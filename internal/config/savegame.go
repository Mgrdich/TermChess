package config

import (
	"fmt"
	"os"

	"github.com/Mgrdich/TermChess/internal/engine"
)

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
	configDir, err := GetConfigDir()
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
