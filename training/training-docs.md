# TermChess RL Training Pipeline

AlphaZero-style self-play training pipeline for chess, built with PyTorch and optimized for Apple Silicon (MPS).

## Requirements

- Python 3.12+
- [uv](https://docs.astral.sh/uv/) package manager
- [Stockfish](https://stockfishchess.org/download/) (for ELO evaluation only)
- macOS with Apple Silicon recommended (MPS acceleration)

## Setup

```bash
cd training
uv sync
```

## Architecture

The neural network is a ResNet with dual heads (policy + value), trained via self-play:

| Component       | Details                                                        |
|-----------------|----------------------------------------------------------------|
| Input           | 18 channels x 8x8 (pieces, castling, en passant, side to move) |
| Residual blocks | 6 blocks, 128 filters each                                     |
| Policy head     | 4096 outputs (64 from-squares x 64 to-squares)                 |
| Value head      | 1 output, tanh activation [-1, 1]                              |
| Parameters      | ~2M                                                            |

## Training

### Quick start

```bash
uv run python -u train.py --verbose-self-play
```

Use `-u` for unbuffered output so logs appear in real time. Use `--verbose-self-play` to see per-game progress during self-play.

### Verbose output

With `--verbose-self-play`, you'll see per-game results as they complete:

```
Game 1/20: 1-0 in 87 moves (CHECKMATE) - 42.3s
Game 2/20: 1/2-1/2 in 124 moves (STALEMATE) - 58.1s
...
--- Self-Play Summary ---
Games played: 20
Total positions: 1965
Average game length: 98.2 moves
Results: White +8, Black +7, Draws =5
Total time: 1153.1s (57.7s per game)
```

Without it, you only see a summary line after each full iteration completes.

Use `--quiet` / `-q` to suppress all console output.

### Full configuration

```bash
uv run python -u train.py \
  --verbose-self-play \
  --iterations 80000 \
  --games-per-iter 100 \
  --batch-size 256 \
  --mcts-sims 400 \
  --num-blocks 6 \
  --num-filters 128
```

### Recommended staged training

Training all 80K iterations in one run takes a very long time. A practical approach is to train in stages, resuming from each checkpoint:

**Stage 1 — Intermediate (target ~1500 ELO)**

```bash
uv run python -u train.py \
  --verbose-self-play \
  --iterations 5000 \
  --games-per-iter 20 \
  --mcts-sims 100 \
  --save-every 500
```

Evaluate, then continue:

**Stage 2 — Advanced (target ~2000 ELO)**

```bash
uv run python -u train.py \
  --verbose-self-play \
  --resume checkpoints/checkpoint_5000.pt \
  --iterations 30000 \
  --games-per-iter 50 \
  --mcts-sims 200 \
  --save-every 1000
```

**Stage 3 — Master (target ~2200 ELO)**

```bash
uv run python -u train.py \
  --verbose-self-play \
  --resume checkpoints/checkpoint_30000.pt \
  --iterations 80000 \
  --games-per-iter 100 \
  --mcts-sims 400
```

You can increase `--games-per-iter` and `--mcts-sims` at later stages since the model benefits more from stronger self-play as it improves.

### Resume from checkpoint

Resume training from any saved checkpoint:

```bash
uv run python -u train.py --verbose-self-play --resume checkpoints/checkpoint_5000.pt
```

The checkpoint restores model weights, optimizer state, learning rate schedule, and iteration count. Training continues from where it left off.

### Training parameters

| Parameter              | Flag                   | Default                  |
|------------------------|------------------------|--------------------------|
| Iterations             | `--iterations`         | 80,000                   |
| Games per iteration    | `--games-per-iter`     | 100                      |
| Batch size             | `--batch-size`         | 256                      |
| Batches per iteration  | `--batches-per-iter`   | 10                       |
| Replay buffer size     | `--buffer-size`        | 500K                     |
| MCTS simulations/move  | `--mcts-sims`          | 400                      |
| Learning rate          | `--lr` / `--lr-final`  | 0.001 -> 0.0001 (decay)  |
| Optimizer              |                        | Adam (weight decay 1e-4) |
| Verbose self-play      | `--verbose-self-play`  | off                      |
| Quiet mode             | `--quiet` / `-q`       | off                      |
| Save interval          | `--save-every`         | 0 (disabled)             |
| Resume                 | `--resume`             | none                     |

Checkpoints are saved at iterations: **5K, 10K, 30K, 80K** (plus any `--save-every` interval).

## ELO Evaluation

After training, evaluate checkpoints against Stockfish to estimate their rating.

### Evaluate a single checkpoint

```bash
uv run python evaluate.py checkpoints/checkpoint_5000.pt \
  --stockfish-path /path/to/stockfish
```

### Evaluate multiple checkpoints

```bash
uv run python evaluate.py checkpoints/checkpoint_*.pt \
  --stockfish-path /path/to/stockfish \
  --num-games 20 \
  --stockfish-depth 5
```

### Evaluation options

| Flag                     | Default     | Description                      |
|--------------------------|-------------|----------------------------------|
| `--stockfish-path`       | `stockfish` | Path to Stockfish binary         |
| `--num-games`            | 20          | Games per checkpoint             |
| `--stockfish-depth`      | 5           | Stockfish search depth           |
| `--stockfish-time-limit` | 1.0         | Stockfish seconds per move       |
| `--mcts-simulations`     | 200         | MCTS simulations for model moves |

### Stockfish depth to approximate ELO

| Depth | ~ELO |
|-------|------|
| 1     | 1300 |
| 2     | 1500 |
| 3     | 1700 |
| 5     | 2000 |
| 8     | 2300 |
| 10    | 2500 |
| 15    | 2800 |
| 20    | 3000 |

### ELO estimation method

The script uses the inverse ELO formula:

```
score = (wins + 0.5 * draws) / total_games
elo_diff = -400 * log10(1/score - 1)
estimated_elo = stockfish_elo + elo_diff
```

A 95% confidence interval is computed using the Wilson score interval.

### Checkpoint to ELO mapping

The goal is to identify checkpoints that correspond to the three target difficulty tiers:

| Target       | ELO  | Suggested approach                                  |
|--------------|------|-----------------------------------------------------|
| Intermediate | 1500 | Evaluate early checkpoints (5K) against depth 2-3   |
| Advanced     | 2000 | Evaluate mid checkpoints (10K-30K) against depth 5  |
| Master       | 2200 | Evaluate late checkpoints (30K-80K) against depth 8 |

Example output when evaluating multiple checkpoints:

```
====================================================
              Checkpoint -> ELO Mapping
====================================================
  checkpoint_5000.pt   | W:8   D:4   L:8   | Score:  50.0% | ELO: ~2000 +/- 150
  checkpoint_10000.pt  | W:10  D:5   L:5   | Score:  62.5% | ELO: ~2088 +/- 130
  checkpoint_30000.pt  | W:14  D:3   L:3   | Score:  77.5% | ELO: ~2210 +/- 110
====================================================
```

## ONNX Export

Once you've identified the right checkpoints, export them to ONNX for the Go runtime:

```bash
uv run python export_onnx.py checkpoints/checkpoint_5000.pt models/rl_1500.onnx
uv run python export_onnx.py checkpoints/checkpoint_30000.pt models/rl_2000.onnx
uv run python export_onnx.py checkpoints/checkpoint_80000.pt models/rl_2200.onnx
```

The exported `.onnx` files are then embedded into the Go binary via `go:embed`.

## Testing

```bash
uv run pytest
```

## Module overview

| File               | Purpose                                    |
|--------------------|--------------------------------------------|
| `train.py`         | Main training loop                         |
| `model.py`         | ChessNet neural network architecture       |
| `mcts.py`          | Monte Carlo Tree Search with UCB selection |
| `board_encoder.py` | Convert board state to 18-channel tensor   |
| `self_play.py`     | Generate self-play games                   |
| `replay_buffer.py` | Store and sample training examples         |
| `evaluate.py`      | ELO estimation against Stockfish           |
| `export_onnx.py`   | Export PyTorch checkpoint to ONNX          |

## End-to-end workflow

```
1. Train        uv run python -u train.py --verbose-self-play --save-every 500
2. Evaluate     uv run python evaluate.py checkpoints/checkpoint_*.pt --stockfish-path stockfish
3. Export       uv run python export_onnx.py <best_checkpoint>.pt model.onnx
4. Integrate    Copy .onnx files to internal/bot/models/ in the Go project
```

To train in stages with resume:

```
1a. Train to 5K    uv run python -u train.py --verbose-self-play --iterations 5000 --mcts-sims 100 --save-every 500
1b. Evaluate       uv run python evaluate.py checkpoints/checkpoint_5000.pt --stockfish-path stockfish
1c. Train to 30K   uv run python -u train.py --verbose-self-play --resume checkpoints/checkpoint_5000.pt --iterations 30000 --mcts-sims 200
1d. Evaluate       uv run python evaluate.py checkpoints/checkpoint_30000.pt --stockfish-path stockfish
1e. Train to 80K   uv run python -u train.py --verbose-self-play --resume checkpoints/checkpoint_30000.pt --iterations 80000 --mcts-sims 400
1f. Evaluate       uv run python evaluate.py checkpoints/checkpoint_80000.pt --stockfish-path stockfish
2.  Export          uv run python export_onnx.py checkpoints/checkpoint_5000.pt models/rl_1500.onnx
                    uv run python export_onnx.py checkpoints/checkpoint_30000.pt models/rl_2000.onnx
                    uv run python export_onnx.py checkpoints/checkpoint_80000.pt models/rl_2200.onnx
3.  Integrate       Copy .onnx files to internal/bot/models/ in the Go project
```
