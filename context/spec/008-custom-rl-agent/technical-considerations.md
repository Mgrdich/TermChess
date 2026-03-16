# Technical Specification: Custom RL Agent

- **Functional Specification:** `context/spec/008-custom-rl-agent/functional-spec.md`
- **Status:** Draft
- **Author(s):** Mgrdich

---

## 1. High-Level Technical Approach

This feature requires two separate components:

1. **Python Training Pipeline** (`training/` directory) - AlphaZero-style self-play training using PyTorch with MPS acceleration, exporting trained models to ONNX format at different strength levels.

2. **Go Runtime Integration** (`internal/bot/rl.go`) - Implements the existing `bot.Engine` interface, loads ONNX models via `onnxruntime-go`, and embeds models in the binary for single-file distribution.

**Key Design Decision:** Training is offline (Python), inference is runtime (Go). No Python dependency in the distributed binary.

---

## 2. Proposed Solution & Implementation Plan (The "How")

### 2.1 Python Training Pipeline

**Directory Structure:**
```
training/
├── pyproject.toml        # Project config and dependencies (uv)
├── train.py              # Main training loop
├── model.py              # Neural network architecture
├── mcts.py               # Monte Carlo Tree Search
├── board_encoder.py      # Convert board state to tensor
├── replay_buffer.py      # Store training examples
├── export_onnx.py        # Export checkpoints to ONNX
└── evaluate.py           # ELO estimation vs Stockfish
```

**Neural Network Architecture (Small Config):**
- Input: 66 channels x 8 x 8 (18 current position + 4 x 12 history piece planes)
  - Channels 0-11: Current piece positions (6 white + 6 black)
  - Channel 12: Side to move
  - Channels 13-16: Castling rights
  - Channel 17: En passant
  - Channels 18-65: Last 4 positions (12 piece planes each, most recent first)
- Body: 6 residual blocks, 128 filters each
- Policy head: 4096 outputs (64 from-squares x 64 to-squares)
- Value head: 1 output with tanh activation [-1, 1]
- Parameters: ~2.4M

**Training Parameters:**
| Parameter | Value |
|-----------|-------|
| MCTS simulations per move | 100-400 (staged) |
| Games per iteration | 20-100 (staged) |
| Training batch size | 256 |
| Replay buffer size | 500K positions |
| Learning rate | 0.001 -> 0.0001 (decay at 50K) |
| Optimizer | Adam with weight decay 1e-4 |
| Gradient clipping | max_norm=1.0 |
| Value loss weight | configurable (default 1.0) |
| Dirichlet noise | alpha=0.3, epsilon=0.25 at MCTS root |
| Repetition draw penalty | -0.2 value target |
| Data augmentation | Horizontal flip (50% during sampling) |
| Move history | Last 4 positions encoded |

**Checkpoint Strategy:**
- Auto-save `checkpoint_latest.pt` + `buffer_latest.npz` every iteration (crash recovery)
- Save numbered checkpoints at: 10, 25, 50, 100, 250, 500, 1K, 2.5K, 5K, 10K, 30K, 80K
- Per-iteration CSV log (`training_log.csv`) with game stats for health monitoring
- Evaluate against Stockfish (fixed depth) to estimate ELO
- Export to ONNX when target ELOs reached (1000, 1200, 1500, 2000, 2200)

**Training Stages:**
| Stage | Iterations | Target ELO | MCTS Sims | Games/Iter |
|-------|-----------|------------|-----------|------------|
| 1 - Beginner | 0-500 | ~1000 | 100 | 20 |
| 2 - Intermediate | 500-2500 | ~1200 | 100 | 20 |
| 3 - Club Player | 2500-5000 | ~1500 | 150 | 30 |
| 4 - Advanced | 5000-30000 | ~2000 | 200 | 50 |
| 5 - Master | 30000-80000 | ~2200 | 400 | 100 |

> **Note:** The 1000 and 1200 ELO models may not yet be available if training has not progressed far enough. The UI should handle this gracefully — see Section 2.5 of the functional spec.

### 2.2 Go Runtime Integration

**New Files:**
```
internal/bot/
├── rl.go                 # rlEngine implementation
├── rl_encoder.go         # Board encoding (66 channels with history)
├── rl_encoder_test.go    # Encoder tests
├── rl_test.go            # Unit tests
├── rl_messages.go        # RL-themed thinking messages
└── models/
    ├── rl_1000.onnx      # Embedded via go:embed (when available)
    ├── rl_1200.onnx      # (when available)
    ├── rl_1500.onnx
    ├── rl_2000.onnx
    └── rl_2200.onnx
```

**RL Engine Implementation:**

```go
type RLDifficulty int

const (
    RLBeginner     RLDifficulty = iota  // 1000 ELO
    RLIntermediate                       // 1200 ELO
    RLClub                               // 1500 ELO
    RLAdvanced                           // 2000 ELO
    RLMaster                             // 2200 ELO
)

type rlEngine struct {
    name       string
    difficulty RLDifficulty
    session    *ort.Session
    timeLimit  time.Duration
    closed     int32
}

func NewRLEngine(difficulty RLDifficulty, opts ...EngineOption) (Engine, error)
func (e *rlEngine) SelectMove(ctx context.Context, board *engine.Board) (engine.Move, error)
func (e *rlEngine) Name() string
func (e *rlEngine) Close() error
func (e *rlEngine) Info() Info  // Implements Inspectable
```

**Model Embedding:**
```go
//go:embed models/rl_1000.onnx
var modelRL1000 []byte  // may not exist yet — handle gracefully

//go:embed models/rl_1200.onnx
var modelRL1200 []byte  // may not exist yet — handle gracefully

//go:embed models/rl_1500.onnx
var modelRL1500 []byte

//go:embed models/rl_2000.onnx
var modelRL2000 []byte

//go:embed models/rl_2200.onnx
var modelRL2200 []byte
```

> **Note:** The 1000 and 1200 ELO models depend on early training stages completing successfully. If these models are not yet available, the `go:embed` directives should be guarded or the factory should return `ErrModelNotLoaded` for those difficulties.

**Board Encoding (Go side for inference):**
- Convert `engine.Board` + history to float32 tensor [1, 66, 8, 8]
- 18 channels for current position + 4 x 12 channels for history piece planes
- Match Python encoder exactly for compatibility
- History is passed as `[]*engine.Board` (nil = zero-filled history planes)
- TODO: Game loop should track board history and pass to `SelectMove()`

**RL Thinking Messages:**
```go
var rlThinkingMessages = []string{
    "Neural pathways firing...",
    "Consulting the matrix...",
    "Adjusting weights...",
    "Running inference...",
    "Propagating through layers...",
    "Calculating policy distribution...",
    "Evaluating position value...",
}
```

### 2.3 UI Integration

**Bot Selection Menu:**
- Add up to five new entries after existing bots:
  - "RL Beginner (1000)" — may show as unavailable if model not yet trained
  - "RL Intermediate (1200)" — may show as unavailable if model not yet trained
  - "RL Club (1500)"
  - "RL Advanced (2000)"
  - "RL Master (2200)"
- Add "Press 'i' for info" hint
- Unavailable models should be visually distinct (greyed out or marked)

**Info Page:**
- Display ELO ratings and descriptions
- Explain RL bots are trained via deep learning
- Accessible via 'i' key from bot selection

**Error Handling:**
- If ONNX model fails to load, display error message
- User cannot proceed until issue resolved
- RL options remain visible in menu

### 2.4 Dependencies

**Python (training only) - managed via uv:**
```bash
uv init training && cd training
uv add torch numpy python-chess onnx onnxruntime
uv add --dev pytest
```
- PyTorch >= 2.0 (MPS support)
- numpy
- python-chess (for Stockfish evaluation)
- onnx, onnxruntime (for export verification)

**Go (runtime):**
- `github.com/yalue/onnxruntime_go` - ONNX Runtime bindings

---

## 3. Impact and Risk Analysis

### System Dependencies

| Component         | Depends On                    | Affects                   |
|-------------------|-------------------------------|---------------------------|
| Training pipeline | None (standalone)             | Produces ONNX models      |
| RL engine         | ONNX Runtime, embedded models | Bot selection, Bot vs Bot |
| UI                | RL engine factory             | Menu, error display       |

### Potential Risks & Mitigations

| Risk                        | Impact              | Mitigation                                                                                      |
|-----------------------------|---------------------|-------------------------------------------------------------------------------------------------|
| Training takes too long     | Delays feature      | Start with Tiny config (4 blocks, 64 filters) to validate pipeline first                        |
| Models too large for binary | Binary bloat        | ONNX models compress well; target ~10-20MB per model; consider external download if >50MB total |
| ONNX Runtime compatibility  | Platform issues     | Test on macOS, Linux, Windows early; use well-supported ops only                                |
| ELO calibration inaccurate  | Difficulty mismatch | Evaluate checkpoints against Stockfish at fixed depth; iterate on targets                       |
| MPS unsupported operations  | Training fails      | Stick to standard PyTorch ops (Conv2d, Linear, ReLU, BatchNorm); avoid exotic layers            |
| Board encoding mismatch     | Wrong moves         | Unit test encoder output matches between Python and Go                                          |

---

## 4. Testing Strategy

### Python Training Tests
- **Board encoder:** Verify output shape [66, 8, 8]; test known positions; test history encoding
- **MCTS:** Verify finds mate-in-1 positions; check visit count distribution
- **Network:** Test forward pass shapes; verify policy sums to ~1 after softmax
- **Export:** Verify ONNX model loads and produces same output as PyTorch

### Go Integration Tests
- **rlEngine:** Unit tests following existing `bot/` patterns
- **Model loading:** Test with valid model, missing model, corrupted model
- **Board encoding:** Compare Go encoder output (66 channels) against Python reference; test with and without history
- **Move decoding:** Verify legal move masking works correctly
- **Unavailable models:** Verify graceful handling when ONNX file not embedded

### End-to-End Tests
- Manual testing of RL bots in Player vs Bot mode
- RL bots in Bot vs Bot mode (all combinations)
- Error message display when model unavailable
- Verify unavailable tiers show appropriate message in UI
