package bot

import (
	"github.com/Mgrdich/TermChess/internal/engine"
)

// pieceValues defines standard chess piece values in pawns.
var pieceValues = map[engine.PieceType]float64{
	engine.Pawn:   1.0,
	engine.Knight: 3.0,
	engine.Bishop: 3.25,
	engine.Rook:   5.0,
	engine.Queen:  9.0,
	engine.King:   0.0, // Invaluable (not counted in material)
}

// evaluate returns a score for the position from White's perspective.
// Positive = White advantage, Negative = Black advantage
func evaluate(board *engine.Board, difficulty Difficulty) float64 {
	// 1. Check terminal states first
	status := board.Status()

	if status == engine.Checkmate {
		winner, _ := board.Winner()
		if winner == engine.White {
			return 10000.0
		}
		return -10000.0
	}

	if status == engine.Stalemate || status == engine.DrawThreefoldRepetition ||
		status == engine.DrawFiftyMoveRule || status == engine.DrawInsufficientMaterial ||
		status == engine.DrawFivefoldRepetition || status == engine.DrawSeventyFiveMoveRule {
		return 0.0
	}

	score := 0.0

	// 2. Material count (all difficulties)
	score += countMaterial(board)

	// Future tasks will add:
	// 3. Piece-square tables (Medium+)
	// 4. Mobility (Medium+)
	// 5. King safety (Hard only)

	return score
}

// countMaterial calculates the material balance from White's perspective.
func countMaterial(board *engine.Board) float64 {
	score := 0.0

	// Iterate all 64 squares
	for sq := 0; sq < 64; sq++ {
		piece := board.PieceAt(engine.Square(sq))

		// Skip empty squares
		if piece.IsEmpty() {
			continue
		}

		pieceType := piece.Type()
		value := pieceValues[pieceType]

		// Add for White pieces, subtract for Black pieces
		if piece.Color() == engine.White {
			score += value
		} else {
			score -= value
		}
	}

	return score
}
