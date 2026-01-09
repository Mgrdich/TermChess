# Manual Testing Guide for Slice 4

## Overview
This guide helps verify that Slice 4 (Parse and Execute Simple Pawn Moves) is working correctly.

## How to Run
```bash
./termchess
```

## Test Cases

### Test 1: Valid Pawn Move (e2e4)
1. Run the application
2. Select "New Game" from the main menu
3. Type: `e2e4`
4. Press Enter

**Expected:**
- Input field clears
- White pawn moves from e2 to e4
- Turn indicator changes to "Black to move"
- No error message displayed

### Test 2: Valid Black Response (e7e5)
1. Continue from Test 1
2. Type: `e7e5`
3. Press Enter

**Expected:**
- Input field clears
- Black pawn moves from e7 to e5
- Turn indicator changes to "White to move"
- No error message displayed

### Test 3: Invalid Move Format
1. Continue from Test 2
2. Type: `invalid`
3. Press Enter

**Expected:**
- Error message appears: "invalid move format: expected 4-5 characters"
- Board state unchanged
- Still White to move

### Test 4: Illegal Move (e2e5)
1. Start a new game
2. Type: `e2e5`
3. Press Enter

**Expected:**
- Error message appears: "illegal move: e2e5"
- Board state unchanged
- Still White to move

### Test 5: Error Clearing
1. Continue from Test 4 (error message visible)
2. Start typing a new move (e.g., type 'e')

**Expected:**
- Error message disappears immediately
- Input shows the character typed

### Test 6: Backspace Handling
1. Type: `e2e4`
2. Press Backspace once

**Expected:**
- Input shows: `e2e`
- Can continue typing or backspace more

### Test 7: Knight Move (g1f3)
1. Start a new game
2. Move white pawn: `e2e4` + Enter
3. Move black pawn: `e7e5` + Enter
4. Move white knight: `g1f3` + Enter

**Expected:**
- White knight moves from g1 to f3
- Board updates correctly
- Turn indicator shows "Black to move"

### Test 8: Sequence of Valid Moves
Enter these moves in sequence:
1. `e2e4`
2. `e7e5`
3. `g1f3`
4. `b8c6`
5. `f1b5`

**Expected:**
- All moves execute successfully
- Board shows the Spanish Opening position
- Turn indicator shows "Black to move"

## Verification Checklist

- [ ] Can enter moves using coordinate notation
- [ ] Valid moves execute and update the board
- [ ] Invalid move formats show error messages
- [ ] Illegal moves show error messages
- [ ] Turn indicator updates correctly after each move
- [ ] Error messages clear when typing new input
- [ ] Backspace removes characters from input
- [ ] Input clears after successful move
- [ ] Board display shows all pieces correctly
- [ ] Can quit with 'q' key

## Known Limitations (To Be Implemented Later)

- No checkmate/stalemate detection (future slice)
- No move history display (future slice)
- No move validation highlighting (future slice)
- No undo/redo functionality (future slice)
