package engine

import "testing"

// TestCharToPiece tests the charToPiece helper function.
func TestCharToPiece(t *testing.T) {
	tests := []struct {
		name      string
		char      rune
		wantColor Color
		wantType  PieceType
		wantErr   bool
	}{
		// White pieces (uppercase)
		{name: "white pawn", char: 'P', wantColor: White, wantType: Pawn, wantErr: false},
		{name: "white knight", char: 'N', wantColor: White, wantType: Knight, wantErr: false},
		{name: "white bishop", char: 'B', wantColor: White, wantType: Bishop, wantErr: false},
		{name: "white rook", char: 'R', wantColor: White, wantType: Rook, wantErr: false},
		{name: "white queen", char: 'Q', wantColor: White, wantType: Queen, wantErr: false},
		{name: "white king", char: 'K', wantColor: White, wantType: King, wantErr: false},
		// Black pieces (lowercase)
		{name: "black pawn", char: 'p', wantColor: Black, wantType: Pawn, wantErr: false},
		{name: "black knight", char: 'n', wantColor: Black, wantType: Knight, wantErr: false},
		{name: "black bishop", char: 'b', wantColor: Black, wantType: Bishop, wantErr: false},
		{name: "black rook", char: 'r', wantColor: Black, wantType: Rook, wantErr: false},
		{name: "black queen", char: 'q', wantColor: Black, wantType: Queen, wantErr: false},
		{name: "black king", char: 'k', wantColor: Black, wantType: King, wantErr: false},
		// Invalid characters
		{name: "invalid digit", char: '1', wantErr: true},
		{name: "invalid symbol", char: '/', wantErr: true},
		{name: "invalid letter", char: 'X', wantErr: true},
		{name: "invalid lowercase", char: 'x', wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			piece, err := charToPiece(tt.char)
			if tt.wantErr {
				if err == nil {
					t.Errorf("charToPiece('%c') expected error, got nil", tt.char)
				}
			} else {
				if err != nil {
					t.Errorf("charToPiece('%c') unexpected error: %v", tt.char, err)
				}
				if piece.Color() != tt.wantColor {
					t.Errorf("charToPiece('%c') color = %v, want %v", tt.char, piece.Color(), tt.wantColor)
				}
				if piece.Type() != tt.wantType {
					t.Errorf("charToPiece('%c') type = %v, want %v", tt.char, piece.Type(), tt.wantType)
				}
			}
		})
	}
}

// TestParseFEN tests the ParseFEN function (alias for FromFEN).
func TestParseFEN(t *testing.T) {
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
			name:    "invalid FEN - too few parts",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := ParseFEN(tt.fen)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseFEN() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ParseFEN() unexpected error: %v", err)
				}
				if board == nil {
					t.Errorf("ParseFEN() returned nil board")
				}
			}
		})
	}
}

// TestParseFENStartingPosition verifies ParseFEN parses the starting position correctly.
func TestParseFENStartingPosition(t *testing.T) {
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	board, err := ParseFEN(fen)
	if err != nil {
		t.Fatalf("ParseFEN() error: %v", err)
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

// TestParseFENVariousPieceLayouts tests parsing positions with different piece configurations.
func TestParseFENVariousPieceLayouts(t *testing.T) {
	tests := []struct {
		name string
		fen  string
	}{
		{
			name: "kiwipete position",
			fen:  "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		},
		{
			name: "endgame position",
			fen:  "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		},
		{
			name: "position 3",
			fen:  "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
		},
		{
			name: "position 4",
			fen:  "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
		},
		{
			name: "position 5",
			fen:  "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := ParseFEN(tt.fen)
			if err != nil {
				t.Fatalf("ParseFEN() error: %v", err)
			}

			// Verify round-trip: ParseFEN -> ToFEN should produce the same FEN
			generatedFEN := board.ToFEN()
			if generatedFEN != tt.fen {
				t.Errorf("Round trip failed:\n  original:  %q\n  generated: %q", tt.fen, generatedFEN)
			}
		})
	}
}

// TestParseFENCastlingRightsCombinations tests all castling right combinations.
func TestParseFENCastlingRightsCombinations(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		expected uint8
	}{
		{
			name:     "all castling rights (KQkq)",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			expected: CastleAll,
		},
		{
			name:     "white kingside only (K)",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w K - 0 1",
			expected: CastleWhiteKing,
		},
		{
			name:     "white queenside only (Q)",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w Q - 0 1",
			expected: CastleWhiteQueen,
		},
		{
			name:     "black kingside only (k)",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w k - 0 1",
			expected: CastleBlackKing,
		},
		{
			name:     "black queenside only (q)",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w q - 0 1",
			expected: CastleBlackQueen,
		},
		{
			name:     "white kingside and black queenside (Kq)",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w Kq - 0 1",
			expected: CastleWhiteKing | CastleBlackQueen,
		},
		{
			name:     "no castling rights (-)",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w - - 0 1",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := ParseFEN(tt.fen)
			if err != nil {
				t.Fatalf("ParseFEN() error: %v", err)
			}

			if board.CastlingRights != tt.expected {
				t.Errorf("CastlingRights = %v, expected %v", board.CastlingRights, tt.expected)
			}
		})
	}
}

// TestParseFENEnPassantSquares tests parsing various en passant squares.
func TestParseFENEnPassantSquares(t *testing.T) {
	tests := []struct {
		name       string
		fen        string
		expectedEP Square
	}{
		{
			name:       "e3 en passant",
			fen:        "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
			expectedEP: NewSquare(4, 2), // e3
		},
		{
			name:       "d6 en passant",
			fen:        "rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2",
			expectedEP: NewSquare(3, 5), // d6
		},
		{
			name:       "a3 en passant",
			fen:        "rnbqkbnr/1ppppppp/8/8/pP6/8/P1PPPPPP/RNBQKBNR w KQkq a3 0 2",
			expectedEP: NewSquare(0, 2), // a3
		},
		{
			name:       "h6 en passant",
			fen:        "rnbqkbnr/pppppp1p/8/6pP/8/8/PPPPPPP1/RNBQKBNR w KQkq h6 0 2",
			expectedEP: NewSquare(7, 5), // h6
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := ParseFEN(tt.fen)
			if err != nil {
				t.Fatalf("ParseFEN() error: %v", err)
			}

			if board.EnPassantSq != int8(tt.expectedEP) {
				t.Errorf("EnPassantSq = %v (square %v), expected %v (square %v)",
					board.EnPassantSq, Square(board.EnPassantSq), int8(tt.expectedEP), tt.expectedEP)
			}
		})
	}
}

// TestParseFENClockValues tests parsing various halfmove and fullmove clock values.
func TestParseFENClockValues(t *testing.T) {
	tests := []struct {
		name         string
		fen          string
		expectedHalf uint8
		expectedFull uint16
	}{
		{
			name:         "initial position (0 halfmove, 1 fullmove)",
			fen:          "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			expectedHalf: 0,
			expectedFull: 1,
		},
		{
			name:         "mid-game (10 halfmove, 6 fullmove)",
			fen:          "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 10 6",
			expectedHalf: 10,
			expectedFull: 6,
		},
		{
			name:         "approaching fifty-move rule (48 halfmove, 30 fullmove)",
			fen:          "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 48 30",
			expectedHalf: 48,
			expectedFull: 30,
		},
		{
			name:         "late game (0 halfmove, 100 fullmove)",
			fen:          "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 100",
			expectedHalf: 0,
			expectedFull: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := ParseFEN(tt.fen)
			if err != nil {
				t.Fatalf("ParseFEN() error: %v", err)
			}

			if board.HalfMoveClock != tt.expectedHalf {
				t.Errorf("HalfMoveClock = %v, expected %v", board.HalfMoveClock, tt.expectedHalf)
			}

			if board.FullMoveNum != tt.expectedFull {
				t.Errorf("FullMoveNum = %v, expected %v", board.FullMoveNum, tt.expectedFull)
			}
		})
	}
}

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
