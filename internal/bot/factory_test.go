package bot

import (
	"strings"
	"testing"
	"time"
)

// TestWithTimeLimit verifies the WithTimeLimit option.
func TestWithTimeLimit(t *testing.T) {
	t.Run("ValidTimeLimit", func(t *testing.T) {
		cfg := &engineConfig{}
		opt := WithTimeLimit(5 * time.Second)

		err := opt(cfg)
		if err != nil {
			t.Errorf("WithTimeLimit(5s) error = %v, want nil", err)
		}

		if cfg.timeLimit != 5*time.Second {
			t.Errorf("timeLimit = %v, want %v", cfg.timeLimit, 5*time.Second)
		}
	})

	t.Run("ZeroTimeLimit", func(t *testing.T) {
		cfg := &engineConfig{}
		opt := WithTimeLimit(0)

		err := opt(cfg)
		if err == nil {
			t.Error("WithTimeLimit(0) error = nil, want error")
		}

		if !strings.Contains(err.Error(), "positive") {
			t.Errorf("WithTimeLimit(0) error = %q, want error containing 'positive'", err.Error())
		}
	})

	t.Run("NegativeTimeLimit", func(t *testing.T) {
		cfg := &engineConfig{}
		opt := WithTimeLimit(-1 * time.Second)

		err := opt(cfg)
		if err == nil {
			t.Error("WithTimeLimit(-1s) error = nil, want error")
		}

		if !strings.Contains(err.Error(), "positive") {
			t.Errorf("WithTimeLimit(-1s) error = %q, want error containing 'positive'", err.Error())
		}
	})
}

// TestWithSearchDepth verifies the WithSearchDepth option.
func TestWithSearchDepth(t *testing.T) {
	validDepths := []int{1, 5, 10, 15, 20}
	for _, depth := range validDepths {
		t.Run("ValidDepth", func(t *testing.T) {
			cfg := &engineConfig{}
			opt := WithSearchDepth(depth)

			err := opt(cfg)
			if err != nil {
				t.Errorf("WithSearchDepth(%d) error = %v, want nil", depth, err)
			}

			if cfg.searchDepth != depth {
				t.Errorf("searchDepth = %d, want %d", cfg.searchDepth, depth)
			}
		})
	}

	t.Run("DepthZero", func(t *testing.T) {
		cfg := &engineConfig{}
		opt := WithSearchDepth(0)

		err := opt(cfg)
		if err == nil {
			t.Error("WithSearchDepth(0) error = nil, want error")
		}

		if !strings.Contains(err.Error(), "1-20") {
			t.Errorf("WithSearchDepth(0) error = %q, want error containing '1-20'", err.Error())
		}
	})

	t.Run("DepthTooHigh", func(t *testing.T) {
		cfg := &engineConfig{}
		opt := WithSearchDepth(21)

		err := opt(cfg)
		if err == nil {
			t.Error("WithSearchDepth(21) error = nil, want error")
		}

		if !strings.Contains(err.Error(), "1-20") {
			t.Errorf("WithSearchDepth(21) error = %q, want error containing '1-20'", err.Error())
		}
	})

	t.Run("NegativeDepth", func(t *testing.T) {
		cfg := &engineConfig{}
		opt := WithSearchDepth(-5)

		err := opt(cfg)
		if err == nil {
			t.Error("WithSearchDepth(-5) error = nil, want error")
		}

		if !strings.Contains(err.Error(), "1-20") {
			t.Errorf("WithSearchDepth(-5) error = %q, want error containing '1-20'", err.Error())
		}
	})
}

// TestWithOptions verifies the WithOptions option.
func TestWithOptions(t *testing.T) {
	t.Run("ValidOptions", func(t *testing.T) {
		cfg := &engineConfig{}
		customOpts := map[string]any{
			"threads":      4,
			"hash":         256,
			"opening_book": true,
		}
		opt := WithOptions(customOpts)

		err := opt(cfg)
		if err != nil {
			t.Errorf("WithOptions() error = %v, want nil", err)
		}

		if cfg.options == nil {
			t.Fatal("options = nil, want non-nil map")
		}

		if cfg.options["threads"] != 4 {
			t.Errorf("options[threads] = %v, want 4", cfg.options["threads"])
		}
		if cfg.options["hash"] != 256 {
			t.Errorf("options[hash] = %v, want 256", cfg.options["hash"])
		}
		if cfg.options["opening_book"] != true {
			t.Errorf("options[opening_book] = %v, want true", cfg.options["opening_book"])
		}
	})

	t.Run("EmptyOptions", func(t *testing.T) {
		cfg := &engineConfig{}
		customOpts := map[string]any{}
		opt := WithOptions(customOpts)

		err := opt(cfg)
		if err != nil {
			t.Errorf("WithOptions(empty) error = %v, want nil", err)
		}

		if cfg.options == nil {
			t.Error("options = nil, want empty map")
		}
	})

	t.Run("NilOptions", func(t *testing.T) {
		cfg := &engineConfig{}
		opt := WithOptions(nil)

		err := opt(cfg)
		if err != nil {
			t.Errorf("WithOptions(nil) error = %v, want nil", err)
		}
	})
}

// TestNewRandomEngine verifies random engine creation.
func TestNewRandomEngine(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		engine, err := NewRandomEngine()

		// Verify engine is created successfully
		if err != nil {
			t.Errorf("NewRandomEngine() error = %v, want nil", err)
		}

		if engine == nil {
			t.Fatal("NewRandomEngine() engine = nil, want non-nil engine")
		}

		// Verify engine name
		if engine.Name() != "Easy Bot" {
			t.Errorf("engine.Name() = %q, want 'Easy Bot'", engine.Name())
		}

		// Clean up
		engine.Close()
	})

	t.Run("CustomTimeLimit", func(t *testing.T) {
		engine, err := NewRandomEngine(WithTimeLimit(3 * time.Second))

		// Should create engine successfully with custom time limit
		if err != nil {
			t.Errorf("NewRandomEngine(WithTimeLimit) error = %v, want nil", err)
		}

		if engine == nil {
			t.Fatal("NewRandomEngine(WithTimeLimit) engine = nil, want non-nil engine")
		}

		// Verify custom time limit
		randomEng, ok := engine.(*randomEngine)
		if !ok {
			t.Fatal("Expected engine to be *randomEngine")
		}
		if randomEng.timeLimit != 3*time.Second {
			t.Errorf("timeLimit = %v, want 3s", randomEng.timeLimit)
		}

		// Clean up
		engine.Close()
	})

	t.Run("InvalidTimeLimit", func(t *testing.T) {
		engine, err := NewRandomEngine(WithTimeLimit(0))

		// Should return validation error
		if err == nil {
			t.Error("NewRandomEngine(invalid time) error = nil, want validation error")
		}

		if !strings.Contains(err.Error(), "positive") {
			t.Errorf("NewRandomEngine(invalid time) error = %q, want error containing 'positive'", err.Error())
		}

		if engine != nil {
			t.Errorf("NewRandomEngine(invalid time) engine = %v, want nil", engine)
		}
	})

	t.Run("CustomSearchDepth", func(t *testing.T) {
		// Random engine doesn't use search depth, but should accept it
		engine, err := NewRandomEngine(WithSearchDepth(5))

		// Should create engine successfully (search depth ignored)
		if err != nil {
			t.Errorf("NewRandomEngine(WithSearchDepth) error = %v, want nil", err)
		}

		if engine == nil {
			t.Fatal("NewRandomEngine(WithSearchDepth) engine = nil, want non-nil engine")
		}

		// Clean up
		engine.Close()
	})

	t.Run("MultipleOptions", func(t *testing.T) {
		engine, err := NewRandomEngine(
			WithTimeLimit(1*time.Second),
			WithOptions(map[string]any{"seed": 42}),
		)

		// Should create engine successfully
		if err != nil {
			t.Errorf("NewRandomEngine(multiple opts) error = %v, want nil", err)
		}

		if engine == nil {
			t.Fatal("NewRandomEngine(multiple opts) engine = nil, want non-nil engine")
		}

		// Clean up
		engine.Close()
	})
}

// TestNewMinimaxEngine verifies minimax engine creation.
func TestNewMinimaxEngine(t *testing.T) {
	t.Run("MediumDifficulty", func(t *testing.T) {
		engine, err := NewMinimaxEngine(Medium)

		// Verify engine is created successfully
		if err != nil {
			t.Errorf("NewMinimaxEngine(Medium) error = %v, want nil", err)
		}

		if engine == nil {
			t.Fatal("NewMinimaxEngine(Medium) engine = nil, want non-nil engine")
		}

		// Verify engine name
		if engine.Name() != "Medium Bot" {
			t.Errorf("engine.Name() = %q, want 'Medium Bot'", engine.Name())
		}

		// Verify it's a minimax engine with correct settings
		minimaxEng, ok := engine.(*minimaxEngine)
		if !ok {
			t.Fatal("Expected engine to be *minimaxEngine")
		}
		if minimaxEng.difficulty != Medium {
			t.Errorf("difficulty = %v, want Medium", minimaxEng.difficulty)
		}
		if minimaxEng.maxDepth != 4 {
			t.Errorf("maxDepth = %d, want 4", minimaxEng.maxDepth)
		}
		if minimaxEng.timeLimit != 4*time.Second {
			t.Errorf("timeLimit = %v, want 4s", minimaxEng.timeLimit)
		}

		// Clean up
		engine.Close()
	})

	t.Run("HardDifficulty", func(t *testing.T) {
		engine, err := NewMinimaxEngine(Hard)

		// Verify engine is created successfully
		if err != nil {
			t.Errorf("NewMinimaxEngine(Hard) error = %v, want nil", err)
		}

		if engine == nil {
			t.Fatal("NewMinimaxEngine(Hard) engine = nil, want non-nil engine")
		}

		// Verify engine name
		if engine.Name() != "Hard Bot" {
			t.Errorf("engine.Name() = %q, want 'Hard Bot'", engine.Name())
		}

		// Verify it's a minimax engine with correct settings
		minimaxEng, ok := engine.(*minimaxEngine)
		if !ok {
			t.Fatal("Expected engine to be *minimaxEngine")
		}
		if minimaxEng.difficulty != Hard {
			t.Errorf("difficulty = %v, want Hard", minimaxEng.difficulty)
		}
		if minimaxEng.maxDepth != 6 {
			t.Errorf("maxDepth = %d, want 6", minimaxEng.maxDepth)
		}
		if minimaxEng.timeLimit != 8*time.Second {
			t.Errorf("timeLimit = %v, want 8s", minimaxEng.timeLimit)
		}

		// Clean up
		engine.Close()
	})

	t.Run("EasyDifficultyInvalid", func(t *testing.T) {
		engine, err := NewMinimaxEngine(Easy)

		// Should return validation error about difficulty, not "not implemented"
		if err == nil {
			t.Error("NewMinimaxEngine(Easy) error = nil, want validation error")
		}

		if !strings.Contains(err.Error(), "invalid difficulty") {
			t.Errorf("NewMinimaxEngine(Easy) error = %q, want error containing 'invalid difficulty'", err.Error())
		}

		if engine != nil {
			t.Errorf("NewMinimaxEngine(Easy) engine = %v, want nil", engine)
		}
	})

	t.Run("CustomSearchDepth", func(t *testing.T) {
		engine, err := NewMinimaxEngine(Medium, WithSearchDepth(8))

		// Should create engine successfully with custom search depth
		if err != nil {
			t.Errorf("NewMinimaxEngine(Medium, depth) error = %v, want nil", err)
		}

		if engine == nil {
			t.Fatal("NewMinimaxEngine(Medium, depth) engine = nil, want non-nil engine")
		}

		// Verify custom search depth
		minimaxEng, ok := engine.(*minimaxEngine)
		if !ok {
			t.Fatal("Expected engine to be *minimaxEngine")
		}
		if minimaxEng.maxDepth != 8 {
			t.Errorf("maxDepth = %d, want 8", minimaxEng.maxDepth)
		}

		// Clean up
		engine.Close()
	})

	t.Run("CustomTimeLimit", func(t *testing.T) {
		engine, err := NewMinimaxEngine(Hard, WithTimeLimit(10*time.Second))

		// Should create engine successfully with custom time limit
		if err != nil {
			t.Errorf("NewMinimaxEngine(Hard, time) error = %v, want nil", err)
		}

		if engine == nil {
			t.Fatal("NewMinimaxEngine(Hard, time) engine = nil, want non-nil engine")
		}

		// Verify custom time limit
		minimaxEng, ok := engine.(*minimaxEngine)
		if !ok {
			t.Fatal("Expected engine to be *minimaxEngine")
		}
		if minimaxEng.timeLimit != 10*time.Second {
			t.Errorf("timeLimit = %v, want 10s", minimaxEng.timeLimit)
		}

		// Clean up
		engine.Close()
	})

	t.Run("InvalidSearchDepth", func(t *testing.T) {
		engine, err := NewMinimaxEngine(Medium, WithSearchDepth(0))

		// Should return validation error about search depth
		if err == nil {
			t.Error("NewMinimaxEngine(invalid depth) error = nil, want validation error")
		}

		if !strings.Contains(err.Error(), "1-20") {
			t.Errorf("NewMinimaxEngine(invalid depth) error = %q, want error containing '1-20'", err.Error())
		}

		if engine != nil {
			t.Errorf("NewMinimaxEngine(invalid depth) engine = %v, want nil", engine)
		}
	})

	t.Run("InvalidTimeLimit", func(t *testing.T) {
		engine, err := NewMinimaxEngine(Hard, WithTimeLimit(-1*time.Second))

		// Should return validation error about time limit
		if err == nil {
			t.Error("NewMinimaxEngine(invalid time) error = nil, want validation error")
		}

		if !strings.Contains(err.Error(), "positive") {
			t.Errorf("NewMinimaxEngine(invalid time) error = %q, want error containing 'positive'", err.Error())
		}

		if engine != nil {
			t.Errorf("NewMinimaxEngine(invalid time) engine = %v, want nil", engine)
		}
	})

	t.Run("MultipleOptions", func(t *testing.T) {
		engine, err := NewMinimaxEngine(
			Hard,
			WithTimeLimit(5*time.Second),
			WithSearchDepth(10),
			WithOptions(map[string]any{"transposition_table": true}),
		)

		// Should create engine successfully with multiple options
		if err != nil {
			t.Errorf("NewMinimaxEngine(multiple opts) error = %v, want nil", err)
		}

		if engine == nil {
			t.Fatal("NewMinimaxEngine(multiple opts) engine = nil, want non-nil engine")
		}

		// Verify all custom settings were applied
		minimaxEng, ok := engine.(*minimaxEngine)
		if !ok {
			t.Fatal("Expected engine to be *minimaxEngine")
		}
		if minimaxEng.timeLimit != 5*time.Second {
			t.Errorf("timeLimit = %v, want 5s", minimaxEng.timeLimit)
		}
		if minimaxEng.maxDepth != 10 {
			t.Errorf("maxDepth = %d, want 10", minimaxEng.maxDepth)
		}

		// Clean up
		engine.Close()
	})
}

// TestEngineConfigDefaults verifies default configurations for each difficulty.
func TestEngineConfigDefaults(t *testing.T) {
	t.Run("EasyDefaults", func(t *testing.T) {
		cfg := &engineConfig{
			difficulty: Easy,
			timeLimit:  2 * time.Second,
		}

		if cfg.difficulty != Easy {
			t.Errorf("difficulty = %v, want Easy", cfg.difficulty)
		}
		if cfg.timeLimit != 2*time.Second {
			t.Errorf("timeLimit = %v, want 2s", cfg.timeLimit)
		}
	})

	t.Run("MediumDefaults", func(t *testing.T) {
		cfg := &engineConfig{difficulty: Medium}

		// Simulate what NewMinimaxEngine does
		cfg.timeLimit = 4 * time.Second
		cfg.searchDepth = 4

		if cfg.difficulty != Medium {
			t.Errorf("difficulty = %v, want Medium", cfg.difficulty)
		}
		if cfg.timeLimit != 4*time.Second {
			t.Errorf("timeLimit = %v, want 4s", cfg.timeLimit)
		}
		if cfg.searchDepth != 4 {
			t.Errorf("searchDepth = %d, want 4", cfg.searchDepth)
		}
	})

	t.Run("HardDefaults", func(t *testing.T) {
		cfg := &engineConfig{difficulty: Hard}

		// Simulate what NewMinimaxEngine does
		cfg.timeLimit = 8 * time.Second
		cfg.searchDepth = 6

		if cfg.difficulty != Hard {
			t.Errorf("difficulty = %v, want Hard", cfg.difficulty)
		}
		if cfg.timeLimit != 8*time.Second {
			t.Errorf("timeLimit = %v, want 8s", cfg.timeLimit)
		}
		if cfg.searchDepth != 6 {
			t.Errorf("searchDepth = %d, want 6", cfg.searchDepth)
		}
	})
}

// TestEngineOptionChaining verifies options can be chained and applied in order.
func TestEngineOptionChaining(t *testing.T) {
	cfg := &engineConfig{}

	options := []EngineOption{
		WithTimeLimit(5 * time.Second),
		WithSearchDepth(10),
		WithOptions(map[string]any{"key": "value"}),
	}

	for _, opt := range options {
		if err := opt(cfg); err != nil {
			t.Errorf("Failed to apply option: %v", err)
		}
	}

	if cfg.timeLimit != 5*time.Second {
		t.Errorf("timeLimit = %v, want 5s", cfg.timeLimit)
	}
	if cfg.searchDepth != 10 {
		t.Errorf("searchDepth = %d, want 10", cfg.searchDepth)
	}
	if cfg.options == nil || cfg.options["key"] != "value" {
		t.Errorf("options = %v, want map with key=value", cfg.options)
	}
}

// TestEngineOptionOverrides verifies later options override earlier ones.
func TestEngineOptionOverrides(t *testing.T) {
	cfg := &engineConfig{}

	// Apply time limit twice, second should win
	opt1 := WithTimeLimit(3 * time.Second)
	opt2 := WithTimeLimit(7 * time.Second)

	if err := opt1(cfg); err != nil {
		t.Fatalf("opt1 error = %v, want nil", err)
	}
	if err := opt2(cfg); err != nil {
		t.Fatalf("opt2 error = %v, want nil", err)
	}

	if cfg.timeLimit != 7*time.Second {
		t.Errorf("timeLimit = %v, want 7s (should be overridden)", cfg.timeLimit)
	}
}
