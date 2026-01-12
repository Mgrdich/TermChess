package ui

import (
	"os"
	"path/filepath"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// SaveGamePath returns the absolute path to the savegame file.
// The savegame is stored at ~/.termchess/savegame.fen on macOS/Linux
// and %USERPROFILE%\.termchess\savegame.fen on Windows.
func SaveGamePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".termchess", "savegame.fen"), nil
}

// SaveGame saves the current board state as FEN to the savegame file.
// It creates the ~/.termchess/ directory if it doesn't exist.
// Returns an error if the save operation fails.
func SaveGame(b *engine.Board) error {
	savePath, err := SaveGamePath()
	if err != nil {
		return err
	}

	// Create the .termchess directory if it doesn't exist
	saveDir := filepath.Dir(savePath)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return err
	}

	// Convert board to FEN
	fen := b.ToFEN()

	// Write FEN to file with proper permissions (0644)
	if err := os.WriteFile(savePath, []byte(fen), 0644); err != nil {
		return err
	}

	return nil
}

// LoadGame loads a saved game from the savegame file.
// Returns the board state or an error if the file cannot be read or parsed.
func LoadGame() (*engine.Board, error) {
	savePath, err := SaveGamePath()
	if err != nil {
		return nil, err
	}

	// Read the FEN string from file
	data, err := os.ReadFile(savePath)
	if err != nil {
		return nil, err
	}

	// Parse the FEN string to create a board
	board, err := engine.FromFEN(string(data))
	if err != nil {
		return nil, err
	}

	return board, nil
}

// DeleteSaveGame deletes the savegame file if it exists.
// This is typically called after a game ends normally.
// Returns nil if the file doesn't exist or was successfully deleted.
func DeleteSaveGame() error {
	savePath, err := SaveGamePath()
	if err != nil {
		return err
	}

	// Check if file exists
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		// File doesn't exist, nothing to delete
		return nil
	}

	// Delete the file
	return os.Remove(savePath)
}

// SaveGameExists checks if a savegame file exists.
// Returns true if the file exists and can be accessed, false otherwise.
func SaveGameExists() bool {
	savePath, err := SaveGamePath()
	if err != nil {
		return false
	}

	// Check if file exists
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		return false
	}

	return true
}
