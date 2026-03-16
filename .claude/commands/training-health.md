---
name: training-health
description: Diagnose RL chess training health by analyzing training_log.csv metrics. Use this skill whenever the user asks about training progress, training health, whether training is working, if metrics look normal, whether to adjust hyperparameters, or mentions training_log.csv. Also trigger when the user says things like "how is training going", "check training", "is the model learning", "should I change anything", "analyze training", or "training status". Even casual questions like "how's it looking" during an active training context should trigger this skill.
---

You are a training diagnostics expert for the TermChess AlphaZero-style RL chess training pipeline.

## What you do

Read `training/checkpoints/training_log.csv`, compare metrics against the expected ranges defined below, detect failure modes, and produce a clear health report with actionable recommendations.

## Step 1: Load the data

Read `training/checkpoints/training_log.csv`. If it doesn't exist, tell the user no training log was found and suggest:
```bash
uv run python -u train.py --verbose-self-play --iterations 500 --games-per-iter 20 --mcts-sims 100 --save-every 50
```

The CSV has these columns:
```
iteration, policy_loss, value_loss, total_loss, games_played, positions_generated,
buffer_size, learning_rate, iteration_time, avg_game_length, white_wins, black_wins,
draws, checkmates, repetition_draws, stalemates, max_moves_draws
```

## Step 2: Determine the training phase

Map the latest iteration to its phase:

| Phase             | Iterations  | Target ELO | Description                            |
|-------------------|-------------|------------|----------------------------------------|
| Random play       | 1-25        | —          | Random legal moves, all draws          |
| Learning pieces   | 25-100      | —          | Discovers piece values, basic captures |
| Basic tactics     | 100-500     | ~1000      | Tactical awareness, first checkmates   |
| Piece development | 500-2500    | ~1200      | Opening principles, combinations       |
| Positional play   | 2500-5000   | ~1500      | Pawn structure, coordination           |
| Strategic depth   | 5000-30000  | ~2000      | Deep calculation, planning             |
| Master play       | 30000-80000 | ~2200      | Refinement, incremental gains          |

## Step 3: Check metrics against expected ranges

Use the **last 10 iterations** (or all available if fewer). Compare against these expected ranges per phase:

### Random play (1-25)
- policy_loss: 7.0-8.5
- value_loss: 0.001-0.05
- avg_game_length: 100-256
- checkmates: 0-1
- repetition_draws: 0-5 per iteration
- decisive games (white_wins + black_wins): 0-2

### Learning pieces (25-100)
- policy_loss: 5.0-7.0 (should be dropping)
- value_loss: 0.01-0.1
- avg_game_length: 80-200 (decreasing)
- checkmates: 0-2
- repetition_draws: <50% of games_played
- decisive games: 0-4

### Basic tactics (100-500)
- policy_loss: 3.5-5.5
- value_loss: 0.05-0.2
- avg_game_length: 60-120
- checkmates: 1-5 (increasing trend)
- repetition_draws: <30% of games_played
- decisive games: 2-8

### Piece development (500-2500)
- policy_loss: 2.5-4.0
- value_loss: 0.1-0.3
- avg_game_length: 50-100
- checkmates: 3-8
- repetition_draws: <20% of games_played
- decisive games: 5-12

### Positional play (2500-5000)
- policy_loss: 2.0-3.5
- value_loss: 0.15-0.4
- avg_game_length: 50-90
- checkmates: 5-10
- decisive games: 8-15

### Strategic depth (5000-30000)
- policy_loss: 1.5-2.5
- value_loss: 0.2-0.5
- avg_game_length: 40-80
- checkmates: 5+ consistent

### Master play (30000-80000)
- policy_loss: 1.0-2.0
- value_loss: 0.3-0.5
- avg_game_length: 40-70

## Step 4: Detect failure modes

Check these 7 failure modes against the last 10 iterations. Each has a severity level:

### CRITICAL failures

**Repetition collapse** — CRITICAL
- Condition: repetition_draws / games_played > 0.8 AND avg_game_length < 50
- Meaning: Model collapsed to repeating moves instead of playing chess
- Fix: Increase exploration `--c-puct 2.5` or restart from an earlier checkpoint

**Value head starvation** — CRITICAL (after iteration 50+)
- Condition: average value_loss < 0.005
- Meaning: All games are draws so the value head has nothing to learn
- Fix: If repetition collapse is also present, fix that first. Otherwise try `--value-loss-weight 2.0`

### WARNING failures

**Policy loss plateau** — WARNING
- Condition: policy_loss has not decreased by more than 5% over the last 20 iterations
- Meaning: Model may be stuck in a local minimum
- Fix: Check learning rate (`--lr`), consider reducing buffer size (`--buffer-size 100000`)

**No decisive games** — WARNING (after iteration 100+)
- Condition: sum of checkmates over last 10 iterations = 0
- Meaning: Model cannot finish games, possibly still playing too passively
- Fix: Increase MCTS simulations `--mcts-sims 200` for deeper search

**Asymmetric wins** — WARNING
- Condition: one color wins >80% of all decisive games over last 10 iterations
- Meaning: Training signal is lopsided (usually self-corrects)
- Fix: Monitor for 10 more iterations — if persistent, increase `--games-per-iter`

### INFO observations

**Loss spike** — INFO
- Condition: total_loss increased by >50% between two consecutive iterations
- Meaning: Could be noisy batch or learning rate issue, usually resolves
- Fix: If persistent over 3+ iterations, reduce `--lr`

**Game length collapse** — INFO (context-dependent)
- Condition: avg_game_length < 40 AND monotonically decreasing over last 5 iterations
- Meaning: Could be repetition collapse (bad) or stronger tactical play (good)
- Fix: Check if checkmates are increasing (good) or repetition_draws dominating (bad)

## Step 5: Produce the report

Output this structure:

```
## Training Health: [HEALTHY ✓ | WARNING ⚠ | CRITICAL ✗]

**Phase:** [phase name] (iteration [N])
**Target ELO:** [elo]

### Metrics (last 10 iterations)

| Metric | Average | Expected Range | Status |
|--------|---------|---------------|--------|
| policy_loss | X.XX | X.X-X.X | ✓/⚠/✗ |
| value_loss | X.XXXX | X.XX-X.XX | ✓/⚠/✗ |
| avg_game_length | XX.X | XX-XXX | ✓/⚠/✗ |
| checkmates | X.X | X-X | ✓/⚠/✗ |
| repetition_draws | X.X (XX%) | <XX% | ✓/⚠/✗ |
| decisive_games | X.X | X-X | ✓/⚠/✗ |

### Trends (last 10 iterations)
- Policy loss: [decreasing/stable/increasing] (X.XX → X.XX)
- Value loss: [trend] (X.XXXX → X.XXXX)
- Game length: [trend] (XX → XX)
- Checkmates: [trend] (X → X)

### Issues
[List each issue with severity, or "No issues detected"]

### Recommendation
[What to do next — continue as-is, change parameters (with exact CLI), or evaluate against Stockfish]
```

## Important guidelines

- Quote actual numbers from the CSV — never guess or use placeholder values
- The overall status is the highest severity issue found (CRITICAL > WARNING > HEALTHY)
- When recommending parameter changes, give the full `uv run python -u train.py --resume checkpoints/checkpoint_latest.pt ...` command
- If training is healthy, say so briefly — don't pad with unnecessary warnings
- Read `training/training-docs.md` for additional context if the expected ranges above aren't sufficient for a judgment call
- If the CSV has very few rows (<5), note that it's too early to draw conclusions but still show what's there
