package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestFENInputScreen_Initialize verifies that the FEN input screen initializes correctly.
func TestFENInputScreen_Initialize(t *testing.T) {
	m := NewModel()
	m.screen = ScreenFENInput
	m.fenInput = ""

	// Verify screen renders without errors
	view := m.View()
	if !strings.Contains(view, "Load Game from FEN") {
		t.Error("FEN input screen should contain header")
	}

	if !strings.Contains(view, "Enter a FEN string to load a chess position") {
		t.Error("FEN input screen should contain instructions")
	}

	if !strings.Contains(view, "Example:") {
		t.Error("FEN input screen should contain example FEN")
	}
}

// TestFENInput_ValidFEN verifies that a valid FEN string loads correctly.
func TestFENInput_ValidFEN(t *testing.T) {
	m := NewModel()
	m.screen = ScreenFENInput
	// Starting position FEN
	m.fenInput = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

	// Simulate Enter key
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.Update(msg)
	m = result.(Model)

	// Should transition to GamePlay
	if m.screen != ScreenGamePlay {
		t.Errorf("Expected ScreenGamePlay, got %v", m.screen)
	}

	// Should have loaded board
	if m.board == nil {
		t.Error("Board should be loaded")
	}

	// Should have no error
	if m.errorMsg != "" {
		t.Errorf("Should have no error, got: %s", m.errorMsg)
	}

	// FEN input should be cleared
	if m.fenInput != "" {
		t.Error("FEN input should be cleared after successful load")
	}

	// Should have status message
	if m.statusMsg == "" {
		t.Error("Should have status message after loading FEN")
	}
}

// TestFENInput_ValidFENWithMovePlayed verifies loading a FEN with a move already played.
func TestFENInput_ValidFENWithMovePlayed(t *testing.T) {
	m := NewModel()
	m.screen = ScreenFENInput
	// Position after 1.e4
	m.fenInput = "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"

	// Simulate Enter key
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.Update(msg)
	m = result.(Model)

	// Should transition to GamePlay
	if m.screen != ScreenGamePlay {
		t.Errorf("Expected ScreenGamePlay, got %v", m.screen)
	}

	// Should have loaded board
	if m.board == nil {
		t.Fatal("Board should be loaded")
	}

	// Should have no error
	if m.errorMsg != "" {
		t.Errorf("Should have no error, got: %s", m.errorMsg)
	}

	// Verify it's Black's turn (since FEN has 'b')
	if m.board.ActiveColor != 1 { // Black = 1
		t.Error("Expected Black to move after loading this FEN")
	}
}

// TestFENInput_InvalidFEN verifies that an invalid FEN string shows an error.
func TestFENInput_InvalidFEN(t *testing.T) {
	m := NewModel()
	m.screen = ScreenFENInput
	m.fenInput = "invalid fen string"

	// Simulate Enter key
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.Update(msg)
	m = result.(Model)

	// Should stay on FEN input screen
	if m.screen != ScreenFENInput {
		t.Errorf("Expected ScreenFENInput, got %v", m.screen)
	}

	// Should have error message
	if m.errorMsg == "" {
		t.Error("Should have error message for invalid FEN")
	}

	if !strings.Contains(m.errorMsg, "Invalid FEN") {
		t.Errorf("Error should mention invalid FEN, got: %s", m.errorMsg)
	}

	// Should still have the input (not cleared)
	if m.fenInput != "invalid fen string" {
		t.Error("FEN input should not be cleared when there's an error")
	}
}

// TestFENInput_EmptyFEN verifies that submitting an empty FEN shows an error.
func TestFENInput_EmptyFEN(t *testing.T) {
	m := NewModel()
	m.screen = ScreenFENInput
	m.fenInput = ""

	// Simulate Enter key
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.Update(msg)
	m = result.(Model)

	// Should stay on FEN input screen
	if m.screen != ScreenFENInput {
		t.Error("Should stay on FEN input screen")
	}

	// Should have error message
	if !strings.Contains(m.errorMsg, "Please enter a FEN string") {
		t.Errorf("Should prompt user to enter FEN, got: %s", m.errorMsg)
	}
}

// TestFENInput_ESCToMainMenu verifies that ESC returns to the main menu.
func TestFENInput_ESCToMainMenu(t *testing.T) {
	m := NewModel()
	m.screen = ScreenFENInput
	m.fenInput = "some input"
	m.errorMsg = "some error"

	// Simulate ESC key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := m.Update(msg)
	m = result.(Model)

	// Should return to main menu
	if m.screen != ScreenMainMenu {
		t.Errorf("Expected ScreenMainMenu, got %v", m.screen)
	}

	// Input should be cleared
	if m.fenInput != "" {
		t.Error("FEN input should be cleared")
	}

	// Error should be cleared
	if m.errorMsg != "" {
		t.Error("Error message should be cleared")
	}

	// Menu options should be reset
	expectedOptions := []string{"New Game", "Load Game", "Settings", "Exit"}
	if len(m.menuOptions) != len(expectedOptions) {
		t.Errorf("Expected %d menu options, got %d", len(expectedOptions), len(m.menuOptions))
	}
}

// TestFENInput_Backspace verifies that backspace removes characters.
func TestFENInput_Backspace(t *testing.T) {
	m := NewModel()
	m.screen = ScreenFENInput
	m.fenInput = "test"

	// Simulate backspace
	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	result, _ := m.Update(msg)
	m = result.(Model)

	if m.fenInput != "tes" {
		t.Errorf("Expected 'tes', got '%s'", m.fenInput)
	}

	// Backspace again
	msg = tea.KeyMsg{Type: tea.KeyBackspace}
	result, _ = m.Update(msg)
	m = result.(Model)

	if m.fenInput != "te" {
		t.Errorf("Expected 'te', got '%s'", m.fenInput)
	}
}

// TestFENInput_BackspaceOnEmpty verifies that backspace on empty input doesn't cause issues.
func TestFENInput_BackspaceOnEmpty(t *testing.T) {
	m := NewModel()
	m.screen = ScreenFENInput
	m.fenInput = ""

	// Simulate backspace on empty input
	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	result, _ := m.Update(msg)
	m = result.(Model)

	// Should remain empty
	if m.fenInput != "" {
		t.Errorf("Expected empty input, got '%s'", m.fenInput)
	}
}

// TestFENInput_CharacterInput verifies that typing adds characters to the input.
func TestFENInput_CharacterInput(t *testing.T) {
	m := NewModel()
	m.screen = ScreenFENInput
	m.fenInput = ""

	// Simulate typing characters
	chars := []string{"r", "n", "b", "q"}
	for _, ch := range chars {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(ch)}
		result, _ := m.Update(msg)
		m = result.(Model)
	}

	if m.fenInput != "rnbq" {
		t.Errorf("Expected 'rnbq', got '%s'", m.fenInput)
	}
}

// TestFENInput_SpecialCharacters verifies that special characters in FEN are accepted.
func TestFENInput_SpecialCharacters(t *testing.T) {
	m := NewModel()
	m.screen = ScreenFENInput
	m.fenInput = ""

	// Simulate typing FEN-relevant characters
	chars := []string{"r", "n", "8", "/", " ", "w", " ", "K", "Q", "k", "q", " ", "-", " ", "0", " ", "1"}
	for _, ch := range chars {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(ch)}
		result, _ := m.Update(msg)
		m = result.(Model)
	}

	expected := "rn8/ w KQkq - 0 1"
	if m.fenInput != expected {
		t.Errorf("Expected '%s', got '%s'", expected, m.fenInput)
	}
}

// TestFENInput_CtrlU_ClearInput verifies that Ctrl+U clears the entire input.
func TestFENInput_CtrlU_ClearInput(t *testing.T) {
	m := NewModel()
	m.screen = ScreenFENInput
	m.fenInput = "some long input text"

	// Simulate Ctrl+U
	msg := tea.KeyMsg{Type: tea.KeyCtrlU}
	result, _ := m.Update(msg)
	m = result.(Model)

	if m.fenInput != "" {
		t.Errorf("Expected empty input after Ctrl+U, got '%s'", m.fenInput)
	}
}

// TestFENInput_ViewWithInput verifies that the view displays the current input.
func TestFENInput_ViewWithInput(t *testing.T) {
	m := NewModel()
	m.screen = ScreenFENInput
	m.fenInput = "rnbqkbnr"

	view := m.View()

	// Should contain the input text
	if !strings.Contains(view, "rnbqkbnr") {
		t.Error("View should display the current input")
	}

	// Should contain cursor
	if !strings.Contains(view, "█") {
		t.Error("View should display cursor")
	}
}

// TestFENInput_ViewWithError verifies that the view displays error messages.
func TestFENInput_ViewWithError(t *testing.T) {
	m := NewModel()
	m.screen = ScreenFENInput
	m.errorMsg = "Test error message"

	view := m.View()

	// Should contain the error message
	if !strings.Contains(view, "Test error message") {
		t.Error("View should display error message")
	}

	if !strings.Contains(view, "Error:") {
		t.Error("View should display error label")
	}
}

// TestFENInput_MainMenuTransition verifies transition from main menu to FEN input.
func TestFENInput_MainMenuTransition(t *testing.T) {
	m := NewModel()
	m.screen = ScreenMainMenu
	m.menuSelection = 1 // "Load Game" is at index 1

	// Simulate pressing Enter on "Load Game"
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.Update(msg)
	m = result.(Model)

	// Should transition to FEN input screen
	if m.screen != ScreenFENInput {
		t.Errorf("Expected ScreenFENInput, got %v", m.screen)
	}

	// FEN input should be empty
	if m.fenInput != "" {
		t.Error("FEN input should be empty on transition")
	}

	// Error and status messages should be cleared
	if m.errorMsg != "" {
		t.Error("Error message should be cleared on transition")
	}
	if m.statusMsg != "" {
		t.Error("Status message should be cleared on transition")
	}
}

// TestFENInput_ComplexFEN verifies loading a complex mid-game position.
func TestFENInput_ComplexFEN(t *testing.T) {
	m := NewModel()
	m.screen = ScreenFENInput
	// A complex mid-game position
	m.fenInput = "r1bqk2r/pppp1ppp/2n2n2/2b1p3/2B1P3/3P1N2/PPP2PPP/RNBQK2R w KQkq - 4 5"

	// Simulate Enter key
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.Update(msg)
	m = result.(Model)

	// Should transition to GamePlay
	if m.screen != ScreenGamePlay {
		t.Errorf("Expected ScreenGamePlay, got %v", m.screen)
	}

	// Should have loaded board
	if m.board == nil {
		t.Fatal("Board should be loaded")
	}

	// Should have no error
	if m.errorMsg != "" {
		t.Errorf("Should have no error for complex FEN, got: %s", m.errorMsg)
	}

	// Verify full move number
	if m.board.FullMoveNum != 5 {
		t.Errorf("Expected full move number 5, got %d", m.board.FullMoveNum)
	}
}

// TestFENInput_PartialFEN verifies that incomplete FEN strings are rejected.
func TestFENInput_PartialFEN(t *testing.T) {
	m := NewModel()
	m.screen = ScreenFENInput
	// Only has 4 parts instead of 6
	m.fenInput = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq"

	// Simulate Enter key
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.Update(msg)
	m = result.(Model)

	// Should stay on FEN input screen
	if m.screen != ScreenFENInput {
		t.Errorf("Expected ScreenFENInput, got %v", m.screen)
	}

	// Should have error message
	if m.errorMsg == "" {
		t.Error("Should have error message for partial FEN")
	}
}

// TestFENInput_MultipleAttempts verifies that multiple FEN inputs can be attempted.
func TestFENInput_MultipleAttempts(t *testing.T) {
	m := NewModel()
	m.screen = ScreenFENInput

	// First attempt - invalid
	m.fenInput = "invalid"
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.Update(msg)
	m = result.(Model)

	if m.errorMsg == "" {
		t.Error("Should have error message after first invalid attempt")
	}

	// Clear input and try again with valid FEN
	m.fenInput = ""
	chars := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	for _, ch := range chars {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		result, _ = m.Update(msg)
		m = result.(Model)
	}

	// Second attempt - valid
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	result, _ = m.Update(msg)
	m = result.(Model)

	// Should succeed and transition to GamePlay
	if m.screen != ScreenGamePlay {
		t.Errorf("Expected ScreenGamePlay after valid FEN, got %v", m.screen)
	}

	if m.errorMsg != "" {
		t.Errorf("Should have no error after valid FEN, got: %s", m.errorMsg)
	}
}
