package bot

import (
	"testing"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// TestEncodeBoard_Shape verifies the output length is exactly 1152 (18*8*8).
func TestEncodeBoard_Shape(t *testing.T) {
	board := engine.NewBoard()
	encoded := encodeBoard(board)

	if len(encoded) != encodingSize {
		t.Errorf("encodeBoard output length = %d, want %d", len(encoded), encodingSize)
	}
}

// TestEncodeBoard_StartingPosition verifies the encoding of the standard
// starting position against known correct values.
func TestEncodeBoard_StartingPosition(t *testing.T) {
	board := engine.NewBoard()
	encoded := encodeBoard(board)

	// Helper to get value at [channel][rank][file].
	at := func(ch, rank, file int) float32 {
		return encoded[ch*64+rank*8+file]
	}

	// Channel 0: white pawns - should be 1.0 at rank=1 for all files.
	for file := 0; file < 8; file++ {
		if v := at(0, 1, file); v != 1.0 {
			t.Errorf("white pawn at rank=1, file=%d: got %f, want 1.0", file, v)
		}
		// No white pawns elsewhere.
		for rank := 0; rank < 8; rank++ {
			if rank == 1 {
				continue
			}
			if v := at(0, rank, file); v != 0.0 {
				t.Errorf("white pawn at rank=%d, file=%d: got %f, want 0.0", rank, file, v)
			}
		}
	}

	// Channel 5: white king - should be 1.0 at rank=0, file=4 (e1).
	if v := at(5, 0, 4); v != 1.0 {
		t.Errorf("white king at e1 (rank=0, file=4): got %f, want 1.0", v)
	}

	// Channel 6: black pawns - should be 1.0 at rank=6 for all files.
	for file := 0; file < 8; file++ {
		if v := at(6, 6, file); v != 1.0 {
			t.Errorf("black pawn at rank=6, file=%d: got %f, want 1.0", file, v)
		}
	}

	// Channel 11: black king - should be 1.0 at rank=7, file=4 (e8).
	if v := at(11, 7, 4); v != 1.0 {
		t.Errorf("black king at e8 (rank=7, file=4): got %f, want 1.0", v)
	}

	// Channel 12: side to move - all 1.0 (white to move).
	for i := 0; i < 64; i++ {
		if encoded[12*64+i] != 1.0 {
			t.Errorf("side-to-move plane index %d: got %f, want 1.0", i, encoded[12*64+i])
		}
	}

	// Channels 13-16: all castling rights available, all 1.0.
	for ch := 13; ch <= 16; ch++ {
		for i := 0; i < 64; i++ {
			if encoded[ch*64+i] != 1.0 {
				t.Errorf("castling channel %d, index %d: got %f, want 1.0", ch, i, encoded[ch*64+i])
			}
		}
	}

	// Channel 17: no en passant, all 0.0.
	for i := 0; i < 64; i++ {
		if encoded[17*64+i] != 0.0 {
			t.Errorf("en passant plane index %d: got %f, want 0.0", i, encoded[17*64+i])
		}
	}
}

// TestEncodeBoard_AfterE4 verifies the encoding after the move 1.e4.
// The pawn moves from e2 (rank=1, file=4) to e4 (rank=3, file=4),
// side to move becomes black, and en passant is available on file 4.
func TestEncodeBoard_AfterE4(t *testing.T) {
	board := engine.NewBoard()

	// Make the move e2-e4 (square 12 to square 28).
	e2 := engine.NewSquare(4, 1) // file=4 (e), rank=1
	e4 := engine.NewSquare(4, 3) // file=4 (e), rank=3
	err := board.MakeMove(engine.Move{From: e2, To: e4})
	if err != nil {
		t.Fatalf("MakeMove e2-e4 failed: %v", err)
	}

	encoded := encodeBoard(board)

	at := func(ch, rank, file int) float32 {
		return encoded[ch*64+rank*8+file]
	}

	// White pawn should NOT be at rank=1, file=4 anymore.
	if v := at(0, 1, 4); v != 0.0 {
		t.Errorf("white pawn at e2 after e4: got %f, want 0.0", v)
	}

	// White pawn should be at rank=3, file=4.
	if v := at(0, 3, 4); v != 1.0 {
		t.Errorf("white pawn at e4 after e4: got %f, want 1.0", v)
	}

	// Channel 12: side to move - all 0.0 (black to move).
	for i := 0; i < 64; i++ {
		if encoded[12*64+i] != 0.0 {
			t.Errorf("side-to-move plane after e4, index %d: got %f, want 0.0", i, encoded[12*64+i])
		}
	}

	// Channel 17: en passant file 4 should have 1.0 for all ranks.
	for rank := 0; rank < 8; rank++ {
		if v := at(17, rank, 4); v != 1.0 {
			t.Errorf("en passant at rank=%d, file=4: got %f, want 1.0", rank, v)
		}
	}

	// Other files in channel 17 should be 0.0.
	for file := 0; file < 8; file++ {
		if file == 4 {
			continue
		}
		for rank := 0; rank < 8; rank++ {
			if v := at(17, rank, file); v != 0.0 {
				t.Errorf("en passant at rank=%d, file=%d: got %f, want 0.0", rank, file, v)
			}
		}
	}

	// All castling rights should still be available.
	for ch := 13; ch <= 16; ch++ {
		for i := 0; i < 64; i++ {
			if encoded[ch*64+i] != 1.0 {
				t.Errorf("castling channel %d after e4, index %d: got %f, want 1.0", ch, i, encoded[ch*64+i])
			}
		}
	}
}

// TestEncodeBoard_PythonReference verifies specific index values that
// can be computed analytically and would match the Python encoder output.
func TestEncodeBoard_PythonReference(t *testing.T) {
	board := engine.NewBoard()
	encoded := encodeBoard(board)

	// Starting position reference values:
	//
	// White rook at a1 (rank=0, file=0):
	//   channel = 3 (Rook, pieceType=4, offset=3)
	//   index = 3*64 + 0*8 + 0 = 192
	if encoded[192] != 1.0 {
		t.Errorf("white rook at a1, index 192: got %f, want 1.0", encoded[192])
	}

	// White rook at h1 (rank=0, file=7):
	//   channel = 3, index = 3*64 + 0*8 + 7 = 199
	if encoded[199] != 1.0 {
		t.Errorf("white rook at h1, index 199: got %f, want 1.0", encoded[199])
	}

	// Black queen at d8 (rank=7, file=3):
	//   channel = 10 (6 + Queen offset 4), index = 10*64 + 7*8 + 3 = 699
	if encoded[699] != 1.0 {
		t.Errorf("black queen at d8, index 699: got %f, want 1.0", encoded[699])
	}

	// White queen at d1 (rank=0, file=3):
	//   channel = 4 (Queen, pieceType=5, offset=4), index = 4*64 + 0*8 + 3 = 259
	if encoded[259] != 1.0 {
		t.Errorf("white queen at d1, index 259: got %f, want 1.0", encoded[259])
	}

	// White knight at b1 (rank=0, file=1):
	//   channel = 1 (Knight, pieceType=2, offset=1), index = 1*64 + 0*8 + 1 = 65
	if encoded[65] != 1.0 {
		t.Errorf("white knight at b1, index 65: got %f, want 1.0", encoded[65])
	}

	// White knight at g1 (rank=0, file=6):
	//   channel = 1, index = 1*64 + 0*8 + 6 = 70
	if encoded[70] != 1.0 {
		t.Errorf("white knight at g1, index 70: got %f, want 1.0", encoded[70])
	}

	// Black knight at b8 (rank=7, file=1):
	//   channel = 7 (6 + Knight offset 1), index = 7*64 + 7*8 + 1 = 505
	if encoded[505] != 1.0 {
		t.Errorf("black knight at b8, index 505: got %f, want 1.0", encoded[505])
	}

	// White bishop at c1 (rank=0, file=2):
	//   channel = 2 (Bishop, pieceType=3, offset=2), index = 2*64 + 0*8 + 2 = 130
	if encoded[130] != 1.0 {
		t.Errorf("white bishop at c1, index 130: got %f, want 1.0", encoded[130])
	}

	// Black bishop at f8 (rank=7, file=5):
	//   channel = 8 (6 + Bishop offset 2), index = 8*64 + 7*8 + 5 = 573
	if encoded[573] != 1.0 {
		t.Errorf("black bishop at f8, index 573: got %f, want 1.0", encoded[573])
	}

	// Black king at e8 (rank=7, file=4):
	//   channel = 11 (6 + King offset 5), index = 11*64 + 7*8 + 4 = 764
	if encoded[764] != 1.0 {
		t.Errorf("black king at e8, index 764: got %f, want 1.0", encoded[764])
	}

	// Side-to-move plane (channel 12), first element:
	//   index = 12*64 + 0 = 768
	if encoded[768] != 1.0 {
		t.Errorf("side-to-move plane start, index 768: got %f, want 1.0", encoded[768])
	}

	// Verify some empty squares are zero.
	// e4 (rank=3, file=4) should have no piece in any channel 0-11.
	for ch := 0; ch < 12; ch++ {
		idx := ch*64 + 3*8 + 4
		if encoded[idx] != 0.0 {
			t.Errorf("empty square e4, channel %d, index %d: got %f, want 0.0", ch, idx, encoded[idx])
		}
	}

	// Total piece count verification: starting position has 32 pieces.
	// Channels 0-11 should have exactly 32 ones total.
	var pieceCount float32
	for i := 0; i < 12*64; i++ {
		pieceCount += encoded[i]
	}
	if pieceCount != 32.0 {
		t.Errorf("total piece count in channels 0-11: got %f, want 32.0", pieceCount)
	}
}

// TestEncodeBoard_NoCastlingRights verifies that channels 13-16 are all 0.0
// when no castling rights are available.
func TestEncodeBoard_NoCastlingRights(t *testing.T) {
	board := engine.NewBoard()
	board.CastlingRights = 0 // Remove all castling rights.

	encoded := encodeBoard(board)

	for ch := 13; ch <= 16; ch++ {
		for i := 0; i < 64; i++ {
			idx := ch*64 + i
			if encoded[idx] != 0.0 {
				t.Errorf("castling channel %d with no rights, index %d: got %f, want 0.0", ch, idx, encoded[idx])
			}
		}
	}
}

// TestEncodeBoard_PartialCastlingRights verifies individual castling rights
// are encoded independently.
func TestEncodeBoard_PartialCastlingRights(t *testing.T) {
	board := engine.NewBoard()

	// Only white kingside and black queenside.
	board.CastlingRights = engine.CastleWhiteKing | engine.CastleBlackQueen

	encoded := encodeBoard(board)

	// Channel 13 (WK): should be all 1.0.
	for i := 0; i < 64; i++ {
		if encoded[13*64+i] != 1.0 {
			t.Errorf("WK castling channel 13, index %d: got %f, want 1.0", i, encoded[13*64+i])
		}
	}

	// Channel 14 (WQ): should be all 0.0.
	for i := 0; i < 64; i++ {
		if encoded[14*64+i] != 0.0 {
			t.Errorf("WQ castling channel 14, index %d: got %f, want 0.0", i, encoded[14*64+i])
		}
	}

	// Channel 15 (BK): should be all 0.0.
	for i := 0; i < 64; i++ {
		if encoded[15*64+i] != 0.0 {
			t.Errorf("BK castling channel 15, index %d: got %f, want 0.0", i, encoded[15*64+i])
		}
	}

	// Channel 16 (BQ): should be all 1.0.
	for i := 0; i < 64; i++ {
		if encoded[16*64+i] != 1.0 {
			t.Errorf("BQ castling channel 16, index %d: got %f, want 1.0", i, encoded[16*64+i])
		}
	}
}

// TestEncodeBoard_EnPassant verifies en passant encoding on channel 17.
func TestEncodeBoard_EnPassant(t *testing.T) {
	board := engine.NewBoard()

	// Simulate en passant available on e3 (square index = rank*8+file = 2*8+4 = 20).
	// This means a pawn just moved e2-e4, en passant target is e3.
	board.EnPassantSq = int8(engine.NewSquare(4, 2)) // e3, file=4

	encoded := encodeBoard(board)

	// Channel 17: file 4 should have 1.0 for all 8 ranks.
	for rank := 0; rank < 8; rank++ {
		idx := 17*64 + rank*8 + 4
		if encoded[idx] != 1.0 {
			t.Errorf("en passant channel 17, rank=%d, file=4: got %f, want 1.0", rank, encoded[idx])
		}
	}

	// All other files in channel 17 should be 0.0.
	for file := 0; file < 8; file++ {
		if file == 4 {
			continue
		}
		for rank := 0; rank < 8; rank++ {
			idx := 17*64 + rank*8 + file
			if encoded[idx] != 0.0 {
				t.Errorf("en passant channel 17, rank=%d, file=%d: got %f, want 0.0", rank, file, encoded[idx])
			}
		}
	}
}

// TestEncodeBoard_EnPassantFileA verifies en passant on the a-file (file=0).
func TestEncodeBoard_EnPassantFileA(t *testing.T) {
	board := engine.NewBoard()
	board.EnPassantSq = int8(engine.NewSquare(0, 2)) // a3, file=0

	encoded := encodeBoard(board)

	for rank := 0; rank < 8; rank++ {
		idx := 17*64 + rank*8 + 0
		if encoded[idx] != 1.0 {
			t.Errorf("en passant a-file, rank=%d: got %f, want 1.0", rank, encoded[idx])
		}
	}

	// File 1 should be 0.
	for rank := 0; rank < 8; rank++ {
		idx := 17*64 + rank*8 + 1
		if encoded[idx] != 0.0 {
			t.Errorf("en passant non-a-file, rank=%d, file=1: got %f, want 0.0", rank, encoded[idx])
		}
	}
}

// TestEncodeBoard_BlackToMove verifies that channel 12 is all 0.0
// when it is black's turn.
func TestEncodeBoard_BlackToMove(t *testing.T) {
	board := engine.NewBoard()
	board.ActiveColor = engine.Black

	encoded := encodeBoard(board)

	for i := 0; i < 64; i++ {
		if encoded[12*64+i] != 0.0 {
			t.Errorf("black-to-move plane, index %d: got %f, want 0.0", i, encoded[12*64+i])
		}
	}
}

// TestEncodeBoard_EmptyBoard verifies encoding of a board with only kings
// (minimal valid position).
func TestEncodeBoard_EmptyBoard(t *testing.T) {
	board := &engine.Board{
		ActiveColor:    engine.White,
		CastlingRights: 0,
		EnPassantSq:    -1,
	}

	// Place only the two kings.
	board.Squares[engine.NewSquare(4, 0)] = engine.NewPiece(engine.White, engine.King) // e1
	board.Squares[engine.NewSquare(4, 7)] = engine.NewPiece(engine.Black, engine.King) // e8

	encoded := encodeBoard(board)

	// Only 2 pieces should be set in channels 0-11.
	var pieceCount float32
	for i := 0; i < 12*64; i++ {
		pieceCount += encoded[i]
	}
	if pieceCount != 2.0 {
		t.Errorf("king-only board piece count: got %f, want 2.0", pieceCount)
	}

	// White king at e1: channel 5, rank=0, file=4.
	if encoded[5*64+0*8+4] != 1.0 {
		t.Errorf("white king at e1: got %f, want 1.0", encoded[5*64+0*8+4])
	}

	// Black king at e8: channel 11, rank=7, file=4.
	if encoded[11*64+7*8+4] != 1.0 {
		t.Errorf("black king at e8: got %f, want 1.0", encoded[11*64+7*8+4])
	}

	// No castling, no en passant.
	for ch := 13; ch <= 17; ch++ {
		for i := 0; i < 64; i++ {
			if encoded[ch*64+i] != 0.0 {
				t.Errorf("empty board channel %d, index %d: got %f, want 0.0", ch, i, encoded[ch*64+i])
			}
		}
	}
}
