"""
Replay Buffer for AlphaZero-style Chess Training

This module implements a replay buffer for storing and sampling training examples
generated during self-play. The buffer supports:

1. FIFO eviction when capacity is exceeded
2. Random sampling for training batches
3. Efficient numpy-based storage

Training Example Format:
------------------------
Each training example contains:
- board_state: numpy array [18, 8, 8] from board encoder
- policy_target: numpy array [4096] from MCTS visit counts (normalized)
- value_target: float in [-1, 1] (game outcome from this player's perspective)

The value_target is determined at the END of the game:
- +1.0 for positions where this player won
- -1.0 for positions where this player lost
- 0.0 for draws
"""

from dataclasses import dataclass
from typing import List, Optional, Tuple
import numpy as np


# Precomputed mapping for horizontal flip of policy indices.
# Policy index = from_square * 64 + to_square
# Flip mirrors files: file' = 7 - file, rank stays the same.
# flip(square) = rank * 8 + (7 - file)
def _build_flip_map() -> np.ndarray:
    """Build a lookup table mapping policy indices to their horizontally flipped indices."""
    flip_map = np.zeros(4096, dtype=np.int32)
    for from_sq in range(64):
        from_rank, from_file = divmod(from_sq, 8)
        flipped_from = from_rank * 8 + (7 - from_file)
        for to_sq in range(64):
            to_rank, to_file = divmod(to_sq, 8)
            flipped_to = to_rank * 8 + (7 - to_file)
            old_idx = from_sq * 64 + to_sq
            new_idx = flipped_from * 64 + flipped_to
            flip_map[old_idx] = new_idx
    return flip_map


_POLICY_FLIP_MAP = _build_flip_map()

# Castling channel pairs that need swapping on horizontal flip:
# Kingside <-> Queenside for each color.
# These are offsets from the start of the current position's metadata channels.
# In the base encoding: channels 13-14 (WK, WQ), channels 15-16 (BK, BQ).
CASTLING_CHANNEL_PAIRS = [(13, 14), (15, 16)]


def flip_board_horizontal(
    board_state: np.ndarray, policy_target: np.ndarray
) -> tuple:
    """
    Apply horizontal (file) flip augmentation to a training example.

    Mirrors the board along the vertical axis (a-file <-> h-file).
    This is a valid symmetry for chess positions.

    Args:
        board_state: Encoded board of shape [C, 8, 8]
        policy_target: Policy vector of shape [4096]

    Returns:
        Tuple of (flipped_board_state, flipped_policy_target)
    """
    # Flip all channels along file axis (axis 2 = last dimension)
    flipped_state = board_state[:, :, ::-1].copy()

    # Swap castling channel pairs (kingside <-> queenside)
    for ch_a, ch_b in CASTLING_CHANNEL_PAIRS:
        if ch_a < flipped_state.shape[0] and ch_b < flipped_state.shape[0]:
            tmp = flipped_state[ch_a].copy()
            flipped_state[ch_a] = flipped_state[ch_b]
            flipped_state[ch_b] = tmp

    # Remap policy indices using precomputed flip map
    flipped_policy = np.zeros_like(policy_target)
    flipped_policy[_POLICY_FLIP_MAP] = policy_target

    return flipped_state, flipped_policy


@dataclass
class TrainingExample:
    """
    A single training example from self-play.

    Attributes:
        board_state: Encoded board position of shape [18, 8, 8]
        policy_target: MCTS visit count distribution of shape [4096]
        value_target: Game outcome from current player's perspective [-1, 1]
    """
    board_state: np.ndarray
    policy_target: np.ndarray
    value_target: float


class ReplayBuffer:
    """
    Replay buffer for storing training examples from self-play games.

    The buffer stores training examples and supports:
    - Adding new examples with automatic FIFO eviction when full
    - Random sampling of batches for training
    - Efficient numpy-based storage

    The default capacity is 500,000 positions as specified in the technical
    requirements, but can be configured.

    Attributes:
        max_size: Maximum number of examples to store
        board_states: numpy array of board states [N, 18, 8, 8]
        policy_targets: numpy array of policy targets [N, 4096]
        value_targets: numpy array of value targets [N]
        size: Current number of examples in the buffer
        index: Next write position (circular buffer)

    Example:
        >>> buffer = ReplayBuffer(max_size=1000)
        >>> example = TrainingExample(
        ...     board_state=np.zeros((18, 8, 8), dtype=np.float32),
        ...     policy_target=np.zeros(4096, dtype=np.float32),
        ...     value_target=1.0
        ... )
        >>> buffer.add(example)
        >>> states, policies, values = buffer.sample(batch_size=32)
    """

    def __init__(self, max_size: int = 500_000):
        """
        Initialize the replay buffer.

        Args:
            max_size: Maximum number of examples to store (default: 500,000)
        """
        self.max_size = max_size

        # Pre-allocate numpy arrays for efficient storage
        # These will be lazily initialized on first add
        self.board_states: Optional[np.ndarray] = None
        self.policy_targets: Optional[np.ndarray] = None
        self.value_targets: Optional[np.ndarray] = None

        # Circular buffer tracking
        self.size = 0  # Current number of valid examples
        self.index = 0  # Next write position

    def _initialize_arrays(self, example: TrainingExample) -> None:
        """
        Initialize storage arrays based on the shape of the first example.

        Args:
            example: First training example to infer shapes from
        """
        board_shape = example.board_state.shape
        policy_shape = example.policy_target.shape

        self.board_states = np.zeros(
            (self.max_size,) + board_shape, dtype=np.float32
        )
        self.policy_targets = np.zeros(
            (self.max_size,) + policy_shape, dtype=np.float32
        )
        self.value_targets = np.zeros(self.max_size, dtype=np.float32)

    def add(self, example: TrainingExample) -> None:
        """
        Add a single training example to the buffer.

        If the buffer is full, the oldest example is overwritten (FIFO eviction).

        Args:
            example: Training example to add
        """
        # Initialize arrays on first add
        if self.board_states is None:
            self._initialize_arrays(example)

        # Store the example at the current index
        self.board_states[self.index] = example.board_state
        self.policy_targets[self.index] = example.policy_target
        self.value_targets[self.index] = example.value_target

        # Update circular buffer pointers
        self.index = (self.index + 1) % self.max_size
        self.size = min(self.size + 1, self.max_size)

    def add_batch(self, examples: List[TrainingExample]) -> None:
        """
        Add multiple training examples to the buffer.

        This is a convenience method that calls add() for each example.

        Args:
            examples: List of training examples to add
        """
        for example in examples:
            self.add(example)

    def sample(
        self, batch_size: int, augment: bool = True
    ) -> Tuple[np.ndarray, np.ndarray, np.ndarray]:
        """
        Sample a random batch of training examples.

        When augment=True, each example has a 50% chance of being horizontally
        flipped (mirroring files a-h to h-a), effectively doubling the
        training data diversity at no storage cost.

        Args:
            batch_size: Number of examples to sample
            augment: Whether to apply random horizontal flip (default: True)

        Returns:
            Tuple of:
            - board_states: numpy array of shape [batch_size, C, 8, 8]
            - policy_targets: numpy array of shape [batch_size, 4096]
            - value_targets: numpy array of shape [batch_size]

        Raises:
            ValueError: If buffer is empty or batch_size exceeds buffer size
        """
        if self.size == 0:
            raise ValueError("Cannot sample from empty buffer")

        if batch_size > self.size:
            raise ValueError(
                f"batch_size ({batch_size}) exceeds buffer size ({self.size})"
            )

        # Sample random indices without replacement
        indices = np.random.choice(self.size, size=batch_size, replace=False)

        states = self.board_states[indices].copy()
        policies = self.policy_targets[indices].copy()
        values = self.value_targets[indices].copy()

        if augment:
            # Apply horizontal flip to ~50% of examples
            flip_mask = np.random.random(batch_size) < 0.5
            for i in np.where(flip_mask)[0]:
                states[i], policies[i] = flip_board_horizontal(
                    states[i], policies[i]
                )

        return states, policies, values

    def sample_all(self) -> Tuple[np.ndarray, np.ndarray, np.ndarray]:
        """
        Return all examples in the buffer.

        This is useful for small buffers or when you need all data.

        Returns:
            Tuple of:
            - board_states: numpy array of shape [size, 18, 8, 8]
            - policy_targets: numpy array of shape [size, 4096]
            - value_targets: numpy array of shape [size]

        Raises:
            ValueError: If buffer is empty
        """
        if self.size == 0:
            raise ValueError("Cannot sample from empty buffer")

        return (
            self.board_states[:self.size].copy(),
            self.policy_targets[:self.size].copy(),
            self.value_targets[:self.size].copy(),
        )

    def __len__(self) -> int:
        """Return the current number of examples in the buffer."""
        return self.size

    def save(self, path: str) -> None:
        """
        Save buffer contents to a .npz file.

        Args:
            path: File path to save to (should end in .npz)
        """
        if self.size == 0 or self.board_states is None:
            return
        np.savez_compressed(
            path,
            board_states=self.board_states[:self.size],
            policy_targets=self.policy_targets[:self.size],
            value_targets=self.value_targets[:self.size],
            size=np.array([self.size]),
            index=np.array([self.index]),
        )

    def load(self, path: str) -> None:
        """
        Load buffer contents from a .npz file.

        Replaces current buffer contents. The max_size is preserved —
        if the saved data exceeds max_size, only the most recent entries
        are kept.

        Args:
            path: File path to load from
        """
        data = np.load(path)
        saved_states = data["board_states"]
        saved_policies = data["policy_targets"]
        saved_values = data["value_targets"]
        saved_size = int(data["size"][0])
        saved_index = int(data["index"][0])

        # Initialize arrays from loaded shapes if needed
        if self.board_states is None:
            board_shape = saved_states.shape[1:]
            policy_shape = saved_policies.shape[1:]
            self.board_states = np.zeros(
                (self.max_size,) + board_shape, dtype=np.float32
            )
            self.policy_targets = np.zeros(
                (self.max_size,) + policy_shape, dtype=np.float32
            )
            self.value_targets = np.zeros(self.max_size, dtype=np.float32)

        # Copy data (truncate if saved data exceeds max_size)
        n = min(saved_size, self.max_size)
        self.board_states[:n] = saved_states[:n]
        self.policy_targets[:n] = saved_policies[:n]
        self.value_targets[:n] = saved_values[:n]
        self.size = n
        self.index = min(saved_index, n) % self.max_size

    def clear(self) -> None:
        """Clear all examples from the buffer."""
        self.size = 0
        self.index = 0
        # Keep arrays allocated for reuse
