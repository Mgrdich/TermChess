package engine

import (
	"fmt"
	"testing"
)

// TestPerft tests the perft function against known-correct results.
// Perft (Performance Test) counts all leaf nodes at a given depth.
// These are well-known test positions used by chess engines worldwide.
func TestPerft(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		depths   []int
		expected []uint64
	}{
		{
			name: "starting position",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			depths: []int{1, 2, 3, 4},
			expected: []uint64{
				20,      // depth 1
				400,     // depth 2
				8902,    // depth 3
				197281,  // depth 4
				// 4865609 is depth 5, takes longer so commented out for faster tests
			},
		},
		{
			name: "kiwipete position",
			fen:  "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
			depths: []int{1, 2, 3, 4},
			expected: []uint64{
				48,      // depth 1
				2039,    // depth 2
				97862,   // depth 3
				4085603, // depth 4
			},
		},
		{
			name:   "position 3",
			fen:    "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
			depths: []int{1, 2, 3, 4},
			expected: []uint64{
				14,    // depth 1
				191,   // depth 2
				2812,  // depth 3
				43238, // depth 4
			},
		},
		{
			name:   "position 4",
			fen:    "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
			depths: []int{1, 2, 3},
			expected: []uint64{
				6,    // depth 1
				264,  // depth 2
				9467, // depth 3
			},
		},
		{
			name: "position 5 (same as kiwipete but different move number)",
			fen:  "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
			depths: []int{1, 2, 3},
			expected: []uint64{
				44,    // depth 1
				1486,  // depth 2
				62379, // depth 3
				// 2103487 is depth 4, takes longer
			},
		},
		{
			name:   "position 6",
			fen:    "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
			depths: []int{1, 2, 3},
			expected: []uint64{
				46,    // depth 1
				2079,  // depth 2
				89890, // depth 3
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("Failed to parse FEN: %v", err)
			}

			for i, depth := range tt.depths {
				t.Run(fmt.Sprintf("depth %d", depth), func(t *testing.T) {
					result := board.Perft(depth)
					if result != tt.expected[i] {
						t.Errorf("Perft(%d) = %d, expected %d", depth, result, tt.expected[i])

						// If test fails, run Divide to help debug
						if depth <= 2 {
							t.Logf("Divide output for depth %d:", depth)
							divide := board.Divide(depth)
							total := uint64(0)
							for move, count := range divide {
								t.Logf("  %s: %d", move, count)
								total += count
							}
							t.Logf("  Total: %d", total)
						}
					}
				})
			}
		})
	}
}

// TestPerftStartingPosition is a focused test for the starting position
// to quickly verify basic move generation is working.
func TestPerftStartingPosition(t *testing.T) {
	board := NewBoard()

	tests := []struct {
		depth    int
		expected uint64
	}{
		{0, 1},
		{1, 20},
		{2, 400},
		{3, 8902},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("depth %d", tt.depth), func(t *testing.T) {
			result := board.Perft(tt.depth)
			if result != tt.expected {
				t.Errorf("Perft(%d) = %d, expected %d", tt.depth, result, tt.expected)
			}
		})
	}
}

// TestDivide tests the Divide function which provides per-move breakdown.
func TestDivide(t *testing.T) {
	board := NewBoard()

	t.Run("starting position depth 1", func(t *testing.T) {
		divide := board.Divide(1)

		// Should have 20 moves
		if len(divide) != 20 {
			t.Errorf("Divide(1) returned %d moves, expected 20", len(divide))
		}

		// Each move should have count 1 at depth 1
		total := uint64(0)
		for move, count := range divide {
			if count != 1 {
				t.Errorf("Move %s has count %d, expected 1", move, count)
			}
			total += count
		}

		if total != 20 {
			t.Errorf("Total nodes = %d, expected 20", total)
		}
	})

	t.Run("starting position depth 2", func(t *testing.T) {
		divide := board.Divide(2)

		// Should have 20 moves
		if len(divide) != 20 {
			t.Errorf("Divide(2) returned %d moves, expected 20", len(divide))
		}

		// Known counts for some starting moves
		expectedCounts := map[string]uint64{
			"a2a3": 20,
			"b2b3": 20,
			"c2c3": 20,
			"d2d3": 20,
			"e2e3": 20,
			"f2f3": 20,
			"g2g3": 20,
			"h2h3": 20,
			"a2a4": 20,
			"b2b4": 20,
			"c2c4": 20,
			"d2d4": 20,
			"e2e4": 20,
			"f2f4": 20,
			"g2g4": 20,
			"h2h4": 20,
			"b1a3": 20,
			"b1c3": 20,
			"g1f3": 20,
			"g1h3": 20,
		}

		for move, expectedCount := range expectedCounts {
			count, ok := divide[move]
			if !ok {
				t.Errorf("Expected move %s not found in divide", move)
			} else if count != expectedCount {
				t.Errorf("Move %s has count %d, expected %d", move, count, expectedCount)
			}
		}

		// Total should be 400
		total := uint64(0)
		for _, count := range divide {
			total += count
		}
		if total != 400 {
			t.Errorf("Total nodes = %d, expected 400", total)
		}
	})
}

// BenchmarkPerft benchmarks the perft function at various depths.
func BenchmarkPerft(b *testing.B) {
	board := NewBoard()

	benchmarks := []struct {
		depth int
	}{
		{1},
		{2},
		{3},
		{4},
	}

	for _, bm := range benchmarks {
		b.Run(fmt.Sprintf("depth_%d", bm.depth), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				board.Perft(bm.depth)
			}
		})
	}
}
