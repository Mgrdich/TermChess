# Software Best Practices — Audit Results

**Date:** 2026-04-21
**Score:** 100% — Grade **A**

## Results

| # | Check | Severity | Status | Evidence |
| - | ----- | -------- | ------ | -------- |
| 1 | SBP-01 Linting configured and enforced | high | PASS | Go: `.golangci.yml` enables 8 linters (gofmt, goimports, govet, errcheck, staticcheck, unused, ineffassign, misspell); Python: `training/pyproject.toml` configures `[tool.ruff]` + `[tool.mypy]`; both wired via `Makefile` (`lint-go`, `lint-py`, `lint`) and `.github/workflows/ci.yml` (golangci-lint, ruff check, mypy). |
| 2 | SBP-02 Formatting automated | medium | PASS | Go: `gofmt` + `goimports` enforced via `golangci-lint` (CI step "golangci-lint"); Python: `ruff format --check` run locally (`make lint-py`) and in CI (`Ruff format check` step). |
| 3 | SBP-03 Type safety enforced | high | PASS | Go: 0 occurrences of `interface{}` in `internal/`; 71 `any` hits across 27 files — all sampled occurrences are in English-language comments, not type decls. Python: 49 `def …->` typed signatures vs 9 untyped `def` (~85% coverage); mypy configured with `check_untyped_defs=true`, `warn_unused_ignores=true`, `no_implicit_optional=true`, and enforced in CI. |
| 4 | SBP-04 Test infrastructure | critical | PASS | Go test files: 44 (matches CLAUDE.md); Python test files: 6 (matches CLAUDE.md). Runners: `make test` → `go test -v ./...`; `make py-test` → `cd training && uv run pytest`. `[tool.pytest.ini_options]` declared in `training/pyproject.toml`. |
| 5 | SBP-05 CI/CD pipeline | high | PASS | 2 workflows: `ci.yml` (Go: golangci-lint → build → test; Python: ruff check → ruff format check → mypy → pytest) and `release.yml` (tag + multi-platform build + checksums + GitHub Release gated on prior CI success). Both build and test stages present for both languages. |
| 6 | SBP-06 Error handling consistency | high | PASS | Go: 603 `if err != nil` sites; 39 `fmt.Errorf(… %w, err)` wraps across 7 files (idiomatic wrapping in `config/savegame.go` `updater/updater.go`, etc.). Python: 0 bare `except:` blocks; all 7 `except` sites in `training/` use typed exceptions (`FileNotFoundError`, `Exception as e`) with logging/print + propagation or explicit `sys.exit`. |
| 7 | SBP-07 Dependencies managed | medium | PASS | Lock files: `go.sum` (57 lines) and `training/uv.lock` (91k, 689 lines) present. `.github/dependabot.yml` configures weekly updates for gomod (`/`), pip (`/training`), and github-actions — all three ecosystems covered with labels and PR limits. |

## Best Practices Summary
- **Linting:** Go configured (8 linters via golangci-lint); Python configured (ruff + mypy) and enforced in CI.
- **Formatting:** Automated — gofmt/goimports via golangci-lint; ruff format check in Makefile + CI.
- **Types:** Go has no `interface{}`/`any` type abuse; Python ~85% function-signature type coverage with mypy enforced.
- **Tests:** Go 44 / Python 6 test files; both runners wired in Makefile.
- **CI:** 2 workflows (ci.yml + release.yml); CI covers lint, format, type-check, and tests for both languages.
- **Deps:** Lock files present for both ecosystems; dependabot configured for gomod, pip, and github-actions.

## Scoring
- SBP-01 high (2): PASS = 2.0
- SBP-02 medium (1): PASS = 1.0
- SBP-03 high (2): PASS = 2.0
- SBP-04 critical (3): PASS = 3.0
- SBP-05 high (2): PASS = 2.0
- SBP-06 high (2): PASS = 2.0
- SBP-07 medium (1): PASS = 1.0

**Total: 13.0 / 13 = 100% — Grade A**
