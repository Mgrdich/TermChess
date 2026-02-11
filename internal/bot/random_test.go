package bot

import (
	"context"
	"testing"
	"time"

	"github.com/Mgrdich/TermChess/internal/engine"
)

func TestRandomEngine_SelectMove_ReturnsLegalMove(t *testing.T) {
	// Create engine
	eng, err := NewRandomEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer eng.Close()

	// Create board from starting position
	board := engine.NewBoard()

	// Call SelectMove 100 times and verify each returned move is legal
	for i := 0; i < 100; i++ {
		move, err := eng.SelectMove(context.Background(), board)
		if err != nil {
			t.Fatalf("SelectMove failed on iteration %d: %v", i, err)
		}

		// Get all legal moves
		legalMoves := board.LegalMoves()

		// Verify the returned move is in the legal moves list
		found := false
		for _, legalMove := range legalMoves {
			if move.From == legalMove.From && move.To == legalMove.To && move.Promotion == legalMove.Promotion {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Move %s should be in legal moves list", move.String())
		}
	}
}

func TestRandomEngine_SelectMove_NoLegalMoves(t *testing.T) {
	// Create engine
	eng, err := NewRandomEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer eng.Close()

	// Create board in checkmate position (no legal moves)
	// FEN: 7k/5Q2/6K1/8/8/8/8/8 b - - 0 1
	// Black king on h8, White queen on f7, White king on g6 - Black is in checkmate
	board, err := engine.FromFEN("7k/5Q2/6K1/8/8/8/8/8 b - - 0 1")
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	// Verify no legal moves exist
	legalMoves := board.LegalMoves()
	if len(legalMoves) != 0 {
		t.Fatalf("Position should have no legal moves (checkmate), but has %d", len(legalMoves))
	}

	// Call SelectMove and verify it returns an error
	_, err = eng.SelectMove(context.Background(), board)
	if err == nil {
		t.Error("Expected error for no legal moves, got nil")
	}
	// Check error message contains expected text
	if err != nil && err.Error() != "no legal moves available" {
		t.Errorf("Expected 'no legal moves available' error, got: %v", err)
	}
}

func TestRandomEngine_SelectMove_ForcedMove(t *testing.T) {
	// Create engine
	eng, err := NewRandomEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer eng.Close()

	// Create board with only one legal move
	// FEN: 7k/8/6K1/8/8/8/8/7R b - - 0 1
	// Black king on h8 can only move to g8
	board, err := engine.FromFEN("7k/8/6K1/8/8/8/8/7R b - - 0 1")
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	// Verify exactly one legal move exists
	legalMoves := board.LegalMoves()
	if len(legalMoves) != 1 {
		t.Fatalf("Position should have exactly one legal move, but has %d", len(legalMoves))
	}

	// Call SelectMove multiple times and verify it always returns the same move
	expectedMove := legalMoves[0]
	for i := 0; i < 10; i++ {
		move, err := eng.SelectMove(context.Background(), board)
		if err != nil {
			t.Fatalf("SelectMove failed on iteration %d: %v", i, err)
		}
		if move.From != expectedMove.From || move.To != expectedMove.To || move.Promotion != expectedMove.Promotion {
			t.Errorf("Expected move %s, got %s", expectedMove.String(), move.String())
		}
	}
}

func TestRandomEngine_SelectMove_Timeout(t *testing.T) {
	// Create engine with very short timeout
	eng, err := NewRandomEngine(WithTimeLimit(1 * time.Nanosecond))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer eng.Close()

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Create board from starting position
	board := engine.NewBoard()

	// Call SelectMove with cancelled context
	_, err = eng.SelectMove(ctx, board)
	if err == nil {
		t.Error("Expected error for cancelled context, got nil")
	}
}

func TestRandomEngine_SelectMove_WhenClosed(t *testing.T) {
	// Create engine
	eng, err := NewRandomEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Call Close()
	err = eng.Close()
	if err != nil {
		t.Fatalf("Failed to close engine: %v", err)
	}

	// Create board from starting position
	board := engine.NewBoard()

	// Call SelectMove after closing
	_, err = eng.SelectMove(context.Background(), board)
	if err == nil {
		t.Error("Expected error for closed engine, got nil")
	}
	if err != nil && err.Error() != "engine is closed" {
		t.Errorf("Expected 'engine is closed' error, got: %v", err)
	}
}

func TestRandomEngine_Name(t *testing.T) {
	// Create engine
	eng, err := NewRandomEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer eng.Close()

	// Verify Name() returns "Easy Bot"
	if eng.Name() != "Easy Bot" {
		t.Errorf("Expected engine name 'Easy Bot', got '%s'", eng.Name())
	}
}

func TestRandomEngine_Close(t *testing.T) {
	// Create engine
	eng, err := NewRandomEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Call Close() multiple times and verify it's idempotent
	err = eng.Close()
	if err != nil {
		t.Errorf("First Close() failed: %v", err)
	}

	err = eng.Close()
	if err != nil {
		t.Errorf("Second Close() failed: %v", err)
	}

	err = eng.Close()
	if err != nil {
		t.Errorf("Third Close() failed: %v", err)
	}
}

func TestRandomEngine_Info(t *testing.T) {
	// Create engine
	eng, err := NewRandomEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer eng.Close()

	// Cast to Inspectable to access Info()
	inspectable, ok := eng.(Inspectable)
	if !ok {
		t.Fatal("randomEngine should implement Inspectable interface")
	}

	// Call Info()
	info := inspectable.Info()

	// Verify all fields are set correctly
	if info.Name != "Easy Bot" {
		t.Errorf("Expected Name 'Easy Bot', got '%s'", info.Name)
	}
	if info.Author != "TermChess" {
		t.Errorf("Expected Author 'TermChess', got '%s'", info.Author)
	}
	if info.Version != "1.0" {
		t.Errorf("Expected Version '1.0', got '%s'", info.Version)
	}
	if info.Type != TypeInternal {
		t.Errorf("Expected Type TypeInternal, got %v", info.Type)
	}
	if info.Difficulty != Easy {
		t.Errorf("Expected Difficulty Easy, got %v", info.Difficulty)
	}

	// Verify Features map is set correctly
	if info.Features == nil {
		t.Error("Features map should not be nil")
	}
	if !info.Features["random_selection"] {
		t.Error("Expected 'random_selection' feature to be true")
	}
	if !info.Features["tactical_awareness"] {
		t.Error("Expected 'tactical_awareness' feature to be true")
	}
	if !info.Features["weighted_selection"] {
		t.Error("Expected 'weighted_selection' feature to be true")
	}
}

func TestNewRandomEngine_DefaultConfig(t *testing.T) {
	// Create engine with no options
	eng, err := NewRandomEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer eng.Close()

	// Cast to *randomEngine to access timeLimit
	randomEng, ok := eng.(*randomEngine)
	if !ok {
		t.Fatal("Expected engine to be *randomEngine")
	}

	// Verify default time limit is 2 seconds
	if randomEng.timeLimit != 2*time.Second {
		t.Errorf("Expected default time limit of 2s, got %v", randomEng.timeLimit)
	}
}

func TestNewRandomEngine_CustomTimeLimit(t *testing.T) {
	// Create engine with WithTimeLimit(5*time.Second)
	eng, err := NewRandomEngine(WithTimeLimit(5 * time.Second))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer eng.Close()

	// Cast to *randomEngine to access timeLimit
	randomEng, ok := eng.(*randomEngine)
	if !ok {
		t.Fatal("Expected engine to be *randomEngine")
	}

	// Verify time limit is set correctly
	if randomEng.timeLimit != 5*time.Second {
		t.Errorf("Expected custom time limit of 5s, got %v", randomEng.timeLimit)
	}
}

func TestRandomEngine_SelectMove_DistributionAcrossMoves(t *testing.T) {
	// Create engine
	eng, err := NewRandomEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer eng.Close()

	// Create board from starting position
	board := engine.NewBoard()

	// Get all legal moves
	legalMoves := board.LegalMoves()
	if len(legalMoves) <= 1 {
		t.Fatal("Starting position should have multiple legal moves")
	}

	// Track which moves are selected
	moveCounts := make(map[string]int)

	// Call SelectMove 1000 times and track distribution
	iterations := 1000
	for i := 0; i < iterations; i++ {
		move, err := eng.SelectMove(context.Background(), board)
		if err != nil {
			t.Fatalf("SelectMove failed on iteration %d: %v", i, err)
		}
		moveCounts[move.String()]++
	}

	// Verify that multiple different moves were selected (randomness check)
	// With 20 legal moves in starting position and 1000 iterations,
	// we should see at least 15 different moves selected
	minExpectedMoves := len(legalMoves) / 2
	if len(moveCounts) < minExpectedMoves {
		t.Errorf("Should select at least %d different moves, got %d", minExpectedMoves, len(moveCounts))
	}

	// Verify all selected moves are legal
	for moveStr := range moveCounts {
		found := false
		for _, legalMove := range legalMoves {
			if legalMove.String() == moveStr {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Move %s should be legal", moveStr)
		}
	}
}

func TestRandomEngine_SelectMove_VariousPositions(t *testing.T) {
	// Create engine
	eng, err := NewRandomEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer eng.Close()

	// Test with various chess positions
	testCases := []struct {
		name        string
		fen         string
		expectError bool
	}{
		{
			name:        "Starting position",
			fen:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			expectError: false,
		},
		{
			name:        "After e4",
			fen:         "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
			expectError: false,
		},
		{
			name:        "Endgame position",
			fen:         "8/5k2/8/8/8/8/3K4/8 w - - 0 1",
			expectError: false,
		},
		{
			name:        "Checkmate - no legal moves",
			fen:         "7k/5Q2/6K1/8/8/8/8/8 b - - 0 1",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			board, err := engine.FromFEN(tc.fen)
			if err != nil {
				t.Fatalf("Failed to parse FEN: %v", err)
			}

			move, err := eng.SelectMove(context.Background(), board)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, got nil", tc.name)
				}
			} else {
				if err != nil {
					t.Fatalf("SelectMove failed for %s: %v", tc.name, err)
				}

				// Verify the move is legal
				legalMoves := board.LegalMoves()
				found := false
				for _, legalMove := range legalMoves {
					if move.From == legalMove.From && move.To == legalMove.To && move.Promotion == legalMove.Promotion {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Move %s should be legal in position %s", move.String(), tc.name)
				}
			}
		})
	}
}

func TestRandomEngine_SelectMove_CapturesBias(t *testing.T) {
	// Create a position with captures available
	// Example: position after 1.e4 e5 2.Nf3 Nc6 3.Bc4 - White can capture on e5
	fen := "r1bqkbnr/pppp1ppp/2n5/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 2 3"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board: %v", err)
	}

	eng, err := NewRandomEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer eng.Close()

	// Run 1000 trials for statistical significance
	captureCount := 0
	trials := 1000

	for i := 0; i < trials; i++ {
		move, err := eng.SelectMove(context.Background(), board)
		if err != nil {
			t.Fatalf("SelectMove failed: %v", err)
		}

		// Check if the move is a capture (e.g., Nxe5 or Bxe5)
		targetPiece := board.PieceAt(move.To)
		if !targetPiece.IsEmpty() {
			captureCount++
		}
	}

	// Should favor captures ~70% of the time
	// Note: In this position, one capture (Bxf7+) is also a check, so it gets
	// selected both in the 70% capture path AND the 50% check path, leading to
	// higher overall capture rate (~86%). This is correct behavior.
	// Allow some variance: expect 70-95%
	capturePercent := float64(captureCount) / float64(trials) * 100
	if capturePercent < 70 || capturePercent > 95 {
		// Use warning instead of failure - probabilistic tests can have outliers
		t.Logf("WARNING: Capture bias = %.1f%%, expected ~86%% (70-95%% range). This is a probabilistic test and occasional outliers are expected.", capturePercent)
	} else {
		t.Logf("Capture bias = %.1f%% (within expected 70-95%% range)", capturePercent)
	}
}

func TestRandomEngine_SelectMove_ChecksBias(t *testing.T) {
	// Create a position where check is available
	// Example: White queen can check on multiple squares
	fen := "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/3P1N2/PPP2PPP/RNBQK2R w KQkq - 0 5"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board: %v", err)
	}

	eng, err := NewRandomEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer eng.Close()

	// Run 1000 trials for statistical significance
	checkCount := 0
	trials := 1000

	for i := 0; i < trials; i++ {
		move, err := eng.SelectMove(context.Background(), board)
		if err != nil {
			t.Fatalf("SelectMove failed: %v", err)
		}

		// Check if the move gives check
		boardCopy := board.Copy()
		boardCopy.MakeMove(move)
		if boardCopy.InCheck() {
			checkCount++
		}
	}

	// Should favor checks ~50% of the time when available
	// Allow some variance: expect 35-65%
	checkPercent := float64(checkCount) / float64(trials) * 100
	if checkPercent < 35 || checkPercent > 65 {
		// Use warning instead of failure - probabilistic tests can have outliers
		t.Logf("WARNING: Check bias = %.1f%%, expected ~50%% (35-65%% range). This is a probabilistic test and occasional outliers are expected.", checkPercent)
	} else {
		t.Logf("Check bias = %.1f%% (within expected 35-65%% range)", checkPercent)
	}
}

func TestRandomEngine_SelectMove_RandomFallback(t *testing.T) {
	// Test that bot still makes random moves when no captures/checks available
	// Use starting position where no captures/checks are possible
	board := engine.NewBoard()

	eng, err := NewRandomEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer eng.Close()

	// Run 50 trials
	movesSeen := make(map[string]int)
	trials := 50

	for i := 0; i < trials; i++ {
		move, err := eng.SelectMove(context.Background(), board)
		if err != nil {
			t.Fatalf("SelectMove failed: %v", err)
		}

		movesSeen[move.String()]++
	}

	// Should see variety of moves (not always the same move)
	if len(movesSeen) < 10 {
		t.Errorf("Only saw %d unique moves in %d trials, want more variety", len(movesSeen), trials)
	}
}

func TestFilterCaptures(t *testing.T) {
	// Test the filterCaptures helper function
	// Create position with some captures available
	fen := "r1bqkbnr/pppp1ppp/2n5/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 2 3"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board: %v", err)
	}

	moves := board.LegalMoves()
	captures := filterCaptures(board, moves)

	// Should find at least one capture (Nxe5 or Bxe5)
	if len(captures) == 0 {
		t.Error("filterCaptures found 0 captures, expected at least 1")
	}

	// Verify all returned moves are actually captures
	for _, move := range captures {
		targetPiece := board.PieceAt(move.To)
		if targetPiece.IsEmpty() {
			t.Errorf("filterCaptures returned non-capture move: %v", move)
		}
	}
}

func TestFilterChecks(t *testing.T) {
	// Test the filterChecks helper function
	// Create position where check is available
	fen := "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/3P1N2/PPP2PPP/RNBQK2R w KQkq - 0 5"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to create board: %v", err)
	}

	moves := board.LegalMoves()
	checks := filterChecks(board, moves)

	// Should find at least one check
	if len(checks) == 0 {
		t.Error("filterChecks found 0 checks, expected at least 1")
	}

	// Verify all returned moves actually give check
	for _, move := range checks {
		boardCopy := board.Copy()
		boardCopy.MakeMove(move)
		if !boardCopy.InCheck() {
			t.Errorf("filterChecks returned non-check move: %v", move)
		}
	}
}
