# Audit Recommendations — 2026-04-21

No P0 or P1 items. No critical or high-severity FAILs and no critical WARNs. All items below are P2 quality improvements.

## P2 — Improve When Possible

### 1. Fix `ErrModelNotLoaded` attribution in CLAUDE.md

- **Dimension:** Documentation Quality
- **Check:** DOC-04 (stale documentation)
- **Effort:** Low
- **Details:** The root `CLAUDE.md` claims `newOnnxSession()` in `internal/bot/rl.go` returns `ErrModelNotLoaded`. That sentinel is actually returned by `SelectMove` at `internal/bot/rl.go:121` when `session` is nil. `newOnnxSession` returns a different inline error (`errors.New("ONNX Runtime not yet configured…")` at `rl.go:213`). Rewrite the CLAUDE.md line to describe the real behavior: `SelectMove` short-circuits to `ErrModelNotLoaded` until `newOnnxSession` is implemented.

### 2. Add agent annotations to tasks.md files

- **Dimension:** Spec-Driven Development
- **Check:** SDD-07 (tasks agent assignments)
- **Effort:** Low
- **Details:** Zero `**[Agent: <name>]**` annotations were found across all 8 `context/spec/*/tasks.md` files. Since spec 008 is active, prioritize annotating its open sub-tasks first (e.g., `**[Agent: go-cli-developer]**` for Go implementation tasks, `**[Agent: ml-trainer]**` for training-loop work, `**[Agent: test-writer]**` for verification tasks). Completed specs (001–007) can be left as-is if historical; do not backfill purely for audit score.

### 3. Register at least one project-level skill

- **Dimension:** AI Development Tooling
- **Check:** AI-03 (skills configured)
- **Effort:** Low
- **Details:** `.claude/skills/` does not exist. Candidates that would earn their keep:
  - `training-health` — convert the existing `.claude/commands/training-health.md` into a skill with progressive disclosure (lightweight trigger description + deeper guidance loaded on demand). The logic is already in place.
  - `onnx-contract-sync` — triggered when editing `training/board_encoder.py` or `internal/bot/rl_encoder.go`, reminding the agent that both files must stay byte-identical and both test suites must pass in the same commit.
  - `spec-slice-commit` — enforces the CLAUDE.md convention that commit messages reference slice IDs (e.g., `008/Slice 11: …`).

### 4. Update the stale CI claim in CLAUDE.md

- **Dimension:** Documentation Quality
- **Check:** DOC-04 (stale documentation) — secondary observation
- **Effort:** Low
- **Details:** Root `CLAUDE.md` says "Python tests are **not** in CI today — run them locally before pushing changes to `training/`." This is no longer true — `.github/workflows/ci.yml` runs ruff check, ruff format check, mypy, and pytest on the `training/` tree. Remove or invert the note.

### 5. Split the two monolithic UI files

- **Dimension:** Code Architecture
- **Check:** ARCH-06 (reasonable file sizes)
- **Effort:** High
- **Details:** `internal/ui/view.go` (2798 LOC) and `internal/ui/update.go` (2290 LOC) both exceed the 2000-LOC hard threshold. CLAUDE.md already flags them as monolithic ("consider splitting by screen before adding new screens"). Suggested decomposition: split by screen into `view_menu.go`, `view_game.go`, `view_settings.go`, `view_bvb.go`, `view_updater.go` (and mirror for `update_*.go`). Keep the MVU shell (Init/Update/View dispatcher) in a thin `view.go` / `update.go`. This is a big refactor — plan it before Phase 5+ UI additions.

### 6. Close the ONNX runtime consumption gap

- **Dimension:** End-to-End Delivery
- **Check:** E2E-04 (no orphaned artifacts)
- **Effort:** High
- **Details:** `internal/bot/rl.go` currently stubs `newOnnxSession` with an inline "not yet configured" error, and the runtime dep `onnxruntime_go` is absent from `go.mod`. The Python side (`training/export_onnx.py`) already produces ONNX models. Close the loop by (a) adding `github.com/yalue/onnxruntime_go` to `go.mod`, (b) implementing `newOnnxSession` to load a model file given its bytes, (c) wiring session lifecycle through the existing interface. This is already tracked as spec 008 Slice 11 — no scope change, just execution.
