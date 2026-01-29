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
