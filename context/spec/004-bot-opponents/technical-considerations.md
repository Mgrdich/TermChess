# Technical Specification: Bot Opponents

- **Functional Specification:** [context/spec/004-bot-opponents/functional-spec.md](./functional-spec.md)
- **Status:** Draft
- **Author(s):** AWOS System

---

## 1. High-Level Technical Approach

The Bot Opponents feature will be implemented using a hybrid algorithm approach:
- **Easy Bot:** Weighted random move selection (70% favoring captures/checks, 30% fully random) to ensure novice players can win
- **Medium/Hard Bots:** Minimax algorithm with alpha-beta pruning, differentiated by search depth and evaluation function sophistication

A new `internal/bot/` package will define an `Engine` interface for bot abstraction, allowing future extensibility for UCI engines and RL agents (Phase 5). Bot move selection will integrate into the existing Bubbletea game loop using asynchronous commands to prevent UI blocking. Go's `context.Context` pattern will manage timeouts (Easy: 2s, Medium: 4s, Hard: 8s).

The implementation requires no changes to the chess engine itself (`internal/engine/`), as move generation, validation, and game state detection are already complete. The UI layer (`internal/ui/`) already has placeholder support for bot games (`GameTypePvBot`, `BotDifficulty`) and only needs bot move execution logic.

---

## 2. Proposed Solution & Implementation Plan

### 2.1 Architecture Changes

**New Package: `internal/bot/`**

Create a new package for all bot-related logic with the following structure:

```
internal/bot/
├── engine.go          # Engine interface definition
├── random.go          # Easy bot implementation (weighted random)
├── minimax.go         # Medium/Hard bot implementation (minimax + alpha-beta)
├── eval.go            # Position evaluation functions
├── engine_test.go     # Interface contract tests
├── random_test.go     # Easy bot tests
├── minimax_test.go    # Minimax algorithm tests
├── eval_test.go       # Evaluation function tests
└── tactics_test.go    # Mate-in-N puzzle tests
```

**Engine Interface** (`internal/bot/engine.go`):

```go
package bot

import (
    "context"
    "github.com/Mgrdich/TermChess/internal/engine"
)

// Engine represents a chess bot that can select moves.
// This is the minimal interface all engines must implement.
type Engine interface {
    // SelectMove returns the bot's chosen move for the given position.
    // The context allows cancellation if the bot exceeds time limits.
    SelectMove(ctx context.Context, board *engine.Board) (engine.Move, error)

    // Name returns a human-readable name for this engine.
    Name() string

    // Close releases any resources held by the engine.
    // Implementations should be idempotent (safe to call multiple times).
    // Internal bots can no-op; UCI engines kill processes; RL agents free model memory.
    Close() error
}

// Configurable engines can accept configuration before or during use.
// Internal bots implement this for difficulty tuning.
// UCI engines implement this for engine options (Threads, Hash, etc.).
type Configurable interface {
    Engine
    Configure(options map[string]any) error
}

// Stateful engines benefit from knowing position history.
// UCI engines use this for opening books and transposition tables.
// RL agents might use this for sequential context.
type Stateful interface {
    Engine
    SetPositionHistory(history []*engine.Board) error
}

// Info provides metadata about the engine.
type Info struct {
    Name       string          // Human-readable name
    Author     string          // Engine author
    Version    string          // Engine version
    Type       EngineType      // Internal, UCI, or RL
    Difficulty Difficulty      // Easy, Medium, Hard (for internal bots)
    Features   map[string]bool // Supported features
}

// Inspectable engines can report metadata.
// Useful for UI display and debugging.
type Inspectable interface {
    Engine
    Info() Info
}

// EngineType categorizes engine implementations.
type EngineType int

const (
    TypeInternal EngineType = iota // Built-in Go implementation
    TypeUCI                         // External UCI engine (Phase 5)
    TypeRL                          // RL agent with ONNX model (Phase 5)
)

// Difficulty levels for internal engines.
type Difficulty int

const (
    Easy Difficulty = iota
    Medium
    Hard
)
```

**Design Rationale:**
- **Minimal base interface:** Only 3 methods required (SelectMove, Name, Close)
- **Resource cleanup:** `Close()` follows Go's `io.Closer` pattern for proper lifecycle management
- **Optional interfaces:** Engines opt-in to capabilities (Configurable, Stateful, Inspectable)
- **Future-proof for Phase 5:** UCI engines need process cleanup, RL agents need model memory management
- **Idiomatic Go:** Follows stdlib patterns (interface composition, type assertions)
- **No breaking changes:** Adding optional interfaces doesn't affect existing implementations

---

### 2.2 Factory Pattern with Functional Options

**Factory Functions** (`internal/bot/factory.go`):

To support future configuration needs (UCI options, model paths, custom timeouts), use the functional options pattern:

```go
// EngineOption is a functional option for engine creation.
type EngineOption func(*engineConfig) error

type engineConfig struct {
    difficulty  Difficulty
    timeLimit   time.Duration
    searchDepth int
    options     map[string]any
}

// Functional options for customization
func WithTimeLimit(d time.Duration) EngineOption {
    return func(c *engineConfig) error {
        if d <= 0 {
            return fmt.Errorf("time limit must be positive")
        }
        c.timeLimit = d
        return nil
    }
}

func WithSearchDepth(depth int) EngineOption {
    return func(c *engineConfig) error {
        if depth < 1 || depth > 20 {
            return fmt.Errorf("search depth must be 1-20")
        }
        c.searchDepth = depth
        return nil
    }
}

// Factory for Easy bot
func NewRandomEngine(opts ...EngineOption) (Engine, error) {
    cfg := &engineConfig{
        difficulty: Easy,
        timeLimit:  2 * time.Second,
    }

    for _, opt := range opts {
        if err := opt(cfg); err != nil {
            return nil, err
        }
    }

    return &randomEngine{
        name:      "Easy Bot",
        timeLimit: cfg.timeLimit,
    }, nil
}

// Factory for Medium/Hard bots
func NewMinimaxEngine(difficulty Difficulty, opts ...EngineOption) (Engine, error) {
    cfg := &engineConfig{difficulty: difficulty}

    // Set defaults based on difficulty
    switch difficulty {
    case Medium:
        cfg.timeLimit = 4 * time.Second
        cfg.searchDepth = 4
    case Hard:
        cfg.timeLimit = 8 * time.Second
        cfg.searchDepth = 6
    default:
        return nil, fmt.Errorf("invalid difficulty: %d", difficulty)
    }

    // Apply custom options
    for _, opt := range opts {
        if err := opt(cfg); err != nil {
            return nil, err
        }
    }

    return &minimaxEngine{
        difficulty: difficulty,
        maxDepth:   cfg.searchDepth,
        timeLimit:  cfg.timeLimit,
    }, nil
}
```

**Usage Examples:**
```go
// Simple case (uses defaults)
easyBot, _ := bot.NewRandomEngine()
mediumBot, _ := bot.NewMinimaxEngine(bot.Medium)

// Custom configuration
customBot, _ := bot.NewMinimaxEngine(bot.Hard,
    bot.WithSearchDepth(8),
    bot.WithTimeLimit(10*time.Second))
```

---

### 2.3 Component Breakdown

#### Easy Bot: Random Engine (`internal/bot/random.go`)

**Implementation:**

```go
type randomEngine struct {
    name      string
    timeLimit time.Duration
    closed    bool
}

func (e *randomEngine) Name() string {
    return e.name
}

func (e *randomEngine) SelectMove(ctx context.Context, board *engine.Board) (engine.Move, error) {
    if e.closed {
        return engine.Move{}, errors.New("engine is closed")
    }
    moves := board.LegalMoves()
    if len(moves) == 0 {
        return engine.Move{}, errors.New("no legal moves available")
    }
    if len(moves) == 1 {
        return moves[0], nil // Only one move, return it
    }

    // Categorize moves
    captures := filterCaptures(board, moves)
    checks := filterChecks(board, moves)

    // Weighted selection: 70% tactical, 30% random
    if rand.Float64() < 0.7 && len(captures) > 0 {
        return captures[rand.Intn(len(captures))], nil
    }
    if rand.Float64() < 0.5 && len(checks) > 0 {
        return checks[rand.Intn(len(checks))], nil
    }

    // Fallback: any legal move
    return moves[rand.Intn(len(moves))], nil
}

func (e *randomEngine) Close() error {
    e.closed = true
    return nil
}

func (e *randomEngine) Info() Info {
    return Info{
        Name:       e.name,
        Author:     "TermChess",
        Version:    "1.0",
        Type:       TypeInternal,
        Difficulty: Easy,
        Features: map[string]bool{
            "tactical_awareness": true,
        },
    }
}

// Helper: filter moves that capture opponent pieces
func filterCaptures(board *engine.Board, moves []engine.Move) []engine.Move {
    var captures []engine.Move
    for _, m := range moves {
        if board.PieceAt(m.To) != engine.EmptyPiece {
            captures = append(captures, m)
        }
    }
    return captures
}

// Helper: filter moves that give check
func filterChecks(board *engine.Board, moves []engine.Move) []engine.Move {
    var checks []engine.Move
    for _, m := range moves {
        boardCopy := board.Copy()
        boardCopy.MakeMove(m)
        if boardCopy.IsInCheck(boardCopy.ActiveColor) {
            checks = append(checks, m)
        }
    }
    return checks
}
```

**Characteristics:**
- No search depth (instant move selection)
- Favors captures and checks but doesn't optimize them
- 30% chance to make completely random moves (ensures unpredictability)
- Frequently misses tactics and hangs pieces (beatable by novices)
- Artificial 1-2 second delay added in UI layer for better UX

---

#### Medium/Hard Bots: Minimax Engine (`internal/bot/minimax.go`)

**Implementation:**

```go
type minimaxEngine struct {
    difficulty  Difficulty
    maxDepth    int
    timeLimit   time.Duration
    evalWeights evalWeights
    closed      bool
}

type evalWeights struct {
    material    float64
    pieceSquare float64
    mobility    float64
    kingSafety  float64
}

func getDefaultWeights(difficulty Difficulty) evalWeights {
    switch difficulty {
    case Medium:
        return evalWeights{
            material:    1.0,
            pieceSquare: 0.1,
            mobility:    0.05,
            kingSafety:  0.0,
        }
    case Hard:
        return evalWeights{
            material:    1.0,
            pieceSquare: 0.15,
            mobility:    0.1,
            kingSafety:  0.2,
        }
    default:
        return evalWeights{material: 1.0}
    }
}

func (e *minimaxEngine) Name() string {
    if e.difficulty == Medium {
        return "Medium Bot"
    }
    return "Hard Bot"
}

func (e *minimaxEngine) SelectMove(ctx context.Context, board *engine.Board) (engine.Move, error) {
    if e.closed {
        return engine.Move{}, errors.New("engine is closed")
    }
    // Create timeout context
    ctx, cancel := context.WithTimeout(ctx, e.timeLimit)
    defer cancel()

    moves := board.LegalMoves()
    if len(moves) == 0 {
        return engine.Move{}, errors.New("no legal moves available")
    }
    if len(moves) == 1 {
        return moves[0], nil // Forced move, no search needed
    }

    var bestMove engine.Move
    var bestScore float64 = math.Inf(-1)

    // Iterative deepening: start at depth 1, gradually increase
    for depth := 1; depth <= e.maxDepth; depth++ {
        select {
        case <-ctx.Done():
            // Timeout reached, return best move from previous iteration
            if bestMove == (engine.Move{}) {
                // Fallback: return first legal move if no iteration completed
                return moves[0], nil
            }
            return bestMove, nil
        default:
            // Search at current depth
            score, move := e.searchDepth(ctx, board, depth)
            if move != (engine.Move{}) {
                bestMove = move
                bestScore = score
            }
        }
    }

    return bestMove, nil
}

func (e *minimaxEngine) searchDepth(ctx context.Context, board *engine.Board, depth int) (float64, engine.Move) {
    moves := board.LegalMoves()
    if len(moves) == 0 {
        return 0, engine.Move{}
    }

    // Order moves: captures first (MVV-LVA), then others
    orderedMoves := orderMoves(board, moves)

    var bestMove engine.Move
    alpha := math.Inf(-1)
    beta := math.Inf(1)

    for _, move := range orderedMoves {
        // Check timeout periodically
        select {
        case <-ctx.Done():
            return alpha, bestMove
        default:
        }

        boardCopy := board.Copy()
        boardCopy.MakeMove(move)

        score := -e.alphaBeta(ctx, boardCopy, depth-1, -beta, -alpha, false)

        if score > alpha {
            alpha = score
            bestMove = move
        }
    }

    return alpha, bestMove
}

func (e *minimaxEngine) alphaBeta(ctx context.Context, board *engine.Board, depth int, alpha, beta float64, maximizing bool) float64 {
    // Base case: depth 0 or game over
    if depth == 0 || board.IsGameOver() {
        return evaluate(board, e.difficulty)
    }

    moves := board.LegalMoves()
    if len(moves) == 0 {
        // Stalemate or checkmate
        return evaluate(board, e.difficulty)
    }

    orderedMoves := orderMoves(board, moves)

    for _, move := range orderedMoves {
        // Periodic timeout check (every ~100 nodes)
        if rand.Intn(100) == 0 {
            select {
            case <-ctx.Done():
                return alpha // Timeout, return best found
            default:
            }
        }

        boardCopy := board.Copy()
        boardCopy.MakeMove(move)

        score := -e.alphaBeta(ctx, boardCopy, depth-1, -beta, -alpha, !maximizing)

        if score >= beta {
            return beta // Beta cutoff
        }
        if score > alpha {
            alpha = score
        }
    }

    return alpha
}

// orderMoves: Simple MVV-LVA (Most Valuable Victim - Least Valuable Attacker)
func orderMoves(board *engine.Board, moves []engine.Move) []engine.Move {
    // Sort captures by victim value, then non-captures
    captures := make([]engine.Move, 0, len(moves))
    nonCaptures := make([]engine.Move, 0, len(moves))

    for _, m := range moves {
        if board.PieceAt(m.To) != engine.EmptyPiece {
            captures = append(captures, m)
        } else {
            nonCaptures = append(nonCaptures, m)
        }
    }

    // TODO: Sort captures by victim value

    return append(captures, nonCaptures...)
}

func (e *minimaxEngine) Close() error {
    e.closed = true
    return nil
}

func (e *minimaxEngine) Info() Info {
    return Info{
        Name:       e.Name(),
        Author:     "TermChess",
        Version:    "1.0",
        Type:       TypeInternal,
        Difficulty: e.difficulty,
        Features: map[string]bool{
            "alpha_beta":          true,
            "iterative_deepening": true,
            "move_ordering":       true,
        },
    }
}

func (e *minimaxEngine) Configure(options map[string]any) error {
    if depth, ok := options["search_depth"].(int); ok {
        if depth < 1 || depth > 20 {
            return fmt.Errorf("search depth must be 1-20")
        }
        e.maxDepth = depth
    }
    if timeLimit, ok := options["time_limit"].(time.Duration); ok {
        if timeLimit <= 0 {
            return fmt.Errorf("time limit must be positive")
        }
        e.timeLimit = timeLimit
    }
    if material, ok := options["eval_weight_material"].(float64); ok {
        e.evalWeights.material = material
    }
    if pieceSquare, ok := options["eval_weight_piece_square"].(float64); ok {
        e.evalWeights.pieceSquare = pieceSquare
    }
    if mobility, ok := options["eval_weight_mobility"].(float64); ok {
        e.evalWeights.mobility = mobility
    }
    if kingSafety, ok := options["eval_weight_king_safety"].(float64); ok {
        e.evalWeights.kingSafety = kingSafety
    }
    return nil
}
```

**Algorithm Details:**

1. **Iterative Deepening:** Start at depth 1, increment to maxDepth. If timeout occurs, return best move from last completed depth (ensures we always have a valid move).

2. **Alpha-Beta Pruning:** Negamax variant with alpha-beta cutoffs to prune branches that won't affect the final decision.

3. **Move Ordering:** Evaluate captures first using MVV-LVA (Most Valuable Victim - Least Valuable Attacker) heuristic. Better move ordering improves pruning effectiveness.

4. **Timeout Handling:** Check `ctx.Done()` periodically (every ~100 nodes) to respect time limits.

**Configuration:**
- **Medium:** maxDepth=4, timeLimit=4s
- **Hard:** maxDepth=6, timeLimit=8s

---

#### Position Evaluation (`internal/bot/eval.go`)

```go
package bot

import (
    "math"
    "github.com/Mgrdich/TermChess/internal/engine"
)

// evaluate returns a score for the position from White's perspective.
// Positive = White advantage, Negative = Black advantage
func evaluate(board *engine.Board, difficulty Difficulty) float64 {
    // 1. Check terminal states first
    status := board.Status()

    if status == engine.Checkmate {
        winner, _ := board.Winner()
        if winner == engine.White {
            return 10000.0
        }
        return -10000.0
    }

    if status == engine.Stalemate || status == engine.DrawByRepetition ||
       status == engine.DrawByFiftyMove || status == engine.DrawByInsufficientMaterial {
        return 0.0
    }

    score := 0.0

    // 2. Material count (all difficulties)
    score += countMaterial(board)

    // 3. Piece-square tables (Medium+)
    if difficulty >= Medium {
        score += evaluatePiecePositions(board)
    }

    // 4. Mobility (Medium+)
    if difficulty >= Medium {
        score += evaluateMobility(board) * 0.1
    }

    // 5. King safety (Hard only)
    if difficulty >= Hard {
        score += evaluateKingSafety(board)
    }

    return score
}

// Material values (standard)
var pieceValues = map[engine.PieceType]float64{
    engine.Pawn:   1.0,
    engine.Knight: 3.0,
    engine.Bishop: 3.25,
    engine.Rook:   5.0,
    engine.Queen:  9.0,
    engine.King:   0.0, // Invaluable
}

func countMaterial(board *engine.Board) float64 {
    score := 0.0

    for sq := 0; sq < 64; sq++ {
        piece := board.PieceAt(engine.Square(sq))
        if piece == engine.EmptyPiece {
            continue
        }

        pieceType := piece.Type()
        value := pieceValues[pieceType]

        if piece.Color() == engine.White {
            score += value
        } else {
            score -= value
        }
    }

    return score
}

func evaluatePiecePositions(board *engine.Board) float64 {
    // TODO: Implement piece-square tables
    // Bonuses for: knights in center, bishops on long diagonals,
    // rooks on open files, pawns advanced, etc.
    return 0.0
}

func evaluateMobility(board *engine.Board) float64 {
    // Number of legal moves for active player
    moves := board.LegalMoves()
    return float64(len(moves))
}

func evaluateKingSafety(board *engine.Board) float64 {
    // TODO: Implement king safety evaluation
    // Factors: pawn shield, open files near king, attackers in king zone
    return 0.0
}
```

**Evaluation Components:**

1. **Material Count (all difficulties):** Standard piece values (Pawn=1, Knight=3, Bishop=3.25, Rook=5, Queen=9)

2. **Piece-Square Tables (Medium+):** Positional bonuses for good piece placement:
   - Pawns: Advancement bonus (e.g., pawns on 6th rank more valuable)
   - Knights: Center control bonus (d4, e4, d5, e5 squares)
   - Bishops: Long diagonal bonus
   - Rooks: Open file and 7th rank bonus
   - King: Safety in opening/middlegame, activity in endgame

3. **Mobility (Medium+):** Number of legal moves weighted by 0.1 (more moves = better position)

4. **King Safety (Hard only):**
   - Pawn shield completeness (pawns in front of king)
   - Open files near king (penalty)
   - Number of enemy pieces attacking king zone (penalty)

**Progressive Complexity:** Easy bot doesn't use evaluation (random moves), Medium adds positional awareness, Hard adds king safety for strategic depth.

---

### 2.3 UI Integration

**Changes to `internal/ui/update.go`:**

**1. Bot Move Trigger (after user move):**

```go
// In handleMoveInput(), after successful user move:
if m.gameType == GameTypePvBot && !m.board.IsGameOver() {
    return m.makeBotMove()
}
```

**2. Bot Move Execution (new method):**

```go
func (m Model) makeBotMove() (tea.Model, tea.Cmd) {
    // Display thinking message
    m.statusMsg = getRandomThinkingMessage()

    // Create bot engine based on difficulty
    var botEngine bot.Engine
    switch m.botDifficulty {
    case BotEasy:
        botEngine = bot.NewRandomEngine()
    case BotMedium:
        botEngine = bot.NewMinimaxEngine(bot.Medium)
    case BotHard:
        botEngine = bot.NewMinimaxEngine(bot.Hard)
    }

    // Execute bot move asynchronously (Bubbletea command)
    return m, tea.Batch(
        m.renderView(),
        func() tea.Msg {
            ctx := context.Background()
            move, err := botEngine.SelectMove(ctx, m.board)
            if err != nil {
                return BotMoveErrorMsg{err: err}
            }
            return BotMoveMsg{move: move}
        },
    )
}
```

**3. Bot Move Message Handlers (new types and handlers):**

```go
// Message types
type BotMoveMsg struct {
    move engine.Move
}

type BotMoveErrorMsg struct {
    err error
}

// In Update() function:
case BotMoveMsg:
    // Apply bot move to board
    err := m.board.MakeMove(msg.move)
    if err != nil {
        m.errorMsg = fmt.Sprintf("Bot error: %v", err)
        return m, nil
    }

    // Update state
    m.moveHistory = append(m.moveHistory, msg.move)
    m.statusMsg = ""

    // Check game over
    if m.board.IsGameOver() {
        m.screen = ScreenGameOver
        _ = config.DeleteSaveGame() // Clear autosave
    }

    return m, nil

case BotMoveErrorMsg:
    m.errorMsg = fmt.Sprintf("Bot failed: %v", msg.err)
    m.statusMsg = ""
    return m, nil
```

**4. Thinking Messages (new file `internal/ui/messages.go`):**

```go
package ui

import "math/rand"

var thinkingMessages = []string{
    "Calculating fork trajectories...",
    "Consulting the chess gods...",
    "Pondering pawn structures...",
    "Analyzing knight maneuvers...",
    "Contemplating castle formations...",
    "Evaluating bishop diagonals...",
    "Reviewing rook highways...",
    "Meditating on the middle game...",
    "Channeling chess grandmasters...",
    "Summoning strategic insights...",
    "Counting material imbalances...",
    "Searching for tactical motifs...",
}

func getRandomThinkingMessage() string {
    return thinkingMessages[rand.Intn(len(thinkingMessages))]
}
```

**Integration Flow:**
1. User makes move → `handleMoveInput()` validates and applies move
2. If bot game and not game over → `makeBotMove()` triggers
3. Thinking message displayed, bot engine created
4. Bot move calculated asynchronously (doesn't block UI)
5. `BotMoveMsg` received → move applied → check game over
6. If error → `BotMoveErrorMsg` displays error to user

**UI Non-Blocking:** Bubbletea's command pattern ensures bot calculation runs in background goroutine, keeping UI responsive.

---

### 2.4 Bot Selection Menu

**Changes to `internal/ui/menu.go`:**

The main menu already supports bot selection via the `GameType` enum. Update the menu rendering to show bot options:

```go
// Menu options
1) Player vs Player
2) vs Easy Bot
3) vs Medium Bot
4) vs Hard Bot
5) Load Game
6) Exit
```

After bot selection, prompt for color choice:

```go
Play as:
1) White
2) Black
3) Random
```

If Random selected, assign color with 50/50 chance:
```go
if colorChoice == Random {
    if rand.Float64() < 0.5 {
        userColor = White
    } else {
        userColor = Black
    }
}
```

If bot plays White (user selected Black or random assigned Black), trigger bot move immediately after board setup.

---

## 3. Impact and Risk Analysis

### System Dependencies

- **Chess Engine (`internal/engine/`):** Bot implementation depends on:
  - `Board.LegalMoves() []Move` - Move generation
  - `Board.MakeMove(Move) error` - Move validation and application
  - `Board.Copy() *Board` - Board duplication for lookahead
  - `Board.Status() GameStatus` - Game termination detection
  - `Board.IsInCheck(Color) bool` - Check detection
  - `Board.PieceAt(Square) Piece` - Board state inspection

- **UI Layer (`internal/ui/`):** Bot move execution integrates with:
  - Bubbletea's Update/Command pattern for async execution
  - Model state (`gameType`, `botDifficulty`, `board`, `moveHistory`)
  - Screen transitions (`ScreenGamePlay` → `ScreenGameOver`)

- **No external dependencies** required (pure Go stdlib + existing packages)

**Dependency Risk:** Low. All required engine methods already exist and are well-tested.

---

### Potential Risks & Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| **Bot exceeds time limit** | UI appears frozen, poor UX | Medium | Use iterative deepening to ensure valid move at any depth; context-based timeouts; return best move from last completed iteration |
| **Bot makes illegal moves** | Game crash, corrupted state | Low | Validate all bot moves with `board.MakeMove()` which returns error; add integration tests verifying move legality |
| **Medium bot too weak/too strong** | Poor difficulty progression, frustrating UX | Medium | Tune depth and evaluation weights during testing; run bot-vs-bot games to measure strength; adjust parameters based on results |
| **Memory allocation during search** | High memory usage, GC pressure | Low | Use existing `board.Copy()` method (already optimized); profile memory with `pprof`; consider object pooling if necessary |
| **UI freezes during bot computation** | Unresponsive interface, bad UX | Low | Use Bubbletea async commands (goroutine-based); never block in `Update()`; display thinking message immediately |
| **Easy bot too predictable/boring** | Users lose interest | Low | Inject 30% randomness into move selection; vary thinking message; add slight delay for realism |
| **Hard bot too slow (>8s)** | Timeout returns weak moves | Medium | Monitor time-per-depth in tests; reduce maxDepth if needed; optimize move ordering to improve pruning |
| **Alpha-beta implementation bugs** | Incorrect move selection, poor play | Medium | Extensive unit tests with known positions; tactical puzzle suite; compare against reference implementations |
| **Evaluation function imbalance** | Bot overvalues/undervalues material | Low | Use standard piece values; test against known positions; tune weights empirically |
| **Resource leaks (Phase 5)** | Memory/process leaks with UCI/RL engines | Low | Implement `Close()` method properly; ensure UI calls `defer engine.Close()`; add cleanup in UI quit handler |

**Critical Path Risks:** The main implementation risks are:
1. **Performance tuning** - Ensuring bots meet time constraints
2. **Difficulty calibration** - Ensuring Easy is beatable and Hard is challenging
3. **UI integration** - Preventing blocking and handling async errors

All risks have clear mitigation strategies and can be addressed through testing and iterative tuning.

---

## 4. Testing Strategy

### Unit Tests

#### 4.1 Evaluation Function Tests (`internal/bot/eval_test.go`)

```go
func TestEvaluateMaterial(t *testing.T)
// Test known material imbalances
// e.g., position with extra queen should score ~+9

func TestEvaluateCheckmate(t *testing.T)
// Checkmate positions should return ±10000

func TestEvaluateStalemate(t *testing.T)
// Stalemate should return 0

func TestEvaluateSymmetry(t *testing.T)
// eval(pos) = -eval(pos_flipped_colors)
// Ensures evaluation is color-symmetric

func TestEvaluateStartPosition(t *testing.T)
// Starting position should evaluate to ~0 (equal)
```

#### 4.2 Minimax Algorithm Tests (`internal/bot/minimax_test.go`)

```go
func TestSelectMove_MateInOne(t *testing.T)
// Load "mate in 1" FEN position
// Verify bot finds the checkmate move

func TestSelectMove_MateInTwo(t *testing.T)
// Load "mate in 2" FEN position
// Verify bot finds the winning first move

func TestSelectMove_AvoidBlunder(t *testing.T)
// Position where hanging queen is possible
// Verify bot doesn't make the blunder

func TestSelectMove_ForcedMove(t *testing.T)
// Only one legal move available
// Verify bot returns it immediately (no search)

func TestSelectMove_Timeout(t *testing.T)
// Set short timeout (100ms)
// Verify bot returns valid move within timeout

func TestAlphaBeta_PruningEffectiveness(t *testing.T)
// Compare nodes searched: minimax vs alpha-beta
// Alpha-beta should search ~sqrt(minimax_nodes)
```

#### 4.3 Easy Bot Tests (`internal/bot/random_test.go`)

```go
func TestRandomEngine_LegalMoves(t *testing.T)
// Run 1000 random move selections
// Verify all returned moves are legal

func TestRandomEngine_FavorsCaptures(t *testing.T)
// Position with captures available
// Run 100 trials, verify ~70% select captures

func TestRandomEngine_NoMoves(t *testing.T)
// Position with no legal moves (checkmate/stalemate)
// Verify returns error
```

---

### Integration Tests

#### 4.4 Tactical Puzzle Suite (`internal/bot/tactics_test.go`)

Test bot performance on standard chess puzzles:

```go
func TestTactical_BackRankMate(t *testing.T)
func TestTactical_SmotheredMate(t *testing.T)
func TestTactical_Fork(t *testing.T)
func TestTactical_Pin(t *testing.T)
func TestTactical_Skewer(t *testing.T)
func TestTactical_DiscoveredAttack(t *testing.T)
```

**Puzzle Sources:**
- Load positions from FEN strings
- Verify Medium/Hard bots find the correct tactical move
- Easy bot is allowed to fail these tests (by design)

**Example Test:**
```go
func TestTactical_BackRankMate(t *testing.T) {
    fen := "6k1/5ppp/8/8/8/8/8/R6K w - - 0 1"
    board, _ := engine.FromFEN(fen)

    bot := NewMinimaxEngine(Medium)
    move, err := bot.SelectMove(context.Background(), board)

    require.NoError(t, err)
    assert.Equal(t, "a1a8", move.String()) // Ra8# is checkmate
}
```

---

#### 4.5 Difficulty Progression Tests (`internal/bot/difficulty_test.go`)

Verify bots follow expected strength hierarchy:

```go
func TestBotVsBot_MediumBeatsEasy(t *testing.T)
// Run 10 automated games: Medium vs Easy
// Assert Medium wins at least 8 games

func TestBotVsBot_HardBeatsMedium(t *testing.T)
// Run 10 automated games: Hard vs Medium
// Assert Hard wins at least 7 games

func runBotGame(white, black Engine) GameResult
// Helper: plays full game between two bots, returns result
```

**Rationale:** Statistical testing ensures difficulty levels are properly calibrated. If Easy wins too often against Medium, tuning is needed.

---

### Performance Tests

#### 4.6 Benchmark Tests (`internal/bot/minimax_test.go`)

```go
func BenchmarkMinimax_Depth2(b *testing.B)
func BenchmarkMinimax_Depth4(b *testing.B)  // Medium bot
func BenchmarkMinimax_Depth6(b *testing.B)  // Hard bot

func BenchmarkEvaluate(b *testing.B)
// Measure evaluation function performance
```

#### 4.7 Time Limit Tests (`internal/bot/timeout_test.go`)

```go
func TestTimeLimit_Medium(t *testing.T)
// Run move selection from complex middlegame position
// Assert completes within 4 seconds

func TestTimeLimit_Hard(t *testing.T)
// Run move selection from complex middlegame position
// Assert completes within 8 seconds

func TestTimeLimit_IterativeDeepeningFallback(t *testing.T)
// Set very short timeout (100ms)
// Verify bot still returns valid move (from depth 1 or 2)
```

---

### UI Integration Tests

#### 4.8 End-to-End Game Tests (`internal/ui/bot_game_test.go`)

```go
func TestFullGame_UserVsEasyBot(t *testing.T)
// Simulate full game with scripted user moves
// Verify bot responds, game completes correctly

func TestFullGame_BotMovesFirst(t *testing.T)
// User selects Black, bot plays White
// Verify bot makes opening move immediately

func TestBotMove_ErrorHandling(t *testing.T)
// Mock bot engine that returns error
// Verify UI displays error message, doesn't crash
```

---

### Test Coverage Goals

- **Unit Test Coverage:** >90% for `internal/bot/` package
- **Integration Tests:** All major tactical patterns covered
- **Performance Tests:** All bots meet time constraints in 95% of positions
- **Difficulty Tests:** Medium beats Easy 70%+, Hard beats Medium 65%+

---

## 5. Implementation Phases

### Phase 1: Foundation (2-3 days)
- Create `internal/bot/engine.go` with interface definition
- Implement `internal/bot/eval.go` with material-only evaluation
- Write unit tests for evaluation function
- **Milestone:** Evaluation function tested and validated

### Phase 2: Easy Bot (0.5 day)
- Implement `internal/bot/random.go` with weighted random selection
- Add unit tests for move legality and capture bias
- **Milestone:** Easy bot functional and tested

### Phase 3: Minimax Core (2-3 days)
- Implement `internal/bot/minimax.go` with alpha-beta pruning
- Start with depth 2, material-only evaluation (simplest version)
- Add iterative deepening and timeout handling
- Write unit tests for mate-in-1, mate-in-2 positions
- **Milestone:** Basic minimax functional with time management

### Phase 4: Difficulty Tuning (1-2 days)
- Add piece-square tables to evaluation
- Add mobility evaluation for Medium/Hard
- Add king safety evaluation for Hard
- Tune depths: Medium=4, Hard=6
- Run bot-vs-bot games to validate difficulty progression
- Performance profiling with `pprof`
- **Milestone:** All difficulty levels properly calibrated

### Phase 5: UI Integration (1 day)
- Add bot move execution to `internal/ui/update.go`
- Create `internal/ui/messages.go` with thinking messages
- Add bot selection menu options
- Add color selection prompt
- Test full PvBot game flow
- **Milestone:** Complete bot game playable in UI

### Phase 6: Polish & Testing (1 day)
- Run tactical puzzle test suite
- Fix any failing tests or edge cases
- Add artificial delay for Easy bot (UX improvement)
- Manual QA: play full games at each difficulty
- **Milestone:** Feature complete and ready for release

**Total Estimated Effort:** 7-10 days for a skilled Go developer

---

## 6. Key Files to Create/Modify

### New Files (Create)
- `/Users/mgo/Documents/TermChess/internal/bot/engine.go` - Engine interface
- `/Users/mgo/Documents/TermChess/internal/bot/random.go` - Easy bot
- `/Users/mgo/Documents/TermChess/internal/bot/minimax.go` - Medium/Hard bots
- `/Users/mgo/Documents/TermChess/internal/bot/eval.go` - Evaluation function
- `/Users/mgo/Documents/TermChess/internal/bot/engine_test.go` - Interface tests
- `/Users/mgo/Documents/TermChess/internal/bot/random_test.go` - Easy bot tests
- `/Users/mgo/Documents/TermChess/internal/bot/minimax_test.go` - Minimax tests
- `/Users/mgo/Documents/TermChess/internal/bot/eval_test.go` - Evaluation tests
- `/Users/mgo/Documents/TermChess/internal/bot/tactics_test.go` - Tactical puzzles
- `/Users/mgo/Documents/TermChess/internal/ui/messages.go` - Thinking messages

### Existing Files (Modify)
- `/Users/mgo/Documents/TermChess/internal/ui/update.go` - Add bot move execution
- `/Users/mgo/Documents/TermChess/internal/ui/menu.go` - Update bot selection menu (if needed)
- `/Users/mgo/Documents/TermChess/internal/ui/model.go` - Minor adjustments (if needed)

---

## 7. Future Extensibility

The `Engine` interface design with optional interfaces supports future Phase 5 features without breaking changes:

### UCI Engine Integration (Phase 5)

```go
type uciEngine struct {
    name            string
    cmd             *exec.Cmd
    stdin           io.WriteCloser
    stdout          *bufio.Scanner
    options         map[string]string
    positionHistory []*engine.Board
    closed          bool
}

func NewUCIEngine(path string, opts ...EngineOption) (Engine, error) {
    // Parse options (e.g., WithUCIOption("Threads", "4"))
    // Start engine process via exec.Command
    // Initialize UCI protocol (send "uci", wait for "uciok")
    // Set engine options (setoption commands)
    // Return configured engine
}

func (e *uciEngine) SelectMove(ctx context.Context, board *engine.Board) (engine.Move, error) {
    // Build position command (use history if available)
    // Send "go" command with time control
    // Wait for "bestmove" response
    // Parse and return move
}

func (e *uciEngine) SetPositionHistory(history []*engine.Board) error {
    e.positionHistory = history
    return nil
}

func (e *uciEngine) Close() error {
    if e.closed {
        return nil
    }
    e.closed = true

    // Send "quit" command
    e.stdin.Write([]byte("quit\n"))
    e.stdin.Close()

    // Wait for process to exit (with timeout)
    // Force kill if necessary
    return e.cmd.Wait()
}

func (e *uciEngine) Info() Info {
    return Info{
        Name:    e.name,
        Type:    TypeUCI,
        Features: map[string]bool{"stateful": true},
    }
}
```

### RL Agent Integration (Phase 5)

```go
type rlEngine struct {
    name    string
    session *onnxruntime.Session
    closed  bool
}

func NewRLEngine(modelPath string, opts ...EngineOption) (Engine, error) {
    // Load ONNX model from embedded FS or file
    // Create ONNX runtime session (CPU or GPU)
    // Return configured engine
}

func (e *rlEngine) SelectMove(ctx context.Context, board *engine.Board) (engine.Move, error) {
    // Convert board to input tensor
    // Run model inference
    // Decode output tensor to move
    // Return move
}

func (e *rlEngine) Close() error {
    if e.closed {
        return nil
    }
    e.closed = true

    // Free ONNX session memory
    return e.session.Destroy()
}

func (e *rlEngine) Info() Info {
    return Info{
        Name:    e.name,
        Type:    TypeRL,
        Features: map[string]bool{"neural_network": true},
    }
}
```

### UI Integration (No Changes Required)

The UI code doesn't need updates for Phase 5 engines:

```go
// Phase 2: Internal bots
easyBot, _ := bot.NewRandomEngine()
mediumBot, _ := bot.NewMinimaxEngine(bot.Medium)

// Phase 5: External engines (same interface!)
stockfish, _ := bot.NewUCIEngine("/usr/local/bin/stockfish",
    bot.WithUCIOption("Threads", "4"),
    bot.WithUCIOption("Hash", "256"))

rlBot, _ := bot.NewRLEngine("models/termchess_v1.onnx",
    bot.WithGPU(true))

// Same usage pattern for all engines
move, err := engine.SelectMove(ctx, board)
defer engine.Close() // Cleanup resources
```

**Key Benefits:**
- No UI changes required
- Consistent interface across all engine types
- Proper resource cleanup via `Close()`
- Optional capabilities via interface composition
- Easy testing via mock implementations

---

## 8. Success Criteria

The implementation will be considered successful when:

1. **Functional Requirements Met:**
   - All three difficulty levels (Easy, Medium, Hard) playable
   - Bot selection and color choice working
   - Thinking messages displayed during bot calculation
   - Game ends properly with correct result detection

2. **Performance Requirements Met:**
   - Easy bot responds within 2 seconds
   - Medium bot responds within 4 seconds
   - Hard bot responds within 8 seconds

3. **Quality Requirements Met:**
   - Easy bot beatable by novice players (loses to scripted beginner play)
   - Medium bot beats Easy bot 70%+ of games
   - Hard bot beats Medium bot 65%+ of games
   - Hard bot finds mate-in-2 positions consistently

4. **Test Coverage:**
   - 90%+ unit test coverage in `internal/bot/`
   - All tactical puzzles pass for Medium/Hard bots
   - Zero crashes or illegal moves in 100+ test games

5. **User Experience:**
   - UI remains responsive during bot calculation
   - No noticeable freezing or lag
   - Thinking messages add personality without distraction
   - Clear feedback when bot makes move

6. **Resource Management:**
   - All engines properly implement `Close()` method
   - UI cleans up bot engine on quit/game over
   - No memory leaks in 100+ consecutive games
   - Ready for Phase 5 external engines (UCI/RL) with process/model cleanup
