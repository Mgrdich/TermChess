# internal/ui

Terminal UI built with [Bubbletea](https://github.com/charmbracelet/bubbletea). Single Go package, MVU (model-view-update) architecture, covers all screens.

See `CLAUDE.md` in this directory for AI-assistant context (screen state machine, conventions, gotchas).

## Files

- `model.go` — application state (`Model` struct)
- `view.go` — rendering for every screen (2798 lines — split pending)
- `update.go` — event handling for every screen (2290 lines — split pending)
- `board.go` — board rendering (ASCII/Unicode)
- `san.go` — SAN move parsing
- `save.go` — save/resume integration with `internal/config`

Run with `make run` or `go run ./cmd/termchess`.
