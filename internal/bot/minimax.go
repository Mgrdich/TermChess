package bot

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// minimaxEngine implements Medium and Hard bots using minimax with alpha-beta pruning.
type minimaxEngine struct {
	name        string
	difficulty  Difficulty
	maxDepth    int
	timeLimit   time.Duration
	evalWeights evalWeights
	closed      bool
}

// evalWeights holds the weights for different evaluation components.
// This will be used in future tasks when we add more evaluation features.
type evalWeights struct {
	material    float64
	pieceSquare float64
	mobility    float64
	kingSafety  float64
}

// getDefaultWeights returns appropriate evaluation weights based on difficulty.
func getDefaultWeights(difficulty Difficulty) evalWeights {
	switch difficulty {
	case Medium:
		return evalWeights{
			material:    1.0,
			pieceSquare: 0.0, // Will be added in future tasks
			mobility:    0.0, // Will be added in future tasks
			kingSafety:  0.0, // Will be added in future tasks
		}
	case Hard:
		return evalWeights{
			material:    1.0,
			pieceSquare: 0.0, // Will be added in future tasks
			mobility:    0.0, // Will be added in future tasks
			kingSafety:  0.0, // Will be added in future tasks
		}
	default:
		// Fallback to Medium weights
		return evalWeights{
			material:    1.0,
			pieceSquare: 0.0,
			mobility:    0.0,
			kingSafety:  0.0,
		}
	}
}

// Name returns the human-readable name of this engine.
func (e *minimaxEngine) Name() string {
	return e.name
}

// Close releases resources held by the engine.
func (e *minimaxEngine) Close() error {
	e.closed = true
	return nil
}

// Configure allows runtime tuning of engine parameters.
func (e *minimaxEngine) Configure(config MinimaxConfig) error {
	// Validate and apply search depth
	if config.SearchDepth != nil {
		if *config.SearchDepth < 1 || *config.SearchDepth > 20 {
			return fmt.Errorf("search depth must be 1-20, got %d", *config.SearchDepth)
		}
		e.maxDepth = *config.SearchDepth
	}

	// Validate and apply time limit
	if config.TimeLimit != nil {
		if *config.TimeLimit <= 0 {
			return fmt.Errorf("time limit must be positive, got %v", *config.TimeLimit)
		}
		e.timeLimit = *config.TimeLimit
	}

	// Apply evaluation weights (no validation needed)
	if config.MaterialWeight != nil {
		e.evalWeights.material = *config.MaterialWeight
	}
	if config.PieceSquareWeight != nil {
		e.evalWeights.pieceSquare = *config.PieceSquareWeight
	}
	if config.MobilityWeight != nil {
		e.evalWeights.mobility = *config.MobilityWeight
	}
	if config.KingSafetyWeight != nil {
		e.evalWeights.kingSafety = *config.KingSafetyWeight
	}

	return nil
}

// Info returns metadata about this engine.
func (e *minimaxEngine) Info() Info {
	return Info{
		Name:       e.name,
		Author:     "TermChess",
		Version:    "1.0",
		Type:       TypeInternal,
		Difficulty: e.difficulty,
		Features: map[string]bool{
			"alpha_beta":          true,
			"iterative_deepening": true,
			"move_ordering":       true,
			"configurable":        true,
			"piece_square_tables": e.difficulty >= Medium,
			"mobility":            e.difficulty >= Medium,
			"king_safety":         e.difficulty >= Hard,
		},
	}
}

// SelectMove returns the best move found by minimax search.
func (e *minimaxEngine) SelectMove(ctx context.Context, board *engine.Board) (engine.Move, error) {
	if e.closed {
		return engine.Move{}, errors.New("engine is closed")
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, e.timeLimit)
	defer cancel()

	// Get all legal moves
	moves := board.LegalMoves()
	if len(moves) == 0 {
		return engine.Move{}, errors.New("no legal moves available")
	}

	// If only one move, return it immediately (forced move)
	if len(moves) == 1 {
		return moves[0], nil
	}

	// Iterative deepening: start at depth 1, increment to maxDepth
	var bestMove engine.Move

	for depth := 1; depth <= e.maxDepth; depth++ {
		// Check timeout before starting new depth
		select {
		case <-ctx.Done():
			// Timeout reached, return best move from previous iteration
			if bestMove == (engine.Move{}) {
				// Fallback: no iteration completed, return first legal move
				return moves[0], nil
			}
			return bestMove, nil
		default:
		}

		// Search at current depth
		move, _, err := e.searchDepth(ctx, board, depth)
		if err != nil {
			// Timeout during search, return best move found so far
			if bestMove == (engine.Move{}) {
				return moves[0], nil
			}
			return bestMove, nil
		}

		// Update best move from this completed iteration
		bestMove = move
	}

	// All depths completed within timeout
	return bestMove, nil
}

// searchDepth performs a minimax search at a specific depth.
func (e *minimaxEngine) searchDepth(ctx context.Context, board *engine.Board, depth int) (engine.Move, float64, error) {
	moves := board.LegalMoves()
	if len(moves) == 0 {
		return engine.Move{}, 0, errors.New("no legal moves available")
	}

	// Order moves to improve alpha-beta pruning
	moves = e.orderMoves(board, moves)

	// Initialize alpha-beta bounds
	alpha := math.Inf(-1)
	beta := math.Inf(1)

	var bestMove engine.Move
	bestScore := math.Inf(-1)

	// Search each move
	for _, move := range moves {
		// Check for timeout
		select {
		case <-ctx.Done():
			return engine.Move{}, 0, ctx.Err()
		default:
		}

		// Make the move on a copy
		boardCopy := board.Copy()
		err := boardCopy.MakeMove(move)
		if err != nil {
			// This should not happen with legal moves, but handle it anyway
			continue
		}

		// Search with negamax (negate the score since we switched sides)
		// Pass ply=1 since we're one move from the root
		score := -e.alphaBeta(ctx, boardCopy, depth-1, -beta, -alpha, 1)

		// Update best move
		if score > bestScore {
			bestScore = score
			bestMove = move
		}

		// Update alpha
		if score > alpha {
			alpha = score
		}

		// Beta cutoff (this shouldn't happen at root with alpha=-inf, beta=+inf)
		if alpha >= beta {
			break
		}
	}

	return bestMove, bestScore, nil
}

// alphaBeta performs recursive negamax search with alpha-beta pruning.
// Returns the score from the perspective of the side to move.
// ply is the distance from the root (0 at root, increments with each recursive call).
func (e *minimaxEngine) alphaBeta(ctx context.Context, board *engine.Board, depth int, alpha, beta float64, ply int) float64 {
	// Check for timeout at the start of each node
	select {
	case <-ctx.Done():
		// Return a neutral score on timeout
		// This ensures the partial search doesn't corrupt the iterative deepening results
		return 0.0
	default:
	}

	// Base case: reached depth 0 or game over
	if depth == 0 || board.IsGameOver() {
		// Evaluate from White's perspective, then adjust for current player
		whiteScore := evaluate(board, e.difficulty)

		// Adjust mate scores to prefer faster mates
		// Mate in 1 ply scores higher than mate in 3 ply
		if whiteScore >= 9999.0 {
			// White wins - prefer faster mate
			whiteScore = whiteScore - float64(ply)
		} else if whiteScore <= -9999.0 {
			// Black wins - prefer faster mate (more negative = worse for us)
			whiteScore = whiteScore + float64(ply)
		}

		// Negamax: flip score if Black is to move
		if board.ActiveColor == engine.Black {
			return -whiteScore
		}
		return whiteScore
	}

	// Get all legal moves
	moves := board.LegalMoves()
	if len(moves) == 0 {
		// No legal moves means checkmate or stalemate
		// evaluate() already handles this, so just evaluate
		whiteScore := evaluate(board, e.difficulty)

		// Adjust mate scores to prefer faster mates
		if whiteScore >= 9999.0 {
			whiteScore = whiteScore - float64(ply)
		} else if whiteScore <= -9999.0 {
			whiteScore = whiteScore + float64(ply)
		}

		if board.ActiveColor == engine.Black {
			return -whiteScore
		}
		return whiteScore
	}

	// Order moves for better pruning
	moves = e.orderMoves(board, moves)

	// Negamax with alpha-beta pruning
	maxScore := math.Inf(-1)

	for _, move := range moves {
		// Make the move on a copy
		boardCopy := board.Copy()
		err := boardCopy.MakeMove(move)
		if err != nil {
			continue
		}

		// Recursive search with negated alpha-beta bounds
		score := -e.alphaBeta(ctx, boardCopy, depth-1, -beta, -alpha, ply+1)

		// Update max score
		if score > maxScore {
			maxScore = score
		}

		// Update alpha
		if score > alpha {
			alpha = score
		}

		// Beta cutoff (pruning)
		if alpha >= beta {
			break
		}
	}

	return maxScore
}

// orderMoves implements simple move ordering (captures first) to improve alpha-beta pruning.
// This is a basic MVV-LVA (Most Valuable Victim - Least Valuable Attacker) implementation.
func (e *minimaxEngine) orderMoves(board *engine.Board, moves []engine.Move) []engine.Move {
	// Separate captures from non-captures
	var captures []engine.Move
	var nonCaptures []engine.Move

	for _, move := range moves {
		targetPiece := board.PieceAt(move.To)
		if !targetPiece.IsEmpty() {
			captures = append(captures, move)
		} else {
			nonCaptures = append(nonCaptures, move)
		}
	}

	// Return captures first, then non-captures
	// Future enhancement: sort captures by MVV-LVA value
	return append(captures, nonCaptures...)
}
