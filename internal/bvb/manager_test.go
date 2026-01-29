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

// TestNewSessionManagerConcurrencyCap verifies that concurrency is capped at maxConcurrentGames.
func TestNewSessionManagerConcurrencyCap(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 10, 100)

	if m.Concurrency() != maxConcurrentGames {
		t.Errorf("Concurrency() = %d, want %d (capped at max)", m.Concurrency(), maxConcurrentGames)
	}
}

// TestNewSessionManagerConcurrencyMinimum verifies that concurrency has a minimum of 1.
func TestNewSessionManagerConcurrencyMinimum(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 10, -5)

	if m.Concurrency() != 1 {
		t.Errorf("Concurrency() = %d, want 1 (minimum)", m.Concurrency())
	}
}
