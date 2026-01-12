package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Mgrdich/TermChess/internal/config"
	"github.com/Mgrdich/TermChess/internal/engine"
)

func TestResumeGameFunctionality(t *testing.T) {
	// Setup: Create a saved game
	board := engine.NewBoard()
	// Make a move to have a non-starting position
	move, _ := engine.ParseMove("e2e4")
	_ = board.MakeMove(move)

	err := config.SaveGame(board)
	if err != nil {
		t.Fatalf("Failed to save game: %v", err)
	}

	// Test 1: SaveGameExists should return true
	if !config.SaveGameExists() {
		t.Error("config.SaveGameExists() returned false, expected true")
	}

	// Test 2: LoadGame should load the saved position
	loadedBoard, err := config.LoadGame()
	if err != nil {
		t.Fatalf("Failed to load game: %v", err)
	}

	// Verify the loaded board matches the saved board
	if loadedBoard.ToFEN() != board.ToFEN() {
		t.Errorf("Loaded board FEN doesn't match saved board FEN\nExpected: %s\nGot: %s",
			board.ToFEN(), loadedBoard.ToFEN())
	}

	// Test 3: NewModel should start at ScreenMainMenu with Resume Game option when save exists
	testCfg := DefaultConfig()
	model := NewModel(testCfg)
	if model.screen != ScreenMainMenu {
		t.Errorf("NewModel screen = %v, expected ScreenMainMenu (%v)", model.screen, ScreenMainMenu)
	}

	// Verify Resume Game is the first menu option
	if len(model.menuOptions) != 5 {
		t.Errorf("Expected 5 menu options with saved game, got %d", len(model.menuOptions))
	}

	if model.menuOptions[0] != "Resume Game" {
		t.Errorf("Expected first menu option to be 'Resume Game', got '%s'", model.menuOptions[0])
	}

	// Test 4: DeleteSaveGame should remove the saved game
	err = config.DeleteSaveGame()
	if err != nil {
		t.Fatalf("Failed to delete save game: %v", err)
	}

	if config.SaveGameExists() {
		t.Error("config.SaveGameExists() returned true after deletion, expected false")
	}

	// Test 5: NewModel should start at ScreenMainMenu when no save exists
	model = NewModel(testCfg)
	if model.screen != ScreenMainMenu {
		t.Errorf("NewModel screen = %v, expected ScreenMainMenu (%v)", model.screen, ScreenMainMenu)
	}

	// Cleanup
	_ = config.DeleteSaveGame()
}

func TestResumePromptSelection(t *testing.T) {
	// Create a saved game
	board := engine.NewBoard()
	_ = config.SaveGame(board)
	defer config.DeleteSaveGame()

	testCfg := DefaultConfig()
	model := NewModel(testCfg)

	// Test "Yes" selection - should load the game
	model.resumePromptSelection = 0
	// We can't easily test the key handler without running the full Bubbletea program,
	// but we can verify the LoadGame function works
	loadedBoard, err := config.LoadGame()
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
	_ = config.SaveGame(board)

	// Verify it exists
	if !config.SaveGameExists() {
		t.Fatal("config.SaveGameExists() returned false, expected true")
	}

	// Simulate game end by calling DeleteSaveGame (this is what happens in handleGamePlayKeys)
	err := config.DeleteSaveGame()
	if err != nil {
		t.Fatalf("Failed to delete save game: %v", err)
	}

	// Verify it's gone
	if config.SaveGameExists() {
		t.Error("config.SaveGameExists() returned true after deletion, expected false")
	}
}

func TestCorruptedFENHandling(t *testing.T) {
	// Create a corrupted save file
	savePath, _ := config.SaveGamePath()
	configDir := filepath.Dir(savePath)
	_ = os.MkdirAll(configDir, 0755)
	_ = os.WriteFile(savePath, []byte("invalid fen string"), 0644)

	// LoadGame should return an error
	_, err := config.LoadGame()
	if err == nil {
		t.Error("LoadGame should return error for corrupted FEN, got nil")
	}

	// Cleanup
	_ = config.DeleteSaveGame()
}
