package bot

import (
	"fmt"
	"math/rand"
	"time"
)

// EngineOption is a functional option for engine creation.
type EngineOption func(*engineConfig) error

// engineConfig holds configuration options for engine creation.
type engineConfig struct {
	difficulty  Difficulty
	timeLimit   time.Duration
	searchDepth int
	options     map[string]any
}

// WithTimeLimit sets a custom time limit for move selection.
func WithTimeLimit(d time.Duration) EngineOption {
	return func(c *engineConfig) error {
		if d <= 0 {
			return fmt.Errorf("time limit must be positive")
		}
		c.timeLimit = d
		return nil
	}
}

// WithSearchDepth sets a custom search depth for minimax engines.
func WithSearchDepth(depth int) EngineOption {
	return func(c *engineConfig) error {
		if depth < 1 || depth > 20 {
			return fmt.Errorf("search depth must be 1-20")
		}
		c.searchDepth = depth
		return nil
	}
}

// WithOptions sets custom options as a map.
func WithOptions(opts map[string]any) EngineOption {
	return func(c *engineConfig) error {
		c.options = opts
		return nil
	}
}

// NewRandomEngine creates an Easy bot with random move selection.
func NewRandomEngine(opts ...EngineOption) (Engine, error) {
	cfg := &engineConfig{
		difficulty: Easy,
		timeLimit:  2 * time.Second,
	}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	return &randomEngine{
		name:      "Easy Bot",
		timeLimit: cfg.timeLimit,
		closed:    0, // atomic: 0 = open
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// NewMinimaxEngine creates a Medium or Hard bot using minimax with alpha-beta pruning.
func NewMinimaxEngine(difficulty Difficulty, opts ...EngineOption) (Engine, error) {
	cfg := &engineConfig{difficulty: difficulty}

	// Set defaults based on difficulty
	switch difficulty {
	case Medium:
		cfg.timeLimit = 4 * time.Second
		cfg.searchDepth = 4
	case Hard:
		cfg.timeLimit = 8 * time.Second
		cfg.searchDepth = 6
	default:
		return nil, fmt.Errorf("invalid difficulty for minimax: %d (expected Medium or Hard)", difficulty)
	}

	// Apply custom options
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	// Create the minimax engine
	name := fmt.Sprintf("%s Bot", difficulty.String())

	return &minimaxEngine{
		name:        name,
		difficulty:  cfg.difficulty,
		maxDepth:    cfg.searchDepth,
		timeLimit:   cfg.timeLimit,
		evalWeights: getDefaultWeights(cfg.difficulty),
		closed:      false,
	}, nil
}
