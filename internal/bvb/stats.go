package bvb

import "time"

// AggregateStats holds computed statistics for a multi-game session.
type AggregateStats struct {
	// TotalGames is the number of completed games.
	TotalGames int
	// WhiteBotName is the name of the white bot.
	WhiteBotName string
	// BlackBotName is the name of the black bot.
	BlackBotName string
	// WhiteWins is the number of games won by the white bot.
	WhiteWins int
	// BlackWins is the number of games won by the black bot.
	BlackWins int
	// Draws is the number of drawn games.
	Draws int
	// WhiteWinPct is the white bot's win percentage (0-100).
	WhiteWinPct float64
	// BlackWinPct is the black bot's win percentage (0-100).
	BlackWinPct float64
	// AvgMoveCount is the average number of moves per game.
	AvgMoveCount float64
	// AvgDuration is the average game duration.
	AvgDuration time.Duration
	// ShortestGame is the game with the fewest moves.
	ShortestGame GameResult
	// LongestGame is the game with the most moves.
	LongestGame GameResult
	// IndividualResults contains all game results in order.
	IndividualResults []GameResult
}

// ComputeStats calculates aggregate statistics from a slice of game results.
func ComputeStats(results []GameResult, whiteName, blackName string) *AggregateStats {
	if len(results) == 0 {
		return &AggregateStats{
			WhiteBotName: whiteName,
			BlackBotName: blackName,
		}
	}

	stats := &AggregateStats{
		TotalGames:        len(results),
		WhiteBotName:      whiteName,
		BlackBotName:      blackName,
		IndividualResults: make([]GameResult, len(results)),
		ShortestGame:      results[0],
		LongestGame:       results[0],
	}
	copy(stats.IndividualResults, results)

	var totalMoves int
	var totalDuration time.Duration

	for _, r := range results {
		// Count wins.
		if r.Winner == "Draw" {
			stats.Draws++
		} else if r.Winner == whiteName {
			stats.WhiteWins++
		} else if r.Winner == blackName {
			stats.BlackWins++
		}

		// Accumulate for averages.
		totalMoves += r.MoveCount
		totalDuration += r.Duration

		// Track shortest/longest by move count.
		if r.MoveCount < stats.ShortestGame.MoveCount {
			stats.ShortestGame = r
		}
		if r.MoveCount > stats.LongestGame.MoveCount {
			stats.LongestGame = r
		}
	}

	// Calculate averages.
	stats.AvgMoveCount = float64(totalMoves) / float64(stats.TotalGames)
	stats.AvgDuration = totalDuration / time.Duration(stats.TotalGames)

	// Calculate win percentages.
	stats.WhiteWinPct = float64(stats.WhiteWins) / float64(stats.TotalGames) * 100
	stats.BlackWinPct = float64(stats.BlackWins) / float64(stats.TotalGames) * 100

	return stats
}
