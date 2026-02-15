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
├── requirements.txt
├── train.py              # Main training loop
├── model.py              # Neural network architecture
├── mcts.py               # Monte Carlo Tree Search
├── board_encoder.py      # Convert board state to tensor
├── replay_buffer.py      # Store training examples
├── export_onnx.py        # Export checkpoints to ONNX
└── evaluate.py           # ELO estimation vs Stockfish
```

**Neural Network Architecture (Small Config):**
- Input: 18 channels x 8 x 8 (pieces, castling, en passant, side to move)
- Body: 6 residual blocks, 128 filters each
- Policy head: 4096 outputs (64 from-squares x 64 to-squares)
- Value head: 1 output with tanh activation [-1, 1]
- Parameters: ~2M

**Training Parameters:**
| Parameter | Value |
|-----------|-------|
| MCTS simulations per move | 400 |
| Games per iteration | 100 |
| Training batch size | 256 |
| Replay buffer size | 500K positions |
| Learning rate | 0.001 -> 0.0001 (decay) |
| Optimizer | Adam with weight decay 1e-4 |

**Checkpoint Strategy:**
- Save checkpoints at intervals: 5K, 10K, 30K, 80K iterations
- Evaluate against Stockfish (fixed depth) to estimate ELO
- Export to ONNX when target ELOs reached (1500, 2000, 2200)

### 2.2 Go Runtime Integration

**New Files:**
```
internal/bot/
├── rl.go                 # rlEngine implementation
├── rl_test.go            # Unit tests
├── rl_messages.go        # RL-themed thinking messages
└── models/
    ├── rl_1500.onnx      # Embedded via go:embed
    ├── rl_2000.onnx
    └── rl_2200.onnx
```

**RL Engine Implementation:**

```go
type RLDifficulty int

const (
    RLIntermediate RLDifficulty = iota  // 1500 ELO
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
//go:embed models/rl_1500.onnx
var modelRL1500 []byte

//go:embed models/rl_2000.onnx
var modelRL2000 []byte

//go:embed models/rl_2200.onnx
var modelRL2200 []byte
```

**Board Encoding (Go side for inference):**
- Convert `engine.Board` to float32 tensor [1, 18, 8, 8]
- Match Python encoder exactly for compatibility

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
- Add three new entries after existing bots:
  - "RL Intermediate (1500)"
  - "RL Advanced (2000)"
  - "RL Master (2200)"
- Add "Press 'i' for info" hint

**Info Page:**
- Display ELO ratings and descriptions
- Explain RL bots are trained via deep learning
- Accessible via 'i' key from bot selection

**Error Handling:**
- If ONNX model fails to load, display error message
- User cannot proceed until issue resolved
- RL options remain visible in menu

### 2.4 Dependencies

**Python (training only):**
- PyTorch >= 2.0 (MPS support)
- numpy
- python-chess (for Stockfish evaluation)
- onnx, onnxruntime (for export verification)

**Go (runtime):**
- `github.com/yalue/onnxruntime_go` - ONNX Runtime bindings

---

## 3. Impact and Risk Analysis

### System Dependencies

| Component | Depends On | Affects |
|-----------|------------|---------|
| Training pipeline | None (standalone) | Produces ONNX models |
| RL engine | ONNX Runtime, embedded models | Bot selection, Bot vs Bot |
| UI | RL engine factory | Menu, error display |

### Potential Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Training takes too long | Delays feature | Start with Tiny config (4 blocks, 64 filters) to validate pipeline first |
| Models too large for binary | Binary bloat | ONNX models compress well; target ~10-20MB per model; consider external download if >50MB total |
| ONNX Runtime compatibility | Platform issues | Test on macOS, Linux, Windows early; use well-supported ops only |
| ELO calibration inaccurate | Difficulty mismatch | Evaluate checkpoints against Stockfish at fixed depth; iterate on targets |
| MPS unsupported operations | Training fails | Stick to standard PyTorch ops (Conv2d, Linear, ReLU, BatchNorm); avoid exotic layers |
| Board encoding mismatch | Wrong moves | Unit test encoder output matches between Python and Go |

---

## 4. Testing Strategy

### Python Training Tests
- **Board encoder:** Verify output shape [18, 8, 8]; test known positions
- **MCTS:** Verify finds mate-in-1 positions; check visit count distribution
- **Network:** Test forward pass shapes; verify policy sums to ~1 after softmax
- **Export:** Verify ONNX model loads and produces same output as PyTorch

### Go Integration Tests
- **rlEngine:** Unit tests following existing `bot/` patterns
- **Model loading:** Test with valid model, missing model, corrupted model
- **Board encoding:** Compare Go encoder output against Python reference
- **Move decoding:** Verify legal move masking works correctly

### End-to-End Tests
- Manual testing of RL bots in Player vs Bot mode
- RL bots in Bot vs Bot mode (all combinations)
- Error message display when model unavailable
