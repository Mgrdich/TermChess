package bot

import (
	"github.com/Mgrdich/TermChess/internal/engine"
)

// totalStartingMaterial is the sum of non-pawn, non-king piece values at game start.
// 2*Queen(9) + 4*Rook(5) + 4*Bishop(3.25) + 4*Knight(3) = 18 + 20 + 13 + 12 = 63
const totalStartingMaterial = 63.0

// endgameThreshold is the material level below which the position is considered a pure endgame.
const endgameThreshold = 16.0

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

// kingMiddlegameTable rewards castled king positions and penalizes exposed kings.
// Used during the opening/middlegame phase for king safety via piece-square evaluation.
var kingMiddlegameTable = [64]float64{
	// Rank 1 - castled king positions are best
	0.2, 0.3, 0.1, 0.0, 0.0, 0.1, 0.3, 0.2,
	// Rank 2 - behind pawns is OK
	0.2, 0.2, 0.0, 0.0, 0.0, 0.0, 0.2, 0.2,
	// Rank 3-8 - penalize exposed king progressively
	-0.1, -0.2, -0.2, -0.3, -0.3, -0.2, -0.2, -0.1,
	-0.2, -0.3, -0.3, -0.4, -0.4, -0.3, -0.3, -0.2,
	-0.3, -0.4, -0.4, -0.5, -0.5, -0.4, -0.4, -0.3,
	-0.3, -0.4, -0.4, -0.5, -0.5, -0.4, -0.4, -0.3,
	-0.3, -0.4, -0.4, -0.5, -0.5, -0.4, -0.4, -0.3,
	-0.3, -0.4, -0.4, -0.5, -0.5, -0.4, -0.4, -0.3,
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

// passedPawnBonus gives bonuses for passed pawns by rank (index = rank 0-7).
// Higher ranks = closer to promotion = bigger bonus.
var passedPawnBonus = [8]float64{
	0.0,  // rank 0 (impossible for pawns)
	0.0,  // rank 1 (White starting rank, Black promotion)
	0.1,  // rank 2
	0.2,  // rank 3
	0.35, // rank 4
	0.6,  // rank 5
	1.0,  // rank 6
	1.5,  // rank 7 (White promotion rank, Black starting)
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

	// 3. Piece-square tables, passed pawns, and mobility (Medium+)
	if difficulty >= Medium {
		phase := computeGamePhase(board)
		score += evaluatePiecePositions(board, phase)
		score += evaluatePassedPawns(board, phase)
		score += evaluateMobility(board) * 0.1 // Weight mobility at 10%
	}

	// 4. King safety (Hard only)
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
// The phase parameter (0.0=endgame to 1.0=opening) is used to interpolate
// between middlegame and endgame king piece-square tables.
// Returns score from White's perspective.
func evaluatePiecePositions(board *engine.Board, phase float64) float64 {
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
			mgBonus := kingMiddlegameTable[squareIndex]
			egBonus := kingEndgameTable[squareIndex]
			bonus = phase*mgBonus + (1.0-phase)*egBonus
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

// computeGamePhase returns a value between 0.0 (endgame) and 1.0 (opening)
// based on remaining non-pawn material on the board.
// Phase is 1.0 at the starting position and 0.0 when only kings/pawns remain.
func computeGamePhase(board *engine.Board) float64 {
	material := countNonPawnMaterial(board)
	if material <= endgameThreshold {
		return 0.0
	}
	if material >= totalStartingMaterial {
		return 1.0
	}
	return (material - endgameThreshold) / (totalStartingMaterial - endgameThreshold)
}

// countNonPawnMaterial sums the piece values for all non-pawn, non-king pieces
// on the board (both colors combined).
func countNonPawnMaterial(board *engine.Board) float64 {
	material := 0.0
	for sq := 0; sq < 64; sq++ {
		piece := board.PieceAt(engine.Square(sq))
		if piece.IsEmpty() {
			continue
		}
		pieceType := piece.Type()
		if pieceType == engine.Pawn || pieceType == engine.King {
			continue
		}
		material += pieceValues[pieceType]
	}
	return material
}

// isPassedPawn checks if a pawn at the given square is a passed pawn.
// A passed pawn is a pawn with no enemy pawns on the same file or adjacent
// files that can block or capture it.
func isPassedPawn(board *engine.Board, sq int, color engine.Color) bool {
	file := sq % 8
	rank := sq / 8

	// Check files: current file and adjacent files
	for f := max(0, file-1); f <= min(7, file+1); f++ {
		// Check ranks ahead of this pawn
		if color == engine.White {
			// White pawns move up (higher ranks)
			for r := rank + 1; r <= 7; r++ {
				checkSq := r*8 + f
				piece := board.PieceAt(engine.Square(checkSq))
				if !piece.IsEmpty() && piece.Type() == engine.Pawn && piece.Color() == engine.Black {
					return false // Blocked by enemy pawn
				}
			}
		} else {
			// Black pawns move down (lower ranks)
			for r := rank - 1; r >= 0; r-- {
				checkSq := r*8 + f
				piece := board.PieceAt(engine.Square(checkSq))
				if !piece.IsEmpty() && piece.Type() == engine.Pawn && piece.Color() == engine.White {
					return false // Blocked by enemy pawn
				}
			}
		}
	}
	return true
}

// evaluatePassedPawns scores passed pawns from White's perspective.
// Bonus is scaled by (1.0 + (1.0 - phase)) to double in pure endgame.
func evaluatePassedPawns(board *engine.Board, phase float64) float64 {
	score := 0.0
	phaseMultiplier := 1.0 + (1.0 - phase) // 1.0 in opening, 2.0 in endgame

	for sq := 0; sq < 64; sq++ {
		piece := board.PieceAt(engine.Square(sq))
		if piece.IsEmpty() || piece.Type() != engine.Pawn {
			continue
		}

		color := piece.Color()
		if !isPassedPawn(board, sq, color) {
			continue
		}

		rank := sq / 8
		var bonus float64
		if color == engine.White {
			bonus = passedPawnBonus[rank] * phaseMultiplier
			score += bonus
		} else {
			// For Black, flip the rank (rank 1 for Black = rank 6 equivalent)
			flippedRank := 7 - rank
			bonus = passedPawnBonus[flippedRank] * phaseMultiplier
			score -= bonus
		}
	}
	return score
}
