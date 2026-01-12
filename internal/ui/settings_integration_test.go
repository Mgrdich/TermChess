package ui

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestSettingsConfigPersistence tests that config changes made in the Settings screen
// are persisted to disk and loaded correctly on next app start
func TestSettingsConfigPersistence(t *testing.T) {
	// Clean up after test
	configPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() failed: %v", err)
	}
	defer os.Remove(configPath)

	// Clean up before test to start fresh
	os.Remove(configPath)

	// Phase 1: Start app (loads default config)
	m1 := NewModel(DefaultConfig())
	initialUseUnicode := m1.config.UseUnicode
	initialShowCoords := m1.config.ShowCoords
	initialUseColors := m1.config.UseColors
	initialShowMoveHistory := m1.config.ShowMoveHistory
	initialShowHelpText := m1.config.ShowHelpText

	// Phase 2: Navigate to Settings and toggle UseUnicode
	m1.screen = ScreenSettings
	m1.settingsSelection = 0 // UseUnicode option

	model, _ := m1.handleSettingsKeys(tea.KeyMsg{Type: tea.KeyEnter})
	m1 = model.(Model)

	// Verify the toggle happened
	if m1.config.UseUnicode == initialUseUnicode {
		t.Errorf("Expected UseUnicode to toggle from %v to %v", initialUseUnicode, !initialUseUnicode)
	}

	// Phase 3: Toggle ShowCoords
	m1.settingsSelection = 1 // ShowCoords option
	model, _ = m1.handleSettingsKeys(tea.KeyMsg{Type: tea.KeyEnter})
	m1 = model.(Model)

	if m1.config.ShowCoords == initialShowCoords {
		t.Errorf("Expected ShowCoords to toggle from %v to %v", initialShowCoords, !initialShowCoords)
	}

	// Phase 4: "Restart" the app by creating a new model (simulates app restart)
	m2 := NewModel(LoadConfig())

	// Verify the config was loaded from disk with the toggled values
	if m2.config.UseUnicode != m1.config.UseUnicode {
		t.Errorf("After restart, UseUnicode = %v, want %v", m2.config.UseUnicode, m1.config.UseUnicode)
	}

	if m2.config.ShowCoords != m1.config.ShowCoords {
		t.Errorf("After restart, ShowCoords = %v, want %v", m2.config.ShowCoords, m1.config.ShowCoords)
	}

	// Verify unchanged options remain the same
	if m2.config.UseColors != initialUseColors {
		t.Errorf("After restart, UseColors = %v, want %v (should be unchanged)", m2.config.UseColors, initialUseColors)
	}

	if m2.config.ShowMoveHistory != initialShowMoveHistory {
		t.Errorf("After restart, ShowMoveHistory = %v, want %v (should be unchanged)", m2.config.ShowMoveHistory, initialShowMoveHistory)
	}

	if m2.config.ShowHelpText != initialShowHelpText {
		t.Errorf("After restart, ShowHelpText = %v, want %v (should be unchanged)", m2.config.ShowHelpText, initialShowHelpText)
	}
}

// TestSettingsToggleAllOptions tests toggling all settings options
func TestSettingsToggleAllOptions(t *testing.T) {
	// Clean up after test
	configPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() failed: %v", err)
	}
	defer os.Remove(configPath)

	// Start with fresh config
	os.Remove(configPath)

	m := NewModel(DefaultConfig())
	m.screen = ScreenSettings

	// Toggle all options
	for i := 0; i < 5; i++ {
		m.settingsSelection = i
		model, _ := m.handleSettingsKeys(tea.KeyMsg{Type: tea.KeyEnter})
		m = model.(Model)
	}

	// Verify all options were toggled
	defaults := DefaultConfig()
	if m.config.UseUnicode == defaults.UseUnicode {
		t.Errorf("UseUnicode should be toggled from default")
	}
	if m.config.ShowCoords == defaults.ShowCoords {
		t.Errorf("ShowCoords should be toggled from default")
	}
	if m.config.UseColors == defaults.UseColors {
		t.Errorf("UseColors should be toggled from default")
	}
	if m.config.ShowMoveHistory == defaults.ShowMoveHistory {
		t.Errorf("ShowMoveHistory should be toggled from default")
	}
	if m.config.ShowHelpText == defaults.ShowHelpText {
		t.Errorf("ShowHelpText should be toggled from default")
	}

	// Verify persistence
	m2 := NewModel(LoadConfig())
	if m2.config != m.config {
		t.Errorf("Config after restart = %+v, want %+v", m2.config, m.config)
	}
}

// TestSettingsStatusMessages tests that status messages are displayed correctly
func TestSettingsStatusMessages(t *testing.T) {
	// Clean up after test
	configPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() failed: %v", err)
	}
	defer os.Remove(configPath)

	m := NewModel(DefaultConfig())
	m.screen = ScreenSettings
	m.settingsSelection = 0

	// Toggle a setting
	model, _ := m.handleSettingsKeys(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(Model)

	// Verify success message is set
	if m.statusMsg == "" {
		t.Error("Expected status message to be set after toggling setting")
	}

	expectedMsg := "Setting saved successfully"
	if m.statusMsg != expectedMsg {
		t.Errorf("Status message = %q, want %q", m.statusMsg, expectedMsg)
	}

	// Verify error message is cleared
	if m.errorMsg != "" {
		t.Errorf("Expected error message to be empty, got %q", m.errorMsg)
	}
}

// TestSettingsNavigationAndReturn tests navigating to Settings and returning to main menu
func TestSettingsNavigationAndReturn(t *testing.T) {
	m := NewModel(DefaultConfig())

	// Navigate from main menu to settings
	m.screen = ScreenMainMenu
	m.menuSelection = 2 // Settings option

	model, _ := m.handleMainMenuSelection()
	m = model.(Model)

	if m.screen != ScreenSettings {
		t.Errorf("Expected screen to be ScreenSettings, got %d", m.screen)
	}

	// Return to main menu
	model, _ = m.handleSettingsKeys(tea.KeyMsg{Type: tea.KeyEsc})
	m = model.(Model)

	if m.screen != ScreenMainMenu {
		t.Errorf("Expected screen to be ScreenMainMenu after ESC, got %d", m.screen)
	}

	// Verify menu selection is reset
	if m.menuSelection != 0 {
		t.Errorf("Expected menuSelection to be reset to 0, got %d", m.menuSelection)
	}
}
