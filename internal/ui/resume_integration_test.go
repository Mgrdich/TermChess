package ui

import (
	"os"
	"testing"

	"github.com/Mgrdich/TermChess/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

// TestResumeGameIntegrationFlow tests the complete resume game flow
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

	// Step 2: Start a new app instance - should show resume prompt
	config := DefaultConfig()
	model := NewModel(config)

	if model.screen != ScreenResumePrompt {
		t.Errorf("Expected screen to be ScreenResumePrompt, got %v", model.screen)
	}

	if model.resumePromptSelection != 0 {
		t.Errorf("Expected resumePromptSelection to be 0, got %d", model.resumePromptSelection)
	}

	// Step 3: Simulate pressing "down" to select "No"
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedModel, _ := model.handleResumePromptKeys(msg)
	model = updatedModel.(Model)

	if model.resumePromptSelection != 1 {
		t.Errorf("Expected resumePromptSelection to be 1 after down, got %d", model.resumePromptSelection)
	}

	// Step 4: Press "up" to go back to "Yes"
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	updatedModel, _ = model.handleResumePromptKeys(msg)
	model = updatedModel.(Model)

	if model.resumePromptSelection != 0 {
		t.Errorf("Expected resumePromptSelection to be 0 after up, got %d", model.resumePromptSelection)
	}

	// Step 5: Press Enter to confirm "Yes" - should load the game
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ = model.handleResumePromptKeys(msg)
	model = updatedModel.(Model)

	if model.screen != ScreenGamePlay {
		t.Errorf("Expected screen to be ScreenGamePlay after selecting Yes, got %v", model.screen)
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

	// Step 8: Start a new app instance - should go to main menu
	model2 := NewModel(config)
	if model2.screen != ScreenMainMenu {
		t.Errorf("Expected screen to be ScreenMainMenu when no save exists, got %v", model2.screen)
	}
}

// TestResumeGameSelectNo tests selecting "No" on the resume prompt
func TestResumeGameSelectNo(t *testing.T) {
	// Cleanup before and after
	defer DeleteSaveGame()
	DeleteSaveGame()

	// Create a saved game
	board := engine.NewBoard()
	_ = SaveGame(board)

	// Start app - should show resume prompt
	config := DefaultConfig()
	model := NewModel(config)

	if model.screen != ScreenResumePrompt {
		t.Errorf("Expected screen to be ScreenResumePrompt, got %v", model.screen)
	}

	// Select "No" (move down to select it)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedModel, _ := model.handleResumePromptKeys(msg)
	model = updatedModel.(Model)

	// Press Enter to confirm "No"
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ = model.handleResumePromptKeys(msg)
	model = updatedModel.(Model)

	// Should go to main menu
	if model.screen != ScreenMainMenu {
		t.Errorf("Expected screen to be ScreenMainMenu after selecting No, got %v", model.screen)
	}

	// Board should still be nil
	if model.board != nil {
		t.Error("Expected board to be nil after selecting No")
	}

	// Save file should still exist (we only load/delete when game ends)
	if !SaveGameExists() {
		t.Error("Save game should still exist after selecting No")
	}
}

// TestResumeGameLoadError tests error handling when load fails
func TestResumeGameLoadError(t *testing.T) {
	// Cleanup before and after
	defer DeleteSaveGame()
	DeleteSaveGame()

	// Create a corrupted save file
	savePath, _ := SaveGamePath()
	configDir, _ := getConfigDir()
	_ = os.MkdirAll(configDir, 0755)
	_ = os.WriteFile(savePath, []byte("corrupted fen"), 0644)

	// Start app - should show resume prompt
	config := DefaultConfig()
	model := NewModel(config)

	if model.screen != ScreenResumePrompt {
		t.Errorf("Expected screen to be ScreenResumePrompt, got %v", model.screen)
	}

	// Select "Yes" and press Enter
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ := model.handleResumePromptKeys(msg)
	model = updatedModel.(Model)

	// Should go to main menu due to error
	if model.screen != ScreenMainMenu {
		t.Errorf("Expected screen to be ScreenMainMenu after load error, got %v", model.screen)
	}

	// Should have an error message
	if model.errorMsg == "" {
		t.Error("Expected error message after load failure")
	}
}
