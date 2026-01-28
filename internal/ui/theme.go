package ui

import "github.com/charmbracelet/lipgloss"

// ThemeName represents a named color theme.
type ThemeName int

const (
	// ThemeClassic is the default theme with original hardcoded colors.
	ThemeClassic ThemeName = iota
	// ThemeModern is a modern theme (placeholder for future implementation).
	ThemeModern
	// ThemeMinimalist is a minimalist theme (placeholder for future implementation).
	ThemeMinimalist
)

// Theme name string constants for serialization and comparison.
// These are the only valid string values for theme names.
const (
	ThemeNameClassic     = "classic"
	ThemeNameModern      = "modern"
	ThemeNameMinimalist  = "minimalist"
)

// String returns the string representation of the theme name.
func (t ThemeName) String() string {
	switch t {
	case ThemeModern:
		return ThemeNameModern
	case ThemeMinimalist:
		return ThemeNameMinimalist
	default:
		return ThemeNameClassic
	}
}

// ParseThemeName converts a string to a ThemeName.
// If the string is not recognized, it returns ThemeClassic as the default.
func ParseThemeName(s string) ThemeName {
	switch s {
	case ThemeNameModern:
		return ThemeModern
	case ThemeNameMinimalist:
		return ThemeMinimalist
	default:
		return ThemeClassic
	}
}

// Theme defines all color values used throughout the UI.
// Themes should use WCAG AA compliant colors (4.5:1 contrast ratio for text).
type Theme struct {
	Name string

	// Board colors
	LightSquare lipgloss.Color
	DarkSquare  lipgloss.Color
	WhitePiece  lipgloss.Color
	BlackPiece  lipgloss.Color

	// Selection colors (for future use)
	SelectedHighlight  lipgloss.Color
	ValidMoveHighlight lipgloss.Color

	// UI colors
	BoardBorder  lipgloss.Color
	MenuSelected lipgloss.Color
	MenuNormal   lipgloss.Color
	TitleText    lipgloss.Color
	HelpText     lipgloss.Color
	ErrorText    lipgloss.Color
	StatusText   lipgloss.Color

	// Turn indicator colors (for future use)
	WhiteTurnText lipgloss.Color
	BlackTurnText lipgloss.Color
}

// themes is a map of ThemeName to Theme, providing type-safe theme access.
// Themes are only accessible via GetTheme(ThemeName).
var themes = map[ThemeName]Theme{
	ThemeClassic: {
		Name: ThemeNameClassic,

		// Board colors - using terminal color codes for broad compatibility
		LightSquare: lipgloss.Color("15"), // Bright white (terminal color 15)
		DarkSquare:  lipgloss.Color("8"),  // Gray (terminal color 8)
		WhitePiece:  lipgloss.Color("15"), // Bright white for white pieces
		BlackPiece:  lipgloss.Color("8"),  // Gray for black pieces

		// Selection colors (for future use)
		SelectedHighlight:  lipgloss.Color("#7D56F4"), // Purple - matches cursor
		ValidMoveHighlight: lipgloss.Color("#50FA7B"), // Green - matches status

		// UI colors - matching original hardcoded values
		BoardBorder:  lipgloss.Color("#FAFAFA"), // White - matches title
		MenuSelected: lipgloss.Color("#7D56F4"), // Purple
		MenuNormal:   lipgloss.Color("#FFFDF5"), // Off-white/cream
		TitleText:    lipgloss.Color("#FAFAFA"), // White
		HelpText:     lipgloss.Color("#626262"), // Gray
		ErrorText:    lipgloss.Color("#FF5555"), // Red
		StatusText:   lipgloss.Color("#50FA7B"), // Green

		// Turn indicator colors (for future use)
		WhiteTurnText: lipgloss.Color("#FAFAFA"), // White
		BlackTurnText: lipgloss.Color("#626262"), // Gray
	},
	// ThemeModern and ThemeMinimalist will be added in Slice 2
}

// GetTheme returns the theme for the given ThemeName.
// If the theme is not found, it returns the classic theme as the default.
// This is the only way to access themes, ensuring type-safe theme selection.
func GetTheme(name ThemeName) Theme {
	if theme, ok := themes[name]; ok {
		return theme
	}
	// Default to classic theme for unknown/unimplemented themes
	return themes[ThemeClassic]
}
