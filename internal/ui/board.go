package ui

import (
	"fmt"
	"strings"

	"github.com/Mgrdich/TermChess/internal/engine"
	"github.com/charmbracelet/lipgloss"
)

// BoardRenderer is responsible for rendering the chess board to the terminal.
// It uses the Config to determine how to display pieces and coordinates.
type BoardRenderer struct {
	config Config
}

// NewBoardRenderer creates a new BoardRenderer with the given configuration.
func NewBoardRenderer(config Config) *BoardRenderer {
	return &BoardRenderer{
		config: config,
	}
}

// Render renders the chess board as a string.
// The board is displayed from White's perspective (rank 8 at top, rank 1 at bottom).
// If the board is nil, returns an error message.
func (r *BoardRenderer) Render(b *engine.Board) string {
	if b == nil {
		return "No board available"
	}

	var result strings.Builder

	// Render each rank from 8 down to 1 (from White's perspective)
	for rank := 7; rank >= 0; rank-- {
		// Show rank number if coordinates are enabled
		if r.config.ShowCoords {
			result.WriteString(fmt.Sprintf("%d ", rank+1))
		}

		// Render pieces for this rank (files a-h, which are 0-7)
		for file := 0; file < 8; file++ {
			sq := engine.NewSquare(file, rank)
			piece := b.PieceAt(sq)
			symbol := r.pieceSymbol(piece)

			// Add spacing between pieces for readability
			if file > 0 {
				result.WriteString(" ")
			}

			result.WriteString(symbol)
		}

		result.WriteString("\n")
	}

	// Show file labels at the bottom if coordinates are enabled
	if r.config.ShowCoords {
		if r.config.ShowCoords {
			result.WriteString("  ") // Indent to align with rank numbers
		}
		result.WriteString("a b c d e f g h")
	}

	return result.String()
}

// pieceSymbol returns the symbol to use for the given piece.
// For ASCII mode, returns uppercase for white pieces, lowercase for black pieces.
// For Unicode mode, returns Unicode chess symbols (implemented in Slice 9).
func (r *BoardRenderer) pieceSymbol(p engine.Piece) string {
	if p.IsEmpty() {
		return "."
	}

	var symbol string

	if r.config.UseUnicode {
		// Unicode symbols - to be implemented in Slice 9
		symbol = r.unicodeSymbol(p)
	} else {
		// ASCII symbols
		symbol = r.asciiSymbol(p)
	}

	// Apply colors if enabled
	if r.config.UseColors {
		return r.colorSymbol(symbol, p)
	}

	return symbol
}

// asciiSymbol returns the ASCII character for the given piece.
// White pieces are uppercase (P, N, B, R, Q, K).
// Black pieces are lowercase (p, n, b, r, q, k).
func (r *BoardRenderer) asciiSymbol(p engine.Piece) string {
	pieceType := p.Type()
	color := p.Color()

	var ch byte
	switch pieceType {
	case engine.Pawn:
		ch = 'P'
	case engine.Knight:
		ch = 'N'
	case engine.Bishop:
		ch = 'B'
	case engine.Rook:
		ch = 'R'
	case engine.Queen:
		ch = 'Q'
	case engine.King:
		ch = 'K'
	default:
		return "."
	}

	// Convert to lowercase for black pieces
	if color == engine.Black {
		ch = ch - 'A' + 'a'
	}

	return string(ch)
}

// unicodeSymbol returns the Unicode chess symbol for the given piece.
// This will be implemented in Slice 9. For now, it falls back to ASCII.
func (r *BoardRenderer) unicodeSymbol(p engine.Piece) string {
	// TODO: Implement Unicode symbols in Slice 9
	// For now, fall back to ASCII
	return r.asciiSymbol(p)
}

// colorSymbol applies color styling to a piece symbol using lipgloss.
// White pieces are rendered in bright/white color.
// Black pieces are rendered in dim/gray color.
func (r *BoardRenderer) colorSymbol(symbol string, p engine.Piece) string {
	if p.Color() == engine.White {
		// White pieces: bright white
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
		return style.Render(symbol)
	} else {
		// Black pieces: dim gray
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
		return style.Render(symbol)
	}
}
