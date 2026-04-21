# Code Architecture — Audit Results

**Date:** 2026-04-21
**Score:** 72% — Grade **C**

## Results

| #       | Check                                          | Severity | Status | Evidence                                                                                                                                                                                                                                                                                                                                                                                                 |
| ------- | ---------------------------------------------- | -------- | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ARCH-01 | Declared or recognizable architectural pattern | high     | PASS   | README declares structure + architecture (README.md lines 286-336). Go project follows the standard `cmd/` + `internal/<layer>/` modular/layered pattern with 8 focused packages under `internal/` (bot, bvb, config, engine, ui, updater, util, version). Python `training/` is a flat ML pipeline module — appropriate for the domain.                                                                   |
| ARCH-02 | Module boundaries are respected                | high     | PASS   | Clean dependency graph: `engine` has no internal imports; `bot`, `config` depend only on `engine`; `bvb` depends on `bot`+`engine`; `updater` depends on `config`; `ui` depends on all lower layers. No reverse imports from `engine`/`bot`/`bvb` back up the stack. No cycles. Verified across 75 files with internal imports.                                                                           |
| ARCH-03 | Single Responsibility Principle in modules     | medium   | WARN   | Most packages have clear, focused names (bot, bvb, engine, ui, config, updater, version). `internal/util/` is a borderline generic name but contains only a small (47-line) cross-platform `clipboard.go` helper — not a god module. No package contains 30+ files mixing concerns (largest: `internal/ui` with 31 files, but all UI-related).                                                           |
| ARCH-04 | Separation of concerns across layers           | high     | WARN   | Layer separation is clean overall, but within the UI layer two files are monolithic and mix sub-concerns: `internal/ui/view.go` (2798 lines) handles rendering for menu/game/settings/bvb/updater, and `internal/ui/update.go` (2290 lines) handles event handling for all those screens plus bot commands, update checks, and timers. Engine and bot packages are well-separated (fen/moves/attacks/etc). |
| ARCH-05 | Consistent file and directory naming           | medium   | PASS   | Go: all 82 files use lower_snake_case (e.g., `game_state.go`, `rl_encoder.go`); `_test.go` colocation verified; package names match directories for all 8 `internal/` packages. Python: all 14 files use snake_case (e.g., `board_encoder.py`, `replay_buffer.py`) with `test_*.py` prefix.                                                                                                               |
| ARCH-06 | Reasonable file sizes                          | medium   | FAIL   | Production code (45 files, tests excluded): 10 files (22.2%) exceed 500 lines (threshold 15%); 2 files (4.4%) exceed 2000 lines — `internal/ui/view.go` (2798 lines) and `internal/ui/update.go` (2290 lines). Tops in production: `ui/view.go` 2798, `ui/update.go` 2290, `bot/eval.go` 617, `ui/san.go` 574, `updater/updater.go` 548, `engine/moves.go` 542.                                            |

## Summary

**Architecture pattern:** Layered modular Go application using the canonical `cmd/`+`internal/` idiom. Layers are: entry point (`cmd/termchess`) → presentation (`internal/ui`) → application/orchestration (`internal/bvb`, `internal/updater`) → domain services (`internal/bot`, `internal/config`) → core domain (`internal/engine`). Cross-cutting: `internal/util`, `internal/version`. The ML training pipeline (`training/`) is a separate flat Python module appropriate for its domain. The `README.md` explicitly documents this structure.

**Boundary health:** Excellent. Strict one-directional dependency flow, no cycles, no reverse imports from domain layers into presentation. `internal/` visibility is leveraged correctly. Inside-out dependency direction holds: `engine` is a pure domain with zero internal deps.

**File-size distribution:** Production code shows a bimodal split. Most files are 100-400 lines and well-focused. However two UI files dominate: `view.go` (2798 lines) and `update.go` (2290 lines) together account for ~40% of production LOC and should be split along screen boundaries (menu, game, settings, bvb, etc.). The chess engine files (e.g., `moves.go` 542, `board.go` 456) are within acceptable range given the domain complexity. Including tests, 34/96 (35.4%) files exceed 500 lines, largely driven by exhaustive table-driven Go tests (e.g., `moves_test.go` 3462, `e2e_test.go` 4230) which is idiomatic.

**Naming consistency:** Strong. snake_case throughout both languages, test colocation follows language conventions, package names align with directory names in all 8 Go packages.

**Top recommendations:**
1. Split `internal/ui/view.go` into screen-specific files (`view_menu.go`, `view_game.go`, `view_settings.go`, `view_bvb.go`).
2. Split `internal/ui/update.go` similarly along screen/event handler boundaries.
3. Consider whether `internal/util/` could be renamed or its contents absorbed (e.g., `internal/clipboard/`).
