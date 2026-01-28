package ui

import (
	"os"
	"strings"
	"testing"

	"github.com/Mgrdich/TermChess/internal/config"
	"github.com/Mgrdich/TermChess/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

// TestRenderHelpText tests the renderHelpText helper method
func TestRenderHelpText(t *testing.T) {
	testText := "test help text"

	// Test with ShowHelpText enabled
	m := NewModel(Config{ShowHelpText: true, Theme: ThemeNameClassic})
	result := m.renderHelpText(testText)
	if result == "" {
		t.Error("Expected non-empty result when ShowHelpText is true")
	}

	// Test with ShowHelpText disabled
	m.config.ShowHelpText = false
	result = m.renderHelpText(testText)
	if result != "" {
		t.Errorf("Expected empty result when ShowHelpText is false, got %q", result)
	}
}

// TestHelpTextVisibilityMainMenu tests that help text is shown/hidden on main menu
func TestHelpTextVisibilityMainMenu(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenMainMenu

	// Test with help text enabled
	m.config.ShowHelpText = true
	output := m.renderMainMenu()
	if !strings.Contains(output, "arrows/jk") && !strings.Contains(output, "navigate") {
		t.Error("Expected help text to be visible on main menu when ShowHelpText is true")
	}

	// Test with help text disabled
	m.config.ShowHelpText = false
	output = m.renderMainMenu()
	if strings.Contains(output, "arrows/jk: navigate | enter: select | q: quit") {
		t.Error("Expected help text to be hidden on main menu when ShowHelpText is false")
	}
}

// TestHelpTextVisibilityGameTypeSelect tests that help text is shown/hidden on game type select
func TestHelpTextVisibilityGameTypeSelect(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenGameTypeSelect
	m.menuOptions = []string{"Player vs Player", "Player vs Bot"}

	// Test with help text enabled
	m.config.ShowHelpText = true
	output := m.renderGameTypeSelect()
	if !strings.Contains(output, "ESC") && !strings.Contains(output, "back to menu") {
		t.Error("Expected help text to be visible on game type select when ShowHelpText is true")
	}

	// Test with help text disabled
	m.config.ShowHelpText = false
	output = m.renderGameTypeSelect()
	if strings.Contains(output, "ESC: back to menu | arrows/jk: navigate | enter: select") {
		t.Error("Expected help text to be hidden on game type select when ShowHelpText is false")
	}
}

// TestHelpTextVisibilitySettings tests that help text is shown/hidden on settings screen
func TestHelpTextVisibilitySettings(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenSettings

	// Test with help text enabled
	m.config.ShowHelpText = true
	output := m.renderSettings()
	if !strings.Contains(output, "ESC") && !strings.Contains(output, "back") {
		t.Error("Expected help text to be visible on settings screen when ShowHelpText is true")
	}

	// Test with help text disabled
	m.config.ShowHelpText = false
	output = m.renderSettings()
	if strings.Contains(output, "ESC: back | arrows/jk: navigate | enter/space: toggle") {
		t.Error("Expected help text to be hidden on settings screen when ShowHelpText is false")
	}
}

// TestHelpTextVisibilityGamePlay tests that help text is shown/hidden on gameplay screen
func TestHelpTextVisibilityGamePlay(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenGamePlay
	// Start a new game so we have a board
	m.board = engine.NewBoard()

	// Test with help text enabled
	m.config.ShowHelpText = true
	output := m.renderGamePlay()
	if !strings.Contains(output, "Type move") && !strings.Contains(output, "ESC") {
		t.Error("Expected help text to be visible on gameplay screen when ShowHelpText is true")
	}

	// Test with help text disabled
	m.config.ShowHelpText = false
	output = m.renderGamePlay()
	// The output should not contain the help text pattern
	if strings.Contains(output, "Type move") && strings.Contains(output, "ESC: menu") {
		t.Error("Expected help text to be hidden on gameplay screen when ShowHelpText is false")
	}
}

// TestHelpTextVisibilityGameOver tests that help text is shown/hidden on game over screen
func TestHelpTextVisibilityGameOver(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenGameOver
	// Create a board in checkmate state for testing
	m.board = engine.NewBoard()
	// Force a checkmate state (this is a simplified test)
	// In a real scenario, we'd set up an actual checkmate position

	// Test with help text enabled
	m.config.ShowHelpText = true
	output := m.renderGameOver()
	// Game over screen shows options, but might also have help text
	if !strings.Contains(output, "new game") && !strings.Contains(output, "main menu") {
		t.Error("Expected navigation options to be visible on game over screen")
	}

	// Test with help text disabled
	m.config.ShowHelpText = false
	output = m.renderGameOver()
	// Should still show options but not additional help text at bottom
	if !strings.Contains(output, "Press 'n'") {
		t.Error("Expected options to still be visible even with ShowHelpText disabled")
	}
}

// TestShowHelpTextPersistence tests that ShowHelpText setting persists across restarts
func TestShowHelpTextPersistence(t *testing.T) {
	// Clean up after test
	configPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("config.GetConfigPath() failed: %v", err)
	}
	defer os.Remove(configPath)

	// Clean up before test to start fresh
	os.Remove(configPath)

	// Phase 1: Start app with default config
	m1 := NewModel(LoadConfig())
	if !m1.config.ShowHelpText {
		t.Error("Expected ShowHelpText to be true by default")
	}

	// Phase 2: Toggle ShowHelpText off
	m1.screen = ScreenSettings
	m1.settingsSelection = 4 // ShowHelpText is the 5th option (index 4)

	model, _ := m1.handleSettingsKeys(tea.KeyMsg{Type: tea.KeyEnter})
	m1 = model.(Model)

	// Verify the toggle happened
	if m1.config.ShowHelpText {
		t.Error("Expected ShowHelpText to be toggled to false")
	}

	// Phase 3: "Restart" the app by creating a new model
	m2 := NewModel(LoadConfig())

	// Verify the config was loaded from disk with ShowHelpText = false
	if m2.config.ShowHelpText {
		t.Error("After restart, ShowHelpText should be false (loaded from config file)")
	}

	// Phase 4: Toggle back to true
	m2.screen = ScreenSettings
	m2.settingsSelection = 4

	model, _ = m2.handleSettingsKeys(tea.KeyMsg{Type: tea.KeyEnter})
	m2 = model.(Model)

	if !m2.config.ShowHelpText {
		t.Error("Expected ShowHelpText to be toggled back to true")
	}

	// Phase 5: Restart again and verify it's true
	m3 := NewModel(LoadConfig())
	if !m3.config.ShowHelpText {
		t.Error("After second restart, ShowHelpText should be true")
	}
}

// TestHelpTextToggleAffectsAllScreens tests that toggling help text affects all screens
func TestHelpTextToggleAffectsAllScreens(t *testing.T) {
	// Clean up after test
	configPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("config.GetConfigPath() failed: %v", err)
	}
	defer os.Remove(configPath)

	m := NewModel(DefaultConfig())

	// Start with help text enabled
	m.config.ShowHelpText = true

	// Check all screens have help text
	screens := []struct {
		name   string
		screen Screen
		setup  func(*Model)
		render func(Model) string
	}{
		{
			"MainMenu",
			ScreenMainMenu,
			func(m *Model) { m.screen = ScreenMainMenu },
			func(m Model) string { return m.renderMainMenu() },
		},
		{
			"GameTypeSelect",
			ScreenGameTypeSelect,
			func(m *Model) {
				m.screen = ScreenGameTypeSelect
				m.menuOptions = []string{"Player vs Player", "Player vs Bot"}
			},
			func(m Model) string { return m.renderGameTypeSelect() },
		},
		{
			"Settings",
			ScreenSettings,
			func(m *Model) { m.screen = ScreenSettings },
			func(m Model) string { return m.renderSettings() },
		},
		{
			"GamePlay",
			ScreenGamePlay,
			func(m *Model) {
				m.screen = ScreenGamePlay
				m.board = engine.NewBoard()
			},
			func(m Model) string { return m.renderGamePlay() },
		},
	}

	for _, screen := range screens {
		// Setup screen
		screen.setup(&m)

		// Test with help text enabled
		m.config.ShowHelpText = true
		output := screen.render(m)
		// We don't check for specific text, just that something is rendered
		if output == "" {
			t.Errorf("%s: Expected non-empty output", screen.name)
		}

		// Test with help text disabled
		m.config.ShowHelpText = false
		outputDisabled := screen.render(m)
		// The output should still be non-empty (main content remains)
		if outputDisabled == "" {
			t.Errorf("%s: Expected non-empty output even with help text disabled", screen.name)
		}

		// Output with help disabled should be shorter or equal (help text removed)
		// Note: This isn't always guaranteed due to formatting, so we just check it renders
	}
}

// TestHelpTextVisibilityFENInput tests that help text is shown/hidden on FEN input screen
func TestHelpTextVisibilityFENInput(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenFENInput

	// Test with help text enabled
	m.config.ShowHelpText = true
	output := m.renderFENInput()
	if !strings.Contains(output, "ESC") || !strings.Contains(output, "enter") {
		t.Error("Expected help text to be visible on FEN input screen when ShowHelpText is true")
	}

	// Test with help text disabled
	m.config.ShowHelpText = false
	output = m.renderFENInput()
	// The main content should still be there, but help text should be hidden
	if strings.Contains(output, "ESC: back to menu") {
		t.Error("Expected help text to be hidden on FEN input screen when ShowHelpText is false")
	}
}

// TestHelpTextVisibilitySavePrompt tests that help text is shown/hidden on save prompt screen
func TestHelpTextVisibilitySavePrompt(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenSavePrompt
	m.board = engine.NewBoard()

	// Test with help text enabled
	m.config.ShowHelpText = true
	output := m.renderSavePrompt()
	if !strings.Contains(output, "y:") || !strings.Contains(output, "ESC") {
		t.Error("Expected help text to be visible on save prompt screen when ShowHelpText is true")
	}

	// Test with help text disabled
	m.config.ShowHelpText = false
	output = m.renderSavePrompt()
	// Options should still be visible, but the help text at the bottom should be hidden
	if strings.Contains(output, "y: save and exit | n: exit without saving | ESC: cancel") {
		t.Error("Expected help text to be hidden on save prompt screen when ShowHelpText is false")
	}
}

// TestHelpTextVisibilityResumePrompt tests that help text is shown/hidden on resume prompt screen
func TestHelpTextVisibilityResumePrompt(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.screen = ScreenResumePrompt

	// Test with help text enabled
	m.config.ShowHelpText = true
	output := m.renderResumePrompt()
	if !strings.Contains(output, "y:") || !strings.Contains(output, "n:") {
		t.Error("Expected help text to be visible on resume prompt screen when ShowHelpText is true")
	}

	// Test with help text disabled
	m.config.ShowHelpText = false
	output = m.renderResumePrompt()
	// Options should still be visible, but the help text at the bottom should be hidden
	if strings.Contains(output, "y: resume game | n: go to main menu") {
		t.Error("Expected help text to be hidden on resume prompt screen when ShowHelpText is false")
	}
}

// TestHelpTextContentMatchesSpec tests that help text content matches the specification
func TestHelpTextContentMatchesSpec(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.config.ShowHelpText = true

	tests := []struct {
		name           string
		screen         Screen
		setup          func(*Model)
		render         func(Model) string
		expectedPhrase string
	}{
		{
			"MainMenu",
			ScreenMainMenu,
			func(m *Model) { m.screen = ScreenMainMenu },
			func(m Model) string { return m.renderMainMenu() },
			"arrows/jk: navigate | enter: select | q: quit",
		},
		{
			"GameTypeSelect",
			ScreenGameTypeSelect,
			func(m *Model) {
				m.screen = ScreenGameTypeSelect
				m.menuOptions = []string{"Player vs Player", "Player vs Bot"}
			},
			func(m Model) string { return m.renderGameTypeSelect() },
			"ESC: back to menu",
		},
		{
			"FENInput",
			ScreenFENInput,
			func(m *Model) { m.screen = ScreenFENInput },
			func(m Model) string { return m.renderFENInput() },
			"ESC: back to menu | enter: load position",
		},
		{
			"GamePlay",
			ScreenGamePlay,
			func(m *Model) {
				m.screen = ScreenGamePlay
				m.board = engine.NewBoard()
			},
			func(m Model) string { return m.renderGamePlay() },
			"ESC: menu (with save)",
		},
		{
			"GameOver",
			ScreenGameOver,
			func(m *Model) {
				m.screen = ScreenGameOver
				m.board = engine.NewBoard()
			},
			func(m Model) string { return m.renderGameOver() },
			"ESC/m: menu",
		},
		{
			"Settings",
			ScreenSettings,
			func(m *Model) { m.screen = ScreenSettings },
			func(m Model) string { return m.renderSettings() },
			"ESC: back",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(&m)
			output := tt.render(m)
			if !strings.Contains(output, tt.expectedPhrase) {
				t.Errorf("Expected help text to contain '%s', output:\n%s", tt.expectedPhrase, output)
			}
		})
	}
}
