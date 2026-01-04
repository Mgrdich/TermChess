package engine

import (
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
		board.MakeMove(move)

		if board.ActiveColor != Black {
			t.Error("active color should be Black after White's move")
		}

		// Black moves
		move2, _ := ParseMove("e7e5")
		board.MakeMove(move2)

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
		board.MakeMove(move)

		if board.FullMoveNum != 1 {
			t.Error("full move number should still be 1 after White's move")
		}

		// Black moves
		move2, _ := ParseMove("e7e5")
		board.MakeMove(move2)

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
}
