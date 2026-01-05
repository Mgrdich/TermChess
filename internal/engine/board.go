package engine

import "fmt"

// Board represents the complete state of a chess game.
type Board struct {
	// Squares holds all 64 squares of the board.
	// Indexed as rank * 8 + file, where a1=0, b1=1, ..., h8=63.
	Squares [64]Piece

	// ActiveColor is the color of the player to move.
	ActiveColor Color

	// CastlingRights encodes available castling options.
	// Bit 0: White kingside (K)
	// Bit 1: White queenside (Q)
	// Bit 2: Black kingside (k)
	// Bit 3: Black queenside (q)
	CastlingRights uint8

	// EnPassantSq is the en passant target square, or -1 if none.
	EnPassantSq int8

	// HalfMoveClock counts half-moves since last pawn move or capture.
	// Used for the fifty-move rule.
	HalfMoveClock uint8

	// FullMoveNum is the current full move number, starting at 1.
	FullMoveNum uint16

	// Hash is the Zobrist hash of the current position.
	Hash uint64

	// History stores Zobrist hashes of previous positions.
	// Used for threefold repetition detection.
	History []uint64
}

// Castling rights bit masks.
const (
	CastleWhiteKing  uint8 = 1 << 0 // K
	CastleWhiteQueen uint8 = 1 << 1 // Q
	CastleBlackKing  uint8 = 1 << 2 // k
	CastleBlackQueen uint8 = 1 << 3 // q
	CastleAll        uint8 = CastleWhiteKing | CastleWhiteQueen | CastleBlackKing | CastleBlackQueen
)

// NewBoard creates a new chess board with the standard starting position.
// White pieces on ranks 1-2, Black pieces on ranks 7-8.
// White is to move, all castling rights are available,
// no en passant square, half-move clock is 0, and full move number is 1.
func NewBoard() *Board {
	b := &Board{
		Squares:        [64]Piece{}, // Will be populated below
		ActiveColor:    White,
		CastlingRights: CastleAll,
		EnPassantSq:    -1,
		HalfMoveClock:  0,
		FullMoveNum:    1,
		Hash:           0,
		History:        []uint64{},
	}

	// Back rank piece order: Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook
	backRank := []PieceType{Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook}

	// Place White pieces on rank 1 (index 0-7)
	for file := 0; file < 8; file++ {
		b.Squares[file] = NewPiece(White, backRank[file])
	}

	// Place White pawns on rank 2 (index 8-15)
	for file := 0; file < 8; file++ {
		b.Squares[8+file] = NewPiece(White, Pawn)
	}

	// Place Black pawns on rank 7 (index 48-55)
	for file := 0; file < 8; file++ {
		b.Squares[48+file] = NewPiece(Black, Pawn)
	}

	// Place Black pieces on rank 8 (index 56-63)
	for file := 0; file < 8; file++ {
		b.Squares[56+file] = NewPiece(Black, backRank[file])
	}

	// Compute the initial Zobrist hash and add it to history
	b.Hash = b.ComputeHash()
	b.History = append(b.History, b.Hash)

	return b
}

// PieceAt returns the piece at the given square.
func (b *Board) PieceAt(sq Square) Piece {
	if !sq.IsValid() {
		return Piece(Empty)
	}
	return b.Squares[sq]
}

// Copy returns a deep copy of the board.
func (b *Board) Copy() *Board {
	newBoard := &Board{
		Squares:        b.Squares, // Array is copied by value
		ActiveColor:    b.ActiveColor,
		CastlingRights: b.CastlingRights,
		EnPassantSq:    b.EnPassantSq,
		HalfMoveClock:  b.HalfMoveClock,
		FullMoveNum:    b.FullMoveNum,
		Hash:           b.Hash,
		History:        make([]uint64, len(b.History)),
	}
	copy(newBoard.History, b.History)
	return newBoard
}

// MakeMove applies a move to the board.
// It validates that the move is legal before applying it.
// Returns an error if the move is illegal (invalid piece, wrong color, or leaves king in check).
// For pawn promotion, the move must include a valid promotion piece (Queen, Rook, Bishop, Knight).
func (b *Board) MakeMove(m Move) error {
	// Get the piece at the from square
	piece := b.Squares[m.From]

	// Check if this is a valid pawn promotion move missing the promotion piece
	// Only trigger this error if the pawn is actually in position to promote (one rank away)
	if piece.Type() == Pawn {
		fromRank := m.From.Rank()
		toRank := m.To.Rank()
		isValidPromotion := (piece.Color() == White && fromRank == 6 && toRank == 7) ||
			(piece.Color() == Black && fromRank == 1 && toRank == 0)
		if isValidPromotion && m.Promotion == Empty {
			return fmt.Errorf("pawn promotion requires specifying a piece (q, r, b, n)")
		}
	}

	// Check if the move is legal using the full legality check
	if !b.IsLegalMove(m) {
		return fmt.Errorf("illegal move: %s", m.String())
	}

	// Apply the move using the internal method (skips legality check)
	b.applyMove(m)
	return nil
}

// applyMove applies a move to the board without checking legality.
// This is used internally by LegalMoves() to test moves on a copy of the board.
// External code should use MakeMove() which validates legality first.
func (b *Board) applyMove(m Move) {
	piece := b.Squares[m.From]
	capturedPiece := b.Squares[m.To]

	// Save old state for hash updates
	oldCastlingRights := b.CastlingRights
	oldEnPassantSq := b.EnPassantSq

	// --- Zobrist: XOR out the old en passant file (if any) ---
	if b.EnPassantSq >= 0 {
		oldEpFile := Square(b.EnPassantSq).File()
		b.Hash ^= zobristEnPassant[oldEpFile]
	}

	// --- Zobrist: XOR out the moving piece from its source square ---
	b.Hash ^= hashPiece(piece, m.From)

	// --- Zobrist: XOR out the captured piece (if any) ---
	if !capturedPiece.IsEmpty() {
		b.Hash ^= hashPiece(capturedPiece, m.To)
	}

	// Handle en passant capture: remove the captured pawn
	// En passant is detected when a pawn moves to the en passant square
	// The captured pawn is on the same file as destination, but different rank
	var enPassantCapturedPawnSq Square = NoSquare
	if piece.Type() == Pawn && oldEnPassantSq >= 0 && m.To == Square(oldEnPassantSq) {
		// White captures: captured pawn is one rank below (rank - 1)
		// Black captures: captured pawn is one rank above (rank + 1)
		capturedPawnRank := m.To.Rank()
		if piece.Color() == White {
			capturedPawnRank-- // White captures pawn below
		} else {
			capturedPawnRank++ // Black captures pawn above
		}
		enPassantCapturedPawnSq = NewSquare(m.To.File(), capturedPawnRank)
		capturedPawn := b.Squares[enPassantCapturedPawnSq]
		// --- Zobrist: XOR out the captured pawn ---
		b.Hash ^= hashPiece(capturedPawn, enPassantCapturedPawnSq)
		b.Squares[enPassantCapturedPawnSq] = Piece(Empty)
	}

	// Move the piece
	b.Squares[m.To] = piece
	b.Squares[m.From] = Piece(Empty)

	// Determine the final piece at the destination (may be promoted)
	finalPiece := piece

	// Handle pawn promotion: replace pawn with promoted piece
	if piece.Type() == Pawn && m.Promotion != Empty {
		// Create the promoted piece with the same color as the pawn
		promotedPiece := NewPiece(piece.Color(), m.Promotion)
		b.Squares[m.To] = promotedPiece
		finalPiece = promotedPiece
	}

	// --- Zobrist: XOR in the final piece at the destination ---
	b.Hash ^= hashPiece(finalPiece, m.To)

	// Handle castling: if king moves 2 squares horizontally, also move the rook
	if piece.Type() == King {
		fileDiff := m.To.File() - m.From.File()
		if fileDiff == 2 {
			// Kingside castling: move rook from h-file to f-file
			rookFromFile := 7 // h-file
			rookToFile := 5   // f-file
			rank := m.From.Rank()
			rookFrom := NewSquare(rookFromFile, rank)
			rookTo := NewSquare(rookToFile, rank)
			rook := b.Squares[rookFrom]
			// --- Zobrist: XOR out rook from old square, XOR in at new square ---
			b.Hash ^= hashPiece(rook, rookFrom)
			b.Hash ^= hashPiece(rook, rookTo)
			b.Squares[rookTo] = rook
			b.Squares[rookFrom] = Piece(Empty)
		} else if fileDiff == -2 {
			// Queenside castling: move rook from a-file to d-file
			rookFromFile := 0 // a-file
			rookToFile := 3   // d-file
			rank := m.From.Rank()
			rookFrom := NewSquare(rookFromFile, rank)
			rookTo := NewSquare(rookToFile, rank)
			rook := b.Squares[rookFrom]
			// --- Zobrist: XOR out rook from old square, XOR in at new square ---
			b.Hash ^= hashPiece(rook, rookFrom)
			b.Hash ^= hashPiece(rook, rookTo)
			b.Squares[rookTo] = rook
			b.Squares[rookFrom] = Piece(Empty)
		}
	}

	// Update castling rights

	// If king moves, remove both castling rights for that color
	if piece.Type() == King {
		if piece.Color() == White {
			b.CastlingRights &^= (CastleWhiteKing | CastleWhiteQueen)
		} else {
			b.CastlingRights &^= (CastleBlackKing | CastleBlackQueen)
		}
	}

	// If rook moves from original square, remove that side's castling right
	if piece.Type() == Rook {
		switch m.From {
		case NewSquare(0, 0): // a1
			b.CastlingRights &^= CastleWhiteQueen
		case NewSquare(7, 0): // h1
			b.CastlingRights &^= CastleWhiteKing
		case NewSquare(0, 7): // a8
			b.CastlingRights &^= CastleBlackQueen
		case NewSquare(7, 7): // h8
			b.CastlingRights &^= CastleBlackKing
		}
	}

	// If a piece is captured on a rook's original square, remove that castling right
	switch m.To {
	case NewSquare(0, 0): // a1
		b.CastlingRights &^= CastleWhiteQueen
	case NewSquare(7, 0): // h1
		b.CastlingRights &^= CastleWhiteKing
	case NewSquare(0, 7): // a8
		b.CastlingRights &^= CastleBlackQueen
	case NewSquare(7, 7): // h8
		b.CastlingRights &^= CastleBlackKing
	}

	// --- Zobrist: Update castling rights hash (XOR out old, XOR in new) ---
	if b.CastlingRights != oldCastlingRights {
		b.Hash ^= zobristCastling[oldCastlingRights]
		b.Hash ^= zobristCastling[b.CastlingRights]
	}

	// Set en passant square if pawn moves two squares
	if piece.Type() == Pawn {
		rankDiff := m.To.Rank() - m.From.Rank()
		if rankDiff == 2 || rankDiff == -2 {
			// En passant square is the square the pawn "passed through"
			epRank := (m.From.Rank() + m.To.Rank()) / 2
			b.EnPassantSq = int8(NewSquare(m.From.File(), epRank))
		} else {
			b.EnPassantSq = -1
		}
	} else {
		b.EnPassantSq = -1
	}

	// --- Zobrist: XOR in the new en passant file (if any) ---
	if b.EnPassantSq >= 0 {
		newEpFile := Square(b.EnPassantSq).File()
		b.Hash ^= zobristEnPassant[newEpFile]
	}

	// --- Zobrist: Toggle side to move ---
	b.Hash ^= zobristSideToMove

	// Toggle active color
	if b.ActiveColor == White {
		b.ActiveColor = Black
	} else {
		b.ActiveColor = White
		// Increment full move number after Black's move
		b.FullMoveNum++
	}

	// Add the new hash to history for repetition detection
	b.History = append(b.History, b.Hash)
}

// InCheck returns true if the active color's king is under attack by the opponent.
func (b *Board) InCheck() bool {
	// Find the active color's king
	kingSquare := NoSquare
	for sq := Square(0); sq < 64; sq++ {
		piece := b.Squares[sq]
		if piece.Type() == King && piece.Color() == b.ActiveColor {
			kingSquare = sq
			break
		}
	}

	// If no king found (shouldn't happen in a valid game), return false
	if kingSquare == NoSquare {
		return false
	}

	// Determine opponent color
	opponentColor := Black
	if b.ActiveColor == Black {
		opponentColor = White
	}

	// Check if the king's square is attacked by the opponent
	return b.IsSquareAttacked(kingSquare, opponentColor)
}

// String returns a simple text representation of the board for debug printing.
// The board is shown from White's perspective (rank 8 at top).
// Uppercase letters for White pieces (PNBRQK), lowercase for Black (pnbrqk).
// Empty squares are shown as '.'.
func (b *Board) String() string {
	// Piece type to character mapping
	pieceChars := [7]byte{'.', 'P', 'N', 'B', 'R', 'Q', 'K'}

	var result string

	// Print ranks from 8 to 1 (top to bottom from White's perspective)
	for rank := 7; rank >= 0; rank-- {
		// Print rank number (rank 0 = '1', rank 7 = '8')
		result += string(rune('1'+rank)) + " "

		// Print pieces for this rank
		for file := 0; file < 8; file++ {
			sq := Square(rank*8 + file)
			piece := b.Squares[sq]

			var ch byte
			if piece.IsEmpty() {
				ch = '.'
			} else {
				ch = pieceChars[piece.Type()]
				// Lowercase for Black pieces
				if piece.Color() == Black {
					ch = ch - 'A' + 'a'
				}
			}

			result += string(ch)
			if file < 7 {
				result += " "
			}
		}
		result += "\n"
	}

	// Print file letters
	result += "  a b c d e f g h"

	return result
}
