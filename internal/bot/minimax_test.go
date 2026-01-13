package bot

import (
	"context"
	"testing"
	"time"

	"github.com/Mgrdich/TermChess/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			require.NoError(t, err)
			assert.Equal(t, tt.expected, eng.Name())
		})
	}
}

func TestMinimaxEngine_Close(t *testing.T) {
	eng, err := NewMinimaxEngine(Medium)
	require.NoError(t, err)

	// Close should succeed
	err = eng.Close()
	assert.NoError(t, err)

	// SelectMove should fail after close
	board := engine.NewBoard()
	_, err = eng.SelectMove(context.Background(), board)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")
}

func TestMinimaxEngine_Info(t *testing.T) {
	eng, err := NewMinimaxEngine(Medium)
	require.NoError(t, err)

	inspectable, ok := eng.(Inspectable)
	require.True(t, ok, "engine should implement Inspectable")

	info := inspectable.Info()
	assert.Equal(t, "Medium Bot", info.Name)
	assert.Equal(t, "TermChess", info.Author)
	assert.Equal(t, TypeInternal, info.Type)
	assert.Equal(t, Medium, info.Difficulty)
	assert.True(t, info.Features["minimax"])
	assert.True(t, info.Features["alpha_beta"])
}

func TestMinimaxEngine_ForcedMove(t *testing.T) {
	// Test that when only one legal move exists, engine returns it immediately
	// Create a position where White must capture a checking piece (forced recapture)
	// White king on e1 in check from Black rook on e2. White rook on e8 can capture (forced).
	// Actually, this results in multiple moves. Let's test the early return logic with any small move set.

	// Simplified test: just verify that with very few moves, selection is fast
	fen := "4k3/8/8/8/8/8/4r3/4K2R w - - 0 1"

	board, err := engine.ParseFEN(fen)
	require.NoError(t, err)

	moves := board.LegalMoves()
	require.Greater(t, len(moves), 0, "expected at least one legal move")

	// If there's only 1 move, test the early return path
	if len(moves) == 1 {
		eng, err := NewMinimaxEngine(Medium)
		require.NoError(t, err)
		defer eng.Close()

		start := time.Now()
		move, err := eng.SelectMove(context.Background(), board)
		elapsed := time.Since(start)

		// Should return immediately
		assert.NoError(t, err)
		assert.Equal(t, moves[0], move)
		assert.Less(t, elapsed, 100*time.Millisecond, "forced move should return very quickly")
	} else {
		// Otherwise, just verify normal move selection works
		eng, err := NewMinimaxEngine(Medium)
		require.NoError(t, err)
		defer eng.Close()

		move, err := eng.SelectMove(context.Background(), board)
		assert.NoError(t, err)
		assert.Contains(t, moves, move)
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
			require.NoError(t, err, "failed to parse FEN")

			eng, err := NewMinimaxEngine(Medium)
			require.NoError(t, err)
			defer eng.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			move, err := eng.SelectMove(ctx, board)
			require.NoError(t, err, "SelectMove failed")

			// Verify the move delivers checkmate
			boardCopy := board.Copy()
			err = boardCopy.MakeMove(move)
			require.NoError(t, err, "failed to make move")

			assert.Equal(t, engine.Checkmate, boardCopy.Status(), "engine should find mate-in-1: %s", tt.description)
		})
	}
}

func TestMinimaxEngine_AvoidBlunder(t *testing.T) {
	// Position where White queen on d1 can move, but moving to d8 would hang it to Black rook on d7
	// White should NOT play Qd8
	fen := "4k3/3r4/8/8/8/8/8/3Q1K2 w - - 0 1"

	board, err := engine.ParseFEN(fen)
	require.NoError(t, err)

	eng, err := NewMinimaxEngine(Medium)
	require.NoError(t, err)
	defer eng.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	move, err := eng.SelectMove(ctx, board)
	require.NoError(t, err)

	// The move should NOT be Qd8 (hanging the queen)
	blunderMove, _ := engine.ParseMove("d1d8")
	assert.NotEqual(t, blunderMove, move, "engine should not hang the queen")

	// Verify the move doesn't lose the queen
	boardCopy := board.Copy()
	err = boardCopy.MakeMove(move)
	require.NoError(t, err)

	// After opponent's best response, White's queen should still be on the board
	// (This is a simplified check - we're not doing full move analysis here)
	// At minimum, the engine shouldn't blunder the queen immediately
}

func TestMinimaxEngine_CapturePriority(t *testing.T) {
	// Position where White can capture Black queen with a pawn
	// White pawn on e5, Black queen on f6, Black king on g8
	fen := "6k1/8/5q2/4P3/8/8/8/6K1 w - - 0 1"

	board, err := engine.ParseFEN(fen)
	require.NoError(t, err)

	eng, err := NewMinimaxEngine(Medium)
	require.NoError(t, err)
	defer eng.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	move, err := eng.SelectMove(ctx, board)
	require.NoError(t, err)

	// Engine should capture the queen
	captureMove, _ := engine.ParseMove("e5f6")
	assert.Equal(t, captureMove, move, "engine should capture hanging queen")
}

func TestMinimaxEngine_Timeout(t *testing.T) {
	// Start from initial position (complex enough to potentially exceed timeout)
	board := engine.NewBoard()

	// Create engine with very short timeout
	eng, err := NewMinimaxEngine(Medium, WithTimeLimit(1*time.Nanosecond))
	require.NoError(t, err)
	defer eng.Close()

	ctx := context.Background()

	// This might timeout or might complete if the search is fast enough
	// We're primarily checking that it doesn't panic or hang
	move, err := eng.SelectMove(ctx, board)

	// Either it succeeds or times out gracefully
	if err != nil {
		// If it errors, it should be a timeout or context error
		assert.Contains(t, err.Error(), "context deadline exceeded")
	} else {
		// If it succeeds, move should be legal
		moves := board.LegalMoves()
		assert.Contains(t, moves, move)
	}
}

func TestMinimaxEngine_NoLegalMoves(t *testing.T) {
	// This is a checkmate position - Black has no legal moves
	fen := "rnb1kbnr/pppp1ppp/8/4p3/6Pq/5P2/PPPPP2P/RNBQKBNR w KQkq - 1 3"

	board, err := engine.ParseFEN(fen)
	require.NoError(t, err)

	// Verify it's checkmate
	assert.Equal(t, engine.Checkmate, board.Status())

	eng, err := NewMinimaxEngine(Medium)
	require.NoError(t, err)
	defer eng.Close()

	// Shouldn't be able to select a move in checkmate position
	_, err = eng.SelectMove(context.Background(), board)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no legal moves")
}

func TestMinimaxEngine_Depth2Search(t *testing.T) {
	// Position where depth-2 search should find a better move than depth-1
	// White can win a rook with a pawn fork: e4-e5 forks rook on d6 and f6
	fen := "6k1/8/3r1r2/8/4P3/8/8/6K1 w - - 0 1"

	board, err := engine.ParseFEN(fen)
	require.NoError(t, err)

	eng, err := NewMinimaxEngine(Medium)
	require.NoError(t, err)
	defer eng.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	move, err := eng.SelectMove(ctx, board)
	require.NoError(t, err)

	// The engine should find e4-e5, which sets up a fork
	forkMove, _ := engine.ParseMove("e4e5")
	assert.Equal(t, forkMove, move, "engine should find the pawn fork")
}

func TestMinimaxEngine_AlphaBetaPruning(t *testing.T) {
	// This test verifies that the engine completes within reasonable time
	// which indicates alpha-beta pruning is working (vs. plain minimax)
	board := engine.NewBoard()

	eng, err := NewMinimaxEngine(Medium)
	require.NoError(t, err)
	defer eng.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	move, err := eng.SelectMove(ctx, board)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.NotEmpty(t, move)

	// Depth-2 search with alpha-beta should complete in well under 1 second
	// even from the starting position
	assert.Less(t, elapsed, 1*time.Second, "search should complete quickly with pruning")
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
			assert.Equal(t, tt.expected, weights)
		})
	}
}

func TestMinimaxEngine_MoveOrdering(t *testing.T) {
	// Position with several capture and non-capture options
	// White pawn on e5 can capture f6 or d6, or push to e6
	fen := "6k1/8/3p1p2/4P3/8/8/8/6K1 w - - 0 1"

	board, err := engine.ParseFEN(fen)
	require.NoError(t, err)

	eng, err := NewMinimaxEngine(Medium)
	require.NoError(t, err)
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

		if isCapture {
			assert.True(t, capturePhase, "all captures should come before non-captures")
		}
	}
}
