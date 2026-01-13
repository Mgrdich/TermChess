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

- [ ] Create `internal/bot/random.go`:
  - [ ] Define `randomEngine` struct with fields (name, timeLimit, closed)
  - [ ] Implement `SelectMove()` - returns random legal move
  - [ ] Implement `Name()` - returns "Easy Bot"
  - [ ] Implement `Close()` - sets closed flag
  - [ ] Implement `Info()` - returns metadata
- [ ] Update `NewRandomEngine()` in `factory.go`:
  - [ ] Parse options (timeLimit)
  - [ ] Create and return `randomEngine` instance
- [ ] Create `internal/bot/random_test.go`:
  - [ ] Test SelectMove returns legal moves (100 iterations)
  - [ ] Test SelectMove returns error when no legal moves
  - [ ] Test Close prevents further use
  - [ ] Test single forced move returned immediately
- [ ] Run tests: `go test ./internal/bot/`
- [ ] Verify: Easy bot makes valid random moves

**Deliverable:** Easy bot functional. Can select random legal moves. Not yet integrated with UI.

---

#### Task 4: Add Weighted Move Selection to Easy Bot
**Goal:** Make Easy bot favor captures/checks (70% tactical bias) while remaining beatable by novices.

- [ ] Update `internal/bot/random.go`:
  - [ ] Add `filterCaptures()` helper function
  - [ ] Add `filterChecks()` helper function
  - [ ] Update `SelectMove()` to use weighted selection:
    - [ ] 70% chance to pick capture if available
    - [ ] 50% chance to pick check if available
    - [ ] Fallback to random legal move
- [ ] Update `internal/bot/random_test.go`:
  - [ ] Test capture bias (statistical test over 100 positions)
  - [ ] Test check bias (statistical test over 100 positions)
  - [ ] Test fallback to random moves works
- [ ] Run tests: `go test ./internal/bot/`
- [ ] Verify: Easy bot has tactical awareness but makes mistakes

**Deliverable:** Easy bot with personality. Favors captures/checks but remains beatable.

---

### Phase 3: Evaluation Function (Prepares for Minimax)

#### Task 5: Implement Material-Only Evaluation Function
**Goal:** Create basic position evaluation. Foundation for minimax bots.

- [ ] Create `internal/bot/eval.go`:
  - [ ] Define `pieceValues` map (Pawn=1, Knight=3, Bishop=3.25, Rook=5, Queen=9)
  - [ ] Implement `evaluate(board, difficulty) float64` function:
    - [ ] Return ±10000 for checkmate
    - [ ] Return 0 for stalemate/draws
    - [ ] Call `countMaterial()` for ongoing games
  - [ ] Implement `countMaterial(board) float64`:
    - [ ] Iterate all squares
    - [ ] Sum piece values (positive for White, negative for Black)
    - [ ] Return total material score
- [ ] Create `internal/bot/eval_test.go`:
  - [ ] Test checkmate positions return ±10000
  - [ ] Test stalemate returns 0
  - [ ] Test material counting (e.g., extra queen = +9)
  - [ ] Test starting position evaluates to ~0
  - [ ] Test symmetry: eval(pos) = -eval(flipped_pos)
- [ ] Run tests: `go test ./internal/bot/`
- [ ] Verify: Material evaluation works correctly

**Deliverable:** Working evaluation function. Can score positions by material. Ready for minimax.

---

### Phase 4: Minimax Algorithm (Medium/Hard Bots)

#### Task 6: Implement Basic Minimax with Alpha-Beta Pruning (Depth 2)
**Goal:** Create a working minimax engine. Start with shallow depth for quick testing.

- [ ] Create `internal/bot/minimax.go`:
  - [ ] Define `minimaxEngine` struct (difficulty, maxDepth, timeLimit, evalWeights, closed)
  - [ ] Define `evalWeights` struct (material, pieceSquare, mobility, kingSafety)
  - [ ] Implement `getDefaultWeights(difficulty)` function
  - [ ] Implement `Name()` - returns "Medium Bot" or "Hard Bot"
  - [ ] Implement `Close()` - sets closed flag
  - [ ] Implement `SelectMove()` - basic version:
    - [ ] Check if closed
    - [ ] Create timeout context
    - [ ] Handle forced moves (only 1 legal move)
    - [ ] Call `searchDepth()` at depth 2 (hardcoded for now)
  - [ ] Implement `searchDepth()`:
    - [ ] Get legal moves
    - [ ] Basic move ordering (captures first)
    - [ ] Alpha-beta search on each move
    - [ ] Return best move and score
  - [ ] Implement `alphaBeta()` recursive function:
    - [ ] Base case: depth 0 or game over → evaluate
    - [ ] Get legal moves
    - [ ] Order moves (captures first)
    - [ ] Negamax with alpha-beta cutoffs
    - [ ] Return best score
  - [ ] Implement `orderMoves()` helper (simple MVV-LVA)
- [ ] Update `NewMinimaxEngine()` in `factory.go`:
  - [ ] Set defaults based on difficulty (Medium: depth 4, Hard: depth 6)
  - [ ] Parse options
  - [ ] Create and return `minimaxEngine` instance
- [ ] Create `internal/bot/minimax_test.go`:
  - [ ] Test forced move returns immediately
  - [ ] Test finds mate-in-1 (use simple FEN position)
  - [ ] Test doesn't make obvious blunder (hanging queen)
- [ ] Run tests: `go test ./internal/bot/`
- [ ] Verify: Basic minimax works at depth 2

**Deliverable:** Working minimax bot (shallow depth). Can find simple tactics.

---

#### Task 7: Add Iterative Deepening and Timeout Handling
**Goal:** Make minimax respect time limits and always return a valid move.

- [ ] Update `SelectMove()` in `internal/bot/minimax.go`:
  - [ ] Implement iterative deepening loop (depth 1 to maxDepth)
  - [ ] Check context timeout in each iteration
  - [ ] Return best move from last completed depth on timeout
  - [ ] Fallback to first legal move if no iteration completed
- [ ] Update `searchDepth()`:
  - [ ] Add periodic context.Done() checks
  - [ ] Return early on timeout
- [ ] Update `alphaBeta()`:
  - [ ] Add periodic context.Done() checks (every ~100 nodes)
  - [ ] Return current alpha on timeout
- [ ] Update `internal/bot/minimax_test.go`:
  - [ ] Test timeout handling (set 100ms timeout, verify returns move)
  - [ ] Test iterative deepening completes multiple depths
  - [ ] Test returns best move from last completed depth
- [ ] Run tests: `go test ./internal/bot/`
- [ ] Verify: Minimax respects timeouts, always returns valid move

**Deliverable:** Minimax with robust timeout handling. Ready for deeper searches.

---

#### Task 8: Add Piece-Square Tables and Mobility Evaluation
**Goal:** Improve bot strength with positional evaluation (Medium/Hard bots).

- [ ] Update `internal/bot/eval.go`:
  - [ ] Define piece-square tables for each piece type:
    - [ ] Pawn table (advancement bonus)
    - [ ] Knight table (center control bonus)
    - [ ] Bishop table (long diagonal bonus)
    - [ ] Rook table (open file bonus)
    - [ ] King table (safety in opening, activity in endgame)
  - [ ] Implement `evaluatePiecePositions(board) float64`:
    - [ ] Iterate all pieces
    - [ ] Look up piece-square table bonus
    - [ ] Sum positional bonuses
  - [ ] Implement `evaluateMobility(board) float64`:
    - [ ] Count legal moves for active player
    - [ ] Return count as float
  - [ ] Update `evaluate()` to use difficulty-based evaluation:
    - [ ] Material (all difficulties)
    - [ ] Piece-square tables (Medium+)
    - [ ] Mobility (Medium+)
- [ ] Update `internal/bot/eval_test.go`:
  - [ ] Test piece-square tables give correct bonuses
  - [ ] Test mobility evaluation counts moves correctly
  - [ ] Test difficulty-based evaluation (Easy vs Medium vs Hard)
- [ ] Run tests: `go test ./internal/bot/`
- [ ] Verify: Medium/Hard bots have positional awareness

**Deliverable:** Bots understand piece positioning and mobility. Medium/Hard bots stronger.

---

#### Task 9: Add King Safety Evaluation (Hard Bot Only)
**Goal:** Make Hard bot understand king safety and play more strategically.

- [ ] Update `internal/bot/eval.go`:
  - [ ] Implement `evaluateKingSafety(board) float64`:
    - [ ] Identify king positions for both colors
    - [ ] Check pawn shield completeness (pawns in front of king)
    - [ ] Penalty for open files near king
    - [ ] Penalty for enemy pieces attacking king zone
    - [ ] Return safety score
  - [ ] Update `evaluate()` to include king safety for Hard difficulty
- [ ] Update `internal/bot/eval_test.go`:
  - [ ] Test pawn shield detection
  - [ ] Test open file penalty
  - [ ] Test attacker detection in king zone
  - [ ] Test king safety only affects Hard bot evaluation
- [ ] Run tests: `go test ./internal/bot/`
- [ ] Verify: Hard bot considers king safety in evaluation

**Deliverable:** Hard bot has strategic depth. Understands king safety.

---

#### Task 10: Implement Configure() for Runtime Tuning
**Goal:** Allow runtime configuration of minimax parameters (depth, timeouts, eval weights).

- [ ] Update `internal/bot/minimax.go`:
  - [ ] Implement `Configure(options map[string]any) error`:
    - [ ] Parse "search_depth" option (validate 1-20)
    - [ ] Parse "time_limit" option (validate positive duration)
    - [ ] Parse eval weight options (material, pieceSquare, mobility, kingSafety)
    - [ ] Update engine fields
    - [ ] Return error for invalid options
  - [ ] Implement `Info()` method:
    - [ ] Return metadata (name, author, version, type, difficulty, features)
- [ ] Update `internal/bot/minimax_test.go`:
  - [ ] Test Configure() updates search depth
  - [ ] Test Configure() updates time limit
  - [ ] Test Configure() updates eval weights
  - [ ] Test Configure() validates input
  - [ ] Test Info() returns correct metadata
- [ ] Run tests: `go test ./internal/bot/`
- [ ] Verify: Minimax engine can be configured at runtime

**Deliverable:** Configurable minimax engine. Can tune parameters for testing/debugging.

---

### Phase 5: UI Integration (Make Bots Playable)

#### Task 11: Add Bot Move Execution to UI Game Loop
**Goal:** Integrate bot engines with Bubbletea UI. First end-to-end bot game!

- [ ] Create `internal/ui/messages.go`:
  - [ ] Define `thinkingMessages` array with 12 chess-themed messages
  - [ ] Implement `getRandomThinkingMessage() string` function
- [ ] Update `internal/ui/model.go`:
  - [ ] Add `botEngine bot.Engine` field to Model struct
- [ ] Update `internal/ui/update.go`:
  - [ ] Define message types:
    - [ ] `BotMoveMsg` struct with move field
    - [ ] `BotMoveErrorMsg` struct with error field
  - [ ] Create `makeBotMove() (Model, tea.Cmd)` method:
    - [ ] Set thinking message
    - [ ] Create bot engine based on difficulty (Easy/Medium/Hard)
    - [ ] Store engine in model.botEngine
    - [ ] Return async command that calls engine.SelectMove()
    - [ ] Return BotMoveMsg or BotMoveErrorMsg
  - [ ] Update `Update()` function to handle messages:
    - [ ] Handle `BotMoveMsg`: apply move, clear status, check game over
    - [ ] Handle `BotMoveErrorMsg`: display error message
  - [ ] Update `handleMoveInput()`:
    - [ ] After successful user move, check if bot game
    - [ ] If bot game and not game over, call `makeBotMove()`
- [ ] Add cleanup logic:
  - [ ] In quit handler: call `model.botEngine.Close()` if not nil
  - [ ] On game over: call `model.botEngine.Close()` if not nil
- [ ] Test manually: Start game vs Easy bot, make moves
- [ ] Verify: Bot responds with moves, UI doesn't freeze, game completes

**Deliverable:** Bots fully integrated with UI. Can play complete games vs Easy/Medium/Hard bots.

---

#### Task 12: Handle Bot Plays White (Bot Moves First)
**Goal:** Support user playing Black. Bot makes opening move immediately after setup.

- [ ] Update `internal/ui/menu.go` (or wherever game setup happens):
  - [ ] After color selection and board setup
  - [ ] Check if bot game AND bot plays White
  - [ ] If so, immediately call `makeBotMove()` before returning to game loop
- [ ] Test manually:
  - [ ] Start game vs Easy bot, select Black
  - [ ] Start game vs Medium bot, select Black
  - [ ] Start game vs Hard bot, select Black
- [ ] Verify: Bot makes opening move immediately, user can respond

**Deliverable:** User can play as Black. Bot makes first move correctly.

---

### Phase 6: Testing & Polish

#### Task 13: Create Tactical Puzzle Test Suite
**Goal:** Validate bot quality with standard chess puzzles (mate-in-N, tactics).

- [ ] Create `internal/bot/tactics_test.go`:
  - [ ] Define test helper: `loadFEN(fenString) *engine.Board`
  - [ ] Test mate-in-1 positions (5 examples):
    - [ ] Back rank mate
    - [ ] Queen + Rook mate
    - [ ] Two rooks mate
    - [ ] Bishop + Knight mate
    - [ ] Smothered mate
  - [ ] Test mate-in-2 positions (3 examples)
  - [ ] Test tactical patterns:
    - [ ] Fork detection
    - [ ] Pin detection
    - [ ] Skewer detection
    - [ ] Discovered attack
  - [ ] Test blunder avoidance:
    - [ ] Don't hang queen
    - [ ] Don't hang rook
    - [ ] Don't allow back rank mate
  - [ ] Run tests on Medium and Hard bots
  - [ ] Allow Easy bot to fail these tests (by design)
- [ ] Run tests: `go test ./internal/bot/ -v`
- [ ] Verify: Medium/Hard bots find most tactics, Easy bot misses them

**Deliverable:** Comprehensive tactical test suite. Validates bot playing strength.

---

#### Task 14: Add Bot vs Bot Automated Testing
**Goal:** Validate difficulty progression. Medium beats Easy, Hard beats Medium.

- [ ] Create `internal/bot/difficulty_test.go`:
  - [ ] Implement helper: `runBotGame(white, black Engine) GameResult`
    - [ ] Create starting position
    - [ ] Alternate moves between white and black
    - [ ] Detect game over
    - [ ] Return winner
  - [ ] Test: Medium vs Easy (run 10 games)
    - [ ] Assert Medium wins at least 7 games
  - [ ] Test: Hard vs Medium (run 10 games)
    - [ ] Assert Hard wins at least 6 games
  - [ ] Test: Easy vs Easy (run 5 games)
    - [ ] Assert games complete without crashes
- [ ] Run tests: `go test ./internal/bot/ -v -timeout 5m`
- [ ] If tests fail, tune evaluation weights or search depths
- [ ] Verify: Difficulty progression is correct

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
