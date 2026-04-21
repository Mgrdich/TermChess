# End-to-End Delivery — Audit Results

**Date:** 2026-04-21
**Score:** 93% — Grade **A**

## Results

| # | Check | Severity | Status | Evidence |
| - | ----- | -------- | ------ | -------- |
| 1 | E2E-01 Cross-layer feature branches/commits | high | PASS | 23 commits in last 3 months, main-only. Substantive feature commits (excl. pure docs/audit/config): 16. Cross-layer (touches `training/**` + `internal/**` or `cmd/**`) = 5: `a40f87d` (audit hardening), `292daac` (feat! move history + augmentation + crash recovery, 12 files Go+Py), `ff4dbc45` (PR review fixes across RL pipeline, 6 files Go+Py), `b69f8add` (PR review fixes RL pipeline, 6 files Go+Py), `5b203a50` (feat: RL playing mechanism, 30 files Py+Go+specs+go.mod+pyproject.toml). Within spec-008 RL feature commits (8 commits), 4/8 = 50% are cross-layer. Overall substantive-cross-layer ratio 5/16 ≈ 31%, but the vertical-delivery headline commit `5b203a5` lands the entire RL-playing mechanism across both languages simultaneously — textbook vertical slice. Single-maintainer project with clear vertical-slice pattern for the cross-layer feature (spec 008) → PASS |
| 2 | E2E-02 No layer-split branching | medium | PASS | `git branch -a`: `main`, `origin/main`, `origin/RL-v1`. No `*-backend`/`*-frontend`/`*-api`/`*-ui` suffix pairs found. `RL-v1` is a feature branch for the RL agent (full-stack: Python + Go), not a layer split. Repo is effectively main-only. Vacuously PASS |
| 3 | E2E-03 Spec-to-delivery traceability | high | PASS | Bidirectional traceability. Specs → commits: spec 008 tasks.md `[x]` marks traceable to Python commits (292daac move history = Slice 7, 304de07 Dirichlet = Slice 8, 5b203a5 RL playing = Slice 10). Commits → specs: 2/23 commit subjects explicitly reference specs ("update specs for 66-channel encoding" in 2ff674f; "Phase 4 mouse interaction specification" in 7f980f1); 5/16 substantive commits modify `context/spec/**` files (per SDD-04). PRs #20, #21, #27, #28 map 1:1 to spec phases. CLAUDE.md mandates slice-ID references in commit messages (e.g., "008/Slice 11: …"). Traceability matrix verified |
| 4 | E2E-04 No orphaned artifacts | medium | WARN | One known orphan, tracked: `training/export_onnx.py` produces ONNX model artifacts that Go `internal/bot/rl.go` cannot yet consume — `newOnnxSession(_ []byte)` at rl.go:213 is stubbed and `ErrModelNotLoaded` ("ONNX runtime integration pending") is returned at rl.go:49. Runtime dep `onnxruntime_go` absent from `go.mod`. Encoder duplication (`training/board_encoder.py` ↔ `internal/bot/rl_encoder.go`) is NOT an orphan — documented invariant (CLAUDE.md cross-language contract) with test parity across both suites. Gap is explicitly tracked as spec 008 Slice 11 (in progress). Given explicit tracking, WARN rather than FAIL |
| 5 | E2E-05 Shared ownership enablers | medium | PASS | Root `Makefile` orchestrates both layers: Go (`build`, `build-all`, `test`, `run`, `lint-go`) + Python (`py-sync`, `py-test`, `train`, `export-onnx`, `lint-py`), plus unified `lint: lint-go lint-py`. `.github/workflows/ci.yml` has two parallel jobs (`go` + `python`), both gated on every push/PR to main. Python job runs ruff lint, ruff format check, mypy type check, and pytest — contradicts CLAUDE.md's note that "Python tests are not in CI today". Cross-layer tooling fully in place |

## Delivery Summary
- **Branching model:** main-only (local `main`; remote `main` + `RL-v1` full-stack feature branch, not a layer split)
- **Cross-layer delivery:** 5/16 substantive commits touch both Go and Python (31%); within spec-008 RL feature work, 4/8 RL commits are cross-layer (50%). Headline vertical-slice commit `5b203a5` lands the RL-playing mechanism across 30 files in Python+Go+specs simultaneously
- **Traceability:** bidirectional — spec 008 tasks.md `[x]` marks align with commit history; commit messages reference spec phases/slices; CLAUDE.md mandates slice-ID references; 5 commits directly modify spec files
- **Orphans:** 1 known, tracked — ONNX runtime consumption gap in `internal/bot/rl.go` (stubbed `newOnnxSession`, `ErrModelNotLoaded`); actively tracked as spec 008 Slice 11. Encoder duplication is a documented invariant, not an orphan
- **Shared tooling:** full coverage — root `Makefile` has Go + Python targets; CI has parallel `go` and `python` jobs both gated on every PR/push to main

## Scoring

| Check  | Severity | Weight | Status | Points |
| ------ | -------- | ------ | ------ | ------ |
| E2E-01 | high     | 2      | PASS   | 2.0    |
| E2E-02 | medium   | 1      | PASS   | 1.0    |
| E2E-03 | high     | 2      | PASS   | 2.0    |
| E2E-04 | medium   | 1      | WARN   | 0.5    |
| E2E-05 | medium   | 1      | PASS   | 1.0    |

Total: 6.5 / 7 = 92.9% → Grade **A**
