package config

import (
	"fmt"
	"os"
	"path/filepath"
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

// SaveGamePath returns the full path to the save game file.
// Exported for testing purposes.
func SaveGamePath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "savegame.fen"), nil
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
