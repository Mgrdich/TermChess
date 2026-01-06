package engine

// GameStatus represents the current state of a chess game.
type GameStatus int

const (
	// Ongoing indicates the game is still in progress.
	Ongoing GameStatus = iota

	// Checkmate indicates the player to move is in checkmate.
	// The opponent wins.
	Checkmate

	// Stalemate indicates the player to move has no legal moves
	// but is not in check. The game is a draw.
	Stalemate

	// DrawInsufficientMaterial indicates a draw due to insufficient
	// material to checkmate (e.g., King vs King, King+Bishop vs King).
	DrawInsufficientMaterial

	// DrawFiftyMoveRule indicates a draw can be claimed under the
	// fifty-move rule (50 moves without pawn move or capture).
	DrawFiftyMoveRule

	// DrawSeventyFiveMoveRule indicates an automatic draw under the
	// seventy-five-move rule (75 moves without pawn move or capture).
	DrawSeventyFiveMoveRule

	// DrawThreefoldRepetition indicates a draw can be claimed due to
	// threefold repetition of the position.
	DrawThreefoldRepetition

	// DrawFivefoldRepetition indicates an automatic draw due to
	// fivefold repetition of the position.
	DrawFivefoldRepetition
)

// String returns a human-readable string representation of the game status.
func (s GameStatus) String() string {
	switch s {
	case Ongoing:
		return "ongoing"
	case Checkmate:
		return "checkmate"
	case Stalemate:
		return "stalemate"
	case DrawInsufficientMaterial:
		return "draw (insufficient material)"
	case DrawFiftyMoveRule:
		return "draw (fifty-move rule)"
	case DrawSeventyFiveMoveRule:
		return "draw (seventy-five-move rule)"
	case DrawThreefoldRepetition:
		return "draw (threefold repetition)"
	case DrawFivefoldRepetition:
		return "draw (fivefold repetition)"
	default:
		return "unknown"
	}
}

// Status returns the current game status by checking for checkmate, stalemate,
// and draw conditions in order of priority.
//
// The algorithm checks:
// 1. If no legal moves exist:
//   - If in check -> Checkmate
//   - If not in check -> Stalemate
//
// 2. Draw conditions:
//   - Fivefold repetition (automatic draw, game ends)
//   - Seventy-five-move rule (automatic draw, game ends)
//   - Insufficient material (automatic draw, game ends)
//   - Threefold repetition (claimable draw, player may claim)
//   - Fifty-move rule (claimable draw, player may claim)
//
// 3. Otherwise -> Ongoing
//
// Note: Claimable draws (threefold, fifty-move) do not automatically end the game.
// Use CanClaimDraw() to check if a draw is available, and IsGameOver() to check
// if the game has actually ended.
func (b *Board) Status() GameStatus {
	// Generate all legal moves for the active player
	legalMoves := b.LegalMoves()

	// If no legal moves exist, check for checkmate or stalemate
	if len(legalMoves) == 0 {
		if b.InCheck() {
			return Checkmate
		}
		return Stalemate
	}

	// Check for automatic draws first (these end the game immediately)

	// Check for fivefold repetition (automatic draw)
	repCount := b.repetitionCount()
	if repCount >= 5 {
		return DrawFivefoldRepetition
	}

	// Check for seventy-five-move rule (automatic draw)
	// 75 full moves = 150 half-moves
	if b.HalfMoveClock >= 150 {
		return DrawSeventyFiveMoveRule
	}

	// Check for insufficient material (automatic draw)
	if b.hasInsufficientMaterial() {
		return DrawInsufficientMaterial
	}

	// Check for claimable draws (these require a player to claim)

	// Check for threefold repetition (claimable draw)
	if repCount >= 3 {
		return DrawThreefoldRepetition
	}

	// Check for fifty-move rule (claimable draw)
	// 50 full moves = 100 half-moves
	if b.HalfMoveClock >= 100 {
		return DrawFiftyMoveRule
	}

	return Ongoing
}

// IsGameOver returns true if the game has ended due to an automatic game-ending
// condition: checkmate, stalemate, or automatic draws (fivefold repetition,
// seventy-five-move rule, insufficient material).
//
// Note: This does NOT include claimable draws (threefold repetition, fifty-move rule).
// Use CanClaimDraw() to check if a draw is available to claim.
func (b *Board) IsGameOver() bool {
	status := b.Status()
	// Game is over for automatic conditions only (not claimable draws)
	switch status {
	case Checkmate, Stalemate, DrawFivefoldRepetition, DrawSeventyFiveMoveRule, DrawInsufficientMaterial:
		return true
	default:
		return false
	}
}

// CanClaimDraw returns true if a draw is available to claim according to FIDE rules.
// This includes:
//   - Threefold repetition: the same position has occurred 3 or more times
//   - Fifty-move rule: 50 moves have been made without a pawn move or capture
//
// Unlike automatic draws (fivefold, seventy-five-move), these draws require a player
// to claim them. The game can continue if the player chooses not to claim.
func (b *Board) CanClaimDraw() bool {
	status := b.Status()
	return status == DrawThreefoldRepetition || status == DrawFiftyMoveRule
}

// Winner returns the color of the winning player and whether there is a winner.
// Returns (Black, true) if White is checkmated, (White, true) if Black is checkmated,
// or (0, false) for stalemate, draws, or ongoing games.
func (b *Board) Winner() (Color, bool) {
	if b.Status() == Checkmate {
		// The player to move is checkmated, so the opponent wins
		if b.ActiveColor == White {
			return Black, true
		}
		return White, true
	}
	return 0, false // No winner (draw, stalemate, or ongoing)
}

// repetitionCount returns the number of times the current position
// has occurred in the game history. The current position's hash
// is included in the history (added after the last move was made).
func (b *Board) repetitionCount() int {
	count := 0
	for _, hash := range b.History {
		if hash == b.Hash {
			count++
		}
	}
	return count
}

// materialCount holds the count of each piece type for both colors.
type materialCount struct {
	whitePawns   int
	whiteKnights int
	whiteBishops int
	whiteRooks   int
	whiteQueens  int
	blackPawns   int
	blackKnights int
	blackBishops int
	blackRooks   int
	blackQueens  int
	// Bishops stored by their square for color detection
	whiteBishopSquares []Square
	blackBishopSquares []Square
}

// countMaterial counts all pieces on the board (excluding kings).
func (b *Board) countMaterial() materialCount {
	mc := materialCount{
		whiteBishopSquares: []Square{},
		blackBishopSquares: []Square{},
	}

	for sq := Square(0); sq < 64; sq++ {
		piece := b.Squares[sq]
		if piece.IsEmpty() {
			continue
		}

		pieceType := piece.Type()
		pieceColor := piece.Color()

		if pieceColor == White {
			switch pieceType {
			case Pawn:
				mc.whitePawns++
			case Knight:
				mc.whiteKnights++
			case Bishop:
				mc.whiteBishops++
				mc.whiteBishopSquares = append(mc.whiteBishopSquares, sq)
			case Rook:
				mc.whiteRooks++
			case Queen:
				mc.whiteQueens++
			}
		} else { // Black
			switch pieceType {
			case Pawn:
				mc.blackPawns++
			case Knight:
				mc.blackKnights++
			case Bishop:
				mc.blackBishops++
				mc.blackBishopSquares = append(mc.blackBishopSquares, sq)
			case Rook:
				mc.blackRooks++
			case Queen:
				mc.blackQueens++
			}
		}
	}

	return mc
}

// hasInsufficientMaterial returns true if neither side can force checkmate.
// This occurs in the following scenarios:
// - K vs K (king versus king)
// - K+B vs K (king and bishop versus king)
// - K+N vs K (king and knight versus king)
// - K+B vs K+B where both bishops are on the same color squares
func (b *Board) hasInsufficientMaterial() bool {
	mc := b.countMaterial()

	// If there are any pawns, rooks, or queens, there is sufficient material
	if mc.whitePawns > 0 || mc.blackPawns > 0 ||
		mc.whiteRooks > 0 || mc.blackRooks > 0 ||
		mc.whiteQueens > 0 || mc.blackQueens > 0 {
		return false
	}

	// Count total minor pieces (knights and bishops) for each side
	whiteMinorPieces := mc.whiteKnights + mc.whiteBishops
	blackMinorPieces := mc.blackKnights + mc.blackBishops

	// K vs K (no minor pieces on either side)
	if whiteMinorPieces == 0 && blackMinorPieces == 0 {
		return true
	}

	// K+B vs K or K vs K+B (one bishop, no other pieces)
	if whiteMinorPieces == 1 && mc.whiteBishops == 1 && blackMinorPieces == 0 {
		return true
	}
	if blackMinorPieces == 1 && mc.blackBishops == 1 && whiteMinorPieces == 0 {
		return true
	}

	// K+N vs K or K vs K+N (one knight, no other pieces)
	if whiteMinorPieces == 1 && mc.whiteKnights == 1 && blackMinorPieces == 0 {
		return true
	}
	if blackMinorPieces == 1 && mc.blackKnights == 1 && whiteMinorPieces == 0 {
		return true
	}

	// K+B vs K+B with same-color bishops
	// (one bishop each, both on the same square color)
	if mc.whiteBishops == 1 && mc.blackBishops == 1 &&
		whiteMinorPieces == 1 && blackMinorPieces == 1 {
		// Check if bishops are on the same color squares
		// A square's color is determined by (rank + file) % 2
		// If both have the same parity, they're on the same color
		whiteBishopSq := mc.whiteBishopSquares[0]
		blackBishopSq := mc.blackBishopSquares[0]

		whiteSquareColor := (whiteBishopSq.Rank() + whiteBishopSq.File()) % 2
		blackSquareColor := (blackBishopSq.Rank() + blackBishopSq.File()) % 2

		if whiteSquareColor == blackSquareColor {
			return true
		}
	}

	return false
}
