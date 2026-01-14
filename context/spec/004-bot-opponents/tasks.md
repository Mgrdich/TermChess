# Task List: Bot Opponents

**Spec Directory:** `context/spec/004-bot-opponents/`
**Status:** Ready for Implementation
**Strategy:** Vertical slicing - each task produces a runnable, testable increment

---

## Overview

This task list breaks down the Bot Opponents feature into small, incremental vertical slices. After completing each main task, the application should remain in a working state with visible progress toward the complete feature.

**Key Principle:** Avoid horizontal layering (all backend, then all UI). Instead, implement thin end-to-end slices that deliver incremental value.

---

## Task Breakdown

### Phase 1: Foundation & Infrastructure

#### Task 1: Create Bot Engine Interface & Package Structure
**Goal:** Establish the bot package with core interfaces and types. No actual bot logic yet, just the contract.

- [x] Create `internal/bot/` package directory
- [x] Create `internal/bot/engine.go` with core interfaces:
  - [x] Define `Engine` interface (SelectMove, Name, Close)
  - [x] Define `Configurable` interface (Configure)
  - [x] Define `Stateful` interface (SetPositionHistory)
  - [x] Define `Inspectable` interface (Info)
  - [x] Define `Info` struct with metadata fields
  - [x] Define `EngineType` enum (TypeInternal, TypeUCI, TypeRL)
  - [x] Define `Difficulty` enum (Easy, Medium, Hard)
- [x] Create `internal/bot/engine_test.go` with interface contract tests
- [x] Run tests: `go test ./internal/bot/`
- [x] Verify: Tests pass, package compiles

**Deliverable:** Bot package exists with clear interface contracts. Foundation ready for implementations.

---

#### Task 2: Create Factory Pattern with Functional Options
**Goal:** Implement the factory pattern and functional options for flexible engine configuration.

- [x] Create `internal/bot/factory.go`:
  - [x] Define `EngineOption` type (functional option pattern)
  - [x] Define `engineConfig` struct with configuration fields
  - [x] Implement `WithTimeLimit(duration)` option
  - [x] Implement `WithSearchDepth(depth)` option
  - [x] Implement `WithOptions(map)` option
  - [x] Create placeholder factories (return nil for now):
    - [x] `NewRandomEngine(opts ...EngineOption) (Engine, error)`
    - [x] `NewMinimaxEngine(difficulty, opts ...) (Engine, error)`
- [x] Create `internal/bot/factory_test.go`:
  - [x] Test option parsing and validation
  - [x] Test error handling for invalid options
- [x] Run tests: `go test ./internal/bot/`
- [x] Verify: Factory pattern works, options validated correctly

**Deliverable:** Factory infrastructure ready. Can create engines with custom configuration (once implementations exist).

---

### Phase 2: Easy Bot (Random Move Selection)

#### Task 3: Implement Easy Bot with Legal Move Selection
**Goal:** Create a working Easy bot that makes random legal moves. First playable bot!

- [x] Create `internal/bot/random.go`:
  - [x] Define `randomEngine` struct with fields (name, timeLimit, closed)
  - [x] Implement `SelectMove()` - returns random legal move
  - [x] Implement `Name()` - returns "Easy Bot"
  - [x] Implement `Close()` - sets closed flag
  - [x] Implement `Info()` - returns metadata
- [x] Update `NewRandomEngine()` in `factory.go`:
  - [x] Parse options (timeLimit)
  - [x] Create and return `randomEngine` instance
- [x] Create `internal/bot/random_test.go`:
  - [x] Test SelectMove returns legal moves (100 iterations)
  - [x] Test SelectMove returns error when no legal moves
  - [x] Test Close prevents further use
  - [x] Test single forced move returned immediately
- [x] Run tests: `go test ./internal/bot/`
- [x] Verify: Easy bot makes valid random moves

**Deliverable:** Easy bot functional. Can select random legal moves. Not yet integrated with UI.

---

#### Task 4: Add Weighted Move Selection to Easy Bot
**Goal:** Make Easy bot favor captures/checks (70% tactical bias) while remaining beatable by novices.

- [x] Update `internal/bot/random.go`:
  - [x] Add `filterCaptures()` helper function
  - [x] Add `filterChecks()` helper function
  - [x] Update `SelectMove()` to use weighted selection:
    - [x] 70% chance to pick capture if available
    - [x] 50% chance to pick check if available
    - [x] Fallback to random legal move
- [x] Update `internal/bot/random_test.go`:
  - [x] Test capture bias (statistical test over 100 positions)
  - [x] Test check bias (statistical test over 100 positions)
  - [x] Test fallback to random moves works
- [x] Run tests: `go test ./internal/bot/`
- [x] Verify: Easy bot has tactical awareness but makes mistakes

**Deliverable:** Easy bot with personality. Favors captures/checks but remains beatable.

---

### Phase 3: Evaluation Function (Prepares for Minimax)

#### Task 5: Implement Material-Only Evaluation Function
**Goal:** Create basic position evaluation. Foundation for minimax bots.

- [x] Create `internal/bot/eval.go`:
  - [x] Define `pieceValues` map (Pawn=1, Knight=3, Bishop=3.25, Rook=5, Queen=9)
  - [x] Implement `evaluate(board, difficulty) float64` function:
    - [x] Return ±10000 for checkmate
    - [x] Return 0 for stalemate/draws
    - [x] Call `countMaterial()` for ongoing games
  - [x] Implement `countMaterial(board) float64`:
    - [x] Iterate all squares
    - [x] Sum piece values (positive for White, negative for Black)
    - [x] Return total material score
- [x] Create `internal/bot/eval_test.go`:
  - [x] Test checkmate positions return ±10000
  - [x] Test stalemate returns 0
  - [x] Test material counting (e.g., extra queen = +9)
  - [x] Test starting position evaluates to ~0
  - [x] Test symmetry: eval(pos) = -eval(flipped_pos)
- [x] Run tests: `go test ./internal/bot/`
- [x] Verify: Material evaluation works correctly

**Deliverable:** Working evaluation function. Can score positions by material. Ready for minimax.

---

### Phase 4: Minimax Algorithm (Medium/Hard Bots)

#### Task 6: Implement Basic Minimax with Alpha-Beta Pruning (Depth 2)
**Goal:** Create a working minimax engine. Start with shallow depth for quick testing.

- [x] Create `internal/bot/minimax.go`:
  - [x] Define `minimaxEngine` struct (difficulty, maxDepth, timeLimit, evalWeights, closed)
  - [x] Define `evalWeights` struct (material, pieceSquare, mobility, kingSafety)
  - [x] Implement `getDefaultWeights(difficulty)` function
  - [x] Implement `Name()` - returns "Medium Bot" or "Hard Bot"
  - [x] Implement `Close()` - sets closed flag
  - [x] Implement `SelectMove()` - basic version:
    - [x] Check if closed
    - [x] Create timeout context
    - [x] Handle forced moves (only 1 legal move)
    - [x] Call `searchDepth()` at depth 2 (hardcoded for now)
  - [x] Implement `searchDepth()`:
    - [x] Get legal moves
    - [x] Basic move ordering (captures first)
    - [x] Alpha-beta search on each move
    - [x] Return best move and score
  - [x] Implement `alphaBeta()` recursive function:
    - [x] Base case: depth 0 or game over → evaluate
    - [x] Get legal moves
    - [x] Order moves (captures first)
    - [x] Negamax with alpha-beta cutoffs
    - [x] Return best score
  - [x] Implement `orderMoves()` helper (simple MVV-LVA)
- [x] Update `NewMinimaxEngine()` in `factory.go`:
  - [x] Set defaults based on difficulty (Medium: depth 4, Hard: depth 6)
  - [x] Parse options
  - [x] Create and return `minimaxEngine` instance
- [x] Create `internal/bot/minimax_test.go`:
  - [x] Test forced move returns immediately
  - [x] Test finds mate-in-1 (use simple FEN position)
  - [x] Test doesn't make obvious blunder (hanging queen)
- [x] Run tests: `go test ./internal/bot/`
- [x] Verify: Basic minimax works at depth 2

**Deliverable:** Working minimax bot (shallow depth). Can find simple tactics.

---

#### Task 7: Add Iterative Deepening and Timeout Handling
**Goal:** Make minimax respect time limits and always return a valid move.

- [x] Update `SelectMove()` in `internal/bot/minimax.go`:
  - [x] Implement iterative deepening loop (depth 1 to maxDepth)
  - [x] Check context timeout in each iteration
  - [x] Return best move from last completed depth on timeout
  - [x] Fallback to first legal move if no iteration completed
- [x] Update `searchDepth()`:
  - [x] Add periodic context.Done() checks
  - [x] Return early on timeout
- [x] Update `alphaBeta()`:
  - [x] Add periodic context.Done() checks (every ~100 nodes)
  - [x] Return current alpha on timeout
- [x] Update `internal/bot/minimax_test.go`:
  - [x] Test timeout handling (set 100ms timeout, verify returns move)
  - [x] Test iterative deepening completes multiple depths
  - [x] Test returns best move from last completed depth
- [x] Run tests: `go test ./internal/bot/`
- [x] Verify: Minimax respects timeouts, always returns valid move

**Deliverable:** Minimax with robust timeout handling. Ready for deeper searches.

---

#### Task 8: Add Piece-Square Tables and Mobility Evaluation
**Goal:** Improve bot strength with positional evaluation (Medium/Hard bots).

- [x] Update `internal/bot/eval.go`:
  - [x] Define piece-square tables for each piece type:
    - [x] Pawn table (advancement bonus)
    - [x] Knight table (center control bonus)
    - [x] Bishop table (long diagonal bonus)
    - [x] Rook table (open file bonus)
    - [x] King table (safety in opening, activity in endgame)
  - [x] Implement `evaluatePiecePositions(board) float64`:
    - [x] Iterate all pieces
    - [x] Look up piece-square table bonus
    - [x] Sum positional bonuses
  - [x] Implement `evaluateMobility(board) float64`:
    - [x] Count legal moves for active player
    - [x] Return count as float
  - [x] Update `evaluate()` to use difficulty-based evaluation:
    - [x] Material (all difficulties)
    - [x] Piece-square tables (Medium+)
    - [x] Mobility (Medium+)
- [x] Update `internal/bot/eval_test.go`:
  - [x] Test piece-square tables give correct bonuses
  - [x] Test mobility evaluation counts moves correctly
  - [x] Test difficulty-based evaluation (Easy vs Medium vs Hard)
- [x] Run tests: `go test ./internal/bot/`
- [x] Verify: Medium/Hard bots have positional awareness

**Deliverable:** Bots understand piece positioning and mobility. Medium/Hard bots stronger.

---

#### Task 9: Add King Safety Evaluation (Hard Bot Only)
**Goal:** Make Hard bot understand king safety and play more strategically.

- [x] Update `internal/bot/eval.go`:
  - [x] Implement `evaluateKingSafety(board) float64`:
    - [x] Identify king positions for both colors
    - [x] Check pawn shield completeness (pawns in front of king)
    - [x] Penalty for open files near king
    - [x] Penalty for enemy pieces attacking king zone
    - [x] Return safety score
  - [x] Update `evaluate()` to include king safety for Hard difficulty
- [x] Update `internal/bot/eval_test.go`:
  - [x] Test pawn shield detection
  - [x] Test open file penalty
  - [x] Test attacker detection in king zone
  - [x] Test king safety only affects Hard bot evaluation
- [x] Run tests: `go test ./internal/bot/`
- [x] Verify: Hard bot considers king safety in evaluation

**Deliverable:** Hard bot has strategic depth. Understands king safety.

---

#### Task 10: Implement Configure() for Runtime Tuning
**Goal:** Allow runtime configuration of minimax parameters (depth, timeouts, eval weights).

- [x] Update `internal/bot/minimax.go`:
  - [x] Implement `Configure(options map[string]any) error`:
    - [x] Parse "search_depth" option (validate 1-20)
    - [x] Parse "time_limit" option (validate positive duration)
    - [x] Parse eval weight options (material, pieceSquare, mobility, kingSafety)
    - [x] Update engine fields
    - [x] Return error for invalid options
  - [x] Implement `Info()` method:
    - [x] Return metadata (name, author, version, type, difficulty, features)
- [x] Update `internal/bot/minimax_test.go`:
  - [x] Test Configure() updates search depth
  - [x] Test Configure() updates time limit
  - [x] Test Configure() updates eval weights
  - [x] Test Configure() validates input
  - [x] Test Info() returns correct metadata
- [x] Run tests: `go test ./internal/bot/`
- [x] Verify: Minimax engine can be configured at runtime

**Deliverable:** Configurable minimax engine. Can tune parameters for testing/debugging.

---

### Phase 5: UI Integration (Make Bots Playable)

#### Task 11: Add Bot Move Execution to UI Game Loop
**Goal:** Integrate bot engines with Bubbletea UI. First end-to-end bot game!

- [x] Create `internal/ui/messages.go`:
  - [x] Define `thinkingMessages` array with 12 chess-themed messages
  - [x] Implement `getRandomThinkingMessage() string` function
- [x] Update `internal/ui/model.go`:
  - [x] Add `botEngine bot.Engine` field to Model struct
- [x] Update `internal/ui/update.go`:
  - [x] Define message types:
    - [x] `BotMoveMsg` struct with move field
    - [x] `BotMoveErrorMsg` struct with error field
  - [x] Create `makeBotMove() (Model, tea.Cmd)` method:
    - [x] Set thinking message
    - [x] Create bot engine based on difficulty (Easy/Medium/Hard)
    - [x] Store engine in model.botEngine
    - [x] Return async command that calls engine.SelectMove()
    - [x] Return BotMoveMsg or BotMoveErrorMsg
  - [x] Update `Update()` function to handle messages:
    - [x] Handle `BotMoveMsg`: apply move, clear status, check game over
    - [x] Handle `BotMoveErrorMsg`: display error message
  - [x] Update `handleMoveInput()`:
    - [x] After successful user move, check if bot game
    - [x] If bot game and not game over, call `makeBotMove()`
- [x] Add cleanup logic:
  - [x] In quit handler: call `model.botEngine.Close()` if not nil
  - [x] On game over: call `model.botEngine.Close()` if not nil
- [x] Test manually: Start game vs Easy bot, make moves
- [x] Verify: Bot responds with moves, UI doesn't freeze, game completes

**Deliverable:** Bots fully integrated with UI. Can play complete games vs Easy/Medium/Hard bots.

---

#### Task 12: Handle Bot Plays White (Bot Moves First)
**Goal:** Support user playing Black. Bot makes opening move immediately after setup.

- [x] Update `internal/ui/menu.go` (or wherever game setup happens):
  - [x] After color selection and board setup
  - [x] Check if bot game AND bot plays White
  - [x] If so, immediately call `makeBotMove()` before returning to game loop
- [x] Test manually:
  - [x] Start game vs Easy bot, select Black
  - [x] Start game vs Medium bot, select Black
  - [x] Start game vs Hard bot, select Black
- [x] Verify: Bot makes opening move immediately, user can respond

**Deliverable:** User can play as Black. Bot makes first move correctly.

---

### Phase 6: Testing & Polish

#### Task 13: Create Tactical Puzzle Test Suite
**Goal:** Validate bot quality with standard chess puzzles (mate-in-N, tactics).

- [x] Create `internal/bot/tactics_test.go`:
  - [x] Define test helper: `loadFEN(fenString) *engine.Board`
  - [x] Test mate-in-1 positions (5 examples):
    - [x] Back rank mate
    - [x] Queen + Rook mate
    - [x] Two rooks mate
    - [x] Bishop + Knight mate
    - [x] Smothered mate
  - [x] Test mate-in-2 positions (3 examples)
  - [x] Test tactical patterns:
    - [x] Fork detection
    - [x] Pin detection
    - [x] Skewer detection
    - [x] Discovered attack
  - [x] Test blunder avoidance:
    - [x] Don't hang queen
    - [x] Don't hang rook
    - [x] Don't allow back rank mate
  - [x] Run tests on Medium and Hard bots
  - [x] Allow Easy bot to fail these tests (by design)
- [x] Run tests: `go test ./internal/bot/ -v`
- [x] Verify: Medium/Hard bots find most tactics, Easy bot misses them

**Deliverable:** Comprehensive tactical test suite. Validates bot playing strength.

---

#### Task 14: Add Bot vs Bot Automated Testing
**Goal:** Validate difficulty progression. Medium beats Easy, Hard beats Medium.

- [x] Create `internal/bot/difficulty_test.go`:
  - [x] Implement helper: `runBotGame(white, black Engine) GameResult`
    - [x] Create starting position
    - [x] Alternate moves between white and black
    - [x] Detect game over
    - [x] Return winner
  - [x] Test: Medium vs Easy (run 10 games)
    - [x] Assert Medium wins at least 7 games
  - [x] Test: Hard vs Medium (run 10 games)
    - [x] Assert Hard wins at least 6 games
  - [x] Test: Easy vs Easy (run 5 games)
    - [x] Assert games complete without crashes
- [x] Run tests: `go test ./internal/bot/ -v -timeout 5m`
- [x] If tests fail, tune evaluation weights or search depths
- [x] Verify: Difficulty progression is correct

**Deliverable:** Automated validation of bot strength. Difficulty levels properly calibrated.

---

#### Task 15: Add Performance Benchmarks and Time Limit Tests
**Goal:** Ensure bots meet time constraints (Easy: 2s, Medium: 4s, Hard: 8s).

- [ ] Create `internal/bot/performance_test.go`:
  - [ ] Benchmark: `BenchmarkEasyBot` (depth 0, random)
  - [ ] Benchmark: `BenchmarkMediumBot_Depth4` (minimax depth 4)
  - [ ] Benchmark: `BenchmarkHardBot_Depth6` (minimax depth 6)
  - [ ] Benchmark: `BenchmarkEvaluate` (evaluation function speed)
- [ ] Add time limit tests:
  - [ ] Test Easy bot completes within 2 seconds
  - [ ] Test Medium bot completes within 4 seconds
  - [ ] Test Hard bot completes within 8 seconds
  - [ ] Use complex middlegame positions
- [ ] Run benchmarks: `go test -bench=. ./internal/bot/`
- [ ] Run time tests: `go test ./internal/bot/ -v -run TimeLimit`
- [ ] If tests fail, adjust maxDepth or optimize move ordering
- [ ] Verify: All bots meet time constraints

**Deliverable:** Performance validated. Bots are fast enough for good UX.

---

#### Task 16: Add Artificial Delay for Easy Bot (UX Polish)
**Goal:** Make Easy bot feel more natural (not instant). Add 1-2 second delay.

- [ ] Update `internal/bot/random.go`:
  - [ ] In `SelectMove()`, after selecting move:
    - [ ] Generate random delay between 1-2 seconds
    - [ ] Sleep for that duration (respect context timeout)
- [ ] Test manually: Play vs Easy bot
- [ ] Verify: Easy bot pauses briefly before moving (feels more natural)

**Deliverable:** Easy bot has realistic timing. Better UX.

---

#### Task 17: Manual QA - Play Full Games at Each Difficulty
**Goal:** End-to-end validation. Play actual games and verify quality.

- [ ] Play 3 full games vs Easy bot:
  - [ ] Verify beatable by novice-level play
  - [ ] Verify thinking messages display
  - [ ] Verify no crashes or freezes
  - [ ] Verify game ends correctly (checkmate/stalemate)
- [ ] Play 3 full games vs Medium bot:
  - [ ] Verify provides reasonable challenge
  - [ ] Verify finds basic tactics (forks, pins)
  - [ ] Verify thinking messages display
  - [ ] Verify no crashes or freezes
- [ ] Play 3 full games vs Hard bot:
  - [ ] Verify challenging for experienced players
  - [ ] Verify finds complex tactics
  - [ ] Verify strategic depth (king safety, positioning)
  - [ ] Verify no crashes or freezes
- [ ] Test edge cases:
  - [ ] Resign during bot game
  - [ ] Offer draw (bot accepts/declines correctly)
  - [ ] Rematch after game ends
  - [ ] Play as White and Black
  - [ ] Random color selection
- [ ] Document any bugs found
- [ ] Verify: All difficulty levels work correctly, good UX

**Deliverable:** Feature fully tested manually. Ready for release.

---

## Summary

**Total Tasks:** 17 tasks organized in 6 phases
**Estimated Effort:** 7-10 developer days
**Strategy:** Vertical slicing with incremental, runnable deliverables

### Key Milestones:
1. **Phase 1-2 (Tasks 1-4):** Easy bot playable
2. **Phase 3-4 (Tasks 5-10):** Medium/Hard bots functional
3. **Phase 5 (Tasks 11-12):** Fully integrated with UI
4. **Phase 6 (Tasks 13-17):** Tested, polished, production-ready

### Testing Coverage:
- Unit tests for all components
- Tactical puzzle validation
- Bot vs bot automated games
- Performance benchmarks
- Manual QA across all difficulties

After each task, the application remains in a working state. Early tasks deliver immediate value (playable Easy bot), while later tasks add sophistication (Medium/Hard bots, polish).
