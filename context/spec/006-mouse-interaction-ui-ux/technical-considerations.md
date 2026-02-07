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

**Terminal Resize Handling** (`internal/ui/model.go` and `internal/ui/update.go`):

The application must handle terminal resize events to ensure all content fits on screen.

**Model Fields** (already exist):
```go
type Model struct {
    // ... existing fields
    termWidth  int  // Current terminal width in characters
    termHeight int  // Current terminal height in lines
}
```

**Resize Event Handling** (`internal/ui/update.go`):
```go
case tea.WindowSizeMsg:
    m.termWidth = msg.Width
    m.termHeight = msg.Height

    // Adjust BvB grid if needed
    if m.screen == ScreenBvBGamePlay && m.gameType == GameTypeBvB {
        m.adjustBvBGridForWidth()
    }

    return m, nil
```

**Grid Auto-Adjustment** (`internal/ui/view.go`):
```go
const (
    minBoardWidth    = 20  // Minimum width for a single board
    bvbCellWidth     = 25  // Width per board cell including padding
    minTerminalWidth = 40  // Minimum usable terminal width
    minTerminalHeight = 20 // Minimum usable terminal height
)

func (m *Model) adjustBvBGridForWidth() {
    if m.termWidth < minTerminalWidth {
        return // Will show warning instead
    }

    // Calculate max columns that fit
    maxCols := m.termWidth / bvbCellWidth
    if maxCols < 1 {
        maxCols = 1
    }

    // If current grid doesn't fit, reduce columns
    if m.bvbGridCols > maxCols {
        m.bvbGridCols = maxCols
        // Recalculate rows to maintain game count
        m.bvbGridRows = (m.bvbGameCount + maxCols - 1) / maxCols
    }

    // If grid can't fit even 1 column, switch to single view
    if m.termWidth < bvbCellWidth {
        m.bvbViewMode = BvBSingleView
    }
}
```

**Minimum Size Warning** (`internal/ui/view.go`):
```go
func (m Model) renderMinSizeWarning() string {
    return lipgloss.NewStyle().
        Foreground(m.theme.ErrorText).
        Bold(true).
        Render("Terminal too small! Resize to at least 40x20")
}
```

**View Integration**:
- Check `termWidth` and `termHeight` at start of `View()` function
- If below minimum, return `renderMinSizeWarning()` instead of normal view
- Pass terminal dimensions to render functions that need responsive layout

**Responsive Elements**:
- Board: Always fixed size (~20 chars wide with coordinates)
- BvB Grid: Variable columns based on width
- Menus: Truncate long options with "..." if needed
- Stats panel: Collapse to narrower format if needed
- Help text: Wrap or hide on very narrow terminals

**Bot vs Bot Concurrency Selection Screen** (`internal/ui/model.go` and `internal/ui/update.go`):

This screen allows users to choose between auto-detected concurrency or enter a custom value with no upper limit.

**New Screen** (`internal/ui/model.go`):
```go
const (
    // ... existing screens
    ScreenBvBConcurrencySelect  // New: Select concurrency before view mode
)
```

**Model Fields**:
```go
type Model struct {
    // ... existing fields
    bvbConcurrencySelection int  // 0 = Recommended, 1 = Custom
    bvbCustomConcurrency    string  // Text input for custom value
    bvbInputtingConcurrency bool    // True when typing custom value
}
```

**Updated Bot vs Bot Flow**:
```
ScreenBvBBotSelect (select White/Black bot difficulties)
    ↓
ScreenBvBGameMode (single game or multi-game)
    ↓ (if multi-game)
ScreenBvBGridConfig (select game count)
    ↓
ScreenBvBConcurrencySelect (NEW: select Recommended or Custom concurrency)
    ↓
ScreenBvBViewModeSelect (select Grid/Single/Stats Only)
    ↓
ScreenBvBGamePlay (session starts)
```

**Concurrency Selection Screen Rendering** (`internal/ui/view.go`):
```go
func (m Model) renderBvBConcurrencySelect() string {
    // Title: "Select Concurrency"
    //
    // Options:
    //   > Recommended (X concurrent games)
    //     Based on your CPU (Y cores)
    //
    //     Custom
    //     Enter your own value (may cause lag)
    //
    // If Custom selected and inputting:
    //   Enter concurrency: [input field]
    //
    // If custom value > 50:
    //   ⚠ High concurrency may cause lag. Consider using Stats Only view mode.
    //
    // Help: arrows: navigate | enter: select | esc: back
}
```

**Key Handler** (`internal/ui/update.go`):
```go
func (m Model) handleBvBConcurrencySelectKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
    if m.bvbInputtingConcurrency {
        // Handle text input for custom value
        switch msg.String() {
        case "enter":
            concurrency, err := parsePositiveInt(m.bvbCustomConcurrency)
            if err != nil {
                m.statusMessage = "Please enter a valid positive number"
                return m, nil
            }
            m.bvbConcurrency = concurrency
            m.bvbInputtingConcurrency = false
            return m.transitionToViewModeSelect()
        case "esc":
            m.bvbInputtingConcurrency = false
            m.bvbCustomConcurrency = ""
            return m, nil
        default:
            // Append digits only
            if isDigit(msg.String()) {
                m.bvbCustomConcurrency += msg.String()
            }
            // Handle backspace
            if msg.String() == "backspace" && len(m.bvbCustomConcurrency) > 0 {
                m.bvbCustomConcurrency = m.bvbCustomConcurrency[:len(m.bvbCustomConcurrency)-1]
            }
        }
        return m, nil
    }

    // Menu navigation
    switch msg.String() {
    case "up", "k":
        m.bvbConcurrencySelection = 0  // Recommended
    case "down", "j":
        m.bvbConcurrencySelection = 1  // Custom
    case "enter":
        if m.bvbConcurrencySelection == 0 {
            // Use recommended (auto-calculated)
            m.bvbConcurrency = calculateDefaultConcurrency()
            return m.transitionToViewModeSelect()
        } else {
            // Switch to custom input mode
            m.bvbInputtingConcurrency = true
            m.bvbCustomConcurrency = ""
        }
    case "esc":
        return m.popScreen()
    }
    return m, nil
}
```

**Removing the Hard Cap**:
```go
// In internal/bvb/manager.go

// Remove or make configurable:
// const maxConcurrentGames = 50

// Update NewSessionManager to accept any concurrency value:
func NewSessionManager(..., concurrency int) *SessionManager {
    // No longer cap at 50 - user chose this value knowingly
    if concurrency < 1 {
        concurrency = 1
    }
    // ... rest of initialization
}
```

**Warning Display Logic**:
- If user enters custom value > 50, show amber warning text
- Warning suggests using Stats Only mode but doesn't prevent proceeding
- Warning disappears if user changes to value <= 50

### 2.5 Accessibility

**WCAG AA Compliance**:
- All three themes will use colors meeting 4.5:1 contrast ratio for text
- Use online contrast checker during theme design
- Test with terminal color schemes (light and dark backgrounds)

**Keyboard Navigation**:
- Ensure every mouse-interactive element has keyboard equivalent
- Focus indicators via distinct styling for selected items
- Tab order follows logical UI flow

### 2.6 Navigation Stack and Linear Back-Navigation

This section addresses the requirement for consistent, mobile-like back navigation across all screens.

**Problem Statement:**
Currently, navigation is inconsistent:
- Some screens use `pushScreen()`/`popScreen()` (GameTypeSelect, BotSelect, ColorSelect, FENInput, Settings)
- BvB screens use direct assignment (`m.screen = ScreenXYZ`)
- ESC handlers are mixed - some use `popScreen()`, others hardcode targets
- `ScreenResumePrompt` is dead code (deprecated but still in codebase)

**Solution: Unified Stack-Based Navigation**

All screen transitions will use the navigation stack, ensuring ESC always returns to the previous screen in exact order.

**Navigation Stack Implementation** (already exists in `internal/ui/navigation.go`):
```go
func (m *Model) pushScreen(screen Screen) {
    if m.screen == screen {
        return // Don't push if already on same screen
    }
    m.navStack = append(m.navStack, m.screen)
    m.screen = screen
}

func (m *Model) popScreen() Screen {
    if len(m.navStack) == 0 {
        m.screen = ScreenMainMenu
        return ScreenMainMenu
    }
    lastIndex := len(m.navStack) - 1
    previousScreen := m.navStack[lastIndex]
    m.navStack = m.navStack[:lastIndex]
    m.screen = previousScreen
    return previousScreen
}
```

**Screen Transition Changes** (`internal/ui/update.go`):

| Location | Current Code | New Code |
|----------|--------------|----------|
| Line 395 (GameType→BvBBotSelect) | `m.screen = ScreenBvBBotSelect` | `m.pushScreen(ScreenBvBBotSelect)` |
| Line 1236 (BvBBot→GameMode) | `m.screen = ScreenBvBGameMode` | `m.pushScreen(ScreenBvBGameMode)` |
| Line 1334 (GameMode→GridConfig) | `m.screen = ScreenBvBGridConfig` | `m.pushScreen(ScreenBvBGridConfig)` |
| Line 1474 (GridConfig→ViewMode) | `m.screen = ScreenBvBViewModeSelect` | `m.pushScreen(ScreenBvBViewModeSelect)` |
| NEW (GridConfig→ConcurrencySelect) | N/A | `m.pushScreen(ScreenBvBConcurrencySelect)` |
| NEW (ConcurrencySelect→ViewMode) | N/A | `m.pushScreen(ScreenBvBViewModeSelect)` |

**ESC Handler Changes** (`internal/ui/update.go`):

Replace all hardcoded ESC handlers with `popScreen()`:

```go
// Before (example from handleBvBViewModeSelectKeys):
case "esc":
    m.screen = ScreenBvBGridConfig  // Hardcoded!
    m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
    m.menuSelection = 0

// After:
case "esc":
    m.popScreen()
    // Menu options will be set by the screen's init/render logic
```

**Screens requiring ESC handler updates:**
- `handleBvBBotSelectKeys()` - Line 1190-1204
- `handleBvBGameModeKeys()` - Line 1275-1282
- `handleBvBGridConfigKeys()` - Line 1399-1404
- `handleBvBViewModeSelectKeys()` - Line 1516-1523
- NEW: `handleBvBConcurrencySelectKeys()`

**BvB Bot Select Special Case:**
The BvB Bot Select screen has internal state (`bvbSelectingWhite` flag) for White/Black selection. This remains a single screen:

```go
func (m Model) handleBvBBotSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    // ...
    case "esc":
        if m.bvbSelectingWhite {
            // At White selection - go back to previous screen
            m.popScreen()
        } else {
            // At Black selection - go back to White selection (same screen)
            m.bvbSelectingWhite = true
            m.menuSelection = 0
        }
    // ...
}
```

**Updated BvB Multi-Game Flow with Stack:**

```
Main Menu
  ↓ pushScreen(GameTypeSelect)
Game Type Select
  ↓ pushScreen(BvBBotSelect)
BvB Bot Select (White → Black internal toggle)
  ↓ pushScreen(BvBGameMode)
BvB Game Mode
  ↓ pushScreen(BvBGridConfig)  [if multi-game]
BvB Grid Config
  ↓ pushScreen(BvBConcurrencySelect)
BvB Concurrency Select
  ↓ pushScreen(BvBViewModeSelect)
BvB View Mode Select
  ↓ clearNavStack() + startBvBSession()
BvB Game Play (stack cleared - terminal destination)
```

**Gameplay as Terminal Destination:**
When entering gameplay (PvP, PvBot, or BvB), the stack is cleared:

```go
func (m Model) startBvBSession() (tea.Model, tea.Cmd) {
    // ... session setup ...
    m.clearNavStack()  // Gameplay is terminal - no back navigation
    m.screen = ScreenBvBGamePlay
    // ...
}
```

**Save/Quit Dialog Changes:**

Current behavior: ESC during gameplay shows Save Prompt, "No" returns to gameplay.

New behavior per spec:
- ESC during gameplay shows Save/Quit dialog
- "Yes" = Save game and return to Main Menu
- "No" = Return to Main Menu WITHOUT saving
- ESC on dialog = Cancel and return to gameplay

```go
func (m Model) handleSavePromptKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "enter":
        if m.savePromptSelection == 0 { // "Yes"
            // Save the game
            err := config.SaveGame(m.board)
            if err != nil {
                m.errorMsg = fmt.Sprintf("Failed to save: %v", err)
                return m, nil
            }
        }
        // Both Yes and No go to Main Menu
        m.cleanupGame()
        m.screen = ScreenMainMenu
        m.menuOptions = buildMainMenuOptions()
        m.menuSelection = 0

    case "esc":
        // Cancel - return to gameplay
        m.screen = ScreenGamePlay
        m.errorMsg = ""
    }
    return m, nil
}
```

**BvB Abort Confirmation Dialog:**

New dialog when ESC pressed during active BvB multi-game session:

```go
type Model struct {
    // ... existing fields
    bvbShowAbortConfirm bool
    bvbAbortSelection   int  // 0 = Cancel, 1 = Abort
}

func (m Model) handleBvBGamePlayKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    // If abort dialog is showing, handle it
    if m.bvbShowAbortConfirm {
        return m.handleBvBAbortConfirmKeys(msg)
    }

    switch msg.String() {
    case "esc":
        // Check if session is still running
        if m.bvbManager != nil && !m.bvbManager.AllFinished() {
            // Show abort confirmation
            m.bvbShowAbortConfirm = true
            m.bvbAbortSelection = 0  // Default to Cancel
            return m, nil
        }
        // Session finished - go directly to stats or menu
        m.screen = ScreenBvBStats
    // ...
    }
}

func (m Model) handleBvBAbortConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "up", "down", "k", "j":
        m.bvbAbortSelection = 1 - m.bvbAbortSelection  // Toggle 0/1

    case "enter":
        if m.bvbAbortSelection == 1 { // "Abort"
            m.bvbManager.Abort()
            m.bvbManager = nil
            m.bvbShowAbortConfirm = false
            m.screen = ScreenMainMenu
            m.menuOptions = buildMainMenuOptions()
        } else { // "Cancel"
            m.bvbShowAbortConfirm = false
        }

    case "esc":
        // ESC on dialog = Cancel
        m.bvbShowAbortConfirm = false
    }
    return m, nil
}
```

**Abort Confirmation Renderer** (`internal/ui/view.go`):
```go
func (m Model) renderBvBAbortConfirm() string {
    // Overlay dialog:
    // ┌─────────────────────────────┐
    // │     Abort Session?          │
    // │                             │
    // │  Games in progress will     │
    // │  be lost.                   │
    // │                             │
    // │  > Cancel                   │
    // │    Abort Session            │
    // │                             │
    // │  esc: cancel | enter: select│
    // └─────────────────────────────┘
}
```

**ScreenResumePrompt Removal:**

Remove all references to the deprecated `ScreenResumePrompt`:

| File | Changes |
|------|---------|
| `internal/ui/model.go` | Remove `ScreenResumePrompt` from Screen constants |
| `internal/ui/update.go` | Remove `case ScreenResumePrompt:` from `handleKeyPress()` |
| `internal/ui/update.go` | Delete `handleResumePromptKeys()` function (~100 lines) |
| `internal/ui/view.go` | Remove `case ScreenResumePrompt:` from `View()` |
| `internal/ui/view.go` | Delete `renderResumePrompt()` function |
| `internal/ui/navigation.go` | Remove `case ScreenResumePrompt:` from `screenName()` |
| Test files | Update tests that reference ScreenResumePrompt |

**Breadcrumb Consistency:**

With all screens using `pushScreen()`, breadcrumbs will automatically reflect the navigation path:

```go
func (m Model) breadcrumb() string {
    if m.screen == ScreenMainMenu || len(m.navStack) == 0 {
        return ""
    }
    // Show: "Parent > Current"
    if len(m.navStack) > 0 {
        parent := m.navStack[len(m.navStack)-1]
        return screenName(parent) + " > " + screenName(m.screen)
    }
    return screenName(m.screen)
}
```

**Menu State Restoration:**

When using `popScreen()`, the previous screen's menu state needs restoration. Options:

1. **Store menu state in stack** (complex, more memory)
2. **Rebuild menu on screen entry** (simpler, recommended)

Recommended approach - each screen handler rebuilds its menu options:

```go
func (m *Model) popScreen() Screen {
    // ... existing pop logic ...

    // Restore menu state based on returned screen
    switch m.screen {
    case ScreenGameTypeSelect:
        m.menuOptions = []string{"Player vs Player", "Player vs Bot", "Bot vs Bot"}
    case ScreenBvBBotSelect:
        m.menuOptions = []string{"Easy", "Medium", "Hard"}
    case ScreenBvBGameMode:
        m.menuOptions = []string{"Single Game", "Multi-Game"}
    case ScreenBvBGridConfig:
        m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
    case ScreenBvBConcurrencySelect:
        m.menuOptions = []string{
            fmt.Sprintf("Recommended (%d concurrent)", bvb.CalculateDefaultConcurrency()),
            "Custom",
        }
    case ScreenBvBViewModeSelect:
        m.menuOptions = []string{"Grid View", "Single Board", "Stats Only"}
    case ScreenMainMenu:
        m.menuOptions = buildMainMenuOptions()
    }
    m.menuSelection = 0
    m.errorMsg = ""

    return m.screen
}
```

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
| Terminal Resize | `Model`, `Update`, `View` | Medium - Affects all screen rendering |
| BvB Concurrency Select | `Model`, `Update`, `View`, `SessionManager` | Medium - New screen, removes hard cap |
| Navigation Stack Refactor | All screen handlers, `navigation.go` | High - Touches all ESC handlers and screen transitions |
| ScreenResumePrompt Removal | `model.go`, `update.go`, `view.go`, tests | Low - Removing dead code |
| BvB Abort Confirmation | `Model`, `View`, `SessionManager` | Low - New dialog, additive |
| Save/Quit Dialog Changes | `update.go` (SavePrompt handler) | Medium - Changes existing behavior |

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
| Layout breaks on small terminals | Medium | Medium | Define minimum size, show warning, auto-adjust grid columns |
| User sets very high concurrency (100+) | Medium | Medium | Show warning, recommend Stats Only mode; user accepts responsibility |
| Navigation stack refactor breaks existing flows | Medium | High | Comprehensive regression tests for all screen transitions; test ESC from every screen |
| Menu state not restored on popScreen | Medium | Medium | Implement menu restoration in `popScreen()` for all screen types |
| ScreenResumePrompt removal breaks tests | Low | Low | Update all test files that reference ScreenResumePrompt |
| Save/Quit behavior change confuses users | Low | Medium | Clear dialog text explaining "No" goes to menu without saving |
| BvB abort dialog interrupts spectating | Low | Low | Make dialog dismissible with ESC; only show when games are in progress |

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
| `adjustBvBGridForWidth()` | Grid columns reduce when terminal narrows |
| `renderMinSizeWarning()` | Warning displays for small terminals |
| `handleBvBConcurrencySelectKeys()` | Navigation, selection, custom input validation |
| Custom concurrency input | Accepts any positive integer, no upper limit |
| `popScreen()` menu restoration | Each screen type gets correct menu options restored |
| `handleSavePromptKeys()` | Yes saves + goes to menu, No goes to menu without saving, ESC returns to game |
| `handleBvBAbortConfirmKeys()` | Cancel returns to session, Abort stops session + goes to menu |
| Navigation stack linear flow | ESC from each BvB screen returns to correct previous screen |
| ScreenResumePrompt removal | No references to ScreenResumePrompt in production code |

### Integration Tests

| Scenario | Verification |
|----------|--------------|
| Theme persistence | Change theme, restart app, verify theme loads |
| BvB with concurrency | Run multi-game session, verify games complete correctly |
| Config file migration | Old config files load without errors, new defaults applied |
| BvB memory cleanup | Run session, exit, verify no memory growth |
| Full BvB navigation flow | Navigate MainMenu → ... → ViewModeSelect, ESC back through entire flow to MainMenu |
| Full PvBot navigation flow | Navigate MainMenu → GameType → BotSelect → ColorSelect, ESC back to MainMenu |
| Save/Quit flow | During gameplay, press ESC, select No, verify at MainMenu without save |
| BvB abort flow | During multi-game session, press ESC, select Abort, verify session stopped |

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
| Terminal resize | Grid adjusts on resize, warning shows for small terminals, no crashes |
| BvB concurrency selection | Recommended works, custom input works, warning shows for >50, high values with Stats Only mode |
| Navigation stack - BvB flow | ESC from each BvB screen returns to previous; multiple ESC presses reach MainMenu |
| Navigation stack - PvP/PvBot | ESC from setup screens returns correctly; ESC during game shows Save dialog |
| Save/Quit dialog | Yes saves game, No doesn't save, both go to MainMenu; ESC cancels |
| BvB abort dialog | Shows only when games running; Cancel returns to session; Abort stops all games |
| Breadcrumb accuracy | Breadcrumb shows correct path at every screen in BvB flow |
| Settings from multiple screens | Press 's' from different screens, ESC returns to correct origin |

### Benchmarks

| Benchmark | Purpose |
|-----------|---------|
| BvB 10-game session (Easy vs Easy) | Baseline performance |
| BvB 10-game session (Hard vs Hard) | CPU-intensive scenario |
| Concurrency sweep (1x, 1.5x, 2x, 3x NumCPU) | Find optimal multiplier |
| Memory profiling during BvB | Verify no leaks after cleanup |
| BvB 100-game session stats-only mode | Verify high concurrency stability |
| Stats-only vs Grid view terminal I/O | Compare rendering overhead |
| BvB 200-game session with custom concurrency | Test extreme concurrency with Stats Only mode |
