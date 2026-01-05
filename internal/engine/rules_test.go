package engine

import "testing"

// Helper function to set up a board from a simple position description.
// This clears the board and places pieces according to the given specifications.
func setupPosition(b *Board, pieces map[Square]Piece, activeColor Color) {
	// Clear the board
	for i := range b.Squares {
		b.Squares[i] = Piece(Empty)
	}

	// Place the pieces
	for sq, piece := range pieces {
		b.Squares[sq] = piece
	}

	// Set active color
	b.ActiveColor = activeColor

	// Reset other board state
	b.CastlingRights = 0
	b.EnPassantSq = -1
	b.HalfMoveClock = 0
	b.FullMoveNum = 1
}

// squareFromNotation converts algebraic notation (e.g., "e4") to a Square.
func squareFromNotation(notation string) Square {
	if len(notation) != 2 {
		return NoSquare
	}
	file := int(notation[0] - 'a')
	rank := int(notation[1] - '1')
	return NewSquare(file, rank)
}

func TestGameStatusString(t *testing.T) {
	tests := []struct {
		status   GameStatus
		expected string
	}{
		{Ongoing, "ongoing"},
		{Checkmate, "checkmate"},
		{Stalemate, "stalemate"},
		{DrawInsufficientMaterial, "draw (insufficient material)"},
		{DrawFiftyMoveRule, "draw (fifty-move rule)"},
		{DrawSeventyFiveMoveRule, "draw (seventy-five-move rule)"},
		{DrawThreefoldRepetition, "draw (threefold repetition)"},
		{DrawFivefoldRepetition, "draw (fivefold repetition)"},
		{GameStatus(100), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("GameStatus.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestStatusOngoing(t *testing.T) {
	t.Run("Starting position is ongoing", func(t *testing.T) {
		board := NewBoard()
		status := board.Status()
		if status != Ongoing {
			t.Errorf("expected Ongoing, got %v", status)
		}
	})

	t.Run("After e4 is ongoing", func(t *testing.T) {
		board := NewBoard()
		move, _ := ParseMove("e2e4")
		_ = board.MakeMove(move)
		status := board.Status()
		if status != Ongoing {
			t.Errorf("expected Ongoing, got %v", status)
		}
	})

	t.Run("King vs King with other pieces is ongoing", func(t *testing.T) {
		board := NewBoard()
		// White King on e1, White Queen on d1, Black King on e8
		pieces := map[Square]Piece{
			squareFromNotation("e1"): NewPiece(White, King),
			squareFromNotation("d1"): NewPiece(White, Queen),
			squareFromNotation("e8"): NewPiece(Black, King),
		}
		setupPosition(board, pieces, White)
		status := board.Status()
		if status != Ongoing {
			t.Errorf("expected Ongoing, got %v", status)
		}
	})
}

// ============================================================================
// CHECKMATE TESTS
// ============================================================================

func TestCheckmatePositions(t *testing.T) {
	// Fool's Mate - The fastest checkmate in chess (2 moves)
	// 1. f3 e5 2. g4 Qh4#
	t.Run("Fool's Mate", func(t *testing.T) {
		board := NewBoard()

		// Play the moves: 1. f3 e5 2. g4 Qh4#
		moves := []string{"f2f3", "e7e5", "g2g4", "d8h4"}
		for _, moveStr := range moves {
			move, err := ParseMove(moveStr)
			if err != nil {
				t.Fatalf("failed to parse move %s: %v", moveStr, err)
			}
			err = board.MakeMove(move)
			if err != nil {
				t.Fatalf("failed to make move %s: %v", moveStr, err)
			}
		}

		status := board.Status()
		if status != Checkmate {
			t.Errorf("Fool's Mate: expected Checkmate, got %v", status)
		}

		// Verify white is checkmated
		if board.ActiveColor != White {
			t.Errorf("expected White to be checkmated, active color is %v", board.ActiveColor)
		}
	})

	// Scholar's Mate - A common 4-move checkmate
	// 1. e4 e5 2. Bc4 Nc6 3. Qh5 Nf6?? 4. Qxf7#
	t.Run("Scholar's Mate", func(t *testing.T) {
		board := NewBoard()

		moves := []string{"e2e4", "e7e5", "f1c4", "b8c6", "d1h5", "g8f6", "h5f7"}
		for _, moveStr := range moves {
			move, err := ParseMove(moveStr)
			if err != nil {
				t.Fatalf("failed to parse move %s: %v", moveStr, err)
			}
			err = board.MakeMove(move)
			if err != nil {
				t.Fatalf("failed to make move %s: %v", moveStr, err)
			}
		}

		status := board.Status()
		if status != Checkmate {
			t.Errorf("Scholar's Mate: expected Checkmate, got %v", status)
		}

		// Verify black is checkmated
		if board.ActiveColor != Black {
			t.Errorf("expected Black to be checkmated, active color is %v", board.ActiveColor)
		}
	})

	// Back Rank Mate - Classic checkmate pattern
	// White Rook on e8 delivers checkmate to Black King on g8
	t.Run("Back Rank Mate", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("e1"): NewPiece(White, King),
			squareFromNotation("e8"): NewPiece(White, Rook),
			squareFromNotation("g8"): NewPiece(Black, King),
			squareFromNotation("f7"): NewPiece(Black, Pawn),
			squareFromNotation("g7"): NewPiece(Black, Pawn),
			squareFromNotation("h7"): NewPiece(Black, Pawn),
		}
		setupPosition(board, pieces, Black)

		status := board.Status()
		if status != Checkmate {
			t.Errorf("Back Rank Mate: expected Checkmate, got %v", status)
		}
	})


	// Two Rooks Mate (Ladder Mate)
	t.Run("Two Rooks Mate", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("e1"): NewPiece(White, King),
			squareFromNotation("a8"): NewPiece(White, Rook), // Delivers check
			squareFromNotation("b7"): NewPiece(White, Rook), // Cuts off escape
			squareFromNotation("h8"): NewPiece(Black, King),
		}
		setupPosition(board, pieces, Black)

		status := board.Status()
		if status != Checkmate {
			t.Errorf("Two Rooks Mate: expected Checkmate, got %v", status)
		}
	})

	// Queen and King Mate - Basic endgame checkmate
	t.Run("Queen and King Mate", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("f6"): NewPiece(White, King),
			squareFromNotation("g7"): NewPiece(White, Queen),
			squareFromNotation("h8"): NewPiece(Black, King),
		}
		setupPosition(board, pieces, Black)

		status := board.Status()
		if status != Checkmate {
			t.Errorf("Queen and King Mate: expected Checkmate, got %v", status)
		}
	})

	// King and Rook Mate - Basic endgame checkmate
	// King on g6 cuts off g7, g8; Rook on a8 delivers checkmate
	t.Run("King and Rook Mate", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("g6"): NewPiece(White, King), // Controls f7, g7, h7
			squareFromNotation("a8"): NewPiece(White, Rook), // Delivers check on 8th rank
			squareFromNotation("h8"): NewPiece(Black, King), // Trapped: g8 controlled by king, h7 controlled by king
		}
		setupPosition(board, pieces, Black)

		status := board.Status()
		if status != Checkmate {
			t.Errorf("King and Rook Mate: expected Checkmate, got %v", status)
		}
	})

}

// Test that check is NOT checkmate when escape is possible
func TestCheckButNotCheckmate(t *testing.T) {
	t.Run("Check with escape square available", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("e1"): NewPiece(White, King),
			squareFromNotation("e8"): NewPiece(White, Rook), // Delivers check
			squareFromNotation("d8"): NewPiece(Black, King), // Can escape to c8, c7, d7
		}
		setupPosition(board, pieces, Black)

		if !board.InCheck() {
			t.Error("expected Black to be in check")
		}

		status := board.Status()
		if status != Ongoing {
			t.Errorf("Check with escape: expected Ongoing, got %v", status)
		}
	})

	t.Run("Check that can be blocked", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("e1"): NewPiece(White, King),
			squareFromNotation("e8"): NewPiece(White, Rook), // Delivers check
			squareFromNotation("h8"): NewPiece(Black, King),
			squareFromNotation("a1"): NewPiece(Black, Rook), // Can block on e1 (wait, that's capture)
			squareFromNotation("f1"): NewPiece(Black, Rook), // Can block on f8
		}
		setupPosition(board, pieces, Black)

		if !board.InCheck() {
			t.Error("expected Black to be in check")
		}

		status := board.Status()
		if status != Ongoing {
			t.Errorf("Check that can be blocked: expected Ongoing, got %v", status)
		}
	})

	t.Run("Check where attacker can be captured", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("e1"): NewPiece(White, King),
			squareFromNotation("e8"): NewPiece(White, Rook), // Delivers check
			squareFromNotation("h8"): NewPiece(Black, King),
			squareFromNotation("a8"): NewPiece(Black, Rook), // Can capture attacking rook
		}
		setupPosition(board, pieces, Black)

		if !board.InCheck() {
			t.Error("expected Black to be in check")
		}

		status := board.Status()
		if status != Ongoing {
			t.Errorf("Check where attacker capturable: expected Ongoing, got %v", status)
		}
	})
}

// ============================================================================
// STALEMATE TESTS
// ============================================================================

func TestStalematePositions(t *testing.T) {
	// Classic stalemate: King in corner with no legal moves
	t.Run("King cornered stalemate", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("f6"): NewPiece(White, King),
			squareFromNotation("g6"): NewPiece(White, Queen), // Controls h7 and h8
			squareFromNotation("h8"): NewPiece(Black, King),  // No legal moves
		}
		setupPosition(board, pieces, Black)

		// Verify Black is NOT in check
		if board.InCheck() {
			t.Error("stalemate position should not be in check")
		}

		status := board.Status()
		if status != Stalemate {
			t.Errorf("King cornered: expected Stalemate, got %v", status)
		}
	})


	// Famous stalemate position: Philidor's position variant
	t.Run("Queen vs King stalemate", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("c6"): NewPiece(White, King),
			squareFromNotation("b6"): NewPiece(White, Queen), // Wrong placement leads to stalemate
			squareFromNotation("a8"): NewPiece(Black, King),  // No legal moves
		}
		setupPosition(board, pieces, Black)

		// Verify Black is NOT in check
		if board.InCheck() {
			t.Error("stalemate position should not be in check")
		}

		status := board.Status()
		if status != Stalemate {
			t.Errorf("Queen vs King stalemate: expected Stalemate, got %v", status)
		}
	})


	// Edge case: King trapped by Rooks
	t.Run("King trapped by rooks stalemate", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("c1"): NewPiece(White, King),
			squareFromNotation("b2"): NewPiece(White, Rook), // Controls a2, c2-h2
			squareFromNotation("b3"): NewPiece(White, Rook), // Controls a3, c3-h3
			squareFromNotation("a1"): NewPiece(Black, King), // Only a2 would be escape, but blocked by rook
		}
		setupPosition(board, pieces, Black)

		// Verify Black is NOT in check
		if board.InCheck() {
			t.Error("stalemate position should not be in check")
		}

		status := board.Status()
		if status != Stalemate {
			t.Errorf("King trapped by rooks: expected Stalemate, got %v", status)
		}
	})

}

// Test that positions with legal moves are not stalemate
func TestNotStalemate(t *testing.T) {
	t.Run("King has escape square", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("a1"): NewPiece(White, King),
			squareFromNotation("c3"): NewPiece(White, Queen),
			squareFromNotation("h8"): NewPiece(Black, King), // Can move to g8, g7, h7
		}
		setupPosition(board, pieces, Black)

		status := board.Status()
		if status != Ongoing {
			t.Errorf("King with escape: expected Ongoing, got %v", status)
		}
	})

	t.Run("Piece can move even if king blocked", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("f6"): NewPiece(White, King),
			squareFromNotation("g6"): NewPiece(White, Queen),
			squareFromNotation("h8"): NewPiece(Black, King), // King is blocked
			squareFromNotation("a1"): NewPiece(Black, Rook), // But rook can move
		}
		setupPosition(board, pieces, Black)

		status := board.Status()
		if status != Ongoing {
			t.Errorf("Piece can move: expected Ongoing, got %v", status)
		}
	})
}

// ============================================================================
// HELPER METHOD TESTS
// ============================================================================

func TestIsGameOver(t *testing.T) {
	t.Run("Starting position is not game over", func(t *testing.T) {
		board := NewBoard()
		if board.IsGameOver() {
			t.Error("starting position should not be game over")
		}
	})

	t.Run("Checkmate is game over", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("f6"): NewPiece(White, King),
			squareFromNotation("g7"): NewPiece(White, Queen),
			squareFromNotation("h8"): NewPiece(Black, King),
		}
		setupPosition(board, pieces, Black)

		if !board.IsGameOver() {
			t.Error("checkmate should be game over")
		}
	})

	t.Run("Stalemate is game over", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("f6"): NewPiece(White, King),
			squareFromNotation("g6"): NewPiece(White, Queen),
			squareFromNotation("h8"): NewPiece(Black, King),
		}
		setupPosition(board, pieces, Black)

		if !board.IsGameOver() {
			t.Error("stalemate should be game over")
		}
	})
}

func TestWinner(t *testing.T) {
	t.Run("No winner in starting position", func(t *testing.T) {
		board := NewBoard()
		if board.Winner() != -1 {
			t.Errorf("expected no winner (-1), got %d", board.Winner())
		}
	})

	t.Run("White wins when Black is checkmated", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("f6"): NewPiece(White, King),
			squareFromNotation("g7"): NewPiece(White, Queen),
			squareFromNotation("h8"): NewPiece(Black, King),
		}
		setupPosition(board, pieces, Black)

		if board.Winner() != int(White) {
			t.Errorf("expected White (%d) to win, got %d", White, board.Winner())
		}
	})

	t.Run("Black wins when White is checkmated", func(t *testing.T) {
		// Simple back rank mate against white
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("g1"): NewPiece(White, King),
			squareFromNotation("f2"): NewPiece(White, Pawn),
			squareFromNotation("g2"): NewPiece(White, Pawn),
			squareFromNotation("h2"): NewPiece(White, Pawn),
			squareFromNotation("e1"): NewPiece(Black, Rook), // Delivering checkmate on back rank
			squareFromNotation("e8"): NewPiece(Black, King),
		}
		setupPosition(board, pieces, White)

		// Verify it's checkmate
		if board.Status() != Checkmate {
			t.Errorf("expected Checkmate, got %v", board.Status())
		}

		if board.Winner() != int(Black) {
			t.Errorf("expected Black (%d) to win, got %d", Black, board.Winner())
		}
	})

	t.Run("No winner in stalemate", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("f6"): NewPiece(White, King),
			squareFromNotation("g6"): NewPiece(White, Queen),
			squareFromNotation("h8"): NewPiece(Black, King),
		}
		setupPosition(board, pieces, Black)

		if board.Winner() != -1 {
			t.Errorf("expected no winner (-1) in stalemate, got %d", board.Winner())
		}
	})
}

// ============================================================================
// EDGE CASES AND REGRESSION TESTS
// ============================================================================

func TestStatusEdgeCases(t *testing.T) {
	// Double check resulting in checkmate (can't block, must move king)
	t.Run("Double check checkmate", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("f6"): NewPiece(White, King),
			squareFromNotation("g1"): NewPiece(White, Rook),   // Check on g-file
			squareFromNotation("a2"): NewPiece(White, Bishop), // Check on a2-g8 diagonal
			squareFromNotation("g8"): NewPiece(Black, King),   // Double checked by rook on g-file and bishop on diagonal
			squareFromNotation("f8"): NewPiece(Black, Rook),   // Blocks f8 escape
			squareFromNotation("f7"): NewPiece(Black, Pawn),   // Blocks f7 escape
			squareFromNotation("h8"): NewPiece(Black, Rook),   // Blocks h8 escape
			squareFromNotation("h7"): NewPiece(Black, Pawn),   // Blocks h7 escape
		}
		setupPosition(board, pieces, Black)

		// Verify Black is in check
		if !board.InCheck() {
			t.Error("expected Black to be in double check")
		}

		status := board.Status()
		if status != Checkmate {
			t.Errorf("Double check checkmate: expected Checkmate, got %v", status)
		}
	})

	// King can capture the checking piece
	t.Run("King can capture checking piece", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("a1"): NewPiece(White, King),
			squareFromNotation("h7"): NewPiece(White, Rook), // Delivers check on 7th rank
			squareFromNotation("g7"): NewPiece(Black, King), // Can capture Rook on h7
		}
		setupPosition(board, pieces, Black)

		// Verify Black is in check
		if !board.InCheck() {
			t.Error("expected Black to be in check")
		}

		status := board.Status()
		if status != Ongoing {
			t.Errorf("King can capture: expected Ongoing, got %v", status)
		}
	})

	// Pawn can block check
	t.Run("Pawn can block check", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("a1"): NewPiece(White, King),
			squareFromNotation("a8"): NewPiece(White, Rook), // Delivers check on 8th rank
			squareFromNotation("h8"): NewPiece(Black, King),
			squareFromNotation("d7"): NewPiece(Black, Pawn), // Can block on d8
		}
		setupPosition(board, pieces, Black)

		// Verify Black is in check
		if !board.InCheck() {
			t.Error("expected Black to be in check")
		}

		status := board.Status()
		if status != Ongoing {
			t.Errorf("Pawn can block: expected Ongoing, got %v", status)
		}
	})

	// Knight can capture checking piece
	t.Run("Knight can capture checking piece", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("a1"): NewPiece(White, King),
			squareFromNotation("h7"): NewPiece(White, Rook), // Delivers check
			squareFromNotation("h8"): NewPiece(Black, King),
			squareFromNotation("g7"): NewPiece(Black, Pawn), // Blocks g7
			squareFromNotation("f6"): NewPiece(Black, Knight), // Can capture rook on h7
		}
		setupPosition(board, pieces, Black)

		// Verify Black is in check
		if !board.InCheck() {
			t.Error("expected Black to be in check")
		}

		status := board.Status()
		if status != Ongoing {
			t.Errorf("Knight can capture: expected Ongoing, got %v", status)
		}
	})
}

// Test that status is correct after a sequence of moves leading to checkmate
func TestCheckmateAfterMoveSequence(t *testing.T) {
	// Test that we can detect checkmate after manually setting up a position
	t.Run("Checkmate delivered by queen", func(t *testing.T) {
		board := NewBoard()
		pieces := map[Square]Piece{
			squareFromNotation("e1"): NewPiece(White, King),
			squareFromNotation("d5"): NewPiece(White, Queen),
			squareFromNotation("e8"): NewPiece(Black, King),
		}
		setupPosition(board, pieces, White)

		// Before moving, verify ongoing
		status := board.Status()
		if status != Ongoing {
			t.Errorf("Before move: expected Ongoing, got %v", status)
		}

		// Create a checkmate position: Queen on e7 delivers checkmate with King on f6
		board.Squares[squareFromNotation("d5")] = Piece(Empty)
		board.Squares[squareFromNotation("e7")] = NewPiece(White, Queen) // Delivers check
		board.Squares[squareFromNotation("f6")] = NewPiece(White, King)  // Cut off escape
		board.Squares[squareFromNotation("e1")] = Piece(Empty)
		board.ActiveColor = Black

		status = board.Status()
		if status != Checkmate {
			t.Errorf("After setup: expected Checkmate, got %v", status)
		}
	})
}

// Test status with castling rights (shouldn't affect checkmate/stalemate)
func TestStatusWithCastlingRights(t *testing.T) {
	t.Run("Checkmate ignores castling rights", func(t *testing.T) {
		board := NewBoard()
		// Back rank mate where castling rights exist but are irrelevant
		pieces := map[Square]Piece{
			squareFromNotation("e1"): NewPiece(White, King),
			squareFromNotation("h1"): NewPiece(White, Rook), // Rook present for potential castling
			squareFromNotation("a8"): NewPiece(White, Rook), // Delivers checkmate on 8th rank
			squareFromNotation("h8"): NewPiece(Black, King),
			squareFromNotation("g7"): NewPiece(Black, Pawn), // Blocks g7
			squareFromNotation("h7"): NewPiece(Black, Pawn), // Blocks h7
		}
		setupPosition(board, pieces, Black)
		board.CastlingRights = CastleWhiteKing | CastleWhiteQueen // White has castling rights

		status := board.Status()
		if status != Checkmate {
			t.Errorf("Checkmate with castling rights: expected Checkmate, got %v", status)
		}
	})
}
