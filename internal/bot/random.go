package bot

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// randomEngine implements the Easy bot using random move selection.
type randomEngine struct {
	name      string
	timeLimit time.Duration
	closed    bool
}

// SelectMove returns a move using weighted selection (70% tactical bias).
func (e *randomEngine) SelectMove(ctx context.Context, board *engine.Board) (engine.Move, error) {
	if e.closed {
		return engine.Move{}, errors.New("engine is closed")
	}

	// Get all legal moves
	moves := board.LegalMoves()
	if len(moves) == 0 {
		return engine.Move{}, errors.New("no legal moves available")
	}

	// If only one move, return it immediately (forced move)
	if len(moves) == 1 {
		return moves[0], nil
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, e.timeLimit)
	defer cancel()

	// Categorize moves
	captures := filterCaptures(board, moves)
	checks := filterChecks(board, moves)

	// Weighted selection: 70% tactical bias
	// Check context to respect timeout
	select {
	case <-ctx.Done():
		return engine.Move{}, ctx.Err()
	default:
		// 70% chance to pick a capture if available
		if rand.Float64() < 0.7 && len(captures) > 0 {
			return captures[rand.Intn(len(captures))], nil
		}

		// 50% chance to pick a check if available
		if rand.Float64() < 0.5 && len(checks) > 0 {
			return checks[rand.Intn(len(checks))], nil
		}

		// Fallback: any random legal move
		return moves[rand.Intn(len(moves))], nil
	}
}

// filterCaptures returns all moves that capture an opponent's piece.
func filterCaptures(board *engine.Board, moves []engine.Move) []engine.Move {
	var captures []engine.Move
	for _, m := range moves {
		// Check if destination square has an opponent piece
		targetPiece := board.PieceAt(m.To)
		if !targetPiece.IsEmpty() {
			captures = append(captures, m)
		}
	}
	return captures
}

// filterChecks returns all moves that give check to the opponent's king.
func filterChecks(board *engine.Board, moves []engine.Move) []engine.Move {
	var checks []engine.Move
	for _, m := range moves {
		// Make move on a copy and check if opponent is in check
		boardCopy := board.Copy()
		boardCopy.MakeMove(m)

		// After making the move, check if the NEW active color (opponent) is in check
		if boardCopy.InCheck() {
			checks = append(checks, m)
		}
	}
	return checks
}

// Name returns the human-readable name of this engine.
func (e *randomEngine) Name() string {
	return e.name
}

// Close releases resources held by the engine.
func (e *randomEngine) Close() error {
	e.closed = true
	return nil
}

// Info returns metadata about this engine.
func (e *randomEngine) Info() Info {
	return Info{
		Name:       e.name,
		Author:     "TermChess",
		Version:    "1.0",
		Type:       TypeInternal,
		Difficulty: Easy,
		Features: map[string]bool{
			"random_selection":   true,
			"tactical_awareness": true,
			"weighted_selection": true,
		},
	}
}
