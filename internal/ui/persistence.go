package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/Mgrdich/TermChess/internal/engine"
)

// getConfigDir returns the path to the TermChess configuration directory.
// It returns ~/.termchess/ or an error if the home directory cannot be determined.
func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".termchess"), nil
}

// getConfigFilePath returns the full path to the configuration file.
func getConfigFilePath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.toml"), nil
}

// defaultConfigFile returns a ConfigFile with default values.
func defaultConfigFile() ConfigFile {
	return ConfigFile{
		Display: DisplayConfig{
			UseUnicode:      false, // ASCII for maximum compatibility
			ShowCoordinates: true,  // Show a-h, 1-8 labels
			UseColors:       true,  // Use colors if terminal supports
			ShowMoveHistory: false, // Hidden by default
		},
		Game: GameConfig{
			DefaultGameType:      "pvp",    // Default to player vs player
			DefaultBotDifficulty: "medium", // Default bot difficulty
		},
	}
}

// configFileToConfig converts a ConfigFile to a Config struct.
func configFileToConfig(cf ConfigFile) Config {
	return Config{
		UseUnicode:      cf.Display.UseUnicode,
		ShowCoords:      cf.Display.ShowCoordinates,
		UseColors:       cf.Display.UseColors,
		ShowMoveHistory: cf.Display.ShowMoveHistory,
		ShowHelpText:    cf.Display.ShowHelpText,
	}
}

// configToConfigFile converts a Config struct to a ConfigFile.
func configToConfigFile(c Config) ConfigFile {
	return ConfigFile{
		Display: DisplayConfig{
			UseUnicode:      c.UseUnicode,
			ShowCoordinates: c.ShowCoords,
			UseColors:       c.UseColors,
			ShowMoveHistory: c.ShowMoveHistory,
			ShowHelpText:    c.ShowHelpText,
		},
		Game: GameConfig{
			DefaultGameType:      "pvp",    // Preserve default
			DefaultBotDifficulty: "medium", // Preserve default
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

// SaveConfig writes the configuration to ~/.termchess/config.toml.
// It creates the ~/.termchess/ directory if it doesn't exist.
// Returns an error if the file cannot be written.
func SaveConfig(config Config) error {
	// Get the config directory path
	configDir, err := getConfigDir()
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

// SaveGamePath returns the full path to the save game file.
// Exported for testing purposes.
func SaveGamePath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "savegame.fen"), nil
}

// SaveGame saves the current game state to ~/.termchess/savegame.fen.
// It converts the board to FEN format and writes it to the file.
// Returns an error if the file cannot be written.
func SaveGame(board *engine.Board) error {
	// Get the save game file path
	savePath, err := SaveGamePath()
	if err != nil {
		return fmt.Errorf("failed to get save game path: %w", err)
	}

	// Get the config directory path
	configDir, err := getConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	// Create the config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Convert board to FEN
	fen := board.ToFEN()

	// Write FEN to file
	if err := os.WriteFile(savePath, []byte(fen), 0644); err != nil {
		return fmt.Errorf("failed to write save game file: %w", err)
	}

	return nil
}

// LoadGame loads a saved game from ~/.termchess/savegame.fen.
// It reads the FEN from the file and creates a Board from it.
// Returns an error if the file cannot be read or the FEN is invalid.
func LoadGame() (*engine.Board, error) {
	// Get the save game file path
	savePath, err := SaveGamePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get save game path: %w", err)
	}

	// Read the FEN from file
	data, err := os.ReadFile(savePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read save game file: %w", err)
	}

	// Parse FEN and create board
	board, err := engine.FromFEN(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse saved game FEN: %w", err)
	}

	return board, nil
}

// DeleteSaveGame deletes the saved game file at ~/.termchess/savegame.fen.
// Returns nil if the file doesn't exist (not an error condition).
// Returns an error only if deletion fails.
func DeleteSaveGame() error {
	// Get the save game file path
	savePath, err := SaveGamePath()
	if err != nil {
		return fmt.Errorf("failed to get save game path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		// File doesn't exist, nothing to delete
		return nil
	}

	// Delete the file
	if err := os.Remove(savePath); err != nil {
		return fmt.Errorf("failed to delete save game file: %w", err)
	}

	return nil
}

// SaveGameExists checks if a saved game file exists at ~/.termchess/savegame.fen.
// Returns true if the file exists, false otherwise.
func SaveGameExists() bool {
	savePath, err := SaveGamePath()
	if err != nil {
		return false
	}

	_, err = os.Stat(savePath)
	return err == nil
}
