"""
Unit Tests for ONNX Export

This module tests the ONNX export functionality to ensure:
1. Export creates valid ONNX file
2. ONNX model loads in onnxruntime
3. Outputs match between PyTorch and ONNX
4. Various batch sizes work correctly
5. Model info extraction works correctly
"""

import os

import numpy as np
import onnx
import onnxruntime as ort
import pytest
import torch

from board_encoder import NUM_CHANNELS
from export_onnx import (
    DEFAULT_OPSET_VERSION,
    INPUT_NAME,
    OUTPUT_POLICY_NAME,
    OUTPUT_VALUE_NAME,
    compare_outputs,
    export_checkpoint,
    export_to_onnx,
    get_model_info,
    load_checkpoint_for_export,
    verify_onnx_model,
)
from model import ChessNet
from train import TrainingConfig, save_checkpoint


@pytest.fixture
def model():
    """Create a ChessNet model for testing."""
    return ChessNet(num_blocks=2, num_filters=32)  # Smaller model for faster tests


@pytest.fixture
def checkpoint_path(model, tmp_path):
    """Create a temporary checkpoint file."""
    from torch.optim import Adam
    from torch.optim.lr_scheduler import StepLR

    # Create optimizer and scheduler
    optimizer = Adam(model.parameters(), lr=0.001)
    scheduler = StepLR(optimizer, step_size=1000, gamma=0.1)

    # Create config
    config = TrainingConfig(
        num_blocks=2,
        num_filters=32,
        num_iterations=100,
    )

    # Save checkpoint (written to checkpoint_dir as checkpoint_100.pt)
    save_checkpoint(
        model=model,
        optimizer=optimizer,
        scheduler=scheduler,
        iteration=100,
        config=config,
        checkpoint_dir=str(tmp_path),
    )

    return str(tmp_path / "checkpoint_100.pt")


class TestExportToOnnx:
    """Tests for export_to_onnx function."""

    def test_export_creates_file(self, model, tmp_path):
        """Export should create an ONNX file."""
        output_path = str(tmp_path / "model.onnx")

        export_to_onnx(model, output_path)

        assert os.path.exists(output_path)

    def test_export_creates_valid_onnx(self, model, tmp_path):
        """Exported file should be valid ONNX."""
        output_path = str(tmp_path / "model.onnx")

        export_to_onnx(model, output_path)

        # Load and check the model
        onnx_model = onnx.load(output_path)
        onnx.checker.check_model(onnx_model)

    def test_export_with_custom_opset(self, model, tmp_path):
        """Export should work with different opset versions.

        Note: PyTorch's ONNX exporter may upgrade lower opset versions to a
        minimum supported version (currently 18). We test that a higher opset
        version is respected.
        """
        output_path = str(tmp_path / "model.onnx")

        # Use opset 18 which is well-supported by current PyTorch
        export_to_onnx(model, output_path, opset_version=18)

        onnx_model = onnx.load(output_path)
        assert onnx_model.opset_import[0].version == 18

    def test_export_input_name(self, model, tmp_path):
        """Exported model should have correct input name."""
        output_path = str(tmp_path / "model.onnx")

        export_to_onnx(model, output_path)

        onnx_model = onnx.load(output_path)
        input_names = [inp.name for inp in onnx_model.graph.input]
        assert INPUT_NAME in input_names

    def test_export_output_names(self, model, tmp_path):
        """Exported model should have correct output names."""
        output_path = str(tmp_path / "model.onnx")

        export_to_onnx(model, output_path)

        onnx_model = onnx.load(output_path)
        output_names = [out.name for out in onnx_model.graph.output]
        assert OUTPUT_POLICY_NAME in output_names
        assert OUTPUT_VALUE_NAME in output_names


class TestVerifyOnnxModel:
    """Tests for verify_onnx_model function."""

    def test_verify_valid_model(self, model, tmp_path):
        """Verification should pass for valid model."""
        output_path = str(tmp_path / "model.onnx")
        export_to_onnx(model, output_path)

        onnx_model = verify_onnx_model(output_path)

        assert onnx_model is not None

    def test_verify_nonexistent_file(self):
        """Verification should raise error for nonexistent file."""
        with pytest.raises(FileNotFoundError):
            verify_onnx_model("/nonexistent/path/model.onnx")


class TestOnnxRuntimeLoading:
    """Tests for loading ONNX model in onnxruntime."""

    def test_onnx_loads_in_onnxruntime(self, model, tmp_path):
        """ONNX model should load in onnxruntime."""
        output_path = str(tmp_path / "model.onnx")
        export_to_onnx(model, output_path)

        session = ort.InferenceSession(output_path, providers=["CPUExecutionProvider"])

        assert session is not None

    def test_onnx_inference_runs(self, model, tmp_path):
        """ONNX model should run inference."""
        output_path = str(tmp_path / "model.onnx")
        export_to_onnx(model, output_path)

        session = ort.InferenceSession(output_path, providers=["CPUExecutionProvider"])

        # Create test input
        test_input = np.random.randn(1, NUM_CHANNELS, 8, 8).astype(np.float32)

        # Run inference
        outputs = session.run(None, {INPUT_NAME: test_input})

        assert len(outputs) == 2

    def test_onnx_output_shapes(self, model, tmp_path):
        """ONNX outputs should have correct shapes."""
        output_path = str(tmp_path / "model.onnx")
        export_to_onnx(model, output_path)

        session = ort.InferenceSession(output_path, providers=["CPUExecutionProvider"])

        # Test with batch size 1
        test_input = np.random.randn(1, NUM_CHANNELS, 8, 8).astype(np.float32)
        outputs = session.run(None, {INPUT_NAME: test_input})

        policy, value = outputs
        assert policy.shape == (1, 4096), f"Expected (1, 4096), got {policy.shape}"
        assert value.shape == (1, 1), f"Expected (1, 1), got {value.shape}"


class TestOutputComparison:
    """Tests for comparing PyTorch and ONNX outputs."""

    def test_outputs_match_single_batch(self, model, tmp_path):
        """PyTorch and ONNX outputs should match for batch size 1."""
        output_path = str(tmp_path / "model.onnx")
        model.eval()
        export_to_onnx(model, output_path)

        test_input = torch.randn(1, NUM_CHANNELS, 8, 8)
        success, message = compare_outputs(model, output_path, test_input)

        assert success, message

    def test_outputs_match_various_batch_sizes(self, model, tmp_path):
        """PyTorch and ONNX outputs should match for various batch sizes."""
        output_path = str(tmp_path / "model.onnx")
        model.eval()
        export_to_onnx(model, output_path)

        for batch_size in [1, 2, 4, 8, 16, 32]:
            test_input = torch.randn(batch_size, NUM_CHANNELS, 8, 8)
            success, message = compare_outputs(model, output_path, test_input)

            assert success, f"Batch size {batch_size}: {message}"

    def test_outputs_match_with_extreme_values(self, model, tmp_path):
        """Outputs should match even with extreme input values."""
        output_path = str(tmp_path / "model.onnx")
        model.eval()
        export_to_onnx(model, output_path)

        # Test with zeros
        test_input = torch.zeros(2, NUM_CHANNELS, 8, 8)
        success, message = compare_outputs(model, output_path, test_input)
        assert success, f"Zero input: {message}"

        # Test with ones
        test_input = torch.ones(2, NUM_CHANNELS, 8, 8)
        success, message = compare_outputs(model, output_path, test_input)
        assert success, f"Ones input: {message}"

    def test_value_output_range(self, model, tmp_path):
        """ONNX value output should be in range [-1, 1]."""
        output_path = str(tmp_path / "model.onnx")
        model.eval()
        export_to_onnx(model, output_path)

        session = ort.InferenceSession(output_path, providers=["CPUExecutionProvider"])

        # Test multiple random inputs
        for _ in range(10):
            test_input = np.random.randn(8, NUM_CHANNELS, 8, 8).astype(np.float32)
            outputs = session.run(None, {INPUT_NAME: test_input})
            value = outputs[1]

            assert np.all(value >= -1), f"Value below -1: {value.min()}"
            assert np.all(value <= 1), f"Value above 1: {value.max()}"


class TestModelInfo:
    """Tests for get_model_info function."""

    def test_get_model_info_returns_dict(self, model, tmp_path):
        """get_model_info should return a dictionary."""
        output_path = str(tmp_path / "model.onnx")
        export_to_onnx(model, output_path)

        info = get_model_info(output_path)

        assert isinstance(info, dict)

    def test_model_info_contains_file_size(self, model, tmp_path):
        """Model info should contain file size."""
        output_path = str(tmp_path / "model.onnx")
        export_to_onnx(model, output_path)

        info = get_model_info(output_path)

        assert "file_size_bytes" in info
        assert "file_size_mb" in info
        assert info["file_size_bytes"] > 0

    def test_model_info_contains_opset(self, model, tmp_path):
        """Model info should contain opset version."""
        output_path = str(tmp_path / "model.onnx")
        export_to_onnx(model, output_path, opset_version=DEFAULT_OPSET_VERSION)

        info = get_model_info(output_path)

        assert "opset_version" in info
        assert info["opset_version"] == DEFAULT_OPSET_VERSION

    def test_model_info_contains_inputs(self, model, tmp_path):
        """Model info should contain input information."""
        output_path = str(tmp_path / "model.onnx")
        export_to_onnx(model, output_path)

        info = get_model_info(output_path)

        assert "inputs" in info
        assert len(info["inputs"]) == 1
        assert info["inputs"][0]["name"] == INPUT_NAME

    def test_model_info_contains_outputs(self, model, tmp_path):
        """Model info should contain output information."""
        output_path = str(tmp_path / "model.onnx")
        export_to_onnx(model, output_path)

        info = get_model_info(output_path)

        assert "outputs" in info
        assert len(info["outputs"]) == 2


class TestLoadCheckpoint:
    """Tests for load_checkpoint_for_export function."""

    def test_load_checkpoint_returns_model(self, checkpoint_path):
        """Loading checkpoint should return a ChessNet model."""
        model = load_checkpoint_for_export(checkpoint_path)

        assert isinstance(model, ChessNet)

    def test_load_checkpoint_model_in_eval_mode(self, checkpoint_path):
        """Loaded model should be in eval mode."""
        model = load_checkpoint_for_export(checkpoint_path)

        assert not model.training

    def test_load_checkpoint_preserves_architecture(self, checkpoint_path):
        """Loaded model should have correct architecture."""
        model = load_checkpoint_for_export(checkpoint_path)

        assert model.num_blocks == 2
        assert model.num_filters == 32

    def test_load_checkpoint_nonexistent_raises(self):
        """Loading nonexistent checkpoint should raise error."""
        with pytest.raises(FileNotFoundError):
            load_checkpoint_for_export("/nonexistent/checkpoint.pt")


class TestExportCheckpoint:
    """Tests for export_checkpoint function (end-to-end)."""

    def test_export_checkpoint_success(self, checkpoint_path, tmp_path):
        """export_checkpoint should succeed for valid checkpoint."""
        output_path = str(tmp_path / "exported.onnx")

        success = export_checkpoint(
            checkpoint_path=checkpoint_path, output_path=output_path, verify=True, verbose=False
        )

        assert success
        assert os.path.exists(output_path)

    def test_export_checkpoint_creates_valid_onnx(self, checkpoint_path, tmp_path):
        """export_checkpoint should create valid ONNX file."""
        output_path = str(tmp_path / "exported.onnx")

        export_checkpoint(checkpoint_path=checkpoint_path, output_path=output_path, verify=False, verbose=False)

        # Verify the file
        onnx_model = onnx.load(output_path)
        onnx.checker.check_model(onnx_model)

    def test_export_checkpoint_without_verification(self, checkpoint_path, tmp_path):
        """export_checkpoint should work without verification."""
        output_path = str(tmp_path / "exported.onnx")

        success = export_checkpoint(
            checkpoint_path=checkpoint_path, output_path=output_path, verify=False, verbose=False
        )

        assert success


class TestDynamicBatchSize:
    """Tests for dynamic batch size support."""

    def test_dynamic_batch_sizes(self, model, tmp_path):
        """ONNX model should support various batch sizes."""
        output_path = str(tmp_path / "model.onnx")
        export_to_onnx(model, output_path)

        session = ort.InferenceSession(output_path, providers=["CPUExecutionProvider"])

        # Test various batch sizes
        for batch_size in [1, 5, 10, 25, 50, 100]:
            test_input = np.random.randn(batch_size, NUM_CHANNELS, 8, 8).astype(np.float32)
            outputs = session.run(None, {INPUT_NAME: test_input})

            policy, value = outputs
            assert policy.shape == (batch_size, 4096)
            assert value.shape == (batch_size, 1)


class TestDefaultArchitecture:
    """Tests with default model architecture (6 blocks, 128 filters)."""

    def test_default_model_export(self, tmp_path):
        """Default architecture model should export correctly."""
        model = ChessNet()  # Default: 6 blocks, 128 filters
        output_path = str(tmp_path / "model.onnx")

        export_to_onnx(model, output_path)

        assert os.path.exists(output_path)

        # Verify the model
        onnx_model = onnx.load(output_path)
        onnx.checker.check_model(onnx_model)

    def test_default_model_outputs_match(self, tmp_path):
        """Default architecture should have matching outputs."""
        model = ChessNet()
        model.eval()
        output_path = str(tmp_path / "model.onnx")

        export_to_onnx(model, output_path)

        test_input = torch.randn(4, NUM_CHANNELS, 8, 8)
        success, message = compare_outputs(model, output_path, test_input)

        assert success, message


class TestReproducibility:
    """Tests for export reproducibility."""

    def test_deterministic_export(self, model, tmp_path):
        """Same model should produce identical ONNX outputs."""
        model.eval()

        path1 = str(tmp_path / "model1.onnx")
        path2 = str(tmp_path / "model2.onnx")

        export_to_onnx(model, path1)
        export_to_onnx(model, path2)

        # Load both sessions
        session1 = ort.InferenceSession(path1, providers=["CPUExecutionProvider"])
        session2 = ort.InferenceSession(path2, providers=["CPUExecutionProvider"])

        # Compare outputs
        test_input = np.random.randn(2, NUM_CHANNELS, 8, 8).astype(np.float32)

        outputs1 = session1.run(None, {INPUT_NAME: test_input})
        outputs2 = session2.run(None, {INPUT_NAME: test_input})

        assert np.allclose(outputs1[0], outputs2[0])
        assert np.allclose(outputs1[1], outputs2[1])


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
