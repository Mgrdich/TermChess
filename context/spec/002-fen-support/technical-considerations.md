# Technical Specification: FEN Support

- **Functional Specification:** `context/spec/002-fen-support/functional-spec.md`
- **Status:** Approved
- **Author(s):** Claude (Tech Architect)

---

## 1. High-Level Technical Approach

FEN support will be implemented in `internal/engine/fen.go` as part of the chess engine package. The implementation provides two core functions: `ToFEN()` for exporting board state to a FEN string, and `ParseFEN()` for importing a FEN string into a new board. Clipboard functionality will be handled separately in a utility package.

**Files:**
- `internal/engine/fen.go` — FEN parsing and export logic
- `internal/engine/fen_test.go` — Comprehensive unit tests
- `internal/util/clipboard.go` — Clipboard helper (UI integration)

---

## 2. Proposed Solution & Implementation Plan

### 2.1 API Design

**FEN Export:**
```go
// ToFEN returns the FEN string representation of the current board position.
func (b *Board) ToFEN() string
```

**FEN Import:**
```go
// ParseFEN parses a FEN string and returns a new Board.
// Returns an error if the FEN string is invalid.
func ParseFEN(fen string) (*Board, error)
```

**Internal Helpers:**
```go
// pieceToChar converts a Piece to its FEN character representation.
func pieceToChar(p Piece) rune

// charToPiece converts a FEN character to a Piece.
func charToPiece(c rune) (Piece, error)
```

### 2.2 FEN Format

Standard FEN has 6 space-separated fields:
```
rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1
│                                            │ │    │ │ │
│                                            │ │    │ │ └─ Fullmove number
│                                            │ │    │ └─── Halfmove clock
│                                            │ │    └───── En passant square
│                                            │ └────────── Castling rights
│                                            └──────────── Active color
└───────────────────────────────────────────────────────── Piece placement
```

**Piece characters:**

| Piece  | White | Black |
|--------|-------|-------|
| King   | K     | k     |
| Queen  | Q     | q     |
| Rook   | R     | r     |
| Bishop | B     | b     |
| Knight | N     | n     |
| Pawn   | P     | p     |

### 2.3 Export Algorithm (`ToFEN`)

1. Iterate ranks 7→0 (8th rank first in FEN)
2. For each rank, iterate files 0→7
3. Count consecutive empty squares, write as digit (1-8)
4. Write piece characters for occupied squares
5. Separate ranks with `/`
6. Append active color (`w` or `b`)
7. Append castling rights (`KQkq`, subset, or `-`)
8. Append en passant square (algebraic or `-`)
9. Append halfmove clock and fullmove number

### 2.4 Import Algorithm (`ParseFEN`)

1. Split string by spaces, validate 6 fields
2. Parse piece placement:
   - Split by `/`, validate 8 ranks
   - For each rank, process characters left-to-right
   - Digits = skip that many files
   - Letters = place piece, advance file
   - Validate each rank sums to 8 squares
3. Parse active color: `w` → White, `b` → Black
4. Parse castling rights: check for `K`, `Q`, `k`, `q` or `-`
5. Parse en passant: algebraic notation or `-` for none
6. Parse halfmove clock: non-negative integer
7. Parse fullmove number: positive integer

### 2.5 Clipboard Support

Using `golang.design/x/clipboard` for cross-platform support:

```go
// internal/util/clipboard.go
package util

import "golang.design/x/clipboard"

func CopyToClipboard(text string) error {
    return clipboard.Write(clipboard.FmtText, []byte(text))
}
```

Fallback: If clipboard fails, FEN is still printed to screen.

---

## 3. Impact and Risk Analysis

### System Dependencies
- Depends on existing `Board` struct and types from `internal/engine/`
- UI layer will depend on these functions for `/fen` command and "Start from FEN" menu
- No external dependencies for core FEN logic; `golang.design/x/clipboard` for clipboard only

### Potential Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Invalid FEN crashes parsing | Comprehensive validation with early error returns |
| Clipboard unavailable on some systems | Graceful fallback: print to screen even if clipboard fails |
| Semantically invalid position (e.g., 3 kings) | Out of scope; accept structurally valid FEN |
| Edge cases in en passant/castling encoding | Dedicated test cases for all combinations |

---

## 4. Testing Strategy

**Unit Tests for Export (`ToFEN`):**
- Starting position → standard FEN string
- Empty board → `8/8/8/8/8/8/8/8 w KQkq - 0 1`
- Position with en passant square set
- Partial castling rights (e.g., `Kq`, `Q`, `-`)
- Various halfmove/fullmove values

**Unit Tests for Import (`ParseFEN`):**
- Parse standard starting position
- Parse positions with various piece layouts
- Parse all castling right combinations
- Parse en passant squares
- Parse various clock values

**Round-trip Tests:**
- Parse FEN → export → verify matches original
- Create board → export → parse → verify board matches

**Error Case Tests:**
- Wrong field count (5, 7 fields)
- Invalid piece characters (`X`, `1` in wrong place)
- Wrong rank count (7, 9 ranks)
- Rank doesn't sum to 8 squares
- Invalid active color (`x`, `W`)
- Invalid castling rights (`KQX`, `abc`)
- Invalid en passant square (`z9`, `a0`)
- Negative halfmove clock
- Zero/negative fullmove number
