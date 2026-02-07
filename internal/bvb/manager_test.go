package bvb

import (
	"testing"
	"time"

	"github.com/Mgrdich/TermChess/internal/bot"
)

func TestNewSessionManager(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "Easy Bot", "Easy Bot", 3, 0)
	if m == nil {
		t.Fatal("NewSessionManager returned nil")
	}
	if m.gameCount != 3 {
		t.Errorf("gameCount = %d, want 3", m.gameCount)
	}
	if m.State() != StateRunning {
		t.Errorf("initial state = %v, want StateRunning", m.State())
	}
}

func TestSessionManagerStartLaunchesSessions(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 3, 0)
	err := m.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer m.Abort()

	sessions := m.Sessions()
	if len(sessions) != 3 {
		t.Fatalf("len(sessions) = %d, want 3", len(sessions))
	}
	for i, s := range sessions {
		if s == nil {
			t.Errorf("session[%d] is nil", i)
		}
	}
}

func TestSessionManagerAllComplete(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 3, 0)
	m.speed = SpeedInstant
	err := m.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	// Wait for all to finish.
	deadline := time.After(60 * time.Second)
	for !m.AllFinished() {
		select {
		case <-deadline:
			m.Abort()
			t.Fatal("games did not complete within timeout")
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}

	if !m.AllFinished() {
		t.Error("AllFinished should be true")
	}

	// Verify all sessions have results.
	for i, s := range m.Sessions() {
		if !s.IsFinished() {
			t.Errorf("session[%d] not finished", i)
		}
		if s.Result() == nil {
			t.Errorf("session[%d] has nil result", i)
		}
	}
}

func TestSessionManagerPauseResume(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 2, 0)
	m.speed = SpeedNormal
	err := m.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer m.Abort()

	time.Sleep(100 * time.Millisecond)
	m.Pause()

	if m.State() != StatePaused {
		t.Errorf("state after Pause = %v, want StatePaused", m.State())
	}

	// Check sessions are paused.
	time.Sleep(100 * time.Millisecond)
	for _, s := range m.Sessions() {
		if s.State() != StatePaused && !s.IsFinished() {
			t.Errorf("session state = %v, want StatePaused", s.State())
		}
	}

	m.Resume()
	if m.State() != StateRunning {
		t.Errorf("state after Resume = %v, want StateRunning", m.State())
	}
}

func TestSessionManagerSetSpeed(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 2, 0)
	m.speed = SpeedNormal
	err := m.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer m.Abort()

	m.SetSpeed(SpeedInstant)

	if m.Speed() != SpeedInstant {
		t.Errorf("Speed() = %v, want SpeedInstant", m.Speed())
	}
}

func TestSessionManagerAbort(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 3, 0)
	m.speed = SpeedNormal
	err := m.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	m.Abort()

	// Wait briefly for goroutines to clean up.
	time.Sleep(200 * time.Millisecond)

	if m.State() != StateFinished {
		t.Errorf("state after Abort = %v, want StateFinished", m.State())
	}

	for i, s := range m.Sessions() {
		if !s.IsFinished() {
			t.Errorf("session[%d] not finished after abort", i)
		}
	}
}

func TestSessionManagerAllFinishedFalseBeforeComplete(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 3, 0)
	// Not started yet.
	if m.AllFinished() {
		t.Error("AllFinished should be false before Start")
	}

	m.speed = SpeedNormal
	err := m.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer m.Abort()

	// Games should still be running (normal speed).
	time.Sleep(100 * time.Millisecond)
	if m.AllFinished() {
		t.Error("AllFinished should be false while games are running")
	}
}

// TestCalculateDefaultConcurrencyWithCPU tests the tiered concurrency formula.
func TestCalculateDefaultConcurrencyWithCPU(t *testing.T) {
	tests := []struct {
		name     string
		numCPU   int
		expected int
	}{
		// Tier 1: numCPU <= 2, use numCPU directly
		{"1 CPU", 1, 1},
		{"2 CPUs", 2, 2},

		// Tier 2: numCPU <= 4, use numCPU * 1.5
		{"3 CPUs", 3, 4},  // 3 * 1.5 = 4.5 -> 4
		{"4 CPUs", 4, 6},  // 4 * 1.5 = 6

		// Tier 3: numCPU > 4, use numCPU * 2
		{"5 CPUs", 5, 10},   // 5 * 2 = 10
		{"8 CPUs", 8, 16},   // 8 * 2 = 16
		{"16 CPUs", 16, 32}, // 16 * 2 = 32

		// Test max cap (maxConcurrentGames = 50)
		{"30 CPUs", 30, 50}, // 30 * 2 = 60, capped at 50

		// Edge case: 0 CPUs should return 1
		{"0 CPUs", 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateDefaultConcurrencyWithCPU(tt.numCPU)
			if got != tt.expected {
				t.Errorf("calculateDefaultConcurrencyWithCPU(%d) = %d, want %d",
					tt.numCPU, got, tt.expected)
			}
		})
	}
}

// TestCalculateDefaultConcurrency verifies the exported function returns a reasonable value.
func TestCalculateDefaultConcurrency(t *testing.T) {
	concurrency := CalculateDefaultConcurrency()

	// Should be at least 1
	if concurrency < 1 {
		t.Errorf("CalculateDefaultConcurrency() = %d, want >= 1", concurrency)
	}

	// Should be at most maxConcurrentGames
	if concurrency > maxConcurrentGames {
		t.Errorf("CalculateDefaultConcurrency() = %d, want <= %d",
			concurrency, maxConcurrentGames)
	}
}

// TestNewSessionManagerAutoDetectConcurrency verifies that concurrency 0 triggers auto-detection.
func TestNewSessionManagerAutoDetectConcurrency(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 10, 0)

	// Concurrency should be auto-detected (not 0)
	if m.Concurrency() == 0 {
		t.Error("Concurrency should be auto-detected when passed 0, but got 0")
	}

	// Should be at least 1
	if m.Concurrency() < 1 {
		t.Errorf("Concurrency() = %d, want >= 1", m.Concurrency())
	}

	// Should be at most maxConcurrentGames
	if m.Concurrency() > maxConcurrentGames {
		t.Errorf("Concurrency() = %d, want <= %d", m.Concurrency(), maxConcurrentGames)
	}
}

// TestNewSessionManagerExplicitConcurrency verifies that explicit concurrency values are used.
func TestNewSessionManagerExplicitConcurrency(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 10, 5)

	if m.Concurrency() != 5 {
		t.Errorf("Concurrency() = %d, want 5", m.Concurrency())
	}
}

// TestNewSessionManagerConcurrencyNoCap verifies that explicit concurrency values are NOT capped.
// Users can specify any value and accept responsibility for potential lag.
func TestNewSessionManagerConcurrencyNoCap(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 10, 100)

	// When user explicitly provides concurrency, it should NOT be capped
	if m.Concurrency() != 100 {
		t.Errorf("Concurrency() = %d, want 100 (no cap for explicit values)", m.Concurrency())
	}
}

// TestNewSessionManagerConcurrencyMinimum verifies that concurrency has a minimum of 1.
func TestNewSessionManagerConcurrencyMinimum(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 10, -5)

	if m.Concurrency() != 1 {
		t.Errorf("Concurrency() = %d, want 1 (minimum)", m.Concurrency())
	}
}

// TestSessionManagerStop verifies that Stop() properly cleans up all sessions.
func TestSessionManagerStop(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 3, 0)
	m.speed = SpeedNormal
	err := m.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	m.Stop()

	// Wait briefly for goroutines to clean up.
	time.Sleep(200 * time.Millisecond)

	if m.State() != StateFinished {
		t.Errorf("state after Stop = %v, want StateFinished", m.State())
	}

	// Sessions should be nil'd after Stop()
	m.mu.Lock()
	sessionsNil := m.sessions == nil
	m.mu.Unlock()

	if !sessionsNil {
		t.Error("sessions should be nil after Stop()")
	}
}

// TestSessionManagerStopCleansUpEngines verifies that Stop() cleans up engines.
func TestSessionManagerStopCleansUpEngines(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 2, 0)
	m.speed = SpeedInstant
	err := m.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	// Get a copy of sessions before Stop
	sessions := m.Sessions()

	// Wait for games to finish
	deadline := time.After(60 * time.Second)
	for !m.AllFinished() {
		select {
		case <-deadline:
			m.Stop()
			t.Fatal("games did not complete within timeout")
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}

	// Now call Stop() to ensure cleanup happens
	m.Stop()

	// Verify engines in the original sessions are nil'd
	for i, s := range sessions {
		s.mu.Lock()
		whiteEngineNil := s.whiteEngine == nil
		blackEngineNil := s.blackEngine == nil
		s.mu.Unlock()

		if !whiteEngineNil {
			t.Errorf("session[%d] whiteEngine should be nil after Stop()", i)
		}
		if !blackEngineNil {
			t.Errorf("session[%d] blackEngine should be nil after Stop()", i)
		}
	}
}

// TestSessionManagerStopIsIdempotent verifies that Stop() can be called multiple times.
func TestSessionManagerStopIsIdempotent(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 2, 0)
	m.speed = SpeedNormal
	err := m.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Call Stop multiple times - should not panic
	m.Stop()
	m.Stop()
	m.Stop()

	if m.State() != StateFinished {
		t.Errorf("state after multiple Stop calls = %v, want StateFinished", m.State())
	}
}

// TestSessionManagerGetSession verifies GetSession returns correct sessions.
func TestSessionManagerGetSession(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 3, 0)
	m.speed = SpeedInstant // Use instant speed so games complete quickly
	err := m.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	// Wait for all games to finish before testing and stopping
	deadline := time.After(60 * time.Second)
	for !m.AllFinished() {
		select {
		case <-deadline:
			m.Stop()
			t.Fatal("games did not complete within timeout")
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
	defer m.Stop()

	// Valid indices should return non-nil sessions
	for i := 0; i < 3; i++ {
		session := m.GetSession(i)
		if session == nil {
			t.Errorf("GetSession(%d) returned nil, want non-nil", i)
		}
	}

	// Invalid indices should return nil
	if m.GetSession(-1) != nil {
		t.Error("GetSession(-1) should return nil")
	}
	if m.GetSession(3) != nil {
		t.Error("GetSession(3) should return nil for 3-game manager")
	}
	if m.GetSession(100) != nil {
		t.Error("GetSession(100) should return nil")
	}
}

// TestSessionManagerGetSessionBeforeStart verifies GetSession returns nil before Start.
func TestSessionManagerGetSessionBeforeStart(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 3, 0)

	// Before Start(), sessions haven't been created
	session := m.GetSession(0)
	if session != nil {
		t.Error("GetSession(0) should return nil before Start()")
	}
}

// TestSessionManagerGameCount verifies GameCount returns correct value.
func TestSessionManagerGameCount(t *testing.T) {
	tests := []struct {
		gameCount int
	}{
		{1},
		{3},
		{10},
		{50},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", tt.gameCount, 0)
			if m.GameCount() != tt.gameCount {
				t.Errorf("GameCount() = %d, want %d", m.GameCount(), tt.gameCount)
			}
		})
	}
}
