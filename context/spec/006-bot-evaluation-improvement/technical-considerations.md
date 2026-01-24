# Technical Specification: Bot Evaluation Improvement

- **Functional Specification:** `context/spec/006-bot-evaluation-improvement/functional-spec.md`
- **Status:** Draft
- **Author(s):** AI Assistant

---

## 1. High-Level Technical Approach

All changes are confined to `internal/bot/eval.go`. The evaluation function gains game phase awareness through material-based phase detection, and several new evaluation terms that scale with the game phase. No changes to the search algorithm, move generation, or UI.

**Core pattern:** Compute a phase value (0.0 = endgame, 1.0 = opening) from remaining non-pawn material. Use this phase to interpolate between middlegame and endgame evaluation terms. Add new evaluation terms (passed pawns, mop-up) that activate primarily in the endgame.

---

## 2. Implementation Details

### 2.1 Game Phase Detection

```go
// Phase constants based on total non-pawn, non-king material value.
// Starting material (no pawns/kings): 2*Q + 4*R + 4*B + 4*N = 2*9 + 4*5 + 4*3.25 + 4*3 = 63
const totalStartingMaterial = 63.0
const endgameThreshold = 16.0 // Below this = pure endgame

// computeGamePhase returns a value from 0.0 (endgame) to 1.0 (opening).
func computeGamePhase(board *engine.Board) float64 {
    material := countNonPawnMaterial(board)
    if material <= endgameThreshold {
        return 0.0
    }
    if material >= totalStartingMaterial {
        return 1.0
    }
    return (material - endgameThreshold) / (totalStartingMaterial - endgameThreshold)
}

// countNonPawnMaterial counts total piece values excluding pawns and kings.
func countNonPawnMaterial(board *engine.Board) float64
```

---

### 2.2 King Piece-Square Tables (Phase-Interpolated)

Add a new `kingMiddlegameTable` that rewards castled/safe king positions:

```go
var kingMiddlegameTable = [64]float64{
    // Rank 1 - castled king positions are best
    0.2, 0.3, 0.1, 0.0, 0.0, 0.1, 0.3, 0.2,
    // Rank 2 - behind pawns
    0.2, 0.2, 0.0, 0.0, 0.0, 0.0, 0.2, 0.2,
    // Rank 3-8 - penalize exposed king
    -0.1, -0.2, -0.2, -0.3, -0.3, -0.2, -0.2, -0.1,
    -0.2, -0.3, -0.3, -0.4, -0.4, -0.3, -0.3, -0.2,
    -0.3, -0.4, -0.4, -0.5, -0.5, -0.4, -0.4, -0.3,
    -0.3, -0.4, -0.4, -0.5, -0.5, -0.4, -0.4, -0.3,
    -0.3, -0.4, -0.4, -0.5, -0.5, -0.4, -0.4, -0.3,
    -0.3, -0.4, -0.4, -0.5, -0.5, -0.4, -0.4, -0.3,
}
```

In `evaluatePiecePositions`, interpolate:
```go
case engine.King:
    phase := computeGamePhase(board)
    mgBonus := kingMiddlegameTable[squareIndex]
    egBonus := kingEndgameTable[squareIndex]
    bonus = phase*mgBonus + (1.0-phase)*egBonus
```

---

### 2.3 Passed Pawn Evaluation

```go
// Passed pawn bonus by rank (from White's perspective, rank 0-7).
// Higher ranks = closer to promotion = bigger bonus.
var passedPawnBonus = [8]float64{
    0.0,  // rank 1 (impossible for pawns)
    0.1,  // rank 2
    0.2,  // rank 3
    0.35, // rank 4
    0.6,  // rank 5
    1.0,  // rank 6
    1.5,  // rank 7
    0.0,  // rank 8 (promoted)
}

// evaluatePassedPawns detects and scores passed pawns.
func evaluatePassedPawns(board *engine.Board, phase float64) float64 {
    // For each pawn:
    //   1. Check if any enemy pawn is on the same file or adjacent files
    //      at a rank >= this pawn's rank (for White) or <= (for Black)
    //   2. If no enemy pawn blocks: it's a passed pawn
    //   3. Apply rank-based bonus, scaled by (1.0 + (1.0 - phase))
    //      (doubles the bonus in pure endgame vs opening)
}
```

---

### 2.4 Mop-Up Evaluation

When one side has a significant material advantage in the endgame, reward:
- Enemy king being far from center (closer to corner)
- Own king being close to enemy king (for checkmate assistance)

```go
const mopUpMaterialThreshold = 3.0 // ~3 pawns advantage to activate

// evaluateMopUp returns a bonus for the winning side to push enemy king to corner.
func evaluateMopUp(board *engine.Board, phase float64, materialBalance float64) float64 {
    // Only active in endgame (phase < 0.5) with material advantage
    if phase > 0.5 || abs(materialBalance) < mopUpMaterialThreshold {
        return 0.0
    }

    // Determine winning side
    // Compute:
    //   1. Enemy king distance from center (manhattan distance, reward far from center)
    //   2. King proximity bonus (close kings = easier to checkmate)
    // Scale by (1.0 - phase) so it's strongest in pure endgame
}

// centerDistance returns manhattan distance from center (0-6 range).
func centerDistance(sq int) float64 {
    file := sq % 8
    rank := sq / 8
    fileDist := abs(float64(file) - 3.5)
    rankDist := abs(float64(rank) - 3.5)
    return fileDist + rankDist
}
```

---

### 2.5 Updated evaluate() Function

```go
func evaluate(board *engine.Board, difficulty Difficulty) float64 {
    // 1. Terminal states (unchanged)
    // 2. Material count (unchanged)
    material := countMaterial(board)
    score := material

    // 3. Game phase detection (NEW)
    phase := computeGamePhase(board)

    // 4. Piece-square tables with phase-interpolated king (UPDATED - Medium+)
    if difficulty >= Medium {
        score += evaluatePiecePositions(board, phase) // signature change: add phase
    }

    // 5. Mobility (unchanged - Medium+)
    if difficulty >= Medium {
        score += evaluateMobility(board) * 0.1
    }

    // 6. King safety (unchanged - Hard only)
    if difficulty >= Hard {
        score += evaluateKingSafety(board)
    }

    // 7. Passed pawns (NEW - Medium+)
    if difficulty >= Medium {
        score += evaluatePassedPawns(board, phase)
    }

    // 8. Mop-up evaluation (NEW - Hard only)
    if difficulty >= Hard {
        score += evaluateMopUp(board, phase, material)
    }

    return score
}
```

---

### 2.6 Difficulty Level Features

| Feature | Easy | Medium | Hard |
|---------|------|--------|------|
| Material count | Yes | Yes | Yes |
| Piece-square tables | No | Yes (phase-interpolated) | Yes (phase-interpolated) |
| Mobility | No | Yes (10%) | Yes (10%) |
| King safety | No | No | Yes |
| Passed pawns | No | Yes | Yes |
| Mop-up evaluation | No | No | Yes |

---

## 3. Impact and Risk Analysis

### System Dependencies
- **`internal/bot/eval.go`** — All changes here
- **`internal/bot/minimax.go`** — No changes (calls `evaluate()` as before)
- **`internal/bot/`** — No API changes; evaluate signature change is internal

### Potential Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Performance regression from phase computation | Slower move computation | Phase computation is O(64) board scan — negligible vs search |
| Passed pawn detection too slow | Slower evaluation | O(pawns * 3_files) per call — well within budget |
| Over-aggressive mop-up distorts middlegame eval | Worse middlegame play | Only active when phase < 0.5 AND material diff > 3.0 |
| King table interpolation affects existing balance | Medium bot plays differently | Phase interpolation preserves existing endgame table values |

---

## 4. Testing Strategy

### Unit Tests (`internal/bot/eval_test.go`)
- `TestComputeGamePhase`: starting position = 1.0, bare kings = 0.0, intermediate values
- `TestEvaluatePassedPawns`: isolated passed pawn detected, blocked pawn not passed, rank-based bonus correct
- `TestEvaluateMopUp`: activates with material advantage in endgame, inactive in middlegame, inactive when material even
- `TestKingPhaseInterpolation`: king in center scores higher in endgame than middlegame
- `TestCountNonPawnMaterial`: correct material count excluding pawns/kings

### Integration Tests
- Run Easy vs Easy, Medium vs Medium, Hard vs Hard games (10 each at instant speed)
- Verify draw rate for Hard vs Medium is lower than before
- Verify no regression in move computation time (stays within 30s timeout)

### Files Modified

| File | Purpose |
|------|---------|
| `internal/bot/eval.go` | All new evaluation terms and phase detection |
| `internal/bot/eval_test.go` | New unit tests for phase, passed pawns, mop-up, king interpolation |
