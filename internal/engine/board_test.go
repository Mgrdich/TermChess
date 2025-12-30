package engine

import "testing"

func TestNewBoard(t *testing.T) {
	board := NewBoard()

	// Test initial metadata
	t.Run("ActiveColor is White", func(t *testing.T) {
		if board.ActiveColor != White {
			t.Errorf("expected ActiveColor to be White (0), got %d", board.ActiveColor)
		}
	})

	t.Run("CastlingRights has all rights", func(t *testing.T) {
		if board.CastlingRights != CastleAll {
			t.Errorf("expected CastlingRights to be %d (all), got %d", CastleAll, board.CastlingRights)
		}
	})

	t.Run("EnPassantSq is -1", func(t *testing.T) {
		if board.EnPassantSq != -1 {
			t.Errorf("expected EnPassantSq to be -1, got %d", board.EnPassantSq)
		}
	})

	t.Run("HalfMoveClock is 0", func(t *testing.T) {
		if board.HalfMoveClock != 0 {
			t.Errorf("expected HalfMoveClock to be 0, got %d", board.HalfMoveClock)
		}
	})

	t.Run("FullMoveNum is 1", func(t *testing.T) {
		if board.FullMoveNum != 1 {
			t.Errorf("expected FullMoveNum to be 1, got %d", board.FullMoveNum)
		}
	})

	t.Run("Hash is 0", func(t *testing.T) {
		if board.Hash != 0 {
			t.Errorf("expected Hash to be 0, got %d", board.Hash)
		}
	})

	t.Run("History is empty", func(t *testing.T) {
		if len(board.History) != 0 {
			t.Errorf("expected History to be empty, got length %d", len(board.History))
		}
	})
}

func TestNewBoardAllSquaresEmpty(t *testing.T) {
	board := NewBoard()

	for sq := Square(0); sq < 64; sq++ {
		piece := board.PieceAt(sq)
		if !piece.IsEmpty() {
			t.Errorf("expected square %s to be empty, got piece type %d", sq.String(), piece.Type())
		}
	}
}

func TestPieceAtInvalidSquare(t *testing.T) {
	board := NewBoard()

	// Test invalid squares return Empty
	invalidSquares := []Square{NoSquare, -5, 64, 100}
	for _, sq := range invalidSquares {
		piece := board.PieceAt(sq)
		if !piece.IsEmpty() {
			t.Errorf("expected invalid square %d to return empty piece, got type %d", sq, piece.Type())
		}
	}
}

func TestSquareIndexing(t *testing.T) {
	tests := []struct {
		name     string
		file     int
		rank     int
		expected Square
	}{
		{"a1", 0, 0, 0},
		{"b1", 1, 0, 1},
		{"h1", 7, 0, 7},
		{"a2", 0, 1, 8},
		{"a8", 0, 7, 56},
		{"h8", 7, 7, 63},
		{"e4", 4, 3, 28},
		{"d5", 3, 4, 35},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sq := NewSquare(tt.file, tt.rank)
			if sq != tt.expected {
				t.Errorf("NewSquare(%d, %d) = %d, expected %d", tt.file, tt.rank, sq, tt.expected)
			}

			// Verify the square's String() method matches the test name
			if sq.String() != tt.name {
				t.Errorf("Square(%d).String() = %s, expected %s", sq, sq.String(), tt.name)
			}

			// Verify File() and Rank() methods
			if sq.File() != tt.file {
				t.Errorf("Square(%d).File() = %d, expected %d", sq, sq.File(), tt.file)
			}
			if sq.Rank() != tt.rank {
				t.Errorf("Square(%d).Rank() = %d, expected %d", sq, sq.Rank(), tt.rank)
			}
		})
	}
}

func TestSquareValidity(t *testing.T) {
	// Valid squares
	for sq := Square(0); sq <= 63; sq++ {
		if !sq.IsValid() {
			t.Errorf("Square %d should be valid", sq)
		}
	}

	// Invalid squares
	invalidSquares := []Square{-1, -10, 64, 100}
	for _, sq := range invalidSquares {
		if sq.IsValid() {
			t.Errorf("Square %d should be invalid", sq)
		}
	}
}

func TestNewSquareInvalidInputs(t *testing.T) {
	invalidInputs := []struct {
		file int
		rank int
	}{
		{-1, 0},
		{0, -1},
		{8, 0},
		{0, 8},
		{-1, -1},
		{8, 8},
	}

	for _, input := range invalidInputs {
		sq := NewSquare(input.file, input.rank)
		if sq != NoSquare {
			t.Errorf("NewSquare(%d, %d) = %d, expected NoSquare (-1)", input.file, input.rank, sq)
		}
	}
}

func TestPieceCreation(t *testing.T) {
	tests := []struct {
		color     Color
		pieceType PieceType
	}{
		{White, Pawn},
		{White, Knight},
		{White, Bishop},
		{White, Rook},
		{White, Queen},
		{White, King},
		{Black, Pawn},
		{Black, Knight},
		{Black, Bishop},
		{Black, Rook},
		{Black, Queen},
		{Black, King},
	}

	for _, tt := range tests {
		piece := NewPiece(tt.color, tt.pieceType)

		if piece.Color() != tt.color {
			t.Errorf("NewPiece(%d, %d).Color() = %d, expected %d", tt.color, tt.pieceType, piece.Color(), tt.color)
		}

		if piece.Type() != tt.pieceType {
			t.Errorf("NewPiece(%d, %d).Type() = %d, expected %d", tt.color, tt.pieceType, piece.Type(), tt.pieceType)
		}

		if piece.IsEmpty() {
			t.Errorf("NewPiece(%d, %d) should not be empty", tt.color, tt.pieceType)
		}
	}
}

func TestEmptyPiece(t *testing.T) {
	piece := Piece(0)

	if !piece.IsEmpty() {
		t.Error("Piece(0) should be empty")
	}

	if piece.Type() != Empty {
		t.Errorf("Piece(0).Type() = %d, expected Empty (0)", piece.Type())
	}
}

func TestCastlingRightsBits(t *testing.T) {
	// Verify individual castling bits
	if CastleWhiteKing != 1 {
		t.Errorf("CastleWhiteKing = %d, expected 1", CastleWhiteKing)
	}
	if CastleWhiteQueen != 2 {
		t.Errorf("CastleWhiteQueen = %d, expected 2", CastleWhiteQueen)
	}
	if CastleBlackKing != 4 {
		t.Errorf("CastleBlackKing = %d, expected 4", CastleBlackKing)
	}
	if CastleBlackQueen != 8 {
		t.Errorf("CastleBlackQueen = %d, expected 8", CastleBlackQueen)
	}
	if CastleAll != 15 {
		t.Errorf("CastleAll = %d, expected 15", CastleAll)
	}
}
