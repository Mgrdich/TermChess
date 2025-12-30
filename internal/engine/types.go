// Package engine implements the chess engine for TermChess.
package engine

// Color represents the color of a chess piece (White or Black).
type Color uint8

const (
	// White is the white player (value 0).
	White Color = 0
	// Black is the black player (value 1).
	Black Color = 1
)

// PieceType represents the type of a chess piece.
type PieceType uint8

const (
	// Empty represents an empty square.
	Empty PieceType = 0
	// Pawn represents a pawn piece.
	Pawn PieceType = 1
	// Knight represents a knight piece.
	Knight PieceType = 2
	// Bishop represents a bishop piece.
	Bishop PieceType = 3
	// Rook represents a rook piece.
	Rook PieceType = 4
	// Queen represents a queen piece.
	Queen PieceType = 5
	// King represents a king piece.
	King PieceType = 6
)

// Piece represents a chess piece encoded as a single byte.
// The high bit stores the color (0=White, 1=Black).
// The low 3 bits store the piece type.
type Piece uint8

// NewPiece creates a new Piece with the given color and piece type.
func NewPiece(color Color, pieceType PieceType) Piece {
	return Piece((uint8(color) << 7) | uint8(pieceType))
}

// Color returns the color of the piece.
func (p Piece) Color() Color {
	return Color(p >> 7)
}

// Type returns the type of the piece.
func (p Piece) Type() PieceType {
	return PieceType(p & 0x07)
}

// IsEmpty returns true if the piece is empty (no piece on square).
func (p Piece) IsEmpty() bool {
	return p.Type() == Empty
}

// Square represents a square on the chess board (0-63).
// Indexed as rank * 8 + file, where a1 = 0, h8 = 63.
type Square int8

const (
	// NoSquare represents an invalid or non-existent square.
	NoSquare Square = -1
)

// NewSquare creates a Square from file and rank (both 0-7).
// file: 0=a, 1=b, ..., 7=h
// rank: 0=1, 1=2, ..., 7=8
func NewSquare(file, rank int) Square {
	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return NoSquare
	}
	return Square(rank*8 + file)
}

// File returns the file of the square (0=a, 1=b, ..., 7=h).
func (s Square) File() int {
	return int(s) % 8
}

// Rank returns the rank of the square (0=1, 1=2, ..., 7=8).
func (s Square) Rank() int {
	return int(s) / 8
}

// IsValid returns true if the square is a valid board square (0-63).
func (s Square) IsValid() bool {
	return s >= 0 && s <= 63
}

// String returns the algebraic notation of the square (e.g., "a1", "h8").
func (s Square) String() string {
	if !s.IsValid() {
		return "-"
	}
	file := 'a' + rune(s.File())
	rank := '1' + rune(s.Rank())
	return string(file) + string(rank)
}
