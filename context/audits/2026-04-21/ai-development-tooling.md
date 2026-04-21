# AI Development Tooling — Audit Results

**Date:** 2026-04-21
**Score:** 53% — Grade **D**

## Results

| #   | Check                                                    | Severity | Status | Evidence                                                                                                                                                                                                                                                                                                                                         |
| --- | -------------------------------------------------------- | -------- | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | AI-01 CLAUDE.md ecosystem provides adequate AI context  | critical | FAIL   | No CLAUDE.md files found anywhere — globbed `**/CLAUDE.md`, `**/CLAUDE*.md`, and `.claude/rules/*.md`: zero matches. No root CLAUDE.md, no service-level CLAUDE.md files for the 7 distinct layers (cmd/termchess, internal/ui, internal/engine, internal/bot, internal/bvb, internal/updater, training). Context lives only in README.md and training/training-docs.md. |
| 2   | AI-02 Custom slash commands exist                       | medium   | PASS   | 10 command files under `.claude/commands/`: `training-health.md` (full diagnostics skill, 180 lines) plus 9 AWOS commands under `.claude/commands/awos/` (architecture, hire, implement, product, roadmap, spec, tasks, tech, verify — all thin wrappers to `.awos/commands/*.md`). 3+ threshold met.                                             |
| 3   | AI-03 Skills are configured                             | low      | FAIL   | No `.claude/skills/` directory exists. Glob `.claude/skills/*/SKILL.md` returned zero files. (`.claude/` contains only `agents/`, `commands/`, `settings.json`, `settings.local.json`.)                                                                                                                                                         |
| 4   | AI-04 MCP servers configured                            | low      | PASS   | `/Users/mgo/Documents/TermChess/.mcp.json` defines the `awos-recruitment` HTTP MCP server (https://recruitment.awos.provectus.pro/mcp). `.claude/settings.local.json` enables it via `enabledMcpjsonServers` and also enables `gopls-lsp` and `pyright-lsp` plugins.                                                                          |
| 5   | AI-05 Hooks are configured                              | low      | FAIL   | `.claude/settings.json` contains only `extraKnownMarketplaces` (awos marketplace registration). No `hooks` key present in either `settings.json` or `settings.local.json`.                                                                                                                                                                      |
| 6   | AI-06 CLAUDE.md files are meaningful and well-structured | high     | SKIP   | Skip-When condition met: no CLAUDE.md files exist in the repository.                                                                                                                                                                                                                                                                           |
| 7   | AI-07 Agent can run and observe the application         | critical | PASS   | Built-in Bash is sufficient for this TUI/CLI + Python project. Run instructions are clearly documented in `README.md`: `make run`, `./bin/termchess`, `make build`, `make test` (lines 134-142, 271-284), and for training `cd training && uv sync && uv run python -u train.py ...` (lines 369-373). Makefile at `/Users/mgo/Documents/TermChess/Makefile` confirms `build`, `test`, `run`, `clean` targets. |

## Scoring

- Max points (excluding skipped AI-06): 3 (AI-01) + 1 (AI-02) + 0.5 (AI-03) + 0.5 (AI-04) + 0.5 (AI-05) + 3 (AI-07) = 8.5
- Deductions: 3 (AI-01 FAIL, critical) + 0.5 (AI-03 FAIL, low) + 0.5 (AI-05 FAIL, low) = 4.0
- Percentage: (8.5 − 4.0) / 8.5 × 100 = **52.94%**
- Grade: **D** (40-59)

## Summary

**CLAUDE.md coverage:** Complete absence. No root CLAUDE.md, no per-layer CLAUDE.md for any of the 7 distinct layers, and no `.claude/rules/` documents. This is the single biggest gap — the project is polyglot (Go + Python), has a non-obvious cross-language contract (ONNX model + duplicated board encoder in Python/Go), and contains complex modules (MCTS self-play, Bubbletea TUI state machine, minimax engine) that would benefit significantly from targeted AI context files. README.md and training/training-docs.md partially compensate but are human-facing and not structured as AI context.

**Custom slash commands:** Well-represented. `training-health` is a substantive, project-specific diagnostics command (180 lines, detailed metric ranges, failure-mode taxonomy). The 9 AWOS commands are mostly redirects to `.awos/commands/*.md` instructions — they count toward the threshold but add little project-specific AI context.

**Skills:** None configured. No `.claude/skills/` directory.

**MCP servers:** One configured — `awos-recruitment` HTTP server. Enabled in local settings.

**Hooks:** None configured. `.claude/settings.json` only declares the AWOS marketplace.

**Agents:** `.claude/agents/` contains three subagents (`awos-guide.md`, `go-cli-expert.md`, `ml-python-trainer-expert.md`) which partially substitute for the missing CLAUDE.md context, but agents are invoked explicitly rather than loaded as ambient context, so they do not satisfy AI-01.

**Observability posture:** Strong. TUI is runnable via `make run`/`./bin/termchess`, training via `uv run python -u train.py ...`. Training emits CSV metrics to `training/checkpoints/training_log.csv` with a companion diagnostic command. Built-in Bash is sufficient for all observation.

**Top recommendation:** Create a root `CLAUDE.md` that documents (1) the two build roots and when to use each, (2) the ONNX cross-language contract and the duplicated board encoder, (3) the map of internal layers, and (4) how to run tests for each language. Consider a `training/CLAUDE.md` for the MCTS/self-play pipeline specifics.
