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

- [ ] Update `internal/bvb/session.go`:
  - [ ] Implement pause handling in `Run()` loop (block on resumeCh when paused)
  - [ ] Implement `Pause()` method
  - [ ] Implement `Resume()` method
  - [ ] Implement `Abort()` method (signals stopCh, closes engines)
  - [ ] Ensure engines are closed on any exit path (finish, abort)
- [ ] Update `internal/bvb/session_test.go`:
  - [ ] Test pause blocks game progress
  - [ ] Test resume continues game after pause
  - [ ] Test abort stops game mid-play, engines closed
  - [ ] Test abort during pause works correctly
- [ ] Run tests: `go test ./internal/bvb/`
- [ ] Verify: Sessions can be controlled externally

**Deliverable:** GameSession fully controllable (pause/resume/abort).

---

#### Task 3: Add Speed Control to GameSession
**Goal:** GameSession respects playback speed and speed changes mid-game.

- [ ] Update `internal/bvb/session.go`:
  - [ ] Add speed-based sleep between moves in `Run()` loop
  - [ ] For Instant: no sleep (compute as fast as possible)
  - [ ] For Fast/Normal/Slow: sleep for configured duration
  - [ ] Speed changes picked up on next iteration (shared pointer)
- [ ] Update `internal/bvb/session_test.go`:
  - [ ] Test Instant speed completes game quickly (< 5s for Easy vs Easy)
  - [ ] Test Normal speed has measurable delays between moves
  - [ ] Test speed change mid-game takes effect
- [ ] Run tests: `go test ./internal/bvb/`
- [ ] Verify: Speed control works correctly

**Deliverable:** Sessions pace themselves according to configured speed.

---

#### Task 4: Create SessionManager for Multi-Game Orchestration
**Goal:** SessionManager creates and manages N parallel game sessions.

- [ ] Create `internal/bvb/manager.go`:
  - [ ] Define `SessionManager` struct (sessions, state, speed, difficulties, names)
  - [ ] Implement `NewSessionManager(whiteDiff, blackDiff, whiteName, blackName, gameCount) *SessionManager`
  - [ ] Implement `Start()` - creates engine instances and launches all sessions as goroutines
  - [ ] Implement `Pause()` - pauses all sessions
  - [ ] Implement `Resume()` - resumes all sessions
  - [ ] Implement `SetSpeed(speed)` - updates speed for all sessions
  - [ ] Implement `Abort()` - stops all sessions, cleans up
  - [ ] Implement `Sessions() []*GameSession` - returns sessions slice
  - [ ] Implement `AllFinished() bool`
  - [ ] Implement `State() SessionState`
- [ ] Create `internal/bvb/manager_test.go`:
  - [ ] Test creating manager with N games
  - [ ] Test Start() launches all sessions
  - [ ] Test all sessions complete (Easy vs Easy, 3 games)
  - [ ] Test Pause/Resume affects all sessions
  - [ ] Test SetSpeed propagates to all sessions
  - [ ] Test Abort stops all sessions and cleans up (no goroutine leaks)
  - [ ] Test AllFinished() returns true only when all done
- [ ] Run tests: `go test ./internal/bvb/`
- [ ] Verify: Manager orchestrates parallel games correctly

**Deliverable:** Multi-game parallel execution working. Core BvB logic complete.

---

#### Task 5: Implement Statistics Collection
**Goal:** Compute aggregate statistics from completed game results.

- [ ] Create `internal/bvb/stats.go`:
  - [ ] Define `AggregateStats` struct (TotalGames, WhiteBotName, BlackBotName, WhiteWins, BlackWins, Draws, WhiteWinPct, BlackWinPct, AvgMoveCount, AvgDuration, ShortestGame, LongestGame, IndividualResults)
  - [ ] Implement `ComputeStats(results []GameResult, whiteName, blackName string) *AggregateStats`
  - [ ] Implement `(m *SessionManager) Stats() *AggregateStats` - collects results from finished sessions
- [ ] Create `internal/bvb/stats_test.go`:
  - [ ] Test with known results: correct win counts, percentages
  - [ ] Test draws counted correctly
  - [ ] Test average move count and duration
  - [ ] Test shortest/longest game identified correctly
  - [ ] Test with single game (stats still work)
  - [ ] Test with all draws
- [ ] Run tests: `go test ./internal/bvb/`
- [ ] Verify: Statistics computed accurately

**Deliverable:** Statistics collection complete. Ready for UI integration.

---

### Phase 2: UI Configuration Screens

#### Task 6: Add "Bot vs Bot" Menu Option and BvB Bot Selection Screen
**Goal:** User can select "Bot vs Bot" from game type menu and choose bot difficulties.

- [ ] Update `internal/ui/model.go`:
  - [ ] Add `GameTypeBvB` to GameType enum
  - [ ] Add `ScreenBvBBotSelect` screen state
  - [ ] Add BvB-related fields to Model (bvbWhiteDiff, bvbBlackDiff, bvbSelectingColor)
- [ ] Update `internal/ui/update.go`:
  - [ ] Add "Bot vs Bot" option to GameTypeSelect screen handler
  - [ ] Handle transition: GameTypeSelect → ScreenBvBBotSelect
- [ ] Create `internal/ui/bvb_screens.go`:
  - [ ] Implement `handleBvBBotSelectKeys()`: navigate difficulties, select White then Black bot
  - [ ] ESC returns to GameTypeSelect
  - [ ] Enter on second selection advances to next screen
- [ ] Update `internal/ui/view.go`:
  - [ ] Add rendering for ScreenBvBBotSelect (show difficulty options, indicate White/Black selection)
  - [ ] Add help text for BvB bot select screen
- [ ] Test: Navigate to Bot vs Bot, select difficulties, ESC goes back
- [ ] Verify: Menu flow works, selections stored in model

**Deliverable:** User can navigate to BvB mode and select bot difficulties.

---

#### Task 7: Add Game Mode Selection Screen (Single/Multi-Game)
**Goal:** User can choose single game or enter number of games for multi-game mode.

- [ ] Update `internal/ui/model.go`:
  - [ ] Add `ScreenBvBGameMode` screen state
  - [ ] Add fields: bvbGameCount, bvbCountInput, bvbGameModeChoice
- [ ] Update `internal/ui/bvb_screens.go`:
  - [ ] Implement `handleBvBGameModeKeys()`:
    - [ ] Navigate between "Single Game" and "Multi-Game" options
    - [ ] If Multi-Game selected, show text input for game count
    - [ ] Validate input (positive integer)
    - [ ] Enter advances to grid config screen
    - [ ] ESC returns to BvB bot select
- [ ] Update `internal/ui/view.go`:
  - [ ] Add rendering for ScreenBvBGameMode
  - [ ] Show game mode options and input field for count
  - [ ] Add help text
- [ ] Test: Select single game, select multi-game with count input, validate error on invalid input
- [ ] Verify: Game mode and count stored correctly

**Deliverable:** User can choose single or multi-game mode with count.

---

#### Task 8: Add Grid Configuration Screen
**Goal:** User can select grid layout (presets or custom) before starting games.

- [ ] Update `internal/ui/model.go`:
  - [ ] Add `ScreenBvBGridConfig` screen state
  - [ ] Add fields: bvbGridRows, bvbGridCols, bvbGridInput, bvbGridPresetIndex
- [ ] Update `internal/ui/bvb_screens.go`:
  - [ ] Implement `handleBvBGridConfigKeys()`:
    - [ ] Show preset options: 1x1, 2x2, 2x3, 2x4
    - [ ] Show "Custom" option with row/col input
    - [ ] Validate max 8 boards total (rows * cols <= 8)
    - [ ] Enter starts the BvB session
    - [ ] ESC returns to game mode screen
  - [ ] On Enter: create SessionManager, call Start(), transition to ScreenBvBGamePlay
- [ ] Update `internal/ui/view.go`:
  - [ ] Add rendering for ScreenBvBGridConfig
  - [ ] Show grid presets and custom input option
  - [ ] Add help text
- [ ] Test: Select grid presets, enter custom dimensions, validate max 8
- [ ] Verify: Grid config stored, session started on confirm

**Deliverable:** Full BvB configuration flow complete. Games start after grid selection.

---

### Phase 3: BvB Gameplay Display

#### Task 9: Implement Single-Board BvB View (1x1 Grid)
**Goal:** User can watch a single bot vs bot game with move history and status.

- [ ] Update `internal/ui/model.go`:
  - [ ] Add `ScreenBvBGamePlay` screen state
  - [ ] Add fields: bvbManager, bvbSpeed, bvbViewMode, bvbSelectedGame
  - [ ] Add `BvBViewMode` type (GridView, SingleView)
- [ ] Create `internal/ui/bvb_view.go`:
  - [ ] Implement `renderBvBGamePlay()` function
  - [ ] For single-board view: render full board, move history, bot names, move count, status
  - [ ] Show "Bot thinking..." or game result as status
  - [ ] Show help text with controls
- [ ] Update `internal/ui/update.go`:
  - [ ] Add `BvBTickMsg` message type
  - [ ] Implement `bvbTickCmd()` function (schedule ticks based on speed)
  - [ ] Handle BvBTickMsg: check AllFinished → go to stats; otherwise re-render
  - [ ] Start ticking when entering ScreenBvBGamePlay
- [ ] Update `internal/ui/bvb_screens.go`:
  - [ ] Implement `handleBvBGamePlayKeys()`:
    - [ ] Space: pause/resume
    - [ ] 1-4: change speed
    - [ ] f: export FEN
    - [ ] ESC: abort and return to menu
- [ ] Test: Start single Easy vs Easy game, watch it play out
- [ ] Verify: Board updates, moves display, game reaches completion

**Deliverable:** First playable BvB experience. Single game watchable end-to-end.

---

#### Task 10: Implement Grid View for Multi-Game Display
**Goal:** User can watch multiple games simultaneously in a grid layout.

- [ ] Update `internal/ui/bvb_view.go`:
  - [ ] Implement `renderBvBGrid()` function:
    - [ ] Render compact boards using lipgloss JoinHorizontal/JoinVertical
    - [ ] Each board shows: position, game number, move count, status
    - [ ] Completed games visually distinguished (dimmed or different border)
  - [ ] Implement compact board renderer (smaller than full-size)
  - [ ] Calculate grid layout based on bvbGridRows/bvbGridCols
- [ ] Update `internal/ui/bvb_view.go`:
  - [ ] Route to grid view or single view based on bvbViewMode
  - [ ] In grid view, show page indicator if games > grid slots
- [ ] Update `internal/ui/bvb_screens.go`:
  - [ ] Add Tab key: toggle between GridView and SingleView
  - [ ] In grid view: ←/→ navigate pages
  - [ ] In single view: ←/→ navigate between games
- [ ] Test: Start 4 games with 2x2 grid, see all 4 boards
- [ ] Test: Start 8 games with 2x2 grid, see pages
- [ ] Verify: Grid renders correctly, page navigation works

**Deliverable:** Multi-game grid display working. Full viewing experience.

---

#### Task 11: Implement Page Navigation and Game Selection
**Goal:** User can navigate pages in grid view and select specific games in single view.

- [ ] Update `internal/ui/model.go`:
  - [ ] Add field: bvbPageIndex
- [ ] Update `internal/ui/bvb_screens.go`:
  - [ ] Grid view: ←/→ changes bvbPageIndex (clamp to valid range)
  - [ ] Single view: ←/→ changes bvbSelectedGame (clamp to valid range)
  - [ ] Show current page/game indicator
- [ ] Update `internal/ui/bvb_view.go`:
  - [ ] Grid view: display correct subset of games based on page index
  - [ ] Single view: display selected game's full details (board, move history, bot names)
  - [ ] Page indicator: "Page 1/3" or "Game 3/10"
- [ ] Test: Navigate between pages, navigate between games in single view
- [ ] Test: Page clamp works (can't go past last page)
- [ ] Verify: Navigation smooth, correct games displayed

**Deliverable:** Full navigation between pages and games working.

---

### Phase 4: Statistics & Polish

#### Task 12: Implement Statistics Screen
**Goal:** After all games finish, user sees comprehensive statistics.

- [ ] Update `internal/ui/model.go`:
  - [ ] Add `ScreenBvBStats` screen state
- [ ] Update `internal/ui/update.go`:
  - [ ] When BvBTickMsg fires and AllFinished(): transition to ScreenBvBStats
- [ ] Update `internal/ui/bvb_view.go`:
  - [ ] Implement `renderBvBStats()` function:
    - [ ] Single game: winner, total moves, duration, final board
    - [ ] Multi-game: wins per bot (with difficulty name), draws, win percentages
    - [ ] Average move count, average duration
    - [ ] Shortest/longest game (with game number)
    - [ ] Individual game results list (scrollable if many)
  - [ ] Show options: "New Session" / "Return to Menu"
- [ ] Update `internal/ui/bvb_screens.go`:
  - [ ] Implement `handleBvBStatsKeys()`:
    - [ ] Navigate between "New Session" and "Return to Menu"
    - [ ] Enter on "New Session": go back to ScreenBvBBotSelect
    - [ ] Enter on "Return to Menu": go to ScreenMainMenu
    - [ ] ESC: go to ScreenMainMenu
- [ ] Test: Run games to completion, verify statistics accuracy
- [ ] Test: Single game shows single-game stats
- [ ] Test: Navigation options work correctly
- [ ] Verify: All statistics display correctly

**Deliverable:** Complete statistics display. Full BvB flow end-to-end.

---

#### Task 13: Add FEN Export During BvB Gameplay
**Goal:** User can export FEN of the currently focused game to clipboard.

- [ ] Update `internal/ui/bvb_screens.go`:
  - [ ] On 'f' key press during ScreenBvBGamePlay:
    - [ ] Get focused game (bvbSelectedGame in single view, first visible in grid view)
    - [ ] Get current board from session via `CurrentBoard()`
    - [ ] Call `board.ToFEN()` to get FEN string
    - [ ] Copy to clipboard using existing clipboard utility
    - [ ] Show status message "FEN copied to clipboard"
- [ ] Test: Press 'f' during game, verify FEN copied
- [ ] Verify: Correct game's FEN is exported

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
