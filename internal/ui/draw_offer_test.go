package ui

import (
	"testing"

	"github.com/Mgrdich/TermChess/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

// TestOfferDraw tests that a player can offer a draw
func TestOfferDraw(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// White offers a draw
	m.input = "offerdraw"
	newModel, _ := m.handleGamePlayInput()
	m = newModel.(Model)

	// Check that we're now on the draw prompt screen
	if m.screen != ScreenDrawPrompt {
		t.Errorf("Expected screen to be ScreenDrawPrompt, got %v", m.screen)
	}

	// Check that White is marked as having offered
	if m.drawOfferedBy != int8(engine.White) {
		t.Errorf("Expected drawOfferedBy to be White (0), got %d", m.drawOfferedBy)
	}

	if !m.drawOfferedByWhite {
		t.Error("Expected drawOfferedByWhite to be true")
	}

	// Check that input was cleared
	if m.input != "" {
		t.Errorf("Expected input to be cleared, got %s", m.input)
	}
}

// TestOfferDrawCaseInsensitive tests that the offerdraw command is case-insensitive
func TestOfferDrawCaseInsensitive(t *testing.T) {
	testCases := []string{
		"offerdraw",
		"OFFERDRAW",
		"OfferDraw",
		"offerDRAW",
	}

	for _, input := range testCases {
		m := NewModel(DefaultConfig())
		m.board = engine.NewBoard()
		m.screen = ScreenGamePlay

		m.input = input
		newModel, _ := m.handleGamePlayInput()
		m = newModel.(Model)

		if m.screen != ScreenDrawPrompt {
			t.Errorf("Input %s: Expected screen to be ScreenDrawPrompt, got %v", input, m.screen)
		}
	}
}

// TestAcceptDrawOffer tests that accepting a draw offer ends the game
func TestAcceptDrawOffer(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// White offers a draw
	m.input = "offerdraw"
	newModel, _ := m.handleGamePlayInput()
	m = newModel.(Model)

	// Verify we're on the draw prompt screen
	if m.screen != ScreenDrawPrompt {
		t.Fatal("Expected to be on draw prompt screen")
	}

	// Black accepts (selection 0 is Accept)
	m.drawPromptSelection = 0
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ = m.handleDrawPromptKeys(msg)
	m = newModel.(Model)

	// Check that the game ended
	if m.screen != ScreenGameOver {
		t.Errorf("Expected screen to be ScreenGameOver, got %v", m.screen)
	}

	// Check that draw by agreement is set
	if !m.drawByAgreement {
		t.Error("Expected drawByAgreement to be true")
	}
}

// TestDeclineDrawOffer tests that declining a draw offer continues the game
func TestDeclineDrawOffer(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// White offers a draw
	m.input = "offerdraw"
	newModel, _ := m.handleGamePlayInput()
	m = newModel.(Model)

	// Verify we're on the draw prompt screen
	if m.screen != ScreenDrawPrompt {
		t.Fatal("Expected to be on draw prompt screen")
	}

	// Black declines (selection 1 is Decline)
	m.drawPromptSelection = 1
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ = m.handleDrawPromptKeys(msg)
	m = newModel.(Model)

	// Check that we're back to gameplay
	if m.screen != ScreenGamePlay {
		t.Errorf("Expected screen to be ScreenGamePlay, got %v", m.screen)
	}

	// Check that draw by agreement is not set
	if m.drawByAgreement {
		t.Error("Expected drawByAgreement to be false")
	}

	// Check status message
	if m.statusMsg != "Draw offer declined" {
		t.Errorf("Expected status message 'Draw offer declined', got %s", m.statusMsg)
	}

	// Check that draw offered by is reset
	if m.drawOfferedBy != -1 {
		t.Errorf("Expected drawOfferedBy to be -1, got %d", m.drawOfferedBy)
	}
}

// TestDrawOfferSpamPrevention tests that a player can't offer draw twice
func TestDrawOfferSpamPrevention(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// White offers a draw
	m.input = "offerdraw"
	newModel, _ := m.handleGamePlayInput()
	m = newModel.(Model)

	// Black declines
	m.drawPromptSelection = 1
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ = m.handleDrawPromptKeys(msg)
	m = newModel.(Model)

	// White tries to offer draw again
	m.input = "offerdraw"
	newModel, _ = m.handleGamePlayInput()
	m = newModel.(Model)

	// Check that we're still on gameplay (not draw prompt)
	if m.screen != ScreenGamePlay {
		t.Errorf("Expected screen to be ScreenGamePlay, got %v", m.screen)
	}

	// Check that an error message was set
	if m.errorMsg != "You have already offered a draw this game" {
		t.Errorf("Expected error message about already offering, got %s", m.errorMsg)
	}
}

// TestDrawOfferBlackCanOffer tests that Black can also offer a draw
func TestDrawOfferBlackCanOffer(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// Make a move to switch to Black's turn
	move, _ := engine.ParseMove("e2e4")
	m.board.MakeMove(move)

	// Black offers a draw
	m.input = "offerdraw"
	newModel, _ := m.handleGamePlayInput()
	m = newModel.(Model)

	// Check that we're now on the draw prompt screen
	if m.screen != ScreenDrawPrompt {
		t.Errorf("Expected screen to be ScreenDrawPrompt, got %v", m.screen)
	}

	// Check that Black is marked as having offered
	if m.drawOfferedBy != int8(engine.Black) {
		t.Errorf("Expected drawOfferedBy to be Black (1), got %d", m.drawOfferedBy)
	}

	if !m.drawOfferedByBlack {
		t.Error("Expected drawOfferedByBlack to be true")
	}
}

// TestDrawOfferBothPlayersCanOfferOnce tests that both players can each offer once
func TestDrawOfferBothPlayersCanOfferOnce(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// White offers a draw
	m.input = "offerdraw"
	newModel, _ := m.handleGamePlayInput()
	m = newModel.(Model)

	// Black declines
	m.drawPromptSelection = 1
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ = m.handleDrawPromptKeys(msg)
	m = newModel.(Model)

	// Make a move to switch to Black's turn
	move, _ := engine.ParseMove("e2e4")
	m.board.MakeMove(move)

	// Black offers a draw (should work since Black hasn't offered yet)
	m.input = "offerdraw"
	newModel, _ = m.handleGamePlayInput()
	m = newModel.(Model)

	// Check that we're on the draw prompt screen
	if m.screen != ScreenDrawPrompt {
		t.Errorf("Expected screen to be ScreenDrawPrompt, got %v", m.screen)
	}

	// Check that Black is marked as having offered
	if m.drawOfferedBy != int8(engine.Black) {
		t.Errorf("Expected drawOfferedBy to be Black (1), got %d", m.drawOfferedBy)
	}
}

// TestDrawOfferEscapeCancel tests that pressing ESC cancels the draw offer
func TestDrawOfferEscapeCancel(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// White offers a draw
	m.input = "offerdraw"
	newModel, _ := m.handleGamePlayInput()
	m = newModel.(Model)

	// Verify we're on the draw prompt screen
	if m.screen != ScreenDrawPrompt {
		t.Fatal("Expected to be on draw prompt screen")
	}

	// Press ESC to cancel
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ = m.handleDrawPromptKeys(msg)
	m = newModel.(Model)

	// Check that we're back to gameplay
	if m.screen != ScreenGamePlay {
		t.Errorf("Expected screen to be ScreenGamePlay, got %v", m.screen)
	}

	// Check status message
	if m.statusMsg != "Draw offer cancelled" {
		t.Errorf("Expected status message 'Draw offer cancelled', got %s", m.statusMsg)
	}

	// Check that draw offer state is fully reset
	if m.drawOfferedBy != -1 {
		t.Errorf("Expected drawOfferedBy to be -1, got %d", m.drawOfferedBy)
	}

	if m.drawOfferedByWhite {
		t.Error("Expected drawOfferedByWhite to be false after cancel")
	}
}

// TestDrawPromptNavigationUpDown tests that arrow keys navigate the draw prompt
func TestDrawPromptNavigationUpDown(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenDrawPrompt
	m.drawPromptSelection = 0

	// Press down to go to Decline
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.handleDrawPromptKeys(msg)
	m = newModel.(Model)

	if m.drawPromptSelection != 1 {
		t.Errorf("Expected selection to be 1 (Decline), got %d", m.drawPromptSelection)
	}

	// Press up to go back to Accept
	msg = tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = m.handleDrawPromptKeys(msg)
	m = newModel.(Model)

	if m.drawPromptSelection != 0 {
		t.Errorf("Expected selection to be 0 (Accept), got %d", m.drawPromptSelection)
	}

	// Test wrapping - press up from Accept should go to Decline
	msg = tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = m.handleDrawPromptKeys(msg)
	m = newModel.(Model)

	if m.drawPromptSelection != 1 {
		t.Errorf("Expected selection to wrap to 1 (Decline), got %d", m.drawPromptSelection)
	}

	// Test wrapping - press down from Decline should go to Accept
	msg = tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ = m.handleDrawPromptKeys(msg)
	m = newModel.(Model)

	if m.drawPromptSelection != 0 {
		t.Errorf("Expected selection to wrap to 0 (Accept), got %d", m.drawPromptSelection)
	}
}

// TestDrawByAgreementMessage tests that the game over message shows "Draw by agreement"
func TestDrawByAgreementMessage(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.drawByAgreement = true

	msg := getGameResultMessage(m.board, -1, true)
	expected := "Draw by agreement"

	if msg != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, msg)
	}
}

// TestNewGameResetsDrawState tests that starting a new game resets draw offer state
func TestNewGameResetsDrawState(t *testing.T) {
	m := NewModel(DefaultConfig())
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// Set draw offer state
	m.drawOfferedBy = int8(engine.White)
	m.drawOfferedByWhite = true
	m.drawByAgreement = true

	// Start a new game via game type selection
	m.screen = ScreenGameTypeSelect
	m.menuOptions = []string{"Player vs Player", "Player vs Bot"}
	m.menuSelection = 0

	newModel, _ := m.handleGameTypeSelection()
	m = newModel.(Model)

	// Check that draw state is reset
	if m.drawOfferedBy != -1 {
		t.Errorf("Expected drawOfferedBy to be -1, got %d", m.drawOfferedBy)
	}

	if m.drawOfferedByWhite {
		t.Error("Expected drawOfferedByWhite to be false")
	}

	if m.drawOfferedByBlack {
		t.Error("Expected drawOfferedByBlack to be false")
	}

	if m.drawByAgreement {
		t.Error("Expected drawByAgreement to be false")
	}
}
