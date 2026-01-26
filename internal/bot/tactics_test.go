package bot

import (
	"context"
	"testing"
	"time"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// loadFEN is a helper function that loads a board position from a FEN string.
// It fails the test if the FEN is invalid.
func loadFEN(t *testing.T, fenString string) *engine.Board {
	t.Helper()
	board, err := engine.FromFEN(fenString)
	if err != nil {
		t.Fatalf("failed to load FEN %q: %v", fenString, err)
	}
	return board
}

// testTacticalPuzzle is a helper that tests if a bot finds the correct tactical move.
// It verifies that the bot selects one of the expected moves (if multiple solutions exist).
func testTacticalPuzzle(t *testing.T, difficulty Difficulty, fen string, expectedMoves []string, description string) {
	t.Helper()

	board := loadFEN(t, fen)

	var bot Engine
	var err error

	// Use increased search depth for tactical tests to ensure correct solution is found
	switch difficulty {
	case Medium:
		bot, err = NewMinimaxEngine(Medium, WithSearchDepth(5))
	case Hard:
		bot, err = NewMinimaxEngine(Hard, WithSearchDepth(6))
	default:
		t.Fatalf("invalid difficulty: %v", difficulty)
	}

	if err != nil {
		t.Fatalf("failed to create bot: %v", err)
	}
	defer bot.Close()

	// Try up to 3 times to account for random tie-breaking
	maxAttempts := 3
	var lastMove string
	found := false

	for attempt := 0; attempt < maxAttempts && !found; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		move, err := bot.SelectMove(ctx, board)
		cancel()

		if err != nil {
			t.Fatalf("SelectMove() error = %v", err)
		}

		lastMove = move.String()
		for _, expected := range expectedMoves {
			if lastMove == expected {
				found = true
				break
			}
		}
	}

	if !found {
		t.Errorf("%s bot should find tactical move %v: %s, but found %s (after %d attempts)",
			difficulty.String(), expectedMoves, description, lastMove, maxAttempts)
	}
}

// testMateDelivery verifies that the selected move delivers checkmate.
func testMateDelivery(t *testing.T, difficulty Difficulty, fen string, description string) {
	t.Helper()

	board := loadFEN(t, fen)

	var bot Engine
	var err error

	switch difficulty {
	case Medium:
		bot, err = NewMinimaxEngine(Medium)
	case Hard:
		bot, err = NewMinimaxEngine(Hard)
	default:
		t.Fatalf("invalid difficulty: %v", difficulty)
	}

	if err != nil {
		t.Fatalf("failed to create bot: %v", err)
	}
	defer bot.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	move, err := bot.SelectMove(ctx, board)
	if err != nil {
		t.Fatalf("SelectMove() error = %v", err)
	}

	// Apply the move and verify it's checkmate
	boardCopy := board.Copy()
	err = boardCopy.MakeMove(move)
	if err != nil {
		t.Fatalf("MakeMove() error = %v", err)
	}

	if boardCopy.Status() != engine.Checkmate {
		t.Errorf("%s bot should find mate-in-1: %s, got move %s with status %v",
			difficulty.String(), description, move.String(), boardCopy.Status())
	}
}

// TestTactical_MateInOne tests that bots can find mate-in-1 positions.
func TestTactical_MateInOne(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		description string
	}{
		{
			name: "back rank mate",
			// White rook on a1, Black king trapped on g8 by its own pawns
			fen:         "6k1/5ppp/8/8/8/8/8/R6K w - - 0 1",
			description: "Ra8# delivers back rank mate",
		},
		{
			name: "queen and bishop mate",
			// White queen and bishop deliver mate
			// White queen on f6, bishop on b2, Black king on h8
			fen:         "7k/8/5Q2/8/8/8/1B6/7K w - - 0 1",
			description: "Qg7# or Qf7 delivers checkmate",
		},
		{
			name: "two rooks mate",
			// White rooks on a7 and b1, Black king on h8
			fen:         "7k/R7/8/8/8/8/8/1R5K w - - 0 1",
			description: "Ra8# or Rb8# delivers checkmate",
		},
		{
			name: "knight and queen mate",
			// White queen on g6 and knight on f6, Black king on h8
			fen:         "7k/8/5NQ1/8/8/8/8/7K w - - 0 1",
			description: "Qg7# or Qh7# delivers checkmate",
		},
		{
			name: "smothered mate",
			// White knight delivers smothered mate, Black king trapped by its own pieces
			// Black king on h8, Black rook on g8, Black pawns on g7 and h7, White knight on f7
			fen:         "6rk/5Npp/8/8/8/8/8/7K w - - 0 1",
			description: "Nh6# or Ng5 delivers checkmate (smothered mate pattern)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("medium bot", func(t *testing.T) {
				testMateDelivery(t, Medium, tt.fen, tt.description)
			})

			t.Run("hard bot", func(t *testing.T) {
				testMateDelivery(t, Hard, tt.fen, tt.description)
			})
		})
	}
}

// TestTactical_MateInTwo tests that bots can find the first move in mate-in-2 sequences.
func TestTactical_MateInTwo(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		firstMoves  []string // Acceptable first moves that lead to mate in 2
		description string
	}{
		{
			name: "queen sacrifice mate in 2",
			// White queen and knight coordinate for mate
			// White queen on d4, knight on f6, Black king on h8, Black rook on g8
			fen:         "6rk/8/5N2/8/3Q4/8/8/7K w - - 0 1",
			firstMoves:  []string{"d4h4", "d4h8", "f6g8"}, // Multiple winning moves
			description: "Queen/knight coordination forcing mate in 2",
		},
		{
			name: "rook mate in 2",
			// Rook penetration leading to mate
			// White rook on a7, Black king on h8, Black pawns blocking
			fen:         "7k/R4ppp/8/8/8/8/8/7K w - - 0 1",
			firstMoves:  []string{"a7a8"}, // Ra8+ followed by mate
			description: "Rook penetration forcing mate in 2",
		},
		{
			name: "queen and bishop mate in 2",
			// White queen and bishop coordinate
			// White queen on d4, bishop on b2, Black king on h8, pawns on g7 and h7
			fen:         "7k/6pp/8/8/3Q4/8/1B6/7K w - - 0 1",
			firstMoves:  []string{"d4g7", "d4h8", "d4h4"}, // Multiple mating patterns
			description: "Queen and bishop coordination forcing mate in 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("medium bot", func(t *testing.T) {
				testTacticalPuzzle(t, Medium, tt.fen, tt.firstMoves, tt.description)
			})

			t.Run("hard bot", func(t *testing.T) {
				testTacticalPuzzle(t, Hard, tt.fen, tt.firstMoves, tt.description)
			})
		})
	}
}

// TestTactical_Fork tests that bots can recognize and execute fork tactics.
func TestTactical_Fork(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		forkMoves   []string
		description string
	}{
		{
			name: "knight fork king and rook",
			// White knight can fork Black king and rook
			// Knight on e3, Black king on d5, Black rook on f5
			fen:         "8/8/8/3k1r2/8/4N3/8/7K w - - 0 1",
			forkMoves:   []string{"e3d5", "e3f5", "e3g4"}, // Several good moves
			description: "Knight forks king and rook",
		},
		{
			name: "pawn fork two rooks",
			// White pawn can capture one of two Black rooks (both are hanging)
			// Pawn on e4, Black rooks on d5 and f5 (both can be captured)
			fen:         "6k1/8/8/3r1r2/4P3/8/8/6K1 w - - 0 1",
			forkMoves:   []string{"e4e5", "e4d5", "e4f5"}, // Capture or advance (all win material)
			description: "Pawn attacks two rooks (fork pattern)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("medium bot", func(t *testing.T) {
				testTacticalPuzzle(t, Medium, tt.fen, tt.forkMoves, tt.description)
			})

			t.Run("hard bot", func(t *testing.T) {
				testTacticalPuzzle(t, Hard, tt.fen, tt.forkMoves, tt.description)
			})
		})
	}
}

// TestTactical_Pin tests that bots can exploit pin tactics.
func TestTactical_Pin(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		pinMoves    []string
		description string
	}{
		{
			name: "exploit absolute pin",
			// Black knight on d5 is pinned to king on d8 by White rook on d1
			// White can capture the pinned knight
			fen:         "3k4/8/8/3n4/8/8/8/3R2K1 w - - 0 1",
			pinMoves:    []string{"d1d5"}, // Capture the pinned knight
			description: "Capture pinned piece",
		},
		{
			name: "create and exploit pin",
			// White bishop can pin Black queen to king
			// White bishop on b2, Black queen on e5, Black king on h8
			fen:         "7k/8/8/4q3/8/8/1B6/7K w - - 0 1",
			pinMoves:    []string{"b2e5"}, // Capture the queen
			description: "Exploit diagonal pin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("medium bot", func(t *testing.T) {
				testTacticalPuzzle(t, Medium, tt.fen, tt.pinMoves, tt.description)
			})

			t.Run("hard bot", func(t *testing.T) {
				testTacticalPuzzle(t, Hard, tt.fen, tt.pinMoves, tt.description)
			})
		})
	}
}

// TestTactical_Skewer tests that bots can execute skewer tactics.
func TestTactical_Skewer(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		skewerMoves []string
		description string
	}{
		{
			name: "rook skewer king and queen",
			// White rook can check king, forcing it to move and exposing queen
			// White rook on a1, Black king on a8, Black queen on a7
			fen:         "k7/q7/8/8/8/8/8/R6K w - - 0 1",
			skewerMoves: []string{"a1a7", "a1a8"}, // Check or capture
			description: "Rook skewers king and queen",
		},
		{
			name: "bishop captures queen",
			// White bishop can simply capture the hanging Black queen
			// White bishop on c4, Black queen on f7, Black king on h8
			fen:         "7k/5q2/8/8/2B5/8/8/7K w - - 0 1",
			skewerMoves: []string{"c4f7"}, // Capture the queen
			description: "Bishop captures hanging queen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("medium bot", func(t *testing.T) {
				testTacticalPuzzle(t, Medium, tt.fen, tt.skewerMoves, tt.description)
			})

			t.Run("hard bot", func(t *testing.T) {
				testTacticalPuzzle(t, Hard, tt.fen, tt.skewerMoves, tt.description)
			})
		})
	}
}

// TestTactical_DiscoveredAttack tests that bots can find discovered attack tactics.
func TestTactical_DiscoveredAttack(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		attackMoves []string
		description string
	}{
		{
			name: "knight wins queen with check",
			// White knight can capture Black queen with check
			// Knight on e5, Black queen on d7, Black king on e8
			fen:         "4k3/3q4/8/4N3/8/8/8/7K w - - 0 1",
			attackMoves: []string{"e5d7"}, // Nxd7+ wins the queen with check
			description: "Knight captures queen with check",
		},
		{
			name: "discovered attack on queen",
			// White knight on e4 blocks White rook on e1 from Black queen on e8
			// Moving knight discovers attack on queen
			fen:         "4q3/8/8/8/4N3/8/8/4R2K w - - 0 1",
			attackMoves: []string{"e4d6", "e4f6", "e4g5", "e4g3", "e4f2", "e4d2", "e4c3", "e4c5"}, // Knight moves discovering rook attack
			description: "Discovered attack wins queen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("medium bot", func(t *testing.T) {
				testTacticalPuzzle(t, Medium, tt.fen, tt.attackMoves, tt.description)
			})

			t.Run("hard bot", func(t *testing.T) {
				testTacticalPuzzle(t, Hard, tt.fen, tt.attackMoves, tt.description)
			})
		})
	}
}

// TestTactical_DontHangQueen tests that bots don't blunder their queen.
func TestTactical_DontHangQueen(t *testing.T) {
	// Position where White queen on d1 can move, but moving to d8 hangs it to Black rook on d7
	fen := "4k3/3r4/8/8/8/8/8/3Q1K2 w - - 0 1"

	board := loadFEN(t, fen)

	// The blunder move (hanging queen)
	blunderMove, err := engine.ParseMove("d1d8")
	if err != nil {
		t.Fatalf("ParseMove() error = %v", err)
	}

	difficulties := []Difficulty{Medium, Hard}
	for _, difficulty := range difficulties {
		t.Run(difficulty.String()+" bot", func(t *testing.T) {
			var bot Engine
			var err error

			switch difficulty {
			case Medium:
				bot, err = NewMinimaxEngine(Medium)
			case Hard:
				bot, err = NewMinimaxEngine(Hard)
			}

			if err != nil {
				t.Fatalf("failed to create bot: %v", err)
			}
			defer bot.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			move, err := bot.SelectMove(ctx, board)
			if err != nil {
				t.Fatalf("SelectMove() error = %v", err)
			}

			if move == blunderMove {
				t.Errorf("%s bot should not hang queen with Qd8, but it did", difficulty.String())
			}
		})
	}
}

// TestTactical_DontHangRook tests that bots don't blunder their rook.
func TestTactical_DontHangRook(t *testing.T) {
	// Position where White rook on a1 can move to a8 but would be captured by Black queen on h8
	// This is a contrived position to test blunder avoidance
	fen := "7q/8/8/8/8/8/8/R6K w - - 0 1"

	board := loadFEN(t, fen)

	// The blunder move (moving rook to 8th rank where queen can capture)
	blunderMove, err := engine.ParseMove("a1a8")
	if err != nil {
		t.Fatalf("ParseMove() error = %v", err)
	}

	difficulties := []Difficulty{Medium, Hard}
	for _, difficulty := range difficulties {
		t.Run(difficulty.String()+" bot", func(t *testing.T) {
			var bot Engine
			var err error

			switch difficulty {
			case Medium:
				bot, err = NewMinimaxEngine(Medium)
			case Hard:
				bot, err = NewMinimaxEngine(Hard)
			}

			if err != nil {
				t.Fatalf("failed to create bot: %v", err)
			}
			defer bot.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			move, err := bot.SelectMove(ctx, board)
			if err != nil {
				t.Fatalf("SelectMove() error = %v", err)
			}

			if move == blunderMove {
				t.Errorf("%s bot should not hang rook with Ra8, but it did", difficulty.String())
			}
		})
	}
}

// TestTactical_DontAllowBackRankMate tests that bots defend against back rank mate threats.
func TestTactical_DontAllowBackRankMate(t *testing.T) {
	// Position where White king is on back rank with pawns in front
	// Black rook threatens back rank mate
	// White king on g1, pawns on f2, g2, h2, Black rook on e8, White has a rook on a1
	fen := "4r3/8/8/8/8/8/5PPP/R5K1 w - - 0 1"

	board := loadFEN(t, fen)

	difficulties := []Difficulty{Medium, Hard}
	for _, difficulty := range difficulties {
		t.Run(difficulty.String()+" bot", func(t *testing.T) {
			var bot Engine
			var err error

			switch difficulty {
			case Medium:
				bot, err = NewMinimaxEngine(Medium)
			case Hard:
				bot, err = NewMinimaxEngine(Hard)
			}

			if err != nil {
				t.Fatalf("failed to create bot: %v", err)
			}
			defer bot.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			move, err := bot.SelectMove(ctx, board)
			if err != nil {
				t.Fatalf("SelectMove() error = %v", err)
			}

			// Apply the move
			boardCopy := board.Copy()
			err = boardCopy.MakeMove(move)
			if err != nil {
				t.Fatalf("MakeMove() error = %v", err)
			}

			// Simulate Black's best response (would be checkmate if White didn't defend properly)
			// Black should deliver mate if possible
			blackMoves := boardCopy.LegalMoves()
			for _, blackMove := range blackMoves {
				testBoard := boardCopy.Copy()
				err = testBoard.MakeMove(blackMove)
				if err != nil {
					continue
				}

				if testBoard.Status() == engine.Checkmate {
					t.Errorf("%s bot should prevent back rank mate, but after %s, Black can deliver mate with %s",
						difficulty.String(), move.String(), blackMove.String())
					return
				}
			}
		})
	}
}
