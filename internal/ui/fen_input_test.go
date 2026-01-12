package ui

import (
	"testing"

	"github.com/Mgrdich/TermChess/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

// TestFENInputNavigation tests that "Load Game" from main menu transitions to FEN input screen.
func TestFENInputNavigation(t *testing.T) {
	config := DefaultConfig()
	m := NewModel(config)

	// Set screen to main menu (if not already there due to saved game)
	m.screen = ScreenMainMenu
	m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}

	// Select "Load Game" option (index 1)
	m.menuSelection = 1

	// Simulate Enter key press
	updatedModelInterface, _ := m.handleMainMenuSelection()
	updatedModel := updatedModelInterface.(Model)

	// Verify we're now on FEN input screen
	if updatedModel.screen != ScreenFENInput {
		t.Errorf("Expected screen to be ScreenFENInput, got %v", updatedModel.screen)
	}

	// Verify text input is ready
	if updatedModel.fenInput.Value() != "" {
		t.Errorf("Expected fenInput to be empty, got %q", updatedModel.fenInput.Value())
	}
}

// TestFENInputValidFEN tests loading a valid FEN string.
func TestFENInputValidFEN(t *testing.T) {
	config := DefaultConfig()
	m := NewModel(config)

	// Set screen to FEN input
	m.screen = ScreenFENInput
	m.fenInput.Focus()

	// Set a valid FEN string (starting position)
	validFEN := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	m.fenInput.SetValue(validFEN)

	// Simulate Enter key press
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModelInterface, _ := m.handleFENInputKeys(msg)
	updatedModel := updatedModelInterface.(Model)

	// Verify we transitioned to gameplay screen
	if updatedModel.screen != ScreenGamePlay {
		t.Errorf("Expected screen to be ScreenGamePlay, got %v", updatedModel.screen)
	}

	// Verify the board was loaded
	if updatedModel.board == nil {
		t.Fatal("Expected board to be loaded, got nil")
	}

	// Verify the board has the correct starting position
	expectedBoard := engine.NewBoard()
	if updatedModel.board.ToFEN() != expectedBoard.ToFEN() {
		t.Errorf("Expected starting position, got %s", updatedModel.board.ToFEN())
	}

	// Verify game type is set to PvP
	if updatedModel.gameType != GameTypePvP {
		t.Errorf("Expected game type to be PvP, got %v", updatedModel.gameType)
	}

	// Verify no error message
	if updatedModel.errorMsg != "" {
		t.Errorf("Expected no error message, got %q", updatedModel.errorMsg)
	}
}

// TestFENInputInvalidFEN tests loading an invalid FEN string.
func TestFENInputInvalidFEN(t *testing.T) {
	config := DefaultConfig()
	m := NewModel(config)

	// Set screen to FEN input
	m.screen = ScreenFENInput
	m.fenInput.Focus()

	// Set an invalid FEN string (missing fields)
	invalidFEN := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR"
	m.fenInput.SetValue(invalidFEN)

	// Simulate Enter key press
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModelInterface, _ := m.handleFENInputKeys(msg)
	updatedModel := updatedModelInterface.(Model)

	// Verify we're still on FEN input screen
	if updatedModel.screen != ScreenFENInput {
		t.Errorf("Expected to stay on ScreenFENInput, got %v", updatedModel.screen)
	}

	// Verify an error message is displayed
	if updatedModel.errorMsg == "" {
		t.Error("Expected error message for invalid FEN")
	}

	// Verify the board was not loaded
	if updatedModel.board != nil {
		t.Error("Expected board to remain nil for invalid FEN")
	}
}

// TestFENInputEmptyString tests submitting an empty FEN string.
func TestFENInputEmptyString(t *testing.T) {
	config := DefaultConfig()
	m := NewModel(config)

	// Set screen to FEN input
	m.screen = ScreenFENInput
	m.fenInput.Focus()
	m.fenInput.SetValue("")

	// Simulate Enter key press
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModelInterface, _ := m.handleFENInputKeys(msg)
	updatedModel := updatedModelInterface.(Model)

	// Verify we're still on FEN input screen
	if updatedModel.screen != ScreenFENInput {
		t.Errorf("Expected to stay on ScreenFENInput, got %v", updatedModel.screen)
	}

	// Verify an error message is displayed
	if updatedModel.errorMsg == "" {
		t.Error("Expected error message for empty FEN string")
	}
}

// TestFENInputEscapeToMenu tests pressing Esc to return to main menu.
func TestFENInputEscapeToMenu(t *testing.T) {
	config := DefaultConfig()
	m := NewModel(config)

	// Set screen to FEN input
	m.screen = ScreenFENInput
	m.fenInput.Focus()
	m.fenInput.SetValue("some test input")

	// Simulate Esc key press
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModelInterface, _ := m.handleFENInputKeys(msg)
	updatedModel := updatedModelInterface.(Model)

	// Verify we're back on main menu
	if updatedModel.screen != ScreenMainMenu {
		t.Errorf("Expected screen to be ScreenMainMenu, got %v", updatedModel.screen)
	}

	// Verify menu options are restored
	expectedOptions := []string{"New Game", "Load Game", "Settings", "Exit"}
	if len(updatedModel.menuOptions) != len(expectedOptions) {
		t.Errorf("Expected %d menu options, got %d", len(expectedOptions), len(updatedModel.menuOptions))
	}

	// Verify fenInput is cleared
	if updatedModel.fenInput.Value() != "" {
		t.Errorf("Expected fenInput to be cleared, got %q", updatedModel.fenInput.Value())
	}
}

// TestFENInputMidGamePosition tests loading a mid-game position.
func TestFENInputMidGamePosition(t *testing.T) {
	config := DefaultConfig()
	m := NewModel(config)

	// Set screen to FEN input
	m.screen = ScreenFENInput
	m.fenInput.Focus()

	// Set a mid-game FEN string
	midGameFEN := "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"
	m.fenInput.SetValue(midGameFEN)

	// Simulate Enter key press
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModelInterface, _ := m.handleFENInputKeys(msg)
	updatedModel := updatedModelInterface.(Model)

	// Verify we transitioned to gameplay screen
	if updatedModel.screen != ScreenGamePlay {
		t.Errorf("Expected screen to be ScreenGamePlay, got %v", updatedModel.screen)
	}

	// Verify the board was loaded
	if updatedModel.board == nil {
		t.Fatal("Expected board to be loaded, got nil")
	}

	// Verify the board matches the mid-game position
	if updatedModel.board.ToFEN() != midGameFEN {
		t.Errorf("Expected FEN %s, got %s", midGameFEN, updatedModel.board.ToFEN())
	}

	// Verify it's White's turn
	if updatedModel.board.ActiveColor != engine.White {
		t.Errorf("Expected White to move, got %v", updatedModel.board.ActiveColor)
	}

	// Verify castling rights are correct
	if updatedModel.board.CastlingRights != (engine.CastleWhiteKing | engine.CastleWhiteQueen | engine.CastleBlackKing | engine.CastleBlackQueen) {
		t.Errorf("Expected all castling rights, got %v", updatedModel.board.CastlingRights)
	}
}
