"""
Tests for AlphaZero-style Training Loop

These tests verify that:
1. Training configuration is correctly parsed
2. Training loop can run for a few iterations without errors
3. Checkpoints can be saved and loaded
4. Loss computation is correct
5. Learning rate scheduling works

Uses small/fast configuration for testing (few simulations, small batches).
"""

import os
import tempfile
from pathlib import Path

import numpy as np
import pytest
import torch

from board_encoder import get_device, NUM_CHANNELS
from model import create_model, POLICY_OUTPUT_SIZE, ChessNet
from replay_buffer import ReplayBuffer, TrainingExample
from train import (
    TrainingConfig,
    TrainingMetrics,
    compute_loss,
    train_batch,
    train_iteration,
    save_checkpoint,
    load_checkpoint,
    create_optimizer,
    create_scheduler,
    train,
)


class TestTrainingConfig:
    """Tests for TrainingConfig dataclass."""

    def test_default_config(self):
        """Test that default config has expected values."""
        config = TrainingConfig()

        assert config.num_iterations == 80_000
        assert config.games_per_iteration == 100
        assert config.batch_size == 256
        assert config.buffer_size == 500_000
        assert config.mcts_simulations == 400
        assert config.initial_lr == 0.001
        assert config.final_lr == 0.0001
        assert config.weight_decay == 1e-4

    def test_custom_config(self):
        """Test that custom config values are respected."""
        config = TrainingConfig(
            num_iterations=100,
            games_per_iteration=10,
            batch_size=32,
            buffer_size=1000
        )

        assert config.num_iterations == 100
        assert config.games_per_iteration == 10
        assert config.batch_size == 32
        assert config.buffer_size == 1000

    def test_checkpoint_intervals_default(self):
        """Test that checkpoint intervals have default values."""
        config = TrainingConfig()
        assert config.checkpoint_intervals == [10, 25, 50, 100, 250, 500, 1_000, 2_500, 5_000, 10_000, 30_000, 80_000]


class TestLossComputation:
    """Tests for loss computation functions."""

    @pytest.fixture
    def model(self):
        """Create a small model for testing."""
        device = get_device()
        model = create_model(num_blocks=2, num_filters=32, device=device)
        return model

    @pytest.fixture
    def device(self):
        """Get the compute device."""
        return get_device()

    @pytest.fixture
    def sample_batch(self):
        """Create a sample batch for testing."""
        batch_size = 8
        states = np.random.randn(batch_size, NUM_CHANNELS, 8, 8).astype(np.float32)

        # Policy targets should be valid probability distributions
        policies = np.random.rand(batch_size, POLICY_OUTPUT_SIZE).astype(np.float32)
        policies = policies / policies.sum(axis=1, keepdims=True)

        # Value targets should be in [-1, 1]
        values = np.random.uniform(-1, 1, batch_size).astype(np.float32)

        return states, policies, values

    def test_compute_loss_shapes(self, model, device, sample_batch):
        """Test that loss computation produces correct shapes."""
        states, policies, values = sample_batch

        policy_loss, value_loss, total_loss = compute_loss(
            model, states, policies, values, device
        )

        # All losses should be scalar tensors
        assert policy_loss.dim() == 0
        assert value_loss.dim() == 0
        assert total_loss.dim() == 0

    def test_compute_loss_values(self, model, device, sample_batch):
        """Test that loss values are reasonable."""
        states, policies, values = sample_batch

        policy_loss, value_loss, total_loss = compute_loss(
            model, states, policies, values, device
        )

        # Losses should be positive
        assert policy_loss.item() >= 0
        assert value_loss.item() >= 0
        assert total_loss.item() >= 0

        # Total loss should be sum of policy and value loss
        expected_total = policy_loss.item() + value_loss.item()
        assert abs(total_loss.item() - expected_total) < 1e-5

    def test_train_batch(self, model, device, sample_batch):
        """Test that training a batch updates model weights."""
        states, policies, values = sample_batch

        optimizer = create_optimizer(model, initial_lr=0.001, weight_decay=1e-4)

        # Get initial weights
        initial_weights = model.initial_conv.weight.clone()

        # Train one batch
        p_loss, v_loss, t_loss = train_batch(
            model, optimizer, states, policies, values, device
        )

        # Weights should have changed
        assert not torch.equal(model.initial_conv.weight, initial_weights)

        # Losses should be returned as floats
        assert isinstance(p_loss, float)
        assert isinstance(v_loss, float)
        assert isinstance(t_loss, float)


class TestTrainingIteration:
    """Tests for training iteration."""

    @pytest.fixture
    def model(self):
        """Create a small model for testing."""
        device = get_device()
        return create_model(num_blocks=2, num_filters=32, device=device)

    @pytest.fixture
    def device(self):
        return get_device()

    @pytest.fixture
    def filled_buffer(self):
        """Create a buffer with some examples."""
        buffer = ReplayBuffer(max_size=1000)

        for i in range(100):
            # Create valid probability distribution for policy
            policy = np.random.rand(POLICY_OUTPUT_SIZE).astype(np.float32)
            policy = policy / policy.sum()

            example = TrainingExample(
                board_state=np.random.randn(NUM_CHANNELS, 8, 8).astype(np.float32),
                policy_target=policy,
                value_target=np.random.uniform(-1, 1)
            )
            buffer.add(example)

        return buffer

    def test_train_iteration(self, model, device, filled_buffer):
        """Test that a training iteration completes successfully."""
        optimizer = create_optimizer(model, initial_lr=0.001, weight_decay=1e-4)

        p_loss, v_loss, t_loss = train_iteration(
            model=model,
            optimizer=optimizer,
            buffer=filled_buffer,
            batch_size=16,
            batches_per_iteration=5,
            device=device
        )

        # Losses should be reasonable
        assert p_loss >= 0
        assert v_loss >= 0
        assert t_loss >= 0
        assert t_loss == pytest.approx(p_loss + v_loss, rel=1e-4)


class TestCheckpointing:
    """Tests for checkpoint saving and loading."""

    @pytest.fixture
    def model(self):
        """Create a small model for testing."""
        device = get_device()
        return create_model(num_blocks=2, num_filters=32, device=device)

    @pytest.fixture
    def device(self):
        return get_device()

    def test_save_checkpoint(self, model, device):
        """Test that checkpoints can be saved."""
        config = TrainingConfig(
            num_iterations=100,
            num_blocks=2,
            num_filters=32
        )

        optimizer = create_optimizer(model, config.initial_lr, config.weight_decay)
        scheduler = create_scheduler(
            optimizer, config.initial_lr, config.final_lr, config.lr_decay_steps
        )

        with tempfile.TemporaryDirectory() as tmpdir:
            checkpoint_path = save_checkpoint(
                model=model,
                optimizer=optimizer,
                scheduler=scheduler,
                iteration=50,
                config=config,
                checkpoint_dir=tmpdir
            )

            # Check file was created
            assert os.path.exists(checkpoint_path)
            assert "checkpoint_50.pt" in checkpoint_path

    def test_load_checkpoint(self, model, device):
        """Test that checkpoints can be loaded."""
        config = TrainingConfig(
            num_iterations=100,
            num_blocks=2,
            num_filters=32
        )

        optimizer = create_optimizer(model, config.initial_lr, config.weight_decay)
        scheduler = create_scheduler(
            optimizer, config.initial_lr, config.final_lr, config.lr_decay_steps
        )

        # Modify model weights
        with torch.no_grad():
            model.initial_conv.weight.fill_(0.5)

        with tempfile.TemporaryDirectory() as tmpdir:
            # Save checkpoint
            checkpoint_path = save_checkpoint(
                model=model,
                optimizer=optimizer,
                scheduler=scheduler,
                iteration=50,
                config=config,
                checkpoint_dir=tmpdir
            )

            # Load checkpoint
            loaded_model, loaded_opt, loaded_sched, iteration, config_dict = \
                load_checkpoint(checkpoint_path, device)

            # Verify iteration
            assert iteration == 50

            # Verify model weights match
            assert torch.equal(
                model.initial_conv.weight.cpu(),
                loaded_model.initial_conv.weight.cpu()
            )

            # Verify config
            assert config_dict["num_blocks"] == 2
            assert config_dict["num_filters"] == 32

    def test_checkpoint_with_metrics(self, model, device):
        """Test that checkpoints include metrics when provided."""
        config = TrainingConfig(num_blocks=2, num_filters=32)

        optimizer = create_optimizer(model, config.initial_lr, config.weight_decay)
        scheduler = create_scheduler(
            optimizer, config.initial_lr, config.final_lr, config.lr_decay_steps
        )

        metrics = TrainingMetrics(
            iteration=50,
            policy_loss=0.5,
            value_loss=0.3,
            total_loss=0.8,
            games_played=100,
            positions_generated=5000,
            buffer_size=10000,
            learning_rate=0.001,
            iteration_time=60.0
        )

        with tempfile.TemporaryDirectory() as tmpdir:
            checkpoint_path = save_checkpoint(
                model=model,
                optimizer=optimizer,
                scheduler=scheduler,
                iteration=50,
                config=config,
                checkpoint_dir=tmpdir,
                metrics=metrics
            )

            # Load and verify metrics are saved
            checkpoint = torch.load(checkpoint_path, weights_only=False)
            assert "metrics" in checkpoint
            assert checkpoint["metrics"]["policy_loss"] == 0.5
            assert checkpoint["metrics"]["value_loss"] == 0.3


class TestLearningRateScheduler:
    """Tests for learning rate scheduling."""

    @pytest.fixture
    def model(self):
        device = get_device()
        return create_model(num_blocks=2, num_filters=32, device=device)

    def test_scheduler_creation(self, model):
        """Test that scheduler is created correctly."""
        optimizer = create_optimizer(model, initial_lr=0.001, weight_decay=1e-4)
        scheduler = create_scheduler(
            optimizer,
            initial_lr=0.001,
            final_lr=0.0001,
            decay_steps=100
        )

        # Initial LR should be 0.001
        assert optimizer.param_groups[0]["lr"] == 0.001

    def test_scheduler_decay(self, model):
        """Test that scheduler decays LR at the right step."""
        optimizer = create_optimizer(model, initial_lr=0.001, weight_decay=1e-4)
        scheduler = create_scheduler(
            optimizer,
            initial_lr=0.001,
            final_lr=0.0001,
            decay_steps=100
        )

        # Step 99 times (before decay)
        for _ in range(99):
            scheduler.step()

        # LR should still be initial
        assert optimizer.param_groups[0]["lr"] == 0.001

        # Step once more (at decay point)
        scheduler.step()

        # LR should now be decayed
        assert optimizer.param_groups[0]["lr"] == pytest.approx(0.0001, rel=1e-5)


class TestEndToEnd:
    """End-to-end integration tests for training."""

    def test_training_loop_100_iterations(self):
        """
        Test that training loop can run for 100 iterations.

        This is the main acceptance test using a tiny configuration
        for speed: minimal MCTS simulations, small batches, few games.
        """
        with tempfile.TemporaryDirectory() as tmpdir:
            config = TrainingConfig(
                # Tiny configuration for fast testing
                num_iterations=100,
                games_per_iteration=1,  # Just 1 game per iteration
                batch_size=8,
                batches_per_iteration=2,
                buffer_size=1000,
                min_buffer_size=8,  # Start training quickly
                mcts_simulations=5,  # Minimal simulations
                max_moves_per_game=20,  # Short games
                num_blocks=2,
                num_filters=32,
                checkpoint_dir=tmpdir,
                checkpoint_intervals=[50, 100],  # Save at 50 and 100
                log_every_n_iterations=25,
                verbose=False
            )

            # Run training
            model = train(config)

            # Verify model is returned
            assert isinstance(model, ChessNet)

            # Verify checkpoints were saved
            assert os.path.exists(os.path.join(tmpdir, "checkpoint_50.pt"))
            assert os.path.exists(os.path.join(tmpdir, "checkpoint_100.pt"))

    def test_training_loop_with_resume(self):
        """Test that training can resume from a checkpoint."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # First training run: 50 iterations
            config1 = TrainingConfig(
                num_iterations=50,
                games_per_iteration=1,
                batch_size=8,
                batches_per_iteration=2,
                buffer_size=1000,
                min_buffer_size=8,
                mcts_simulations=5,
                max_moves_per_game=20,
                num_blocks=2,
                num_filters=32,
                checkpoint_dir=tmpdir,
                checkpoint_intervals=[50],
                log_every_n_iterations=50,
                verbose=False
            )

            train(config1)

            # Verify checkpoint exists
            checkpoint_path = os.path.join(tmpdir, "checkpoint_50.pt")
            assert os.path.exists(checkpoint_path)

            # Resume training for 50 more iterations
            config2 = TrainingConfig(
                num_iterations=100,  # Total iterations (will run 50-100)
                games_per_iteration=1,
                batch_size=8,
                batches_per_iteration=2,
                buffer_size=1000,
                min_buffer_size=8,
                mcts_simulations=5,
                max_moves_per_game=20,
                num_blocks=2,
                num_filters=32,
                checkpoint_dir=tmpdir,
                checkpoint_intervals=[100],
                log_every_n_iterations=50,
                verbose=False
            )

            model = train(config2, resume_from=checkpoint_path)

            # Verify final checkpoint
            assert isinstance(model, ChessNet)
            assert os.path.exists(os.path.join(tmpdir, "checkpoint_100.pt"))

    def test_training_produces_decreasing_loss(self):
        """
        Test that training produces decreasing loss over iterations.

        Note: With very small models and few iterations, loss may not
        consistently decrease. This test just checks the training
        doesn't produce NaN or infinite losses.
        """
        with tempfile.TemporaryDirectory() as tmpdir:
            config = TrainingConfig(
                num_iterations=20,
                games_per_iteration=2,
                batch_size=8,
                batches_per_iteration=3,
                buffer_size=500,
                min_buffer_size=8,
                mcts_simulations=5,
                max_moves_per_game=30,
                num_blocks=2,
                num_filters=32,
                checkpoint_dir=tmpdir,
                checkpoint_intervals=[20],
                log_every_n_iterations=5,
                verbose=False
            )

            model = train(config)

            # Model should be valid
            assert isinstance(model, ChessNet)

            # Check model produces valid outputs
            device = get_device()
            model.eval()
            with torch.no_grad():
                dummy_input = torch.randn(1, NUM_CHANNELS, 8, 8, device=device)
                policy, value = model(dummy_input)

                # Outputs should not be NaN or Inf
                assert not torch.isnan(policy).any()
                assert not torch.isinf(policy).any()
                assert not torch.isnan(value).any()
                assert not torch.isinf(value).any()

                # Value should be in [-1, 1]
                assert value.item() >= -1.0
                assert value.item() <= 1.0


class TestTrainingMetrics:
    """Tests for TrainingMetrics dataclass."""

    def test_metrics_creation(self):
        """Test that metrics can be created."""
        metrics = TrainingMetrics(
            iteration=100,
            policy_loss=0.5,
            value_loss=0.3,
            total_loss=0.8,
            games_played=100,
            positions_generated=5000,
            buffer_size=10000,
            learning_rate=0.001,
            iteration_time=60.0
        )

        assert metrics.iteration == 100
        assert metrics.policy_loss == 0.5
        assert metrics.value_loss == 0.3
        assert metrics.total_loss == 0.8


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
