# Project Topology — Audit Results

**Date:** 2026-04-21
**Score:** 100% — Grade **A**

## Results

| #   | Check | Severity | Status | Evidence |
| --- | ----- | -------- | ------ | -------- |
| 1   | TOPO-01 Repository structure type | medium | PASS | Monorepo: two independent build roots — `/Users/mgo/Documents/TermChess/go.mod` (Go) and `/Users/mgo/Documents/TermChess/training/pyproject.toml` (Python/uv). No other build manifests (no `package.json`, `Cargo.toml`, `pom.xml`, `build.gradle*`). |
| 2   | TOPO-02 Application layer inventory | medium | PASS | 3 layers: (a) Go CLI/TUI entrypoint at `cmd/termchess/main.go`; (b) Go internal packages under `internal/` (8 subpackages: `bot`, `bvb`, `config`, `engine`, `ui`, `updater`, `util`, `version`); (c) Python RL training pipeline at `training/` (13 `.py` modules incl. `train.py`, `self_play.py`, `mcts.py`, `model.py`, `board_encoder.py`, `export_onnx.py`, `evaluate.py`). |
| 3   | TOPO-03 Database and storage detection | medium | PASS | Local-filesystem storage only — no DBMS. TOML config at `~/.termchess/config.toml` (`internal/config/config.go`, `internal/config/paths.go`), FEN savegame at `~/.termchess/savegame.fen` (`internal/config/savegame.go`), Python replay buffer NPZ + checkpoints at `training/checkpoints/` (gitignored). Grep for `database|sqlite|postgres|mysql|mongo|redis|migrate` in source returned 0 hits. |
| 4   | TOPO-04 Infrastructure layer detection | medium | PASS | GitHub Actions CI/CD: `.github/workflows/ci.yml` (Go+Python lint/build/test) and `.github/workflows/release.yml` (multi-platform release). `Makefile` orchestrates cross-compile for darwin/linux × amd64/arm64 (`build-all` target). `scripts/install.sh` installer. No Dockerfile, no docker-compose, no Terraform, no k8s, no Helm (0 matches). |
| 5   | TOPO-05 Language inventory | medium | PASS | Go 82 files, Python 14 files (excl. `.venv`, `__pycache__`, `checkpoints/`), Markdown 75 files, Shell 2 files (`scripts/install.sh`, plus one other), TOML 1 (`training/pyproject.toml`). Go + Python are the only runtime languages. |
| 6   | TOPO-06 Inter-layer communication patterns | medium | PASS | Cross-language contract via ONNX tensor I/O: Python exports PyTorch model via `training/export_onnx.py` (422 LOC); Go consumes via `internal/bot/rl.go` (`newOnnxSession()`, currently returns `ErrModelNotLoaded`). Byte-identical board encoder duplicated across `training/board_encoder.py` (191 LOC) and `internal/bot/rl_encoder.go` (118 LOC) producing `[batch,18,8,8]` float32 tensor. No OpenAPI/.proto/GraphQL/message-queue (only `.proto` hits are inside `training/.venv/.../onnx/`, excluded). |

## Topology Summary

- **Structure:** monorepo (2 independent build roots: Go `go.mod`, Python `training/pyproject.toml`)
- **Layers:**
  - CLI/TUI entrypoint: Bubbletea MVU at `cmd/termchess/` (primary language: Go)
  - Domain/application libraries: 8 Go packages at `internal/` — `engine` (chess rules), `bot` (random/minimax/RL engines), `bvb` (bot-vs-bot), `ui` (Bubbletea screens), `config`, `updater`, `util`, `version` (primary language: Go)
  - RL training pipeline: PyTorch + python-chess + ONNX at `training/` (primary language: Python 3.12, managed with `uv`)
- **Storage:** local filesystem only — TOML config + FEN savegame at `~/.termchess/`; Python replay buffer / checkpoints at `training/checkpoints/` (gitignored). No database.
- **Infrastructure:** GitHub Actions (`.github/workflows/ci.yml`, `release.yml`), `Makefile` cross-compile for darwin/linux × amd64/arm64, `scripts/install.sh`. No container/IaC/k8s.
- **Languages:** Go (82 files), Python (14 files), Markdown (75 files), Shell (2 files), TOML (1 file).
- **Communication:** ONNX model file as cross-language artifact; duplicated board-encoder contract (`training/board_encoder.py` ↔ `internal/bot/rl_encoder.go`) producing `[batch,18,8,8]` float32 tensor. Runtime ONNX consumption in Go currently stubbed (`ErrModelNotLoaded`).
- **Service directories:** `cmd/termchess`, `internal/bot`, `internal/bvb`, `internal/config`, `internal/engine`, `internal/ui`, `internal/updater`, `internal/util`, `internal/version`, `training`
