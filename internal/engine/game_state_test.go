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
		_, hasWinner := board.Winner()
		if hasWinner {
			t.Error("expected no winner in starting position")
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

		winner, hasWinner := board.Winner()
		if !hasWinner {
			t.Error("expected a winner")
		}
		if winner != White {
			t.Errorf("expected White to win, got %v", winner)
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

		winner, hasWinner := board.Winner()
		if !hasWinner {
			t.Error("expected a winner")
		}
		if winner != Black {
			t.Errorf("expected Black to win, got %v", winner)
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

		_, hasWinner := board.Winner()
		if hasWinner {
			t.Error("expected no winner in stalemate")
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

// ============================================================================
// REPETITION TESTS
// ============================================================================

func TestThreefoldRepetition(t *testing.T) {
	// Test threefold repetition by moving knights back and forth
	// This creates a repetition of the same position three times
	t.Run("Threefold repetition via knight moves", func(t *testing.T) {
		board := NewBoard()

		// Move sequence that returns to the same position:
		// 1. Ng1-f3 Ng8-f6
		// 2. Nf3-g1 Nf6-g8 (position repeated twice)
		// 3. Ng1-f3 Ng8-f6
		// 4. Nf3-g1 Nf6-g8 (position repeated three times - threefold!)
		moves := []string{
			"g1f3", "g8f6", // Position 1
			"f3g1", "f6g8", // Back to start (position 2)
			"g1f3", "g8f6", // Position 3
			"f3g1", "f6g8", // Back to start (position 3 = threefold)
		}

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
		if status != DrawThreefoldRepetition {
			t.Errorf("expected DrawThreefoldRepetition, got %v", status)
		}
	})

	t.Run("Two repetitions is not a draw", func(t *testing.T) {
		board := NewBoard()

		// Only repeat position twice (not threefold)
		// 1. Ng1-f3 Ng8-f6
		// 2. Nf3-g1 Nf6-g8 (position repeated twice)
		moves := []string{
			"g1f3", "g8f6",
			"f3g1", "f6g8", // Back to start (position 2 only)
		}

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
		if status != Ongoing {
			t.Errorf("expected Ongoing (only 2 repetitions), got %v", status)
		}
	})

	t.Run("Threefold with same position after castling rights lost", func(t *testing.T) {
		board := NewBoard()

		// Move the king (loses castling rights), then repeat position 3 times
		// After castling rights are lost, the repeated positions should be identical
		moves := []string{
			"e2e4", "e7e5",
			"e1e2", "e8e7", // Kings move - both lose castling rights, Position A (1st)
			"e2e1", "e7e8", // Position B (1st)
			"e1e2", "e8e7", // Position A (2nd)
			"e2e1", "e7e8", // Position B (2nd)
			"e1e2", "e8e7", // Position A (3rd = threefold!)
		}

		for i, moveStr := range moves {
			move, err := ParseMove(moveStr)
			if err != nil {
				t.Fatalf("failed to parse move %s: %v", moveStr, err)
			}
			err = board.MakeMove(move)
			if err != nil {
				t.Fatalf("failed to make move %s at move %d: %v", moveStr, i+1, err)
			}
		}

		// After the third occurrence of Position A, we should have threefold
		status := board.Status()
		if status != DrawThreefoldRepetition {
			t.Errorf("expected DrawThreefoldRepetition, got %v", status)
		}
	})

	t.Run("Different castling rights means different positions", func(t *testing.T) {
		board := NewBoard()

		// The starting position has castling rights.
		// After King moves and returns, the position LOOKS the same but is NOT
		// because castling rights are different.
		initialHash := board.Hash

		// First move pawns to make room for kings
		// Then move king out and back
		moves := []string{
			"e2e4", "e7e5", // Open the way for kings
			"e1e2", // King moves - loses White castling rights
			"e8e7", // Black king moves - loses Black castling rights
			"e2e1", // King back to e1
			"e7e8", // Black king back to e8
		}

		for i, moveStr := range moves {
			move, err := ParseMove(moveStr)
			if err != nil {
				t.Fatalf("failed to parse move %s: %v", moveStr, err)
			}
			err = board.MakeMove(move)
			if err != nil {
				t.Fatalf("failed to make move %s at move %d: %v", moveStr, i+1, err)
			}
		}

		// Position has pawns on e4/e5 and kings on e1/e8.
		// This is NOT the initial position, so we can't compare directly.
		// But we can verify the initial hash is not repeated.

		// The initial hash (starting position with all castling rights) should NOT
		// appear again, even though the pieces might look similar after moving back.
		found := false
		for _, h := range board.History {
			if h == initialHash {
				found = true
				break
			}
		}
		// The initial position is in history (it's the first entry), so we expect it to be found once
		if !found {
			t.Error("expected initial position to be in history")
		}

		// But the CURRENT position should be different (has pawns on e4/e5, no castling rights)
		if board.Hash == initialHash {
			t.Error("expected current position hash to differ from initial (pawns moved, no castling)")
		}
	})
}

func TestFivefoldRepetition(t *testing.T) {
	t.Run("Fivefold repetition via knight moves", func(t *testing.T) {
		board := NewBoard()

		// Move sequence that returns to the same position five times:
		moves := []string{
			"g1f3", "g8f6", // Position 1
			"f3g1", "f6g8", // Back to start (position 2)
			"g1f3", "g8f6", // Position 3
			"f3g1", "f6g8", // Back to start (position 3)
			"g1f3", "g8f6", // Position 4
			"f3g1", "f6g8", // Back to start (position 4)
			"g1f3", "g8f6", // Position 5
			"f3g1", "f6g8", // Back to start (position 5 = fivefold!)
		}

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
		if status != DrawFivefoldRepetition {
			t.Errorf("expected DrawFivefoldRepetition, got %v", status)
		}
	})

	t.Run("Four repetitions is threefold not fivefold", func(t *testing.T) {
		board := NewBoard()

		// Only repeat position four times
		moves := []string{
			"g1f3", "g8f6",
			"f3g1", "f6g8", // Position 2
			"g1f3", "g8f6",
			"f3g1", "f6g8", // Position 3
			"g1f3", "g8f6",
			"f3g1", "f6g8", // Position 4 (threefold but not fivefold)
		}

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
		if status != DrawThreefoldRepetition {
			t.Errorf("expected DrawThreefoldRepetition (4 repetitions), got %v", status)
		}
	})
}

func TestRepetitionCount(t *testing.T) {
	t.Run("Initial position has count 1", func(t *testing.T) {
		board := NewBoard()
		count := board.repetitionCount()
		if count != 1 {
			t.Errorf("expected repetition count 1 for initial position, got %d", count)
		}
	})

	t.Run("After one move, count is 1", func(t *testing.T) {
		board := NewBoard()
		move, _ := ParseMove("e2e4")
		_ = board.MakeMove(move)

		count := board.repetitionCount()
		if count != 1 {
			t.Errorf("expected repetition count 1 after first move, got %d", count)
		}
	})

	t.Run("Count increases with repeated positions", func(t *testing.T) {
		board := NewBoard()

		moves := []string{
			"g1f3", "g8f6",
			"f3g1", "f6g8", // Back to start
		}

		for _, moveStr := range moves {
			move, err := ParseMove(moveStr)
			if err != nil {
				t.Fatalf("failed to parse move %s: %v", moveStr, err)
			}
			_ = board.MakeMove(move)
		}

		count := board.repetitionCount()
		if count != 2 {
			t.Errorf("expected repetition count 2, got %d", count)
		}
	})
}

func TestRepetitionIsGameOver(t *testing.T) {
	t.Run("Fivefold repetition is game over", func(t *testing.T) {
		board := NewBoard()

		moves := []string{
			"g1f3", "g8f6",
			"f3g1", "f6g8",
			"g1f3", "g8f6",
			"f3g1", "f6g8",
			"g1f3", "g8f6",
			"f3g1", "f6g8",
			"g1f3", "g8f6",
			"f3g1", "f6g8", // Fivefold
		}

		for _, moveStr := range moves {
			move, _ := ParseMove(moveStr)
			_ = board.MakeMove(move)
		}

		if !board.IsGameOver() {
			t.Error("fivefold repetition should be game over (automatic draw)")
		}
	})

	t.Run("Threefold repetition has no winner", func(t *testing.T) {
		board := NewBoard()

		moves := []string{
			"g1f3", "g8f6",
			"f3g1", "f6g8",
			"g1f3", "g8f6",
			"f3g1", "f6g8", // Threefold
		}

		for _, moveStr := range moves {
			move, _ := ParseMove(moveStr)
			_ = board.MakeMove(move)
		}

		_, hasWinner := board.Winner()
		if hasWinner {
			t.Error("threefold repetition should have no winner")
		}
	})
}

func TestCanClaimDraw(t *testing.T) {
	t.Run("Initial position cannot claim draw", func(t *testing.T) {
		board := NewBoard()
		if board.CanClaimDraw() {
			t.Error("initial position should not be able to claim draw")
		}
	})

	t.Run("After one move cannot claim draw", func(t *testing.T) {
		board := NewBoard()
		move, _ := ParseMove("e2e4")
		_ = board.MakeMove(move)

		if board.CanClaimDraw() {
			t.Error("should not be able to claim draw after one move")
		}
	})

	t.Run("Threefold repetition can claim draw", func(t *testing.T) {
		board := NewBoard()

		moves := []string{
			"g1f3", "g8f6",
			"f3g1", "f6g8",
			"g1f3", "g8f6",
			"f3g1", "f6g8", // Threefold
		}

		for _, moveStr := range moves {
			move, _ := ParseMove(moveStr)
			_ = board.MakeMove(move)
		}

		if !board.CanClaimDraw() {
			t.Error("should be able to claim draw after threefold repetition")
		}
	})

	t.Run("Threefold repetition is NOT game over", func(t *testing.T) {
		board := NewBoard()

		moves := []string{
			"g1f3", "g8f6",
			"f3g1", "f6g8",
			"g1f3", "g8f6",
			"f3g1", "f6g8", // Threefold
		}

		for _, moveStr := range moves {
			move, _ := ParseMove(moveStr)
			_ = board.MakeMove(move)
		}

		if board.IsGameOver() {
			t.Error("threefold repetition should NOT automatically end the game")
		}

		if !board.CanClaimDraw() {
			t.Error("should be able to claim draw")
		}
	})

	t.Run("Two repetitions cannot claim draw", func(t *testing.T) {
		board := NewBoard()

		moves := []string{
			"g1f3", "g8f6",
			"f3g1", "f6g8", // Only 2 repetitions
		}

		for _, moveStr := range moves {
			move, _ := ParseMove(moveStr)
			_ = board.MakeMove(move)
		}

		if board.CanClaimDraw() {
			t.Error("should not be able to claim draw with only 2 repetitions")
		}
	})

	t.Run("Four repetitions can still claim draw (not yet fivefold)", func(t *testing.T) {
		board := NewBoard()

		moves := []string{
			"g1f3", "g8f6",
			"f3g1", "f6g8",
			"g1f3", "g8f6",
			"f3g1", "f6g8",
			"g1f3", "g8f6",
			"f3g1", "f6g8", // 4 repetitions
		}

		for _, moveStr := range moves {
			move, _ := ParseMove(moveStr)
			_ = board.MakeMove(move)
		}

		if board.IsGameOver() {
			t.Error("4 repetitions should not be game over (not yet fivefold)")
		}

		if !board.CanClaimDraw() {
			t.Error("should be able to claim draw with 4 repetitions")
		}
	})

	t.Run("Fivefold repetition is game over (not just claimable)", func(t *testing.T) {
		board := NewBoard()

		moves := []string{
			"g1f3", "g8f6",
			"f3g1", "f6g8",
			"g1f3", "g8f6",
			"f3g1", "f6g8",
			"g1f3", "g8f6",
			"f3g1", "f6g8",
			"g1f3", "g8f6",
			"f3g1", "f6g8", // 5 repetitions
		}

		for _, moveStr := range moves {
			move, _ := ParseMove(moveStr)
			_ = board.MakeMove(move)
		}

		if !board.IsGameOver() {
			t.Error("fivefold repetition should automatically end the game")
		}

		if board.Status() != DrawFivefoldRepetition {
			t.Errorf("expected DrawFivefoldRepetition, got %v", board.Status())
		}
	})

	t.Run("Checkmate is game over but not claimable draw", func(t *testing.T) {
		board := NewBoard()
		// Fool's mate
		moves := []string{"f2f3", "e7e5", "g2g4", "d8h4"}

		for _, moveStr := range moves {
			move, _ := ParseMove(moveStr)
			_ = board.MakeMove(move)
		}

		if !board.IsGameOver() {
			t.Error("checkmate should be game over")
		}

		if board.CanClaimDraw() {
			t.Error("checkmate should not be claimable as a draw")
		}
	})

	t.Run("Stalemate is game over but not claimable draw", func(t *testing.T) {
		board := NewBoard()
		// Clear the board and set up a stalemate position
		for i := range board.Squares {
			board.Squares[i] = Piece(Empty)
		}

		// King on a8, opponent King on c6, opponent Queen on b6
		// Black to move, in stalemate
		a8 := NewSquare(0, 7)
		c6 := NewSquare(2, 5)
		b6 := NewSquare(1, 5)

		board.Squares[a8] = NewPiece(Black, King)
		board.Squares[c6] = NewPiece(White, King)
		board.Squares[b6] = NewPiece(White, Queen)
		board.ActiveColor = Black
		board.CastlingRights = 0
		board.EnPassantSq = -1
		board.HalfMoveClock = 0

		if !board.IsGameOver() {
			t.Error("stalemate should be game over")
		}

		if board.CanClaimDraw() {
			t.Error("stalemate should not be claimable as a draw")
		}

		if board.Status() != Stalemate {
			t.Errorf("expected Stalemate, got %v", board.Status())
		}
	})
}
