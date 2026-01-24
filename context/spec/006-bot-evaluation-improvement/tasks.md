# Task List: Bot Evaluation Improvement

**Spec Directory:** `context/spec/006-bot-evaluation-improvement/`
**Status:** Ready for Implementation
**Strategy:** Incremental evaluation improvements — each task adds one evaluation term with tests

---

## Overview

This task list breaks down the bot evaluation improvement into incremental steps. Each task adds one evaluation term, with unit tests to verify correctness. The existing evaluation behavior is preserved for Easy bots (no changes).

---

## Task Breakdown

### Task 1: Add Game Phase Detection

**Goal:** Compute a game phase value (0.0 = endgame, 1.0 = opening) based on remaining non-pawn material.

- [ ] Add `computeGamePhase(board) float64` function to `internal/bot/eval.go`
- [ ] Add `countNonPawnMaterial(board) float64` helper function
- [ ] Define constants: `totalStartingMaterial`, `endgameThreshold`
- [ ] Phase = 1.0 at starting position, 0.0 when only kings/pawns remain
- [ ] Add unit tests in `internal/bot/eval_test.go`:
  - [ ] Starting position returns ~1.0
  - [ ] Bare kings return 0.0
  - [ ] Kings + pawns only return 0.0
  - [ ] One minor piece above threshold returns small positive phase
  - [ ] Half material returns ~0.5
- [ ] Run tests: `go test ./internal/bot/ -run TestComputeGamePhase`
- [ ] Verify: Phase detection works correctly

**Deliverable:** Game phase detection function ready for use by other evaluation terms.

---

### Task 2: Phase-Interpolated King Piece-Square Tables

**Goal:** King evaluation uses middlegame table (safety) vs endgame table (centralization) based on game phase.

- [ ] Add `kingMiddlegameTable` to `internal/bot/eval.go` (rewards castled positions, penalizes exposed king)
- [ ] Update `evaluatePiecePositions` signature to accept `phase float64` parameter
- [ ] Interpolate king bonus: `phase*mgBonus + (1.0-phase)*egBonus`
- [ ] Update `evaluate()` to pass phase to `evaluatePiecePositions`
- [ ] Add unit tests:
  - [ ] King in center scores higher in endgame (phase=0) than middlegame (phase=1)
  - [ ] King on g1 (castled) scores higher in middlegame than endgame
  - [ ] Interpolation produces intermediate values at phase=0.5
- [ ] Run tests: `go test ./internal/bot/ -run "TestKing|TestEvaluate"`
- [ ] Verify: Existing tests still pass with phase-aware king eval

**Deliverable:** King evaluation is phase-appropriate — safe early, centralized late.

---

### Task 3: Passed Pawn Detection and Bonus

**Goal:** Detect passed pawns and assign rank-scaled bonuses that increase in the endgame.

- [ ] Add `passedPawnBonus` table (rank-indexed bonuses, higher for advanced pawns)
- [ ] Add `isPassedPawn(board, sq, color) bool` helper function
  - [ ] Check same file and adjacent files for enemy pawns ahead
- [ ] Add `evaluatePassedPawns(board, phase) float64` function
  - [ ] For each pawn, check if passed
  - [ ] Apply rank-based bonus scaled by `(1.0 + (1.0 - phase))` (doubles in endgame)
  - [ ] Score from White's perspective
- [ ] Wire into `evaluate()` for Medium+ difficulty
- [ ] Add unit tests:
  - [ ] Isolated passed pawn on e5 detected correctly
  - [ ] Pawn blocked by enemy pawn on same file is NOT passed
  - [ ] Pawn blocked by enemy pawn on adjacent file is NOT passed
  - [ ] Advanced passed pawn (rank 6-7) gets higher bonus than rank 3-4
  - [ ] Endgame phase amplifies passed pawn bonus
  - [ ] Both White and Black passed pawns scored correctly
- [ ] Run tests: `go test ./internal/bot/ -run TestPassedPawn`
- [ ] Verify: Passed pawn evaluation correct

**Deliverable:** Bots recognize and push passed pawns, especially in endgames.

---

### Task 4: Mop-Up Evaluation (Winning Endgame)

**Goal:** When ahead in material in the endgame, reward driving the enemy king to the corner.

- [ ] Add `centerDistance(sq) float64` helper (manhattan distance from center, 0-6 range)
- [ ] Add `evaluateMopUp(board, phase, materialBalance) float64` function
  - [ ] Only active when `phase < 0.5` AND `abs(materialBalance) >= 3.0`
  - [ ] Find both kings
  - [ ] Reward enemy king far from center (higher center distance)
  - [ ] Reward own king close to enemy king (king proximity)
  - [ ] Scale by `(1.0 - phase)` for endgame strength
  - [ ] Return positive for White advantage, negative for Black advantage
- [ ] Wire into `evaluate()` for Hard difficulty only
- [ ] Add unit tests:
  - [ ] Mop-up inactive in middlegame (phase=0.8)
  - [ ] Mop-up inactive when material is even
  - [ ] Mop-up active with +4 material advantage in endgame
  - [ ] Enemy king in corner scores higher than enemy king in center
  - [ ] Own king close to enemy king scores higher than far away
  - [ ] Works for both White and Black advantages (sign flips)
- [ ] Run tests: `go test ./internal/bot/ -run TestMopUp`
- [ ] Verify: Winning side actively pursues checkmate in endgames

**Deliverable:** Hard bot can convert material advantages to checkmate.

---

### Task 5: Integration Testing and Validation

**Goal:** Verify the complete evaluation improvement works end-to-end with reduced draw rates.

- [ ] Run all existing eval tests: `go test ./internal/bot/ -run TestEvaluate`
- [ ] Run all existing minimax tests: `go test ./internal/bot/ -run TestMinimax`
- [ ] Run `go vet ./...`
- [ ] Verify no performance regression (minimax tests complete in similar time)
- [ ] Run a BvB session manually: Medium vs Hard, 5 games at Instant speed
- [ ] Verify: Fewer draws than before, Hard bot wins more consistently
- [ ] Run a BvB session: Easy vs Medium, verify Medium dominates
- [ ] All tests pass: `go test ./internal/bot/ ./internal/bvb/ ./internal/ui/`

**Deliverable:** Evaluation improvement validated. More decisive games.

---

## Summary

**Total Tasks:** 5 tasks
**Strategy:** Incremental additions — each task adds one evaluation term with tests

### Key Milestones:
1. **Task 1:** Game phase detection foundation
2. **Task 2:** King evaluation becomes phase-aware
3. **Task 3:** Passed pawns incentivize pawn promotion
4. **Task 4:** Mop-up helps convert winning endgames to checkmate
5. **Task 5:** End-to-end validation of reduced draw rates
