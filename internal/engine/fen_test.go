package engine

import (
	"strings"
	"testing"
)

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

// TestParseFENValidation tests validation of FEN strings with comprehensive error cases.
// This validates Slice 3: FEN Validation & Error Handling.
func TestParseFENValidation(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		wantErr     bool
		errContains string // Substring that should appear in error message
	}{
		// Valid FEN strings (should not error)
		{
			name:    "valid - starting position",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			wantErr: false,
		},
		{
			name:    "valid - position with en passant",
			fen:     "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
			wantErr: false,
		},
		{
			name:    "valid - no castling rights",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w - - 0 1",
			wantErr: false,
		},

		// Field count validation
		{
			name:        "invalid - too few fields (5)",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0",
			wantErr:     true,
			errContains: "6 parts",
		},
		{
			name:        "invalid - too few fields (3)",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq",
			wantErr:     true,
			errContains: "6 parts",
		},
		{
			name:        "invalid - too many fields (7)",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1 extra",
			wantErr:     true,
			errContains: "6 parts",
		},

		// Rank count validation
		{
			name:        "invalid - too few ranks (7)",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP w KQkq - 0 1",
			wantErr:     true,
			errContains: "8 ranks",
		},
		{
			name:        "invalid - too many ranks (9)",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			wantErr:     true,
			errContains: "8 ranks",
		},

		// Invalid piece characters
		{
			name:        "invalid - invalid piece character (X)",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPXPPP/RNBQKBNR w KQkq - 0 1",
			wantErr:     true,
			errContains: "invalid piece character",
		},
		{
			name:        "invalid - invalid piece character (lowercase w in active color position)",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPP1PPP/RNBQKBNR w KQkq - 0 1",
			wantErr:     false, // This is valid
		},
		{
			name:        "invalid - digit 9 in piece placement",
			fen:         "rnbqkbnr/pppppppp/9/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			wantErr:     true,
			errContains: "invalid piece character",
		},
		{
			name:        "invalid - digit 0 in piece placement",
			fen:         "rnbqkbnr/pppppppp/0/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			wantErr:     true,
			errContains: "invalid piece character",
		},

		// Rank square count validation
		{
			name:        "invalid - rank has too many squares (9)",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPPP/RNBQKBNR w KQkq - 0 1",
			wantErr:     true,
			errContains: "squares",
		},
		{
			name:        "invalid - rank has too few squares (7)",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPP/RNBQKBNR w KQkq - 0 1",
			wantErr:     true,
			errContains: "squares",
		},
		{
			name:        "invalid - rank with numbers summing to 9",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/54/RNBQKBNR w KQkq - 0 1",
			wantErr:     true,
			errContains: "squares",
		},

		// Active color validation
		{
			name:        "invalid - active color is 'x'",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR x KQkq - 0 1",
			wantErr:     true,
			errContains: "invalid active color",
		},
		{
			name:        "invalid - active color is 'W' (uppercase)",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR W KQkq - 0 1",
			wantErr:     true,
			errContains: "invalid active color",
		},
		{
			name:        "invalid - active color is 'B' (uppercase)",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR B KQkq - 0 1",
			wantErr:     true,
			errContains: "invalid active color",
		},
		{
			name:        "invalid - active color is empty",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR  KQkq - 0 1",
			wantErr:     true,
			errContains: "6 parts",
		},

		// Castling rights validation
		{
			name:        "invalid - castling rights contains 'X'",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQX - 0 1",
			wantErr:     true,
			errContains: "invalid castling character",
		},
		{
			name:        "invalid - castling rights contains lowercase letters other than kq",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w abc - 0 1",
			wantErr:     true,
			errContains: "invalid castling character",
		},
		{
			name:        "invalid - castling rights contains number",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w K1 - 0 1",
			wantErr:     true,
			errContains: "invalid castling character",
		},
		{
			name:    "valid - castling rights subset (Kq)",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w Kq - 0 1",
			wantErr: false,
		},
		{
			name:    "valid - castling rights subset (Q)",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w Q - 0 1",
			wantErr: false,
		},

		// En passant validation
		{
			name:        "invalid - en passant square outside board (z9)",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq z9 0 1",
			wantErr:     true,
			errContains: "invalid en passant square",
		},
		{
			name:        "invalid - en passant square invalid file (i3)",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq i3 0 1",
			wantErr:     true,
			errContains: "invalid en passant square",
		},
		{
			name:        "invalid - en passant square invalid rank (a0)",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq a0 0 1",
			wantErr:     true,
			errContains: "invalid en passant square",
		},
		{
			name:        "invalid - en passant square invalid rank (a9)",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq a9 0 1",
			wantErr:     true,
			errContains: "invalid en passant square",
		},
		{
			name:        "invalid - en passant square too short (e)",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq e 0 1",
			wantErr:     true,
			errContains: "invalid en passant square",
		},
		{
			name:        "invalid - en passant square too long (e33)",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq e33 0 1",
			wantErr:     true,
			errContains: "invalid en passant square",
		},
		{
			name:    "valid - en passant is dash",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			wantErr: false,
		},

		// Halfmove clock validation
		{
			name:        "invalid - halfmove clock is negative",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - -1 1",
			wantErr:     true,
			errContains: "invalid half-move clock",
		},
		{
			name:        "invalid - halfmove clock is not a number",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - abc 1",
			wantErr:     true,
			errContains: "invalid half-move clock",
		},
		{
			name:        "invalid - halfmove clock is too large (> 255)",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 256 1",
			wantErr:     true,
			errContains: "half-move clock out of range",
		},
		{
			name:    "valid - halfmove clock is 0",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			wantErr: false,
		},
		{
			name:    "valid - halfmove clock is 50",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 50 1",
			wantErr: false,
		},

		// Fullmove number validation
		{
			name:        "invalid - fullmove number is zero",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 0",
			wantErr:     true,
			errContains: "full move number out of range",
		},
		{
			name:        "invalid - fullmove number is negative",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 -1",
			wantErr:     true,
			errContains: "full move number out of range",
		},
		{
			name:        "invalid - fullmove number is not a number",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 abc",
			wantErr:     true,
			errContains: "invalid full move number",
		},
		{
			name:        "invalid - fullmove number is too large (> 65535)",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 65536",
			wantErr:     true,
			errContains: "full move number out of range",
		},
		{
			name:    "valid - fullmove number is 1",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			wantErr: false,
		},
		{
			name:    "valid - fullmove number is large (1000)",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1000",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseFEN(tt.fen)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseFEN() expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ParseFEN() error = %q, should contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ParseFEN() unexpected error: %v", err)
				}
			}
		})
	}
}

// TestRoundTripParseFENToFEN tests FEN string -> ParseFEN -> ToFEN -> verify matches original.
// This test verifies that parsing a FEN string and exporting it produces the same FEN string.
func TestRoundTripParseFENToFEN(t *testing.T) {
	tests := []struct {
		name string
		fen  string
	}{
		{
			name: "starting position",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		},
		{
			name: "empty board",
			fen:  "8/8/8/8/8/8/8/8 w - - 0 1",
		},
		{
			name: "complex position - kiwipete",
			fen:  "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		},
		{
			name: "endgame position",
			fen:  "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		},
		{
			name: "position with en passant - e3",
			fen:  "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
		},
		{
			name: "position with en passant - d6",
			fen:  "rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2",
		},
		{
			name: "position with en passant - a3",
			fen:  "rnbqkbnr/1ppppppp/8/8/pP6/8/P1PPPPPP/RNBQKBNR w KQkq a3 0 2",
		},
		{
			name: "position with en passant - h6",
			fen:  "rnbqkbnr/pppppp1p/8/6pP/8/8/PPPPPPP1/RNBQKBNR w KQkq h6 0 2",
		},
		{
			name: "no castling rights",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w - - 0 1",
		},
		{
			name: "white kingside castling only",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w K - 0 1",
		},
		{
			name: "white queenside castling only",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w Q - 0 1",
		},
		{
			name: "black kingside castling only",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w k - 0 1",
		},
		{
			name: "black queenside castling only",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w q - 0 1",
		},
		{
			name: "partial castling rights - Kq",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w Kq - 0 1",
		},
		{
			name: "partial castling rights - Qk",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w Qk - 0 1",
		},
		{
			name: "black to move",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1",
		},
		{
			name: "non-zero halfmove clock",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 10 6",
		},
		{
			name: "large fullmove number",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 100",
		},
		{
			name: "approaching fifty-move rule",
			fen:  "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 48 30",
		},
		{
			name: "complex position with en passant and partial castling",
			fen:  "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R b Kq e3 5 10",
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
		// Additional famous perft test positions from chessprogramming.org
		{
			name: "perft - position with bishops and rooks",
			fen:  "r6r/1b2k1bq/8/8/7B/8/8/R3K2R b KQ - 3 2",
		},
		{
			name: "perft - en passant promotion",
			fen:  "8/8/8/2k5/2pP4/8/B7/4K3 b - d3 0 3",
		},
		{
			name: "perft - discovered check position",
			fen:  "r3k2r/p1pp1pb1/bn2Qnp1/2qPN3/1p2P3/2N5/PPPBBPPP/R3K2R b KQkq - 3 2",
		},
		{
			name: "perft - checkmate in few moves",
			fen:  "2kr3r/p1ppqpb1/bn2Qnp1/3PN3/1p2P3/2N5/PPPBBPPP/R3K2R b KQ - 3 2",
		},
		{
			name: "perft - complex promotion scenario",
			fen:  "rnb2k1r/pp1Pbppp/2p5/q7/2B5/8/PPPQNnPP/RNB1K2R w KQ - 3 9",
		},
		{
			name: "perft - endgame pawn race",
			fen:  "2r5/3pk3/8/2P5/8/2K5/8/8 w - - 5 4",
		},
		{
			name: "perft - rook endgame",
			fen:  "3k4/3p4/8/K1P4r/8/8/8/8 b - - 0 1",
		},
		{
			name: "perft - minimal pieces",
			fen:  "8/8/4k3/8/2p5/8/B2P2K1/8 w - - 0 1",
		},
		{
			name: "perft - en passant edge case",
			fen:  "8/8/1k6/2b5/2pP4/8/5K2/8 b - d3 0 1",
		},
		{
			name: "perft - kingside castling rights only",
			fen:  "5k2/8/8/8/8/8/8/4K2R w K - 0 1",
		},
		{
			name: "perft - queenside castling rights only",
			fen:  "3k4/8/8/8/8/8/8/R3K3 w Q - 0 1",
		},
		{
			name: "perft - bishops and rooks symmetry",
			fen:  "r3k2r/1b4bq/8/8/8/8/7B/R3K2R w KQkq - 0 1",
		},
		{
			name: "perft - queens and rooks",
			fen:  "r3k2r/8/3Q4/8/8/5q2/8/R3K2R b KQkq - 0 1",
		},
		{
			name: "perft - promotion to rook",
			fen:  "2K2r2/4P3/8/8/8/8/8/3k4 w - - 0 1",
		},
		{
			name: "perft - knight and queen endgame",
			fen:  "8/8/1P2K3/8/2n5/1q6/8/5k2 b - - 0 1",
		},
		{
			name: "perft - pawn promotion on 7th rank",
			fen:  "4k3/1P6/8/8/8/8/K7/8 w - - 0 1",
		},
		{
			name: "perft - pawn promotion corner",
			fen:  "8/P1k5/K7/8/8/8/8/8 w - - 0 1",
		},
		{
			name: "perft - pawn promotion king corner",
			fen:  "K1k5/8/P7/8/8/8/8/8 w - - 0 1",
		},
		{
			name: "perft - black pawn promotion",
			fen:  "8/k1P5/8/1K6/8/8/8/8 w - - 0 1",
		},
		{
			name: "perft - checkmate pattern",
			fen:  "8/8/2k5/5q2/5n2/8/5K2/8 b - - 0 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse FEN to board
			board, err := ParseFEN(tt.fen)
			if err != nil {
				t.Fatalf("ParseFEN() error: %v", err)
			}

			// Convert back to FEN
			generatedFEN := board.ToFEN()

			// Should match the original
			if generatedFEN != tt.fen {
				t.Errorf("Round trip failed:\n  original:  %q\n  generated: %q", tt.fen, generatedFEN)
			}
		})
	}
}

// TestRoundTripBoardToFENToBoard tests Board -> ToFEN -> ParseFEN -> verify board matches.
// This test verifies that creating a board, exporting to FEN, and parsing back produces
// an equivalent board state.
func TestRoundTripBoardToFENToBoard(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *Board
	}{
		{
			name: "standard starting position",
			setup: func() *Board {
				return NewBoard()
			},
		},
		{
			name: "empty board",
			setup: func() *Board {
				return &Board{
					Squares:        [64]Piece{},
					ActiveColor:    White,
					CastlingRights: 0,
					EnPassantSq:    -1,
					HalfMoveClock:  0,
					FullMoveNum:    1,
					Hash:           0,
					History:        []uint64{},
				}
			},
		},
		{
			name: "board with partial castling rights",
			setup: func() *Board {
				b := NewBoard()
				b.CastlingRights = CastleWhiteKing | CastleBlackQueen
				return b
			},
		},
		{
			name: "board with en passant square",
			setup: func() *Board {
				b := NewBoard()
				b.EnPassantSq = int8(NewSquare(4, 2)) // e3
				return b
			},
		},
		{
			name: "board with black to move",
			setup: func() *Board {
				b := NewBoard()
				b.ActiveColor = Black
				return b
			},
		},
		{
			name: "board with non-zero clocks",
			setup: func() *Board {
				b := NewBoard()
				b.HalfMoveClock = 25
				b.FullMoveNum = 50
				return b
			},
		},
		{
			name: "board with no castling rights",
			setup: func() *Board {
				b := NewBoard()
				b.CastlingRights = 0
				return b
			},
		},
		{
			name: "complex board state",
			setup: func() *Board {
				b := NewBoard()
				b.ActiveColor = Black
				b.CastlingRights = CastleBlackKing
				b.EnPassantSq = int8(NewSquare(3, 5)) // d6
				b.HalfMoveClock = 10
				b.FullMoveNum = 25
				return b
			},
		},
		{
			name: "board with custom piece placement",
			setup: func() *Board {
				b := &Board{
					Squares:        [64]Piece{},
					ActiveColor:    White,
					CastlingRights: CastleAll,
					EnPassantSq:    -1,
					HalfMoveClock:  0,
					FullMoveNum:    1,
					Hash:           0,
					History:        []uint64{},
				}
				// Place white king on e1
				b.Squares[NewSquare(4, 0)] = NewPiece(White, King)
				// Place black king on e8
				b.Squares[NewSquare(4, 7)] = NewPiece(Black, King)
				// Place white rook on a1
				b.Squares[NewSquare(0, 0)] = NewPiece(White, Rook)
				// Place black rook on h8
				b.Squares[NewSquare(7, 7)] = NewPiece(Black, Rook)
				return b
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the original board
			original := tt.setup()

			// Export to FEN
			fen := original.ToFEN()

			// Parse back to board
			parsed, err := ParseFEN(fen)
			if err != nil {
				t.Fatalf("ParseFEN() error: %v", err)
			}

			// Compare all board fields
			if parsed.ActiveColor != original.ActiveColor {
				t.Errorf("ActiveColor mismatch: got %v, want %v", parsed.ActiveColor, original.ActiveColor)
			}

			if parsed.CastlingRights != original.CastlingRights {
				t.Errorf("CastlingRights mismatch: got %v, want %v", parsed.CastlingRights, original.CastlingRights)
			}

			if parsed.EnPassantSq != original.EnPassantSq {
				t.Errorf("EnPassantSq mismatch: got %v, want %v", parsed.EnPassantSq, original.EnPassantSq)
			}

			if parsed.HalfMoveClock != original.HalfMoveClock {
				t.Errorf("HalfMoveClock mismatch: got %v, want %v", parsed.HalfMoveClock, original.HalfMoveClock)
			}

			if parsed.FullMoveNum != original.FullMoveNum {
				t.Errorf("FullMoveNum mismatch: got %v, want %v", parsed.FullMoveNum, original.FullMoveNum)
			}

			// Compare all squares
			for sq := Square(0); sq < 64; sq++ {
				if parsed.Squares[sq] != original.Squares[sq] {
					t.Errorf("Square %v mismatch: got %v, want %v", sq, parsed.Squares[sq], original.Squares[sq])
				}
			}
		})
	}
}

// TestRoundTripStartingPosition specifically tests the standard starting position round-trip.
// This verifies that the starting position can be reliably converted between board and FEN representations.
func TestRoundTripStartingPosition(t *testing.T) {
	startingFEN := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

	// Test direction 1: FEN -> Board -> FEN
	t.Run("FEN to Board to FEN", func(t *testing.T) {
		board, err := ParseFEN(startingFEN)
		if err != nil {
			t.Fatalf("ParseFEN() error: %v", err)
		}

		generatedFEN := board.ToFEN()
		if generatedFEN != startingFEN {
			t.Errorf("Round trip failed:\n  original:  %q\n  generated: %q", startingFEN, generatedFEN)
		}
	})

	// Test direction 2: Board -> FEN -> Board
	t.Run("Board to FEN to Board", func(t *testing.T) {
		original := NewBoard()
		fen := original.ToFEN()

		// Verify the FEN matches the standard starting position
		if fen != startingFEN {
			t.Errorf("NewBoard().ToFEN() = %q, want %q", fen, startingFEN)
		}

		// Parse it back
		parsed, err := ParseFEN(fen)
		if err != nil {
			t.Fatalf("ParseFEN() error: %v", err)
		}

		// Compare all fields
		if parsed.ActiveColor != original.ActiveColor {
			t.Errorf("ActiveColor mismatch: got %v, want %v", parsed.ActiveColor, original.ActiveColor)
		}
		if parsed.CastlingRights != original.CastlingRights {
			t.Errorf("CastlingRights mismatch: got %v, want %v", parsed.CastlingRights, original.CastlingRights)
		}
		if parsed.EnPassantSq != original.EnPassantSq {
			t.Errorf("EnPassantSq mismatch: got %v, want %v", parsed.EnPassantSq, original.EnPassantSq)
		}
		if parsed.HalfMoveClock != original.HalfMoveClock {
			t.Errorf("HalfMoveClock mismatch: got %v, want %v", parsed.HalfMoveClock, original.HalfMoveClock)
		}
		if parsed.FullMoveNum != original.FullMoveNum {
			t.Errorf("FullMoveNum mismatch: got %v, want %v", parsed.FullMoveNum, original.FullMoveNum)
		}

		// Compare all squares
		for sq := Square(0); sq < 64; sq++ {
			if parsed.Squares[sq] != original.Squares[sq] {
				t.Errorf("Square %v mismatch: got %v, want %v", sq, parsed.Squares[sq], original.Squares[sq])
			}
		}
	})
}

// TestRoundTripComplexPositions tests round-trip conversion for complex positions with
// various combinations of en passant, partial castling rights, and non-zero clocks.
func TestRoundTripComplexPositions(t *testing.T) {
	tests := []struct {
		name string
		fen  string
	}{
		{
			name: "en passant with partial castling - white kingside only",
			fen:  "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b K e3 0 1",
		},
		{
			name: "en passant with partial castling - black queenside only",
			fen:  "rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR w q d6 0 2",
		},
		{
			name: "en passant with no castling rights",
			fen:  "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b - e3 0 1",
		},
		{
			name: "en passant with non-zero halfmove clock",
			fen:  "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 5 1",
		},
		{
			name: "partial castling Kq with non-zero clocks",
			fen:  "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w Kq - 10 15",
		},
		{
			name: "partial castling Qk with en passant",
			fen:  "rnbqkbnr/pp1ppppp/8/2pP4/8/8/PPP1PPPP/RNBQKBNR w Qk c6 0 2",
		},
		{
			name: "complex endgame with high clocks",
			fen:  "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 b - - 47 100",
		},
		{
			name: "all features combined - en passant, partial castling, black to move, non-zero clocks",
			fen:  "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R b Kq e3 5 10",
		},
		{
			name: "single castling right with en passant and clocks",
			fen:  "rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR w K d6 3 2",
		},
		{
			name: "white queenside castling only with en passant",
			fen:  "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b Q e3 0 1",
		},
		{
			name: "black kingside castling only with en passant",
			fen:  "rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR w k d6 0 2",
		},
		{
			name: "approaching fifty-move rule with partial castling",
			fen:  "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w q - 49 30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse FEN to board
			board, err := ParseFEN(tt.fen)
			if err != nil {
				t.Fatalf("ParseFEN() error: %v", err)
			}

			// Convert back to FEN
			generatedFEN := board.ToFEN()

			// Should match the original
			if generatedFEN != tt.fen {
				t.Errorf("Round trip failed:\n  original:  %q\n  generated: %q", tt.fen, generatedFEN)
			}
		})
	}
}
