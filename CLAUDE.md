# TermChess — AI Context

Terminal chess TUI in Go with a Python RL training pipeline. Built via AWOS spec-driven development — see `context/spec/` and `context/product/`.

## Build roots (polyglot monorepo)

Three independent build roots. Switch toolchains based on what you're touching:

| Root | Use when | Commands |
|------|----------|----------|
| `go.mod` (root) | Editing Go code under `cmd/`, `internal/` | `make build-go`, `make test`, `make run-go` (`make build` aliases the Go build) |
| `rust/Cargo.toml` | Editing the Rust port under `rust/crates/`, `rust/app/` | `make build-rust`, `make test-rust`, `make run-rust`; or `cd rust && cargo build/test/run -p termchess` |
| `training/pyproject.toml` | Editing Python RL pipeline | `cd training && uv run pytest`, `cd training && uv run python -u train.py ...` |

The Go CI job runs `make build` + `make test`. Python and Rust tests are **not** in CI today — run them locally before pushing changes to `training/` or `rust/`. The Rust port lives entirely under `rust/`; see `docs/MIGRATION.md` for the crate ↔ Go-package map. **Never modify Go code from the Rust side** — the two trees are independent.

## Cross-language contract (ONNX)

Python trains a PyTorch model → exports ONNX (`training/export_onnx.py`) → Go consumes it for the RL bot.

**Critical invariant: the board encoder is duplicated across languages and must stay byte-identical:**

- `training/board_encoder.py` — Python encoder, produces `[batch, 18, 8, 8]` float32 tensor
- `internal/bot/rl_encoder.go` — Go encoder, produces the same tensor layout
- `rust/crates/bot/src/rl_encoder.rs` — Rust encoder, produces the same tensor layout

The invariant is now **tri-language**: any change to channel layout, channel count, or encoding semantics must be applied in all three files in the same commit, and all three test suites (`training/test_board_encoder.py` + `internal/bot/rl_encoder_test.go` + `rust/crates/bot/src/tests/rl_encoder_tests.rs`) must pass. The 66-channel / 18-channel debate is tracked in spec `008-custom-rl-agent`.

Go-side ONNX inference is wired (interface, session, selector, mock-session tests) but `newOnnxSession()` in `internal/bot/rl.go` currently returns `ErrModelNotLoaded` — the `onnxruntime_go` dependency is not yet in `go.mod`. Tracked as spec 008 Slice 11.

## Layer map (`internal/`)

One-way dependency flow; `engine` is a pure domain with zero internal deps.

```
cmd/termchess (entry) → internal/ui → internal/bvb, internal/updater → internal/bot, internal/config → internal/engine
```

- `engine/` — chess rules, board, FEN, moves, attacks, Zobrist hashing. **No imports from other `internal/` packages.**
- `bot/` — engine interface + implementations: `random.go`, `minimax.go` (alpha-beta + `eval.go`), `rl.go` (RL/ONNX).
- `bvb/` — bot-vs-bot session controller, stats, multi-game queue.
- `ui/` — Bubbletea MVU (model/view/update). Screens: menu, game, settings, bvb, updater. Two files are monolithic (`view.go` 2798 lines, `update.go` 2290 lines) — consider splitting by screen before adding new screens.
- `config/` — TOML settings at `~/.termchess/config.toml`; FEN savegame at `~/.termchess/savegame.fen`.
- `updater/` — GitHub releases check and in-place binary swap (`termchess --upgrade`).
- `version/` — build-time injected version metadata (via LDFLAGS in Makefile).
- `util/` — cross-platform clipboard helper only.

## Conventions

- **Go naming:** snake_case filenames, tests colocated as `*_test.go`, package names match directory names.
- **Python naming:** snake_case filenames, tests as `test_*.py`.
- **Error handling:** idiomatic `if err != nil { return err }`; no silent swallowing. Python uses typed `except` blocks with logging; no bare `except:`.
- **Linting:** Go uses `.golangci.yml` with 8 linters. Python has no linter configured yet.
- **Test coverage:** 44 Go test files + 6 Python test files. UI layer has ~83% coverage. Run `go test -v ./...` or `cd training && uv run pytest`.

## Spec workflow (AWOS)

Significant features go through: `/awos:spec` → `/awos:tech` → `/awos:tasks` → `/awos:implement` → `/awos:verify`. When touching feature work, check `context/spec/<slice>/tasks.md` for open items. Commit messages should reference the slice ID (e.g., "008/Slice 11: …").

## Gotchas

- **Large UI files:** `internal/ui/view.go` and `internal/ui/update.go` handle every screen in a single file. Locate your screen by section comment, not by filename.
- **RL bot returns `ErrModelNotLoaded`:** The RL engine is scaffolded but the ONNX model is not yet consumed at runtime. The Go encoder, selector, and tests exist; only the runtime dep is missing.
- **Training checkpoints are gitignored** (`training/checkpoints/*`). `training_log.csv` is written there during runs; the `/training-health` slash command analyzes it.
- **Config path is OS-dependent:** `internal/config/paths.go` resolves `~/.termchess/` per platform.
- **Go and Rust CLIs are kept in parity:** the Rust binary (`rust/app/src/main.rs`) mirrors `cmd/termchess/main.go` flag-for-flag (`--version`/`-v`, `--help`/`-h`, `--upgrade`, `--uninstall`). Any CLI, screen, or on-disk-format change must be applied to both. The board-encoder invariant is now tri-language (Python + Go + Rust) — see the ONNX section above.
