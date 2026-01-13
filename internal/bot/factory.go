package bot

import (
	"fmt"
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

// NewRandomEngine creates an Easy bot with weighted random selection.
// This is a PLACEHOLDER that returns nil until Task 3 implements the actual bot.
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

	// TODO: Task 3 will implement the actual randomEngine
	return nil, fmt.Errorf("not implemented: random engine will be created in Task 3")
}

// NewMinimaxEngine creates a Medium or Hard bot using minimax with alpha-beta pruning.
// This is a PLACEHOLDER that returns nil until Task 6 implements the actual bot.
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

	// TODO: Task 6 will implement the actual minimaxEngine
	return nil, fmt.Errorf("not implemented: minimax engine will be created in Task 6")
}
