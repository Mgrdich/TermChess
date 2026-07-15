# TermChess

A terminal-based chess application written in Go. Play chess against friends locally or challenge AI opponents of varying difficulty — all from your command line.

[![CI](https://github.com/Mgrdich/TermChess/actions/workflows/ci.yml/badge.svg)](https://github.com/Mgrdich/TermChess/actions/workflows/ci.yml)

> **100% AI-Generated Project**
>
> This entire project — every line of code, all specifications, tests, and documentation — was created using [AWOS (Agentic Way of Software)](https://github.com/provectus/awos). Not a single line of manual code was written.

## Two implementations (Go & Rust)

TermChess ships as **two independent build roots that produce the same terminal chess TUI** and are kept in sync:

| Implementation | Build root | Toolchain | Build | Run | Test |
|----------------|-----------|-----------|-------|-----|------|
| **Go** (original) | `go.mod` (repo root) | Go 1.21+ | `make build-go` | `make run-go` | `make test` |
| **Rust** (port) | `rust/` Cargo workspace | cargo 1.95+ | `make build-rust` | `make run-rust` | `make test-rust` |

Both implementations share the same features, screens, CLI flags, and on-disk config/savegame format. The Rust port lives entirely under `rust/` and never touches the Go tree. For the full crate ↔ Go-package mapping, the architectural changes, and current parity notes, see [docs/MIGRATION.md](docs/MIGRATION.md).

> `make build` (with no suffix) remains an alias for the Go build so existing tooling keeps working.

## Features

- **Interactive Terminal UI** — Full-featured TUI built with Bubbletea
- **Local PvP** — Two players on the same machine
- **SAN Move Input** — Enter moves using standard algebraic notation (e4, Nf3, O-O, etc.)
- **Board Rendering** — ASCII and Unicode display options with configurable colors
- **FEN Support** — Save/load positions using standard FEN notation
- **Game Management** — Auto-save on exit, resume games, settings persistence
- **Standard Chess Rules** — Castling, en passant, pawn promotion, checkmate/stalemate detection
- **Draw System** — Draw offers, resignation, automatic draw detection
- **Move History** — Optional move list display in SAN format
- **Bot Opponents** — AI players with easy, medium, and hard difficulty levels
- **Bot vs Bot Mode** — Watch AI opponents battle each other with configurable speed

## Installation

### Quick Install (Recommended)

Install TermChess with a single command:

```bash
curl -fsSL https://raw.githubusercontent.com/Mgrdich/TermChess/main/scripts/install.sh | bash
```

To install a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/Mgrdich/TermChess/main/scripts/install.sh | bash -s -- v1.0.0
```

> **Why curl install is recommended:** Installing via the curl script or manual download enables the built-in `--upgrade` and `--uninstall` commands, making it easy to keep TermChess up to date.

### Manual Download

If you prefer not to pipe to bash, download the binary directly:

**Available platforms:**

| Platform              | Binary Name                     |
|-----------------------|---------------------------------|
| macOS (Apple Silicon) | `termchess-vX.X.X-darwin-arm64` |
| macOS (Intel)         | `termchess-vX.X.X-darwin-amd64` |
| Linux (x86_64)        | `termchess-vX.X.X-linux-amd64`  |
| Linux (ARM64)         | `termchess-vX.X.X-linux-arm64`  |

**Option 1: Download from browser**
1. Go to the [Releases page](https://github.com/Mgrdich/TermChess/releases)
2. Download the binary for your platform
3. Make it executable and move to your PATH

**Option 2: Download with curl**
```bash
# Set the version and platform
VERSION="v0.1.0"
OS="darwin"      # or "linux"
ARCH="arm64"     # or "amd64"

# Download the binary
curl -fsSL "https://github.com/Mgrdich/TermChess/releases/download/${VERSION}/termchess-${VERSION}-${OS}-${ARCH}" -o termchess

# Make executable and install
chmod +x termchess
mv termchess ~/.local/bin/
```

**Verify checksum (optional but recommended):**
```bash
# Download checksums file
curl -fsSL "https://github.com/Mgrdich/TermChess/releases/download/${VERSION}/checksums.txt" -o checksums.txt

# Verify (on macOS)
shasum -a 256 termchess | grep -f checksums.txt

# Verify (on Linux)
sha256sum termchess | grep -f checksums.txt
```

### From Source

Requires Go 1.21 or later.

```bash
git clone https://github.com/Mgrdich/TermChess.git
cd TermChess
make build
```

The binary will be created at `bin/termchess`.

### Upgrading

The `--upgrade` command works when TermChess is installed via the **curl script** or **manual download** to `~/.local/bin` or `/usr/local/bin`.

To upgrade to the latest version:

```bash
termchess --upgrade
```

To upgrade (or downgrade) to a specific version:

```bash
termchess --upgrade v1.0.0
```

> **Note:** If you installed via `go install`, use `go install github.com/Mgrdich/TermChess/cmd/termchess@latest` to upgrade instead.

### Uninstalling

To remove TermChess and its configuration:

```bash
termchess --uninstall
```

### Version Information

To check your installed version:

```bash
termchess --version
```

## Usage

```bash
# Run the application
make run

# Or run the built binary
./bin/termchess
```

The application features a full interactive menu system:
- **Main Menu** — New game, load game from FEN, resume saved game, settings, exit
- **Game Types** — Player vs Player (local), Player vs Bot, Bot vs Bot
- **Gameplay** — Enter moves using SAN notation (e4, Nf3, Bxc5, O-O, etc.)
- **Commands** — Type `resign`, `offerdraw`, `showfen`, or `menu` during gameplay
- **Navigation** — Use arrow keys or j/k, press ESC to go back, Ctrl+C to exit

**Main Menu:**
```
TermChess

> New Game
  Load from FEN
  Resume Game
  Settings
  Exit

↑/↓: navigate | Enter: select
```

**Game Type Selection:**
```
TermChess

Select Game Type:

> Player vs Player
  Player vs Bot
  Bot vs Bot

↑/↓: navigate | Enter: select | ESC: back
```

### Board Display

The board can be displayed in ASCII or Unicode mode (configurable in Settings):

**ASCII Mode:**
```
8 r n b q k b n r
7 p p p p p p p p
6 . . . . . . . .
5 . . . . . . . .
4 . . . . . . . .
3 . . . . . . . .
2 P P P P P P P P
1 R N B Q K B N R
  a b c d e f g h
```

**Unicode Mode:**
```
8 ♜ ♞ ♝ ♛ ♚ ♝ ♞ ♜
7 ♟ ♟ ♟ ♟ ♟ ♟ ♟ ♟
6 · · · · · · · ·
5 · · · · · · · ·
4 · · · · · · · ·
3 · · · · · · · ·
2 ♙ ♙ ♙ ♙ ♙ ♙ ♙ ♙
1 ♖ ♘ ♗ ♕ ♔ ♗ ♘ ♖
  a b c d e f g h
```

### Bot vs Bot Mode

Watch two AI opponents play against each other:

1. Select **Bot vs Bot** from the main menu
2. Choose difficulty for the White bot (Easy, Medium, or Hard)
3. Choose difficulty for the Black bot
4. Select Single Game or Multi-Game mode
5. Watch the game unfold automatically

**Example Bot vs Bot display:**
```
TermChess - Bot vs Bot

Easy Bot (White) vs Hard Bot (Black)
Game 1/1 | 15 moves

8 ♜ · ♝ · ♚ ♝ · ♜
7 ♟ ♟ ♟ · · ♟ ♟ ♟
6 · · ♞ ♟ · ♞ · ·
5 · · · · ♟ · · ·
4 · · ♗ · ♙ · · ·
3 · · · · · ♘ · ·
2 ♙ ♙ ♙ ♙ · ♙ ♙ ♙
1 ♖ ♘ ♗ ♕ ♔ · · ♖
  a b c d e f g h

White to move | Speed: Normal

Space: pause | 1-4: speed | Tab: view | ESC: abort
```

**Controls during Bot vs Bot games:**
- **Space** — Pause/resume the game
- **1-4** — Change playback speed (1=Instant, 2=Fast, 3=Normal, 4=Slow)
- **Tab** — Toggle between single board and grid view (multi-game)
- **←/→** — Navigate between games (multi-game mode)
- **f** — Show current position FEN
- **ESC** — Abort and return to menu

**Multi-Game Mode:**
Run multiple games and view them in a grid layout. Games are queued and executed 50 at a time to maintain UI responsiveness. The status bar shows completed, running, and queued game counts. After all games complete, see detailed statistics including win rates, average game length, and individual game results.

### Bot Difficulty Levels

| Difficulty | Engine | Search Depth | Time Limit | Description |
|------------|--------|--------------|------------|-------------|
| Easy       | Random | N/A          | 2s         | Weighted random moves, beatable by beginners |
| Medium     | Minimax | 4           | 4s         | Alpha-beta pruning, finds basic tactics |
| Hard       | Minimax | 7           | 8s         | Deeper search, finds complex tactics |

Hard bot consistently beats Medium in automated testing due to its 3-ply depth advantage.

### Configuration

Settings are saved to `~/.termchess/config.toml` and include:
- **Use Unicode Pieces** — Display board with Unicode chess symbols
- **Show Coordinates** — Display file/rank labels around board
- **Use Colors** — Color pieces for better visibility
- **Show Move History** — Display move list during gameplay
- **Show Help Text** — Display navigation hints on each screen
- **Bot Move Delay** — Adjust speed of bot moves in Bot vs Bot mode

## Development

### Prerequisites

- Go 1.21+
- Make

### Commands

```bash
make build    # Build the binary
make test     # Run all tests
make run      # Run the application
make clean    # Remove build artifacts
```

### Project Structure

```
termchess/
├── cmd/
│   └── termchess/
│       └── main.go           # Entry point
├── internal/
│   ├── config/               # Configuration management
│   │   ├── config.go         # Load/save user preferences
│   │   └── config_test.go
│   ├── engine/               # Chess engine
│   │   ├── types.go          # Core types (Color, Piece, Square)
│   │   ├── board.go          # Board state and operations
│   │   ├── moves.go          # Move generation and validation
│   │   ├── fen.go            # FEN import/export
│   │   ├── game_state.go     # Game status detection
│   │   ├── attacks.go        # Attack calculations
│   │   ├── zobrist.go        # Position hashing
│   │   └── *_test.go         # Comprehensive test suite
│   ├── bot/                  # Bot engine implementations
│   │   ├── engine.go         # Engine interface
│   │   ├── random.go         # Easy bot (random moves)
│   │   ├── minimax.go        # Medium/Hard bot (minimax + alpha-beta)
│   │   └── eval.go           # Position evaluation
│   ├── bvb/                  # Bot vs Bot game management
│   │   ├── session.go        # Single-game controller
│   │   ├── manager.go        # Multi-game queue
│   │   ├── stats.go          # Aggregate statistics
│   │   └── export.go         # Game export
│   ├── ui/                   # Terminal UI (Bubbletea)
│   │   ├── model.go          # Application state
│   │   ├── view.go           # Screen rendering
│   │   ├── update.go         # Event handling
│   │   ├── board.go          # Board rendering
│   │   ├── san.go            # SAN move parsing
│   │   ├── save.go           # Game save/load
│   │   └── *_test.go         # UI tests (83.5% coverage)
│   ├── updater/              # Self-upgrade via GitHub Releases
│   │   └── updater.go
│   ├── version/              # Build-time version metadata
│   │   └── version.go
│   └── util/                 # Cross-cutting utilities
│       └── clipboard.go      # Cross-platform clipboard
├── training/                 # Python RL training pipeline (uv-managed)
├── Makefile
├── go.mod
└── README.md
```

### Architecture

- **CLI Application:** Go + [Bubbletea](https://github.com/charmbracelet/bubbletea) (TUI framework)
- **Chess Engine:** Pure Go implementation with Zobrist hashing
- **Configuration:** TOML-based persistent settings
- **Save System:** FEN-based game state persistence
- **Testing:** 83.5% test coverage on UI, comprehensive engine tests

## Roadmap

### Completed ✅
- [x] Chess engine foundation (board, pieces, move generation)
- [x] Check detection and legal move filtering
- [x] Special moves (castling, en passant, promotion)
- [x] Game state detection (checkmate, stalemate, draws)
- [x] FEN import/export
- [x] Terminal UI with Bubbletea
- [x] SAN move input parsing
- [x] Game save/resume functionality
- [x] Settings and configuration management
- [x] Move history display
- [x] Draw offers and resignation
- [x] Bot opponents (easy/medium/hard)
- [x] Bot vs Bot spectator mode
- [x] CLI distribution (install script, self-upgrade, self-uninstall)
- [x] RL-trained agent — encoder, inference interface, difficulty tiers, training pipeline

### In Progress / Planned 🚧
- [ ] RL ONNX Runtime integration in Go (spec 008 Slice 11 — consumes trained models at runtime)
- [ ] Opening book integration
- [ ] PGN import/export
- [ ] Time controls

## RL Training Pipeline

The `training/` directory contains an AlphaZero-style self-play training pipeline for chess, built with PyTorch and optimized for Apple Silicon (MPS). It trains a neural network through self-play using MCTS, then exports to ONNX for use as a bot opponent in the game.

See [training/training-docs.md](training/training-docs.md) for full documentation.

### Quick start

```bash
cd training
uv sync
uv run python -u train.py --verbose-self-play --iterations 500 --games-per-iter 20 --mcts-sims 100 --save-every 50
```

### Monitoring training health

Training writes per-iteration metrics to `training/checkpoints/training_log.csv`. If you use [Claude Code](https://claude.ai/claude-code), the `/training-health` slash command analyzes this log and reports whether training is progressing normally:

```
/training-health
```

It compares metrics against expected ranges for each training phase, detects failure modes (repetition collapse, value head starvation, policy plateau, etc.), and suggests specific parameter changes if something is off.

You can also monitor the CSV directly — key columns to watch:

| Column | What it tells you |
|--------|-------------------|
| `policy_loss` | Should decrease over time (8.0 → 2.0) |
| `value_loss` | Should be >0.01 (means decisive games are happening) |
| `checkmates` | Should appear by iteration ~100 and increase |
| `repetition_draws` | Should be <30% of games after early training |
| `avg_game_length` | 50-150 is healthy; <40 with high repetition is a problem |

### Training stages and ELO targets

| Stage | Iterations | Target ELO | Stockfish Eval Depth |
|-------|-----------|------------|---------------------|
| Beginner | 0-500 | ~1000 | depth 1 |
| Intermediate | 500-2500 | ~1200 | depth 1-2 |
| Club Player | 2500-5000 | ~1500 | depth 2-3 |
| Advanced | 5000-30000 | ~2000 | depth 5 |
| Master | 30000-80000 | ~2200 | depth 8 |

## License

MIT
