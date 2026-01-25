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
	board := engine.NewBoard()

	scoreEasy := evaluate(board, Easy)
	scoreMedium := evaluate(board, Medium)
	scoreHard := evaluate(board, Hard)

	// Easy should use only material evaluation
	if math.Abs(scoreEasy) > 0.01 {
		t.Errorf("Starting position Easy evaluation should be ~0, got %v", scoreEasy)
	}

	// Medium and Hard use additional evaluation components
	// They should be equal to each other (same components, different search depth in minimax)
	if scoreMedium != scoreHard {
		t.Errorf("Medium and Hard should evaluate the same: Medium=%v, Hard=%v",
			scoreMedium, scoreHard)
	}

	// Starting position should still be close to 0 for all difficulties (symmetric position)
	if math.Abs(scoreMedium) > 3.0 {
		t.Errorf("Starting position Medium/Hard evaluation too far from 0: %v", scoreMedium)
	}
}

func TestEvaluatePiecePositions(t *testing.T) {
	// Test that piece-square tables give reasonable bonuses

	// Test 1: Knight on e4 (central square) should score higher than knight on a1 (corner)
	fenCentralKnight := "8/8/8/8/4N3/8/8/8 w - - 0 1"
	boardCentral, err := engine.FromFEN(fenCentralKnight)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	fenCornerKnight := "N7/8/8/8/8/8/8/8 w - - 0 1"
	boardCorner, err := engine.FromFEN(fenCornerKnight)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	scoreCentral := evaluatePiecePositions(boardCentral, 0.5)
	scoreCorner := evaluatePiecePositions(boardCorner, 0.5)

	if scoreCentral <= scoreCorner {
		t.Errorf("Central knight should score higher than corner knight: central=%v, corner=%v",
			scoreCentral, scoreCorner)
	}

	// Test 2: Advanced pawn (rank 7) should score higher than starting pawn (rank 2)
	fenAdvancedPawn := "8/4P3/8/8/8/8/8/8 w - - 0 1"
	boardAdvanced, err := engine.FromFEN(fenAdvancedPawn)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	fenStartingPawn := "8/8/8/8/8/8/4P3/8 w - - 0 1"
	boardStarting, err := engine.FromFEN(fenStartingPawn)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	scoreAdvanced := evaluatePiecePositions(boardAdvanced, 0.5)
	scoreStarting := evaluatePiecePositions(boardStarting, 0.5)

	if scoreAdvanced <= scoreStarting {
		t.Errorf("Advanced pawn (rank 7) should score higher than starting pawn (rank 2): advanced=%v, starting=%v",
			scoreAdvanced, scoreStarting)
	}

	// Test 3: Symmetric position for Black
	// Black knight on e5 (central) should have same magnitude as White knight on e4
	fenWhiteKnight := "8/8/8/8/4N3/8/8/8 w - - 0 1"
	boardWhite, err := engine.FromFEN(fenWhiteKnight)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	fenBlackKnight := "8/8/8/4n3/8/8/8/8 w - - 0 1"
	boardBlack, err := engine.FromFEN(fenBlackKnight)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	scoreWhiteKnight := evaluatePiecePositions(boardWhite, 0.5)
	scoreBlackKnight := evaluatePiecePositions(boardBlack, 0.5)

	// Scores should be opposite (White positive, Black negative)
	if math.Abs(scoreWhiteKnight+scoreBlackKnight) > 0.01 {
		t.Errorf("Symmetric knight positions should have opposite scores: white=%v, black=%v",
			scoreWhiteKnight, scoreBlackKnight)
	}
}

func TestEvaluateMobility(t *testing.T) {
	// Test mobility scoring

	// Test 1: Open position should have more mobility than cramped position
	fenOpen := "8/8/8/8/8/8/8/R3K3 w - - 0 1" // Rook and King, rook has many moves
	boardOpen, err := engine.FromFEN(fenOpen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	fenCramped := "8/8/8/8/8/8/PPP5/RK6 w - - 0 1" // Rook and King blocked by pawns
	boardCramped, err := engine.FromFEN(fenCramped)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	mobilityOpen := evaluateMobility(boardOpen)
	mobilityCramped := evaluateMobility(boardCramped)

	if mobilityOpen <= mobilityCramped {
		t.Errorf("Open position should have higher mobility: open=%v, cramped=%v",
			mobilityOpen, mobilityCramped)
	}

	// Test 2: Mobility should be positive for White to move
	if mobilityOpen <= 0 {
		t.Errorf("White to move should have positive mobility, got %v", mobilityOpen)
	}

	// Test 3: Mobility should be negative for Black to move
	fenBlackToMove := "8/8/8/8/8/8/8/r3k3 b - - 0 1"
	boardBlackToMove, err := engine.FromFEN(fenBlackToMove)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	mobilityBlack := evaluateMobility(boardBlackToMove)
	if mobilityBlack >= 0 {
		t.Errorf("Black to move should have negative mobility (from White's perspective), got %v", mobilityBlack)
	}

	// Test 4: Starting position mobility check
	board := engine.NewBoard()
	mobilityStart := evaluateMobility(board)

	// White has 20 legal moves in starting position
	if mobilityStart != 20.0 {
		t.Errorf("Starting position should have 20 legal moves for White, got %v", mobilityStart)
	}
}

func TestEvaluate_DifficultyLevels(t *testing.T) {
	// Verify Easy bot doesn't use piece-square tables or mobility
	// Verify Medium bot uses piece-square tables and mobility

	// Position with good piece placement (knight on e4, advanced pawn)
	// Need to include kings for legal position
	fen := "4k3/8/8/8/3NP3/8/8/4K3 w - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	scoreEasy := evaluate(board, Easy)
	scoreMedium := evaluate(board, Medium)
	scoreHard := evaluate(board, Hard)

	// Easy should only count material (N=3, P=1) = 4.0
	expectedEasy := 4.0
	if math.Abs(scoreEasy-expectedEasy) > 0.01 {
		t.Errorf("Easy difficulty should only use material: got %v, want ~%v", scoreEasy, expectedEasy)
	}

	// Medium and Hard should give higher scores due to positional bonuses
	// (knight on e4 gets +0.2, pawn on e4 gets +0.35, plus mobility bonus)
	if scoreMedium <= scoreEasy {
		t.Errorf("Medium should score higher than Easy due to positional bonuses: Medium=%v, Easy=%v",
			scoreMedium, scoreEasy)
	}

	// Medium and Hard should be equal (both use same evaluation components)
	if scoreMedium != scoreHard {
		t.Errorf("Medium and Hard should evaluate the same: Medium=%v, Hard=%v",
			scoreMedium, scoreHard)
	}

	// Test with starting position - should be close to balanced for all
	boardStart := engine.NewBoard()
	easyStart := evaluate(boardStart, Easy)
	mediumStart := evaluate(boardStart, Medium)
	hardStart := evaluate(boardStart, Hard)

	if math.Abs(easyStart) > 0.01 {
		t.Errorf("Easy: Starting position should be ~0, got %v", easyStart)
	}

	// Medium/Hard might not be exactly 0 due to mobility (White moves first)
	if math.Abs(mediumStart) > 3.0 {
		t.Errorf("Medium: Starting position should be close to 0, got %v", mediumStart)
	}

	if math.Abs(hardStart) > 3.0 {
		t.Errorf("Hard: Starting position should be close to 0, got %v", hardStart)
	}
}

func TestEvaluatePiecePositions_AllPieces(t *testing.T) {
	// Test that all piece types get evaluated correctly

	tests := []struct {
		name     string
		fen      string
		minScore float64
		maxScore float64
	}{
		{
			name:     "WhitePawnRank7",
			fen:      "8/4P3/8/8/8/8/8/8 w - - 0 1",
			minScore: 0.6, // Should get a good bonus near promotion (rank 7)
			maxScore: 0.8,
		},
		{
			name:     "BlackPawnRank2",
			fen:      "8/8/8/8/8/8/4p3/8 w - - 0 1",
			minScore: -0.8, // Black pawn near promotion (negative, flipped to rank 7)
			maxScore: -0.6,
		},
		{
			name:     "WhiteRookRank7",
			fen:      "8/4R3/8/8/8/8/8/8 w - - 0 1",
			minScore: 0.2, // 7th rank bonus
			maxScore: 0.3,
		},
		{
			name:     "WhiteBishopCenter",
			fen:      "8/8/8/3B4/8/8/8/8 w - - 0 1",
			minScore: 0.05,
			maxScore: 0.15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := engine.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to parse FEN: %v", err)
			}

			score := evaluatePiecePositions(board, 0.5)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("%s: evaluatePiecePositions() = %v, want between %v and %v",
					tt.name, score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestEvaluateKingSafety(t *testing.T) {
	// Test king safety evaluation

	// Test 1: King with complete pawn shield should be safer
	fenGoodShield := "4k3/ppp5/8/8/8/8/PPP5/4K3 w - - 0 1"
	boardGood, err := engine.FromFEN(fenGoodShield)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	fenNoShield := "4k3/8/8/8/8/8/8/4K3 w - - 0 1"
	boardBad, err := engine.FromFEN(fenNoShield)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	safetyGood := evaluateKingSafety(boardGood)
	safetyBad := evaluateKingSafety(boardBad)

	// King with pawn shield should have better (less negative) score
	if safetyGood < safetyBad {
		t.Errorf("King with pawn shield should be safer: good=%v, bad=%v",
			safetyGood, safetyBad)
	}
}

func TestFindKing(t *testing.T) {
	// Test finding kings on the board

	// Test 1: Standard position with both kings
	board := engine.NewBoard()

	whiteKingSq := findKing(board, engine.White)
	if whiteKingSq != 4 { // e1 = 4
		t.Errorf("White king should be at square 4 (e1), got %v", whiteKingSq)
	}

	blackKingSq := findKing(board, engine.Black)
	if blackKingSq != 60 { // e8 = 60
		t.Errorf("Black king should be at square 60 (e8), got %v", blackKingSq)
	}

	// Test 2: Custom position
	fen := "4k3/8/8/8/8/8/8/4K3 w - - 0 1"
	board2, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	whiteKingSq2 := findKing(board2, engine.White)
	if whiteKingSq2 != 4 { // e1 = 4
		t.Errorf("White king should be at square 4 (e1), got %v", whiteKingSq2)
	}

	blackKingSq2 := findKing(board2, engine.Black)
	if blackKingSq2 != 60 { // e8 = 60
		t.Errorf("Black king should be at square 60 (e8), got %v", blackKingSq2)
	}
}

func TestEvaluatePawnShield(t *testing.T) {
	// Test pawn shield detection specifically

	// Test 1: Complete pawn shield (3 pawns)
	fenComplete := "8/8/8/8/8/8/PPP5/1K6 w - - 0 1"
	boardComplete, err := engine.FromFEN(fenComplete)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	whiteKingSq := findKing(boardComplete, engine.White)
	penaltyComplete := evaluatePawnShield(boardComplete, whiteKingSq, engine.White)

	// Should have 0 penalty (all 3 pawns present)
	if math.Abs(penaltyComplete) > 0.01 {
		t.Errorf("Complete pawn shield should have no penalty, got %v", penaltyComplete)
	}

	// Test 2: Partial pawn shield (1 pawn)
	fenPartial := "8/8/8/8/8/8/P7/1K6 w - - 0 1"
	boardPartial, err := engine.FromFEN(fenPartial)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	whiteKingSq2 := findKing(boardPartial, engine.White)
	penaltyPartial := evaluatePawnShield(boardPartial, whiteKingSq2, engine.White)

	// Should have penalty for 2 missing pawns
	expectedPenalty := 2.0 * 0.3 // 2 missing pawns * 0.3
	if math.Abs(penaltyPartial-expectedPenalty) > 0.01 {
		t.Errorf("Partial pawn shield should have penalty %v, got %v", expectedPenalty, penaltyPartial)
	}

	// Test 3: No pawn shield
	fenNone := "8/8/8/8/8/8/8/1K6 w - - 0 1"
	boardNone, err := engine.FromFEN(fenNone)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	whiteKingSq3 := findKing(boardNone, engine.White)
	penaltyNone := evaluatePawnShield(boardNone, whiteKingSq3, engine.White)

	// Should have penalty for 3 missing pawns
	expectedPenaltyNone := 3.0 * 0.3 // 3 missing pawns * 0.3
	if math.Abs(penaltyNone-expectedPenaltyNone) > 0.01 {
		t.Errorf("No pawn shield should have penalty %v, got %v", expectedPenaltyNone, penaltyNone)
	}
}

func TestEvaluateOpenFilesNearKing(t *testing.T) {
	// Test open file penalty

	// Test 1: No open files (all files have pawns)
	fenClosed := "8/8/8/8/8/ppp5/PPP5/1K1k4 w - - 0 1"
	boardClosed, err := engine.FromFEN(fenClosed)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	whiteKingSq := findKing(boardClosed, engine.White)
	penaltyClosed := evaluateOpenFilesNearKing(boardClosed, whiteKingSq, engine.White)

	// Should have 0 penalty (no open files)
	if math.Abs(penaltyClosed) > 0.01 {
		t.Errorf("Closed files should have no penalty, got %v", penaltyClosed)
	}

	// Test 2: One open file near king
	fenOneOpen := "8/8/8/8/8/8/PP6/1K6 w - - 0 1"
	boardOneOpen, err := engine.FromFEN(fenOneOpen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	whiteKingSq2 := findKing(boardOneOpen, engine.White)
	penaltyOneOpen := evaluateOpenFilesNearKing(boardOneOpen, whiteKingSq2, engine.White)

	// Should have penalty for 1 open file (c-file is open)
	expectedPenalty := 0.25
	if math.Abs(penaltyOneOpen-expectedPenalty) > 0.01 {
		t.Errorf("One open file should have penalty %v, got %v", expectedPenalty, penaltyOneOpen)
	}

	// Test 3: All files open near king
	fenAllOpen := "8/8/8/8/8/8/8/1K6 w - - 0 1"
	boardAllOpen, err := engine.FromFEN(fenAllOpen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	whiteKingSq3 := findKing(boardAllOpen, engine.White)
	penaltyAllOpen := evaluateOpenFilesNearKing(boardAllOpen, whiteKingSq3, engine.White)

	// Should have penalty for 3 open files (a, b, c files all open)
	expectedPenaltyAll := 3.0 * 0.25
	if math.Abs(penaltyAllOpen-expectedPenaltyAll) > 0.01 {
		t.Errorf("All open files should have penalty %v, got %v", expectedPenaltyAll, penaltyAllOpen)
	}
}

func TestEvaluateAttackersInKingZone(t *testing.T) {
	// Test attacker detection in king zone

	// Test 1: No attackers in king zone
	fenSafe := "4k3/8/8/8/8/8/8/4K3 w - - 0 1"
	boardSafe, err := engine.FromFEN(fenSafe)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	whiteKingSq := findKing(boardSafe, engine.White)
	penaltySafe := evaluateAttackersInKingZone(boardSafe, whiteKingSq, engine.White)

	// Should have 0 penalty (no attackers)
	if math.Abs(penaltySafe) > 0.01 {
		t.Errorf("No attackers should have no penalty, got %v", penaltySafe)
	}

	// Test 2: Enemy queen near king (attacking multiple squares)
	fenDanger := "4k3/8/8/8/8/8/8/q3K3 w - - 0 1"
	boardDanger, err := engine.FromFEN(fenDanger)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	whiteKingSq2 := findKing(boardDanger, engine.White)
	penaltyDanger := evaluateAttackersInKingZone(boardDanger, whiteKingSq2, engine.White)

	// Should have penalty (queen attacks multiple squares around king)
	if penaltyDanger <= 0 {
		t.Errorf("Enemy queen should create penalty, got %v", penaltyDanger)
	}

	// Test 3: Enemy rook on open file near king
	fenRook := "4k3/8/8/8/8/8/8/R3K3 w - - 0 1"
	boardRook, err := engine.FromFEN(fenRook)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	blackKingSq := findKing(boardRook, engine.Black)
	penaltyRook := evaluateAttackersInKingZone(boardRook, blackKingSq, engine.Black)

	// Black king should be safe (White rook far away on rank 1)
	// No white pieces attacking squares around black king on e8
	if penaltyRook != 0 {
		t.Errorf("Black king should be safe from white rook, got penalty %v", penaltyRook)
	}
}

func TestEvaluate_KingSafetyOnlyHard(t *testing.T) {
	// Verify king safety only affects Hard difficulty
	// Medium and Hard should differ when king safety matters

	// Position where White king is exposed but Black king has pawn shield
	// Black king on e7, pawns on rank 6 (d6, e6, f6) provide shield
	// White king on e1, no pawns on rank 2
	fenAsymmetric := "8/4k3/3ppp2/8/8/8/8/4K3 w - - 0 1"
	board, err := engine.FromFEN(fenAsymmetric)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	// First, verify that evaluateKingSafety returns non-zero for this position
	kingSafetyScore := evaluateKingSafety(board)
	if math.Abs(kingSafetyScore) < 0.01 {
		t.Errorf("evaluateKingSafety should return non-zero for asymmetric position, got %v", kingSafetyScore)
	}

	scoreMedium := evaluate(board, Medium)
	scoreHard := evaluate(board, Hard)

	// Hard should penalize White's exposed king (making score more negative)
	// Since Black has a pawn shield but White doesn't
	if scoreHard >= scoreMedium {
		t.Errorf("Hard bot should penalize exposed White king: Hard=%v, Medium=%v, kingSafety=%v",
			scoreHard, scoreMedium, kingSafetyScore)
	}

	// Test with a position where both kings have good pawn shields
	fenBothSafe := "4k3/ppp5/8/8/8/8/PPP5/4K3 w - - 0 1"
	boardBothSafe, err := engine.FromFEN(fenBothSafe)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	scoreMediumBothSafe := evaluate(boardBothSafe, Medium)
	scoreHardBothSafe := evaluate(boardBothSafe, Hard)

	// When both kings are equally safe, the difference should be smaller
	diffAsymmetric := math.Abs(scoreHard - scoreMedium)
	diffBothSafe := math.Abs(scoreHardBothSafe - scoreMediumBothSafe)

	if diffBothSafe >= diffAsymmetric {
		t.Errorf("Asymmetric king safety should create larger diff than symmetric: asymmetric=%v, bothSafe=%v",
			diffAsymmetric, diffBothSafe)
	}
}

func TestComputeGamePhase_StartingPosition(t *testing.T) {
	// Starting position has full material (63.0), should return 1.0
	board := engine.NewBoard()
	phase := computeGamePhase(board)

	if math.Abs(phase-1.0) > 0.01 {
		t.Errorf("Starting position: computeGamePhase() = %v, want 1.0", phase)
	}
}

func TestComputeGamePhase_BareKings(t *testing.T) {
	// Only kings on the board - should return 0.0
	fen := "4k3/8/8/8/8/8/8/4K3 w - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	phase := computeGamePhase(board)
	if phase != 0.0 {
		t.Errorf("Bare kings: computeGamePhase() = %v, want 0.0", phase)
	}
}

func TestComputeGamePhase_KingsAndPawns(t *testing.T) {
	// Kings and pawns only - no non-pawn material, should return 0.0
	fen := "4k3/pppppppp/8/8/8/8/PPPPPPPP/4K3 w - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	phase := computeGamePhase(board)
	if phase != 0.0 {
		t.Errorf("Kings and pawns: computeGamePhase() = %v, want 0.0", phase)
	}
}

func TestComputeGamePhase_MinorPieceAboveThreshold(t *testing.T) {
	// Material just above threshold: 16 + 3 = 19 (threshold material + one knight)
	// Use 2 rooks (10) + 2 bishops (6.5) = 16.5 which is at threshold,
	// then add a knight (3) = 19.5 total non-pawn material.
	// phase = (19.5 - 16) / (63 - 16) = 3.5 / 47 ~ 0.0745
	// Position: White has Rook, Bishop, Knight; Black has Rook, Bishop
	// White: R(5) + B(3.25) + N(3) = 11.25; Black: R(5) + B(3.25) = 8.25; Total = 19.5
	fen := "1rb1k3/8/8/8/8/8/8/1RB1KN2 w - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	phase := computeGamePhase(board)

	// Expected: (19.5 - 16) / (63 - 16) = 3.5/47 ~ 0.0745
	expectedPhase := (19.5 - endgameThreshold) / (totalStartingMaterial - endgameThreshold)
	if math.Abs(phase-expectedPhase) > 0.01 {
		t.Errorf("Minor piece above threshold: computeGamePhase() = %v, want ~%v", phase, expectedPhase)
	}

	// Should be a small positive value
	if phase <= 0.0 {
		t.Errorf("Phase should be positive above threshold, got %v", phase)
	}
}

func TestComputeGamePhase_HalfMaterial(t *testing.T) {
	// Approximately half material to get phase ~0.5
	// Need material = endgameThreshold + (totalStartingMaterial - endgameThreshold) * 0.5
	// = 16 + 47 * 0.5 = 39.5
	// White: Q(9) + 2R(10) + B(3.25) = 22.25; Black: Q(9) + R(5) + B(3.25) = 17.25
	// Total = 39.5
	fen := "1r1qkb2/8/8/8/8/8/8/1RRQKB2 w - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	phase := computeGamePhase(board)

	// Expected: (39.5 - 16) / (63 - 16) = 23.5/47 = 0.5
	if math.Abs(phase-0.5) > 0.01 {
		t.Errorf("Half material: computeGamePhase() = %v, want ~0.5", phase)
	}
}

func TestCountNonPawnMaterial(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		expected float64
	}{
		{
			name:     "StartingPosition",
			fen:      "", // Use NewBoard
			expected: 63.0,
		},
		{
			name:     "BareKings",
			fen:      "4k3/8/8/8/8/8/8/4K3 w - - 0 1",
			expected: 0.0,
		},
		{
			name:     "KingsAndPawns",
			fen:      "4k3/pppppppp/8/8/8/8/PPPPPPPP/4K3 w - - 0 1",
			expected: 0.0,
		},
		{
			name:     "OneWhiteQueen",
			fen:      "4k3/8/8/8/8/8/8/3QK3 w - - 0 1",
			expected: 9.0,
		},
		{
			name:     "OneBlackRook",
			fen:      "r3k3/8/8/8/8/8/8/4K3 w - - 0 1",
			expected: 5.0,
		},
		{
			name:     "MixedPieces",
			fen:      "rnb1k3/8/8/8/8/8/8/RNB1K3 w - - 0 1",
			// White: R(5) + N(3) + B(3.25) = 11.25
			// Black: R(5) + N(3) + B(3.25) = 11.25
			// Total = 22.5
			expected: 22.5,
		},
		{
			name:     "PawnsDoNotCount",
			fen:      "4k3/pppppppp/8/8/8/8/PPPPPPPP/4K3 w - - 0 1",
			expected: 0.0,
		},
		{
			name:     "AllQueensAndRooks",
			fen:      "r2qk2r/8/8/8/8/8/8/R2QK2R w - - 0 1",
			// White: R(5) + Q(9) + R(5) = 19
			// Black: R(5) + Q(9) + R(5) = 19
			// Total = 38
			expected: 38.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var board *engine.Board
			if tt.fen == "" {
				board = engine.NewBoard()
			} else {
				var err error
				board, err = engine.FromFEN(tt.fen)
				if err != nil {
					t.Fatalf("Failed to parse FEN: %v", err)
				}
			}

			material := countNonPawnMaterial(board)
			if math.Abs(material-tt.expected) > 0.01 {
				t.Errorf("%s: countNonPawnMaterial() = %v, want %v",
					tt.name, material, tt.expected)
			}
		})
	}
}

func TestEvaluateKingSafety_Symmetry(t *testing.T) {
	// Test that king safety evaluation is symmetric for both colors

	// Position with both kings having no pawn shield
	fenNoShield := "4k3/8/8/8/8/8/8/4K3 w - - 0 1"
	boardNoShield, err := engine.FromFEN(fenNoShield)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	safetyNoShield := evaluateKingSafety(boardNoShield)

	// Both kings should have same penalty, so net should be ~0
	if math.Abs(safetyNoShield) > 0.1 {
		t.Errorf("Symmetric position should have ~0 king safety score, got %v", safetyNoShield)
	}

	// Position with both kings having good pawn shields
	fenGoodShields := "4k3/ppp5/8/8/8/8/PPP5/4K3 w - - 0 1"
	boardGoodShields, err := engine.FromFEN(fenGoodShields)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	safetyGoodShields := evaluateKingSafety(boardGoodShields)

	// Both kings should have similar safety, so net should be ~0
	if math.Abs(safetyGoodShields) > 0.1 {
		t.Errorf("Symmetric position with shields should have ~0 king safety score, got %v", safetyGoodShields)
	}
}

func TestKingPhaseInterpolation_EndgameCenter(t *testing.T) {
	// King in center (e4) should score higher when phase=0.0 (endgame)
	// than when phase=1.0 (middlegame), because the endgame table rewards
	// centralization while the middlegame table penalizes exposed kings.
	// White King on e4 = square 28
	// kingEndgameTable[28] = 0.2 (rewards center)
	// kingMiddlegameTable[28] = -0.4 (penalizes center)
	fen := "8/8/8/8/4K3/8/8/8 w - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	scoreEndgame := evaluatePiecePositions(board, 0.0)   // Pure endgame
	scoreMiddlegame := evaluatePiecePositions(board, 1.0) // Pure middlegame

	// In endgame, center king gets positive bonus (+0.2 from endgame table)
	// In middlegame, center king gets penalty (-0.4 from middlegame table)
	if scoreEndgame <= scoreMiddlegame {
		t.Errorf("King in center should score higher in endgame than middlegame: endgame=%v, middlegame=%v",
			scoreEndgame, scoreMiddlegame)
	}
}

func TestKingPhaseInterpolation_MiddlegameCastled(t *testing.T) {
	// King on g1 (castled position) should score higher when phase=1.0 (middlegame)
	// than when phase=0.0 (endgame), because the middlegame table rewards castled
	// positions while the endgame table penalizes edge squares.
	// White King on g1 = square 6
	// kingMiddlegameTable[6] = 0.3 (rewards castled position)
	// kingEndgameTable[6] = -0.4 (penalizes edge in endgame)
	fen := "8/8/8/8/8/8/8/6K1 w - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	scoreMiddlegame := evaluatePiecePositions(board, 1.0) // Pure middlegame
	scoreEndgame := evaluatePiecePositions(board, 0.0)    // Pure endgame

	// In middlegame, castled king gets bonus (+0.3 from middlegame table)
	// In endgame, edge king gets penalty (-0.4 from endgame table)
	if scoreMiddlegame <= scoreEndgame {
		t.Errorf("Castled king should score higher in middlegame than endgame: middlegame=%v, endgame=%v",
			scoreMiddlegame, scoreEndgame)
	}
}

func TestKingPhaseInterpolation_HalfPhase(t *testing.T) {
	// At phase=0.5, the king bonus should be the average of the middlegame
	// and endgame table values.
	// White King on g1 = square 6
	// kingMiddlegameTable[6] = 0.3
	// kingEndgameTable[6] = -0.4
	// Expected at phase=0.5: 0.5*0.3 + 0.5*(-0.4) = 0.15 - 0.2 = -0.05
	fen := "8/8/8/8/8/8/8/6K1 w - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	scoreHalf := evaluatePiecePositions(board, 0.5)

	mgBonus := kingMiddlegameTable[6]  // 0.3
	egBonus := kingEndgameTable[6]     // -0.4
	expectedBonus := 0.5*mgBonus + 0.5*egBonus // -0.05

	if math.Abs(scoreHalf-expectedBonus) > 0.001 {
		t.Errorf("At phase=0.5, king bonus should be average of mg and eg tables: got %v, want %v (mg=%v, eg=%v)",
			scoreHalf, expectedBonus, mgBonus, egBonus)
	}
}

// Passed pawn tests

func TestIsPassedPawn_IsolatedPassed(t *testing.T) {
	// White pawn on e5 with no Black pawns on d, e, f files -> is passed
	// FEN: White King on a1, Black King on h8, White Pawn on e5
	fen := "7k/8/8/4P3/8/8/8/K7 w - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	// e5 = rank 4 (0-indexed), file 4 = 4*8 + 4 = 36
	sq := 36
	if !isPassedPawn(board, sq, engine.White) {
		t.Errorf("White pawn on e5 with no blocking pawns should be passed")
	}
}

func TestIsPassedPawn_BlockedSameFile(t *testing.T) {
	// White pawn on e4 with Black pawn on e6 -> NOT passed
	// FEN: White King on a1, Black King on h8, White Pawn on e4, Black Pawn on e6
	fen := "7k/8/4p3/8/4P3/8/8/K7 w - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	// e4 = rank 3, file 4 = 3*8 + 4 = 28
	sq := 28
	if isPassedPawn(board, sq, engine.White) {
		t.Errorf("White pawn on e4 with Black pawn on e6 should NOT be passed")
	}
}

func TestIsPassedPawn_BlockedAdjacentFile(t *testing.T) {
	// White pawn on e4 with Black pawn on d5 -> NOT passed
	// FEN: White King on a1, Black King on h8, White Pawn on e4, Black Pawn on d5
	fen := "7k/8/8/3p4/4P3/8/8/K7 w - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	// e4 = rank 3, file 4 = 3*8 + 4 = 28
	sq := 28
	if isPassedPawn(board, sq, engine.White) {
		t.Errorf("White pawn on e4 with Black pawn on d5 should NOT be passed")
	}
}

func TestIsPassedPawn_BlackPawn(t *testing.T) {
	// Black pawn on d4 with no White pawns blocking -> is passed
	// FEN: White King on a1, Black King on h8, Black Pawn on d4
	fen := "7k/8/8/8/3p4/8/8/K7 w - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	// d4 = rank 3, file 3 = 3*8 + 3 = 27
	sq := 27
	if !isPassedPawn(board, sq, engine.Black) {
		t.Errorf("Black pawn on d4 with no blocking pawns should be passed")
	}
}

func TestEvaluatePassedPawns_SingleWhite(t *testing.T) {
	// Single White passed pawn on e6 -> positive bonus
	// FEN: White King on a1, Black King on h8, White Pawn on e6
	fen := "7k/8/4P3/8/8/8/8/K7 w - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	// Phase 1.0 (opening) means phaseMultiplier = 1.0
	score := evaluatePassedPawns(board, 1.0)
	if score <= 0 {
		t.Errorf("Single White passed pawn should give positive score, got %v", score)
	}

	// e6 is rank 5 (0-indexed), so bonus should be passedPawnBonus[5] = 1.0
	expectedBonus := passedPawnBonus[5] * 1.0 // 1.0 * 1.0 = 1.0
	if math.Abs(score-expectedBonus) > 0.01 {
		t.Errorf("White passed pawn on e6: expected %v, got %v", expectedBonus, score)
	}
}

func TestEvaluatePassedPawns_SingleBlack(t *testing.T) {
	// Single Black passed pawn on d3 -> negative score (from White's perspective)
	// FEN: White King on a1, Black King on h8, Black Pawn on d3
	fen := "7k/8/8/8/8/3p4/8/K7 w - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	// Phase 1.0 (opening) means phaseMultiplier = 1.0
	score := evaluatePassedPawns(board, 1.0)
	if score >= 0 {
		t.Errorf("Single Black passed pawn should give negative score, got %v", score)
	}

	// d3 is rank 2, flipped rank = 7-2 = 5, so bonus = passedPawnBonus[5] = 1.0
	expectedBonus := -passedPawnBonus[5] * 1.0 // -1.0 * 1.0 = -1.0
	if math.Abs(score-expectedBonus) > 0.01 {
		t.Errorf("Black passed pawn on d3: expected %v, got %v", expectedBonus, score)
	}
}

func TestEvaluatePassedPawns_RankBonus(t *testing.T) {
	// Advanced pawn (rank 6) gets higher bonus than rank 3
	// e7 = rank 6 (0-indexed)
	fenAdvanced := "7k/4P3/8/8/8/8/8/K7 w - - 0 1"
	boardAdvanced, err := engine.FromFEN(fenAdvanced)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	// e4 = rank 3 (0-indexed)
	fenEarly := "7k/8/8/8/4P3/8/8/K7 w - - 0 1"
	boardEarly, err := engine.FromFEN(fenEarly)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	scoreAdvanced := evaluatePassedPawns(boardAdvanced, 1.0)
	scoreEarly := evaluatePassedPawns(boardEarly, 1.0)

	if scoreAdvanced <= scoreEarly {
		t.Errorf("Advanced pawn (e7, rank 6) should score higher than early pawn (e4, rank 3): advanced=%v, early=%v",
			scoreAdvanced, scoreEarly)
	}

	// Verify specific values: e7=rank6 -> bonus=1.5, e4=rank3 -> bonus=0.2
	expectedAdvanced := passedPawnBonus[6] * 1.0 // 1.5
	expectedEarly := passedPawnBonus[3] * 1.0    // 0.2
	if math.Abs(scoreAdvanced-expectedAdvanced) > 0.01 {
		t.Errorf("Advanced pawn bonus: expected %v, got %v", expectedAdvanced, scoreAdvanced)
	}
	if math.Abs(scoreEarly-expectedEarly) > 0.01 {
		t.Errorf("Early pawn bonus: expected %v, got %v", expectedEarly, scoreEarly)
	}
}

func TestEvaluatePassedPawns_EndgameAmplification(t *testing.T) {
	// Same pawn scores higher with phase=0.0 (endgame) than phase=1.0 (opening)
	// FEN: White King on a1, Black King on h8, White Pawn on e6
	fen := "7k/8/4P3/8/8/8/8/K7 w - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	scoreOpening := evaluatePassedPawns(board, 1.0) // phaseMultiplier = 1.0
	scoreEndgame := evaluatePassedPawns(board, 0.0) // phaseMultiplier = 2.0

	if scoreEndgame <= scoreOpening {
		t.Errorf("Passed pawn should score higher in endgame than opening: endgame=%v, opening=%v",
			scoreEndgame, scoreOpening)
	}

	// Verify the endgame bonus is exactly double the opening bonus
	if math.Abs(scoreEndgame-2*scoreOpening) > 0.01 {
		t.Errorf("Endgame score should be 2x opening score: endgame=%v, opening=%v", scoreEndgame, scoreOpening)
	}
}
