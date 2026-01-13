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

// Piece-square tables give positional bonuses (from White's perspective).
// Values are in fractions of a pawn (0.5 = half a pawn bonus).
// For Black pieces, the rank is flipped: sq_flipped = (7-rank)*8 + file

// pawnTable encourages pawn advancement and central control.
var pawnTable = [64]float64{
	// Rank 1 (White's back rank) - pawns shouldn't be here
	0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
	// Rank 2 - starting position
	0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
	// Rank 3 - slight advancement bonus
	0.1, 0.1, 0.2, 0.3, 0.3, 0.2, 0.1, 0.1,
	// Rank 4 - good advancement, central pawns more valuable
	0.15, 0.15, 0.2, 0.35, 0.35, 0.2, 0.15, 0.15,
	// Rank 5 - strong advancement
	0.2, 0.2, 0.3, 0.4, 0.4, 0.3, 0.2, 0.2,
	// Rank 6 - near promotion
	0.3, 0.3, 0.4, 0.5, 0.5, 0.4, 0.3, 0.3,
	// Rank 7 - very close to promotion
	0.5, 0.5, 0.6, 0.7, 0.7, 0.6, 0.5, 0.5,
	// Rank 8 - promotion square (shouldn't have pawns here)
	0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
}

// knightTable encourages centralization and development.
var knightTable = [64]float64{
	// Rank 1 - back rank, need development
	-0.5, -0.4, -0.3, -0.3, -0.3, -0.3, -0.4, -0.5,
	// Rank 2 - still on back ranks
	-0.4, -0.2, 0.0, 0.0, 0.0, 0.0, -0.2, -0.4,
	// Rank 3 - developed position
	-0.3, 0.0, 0.1, 0.15, 0.15, 0.1, 0.0, -0.3,
	// Rank 4 - good central squares
	-0.3, 0.05, 0.15, 0.2, 0.2, 0.15, 0.05, -0.3,
	// Rank 5 - excellent central control
	-0.3, 0.0, 0.15, 0.2, 0.2, 0.15, 0.0, -0.3,
	// Rank 6 - advanced but less stable
	-0.3, 0.05, 0.1, 0.15, 0.15, 0.1, 0.05, -0.3,
	// Rank 7 - too far advanced
	-0.4, -0.2, 0.0, 0.05, 0.05, 0.0, -0.2, -0.4,
	// Rank 8 - rim knights are dim
	-0.5, -0.4, -0.3, -0.3, -0.3, -0.3, -0.4, -0.5,
}

// bishopTable encourages long diagonals and central control.
var bishopTable = [64]float64{
	// Rank 1
	-0.2, -0.1, -0.1, -0.1, -0.1, -0.1, -0.1, -0.2,
	// Rank 2
	-0.1, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, -0.1,
	// Rank 3
	-0.1, 0.0, 0.05, 0.1, 0.1, 0.05, 0.0, -0.1,
	// Rank 4
	-0.1, 0.05, 0.05, 0.1, 0.1, 0.05, 0.05, -0.1,
	// Rank 5
	-0.1, 0.0, 0.1, 0.1, 0.1, 0.1, 0.0, -0.1,
	// Rank 6
	-0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, -0.1,
	// Rank 7
	-0.1, 0.05, 0.0, 0.0, 0.0, 0.0, 0.05, -0.1,
	// Rank 8
	-0.2, -0.1, -0.1, -0.1, -0.1, -0.1, -0.1, -0.2,
}

// rookTable encourages 7th rank occupation and central files.
var rookTable = [64]float64{
	// Rank 1 - back rank, OK for castled position
	0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
	// Rank 2
	0.05, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.05,
	// Rank 3
	-0.05, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, -0.05,
	// Rank 4
	-0.05, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, -0.05,
	// Rank 5
	-0.05, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, -0.05,
	// Rank 6
	-0.05, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, -0.05,
	// Rank 7 - the famous 7th rank bonus
	0.25, 0.25, 0.25, 0.25, 0.25, 0.25, 0.25, 0.25,
	// Rank 8
	0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
}

// kingEndgameTable encourages king centralization in the endgame.
// In opening/middlegame, the king should stay protected (corners).
var kingEndgameTable = [64]float64{
	// Rank 1 - corners are safer early, but we use endgame table
	-0.3, -0.4, -0.4, -0.5, -0.5, -0.4, -0.4, -0.3,
	// Rank 2
	-0.3, -0.4, -0.4, -0.5, -0.5, -0.4, -0.4, -0.3,
	// Rank 3
	-0.3, -0.4, -0.2, 0.0, 0.0, -0.2, -0.4, -0.3,
	// Rank 4 - center becomes better
	-0.3, -0.3, 0.0, 0.2, 0.2, 0.0, -0.3, -0.3,
	// Rank 5
	-0.3, -0.3, 0.0, 0.2, 0.2, 0.0, -0.3, -0.3,
	// Rank 6
	-0.3, -0.4, -0.2, 0.0, 0.0, -0.2, -0.4, -0.3,
	// Rank 7
	-0.3, -0.4, -0.4, -0.5, -0.5, -0.4, -0.4, -0.3,
	// Rank 8
	-0.3, -0.4, -0.4, -0.5, -0.5, -0.4, -0.4, -0.3,
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

	// 3. Piece-square tables (Medium+)
	if difficulty >= Medium {
		score += evaluatePiecePositions(board)
	}

	// 4. Mobility (Medium+)
	if difficulty >= Medium {
		score += evaluateMobility(board) * 0.1 // Weight mobility at 10%
	}

	// Future tasks will add:
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

// evaluatePiecePositions calculates positional bonuses using piece-square tables.
// Returns score from White's perspective.
func evaluatePiecePositions(board *engine.Board) float64 {
	score := 0.0

	for sq := 0; sq < 64; sq++ {
		piece := board.PieceAt(engine.Square(sq))
		if piece.IsEmpty() {
			continue
		}

		pieceType := piece.Type()
		color := piece.Color()

		// Get piece-square table bonus
		var bonus float64
		squareIndex := sq

		// Flip square for Black pieces (Black plays from rank 7)
		if color == engine.Black {
			rank := sq / 8
			file := sq % 8
			squareIndex = (7-rank)*8 + file
		}

		switch pieceType {
		case engine.Pawn:
			bonus = pawnTable[squareIndex]
		case engine.Knight:
			bonus = knightTable[squareIndex]
		case engine.Bishop:
			bonus = bishopTable[squareIndex]
		case engine.Rook:
			bonus = rookTable[squareIndex]
		case engine.King:
			bonus = kingEndgameTable[squareIndex]
		case engine.Queen:
			// Queens don't have a specific table, use 0
			bonus = 0.0
		}

		// Add bonus for White, subtract for Black
		if color == engine.White {
			score += bonus
		} else {
			score -= bonus
		}
	}

	return score
}

// evaluateMobility calculates a mobility score based on legal move count.
// More legal moves = better position (more options).
// Returns score from White's perspective.
func evaluateMobility(board *engine.Board) float64 {
	// Count legal moves for active player
	moves := board.LegalMoves()
	mobilityScore := float64(len(moves))

	// Return from White's perspective
	if board.ActiveColor == engine.Black {
		return -mobilityScore
	}
	return mobilityScore
}
