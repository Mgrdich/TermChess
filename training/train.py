"""
AlphaZero-style Training Loop for Chess Neural Network

This module implements the main training loop that:
1. Generates self-play games using MCTS
2. Stores training examples in a replay buffer
3. Trains the neural network on sampled batches
4. Saves checkpoints at specified intervals

Training Flow:
--------------
For each iteration:
    1. Generate self-play games with current model
    2. Add training examples to replay buffer
    3. Sample batches and train:
        - Policy loss: Cross-entropy between predicted logits and MCTS policy targets
        - Value loss: MSE between predicted value and game outcome
    4. Log training metrics
    5. Save checkpoint at intervals

Command-line Usage:
-------------------
    uv run python train.py --iterations 1000 --games-per-iter 100 --batch-size 256

Resume from checkpoint:
    uv run python train.py --resume checkpoints/checkpoint_5000.pt
"""

import argparse
import csv
import logging
import os
import sys
import time
from dataclasses import dataclass, fields
from typing import Dict, List, Optional, Tuple

import numpy as np
import torch
import torch.nn as nn
import torch.nn.functional as F
import torch.optim as optim
from torch.optim.lr_scheduler import LambdaLR

from board_encoder import get_device
from model import ChessNet, create_model
from replay_buffer import ReplayBuffer
from self_play import SelfPlayManager, GameStats


# Default training parameters (from technical requirements)
DEFAULT_NUM_ITERATIONS = 80_000
DEFAULT_GAMES_PER_ITERATION = 100
DEFAULT_BATCH_SIZE = 256
DEFAULT_BATCHES_PER_ITERATION = 10
DEFAULT_BUFFER_SIZE = 500_000
DEFAULT_MCTS_SIMULATIONS = 400
DEFAULT_INITIAL_LR = 0.001
DEFAULT_FINAL_LR = 0.0001
DEFAULT_WEIGHT_DECAY = 1e-4
DEFAULT_CHECKPOINT_DIR = "checkpoints"

# Checkpoint intervals — save early and often for ELO evaluation
DEFAULT_CHECKPOINT_INTERVALS = [10, 25, 50, 100, 250, 500, 1_000, 2_500, 5_000, 10_000, 30_000, 80_000]

# Model architecture defaults
DEFAULT_NUM_BLOCKS = 6
DEFAULT_NUM_FILTERS = 128


@dataclass
class TrainingConfig:
    """Configuration for training run."""

    # Training loop parameters
    num_iterations: int = DEFAULT_NUM_ITERATIONS
    games_per_iteration: int = DEFAULT_GAMES_PER_ITERATION
    batch_size: int = DEFAULT_BATCH_SIZE
    batches_per_iteration: int = DEFAULT_BATCHES_PER_ITERATION

    # Replay buffer
    buffer_size: int = DEFAULT_BUFFER_SIZE
    min_buffer_size: int = 1000  # Minimum examples before training starts

    # MCTS parameters
    mcts_simulations: int = DEFAULT_MCTS_SIMULATIONS
    c_puct: float = 1.5
    max_moves_per_game: int = 512

    # Optimizer parameters
    initial_lr: float = DEFAULT_INITIAL_LR
    final_lr: float = DEFAULT_FINAL_LR
    weight_decay: float = DEFAULT_WEIGHT_DECAY
    lr_decay_steps: int = 50_000  # Decay LR at this point

    # Model architecture
    num_blocks: int = DEFAULT_NUM_BLOCKS
    num_filters: int = DEFAULT_NUM_FILTERS

    # Checkpointing
    checkpoint_dir: str = DEFAULT_CHECKPOINT_DIR
    checkpoint_intervals: List[int] = None  # Uses DEFAULT_CHECKPOINT_INTERVALS
    save_every_n_iterations: int = 0  # Also save every N iterations (0=disabled)

    # Logging
    log_every_n_iterations: int = 1
    verbose: bool = True
    verbose_self_play: bool = False

    def __post_init__(self):
        if self.checkpoint_intervals is None:
            self.checkpoint_intervals = DEFAULT_CHECKPOINT_INTERVALS.copy()
        if self.batches_per_iteration < 1:
            raise ValueError(f"batches_per_iteration must be >= 1, got {self.batches_per_iteration}")


@dataclass
class TrainingMetrics:
    """Metrics from a single training iteration."""

    iteration: int
    policy_loss: float
    value_loss: float
    total_loss: float
    games_played: int
    positions_generated: int
    buffer_size: int
    learning_rate: float
    iteration_time: float
    # Game statistics for analysis
    avg_game_length: float = 0.0
    white_wins: int = 0
    black_wins: int = 0
    draws: int = 0
    checkmates: int = 0
    repetition_draws: int = 0
    stalemates: int = 0
    max_moves_draws: int = 0


def setup_logging(verbose: bool = True) -> logging.Logger:
    """
    Set up logging for training.

    Args:
        verbose: If True, log to console. If False, no handlers are added.

    Returns:
        Configured logger instance.
    """
    logger = logging.getLogger("train")
    logger.setLevel(logging.INFO)

    # Avoid adding duplicate handlers on repeated calls
    if logger.handlers:
        return logger

    # Console handler
    if verbose:
        console_handler = logging.StreamHandler(sys.stdout)
        console_handler.setLevel(logging.INFO)
        console_format = logging.Formatter(
            "%(asctime)s - %(levelname)s - %(message)s",
            datefmt="%Y-%m-%d %H:%M:%S"
        )
        console_handler.setFormatter(console_format)
        logger.addHandler(console_handler)

    return logger


def _compute_game_stats(game_stats: List["GameStats"]) -> Dict:
    """Compute summary statistics from a list of game stats."""
    if not game_stats:
        return {}
    return {
        "avg_game_length": np.mean([s.num_moves for s in game_stats]),
        "white_wins": sum(1 for s in game_stats if s.winner is True),
        "black_wins": sum(1 for s in game_stats if s.winner is False),
        "draws": sum(1 for s in game_stats if s.winner is None),
        "checkmates": sum(1 for s in game_stats if s.termination == "CHECKMATE"),
        "repetition_draws": sum(
            1 for s in game_stats
            if s.termination in ("FIVEFOLD_REPETITION", "THREEFOLD_REPETITION")
        ),
        "stalemates": sum(1 for s in game_stats if s.termination == "STALEMATE"),
        "max_moves_draws": sum(1 for s in game_stats if s.termination == "MAX_MOVES"),
    }


def _init_csv_log(log_path: str) -> None:
    """Initialize the CSV training log with headers."""
    header = [f.name for f in fields(TrainingMetrics)]
    with open(log_path, "w", newline="") as f:
        writer = csv.writer(f)
        writer.writerow(header)


def _append_csv_log(log_path: str, metrics: "TrainingMetrics") -> None:
    """Append a single metrics row to the CSV training log."""
    row = [getattr(metrics, f.name) for f in fields(TrainingMetrics)]
    with open(log_path, "a", newline="") as f:
        writer = csv.writer(f)
        writer.writerow(row)


def create_optimizer(
    model: ChessNet,
    initial_lr: float,
    weight_decay: float
) -> optim.Optimizer:
    """
    Create Adam optimizer with weight decay.

    Args:
        model: ChessNet model to optimize
        initial_lr: Initial learning rate
        weight_decay: L2 regularization weight

    Returns:
        Configured optimizer
    """
    return optim.Adam(
        model.parameters(),
        lr=initial_lr,
        weight_decay=weight_decay
    )


def create_scheduler(
    optimizer: optim.Optimizer,
    initial_lr: float,
    final_lr: float,
    decay_steps: int
) -> optim.lr_scheduler.LRScheduler:
    """
    Create learning rate scheduler for decay.

    Uses LambdaLR to decay from initial_lr to final_lr at decay_steps.
    The learning rate is clamped at final_lr and will not decay further.

    Args:
        optimizer: Optimizer to schedule
        initial_lr: Starting learning rate
        final_lr: Final learning rate after decay
        decay_steps: Step at which to apply decay

    Returns:
        Learning rate scheduler
    """
    ratio = final_lr / initial_lr

    def lr_lambda(step: int) -> float:
        if step >= decay_steps:
            return ratio
        return 1.0

    return LambdaLR(optimizer, lr_lambda=lr_lambda)


def compute_loss(
    model: ChessNet,
    states: torch.Tensor,
    policy_targets: torch.Tensor,
    value_targets: torch.Tensor,
    device: torch.device
) -> Tuple[torch.Tensor, torch.Tensor, torch.Tensor]:
    """
    Compute policy and value losses.

    Args:
        model: ChessNet model
        states: Batch of board states [batch, 18, 8, 8]
        policy_targets: MCTS policy targets [batch, 4096]
        value_targets: Game outcome targets [batch]
        device: Torch device

    Returns:
        Tuple of (policy_loss, value_loss, total_loss)
    """
    # Move data to device
    states = torch.from_numpy(states).to(device)
    policy_targets = torch.from_numpy(policy_targets).to(device)
    value_targets = torch.from_numpy(value_targets).to(device)

    # Forward pass
    policy_logits, value_pred = model(states)

    # Policy loss: Cross-entropy between predicted logits and MCTS policy
    # Note: policy_targets are probabilities, so we use cross_entropy with soft targets
    # Cross-entropy with soft targets: -sum(target * log_softmax(pred))
    log_probs = F.log_softmax(policy_logits, dim=1)
    policy_loss = -torch.sum(policy_targets * log_probs, dim=1).mean()

    # Value loss: MSE between predicted value and game outcome
    value_loss = F.mse_loss(value_pred.squeeze(-1), value_targets)

    # Total loss: equal weighting
    total_loss = policy_loss + value_loss

    return policy_loss, value_loss, total_loss


def train_batch(
    model: ChessNet,
    optimizer: optim.Optimizer,
    states: np.ndarray,
    policy_targets: np.ndarray,
    value_targets: np.ndarray,
    device: torch.device
) -> Tuple[float, float, float]:
    """
    Train on a single batch.

    Args:
        model: ChessNet model
        optimizer: Optimizer
        states: Batch of board states
        policy_targets: MCTS policy targets
        value_targets: Game outcome targets
        device: Torch device

    Returns:
        Tuple of (policy_loss, value_loss, total_loss) as floats
    """
    model.train()

    # Compute loss
    policy_loss, value_loss, total_loss = compute_loss(
        model, states, policy_targets, value_targets, device
    )

    # Backward pass
    optimizer.zero_grad()
    total_loss.backward()
    optimizer.step()

    return (
        policy_loss.item(),
        value_loss.item(),
        total_loss.item()
    )


def train_iteration(
    model: ChessNet,
    optimizer: optim.Optimizer,
    buffer: ReplayBuffer,
    batch_size: int,
    batches_per_iteration: int,
    device: torch.device
) -> Tuple[float, float, float]:
    """
    Run one training iteration (multiple batches).

    Args:
        model: ChessNet model
        optimizer: Optimizer
        buffer: Replay buffer to sample from
        batch_size: Size of each batch
        batches_per_iteration: Number of batches to train per iteration
        device: Torch device

    Returns:
        Tuple of average (policy_loss, value_loss, total_loss)
    """
    total_policy_loss = 0.0
    total_value_loss = 0.0
    total_loss = 0.0

    for _ in range(batches_per_iteration):
        # Sample batch from buffer
        states, policies, values = buffer.sample(batch_size)

        # Train on batch
        p_loss, v_loss, t_loss = train_batch(
            model, optimizer, states, policies, values, device
        )

        total_policy_loss += p_loss
        total_value_loss += v_loss
        total_loss += t_loss

    # Return averages
    n = batches_per_iteration
    return (
        total_policy_loss / n,
        total_value_loss / n,
        total_loss / n
    )


def save_checkpoint(
    model: ChessNet,
    optimizer: optim.Optimizer,
    scheduler: optim.lr_scheduler.LRScheduler,
    iteration: int,
    config: TrainingConfig,
    checkpoint_dir: str,
    metrics: Optional[TrainingMetrics] = None
) -> str:
    """
    Save a training checkpoint.

    Args:
        model: ChessNet model
        optimizer: Optimizer state
        scheduler: Learning rate scheduler state
        iteration: Current iteration number
        config: Training configuration
        checkpoint_dir: Directory to save checkpoint
        metrics: Optional training metrics to save

    Returns:
        Path to saved checkpoint
    """
    # Create checkpoint directory if it doesn't exist
    os.makedirs(checkpoint_dir, exist_ok=True)

    # Build checkpoint path
    checkpoint_path = os.path.join(
        checkpoint_dir,
        f"checkpoint_{iteration}.pt"
    )

    # Prepare checkpoint data
    checkpoint = {
        "iteration": iteration,
        "model_state_dict": model.state_dict(),
        "optimizer_state_dict": optimizer.state_dict(),
        "scheduler_state_dict": scheduler.state_dict(),
        "config": {
            "num_blocks": config.num_blocks,
            "num_filters": config.num_filters,
            "num_iterations": config.num_iterations,
            "games_per_iteration": config.games_per_iteration,
            "batch_size": config.batch_size,
            "batches_per_iteration": config.batches_per_iteration,
            "initial_lr": config.initial_lr,
            "final_lr": config.final_lr,
            "weight_decay": config.weight_decay,
            "lr_decay_steps": config.lr_decay_steps,
        }
    }

    if metrics is not None:
        checkpoint["metrics"] = {
            "policy_loss": metrics.policy_loss,
            "value_loss": metrics.value_loss,
            "total_loss": metrics.total_loss,
            "buffer_size": metrics.buffer_size,
            "avg_game_length": metrics.avg_game_length,
            "white_wins": metrics.white_wins,
            "black_wins": metrics.black_wins,
            "draws": metrics.draws,
            "checkmates": metrics.checkmates,
            "repetition_draws": metrics.repetition_draws,
            "stalemates": metrics.stalemates,
            "max_moves_draws": metrics.max_moves_draws,
        }

    # Save checkpoint
    torch.save(checkpoint, checkpoint_path)

    return checkpoint_path


def load_checkpoint(
    checkpoint_path: str,
    device: torch.device
) -> Tuple[ChessNet, optim.Optimizer, optim.lr_scheduler.LRScheduler, int, Dict]:
    """
    Load a training checkpoint.

    Args:
        checkpoint_path: Path to checkpoint file
        device: Device to load model onto

    Returns:
        Tuple of (model, optimizer, scheduler, iteration, config_dict)
    """
    checkpoint = torch.load(checkpoint_path, map_location=device, weights_only=False)

    config_dict = checkpoint["config"]

    # Create model with same architecture
    model = create_model(
        num_blocks=config_dict["num_blocks"],
        num_filters=config_dict["num_filters"],
        device=device
    )
    model.load_state_dict(checkpoint["model_state_dict"])

    # Create optimizer
    optimizer = create_optimizer(
        model,
        initial_lr=config_dict["initial_lr"],
        weight_decay=config_dict["weight_decay"]
    )
    optimizer.load_state_dict(checkpoint["optimizer_state_dict"])

    # Create scheduler using saved config (with fallback defaults)
    scheduler = create_scheduler(
        optimizer,
        initial_lr=config_dict.get("initial_lr", DEFAULT_INITIAL_LR),
        final_lr=config_dict.get("final_lr", DEFAULT_FINAL_LR),
        decay_steps=config_dict.get("lr_decay_steps", 50_000),
    )
    scheduler.load_state_dict(checkpoint["scheduler_state_dict"])

    iteration = checkpoint["iteration"]

    return model, optimizer, scheduler, iteration, config_dict


def train(config: TrainingConfig, resume_from: Optional[str] = None) -> ChessNet:
    """
    Main training loop.

    Args:
        config: Training configuration
        resume_from: Optional path to checkpoint to resume from

    Returns:
        Trained ChessNet model
    """
    logger = setup_logging(config.verbose)
    device = get_device()

    logger.info(f"Starting training on device: {device}")
    logger.info(f"Configuration: {config.num_iterations} iterations, "
                f"{config.games_per_iteration} games/iter, "
                f"batch_size={config.batch_size}")

    # Initialize or resume
    start_iteration = 0

    if resume_from is not None:
        logger.info(f"Resuming from checkpoint: {resume_from}")
        model, optimizer, scheduler, start_iteration, _ = load_checkpoint(
            resume_from, device
        )
        logger.info(f"Resumed at iteration {start_iteration}")
    else:
        # Create new model
        model = create_model(
            num_blocks=config.num_blocks,
            num_filters=config.num_filters,
            device=device
        )

        # Create optimizer and scheduler
        optimizer = create_optimizer(model, config.initial_lr, config.weight_decay)
        scheduler = create_scheduler(
            optimizer,
            config.initial_lr,
            config.final_lr,
            config.lr_decay_steps
        )

    logger.info(f"Model parameters: {model.count_parameters():,}")

    # Create replay buffer
    buffer = ReplayBuffer(max_size=config.buffer_size)

    # Create self-play manager
    manager = SelfPlayManager(
        model=model,
        buffer=buffer,
        num_simulations=config.mcts_simulations,
        c_puct=config.c_puct,
        max_moves=config.max_moves_per_game,
        device=device
    )

    # Track training metrics
    all_metrics: List[TrainingMetrics] = []

    # Initialize CSV training log
    csv_log_path = os.path.join(config.checkpoint_dir, "training_log.csv")
    os.makedirs(config.checkpoint_dir, exist_ok=True)
    if start_iteration == 0:
        _init_csv_log(csv_log_path)

    # Main training loop
    for iteration in range(start_iteration, config.num_iterations):
        iteration_start = time.time()

        # 1. Generate self-play games
        model.eval()
        game_stats = manager.generate(
            num_games=config.games_per_iteration,
            verbose=config.verbose_self_play
        )
        positions_generated = sum(s.num_moves for s in game_stats)

        # 2. Train on sampled batches (if buffer has enough examples)
        policy_loss, value_loss, total_loss = 0.0, 0.0, 0.0

        if len(buffer) >= max(config.min_buffer_size, config.batch_size):
            policy_loss, value_loss, total_loss = train_iteration(
                model=model,
                optimizer=optimizer,
                buffer=buffer,
                batch_size=config.batch_size,
                batches_per_iteration=config.batches_per_iteration,
                device=device
            )

            # Step the learning rate scheduler
            scheduler.step()

        # 3. Update manager with trained model
        manager.update_model(model)

        iteration_time = time.time() - iteration_start

        # Compute game statistics for this iteration
        gstats = _compute_game_stats(game_stats)

        # Create metrics for this iteration
        current_lr = optimizer.param_groups[0]["lr"]
        metrics = TrainingMetrics(
            iteration=iteration,
            policy_loss=policy_loss,
            value_loss=value_loss,
            total_loss=total_loss,
            games_played=len(game_stats),
            positions_generated=positions_generated,
            buffer_size=len(buffer),
            learning_rate=current_lr,
            iteration_time=iteration_time,
            avg_game_length=gstats.get("avg_game_length", 0.0),
            white_wins=gstats.get("white_wins", 0),
            black_wins=gstats.get("black_wins", 0),
            draws=gstats.get("draws", 0),
            checkmates=gstats.get("checkmates", 0),
            repetition_draws=gstats.get("repetition_draws", 0),
            stalemates=gstats.get("stalemates", 0),
            max_moves_draws=gstats.get("max_moves_draws", 0),
        )
        all_metrics.append(metrics)

        # Write metrics to CSV log (append each iteration for crash-safety)
        _append_csv_log(csv_log_path, metrics)

        # 4. Log progress
        if (iteration + 1) % config.log_every_n_iterations == 0:
            decisive = metrics.checkmates
            rep = metrics.repetition_draws
            logger.info(
                f"Iteration {iteration + 1}/{config.num_iterations} | "
                f"Loss: {total_loss:.4f} (P: {policy_loss:.4f}, V: {value_loss:.4f}) | "
                f"Games: {len(game_stats)}, Positions: {positions_generated} | "
                f"AvgLen: {metrics.avg_game_length:.0f} | "
                f"W/B/D: {metrics.white_wins}/{metrics.black_wins}/{metrics.draws} | "
                f"Mates: {decisive}, Reps: {rep} | "
                f"Buffer: {len(buffer):,} | LR: {current_lr:.6f} | "
                f"Time: {iteration_time:.1f}s"
            )

        # 5. Save checkpoint at intervals
        should_save = False

        # Check if we hit a predefined checkpoint interval
        if (iteration + 1) in config.checkpoint_intervals:
            should_save = True

        # Check if we should save every N iterations
        if config.save_every_n_iterations > 0:
            if (iteration + 1) % config.save_every_n_iterations == 0:
                should_save = True

        if should_save:
            checkpoint_path = save_checkpoint(
                model=model,
                optimizer=optimizer,
                scheduler=scheduler,
                iteration=iteration + 1,
                config=config,
                checkpoint_dir=config.checkpoint_dir,
                metrics=metrics
            )
            logger.info(f"Saved checkpoint: {checkpoint_path}")

    # Save final checkpoint if not already saved
    final_iteration = config.num_iterations
    if final_iteration not in config.checkpoint_intervals:
        checkpoint_path = save_checkpoint(
            model=model,
            optimizer=optimizer,
            scheduler=scheduler,
            iteration=final_iteration,
            config=config,
            checkpoint_dir=config.checkpoint_dir,
            metrics=all_metrics[-1] if all_metrics else None
        )
        logger.info(f"Saved final checkpoint: {checkpoint_path}")

    logger.info("Training complete!")

    return model


def parse_args() -> argparse.Namespace:
    """Parse command-line arguments."""
    parser = argparse.ArgumentParser(
        description="AlphaZero-style chess training",
        formatter_class=argparse.ArgumentDefaultsHelpFormatter
    )

    # Training loop parameters
    parser.add_argument(
        "--iterations", "-i",
        type=int,
        default=DEFAULT_NUM_ITERATIONS,
        help="Number of training iterations"
    )
    parser.add_argument(
        "--games-per-iter", "-g",
        type=int,
        default=DEFAULT_GAMES_PER_ITERATION,
        help="Number of self-play games per iteration"
    )
    parser.add_argument(
        "--batch-size", "-b",
        type=int,
        default=DEFAULT_BATCH_SIZE,
        help="Training batch size"
    )
    parser.add_argument(
        "--batches-per-iter",
        type=int,
        default=DEFAULT_BATCHES_PER_ITERATION,
        help="Number of training batches per iteration"
    )

    # Replay buffer
    parser.add_argument(
        "--buffer-size",
        type=int,
        default=DEFAULT_BUFFER_SIZE,
        help="Replay buffer size"
    )

    # MCTS parameters
    parser.add_argument(
        "--mcts-sims",
        type=int,
        default=DEFAULT_MCTS_SIMULATIONS,
        help="MCTS simulations per move"
    )
    parser.add_argument(
        "--c-puct",
        type=float,
        default=1.5,
        help="MCTS exploration constant"
    )

    # Optimizer parameters
    parser.add_argument(
        "--lr",
        type=float,
        default=DEFAULT_INITIAL_LR,
        help="Initial learning rate"
    )
    parser.add_argument(
        "--lr-final",
        type=float,
        default=DEFAULT_FINAL_LR,
        help="Final learning rate after decay"
    )
    parser.add_argument(
        "--weight-decay",
        type=float,
        default=DEFAULT_WEIGHT_DECAY,
        help="Weight decay (L2 regularization)"
    )

    # Model architecture
    parser.add_argument(
        "--num-blocks",
        type=int,
        default=DEFAULT_NUM_BLOCKS,
        help="Number of residual blocks"
    )
    parser.add_argument(
        "--num-filters",
        type=int,
        default=DEFAULT_NUM_FILTERS,
        help="Number of convolutional filters"
    )

    # Checkpointing
    parser.add_argument(
        "--checkpoint-dir",
        type=str,
        default=DEFAULT_CHECKPOINT_DIR,
        help="Directory to save checkpoints"
    )
    parser.add_argument(
        "--save-every",
        type=int,
        default=0,
        help="Save checkpoint every N iterations (0=disabled)"
    )
    parser.add_argument(
        "--resume",
        type=str,
        default=None,
        help="Path to checkpoint to resume from"
    )

    # Logging
    parser.add_argument(
        "--log-every",
        type=int,
        default=1,
        help="Log metrics every N iterations"
    )
    parser.add_argument(
        "--quiet", "-q",
        action="store_true",
        help="Reduce output verbosity"
    )
    parser.add_argument(
        "--verbose-self-play",
        action="store_true",
        help="Log per-game self-play progress"
    )

    return parser.parse_args()


def main():
    """Main entry point for training."""
    args = parse_args()

    # Build configuration from arguments
    config = TrainingConfig(
        num_iterations=args.iterations,
        games_per_iteration=args.games_per_iter,
        batch_size=args.batch_size,
        batches_per_iteration=args.batches_per_iter,
        buffer_size=args.buffer_size,
        mcts_simulations=args.mcts_sims,
        c_puct=args.c_puct,
        initial_lr=args.lr,
        final_lr=args.lr_final,
        weight_decay=args.weight_decay,
        num_blocks=args.num_blocks,
        num_filters=args.num_filters,
        checkpoint_dir=args.checkpoint_dir,
        save_every_n_iterations=args.save_every,
        log_every_n_iterations=args.log_every,
        verbose=not args.quiet,
        verbose_self_play=args.verbose_self_play
    )

    # Run training
    train(config, resume_from=args.resume)


if __name__ == "__main__":
    main()
