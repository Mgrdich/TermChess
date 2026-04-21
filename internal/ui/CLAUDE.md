# internal/ui — AI Context

Bubbletea MVU (Model-View-Update) for the terminal UI. This is a single Go package with 31 files covering every screen in the app.

## Structure

- `model.go` — application state. The `Model` struct holds screen, board, menu selections, settings, bvb session, updater state, etc.
- `view.go` — rendering (2798 lines — spans every screen). Search by section comment: "Menu", "Game", "Settings", "BvB", "Updater".
- `update.go` — event handling (2290 lines — same screen structure as `view.go`).
- `board.go` — board rendering (ASCII/Unicode, colors, coordinates).
- `san.go` — SAN move parsing and disambiguation.
- `save.go` — save/resume game integration with `internal/config`.

## MVU loop

```
tea.Program → (Msg) → Model.Update(Msg) → (Model, Cmd) → Model.View() → terminal
```

Messages include key events, bot move completion, updater check results, timers. Commands include async bot moves, HTTP calls to GitHub releases, and timer ticks.

## Screen state machine

The `screen` field on `Model` enumerates the active view. Transitions happen in `update.go` via `m.screen = screenX`. When adding a new screen:

1. Add a `screenX` constant.
2. Add a `viewX()` function called from `view.go`'s top-level dispatcher.
3. Add an `updateX()` function called from `update.go`'s top-level dispatcher.

Because `view.go` and `update.go` are already 2000+ lines each, a new screen is a good moment to start splitting by screen (`view_menu.go`, `update_menu.go`, etc.) — see ARCH-06 in the 2026-04-21 audit.

## Conventions

- **Widget composition:** `bubbles/textinput`, `bubbles/viewport`, `bubbles/list` are used — prefer these over building input widgets from scratch.
- **Styling:** `lipgloss` for colors/borders. Respect `settings.UseColors` everywhere.
- **No direct engine mutation:** UI code creates new `engine.Board` values or calls engine methods; it never mutates internal engine state behind the engine's back.
- **Async work returns `tea.Cmd`:** bot moves, update checks, file I/O — never block in `Update()`.

## Gotchas

- **Two screens can conflict on the same key:** if you're adding a key binding, grep `update.go` for the keybinding first — e.g., `space` toggles pause in BvB but enters a move on the game screen.
- **Dragons in bot-move integration:** bot moves are dispatched via `tea.Cmd` and return as `botMoveMsg`. If you add a new bot type, wire it into the factory in `internal/bot/engine.go` and the message handler in `update.go` → "Game" section.
- **`view.go` uses string concatenation heavily.** If you're tempted to add a `fmt.Sprintf` with many `%s`, consider building a strings.Builder instead — view functions are called every frame.
