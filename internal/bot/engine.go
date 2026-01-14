// Package bot provides interfaces and types for chess bot opponents.
package bot

import (
	"context"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// Engine represents a chess bot that can select moves.
// This is the minimal interface all engines must implement.
type Engine interface {
	// SelectMove returns the bot's chosen move for the given position.
	// The context allows cancellation if the bot exceeds time limits.
	SelectMove(ctx context.Context, board *engine.Board) (engine.Move, error)

	// Name returns a human-readable name for this engine.
	Name() string

	// Close releases any resources held by the engine.
	// Implementations should be idempotent (safe to call multiple times).
	// Internal bots can no-op; UCI engines kill processes; RL agents free model memory.
	Close() error
}

// Configurable engines can accept configuration before or during use.
// Internal bots implement this for difficulty tuning.
// UCI engines implement this for engine options (Threads, Hash, etc.).
type Configurable interface {
	Engine
	Configure(options map[string]any) error
}

// Stateful engines benefit from knowing position history.
// UCI engines use this for opening books and transposition tables.
// RL agents might use this for sequential context.
type Stateful interface {
	Engine
	SetPositionHistory(history []*engine.Board) error
}

// Info provides metadata about the engine.
type Info struct {
	Name       string          // Human-readable name
	Author     string          // Engine author
	Version    string          // Engine version
	Type       EngineType      // Internal, UCI, or RL
	Difficulty Difficulty      // Easy, Medium, Hard (for internal bots)
	Features   map[string]bool // Supported features
}

// Inspectable engines can report metadata.
// Useful for UI display and debugging.
type Inspectable interface {
	Engine
	Info() Info
}

// EngineType categorizes engine implementations.
type EngineType int

const (
	// TypeInternal represents built-in Go implementations.
	TypeInternal EngineType = iota
	// TypeUCI represents external UCI engines (Phase 5).
	TypeUCI
	// TypeRL represents RL agents with ONNX models (Phase 6).
	TypeRL
)

// String returns a string representation of the engine type.
func (t EngineType) String() string {
	switch t {
	case TypeInternal:
		return "Internal"
	case TypeUCI:
		return "UCI"
	case TypeRL:
		return "RL"
	default:
		return "Unknown"
	}
}

// Difficulty levels for internal engines.
type Difficulty int

const (
	// Easy difficulty: fast responses, simpler evaluation.
	Easy Difficulty = iota
	// Medium difficulty: balanced play.
	Medium
	// Hard difficulty: stronger evaluation, deeper search.
	Hard
)

// String returns a string representation of the difficulty level.
func (d Difficulty) String() string {
	switch d {
	case Easy:
		return "Easy"
	case Medium:
		return "Medium"
	case Hard:
		return "Hard"
	default:
		return "Unknown"
	}
}
