package ui

import (
	"strings"
	"testing"

	"github.com/Mgrdich/TermChess/internal/engine"
)

func TestBoardRenderer_ASCII(t *testing.T) {
	// Create a new board with starting position
	board := engine.NewBoard()

	// Create a renderer with ASCII mode and no colors (for easier testing)
	config := Config{
		UseUnicode:      false,
		ShowCoords:      true,
		UseColors:       false, // Disable colors for testing
		ShowMoveHistory: false,
	}
	renderer := NewBoardRenderer(config)

	// Render the board
	result := renderer.Render(board)

	// Verify that the result contains expected elements
	if !strings.Contains(result, "r n b q k b n r") {
		t.Errorf("Expected to find black back rank pieces, got:\n%s", result)
	}

	if !strings.Contains(result, "R N B Q K B N R") {
		t.Errorf("Expected to find white back rank pieces, got:\n%s", result)
	}

	if !strings.Contains(result, "a b c d e f g h") {
		t.Errorf("Expected to find file labels, got:\n%s", result)
	}

	// Check that rank numbers are present
	for rank := 1; rank <= 8; rank++ {
		rankStr := string(rune('0' + rank))
		if !strings.Contains(result, rankStr) {
			t.Errorf("Expected to find rank %d, got:\n%s", rank, result)
		}
	}
}

func TestBoardRenderer_NilBoard(t *testing.T) {
	config := DefaultConfig()
	renderer := NewBoardRenderer(config)

	result := renderer.Render(nil)

	if result != "No board available" {
		t.Errorf("Expected 'No board available', got: %s", result)
	}
}

func TestPieceSymbol(t *testing.T) {
	config := Config{
		UseUnicode: false,
		UseColors:  false,
	}
	renderer := NewBoardRenderer(config)

	tests := []struct {
		piece    engine.Piece
		expected string
	}{
		{engine.NewPiece(engine.White, engine.Pawn), "P"},
		{engine.NewPiece(engine.White, engine.Knight), "N"},
		{engine.NewPiece(engine.White, engine.Bishop), "B"},
		{engine.NewPiece(engine.White, engine.Rook), "R"},
		{engine.NewPiece(engine.White, engine.Queen), "Q"},
		{engine.NewPiece(engine.White, engine.King), "K"},
		{engine.NewPiece(engine.Black, engine.Pawn), "p"},
		{engine.NewPiece(engine.Black, engine.Knight), "n"},
		{engine.NewPiece(engine.Black, engine.Bishop), "b"},
		{engine.NewPiece(engine.Black, engine.Rook), "r"},
		{engine.NewPiece(engine.Black, engine.Queen), "q"},
		{engine.NewPiece(engine.Black, engine.King), "k"},
		{engine.Piece(engine.Empty), "."},
	}

	for _, tt := range tests {
		result := renderer.pieceSymbol(tt.piece)
		if result != tt.expected {
			t.Errorf("Expected piece symbol '%s' for piece type %d color %d, got '%s'",
				tt.expected, tt.piece.Type(), tt.piece.Color(), result)
		}
	}
}
