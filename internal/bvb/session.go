package bvb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Mgrdich/TermChess/internal/bot"
	"github.com/Mgrdich/TermChess/internal/engine"
)

// maxMoveCount is the maximum number of moves before a forced draw.
const maxMoveCount = 500

// GameSession manages a single Bot vs Bot chess game.
// It runs the game loop in a goroutine and provides thread-safe
// access to the current board state and move history.
type GameSession struct {
	mu          sync.Mutex
	gameNumber  int
	board       *engine.Board
	whiteEngine bot.Engine
	blackEngine bot.Engine
	whiteName   string
	blackName   string
	moveHistory []engine.Move
	state       SessionState
	paused      bool
	result      *GameResult
	startTime   time.Time
	speed       *PlaybackSpeed
	stopCh      chan struct{}
	pauseCh     chan struct{}
	resumeCh    chan struct{}
}

// NewGameSession creates a new game session ready to be run.
// The speed parameter is a pointer to a shared PlaybackSpeed value
// that can be modified externally to change the delay between moves.
func NewGameSession(gameNumber int, whiteEngine bot.Engine, blackEngine bot.Engine, whiteName string, blackName string, speed *PlaybackSpeed) *GameSession {
	return &GameSession{
		gameNumber:  gameNumber,
		board:       engine.NewBoard(),
		whiteEngine: whiteEngine,
		blackEngine: blackEngine,
		whiteName:   whiteName,
		blackName:   blackName,
		moveHistory: make([]engine.Move, 0, 80),
		state:       StateRunning,
		speed:       speed,
		stopCh:      make(chan struct{}),
		pauseCh:     make(chan struct{}, 1),
		resumeCh:    make(chan struct{}, 1),
	}
}

// Run executes the game loop. This is intended to be called as a goroutine.
// It plays moves alternately until the game ends, an error occurs, or
// the session is stopped via the stop channel.
func (s *GameSession) Run() {
	s.mu.Lock()
	s.startTime = time.Now()
	s.mu.Unlock()

	defer s.closeEngines()

	for {
		// Check for abort signal.
		select {
		case <-s.stopCh:
			s.mu.Lock()
			s.state = StateFinished
			s.mu.Unlock()
			return
		default:
		}

		// Check for pause signal.
		select {
		case <-s.pauseCh:
			// Wait for resume or stop.
			select {
			case <-s.resumeCh:
				// Continue.
			case <-s.stopCh:
				s.mu.Lock()
				s.state = StateFinished
				s.mu.Unlock()
				return
			}
		case <-s.stopCh:
			s.mu.Lock()
			s.state = StateFinished
			s.mu.Unlock()
			return
		default:
			// Not paused, continue.
		}

		// Determine the current engine based on active color.
		s.mu.Lock()
		activeColor := s.board.ActiveColor
		var currentEngine bot.Engine
		var currentName string
		if activeColor == engine.White {
			currentEngine = s.whiteEngine
			currentName = s.whiteName
		} else {
			currentEngine = s.blackEngine
			currentName = s.blackName
		}
		boardCopy := s.board.Copy()
		s.mu.Unlock()

		// Ask the engine to select a move.
		move, err := currentEngine.SelectMove(context.Background(), boardCopy)
		if err != nil {
			s.finishWithError(currentName, activeColor, err)
			return
		}

		// Apply the move to the board.
		s.mu.Lock()
		if err := s.board.MakeMove(move); err != nil {
			s.mu.Unlock()
			s.finishWithError(currentName, activeColor, err)
			return
		}
		s.moveHistory = append(s.moveHistory, move)
		moveCount := len(s.moveHistory)

		// Check for game over conditions.
		status := s.board.Status()
		if status != engine.Ongoing {
			s.finishWithStatus(status, moveCount)
			s.mu.Unlock()
			return
		}

		// Check for forced draw due to excessive moves.
		if moveCount >= maxMoveCount {
			s.result = &GameResult{
				GameNumber:  s.gameNumber,
				Winner:      "Draw",
				EndReason:   "move limit exceeded",
				MoveCount:   moveCount,
				Duration:    time.Since(s.startTime),
				FinalFEN:    s.board.ToFEN(),
				MoveHistory: s.copyMoveHistory(),
			}
			s.state = StateFinished
			s.mu.Unlock()
			return
		}
		s.mu.Unlock()

		// Sleep for the configured playback speed, interruptible by stop signal.
		s.mu.Lock()
		delay := s.speed.Duration()
		s.mu.Unlock()

		if delay > 0 {
			select {
			case <-time.After(delay):
			case <-s.stopCh:
				s.mu.Lock()
				s.state = StateFinished
				s.mu.Unlock()
				return
			}
		}
	}
}

// Pause signals the game session to pause. It is safe to call multiple times.
// If the session is already paused or finished, this is a no-op.
func (s *GameSession) Pause() {
	s.mu.Lock()
	if s.paused || s.state == StateFinished {
		s.mu.Unlock()
		return
	}
	s.paused = true
	s.state = StatePaused
	s.mu.Unlock()

	// Non-blocking send on buffered channel.
	select {
	case s.pauseCh <- struct{}{}:
	default:
	}
}

// Resume signals the game session to continue after a pause.
// If the session is not paused, this is a no-op.
func (s *GameSession) Resume() {
	s.mu.Lock()
	if !s.paused {
		s.mu.Unlock()
		return
	}
	s.paused = false
	s.state = StateRunning
	s.mu.Unlock()

	// Non-blocking send on buffered channel.
	select {
	case s.resumeCh <- struct{}{}:
	default:
	}
}

// Abort signals the game session to stop immediately. It is safe to call multiple times.
func (s *GameSession) Abort() {
	select {
	case <-s.stopCh:
		// Already closed.
	default:
		close(s.stopCh)
	}
}

// CurrentBoard returns a deep copy of the current board state.
func (s *GameSession) CurrentBoard() *engine.Board {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.board.Copy()
}

// CurrentMoveHistory returns a copy of the move history so far.
func (s *GameSession) CurrentMoveHistory() []engine.Move {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.copyMoveHistory()
}

// IsFinished returns true if the game session has completed.
func (s *GameSession) IsFinished() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state == StateFinished
}

// Result returns the game result, or nil if the game is not finished.
func (s *GameSession) Result() *GameResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.result
}

// GameNumber returns the sequence number of this game.
func (s *GameSession) GameNumber() int {
	return s.gameNumber
}

// State returns the current session state.
func (s *GameSession) State() SessionState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

// finishWithStatus records the game result based on the board's game status.
// Must be called with s.mu held.
func (s *GameSession) finishWithStatus(status engine.GameStatus, moveCount int) {
	winner := "Draw"
	var winnerColor engine.Color

	if status == engine.Checkmate {
		// The active color is the one checkmated, so the opponent wins.
		if s.board.ActiveColor == engine.White {
			winner = s.blackName
			winnerColor = engine.Black
		} else {
			winner = s.whiteName
			winnerColor = engine.White
		}
	}

	s.result = &GameResult{
		GameNumber:  s.gameNumber,
		Winner:      winner,
		WinnerColor: winnerColor,
		EndReason:   status.String(),
		MoveCount:   moveCount,
		Duration:    time.Since(s.startTime),
		FinalFEN:    s.board.ToFEN(),
		MoveHistory: s.copyMoveHistory(),
	}
	s.state = StateFinished
}

// finishWithError records the game result when an engine produces an error.
func (s *GameSession) finishWithError(engineName string, engineColor engine.Color, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// The engine that errored loses; the opponent wins.
	var winner string
	var winnerColor engine.Color
	if engineColor == engine.White {
		winner = s.blackName
		winnerColor = engine.Black
	} else {
		winner = s.whiteName
		winnerColor = engine.White
	}

	s.result = &GameResult{
		GameNumber:  s.gameNumber,
		Winner:      winner,
		WinnerColor: winnerColor,
		EndReason:   fmt.Sprintf("engine error: %v", err),
		MoveCount:   len(s.moveHistory),
		Duration:    time.Since(s.startTime),
		FinalFEN:    s.board.ToFEN(),
		MoveHistory: s.copyMoveHistory(),
	}
	s.state = StateFinished
}

// copyMoveHistory returns a copy of the move history slice.
// Must be called with s.mu held.
func (s *GameSession) copyMoveHistory() []engine.Move {
	moves := make([]engine.Move, len(s.moveHistory))
	copy(moves, s.moveHistory)
	return moves
}

// closeEngines closes both bot engines, ignoring errors.
func (s *GameSession) closeEngines() {
	_ = s.whiteEngine.Close()
	_ = s.blackEngine.Close()
}
