# Tasks: FEN Support

## Slice 1: FEN Export (ToFEN)
*Goal: Board can be exported to a valid FEN string.*

- [ ] **Slice 1: Export board position to FEN string**
  - [ ] Create `internal/engine/fen.go` with `pieceToChar(p Piece) rune` helper function
  - [ ] Implement `(b *Board) ToFEN() string` method
  - [ ] Add unit tests for `ToFEN()` with empty board
  - [ ] Add unit tests for `ToFEN()` with standard starting position
  - [ ] Add unit tests for `ToFEN()` with various castling rights combinations
  - [ ] Add unit tests for `ToFEN()` with en passant square set
  - [ ] Add unit tests for `ToFEN()` with various halfmove/fullmove values

---

## Slice 2: FEN Import (ParseFEN)
*Goal: Valid FEN strings can be parsed into a Board.*

- [ ] **Slice 2: Parse FEN string into board position**
  - [ ] Implement `charToPiece(c rune) (Piece, error)` helper function
  - [ ] Implement `ParseFEN(fen string) (*Board, error)` function
  - [ ] Add unit tests for parsing standard starting position
  - [ ] Add unit tests for parsing positions with various piece layouts
  - [ ] Add unit tests for parsing all castling right combinations
  - [ ] Add unit tests for parsing en passant squares
  - [ ] Add unit tests for parsing various clock values

---

## Slice 3: FEN Validation & Error Handling
*Goal: Invalid FEN strings are rejected with errors.*

- [ ] **Slice 3: Validate FEN and return meaningful errors**
  - [ ] Add validation for wrong field count (not 6 fields)
  - [ ] Add validation for wrong rank count (not 8 ranks)
  - [ ] Add validation for invalid piece characters
  - [ ] Add validation for rank not summing to 8 squares
  - [ ] Add validation for invalid active color
  - [ ] Add validation for invalid castling rights
  - [ ] Add validation for invalid en passant square
  - [ ] Add validation for invalid halfmove clock (negative)
  - [ ] Add validation for invalid fullmove number (zero/negative)
  - [ ] Add unit tests for each error case

---

## Slice 4: Round-trip Tests
*Goal: Verify FEN export and import are consistent.*

- [ ] **Slice 4: Round-trip verification tests**
  - [ ] Add test: parse FEN → export → verify matches original
  - [ ] Add test: create board manually → export → parse → verify board matches
  - [ ] Add test: standard starting position round-trip
  - [ ] Add test: complex position with en passant and partial castling round-trip

---

## Slice 5: Clipboard Utility
*Goal: FEN strings can be copied to system clipboard.*

- [ ] **Slice 5: Clipboard support for FEN export**
  - [ ] Add `golang.design/x/clipboard` dependency to `go.mod`
  - [ ] Create `internal/util/clipboard.go` with `CopyToClipboard(text string) error`
  - [ ] Add graceful error handling (fallback if clipboard unavailable)
  - [ ] Add basic test for clipboard functionality
