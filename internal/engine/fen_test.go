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

// TestToFENEmptyBoard tests ToFEN with an empty board.
func TestToFENEmptyBoard(t *testing.T) {
	// Create an empty board
	b := &Board{
		Squares:        [64]Piece{}, // All empty
		ActiveColor:    White,
		CastlingRights: 0, // No castling rights for empty board
		EnPassantSq:    -1,
		HalfMoveClock:  0,
		FullMoveNum:    1,
	}

	fen := b.ToFEN()
	expected := "8/8/8/8/8/8/8/8 w - - 0 1"

	if fen != expected {
		t.Errorf("ToFEN() = %q, expected %q", fen, expected)
	}
}

// TestToFENStartingPosition tests ToFEN with the standard starting position.
func TestToFENStartingPosition(t *testing.T) {
	b := NewBoard()
	fen := b.ToFEN()
	expected := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

	if fen != expected {
		t.Errorf("ToFEN() = %q, expected %q", fen, expected)
	}
}

// TestToFENCastlingRights tests ToFEN with various castling rights combinations.
func TestToFENCastlingRights(t *testing.T) {
	tests := []struct {
		name           string
		castlingRights uint8
		expectedCastle string
	}{
		{
			name:           "all castling rights",
			castlingRights: CastleAll,
			expectedCastle: "KQkq",
		},
		{
			name:           "white kingside only",
			castlingRights: CastleWhiteKing,
			expectedCastle: "K",
		},
		{
			name:           "white queenside only",
			castlingRights: CastleWhiteQueen,
			expectedCastle: "Q",
		},
		{
			name:           "black kingside only",
			castlingRights: CastleBlackKing,
			expectedCastle: "k",
		},
		{
			name:           "black queenside only",
			castlingRights: CastleBlackQueen,
			expectedCastle: "q",
		},
		{
			name:           "white kingside and black queenside",
			castlingRights: CastleWhiteKing | CastleBlackQueen,
			expectedCastle: "Kq",
		},
		{
			name:           "no castling rights",
			castlingRights: 0,
			expectedCastle: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBoard()
			b.CastlingRights = tt.castlingRights
			fen := b.ToFEN()

			// Expected FEN with the specific castling rights
			expected := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w " + tt.expectedCastle + " - 0 1"

			if fen != expected {
				t.Errorf("ToFEN() = %q, expected %q", fen, expected)
			}
		})
	}
}

// TestToFENEnPassant tests ToFEN with en passant square set.
func TestToFENEnPassant(t *testing.T) {
	tests := []struct {
		name       string
		epSquare   Square
		expectedEP string
	}{
		{
			name:       "e3 en passant",
			epSquare:   NewSquare(4, 2), // e3
			expectedEP: "e3",
		},
		{
			name:       "d6 en passant",
			epSquare:   NewSquare(3, 5), // d6
			expectedEP: "d6",
		},
		{
			name:       "a3 en passant",
			epSquare:   NewSquare(0, 2), // a3
			expectedEP: "a3",
		},
		{
			name:       "h6 en passant",
			epSquare:   NewSquare(7, 5), // h6
			expectedEP: "h6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBoard()
			b.EnPassantSq = int8(tt.epSquare)
			fen := b.ToFEN()

			// Expected FEN with the specific en passant square
			expected := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq " + tt.expectedEP + " 0 1"

			if fen != expected {
				t.Errorf("ToFEN() = %q, expected %q", fen, expected)
			}
		})
	}
}

// TestToFENHalfmoveFullmove tests ToFEN with various halfmove and fullmove values.
func TestToFENHalfmoveFullmove(t *testing.T) {
	tests := []struct {
		name         string
		halfmove     uint8
		fullmove     uint16
		expectedHalf string
		expectedFull string
	}{
		{
			name:         "initial position",
			halfmove:     0,
			fullmove:     1,
			expectedHalf: "0",
			expectedFull: "1",
		},
		{
			name:         "after 5 moves, no capture/pawn move",
			halfmove:     10,
			fullmove:     6,
			expectedHalf: "10",
			expectedFull: "6",
		},
		{
			name:         "near fifty-move rule",
			halfmove:     48,
			fullmove:     30,
			expectedHalf: "48",
			expectedFull: "30",
		},
		{
			name:         "late game",
			halfmove:     0,
			fullmove:     100,
			expectedHalf: "0",
			expectedFull: "100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBoard()
			b.HalfMoveClock = tt.halfmove
			b.FullMoveNum = tt.fullmove
			fen := b.ToFEN()

			// Expected FEN with the specific halfmove and fullmove numbers
			expected := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - " + tt.expectedHalf + " " + tt.expectedFull

			if fen != expected {
				t.Errorf("ToFEN() = %q, expected %q", fen, expected)
			}
		})
	}
}

// TestToFENRoundTrip tests that FromFEN and ToFEN are inverses of each other.
func TestToFENRoundTrip(t *testing.T) {
	testFENs := []string{
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
		"rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 0 2",
		"rnbqkbnr/pp1ppppp/8/2p5/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2",
	}

	for _, originalFEN := range testFENs {
		t.Run(originalFEN, func(t *testing.T) {
			// Parse FEN to board
			board, err := FromFEN(originalFEN)
			if err != nil {
				t.Fatalf("FromFEN() error: %v", err)
			}

			// Convert back to FEN
			generatedFEN := board.ToFEN()

			// Should match the original
			if generatedFEN != originalFEN {
				t.Errorf("Round trip failed:\n  original: %q\n  generated: %q", originalFEN, generatedFEN)
			}
		})
	}
}
