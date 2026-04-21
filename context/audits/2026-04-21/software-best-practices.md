# Software Best Practices — Audit Results

**Date:** 2026-04-21
**Score:** 69% — Grade **C**

## Results

| #      | Check                                    | Severity | Status | Evidence                                                                                                                                                                 |
| ------ | ---------------------------------------- | -------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| SBP-01 | Linting is configured and enforced       | high     | WARN   | Go: `.golangci.yml` enables gofmt, goimports, govet, errcheck, staticcheck, unused, ineffassign, misspell. Python: no `[tool.ruff]`/`[tool.flake8]`/`[tool.mypy]` in `training/pyproject.toml`; no `.flake8`/`.pylintrc`/`ruff.toml`/`mypy.ini`. |
| SBP-02 | Formatting is automated                  | medium   | WARN   | Go: `gofmt` + `goimports` enforced via golangci-lint config. However `.github/workflows/ci.yml` only runs `make build` + `make test` — no `golangci-lint` invocation and no pre-commit hooks. Python: no `black`/`ruff format`/`autopep8` configured.        |
| SBP-03 | Type safety is enforced                  | high     | WARN   | Go: statically typed; 0 matches for `interface{}`. Python: 49 typed function signatures across 8 files in `training/` (mcts.py:13, train.py:7, evaluate.py:8, self_play.py:3, replay_buffer.py:9, model.py:3, board_encoder.py:2, export_onnx.py:4), but no mypy/pyright config.                     |
| SBP-04 | Test infrastructure exists               | critical | PASS   | 44 Go `*_test.go` files under `internal/`; 6 Python `test_*.py` in `training/` (test_board_encoder, test_mcts, test_model, test_self_play, test_train, test_export_onnx). pytest >=9.0.2 in `training/pyproject.toml` dev deps. Total = 50 test files.            |
| SBP-05 | CI/CD pipeline exists                    | high     | WARN   | `.github/workflows/ci.yml` runs `make build` and `make test` on push/PR to main. `.github/workflows/release.yml` handles tagged releases. No lint/format quality gate; Python tests are not run in CI (only Go via `go test ./...`).                   |
| SBP-06 | Error handling patterns are consistent   | high     | PASS   | Go: 603 `if err != nil` sites across 49 files; idiomatic returns (e.g. `internal/engine/fen.go:90-92,162-163`). Only 8 `_ = ` suppressions, mostly legitimate (deferred `Close()` in `internal/bvb/session.go:384,395`, test fixtures). Python: 6 `except` blocks, all typed (FileNotFoundError / Exception with logging + sys.exit or return False); no bare `except:` or silent `pass`.                                         |
| SBP-07 | Dependencies are managed                 | medium   | WARN   | `go.sum` present at repo root; `training/uv.lock` present alongside `pyproject.toml`. No `.github/dependabot.yml` or `renovate.json` — no automated dependency updates. |

## Summary

**Linting:** Strong for Go (golangci-lint v2 with 8 linters), absent for Python (no ruff/mypy/flake8/pylint config).

**Formatting:** Go formatting enforced through golangci-lint but not invoked in CI. No Python formatter.

**Type safety:** Go is statically typed with no `interface{}` usage. Python code has substantial type hints (49 annotated signatures across 8 training modules) but no static type checker configured.

**Tests:** 50 test files total (44 Go + 6 Python). pytest is declared as a dev dependency, and `go test -v ./...` is wired into the Makefile.

**CI stages:** A single `build-and-test` job on Ubuntu running `make build` + `make test` (Go only). Separate release workflow triggered on VERSION file change. No lint, no Python test execution, no coverage, no security scan.

**Dependency management:** Lock files present for both ecosystems (`go.sum`, `training/uv.lock`). No automated dependency update tooling (dependabot/renovate) configured.

### Score calculation

Penalties — SBP-01 (WARN high=1), SBP-02 (WARN medium=0.5), SBP-03 (WARN high=1), SBP-05 (WARN high=1), SBP-07 (WARN medium=0.5). Total = 4.0.

Max weight = 3 (critical) + 4×2 (high) + 2×1 (medium) = 13.

Score = (13 - 4.0) / 13 = 69.2% → Grade **C**.
