package ui

import (
	"strings"
	"testing"

	"github.com/Mgrdich/TermChess/internal/engine"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
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

func TestPieceSymbol_ASCII(t *testing.T) {
	config := Config{
		UseUnicode: false,
		UseColors:  false,
	}
	renderer := NewBoardRenderer(config)

	tests := []struct {
		name     string
		piece    engine.Piece
		expected string
	}{
		{"white pawn", engine.NewPiece(engine.White, engine.Pawn), "P"},
		{"white knight", engine.NewPiece(engine.White, engine.Knight), "N"},
		{"white bishop", engine.NewPiece(engine.White, engine.Bishop), "B"},
		{"white rook", engine.NewPiece(engine.White, engine.Rook), "R"},
		{"white queen", engine.NewPiece(engine.White, engine.Queen), "Q"},
		{"white king", engine.NewPiece(engine.White, engine.King), "K"},
		{"black pawn", engine.NewPiece(engine.Black, engine.Pawn), "p"},
		{"black knight", engine.NewPiece(engine.Black, engine.Knight), "n"},
		{"black bishop", engine.NewPiece(engine.Black, engine.Bishop), "b"},
		{"black rook", engine.NewPiece(engine.Black, engine.Rook), "r"},
		{"black queen", engine.NewPiece(engine.Black, engine.Queen), "q"},
		{"black king", engine.NewPiece(engine.Black, engine.King), "k"},
		{"empty square", engine.Piece(engine.Empty), "."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.pieceSymbol(tt.piece)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestPieceSymbol_Unicode(t *testing.T) {
	config := Config{
		UseUnicode: true,
		UseColors:  false,
	}
	renderer := NewBoardRenderer(config)

	tests := []struct {
		name     string
		piece    engine.Piece
		expected string
	}{
		{"white king", engine.NewPiece(engine.White, engine.King), "♔"},
		{"white queen", engine.NewPiece(engine.White, engine.Queen), "♕"},
		{"white rook", engine.NewPiece(engine.White, engine.Rook), "♖"},
		{"white bishop", engine.NewPiece(engine.White, engine.Bishop), "♗"},
		{"white knight", engine.NewPiece(engine.White, engine.Knight), "♘"},
		{"white pawn", engine.NewPiece(engine.White, engine.Pawn), "♙"},
		{"black king", engine.NewPiece(engine.Black, engine.King), "♚"},
		{"black queen", engine.NewPiece(engine.Black, engine.Queen), "♛"},
		{"black rook", engine.NewPiece(engine.Black, engine.Rook), "♜"},
		{"black bishop", engine.NewPiece(engine.Black, engine.Bishop), "♝"},
		{"black knight", engine.NewPiece(engine.Black, engine.Knight), "♞"},
		{"black pawn", engine.NewPiece(engine.Black, engine.Pawn), "♟"},
		{"empty square", engine.Piece(engine.Empty), "·"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.pieceSymbol(tt.piece)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBoardRenderer_Unicode(t *testing.T) {
	// Create a new board with starting position
	board := engine.NewBoard()

	// Create a renderer with Unicode mode enabled
	config := Config{
		UseUnicode:      true,
		ShowCoords:      true,
		UseColors:       false, // Disable colors for easier testing
		ShowMoveHistory: false,
	}
	renderer := NewBoardRenderer(config)

	// Render the board
	result := renderer.Render(board)

	// Check that Unicode symbols appear in the output
	unicodeSymbols := []string{"♔", "♕", "♖", "♗", "♘", "♙", "♚", "♛", "♜", "♝", "♞", "♟"}
	for _, symbol := range unicodeSymbols {
		if !strings.Contains(result, symbol) {
			t.Errorf("Expected to find Unicode symbol '%s' in output:\n%s", symbol, result)
		}
	}

	// Verify specific pieces in starting position
	// Black back rank (rank 8)
	if !strings.Contains(result, "♜ ♞ ♝ ♛ ♚ ♝ ♞ ♜") {
		t.Errorf("Expected to find black back rank pieces, got:\n%s", result)
	}

	// White back rank (rank 1)
	if !strings.Contains(result, "♖ ♘ ♗ ♕ ♔ ♗ ♘ ♖") {
		t.Errorf("Expected to find white back rank pieces, got:\n%s", result)
	}

	// Should not contain ASCII piece symbols in piece positions
	// We check for patterns that would indicate ASCII pieces are being rendered
	// For lowercase letters, we avoid 'a' through 'h' as they appear in file labels
	if strings.Contains(result, "r n b q k") || strings.Contains(result, "R N B Q K") {
		t.Errorf("Should not contain ASCII piece symbols when UseUnicode is true, got:\n%s", result)
	}

	// Verify coordinates are still present
	if !strings.Contains(result, "a b c d e f g h") {
		t.Errorf("Expected to find file labels, got:\n%s", result)
	}
}

func TestBoardRenderer_ASCII_NoUnicodeSymbols(t *testing.T) {
	// Create a new board with starting position
	board := engine.NewBoard()

	// Create a renderer with ASCII mode (UseUnicode = false)
	config := Config{
		UseUnicode:      false,
		ShowCoords:      true,
		UseColors:       false,
		ShowMoveHistory: false,
	}
	renderer := NewBoardRenderer(config)

	// Render the board
	result := renderer.Render(board)

	// Should not contain Unicode symbols
	unicodeSymbols := []string{"♔", "♕", "♖", "♗", "♘", "♙", "♚", "♛", "♜", "♝", "♞", "♟"}
	for _, symbol := range unicodeSymbols {
		if strings.Contains(result, symbol) {
			t.Errorf("Should not contain Unicode symbol '%s' when UseUnicode is false, got:\n%s", symbol, result)
		}
	}

	// Should contain ASCII symbols
	if !strings.Contains(result, "r n b q k b n r") {
		t.Errorf("Expected to find ASCII black back rank pieces, got:\n%s", result)
	}

	if !strings.Contains(result, "R N B Q K B N R") {
		t.Errorf("Expected to find ASCII white back rank pieces, got:\n%s", result)
	}
}

// TestColorSymbol_WhitePiece tests that white pieces get colored when UseColors is true
func TestColorSymbol_WhitePiece(t *testing.T) {
	// Force color output in tests
	lipgloss.SetColorProfile(termenv.ANSI256)

	config := Config{
		UseUnicode: false,
		UseColors:  true,
	}
	renderer := NewBoardRenderer(config)

	// Test white pawn gets colored
	whitePawn := engine.NewPiece(engine.White, engine.Pawn)
	result := renderer.pieceSymbol(whitePawn)

	// Result should contain ANSI color codes
	if !strings.Contains(result, "\x1b[") {
		t.Errorf("Expected colored output to contain ANSI escape codes, got: %q", result)
	}

	// Should contain the piece symbol
	if !strings.Contains(result, "P") {
		t.Errorf("Expected to find 'P' in colored output, got: %s", result)
	}
}

// TestColorSymbol_BlackPiece tests that black pieces get colored when UseColors is true
func TestColorSymbol_BlackPiece(t *testing.T) {
	// Force color output in tests
	lipgloss.SetColorProfile(termenv.ANSI256)

	config := Config{
		UseUnicode: false,
		UseColors:  true,
	}
	renderer := NewBoardRenderer(config)

	// Test black pawn gets colored
	blackPawn := engine.NewPiece(engine.Black, engine.Pawn)
	result := renderer.pieceSymbol(blackPawn)

	// Result should contain ANSI color codes
	if !strings.Contains(result, "\x1b[") {
		t.Errorf("Expected colored output to contain ANSI escape codes, got: %q", result)
	}

	// Should contain the piece symbol
	if !strings.Contains(result, "p") {
		t.Errorf("Expected to find 'p' in colored output, got: %s", result)
	}
}

// TestColorSymbol_DifferentColors tests that white and black pieces have different colors
func TestColorSymbol_DifferentColors(t *testing.T) {
	// Force color output in tests
	lipgloss.SetColorProfile(termenv.ANSI256)

	config := Config{
		UseUnicode: false,
		UseColors:  true,
	}
	renderer := NewBoardRenderer(config)

	whitePawn := engine.NewPiece(engine.White, engine.Pawn)
	blackPawn := engine.NewPiece(engine.Black, engine.Pawn)

	whiteResult := renderer.pieceSymbol(whitePawn)
	blackResult := renderer.pieceSymbol(blackPawn)

	// White and black should have different ANSI escape sequences
	if whiteResult == blackResult {
		t.Error("White and black pieces should have different color codes")
	}

	// Both should contain ANSI color codes
	if !strings.Contains(whiteResult, "\x1b[") || !strings.Contains(blackResult, "\x1b[") {
		t.Errorf("Both white and black pieces should contain ANSI escape codes when UseColors is true. White: %q, Black: %q", whiteResult, blackResult)
	}
}

// TestColorSymbol_Disabled tests that colors are not applied when UseColors is false
func TestColorSymbol_Disabled(t *testing.T) {
	config := Config{
		UseUnicode: false,
		UseColors:  false, // Colors disabled
	}
	renderer := NewBoardRenderer(config)

	whitePawn := engine.NewPiece(engine.White, engine.Pawn)
	result := renderer.pieceSymbol(whitePawn)

	// Should not contain ANSI color codes when colors are disabled
	if strings.Contains(result, "\x1b[") {
		t.Error("Should not contain ANSI escape codes when UseColors is false")
	}

	// Should just be "P" for white pawn in ASCII mode
	if result != "P" {
		t.Errorf("Expected 'P', got '%s'", result)
	}
}

// TestColorSymbol_EmptySquare tests that empty squares are never colored
func TestColorSymbol_EmptySquare(t *testing.T) {
	config := Config{
		UseUnicode: false,
		UseColors:  true,
	}
	renderer := NewBoardRenderer(config)

	emptyPiece := engine.Piece(engine.Empty)
	result := renderer.pieceSymbol(emptyPiece)

	// Empty squares should not be colored even when UseColors is true
	if strings.Contains(result, "\x1b[") {
		t.Error("Empty squares should not be colored")
	}

	if result != "." {
		t.Errorf("Expected '.', got '%s'", result)
	}
}

// TestColorSymbol_Unicode tests that colors work with Unicode pieces
func TestColorSymbol_Unicode(t *testing.T) {
	// Force color output in tests
	lipgloss.SetColorProfile(termenv.ANSI256)

	config := Config{
		UseUnicode: true,
		UseColors:  true,
	}
	renderer := NewBoardRenderer(config)

	whiteKing := engine.NewPiece(engine.White, engine.King)
	blackKing := engine.NewPiece(engine.Black, engine.King)

	whiteResult := renderer.pieceSymbol(whiteKing)
	blackResult := renderer.pieceSymbol(blackKing)

	// Both should contain ANSI color codes
	if !strings.Contains(whiteResult, "\x1b[") {
		t.Errorf("White Unicode piece should be colored, got: %q", whiteResult)
	}

	if !strings.Contains(blackResult, "\x1b[") {
		t.Errorf("Black Unicode piece should be colored, got: %q", blackResult)
	}

	// Should contain Unicode symbols
	if !strings.Contains(whiteResult, "♔") {
		t.Errorf("Expected to find '♔' in colored output, got: %s", whiteResult)
	}

	if !strings.Contains(blackResult, "♚") {
		t.Errorf("Expected to find '♚' in colored output, got: %s", blackResult)
	}

	// Colors should be different
	if whiteResult == blackResult {
		t.Error("White and black Unicode pieces should have different colors")
	}
}

// TestColorSymbol_AllPieceTypes tests that all piece types can be colored
func TestColorSymbol_AllPieceTypes(t *testing.T) {
	// Force color output in tests
	lipgloss.SetColorProfile(termenv.ANSI256)

	config := Config{
		UseUnicode: false,
		UseColors:  true,
	}
	renderer := NewBoardRenderer(config)

	pieceTypes := []engine.PieceType{
		engine.Pawn, engine.Knight, engine.Bishop,
		engine.Rook, engine.Queen, engine.King,
	}

	for _, pt := range pieceTypes {
		whitePiece := engine.NewPiece(engine.White, pt)
		blackPiece := engine.NewPiece(engine.Black, pt)

		whiteResult := renderer.pieceSymbol(whitePiece)
		blackResult := renderer.pieceSymbol(blackPiece)

		// Both should contain ANSI color codes
		if !strings.Contains(whiteResult, "\x1b[") {
			t.Errorf("White %v should be colored, got: %q", pt, whiteResult)
		}

		if !strings.Contains(blackResult, "\x1b[") {
			t.Errorf("Black %v should be colored, got: %q", pt, blackResult)
		}

		// Colors should be different
		if whiteResult == blackResult {
			t.Errorf("White and black %v should have different colors", pt)
		}
	}
}
