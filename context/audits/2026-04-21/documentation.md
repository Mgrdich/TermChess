# Documentation Quality — Audit Results

**Date:** 2026-04-21
**Score:** 92% — Grade **A−**

## Results

| #   | Check | Severity | Status | Evidence |
| --- | ----- | -------- | ------ | -------- |
| 1   | DOC-01 Root README exists and is useful | critical | PASS | `README.md` (415 lines): project name + description, feature list, install options (curl/manual/source), `make build`/`make test`/`make run` usage, TUI screenshots, bot-difficulty table, config keys, RL training quickstart with pointer to `training/training-docs.md`. A new dev can clone → `make build` → `./bin/termchess`. |
| 2   | DOC-02 Service-level READMEs exist | high | PASS | All service dirs have READMEs: `cmd/termchess/README.md`, `internal/engine/README.md`, `internal/bot/README.md`, `internal/bvb/README.md`, `internal/ui/README.md`, `internal/config/README.md`, `internal/updater/README.md`, `internal/version/README.md`, `internal/util/README.md`, `training/README.md`. Each describes responsibilities, key files, and build/run guidance where relevant. Coverage: 10/10. |
| 3   | DOC-03 API documentation is available | high | SKIP | Skip-When met: no REST/gRPC/GraphQL API per topology. The only inter-layer contract is ONNX (binary tensor `[batch, 18, 8, 8]`), documented in 4 places: root `CLAUDE.md` lines 16–27, `internal/bot/README.md` lines 15–21, `training/README.md`, and spec `008-custom-rl-agent`. |
| 4   | DOC-04 No stale documentation | medium | WARN | 1 of 6 sampled claims inaccurate (see table below). |

### Stale-claim sampling (DOC-04)

| # | Claim | Source | Verified result | Verdict |
|---|-------|--------|-----------------|---------|
| 1 | `view.go` is 2798 lines | `CLAUDE.md` | `wc -l internal/ui/view.go` → 2798 | Accurate |
| 2 | `update.go` is 2290 lines | `CLAUDE.md` | `wc -l internal/ui/update.go` → 2290 | Accurate |
| 3 | 44 Go test files + 6 Python test files | `CLAUDE.md` | Glob `**/*_test.go` → 44 matches; Glob `training/test_*.py` → 6 matches | Accurate |
| 4 | `newOnnxSession()` in `internal/bot/rl.go` returns `ErrModelNotLoaded` | `CLAUDE.md` lines 27 & 61 | `rl.go:213` `newOnnxSession` returns `errors.New("ONNX Runtime not yet configured…")`; `ErrModelNotLoaded` is returned by `SelectMove` at line 121 when session is nil | **Inaccurate** (wrong function attribution; runtime behavior described is correct in spirit) |
| 5 | `.golangci.yml` has 8 linters | `CLAUDE.md` | File enumerates gofmt, goimports, govet, errcheck, staticcheck, unused, ineffassign, misspell = 8 | Accurate |
| 6 | README commands `make build`/`make test`/`make run` | `README.md` | Makefile defines all three targets | Accurate |

Unverifiable without running tests: the "~83% UI coverage" figure in README.md/CLAUDE.md. No `cover` target exists in the Makefile; the claim is plausible and not contradicted by any other evidence.

## Documentation Summary

- **Root README:** present, 415 lines. Covers install (curl/manual/from-source/go-install), upgrade, uninstall, TUI usage with screenshots, bot difficulty tiers, config keys, development commands, project structure tree, architecture overview, roadmap, and link to the RL training pipeline docs.
- **Service READMEs:** 10/10 coverage. Present in `cmd/termchess/`, all 8 `internal/*` packages, and `training/`. Each is concise and describes responsibilities, key files, and — where relevant — build/run/testing commands.
- **API docs:** N/A (no HTTP/RPC API). The ONNX cross-language tensor contract (`[batch, 18, 8, 8]` float32, byte-identical between `training/board_encoder.py` and `internal/bot/rl_encoder.go`) is clearly documented as a load-bearing invariant in both the root `CLAUDE.md` and `internal/bot/README.md`.
- **Stale claims found:** 1 of 6 sampled — `CLAUDE.md` mis-attributes the `ErrModelNotLoaded` return to `newOnnxSession`; the sentinel is actually returned by `SelectMove` when the session is nil, while `newOnnxSession` returns a different inline error. Minor drift, not user-impacting.

## Scoring

- DOC-03 is SKIP → max_points = 3 (DOC-01 critical) + 2 (DOC-02 high) + 1 (DOC-04 medium) = **6**.
- Deductions: DOC-04 WARN medium = 0.5. Total earned = 6 − 0.5 = **5.5**.
- Score: 5.5 / 6 = **91.7%** → rounded **92%** → Grade **A−** (AWOS band A− = 90–92%).
