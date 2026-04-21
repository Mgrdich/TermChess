# TermChess RL Training Pipeline

AlphaZero-style self-play training pipeline in PyTorch. Trains `ChessNet` (policy + value heads) through MCTS self-play and exports to ONNX for use as the RL bot in the TermChess TUI.

See [`training-docs.md`](./training-docs.md) for the full documentation, or [`CLAUDE.md`](./CLAUDE.md) for AI-assistant context.

## Quick start

```bash
cd training
uv sync
uv run python -u train.py --verbose-self-play --iterations 500 --games-per-iter 20 --mcts-sims 100 --save-every 50
uv run pytest
```

Training metrics stream to `checkpoints/training_log.csv`; use the `/training-health` slash command in Claude Code to diagnose.
