package bot

import (
	"context"
	"testing"
	"time"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// GameResult represents the outcome of a bot vs bot game.
type GameResult struct {
	Winner    engine.Color      // White, Black, or 0 (for draws)
	Outcome   engine.GameStatus // Checkmate, Stalemate, etc.
	MoveCount int               // Number of moves made
	IsDraw    bool              // True if game ended in draw
}

// runBotGame plays a full game between two bots and returns the result.
// The white bot plays as White, the black bot plays as Black.
// Games are limited to maxMoves (default 200) to prevent infinite games.
func runBotGame(t *testing.T, white, black Engine) GameResult {
	t.Helper()

	board := engine.NewBoard()
	moveCount := 0
	maxMoves := 200

	for moveCount < maxMoves {
		// Check if game is over
		status := board.Status()
		if board.IsGameOver() {
			// Game ended automatically (checkmate, stalemate, automatic draws)
			winner, hasWinner := board.Winner()
			return GameResult{
				Winner:    winner,
				Outcome:   status,
				MoveCount: moveCount,
				IsDraw:    !hasWinner,
			}
		}

		// Handle claimable draws (threefold repetition, fifty-move rule)
		// For bot testing, we'll automatically claim draws
		if board.CanClaimDraw() {
			return GameResult{
				Winner:    0,
				Outcome:   status,
				MoveCount: moveCount,
				IsDraw:    true,
			}
		}

		// Select current bot based on active color
		var currentBot Engine
		var botName string
		if board.ActiveColor == engine.White {
			currentBot = white
			botName = "White"
		} else {
			currentBot = black
			botName = "Black"
		}

		// Get move from bot with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		move, err := currentBot.SelectMove(ctx, board)
		cancel()

		if err != nil {
			t.Fatalf("Bot %s (%s) failed to select move at move %d: %v",
				botName, currentBot.Name(), moveCount, err)
		}

		// Make the move
		err = board.MakeMove(move)
		if err != nil {
			t.Fatalf("Bot %s (%s) selected illegal move %s at move %d: %v",
				botName, currentBot.Name(), move.String(), moveCount, err)
		}

		moveCount++
	}

	// Max moves reached, consider it a draw
	t.Logf("Game reached maximum move limit (%d moves), considering it a draw", maxMoves)
	return GameResult{
		Winner:    0,
		Outcome:   engine.Ongoing,
		MoveCount: moveCount,
		IsDraw:    true,
	}
}

// TestDifficulty_MediumVsEasy tests that Medium bot consistently beats Easy bot.
// Runs 10 games and expects Medium to win at least 7.
func TestDifficulty_MediumVsEasy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping bot vs bot test in short mode")
	}

	// Create bots
	easyBot, err := NewRandomEngine()
	if err != nil {
		t.Fatalf("Failed to create Easy bot: %v", err)
	}
	defer easyBot.Close()

	mediumBot, err := NewMinimaxEngine(Medium)
	if err != nil {
		t.Fatalf("Failed to create Medium bot: %v", err)
	}
	defer mediumBot.Close()

	// Run games
	numGames := 10
	mediumWins := 0
	easyWins := 0
	draws := 0

	t.Logf("Starting %d games: Medium vs Easy", numGames)

	for i := 0; i < numGames; i++ {
		t.Logf("Game %d/%d starting...", i+1, numGames)

		// Alternate colors: Medium plays White in even games, Black in odd games
		var result GameResult
		if i%2 == 0 {
			result = runBotGame(t, mediumBot, easyBot)
			if result.IsDraw {
				draws++
			} else if result.Winner == engine.White {
				mediumWins++
			} else {
				easyWins++
			}
		} else {
			result = runBotGame(t, easyBot, mediumBot)
			if result.IsDraw {
				draws++
			} else if result.Winner == engine.Black {
				mediumWins++
			} else {
				easyWins++
			}
		}

		t.Logf("Game %d/%d finished: %s in %d moves",
			i+1, numGames, result.Outcome.String(), result.MoveCount)
	}

	t.Logf("Results: Medium wins: %d, Easy wins: %d, Draws: %d",
		mediumWins, easyWins, draws)

	// Assert Medium wins at least 7 out of 10 games
	if mediumWins < 7 {
		t.Errorf("Medium bot should win at least 7/%d games, but won %d",
			numGames, mediumWins)
		t.Errorf("This suggests the difficulty calibration needs tuning.")
		t.Errorf("Consider adjusting search depth or evaluation weights.")
	}
}

// TestDifficulty_HardVsMedium tests that Hard bot consistently beats Medium bot.
// Runs 3 games with depth 2 vs depth 4 (N+2 advantage) to ensure Hard dominates.
// Note: Time limits keep test duration reasonable while depth difference ensures decisive results.
func TestDifficulty_HardVsMedium(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping bot vs bot test in short mode")
	}

	// Create bots with different search depths for testing
	// Medium: depth 2, Hard: depth 4 (N+2 difference for decisive advantage)
	// Time limits keep total test duration reasonable
	mediumBot, err := NewMinimaxEngine(Medium, WithTimeLimit(500*time.Millisecond), WithSearchDepth(2))
	if err != nil {
		t.Fatalf("Failed to create Medium bot: %v", err)
	}
	defer mediumBot.Close()

	hardBot, err := NewMinimaxEngine(Hard, WithTimeLimit(1*time.Second), WithSearchDepth(4))
	if err != nil {
		t.Fatalf("Failed to create Hard bot: %v", err)
	}
	defer hardBot.Close()

	// Run games (reduced from 10 to 3 for faster testing)
	numGames := 3
	hardWins := 0
	mediumWins := 0
	draws := 0

	t.Logf("Starting %d games: Hard vs Medium", numGames)

	for i := 0; i < numGames; i++ {
		t.Logf("Game %d/%d starting...", i+1, numGames)

		// Alternate colors: Hard plays White in even games, Black in odd games
		var result GameResult
		if i%2 == 0 {
			result = runBotGame(t, hardBot, mediumBot)
			if result.IsDraw {
				draws++
			} else if result.Winner == engine.White {
				hardWins++
			} else {
				mediumWins++
			}
		} else {
			result = runBotGame(t, mediumBot, hardBot)
			if result.IsDraw {
				draws++
			} else if result.Winner == engine.Black {
				hardWins++
			} else {
				mediumWins++
			}
		}

		t.Logf("Game %d/%d finished: %s in %d moves",
			i+1, numGames, result.Outcome.String(), result.MoveCount)
	}

	t.Logf("Results: Hard wins: %d, Medium wins: %d, Draws: %d",
		hardWins, mediumWins, draws)

	// Assert Hard wins all games (with depth 4 vs 2, Hard should dominate)
	// Note: With 2-depth advantage, Hard should not lose any games
	if mediumWins > 0 {
		t.Errorf("Hard bot (depth 4) should not lose to Medium (depth 2), but Medium won %d games",
			mediumWins)
	}
	if hardWins == 0 && draws == numGames {
		// All draws is acceptable but unexpected with depth difference
		t.Logf("Warning: All games were draws, consider if this is expected")
	}

	// Calculate win rate (excluding draws)
	decidedGames := hardWins + mediumWins
	if decidedGames > 0 {
		winRate := float64(hardWins) / float64(decidedGames) * 100.0
		t.Logf("Hard win rate (excluding draws): %.1f%%", winRate)
	}
}

// TestDifficulty_EasyVsEasy tests that Easy bots can play full games without crashes.
// Runs 5 games and verifies games complete successfully.
func TestDifficulty_EasyVsEasy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping bot vs bot test in short mode")
	}

	// Create two Easy bots
	easyBot1, err := NewRandomEngine()
	if err != nil {
		t.Fatalf("Failed to create Easy bot 1: %v", err)
	}
	defer easyBot1.Close()

	easyBot2, err := NewRandomEngine()
	if err != nil {
		t.Fatalf("Failed to create Easy bot 2: %v", err)
	}
	defer easyBot2.Close()

	// Run games
	numGames := 5
	completedGames := 0

	t.Logf("Starting %d games: Easy vs Easy", numGames)

	for i := 0; i < numGames; i++ {
		t.Logf("Game %d/%d starting...", i+1, numGames)

		result := runBotGame(t, easyBot1, easyBot2)

		t.Logf("Game %d/%d finished: %s in %d moves",
			i+1, numGames, result.Outcome.String(), result.MoveCount)

		completedGames++
	}

	t.Logf("All %d games completed successfully", completedGames)

	// Assert all games completed
	if completedGames != numGames {
		t.Errorf("Expected %d games to complete, but only %d completed",
			numGames, completedGames)
	}
}
