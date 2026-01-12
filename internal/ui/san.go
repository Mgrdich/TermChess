package ui

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// ParseSAN converts Standard Algebraic Notation to a Move.
// Supports:
// - Pawn moves: "e4", "d5", "exd5", "e8=Q"
// - Piece moves: "Nf3", "Bc4", "Qh5", "Kf1"
// - Disambiguation: "Nbd2" (file), "N1f3" (rank), "Nb1d2" (both)
// - Captures: "Bxc5", "Nxe5", "Nbxd4" (with disambiguation)
// - Castling: "O-O", "O-O-O"
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

	// Check for castling notation
	if san == "O-O" || san == "0-0" {
		return parseCastling(b, true) // kingside
	}
	if san == "O-O-O" || san == "0-0-0" {
		return parseCastling(b, false) // queenside
	}

	// Check if it's a piece move (starts with uppercase letter for piece type)
	// K, Q, R, B, N indicate piece moves
	firstChar := rune(san[0])
	if unicode.IsUpper(firstChar) && (firstChar == 'K' || firstChar == 'Q' ||
		firstChar == 'R' || firstChar == 'B' || firstChar == 'N') {
		return parsePieceMove(b, san)
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

// parsePieceMove parses a piece move in SAN notation.
// Formats:
// - "Nf3" - knight to f3
// - "Bc4" - bishop to c4
// - "Qh5" - queen to h5
// - "Bxc5" - bishop captures on c5
// - "Nxe5" - knight captures on e5
// - "Nbd2" - knight from b-file to d2 (file disambiguation)
// - "N1d2" - knight from rank 1 to d2 (rank disambiguation)
// - "Nb1d2" - knight from b1 to d2 (file+rank disambiguation)
func parsePieceMove(b *engine.Board, san string) (engine.Move, error) {
	if len(san) < 2 {
		return engine.Move{}, fmt.Errorf("invalid piece move format: %s", san)
	}

	// Parse piece type (first character)
	pieceType, err := parsePieceType(rune(san[0]))
	if err != nil {
		return engine.Move{}, err
	}

	// Remove piece type from the string
	moveStr := san[1:]

	// Parse disambiguation (optional file and/or rank)
	fromFile := -1
	fromRank := -1

	// First, check for and remove the capture marker 'x'
	// We need to do this early to know where the destination square starts
	captureIdx := strings.Index(moveStr, "x")
	var disambiguationPart string
	var remainingPart string

	if captureIdx >= 0 {
		// There's a capture marker
		disambiguationPart = moveStr[:captureIdx]
		remainingPart = moveStr[captureIdx+1:] // Skip the 'x'
	} else {
		// No capture marker, but we need to figure out what's disambiguation vs destination
		// The destination square is always the last 2 characters
		if len(moveStr) > 2 {
			disambiguationPart = moveStr[:len(moveStr)-2]
			remainingPart = moveStr[len(moveStr)-2:]
		} else {
			disambiguationPart = ""
			remainingPart = moveStr
		}
	}

	// Parse the disambiguation part
	for i := 0; i < len(disambiguationPart); i++ {
		ch := disambiguationPart[i]
		if ch >= 'a' && ch <= 'h' {
			fromFile = int(ch - 'a')
		} else if ch >= '1' && ch <= '8' {
			fromRank = int(ch - '1')
		}
	}

	moveStr = remainingPart

	// The remaining string should be the destination square (e.g., "f3")
	if len(moveStr) != 2 {
		return engine.Move{}, fmt.Errorf("invalid piece move format: %s", san)
	}

	destSquare, err := parseSquare(moveStr)
	if err != nil {
		return engine.Move{}, fmt.Errorf("invalid destination square: %v", err)
	}

	// Get all legal moves
	legalMoves := b.LegalMoves()

	// Filter for moves that match:
	// - Piece type
	// - Destination square
	// - File disambiguation (if specified)
	// - Rank disambiguation (if specified)
	var candidates []engine.Move
	for _, move := range legalMoves {
		piece := b.PieceAt(move.From)

		// Must be the correct piece type
		if piece.Type() != pieceType {
			continue
		}

		// Must move to the destination square
		if move.To != destSquare {
			continue
		}

		// Check file disambiguation
		if fromFile >= 0 && move.From.File() != fromFile {
			continue
		}

		// Check rank disambiguation
		if fromRank >= 0 && move.From.Rank() != fromRank {
			continue
		}

		candidates = append(candidates, move)
	}

	// Return the unique match or error
	if len(candidates) == 0 {
		return engine.Move{}, fmt.Errorf("no legal move matches: %s", san)
	}

	if len(candidates) > 1 {
		return engine.Move{}, fmt.Errorf("move is still ambiguous: %s (multiple candidates)", san)
	}

	return candidates[0], nil
}

// parseCastling parses a castling move.
// kingside: true for O-O (kingside), false for O-O-O (queenside)
func parseCastling(b *engine.Board, kingside bool) (engine.Move, error) {
	// Determine the king's starting square based on the active color
	var kingFrom, kingTo engine.Square

	if b.ActiveColor == engine.White {
		kingFrom = engine.NewSquare(4, 0) // e1
		if kingside {
			kingTo = engine.NewSquare(6, 0) // g1
		} else {
			kingTo = engine.NewSquare(2, 0) // c1
		}
	} else {
		kingFrom = engine.NewSquare(4, 7) // e8
		if kingside {
			kingTo = engine.NewSquare(6, 7) // g8
		} else {
			kingTo = engine.NewSquare(2, 7) // c8
		}
	}

	// Create the castling move
	castleMove := engine.Move{From: kingFrom, To: kingTo}

	// Verify this is a legal move
	legalMoves := b.LegalMoves()
	for _, move := range legalMoves {
		if move.From == castleMove.From && move.To == castleMove.To {
			return move, nil
		}
	}

	// Castling is not legal
	if kingside {
		return engine.Move{}, fmt.Errorf("kingside castling is not legal")
	}
	return engine.Move{}, fmt.Errorf("queenside castling is not legal")
}

// parsePieceType converts a piece character to a PieceType.
// Accepts: K, Q, R, B, N (uppercase).
func parsePieceType(r rune) (engine.PieceType, error) {
	switch r {
	case 'K':
		return engine.King, nil
	case 'Q':
		return engine.Queen, nil
	case 'R':
		return engine.Rook, nil
	case 'B':
		return engine.Bishop, nil
	case 'N':
		return engine.Knight, nil
	default:
		return engine.Empty, fmt.Errorf("invalid piece type: %c (expected K, Q, R, B, or N)", r)
	}
}

// FormatSAN converts a Move to Standard Algebraic Notation (SAN).
// Takes the board state BEFORE the move and the move to format.
// Returns the SAN string (e.g., "e4", "Nf3", "Bxc5", "O-O", "e8=Q+").
//
// Algorithm:
// 1. Check for castling (king moves 2 squares) -> "O-O" or "O-O-O"
// 2. Get piece type (Pawn, Knight, Bishop, Rook, Queen, King)
// 3. Check if it's a capture (destination square has enemy piece or en passant)
// 4. For disambiguation: find all legal moves by same piece type to same destination
// 5. Build string: Piece + disambiguation + capture marker + destination + promotion + check
func FormatSAN(board *engine.Board, move engine.Move) string {
	piece := board.PieceAt(move.From)
	if piece.IsEmpty() {
		return move.String() // Fallback to coordinate notation
	}

	// Check for castling notation
	if piece.Type() == engine.King {
		fileDiff := move.To.File() - move.From.File()
		if fileDiff == 2 {
			return "O-O" // Kingside castling
		} else if fileDiff == -2 {
			return "O-O-O" // Queenside castling
		}
	}

	var result strings.Builder

	// Add piece letter (empty for pawns)
	pieceType := piece.Type()
	if pieceType != engine.Pawn {
		result.WriteRune(pieceTypeToRune(pieceType))
	}

	// Check if this is a capture
	targetPiece := board.PieceAt(move.To)
	isCapture := !targetPiece.IsEmpty()

	// Check for en passant capture
	if pieceType == engine.Pawn && board.EnPassantSq >= 0 && move.To == engine.Square(board.EnPassantSq) {
		isCapture = true
	}

	// Add disambiguation for non-pawn pieces
	if pieceType != engine.Pawn {
		disambiguation := getDisambiguation(board, move)
		result.WriteString(disambiguation)
	} else if isCapture {
		// For pawn captures, always add the source file
		result.WriteRune(rune('a' + move.From.File()))
	}

	// Add capture marker
	if isCapture {
		result.WriteRune('x')
	}

	// Add destination square
	result.WriteString(move.To.String())

	// Add promotion notation
	if move.Promotion != engine.Empty {
		result.WriteRune('=')
		result.WriteRune(pieceTypeToRune(move.Promotion))
	}

	// Check for check or checkmate by making the move on a copy
	boardCopy := board.Copy()
	boardCopy.MakeMove(move)

	if boardCopy.InCheck() {
		// Check if it's checkmate
		if len(boardCopy.LegalMoves()) == 0 {
			result.WriteRune('#')
		} else {
			result.WriteRune('+')
		}
	}

	return result.String()
}

// pieceTypeToRune converts a PieceType to its SAN character representation.
func pieceTypeToRune(pt engine.PieceType) rune {
	switch pt {
	case engine.King:
		return 'K'
	case engine.Queen:
		return 'Q'
	case engine.Rook:
		return 'R'
	case engine.Bishop:
		return 'B'
	case engine.Knight:
		return 'N'
	default:
		return '?'
	}
}

// getDisambiguation returns the disambiguation string needed for a piece move.
// This is necessary when multiple pieces of the same type can move to the same square.
// Returns:
// - "" if no disambiguation needed
// - "a" (file) if file alone is sufficient to disambiguate
// - "1" (rank) if rank alone is sufficient to disambiguate
// - "a1" (both) if both file and rank are needed to disambiguate
func getDisambiguation(board *engine.Board, move engine.Move) string {
	piece := board.PieceAt(move.From)
	pieceType := piece.Type()

	// Find all legal moves by the same piece type to the same destination
	legalMoves := board.LegalMoves()
	var candidates []engine.Move

	for _, m := range legalMoves {
		if m.To == move.To && m.From != move.From {
			candidatePiece := board.PieceAt(m.From)
			if candidatePiece.Type() == pieceType {
				candidates = append(candidates, m)
			}
		}
	}

	// No disambiguation needed if this is the only piece that can move there
	if len(candidates) == 0 {
		return ""
	}

	fromFile := move.From.File()
	fromRank := move.From.Rank()

	// Check if file alone is sufficient (no other candidate on same file)
	fileUnique := true
	for _, m := range candidates {
		if m.From.File() == fromFile {
			fileUnique = false
			break
		}
	}

	if fileUnique {
		return string(rune('a' + fromFile))
	}

	// Check if rank alone is sufficient (no other candidate on same rank)
	rankUnique := true
	for _, m := range candidates {
		if m.From.Rank() == fromRank {
			rankUnique = false
			break
		}
	}

	if rankUnique {
		return string(rune('1' + fromRank))
	}

	// Need both file and rank
	return string(rune('a'+fromFile)) + string(rune('1'+fromRank))
}
