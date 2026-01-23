package bvb

import (
	"sync"
	"testing"
	"time"

	"github.com/Mgrdich/TermChess/internal/bot"
	"github.com/Mgrdich/TermChess/internal/engine"
)

func TestPlaybackSpeedDuration(t *testing.T) {
	tests := []struct {
		name  string
		speed PlaybackSpeed
		want  time.Duration
	}{
		{"instant", SpeedInstant, 0},
		{"fast", SpeedFast, 500 * time.Millisecond},
		{"normal", SpeedNormal, 1500 * time.Millisecond},
		{"slow", SpeedSlow, 3000 * time.Millisecond},
		{"unknown defaults to zero", PlaybackSpeed(99), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.speed.Duration()
			if got != tt.want {
				t.Errorf("PlaybackSpeed(%d).Duration() = %v, want %v", tt.speed, got, tt.want)
			}
		})
	}
}

func TestGameSessionRunsToCompletion(t *testing.T) {
	whiteEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create white engine: %v", err)
	}
	blackEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create black engine: %v", err)
	}

	speed := SpeedInstant
	session := NewGameSession(1, whiteEngine, blackEngine, "White Bot", "Black Bot", &speed)

	done := make(chan struct{})
	go func() {
		session.Run()
		close(done)
	}()

	// Wait for the game to finish with a generous timeout.
	select {
	case <-done:
		// Game completed.
	case <-time.After(60 * time.Second):
		session.Stop()
		t.Fatal("game did not complete within timeout")
	}

	if !session.IsFinished() {
		t.Error("session should be finished after Run() returns")
	}
}

func TestGameSessionResultPopulated(t *testing.T) {
	whiteEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create white engine: %v", err)
	}
	blackEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create black engine: %v", err)
	}

	speed := SpeedInstant
	session := NewGameSession(42, whiteEngine, blackEngine, "White Bot", "Black Bot", &speed)

	done := make(chan struct{})
	go func() {
		session.Run()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(60 * time.Second):
		session.Stop()
		t.Fatal("game did not complete within timeout")
	}

	result := session.Result()
	if result == nil {
		t.Fatal("result should not be nil after game completes")
	}

	if result.GameNumber != 42 {
		t.Errorf("GameNumber = %d, want 42", result.GameNumber)
	}

	if result.Winner == "" {
		t.Error("Winner should not be empty")
	}

	if result.EndReason == "" {
		t.Error("EndReason should not be empty")
	}

	if result.MoveCount <= 0 {
		t.Errorf("MoveCount = %d, want > 0", result.MoveCount)
	}

	if result.Duration <= 0 {
		t.Errorf("Duration = %v, want > 0", result.Duration)
	}

	if result.FinalFEN == "" {
		t.Error("FinalFEN should not be empty")
	}

	if len(result.MoveHistory) != result.MoveCount {
		t.Errorf("MoveHistory length = %d, want %d", len(result.MoveHistory), result.MoveCount)
	}

	// Verify winner is one of the expected values.
	validWinners := map[string]bool{
		"White Bot": true,
		"Black Bot": true,
		"Draw":      true,
	}
	if !validWinners[result.Winner] {
		t.Errorf("Winner = %q, want one of White Bot, Black Bot, or Draw", result.Winner)
	}
}

func TestGameSessionStop(t *testing.T) {
	whiteEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create white engine: %v", err)
	}
	blackEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create black engine: %v", err)
	}

	// Use a slow speed so the game does not finish instantly.
	speed := SpeedSlow
	session := NewGameSession(1, whiteEngine, blackEngine, "White Bot", "Black Bot", &speed)

	done := make(chan struct{})
	go func() {
		session.Run()
		close(done)
	}()

	// Give it a moment to start, then stop.
	time.Sleep(50 * time.Millisecond)
	session.Stop()

	select {
	case <-done:
		// Stopped successfully.
	case <-time.After(5 * time.Second):
		t.Fatal("session did not stop within timeout")
	}

	if !session.IsFinished() {
		t.Error("session should be finished after stop")
	}
}

func TestGameSessionConcurrentAccessors(t *testing.T) {
	whiteEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create white engine: %v", err)
	}
	blackEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create black engine: %v", err)
	}

	speed := SpeedFast
	session := NewGameSession(1, whiteEngine, blackEngine, "White Bot", "Black Bot", &speed)

	done := make(chan struct{})
	go func() {
		session.Run()
		close(done)
	}()

	// Run concurrent reads while the game is in progress.
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				board := session.CurrentBoard()
				if board == nil {
					t.Error("CurrentBoard() returned nil")
					return
				}

				_ = session.CurrentMoveHistory()
				_ = session.IsFinished()
				_ = session.State()
				_ = session.Result()
				_ = session.GameNumber()

				time.Sleep(5 * time.Millisecond)
			}
		}()
	}

	wg.Wait()

	// Stop the game and wait for completion.
	session.Stop()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("session did not stop within timeout")
	}
}

func TestGameSessionGameNumber(t *testing.T) {
	whiteEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create white engine: %v", err)
	}
	blackEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create black engine: %v", err)
	}

	speed := SpeedInstant
	session := NewGameSession(7, whiteEngine, blackEngine, "A", "B", &speed)

	if session.GameNumber() != 7 {
		t.Errorf("GameNumber() = %d, want 7", session.GameNumber())
	}
}

func TestGameSessionInitialState(t *testing.T) {
	whiteEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create white engine: %v", err)
	}
	blackEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create black engine: %v", err)
	}

	speed := SpeedInstant
	session := NewGameSession(1, whiteEngine, blackEngine, "W", "B", &speed)

	if session.State() != StateRunning {
		t.Errorf("initial state = %d, want StateRunning (%d)", session.State(), StateRunning)
	}

	if session.IsFinished() {
		t.Error("session should not be finished before Run()")
	}

	if session.Result() != nil {
		t.Error("result should be nil before game completes")
	}

	board := session.CurrentBoard()
	if board == nil {
		t.Fatal("CurrentBoard() should not return nil")
	}
	if board.ActiveColor != engine.White {
		t.Error("initial board should have White to move")
	}

	moves := session.CurrentMoveHistory()
	if len(moves) != 0 {
		t.Errorf("initial move history length = %d, want 0", len(moves))
	}

	// Clean up engines since Run() was never called.
	_ = whiteEngine.Close()
	_ = blackEngine.Close()
}

func TestMaxMoveCountConstant(t *testing.T) {
	if maxMoveCount != 500 {
		t.Errorf("maxMoveCount = %d, want 500", maxMoveCount)
	}
}
