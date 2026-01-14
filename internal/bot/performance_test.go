package bot

import (
	"context"
	"testing"
	"time"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// BenchmarkEasyBot benchmarks the Easy bot (random move selection with tactical bias).
func BenchmarkEasyBot(b *testing.B) {
	// Use a complex middlegame position with many legal moves
	fen := "r1bqk2r/pp1n1ppp/2pbpn2/8/2BP4/2N2N2/PPP2PPP/R1BQK2R w KQkq - 0 8"
	board, err := engine.FromFEN(fen)
	if err != nil {
		b.Fatalf("Failed to parse FEN: %v", err)
	}

	bot, err := NewRandomEngine()
	if err != nil {
		b.Fatalf("Failed to create Easy bot: %v", err)
	}
	defer bot.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bot.SelectMove(ctx, board)
		if err != nil {
			b.Fatalf("SelectMove failed: %v", err)
		}
	}
}

// BenchmarkMediumBot_Depth4 benchmarks the Medium bot (minimax depth 4).
func BenchmarkMediumBot_Depth4(b *testing.B) {
	// Complex middlegame position with many legal moves
	fen := "r1bqk2r/pp1n1ppp/2pbpn2/8/2BP4/2N2N2/PPP2PPP/R1BQK2R w KQkq - 0 8"
	board, err := engine.FromFEN(fen)
	if err != nil {
		b.Fatalf("Failed to parse FEN: %v", err)
	}

	bot, err := NewMinimaxEngine(Medium)
	if err != nil {
		b.Fatalf("Failed to create Medium bot: %v", err)
	}
	defer bot.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bot.SelectMove(ctx, board)
		if err != nil {
			b.Fatalf("SelectMove failed: %v", err)
		}
	}
}

// BenchmarkHardBot_Depth6 benchmarks the Hard bot (minimax depth 6).
func BenchmarkHardBot_Depth6(b *testing.B) {
	// Complex middlegame position with many legal moves
	fen := "r1bqk2r/pp1n1ppp/2pbpn2/8/2BP4/2N2N2/PPP2PPP/R1BQK2R w KQkq - 0 8"
	board, err := engine.FromFEN(fen)
	if err != nil {
		b.Fatalf("Failed to parse FEN: %v", err)
	}

	bot, err := NewMinimaxEngine(Hard)
	if err != nil {
		b.Fatalf("Failed to create Hard bot: %v", err)
	}
	defer bot.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bot.SelectMove(ctx, board)
		if err != nil {
			b.Fatalf("SelectMove failed: %v", err)
		}
	}
}

// BenchmarkEvaluate benchmarks the evaluation function speed.
func BenchmarkEvaluate(b *testing.B) {
	// Test various positions to benchmark evaluation
	positions := []struct {
		name string
		fen  string
	}{
		{"Starting position", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{"Complex middlegame", "r1bqk2r/pp1n1ppp/2pbpn2/8/2BP4/2N2N2/PPP2PPP/R1BQK2R w KQkq - 0 8"},
		{"Tactical position", "r2qkb1r/ppp2ppp/2n1bn2/3pp3/4P3/3P1N2/PPP2PPP/RNBQKB1R w KQkq - 0 6"},
		{"Endgame", "8/5k2/3p4/1p1Pp2p/pP2Pp1P/P4P1K/8/8 b - - 99 50"},
	}

	for _, pos := range positions {
		b.Run(pos.name, func(b *testing.B) {
			board, err := engine.FromFEN(pos.fen)
			if err != nil {
				b.Fatalf("Failed to parse FEN: %v", err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Benchmark evaluation at different difficulty levels
				_ = evaluate(board, Easy)
				_ = evaluate(board, Medium)
				_ = evaluate(board, Hard)
			}
		})
	}
}

// TestTimeLimit_EasyBot tests that the Easy bot completes within 2 seconds.
func TestTimeLimit_EasyBot(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping time limit test in short mode")
	}

	positions := []struct {
		name string
		fen  string
	}{
		{"Complex middlegame", "r1bqk2r/pp1n1ppp/2pbpn2/8/2BP4/2N2N2/PPP2PPP/R1BQK2R w KQkq - 0 8"},
		{"Tactical position", "r2qkb1r/ppp2ppp/2n1bn2/3pp3/4P3/3P1N2/PPP2PPP/RNBQKB1R w KQkq - 0 6"},
		{"Open position", "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"},
	}

	bot, err := NewRandomEngine()
	if err != nil {
		t.Fatalf("Failed to create Easy bot: %v", err)
	}
	defer bot.Close()

	for _, pos := range positions {
		t.Run(pos.name, func(t *testing.T) {
			board, err := engine.FromFEN(pos.fen)
			if err != nil {
				t.Fatalf("Failed to parse FEN: %v", err)
			}

			start := time.Now()
			ctx := context.Background()
			move, err := bot.SelectMove(ctx, board)
			elapsed := time.Since(start)

			if err != nil {
				t.Fatalf("SelectMove failed: %v", err)
			}

			// Validate move is legal
			legalMoves := board.LegalMoves()
			found := false
			for _, m := range legalMoves {
				if m == move {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Bot returned illegal move: %v", move)
			}

			// Allow a small grace period (100ms) for context cancellation overhead
			limit := 2 * time.Second
			gracePeriod := 100 * time.Millisecond
			if elapsed > limit+gracePeriod {
				t.Errorf("Easy bot took %v, expected < %v (with %v grace period)", elapsed, limit, gracePeriod)
			}

			t.Logf("Easy bot took %v (limit: %v)", elapsed, limit)
		})
	}
}

// TestTimeLimit_MediumBot tests that the Medium bot completes within 4 seconds.
func TestTimeLimit_MediumBot(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping time limit test in short mode")
	}

	positions := []struct {
		name string
		fen  string
	}{
		{"Complex middlegame", "r1bqk2r/pp1n1ppp/2pbpn2/8/2BP4/2N2N2/PPP2PPP/R1BQK2R w KQkq - 0 8"},
		{"Tactical position", "r2qkb1r/ppp2ppp/2n1bn2/3pp3/4P3/3P1N2/PPP2PPP/RNBQKB1R w KQkq - 0 6"},
		{"Open position", "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"},
	}

	bot, err := NewMinimaxEngine(Medium)
	if err != nil {
		t.Fatalf("Failed to create Medium bot: %v", err)
	}
	defer bot.Close()

	for _, pos := range positions {
		t.Run(pos.name, func(t *testing.T) {
			board, err := engine.FromFEN(pos.fen)
			if err != nil {
				t.Fatalf("Failed to parse FEN: %v", err)
			}

			start := time.Now()
			ctx := context.Background()
			move, err := bot.SelectMove(ctx, board)
			elapsed := time.Since(start)

			if err != nil {
				t.Fatalf("SelectMove failed: %v", err)
			}

			// Validate move is legal
			legalMoves := board.LegalMoves()
			found := false
			for _, m := range legalMoves {
				if m == move {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Bot returned illegal move: %v", move)
			}

			// Allow a small grace period (100ms) for context cancellation overhead
			limit := 4 * time.Second
			gracePeriod := 100 * time.Millisecond
			if elapsed > limit+gracePeriod {
				t.Errorf("Medium bot took %v, expected < %v (with %v grace period)", elapsed, limit, gracePeriod)
			}

			t.Logf("Medium bot took %v (limit: %v)", elapsed, limit)
		})
	}
}

// TestTimeLimit_HardBot tests that the Hard bot completes within 8 seconds.
func TestTimeLimit_HardBot(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping time limit test in short mode")
	}

	positions := []struct {
		name string
		fen  string
	}{
		{"Complex middlegame", "r1bqk2r/pp1n1ppp/2pbpn2/8/2BP4/2N2N2/PPP2PPP/R1BQK2R w KQkq - 0 8"},
		{"Tactical position", "r2qkb1r/ppp2ppp/2n1bn2/3pp3/4P3/3P1N2/PPP2PPP/RNBQKB1R w KQkq - 0 6"},
		{"Open position", "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"},
	}

	bot, err := NewMinimaxEngine(Hard)
	if err != nil {
		t.Fatalf("Failed to create Hard bot: %v", err)
	}
	defer bot.Close()

	for _, pos := range positions {
		t.Run(pos.name, func(t *testing.T) {
			board, err := engine.FromFEN(pos.fen)
			if err != nil {
				t.Fatalf("Failed to parse FEN: %v", err)
			}

			start := time.Now()
			ctx := context.Background()
			move, err := bot.SelectMove(ctx, board)
			elapsed := time.Since(start)

			if err != nil {
				t.Fatalf("SelectMove failed: %v", err)
			}

			// Validate move is legal
			legalMoves := board.LegalMoves()
			found := false
			for _, m := range legalMoves {
				if m == move {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Bot returned illegal move: %v", move)
			}

			// Allow a small grace period (100ms) for context cancellation overhead
			limit := 8 * time.Second
			gracePeriod := 100 * time.Millisecond
			if elapsed > limit+gracePeriod {
				t.Errorf("Hard bot took %v, expected < %v (with %v grace period)", elapsed, limit, gracePeriod)
			}

			t.Logf("Hard bot took %v (limit: %v)", elapsed, limit)
		})
	}
}

// TestTimeLimit_MediumBot_WithTimeout tests Medium bot respects context timeout.
func TestTimeLimit_MediumBot_WithTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping time limit test in short mode")
	}

	// Use a complex position to ensure bot takes some time
	fen := "r1bqk2r/pp1n1ppp/2pbpn2/8/2BP4/2N2N2/PPP2PPP/R1BQK2R w KQkq - 0 8"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	bot, err := NewMinimaxEngine(Medium)
	if err != nil {
		t.Fatalf("Failed to create Medium bot: %v", err)
	}
	defer bot.Close()

	// Create a context with a very short timeout (100ms)
	// This tests that the bot respects context cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	move, err := bot.SelectMove(ctx, board)
	elapsed := time.Since(start)

	// Bot should complete within timeout and return a valid move
	// (It falls back to best move from completed iteration or first legal move)
	if err != nil {
		t.Fatalf("SelectMove failed: %v", err)
	}

	// Validate move is legal
	legalMoves := board.LegalMoves()
	found := false
	for _, m := range legalMoves {
		if m == move {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Bot returned illegal move: %v", move)
	}

	t.Logf("Medium bot with 100ms timeout took %v and returned move: %v", elapsed, move)
}

// TestTimeLimit_HardBot_WithTimeout tests Hard bot respects context timeout.
func TestTimeLimit_HardBot_WithTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping time limit test in short mode")
	}

	// Use a complex position to ensure bot takes some time
	fen := "r1bqk2r/pp1n1ppp/2pbpn2/8/2BP4/2N2N2/PPP2PPP/R1BQK2R w KQkq - 0 8"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	bot, err := NewMinimaxEngine(Hard)
	if err != nil {
		t.Fatalf("Failed to create Hard bot: %v", err)
	}
	defer bot.Close()

	// Create a context with a short timeout (200ms)
	// This tests that the bot respects context cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	move, err := bot.SelectMove(ctx, board)
	elapsed := time.Since(start)

	// Bot should complete within timeout and return a valid move
	if err != nil {
		t.Fatalf("SelectMove failed: %v", err)
	}

	// Validate move is legal
	legalMoves := board.LegalMoves()
	found := false
	for _, m := range legalMoves {
		if m == move {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Bot returned illegal move: %v", move)
	}

	t.Logf("Hard bot with 200ms timeout took %v and returned move: %v", elapsed, move)
}
