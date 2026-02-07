package bvb

import (
	"math"
	"testing"
	"time"

	"github.com/Mgrdich/TermChess/internal/bot"
	"github.com/Mgrdich/TermChess/internal/engine"
)

func TestComputeStatsWinCounts(t *testing.T) {
	results := []GameResult{
		{GameNumber: 1, Winner: "White Bot", WinnerColor: engine.White, MoveCount: 30, Duration: 5 * time.Second, EndReason: "checkmate"},
		{GameNumber: 2, Winner: "White Bot", WinnerColor: engine.White, MoveCount: 25, Duration: 4 * time.Second, EndReason: "checkmate"},
		{GameNumber: 3, Winner: "White Bot", WinnerColor: engine.White, MoveCount: 35, Duration: 6 * time.Second, EndReason: "checkmate"},
		{GameNumber: 4, Winner: "Black Bot", WinnerColor: engine.Black, MoveCount: 40, Duration: 7 * time.Second, EndReason: "checkmate"},
		{GameNumber: 5, Winner: "Black Bot", WinnerColor: engine.Black, MoveCount: 50, Duration: 9 * time.Second, EndReason: "checkmate"},
		{GameNumber: 6, Winner: "Draw", MoveCount: 60, Duration: 10 * time.Second, EndReason: "stalemate"},
	}

	stats := ComputeStats(results, "White Bot", "Black Bot")

	if stats.TotalGames != 6 {
		t.Errorf("TotalGames = %d, want 6", stats.TotalGames)
	}
	if stats.WhiteWins != 3 {
		t.Errorf("WhiteWins = %d, want 3", stats.WhiteWins)
	}
	if stats.BlackWins != 2 {
		t.Errorf("BlackWins = %d, want 2", stats.BlackWins)
	}
	if stats.Draws != 1 {
		t.Errorf("Draws = %d, want 1", stats.Draws)
	}

	expectedWhitePct := 50.0
	if math.Abs(stats.WhiteWinPct-expectedWhitePct) > 0.01 {
		t.Errorf("WhiteWinPct = %f, want %f", stats.WhiteWinPct, expectedWhitePct)
	}

	expectedBlackPct := 100.0 * 2.0 / 6.0
	if math.Abs(stats.BlackWinPct-expectedBlackPct) > 0.01 {
		t.Errorf("BlackWinPct = %f, want %f", stats.BlackWinPct, expectedBlackPct)
	}

	if stats.WhiteBotName != "White Bot" {
		t.Errorf("WhiteBotName = %q, want %q", stats.WhiteBotName, "White Bot")
	}
	if stats.BlackBotName != "Black Bot" {
		t.Errorf("BlackBotName = %q, want %q", stats.BlackBotName, "Black Bot")
	}
}

func TestComputeStatsAllDraws(t *testing.T) {
	results := []GameResult{
		{GameNumber: 1, Winner: "Draw", MoveCount: 50, Duration: 8 * time.Second, EndReason: "stalemate"},
		{GameNumber: 2, Winner: "Draw", MoveCount: 55, Duration: 9 * time.Second, EndReason: "stalemate"},
		{GameNumber: 3, Winner: "Draw", MoveCount: 60, Duration: 10 * time.Second, EndReason: "move limit exceeded"},
	}

	stats := ComputeStats(results, "Alpha", "Beta")

	if stats.TotalGames != 3 {
		t.Errorf("TotalGames = %d, want 3", stats.TotalGames)
	}
	if stats.Draws != 3 {
		t.Errorf("Draws = %d, want 3", stats.Draws)
	}
	if stats.WhiteWins != 0 {
		t.Errorf("WhiteWins = %d, want 0", stats.WhiteWins)
	}
	if stats.BlackWins != 0 {
		t.Errorf("BlackWins = %d, want 0", stats.BlackWins)
	}
	if stats.WhiteWinPct != 0 {
		t.Errorf("WhiteWinPct = %f, want 0", stats.WhiteWinPct)
	}
	if stats.BlackWinPct != 0 {
		t.Errorf("BlackWinPct = %f, want 0", stats.BlackWinPct)
	}
}

func TestComputeStatsAverages(t *testing.T) {
	results := []GameResult{
		{GameNumber: 1, Winner: "White Bot", MoveCount: 10, Duration: 2 * time.Second, EndReason: "checkmate"},
		{GameNumber: 2, Winner: "Black Bot", MoveCount: 20, Duration: 4 * time.Second, EndReason: "checkmate"},
		{GameNumber: 3, Winner: "Draw", MoveCount: 30, Duration: 6 * time.Second, EndReason: "stalemate"},
	}

	stats := ComputeStats(results, "White Bot", "Black Bot")

	expectedAvgMoves := 20.0
	if math.Abs(stats.AvgMoveCount-expectedAvgMoves) > 0.01 {
		t.Errorf("AvgMoveCount = %f, want %f", stats.AvgMoveCount, expectedAvgMoves)
	}

	expectedAvgDuration := 4 * time.Second
	if stats.AvgDuration != expectedAvgDuration {
		t.Errorf("AvgDuration = %v, want %v", stats.AvgDuration, expectedAvgDuration)
	}
}

func TestComputeStatsShortestLongest(t *testing.T) {
	results := []GameResult{
		{GameNumber: 1, Winner: "White Bot", MoveCount: 40, Duration: 5 * time.Second, EndReason: "checkmate"},
		{GameNumber: 2, Winner: "Black Bot", MoveCount: 15, Duration: 3 * time.Second, EndReason: "checkmate"},
		{GameNumber: 3, Winner: "Draw", MoveCount: 80, Duration: 12 * time.Second, EndReason: "stalemate"},
		{GameNumber: 4, Winner: "White Bot", MoveCount: 55, Duration: 7 * time.Second, EndReason: "checkmate"},
	}

	stats := ComputeStats(results, "White Bot", "Black Bot")

	if stats.ShortestGame.GameNumber != 2 {
		t.Errorf("ShortestGame.GameNumber = %d, want 2", stats.ShortestGame.GameNumber)
	}
	if stats.ShortestGame.MoveCount != 15 {
		t.Errorf("ShortestGame.MoveCount = %d, want 15", stats.ShortestGame.MoveCount)
	}

	if stats.LongestGame.GameNumber != 3 {
		t.Errorf("LongestGame.GameNumber = %d, want 3", stats.LongestGame.GameNumber)
	}
	if stats.LongestGame.MoveCount != 80 {
		t.Errorf("LongestGame.MoveCount = %d, want 80", stats.LongestGame.MoveCount)
	}
}

func TestComputeStatsSingleGame(t *testing.T) {
	results := []GameResult{
		{GameNumber: 1, Winner: "White Bot", WinnerColor: engine.White, MoveCount: 42, Duration: 7 * time.Second, EndReason: "checkmate"},
	}

	stats := ComputeStats(results, "White Bot", "Black Bot")

	if stats.TotalGames != 1 {
		t.Errorf("TotalGames = %d, want 1", stats.TotalGames)
	}
	if stats.WhiteWins != 1 {
		t.Errorf("WhiteWins = %d, want 1", stats.WhiteWins)
	}
	if stats.BlackWins != 0 {
		t.Errorf("BlackWins = %d, want 0", stats.BlackWins)
	}
	if stats.Draws != 0 {
		t.Errorf("Draws = %d, want 0", stats.Draws)
	}
	if stats.WhiteWinPct != 100 {
		t.Errorf("WhiteWinPct = %f, want 100", stats.WhiteWinPct)
	}
	if stats.BlackWinPct != 0 {
		t.Errorf("BlackWinPct = %f, want 0", stats.BlackWinPct)
	}
	if stats.AvgMoveCount != 42 {
		t.Errorf("AvgMoveCount = %f, want 42", stats.AvgMoveCount)
	}
	if stats.AvgDuration != 7*time.Second {
		t.Errorf("AvgDuration = %v, want %v", stats.AvgDuration, 7*time.Second)
	}

	// Shortest and longest should both be the single game.
	if stats.ShortestGame.GameNumber != 1 {
		t.Errorf("ShortestGame.GameNumber = %d, want 1", stats.ShortestGame.GameNumber)
	}
	if stats.LongestGame.GameNumber != 1 {
		t.Errorf("LongestGame.GameNumber = %d, want 1", stats.LongestGame.GameNumber)
	}
	if stats.ShortestGame.MoveCount != stats.LongestGame.MoveCount {
		t.Errorf("ShortestGame.MoveCount (%d) != LongestGame.MoveCount (%d) for single game",
			stats.ShortestGame.MoveCount, stats.LongestGame.MoveCount)
	}
}

func TestComputeStatsEmpty(t *testing.T) {
	stats := ComputeStats(nil, "White Bot", "Black Bot")

	if stats == nil {
		t.Fatal("ComputeStats returned nil for empty results")
	}
	if stats.TotalGames != 0 {
		t.Errorf("TotalGames = %d, want 0", stats.TotalGames)
	}
	if stats.WhiteWins != 0 {
		t.Errorf("WhiteWins = %d, want 0", stats.WhiteWins)
	}
	if stats.BlackWins != 0 {
		t.Errorf("BlackWins = %d, want 0", stats.BlackWins)
	}
	if stats.Draws != 0 {
		t.Errorf("Draws = %d, want 0", stats.Draws)
	}
	if stats.WhiteWinPct != 0 {
		t.Errorf("WhiteWinPct = %f, want 0", stats.WhiteWinPct)
	}
	if stats.BlackWinPct != 0 {
		t.Errorf("BlackWinPct = %f, want 0", stats.BlackWinPct)
	}
	if stats.AvgMoveCount != 0 {
		t.Errorf("AvgMoveCount = %f, want 0", stats.AvgMoveCount)
	}
	if stats.AvgDuration != 0 {
		t.Errorf("AvgDuration = %v, want 0", stats.AvgDuration)
	}
	if stats.WhiteBotName != "White Bot" {
		t.Errorf("WhiteBotName = %q, want %q", stats.WhiteBotName, "White Bot")
	}
	if stats.BlackBotName != "Black Bot" {
		t.Errorf("BlackBotName = %q, want %q", stats.BlackBotName, "Black Bot")
	}
	if stats.IndividualResults != nil {
		t.Errorf("IndividualResults should be nil for empty results, got length %d", len(stats.IndividualResults))
	}
}

func TestComputeStatsEmptySlice(t *testing.T) {
	stats := ComputeStats([]GameResult{}, "A", "B")

	if stats == nil {
		t.Fatal("ComputeStats returned nil for empty slice")
	}
	if stats.TotalGames != 0 {
		t.Errorf("TotalGames = %d, want 0", stats.TotalGames)
	}
	if stats.WhiteBotName != "A" {
		t.Errorf("WhiteBotName = %q, want %q", stats.WhiteBotName, "A")
	}
	if stats.BlackBotName != "B" {
		t.Errorf("BlackBotName = %q, want %q", stats.BlackBotName, "B")
	}
}

func TestComputeStatsIndividualResults(t *testing.T) {
	results := []GameResult{
		{GameNumber: 1, Winner: "White Bot", MoveCount: 30, Duration: 5 * time.Second, EndReason: "checkmate"},
		{GameNumber: 2, Winner: "Black Bot", MoveCount: 45, Duration: 8 * time.Second, EndReason: "checkmate"},
	}

	stats := ComputeStats(results, "White Bot", "Black Bot")

	if len(stats.IndividualResults) != 2 {
		t.Fatalf("IndividualResults length = %d, want 2", len(stats.IndividualResults))
	}

	// Verify the results are copies (not the same slice).
	if stats.IndividualResults[0].GameNumber != 1 {
		t.Errorf("IndividualResults[0].GameNumber = %d, want 1", stats.IndividualResults[0].GameNumber)
	}
	if stats.IndividualResults[1].GameNumber != 2 {
		t.Errorf("IndividualResults[1].GameNumber = %d, want 2", stats.IndividualResults[1].GameNumber)
	}

	// Modify original to ensure it is a copy.
	results[0].GameNumber = 99
	if stats.IndividualResults[0].GameNumber == 99 {
		t.Error("IndividualResults should be a copy, not a reference to the original slice")
	}
}

func TestSessionManagerStats(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "Easy White", "Easy Black", 3, 0)
	m.speed = SpeedInstant
	err := m.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	// Wait for all games to finish.
	deadline := time.After(60 * time.Second)
	for !m.AllFinished() {
		select {
		case <-deadline:
			m.Abort()
			t.Fatal("games did not complete within timeout")
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}

	stats := m.Stats()
	if stats == nil {
		t.Fatal("Stats() returned nil")
	}

	if stats.TotalGames != 3 {
		t.Errorf("TotalGames = %d, want 3", stats.TotalGames)
	}

	if stats.WhiteBotName != "Easy White" {
		t.Errorf("WhiteBotName = %q, want %q", stats.WhiteBotName, "Easy White")
	}
	if stats.BlackBotName != "Easy Black" {
		t.Errorf("BlackBotName = %q, want %q", stats.BlackBotName, "Easy Black")
	}

	// Total outcomes should sum to TotalGames.
	totalOutcomes := stats.WhiteWins + stats.BlackWins + stats.Draws
	if totalOutcomes != stats.TotalGames {
		t.Errorf("WhiteWins(%d) + BlackWins(%d) + Draws(%d) = %d, want %d",
			stats.WhiteWins, stats.BlackWins, stats.Draws, totalOutcomes, stats.TotalGames)
	}

	if stats.AvgMoveCount <= 0 {
		t.Errorf("AvgMoveCount = %f, want > 0", stats.AvgMoveCount)
	}

	if stats.AvgDuration <= 0 {
		t.Errorf("AvgDuration = %v, want > 0", stats.AvgDuration)
	}

	if stats.ShortestGame.MoveCount <= 0 {
		t.Errorf("ShortestGame.MoveCount = %d, want > 0", stats.ShortestGame.MoveCount)
	}

	if stats.LongestGame.MoveCount <= 0 {
		t.Errorf("LongestGame.MoveCount = %d, want > 0", stats.LongestGame.MoveCount)
	}

	if stats.ShortestGame.MoveCount > stats.LongestGame.MoveCount {
		t.Errorf("ShortestGame.MoveCount (%d) > LongestGame.MoveCount (%d)",
			stats.ShortestGame.MoveCount, stats.LongestGame.MoveCount)
	}

	if len(stats.IndividualResults) != 3 {
		t.Errorf("IndividualResults length = %d, want 3", len(stats.IndividualResults))
	}

	// Each individual result should have valid data.
	for i, r := range stats.IndividualResults {
		if r.MoveCount <= 0 {
			t.Errorf("IndividualResults[%d].MoveCount = %d, want > 0", i, r.MoveCount)
		}
		if r.Winner == "" {
			t.Errorf("IndividualResults[%d].Winner is empty", i)
		}
		if r.EndReason == "" {
			t.Errorf("IndividualResults[%d].EndReason is empty", i)
		}
	}
}

func TestSessionManagerStatsBeforeStart(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "W", "B", 3, 0)

	stats := m.Stats()
	if stats == nil {
		t.Fatal("Stats() returned nil before Start")
	}
	if stats.TotalGames != 0 {
		t.Errorf("TotalGames = %d, want 0 before Start", stats.TotalGames)
	}
	if stats.WhiteBotName != "W" {
		t.Errorf("WhiteBotName = %q, want %q", stats.WhiteBotName, "W")
	}
	if stats.BlackBotName != "B" {
		t.Errorf("BlackBotName = %q, want %q", stats.BlackBotName, "B")
	}
}
