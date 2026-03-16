"""
Board Encoder for AlphaZero-style Chess Training Pipeline

This module converts chess positions (python-chess Board objects) into tensor
representations suitable for neural network input.

Input Representation (NUM_CHANNELS channels x 8 x 8):
------------------------------------------------------
Current position (18 channels):
  Channels 0-5:   White pieces (Pawn, Knight, Bishop, Rook, Queen, King)
  Channels 6-11:  Black pieces (Pawn, Knight, Bishop, Rook, Queen, King)
  Channel 12:     Side to move (all 1s if White to move, all 0s if Black)
  Channels 13-16: Castling rights (4 planes: WK, WQ, BK, BQ)
  Channel 17:     En passant file (column marked with 1s if en passant possible)

History positions (NUM_HISTORY_POSITIONS * 12 channels):
  For each of the last N positions (most recent first):
    12 piece planes (6 white + 6 black), same layout as channels 0-11

Total channels = 18 + NUM_HISTORY_POSITIONS * 12

Board Indexing:
---------------
- Rank 0 = row 1 (white's back rank), Rank 7 = row 8 (black's back rank)
- File 0 = column a, File 7 = column h
- Tensor shape: [NUM_CHANNELS, 8, 8] where dimensions are [channel, rank, file]

This encoding is deterministic: the same board position and history always
produces the same tensor output.
"""

import chess
import numpy as np
import torch
from typing import List, Optional


# Number of previous positions to include in the encoding.
# 4 is enough to detect threefold repetition patterns while keeping
# the input manageable for the ~2M parameter model.
NUM_HISTORY_POSITIONS = 4

# Derived constants
PIECE_CHANNELS = 12  # 6 piece types x 2 colors
CURRENT_POSITION_CHANNELS = 18  # 12 pieces + 1 side-to-move + 4 castling + 1 en passant
NUM_CHANNELS = CURRENT_POSITION_CHANNELS + NUM_HISTORY_POSITIONS * PIECE_CHANNELS

# Piece type indices (python-chess uses 1-6, we use 0-5 for array indexing)
# PAWN=1, KNIGHT=2, BISHOP=3, ROOK=4, QUEEN=5, KING=6 in python-chess
PIECE_TYPES = [chess.PAWN, chess.KNIGHT, chess.BISHOP, chess.ROOK, chess.QUEEN, chess.KING]

# Channel assignments for current position
WHITE_PIECE_OFFSET = 0   # Channels 0-5 for white pieces
BLACK_PIECE_OFFSET = 6   # Channels 6-11 for black pieces
SIDE_TO_MOVE_CHANNEL = 12
CASTLING_WK_CHANNEL = 13  # White kingside castling
CASTLING_WQ_CHANNEL = 14  # White queenside castling
CASTLING_BK_CHANNEL = 15  # Black kingside castling
CASTLING_BQ_CHANNEL = 16  # Black queenside castling
EN_PASSANT_CHANNEL = 17

# History channels start after the current position
HISTORY_OFFSET = CURRENT_POSITION_CHANNELS  # = 18


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


def _encode_pieces(board: chess.Board, encoding: np.ndarray, offset: int) -> None:
    """
    Encode piece positions into 12 channels starting at the given offset.

    Args:
        board: Chess board to encode
        encoding: Target array to write into [C, 8, 8]
        offset: Channel offset to start writing at
    """
    for square in chess.SQUARES:
        piece = board.piece_at(square)
        if piece is not None:
            rank = chess.square_rank(square)
            file = chess.square_file(square)
            piece_index = piece.piece_type - 1

            if piece.color == chess.WHITE:
                channel = offset + piece_index
            else:
                channel = offset + 6 + piece_index

            encoding[channel, rank, file] = 1.0


def encode_board(
    board: chess.Board,
    history: Optional[List[chess.Board]] = None
) -> np.ndarray:
    """
    Encode a chess board position into a multi-channel numpy array.

    The encoding includes the current position (18 channels) plus
    piece planes from the last NUM_HISTORY_POSITIONS positions (12 channels each).
    If fewer history positions are available, the remaining channels are zero-filled.

    Args:
        board: A python-chess Board object representing the current position.
        history: Optional list of previous board states (most recent last).
                 Only the last NUM_HISTORY_POSITIONS entries are used.

    Returns:
        np.ndarray: A float32 array of shape [NUM_CHANNELS, 8, 8].

    Example:
        >>> import chess
        >>> board = chess.Board()
        >>> encoded = encode_board(board)
        >>> encoded.shape[0] == NUM_CHANNELS
        True
    """
    encoding = np.zeros((NUM_CHANNELS, 8, 8), dtype=np.float32)

    # --- Current position: piece planes (channels 0-11) ---
    _encode_pieces(board, encoding, offset=0)

    # --- Current position: metadata (channels 12-17) ---
    if board.turn == chess.WHITE:
        encoding[SIDE_TO_MOVE_CHANNEL, :, :] = 1.0

    if board.has_kingside_castling_rights(chess.WHITE):
        encoding[CASTLING_WK_CHANNEL, :, :] = 1.0
    if board.has_queenside_castling_rights(chess.WHITE):
        encoding[CASTLING_WQ_CHANNEL, :, :] = 1.0
    if board.has_kingside_castling_rights(chess.BLACK):
        encoding[CASTLING_BK_CHANNEL, :, :] = 1.0
    if board.has_queenside_castling_rights(chess.BLACK):
        encoding[CASTLING_BQ_CHANNEL, :, :] = 1.0

    if board.ep_square is not None:
        ep_file = chess.square_file(board.ep_square)
        encoding[EN_PASSANT_CHANNEL, :, ep_file] = 1.0

    # --- History positions: piece planes only (12 channels each) ---
    if history:
        # Take the most recent NUM_HISTORY_POSITIONS entries
        recent = history[-NUM_HISTORY_POSITIONS:]
        for i, hist_board in enumerate(reversed(recent)):
            # i=0 is most recent history position (t-1)
            ch_offset = HISTORY_OFFSET + i * PIECE_CHANNELS
            _encode_pieces(hist_board, encoding, offset=ch_offset)

    return encoding


def encode_board_tensor(
    board: chess.Board,
    device: torch.device,
    history: Optional[List[chess.Board]] = None
) -> torch.Tensor:
    """
    Encode a chess board position into a PyTorch tensor on the specified device.

    Args:
        board: A python-chess Board object representing the position.
        device: The torch.device to place the tensor on.
        history: Optional list of previous board states.

    Returns:
        torch.Tensor: A float32 tensor of shape [NUM_CHANNELS, 8, 8].
    """
    encoding = encode_board(board, history=history)
    tensor = torch.from_numpy(encoding)
    return tensor.to(device)
