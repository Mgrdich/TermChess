// Package bvb provides Bot vs Bot game sessions for automated chess matches.
package bvb

import (
	"time"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// PlaybackSpeed controls the delay between moves in a Bot vs Bot game.
type PlaybackSpeed int

const (
	// SpeedInstant applies no delay between moves.
	SpeedInstant PlaybackSpeed = iota
	// SpeedNormal applies a 1 second delay between moves.
	SpeedNormal
)

// Duration returns the time delay associated with this playback speed.
func (s PlaybackSpeed) Duration() time.Duration {
	switch s {
	case SpeedInstant:
		return 0
	case SpeedNormal:
		return time.Second
	default:
		return 0
	}
}

// SessionState represents the current state of a game session.
type SessionState int

const (
	// StateRunning indicates the game is currently in progress.
	StateRunning SessionState = iota
	// StatePaused indicates the game is paused.
	StatePaused
	// StateFinished indicates the game has completed.
	StateFinished
)

// GameResult holds the outcome of a completed Bot vs Bot game.
type GameResult struct {
	// GameNumber is the sequence number of this game in a series.
	GameNumber int
	// Winner is the name of the winning engine, or "Draw" for draws.
	Winner string
	// WinnerColor is the color of the winning engine.
	WinnerColor engine.Color
	// EndReason describes why the game ended (e.g., "checkmate", "stalemate").
	EndReason string
	// MoveCount is the total number of moves played.
	MoveCount int
	// Duration is how long the game took to complete.
	Duration time.Duration
	// FinalFEN is the FEN string of the final board position.
	FinalFEN string
	// MoveHistory contains all moves played in order.
	MoveHistory []engine.Move
}
