package ui

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// ParseSAN converts Standard Algebraic Notation to a Move.
// For Slice 5, only handles pawn moves:
// - Simple: "e4", "d5"
// - Captures: "exd5", "axb3"
// - Promotions: "e8=Q", "a1=N"
// - Combined: "exd8=Q"
func ParseSAN(b *engine.Board, san string) (engine.Move, error) {
	if san == "" {
		return engine.Move{}, fmt.Errorf("empty move notation")
	}

	// Strip check/checkmate symbols (+, #) from the end
	san = strings.TrimSuffix(san, "+")
	san = strings.TrimSuffix(san, "#")

	if san == "" {
		return engine.Move{}, fmt.Errorf("invalid move notation")
	}

	// Check if it's a piece move (starts with uppercase letter for piece type)
	// K, Q, R, B, N indicate piece moves - not supported in Slice 5
	firstChar := rune(san[0])
	if unicode.IsUpper(firstChar) && (firstChar == 'K' || firstChar == 'Q' ||
		firstChar == 'R' || firstChar == 'B' || firstChar == 'N') {
		return engine.Move{}, fmt.Errorf("piece moves not yet supported (slice 6)")
	}

	// Check for castling notation
	if san == "O-O" || san == "O-O-O" || san == "0-0" || san == "0-0-0" {
		return engine.Move{}, fmt.Errorf("castling not yet supported (slice 6)")
	}

	// Must be a pawn move - parse it
	return parsePawnMove(b, san)
}

// parsePawnMove parses a pawn move in SAN notation.
// Formats:
// - "e4" - simple pawn move
// - "e8=Q" - pawn move with promotion
// - "exd5" - pawn capture
// - "exd8=Q" - pawn capture with promotion
func parsePawnMove(b *engine.Board, san string) (engine.Move, error) {
	// Parse promotion suffix first (=Q, =R, =B, =N)
	var promotion engine.PieceType = engine.Empty
	var moveStr = san

	if strings.Contains(san, "=") {
		parts := strings.Split(san, "=")
		if len(parts) != 2 {
			return engine.Move{}, fmt.Errorf("invalid promotion format: %s", san)
		}
		moveStr = parts[0]

		var err error
		promotion, err = parsePromotion(parts[1])
		if err != nil {
			return engine.Move{}, err
		}
	}

	// Parse capture indicator (x)
	isCapture := strings.Contains(moveStr, "x")
	var sourceFile int = -1
	var destSquare engine.Square

	if isCapture {
		// Format: "exd5" or "axb3"
		parts := strings.Split(moveStr, "x")
		if len(parts) != 2 {
			return engine.Move{}, fmt.Errorf("invalid capture format: %s", san)
		}

		// Parse source file (e.g., 'e' from "exd5")
		if len(parts[0]) != 1 {
			return engine.Move{}, fmt.Errorf("invalid source file in capture: %s", san)
		}

		var err error
		sourceFile, err = parseFile(rune(parts[0][0]))
		if err != nil {
			return engine.Move{}, fmt.Errorf("invalid source file: %v", err)
		}

		// Parse destination square (e.g., "d5" from "exd5")
		destSquare, err = parseSquare(parts[1])
		if err != nil {
			return engine.Move{}, fmt.Errorf("invalid destination square: %v", err)
		}
	} else {
		// Simple pawn move: "e4" or just the destination square
		var err error
		destSquare, err = parseSquare(moveStr)
		if err != nil {
			return engine.Move{}, fmt.Errorf("invalid destination square: %v", err)
		}
	}

	// Get all legal moves for the current player
	legalMoves := b.LegalMoves()

	// Filter for pawn moves to the destination square
	var candidates []engine.Move
	for _, move := range legalMoves {
		piece := b.PieceAt(move.From)

		// Must be a pawn
		if piece.Type() != engine.Pawn {
			continue
		}

		// Must move to the destination square
		if move.To != destSquare {
			continue
		}

		// If capture, must match source file
		if isCapture {
			if move.From.File() != sourceFile {
				continue
			}
		}

		// If promotion specified, must match
		if promotion != engine.Empty {
			if move.Promotion != promotion {
				continue
			}
		}

		candidates = append(candidates, move)
	}

	// Return the unique match or error
	if len(candidates) == 0 {
		return engine.Move{}, fmt.Errorf("no legal pawn move matches: %s", san)
	}

	if len(candidates) > 1 {
		return engine.Move{}, fmt.Errorf("ambiguous pawn move: %s (multiple candidates)", san)
	}

	return candidates[0], nil
}

// parseSquare converts algebraic notation like "e4" to a Square.
// File must be 'a'-'h', rank must be '1'-'8'.
func parseSquare(s string) (engine.Square, error) {
	if len(s) != 2 {
		return engine.NoSquare, fmt.Errorf("invalid square notation: %s (expected 2 characters)", s)
	}

	file := int(s[0] - 'a')
	rank := int(s[1] - '1')

	if file < 0 || file > 7 {
		return engine.NoSquare, fmt.Errorf("invalid file: %c (expected a-h)", s[0])
	}

	if rank < 0 || rank > 7 {
		return engine.NoSquare, fmt.Errorf("invalid rank: %c (expected 1-8)", s[1])
	}

	return engine.NewSquare(file, rank), nil
}

// parsePromotion converts a promotion character to a PieceType.
// Accepts: Q, R, B, N (uppercase or lowercase).
func parsePromotion(s string) (engine.PieceType, error) {
	if len(s) != 1 {
		return engine.Empty, fmt.Errorf("invalid promotion piece: %s", s)
	}

	switch unicode.ToUpper(rune(s[0])) {
	case 'Q':
		return engine.Queen, nil
	case 'R':
		return engine.Rook, nil
	case 'B':
		return engine.Bishop, nil
	case 'N':
		return engine.Knight, nil
	default:
		return engine.Empty, fmt.Errorf("invalid promotion piece: %s (expected Q, R, B, or N)", s)
	}
}

// parseFile converts a file character ('a'-'h') to a file index (0-7).
func parseFile(r rune) (int, error) {
	file := int(r - 'a')
	if file < 0 || file > 7 {
		return -1, fmt.Errorf("invalid file: %c (expected a-h)", r)
	}
	return file, nil
}
