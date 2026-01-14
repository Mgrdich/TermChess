package ui

import (
	"errors"
	"testing"

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

			// Should transition to GamePlay screen
			if m.screen != ScreenGamePlay {
				t.Errorf("Expected screen to be ScreenGamePlay, got: %v", m.screen)
			}

			// Should set bot difficulty
			if m.botDifficulty != tt.difficulty {
				t.Errorf("Expected difficulty to be %v, got: %v", tt.difficulty, m.botDifficulty)
			}

			// Should create a new board
			if m.board == nil {
				t.Errorf("Expected board to be created")
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
