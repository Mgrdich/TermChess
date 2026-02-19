package bot

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// RLDifficulty represents difficulty levels for RL-based engines.
type RLDifficulty int

const (
	// RLIntermediate targets approximately 1500 ELO.
	RLIntermediate RLDifficulty = iota
	// RLAdvanced targets approximately 2000 ELO.
	RLAdvanced
	// RLMaster targets approximately 2200 ELO.
	RLMaster
)

// String returns a human-readable description of the RL difficulty level.
func (d RLDifficulty) String() string {
	switch d {
	case RLIntermediate:
		return "RL Intermediate (1500)"
	case RLAdvanced:
		return "RL Advanced (2000)"
	case RLMaster:
		return "RL Master (2200)"
	default:
		return "Unknown"
	}
}

// ErrModelNotLoaded is returned when SelectMove is called before
// the ONNX model has been loaded. This is a placeholder until
// ONNX runtime integration is added.
var ErrModelNotLoaded = errors.New("RL model not loaded: ONNX runtime integration pending")

// rlEngine implements the Engine and Inspectable interfaces for RL-based
// chess agents. This is a skeleton; actual ONNX model loading and inference
// will be added in a later slice.
type rlEngine struct {
	name       string
	difficulty RLDifficulty
	timeLimit  time.Duration
	closed     int32 // atomic: 0 = open, 1 = closed
}

// NewRLEngine creates an RL-based engine with the given difficulty.
// The engine is a skeleton until ONNX runtime integration is added;
// SelectMove will return ErrModelNotLoaded.
func NewRLEngine(difficulty RLDifficulty, opts ...EngineOption) (Engine, error) {
	// Validate difficulty
	var defaultTimeLimit time.Duration
	switch difficulty {
	case RLIntermediate:
		defaultTimeLimit = 5 * time.Second
	case RLAdvanced:
		defaultTimeLimit = 8 * time.Second
	case RLMaster:
		defaultTimeLimit = 10 * time.Second
	default:
		return nil, fmt.Errorf("invalid RL difficulty: %d (expected RLIntermediate, RLAdvanced, or RLMaster)", difficulty)
	}

	cfg := &engineConfig{
		timeLimit: defaultTimeLimit,
	}

	// Apply custom options
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	return &rlEngine{
		name:       difficulty.String(),
		difficulty: difficulty,
		timeLimit:  cfg.timeLimit,
		closed:     0, // atomic: 0 = open
	}, nil
}

// SelectMove returns the engine's chosen move for the given position.
// Currently returns ErrModelNotLoaded as a placeholder until ONNX
// runtime integration is added.
func (e *rlEngine) SelectMove(ctx context.Context, board *engine.Board) (engine.Move, error) {
	if atomic.LoadInt32(&e.closed) == 1 {
		return engine.Move{}, errors.New("engine is closed")
	}

	return engine.Move{}, ErrModelNotLoaded
}

// Name returns the human-readable name of this engine.
func (e *rlEngine) Name() string {
	return e.name
}

// Close releases resources held by the engine.
// Returns an error if the engine is already closed.
func (e *rlEngine) Close() error {
	if !atomic.CompareAndSwapInt32(&e.closed, 0, 1) {
		return errors.New("engine already closed")
	}
	return nil
}

// Info returns metadata about this engine.
func (e *rlEngine) Info() Info {
	return Info{
		Name:       e.name,
		Author:     "TermChess",
		Version:    "0.1.0",
		Type:       TypeRL,
		Difficulty: Hard,
		Features:   map[string]bool{"onnx": false},
	}
}
