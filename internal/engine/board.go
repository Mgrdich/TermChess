package engine

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

// NewBoard creates a new empty chess board with default game state.
// All squares are empty, White is to move, all castling rights are available,
// no en passant square, half-move clock is 0, and full move number is 1.
func NewBoard() *Board {
	return &Board{
		Squares:        [64]Piece{}, // All zeros = all Empty pieces
		ActiveColor:    White,
		CastlingRights: CastleAll,
		EnPassantSq:    -1,
		HalfMoveClock:  0,
		FullMoveNum:    1,
		Hash:           0,
		History:        []uint64{},
	}
}

// PieceAt returns the piece at the given square.
func (b *Board) PieceAt(sq Square) Piece {
	if !sq.IsValid() {
		return Piece(Empty)
	}
	return b.Squares[sq]
}
