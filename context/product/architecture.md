# System Architecture Overview: TermChess

---

## 1. Application & Technology Stack

- **CLI Application:** Go + Bubbletea — Terminal UI, menus, keyboard input, board rendering
- **Chess Engine:** Go — Board state, move validation, all chess rules, FEN import/export
- **Built-in Bots:** Go — Easy/Medium/Hard using minimax with alpha-beta pruning
- **RL Model Inference:** Go + ONNX Runtime — Load trained .onnx model directly in Go
- **RL Training:** Python + PyTorch — Offline training, exports to ONNX
- **External Engines:** UCI Protocol (stdin/stdout) — Stockfish, Komodo, etc.

### Engine Abstraction

Go interface pattern for bot/engine interoperability:

```
Engine interface {
    GetMove(position) -> move
}

├── InternalEngine (Go bots + ONNX RL agent)
└── UCIEngine (external engines via stdin/stdout)
```

**Result:** Single Go binary for the application. Python only used offline for training.

---

## 2. Data & Persistence

- **User Config:** File at `~/.termchess/config.toml` — Board style, default difficulty, preferences
- **Saved Games:** File(s) at `~/.termchess/saves/` — FEN strings with game names
- **RL Model:** Bundled with binary via Go `embed` package — `.onnx` format

### Storage Structure

```
~/.termchess/
├── config.toml          # user preferences
└── saves/
    └── games.toml       # saved positions with names
```

No database — simple file-based persistence read/written directly by the CLI.

---

## 3. Project Structure & Build

### Directory Layout

```
termchess/
├── cmd/
│   └── termchess/
│       └── main.go              # Entry point
├── internal/
│   ├── engine/
│   │   ├── board.go             # Board representation
│   │   ├── moves.go             # Move generation & validation
│   │   ├── fen.go               # FEN parsing/export
│   │   └── rules.go             # Chess rules, game state detection
│   ├── bot/
│   │   ├── engine.go            # Engine interface
│   │   ├── random.go            # Easy bot
│   │   ├── minimax.go           # Medium/Hard bots
│   │   ├── rl.go                # ONNX RL agent
│   │   └── uci.go               # UCI engine adapter
│   ├── ui/
│   │   ├── app.go               # Bubbletea app
│   │   ├── board.go             # Board rendering
│   │   └── menu.go              # Menus & navigation
│   └── config/
│       ├── config.go            # Config loading/saving
│       └── saves.go             # Game save/load
├── training/                     # Python RL training (separate, managed by uv)
│   ├── pyproject.toml
│   ├── train.py
│   ├── model.py
│   └── export_onnx.py
├── go.mod
├── go.sum
└── Makefile
```

### Build Tools

- **Go Modules:** Dependency management for Go
- **Makefile:** Build, test, run commands
- **uv:** Python package management for RL training (fast, deterministic lockfile)
- **Model Distribution:** Bundled via Go `embed` package — single self-contained binary

---

## 4. Testing Strategy

- **Chess Engine:** Unit tests — move validation, FEN parsing, checkmate detection
- **Bots:** Unit tests + known position tests — "mate in 1" scenarios, tactical puzzles
- **UCI Adapter:** Integration tests with mock engine
- **UI:** Manual testing (Bubbletea has limited test tooling)

---

## 5. Key Design Decisions

| Decision           | Choice                           | Rationale                                               |
|--------------------|----------------------------------|---------------------------------------------------------|
| Language split     | Go (runtime) / Python (training) | Single binary distribution; Python only for ML training |
| Chess logic        | Native Go implementation         | No external dependencies at runtime; full control       |
| Bot communication  | Engine interface abstraction     | Unified API for internal bots and external UCI engines  |
| RL inference       | ONNX Runtime in Go               | No Python dependency at runtime; fast inference         |
| Persistence        | File-based (TOML)                | Simple, human-readable, no database overhead            |
| Model distribution | Embedded in binary               | Single-file distribution; no separate downloads         |
