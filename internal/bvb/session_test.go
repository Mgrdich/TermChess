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
		session.Abort()
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
		session.Abort()
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

func TestGameSessionAbort(t *testing.T) {
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
	session.Abort()

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
	session.Abort()
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

func TestGameSessionPauseBlocksProgress(t *testing.T) {
	whiteEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create white engine: %v", err)
	}
	blackEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create black engine: %v", err)
	}

	// Use SpeedFast so the game does not finish before we can pause it.
	speed := SpeedFast
	session := NewGameSession(1, whiteEngine, blackEngine, "White Bot", "Black Bot", &speed)

	done := make(chan struct{})
	go func() {
		session.Run()
		close(done)
	}()

	// Give it a moment to start and make some moves.
	time.Sleep(100 * time.Millisecond)

	session.Pause()

	// Allow any in-flight move to complete after pause signal is sent.
	time.Sleep(50 * time.Millisecond)

	// Record move count after pausing and settling.
	movesBefore := len(session.CurrentMoveHistory())

	// Wait and verify no progress is made.
	time.Sleep(300 * time.Millisecond)

	movesAfter := len(session.CurrentMoveHistory())
	if movesAfter != movesBefore {
		t.Errorf("moves changed during pause: before=%d, after=%d", movesBefore, movesAfter)
	}

	if session.State() != StatePaused {
		t.Errorf("state = %d, want StatePaused (%d)", session.State(), StatePaused)
	}

	// Resume and abort to clean up.
	session.Resume()
	session.Abort()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("session did not stop within timeout")
	}
}

func TestGameSessionResumeAfterPause(t *testing.T) {
	whiteEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create white engine: %v", err)
	}
	blackEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create black engine: %v", err)
	}

	// Use SpeedSlow to ensure the game doesn't finish between Resume() and state check.
	speed := SpeedSlow
	session := NewGameSession(1, whiteEngine, blackEngine, "White Bot", "Black Bot", &speed)

	done := make(chan struct{})
	go func() {
		session.Run()
		close(done)
	}()

	// Give it a moment to start.
	time.Sleep(50 * time.Millisecond)

	// Pause the game.
	session.Pause()
	time.Sleep(100 * time.Millisecond)

	// Resume the game.
	session.Resume()

	if session.State() != StateRunning {
		t.Errorf("state after resume = %d, want StateRunning (%d)", session.State(), StateRunning)
	}

	// Abort the session (it's using SpeedSlow so it won't finish naturally in a test).
	session.Abort()

	// Wait for the goroutine to finish.
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("session did not stop within timeout after abort")
	}

	if !session.IsFinished() {
		t.Error("session should be finished after abort")
	}
}

func TestGameSessionAbortStopsGame(t *testing.T) {
	whiteEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create white engine: %v", err)
	}
	blackEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create black engine: %v", err)
	}

	speed := SpeedSlow
	session := NewGameSession(1, whiteEngine, blackEngine, "White Bot", "Black Bot", &speed)

	done := make(chan struct{})
	go func() {
		session.Run()
		close(done)
	}()

	// Give it a moment to start.
	time.Sleep(50 * time.Millisecond)

	// Abort.
	session.Abort()

	// Verify Run() returns quickly.
	select {
	case <-done:
		// Stopped successfully.
	case <-time.After(2 * time.Second):
		t.Fatal("session did not abort within timeout")
	}

	if !session.IsFinished() {
		t.Error("session should be finished after abort")
	}
}

func TestGameSessionAbortDuringPause(t *testing.T) {
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

	// Give it a moment to start.
	time.Sleep(50 * time.Millisecond)

	// Pause the game.
	session.Pause()
	time.Sleep(50 * time.Millisecond)

	// Abort while paused.
	session.Abort()

	// Verify Run() returns quickly.
	select {
	case <-done:
		// Stopped successfully.
	case <-time.After(2 * time.Second):
		t.Fatal("session did not abort during pause within timeout")
	}

	if !session.IsFinished() {
		t.Error("session should be finished after abort during pause")
	}
}

func TestGameSessionInstantSpeed(t *testing.T) {
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

	// Instant speed with random engines should complete well within 5 seconds.
	select {
	case <-done:
		// Game completed quickly as expected.
	case <-time.After(5 * time.Second):
		session.Abort()
		t.Fatal("instant speed game did not complete within 5 seconds")
	}

	if !session.IsFinished() {
		t.Error("session should be finished after Run() returns")
	}

	result := session.Result()
	if result == nil {
		t.Fatal("result should not be nil after game completes")
	}
	if result.MoveCount <= 0 {
		t.Errorf("MoveCount = %d, want > 0", result.MoveCount)
	}
}

func TestGameSessionNormalSpeedHasDelays(t *testing.T) {
	whiteEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create white engine: %v", err)
	}
	blackEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create black engine: %v", err)
	}

	// Normal speed = 1500ms delay per move.
	speed := SpeedNormal
	session := NewGameSession(1, whiteEngine, blackEngine, "White Bot", "Black Bot", &speed)

	done := make(chan struct{})
	go func() {
		session.Run()
		close(done)
	}()

	// Wait 2 seconds. With 1500ms delay per move, expect at most 1-2 moves.
	time.Sleep(2 * time.Second)

	moveCount := len(session.CurrentMoveHistory())
	if moveCount > 3 {
		t.Errorf("expected at most 3 moves in 2s at normal speed (1500ms/move), got %d", moveCount)
	}

	// Clean up.
	session.Abort()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("session did not stop within timeout after abort")
	}
}

func TestGameSessionSpeedChangeMidGame(t *testing.T) {
	whiteEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create white engine: %v", err)
	}
	blackEngine, err := bot.NewRandomEngine()
	if err != nil {
		t.Fatalf("failed to create black engine: %v", err)
	}

	// Start with slow speed (3000ms per move).
	speed := SpeedSlow
	session := NewGameSession(1, whiteEngine, blackEngine, "White Bot", "Black Bot", &speed)

	done := make(chan struct{})
	go func() {
		session.Run()
		close(done)
	}()

	// Wait 1 second. At 3000ms/move, game should have 0-1 moves.
	time.Sleep(1 * time.Second)

	movesBeforeChange := len(session.CurrentMoveHistory())
	if movesBeforeChange > 1 {
		t.Errorf("expected at most 1 move in 1s at slow speed (3000ms/move), got %d", movesBeforeChange)
	}

	// Change speed to instant using the thread-safe SetSpeed method.
	session.SetSpeed(SpeedInstant)

	// Wait 3 seconds for the game to progress rapidly at instant speed.
	time.Sleep(3 * time.Second)

	movesAfterChange := len(session.CurrentMoveHistory())

	// After switching to instant, many more moves should have been made
	// (or the game should have finished entirely).
	if !session.IsFinished() && movesAfterChange <= movesBeforeChange+2 {
		t.Errorf("expected significant progress after speed change to instant: before=%d, after=%d",
			movesBeforeChange, movesAfterChange)
	}

	// Clean up if not already finished.
	if !session.IsFinished() {
		session.Abort()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("session did not stop within timeout after abort")
		}
	}
}
