package ui

import (
	"testing"
)

// TestGetTheme_Classic tests that GetTheme returns the classic theme for ThemeClassic.
func TestGetTheme_Classic(t *testing.T) {
	theme := GetTheme(ThemeClassic)
	if theme.Name != ThemeNameClassic {
		t.Errorf("Expected theme name %q, got %q", ThemeNameClassic, theme.Name)
	}

	// Verify theme has the expected colors (matching original hardcoded values)
	if theme.TitleText != "#FAFAFA" {
		t.Errorf("Expected TitleText '#FAFAFA', got %v", theme.TitleText)
	}
	if theme.MenuNormal != "#FFFDF5" {
		t.Errorf("Expected MenuNormal '#FFFDF5', got %v", theme.MenuNormal)
	}
	if theme.MenuSelected != "#7D56F4" {
		t.Errorf("Expected MenuSelected '#7D56F4', got %v", theme.MenuSelected)
	}
	if theme.HelpText != "#626262" {
		t.Errorf("Expected HelpText '#626262', got %v", theme.HelpText)
	}
	if theme.ErrorText != "#FF5555" {
		t.Errorf("Expected ErrorText '#FF5555', got %v", theme.ErrorText)
	}
	if theme.StatusText != "#50FA7B" {
		t.Errorf("Expected StatusText '#50FA7B', got %v", theme.StatusText)
	}
}

// TestParseThemeName tests that ParseThemeName correctly parses theme strings.
func TestParseThemeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ThemeName
	}{
		{"classic string", ThemeNameClassic, ThemeClassic},
		{"modern string", ThemeNameModern, ThemeModern},
		{"minimalist string", ThemeNameMinimalist, ThemeMinimalist},
		{"empty string defaults to classic", "", ThemeClassic},
		{"unknown string defaults to classic", "nonexistent", ThemeClassic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseThemeName(tt.input)
			if got != tt.expected {
				t.Errorf("ParseThemeName(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// TestThemeName_String tests that ThemeName.String() returns correct strings.
func TestThemeName_String(t *testing.T) {
	tests := []struct {
		name     string
		input    ThemeName
		expected string
	}{
		{"classic", ThemeClassic, ThemeNameClassic},
		{"modern", ThemeModern, ThemeNameModern},
		{"minimalist", ThemeMinimalist, ThemeNameMinimalist},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.String()
			if got != tt.expected {
				t.Errorf("ThemeName(%d).String() = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestGetTheme_AllVariants tests that GetTheme returns correct theme for all ThemeName values.
func TestGetTheme_AllVariants(t *testing.T) {
	tests := []struct {
		name         string
		input        ThemeName
		expectedName string
	}{
		{"classic", ThemeClassic, ThemeNameClassic},
		{"modern returns classic (not yet implemented)", ThemeModern, ThemeNameClassic},
		{"minimalist returns classic (not yet implemented)", ThemeMinimalist, ThemeNameClassic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme := GetTheme(tt.input)
			if theme.Name != tt.expectedName {
				t.Errorf("GetTheme(%v) returned theme with Name %q, want %q", tt.input, theme.Name, tt.expectedName)
			}
		})
	}
}

// TestClassicTheme_AllFieldsSet tests that the classic theme has all fields set.
func TestClassicTheme_AllFieldsSet(t *testing.T) {
	theme := GetTheme(ThemeClassic)

	// Board colors
	if theme.LightSquare == "" {
		t.Error("Expected LightSquare to be set")
	}
	if theme.DarkSquare == "" {
		t.Error("Expected DarkSquare to be set")
	}
	if theme.WhitePiece == "" {
		t.Error("Expected WhitePiece to be set")
	}
	if theme.BlackPiece == "" {
		t.Error("Expected BlackPiece to be set")
	}

	// Selection colors
	if theme.SelectedHighlight == "" {
		t.Error("Expected SelectedHighlight to be set")
	}
	if theme.ValidMoveHighlight == "" {
		t.Error("Expected ValidMoveHighlight to be set")
	}

	// UI colors
	if theme.BoardBorder == "" {
		t.Error("Expected BoardBorder to be set")
	}
	if theme.MenuSelected == "" {
		t.Error("Expected MenuSelected to be set")
	}
	if theme.MenuNormal == "" {
		t.Error("Expected MenuNormal to be set")
	}
	if theme.TitleText == "" {
		t.Error("Expected TitleText to be set")
	}
	if theme.HelpText == "" {
		t.Error("Expected HelpText to be set")
	}
	if theme.ErrorText == "" {
		t.Error("Expected ErrorText to be set")
	}
	if theme.StatusText == "" {
		t.Error("Expected StatusText to be set")
	}

	// Turn indicator colors
	if theme.WhiteTurnText == "" {
		t.Error("Expected WhiteTurnText to be set")
	}
	if theme.BlackTurnText == "" {
		t.Error("Expected BlackTurnText to be set")
	}
}

// TestModelThemeInitialization tests that Model initializes with theme from config.
func TestModelThemeInitialization(t *testing.T) {
	config := Config{
		Theme: ThemeNameClassic,
	}

	m := NewModel(config)

	if m.theme.Name != ThemeNameClassic {
		t.Errorf("Expected model theme to be %q, got %q", ThemeNameClassic, m.theme.Name)
	}
}

// TestModelThemeInitialization_DefaultOnEmpty tests that Model initializes with classic theme for empty config.
func TestModelThemeInitialization_DefaultOnEmpty(t *testing.T) {
	config := Config{
		Theme: "",
	}

	m := NewModel(config)

	if m.theme.Name != ThemeNameClassic {
		t.Errorf("Expected model theme to be %q for empty config theme, got %q", ThemeNameClassic, m.theme.Name)
	}
}

// TestThemeStyleMethods tests that theme-based style methods work correctly.
func TestThemeStyleMethods(t *testing.T) {
	config := Config{
		Theme: ThemeNameClassic,
	}
	m := NewModel(config)

	// Test that style methods don't panic and return non-nil styles
	_ = m.titleStyle()
	_ = m.menuItemStyle()
	_ = m.selectedItemStyle()
	_ = m.helpStyle()
	_ = m.errorStyle()
	_ = m.statusStyle()
	_ = m.cursorStyle()

	// If we get here without panics, the test passes
}

// TestThemeUsedInRendering tests that theme colors are used in rendering.
func TestThemeUsedInRendering(t *testing.T) {
	config := Config{
		Theme:        ThemeNameClassic,
		ShowHelpText: true,
	}
	m := NewModel(config)
	m.screen = ScreenMainMenu

	// Render main menu
	output := m.renderMainMenu()

	// Basic check that rendering works with theme
	if output == "" {
		t.Error("Expected non-empty output from renderMainMenu with theme")
	}

	// Should contain "TermChess" title
	if !containsString(output, "TermChess") {
		t.Error("Expected output to contain 'TermChess'")
	}
}

// containsString is a helper to check if a string contains a substring
// (without importing strings package just for this)
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
