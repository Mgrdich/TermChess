package engine

import (
	"errors"
	"fmt"
)

// Move represents a chess move from one square to another.
type Move struct {
	From      Square    // Source square
	To        Square    // Destination square
	Promotion PieceType // Promotion piece type (Empty if not a promotion)
}

// ParseMove parses a move from coordinate notation (e.g., "e2e4", "a7a8q").
// Format: from_file, from_rank, to_file, to_rank + optional promotion char.
// Promotion chars: q=Queen, r=Rook, b=Bishop, n=Knight (lowercase).
func ParseMove(s string) (Move, error) {
	if len(s) < 4 || len(s) > 5 {
		return Move{}, errors.New("invalid move format: expected 4-5 characters")
	}

	// Parse from square
	fromFile := int(s[0] - 'a')
	fromRank := int(s[1] - '1')
	if fromFile < 0 || fromFile > 7 || fromRank < 0 || fromRank > 7 {
		return Move{}, fmt.Errorf("invalid from square: %s", s[0:2])
	}

	// Parse to square
	toFile := int(s[2] - 'a')
	toRank := int(s[3] - '1')
	if toFile < 0 || toFile > 7 || toRank < 0 || toRank > 7 {
		return Move{}, fmt.Errorf("invalid to square: %s", s[2:4])
	}

	from := NewSquare(fromFile, fromRank)
	to := NewSquare(toFile, toRank)

	// Parse promotion if present
	var promotion PieceType = Empty
	if len(s) == 5 {
		switch s[4] {
		case 'q':
			promotion = Queen
		case 'r':
			promotion = Rook
		case 'b':
			promotion = Bishop
		case 'n':
			promotion = Knight
		default:
			return Move{}, fmt.Errorf("invalid promotion character: %c", s[4])
		}
	}

	return Move{
		From:      from,
		To:        to,
		Promotion: promotion,
	}, nil
}

// String returns the move in coordinate notation (e.g., "e2e4", "a7a8q").
func (m Move) String() string {
	s := m.From.String() + m.To.String()

	// Add promotion suffix if applicable
	if m.Promotion != Empty {
		switch m.Promotion {
		case Queen:
			s += "q"
		case Rook:
			s += "r"
		case Bishop:
			s += "b"
		case Knight:
			s += "n"
		}
	}

	return s
}

// generatePawnMoves generates all pseudo-legal pawn moves for the active color.
// This does not check for check or en passant (handled in later slices).
func (b *Board) generatePawnMoves() []Move {
	var moves []Move

	// Direction and starting rank depend on color
	var direction int
	var startRank int

	if b.ActiveColor == White {
		direction = 1 // White pawns move up (increasing rank)
		startRank = 1 // White pawns start on rank 2 (index 1)
	} else {
		direction = -1 // Black pawns move down (decreasing rank)
		startRank = 6  // Black pawns start on rank 7 (index 6)
	}

	// Iterate through all squares looking for pawns of the active color
	for sq := Square(0); sq < 64; sq++ {
		piece := b.Squares[sq]
		if piece.IsEmpty() || piece.Type() != Pawn || piece.Color() != b.ActiveColor {
			continue
		}

		file := sq.File()
		rank := sq.Rank()

		// One square forward
		forwardRank := rank + direction
		if forwardRank >= 0 && forwardRank <= 7 {
			forwardSq := NewSquare(file, forwardRank)
			if b.Squares[forwardSq].IsEmpty() {
				moves = append(moves, Move{From: sq, To: forwardSq})

				// Two squares forward from starting position
				if rank == startRank {
					twoForwardRank := rank + 2*direction
					twoForwardSq := NewSquare(file, twoForwardRank)
					if b.Squares[twoForwardSq].IsEmpty() {
						moves = append(moves, Move{From: sq, To: twoForwardSq})
					}
				}
			}
		}

		// Diagonal captures
		for _, fileOffset := range []int{-1, 1} {
			captureFile := file + fileOffset
			captureRank := rank + direction

			if captureFile >= 0 && captureFile <= 7 && captureRank >= 0 && captureRank <= 7 {
				captureSq := NewSquare(captureFile, captureRank)
				targetPiece := b.Squares[captureSq]

				// Can capture if there's an enemy piece
				if !targetPiece.IsEmpty() && targetPiece.Color() != b.ActiveColor {
					moves = append(moves, Move{From: sq, To: captureSq})
				}
			}
		}
	}

	return moves
}

// generateKnightMoves generates all pseudo-legal knight moves for the active color.
// Knights move in an L-shape: 2 squares in one direction, 1 square perpendicular.
func (b *Board) generateKnightMoves() []Move {
	var moves []Move

	// Knight move offsets: (file delta, rank delta)
	offsets := [][2]int{
		{+2, +1}, {+2, -1}, {-2, +1}, {-2, -1},
		{+1, +2}, {+1, -2}, {-1, +2}, {-1, -2},
	}

	// Iterate through all squares looking for knights of the active color
	for sq := Square(0); sq < 64; sq++ {
		piece := b.Squares[sq]
		if piece.IsEmpty() || piece.Type() != Knight || piece.Color() != b.ActiveColor {
			continue
		}

		file := sq.File()
		rank := sq.Rank()

		// Try each knight move offset
		for _, offset := range offsets {
			newFile := file + offset[0]
			newRank := rank + offset[1]

			// Check bounds
			if newFile < 0 || newFile > 7 || newRank < 0 || newRank > 7 {
				continue
			}

			targetSq := NewSquare(newFile, newRank)
			targetPiece := b.Squares[targetSq]

			// Can move to empty square or capture enemy piece
			if targetPiece.IsEmpty() || targetPiece.Color() != b.ActiveColor {
				moves = append(moves, Move{From: sq, To: targetSq})
			}
		}
	}

	return moves
}

// generateSlidingMoves generates all pseudo-legal moves for sliding pieces (bishop, rook, queen).
// It takes the piece type to look for and the directions to slide in.
func (b *Board) generateSlidingMoves(pieceType PieceType, directions [][2]int) []Move {
	var moves []Move

	// Iterate through all squares looking for pieces of the specified type and active color
	for sq := Square(0); sq < 64; sq++ {
		piece := b.Squares[sq]
		if piece.IsEmpty() || piece.Type() != pieceType || piece.Color() != b.ActiveColor {
			continue
		}

		file := sq.File()
		rank := sq.Rank()

		// Try each direction
		for _, dir := range directions {
			// Slide in this direction until blocked
			for dist := 1; dist <= 7; dist++ {
				newFile := file + dir[0]*dist
				newRank := rank + dir[1]*dist

				// Check bounds
				if newFile < 0 || newFile > 7 || newRank < 0 || newRank > 7 {
					break
				}

				targetSq := NewSquare(newFile, newRank)
				targetPiece := b.Squares[targetSq]

				if targetPiece.IsEmpty() {
					// Empty square - can move here and continue sliding
					moves = append(moves, Move{From: sq, To: targetSq})
				} else if targetPiece.Color() != b.ActiveColor {
					// Enemy piece - can capture but then stop
					moves = append(moves, Move{From: sq, To: targetSq})
					break
				} else {
					// Own piece - stop before it
					break
				}
			}
		}
	}

	return moves
}

// generateBishopMoves generates all pseudo-legal bishop moves for the active color.
// Bishops move diagonally any number of squares.
func (b *Board) generateBishopMoves() []Move {
	// Diagonal directions: (+1,+1), (+1,-1), (-1,+1), (-1,-1)
	directions := [][2]int{
		{+1, +1}, {+1, -1}, {-1, +1}, {-1, -1},
	}
	return b.generateSlidingMoves(Bishop, directions)
}

// generateRookMoves generates all pseudo-legal rook moves for the active color.
// Rooks move orthogonally (horizontal/vertical) any number of squares.
func (b *Board) generateRookMoves() []Move {
	// Orthogonal directions: (+1,0), (-1,0), (0,+1), (0,-1)
	directions := [][2]int{
		{+1, 0}, {-1, 0}, {0, +1}, {0, -1},
	}
	return b.generateSlidingMoves(Rook, directions)
}

// generateQueenMoves generates all pseudo-legal queen moves for the active color.
// Queens combine bishop and rook movement (all 8 directions).
func (b *Board) generateQueenMoves() []Move {
	// All 8 directions (diagonal + orthogonal)
	directions := [][2]int{
		{+1, +1}, {+1, -1}, {-1, +1}, {-1, -1}, // diagonal
		{+1, 0}, {-1, 0}, {0, +1}, {0, -1}, // orthogonal
	}
	return b.generateSlidingMoves(Queen, directions)
}

// generateKingMoves generates all pseudo-legal king moves for the active color.
// Kings move one square in any direction. This also includes castling moves
// when the conditions are met (rights, empty squares, not in/through/into check).
func (b *Board) generateKingMoves() []Move {
	var moves []Move

	// King move offsets: all 8 adjacent squares
	offsets := [][2]int{
		{+1, +1}, {+1, -1}, {-1, +1}, {-1, -1}, // diagonal
		{+1, 0}, {-1, 0}, {0, +1}, {0, -1}, // orthogonal
	}

	// Iterate through all squares looking for kings of the active color
	for sq := Square(0); sq < 64; sq++ {
		piece := b.Squares[sq]
		if piece.IsEmpty() || piece.Type() != King || piece.Color() != b.ActiveColor {
			continue
		}

		file := sq.File()
		rank := sq.Rank()

		// Try each king move offset
		for _, offset := range offsets {
			newFile := file + offset[0]
			newRank := rank + offset[1]

			// Check bounds
			if newFile < 0 || newFile > 7 || newRank < 0 || newRank > 7 {
				continue
			}

			targetSq := NewSquare(newFile, newRank)
			targetPiece := b.Squares[targetSq]

			// Can move to empty square or capture enemy piece
			if targetPiece.IsEmpty() || targetPiece.Color() != b.ActiveColor {
				moves = append(moves, Move{From: sq, To: targetSq})
			}
		}

		// Generate castling moves
		moves = append(moves, b.generateCastlingMoves(sq)...)
	}

	return moves
}

// generateCastlingMoves generates castling moves for the king at the given square.
// Castling requires:
// 1. King has castling rights (king and relevant rook haven't moved)
// 2. No pieces between king and rook
// 3. King is NOT currently in check
// 4. King does not pass through a square that is attacked
// 5. King does not land on a square that is attacked
func (b *Board) generateCastlingMoves(kingSq Square) []Move {
	var moves []Move

	// Determine opponent color for attack checks
	opponentColor := Black
	if b.ActiveColor == Black {
		opponentColor = White
	}

	// King must not be in check to castle
	if b.IsSquareAttacked(kingSq, opponentColor) {
		return moves
	}

	if b.ActiveColor == White {
		// White king should be on e1 (square 4) for castling
		if kingSq != NewSquare(4, 0) {
			return moves
		}

		// White kingside castling (O-O): King e1 -> g1, Rook h1 -> f1
		if b.CastlingRights&CastleWhiteKing != 0 {
			// Check squares f1 and g1 are empty
			f1 := NewSquare(5, 0)
			g1 := NewSquare(6, 0)

			if b.Squares[f1].IsEmpty() && b.Squares[g1].IsEmpty() {
				// Check that king doesn't pass through or land on attacked square
				if !b.IsSquareAttacked(f1, opponentColor) && !b.IsSquareAttacked(g1, opponentColor) {
					moves = append(moves, Move{From: kingSq, To: g1})
				}
			}
		}

		// White queenside castling (O-O-O): King e1 -> c1, Rook a1 -> d1
		if b.CastlingRights&CastleWhiteQueen != 0 {
			// Check squares b1, c1, and d1 are empty
			b1 := NewSquare(1, 0)
			c1 := NewSquare(2, 0)
			d1 := NewSquare(3, 0)

			if b.Squares[b1].IsEmpty() && b.Squares[c1].IsEmpty() && b.Squares[d1].IsEmpty() {
				// Check that king doesn't pass through or land on attacked square
				// (only c1 and d1 matter for the king's path, b1 just needs to be empty for the rook)
				if !b.IsSquareAttacked(c1, opponentColor) && !b.IsSquareAttacked(d1, opponentColor) {
					moves = append(moves, Move{From: kingSq, To: c1})
				}
			}
		}
	} else {
		// Black king should be on e8 (square 60) for castling
		if kingSq != NewSquare(4, 7) {
			return moves
		}

		// Black kingside castling (O-O): King e8 -> g8, Rook h8 -> f8
		if b.CastlingRights&CastleBlackKing != 0 {
			// Check squares f8 and g8 are empty
			f8 := NewSquare(5, 7)
			g8 := NewSquare(6, 7)

			if b.Squares[f8].IsEmpty() && b.Squares[g8].IsEmpty() {
				// Check that king doesn't pass through or land on attacked square
				if !b.IsSquareAttacked(f8, opponentColor) && !b.IsSquareAttacked(g8, opponentColor) {
					moves = append(moves, Move{From: kingSq, To: g8})
				}
			}
		}

		// Black queenside castling (O-O-O): King e8 -> c8, Rook a8 -> d8
		if b.CastlingRights&CastleBlackQueen != 0 {
			// Check squares b8, c8, and d8 are empty
			b8 := NewSquare(1, 7)
			c8 := NewSquare(2, 7)
			d8 := NewSquare(3, 7)

			if b.Squares[b8].IsEmpty() && b.Squares[c8].IsEmpty() && b.Squares[d8].IsEmpty() {
				// Check that king doesn't pass through or land on attacked square
				// (only c8 and d8 matter for the king's path, b8 just needs to be empty for the rook)
				if !b.IsSquareAttacked(c8, opponentColor) && !b.IsSquareAttacked(d8, opponentColor) {
					moves = append(moves, Move{From: kingSq, To: c8})
				}
			}
		}
	}

	return moves
}

// PseudoLegalMoves generates all pseudo-legal moves for the active color.
// Pseudo-legal moves are moves that follow piece movement rules but may leave
// the king in check. Filtering for check is done in a later slice.
func (b *Board) PseudoLegalMoves() []Move {
	var moves []Move

	// Generate moves for each piece type
	moves = append(moves, b.generatePawnMoves()...)
	moves = append(moves, b.generateKnightMoves()...)
	moves = append(moves, b.generateBishopMoves()...)
	moves = append(moves, b.generateRookMoves()...)
	moves = append(moves, b.generateQueenMoves()...)
	moves = append(moves, b.generateKingMoves()...)

	return moves
}

// LegalMoves generates all legal moves for the active color.
// A legal move is a pseudo-legal move that does not leave the king in check.
// This is done by filtering pseudo-legal moves: for each move, we make it on
// a copy of the board and verify the king is not in check afterwards.
func (b *Board) LegalMoves() []Move {
	pseudoLegalMoves := b.PseudoLegalMoves()
	var legalMoves []Move

	// Remember which color is moving (before applyMove switches it)
	movingColor := b.ActiveColor

	for _, move := range pseudoLegalMoves {
		// Create a copy of the board to test the move
		boardCopy := b.Copy()

		// Apply the move on the copy using the internal method (skips legality check)
		// This avoids infinite recursion since MakeMove calls IsLegalMove which calls LegalMoves
		boardCopy.applyMove(move)

		// After applyMove, ActiveColor has switched to the opponent.
		// We need to check if the king of the color that JUST moved is in check.
		// The opponent (now active) would be attacking, so we check if the
		// moving color's king is attacked by the new active color.
		kingSquare := NoSquare
		for sq := Square(0); sq < 64; sq++ {
			piece := boardCopy.Squares[sq]
			if piece.Type() == King && piece.Color() == movingColor {
				kingSquare = sq
				break
			}
		}

		// If no king found (shouldn't happen), skip this move
		if kingSquare == NoSquare {
			continue
		}

		// Check if the king is attacked by the opponent (who is now the active color)
		if !boardCopy.IsSquareAttacked(kingSquare, boardCopy.ActiveColor) {
			// King is not in check after this move - it's legal
			legalMoves = append(legalMoves, move)
		}
	}

	return legalMoves
}

// IsLegalMove checks if a specific move is legal for the current position.
// Returns true if the move is in the list of legal moves, false otherwise.
// This is a convenience method for validating a single move without generating
// all legal moves manually.
func (b *Board) IsLegalMove(m Move) bool {
	legalMoves := b.LegalMoves()
	for _, legalMove := range legalMoves {
		if legalMove.From == m.From && legalMove.To == m.To && legalMove.Promotion == m.Promotion {
			return true
		}
	}
	return false
}
