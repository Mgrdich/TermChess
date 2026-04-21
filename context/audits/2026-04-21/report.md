# Code Audit Report

**Date:** 2026-04-21
**Scope:** all dimensions (project-topology, documentation, security, ai-development-tooling, spec-driven-development, code-architecture, software-best-practices, end-to-end-delivery)
**Overall Score:** 95% — Grade **A**
**Previous Audit:** none

## Summary

| #   | Dimension                  | Score | Grade | Delta | Critical | High | Medium | Low |
| --- | -------------------------- | ----- | ----- | ----- | -------- | ---- | ------ | --- |
| 1   | Project Topology           | 100%  | A     | —     | 0        | 0    | 0      | 0   |
| 2   | Documentation Quality      | 92%   | A−    | —     | 0        | 0    | 1      | 0   |
| 3   | Security Guardrails        | 100%  | A     | —     | 0        | 0    | 0      | 0   |
| 4   | AI Development Tooling     | 95%   | A     | —     | 0        | 0    | 0      | 1   |
| 5   | Spec-Driven Development    | 93%   | A     | —     | 0        | 0    | 1      | 0   |
| 6   | Code Architecture          | 89%   | B     | —     | 0        | 0    | 1      | 0   |
| 7   | Software Best Practices    | 100%  | A     | —     | 0        | 0    | 0      | 0   |
| 8   | End-to-End Delivery        | 93%   | A     | —     | 0        | 0    | 1      | 0   |

*Severity columns count FAIL + WARN issues at that severity.*

Headline: TermChess is a well-run polyglot monorepo. Linting, tests, CI, security hooks, AWOS specs, and cross-layer tooling are all in place and consistent with the project's declared architecture. Two issues stand out: the two monolithic UI files (>2000 LOC) flagged in CLAUDE.md, and a small documentation drift item. The ONNX runtime integration gap in Go is the only open cross-layer orphan, and it's actively tracked as spec 008 Slice 11.

---

## Dimension: Project Topology

**Score:** 100% — Grade **A**

| # | Check | Severity | Status | Evidence |
| - | ----- | -------- | ------ | -------- |
| 1 | TOPO-01 Repository structure type | medium | PASS | Monorepo: `go.mod` + `training/pyproject.toml` (2 independent build roots) |
| 2 | TOPO-02 Application layer inventory | medium | PASS | 3 layers: CLI/TUI (`cmd/termchess`), 8 Go packages (`internal/*`), Python RL pipeline (`training/`) |
| 3 | TOPO-03 Database and storage detection | medium | PASS | Local filesystem only — TOML config + FEN savegame at `~/.termchess/`, NPZ replay buffer + checkpoints (gitignored). No DB |
| 4 | TOPO-04 Infrastructure layer detection | medium | PASS | GitHub Actions (`ci.yml`, `release.yml`), Makefile, `scripts/install.sh`. No container/IaC |
| 5 | TOPO-05 Language inventory | medium | PASS | Go 82, Python 14, Markdown 75, Shell 2, TOML 1 |
| 6 | TOPO-06 Inter-layer communication patterns | medium | PASS | ONNX cross-language contract; duplicated board encoder producing `[batch, 18, 8, 8]` tensor |

## Dimension: Documentation Quality

**Score:** 92% — Grade **A−**

| # | Check | Severity | Status | Evidence |
| - | ----- | -------- | ------ | -------- |
| 1 | DOC-01 Root README exists and is useful | critical | PASS | `README.md` (415 lines): install/upgrade/uninstall, TUI usage, bot difficulty, dev commands, RL pointer |
| 2 | DOC-02 Service-level READMEs exist | high | PASS | 10/10 service dirs have dedicated READMEs |
| 3 | DOC-03 API documentation is available | high | SKIP | No REST/gRPC/GraphQL API; ONNX contract documented in CLAUDE.md + `internal/bot/README.md` + `training/README.md` |
| 4 | DOC-04 No stale documentation | medium | WARN | 1 of 6 sampled claims inaccurate — `CLAUDE.md` mis-attributes `ErrModelNotLoaded` return to `newOnnxSession` (sentinel actually returned by `SelectMove` at `internal/bot/rl.go:121`) |

## Dimension: Security Guardrails

**Score:** 100% — Grade **A**

| # | Check | Severity | Status | Evidence |
| - | ----- | -------- | ------ | -------- |
| 1 | SEC-01 .env files gitignored | critical | PASS | `.env*` patterns in `.gitignore`; 0 `.env` files tracked |
| 2 | SEC-02 AI hooks restrict sensitive files | critical | PASS | `.claude/settings.json` PreToolUse hooks block Read/Edit/Write/Bash on `.env`, `*.pem`, `*.key`, `*.p12`, `*.pfx`, `credentials*`, `secrets*`, SSH keys, AWS/K8s creds |
| 3 | SEC-03 .env template exists | high | SKIP | No runtime env-var usage (only `os.Getenv("CI")` in a build-flag test) |
| 4 | SEC-04 No secrets in committed files | critical | PASS | 0 matches across 5 secret pattern families |
| 5 | SEC-05 Sensitive file coverage in .gitignore | high | PASS | Stack-relevant coverage: `.env*`, `*.pem`/`*.key`/`*.p12`/`*.pfx`, `.DS_Store`, Python venv/cache, checkpoints, binaries |

## Dimension: AI Development Tooling

**Score:** 95% — Grade **A**

| # | Check | Severity | Status | Evidence |
| - | ----- | -------- | ------ | -------- |
| 1 | AI-01 CLAUDE.md ecosystem | critical | PASS | 3 files (root 63L, `training/` 60L, `internal/ui/` 43L) covering purpose, commands, ONNX contract, layer map, gotchas |
| 2 | AI-02 Custom slash commands | medium | PASS | 10 commands (`training-health` + 9 AWOS wrappers) |
| 3 | AI-03 Skills configured | low | FAIL | `.claude/skills/` does not exist |
| 4 | AI-04 MCP servers configured | low | PASS | `.mcp.json` declares `awos-recruitment` HTTP server |
| 5 | AI-05 Hooks configured | low | PASS | 2 PreToolUse hooks enforcing SEC-02 |
| 6 | AI-06 CLAUDE.md quality | high | PASS | All files <200 lines, concrete bullets/tables, no discoverable-content bloat |
| 7 | AI-07 Agent can run/observe app | critical | PASS | Go + Python run/test paths documented; Bash + Read cover both app types |

## Dimension: Spec-Driven Development

**Score:** 93% — Grade **A**

| # | Check | Severity | Status | Evidence |
| - | ----- | -------- | ------ | -------- |
| 1 | SDD-01 AWOS installed | critical | PASS | `.awos/commands/` (9 files) + `.claude/commands/awos/` (9 wrappers) + `context/product/` + `context/spec/` |
| 2 | SDD-02 Product context complete | high | PASS | `product-definition.md` (122L), `roadmap.md` (110L, 6 phases), `architecture.md` (122L, 5 sections) |
| 3 | SDD-03 Architecture reflects reality | high | PASS | Declared stack confirmed by `go.mod` + `training/pyproject.toml`. `onnxruntime_go` pending Slice 11 is explicitly flagged in architecture.md |
| 4 | SDD-04 Features implemented via specs | critical | PASS | Adapted ratio: 71% of substantive feature commits touched `context/spec/**`; RL slice commits map 1:1 to spec 008 tasks.md items |
| 5 | SDD-05 Spec dirs complete | high | PASS | 8/8 spec directories have full triad (functional-spec, technical-considerations, tasks) |
| 6 | SDD-06 No stale specs | medium | PASS | 7 completed specs; spec 008 actively progressing (52/76 tasks) |
| 7 | SDD-07 Tasks agent assignments | medium | FAIL | 0 `**[Agent:...]**` annotations across all 8 tasks.md files — tasks use slice-based organization instead |

## Dimension: Code Architecture

**Score:** 89% — Grade **B**

| # | Check | Severity | Status | Evidence |
| - | ----- | -------- | ------ | -------- |
| 1 | ARCH-01 Pattern declared/recognizable | high | PASS | Layer map in CLAUDE.md verified against 8 Go packages; Python `training/` is a modular functional pipeline |
| 2 | ARCH-02 Module boundaries respected | high | PASS | `internal/engine/` imports zero internal packages; Go import DAG is strictly downward; Python DAG acyclic |
| 3 | ARCH-03 SRP in modules | medium | PASS | Every package is cohesive; no god dirs; no generic `helpers/`/`common/`/`misc/` dumping grounds |
| 4 | ARCH-04 Separation of concerns | high | PASS | Engine is pure (only imports `fmt`); UI imports engine for domain data rather than embedding chess logic |
| 5 | ARCH-05 Consistent naming | medium | PASS | 100% snake_case across Go and Python; zero deviations |
| 6 | ARCH-06 Reasonable file sizes | medium | FAIL | `internal/ui/view.go` 2798 LOC and `internal/ui/update.go` 2290 LOC both exceed the 2000-LOC hard threshold. 6/37 non-test Go files (16%) exceed 500 LOC |

## Dimension: Software Best Practices

**Score:** 100% — Grade **A**

| # | Check | Severity | Status | Evidence |
| - | ----- | -------- | ------ | -------- |
| 1 | SBP-01 Linting configured | high | PASS | Go: `.golangci.yml` with 8 linters; Python: ruff + mypy in `training/pyproject.toml`; both enforced in CI |
| 2 | SBP-02 Formatting automated | medium | PASS | `gofmt`/`goimports` via golangci-lint; ruff format check locally + in CI |
| 3 | SBP-03 Type safety enforced | high | PASS | Go: 0 `interface{}` in `internal/`; Python: ~85% typed signatures with mypy strict flags in CI |
| 4 | SBP-04 Test infrastructure | critical | PASS | 44 Go `*_test.go` + 6 Python `test_*.py`; both runners wired via Makefile |
| 5 | SBP-05 CI/CD pipeline | high | PASS | `ci.yml` (Go + Python lint/test/type-check) + `release.yml` (multi-platform build + checksums + GH Release) |
| 6 | SBP-06 Error handling | high | PASS | 603 Go `if err != nil` sites; 39 `fmt.Errorf(%w)` wraps; 0 bare Python `except:` |
| 7 | SBP-07 Dependencies managed | medium | PASS | `go.sum` + `training/uv.lock`; `.github/dependabot.yml` covers gomod, pip, github-actions |

## Dimension: End-to-End Delivery

**Score:** 93% — Grade **A**

| # | Check | Severity | Status | Evidence |
| - | ----- | -------- | ------ | -------- |
| 1 | E2E-01 Cross-layer feature commits | high | PASS | 5/16 substantive commits cross-layer (31% overall); 4/8 within spec-008 (50%); headline vertical-slice commit `5b203a5` lands 30-file RL-playing mechanism across Python + Go simultaneously |
| 2 | E2E-02 No layer-split branching | medium | PASS | No `*-backend` / `*-frontend` branch pairs; `RL-v1` is a full-stack feature branch, not a layer split |
| 3 | E2E-03 Spec-to-delivery traceability | high | PASS | Bidirectional — spec 008 `[x]` marks align with Python commits; commit messages reference spec phases; PRs #20/#21/#27/#28 map 1:1 to spec phases |
| 4 | E2E-04 No orphaned artifacts | medium | WARN | `training/export_onnx.py` produces ONNX model artifacts that `internal/bot/rl.go` cannot yet consume (`newOnnxSession` stubbed, `ErrModelNotLoaded` returned). Explicitly tracked as spec 008 Slice 11 |
| 5 | E2E-05 Shared ownership enablers | medium | PASS | Root `Makefile` orchestrates both layers; CI has parallel `go` + `python` jobs on every push/PR |

---

## Top Recommendations

| #   | Priority | Effort | Dimension              | Recommendation                                                                                                      |
| --- | -------- | ------ | ---------------------- | ------------------------------------------------------------------------------------------------------------------- |
| 1   | P2       | Low    | Documentation          | Fix `ErrModelNotLoaded` attribution in `CLAUDE.md` — the sentinel is returned by `SelectMove`, not `newOnnxSession` |
| 2   | P2       | Low    | Spec-Driven Development| Add `**[Agent: <name>]**` annotations to sub-tasks in `context/spec/*/tasks.md` so future runs hit SDD-07           |
| 3   | P2       | Low    | AI Development Tooling | Add `.claude/skills/<name>/SKILL.md` for at least one project-specific skill (candidates: `training-health`, `onnx-contract-sync`) |
| 4   | P2       | Medium | Documentation          | Update the stale CLAUDE.md note "Python tests are not in CI today" — they are (CI runs ruff, mypy, and pytest)     |
| 5   | P2       | High   | Code Architecture      | Split `internal/ui/view.go` (2798 LOC) and `internal/ui/update.go` (2290 LOC) by screen (menu, game, settings, bvb, updater) — already flagged as a known issue in CLAUDE.md |
| 6   | P2       | High   | End-to-End Delivery    | Close the ONNX runtime consumption gap in `internal/bot/rl.go` (add `onnxruntime_go` dep, implement `newOnnxSession`) — tracked as spec 008 Slice 11 |

No P0 or P1 items — no critical or high-severity FAILs, and no critical WARNs. All remaining items are P2 quality improvements.
