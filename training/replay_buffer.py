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
        self, batch_size: int
    ) -> Tuple[np.ndarray, np.ndarray, np.ndarray]:
        """
        Sample a random batch of training examples.

        Args:
            batch_size: Number of examples to sample

        Returns:
            Tuple of:
            - board_states: numpy array of shape [batch_size, 18, 8, 8]
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

        return (
            self.board_states[indices].copy(),
            self.policy_targets[indices].copy(),
            self.value_targets[indices].copy(),
        )

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

    def clear(self) -> None:
        """Clear all examples from the buffer."""
        self.size = 0
        self.index = 0
        # Keep arrays allocated for reuse
