package ui

import (
	"testing"

	"github.com/Mgrdich/TermChess/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleGamePlayKeys_ValidMove(t *testing.T) {
	// Create a model with a new board
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// Simulate typing "e2e4"
	keys := []rune{'e', '2', 'e', '4'}
	for _, key := range keys {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}}
		result, _ := m.handleGamePlayKeys(msg)
		m = result.(Model)
	}

	// Verify input was captured
	if m.input != "e2e4" {
		t.Errorf("Expected input 'e2e4', got '%s'", m.input)
	}

	// Simulate pressing Enter
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleGamePlayKeys(msg)
	m = result.(Model)

	// Verify move was executed
	if m.input != "" {
		t.Errorf("Expected input to be cleared after valid move, got '%s'", m.input)
	}

	if m.errorMsg != "" {
		t.Errorf("Expected no error message after valid move, got '%s'", m.errorMsg)
	}

	// Verify the board was updated - pawn should have moved from e2 to e4
	e2 := engine.NewSquare(4, 1) // e2
	e4 := engine.NewSquare(4, 3) // e4

	if !m.board.PieceAt(e2).IsEmpty() {
		t.Error("Expected e2 to be empty after move")
	}

	pieceAtE4 := m.board.PieceAt(e4)
	if pieceAtE4.Type() != engine.Pawn || pieceAtE4.Color() != engine.White {
		t.Error("Expected white pawn at e4 after move")
	}

	// Verify turn changed to Black
	if m.board.ActiveColor != engine.Black {
		t.Error("Expected active color to be Black after White's move")
	}

	// Verify move was added to history
	if len(m.moveHistory) != 1 {
		t.Errorf("Expected 1 move in history, got %d", len(m.moveHistory))
	}
}

func TestHandleGamePlayKeys_InvalidMoveFormat(t *testing.T) {
	// Create a model with a new board
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// Simulate typing an invalid move format
	invalidInputs := []string{"invalid", "e2", "e2e", "e2e99", "z2z4"}

	for _, input := range invalidInputs {
		m.input = input
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		result, _ := m.handleGamePlayKeys(msg)
		m = result.(Model)

		// Verify error message was set
		if m.errorMsg == "" {
			t.Errorf("Expected error message for invalid input '%s', got none", input)
		}

		// Verify board was not modified (still White to move)
		if m.board.ActiveColor != engine.White {
			t.Errorf("Expected active color to still be White after invalid move '%s'", input)
		}
	}
}

func TestHandleGamePlayKeys_IllegalMove(t *testing.T) {
	// Create a model with a new board
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// Try an illegal move: e2e5 (pawn can't move 3 squares)
	m.input = "e2e5"
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleGamePlayKeys(msg)
	m = result.(Model)

	// Verify error message was set
	if m.errorMsg == "" {
		t.Error("Expected error message for illegal move e2e5")
	}

	// Verify board was not modified
	e2 := engine.NewSquare(4, 1) // e2
	e5 := engine.NewSquare(4, 4) // e5

	pieceAtE2 := m.board.PieceAt(e2)
	if pieceAtE2.Type() != engine.Pawn || pieceAtE2.Color() != engine.White {
		t.Error("Expected white pawn to still be at e2 after illegal move")
	}

	if !m.board.PieceAt(e5).IsEmpty() {
		t.Error("Expected e5 to be empty after illegal move")
	}

	// Verify still White to move
	if m.board.ActiveColor != engine.White {
		t.Error("Expected active color to still be White after illegal move")
	}
}

func TestHandleGamePlayKeys_ErrorClearingOnNewInput(t *testing.T) {
	// Create a model with a new board and an error message
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay
	m.errorMsg = "Previous error"

	// Simulate typing a character
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	result, _ := m.handleGamePlayKeys(msg)
	m = result.(Model)

	// Verify error message was cleared
	if m.errorMsg != "" {
		t.Errorf("Expected error message to be cleared when typing, got '%s'", m.errorMsg)
	}

	// Verify input was updated
	if m.input != "e" {
		t.Errorf("Expected input 'e', got '%s'", m.input)
	}
}

func TestHandleGamePlayKeys_BackspaceHandling(t *testing.T) {
	// Create a model with some input
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay
	m.input = "e2e4"
	m.errorMsg = "Some error"

	// Simulate pressing Backspace
	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	result, _ := m.handleGamePlayKeys(msg)
	m = result.(Model)

	// Verify last character was removed
	if m.input != "e2e" {
		t.Errorf("Expected input 'e2e' after backspace, got '%s'", m.input)
	}

	// Verify error message was cleared
	if m.errorMsg != "" {
		t.Errorf("Expected error message to be cleared on backspace, got '%s'", m.errorMsg)
	}

	// Backspace on empty input should not cause issues
	m.input = ""
	msg = tea.KeyMsg{Type: tea.KeyBackspace}
	result, _ = m.handleGamePlayKeys(msg)
	m = result.(Model)

	if m.input != "" {
		t.Errorf("Expected input to remain empty after backspace on empty string, got '%s'", m.input)
	}
}

func TestHandleGamePlayKeys_SequenceOfMoves(t *testing.T) {
	// Create a model with a new board
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// Execute a sequence of valid moves
	moves := []string{"e2e4", "e7e5", "g1f3", "b8c6"}

	for i, moveStr := range moves {
		m.input = moveStr
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		result, _ := m.handleGamePlayKeys(msg)
		m = result.(Model)

		// Verify no error
		if m.errorMsg != "" {
			t.Errorf("Move %d (%s) failed with error: %s", i+1, moveStr, m.errorMsg)
		}

		// Verify input was cleared
		if m.input != "" {
			t.Errorf("Expected input to be cleared after move %d, got '%s'", i+1, m.input)
		}
	}

	// Verify move history has all moves
	if len(m.moveHistory) != 4 {
		t.Errorf("Expected 4 moves in history, got %d", len(m.moveHistory))
	}

	// Verify it's White's turn again
	if m.board.ActiveColor != engine.White {
		t.Error("Expected active color to be White after 4 moves")
	}
}

func TestHandleGamePlayKeys_EmptyInputEnter(t *testing.T) {
	// Create a model with a new board
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay
	m.input = ""

	// Simulate pressing Enter with empty input
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleGamePlayKeys(msg)
	m = result.(Model)

	// Verify nothing happened (no error, board unchanged)
	if m.errorMsg != "" {
		t.Errorf("Expected no error for empty input, got '%s'", m.errorMsg)
	}

	if m.board.ActiveColor != engine.White {
		t.Error("Expected active color to still be White")
	}

	if len(m.moveHistory) != 0 {
		t.Errorf("Expected empty move history, got %d moves", len(m.moveHistory))
	}
}

func TestHandleGamePlayKeys_PromotionMove(t *testing.T) {
	// Create a custom board position where a pawn can promote
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// Set up a position: White pawn on e7, ready to promote
	e7 := engine.NewSquare(4, 6) // e7
	e8 := engine.NewSquare(4, 7) // e8

	// Clear e2 pawn and place it on e7
	m.board.Squares[engine.NewSquare(4, 1)] = engine.Piece(engine.Empty)
	m.board.Squares[e7] = engine.NewPiece(engine.White, engine.Pawn)
	// Clear black pieces on rank 8 that might block
	m.board.Squares[e8] = engine.Piece(engine.Empty)

	// Try to move without promotion piece (should fail)
	m.input = "e7e8"
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleGamePlayKeys(msg)
	m = result.(Model)

	if m.errorMsg == "" {
		t.Error("Expected error for promotion move without promotion piece")
	}

	// Try with promotion piece (should succeed)
	m.input = "e7e8q"
	m.errorMsg = "" // Clear previous error
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	result, _ = m.handleGamePlayKeys(msg)
	m = result.(Model)

	if m.errorMsg != "" {
		t.Errorf("Expected no error for promotion move with piece, got '%s'", m.errorMsg)
	}

	// Verify queen was placed at e8
	pieceAtE8 := m.board.PieceAt(e8)
	if pieceAtE8.Type() != engine.Queen || pieceAtE8.Color() != engine.White {
		t.Error("Expected white queen at e8 after promotion")
	}
}

func TestHandleMainMenuSelection_Settings(t *testing.T) {
	// Create a model at the main menu
	m := NewModel(DefaultConfig())
	m.screen = ScreenMainMenu
	m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
	m.menuSelection = 2 // Select "Settings"

	// Simulate pressing Enter
	result, _ := m.handleMainMenuSelection()
	m = result.(Model)

	// Verify transitioned to Settings screen
	if m.screen != ScreenSettings {
		t.Errorf("Expected screen to be ScreenSettings, got %v", m.screen)
	}

	// Verify settings selection is initialized to 0
	if m.settingsSelection != 0 {
		t.Errorf("Expected settingsSelection to be 0, got %d", m.settingsSelection)
	}

	// Verify messages are cleared
	if m.errorMsg != "" {
		t.Errorf("Expected errorMsg to be cleared, got '%s'", m.errorMsg)
	}

	if m.statusMsg != "" {
		t.Errorf("Expected statusMsg to be cleared, got '%s'", m.statusMsg)
	}
}

func TestHandleSettingsKeys_Navigation(t *testing.T) {
	// Create a model at the settings screen
	m := NewModel(DefaultConfig())
	m.screen = ScreenSettings
	m.settingsSelection = 0

	// Test moving down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	result, _ := m.handleSettingsKeys(msg)
	m = result.(Model)

	if m.settingsSelection != 1 {
		t.Errorf("Expected settingsSelection to be 1 after down, got %d", m.settingsSelection)
	}

	// Test moving down with arrow key
	msg = tea.KeyMsg{Type: tea.KeyDown}
	result, _ = m.handleSettingsKeys(msg)
	m = result.(Model)

	if m.settingsSelection != 2 {
		t.Errorf("Expected settingsSelection to be 2 after down, got %d", m.settingsSelection)
	}

	// Test moving up
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	result, _ = m.handleSettingsKeys(msg)
	m = result.(Model)

	if m.settingsSelection != 1 {
		t.Errorf("Expected settingsSelection to be 1 after up, got %d", m.settingsSelection)
	}

	// Test wrapping at bottom (move to index 3, then down should wrap to 0)
	m.settingsSelection = 3
	msg = tea.KeyMsg{Type: tea.KeyDown}
	result, _ = m.handleSettingsKeys(msg)
	m = result.(Model)

	if m.settingsSelection != 0 {
		t.Errorf("Expected settingsSelection to wrap to 0, got %d", m.settingsSelection)
	}

	// Test wrapping at top (at index 0, up should wrap to 3)
	msg = tea.KeyMsg{Type: tea.KeyUp}
	result, _ = m.handleSettingsKeys(msg)
	m = result.(Model)

	if m.settingsSelection != 3 {
		t.Errorf("Expected settingsSelection to wrap to 3, got %d", m.settingsSelection)
	}
}

func TestHandleSettingsKeys_Toggle(t *testing.T) {
	// Create a model at the settings screen
	m := NewModel(DefaultConfig())
	m.screen = ScreenSettings
	m.settingsSelection = 0 // Use Unicode Pieces

	// Store initial value
	initialValue := m.config.UseUnicode

	// Toggle with Space
	msg := tea.KeyMsg{Type: tea.KeySpace}
	result, _ := m.handleSettingsKeys(msg)
	m = result.(Model)

	// Verify value was toggled
	if m.config.UseUnicode == initialValue {
		t.Error("Expected UseUnicode to be toggled")
	}

	// Toggle with Enter
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	result, _ = m.handleSettingsKeys(msg)
	m = result.(Model)

	// Verify value was toggled back
	if m.config.UseUnicode != initialValue {
		t.Error("Expected UseUnicode to be toggled back to initial value")
	}
}

func TestHandleSettingsKeys_ToggleAllSettings(t *testing.T) {
	// Create a model at the settings screen
	m := NewModel(DefaultConfig())
	m.screen = ScreenSettings

	// Test toggling each setting
	settings := []struct {
		index    int
		getName  func() string
		getValue func() bool
	}{
		{0, func() string { return "UseUnicode" }, func() bool { return m.config.UseUnicode }},
		{1, func() string { return "ShowCoords" }, func() bool { return m.config.ShowCoords }},
		{2, func() string { return "UseColors" }, func() bool { return m.config.UseColors }},
		{3, func() string { return "ShowMoveHistory" }, func() bool { return m.config.ShowMoveHistory }},
	}

	for _, setting := range settings {
		m.settingsSelection = setting.index
		initialValue := setting.getValue()

		// Toggle the setting
		msg := tea.KeyMsg{Type: tea.KeySpace}
		result, _ := m.handleSettingsKeys(msg)
		m = result.(Model)

		// Verify value was toggled
		if setting.getValue() == initialValue {
			t.Errorf("Expected %s to be toggled from %v", setting.getName(), initialValue)
		}
	}
}

func TestHandleSettingsKeys_ESC(t *testing.T) {
	// Create a model at the settings screen
	m := NewModel(DefaultConfig())
	m.screen = ScreenSettings
	m.settingsSelection = 2

	// Press ESC
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := m.handleSettingsKeys(msg)
	m = result.(Model)

	// Verify returned to main menu
	if m.screen != ScreenMainMenu {
		t.Errorf("Expected screen to be ScreenMainMenu, got %v", m.screen)
	}

	// Verify menu was reset
	if m.menuSelection != 0 {
		t.Errorf("Expected menuSelection to be reset to 0, got %d", m.menuSelection)
	}

	expectedOptions := []string{"New Game", "Load Game", "Settings", "Exit"}
	if len(m.menuOptions) != len(expectedOptions) {
		t.Errorf("Expected %d menu options, got %d", len(expectedOptions), len(m.menuOptions))
	}

	// Verify messages are cleared
	if m.errorMsg != "" {
		t.Errorf("Expected errorMsg to be cleared, got '%s'", m.errorMsg)
	}

	if m.statusMsg != "" {
		t.Errorf("Expected statusMsg to be cleared, got '%s'", m.statusMsg)
	}
}
