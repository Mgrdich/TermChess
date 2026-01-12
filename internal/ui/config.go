package ui

import (
	"os"
	"path/filepath"
)

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
	// ShowHelpText determines whether to display navigation help text at the bottom of screens
	ShowHelpText bool
}

// DefaultConfig returns a Config with default values for maximum compatibility
// and user-friendliness.
func DefaultConfig() Config {
	return Config{
		UseUnicode:      false, // ASCII for maximum compatibility (change to true to test Unicode)
		ShowCoords:      true,  // Show a-h, 1-8 labels
		UseColors:       true,  // Use colors if terminal supports
		ShowMoveHistory: false, // Hidden by default
		ShowHelpText:    true,  // Show help text by default
	}
}

// ConfigFile represents the structure of the TOML configuration file.
// It uses separate sections for display and game settings.
type ConfigFile struct {
	Display DisplayConfig `toml:"display"`
	Game    GameConfig    `toml:"game"`
}

// DisplayConfig holds display-related configuration options for the TOML file.
type DisplayConfig struct {
	UseUnicode      bool `toml:"use_unicode"`
	ShowCoordinates bool `toml:"show_coordinates"`
	UseColors       bool `toml:"use_colors"`
	ShowMoveHistory bool `toml:"show_move_history"`
	ShowHelpText    bool `toml:"show_help_text"`
}

// GameConfig holds game-related configuration options for the TOML file.
type GameConfig struct {
	DefaultGameType      string `toml:"default_game_type"`
	DefaultBotDifficulty string `toml:"default_bot_difficulty"`
}

// GetConfigPath returns the absolute path to the configuration file.
// The config file is stored at ~/.termchess/config.toml
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".termchess", "config.toml"), nil
}
