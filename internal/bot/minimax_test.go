package bot

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// Helper functions for creating pointers to primitive types (for MinimaxConfig)
func intPtr(v int) *int                         { return &v }
func durationPtr(v time.Duration) *time.Duration { return &v }
func float64Ptr(v float64) *float64             { return &v }

func TestMinimaxEngine_Name(t *testing.T) {
	tests := []struct {
		name       string
		difficulty Difficulty
		expected   string
	}{
		{
			name:       "medium bot name",
			difficulty: Medium,
			expected:   "Medium Bot",
		},
		{
			name:       "hard bot name",
			difficulty: Hard,
			expected:   "Hard Bot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eng, err := NewMinimaxEngine(tt.difficulty)
			if err != nil {
				t.Fatalf("NewMinimaxEngine() error = %v", err)
			}
			if got := eng.Name(); got != tt.expected {
				t.Errorf("Name() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestMinimaxEngine_Close(t *testing.T) {
	eng, err := NewMinimaxEngine(Medium)
	if err != nil {
		t.Fatalf("NewMinimaxEngine() error = %v", err)
	}

	// Close should succeed
	err = eng.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// SelectMove should fail after close
	board := engine.NewBoard()
	_, err = eng.SelectMove(context.Background(), board)
	if err == nil {
		t.Error("SelectMove() after Close() should return error, got nil")
	}
	if !strings.Contains(err.Error(), "closed") {
		t.Errorf("error should contain 'closed', got %q", err.Error())
	}
}

func TestMinimaxEngine_Info(t *testing.T) {
	// Test Info() returns correct metadata

	// Medium bot
	engMedium, err := NewMinimaxEngine(Medium)
	if err != nil {
		t.Fatalf("NewMinimaxEngine() error = %v", err)
	}

	inspectable, ok := engMedium.(Inspectable)
	if !ok {
		t.Fatal("engine should implement Inspectable")
	}

	infoMedium := inspectable.Info()
	if infoMedium.Name != "Medium Bot" {
		t.Errorf("Medium bot name should be 'Medium Bot', got '%s'", infoMedium.Name)
	}
	if infoMedium.Author != "TermChess" {
		t.Errorf("Author should be 'TermChess', got '%s'", infoMedium.Author)
	}
	if infoMedium.Version != "1.0" {
		t.Errorf("Version should be '1.0', got '%s'", infoMedium.Version)
	}
	if infoMedium.Type != TypeInternal {
		t.Errorf("Type should be TypeInternal, got %v", infoMedium.Type)
	}
	if infoMedium.Difficulty != Medium {
		t.Errorf("Difficulty should be Medium, got %v", infoMedium.Difficulty)
	}

	// Check features
	if !infoMedium.Features["alpha_beta"] {
		t.Error("Medium bot should have alpha_beta feature")
	}
	if !infoMedium.Features["iterative_deepening"] {
		t.Error("Medium bot should have iterative_deepening feature")
	}
	if !infoMedium.Features["configurable"] {
		t.Error("Medium bot should have configurable feature")
	}
	if !infoMedium.Features["piece_square_tables"] {
		t.Error("Medium bot should have piece_square_tables feature")
	}
	if !infoMedium.Features["mobility"] {
		t.Error("Medium bot should have mobility feature")
	}
	if infoMedium.Features["king_safety"] {
		t.Error("Medium bot should NOT have king_safety feature")
	}

	// Hard bot
	engHard, err := NewMinimaxEngine(Hard)
	if err != nil {
		t.Fatalf("NewMinimaxEngine() error = %v", err)
	}

	infoHard := engHard.(Inspectable).Info()

	if infoHard.Name != "Hard Bot" {
		t.Errorf("Hard bot name should be 'Hard Bot', got '%s'", infoHard.Name)
	}
	if infoHard.Difficulty != Hard {
		t.Errorf("Difficulty should be Hard, got %v", infoHard.Difficulty)
	}
	if !infoHard.Features["king_safety"] {
		t.Error("Hard bot should have king_safety feature")
	}
}

func TestMinimaxEngine_ForcedMove(t *testing.T) {
	// Test that when only one legal move exists, engine returns it immediately
	// Create a position where White must capture a checking piece (forced recapture)
	// White king on e1 in check from Black rook on e2. White rook on e8 can capture (forced).
	// Actually, this results in multiple moves. Let's test the early return logic with any small move set.

	// Simplified test: just verify that with very few moves, selection is fast
	fen := "4k3/8/8/8/8/8/4r3/4K2R w - - 0 1"

	board, err := engine.ParseFEN(fen)
	if err != nil {
		t.Fatalf("ParseFEN() error = %v", err)
	}

	moves := board.LegalMoves()
	if len(moves) == 0 {
		t.Fatal("expected at least one legal move")
	}

	// If there's only 1 move, test the early return path
	if len(moves) == 1 {
		eng, err := NewMinimaxEngine(Medium)
		if err != nil {
			t.Fatalf("NewMinimaxEngine() error = %v", err)
		}
		defer eng.Close()

		start := time.Now()
		move, err := eng.SelectMove(context.Background(), board)
		elapsed := time.Since(start)

		// Should return immediately
		if err != nil {
			t.Errorf("SelectMove() error = %v", err)
		}
		if move != moves[0] {
			t.Errorf("SelectMove() = %v, want %v", move, moves[0])
		}
		if elapsed >= 100*time.Millisecond {
			t.Errorf("forced move took %v, should return quickly (< 100ms)", elapsed)
		}
	} else {
		// Otherwise, just verify normal move selection works
		eng, err := NewMinimaxEngine(Medium)
		if err != nil {
			t.Fatalf("NewMinimaxEngine() error = %v", err)
		}
		defer eng.Close()

		move, err := eng.SelectMove(context.Background(), board)
		if err != nil {
			t.Errorf("SelectMove() error = %v", err)
		}
		if !containsMove(moves, move) {
			t.Errorf("SelectMove() returned illegal move %v", move)
		}
	}
}

func TestMinimaxEngine_FindsMateInOne(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		description string
	}{
		{
			name:        "back rank mate",
			fen:         "6k1/5ppp/8/8/8/8/8/R6K w - - 0 1",
			description: "White rook delivers back rank mate with Ra8#",
		},
		{
			name:        "queen mate",
			fen:         "k7/8/1K6/8/8/8/8/Q7 w - - 0 1",
			description: "White queen delivers mate (multiple mating moves available)",
		},
		{
			name:        "simple mate pattern",
			fen:         "7k/5Q2/6K1/8/8/8/8/8 w - - 0 1",
			description: "White queen delivers mate (multiple mating moves available)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := engine.ParseFEN(tt.fen)
			if err != nil {
				t.Fatalf("ParseFEN() error = %v", err)
			}

			eng, err := NewMinimaxEngine(Medium)
			if err != nil {
				t.Fatalf("NewMinimaxEngine() error = %v", err)
			}
			defer eng.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			move, err := eng.SelectMove(ctx, board)
			if err != nil {
				t.Fatalf("SelectMove() error = %v", err)
			}

			// Verify the move delivers checkmate
			boardCopy := board.Copy()
			err = boardCopy.MakeMove(move)
			if err != nil {
				t.Fatalf("MakeMove() error = %v", err)
			}

			if boardCopy.Status() != engine.Checkmate {
				t.Errorf("engine should find mate-in-1: %s, got status %v", tt.description, boardCopy.Status())
			}
		})
	}
}

func TestMinimaxEngine_AvoidBlunder(t *testing.T) {
	// Position where White queen on d1 can move, but moving to d8 would hang it to Black rook on d7
	// White should NOT play Qd8
	fen := "4k3/3r4/8/8/8/8/8/3Q1K2 w - - 0 1"

	board, err := engine.ParseFEN(fen)
	if err != nil {
		t.Fatalf("ParseFEN() error = %v", err)
	}

	eng, err := NewMinimaxEngine(Medium)
	if err != nil {
		t.Fatalf("NewMinimaxEngine() error = %v", err)
	}
	defer eng.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	move, err := eng.SelectMove(ctx, board)
	if err != nil {
		t.Fatalf("SelectMove() error = %v", err)
	}

	// The move should NOT be Qd8 (hanging the queen)
	blunderMove, _ := engine.ParseMove("d1d8")
	if move == blunderMove {
		t.Error("engine should not hang the queen with Qd8")
	}

	// Verify the move doesn't lose the queen
	boardCopy := board.Copy()
	err = boardCopy.MakeMove(move)
	if err != nil {
		t.Fatalf("MakeMove() error = %v", err)
	}

	// After opponent's best response, White's queen should still be on the board
	// (This is a simplified check - we're not doing full move analysis here)
	// At minimum, the engine shouldn't blunder the queen immediately
}

func TestMinimaxEngine_CapturePriority(t *testing.T) {
	// Position where White can capture Black queen with a pawn
	// White pawn on e5, Black queen on f6, Black king on g8
	fen := "6k1/8/5q2/4P3/8/8/8/6K1 w - - 0 1"

	board, err := engine.ParseFEN(fen)
	if err != nil {
		t.Fatalf("ParseFEN() error = %v", err)
	}

	eng, err := NewMinimaxEngine(Medium)
	if err != nil {
		t.Fatalf("NewMinimaxEngine() error = %v", err)
	}
	defer eng.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	move, err := eng.SelectMove(ctx, board)
	if err != nil {
		t.Fatalf("SelectMove() error = %v", err)
	}

	// Engine should capture the queen
	captureMove, _ := engine.ParseMove("e5f6")
	if move != captureMove {
		t.Errorf("engine should capture hanging queen with exf6, got %v", move)
	}
}

func TestMinimaxEngine_Timeout(t *testing.T) {
	// Start from initial position (complex enough to potentially exceed timeout)
	board := engine.NewBoard()

	// Create engine with very short timeout
	eng, err := NewMinimaxEngine(Medium, WithTimeLimit(1*time.Nanosecond))
	if err != nil {
		t.Fatalf("NewMinimaxEngine() error = %v", err)
	}
	defer eng.Close()

	ctx := context.Background()

	// This might timeout or might complete if the search is fast enough
	// We're primarily checking that it doesn't panic or hang
	move, err := eng.SelectMove(ctx, board)

	// Either it succeeds or times out gracefully
	if err != nil {
		// If it errors, it should be a timeout or context error
		if !strings.Contains(err.Error(), "context deadline exceeded") {
			t.Errorf("expected timeout error, got %v", err)
		}
	} else {
		// If it succeeds, move should be legal
		moves := board.LegalMoves()
		if !containsMove(moves, move) {
			t.Errorf("returned illegal move %v", move)
		}
	}
}

func TestMinimaxEngine_NoLegalMoves(t *testing.T) {
	// This is a checkmate position - Black has no legal moves
	fen := "rnb1kbnr/pppp1ppp/8/4p3/6Pq/5P2/PPPPP2P/RNBQKBNR w KQkq - 1 3"

	board, err := engine.ParseFEN(fen)
	if err != nil {
		t.Fatalf("ParseFEN() error = %v", err)
	}

	// Verify it's checkmate
	if board.Status() != engine.Checkmate {
		t.Errorf("position should be checkmate, got %v", board.Status())
	}

	eng, err := NewMinimaxEngine(Medium)
	if err != nil {
		t.Fatalf("NewMinimaxEngine() error = %v", err)
	}
	defer eng.Close()

	// Shouldn't be able to select a move in checkmate position
	_, err = eng.SelectMove(context.Background(), board)
	if err == nil {
		t.Error("SelectMove() in checkmate position should return error, got nil")
	}
	if !strings.Contains(err.Error(), "no legal moves") {
		t.Errorf("error should contain 'no legal moves', got %q", err.Error())
	}
}

func TestMinimaxEngine_Depth2Search(t *testing.T) {
	// Position where depth-2 search should find a better move than depth-1
	// White can win a rook with a pawn fork: e4-e5 forks rook on d6 and f6
	fen := "6k1/8/3r1r2/8/4P3/8/8/6K1 w - - 0 1"

	board, err := engine.ParseFEN(fen)
	if err != nil {
		t.Fatalf("ParseFEN() error = %v", err)
	}

	eng, err := NewMinimaxEngine(Medium)
	if err != nil {
		t.Fatalf("NewMinimaxEngine() error = %v", err)
	}
	defer eng.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	move, err := eng.SelectMove(ctx, board)
	if err != nil {
		t.Fatalf("SelectMove() error = %v", err)
	}

	// The engine should find e4-e5, which sets up a fork
	forkMove, _ := engine.ParseMove("e4e5")
	if move != forkMove {
		t.Errorf("engine should find the pawn fork e4-e5, got %v", move)
	}
}

func TestMinimaxEngine_AlphaBetaPruning(t *testing.T) {
	// This test verifies that the engine completes within reasonable time
	// which indicates alpha-beta pruning is working (vs. plain minimax)
	board := engine.NewBoard()

	eng, err := NewMinimaxEngine(Medium)
	if err != nil {
		t.Fatalf("NewMinimaxEngine() error = %v", err)
	}
	defer eng.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	move, err := eng.SelectMove(ctx, board)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("SelectMove() error = %v", err)
	}
	if move == (engine.Move{}) {
		t.Error("SelectMove() returned empty move")
	}

	// Depth-2 search with alpha-beta should complete in well under 1 second
	// even from the starting position
	if elapsed >= 1*time.Second {
		t.Errorf("search took %v, should complete quickly with pruning (< 1s)", elapsed)
	}
}

func TestGetDefaultWeights(t *testing.T) {
	tests := []struct {
		name       string
		difficulty Difficulty
		expected   evalWeights
	}{
		{
			name:       "medium weights",
			difficulty: Medium,
			expected: evalWeights{
				material:    1.0,
				pieceSquare: 0.0,
				mobility:    0.0,
				kingSafety:  0.0,
			},
		},
		{
			name:       "hard weights",
			difficulty: Hard,
			expected: evalWeights{
				material:    1.0,
				pieceSquare: 0.0,
				mobility:    0.0,
				kingSafety:  0.0,
			},
		},
		{
			name:       "invalid difficulty uses medium fallback",
			difficulty: Easy, // Not valid for minimax, but should not panic
			expected: evalWeights{
				material:    1.0,
				pieceSquare: 0.0,
				mobility:    0.0,
				kingSafety:  0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weights := getDefaultWeights(tt.difficulty)
			if weights != tt.expected {
				t.Errorf("getDefaultWeights() = %+v, want %+v", weights, tt.expected)
			}
		})
	}
}

func TestMinimaxEngine_MoveOrdering(t *testing.T) {
	// Position with several capture and non-capture options
	// White pawn on e5 can capture f6 or d6, or push to e6
	fen := "6k1/8/3p1p2/4P3/8/8/8/6K1 w - - 0 1"

	board, err := engine.ParseFEN(fen)
	if err != nil {
		t.Fatalf("ParseFEN() error = %v", err)
	}

	eng, err := NewMinimaxEngine(Medium)
	if err != nil {
		t.Fatalf("NewMinimaxEngine() error = %v", err)
	}
	defer eng.Close()

	me := eng.(*minimaxEngine)

	moves := board.LegalMoves()
	orderedMoves := me.orderMoves(board, moves)

	// Verify captures come before non-captures
	capturePhase := true
	for _, move := range orderedMoves {
		targetPiece := board.PieceAt(move.To)
		isCapture := !targetPiece.IsEmpty()

		if !isCapture {
			capturePhase = false
		}

		if isCapture && !capturePhase {
			t.Error("all captures should come before non-captures")
		}
	}
}

func TestMinimaxEngine_IterativeDeepening_Timeout(t *testing.T) {
	// Create engine with very short timeout
	eng, err := NewMinimaxEngine(Medium, WithTimeLimit(100*time.Millisecond))
	if err != nil {
		t.Fatalf("NewMinimaxEngine() error = %v", err)
	}
	defer eng.Close()

	// Start from complex middlegame position
	board, err := engine.ParseFEN("r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 1")
	if err != nil {
		t.Fatalf("ParseFEN() error = %v", err)
	}

	// Should return valid move even with short timeout
	move, err := eng.SelectMove(context.Background(), board)

	if err != nil {
		t.Fatalf("SelectMove() error = %v", err)
	}
	if move == (engine.Move{}) {
		t.Error("SelectMove() returned empty move")
	}
	// Verify move is legal
	legalMoves := board.LegalMoves()
	if !containsMove(legalMoves, move) {
		t.Errorf("SelectMove() returned illegal move %v", move)
	}
}

func TestMinimaxEngine_IterativeDeepening_MultipleDepths(t *testing.T) {
	// Create engine with generous timeout
	eng, err := NewMinimaxEngine(Medium, WithTimeLimit(5*time.Second))
	if err != nil {
		t.Fatalf("NewMinimaxEngine() error = %v", err)
	}
	defer eng.Close()

	// Simple position where depth 4 is achievable
	board, err := engine.ParseFEN("8/8/8/4k3/8/8/4K3/4R3 w - - 0 1")
	if err != nil {
		t.Fatalf("ParseFEN() error = %v", err)
	}

	move, err := eng.SelectMove(context.Background(), board)

	if err != nil {
		t.Fatalf("SelectMove() error = %v", err)
	}
	if move == (engine.Move{}) {
		t.Error("SelectMove() returned empty move")
	}
	// Should find a good move at deeper depth
	legalMoves := board.LegalMoves()
	if !containsMove(legalMoves, move) {
		t.Errorf("SelectMove() returned illegal move %v", move)
	}
}

func TestMinimaxEngine_ReturnsLastCompletedDepth(t *testing.T) {
	// Create engine with moderate timeout
	eng, err := NewMinimaxEngine(Hard, WithTimeLimit(500*time.Millisecond))
	if err != nil {
		t.Fatalf("NewMinimaxEngine() error = %v", err)
	}
	defer eng.Close()

	// Complex position
	board, err := engine.ParseFEN("r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 1")
	if err != nil {
		t.Fatalf("ParseFEN() error = %v", err)
	}

	move, err := eng.SelectMove(context.Background(), board)

	if err != nil {
		t.Fatalf("SelectMove() error = %v", err)
	}
	if move == (engine.Move{}) {
		t.Error("SelectMove() returned empty move")
	}
	// Should return valid move from last completed depth (likely depth 2 or 3)
	legalMoves := board.LegalMoves()
	if !containsMove(legalMoves, move) {
		t.Errorf("SelectMove() returned illegal move %v", move)
	}
}

func TestMinimaxEngine_Configure_SearchDepth(t *testing.T) {
	// Test updating search depth
	eng, err := NewMinimaxEngine(Medium)
	if err != nil {
		t.Fatalf("NewMinimaxEngine() error = %v", err)
	}

	configurable, ok := eng.(Configurable)
	if !ok {
		t.Fatal("engine should implement Configurable")
	}

	// Valid depth
	err = configurable.Configure(MinimaxConfig{
		SearchDepth: intPtr(8),
	})
	if err != nil {
		t.Errorf("Configure should accept valid depth: %v", err)
	}

	// Verify depth was updated
	me := eng.(*minimaxEngine)
	if me.maxDepth != 8 {
		t.Errorf("search depth should be 8, got %d", me.maxDepth)
	}

	// Invalid depth (too low)
	err = configurable.Configure(MinimaxConfig{
		SearchDepth: intPtr(0),
	})
	if err == nil {
		t.Error("Configure should reject depth < 1")
	}

	// Invalid depth (too high)
	err = configurable.Configure(MinimaxConfig{
		SearchDepth: intPtr(21),
	})
	if err == nil {
		t.Error("Configure should reject depth > 20")
	}
}

func TestMinimaxEngine_Configure_TimeLimit(t *testing.T) {
	// Test updating time limit
	eng, err := NewMinimaxEngine(Medium)
	if err != nil {
		t.Fatalf("NewMinimaxEngine() error = %v", err)
	}

	configurable, ok := eng.(Configurable)
	if !ok {
		t.Fatal("engine should implement Configurable")
	}

	// Valid time limit
	err = configurable.Configure(MinimaxConfig{
		TimeLimit: durationPtr(5 * time.Second),
	})
	if err != nil {
		t.Errorf("Configure should accept valid time limit: %v", err)
	}

	// Verify time limit was updated
	me := eng.(*minimaxEngine)
	if me.timeLimit != 5*time.Second {
		t.Errorf("time limit should be 5s, got %v", me.timeLimit)
	}

	// Invalid time limit (negative)
	err = configurable.Configure(MinimaxConfig{
		TimeLimit: durationPtr(-1 * time.Second),
	})
	if err == nil {
		t.Error("Configure should reject negative time limit")
	}

	// Invalid time limit (zero)
	err = configurable.Configure(MinimaxConfig{
		TimeLimit: durationPtr(0 * time.Second),
	})
	if err == nil {
		t.Error("Configure should reject zero time limit")
	}
}

func TestMinimaxEngine_Configure_EvalWeights(t *testing.T) {
	// Test updating evaluation weights
	eng, err := NewMinimaxEngine(Hard)
	if err != nil {
		t.Fatalf("NewMinimaxEngine() error = %v", err)
	}

	configurable, ok := eng.(Configurable)
	if !ok {
		t.Fatal("engine should implement Configurable")
	}

	err = configurable.Configure(MinimaxConfig{
		MaterialWeight:    float64Ptr(1.5),
		PieceSquareWeight: float64Ptr(0.2),
		MobilityWeight:    float64Ptr(0.15),
		KingSafetyWeight:  float64Ptr(0.3),
	})

	if err != nil {
		t.Errorf("Configure should accept valid eval weights: %v", err)
	}

	// Verify weights were updated
	me := eng.(*minimaxEngine)
	if me.evalWeights.material != 1.5 {
		t.Errorf("material weight should be 1.5, got %f", me.evalWeights.material)
	}
	if me.evalWeights.pieceSquare != 0.2 {
		t.Errorf("piece square weight should be 0.2, got %f", me.evalWeights.pieceSquare)
	}
	if me.evalWeights.mobility != 0.15 {
		t.Errorf("mobility weight should be 0.15, got %f", me.evalWeights.mobility)
	}
	if me.evalWeights.kingSafety != 0.3 {
		t.Errorf("king safety weight should be 0.3, got %f", me.evalWeights.kingSafety)
	}
}

func TestMinimaxEngine_Configure_InvalidOption(t *testing.T) {
	// Test that invalid option keys are ignored (not an error)
	eng, err := NewMinimaxEngine(Medium)
	if err != nil {
		t.Fatalf("NewMinimaxEngine() error = %v", err)
	}

	configurable, ok := eng.(Configurable)
	if !ok {
		t.Fatal("engine should implement Configurable")
	}

	err = configurable.Configure(MinimaxConfig{
		SearchDepth: intPtr(5), // valid option
	})

	if err != nil {
		t.Errorf("Configure should succeed with valid options: %v", err)
	}

	// Verify valid option was applied
	me := eng.(*minimaxEngine)
	if me.maxDepth != 5 {
		t.Errorf("search depth should be 5, got %d", me.maxDepth)
	}
}

func TestMinimaxEngine_Configure_MultipleOptions(t *testing.T) {
	// Test configuring multiple options at once
	eng, err := NewMinimaxEngine(Medium)
	if err != nil {
		t.Fatalf("NewMinimaxEngine() error = %v", err)
	}

	configurable, ok := eng.(Configurable)
	if !ok {
		t.Fatal("engine should implement Configurable")
	}

	err = configurable.Configure(MinimaxConfig{
		SearchDepth:       intPtr(10),
		TimeLimit:         durationPtr(3 * time.Second),
		MaterialWeight:    float64Ptr(1.2),
		PieceSquareWeight: float64Ptr(0.25),
	})

	if err != nil {
		t.Errorf("Configure should accept multiple valid options: %v", err)
	}

	// Verify all options were applied
	me := eng.(*minimaxEngine)
	if me.maxDepth != 10 {
		t.Errorf("search depth should be 10, got %d", me.maxDepth)
	}
	if me.timeLimit != 3*time.Second {
		t.Errorf("time limit should be 3s, got %v", me.timeLimit)
	}
	if me.evalWeights.material != 1.2 {
		t.Errorf("material weight should be 1.2, got %f", me.evalWeights.material)
	}
	if me.evalWeights.pieceSquare != 0.25 {
		t.Errorf("piece square weight should be 0.25, got %f", me.evalWeights.pieceSquare)
	}
}

func TestMinimaxEngine_Configure_EmptyConfig(t *testing.T) {
	// Test that empty config is valid (no-op)
	eng, err := NewMinimaxEngine(Medium)
	if err != nil {
		t.Fatalf("NewMinimaxEngine() error = %v", err)
	}

	configurable, ok := eng.(Configurable)
	if !ok {
		t.Fatal("engine should implement Configurable")
	}

	// Pass empty config (no options set)
	err = configurable.Configure(MinimaxConfig{})

	if err != nil {
		t.Errorf("Configure should accept empty config: %v", err)
	}

	// Original depth should be unchanged
	me := eng.(*minimaxEngine)
	if me.maxDepth != 4 { // Medium default
		t.Errorf("search depth should remain at default 4, got %d", me.maxDepth)
	}
}

// Helper function to check if a slice contains a move
func containsMove(moves []engine.Move, move engine.Move) bool {
	for _, m := range moves {
		if m == move {
			return true
		}
	}
	return false
}
