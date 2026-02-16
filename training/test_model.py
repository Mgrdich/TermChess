"""
Unit Tests for ChessNet Neural Network

This module tests the neural network architecture to ensure:
1. Correct output shapes for policy and value heads
2. Policy logits can be softmaxed to valid probabilities
3. Value output is in the correct range [-1, 1]
4. Model runs on MPS device (Mac ARM)
5. Forward pass works with both random and actual encoded inputs
6. Parameter count is approximately 2M
"""

import chess
import numpy as np
import pytest
import torch
import torch.nn.functional as F

from model import (
    ChessNet,
    ResidualBlock,
    create_model,
    DEFAULT_NUM_BLOCKS,
    DEFAULT_NUM_FILTERS,
    POLICY_OUTPUT_SIZE,
    BOARD_SIZE,
)
from board_encoder import (
    encode_board_tensor,
    get_device,
    NUM_CHANNELS,
)


class TestResidualBlock:
    """Tests for the ResidualBlock class."""

    def test_residual_block_output_shape(self):
        """Residual block should preserve input shape."""
        block = ResidualBlock(num_filters=128)
        x = torch.randn(4, 128, 8, 8)  # [batch, channels, height, width]

        output = block(x)

        assert output.shape == x.shape, (
            f"Expected shape {x.shape}, got {output.shape}"
        )

    def test_residual_block_different_filter_counts(self):
        """Test residual blocks with different filter counts."""
        for num_filters in [32, 64, 128, 256]:
            block = ResidualBlock(num_filters=num_filters)
            x = torch.randn(2, num_filters, 8, 8)

            output = block(x)

            assert output.shape == (2, num_filters, 8, 8)

    def test_residual_block_skip_connection(self):
        """Verify skip connection is functional."""
        block = ResidualBlock(num_filters=128)

        # Set all weights to zero - output should equal ReLU of input due to skip
        with torch.no_grad():
            for param in block.parameters():
                param.zero_()

        x = torch.randn(1, 128, 8, 8)
        output = block(x)

        # With zero weights, conv outputs are zero, so output = ReLU(0 + x) = ReLU(x)
        expected = F.relu(x)
        assert torch.allclose(output, expected), (
            "With zero weights, residual block should output ReLU of input"
        )


class TestChessNetOutputShapes:
    """Tests for verifying output shapes of ChessNet."""

    def test_policy_output_shape_single_position(self):
        """Policy output should have shape [batch, 4096] for single position."""
        model = ChessNet()
        x = torch.randn(1, NUM_CHANNELS, 8, 8)

        policy_logits, _ = model(x)

        assert policy_logits.shape == (1, POLICY_OUTPUT_SIZE), (
            f"Expected shape (1, {POLICY_OUTPUT_SIZE}), got {policy_logits.shape}"
        )

    def test_value_output_shape_single_position(self):
        """Value output should have shape [batch, 1] for single position."""
        model = ChessNet()
        x = torch.randn(1, NUM_CHANNELS, 8, 8)

        _, value = model(x)

        assert value.shape == (1, 1), (
            f"Expected shape (1, 1), got {value.shape}"
        )

    def test_output_shapes_batch(self):
        """Output shapes should scale correctly with batch size."""
        model = ChessNet()

        for batch_size in [1, 4, 16, 32]:
            x = torch.randn(batch_size, NUM_CHANNELS, 8, 8)
            policy_logits, value = model(x)

            assert policy_logits.shape == (batch_size, POLICY_OUTPUT_SIZE), (
                f"Policy shape mismatch for batch_size={batch_size}"
            )
            assert value.shape == (batch_size, 1), (
                f"Value shape mismatch for batch_size={batch_size}"
            )


class TestPolicyOutput:
    """Tests for policy head output properties."""

    def test_policy_softmax_sums_to_one(self):
        """Policy logits should produce valid probability distribution after softmax."""
        model = ChessNet()
        x = torch.randn(4, NUM_CHANNELS, 8, 8)

        policy_logits, _ = model(x)

        # Apply softmax to get probabilities
        policy_probs = F.softmax(policy_logits, dim=1)

        # Sum should be approximately 1 for each batch element
        sums = policy_probs.sum(dim=1)

        assert torch.allclose(sums, torch.ones_like(sums), atol=1e-5), (
            f"Policy probabilities should sum to 1, got {sums}"
        )

    def test_policy_all_probabilities_positive(self):
        """All policy probabilities should be positive after softmax."""
        model = ChessNet()
        x = torch.randn(4, NUM_CHANNELS, 8, 8)

        policy_logits, _ = model(x)
        policy_probs = F.softmax(policy_logits, dim=1)

        assert (policy_probs >= 0).all(), (
            "All policy probabilities should be non-negative"
        )

    def test_policy_all_probabilities_at_most_one(self):
        """All policy probabilities should be at most 1 after softmax."""
        model = ChessNet()
        x = torch.randn(4, NUM_CHANNELS, 8, 8)

        policy_logits, _ = model(x)
        policy_probs = F.softmax(policy_logits, dim=1)

        assert (policy_probs <= 1).all(), (
            "All policy probabilities should be at most 1"
        )

    def test_policy_logits_are_finite(self):
        """Policy logits should not contain NaN or Inf."""
        model = ChessNet()
        x = torch.randn(4, NUM_CHANNELS, 8, 8)

        policy_logits, _ = model(x)

        assert torch.isfinite(policy_logits).all(), (
            "Policy logits should be finite (no NaN or Inf)"
        )


class TestValueOutput:
    """Tests for value head output properties."""

    def test_value_in_valid_range(self):
        """Value output should be in range [-1, 1] due to tanh activation."""
        model = ChessNet()

        # Test with multiple random inputs
        for _ in range(10):
            x = torch.randn(8, NUM_CHANNELS, 8, 8)
            _, value = model(x)

            assert (value >= -1).all() and (value <= 1).all(), (
                f"Value should be in [-1, 1], got min={value.min()}, max={value.max()}"
            )

    def test_value_output_scalar_per_position(self):
        """Each position should produce exactly one scalar value."""
        model = ChessNet()
        batch_size = 5
        x = torch.randn(batch_size, NUM_CHANNELS, 8, 8)

        _, value = model(x)

        assert value.numel() == batch_size, (
            f"Expected {batch_size} values, got {value.numel()}"
        )

    def test_value_is_finite(self):
        """Value output should not contain NaN or Inf."""
        model = ChessNet()
        x = torch.randn(4, NUM_CHANNELS, 8, 8)

        _, value = model(x)

        assert torch.isfinite(value).all(), (
            "Value output should be finite (no NaN or Inf)"
        )


class TestForwardPassWithRandomInput:
    """Tests for forward pass with random tensor input."""

    def test_forward_pass_returns_tuple(self):
        """Forward pass should return a tuple of (policy, value)."""
        model = ChessNet()
        x = torch.randn(1, NUM_CHANNELS, 8, 8)

        result = model(x)

        assert isinstance(result, tuple), "Forward pass should return a tuple"
        assert len(result) == 2, "Forward pass should return exactly 2 elements"

    def test_forward_pass_gradient_flow(self):
        """Gradients should flow through the entire network."""
        model = ChessNet()
        x = torch.randn(2, NUM_CHANNELS, 8, 8, requires_grad=True)

        policy_logits, value = model(x)

        # Compute a dummy loss and backpropagate
        loss = policy_logits.sum() + value.sum()
        loss.backward()

        # Check that input has gradients
        assert x.grad is not None, "Input should have gradients"
        assert x.grad.shape == x.shape, "Gradient shape should match input shape"

    def test_forward_pass_deterministic_in_eval_mode(self):
        """Forward pass should be deterministic in eval mode."""
        model = ChessNet()
        model.eval()

        x = torch.randn(2, NUM_CHANNELS, 8, 8)

        with torch.no_grad():
            policy1, value1 = model(x)
            policy2, value2 = model(x)

        assert torch.equal(policy1, policy2), "Policy should be deterministic in eval mode"
        assert torch.equal(value1, value2), "Value should be deterministic in eval mode"


class TestForwardPassWithEncodedBoard:
    """Tests for forward pass with actual encoded chess positions."""

    def test_forward_pass_with_starting_position(self):
        """Forward pass should work with encoded starting position."""
        model = ChessNet()
        model.eval()

        board = chess.Board()
        device = torch.device("cpu")
        encoded = encode_board_tensor(board, device)

        # Add batch dimension
        x = encoded.unsqueeze(0)

        with torch.no_grad():
            policy_logits, value = model(x)

        assert policy_logits.shape == (1, POLICY_OUTPUT_SIZE)
        assert value.shape == (1, 1)
        assert -1 <= value.item() <= 1

    def test_forward_pass_with_sicilian_defense(self):
        """Forward pass should work with a Sicilian Defense position."""
        model = ChessNet()
        model.eval()

        board = chess.Board()
        board.push_san("e4")
        board.push_san("c5")

        device = torch.device("cpu")
        encoded = encode_board_tensor(board, device)
        x = encoded.unsqueeze(0)

        with torch.no_grad():
            policy_logits, value = model(x)

        # Verify outputs are valid
        policy_probs = F.softmax(policy_logits, dim=1)
        assert torch.allclose(policy_probs.sum(), torch.tensor(1.0), atol=1e-5)
        assert -1 <= value.item() <= 1

    def test_forward_pass_with_batch_of_positions(self):
        """Forward pass should work with a batch of different positions."""
        model = ChessNet()
        model.eval()
        device = torch.device("cpu")

        # Create different positions
        positions = [
            chess.Board(),  # Starting
            chess.Board("rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 2"),  # 1.e4 e5
            chess.Board("r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3"),  # Nf3 Nc6
            chess.Board("r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQkq - 0 1"),  # Both can castle
        ]

        # Encode all positions
        encoded_list = [encode_board_tensor(pos, device) for pos in positions]
        x = torch.stack(encoded_list)

        with torch.no_grad():
            policy_logits, value = model(x)

        assert policy_logits.shape == (4, POLICY_OUTPUT_SIZE)
        assert value.shape == (4, 1)
        assert (value >= -1).all() and (value <= 1).all()

    def test_different_positions_different_outputs(self):
        """Different positions should generally produce different outputs."""
        model = ChessNet()
        model.eval()
        device = torch.device("cpu")

        board1 = chess.Board()  # Starting position
        board2 = chess.Board("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1")  # After 1.e4

        x1 = encode_board_tensor(board1, device).unsqueeze(0)
        x2 = encode_board_tensor(board2, device).unsqueeze(0)

        with torch.no_grad():
            policy1, value1 = model(x1)
            policy2, value2 = model(x2)

        # Outputs should be different (not exactly equal)
        assert not torch.equal(policy1, policy2), (
            "Different positions should produce different policy outputs"
        )


class TestParameterCount:
    """Tests for model parameter count."""

    def test_parameter_count_approximately_2m(self):
        """Model should have approximately 2M parameters.

        With 6 residual blocks and 128 filters, the actual count is ~2.3M.
        This is reasonable for the architecture and within expected bounds.
        """
        model = ChessNet()
        param_count = model.count_parameters()

        # Should be approximately 2-3M parameters for this architecture
        # The 6-block, 128-filter config yields ~2.34M parameters
        lower_bound = 1_500_000  # 1.5M minimum
        upper_bound = 3_000_000  # 3M maximum

        assert lower_bound <= param_count <= upper_bound, (
            f"Expected 1.5M-3M parameters, got {param_count:,}"
        )

        # Print actual count for reference
        print(f"Model has {param_count:,} parameters")

    def test_count_parameters_method(self):
        """count_parameters method should return correct count."""
        model = ChessNet()

        # Manual count
        manual_count = sum(p.numel() for p in model.parameters() if p.requires_grad)

        assert model.count_parameters() == manual_count

    def test_parameter_count_scales_with_filters(self):
        """More filters should mean more parameters."""
        model_small = ChessNet(num_filters=64)
        model_large = ChessNet(num_filters=256)

        assert model_small.count_parameters() < model_large.count_parameters()

    def test_parameter_count_scales_with_blocks(self):
        """More blocks should mean more parameters."""
        model_small = ChessNet(num_blocks=3)
        model_large = ChessNet(num_blocks=12)

        assert model_small.count_parameters() < model_large.count_parameters()


class TestModelConfiguration:
    """Tests for model configuration options."""

    def test_default_configuration(self):
        """Model should use default configuration when not specified."""
        model = ChessNet()

        assert model.num_blocks == DEFAULT_NUM_BLOCKS
        assert model.num_filters == DEFAULT_NUM_FILTERS

    def test_custom_configuration(self):
        """Model should accept custom configuration."""
        model = ChessNet(num_blocks=10, num_filters=256)

        assert model.num_blocks == 10
        assert model.num_filters == 256

    def test_create_model_factory_function(self):
        """create_model should create a properly configured model."""
        model = create_model(num_blocks=4, num_filters=64, device=torch.device("cpu"))

        assert isinstance(model, ChessNet)
        assert model.num_blocks == 4
        assert model.num_filters == 64


class TestDeviceCompatibility:
    """Tests for device compatibility (CPU, MPS)."""

    def test_model_on_cpu(self):
        """Model should work on CPU."""
        device = torch.device("cpu")
        model = ChessNet().to(device)

        x = torch.randn(2, NUM_CHANNELS, 8, 8, device=device)

        policy_logits, value = model(x)

        assert policy_logits.device.type == "cpu"
        assert value.device.type == "cpu"

    def test_create_model_with_cpu(self):
        """create_model should work with explicit CPU device."""
        model = create_model(device=torch.device("cpu"))

        # Check model parameters are on CPU
        for param in model.parameters():
            assert param.device.type == "cpu"


class TestMPSDevice:
    """Tests for MPS (Apple Silicon) device compatibility."""

    @pytest.mark.skipif(
        not torch.backends.mps.is_available(),
        reason="MPS not available on this system"
    )
    def test_model_runs_on_mps(self):
        """Model should run on MPS device (Mac ARM)."""
        device = torch.device("mps")
        model = ChessNet().to(device)

        x = torch.randn(2, NUM_CHANNELS, 8, 8, device=device)

        policy_logits, value = model(x)

        assert policy_logits.device.type == "mps"
        assert value.device.type == "mps"

    @pytest.mark.skipif(
        not torch.backends.mps.is_available(),
        reason="MPS not available on this system"
    )
    def test_mps_output_values_match_cpu(self):
        """MPS outputs should match CPU outputs (within floating point tolerance)."""
        model = ChessNet()
        model.eval()

        # Create random input
        x_cpu = torch.randn(2, NUM_CHANNELS, 8, 8)

        # CPU forward pass
        with torch.no_grad():
            policy_cpu, value_cpu = model(x_cpu)

        # Move model and input to MPS
        device = torch.device("mps")
        model = model.to(device)
        x_mps = x_cpu.to(device)

        # MPS forward pass
        with torch.no_grad():
            policy_mps, value_mps = model(x_mps)

        # Move results back to CPU for comparison
        policy_mps_cpu = policy_mps.cpu()
        value_mps_cpu = value_mps.cpu()

        # Results should be close (allowing for floating point differences)
        assert torch.allclose(policy_cpu, policy_mps_cpu, atol=1e-4), (
            "MPS policy output should match CPU"
        )
        assert torch.allclose(value_cpu, value_mps_cpu, atol=1e-4), (
            "MPS value output should match CPU"
        )

    @pytest.mark.skipif(
        not torch.backends.mps.is_available(),
        reason="MPS not available on this system"
    )
    def test_mps_with_encoded_board(self):
        """Model on MPS should work with encoded chess board."""
        device = torch.device("mps")
        model = ChessNet().to(device)
        model.eval()

        # Encode a chess position
        board = chess.Board()
        encoded = encode_board_tensor(board, device)
        x = encoded.unsqueeze(0)  # Add batch dimension

        with torch.no_grad():
            policy_logits, value = model(x)

        # Verify outputs
        assert policy_logits.shape == (1, POLICY_OUTPUT_SIZE)
        assert value.shape == (1, 1)

        # Move to CPU to check values
        value_cpu = value.cpu()
        assert -1 <= value_cpu.item() <= 1

    @pytest.mark.skipif(
        not torch.backends.mps.is_available(),
        reason="MPS not available on this system"
    )
    def test_create_model_auto_detects_mps(self):
        """create_model should auto-detect and use MPS when available."""
        # When MPS is available, get_device() should return MPS
        device = get_device()
        assert device.type == "mps"

        # create_model should use this device
        model = create_model()

        # Check model parameters are on MPS
        for param in model.parameters():
            assert param.device.type == "mps"

    @pytest.mark.skipif(
        not torch.backends.mps.is_available(),
        reason="MPS not available on this system"
    )
    def test_mps_batch_processing(self):
        """Model on MPS should handle batch processing correctly."""
        device = torch.device("mps")
        model = ChessNet().to(device)
        model.eval()

        # Test various batch sizes
        for batch_size in [1, 4, 8, 16]:
            x = torch.randn(batch_size, NUM_CHANNELS, 8, 8, device=device)

            with torch.no_grad():
                policy_logits, value = model(x)

            assert policy_logits.shape == (batch_size, POLICY_OUTPUT_SIZE)
            assert value.shape == (batch_size, 1)

    @pytest.mark.skipif(
        not torch.backends.mps.is_available(),
        reason="MPS not available on this system"
    )
    def test_mps_training_mode(self):
        """Model should support training on MPS."""
        device = torch.device("mps")
        model = ChessNet().to(device)
        model.train()

        x = torch.randn(4, NUM_CHANNELS, 8, 8, device=device)

        # Forward pass
        policy_logits, value = model(x)

        # Compute dummy loss
        policy_target = torch.zeros(4, POLICY_OUTPUT_SIZE, device=device)
        policy_target[:, 0] = 1  # Set first move as target
        value_target = torch.zeros(4, 1, device=device)

        policy_loss = F.cross_entropy(policy_logits, policy_target)
        value_loss = F.mse_loss(value, value_target)
        total_loss = policy_loss + value_loss

        # Backward pass should work
        total_loss.backward()

        # Check gradients exist
        for param in model.parameters():
            if param.requires_grad:
                assert param.grad is not None, "Parameters should have gradients"


class TestTrainingCapability:
    """Tests to verify model can be trained."""

    def test_model_parameters_require_grad(self):
        """All model parameters should require gradients by default."""
        model = ChessNet()

        for name, param in model.named_parameters():
            assert param.requires_grad, f"Parameter {name} should require gradients"

    def test_optimizer_can_update_parameters(self):
        """Optimizer should be able to update model parameters."""
        model = ChessNet()
        optimizer = torch.optim.Adam(model.parameters(), lr=0.001)

        # Save initial parameters
        initial_params = {name: param.clone() for name, param in model.named_parameters()}

        # Forward pass
        x = torch.randn(2, NUM_CHANNELS, 8, 8)
        policy_logits, value = model(x)

        # Compute loss and update
        loss = policy_logits.sum() + value.sum()
        loss.backward()
        optimizer.step()

        # Check parameters changed
        params_changed = False
        for name, param in model.named_parameters():
            if not torch.equal(param, initial_params[name]):
                params_changed = True
                break

        assert params_changed, "Optimizer should update parameters"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
