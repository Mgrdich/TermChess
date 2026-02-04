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

- [x] **Slice 9: Mouse Move Execution**
  - [x] Add `validMoves []engine.Square` field to Model
  - [x] When piece selected, compute and store valid moves using `engine.Board.LegalMoves()`
  - [x] Click on valid destination executes the move
  - [x] Click on invalid destination keeps piece selected
  - [x] Clear selection after successful move
  - [x] Verify complete mouse-based moves work in PvP and vs Bot modes

## Slice 10: Selection Blinking Effect
*Add visual feedback for selected piece and valid moves*

- [x] **Slice 10: Blinking Highlight Effect**
  - [x] Add `blinkOn bool` field to Model
  - [x] Create `BlinkTickMsg` message type
  - [x] Start tick command (500ms interval) when piece is selected
  - [x] Stop tick when selection is cleared
  - [x] Toggle `blinkOn` on each tick
  - [x] Add `SelectedHighlight` and `ValidMoveHighlight` colors to all themes
  - [x] Update `BoardRenderer.Render()` to accept selection state parameters
  - [x] Apply blinking highlight style to selected square when `blinkOn` is true
  - [x] Apply blinking highlight style to valid move squares when `blinkOn` is true
  - [x] Verify blinking effect displays correctly at ~500ms intervals

## Slice 11: Bot vs Bot Speed Simplification
*Reduce speed options to Normal and Instant*

- [x] **Slice 11: Simplify Bot vs Bot Speed Options**
  - [x] Remove `SpeedFast` and `SpeedSlow` constants from `internal/bvb/types.go`
  - [x] Update `SpeedNormal` to 1 second delay
  - [x] Keep `SpeedInstant` as 0 delay
  - [x] Update UI to show only two speed options
  - [x] Update speed toggle controls (remove `1-4` keys, use simpler toggle)
  - [x] Verify speed changes work correctly during Bot vs Bot

## Slice 12: Bot vs Bot Concurrency Config
*Add configurable concurrency with auto-detection*

- [x] **Slice 12: Bot vs Bot Concurrency Control**
  - [x] Add `BvBConcurrency int` to `GameConfig` in `config.go`
  - [x] Implement `calculateDefaultConcurrency()` with tiered formula
  - [x] Modify `SessionManager` constructor to accept concurrency parameter
  - [x] If concurrency is 0, use auto-detected value
  - [x] Cap concurrency at `maxConcurrentGames`
  - [x] Display current concurrency setting when starting multi-game session
  - [x] Add concurrency setting to Settings screen or Bot vs Bot config screen
  - [x] Write unit tests for `calculateDefaultConcurrency()` formula
  - [x] Verify concurrency setting affects actual game parallelism

## Slice 13: Bot vs Bot Engine Cleanup
*Prevent memory leaks by destroying engines after games*

- [x] **Slice 13: Engine Lifecycle Management**
  - [x] Add `cleanup()` method to `GameSession` struct
  - [x] Implement engine destruction logic (check for `io.Closer`, set to nil)
  - [x] Call `cleanup()` via defer in `GameSession.Run()`
  - [x] Update `SessionManager.Stop()` to cleanup all sessions
  - [x] Ensure abort channel properly signals goroutines to exit
  - [x] Run memory profiling to verify no leaks after session completion
  - [x] Verify Bot vs Bot sessions cleanup properly when user exits

## Slice 14: Bot vs Bot Game Jump Navigation
*Add ability to jump to specific game number*

- [x] **Slice 14: Jump to Game Number**
  - [x] Add `bvbJumpInput string` and `bvbShowJumpPrompt bool` to Model
  - [x] Handle `g` key in Bot vs Bot to show jump prompt
  - [x] Implement text input for game number
  - [x] Parse input and validate (numeric, within range)
  - [x] Navigate to specified game on valid input
  - [x] Show error message for invalid input
  - [x] Display "Game X of Y" indicator prominently
  - [x] Verify jump navigation works correctly

## Slice 15: Bot vs Bot Basic Live Statistics
*Show score and progress during games*

- [x] **Slice 15: Basic Live Statistics Panel**
  - [x] Create `renderBvBStats()` function in `view.go`
  - [x] Display during `ScreenBvBGamePlay`
  - [x] Show current score: White Wins / Black Wins / Draws
  - [x] Show progress: Games Completed / Total
  - [x] Update stats on each `BvBTickMsg`
  - [x] Verify basic stats display and update correctly

## Slice 16: Bot vs Bot Detailed Statistics
*Add comprehensive statistics panel*

- [x] **Slice 16: Comprehensive Live Statistics**
  - [x] Add average move count per completed game
  - [x] Add current game duration timer
  - [x] Add longest game (by moves) tracking
  - [x] Add shortest game (by moves) tracking
  - [x] Add move history summary for current game (last 10 moves)
  - [x] Add captured pieces display for current game
  - [x] Verify all statistics display and update in real-time

## Slice 17: Bot vs Bot Stats-Only Mode
*Add stats-only view mode for high-concurrency sessions*

- [x] **Slice 17: Stats-Only View Mode**
  - [x] Add `ScreenBvBViewModeSelect` constant to Screen type in `model.go`
  - [x] Add `BvBStatsOnlyView` constant to `BvBViewMode` type in `model.go`
  - [x] Create `renderBvBViewModeSelect()` function in `view.go` with three options
  - [x] Add descriptions for each view mode option (Grid, Single, Stats Only)
  - [x] Include "(Recommended for 50+ games)" hint on Stats Only option
  - [x] Update `ScreenBvBGridConfig` to navigate to `ScreenBvBViewModeSelect` after game count entry
  - [x] Handle arrow keys and Enter for view mode selection
  - [x] Handle Esc to go back to game count input
  - [x] Set `bvbViewMode` based on selection before starting session
  - [x] Update `v` key handler to cycle through Grid → Single → Stats Only → Grid during session
  - [x] Create `renderBvBStatsOnly()` function in `view.go`
  - [x] Display progress bar showing completed/total games
  - [x] Display score summary (White wins / Black wins / Draws)
  - [x] Display average moves per completed game
  - [x] Display "X games in progress" indicator
  - [x] Display recent completions log (last 5 game results)
  - [x] Add `BvBDefaultViewMode string` to `GameConfig` in `config.go`
  - [x] Update config loading/saving for new field
  - [x] Verify stats-only mode works correctly with high concurrency (50+ games)
  - [x] Verify view mode can be toggled during active session

## Slice 18: Bot vs Bot Grid Layout Stability
*Fix board position shifting when games end*

- [x] **Slice 18: Grid Layout Stability**
  - [x] Define `bvbCellHeight` constant in `view.go` (board + header + status + result + spacing)
  - [x] Define `bvbCellWidth` constant based on board width with padding
  - [x] Create `renderBvBGridCell(gameIndex int) string` function with fixed dimensions
  - [x] Always reserve space for result text line (empty placeholder when game in progress)
  - [x] Pad or truncate each cell to exactly `bvbCellHeight` lines
  - [x] Use `lipgloss.Width()` to ensure consistent cell widths
  - [x] Update `renderBvBGrid()` to use fixed-dimension cells
  - [x] Verify boards don't shift when games complete at different times
  - [x] Verify all boards in a row maintain consistent vertical alignment
  - [x] Test with various grid configurations (2x2, 3x3, 4x4)

## Slice 19: Bot vs Bot Statistics Export
*Save session statistics and game data to file*

- [x] **Slice 19: Statistics Export**
  - [x] Create `internal/bvb/export.go` with `SessionExport` and `GameExport` structs
  - [x] Add move history tracking to `GameSession` (store moves as they're made)
  - [x] Add termination reason tracking when games end
  - [x] Implement `ExportStats()` method on `SessionManager` to gather all game data
  - [x] Implement `SaveSessionExport()` function to write JSON file
  - [x] Create stats directory (`~/.termchess/stats/`) if not exists
  - [x] Generate filename with timestamp (e.g., `bvb_session_2024-01-15_14-30-00.json`)
  - [x] Handle `s` key on BvB stats screen to trigger save
  - [x] Display success message with filepath after save
  - [x] Display error message if save fails
  - [x] Write unit tests for `ExportStats()` and `SaveSessionExport()`
  - [x] Verify exported JSON contains all session and game data
  - [x] Verify move history is in standard algebraic notation

## Slice 20: Terminal Resize and Responsive Layout
*Ensure UI adapts to terminal size changes*

- [x] **Slice 20: Terminal Resize Handling**
  - [x] Verify `termWidth` and `termHeight` are updated on `tea.WindowSizeMsg`
  - [x] Define constants: `minTerminalWidth` (40), `minTerminalHeight` (20), `bvbCellWidth`
  - [x] Create `renderMinSizeWarning()` function for small terminal warning
  - [x] Update `View()` to check terminal size and show warning if too small
  - [x] Create `adjustBvBGridForWidth()` function to auto-adjust grid columns
  - [x] Call `adjustBvBGridForWidth()` on resize during BvB gameplay
  - [x] Auto-switch to single view if terminal too narrow for grid
  - [x] Ensure menus truncate gracefully on narrow terminals
  - [x] Ensure stats panel adjusts to available width
  - [x] Test resize during active BvB session
  - [x] Test resize during gameplay and menu screens
  - [x] Verify no crashes or rendering issues on resize

## Slice 21: Final Polish and Accessibility Verification
*Ensure WCAG compliance and keyboard accessibility*

- [x] **Slice 21: Accessibility and Final Polish**
  - [x] Verify all three themes meet WCAG AA contrast standards using contrast checker
  - [x] Verify every interactive element is reachable via keyboard
  - [x] Verify focus indicators are visible throughout the app
  - [x] Verify mouse interaction has keyboard equivalents (algebraic notation still works)
  - [x] Run full manual test suite across all features
  - [x] Fix any remaining visual or interaction issues
