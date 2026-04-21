# Project Topology — Audit Results

**Date:** 2026-04-21
**Score:** 100% — Grade **A**

## Results

| #   | Check                                | Severity | Status | Evidence                                                                                                                                                                                                                                                   |
| --- | ------------------------------------ | -------- | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Repository structure type            | medium   | PASS   | Monorepo (polyglot). Two independent build roots: `/go.mod` (Go, module `github.com/Mgrdich/TermChess`) and `/training/pyproject.toml` (Python project `training` v0.1.0).                                                                                 |
| 2   | Application layer inventory          | medium   | PASS   | 7 layers detected across Go (binary + internal packages) and Python (ML/RL pipeline). See Topology Summary below.                                                                                                                                          |
| 3   | Database and storage detection       | medium   | PASS   | No database engine. File-based local persistence: TOML config at `~/.termchess/config.toml` (`internal/config/config.go`, BurntSushi/toml dep), FEN save games at `~/.termchess/savegame.fen` (`internal/config/savegame.go`), and PyTorch `.pt` / NumPy `.npz` checkpoints under `training/checkpoints/`. |
| 4   | Infrastructure layer detection       | medium   | SKIP   | No IaC found. No Dockerfile, docker-compose, `*.tf`, K8s manifests, Helm `Chart.yaml`, Pulumi, CDK, CloudFormation, serverless, or Ansible. Only CI/CD GitHub Actions present (`.github/workflows/ci.yml`, `release.yml`), which are not IaC.              |
| 5   | Language inventory                   | medium   | PASS   | Go: 82 `.go` files; Python: 14 `.py` files (training only, excluding `.venv`/`__pycache__`); Shell: 2 `.sh` files (`scripts/install.sh`, release helper). Config/docs: ~39 markdown/TOML/YAML files.                                                       |
| 6   | Inter-layer communication patterns   | medium   | PASS   | No OpenAPI/Swagger/gRPC/GraphQL. Cross-language contract via ONNX model export: `training/export_onnx.py` exports PyTorch ChessNet → `.onnx`; Go side loads it in `internal/bot/rl.go` (currently flagged `ErrModelNotLoaded`: "ONNX runtime integration pending"). Shared tensor contract: `[batch, 18, 8, 8]` board encoding mirrored in `training/board_encoder.py` and `internal/bot/rl_encoder.go`. |

## Topology Summary

- **Structure:** monorepo (polyglot, 2 independent build roots: Go + Python)
- **Layers:**
  - CLI/TUI entry point: Bubbletea (`github.com/charmbracelet/bubbletea`, `bubbles`, `lipgloss`, `muesli/termenv`) at `cmd/termchess/main.go` (primary language: Go)
  - Terminal UI (Bubbletea MVU): framework Bubbletea at `internal/ui/` (primary language: Go)
  - Chess engine (rules, board, FEN, moves, Zobrist hashing): standard library at `internal/engine/` (primary language: Go)
  - Bot / AI opponents (random, minimax + eval, RL/ONNX-backed): internal Go packages at `internal/bot/` (primary language: Go)
  - Bot-vs-Bot session manager / stats / export: internal Go package at `internal/bvb/` (primary language: Go)
  - Self-updater and versioning: internal Go packages at `internal/updater/` and `internal/version/` (primary language: Go)
  - RL training pipeline: PyTorch + `python-chess` + `onnx` / `onnxruntime` / `onnxscript` + `numpy` at `training/` — modules: `board_encoder.py`, `model.py`, `mcts.py`, `self_play.py`, `replay_buffer.py`, `train.py`, `evaluate.py`, `export_onnx.py` (primary language: Python, >=3.12, uv-managed)
  - Shared utilities (clipboard helper): at `internal/util/` (primary language: Go)
  - Config persistence (TOML + FEN savegame): at `internal/config/` (primary language: Go)
- **Storage:** Local filesystem persistence only — TOML config (`~/.termchess/config.toml` via `BurntSushi/toml`), FEN save state (`~/.termchess/savegame.fen`), PyTorch checkpoints (`training/checkpoints/*.pt`), NumPy replay buffer (`training/checkpoints/buffer_latest.npz`), CSV training log (`training_log.csv`). No RDBMS, KV store, or cache service.
- **Infrastructure:** not detected (no IaC; only GitHub Actions CI at `.github/workflows/ci.yml` and `release.yml`)
- **Languages:** Go (82 files), Python (14 files, training), Shell (2 files), Markdown/TOML/YAML configs (~39 files)
- **Communication:** Cross-language integration via ONNX model artifact exchange (Python trains → exports ONNX → Go inference intended). Shared tensor contract `[batch, 18, 8, 8]` implemented twice (Python `board_encoder.py` and Go `internal/bot/rl_encoder.go`). No network APIs, message queues, or schema registries.
- **Service directories:** `cmd/termchess` (Go binary), `internal/engine`, `internal/bot`, `internal/bvb`, `internal/ui`, `internal/config`, `internal/updater`, `internal/util`, `internal/version`, `training` (Python RL pipeline), `scripts` (install helper).
