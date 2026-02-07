package bvb

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Mgrdich/TermChess/internal/bot"
	"github.com/Mgrdich/TermChess/internal/engine"
)

func TestExportStatsBasic(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "Easy White", "Easy Black", 2, 0)
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

	export := m.ExportStats("Easy", "Easy")
	if export == nil {
		t.Fatal("ExportStats returned nil")
	}

	if export.WhiteBot != "Easy" {
		t.Errorf("WhiteBot = %q, want %q", export.WhiteBot, "Easy")
	}
	if export.BlackBot != "Easy" {
		t.Errorf("BlackBot = %q, want %q", export.BlackBot, "Easy")
	}
	if export.TotalGames != 2 {
		t.Errorf("TotalGames = %d, want 2", export.TotalGames)
	}

	// Total outcomes should sum to TotalGames.
	totalOutcomes := export.WhiteWins + export.BlackWins + export.Draws
	if totalOutcomes != export.TotalGames {
		t.Errorf("WhiteWins(%d) + BlackWins(%d) + Draws(%d) = %d, want %d",
			export.WhiteWins, export.BlackWins, export.Draws, totalOutcomes, export.TotalGames)
	}

	if export.AverageMoves <= 0 {
		t.Errorf("AverageMoves = %f, want > 0", export.AverageMoves)
	}

	if len(export.Games) != 2 {
		t.Errorf("Games length = %d, want 2", len(export.Games))
	}

	// Verify each game export has valid data.
	for i, game := range export.Games {
		if game.GameNumber <= 0 {
			t.Errorf("Games[%d].GameNumber = %d, want > 0", i, game.GameNumber)
		}
		if game.Result != "White" && game.Result != "Black" && game.Result != "Draw" {
			t.Errorf("Games[%d].Result = %q, want White/Black/Draw", i, game.Result)
		}
		if game.TerminationReason == "" {
			t.Errorf("Games[%d].TerminationReason is empty", i)
		}
		if game.MoveCount <= 0 {
			t.Errorf("Games[%d].MoveCount = %d, want > 0", i, game.MoveCount)
		}
		if len(game.Moves) != game.MoveCount {
			t.Errorf("Games[%d].Moves length = %d, want %d", i, len(game.Moves), game.MoveCount)
		}
		if game.FinalFEN == "" {
			t.Errorf("Games[%d].FinalFEN is empty", i)
		}
	}
}

func TestExportStatsMoveHistory(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "Easy White", "Easy Black", 1, 0)
	m.speed = SpeedInstant
	err := m.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	// Wait for game to finish.
	deadline := time.After(60 * time.Second)
	for !m.AllFinished() {
		select {
		case <-deadline:
			m.Abort()
			t.Fatal("game did not complete within timeout")
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}

	export := m.ExportStats("Easy", "Easy")
	if len(export.Games) != 1 {
		t.Fatalf("Games length = %d, want 1", len(export.Games))
	}

	game := export.Games[0]

	// Verify move format (coordinate notation like "e2e4").
	for i, move := range game.Moves {
		if len(move) < 4 || len(move) > 5 {
			t.Errorf("Games[0].Moves[%d] = %q, invalid length (expected 4-5 chars)", i, move)
		}
		// First two chars should be from square (e.g., "e2").
		if move[0] < 'a' || move[0] > 'h' {
			t.Errorf("Games[0].Moves[%d] = %q, invalid from file", i, move)
		}
		if move[1] < '1' || move[1] > '8' {
			t.Errorf("Games[0].Moves[%d] = %q, invalid from rank", i, move)
		}
		// Next two chars should be to square.
		if move[2] < 'a' || move[2] > 'h' {
			t.Errorf("Games[0].Moves[%d] = %q, invalid to file", i, move)
		}
		if move[3] < '1' || move[3] > '8' {
			t.Errorf("Games[0].Moves[%d] = %q, invalid to rank", i, move)
		}
	}
}

func TestExportStatsEmpty(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "W", "B", 3, 0)

	// Don't start the session - export should handle empty state.
	export := m.ExportStats("Easy", "Easy")
	if export == nil {
		t.Fatal("ExportStats returned nil for empty session")
	}

	if export.TotalGames != 0 {
		t.Errorf("TotalGames = %d, want 0", export.TotalGames)
	}
	if export.WhiteWins != 0 {
		t.Errorf("WhiteWins = %d, want 0", export.WhiteWins)
	}
	if export.BlackWins != 0 {
		t.Errorf("BlackWins = %d, want 0", export.BlackWins)
	}
	if export.Draws != 0 {
		t.Errorf("Draws = %d, want 0", export.Draws)
	}
	if export.AverageMoves != 0 {
		t.Errorf("AverageMoves = %f, want 0", export.AverageMoves)
	}
	if len(export.Games) != 0 {
		t.Errorf("Games length = %d, want 0", len(export.Games))
	}
}

func TestSaveSessionExportCreatesFile(t *testing.T) {
	// Create a temporary directory for test.
	tmpDir := t.TempDir()

	export := &SessionExport{
		Timestamp:    time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
		WhiteBot:     "Easy",
		BlackBot:     "Medium",
		TotalGames:   3,
		WhiteWins:    1,
		BlackWins:    1,
		Draws:        1,
		AverageMoves: 42.5,
		Games: []GameExport{
			{
				GameNumber:        1,
				Result:            "White",
				TerminationReason: "Checkmate",
				MoveCount:         40,
				Moves:             []string{"e2e4", "e7e5", "g1f3"},
				FinalFEN:          "rnbqkbnr/pppp1ppp/8/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2",
			},
		},
	}

	filepath, err := SaveSessionExport(export, tmpDir)
	if err != nil {
		t.Fatalf("SaveSessionExport() error: %v", err)
	}

	// Verify file was created.
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Fatalf("File was not created at %s", filepath)
	}

	// Verify filename format.
	expectedFilename := "bvb_session_2024-01-15_10-30-45.json"
	if !strings.HasSuffix(filepath, expectedFilename) {
		t.Errorf("Filename = %s, want suffix %s", filepath, expectedFilename)
	}
}

func TestSaveSessionExportJSONFormat(t *testing.T) {
	tmpDir := t.TempDir()

	export := &SessionExport{
		Timestamp:    time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
		WhiteBot:     "Hard",
		BlackBot:     "Hard",
		TotalGames:   2,
		WhiteWins:    1,
		BlackWins:    0,
		Draws:        1,
		AverageMoves: 50.0,
		Games: []GameExport{
			{
				GameNumber:        1,
				Result:            "White",
				TerminationReason: "Checkmate",
				MoveCount:         45,
				Moves:             []string{"d2d4", "d7d5"},
				FinalFEN:          "some-fen-1",
			},
			{
				GameNumber:        2,
				Result:            "Draw",
				TerminationReason: "Stalemate",
				MoveCount:         55,
				Moves:             []string{"e2e4"},
				FinalFEN:          "some-fen-2",
			},
		},
	}

	filepath, err := SaveSessionExport(export, tmpDir)
	if err != nil {
		t.Fatalf("SaveSessionExport() error: %v", err)
	}

	// Read file and verify JSON structure.
	data, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Verify it's valid JSON.
	var loaded SessionExport
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify key fields.
	if loaded.WhiteBot != "Hard" {
		t.Errorf("WhiteBot = %q, want %q", loaded.WhiteBot, "Hard")
	}
	if loaded.BlackBot != "Hard" {
		t.Errorf("BlackBot = %q, want %q", loaded.BlackBot, "Hard")
	}
	if loaded.TotalGames != 2 {
		t.Errorf("TotalGames = %d, want 2", loaded.TotalGames)
	}
	if loaded.WhiteWins != 1 {
		t.Errorf("WhiteWins = %d, want 1", loaded.WhiteWins)
	}
	if loaded.Draws != 1 {
		t.Errorf("Draws = %d, want 1", loaded.Draws)
	}
	if loaded.AverageMoves != 50.0 {
		t.Errorf("AverageMoves = %f, want 50.0", loaded.AverageMoves)
	}

	if len(loaded.Games) != 2 {
		t.Fatalf("Games length = %d, want 2", len(loaded.Games))
	}

	if loaded.Games[0].Result != "White" {
		t.Errorf("Games[0].Result = %q, want %q", loaded.Games[0].Result, "White")
	}
	if loaded.Games[1].Result != "Draw" {
		t.Errorf("Games[1].Result = %q, want %q", loaded.Games[1].Result, "Draw")
	}
}

func TestSaveSessionExportCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "nested", "deep", "stats")

	export := &SessionExport{
		Timestamp:  time.Now(),
		WhiteBot:   "Easy",
		BlackBot:   "Easy",
		TotalGames: 0,
		Games:      []GameExport{},
	}

	filepath, err := SaveSessionExport(export, nestedDir)
	if err != nil {
		t.Fatalf("SaveSessionExport() error: %v", err)
	}

	// Verify directory was created.
	info, err := os.Stat(nestedDir)
	if err != nil {
		t.Fatalf("Directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("Expected directory, got file")
	}

	// Verify file exists.
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Fatalf("File was not created at %s", filepath)
	}
}

func TestSaveSessionExportNilExport(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := SaveSessionExport(nil, tmpDir)
	if err == nil {
		t.Error("SaveSessionExport(nil) should return error")
	}
}

func TestSaveSessionExportDefaultDirectory(t *testing.T) {
	export := &SessionExport{
		Timestamp:  time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
		WhiteBot:   "Test",
		BlackBot:   "Test",
		TotalGames: 0,
		Games:      []GameExport{},
	}

	// Use empty string to trigger default directory.
	filepath, err := SaveSessionExport(export, "")
	if err != nil {
		t.Fatalf("SaveSessionExport() error: %v", err)
	}

	// Verify file was created.
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Fatalf("File was not created at %s", filepath)
	}

	// Verify path contains expected components.
	if !strings.Contains(filepath, ".termchess") {
		t.Errorf("Path %s does not contain .termchess", filepath)
	}
	if !strings.Contains(filepath, "stats") {
		t.Errorf("Path %s does not contain stats", filepath)
	}

	// Clean up the created file.
	_ = os.Remove(filepath)
}

func TestGameExportTerminationReasons(t *testing.T) {
	// Test various termination reasons are preserved.
	reasons := []string{"Checkmate", "Stalemate", "Insufficient Material", "move limit exceeded"}

	for _, reason := range reasons {
		export := &SessionExport{
			Timestamp:  time.Now(),
			WhiteBot:   "Test",
			BlackBot:   "Test",
			TotalGames: 1,
			Games: []GameExport{
				{
					GameNumber:        1,
					Result:            "White",
					TerminationReason: reason,
					MoveCount:         10,
					Moves:             []string{"e2e4"},
					FinalFEN:          "test-fen",
				},
			},
		}

		tmpDir := t.TempDir()
		filepath, err := SaveSessionExport(export, tmpDir)
		if err != nil {
			t.Fatalf("SaveSessionExport() error for reason %q: %v", reason, err)
		}

		// Read back and verify.
		data, err := os.ReadFile(filepath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		var loaded SessionExport
		if err := json.Unmarshal(data, &loaded); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		if loaded.Games[0].TerminationReason != reason {
			t.Errorf("TerminationReason = %q, want %q", loaded.Games[0].TerminationReason, reason)
		}
	}
}

func TestExportStatsTimestamp(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "W", "B", 1, 0)

	beforeExport := time.Now()
	export := m.ExportStats("Easy", "Easy")
	afterExport := time.Now()

	if export.Timestamp.Before(beforeExport) {
		t.Errorf("Timestamp %v is before export time %v", export.Timestamp, beforeExport)
	}
	if export.Timestamp.After(afterExport) {
		t.Errorf("Timestamp %v is after export time %v", export.Timestamp, afterExport)
	}
}

func TestMoveHistoryRecordedCorrectly(t *testing.T) {
	// Run a real game and verify move history is properly recorded.
	m := NewSessionManager(bot.Easy, bot.Easy, "Easy White", "Easy Black", 1, 0)
	m.speed = SpeedInstant
	err := m.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	// Wait for game to finish.
	deadline := time.After(60 * time.Second)
	for !m.AllFinished() {
		select {
		case <-deadline:
			m.Abort()
			t.Fatal("game did not complete within timeout")
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}

	// Get the session result directly.
	sessions := m.Sessions()
	if len(sessions) != 1 {
		t.Fatalf("Sessions length = %d, want 1", len(sessions))
	}

	result := sessions[0].Result()
	if result == nil {
		t.Fatal("Session result is nil")
	}

	// Export and compare.
	export := m.ExportStats("Easy", "Easy")
	if len(export.Games) != 1 {
		t.Fatalf("Exported games = %d, want 1", len(export.Games))
	}

	// Verify move count matches.
	if export.Games[0].MoveCount != result.MoveCount {
		t.Errorf("Exported MoveCount = %d, result MoveCount = %d",
			export.Games[0].MoveCount, result.MoveCount)
	}

	// Verify moves array length matches.
	if len(export.Games[0].Moves) != len(result.MoveHistory) {
		t.Errorf("Exported Moves length = %d, MoveHistory length = %d",
			len(export.Games[0].Moves), len(result.MoveHistory))
	}

	// Verify each move matches.
	for i, move := range result.MoveHistory {
		if export.Games[0].Moves[i] != move.String() {
			t.Errorf("Moves[%d] = %q, want %q", i, export.Games[0].Moves[i], move.String())
		}
	}
}

func TestExportStatsResultMapping(t *testing.T) {
	// Create a mock scenario with known results.
	// We'll need to run actual games to test this properly.
	m := NewSessionManager(bot.Easy, bot.Easy, "Easy White", "Easy Black", 3, 0)
	m.speed = SpeedInstant
	err := m.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	// Wait for games to finish.
	deadline := time.After(90 * time.Second)
	for !m.AllFinished() {
		select {
		case <-deadline:
			m.Abort()
			t.Fatal("games did not complete within timeout")
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}

	export := m.ExportStats("Easy", "Easy")

	// Verify result mapping is consistent.
	whiteCount := 0
	blackCount := 0
	drawCount := 0

	for _, game := range export.Games {
		switch game.Result {
		case "White":
			whiteCount++
		case "Black":
			blackCount++
		case "Draw":
			drawCount++
		default:
			t.Errorf("Invalid result %q for game %d", game.Result, game.GameNumber)
		}
	}

	if whiteCount != export.WhiteWins {
		t.Errorf("White count mismatch: games=%d, WhiteWins=%d", whiteCount, export.WhiteWins)
	}
	if blackCount != export.BlackWins {
		t.Errorf("Black count mismatch: games=%d, BlackWins=%d", blackCount, export.BlackWins)
	}
	if drawCount != export.Draws {
		t.Errorf("Draw count mismatch: games=%d, Draws=%d", drawCount, export.Draws)
	}
}

func TestFinalFENRecorded(t *testing.T) {
	m := NewSessionManager(bot.Easy, bot.Easy, "Easy White", "Easy Black", 1, 0)
	m.speed = SpeedInstant
	err := m.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	// Wait for game to finish.
	deadline := time.After(60 * time.Second)
	for !m.AllFinished() {
		select {
		case <-deadline:
			m.Abort()
			t.Fatal("game did not complete within timeout")
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}

	export := m.ExportStats("Easy", "Easy")
	if len(export.Games) != 1 {
		t.Fatalf("Games length = %d, want 1", len(export.Games))
	}

	// Verify FEN is not empty and has valid format.
	fen := export.Games[0].FinalFEN
	if fen == "" {
		t.Error("FinalFEN is empty")
	}

	// Basic FEN validation: should contain spaces separating fields.
	parts := strings.Split(fen, " ")
	if len(parts) < 4 {
		t.Errorf("FinalFEN has %d parts, expected at least 4", len(parts))
	}

	// Verify it can be parsed back.
	_, err = engine.FromFEN(fen)
	if err != nil {
		t.Errorf("FinalFEN is not valid FEN: %v", err)
	}
}
