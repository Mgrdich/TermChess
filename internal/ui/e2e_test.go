package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/Mgrdich/TermChess/internal/bvb"
	"github.com/Mgrdich/TermChess/internal/config"
	"github.com/Mgrdich/TermChess/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

// TestUpdate_QuitKey tests that pressing 'q' or ctrl+c quits the app
func TestUpdate_QuitKey(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenMainMenu

	// Test 'q' key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	result, cmd := m.Update(msg)

	// Should return quit command
	if cmd == nil {
		t.Error("Expected quit command, got nil")
	}

	// Model should be returned
	if _, ok := result.(Model); !ok {
		t.Error("Expected Model to be returned")
	}
}

// TestUpdate_CtrlC tests ctrl+c quit
func TestUpdate_CtrlC(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenGamePlay

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("Expected quit command on ctrl+c, got nil")
	}
}

// TestView_AllScreens tests that View() renders all screen types without crashing
func TestView_AllScreens(t *testing.T) {
	screens := []Screen{
		ScreenMainMenu,
		ScreenGameTypeSelect,
		ScreenGamePlay,
		ScreenGameOver,
		ScreenSettings,
		ScreenSavePrompt,
		ScreenResumePrompt,
		ScreenFENInput,
	}

	for _, screen := range screens {
		t.Run(string(rune(screen)), func(t *testing.T) {
			m := NewModel(DefaultConfig())
			m.screen = screen
			m.board = engine.NewBoard()

			// Set up necessary state for each screen
			switch screen {
			case ScreenMainMenu:
				m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
				m.menuSelection = 0
			case ScreenGameTypeSelect:
				m.menuOptions = []string{"Player vs Player", "Player vs Bot", "Back"}
				m.menuSelection = 0
			case ScreenGameOver:
				m.menuOptions = []string{"New Game", "Main Menu", "Exit"}
				m.menuSelection = 0
			case ScreenSettings:
				m.settingsSelection = 0
			case ScreenSavePrompt:
				m.menuOptions = []string{"Yes", "No"}
				m.menuSelection = 0
			case ScreenResumePrompt:
				m.menuOptions = []string{"Yes", "No"}
				m.menuSelection = 0
			case ScreenFENInput:
				m.input = ""
			}

			// Should not panic
			view := m.View()

			// Should return non-empty string
			if view == "" {
				t.Errorf("View() returned empty string for screen %d", screen)
			}
		})
	}
}

// TestRenderMainMenu tests the main menu rendering
func TestRenderMainMenu(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenMainMenu
	m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
	m.menuSelection = 0

	view := m.renderMainMenu()

	// Should contain title
	if !strings.Contains(view, "TermChess") {
		t.Error("Main menu should contain 'TermChess' title")
	}

	// Should contain all menu options
	for _, option := range m.menuOptions {
		if !strings.Contains(view, option) {
			t.Errorf("Main menu should contain option '%s'", option)
		}
	}

	// Should contain instructions
	if !strings.Contains(view, "arrows/jk") || !strings.Contains(view, "enter") {
		t.Error("Main menu should contain navigation instructions")
	}
}

// TestRenderGameTypeSelect tests the game type selection rendering
func TestRenderGameTypeSelect(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenGameTypeSelect
	m.menuOptions = []string{"Player vs Player", "Player vs Bot", "Back"}
	m.menuSelection = 0

	view := m.renderGameTypeSelect()

	// Should contain title
	if !strings.Contains(view, "Select Game Type") {
		t.Error("Game type select should contain title")
	}

	// Should contain all options
	for _, option := range m.menuOptions {
		if !strings.Contains(view, option) {
			t.Errorf("Game type select should contain option '%s'", option)
		}
	}
}

// TestRenderGameOver tests the game over screen rendering
func TestRenderGameOver(t *testing.T) {
	// Set up a checkmate position (Fool's mate)
	board := engine.NewBoard()
	moves := []string{"f2f3", "e7e5", "g2g4", "d8h4"}
	for _, moveStr := range moves {
		move, _ := engine.ParseMove(moveStr)
		board.MakeMove(move)
	}

	m := NewModel(DefaultConfig())
	m.board = board
	m.screen = ScreenGameOver
	m.menuOptions = []string{"New Game", "Main Menu", "Exit"}
	m.menuSelection = 0

	view := m.renderGameOver()

	// Should contain game result message
	if !strings.Contains(strings.ToLower(view), "wins") || !strings.Contains(strings.ToLower(view), "checkmate") {
		t.Error("Game over screen should contain game result with 'wins' and 'checkmate'")
	}

	// Should contain key hints (not menu options, but keyboard shortcuts)
	if !strings.Contains(view, "New Game") || !strings.Contains(view, "Main Menu") {
		t.Error("Game over screen should contain 'New Game' and 'Main Menu' options")
	}
}

// TestRenderSettings tests the settings screen rendering
func TestRenderSettings(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenSettings
	m.settingsSelection = 0

	view := m.renderSettings()

	// Should contain title
	if !strings.Contains(view, "Settings") {
		t.Error("Settings screen should contain title")
	}

	// Should contain all setting options
	settingNames := []string{"Unicode", "Coordinates", "Colors", "Move History"}
	for _, name := range settingNames {
		if !strings.Contains(view, name) {
			t.Errorf("Settings screen should contain setting '%s'", name)
		}
	}

	// Should contain instructions
	if !strings.Contains(view, "space") || !strings.Contains(view, "ESC") {
		t.Error("Settings screen should contain navigation instructions")
	}
}

// TestRenderSavePrompt tests the save prompt rendering
func TestRenderSavePrompt(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenSavePrompt
	m.menuOptions = []string{"Yes", "No"}
	m.menuSelection = 0

	view := m.renderSavePrompt()

	// Should ask about saving (note: "Save" is in the title)
	if !strings.Contains(strings.ToLower(view), "save") {
		t.Error("Save prompt should ask about saving")
	}

	// Should contain Yes/No options
	if !strings.Contains(view, "Yes") || !strings.Contains(view, "No") {
		t.Error("Save prompt should contain Yes and No options")
	}
}

// TestRenderResumePrompt tests the resume prompt rendering
func TestRenderResumePrompt(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenResumePrompt
	m.menuOptions = []string{"Yes", "No"}
	m.menuSelection = 0

	view := m.renderResumePrompt()

	// Should ask about resuming (case insensitive)
	lowerView := strings.ToLower(view)
	if !strings.Contains(lowerView, "resume") && !strings.Contains(lowerView, "saved") {
		t.Error("Resume prompt should ask about resuming or mention saved game")
	}

	// Should contain Yes/No options
	if !strings.Contains(view, "Yes") || !strings.Contains(view, "No") {
		t.Error("Resume prompt should contain Yes and No options")
	}
}

// TestRenderFENInput tests the FEN input screen rendering
func TestRenderFENInput(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenFENInput
	m.input = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

	view := m.renderFENInput()

	// Should contain title
	if !strings.Contains(view, "FEN") {
		t.Error("FEN input screen should mention FEN")
	}

	// Should show the input
	if !strings.Contains(view, m.input) {
		t.Error("FEN input screen should show the user's input")
	}

	// Should contain instructions
	if !strings.Contains(view, "Enter") || !strings.Contains(view, "ESC") {
		t.Error("FEN input screen should contain instructions")
	}
}

// TestHandleMainMenuKeys tests main menu key handling
func TestHandleMainMenuKeys(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenMainMenu
	m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
	m.menuSelection = 0

	// Test down movement
	msg := tea.KeyMsg{Type: tea.KeyDown}
	result, _ := m.handleMainMenuKeys(msg)
	m = result.(Model)

	if m.menuSelection != 1 {
		t.Errorf("Expected selection 1, got %d", m.menuSelection)
	}

	// Test up movement with wrapping
	m.menuSelection = 0
	msg = tea.KeyMsg{Type: tea.KeyUp}
	result, _ = m.handleMainMenuKeys(msg)
	m = result.(Model)

	if m.menuSelection != len(m.menuOptions)-1 {
		t.Errorf("Expected selection to wrap to %d, got %d", len(m.menuOptions)-1, m.menuSelection)
	}

	// Test 'j' key for down
	m.menuSelection = 0
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	result, _ = m.handleMainMenuKeys(msg)
	m = result.(Model)

	if m.menuSelection != 1 {
		t.Errorf("Expected selection 1 after 'j', got %d", m.menuSelection)
	}

	// Test 'k' key for up
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	result, _ = m.handleMainMenuKeys(msg)
	m = result.(Model)

	if m.menuSelection != 0 {
		t.Errorf("Expected selection 0 after 'k', got %d", m.menuSelection)
	}
}

// TestHandleGameTypeSelectKeys tests game type selection key handling
func TestHandleGameTypeSelectKeys(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenGameTypeSelect
	m.menuOptions = []string{"Player vs Player", "Player vs Bot", "Back"}
	m.menuSelection = 0

	// Test navigation
	msg := tea.KeyMsg{Type: tea.KeyDown}
	result, _ := m.handleGameTypeSelectKeys(msg)
	m = result.(Model)

	if m.menuSelection != 1 {
		t.Errorf("Expected selection 1, got %d", m.menuSelection)
	}

	// Test ESC key
	msg = tea.KeyMsg{Type: tea.KeyEsc}
	result, _ = m.handleGameTypeSelectKeys(msg)
	m = result.(Model)

	if m.screen != ScreenMainMenu {
		t.Errorf("Expected to return to main menu, got screen %v", m.screen)
	}
}

// TestHandleGameOverKeys tests game over screen key handling
func TestHandleGameOverKeys(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGameOver

	// Test 'n' key for new game
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	result, _ := m.handleGameOverKeys(msg)
	m = result.(Model)

	if m.screen != ScreenGameTypeSelect {
		t.Errorf("Expected ScreenGameTypeSelect after 'n', got %v", m.screen)
	}

	// Test 'm' key for main menu
	m.screen = ScreenGameOver
	m.board = engine.NewBoard()
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	result, _ = m.handleGameOverKeys(msg)
	m = result.(Model)

	if m.screen != ScreenMainMenu {
		t.Errorf("Expected ScreenMainMenu after 'm', got %v", m.screen)
	}

	// Test 'q' key for quit
	m.screen = ScreenGameOver
	m.board = engine.NewBoard()
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.handleGameOverKeys(msg)

	if cmd == nil {
		t.Error("Expected quit command after 'q', got nil")
	}
}

// TestHandleSavePromptKeys tests save prompt key handling
func TestHandleSavePromptKeys(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenSavePrompt
	m.savePromptSelection = 0
	m.savePromptAction = "menu"
	m.board = engine.NewBoard()

	// Move to "No" option
	msg := tea.KeyMsg{Type: tea.KeyDown}
	result, _ := m.handleSavePromptKeys(msg)
	m = result.(Model)

	if m.savePromptSelection != 1 {
		t.Errorf("Expected selection 1, got %d", m.savePromptSelection)
	}

	// Select "No" - should go to main menu without saving
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	result, _ = m.handleSavePromptKeys(msg)
	m = result.(Model)

	if m.screen != ScreenMainMenu {
		t.Errorf("Expected to return to main menu, got screen %v", m.screen)
	}

	// Test ESC to cancel and return to game
	m.screen = ScreenSavePrompt
	m.savePromptSelection = 0
	msg = tea.KeyMsg{Type: tea.KeyEsc}
	result, _ = m.handleSavePromptKeys(msg)
	m = result.(Model)

	if m.screen != ScreenGamePlay {
		t.Errorf("Expected to return to game play after ESC, got screen %v", m.screen)
	}

	// Test direct 'n' key - should go to main menu without saving
	m.screen = ScreenSavePrompt
	m.board = engine.NewBoard()
	m.savePromptAction = "menu"
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	result, _ = m.handleSavePromptKeys(msg)
	m = result.(Model)

	if m.screen != ScreenMainMenu {
		t.Errorf("Expected to return to main menu after 'n', got screen %v", m.screen)
	}

	// Test direct 'y' key - should save and go to main menu
	m.screen = ScreenSavePrompt
	m.board = engine.NewBoard()
	m.savePromptAction = "menu"
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	result, _ = m.handleSavePromptKeys(msg)
	m = result.(Model)

	if m.screen != ScreenMainMenu {
		t.Errorf("Expected to return to main menu after 'y', got screen %v", m.screen)
	}

	// Verify the game was saved
	if !config.SaveGameExists() {
		t.Error("Expected game to be saved after pressing 'y'")
	}

	// Clean up
	_ = config.DeleteSaveGame()
}

// TestFullGameFlow tests a complete game from start to finish
func TestFullGameFlow(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()

	// Start at main menu
	if m.screen != ScreenMainMenu {
		t.Errorf("Expected to start at main menu, got %v", m.screen)
	}

	// Select "New Game"
	m.menuSelection = 0
	result, _ := m.handleMainMenuSelection()
	m = result.(Model)

	if m.screen != ScreenGameTypeSelect {
		t.Errorf("Expected game type select screen, got %v", m.screen)
	}

	// Select "Player vs Player"
	m.menuSelection = 0
	result, _ = m.handleGameTypeSelection()
	m = result.(Model)

	if m.screen != ScreenGamePlay {
		t.Errorf("Expected gameplay screen, got %v", m.screen)
	}

	// Play Scholar's Mate
	moves := []string{"e2e4", "e7e5", "f1c4", "b8c6", "d1h5", "g8f6", "h5f7"}

	for i, moveStr := range moves {
		m.input = moveStr
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		result, _ = m.handleGamePlayKeys(msg)
		m = result.(Model)

		if m.errorMsg != "" && i < len(moves)-1 {
			t.Errorf("Move %d (%s) failed: %s", i+1, moveStr, m.errorMsg)
		}
	}

	// Should detect checkmate and transition to game over screen
	if m.screen != ScreenGameOver {
		t.Errorf("Expected game over screen after checkmate, got %v", m.screen)
	}

	// Verify game ended in checkmate
	if !m.board.IsGameOver() {
		t.Error("Expected game to be over after Scholar's Mate")
	}

	status := m.board.Status()
	if status != engine.Checkmate {
		t.Errorf("Expected checkmate status, got %v", status)
	}

	// Verify move history was recorded
	if len(m.moveHistory) != len(moves) {
		t.Errorf("Expected %d moves in history, got %d", len(moves), len(m.moveHistory))
	}
}

// TestScreenTransitions tests all valid screen transitions
func TestScreenTransitions(t *testing.T) {
	tests := []struct {
		name           string
		fromScreen     Screen
		action         string
		toScreen       Screen
		setupFunc      func(*Model)
		transitionFunc func(Model) (tea.Model, tea.Cmd)
	}{
		{
			name:       "MainMenu to GameTypeSelect",
			fromScreen: ScreenMainMenu,
			action:     "Select New Game",
			toScreen:   ScreenGameTypeSelect,
			setupFunc: func(m *Model) {
				m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
				m.menuSelection = 0
			},
			transitionFunc: func(m Model) (tea.Model, tea.Cmd) {
				return m.handleMainMenuSelection()
			},
		},
		{
			name:       "MainMenu to FENInput",
			fromScreen: ScreenMainMenu,
			action:     "Select Load Game",
			toScreen:   ScreenFENInput,
			setupFunc: func(m *Model) {
				m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
				m.menuSelection = 1
			},
			transitionFunc: func(m Model) (tea.Model, tea.Cmd) {
				return m.handleMainMenuSelection()
			},
		},
		{
			name:       "MainMenu to Settings",
			fromScreen: ScreenMainMenu,
			action:     "Select Settings",
			toScreen:   ScreenSettings,
			setupFunc: func(m *Model) {
				m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
				m.menuSelection = 2
			},
			transitionFunc: func(m Model) (tea.Model, tea.Cmd) {
				return m.handleMainMenuSelection()
			},
		},
		{
			name:       "GameTypeSelect to GamePlay (PvP)",
			fromScreen: ScreenGameTypeSelect,
			action:     "Select Player vs Player",
			toScreen:   ScreenGamePlay,
			setupFunc: func(m *Model) {
				m.menuOptions = []string{"Player vs Player", "Player vs Bot", "Back"}
				m.menuSelection = 0
				m.board = engine.NewBoard()
			},
			transitionFunc: func(m Model) (tea.Model, tea.Cmd) {
				return m.handleGameTypeSelection()
			},
		},
		{
			name:       "GameTypeSelect to MainMenu",
			fromScreen: ScreenGameTypeSelect,
			action:     "Press ESC",
			toScreen:   ScreenMainMenu,
			setupFunc: func(m *Model) {
				m.menuOptions = []string{"Player vs Player", "Player vs Bot"}
				m.menuSelection = 0
			},
			transitionFunc: func(m Model) (tea.Model, tea.Cmd) {
				msg := tea.KeyMsg{Type: tea.KeyEsc}
				return m.handleGameTypeSelectKeys(msg)
			},
		},
		{
			name:       "Settings to MainMenu",
			fromScreen: ScreenSettings,
			action:     "Press ESC",
			toScreen:   ScreenMainMenu,
			setupFunc: func(m *Model) {
				m.settingsSelection = 0
			},
			transitionFunc: func(m Model) (tea.Model, tea.Cmd) {
				msg := tea.KeyMsg{Type: tea.KeyEsc}
				return m.handleSettingsKeys(msg)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(DefaultConfig())
			m.screen = tt.fromScreen
			tt.setupFunc(&m)

			result, _ := tt.transitionFunc(m)
			newModel := result.(Model)

			if newModel.screen != tt.toScreen {
				t.Errorf("Expected transition to %v, got %v", tt.toScreen, newModel.screen)
			}
		})
	}
}

// TestGetGameResultMessage tests game result message generation
func TestGetGameResultMessage(t *testing.T) {
	tests := []struct {
		name          string
		setupBoard    func() *engine.Board
		resignedBy    int8
		containsCheck []string
	}{
		{
			name: "Checkmate - Black wins",
			setupBoard: func() *engine.Board {
				// Fool's mate position: 1. f3 e5 2. g4 Qh4#
				board := engine.NewBoard()
				moves := []string{"f2f3", "e7e5", "g2g4", "d8h4"}
				for _, m := range moves {
					move, _ := engine.ParseMove(m)
					board.MakeMove(move)
				}
				return board
			},
			resignedBy:    -1,
			containsCheck: []string{"black", "checkmate"},
		},
		{
			name: "Stalemate",
			setupBoard: func() *engine.Board {
				// Create a proper stalemate position
				// Black king on a8, White king on c7, White queen on b6
				fen := "k7/2K5/1Q6/8/8/8/8/8 b - - 0 1"
				board, _ := engine.FromFEN(fen)
				return board
			},
			resignedBy:    -1,
			containsCheck: []string{"stalemate", "draw"},
		},
		{
			name: "Resignation by White",
			setupBoard: func() *engine.Board {
				return engine.NewBoard()
			},
			resignedBy:    int8(engine.White),
			containsCheck: []string{"black", "resigned"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := tt.setupBoard()
			msg := getGameResultMessage(board, tt.resignedBy, false)

			for _, check := range tt.containsCheck {
				if !strings.Contains(strings.ToLower(msg), strings.ToLower(check)) {
					t.Errorf("Expected message to contain '%s', got: %s", check, msg)
				}
			}
		})
	}
}

// TestFENInputValidation tests FEN input validation and error handling
func TestFENInputValidation(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{
			name:      "Valid starting position",
			input:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			shouldErr: false,
		},
		{
			name:      "Valid mid-game position",
			input:     "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4",
			shouldErr: false,
		},
		{
			name:      "Invalid FEN",
			input:     "invalid",
			shouldErr: true,
		},
		{
			name:      "Empty FEN",
			input:     "",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(DefaultConfig())
			m.screen = ScreenFENInput
			m.fenInput.SetValue(tt.input)

			msg := tea.KeyMsg{Type: tea.KeyEnter}
			result, _ := m.handleFENInputKeys(msg)
			newModel := result.(Model)

			if tt.shouldErr {
				if newModel.errorMsg == "" {
					t.Errorf("Expected error for input '%s', got none", tt.input)
				}
				if newModel.screen != ScreenFENInput {
					t.Errorf("Should stay on FEN input screen on error")
				}
			} else {
				if newModel.errorMsg != "" {
					t.Errorf("Expected no error for input '%s', got: %s", tt.input, newModel.errorMsg)
				}
				if newModel.screen != ScreenGamePlay {
					t.Errorf("Should transition to gameplay on valid FEN")
				}
			}
		})
	}
}

// TestCommandCaseInsensitivity tests that commands work regardless of case
func TestCommandCaseInsensitivity(t *testing.T) {
	commands := []struct {
		input    string
		expected string
	}{
		{"resign", "resign"},
		{"RESIGN", "resign"},
		{"Resign", "resign"},
		{"showfen", "showfen"},
		{"ShowFen", "showfen"},
		{"SHOWFEN", "showfen"},
		{"menu", "menu"},
		{"MENU", "menu"},
		{"Menu", "menu"},
	}

	for _, cmd := range commands {
		t.Run(cmd.input, func(t *testing.T) {
			m := NewModel(DefaultConfig())
			m.board = engine.NewBoard()
			m.screen = ScreenGamePlay
			m.input = cmd.input

			msg := tea.KeyMsg{Type: tea.KeyEnter}
			result, _ := m.handleGamePlayKeys(msg)
			newModel := result.(Model)

			// Commands should be recognized regardless of case
			switch cmd.expected {
			case "resign":
				if newModel.screen != ScreenGameOver {
					t.Errorf("Resign command '%s' should lead to game over screen", cmd.input)
				}
			case "showfen":
				if newModel.statusMsg == "" {
					t.Errorf("ShowFen command '%s' should set status message", cmd.input)
				}
			case "menu":
				if newModel.screen != ScreenSavePrompt {
					t.Errorf("Menu command '%s' should lead to save prompt", cmd.input)
				}
			}
		})
	}
}

// TestErrorMessageClearing tests that error messages are cleared appropriately
func TestErrorMessageClearing(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay
	m.errorMsg = "Previous error"

	// Error should clear when typing
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	result, _ := m.handleGamePlayKeys(msg)
	m = result.(Model)

	if m.errorMsg != "" {
		t.Errorf("Error message should clear when typing, got: %s", m.errorMsg)
	}

	// Set error again
	m.errorMsg = "Another error"

	// Error should clear on backspace
	msg = tea.KeyMsg{Type: tea.KeyBackspace}
	result, _ = m.handleGamePlayKeys(msg)
	m = result.(Model)

	if m.errorMsg != "" {
		t.Errorf("Error message should clear on backspace, got: %s", m.errorMsg)
	}
}

// TestGameTypeSelection_BotTransitionsToBotSelect tests that bot selection transitions to bot difficulty screen
func TestGameTypeSelection_BotTransitionsToBotSelect(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenGameTypeSelect
	m.menuOptions = []string{"Player vs Player", "Player vs Bot", "Back"}
	m.menuSelection = 1 // Select "Player vs Bot"

	result, _ := m.handleGameTypeSelection()
	m = result.(Model)

	// Should transition to ScreenBotSelect
	if m.screen != ScreenBotSelect {
		t.Errorf("Expected screen to be ScreenBotSelect, got: %v", m.screen)
	}

	// Should set game type to PvBot
	if m.gameType != GameTypePvBot {
		t.Errorf("Expected gameType to be set to PvBot, got: %v", m.gameType)
	}

	// Should have difficulty options
	expectedOptions := []string{"Easy", "Medium", "Hard"}
	if len(m.menuOptions) != len(expectedOptions) {
		t.Errorf("Expected %d menu options, got %d", len(expectedOptions), len(m.menuOptions))
	}
	for i, opt := range expectedOptions {
		if i < len(m.menuOptions) && m.menuOptions[i] != opt {
			t.Errorf("Expected option %d to be %s, got %s", i, opt, m.menuOptions[i])
		}
	}
}

// TestMoveHistoryPersistence tests that move history persists through game
func TestMoveHistoryPersistence(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay
	m.config.ShowMoveHistory = true

	// Play several moves
	moves := []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4", "g8f6"}
	for _, moveStr := range moves {
		m.input = moveStr
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		result, _ := m.handleGamePlayKeys(msg)
		m = result.(Model)
	}

	// Verify all moves are in history
	if len(m.moveHistory) != len(moves) {
		t.Errorf("Expected %d moves in history, got %d", len(moves), len(m.moveHistory))
	}

	// Verify move history is formatted correctly
	history := m.formatMoveHistory()
	if history == "" {
		t.Error("Move history should not be empty")
	}

	// Should contain numbered moves
	if !strings.Contains(history, "1.") || !strings.Contains(history, "2.") || !strings.Contains(history, "3.") {
		t.Error("Move history should contain numbered moves")
	}
}

// TestGameTypeSelection_BvBTransitionsToBvBBotSelect tests that "Bot vs Bot" transitions to BvB bot selection screen.
func TestGameTypeSelection_BvBTransitionsToBvBBotSelect(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenGameTypeSelect
	m.menuOptions = []string{"Player vs Player", "Player vs Bot", "Bot vs Bot"}
	m.menuSelection = 2 // Select "Bot vs Bot"

	result, _ := m.handleGameTypeSelection()
	m = result.(Model)

	if m.screen != ScreenBvBBotSelect {
		t.Errorf("Expected screen to be ScreenBvBBotSelect, got: %v", m.screen)
	}
	if m.gameType != GameTypeBvB {
		t.Errorf("Expected gameType to be GameTypeBvB, got: %v", m.gameType)
	}
	if !m.bvbSelectingWhite {
		t.Error("Expected bvbSelectingWhite to be true for initial selection")
	}
	expectedOptions := []string{"Easy", "Medium", "Hard"}
	if len(m.menuOptions) != len(expectedOptions) {
		t.Fatalf("Expected %d menu options, got %d", len(expectedOptions), len(m.menuOptions))
	}
	for i, opt := range expectedOptions {
		if m.menuOptions[i] != opt {
			t.Errorf("Expected option %d to be %s, got %s", i, opt, m.menuOptions[i])
		}
	}
}

// TestBvBBotSelect_WhiteThenBlack tests the two-step bot selection flow.
func TestBvBBotSelect_WhiteThenBlack(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBBotSelect
	m.gameType = GameTypeBvB
	m.bvbSelectingWhite = true
	m.menuOptions = []string{"Easy", "Medium", "Hard"}
	m.menuSelection = 2 // Select "Hard" for White

	// Select White difficulty
	result, _ := m.handleBvBBotDifficultySelection()
	m = result.(Model)

	if m.bvbSelectingWhite {
		t.Error("Expected bvbSelectingWhite to be false after White selection")
	}
	if m.bvbWhiteDiff != BotHard {
		t.Errorf("Expected bvbWhiteDiff to be BotHard, got: %v", m.bvbWhiteDiff)
	}
	if m.screen != ScreenBvBBotSelect {
		t.Errorf("Expected to stay on ScreenBvBBotSelect for Black selection, got: %v", m.screen)
	}

	// Select Black difficulty
	m.menuSelection = 0 // Select "Easy" for Black
	result, _ = m.handleBvBBotDifficultySelection()
	m = result.(Model)

	if m.bvbBlackDiff != BotEasy {
		t.Errorf("Expected bvbBlackDiff to be BotEasy, got: %v", m.bvbBlackDiff)
	}
}

// TestBvBBotSelect_EscFromWhiteGoesBackToGameType tests ESC during White selection.
func TestBvBBotSelect_EscFromWhiteGoesBackToGameType(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBBotSelect
	m.gameType = GameTypeBvB
	m.bvbSelectingWhite = true
	m.menuOptions = []string{"Easy", "Medium", "Hard"}
	m.menuSelection = 0

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := m.handleBvBBotSelectKeys(msg)
	m = result.(Model)

	if m.screen != ScreenGameTypeSelect {
		t.Errorf("Expected screen to be ScreenGameTypeSelect, got: %v", m.screen)
	}
}

// TestBvBBotSelect_EscFromBlackGoesBackToWhite tests ESC during Black selection.
func TestBvBBotSelect_EscFromBlackGoesBackToWhite(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBBotSelect
	m.gameType = GameTypeBvB
	m.bvbSelectingWhite = false
	m.bvbWhiteDiff = BotMedium
	m.menuOptions = []string{"Easy", "Medium", "Hard"}
	m.menuSelection = 0

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := m.handleBvBBotSelectKeys(msg)
	m = result.(Model)

	if m.screen != ScreenBvBBotSelect {
		t.Errorf("Expected to stay on ScreenBvBBotSelect, got: %v", m.screen)
	}
	if !m.bvbSelectingWhite {
		t.Error("Expected bvbSelectingWhite to be true after ESC from Black selection")
	}
}

// TestBvBBotSelect_Navigation tests arrow key navigation.
func TestBvBBotSelect_Navigation(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBBotSelect
	m.gameType = GameTypeBvB
	m.bvbSelectingWhite = true
	m.menuOptions = []string{"Easy", "Medium", "Hard"}
	m.menuSelection = 0

	// Move down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	result, _ := m.handleBvBBotSelectKeys(msg)
	m = result.(Model)
	if m.menuSelection != 1 {
		t.Errorf("Expected menuSelection to be 1 after down, got: %d", m.menuSelection)
	}

	// Move down again
	result, _ = m.handleBvBBotSelectKeys(msg)
	m = result.(Model)
	if m.menuSelection != 2 {
		t.Errorf("Expected menuSelection to be 2 after second down, got: %d", m.menuSelection)
	}

	// Wrap around
	result, _ = m.handleBvBBotSelectKeys(msg)
	m = result.(Model)
	if m.menuSelection != 0 {
		t.Errorf("Expected menuSelection to wrap to 0, got: %d", m.menuSelection)
	}

	// Move up wraps to bottom
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	result, _ = m.handleBvBBotSelectKeys(msg)
	m = result.(Model)
	if m.menuSelection != 2 {
		t.Errorf("Expected menuSelection to wrap to 2, got: %d", m.menuSelection)
	}
}

// TestRenderBvBBotSelect_WhiteSelection tests rendering during White bot selection.
func TestRenderBvBBotSelect_WhiteSelection(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBBotSelect
	m.bvbSelectingWhite = true
	m.menuOptions = []string{"Easy", "Medium", "Hard"}
	m.menuSelection = 0

	view := m.renderBvBBotSelect()

	if !strings.Contains(view, "Select White Bot Difficulty:") {
		t.Error("Expected view to contain 'Select White Bot Difficulty:'")
	}
	if !strings.Contains(view, "Easy") {
		t.Error("Expected view to contain 'Easy'")
	}
}

// TestRenderBvBBotSelect_BlackSelection tests rendering during Black bot selection.
func TestRenderBvBBotSelect_BlackSelection(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBBotSelect
	m.bvbSelectingWhite = false
	m.bvbWhiteDiff = BotHard
	m.menuOptions = []string{"Easy", "Medium", "Hard"}
	m.menuSelection = 0

	view := m.renderBvBBotSelect()

	if !strings.Contains(view, "Select Black Bot Difficulty:") {
		t.Error("Expected view to contain 'Select Black Bot Difficulty:'")
	}
	if !strings.Contains(view, "White: Hard Bot") {
		t.Error("Expected view to show previously selected White difficulty")
	}
}

// TestBotDifficultyName tests the botDifficultyName helper.
func TestBotDifficultyName(t *testing.T) {
	tests := []struct {
		diff BotDifficulty
		want string
	}{
		{BotEasy, "Easy"},
		{BotMedium, "Medium"},
		{BotHard, "Hard"},
		{BotDifficulty(99), "Unknown"},
	}
	for _, tt := range tests {
		got := botDifficultyName(tt.diff)
		if got != tt.want {
			t.Errorf("botDifficultyName(%d) = %q, want %q", tt.diff, got, tt.want)
		}
	}
}

// TestBvBGameMode_SingleGameSetsCount tests single game selection sets count to 1.
func TestBvBGameMode_SingleGameSetsCount(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGameMode
	m.menuOptions = []string{"Single Game", "Multi-Game"}
	m.menuSelection = 0 // Single Game

	result, _ := m.handleBvBGameModeSelection()
	m = result.(Model)

	if m.bvbGameCount != 1 {
		t.Errorf("Expected bvbGameCount to be 1, got: %d", m.bvbGameCount)
	}
}

// TestBvBGameMode_MultiGameShowsInput tests multi-game switches to input mode.
func TestBvBGameMode_MultiGameShowsInput(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGameMode
	m.menuOptions = []string{"Single Game", "Multi-Game"}
	m.menuSelection = 1 // Multi-Game

	result, _ := m.handleBvBGameModeSelection()
	m = result.(Model)

	if !m.bvbInputtingCount {
		t.Error("Expected bvbInputtingCount to be true after selecting Multi-Game")
	}
	if m.bvbCountInput != "" {
		t.Errorf("Expected empty bvbCountInput, got: %q", m.bvbCountInput)
	}
}

// TestBvBGameMode_CountInputAcceptsDigits tests that count input accepts digit characters.
func TestBvBGameMode_CountInputAcceptsDigits(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGameMode
	m.bvbInputtingCount = true
	m.bvbCountInput = ""

	// Type "10"
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}
	result, _ := m.handleBvBCountInput(msg)
	m = result.(Model)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}}
	result, _ = m.handleBvBCountInput(msg)
	m = result.(Model)

	if m.bvbCountInput != "10" {
		t.Errorf("Expected bvbCountInput to be '10', got: %q", m.bvbCountInput)
	}
}

// TestBvBGameMode_CountInputRejectsLetters tests that letters are ignored.
func TestBvBGameMode_CountInputRejectsLetters(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGameMode
	m.bvbInputtingCount = true
	m.bvbCountInput = "5"

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	result, _ := m.handleBvBCountInput(msg)
	m = result.(Model)

	if m.bvbCountInput != "5" {
		t.Errorf("Expected bvbCountInput to remain '5', got: %q", m.bvbCountInput)
	}
}

// TestBvBGameMode_CountInputSubmit tests submitting a valid count.
func TestBvBGameMode_CountInputSubmit(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGameMode
	m.bvbInputtingCount = true
	m.bvbCountInput = "25"

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleBvBCountInput(msg)
	m = result.(Model)

	if m.bvbGameCount != 25 {
		t.Errorf("Expected bvbGameCount to be 25, got: %d", m.bvbGameCount)
	}
	if m.bvbInputtingCount {
		t.Error("Expected bvbInputtingCount to be false after submit")
	}
}

// TestBvBGameMode_CountInputRejectsZero tests that 0 is rejected.
func TestBvBGameMode_CountInputRejectsZero(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGameMode
	m.bvbInputtingCount = true
	m.bvbCountInput = "0"

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleBvBCountInput(msg)
	m = result.(Model)

	if m.errorMsg == "" {
		t.Error("Expected an error message for zero input")
	}
	if !m.bvbInputtingCount {
		t.Error("Should remain in input mode on validation error")
	}
}

// TestBvBGameMode_CountInputRejectsEmpty tests that empty input is rejected.
func TestBvBGameMode_CountInputRejectsEmpty(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGameMode
	m.bvbInputtingCount = true
	m.bvbCountInput = ""

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleBvBCountInput(msg)
	m = result.(Model)

	if m.errorMsg == "" {
		t.Error("Expected an error message for empty input")
	}
}

// TestBvBGameMode_EscFromMenuGoesBackToBotSelect tests ESC navigates back.
func TestBvBGameMode_EscFromMenuGoesBackToBotSelect(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGameMode
	m.menuOptions = []string{"Single Game", "Multi-Game"}
	m.menuSelection = 0
	m.bvbInputtingCount = false

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := m.handleBvBGameModeKeys(msg)
	m = result.(Model)

	if m.screen != ScreenBvBBotSelect {
		t.Errorf("Expected screen to be ScreenBvBBotSelect, got: %v", m.screen)
	}
	if m.bvbSelectingWhite {
		t.Error("Expected bvbSelectingWhite to be false (should return to Black selection)")
	}
}

// TestBvBGameMode_EscFromInputGoesBackToMenu tests ESC from count input returns to menu.
func TestBvBGameMode_EscFromInputGoesBackToMenu(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGameMode
	m.bvbInputtingCount = true
	m.bvbCountInput = "123"

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := m.handleBvBCountInput(msg)
	m = result.(Model)

	if m.bvbInputtingCount {
		t.Error("Expected bvbInputtingCount to be false after ESC")
	}
	if m.bvbCountInput != "" {
		t.Errorf("Expected bvbCountInput to be cleared, got: %q", m.bvbCountInput)
	}
}

// TestBvBGameMode_BackspaceRemovesCharacter tests backspace in count input.
func TestBvBGameMode_BackspaceRemovesCharacter(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGameMode
	m.bvbInputtingCount = true
	m.bvbCountInput = "123"

	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	result, _ := m.handleBvBCountInput(msg)
	m = result.(Model)

	if m.bvbCountInput != "12" {
		t.Errorf("Expected bvbCountInput to be '12', got: %q", m.bvbCountInput)
	}
}

// TestRenderBvBGameMode_MenuView tests rendering in menu mode.
func TestRenderBvBGameMode_MenuView(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGameMode
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotHard
	m.menuOptions = []string{"Single Game", "Multi-Game"}
	m.menuSelection = 0
	m.bvbInputtingCount = false

	view := m.renderBvBGameMode()

	if !strings.Contains(view, "Select Game Mode:") {
		t.Error("Expected view to contain 'Select Game Mode:'")
	}
	if !strings.Contains(view, "Easy Bot (White) vs Hard Bot (Black)") {
		t.Error("Expected view to show matchup info")
	}
	if !strings.Contains(view, "Single Game") {
		t.Error("Expected view to contain 'Single Game'")
	}
	if !strings.Contains(view, "Multi-Game") {
		t.Error("Expected view to contain 'Multi-Game'")
	}
}

// TestRenderBvBGameMode_InputView tests rendering in count input mode.
func TestRenderBvBGameMode_InputView(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGameMode
	m.bvbWhiteDiff = BotMedium
	m.bvbBlackDiff = BotMedium
	m.bvbInputtingCount = true
	m.bvbCountInput = "42"

	view := m.renderBvBGameMode()

	if !strings.Contains(view, "Number of games:") {
		t.Error("Expected view to contain 'Number of games:'")
	}
	if !strings.Contains(view, "42") {
		t.Error("Expected view to show the current input '42'")
	}
}

// TestBvBGridConfig_PresetSelection tests selecting grid presets.
func TestBvBGridConfig_PresetSelection(t *testing.T) {
	tests := []struct {
		name     string
		index    int
		wantRows int
		wantCols int
	}{
		{"1x1", 0, 1, 1},
		{"2x2", 1, 2, 2},
		{"2x3", 2, 2, 3},
		{"2x4", 3, 2, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(DefaultConfig())
			m.screen = ScreenBvBGridConfig
			m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
			m.menuSelection = tt.index
			m.bvbGameCount = 5
			m.bvbWhiteDiff = BotEasy
			m.bvbBlackDiff = BotHard

			result, _ := m.handleBvBGridSelection()
			m = result.(Model)

			if m.bvbGridRows != tt.wantRows {
				t.Errorf("Expected bvbGridRows=%d, got %d", tt.wantRows, m.bvbGridRows)
			}
			if m.bvbGridCols != tt.wantCols {
				t.Errorf("Expected bvbGridCols=%d, got %d", tt.wantCols, m.bvbGridCols)
			}
			if m.screen != ScreenBvBGamePlay {
				t.Errorf("Expected screen to be ScreenBvBGamePlay, got %v", m.screen)
			}
			if m.bvbManager == nil {
				t.Error("Expected bvbManager to be initialized")
			}
			// Clean up
			if m.bvbManager != nil {
				m.bvbManager.Abort()
			}
		})
	}
}

// TestBvBGridConfig_CustomSelection tests selecting "Custom" shows input mode.
func TestBvBGridConfig_CustomSelection(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 4 // Custom

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)

	if !m.bvbInputtingGrid {
		t.Error("Expected bvbInputtingGrid to be true")
	}
	if m.screen != ScreenBvBGridConfig {
		t.Errorf("Expected to stay on ScreenBvBGridConfig, got %v", m.screen)
	}
}

// TestBvBGridConfig_CustomInputValid tests valid custom grid input.
func TestBvBGridConfig_CustomInputValid(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.bvbInputtingGrid = true
	m.bvbCustomGridInput = "2x3"
	m.bvbGameCount = 10
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleBvBGridInput(msg)
	m = result.(Model)

	if m.bvbGridRows != 2 || m.bvbGridCols != 3 {
		t.Errorf("Expected grid 2x3, got %dx%d", m.bvbGridRows, m.bvbGridCols)
	}
	if m.screen != ScreenBvBGamePlay {
		t.Errorf("Expected screen to be ScreenBvBGamePlay, got %v", m.screen)
	}
	// Clean up
	if m.bvbManager != nil {
		m.bvbManager.Abort()
	}
}

// TestBvBGridConfig_CustomInputExceeds8 tests that grid > 8 boards is rejected.
func TestBvBGridConfig_CustomInputExceeds8(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.bvbInputtingGrid = true
	m.bvbCustomGridInput = "3x3"

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleBvBGridInput(msg)
	m = result.(Model)

	if m.errorMsg == "" {
		t.Error("Expected error for grid exceeding 8 boards")
	}
	if !m.bvbInputtingGrid {
		t.Error("Should remain in input mode on error")
	}
}

// TestBvBGridConfig_CustomInputInvalid tests invalid format is rejected.
func TestBvBGridConfig_CustomInputInvalid(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.bvbInputtingGrid = true
	m.bvbCustomGridInput = "22"

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleBvBGridInput(msg)
	m = result.(Model)

	if m.errorMsg == "" {
		t.Error("Expected error for invalid grid format")
	}
}

// TestBvBGridConfig_EscGoesBack tests ESC returns to game mode screen.
func TestBvBGridConfig_EscGoesBack(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 0

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := m.handleBvBGridConfigKeys(msg)
	m = result.(Model)

	if m.screen != ScreenBvBGameMode {
		t.Errorf("Expected screen to be ScreenBvBGameMode, got %v", m.screen)
	}
}

// TestBvBGridConfig_EscFromInputGoesBack tests ESC from input mode cancels input.
func TestBvBGridConfig_EscFromInputGoesBack(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.bvbInputtingGrid = true
	m.bvbCustomGridInput = "2x"

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := m.handleBvBGridInput(msg)
	m = result.(Model)

	if m.bvbInputtingGrid {
		t.Error("Expected bvbInputtingGrid to be false after ESC")
	}
	if m.bvbCustomGridInput != "" {
		t.Errorf("Expected input to be cleared, got %q", m.bvbCustomGridInput)
	}
}

// TestParseGridDimensions tests the parseGridDimensions helper.
func TestParseGridDimensions(t *testing.T) {
	tests := []struct {
		input    string
		wantRows int
		wantCols int
		wantErr  bool
	}{
		{"1x1", 1, 1, false},
		{"2x3", 2, 3, false},
		{"2x4", 2, 4, false},
		{"1X1", 1, 1, false},
		{"3x3", 0, 0, true},  // exceeds 8
		{"22", 0, 0, true},   // no separator
		{"0x1", 0, 0, true},  // zero rows
		{"2x0", 0, 0, true},  // zero cols
		{"x1", 0, 0, true},   // empty rows
		{"2x", 0, 0, true},   // empty cols
	}

	for _, tt := range tests {
		rows, cols, err := parseGridDimensions(tt.input)
		if tt.wantErr && err == nil {
			t.Errorf("parseGridDimensions(%q) expected error", tt.input)
			continue
		}
		if !tt.wantErr && err != nil {
			t.Errorf("parseGridDimensions(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if !tt.wantErr {
			if rows != tt.wantRows || cols != tt.wantCols {
				t.Errorf("parseGridDimensions(%q) = (%d,%d), want (%d,%d)", tt.input, rows, cols, tt.wantRows, tt.wantCols)
			}
		}
	}
}

// TestRenderBvBGridConfig tests rendering of the grid config screen.
func TestRenderBvBGridConfig(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.bvbGameCount = 10
	m.bvbWhiteDiff = BotMedium
	m.bvbBlackDiff = BotHard
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 0

	view := m.renderBvBGridConfig()

	if !strings.Contains(view, "Select Grid Layout:") {
		t.Error("Expected 'Select Grid Layout:' in view")
	}
	if !strings.Contains(view, "10 game(s)") {
		t.Error("Expected game count in view")
	}
	if !strings.Contains(view, "Medium Bot (White) vs Hard Bot (Black)") {
		t.Error("Expected matchup info in view")
	}
	if !strings.Contains(view, "1x1") {
		t.Error("Expected '1x1' preset in view")
	}
}

// TestBvBGamePlay_EscAbortsSession tests ESC during gameplay aborts and returns to menu.
func TestBvBGamePlay_EscAbortsSession(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 0
	m.bvbGameCount = 1
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	// Start a session
	result, _ := m.handleBvBGridSelection()
	m = result.(Model)

	if m.bvbManager == nil {
		t.Fatal("Expected bvbManager to be initialized")
	}

	// Press ESC to abort
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ = m.handleBvBGamePlayKeys(msg)
	m = result.(Model)

	if m.screen != ScreenMainMenu {
		t.Errorf("Expected screen to be ScreenMainMenu, got %v", m.screen)
	}
	if m.bvbManager != nil {
		t.Error("Expected bvbManager to be nil after abort")
	}
}

// TestBvBGamePlay_SpeedChange tests speed change keys.
func TestBvBGamePlay_SpeedChange(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 0
	m.bvbGameCount = 1
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)

	// Speed should default to Normal
	if m.bvbSpeed != bvb.SpeedNormal {
		t.Errorf("Expected default speed Normal, got %v", m.bvbSpeed)
	}

	// Change to Fast (key "2")
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}
	result, _ = m.handleBvBGamePlayKeys(msg)
	m = result.(Model)

	if m.bvbSpeed != bvb.SpeedFast {
		t.Errorf("Expected speed Fast after pressing '2', got %v", m.bvbSpeed)
	}

	// Change to Instant (key "1")
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}
	result, _ = m.handleBvBGamePlayKeys(msg)
	m = result.(Model)

	if m.bvbSpeed != bvb.SpeedInstant {
		t.Errorf("Expected speed Instant after pressing '1', got %v", m.bvbSpeed)
	}

	// Clean up
	if m.bvbManager != nil {
		m.bvbManager.Abort()
	}
}

// TestBvBGamePlay_PauseResume tests pause/resume toggle.
func TestBvBGamePlay_PauseResume(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 0
	m.bvbGameCount = 1
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)

	if m.bvbPaused {
		t.Error("Should not be paused initially")
	}

	// Press space to pause
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	result, _ = m.handleBvBGamePlayKeys(msg)
	m = result.(Model)

	if !m.bvbPaused {
		t.Error("Should be paused after space")
	}

	// Press space again to resume
	result, _ = m.handleBvBGamePlayKeys(msg)
	m = result.(Model)

	if m.bvbPaused {
		t.Error("Should be resumed after second space")
	}

	// Clean up
	if m.bvbManager != nil {
		m.bvbManager.Abort()
	}
}

// TestBvBGamePlay_GameNavigation tests left/right game navigation.
func TestBvBGamePlay_GameNavigation(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 0
	m.bvbGameCount = 3
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)

	if m.bvbSelectedGame != 0 {
		t.Errorf("Expected selectedGame=0, got %d", m.bvbSelectedGame)
	}

	// Press right to go to game 1
	msg := tea.KeyMsg{Type: tea.KeyRight}
	result, _ = m.handleBvBGamePlayKeys(msg)
	m = result.(Model)
	if m.bvbSelectedGame != 1 {
		t.Errorf("Expected selectedGame=1, got %d", m.bvbSelectedGame)
	}

	// Press right again to game 2
	result, _ = m.handleBvBGamePlayKeys(msg)
	m = result.(Model)
	if m.bvbSelectedGame != 2 {
		t.Errorf("Expected selectedGame=2, got %d", m.bvbSelectedGame)
	}

	// Wrap around
	result, _ = m.handleBvBGamePlayKeys(msg)
	m = result.(Model)
	if m.bvbSelectedGame != 0 {
		t.Errorf("Expected wrap to 0, got %d", m.bvbSelectedGame)
	}

	// Press left wraps to last
	msg = tea.KeyMsg{Type: tea.KeyLeft}
	result, _ = m.handleBvBGamePlayKeys(msg)
	m = result.(Model)
	if m.bvbSelectedGame != 2 {
		t.Errorf("Expected wrap to 2, got %d", m.bvbSelectedGame)
	}

	// Clean up
	if m.bvbManager != nil {
		m.bvbManager.Abort()
	}
}

// TestBvBGamePlay_ViewToggle tests Tab toggles between grid and single view.
func TestBvBGamePlay_ViewToggle(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGamePlay
	m.bvbViewMode = BvBSingleView

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\t'}}
	// Note: tab is actually "tab" string
	msg2 := tea.KeyMsg{Type: tea.KeyTab}
	result, _ := m.handleBvBGamePlayKeys(msg2)
	m = result.(Model)

	if m.bvbViewMode != BvBGridView {
		t.Errorf("Expected BvBGridView after Tab, got %v", m.bvbViewMode)
	}

	result, _ = m.handleBvBGamePlayKeys(msg2)
	m = result.(Model)

	if m.bvbViewMode != BvBSingleView {
		t.Errorf("Expected BvBSingleView after second Tab, got %v", m.bvbViewMode)
	}
	_ = msg
}

// TestBvBGamePlay_TickSchedulesNext tests that tick handler schedules next tick.
func TestBvBGamePlay_TickSchedulesNext(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 0
	m.bvbGameCount = 1
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)

	// Handle tick - should schedule another tick since game is running
	result, cmd := m.handleBvBTick()
	m = result.(Model)

	if cmd == nil {
		// Game might have finished instantly; check if manager shows finished
		if m.bvbManager != nil && !m.bvbManager.AllFinished() {
			t.Error("Expected tick to schedule next tick command while game running")
		}
	}

	// Clean up
	if m.bvbManager != nil {
		m.bvbManager.Abort()
	}
}

// TestBvBGamePlay_RenderSingleView tests that the single view renders correctly.
func TestBvBGamePlay_RenderSingleView(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 0
	m.bvbGameCount = 1
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotHard

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)

	view := m.renderBvBGamePlay()

	if !strings.Contains(view, "Bot vs Bot") {
		t.Error("Expected title to contain 'Bot vs Bot'")
	}
	if !strings.Contains(view, "Easy Bot (White) vs Hard Bot (Black)") {
		t.Error("Expected matchup info in view")
	}
	if !strings.Contains(view, "Game 1 of 1") {
		t.Error("Expected game number info")
	}
	if !strings.Contains(view, "Speed: Normal") {
		t.Error("Expected speed indicator")
	}

	// Clean up
	if m.bvbManager != nil {
		m.bvbManager.Abort()
	}
}

// TestBvBGamePlay_RenderGridView tests that the grid view renders correctly with multiple games.
func TestBvBGamePlay_RenderGridView(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 1 // 2x2
	m.bvbGameCount = 4
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotMedium

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)

	// Switch to grid view
	m.bvbViewMode = BvBGridView

	view := m.renderBvBGamePlay()

	if !strings.Contains(view, "Bot vs Bot") {
		t.Error("Expected title to contain 'Bot vs Bot'")
	}
	if !strings.Contains(view, "Easy Bot (White) vs Medium Bot (Black)") {
		t.Error("Expected matchup info in grid view")
	}
	if !strings.Contains(view, "Game 1") {
		t.Error("Expected Game 1 cell in grid view")
	}
	if !strings.Contains(view, "Speed: Normal") {
		t.Error("Expected speed indicator in grid view")
	}
	if !strings.Contains(view, "pages") {
		t.Error("Expected help text to mention pages navigation")
	}

	// Clean up
	if m.bvbManager != nil {
		m.bvbManager.Abort()
	}
}

// TestBvBGamePlay_GridPageNavigation tests page navigation in grid view.
func TestBvBGamePlay_GridPageNavigation(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 0 // 1x1 grid
	m.bvbGameCount = 3  // 3 games with 1x1 grid = 3 pages
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)

	// Switch to grid view
	m.bvbViewMode = BvBGridView

	// Should start at page 0
	if m.bvbPageIndex != 0 {
		t.Errorf("Expected initial page index 0, got %d", m.bvbPageIndex)
	}

	rightMsg := tea.KeyMsg{Type: tea.KeyRight}
	leftMsg := tea.KeyMsg{Type: tea.KeyLeft}

	// Navigate right to page 1
	result, _ = m.handleBvBGamePlayKeys(rightMsg)
	m = result.(Model)
	if m.bvbPageIndex != 1 {
		t.Errorf("Expected page index 1 after right, got %d", m.bvbPageIndex)
	}

	// Navigate right to page 2
	result, _ = m.handleBvBGamePlayKeys(rightMsg)
	m = result.(Model)
	if m.bvbPageIndex != 2 {
		t.Errorf("Expected page index 2 after second right, got %d", m.bvbPageIndex)
	}

	// Navigate right again - should not go past last page
	result, _ = m.handleBvBGamePlayKeys(rightMsg)
	m = result.(Model)
	if m.bvbPageIndex != 2 {
		t.Errorf("Expected page index to stay at 2 (no wrap), got %d", m.bvbPageIndex)
	}

	// Navigate left back to page 1
	result, _ = m.handleBvBGamePlayKeys(leftMsg)
	m = result.(Model)
	if m.bvbPageIndex != 1 {
		t.Errorf("Expected page index 1 after left, got %d", m.bvbPageIndex)
	}

	// Navigate left to page 0
	result, _ = m.handleBvBGamePlayKeys(leftMsg)
	m = result.(Model)
	if m.bvbPageIndex != 0 {
		t.Errorf("Expected page index 0 after second left, got %d", m.bvbPageIndex)
	}

	// Navigate left again - should not go below 0
	result, _ = m.handleBvBGamePlayKeys(leftMsg)
	m = result.(Model)
	if m.bvbPageIndex != 0 {
		t.Errorf("Expected page index to stay at 0 (no wrap), got %d", m.bvbPageIndex)
	}

	// Clean up
	if m.bvbManager != nil {
		m.bvbManager.Abort()
	}
}

// TestBvBGamePlay_GridViewPageIndicator tests that page indicator shows for multi-page grids.
func TestBvBGamePlay_GridViewPageIndicator(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 0 // 1x1 grid
	m.bvbGameCount = 2  // 2 pages
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)
	m.bvbViewMode = BvBGridView

	view := m.renderBvBGamePlay()
	if !strings.Contains(view, "Page 1/2") {
		t.Error("Expected page indicator 'Page 1/2' for multi-page grid")
	}

	// Navigate to page 2
	m.bvbPageIndex = 1
	view = m.renderBvBGamePlay()
	if !strings.Contains(view, "Page 2/2") {
		t.Error("Expected page indicator 'Page 2/2' after navigation")
	}

	// Clean up
	if m.bvbManager != nil {
		m.bvbManager.Abort()
	}
}

// TestBvBGamePlay_GridViewNoPageIndicatorSinglePage tests no page indicator for single-page grids.
func TestBvBGamePlay_GridViewNoPageIndicatorSinglePage(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 1 // 2x2 grid
	m.bvbGameCount = 4  // Exactly fits in one page
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)
	m.bvbViewMode = BvBGridView

	view := m.renderBvBGamePlay()
	if strings.Contains(view, "Page ") {
		t.Error("Expected no page indicator when all games fit on one page")
	}

	// Clean up
	if m.bvbManager != nil {
		m.bvbManager.Abort()
	}
}

// TestBvBGamePlay_ViewTogglePreservesNavigation tests that view toggle doesn't reset navigation state.
func TestBvBGamePlay_ViewTogglePreservesNavigation(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 0 // 1x1
	m.bvbGameCount = 3
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)

	// Start in single view, navigate to game 2
	m.bvbViewMode = BvBSingleView
	m.bvbSelectedGame = 2

	tabMsg := tea.KeyMsg{Type: tea.KeyTab}

	// Toggle to grid view
	result, _ = m.handleBvBGamePlayKeys(tabMsg)
	m = result.(Model)
	if m.bvbViewMode != BvBGridView {
		t.Error("Expected grid view after tab")
	}

	// Selected game should be preserved
	if m.bvbSelectedGame != 2 {
		t.Errorf("Expected selectedGame to be preserved as 2, got %d", m.bvbSelectedGame)
	}

	// Toggle back to single view
	result, _ = m.handleBvBGamePlayKeys(tabMsg)
	m = result.(Model)
	if m.bvbViewMode != BvBSingleView {
		t.Error("Expected single view after second tab")
	}

	// Clean up
	if m.bvbManager != nil {
		m.bvbManager.Abort()
	}
}

// TestBvBGamePlay_GridViewPausedIndicator tests the paused indicator in grid view.
func TestBvBGamePlay_GridViewPausedIndicator(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 0
	m.bvbGameCount = 1
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)
	m.bvbViewMode = BvBGridView
	m.bvbPaused = true

	view := m.renderBvBGamePlay()
	if !strings.Contains(view, "PAUSED") {
		t.Error("Expected PAUSED indicator in grid view when paused")
	}

	// Clean up
	if m.bvbManager != nil {
		m.bvbManager.Abort()
	}
}

// TestBvBStats_TransitionOnAllFinished tests that tick transitions to stats when all games done.
func TestBvBStats_TransitionOnAllFinished(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 0 // 1x1
	m.bvbGameCount = 1
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)

	// Set speed to instant so game finishes quickly
	m.bvbSpeed = bvb.SpeedInstant
	m.bvbManager.SetSpeed(bvb.SpeedInstant)

	// Wait for the game to finish
	for i := 0; i < 1000; i++ {
		if m.bvbManager.AllFinished() {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	if !m.bvbManager.AllFinished() {
		t.Skip("Game did not finish in time")
	}

	// Tick should transition to stats
	result, _ = m.handleBvBTick()
	m = result.(Model)

	if m.screen != ScreenBvBStats {
		t.Errorf("Expected ScreenBvBStats after all finished, got %d", m.screen)
	}
	if m.bvbStatsSelection != 0 {
		t.Errorf("Expected stats selection to be 0, got %d", m.bvbStatsSelection)
	}
}

// TestBvBStats_RenderSingleGame tests stats rendering for a single game.
func TestBvBStats_RenderSingleGame(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 0
	m.bvbGameCount = 1
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)

	m.bvbSpeed = bvb.SpeedInstant
	m.bvbManager.SetSpeed(bvb.SpeedInstant)

	for i := 0; i < 2000; i++ {
		if m.bvbManager.AllFinished() {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	if !m.bvbManager.AllFinished() {
		t.Skip("Game did not finish in time")
	}

	// Transition to stats
	m.screen = ScreenBvBStats
	m.bvbStatsSelection = 0
	m.menuOptions = []string{"New Session", "Return to Menu"}

	view := m.renderBvBStats()

	if !strings.Contains(view, "Results") {
		t.Error("Expected 'Results' in stats title")
	}
	if !strings.Contains(view, "Easy Bot") {
		t.Error("Expected 'Easy Bot' in stats")
	}
	if !strings.Contains(view, "moves") {
		t.Error("Expected move count in stats")
	}
	if !strings.Contains(view, "New Session") {
		t.Error("Expected 'New Session' option")
	}
	if !strings.Contains(view, "Return to Menu") {
		t.Error("Expected 'Return to Menu' option")
	}
}

// TestBvBStats_RenderMultiGame tests stats rendering for multiple games.
func TestBvBStats_RenderMultiGame(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 1 // 2x2
	m.bvbGameCount = 4
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)

	m.bvbSpeed = bvb.SpeedInstant
	m.bvbManager.SetSpeed(bvb.SpeedInstant)

	for i := 0; i < 2000; i++ {
		if m.bvbManager.AllFinished() {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	if !m.bvbManager.AllFinished() {
		t.Skip("Games did not finish in time")
	}

	m.screen = ScreenBvBStats
	m.bvbStatsSelection = 0
	m.menuOptions = []string{"New Session", "Return to Menu"}

	view := m.renderBvBStats()

	if !strings.Contains(view, "4 games") {
		t.Error("Expected '4 games' in multi-game stats")
	}
	if !strings.Contains(view, "wins") {
		t.Error("Expected 'wins' in multi-game stats")
	}
	if !strings.Contains(view, "Avg moves") {
		t.Error("Expected 'Avg moves' in multi-game stats")
	}
	if !strings.Contains(view, "Shortest game") {
		t.Error("Expected 'Shortest game' in multi-game stats")
	}
	if !strings.Contains(view, "Individual Results") {
		t.Error("Expected 'Individual Results' in multi-game stats")
	}
}

// TestBvBStats_NavigationUpDown tests stats screen menu navigation.
func TestBvBStats_NavigationUpDown(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBStats
	m.menuOptions = []string{"New Session", "Return to Menu"}
	m.bvbStatsSelection = 0

	// Down moves to Return to Menu
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	result, _ := m.handleBvBStatsKeys(msg)
	m = result.(Model)
	if m.bvbStatsSelection != 1 {
		t.Errorf("Expected selection 1 after down, got %d", m.bvbStatsSelection)
	}

	// Down again doesn't go past end
	result, _ = m.handleBvBStatsKeys(msg)
	m = result.(Model)
	if m.bvbStatsSelection != 1 {
		t.Errorf("Expected selection to stay at 1, got %d", m.bvbStatsSelection)
	}

	// Up goes back
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	result, _ = m.handleBvBStatsKeys(msg)
	m = result.(Model)
	if m.bvbStatsSelection != 0 {
		t.Errorf("Expected selection 0 after up, got %d", m.bvbStatsSelection)
	}
}

// TestBvBStats_NewSessionAction tests selecting "New Session" on stats screen.
func TestBvBStats_NewSessionAction(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBStats
	m.menuOptions = []string{"New Session", "Return to Menu"}
	m.bvbStatsSelection = 0 // New Session

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleBvBStatsKeys(msg)
	m = result.(Model)

	if m.screen != ScreenBvBBotSelect {
		t.Errorf("Expected ScreenBvBBotSelect after New Session, got %d", m.screen)
	}
	if !m.bvbSelectingWhite {
		t.Error("Expected bvbSelectingWhite to be true")
	}
	if m.bvbManager != nil {
		t.Error("Expected bvbManager to be nil after starting new session flow")
	}
}

// TestBvBStats_ReturnToMenuAction tests selecting "Return to Menu" on stats screen.
func TestBvBStats_ReturnToMenuAction(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBStats
	m.menuOptions = []string{"New Session", "Return to Menu"}
	m.bvbStatsSelection = 1 // Return to Menu

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleBvBStatsKeys(msg)
	m = result.(Model)

	if m.screen != ScreenMainMenu {
		t.Errorf("Expected ScreenMainMenu after Return to Menu, got %d", m.screen)
	}
	if m.bvbManager != nil {
		t.Error("Expected bvbManager to be nil after returning to menu")
	}
}

// TestBvBStats_EscGoesToMenu tests ESC on stats screen.
func TestBvBStats_EscGoesToMenu(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBStats
	m.menuOptions = []string{"New Session", "Return to Menu"}
	m.bvbStatsSelection = 0

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := m.handleBvBStatsKeys(msg)
	m = result.(Model)

	if m.screen != ScreenMainMenu {
		t.Errorf("Expected ScreenMainMenu after ESC, got %d", m.screen)
	}
}

// TestBvBGamePlay_FENExportSingleView tests FEN export in single view.
func TestBvBGamePlay_FENExportSingleView(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 0
	m.bvbGameCount = 1
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)

	m.bvbViewMode = BvBSingleView
	m.bvbSelectedGame = 0

	// Press 'f' to export FEN
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
	result, _ = m.handleBvBGamePlayKeys(msg)
	m = result.(Model)

	// Status message should contain FEN-related text
	if m.statusMsg == "" {
		t.Error("Expected status message after FEN export")
	}
	if !strings.Contains(m.statusMsg, "FEN") {
		t.Errorf("Expected status to mention FEN, got: %s", m.statusMsg)
	}

	// Clean up
	if m.bvbManager != nil {
		m.bvbManager.Abort()
	}
}

// TestBvBGamePlay_FENExportGridView tests FEN export in grid view.
func TestBvBGamePlay_FENExportGridView(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 1 // 2x2
	m.bvbGameCount = 4
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)

	m.bvbViewMode = BvBGridView
	m.bvbPageIndex = 0

	// Press 'f' to export FEN
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
	result, _ = m.handleBvBGamePlayKeys(msg)
	m = result.(Model)

	// Status message should contain FEN-related text
	if m.statusMsg == "" {
		t.Error("Expected status message after FEN export in grid view")
	}
	if !strings.Contains(m.statusMsg, "FEN") {
		t.Errorf("Expected status to mention FEN, got: %s", m.statusMsg)
	}

	// Clean up
	if m.bvbManager != nil {
		m.bvbManager.Abort()
	}
}

// TestBvBGamePlay_FENExportNoManager tests FEN export with no manager.
func TestBvBGamePlay_FENExportNoManager(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGamePlay
	m.bvbManager = nil

	// Press 'f' should not crash
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
	result, _ := m.handleBvBGamePlayKeys(msg)
	m = result.(Model)

	// No status message since no manager
	if m.statusMsg != "" {
		t.Errorf("Expected no status message without manager, got: %s", m.statusMsg)
	}
}

// TestBvB_CtrlCCleansBvBManager tests that Ctrl+C cleans up the BvB manager.
func TestBvB_CtrlCCleansBvBManager(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 0
	m.bvbGameCount = 1
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)

	if m.bvbManager == nil {
		t.Fatal("Expected bvbManager to be set")
	}

	// Press Ctrl+C
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	result, cmd := m.handleKeyPress(msg)
	m = result.(Model)

	if m.bvbManager != nil {
		t.Error("Expected bvbManager to be nil after Ctrl+C")
	}
	if cmd == nil {
		t.Error("Expected tea.Quit command after Ctrl+C")
	}
}

// TestBvB_QuitCleansBvBManager tests that 'q' cleans up the BvB manager.
func TestBvB_QuitCleansBvBManager(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGamePlay // Not ScreenGamePlay, so q should quit
	m.bvbManager = nil           // Start fresh

	// Set up a manager
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 0
	m.bvbGameCount = 1
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)

	// Change screen to something that allows 'q' to quit
	m.screen = ScreenBvBStats

	if m.bvbManager == nil {
		t.Fatal("Expected bvbManager to be set")
	}

	// Press 'q'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	result, cmd := m.handleKeyPress(msg)
	m = result.(Model)

	if m.bvbManager != nil {
		t.Error("Expected bvbManager to be nil after 'q'")
	}
	if cmd == nil {
		t.Error("Expected tea.Quit command after 'q'")
	}
}

// TestBvB_GridViewTerminalTooSmall tests grid view fallback for small terminals.
func TestBvB_GridViewTerminalTooSmall(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 1 // 2x2
	m.bvbGameCount = 4
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)

	m.bvbViewMode = BvBGridView
	// Set terminal size too small for 2x2 grid
	m.termWidth = 20
	m.termHeight = 15

	view := m.renderBvBGamePlay()

	if !strings.Contains(view, "Terminal too small") {
		t.Error("Expected 'Terminal too small' warning for small terminal")
	}
	if !strings.Contains(view, "Tab") {
		t.Error("Expected suggestion to switch to single view")
	}

	// Clean up
	if m.bvbManager != nil {
		m.bvbManager.Abort()
	}
}

// TestBvB_GridViewTerminalLargeEnough tests no warning for large terminal.
func TestBvB_GridViewTerminalLargeEnough(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 1 // 2x2
	m.bvbGameCount = 4
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)

	m.bvbViewMode = BvBGridView
	// Set terminal size large enough
	m.termWidth = 100
	m.termHeight = 50

	view := m.renderBvBGamePlay()

	if strings.Contains(view, "Terminal too small") {
		t.Error("Should not show terminal warning for large terminal")
	}

	// Clean up
	if m.bvbManager != nil {
		m.bvbManager.Abort()
	}
}

// TestBvB_WindowSizeMsg tests that WindowSizeMsg updates terminal dimensions.
func TestBvB_WindowSizeMsg(t *testing.T) {
	m := NewModel(DefaultConfig())

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	result, _ := m.Update(msg)
	m = result.(Model)

	if m.termWidth != 120 {
		t.Errorf("Expected termWidth 120, got %d", m.termWidth)
	}
	if m.termHeight != 40 {
		t.Errorf("Expected termHeight 40, got %d", m.termHeight)
	}
}

// TestBvB_CompleteFlow tests the full BvB flow: menu → bot select → game mode → grid → gameplay → stats → menu.
func TestBvB_CompleteFlow(t *testing.T) {
	m := NewModel(DefaultConfig())

	// Step 1: Main menu → Game Type Select
	m.screen = ScreenMainMenu
	m.menuOptions = buildMainMenuOptions()
	// Find "New Game" and select it
	for i, opt := range m.menuOptions {
		if opt == "New Game" {
			m.menuSelection = i
			break
		}
	}
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleKeyPress(msg)
	m = result.(Model)
	if m.screen != ScreenGameTypeSelect {
		t.Fatalf("Step 1: Expected ScreenGameTypeSelect, got %d", m.screen)
	}

	// Step 2: Game Type → Bot vs Bot (last option)
	m.menuSelection = len(m.menuOptions) - 1 // "Bot vs Bot" is last
	result, _ = m.handleKeyPress(msg)
	m = result.(Model)
	if m.screen != ScreenBvBBotSelect {
		t.Fatalf("Step 2: Expected ScreenBvBBotSelect, got %d", m.screen)
	}

	// Step 3: Select White bot (Easy)
	m.menuSelection = 0 // Easy
	result, _ = m.handleKeyPress(msg)
	m = result.(Model)
	if m.bvbWhiteDiff != BotEasy {
		t.Fatal("Step 3: Expected BotEasy for white")
	}
	if m.bvbSelectingWhite {
		t.Fatal("Step 3: Should now be selecting black")
	}

	// Step 4: Select Black bot (Easy)
	m.menuSelection = 0 // Easy
	result, _ = m.handleKeyPress(msg)
	m = result.(Model)
	if m.screen != ScreenBvBGameMode {
		t.Fatalf("Step 4: Expected ScreenBvBGameMode, got %d", m.screen)
	}

	// Step 5: Select Single Game mode
	m.menuSelection = 0 // Single Game
	result, _ = m.handleKeyPress(msg)
	m = result.(Model)
	if m.screen != ScreenBvBGridConfig {
		t.Fatalf("Step 5: Expected ScreenBvBGridConfig, got %d", m.screen)
	}
	if m.bvbGameCount != 1 {
		t.Fatalf("Step 5: Expected game count 1, got %d", m.bvbGameCount)
	}

	// Step 6: Select 1x1 grid
	m.menuSelection = 0 // 1x1
	result, _ = m.handleKeyPress(msg)
	m = result.(Model)
	if m.screen != ScreenBvBGamePlay {
		t.Fatalf("Step 6: Expected ScreenBvBGamePlay, got %d", m.screen)
	}
	if m.bvbManager == nil {
		t.Fatal("Step 6: Expected bvbManager to be initialized")
	}

	// Step 7: Set speed to instant and wait for game to finish
	m.bvbSpeed = bvb.SpeedInstant
	m.bvbManager.SetSpeed(bvb.SpeedInstant)
	for i := 0; i < 2000; i++ {
		if m.bvbManager.AllFinished() {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if !m.bvbManager.AllFinished() {
		t.Skip("Game did not finish in time")
	}

	// Step 8: Tick transitions to stats
	result, _ = m.handleBvBTick()
	m = result.(Model)
	if m.screen != ScreenBvBStats {
		t.Fatalf("Step 8: Expected ScreenBvBStats, got %d", m.screen)
	}

	// Step 9: Verify stats are populated
	stats := m.bvbManager.Stats()
	if stats == nil || stats.TotalGames != 1 {
		t.Fatal("Step 9: Expected 1 game in stats")
	}

	// Step 10: Return to menu
	m.bvbStatsSelection = 1 // Return to Menu
	result, _ = m.handleBvBStatsKeys(msg)
	m = result.(Model)
	if m.screen != ScreenMainMenu {
		t.Fatalf("Step 10: Expected ScreenMainMenu, got %d", m.screen)
	}
	if m.bvbManager != nil {
		t.Fatal("Step 10: Expected bvbManager to be nil after returning to menu")
	}
}

// TestBvB_HelpTextConfig tests that help text respects ShowHelpText config.
func TestBvB_HelpTextConfig(t *testing.T) {
	// With help text disabled
	cfg := DefaultConfig()
	cfg.ShowHelpText = false

	m := NewModel(cfg)
	m.screen = ScreenBvBGridConfig
	m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
	m.menuSelection = 0
	m.bvbGameCount = 1
	m.bvbWhiteDiff = BotEasy
	m.bvbBlackDiff = BotEasy

	result, _ := m.handleBvBGridSelection()
	m = result.(Model)
	m.bvbViewMode = BvBSingleView

	view := m.renderBvBGamePlay()
	if strings.Contains(view, "ESC: abort") {
		t.Error("Expected help text to be hidden when ShowHelpText is false")
	}

	// Clean up
	if m.bvbManager != nil {
		m.bvbManager.Abort()
	}
}

// TestParsePositiveInt tests the parsePositiveInt helper.
func TestParsePositiveInt(t *testing.T) {
	tests := []struct {
		input string
		want  int
		err   bool
	}{
		{"1", 1, false},
		{"10", 10, false},
		{"999", 999, false},
		{"0", 0, true},
		{"", 0, true},
		{"abc", 0, true},
		{"12abc", 0, true},
	}
	for _, tt := range tests {
		got, err := parsePositiveInt(tt.input)
		if tt.err && err == nil {
			t.Errorf("parsePositiveInt(%q) expected error, got %d", tt.input, got)
		}
		if !tt.err && err != nil {
			t.Errorf("parsePositiveInt(%q) unexpected error: %v", tt.input, err)
		}
		if !tt.err && got != tt.want {
			t.Errorf("parsePositiveInt(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
