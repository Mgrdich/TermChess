package ui

import (
	"github.com/Mgrdich/TermChess/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

// Board rendering constants for mouse coordinate calculation.
// These values are based on the renderGamePlay() layout in view.go:
// - Title with Padding(1, 0) = 3 lines (padding above, text, padding below)
// - "\n\n" after title = 2 more lines
// - Board starts at row 4 (0-indexed)
const (
	// boardStartY is the terminal row where the board's first rank (rank 8) is rendered.
	// Calculated from: title padding (1) + title text (1) + title padding (1) + 2 newlines = 4
	boardStartY = 4

	// boardStartXWithCoords is the column where the first piece starts when ShowCoords is true.
	// The rank label "8 " takes 2 characters.
	boardStartXWithCoords = 2

	// boardStartXNoCoords is the column where the first piece starts when ShowCoords is false.
	boardStartXNoCoords = 0

	// squareWidth is the width of each square in characters.
	// Each piece is followed by a space (except handled in rendering), so effectively 2 chars per square.
	squareWidth = 2
)

// squareFromMouse converts mouse coordinates to a chess square.
// Returns nil if the coordinates are outside the board.
//
// The calculation accounts for:
// - Board Y offset from title and spacing
// - Board X offset from rank labels (if ShowCoords is enabled)
// - Each square being 2 characters wide
// - Rank 8 at the top (y=0 relative to board), rank 1 at the bottom
func squareFromMouse(x, y int, config Config) *engine.Square {
	// Calculate board start X based on whether coordinates are shown
	boardStartX := boardStartXNoCoords
	if config.ShowCoords {
		boardStartX = boardStartXWithCoords
	}

	// Check if click is above or to the left of the board
	if x < boardStartX || y < boardStartY {
		return nil
	}

	// Calculate file (0-7) from X coordinate
	// Each square is squareWidth characters wide
	file := (x - boardStartX) / squareWidth

	// Calculate rank (0-7) from Y coordinate
	// Rank 8 (index 7) is at the top, rank 1 (index 0) is at the bottom
	rank := 7 - (y - boardStartY)

	// Validate file and rank are within bounds
	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return nil
	}

	// Create and return the square
	sq := engine.NewSquare(file, rank)
	return &sq
}

// handleMouseEvent processes mouse events during gameplay.
// It handles piece selection for interactive game modes (PvP and PvBot).
// Returns the updated model and any commands to execute.
func (m Model) handleMouseEvent(msg tea.MouseMsg) (Model, tea.Cmd) {
	// Only process left mouse button clicks
	if msg.Button != tea.MouseButtonLeft || msg.Action != tea.MouseActionPress {
		return m, nil
	}

	// Convert mouse coordinates to chess square
	sq := squareFromMouse(msg.X, msg.Y, m.config)
	if sq == nil {
		// Click was outside the board, ignore
		return m, nil
	}

	// For PvBot games, only allow selection when it's the human's turn
	if m.gameType == GameTypePvBot && m.board.ActiveColor != m.userColor {
		return m, nil
	}

	// Get the piece at the clicked square
	piece := m.board.PieceAt(*sq)

	// Check if the clicked square contains a piece belonging to the current player
	if !piece.IsEmpty() && piece.Color() == m.board.ActiveColor {
		// Select this piece (or change selection to a different own piece)
		m.selectedSquare = sq
	}
	// Note: Clicking on empty squares or opponent pieces does NOT clear selection
	// That behavior will be added in Slice 9 for move execution

	return m, nil
}
