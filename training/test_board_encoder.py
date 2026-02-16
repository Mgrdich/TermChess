"""
Unit Tests for Board Encoder

This module tests the board encoding functionality to ensure:
1. Correct tensor shapes
2. Accurate piece placement encoding
3. Proper handling of side to move
4. Correct castling rights encoding
5. En passant file encoding
6. Device compatibility (MPS, CUDA, CPU)
"""

import chess
import numpy as np
import pytest
import torch

from board_encoder import (
    encode_board,
    encode_board_tensor,
    get_device,
    NUM_CHANNELS,
    WHITE_PIECE_OFFSET,
    BLACK_PIECE_OFFSET,
    SIDE_TO_MOVE_CHANNEL,
    CASTLING_WK_CHANNEL,
    CASTLING_WQ_CHANNEL,
    CASTLING_BK_CHANNEL,
    CASTLING_BQ_CHANNEL,
    EN_PASSANT_CHANNEL,
)


class TestOutputShape:
    """Tests for verifying the output tensor shape."""

    def test_output_shape_starting_position(self):
        """The encoded starting position should have shape [18, 8, 8]."""
        board = chess.Board()
        encoded = encode_board(board)

        assert encoded.shape == (18, 8, 8), (
            f"Expected shape (18, 8, 8), got {encoded.shape}"
        )

    def test_output_shape_empty_board(self):
        """An empty board should still produce shape [18, 8, 8]."""
        board = chess.Board.empty()
        encoded = encode_board(board)

        assert encoded.shape == (18, 8, 8)

    def test_output_dtype(self):
        """The encoded board should be float32."""
        board = chess.Board()
        encoded = encode_board(board)

        assert encoded.dtype == np.float32


class TestStartingPosition:
    """Tests for verifying the starting position is encoded correctly."""

    def test_white_pawns_on_rank_1(self):
        """White pawns should be on rank 1 (index 1) in the starting position."""
        board = chess.Board()
        encoded = encode_board(board)

        # Channel 0 is white pawns (PAWN=1, so index 0)
        white_pawn_channel = encoded[WHITE_PIECE_OFFSET + 0]  # Pawn is piece_type 1, index 0

        # Check rank 1 (index 1) has all pawns
        assert np.all(white_pawn_channel[1, :] == 1.0), (
            "White pawns should be on rank 1 (second row)"
        )
        # Check other ranks are empty
        assert np.sum(white_pawn_channel[0, :]) == 0, "Rank 0 should have no white pawns"
        assert np.sum(white_pawn_channel[2:, :]) == 0, "Ranks 2-7 should have no white pawns"

    def test_black_pawns_on_rank_6(self):
        """Black pawns should be on rank 6 (index 6) in the starting position."""
        board = chess.Board()
        encoded = encode_board(board)

        # Channel 6 is black pawns
        black_pawn_channel = encoded[BLACK_PIECE_OFFSET + 0]

        # Check rank 6 has all pawns
        assert np.all(black_pawn_channel[6, :] == 1.0), (
            "Black pawns should be on rank 6 (seventh row)"
        )
        # Check other ranks are empty
        assert np.sum(black_pawn_channel[:6, :]) == 0, "Ranks 0-5 should have no black pawns"
        assert np.sum(black_pawn_channel[7, :]) == 0, "Rank 7 should have no black pawns"

    def test_white_pieces_on_rank_0(self):
        """White pieces (non-pawns) should be on rank 0 in the starting position."""
        board = chess.Board()
        encoded = encode_board(board)

        # Check knights on b1 and g1 (files 1 and 6)
        knight_channel = encoded[WHITE_PIECE_OFFSET + 1]  # Knight is piece_type 2, index 1
        assert knight_channel[0, 1] == 1.0, "White knight should be on b1"
        assert knight_channel[0, 6] == 1.0, "White knight should be on g1"
        assert np.sum(knight_channel) == 2, "Should have exactly 2 white knights"

        # Check bishops on c1 and f1 (files 2 and 5)
        bishop_channel = encoded[WHITE_PIECE_OFFSET + 2]
        assert bishop_channel[0, 2] == 1.0, "White bishop should be on c1"
        assert bishop_channel[0, 5] == 1.0, "White bishop should be on f1"

        # Check rooks on a1 and h1 (files 0 and 7)
        rook_channel = encoded[WHITE_PIECE_OFFSET + 3]
        assert rook_channel[0, 0] == 1.0, "White rook should be on a1"
        assert rook_channel[0, 7] == 1.0, "White rook should be on h1"

        # Check queen on d1 (file 3)
        queen_channel = encoded[WHITE_PIECE_OFFSET + 4]
        assert queen_channel[0, 3] == 1.0, "White queen should be on d1"

        # Check king on e1 (file 4)
        king_channel = encoded[WHITE_PIECE_OFFSET + 5]
        assert king_channel[0, 4] == 1.0, "White king should be on e1"

    def test_black_pieces_on_rank_7(self):
        """Black pieces (non-pawns) should be on rank 7 in the starting position."""
        board = chess.Board()
        encoded = encode_board(board)

        # Check knights on b8 and g8
        knight_channel = encoded[BLACK_PIECE_OFFSET + 1]
        assert knight_channel[7, 1] == 1.0, "Black knight should be on b8"
        assert knight_channel[7, 6] == 1.0, "Black knight should be on g8"

        # Check king on e8 (file 4)
        king_channel = encoded[BLACK_PIECE_OFFSET + 5]
        assert king_channel[7, 4] == 1.0, "Black king should be on e8"

    def test_total_piece_count(self):
        """Starting position should have 32 pieces total."""
        board = chess.Board()
        encoded = encode_board(board)

        # Sum all piece channels (0-11)
        total_pieces = np.sum(encoded[:12])
        assert total_pieces == 32, f"Expected 32 pieces, got {total_pieces}"


class TestSideToMove:
    """Tests for the side to move channel."""

    def test_white_to_move_starting_position(self):
        """In the starting position, White is to move (channel 12 all 1s)."""
        board = chess.Board()
        encoded = encode_board(board)

        side_to_move = encoded[SIDE_TO_MOVE_CHANNEL]
        assert np.all(side_to_move == 1.0), (
            "Side to move channel should be all 1s when White to move"
        )

    def test_black_to_move(self):
        """After one move, Black is to move (channel 12 all 0s)."""
        board = chess.Board()
        board.push_san("e4")  # White's move
        encoded = encode_board(board)

        side_to_move = encoded[SIDE_TO_MOVE_CHANNEL]
        assert np.all(side_to_move == 0.0), (
            "Side to move channel should be all 0s when Black to move"
        )

    def test_white_to_move_after_two_moves(self):
        """After two moves, White is to move again."""
        board = chess.Board()
        board.push_san("e4")
        board.push_san("e5")
        encoded = encode_board(board)

        side_to_move = encoded[SIDE_TO_MOVE_CHANNEL]
        assert np.all(side_to_move == 1.0), (
            "Side to move channel should be all 1s when White to move"
        )


class TestCastlingRights:
    """Tests for the castling rights channels."""

    def test_all_castling_rights_starting_position(self):
        """Starting position has all four castling rights."""
        board = chess.Board()
        encoded = encode_board(board)

        # All castling channels should be all 1s
        assert np.all(encoded[CASTLING_WK_CHANNEL] == 1.0), "WK castling should be available"
        assert np.all(encoded[CASTLING_WQ_CHANNEL] == 1.0), "WQ castling should be available"
        assert np.all(encoded[CASTLING_BK_CHANNEL] == 1.0), "BK castling should be available"
        assert np.all(encoded[CASTLING_BQ_CHANNEL] == 1.0), "BQ castling should be available"

    def test_no_castling_rights(self):
        """A position with no castling rights should have all 0s in castling channels."""
        # Create a position with no castling rights using FEN
        # This is a simple position with both kings having moved
        board = chess.Board("r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w - - 0 1")
        encoded = encode_board(board)

        # All castling channels should be all 0s
        assert np.all(encoded[CASTLING_WK_CHANNEL] == 0.0), "WK castling should not be available"
        assert np.all(encoded[CASTLING_WQ_CHANNEL] == 0.0), "WQ castling should not be available"
        assert np.all(encoded[CASTLING_BK_CHANNEL] == 0.0), "BK castling should not be available"
        assert np.all(encoded[CASTLING_BQ_CHANNEL] == 0.0), "BQ castling should not be available"

    def test_partial_castling_rights(self):
        """Test position with only some castling rights."""
        # Only white kingside and black queenside castling available
        board = chess.Board("r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w Kq - 0 1")
        encoded = encode_board(board)

        assert np.all(encoded[CASTLING_WK_CHANNEL] == 1.0), "WK castling should be available"
        assert np.all(encoded[CASTLING_WQ_CHANNEL] == 0.0), "WQ castling should not be available"
        assert np.all(encoded[CASTLING_BK_CHANNEL] == 0.0), "BK castling should not be available"
        assert np.all(encoded[CASTLING_BQ_CHANNEL] == 1.0), "BQ castling should be available"

    def test_castling_lost_after_rook_move(self):
        """Moving a rook should remove castling rights on that side."""
        board = chess.Board()
        board.push_san("a4")  # Move pawn to allow rook out
        board.push_san("a5")
        board.push_san("Ra3")  # Move queenside rook
        board.push_san("Ra6")

        encoded = encode_board(board)

        # White still has kingside but not queenside
        assert np.all(encoded[CASTLING_WK_CHANNEL] == 1.0), "WK castling should still be available"
        assert np.all(encoded[CASTLING_WQ_CHANNEL] == 0.0), "WQ castling should be lost after Ra3"


class TestEnPassant:
    """Tests for the en passant channel."""

    def test_no_en_passant_starting_position(self):
        """Starting position has no en passant square."""
        board = chess.Board()
        encoded = encode_board(board)

        ep_channel = encoded[EN_PASSANT_CHANNEL]
        assert np.all(ep_channel == 0.0), (
            "En passant channel should be all 0s in starting position"
        )

    def test_en_passant_after_pawn_push(self):
        """After a double pawn push, the en passant file should be marked."""
        board = chess.Board()
        board.push_san("e4")  # Double pawn push

        encoded = encode_board(board)
        ep_channel = encoded[EN_PASSANT_CHANNEL]

        # En passant square is e3, so file e (index 4) should be marked
        # The entire column should be 1s
        assert np.all(ep_channel[:, 4] == 1.0), (
            "E-file should be marked for en passant after e4"
        )
        # Other columns should be 0
        assert np.sum(ep_channel[:, :4]) == 0, "Files a-d should have no en passant"
        assert np.sum(ep_channel[:, 5:]) == 0, "Files f-h should have no en passant"

    def test_en_passant_different_file(self):
        """Test en passant on a different file."""
        board = chess.Board()
        board.push_san("d3")
        board.push_san("a5")  # Black double pawn push on a-file

        encoded = encode_board(board)
        ep_channel = encoded[EN_PASSANT_CHANNEL]

        # En passant square is a6, so file a (index 0) should be marked
        assert np.all(ep_channel[:, 0] == 1.0), (
            "A-file should be marked for en passant after a5"
        )

    def test_en_passant_disappears_after_other_move(self):
        """En passant opportunity disappears after the next move."""
        board = chess.Board()
        board.push_san("e4")  # Creates en passant opportunity
        board.push_san("d6")  # Black makes a different move

        encoded = encode_board(board)
        ep_channel = encoded[EN_PASSANT_CHANNEL]

        # En passant is no longer available
        assert np.all(ep_channel == 0.0), (
            "En passant should disappear after opponent's move"
        )


class TestDeterminism:
    """Tests to ensure encoding is deterministic."""

    def test_same_position_same_encoding(self):
        """The same position should always produce the same encoding."""
        board1 = chess.Board()
        board2 = chess.Board()

        encoded1 = encode_board(board1)
        encoded2 = encode_board(board2)

        assert np.array_equal(encoded1, encoded2), (
            "Same position should produce identical encodings"
        )

    def test_encoding_reproducible_after_moves(self):
        """Reaching the same position via different move orders should give same encoding."""
        # Create position via one move order
        board1 = chess.Board()
        board1.push_san("e4")
        board1.push_san("e5")
        board1.push_san("Nf3")

        # Create same position via FEN
        board2 = chess.Board("rnbqkbnr/pppp1ppp/8/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2")

        encoded1 = encode_board(board1)
        encoded2 = encode_board(board2)

        assert np.array_equal(encoded1, encoded2), (
            "Same position via different paths should produce identical encodings"
        )


class TestTensorEncoding:
    """Tests for the tensor encoding function."""

    def test_tensor_shape(self):
        """Tensor output should have correct shape."""
        board = chess.Board()
        device = torch.device("cpu")
        tensor = encode_board_tensor(board, device)

        assert tensor.shape == torch.Size([18, 8, 8])

    def test_tensor_dtype(self):
        """Tensor should be float32."""
        board = chess.Board()
        device = torch.device("cpu")
        tensor = encode_board_tensor(board, device)

        assert tensor.dtype == torch.float32

    def test_tensor_device_cpu(self):
        """Tensor should be on the specified CPU device."""
        board = chess.Board()
        device = torch.device("cpu")
        tensor = encode_board_tensor(board, device)

        assert tensor.device.type == "cpu"

    def test_tensor_matches_numpy(self):
        """Tensor values should match numpy encoding."""
        board = chess.Board()

        np_encoded = encode_board(board)
        tensor = encode_board_tensor(board, torch.device("cpu"))

        assert np.allclose(np_encoded, tensor.numpy())


class TestDeviceDetection:
    """Tests for device detection functionality."""

    def test_get_device_returns_device(self):
        """get_device should return a torch.device object."""
        device = get_device()
        assert isinstance(device, torch.device)

    def test_get_device_valid_type(self):
        """Device type should be one of cpu, cuda, or mps."""
        device = get_device()
        assert device.type in ("cpu", "cuda", "mps")


class TestMPSDevice:
    """Tests for MPS (Apple Silicon) device compatibility."""

    @pytest.mark.skipif(
        not torch.backends.mps.is_available(),
        reason="MPS not available on this system"
    )
    def test_encoder_runs_on_mps(self):
        """Encoder should work on MPS device (Mac ARM)."""
        board = chess.Board()
        device = torch.device("mps")

        tensor = encode_board_tensor(board, device)

        assert tensor.device.type == "mps"
        assert tensor.shape == torch.Size([18, 8, 8])

    @pytest.mark.skipif(
        not torch.backends.mps.is_available(),
        reason="MPS not available on this system"
    )
    def test_mps_tensor_values_correct(self):
        """MPS tensor should have same values as CPU tensor."""
        board = chess.Board()

        cpu_tensor = encode_board_tensor(board, torch.device("cpu"))
        mps_tensor = encode_board_tensor(board, torch.device("mps"))

        # Move MPS tensor to CPU for comparison
        mps_on_cpu = mps_tensor.cpu()

        assert torch.allclose(cpu_tensor, mps_on_cpu)

    @pytest.mark.skipif(
        not torch.backends.mps.is_available(),
        reason="MPS not available on this system"
    )
    def test_complex_position_on_mps(self):
        """Test a more complex position on MPS."""
        # Sicilian Defense position
        board = chess.Board()
        board.push_san("e4")
        board.push_san("c5")
        board.push_san("Nf3")
        board.push_san("d6")

        device = torch.device("mps")
        tensor = encode_board_tensor(board, device)

        assert tensor.device.type == "mps"
        # Verify piece count is correct (32 pieces still on board)
        piece_sum = tensor[:12].sum().item()
        assert piece_sum == 32


class TestEdgeCases:
    """Tests for edge cases and unusual positions."""

    def test_empty_board(self):
        """Empty board should encode with all zeros in piece channels."""
        board = chess.Board.empty()
        encoded = encode_board(board)

        # All piece channels should be zero
        assert np.sum(encoded[:12]) == 0

    def test_kings_only(self):
        """A position with only kings should encode correctly."""
        board = chess.Board("4k3/8/8/8/8/8/8/4K3 w - - 0 1")
        encoded = encode_board(board)

        # Only 2 pieces
        assert np.sum(encoded[:12]) == 2

        # White king on e1
        assert encoded[WHITE_PIECE_OFFSET + 5, 0, 4] == 1.0
        # Black king on e8
        assert encoded[BLACK_PIECE_OFFSET + 5, 7, 4] == 1.0

    def test_promoted_piece(self):
        """A promoted piece should encode correctly."""
        # Position with a promoted queen
        board = chess.Board("Q3k3/8/8/8/8/8/8/4K3 w - - 0 1")
        encoded = encode_board(board)

        # White queen on a8 (rank 7, file 0)
        assert encoded[WHITE_PIECE_OFFSET + 4, 7, 0] == 1.0

    def test_multiple_queens(self):
        """Multiple queens should all be encoded."""
        board = chess.Board("QQ2k3/8/8/8/8/8/8/4K3 w - - 0 1")
        encoded = encode_board(board)

        queen_channel = encoded[WHITE_PIECE_OFFSET + 4]
        assert np.sum(queen_channel) == 2
        assert queen_channel[7, 0] == 1.0  # a8
        assert queen_channel[7, 1] == 1.0  # b8


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
