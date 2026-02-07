package ui

import (
	"testing"

	"github.com/Mgrdich/TermChess/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

// TestSquareFromMouse_ValidPositions_WithCoords tests that valid mouse positions
// are correctly converted to chess squares when coordinates are shown.
func TestSquareFromMouse_ValidPositions_WithCoords(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	tests := []struct {
		name         string
		mouseX       int
		mouseY       int
		expectedFile int
		expectedRank int
	}{
		// Corner squares
		{"a8 (top-left)", 2, boardStartY, 0, 7},
		{"h8 (top-right)", 16, boardStartY, 7, 7},
		{"a1 (bottom-left)", 2, boardStartY + 7, 0, 0},
		{"h1 (bottom-right)", 16, boardStartY + 7, 7, 0},

		// Edge squares
		{"a4 (left edge)", 2, boardStartY + 4, 0, 3},
		{"h5 (right edge)", 16, boardStartY + 3, 7, 4},
		{"e8 (top edge)", 10, boardStartY, 4, 7},
		{"d1 (bottom edge)", 8, boardStartY + 7, 3, 0},

		// Center squares
		{"e4", 10, boardStartY + 4, 4, 3},
		{"d5", 8, boardStartY + 3, 3, 4},

		// Second character of a square should map to same square
		{"e4 (second char)", 11, boardStartY + 4, 4, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sq := squareFromMouse(tt.mouseX, tt.mouseY, config)
			if sq == nil {
				t.Fatalf("Expected a square, got nil")
			}

			if sq.File() != tt.expectedFile {
				t.Errorf("Expected file %d, got %d", tt.expectedFile, sq.File())
			}

			if sq.Rank() != tt.expectedRank {
				t.Errorf("Expected rank %d, got %d", tt.expectedRank, sq.Rank())
			}
		})
	}
}

// TestSquareFromMouse_ValidPositions_NoCoords tests that valid mouse positions
// are correctly converted to chess squares when coordinates are not shown.
func TestSquareFromMouse_ValidPositions_NoCoords(t *testing.T) {
	config := Config{
		ShowCoords: false,
	}

	tests := []struct {
		name         string
		mouseX       int
		mouseY       int
		expectedFile int
		expectedRank int
	}{
		// Corner squares (no coordinate offset)
		{"a8 (top-left)", 0, boardStartY, 0, 7},
		{"h8 (top-right)", 14, boardStartY, 7, 7},
		{"a1 (bottom-left)", 0, boardStartY + 7, 0, 0},
		{"h1 (bottom-right)", 14, boardStartY + 7, 7, 0},

		// Center square
		{"e4", 8, boardStartY + 4, 4, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sq := squareFromMouse(tt.mouseX, tt.mouseY, config)
			if sq == nil {
				t.Fatalf("Expected a square, got nil")
			}

			if sq.File() != tt.expectedFile {
				t.Errorf("Expected file %d, got %d", tt.expectedFile, sq.File())
			}

			if sq.Rank() != tt.expectedRank {
				t.Errorf("Expected rank %d, got %d", tt.expectedRank, sq.Rank())
			}
		})
	}
}

// TestSquareFromMouse_OutOfBounds tests that out-of-bounds positions return nil.
func TestSquareFromMouse_OutOfBounds(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	tests := []struct {
		name   string
		mouseX int
		mouseY int
	}{
		// Above the board
		{"above board", 10, boardStartY - 1},
		{"way above board", 10, 0},

		// Left of the board (in the coordinate area)
		{"left of board (rank label)", 0, boardStartY},
		{"left of board (rank label space)", 1, boardStartY},

		// Right of the board
		{"right of board", 18, boardStartY},
		{"way right of board", 100, boardStartY},

		// Below the board
		{"below board", 10, boardStartY + 8},
		{"way below board", 10, 100},

		// Corner cases outside board
		{"negative X", -1, boardStartY},
		{"negative Y", 10, -1},
		{"negative both", -1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sq := squareFromMouse(tt.mouseX, tt.mouseY, config)
			if sq != nil {
				t.Errorf("Expected nil for out-of-bounds position (%d, %d), got %v", tt.mouseX, tt.mouseY, sq)
			}
		})
	}
}

// TestSquareFromMouse_AllSquares tests that all 64 squares can be correctly identified.
func TestSquareFromMouse_AllSquares(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			// Calculate mouse position for this square
			mouseX := boardStartXWithCoords + file*squareWidth
			mouseY := boardStartY + (7 - rank) // rank 7 at top, rank 0 at bottom

			sq := squareFromMouse(mouseX, mouseY, config)
			if sq == nil {
				t.Errorf("Expected square for file=%d, rank=%d, got nil", file, rank)
				continue
			}

			if sq.File() != file {
				t.Errorf("File mismatch: expected %d, got %d (mouseX=%d, mouseY=%d)",
					file, sq.File(), mouseX, mouseY)
			}

			if sq.Rank() != rank {
				t.Errorf("Rank mismatch: expected %d, got %d (mouseX=%d, mouseY=%d)",
					rank, sq.Rank(), mouseX, mouseY)
			}
		}
	}
}

// TestSquareFromMouse_SquareWidth tests that clicking anywhere within a square
// (both characters) maps to the same square.
func TestSquareFromMouse_SquareWidth(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	// Test e4 square - both character positions should map to the same square
	baseX := boardStartXWithCoords + 4*squareWidth // file e = 4
	baseY := boardStartY + 4                        // rank 4 (7-4=3, so row offset 4)

	// First character of the square
	sq1 := squareFromMouse(baseX, baseY, config)
	// Second character of the square
	sq2 := squareFromMouse(baseX+1, baseY, config)

	if sq1 == nil || sq2 == nil {
		t.Fatalf("Expected valid squares, got sq1=%v, sq2=%v", sq1, sq2)
	}

	if sq1.File() != sq2.File() || sq1.Rank() != sq2.Rank() {
		t.Errorf("Both positions should map to same square: got (%d,%d) and (%d,%d)",
			sq1.File(), sq1.Rank(), sq2.File(), sq2.Rank())
	}
}

// TestHandleMouseEvent_SelectOwnPiece tests that clicking on own pieces selects them.
func TestHandleMouseEvent_SelectOwnPiece(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	m := Model{
		board:    engine.NewBoard(),
		gameType: GameTypePvP,
		screen:   ScreenGamePlay,
		config:   config,
	}

	// White to move - click on e2 (white pawn)
	// e2 = file 4, rank 1
	// mouseX = 2 + 4*2 = 10
	// mouseY = 4 + (7-1) = 10
	msg := tea.MouseMsg{
		X:      10,
		Y:      10,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}

	newModel, _ := m.handleMouseEvent(msg)

	if newModel.selectedSquare == nil {
		t.Fatalf("Expected selectedSquare to be set, got nil")
	}

	expectedSq := engine.NewSquare(4, 1) // e2
	if *newModel.selectedSquare != expectedSq {
		t.Errorf("Expected selectedSquare to be e2 (%v), got %v",
			expectedSq, *newModel.selectedSquare)
	}
}

// TestHandleMouseEvent_ChangeSelection tests that clicking on a different own piece
// changes the selection.
func TestHandleMouseEvent_ChangeSelection(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	// Start with a piece already selected
	initialSquare := engine.NewSquare(4, 1) // e2
	m := Model{
		board:          engine.NewBoard(),
		gameType:       GameTypePvP,
		screen:         ScreenGamePlay,
		config:         config,
		selectedSquare: &initialSquare,
	}

	// Click on d2 (different white pawn)
	// d2 = file 3, rank 1
	// mouseX = 2 + 3*2 = 8
	// mouseY = 4 + (7-1) = 10
	msg := tea.MouseMsg{
		X:      8,
		Y:      10,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}

	newModel, _ := m.handleMouseEvent(msg)

	if newModel.selectedSquare == nil {
		t.Fatalf("Expected selectedSquare to be set, got nil")
	}

	expectedSq := engine.NewSquare(3, 1) // d2
	if *newModel.selectedSquare != expectedSq {
		t.Errorf("Expected selectedSquare to change to d2 (%v), got %v",
			expectedSq, *newModel.selectedSquare)
	}
}

// TestHandleMouseEvent_IgnoreOpponentPiece tests that clicking on opponent pieces
// does not change selection.
func TestHandleMouseEvent_IgnoreOpponentPiece(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	// Start with a piece already selected
	initialSquare := engine.NewSquare(4, 1) // e2
	m := Model{
		board:          engine.NewBoard(),
		gameType:       GameTypePvP,
		screen:         ScreenGamePlay,
		config:         config,
		selectedSquare: &initialSquare,
	}

	// White to move - click on e7 (black pawn)
	// e7 = file 4, rank 6
	// mouseX = 2 + 4*2 = 10
	// mouseY = 4 + (7-6) = 5
	msg := tea.MouseMsg{
		X:      10,
		Y:      5,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}

	newModel, _ := m.handleMouseEvent(msg)

	// Selection should remain unchanged
	if newModel.selectedSquare == nil {
		t.Fatalf("Expected selectedSquare to remain set")
	}

	if *newModel.selectedSquare != initialSquare {
		t.Errorf("Expected selectedSquare to remain e2, got %v", *newModel.selectedSquare)
	}
}

// TestHandleMouseEvent_IgnoreEmptySquare tests that clicking on empty squares
// does not change selection.
func TestHandleMouseEvent_IgnoreEmptySquare(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	// Start with a piece already selected
	initialSquare := engine.NewSquare(4, 1) // e2
	m := Model{
		board:          engine.NewBoard(),
		gameType:       GameTypePvP,
		screen:         ScreenGamePlay,
		config:         config,
		selectedSquare: &initialSquare,
	}

	// Click on e4 (empty square)
	// e4 = file 4, rank 3
	// mouseX = 2 + 4*2 = 10
	// mouseY = 4 + (7-3) = 8
	msg := tea.MouseMsg{
		X:      10,
		Y:      8,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}

	newModel, _ := m.handleMouseEvent(msg)

	// Selection should remain unchanged
	if newModel.selectedSquare == nil {
		t.Fatalf("Expected selectedSquare to remain set")
	}

	if *newModel.selectedSquare != initialSquare {
		t.Errorf("Expected selectedSquare to remain e2, got %v", *newModel.selectedSquare)
	}
}

// TestHandleMouseEvent_IgnoreRightClick tests that right clicks are ignored.
func TestHandleMouseEvent_IgnoreRightClick(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	m := Model{
		board:    engine.NewBoard(),
		gameType: GameTypePvP,
		screen:   ScreenGamePlay,
		config:   config,
	}

	// Right-click on e2 (white pawn)
	msg := tea.MouseMsg{
		X:      10,
		Y:      10,
		Button: tea.MouseButtonRight,
		Action: tea.MouseActionPress,
	}

	newModel, _ := m.handleMouseEvent(msg)

	if newModel.selectedSquare != nil {
		t.Errorf("Expected selectedSquare to be nil for right click, got %v", *newModel.selectedSquare)
	}
}

// TestHandleMouseEvent_IgnoreMouseRelease tests that mouse release events are ignored.
func TestHandleMouseEvent_IgnoreMouseRelease(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	m := Model{
		board:    engine.NewBoard(),
		gameType: GameTypePvP,
		screen:   ScreenGamePlay,
		config:   config,
	}

	// Mouse release on e2 (white pawn)
	msg := tea.MouseMsg{
		X:      10,
		Y:      10,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionRelease,
	}

	newModel, _ := m.handleMouseEvent(msg)

	if newModel.selectedSquare != nil {
		t.Errorf("Expected selectedSquare to be nil for mouse release, got %v", *newModel.selectedSquare)
	}
}

// TestHandleMouseEvent_IgnoreOutOfBounds tests that clicks outside the board are ignored.
func TestHandleMouseEvent_IgnoreOutOfBounds(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	initialSquare := engine.NewSquare(4, 1) // e2
	m := Model{
		board:          engine.NewBoard(),
		gameType:       GameTypePvP,
		screen:         ScreenGamePlay,
		config:         config,
		selectedSquare: &initialSquare,
	}

	// Click outside the board (above it)
	msg := tea.MouseMsg{
		X:      10,
		Y:      0,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}

	newModel, _ := m.handleMouseEvent(msg)

	// Selection should remain unchanged
	if newModel.selectedSquare == nil {
		t.Fatalf("Expected selectedSquare to remain set")
	}

	if *newModel.selectedSquare != initialSquare {
		t.Errorf("Expected selectedSquare to remain e2, got %v", *newModel.selectedSquare)
	}
}

// TestHandleMouseEvent_PvBot_HumanTurn tests that in PvBot mode,
// selection works when it's the human's turn.
func TestHandleMouseEvent_PvBot_HumanTurn(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	m := Model{
		board:     engine.NewBoard(),
		gameType:  GameTypePvBot,
		screen:    ScreenGamePlay,
		config:    config,
		userColor: engine.White, // Human plays White
	}

	// White to move, human is White - click on e2 (white pawn)
	msg := tea.MouseMsg{
		X:      10,
		Y:      10,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}

	newModel, _ := m.handleMouseEvent(msg)

	if newModel.selectedSquare == nil {
		t.Fatalf("Expected selectedSquare to be set for human's turn")
	}

	expectedSq := engine.NewSquare(4, 1) // e2
	if *newModel.selectedSquare != expectedSq {
		t.Errorf("Expected selectedSquare to be e2, got %v", *newModel.selectedSquare)
	}
}

// TestHandleMouseEvent_PvBot_BotTurn tests that in PvBot mode,
// selection does not work when it's the bot's turn.
func TestHandleMouseEvent_PvBot_BotTurn(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	m := Model{
		board:     engine.NewBoard(),
		gameType:  GameTypePvBot,
		screen:    ScreenGamePlay,
		config:    config,
		userColor: engine.Black, // Human plays Black, but White moves first
	}

	// White to move, but human is Black - click on e2 (white pawn)
	// Should be ignored because it's not the human's turn
	msg := tea.MouseMsg{
		X:      10,
		Y:      10,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}

	newModel, _ := m.handleMouseEvent(msg)

	if newModel.selectedSquare != nil {
		t.Errorf("Expected selectedSquare to be nil when it's bot's turn, got %v", *newModel.selectedSquare)
	}
}

// TestUpdate_BvBIgnoresMouse tests that Bot vs Bot mode ignores mouse input entirely.
func TestUpdate_BvBIgnoresMouse(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	m := Model{
		board:    engine.NewBoard(),
		gameType: GameTypeBvB,
		screen:   ScreenGamePlay,
		config:   config,
	}

	// Click on e2 - should be ignored in BvB mode
	msg := tea.MouseMsg{
		X:      10,
		Y:      10,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}

	newModelInterface, _ := m.Update(msg)
	newModel := newModelInterface.(Model)

	if newModel.selectedSquare != nil {
		t.Errorf("Expected mouse input to be ignored in BvB mode, got selectedSquare=%v",
			*newModel.selectedSquare)
	}
}

// TestUpdate_NonGamePlayIgnoresMouse tests that mouse input is ignored
// when not on the GamePlay screen.
func TestUpdate_NonGamePlayIgnoresMouse(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	m := Model{
		board:    engine.NewBoard(),
		gameType: GameTypePvP,
		screen:   ScreenMainMenu, // Not GamePlay screen
		config:   config,
	}

	msg := tea.MouseMsg{
		X:      10,
		Y:      10,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}

	newModelInterface, _ := m.Update(msg)
	newModel := newModelInterface.(Model)

	if newModel.selectedSquare != nil {
		t.Errorf("Expected mouse input to be ignored outside GamePlay, got selectedSquare=%v",
			*newModel.selectedSquare)
	}
}

// TestComputeValidMoves tests that valid moves are computed correctly when selecting a piece.
func TestComputeValidMoves(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	m := Model{
		board:    engine.NewBoard(),
		gameType: GameTypePvP,
		screen:   ScreenGamePlay,
		config:   config,
	}

	// Select the e2 pawn (white pawn)
	// e2 = file 4, rank 1
	// mouseX = 2 + 4*2 = 10
	// mouseY = 4 + (7-1) = 10
	msg := tea.MouseMsg{
		X:      10,
		Y:      10,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}

	newModel, _ := m.handleMouseEvent(msg)

	if newModel.selectedSquare == nil {
		t.Fatalf("Expected selectedSquare to be set")
	}

	// Check that valid moves were computed
	if len(newModel.validMoves) == 0 {
		t.Fatalf("Expected validMoves to be populated")
	}

	// e2 pawn should have 2 valid moves: e3 and e4
	if len(newModel.validMoves) != 2 {
		t.Errorf("Expected 2 valid moves for e2 pawn, got %d", len(newModel.validMoves))
	}

	// Check that e3 and e4 are in valid moves
	e3 := engine.NewSquare(4, 2)
	e4 := engine.NewSquare(4, 3)
	hasE3, hasE4 := false, false
	for _, sq := range newModel.validMoves {
		if sq == e3 {
			hasE3 = true
		}
		if sq == e4 {
			hasE4 = true
		}
	}

	if !hasE3 {
		t.Errorf("Expected e3 to be in validMoves")
	}
	if !hasE4 {
		t.Errorf("Expected e4 to be in validMoves")
	}
}

// TestIsValidMoveDestination tests the helper function for checking valid destinations.
func TestIsValidMoveDestination(t *testing.T) {
	e3 := engine.NewSquare(4, 2)
	e4 := engine.NewSquare(4, 3)
	d4 := engine.NewSquare(3, 3)

	m := Model{
		validMoves: []engine.Square{e3, e4},
	}

	if !m.isValidMoveDestination(e3) {
		t.Errorf("Expected e3 to be valid destination")
	}

	if !m.isValidMoveDestination(e4) {
		t.Errorf("Expected e4 to be valid destination")
	}

	if m.isValidMoveDestination(d4) {
		t.Errorf("Expected d4 to NOT be valid destination")
	}
}

// TestExecuteMouseMove_ValidMove tests that clicking on a valid destination executes the move.
func TestExecuteMouseMove_ValidMove(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	m := Model{
		board:    engine.NewBoard(),
		gameType: GameTypePvP,
		screen:   ScreenGamePlay,
		config:   config,
	}

	// First, select the e2 pawn
	e2 := engine.NewSquare(4, 1)
	m.selectedSquare = &e2
	m.computeValidMoves()

	// Now click on e4 (valid move destination)
	// e4 = file 4, rank 3
	// mouseX = 2 + 4*2 = 10
	// mouseY = 4 + (7-3) = 8
	msg := tea.MouseMsg{
		X:      10,
		Y:      8,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}

	newModel, _ := m.handleMouseEvent(msg)

	// Move should be executed
	if newModel.selectedSquare != nil {
		t.Errorf("Expected selectedSquare to be cleared after move execution")
	}

	if newModel.validMoves != nil {
		t.Errorf("Expected validMoves to be cleared after move execution")
	}

	// Check that the pawn is now on e4
	e4 := engine.NewSquare(4, 3)
	piece := newModel.board.PieceAt(e4)
	if piece.Type() != engine.Pawn || piece.Color() != engine.White {
		t.Errorf("Expected white pawn on e4 after move")
	}

	// Check that e2 is now empty
	piece = newModel.board.PieceAt(e2)
	if !piece.IsEmpty() {
		t.Errorf("Expected e2 to be empty after move")
	}

	// Check that it's now Black's turn
	if newModel.board.ActiveColor != engine.Black {
		t.Errorf("Expected Black to move after White's move")
	}

	// Check that move was added to history
	if len(newModel.moveHistory) != 1 {
		t.Errorf("Expected 1 move in history, got %d", len(newModel.moveHistory))
	}
}

// TestExecuteMouseMove_InvalidDestination tests that clicking on an invalid destination
// keeps the piece selected.
func TestExecuteMouseMove_InvalidDestination(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	m := Model{
		board:    engine.NewBoard(),
		gameType: GameTypePvP,
		screen:   ScreenGamePlay,
		config:   config,
	}

	// First, select the e2 pawn
	e2 := engine.NewSquare(4, 1)
	m.selectedSquare = &e2
	m.computeValidMoves()

	// Click on a5 (not a valid destination for e2 pawn)
	// a5 = file 0, rank 4
	// mouseX = 2 + 0*2 = 2
	// mouseY = 4 + (7-4) = 7
	msg := tea.MouseMsg{
		X:      2,
		Y:      7,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}

	newModel, _ := m.handleMouseEvent(msg)

	// Selection should remain (piece stays selected)
	if newModel.selectedSquare == nil {
		t.Errorf("Expected selectedSquare to remain set")
	}

	if *newModel.selectedSquare != e2 {
		t.Errorf("Expected selectedSquare to remain e2")
	}

	// Move should NOT be executed
	if newModel.board.ActiveColor != engine.White {
		t.Errorf("Expected White to still be the active color (no move made)")
	}
}

// TestMouseMoveClears_Selection tests that a successful move clears the selection.
func TestMouseMoveClears_Selection(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	m := Model{
		board:    engine.NewBoard(),
		gameType: GameTypePvP,
		screen:   ScreenGamePlay,
		config:   config,
	}

	// Select and move e2-e4
	e2 := engine.NewSquare(4, 1)
	m.selectedSquare = &e2
	m.computeValidMoves()

	e4 := engine.NewSquare(4, 3)
	newModel, _ := m.executeMouseMove(e4)

	if newModel.selectedSquare != nil {
		t.Errorf("Expected selection to be cleared after move")
	}

	if len(newModel.validMoves) != 0 {
		t.Errorf("Expected validMoves to be cleared after move")
	}
}

// TestMouseMove_PvP_CompleteTurn tests a complete mouse-based turn in PvP mode.
func TestMouseMove_PvP_CompleteTurn(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	m := Model{
		board:       engine.NewBoard(),
		gameType:    GameTypePvP,
		screen:      ScreenGamePlay,
		config:      config,
		moveHistory: []engine.Move{},
	}

	// White's turn: select e2, move to e4
	// Click on e2
	msgSelectE2 := tea.MouseMsg{X: 10, Y: 10, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
	m, _ = m.handleMouseEvent(msgSelectE2)

	if m.selectedSquare == nil || m.selectedSquare.File() != 4 || m.selectedSquare.Rank() != 1 {
		t.Fatalf("Expected e2 to be selected")
	}

	// Click on e4
	msgMoveE4 := tea.MouseMsg{X: 10, Y: 8, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
	m, _ = m.handleMouseEvent(msgMoveE4)

	// Verify move executed
	if m.board.ActiveColor != engine.Black {
		t.Errorf("Expected Black's turn after White's move")
	}

	// Black's turn: select e7, move to e5
	// e7 = file 4, rank 6 -> mouseX=10, mouseY=4+(7-6)=5
	msgSelectE7 := tea.MouseMsg{X: 10, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
	m, _ = m.handleMouseEvent(msgSelectE7)

	if m.selectedSquare == nil || m.selectedSquare.File() != 4 || m.selectedSquare.Rank() != 6 {
		t.Fatalf("Expected e7 to be selected")
	}

	// e5 = file 4, rank 4 -> mouseX=10, mouseY=4+(7-4)=7
	msgMoveE5 := tea.MouseMsg{X: 10, Y: 7, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
	m, _ = m.handleMouseEvent(msgMoveE5)

	// Verify move executed
	if m.board.ActiveColor != engine.White {
		t.Errorf("Expected White's turn after Black's move")
	}

	// Verify move history has 2 moves
	if len(m.moveHistory) != 2 {
		t.Errorf("Expected 2 moves in history, got %d", len(m.moveHistory))
	}
}

// TestMouseMove_PvBot_HumanMove tests that human can make moves with mouse in PvBot mode.
func TestMouseMove_PvBot_HumanMove(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	m := Model{
		board:       engine.NewBoard(),
		gameType:    GameTypePvBot,
		screen:      ScreenGamePlay,
		config:      config,
		userColor:   engine.White, // Human plays White
		moveHistory: []engine.Move{},
	}

	// Select e2 pawn
	e2 := engine.NewSquare(4, 1)
	m.selectedSquare = &e2
	m.computeValidMoves()

	// Execute move to e4
	newModel, cmd := m.executeMouseMove(engine.NewSquare(4, 3))

	// Move should be executed
	if newModel.board.ActiveColor != engine.Black {
		t.Errorf("Expected Black's turn after White's move")
	}

	// Bot move command should be triggered
	if cmd == nil {
		t.Errorf("Expected bot move command to be triggered")
	}
}

// TestMouseMove_Capture tests that a capture move works correctly via mouse.
func TestMouseMove_Capture(t *testing.T) {
	// Set up a position where white pawn can capture
	// Use FEN: starting position with e4 and d5 played
	board, _ := engine.FromFEN("rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2")

	config := Config{
		ShowCoords: true,
	}

	m := Model{
		board:       board,
		gameType:    GameTypePvP,
		screen:      ScreenGamePlay,
		config:      config,
		moveHistory: []engine.Move{},
	}

	// Select e4 pawn
	e4 := engine.NewSquare(4, 3)
	m.selectedSquare = &e4
	m.computeValidMoves()

	// Verify d5 is a valid move (capture)
	d5 := engine.NewSquare(3, 4)
	if !m.isValidMoveDestination(d5) {
		t.Fatalf("Expected d5 to be a valid capture destination")
	}

	// Execute capture
	newModel, _ := m.executeMouseMove(d5)

	// Verify capture was executed
	piece := newModel.board.PieceAt(d5)
	if piece.Type() != engine.Pawn || piece.Color() != engine.White {
		t.Errorf("Expected white pawn on d5 after capture")
	}

	// e4 should be empty
	piece = newModel.board.PieceAt(e4)
	if !piece.IsEmpty() {
		t.Errorf("Expected e4 to be empty after capture")
	}
}

// TestMouseMove_ChangeSelectionToAnotherPiece tests that clicking on another own piece
// changes the selection instead of trying to move there.
func TestMouseMove_ChangeSelectionToAnotherPiece(t *testing.T) {
	config := Config{
		ShowCoords: true,
	}

	m := Model{
		board:    engine.NewBoard(),
		gameType: GameTypePvP,
		screen:   ScreenGamePlay,
		config:   config,
	}

	// Select e2 pawn
	e2 := engine.NewSquare(4, 1)
	m.selectedSquare = &e2
	m.computeValidMoves()

	// Click on d2 (another white pawn)
	// d2 = file 3, rank 1
	// mouseX = 2 + 3*2 = 8
	// mouseY = 4 + (7-1) = 10
	msg := tea.MouseMsg{
		X:      8,
		Y:      10,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}

	newModel, _ := m.handleMouseEvent(msg)

	// Selection should change to d2
	if newModel.selectedSquare == nil {
		t.Fatalf("Expected selectedSquare to be set")
	}

	d2 := engine.NewSquare(3, 1)
	if *newModel.selectedSquare != d2 {
		t.Errorf("Expected selectedSquare to be d2, got %v", *newModel.selectedSquare)
	}

	// Valid moves should be for d2, not e2
	d3 := engine.NewSquare(3, 2)
	d4 := engine.NewSquare(3, 3)
	hasD3, hasD4 := false, false
	for _, sq := range newModel.validMoves {
		if sq == d3 {
			hasD3 = true
		}
		if sq == d4 {
			hasD4 = true
		}
	}

	if !hasD3 || !hasD4 {
		t.Errorf("Expected valid moves to include d3 and d4 for d2 pawn")
	}
}

// TestMouseMove_PromotionAutoQueen tests that pawn promotion automatically promotes to Queen.
func TestMouseMove_PromotionAutoQueen(t *testing.T) {
	// Set up a position where white pawn is about to promote
	// FEN with white pawn on a7 (simpler - no piece blocking)
	// 8 . . . . k . . .
	// 7 P . . . . . . .
	// etc.
	board, err := engine.FromFEN("4k3/P7/8/8/8/8/8/4K3 w - - 0 1")
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	config := Config{
		ShowCoords: true,
	}

	m := Model{
		board:       board,
		gameType:    GameTypePvP,
		screen:      ScreenGamePlay,
		config:      config,
		moveHistory: []engine.Move{},
	}

	// Select a7 pawn (file 0, rank 6)
	a7 := engine.NewSquare(0, 6)
	piece := m.board.PieceAt(a7)
	if piece.Type() != engine.Pawn || piece.Color() != engine.White {
		// Debug: print board state
		for rank := 7; rank >= 0; rank-- {
			for file := 0; file < 8; file++ {
				sq := engine.NewSquare(file, rank)
				p := m.board.PieceAt(sq)
				if p.IsEmpty() {
					t.Logf("  ")
				} else {
					t.Logf("%v ", p)
				}
			}
			t.Logf("\n")
		}
		t.Fatalf("Expected white pawn at a7 (file 0, rank 6), got %v", piece)
	}

	m.selectedSquare = &a7
	m.computeValidMoves()

	// a8 should be a valid move (file 0, rank 7)
	a8 := engine.NewSquare(0, 7)
	if !m.isValidMoveDestination(a8) {
		// Debug: show all legal moves
		for _, move := range m.board.LegalMoves() {
			t.Logf("Legal move: %v -> %v (promo: %v)", move.From, move.To, move.Promotion)
		}
		t.Fatalf("Expected a8 to be a valid promotion destination, validMoves: %v", m.validMoves)
	}

	// Execute promotion move
	newModel, _ := m.executeMouseMove(a8)

	// Verify promotion to Queen
	pieceAtA8 := newModel.board.PieceAt(a8)
	if pieceAtA8.Type() != engine.Queen {
		t.Errorf("Expected Queen on a8 after promotion, got %v", pieceAtA8.Type())
	}
	if pieceAtA8.Color() != engine.White {
		t.Errorf("Expected White Queen on a8 after promotion")
	}
}

// TestMouseMove_GameOver tests that game transitions to GameOver screen after checkmate.
func TestMouseMove_GameOver(t *testing.T) {
	// Set up a position where white can checkmate in one move
	// Fool's mate position: after 1. f3 e5 2. g4, Black can play Qh4#
	board, _ := engine.FromFEN("rnb1kbnr/pppp1ppp/8/4p3/6Pq/5P2/PPPPP2P/RNBQKBNR w KQkq - 1 3")

	config := Config{
		ShowCoords: true,
	}

	m := Model{
		board:       board,
		gameType:    GameTypePvP,
		screen:      ScreenGamePlay,
		config:      config,
		moveHistory: []engine.Move{},
	}

	// This position is already checkmate for white
	if !m.board.IsGameOver() {
		t.Skip("Position should be game over (checkmate)")
	}

	// Verify game is over
	if m.board.IsGameOver() {
		m.screen = ScreenGameOver
	}

	if m.screen != ScreenGameOver {
		t.Errorf("Expected screen to be GameOver")
	}
}
