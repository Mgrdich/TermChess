package ui

import (
	"os"
	"testing"

	"github.com/Mgrdich/TermChess/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

// TestResumeGameIntegrationFlow tests the complete resume game flow with new menu-based approach
func TestResumeGameIntegrationFlow(t *testing.T) {
	// Cleanup before and after
	defer DeleteSaveGame()
	DeleteSaveGame()

	// Step 1: Create and save a game in progress
	board := engine.NewBoard()
	move, _ := engine.ParseMove("e2e4")
	_ = board.MakeMove(move)
	savedFEN := board.ToFEN()

	err := SaveGame(board)
	if err != nil {
		t.Fatalf("Failed to save game: %v", err)
	}

	// Step 2: Start a new app instance - should show main menu with Resume Game option
	config := DefaultConfig()
	model := NewModel(config)

	if model.screen != ScreenMainMenu {
		t.Errorf("Expected screen to be ScreenMainMenu, got %v", model.screen)
	}

	// Verify Resume Game is the first option
	if len(model.menuOptions) != 5 {
		t.Errorf("Expected 5 menu options with saved game, got %d", len(model.menuOptions))
	}

	if model.menuOptions[0] != "Resume Game" {
		t.Errorf("Expected first menu option to be 'Resume Game', got '%s'", model.menuOptions[0])
	}

	if model.menuSelection != 0 {
		t.Errorf("Expected menuSelection to be 0, got %d", model.menuSelection)
	}

	// Step 3: Simulate pressing "down" to navigate away from Resume Game
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedModel, _ := model.handleMainMenuKeys(msg)
	model = updatedModel.(Model)

	if model.menuSelection != 1 {
		t.Errorf("Expected menuSelection to be 1 after down, got %d", model.menuSelection)
	}

	// Step 4: Press "up" to go back to Resume Game
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	updatedModel, _ = model.handleMainMenuKeys(msg)
	model = updatedModel.(Model)

	if model.menuSelection != 0 {
		t.Errorf("Expected menuSelection to be 0 after up, got %d", model.menuSelection)
	}

	// Step 5: Press Enter to select Resume Game - should load the game
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ = model.handleMainMenuKeys(msg)
	model = updatedModel.(Model)

	if model.screen != ScreenGamePlay {
		t.Errorf("Expected screen to be ScreenGamePlay after selecting Resume Game, got %v", model.screen)
	}

	if model.board == nil {
		t.Fatal("Expected board to be loaded, got nil")
	}

	if model.board.ToFEN() != savedFEN {
		t.Errorf("Loaded board FEN doesn't match saved FEN\nExpected: %s\nGot: %s",
			savedFEN, model.board.ToFEN())
	}

	// Step 6: Simulate completing the game (game ends)
	// Set up a checkmate position
	checkmateBoard, _ := engine.FromFEN("r1bqkb1r/pppp1Qpp/2n2n2/4p3/2B1P3/8/PPPP1PPP/RNB1K1NR b KQkq - 0 4")
	model.board = checkmateBoard

	// Verify game is over
	if !model.board.IsGameOver() {
		t.Error("Expected game to be over")
	}

	// Step 7: When game ends, save file should be deleted
	// Simulate the handleGamePlayKeys behavior when game ends
	_ = DeleteSaveGame()

	if SaveGameExists() {
		t.Error("Save game should not exist after game ends")
	}

	// Step 8: Start a new app instance - should go to main menu without Resume Game option
	model2 := NewModel(config)
	if model2.screen != ScreenMainMenu {
		t.Errorf("Expected screen to be ScreenMainMenu when no save exists, got %v", model2.screen)
	}

	// Verify no Resume Game option
	if len(model2.menuOptions) != 4 {
		t.Errorf("Expected 4 menu options without saved game, got %d", len(model2.menuOptions))
	}

	for _, opt := range model2.menuOptions {
		if opt == "Resume Game" {
			t.Error("Did not expect 'Resume Game' option when no saved game exists")
		}
	}
}

// TestResumeGameSelectNo tests navigating to a different menu option instead of Resume Game
func TestResumeGameSelectNo(t *testing.T) {
	// Cleanup before and after
	defer DeleteSaveGame()
	DeleteSaveGame()

	// Create a saved game
	board := engine.NewBoard()
	_ = SaveGame(board)

	// Start app - should show main menu with Resume Game option
	config := DefaultConfig()
	model := NewModel(config)

	if model.screen != ScreenMainMenu {
		t.Errorf("Expected screen to be ScreenMainMenu, got %v", model.screen)
	}

	// Verify Resume Game is present
	if model.menuOptions[0] != "Resume Game" {
		t.Errorf("Expected first option to be 'Resume Game', got '%s'", model.menuOptions[0])
	}

	// Navigate down to "New Game" (index 1)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedModel, _ := model.handleMainMenuKeys(msg)
	model = updatedModel.(Model)

	if model.menuSelection != 1 {
		t.Errorf("Expected menuSelection to be 1, got %d", model.menuSelection)
	}

	// Press Enter to select "New Game"
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ = model.handleMainMenuKeys(msg)
	model = updatedModel.(Model)

	// Should go to game type selection screen
	if model.screen != ScreenGameTypeSelect {
		t.Errorf("Expected screen to be ScreenGameTypeSelect after selecting New Game, got %v", model.screen)
	}

	// Board should still be nil
	if model.board != nil {
		t.Error("Expected board to be nil after selecting New Game")
	}

	// Save file should still exist (we only delete when game ends)
	if !SaveGameExists() {
		t.Error("Save game should still exist after selecting New Game")
	}
}

// TestResumeGameLoadError tests error handling when resuming a corrupted saved game
func TestResumeGameLoadError(t *testing.T) {
	// Cleanup before and after
	defer DeleteSaveGame()
	DeleteSaveGame()

	// Create a corrupted save file
	savePath, _ := SaveGamePath()
	configDir, _ := getConfigDir()
	_ = os.MkdirAll(configDir, 0755)
	_ = os.WriteFile(savePath, []byte("corrupted fen"), 0644)

	// Start app - should show main menu with Resume Game option
	config := DefaultConfig()
	model := NewModel(config)

	if model.screen != ScreenMainMenu {
		t.Errorf("Expected screen to be ScreenMainMenu, got %v", model.screen)
	}

	// Verify Resume Game is present
	if model.menuOptions[0] != "Resume Game" {
		t.Errorf("Expected first option to be 'Resume Game', got '%s'", model.menuOptions[0])
	}

	// Select Resume Game and press Enter
	model.menuSelection = 0
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ := model.handleMainMenuKeys(msg)
	model = updatedModel.(Model)

	// Should stay on main menu due to error
	if model.screen != ScreenMainMenu {
		t.Errorf("Expected screen to be ScreenMainMenu after load error, got %v", model.screen)
	}

	// Should have an error message
	if model.errorMsg == "" {
		t.Error("Expected error message after load failure")
	}

	// Board should still be nil
	if model.board != nil {
		t.Error("Expected board to be nil after failed load")
	}
}
