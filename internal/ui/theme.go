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
	ThemeNameClassic    = "classic"
	ThemeNameModern     = "modern"
	ThemeNameMinimalist = "minimalist"
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

	// Menu hierarchy colors for visual distinction
	MenuPrimary   lipgloss.Color // For primary actions (New Game, Start, Resume)
	MenuSecondary lipgloss.Color // For secondary actions (Settings, Load Game, Exit)
	MenuSeparator lipgloss.Color // For visual separators between menu groups

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

		// Menu hierarchy colors
		MenuPrimary:   lipgloss.Color("#FAFAFA"), // Bright white for primary actions
		MenuSecondary: lipgloss.Color("#A0A0A0"), // Dimmer for secondary actions
		MenuSeparator: lipgloss.Color("#444444"), // Dark gray for separators

		// Turn indicator colors (for future use)
		WhiteTurnText: lipgloss.Color("#FAFAFA"), // White
		BlackTurnText: lipgloss.Color("#626262"), // Gray
	},
	ThemeModern: {
		Name: ThemeNameModern,

		// Board colors - blues and teals for a modern aesthetic
		LightSquare: lipgloss.Color("#E8EEF2"), // Light gray-blue
		DarkSquare:  lipgloss.Color("#5D8AA8"), // Steel blue
		WhitePiece:  lipgloss.Color("#FFFFFF"), // Pure white for white pieces
		BlackPiece:  lipgloss.Color("#1A1A2E"), // Dark navy for black pieces

		// Selection colors
		SelectedHighlight:  lipgloss.Color("#00A0B0"), // Teal
		ValidMoveHighlight: lipgloss.Color("#4ECDC4"), // Light teal

		// UI colors - clean modern look with blues and teals
		BoardBorder:  lipgloss.Color("#B8C5D0"), // Light steel
		MenuSelected: lipgloss.Color("#00A0B0"), // Teal
		MenuNormal:   lipgloss.Color("#E0E0E0"), // Light gray
		TitleText:    lipgloss.Color("#E0E0E0"), // Light gray (WCAG AA on dark bg)
		HelpText:     lipgloss.Color("#8899A6"), // Muted blue-gray
		ErrorText:    lipgloss.Color("#E74C3C"), // Modern red
		StatusText:   lipgloss.Color("#4ECDC4"), // Light teal

		// Menu hierarchy colors
		MenuPrimary:   lipgloss.Color("#E0E0E0"), // Bright for primary actions
		MenuSecondary: lipgloss.Color("#8899A6"), // Muted for secondary actions
		MenuSeparator: lipgloss.Color("#3D4F5F"), // Dark steel for separators

		// Turn indicator colors
		WhiteTurnText: lipgloss.Color("#E0E0E0"), // Light gray
		BlackTurnText: lipgloss.Color("#8899A6"), // Muted blue-gray
	},
	ThemeMinimalist: {
		Name: ThemeNameMinimalist,

		// Board colors - simple grayscale for distraction-free play
		LightSquare: lipgloss.Color("#D0D0D0"), // Light gray
		DarkSquare:  lipgloss.Color("#808080"), // Medium gray
		WhitePiece:  lipgloss.Color("#FFFFFF"), // Pure white for white pieces
		BlackPiece:  lipgloss.Color("#2D2D2D"), // Dark gray for black pieces

		// Selection colors - subtle accents
		SelectedHighlight:  lipgloss.Color("#A0A0A0"), // Gray
		ValidMoveHighlight: lipgloss.Color("#B8B8B8"), // Light gray

		// UI colors - muted grayscale palette
		BoardBorder:  lipgloss.Color("#A0A0A0"), // Gray
		MenuSelected: lipgloss.Color("#A0A0A0"), // Gray
		MenuNormal:   lipgloss.Color("#C0C0C0"), // Light gray
		TitleText:    lipgloss.Color("#C0C0C0"), // Light gray (WCAG AA on dark bg)
		HelpText:     lipgloss.Color("#707070"), // Medium gray
		ErrorText:    lipgloss.Color("#CC6666"), // Muted red
		StatusText:   lipgloss.Color("#88AA88"), // Muted green

		// Menu hierarchy colors
		MenuPrimary:   lipgloss.Color("#C0C0C0"), // Bright for primary actions
		MenuSecondary: lipgloss.Color("#888888"), // Dimmer for secondary actions
		MenuSeparator: lipgloss.Color("#505050"), // Dark gray for separators

		// Turn indicator colors
		WhiteTurnText: lipgloss.Color("#C0C0C0"), // Light gray
		BlackTurnText: lipgloss.Color("#707070"), // Medium gray
	},
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
