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

	// 5. King safety (Hard only)
	if difficulty >= Hard {
		score += evaluateKingSafety(board)
	}

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

// evaluateKingSafety calculates king safety scores for both kings.
// Returns score from White's perspective (positive = White's king is safer).
func evaluateKingSafety(board *engine.Board) float64 {
	score := 0.0

	// Evaluate White king safety
	whiteKingSq := findKing(board, engine.White)
	if whiteKingSq != -1 {
		whiteKingSafety := evaluateKingSafetyForColor(board, whiteKingSq, engine.White)
		score += whiteKingSafety
	}

	// Evaluate Black king safety
	blackKingSq := findKing(board, engine.Black)
	if blackKingSq != -1 {
		blackKingSafety := evaluateKingSafetyForColor(board, blackKingSq, engine.Black)
		score -= blackKingSafety // Subtract because negative is bad for White
	}

	return score
}

// findKing finds and returns the square index of the king for the given color.
// Returns -1 if the king is not found (should never happen in a legal position).
func findKing(board *engine.Board, color engine.Color) int {
	for sq := 0; sq < 64; sq++ {
		piece := board.PieceAt(engine.Square(sq))
		if piece.Type() == engine.King && piece.Color() == color {
			return sq
		}
	}
	return -1
}

// evaluateKingSafetyForColor evaluates king safety for a specific color.
// Returns a negative penalty (lower = worse safety).
func evaluateKingSafetyForColor(board *engine.Board, kingSq int, color engine.Color) float64 {
	penalty := 0.0

	// 1. Check pawn shield completeness
	penalty += evaluatePawnShield(board, kingSq, color)

	// 2. Check for open files near king
	penalty += evaluateOpenFilesNearKing(board, kingSq, color)

	// 3. Count enemy attackers in king zone
	penalty += evaluateAttackersInKingZone(board, kingSq, color)

	// Return negative penalty (lower score = worse king safety)
	return -penalty
}

// evaluatePawnShield checks if the king has a protective pawn shield.
// Returns a penalty for missing pawns in the shield.
func evaluatePawnShield(board *engine.Board, kingSq int, color engine.Color) float64 {
	kingFile := kingSq % 8
	kingRank := kingSq / 8

	penalty := 0.0
	pawnCount := 0

	// Check 3 files: king's file, and files ±1
	for fileOffset := -1; fileOffset <= 1; fileOffset++ {
		file := kingFile + fileOffset
		if file < 0 || file >= 8 {
			continue
		}

		// Check one rank ahead of king
		var targetRank int
		if color == engine.White {
			targetRank = kingRank + 1
		} else {
			targetRank = kingRank - 1
		}

		if targetRank >= 0 && targetRank < 8 {
			sq := targetRank*8 + file
			piece := board.PieceAt(engine.Square(sq))
			if piece.Type() == engine.Pawn && piece.Color() == color {
				pawnCount++
			}
		}
	}

	// Penalty for missing pawns in shield
	// Ideal: 3 pawns in front, acceptable: 2 pawns
	missingPawns := 3 - pawnCount
	penalty = float64(missingPawns) * 0.3 // 0.3 penalty per missing pawn

	return penalty
}

// evaluateOpenFilesNearKing checks for open files near the king.
// Open files (no pawns) allow enemy rooks/queens to attack.
// Returns a penalty for each open file near the king.
func evaluateOpenFilesNearKing(board *engine.Board, kingSq int, color engine.Color) float64 {
	kingFile := kingSq % 8
	penalty := 0.0

	// Check files around king (king's file ± 1)
	for fileOffset := -1; fileOffset <= 1; fileOffset++ {
		file := kingFile + fileOffset
		if file < 0 || file >= 8 {
			continue
		}

		// Check if this file has any pawns
		hasPawn := false
		for rank := 0; rank < 8; rank++ {
			sq := rank*8 + file
			piece := board.PieceAt(engine.Square(sq))
			if piece.Type() == engine.Pawn {
				hasPawn = true
				break
			}
		}

		// If no pawns on this file, it's open → penalty
		if !hasPawn {
			penalty += 0.25 // 0.25 penalty per open file near king
		}
	}

	return penalty
}

// evaluateAttackersInKingZone counts enemy pieces attacking the king zone.
// King zone is a 3x3 area around the king.
// Returns a penalty based on the number of attacked squares.
func evaluateAttackersInKingZone(board *engine.Board, kingSq int, color engine.Color) float64 {
	kingFile := kingSq % 8
	kingRank := kingSq / 8

	opponentColor := engine.Black
	if color == engine.Black {
		opponentColor = engine.White
	}

	attackerCount := 0

	// Check 3x3 area around king
	for rankOffset := -1; rankOffset <= 1; rankOffset++ {
		for fileOffset := -1; fileOffset <= 1; fileOffset++ {
			targetRank := kingRank + rankOffset
			targetFile := kingFile + fileOffset

			if targetRank < 0 || targetRank >= 8 || targetFile < 0 || targetFile >= 8 {
				continue
			}

			targetSq := engine.Square(targetRank*8 + targetFile)

			// Check if this square is attacked by opponent
			if board.IsSquareAttacked(targetSq, opponentColor) {
				attackerCount++
			}
		}
	}

	// Penalty based on number of attacked squares in king zone
	penalty := float64(attackerCount) * 0.1 // 0.1 penalty per attacked square

	return penalty
}
