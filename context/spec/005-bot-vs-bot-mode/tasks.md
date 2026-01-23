# Task List: Bot vs Bot Mode

**Spec Directory:** `context/spec/005-bot-vs-bot-mode/`
**Status:** Ready for Implementation
**Strategy:** Vertical slicing - each task produces a runnable, testable increment

---

## Overview

This task list breaks down the Bot vs Bot Mode feature into small, incremental vertical slices. After completing each main task, the application should remain in a working state with visible progress toward the complete feature.

---

## Task Breakdown

### Phase 1: Core BvB Package & Single Game

#### Task 1: Create BvB Types and GameSession (Single Game, No UI)
**Goal:** Establish the `internal/bvb/` package with types and a working single-game session that runs to completion.

- [x] Create `internal/bvb/` package directory
- [x] Create `internal/bvb/types.go`:
  - [x] Define `PlaybackSpeed` enum (Instant, Fast, Normal, Slow) with `Duration()` method
  - [x] Define `SessionState` enum (Running, Paused, Finished)
  - [x] Define `GameResult` struct (GameNumber, Winner, WinnerColor, EndReason, MoveCount, Duration, FinalFEN, MoveHistory)
- [x] Create `internal/bvb/session.go`:
  - [x] Define `GameSession` struct with board, engines, state, mutex, channels
  - [x] Implement `NewGameSession(gameNumber, whiteEngine, blackEngine, whiteName, blackName, speed *PlaybackSpeed) *GameSession`
  - [x] Implement `Run()` goroutine: loop computing moves, applying them, checking game over
  - [x] Implement `CurrentBoard()` thread-safe board snapshot
  - [x] Implement `CurrentMoveHistory()` thread-safe history copy
  - [x] Implement `IsFinished() bool`
  - [x] Implement `Result() *GameResult`
  - [x] Implement max move limit (500 moves → forced draw)
- [x] Create `internal/bvb/session_test.go`:
  - [x] Test single Easy vs Easy game runs to completion
  - [x] Test game result is populated correctly
  - [x] Test max move limit triggers forced draw
  - [x] Test thread-safe accessors work during active game
- [x] Run tests: `go test ./internal/bvb/`
- [x] Verify: Single game session runs in a goroutine and completes

**Deliverable:** BvB package exists with working single-game execution. Foundation for manager.

---

#### Task 2: Add Pause/Resume/Abort to GameSession
**Goal:** GameSession supports pause, resume, and abort via channel signals.

- [x] Update `internal/bvb/session.go`:
  - [x] Implement pause handling in `Run()` loop (block on resumeCh when paused)
  - [x] Implement `Pause()` method
  - [x] Implement `Resume()` method
  - [x] Implement `Abort()` method (signals stopCh, closes engines)
  - [x] Ensure engines are closed on any exit path (finish, abort)
- [x] Update `internal/bvb/session_test.go`:
  - [x] Test pause blocks game progress
  - [x] Test resume continues game after pause
  - [x] Test abort stops game mid-play, engines closed
  - [x] Test abort during pause works correctly
- [x] Run tests: `go test ./internal/bvb/`
- [x] Verify: Sessions can be controlled externally

**Deliverable:** GameSession fully controllable (pause/resume/abort).

---

#### Task 3: Add Speed Control to GameSession
**Goal:** GameSession respects playback speed and speed changes mid-game.

- [x] Update `internal/bvb/session.go`:
  - [x] Add speed-based sleep between moves in `Run()` loop
  - [x] For Instant: no sleep (compute as fast as possible)
  - [x] For Fast/Normal/Slow: sleep for configured duration
  - [x] Speed changes picked up on next iteration (shared pointer)
- [x] Update `internal/bvb/session_test.go`:
  - [x] Test Instant speed completes game quickly (< 5s for Easy vs Easy)
  - [x] Test Normal speed has measurable delays between moves
  - [x] Test speed change mid-game takes effect
- [x] Run tests: `go test ./internal/bvb/`
- [x] Verify: Speed control works correctly

**Deliverable:** Sessions pace themselves according to configured speed.

---

#### Task 4: Create SessionManager for Multi-Game Orchestration
**Goal:** SessionManager creates and manages N parallel game sessions.

- [x] Create `internal/bvb/manager.go`:
  - [x] Define `SessionManager` struct (sessions, state, speed, difficulties, names)
  - [x] Implement `NewSessionManager(whiteDiff, blackDiff, whiteName, blackName, gameCount) *SessionManager`
  - [x] Implement `Start()` - creates engine instances and launches all sessions as goroutines
  - [x] Implement `Pause()` - pauses all sessions
  - [x] Implement `Resume()` - resumes all sessions
  - [x] Implement `SetSpeed(speed)` - updates speed for all sessions
  - [x] Implement `Abort()` - stops all sessions, cleans up
  - [x] Implement `Sessions() []*GameSession` - returns sessions slice
  - [x] Implement `AllFinished() bool`
  - [x] Implement `State() SessionState`
- [x] Create `internal/bvb/manager_test.go`:
  - [x] Test creating manager with N games
  - [x] Test Start() launches all sessions
  - [x] Test all sessions complete (Easy vs Easy, 3 games)
  - [x] Test Pause/Resume affects all sessions
  - [x] Test SetSpeed propagates to all sessions
  - [x] Test Abort stops all sessions and cleans up (no goroutine leaks)
  - [x] Test AllFinished() returns true only when all done
- [x] Run tests: `go test ./internal/bvb/`
- [x] Verify: Manager orchestrates parallel games correctly

**Deliverable:** Multi-game parallel execution working. Core BvB logic complete.

---

#### Task 5: Implement Statistics Collection
**Goal:** Compute aggregate statistics from completed game results.

- [x] Create `internal/bvb/stats.go`:
  - [x] Define `AggregateStats` struct (TotalGames, WhiteBotName, BlackBotName, WhiteWins, BlackWins, Draws, WhiteWinPct, BlackWinPct, AvgMoveCount, AvgDuration, ShortestGame, LongestGame, IndividualResults)
  - [x] Implement `ComputeStats(results []GameResult, whiteName, blackName string) *AggregateStats`
  - [x] Implement `(m *SessionManager) Stats() *AggregateStats` - collects results from finished sessions
- [x] Create `internal/bvb/stats_test.go`:
  - [x] Test with known results: correct win counts, percentages
  - [x] Test draws counted correctly
  - [x] Test average move count and duration
  - [x] Test shortest/longest game identified correctly
  - [x] Test with single game (stats still work)
  - [x] Test with all draws
- [x] Run tests: `go test ./internal/bvb/`
- [x] Verify: Statistics computed accurately

**Deliverable:** Statistics collection complete. Ready for UI integration.

---

### Phase 2: UI Configuration Screens

#### Task 6: Add "Bot vs Bot" Menu Option and BvB Bot Selection Screen
**Goal:** User can select "Bot vs Bot" from game type menu and choose bot difficulties.

- [x] Update `internal/ui/model.go`:
  - [x] Add `GameTypeBvB` to GameType enum
  - [x] Add `ScreenBvBBotSelect` screen state
  - [x] Add BvB-related fields to Model (bvbWhiteDiff, bvbBlackDiff, bvbSelectingWhite)
- [x] Update `internal/ui/update.go`:
  - [x] Add "Bot vs Bot" option to GameTypeSelect screen handler
  - [x] Handle transition: GameTypeSelect → ScreenBvBBotSelect
  - [x] Implement `handleBvBBotSelectKeys()`: navigate difficulties, select White then Black bot
  - [x] ESC returns to GameTypeSelect (from White) or back to White selection (from Black)
  - [x] Enter on second selection advances to next screen
- [x] Update `internal/ui/view.go`:
  - [x] Add rendering for ScreenBvBBotSelect (show difficulty options, indicate White/Black selection)
  - [x] Show previously selected White difficulty when selecting Black
  - [x] Add help text for BvB bot select screen
- [x] Test: Navigate to Bot vs Bot, select difficulties, ESC goes back
- [x] Verify: Menu flow works, selections stored in model

**Deliverable:** User can navigate to BvB mode and select bot difficulties.

---

#### Task 7: Add Game Mode Selection Screen (Single/Multi-Game)
**Goal:** User can choose single game or enter number of games for multi-game mode.

- [x] Update `internal/ui/model.go`:
  - [x] Add `ScreenBvBGameMode` screen state
  - [x] Add fields: bvbGameCount, bvbCountInput, bvbInputtingCount
- [x] Update `internal/ui/update.go`:
  - [x] Implement `handleBvBGameModeKeys()`:
    - [x] Navigate between "Single Game" and "Multi-Game" options
    - [x] If Multi-Game selected, show text input for game count
    - [x] Validate input (positive integer, only digits allowed)
    - [x] Enter advances to next screen
    - [x] ESC returns to BvB bot select (Black selection)
  - [x] Implement `handleBvBCountInput()` for text input mode
  - [x] Implement `parsePositiveInt()` helper
- [x] Update `internal/ui/view.go`:
  - [x] Add rendering for ScreenBvBGameMode
  - [x] Show game mode options and input field for count
  - [x] Show matchup info (bot difficulties)
  - [x] Add help text
- [x] Test: Select single game, select multi-game with count input, validate error on invalid input
- [x] Verify: Game mode and count stored correctly

**Deliverable:** User can choose single or multi-game mode with count.

---

#### Task 8: Add Grid Configuration Screen
**Goal:** User can select grid layout (presets or custom) before starting games.

- [x] Update `internal/ui/model.go`:
  - [x] Add `ScreenBvBGridConfig` and `ScreenBvBGamePlay` screen states
  - [x] Add fields: bvbGridRows, bvbGridCols, bvbCustomGridInput, bvbInputtingGrid
  - [x] Add bvbManager field (*bvb.SessionManager)
- [x] Update `internal/ui/update.go`:
  - [x] Implement `handleBvBGridConfigKeys()`:
    - [x] Show preset options: 1x1, 2x2, 2x3, 2x4
    - [x] Show "Custom" option with row/col input (format: RxC)
    - [x] Validate max 8 boards total (rows * cols <= 8)
    - [x] Enter starts the BvB session
    - [x] ESC returns to game mode screen
  - [x] Implement `startBvBSession()`: create SessionManager, call Start(), transition to ScreenBvBGamePlay
  - [x] Implement minimal `handleBvBGamePlayKeys()` (ESC to abort)
  - [x] Implement `parseGridDimensions()` helper
- [x] Update `internal/ui/view.go`:
  - [x] Add rendering for ScreenBvBGridConfig (presets and custom input)
  - [x] Add minimal rendering for ScreenBvBGamePlay (status display)
  - [x] Add help text
- [x] Test: Select grid presets, enter custom dimensions, validate max 8
- [x] Verify: Grid config stored, session started on confirm

**Deliverable:** Full BvB configuration flow complete. Games start after grid selection.

---

### Phase 3: BvB Gameplay Display

#### Task 9: Implement Single-Board BvB View (1x1 Grid)
**Goal:** User can watch a single bot vs bot game with move history and status.

- [x] Update `internal/ui/model.go`:
  - [x] Add fields: bvbSpeed, bvbSelectedGame, bvbViewMode, bvbPaused
  - [x] Add `BvBViewMode` type (BvBGridView, BvBSingleView)
- [x] Update `internal/ui/view.go`:
  - [x] Implement `renderBvBGamePlay()` with view mode routing
  - [x] Implement `renderBvBSingleView()`: full board, move history, bot names, move count, status, speed indicator
  - [x] Show game result when finished, active color when running
  - [x] Show help text with controls
- [x] Update `internal/ui/update.go`:
  - [x] Add `BvBTickMsg` message type
  - [x] Implement `bvbTickCmd()` function (schedule ticks based on speed)
  - [x] Handle BvBTickMsg: check AllFinished → stop ticking; otherwise re-render
  - [x] Start ticking when entering ScreenBvBGamePlay (from startBvBSession)
  - [x] Implement full `handleBvBGamePlayKeys()`:
    - [x] Space: pause/resume
    - [x] 1-4: change speed (Instant/Fast/Normal/Slow)
    - [x] Tab: toggle grid/single view
    - [x] Left/Right (h/l): navigate between games in single view
    - [x] ESC: abort and return to menu
- [x] Test: Speed change, pause/resume, game navigation, view toggle, tick scheduling, render
- [x] Verify: Board updates, moves display, controls work

**Deliverable:** First playable BvB experience. Single game watchable end-to-end.

---

#### Task 10: Implement Grid View for Multi-Game Display
**Goal:** User can watch multiple games simultaneously in a grid layout.

- [x] Update `internal/ui/view.go`:
  - [x] Implement `renderBvBGridView()` function:
    - [x] Render compact boards using lipgloss JoinHorizontal/JoinVertical
    - [x] Each board shows: position, game number, move count, status
    - [x] Completed games visually distinguished (dimmed style)
  - [x] Implement `renderCompactBoardCell()` compact board renderer
  - [x] Calculate grid layout based on bvbGridRows/bvbGridCols
- [x] Update `internal/ui/view.go`:
  - [x] Route to grid view or single view based on bvbViewMode
  - [x] In grid view, show page indicator if games > grid slots
- [x] Update `internal/ui/update.go`:
  - [x] Tab key: toggle between GridView and SingleView
  - [x] In grid view: ←/→ navigate pages (no wrap)
  - [x] In single view: ←/→ navigate between games (with wrap)
- [x] Test: Grid view renders with multiple games
- [x] Test: Page navigation in grid view
- [x] Verify: Grid renders correctly, page navigation works, page indicator shows/hides

**Deliverable:** Multi-game grid display working. Full viewing experience.

---

#### Task 11: Implement Page Navigation and Game Selection
**Goal:** User can navigate pages in grid view and select specific games in single view.

- [x] Update `internal/ui/model.go`:
  - [x] Add field: bvbPageIndex (implemented in Task 10)
- [x] Update `internal/ui/update.go`:
  - [x] Grid view: ←/→ changes bvbPageIndex (clamped, no wrap) (Task 10)
  - [x] Single view: ←/→ changes bvbSelectedGame (with wrap) (Task 9)
  - [x] Show current page/game indicator
- [x] Update `internal/ui/view.go`:
  - [x] Grid view: display correct subset of games based on page index (Task 10)
  - [x] Single view: display selected game's full details (board, move history, bot names) (Task 9)
  - [x] Page indicator: "Page 1/3" in grid view, "Game X of Y" in single view
- [x] Test: Navigate between pages, navigate between games in single view
- [x] Test: Page clamp works (can't go past last page)
- [x] Verify: Navigation smooth, correct games displayed

**Deliverable:** Full navigation between pages and games working.

---

### Phase 4: Statistics & Polish

#### Task 12: Implement Statistics Screen
**Goal:** After all games finish, user sees comprehensive statistics.

- [x] Update `internal/ui/model.go`:
  - [x] Add `ScreenBvBStats` screen state
  - [x] Add `bvbStatsSelection` field
- [x] Update `internal/ui/update.go`:
  - [x] When BvBTickMsg fires and AllFinished(): transition to ScreenBvBStats
  - [x] Implement `handleBvBStatsKeys()` with up/down/enter/esc
  - [x] Implement `handleBvBStatsSelection()` for New Session / Return to Menu
- [x] Update `internal/ui/view.go`:
  - [x] Implement `renderBvBStats()` function:
    - [x] Single game: winner/draw, total moves, duration
    - [x] Multi-game: wins per bot (with name), draws, win percentages
    - [x] Average move count, average duration
    - [x] Shortest/longest game (with game number)
    - [x] Individual game results list
  - [x] Show options: "New Session" / "Return to Menu"
- [x] Test: Transition to stats when all games finish
- [x] Test: Single game shows single-game stats
- [x] Test: Multi-game shows aggregate stats
- [x] Test: Navigation options work correctly (up/down/enter/esc)
- [x] Verify: All statistics display correctly

**Deliverable:** Complete statistics display. Full BvB flow end-to-end.

---

#### Task 13: Add FEN Export During BvB Gameplay
**Goal:** User can export FEN of the currently focused game to clipboard.

- [x] Update `internal/ui/update.go`:
  - [x] On 'f' key press during ScreenBvBGamePlay:
    - [x] Get focused game (bvbSelectedGame in single view, first visible in grid view)
    - [x] Get current board from session via `CurrentBoard()`
    - [x] Call `board.ToFEN()` to get FEN string
    - [x] Copy to clipboard using existing clipboard utility
    - [x] Show status message "FEN copied to clipboard"
  - [x] Updated help text in both views to show 'f: FEN'
- [x] Test: Press 'f' during game in single view
- [x] Test: Press 'f' during game in grid view
- [x] Test: Press 'f' with no manager (no crash)
- [x] Verify: Correct game's FEN is exported

**Deliverable:** FEN export working during BvB gameplay.

---

#### Task 14: Handle Edge Cases and Error Conditions
**Goal:** Graceful handling of terminal size issues, engine errors, and cleanup.

- [ ] Update `internal/ui/bvb_view.go`:
  - [ ] Check terminal size before rendering grid
  - [ ] If terminal too small for selected grid: show warning message, fallback to single-board view
- [ ] Update `internal/bvb/session.go`:
  - [ ] Handle engine.SelectMove() errors gracefully (log error, end game as error result)
  - [ ] Ensure context timeout per move (prevent infinite engine computation)
- [ ] Update `internal/ui/bvb_screens.go`:
  - [ ] On ESC during ScreenBvBGamePlay: abort manager, clean up all goroutines, return to menu
  - [ ] On Ctrl+C: abort manager before quitting application
- [ ] Add integration tests:
  - [ ] Test abort during active multi-game session (no goroutine leaks)
  - [ ] Test engine error during game (game ends with error result)
  - [ ] Test terminal size fallback
- [ ] Run all tests: `go test ./internal/bvb/ ./internal/ui/`
- [ ] Verify: No panics, no leaks, graceful degradation

**Deliverable:** Robust error handling. Production-ready quality.

---

#### Task 15: Final Integration Testing and Polish
**Goal:** End-to-end validation of the complete BvB feature.

- [ ] Run complete flow: menu → bot select → game mode → grid → watch → stats → menu
- [ ] Test all difficulty combinations (Easy/Easy, Easy/Hard, Medium/Hard, Hard/Hard)
- [ ] Test single game mode with all grid sizes
- [ ] Test multi-game mode (5 games, 10 games) with various grid sizes
- [ ] Test all speed settings (Instant, Fast, Normal, Slow)
- [ ] Test speed change mid-game
- [ ] Test pause/resume during active games
- [ ] Test abort during active games
- [ ] Test FEN export at various game states
- [ ] Test page navigation with more games than grid slots
- [ ] Test single-board view navigation
- [ ] Test view toggle (Tab) between grid and single
- [ ] Verify statistics accuracy across multiple runs
- [ ] Verify help text displays correctly (respects ShowHelpText config)
- [ ] Run `go vet ./...` and fix any issues
- [ ] Run all tests: `go test ./...`
- [ ] Verify: Feature complete, stable, no regressions

**Deliverable:** Bot vs Bot Mode fully implemented and tested.

---

## Summary

**Total Tasks:** 15 tasks organized in 4 phases
**Strategy:** Vertical slicing with incremental, runnable deliverables

### Key Milestones:
1. **Phase 1 (Tasks 1-5):** Core BvB logic package complete (sessions, manager, stats)
2. **Phase 2 (Tasks 6-8):** UI configuration flow complete (menu → bot select → game mode → grid)
3. **Phase 3 (Tasks 9-11):** BvB gameplay display working (single view, grid view, navigation)
4. **Phase 4 (Tasks 12-15):** Statistics, polish, edge cases, final testing
