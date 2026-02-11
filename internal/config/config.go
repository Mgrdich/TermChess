// Package config provides configuration and game state persistence for TermChess.
//
// Configuration files are stored in ~/.termchess/ and use TOML format.
// Game saves are stored as FEN strings in ~/.termchess/savegame.fen.
//
// The package provides:
//   - Config types and default values
//   - Config file loading and saving
//   - Game state save/load/delete operations
//   - Path helpers for config directory and files
//
// Config directory permissions: 0755 (rwxr-xr-x)
// Config file permissions: 0644 (rw-r--r--)
// Save game file permissions: 0644 (rw-r--r--)
package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// DefaultTheme is the default theme name.
// Valid theme values are: "classic", "modern", "minimalist"
// These must match the ui.ThemeNameX constants.
// Invalid theme values will be normalized to DefaultTheme by ui.ParseThemeName.
const DefaultTheme = "classic"

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
	// Theme is the name of the color theme to use (e.g., "classic")
	Theme string
}

// DefaultConfig returns a Config with default values for maximum compatibility
// and user-friendliness.
func DefaultConfig() Config {
	return Config{
		UseUnicode:      false,     // ASCII for maximum compatibility (change to true to test Unicode)
		ShowCoords:      true,      // Show a-h, 1-8 labels
		UseColors:       true,      // Use colors if terminal supports
		ShowMoveHistory: false,     // Hidden by default
		ShowHelpText:    true,      // Show help text by default
		Theme:           DefaultTheme, // Classic theme by default
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
	UseUnicode      bool   `toml:"use_unicode"`
	ShowCoordinates bool   `toml:"show_coordinates"`
	UseColors       bool   `toml:"use_colors"`
	ShowMoveHistory bool   `toml:"show_move_history"`
	ShowHelpText    bool   `toml:"show_help_text"`
	Theme           string `toml:"theme"`
}

// GameConfig holds game-related configuration options for the TOML file.
type GameConfig struct {
	DefaultGameType      string `toml:"default_game_type"`
	DefaultBotDifficulty string `toml:"default_bot_difficulty"`
	// BvBConcurrency controls how many Bot vs Bot games run simultaneously.
	// 0 = auto-detect based on CPU count, positive values specify exact count.
	BvBConcurrency int `toml:"bvb_concurrency"`
	// BvBDefaultViewMode specifies the default view mode for Bot vs Bot sessions.
	// Valid values: "grid", "single", "stats_only"
	BvBDefaultViewMode string `toml:"bvb_default_view_mode"`
}

// defaultConfigFile returns a ConfigFile with default values.
func defaultConfigFile() ConfigFile {
	return ConfigFile{
		Display: DisplayConfig{
			UseUnicode:      false,     // ASCII for maximum compatibility
			ShowCoordinates: true,      // Show a-h, 1-8 labels
			UseColors:       true,      // Use colors if terminal supports
			ShowMoveHistory: false,     // Hidden by default
			Theme:           DefaultTheme, // Classic theme by default
		},
		Game: GameConfig{
			DefaultGameType:      "pvp",    // Default to player vs player
			DefaultBotDifficulty: "medium", // Default bot difficulty
			BvBDefaultViewMode:   "grid",   // Default to grid view for BvB
		},
	}
}

// configFileToConfig converts a ConfigFile to a Config struct.
func configFileToConfig(cf ConfigFile) Config {
	theme := cf.Display.Theme
	if theme == "" {
		theme = DefaultTheme
	}
	return Config{
		UseUnicode:      cf.Display.UseUnicode,
		ShowCoords:      cf.Display.ShowCoordinates,
		UseColors:       cf.Display.UseColors,
		ShowMoveHistory: cf.Display.ShowMoveHistory,
		ShowHelpText:    cf.Display.ShowHelpText,
		Theme:           theme,
	}
}

// configToConfigFile converts a Config struct to a ConfigFile.
func configToConfigFile(c Config) ConfigFile {
	theme := c.Theme
	if theme == "" {
		theme = DefaultTheme
	}
	return ConfigFile{
		Display: DisplayConfig{
			UseUnicode:      c.UseUnicode,
			ShowCoordinates: c.ShowCoords,
			UseColors:       c.UseColors,
			ShowMoveHistory: c.ShowMoveHistory,
			ShowHelpText:    c.ShowHelpText,
			Theme:           theme,
		},
		Game: GameConfig{
			DefaultGameType:      "pvp",    // Preserve default
			DefaultBotDifficulty: "medium", // Preserve default
			BvBDefaultViewMode:   "grid",   // Preserve default
		},
	}
}

// LoadConfig reads the configuration file from ~/.termchess/config.toml.
// If the file doesn't exist or cannot be parsed, it returns the default configuration.
// This function never returns an error - it always returns a valid configuration.
func LoadConfig() Config {
	configPath, err := getConfigFilePath()
	if err != nil {
		// Cannot determine config path, use defaults
		return DefaultConfig()
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, use defaults
		return DefaultConfig()
	}

	// Read and parse the config file
	var cf ConfigFile
	if _, err := toml.DecodeFile(configPath, &cf); err != nil {
		// Failed to parse config file, use defaults
		return DefaultConfig()
	}

	// Convert ConfigFile to Config and return
	return configFileToConfig(cf)
}

// LoadGameConfig reads the game configuration from ~/.termchess/config.toml.
// If the file doesn't exist or cannot be parsed, it returns the default game configuration.
// This function never returns an error - it always returns a valid configuration.
func LoadGameConfig() GameConfig {
	configPath, err := getConfigFilePath()
	if err != nil {
		// Cannot determine config path, use defaults
		return defaultConfigFile().Game
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, use defaults
		return defaultConfigFile().Game
	}

	// Read and parse the config file
	var cf ConfigFile
	if _, err := toml.DecodeFile(configPath, &cf); err != nil {
		// Failed to parse config file, use defaults
		return defaultConfigFile().Game
	}

	return cf.Game
}

// SaveConfig writes the configuration to ~/.termchess/config.toml.
// It creates the ~/.termchess/ directory if it doesn't exist.
// Returns an error if the file cannot be written.
func SaveConfig(config Config) error {
	// Get the config directory path
	configDir, err := GetConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	// Create the config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Get the config file path
	configPath, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("failed to get config file path: %w", err)
	}

	// Convert Config to ConfigFile
	cf := configToConfigFile(config)

	// Create the config file
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	// Encode the config to TOML and write to file
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(cf); err != nil {
		return fmt.Errorf("failed to encode config to TOML: %w", err)
	}

	return nil
}
