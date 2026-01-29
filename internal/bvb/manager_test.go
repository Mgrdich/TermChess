package bvb

import (
	"testing"
	"time"

	"github.com/Mgrdich/TermChess/internal/bot"
)

func TestNewSessionManager(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "Easy Bot", "Easy Bot", 3)
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
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 3)
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
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 3)
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
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 2)
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
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 2)
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
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 3)
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
	m := NewSessionManager(bot.Easy, bot.Easy, "White", "Black", 3)
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
