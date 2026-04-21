# Documentation Quality — Audit Results

**Date:** 2026-04-21
**Score:** 75% — Grade **B**

## Results

| #      | Check                                 | Severity | Status | Evidence                                                                                                                                                                                                                                                                                           |
| ------ | ------------------------------------- | -------- | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| DOC-01 | Root README exists and is useful      | critical | PASS   | `/Users/mgo/Documents/TermChess/README.md` (407 lines): project name/description, install (curl + manual + from source), `make build`/`make run`/`make test` dev commands, Usage, project structure, architecture, roadmap, RL training quick-start. A new dev can follow setup.                    |
| DOC-02 | Service-level READMEs exist           | high     | FAIL   | Zero per-service READMEs in Go services. No README in `cmd/termchess/`, `internal/bot/`, `internal/bvb/`, `internal/config/`, `internal/engine/`, `internal/ui/`, `internal/updater/`, `internal/version/`, `internal/util/`. `training/README.md` exists but is **empty (0 lines)**. `internal/config/` service dir is missing from topology but also lacks README. |
| DOC-03 | API documentation is available        | high     | SKIP   | TUI-only app, no network APIs (per topology summary). ONNX artifact contract is documented in `training/training-docs.md`.                                                                                                                                                                         |
| DOC-04 | No stale documentation                | medium   | WARN   | Sampled 5 claims: (1) `make build` target in Makefile — accurate; (2) `~/.termchess/config.toml` — accurate (`internal/config/paths.go:45`); (3) `training/training-docs.md` link — accurate; (4) `/training-health` slash command — accurate (`.claude/commands/training-health.md`); (5) Roadmap lists "RL-trained agent" as "In Progress / Planned" but `internal/bot/rl.go` and `internal/bot/rl_encoder.go` are already implemented — **stale**. Additionally `training/README.md` is referenced implicitly but is empty. |

## Summary

**Root README:** Strong. 407-line `README.md` covers installation (curl one-liner, manual download, from source), usage, board display, bot modes, difficulty table, configuration paths, development commands, project structure, architecture, roadmap, and RL training quick-start. Badges, code blocks, and examples are well-formed.

**Per-service README coverage:** 0 of 8 Go service directories (`cmd/termchess`, `internal/bot`, `internal/bvb`, `internal/config`, `internal/engine`, `internal/ui`, `internal/updater`, `internal/version`) have any README. `training/` has an empty `README.md` but mitigates this with a comprehensive 469-line `training-docs.md`. Overall: 0/9 service dirs have a usable `README.md`; 1/9 has alternative docs (training).

**Staleness findings:**
- One confirmed stale claim: Roadmap section marks "RL-trained agent" as "In Progress / Planned" but RL bot source files (`internal/bot/rl.go`, `internal/bot/rl_encoder.go`, `internal/bot/rl_test.go`, `internal/bot/rl_encoder_test.go`) exist and are integrated into the bot factory.
- Empty `training/README.md` is a documentation hygiene issue (file exists, should either be removed or populated with a pointer to `training-docs.md`).
- Most concrete command/path claims (Makefile targets, config path, slash command reference) are accurate.

**Recommendations:**
1. Add lightweight READMEs (5-15 lines each) to each `internal/*` service dir describing the package purpose and key entry points.
2. Populate `training/README.md` with a brief overview and pointer to `training-docs.md`, or delete it.
3. Update README roadmap to move "RL-trained agent" from "In Progress / Planned" to "Completed" (or clarify what remains in progress).
