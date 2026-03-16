package bot

import "github.com/Mgrdich/TermChess/internal/engine"

const (
	pieceChannels        = 12
	currentPosChannels   = 18
	numHistoryPositions  = 4
	numChannels          = currentPosChannels + numHistoryPositions*pieceChannels // 66
	boardSize            = 8
	encodingSize         = numChannels * boardSize * boardSize // 4224
	policySize           = 4096                                // 64 * 64 from-to squares
)

// encodeBoard converts a chess board position and its history into a 66-channel
// float32 tensor matching the Python training encoder exactly.
// Returns a flat []float32 of length 4224 in [channel, rank, file] order.
//
// Channel layout:
//
//	 0-5:  White pieces (Pawn, Knight, Bishop, Rook, Queen, King) — current position
//	 6-11: Black pieces — current position
//	 12:   Side to move (1.0 if White to move)
//	 13:   White kingside castling available
//	 14:   White queenside castling available
//	 15:   Black kingside castling available
//	 16:   Black queenside castling available
//	 17:   En passant file (column filled with 1.0 if en passant available)
//	 18-29: Piece planes for position t-1 (most recent history)
//	 30-41: Piece planes for position t-2
//	 42-53: Piece planes for position t-3
//	 54-65: Piece planes for position t-4
func encodeBoard(board *engine.Board, history []*engine.Board) []float32 {
	encoding := make([]float32, encodingSize)

	// Channels 0-11: current position piece placement.
	encodePieces(encoding, board, 0)

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
	if board.EnPassantSq >= 0 {
		epFile := int(board.EnPassantSq) % 8
		for rank := 0; rank < 8; rank++ {
			encoding[17*64+rank*8+epFile] = 1.0
		}
	}

	// Channels 18+: history position piece planes (12 channels each).
	// Most recent history position first.
	if len(history) > 0 {
		// Take the last numHistoryPositions entries
		start := 0
		if len(history) > numHistoryPositions {
			start = len(history) - numHistoryPositions
		}
		histSlice := history[start:]

		for i := 0; i < len(histSlice); i++ {
			// i=0 in reversed order = most recent history position
			histIdx := len(histSlice) - 1 - i
			channelOffset := currentPosChannels + i*pieceChannels
			encodePieces(encoding, histSlice[histIdx], channelOffset)
		}
	}

	return encoding
}

// encodePieces writes piece placement data for a board into 12 channels
// starting at the given channel offset.
func encodePieces(encoding []float32, board *engine.Board, channelOffset int) {
	for sq := 0; sq < 64; sq++ {
		piece := board.Squares[sq]
		if piece.IsEmpty() {
			continue
		}

		pt := piece.Type() // 1=Pawn, 2=Knight, ..., 6=King
		pieceIdx := int(pt) - 1

		var channel int
		if piece.Color() == engine.White {
			channel = channelOffset + pieceIdx
		} else {
			channel = channelOffset + 6 + pieceIdx
		}

		rank := sq / 8
		file := sq % 8
		encoding[channel*64+rank*8+file] = 1.0
	}
}

// fillPlane sets all 64 values in the given channel to 1.0.
func fillPlane(encoding []float32, channel int) {
	start := channel * 64
	for i := start; i < start+64; i++ {
		encoding[i] = 1.0
	}
}
