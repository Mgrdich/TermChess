# Code Architecture — Audit Results

**Date:** 2026-04-21
**Score:** 89% — Grade **B**

## Results

| # | Check | Severity | Status | Evidence |
| - | ----- | -------- | ------ | -------- |
| 1 | ARCH-01 Pattern declared/recognizable | high | PASS | Layer map declared in `CLAUDE.md` (cmd -> ui -> bvb/updater -> bot/config -> engine). Verified: 8 Go packages (engine, bot, bvb, ui, config, updater, util, version) match declaration. Python `training/` is modular functional pipeline (board_encoder, model, mcts, self_play, replay_buffer, train, evaluate, export_onnx). |
| 2 | ARCH-02 Module boundaries respected | high | PASS | `internal/engine/` imports zero internal packages (grep `Mgrdich/TermChess/internal` in engine/ -> 0 matches). `bot/` imports only `engine`. `config/` imports only `engine`. `bvb/` imports `bot`+`engine`. `updater/` imports only `config`. No backwards imports. Python DAG verified acyclic (board_encoder and replay_buffer are leaves; train depends on self_play -> mcts -> model -> board_encoder). |
| 3 | ARCH-03 SRP in modules | medium | PASS | No god dirs: `util/` contains 2 files (clipboard only); `version/` 1 file; each `internal/*` package is cohesive (engine=rules, bot=AI, bvb=bot-vs-bot, ui=TUI, config=persistence, updater=self-update). No `helpers/`, `common/`, `misc/` generic dumping grounds. |
| 4 | ARCH-04 Separation of concerns | high | PASS | `internal/engine/*.go` (non-test) uses only `fmt` — no `os.`, no `bubbletea`, no `lipgloss`, no `net/http`. UI files `view.go`/`update.go` import `engine` for domain data (confirmed via grep) rather than embedding chess logic. Python `model.py` contains no IO; `train.py` imports exporter only indirectly via checkpoint files. |
| 5 | ARCH-05 Consistent naming | medium | PASS | All 36 Go files under `internal/` and `cmd/` are snake_case with `_test.go` suffix. All 8 Python source files + 6 test files in `training/` are snake_case with `test_*.py` prefix. Zero CamelCase or kebab-case filenames found. |
| 6 | ARCH-06 Reasonable file sizes | medium | FAIL | 2 non-test Go files exceed 2000 LOC: `internal/ui/view.go` 2798, `internal/ui/update.go` 2290. 6/37 non-test Go files (16%) exceed 500 LOC. Other >500 files: `bot/eval.go` 617, `ui/san.go` 574, `updater/updater.go` 548, `engine/moves.go` 542. Python: all source files <1000 LOC (max `train.py` 953). Threshold ">2000 LOC = FAIL" triggered. |

## Architecture Summary

- **Declared pattern (Go):** Layered architecture with pure domain core (engine), application layer (bot, bvb, config, updater), presentation (ui), entry (cmd). Explicitly documented in `/Users/mgo/Documents/TermChess/CLAUDE.md` layer map.
- **Declared pattern (Python):** Modular functional pipeline in `training/` — leaf modules (`board_encoder`, `replay_buffer`) feed domain (`model`, `mcts`) feed orchestration (`self_play`, `train`, `evaluate`, `export_onnx`).
- **Engine purity:** VERIFIED. `internal/engine/` has zero imports of other `internal/*` packages and uses only `fmt` from stdlib in non-test code. No UI/IO/network leakage.
- **Cross-language contract:** Documented in CLAUDE.md — `training/board_encoder.py` and `internal/bot/rl_encoder.go` must stay byte-identical. Contract tests exist on both sides.
- **Large files (>500 LOC, non-test):**
  - `internal/ui/view.go` — 2798 LOC (FAIL threshold, already flagged in CLAUDE.md as monolithic)
  - `internal/ui/update.go` — 2290 LOC (FAIL threshold, already flagged in CLAUDE.md as monolithic)
  - `internal/bot/eval.go` — 617 LOC
  - `internal/ui/san.go` — 574 LOC
  - `internal/updater/updater.go` — 548 LOC
  - `internal/engine/moves.go` — 542 LOC
- **Test-file sizes (informational):** Large test files are permissible but worth noting: `internal/ui/e2e_test.go` 4230, `internal/engine/moves_test.go` 3462, `internal/engine/board_test.go` 2381, `internal/engine/game_state_test.go` 1839, `internal/ui/san_test.go` 1526, `internal/engine/fen_test.go` 1517, `internal/bot/eval_test.go` 1509.
- **Naming:** CONSISTENT — 100% snake_case across Go and Python; zero deviations.
- **Circular imports:** None detected in Go or Python.

## Scoring calculation

- Max points: 2 (ARCH-01) + 2 (ARCH-02) + 1 (ARCH-03) + 2 (ARCH-04) + 1 (ARCH-05) + 1 (ARCH-06) = 9
- Earned: 2 + 2 + 1 + 2 + 1 + 0 = 8
- Score: 8/9 = 88.89% -> **89%** -> Grade **B**

Grade B reflects strong architectural discipline (clean layers, pure domain, consistent naming, clean import DAG) undermined by two UI files exceeding 2000 LOC — a known issue already flagged in project documentation.
