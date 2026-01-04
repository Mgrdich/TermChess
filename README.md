# TermChess

A terminal-based chess application written in Go. Play chess against friends locally or challenge AI opponents of varying difficulty — all from your command line.

[![CI](https://github.com/Mgrdich/TermChess/actions/workflows/ci.yml/badge.svg)](https://github.com/Mgrdich/TermChess/actions/workflows/ci.yml)

## Features

- **Local PvP** — Two players on the same machine
- **Bot Opponents** — Easy, medium, and hard difficulty levels (coming soon)
- **ASCII Board Display** — Clear terminal rendering with coordinates
- **FEN Support** — Save/load positions using standard FEN notation (coming soon)
- **Standard Chess Rules** — Castling, en passant, pawn promotion, checkmate/stalemate detection

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

### Board Display

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

- Uppercase letters = White pieces (P N B R Q K)
- Lowercase letters = Black pieces (p n b r q k)
- Dots = Empty squares

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
│   └── engine/
│       ├── types.go          # Core types (Color, Piece, Square)
│       ├── board.go          # Board state and operations
│       ├── moves.go          # Move generation
│       ├── board_test.go     # Board tests
│       └── moves_test.go     # Move generation tests
├── Makefile
├── go.mod
└── README.md
```

### Architecture

- **CLI Application:** Go + Bubbletea (TUI framework)
- **Chess Engine:** Pure Go implementation
- **Bot Opponents:** Minimax with alpha-beta pruning
- **RL Agent:** ONNX model inference in Go

## Roadmap

- [x] Chess engine foundation (board, pieces, move generation)
- [ ] Check detection and legal move filtering
- [ ] Special moves (castling, en passant, promotion)
- [ ] Game state detection (checkmate, stalemate, draws)
- [ ] FEN import/export
- [ ] Terminal UI with Bubbletea
- [ ] Bot opponents (easy/medium/hard)
- [ ] RL-trained agent

## License

MIT
