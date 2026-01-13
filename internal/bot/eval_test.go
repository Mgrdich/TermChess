package bot

import (
	"math"
	"testing"

	"github.com/Mgrdich/TermChess/internal/engine"
)

func TestEvaluate_Checkmate(t *testing.T) {
	// White checkmate (White wins) - Queen and King mate
	fen := "7k/6Q1/5K2/8/8/8/8/8 b - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	score := evaluate(board, Easy)
	if score != 10000.0 {
		t.Errorf("White checkmate: evaluate() = %v, want 10000", score)
	}

	// Black checkmate (Black wins) - Queen and King mate
	fen = "8/8/8/8/8/5k2/6q1/7K w - - 0 1"
	board, err = engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	score = evaluate(board, Easy)
	if score != -10000.0 {
		t.Errorf("Black checkmate: evaluate() = %v, want -10000", score)
	}
}

func TestEvaluate_Stalemate(t *testing.T) {
	// Stalemate position
	fen := "7k/5Q2/5K2/8/8/8/8/8 b - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	score := evaluate(board, Easy)
	if score != 0.0 {
		t.Errorf("Stalemate: evaluate() = %v, want 0", score)
	}
}

func TestEvaluate_StartPosition(t *testing.T) {
	// Starting position should be equal (score ~0)
	board := engine.NewBoard()

	score := evaluate(board, Easy)
	if math.Abs(score) > 0.01 {
		t.Errorf("Starting position: evaluate() = %v, want ~0", score)
	}
}

func TestEvaluate_MaterialAdvantage(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		wantMin  float64
		wantMax  float64
		desc     string
	}{
		{
			name:    "WhiteExtraQueen",
			fen:     "rnb1kbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			wantMin: 8.0,
			wantMax: 10.0,
			desc:    "White has extra queen (~9 pawns)",
		},
		{
			name:    "WhiteExtraRook",
			fen:     "rnbqkbn1/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			wantMin: 4.0,
			wantMax: 6.0,
			desc:    "White has extra rook (~5 pawns)",
		},
		{
			name:    "BlackExtraQueen",
			fen:     "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNB1KBNR w KQkq - 0 1",
			wantMin: -10.0,
			wantMax: -8.0,
			desc:    "Black has extra queen (~-9 pawns)",
		},
		{
			name:    "WhiteExtraKnight",
			fen:     "rnbqkb1r/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			wantMin: 2.0,
			wantMax: 4.0,
			desc:    "White has extra knight (~3 pawns)",
		},
		{
			name:    "WhiteExtraBishop",
			fen:     "rn1qkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			wantMin: 2.5,
			wantMax: 4.0,
			desc:    "White has extra bishop (~3.25 pawns)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := engine.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to parse FEN: %v", err)
			}

			score := evaluate(board, Easy)
			if score < tt.wantMin || score > tt.wantMax {
				t.Errorf("%s: evaluate() = %v, want between %v and %v (%s)",
					tt.name, score, tt.wantMin, tt.wantMax, tt.desc)
			}
		})
	}
}

func TestCountMaterial(t *testing.T) {
	// Test starting position (equal material)
	board := engine.NewBoard()
	score := countMaterial(board)
	if math.Abs(score) > 0.01 {
		t.Errorf("Starting position: countMaterial() = %v, want ~0", score)
	}

	// Test position with material imbalance
	// Position: White has extra queen (missing black queen)
	fen := "rnb1kbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	score = countMaterial(board)
	expectedScore := 9.0 // White has extra queen worth 9 pawns
	if math.Abs(score-expectedScore) > 0.1 {
		t.Errorf("Extra queen: countMaterial() = %v, want ~%v", score, expectedScore)
	}

	// Test position with Black material advantage
	fen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNB1KBNR w KQkq - 0 1"
	board, err = engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	score = countMaterial(board)
	expectedScore = -9.0 // Black has extra queen worth -9 pawns (from White's perspective)
	if math.Abs(score-expectedScore) > 0.1 {
		t.Errorf("Black extra queen: countMaterial() = %v, want ~%v", score, expectedScore)
	}

	// Test endgame position with rook vs pawns
	fen = "7k/8/8/8/8/8/PPPPPPPP/7K w - - 0 1"
	board, err = engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	score = countMaterial(board)
	expectedScore = 8.0 // 8 white pawns = 8.0
	if math.Abs(score-expectedScore) > 0.1 {
		t.Errorf("White 8 pawns: countMaterial() = %v, want ~%v", score, expectedScore)
	}
}

func TestEvaluate_Symmetry(t *testing.T) {
	// Test that eval(position) has consistent scoring
	// For symmetric positions, the evaluation should reflect material balance

	// Original position (White has moved e2-e4)
	fen := "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"
	board1, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	score1 := evaluate(board1, Easy)

	// Flipped position (Black has moved e7-e5)
	fenFlipped := "rnbqkbnr/pppp1ppp/8/4p3/8/8/PPPPPPPP/RNBQKBNR w KQkq e6 0 1"
	board2, err := engine.FromFEN(fenFlipped)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	score2 := evaluate(board2, Easy)

	// Both positions have equal material, so scores should be ~0
	// The test verifies material counting is symmetric
	if math.Abs(score1) > 0.01 {
		t.Errorf("Position 1: eval() = %v, want ~0 (equal material)", score1)
	}
	if math.Abs(score2) > 0.01 {
		t.Errorf("Position 2: eval() = %v, want ~0 (equal material)", score2)
	}
}

func TestEvaluate_DrawByRepetition(t *testing.T) {
	// Test draw by insufficient material
	fen := "8/8/8/8/8/4k3/8/4K3 w - - 0 1" // Only kings
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	score := evaluate(board, Easy)
	if score != 0.0 {
		t.Errorf("Draw by insufficient material: evaluate() = %v, want 0", score)
	}

	// Test king and bishop vs king
	fen = "8/8/8/8/8/4k3/8/4KB2 w - - 0 1"
	board, err = engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	score = evaluate(board, Easy)
	if score != 0.0 {
		t.Errorf("Draw by insufficient material (K+B vs K): evaluate() = %v, want 0", score)
	}

	// Test king and knight vs king
	fen = "8/8/8/8/8/4k3/8/4KN2 w - - 0 1"
	board, err = engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	score = evaluate(board, Easy)
	if score != 0.0 {
		t.Errorf("Draw by insufficient material (K+N vs K): evaluate() = %v, want 0", score)
	}
}

func TestEvaluate_DrawFiftyMoveRule(t *testing.T) {
	// Test fifty-move rule draw
	// Create a position where the half-move clock is at 100 (50 full moves)
	fen := "8/8/4k3/8/8/4K3/8/8 w - - 100 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	score := evaluate(board, Easy)
	if score != 0.0 {
		t.Errorf("Draw by fifty-move rule: evaluate() = %v, want 0", score)
	}

	// Test seventy-five-move rule (automatic draw)
	fen = "8/8/4k3/8/8/4K3/8/8 w - - 150 1"
	board, err = engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	score = evaluate(board, Easy)
	if score != 0.0 {
		t.Errorf("Draw by seventy-five-move rule: evaluate() = %v, want 0", score)
	}
}

func TestPieceValues(t *testing.T) {
	// Verify piece values are set correctly
	expectedValues := map[engine.PieceType]float64{
		engine.Pawn:   1.0,
		engine.Knight: 3.0,
		engine.Bishop: 3.25,
		engine.Rook:   5.0,
		engine.Queen:  9.0,
		engine.King:   0.0,
	}

	for pieceType, expectedValue := range expectedValues {
		actualValue := pieceValues[pieceType]
		if actualValue != expectedValue {
			t.Errorf("pieceValues[%v] = %v, want %v", pieceType, actualValue, expectedValue)
		}
	}
}

func TestEvaluate_ComplexPositions(t *testing.T) {
	tests := []struct {
		name        string
		fen         string
		wantScore   float64
		tolerance   float64
		description string
	}{
		{
			name:        "EmptyBoard",
			fen:         "8/8/8/8/8/8/8/8 w - - 0 1",
			wantScore:   0.0,
			tolerance:   0.01,
			description: "Empty board should score 0",
		},
		{
			name:        "OnlyKings",
			fen:         "4k3/8/8/8/8/8/8/4K3 w - - 0 1",
			wantScore:   0.0,
			tolerance:   0.01,
			description: "Only kings (draw by insufficient material)",
		},
		{
			name:        "WhiteRookVsBlackKnight",
			fen:         "4k3/8/8/8/8/8/8/4K2R w - - 0 1",
			wantScore:   5.0,
			tolerance:   0.1,
			description: "White rook (5) vs nothing",
		},
		{
			name:        "BlackQueenVsWhiteTwoRooks",
			fen:         "4k3/8/8/8/8/8/q7/4K2R w - - 0 1",
			wantScore:   -4.0,
			tolerance:   0.1,
			description: "White rook (5) vs Black queen (9) = -4",
		},
		{
			name:        "ComplexMaterial",
			fen:         "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 1",
			wantScore:   0.0,
			tolerance:   0.5,
			description: "Italian Game position (roughly equal material)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := engine.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to parse FEN: %v", err)
			}

			score := evaluate(board, Easy)
			diff := math.Abs(score - tt.wantScore)
			if diff > tt.tolerance {
				t.Errorf("%s: evaluate() = %v, want %v (±%v) - %s",
					tt.name, score, tt.wantScore, tt.tolerance, tt.description)
			}
		})
	}
}

func TestEvaluate_DifficultyParameter(t *testing.T) {
	// Test that evaluate() accepts all difficulty levels
	// (Currently difficulty doesn't affect material evaluation,
	// but this test ensures the function signature works correctly)
	board := engine.NewBoard()

	scoreEasy := evaluate(board, Easy)
	scoreMedium := evaluate(board, Medium)
	scoreHard := evaluate(board, Hard)

	// For material-only evaluation, all difficulties should give same result
	if scoreEasy != scoreMedium || scoreMedium != scoreHard {
		t.Errorf("Material evaluation should be same for all difficulties: Easy=%v, Medium=%v, Hard=%v",
			scoreEasy, scoreMedium, scoreHard)
	}

	if math.Abs(scoreEasy) > 0.01 {
		t.Errorf("Starting position should evaluate to ~0, got %v", scoreEasy)
	}
}
