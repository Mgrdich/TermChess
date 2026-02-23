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
// the ONNX model has been loaded.
var ErrModelNotLoaded = errors.New("RL model not loaded: ONNX runtime integration pending")

// rlInferenceSession abstracts the ONNX inference so it can be mocked in tests
// and implemented with real ONNX Runtime.
type rlInferenceSession interface {
	// RunInference takes the encoded board (flat [1, 18, 8, 8] = 1152 floats)
	// and returns policy logits [4096] and a value in [-1, 1].
	RunInference(input []float32) (policy []float32, value float32, err error)

	// Close releases resources held by the inference session.
	Close() error
}

// rlEngine implements the Engine and Inspectable interfaces for RL-based
// chess agents. When session is nil, SelectMove returns ErrModelNotLoaded.
// The actual ONNX session will be wired in when models are embedded.
type rlEngine struct {
	name       string
	difficulty RLDifficulty
	timeLimit  time.Duration
	closed     int32              // atomic: 0 = open, 1 = closed
	session    rlInferenceSession // nil until model is loaded
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
// Returns ErrModelNotLoaded if the ONNX session has not been loaded.
func (e *rlEngine) SelectMove(ctx context.Context, board *engine.Board) (engine.Move, error) {
	if atomic.LoadInt32(&e.closed) == 1 {
		return engine.Move{}, errors.New("engine is closed")
	}

	if e.session == nil {
		return engine.Move{}, ErrModelNotLoaded
	}

	// TODO: encode board, run inference, decode policy into a legal move.
	// This will be implemented when MCTS + move selection is added.
	return engine.Move{}, ErrModelNotLoaded
}

// Name returns the human-readable name of this engine.
func (e *rlEngine) Name() string {
	return e.name
}

// Close releases resources held by the engine.
// If an inference session is loaded, it is also closed.
// Returns an error if the engine is already closed.
func (e *rlEngine) Close() error {
	if !atomic.CompareAndSwapInt32(&e.closed, 0, 1) {
		return errors.New("engine already closed")
	}

	if e.session != nil {
		return e.session.Close()
	}

	return nil
}

// newOnnxSession creates an ONNX inference session from model bytes.
// This is the bridge to the onnxruntime_go library.
// Returns an error until ONNX Runtime is configured with embedded models.
func newOnnxSession(_ []byte) (rlInferenceSession, error) {
	return nil, errors.New("ONNX Runtime not yet configured: run with embedded models")
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
