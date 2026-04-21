# Code Audit Report

**Date:** 2026-04-21
**Scope:** all dimensions
**Overall Score:** 71% — Grade **C**
**Previous Audit:** none

## Summary

| #   | Dimension                | Score | Grade | Delta | Critical | High | Medium | Low |
| --- | ------------------------ | ----- | ----- | ----- | -------- | ---- | ------ | --- |
| 1   | Project Topology         | 100%  | A     | —     | 0        | 0    | 0      | 0   |
| 2   | AI Development Tooling   | 53%   | D     | —     | 1F       | 0    | 0      | 2F  |
| 3   | Code Architecture        | 72%   | C     | —     | 0        | 1W   | 1F/1W  | 0   |
| 4   | Documentation Quality    | 75%   | B     | —     | 0        | 1F   | 1W     | 0   |
| 5   | Security Guardrails      | 30%   | F     | —     | 2F       | 1W   | 0      | 0   |
| 6   | Software Best Practices  | 69%   | C     | —     | 0        | 3W   | 2W     | 0   |
| 7   | Spec-Driven Development  | 92%   | A     | —     | 0        | 1W   | 1F     | 0   |
| 8   | End-to-End Delivery      | 79%   | B     | —     | 0        | 0    | 1F/1W  | 0   |

Legend: F = FAIL, W = WARN. Columns tally findings by check severity.

## Dimension: Project Topology

**Score:** 100% — Grade **A**

| #   | Check                              | Severity | Status | Evidence |
| --- | ---------------------------------- | -------- | ------ | -------- |
| 1   | Repository structure type          | medium   | PASS   | Polyglot monorepo: 2 independent build roots (`go.mod`, `training/pyproject.toml`). |
| 2   | Application layer inventory        | medium   | PASS   | 7 layers detected: TUI entry (cmd/termchess), UI, engine, bot, bvb, updater, Python RL training. |
| 3   | Database and storage detection     | medium   | PASS   | Filesystem only — TOML config, FEN savegame, `.pt` checkpoints, `.npz` replay buffer. |
| 4   | Infrastructure layer detection     | medium   | SKIP   | No IaC; only GitHub Actions CI. |
| 5   | Language inventory                 | medium   | PASS   | Go (82 files), Python (14 files), Shell (2 files). |
| 6   | Inter-layer communication patterns | medium   | PASS   | ONNX model artifact is cross-language contract; board encoding `[batch, 18, 8, 8]` mirrored in Python & Go. |

## Dimension: AI Development Tooling

**Score:** 53% — Grade **D**

| #   | Check                                            | Severity | Status | Evidence |
| --- | ------------------------------------------------ | -------- | ------ | -------- |
| 1   | AI-01 CLAUDE.md ecosystem                        | critical | FAIL   | Zero CLAUDE.md files anywhere — no root, no per-layer. Context only in README + `training/training-docs.md`. |
| 2   | AI-02 Custom slash commands                      | medium   | PASS   | 10 commands (`training-health` plus 9 AWOS wrappers). |
| 3   | AI-03 Skills configured                          | low      | FAIL   | No `.claude/skills/` directory. |
| 4   | AI-04 MCP servers                                | low      | PASS   | `.mcp.json` configures `awos-recruitment` HTTP MCP. |
| 5   | AI-05 Hooks configured                           | low      | FAIL   | `.claude/settings.json` has no `hooks` key. |
| 6   | AI-06 CLAUDE.md meaningful and well-structured   | high     | SKIP   | No CLAUDE.md files to evaluate. |
| 7   | AI-07 Agent can run and observe application      | critical | PASS   | TUI is runnable via `make run`; Python via `uv run`. Built-in Bash is sufficient. |

## Dimension: Code Architecture

**Score:** 72% — Grade **C**

| #   | Check                                          | Severity | Status | Evidence |
| --- | ---------------------------------------------- | -------- | ------ | -------- |
| 1   | ARCH-01 Declared/recognizable pattern          | high     | PASS   | Canonical Go `cmd/`+`internal/` layered modular, declared in README. |
| 2   | ARCH-02 Module boundaries respected            | high     | PASS   | One-way dependency graph, no cycles, no reverse imports; `engine` is pure domain. |
| 3   | ARCH-03 SRP in modules                         | medium   | WARN   | `internal/util/` has generic name (though content is a 47-line clipboard helper only). |
| 4   | ARCH-04 Separation of concerns across layers   | high     | WARN   | `internal/ui/view.go` (2798 lines) and `internal/ui/update.go` (2290 lines) mix many screens. |
| 5   | ARCH-05 Consistent naming conventions          | medium   | PASS   | Snake_case throughout Go + Python; `_test.go` colocation; package/dir alignment. |
| 6   | ARCH-06 Reasonable file sizes                  | medium   | FAIL   | 22.2% of production files >500 lines (threshold 15%); 2 files >2000 lines. |

## Dimension: Documentation Quality

**Score:** 75% — Grade **B**

| #   | Check                              | Severity | Status | Evidence |
| --- | ---------------------------------- | -------- | ------ | -------- |
| 1   | DOC-01 Root README                 | critical | PASS   | 407 lines: install, usage, dev commands, structure, roadmap, RL training quick-start. |
| 2   | DOC-02 Service-level READMEs       | high     | FAIL   | 0/8 Go service dirs have a README; `training/README.md` is empty (0 lines). |
| 3   | DOC-03 API documentation           | high     | SKIP   | No network APIs (TUI app). |
| 4   | DOC-04 No stale documentation      | medium   | WARN   | Roadmap lists "RL-trained agent" as planned, but `internal/bot/rl.go` is implemented. |

## Dimension: Security Guardrails

**Score:** 30% — Grade **F**

| #   | Check                                               | Severity | Status | Evidence |
| --- | --------------------------------------------------- | -------- | ------ | -------- |
| 1   | SEC-01 `.env` files gitignored                      | critical | FAIL   | `.gitignore` has no `.env*` patterns. No tracked `.env` today, but guardrail absent. |
| 2   | SEC-02 AI hooks restrict sensitive file access      | critical | FAIL   | `.claude/settings.json` has no `PreToolUse` hooks; no deny coverage for `.env`, `*.pem`, `*.key`, credentials. |
| 3   | SEC-03 `.env.example`/template exists               | high     | SKIP   | No real env var usage (only `CI` in tests). |
| 4   | SEC-04 No secrets in committed files                | critical | PASS   | Clean — no api_key, password, token, private key, AWS key patterns. |
| 5   | SEC-05 Sensitive files in .gitignore coverage       | high     | WARN   | Python artifacts covered; missing `.env*`, `.DS_Store`, `Thumbs.db`, `*.pem`, `*.key`, `*.p12`, `*.pfx`. |

## Dimension: Software Best Practices

**Score:** 69% — Grade **C**

| #   | Check                                    | Severity | Status | Evidence |
| --- | ---------------------------------------- | -------- | ------ | -------- |
| 1   | SBP-01 Linting configured                | high     | WARN   | Go has `.golangci.yml` (8 linters); Python has no ruff/mypy/flake8/pylint config. |
| 2   | SBP-02 Formatting automated              | medium   | WARN   | Go gofmt via golangci-lint but not invoked in CI; no Python formatter. |
| 3   | SBP-03 Type safety enforced              | high     | WARN   | Go native; Python has 49 typed signatures but no mypy/pyright config. |
| 4   | SBP-04 Test infrastructure exists        | critical | PASS   | 50 test files (44 Go + 6 Python); pytest in dev deps. |
| 5   | SBP-05 CI/CD pipeline exists             | high     | WARN   | CI runs Go build+test only — no lint gate, no Python tests, no coverage. |
| 6   | SBP-06 Error handling consistent         | high     | PASS   | 603 idiomatic `if err != nil` sites; Python except blocks typed; no bare `except:`. |
| 7   | SBP-07 Dependencies managed              | medium   | WARN   | `go.sum` + `training/uv.lock` present; no dependabot/renovate automation. |

## Dimension: Spec-Driven Development

**Score:** 92% — Grade **A**

| #   | Check                                           | Severity | Status | Evidence |
| --- | ----------------------------------------------- | -------- | ------ | -------- |
| 1   | SDD-01 AWOS installed                           | critical | PASS   | 9 `.awos/commands/` + 9 `.claude/commands/awos/` wrappers; both context dirs present. |
| 2   | SDD-02 Product context documents complete       | high     | PASS   | product-definition (123 lines), roadmap (111 lines, 6 phases), architecture (116 lines). |
| 3   | SDD-03 Architecture reflects codebase reality   | high     | WARN   | ONNX Runtime declared but `go.mod` has no dep (tracked as 008/Slice 11); directory diagram omits `bvb/`, `updater/`, `util/`, `version/`. |
| 4   | SDD-04 Features implemented through specs       | critical | PASS   | ~75% of significant feature PRs modified `context/spec/`. |
| 5   | SDD-05 Spec directories structurally complete   | high     | PASS   | 8/8 dirs have full triad (functional-spec + tech + tasks). |
| 6   | SDD-06 No stale or abandoned specs              | medium   | PASS   | Zero stale — 6 Completed, 1 Implementation Complete, 1 actively-worked Draft. |
| 7   | SDD-07 Tasks have meaningful agent assignments  | medium   | FAIL   | No `[Agent: ...]` annotations in any tasks.md; 0% coverage. |

## Dimension: End-to-End Delivery

**Score:** 79% — Grade **B**

| #   | Check                                  | Severity | Status | Evidence |
| --- | -------------------------------------- | -------- | ------ | -------- |
| 1   | E2E-01 Cross-layer feature branches    | high     | PASS   | 4 commits in 3 months touch both Go and Python in a single commit (RL work, board encoder). |
| 2   | E2E-02 No layer-split branching        | medium   | PASS   | Only `main` and `RL-v1`; zero layer-split suffixes. |
| 3   | E2E-03 Spec-to-delivery traceability   | high     | PASS   | Bidirectional — tasks.md `[x]` matches shipped code; commit bodies reference slice IDs. |
| 4   | E2E-04 No orphaned artifacts           | medium   | WARN   | ONNX artifact is partially orphaned: Python exports, Go doesn't yet consume (tracked as Slice 11 `[ ]`). |
| 5   | E2E-05 Shared ownership enablers       | medium   | FAIL   | `Makefile` is Go-only; CI doesn't run Python tests/lint; no cross-layer orchestrator. |

## Top Recommendations

| #   | Priority | Effort | Dimension                | Recommendation |
| --- | -------- | ------ | ------------------------ | -------------- |
| 1   | P0       | Low    | Security                 | Add `.env`, `.env.local`, `.env.*.local` to `.gitignore`. Zero-risk preventive guardrail. |
| 2   | P0       | Medium | Security                 | Add `PreToolUse` deny hooks in `.claude/settings.json` for `.env`, `*.pem`, `*.key`, `credentials*`, `*secret*` patterns on Read/Glob/Bash. |
| 3   | P0       | Medium | AI Development Tooling   | Create root `CLAUDE.md` covering (a) the two build roots, (b) ONNX cross-language contract + duplicated board encoder, (c) layer map, (d) how to run each language's tests. Add `training/CLAUDE.md` for MCTS/self-play specifics. |
| 4   | P1       | Low    | Documentation            | Add 5-15 line READMEs to each `internal/*` service dir (purpose + key entry points); populate or delete empty `training/README.md`; correct roadmap "RL-trained agent" status. |
| 5   | P2       | Low    | End-to-End Delivery      | Add root Makefile targets (`train`, `export-onnx`, `py-test`, `lint-py`) and a second CI job that runs `pytest` + Python lint on `training/` — hardens the ONNX contract. |
| 6   | P2       | Low    | Software Best Practices  | Add `[tool.ruff]` + mypy/pyright config to `training/pyproject.toml`; invoke both plus `golangci-lint run` in CI (closes SBP-01, SBP-02, SBP-03, SBP-05 simultaneously). |
| 7   | P2       | Low    | Security                 | Extend `.gitignore` with `.DS_Store`, `Thumbs.db`, `*.pem`, `*.key`, `*.p12`, `*.pfx`. |
| 8   | P2       | Low    | Software Best Practices  | Add `.github/dependabot.yml` covering `gomod`, `pip`, and `github-actions` ecosystems. |
| 9   | P2       | Medium | Spec-Driven Development  | Add `**[Agent: agent-name]**` annotations to sub-tasks in existing `context/spec/*/tasks.md` files; adopt for new specs going forward. |
| 10  | P2       | Medium | Code Architecture        | Split `internal/ui/view.go` (2798 lines) and `internal/ui/update.go` (2290 lines) along screen boundaries (menu / game / settings / bvb / updater). Also addresses ARCH-04 WARN. |

Sorted by priority (P0 → P2), then by effort (Low → High).
