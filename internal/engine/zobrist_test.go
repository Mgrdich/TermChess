package engine

import "testing"

func TestZobristInitialization(t *testing.T) {
	// Test that Zobrist tables are initialized (non-zero values exist)
	t.Run("Piece tables initialized", func(t *testing.T) {
		nonZeroCount := 0
		for pieceIdx := 0; pieceIdx < 12; pieceIdx++ {
			for sq := 0; sq < 64; sq++ {
				if zobristPieces[pieceIdx][sq] != 0 {
					nonZeroCount++
				}
			}
		}
		// Should have many non-zero values (extremely unlikely all would be 0)
		if nonZeroCount < 100 {
			t.Errorf("expected many non-zero Zobrist piece values, got only %d", nonZeroCount)
		}
	})

	t.Run("Side to move initialized", func(t *testing.T) {
		if zobristSideToMove == 0 {
			t.Error("expected zobristSideToMove to be non-zero")
		}
	})

	t.Run("Castling tables initialized", func(t *testing.T) {
		nonZeroCount := 0
		for i := 0; i < 16; i++ {
			if zobristCastling[i] != 0 {
				nonZeroCount++
			}
		}
		if nonZeroCount < 8 {
			t.Errorf("expected many non-zero castling values, got only %d", nonZeroCount)
		}
	})

	t.Run("En passant tables initialized", func(t *testing.T) {
		nonZeroCount := 0
		for i := 0; i < 8; i++ {
			if zobristEnPassant[i] != 0 {
				nonZeroCount++
			}
		}
		if nonZeroCount < 4 {
			t.Errorf("expected many non-zero en passant values, got only %d", nonZeroCount)
		}
	})
}

func TestNewBoardHasNonZeroHash(t *testing.T) {
	board := NewBoard()
	if board.Hash == 0 {
		t.Error("expected NewBoard() to have non-zero hash")
	}
}

func TestNewBoardHashIsDeterministic(t *testing.T) {
	board1 := NewBoard()
	board2 := NewBoard()

	if board1.Hash != board2.Hash {
		t.Errorf("expected same hash for identical boards, got %x and %x", board1.Hash, board2.Hash)
	}
}

func TestHashChangesAfterMove(t *testing.T) {
	board := NewBoard()
	initialHash := board.Hash

	// Make a simple pawn move
	move, _ := ParseMove("e2e4")
	err := board.MakeMove(move)
	if err != nil {
		t.Fatalf("failed to make move: %v", err)
	}

	if board.Hash == initialHash {
		t.Error("expected hash to change after move")
	}
}

func TestDifferentMovesProduceDifferentHashes(t *testing.T) {
	// Play e2e4
	board1 := NewBoard()
	move1, _ := ParseMove("e2e4")
	board1.MakeMove(move1)

	// Play d2d4
	board2 := NewBoard()
	move2, _ := ParseMove("d2d4")
	board2.MakeMove(move2)

	if board1.Hash == board2.Hash {
		t.Error("expected different moves to produce different hashes")
	}
}

func TestHistoryGrows(t *testing.T) {
	board := NewBoard()

	// Initial position should be in history
	if len(board.History) != 1 {
		t.Errorf("expected initial History length 1 (starting position), got %d", len(board.History))
	}

	// Make a move
	move1, _ := ParseMove("e2e4")
	board.MakeMove(move1)

	if len(board.History) != 2 {
		t.Errorf("expected History length 2 after one move, got %d", len(board.History))
	}

	// Make another move
	move2, _ := ParseMove("e7e5")
	board.MakeMove(move2)

	if len(board.History) != 3 {
		t.Errorf("expected History length 3 after two moves, got %d", len(board.History))
	}
}

func TestInitialPositionInHistory(t *testing.T) {
	board := NewBoard()

	// The initial position's hash should be the first entry in history
	if len(board.History) == 0 {
		t.Fatal("expected initial position hash to be in history")
	}

	if board.History[0] != board.Hash {
		t.Errorf("expected first history entry %x to match current hash %x", board.History[0], board.Hash)
	}
}

func TestThreefoldRepetitionWithStartingPosition(t *testing.T) {
	// This tests that returning to the starting position is correctly counted
	// for threefold repetition detection.
	//
	// Play knight moves that return to the starting position:
	// 1. Nf3 Nf6 2. Ng1 Ng8 (back to start - 2nd occurrence)
	// 3. Nf3 Nf6 4. Ng1 Ng8 (back to start - 3rd occurrence)
	board := NewBoard()
	startingHash := board.Hash

	// First occurrence is already in history (initial position)
	countOccurrences := func() int {
		count := 0
		for _, h := range board.History {
			if h == startingHash {
				count++
			}
		}
		return count
	}

	if countOccurrences() != 1 {
		t.Errorf("expected 1 occurrence of starting position initially, got %d", countOccurrences())
	}

	// Play moves to return to starting position
	moves := []string{
		"g1f3", "g8f6", // Knights out
		"f3g1", "f6g8", // Knights back - 2nd occurrence
	}
	for _, moveStr := range moves {
		move, _ := ParseMove(moveStr)
		board.MakeMove(move)
	}

	if countOccurrences() != 2 {
		t.Errorf("expected 2 occurrences of starting position after returning once, got %d", countOccurrences())
	}

	// Verify the current position matches starting position
	if board.Hash != startingHash {
		t.Error("expected to be back at starting position")
	}

	// Return to starting position again
	moves2 := []string{
		"g1f3", "g8f6", // Knights out again
		"f3g1", "f6g8", // Knights back - 3rd occurrence
	}
	for _, moveStr := range moves2 {
		move, _ := ParseMove(moveStr)
		board.MakeMove(move)
	}

	if countOccurrences() != 3 {
		t.Errorf("expected 3 occurrences of starting position (threefold repetition), got %d", countOccurrences())
	}
}

func TestMultipleGamesHaveIndependentHistory(t *testing.T) {
	// Test that creating multiple boards doesn't cause any initialization issues
	board1 := NewBoard()
	board2 := NewBoard()

	// Both should have independent histories
	if len(board1.History) != 1 || len(board2.History) != 1 {
		t.Errorf("expected both boards to have history length 1, got %d and %d",
			len(board1.History), len(board2.History))
	}

	// Make moves on board1 only
	move, _ := ParseMove("e2e4")
	board1.MakeMove(move)

	// board1 history should grow, board2 should be unchanged
	if len(board1.History) != 2 {
		t.Errorf("expected board1 history length 2, got %d", len(board1.History))
	}
	if len(board2.History) != 1 {
		t.Errorf("expected board2 history length 1 (unchanged), got %d", len(board2.History))
	}

	// Verify they don't share the same underlying slice
	board1.History[0] = 0xDEADBEEF
	if board2.History[0] == 0xDEADBEEF {
		t.Error("boards appear to share the same history slice - initialization bug!")
	}
}

func TestHashMatchesComputedHash(t *testing.T) {
	// This verifies that incremental hash updates match a full recomputation
	board := NewBoard()

	// Play several moves
	moves := []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1b5"}
	for _, moveStr := range moves {
		move, err := ParseMove(moveStr)
		if err != nil {
			t.Fatalf("failed to parse move %s: %v", moveStr, err)
		}
		err = board.MakeMove(move)
		if err != nil {
			t.Fatalf("failed to make move %s: %v", moveStr, err)
		}

		// Verify incremental hash matches full computation
		computedHash := board.ComputeHash()
		if board.Hash != computedHash {
			t.Errorf("after move %s: incremental hash %x != computed hash %x", moveStr, board.Hash, computedHash)
		}
	}
}

func TestSamePositionSameHash(t *testing.T) {
	// Test that the same position reached via different move orders has the same hash
	// Note: We must ensure en passant state is the same too!
	// Using: 1.Nf3 Nc6 2.Nc3 Nf6 vs 1.Nc3 Nf6 2.Nf3 Nc6
	// Both reach the same position with no en passant and all castling rights

	// Path 1: g1f3, b8c6, b1c3, g8f6
	board1 := NewBoard()
	moves1 := []string{"g1f3", "b8c6", "b1c3", "g8f6"}
	for _, moveStr := range moves1 {
		move, _ := ParseMove(moveStr)
		board1.MakeMove(move)
	}

	// Path 2: b1c3, g8f6, g1f3, b8c6
	board2 := NewBoard()
	moves2 := []string{"b1c3", "g8f6", "g1f3", "b8c6"}
	for _, moveStr := range moves2 {
		move, _ := ParseMove(moveStr)
		board2.MakeMove(move)
	}

	// Both should have the same hash (same position, same side to move, same castling, no EP)
	if board1.Hash != board2.Hash {
		t.Errorf("same position via different move orders should have same hash: %x vs %x", board1.Hash, board2.Hash)
	}
}

func TestCopyPreservesHash(t *testing.T) {
	board := NewBoard()
	move, _ := ParseMove("e2e4")
	board.MakeMove(move)

	boardCopy := board.Copy()

	if boardCopy.Hash != board.Hash {
		t.Errorf("copied board should have same hash: original %x, copy %x", board.Hash, boardCopy.Hash)
	}
}

func TestCaptureChangesHash(t *testing.T) {
	// Set up a position where a capture can happen
	board := NewBoard()
	moves := []string{"e2e4", "d7d5"}
	for _, moveStr := range moves {
		move, _ := ParseMove(moveStr)
		board.MakeMove(move)
	}

	hashBeforeCapture := board.Hash

	// Capture: e4xd5
	captureMove, _ := ParseMove("e4d5")
	board.MakeMove(captureMove)

	if board.Hash == hashBeforeCapture {
		t.Error("expected hash to change after capture")
	}

	// Verify computed hash matches
	computedHash := board.ComputeHash()
	if board.Hash != computedHash {
		t.Errorf("after capture: incremental hash %x != computed hash %x", board.Hash, computedHash)
	}
}

func TestCastlingChangesHash(t *testing.T) {
	// Set up kingside castling for white
	board := NewBoard()
	// Clear squares between king and rook
	board.Squares[NewSquare(5, 0)] = Piece(Empty) // f1
	board.Squares[NewSquare(6, 0)] = Piece(Empty) // g1
	board.Hash = board.ComputeHash()              // Recompute hash after manual setup

	hashBeforeCastling := board.Hash

	// Kingside castle: e1g1
	castleMove, _ := ParseMove("e1g1")
	board.MakeMove(castleMove)

	if board.Hash == hashBeforeCastling {
		t.Error("expected hash to change after castling")
	}

	// Verify computed hash matches
	computedHash := board.ComputeHash()
	if board.Hash != computedHash {
		t.Errorf("after castling: incremental hash %x != computed hash %x", board.Hash, computedHash)
	}
}

func TestEnPassantCaptureChangesHash(t *testing.T) {
	// Set up en passant position
	board := NewBoard()
	moves := []string{"e2e4", "a7a6", "e4e5", "d7d5"}
	for _, moveStr := range moves {
		move, _ := ParseMove(moveStr)
		board.MakeMove(move)
	}

	// Now white can capture en passant: e5xd6
	hashBeforeEp := board.Hash

	epMove, _ := ParseMove("e5d6")
	board.MakeMove(epMove)

	if board.Hash == hashBeforeEp {
		t.Error("expected hash to change after en passant capture")
	}

	// Verify computed hash matches
	computedHash := board.ComputeHash()
	if board.Hash != computedHash {
		t.Errorf("after en passant: incremental hash %x != computed hash %x", board.Hash, computedHash)
	}
}

func TestPromotionChangesHash(t *testing.T) {
	// Set up promotion position - put a white pawn on 7th rank
	board := NewBoard()
	// Clear the board except for kings
	for i := 0; i < 64; i++ {
		board.Squares[i] = Piece(Empty)
	}
	// Place kings
	board.Squares[NewSquare(4, 0)] = NewPiece(White, King) // e1
	board.Squares[NewSquare(4, 7)] = NewPiece(Black, King) // e8
	// Place white pawn on 7th rank
	board.Squares[NewSquare(0, 6)] = NewPiece(White, Pawn) // a7
	board.CastlingRights = 0
	board.EnPassantSq = -1
	board.Hash = board.ComputeHash()

	hashBeforePromotion := board.Hash

	// Promote: a7a8q
	promoteMove, _ := ParseMove("a7a8q")
	board.MakeMove(promoteMove)

	if board.Hash == hashBeforePromotion {
		t.Error("expected hash to change after promotion")
	}

	// Verify computed hash matches
	computedHash := board.ComputeHash()
	if board.Hash != computedHash {
		t.Errorf("after promotion: incremental hash %x != computed hash %x", board.Hash, computedHash)
	}
}

func TestEnPassantFileAffectsHash(t *testing.T) {
	// Test that en passant file is properly included in hash
	// Two positions that differ only in en passant file should have different hashes

	// Position 1: after e2e4
	board1 := NewBoard()
	move1, _ := ParseMove("e2e4")
	board1.MakeMove(move1)
	// Now Black to move, ep square is e3

	// Position 2: after d2d4
	board2 := NewBoard()
	move2, _ := ParseMove("d2d4")
	board2.MakeMove(move2)
	// Now Black to move, ep square is d3

	// Hashes should be different (different pawn positions AND different ep files)
	if board1.Hash == board2.Hash {
		t.Error("positions with different en passant files should have different hashes")
	}
}

func TestCastlingRightsAffectHash(t *testing.T) {
	// Create two boards with same piece positions but different castling rights
	board1 := NewBoard()
	board2 := NewBoard()

	// Remove castling rights from board2
	board2.CastlingRights = 0
	board2.Hash = board2.ComputeHash()

	if board1.Hash == board2.Hash {
		t.Error("positions with different castling rights should have different hashes")
	}
}

func TestSideToMoveAffectsHash(t *testing.T) {
	// Create two boards with same piece positions but different side to move
	board1 := NewBoard()
	board2 := NewBoard()

	// Change side to move in board2
	board2.ActiveColor = Black
	board2.Hash = board2.ComputeHash()

	if board1.Hash == board2.Hash {
		t.Error("positions with different side to move should have different hashes")
	}
}

func TestPieceZobristIndex(t *testing.T) {
	testCases := []struct {
		piece    Piece
		expected int
	}{
		{Piece(Empty), -1},           // Empty square
		{NewPiece(White, Pawn), 0},   // White Pawn
		{NewPiece(White, Knight), 1}, // White Knight
		{NewPiece(White, Bishop), 2}, // White Bishop
		{NewPiece(White, Rook), 3},   // White Rook
		{NewPiece(White, Queen), 4},  // White Queen
		{NewPiece(White, King), 5},   // White King
		{NewPiece(Black, Pawn), 6},   // Black Pawn
		{NewPiece(Black, Knight), 7}, // Black Knight
		{NewPiece(Black, Bishop), 8}, // Black Bishop
		{NewPiece(Black, Rook), 9},   // Black Rook
		{NewPiece(Black, Queen), 10}, // Black Queen
		{NewPiece(Black, King), 11},  // Black King
	}

	for _, tc := range testCases {
		got := pieceZobristIndex(tc.piece)
		if got != tc.expected {
			t.Errorf("pieceZobristIndex(%v) = %d, expected %d", tc.piece, got, tc.expected)
		}
	}
}
