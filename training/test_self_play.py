"""
Tests for Self-Play Game Generation and Replay Buffer

These tests verify that:
1. The replay buffer correctly stores and samples training examples
2. Self-play games can be played to completion
3. Training examples are generated correctly
4. Multiple games can be generated end-to-end
"""

import numpy as np
import pytest
import chess
import torch

from board_encoder import encode_board, NUM_CHANNELS, get_device
from model import create_model, POLICY_OUTPUT_SIZE
from mcts import MCTS
from replay_buffer import ReplayBuffer, TrainingExample
from self_play import (
    play_game,
    generate_games,
    generate_games_to_buffer,
    SelfPlayManager,
    _moves_to_policy_vector,
    _get_temperature,
    _get_game_outcome,
    EXPLORATION_MOVES,
    HIGH_TEMPERATURE,
    LOW_TEMPERATURE,
)


class TestReplayBuffer:
    """Tests for the ReplayBuffer class."""

    def test_buffer_initialization(self):
        """Test that buffer initializes with correct size."""
        buffer = ReplayBuffer(max_size=1000)
        assert len(buffer) == 0
        assert buffer.max_size == 1000

    def test_add_single_example(self):
        """Test adding a single example to the buffer."""
        buffer = ReplayBuffer(max_size=100)

        example = TrainingExample(
            board_state=np.zeros((NUM_CHANNELS, 8, 8), dtype=np.float32),
            policy_target=np.zeros(POLICY_OUTPUT_SIZE, dtype=np.float32),
            value_target=1.0
        )

        buffer.add(example)
        assert len(buffer) == 1

    def test_add_batch(self):
        """Test adding multiple examples at once."""
        buffer = ReplayBuffer(max_size=100)

        examples = [
            TrainingExample(
                board_state=np.ones((NUM_CHANNELS, 8, 8), dtype=np.float32) * i,
                policy_target=np.ones(POLICY_OUTPUT_SIZE, dtype=np.float32) * i,
                value_target=float(i) / 10
            )
            for i in range(10)
        ]

        buffer.add_batch(examples)
        assert len(buffer) == 10

    def test_fifo_eviction(self):
        """Test that oldest examples are evicted when buffer is full."""
        buffer = ReplayBuffer(max_size=5)

        # Add 10 examples to a buffer with capacity 5
        for i in range(10):
            example = TrainingExample(
                board_state=np.ones((NUM_CHANNELS, 8, 8), dtype=np.float32) * i,
                policy_target=np.zeros(POLICY_OUTPUT_SIZE, dtype=np.float32),
                value_target=float(i)
            )
            buffer.add(example)

        # Buffer should contain only 5 examples
        assert len(buffer) == 5

        # Sample all and verify we have the newest 5 examples (indices 5-9)
        states, _, values = buffer.sample_all()
        assert len(values) == 5
        # The values should be 5, 6, 7, 8, 9 (but possibly in different order in storage)
        assert set(values) == {5.0, 6.0, 7.0, 8.0, 9.0}

    def test_sample_batch(self):
        """Test sampling a batch of examples."""
        buffer = ReplayBuffer(max_size=100)

        # Add 50 examples
        for i in range(50):
            example = TrainingExample(
                board_state=np.random.randn(NUM_CHANNELS, 8, 8).astype(np.float32),
                policy_target=np.random.randn(POLICY_OUTPUT_SIZE).astype(np.float32),
                value_target=np.random.uniform(-1, 1)
            )
            buffer.add(example)

        # Sample a batch
        states, policies, values = buffer.sample(batch_size=16)

        assert states.shape == (16, NUM_CHANNELS, 8, 8)
        assert policies.shape == (16, POLICY_OUTPUT_SIZE)
        assert values.shape == (16,)

    def test_sample_empty_buffer_raises(self):
        """Test that sampling from empty buffer raises error."""
        buffer = ReplayBuffer(max_size=100)

        with pytest.raises(ValueError, match="empty buffer"):
            buffer.sample(batch_size=10)

    def test_sample_too_large_batch_raises(self):
        """Test that sampling more than buffer size raises error."""
        buffer = ReplayBuffer(max_size=100)

        example = TrainingExample(
            board_state=np.zeros((NUM_CHANNELS, 8, 8), dtype=np.float32),
            policy_target=np.zeros(POLICY_OUTPUT_SIZE, dtype=np.float32),
            value_target=0.0
        )
        buffer.add(example)

        with pytest.raises(ValueError, match="exceeds buffer size"):
            buffer.sample(batch_size=10)

    def test_clear(self):
        """Test clearing the buffer."""
        buffer = ReplayBuffer(max_size=100)

        for i in range(10):
            example = TrainingExample(
                board_state=np.zeros((NUM_CHANNELS, 8, 8), dtype=np.float32),
                policy_target=np.zeros(POLICY_OUTPUT_SIZE, dtype=np.float32),
                value_target=0.0
            )
            buffer.add(example)

        assert len(buffer) == 10
        buffer.clear()
        assert len(buffer) == 0


class TestHelperFunctions:
    """Tests for self-play helper functions."""

    def test_moves_to_policy_vector_shape(self):
        """Test that policy vector has correct shape."""
        board = chess.Board()
        # Create fake visit counts
        visit_counts = {move: 10 for move in list(board.legal_moves)[:5]}

        policy = _moves_to_policy_vector(visit_counts)

        assert policy.shape == (POLICY_OUTPUT_SIZE,)
        assert policy.dtype == np.float32

    def test_moves_to_policy_vector_normalization(self):
        """Test that policy vector sums to 1."""
        board = chess.Board()
        visit_counts = {move: i + 1 for i, move in enumerate(list(board.legal_moves)[:10])}

        policy = _moves_to_policy_vector(visit_counts)

        # Sum should be 1.0 (normalized)
        assert np.isclose(policy.sum(), 1.0)

    def test_moves_to_policy_vector_empty(self):
        """Test policy vector with no visits."""
        policy = _moves_to_policy_vector({})
        assert policy.sum() == 0.0

    def test_get_temperature_early_moves(self):
        """Test that early moves use high temperature."""
        for move_num in range(EXPLORATION_MOVES):
            assert _get_temperature(move_num) == HIGH_TEMPERATURE

    def test_get_temperature_late_moves(self):
        """Test that late moves use low temperature."""
        for move_num in range(EXPLORATION_MOVES, EXPLORATION_MOVES + 10):
            assert _get_temperature(move_num) == LOW_TEMPERATURE

    def test_get_game_outcome_white_wins(self):
        """Test game outcome when white wins."""
        # Scholar's mate position (after Qxf7#)
        board = chess.Board()
        moves = ["e4", "e5", "Bc4", "Nc6", "Qh5", "Nf6", "Qxf7"]
        for move in moves:
            board.push_san(move)

        assert board.is_checkmate()
        # From white's perspective: white won = +1
        assert _get_game_outcome(board, chess.WHITE) == 1.0
        # From black's perspective: white won = -1
        assert _get_game_outcome(board, chess.BLACK) == -1.0

    def test_get_game_outcome_draw(self):
        """Test game outcome for a draw (stalemate)."""
        # A known stalemate position: black king trapped in corner
        # Black to move but has no legal moves (stalemate)
        board = chess.Board("k7/2Q5/1K6/8/8/8/8/8 b - - 0 1")
        assert board.is_stalemate()

        # Draw should return 0.0 for both perspectives
        assert _get_game_outcome(board, chess.WHITE) == 0.0
        assert _get_game_outcome(board, chess.BLACK) == 0.0


class TestSelfPlay:
    """Tests for self-play game generation."""

    @pytest.fixture
    def model(self):
        """Create a small model for testing."""
        device = get_device()
        # Use smaller model for faster tests
        model = create_model(num_blocks=2, num_filters=32, device=device)
        model.eval()
        return model

    @pytest.fixture
    def mcts(self, model):
        """Create MCTS with few simulations for fast testing."""
        return MCTS(
            model=model,
            num_simulations=10,  # Very few for speed
            c_puct=1.5
        )

    def test_play_single_game(self, mcts):
        """Test that a single game can be played to completion."""
        game_data = play_game(mcts, max_moves=50)

        # Game should have some moves
        assert game_data.stats.num_moves > 0

        # Should have generated some training examples
        assert len(game_data.examples) > 0

        # Each example should have correct shapes
        for example in game_data.examples:
            assert example.board_state.shape == (NUM_CHANNELS, 8, 8)
            assert example.policy_target.shape == (POLICY_OUTPUT_SIZE,)
            assert -1.0 <= example.value_target <= 1.0

        # Stats should be populated
        assert game_data.stats.result in ["1-0", "0-1", "1/2-1/2"]
        assert game_data.stats.duration_seconds > 0

    def test_generate_multiple_games(self, model):
        """Test generating multiple games."""
        examples, stats = generate_games(
            model=model,
            num_games=3,
            num_simulations=10,
            max_moves=30,  # Short games for testing
            verbose=False
        )

        # Should have 3 games worth of stats
        assert len(stats) == 3

        # Should have generated training examples
        assert len(examples) > 0

        # All examples should have valid shapes
        for example in examples:
            assert example.board_state.shape == (NUM_CHANNELS, 8, 8)
            assert example.policy_target.shape == (POLICY_OUTPUT_SIZE,)

    def test_generate_games_to_buffer(self, model):
        """Test that games are correctly added to buffer."""
        buffer = ReplayBuffer(max_size=10000)

        stats = generate_games_to_buffer(
            model=model,
            buffer=buffer,
            num_games=2,
            num_simulations=10,
            max_moves=30,
            verbose=False
        )

        # Buffer should have examples
        assert len(buffer) > 0

        # Should be able to sample from buffer
        if len(buffer) >= 16:
            states, policies, values = buffer.sample(batch_size=16)
            assert states.shape[0] == 16

    def test_training_examples_have_correct_values(self, mcts):
        """Test that value targets are set correctly based on game outcome."""
        game_data = play_game(mcts, max_moves=50)

        # Get the game result
        winner = game_data.stats.winner

        # Check that values are consistent with game outcome
        for example in game_data.examples:
            # Value should be one of -1, 0, or 1
            assert example.value_target in [-1.0, 0.0, 1.0]

        # If there's a winner, values should reflect that
        if winner is not None:
            # At least some positions should have non-zero values
            values = [e.value_target for e in game_data.examples]
            assert any(v != 0.0 for v in values)


class TestSelfPlayManager:
    """Tests for the SelfPlayManager class."""

    @pytest.fixture
    def model(self):
        """Create a small model for testing."""
        device = get_device()
        model = create_model(num_blocks=2, num_filters=32, device=device)
        model.eval()
        return model

    def test_manager_initialization(self, model):
        """Test manager initializes correctly."""
        manager = SelfPlayManager(
            model=model,
            num_simulations=10,
            buffer_size=1000
        )

        assert manager.total_games == 0
        assert manager.total_positions == 0
        assert len(manager.buffer) == 0

    def test_manager_generate(self, model):
        """Test manager can generate games."""
        manager = SelfPlayManager(
            model=model,
            num_simulations=10,
            buffer_size=1000
        )

        stats = manager.generate(num_games=2, verbose=False)

        assert len(stats) == 2
        assert manager.total_games == 2
        assert len(manager.buffer) > 0

    def test_manager_sample_batch(self, model):
        """Test manager can sample batches after generation."""
        manager = SelfPlayManager(
            model=model,
            num_simulations=10,
            buffer_size=1000
        )

        manager.generate(num_games=2, verbose=False)

        if len(manager.buffer) >= 8:
            states, policies, values = manager.sample_batch(8)
            assert states.shape[0] == 8
            assert policies.shape[0] == 8
            assert values.shape[0] == 8

    def test_manager_get_statistics(self, model):
        """Test manager statistics collection."""
        manager = SelfPlayManager(
            model=model,
            num_simulations=10,
            buffer_size=1000
        )

        manager.generate(num_games=3, verbose=False)

        stats = manager.get_statistics()

        assert stats["total_games"] == 3
        assert stats["buffer_size"] > 0
        assert "avg_game_length" in stats
        assert "white_wins" in stats
        assert "black_wins" in stats
        assert "draws" in stats

    def test_manager_update_model(self, model):
        """Test that model can be updated."""
        manager = SelfPlayManager(
            model=model,
            num_simulations=10,
            buffer_size=1000
        )

        # Create a new model
        device = get_device()
        new_model = create_model(num_blocks=2, num_filters=32, device=device)
        new_model.eval()

        manager.update_model(new_model)
        assert manager.model is new_model


class TestEndToEnd:
    """End-to-end integration tests."""

    def test_full_pipeline_10_games(self):
        """
        Test that 10 complete self-play games can be generated.

        This is the main acceptance test verifying the full pipeline works.
        Uses minimal MCTS simulations for speed.
        """
        device = get_device()

        # Create model
        model = create_model(num_blocks=2, num_filters=32, device=device)
        model.eval()

        # Create buffer
        buffer = ReplayBuffer(max_size=10000)

        # Generate 10 games
        stats = generate_games_to_buffer(
            model=model,
            buffer=buffer,
            num_games=10,
            num_simulations=10,  # Minimal for speed
            max_moves=50,  # Cap game length
            verbose=False
        )

        # Verify we got 10 games
        assert len(stats) == 10

        # Verify buffer has examples
        assert len(buffer) > 0

        # Verify we can sample training batches
        if len(buffer) >= 32:
            states, policies, values = buffer.sample(batch_size=32)

            # Verify shapes
            assert states.shape == (32, NUM_CHANNELS, 8, 8)
            assert policies.shape == (32, POLICY_OUTPUT_SIZE)
            assert values.shape == (32,)

            # Verify data types
            assert states.dtype == np.float32
            assert policies.dtype == np.float32
            assert values.dtype == np.float32

            # Verify value range
            assert np.all(values >= -1.0)
            assert np.all(values <= 1.0)

        # Print summary for debugging
        print(f"\nGenerated {len(stats)} games with {len(buffer)} positions")
        print(f"Game lengths: {[s.num_moves for s in stats]}")
        print(f"Results: {[s.result for s in stats]}")


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
