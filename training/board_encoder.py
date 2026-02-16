"""
Board Encoder for AlphaZero-style Chess Training Pipeline

This module converts chess positions (python-chess Board objects) into tensor
representations suitable for neural network input.

Input Representation (18 channels x 8 x 8):
-------------------------------------------
Channels 0-5:   White pieces (Pawn, Knight, Bishop, Rook, Queen, King)
Channels 6-11:  Black pieces (Pawn, Knight, Bishop, Rook, Queen, King)
Channel 12:     Side to move (all 1s if White to move, all 0s if Black)
Channels 13-16: Castling rights (4 planes: WK, WQ, BK, BQ)
Channel 17:     En passant file (column marked with 1s if en passant possible)

Board Indexing:
---------------
- Rank 0 = row 1 (white's back rank), Rank 7 = row 8 (black's back rank)
- File 0 = column a, File 7 = column h
- Tensor shape: [18, 8, 8] where dimensions are [channel, rank, file]

This encoding is deterministic: the same board position always produces
the same tensor output.
"""

import chess
import numpy as np
import torch


# Number of channels in the board encoding
NUM_CHANNELS = 18

# Piece type indices (python-chess uses 1-6, we use 0-5 for array indexing)
# PAWN=1, KNIGHT=2, BISHOP=3, ROOK=4, QUEEN=5, KING=6 in python-chess
PIECE_TYPES = [chess.PAWN, chess.KNIGHT, chess.BISHOP, chess.ROOK, chess.QUEEN, chess.KING]

# Channel assignments
WHITE_PIECE_OFFSET = 0   # Channels 0-5 for white pieces
BLACK_PIECE_OFFSET = 6   # Channels 6-11 for black pieces
SIDE_TO_MOVE_CHANNEL = 12
CASTLING_WK_CHANNEL = 13  # White kingside castling
CASTLING_WQ_CHANNEL = 14  # White queenside castling
CASTLING_BK_CHANNEL = 15  # Black kingside castling
CASTLING_BQ_CHANNEL = 16  # Black queenside castling
EN_PASSANT_CHANNEL = 17


def get_device() -> torch.device:
    """
    Detect the best available compute device.

    Priority order:
    1. MPS (Apple Silicon GPU) - for Mac ARM chips
    2. CUDA (NVIDIA GPU) - for systems with NVIDIA GPUs
    3. CPU - fallback for all systems

    Returns:
        torch.device: The best available device for tensor computation.
    """
    if torch.backends.mps.is_available():
        # Apple Silicon (M1/M2/M3) GPU acceleration
        return torch.device("mps")
    elif torch.cuda.is_available():
        # NVIDIA GPU acceleration
        return torch.device("cuda")
    else:
        # CPU fallback
        return torch.device("cpu")


def encode_board(board: chess.Board) -> np.ndarray:
    """
    Encode a chess board position into an 18-channel numpy array.

    This function converts a python-chess Board object into a tensor
    representation suitable for neural network input. The encoding
    captures all information needed to fully describe a chess position:
    piece positions, side to move, castling rights, and en passant.

    Args:
        board: A python-chess Board object representing the position.

    Returns:
        np.ndarray: A float32 array of shape [18, 8, 8] containing the
                   encoded position. Values are 0.0 or 1.0 (binary planes).

    Example:
        >>> import chess
        >>> board = chess.Board()  # Starting position
        >>> encoded = encode_board(board)
        >>> encoded.shape
        (18, 8, 8)
    """
    # Initialize the encoding array with zeros
    # Shape: [channels, ranks, files] = [18, 8, 8]
    encoding = np.zeros((NUM_CHANNELS, 8, 8), dtype=np.float32)

    # --- Encode piece positions (channels 0-11) ---
    # Iterate over all 64 squares on the board
    for square in chess.SQUARES:
        piece = board.piece_at(square)
        if piece is not None:
            # Convert square index (0-63) to rank and file
            # python-chess square indexing: a1=0, b1=1, ..., h8=63
            rank = chess.square_rank(square)  # 0-7 (rank 1-8)
            file = chess.square_file(square)  # 0-7 (file a-h)

            # Determine channel based on piece color and type
            # piece.piece_type returns 1-6 (PAWN to KING)
            # We subtract 1 to get 0-5 for array indexing
            piece_index = piece.piece_type - 1

            if piece.color == chess.WHITE:
                channel = WHITE_PIECE_OFFSET + piece_index
            else:
                channel = BLACK_PIECE_OFFSET + piece_index

            # Set the piece's position in the appropriate channel
            encoding[channel, rank, file] = 1.0

    # --- Encode side to move (channel 12) ---
    # All 1s if White to move, all 0s if Black to move
    if board.turn == chess.WHITE:
        encoding[SIDE_TO_MOVE_CHANNEL, :, :] = 1.0
    # If Black to move, channel stays all zeros (already initialized)

    # --- Encode castling rights (channels 13-16) ---
    # Each castling right gets its own binary plane
    # If the right exists, the entire plane is filled with 1s
    if board.has_kingside_castling_rights(chess.WHITE):
        encoding[CASTLING_WK_CHANNEL, :, :] = 1.0
    if board.has_queenside_castling_rights(chess.WHITE):
        encoding[CASTLING_WQ_CHANNEL, :, :] = 1.0
    if board.has_kingside_castling_rights(chess.BLACK):
        encoding[CASTLING_BK_CHANNEL, :, :] = 1.0
    if board.has_queenside_castling_rights(chess.BLACK):
        encoding[CASTLING_BQ_CHANNEL, :, :] = 1.0

    # --- Encode en passant square (channel 17) ---
    # If en passant is possible, mark the entire file (column) with 1s
    if board.ep_square is not None:
        ep_file = chess.square_file(board.ep_square)
        encoding[EN_PASSANT_CHANNEL, :, ep_file] = 1.0

    return encoding


def encode_board_tensor(board: chess.Board, device: torch.device) -> torch.Tensor:
    """
    Encode a chess board position into a PyTorch tensor on the specified device.

    This is a convenience function that wraps encode_board() and converts
    the numpy array to a PyTorch tensor on the desired device (CPU, CUDA, or MPS).

    Args:
        board: A python-chess Board object representing the position.
        device: The torch.device to place the tensor on.

    Returns:
        torch.Tensor: A float32 tensor of shape [18, 8, 8] on the specified device.

    Example:
        >>> import chess
        >>> import torch
        >>> board = chess.Board()
        >>> device = get_device()
        >>> tensor = encode_board_tensor(board, device)
        >>> tensor.shape
        torch.Size([18, 8, 8])
    """
    # First encode to numpy array
    encoding = encode_board(board)

    # Convert to PyTorch tensor
    # We create the tensor on CPU first, then move to device
    # This is the most compatible approach across all devices (CPU, CUDA, MPS)
    tensor = torch.from_numpy(encoding)

    # Move to the target device
    return tensor.to(device)
