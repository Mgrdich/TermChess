"""
ONNX Export for ChessNet Neural Network

This module exports trained PyTorch ChessNet models to ONNX format for inference
in Go or other languages using ONNX Runtime.

The exported model:
- Input: board_state [batch, 18, 8, 8] float32 tensor
- Output: policy_logits [batch, 4096] float32
- Output: value [batch, 1] float32

Usage:
    uv run python export_onnx.py checkpoint.pt output.onnx

The script will:
1. Load the PyTorch checkpoint
2. Export to ONNX format with dynamic batch size
3. Verify the ONNX model loads in onnxruntime
4. Compare PyTorch vs ONNX outputs to ensure correctness
5. Print model information (size, input/output shapes)
"""

import argparse
import os
import sys

import numpy as np
import onnx
import onnxruntime as ort
import torch

from board_encoder import NUM_CHANNELS
from model import ChessNet

# Default ONNX opset version (18 is the minimum supported by PyTorch 2.x ONNX exporter)
DEFAULT_OPSET_VERSION = 18

# Input/output names for ONNX model
INPUT_NAME = "board_state"
OUTPUT_POLICY_NAME = "policy_logits"
OUTPUT_VALUE_NAME = "value"


def load_checkpoint_for_export(
    checkpoint_path: str,
    device: torch.device | None = None,
) -> ChessNet:
    """
    Load a PyTorch checkpoint for ONNX export.

    Args:
        checkpoint_path: Path to the checkpoint file
        device: Device to load the model onto (CPU recommended for export)

    Returns:
        Loaded ChessNet model in eval mode

    Raises:
        FileNotFoundError: If checkpoint file does not exist
        KeyError: If checkpoint is missing required keys
    """
    if device is None:
        device = torch.device("cpu")

    if not os.path.exists(checkpoint_path):
        raise FileNotFoundError(f"Checkpoint not found: {checkpoint_path}")

    # Load checkpoint
    checkpoint = torch.load(checkpoint_path, map_location=device, weights_only=False)

    # Extract model configuration
    config = checkpoint.get("config", {})
    num_blocks = config.get("num_blocks", 6)
    num_filters = config.get("num_filters", 128)

    # Create model with same architecture
    model = ChessNet(num_blocks=num_blocks, num_filters=num_filters)
    model.load_state_dict(checkpoint["model_state_dict"])
    model = model.to(device)
    model.eval()

    return model


def export_to_onnx(model: ChessNet, output_path: str, opset_version: int = DEFAULT_OPSET_VERSION) -> None:
    """
    Export a ChessNet model to ONNX format.

    The exported model supports dynamic batch sizes and includes:
    - Input: board_state [batch, 18, 8, 8]
    - Output: policy_logits [batch, 4096]
    - Output: value [batch, 1]

    Args:
        model: ChessNet model to export (should be in eval mode)
        output_path: Path for the output ONNX file
        opset_version: ONNX opset version to use (default: 18)

    Raises:
        RuntimeError: If export fails
    """
    # Ensure model is on CPU for ONNX export
    model = model.cpu()
    model.eval()

    # Create dummy input for tracing (CPU to match model)
    # Shape: [batch, channels, height, width] = [1, 18, 8, 8]
    dummy_input = torch.randn(1, NUM_CHANNELS, 8, 8)

    # Export to ONNX
    torch.onnx.export(
        model,
        (dummy_input,),
        output_path,
        input_names=[INPUT_NAME],
        output_names=[OUTPUT_POLICY_NAME, OUTPUT_VALUE_NAME],
        dynamic_axes={
            INPUT_NAME: {0: "batch_size"},
            OUTPUT_POLICY_NAME: {0: "batch_size"},
            OUTPUT_VALUE_NAME: {0: "batch_size"},
        },
        opset_version=opset_version,
        do_constant_folding=True,  # Optimize constants
    )


def verify_onnx_model(onnx_path: str) -> onnx.ModelProto:
    """
    Load and verify an ONNX model.

    Args:
        onnx_path: Path to the ONNX file

    Returns:
        Loaded and verified ONNX model

    Raises:
        onnx.checker.ValidationError: If model is invalid
        FileNotFoundError: If file does not exist
    """
    if not os.path.exists(onnx_path):
        raise FileNotFoundError(f"ONNX file not found: {onnx_path}")

    # Load the model
    onnx_model = onnx.load(onnx_path)

    # Check the model is valid
    onnx.checker.check_model(onnx_model)

    return onnx_model


def compare_outputs(
    pytorch_model: ChessNet, onnx_path: str, test_input: torch.Tensor, rtol: float = 1e-5, atol: float = 1e-5
) -> tuple[bool, str]:
    """
    Compare PyTorch and ONNX model outputs.

    Args:
        pytorch_model: PyTorch ChessNet model
        onnx_path: Path to ONNX model file
        test_input: Input tensor for testing
        rtol: Relative tolerance for comparison
        atol: Absolute tolerance for comparison

    Returns:
        Tuple of (success: bool, message: str)
    """
    # Get PyTorch outputs
    pytorch_model.eval()
    with torch.no_grad():
        pt_policy, pt_value = pytorch_model(test_input)

    # Get ONNX outputs
    ort_session = ort.InferenceSession(onnx_path, providers=["CPUExecutionProvider"])
    ort_inputs = {INPUT_NAME: test_input.detach().cpu().numpy()}
    ort_outputs = ort_session.run(None, ort_inputs)
    ort_policy, ort_value = ort_outputs

    # Compare policy outputs
    pt_policy_np = pt_policy.detach().cpu().numpy()
    policy_match = np.allclose(pt_policy_np, ort_policy, rtol=rtol, atol=atol)

    if not policy_match:
        max_diff = np.max(np.abs(pt_policy_np - ort_policy))
        return False, f"Policy mismatch: max difference = {max_diff}"

    # Compare value outputs
    pt_value_np = pt_value.detach().cpu().numpy()
    value_match = np.allclose(pt_value_np, ort_value, rtol=rtol, atol=atol)

    if not value_match:
        max_diff = np.max(np.abs(pt_value_np - ort_value))
        return False, f"Value mismatch: max difference = {max_diff}"

    return True, "Outputs match within tolerance"


def get_model_info(onnx_path: str) -> dict:
    """
    Get information about an ONNX model.

    Args:
        onnx_path: Path to ONNX file

    Returns:
        Dictionary with model information
    """
    # Get file size
    file_size = os.path.getsize(onnx_path)

    # Load model
    onnx_model = onnx.load(onnx_path)

    # Get input info
    inputs = []
    for inp in onnx_model.graph.input:
        shape = [dim.dim_value if dim.dim_value > 0 else "dynamic" for dim in inp.type.tensor_type.shape.dim]
        inputs.append(
            {"name": inp.name, "shape": shape, "dtype": onnx.TensorProto.DataType.Name(inp.type.tensor_type.elem_type)}
        )

    # Get output info
    outputs = []
    for out in onnx_model.graph.output:
        shape = [dim.dim_value if dim.dim_value > 0 else "dynamic" for dim in out.type.tensor_type.shape.dim]
        outputs.append(
            {"name": out.name, "shape": shape, "dtype": onnx.TensorProto.DataType.Name(out.type.tensor_type.elem_type)}
        )

    return {
        "file_size_bytes": file_size,
        "file_size_mb": file_size / (1024 * 1024),
        "opset_version": onnx_model.opset_import[0].version,
        "inputs": inputs,
        "outputs": outputs,
    }


def print_model_info(info: dict) -> None:
    """Print model information in a readable format."""
    print("\n" + "=" * 60)
    print("ONNX Model Information")
    print("=" * 60)
    print(f"File size: {info['file_size_mb']:.2f} MB ({info['file_size_bytes']:,} bytes)")
    print(f"Opset version: {info['opset_version']}")

    print("\nInputs:")
    for inp in info["inputs"]:
        print(f"  - {inp['name']}: {inp['shape']} ({inp['dtype']})")

    print("\nOutputs:")
    for out in info["outputs"]:
        print(f"  - {out['name']}: {out['shape']} ({out['dtype']})")

    print("=" * 60 + "\n")


def export_checkpoint(
    checkpoint_path: str,
    output_path: str,
    opset_version: int = DEFAULT_OPSET_VERSION,
    verify: bool = True,
    verbose: bool = True,
) -> bool:
    """
    Export a checkpoint to ONNX format with verification.

    This is the main entry point that combines loading, exporting,
    and verification.

    Args:
        checkpoint_path: Path to PyTorch checkpoint
        output_path: Path for output ONNX file
        opset_version: ONNX opset version
        verify: Whether to verify the exported model
        verbose: Whether to print progress information

    Returns:
        True if export and verification succeeded, False otherwise
    """
    if verbose:
        print(f"Loading checkpoint: {checkpoint_path}")

    # Load the model
    model = load_checkpoint_for_export(checkpoint_path)

    if verbose:
        print(f"Model loaded: {model.count_parameters():,} parameters")
        print(f"Architecture: {model.num_blocks} blocks, {model.num_filters} filters")

    if verbose:
        print(f"\nExporting to ONNX: {output_path}")

    # Export to ONNX
    export_to_onnx(model, output_path, opset_version)

    if verbose:
        print("Export complete.")

    if verify:
        if verbose:
            print("\nVerifying ONNX model...")

        # Verify the model structure
        try:
            verify_onnx_model(output_path)
            if verbose:
                print("ONNX model structure is valid.")
        except Exception as e:
            print(f"ONNX verification failed: {e}")
            return False

        # Compare outputs with various batch sizes
        batch_sizes = [1, 4, 8]
        for batch_size in batch_sizes:
            test_input = torch.randn(batch_size, NUM_CHANNELS, 8, 8)
            success, message = compare_outputs(model, output_path, test_input)

            if not success:
                print(f"Output comparison failed (batch_size={batch_size}): {message}")
                return False

            if verbose:
                print(f"Output comparison passed (batch_size={batch_size})")

    # Print model info
    if verbose:
        info = get_model_info(output_path)
        print_model_info(info)

    return True


def parse_args() -> argparse.Namespace:
    """Parse command-line arguments."""
    parser = argparse.ArgumentParser(
        description="Export ChessNet PyTorch model to ONNX format",
        formatter_class=argparse.ArgumentDefaultsHelpFormatter,
    )

    parser.add_argument("checkpoint", type=str, help="Path to PyTorch checkpoint (.pt file)")

    parser.add_argument("output", type=str, help="Path for output ONNX file (.onnx)")

    parser.add_argument("--opset", type=int, default=DEFAULT_OPSET_VERSION, help="ONNX opset version")

    parser.add_argument("--no-verify", action="store_true", help="Skip verification step")

    parser.add_argument("--quiet", "-q", action="store_true", help="Reduce output verbosity")

    return parser.parse_args()


def main():
    """Main entry point for ONNX export."""
    args = parse_args()

    try:
        success = export_checkpoint(
            checkpoint_path=args.checkpoint,
            output_path=args.output,
            opset_version=args.opset,
            verify=not args.no_verify,
            verbose=not args.quiet,
        )

        if success:
            print(f"Successfully exported model to: {args.output}")
            sys.exit(0)
        else:
            print("Export failed.")
            sys.exit(1)

    except FileNotFoundError as e:
        print(f"Error: {e}")
        sys.exit(1)
    except Exception as e:
        print(f"Unexpected error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
