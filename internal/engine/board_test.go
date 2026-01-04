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

func TestNewBoardStartingPosition(t *testing.T) {
	board := NewBoard()

	// Test White back rank (rank 1, index 0-7)
	t.Run("White back rank pieces", func(t *testing.T) {
		expectedPieces := []struct {
			square    string
			pieceType PieceType
		}{
			{"a1", Rook},
			{"b1", Knight},
			{"c1", Bishop},
			{"d1", Queen},
			{"e1", King},
			{"f1", Bishop},
			{"g1", Knight},
			{"h1", Rook},
		}

		for i, expected := range expectedPieces {
			sq := Square(i)
			piece := board.PieceAt(sq)
			if piece.Type() != expected.pieceType {
				t.Errorf("expected %s to have piece type %d, got %d", expected.square, expected.pieceType, piece.Type())
			}
			if piece.Color() != White {
				t.Errorf("expected %s to have White piece, got color %d", expected.square, piece.Color())
			}
		}
	})

	// Test White pawns (rank 2, index 8-15)
	t.Run("White pawns on rank 2", func(t *testing.T) {
		for file := 0; file < 8; file++ {
			sq := Square(8 + file)
			piece := board.PieceAt(sq)
			if piece.Type() != Pawn {
				t.Errorf("expected square %s to have Pawn, got piece type %d", sq.String(), piece.Type())
			}
			if piece.Color() != White {
				t.Errorf("expected square %s to have White piece, got color %d", sq.String(), piece.Color())
			}
		}
	})

	// Test empty squares (ranks 3-6, index 16-47)
	t.Run("Empty squares on ranks 3-6", func(t *testing.T) {
		for sq := Square(16); sq < 48; sq++ {
			piece := board.PieceAt(sq)
			if !piece.IsEmpty() {
				t.Errorf("expected square %s to be empty, got piece type %d", sq.String(), piece.Type())
			}
		}
	})

	// Test Black pawns (rank 7, index 48-55)
	t.Run("Black pawns on rank 7", func(t *testing.T) {
		for file := 0; file < 8; file++ {
			sq := Square(48 + file)
			piece := board.PieceAt(sq)
			if piece.Type() != Pawn {
				t.Errorf("expected square %s to have Pawn, got piece type %d", sq.String(), piece.Type())
			}
			if piece.Color() != Black {
				t.Errorf("expected square %s to have Black piece, got color %d", sq.String(), piece.Color())
			}
		}
	})

	// Test Black back rank (rank 8, index 56-63)
	t.Run("Black back rank pieces", func(t *testing.T) {
		expectedPieces := []struct {
			square    string
			pieceType PieceType
		}{
			{"a8", Rook},
			{"b8", Knight},
			{"c8", Bishop},
			{"d8", Queen},
			{"e8", King},
			{"f8", Bishop},
			{"g8", Knight},
			{"h8", Rook},
		}

		for i, expected := range expectedPieces {
			sq := Square(56 + i)
			piece := board.PieceAt(sq)
			if piece.Type() != expected.pieceType {
				t.Errorf("expected %s to have piece type %d, got %d", expected.square, expected.pieceType, piece.Type())
			}
			if piece.Color() != Black {
				t.Errorf("expected %s to have Black piece, got color %d", expected.square, piece.Color())
			}
		}
	})

	// Test total piece count
	t.Run("Total piece count is 32", func(t *testing.T) {
		count := 0
		for sq := Square(0); sq < 64; sq++ {
			if !board.PieceAt(sq).IsEmpty() {
				count++
			}
		}
		if count != 32 {
			t.Errorf("expected 32 pieces, got %d", count)
		}
	})
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

func TestBoardString(t *testing.T) {
	board := NewBoard()
	boardStr := board.String()

	expected := `8 r n b q k b n r
7 p p p p p p p p
6 . . . . . . . .
5 . . . . . . . .
4 . . . . . . . .
3 . . . . . . . .
2 P P P P P P P P
1 R N B Q K B N R
  a b c d e f g h`

	if boardStr != expected {
		t.Errorf("Board.String() output does not match expected.\nGot:\n%s\n\nExpected:\n%s", boardStr, expected)
	}
}

func TestInCheck(t *testing.T) {
	t.Run("starting position - not in check", func(t *testing.T) {
		board := NewBoard()
		if board.InCheck() {
			t.Error("starting position should not be in check for White")
		}

		// Switch to Black's turn
		board.ActiveColor = Black
		if board.InCheck() {
			t.Error("starting position should not be in check for Black")
		}
	})

	// Basic check scenarios - Rook
	t.Run("white king in check from rook horizontal", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 3)] = NewPiece(White, King) // e4
		board.Squares[NewSquare(0, 3)] = NewPiece(Black, Rook) // a4 - attacking horizontally

		if !board.InCheck() {
			t.Error("white king should be in check by black rook on a4 (horizontal)")
		}
	})

	t.Run("white king in check from rook vertical", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King) // e1
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, Rook) // e8 - attacking vertically

		if !board.InCheck() {
			t.Error("white king should be in check by black rook on e8 (vertical)")
		}
	})

	t.Run("black king in check from rook horizontal", func(t *testing.T) {
		board := &Board{ActiveColor: Black}
		board.Squares[NewSquare(3, 4)] = NewPiece(Black, King) // d5
		board.Squares[NewSquare(7, 4)] = NewPiece(White, Rook) // h5 - attacking horizontally

		if !board.InCheck() {
			t.Error("black king should be in check by white rook on h5 (horizontal)")
		}
	})

	t.Run("black king in check from rook vertical", func(t *testing.T) {
		board := &Board{ActiveColor: Black}
		board.Squares[NewSquare(2, 6)] = NewPiece(Black, King) // c7
		board.Squares[NewSquare(2, 0)] = NewPiece(White, Rook) // c1 - attacking vertically

		if !board.InCheck() {
			t.Error("black king should be in check by white rook on c1 (vertical)")
		}
	})

	// Basic check scenarios - Bishop
	t.Run("white king in check from bishop diagonal up-right", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 3)] = NewPiece(White, King)   // e4
		board.Squares[NewSquare(7, 6)] = NewPiece(Black, Bishop) // h7 - attacking diagonally

		if !board.InCheck() {
			t.Error("white king should be in check by black bishop on h7")
		}
	})

	t.Run("white king in check from bishop diagonal down-left", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 4)] = NewPiece(White, King)   // e5
		board.Squares[NewSquare(0, 0)] = NewPiece(Black, Bishop) // a1 - attacking diagonally

		if !board.InCheck() {
			t.Error("white king should be in check by black bishop on a1")
		}
	})

	t.Run("black king in check from bishop diagonal up-left", func(t *testing.T) {
		board := &Board{ActiveColor: Black}
		board.Squares[NewSquare(5, 2)] = NewPiece(Black, King)   // f3
		board.Squares[NewSquare(2, 5)] = NewPiece(White, Bishop) // c6 - attacking diagonally

		if !board.InCheck() {
			t.Error("black king should be in check by white bishop on c6")
		}
	})

	t.Run("black king in check from bishop diagonal down-right", func(t *testing.T) {
		board := &Board{ActiveColor: Black}
		board.Squares[NewSquare(3, 5)] = NewPiece(Black, King)   // d6
		board.Squares[NewSquare(6, 2)] = NewPiece(White, Bishop) // g3 - attacking diagonally

		if !board.InCheck() {
			t.Error("black king should be in check by white bishop on g3")
		}
	})

	// Basic check scenarios - Queen
	t.Run("white king in check from queen horizontal", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 3)] = NewPiece(White, King)  // e4
		board.Squares[NewSquare(1, 3)] = NewPiece(Black, Queen) // b4 - attacking horizontally

		if !board.InCheck() {
			t.Error("white king should be in check by black queen on b4 (horizontal)")
		}
	})

	t.Run("white king in check from queen vertical", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King)  // e1
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, Queen) // e8 - attacking vertically

		if !board.InCheck() {
			t.Error("white king should be in check by black queen on e8 (vertical)")
		}
	})

	t.Run("white king in check from queen diagonal", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(3, 3)] = NewPiece(White, King)  // d4
		board.Squares[NewSquare(6, 6)] = NewPiece(Black, Queen) // g7 - attacking diagonally

		if !board.InCheck() {
			t.Error("white king should be in check by black queen on g7 (diagonal)")
		}
	})

	t.Run("black king in check from queen horizontal", func(t *testing.T) {
		board := &Board{ActiveColor: Black}
		board.Squares[NewSquare(4, 5)] = NewPiece(Black, King)  // e6
		board.Squares[NewSquare(7, 5)] = NewPiece(White, Queen) // h6 - attacking horizontally

		if !board.InCheck() {
			t.Error("black king should be in check by white queen on h6 (horizontal)")
		}
	})

	t.Run("black king in check from queen vertical", func(t *testing.T) {
		board := &Board{ActiveColor: Black}
		board.Squares[NewSquare(3, 7)] = NewPiece(Black, King)  // d8
		board.Squares[NewSquare(3, 2)] = NewPiece(White, Queen) // d3 - attacking vertically

		if !board.InCheck() {
			t.Error("black king should be in check by white queen on d3 (vertical)")
		}
	})

	t.Run("black king in check from queen diagonal", func(t *testing.T) {
		board := &Board{ActiveColor: Black}
		board.Squares[NewSquare(5, 5)] = NewPiece(Black, King)  // f6
		board.Squares[NewSquare(2, 2)] = NewPiece(White, Queen) // c3 - attacking diagonally

		if !board.InCheck() {
			t.Error("black king should be in check by white queen on c3 (diagonal)")
		}
	})

	// Basic check scenarios - Knight
	t.Run("white king in check from knight", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 3)] = NewPiece(White, King)   // e4
		board.Squares[NewSquare(5, 5)] = NewPiece(Black, Knight) // f6 - L-shape attack

		if !board.InCheck() {
			t.Error("white king should be in check by black knight on f6")
		}
	})

	t.Run("white king in check from knight different L-shape", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 4)] = NewPiece(White, King)   // e5
		board.Squares[NewSquare(2, 3)] = NewPiece(Black, Knight) // c4 - L-shape attack

		if !board.InCheck() {
			t.Error("white king should be in check by black knight on c4")
		}
	})

	t.Run("black king in check from knight", func(t *testing.T) {
		board := &Board{ActiveColor: Black}
		board.Squares[NewSquare(3, 6)] = NewPiece(Black, King)   // d7
		board.Squares[NewSquare(4, 4)] = NewPiece(White, Knight) // e5 - L-shape attack

		if !board.InCheck() {
			t.Error("black king should be in check by white knight on e5")
		}
	})

	t.Run("black king in check from knight different L-shape", func(t *testing.T) {
		board := &Board{ActiveColor: Black}
		board.Squares[NewSquare(5, 5)] = NewPiece(Black, King)   // f6
		board.Squares[NewSquare(6, 3)] = NewPiece(White, Knight) // g4 - L-shape attack

		if !board.InCheck() {
			t.Error("black king should be in check by white knight on g4")
		}
	})

	// Basic check scenarios - Pawn
	t.Run("white king in check from pawn diagonal left", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 3)] = NewPiece(White, King) // e4
		board.Squares[NewSquare(3, 4)] = NewPiece(Black, Pawn) // d5 - attacking diagonally

		if !board.InCheck() {
			t.Error("white king should be in check by black pawn on d5")
		}
	})

	t.Run("white king in check from pawn diagonal right", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 3)] = NewPiece(White, King) // e4
		board.Squares[NewSquare(5, 4)] = NewPiece(Black, Pawn) // f5 - attacking diagonally

		if !board.InCheck() {
			t.Error("white king should be in check by black pawn on f5")
		}
	})

	t.Run("black king in check from pawn diagonal left", func(t *testing.T) {
		board := &Board{ActiveColor: Black}
		board.Squares[NewSquare(4, 3)] = NewPiece(Black, King) // e4
		board.Squares[NewSquare(3, 2)] = NewPiece(White, Pawn) // d3 - attacking diagonally

		if !board.InCheck() {
			t.Error("black king should be in check by white pawn on d3")
		}
	})

	t.Run("black king in check from pawn diagonal right", func(t *testing.T) {
		board := &Board{ActiveColor: Black}
		board.Squares[NewSquare(4, 3)] = NewPiece(Black, King) // e4
		board.Squares[NewSquare(5, 2)] = NewPiece(White, Pawn) // f3 - attacking diagonally

		if !board.InCheck() {
			t.Error("black king should be in check by white pawn on f3")
		}
	})

	// No check scenarios
	t.Run("king not in check with pieces nearby but not attacking", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 3)] = NewPiece(White, King)   // e4 (4,3)
		board.Squares[NewSquare(6, 4)] = NewPiece(Black, Rook)   // g5 (6,4) - not on same file/rank, not diagonal
		board.Squares[NewSquare(1, 1)] = NewPiece(Black, Bishop) // b2 (1,1) - file diff=3, rank diff=2, not diagonal
		board.Squares[NewSquare(7, 5)] = NewPiece(Black, Queen)  // h6 (7,5) - file diff=3, rank diff=2, not diagonal/file/rank

		if board.InCheck() {
			t.Error("white king should not be in check - pieces not attacking")
		}
	})

	t.Run("king not in check - rook attack blocked by own piece", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King) // e1
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, Rook) // e8
		board.Squares[NewSquare(4, 1)] = NewPiece(White, Pawn) // e2 - blocking

		if board.InCheck() {
			t.Error("white king should not be in check - rook blocked by pawn")
		}
	})

	t.Run("king not in check - rook attack blocked by enemy piece", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King)   // e1
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, Rook)   // e8
		board.Squares[NewSquare(4, 3)] = NewPiece(Black, Knight) // e4 - blocking (even enemy blocks)

		if board.InCheck() {
			t.Error("white king should not be in check - rook blocked by knight")
		}
	})

	t.Run("king not in check - bishop attack blocked", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 3)] = NewPiece(White, King)   // e4
		board.Squares[NewSquare(7, 6)] = NewPiece(Black, Bishop) // h7
		board.Squares[NewSquare(5, 4)] = NewPiece(White, Pawn)   // f5 - blocking diagonal

		if board.InCheck() {
			t.Error("white king should not be in check - bishop blocked by pawn")
		}
	})

	t.Run("king not in check - queen attack blocked", func(t *testing.T) {
		board := &Board{ActiveColor: Black}
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, King)   // e8
		board.Squares[NewSquare(4, 0)] = NewPiece(White, Queen)  // e1
		board.Squares[NewSquare(4, 4)] = NewPiece(Black, Bishop) // e5 - blocking

		if board.InCheck() {
			t.Error("black king should not be in check - queen blocked by bishop")
		}
	})

	t.Run("king not in check - pawn directly in front does not attack", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 3)] = NewPiece(White, King) // e4
		board.Squares[NewSquare(4, 4)] = NewPiece(Black, Pawn) // e5 - directly in front, not diagonal

		if board.InCheck() {
			t.Error("white king should not be in check - pawn doesn't attack forward")
		}
	})

	t.Run("king not in check - knight not in L-shape position", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 3)] = NewPiece(White, King)   // e4
		board.Squares[NewSquare(4, 5)] = NewPiece(Black, Knight) // e6 - not in L-shape

		if board.InCheck() {
			t.Error("white king should not be in check - knight not in attacking position")
		}
	})

	t.Run("own pieces do not put king in check", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 3)] = NewPiece(White, King)  // e4
		board.Squares[NewSquare(4, 7)] = NewPiece(White, Queen) // e8 - same color
		board.Squares[NewSquare(0, 3)] = NewPiece(White, Rook)  // a4 - same color

		if board.InCheck() {
			t.Error("white king should not be in check by own pieces")
		}
	})

	// Double check scenarios
	t.Run("double check - queen and knight", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 3)] = NewPiece(White, King)   // e4
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, Queen)  // e8 - vertical attack
		board.Squares[NewSquare(5, 5)] = NewPiece(Black, Knight) // f6 - L-shape attack

		if !board.InCheck() {
			t.Error("white king should be in check (double check by queen and knight)")
		}
	})

	t.Run("double check - two rooks", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 3)] = NewPiece(White, King) // e4
		board.Squares[NewSquare(0, 3)] = NewPiece(Black, Rook) // a4 - horizontal attack
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, Rook) // e8 - vertical attack

		if !board.InCheck() {
			t.Error("white king should be in check (double check by two rooks)")
		}
	})

	t.Run("double check - bishop and rook", func(t *testing.T) {
		board := &Board{ActiveColor: Black}
		board.Squares[NewSquare(3, 3)] = NewPiece(Black, King)   // d4
		board.Squares[NewSquare(0, 0)] = NewPiece(White, Bishop) // a1 - diagonal attack
		board.Squares[NewSquare(7, 3)] = NewPiece(White, Rook)   // h4 - horizontal attack

		if !board.InCheck() {
			t.Error("black king should be in check (double check by bishop and rook)")
		}
	})

	// Edge cases - King in corner positions
	t.Run("king in corner a1 - in check from rook", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(0, 0)] = NewPiece(White, King) // a1 - corner
		board.Squares[NewSquare(7, 0)] = NewPiece(Black, Rook) // h1 - horizontal attack

		if !board.InCheck() {
			t.Error("white king in corner a1 should be in check by rook on h1")
		}
	})

	t.Run("king in corner h8 - in check from bishop", func(t *testing.T) {
		board := &Board{ActiveColor: Black}
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King)   // h8 - corner
		board.Squares[NewSquare(4, 4)] = NewPiece(White, Bishop) // e5 - diagonal attack

		if !board.InCheck() {
			t.Error("black king in corner h8 should be in check by bishop on e5")
		}
	})

	t.Run("king in corner a8 - in check from knight", func(t *testing.T) {
		board := &Board{ActiveColor: Black}
		board.Squares[NewSquare(0, 7)] = NewPiece(Black, King)   // a8 - corner
		board.Squares[NewSquare(1, 5)] = NewPiece(White, Knight) // b6 - L-shape attack

		if !board.InCheck() {
			t.Error("black king in corner a8 should be in check by knight on b6")
		}
	})

	t.Run("king in corner h1 - not in check", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(7, 0)] = NewPiece(White, King)  // h1 - corner
		board.Squares[NewSquare(4, 4)] = NewPiece(Black, Queen) // e5 - not attacking corner

		if board.InCheck() {
			t.Error("white king in corner h1 should not be in check")
		}
	})

	// Edge cases - King on edge of board
	t.Run("king on edge a4 - in check from rook", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(0, 3)] = NewPiece(White, King) // a4 - edge
		board.Squares[NewSquare(0, 7)] = NewPiece(Black, Rook) // a8 - vertical attack

		if !board.InCheck() {
			t.Error("white king on edge a4 should be in check by rook on a8")
		}
	})

	t.Run("king on edge e1 - in check from queen", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King)  // e1 - edge
		board.Squares[NewSquare(7, 3)] = NewPiece(Black, Queen) // h4 - diagonal attack

		if !board.InCheck() {
			t.Error("white king on edge e1 should be in check by queen on h4")
		}
	})

	t.Run("king on edge h5 - not in check", func(t *testing.T) {
		board := &Board{ActiveColor: Black}
		board.Squares[NewSquare(7, 4)] = NewPiece(Black, King)   // h5 - edge
		board.Squares[NewSquare(4, 6)] = NewPiece(White, Bishop) // e7 - not on diagonal with h5

		if board.InCheck() {
			t.Error("black king on edge h5 should not be in check")
		}
	})

	t.Run("no king on board returns false", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		// No pieces on board

		if board.InCheck() {
			t.Error("board with no king should return false for InCheck")
		}
	})

	// Discovered check scenario
	t.Run("discovered check simulation - piece moved away reveals attack", func(t *testing.T) {
		// This tests the concept: if we move a piece that was blocking,
		// the king becomes in check. We simulate the position after the move.
		board := &Board{ActiveColor: White}
		// Setup: White king on e1, Black rook on e8, but nothing blocking
		// (as if a piece was just moved away from e-file)
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King) // e1
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, Rook) // e8 - now has clear line

		if !board.InCheck() {
			t.Error("white king should be in check after blocking piece moved (discovered check)")
		}
	})

	t.Run("discovered check simulation - diagonal", func(t *testing.T) {
		// Simulate position after a piece moved off the diagonal
		board := &Board{ActiveColor: Black}
		board.Squares[NewSquare(5, 5)] = NewPiece(Black, King)   // f6
		board.Squares[NewSquare(2, 2)] = NewPiece(White, Bishop) // c3 - now has clear diagonal

		if !board.InCheck() {
			t.Error("black king should be in check after blocking piece moved (discovered check)")
		}
	})
}

func TestApplyMoveCastling(t *testing.T) {
	// Test white kingside castling (O-O): King e1 -> g1, Rook h1 -> f1
	t.Run("white kingside castling moves king and rook", func(t *testing.T) {
		board := &Board{ActiveColor: White, CastlingRights: CastleAll}
		// Setup: King on e1, Rook on h1, squares f1 and g1 empty
		e1 := NewSquare(4, 0) // 4
		f1 := NewSquare(5, 0) // 5
		g1 := NewSquare(6, 0) // 6
		h1 := NewSquare(7, 0) // 7

		board.Squares[e1] = NewPiece(White, King)
		board.Squares[h1] = NewPiece(White, Rook)

		// Apply castling move (king moves from e1 to g1)
		move := Move{From: e1, To: g1}
		board.applyMove(move)

		// Verify king is on g1
		if board.Squares[g1].Type() != King || board.Squares[g1].Color() != White {
			t.Errorf("expected white king on g1 after kingside castling, got %v", board.Squares[g1])
		}

		// Verify rook is on f1
		if board.Squares[f1].Type() != Rook || board.Squares[f1].Color() != White {
			t.Errorf("expected white rook on f1 after kingside castling, got %v", board.Squares[f1])
		}

		// Verify e1 is empty
		if !board.Squares[e1].IsEmpty() {
			t.Errorf("expected e1 to be empty after castling, got %v", board.Squares[e1])
		}

		// Verify h1 is empty
		if !board.Squares[h1].IsEmpty() {
			t.Errorf("expected h1 to be empty after castling, got %v", board.Squares[h1])
		}
	})

	// Test white queenside castling (O-O-O): King e1 -> c1, Rook a1 -> d1
	t.Run("white queenside castling moves king and rook", func(t *testing.T) {
		board := &Board{ActiveColor: White, CastlingRights: CastleAll}
		// Setup: King on e1, Rook on a1, squares b1, c1, d1 empty
		a1 := NewSquare(0, 0) // 0
		c1 := NewSquare(2, 0) // 2
		d1 := NewSquare(3, 0) // 3
		e1 := NewSquare(4, 0) // 4

		board.Squares[e1] = NewPiece(White, King)
		board.Squares[a1] = NewPiece(White, Rook)

		// Apply castling move (king moves from e1 to c1)
		move := Move{From: e1, To: c1}
		board.applyMove(move)

		// Verify king is on c1
		if board.Squares[c1].Type() != King || board.Squares[c1].Color() != White {
			t.Errorf("expected white king on c1 after queenside castling, got %v", board.Squares[c1])
		}

		// Verify rook is on d1
		if board.Squares[d1].Type() != Rook || board.Squares[d1].Color() != White {
			t.Errorf("expected white rook on d1 after queenside castling, got %v", board.Squares[d1])
		}

		// Verify e1 is empty
		if !board.Squares[e1].IsEmpty() {
			t.Errorf("expected e1 to be empty after castling, got %v", board.Squares[e1])
		}

		// Verify a1 is empty
		if !board.Squares[a1].IsEmpty() {
			t.Errorf("expected a1 to be empty after castling, got %v", board.Squares[a1])
		}
	})

	// Test black kingside castling (O-O): King e8 -> g8, Rook h8 -> f8
	t.Run("black kingside castling moves king and rook", func(t *testing.T) {
		board := &Board{ActiveColor: Black, CastlingRights: CastleAll}
		// Setup: King on e8, Rook on h8, squares f8 and g8 empty
		e8 := NewSquare(4, 7) // 60
		f8 := NewSquare(5, 7) // 61
		g8 := NewSquare(6, 7) // 62
		h8 := NewSquare(7, 7) // 63

		board.Squares[e8] = NewPiece(Black, King)
		board.Squares[h8] = NewPiece(Black, Rook)

		// Apply castling move (king moves from e8 to g8)
		move := Move{From: e8, To: g8}
		board.applyMove(move)

		// Verify king is on g8
		if board.Squares[g8].Type() != King || board.Squares[g8].Color() != Black {
			t.Errorf("expected black king on g8 after kingside castling, got %v", board.Squares[g8])
		}

		// Verify rook is on f8
		if board.Squares[f8].Type() != Rook || board.Squares[f8].Color() != Black {
			t.Errorf("expected black rook on f8 after kingside castling, got %v", board.Squares[f8])
		}

		// Verify e8 is empty
		if !board.Squares[e8].IsEmpty() {
			t.Errorf("expected e8 to be empty after castling, got %v", board.Squares[e8])
		}

		// Verify h8 is empty
		if !board.Squares[h8].IsEmpty() {
			t.Errorf("expected h8 to be empty after castling, got %v", board.Squares[h8])
		}
	})

	// Test black queenside castling (O-O-O): King e8 -> c8, Rook a8 -> d8
	t.Run("black queenside castling moves king and rook", func(t *testing.T) {
		board := &Board{ActiveColor: Black, CastlingRights: CastleAll}
		// Setup: King on e8, Rook on a8, squares b8, c8, d8 empty
		a8 := NewSquare(0, 7) // 56
		c8 := NewSquare(2, 7) // 58
		d8 := NewSquare(3, 7) // 59
		e8 := NewSquare(4, 7) // 60

		board.Squares[e8] = NewPiece(Black, King)
		board.Squares[a8] = NewPiece(Black, Rook)

		// Apply castling move (king moves from e8 to c8)
		move := Move{From: e8, To: c8}
		board.applyMove(move)

		// Verify king is on c8
		if board.Squares[c8].Type() != King || board.Squares[c8].Color() != Black {
			t.Errorf("expected black king on c8 after queenside castling, got %v", board.Squares[c8])
		}

		// Verify rook is on d8
		if board.Squares[d8].Type() != Rook || board.Squares[d8].Color() != Black {
			t.Errorf("expected black rook on d8 after queenside castling, got %v", board.Squares[d8])
		}

		// Verify e8 is empty
		if !board.Squares[e8].IsEmpty() {
			t.Errorf("expected e8 to be empty after castling, got %v", board.Squares[e8])
		}

		// Verify a8 is empty
		if !board.Squares[a8].IsEmpty() {
			t.Errorf("expected a8 to be empty after castling, got %v", board.Squares[a8])
		}
	})

	// Test that normal king moves don't trigger rook movement
	t.Run("normal king move does not move rook", func(t *testing.T) {
		board := &Board{ActiveColor: White, CastlingRights: CastleAll}
		e1 := NewSquare(4, 0)
		f1 := NewSquare(5, 0)
		h1 := NewSquare(7, 0)

		board.Squares[e1] = NewPiece(White, King)
		board.Squares[h1] = NewPiece(White, Rook)

		// Apply a normal king move (one square to the right)
		move := Move{From: e1, To: f1}
		board.applyMove(move)

		// Verify king is on f1
		if board.Squares[f1].Type() != King || board.Squares[f1].Color() != White {
			t.Errorf("expected white king on f1 after normal move, got %v", board.Squares[f1])
		}

		// Verify rook is still on h1 (not moved)
		if board.Squares[h1].Type() != Rook || board.Squares[h1].Color() != White {
			t.Errorf("expected white rook to remain on h1 after normal king move, got %v", board.Squares[h1])
		}

		// Verify e1 is empty
		if !board.Squares[e1].IsEmpty() {
			t.Errorf("expected e1 to be empty after king move, got %v", board.Squares[e1])
		}
	})
}

func TestEnPassantSquare(t *testing.T) {
	// Test white pawn e2->e4 sets en passant square to e3
	t.Run("white pawn e2-e4 sets en passant to e3", func(t *testing.T) {
		board := NewBoard()
		e2 := NewSquare(4, 1) // e2
		e4 := NewSquare(4, 3) // e4
		e3 := NewSquare(4, 2) // e3 - expected en passant square

		move := Move{From: e2, To: e4}
		board.applyMove(move)

		if board.EnPassantSq != int8(e3) {
			t.Errorf("expected EnPassantSq to be %d (e3), got %d", e3, board.EnPassantSq)
		}

		// Verify the en passant square has correct file and rank
		epSquare := Square(board.EnPassantSq)
		if epSquare.File() != 4 {
			t.Errorf("expected en passant file to be 4 (e-file), got %d", epSquare.File())
		}
		if epSquare.Rank() != 2 {
			t.Errorf("expected en passant rank to be 2 (rank 3), got %d", epSquare.Rank())
		}
	})

	// Test black pawn d7->d5 sets en passant square to d6
	t.Run("black pawn d7-d5 sets en passant to d6", func(t *testing.T) {
		board := NewBoard()
		board.ActiveColor = Black
		d7 := NewSquare(3, 6) // d7
		d5 := NewSquare(3, 4) // d5
		d6 := NewSquare(3, 5) // d6 - expected en passant square

		move := Move{From: d7, To: d5}
		board.applyMove(move)

		if board.EnPassantSq != int8(d6) {
			t.Errorf("expected EnPassantSq to be %d (d6), got %d", d6, board.EnPassantSq)
		}

		// Verify the en passant square has correct file and rank
		epSquare := Square(board.EnPassantSq)
		if epSquare.File() != 3 {
			t.Errorf("expected en passant file to be 3 (d-file), got %d", epSquare.File())
		}
		if epSquare.Rank() != 5 {
			t.Errorf("expected en passant rank to be 5 (rank 6), got %d", epSquare.Rank())
		}
	})

	// Test single square pawn move clears en passant square
	t.Run("single square pawn move clears en passant", func(t *testing.T) {
		board := NewBoard()
		// First set up an en passant square
		board.EnPassantSq = int8(NewSquare(4, 2)) // e3

		e2 := NewSquare(4, 1) // e2
		e3 := NewSquare(4, 2) // e3

		move := Move{From: e2, To: e3}
		board.applyMove(move)

		if board.EnPassantSq != -1 {
			t.Errorf("expected EnPassantSq to be -1 after single square pawn move, got %d", board.EnPassantSq)
		}
	})

	// Test non-pawn moves clear en passant square
	t.Run("non-pawn move clears en passant", func(t *testing.T) {
		board := NewBoard()
		// First set up an en passant square
		board.EnPassantSq = int8(NewSquare(4, 2)) // e3

		// Move the knight
		g1 := NewSquare(6, 0) // g1
		f3 := NewSquare(5, 2) // f3

		move := Move{From: g1, To: f3}
		board.applyMove(move)

		if board.EnPassantSq != -1 {
			t.Errorf("expected EnPassantSq to be -1 after knight move, got %d", board.EnPassantSq)
		}
	})

	// Test all files for white pawn double push
	t.Run("white pawn double push sets correct en passant for all files", func(t *testing.T) {
		for file := 0; file < 8; file++ {
			board := NewBoard()
			from := NewSquare(file, 1) // rank 2
			to := NewSquare(file, 3)   // rank 4
			expectedEP := NewSquare(file, 2) // rank 3

			move := Move{From: from, To: to}
			board.applyMove(move)

			if board.EnPassantSq != int8(expectedEP) {
				t.Errorf("file %d: expected EnPassantSq to be %d, got %d", file, expectedEP, board.EnPassantSq)
			}
		}
	})

	// Test all files for black pawn double push
	t.Run("black pawn double push sets correct en passant for all files", func(t *testing.T) {
		for file := 0; file < 8; file++ {
			board := NewBoard()
			board.ActiveColor = Black
			from := NewSquare(file, 6) // rank 7
			to := NewSquare(file, 4)   // rank 5
			expectedEP := NewSquare(file, 5) // rank 6

			move := Move{From: from, To: to}
			board.applyMove(move)

			if board.EnPassantSq != int8(expectedEP) {
				t.Errorf("file %d: expected EnPassantSq to be %d, got %d", file, expectedEP, board.EnPassantSq)
			}
		}
	})

	// Test king move clears en passant
	t.Run("king move clears en passant", func(t *testing.T) {
		board := &Board{ActiveColor: White, CastlingRights: CastleAll}
		board.EnPassantSq = int8(NewSquare(4, 2)) // e3
		e1 := NewSquare(4, 0)
		e2 := NewSquare(4, 1)
		board.Squares[e1] = NewPiece(White, King)

		move := Move{From: e1, To: e2}
		board.applyMove(move)

		if board.EnPassantSq != -1 {
			t.Errorf("expected EnPassantSq to be -1 after king move, got %d", board.EnPassantSq)
		}
	})

	// Test rook move clears en passant
	t.Run("rook move clears en passant", func(t *testing.T) {
		board := &Board{ActiveColor: White, CastlingRights: CastleAll}
		board.EnPassantSq = int8(NewSquare(4, 2)) // e3
		a1 := NewSquare(0, 0)
		a3 := NewSquare(0, 2)
		board.Squares[a1] = NewPiece(White, Rook)

		move := Move{From: a1, To: a3}
		board.applyMove(move)

		if board.EnPassantSq != -1 {
			t.Errorf("expected EnPassantSq to be -1 after rook move, got %d", board.EnPassantSq)
		}
	})

	// Test capture clears en passant (non-pawn)
	t.Run("capture by non-pawn clears en passant", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.EnPassantSq = int8(NewSquare(4, 2)) // e3
		c3 := NewSquare(2, 2)
		d5 := NewSquare(3, 4)
		board.Squares[c3] = NewPiece(White, Bishop)
		board.Squares[d5] = NewPiece(Black, Pawn)

		move := Move{From: c3, To: d5}
		board.applyMove(move)

		if board.EnPassantSq != -1 {
			t.Errorf("expected EnPassantSq to be -1 after bishop capture, got %d", board.EnPassantSq)
		}
	})

	// Test pawn capture clears en passant (diagonal move, not double push)
	t.Run("pawn capture clears en passant", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		board.EnPassantSq = int8(NewSquare(4, 2)) // e3
		d4 := NewSquare(3, 3)
		e5 := NewSquare(4, 4)
		board.Squares[d4] = NewPiece(White, Pawn)
		board.Squares[e5] = NewPiece(Black, Pawn)

		move := Move{From: d4, To: e5}
		board.applyMove(move)

		if board.EnPassantSq != -1 {
			t.Errorf("expected EnPassantSq to be -1 after pawn capture, got %d", board.EnPassantSq)
		}
	})
}

func TestCastlingRightsUpdate(t *testing.T) {
	// Test that white king move removes both white castling rights
	t.Run("white king move removes both white castling rights", func(t *testing.T) {
		board := &Board{ActiveColor: White, CastlingRights: CastleAll}
		e1 := NewSquare(4, 0) // e1
		e2 := NewSquare(4, 1) // e2

		board.Squares[e1] = NewPiece(White, King)

		move := Move{From: e1, To: e2}
		board.applyMove(move)

		// White castling rights should be removed
		if board.CastlingRights&CastleWhiteKing != 0 {
			t.Error("CastleWhiteKing should be removed after white king moves")
		}
		if board.CastlingRights&CastleWhiteQueen != 0 {
			t.Error("CastleWhiteQueen should be removed after white king moves")
		}
		// Black castling rights should remain
		if board.CastlingRights&CastleBlackKing == 0 {
			t.Error("CastleBlackKing should remain after white king moves")
		}
		if board.CastlingRights&CastleBlackQueen == 0 {
			t.Error("CastleBlackQueen should remain after white king moves")
		}
	})

	// Test that black king move removes both black castling rights
	t.Run("black king move removes both black castling rights", func(t *testing.T) {
		board := &Board{ActiveColor: Black, CastlingRights: CastleAll}
		e8 := NewSquare(4, 7) // e8
		e7 := NewSquare(4, 6) // e7

		board.Squares[e8] = NewPiece(Black, King)

		move := Move{From: e8, To: e7}
		board.applyMove(move)

		// Black castling rights should be removed
		if board.CastlingRights&CastleBlackKing != 0 {
			t.Error("CastleBlackKing should be removed after black king moves")
		}
		if board.CastlingRights&CastleBlackQueen != 0 {
			t.Error("CastleBlackQueen should be removed after black king moves")
		}
		// White castling rights should remain
		if board.CastlingRights&CastleWhiteKing == 0 {
			t.Error("CastleWhiteKing should remain after black king moves")
		}
		if board.CastlingRights&CastleWhiteQueen == 0 {
			t.Error("CastleWhiteQueen should remain after black king moves")
		}
	})

	// Test that white h1 rook move removes white kingside castling
	t.Run("white h1 rook move removes white kingside castling", func(t *testing.T) {
		board := &Board{ActiveColor: White, CastlingRights: CastleAll}
		h1 := NewSquare(7, 0) // h1
		h2 := NewSquare(7, 1) // h2

		board.Squares[h1] = NewPiece(White, Rook)

		move := Move{From: h1, To: h2}
		board.applyMove(move)

		// Only CastleWhiteKing should be removed
		if board.CastlingRights&CastleWhiteKing != 0 {
			t.Error("CastleWhiteKing should be removed after h1 rook moves")
		}
		if board.CastlingRights&CastleWhiteQueen == 0 {
			t.Error("CastleWhiteQueen should remain after h1 rook moves")
		}
		if board.CastlingRights&CastleBlackKing == 0 {
			t.Error("CastleBlackKing should remain after h1 rook moves")
		}
		if board.CastlingRights&CastleBlackQueen == 0 {
			t.Error("CastleBlackQueen should remain after h1 rook moves")
		}
	})

	// Test that white a1 rook move removes white queenside castling
	t.Run("white a1 rook move removes white queenside castling", func(t *testing.T) {
		board := &Board{ActiveColor: White, CastlingRights: CastleAll}
		a1 := NewSquare(0, 0) // a1
		a2 := NewSquare(0, 1) // a2

		board.Squares[a1] = NewPiece(White, Rook)

		move := Move{From: a1, To: a2}
		board.applyMove(move)

		// Only CastleWhiteQueen should be removed
		if board.CastlingRights&CastleWhiteQueen != 0 {
			t.Error("CastleWhiteQueen should be removed after a1 rook moves")
		}
		if board.CastlingRights&CastleWhiteKing == 0 {
			t.Error("CastleWhiteKing should remain after a1 rook moves")
		}
		if board.CastlingRights&CastleBlackKing == 0 {
			t.Error("CastleBlackKing should remain after a1 rook moves")
		}
		if board.CastlingRights&CastleBlackQueen == 0 {
			t.Error("CastleBlackQueen should remain after a1 rook moves")
		}
	})

	// Test that black h8 rook move removes black kingside castling
	t.Run("black h8 rook move removes black kingside castling", func(t *testing.T) {
		board := &Board{ActiveColor: Black, CastlingRights: CastleAll}
		h8 := NewSquare(7, 7) // h8
		h7 := NewSquare(7, 6) // h7

		board.Squares[h8] = NewPiece(Black, Rook)

		move := Move{From: h8, To: h7}
		board.applyMove(move)

		// Only CastleBlackKing should be removed
		if board.CastlingRights&CastleBlackKing != 0 {
			t.Error("CastleBlackKing should be removed after h8 rook moves")
		}
		if board.CastlingRights&CastleBlackQueen == 0 {
			t.Error("CastleBlackQueen should remain after h8 rook moves")
		}
		if board.CastlingRights&CastleWhiteKing == 0 {
			t.Error("CastleWhiteKing should remain after h8 rook moves")
		}
		if board.CastlingRights&CastleWhiteQueen == 0 {
			t.Error("CastleWhiteQueen should remain after h8 rook moves")
		}
	})

	// Test that black a8 rook move removes black queenside castling
	t.Run("black a8 rook move removes black queenside castling", func(t *testing.T) {
		board := &Board{ActiveColor: Black, CastlingRights: CastleAll}
		a8 := NewSquare(0, 7) // a8
		a7 := NewSquare(0, 6) // a7

		board.Squares[a8] = NewPiece(Black, Rook)

		move := Move{From: a8, To: a7}
		board.applyMove(move)

		// Only CastleBlackQueen should be removed
		if board.CastlingRights&CastleBlackQueen != 0 {
			t.Error("CastleBlackQueen should be removed after a8 rook moves")
		}
		if board.CastlingRights&CastleBlackKing == 0 {
			t.Error("CastleBlackKing should remain after a8 rook moves")
		}
		if board.CastlingRights&CastleWhiteKing == 0 {
			t.Error("CastleWhiteKing should remain after a8 rook moves")
		}
		if board.CastlingRights&CastleWhiteQueen == 0 {
			t.Error("CastleWhiteQueen should remain after a8 rook moves")
		}
	})

	// Test capture on h1 removes white kingside castling
	t.Run("capture on h1 removes white kingside castling", func(t *testing.T) {
		board := &Board{ActiveColor: Black, CastlingRights: CastleAll}
		h1 := NewSquare(7, 0) // h1
		g2 := NewSquare(6, 1) // g2

		board.Squares[h1] = NewPiece(White, Rook)
		board.Squares[g2] = NewPiece(Black, Bishop) // Black bishop to capture the rook

		move := Move{From: g2, To: h1}
		board.applyMove(move)

		// CastleWhiteKing should be removed due to capture on h1
		if board.CastlingRights&CastleWhiteKing != 0 {
			t.Error("CastleWhiteKing should be removed after capture on h1")
		}
		if board.CastlingRights&CastleWhiteQueen == 0 {
			t.Error("CastleWhiteQueen should remain after capture on h1")
		}
	})

	// Test capture on a1 removes white queenside castling
	t.Run("capture on a1 removes white queenside castling", func(t *testing.T) {
		board := &Board{ActiveColor: Black, CastlingRights: CastleAll}
		a1 := NewSquare(0, 0) // a1
		b2 := NewSquare(1, 1) // b2

		board.Squares[a1] = NewPiece(White, Rook)
		board.Squares[b2] = NewPiece(Black, Bishop) // Black bishop to capture the rook

		move := Move{From: b2, To: a1}
		board.applyMove(move)

		// CastleWhiteQueen should be removed due to capture on a1
		if board.CastlingRights&CastleWhiteQueen != 0 {
			t.Error("CastleWhiteQueen should be removed after capture on a1")
		}
		if board.CastlingRights&CastleWhiteKing == 0 {
			t.Error("CastleWhiteKing should remain after capture on a1")
		}
	})

	// Test capture on h8 removes black kingside castling
	t.Run("capture on h8 removes black kingside castling", func(t *testing.T) {
		board := &Board{ActiveColor: White, CastlingRights: CastleAll}
		h8 := NewSquare(7, 7) // h8
		g7 := NewSquare(6, 6) // g7

		board.Squares[h8] = NewPiece(Black, Rook)
		board.Squares[g7] = NewPiece(White, Bishop) // White bishop to capture the rook

		move := Move{From: g7, To: h8}
		board.applyMove(move)

		// CastleBlackKing should be removed due to capture on h8
		if board.CastlingRights&CastleBlackKing != 0 {
			t.Error("CastleBlackKing should be removed after capture on h8")
		}
		if board.CastlingRights&CastleBlackQueen == 0 {
			t.Error("CastleBlackQueen should remain after capture on h8")
		}
	})

	// Test capture on a8 removes black queenside castling
	t.Run("capture on a8 removes black queenside castling", func(t *testing.T) {
		board := &Board{ActiveColor: White, CastlingRights: CastleAll}
		a8 := NewSquare(0, 7) // a8
		b7 := NewSquare(1, 6) // b7

		board.Squares[a8] = NewPiece(Black, Rook)
		board.Squares[b7] = NewPiece(White, Bishop) // White bishop to capture the rook

		move := Move{From: b7, To: a8}
		board.applyMove(move)

		// CastleBlackQueen should be removed due to capture on a8
		if board.CastlingRights&CastleBlackQueen != 0 {
			t.Error("CastleBlackQueen should be removed after capture on a8")
		}
		if board.CastlingRights&CastleBlackKing == 0 {
			t.Error("CastleBlackKing should remain after capture on a8")
		}
	})

	// Test that already removed castling rights stay removed
	t.Run("already removed castling rights stay removed", func(t *testing.T) {
		// Start with only black castling rights
		board := &Board{ActiveColor: White, CastlingRights: CastleBlackKing | CastleBlackQueen}
		e1 := NewSquare(4, 0) // e1
		e2 := NewSquare(4, 1) // e2

		board.Squares[e1] = NewPiece(White, King)

		move := Move{From: e1, To: e2}
		board.applyMove(move)

		// White rights were already 0, should still be 0
		if board.CastlingRights&CastleWhiteKing != 0 {
			t.Error("CastleWhiteKing should still be 0")
		}
		if board.CastlingRights&CastleWhiteQueen != 0 {
			t.Error("CastleWhiteQueen should still be 0")
		}
		// Black rights should remain unchanged
		if board.CastlingRights&CastleBlackKing == 0 {
			t.Error("CastleBlackKing should remain")
		}
		if board.CastlingRights&CastleBlackQueen == 0 {
			t.Error("CastleBlackQueen should remain")
		}
	})

	// Test that other moves don't affect castling rights
	t.Run("other moves do not affect castling rights", func(t *testing.T) {
		board := &Board{ActiveColor: White, CastlingRights: CastleAll}
		e2 := NewSquare(4, 1) // e2
		e4 := NewSquare(4, 3) // e4

		board.Squares[e2] = NewPiece(White, Pawn)

		move := Move{From: e2, To: e4}
		board.applyMove(move)

		// All castling rights should remain
		if board.CastlingRights != CastleAll {
			t.Errorf("expected all castling rights to remain, got %d", board.CastlingRights)
		}
	})

	// Test that knight moves don't affect castling rights
	t.Run("knight moves do not affect castling rights", func(t *testing.T) {
		board := &Board{ActiveColor: White, CastlingRights: CastleAll}
		g1 := NewSquare(6, 0) // g1
		f3 := NewSquare(5, 2) // f3

		board.Squares[g1] = NewPiece(White, Knight)

		move := Move{From: g1, To: f3}
		board.applyMove(move)

		// All castling rights should remain
		if board.CastlingRights != CastleAll {
			t.Errorf("expected all castling rights to remain after knight move, got %d", board.CastlingRights)
		}
	})

	// Test that rook moving from non-original square doesn't affect castling rights
	t.Run("rook from non-original square does not affect castling rights", func(t *testing.T) {
		board := &Board{ActiveColor: White, CastlingRights: CastleAll}
		d1 := NewSquare(3, 0) // d1 (not a1 or h1)
		d4 := NewSquare(3, 3) // d4

		board.Squares[d1] = NewPiece(White, Rook)

		move := Move{From: d1, To: d4}
		board.applyMove(move)

		// All castling rights should remain
		if board.CastlingRights != CastleAll {
			t.Errorf("expected all castling rights to remain after rook moves from d1, got %d", board.CastlingRights)
		}
	})

	// Test castling itself removes both rights for that color
	t.Run("white kingside castling removes both white castling rights", func(t *testing.T) {
		board := &Board{ActiveColor: White, CastlingRights: CastleAll}
		e1 := NewSquare(4, 0) // e1
		g1 := NewSquare(6, 0) // g1
		h1 := NewSquare(7, 0) // h1

		board.Squares[e1] = NewPiece(White, King)
		board.Squares[h1] = NewPiece(White, Rook)

		// Castling move (king moves 2 squares)
		move := Move{From: e1, To: g1}
		board.applyMove(move)

		// Both white castling rights should be removed
		if board.CastlingRights&CastleWhiteKing != 0 {
			t.Error("CastleWhiteKing should be removed after castling")
		}
		if board.CastlingRights&CastleWhiteQueen != 0 {
			t.Error("CastleWhiteQueen should be removed after castling")
		}
		// Black castling rights should remain
		if board.CastlingRights&CastleBlackKing == 0 {
			t.Error("CastleBlackKing should remain after white castles")
		}
		if board.CastlingRights&CastleBlackQueen == 0 {
			t.Error("CastleBlackQueen should remain after white castles")
		}
	})

	// Test black queenside castling removes both black castling rights
	t.Run("black queenside castling removes both black castling rights", func(t *testing.T) {
		board := &Board{ActiveColor: Black, CastlingRights: CastleAll}
		e8 := NewSquare(4, 7) // e8
		c8 := NewSquare(2, 7) // c8
		a8 := NewSquare(0, 7) // a8

		board.Squares[e8] = NewPiece(Black, King)
		board.Squares[a8] = NewPiece(Black, Rook)

		// Castling move (king moves 2 squares)
		move := Move{From: e8, To: c8}
		board.applyMove(move)

		// Both black castling rights should be removed
		if board.CastlingRights&CastleBlackKing != 0 {
			t.Error("CastleBlackKing should be removed after castling")
		}
		if board.CastlingRights&CastleBlackQueen != 0 {
			t.Error("CastleBlackQueen should be removed after castling")
		}
		// White castling rights should remain
		if board.CastlingRights&CastleWhiteKing == 0 {
			t.Error("CastleWhiteKing should remain after black castles")
		}
		if board.CastlingRights&CastleWhiteQueen == 0 {
			t.Error("CastleWhiteQueen should remain after black castles")
		}
	})
}

func TestEnPassantCaptureExecution(t *testing.T) {
	// Test basic en passant capture by White
	t.Run("white en passant capture removes black pawn", func(t *testing.T) {
		// Setup: White pawn on e5, Black pawn moves d7->d5 (setting en passant square d6)
		board := &Board{ActiveColor: White}
		e5 := NewSquare(4, 4) // e5 (white pawn on 5th rank)
		d5 := NewSquare(3, 4) // d5 (black pawn just moved here)
		d6 := NewSquare(3, 5) // d6 (en passant target square)

		// Place kings (required for legal move generation)
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King) // e1
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, King) // e8

		// Place white pawn on e5 and black pawn on d5
		board.Squares[e5] = NewPiece(White, Pawn)
		board.Squares[d5] = NewPiece(Black, Pawn)

		// Set en passant square to d6 (as if black just played d7-d5)
		board.EnPassantSq = int8(d6)

		// Generate legal moves - should include en passant capture e5xd6
		legalMoves := board.LegalMoves()

		// Find the en passant capture move
		var epMove Move
		foundEP := false
		for _, m := range legalMoves {
			if m.From == e5 && m.To == d6 {
				epMove = m
				foundEP = true
				break
			}
		}

		if !foundEP {
			t.Error("en passant capture e5xd6 should be in legal moves")
		}

		// Apply the en passant capture
		err := board.MakeMove(epMove)
		if err != nil {
			t.Fatalf("MakeMove failed: %v", err)
		}

		// Verify the white pawn is now on d6
		if board.Squares[d6].Type() != Pawn || board.Squares[d6].Color() != White {
			t.Errorf("expected white pawn on d6, got %v", board.Squares[d6])
		}

		// Verify e5 is empty (pawn moved from there)
		if !board.Squares[e5].IsEmpty() {
			t.Errorf("expected e5 to be empty, got %v", board.Squares[e5])
		}

		// Verify d5 is empty (black pawn was captured)
		if !board.Squares[d5].IsEmpty() {
			t.Errorf("expected d5 to be empty (black pawn captured), got %v", board.Squares[d5])
		}

		// Verify en passant square is cleared
		if board.EnPassantSq != -1 {
			t.Errorf("expected EnPassantSq to be -1 after move, got %d", board.EnPassantSq)
		}
	})

	// Test basic en passant capture by Black
	t.Run("black en passant capture removes white pawn", func(t *testing.T) {
		// Setup: Black pawn on d4, White pawn moves e2->e4 (setting en passant square e3)
		board := &Board{ActiveColor: Black}
		d4 := NewSquare(3, 3) // d4 (black pawn on 4th rank)
		e4 := NewSquare(4, 3) // e4 (white pawn just moved here)
		e3 := NewSquare(4, 2) // e3 (en passant target square)

		// Place kings (required for legal move generation)
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King) // e1
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, King) // e8

		// Place black pawn on d4 and white pawn on e4
		board.Squares[d4] = NewPiece(Black, Pawn)
		board.Squares[e4] = NewPiece(White, Pawn)

		// Set en passant square to e3 (as if white just played e2-e4)
		board.EnPassantSq = int8(e3)

		// Generate legal moves - should include en passant capture d4xe3
		legalMoves := board.LegalMoves()

		// Find the en passant capture move
		var epMove Move
		foundEP := false
		for _, m := range legalMoves {
			if m.From == d4 && m.To == e3 {
				epMove = m
				foundEP = true
				break
			}
		}

		if !foundEP {
			t.Error("en passant capture d4xe3 should be in legal moves")
		}

		// Apply the en passant capture
		err := board.MakeMove(epMove)
		if err != nil {
			t.Fatalf("MakeMove failed: %v", err)
		}

		// Verify the black pawn is now on e3
		if board.Squares[e3].Type() != Pawn || board.Squares[e3].Color() != Black {
			t.Errorf("expected black pawn on e3, got %v", board.Squares[e3])
		}

		// Verify d4 is empty (pawn moved from there)
		if !board.Squares[d4].IsEmpty() {
			t.Errorf("expected d4 to be empty, got %v", board.Squares[d4])
		}

		// Verify e4 is empty (white pawn was captured)
		if !board.Squares[e4].IsEmpty() {
			t.Errorf("expected e4 to be empty (white pawn captured), got %v", board.Squares[e4])
		}

		// Verify en passant square is cleared
		if board.EnPassantSq != -1 {
			t.Errorf("expected EnPassantSq to be -1 after move, got %d", board.EnPassantSq)
		}
	})

	// Test en passant removes the correct pawn (verifies all three squares)
	t.Run("en passant removes correct pawn from correct squares", func(t *testing.T) {
		// White pawn on f5 captures en passant on g6, removing black pawn on g5
		board := &Board{ActiveColor: White}
		f5 := NewSquare(5, 4) // f5
		g5 := NewSquare(6, 4) // g5 (black pawn location)
		g6 := NewSquare(6, 5) // g6 (en passant target)

		board.Squares[f5] = NewPiece(White, Pawn)
		board.Squares[g5] = NewPiece(Black, Pawn)
		board.EnPassantSq = int8(g6)

		// Apply en passant capture directly
		move := Move{From: f5, To: g6}
		board.applyMove(move)

		// Verify capturing pawn moved to en passant square (g6)
		if board.Squares[g6].Type() != Pawn || board.Squares[g6].Color() != White {
			t.Errorf("expected white pawn on g6, got %v", board.Squares[g6])
		}

		// Verify original square of capturing pawn is empty (f5)
		if !board.Squares[f5].IsEmpty() {
			t.Errorf("expected f5 to be empty, got %v", board.Squares[f5])
		}

		// Verify captured pawn's square is empty (g5, NOT g6)
		if !board.Squares[g5].IsEmpty() {
			t.Errorf("expected g5 to be empty (captured pawn removed), got %v", board.Squares[g5])
		}
	})

	// Test en passant from both sides (left and right captures)
	t.Run("white pawn can capture en passant from both left and right", func(t *testing.T) {
		// Setup: White pawn on d5, black pawns could have moved to c5 or e5
		// Test left capture (d5xc6)
		board := &Board{ActiveColor: White}
		d5 := NewSquare(3, 4) // d5
		c5 := NewSquare(2, 4) // c5
		c6 := NewSquare(2, 5) // c6

		// Place kings (required for legal move generation)
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King) // e1
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, King) // e8

		board.Squares[d5] = NewPiece(White, Pawn)
		board.Squares[c5] = NewPiece(Black, Pawn)
		board.EnPassantSq = int8(c6)

		// Verify en passant move is legal
		legalMoves := board.LegalMoves()
		foundLeft := false
		for _, m := range legalMoves {
			if m.From == d5 && m.To == c6 {
				foundLeft = true
				break
			}
		}
		if !foundLeft {
			t.Error("en passant capture d5xc6 (left) should be in legal moves")
		}

		// Test right capture (d5xe6)
		board = &Board{ActiveColor: White}
		e5 := NewSquare(4, 4) // e5
		e6 := NewSquare(4, 5) // e6

		// Place kings (required for legal move generation)
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King) // e1
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, King) // e8

		board.Squares[d5] = NewPiece(White, Pawn)
		board.Squares[e5] = NewPiece(Black, Pawn)
		board.EnPassantSq = int8(e6)

		legalMoves = board.LegalMoves()
		foundRight := false
		for _, m := range legalMoves {
			if m.From == d5 && m.To == e6 {
				foundRight = true
				break
			}
		}
		if !foundRight {
			t.Error("en passant capture d5xe6 (right) should be in legal moves")
		}
	})

	// Test en passant only available on correct ranks
	t.Run("en passant only available on correct ranks", func(t *testing.T) {
		// White pawns can only en passant from rank 5
		// Setup white pawn on wrong rank (e4 instead of e5)
		board := &Board{ActiveColor: White}
		e4 := NewSquare(4, 3) // e4 (rank 4, not rank 5)
		d5 := NewSquare(3, 4) // d5
		d6 := NewSquare(3, 5) // d6

		board.Squares[e4] = NewPiece(White, Pawn)
		board.Squares[d5] = NewPiece(Black, Pawn)
		board.EnPassantSq = int8(d6)

		// Generate pseudo-legal moves (before legality check)
		pseudoMoves := board.PseudoLegalMoves()
		foundEP := false
		for _, m := range pseudoMoves {
			if m.From == e4 && m.To == d6 {
				foundEP = true
				break
			}
		}

		if foundEP {
			t.Error("pawn on rank 4 should not be able to capture en passant")
		}

		// Black pawns can only en passant from rank 4
		// Setup black pawn on wrong rank (d5 instead of d4)
		board = &Board{ActiveColor: Black}
		d5 = NewSquare(3, 4) // d5 (rank 5, not rank 4)
		e4 = NewSquare(4, 3) // e4
		e3 := NewSquare(4, 2) // e3

		board.Squares[d5] = NewPiece(Black, Pawn)
		board.Squares[e4] = NewPiece(White, Pawn)
		board.EnPassantSq = int8(e3)

		pseudoMoves = board.PseudoLegalMoves()
		foundEP = false
		for _, m := range pseudoMoves {
			if m.From == d5 && m.To == e3 {
				foundEP = true
				break
			}
		}

		if foundEP {
			t.Error("black pawn on rank 5 should not be able to capture en passant")
		}
	})

	// Test en passant when the capturing pawn is pinned (edge case)
	t.Run("en passant not legal when pawn is pinned to king", func(t *testing.T) {
		// Setup: White king on e5, white pawn on d5, black rook on a5
		// Black pawn on c5 (en passant available on c6)
		// If white pawn captures en passant, it exposes the king to check from rook
		board := &Board{ActiveColor: White}

		e5 := NewSquare(4, 4) // e5 (white king)
		d5 := NewSquare(3, 4) // d5 (white pawn - pinned)
		c5 := NewSquare(2, 4) // c5 (black pawn)
		c6 := NewSquare(2, 5) // c6 (en passant target)
		a5 := NewSquare(0, 4) // a5 (black rook - pinning)

		board.Squares[e5] = NewPiece(White, King)
		board.Squares[d5] = NewPiece(White, Pawn)
		board.Squares[c5] = NewPiece(Black, Pawn)
		board.Squares[a5] = NewPiece(Black, Rook)
		board.EnPassantSq = int8(c6)

		// Generate legal moves
		legalMoves := board.LegalMoves()

		// The en passant capture should NOT be in legal moves (would expose king)
		foundEP := false
		for _, m := range legalMoves {
			if m.From == d5 && m.To == c6 {
				foundEP = true
				break
			}
		}

		if foundEP {
			t.Error("en passant capture should not be legal when pawn is pinned to king")
		}
	})

	// Test en passant across all files
	t.Run("en passant works on all files", func(t *testing.T) {
		// Test white capturing on each file
		for file := 0; file < 7; file++ { // 0-6 (a-g files can capture to the right)
			board := &Board{ActiveColor: White}

			// Place kings (required for legal move generation)
			board.Squares[NewSquare(4, 0)] = NewPiece(White, King) // e1
			board.Squares[NewSquare(4, 7)] = NewPiece(Black, King) // e8

			// White pawn on current file, rank 5
			pawnSq := NewSquare(file, 4)
			// Black pawn on next file, rank 5
			enemySq := NewSquare(file+1, 4)
			// En passant target on next file, rank 6
			epSq := NewSquare(file+1, 5)

			board.Squares[pawnSq] = NewPiece(White, Pawn)
			board.Squares[enemySq] = NewPiece(Black, Pawn)
			board.EnPassantSq = int8(epSq)

			legalMoves := board.LegalMoves()
			foundEP := false
			for _, m := range legalMoves {
				if m.From == pawnSq && m.To == epSq {
					foundEP = true
					break
				}
			}

			if !foundEP {
				t.Errorf("en passant should work on file %d (capturing to file %d)", file, file+1)
			}
		}

		// Test capturing to the left as well (files b-h can capture to the left)
		for file := 1; file < 8; file++ {
			board := &Board{ActiveColor: White}

			// Place kings (required for legal move generation)
			board.Squares[NewSquare(4, 0)] = NewPiece(White, King) // e1
			board.Squares[NewSquare(4, 7)] = NewPiece(Black, King) // e8

			pawnSq := NewSquare(file, 4)
			enemySq := NewSquare(file-1, 4)
			epSq := NewSquare(file-1, 5)

			board.Squares[pawnSq] = NewPiece(White, Pawn)
			board.Squares[enemySq] = NewPiece(Black, Pawn)
			board.EnPassantSq = int8(epSq)

			legalMoves := board.LegalMoves()
			foundEP := false
			for _, m := range legalMoves {
				if m.From == pawnSq && m.To == epSq {
					foundEP = true
					break
				}
			}

			if !foundEP {
				t.Errorf("en passant should work on file %d (capturing to file %d)", file, file-1)
			}
		}
	})

	// Test no en passant when square is not set
	t.Run("no en passant when en passant square is not set", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		e5 := NewSquare(4, 4)
		d5 := NewSquare(3, 4)
		d6 := NewSquare(3, 5)

		board.Squares[e5] = NewPiece(White, Pawn)
		board.Squares[d5] = NewPiece(Black, Pawn)
		board.EnPassantSq = -1 // No en passant available

		legalMoves := board.LegalMoves()
		foundEP := false
		for _, m := range legalMoves {
			if m.From == e5 && m.To == d6 {
				foundEP = true
				break
			}
		}

		if foundEP {
			t.Error("en passant should not be available when EnPassantSq is -1")
		}
	})

	// Test en passant in a realistic game sequence
	t.Run("realistic en passant sequence", func(t *testing.T) {
		board := NewBoard()

		// 1. White plays e2-e4
		e2 := NewSquare(4, 1)
		e4 := NewSquare(4, 3)
		move1 := Move{From: e2, To: e4}
		err := board.MakeMove(move1)
		if err != nil {
			t.Fatalf("Move 1 failed: %v", err)
		}

		// Verify en passant square is set
		e3 := NewSquare(4, 2)
		if board.EnPassantSq != int8(e3) {
			t.Errorf("expected en passant square e3, got %d", board.EnPassantSq)
		}

		// 2. Black plays a7-a6
		a7 := NewSquare(0, 6)
		a6 := NewSquare(0, 5)
		move2 := Move{From: a7, To: a6}
		err = board.MakeMove(move2)
		if err != nil {
			t.Fatalf("Move 2 failed: %v", err)
		}

		// Verify en passant square is cleared
		if board.EnPassantSq != -1 {
			t.Errorf("expected en passant square to be cleared, got %d", board.EnPassantSq)
		}

		// 3. White plays e4-e5 (advancing the pawn to rank 5)
		e5 := NewSquare(4, 4)
		move3 := Move{From: e4, To: e5}
		err = board.MakeMove(move3)
		if err != nil {
			t.Fatalf("Move 3 failed: %v", err)
		}

		// 4. Black plays d7-d5 (now adjacent to white e5 pawn, setting up en passant)
		d7 := NewSquare(3, 6)
		d5 := NewSquare(3, 4)
		move4 := Move{From: d7, To: d5}
		err = board.MakeMove(move4)
		if err != nil {
			t.Fatalf("Move 4 failed: %v", err)
		}

		// Verify en passant square is set to d6
		d6 := NewSquare(3, 5)
		if board.EnPassantSq != int8(d6) {
			t.Errorf("expected en passant square d6, got %d", board.EnPassantSq)
		}

		// 5. White captures en passant e5xd6
		// But first verify the move is legal
		legalMoves := board.LegalMoves()
		foundEP := false
		var epMove Move
		for _, m := range legalMoves {
			if m.From == e5 && m.To == d6 {
				foundEP = true
				epMove = m
				break
			}
		}

		if !foundEP {
			t.Fatal("en passant capture e5xd6 should be legal")
		}

		err = board.MakeMove(epMove)
		if err != nil {
			t.Fatalf("En passant capture failed: %v", err)
		}

		// Verify the position after en passant
		// White pawn should be on d6
		if board.Squares[d6].Type() != Pawn || board.Squares[d6].Color() != White {
			t.Errorf("expected white pawn on d6, got %v", board.Squares[d6])
		}

		// d5 should be empty (black pawn captured)
		if !board.Squares[d5].IsEmpty() {
			t.Errorf("expected d5 to be empty, got %v", board.Squares[d5])
		}

		// e5 should be empty (white pawn moved from there)
		if !board.Squares[e5].IsEmpty() {
			t.Errorf("expected e5 to be empty, got %v", board.Squares[e5])
		}
	})

	// Test en passant discovery check (capturing reveals check on opponent's king)
	t.Run("en passant can give discovered check", func(t *testing.T) {
		// Setup: White rook on a5, black king on e5, black pawn on d5
		// White pawn on c5, en passant on d6
		// When white captures d5 en passant, it discovers check on black king
		board := &Board{ActiveColor: White}

		a5 := NewSquare(0, 4) // a5 (white rook)
		c5 := NewSquare(2, 4) // c5 (white pawn)
		d5 := NewSquare(3, 4) // d5 (black pawn)
		d6 := NewSquare(3, 5) // d6 (en passant target)
		e5 := NewSquare(4, 4) // e5 (black king)

		// Place white king (required for legal move generation)
		board.Squares[NewSquare(0, 0)] = NewPiece(White, King) // a1

		board.Squares[a5] = NewPiece(White, Rook)
		board.Squares[c5] = NewPiece(White, Pawn)
		board.Squares[d5] = NewPiece(Black, Pawn)
		board.Squares[e5] = NewPiece(Black, King)
		board.EnPassantSq = int8(d6)

		// The en passant capture should be legal
		legalMoves := board.LegalMoves()
		foundEP := false
		var epMove Move
		for _, m := range legalMoves {
			if m.From == c5 && m.To == d6 {
				foundEP = true
				epMove = m
				break
			}
		}

		if !foundEP {
			t.Error("en passant capture c5xd6 should be legal")
		}

		// Apply the move
		err := board.MakeMove(epMove)
		if err != nil {
			t.Fatalf("En passant capture failed: %v", err)
		}

		// After the move, it's Black's turn and Black should be in check
		if !board.InCheck() {
			t.Error("black king should be in check after en passant (discovered check)")
		}
	})
}
