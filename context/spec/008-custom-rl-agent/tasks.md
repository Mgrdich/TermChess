# Task List: Custom RL Agent

---

## Phase A: Python Training Pipeline

- [ ] **Slice 1: Board encoder with tests**
  - [ ] Initialize `training/` project with `uv init` and add dependencies (torch, numpy, python-chess, pytest)
  - [ ] Create `training/board_encoder.py` - convert chess position to 18-channel tensor
  - [ ] Create `training/test_board_encoder.py` - verify output shape [18, 8, 8], test known positions
  - [ ] Verify encoder runs on MPS device

- [ ] **Slice 2: Neural network architecture with forward pass**
  - [ ] Create `training/model.py` - ResNet with 6 blocks, 128 filters, dual heads (policy + value)
  - [ ] Create `training/test_model.py` - verify forward pass shapes, policy sums to ~1
  - [ ] Verify model runs on MPS device

- [ ] **Slice 3: MCTS implementation**
  - [ ] Create `training/mcts.py` - Monte Carlo Tree Search with UCB selection
  - [ ] Create `training/test_mcts.py` - verify finds mate-in-1 positions
  - [ ] Integrate neural network for position evaluation

- [ ] **Slice 4: Self-play game generation**
  - [ ] Create `training/self_play.py` - play games using MCTS + neural network
  - [ ] Create `training/replay_buffer.py` - store training examples
  - [ ] Verify can generate 10 self-play games end-to-end

- [ ] **Slice 5: Training loop (minimal)**
  - [ ] Create `training/train.py` - main training loop with MPS support
  - [ ] Implement 1 iteration: generate games → sample batches → train → save checkpoint
  - [ ] Verify training runs for 100 iterations without errors

- [ ] **Slice 6: ONNX export**
  - [ ] Create `training/export_onnx.py` - export PyTorch checkpoint to ONNX
  - [ ] Verify exported model loads in onnxruntime
  - [ ] Verify outputs match between PyTorch and ONNX

- [ ] **Slice 7: ELO evaluation**
  - [ ] Create `training/evaluate.py` - play model vs Stockfish at fixed depth
  - [ ] Estimate ELO from win rate
  - [ ] Document checkpoint → ELO mapping

---

## Phase B: Go Runtime Integration

- [ ] **Slice 8: RL engine skeleton**
  - [ ] Create `internal/bot/rl.go` - implement `rlEngine` struct with `Engine` interface
  - [ ] Add `RLDifficulty` enum (RLIntermediate, RLAdvanced, RLMaster)
  - [ ] Create factory function `NewRLEngine()` returning error (model not yet available)
  - [ ] Add unit tests for factory and interface compliance

- [ ] **Slice 9: ONNX Runtime integration**
  - [ ] Add `github.com/yalue/onnxruntime_go` dependency
  - [ ] Implement model loading from embedded bytes
  - [ ] Create Go board encoder matching Python encoder exactly
  - [ ] Unit test: encoder output matches Python reference

- [ ] **Slice 10: Inference and move selection**
  - [ ] Implement `SelectMove()` - run inference, decode policy, select legal move
  - [ ] Add legal move masking
  - [ ] Unit test with a dummy/test ONNX model

- [ ] **Slice 11: Embed trained models**
  - [ ] Export 1500/2000/2200 ELO models from training
  - [ ] Embed models via `go:embed` in `internal/bot/models/`
  - [ ] Update factory to load correct model based on difficulty
  - [ ] Verify RL bot can play a complete game

---

## Phase C: UI Integration

- [ ] **Slice 12: Add RL bots to selection menu**
  - [ ] Add "RL Intermediate (1500)", "RL Advanced (2000)", "RL Master (2200)" to bot list
  - [ ] Wire selection to `NewRLEngine()` with appropriate difficulty
  - [ ] Handle error case: display message if model unavailable

- [ ] **Slice 13: RL thinking messages**
  - [ ] Create `internal/bot/rl_messages.go` with RL-themed messages
  - [ ] Integrate with existing `getRandomThinkingMessage()` pattern
  - [ ] Verify messages display during RL bot moves

- [ ] **Slice 14: Info page for RL bots**
  - [ ] Add "Press 'i' for info" hint on bot selection screen
  - [ ] Create info page displaying ELO ratings and descriptions
  - [ ] Allow closing info page to return to selection

- [ ] **Slice 15: Bot vs Bot support**
  - [ ] Enable RL bots in Bot vs Bot mode selection
  - [ ] Test combinations: RL vs RL, RL vs Hard, etc.
  - [ ] Verify cleanup on session end
