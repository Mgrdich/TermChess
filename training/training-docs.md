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

**Stage 1 — Beginner (target ~1000 ELO)**

```bash
uv run python -u train.py \
  --verbose-self-play \
  --iterations 500 \
  --games-per-iter 20 \
  --mcts-sims 100 \
  --save-every 50
```

At ~1000 ELO the model should avoid blundering pieces, make basic captures, and play legal-looking chess. Evaluate early checkpoints against Stockfish depth 1.

**Stage 2 — Intermediate (target ~1200 ELO)**

```bash
uv run python -u train.py \
  --verbose-self-play \
  --resume checkpoints/checkpoint_500.pt \
  --iterations 2500 \
  --games-per-iter 20 \
  --mcts-sims 100 \
  --save-every 250
```

At ~1200 ELO the model should have basic tactical awareness (pins, forks), develop pieces, and avoid trivial draws. Evaluate against Stockfish depth 1-2.

**Stage 3 — Club Player (target ~1500 ELO)**

```bash
uv run python -u train.py \
  --verbose-self-play \
  --resume checkpoints/checkpoint_2500.pt \
  --iterations 5000 \
  --games-per-iter 30 \
  --mcts-sims 150 \
  --save-every 500
```

**Stage 4 — Advanced (target ~2000 ELO)**

```bash
uv run python -u train.py \
  --verbose-self-play \
  --resume checkpoints/checkpoint_5000.pt \
  --iterations 30000 \
  --games-per-iter 50 \
  --mcts-sims 200 \
  --save-every 1000
```

**Stage 5 — Master (target ~2200 ELO)**

```bash
uv run python -u train.py \
  --verbose-self-play \
  --resume checkpoints/checkpoint_30000.pt \
  --iterations 80000 \
  --games-per-iter 100 \
  --mcts-sims 400
```

You can increase `--games-per-iter` and `--mcts-sims` at later stages since the model benefits more from stronger self-play as it improves.

### Training health indicators

The training loop writes a CSV log to `checkpoints/training_log.csv` with per-iteration metrics. Use this to diagnose training health:

| Metric              | Healthy Sign                                   | Problem Sign                                     |
|---------------------|-------------------------------------------------|--------------------------------------------------|
| `checkmates`        | Increasing over time                            | Stuck at 0 after many iterations                 |
| `repetition_draws`  | Decreasing or low fraction                      | 100% of games end in repetition                  |
| `avg_game_length`   | 40-150 moves, not monotonically decreasing      | Collapsing to <40 moves (repetition collapse)    |
| `value_loss`        | >0.01, meaningfully contributing to total loss   | Near 0 (all games are draws, value head starved) |
| `white_wins/black_wins` | Both >0, roughly balanced                   | Both stuck at 0 (no decisive games)              |

If you see repetition collapse (all games ending in FIVEFOLD_REPETITION with short game lengths), the Dirichlet noise and repetition penalty should help. If it persists, try increasing `--c-puct` (e.g., 2.0-3.0) to encourage more exploration.

### What to expect at each training phase

Use this as a reference when monitoring `checkpoints/training_log.csv`. Numbers are approximate — your run may differ, but the **trends** should match. Iteration ranges below align with the staged training plan above.

#### Iterations 1-25 (Random play)

```
Expected:  policy_loss ~7-8, value_loss ~0.01-0.05, avg_game_length 100-256
           checkmates: 0-1, repetition_draws: 0-5, max_moves_draws: 5-15
           white_wins: 0-1, black_wins: 0-1, draws: 18-20
```

The model plays essentially random legal moves. Games are long and almost all end in draws (max moves, insufficient material, or stalemate). This is normal — the model has no chess knowledge yet.

**What's OK:** All draws, no checkmates, high policy loss.
**Red flag:** If games are already very short (<50 moves) with all repetition draws, Dirichlet noise may not be working.

#### Iterations 25-100 (Learning piece values)

```
Expected:  policy_loss ~5-6 (dropping steadily), value_loss ~0.01-0.1
           avg_game_length: 80-200, gradually decreasing
           checkmates: 0-2 per iteration, draws: 15-20
           repetition_draws: should be <50% of games
```

The model starts learning which pieces are valuable and basic captures. Games get shorter as the model learns to take hanging pieces. Policy loss should drop noticeably.

**What's OK:** Mostly draws still, but some decisive games appearing. Game lengths decreasing.
**Red flag:** Policy loss not decreasing, or avg_game_length dropping below 40 with all repetition draws.

#### Iterations 100-500 (Basic tactics, approaching ~1000 ELO)

```
Expected:  policy_loss ~3.5-5, value_loss ~0.05-0.2
           avg_game_length: 60-120
           checkmates: 1-5 per iteration (increasing trend)
           white_wins + black_wins: 2-8 per iteration
           repetition_draws: <30% of games
```

The model develops basic tactical awareness — it can capture pieces intentionally and starts mating in simple endgames. Value loss should be rising as the model sees more decisive games and the value head gets meaningful training signal.

**Evaluate at iteration 500:** Run against Stockfish depth 1. Target: win rate >30% → ~1000 ELO.

**What's OK:** Mix of decisive games and draws. Checkmates appearing semi-regularly.
**Red flag:** Value loss still near 0, zero checkmates after 250 iterations. Try increasing `--c-puct` to 2.0-3.0.

#### Iterations 500-2500 (Piece development, ~1200 ELO)

```
Expected:  policy_loss ~2.5-3.5, value_loss ~0.1-0.3
           avg_game_length: 50-100
           checkmates: 3-8 per iteration
           white_wins + black_wins: 5-12 per iteration
           repetition_draws: <20% of games
```

The model learns basic opening principles (develop pieces, control center) and can execute simple tactical combinations. Games should have a healthy mix of decisive outcomes and draws.

**Evaluate at iteration 2500:** Run against Stockfish depth 1-2. Target: ~50% score vs depth 1 → ~1200 ELO.

**What's OK:** Roughly balanced white/black wins, checkmates in most iterations.
**Red flag:** One side winning much more than the other (>80% of decisive games).

#### Iterations 2500-5000 (Positional play, ~1500 ELO)

```
Expected:  policy_loss ~2.0-3.0, value_loss ~0.15-0.4
           avg_game_length: 50-90
           checkmates: 5-10 per iteration
           white_wins + black_wins: 8-15 per iteration
```

The model develops positional understanding — pawn structure, piece coordination, king safety. Games should be shorter and more decisive.

**Evaluate at iteration 5000:** Run against Stockfish depth 2-3. Target: ~50% score vs depth 2 → ~1500 ELO.

#### Iterations 5000-30000 (Strategic depth, ~2000 ELO)

```
Expected:  policy_loss ~1.5-2.5, value_loss ~0.2-0.5
           avg_game_length: 40-80
           checkmates: consistent 5+ per iteration
```

Loss will plateau — the model is refining rather than learning new concepts. Increase `--mcts-sims` to 200 and `--games-per-iter` to 50 for higher quality self-play.

**Evaluate at iteration 30000:** Run against Stockfish depth 5. Target: ~50% score → ~2000 ELO.

#### Iterations 30000-80000 (Master play, ~2200 ELO)

```
Expected:  policy_loss ~1.0-2.0, value_loss ~0.3-0.5
           avg_game_length: 40-70
```

Increase `--mcts-sims` to 400 and `--games-per-iter` to 100. Improvements will be incremental.

**Evaluate at iteration 80000:** Run against Stockfish depth 8. Target: ~50% score → ~2200 ELO.

### Troubleshooting guide

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| All games end in FIVEFOLD_REPETITION | Dirichlet noise too low, or model collapsed to repetitive policy | Increase `--c-puct` to 2.0-3.0, or restart from earlier checkpoint |
| Policy loss stuck / not decreasing | Learning rate too low, or buffer too large (stale data dilutes signal) | Decrease `--buffer-size` to 100K, or increase `--lr` |
| Value loss stays near 0 | All games are draws, value head is starved | Check that Dirichlet noise is working (games should be diverse). Increase repetition penalty |
| Policy loss oscillating wildly | Gradient explosions | Decrease `--lr`, gradient clipping should help (already enabled, max_norm=1.0) |
| One side always wins | Asymmetric training signal | Likely fine — it can self-correct as both sides improve. Monitor over 10+ iterations |
| Training very slow per iteration | MCTS simulations are bottleneck | Reduce `--mcts-sims` for early stages (100 is fine for <1500 ELO) |
| Loss suddenly jumps up after resume | Learning rate schedule mismatch or buffer was emptied | Normal — new self-play data from improved model takes time to stabilize |
| avg_game_length increasing after initial decrease | Model entering positional play phase (longer games = more strategic) | Good sign — this is healthy if checkmates are still occurring |

### Resume from checkpoint

Resume training from any saved checkpoint:

```bash
uv run python -u train.py --verbose-self-play --resume checkpoints/checkpoint_5000.pt
```

The checkpoint restores model weights, optimizer state, learning rate schedule, and iteration count. Training continues from where it left off.

### Crash recovery

A `checkpoint_latest.pt` and `buffer_latest.npz` are saved **every iteration**, overwriting the previous version. If training crashes or is interrupted:

```bash
# Resume from the last completed iteration (no lost progress)
uv run python -u train.py --verbose-self-play --resume checkpoints/checkpoint_latest.pt
```

This restores:
- Model weights, optimizer state, and LR schedule from `checkpoint_latest.pt`
- Replay buffer contents from `buffer_latest.npz` (so you don't regenerate all self-play data)
- Iteration count (continues from exactly where it stopped)
- The CSV training log is append-only, so all previous metrics are preserved

The numbered checkpoints (e.g., `checkpoint_500.pt`) are still saved at milestone intervals for ELO evaluation and export. The `checkpoint_latest.pt` is only for crash recovery.

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

Checkpoints are saved at iterations: **10, 25, 50, 100, 250, 500, 1K, 2.5K, 5K, 10K, 30K, 80K** (plus any `--save-every` interval). A CSV training log is also written to `checkpoints/training_log.csv` with per-iteration metrics including game outcomes, termination types, and average game length.

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

The goal is to identify checkpoints that correspond to the target difficulty tiers:

| Target       | ELO  | Suggested approach                                     |
|--------------|------|--------------------------------------------------------|
| Beginner     | 1000 | Evaluate very early checkpoints (50-500) vs depth 1    |
| Intermediate | 1200 | Evaluate early checkpoints (500-2500) vs depth 1-2     |
| Club Player  | 1500 | Evaluate mid checkpoints (2500-5K) vs depth 2-3        |
| Advanced     | 2000 | Evaluate later checkpoints (5K-30K) vs depth 5         |
| Master       | 2200 | Evaluate late checkpoints (30K-80K) vs depth 8         |

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

Once you've identified the right checkpoints via ELO evaluation, export them to ONNX for the Go runtime:

```bash
# Replace checkpoint filenames with the ones that matched each ELO target
uv run python export_onnx.py checkpoints/<best_1000_elo>.pt models/rl_1000.onnx   # Stage 1: iter ~100-500
uv run python export_onnx.py checkpoints/<best_1200_elo>.pt models/rl_1200.onnx   # Stage 2: iter ~500-2500
uv run python export_onnx.py checkpoints/<best_1500_elo>.pt models/rl_1500.onnx   # Stage 3: iter ~2500-5000
uv run python export_onnx.py checkpoints/<best_2000_elo>.pt models/rl_2000.onnx   # Stage 4: iter ~5000-30000
uv run python export_onnx.py checkpoints/<best_2200_elo>.pt models/rl_2200.onnx   # Stage 5: iter ~30000-80000
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
1a. Train to 500   uv run python -u train.py --verbose-self-play --iterations 500 --mcts-sims 100 --save-every 50
1b. Evaluate       uv run python evaluate.py checkpoints/checkpoint_500.pt --stockfish-path stockfish --stockfish-depth 1
1c. Train to 2500  uv run python -u train.py --verbose-self-play --resume checkpoints/checkpoint_500.pt --iterations 2500 --mcts-sims 100 --save-every 250
1d. Evaluate       uv run python evaluate.py checkpoints/checkpoint_2500.pt --stockfish-path stockfish --stockfish-depth 2
1e. Train to 5K    uv run python -u train.py --verbose-self-play --resume checkpoints/checkpoint_2500.pt --iterations 5000 --mcts-sims 150 --save-every 500
1f. Evaluate       uv run python evaluate.py checkpoints/checkpoint_5000.pt --stockfish-path stockfish --stockfish-depth 3
1g. Train to 30K   uv run python -u train.py --verbose-self-play --resume checkpoints/checkpoint_5000.pt --iterations 30000 --mcts-sims 200
1h. Train to 80K   uv run python -u train.py --verbose-self-play --resume checkpoints/checkpoint_30000.pt --iterations 80000 --mcts-sims 400
2.  Export          uv run python export_onnx.py <best_1000_elo>.pt models/rl_1000.onnx
                    uv run python export_onnx.py <best_1200_elo>.pt models/rl_1200.onnx
                    uv run python export_onnx.py <best_1500_elo>.pt models/rl_1500.onnx
                    uv run python export_onnx.py <best_2000_elo>.pt models/rl_2000.onnx
                    uv run python export_onnx.py <best_2200_elo>.pt models/rl_2200.onnx
3.  Integrate       Copy .onnx files to internal/bot/models/ in the Go project
4.  Monitor         Review checkpoints/training_log.csv for training health
```
