package bvb

import (
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/Mgrdich/TermChess/internal/bot"
)

// maxConcurrentGames limits how many games run simultaneously.
// This prevents excessive CPU usage when running many games.
const maxConcurrentGames = 50

// MaxConcurrentGames returns the maximum number of concurrent games.
// Exported for UI display purposes.
func MaxConcurrentGames() int {
	return maxConcurrentGames
}

// CalculateDefaultConcurrency returns the recommended concurrency based on CPU count.
// It uses a tiered formula:
//   - numCPU <= 2: use numCPU
//   - numCPU <= 4: use numCPU * 1.5
//   - numCPU > 4: use numCPU * 2
//
// The result is capped at maxConcurrentGames and has a minimum of 1.
func CalculateDefaultConcurrency() int {
	return calculateDefaultConcurrencyWithCPU(runtime.NumCPU())
}

// calculateDefaultConcurrencyWithCPU is the internal implementation that accepts
// the CPU count as a parameter for testing purposes.
func calculateDefaultConcurrencyWithCPU(numCPU int) int {
	var concurrency int
	switch {
	case numCPU <= 2:
		concurrency = numCPU
	case numCPU <= 4:
		concurrency = int(float64(numCPU) * 1.5)
	default:
		concurrency = numCPU * 2
	}

	// Cap at reasonable maximum
	if concurrency > maxConcurrentGames {
		concurrency = maxConcurrentGames
	}
	if concurrency < 1 {
		concurrency = 1
	}
	return concurrency
}

// SessionManager orchestrates N parallel game sessions.
type SessionManager struct {
	mu          sync.Mutex
	sessions    []*GameSession
	state       SessionState
	speed       PlaybackSpeed
	whiteDiff   bot.Difficulty
	blackDiff   bot.Difficulty
	whiteName   string
	blackName   string
	gameCount   int
	concurrency int           // effective concurrency (auto-detected or user-specified)
	semaphore   chan struct{} // limits concurrent game execution
	abortCh     chan struct{} // signals all waiting goroutines to abort
	activeCount int32         // atomic counter for currently running games
}

// NewSessionManager creates a new manager configured for the given matchup.
// The concurrency parameter controls how many games run in parallel.
// If concurrency is 0, it auto-detects based on CPU count.
// If concurrency exceeds maxConcurrentGames, it is capped.
func NewSessionManager(whiteDiff, blackDiff bot.Difficulty, whiteName, blackName string, gameCount, concurrency int) *SessionManager {
	// Auto-detect if concurrency is 0
	effectiveConcurrency := concurrency
	if effectiveConcurrency == 0 {
		effectiveConcurrency = CalculateDefaultConcurrency()
	}

	// Cap at maxConcurrentGames
	if effectiveConcurrency > maxConcurrentGames {
		effectiveConcurrency = maxConcurrentGames
	}

	// Ensure at least 1
	if effectiveConcurrency < 1 {
		effectiveConcurrency = 1
	}

	return &SessionManager{
		state:       StateRunning,
		speed:       SpeedNormal,
		whiteDiff:   whiteDiff,
		blackDiff:   blackDiff,
		whiteName:   whiteName,
		blackName:   blackName,
		gameCount:   gameCount,
		concurrency: effectiveConcurrency,
	}
}

// Start creates engine instances for each game and launches them via a coordinator.
// Games are started in order (1, 2, 3, ...) with up to concurrency running at once.
func (m *SessionManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions = make([]*GameSession, m.gameCount)

	// Create semaphore to limit concurrent games
	// Use the smaller of concurrency and gameCount
	semaphoreSize := m.concurrency
	if m.gameCount < semaphoreSize {
		semaphoreSize = m.gameCount
	}
	m.semaphore = make(chan struct{}, semaphoreSize)
	m.abortCh = make(chan struct{})

	// Pre-create all sessions and their engines
	for i := 0; i < m.gameCount; i++ {
		whiteEngine, err := createEngine(m.whiteDiff)
		if err != nil {
			m.abortSessions()
			return err
		}
		blackEngine, err := createEngine(m.blackDiff)
		if err != nil {
			whiteEngine.Close()
			m.abortSessions()
			return err
		}

		sessionSpeed := new(PlaybackSpeed)
		*sessionSpeed = m.speed
		session := NewGameSession(i+1, whiteEngine, blackEngine, m.whiteName, m.blackName, sessionSpeed)
		m.sessions[i] = session
	}

	// Launch coordinator goroutine that starts games in order
	go m.coordinateGames()

	return nil
}

// coordinateGames starts games sequentially as semaphore slots become available.
// This ensures games start in order: 1-25 first, then 26, 27, etc.
func (m *SessionManager) coordinateGames() {
	for i := 0; i < m.gameCount; i++ {
		// Wait for a semaphore slot or abort signal
		select {
		case m.semaphore <- struct{}{}: // acquired slot
			// Start game i
			atomic.AddInt32(&m.activeCount, 1)
			go func(idx int) {
				defer func() {
					atomic.AddInt32(&m.activeCount, -1)
					<-m.semaphore // release slot when done
				}()
				m.sessions[idx].Run()
			}(i)
		case <-m.abortCh:
			// Aborted, stop starting new games
			return
		}
	}
}

// createEngine creates a bot engine based on difficulty.
func createEngine(diff bot.Difficulty) (bot.Engine, error) {
	switch diff {
	case bot.Easy:
		return bot.NewRandomEngine()
	case bot.Medium:
		return bot.NewMinimaxEngine(bot.Medium)
	case bot.Hard:
		return bot.NewMinimaxEngine(bot.Hard)
	default:
		return bot.NewRandomEngine()
	}
}

// Pause pauses all running sessions.
func (m *SessionManager) Pause() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = StatePaused
	for _, s := range m.sessions {
		if s != nil && !s.IsFinished() {
			s.Pause()
		}
	}
}

// Resume resumes all paused sessions.
func (m *SessionManager) Resume() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = StateRunning
	for _, s := range m.sessions {
		if s != nil && s.State() == StatePaused {
			s.Resume()
		}
	}
}

// SetSpeed updates the playback speed for all sessions.
func (m *SessionManager) SetSpeed(speed PlaybackSpeed) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.speed = speed
	for _, s := range m.sessions {
		if s != nil {
			s.SetSpeed(speed)
		}
	}
}

// Abort stops all sessions and cleans up.
func (m *SessionManager) Abort() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = StateFinished
	m.closeAbortChannel()
	m.abortSessions()
}

// Stop stops the session manager and cleans up all sessions and their resources.
// This is the preferred method for graceful shutdown as it ensures all engines
// are properly closed and resources are freed. It signals all goroutines to exit
// via the abort channel and then cleans up each session.
func (m *SessionManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.state = StateFinished

	// Signal all waiting goroutines to stop
	m.closeAbortChannel()

	// Abort all running sessions
	m.abortSessions()

	// Cleanup all sessions (call cleanup on each to ensure engines are properly closed)
	m.cleanupAllSessions()

	// Nil out the sessions slice to allow garbage collection
	m.sessions = nil
}

// closeAbortChannel safely closes the abort channel if not already closed.
// Must be called with m.mu held.
func (m *SessionManager) closeAbortChannel() {
	if m.abortCh != nil {
		select {
		case <-m.abortCh:
			// Already closed
		default:
			close(m.abortCh)
		}
	}
}

// abortSessions stops all non-nil sessions. Must be called with m.mu held.
func (m *SessionManager) abortSessions() {
	for _, s := range m.sessions {
		if s != nil && !s.IsFinished() {
			s.Abort()
		}
	}
}

// cleanupAllSessions calls cleanup on all sessions to ensure engines are closed.
// Must be called with m.mu held.
func (m *SessionManager) cleanupAllSessions() {
	for _, s := range m.sessions {
		if s != nil {
			s.cleanup()
		}
	}
}

// Sessions returns the list of game sessions.
func (m *SessionManager) Sessions() []*GameSession {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]*GameSession, len(m.sessions))
	copy(result, m.sessions)
	return result
}

// AllFinished returns true if all sessions have completed.
func (m *SessionManager) AllFinished() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.sessions) == 0 {
		return false
	}
	for _, s := range m.sessions {
		if s == nil || !s.IsFinished() {
			return false
		}
	}
	return true
}

// State returns the current manager state.
func (m *SessionManager) State() SessionState {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}

// Speed returns the current playback speed.
func (m *SessionManager) Speed() PlaybackSpeed {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.speed
}

// Concurrency returns the effective concurrency setting.
// This is the number of games that can run in parallel.
func (m *SessionManager) Concurrency() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.concurrency
}

// RunningCount returns the number of games currently executing.
func (m *SessionManager) RunningCount() int {
	return int(atomic.LoadInt32(&m.activeCount))
}

// QueuedCount returns the number of games waiting to start.
func (m *SessionManager) QueuedCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	finished := 0
	for _, s := range m.sessions {
		if s != nil && s.IsFinished() {
			finished++
		}
	}
	running := int(atomic.LoadInt32(&m.activeCount))
	queued := m.gameCount - finished - running
	if queued < 0 {
		queued = 0
	}
	return queued
}

// Stats computes aggregate statistics from all finished sessions.
func (m *SessionManager) Stats() *AggregateStats {
	m.mu.Lock()
	defer m.mu.Unlock()

	var results []GameResult
	for _, s := range m.sessions {
		if s != nil && s.IsFinished() {
			if r := s.Result(); r != nil {
				results = append(results, *r)
			}
		}
	}

	return ComputeStats(results, m.whiteName, m.blackName)
}
