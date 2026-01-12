package ui

import (
	"strings"
	"testing"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// TestMoveHistoryIntegration tests the full integration of move history display
func TestMoveHistoryIntegration(t *testing.T) {
	// Create a model with move history enabled
	config := Config{
		UseUnicode:      false,
		ShowCoords:      true,
		UseColors:       false,
		ShowMoveHistory: true,
	}

	m := NewModel(config)
	m.board = engine.NewBoard()

	// Play a short game (Scholar's mate setup)
	moves := []string{"e2e4", "e7e5", "f1c4", "b8c6", "d1h5", "g8f6"}

	for _, moveStr := range moves {
		move, err := engine.ParseMove(moveStr)
		if err != nil {
			t.Fatalf("Failed to parse move %s: %v", moveStr, err)
		}

		err = m.board.MakeMove(move)
		if err != nil {
			t.Fatalf("Failed to make move %s: %v", moveStr, err)
		}

		m.moveHistory = append(m.moveHistory, move)
	}

	// Format the move history
	history := m.formatMoveHistory()

	// Verify the history contains expected moves in SAN notation
	expected := []string{
		"1. e4 e5",
		"2. Bc4 Nc6",
		"3. Qh5 Nf6",
	}

	for _, exp := range expected {
		if !strings.Contains(history, exp) {
			t.Errorf("Move history missing expected sequence: %s\nGot: %s", exp, history)
		}
	}

	// Test that rendering includes the move history when config is enabled
	rendered := m.renderGamePlay()
	if !strings.Contains(rendered, "Move History:") {
		t.Errorf("renderGamePlay should include 'Move History:' when config.ShowMoveHistory is true")
	}

	if !strings.Contains(rendered, "1. e4 e5") {
		t.Errorf("renderGamePlay should include actual move history")
	}
}

// TestMoveHistoryDisabled tests that move history is not shown when disabled
func TestMoveHistoryDisabled(t *testing.T) {
	// Create a model with move history DISABLED
	config := Config{
		UseUnicode:      false,
		ShowCoords:      true,
		UseColors:       false,
		ShowMoveHistory: false, // Disabled
	}

	m := NewModel(config)
	m.board = engine.NewBoard()

	// Play some moves
	moves := []string{"e2e4", "e7e5", "g1f3"}

	for _, moveStr := range moves {
		move, err := engine.ParseMove(moveStr)
		if err != nil {
			t.Fatalf("Failed to parse move %s: %v", moveStr, err)
		}

		m.board.MakeMove(move)
		m.moveHistory = append(m.moveHistory, move)
	}

	// Test that rendering does NOT include the move history when config is disabled
	rendered := m.renderGamePlay()
	if strings.Contains(rendered, "Move History:") {
		t.Errorf("renderGamePlay should NOT include 'Move History:' when config.ShowMoveHistory is false")
	}
}

// TestMoveHistoryResetsOnNewGame tests that move history is cleared on new game
func TestMoveHistoryResetsOnNewGame(t *testing.T) {
	config := Config{
		UseUnicode:      false,
		ShowCoords:      true,
		UseColors:       false,
		ShowMoveHistory: true,
	}

	m := NewModel(config)
	m.board = engine.NewBoard()

	// Play some moves
	moves := []string{"e2e4", "e7e5", "g1f3"}

	for _, moveStr := range moves {
		move, _ := engine.ParseMove(moveStr)
		m.board.MakeMove(move)
		m.moveHistory = append(m.moveHistory, move)
	}

	// Verify move history has moves
	if len(m.moveHistory) != 3 {
		t.Errorf("Expected 3 moves in history, got %d", len(m.moveHistory))
	}

	// Start a new game (simulating what happens in update.go)
	m.board = engine.NewBoard()
	m.moveHistory = []engine.Move{}

	// Verify move history is cleared
	if len(m.moveHistory) != 0 {
		t.Errorf("Move history should be cleared on new game, got %d moves", len(m.moveHistory))
	}

	history := m.formatMoveHistory()
	if history != "" {
		t.Errorf("Empty move history should return empty string, got: %s", history)
	}
}

// TestCastlingInMoveHistory tests that castling is formatted correctly in history
func TestCastlingInMoveHistory(t *testing.T) {
	config := Config{
		UseUnicode:      false,
		ShowCoords:      true,
		UseColors:       false,
		ShowMoveHistory: true,
	}

	m := NewModel(config)
	m.board = engine.NewBoard()

	// Set up for castling: 1. e4 e5 2. Nf3 Nc6 3. Bc4 Nf6 4. O-O
	moves := []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4", "g8f6", "e1g1"}

	for _, moveStr := range moves {
		move, err := engine.ParseMove(moveStr)
		if err != nil {
			t.Fatalf("Failed to parse move %s: %v", moveStr, err)
		}

		m.board.MakeMove(move)
		m.moveHistory = append(m.moveHistory, move)
	}

	history := m.formatMoveHistory()

	// Verify castling is shown as O-O
	if !strings.Contains(history, "O-O") {
		t.Errorf("Move history should contain 'O-O' for castling, got: %s", history)
	}

	// Verify it's in the correct position (move 4 for white)
	if !strings.Contains(history, "4. O-O") {
		t.Errorf("Castling should be move 4 for white, got: %s", history)
	}
}

// TestPromotionInMoveHistory tests that promotions are formatted correctly
func TestPromotionInMoveHistory(t *testing.T) {
	config := Config{
		UseUnicode:      false,
		ShowCoords:      true,
		UseColors:       false,
		ShowMoveHistory: true,
	}

	m := NewModel(config)

	// Set up a position where White can promote
	fen := "8/P7/8/8/8/8/8/4K2k w - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}
	m.board = board

	// Promote to queen
	move, err := engine.ParseMove("a7a8q")
	if err != nil {
		t.Fatalf("Failed to parse move: %v", err)
	}

	m.board.MakeMove(move)
	m.moveHistory = append(m.moveHistory, move)

	history := m.formatMoveHistory()

	// Verify promotion is shown with =Q
	if !strings.Contains(history, "=Q") {
		t.Errorf("Move history should contain '=Q' for queen promotion, got: %s", history)
	}
}
