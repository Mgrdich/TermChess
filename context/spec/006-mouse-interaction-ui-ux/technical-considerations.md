# Technical Specification: Mouse Interaction & UI/UX Enhancements

- **Functional Specification:** `context/spec/006-mouse-interaction-ui-ux/functional-spec.md`
- **Status:** Draft
- **Author(s):** AI Assistant

---

## 1. High-Level Technical Approach

Phase 4 enhances TermChess with mouse interaction, visual themes, improved navigation, and Bot vs Bot optimizations. The implementation builds on the existing Bubbletea architecture:

1. **Mouse Interaction**: Enable `tea.MouseMsg` handling, track selection state in Model, implement blinking highlight effects via tick-based animation. Mouse only enabled for PvP and Player vs Bot modes.
2. **Theme System**: Create a new `Theme` struct with preset themes, integrate into Config, replace hardcoded styles
3. **Navigation**: Add navigation stack for breadcrumbs, implement global keyboard shortcuts, add help overlay
4. **Bot vs Bot**: Configurable concurrency with CPU auto-detection using tiered formula, simplified speed options, live statistics panel, game number jump, proper engine cleanup

All changes are contained within the existing Go codebase, primarily affecting `internal/ui/` and `internal/bvb/` packages.

---

## 2. Proposed Solution & Implementation Plan (The "How")

### 2.1 Mouse Interaction

**Scope**: Mouse interaction is **only enabled** for interactive game modes:
- **Player vs Player (1v1)**: Both players can use mouse
- **Player vs Bot (1 vs Bot)**: Human player can use mouse
- **Bot vs Bot**: Disabled - user is spectating, not playing

**New Model Fields** (`internal/ui/model.go`):
```go
type Model struct {
    // ... existing fields

    // Mouse interaction state
    selectedSquare  *engine.Square    // Currently selected piece
    validMoves      []engine.Square   // Legal destinations for selected piece
    blinkOn         bool              // Toggle for blink animation
}
```

**Update Handler** (`internal/ui/update.go`):
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.MouseMsg:
        // Only handle mouse in interactive game modes
        if m.screen == ScreenGamePlay && m.gameType != GameTypeBotVsBot {
            return m.handleMouseEvent(msg)
        }
        // Ignore mouse events in Bot vs Bot and other screens
        return m, nil
    // ... rest of cases
    }
}

func (m Model) isInteractiveGame() bool {
    return m.screen == ScreenGamePlay &&
           (m.gameType == GameTypePvP || m.gameType == GameTypeVsBot)
}
```

- Create `handleMouseEvent(msg tea.MouseMsg)` method
- Implement `squareFromMouse(x, y int) *engine.Square` helper
- Board position assumed fixed at (2, 1) with coordinate labels

**Blink Animation**:
- Create new `BlinkTickMsg` message type
- Add tick command that fires every 500ms when a piece is selected
- Toggle `blinkOn` boolean on each tick
- Stop tick when selection is cleared

**Board Renderer Updates** (`internal/ui/board.go`):
- Modify `Render()` to accept selection state: `Render(b *engine.Board, selected *engine.Square, validMoves []engine.Square, blinkOn bool)`
- Apply highlight styles when `blinkOn` is true for selected square and valid moves
- Use theme colors for highlights

**Mouse-to-Square Calculation**:
- Board starts at fixed position (column 2 after rank labels, row 1 after title)
- Each square is 2 characters wide (piece + space)
- Formula: `file = (mouseX - boardStartX) / 2`, `rank = 7 - (mouseY - boardStartY)`
- Validate coordinates are within 0-7 range

### 2.2 Theme System

**New File** (`internal/ui/theme.go`):
```go
type Theme struct {
    Name            string

    // Board colors
    LightSquare     lipgloss.Color
    DarkSquare      lipgloss.Color
    WhitePiece      lipgloss.Color
    BlackPiece      lipgloss.Color

    // Selection colors
    SelectedHighlight lipgloss.Color
    ValidMoveHighlight lipgloss.Color

    // UI colors
    BoardBorder     lipgloss.Color
    MenuSelected    lipgloss.Color
    MenuNormal      lipgloss.Color
    TitleText       lipgloss.Color
    HelpText        lipgloss.Color
    ErrorText       lipgloss.Color
    StatusText      lipgloss.Color

    // Turn indicator colors
    WhiteTurnText   lipgloss.Color
    BlackTurnText   lipgloss.Color
}

var (
    ClassicTheme = Theme{
        Name:              "Classic",
        LightSquare:       lipgloss.Color("#F0D9B5"),
        DarkSquare:        lipgloss.Color("#B58863"),
        WhitePiece:        lipgloss.Color("#FFFFFF"),
        BlackPiece:        lipgloss.Color("#000000"),
        SelectedHighlight: lipgloss.Color("#7D56F4"),
        ValidMoveHighlight: lipgloss.Color("#50FA7B"),
        // ... remaining colors
    }

    ModernTheme = Theme{...}
    MinimalistTheme = Theme{...}
)

func GetTheme(name string) Theme {
    switch name {
    case "modern":
        return ModernTheme
    case "minimalist":
        return MinimalistTheme
    default:
        return ClassicTheme
    }
}
```

**Config Updates** (`internal/config/config.go`):
- Add `Theme string` to `DisplayConfig` struct with TOML tag `theme`
- Default value: `"classic"`
- Update `configFileToConfig()` and `configToConfigFile()` conversion functions

**View Updates** (`internal/ui/view.go`):
- Add `theme Theme` field to Model
- Replace all hardcoded lipgloss styles with theme-based style getters
- Create style helper methods: `titleStyle()`, `menuItemStyle()`, `selectedItemStyle()`, etc.

**Turn Indicator**:
- Use `theme.WhiteTurnText` when `board.Turn() == engine.White`
- Use `theme.BlackTurnText` when `board.Turn() == engine.Black`
- Apply to move input prompt and status messages

### 2.3 Menu and Navigation Improvements

**Navigation Stack** (`internal/ui/model.go`):
```go
type Model struct {
    // ... existing fields

    navStack             []Screen  // Navigation history
    showShortcutsOverlay bool      // Display help overlay
}

func (m *Model) pushScreen(screen Screen) {
    m.navStack = append(m.navStack, m.screen)
    m.screen = screen
}

func (m *Model) popScreen() {
    if len(m.navStack) > 0 {
        m.screen = m.navStack[len(m.navStack)-1]
        m.navStack = m.navStack[:len(m.navStack)-1]
    } else {
        m.screen = ScreenMainMenu
    }
}

func (m Model) breadcrumb() string {
    // Returns "Main Menu > Bot vs Bot > Game 3"
}
```

**Global Keyboard Shortcuts** (`internal/ui/update.go`):
- Handle in `Update()` before screen-specific handlers
- `?` - Toggle shortcuts overlay (modal, press any key to dismiss)
- `Esc` - Call `popScreen()` (unless in specific contexts)
- Shortcuts only active when not in text input mode

**Shortcuts Overlay**:
- Full-screen modal rendered over current screen
- Displays all keyboard shortcuts organized by context
- Dismiss with any key press

**Menu Reorganization**:
- Group less-common options under "More..." submenu
- Add visual separators between option groups
- Ensure primary actions are prominently styled

### 2.4 Bot vs Bot Improvements

**Concurrency Formula & Experimentation** (`internal/bvb/manager.go`):

The optimal concurrency setting requires experimentation since goroutines are lightweight but engine calculations are CPU-bound:

```go
func calculateDefaultConcurrency() int {
    numCPU := runtime.NumCPU()

    // Option A: Conservative (NumCPU - 1)
    // Pros: Leaves headroom for UI, safe default
    // Cons: May underutilize on high-core systems

    // Option B: Moderate (NumCPU * 1.5)
    // Pros: Better utilization, goroutines can context-switch during I/O
    // Cons: May cause contention on CPU-bound minimax

    // Option C: Aggressive (NumCPU * 2)
    // Pros: Maximizes throughput if engines have any wait states
    // Cons: Risk of thrashing on pure CPU workloads

    // Option D: Tiered approach based on core count (RECOMMENDED)
}
```

**Recommended Approach - Tiered with Experimentation**:
```go
func calculateDefaultConcurrency() int {
    numCPU := runtime.NumCPU()

    var concurrency int
    switch {
    case numCPU <= 2:
        concurrency = numCPU
    case numCPU <= 4:
        concurrency = int(float64(numCPU) * 1.5)
    default:
        concurrency = numCPU * 2
    }

    // Cap at reasonable maximum
    if concurrency > maxConcurrentGames {
        concurrency = maxConcurrentGames
    }
    if concurrency < 1 {
        concurrency = 1
    }
    return concurrency
}
```

**Goroutine Optimization Notes**:
- The number of concurrent game goroutines should **correlate with** but **not equal** the CPU core count
- Rationale: Each game spawns goroutines for engine calculations (minimax search), so the actual goroutine count is a multiple of the concurrency setting
- The tiered formula accounts for different system sizes
- For CPU-intensive minimax bots, fewer concurrent games with deeper searches often outperforms many shallow concurrent games
- Consider adding `runtime.GOMAXPROCS` awareness if users report issues on systems with many cores

**Testing Matrix for Concurrency**:
| CPU Cores | Formula | Concurrency | Test Result |
|-----------|---------|-------------|-------------|
| 2 | NumCPU | 2 | [TBD] |
| 4 | NumCPU * 1.5 | 6 | [TBD] |
| 8 | NumCPU * 2 | 16 | [TBD] |
| 16 | NumCPU * 2 | 32 | [TBD] |

**Implementation Note**: Run benchmarks during development to validate the tiered formula. Metrics to track:
- Total time to complete N games
- CPU utilization percentage
- UI responsiveness (input lag)
- Memory usage over time

**Config Addition** (`internal/config/config.go`):
- Add `BvBConcurrency int` to `GameConfig` with TOML tag `bvb_concurrency`
- Default value: `0` (auto-detect)

**Speed Simplification** (`internal/bvb/types.go`):
- Remove `SpeedFast` and `SpeedSlow` constants
- Keep only `SpeedNormal` (1 second delay) and `SpeedInstant` (0 delay)
- Update UI to show only two speed options

**Game Number Jump** (`internal/ui/update.go`):
- Add `bvbJumpInput string` and `bvbShowJumpPrompt bool` to Model
- Handle `g` key to show jump prompt
- Parse input and navigate to valid game number
- Show error for invalid input

**Live Statistics Panel** (`internal/ui/view.go`):
- Create `renderBvBStats()` function
- Display during `ScreenBvBGamePlay`:
  - Score: White Wins / Black Wins / Draws
  - Games: Completed / Total
  - Avg moves per game
  - Current game duration
  - Longest/shortest game (moves)
  - Current game move history (last 10 moves)
  - Captured pieces for current game
- Update stats on each `BvBTickMsg`

**Engine Lifecycle Management** (`internal/bvb/session.go` and `manager.go`):

```go
// In GameSession - cleanup after game completion
func (s *GameSession) Run() {
    defer s.cleanup()  // Ensure cleanup runs even on panic

    // ... game loop ...
}

func (s *GameSession) cleanup() {
    // Destroy engines to free resources
    if s.whiteEngine != nil {
        if closer, ok := s.whiteEngine.(io.Closer); ok {
            closer.Close()
        }
        s.whiteEngine = nil
    }
    if s.blackEngine != nil {
        if closer, ok := s.blackEngine.(io.Closer); ok {
            closer.Close()
        }
        s.blackEngine = nil
    }
}

// In SessionManager - cleanup all sessions when session ends
func (m *SessionManager) Stop() {
    close(m.abortCh)  // Signal all goroutines to stop

    // Wait for active games to finish, then cleanup
    m.mu.Lock()
    defer m.mu.Unlock()

    for _, session := range m.sessions {
        if session != nil {
            session.cleanup()
        }
    }
    m.sessions = nil
}
```

**Engine Interface Update** (`internal/bot/engine.go`):
- Consider adding `io.Closer` to the `Engine` interface or create a separate `CloseableEngine` interface
- Minimax engines should release any allocated resources (evaluation caches, transposition tables)
- Future UCI engines will need proper process termination

**Memory Leak Prevention Checklist**:
- [ ] Each `GameSession` calls `cleanup()` via defer when `Run()` completes
- [ ] `SessionManager.Stop()` cleans up all sessions when user exits Bot vs Bot mode
- [ ] Engines are set to `nil` after cleanup to allow garbage collection
- [ ] Abort channel properly signals waiting goroutines to exit
- [ ] No references to completed sessions are retained in the manager

**Stats-Only Mode** (`internal/ui/model.go` and `internal/ui/view.go`):

Stats-Only mode addresses terminal performance issues when running high-concurrency Bot vs Bot sessions. Rendering many boards simultaneously causes terminal lag; Stats-Only mode eliminates this bottleneck.

**New Screen** (`internal/ui/model.go`):
```go
const (
    // ... existing screens
    ScreenBvBViewModeSelect  // New: Select view mode before starting session
)
```

**Updated Bot vs Bot Flow**:
```
ScreenBvBBotSelect (select White/Black bot difficulties)
    ↓
ScreenBvBGameMode (single game or multi-game)
    ↓ (if multi-game)
ScreenBvBGridConfig (select game count)
    ↓
ScreenBvBViewModeSelect (NEW: select Grid/Single/Stats Only)
    ↓
ScreenBvBGamePlay (session starts with selected view mode)
```

**View Mode Selection Screen**:
- Menu options:
  1. "Grid View" - Watch multiple games in a grid layout
  2. "Single Board" - Focus on one game at a time
  3. "Stats Only (Recommended for 50+ games)" - No boards, just statistics
- Arrow keys to navigate, Enter to select
- Esc to go back to game count input

```go
type BvBViewMode int

const (
    BvBGridView BvBViewMode = iota
    BvBSingleView
    BvBStatsOnlyView  // New: No board rendering
)
```

**Model Fields**:
```go
type Model struct {
    // ... existing fields
    bvbViewMode BvBViewMode  // Extended to include BvBStatsOnlyView
}
```

**Stats-Only View Rendering** (`internal/ui/view.go`):
```go
func (m Model) renderBvBStatsOnly() string {
    // Render comprehensive statistics without any board visualization
    // - Session title and configuration
    // - Progress bar: [████████░░░░░░░░] 45% (45/100 games)
    // - Score summary: White: 20 | Black: 15 | Draws: 10
    // - Average moves per completed game
    // - Estimated time remaining (based on avg game duration)
    // - Current active games indicator (e.g., "12 games in progress")
    // - Recent completions log (last 5 game results)
}
```

**View Mode Toggle**:
- `v` key cycles through view modes: Grid → Single → Stats Only → Grid
- Mode can be changed during active session
- Current mode persists until user changes it or session ends

**Performance Benefits**:
- No board string rendering (each board is ~20 lines × 30 chars)
- No per-game state tracking for display purposes
- Reduced terminal I/O (single stats panel vs. multiple boards)
- Allows safely running 50+ concurrent games without terminal lag

**Config Integration** (`internal/config/config.go`):
```go
type GameConfig struct {
    // ... existing fields
    BvBDefaultViewMode string `toml:"bvb_default_view_mode"` // "grid", "single", "stats_only"
}
```

**Grid Layout Stability** (`internal/ui/view.go`):

The grid view for Bot vs Bot multi-game sessions must maintain stable board positions. When games complete at different times and result text is added, boards should not shift vertically.

**Problem**: Currently, when a game ends, result text (e.g., "White wins by checkmate") is appended below the board, causing that cell to grow and pushing other content down.

**Solution - Fixed Cell Heights**:
```go
const (
    // Board is 8 rows + 2 for coordinates = 10 lines
    // Plus 1 line for game number header
    // Plus 2 lines for status/result text (always reserved)
    // Plus 1 line for spacing
    bvbCellHeight = 14
)

func (m Model) renderBvBGridCell(gameIndex int) string {
    var lines []string

    // Add game header (1 line)
    lines = append(lines, fmt.Sprintf("Game %d", gameIndex+1))

    // Add board (10 lines with coordinates)
    boardLines := strings.Split(boardStr, "\n")
    lines = append(lines, boardLines...)

    // Add status line (always 1 line, even if empty)
    lines = append(lines, statusText)

    // Add result line (always 1 line, even if game in progress)
    if game.IsOver() {
        lines = append(lines, resultText)
    } else {
        lines = append(lines, "") // Empty placeholder
    }

    // Pad to fixed height if needed
    for len(lines) < bvbCellHeight {
        lines = append(lines, "")
    }

    return strings.Join(lines[:bvbCellHeight], "\n")
}
```

**Key Implementation Points**:
- Define a constant `bvbCellHeight` that accounts for all possible content
- Always reserve space for result text, even when game is in progress (use empty line)
- Truncate or pad each cell to exactly `bvbCellHeight` lines
- Use `lipgloss.Height()` to verify consistent cell heights
- Apply same fixed width per cell using `lipgloss.Width()`

**Grid Assembly**:
```go
func (m Model) renderBvBGrid() string {
    // Render each cell with fixed dimensions
    // Use lipgloss.JoinHorizontal for rows
    // Use lipgloss.JoinVertical for the full grid
    // Each cell is padded/truncated to bvbCellHeight × bvbCellWidth
}
```

**Bot vs Bot Statistics Export** (`internal/bvb/export.go`):

Statistics export allows users to save session results and game data to a file for later review or analysis.

**Data Structure**:
```go
type SessionExport struct {
    Timestamp       time.Time           `json:"timestamp"`
    WhiteBot        string              `json:"white_bot"`       // e.g., "Easy", "Medium", "Hard"
    BlackBot        string              `json:"black_bot"`
    TotalGames      int                 `json:"total_games"`
    WhiteWins       int                 `json:"white_wins"`
    BlackWins       int                 `json:"black_wins"`
    Draws           int                 `json:"draws"`
    AverageMoves    float64             `json:"average_moves"`
    Games           []GameExport        `json:"games"`
}

type GameExport struct {
    GameNumber      int                 `json:"game_number"`
    Result          string              `json:"result"`          // "White", "Black", "Draw"
    TerminationReason string            `json:"termination"`     // "Checkmate", "Stalemate", "Insufficient Material", etc.
    MoveCount       int                 `json:"move_count"`
    Moves           []string            `json:"moves"`           // Standard algebraic notation
    FinalFEN        string              `json:"final_fen"`       // Final position
}
```

**Export Function**:
```go
func (m *SessionManager) ExportStats() (*SessionExport, error) {
    // Collect all game data from completed sessions
    // Build SessionExport struct
    // Return for serialization
}

func SaveSessionExport(export *SessionExport, dir string) (string, error) {
    // Create directory if not exists
    // Generate filename with timestamp
    // Marshal to JSON with indentation
    // Write to file
    // Return filepath
}
```

**File Location**:
- Default directory: `~/.termchess/stats/`
- Create directory if it doesn't exist
- Filename format: `bvb_session_YYYY-MM-DD_HH-mm-ss.json`

**UI Integration** (`internal/ui/update.go`):
- On BvB stats screen, handle `s` key to trigger save
- Call `ExportStats()` to gather data
- Call `SaveSessionExport()` to write file
- Display success message with filepath or error message

**Move History Collection**:
- Each `GameSession` should store moves as they're made
- Convert moves to standard algebraic notation for export
- Store termination reason when game ends

### 2.5 Accessibility

**WCAG AA Compliance**:
- All three themes will use colors meeting 4.5:1 contrast ratio for text
- Use online contrast checker during theme design
- Test with terminal color schemes (light and dark backgrounds)

**Keyboard Navigation**:
- Ensure every mouse-interactive element has keyboard equivalent
- Focus indicators via distinct styling for selected items
- Tab order follows logical UI flow

---

## 3. Impact and Risk Analysis

### System Dependencies

| Component | Dependencies | Impact |
|-----------|--------------|--------|
| Mouse Interaction | `engine.Board` (move validation), `BoardRenderer` | Medium - New feature, isolated changes |
| Theme System | `Config`, all UI rendering | High - Touches all view code |
| Navigation | `Model`, `Update`, `View` | Medium - Cross-cutting but isolated |
| BvB Improvements | `SessionManager`, `Config`, UI | Medium - Changes to existing feature |
| BvB Stats-Only Mode | `Model`, `View`, `Config` | Low - New view mode, additive change |

### Potential Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Mouse coordinate calculation errors | Medium | Medium | Extensive manual testing across terminal sizes; add bounds checking |
| Theme color contrast issues | Medium | Low | Use contrast checker tools; test on multiple terminal themes |
| Blink animation performance | Low | Low | Use efficient tick handling; disable blink when not needed |
| Breaking existing keyboard navigation | Medium | High | Comprehensive regression testing; keep existing shortcuts working |
| BvB concurrency regression | Low | Medium | Preserve existing behavior when concurrency=0 (auto); add unit tests |
| Memory leaks from undestroyed engines | High (if not addressed) | High | Implement `cleanup()` with defer pattern; nil out engine references; add `io.Closer` interface support |
| Stats-only mode missing critical info | Low | Low | Include all essential statistics; allow toggling back to board view |

---

## 4. Testing Strategy

### Unit Tests

| Component | Test Coverage |
|-----------|---------------|
| `squareFromMouse()` | Table-driven tests for various mouse positions, edge cases, out-of-bounds |
| `Theme.GetTheme()` | All theme names resolve correctly, default fallback works |
| `Model.pushScreen/popScreen` | Stack behavior, empty stack handling |
| `SessionManager` concurrency | Auto-detection logic (tiered formula), manual override, bounds checking |
| `GameSession.cleanup()` | Engines properly nil'd, Close() called if implemented |
| `BvBViewMode` toggle | Cycling through Grid → Single → Stats Only → Grid |
| `renderBvBStatsOnly()` | Output contains score, progress, avg moves |
| `renderBvBGridCell()` | Output has exactly `bvbCellHeight` lines |
| `ExportStats()` | Returns valid SessionExport with all game data |
| `SaveSessionExport()` | Creates file with correct JSON format |

### Integration Tests

| Scenario | Verification |
|----------|--------------|
| Theme persistence | Change theme, restart app, verify theme loads |
| BvB with concurrency | Run multi-game session, verify games complete correctly |
| Config file migration | Old config files load without errors, new defaults applied |
| BvB memory cleanup | Run session, exit, verify no memory growth |

### Manual Testing

| Area | Test Cases |
|------|------------|
| Mouse interaction | Select piece, view highlights, make move, invalid clicks, edge squares |
| Mouse scope | Verify mouse works in PvP and vs Bot, disabled in Bot vs Bot |
| Blink animation | Timing correct (~500ms), stops when deselected |
| All three themes | Visual appearance, contrast, piece visibility |
| Keyboard shortcuts | All shortcuts work, overlay displays correctly |
| BvB statistics | All stats display and update correctly |
| Breadcrumb navigation | Back navigation works from all screens |
| Concurrency testing | Test on 2, 4, 8, 16 core systems with various multipliers |
| BvB stats-only mode | View toggle works, stats update correctly, no terminal lag at high concurrency |
| BvB grid layout stability | Boards don't shift when games end, consistent cell heights across all states |
| BvB statistics export | Save works, file contains all data, move history is correct |

### Benchmarks

| Benchmark | Purpose |
|-----------|---------|
| BvB 10-game session (Easy vs Easy) | Baseline performance |
| BvB 10-game session (Hard vs Hard) | CPU-intensive scenario |
| Concurrency sweep (1x, 1.5x, 2x, 3x NumCPU) | Find optimal multiplier |
| Memory profiling during BvB | Verify no leaks after cleanup |
| BvB 100-game session stats-only mode | Verify high concurrency stability |
| Stats-only vs Grid view terminal I/O | Compare rendering overhead |
