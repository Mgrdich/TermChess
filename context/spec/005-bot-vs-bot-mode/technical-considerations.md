# Technical Specification: Bot vs Bot Mode

- **Functional Specification:** `context/spec/005-bot-vs-bot-mode/functional-spec.md`
- **Status:** Draft
- **Author(s):** AI Assistant

---

## 1. High-Level Technical Approach

Bot vs Bot mode introduces parallel game execution, grid-based multi-board rendering, and statistics aggregation. The implementation splits into:

- **`internal/bvb/`** — Pure logic package: game sessions (each running in its own goroutine), a session manager for orchestration (pause/resume, speed, state), and statistics collection.
- **`internal/ui/`** — New Bubbletea screens for the BvB configuration flow, grid rendering with lipgloss, and tick-based UI updates that poll the session manager for latest state.

**Core pattern:** Each game runs in a goroutine. Engines compute moves concurrently. The UI uses `tea.Tick` at the configured speed interval to poll session state and re-render. Pause/resume controls whether ticks trigger move advancement or just re-render. Sessions self-pace: each goroutine sleeps for the configured delay between moves. When speed changes, all sessions pick up the new delay. For Instant speed, no sleep at all.

---

## 2. Proposed Solution & Implementation Plan

### 2.1 New Package: `internal/bvb/types.go`

```go
type PlaybackSpeed int
const (
    SpeedInstant PlaybackSpeed = iota
    SpeedFast      // 500ms per move
    SpeedNormal    // 1500ms per move
    SpeedSlow      // 3000ms per move
)

type SessionState int
const (
    StateRunning SessionState = iota
    StatePaused
    StateFinished
)

type GameResult struct {
    GameNumber    int
    Winner        string           // e.g. "Easy Bot" or "Draw"
    WinnerColor   engine.Color
    EndReason     string           // "checkmate", "stalemate", "draw by repetition", etc.
    MoveCount     int
    Duration      time.Duration
    FinalFEN      string
    MoveHistory   []engine.Move
}
```

---

### 2.2 New Package: `internal/bvb/session.go`

A `GameSession` encapsulates a single game running in a goroutine:

```go
type GameSession struct {
    mu            sync.Mutex
    gameNumber    int
    board         *engine.Board
    whiteEngine   bot.Engine
    blackEngine   bot.Engine
    whiteName     string          // "Easy Bot", "Medium Bot", "Hard Bot"
    blackName     string
    moveHistory   []engine.Move
    state         SessionState
    result        *GameResult
    startTime     time.Time
    speed         *PlaybackSpeed  // pointer to manager's speed (shared)
    pauseCh       chan struct{}    // Signal pause
    resumeCh      chan struct{}    // Signal resume
    stopCh        chan struct{}    // Signal abort
}
```

**Key methods:**
- `Run()` — Goroutine entry point. Loops: check pause/stop → compute next move via engine.SelectMove() → apply move → sleep for speed delay → check game over → repeat
- `CurrentBoard() *engine.Board` — Thread-safe board snapshot for rendering
- `CurrentMoveHistory() []engine.Move` — Thread-safe move history copy
- `IsFinished() bool`
- `Result() *GameResult`

**Lifecycle:**
- Each session owns its engine instances (created by the manager)
- Engines are closed when the session finishes or is aborted
- Max move count per game (500 moves) enforced as forced draw to prevent infinite games

---

### 2.3 New Package: `internal/bvb/manager.go`

The `SessionManager` orchestrates N sessions:

```go
type SessionManager struct {
    mu           sync.Mutex
    sessions     []*GameSession
    state        SessionState
    speed        PlaybackSpeed
    whiteDiff    bot.Difficulty
    blackDiff    bot.Difficulty
    whiteName    string
    blackName    string
    results      []GameResult
}
```

**Key methods:**
- `NewSessionManager(whiteDiff, blackDiff, gameCount) *SessionManager`
- `Start()` — Creates and launches all game sessions as goroutines
- `Pause()` — Signals all sessions to pause
- `Resume()` — Signals all sessions to resume
- `SetSpeed(speed)` — Updates move interval for all sessions
- `Abort()` — Stops all sessions, cleans up engines
- `Sessions() []*GameSession` — Returns sessions for rendering
- `Stats() *AggregateStats` — Computes statistics from completed results
- `AllFinished() bool`

---

### 2.4 New Package: `internal/bvb/stats.go`

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
// Add to GameType enum:
GameTypeBvB GameType = 2

// Add new screen states:
ScreenBvBBotSelect      // Select white/black bot difficulties
ScreenBvBGameMode       // Single game or multi-game + count input
ScreenBvBGridConfig     // Grid size configuration
ScreenBvBGamePlay       // Watching games (separate from ScreenGamePlay)
ScreenBvBStats          // Statistics display

// Add to Model struct:
bvbManager        *bvb.SessionManager
bvbSpeed          bvb.PlaybackSpeed
bvbGridRows       int
bvbGridCols       int
bvbViewMode       BvBViewMode  // GridView or SingleView
bvbSelectedGame   int          // Which game is focused in single view
bvbPageIndex      int          // Current page in grid view
bvbWhiteDiff      BotDifficulty
bvbBlackDiff      BotDifficulty
bvbGameCount      int
bvbCountInput     string       // Text input for game count
bvbGridInput      string       // Text input for custom grid dimensions
```

---

### 2.6 UI Changes: Screen Flow (`internal/ui/bvb_screens.go`)

Navigation flow:
```
ScreenGameTypeSelect → "Bot vs Bot"
    → ScreenBvBBotSelect (select White difficulty, then Black difficulty)
    → ScreenBvBGameMode (Single Game / Multi-Game + count input)
    → ScreenBvBGridConfig (grid size presets or custom input)
    → ScreenBvBGamePlay (watching games)
    → ScreenBvBStats (statistics, option to replay or return to menu)
```

Each screen handler follows existing patterns:
- Arrow keys / j/k for navigation
- Enter to confirm
- ESC to go back to previous screen

---

### 2.7 UI Changes: Tick-Based Updates (`internal/ui/update.go`)

```go
// New message type:
type BvBTickMsg struct{}

// Tick command:
func bvbTickCmd(speed bvb.PlaybackSpeed) tea.Cmd {
    if speed == bvb.SpeedInstant {
        return func() tea.Msg { return BvBTickMsg{} }  // immediate re-render
    }
    return tea.Tick(speed.Duration(), func(time.Time) tea.Msg {
        return BvBTickMsg{}
    })
}

// In Update():
case BvBTickMsg:
    if m.bvbManager.AllFinished() {
        m.screen = ScreenBvBStats
        return m, nil
    }
    // Tick triggers re-render; goroutines advance games independently
    return m, bvbTickCmd(m.bvbSpeed)  // schedule next tick
```

**Instant mode:** Ticks fire immediately (no delay). Goroutines compute moves as fast as engines can. UI re-renders on each tick showing current progress.

---

### 2.8 UI Changes: Grid Rendering (`internal/ui/bvb_view.go`)

- Uses lipgloss `JoinHorizontal` and `JoinVertical` to compose boards in a grid
- Each board in the grid is a compact version (smaller than full-size, minimal info)
- **Grid board shows:** position, game number, move count, status indicator (ongoing/finished)
- **Single-board view shows:** full-size board, move history, bot names, detailed status
- Page indicator when games > grid slots (e.g., "Page 1/3 | ←/→ to navigate")
- Help text shows available controls (speed, pause, view toggle, abort, FEN export)
- Completed games visually distinguished (e.g., dimmed or bordered differently)

---

### 2.9 UI Changes: Key Bindings During BvB Gameplay

| Key | Action |
|-----|--------|
| Space | Pause / Resume all games |
| 1-4 | Set speed (1=Instant, 2=Fast, 3=Normal, 4=Slow) |
| Tab | Toggle grid view / single-board view |
| ←/→ | Navigate pages (grid) or games (single view) |
| f | Export FEN of selected/focused game |
| ESC | Abort session, return to menu |
| Ctrl+C | Exit application |

---

### 2.10 Goroutine Lifecycle

```
Start() called:
  ├── For each game (1..N):
  │   ├── Create white engine (factory)
  │   ├── Create black engine (factory)
  │   ├── Create GameSession
  │   └── go session.Run()
  │         ├── Loop:
  │         │   ├── Check stopCh → cleanup and return
  │         │   ├── Check pause → block on resumeCh
  │         │   ├── Call engine.SelectMove(ctx, board)
  │         │   ├── Apply move to board
  │         │   ├── Check game over → record result, close engines, return
  │         │   ├── Sleep for speed delay (0 for Instant)
  │         │   └── Continue
  │         └── Close engines on exit
  └── Manager tracks completion via session states
```

---

### 2.11 Files Summary

| File | Location | Purpose |
|------|----------|---------|
| `types.go` | `internal/bvb/` | PlaybackSpeed, SessionState, GameResult types |
| `session.go` | `internal/bvb/` | GameSession struct, Run() goroutine, thread-safe accessors |
| `manager.go` | `internal/bvb/` | SessionManager, orchestration, pause/resume/speed/abort |
| `stats.go` | `internal/bvb/` | AggregateStats, ComputeStats() |
| `bvb_screens.go` | `internal/ui/` | Screen handlers for BvB config flow |
| `bvb_view.go` | `internal/ui/` | Grid rendering, single-board view, stats view |
| `model.go` | `internal/ui/` | Extended with BvB fields, GameTypeBvB, new screen states |
| `update.go` | `internal/ui/` | BvBTickMsg handling, tick scheduling |
| `view.go` | `internal/ui/` | Route to BvB views for new screen states |

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
| Many goroutines with Hard bots (depth 6) causing high CPU | UI lag, system unresponsive | Context timeouts per move; engine time limits already enforced |
| Race conditions on shared board state | Panics, incorrect renders | Each session uses sync.Mutex; UI reads copies via thread-safe methods |
| Memory usage with many concurrent games | High memory for large counts | Each game is lightweight (~KB); practical limit is CPU, not memory |
| Engines never finishing (infinite loops) | Games hang indefinitely | Max move count per game (500 moves → forced draw); context timeout per move |
| Terminal too small for grid | Broken layout | Detect terminal size; fallback to smaller grid or single-board view if insufficient space |
| Goroutine leaks on abort | Resource leak | Explicit stopCh signaling; defer cleanup in Run(); manager tracks all sessions |

---

## 4. Testing Strategy

### Unit Tests (`internal/bvb/`)
- `session_test.go`: Single game runs to completion; pause/resume works; abort stops cleanly; max move limit enforced; thread-safe accessors return correct data
- `manager_test.go`: Multiple sessions run in parallel and all finish; pause/resume affects all sessions; speed change propagates; abort cleans up all goroutines
- `stats_test.go`: Statistics computation with known results (wins, draws, averages, shortest/longest)
- `types_test.go`: Speed duration mapping; state transitions

### Unit Tests (`internal/ui/`)
- `bvb_screens_test.go`: Screen navigation (ESC goes back, Enter advances); input validation for game count and grid dimensions
- `bvb_view_test.go`: Grid rendering produces expected layout; page calculation correct; single-board view shows move history

### Integration Tests
- Run 5 Easy vs Easy games, verify all complete without panics or goroutine leaks
- Run single Hard vs Hard game, verify timeout handling works
- Test pause during active games, resume, verify games continue correctly
- Test abort during active games, verify all goroutines cleaned up (no leaks)
- Test speed change mid-game propagates to all sessions

### Manual Testing
- Visual verification of grid layouts at all preset sizes (1x1, 2x2, 2x3, 2x4)
- Custom grid input validation
- Single-board navigation between games
- Statistics screen accuracy
- Terminal resize handling
- Help text display respects ShowHelpText config
