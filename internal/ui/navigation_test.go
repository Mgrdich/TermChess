package ui

import (
	"strings"
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

// TestScreenName tests the screenName helper function
func TestScreenName(t *testing.T) {
	tests := []struct {
		screen   Screen
		expected string
	}{
		{ScreenMainMenu, "Main Menu"},
		{ScreenGameTypeSelect, "New Game"},
		{ScreenBotSelect, "Bot Difficulty"},
		{ScreenColorSelect, "Choose Color"},
		{ScreenFENInput, "Load Game"},
		{ScreenSettings, "Settings"},
		{ScreenGamePlay, "Game"},
		{ScreenGameOver, "Game Over"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := screenName(tt.screen)
			if got != tt.expected {
				t.Errorf("screenName(%v) = %q, want %q", tt.screen, got, tt.expected)
			}
		})
	}
}

// TestPushScreen tests the pushScreen method
func TestPushScreen(t *testing.T) {
	m := NewModel(DefaultConfig())

	// Initially at MainMenu with empty stack
	if m.screen != ScreenMainMenu {
		t.Errorf("Expected initial screen to be MainMenu, got %v", m.screen)
	}
	if len(m.navStack) != 0 {
		t.Errorf("Expected empty nav stack, got %d items", len(m.navStack))
	}

	// Push to GameTypeSelect
	m.pushScreen(ScreenGameTypeSelect)
	if m.screen != ScreenGameTypeSelect {
		t.Errorf("Expected screen to be GameTypeSelect, got %v", m.screen)
	}
	if len(m.navStack) != 1 {
		t.Errorf("Expected 1 item in nav stack, got %d", len(m.navStack))
	}
	if m.navStack[0] != ScreenMainMenu {
		t.Errorf("Expected MainMenu in nav stack, got %v", m.navStack[0])
	}

	// Push to BotSelect
	m.pushScreen(ScreenBotSelect)
	if m.screen != ScreenBotSelect {
		t.Errorf("Expected screen to be BotSelect, got %v", m.screen)
	}
	if len(m.navStack) != 2 {
		t.Errorf("Expected 2 items in nav stack, got %d", len(m.navStack))
	}

	// Pushing same screen should not add to stack
	m.pushScreen(ScreenBotSelect)
	if len(m.navStack) != 2 {
		t.Errorf("Expected stack to remain at 2 items, got %d", len(m.navStack))
	}
}

// TestPopScreen tests the popScreen method
func TestPopScreen(t *testing.T) {
	m := NewModel(DefaultConfig())

	// Set up a navigation stack
	m.navStack = []Screen{ScreenMainMenu, ScreenGameTypeSelect}
	m.screen = ScreenBotSelect

	// Pop should return to GameTypeSelect
	result := m.popScreen()
	if result != ScreenGameTypeSelect {
		t.Errorf("popScreen() returned %v, expected GameTypeSelect", result)
	}
	if m.screen != ScreenGameTypeSelect {
		t.Errorf("Expected screen to be GameTypeSelect, got %v", m.screen)
	}
	if len(m.navStack) != 1 {
		t.Errorf("Expected 1 item in nav stack, got %d", len(m.navStack))
	}

	// Pop again should return to MainMenu
	result = m.popScreen()
	if result != ScreenMainMenu {
		t.Errorf("popScreen() returned %v, expected MainMenu", result)
	}
	if m.screen != ScreenMainMenu {
		t.Errorf("Expected screen to be MainMenu, got %v", m.screen)
	}
	if len(m.navStack) != 0 {
		t.Errorf("Expected empty nav stack, got %d items", len(m.navStack))
	}

	// Pop with empty stack should go to MainMenu
	m.screen = ScreenSettings
	result = m.popScreen()
	if result != ScreenMainMenu {
		t.Errorf("popScreen() with empty stack returned %v, expected MainMenu", result)
	}
}

// TestClearNavStack tests the clearNavStack method
func TestClearNavStack(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.navStack = []Screen{ScreenMainMenu, ScreenGameTypeSelect, ScreenBotSelect}

	m.clearNavStack()

	if len(m.navStack) != 0 {
		t.Errorf("Expected empty nav stack after clear, got %d items", len(m.navStack))
	}
}

// TestBreadcrumb tests the breadcrumb method
func TestBreadcrumb(t *testing.T) {
	m := NewModel(DefaultConfig())

	// At MainMenu with empty stack - should return empty
	bc := m.breadcrumb()
	if bc != "" {
		t.Errorf("Expected empty breadcrumb at MainMenu, got %q", bc)
	}

	// Push to Settings - should show "Main Menu > Settings"
	m.pushScreen(ScreenSettings)
	bc = m.breadcrumb()
	expected := "Main Menu > Settings"
	if bc != expected {
		t.Errorf("Expected breadcrumb %q, got %q", expected, bc)
	}

	// Push to GameTypeSelect from MainMenu, then BotSelect
	m = NewModel(DefaultConfig())
	m.pushScreen(ScreenGameTypeSelect)
	m.pushScreen(ScreenBotSelect)
	bc = m.breadcrumb()
	expected = "New Game > Bot Difficulty"
	if bc != expected {
		t.Errorf("Expected breadcrumb %q, got %q", expected, bc)
	}
}

// TestCanGoBack tests the canGoBack method
func TestCanGoBack(t *testing.T) {
	m := NewModel(DefaultConfig())

	if m.canGoBack() {
		t.Error("Expected canGoBack() to return false with empty stack")
	}

	m.pushScreen(ScreenSettings)

	if !m.canGoBack() {
		t.Error("Expected canGoBack() to return true with non-empty stack")
	}
}

// TestNavStackClearedOnGameStart tests that nav stack is cleared when starting a game
func TestNavStackClearedOnGameStart(t *testing.T) {
	m := NewModel(DefaultConfig())

	// Navigate to game type select
	m.pushScreen(ScreenGameTypeSelect)
	m.menuOptions = []string{"Player vs Player", "Player vs Bot", "Bot vs Bot"}
	m.menuSelection = 0

	// Select Player vs Player
	result, _ := m.handleGameTypeSelection()
	m = result.(Model)

	// Nav stack should be cleared
	if len(m.navStack) != 0 {
		t.Errorf("Expected empty nav stack after starting game, got %d items", len(m.navStack))
	}
	if m.screen != ScreenGamePlay {
		t.Errorf("Expected screen to be GamePlay, got %v", m.screen)
	}
}

// TestShortcutsOverlayToggle tests that '?' key toggles the shortcuts overlay
func TestShortcutsOverlayToggle(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenMainMenu

	// Initially overlay should be hidden
	if m.showShortcutsOverlay {
		t.Error("Expected shortcuts overlay to be hidden initially")
	}

	// Press '?' key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	result, _ := m.handleKeyPress(msg)
	m = result.(Model)

	// Overlay should now be visible
	if !m.showShortcutsOverlay {
		t.Error("Expected shortcuts overlay to be visible after pressing '?'")
	}

	// Verify screen didn't change
	if m.screen != ScreenMainMenu {
		t.Errorf("Expected screen to remain MainMenu, got %v", m.screen)
	}
}

// TestShortcutsOverlayDismissOnAnyKey tests that any key dismisses the overlay
func TestShortcutsOverlayDismissOnAnyKey(t *testing.T) {
	testKeys := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'a'}},
		{Type: tea.KeyEnter},
		{Type: tea.KeyEsc},
		{Type: tea.KeySpace},
		{Type: tea.KeyRunes, Runes: []rune{'?'}},
	}

	for _, key := range testKeys {
		m := NewModel(DefaultConfig())
		m.screen = ScreenMainMenu
		m.showShortcutsOverlay = true

		result, _ := m.handleKeyPress(key)
		m = result.(Model)

		if m.showShortcutsOverlay {
			t.Errorf("Expected overlay to be dismissed by key %v", key)
		}

		// Screen should remain unchanged
		if m.screen != ScreenMainMenu {
			t.Errorf("Expected screen to remain MainMenu after dismiss, got %v", m.screen)
		}
	}
}

// TestShortcutsOverlayNotShownInTextInputMode tests that '?' doesn't show overlay in text input modes
func TestShortcutsOverlayNotShownInTextInputMode(t *testing.T) {
	// Test FEN input screen (text input mode)
	m := NewModel(DefaultConfig())
	m.screen = ScreenFENInput

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	result, _ := m.handleKeyPress(msg)
	m = result.(Model)

	if m.showShortcutsOverlay {
		t.Error("Expected shortcuts overlay NOT to show on FEN input screen")
	}

	// Test GamePlay screen (text input mode)
	m = NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	result, _ = m.handleKeyPress(msg)
	m = result.(Model)

	if m.showShortcutsOverlay {
		t.Error("Expected shortcuts overlay NOT to show on GamePlay screen")
	}

	// Test BvB count input mode
	m = NewModel(DefaultConfig())
	m.screen = ScreenBvBGameMode
	m.bvbInputtingCount = true

	result, _ = m.handleKeyPress(msg)
	m = result.(Model)

	if m.showShortcutsOverlay {
		t.Error("Expected shortcuts overlay NOT to show in BvB count input mode")
	}

	// Test BvB grid input mode
	m = NewModel(DefaultConfig())
	m.screen = ScreenBvBGridConfig
	m.bvbInputtingGrid = true

	result, _ = m.handleKeyPress(msg)
	m = result.(Model)

	if m.showShortcutsOverlay {
		t.Error("Expected shortcuts overlay NOT to show in BvB grid input mode")
	}
}

// TestShortcutsOverlayShownOnNonTextInputScreens tests that '?' shows overlay on non-text-input screens
func TestShortcutsOverlayShownOnNonTextInputScreens(t *testing.T) {
	screens := []Screen{
		ScreenMainMenu,
		ScreenGameTypeSelect,
		ScreenBotSelect,
		ScreenColorSelect,
		ScreenSettings,
		ScreenGameOver,
		ScreenBvBBotSelect,
	}

	for _, screen := range screens {
		m := NewModel(DefaultConfig())
		m.screen = screen

		// Set up any required state for certain screens
		if screen == ScreenGameOver || screen == ScreenGamePlay {
			m.board = engine.NewBoard()
		}
		if screen == ScreenGameTypeSelect || screen == ScreenBotSelect || screen == ScreenColorSelect || screen == ScreenBvBBotSelect {
			m.menuOptions = []string{"Option 1", "Option 2"}
		}

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
		result, _ := m.handleKeyPress(msg)
		m = result.(Model)

		if !m.showShortcutsOverlay {
			t.Errorf("Expected shortcuts overlay to show on screen %v", screen)
		}
	}
}

// TestShortcutsOverlayRendersContent tests that the overlay renders expected content
func TestShortcutsOverlayRendersContent(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.showShortcutsOverlay = true

	view := m.View()

	// Check for title
	if !containsText(view, "Keyboard Shortcuts") {
		t.Error("Expected overlay to contain 'Keyboard Shortcuts' title")
	}

	// Check for section headers
	expectedSections := []string{
		"Global",
		"Menu Navigation",
		"Settings",
		"Gameplay",
		"Bot vs Bot",
	}
	for _, section := range expectedSections {
		if !containsText(view, section) {
			t.Errorf("Expected overlay to contain section '%s'", section)
		}
	}

	// Check for some key shortcuts
	expectedShortcuts := []string{
		"Ctrl+C",
		"Esc",
		"Enter",
		"resign",
		"offerdraw",
		"Space",
	}
	for _, shortcut := range expectedShortcuts {
		if !containsText(view, shortcut) {
			t.Errorf("Expected overlay to contain shortcut '%s'", shortcut)
		}
	}

	// Check for dismiss hint
	if !containsText(view, "Press any key to close") {
		t.Error("Expected overlay to contain dismiss hint")
	}
}

// TestIsInTextInputMode tests the isInTextInputMode helper function
func TestIsInTextInputMode(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*Model)
		expectedResult bool
	}{
		{
			name: "MainMenu is not text input",
			setup: func(m *Model) {
				m.screen = ScreenMainMenu
			},
			expectedResult: false,
		},
		{
			name: "FENInput is text input",
			setup: func(m *Model) {
				m.screen = ScreenFENInput
			},
			expectedResult: true,
		},
		{
			name: "GamePlay is text input",
			setup: func(m *Model) {
				m.screen = ScreenGamePlay
				m.board = engine.NewBoard()
			},
			expectedResult: true,
		},
		{
			name: "BvBGameMode without input is not text input",
			setup: func(m *Model) {
				m.screen = ScreenBvBGameMode
				m.bvbInputtingCount = false
			},
			expectedResult: false,
		},
		{
			name: "BvBGameMode with input is text input",
			setup: func(m *Model) {
				m.screen = ScreenBvBGameMode
				m.bvbInputtingCount = true
			},
			expectedResult: true,
		},
		{
			name: "BvBGridConfig without input is not text input",
			setup: func(m *Model) {
				m.screen = ScreenBvBGridConfig
				m.bvbInputtingGrid = false
			},
			expectedResult: false,
		},
		{
			name: "BvBGridConfig with input is text input",
			setup: func(m *Model) {
				m.screen = ScreenBvBGridConfig
				m.bvbInputtingGrid = true
			},
			expectedResult: true,
		},
		{
			name: "Settings is not text input",
			setup: func(m *Model) {
				m.screen = ScreenSettings
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(DefaultConfig())
			tt.setup(&m)

			result := m.isInTextInputMode()
			if result != tt.expectedResult {
				t.Errorf("isInTextInputMode() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

// containsText is a helper function to check if a string contains a substring
func containsText(s, substr string) bool {
	return strings.Contains(s, substr)
}
