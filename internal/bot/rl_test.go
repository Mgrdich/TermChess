package bot

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// TestRLEngineInterface is a compile-time check that *rlEngine implements Engine.
func TestRLEngineInterface(t *testing.T) {
	var _ Engine = (*rlEngine)(nil)
}

// TestRLEngineInspectable is a compile-time check that *rlEngine implements Inspectable.
func TestRLEngineInspectable(t *testing.T) {
	var _ Inspectable = (*rlEngine)(nil)
}

// TestNewRLEngine_Defaults verifies that each RLDifficulty creates an engine
// with the correct name and no error from the factory.
func TestNewRLEngine_Defaults(t *testing.T) {
	tests := []struct {
		difficulty RLDifficulty
		wantName   string
	}{
		{RLIntermediate, "RL Intermediate (1500)"},
		{RLAdvanced, "RL Advanced (2000)"},
		{RLMaster, "RL Master (2200)"},
	}

	for _, tc := range tests {
		t.Run(tc.wantName, func(t *testing.T) {
			eng, err := NewRLEngine(tc.difficulty)
			if err != nil {
				t.Fatalf("NewRLEngine(%d) error = %v, want nil", tc.difficulty, err)
			}
			defer eng.Close()

			if eng.Name() != tc.wantName {
				t.Errorf("Name() = %q, want %q", eng.Name(), tc.wantName)
			}
		})
	}
}

// TestNewRLEngine_InvalidDifficulty verifies that an invalid difficulty
// returns an error from the factory.
func TestNewRLEngine_InvalidDifficulty(t *testing.T) {
	eng, err := NewRLEngine(RLDifficulty(99))
	if err == nil {
		t.Error("NewRLEngine(99) error = nil, want error for invalid difficulty")
	}
	if eng != nil {
		t.Errorf("NewRLEngine(99) engine = %v, want nil", eng)
	}
}

// TestNewRLEngine_WithTimeLimit verifies that a custom time limit
// option overrides the default.
func TestNewRLEngine_WithTimeLimit(t *testing.T) {
	eng, err := NewRLEngine(RLIntermediate, WithTimeLimit(15*time.Second))
	if err != nil {
		t.Fatalf("NewRLEngine with WithTimeLimit error = %v, want nil", err)
	}
	defer eng.Close()

	rlEng, ok := eng.(*rlEngine)
	if !ok {
		t.Fatal("Expected engine to be *rlEngine")
	}

	if rlEng.timeLimit != 15*time.Second {
		t.Errorf("timeLimit = %v, want 15s", rlEng.timeLimit)
	}
}

// TestRLEngine_SelectMove_ReturnsModelNotLoaded verifies that SelectMove
// returns ErrModelNotLoaded since the ONNX model is not yet integrated.
func TestRLEngine_SelectMove_ReturnsModelNotLoaded(t *testing.T) {
	eng, err := NewRLEngine(RLIntermediate)
	if err != nil {
		t.Fatalf("NewRLEngine error = %v, want nil", err)
	}
	defer eng.Close()

	board := engine.NewBoard()

	_, err = eng.SelectMove(context.Background(), board)
	if err == nil {
		t.Fatal("SelectMove error = nil, want ErrModelNotLoaded")
	}

	if !errors.Is(err, ErrModelNotLoaded) {
		t.Errorf("SelectMove error = %v, want ErrModelNotLoaded", err)
	}
}

// TestRLEngine_Close verifies that Close works and that calling it
// a second time returns an error.
func TestRLEngine_Close(t *testing.T) {
	eng, err := NewRLEngine(RLAdvanced)
	if err != nil {
		t.Fatalf("NewRLEngine error = %v, want nil", err)
	}

	// First close should succeed
	err = eng.Close()
	if err != nil {
		t.Errorf("First Close() error = %v, want nil", err)
	}

	// Second close should return an error
	err = eng.Close()
	if err == nil {
		t.Error("Second Close() error = nil, want error for already closed")
	}
}

// TestRLEngine_Info verifies that Info() returns correct metadata.
func TestRLEngine_Info(t *testing.T) {
	eng, err := NewRLEngine(RLMaster)
	if err != nil {
		t.Fatalf("NewRLEngine error = %v, want nil", err)
	}
	defer eng.Close()

	inspectable, ok := eng.(Inspectable)
	if !ok {
		t.Fatal("rlEngine should implement Inspectable interface")
	}

	info := inspectable.Info()

	if info.Type != TypeRL {
		t.Errorf("Info().Type = %v, want TypeRL", info.Type)
	}

	if info.Name != "RL Master (2200)" {
		t.Errorf("Info().Name = %q, want %q", info.Name, "RL Master (2200)")
	}

	if info.Features == nil {
		t.Fatal("Info().Features = nil, want non-nil map")
	}

	if info.Features["onnx"] != false {
		t.Errorf("Info().Features[\"onnx\"] = %v, want false", info.Features["onnx"])
	}
}

// TestRLEngine_SelectMove_AfterClose verifies that SelectMove returns
// an error after the engine has been closed.
func TestRLEngine_SelectMove_AfterClose(t *testing.T) {
	eng, err := NewRLEngine(RLIntermediate)
	if err != nil {
		t.Fatalf("NewRLEngine error = %v, want nil", err)
	}

	err = eng.Close()
	if err != nil {
		t.Fatalf("Close() error = %v, want nil", err)
	}

	board := engine.NewBoard()

	_, err = eng.SelectMove(context.Background(), board)
	if err == nil {
		t.Error("SelectMove after Close error = nil, want error")
	}
	if err != nil && err.Error() != "engine is closed" {
		t.Errorf("SelectMove after Close error = %q, want %q", err.Error(), "engine is closed")
	}
}
