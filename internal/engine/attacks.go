package engine

// IsSquareAttacked returns true if the given square is attacked by any piece of the specified color.
// This is an efficient check that works backwards from the target square to potential attackers.
func (b *Board) IsSquareAttacked(sq Square, byColor Color) bool {
	if !sq.IsValid() {
		return false
	}

	file := sq.File()
	rank := sq.Rank()

	// Check pawn attacks
	if b.isSquareAttackedByPawn(sq, file, rank, byColor) {
		return true
	}

	// Check knight attacks
	if b.isSquareAttackedByKnight(sq, file, rank, byColor) {
		return true
	}

	// Check king attacks
	if b.isSquareAttackedByKing(sq, file, rank, byColor) {
		return true
	}

	// Check diagonal attacks (bishop or queen)
	if b.isSquareAttackedDiagonally(file, rank, byColor) {
		return true
	}

	// Check orthogonal attacks (rook or queen)
	if b.isSquareAttackedOrthogonally(file, rank, byColor) {
		return true
	}

	return false
}

// isSquareAttackedByPawn checks if the square is attacked by a pawn of the given color.
// Pawns attack diagonally: white pawns attack upward (increasing rank),
// black pawns attack downward (decreasing rank).
func (b *Board) isSquareAttackedByPawn(sq Square, file, rank int, byColor Color) bool {
	// The pawn that could attack this square is one rank behind (from the attacker's perspective)
	// For white pawns attacking upward: they would be on rank-1
	// For black pawns attacking downward: they would be on rank+1
	var attackerRank int
	if byColor == White {
		attackerRank = rank - 1 // White pawn is below the target square
	} else {
		attackerRank = rank + 1 // Black pawn is above the target square
	}

	// Check bounds
	if attackerRank < 0 || attackerRank > 7 {
		return false
	}

	// Check both diagonal attack positions (file-1 and file+1)
	for _, attackerFile := range []int{file - 1, file + 1} {
		if attackerFile < 0 || attackerFile > 7 {
			continue
		}

		attackerSq := NewSquare(attackerFile, attackerRank)
		piece := b.Squares[attackerSq]
		if piece.Type() == Pawn && piece.Color() == byColor {
			return true
		}
	}

	return false
}

// isSquareAttackedByKnight checks if the square is attacked by a knight of the given color.
// Knights move in an L-shape: 2 squares in one direction, 1 square perpendicular.
func (b *Board) isSquareAttackedByKnight(sq Square, file, rank int, byColor Color) bool {
	// Knight move offsets: (file delta, rank delta)
	offsets := [][2]int{
		{+2, +1}, {+2, -1}, {-2, +1}, {-2, -1},
		{+1, +2}, {+1, -2}, {-1, +2}, {-1, -2},
	}

	for _, offset := range offsets {
		attackerFile := file + offset[0]
		attackerRank := rank + offset[1]

		// Check bounds
		if attackerFile < 0 || attackerFile > 7 || attackerRank < 0 || attackerRank > 7 {
			continue
		}

		attackerSq := NewSquare(attackerFile, attackerRank)
		piece := b.Squares[attackerSq]
		if piece.Type() == Knight && piece.Color() == byColor {
			return true
		}
	}

	return false
}

// isSquareAttackedByKing checks if the square is attacked by the king of the given color.
// Kings can attack any of the 8 adjacent squares.
func (b *Board) isSquareAttackedByKing(sq Square, file, rank int, byColor Color) bool {
	// King move offsets: all 8 adjacent squares
	offsets := [][2]int{
		{+1, +1}, {+1, -1}, {-1, +1}, {-1, -1}, // diagonal
		{+1, 0}, {-1, 0}, {0, +1}, {0, -1}, // orthogonal
	}

	for _, offset := range offsets {
		attackerFile := file + offset[0]
		attackerRank := rank + offset[1]

		// Check bounds
		if attackerFile < 0 || attackerFile > 7 || attackerRank < 0 || attackerRank > 7 {
			continue
		}

		attackerSq := NewSquare(attackerFile, attackerRank)
		piece := b.Squares[attackerSq]
		if piece.Type() == King && piece.Color() == byColor {
			return true
		}
	}

	return false
}

// isSquareAttackedDiagonally checks if the square is attacked by a bishop or queen diagonally.
// Slides along diagonals until hitting a piece or the board edge.
func (b *Board) isSquareAttackedDiagonally(file, rank int, byColor Color) bool {
	// Diagonal directions: (+1,+1), (+1,-1), (-1,+1), (-1,-1)
	directions := [][2]int{
		{+1, +1}, {+1, -1}, {-1, +1}, {-1, -1},
	}

	for _, dir := range directions {
		for dist := 1; dist <= 7; dist++ {
			attackerFile := file + dir[0]*dist
			attackerRank := rank + dir[1]*dist

			// Check bounds
			if attackerFile < 0 || attackerFile > 7 || attackerRank < 0 || attackerRank > 7 {
				break
			}

			attackerSq := NewSquare(attackerFile, attackerRank)
			piece := b.Squares[attackerSq]

			if piece.IsEmpty() {
				// Empty square, continue sliding
				continue
			}

			// Found a piece
			if piece.Color() == byColor {
				// It's the attacker's piece - check if it can attack diagonally
				if piece.Type() == Bishop || piece.Type() == Queen {
					return true
				}
			}
			// Either wrong color or a piece that can't attack diagonally - stop this direction
			break
		}
	}

	return false
}

// isSquareAttackedOrthogonally checks if the square is attacked by a rook or queen orthogonally.
// Slides along ranks and files until hitting a piece or the board edge.
func (b *Board) isSquareAttackedOrthogonally(file, rank int, byColor Color) bool {
	// Orthogonal directions: (+1,0), (-1,0), (0,+1), (0,-1)
	directions := [][2]int{
		{+1, 0}, {-1, 0}, {0, +1}, {0, -1},
	}

	for _, dir := range directions {
		for dist := 1; dist <= 7; dist++ {
			attackerFile := file + dir[0]*dist
			attackerRank := rank + dir[1]*dist

			// Check bounds
			if attackerFile < 0 || attackerFile > 7 || attackerRank < 0 || attackerRank > 7 {
				break
			}

			attackerSq := NewSquare(attackerFile, attackerRank)
			piece := b.Squares[attackerSq]

			if piece.IsEmpty() {
				// Empty square, continue sliding
				continue
			}

			// Found a piece
			if piece.Color() == byColor {
				// It's the attacker's piece - check if it can attack orthogonally
				if piece.Type() == Rook || piece.Type() == Queen {
					return true
				}
			}
			// Either wrong color or a piece that can't attack orthogonally - stop this direction
			break
		}
	}

	return false
}
