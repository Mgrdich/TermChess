package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// TestSaveGamePath tests that SaveGamePath returns a valid path
func TestSaveGamePath(t *testing.T) {
	path, err := SaveGamePath()
	if err != nil {
		t.Fatalf("SaveGamePath returned error: %v", err)
	}

	if path == "" {
		t.Fatal("SaveGamePath returned empty string")
	}

	// Check that path contains .termchess directory
	if !strings.Contains(path, ".termchess") {
		t.Errorf("SaveGamePath %q does not contain .termchess", path)
	}

	// Check that path ends with savegame.fen
	if !strings.HasSuffix(path, "savegame.fen") {
		t.Errorf("SaveGamePath %q does not end with savegame.fen", path)
	}
}

// TestSaveGame tests saving a board to file
func TestSaveGame(t *testing.T) {
	// Create a board with a known position
	board := engine.NewBoard()

	// Save the board
	err := SaveGame(board)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Verify file exists
	path, _ := SaveGamePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Savegame file was not created at %s", path)
	}

	// Read the file and verify it contains valid FEN
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read savegame file: %v", err)
	}

	fenStr := string(data)
	if fenStr == "" {
		t.Fatal("Savegame file is empty")
	}

	// Verify it's a valid FEN by trying to parse it
	_, err = engine.FromFEN(fenStr)
	if err != nil {
		t.Fatalf("Savegame contains invalid FEN: %v", err)
	}

	// Clean up
	os.Remove(path)
}

// TestSaveGameCreatesDirectory tests that SaveGame creates the .termchess directory
func TestSaveGameCreatesDirectory(t *testing.T) {
	// Get the .termchess directory path
	path, _ := SaveGamePath()
	saveDir := filepath.Dir(path)

	// Remove the directory if it exists (to test creation)
	os.RemoveAll(saveDir)

	// Create a board and save it
	board := engine.NewBoard()
	err := SaveGame(board)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		t.Fatalf("SaveGame did not create .termchess directory at %s", saveDir)
	}

	// Clean up
	os.Remove(path)
}

// TestLoadGame tests loading a saved game
func TestLoadGame(t *testing.T) {
	// Create a board with a known position (after 1.e4)
	originalBoard := engine.NewBoard()
	move, _ := engine.ParseMove("e2e4")
	originalBoard.MakeMove(move)

	// Save the board
	err := SaveGame(originalBoard)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Load the board
	loadedBoard, err := LoadGame()
	if err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}

	// Verify the loaded board matches the original
	if loadedBoard.ToFEN() != originalBoard.ToFEN() {
		t.Errorf("Loaded board FEN does not match original.\nExpected: %s\nGot: %s",
			originalBoard.ToFEN(), loadedBoard.ToFEN())
	}

	// Clean up
	path, _ := SaveGamePath()
	os.Remove(path)
}

// TestLoadGameNonExistent tests loading when no save file exists
func TestLoadGameNonExistent(t *testing.T) {
	// Ensure no save file exists
	path, _ := SaveGamePath()
	os.Remove(path)

	// Try to load - should return error
	_, err := LoadGame()
	if err == nil {
		t.Fatal("LoadGame should return error when file doesn't exist")
	}
}

// TestLoadGameInvalidFEN tests loading a file with invalid FEN
func TestLoadGameInvalidFEN(t *testing.T) {
	// Write invalid FEN to save file
	path, _ := SaveGamePath()
	saveDir := filepath.Dir(path)
	os.MkdirAll(saveDir, 0755)

	err := os.WriteFile(path, []byte("invalid fen string"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Try to load - should return error
	_, err = LoadGame()
	if err == nil {
		t.Fatal("LoadGame should return error for invalid FEN")
	}

	// Clean up
	os.Remove(path)
}

// TestSaveLoadRoundTrip tests that save and load preserve the game state
func TestSaveLoadRoundTrip(t *testing.T) {
	// Create a board and make several moves
	board := engine.NewBoard()
	moves := []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4"}

	for _, moveStr := range moves {
		move, err := engine.ParseMove(moveStr)
		if err != nil {
			t.Fatalf("Failed to parse move %s: %v", moveStr, err)
		}
		err = board.MakeMove(move)
		if err != nil {
			t.Fatalf("Failed to make move %s: %v", moveStr, err)
		}
	}

	originalFEN := board.ToFEN()

	// Save the board
	err := SaveGame(board)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Load the board
	loadedBoard, err := LoadGame()
	if err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}

	loadedFEN := loadedBoard.ToFEN()

	// Verify FEN strings match
	if originalFEN != loadedFEN {
		t.Errorf("Round-trip FEN mismatch.\nOriginal: %s\nLoaded:   %s",
			originalFEN, loadedFEN)
	}

	// Verify specific board properties
	if board.ActiveColor != loadedBoard.ActiveColor {
		t.Errorf("ActiveColor mismatch: expected %d, got %d",
			board.ActiveColor, loadedBoard.ActiveColor)
	}

	if board.CastlingRights != loadedBoard.CastlingRights {
		t.Errorf("CastlingRights mismatch: expected %d, got %d",
			board.CastlingRights, loadedBoard.CastlingRights)
	}

	if board.EnPassantSq != loadedBoard.EnPassantSq {
		t.Errorf("EnPassantSq mismatch: expected %d, got %d",
			board.EnPassantSq, loadedBoard.EnPassantSq)
	}

	if board.HalfMoveClock != loadedBoard.HalfMoveClock {
		t.Errorf("HalfMoveClock mismatch: expected %d, got %d",
			board.HalfMoveClock, loadedBoard.HalfMoveClock)
	}

	if board.FullMoveNum != loadedBoard.FullMoveNum {
		t.Errorf("FullMoveNum mismatch: expected %d, got %d",
			board.FullMoveNum, loadedBoard.FullMoveNum)
	}

	// Clean up
	path, _ := SaveGamePath()
	os.Remove(path)
}

// TestDeleteSaveGame tests deleting the save file
func TestDeleteSaveGame(t *testing.T) {
	// Create and save a game
	board := engine.NewBoard()
	err := SaveGame(board)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Verify file exists
	path, _ := SaveGamePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Savegame file was not created")
	}

	// Delete the save
	err = DeleteSaveGame()
	if err != nil {
		t.Fatalf("DeleteSaveGame failed: %v", err)
	}

	// Verify file no longer exists
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("Savegame file still exists after deletion")
	}
}

// TestDeleteSaveGameNonExistent tests deleting when no save file exists
func TestDeleteSaveGameNonExistent(t *testing.T) {
	// Ensure no save file exists
	path, _ := SaveGamePath()
	os.Remove(path)

	// Delete should not return error
	err := DeleteSaveGame()
	if err != nil {
		t.Fatalf("DeleteSaveGame should not error when file doesn't exist: %v", err)
	}
}

// TestSaveGameExists tests checking if a save file exists
func TestSaveGameExists(t *testing.T) {
	// Ensure no save file exists initially
	path, _ := SaveGamePath()
	os.Remove(path)

	// Should return false
	if SaveGameExists() {
		t.Fatal("SaveGameExists should return false when no save file exists")
	}

	// Create a save file
	board := engine.NewBoard()
	err := SaveGame(board)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Should return true
	if !SaveGameExists() {
		t.Fatal("SaveGameExists should return true when save file exists")
	}

	// Clean up
	os.Remove(path)
}

// TestSaveGameFilePermissions tests that the save file has correct permissions
func TestSaveGameFilePermissions(t *testing.T) {
	// Create and save a game
	board := engine.NewBoard()
	err := SaveGame(board)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Check file permissions
	path, _ := SaveGamePath()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Failed to stat save file: %v", err)
	}

	// Check that file is readable by owner (at minimum)
	mode := info.Mode()
	if mode&0400 == 0 {
		t.Errorf("Save file is not readable by owner: %v", mode)
	}

	// Clean up
	os.Remove(path)
}
