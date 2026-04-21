# training/ — AI Context

AlphaZero-style self-play training pipeline for the TermChess RL bot. PyTorch + `python-chess` + ONNX export. Managed with `uv`.

## Run commands

```bash
uv sync                                    # install deps
uv run pytest                              # run tests (not in CI — run locally!)
uv run python -u train.py --help           # training CLI
uv run python -u export_onnx.py --help     # export PyTorch checkpoint → ONNX
uv run python -u evaluate.py --help        # evaluate model vs Stockfish
```

## Module map

- `board_encoder.py` — encodes `chess.Board` to `[batch, 18, 8, 8]` float32 tensor. **Must stay byte-identical to `internal/bot/rl_encoder.go`** — this is the cross-language contract.
- `model.py` — `ChessNet` PyTorch module (policy + value heads).
- `mcts.py` — Monte Carlo Tree Search with PUCT, Dirichlet noise at root, virtual loss.
- `self_play.py` — one self-play game driver; returns `(encoded_positions, policies, value_labels)`.
- `replay_buffer.py` — NumPy-backed buffer persisted to `checkpoints/buffer_latest.npz`.
- `train.py` — training loop: self-play → buffer → SGD step → checkpoint → iterate.
- `evaluate.py` — play against Stockfish at configured depth; report ELO estimate.
- `export_onnx.py` — PyTorch → ONNX conversion using `onnxscript`.

## Training output

All under `checkpoints/` (gitignored):

- `model_iter_<N>.pt` — periodic checkpoints per `--save-every`
- `buffer_latest.npz` — replay buffer snapshot
- `training_log.csv` — per-iteration metrics consumed by `/training-health` slash command

## Critical metrics (in `training_log.csv`)

| Column | Healthy range | Failure signal |
|--------|--------------|---------------|
| `policy_loss` | 2.0–8.0 (decreasing) | Stuck >6 after 200 iter → LR too low / policy plateau |
| `value_loss` | >0.01 | Near zero → value head starvation (no decisive games) |
| `checkmates` | Appearing by iter ~100 | Zero after iter 200 → repetition collapse |
| `repetition_draws` | <30% after early training | >50% → repetition collapse — see recent anti-repetition commits |
| `avg_game_length` | 50–150 | <40 with high repetition → collapse; >200 → search too slow |

## ELO targets per stage

| Stage | Iterations | Target ELO | Stockfish eval depth |
|-------|-----------|------------|----------------------|
| Beginner | 0–500 | ~1000 | 1 |
| Intermediate | 500–2500 | ~1200 | 1–2 |
| Club Player | 2500–5000 | ~1500 | 2–3 |
| Advanced | 5000–30000 | ~2000 | 5 |
| Master | 30000–80000 | ~2200 | 8 |

## Gotchas

- **MPS (Apple Silicon) is the default device.** CUDA works but has not been exercised as heavily.
- **Replay buffer reloading:** on restart, `train.py` rehydrates `buffer_latest.npz` — do not delete it unless you want to restart from scratch.
- **Anti-repetition measures:** recent changes (see `git log --oneline training/`) added temperature scheduling and draw-penalty value labels. Do not revert without replicating the training-health evidence.
- **ONNX export invariant:** the exported graph must match the `[18, 8, 8]` input shape. Breaking the encoder in Python without updating `internal/bot/rl_encoder.go` breaks inference silently.
- See `training-docs.md` for the long-form narrative of the pipeline.
