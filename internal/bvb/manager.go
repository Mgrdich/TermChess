package bvb

import (
	"sync"

	"github.com/Mgrdich/TermChess/internal/bot"
)

// SessionManager orchestrates N parallel game sessions.
type SessionManager struct {
	mu        sync.Mutex
	sessions  []*GameSession
	state     SessionState
	speed     PlaybackSpeed
	whiteDiff bot.Difficulty
	blackDiff bot.Difficulty
	whiteName string
	blackName string
	gameCount int
}

// NewSessionManager creates a new manager configured for the given matchup.
func NewSessionManager(whiteDiff, blackDiff bot.Difficulty, whiteName, blackName string, gameCount int) *SessionManager {
	return &SessionManager{
		state:     StateRunning,
		speed:     SpeedNormal,
		whiteDiff: whiteDiff,
		blackDiff: blackDiff,
		whiteName: whiteName,
		blackName: blackName,
		gameCount: gameCount,
	}
}

// Start creates engine instances for each game and launches all sessions as goroutines.
func (m *SessionManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions = make([]*GameSession, m.gameCount)

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
		go session.Run()
	}

	return nil
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
	m.abortSessions()
}

// abortSessions stops all non-nil sessions. Must be called with m.mu held.
func (m *SessionManager) abortSessions() {
	for _, s := range m.sessions {
		if s != nil && !s.IsFinished() {
			s.Abort()
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
