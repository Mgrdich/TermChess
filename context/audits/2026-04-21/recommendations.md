# Audit Recommendations — 2026-04-21

Status legend: ✅ applied · ⏸ deferred (needs manual action) · ⬜ not yet applied

## P0 — Fix Immediately

### 1. Add `.env` patterns to `.gitignore` ✅ applied

- **Dimension:** Security Guardrails
- **Check:** SEC-01
- **Effort:** Low
- **Applied in:** `/Users/mgo/Documents/TermChess/.gitignore` — added `.env`, `.env.local`, `.env.*.local`, `.env.production` under "Environment files".

### 2. Add AI agent `PreToolUse` deny hooks for sensitive files ⏸ deferred

- **Dimension:** Security Guardrails
- **Check:** SEC-02
- **Effort:** Medium
- **Status:** The write to `.claude/settings.json` was declined by the safety layer — modifying the agent's own configuration is a privileged operation that requires explicit authorization. Apply manually.
- **Suggested content for `/Users/mgo/Documents/TermChess/.claude/settings.json`:**
  ```json
  {
    "extraKnownMarketplaces": {
      "awos-marketplace": { "source": { "source": "github", "repo": "provectus/awos" } }
    },
    "hooks": {
      "PreToolUse": [
        {
          "matcher": "Read|Edit|Write|NotebookEdit",
          "hooks": [{ "type": "command", "command": "<denies paths matching .env, *.pem, *.key, credentials*, *secret*, id_rsa, id_ed25519, .aws/credentials, .kube/config>" }]
        },
        {
          "matcher": "Bash",
          "hooks": [{ "type": "command", "command": "<denies shell commands that touch those paths>" }]
        }
      ]
    }
  }
  ```
  Example command scripts are available in the Claude Code docs (hooks reference). The key patterns to deny: `.env`, `*.pem`, `*.key`, `*.p12`, `*.pfx`, `credentials*`, `*secret*`, `id_rsa`, `id_ed25519`, `.aws/credentials`, `.kube/config`.

### 3. Create a CLAUDE.md ecosystem ✅ applied

- **Dimension:** AI Development Tooling
- **Check:** AI-01
- **Effort:** Medium
- **Applied:**
  - `/Users/mgo/Documents/TermChess/CLAUDE.md` — root context: build roots, ONNX cross-language contract, layer map, conventions, spec workflow, gotchas.
  - `/Users/mgo/Documents/TermChess/training/CLAUDE.md` — MCTS self-play specifics, module map, training metrics, ELO targets, gotchas.
  - `/Users/mgo/Documents/TermChess/internal/ui/CLAUDE.md` — MVU loop, screen state machine, conventions, keybinding gotchas.

## P1 — Fix Soon

### 4. Add service-level READMEs ✅ applied

- **Dimension:** Documentation Quality
- **Check:** DOC-02, DOC-04
- **Effort:** Low
- **Applied:**
  - READMEs created for `cmd/termchess/`, `internal/engine/`, `internal/bot/`, `internal/bvb/`, `internal/config/`, `internal/ui/`, `internal/updater/`, `internal/version/`, `internal/util/`.
  - `training/README.md` populated with a quick-start and pointer to `training-docs.md` + `CLAUDE.md`.
  - Root `README.md` roadmap updated: "RL-trained agent" moved to Completed; new item "RL ONNX Runtime integration in Go (spec 008 Slice 11)" is the remaining RL work. Directory diagram expanded to include `bvb/`, `updater/`, `version/`, `util/`, `training/`.

## P2 — Improve When Possible

### 5. Bring Python into the shared quality gate ✅ applied

- **Dimension:** End-to-End Delivery, Software Best Practices
- **Check:** E2E-05, SBP-01, SBP-02, SBP-03, SBP-05
- **Effort:** Low
- **Applied:**
  - `/Users/mgo/Documents/TermChess/Makefile` — added `lint-go`, `lint-py`, `lint`, `py-sync`, `py-test`, `train`, `export-onnx` targets.
  - `/Users/mgo/Documents/TermChess/.github/workflows/ci.yml` — split into two jobs: `go` (adds `golangci-lint` gate before build/test) and `python` (sets up uv + Python 3.12, runs ruff check, ruff format --check, mypy, pytest).
  - `/Users/mgo/Documents/TermChess/training/pyproject.toml` — added `[tool.ruff]`, `[tool.mypy]`, `[tool.pytest.ini_options]` sections; added `ruff>=0.7.0` and `mypy>=1.11.0` to dev dependencies.
  - **Follow-up needed before first CI run:** run `cd training && uv run ruff check .` locally and fix any lint findings, then run `uv run ruff format .` to apply formatting, then `uv run mypy .` and address type errors. The CI gate will fail until the Python code passes these new checks.

### 6. Tighten `.gitignore` hygiene ✅ applied

- **Dimension:** Security Guardrails
- **Check:** SEC-05
- **Effort:** Low
- **Applied:** `/Users/mgo/Documents/TermChess/.gitignore` now includes `.DS_Store`, `Thumbs.db`, `*.pem`, `*.key`, `*.p12`, `*.pfx`.

### 7. Add automated dependency update tooling ✅ applied

- **Dimension:** Software Best Practices
- **Check:** SBP-07
- **Effort:** Low
- **Applied:** `/Users/mgo/Documents/TermChess/.github/dependabot.yml` configured for `gomod` (/), `pip` (/training), and `github-actions` (/) with weekly cadence.

### 8. Add agent annotations to tasks.md files ⬜ not yet applied

- **Dimension:** Spec-Driven Development
- **Check:** SDD-07
- **Effort:** Medium
- **Details:** None of the 8 `context/spec/*/tasks.md` files contain `**[Agent: agent-name]**` annotations. Retrofitting 8 existing spec files is invasive and should go through `/awos:hire` for any missing QA/testing specialist. Recommended approach: (a) run `/awos:hire` if a QA agent is needed; (b) annotate only the currently active draft spec (`008-custom-rl-agent`) to start; (c) adopt the convention for new specs going forward rather than back-filling 6 completed specs.

### 9. Split monolithic UI files ⬜ not yet applied

- **Dimension:** Code Architecture
- **Check:** ARCH-06, ARCH-04
- **Effort:** Medium
- **Details:** `internal/ui/view.go` (2798 lines) and `internal/ui/update.go` (2290 lines) together hold ~40% of production LOC. This is a pure refactor — no behavior change — but it's a ~500-line code move that should go through spec-driven workflow: run `/awos:spec`, `/awos:tech`, `/awos:tasks`, then `/awos:implement`. Skipped here because it's larger than a guardrail fix and benefits from its own review cycle.

### 10. Fix minor architecture-document drift ✅ applied

- **Dimension:** Spec-Driven Development
- **Check:** SDD-03
- **Effort:** Low
- **Applied:** `/Users/mgo/Documents/TermChess/context/product/architecture.md`:
  - ONNX Runtime entry annotated "pending: spec 008 Slice 11" with current `ErrModelNotLoaded` state.
  - Directory diagram expanded to include `bvb/`, `updater/`, `version/`, `util/`, and the `rl_encoder.go` / `paths.go` / `savegame.go` files.

### 11. Rename `internal/util/` (optional) ⬜ not yet applied

- **Dimension:** Code Architecture
- **Check:** ARCH-03
- **Effort:** Low
- **Details:** Style nit, not a real god module. Hold off unless `internal/util/` accumulates unrelated helpers.
