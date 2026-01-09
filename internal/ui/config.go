package ui

// Config holds display configuration options that control how the UI is rendered.
type Config struct {
	// UseUnicode determines whether to use Unicode chess pieces (♔♕) or ASCII (K, Q)
	UseUnicode bool
	// ShowCoords determines whether to show file/rank labels (a-h, 1-8)
	ShowCoords bool
	// UseColors determines whether to color piece symbols
	UseColors bool
	// ShowMoveHistory determines whether to display the move history panel
	ShowMoveHistory bool
}

// DefaultConfig returns a Config with default values for maximum compatibility
// and user-friendliness.
func DefaultConfig() Config {
	return Config{
		UseUnicode:      false, // ASCII for maximum compatibility (change to true to test Unicode)
		ShowCoords:      true,  // Show a-h, 1-8 labels
		UseColors:       true,  // Use colors if terminal supports
		ShowMoveHistory: false, // Hidden by default
	}
}
