package ui

import (
	"testing"

	"github.com/Mgrdich/TermChess/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

// TestESCKeyGameTypeSelectToMainMenu tests ESC key navigation from GameTypeSelect to MainMenu
func TestESCKeyGameTypeSelectToMainMenu(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenGameTypeSelect
	m.menuOptions = []string{"Player vs Player", "Player vs Bot"}
	m.menuSelection = 1

	// Press ESC key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := m.handleGameTypeSelectKeys(msg)
	m = result.(Model)

	// Verify returned to main menu
	if m.screen != ScreenMainMenu {
		t.Errorf("Expected screen to be ScreenMainMenu, got %v", m.screen)
	}

	// Verify menu was reset to main menu options
	expectedOptions := []string{"New Game", "Load Game", "Settings", "Exit"}
	if len(m.menuOptions) != len(expectedOptions) {
		t.Errorf("Expected %d menu options, got %d", len(expectedOptions), len(m.menuOptions))
	}

	// Verify menu selection was reset
	if m.menuSelection != 0 {
		t.Errorf("Expected menuSelection to be reset to 0, got %d", m.menuSelection)
	}

	// Verify messages are cleared
	if m.errorMsg != "" {
		t.Errorf("Expected errorMsg to be cleared, got '%s'", m.errorMsg)
	}
	if m.statusMsg != "" {
		t.Errorf("Expected statusMsg to be cleared, got '%s'", m.statusMsg)
	}
}

// TestESCKeyFENInputToMainMenu tests ESC key navigation from FENInput to MainMenu
func TestESCKeyFENInputToMainMenu(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenFENInput
	m.fenInput.SetValue("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

	// Press ESC key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := m.handleFENInputKeys(msg)
	m = result.(Model)

	// Verify returned to main menu
	if m.screen != ScreenMainMenu {
		t.Errorf("Expected screen to be ScreenMainMenu, got %v", m.screen)
	}

	// Verify menu was reset to main menu options
	expectedOptions := []string{"New Game", "Load Game", "Settings", "Exit"}
	if len(m.menuOptions) != len(expectedOptions) {
		t.Errorf("Expected %d menu options, got %d", len(expectedOptions), len(m.menuOptions))
	}

	// Verify FEN input was cleared
	if m.fenInput.Value() != "" {
		t.Errorf("Expected fenInput to be cleared, got '%s'", m.fenInput.Value())
	}

	// Verify messages are cleared
	if m.errorMsg != "" {
		t.Errorf("Expected errorMsg to be cleared, got '%s'", m.errorMsg)
	}
	if m.statusMsg != "" {
		t.Errorf("Expected statusMsg to be cleared, got '%s'", m.statusMsg)
	}
}

// TestESCKeySettingsToMainMenu tests ESC key navigation from Settings to MainMenu
func TestESCKeySettingsToMainMenu(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenSettings
	m.settingsSelection = 2

	// Press ESC key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := m.handleSettingsKeys(msg)
	m = result.(Model)

	// Verify returned to main menu
	if m.screen != ScreenMainMenu {
		t.Errorf("Expected screen to be ScreenMainMenu, got %v", m.screen)
	}

	// Verify menu was reset
	if m.menuSelection != 0 {
		t.Errorf("Expected menuSelection to be reset to 0, got %d", m.menuSelection)
	}

	expectedOptions := []string{"New Game", "Load Game", "Settings", "Exit"}
	if len(m.menuOptions) != len(expectedOptions) {
		t.Errorf("Expected %d menu options, got %d", len(expectedOptions), len(m.menuOptions))
	}
}

// TestESCKeyGamePlayShowsSavePrompt tests ESC key shows save prompt during active gameplay
func TestESCKeyGamePlayShowsSavePrompt(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay
	m.input = "e4"

	// Press ESC key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := m.handleGamePlayKeys(msg)
	m = result.(Model)

	// Verify transitioned to save prompt
	if m.screen != ScreenSavePrompt {
		t.Errorf("Expected screen to be ScreenSavePrompt, got %v", m.screen)
	}

	// Verify save prompt action is set to "menu"
	if m.savePromptAction != "menu" {
		t.Errorf("Expected savePromptAction to be 'menu', got '%s'", m.savePromptAction)
	}

	// Verify save prompt selection is reset
	if m.savePromptSelection != 0 {
		t.Errorf("Expected savePromptSelection to be 0, got %d", m.savePromptSelection)
	}

	// Verify messages are cleared
	if m.errorMsg != "" {
		t.Errorf("Expected errorMsg to be cleared, got '%s'", m.errorMsg)
	}
	if m.statusMsg != "" {
		t.Errorf("Expected statusMsg to be cleared, got '%s'", m.statusMsg)
	}
}

// TestESCKeySavePromptReturnsToGamePlay tests ESC key returns to gameplay from save prompt
func TestESCKeySavePromptReturnsToGamePlay(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenSavePrompt
	m.savePromptSelection = 1
	m.savePromptAction = "menu"

	// Press ESC key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := m.handleSavePromptKeys(msg)
	m = result.(Model)

	// Verify returned to gameplay
	if m.screen != ScreenGamePlay {
		t.Errorf("Expected screen to be ScreenGamePlay, got %v", m.screen)
	}

	// Verify messages are cleared
	if m.errorMsg != "" {
		t.Errorf("Expected errorMsg to be cleared, got '%s'", m.errorMsg)
	}
	if m.statusMsg != "" {
		t.Errorf("Expected statusMsg to be cleared, got '%s'", m.statusMsg)
	}
}

// TestESCKeyGameOverToMainMenu tests ESC key navigation from GameOver to MainMenu
func TestESCKeyGameOverToMainMenu(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGameOver
	m.resignedBy = int8(engine.White)

	// Press ESC key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := m.handleGameOverKeys(msg)
	m = result.(Model)

	// Verify returned to main menu
	if m.screen != ScreenMainMenu {
		t.Errorf("Expected screen to be ScreenMainMenu, got %v", m.screen)
	}

	// Verify board was cleared
	if m.board != nil {
		t.Error("Expected board to be nil after returning to main menu")
	}

	// Verify move history was cleared
	if len(m.moveHistory) != 0 {
		t.Errorf("Expected moveHistory to be empty, got %d moves", len(m.moveHistory))
	}

	// Verify menu was reset
	expectedOptions := []string{"New Game", "Load Game", "Settings", "Exit"}
	if len(m.menuOptions) != len(expectedOptions) {
		t.Errorf("Expected %d menu options, got %d", len(expectedOptions), len(m.menuOptions))
	}

	if m.menuSelection != 0 {
		t.Errorf("Expected menuSelection to be reset to 0, got %d", m.menuSelection)
	}
}

// TestESCKeyDrawPromptReturnsToGamePlay tests ESC key returns to gameplay from draw prompt
func TestESCKeyDrawPromptReturnsToGamePlay(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenDrawPrompt
	m.drawOfferedBy = int8(engine.White)
	m.drawOfferedByWhite = true
	m.drawPromptSelection = 0

	// Press ESC key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := m.handleDrawPromptKeys(msg)
	m = result.(Model)

	// Verify returned to gameplay
	if m.screen != ScreenGamePlay {
		t.Errorf("Expected screen to be ScreenGamePlay, got %v", m.screen)
	}

	// Verify draw offer state was reset
	if m.drawOfferedBy != -1 {
		t.Errorf("Expected drawOfferedBy to be reset to -1, got %d", m.drawOfferedBy)
	}

	if m.drawOfferedByWhite {
		t.Error("Expected drawOfferedByWhite to be false after canceling draw offer")
	}

	// Verify status message was set
	if m.statusMsg != "Draw offer cancelled" {
		t.Errorf("Expected statusMsg to be 'Draw offer cancelled', got '%s'", m.statusMsg)
	}
}

// TestNavigationFlowNewGameToMenu tests full navigation flow from main menu through new game and back
func TestNavigationFlowNewGameToMenu(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenMainMenu
	m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
	m.menuSelection = 0

	// Select "New Game"
	result, _ := m.handleMainMenuSelection()
	m = result.(Model)

	// Should be at GameTypeSelect
	if m.screen != ScreenGameTypeSelect {
		t.Fatalf("Expected screen to be ScreenGameTypeSelect, got %v", m.screen)
	}

	// Press ESC to go back to main menu
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ = m.handleGameTypeSelectKeys(msg)
	m = result.(Model)

	// Should be back at main menu
	if m.screen != ScreenMainMenu {
		t.Errorf("Expected screen to be ScreenMainMenu, got %v", m.screen)
	}
}

// TestNavigationFlowLoadGameToMenu tests navigation flow from Load Game back to menu
func TestNavigationFlowLoadGameToMenu(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenMainMenu
	m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
	m.menuSelection = 1

	// Select "Load Game"
	result, _ := m.handleMainMenuSelection()
	m = result.(Model)

	// Should be at FENInput
	if m.screen != ScreenFENInput {
		t.Fatalf("Expected screen to be ScreenFENInput, got %v", m.screen)
	}

	// Press ESC to go back to main menu
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ = m.handleFENInputKeys(msg)
	m = result.(Model)

	// Should be back at main menu
	if m.screen != ScreenMainMenu {
		t.Errorf("Expected screen to be ScreenMainMenu, got %v", m.screen)
	}
}

// TestNavigationFlowSettingsToMenu tests navigation flow from Settings back to menu
func TestNavigationFlowSettingsToMenu(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenMainMenu
	m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
	m.menuSelection = 2

	// Select "Settings"
	result, _ := m.handleMainMenuSelection()
	m = result.(Model)

	// Should be at Settings
	if m.screen != ScreenSettings {
		t.Fatalf("Expected screen to be ScreenSettings, got %v", m.screen)
	}

	// Press ESC to go back to main menu
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ = m.handleSettingsKeys(msg)
	m = result.(Model)

	// Should be back at main menu
	if m.screen != ScreenMainMenu {
		t.Errorf("Expected screen to be ScreenMainMenu, got %v", m.screen)
	}
}

// TestCtrlCExitsFromAnyScreen tests that Ctrl+C exits the application from any screen
func TestCtrlCExitsFromAnyScreen(t *testing.T) {
	screens := []Screen{
		ScreenMainMenu,
		ScreenGameTypeSelect,
		ScreenFENInput,
		ScreenSettings,
		ScreenGameOver,
	}

	for _, screen := range screens {
		m := NewModel(DefaultConfig())
		m.screen = screen
		if screen == ScreenGameOver || screen == ScreenGamePlay {
			m.board = engine.NewBoard()
		}
		if screen == ScreenGameTypeSelect {
			m.menuOptions = []string{"Player vs Player", "Player vs Bot"}
		}

		// Press Ctrl+C
		msg := tea.KeyMsg{Type: tea.KeyCtrlC}
		result, cmd := m.handleKeyPress(msg)

		// Verify quit command was returned
		if cmd == nil {
			t.Errorf("Expected quit command for screen %v, got nil", screen)
		}

		// The model should still be valid (quit happens via command)
		if _, ok := result.(Model); !ok {
			t.Errorf("Expected result to be a Model for screen %v", screen)
		}
	}
}

// TestQKeyBehaviorDiffersByScreen tests that 'q' key behaves differently based on screen
func TestQKeyBehaviorDiffersByScreen(t *testing.T) {
	// 'q' should quit from main menu
	m := NewModel(DefaultConfig())
	m.screen = ScreenMainMenu

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.handleKeyPress(msg)

	if cmd == nil {
		t.Error("Expected quit command from main menu when pressing 'q'")
	}

	// 'q' should show save prompt from GamePlay
	m = NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	result, cmd := m.handleKeyPress(msg)
	m = result.(Model)

	if m.screen != ScreenSavePrompt {
		t.Errorf("Expected screen to be ScreenSavePrompt after pressing 'q' in GamePlay, got %v", m.screen)
	}

	if cmd != nil {
		t.Error("Did not expect quit command from GamePlay when pressing 'q' (should show save prompt)")
	}
}
