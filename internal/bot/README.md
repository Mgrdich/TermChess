# internal/bot

AI opponents. Each implementation satisfies the `Engine` interface in `engine.go`.

## Implementations

| File | Engine | Difficulty |
|------|--------|-----------|
| `random.go` | Weighted random moves | Easy |
| `minimax.go` + `eval.go` | Minimax with alpha-beta pruning | Medium (depth 4) / Hard (depth 7) |
| `rl.go` + `rl_encoder.go` | ONNX-loaded neural net (AlphaZero-style) | RL Beginner/Intermediate/Club/Advanced/Master |

`factory.go` selects the right engine based on the user's difficulty pick.

## ONNX inference status

RL engine is wired (encoder, session interface, move selector, tests with mock session) but `newOnnxSession()` in `rl.go` returns `ErrModelNotLoaded` — the `onnxruntime_go` dependency is not yet in `go.mod`. Tracked as spec `008-custom-rl-agent` Slice 11.

## Cross-language contract

`rl_encoder.go` must stay byte-identical to `training/board_encoder.py`. Both produce `[batch, 18, 8, 8]` float32 tensors. Any channel/layout change must be made in both files in the same commit.
