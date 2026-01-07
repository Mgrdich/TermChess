# Tasks: FEN Support

## Slice 1: FEN Export (ToFEN)
*Goal: Board can be exported to a valid FEN string.*

- [x] **Slice 1: Export board position to FEN string**
  - [x] Create `internal/engine/fen.go` with `pieceToChar(p Piece) rune` helper function
  - [x] Implement `(b *Board) ToFEN() string` method
  - [x] Add unit tests for `ToFEN()` with empty board
  - [x] Add unit tests for `ToFEN()` with standard starting position
  - [x] Add unit tests for `ToFEN()` with various castling rights combinations
  - [x] Add unit tests for `ToFEN()` with en passant square set
  - [x] Add unit tests for `ToFEN()` with various halfmove/fullmove values

---

## Slice 2: FEN Import (ParseFEN)
*Goal: Valid FEN strings can be parsed into a Board.*

- [x] **Slice 2: Parse FEN string into board position**
  - [x] Implement `charToPiece(c rune) (Piece, error)` helper function
  - [x] Implement `ParseFEN(fen string) (*Board, error)` function
  - [x] Add unit tests for parsing standard starting position
  - [x] Add unit tests for parsing positions with various piece layouts
  - [x] Add unit tests for parsing all castling right combinations
  - [x] Add unit tests for parsing en passant squares
  - [x] Add unit tests for parsing various clock values

---

## Slice 3: FEN Validation & Error Handling
*Goal: Invalid FEN strings are rejected with errors.*

- [x] **Slice 3: Validate FEN and return meaningful errors**
  - [x] Add validation for wrong field count (not 6 fields)
  - [x] Add validation for wrong rank count (not 8 ranks)
  - [x] Add validation for invalid piece characters
  - [x] Add validation for rank not summing to 8 squares
  - [x] Add validation for invalid active color
  - [x] Add validation for invalid castling rights
  - [x] Add validation for invalid en passant square
  - [x] Add validation for invalid halfmove clock (negative)
  - [x] Add validation for invalid fullmove number (zero/negative)
  - [x] Add unit tests for each error case

---

## Slice 4: Round-trip Tests
*Goal: Verify FEN export and import are consistent.*

- [x] **Slice 4: Round-trip verification tests**
  - [x] Add test: parse FEN → export → verify matches original
  - [x] Add test: create board manually → export → parse → verify board matches
  - [x] Add test: standard starting position round-trip
  - [x] Add test: complex position with en passant and partial castling round-trip

---

## Slice 5: Clipboard Utility
*Goal: FEN strings can be copied to system clipboard.*

- [ ] **Slice 5: Clipboard support for FEN export**
  - [ ] Add `golang.design/x/clipboard` dependency to `go.mod`
  - [ ] Create `internal/util/clipboard.go` with `CopyToClipboard(text string) error`
  - [ ] Add graceful error handling (fallback if clipboard unavailable)
  - [ ] Add basic test for clipboard functionality
