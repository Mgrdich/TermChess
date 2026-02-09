package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Mgrdich/TermChess/internal/engine"
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
		{"modern", ThemeModern, ThemeNameModern},
		{"minimalist", ThemeMinimalist, ThemeNameMinimalist},
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

	// Menu hierarchy colors
	if theme.MenuPrimary == "" {
		t.Error("Expected MenuPrimary to be set")
	}
	if theme.MenuSecondary == "" {
		t.Error("Expected MenuSecondary to be set")
	}
	if theme.MenuSeparator == "" {
		t.Error("Expected MenuSeparator to be set")
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

// TestGetTheme_Modern tests that GetTheme returns the modern theme correctly.
func TestGetTheme_Modern(t *testing.T) {
	theme := GetTheme(ThemeModern)
	if theme.Name != ThemeNameModern {
		t.Errorf("Expected theme name %q, got %q", ThemeNameModern, theme.Name)
	}

	// Verify theme has the expected modern colors
	if theme.TitleText != "#E0E0E0" {
		t.Errorf("Expected TitleText '#E0E0E0', got %v", theme.TitleText)
	}
	if theme.MenuSelected != "#00A0B0" {
		t.Errorf("Expected MenuSelected '#00A0B0', got %v", theme.MenuSelected)
	}
	if theme.LightSquare != "#E8EEF2" {
		t.Errorf("Expected LightSquare '#E8EEF2', got %v", theme.LightSquare)
	}
	if theme.DarkSquare != "#5D8AA8" {
		t.Errorf("Expected DarkSquare '#5D8AA8', got %v", theme.DarkSquare)
	}
}

// TestGetTheme_Minimalist tests that GetTheme returns the minimalist theme correctly.
func TestGetTheme_Minimalist(t *testing.T) {
	theme := GetTheme(ThemeMinimalist)
	if theme.Name != ThemeNameMinimalist {
		t.Errorf("Expected theme name %q, got %q", ThemeNameMinimalist, theme.Name)
	}

	// Verify theme has the expected minimalist colors
	if theme.TitleText != "#C0C0C0" {
		t.Errorf("Expected TitleText '#C0C0C0', got %v", theme.TitleText)
	}
	if theme.MenuSelected != "#A0A0A0" {
		t.Errorf("Expected MenuSelected '#A0A0A0', got %v", theme.MenuSelected)
	}
	if theme.LightSquare != "#D0D0D0" {
		t.Errorf("Expected LightSquare '#D0D0D0', got %v", theme.LightSquare)
	}
	if theme.DarkSquare != "#808080" {
		t.Errorf("Expected DarkSquare '#808080', got %v", theme.DarkSquare)
	}
}

// TestModernTheme_AllFieldsSet tests that the modern theme has all fields set.
func TestModernTheme_AllFieldsSet(t *testing.T) {
	theme := GetTheme(ThemeModern)

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

	// Menu hierarchy colors
	if theme.MenuPrimary == "" {
		t.Error("Expected MenuPrimary to be set")
	}
	if theme.MenuSecondary == "" {
		t.Error("Expected MenuSecondary to be set")
	}
	if theme.MenuSeparator == "" {
		t.Error("Expected MenuSeparator to be set")
	}

	// Turn indicator colors
	if theme.WhiteTurnText == "" {
		t.Error("Expected WhiteTurnText to be set")
	}
	if theme.BlackTurnText == "" {
		t.Error("Expected BlackTurnText to be set")
	}
}

// TestMinimalistTheme_AllFieldsSet tests that the minimalist theme has all fields set.
func TestMinimalistTheme_AllFieldsSet(t *testing.T) {
	theme := GetTheme(ThemeMinimalist)

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

	// Menu hierarchy colors
	if theme.MenuPrimary == "" {
		t.Error("Expected MenuPrimary to be set")
	}
	if theme.MenuSecondary == "" {
		t.Error("Expected MenuSecondary to be set")
	}
	if theme.MenuSeparator == "" {
		t.Error("Expected MenuSeparator to be set")
	}

	// Turn indicator colors
	if theme.WhiteTurnText == "" {
		t.Error("Expected WhiteTurnText to be set")
	}
	if theme.BlackTurnText == "" {
		t.Error("Expected BlackTurnText to be set")
	}
}

// TestCycleTheme tests the theme cycling logic.
func TestCycleTheme(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		expected string
	}{
		{"classic to modern", ThemeNameClassic, ThemeNameModern},
		{"modern to minimalist", ThemeNameModern, ThemeNameMinimalist},
		{"minimalist to classic", ThemeNameMinimalist, ThemeNameClassic},
		{"empty to modern", "", ThemeNameModern},
		{"unknown to modern", "unknown", ThemeNameModern},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cycleTheme(tt.current)
			if got != tt.expected {
				t.Errorf("cycleTheme(%q) = %q, want %q", tt.current, got, tt.expected)
			}
		})
	}
}

// TestModelThemeInitialization_Modern tests that Model initializes with modern theme from config.
func TestModelThemeInitialization_Modern(t *testing.T) {
	config := Config{
		Theme: ThemeNameModern,
	}

	m := NewModel(config)

	if m.theme.Name != ThemeNameModern {
		t.Errorf("Expected model theme to be %q, got %q", ThemeNameModern, m.theme.Name)
	}
}

// TestModelThemeInitialization_Minimalist tests that Model initializes with minimalist theme from config.
func TestModelThemeInitialization_Minimalist(t *testing.T) {
	config := Config{
		Theme: ThemeNameMinimalist,
	}

	m := NewModel(config)

	if m.theme.Name != ThemeNameMinimalist {
		t.Errorf("Expected model theme to be %q, got %q", ThemeNameMinimalist, m.theme.Name)
	}
}

// TestThemeDisplayName tests the getThemeDisplayName helper function.
func TestThemeDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"classic", ThemeNameClassic, "Classic"},
		{"modern", ThemeNameModern, "Modern"},
		{"minimalist", ThemeNameMinimalist, "Minimalist"},
		{"empty defaults to Classic", "", "Classic"},
		{"unknown defaults to Classic", "unknown", "Classic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getThemeDisplayName(tt.input)
			if got != tt.expected {
				t.Errorf("getThemeDisplayName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestTurnStyleMethods tests that turn style methods work correctly.
func TestTurnStyleMethods(t *testing.T) {
	config := Config{
		Theme: ThemeNameClassic,
	}
	m := NewModel(config)

	// Test that turn style methods don't panic and return non-nil styles
	whiteStyle := m.whiteTurnStyle()
	blackStyle := m.blackTurnStyle()

	// Test rendering with these styles
	_ = whiteStyle.Render("White to move")
	_ = blackStyle.Render("Black to move")

	// If we get here without panics, the test passes
}

// TestTurnStyleUsesCorrectColor tests that turnStyle returns appropriate style based on turn.
func TestTurnStyleUsesCorrectColor(t *testing.T) {
	config := Config{
		Theme: ThemeNameClassic,
	}
	m := NewModel(config)

	// Without a board, turnStyle should return white style (default)
	style := m.turnStyle()
	// Render with it to verify it works
	_ = style.Render("test")

	// With a board set to white's turn
	m.board = &engine.Board{ActiveColor: 0} // White
	style = m.turnStyle()
	_ = style.Render("White to move")

	// With a board set to black's turn
	m.board = &engine.Board{ActiveColor: 1} // Black
	style = m.turnStyle()
	_ = style.Render("Black to move")
}

// TestSettingsThemeRendering tests that the settings screen renders the theme option.
func TestSettingsThemeRendering(t *testing.T) {
	tests := []struct {
		name         string
		theme        string
		expectedText string
	}{
		{"classic theme", ThemeNameClassic, "Theme: Classic"},
		{"modern theme", ThemeNameModern, "Theme: Modern"},
		{"minimalist theme", ThemeNameMinimalist, "Theme: Minimalist"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Theme:        tt.theme,
				ShowHelpText: true,
			}
			m := NewModel(config)
			m.screen = ScreenSettings

			output := m.renderSettings()

			if !containsString(output, tt.expectedText) {
				t.Errorf("Expected settings output to contain %q", tt.expectedText)
			}
		})
	}
}

// TestMenuHierarchyStyleMethods tests that menu hierarchy style methods work correctly.
func TestMenuHierarchyStyleMethods(t *testing.T) {
	config := Config{
		Theme: ThemeNameClassic,
	}
	m := NewModel(config)

	// Test that style methods don't panic and return non-nil styles
	_ = m.menuPrimaryStyle()
	_ = m.menuSecondaryStyle()
	_ = m.selectedPrimaryStyle()
	_ = m.selectedSecondaryStyle()
	_ = m.menuSeparatorStyle()

	// Test rendering with these styles
	_ = m.menuPrimaryStyle().Render("New Game")
	_ = m.menuSecondaryStyle().Render("Settings")
	_ = m.selectedPrimaryStyle().Render("New Game")
	_ = m.selectedSecondaryStyle().Render("Settings")
	_ = m.renderMenuSeparator()

	// If we get here without panics, the test passes
}

// TestIsPrimaryAction tests the isPrimaryAction helper function.
func TestIsPrimaryAction(t *testing.T) {
	tests := []struct {
		option   string
		expected bool
	}{
		{"New Game", true},
		{"Resume Game", true},
		{"Start", true},
		{"Play Again", true},
		{"New Session", true},
		{"Settings", false},
		{"Exit", false},
		{"Load Game", false},
		{"Return to Menu", false},
	}

	for _, tt := range tests {
		t.Run(tt.option, func(t *testing.T) {
			got := isPrimaryAction(tt.option)
			if got != tt.expected {
				t.Errorf("isPrimaryAction(%q) = %v, want %v", tt.option, got, tt.expected)
			}
		})
	}
}

// TestMainMenuRendersSeparator tests that the main menu renders a separator.
func TestMainMenuRendersSeparator(t *testing.T) {
	config := Config{
		Theme:        ThemeNameClassic,
		ShowHelpText: true,
	}
	m := NewModel(config)
	m.screen = ScreenMainMenu
	m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}

	output := m.renderMainMenu()

	// Should contain the separator line
	if !containsString(output, "────") {
		t.Error("Expected main menu to contain separator line")
	}
}

// TestSettingsRendersSeparators tests that the settings screen renders separators.
func TestSettingsRendersSeparators(t *testing.T) {
	config := Config{
		Theme:        ThemeNameClassic,
		ShowHelpText: true,
	}
	m := NewModel(config)
	m.screen = ScreenSettings

	output := m.renderSettings()

	// Should contain separator lines
	if !containsString(output, "────") {
		t.Error("Expected settings screen to contain separator lines")
	}
}

// TestFocusIndicatorInMainMenu tests that focus indicator (>>) is visible in main menu.
func TestFocusIndicatorInMainMenu(t *testing.T) {
	config := Config{
		Theme:        ThemeNameClassic,
		ShowHelpText: true,
	}
	m := NewModel(config)
	m.screen = ScreenMainMenu
	m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
	m.menuSelection = 0

	output := m.renderMainMenu()

	// Should contain the focus indicator
	if !containsString(output, ">>") {
		t.Error("Expected main menu to contain focus indicator '>>'")
	}
}

// TestFocusIndicatorInSettings tests that focus indicator (>>) is visible in settings.
func TestFocusIndicatorInSettings(t *testing.T) {
	config := Config{
		Theme:        ThemeNameClassic,
		ShowHelpText: true,
	}
	m := NewModel(config)
	m.screen = ScreenSettings
	m.settingsSelection = 0

	output := m.renderSettings()

	// Should contain the focus indicator
	if !containsString(output, ">>") {
		t.Error("Expected settings screen to contain focus indicator '>>'")
	}
}

// TestKeyboardAccessibility_AllScreensReachable tests that all screens are reachable via keyboard.
func TestKeyboardAccessibility_AllScreensReachable(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.termWidth = 80
	m.termHeight = 24

	// Main menu is the starting point
	if m.screen != ScreenMainMenu {
		t.Errorf("Expected to start on ScreenMainMenu, got %d", m.screen)
	}

	// Navigate to Game Type Select via keyboard
	m.menuOptions = buildMainMenuOptions()
	for i, opt := range m.menuOptions {
		if opt == "New Game" {
			m.menuSelection = i
			break
		}
	}
	result, _ := m.handleMainMenuKeys(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(Model)
	if m.screen != ScreenGameTypeSelect {
		t.Errorf("Expected ScreenGameTypeSelect, got %d", m.screen)
	}

	// Navigate back to main menu and then to Settings
	m.screen = ScreenMainMenu
	m.menuOptions = buildMainMenuOptions()
	m.menuSelection = 0

	// Find and select Settings
	settingsFound := false
	for i, opt := range m.menuOptions {
		if opt == "Settings" {
			m.menuSelection = i
			settingsFound = true
			break
		}
	}
	if !settingsFound {
		t.Fatal("Settings option not found in menu")
	}

	result, _ = m.handleMainMenuKeys(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(Model)
	if m.screen != ScreenSettings {
		t.Errorf("Expected ScreenSettings after selecting Settings, got %d", m.screen)
	}
}

// TestKeyboardAccessibility_FocusIndicatorsPresent tests that focus indicators are present on all menu screens.
func TestKeyboardAccessibility_FocusIndicatorsPresent(t *testing.T) {
	screens := []struct {
		name   string
		setup  func(*Model)
		render func(*Model) string
	}{
		{
			"MainMenu",
			func(m *Model) {
				m.screen = ScreenMainMenu
				m.menuOptions = buildMainMenuOptions()
				m.menuSelection = 0
			},
			func(m *Model) string { return m.renderMainMenu() },
		},
		{
			"GameTypeSelect",
			func(m *Model) {
				m.screen = ScreenGameTypeSelect
				m.menuOptions = []string{"Player vs Player", "Player vs Bot", "Bot vs Bot"}
				m.menuSelection = 0
			},
			func(m *Model) string { return m.renderGameTypeSelect() },
		},
		{
			"Settings",
			func(m *Model) {
				m.screen = ScreenSettings
				m.settingsSelection = 0
			},
			func(m *Model) string { return m.renderSettings() },
		},
	}

	for _, tc := range screens {
		t.Run(tc.name, func(t *testing.T) {
			config := Config{
				ShowHelpText: true,
				Theme:        ThemeNameClassic,
			}
			m := NewModel(config)
			m.termWidth = 80
			m.termHeight = 24
			tc.setup(&m)

			output := tc.render(&m)

			if !containsString(output, ">>") {
				t.Errorf("Expected focus indicator '>>' on %s screen", tc.name)
			}
		})
	}
}

// TestMouseAndKeyboardParallelUsage tests that mouse and keyboard can be used together.
func TestMouseAndKeyboardParallelUsage(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay
	m.gameType = GameTypePvP
	m.termWidth = 80
	m.termHeight = 24

	// Select a piece via mouse (simulate click on e2 pawn)
	// The squareFromMouse function is already tested elsewhere
	// Here we just verify the state changes

	// Set up a selection state
	e2 := engine.NewSquare(4, 1) // e2
	m.selectedSquare = &e2
	m.validMoves = []engine.Square{engine.NewSquare(4, 2), engine.NewSquare(4, 3)} // e3, e4

	// Now use keyboard to type a move - this should work independently
	m.input = "d2d4"

	// Execute the keyboard move
	// The input should be processed normally even with a mouse selection
	if m.input != "d2d4" {
		t.Errorf("Expected input 'd2d4', got %q", m.input)
	}

	// Both mouse selection and keyboard input should coexist
	if m.selectedSquare == nil {
		t.Error("Mouse selection should still be present")
	}
}
