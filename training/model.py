"""
ChessNet: AlphaZero-style Neural Network for Chess

This module implements a residual neural network for chess position evaluation,
following the AlphaZero architecture with separate policy and value heads.

Architecture Overview:
----------------------
1. Input: 18 channels x 8 x 8 (from board encoder)
   - Piece positions, side to move, castling rights, en passant

2. Body (Residual Tower):
   - Initial convolution: 18 -> 128 channels (3x3 kernel)
   - 6 residual blocks, each with:
     * Conv 3x3 -> BatchNorm -> ReLU -> Conv 3x3 -> BatchNorm
     * Skip connection (add input to output)
     * Final ReLU

3. Policy Head (move probabilities):
   - Conv 1x1 -> BatchNorm -> ReLU (128 -> 2 channels)
   - Flatten -> Linear (2*8*8 -> 4096 logits)
   - Output: 4096 values (64 from-squares x 64 to-squares)

4. Value Head (position evaluation):
   - Conv 1x1 -> BatchNorm -> ReLU (128 -> 1 channel)
   - Flatten -> Linear -> ReLU (64 -> 256)
   - Linear -> Tanh (256 -> 1)
   - Output: scalar in [-1, 1] (-1=Black wins, 0=Draw, +1=White wins)

The network has approximately 2M parameters with the default configuration
(6 blocks, 128 filters).
"""

import torch
import torch.nn as nn
import torch.nn.functional as F

# Import device detection from board_encoder
from board_encoder import get_device, NUM_CHANNELS


# Default architecture parameters
DEFAULT_NUM_BLOCKS = 6
DEFAULT_NUM_FILTERS = 128

# Policy output size: 64 from-squares x 64 to-squares
POLICY_OUTPUT_SIZE = 4096

# Board dimensions
BOARD_SIZE = 8


class ResidualBlock(nn.Module):
    """
    A single residual block with skip connection.

    Architecture:
        Input -> Conv3x3 -> BatchNorm -> ReLU -> Conv3x3 -> BatchNorm -> (+Input) -> ReLU -> Output

    The skip connection allows gradients to flow directly through the network,
    making it easier to train deeper networks. This is the key innovation from
    the ResNet paper (He et al., 2015).

    Args:
        num_filters: Number of convolutional filters (input and output channels).
    """

    def __init__(self, num_filters: int = DEFAULT_NUM_FILTERS):
        super().__init__()

        # First convolutional layer
        # 3x3 kernel with padding=1 preserves spatial dimensions (8x8 -> 8x8)
        self.conv1 = nn.Conv2d(
            in_channels=num_filters,
            out_channels=num_filters,
            kernel_size=3,
            padding=1,
            bias=False  # No bias needed when using BatchNorm
        )
        self.bn1 = nn.BatchNorm2d(num_filters)

        # Second convolutional layer
        self.conv2 = nn.Conv2d(
            in_channels=num_filters,
            out_channels=num_filters,
            kernel_size=3,
            padding=1,
            bias=False
        )
        self.bn2 = nn.BatchNorm2d(num_filters)

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        """
        Forward pass through the residual block.

        Args:
            x: Input tensor of shape [batch, num_filters, 8, 8]

        Returns:
            Output tensor of shape [batch, num_filters, 8, 8]
        """
        # Save input for skip connection
        identity = x

        # First conv -> batchnorm -> relu
        out = self.conv1(x)
        out = self.bn1(out)
        out = F.relu(out)

        # Second conv -> batchnorm
        out = self.conv2(out)
        out = self.bn2(out)

        # Add skip connection (this is the "residual" part)
        # The network learns to predict the residual: f(x) = h(x) - x
        # where h(x) is the desired output
        out = out + identity

        # Final ReLU after adding skip connection
        out = F.relu(out)

        return out


class ChessNet(nn.Module):
    """
    AlphaZero-style neural network for chess.

    This network takes an encoded chess position and outputs:
    1. Policy logits: A distribution over all possible moves (4096 values)
    2. Value: An evaluation of the position in [-1, 1]

    The architecture uses a residual tower followed by separate policy and
    value heads. This dual-head design allows the network to learn both
    move selection and position evaluation simultaneously.

    Args:
        num_blocks: Number of residual blocks in the tower (default: 6)
        num_filters: Number of convolutional filters per layer (default: 128)

    Example:
        >>> model = ChessNet()
        >>> x = torch.randn(1, 18, 8, 8)  # Single position
        >>> policy_logits, value = model(x)
        >>> policy_logits.shape
        torch.Size([1, 4096])
        >>> value.shape
        torch.Size([1, 1])
    """

    def __init__(
        self,
        num_blocks: int = DEFAULT_NUM_BLOCKS,
        num_filters: int = DEFAULT_NUM_FILTERS
    ):
        super().__init__()

        self.num_blocks = num_blocks
        self.num_filters = num_filters

        # =====================================================================
        # Initial Convolutional Layer
        # =====================================================================
        # Transform from input channels (18) to feature channels (128)
        # This initial layer extracts basic features from the board representation
        self.initial_conv = nn.Conv2d(
            in_channels=NUM_CHANNELS,  # 18 input channels from board encoder
            out_channels=num_filters,  # 128 output channels
            kernel_size=3,
            padding=1,
            bias=False
        )
        self.initial_bn = nn.BatchNorm2d(num_filters)

        # =====================================================================
        # Residual Tower
        # =====================================================================
        # Stack of residual blocks that learn hierarchical features
        # Each block preserves spatial dimensions (8x8) and channel count
        self.residual_blocks = nn.ModuleList([
            ResidualBlock(num_filters) for _ in range(num_blocks)
        ])

        # =====================================================================
        # Policy Head
        # =====================================================================
        # Reduces to 2 channels, then flattens to predict move probabilities
        # 4096 outputs = 64 possible from-squares x 64 possible to-squares

        # 1x1 convolution to reduce channels: 128 -> 2
        self.policy_conv = nn.Conv2d(
            in_channels=num_filters,
            out_channels=2,
            kernel_size=1,
            bias=False
        )
        self.policy_bn = nn.BatchNorm2d(2)

        # Fully connected layer: 2 * 8 * 8 = 128 -> 4096
        self.policy_fc = nn.Linear(2 * BOARD_SIZE * BOARD_SIZE, POLICY_OUTPUT_SIZE)

        # =====================================================================
        # Value Head
        # =====================================================================
        # Reduces to 1 channel, then predicts scalar position evaluation

        # 1x1 convolution to reduce channels: 128 -> 1
        self.value_conv = nn.Conv2d(
            in_channels=num_filters,
            out_channels=1,
            kernel_size=1,
            bias=False
        )
        self.value_bn = nn.BatchNorm2d(1)

        # First fully connected layer: 1 * 8 * 8 = 64 -> 256
        self.value_fc1 = nn.Linear(1 * BOARD_SIZE * BOARD_SIZE, 256)

        # Second fully connected layer: 256 -> 1
        self.value_fc2 = nn.Linear(256, 1)

    def forward(self, x: torch.Tensor) -> tuple[torch.Tensor, torch.Tensor]:
        """
        Forward pass through the network.

        Args:
            x: Input tensor of shape [batch, 18, 8, 8] representing encoded positions

        Returns:
            Tuple of:
            - policy_logits: Tensor of shape [batch, 4096] with raw policy scores
                            (apply softmax for probabilities)
            - value: Tensor of shape [batch, 1] with position evaluation in [-1, 1]
        """
        # =====================================================================
        # Initial Convolution
        # =====================================================================
        # Transform input: [batch, 18, 8, 8] -> [batch, 128, 8, 8]
        x = self.initial_conv(x)
        x = self.initial_bn(x)
        x = F.relu(x)

        # =====================================================================
        # Residual Tower
        # =====================================================================
        # Process through each residual block
        # Shape remains: [batch, 128, 8, 8]
        for block in self.residual_blocks:
            x = block(x)

        # =====================================================================
        # Policy Head
        # =====================================================================
        # [batch, 128, 8, 8] -> [batch, 2, 8, 8]
        policy = self.policy_conv(x)
        policy = self.policy_bn(policy)
        policy = F.relu(policy)

        # Flatten: [batch, 2, 8, 8] -> [batch, 128]
        policy = policy.view(policy.size(0), -1)

        # Linear: [batch, 128] -> [batch, 4096]
        policy_logits = self.policy_fc(policy)

        # =====================================================================
        # Value Head
        # =====================================================================
        # [batch, 128, 8, 8] -> [batch, 1, 8, 8]
        value = self.value_conv(x)
        value = self.value_bn(value)
        value = F.relu(value)

        # Flatten: [batch, 1, 8, 8] -> [batch, 64]
        value = value.view(value.size(0), -1)

        # First linear + ReLU: [batch, 64] -> [batch, 256]
        value = self.value_fc1(value)
        value = F.relu(value)

        # Second linear + Tanh: [batch, 256] -> [batch, 1]
        value = self.value_fc2(value)
        value = torch.tanh(value)  # Squash to [-1, 1] range

        return policy_logits, value

    def count_parameters(self) -> int:
        """
        Count the total number of trainable parameters in the model.

        Returns:
            Total number of trainable parameters.
        """
        return sum(p.numel() for p in self.parameters() if p.requires_grad)


def create_model(
    num_blocks: int = DEFAULT_NUM_BLOCKS,
    num_filters: int = DEFAULT_NUM_FILTERS,
    device: torch.device | None = None
) -> ChessNet:
    """
    Factory function to create and initialize a ChessNet model.

    Args:
        num_blocks: Number of residual blocks (default: 6)
        num_filters: Number of convolutional filters (default: 128)
        device: Device to place the model on. If None, uses get_device().

    Returns:
        An initialized ChessNet model on the specified device.

    Example:
        >>> model = create_model()
        >>> print(f"Parameters: {model.count_parameters():,}")
        Parameters: 2,003,010
    """
    if device is None:
        device = get_device()

    model = ChessNet(num_blocks=num_blocks, num_filters=num_filters)
    model = model.to(device)

    return model
