# Task List: Custom RL Agent

---

## Phase A: Python Training Pipeline

- [x] **Slice 1: Board encoder with tests**
  - [x] Initialize `training/` project with `uv init` and add dependencies (torch, numpy, python-chess, pytest)
  - [x] Create `training/board_encoder.py` - convert chess position to 66-channel tensor (18 current + 4x12 history)
  - [x] Create `training/test_board_encoder.py` - verify output shape [66, 8, 8], test known positions, test history encoding
  - [x] Verify encoder runs on MPS device

- [x] **Slice 2: Neural network architecture with forward pass**
  - [x] Create `training/model.py` - ResNet with 6 blocks, 128 filters, dual heads (policy + value)
  - [x] Create `training/test_model.py` - verify forward pass shapes, policy sums to ~1
  - [x] Verify model runs on MPS device

- [x] **Slice 3: MCTS implementation**
  - [x] Create `training/mcts.py` - Monte Carlo Tree Search with UCB selection
  - [x] Create `training/test_mcts.py` - verify finds mate-in-1 positions
  - [x] Integrate neural network for position evaluation

- [x] **Slice 4: Self-play game generation**
  - [x] Create `training/self_play.py` - play games using MCTS + neural network
  - [x] Create `training/replay_buffer.py` - store training examples
  - [x] Verify can generate 10 self-play games end-to-end

- [x] **Slice 5: Training loop (full)**
  - [x] Create `training/train.py` - main training loop with MPS support
  - [x] Implement iteration: generate games → sample batches → train → save checkpoint
  - [x] Dirichlet noise at MCTS root for exploration
  - [x] Repetition draw penalty (-0.2 value target)
  - [x] Gradient clipping (max_norm=1.0)
  - [x] Configurable value loss weight
  - [x] Horizontal flip data augmentation (50% during sampling)
  - [x] Move history encoding (last 4 positions)
  - [x] Per-iteration CSV metrics log (`training_log.csv`)
  - [x] Auto-save `checkpoint_latest.pt` + `buffer_latest.npz` every iteration (crash recovery)
  - [x] Replay buffer save/load for resume
  - [x] Verify training runs for 100 iterations without errors

- [x] **Slice 6: ONNX export**
  - [x] Create `training/export_onnx.py` - export PyTorch checkpoint to ONNX
  - [x] Verify exported model loads in onnxruntime
  - [x] Verify outputs match between PyTorch and ONNX

- [x] **Slice 7: ELO evaluation**
  - [x] Create `training/evaluate.py` - play model vs Stockfish at fixed depth
  - [x] Estimate ELO from win rate
  - [x] Document checkpoint → ELO mapping

---

## Phase B: Go Runtime Integration

- [x] **Slice 8: RL engine skeleton**
  - [x] Create `internal/bot/rl.go` - implement `rlEngine` struct with `Engine` interface
  - [x] Add `RLDifficulty` enum (RLIntermediate, RLAdvanced, RLMaster)
  - [x] Create factory function `NewRLEngine()` returning error (model not yet available)
  - [x] Add unit tests for factory and interface compliance

- [x] **Slice 9: ONNX Runtime integration** *(interface + encoder; ONNX session wiring in Slice 11)*
  - [x] Add `github.com/yalue/onnxruntime_go` dependency
  - [x] Define `rlInferenceSession` interface and stubbed `newOnnxSession()`
  - [x] Create Go board encoder matching Python encoder exactly (66 channels with history support)
  - [x] Unit test: encoder output matches Python reference

- [x] **Slice 10: Inference and move selection** *(uses mock session; real ONNX in Slice 11)*
  - [x] Implement `SelectMove()` - run inference, decode policy, select legal move
  - [x] Add legal move masking
  - [x] Unit test with mock inference session

- [ ] **Slice 11: Embed trained models**
  - [ ] Train to at least iteration 500 (target ~1000 ELO)
  - [ ] Evaluate checkpoints against Stockfish to identify ELO targets
  - [ ] Export 1000/1200/1500/2000/2200 ELO models to ONNX (export what's available, mark rest as unavailable)
  - [ ] Embed available models via `go:embed` in `internal/bot/models/`
  - [ ] Update factory to load correct model based on difficulty; return `ErrModelNotLoaded` for unavailable tiers
  - [ ] Verify RL bot can play a complete game

---

## Phase C: UI Integration

- [ ] **Slice 12: Add RL bots to selection menu**
  - [ ] Add "RL Beginner (1000)", "RL Intermediate (1200)", "RL Club (1500)", "RL Advanced (2000)", "RL Master (2200)" to bot list
  - [ ] Wire selection to `NewRLEngine()` with appropriate difficulty
  - [ ] Handle error case: display message if model unavailable (greyed out or "Model not yet trained")
  - [ ] Unavailable tiers should be visible but not selectable

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
