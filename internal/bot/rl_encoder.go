package bot

import "github.com/Mgrdich/TermChess/internal/engine"

const (
	numChannels  = 18
	boardSize    = 8
	encodingSize = numChannels * boardSize * boardSize // 1152
	policySize   = 4096                                // 64 * 64 from-to squares
)

// encodeBoard converts a chess board position into an 18-channel float32 tensor
// matching the Python training encoder exactly.
// Returns a flat []float32 of length 1152 in [channel, rank, file] order.
//
// Channel layout:
//
//	 0-5:  White pieces (Pawn, Knight, Bishop, Rook, Queen, King)
//	 6-11: Black pieces (Pawn, Knight, Bishop, Rook, Queen, King)
//	 12:   Side to move (1.0 if White to move)
//	 13:   White kingside castling available
//	 14:   White queenside castling available
//	 15:   Black kingside castling available
//	 16:   Black queenside castling available
//	 17:   En passant file (column filled with 1.0 if en passant available)
func encodeBoard(board *engine.Board) []float32 {
	encoding := make([]float32, encodingSize)

	// Channels 0-11: piece placement.
	// For each square, if a piece is present, set the corresponding channel.
	// White pieces: channel = pieceType - 1 (Pawn=0, Knight=1, ..., King=5)
	// Black pieces: channel = 6 + (pieceType - 1)
	for sq := 0; sq < 64; sq++ {
		piece := board.Squares[sq]
		if piece.IsEmpty() {
			continue
		}

		pt := piece.Type() // 1=Pawn, 2=Knight, ..., 6=King
		channelOffset := int(pt) - 1

		var channel int
		if piece.Color() == engine.White {
			channel = channelOffset
		} else {
			channel = 6 + channelOffset
		}

		rank := sq / 8
		file := sq % 8
		encoding[channel*64+rank*8+file] = 1.0
	}

	// Channel 12: side to move (entire plane 1.0 if White to move).
	if board.ActiveColor == engine.White {
		fillPlane(encoding, 12)
	}

	// Channels 13-16: castling rights.
	if board.CastlingRights&engine.CastleWhiteKing != 0 {
		fillPlane(encoding, 13)
	}
	if board.CastlingRights&engine.CastleWhiteQueen != 0 {
		fillPlane(encoding, 14)
	}
	if board.CastlingRights&engine.CastleBlackKing != 0 {
		fillPlane(encoding, 15)
	}
	if board.CastlingRights&engine.CastleBlackQueen != 0 {
		fillPlane(encoding, 16)
	}

	// Channel 17: en passant.
	// If en passant is available, fill the entire column (file) with 1.0.
	if board.EnPassantSq >= 0 {
		epFile := int(board.EnPassantSq) % 8
		for rank := 0; rank < 8; rank++ {
			encoding[17*64+rank*8+epFile] = 1.0
		}
	}

	return encoding
}

// fillPlane sets all 64 values in the given channel to 1.0.
func fillPlane(encoding []float32, channel int) {
	start := channel * 64
	for i := start; i < start+64; i++ {
		encoding[i] = 1.0
	}
}
