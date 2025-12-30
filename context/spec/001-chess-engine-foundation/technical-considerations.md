# Technical Specification: Chess Engine Foundation

- **Functional Specification:** `context/spec/001-chess-engine-foundation/functional-spec.md`
- **Status:** Approved
- **Author(s):** Claude (Tech Architect)

---

## 1. High-Level Technical Approach

The Chess Engine Foundation will be implemented as a pure Go package in `internal/engine/`. It provides board state management, legal move generation, and game state detection. The engine uses a struct-based board representation with a flat 64-square array and Zobrist hashing for repetition detection. Move generation follows a "generate pseudo-legal, then filter" approach for correctness.

**Files:**
- `board.go` — Board representation, initialization, state queries
- `moves.go` — Move struct, parsing, generation, validation
- `rules.go` — Game state detection (checkmate, stalemate, draws)

---

## 2. Proposed Solution & Implementation Plan

### 2.1 Data Model

**Piece Encoding:**
```go
type Color uint8
const (White Color = 0; Black Color = 1)

type PieceType uint8
const (Empty PieceType = 0; Pawn = 1; Knight = 2; Bishop = 3; Rook = 4; Queen = 5; King = 6)

type Piece uint8  // Color in high bit, PieceType in low 3 bits
```

**Board State:**
```go
type Board struct {
    Squares        [64]Piece   // a1=0, b1=1, ..., h8=63
    ActiveColor    Color
    CastlingRights uint8       // KQkq bits
    EnPassantSq    int8        // -1 if none
    HalfMoveClock  uint8
    FullMoveNum    uint16
    Hash           uint64      // Zobrist hash of current position
    History        []uint64    // History of hashes for repetition
}
```

**Square indexing:** `rank * 8 + file` where a1 is rank=0, file=0.

### 2.2 API Design

**Board operations:**
```go
func NewBoard() *Board                    // Standard starting position
func NewBoardFromPosition(pieces [64]Piece, opts BoardOptions) *Board
func (b *Board) PieceAt(sq Square) Piece
func (b *Board) Copy() *Board
```

**Move operations:**
```go
type Move struct { From, To Square; Promotion PieceType }

func ParseMove(s string) (Move, error)    // "e2e4", "e7e8q"
func (m Move) String() string

func (b *Board) LegalMoves() []Move
func (b *Board) MakeMove(m Move) error    // Validates and applies
func (b *Board) IsLegalMove(m Move) bool
```

**ParseMove** converts a coordinate notation string into a Move struct:
- `ParseMove("e2e4")` → `Move{From: 12, To: 28, Promotion: 0}` (e2=index 12, e4=index 28)
- `ParseMove("e7e8q")` → `Move{From: 52, To: 60, Promotion: Queen}` (pawn promotes to queen)
- `ParseMove("xyz")` → error (invalid format)

**Game state:**
```go
type GameStatus int
// Constants: Ongoing, Checkmate, Stalemate, Draw* variants

func (b *Board) Status() GameStatus
func (b *Board) InCheck() bool
```

### 2.3 Algorithm: Move Generation

1. For each piece of active color, generate pseudo-legal moves (ignoring pins/checks)
2. For each candidate move:
   a. Make move on board copy
   b. Check if own king is in check
   c. If not in check, move is legal
3. Return list of legal moves

### 2.4 Algorithm: Game State Detection

After each move, check in order:
1. **Legal moves = 0?**
   - If in check → Checkmate
   - If not in check → Stalemate
2. **Insufficient material?** → Draw
3. **Fivefold repetition?** → Automatic draw
4. **75-move rule?** → Automatic draw
5. **Threefold repetition?** → Draw
6. **50-move rule?** → Draw
7. Otherwise → Ongoing

### 2.5 Zobrist Hashing

Pre-computed random values for:
- Each piece type on each square (12 x 64 values)
- Side to move (1 value)
- Castling rights (16 values)
- En passant file (8 values)

Hash updated incrementally on each move via XOR.

---

## 3. Impact and Risk Analysis

### System Dependencies
- This is the foundational component — no dependencies on other internal packages
- Future components (UI, bots, FEN support) will depend on this engine

### Potential Risks & Mitigations

| Risk                        | Mitigation                                                       |
|-----------------------------|------------------------------------------------------------------|
| Move generation bugs        | Perft testing against known results                              |
| Edge cases in special rules | Dedicated test cases for castling, en passant, promotion         |
| Repetition detection errors | Verify Zobrist implementation with known positions               |
| Performance bottlenecks     | Profile if needed; pseudo-legal filter is adequate for TUI chess |

---

## 4. Testing Strategy

**Unit Tests:**
- Board initialization (standard + custom positions)
- Piece movement for each piece type
- Special moves: castling (all 4 variations + blocking scenarios), en passant, promotion
- Check detection
- Game state detection (checkmate, stalemate, all draw types)

**Perft Tests:**
- Starting position: depth 1-4 (20, 400, 8902, 197281 nodes)
- Known perft positions with established results
- These comprehensively verify move generation correctness

**Edge Case Tests:**
- Castling through check (illegal)
- Castling after rook/king moved (illegal)
- En passant pin (pawn is pinned to king)
- Discovered check
- Double check
- Underpromotion
