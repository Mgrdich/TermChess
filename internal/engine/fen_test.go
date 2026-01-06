package engine

import "testing"

func TestFromFEN(t *testing.T) {
	tests := []struct {
		name    string
		fen     string
		wantErr bool
	}{
		{
			name:    "starting position",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			wantErr: false,
		},
		{
			name:    "kiwipete position",
			fen:     "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
			wantErr: false,
		},
		{
			name:    "position with en passant",
			fen:     "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
			wantErr: false,
		},
		{
			name:    "position without castling rights",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w - - 0 1",
			wantErr: false,
		},
		{
			name:    "invalid FEN - too few parts",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq",
			wantErr: true,
		},
		{
			name:    "invalid FEN - invalid piece character",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPXPPP/RNBQKBNR w KQkq - 0 1",
			wantErr: true,
		},
		{
			name:    "invalid FEN - invalid active color",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR x KQkq - 0 1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := FromFEN(tt.fen)
			if tt.wantErr {
				if err == nil {
					t.Errorf("FromFEN() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("FromFEN() unexpected error: %v", err)
				}
				if board == nil {
					t.Errorf("FromFEN() returned nil board")
				}
			}
		})
	}
}

// TestFromFENStartingPosition verifies the starting position is parsed correctly.
func TestFromFENStartingPosition(t *testing.T) {
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	board, err := FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN() error: %v", err)
	}

	// Compare with NewBoard()
	expected := NewBoard()

	// Check active color
	if board.ActiveColor != expected.ActiveColor {
		t.Errorf("ActiveColor = %v, expected %v", board.ActiveColor, expected.ActiveColor)
	}

	// Check castling rights
	if board.CastlingRights != expected.CastlingRights {
		t.Errorf("CastlingRights = %v, expected %v", board.CastlingRights, expected.CastlingRights)
	}

	// Check en passant
	if board.EnPassantSq != expected.EnPassantSq {
		t.Errorf("EnPassantSq = %v, expected %v", board.EnPassantSq, expected.EnPassantSq)
	}

	// Check half-move clock
	if board.HalfMoveClock != expected.HalfMoveClock {
		t.Errorf("HalfMoveClock = %v, expected %v", board.HalfMoveClock, expected.HalfMoveClock)
	}

	// Check full move number
	if board.FullMoveNum != expected.FullMoveNum {
		t.Errorf("FullMoveNum = %v, expected %v", board.FullMoveNum, expected.FullMoveNum)
	}

	// Check all squares
	for sq := Square(0); sq < 64; sq++ {
		if board.Squares[sq] != expected.Squares[sq] {
			t.Errorf("Square %v: piece = %v, expected %v", sq, board.Squares[sq], expected.Squares[sq])
		}
	}
}

// TestFromFENEnPassant verifies en passant square is parsed correctly.
func TestFromFENEnPassant(t *testing.T) {
	fen := "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"
	board, err := FromFEN(fen)
	if err != nil {
		t.Fatalf("FromFEN() error: %v", err)
	}

	expectedEpSq := NewSquare(4, 2) // e3
	if board.EnPassantSq != int8(expectedEpSq) {
		t.Errorf("EnPassantSq = %v (square %v), expected %v (e3)", board.EnPassantSq, Square(board.EnPassantSq), expectedEpSq)
	}
}

// TestFromFENCastlingRights verifies castling rights are parsed correctly.
func TestFromFENCastlingRights(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		expected uint8
	}{
		{
			name:     "all castling rights",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			expected: CastleAll,
		},
		{
			name:     "white only",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQ - 0 1",
			expected: CastleWhiteKing | CastleWhiteQueen,
		},
		{
			name:     "black only",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w kq - 0 1",
			expected: CastleBlackKing | CastleBlackQueen,
		},
		{
			name:     "kingside only",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w Kk - 0 1",
			expected: CastleWhiteKing | CastleBlackKing,
		},
		{
			name:     "queenside only",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w Qq - 0 1",
			expected: CastleWhiteQueen | CastleBlackQueen,
		},
		{
			name:     "no castling rights",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w - - 0 1",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("FromFEN() error: %v", err)
			}

			if board.CastlingRights != tt.expected {
				t.Errorf("CastlingRights = %v, expected %v", board.CastlingRights, tt.expected)
			}
		})
	}
}
