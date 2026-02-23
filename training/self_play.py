"""
Self-Play Game Generation for AlphaZero-style Chess Training

This module implements self-play game generation where a neural network plays
against itself using MCTS to generate training data. The generated games
produce training examples that capture:

1. Board positions encountered during play
2. MCTS-derived move probabilities (policy targets)
3. Game outcomes (value targets)

Temperature Schedule:
---------------------
Following AlphaZero, we use a temperature schedule for move selection:
- First 30 moves: temperature = 1.0 (more exploration)
- After 30 moves: temperature = 0.2 (more exploitation)

This encourages diverse opening play while still playing strongly in the
middle and endgame.

Training Example Generation:
----------------------------
For each position in a game, we store:
- The encoded board state [18, 8, 8]
- The MCTS visit count policy [4096]
- The game outcome from that player's perspective [-1, 1]

The value targets are determined after the game ends:
- +1.0 if the player at that position won
- -1.0 if the player at that position lost
- 0.0 for draws
"""

from dataclasses import dataclass
from typing import Dict, List, Optional, Tuple
import time

import chess
import numpy as np
import torch

from board_encoder import encode_board, get_device
from model import ChessNet, create_model, POLICY_OUTPUT_SIZE
from mcts import MCTS
from replay_buffer import TrainingExample, ReplayBuffer


# Temperature schedule constants
EXPLORATION_MOVES = 30  # Number of moves to use high temperature
HIGH_TEMPERATURE = 1.0  # Temperature for first N moves
LOW_TEMPERATURE = 0.2   # Temperature for remaining moves


@dataclass
class GameStats:
    """
    Statistics for a completed self-play game.

    Attributes:
        num_moves: Total number of moves (half-moves/plies) in the game
        result: Game result as string ("1-0", "0-1", or "1/2-1/2")
        winner: chess.WHITE, chess.BLACK, or None for draw
        termination: How the game ended (checkmate, stalemate, etc.)
        duration_seconds: Time taken to play the game
    """
    num_moves: int
    result: str
    winner: Optional[bool]  # chess.WHITE (True), chess.BLACK (False), or None
    termination: str
    duration_seconds: float


@dataclass
class GameData:
    """
    Data from a completed self-play game.

    Attributes:
        examples: List of training examples from the game
        stats: Game statistics
    """
    examples: List[TrainingExample]
    stats: GameStats


def _moves_to_policy_vector(
    visit_counts: Dict[chess.Move, int]
) -> np.ndarray:
    """
    Convert MCTS visit counts to a policy vector.

    The policy vector has 4096 elements (64 from-squares x 64 to-squares).
    Visit counts are normalized to sum to 1.

    Args:
        visit_counts: Dict mapping moves to their visit counts

    Returns:
        numpy array of shape [4096] with normalized probabilities
    """
    policy = np.zeros(POLICY_OUTPUT_SIZE, dtype=np.float32)

    total_visits = sum(visit_counts.values())
    if total_visits == 0:
        return policy

    for move, count in visit_counts.items():
        # Policy index = from_square * 64 + to_square
        idx = move.from_square * 64 + move.to_square
        policy[idx] = count / total_visits

    return policy


def _get_temperature(move_number: int) -> float:
    """
    Get the temperature for move selection based on move number.

    Uses high temperature (1.0) for the first 30 moves to encourage
    exploration, then low temperature (0.2) for remaining moves.

    Args:
        move_number: The current move number (0-indexed, counts half-moves)

    Returns:
        Temperature value for move selection
    """
    if move_number < EXPLORATION_MOVES:
        return HIGH_TEMPERATURE
    return LOW_TEMPERATURE


def _get_game_outcome(
    board: chess.Board, perspective: bool
) -> float:
    """
    Get the game outcome from a specific player's perspective.

    Args:
        board: Chess board at game end
        perspective: The player's perspective (chess.WHITE or chess.BLACK)

    Returns:
        +1.0 if the player won
        -1.0 if the player lost
        0.0 for draws
    """
    result = board.outcome()
    if result is None or result.winner is None:
        # Draw
        return 0.0

    if result.winner == perspective:
        return 1.0
    return -1.0


def play_game(
    mcts: MCTS,
    max_moves: int = 512
) -> GameData:
    """
    Play a single self-play game using MCTS.

    This function plays a complete game where both sides use the same
    neural network and MCTS for move selection. Training examples are
    collected at each position.

    Args:
        mcts: MCTS instance with neural network
        max_moves: Maximum number of moves before declaring draw (default: 512)

    Returns:
        GameData containing training examples and game statistics
    """
    start_time = time.time()
    board = chess.Board()

    # Store positions, policies, and whose turn it was
    positions: List[Tuple[np.ndarray, np.ndarray, bool]] = []

    move_number = 0
    while not board.is_game_over() and move_number < max_moves:
        # Get temperature for this move
        temperature = _get_temperature(move_number)

        # Run MCTS search
        visit_counts = mcts.search(board)

        if len(visit_counts) == 0:
            # No legal moves (game should be over)
            break

        # Convert visit counts to policy vector
        policy = _moves_to_policy_vector(visit_counts)

        # Store the position data
        board_state = encode_board(board)
        positions.append((board_state, policy, board.turn))

        # Get action probabilities and select move
        moves, probs = mcts.get_action_probabilities(board, temperature)

        # Sample from distribution
        idx = np.random.choice(len(moves), p=probs)
        selected_move = moves[idx]

        # Make the move
        board.push(selected_move)
        move_number += 1

    # Game is over - determine outcome
    duration = time.time() - start_time

    # Get game result
    outcome = board.outcome()
    if outcome is not None:
        result_str = outcome.result()
        winner = outcome.winner
        termination = outcome.termination.name
    else:
        # Max moves reached
        result_str = "1/2-1/2"
        winner = None
        termination = "MAX_MOVES"

    # Create training examples with correct value targets
    examples: List[TrainingExample] = []
    for board_state, policy, perspective in positions:
        value = _get_game_outcome(board, perspective)
        examples.append(TrainingExample(
            board_state=board_state,
            policy_target=policy,
            value_target=value
        ))

    stats = GameStats(
        num_moves=move_number,
        result=result_str,
        winner=winner,
        termination=termination,
        duration_seconds=duration
    )

    return GameData(examples=examples, stats=stats)


def generate_games(
    model: ChessNet,
    num_games: int = 100,
    num_simulations: int = 400,
    c_puct: float = 1.5,
    max_moves: int = 512,
    device: Optional[torch.device] = None,
    verbose: bool = True
) -> Tuple[List[TrainingExample], List[GameStats]]:
    """
    Generate multiple self-play games.

    This is the main entry point for self-play game generation. It plays
    multiple games sequentially and collects all training examples.

    Args:
        model: ChessNet neural network for evaluation
        num_games: Number of games to generate (default: 100)
        num_simulations: MCTS simulations per move (default: 400)
        c_puct: MCTS exploration constant (default: 1.5)
        max_moves: Maximum moves per game before draw (default: 512)
        device: Torch device for inference (default: auto-detect)
        verbose: Whether to print progress (default: True)

    Returns:
        Tuple of:
        - List of all training examples from all games
        - List of game statistics
    """
    if device is None:
        device = get_device()

    # Create MCTS instance
    mcts = MCTS(
        model=model,
        c_puct=c_puct,
        num_simulations=num_simulations,
        device=device
    )

    all_examples: List[TrainingExample] = []
    all_stats: List[GameStats] = []

    for game_idx in range(num_games):
        game_data = play_game(mcts, max_moves=max_moves)

        all_examples.extend(game_data.examples)
        all_stats.append(game_data.stats)

        if verbose:
            stats = game_data.stats
            print(
                f"Game {game_idx + 1}/{num_games}: "
                f"{stats.result} in {stats.num_moves} moves "
                f"({stats.termination}) - {stats.duration_seconds:.1f}s"
            )

    if verbose:
        # Print summary statistics
        total_positions = len(all_examples)
        avg_moves = np.mean([s.num_moves for s in all_stats])
        white_wins = sum(1 for s in all_stats if s.winner == chess.WHITE)
        black_wins = sum(1 for s in all_stats if s.winner == chess.BLACK)
        draws = sum(1 for s in all_stats if s.winner is None)
        total_time = sum(s.duration_seconds for s in all_stats)

        print(f"\n--- Self-Play Summary ---")
        print(f"Games played: {num_games}")
        print(f"Total positions: {total_positions}")
        print(f"Average game length: {avg_moves:.1f} moves")
        print(f"Results: White +{white_wins}, Black +{black_wins}, Draws ={draws}")
        print(f"Total time: {total_time:.1f}s ({total_time/num_games:.1f}s per game)")

    return all_examples, all_stats


def generate_games_to_buffer(
    model: ChessNet,
    buffer: ReplayBuffer,
    num_games: int = 100,
    num_simulations: int = 400,
    c_puct: float = 1.5,
    max_moves: int = 512,
    device: Optional[torch.device] = None,
    verbose: bool = True
) -> List[GameStats]:
    """
    Generate self-play games and add examples directly to a replay buffer.

    This is a convenience function that combines game generation with
    buffer storage. Use this when you want to populate a buffer for training.

    Args:
        model: ChessNet neural network for evaluation
        buffer: ReplayBuffer to store examples
        num_games: Number of games to generate (default: 100)
        num_simulations: MCTS simulations per move (default: 400)
        c_puct: MCTS exploration constant (default: 1.5)
        max_moves: Maximum moves per game before draw (default: 512)
        device: Torch device for inference (default: auto-detect)
        verbose: Whether to print progress (default: True)

    Returns:
        List of game statistics
    """
    examples, stats = generate_games(
        model=model,
        num_games=num_games,
        num_simulations=num_simulations,
        c_puct=c_puct,
        max_moves=max_moves,
        device=device,
        verbose=verbose
    )

    buffer.add_batch(examples)

    if verbose:
        print(f"Added {len(examples)} examples to buffer (total: {len(buffer)})")

    return stats


class SelfPlayManager:
    """
    Manager for self-play game generation with configurable parameters.

    This class provides a higher-level interface for self-play that maintains
    state across multiple generation rounds. It's useful for iterative training
    where you generate games, train, and repeat.

    Attributes:
        model: ChessNet neural network
        buffer: ReplayBuffer for storing examples
        num_simulations: MCTS simulations per move
        c_puct: MCTS exploration constant
        max_moves: Maximum moves per game
        device: Torch device
        total_games: Total number of games generated
        total_positions: Total number of positions generated

    Example:
        >>> model = create_model()
        >>> manager = SelfPlayManager(model, num_simulations=100)
        >>> manager.generate(num_games=10)
        >>> batch = manager.sample_batch(32)
    """

    def __init__(
        self,
        model: ChessNet,
        buffer: Optional[ReplayBuffer] = None,
        num_simulations: int = 400,
        c_puct: float = 1.5,
        max_moves: int = 512,
        buffer_size: int = 500_000,
        device: Optional[torch.device] = None
    ):
        """
        Initialize the self-play manager.

        Args:
            model: ChessNet neural network
            buffer: ReplayBuffer to use (creates new one if None)
            num_simulations: MCTS simulations per move (default: 400)
            c_puct: MCTS exploration constant (default: 1.5)
            max_moves: Maximum moves per game (default: 512)
            buffer_size: Size of replay buffer if creating new (default: 500,000)
            device: Torch device (default: auto-detect)
        """
        self.model = model
        self.buffer = buffer if buffer is not None else ReplayBuffer(buffer_size)
        self.num_simulations = num_simulations
        self.c_puct = c_puct
        self.max_moves = max_moves
        self.device = device if device is not None else get_device()

        # Statistics
        self.total_games = 0
        self.total_positions = 0
        self.all_stats: List[GameStats] = []

    def update_model(self, model: ChessNet) -> None:
        """
        Update the neural network model.

        Call this after training to use the improved model for future games.

        Args:
            model: New ChessNet model
        """
        self.model = model

    def generate(
        self,
        num_games: int = 100,
        verbose: bool = True
    ) -> List[GameStats]:
        """
        Generate self-play games and add to buffer.

        Args:
            num_games: Number of games to generate
            verbose: Whether to print progress

        Returns:
            List of game statistics for this generation round
        """
        stats = generate_games_to_buffer(
            model=self.model,
            buffer=self.buffer,
            num_games=num_games,
            num_simulations=self.num_simulations,
            c_puct=self.c_puct,
            max_moves=self.max_moves,
            device=self.device,
            verbose=verbose
        )

        self.total_games += num_games
        self.total_positions = len(self.buffer)
        self.all_stats.extend(stats)

        return stats

    def sample_batch(
        self, batch_size: int
    ) -> Tuple[np.ndarray, np.ndarray, np.ndarray]:
        """
        Sample a batch of training examples from the buffer.

        Args:
            batch_size: Number of examples to sample

        Returns:
            Tuple of (board_states, policy_targets, value_targets)
        """
        return self.buffer.sample(batch_size)

    def get_statistics(self) -> Dict:
        """
        Get summary statistics for all games generated.

        Returns:
            Dict with summary statistics
        """
        if not self.all_stats:
            return {
                "total_games": 0,
                "total_positions": 0,
                "buffer_size": len(self.buffer)
            }

        return {
            "total_games": self.total_games,
            "total_positions": self.total_positions,
            "buffer_size": len(self.buffer),
            "avg_game_length": np.mean([s.num_moves for s in self.all_stats]),
            "white_wins": sum(1 for s in self.all_stats if s.winner == chess.WHITE),
            "black_wins": sum(1 for s in self.all_stats if s.winner == chess.BLACK),
            "draws": sum(1 for s in self.all_stats if s.winner is None),
            "avg_game_time": np.mean([s.duration_seconds for s in self.all_stats])
        }
