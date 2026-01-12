package ui

import (
	"strings"
	"testing"

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

// TestGameTypeSelection_BotNotImplemented tests that bot selection shows coming soon message
func TestGameTypeSelection_BotNotImplemented(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenGameTypeSelect
	m.menuOptions = []string{"Player vs Player", "Player vs Bot", "Back"}
	m.menuSelection = 1 // Select "Player vs Bot"

	result, _ := m.handleGameTypeSelection()
	m = result.(Model)

	// Should show "coming soon" message
	if !strings.Contains(strings.ToLower(m.statusMsg), "coming soon") {
		t.Errorf("Expected 'coming soon' message for bot play, got: %s", m.statusMsg)
	}

	// Should set game type to PvBot even though not implemented
	if m.gameType != GameTypePvBot {
		t.Errorf("Expected gameType to be set to PvBot, got: %v", m.gameType)
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
