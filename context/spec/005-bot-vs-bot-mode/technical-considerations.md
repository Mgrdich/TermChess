# Technical Specification: Bot vs Bot Mode

- **Functional Specification:** `context/spec/005-bot-vs-bot-mode/functional-spec.md`
- **Status:** Implemented
- **Author(s):** AI Assistant

---

## 1. High-Level Technical Approach

Bot vs Bot mode introduces parallel game execution, grid-based multi-board rendering, and statistics aggregation. The implementation splits into:

- **`internal/bvb/`** — Pure logic package: game sessions (each running in its own goroutine), a session manager for orchestration (pause/resume, speed, state), and statistics collection.
- **`internal/ui/`** — Extended Bubbletea screens for the BvB configuration flow, grid rendering with lipgloss, and tick-based UI updates that poll the session manager for latest state.

**Core pattern:** Each game runs in a goroutine. Engines compute moves concurrently. The UI uses `tea.Tick` at the configured speed interval to poll session state and re-render. Sessions self-pace: each goroutine sleeps for the configured delay between moves using a shared speed pointer. When speed changes, all sessions pick up the new delay via the shared pointer. For Instant speed, no sleep at all (UI polls at 100ms intervals for rendering).

---

## 2. Implementation Details

### 2.1 Package: `internal/bvb/types.go`

```go
type PlaybackSpeed int
const (
    SpeedInstant PlaybackSpeed = iota  // 0ms delay
    SpeedFast                          // 500ms per move
    SpeedNormal                        // 1500ms per move
    SpeedSlow                          // 3000ms per move
)

func (s PlaybackSpeed) Duration() time.Duration  // Returns the delay

type SessionState int
const (
    StateRunning SessionState = iota
    StatePaused
    StateFinished
)

type GameResult struct {
    GameNumber    int
    Winner        string           // Engine name or "Draw"
    WinnerColor   engine.Color
    EndReason     string           // "checkmate", "stalemate", "max moves reached", "engine error: ..."
    MoveCount     int
    Duration      time.Duration
    FinalFEN      string
    MoveHistory   []engine.Move
}
```

---

### 2.2 Package: `internal/bvb/session.go`

A `GameSession` encapsulates a single game running in a goroutine:

```go
const maxMoveCount = 500  // Forced draw after 500 moves

type GameSession struct {
    mu          sync.Mutex
    gameNumber  int
    board       *engine.Board
    whiteEngine bot.Engine
    blackEngine bot.Engine
    whiteName   string          // "Easy Bot", "Medium Bot", "Hard Bot"
    blackName   string
    moveHistory []engine.Move
    state       SessionState
    paused      bool
    result      *GameResult
    startTime   time.Time
    speed       *PlaybackSpeed  // Pointer to manager's speed (shared, read atomically via mutex)
    stopCh      chan struct{}    // Signal abort
    pauseCh     chan struct{}    // Signal pause (buffered, cap 1)
    resumeCh    chan struct{}    // Signal resume (buffered, cap 1)
}
```

**Constructor:**
```go
func NewGameSession(gameNumber int, whiteEngine, blackEngine bot.Engine,
    whiteName, blackName string, speed *PlaybackSpeed) *GameSession
```

**Key methods:**
- `Run()` — Goroutine entry point. Loop: check stop → check pause → compute move with 30s context timeout → apply move → check game over → sleep for speed delay → repeat
- `Pause()` — Sends on pauseCh
- `Resume()` — Sends on resumeCh
- `SetSpeed(speed)` — Updates the shared speed pointer
- `Abort()` — Closes stopCh, closes engines
- `CurrentBoard() *engine.Board` — Thread-safe board reference for rendering
- `CurrentMoveHistory() []engine.Move` — Thread-safe move history copy
- `IsFinished() bool`
- `Result() *GameResult`
- `GameNumber() int`
- `State() SessionState`

**Error handling:**
- `finishWithError(engineName, color, err)` — Engine that errors loses; opponent wins with "engine error" reason
- `finishWithStatus(status, moveCount)` — Normal game end (checkmate, stalemate, draw conditions)
- Per-move context timeout of 30 seconds prevents infinite engine computation
- Max 500 moves enforced as forced draw

**Lifecycle:**
- Each session owns its engine instances (created by the manager via factory functions)
- Engines are closed when the session finishes or is aborted
- Goroutine exits cleanly on stop signal, game over, or error

---

### 2.3 Package: `internal/bvb/manager.go`

The `SessionManager` orchestrates N sessions:

```go
type SessionManager struct {
    mu        sync.Mutex
    sessions  []*GameSession
    state     SessionState
    speed     PlaybackSpeed
    whiteDiff bot.Difficulty
    blackDiff bot.Difficulty
    whiteName string
    blackName string
    gameCount int
}
```

**Constructor:**
```go
func NewSessionManager(whiteDiff, blackDiff bot.Difficulty,
    whiteName, blackName string, gameCount int) *SessionManager
```

**Key methods:**
- `Start() error` — Creates engine pairs for each game, launches all sessions as goroutines
- `Pause()` — Signals all sessions to pause
- `Resume()` — Signals all sessions to resume
- `SetSpeed(speed PlaybackSpeed)` — Updates speed for all sessions
- `Abort()` — Stops all sessions, closes all engines
- `Sessions() []*GameSession` — Returns sessions slice for rendering
- `Stats() *AggregateStats` — Collects results from finished sessions, calls ComputeStats
- `AllFinished() bool` — Returns true when all sessions are in StateFinished
- `State() SessionState`
- `Speed() PlaybackSpeed`

---

### 2.4 Package: `internal/bvb/stats.go`

```go
type AggregateStats struct {
    TotalGames        int
    WhiteBotName      string
    BlackBotName      string
    WhiteWins         int
    BlackWins         int
    Draws             int
    WhiteWinPct       float64
    BlackWinPct       float64
    AvgMoveCount      float64
    AvgDuration       time.Duration
    ShortestGame      GameResult
    LongestGame       GameResult
    IndividualResults []GameResult
}

func ComputeStats(results []GameResult, whiteName, blackName string) *AggregateStats
```

---

### 2.5 UI Changes: Model Extensions (`internal/ui/model.go`)

```go
// GameType enum addition:
GameTypeBvB GameType = 2

// New screen states:
ScreenBvBBotSelect      // Select white/black bot difficulties (two-step)
ScreenBvBGameMode       // Single game or multi-game + count input
ScreenBvBGridConfig     // Grid size configuration
ScreenBvBGamePlay       // Watching games (tick-driven updates)
ScreenBvBStats          // Statistics display after all games finish

// BvBViewMode type:
type BvBViewMode int
const (
    BvBGridView   BvBViewMode = iota
    BvBSingleView
)

// Model struct additions:
bvbWhiteDiff        BotDifficulty
bvbBlackDiff        BotDifficulty
bvbSelectingWhite   bool              // True when selecting White bot, false for Black
bvbGameCount        int
bvbCountInput       string            // Text input for game count
bvbInputtingCount   bool              // Whether in text input mode for count
bvbGridRows         int
bvbGridCols         int
bvbCustomGridInput  string            // Text input for custom grid dimensions
bvbInputtingGrid    bool              // Whether in text input mode for grid
bvbManager          *bvb.SessionManager
bvbSpeed            bvb.PlaybackSpeed
bvbSelectedGame     int               // Which game is focused in single view (0-indexed)
bvbViewMode         BvBViewMode
bvbPaused           bool
bvbPageIndex        int               // Current page in grid view
bvbStatsSelection   int               // Selected option on stats screen (0=New Session, 1=Return)
termWidth           int               // Terminal width (from WindowSizeMsg)
termHeight          int               // Terminal height (from WindowSizeMsg)
```

---

### 2.6 UI Changes: Screen Flow (`internal/ui/update.go`)

Navigation flow:
```
ScreenGameTypeSelect → "Bot vs Bot"
    → ScreenBvBBotSelect (select White difficulty, then Black difficulty)
    → ScreenBvBGameMode (Single Game / Multi-Game + count input)
        → Single Game: directly to ScreenBvBGamePlay (1x1 grid, single-board view)
        → Multi-Game: → ScreenBvBGridConfig (grid size presets or custom input)
                      → ScreenBvBGamePlay (watching games, tick-based updates)
    → ScreenBvBStats (statistics, New Session or Return to Menu)
```

Key handler functions (all in `update.go`):
- `handleBvBBotSelectKeys(msg)` — Up/down/enter/esc; two-step White then Black selection
- `handleBvBGameModeKeys(msg)` — Menu or text input mode; validates positive integers
- `handleBvBGridConfigKeys(msg)` — Menu presets or custom "RxC" input
- `handleBvBGamePlayKeys(msg)` — Space, 1-4, Tab, ←/→, f, ESC; view-mode-aware navigation
- `handleBvBStatsKeys(msg)` — Up/down/enter/esc for New Session or Return to Menu

Helper functions:
- `startBvBSession()` — Creates SessionManager, starts it, sets initial speed/view
- `uiBotDiffToBvB(BotDifficulty) bot.Difficulty` — Maps UI enum to bot package enum
- `parsePositiveInt(string) (int, error)` — Validates game count input
- `parseGridDimensions(string) (rows, cols int, err error)` — Validates "RxC" grid input

---

### 2.7 UI Changes: Tick-Based Updates (`internal/ui/update.go`)

```go
// Message type:
type BvBTickMsg struct{}

// Tick command:
func bvbTickCmd(speed bvb.PlaybackSpeed) tea.Cmd {
    delay := speed.Duration()
    if delay == 0 {
        // For instant speed, use 100ms tick interval for rendering
        delay = 100 * time.Millisecond
    }
    return tea.Tick(delay, func(time.Time) tea.Msg {
        return BvBTickMsg{}
    })
}

// In Update():
case tea.WindowSizeMsg:
    m.termWidth = msg.Width
    m.termHeight = msg.Height
case BvBTickMsg:
    return m.handleBvBTick()  // Transitions to ScreenBvBStats when AllFinished()
```

**Instant mode:** Ticks fire every 100ms. Goroutines compute moves as fast as engines can (no sleep). UI re-renders on each tick showing current progress.

---

### 2.8 UI Changes: Rendering (`internal/ui/view.go`)

All BvB rendering functions are in `view.go` alongside existing render functions:

- `renderBvBBotSelect()` — Bot difficulty selection with White/Black indicator
- `renderBvBGameMode()` — Menu or text input for game count
- `renderBvBGridConfig()` — Grid preset menu or custom dimension input
- `renderBvBGamePlay()` — Routes to grid or single view based on bvbViewMode
- `renderBvBSingleView()` — Full board, matchup, game progress, move history, speed, help
- `renderBvBGridView()` — Paginated compact board grid with matchup info, page/speed indicators
- `renderBoardGrid(sessions, cols)` — Arranges cells using lipgloss JoinHorizontal/JoinVertical
- `renderCompactBoardCell(session)` — Game number, compact board (no coords/color), moves, status
- `renderBvBStats()` — Single or multi-game statistics with menu options

**Grid rendering details:**
- Uses lipgloss `JoinHorizontal` for rows and `JoinVertical` to stack rows
- Compact boards rendered with `BoardRenderer` using `ShowCoords: false, UseColors: false`
- Finished games styled with dimmed foreground color (`#626262`)
- Page indicator shown when totalPages > 1
- Terminal size check: if `termWidth < cols*14` or `termHeight < rows*11+8`, shows warning

---

### 2.9 Key Bindings During BvB Gameplay

| Key | Action |
|-----|--------|
| Space | Pause / Resume all games |
| 1 | Set speed to Instant |
| 2 | Set speed to Fast |
| 3 | Set speed to Normal |
| 4 | Set speed to Slow |
| Tab | Toggle grid view / single-board view |
| ←/→ or h/l | Navigate pages (grid) or games (single view) |
| f | Export FEN of focused game to clipboard |
| ESC | Abort session, return to menu |
| q | Quit application (cleans up bvbManager) |
| Ctrl+C | Exit application (cleans up bvbManager) |

---

### 2.10 Goroutine Lifecycle

```
Start() called:
  ├── For each game (1..N):
  │   ├── Create white engine via bot.NewRandomEngine() or bot.NewMinimaxEngine()
  │   ├── Create black engine via bot.NewRandomEngine() or bot.NewMinimaxEngine()
  │   ├── Create GameSession with shared speed pointer
  │   └── go session.Run()
  │         ├── Loop:
  │         │   ├── Check stopCh (select) → close engines, return
  │         │   ├── Check pauseCh → block on resumeCh until signaled
  │         │   ├── Call engine.SelectMove(ctx, board) with 30s timeout
  │         │   ├── If error: finishWithError() (opponent wins), return
  │         │   ├── Apply move to board
  │         │   ├── Check game over (checkmate/stalemate/draw/max moves)
  │         │   │   └── If over: finishWithStatus(), close engines, return
  │         │   ├── Read speed via shared pointer
  │         │   ├── If speed > 0: sleep for speed.Duration() (interruptible by stop)
  │         │   └── Continue
  │         └── Deferred: close engines on any exit path
  └── Manager tracks completion via session.IsFinished()
```

---

### 2.11 Files Summary

| File | Location | Purpose |
|------|----------|---------|
| `types.go` | `internal/bvb/` | PlaybackSpeed, SessionState, GameResult types |
| `session.go` | `internal/bvb/` | GameSession struct, Run() goroutine, thread-safe accessors |
| `manager.go` | `internal/bvb/` | SessionManager, orchestration, pause/resume/speed/abort |
| `stats.go` | `internal/bvb/` | AggregateStats, ComputeStats() |
| `model.go` | `internal/ui/` | Extended with BvB fields, GameTypeBvB, new screen states, BvBViewMode |
| `update.go` | `internal/ui/` | BvBTickMsg, all BvB key handlers, tick scheduling, session startup |
| `view.go` | `internal/ui/` | All BvB render functions (bot select, game mode, grid config, gameplay, stats) |

**Files NOT modified:** `internal/bot/*`, `internal/engine/*`, `internal/ui/board.go` (reused as-is)

---

## 3. Impact and Risk Analysis

### System Dependencies
- **`internal/bot/`** — Used as-is. Factory functions create engine instances per session. No changes needed.
- **`internal/engine/`** — Board and move types used by sessions. No changes needed.
- **`internal/ui/`** — Extended with new screens, message types, and view functions. Existing PvP and PvBot flows unchanged.

### Potential Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Many goroutines with Hard bots (depth 6) causing high CPU | UI lag, system unresponsive | 30-second context timeout per move; games run independently |
| Race conditions on shared board state | Panics, incorrect renders | Each session uses sync.Mutex; UI reads via thread-safe methods |
| Memory usage with many concurrent games | High memory for large counts | Each game is lightweight (~KB); practical limit is CPU, not memory |
| Engines never finishing (infinite loops) | Games hang indefinitely | Max 500 moves → forced draw; 30s context timeout per move |
| Terminal too small for grid | Broken layout | Detect terminal size via WindowSizeMsg; show warning if too small |
| Goroutine leaks on abort | Resource leak | Explicit stopCh close; deferred cleanup in Run(); manager aborts all on Ctrl+C/q/ESC |

---

## 4. Testing Strategy

### Unit Tests (`internal/bvb/`)
- `session_test.go`: Single game runs to completion; pause/resume works; abort stops cleanly; max move limit enforced; thread-safe accessors return correct data; engine error handling
- `manager_test.go`: Multiple sessions run in parallel and all finish; pause/resume affects all sessions; speed change propagates; abort cleans up all goroutines; Stats() computes correctly
- `stats_test.go`: Statistics computation with known results (wins, draws, averages, shortest/longest)
- `types_test.go`: Speed duration mapping

### Unit Tests (`internal/ui/`)
- `e2e_test.go`: Screen navigation (ESC goes back, Enter advances); input validation for game count and grid dimensions; view toggle; page navigation; speed changes; pause/resume; FEN export; stats rendering; complete flow integration; help text config; terminal size fallback; cleanup on quit

### Integration Tests
- Complete flow test: menu → bot select → game mode → grid → gameplay → stats → menu
- Ctrl+C and 'q' properly clean up bvbManager
- Terminal size warning when too small
- WindowSizeMsg updates dimensions
- Stats accuracy with real game completions (Easy vs Easy at instant speed)

### Test Coverage Summary
- 60+ BvB-specific tests across `internal/bvb/` and `internal/ui/`
- All tests pass with `go test ./...`
- `go vet ./...` clean
