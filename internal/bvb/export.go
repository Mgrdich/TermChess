package bvb

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SessionExport represents the complete export data for a Bot vs Bot session.
// This structure is serialized to JSON when exporting statistics.
type SessionExport struct {
	Timestamp    time.Time    `json:"timestamp"`
	WhiteBot     string       `json:"white_bot"`
	BlackBot     string       `json:"black_bot"`
	TotalGames   int          `json:"total_games"`
	WhiteWins    int          `json:"white_wins"`
	BlackWins    int          `json:"black_wins"`
	Draws        int          `json:"draws"`
	AverageMoves float64      `json:"average_moves"`
	Games        []GameExport `json:"games"`
}

// GameExport represents the export data for a single game.
type GameExport struct {
	GameNumber        int      `json:"game_number"`
	Result            string   `json:"result"`      // "White", "Black", "Draw"
	TerminationReason string   `json:"termination"` // "Checkmate", "Stalemate", etc.
	MoveCount         int      `json:"move_count"`
	Moves             []string `json:"moves"`     // Coordinate notation (e.g., "e2e4")
	FinalFEN          string   `json:"final_fen"` // Final position in FEN
}

// ExportStats generates a SessionExport from the SessionManager's completed games.
// It collects all game data and calculates aggregate statistics.
func (m *SessionManager) ExportStats(whiteBot, blackBot string) *SessionExport {
	m.mu.Lock()
	defer m.mu.Unlock()

	export := &SessionExport{
		Timestamp: time.Now(),
		WhiteBot:  whiteBot,
		BlackBot:  blackBot,
		Games:     make([]GameExport, 0),
	}

	var totalMoves int
	for _, s := range m.sessions {
		if s == nil || !s.IsFinished() {
			continue
		}

		result := s.Result()
		if result == nil {
			continue
		}

		export.TotalGames++

		// Determine result string for export
		var resultStr string
		if result.Winner == "Draw" {
			resultStr = "Draw"
			export.Draws++
		} else if result.Winner == m.whiteName {
			resultStr = "White"
			export.WhiteWins++
		} else {
			resultStr = "Black"
			export.BlackWins++
		}

		// Convert move history to string notation
		moves := make([]string, len(result.MoveHistory))
		for i, move := range result.MoveHistory {
			moves[i] = move.String()
		}

		totalMoves += result.MoveCount

		gameExport := GameExport{
			GameNumber:        result.GameNumber,
			Result:            resultStr,
			TerminationReason: result.EndReason,
			MoveCount:         result.MoveCount,
			Moves:             moves,
			FinalFEN:          result.FinalFEN,
		}
		export.Games = append(export.Games, gameExport)
	}

	// Calculate average moves
	if export.TotalGames > 0 {
		export.AverageMoves = float64(totalMoves) / float64(export.TotalGames)
	}

	return export
}

// SaveSessionExport saves a SessionExport to a JSON file.
// If dir is empty, it uses the default directory ~/.termchess/stats/.
// Returns the full path to the created file, or an error if the operation fails.
func SaveSessionExport(export *SessionExport, dir string) (string, error) {
	if export == nil {
		return "", fmt.Errorf("export cannot be nil")
	}

	// Use default directory if not specified
	if dir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		dir = filepath.Join(homeDir, ".termchess", "stats")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate filename with timestamp
	timestamp := export.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	filename := fmt.Sprintf("bvb_session_%s.json", timestamp.Format("2006-01-02_15-04-05"))
	filepath := filepath.Join(dir, filename)

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal export: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filepath, nil
}
