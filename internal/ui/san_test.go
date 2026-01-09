package ui

import (
	"testing"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// TestParseSAN_SimplePawnMoves tests parsing of simple pawn moves like "e4", "d5".
func TestParseSAN_SimplePawnMoves(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		san      string
		wantMove string
		wantErr  bool
	}{
		{
			name:     "white pawn e4 from start",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "e4",
			wantMove: "e2e4",
			wantErr:  false,
		},
		{
			name:     "white pawn d4 from start",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "d4",
			wantMove: "d2d4",
			wantErr:  false,
		},
		{
			name:     "white pawn a3 from start",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "a3",
			wantMove: "a2a3",
			wantErr:  false,
		},
		{
			name:     "white pawn h4 from start",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "h4",
			wantMove: "h2h4",
			wantErr:  false,
		},
		{
			name:     "black pawn e5 after white e4",
			fen:      "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
			san:      "e5",
			wantMove: "e7e5",
			wantErr:  false,
		},
		{
			name:     "black pawn d5 from start",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1",
			san:      "d5",
			wantMove: "d7d5",
			wantErr:  false,
		},
		{
			name:     "white pawn e3 single square",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "e3",
			wantMove: "e2e3",
			wantErr:  false,
		},
		{
			name:     "invalid pawn move - no pawn can reach",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "e5",
			wantMove: "",
			wantErr:  true,
		},
		{
			name:     "invalid square - off board",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "e9",
			wantMove: "",
			wantErr:  true,
		},
		{
			name:     "invalid file - i4",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "i4",
			wantMove: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := engine.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("failed to parse FEN: %v", err)
			}

			move, err := ParseSAN(board, tt.san)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if move.String() != tt.wantMove {
				t.Errorf("expected %s, got %s", tt.wantMove, move.String())
			}
		})
	}
}

// TestParseSAN_PawnCaptures tests parsing of pawn captures like "exd5", "axb3".
func TestParseSAN_PawnCaptures(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		san      string
		wantMove string
		wantErr  bool
	}{
		{
			name:     "white exd5 capture",
			fen:      "rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2",
			san:      "exd5",
			wantMove: "e4d5",
			wantErr:  false,
		},
		{
			name:     "black exd4 capture",
			fen:      "rnbqkbnr/pppp1ppp/8/4p3/3P4/8/PPP1PPPP/RNBQKBNR b KQkq d3 0 2",
			san:      "exd4",
			wantMove: "e5d4",
			wantErr:  false,
		},
		{
			name:     "white axb3 capture",
			fen:      "rnbqkbnr/1ppppppp/8/8/8/1p6/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "axb3",
			wantMove: "a2b3",
			wantErr:  false,
		},
		{
			name:     "white dxe5 capture",
			fen:      "rnbqkbnr/pppp1ppp/8/4p3/3P4/8/PPP1PPPP/RNBQKBNR w KQkq e6 0 2",
			san:      "dxe5",
			wantMove: "d4e5",
			wantErr:  false,
		},
		{
			name:     "invalid capture - no pawn on source file",
			fen:      "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 1",
			san:      "axb3",
			wantMove: "",
			wantErr:  true,
		},
		{
			name:     "invalid capture - pawn cannot capture there",
			fen:      "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 1",
			san:      "exf5",
			wantMove: "",
			wantErr:  true,
		},
		{
			name:     "en passant capture",
			fen:      "rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2",
			san:      "exd6",
			wantMove: "e5d6",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := engine.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("failed to parse FEN: %v", err)
			}

			move, err := ParseSAN(board, tt.san)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if move.String() != tt.wantMove {
				t.Errorf("expected %s, got %s", tt.wantMove, move.String())
			}
		})
	}
}

// TestParseSAN_PawnPromotions tests parsing of pawn promotions like "e8=Q", "a1=N".
func TestParseSAN_PawnPromotions(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		san      string
		wantMove string
		wantErr  bool
	}{
		{
			name:     "white e8=Q promotion",
			fen:      "3k4/4P3/8/8/8/8/8/4K3 w - - 0 1",
			san:      "e8=Q",
			wantMove: "e7e8q",
			wantErr:  false,
		},
		{
			name:     "white e8=R promotion",
			fen:      "3k4/4P3/8/8/8/8/8/4K3 w - - 0 1",
			san:      "e8=R",
			wantMove: "e7e8r",
			wantErr:  false,
		},
		{
			name:     "white e8=B promotion",
			fen:      "3k4/4P3/8/8/8/8/8/4K3 w - - 0 1",
			san:      "e8=B",
			wantMove: "e7e8b",
			wantErr:  false,
		},
		{
			name:     "white e8=N promotion",
			fen:      "3k4/4P3/8/8/8/8/8/4K3 w - - 0 1",
			san:      "e8=N",
			wantMove: "e7e8n",
			wantErr:  false,
		},
		{
			name:     "black a1=Q promotion",
			fen:      "4k3/8/8/8/8/8/p7/4K3 b - - 0 1",
			san:      "a1=Q",
			wantMove: "a2a1q",
			wantErr:  false,
		},
		{
			name:     "black h1=N promotion",
			fen:      "4k3/8/8/8/8/8/7p/4K3 b - - 0 1",
			san:      "h1=N",
			wantMove: "h2h1n",
			wantErr:  false,
		},
		{
			name:     "lowercase promotion - e8=q",
			fen:      "3k4/4P3/8/8/8/8/8/4K3 w - - 0 1",
			san:      "e8=q",
			wantMove: "e7e8q",
			wantErr:  false,
		},
		{
			name:     "invalid promotion piece - e8=K",
			fen:      "3k4/4P3/8/8/8/8/8/4K3 w - - 0 1",
			san:      "e8=K",
			wantMove: "",
			wantErr:  true,
		},
		{
			name:     "invalid promotion format - e8Q",
			fen:      "3k4/4P3/8/8/8/8/8/4K3 w - - 0 1",
			san:      "e8Q",
			wantMove: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := engine.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("failed to parse FEN: %v", err)
			}

			move, err := ParseSAN(board, tt.san)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if move.String() != tt.wantMove {
				t.Errorf("expected %s, got %s", tt.wantMove, move.String())
			}
		})
	}
}

// TestParseSAN_CombinedCapturePromotion tests parsing of combined capture and promotion like "exd8=Q".
func TestParseSAN_CombinedCapturePromotion(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		san      string
		wantMove string
		wantErr  bool
	}{
		{
			name:     "white exd8=Q capture promotion",
			fen:      "3nk3/4P3/8/8/8/8/8/4K3 w - - 0 1",
			san:      "exd8=Q",
			wantMove: "e7d8q",
			wantErr:  false,
		},
		{
			name:     "white exd8=R capture promotion",
			fen:      "3nk3/4P3/8/8/8/8/8/4K3 w - - 0 1",
			san:      "exd8=R",
			wantMove: "e7d8r",
			wantErr:  false,
		},
		{
			name:     "black axb1=Q capture promotion",
			fen:      "4k3/8/8/8/8/8/p7/1N2K3 b - - 0 1",
			san:      "axb1=Q",
			wantMove: "a2b1q",
			wantErr:  false,
		},
		{
			name:     "black hxg1=N capture promotion",
			fen:      "4k3/8/8/8/8/8/7p/4K1B1 b - - 0 1",
			san:      "hxg1=N",
			wantMove: "h2g1n",
			wantErr:  false,
		},
		{
			name:     "lowercase promotion - exd8=q",
			fen:      "3nk3/4P3/8/8/8/8/8/4K3 w - - 0 1",
			san:      "exd8=q",
			wantMove: "e7d8q",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := engine.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("failed to parse FEN: %v", err)
			}

			move, err := ParseSAN(board, tt.san)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if move.String() != tt.wantMove {
				t.Errorf("expected %s, got %s", tt.wantMove, move.String())
			}
		})
	}
}

// TestParseSAN_CheckSymbolsStripped tests that check/checkmate symbols are properly stripped.
func TestParseSAN_CheckSymbolsStripped(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		san      string
		wantMove string
		wantErr  bool
	}{
		{
			name:     "e4+ with check symbol",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "e4+",
			wantMove: "e2e4",
			wantErr:  false,
		},
		{
			name:     "e4# with checkmate symbol",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "e4#",
			wantMove: "e2e4",
			wantErr:  false,
		},
		{
			name:     "exd5+ with check",
			fen:      "rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2",
			san:      "exd5+",
			wantMove: "e4d5",
			wantErr:  false,
		},
		{
			name:     "e8=Q# with checkmate",
			fen:      "3k4/4P3/8/8/8/8/8/4K3 w - - 0 1",
			san:      "e8=Q#",
			wantMove: "e7e8q",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := engine.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("failed to parse FEN: %v", err)
			}

			move, err := ParseSAN(board, tt.san)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if move.String() != tt.wantMove {
				t.Errorf("expected %s, got %s", tt.wantMove, move.String())
			}
		})
	}
}

// TestParseSAN_InvalidPawnMoves tests error cases for invalid pawn moves.
func TestParseSAN_InvalidPawnMoves(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		san  string
	}{
		{
			name: "empty string",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:  "",
		},
		{
			name: "only check symbol",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:  "+",
		},
		{
			name: "piece move - not supported",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:  "Nf3",
		},
		{
			name: "castling - not supported",
			fen:  "r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQkq - 0 1",
			san:  "O-O",
		},
		{
			name: "invalid square format - too long",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:  "e44",
		},
		{
			name: "invalid square format - too short",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:  "e",
		},
		{
			name: "invalid capture format - multiple x",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:  "exxd5",
		},
		{
			name: "invalid promotion format - multiple =",
			fen:  "4k3/4P3/8/8/8/8/8/4K3 w - - 0 1",
			san:  "e8==Q",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := engine.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("failed to parse FEN: %v", err)
			}

			_, err = ParseSAN(board, tt.san)

			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

// TestParseSAN_WrongTurn tests that moves for the wrong color are rejected.
func TestParseSAN_WrongTurn(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		san  string
	}{
		{
			name: "white's turn but trying black pawn move",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:  "e5",
		},
		{
			name: "black's turn but trying white pawn move",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1",
			san:  "e4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := engine.FromFEN(tt.fen)
			if err != nil {
				t.Fatalf("failed to parse FEN: %v", err)
			}

			_, err = ParseSAN(board, tt.san)

			if err == nil {
				t.Error("expected error for wrong turn, got nil")
			}
		})
	}
}
