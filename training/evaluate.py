#!/usr/bin/env python3
"""
ELO Evaluation: Play Trained Model vs Stockfish to Estimate Rating

This module evaluates trained ChessNet models against Stockfish at various
depths to estimate their ELO rating. It uses the model's MCTS for move
selection and interfaces with Stockfish via python-chess's engine module.

ELO Estimation:
---------------
The script uses the standard ELO formula to estimate the model's rating
based on its win/draw/loss record against Stockfish at a known depth.

    expected_score = 1 / (1 + 10^((elo_opponent - elo_player) / 400))

Inverting this formula given the actual score:

    elo_diff = -400 * log10(1/score - 1)
    elo_player = elo_opponent + elo_diff

Stockfish Depth to Approximate ELO Mapping:
    depth 1  ~ 1300    depth 8  ~ 2300
    depth 2  ~ 1500    depth 10 ~ 2500
    depth 3  ~ 1700    depth 15 ~ 2800
    depth 5  ~ 2000    depth 20 ~ 3000

Usage:
------
    uv run python evaluate.py checkpoint.pt --stockfish-path /path/to/stockfish
    uv run python evaluate.py checkpoint_*.pt --num-games 20 --stockfish-depth 5
    uv run python evaluate.py ckpt1.pt ckpt2.pt --mcts-simulations 200
"""

import argparse
import glob
import math
import os
import sys
import time
from dataclasses import dataclass, field

import chess
import chess.engine
import torch

from board_encoder import get_device
from mcts import MCTS
from model import ChessNet

# Approximate ELO ratings for Stockfish at various search depths.
# These are rough reference points; actual strength depends on hardware
# and Stockfish version.
STOCKFISH_DEPTH_TO_ELO: dict[int, int] = {
    1: 1300,
    2: 1500,
    3: 1700,
    5: 2000,
    8: 2300,
    10: 2500,
    15: 2800,
    20: 3000,
}

# Maximum number of moves before declaring a draw to avoid infinite games.
MAX_GAME_MOVES = 300


@dataclass
class GameResult:
    """Result of a single game between model and Stockfish."""

    game_number: int
    model_color: str  # "white" or "black"
    outcome: str  # "win", "draw", "loss" (from model's perspective)
    num_moves: int
    termination: str  # e.g. "checkmate", "stalemate", "max_moves", etc.
    pgn_moves: list[str] = field(default_factory=list)


@dataclass
class MatchResults:
    """Aggregated results from a match (multiple games)."""

    wins: int = 0
    draws: int = 0
    losses: int = 0
    games: list[GameResult] = field(default_factory=list)

    @property
    def total_games(self) -> int:
        return self.wins + self.draws + self.losses

    @property
    def score(self) -> float:
        """Score as a fraction: (wins + 0.5 * draws) / total_games."""
        if self.total_games == 0:
            return 0.0
        return (self.wins + 0.5 * self.draws) / self.total_games

    @property
    def score_percentage(self) -> float:
        """Score as a percentage."""
        return self.score * 100.0


@dataclass
class EvaluationResult:
    """Full evaluation result for a single checkpoint."""

    checkpoint_path: str
    checkpoint_name: str
    match_results: MatchResults
    stockfish_depth: int
    estimated_elo: float
    elo_ci_low: float
    elo_ci_high: float
    mcts_simulations: int
    evaluation_time: float  # seconds


def get_stockfish_elo(depth: int) -> int:
    """
    Get the approximate ELO rating for Stockfish at a given depth.

    If the exact depth is not in the lookup table, linearly interpolates
    between the two nearest known depths.

    Args:
        depth: Stockfish search depth.

    Returns:
        Approximate ELO rating for Stockfish at that depth.
    """
    if depth in STOCKFISH_DEPTH_TO_ELO:
        return STOCKFISH_DEPTH_TO_ELO[depth]

    # Interpolate between known depths
    known_depths = sorted(STOCKFISH_DEPTH_TO_ELO.keys())

    # Clamp to range
    if depth <= known_depths[0]:
        return STOCKFISH_DEPTH_TO_ELO[known_depths[0]]
    if depth >= known_depths[-1]:
        return STOCKFISH_DEPTH_TO_ELO[known_depths[-1]]

    # Find surrounding known depths
    for i in range(len(known_depths) - 1):
        if known_depths[i] <= depth <= known_depths[i + 1]:
            d_low = known_depths[i]
            d_high = known_depths[i + 1]
            elo_low = STOCKFISH_DEPTH_TO_ELO[d_low]
            elo_high = STOCKFISH_DEPTH_TO_ELO[d_high]

            # Linear interpolation
            fraction = (depth - d_low) / (d_high - d_low)
            return int(elo_low + fraction * (elo_high - elo_low))

    # Fallback (should not reach here)
    return STOCKFISH_DEPTH_TO_ELO[known_depths[-1]]


def estimate_elo(
    score: float,
    num_games: int,
    stockfish_depth: int,
) -> tuple[float, float, float]:
    """
    Estimate the model's ELO rating from its score against Stockfish.

    Uses the inverse ELO formula:
        elo_diff = -400 * log10(1/score - 1)
        elo_player = elo_stockfish + elo_diff

    Also computes a 95% confidence interval using the Wilson score interval
    for the underlying win proportion, then maps the CI bounds through the
    ELO formula.

    Args:
        score: Fractional score (wins + 0.5 * draws) / total_games, in [0, 1].
        num_games: Number of games played.
        stockfish_depth: Stockfish search depth used.

    Returns:
        Tuple of (estimated_elo, ci_low, ci_high) where ci_low and ci_high
        define the 95% confidence interval.
    """
    stockfish_elo = get_stockfish_elo(stockfish_depth)

    # Handle edge cases where the ELO formula is undefined
    if score <= 0.0:
        # Model lost every game; assign a large penalty
        elo_diff = -800.0
    elif score >= 1.0:
        # Model won every game; assign a large bonus
        elo_diff = 800.0
    else:
        elo_diff = -400.0 * math.log10(1.0 / score - 1.0)

    estimated_elo = stockfish_elo + elo_diff

    # Compute 95% confidence interval using Wilson score interval
    # z = 1.96 for 95% confidence
    z = 1.96
    ci_low_score, ci_high_score = _wilson_score_interval(score, num_games, z)

    # Map confidence interval bounds through the ELO formula
    if ci_low_score <= 0.0:
        ci_low_elo = stockfish_elo - 800.0
    else:
        ci_low_elo = stockfish_elo + (-400.0 * math.log10(1.0 / ci_low_score - 1.0))

    if ci_high_score >= 1.0:
        ci_high_elo = stockfish_elo + 800.0
    else:
        ci_high_elo = stockfish_elo + (-400.0 * math.log10(1.0 / ci_high_score - 1.0))

    return estimated_elo, ci_low_elo, ci_high_elo


def _wilson_score_interval(p: float, n: int, z: float = 1.96) -> tuple[float, float]:
    """
    Compute the Wilson score confidence interval for a proportion.

    The Wilson score interval is better-behaved than the normal approximation
    for small sample sizes and proportions near 0 or 1.

    Args:
        p: Observed proportion (score), in [0, 1].
        n: Number of observations (games).
        z: Z-score for the desired confidence level (1.96 for 95%).

    Returns:
        Tuple of (lower_bound, upper_bound) for the confidence interval,
        each clamped to [0.001, 0.999] to avoid division by zero in ELO calc.
    """
    if n == 0:
        return 0.001, 0.999

    denominator = 1.0 + z * z / n
    centre = p + z * z / (2.0 * n)

    # The term under the square root
    discriminant = p * (1.0 - p) / n + z * z / (4.0 * n * n)
    if discriminant < 0:
        discriminant = 0.0
    spread = z * math.sqrt(discriminant)

    lower = (centre - spread) / denominator
    upper = (centre + spread) / denominator

    # Clamp to avoid log10(0) in ELO calculation
    lower = max(0.001, min(lower, 0.999))
    upper = max(0.001, min(upper, 0.999))

    return lower, upper


def load_model_from_checkpoint(
    checkpoint_path: str,
    device: torch.device,
) -> ChessNet:
    """
    Load a ChessNet model from a training checkpoint.

    Follows the same pattern as export_onnx.py for checkpoint loading.

    Args:
        checkpoint_path: Path to the .pt checkpoint file.
        device: Torch device to load the model onto.

    Returns:
        ChessNet model in eval mode on the specified device.

    Raises:
        FileNotFoundError: If the checkpoint file does not exist.
        KeyError: If the checkpoint is missing required keys.
    """
    if not os.path.exists(checkpoint_path):
        raise FileNotFoundError(f"Checkpoint not found: {checkpoint_path}")

    checkpoint = torch.load(checkpoint_path, map_location=device, weights_only=False)

    # Extract model configuration from checkpoint
    config = checkpoint.get("config", {})
    num_blocks = config.get("num_blocks", 6)
    num_filters = config.get("num_filters", 128)

    # Create model with the same architecture
    model = ChessNet(num_blocks=num_blocks, num_filters=num_filters)
    model.load_state_dict(checkpoint["model_state_dict"])
    model = model.to(device)
    model.eval()

    return model


def _determine_outcome(
    board: chess.Board,
    model_is_white: bool,
) -> tuple[str, str]:
    """
    Determine the game outcome from the model's perspective.

    Args:
        board: The board at the end of the game.
        model_is_white: Whether the model was playing as white.

    Returns:
        Tuple of (outcome, termination) where outcome is "win", "draw",
        or "loss" from the model's perspective and termination describes
        how the game ended.
    """
    result = board.outcome()

    if result is None:
        # Game did not end naturally (e.g. max moves reached)
        return "draw", "max_moves"

    termination = result.termination.name.lower()

    if result.winner is None:
        return "draw", termination

    model_won = (result.winner == chess.WHITE) == model_is_white
    if model_won:
        return "win", termination
    else:
        return "loss", termination


def play_match(
    model: ChessNet,
    device: torch.device,
    stockfish_path: str,
    num_games: int,
    stockfish_depth: int,
    mcts_simulations: int,
    stockfish_time_limit: float = 1.0,
) -> MatchResults:
    """
    Play a match of multiple games between the model and Stockfish.

    The model uses MCTS with temperature=0 (deterministic best move) for
    move selection. Stockfish plays at the specified depth and time limit.
    Colors alternate each game: the model plays white in odd-numbered games
    (1, 3, 5, ...) and black in even-numbered games (2, 4, 6, ...).

    Args:
        model: ChessNet model in eval mode.
        device: Torch device for the model.
        stockfish_path: Path to the Stockfish binary.
        num_games: Number of games to play.
        stockfish_depth: Stockfish search depth limit.
        mcts_simulations: Number of MCTS simulations per model move.
        stockfish_time_limit: Stockfish time limit per move in seconds.

    Returns:
        MatchResults containing wins, draws, losses, and per-game details.

    Raises:
        chess.engine.EngineTerminatedError: If Stockfish crashes.
        FileNotFoundError: If Stockfish binary is not found.
    """
    # Create MCTS instance for the model
    mcts = MCTS(
        model=model,
        c_puct=1.5,
        num_simulations=mcts_simulations,
        device=device,
    )

    results = MatchResults()

    # Open Stockfish engine
    try:
        engine = chess.engine.SimpleEngine.popen_uci(stockfish_path)
    except FileNotFoundError:
        print(f"Error: Stockfish binary not found at '{stockfish_path}'")
        print("Please provide a valid path with --stockfish-path")
        sys.exit(1)
    except Exception as e:
        print(f"Error: Could not start Stockfish engine: {e}")
        sys.exit(1)

    try:
        for game_num in range(1, num_games + 1):
            # Alternate colors: model plays white in odd games, black in even
            model_is_white = game_num % 2 == 1
            model_color = "white" if model_is_white else "black"

            board = chess.Board()
            move_list: list[str] = []
            board_history: list[chess.Board] = []
            move_count = 0

            while not board.is_game_over() and move_count < MAX_GAME_MOVES:
                is_model_turn = (board.turn == chess.WHITE) == model_is_white

                if is_model_turn:
                    # Model's turn: use MCTS with history
                    move = mcts.select_move(board, temperature=0, history=board_history)
                else:
                    # Stockfish's turn
                    sf_result = engine.play(
                        board,
                        chess.engine.Limit(
                            depth=stockfish_depth,
                            time=stockfish_time_limit,
                        ),
                    )
                    sf_move = sf_result.move
                    assert sf_move is not None, "Stockfish returned no move"
                    move = sf_move

                move_list.append(board.san(move))
                board_history.append(board.copy())
                board.push(move)
                move_count += 1

            # Determine outcome
            if move_count >= MAX_GAME_MOVES and not board.is_game_over():
                outcome = "draw"
                termination = "max_moves"
            else:
                outcome, termination = _determine_outcome(board, model_is_white)

            # Record result
            game_result = GameResult(
                game_number=game_num,
                model_color=model_color,
                outcome=outcome,
                num_moves=move_count,
                termination=termination,
                pgn_moves=move_list,
            )
            results.games.append(game_result)

            if outcome == "win":
                results.wins += 1
            elif outcome == "draw":
                results.draws += 1
            else:
                results.losses += 1

            # Print progress
            outcome_display = {
                "win": "Model wins",
                "draw": "Draw",
                "loss": "Model loses",
            }[outcome]
            color_display = model_color.capitalize()
            print(
                f"  Game {game_num:>{len(str(num_games))}}/{num_games}: "
                f"{outcome_display} as {color_display} "
                f"({termination}, {move_count} moves)"
            )

    finally:
        engine.quit()

    return results


def evaluate_checkpoint(
    checkpoint_path: str,
    stockfish_path: str,
    num_games: int,
    stockfish_depth: int,
    mcts_simulations: int,
    stockfish_time_limit: float = 1.0,
) -> EvaluationResult:
    """
    Evaluate a single checkpoint against Stockfish and estimate its ELO.

    This is the main evaluation function that ties together model loading,
    match play, and ELO estimation.

    Args:
        checkpoint_path: Path to the .pt checkpoint file.
        stockfish_path: Path to the Stockfish binary.
        num_games: Number of games to play.
        stockfish_depth: Stockfish search depth.
        mcts_simulations: Number of MCTS simulations per model move.
        stockfish_time_limit: Stockfish time limit per move in seconds.

    Returns:
        EvaluationResult with match results and ELO estimate.
    """
    checkpoint_name = os.path.basename(checkpoint_path)
    device = get_device()

    print(f"\n{'=' * 60}")
    print(f"Evaluating: {checkpoint_name}")
    print(f"{'=' * 60}")
    print(f"  Device:           {device}")
    print(f"  MCTS simulations: {mcts_simulations}")
    print(f"  Stockfish depth:  {stockfish_depth}")
    print(f"  Stockfish time:   {stockfish_time_limit}s/move")
    print(f"  Games to play:    {num_games}")
    print()

    # Load the model
    print(f"  Loading checkpoint: {checkpoint_path}")
    model = load_model_from_checkpoint(checkpoint_path, device)
    print(f"  Model loaded: {model.count_parameters():,} parameters")
    print(f"  Architecture: {model.num_blocks} blocks, {model.num_filters} filters")
    print()

    # Play the match
    start_time = time.time()
    match_results = play_match(
        model=model,
        device=device,
        stockfish_path=stockfish_path,
        num_games=num_games,
        stockfish_depth=stockfish_depth,
        mcts_simulations=mcts_simulations,
        stockfish_time_limit=stockfish_time_limit,
    )
    evaluation_time = time.time() - start_time

    # Estimate ELO
    elo, elo_ci_low, elo_ci_high = estimate_elo(
        score=match_results.score,
        num_games=match_results.total_games,
        stockfish_depth=stockfish_depth,
    )

    return EvaluationResult(
        checkpoint_path=checkpoint_path,
        checkpoint_name=checkpoint_name,
        match_results=match_results,
        stockfish_depth=stockfish_depth,
        estimated_elo=elo,
        elo_ci_low=elo_ci_low,
        elo_ci_high=elo_ci_high,
        mcts_simulations=mcts_simulations,
        evaluation_time=evaluation_time,
    )


def print_results(result: EvaluationResult) -> None:
    """
    Print a formatted summary of a single evaluation result.

    Args:
        result: The evaluation result to display.
    """
    mr = result.match_results

    print(f"\n{'=' * 60}")
    print(f"Results: {result.checkpoint_name}")
    print(f"{'=' * 60}")
    print(f"  Record:        W:{mr.wins}  D:{mr.draws}  L:{mr.losses}")
    print(f"  Score:         {mr.score_percentage:.1f}%")
    print(f"  vs Stockfish:  depth {result.stockfish_depth} (~{get_stockfish_elo(result.stockfish_depth)} ELO)")
    print(f"  Estimated ELO: ~{result.estimated_elo:.0f} (95% CI: {result.elo_ci_low:.0f} - {result.elo_ci_high:.0f})")
    print(f"  MCTS sims:     {result.mcts_simulations}")
    print(
        f"  Time:          {result.evaluation_time:.1f}s ({result.evaluation_time / max(mr.total_games, 1):.1f}s/game)"
    )
    print(f"{'=' * 60}")


def print_summary_table(results: list[EvaluationResult]) -> None:
    """
    Print a summary table mapping checkpoints to their estimated ELO.

    This is displayed after evaluating multiple checkpoints to provide
    a convenient comparison.

    Args:
        results: List of evaluation results to display.
    """
    if not results:
        return

    # Determine column widths
    max_name_len = max(len(r.checkpoint_name) for r in results)
    name_width = max(max_name_len, len("Checkpoint"))

    print(f"\n{'=' * (name_width + 65)}")
    print(f"{'Checkpoint -> ELO Mapping':^{name_width + 65}}")
    print(f"{'=' * (name_width + 65)}")

    for r in results:
        mr = r.match_results
        ci_half = (r.elo_ci_high - r.elo_ci_low) / 2.0
        print(
            f"  {r.checkpoint_name:<{name_width}} | "
            f"W:{mr.wins:<3} D:{mr.draws:<3} L:{mr.losses:<3} | "
            f"Score: {mr.score_percentage:5.1f}% | "
            f"ELO: ~{r.estimated_elo:>5.0f} +/- {ci_half:<.0f}"
        )

    print(f"{'=' * (name_width + 65)}")


def resolve_checkpoint_paths(patterns: list[str]) -> list[str]:
    """
    Resolve checkpoint path patterns (which may contain globs) to actual files.

    Args:
        patterns: List of file paths or glob patterns.

    Returns:
        Sorted list of resolved, unique checkpoint file paths.

    Raises:
        SystemExit: If no checkpoint files are found.
    """
    paths: list[str] = []
    for pattern in patterns:
        # Try glob expansion
        expanded = glob.glob(pattern)
        if expanded:
            paths.extend(expanded)
        elif os.path.exists(pattern):
            paths.append(pattern)
        else:
            print(f"Warning: No files matched pattern '{pattern}'")

    # Remove duplicates and sort
    paths = sorted(set(paths))

    if not paths:
        print("Error: No checkpoint files found.")
        sys.exit(1)

    return paths


def parse_args() -> argparse.Namespace:
    """Parse command-line arguments."""
    parser = argparse.ArgumentParser(
        description="Evaluate chess model checkpoints against Stockfish to estimate ELO",
        formatter_class=argparse.ArgumentDefaultsHelpFormatter,
    )

    parser.add_argument(
        "checkpoint_paths",
        nargs="+",
        type=str,
        help="Path(s) to checkpoint .pt file(s). Supports glob patterns.",
    )

    parser.add_argument(
        "--stockfish-path",
        type=str,
        default="stockfish",
        help="Path to the Stockfish binary (default: 'stockfish' in PATH)",
    )

    parser.add_argument(
        "--num-games",
        type=int,
        default=20,
        help="Number of games to play per checkpoint",
    )

    parser.add_argument(
        "--stockfish-depth",
        type=int,
        default=5,
        help="Stockfish search depth",
    )

    parser.add_argument(
        "--stockfish-time-limit",
        type=float,
        default=1.0,
        help="Stockfish time limit per move in seconds",
    )

    parser.add_argument(
        "--mcts-simulations",
        type=int,
        default=200,
        help="Number of MCTS simulations for model move selection",
    )

    return parser.parse_args()


def main():
    """Main entry point for ELO evaluation."""
    args = parse_args()

    # Resolve checkpoint paths (handles glob patterns)
    checkpoint_paths = resolve_checkpoint_paths(args.checkpoint_paths)

    print(f"Found {len(checkpoint_paths)} checkpoint(s) to evaluate:")
    for path in checkpoint_paths:
        print(f"  - {path}")
    print(f"\nStockfish:  {args.stockfish_path}")
    print(f"Depth:      {args.stockfish_depth} (~{get_stockfish_elo(args.stockfish_depth)} ELO)")
    print(f"Games:      {args.num_games} per checkpoint")
    print(f"MCTS sims:  {args.mcts_simulations}")

    # Evaluate each checkpoint
    all_results: list[EvaluationResult] = []

    for checkpoint_path in checkpoint_paths:
        try:
            result = evaluate_checkpoint(
                checkpoint_path=checkpoint_path,
                stockfish_path=args.stockfish_path,
                num_games=args.num_games,
                stockfish_depth=args.stockfish_depth,
                mcts_simulations=args.mcts_simulations,
                stockfish_time_limit=args.stockfish_time_limit,
            )
            all_results.append(result)

            # Print individual results
            print_results(result)

        except FileNotFoundError as e:
            print(f"\nError: {e}")
            print(f"Skipping checkpoint: {checkpoint_path}")
            continue
        except Exception as e:
            print(f"\nUnexpected error evaluating {checkpoint_path}: {e}")
            print(f"Skipping checkpoint: {checkpoint_path}")
            continue

    # Print summary table if multiple checkpoints were evaluated
    if len(all_results) > 1:
        print_summary_table(all_results)
    elif len(all_results) == 0:
        print("\nNo checkpoints were successfully evaluated.")
        sys.exit(1)

    print("\nEvaluation complete.")


if __name__ == "__main__":
    main()
