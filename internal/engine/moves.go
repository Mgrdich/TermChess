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
		direction = 1  // White pawns move up (increasing rank)
		startRank = 1  // White pawns start on rank 2 (index 1)
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
