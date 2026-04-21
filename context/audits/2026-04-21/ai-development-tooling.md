# AI Development Tooling — Audit Results

**Date:** 2026-04-21
**Score:** 95% — Grade **A**

## Results

| # | Check | Severity | Status | Evidence |
| - | ----- | -------- | ------ | -------- |
| 1 | AI-01 CLAUDE.md ecosystem | critical | PASS | 3 files (root 63L, `training/` 60L, `internal/ui/` 43L) cover purpose, build roots + commands, cross-language ONNX contract, layer map, conventions, screen-level gotchas |
| 2 | AI-02 Custom slash commands | medium | PASS | 10 commands in `.claude/commands/` tracked by git: `training-health.md` + `awos/{architecture,hire,implement,product,roadmap,spec,tasks,tech,verify}.md` |
| 3 | AI-03 Skills configured | low | FAIL | `.claude/skills/` does not exist; no `SKILL.md` files found |
| 4 | AI-04 MCP servers configured | low | PASS | `.mcp.json` at repo root declares HTTP server `awos-recruitment` (`https://recruitment.awos.provectus.pro/mcp`); enabled in `settings.local.json` |
| 5 | AI-05 Hooks configured | low | PASS | `.claude/settings.json` has 2 PreToolUse hooks (SEC-02): Read/Edit/Write/NotebookEdit sensitive-file blocker and Bash sensitive-path blocker |
| 6 | AI-06 CLAUDE.md quality | high | PASS | All 3 files well under 200L (63/60/43), use concrete tables and bullets, cite real file paths (`training/board_encoder.py`, `internal/bot/rl_encoder.go`), and avoid vague platitudes or directory-tree repetition |
| 7 | AI-07 Agent can run/observe app | critical | PASS | Go run/test/build documented in root `CLAUDE.md` (`make build`, `make test`, `make run`, `go test -v ./...`); Python pipeline documented in `training/CLAUDE.md` (`uv sync`, `uv run pytest`, `uv run python -u train.py --help`). Bash + Read cover both app types — no web/API surfaces needing browser MCP |

## AI Tooling Summary

- **CLAUDE.md files:** 3 files — `/Users/mgo/Documents/TermChess/CLAUDE.md` (63L, root/cross-cutting), `/Users/mgo/Documents/TermChess/training/CLAUDE.md` (60L, Python RL pipeline), `/Users/mgo/Documents/TermChess/internal/ui/CLAUDE.md` (43L, Bubbletea MVU); total 166 lines.
- **Commands:** 10 custom — `training-health`, `awos/architecture`, `awos/hire`, `awos/implement`, `awos/product`, `awos/roadmap`, `awos/spec`, `awos/tasks`, `awos/tech`, `awos/verify`.
- **Skills:** 0 — no `.claude/skills/` directory.
- **MCP / Plugins:** 1 MCP server (`awos-recruitment` via HTTP) in `.mcp.json`; local enabled plugins `gopls-lsp@claude-plugins-official`, `pyright-lsp@claude-plugins-official` (LSP, not MCP); marketplace `awos-marketplace` registered.
- **Hooks:** 2 PreToolUse hooks enforcing SEC-02 (block reads/writes of `.env`, `.pem`, `.key`, `.p12`, `.pfx`, `credentials`, `id_rsa`, `id_ed25519`, `.aws/credentials`, `.kube/config`; block corresponding `cat/less/curl/scp/cp/mv/base64/...` shell invocations).
- **Run/observe:** Primary = Go CLI/TUI (runnable via `make run` / `go test -v ./...`); secondary = Python RL training (runnable via `uv run pytest` / `uv run python -u train.py`). Both documented in respective CLAUDE.md files and covered by default Bash + Read tools.

## Deductions

- AI-03 FAIL (low): −0.5

Total max = 10.5; deductions = 0.5; final = 10.0 / 10.5 ≈ 95.2% → Grade **A**.
