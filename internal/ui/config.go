package ui

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
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

// LoadConfig reads the configuration from ~/.termchess/config.toml.
// If the file doesn't exist or cannot be parsed, it returns the default configuration.
// This ensures the application always has valid configuration values.
func LoadConfig() Config {
	configPath, err := GetConfigPath()
	if err != nil {
		// Cannot determine config path, return defaults
		return DefaultConfig()
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, return defaults
		return DefaultConfig()
	}

	// Read the config file
	var configFile ConfigFile
	if _, err := toml.DecodeFile(configPath, &configFile); err != nil {
		// Failed to parse config file, return defaults
		return DefaultConfig()
	}

	// Convert ConfigFile to Config
	config := Config{
		UseUnicode:      configFile.Display.UseUnicode,
		ShowCoords:      configFile.Display.ShowCoordinates,
		UseColors:       configFile.Display.UseColors,
		ShowMoveHistory: configFile.Display.ShowMoveHistory,
		ShowHelpText:    configFile.Display.ShowHelpText,
	}

	// Apply defaults for new fields if they're missing (backward compatibility)
	// If ShowHelpText is false and all other fields are also false, it's likely
	// an old config file without ShowHelpText, so we default it to true
	if !config.ShowHelpText && !config.UseUnicode && !config.ShowCoords && !config.UseColors && !config.ShowMoveHistory {
		config.ShowHelpText = true
	}

	return config
}

// SaveConfig writes the current configuration to ~/.termchess/config.toml.
// It creates the ~/.termchess/ directory if it doesn't exist.
// Returns an error if the file cannot be written.
func SaveConfig(config Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create the .termchess directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Convert Config to ConfigFile for TOML serialization
	configFile := ConfigFile{
		Display: DisplayConfig{
			UseUnicode:      config.UseUnicode,
			ShowCoordinates: config.ShowCoords,
			UseColors:       config.UseColors,
			ShowMoveHistory: config.ShowMoveHistory,
			ShowHelpText:    config.ShowHelpText,
		},
		Game: GameConfig{
			DefaultGameType:      "pvp",
			DefaultBotDifficulty: "medium",
		},
	}

	// Create the config file
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Set proper file permissions
	if err := file.Chmod(0644); err != nil {
		return err
	}

	// Encode the config to TOML and write to file
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(configFile); err != nil {
		return err
	}

	return nil
}
