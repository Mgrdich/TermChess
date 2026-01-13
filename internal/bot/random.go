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

// SelectMove returns a random legal move from the current position.
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

	// Select random move
	// Check context to respect timeout (though this is instant)
	select {
	case <-ctx.Done():
		return engine.Move{}, ctx.Err()
	default:
		return moves[rand.Intn(len(moves))], nil
	}
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
			"random_selection": true,
		},
	}
}
