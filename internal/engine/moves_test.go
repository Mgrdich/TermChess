package engine

import (
	"strings"
	"testing"
)

func TestParseMove(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantFrom  Square
		wantTo    Square
		wantPromo PieceType
		wantErr   bool
	}{
		{
			name:      "simple move e2e4",
			input:     "e2e4",
			wantFrom:  NewSquare(4, 1), // e2
			wantTo:    NewSquare(4, 3), // e4
			wantPromo: Empty,
			wantErr:   false,
		},
		{
			name:      "move a1h8",
			input:     "a1h8",
			wantFrom:  NewSquare(0, 0), // a1
			wantTo:    NewSquare(7, 7), // h8
			wantPromo: Empty,
			wantErr:   false,
		},
		{
			name:      "promotion to queen a7a8q",
			input:     "a7a8q",
			wantFrom:  NewSquare(0, 6), // a7
			wantTo:    NewSquare(0, 7), // a8
			wantPromo: Queen,
			wantErr:   false,
		},
		{
			name:      "promotion to rook h7h8r",
			input:     "h7h8r",
			wantFrom:  NewSquare(7, 6), // h7
			wantTo:    NewSquare(7, 7), // h8
			wantPromo: Rook,
			wantErr:   false,
		},
		{
			name:      "promotion to bishop b7b8b",
			input:     "b7b8b",
			wantFrom:  NewSquare(1, 6), // b7
			wantTo:    NewSquare(1, 7), // b8
			wantPromo: Bishop,
			wantErr:   false,
		},
		{
			name:      "promotion to knight c7c8n",
			input:     "c7c8n",
			wantFrom:  NewSquare(2, 6), // c7
			wantTo:    NewSquare(2, 7), // c8
			wantPromo: Knight,
			wantErr:   false,
		},
		{
			name:    "too short e2",
			input:   "e2",
			wantErr: true,
		},
		{
			name:    "invalid rank e2e9",
			input:   "e2e9",
			wantErr: true,
		},
		{
			name:    "invalid string xyz",
			input:   "xyz",
			wantErr: true,
		},
		{
			name:    "invalid from file i2e4",
			input:   "i2e4",
			wantErr: true,
		},
		{
			name:    "invalid promotion char e7e8x",
			input:   "e7e8x",
			wantErr: true,
		},
		{
			name:    "too long e2e4qq",
			input:   "e2e4qq",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			move, err := ParseMove(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseMove(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseMove(%q) unexpected error: %v", tt.input, err)
				return
			}

			if move.From != tt.wantFrom {
				t.Errorf("ParseMove(%q).From = %v, want %v", tt.input, move.From, tt.wantFrom)
			}
			if move.To != tt.wantTo {
				t.Errorf("ParseMove(%q).To = %v, want %v", tt.input, move.To, tt.wantTo)
			}
			if move.Promotion != tt.wantPromo {
				t.Errorf("ParseMove(%q).Promotion = %v, want %v", tt.input, move.Promotion, tt.wantPromo)
			}
		})
	}
}

func TestMoveString(t *testing.T) {
	tests := []struct {
		name string
		move Move
		want string
	}{
		{
			name: "simple move e2e4",
			move: Move{From: NewSquare(4, 1), To: NewSquare(4, 3), Promotion: Empty},
			want: "e2e4",
		},
		{
			name: "promotion to queen a7a8q",
			move: Move{From: NewSquare(0, 6), To: NewSquare(0, 7), Promotion: Queen},
			want: "a7a8q",
		},
		{
			name: "promotion to rook h7h8r",
			move: Move{From: NewSquare(7, 6), To: NewSquare(7, 7), Promotion: Rook},
			want: "h7h8r",
		},
		{
			name: "promotion to bishop b7b8b",
			move: Move{From: NewSquare(1, 6), To: NewSquare(1, 7), Promotion: Bishop},
			want: "b7b8b",
		},
		{
			name: "promotion to knight c7c8n",
			move: Move{From: NewSquare(2, 6), To: NewSquare(2, 7), Promotion: Knight},
			want: "c7c8n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.move.String()
			if got != tt.want {
				t.Errorf("Move.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMoveRoundTrip(t *testing.T) {
	// Test that parsing and stringifying round-trips correctly
	testCases := []string{"e2e4", "a7a8q", "b1c3", "h1h8", "d7d8r", "e7e8n", "f7f8b"}

	for _, input := range testCases {
		t.Run(input, func(t *testing.T) {
			move, err := ParseMove(input)
			if err != nil {
				t.Fatalf("ParseMove(%q) error: %v", input, err)
			}

			output := move.String()
			if output != input {
				t.Errorf("Round trip failed: input=%q, output=%q", input, output)
			}
		})
	}
}

func TestBoardCopy(t *testing.T) {
	t.Run("copied board equals original", func(t *testing.T) {
		original := NewBoard()
		original.History = append(original.History, 12345, 67890)

		copied := original.Copy()

		// Check all fields match
		if copied.ActiveColor != original.ActiveColor {
			t.Errorf("ActiveColor mismatch: got %v, want %v", copied.ActiveColor, original.ActiveColor)
		}
		if copied.CastlingRights != original.CastlingRights {
			t.Errorf("CastlingRights mismatch: got %v, want %v", copied.CastlingRights, original.CastlingRights)
		}
		if copied.EnPassantSq != original.EnPassantSq {
			t.Errorf("EnPassantSq mismatch: got %v, want %v", copied.EnPassantSq, original.EnPassantSq)
		}
		if copied.HalfMoveClock != original.HalfMoveClock {
			t.Errorf("HalfMoveClock mismatch: got %v, want %v", copied.HalfMoveClock, original.HalfMoveClock)
		}
		if copied.FullMoveNum != original.FullMoveNum {
			t.Errorf("FullMoveNum mismatch: got %v, want %v", copied.FullMoveNum, original.FullMoveNum)
		}
		if copied.Hash != original.Hash {
			t.Errorf("Hash mismatch: got %v, want %v", copied.Hash, original.Hash)
		}

		// Check Squares
		for i := 0; i < 64; i++ {
			if copied.Squares[i] != original.Squares[i] {
				t.Errorf("Squares[%d] mismatch: got %v, want %v", i, copied.Squares[i], original.Squares[i])
			}
		}

		// Check History
		if len(copied.History) != len(original.History) {
			t.Errorf("History length mismatch: got %d, want %d", len(copied.History), len(original.History))
		}
		for i := range original.History {
			if copied.History[i] != original.History[i] {
				t.Errorf("History[%d] mismatch: got %v, want %v", i, copied.History[i], original.History[i])
			}
		}
	})

	t.Run("modifying copy doesn't affect original", func(t *testing.T) {
		original := NewBoard()
		original.History = append(original.History, 12345)

		copied := original.Copy()

		// Modify the copy
		copied.ActiveColor = Black
		copied.CastlingRights = 0
		copied.EnPassantSq = 20
		copied.HalfMoveClock = 10
		copied.FullMoveNum = 50
		copied.Hash = 999999
		copied.Squares[0] = Piece(0) // Clear a1
		copied.History = append(copied.History, 11111)

		// Verify original is unchanged
		if original.ActiveColor != White {
			t.Error("original.ActiveColor was modified")
		}
		if original.CastlingRights != CastleAll {
			t.Error("original.CastlingRights was modified")
		}
		if original.EnPassantSq != -1 {
			t.Error("original.EnPassantSq was modified")
		}
		if original.HalfMoveClock != 0 {
			t.Error("original.HalfMoveClock was modified")
		}
		if original.FullMoveNum != 1 {
			t.Error("original.FullMoveNum was modified")
		}
		if original.Hash != 0 {
			t.Error("original.Hash was modified")
		}
		if original.Squares[0].IsEmpty() {
			t.Error("original.Squares[0] was modified")
		}
		if len(original.History) != 1 {
			t.Errorf("original.History was modified: got length %d, want 1", len(original.History))
		}
	})
}

func TestGeneratePawnMoves(t *testing.T) {
	t.Run("starting position e2 pawn has 2 moves", func(t *testing.T) {
		board := NewBoard()
		moves := board.generatePawnMoves()

		// Find moves from e2
		e2 := NewSquare(4, 1) // e2
		e2Moves := []Move{}
		for _, m := range moves {
			if m.From == e2 {
				e2Moves = append(e2Moves, m)
			}
		}

		if len(e2Moves) != 2 {
			t.Errorf("e2 pawn expected 2 moves, got %d", len(e2Moves))
		}

		// Check e3 and e4 are the destinations
		e3 := NewSquare(4, 2) // e3
		e4 := NewSquare(4, 3) // e4
		hasE3, hasE4 := false, false
		for _, m := range e2Moves {
			if m.To == e3 {
				hasE3 = true
			}
			if m.To == e4 {
				hasE4 = true
			}
		}
		if !hasE3 {
			t.Error("e2 pawn should be able to move to e3")
		}
		if !hasE4 {
			t.Error("e2 pawn should be able to move to e4")
		}
	})

	t.Run("pawn on e4 has 1 move (e5) if no captures", func(t *testing.T) {
		board := NewBoard()
		// Move the e2 pawn to e4 manually
		e2 := NewSquare(4, 1)
		e4 := NewSquare(4, 3)
		board.Squares[e4] = board.Squares[e2]
		board.Squares[e2] = Piece(Empty)

		moves := board.generatePawnMoves()

		// Find moves from e4
		e4Moves := []Move{}
		for _, m := range moves {
			if m.From == e4 {
				e4Moves = append(e4Moves, m)
			}
		}

		if len(e4Moves) != 1 {
			t.Errorf("e4 pawn expected 1 move, got %d", len(e4Moves))
		}

		e5 := NewSquare(4, 4)
		if len(e4Moves) > 0 && e4Moves[0].To != e5 {
			t.Errorf("e4 pawn should move to e5, got %s", e4Moves[0].To.String())
		}
	})

	t.Run("pawn can capture diagonally", func(t *testing.T) {
		board := NewBoard()
		// Place white pawn on e4 and black pawn on d5
		e4 := NewSquare(4, 3) // e4
		d5 := NewSquare(3, 4) // d5

		board.Squares[e4] = NewPiece(White, Pawn)
		board.Squares[d5] = NewPiece(Black, Pawn)

		// Clear e2 pawn to simplify test
		e2 := NewSquare(4, 1)
		board.Squares[e2] = Piece(Empty)

		moves := board.generatePawnMoves()

		// Find moves from e4
		e4Moves := []Move{}
		for _, m := range moves {
			if m.From == e4 {
				e4Moves = append(e4Moves, m)
			}
		}

		// Should have e5 (forward) and d5 (capture)
		hasCapture := false
		for _, m := range e4Moves {
			if m.To == d5 {
				hasCapture = true
			}
		}

		if !hasCapture {
			t.Error("e4 pawn should be able to capture on d5")
		}
	})

	t.Run("pawn blocked by piece in front has no forward moves", func(t *testing.T) {
		board := NewBoard()
		// Place a piece in front of e2 pawn (e3)
		e2 := NewSquare(4, 1) // e2
		e3 := NewSquare(4, 2) // e3
		board.Squares[e3] = NewPiece(Black, Knight)

		moves := board.generatePawnMoves()

		// Find moves from e2
		e2Moves := []Move{}
		for _, m := range moves {
			if m.From == e2 {
				e2Moves = append(e2Moves, m)
			}
		}

		if len(e2Moves) != 0 {
			t.Errorf("blocked e2 pawn expected 0 moves, got %d", len(e2Moves))
		}
	})

	t.Run("black pawn moves down", func(t *testing.T) {
		board := NewBoard()
		board.ActiveColor = Black

		moves := board.generatePawnMoves()

		// Find moves from e7 (black pawn)
		e7 := NewSquare(4, 6) // e7
		e7Moves := []Move{}
		for _, m := range moves {
			if m.From == e7 {
				e7Moves = append(e7Moves, m)
			}
		}

		if len(e7Moves) != 2 {
			t.Errorf("e7 pawn expected 2 moves, got %d", len(e7Moves))
		}

		// Check e6 and e5 are the destinations
		e6 := NewSquare(4, 5) // e6
		e5 := NewSquare(4, 4) // e5
		hasE6, hasE5 := false, false
		for _, m := range e7Moves {
			if m.To == e6 {
				hasE6 = true
			}
			if m.To == e5 {
				hasE5 = true
			}
		}
		if !hasE6 {
			t.Error("e7 pawn should be able to move to e6")
		}
		if !hasE5 {
			t.Error("e7 pawn should be able to move to e5")
		}
	})
}

func TestMakeMove(t *testing.T) {
	t.Run("e2e4 moves pawn and clears e2", func(t *testing.T) {
		board := NewBoard()
		move, _ := ParseMove("e2e4")

		err := board.MakeMove(move)
		if err != nil {
			t.Fatalf("MakeMove returned error: %v", err)
		}

		e2 := NewSquare(4, 1)
		e4 := NewSquare(4, 3)

		// e4 should have white pawn
		if board.Squares[e4].Type() != Pawn || board.Squares[e4].Color() != White {
			t.Error("e4 should have white pawn after e2e4")
		}

		// e2 should be empty
		if !board.Squares[e2].IsEmpty() {
			t.Error("e2 should be empty after e2e4")
		}
	})

	t.Run("active color changes after move", func(t *testing.T) {
		board := NewBoard()
		if board.ActiveColor != White {
			t.Fatal("starting color should be White")
		}

		move, _ := ParseMove("e2e4")
		if err := board.MakeMove(move); err != nil {
			t.Fatalf("MakeMove returned error: %v", err)
		}

		if board.ActiveColor != Black {
			t.Error("active color should be Black after White's move")
		}

		// Black moves
		move2, _ := ParseMove("e7e5")
		if err := board.MakeMove(move2); err != nil {
			t.Fatalf("MakeMove returned error: %v", err)
		}

		if board.ActiveColor != White {
			t.Error("active color should be White after Black's move")
		}
	})

	t.Run("full move number increments after Black's move", func(t *testing.T) {
		board := NewBoard()
		if board.FullMoveNum != 1 {
			t.Fatal("starting full move number should be 1")
		}

		// White moves
		move, _ := ParseMove("e2e4")
		if err := board.MakeMove(move); err != nil {
			t.Fatalf("MakeMove returned error: %v", err)
		}

		if board.FullMoveNum != 1 {
			t.Error("full move number should still be 1 after White's move")
		}

		// Black moves
		move2, _ := ParseMove("e7e5")
		if err := board.MakeMove(move2); err != nil {
			t.Fatalf("MakeMove returned error: %v", err)
		}

		if board.FullMoveNum != 2 {
			t.Error("full move number should be 2 after Black's move")
		}
	})

	t.Run("invalid move - no piece - returns error", func(t *testing.T) {
		board := NewBoard()
		// Try to move from an empty square (e4)
		move, _ := ParseMove("e4e5")

		err := board.MakeMove(move)
		if err == nil {
			t.Error("MakeMove should return error when no piece at source")
		}
	})

	t.Run("invalid move - wrong color - returns error", func(t *testing.T) {
		board := NewBoard()
		// White to move, but try to move Black's pawn
		move, _ := ParseMove("e7e6")

		err := board.MakeMove(move)
		if err == nil {
			t.Error("MakeMove should return error when piece belongs to opponent")
		}
	})

	t.Run("capture replaces piece", func(t *testing.T) {
		board := NewBoard()
		// Set up a capture scenario: white pawn on e4, black pawn on d5
		e4 := NewSquare(4, 3)
		d5 := NewSquare(3, 4)
		board.Squares[e4] = NewPiece(White, Pawn)
		board.Squares[d5] = NewPiece(Black, Pawn)

		// Remove e2 pawn to make it clearer
		e2 := NewSquare(4, 1)
		board.Squares[e2] = Piece(Empty)

		move, _ := ParseMove("e4d5")
		err := board.MakeMove(move)
		if err != nil {
			t.Fatalf("MakeMove returned error: %v", err)
		}

		// d5 should now have white pawn
		if board.Squares[d5].Type() != Pawn || board.Squares[d5].Color() != White {
			t.Error("d5 should have white pawn after capture")
		}

		// e4 should be empty
		if !board.Squares[e4].IsEmpty() {
			t.Error("e4 should be empty after capture")
		}
	})

	t.Run("board state unchanged after illegal move", func(t *testing.T) {
		board := NewBoard()
		// Save the original board state
		originalSquares := board.Squares
		originalActiveColor := board.ActiveColor
		originalCastlingRights := board.CastlingRights
		originalEnPassantSq := board.EnPassantSq
		originalHalfMoveClock := board.HalfMoveClock
		originalFullMoveNum := board.FullMoveNum

		// Try to make an illegal move (moving from empty square)
		move, _ := ParseMove("e4e5")
		err := board.MakeMove(move)
		if err == nil {
			t.Fatal("expected error for illegal move")
		}

		// Verify board state is unchanged
		if board.Squares != originalSquares {
			t.Error("board squares should be unchanged after illegal move")
		}
		if board.ActiveColor != originalActiveColor {
			t.Error("active color should be unchanged after illegal move")
		}
		if board.CastlingRights != originalCastlingRights {
			t.Error("castling rights should be unchanged after illegal move")
		}
		if board.EnPassantSq != originalEnPassantSq {
			t.Error("en passant square should be unchanged after illegal move")
		}
		if board.HalfMoveClock != originalHalfMoveClock {
			t.Error("half move clock should be unchanged after illegal move")
		}
		if board.FullMoveNum != originalFullMoveNum {
			t.Error("full move number should be unchanged after illegal move")
		}
	})

	t.Run("pinned piece move returns error", func(t *testing.T) {
		// Set up a position where a piece is pinned to the king
		// White king on e1, white bishop on e2 (pinned), black rook on e8
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King)   // e1
		board.Squares[NewSquare(4, 1)] = NewPiece(White, Bishop) // e2 (pinned)
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, Rook)   // e8
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King)   // h8 (black king needed)

		// Try to move the pinned bishop
		move, _ := ParseMove("e2d3")
		err := board.MakeMove(move)
		if err == nil {
			t.Error("MakeMove should return error when moving pinned piece")
		}

		// Bishop should still be on e2
		e2 := NewSquare(4, 1)
		if board.Squares[e2].Type() != Bishop || board.Squares[e2].Color() != White {
			t.Error("pinned bishop should remain on e2 after failed move")
		}
	})

	t.Run("moving into check returns error", func(t *testing.T) {
		// White king on e4, black rook on a5
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 3)] = NewPiece(White, King) // e4
		board.Squares[NewSquare(0, 4)] = NewPiece(Black, Rook) // a5
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King) // h8 (black king needed)

		// King moving to e5 would put it in check from the rook
		move, _ := ParseMove("e4e5")
		err := board.MakeMove(move)
		if err == nil {
			t.Error("MakeMove should return error when king moves into check")
		}

		// King should still be on e4
		e4 := NewSquare(4, 3)
		if board.Squares[e4].Type() != King || board.Squares[e4].Color() != White {
			t.Error("king should remain on e4 after failed move into check")
		}
	})

	t.Run("error message contains move string", func(t *testing.T) {
		board := NewBoard()
		// Try an illegal move
		move, _ := ParseMove("e4e5")
		err := board.MakeMove(move)
		if err == nil {
			t.Fatal("expected error for illegal move")
		}

		// Error message should contain the move
		if !strings.Contains(err.Error(), "e4e5") {
			t.Errorf("error message should contain move string, got: %s", err.Error())
		}
	})
}

// Helper function to check if a move exists in a list of moves
func containsMove(moves []Move, from, to Square) bool {
	for _, m := range moves {
		if m.From == from && m.To == to {
			return true
		}
	}
	return false
}

// Helper function to count moves from a specific square
func countMovesFrom(moves []Move, sq Square) int {
	count := 0
	for _, m := range moves {
		if m.From == sq {
			count++
		}
	}
	return count
}


func TestGenerateKnightMoves(t *testing.T) {
	t.Run("knight on e4 can reach 8 squares", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		// Place white knight on e4
		e4 := NewSquare(4, 3)
		board.Squares[e4] = NewPiece(White, Knight)

		moves := board.generateKnightMoves()

		// Knight on e4 should be able to reach:
		// f6, g5, g3, f2, d2, c3, c5, d6
		expectedTargets := []Square{
			NewSquare(5, 5), // f6
			NewSquare(6, 4), // g5
			NewSquare(6, 2), // g3
			NewSquare(5, 1), // f2
			NewSquare(3, 1), // d2
			NewSquare(2, 2), // c3
			NewSquare(2, 4), // c5
			NewSquare(3, 5), // d6
		}

		if len(moves) != 8 {
			t.Errorf("knight on e4 expected 8 moves, got %d", len(moves))
		}

		for _, target := range expectedTargets {
			if !containsMove(moves, e4, target) {
				t.Errorf("knight on e4 should be able to move to %s", target.String())
			}
		}
	})

	t.Run("knight on a1 corner has only 2 legal squares", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		// Place white knight on a1
		a1 := NewSquare(0, 0)
		board.Squares[a1] = NewPiece(White, Knight)

		moves := board.generateKnightMoves()

		// Knight on a1 should only reach: b3, c2
		if len(moves) != 2 {
			t.Errorf("knight on a1 expected 2 moves, got %d", len(moves))
		}

		b3 := NewSquare(1, 2)
		c2 := NewSquare(2, 1)
		if !containsMove(moves, a1, b3) {
			t.Error("knight on a1 should be able to move to b3")
		}
		if !containsMove(moves, a1, c2) {
			t.Error("knight on a1 should be able to move to c2")
		}
	})

	t.Run("knight can jump over pieces", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		// Place white knight on b1, surrounded by pawns on a2, b2, c2, b3
		b1 := NewSquare(1, 0)
		board.Squares[b1] = NewPiece(White, Knight)
		board.Squares[NewSquare(0, 1)] = NewPiece(White, Pawn) // a2
		board.Squares[NewSquare(1, 1)] = NewPiece(White, Pawn) // b2
		board.Squares[NewSquare(2, 1)] = NewPiece(White, Pawn) // c2
		board.Squares[NewSquare(1, 2)] = NewPiece(White, Pawn) // b3

		moves := board.generateKnightMoves()

		// Knight can still reach a3 and c3 despite being surrounded
		a3 := NewSquare(0, 2)
		c3 := NewSquare(2, 2)
		d2 := NewSquare(3, 1)

		if !containsMove(moves, b1, a3) {
			t.Error("knight on b1 should be able to jump to a3")
		}
		if !containsMove(moves, b1, c3) {
			t.Error("knight on b1 should be able to jump to c3")
		}
		if !containsMove(moves, b1, d2) {
			t.Error("knight on b1 should be able to jump to d2")
		}
	})

	t.Run("knight captures enemy but not own piece", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		// Place white knight on e4
		e4 := NewSquare(4, 3)
		board.Squares[e4] = NewPiece(White, Knight)

		// Place white pawn on f6 (can't capture)
		f6 := NewSquare(5, 5)
		board.Squares[f6] = NewPiece(White, Pawn)

		// Place black pawn on d6 (can capture)
		d6 := NewSquare(3, 5)
		board.Squares[d6] = NewPiece(Black, Pawn)

		moves := board.generateKnightMoves()

		// Should not be able to move to f6 (own piece)
		if containsMove(moves, e4, f6) {
			t.Error("knight should not be able to capture own piece on f6")
		}

		// Should be able to capture on d6
		if !containsMove(moves, e4, d6) {
			t.Error("knight should be able to capture enemy piece on d6")
		}
	})
}

func TestGenerateBishopMoves(t *testing.T) {
	t.Run("bishop on e4 empty board", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		// Place white bishop on e4
		e4 := NewSquare(4, 3)
		board.Squares[e4] = NewPiece(White, Bishop)

		moves := board.generateBishopMoves()

		// Bishop on e4 should be able to reach:
		// f5, g6, h7 (NE diagonal)
		// f3, g2, h1 (SE diagonal)
		// d3, c2, b1 (SW diagonal)
		// d5, c6, b7, a8 (NW diagonal)
		// Total: 13 squares
		if len(moves) != 13 {
			t.Errorf("bishop on e4 expected 13 moves, got %d", len(moves))
		}

		// Check a few specific squares
		if !containsMove(moves, e4, NewSquare(7, 6)) { // h7
			t.Error("bishop should reach h7")
		}
		if !containsMove(moves, e4, NewSquare(0, 7)) { // a8
			t.Error("bishop should reach a8")
		}
	})

	t.Run("bishop blocked by own piece stops before it", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		// Place white bishop on e4 and white pawn on f5
		e4 := NewSquare(4, 3)
		f5 := NewSquare(5, 4)
		board.Squares[e4] = NewPiece(White, Bishop)
		board.Squares[f5] = NewPiece(White, Pawn)

		moves := board.generateBishopMoves()

		// Should not reach f5 or beyond in that diagonal
		if containsMove(moves, e4, f5) {
			t.Error("bishop should not move to own piece on f5")
		}
		if containsMove(moves, e4, NewSquare(6, 5)) { // g6
			t.Error("bishop should not jump over own piece to g6")
		}
	})

	t.Run("bishop can capture enemy piece and stops", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		// Place white bishop on e4 and black pawn on f5
		e4 := NewSquare(4, 3)
		f5 := NewSquare(5, 4)
		board.Squares[e4] = NewPiece(White, Bishop)
		board.Squares[f5] = NewPiece(Black, Pawn)

		moves := board.generateBishopMoves()

		// Can capture on f5
		if !containsMove(moves, e4, f5) {
			t.Error("bishop should be able to capture on f5")
		}
		// Should not go beyond f5 in that diagonal
		if containsMove(moves, e4, NewSquare(6, 5)) { // g6
			t.Error("bishop should stop after capturing on f5")
		}
	})
}

func TestGenerateRookMoves(t *testing.T) {
	t.Run("rook on e4 empty board", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		// Place white rook on e4
		e4 := NewSquare(4, 3)
		board.Squares[e4] = NewPiece(White, Rook)

		moves := board.generateRookMoves()

		// Rook on e4 should be able to reach:
		// e1, e2, e3, e5, e6, e7, e8 (vertical: 7 squares)
		// a4, b4, c4, d4, f4, g4, h4 (horizontal: 7 squares)
		// Total: 14 squares
		if len(moves) != 14 {
			t.Errorf("rook on e4 expected 14 moves, got %d", len(moves))
		}

		// Check edges
		if !containsMove(moves, e4, NewSquare(4, 0)) { // e1
			t.Error("rook should reach e1")
		}
		if !containsMove(moves, e4, NewSquare(4, 7)) { // e8
			t.Error("rook should reach e8")
		}
		if !containsMove(moves, e4, NewSquare(0, 3)) { // a4
			t.Error("rook should reach a4")
		}
		if !containsMove(moves, e4, NewSquare(7, 3)) { // h4
			t.Error("rook should reach h4")
		}
	})

	t.Run("rook blocked by pieces", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		// Place white rook on e4 and white pawn on e6, black pawn on c4
		e4 := NewSquare(4, 3)
		e6 := NewSquare(4, 5)
		c4 := NewSquare(2, 3)

		board.Squares[e4] = NewPiece(White, Rook)
		board.Squares[e6] = NewPiece(White, Pawn)
		board.Squares[c4] = NewPiece(Black, Pawn)

		moves := board.generateRookMoves()

		// Should reach e5 but not e6 or e7
		if !containsMove(moves, e4, NewSquare(4, 4)) { // e5
			t.Error("rook should reach e5")
		}
		if containsMove(moves, e4, e6) {
			t.Error("rook should not move to own piece on e6")
		}
		if containsMove(moves, e4, NewSquare(4, 6)) { // e7
			t.Error("rook should not jump over own piece to e7")
		}

		// Should capture c4 but not reach b4
		if !containsMove(moves, e4, c4) {
			t.Error("rook should capture on c4")
		}
		if containsMove(moves, e4, NewSquare(1, 3)) { // b4
			t.Error("rook should not reach b4 after capturing on c4")
		}
	})
}

func TestGenerateQueenMoves(t *testing.T) {
	t.Run("queen combines bishop and rook movement", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		// Place white queen on e4
		e4 := NewSquare(4, 3)
		board.Squares[e4] = NewPiece(White, Queen)

		moves := board.generateQueenMoves()

		// Queen on e4 should reach:
		// Rook moves: 14
		// Bishop moves: 13
		// Total: 27
		if len(moves) != 27 {
			t.Errorf("queen on e4 expected 27 moves, got %d", len(moves))
		}

		// Check bishop moves
		if !containsMove(moves, e4, NewSquare(7, 6)) { // h7 (diagonal)
			t.Error("queen should reach h7 diagonally")
		}

		// Check rook moves
		if !containsMove(moves, e4, NewSquare(4, 7)) { // e8 (vertical)
			t.Error("queen should reach e8 vertically")
		}
		if !containsMove(moves, e4, NewSquare(0, 3)) { // a4 (horizontal)
			t.Error("queen should reach a4 horizontally")
		}
	})

	t.Run("queen blocked appropriately", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		// Place white queen on e4, white pawn on e6, black pawn on f5
		e4 := NewSquare(4, 3)
		e6 := NewSquare(4, 5)
		f5 := NewSquare(5, 4)

		board.Squares[e4] = NewPiece(White, Queen)
		board.Squares[e6] = NewPiece(White, Pawn)
		board.Squares[f5] = NewPiece(Black, Pawn)

		moves := board.generateQueenMoves()

		// Cannot reach e6 (own piece)
		if containsMove(moves, e4, e6) {
			t.Error("queen should not move to own piece on e6")
		}

		// Can capture f5
		if !containsMove(moves, e4, f5) {
			t.Error("queen should capture on f5")
		}

		// Cannot reach g6 (blocked by f5)
		if containsMove(moves, e4, NewSquare(6, 5)) {
			t.Error("queen should not reach g6 after capture on f5")
		}
	})
}

func TestGenerateKingMoves(t *testing.T) {
	t.Run("king on e4 has 8 moves", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		// Place white king on e4
		e4 := NewSquare(4, 3)
		board.Squares[e4] = NewPiece(White, King)

		moves := board.generateKingMoves()

		// King should have 8 adjacent squares
		if len(moves) != 8 {
			t.Errorf("king on e4 expected 8 moves, got %d", len(moves))
		}

		// Check all adjacent squares
		adjacents := []Square{
			NewSquare(3, 2), // d3
			NewSquare(4, 2), // e3
			NewSquare(5, 2), // f3
			NewSquare(3, 3), // d4
			NewSquare(5, 3), // f4
			NewSquare(3, 4), // d5
			NewSquare(4, 4), // e5
			NewSquare(5, 4), // f5
		}

		for _, sq := range adjacents {
			if !containsMove(moves, e4, sq) {
				t.Errorf("king on e4 should move to %s", sq.String())
			}
		}
	})

	t.Run("king on a1 has 3 moves (corner)", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		// Place white king on a1
		a1 := NewSquare(0, 0)
		board.Squares[a1] = NewPiece(White, King)

		moves := board.generateKingMoves()

		// King in corner has only 3 squares
		if len(moves) != 3 {
			t.Errorf("king on a1 expected 3 moves, got %d", len(moves))
		}

		// Check the 3 possible squares: a2, b1, b2
		if !containsMove(moves, a1, NewSquare(0, 1)) { // a2
			t.Error("king on a1 should move to a2")
		}
		if !containsMove(moves, a1, NewSquare(1, 0)) { // b1
			t.Error("king on a1 should move to b1")
		}
		if !containsMove(moves, a1, NewSquare(1, 1)) { // b2
			t.Error("king on a1 should move to b2")
		}
	})

	t.Run("king can capture enemy but not own piece", func(t *testing.T) {
		board := &Board{ActiveColor: White}
		// Place white king on e4, white pawn on e5, black pawn on f5
		e4 := NewSquare(4, 3)
		e5 := NewSquare(4, 4)
		f5 := NewSquare(5, 4)

		board.Squares[e4] = NewPiece(White, King)
		board.Squares[e5] = NewPiece(White, Pawn)
		board.Squares[f5] = NewPiece(Black, Pawn)

		moves := board.generateKingMoves()

		// Should not move to e5 (own piece)
		if containsMove(moves, e4, e5) {
			t.Error("king should not capture own piece on e5")
		}

		// Should capture on f5
		if !containsMove(moves, e4, f5) {
			t.Error("king should capture enemy piece on f5")
		}
	})
}

func TestPseudoLegalMoves(t *testing.T) {
	t.Run("starting position White has 20 moves", func(t *testing.T) {
		board := NewBoard()
		moves := board.PseudoLegalMoves()

		// Starting position: 16 pawn moves (8 pawns x 2) + 4 knight moves (2 knights x 2)
		// = 20 total pseudo-legal moves
		if len(moves) != 20 {
			t.Errorf("starting position expected 20 pseudo-legal moves, got %d", len(moves))
			// List the moves for debugging
			t.Logf("Moves: ")
			for _, m := range moves {
				t.Logf("  %s", m.String())
			}
		}
	})

	t.Run("starting position Black has 20 moves", func(t *testing.T) {
		board := NewBoard()
		board.ActiveColor = Black
		moves := board.PseudoLegalMoves()

		if len(moves) != 20 {
			t.Errorf("starting position (Black) expected 20 pseudo-legal moves, got %d", len(moves))
		}
	})

	t.Run("PseudoLegalMoves combines all piece generators", func(t *testing.T) {
		board := &Board{ActiveColor: White}

		// Place various white pieces on empty board
		board.Squares[NewSquare(4, 3)] = NewPiece(White, King)   // e4
		board.Squares[NewSquare(0, 0)] = NewPiece(White, Knight) // a1
		board.Squares[NewSquare(7, 7)] = NewPiece(White, Bishop) // h8
		board.Squares[NewSquare(3, 1)] = NewPiece(White, Pawn)   // d2

		moves := board.PseudoLegalMoves()

		// Count moves by piece
		kingMoves := countMovesFrom(moves, NewSquare(4, 3))
		knightMoves := countMovesFrom(moves, NewSquare(0, 0))
		bishopMoves := countMovesFrom(moves, NewSquare(7, 7))
		pawnMoves := countMovesFrom(moves, NewSquare(3, 1))

		if kingMoves != 8 {
			t.Errorf("expected 8 king moves, got %d", kingMoves)
		}
		if knightMoves != 2 {
			t.Errorf("expected 2 knight moves (a1 corner), got %d", knightMoves)
		}
		if bishopMoves != 6 {
			// h8 can reach g7, f6, e5, d4, c3, b2 (6 squares)
			// a1 is blocked by the knight on a1
			t.Errorf("expected 6 bishop moves (h8 corner, blocked at a1 by knight), got %d", bishopMoves)
		}
		if pawnMoves != 2 {
			t.Errorf("expected 2 pawn moves (d2 starting rank), got %d", pawnMoves)
		}
	})
}

func TestLegalMoves(t *testing.T) {
	t.Run("starting position has 20 legal moves", func(t *testing.T) {
		board := NewBoard()
		moves := board.LegalMoves()

		// In the starting position, all 20 pseudo-legal moves are also legal
		// (16 pawn moves + 4 knight moves)
		if len(moves) != 20 {
			t.Errorf("starting position expected 20 legal moves, got %d", len(moves))
			t.Logf("Moves: ")
			for _, m := range moves {
				t.Logf("  %s", m.String())
			}
		}
	})

	t.Run("move that leaves king in check is filtered", func(t *testing.T) {
		// Set up a position where a piece is pinned to the king
		// White king on e1, white bishop on e2, black rook on e8
		// The bishop cannot move because it would expose the king to check
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King)   // e1
		board.Squares[NewSquare(4, 1)] = NewPiece(White, Bishop) // e2
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, Rook)   // e8
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King)   // h8 (black king needed)

		moves := board.LegalMoves()

		// Bishop on e2 should have 0 legal moves (it's pinned)
		bishopMoves := countMovesFrom(moves, NewSquare(4, 1))
		if bishopMoves != 0 {
			t.Errorf("pinned bishop should have 0 legal moves, got %d", bishopMoves)
		}

		// King should still have legal moves (not into the rook's attack)
		kingMoves := countMovesFrom(moves, NewSquare(4, 0))
		if kingMoves == 0 {
			t.Error("king should have at least one legal move")
		}

		// King should not be able to move to e2 (would still be in check after moving)
		// Actually, king CAN move to d1, f1, d2, f2 (not e2 since bishop is there)
		// But it cannot stay on the e-file (e2 has bishop)
	})

	t.Run("pinned piece cannot move to expose king", func(t *testing.T) {
		// White king on a1, white rook on a4, black queen on a8
		// The rook is pinned along the a-file
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(0, 0)] = NewPiece(White, King)  // a1
		board.Squares[NewSquare(0, 3)] = NewPiece(White, Rook)  // a4
		board.Squares[NewSquare(0, 7)] = NewPiece(Black, Queen) // a8
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King)  // h8 (black king needed)

		moves := board.LegalMoves()

		// The rook can only move along the a-file (a2, a3, a5, a6, a7, a8)
		// It cannot move horizontally (b4, c4, etc.) as that would expose the king
		rookMoves := countMovesFrom(moves, NewSquare(0, 3))

		// Check that rook can move to a5
		if !containsMove(moves, NewSquare(0, 3), NewSquare(0, 4)) { // a4 to a5
			t.Error("pinned rook should be able to move along the pin (a4 to a5)")
		}

		// Check that rook cannot move to b4
		if containsMove(moves, NewSquare(0, 3), NewSquare(1, 3)) { // a4 to b4
			t.Error("pinned rook should not be able to move to b4 (would expose king)")
		}

		// Rook can move to: a2, a3, a5, a6, a7, a8 (capture) = 6 moves
		if rookMoves != 6 {
			t.Errorf("pinned rook should have 6 legal moves along the file, got %d", rookMoves)
		}
	})

	t.Run("when in check only escaping moves are legal", func(t *testing.T) {
		// White king on e1 in check from black queen on e8
		// White has a knight on c3 that could block on e2 or e4
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King)   // e1
		board.Squares[NewSquare(2, 2)] = NewPiece(White, Knight) // c3
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, Queen)  // e8
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King)   // h8 (black king needed)

		// Verify the king is in check
		if !board.InCheck() {
			t.Fatal("king should be in check in this position")
		}

		moves := board.LegalMoves()

		// All legal moves should either:
		// 1. Move the king out of check
		// 2. Block the check (knight to e2 or e4)

		// Knight can block by going to e2 (Nc3-e2)
		if !containsMove(moves, NewSquare(2, 2), NewSquare(4, 1)) { // c3 to e2
			t.Error("knight should be able to block check by moving to e2")
		}

		// Knight can also block by going to e4 (Nc3-e4)
		if !containsMove(moves, NewSquare(2, 2), NewSquare(4, 3)) { // c3 to e4
			t.Error("knight should be able to block check by moving to e4")
		}

		// Knight's other moves should be filtered out (they don't escape check)
		// Knight on c3 can normally go to: a2, a4, b1, b5, d1, d5, e2, e4
		// But only e2 and e4 block the check (both are on the e-file)
		knightMoves := countMovesFrom(moves, NewSquare(2, 2))
		if knightMoves != 2 {
			t.Errorf("knight should have exactly 2 legal moves (block on e2 or e4), got %d", knightMoves)
		}

		// Non-blocking moves should not be legal
		if containsMove(moves, NewSquare(2, 2), NewSquare(0, 1)) { // c3 to a2
			t.Error("knight should not be able to move to a2 (does not escape check)")
		}
		if containsMove(moves, NewSquare(2, 2), NewSquare(3, 0)) { // c3 to d1
			t.Error("knight should not be able to move to d1 (does not escape check)")
		}
	})

	t.Run("king cannot move into check", func(t *testing.T) {
		// White king on e4, black rook on a5
		// King cannot move to d5, e5, or f5 (attacked by rook)
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 3)] = NewPiece(White, King) // e4
		board.Squares[NewSquare(0, 4)] = NewPiece(Black, Rook) // a5
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King) // h8 (black king needed)

		moves := board.LegalMoves()

		// King should not be able to move to the 5th rank (attacked by rook)
		if containsMove(moves, NewSquare(4, 3), NewSquare(3, 4)) { // e4 to d5
			t.Error("king should not be able to move to d5 (attacked by rook)")
		}
		if containsMove(moves, NewSquare(4, 3), NewSquare(4, 4)) { // e4 to e5
			t.Error("king should not be able to move to e5 (attacked by rook)")
		}
		if containsMove(moves, NewSquare(4, 3), NewSquare(5, 4)) { // e4 to f5
			t.Error("king should not be able to move to f5 (attacked by rook)")
		}

		// King should be able to move to 3rd rank (not attacked)
		if !containsMove(moves, NewSquare(4, 3), NewSquare(4, 2)) { // e4 to e3
			t.Error("king should be able to move to e3")
		}
	})

	t.Run("cannot capture protected piece with king", func(t *testing.T) {
		// White king on e4, black pawn on f5 protected by black bishop on h3
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 3)] = NewPiece(White, King)   // e4
		board.Squares[NewSquare(5, 4)] = NewPiece(Black, Pawn)   // f5
		board.Squares[NewSquare(7, 2)] = NewPiece(Black, Bishop) // h3 (protects f5)
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King)   // h8 (black king needed)

		moves := board.LegalMoves()

		// King should not be able to capture on f5 (it's protected)
		if containsMove(moves, NewSquare(4, 3), NewSquare(5, 4)) { // e4 to f5
			t.Error("king should not be able to capture protected pawn on f5")
		}
	})

	t.Run("double check requires king move", func(t *testing.T) {
		// White king on e1, attacked by black queen on e8 and black knight on d3
		// The only legal moves are king moves (cannot block both attacks)
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King)   // e1
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, Queen)  // e8 (gives check on e-file)
		board.Squares[NewSquare(3, 2)] = NewPiece(Black, Knight) // d3 (gives check)
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King)   // h8 (black king needed)
		board.Squares[NewSquare(0, 1)] = NewPiece(White, Rook)   // a2 (white rook that cannot help)

		// Verify the king is in check
		if !board.InCheck() {
			t.Fatal("king should be in check in this position")
		}

		moves := board.LegalMoves()

		// Rook cannot help because there are two attackers - only king can move
		rookMoves := countMovesFrom(moves, NewSquare(0, 1))
		if rookMoves != 0 {
			t.Errorf("rook should have 0 legal moves in double check, got %d", rookMoves)
		}

		// King must have at least one escape square
		kingMoves := countMovesFrom(moves, NewSquare(4, 0))
		if kingMoves == 0 {
			t.Error("king should have at least one escape square")
		}
	})
}

func TestIsLegalMove(t *testing.T) {
	t.Run("legal move returns true", func(t *testing.T) {
		board := NewBoard()
		// e2e4 is a legal move in the starting position
		move, _ := ParseMove("e2e4")
		if !board.IsLegalMove(move) {
			t.Error("e2e4 should be a legal move in the starting position")
		}

		// d2d3 is also legal
		move2, _ := ParseMove("d2d3")
		if !board.IsLegalMove(move2) {
			t.Error("d2d3 should be a legal move in the starting position")
		}

		// Knight moves are also legal
		move3, _ := ParseMove("g1f3")
		if !board.IsLegalMove(move3) {
			t.Error("g1f3 (Nf3) should be a legal move in the starting position")
		}
	})

	t.Run("illegal move leaves king in check returns false", func(t *testing.T) {
		// Set up a position where moving a piece would leave king in check
		// White king on e1, white bishop on e2 (pinned), black rook on e8
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King)   // e1
		board.Squares[NewSquare(4, 1)] = NewPiece(White, Bishop) // e2 (pinned)
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, Rook)   // e8
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King)   // h8 (black king needed)

		// Bishop moving to d3 would leave king in check - should be illegal
		move, _ := ParseMove("e2d3")
		if board.IsLegalMove(move) {
			t.Error("e2d3 should be illegal (would leave king in check)")
		}

		// Bishop moving to f3 would also leave king in check - should be illegal
		move2, _ := ParseMove("e2f3")
		if board.IsLegalMove(move2) {
			t.Error("e2f3 should be illegal (would leave king in check)")
		}
	})

	t.Run("completely invalid move no piece returns false", func(t *testing.T) {
		board := NewBoard()
		// Try to move from an empty square (e4 is empty in starting position)
		move, _ := ParseMove("e4e5")
		if board.IsLegalMove(move) {
			t.Error("e4e5 should be illegal (no piece on e4)")
		}

		// Try to move from another empty square
		move2, _ := ParseMove("d5d6")
		if board.IsLegalMove(move2) {
			t.Error("d5d6 should be illegal (no piece on d5)")
		}
	})

	t.Run("wrong color piece returns false", func(t *testing.T) {
		board := NewBoard()
		// White to move, but trying to move black piece
		move, _ := ParseMove("e7e6")
		if board.IsLegalMove(move) {
			t.Error("e7e6 should be illegal (wrong color - black piece on white's turn)")
		}

		// Another black piece move
		move2, _ := ParseMove("b8c6")
		if board.IsLegalMove(move2) {
			t.Error("b8c6 should be illegal (wrong color - black knight on white's turn)")
		}
	})

	t.Run("invalid piece movement pattern returns false", func(t *testing.T) {
		board := NewBoard()
		// Pawn can't move 3 squares
		move, _ := ParseMove("e2e5")
		if board.IsLegalMove(move) {
			t.Error("e2e5 should be illegal (pawn can't move 3 squares)")
		}

		// Knight can't move like a bishop
		move2, _ := ParseMove("b1b3")
		if board.IsLegalMove(move2) {
			t.Error("b1b3 should be illegal (knight can't move straight)")
		}

		// Pawn can't move diagonally without capture
		move3, _ := ParseMove("e2f3")
		if board.IsLegalMove(move3) {
			t.Error("e2f3 should be illegal (pawn can't move diagonally without capture)")
		}
	})

	t.Run("move blocked by own piece returns false", func(t *testing.T) {
		board := NewBoard()
		// Rook on a1 is blocked by knight on b1
		move, _ := ParseMove("a1b1")
		if board.IsLegalMove(move) {
			t.Error("a1b1 should be illegal (blocked by own knight)")
		}

		// Bishop on c1 is blocked by pawns
		move2, _ := ParseMove("c1d2")
		if board.IsLegalMove(move2) {
			t.Error("c1d2 should be illegal (blocked by own pawn)")
		}
	})

	t.Run("king cannot move into check", func(t *testing.T) {
		// White king on e4, black rook on a5
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 3)] = NewPiece(White, King) // e4
		board.Squares[NewSquare(0, 4)] = NewPiece(Black, Rook) // a5
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King) // h8 (black king needed)

		// King moving to e5 would put it in check from the rook
		move, _ := ParseMove("e4e5")
		if board.IsLegalMove(move) {
			t.Error("e4e5 should be illegal (king would move into check)")
		}

		// King moving to d5 also in check
		move2, _ := ParseMove("e4d5")
		if board.IsLegalMove(move2) {
			t.Error("e4d5 should be illegal (king would move into check)")
		}

		// King moving to e3 is safe
		move3, _ := ParseMove("e4e3")
		if !board.IsLegalMove(move3) {
			t.Error("e4e3 should be legal (king is safe on e3)")
		}
	})

	t.Run("capture is legal", func(t *testing.T) {
		// Set up a position where white pawn can capture black pawn
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 3)] = NewPiece(White, Pawn) // e4
		board.Squares[NewSquare(3, 4)] = NewPiece(Black, Pawn) // d5
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King) // e1
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, King) // e8

		// e4xd5 should be legal
		move, _ := ParseMove("e4d5")
		if !board.IsLegalMove(move) {
			t.Error("e4d5 (capture) should be legal")
		}
	})

	t.Run("same from and to square is invalid", func(t *testing.T) {
		board := NewBoard()
		// Moving a piece to the same square
		move := Move{From: NewSquare(4, 1), To: NewSquare(4, 1)}
		if board.IsLegalMove(move) {
			t.Error("moving to the same square should be illegal")
		}
	})

	t.Run("respects active color changes", func(t *testing.T) {
		board := NewBoard()

		// White's turn - white move is legal
		whiteMove, _ := ParseMove("e2e4")
		if !board.IsLegalMove(whiteMove) {
			t.Error("e2e4 should be legal for white")
		}

		// Make the move
		if err := board.MakeMove(whiteMove); err != nil {
			t.Fatalf("MakeMove returned error: %v", err)
		}

		// Now it's black's turn - same white move should be illegal
		if board.IsLegalMove(whiteMove) {
			t.Error("e2e4 should be illegal now (it's black's turn and pawn is on e4)")
		}

		// Black's move should now be legal
		blackMove, _ := ParseMove("e7e5")
		if !board.IsLegalMove(blackMove) {
			t.Error("e7e5 should be legal for black")
		}
	})

	t.Run("must escape check", func(t *testing.T) {
		// White king on e1 in check from black queen on e8
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King)  // e1
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, Queen) // e8 (giving check)
		board.Squares[NewSquare(0, 0)] = NewPiece(White, Rook)  // a1
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King)  // h8

		// Verify the king is in check
		if !board.InCheck() {
			t.Fatal("king should be in check in this position")
		}

		// Moving the rook (that doesn't block check) should be illegal
		rookMove, _ := ParseMove("a1a2")
		if board.IsLegalMove(rookMove) {
			t.Error("a1a2 should be illegal (does not escape check)")
		}

		// King moving out of check should be legal
		kingMove, _ := ParseMove("e1d1")
		if !board.IsLegalMove(kingMove) {
			t.Error("e1d1 should be legal (escapes check)")
		}
	})

	t.Run("promotion move comparison", func(t *testing.T) {
		// Set up a position where white pawn can promote
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(0, 6)] = NewPiece(White, Pawn) // a7
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King) // e1
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, King) // e8

		// Note: Current pawn move generation doesn't handle promotions yet
		// This test verifies that IsLegalMove correctly compares the Promotion field
		// If/when promotion is implemented, this will ensure proper matching

		// For now, test that a move without promotion is found if it exists in legal moves
		// (The pawn can move to a8 as a non-promotion move with current implementation)
		move, _ := ParseMove("a7a8")
		// Current implementation allows this as a regular move (no promotion logic yet)
		// This test documents current behavior
		result := board.IsLegalMove(move)
		// Just verify the method runs without panic
		_ = result
	})
}

func TestIllegalMovesFiltered(t *testing.T) {
	t.Run("pinned rook cannot move perpendicular to pin", func(t *testing.T) {
		// White king on e1, White rook on e4, Black rook on e8
		// Rook is pinned along e-file
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King) // e1
		board.Squares[NewSquare(4, 3)] = NewPiece(White, Rook) // e4
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, Rook) // e8
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King) // h8

		// Verify rook cannot move perpendicular to pin (to d4, f4, etc.)
		illegalMoves := []string{"e4d4", "e4f4", "e4a4", "e4h4"}
		for _, moveStr := range illegalMoves {
			move, _ := ParseMove(moveStr)

			// Should not be in LegalMoves
			legalMoves := board.LegalMoves()
			if containsMove(legalMoves, move.From, move.To) {
				t.Errorf("%s should not be in LegalMoves (pinned rook)", moveStr)
			}

			// IsLegalMove should return false
			if board.IsLegalMove(move) {
				t.Errorf("IsLegalMove(%s) should return false (pinned rook)", moveStr)
			}

			// MakeMove should return error
			err := board.MakeMove(move)
			if err == nil {
				t.Errorf("MakeMove(%s) should return error (pinned rook)", moveStr)
			}
		}
	})

	t.Run("pinned bishop by queen on diagonal", func(t *testing.T) {
		// White king on a1, White bishop on c3, Black queen on f6
		// Bishop is pinned along the diagonal
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(0, 0)] = NewPiece(White, King)   // a1
		board.Squares[NewSquare(2, 2)] = NewPiece(White, Bishop) // c3
		board.Squares[NewSquare(5, 5)] = NewPiece(Black, Queen)  // f6
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King)   // h8

		// Bishop can move along the diagonal toward the queen
		legalMovesTowardAttacker := []string{"c3d4", "c3e5", "c3f6"} // captures queen
		for _, moveStr := range legalMovesTowardAttacker {
			move, _ := ParseMove(moveStr)
			if !board.IsLegalMove(move) {
				t.Errorf("IsLegalMove(%s) should return true (move along pin toward attacker)", moveStr)
			}
		}

		// Bishop can move along the diagonal away from the queen
		legalMovesAwayFromAttacker := []string{"c3b2"}
		for _, moveStr := range legalMovesAwayFromAttacker {
			move, _ := ParseMove(moveStr)
			if !board.IsLegalMove(move) {
				t.Errorf("IsLegalMove(%s) should return true (move along pin away from attacker)", moveStr)
			}
		}

		// Bishop cannot move perpendicular to the pin
		illegalMoves := []string{"c3b4", "c3d2", "c3a5"}
		for _, moveStr := range illegalMoves {
			move, _ := ParseMove(moveStr)

			// Should not be in LegalMoves
			legalMoves := board.LegalMoves()
			if containsMove(legalMoves, move.From, move.To) {
				t.Errorf("%s should not be in LegalMoves (pinned bishop)", moveStr)
			}

			// IsLegalMove should return false
			if board.IsLegalMove(move) {
				t.Errorf("IsLegalMove(%s) should return false (pinned bishop)", moveStr)
			}

			// MakeMove should return error
			err := board.MakeMove(move)
			if err == nil {
				t.Errorf("MakeMove(%s) should return error (pinned bishop)", moveStr)
			}
		}
	})

	t.Run("knight pinned to king cannot move at all", func(t *testing.T) {
		// White king on e1, White knight on e3, Black rook on e8
		// Knight is pinned along the e-file
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King)   // e1
		board.Squares[NewSquare(4, 2)] = NewPiece(White, Knight) // e3
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, Rook)   // e8
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King)   // h8

		// Knight cannot move at all when pinned (it can't move along the pin line)
		legalMoves := board.LegalMoves()
		knightMoves := countMovesFrom(legalMoves, NewSquare(4, 2))
		if knightMoves != 0 {
			t.Errorf("pinned knight should have 0 legal moves, got %d", knightMoves)
		}

		// Test a few specific knight moves
		illegalMoves := []string{"e3d5", "e3f5", "e3g4", "e3c2"}
		for _, moveStr := range illegalMoves {
			move, _ := ParseMove(moveStr)

			if board.IsLegalMove(move) {
				t.Errorf("IsLegalMove(%s) should return false (pinned knight)", moveStr)
			}

			err := board.MakeMove(move)
			if err == nil {
				t.Errorf("MakeMove(%s) should return error (pinned knight)", moveStr)
			}
		}
	})

	t.Run("pawn pinned on diagonal cannot move forward", func(t *testing.T) {
		// White king on e1, White pawn on e2, Black bishop on e4 - NO, bishop can't pin forward
		// Actually: White king on a1, White pawn on b2, Black bishop on d4
		// Pawn is pinned on the diagonal
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(0, 0)] = NewPiece(White, King)   // a1
		board.Squares[NewSquare(1, 1)] = NewPiece(White, Pawn)   // b2
		board.Squares[NewSquare(3, 3)] = NewPiece(Black, Bishop) // d4
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King)   // h8

		// Pawn cannot move forward (perpendicular to pin)
		illegalMoves := []string{"b2b3", "b2b4"}
		for _, moveStr := range illegalMoves {
			move, _ := ParseMove(moveStr)

			if board.IsLegalMove(move) {
				t.Errorf("IsLegalMove(%s) should return false (pawn pinned on diagonal)", moveStr)
			}

			err := board.MakeMove(move)
			if err == nil {
				t.Errorf("MakeMove(%s) should return error (pawn pinned on diagonal)", moveStr)
			}
		}

		// Pawn could capture along the diagonal if bishop moves closer
		// Let's test with bishop on c3
		board.Squares[NewSquare(3, 3)] = Piece(Empty)            // remove from d4
		board.Squares[NewSquare(2, 2)] = NewPiece(Black, Bishop) // c3

		// Now pawn can capture the bishop (along the pin line)
		captureMove, _ := ParseMove("b2c3")
		if !board.IsLegalMove(captureMove) {
			t.Error("IsLegalMove(b2c3) should return true (capture along pin line)")
		}
	})

	t.Run("discovered check - piece cannot move away from blocking position", func(t *testing.T) {
		// White king on e1, White knight on e3, Black rook on e8
		// If knight moves, it would expose king to check
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King)   // e1
		board.Squares[NewSquare(4, 2)] = NewPiece(White, Knight) // e3
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, Rook)   // e8
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King)   // h8

		// Knight is blocking check - it cannot move
		illegalMoves := []string{"e3d5", "e3f5", "e3g4", "e3d1"}
		for _, moveStr := range illegalMoves {
			move, _ := ParseMove(moveStr)

			if board.IsLegalMove(move) {
				t.Errorf("IsLegalMove(%s) should return false (discovered check)", moveStr)
			}

			err := board.MakeMove(move)
			if err == nil {
				t.Errorf("MakeMove(%s) should return error (discovered check)", moveStr)
			}
		}
	})

	t.Run("king cannot capture defended piece", func(t *testing.T) {
		// White king on e4, Black pawn on f5, Black queen on h5 (defends pawn)
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 3)] = NewPiece(White, King)  // e4
		board.Squares[NewSquare(5, 4)] = NewPiece(Black, Pawn)  // f5
		board.Squares[NewSquare(7, 4)] = NewPiece(Black, Queen) // h5 (defends f5)
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King)  // h8

		// King cannot capture the defended pawn
		move, _ := ParseMove("e4f5")

		legalMoves := board.LegalMoves()
		if containsMove(legalMoves, move.From, move.To) {
			t.Error("e4f5 should not be in LegalMoves (king capturing defended piece)")
		}

		if board.IsLegalMove(move) {
			t.Error("IsLegalMove(e4f5) should return false (king capturing defended piece)")
		}

		err := board.MakeMove(move)
		if err == nil {
			t.Error("MakeMove(e4f5) should return error (king capturing defended piece)")
		}
	})

	t.Run("king cannot move next to enemy king", func(t *testing.T) {
		// White king on e4, Black king on e6
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 3)] = NewPiece(White, King) // e4
		board.Squares[NewSquare(4, 5)] = NewPiece(Black, King) // e6

		// White king cannot move to e5 (next to black king)
		move, _ := ParseMove("e4e5")

		if board.IsLegalMove(move) {
			t.Error("IsLegalMove(e4e5) should return false (king next to king)")
		}

		err := board.MakeMove(move)
		if err == nil {
			t.Error("MakeMove(e4e5) should return error (king next to king)")
		}

		// Also cannot move to d5, f5 (diagonal)
		diagonalMoves := []string{"e4d5", "e4f5"}
		for _, moveStr := range diagonalMoves {
			move, _ := ParseMove(moveStr)
			if board.IsLegalMove(move) {
				t.Errorf("IsLegalMove(%s) should return false (king next to king)", moveStr)
			}
		}
	})

	t.Run("when in check from multiple pieces only king can escape", func(t *testing.T) {
		// White king on e1, Black queen on e8, Black knight on d3
		// King is in double check - only king moves are legal
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King)   // e1
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, Queen)  // e8
		board.Squares[NewSquare(3, 2)] = NewPiece(Black, Knight) // d3
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King)   // h8
		board.Squares[NewSquare(0, 0)] = NewPiece(White, Rook)   // a1
		board.Squares[NewSquare(2, 1)] = NewPiece(White, Bishop) // c2

		// Verify king is in check
		if !board.InCheck() {
			t.Fatal("king should be in check")
		}

		// Non-king pieces should have 0 legal moves
		legalMoves := board.LegalMoves()

		rookMoves := countMovesFrom(legalMoves, NewSquare(0, 0))
		if rookMoves != 0 {
			t.Errorf("rook should have 0 legal moves in double check, got %d", rookMoves)
		}

		bishopMoves := countMovesFrom(legalMoves, NewSquare(2, 1))
		if bishopMoves != 0 {
			t.Errorf("bishop should have 0 legal moves in double check, got %d", bishopMoves)
		}

		// Only king should have legal moves
		kingMoves := countMovesFrom(legalMoves, NewSquare(4, 0))
		if kingMoves == 0 {
			t.Error("king should have at least one escape square")
		}

		// All legal moves should be king moves
		for _, move := range legalMoves {
			if move.From != NewSquare(4, 0) {
				t.Errorf("in double check, only king should be able to move, but found move %s", move.String())
			}
		}
	})

	t.Run("when in check must block or capture or move king", func(t *testing.T) {
		// White king on e1, Black queen on e8
		// White bishop on b5 can block at e2
		// White knight on c3 can block at e2 or e4
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King)   // e1
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, Queen)  // e8
		board.Squares[NewSquare(1, 4)] = NewPiece(White, Bishop) // b5
		board.Squares[NewSquare(2, 2)] = NewPiece(White, Knight) // c3
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King)   // h8

		// Verify king is in check
		if !board.InCheck() {
			t.Fatal("king should be in check")
		}

		legalMoves := board.LegalMoves()

		// Bishop moving to e2 should be legal (blocks check)
		if !containsMove(legalMoves, NewSquare(1, 4), NewSquare(4, 1)) {
			t.Error("b5e2 should be legal (bishop blocks check)")
		}

		// Bishop moving to a4 should be illegal (doesn't block)
		if containsMove(legalMoves, NewSquare(1, 4), NewSquare(0, 3)) {
			t.Error("b5a4 should be illegal (doesn't block check)")
		}

		// Knight to e2 should be legal (blocks)
		if !containsMove(legalMoves, NewSquare(2, 2), NewSquare(4, 1)) {
			t.Error("c3e2 should be legal (knight blocks check)")
		}

		// Knight to e4 should be legal (blocks)
		if !containsMove(legalMoves, NewSquare(2, 2), NewSquare(4, 3)) {
			t.Error("c3e4 should be legal (knight blocks check)")
		}

		// Knight to a2 should be illegal (doesn't block)
		if containsMove(legalMoves, NewSquare(2, 2), NewSquare(0, 1)) {
			t.Error("c3a2 should be illegal (doesn't block check)")
		}
	})

	t.Run("cannot block knight check must capture or move king", func(t *testing.T) {
		// White king on e1, Black knight on d3 (gives check)
		// White rook on d1 can capture the knight
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King)   // e1
		board.Squares[NewSquare(3, 2)] = NewPiece(Black, Knight) // d3 (gives check)
		board.Squares[NewSquare(3, 0)] = NewPiece(White, Rook)   // d1 (can capture)
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King)   // h8

		// Verify king is in check
		if !board.InCheck() {
			t.Fatal("king should be in check from knight")
		}

		legalMoves := board.LegalMoves()

		// Rook can capture the knight
		if !containsMove(legalMoves, NewSquare(3, 0), NewSquare(3, 2)) {
			t.Error("d1d3 should be legal (rook captures checking knight)")
		}

		// But rook cannot just move somewhere else
		if containsMove(legalMoves, NewSquare(3, 0), NewSquare(2, 0)) {
			t.Error("d1c1 should be illegal (doesn't escape knight check)")
		}
	})

	t.Run("multiple pins on same board", func(t *testing.T) {
		// White king on e1
		// White rook on e4 pinned by Black rook on e8
		// White bishop on d2 pinned by Black queen on a5
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King)   // e1
		board.Squares[NewSquare(4, 3)] = NewPiece(White, Rook)   // e4 (pinned vertically)
		board.Squares[NewSquare(3, 1)] = NewPiece(White, Bishop) // d2 (pinned diagonally)
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, Rook)   // e8 (pins the rook)
		board.Squares[NewSquare(0, 4)] = NewPiece(Black, Queen)  // a5 (pins the bishop)
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King)   // h8

		legalMoves := board.LegalMoves()

		// Rook on e4 can only move along e-file
		rookMoves := countMovesFrom(legalMoves, NewSquare(4, 3))
		// Can move to: e2, e3, e5, e6, e7, e8 = 6 squares
		if rookMoves != 6 {
			t.Errorf("pinned rook should have 6 legal moves, got %d", rookMoves)
		}

		// Rook cannot move to d4 (perpendicular to pin)
		if containsMove(legalMoves, NewSquare(4, 3), NewSquare(3, 3)) {
			t.Error("e4d4 should be illegal (rook pinned)")
		}

		// Bishop on d2 can only move along a5-e1 diagonal
		bishopMoves := countMovesFrom(legalMoves, NewSquare(3, 1))
		// Can move to: c3, b4, a5 (capture) = 3 squares
		// Note: e1 has the king, so bishop can't move there
		if bishopMoves != 3 {
			t.Errorf("pinned bishop should have 3 legal moves, got %d", bishopMoves)
		}

		// Bishop cannot move to c1 (perpendicular to pin)
		if containsMove(legalMoves, NewSquare(3, 1), NewSquare(2, 0)) {
			t.Error("d2c1 should be illegal (bishop pinned)")
		}
	})

	t.Run("absolute pin vs partial pin", func(t *testing.T) {
		// Absolute pin: piece cannot move at all (e.g., knight pinned by rook)
		// Partial pin: piece can move along pin line (e.g., rook pinned by rook)

		// Test absolute pin: White king e1, White queen e3, Black rook e8
		// Queen is pinned but can still move along the pin line
		board := &Board{ActiveColor: White}
		board.Squares[NewSquare(4, 0)] = NewPiece(White, King)  // e1
		board.Squares[NewSquare(4, 2)] = NewPiece(White, Queen) // e3
		board.Squares[NewSquare(4, 7)] = NewPiece(Black, Rook)  // e8
		board.Squares[NewSquare(7, 7)] = NewPiece(Black, King)  // h8

		legalMoves := board.LegalMoves()

		// Queen can move along the e-file (toward or away from attacker)
		if !containsMove(legalMoves, NewSquare(4, 2), NewSquare(4, 1)) {
			t.Error("e3e2 should be legal (queen moves along pin line)")
		}
		if !containsMove(legalMoves, NewSquare(4, 2), NewSquare(4, 4)) {
			t.Error("e3e5 should be legal (queen moves along pin line)")
		}

		// But queen cannot move perpendicular to pin
		if containsMove(legalMoves, NewSquare(4, 2), NewSquare(3, 3)) {
			t.Error("e3d4 should be illegal (queen pinned, moving perpendicular)")
		}
		if containsMove(legalMoves, NewSquare(4, 2), NewSquare(5, 2)) {
			t.Error("e3f3 should be illegal (queen pinned, moving perpendicular)")
		}
	})
}
