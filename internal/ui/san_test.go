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

// TestParseSAN_KnightMoves tests parsing of knight moves.
func TestParseSAN_KnightMoves(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		san      string
		wantMove string
		wantErr  bool
	}{
		{
			name:     "white Nf3 from start",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "Nf3",
			wantMove: "g1f3",
			wantErr:  false,
		},
		{
			name:     "white Nc3 from start",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "Nc3",
			wantMove: "b1c3",
			wantErr:  false,
		},
		{
			name:     "white Nh3 from start",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "Nh3",
			wantMove: "g1h3",
			wantErr:  false,
		},
		{
			name:     "white Na3 from start",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "Na3",
			wantMove: "b1a3",
			wantErr:  false,
		},
		{
			name:     "black Nf6 after e4",
			fen:      "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
			san:      "Nf6",
			wantMove: "g8f6",
			wantErr:  false,
		},
		{
			name:     "black Nc6 after e4",
			fen:      "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
			san:      "Nc6",
			wantMove: "b8c6",
			wantErr:  false,
		},
		{
			name:     "knight move with check symbol stripped",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "Nf3+",
			wantMove: "g1f3",
			wantErr:  false,
		},
		{
			name:     "invalid knight move - square blocked",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "Ne2",
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

// TestParseSAN_BishopMoves tests parsing of bishop moves.
func TestParseSAN_BishopMoves(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		san      string
		wantMove string
		wantErr  bool
	}{
		{
			name:     "white Bc4 after e4",
			fen:      "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e3 0 1",
			san:      "Bc4",
			wantMove: "f1c4",
			wantErr:  false,
		},
		{
			name:     "white Bb5 after e4 Nc6 Nf3",
			fen:      "r1bqkbnr/pppppppp/2n5/8/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 0 3",
			san:      "Bb5",
			wantMove: "f1b5",
			wantErr:  false,
		},
		{
			name:     "white Be2 after e4",
			fen:      "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e3 0 1",
			san:      "Be2",
			wantMove: "f1e2",
			wantErr:  false,
		},
		{
			name:     "black Bc5 after e4 e5 Nf3",
			fen:      "rnbqkbnr/pppp1ppp/8/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 0 2",
			san:      "Bc5",
			wantMove: "f8c5",
			wantErr:  false,
		},
		{
			name:     "invalid bishop move - blocked by pawn",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "Bc4",
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

// TestParseSAN_RookMoves tests parsing of rook moves.
func TestParseSAN_RookMoves(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		san      string
		wantMove string
		wantErr  bool
	}{
		{
			name:     "white Ra3 - rook moves from a1",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/1PPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "Ra3",
			wantMove: "a1a3",
			wantErr:  false,
		},
		{
			name:     "white Rh3 - rook moves from h1",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPP1/RNBQKBNR w KQkq - 0 1",
			san:      "Rh3",
			wantMove: "h1h3",
			wantErr:  false,
		},
		{
			name:     "white Re1 - rook in center",
			fen:      "rnbqkbnr/pppppppp/8/8/8/4R3/PPPP1PPP/RNBQ1KN1 w kq - 0 1",
			san:      "Re1",
			wantMove: "e3e1",
			wantErr:  false,
		},
		{
			name:     "invalid rook move - blocked",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "Ra5",
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

// TestParseSAN_QueenMoves tests parsing of queen moves.
func TestParseSAN_QueenMoves(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		san      string
		wantMove string
		wantErr  bool
	}{
		{
			name:     "white Qh5 - scholar's mate setup",
			fen:      "rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2",
			san:      "Qh5",
			wantMove: "d1h5",
			wantErr:  false,
		},
		{
			name:     "white Qf3 - queen to f3",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPP1PPP/RNBQKBNR w KQkq - 0 1",
			san:      "Qf3",
			wantMove: "d1f3",
			wantErr:  false,
		},
		{
			name:     "black Qh4 - attacking move",
			fen:      "rnbqkbnr/pppp1ppp/8/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 0 2",
			san:      "Qh4",
			wantMove: "d8h4",
			wantErr:  false,
		},
		{
			name:     "invalid queen move - blocked",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "Qh5",
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

// TestParseSAN_KingMoves tests parsing of king moves.
func TestParseSAN_KingMoves(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		san      string
		wantMove string
		wantErr  bool
	}{
		{
			name:     "white Ke2 - king forward",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPP1PPP/RNBQKBNR w KQkq - 0 1",
			san:      "Ke2",
			wantMove: "e1e2",
			wantErr:  false,
		},
		{
			name:     "white Kf1 - king to f1",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQK1NR w KQkq - 0 1",
			san:      "Kf1",
			wantMove: "e1f1",
			wantErr:  false,
		},
		{
			name:     "black Ke7 - king forward",
			fen:      "rnbqkbnr/pppp1ppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1",
			san:      "Ke7",
			wantMove: "e8e7",
			wantErr:  false,
		},
		{
			name:     "invalid king move - too far",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "Ke3",
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

// TestParseSAN_PieceCaptures tests parsing of piece captures.
func TestParseSAN_PieceCaptures(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		san      string
		wantMove string
		wantErr  bool
	}{
		{
			name:     "white Nxe5 - knight captures",
			fen:      "rnbqkbnr/pppp1ppp/8/4p3/8/5N2/PPPPPPPP/RNBQKB1R w KQkq e6 0 2",
			san:      "Nxe5",
			wantMove: "f3e5",
			wantErr:  false,
		},
		{
			name:     "white Bxf7 - bishop captures (check)",
			fen:      "rnbqkbnr/pppp1ppp/8/4p3/2B1P3/8/PPPP1PPP/RNBQK1NR w KQkq - 0 3",
			san:      "Bxf7+",
			wantMove: "c4f7",
			wantErr:  false,
		},
		{
			name:     "black Nxe4 - knight captures",
			fen:      "rnbqkb1r/pppp1ppp/5n2/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 0 3",
			san:      "Nxe4",
			wantMove: "f6e4",
			wantErr:  false,
		},
		{
			name:     "white Rxe5 - rook captures",
			fen:      "rnbqkbnr/pppp1ppp/8/4p3/8/4R3/PPPP1PPP/RNBQKBN1 w Qkq - 0 2",
			san:      "Rxe5",
			wantMove: "e3e5",
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

// TestParseSAN_Castling tests parsing of castling moves.
func TestParseSAN_Castling(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		san      string
		wantMove string
		wantErr  bool
	}{
		{
			name:     "white O-O kingside castling",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQK2R w KQkq - 0 1",
			san:      "O-O",
			wantMove: "e1g1",
			wantErr:  false,
		},
		{
			name:     "white O-O-O queenside castling",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/R3KBNR w KQkq - 0 1",
			san:      "O-O-O",
			wantMove: "e1c1",
			wantErr:  false,
		},
		{
			name:     "black O-O kingside castling",
			fen:      "rnbqk2r/pppppppp/5n2/8/8/5N2/PPPPPPPP/RNBQKB1R b KQkq - 0 1",
			san:      "O-O",
			wantMove: "e8g8",
			wantErr:  false,
		},
		{
			name:     "black O-O-O queenside castling",
			fen:      "r3kbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1",
			san:      "O-O-O",
			wantMove: "e8c8",
			wantErr:  false,
		},
		{
			name:     "white 0-0 with zeros",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQK2R w KQkq - 0 1",
			san:      "0-0",
			wantMove: "e1g1",
			wantErr:  false,
		},
		{
			name:     "white 0-0-0 with zeros",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/R3KBNR w KQkq - 0 1",
			san:      "0-0-0",
			wantMove: "e1c1",
			wantErr:  false,
		},
		{
			name:     "castling with check symbol",
			fen:      "rnbqk2r/pppp1ppp/5n2/2b1p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 4",
			san:      "O-O+",
			wantMove: "e1g1",
			wantErr:  false,
		},
		{
			name:     "invalid castling - pieces in way",
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			san:      "O-O",
			wantMove: "",
			wantErr:  true,
		},
		{
			name:     "invalid castling - no rights",
			fen:      "rnbqkbnr/pppppppp/8/8/8/5N2/PPPPPPPP/RNBQKB1R w kq - 0 1",
			san:      "O-O",
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

// TestParseSAN_FullGameSequence tests a sequence of moves for a real game.
func TestParseSAN_FullGameSequence(t *testing.T) {
	board := engine.NewBoard()

	// Test: 1. e4 e5 2. Nf3 Nc6 3. Bc4
	moves := []struct {
		san      string
		wantMove string
	}{
		{"e4", "e2e4"},
		{"e5", "e7e5"},
		{"Nf3", "g1f3"},
		{"Nc6", "b8c6"},
		{"Bc4", "f1c4"},
	}

	for i, m := range moves {
		move, err := ParseSAN(board, m.san)
		if err != nil {
			t.Fatalf("move %d (%s) failed: %v", i+1, m.san, err)
		}

		if move.String() != m.wantMove {
			t.Errorf("move %d: expected %s, got %s", i+1, m.wantMove, move.String())
		}

		// Apply the move to the board
		err = board.MakeMove(move)
		if err != nil {
			t.Fatalf("failed to apply move %d (%s): %v", i+1, m.san, err)
		}
	}
}

// TestParseSAN_FileDisambiguation tests parsing of moves with file disambiguation.
func TestParseSAN_FileDisambiguation(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		san      string
		wantMove string
		wantErr  bool
	}{
		{
			name: "Nbd2 - knight from b-file",
			// Two knights can reach d2: one on b1, one on f3 (d2 is empty)
			fen:      "rnbqkb1r/pppppppp/5n2/8/8/5N2/PPP1PPPP/RNBQKB1R w KQkq - 0 1",
			san:      "Nbd2",
			wantMove: "b1d2",
			wantErr:  false,
		},
		{
			name: "Nfd2 - knight from f-file",
			// Two knights can reach d2: one on b1, one on f3 (d2 is empty)
			fen:      "rnbqkb1r/pppppppp/5n2/8/8/5N2/PPP1PPPP/RNBQKB1R w KQkq - 0 1",
			san:      "Nfd2",
			wantMove: "f3d2",
			wantErr:  false,
		},
		{
			name: "Rfe1 - rook from f-file",
			// Two rooks on a1 and f1 can reach e1 (king on g1)
			fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/R4RK1 w kq - 0 1",
			san:      "Rfe1",
			wantMove: "f1e1",
			wantErr:  false,
		},
		{
			name: "Rae5 - rook from a-file",
			// Rooks on a5 and h5 can both reach e5
			fen:      "4k3/8/8/R6R/8/8/8/4K3 w - - 0 1",
			san:      "Rae5",
			wantMove: "a5e5",
			wantErr:  false,
		},
		{
			name: "Qae5 - queen from a-file",
			// Queens on a1 and h8 can both reach e5
			fen:      "4k2Q/8/8/8/8/8/8/Q2K4 w - - 0 1",
			san:      "Qae5",
			wantMove: "a1e5",
			wantErr:  false,
		},
		{
			name: "Qhe5 - queen from h-file",
			// Queens on a1 and h8 can both reach e5
			fen:      "4k2Q/8/8/8/8/8/8/Q2K4 w - - 0 1",
			san:      "Qhe5",
			wantMove: "h8e5",
			wantErr:  false,
		},
		{
			name: "Rhe5 - rook from h-file",
			// Rooks on a5 and h5 can both reach e5
			fen:      "4k3/8/8/R6R/8/8/8/4K3 w - - 0 1",
			san:      "Rhe5",
			wantMove: "h5e5",
			wantErr:  false,
		},
		{
			name: "ambiguous without disambiguation",
			// Two knights can reach d2 but no disambiguation provided
			fen:      "rnbqkb1r/pppppppp/5n2/8/8/5N2/PPP1PPPP/RNBQKB1R w KQkq - 0 1",
			san:      "Nd2",
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

// TestParseSAN_RankDisambiguation tests parsing of moves with rank disambiguation.
func TestParseSAN_RankDisambiguation(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		san      string
		wantMove string
		wantErr  bool
	}{
		{
			name: "N1c3 - knight from rank 1",
			// Knights on b1 and d4 can both reach c3
			fen:      "rnbqkb1r/pppppppp/8/8/3N4/8/PPPPPPPP/RNBQKB1R w KQkq - 0 1",
			san:      "N1c3",
			wantMove: "b1c3",
			wantErr:  false,
		},
		{
			name: "N5c3 - knight from rank 5",
			// Knights on b1 and d5 can both reach c3 (c3 is empty)
			fen:      "rnbqkb1r/pppppppp/8/3N4/8/8/PP1PPPPP/RNBQKB1R w KQkq - 0 1",
			san:      "N5c3",
			wantMove: "d5c3",
			wantErr:  false,
		},
		{
			name: "R1a3 - rook from rank 1",
			// Rooks on a1 and a5 can both reach a3
			fen:      "4k3/8/8/r7/8/8/8/R3K3 w - - 0 1",
			san:      "R1a3",
			wantMove: "a1a3",
			wantErr:  false,
		},
		{
			name: "R5a3 - rook from rank 5",
			// Rooks on a1 and a5 can both reach a3
			fen:      "4k3/8/8/r7/8/8/8/R3K3 b - - 0 1",
			san:      "R5a3",
			wantMove: "a5a3",
			wantErr:  false,
		},
		{
			name: "Q1d4 - queen from rank 1",
			// Queens on d1 and d8 can both reach d4
			fen:      "3Qk3/8/8/8/8/8/8/3QK3 w - - 0 1",
			san:      "Q1d4",
			wantMove: "d1d4",
			wantErr:  false,
		},
		{
			name: "Q4h4 - queen from rank 4",
			// Queens on d1 and d4 can both reach h4
			fen:      "4k3/8/8/8/3Q4/8/8/3QK3 w - - 0 1",
			san:      "Q4h4",
			wantMove: "d4h4",
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

// TestParseSAN_FileAndRankDisambiguation tests parsing of moves with both file and rank disambiguation.
func TestParseSAN_FileAndRankDisambiguation(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		san      string
		wantMove string
		wantErr  bool
	}{
		{
			name: "Nb1d2 - knight from b1",
			// Multiple knights can reach d2: b1, f3 (d2 is empty)
			fen:      "rnbqkb1r/pppppppp/5n2/8/8/5N2/PPP1PPPP/RNBQKB1R w KQkq - 0 1",
			san:      "Nb1d2",
			wantMove: "b1d2",
			wantErr:  false,
		},
		{
			name: "Nf3d2 - knight from f3",
			// Multiple knights can reach d2: b1, f3 (d2 is empty)
			fen:      "rnbqkb1r/pppppppp/5n2/8/8/5N2/PPP1PPPP/RNBQKB1R w KQkq - 0 1",
			san:      "Nf3d2",
			wantMove: "f3d2",
			wantErr:  false,
		},
		{
			name: "Ra1a3 - rook from a1 to a3",
			// Rooks on a1, a8, h1 where a1 and h1 can reach a3
			fen:      "r3k3/8/8/8/8/8/8/R3K2R w KQ - 0 1",
			san:      "Ra1a3",
			wantMove: "a1a3",
			wantErr:  false,
		},
		{
			name: "Rh1h3 - rook from h1 to h3",
			// Rooks on a1, h1, h8 where h1 and h8 can reach h3
			fen:      "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
			san:      "Rh1h3",
			wantMove: "h1h3",
			wantErr:  false,
		},
		{
			name: "Qd1d4 - queen from d1 to d4",
			// Multiple queens on d1, h4
			fen:      "4k3/8/8/8/7Q/8/8/3QK3 w - - 0 1",
			san:      "Qd1d4",
			wantMove: "d1d4",
			wantErr:  false,
		},
		{
			name: "Qh4d4 - queen from h4 to d4",
			// Multiple queens on d1, h4
			fen:      "4k3/8/8/8/7Q/8/8/3QK3 w - - 0 1",
			san:      "Qh4d4",
			wantMove: "h4d4",
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

// TestParseSAN_DisambiguationWithCaptures tests disambiguation combined with captures.
func TestParseSAN_DisambiguationWithCaptures(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		san      string
		wantMove string
		wantErr  bool
	}{
		{
			name: "Nbxd4 - knight from b-file captures on d4",
			// Knights on b3 and f3 can both capture on d4
			fen:      "rnbqkb1r/ppp1pppp/8/3p4/8/1N3N2/PPPPPPPP/R1BQKB1R w KQkq - 0 1",
			san:      "Nbxd4",
			wantMove: "b3d4",
			wantErr:  false,
		},
		{
			name: "Nfxd4 - knight from f-file captures on d4",
			// Knights on b3 and f3 can both capture on d4
			fen:      "rnbqkb1r/ppp1pppp/8/3p4/8/1N3N2/PPPPPPPP/R1BQKB1R w KQkq - 0 1",
			san:      "Nfxd4",
			wantMove: "f3d4",
			wantErr:  false,
		},
		{
			name: "R1xe5 - rook from rank 1 captures on e5",
			// Rooks on e1 and e8 can both capture on e5
			fen:      "3kr3/8/8/4p3/8/8/8/3KR3 w - - 0 1",
			san:      "R1xe5",
			wantMove: "e1e5",
			wantErr:  false,
		},
		{
			name: "Raxe5 - rook from a-file captures on e5",
			// Rooks on a5 and h5 can both capture on e5 (black pawns on e5)
			fen:      "4k3/8/8/r3P2r/8/8/8/4K3 b - - 0 1",
			san:      "Raxe5",
			wantMove: "a5e5",
			wantErr:  false,
		},
		{
			name: "Qa1xd4 - full disambiguation with capture",
			// Queens on a1 and a8 can both capture on d4
			fen:      "Q3k3/8/8/8/3p4/8/8/Q3K3 w - - 0 1",
			san:      "Qa1xd4",
			wantMove: "a1d4",
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

// TestParseSAN_GameWithDisambiguation tests a game sequence that requires disambiguation.
func TestParseSAN_GameWithDisambiguation(t *testing.T) {
	board := engine.NewBoard()

	// Play a game that creates ambiguous positions
	moves := []struct {
		san      string
		wantMove string
	}{
		{"Nf3", "g1f3"},
		{"Nf6", "g8f6"},
		{"Nc3", "b1c3"},
		{"Nc6", "b8c6"},
		{"e4", "e2e4"},
		{"e5", "e7e5"},
		{"Bb5", "f1b5"},
		{"Bb4", "f8b4"},
		{"O-O", "e1g1"},
		{"O-O", "e8g8"},
		{"d3", "d2d3"},
		{"d6", "d7d6"},
		{"Bg5", "c1g5"},
		{"Bg4", "c8g4"},
		{"Nd5", "c3d5"}, // This requires disambiguation eventually
		{"Nxd5", "f6d5"},
		{"exd5", "e4d5"},
	}

	for i, m := range moves {
		move, err := ParseSAN(board, m.san)
		if err != nil {
			t.Fatalf("move %d (%s) failed: %v", i+1, m.san, err)
		}

		if move.String() != m.wantMove {
			t.Errorf("move %d (%s): expected %s, got %s", i+1, m.san, m.wantMove, move.String())
		}

		// Apply the move to the board
		err = board.MakeMove(move)
		if err != nil {
			t.Fatalf("failed to apply move %d (%s): %v", i+1, m.san, err)
		}
	}
}
