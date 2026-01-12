package ui

import (
	"strings"
	"testing"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// TestFormatSAN_SimplePawnMoves tests basic pawn moves
func TestFormatSAN_SimplePawnMoves(t *testing.T) {
	board := engine.NewBoard()

	tests := []struct {
		name     string
		moveStr  string
		expected string
	}{
		{"e2-e4", "e2e4", "e4"},
		{"d2-d4", "d2d4", "d4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			move, err := engine.ParseMove(tt.moveStr)
			if err != nil {
				t.Fatalf("Failed to parse move %s: %v", tt.moveStr, err)
			}

			san := FormatSAN(board, move)
			if san != tt.expected {
				t.Errorf("FormatSAN(%s) = %s, want %s", tt.moveStr, san, tt.expected)
			}
		})
	}
}

// TestFormatSAN_PawnCaptures tests pawn captures
func TestFormatSAN_PawnCaptures(t *testing.T) {
	// Set up a position where pawns can capture
	// 1. e4 e5 2. Nf3 Nc6 3. d4 exd4
	testBoard := engine.NewBoard()
	moves := []string{"e2e4", "e7e5", "g1f3", "b8c6", "d2d4"}

	for _, moveStr := range moves {
		move, _ := engine.ParseMove(moveStr)
		testBoard.MakeMove(move)
	}

	// Now Black can capture with exd4
	move, err := engine.ParseMove("e5d4")
	if err != nil {
		t.Fatalf("Failed to parse move: %v", err)
	}

	san := FormatSAN(testBoard, move)
	if san != "exd4" {
		t.Errorf("FormatSAN(exd4) = %s, want exd4", san)
	}
}

// TestFormatSAN_PieceMove tests basic piece moves
func TestFormatSAN_PieceMove(t *testing.T) {
	tests := []struct {
		name     string
		setup    []string // Moves to set up the position
		moveStr  string
		expected string
	}{
		{
			name:     "Nf3",
			setup:    []string{},
			moveStr:  "g1f3",
			expected: "Nf3",
		},
		{
			name:     "Bc4",
			setup:    []string{"e2e4", "e7e5"},
			moveStr:  "f1c4",
			expected: "Bc4",
		},
		{
			name:     "Qh5",
			setup:    []string{"e2e4", "e7e5"},
			moveStr:  "d1h5",
			expected: "Qh5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testBoard := engine.NewBoard()
			for _, setupMove := range tt.setup {
				move, _ := engine.ParseMove(setupMove)
				testBoard.MakeMove(move)
			}

			move, err := engine.ParseMove(tt.moveStr)
			if err != nil {
				t.Fatalf("Failed to parse move %s: %v", tt.moveStr, err)
			}

			san := FormatSAN(testBoard, move)
			if san != tt.expected {
				t.Errorf("FormatSAN(%s) = %s, want %s", tt.moveStr, san, tt.expected)
			}
		})
	}
}

// TestFormatSAN_PieceCaptures tests piece captures
func TestFormatSAN_PieceCaptures(t *testing.T) {
	// Set up position: 1. e4 e5 2. Nf3 Nc6 3. Bc4 Nf6 4. Ng5 d5 5. exd5 Nxd5
	testBoard := engine.NewBoard()
	moves := []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4", "g8f6", "f3g5", "d7d5", "e4d5"}

	for _, moveStr := range moves {
		move, _ := engine.ParseMove(moveStr)
		testBoard.MakeMove(move)
	}

	// Black's knight captures on d5
	move, err := engine.ParseMove("f6d5")
	if err != nil {
		t.Fatalf("Failed to parse move: %v", err)
	}

	san := FormatSAN(testBoard, move)
	if san != "Nxd5" {
		t.Errorf("FormatSAN(Nxd5) = %s, want Nxd5", san)
	}
}

// TestFormatSAN_Castling tests castling notation
func TestFormatSAN_Castling(t *testing.T) {
	tests := []struct {
		name     string
		setup    []string // Moves to set up the position
		moveStr  string
		expected string
	}{
		{
			name:     "White kingside castling",
			setup:    []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4", "g8f6"},
			moveStr:  "e1g1",
			expected: "O-O",
		},
		{
			name:     "White queenside castling",
			setup:    []string{"d2d4", "d7d5", "b1c3", "b8c6", "c1f4", "c8f5", "d1d3", "d8d7"},
			moveStr:  "e1c1",
			expected: "O-O-O",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testBoard := engine.NewBoard()
			for _, setupMove := range tt.setup {
				move, _ := engine.ParseMove(setupMove)
				testBoard.MakeMove(move)
			}

			move, err := engine.ParseMove(tt.moveStr)
			if err != nil {
				t.Fatalf("Failed to parse move %s: %v", tt.moveStr, err)
			}

			san := FormatSAN(testBoard, move)
			if san != tt.expected {
				t.Errorf("FormatSAN(%s) = %s, want %s", tt.moveStr, san, tt.expected)
			}
		})
	}
}

// TestFormatSAN_Promotion tests pawn promotion
func TestFormatSAN_Promotion(t *testing.T) {
	// Set up a position where White can promote
	fen := "8/P7/8/8/8/8/8/4K2k w - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	tests := []struct {
		name     string
		moveStr  string
		expected string
	}{
		{"Promote to Queen", "a7a8q", "a8=Q"},
		{"Promote to Rook", "a7a8r", "a8=R"},
		{"Promote to Bishop", "a7a8b", "a8=B"},
		{"Promote to Knight", "a7a8n", "a8=N"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a copy of the board for each test
			testBoard := board.Copy()

			move, err := engine.ParseMove(tt.moveStr)
			if err != nil {
				t.Fatalf("Failed to parse move %s: %v", tt.moveStr, err)
			}

			san := FormatSAN(testBoard, move)
			// The move should end with checkmate symbol
			if !strings.HasPrefix(san, tt.expected) {
				t.Errorf("FormatSAN(%s) = %s, want prefix %s", tt.moveStr, san, tt.expected)
			}
		})
	}
}

// TestFormatSAN_PromotionWithCapture tests pawn promotion with capture
func TestFormatSAN_PromotionWithCapture(t *testing.T) {
	// Set up a position where White can promote with capture
	fen := "1n6/P7/8/8/8/8/8/4K2k w - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	move, err := engine.ParseMove("a7b8q")
	if err != nil {
		t.Fatalf("Failed to parse move: %v", err)
	}

	san := FormatSAN(board, move)
	if !strings.HasPrefix(san, "axb8=Q") {
		t.Errorf("FormatSAN(axb8=Q) = %s, want prefix axb8=Q", san)
	}
}

// TestFormatSAN_Check tests check notation
func TestFormatSAN_Check(t *testing.T) {
	// Set up a position where White can give check with discovered attack
	// Position where moving the knight gives check
	fen := "rnbqkb1r/pppp1ppp/5n2/4p3/3PP3/5N2/PPP2PPP/RNBQKB1R b KQkq - 0 3"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	// Bb4+ is check (bishop checks the king on e1)
	move, err := engine.ParseMove("f8b4")
	if err != nil {
		t.Fatalf("Failed to parse move: %v", err)
	}

	san := FormatSAN(board, move)
	if san != "Bb4+" {
		t.Errorf("FormatSAN(Bb4+) = %s, want Bb4+", san)
	}
}

// TestFormatSAN_Checkmate tests checkmate notation
func TestFormatSAN_Checkmate(t *testing.T) {
	// Set up a simple checkmate position
	fen := "6k1/5ppp/8/8/8/8/5PPP/4R1K1 w - - 0 1"
	board, err := engine.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	// Re8# is checkmate
	move, err := engine.ParseMove("e1e8")
	if err != nil {
		t.Fatalf("Failed to parse move: %v", err)
	}

	san := FormatSAN(board, move)
	if san != "Re8#" {
		t.Errorf("FormatSAN(Re8#) = %s, want Re8#", san)
	}
}

// TestFormatSAN_Disambiguation tests disambiguation
func TestFormatSAN_Disambiguation(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		moveStr  string
		expected string
	}{
		{
			name:     "File disambiguation - two knights can go to same square",
			fen:      "rnbqkb1r/pppppppp/5n2/8/8/3N1N2/PPPPPPPP/R1BQKB1R w KQkq - 0 1",
			moveStr:  "f3e5",
			expected: "Nfe5",
		},
		{
			name:     "Rank disambiguation - two rooks on same file",
			fen:      "4k3/8/8/8/8/8/4R3/4R1K1 w - - 0 1",
			moveStr:  "e1e4",
			expected: "R1e4",
		},
		{
			name:     "File and rank disambiguation",
			fen:      "4k3/8/8/3Q4/8/8/3Q4/3QK3 w - - 0 1",
			moveStr:  "d1d3",
			expected: "Q1d3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := engine.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to parse FEN: %v", err)
			}

			move, err := engine.ParseMove(tt.moveStr)
			if err != nil {
				t.Fatalf("Failed to parse move %s: %v", tt.moveStr, err)
			}

			san := FormatSAN(board, move)
			if san != tt.expected {
				t.Errorf("FormatSAN(%s) = %s, want %s", tt.moveStr, san, tt.expected)
			}
		})
	}
}

// TestFormatMoveHistory tests the move history formatting
func TestFormatMoveHistory(t *testing.T) {
	config := Config{
		UseUnicode:      false,
		ShowCoords:      true,
		UseColors:       false,
		ShowMoveHistory: true,
	}

	m := NewModel(config)
	m.board = engine.NewBoard()

	// Play some moves
	movesStr := []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4"}

	for _, moveStr := range movesStr {
		move, err := engine.ParseMove(moveStr)
		if err != nil {
			t.Fatalf("Failed to parse move %s: %v", moveStr, err)
		}

		m.board.MakeMove(move)
		m.moveHistory = append(m.moveHistory, move)
	}

	// Format the move history
	history := m.formatMoveHistory()

	// Check that it contains expected moves
	expectedMoves := []string{"1. e4 e5", "2. Nf3 Nc6", "3. Bc4"}
	for _, expected := range expectedMoves {
		if !strings.Contains(history, expected) {
			t.Errorf("Move history missing expected sequence: %s\nGot: %s", expected, history)
		}
	}

	// Check that it starts with "Move History:"
	if !strings.HasPrefix(history, "Move History:") {
		t.Errorf("Move history should start with 'Move History:', got: %s", history)
	}
}

// TestMoveHistoryEmpty tests that empty move history returns empty string
func TestMoveHistoryEmpty(t *testing.T) {
	config := Config{
		UseUnicode:      false,
		ShowCoords:      true,
		UseColors:       false,
		ShowMoveHistory: true,
	}

	m := NewModel(config)
	m.board = engine.NewBoard()

	history := m.formatMoveHistory()
	if history != "" {
		t.Errorf("Empty move history should return empty string, got: %s", history)
	}
}

// TestMoveHistoryReplayAccuracy tests that replaying moves produces correct SAN
func TestMoveHistoryReplayAccuracy(t *testing.T) {
	config := Config{
		UseUnicode:      false,
		ShowCoords:      true,
		UseColors:       false,
		ShowMoveHistory: true,
	}

	m := NewModel(config)
	m.board = engine.NewBoard()

	// Play Italian Game opening
	expectedSAN := []string{"e4", "e5", "Nf3", "Nc6", "Bc4", "Nf6", "d3", "Bc5"}
	movesStr := []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4", "g8f6", "d2d3", "f8c5"}

	for i, moveStr := range movesStr {
		move, err := engine.ParseMove(moveStr)
		if err != nil {
			t.Fatalf("Failed to parse move %s: %v", moveStr, err)
		}

		// Format before making the move
		san := FormatSAN(m.board, move)
		if san != expectedSAN[i] {
			t.Errorf("Move %d: expected %s, got %s", i+1, expectedSAN[i], san)
		}

		m.board.MakeMove(move)
		m.moveHistory = append(m.moveHistory, move)
	}

	// Verify the full history
	history := m.formatMoveHistory()
	expected := "Move History: 1. e4 e5 2. Nf3 Nc6 3. Bc4 Nf6 4. d3 Bc5"
	if history != expected {
		t.Errorf("Full move history mismatch\nExpected: %s\nGot: %s", expected, history)
	}
}
