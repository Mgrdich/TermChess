package bot

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Mgrdich/TermChess/internal/engine"
)

// mockInferenceSession is a test double for rlInferenceSession.
type mockInferenceSession struct {
	policy []float32
	value  float32
	err    error
	closed bool
}

func (m *mockInferenceSession) RunInference(input []float32) ([]float32, float32, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	return m.policy, m.value, nil
}

func (m *mockInferenceSession) Close() error {
	m.closed = true
	return nil
}

// TestRLEngineInterface is a compile-time check that *rlEngine implements Engine.
func TestRLEngineInterface(t *testing.T) {
	var _ Engine = (*rlEngine)(nil)
}

// TestRLEngineInspectable is a compile-time check that *rlEngine implements Inspectable.
func TestRLEngineInspectable(t *testing.T) {
	var _ Inspectable = (*rlEngine)(nil)
}

// TestNewRLEngine_Defaults verifies that each RLDifficulty creates an engine
// with the correct name and no error from the factory.
func TestNewRLEngine_Defaults(t *testing.T) {
	tests := []struct {
		difficulty RLDifficulty
		wantName   string
	}{
		{RLBeginner, "RL Beginner (1000)"},
		{RLIntermediate, "RL Intermediate (1200)"},
		{RLClub, "RL Club (1500)"},
		{RLAdvanced, "RL Advanced (2000)"},
		{RLMaster, "RL Master (2200)"},
	}

	for _, tc := range tests {
		t.Run(tc.wantName, func(t *testing.T) {
			eng, err := NewRLEngine(tc.difficulty)
			if err != nil {
				t.Fatalf("NewRLEngine(%d) error = %v, want nil", tc.difficulty, err)
			}
			defer eng.Close()

			if eng.Name() != tc.wantName {
				t.Errorf("Name() = %q, want %q", eng.Name(), tc.wantName)
			}
		})
	}
}

// TestNewRLEngine_InvalidDifficulty verifies that an invalid difficulty
// returns an error from the factory.
func TestNewRLEngine_InvalidDifficulty(t *testing.T) {
	eng, err := NewRLEngine(RLDifficulty(99))
	if err == nil {
		t.Error("NewRLEngine(99) error = nil, want error for invalid difficulty")
	}
	if eng != nil {
		t.Errorf("NewRLEngine(99) engine = %v, want nil", eng)
	}
}

// TestNewRLEngine_WithTimeLimit verifies that a custom time limit
// option overrides the default.
func TestNewRLEngine_WithTimeLimit(t *testing.T) {
	eng, err := NewRLEngine(RLIntermediate, WithTimeLimit(15*time.Second))
	if err != nil {
		t.Fatalf("NewRLEngine with WithTimeLimit error = %v, want nil", err)
	}
	defer eng.Close()

	rlEng, ok := eng.(*rlEngine)
	if !ok {
		t.Fatal("Expected engine to be *rlEngine")
	}

	if rlEng.timeLimit != 15*time.Second {
		t.Errorf("timeLimit = %v, want 15s", rlEng.timeLimit)
	}
}

// TestRLEngine_SelectMove_ReturnsModelNotLoaded verifies that SelectMove
// returns ErrModelNotLoaded since the ONNX model is not yet integrated.
func TestRLEngine_SelectMove_ReturnsModelNotLoaded(t *testing.T) {
	eng, err := NewRLEngine(RLIntermediate)
	if err != nil {
		t.Fatalf("NewRLEngine error = %v, want nil", err)
	}
	defer eng.Close()

	board := engine.NewBoard()

	_, err = eng.SelectMove(context.Background(), board)
	if err == nil {
		t.Fatal("SelectMove error = nil, want ErrModelNotLoaded")
	}

	if !errors.Is(err, ErrModelNotLoaded) {
		t.Errorf("SelectMove error = %v, want ErrModelNotLoaded", err)
	}
}

// TestRLEngine_Close verifies that Close works and that calling it
// a second time returns an error.
func TestRLEngine_Close(t *testing.T) {
	eng, err := NewRLEngine(RLAdvanced)
	if err != nil {
		t.Fatalf("NewRLEngine error = %v, want nil", err)
	}

	// First close should succeed
	err = eng.Close()
	if err != nil {
		t.Errorf("First Close() error = %v, want nil", err)
	}

	// Second close should return an error
	err = eng.Close()
	if err == nil {
		t.Error("Second Close() error = nil, want error for already closed")
	}
}

// TestRLEngine_Info verifies that Info() returns correct metadata.
func TestRLEngine_Info(t *testing.T) {
	eng, err := NewRLEngine(RLMaster)
	if err != nil {
		t.Fatalf("NewRLEngine error = %v, want nil", err)
	}
	defer eng.Close()

	inspectable, ok := eng.(Inspectable)
	if !ok {
		t.Fatal("rlEngine should implement Inspectable interface")
	}

	info := inspectable.Info()

	if info.Type != TypeRL {
		t.Errorf("Info().Type = %v, want TypeRL", info.Type)
	}

	if info.Name != "RL Master (2200)" {
		t.Errorf("Info().Name = %q, want %q", info.Name, "RL Master (2200)")
	}

	if info.Features == nil {
		t.Fatal("Info().Features = nil, want non-nil map")
	}

	if info.Features["onnx"] != false {
		t.Errorf("Info().Features[\"onnx\"] = %v, want false", info.Features["onnx"])
	}
}

// TestRLEngine_SelectMove_AfterClose verifies that SelectMove returns
// an error after the engine has been closed.
func TestRLEngine_SelectMove_AfterClose(t *testing.T) {
	eng, err := NewRLEngine(RLIntermediate)
	if err != nil {
		t.Fatalf("NewRLEngine error = %v, want nil", err)
	}

	err = eng.Close()
	if err != nil {
		t.Fatalf("Close() error = %v, want nil", err)
	}

	board := engine.NewBoard()

	_, err = eng.SelectMove(context.Background(), board)
	if err == nil {
		t.Error("SelectMove after Close error = nil, want error")
	}
	if err != nil && err.Error() != "engine is closed" {
		t.Errorf("SelectMove after Close error = %q, want %q", err.Error(), "engine is closed")
	}
}

// TestMoveToPolicyIndex verifies the from-to policy index encoding.
func TestMoveToPolicyIndex(t *testing.T) {
	tests := []struct {
		name string
		move engine.Move
		want int
	}{
		{
			name: "e2e4",
			move: engine.Move{From: engine.NewSquare(4, 1), To: engine.NewSquare(4, 3)},
			want: 12*64 + 28, // from=12 (rank1*8+file4), to=28 (rank3*8+file4) = 796
		},
		{
			name: "a1a2",
			move: engine.Move{From: engine.NewSquare(0, 0), To: engine.NewSquare(0, 1)},
			want: 0*64 + 8, // from=0, to=8 = 8
		},
		{
			name: "h8h7",
			move: engine.Move{From: engine.NewSquare(7, 7), To: engine.NewSquare(7, 6)},
			want: 63*64 + 55, // from=63, to=55 = 4087
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := moveToPolicyIndex(tc.move)
			if got != tc.want {
				t.Errorf("moveToPolicyIndex(%v) = %d, want %d", tc.move, got, tc.want)
			}
		})
	}
}

// TestSelectBestMove_PicksHighestScore verifies that the move with
// the highest policy score is selected from the legal moves.
func TestSelectBestMove_PicksHighestScore(t *testing.T) {
	// Create three legal moves
	moveA := engine.Move{From: engine.NewSquare(4, 1), To: engine.NewSquare(4, 3)} // e2e4, idx=796
	moveB := engine.Move{From: engine.NewSquare(3, 1), To: engine.NewSquare(3, 3)} // d2d4, idx=11*64+27=731
	moveC := engine.Move{From: engine.NewSquare(6, 0), To: engine.NewSquare(5, 2)} // g1f3, idx=6*64+21=405

	legalMoves := []engine.Move{moveA, moveB, moveC}

	policy := make([]float32, policySize)
	policy[moveToPolicyIndex(moveA)] = 1.0
	policy[moveToPolicyIndex(moveB)] = 5.0 // Highest score
	policy[moveToPolicyIndex(moveC)] = 2.0

	got, err := selectBestMove(legalMoves, policy)
	if err != nil {
		t.Fatalf("selectBestMove error = %v, want nil", err)
	}

	if got.From != moveB.From || got.To != moveB.To {
		t.Errorf("selectBestMove picked %v, want %v (d2d4 with highest score)", got, moveB)
	}
}

// TestSelectBestMove_LegalMovesMasking verifies that illegal moves with higher
// scores are effectively masked out because only legal moves are considered.
func TestSelectBestMove_LegalMovesMasking(t *testing.T) {
	// Only two legal moves
	legalA := engine.Move{From: engine.NewSquare(4, 1), To: engine.NewSquare(4, 3)} // e2e4, idx=796
	legalB := engine.Move{From: engine.NewSquare(3, 1), To: engine.NewSquare(3, 3)} // d2d4, idx=731

	legalMoves := []engine.Move{legalA, legalB}

	policy := make([]float32, policySize)
	// Give high scores to illegal moves
	policy[0] = 100.0   // a1a1 - not a legal move
	policy[100] = 50.0  // some other illegal move
	policy[2000] = 80.0 // another illegal move

	// Legal moves get lower scores
	policy[moveToPolicyIndex(legalA)] = 3.0
	policy[moveToPolicyIndex(legalB)] = 7.0 // Best among legal moves

	got, err := selectBestMove(legalMoves, policy)
	if err != nil {
		t.Fatalf("selectBestMove error = %v, want nil", err)
	}

	if got.From != legalB.From || got.To != legalB.To {
		t.Errorf("selectBestMove picked %v, want %v (d2d4, best legal move)", got, legalB)
	}
}

// TestSelectBestMove_QueenPromotionPreference verifies that when multiple
// promotion moves share the same policy index and have equal scores,
// the queen promotion is preferred.
func TestSelectBestMove_QueenPromotionPreference(t *testing.T) {
	// Pawn on a7 promoting to a8 — all 4 promotions map to the same policy index
	from := engine.NewSquare(0, 6) // a7
	to := engine.NewSquare(0, 7)   // a8
	idx := moveToPolicyIndex(engine.Move{From: from, To: to})

	promoMoves := []engine.Move{
		{From: from, To: to, Promotion: engine.Rook},
		{From: from, To: to, Promotion: engine.Knight},
		{From: from, To: to, Promotion: engine.Bishop},
		{From: from, To: to, Promotion: engine.Queen},
	}

	policy := make([]float32, policySize)
	policy[idx] = 5.0 // All promotions share this score

	got, err := selectBestMove(promoMoves, policy)
	if err != nil {
		t.Fatalf("selectBestMove error = %v, want nil", err)
	}

	if got.Promotion != engine.Queen {
		t.Errorf("selectBestMove promotion = %v, want Queen", got.Promotion)
	}
}

// TestRLEngine_SelectMove_WithMockSession verifies that SelectMove
// correctly runs inference via the session and returns the best legal move.
func TestRLEngine_SelectMove_WithMockSession(t *testing.T) {
	eng, err := NewRLEngine(RLIntermediate)
	if err != nil {
		t.Fatalf("NewRLEngine error = %v", err)
	}
	defer eng.Close()

	// Inject mock session with a policy that favors e2e4 (index 796)
	policy := make([]float32, policySize)
	policy[796] = 10.0 // e2e4 gets highest score

	rlEng := eng.(*rlEngine)
	rlEng.session = &mockInferenceSession{
		policy: policy,
		value:  0.5,
	}

	board := engine.NewBoard()
	move, err := eng.SelectMove(context.Background(), board)
	if err != nil {
		t.Fatalf("SelectMove error = %v", err)
	}

	// Verify the selected move is e2e4
	expectedFrom := engine.NewSquare(4, 1) // e2
	expectedTo := engine.NewSquare(4, 3)   // e4
	if move.From != expectedFrom || move.To != expectedTo {
		t.Errorf("SelectMove = %v, want e2e4 (from=%v, to=%v)", move, expectedFrom, expectedTo)
	}
}

// TestRLEngine_SelectMove_SingleLegalMove verifies that when there is only
// one legal move, it is returned immediately without running inference.
func TestRLEngine_SelectMove_SingleLegalMove(t *testing.T) {
	// Position where Black king has only one legal move.
	// FEN: "K7/8/1k6/8/8/8/8/8 b - - 0 1" - Black king on b6, can move to many squares
	// Use a more constrained position: Black king in corner with limited moves.
	// FEN: "8/8/8/8/8/8/1Q6/k1K5 b - - 0 1"
	// Black king on a1, white queen on b2, white king on c1.
	// Black king's only legal move would depend on the position.

	// Let's use a simpler approach: set up a position via FEN where
	// the side to move has exactly one legal move.
	// "k7/1R6/1K6/8/8/8/8/8 b - - 0 1"
	// Black king on a8, white rook on b7, white king on b6.
	// Black king can only go to... let's check: a8 is black king.
	// Squares around a8: a7 (attacked by Rb7), b8 (attacked by Rb7 and Kb6), b7 (white rook).
	// Actually that's stalemate or checkmate. Let me pick something simpler.

	// Use a position where we inject a mock session that will error if called,
	// to verify inference is NOT called for single legal move.
	// "8/8/8/8/8/b1k5/8/K7 w - - 0 1"
	// White king on a1. Bishop on a3 covers b2 and c1. King on c3 covers b2, d2.
	// a1 king can go to: a2 (not attacked), b1 (not attacked by anything? c3 king doesn't reach b1).
	// That's 2 moves. We need exactly 1.

	// Simplest: just create an engine with a mock that panics on RunInference,
	// and manually test with a board. If there's exactly one legal move,
	// inference should not be called.

	eng, err := NewRLEngine(RLIntermediate)
	if err != nil {
		t.Fatalf("NewRLEngine error = %v", err)
	}
	defer eng.Close()

	// Use a FEN position where only one move is legal
	// "K1k5/8/8/8/8/8/8/1r6 w - - 0 1"
	// White king on a8, black king on c8, black rook on b1.
	// King on a8: possible squares a7 (attacked by Rb1? No, rook on b1 attacks b-file and rank 1).
	// a7 is safe, b8 is attacked by Kc8, b7 is attacked by Kc8.
	// So white king has only a7. That's one legal move.
	board, fenErr := engine.ParseFEN("K1k5/8/8/8/8/8/8/1r6 w - - 0 1")
	if fenErr != nil {
		t.Fatalf("ParseFEN error = %v", fenErr)
	}

	legalMoves := board.LegalMoves()
	if len(legalMoves) != 1 {
		t.Fatalf("Expected exactly 1 legal move, got %d: %v", len(legalMoves), legalMoves)
	}

	// Inject a mock session that returns an error if called —
	// this verifies that inference is NOT invoked for a single legal move.
	rlEng := eng.(*rlEngine)
	rlEng.session = &mockInferenceSession{
		err: fmt.Errorf("should not be called"),
	}

	move, err := eng.SelectMove(context.Background(), board)
	if err != nil {
		t.Fatalf("SelectMove error = %v, want nil", err)
	}

	if move.From != legalMoves[0].From || move.To != legalMoves[0].To {
		t.Errorf("SelectMove = %v, want %v", move, legalMoves[0])
	}
}

// TestRLEngine_SelectMove_InferenceError verifies that when the inference
// session returns an error, SelectMove propagates it.
func TestRLEngine_SelectMove_InferenceError(t *testing.T) {
	eng, err := NewRLEngine(RLIntermediate)
	if err != nil {
		t.Fatalf("NewRLEngine error = %v", err)
	}
	defer eng.Close()

	rlEng := eng.(*rlEngine)
	rlEng.session = &mockInferenceSession{
		err: fmt.Errorf("onnx runtime error"),
	}

	board := engine.NewBoard()
	_, err = eng.SelectMove(context.Background(), board)
	if err == nil {
		t.Fatal("SelectMove error = nil, want inference error")
	}

	wantMsg := "inference failed: onnx runtime error"
	if err.Error() != wantMsg {
		t.Errorf("SelectMove error = %q, want %q", err.Error(), wantMsg)
	}
}

// TestRLEngine_SelectMove_ClosedSession verifies that SelectMove returns
// an error after the engine has been closed, even with a mock session.
func TestRLEngine_SelectMove_ClosedSession(t *testing.T) {
	eng, err := NewRLEngine(RLIntermediate)
	if err != nil {
		t.Fatalf("NewRLEngine error = %v", err)
	}

	rlEng := eng.(*rlEngine)
	mock := &mockInferenceSession{
		policy: make([]float32, policySize),
		value:  0.0,
	}
	rlEng.session = mock

	// Close the engine
	err = eng.Close()
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Verify the mock session was closed
	if !mock.closed {
		t.Error("Expected mock session to be closed")
	}

	board := engine.NewBoard()
	_, err = eng.SelectMove(context.Background(), board)
	if err == nil {
		t.Fatal("SelectMove after Close error = nil, want error")
	}
	if err.Error() != "engine is closed" {
		t.Errorf("SelectMove after Close error = %q, want %q", err.Error(), "engine is closed")
	}
}
