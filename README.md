# TermChess

A terminal-based chess application written in Go. Play chess against friends locally or challenge AI opponents of varying difficulty — all from your command line.

[![CI](https://github.com/Mgrdich/TermChess/actions/workflows/ci.yml/badge.svg)](https://github.com/Mgrdich/TermChess/actions/workflows/ci.yml)

> **100% AI-Generated Project**
>
> This entire project — every line of code, all specifications, tests, and documentation — was created using [AWOS (Agentic Way of Software)](https://github.com/provectus/awos). Not a single line of manual code was written.

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

## Installation

### From Source

Requires Go 1.21 or later.

```bash
git clone https://github.com/Mgrdich/TermChess.git
cd TermChess
make build
```

The binary will be created at `bin/termchess`.

## Usage

```bash
# Run the application
make run

# Or run the built binary
./bin/termchess
```

The application features a full interactive menu system:
- **Main Menu** — New game, load game from FEN, resume saved game, settings, exit
- **Game Types** — Player vs Player (local), Bot Opponents (easy/medium/hard)
- **Gameplay** — Enter moves using SAN notation (e4, Nf3, Bxc5, O-O, etc.)
- **Commands** — Type `resign`, `offerdraw`, `showfen`, or `menu` during gameplay
- **Navigation** — Use arrow keys or j/k, press ESC to go back, Ctrl+C to exit

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

### Configuration

Settings are saved to `~/.termchess/config.toml` and include:
- **Use Unicode Pieces** — Display board with Unicode chess symbols
- **Show Coordinates** — Display file/rank labels around board
- **Use Colors** — Color pieces for better visibility
- **Show Move History** — Display move list during gameplay
- **Show Help Text** — Display navigation hints on each screen

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
│   ├── ui/                   # Terminal UI (Bubbletea)
│   │   ├── model.go          # Application state
│   │   ├── view.go           # Screen rendering
│   │   ├── update.go         # Event handling
│   │   ├── board.go          # Board rendering
│   │   ├── san.go            # SAN move parsing
│   │   ├── save.go           # Game save/load
│   │   └── *_test.go         # UI tests (83.5% coverage)
│   └── util/                 # Utilities
│       └── clipboard.go      # Cross-platform clipboard
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

### In Progress / Planned 🚧
- [ ] RL-trained agent
- [ ] Opening book integration
- [ ] PGN import/export
- [ ] Time controls

## License

MIT
