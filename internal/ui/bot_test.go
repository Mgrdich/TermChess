package ui

import (
	"errors"
	"testing"
	"time"

	"github.com/Mgrdich/TermChess/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

// TestBotMoveExecution tests the bot move execution flow
func TestBotMoveExecution(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenGamePlay
	m.gameType = GameTypePvBot
	m.botDifficulty = BotEasy
	m.board = engine.NewBoard()
	m.moveHistory = []engine.Move{}

	// Player makes a move (e2e4)
	m.input = "e4"
	result, cmd := m.handleMoveInput()
	m = result.(Model)

	// Should have cleared input
	if m.input != "" {
		t.Errorf("Expected input to be cleared, got: %s", m.input)
	}

	// Should have a thinking message
	if m.statusMsg == "" {
		t.Errorf("Expected status message with thinking text, got empty string")
	}

	// Should have returned a command to execute bot move
	if cmd == nil {
		t.Errorf("Expected a command to be returned for bot move execution")
	}

	// Should have bot engine created
	if m.botEngine == nil {
		t.Errorf("Expected bot engine to be created")
	}

	// Execute the bot move command to get the message
	if cmd != nil {
		msg := cmd()

		// Should return either BotMoveMsg or BotMoveErrorMsg
		switch msg.(type) {
		case BotMoveMsg, BotMoveErrorMsg:
			// Expected message types
		default:
			t.Errorf("Expected BotMoveMsg or BotMoveErrorMsg, got: %T", msg)
		}
	}

	// Clean up
	if m.botEngine != nil {
		_ = m.botEngine.Close()
	}
}

// TestBotDifficultySelection tests the bot difficulty selection flow
func TestBotDifficultySelection(t *testing.T) {
	tests := []struct {
		name       string
		selection  int
		option     string
		difficulty BotDifficulty
	}{
		{"Easy", 0, "Easy", BotEasy},
		{"Medium", 1, "Medium", BotMedium},
		{"Hard", 2, "Hard", BotHard},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(DefaultConfig())
			m.screen = ScreenBotSelect
			m.gameType = GameTypePvBot
			m.menuOptions = []string{"Easy", "Medium", "Hard"}
			m.menuSelection = tt.selection

			result, _ := m.handleBotDifficultySelection()
			m = result.(Model)

			// Should transition to ColorSelect screen
			if m.screen != ScreenColorSelect {
				t.Errorf("Expected screen to be ScreenColorSelect, got: %v", m.screen)
			}

			// Should set bot difficulty
			if m.botDifficulty != tt.difficulty {
				t.Errorf("Expected difficulty to be %v, got: %v", tt.difficulty, m.botDifficulty)
			}

			// Menu options should be color choices
			if len(m.menuOptions) != 2 {
				t.Errorf("Expected 2 menu options, got: %d", len(m.menuOptions))
			}
			if m.menuOptions[0] != "Play as White" {
				t.Errorf("Expected first option to be 'Play as White', got: %s", m.menuOptions[0])
			}
			if m.menuOptions[1] != "Play as Black" {
				t.Errorf("Expected second option to be 'Play as Black', got: %s", m.menuOptions[1])
			}

			// Should have cleared status messages
			if m.statusMsg != "" {
				t.Errorf("Expected status message to be cleared, got: %s", m.statusMsg)
			}
			if m.errorMsg != "" {
				t.Errorf("Expected error message to be cleared, got: %s", m.errorMsg)
			}
		})
	}
}

// TestBotMoveHandling tests handling of bot move messages
func TestBotMoveHandling(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenGamePlay
	m.gameType = GameTypePvBot
	m.botDifficulty = BotEasy
	m.board = engine.NewBoard()
	m.moveHistory = []engine.Move{}
	m.statusMsg = "Thinking..."

	// Create a valid bot move (e7e5)
	move, _ := engine.ParseMove("e7e5")

	// Make player move first (e2e4)
	playerMove, _ := engine.ParseMove("e2e4")
	_ = m.board.MakeMove(playerMove)
	m.moveHistory = append(m.moveHistory, playerMove)

	// Handle bot move message
	msg := BotMoveMsg{move: move}
	result, _ := m.handleBotMove(msg)
	m = result.(Model)

	// Should clear status message
	if m.statusMsg != "" {
		t.Errorf("Expected status message to be cleared, got: %s", m.statusMsg)
	}

	// Should add move to history
	if len(m.moveHistory) != 2 {
		t.Errorf("Expected 2 moves in history, got: %d", len(m.moveHistory))
	}

	// Should not show error
	if m.errorMsg != "" {
		t.Errorf("Expected no error message, got: %s", m.errorMsg)
	}
}

// TestBotMoveErrorHandling tests handling of bot move errors
func TestBotMoveErrorHandling(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenGamePlay
	m.gameType = GameTypePvBot
	m.statusMsg = "Thinking..."

	// Create an error message
	msg := BotMoveErrorMsg{err: errors.New("test error")}
	result, _ := m.handleBotMoveError(msg)
	m = result.(Model)

	// Should clear status message
	if m.statusMsg != "" {
		t.Errorf("Expected status message to be cleared, got: %s", m.statusMsg)
	}

	// Should show error
	if m.errorMsg == "" {
		t.Errorf("Expected error message to be set")
	}
}

// TestBotSelectNavigation tests navigation in bot select screen
func TestBotSelectNavigation(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBotSelect
	m.menuOptions = []string{"Easy", "Medium", "Hard"}
	m.menuSelection = 0

	// Test down navigation
	result, _ := m.handleBotSelectKeys(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(Model)
	if m.menuSelection != 1 {
		t.Errorf("Expected selection 1, got: %d", m.menuSelection)
	}

	// Test up navigation
	result, _ = m.handleBotSelectKeys(tea.KeyMsg{Type: tea.KeyUp})
	m = result.(Model)
	if m.menuSelection != 0 {
		t.Errorf("Expected selection 0, got: %d", m.menuSelection)
	}

	// Test wrap around down
	m.menuSelection = 2
	result, _ = m.handleBotSelectKeys(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(Model)
	if m.menuSelection != 0 {
		t.Errorf("Expected selection to wrap to 0, got: %d", m.menuSelection)
	}

	// Test wrap around up
	m.menuSelection = 0
	result, _ = m.handleBotSelectKeys(tea.KeyMsg{Type: tea.KeyUp})
	m = result.(Model)
	if m.menuSelection != 2 {
		t.Errorf("Expected selection to wrap to 2, got: %d", m.menuSelection)
	}

	// Test ESC returns to game type select
	result, _ = m.handleBotSelectKeys(tea.KeyMsg{Type: tea.KeyEsc})
	m = result.(Model)
	if m.screen != ScreenGameTypeSelect {
		t.Errorf("Expected screen to be ScreenGameTypeSelect, got: %v", m.screen)
	}
}

// TestColorSelection tests the color selection flow
func TestColorSelection(t *testing.T) {
	tests := []struct {
		name           string
		selection      int
		option         string
		expectedColor  engine.Color
		expectBotMove  bool
	}{
		{"Play as White", 0, "Play as White", engine.White, false},
		{"Play as Black", 1, "Play as Black", engine.Black, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(DefaultConfig())
			m.screen = ScreenColorSelect
			m.gameType = GameTypePvBot
			m.botDifficulty = BotEasy
			m.menuOptions = []string{"Play as White", "Play as Black"}
			m.menuSelection = tt.selection

			result, cmd := m.handleColorSelection()
			m = result.(Model)

			// Should transition to GamePlay screen
			if m.screen != ScreenGamePlay {
				t.Errorf("Expected screen to be ScreenGamePlay, got: %v", m.screen)
			}

			// Should set user color
			if m.userColor != tt.expectedColor {
				t.Errorf("Expected user color to be %v, got: %v", tt.expectedColor, m.userColor)
			}

			// Should create a new board
			if m.board == nil {
				t.Errorf("Expected board to be created")
			}

			// Should have cleared status messages initially
			if m.errorMsg != "" {
				t.Errorf("Expected error message to be cleared, got: %s", m.errorMsg)
			}

			// If user plays Black, should trigger bot move
			if tt.expectBotMove {
				if cmd == nil {
					t.Errorf("Expected a command to be returned for bot move execution")
				}
				if m.botEngine == nil {
					t.Errorf("Expected bot engine to be created")
				}
				if m.statusMsg == "" {
					t.Errorf("Expected status message with thinking text when bot moves first")
				}
				// Clean up bot engine
				if m.botEngine != nil {
					_ = m.botEngine.Close()
				}
			} else {
				// If user plays White, should not trigger bot move
				if m.statusMsg != "" {
					t.Errorf("Expected no status message when user plays White, got: %s", m.statusMsg)
				}
			}
		})
	}
}

// TestColorSelectNavigation tests navigation in color select screen
func TestColorSelectNavigation(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenColorSelect
	m.menuOptions = []string{"Play as White", "Play as Black"}
	m.menuSelection = 0

	// Test down navigation
	result, _ := m.handleColorSelectKeys(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(Model)
	if m.menuSelection != 1 {
		t.Errorf("Expected selection 1, got: %d", m.menuSelection)
	}

	// Test up navigation
	result, _ = m.handleColorSelectKeys(tea.KeyMsg{Type: tea.KeyUp})
	m = result.(Model)
	if m.menuSelection != 0 {
		t.Errorf("Expected selection 0, got: %d", m.menuSelection)
	}

	// Test wrap around down
	m.menuSelection = 1
	result, _ = m.handleColorSelectKeys(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(Model)
	if m.menuSelection != 0 {
		t.Errorf("Expected selection to wrap to 0, got: %d", m.menuSelection)
	}

	// Test wrap around up
	m.menuSelection = 0
	result, _ = m.handleColorSelectKeys(tea.KeyMsg{Type: tea.KeyUp})
	m = result.(Model)
	if m.menuSelection != 1 {
		t.Errorf("Expected selection to wrap to 1, got: %d", m.menuSelection)
	}

	// Test ESC returns to bot difficulty select
	result, _ = m.handleColorSelectKeys(tea.KeyMsg{Type: tea.KeyEsc})
	m = result.(Model)
	if m.screen != ScreenBotSelect {
		t.Errorf("Expected screen to be ScreenBotSelect, got: %v", m.screen)
	}
}

// TestBotPlaysWhite_FullFlow tests the complete flow when user plays as Black
func TestBotPlaysWhite_FullFlow(t *testing.T) {
	// Start with bot difficulty selection
	m := NewModel(DefaultConfig())
	m.screen = ScreenBotSelect
	m.gameType = GameTypePvBot
	m.menuOptions = []string{"Easy", "Medium", "Hard"}
	m.menuSelection = 0 // Select Easy

	// Select bot difficulty
	result, _ := m.handleBotDifficultySelection()
	m = result.(Model)

	// Should be at color selection screen
	if m.screen != ScreenColorSelect {
		t.Fatalf("Expected ScreenColorSelect, got: %v", m.screen)
	}

	// Select "Play as Black" (index 1)
	m.menuSelection = 1
	result, cmd := m.handleColorSelection()
	m = result.(Model)

	// Should be at GamePlay screen
	if m.screen != ScreenGamePlay {
		t.Errorf("Expected ScreenGamePlay, got: %v", m.screen)
	}

	// User color should be Black
	if m.userColor != engine.Black {
		t.Errorf("Expected userColor to be Black, got: %v", m.userColor)
	}

	// Board should be created
	if m.board == nil {
		t.Fatalf("Expected board to be created")
	}

	// Board should still be at initial position (White to move)
	if m.board.ActiveColor != engine.White {
		t.Errorf("Expected ActiveColor to be White, got: %v", m.board.ActiveColor)
	}

	// Bot move command should be returned
	if cmd == nil {
		t.Fatalf("Expected bot move command to be returned")
	}

	// Bot engine should be created
	if m.botEngine == nil {
		t.Fatalf("Expected bot engine to be created")
	}

	// Status message should indicate bot is thinking
	if m.statusMsg == "" {
		t.Errorf("Expected thinking status message")
	}

	// Execute the bot move command
	msg := cmd()

	// Should return BotMoveMsg (bot should make a valid move)
	botMoveMsg, ok := msg.(BotMoveMsg)
	if !ok {
		t.Fatalf("Expected BotMoveMsg, got: %T", msg)
	}

	// Process the bot move
	result, _ = m.handleBotMove(botMoveMsg)
	m = result.(Model)

	// After bot's move, it should be Black's turn
	if m.board.ActiveColor != engine.Black {
		t.Errorf("Expected ActiveColor to be Black after bot move, got: %v", m.board.ActiveColor)
	}

	// Move history should have one move (bot's opening move)
	if len(m.moveHistory) != 1 {
		t.Errorf("Expected 1 move in history, got: %d", len(m.moveHistory))
	}

	// Status message should be cleared after bot move
	if m.statusMsg != "" {
		t.Errorf("Expected status message to be cleared after bot move, got: %s", m.statusMsg)
	}

	// Clean up
	if m.botEngine != nil {
		_ = m.botEngine.Close()
	}
}

// TestBotPlaysBlack_FullFlow tests the complete flow when user plays as White
func TestBotPlaysBlack_FullFlow(t *testing.T) {
	// Start with bot difficulty selection
	m := NewModel(DefaultConfig())
	m.screen = ScreenBotSelect
	m.gameType = GameTypePvBot
	m.menuOptions = []string{"Easy", "Medium", "Hard"}
	m.menuSelection = 1 // Select Medium

	// Select bot difficulty
	result, _ := m.handleBotDifficultySelection()
	m = result.(Model)

	// Should be at color selection screen
	if m.screen != ScreenColorSelect {
		t.Fatalf("Expected ScreenColorSelect, got: %v", m.screen)
	}

	// Select "Play as White" (index 0)
	m.menuSelection = 0
	result, cmd := m.handleColorSelection()
	m = result.(Model)

	// Should be at GamePlay screen
	if m.screen != ScreenGamePlay {
		t.Errorf("Expected ScreenGamePlay, got: %v", m.screen)
	}

	// User color should be White
	if m.userColor != engine.White {
		t.Errorf("Expected userColor to be White, got: %v", m.userColor)
	}

	// Board should be created
	if m.board == nil {
		t.Fatalf("Expected board to be created")
	}

	// Board should be at initial position (White to move)
	if m.board.ActiveColor != engine.White {
		t.Errorf("Expected ActiveColor to be White, got: %v", m.board.ActiveColor)
	}

	// No bot move command should be returned (user moves first)
	if cmd != nil {
		t.Errorf("Expected no bot move command when user plays White")
	}

	// Status message should be empty (no thinking message)
	if m.statusMsg != "" {
		t.Errorf("Expected no status message when user plays White, got: %s", m.statusMsg)
	}

	// Move history should be empty
	if len(m.moveHistory) != 0 {
		t.Errorf("Expected empty move history, got: %d moves", len(m.moveHistory))
	}
}

// TestBotEngineCleanup tests that bot engine is cleaned up properly
func TestBotEngineCleanup(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenGamePlay
	m.gameType = GameTypePvBot
	m.botDifficulty = BotEasy
	m.board = engine.NewBoard()

	// Create a bot engine by making a bot move
	m, _ = m.makeBotMove()
	if m.botEngine == nil {
		t.Fatalf("Expected bot engine to be created")
	}

	// Simulate game over
	result, _ := m.handleGameOverKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = result.(Model)

	// Bot engine should be cleaned up (we can't check if Close was called,
	// but we can verify the function doesn't panic)
}

// TestBotMoveDelay tests that bot moves enforce a minimum delay
func TestBotMoveDelay(t *testing.T) {
	tests := []struct {
		name       string
		difficulty BotDifficulty
		minDelay   time.Duration
		maxDelay   time.Duration
	}{
		// Note: Total time = minimum delay (1-2s) + computation time
		// CI environments may be slower, so we allow extra time
		{"Easy bot has 1-2s delay", BotEasy, 1 * time.Second, 3 * time.Second},
		{"Medium bot has 1-2s delay", BotMedium, 1 * time.Second, 4 * time.Second},
		{"Hard bot has 1s+ delay", BotHard, 1 * time.Second, 10 * time.Second}, // Hard can take longer naturally
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(DefaultConfig())
			m.screen = ScreenGamePlay
			m.gameType = GameTypePvBot
			m.botDifficulty = tt.difficulty
			m.board = engine.NewBoard()
			m.moveHistory = []engine.Move{}

			// Player makes a move (e2e4)
			move, _ := engine.ParseMove("e2e4")
			err := m.board.MakeMove(move)
			if err != nil {
				t.Fatalf("Failed to make player move: %v", err)
			}
			m.moveHistory = append(m.moveHistory, move)

			// Bot should respond
			m, cmd := m.makeBotMove()

			if cmd == nil {
				t.Fatal("Expected bot move command")
			}

			// Measure time for bot move
			start := time.Now()
			msg := cmd()
			elapsed := time.Since(start)

			// Verify it's a successful move
			botMoveMsg, ok := msg.(BotMoveMsg)
			if !ok {
				t.Fatalf("Expected BotMoveMsg, got: %T", msg)
			}

			// Verify minimum delay was enforced
			if elapsed < tt.minDelay {
				t.Errorf("Expected delay >= %v, got %v", tt.minDelay, elapsed)
			}

			// Verify it's not excessively long (sanity check)
			if elapsed > tt.maxDelay {
				t.Errorf("Expected delay <= %v, got %v (might indicate timeout issue)", tt.maxDelay, elapsed)
			}

			// Verify move is valid
			if botMoveMsg.move == (engine.Move{}) {
				t.Error("Bot returned empty move")
			}

			// Clean up
			if m.botEngine != nil {
				_ = m.botEngine.Close()
			}
		})
	}
}

// TestGetMinimumBotDelay tests the delay calculation function
func TestGetMinimumBotDelay(t *testing.T) {
	tests := []struct {
		name       string
		difficulty BotDifficulty
		minBound   time.Duration
		maxBound   time.Duration
	}{
		{"Easy: 1-2 seconds", BotEasy, 1 * time.Second, 2 * time.Second},
		{"Medium: 1-2 seconds", BotMedium, 1 * time.Second, 2 * time.Second},
		{"Hard: 1 second", BotHard, 1 * time.Second, 1 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test multiple times to verify randomization (for Easy/Medium)
			for i := 0; i < 10; i++ {
				delay := getMinimumBotDelay(tt.difficulty)

				if delay < tt.minBound {
					t.Errorf("Delay %v is less than minimum %v", delay, tt.minBound)
				}

				if delay > tt.maxBound {
					t.Errorf("Delay %v is greater than maximum %v", delay, tt.maxBound)
				}
			}
		})
	}
}
