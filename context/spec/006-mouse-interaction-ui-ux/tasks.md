# Tasks: Mouse Interaction & UI/UX Enhancements

## Slice 1: Theme System Foundation
*Establish the theme infrastructure with Classic theme only - app remains runnable with new theme system*

- [x] **Slice 1: Implement Theme System with Classic Theme**
  - [x] Create `internal/ui/theme.go` with `Theme` struct containing all color fields
  - [x] Define `ClassicTheme` variable with WCAG AA compliant colors
  - [x] Implement `GetTheme(name ThemeName) Theme` function with enum type (returns Classic for now)
  - [x] Add `Theme string` field to `DisplayConfig` in `internal/config/config.go` with default `"classic"`
  - [x] Update `configFileToConfig()` and `configToConfigFile()` conversion functions
  - [x] Add `theme Theme` field to UI Model and load theme on initialization
  - [x] Replace hardcoded lipgloss styles in `view.go` with theme-based style getters
  - [x] Verify app runs with new theme system (visual appearance unchanged)

## Slice 2: Additional Themes
*Add Modern and Minimalist themes with settings selection*

- [x] **Slice 2: Add Modern and Minimalist Themes with Settings Selection**
  - [x] Define `ModernTheme` variable with distinct WCAG AA compliant colors
  - [x] Define `MinimalistTheme` variable with distinct WCAG AA compliant colors
  - [x] Update `GetTheme()` to return correct theme by name
  - [x] Add theme selection option to Settings screen
  - [x] Theme selection updates config and persists to file
  - [x] Verify all three themes render correctly and persist across restarts

## Slice 3: Turn Indicator Text Styling
*Add turn-colored text using theme system*

- [x] **Slice 3: Turn-Colored Text Indicators**
  - [x] Add `WhiteTurnText` and `BlackTurnText` colors to Theme struct
  - [x] Update all three themes with appropriate turn colors
  - [x] Modify move input prompt to use turn-based color
  - [x] Modify turn status text to use turn-based color
  - [x] Verify turn indicator is clearly visible in all themes

## Slice 4: Navigation Stack and Breadcrumbs
*Add back navigation and location indicator*

- [x] **Slice 4: Navigation Stack with Breadcrumbs**
  - [x] Add `navStack []Screen` field to Model
  - [x] Implement `pushScreen(screen Screen)` method
  - [x] Implement `popScreen()` method
  - [x] Implement `breadcrumb() string` method
  - [x] Update all screen transitions to use `pushScreen()` instead of direct assignment
  - [x] Handle `Esc` key globally to call `popScreen()` (with appropriate exceptions)
  - [x] Render breadcrumb in UI header area
  - [x] Verify back navigation works from all screens

## Slice 5: Keyboard Shortcuts Help Overlay
*Add ? shortcut to show help modal*

- [x] **Slice 5: Keyboard Shortcuts Overlay**
  - [x] Add `showShortcutsOverlay bool` field to Model
  - [x] Handle `?` key globally to toggle overlay
  - [x] Create `renderShortcutsOverlay()` function with all shortcuts organized by context
  - [x] Render overlay as full-screen modal over current view
  - [x] Dismiss overlay on any key press
  - [x] Verify overlay displays correctly and dismisses properly

## Slice 6: Global Keyboard Shortcuts
*Implement remaining global shortcuts (n, q, s)*

- [x] **Slice 6: Global Keyboard Shortcuts**
  - [x] Handle `n` key globally for new game (navigate to game type selection)
  - [x] Handle `q` key globally for quit (with confirmation if game in progress)
  - [x] Handle `s` key globally for settings
  - [x] Ensure shortcuts are disabled during text input modes
  - [x] Update shortcuts overlay with all implemented shortcuts
  - [x] Verify all shortcuts work from appropriate screens

## Slice 7: Menu Visual Hierarchy
*Reorganize menus for better clarity*

- [x] **Slice 7: Menu Reorganization and Visual Hierarchy**
  - [x] Identify less-common menu options to group
  - [x] Add visual separators between option groups using theme colors
  - [x] Style primary actions more prominently than secondary actions
  - [x] Ensure focus indicators are visible when navigating with keyboard
  - [x] Verify menus are clearer and easier to navigate

## Slice 8: Mouse Selection (No Highlighting Yet)
*Basic click-to-select without visual feedback*

- [x] **Slice 8: Basic Mouse Piece Selection**
  - [x] Add `selectedSquare *engine.Square` field to Model
  - [x] Add `tea.MouseMsg` case to `Update()` function
  - [x] Implement `squareFromMouse(x, y int) *engine.Square` helper with fixed board position
  - [x] Implement `handleMouseEvent(msg tea.MouseMsg)` method
  - [x] Only process mouse in PvP and Player vs Bot modes (not Bot vs Bot)
  - [x] Click on own piece sets `selectedSquare`
  - [x] Click on different own piece changes selection
  - [x] Write unit tests for `squareFromMouse()` with various positions
  - [x] Verify piece selection works (no visual feedback yet, but state changes)

## Slice 9: Mouse Move Execution
*Complete moves by clicking destination*

- [ ] **Slice 9: Mouse Move Execution**
  - [ ] Add `validMoves []engine.Square` field to Model
  - [ ] When piece selected, compute and store valid moves using `engine.Board.LegalMoves()`
  - [ ] Click on valid destination executes the move
  - [ ] Click on invalid destination keeps piece selected
  - [ ] Clear selection after successful move
  - [ ] Verify complete mouse-based moves work in PvP and vs Bot modes

## Slice 10: Selection Blinking Effect
*Add visual feedback for selected piece and valid moves*

- [ ] **Slice 10: Blinking Highlight Effect**
  - [ ] Add `blinkOn bool` field to Model
  - [ ] Create `BlinkTickMsg` message type
  - [ ] Start tick command (500ms interval) when piece is selected
  - [ ] Stop tick when selection is cleared
  - [ ] Toggle `blinkOn` on each tick
  - [ ] Add `SelectedHighlight` and `ValidMoveHighlight` colors to all themes
  - [ ] Update `BoardRenderer.Render()` to accept selection state parameters
  - [ ] Apply blinking highlight style to selected square when `blinkOn` is true
  - [ ] Apply blinking highlight style to valid move squares when `blinkOn` is true
  - [ ] Verify blinking effect displays correctly at ~500ms intervals

## Slice 11: Bot vs Bot Speed Simplification
*Reduce speed options to Normal and Instant*

- [ ] **Slice 11: Simplify Bot vs Bot Speed Options**
  - [ ] Remove `SpeedFast` and `SpeedSlow` constants from `internal/bvb/types.go`
  - [ ] Update `SpeedNormal` to 1 second delay
  - [ ] Keep `SpeedInstant` as 0 delay
  - [ ] Update UI to show only two speed options
  - [ ] Update speed toggle controls (remove `1-4` keys, use simpler toggle)
  - [ ] Verify speed changes work correctly during Bot vs Bot

## Slice 12: Bot vs Bot Concurrency Config
*Add configurable concurrency with auto-detection*

- [ ] **Slice 12: Bot vs Bot Concurrency Control**
  - [ ] Add `BvBConcurrency int` to `GameConfig` in `config.go`
  - [ ] Implement `calculateDefaultConcurrency()` with tiered formula
  - [ ] Modify `SessionManager` constructor to accept concurrency parameter
  - [ ] If concurrency is 0, use auto-detected value
  - [ ] Cap concurrency at `maxConcurrentGames`
  - [ ] Display current concurrency setting when starting multi-game session
  - [ ] Add concurrency setting to Settings screen or Bot vs Bot config screen
  - [ ] Write unit tests for `calculateDefaultConcurrency()` formula
  - [ ] Verify concurrency setting affects actual game parallelism

## Slice 13: Bot vs Bot Engine Cleanup
*Prevent memory leaks by destroying engines after games*

- [ ] **Slice 13: Engine Lifecycle Management**
  - [ ] Add `cleanup()` method to `GameSession` struct
  - [ ] Implement engine destruction logic (check for `io.Closer`, set to nil)
  - [ ] Call `cleanup()` via defer in `GameSession.Run()`
  - [ ] Update `SessionManager.Stop()` to cleanup all sessions
  - [ ] Ensure abort channel properly signals goroutines to exit
  - [ ] Run memory profiling to verify no leaks after session completion
  - [ ] Verify Bot vs Bot sessions cleanup properly when user exits

## Slice 14: Bot vs Bot Game Jump Navigation
*Add ability to jump to specific game number*

- [ ] **Slice 14: Jump to Game Number**
  - [ ] Add `bvbJumpInput string` and `bvbShowJumpPrompt bool` to Model
  - [ ] Handle `g` key in Bot vs Bot to show jump prompt
  - [ ] Implement text input for game number
  - [ ] Parse input and validate (numeric, within range)
  - [ ] Navigate to specified game on valid input
  - [ ] Show error message for invalid input
  - [ ] Display "Game X of Y" indicator prominently
  - [ ] Verify jump navigation works correctly

## Slice 15: Bot vs Bot Basic Live Statistics
*Show score and progress during games*

- [ ] **Slice 15: Basic Live Statistics Panel**
  - [ ] Create `renderBvBStats()` function in `view.go`
  - [ ] Display during `ScreenBvBGamePlay`
  - [ ] Show current score: White Wins / Black Wins / Draws
  - [ ] Show progress: Games Completed / Total
  - [ ] Update stats on each `BvBTickMsg`
  - [ ] Verify basic stats display and update correctly

## Slice 16: Bot vs Bot Detailed Statistics
*Add comprehensive statistics panel*

- [ ] **Slice 16: Comprehensive Live Statistics**
  - [ ] Add average move count per completed game
  - [ ] Add current game duration timer
  - [ ] Add longest game (by moves) tracking
  - [ ] Add shortest game (by moves) tracking
  - [ ] Add move history summary for current game (last 10 moves)
  - [ ] Add captured pieces display for current game
  - [ ] Verify all statistics display and update in real-time

## Slice 17: Bot vs Bot Stats-Only Mode
*Add stats-only view mode for high-concurrency sessions*

- [ ] **Slice 17: Stats-Only View Mode**
  - [ ] Add `ScreenBvBViewModeSelect` constant to Screen type in `model.go`
  - [ ] Add `BvBStatsOnlyView` constant to `BvBViewMode` type in `model.go`
  - [ ] Create `renderBvBViewModeSelect()` function in `view.go` with three options
  - [ ] Add descriptions for each view mode option (Grid, Single, Stats Only)
  - [ ] Include "(Recommended for 50+ games)" hint on Stats Only option
  - [ ] Update `ScreenBvBGridConfig` to navigate to `ScreenBvBViewModeSelect` after game count entry
  - [ ] Handle arrow keys and Enter for view mode selection
  - [ ] Handle Esc to go back to game count input
  - [ ] Set `bvbViewMode` based on selection before starting session
  - [ ] Update `v` key handler to cycle through Grid → Single → Stats Only → Grid during session
  - [ ] Create `renderBvBStatsOnly()` function in `view.go`
  - [ ] Display progress bar showing completed/total games
  - [ ] Display score summary (White wins / Black wins / Draws)
  - [ ] Display average moves per completed game
  - [ ] Display "X games in progress" indicator
  - [ ] Display recent completions log (last 5 game results)
  - [ ] Add `BvBDefaultViewMode string` to `GameConfig` in `config.go`
  - [ ] Update config loading/saving for new field
  - [ ] Verify stats-only mode works correctly with high concurrency (50+ games)
  - [ ] Verify view mode can be toggled during active session

## Slice 18: Bot vs Bot Grid Layout Stability
*Fix board position shifting when games end*

- [ ] **Slice 18: Grid Layout Stability**
  - [ ] Define `bvbCellHeight` constant in `view.go` (board + header + status + result + spacing)
  - [ ] Define `bvbCellWidth` constant based on board width with padding
  - [ ] Create `renderBvBGridCell(gameIndex int) string` function with fixed dimensions
  - [ ] Always reserve space for result text line (empty placeholder when game in progress)
  - [ ] Pad or truncate each cell to exactly `bvbCellHeight` lines
  - [ ] Use `lipgloss.Width()` to ensure consistent cell widths
  - [ ] Update `renderBvBGrid()` to use fixed-dimension cells
  - [ ] Verify boards don't shift when games complete at different times
  - [ ] Verify all boards in a row maintain consistent vertical alignment
  - [ ] Test with various grid configurations (2x2, 3x3, 4x4)

## Slice 19: Final Polish and Accessibility Verification
*Ensure WCAG compliance and keyboard accessibility*

- [ ] **Slice 19: Accessibility and Final Polish**
  - [ ] Verify all three themes meet WCAG AA contrast standards using contrast checker
  - [ ] Verify every interactive element is reachable via keyboard
  - [ ] Verify focus indicators are visible throughout the app
  - [ ] Verify mouse interaction has keyboard equivalents (algebraic notation still works)
  - [ ] Run full manual test suite across all features
  - [ ] Fix any remaining visual or interaction issues
