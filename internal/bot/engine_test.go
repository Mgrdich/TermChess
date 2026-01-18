package bot

import (
	"context"
	"testing"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// mockEngine is a minimal mock implementation for testing the Engine interface.
type mockEngine struct {
	name   string
	closed bool
}

func (m *mockEngine) SelectMove(ctx context.Context, board *engine.Board) (engine.Move, error) {
	// Return a simple mock move (e2e4)
	return engine.Move{
		From:      engine.NewSquare(4, 1), // e2
		To:        engine.NewSquare(4, 3), // e4
		Promotion: engine.Empty,
	}, nil
}

func (m *mockEngine) Name() string {
	return m.name
}

func (m *mockEngine) Close() error {
	m.closed = true
	return nil
}

// mockConfigurableEngine implements both Engine and Configurable.
type mockConfigurableEngine struct {
	mockEngine
	configured bool
	config     MinimaxConfig
}

func (m *mockConfigurableEngine) Configure(config MinimaxConfig) error {
	m.configured = true
	m.config = config
	return nil
}

// mockStatefulEngine implements both Engine and Stateful.
type mockStatefulEngine struct {
	mockEngine
	history []*engine.Board
}

func (m *mockStatefulEngine) SetPositionHistory(history []*engine.Board) error {
	m.history = history
	return nil
}

// mockInspectableEngine implements both Engine and Inspectable.
type mockInspectableEngine struct {
	mockEngine
	info Info
}

func (m *mockInspectableEngine) Info() Info {
	return m.info
}

// TestEngineInterface verifies that mockEngine implements the Engine interface.
func TestEngineInterface(t *testing.T) {
	var _ Engine = (*mockEngine)(nil)

	mock := &mockEngine{name: "TestEngine"}

	// Test Name method
	if got := mock.Name(); got != "TestEngine" {
		t.Errorf("Name() = %q, want %q", got, "TestEngine")
	}

	// Test SelectMove method
	board := engine.NewBoard()
	ctx := context.Background()
	move, err := mock.SelectMove(ctx, board)
	if err != nil {
		t.Errorf("SelectMove() error = %v, want nil", err)
	}
	if !move.From.IsValid() || !move.To.IsValid() {
		t.Errorf("SelectMove() returned invalid move: %v", move)
	}

	// Test Close method
	if mock.closed {
		t.Error("Close() should not be called yet")
	}
	if err := mock.Close(); err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
	if !mock.closed {
		t.Error("Close() did not mark engine as closed")
	}
}

// TestOptionalInterfaces verifies type assertions work for optional interfaces.
func TestOptionalInterfaces(t *testing.T) {
	t.Run("Configurable", func(t *testing.T) {
		var eng Engine = &mockConfigurableEngine{
			mockEngine: mockEngine{name: "ConfigurableEngine"},
		}

		// Type assertion should succeed
		configurable, ok := eng.(Configurable)
		if !ok {
			t.Fatal("Type assertion to Configurable failed")
		}

		// Test Configure method
		config := MinimaxConfig{SearchDepth: intPtr(5)}
		if err := configurable.Configure(config); err != nil {
			t.Errorf("Configure() error = %v, want nil", err)
		}

		// Verify configuration was applied
		mockCfg := eng.(*mockConfigurableEngine)
		if !mockCfg.configured {
			t.Error("Configure() did not set configured flag")
		}
		if mockCfg.config.SearchDepth == nil || *mockCfg.config.SearchDepth != 5 {
			t.Errorf("Configure() SearchDepth = %v, want 5", mockCfg.config.SearchDepth)
		}
	})

	t.Run("Stateful", func(t *testing.T) {
		var eng Engine = &mockStatefulEngine{
			mockEngine: mockEngine{name: "StatefulEngine"},
		}

		// Type assertion should succeed
		stateful, ok := eng.(Stateful)
		if !ok {
			t.Fatal("Type assertion to Stateful failed")
		}

		// Test SetPositionHistory method
		history := []*engine.Board{engine.NewBoard()}
		if err := stateful.SetPositionHistory(history); err != nil {
			t.Errorf("SetPositionHistory() error = %v, want nil", err)
		}

		// Verify history was stored
		mockStateful := eng.(*mockStatefulEngine)
		if len(mockStateful.history) != 1 {
			t.Errorf("SetPositionHistory() history length = %d, want 1", len(mockStateful.history))
		}
	})

	t.Run("Inspectable", func(t *testing.T) {
		expectedInfo := Info{
			Name:       "InspectableEngine",
			Author:     "Test Author",
			Version:    "1.0.0",
			Type:       TypeInternal,
			Difficulty: Medium,
			Features:   map[string]bool{"analysis": true},
		}

		var eng Engine = &mockInspectableEngine{
			mockEngine: mockEngine{name: "InspectableEngine"},
			info:       expectedInfo,
		}

		// Type assertion should succeed
		inspectable, ok := eng.(Inspectable)
		if !ok {
			t.Fatal("Type assertion to Inspectable failed")
		}

		// Test Info method
		info := inspectable.Info()
		if info.Name != expectedInfo.Name {
			t.Errorf("Info().Name = %q, want %q", info.Name, expectedInfo.Name)
		}
		if info.Author != expectedInfo.Author {
			t.Errorf("Info().Author = %q, want %q", info.Author, expectedInfo.Author)
		}
		if info.Version != expectedInfo.Version {
			t.Errorf("Info().Version = %q, want %q", info.Version, expectedInfo.Version)
		}
		if info.Type != expectedInfo.Type {
			t.Errorf("Info().Type = %v, want %v", info.Type, expectedInfo.Type)
		}
		if info.Difficulty != expectedInfo.Difficulty {
			t.Errorf("Info().Difficulty = %v, want %v", info.Difficulty, expectedInfo.Difficulty)
		}
	})

	t.Run("NonConfigurable", func(t *testing.T) {
		var eng Engine = &mockEngine{name: "BasicEngine"}

		// Type assertion should fail for basic engine
		_, ok := eng.(Configurable)
		if ok {
			t.Error("Type assertion to Configurable should fail for basic engine")
		}
	})
}

// TestEngineTypeConstants verifies EngineType constants are properly defined.
func TestEngineTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		typ      EngineType
		expected string
	}{
		{"Internal", TypeInternal, "Internal"},
		{"UCI", TypeUCI, "UCI"},
		{"RL", TypeRL, "RL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.typ.String(); got != tt.expected {
				t.Errorf("EngineType.String() = %q, want %q", got, tt.expected)
			}
		})
	}

	// Verify constant values
	if TypeInternal != 0 {
		t.Errorf("TypeInternal = %d, want 0", TypeInternal)
	}
	if TypeUCI != 1 {
		t.Errorf("TypeUCI = %d, want 1", TypeUCI)
	}
	if TypeRL != 2 {
		t.Errorf("TypeRL = %d, want 2", TypeRL)
	}
}

// TestDifficultyConstants verifies Difficulty constants are properly defined.
func TestDifficultyConstants(t *testing.T) {
	tests := []struct {
		name       string
		difficulty Difficulty
		expected   string
	}{
		{"Easy", Easy, "Easy"},
		{"Medium", Medium, "Medium"},
		{"Hard", Hard, "Hard"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.difficulty.String(); got != tt.expected {
				t.Errorf("Difficulty.String() = %q, want %q", got, tt.expected)
			}
		})
	}

	// Verify constant values
	if Easy != 0 {
		t.Errorf("Easy = %d, want 0", Easy)
	}
	if Medium != 1 {
		t.Errorf("Medium = %d, want 1", Medium)
	}
	if Hard != 2 {
		t.Errorf("Hard = %d, want 2", Hard)
	}
}

// TestInfoStruct verifies the Info struct can be created and used.
func TestInfoStruct(t *testing.T) {
	info := Info{
		Name:       "TestBot",
		Author:     "Test Author",
		Version:    "1.0.0",
		Type:       TypeInternal,
		Difficulty: Hard,
		Features: map[string]bool{
			"opening_book": true,
			"endgame_tb":   false,
		},
	}

	if info.Name != "TestBot" {
		t.Errorf("Info.Name = %q, want %q", info.Name, "TestBot")
	}
	if info.Author != "Test Author" {
		t.Errorf("Info.Author = %q, want %q", info.Author, "Test Author")
	}
	if info.Version != "1.0.0" {
		t.Errorf("Info.Version = %q, want %q", info.Version, "1.0.0")
	}
	if info.Type != TypeInternal {
		t.Errorf("Info.Type = %v, want %v", info.Type, TypeInternal)
	}
	if info.Difficulty != Hard {
		t.Errorf("Info.Difficulty = %v, want %v", info.Difficulty, Hard)
	}
	if !info.Features["opening_book"] {
		t.Error("Info.Features[opening_book] = false, want true")
	}
	if info.Features["endgame_tb"] {
		t.Error("Info.Features[endgame_tb] = true, want false")
	}
}

// TestEngineCloseIdempotency verifies Close() can be called multiple times safely.
func TestEngineCloseIdempotency(t *testing.T) {
	mock := &mockEngine{name: "TestEngine"}

	// Close multiple times should not error
	if err := mock.Close(); err != nil {
		t.Errorf("First Close() error = %v, want nil", err)
	}
	if err := mock.Close(); err != nil {
		t.Errorf("Second Close() error = %v, want nil", err)
	}
	if err := mock.Close(); err != nil {
		t.Errorf("Third Close() error = %v, want nil", err)
	}
}

// TestEngineWithContext verifies SelectMove respects context.
func TestEngineWithContext(t *testing.T) {
	mock := &mockEngine{name: "TestEngine"}
	board := engine.NewBoard()

	t.Run("ValidContext", func(t *testing.T) {
		ctx := context.Background()
		_, err := mock.SelectMove(ctx, board)
		if err != nil {
			t.Errorf("SelectMove() with valid context error = %v, want nil", err)
		}
	})

	t.Run("CancelledContext", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Mock doesn't check context, but interface allows it
		_, err := mock.SelectMove(ctx, board)
		// Mock implementation doesn't fail on cancelled context,
		// but real implementations should check ctx.Done()
		if err != nil {
			// This is actually the desired behavior for real engines
			t.Logf("SelectMove() correctly returned error on cancelled context: %v", err)
		}
	})
}
