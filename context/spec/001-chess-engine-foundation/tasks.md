# Tasks: Chess Engine Foundation

## Slice 1: Project Setup & Empty Board
*Goal: Runnable project with basic board structure that can be instantiated and queried.*

- [x] **Slice 1: Project scaffolding and empty board representation**
  - [x] Initialize Go module (`go mod init`) and create directory structure (`cmd/termchess/`, `internal/engine/`)
  - [x] Create `internal/engine/types.go`: Define `Color`, `PieceType`, `Piece`, `Square` types and constants
  - [x] Create `internal/engine/board.go`: Define `Board` struct with `Squares [64]Piece` and metadata fields
  - [x] Implement `NewBoard()` that returns an empty board (all squares empty)
  - [x] Implement `PieceAt(sq Square) Piece` query method
  - [x] Create `cmd/termchess/main.go` with minimal main that instantiates a board (smoke test)
  - [x] Add unit tests for board creation and `PieceAt` queries
  - [x] Set up Makefile with `build`, `test`, `run` targets

---

## Slice 2: Standard Starting Position
*Goal: Board can initialize to standard chess starting position.*

- [x] **Slice 2: Initialize board with standard starting position**
  - [x] Update `NewBoard()` to place all 32 pieces in standard starting positions
  - [x] Set initial metadata: White to move, all castling rights, no en passant, clocks at 0/1
  - [x] Implement `Board.String()` for debug printing (simple text representation)
  - [x] Add unit tests verifying all pieces are in correct positions
  - [x] Add unit tests verifying initial metadata values

---

## Slice 3: Basic Pawn Moves (No Special Rules)
*Goal: Generate and apply simple pawn moves (1 forward, 2 from start, diagonal capture).*

- [x] **Slice 3: Pawn move generation and application (basic)**
  - [x] Create `internal/engine/moves.go`: Define `Move` struct with `From`, `To`, `Promotion` fields
  - [x] Implement `ParseMove(s string) (Move, error)` for coordinate notation ("e2e4")
  - [x] Implement `Move.String()` to convert back to coordinate notation
  - [x] Implement `Board.Copy()` to clone board state
  - [x] Implement pawn pseudo-legal move generation (1 forward, 2 from start rank, diagonal captures)
  - [x] Implement `Board.MakeMove(m Move) error` that applies a move and updates state
  - [x] Add unit tests for pawn move generation from various positions
  - [x] Add unit tests for `ParseMove` and `Move.String()` round-trip

---

## Slice 4: Knight, Bishop, Rook, Queen Moves
*Goal: Generate moves for sliding and jumping pieces.*

- [x] **Slice 4: Non-pawn piece move generation**
  - [x] Implement knight move generation (L-shape, can jump)
  - [x] Implement bishop move generation (diagonal sliding, blocked by pieces)
  - [x] Implement rook move generation (orthogonal sliding, blocked by pieces)
  - [x] Implement queen move generation (combines bishop + rook)
  - [x] Implement king move generation (one square any direction, no castling yet)
  - [x] Implement `Board.PseudoLegalMoves()` that generates all pseudo-legal moves for active color
  - [x] Add unit tests for each piece type from various positions

---

## Slice 5: Check Detection & Legal Move Filtering
*Goal: Filter out moves that leave own king in check; detect check state.*

- [ ] **Slice 5: Check detection and legal move generation**
  - [x] Implement `Board.IsSquareAttacked(sq Square, byColor Color) bool`
  - [x] Implement `Board.InCheck() bool` (is active color's king attacked?)
  - [ ] Implement `Board.LegalMoves()` that filters pseudo-legal moves by checking if king is left in check
  - [ ] Implement `Board.IsLegalMove(m Move) bool` convenience method
  - [ ] Update `Board.MakeMove()` to reject illegal moves with error
  - [ ] Add unit tests for check detection in various positions
  - [ ] Add unit tests verifying illegal moves are filtered (moving pinned pieces, etc.)

---

## Slice 6: Castling
*Goal: Implement castling move generation and validation.*

- [ ] **Slice 6: Castling implementation**
  - [ ] Implement castling move generation (kingside and queenside for both colors)
  - [ ] Validate castling conditions: rights not lost, no pieces between, king not in check, king doesn't pass through check
  - [ ] Update `MakeMove()` to handle castling (move both king and rook)
  - [ ] Update castling rights when king or rook moves
  - [ ] Add unit tests for all 4 castling variations
  - [ ] Add unit tests for castling being blocked (pieces in way, through check, rights lost)

---

## Slice 7: En Passant
*Goal: Implement en passant capture.*

- [ ] **Slice 7: En passant implementation**
  - [ ] Update `MakeMove()` to set en passant square when pawn moves two squares
  - [ ] Update pawn move generation to include en passant captures when available
  - [ ] Update `MakeMove()` to handle en passant capture (remove captured pawn)
  - [ ] Clear en passant square after any move (only valid immediately)
  - [ ] Add unit tests for en passant capture
  - [ ] Add unit test for en passant expiring after one move

---

## Slice 8: Pawn Promotion
*Goal: Implement pawn promotion.*

- [ ] **Slice 8: Pawn promotion implementation**
  - [ ] Update pawn move generation to require promotion piece when reaching 8th rank
  - [ ] Update `MakeMove()` to replace pawn with promoted piece
  - [ ] Update `ParseMove()` to handle promotion suffix ("e7e8q")
  - [ ] Add unit tests for promotion to all 4 piece types (Q, R, B, N)
  - [ ] Add unit test that promotion is required (move without promotion piece fails)

---

## Slice 9: Checkmate & Stalemate Detection
*Goal: Detect game-ending conditions when no legal moves exist.*

- [ ] **Slice 9: Checkmate and stalemate detection**
  - [ ] Create `internal/engine/rules.go`: Define `GameStatus` enum (Ongoing, Checkmate, Stalemate, Draw variants)
  - [ ] Implement `Board.Status() GameStatus` that checks for checkmate/stalemate
  - [ ] Checkmate: in check + no legal moves
  - [ ] Stalemate: not in check + no legal moves
  - [ ] Add unit tests with known checkmate positions
  - [ ] Add unit tests with known stalemate positions

---

## Slice 10: Zobrist Hashing & Position History
*Goal: Implement incremental hashing for repetition detection.*

- [ ] **Slice 10: Zobrist hashing implementation**
  - [ ] Create `internal/engine/zobrist.go`: Generate random values for pieces, side, castling, en passant
  - [ ] Add `Hash uint64` field to `Board`
  - [ ] Add `History []uint64` field to track position hashes
  - [ ] Compute initial hash in `NewBoard()`
  - [ ] Update hash incrementally in `MakeMove()` via XOR
  - [ ] Add hash to history after each move
  - [ ] Add unit tests verifying hash changes appropriately
  - [ ] Add unit test verifying same position produces same hash

---

## Slice 11: Draw by Repetition
*Goal: Detect threefold and fivefold repetition.*

- [ ] **Slice 11: Repetition draw detection**
  - [ ] Implement threefold repetition check (count occurrences in history)
  - [ ] Implement fivefold repetition check (automatic draw)
  - [ ] Update `Board.Status()` to return appropriate draw status
  - [ ] Add unit tests with positions that repeat 3 times
  - [ ] Add unit tests with positions that repeat 5 times

---

## Slice 12: Draw by Move Rules (50/75)
*Goal: Detect draws from 50-move and 75-move rules.*

- [ ] **Slice 12: Move count draw detection**
  - [ ] Update `MakeMove()` to increment/reset half-move clock appropriately
  - [ ] Implement 50-move rule detection
  - [ ] Implement 75-move rule detection (automatic draw)
  - [ ] Update `Board.Status()` to return appropriate draw status
  - [ ] Add unit tests for 50-move and 75-move scenarios

---

## Slice 13: Insufficient Material Draw
*Goal: Detect draws from insufficient mating material.*

- [ ] **Slice 13: Insufficient material detection**
  - [ ] Implement material counting
  - [ ] Detect K vs K
  - [ ] Detect K+B vs K, K+N vs K
  - [ ] Detect K+B vs K+B (same color bishops)
  - [ ] Update `Board.Status()` to return `DrawInsufficientMaterial`
  - [ ] Add unit tests for each insufficient material scenario

---

## Slice 14: Perft Testing & Validation
*Goal: Validate move generation correctness with perft tests.*

- [ ] **Slice 14: Perft testing suite**
  - [ ] Implement `Perft(depth int) uint64` that counts leaf nodes
  - [ ] Add perft test for starting position: depth 1 (20), depth 2 (400), depth 3 (8902), depth 4 (197281)
  - [ ] Add perft tests for known "Kiwipete" and other tricky positions
  - [ ] Fix any bugs discovered by perft mismatches
