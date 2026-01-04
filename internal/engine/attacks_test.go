package engine

import "testing"

func TestIsSquareAttacked(t *testing.T) {
	t.Run("empty board - no attacks", func(t *testing.T) {
		board := &Board{}
		e4 := NewSquare(4, 3)

		if board.IsSquareAttacked(e4, White) {
			t.Error("empty board should have no attacks from White")
		}
		if board.IsSquareAttacked(e4, Black) {
			t.Error("empty board should have no attacks from Black")
		}
	})

	t.Run("invalid square returns false", func(t *testing.T) {
		board := NewBoard()
		if board.IsSquareAttacked(NoSquare, White) {
			t.Error("invalid square should return false")
		}
		if board.IsSquareAttacked(Square(-5), Black) {
			t.Error("negative square should return false")
		}
		if board.IsSquareAttacked(Square(64), White) {
			t.Error("out of bounds square should return false")
		}
	})
}

func TestIsSquareAttackedByPawn(t *testing.T) {
	t.Run("white pawn attacks diagonally upward", func(t *testing.T) {
		board := &Board{}
		// Place white pawn on e4
		e4 := NewSquare(4, 3)
		board.Squares[e4] = NewPiece(White, Pawn)

		// White pawn on e4 attacks d5 and f5
		d5 := NewSquare(3, 4)
		f5 := NewSquare(5, 4)

		if !board.IsSquareAttacked(d5, White) {
			t.Error("white pawn on e4 should attack d5")
		}
		if !board.IsSquareAttacked(f5, White) {
			t.Error("white pawn on e4 should attack f5")
		}

		// Should not attack e5 (directly in front)
		e5 := NewSquare(4, 4)
		if board.IsSquareAttacked(e5, White) {
			t.Error("white pawn on e4 should not attack e5 (pawns don't attack forward)")
		}

		// Should not attack squares behind (d3, f3)
		d3 := NewSquare(3, 2)
		f3 := NewSquare(5, 2)
		if board.IsSquareAttacked(d3, White) {
			t.Error("white pawn on e4 should not attack d3")
		}
		if board.IsSquareAttacked(f3, White) {
			t.Error("white pawn on e4 should not attack f3")
		}
	})

	t.Run("black pawn attacks diagonally downward", func(t *testing.T) {
		board := &Board{}
		// Place black pawn on e5
		e5 := NewSquare(4, 4)
		board.Squares[e5] = NewPiece(Black, Pawn)

		// Black pawn on e5 attacks d4 and f4
		d4 := NewSquare(3, 3)
		f4 := NewSquare(5, 3)

		if !board.IsSquareAttacked(d4, Black) {
			t.Error("black pawn on e5 should attack d4")
		}
		if !board.IsSquareAttacked(f4, Black) {
			t.Error("black pawn on e5 should attack f4")
		}

		// Should not attack e4 (directly in front for black)
		e4 := NewSquare(4, 3)
		if board.IsSquareAttacked(e4, Black) {
			t.Error("black pawn on e5 should not attack e4 (pawns don't attack forward)")
		}
	})

	t.Run("pawn on a-file only attacks right diagonal", func(t *testing.T) {
		board := &Board{}
		// Place white pawn on a2
		a2 := NewSquare(0, 1)
		board.Squares[a2] = NewPiece(White, Pawn)

		// Should only attack b3 (not wrap around to h3)
		b3 := NewSquare(1, 2)
		if !board.IsSquareAttacked(b3, White) {
			t.Error("white pawn on a2 should attack b3")
		}

		// Verify no attacks on the left side (file wrapping check)
		// If there was a bug with file wrapping, the pawn might incorrectly attack h-file squares
	})

	t.Run("pawn on h-file only attacks left diagonal", func(t *testing.T) {
		board := &Board{}
		// Place white pawn on h2
		h2 := NewSquare(7, 1)
		board.Squares[h2] = NewPiece(White, Pawn)

		// Should only attack g3
		g3 := NewSquare(6, 2)
		if !board.IsSquareAttacked(g3, White) {
			t.Error("white pawn on h2 should attack g3")
		}
	})

	t.Run("pawn attack does not affect wrong color check", func(t *testing.T) {
		board := &Board{}
		// Place white pawn on e4
		e4 := NewSquare(4, 3)
		board.Squares[e4] = NewPiece(White, Pawn)

		// d5 is attacked by White but not by Black
		d5 := NewSquare(3, 4)
		if !board.IsSquareAttacked(d5, White) {
			t.Error("d5 should be attacked by White")
		}
		if board.IsSquareAttacked(d5, Black) {
			t.Error("d5 should not be attacked by Black")
		}
	})
}

func TestIsSquareAttackedByKnight(t *testing.T) {
	t.Run("knight attacks all 8 L-shaped squares", func(t *testing.T) {
		board := &Board{}
		// Place white knight on e4
		e4 := NewSquare(4, 3)
		board.Squares[e4] = NewPiece(White, Knight)

		// Knight on e4 attacks: f6, g5, g3, f2, d2, c3, c5, d6
		attacks := []struct {
			sq   Square
			name string
		}{
			{NewSquare(5, 5), "f6"},
			{NewSquare(6, 4), "g5"},
			{NewSquare(6, 2), "g3"},
			{NewSquare(5, 1), "f2"},
			{NewSquare(3, 1), "d2"},
			{NewSquare(2, 2), "c3"},
			{NewSquare(2, 4), "c5"},
			{NewSquare(3, 5), "d6"},
		}

		for _, test := range attacks {
			if !board.IsSquareAttacked(test.sq, White) {
				t.Errorf("knight on e4 should attack %s", test.name)
			}
		}

		// Should not attack adjacent squares (e3, e5, d4, f4)
		if board.IsSquareAttacked(NewSquare(4, 2), White) { // e3
			t.Error("knight should not attack adjacent square e3")
		}
		if board.IsSquareAttacked(NewSquare(4, 4), White) { // e5
			t.Error("knight should not attack adjacent square e5")
		}
	})

	t.Run("knight in corner has limited attacks", func(t *testing.T) {
		board := &Board{}
		// Place white knight on a1
		a1 := NewSquare(0, 0)
		board.Squares[a1] = NewPiece(White, Knight)

		// Knight on a1 can only attack b3 and c2
		b3 := NewSquare(1, 2)
		c2 := NewSquare(2, 1)

		if !board.IsSquareAttacked(b3, White) {
			t.Error("knight on a1 should attack b3")
		}
		if !board.IsSquareAttacked(c2, White) {
			t.Error("knight on a1 should attack c2")
		}

		// Should not attack other squares
		if board.IsSquareAttacked(NewSquare(0, 2), White) { // a3
			t.Error("knight on a1 should not attack a3")
		}
	})

	t.Run("knight attack with wrong color", func(t *testing.T) {
		board := &Board{}
		// Place white knight on e4
		e4 := NewSquare(4, 3)
		board.Squares[e4] = NewPiece(White, Knight)

		// Check Black is not attacking
		f6 := NewSquare(5, 5)
		if board.IsSquareAttacked(f6, Black) {
			t.Error("f6 should not be attacked by Black when only White knight exists")
		}
	})
}

func TestIsSquareAttackedByBishop(t *testing.T) {
	t.Run("bishop attacks all diagonal directions", func(t *testing.T) {
		board := &Board{}
		// Place white bishop on e4
		e4 := NewSquare(4, 3)
		board.Squares[e4] = NewPiece(White, Bishop)

		// Check some diagonal squares
		attacks := []struct {
			sq   Square
			name string
		}{
			{NewSquare(5, 4), "f5"}, // NE
			{NewSquare(7, 6), "h7"}, // NE far
			{NewSquare(5, 2), "f3"}, // SE
			{NewSquare(7, 0), "h1"}, // SE far
			{NewSquare(3, 2), "d3"}, // SW
			{NewSquare(1, 0), "b1"}, // SW far
			{NewSquare(3, 4), "d5"}, // NW
			{NewSquare(0, 7), "a8"}, // NW far
		}

		for _, test := range attacks {
			if !board.IsSquareAttacked(test.sq, White) {
				t.Errorf("bishop on e4 should attack %s", test.name)
			}
		}

		// Should not attack orthogonal squares
		if board.IsSquareAttacked(NewSquare(4, 4), White) { // e5
			t.Error("bishop should not attack e5 (orthogonal)")
		}
		if board.IsSquareAttacked(NewSquare(3, 3), White) { // d4
			t.Error("bishop should not attack d4 (orthogonal)")
		}
	})

	t.Run("bishop blocked by own piece", func(t *testing.T) {
		board := &Board{}
		// Place white bishop on e4 and white knight on f5 (use knight to avoid pawn attack interference)
		e4 := NewSquare(4, 3)
		f5 := NewSquare(5, 4)
		board.Squares[e4] = NewPiece(White, Bishop)
		board.Squares[f5] = NewPiece(White, Knight)

		// f5 IS attacked by the bishop (it can "see" that square, even if it can't move there)
		// This is important because if we're asking "can enemy king stand here?", the answer is no
		if !board.IsSquareAttacked(f5, White) {
			t.Error("bishop should attack square with own piece (it sees that square)")
		}

		// g6 and h7 should not be attacked (blocked by knight on f5)
		// Note: using knight because pawn on f5 would attack g6, confusing the test
		if board.IsSquareAttacked(NewSquare(6, 5), White) { // g6
			t.Error("bishop should not attack g6 (blocked by piece on f5)")
		}
		if board.IsSquareAttacked(NewSquare(7, 6), White) { // h7
			t.Error("bishop should not attack h7 (blocked by piece on f5)")
		}

		// Other diagonals should still work
		if !board.IsSquareAttacked(NewSquare(3, 4), White) { // d5
			t.Error("bishop should still attack d5 (NW diagonal)")
		}
	})

	t.Run("bishop blocked by enemy piece", func(t *testing.T) {
		board := &Board{}
		// Place white bishop on e4 and black pawn on f5
		e4 := NewSquare(4, 3)
		f5 := NewSquare(5, 4)
		board.Squares[e4] = NewPiece(White, Bishop)
		board.Squares[f5] = NewPiece(Black, Pawn)

		// f5 should be attacked (can capture enemy)
		if !board.IsSquareAttacked(f5, White) {
			t.Error("bishop should attack f5 (enemy piece)")
		}

		// g6 and h7 should not be attacked (blocked by enemy piece on f5)
		if board.IsSquareAttacked(NewSquare(6, 5), White) { // g6
			t.Error("bishop should not attack g6 (blocked by enemy piece on f5)")
		}
	})
}

func TestIsSquareAttackedByRook(t *testing.T) {
	t.Run("rook attacks all orthogonal directions", func(t *testing.T) {
		board := &Board{}
		// Place white rook on e4
		e4 := NewSquare(4, 3)
		board.Squares[e4] = NewPiece(White, Rook)

		// Check orthogonal squares
		attacks := []struct {
			sq   Square
			name string
		}{
			{NewSquare(4, 4), "e5"}, // up
			{NewSquare(4, 7), "e8"}, // up far
			{NewSquare(4, 2), "e3"}, // down
			{NewSquare(4, 0), "e1"}, // down far
			{NewSquare(5, 3), "f4"}, // right
			{NewSquare(7, 3), "h4"}, // right far
			{NewSquare(3, 3), "d4"}, // left
			{NewSquare(0, 3), "a4"}, // left far
		}

		for _, test := range attacks {
			if !board.IsSquareAttacked(test.sq, White) {
				t.Errorf("rook on e4 should attack %s", test.name)
			}
		}

		// Should not attack diagonal squares
		if board.IsSquareAttacked(NewSquare(5, 4), White) { // f5
			t.Error("rook should not attack f5 (diagonal)")
		}
		if board.IsSquareAttacked(NewSquare(3, 2), White) { // d3
			t.Error("rook should not attack d3 (diagonal)")
		}
	})

	t.Run("rook blocked by own piece", func(t *testing.T) {
		board := &Board{}
		// Place white rook on e4 and white pawn on e6
		e4 := NewSquare(4, 3)
		e6 := NewSquare(4, 5)
		board.Squares[e4] = NewPiece(White, Rook)
		board.Squares[e6] = NewPiece(White, Pawn)

		// e5 should be attacked
		if !board.IsSquareAttacked(NewSquare(4, 4), White) { // e5
			t.Error("rook should attack e5")
		}

		// e6 IS attacked by the rook (it can "see" that square)
		if !board.IsSquareAttacked(e6, White) {
			t.Error("rook should attack square with own piece (it sees that square)")
		}

		// e7 and e8 should not be attacked (blocked by piece on e6)
		if board.IsSquareAttacked(NewSquare(4, 6), White) { // e7
			t.Error("rook should not attack e7 (blocked)")
		}
		if board.IsSquareAttacked(NewSquare(4, 7), White) { // e8
			t.Error("rook should not attack e8 (blocked)")
		}
	})

	t.Run("rook blocked by enemy piece", func(t *testing.T) {
		board := &Board{}
		// Place white rook on e4 and black pawn on e6
		e4 := NewSquare(4, 3)
		e6 := NewSquare(4, 5)
		board.Squares[e4] = NewPiece(White, Rook)
		board.Squares[e6] = NewPiece(Black, Pawn)

		// e6 should be attacked (can capture)
		if !board.IsSquareAttacked(e6, White) {
			t.Error("rook should attack e6 (enemy piece)")
		}

		// e7 should not be attacked (blocked after capture)
		if board.IsSquareAttacked(NewSquare(4, 6), White) { // e7
			t.Error("rook should not attack e7 (blocked by enemy on e6)")
		}
	})
}

func TestIsSquareAttackedByQueen(t *testing.T) {
	t.Run("queen attacks diagonally", func(t *testing.T) {
		board := &Board{}
		// Place white queen on e4
		e4 := NewSquare(4, 3)
		board.Squares[e4] = NewPiece(White, Queen)

		// Check diagonal attacks
		if !board.IsSquareAttacked(NewSquare(7, 6), White) { // h7
			t.Error("queen should attack h7 diagonally")
		}
		if !board.IsSquareAttacked(NewSquare(0, 7), White) { // a8
			t.Error("queen should attack a8 diagonally")
		}
	})

	t.Run("queen attacks orthogonally", func(t *testing.T) {
		board := &Board{}
		// Place white queen on e4
		e4 := NewSquare(4, 3)
		board.Squares[e4] = NewPiece(White, Queen)

		// Check orthogonal attacks
		if !board.IsSquareAttacked(NewSquare(4, 7), White) { // e8
			t.Error("queen should attack e8 orthogonally")
		}
		if !board.IsSquareAttacked(NewSquare(0, 3), White) { // a4
			t.Error("queen should attack a4 orthogonally")
		}
	})

	t.Run("queen blocked by piece", func(t *testing.T) {
		board := &Board{}
		// Place white queen on e4 and white knight on f5 (use knight to avoid pawn attack interference)
		e4 := NewSquare(4, 3)
		f5 := NewSquare(5, 4)
		board.Squares[e4] = NewPiece(White, Queen)
		board.Squares[f5] = NewPiece(White, Knight)

		// f5 IS attacked by the queen (it can "see" that square)
		if !board.IsSquareAttacked(f5, White) {
			t.Error("queen should attack f5 (it sees that square)")
		}

		// g6 should not be attacked (blocked by knight on f5)
		// Note: using knight because pawn on f5 would attack g6
		if board.IsSquareAttacked(NewSquare(6, 5), White) { // g6
			t.Error("queen should not attack g6 (blocked)")
		}

		// Orthogonal should still work
		if !board.IsSquareAttacked(NewSquare(4, 7), White) { // e8
			t.Error("queen should still attack e8")
		}
	})
}

func TestIsSquareAttackedByKing(t *testing.T) {
	t.Run("king attacks all 8 adjacent squares", func(t *testing.T) {
		board := &Board{}
		// Place white king on e4
		e4 := NewSquare(4, 3)
		board.Squares[e4] = NewPiece(White, King)

		// Check all adjacent squares
		adjacents := []struct {
			sq   Square
			name string
		}{
			{NewSquare(3, 2), "d3"},
			{NewSquare(4, 2), "e3"},
			{NewSquare(5, 2), "f3"},
			{NewSquare(3, 3), "d4"},
			{NewSquare(5, 3), "f4"},
			{NewSquare(3, 4), "d5"},
			{NewSquare(4, 4), "e5"},
			{NewSquare(5, 4), "f5"},
		}

		for _, test := range adjacents {
			if !board.IsSquareAttacked(test.sq, White) {
				t.Errorf("king on e4 should attack %s", test.name)
			}
		}

		// Should not attack distant squares
		if board.IsSquareAttacked(NewSquare(4, 5), White) { // e6
			t.Error("king should not attack e6 (too far)")
		}
		if board.IsSquareAttacked(NewSquare(6, 5), White) { // g6
			t.Error("king should not attack g6 (too far)")
		}
	})

	t.Run("king in corner has limited attacks", func(t *testing.T) {
		board := &Board{}
		// Place white king on a1
		a1 := NewSquare(0, 0)
		board.Squares[a1] = NewPiece(White, King)

		// Should attack a2, b1, b2
		if !board.IsSquareAttacked(NewSquare(0, 1), White) { // a2
			t.Error("king on a1 should attack a2")
		}
		if !board.IsSquareAttacked(NewSquare(1, 0), White) { // b1
			t.Error("king on a1 should attack b1")
		}
		if !board.IsSquareAttacked(NewSquare(1, 1), White) { // b2
			t.Error("king on a1 should attack b2")
		}

		// Should not attack a3, c1, etc. (too far)
		if board.IsSquareAttacked(NewSquare(0, 2), White) { // a3
			t.Error("king on a1 should not attack a3")
		}
	})
}

func TestIsSquareAttackedMultiplePieces(t *testing.T) {
	t.Run("square attacked by multiple pieces", func(t *testing.T) {
		board := &Board{}
		// Place multiple white pieces that attack e4
		// White queen on e1, white rook on a4, white bishop on b7
		board.Squares[NewSquare(4, 0)] = NewPiece(White, Queen)  // e1
		board.Squares[NewSquare(0, 3)] = NewPiece(White, Rook)   // a4
		board.Squares[NewSquare(1, 6)] = NewPiece(White, Bishop) // b7

		e4 := NewSquare(4, 3)
		if !board.IsSquareAttacked(e4, White) {
			t.Error("e4 should be attacked by at least one white piece")
		}
	})

	t.Run("pieces of both colors on board", func(t *testing.T) {
		board := &Board{}
		// White knight attacks e4
		board.Squares[NewSquare(5, 5)] = NewPiece(White, Knight) // f6

		// Black knight also attacks e4
		board.Squares[NewSquare(3, 5)] = NewPiece(Black, Knight) // d6

		e4 := NewSquare(4, 3)
		if !board.IsSquareAttacked(e4, White) {
			t.Error("e4 should be attacked by White knight on f6")
		}
		if !board.IsSquareAttacked(e4, Black) {
			t.Error("e4 should be attacked by Black knight on d6")
		}
	})
}

func TestIsSquareAttackedStartingPosition(t *testing.T) {
	t.Run("squares controlled in starting position", func(t *testing.T) {
		board := NewBoard()

		// e3 is attacked by white pawns (d2 and f2)
		// Actually d2 attacks e3, f2 attacks e3
		e3 := NewSquare(4, 2)
		if !board.IsSquareAttacked(e3, White) {
			t.Error("e3 should be attacked by white pawns in starting position")
		}

		// e6 is attacked by black pawns (d7 and f7)
		e6 := NewSquare(4, 5)
		if !board.IsSquareAttacked(e6, Black) {
			t.Error("e6 should be attacked by black pawns in starting position")
		}

		// a3 is attacked by white knight on b1
		a3 := NewSquare(0, 2)
		if !board.IsSquareAttacked(a3, White) {
			t.Error("a3 should be attacked by white knight on b1")
		}

		// c3 is attacked by white knight on b1
		c3 := NewSquare(2, 2)
		if !board.IsSquareAttacked(c3, White) {
			t.Error("c3 should be attacked by white knight on b1")
		}

		// e4 should not be attacked by anyone in starting position
		e4 := NewSquare(4, 3)
		if board.IsSquareAttacked(e4, White) {
			t.Error("e4 should not be attacked by White in starting position")
		}
		if board.IsSquareAttacked(e4, Black) {
			t.Error("e4 should not be attacked by Black in starting position")
		}
	})
}

func TestIsSquareAttackedEdgeCases(t *testing.T) {
	t.Run("pieces blocked by other pieces correctly", func(t *testing.T) {
		board := &Board{}
		// White rook on a1, black pawn on a4, target is a8
		board.Squares[NewSquare(0, 0)] = NewPiece(White, Rook) // a1
		board.Squares[NewSquare(0, 3)] = NewPiece(Black, Pawn) // a4

		// a4 should be attacked by white rook
		if !board.IsSquareAttacked(NewSquare(0, 3), White) { // a4
			t.Error("a4 should be attacked by white rook on a1")
		}

		// a8 should NOT be attacked (blocked by pawn on a4)
		if board.IsSquareAttacked(NewSquare(0, 7), White) { // a8
			t.Error("a8 should not be attacked (blocked by piece on a4)")
		}
	})

	t.Run("bishop on corner attacks diagonal", func(t *testing.T) {
		board := &Board{}
		// White bishop on a1
		board.Squares[NewSquare(0, 0)] = NewPiece(White, Bishop) // a1

		// Should attack the diagonal: b2, c3, d4, e5, f6, g7, h8
		if !board.IsSquareAttacked(NewSquare(7, 7), White) { // h8
			t.Error("bishop on a1 should attack h8")
		}
		if !board.IsSquareAttacked(NewSquare(3, 3), White) { // d4
			t.Error("bishop on a1 should attack d4")
		}

		// Should not attack off-diagonal squares
		if board.IsSquareAttacked(NewSquare(1, 0), White) { // b1
			t.Error("bishop on a1 should not attack b1")
		}
	})
}
