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
		result.WriteString("  ") // Indent to align with rank numbers
		result.WriteString("a b c d e f g h")
	}

	return result.String()
}

// pieceSymbol returns the symbol to use for the given piece.
// For ASCII mode, returns uppercase for white pieces, lowercase for black pieces.
// For Unicode mode, returns Unicode chess symbols.
func (r *BoardRenderer) pieceSymbol(p engine.Piece) string {
	var symbol string

	if r.config.UseUnicode {
		// Unicode symbols
		symbol = r.unicodeSymbol(p)
	} else {
		// ASCII symbols
		symbol = r.asciiSymbol(p)
	}

	// Apply colors if enabled (but not for empty squares)
	if r.config.UseColors && !p.IsEmpty() {
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
// Uses standard Unicode chess piece symbols (U+2654 through U+265F).
// Empty squares are represented by a middle dot (·).
func (r *BoardRenderer) unicodeSymbol(p engine.Piece) string {
	pieceType := p.Type()

	if pieceType == engine.Empty {
		return "·" // Middle dot for empty squares
	}

	color := p.Color()

	// White pieces: U+2654 to U+2659
	if color == engine.White {
		switch pieceType {
		case engine.King:
			return "♔" // U+2654
		case engine.Queen:
			return "♕" // U+2655
		case engine.Rook:
			return "♖" // U+2656
		case engine.Bishop:
			return "♗" // U+2657
		case engine.Knight:
			return "♘" // U+2658
		case engine.Pawn:
			return "♙" // U+2659
		}
	} else {
		// Black pieces: U+265A to U+265F
		switch pieceType {
		case engine.King:
			return "♚" // U+265A
		case engine.Queen:
			return "♛" // U+265B
		case engine.Rook:
			return "♜" // U+265C
		case engine.Bishop:
			return "♝" // U+265D
		case engine.Knight:
			return "♞" // U+265E
		case engine.Pawn:
			return "♟" // U+265F
		}
	}

	// Fallback for unknown pieces
	return "?"
}

// colorSymbol applies color styling to a piece symbol using lipgloss.
// White pieces are rendered in bright/white color (terminal color 15).
// Black pieces are rendered in dim/gray color (terminal color 8).
// Using terminal color codes (0-15) provides better compatibility across different terminals.
func (r *BoardRenderer) colorSymbol(symbol string, p engine.Piece) string {
	if p.Color() == engine.White {
		// White pieces: bright white (terminal color 15) with bold
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
		return style.Render(symbol)
	} else {
		// Black pieces: gray (terminal color 8)
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		return style.Render(symbol)
	}
}
