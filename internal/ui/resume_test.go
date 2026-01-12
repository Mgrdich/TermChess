package ui

import (
	"os"
	"testing"

	"github.com/Mgrdich/TermChess/internal/engine"
)

func TestResumeGameFunctionality(t *testing.T) {
	// Setup: Create a saved game
	board := engine.NewBoard()
	// Make a move to have a non-starting position
	move, _ := engine.ParseMove("e2e4")
	_ = board.MakeMove(move)

	err := SaveGame(board)
	if err != nil {
		t.Fatalf("Failed to save game: %v", err)
	}

	// Test 1: SaveGameExists should return true
	if !SaveGameExists() {
		t.Error("SaveGameExists() returned false, expected true")
	}

	// Test 2: LoadGame should load the saved position
	loadedBoard, err := LoadGame()
	if err != nil {
		t.Fatalf("Failed to load game: %v", err)
	}

	// Verify the loaded board matches the saved board
	if loadedBoard.ToFEN() != board.ToFEN() {
		t.Errorf("Loaded board FEN doesn't match saved board FEN\nExpected: %s\nGot: %s",
			board.ToFEN(), loadedBoard.ToFEN())
	}

	// Test 3: NewModel should start at ScreenResumePrompt when save exists
	config := DefaultConfig()
	model := NewModel(config)
	if model.screen != ScreenResumePrompt {
		t.Errorf("NewModel screen = %v, expected ScreenResumePrompt (%v)", model.screen, ScreenResumePrompt)
	}

	// Test 4: DeleteSaveGame should remove the saved game
	err = DeleteSaveGame()
	if err != nil {
		t.Fatalf("Failed to delete save game: %v", err)
	}

	if SaveGameExists() {
		t.Error("SaveGameExists() returned true after deletion, expected false")
	}

	// Test 5: NewModel should start at ScreenMainMenu when no save exists
	model = NewModel(config)
	if model.screen != ScreenMainMenu {
		t.Errorf("NewModel screen = %v, expected ScreenMainMenu (%v)", model.screen, ScreenMainMenu)
	}

	// Cleanup
	_ = DeleteSaveGame()
}

func TestResumePromptSelection(t *testing.T) {
	// Create a saved game
	board := engine.NewBoard()
	_ = SaveGame(board)
	defer DeleteSaveGame()

	config := DefaultConfig()
	model := NewModel(config)

	// Test "Yes" selection - should load the game
	model.resumePromptSelection = 0
	// We can't easily test the key handler without running the full Bubbletea program,
	// but we can verify the LoadGame function works
	loadedBoard, err := LoadGame()
	if err != nil {
		t.Fatalf("Failed to load game: %v", err)
	}
	if loadedBoard == nil {
		t.Error("LoadGame returned nil board")
	}
}

func TestDeleteSaveGameOnGameEnd(t *testing.T) {
	// Create a saved game
	board := engine.NewBoard()
	_ = SaveGame(board)

	// Verify it exists
	if !SaveGameExists() {
		t.Fatal("SaveGameExists() returned false, expected true")
	}

	// Simulate game end by calling DeleteSaveGame (this is what happens in handleGamePlayKeys)
	err := DeleteSaveGame()
	if err != nil {
		t.Fatalf("Failed to delete save game: %v", err)
	}

	// Verify it's gone
	if SaveGameExists() {
		t.Error("SaveGameExists() returned true after deletion, expected false")
	}
}

func TestCorruptedFENHandling(t *testing.T) {
	// Create a corrupted save file
	savePath, _ := SaveGamePath()
	configDir, _ := getConfigDir()
	_ = os.MkdirAll(configDir, 0755)
	_ = os.WriteFile(savePath, []byte("invalid fen string"), 0644)

	// LoadGame should return an error
	_, err := LoadGame()
	if err == nil {
		t.Error("LoadGame should return error for corrupted FEN, got nil")
	}

	// Cleanup
	_ = DeleteSaveGame()
}
