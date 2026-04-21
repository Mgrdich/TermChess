# internal/engine

Pure Go chess rules engine. Zero dependencies on other `internal/` packages — this is the domain core.

## Responsibilities

- Board state (`board.go`) and piece/color/square types (`types.go`)
- Move generation + legality filtering (`moves.go`, `attacks.go`)
- Game state detection: check, checkmate, stalemate, draws by rule (`game_state.go`)
- FEN import/export (`fen.go`)
- Zobrist position hashing for repetition detection (`zobrist.go`)

## Testing

`go test ./internal/engine/...` — comprehensive unit tests plus `perft_test.go` for move-generation correctness against known node counts.
