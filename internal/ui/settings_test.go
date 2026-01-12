package ui

import (
	"os"
	"strings"
	"testing"

	"github.com/Mgrdich/TermChess/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

// TestSettingsNavigation tests that the settings screen navigation works correctly
func TestSettingsNavigation(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenSettings
	m.settingsSelection = 0

	// Test moving down
	model, _ := m.handleSettingsKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = model.(Model)
	if m.settingsSelection != 1 {
		t.Errorf("Expected settingsSelection to be 1, got %d", m.settingsSelection)
	}

	// Test moving down again
	model, _ = m.handleSettingsKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = model.(Model)
	if m.settingsSelection != 2 {
		t.Errorf("Expected settingsSelection to be 2, got %d", m.settingsSelection)
	}

	// Test moving up
	model, _ = m.handleSettingsKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = model.(Model)
	if m.settingsSelection != 1 {
		t.Errorf("Expected settingsSelection to be 1, got %d", m.settingsSelection)
	}

	// Test wrapping at bottom (should go from 4 to 0)
	m.settingsSelection = 4
	model, _ = m.handleSettingsKeys(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(Model)
	if m.settingsSelection != 0 {
		t.Errorf("Expected settingsSelection to wrap to 0, got %d", m.settingsSelection)
	}

	// Test wrapping at top (should go from 0 to 4)
	m.settingsSelection = 0
	model, _ = m.handleSettingsKeys(tea.KeyMsg{Type: tea.KeyUp})
	m = model.(Model)
	if m.settingsSelection != 4 {
		t.Errorf("Expected settingsSelection to wrap to 4, got %d", m.settingsSelection)
	}
}

// TestSettingsToggle tests that toggling settings works correctly
func TestSettingsToggle(t *testing.T) {
	// Clean up any existing config file after test
	defer func() {
		configPath, _ := config.GetConfigPath()
		os.Remove(configPath)
	}()

	m := NewModel(DefaultConfig())
	m.screen = ScreenSettings

	// Test toggling UseUnicode (option 0)
	m.settingsSelection = 0
	initialValue := m.config.UseUnicode
	model, _ := m.handleSettingsKeys(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(Model)
	if m.config.UseUnicode == initialValue {
		t.Errorf("Expected UseUnicode to toggle from %v to %v", initialValue, !initialValue)
	}

	// Test toggling ShowCoords (option 1)
	m.settingsSelection = 1
	initialValue = m.config.ShowCoords
	model, _ = m.handleSettingsKeys(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(Model)
	if m.config.ShowCoords == initialValue {
		t.Errorf("Expected ShowCoords to toggle from %v to %v", initialValue, !initialValue)
	}

	// Test toggling UseColors (option 2)
	m.settingsSelection = 2
	initialValue = m.config.UseColors
	model, _ = m.handleSettingsKeys(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(Model)
	if m.config.UseColors == initialValue {
		t.Errorf("Expected UseColors to toggle from %v to %v", initialValue, !initialValue)
	}

	// Test toggling ShowMoveHistory (option 3)
	m.settingsSelection = 3
	initialValue = m.config.ShowMoveHistory
	model, _ = m.handleSettingsKeys(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(Model)
	if m.config.ShowMoveHistory == initialValue {
		t.Errorf("Expected ShowMoveHistory to toggle from %v to %v", initialValue, !initialValue)
	}

	// Test toggling ShowHelpText (option 4)
	m.settingsSelection = 4
	initialValue = m.config.ShowHelpText
	model, _ = m.handleSettingsKeys(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(Model)
	if m.config.ShowHelpText == initialValue {
		t.Errorf("Expected ShowHelpText to toggle from %v to %v", initialValue, !initialValue)
	}
}

// TestSettingsReturnToMenu tests that ESC/q/b/backspace return to main menu
func TestSettingsReturnToMenu(t *testing.T) {
	testCases := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"esc", tea.KeyMsg{Type: tea.KeyEsc}},
		{"q", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}},
		{"b", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}},
		{"backspace", tea.KeyMsg{Type: tea.KeyBackspace}},
	}

	for _, tc := range testCases {
		m := NewModel(DefaultConfig())
		m.screen = ScreenSettings

		model, _ := m.handleSettingsKeys(tc.key)
		m = model.(Model)

		if m.screen != ScreenMainMenu {
			t.Errorf("Expected screen to be ScreenMainMenu after pressing '%s', got %d", tc.name, m.screen)
		}
	}
}

// TestMainMenuToSettings tests navigation from main menu to settings
func TestMainMenuToSettings(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenMainMenu
	m.menuSelection = 2 // "Settings" is the 3rd option (index 2)

	model, _ := m.handleMainMenuSelection()
	m = model.(Model)

	if m.screen != ScreenSettings {
		t.Errorf("Expected screen to be ScreenSettings, got %d", m.screen)
	}

	if m.settingsSelection != 0 {
		t.Errorf("Expected settingsSelection to be 0, got %d", m.settingsSelection)
	}
}

// TestSettingsRender tests that the settings screen renders without errors
func TestSettingsRender(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenSettings

	// Test with different settings values
	m.config.UseUnicode = true
	m.config.ShowCoords = false
	m.config.UseColors = true
	m.config.ShowMoveHistory = false

	output := m.renderSettings()

	if output == "" {
		t.Error("Expected non-empty output from renderSettings()")
	}

	// Check that output contains expected strings
	expectedStrings := []string{"Settings", "Use Unicode Pieces", "Show Coordinates", "Use Colors", "Show Move History", "Show Help Text"}
	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}
}
