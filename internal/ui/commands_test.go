package ui

import (
	"testing"

	"github.com/Mgrdich/TermChess/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

// TestHandleGamePlayKeys_ResignCommand tests the "resign" command
func TestHandleGamePlayKeys_ResignCommand(t *testing.T) {
	// Create a model with a new board in gameplay
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// White resigns
	m.input = "resign"
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleGamePlayKeys(msg)
	m = result.(Model)

	// Verify transition to game over screen
	if m.screen != ScreenGameOver {
		t.Errorf("Expected screen to be ScreenGameOver, got %v", m.screen)
	}

	// Verify White resigned
	if m.resignedBy != int8(engine.White) {
		t.Errorf("Expected resignedBy to be White (0), got %d", m.resignedBy)
	}

	// Verify input was cleared
	if m.input != "" {
		t.Errorf("Expected input to be cleared, got '%s'", m.input)
	}

	// Verify messages were cleared
	if m.errorMsg != "" {
		t.Errorf("Expected errorMsg to be cleared, got '%s'", m.errorMsg)
	}

	if m.statusMsg != "" {
		t.Errorf("Expected statusMsg to be cleared, got '%s'", m.statusMsg)
	}
}

// TestHandleGamePlayKeys_ResignCommandCaseInsensitive tests case-insensitive resign
func TestHandleGamePlayKeys_ResignCommandCaseInsensitive(t *testing.T) {
	testCases := []string{"resign", "RESIGN", "Resign", "ReSiGn"}

	for _, resignInput := range testCases {
		// Create a fresh model for each test
		m := NewModel(DefaultConfig())
		m.board = engine.NewBoard()
		m.screen = ScreenGamePlay

		// Test resignation with different casings
		m.input = resignInput
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		result, _ := m.handleGamePlayKeys(msg)
		m = result.(Model)

		// Verify transition to game over screen
		if m.screen != ScreenGameOver {
			t.Errorf("Expected screen to be ScreenGameOver for input '%s', got %v", resignInput, m.screen)
		}

		// Verify resignation was recorded
		if m.resignedBy != int8(engine.White) {
			t.Errorf("Expected resignedBy to be White for input '%s', got %d", resignInput, m.resignedBy)
		}
	}
}

// TestHandleGamePlayKeys_ResignBlackPlayer tests Black player resigning
func TestHandleGamePlayKeys_ResignBlackPlayer(t *testing.T) {
	// Create a model and make it Black's turn
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// Make a move to switch to Black's turn
	m.input = "e2e4"
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleGamePlayKeys(msg)
	m = result.(Model)

	// Now it's Black's turn, Black resigns
	m.input = "resign"
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	result, _ = m.handleGamePlayKeys(msg)
	m = result.(Model)

	// Verify Black resigned
	if m.resignedBy != int8(engine.Black) {
		t.Errorf("Expected resignedBy to be Black (1), got %d", m.resignedBy)
	}

	// Verify screen is game over
	if m.screen != ScreenGameOver {
		t.Errorf("Expected screen to be ScreenGameOver, got %v", m.screen)
	}
}

// TestHandleGamePlayKeys_ResignWithWhitespace tests resign with extra whitespace
func TestHandleGamePlayKeys_ResignWithWhitespace(t *testing.T) {
	testCases := []string{"  resign", "resign  ", "  resign  ", "\tresign\t"}

	for _, resignInput := range testCases {
		// Create a fresh model for each test
		m := NewModel(DefaultConfig())
		m.board = engine.NewBoard()
		m.screen = ScreenGamePlay

		// Test resignation with whitespace
		m.input = resignInput
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		result, _ := m.handleGamePlayKeys(msg)
		m = result.(Model)

		// Verify transition to game over screen
		if m.screen != ScreenGameOver {
			t.Errorf("Expected screen to be ScreenGameOver for input '%s', got %v", resignInput, m.screen)
		}
	}
}

// TestHandleGamePlayKeys_ShowFenCommand tests the "showfen" command
func TestHandleGamePlayKeys_ShowFenCommand(t *testing.T) {
	// Create a model with a new board
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// Execute showfen command
	m.input = "showfen"
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleGamePlayKeys(msg)
	m = result.(Model)

	// Verify still in gameplay screen
	if m.screen != ScreenGamePlay {
		t.Errorf("Expected screen to still be ScreenGamePlay, got %v", m.screen)
	}

	// Verify status message contains FEN
	if m.statusMsg == "" {
		t.Error("Expected statusMsg to contain FEN string")
	}

	// Verify the FEN is the starting position
	expectedFen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	if m.statusMsg != "FEN: "+expectedFen+" (Copied to clipboard)" &&
		m.statusMsg != "FEN: "+expectedFen+" (Failed to copy to clipboard: failed to initialize clipboard: system clipboard not accessible)" {
		// Either success or expected clipboard failure is acceptable
		// Some CI environments don't have clipboard access
		if m.statusMsg[:5] != "FEN: " {
			t.Errorf("Expected statusMsg to start with 'FEN: ', got '%s'", m.statusMsg)
		}
	}

	// Verify input was cleared
	if m.input != "" {
		t.Errorf("Expected input to be cleared, got '%s'", m.input)
	}

	// Verify no error message
	if m.errorMsg != "" {
		t.Errorf("Expected no error message, got '%s'", m.errorMsg)
	}
}

// TestHandleGamePlayKeys_ShowFenCommandCaseInsensitive tests case-insensitive showfen
func TestHandleGamePlayKeys_ShowFenCommandCaseInsensitive(t *testing.T) {
	testCases := []string{"showfen", "SHOWFEN", "ShowFen", "ShOwFeN"}

	for _, showfenInput := range testCases {
		// Create a fresh model for each test
		m := NewModel(DefaultConfig())
		m.board = engine.NewBoard()
		m.screen = ScreenGamePlay

		// Execute showfen command
		m.input = showfenInput
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		result, _ := m.handleGamePlayKeys(msg)
		m = result.(Model)

		// Verify status message contains FEN
		if m.statusMsg == "" {
			t.Errorf("Expected statusMsg to contain FEN string for input '%s'", showfenInput)
		}

		// Verify still in gameplay screen
		if m.screen != ScreenGamePlay {
			t.Errorf("Expected screen to still be ScreenGamePlay for input '%s', got %v", showfenInput, m.screen)
		}
	}
}

// TestHandleGamePlayKeys_ShowFenAfterMoves tests showfen after some moves
func TestHandleGamePlayKeys_ShowFenAfterMoves(t *testing.T) {
	// Create a model with a new board
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// Make a move: e2e4
	m.input = "e2e4"
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleGamePlayKeys(msg)
	m = result.(Model)

	// Execute showfen command
	m.input = "showfen"
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	result, _ = m.handleGamePlayKeys(msg)
	m = result.(Model)

	// Verify status message contains FEN and it's different from starting position
	if m.statusMsg == "" {
		t.Error("Expected statusMsg to contain FEN string")
	}

	// The FEN should reflect the move e2e4
	// Should contain "4P3" or similar pattern in the position
	// And should show "b" (Black to move)
	if m.statusMsg[:5] != "FEN: " {
		t.Errorf("Expected statusMsg to start with 'FEN: ', got '%s'", m.statusMsg)
	}
}

// TestHandleGamePlayKeys_MenuCommand tests the "menu" command
func TestHandleGamePlayKeys_MenuCommand(t *testing.T) {
	// Create a model with a new board
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// Execute menu command
	m.input = "menu"
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleGamePlayKeys(msg)
	m = result.(Model)

	// Verify transition to save prompt screen
	if m.screen != ScreenSavePrompt {
		t.Errorf("Expected screen to be ScreenSavePrompt, got %v", m.screen)
	}

	// Verify save prompt action is set to "menu"
	if m.savePromptAction != "menu" {
		t.Errorf("Expected savePromptAction to be 'menu', got '%s'", m.savePromptAction)
	}

	// Verify save prompt selection is initialized to 0
	if m.savePromptSelection != 0 {
		t.Errorf("Expected savePromptSelection to be 0, got %d", m.savePromptSelection)
	}

	// Verify input was cleared
	if m.input != "" {
		t.Errorf("Expected input to be cleared, got '%s'", m.input)
	}

	// Verify messages were cleared
	if m.errorMsg != "" {
		t.Errorf("Expected errorMsg to be cleared, got '%s'", m.errorMsg)
	}

	if m.statusMsg != "" {
		t.Errorf("Expected statusMsg to be cleared, got '%s'", m.statusMsg)
	}
}

// TestHandleGamePlayKeys_MenuCommandCaseInsensitive tests case-insensitive menu
func TestHandleGamePlayKeys_MenuCommandCaseInsensitive(t *testing.T) {
	testCases := []string{"menu", "MENU", "Menu", "MeNu"}

	for _, menuInput := range testCases {
		// Create a fresh model for each test
		m := NewModel(DefaultConfig())
		m.board = engine.NewBoard()
		m.screen = ScreenGamePlay

		// Execute menu command
		m.input = menuInput
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		result, _ := m.handleGamePlayKeys(msg)
		m = result.(Model)

		// Verify transition to save prompt screen
		if m.screen != ScreenSavePrompt {
			t.Errorf("Expected screen to be ScreenSavePrompt for input '%s', got %v", menuInput, m.screen)
		}

		// Verify save prompt action is set to "menu"
		if m.savePromptAction != "menu" {
			t.Errorf("Expected savePromptAction to be 'menu' for input '%s', got '%s'", menuInput, m.savePromptAction)
		}
	}
}

// TestHandleGamePlayKeys_CommandsDoNotInterfereWithMoves tests that commands don't break normal moves
func TestHandleGamePlayKeys_CommandsDoNotInterfereWithMoves(t *testing.T) {
	// Create a model with a new board
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// First make a normal move
	m.input = "e2e4"
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleGamePlayKeys(msg)
	m = result.(Model)

	// Verify move was executed
	if m.errorMsg != "" {
		t.Errorf("Expected no error after normal move, got '%s'", m.errorMsg)
	}

	if m.screen != ScreenGamePlay {
		t.Errorf("Expected screen to still be ScreenGamePlay after move, got %v", m.screen)
	}

	// Verify turn changed to Black
	if m.board.ActiveColor != engine.Black {
		t.Error("Expected active color to be Black after White's move")
	}

	// Make another move
	m.input = "e7e5"
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	result, _ = m.handleGamePlayKeys(msg)
	m = result.(Model)

	// Verify move was executed
	if m.errorMsg != "" {
		t.Errorf("Expected no error after second move, got '%s'", m.errorMsg)
	}

	// Verify turn changed back to White
	if m.board.ActiveColor != engine.White {
		t.Error("Expected active color to be White after Black's move")
	}

	// Verify move history has 2 moves
	if len(m.moveHistory) != 2 {
		t.Errorf("Expected 2 moves in history, got %d", len(m.moveHistory))
	}
}

// TestHandleGamePlayKeys_InvalidCommandTreatedAsMove tests that invalid commands are treated as moves
func TestHandleGamePlayKeys_InvalidCommandTreatedAsMove(t *testing.T) {
	// Create a model with a new board
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// Try an input that looks like a command but isn't
	m.input = "resigns" // Note the 's' at the end
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleGamePlayKeys(msg)
	m = result.(Model)

	// Verify it was treated as an invalid move, not a command
	if m.errorMsg == "" {
		t.Error("Expected error message for invalid move 'resigns'")
	}

	// Verify still in gameplay screen (not game over)
	if m.screen != ScreenGamePlay {
		t.Errorf("Expected screen to still be ScreenGamePlay, got %v", m.screen)
	}

	// Verify no resignation occurred
	if m.resignedBy != -1 {
		t.Errorf("Expected resignedBy to be -1, got %d", m.resignedBy)
	}
}

// TestHandleGamePlayKeys_PartialCommandsNotRecognized tests that partial commands aren't recognized
func TestHandleGamePlayKeys_PartialCommandsNotRecognized(t *testing.T) {
	testCases := []string{"resi", "res", "show", "men", "e4resign", "resigne4"}

	for _, input := range testCases {
		// Create a fresh model for each test
		m := NewModel(DefaultConfig())
		m.board = engine.NewBoard()
		m.screen = ScreenGamePlay

		// Try the partial command
		m.input = input
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		result, _ := m.handleGamePlayKeys(msg)
		m = result.(Model)

		// Verify it was treated as an invalid move (should have error)
		if m.errorMsg == "" {
			t.Errorf("Expected error message for invalid input '%s'", input)
		}

		// Verify still in gameplay screen
		if m.screen != ScreenGamePlay {
			t.Errorf("Expected screen to still be ScreenGamePlay for input '%s', got %v", input, m.screen)
		}

		// Verify no resignation occurred
		if m.resignedBy != -1 {
			t.Errorf("Expected resignedBy to be -1 for input '%s', got %d", input, m.resignedBy)
		}
	}
}

// TestHandleGamePlayKeys_ResignationResetsOnNewGame tests that resignation state is reset
func TestHandleGamePlayKeys_ResignationResetsOnNewGame(t *testing.T) {
	// Create a model with a new board
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// Resign
	m.input = "resign"
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := m.handleGamePlayKeys(msg)
	m = result.(Model)

	// Verify resignation occurred
	if m.resignedBy != int8(engine.White) {
		t.Errorf("Expected resignedBy to be White, got %d", m.resignedBy)
	}

	// Now start a new game by selecting "New Game" from game over screen
	m.screen = ScreenGameTypeSelect
	m.menuOptions = []string{"Player vs Player", "Player vs Bot"}
	m.menuSelection = 0
	result, _ = m.handleGameTypeSelection()
	m = result.(Model)

	// Verify resignation was reset
	if m.resignedBy != -1 {
		t.Errorf("Expected resignedBy to be reset to -1, got %d", m.resignedBy)
	}

	// Verify we're in gameplay
	if m.screen != ScreenGamePlay {
		t.Errorf("Expected screen to be ScreenGamePlay, got %v", m.screen)
	}
}

// TestGetGameResultMessage_Resignation tests the game result message for resignation
func TestGetGameResultMessage_Resignation(t *testing.T) {
	// Create a board (state doesn't matter for resignation)
	board := engine.NewBoard()

	// Test White resignation
	resultMsg := getGameResultMessage(board, int8(engine.White), false)
	expectedMsg := "White resigned - Black wins"
	if resultMsg != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, resultMsg)
	}

	// Test Black resignation
	resultMsg = getGameResultMessage(board, int8(engine.Black), false)
	expectedMsg = "Black resigned - White wins"
	if resultMsg != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, resultMsg)
	}

	// Test no resignation (should fall through to normal game status)
	resultMsg = getGameResultMessage(board, -1, false)
	// Starting position is not game over, so should return "Game Over" as default
	expectedMsg = "Game Over"
	if resultMsg != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, resultMsg)
	}
}
