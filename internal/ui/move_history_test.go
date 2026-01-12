package ui

import (
	"strings"
	"testing"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// TestMoveHistoryAppend tests that moves are properly appended to move history.
func TestMoveHistoryAppend(t *testing.T) {
	m := NewModel()
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay
	m.moveHistory = []engine.Move{}

	// Parse and make a move
	move, err := engine.ParseMove("e2e4")
	if err != nil {
		t.Fatalf("failed to parse move: %v", err)
	}

	err = m.board.MakeMove(move)
	if err != nil {
		t.Fatalf("failed to make move: %v", err)
	}

	m.moveHistory = append(m.moveHistory, move)

	if len(m.moveHistory) != 1 {
		t.Errorf("expected 1 move in history, got %d", len(m.moveHistory))
	}

	if m.moveHistory[0].String() != "e2e4" {
		t.Errorf("expected e2e4, got %s", m.moveHistory[0].String())
	}
}

// TestMoveHistoryMultipleMoves tests appending multiple moves.
func TestMoveHistoryMultipleMoves(t *testing.T) {
	m := NewModel()
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay
	m.moveHistory = []engine.Move{}

	moves := []string{"e2e4", "e7e5", "g1f3", "b8c6"}

	for i, moveStr := range moves {
		move, err := engine.ParseMove(moveStr)
		if err != nil {
			t.Fatalf("failed to parse move %d: %v", i+1, err)
		}

		err = m.board.MakeMove(move)
		if err != nil {
			t.Fatalf("failed to make move %d: %v", i+1, err)
		}

		m.moveHistory = append(m.moveHistory, move)
	}

	if len(m.moveHistory) != 4 {
		t.Errorf("expected 4 moves in history, got %d", len(m.moveHistory))
	}

	// Verify all moves are stored correctly
	for i, moveStr := range moves {
		if m.moveHistory[i].String() != moveStr {
			t.Errorf("move %d: expected %s, got %s", i+1, moveStr, m.moveHistory[i].String())
		}
	}
}

// TestMoveHistoryDisplayEnabled tests that move history is shown when config is enabled.
func TestMoveHistoryDisplayEnabled(t *testing.T) {
	m := NewModel()
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay
	m.config.ShowMoveHistory = true

	// Add some moves to history
	m.moveHistory = []engine.Move{
		{From: parseSquareHelper("e2"), To: parseSquareHelper("e4")},
		{From: parseSquareHelper("e7"), To: parseSquareHelper("e5")},
	}

	view := m.View()

	// Should contain move history header
	if !strings.Contains(view, "Move History:") {
		t.Error("expected move history header in view")
	}

	// Should contain moves
	if !strings.Contains(view, "e2e4") {
		t.Error("expected moves in view")
	}

	if !strings.Contains(view, "e7e5") {
		t.Error("expected moves in view")
	}
}

// TestMoveHistoryDisplayDisabled tests that move history is hidden when config is disabled.
func TestMoveHistoryDisplayDisabled(t *testing.T) {
	m := NewModel()
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay
	m.config.ShowMoveHistory = false // Disabled

	// Add moves to history
	m.moveHistory = []engine.Move{
		{From: parseSquareHelper("e2"), To: parseSquareHelper("e4")},
	}

	view := m.View()

	// Should NOT contain move history
	if strings.Contains(view, "Move History:") {
		t.Error("move history should be hidden when config is false")
	}
}

// TestMoveHistoryDisplayEmpty tests that move history section is not shown when empty.
func TestMoveHistoryDisplayEmpty(t *testing.T) {
	m := NewModel()
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay
	m.config.ShowMoveHistory = true
	m.moveHistory = []engine.Move{} // Empty

	view := m.View()

	// Should NOT contain move history header when empty
	if strings.Contains(view, "Move History:") {
		t.Error("move history should not be shown when empty")
	}
}

// TestMoveHistoryClearedOnNewGame tests that move history is cleared when starting a new game.
func TestMoveHistoryClearedOnNewGame(t *testing.T) {
	m := NewModel()
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// Add some moves to history
	m.moveHistory = []engine.Move{
		{From: parseSquareHelper("e2"), To: parseSquareHelper("e4")},
		{From: parseSquareHelper("e7"), To: parseSquareHelper("e5")},
	}

	// Simulate starting a new game from GameOver screen
	m.screen = ScreenGameOver
	m.board = engine.NewBoard()
	m.moveHistory = []engine.Move{} // This is what should happen

	if len(m.moveHistory) != 0 {
		t.Errorf("expected move history to be cleared, got %d moves", len(m.moveHistory))
	}
}

// TestMoveHistoryClearedOnFENLoad tests that move history is cleared when loading from FEN.
func TestMoveHistoryClearedOnFENLoad(t *testing.T) {
	m := NewModel()
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay

	// Add some moves to history
	m.moveHistory = []engine.Move{
		{From: parseSquareHelper("e2"), To: parseSquareHelper("e4")},
		{From: parseSquareHelper("e7"), To: parseSquareHelper("e5")},
	}

	// Simulate loading a FEN position
	fenString := "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"
	board, err := engine.FromFEN(fenString)
	if err != nil {
		t.Fatalf("failed to parse FEN: %v", err)
	}

	m.board = board
	m.moveHistory = []engine.Move{} // This is what should happen when loading FEN

	if len(m.moveHistory) != 0 {
		t.Errorf("expected move history to be cleared when loading FEN, got %d moves", len(m.moveHistory))
	}
}

// TestMoveHistoryFormatting tests the formatting of move history in the view.
func TestMoveHistoryFormatting(t *testing.T) {
	m := NewModel()
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay
	m.config.ShowMoveHistory = true

	// Add a sequence of moves
	m.moveHistory = []engine.Move{
		{From: parseSquareHelper("e2"), To: parseSquareHelper("e4")},
		{From: parseSquareHelper("e7"), To: parseSquareHelper("e5")},
		{From: parseSquareHelper("g1"), To: parseSquareHelper("f3")},
		{From: parseSquareHelper("b8"), To: parseSquareHelper("c6")},
	}

	view := m.View()

	// Should contain properly formatted move pairs with numbers
	if !strings.Contains(view, "1. e2e4 e7e5") {
		t.Error("expected formatted first move pair '1. e2e4 e7e5'")
	}

	if !strings.Contains(view, "2. g1f3 b8c6") {
		t.Error("expected formatted second move pair '2. g1f3 b8c6'")
	}
}

// TestMoveHistoryWithPromotion tests move history display with pawn promotions.
func TestMoveHistoryWithPromotion(t *testing.T) {
	m := NewModel()
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay
	m.config.ShowMoveHistory = true

	// Add a move with promotion
	m.moveHistory = []engine.Move{
		{From: parseSquareHelper("e2"), To: parseSquareHelper("e4")},
		{
			From:      parseSquareHelper("e7"),
			To:        parseSquareHelper("e8"),
			Promotion: engine.Queen,
		},
	}

	view := m.View()

	// Should contain promotion notation
	if !strings.Contains(view, "e7e8=Q") {
		t.Error("expected promotion notation 'e7e8=Q' in move history")
	}
}

// TestMoveHistoryOddNumberOfMoves tests formatting with an odd number of moves.
func TestMoveHistoryOddNumberOfMoves(t *testing.T) {
	m := NewModel()
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay
	m.config.ShowMoveHistory = true

	// Add three moves (white has an extra move)
	m.moveHistory = []engine.Move{
		{From: parseSquareHelper("e2"), To: parseSquareHelper("e4")},
		{From: parseSquareHelper("e7"), To: parseSquareHelper("e5")},
		{From: parseSquareHelper("g1"), To: parseSquareHelper("f3")},
	}

	view := m.View()

	// Should show the incomplete pair correctly
	if !strings.Contains(view, "1. e2e4 e7e5") {
		t.Error("expected complete first move pair")
	}

	if !strings.Contains(view, "2. g1f3") {
		t.Error("expected second move with only white's move")
	}

	// Should not show a black move after the second move number
	// The format should be "2. g1f3" without a black move following it
	historyText := FormatMoveHistory(m.moveHistory)
	expected := "1. e2e4 e7e5 2. g1f3"
	if !strings.HasPrefix(historyText, expected) {
		t.Errorf("expected history to start with '%s', got '%s'", expected, historyText)
	}
}

// TestMoveHistoryIntegrationWithGamePlay tests the full integration of move history
// with actual game play.
func TestMoveHistoryIntegrationWithGamePlay(t *testing.T) {
	m := NewModel()
	m.board = engine.NewBoard()
	m.screen = ScreenGamePlay
	m.config.ShowMoveHistory = true
	m.moveHistory = []engine.Move{}

	// Simulate a short game
	movesToPlay := []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4"}

	for i, moveStr := range movesToPlay {
		move, err := engine.ParseMove(moveStr)
		if err != nil {
			t.Fatalf("move %d: failed to parse: %v", i+1, err)
		}

		err = m.board.MakeMove(move)
		if err != nil {
			t.Fatalf("move %d: failed to make move: %v", i+1, err)
		}

		m.moveHistory = append(m.moveHistory, move)
	}

	// Check that all moves are in history
	if len(m.moveHistory) != 5 {
		t.Errorf("expected 5 moves in history, got %d", len(m.moveHistory))
	}

	// Check that view contains the move history
	view := m.View()

	if !strings.Contains(view, "Move History:") {
		t.Error("expected move history header in view")
	}

	// Check for specific moves
	expectedMoves := []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4"}
	for _, moveStr := range expectedMoves {
		if !strings.Contains(view, moveStr) {
			t.Errorf("expected move %s in view", moveStr)
		}
	}

	// Check for proper formatting with move numbers
	if !strings.Contains(view, "1. e2e4 e7e5") {
		t.Error("expected move pair '1. e2e4 e7e5'")
	}

	if !strings.Contains(view, "2. g1f3 b8c6") {
		t.Error("expected move pair '2. g1f3 b8c6'")
	}

	if !strings.Contains(view, "3. f1c4") {
		t.Error("expected incomplete move '3. f1c4'")
	}
}
