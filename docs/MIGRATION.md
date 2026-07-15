# TermChess: Go → Rust Migration

This is the **authoritative migration document** for the Go → Rust port of TermChess.
The Rust workspace lives under `rust/` and builds independently of the Go tree; both
implementations produce the same terminal chess TUI and are kept in parity.

- **Go build root:** `go.mod` at the repo root (packages under `cmd/`, `internal/`).
- **Rust build root:** `rust/Cargo.toml` — a Cargo workspace. Library crates live in
  `rust/crates/`, and the binary crate is `rust/app/` (package/binary name `termchess`).

The Rust migration is **complete and fully green**: the whole workspace builds and every
crate's test suite passes. No Go code was modified during the migration.

## Build & run

All commands are driven from the repo-root `Makefile`:

| Task  | Go              | Rust                                                     |
|-------|-----------------|----------------------------------------------------------|
| Build | `make build-go` | `make build-rust` — `cd rust && cargo build --release -p termchess` |
| Run   | `make run-go`   | `make run-rust`   — `cd rust && cargo run -p termchess`  |
| Test  | `make test`     | `make test-rust`  — `cd rust && cargo test`              |

`make build` (no suffix) is an alias for `make build-go`.

You can also drive Cargo directly from inside `rust/`:

```sh
cd rust
cargo build                 # build the whole workspace
cargo build -p termchess    # build just the binary crate
cargo test                  # run all crate test suites
cargo run -p termchess -- --version
cargo run -p termchess -- --help
cargo run -p termchess      # launch the interactive TUI (needs a real TTY)
```

### CLI flags (parity with `cmd/termchess/main.go`)

- `--version` / `-version` / `-v` — print version, build date, git commit, then exit.
- `--help` / `-h` / `-help` — print usage and exit.
- `--upgrade [version]` — download and swap the binary in place.
- `--uninstall` — remove the binary and config after confirmation.
- _no flags_ — launch the interactive TUI.

The `--help`/`-h` usage output and the `-v` short alias were added to the Rust binary to
close a CLI-parity gap: previously unknown flags fell through into the raw-mode TUI.

## Crate ↔ Go-package mapping

The migration preserves the Go layer map one-for-one. Each Rust crate mirrors exactly one
Go package, keeping the same one-way dependency flow (`engine` is a pure domain crate with
no internal dependencies).

| Rust crate (`rust/crates/…`)      | Go package         | Responsibility |
|-----------------------------------|--------------------|----------------|
| `engine`                          | `internal/engine`  | Chess rules, board, FEN, moves, attacks, Zobrist hashing (pure domain, no internal deps) |
| `bot`                             | `internal/bot`     | Engine trait + `random`, `minimax` (alpha-beta + eval), `rl`/ONNX skeleton |
| `bvb`                             | `internal/bvb`     | Bot-vs-bot session controller, stats, multi-game queue |
| `config`                          | `internal/config`  | TOML settings + FEN savegame under `~/.termchess/` |
| `updater`                         | `internal/updater` | GitHub releases check + in-place binary swap |
| `util`                            | `internal/util`    | Cross-platform clipboard helper |
| `version`                         | `internal/version` | Build-time-injected version metadata |
| `ui`                              | `internal/ui`      | Terminal UI; `ui::run(cfg)` is the entrypoint |
| `app` (binary `termchess`)        | `cmd/termchess`    | CLI entrypoint: flag parsing + launching the UI |

Dependency flow (unchanged from Go):

```
app (termchess) → ui → bvb, updater → bot, config → engine
```

Each crate's `src/lib.rs` (and `rust/app/src/main.rs`) carries a `//!` rustdoc header that
names the Go package it mirrors; run `cd rust && cargo doc --no-deps` to browse them.

## Architectural change: Bubbletea MVU → ratatui event loop

The single largest structural change is the UI layer. The Go UI is built on
[Bubbletea](https://github.com/charmbracelet/bubbletea), which uses the Elm-style
Model–View–Update (MVU) architecture with `tea.Cmd` for async effects. The Rust `ui` crate
ports this onto [ratatui](https://ratatui.rs) + crossterm:

| Bubbletea (Go)                         | ratatui (Rust `ui` crate)                              |
|----------------------------------------|--------------------------------------------------------|
| `Model` (immutable state, returned)    | `App` — owns all state, mutated in place               |
| `Update(msg) (Model, Cmd)`             | `App::update(event)` — mutates state in response to an `AppEvent` |
| `View() string`                        | `App::draw(frame)` — renders the current screen with ratatui widgets |
| `tea.Cmd` async effects                | worker threads + an `mpsc` channel feeding `AppEvent`s |
| `tea.Msg`                              | `AppEvent` enum (keyboard, mouse, bot-move, tick, …)   |
| `Screen` iota constants                | `Screen` Rust enum                                     |

All **17 UI screens** from Go's `Screen` enum are ported 1:1 (verified against
`internal/ui/model.go`): `MainMenu`, `GameTypeSelect`, `BotSelect`, `ColorSelect`,
`FenInput`, `GamePlay`, `GameOver`, `Settings`, `SavePrompt`, `DrawPrompt`, `BvBBotSelect`,
`BvBGameMode`, `BvBGridConfig`, `BvBGamePlay`, `BvBStats`, `BvBViewModeSelect`,
`BvBConcurrencySelect`. As in Go, the updater is CLI-driven rather than a dedicated screen.

### Idiom mapping

The port follows idiomatic Rust throughout:

- Go interfaces → Rust **traits** (e.g. `bot::Engine`, `Configurable`, `Stateful`, `Inspectable`).
- Go `error` / `if err != nil { return err }` → `Result<T, E>` propagated with `?`.
- Error types → `enum`s deriving [`thiserror::Error`] (e.g. `EngineError`, `FenError`, `UpdaterError`).
- Go structs + methods → `struct` + `impl`.
- Go `iota` enum constants → Rust `enum`s.
- Go `context.Context` → a `Context` type in the `bot` crate.
- Library code avoids `unwrap()` and propagates with `?`/`Result`.

## The board-encoder / ONNX invariant is now tri-language

TermChess's cross-language contract is the board encoder used by the RL bot. It was already
duplicated across Go and Python and must stay byte-identical; the Rust port adds a **third**
copy of the same tensor layout that must be kept in lockstep:

| Language | File                              | Output tensor         |
|----------|-----------------------------------|-----------------------|
| Python   | `training/board_encoder.py`       | `[batch, 18, 8, 8]` f32 |
| Go       | `internal/bot/rl_encoder.go`      | same layout           |
| Rust     | `rust/crates/bot/src/rl_encoder.rs` | same layout          |

**Invariant:** any change to channel layout, channel count, or encoding semantics must be
applied in **all three** files in the same commit, and all three test suites must pass:

- `training/test_board_encoder.py`
- `internal/bot/rl_encoder_test.go`
- the encoder tests in `rust/crates/bot/src/tests/rl_encoder_tests.rs`

The 66-channel / 18-channel debate is tracked in spec `008-custom-rl-agent`. The Python side
trains a PyTorch model and exports ONNX via `training/export_onnx.py`, which both the Go and
Rust bots are intended to consume.

## Current parity gaps

- **RL bot / ONNX not consumed at runtime.** As in the Go build, the RL engine is scaffolded
  but the ONNX model is not wired into runtime inference (Go's unresolved spec 008 Slice 11).
  On the Rust side, `RlEngine`'s `difficulty` and `time_limit` fields are currently unread,
  producing the workspace's single `dead_code` warning. This mirrors the Go build exactly and
  is not a regression.
- **Interactive TUI not yet visually validated against Go.** `ui::run` could not be exercised
  during the port because the migration environment had no TTY. Screen-by-screen visual and
  behavioral parity against the Go UI should be validated manually in a real terminal.

## Parity status (verification)

All commands below were run with `cargo` exit code 0:

- `cargo build` (full workspace) — 0
- `cargo build -p termchess` — 0
- `cargo test` — all 18 test-result lines `ok`, 0 failed:
  engine 93, bot 148 (+9 ignored), bvb 61, config 20, ui 30, updater 27 (+8), util 2, version 4.
- Non-interactive smoke tests: `cargo run -p termchess -- --version`, `-v`, and `--help` each
  print and exit 0 without entering the TUI.
- `make build-rust` produces a release binary successfully.

Green (build + tests pass): `engine`, `config`, `util`, `version`, `bot`, `bvb`, `updater`,
`ui`, `termchess` (app). Red: none.

## Follow-ups

- Wire ONNX runtime inference in both Go and Rust so `RlEngine` consumes trained models
  (removes the `difficulty`/`time_limit` dead-code warning); tracked as spec 008 Slice 11.
- Manually validate the interactive TUI screen-by-screen against the Go build in a real terminal.
- Commit the `rust/` tree on the `rust-migration` branch and review the `Makefile` diff before merge.

---

_This document supersedes the earlier `rust/README.md`, which now points here to avoid
duplication._
