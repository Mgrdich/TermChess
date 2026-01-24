# Functional Specification: Bot Evaluation Improvement

- **Roadmap Item:** Bot Evaluation Improvement - Endgame awareness and game phase detection
- **Status:** Draft
- **Author:** AI Assistant

---

## 1. Overview and Rationale (The "Why")

### Purpose
Improve the bot evaluation function to produce more decisive games by adding game phase detection, endgame-specific evaluation terms, and passed pawn awareness. Currently, Medium and Hard bots frequently draw against each other because the evaluation lacks the ability to make progress in endgame positions.

### Problem Being Solved
The current evaluation function:
- Uses static piece-square tables regardless of game phase (opening vs endgame)
- Has no passed pawn detection or bonus
- Does not incentivize pawn promotion in endgames
- Cannot drive the enemy king to the corner when ahead in material
- Results in repetitive position cycling and fifty-move rule draws between Medium/Hard bots

### Desired Outcome
Bots at Medium and Hard difficulty play more decisive games with fewer draws. When ahead in material, they actively push for checkmate rather than shuffling pieces. Endgame positions are evaluated with phase-appropriate heuristics.

### Success Metrics
- Significant reduction in draw rate for Medium vs Hard games
- Hard bot consistently wins more often against Medium (demonstrating deeper eval advantage)
- Endgames with material advantage are converted to wins more frequently
- No regression in opening/middlegame play quality

---

## 2. Functional Requirements (The "What")

### 2.1 Game Phase Detection

- **As a** bot engine, **I need to** detect the current game phase, **so that** evaluation weights are appropriate for the position.
  - **Acceptance Criteria:**
    - [ ] Game phase determined by remaining material (opening/middlegame/endgame)
    - [ ] Endgame triggered when total non-pawn material drops below a threshold
    - [ ] Phase is a continuous value (0.0 = endgame, 1.0 = opening) for smooth interpolation

---

### 2.2 Phase-Dependent King Evaluation

- **As a** bot, **I need to** evaluate king position differently in middlegame vs endgame, **so that** the king stays safe early but centralizes late.
  - **Acceptance Criteria:**
    - [ ] Middlegame: king rewarded for staying near castled position (corners/edges)
    - [ ] Endgame: king rewarded for centralization (as currently in kingEndgameTable)
    - [ ] Smooth interpolation between the two tables based on game phase

---

### 2.3 Passed Pawn Detection and Bonus

- **As a** bot, **I need to** recognize passed pawns and value them highly, **so that** I push them toward promotion.
  - **Acceptance Criteria:**
    - [ ] Passed pawn detected: no enemy pawns on same file or adjacent files ahead
    - [ ] Bonus scales with advancement (closer to promotion = higher bonus)
    - [ ] Connected passed pawns (adjacent files) get additional bonus
    - [ ] Bonus is more significant in endgame than middlegame

---

### 2.4 Mop-Up Evaluation (Winning Endgame)

- **As a** bot, **I need to** drive the enemy king to the corner when ahead in material, **so that** I can deliver checkmate.
  - **Acceptance Criteria:**
    - [ ] When significantly ahead in material in endgame phase:
      - [ ] Bonus for enemy king being closer to corners/edges
      - [ ] Bonus for own king being closer to enemy king
    - [ ] Only active when material advantage exceeds a threshold (e.g., 3+ pawns worth)

---

### 2.5 Enhanced Pawn Advancement in Endgame

- **As a** bot, **I need to** value pawn advancement more highly in endgames, **so that** I push pawns toward promotion.
  - **Acceptance Criteria:**
    - [ ] Pawn advancement bonus increases in endgame phase
    - [ ] Pawns on 6th/7th rank get significant bonus in endgame

---

## 3. Scope and Boundaries

### In-Scope

- Game phase detection based on material count
- Phase-interpolated king piece-square tables (middlegame safety vs endgame centralization)
- Passed pawn detection and rank-scaled bonus
- Mop-up evaluation for winning endgames
- Enhanced pawn advancement bonus in endgame
- All changes confined to `internal/bot/eval.go`

### Out-of-Scope

- Opening book or predefined opening moves
- Transposition tables
- Quiescence search
- Endgame tablebases (Syzygy, etc.)
- Changes to search depth or iterative deepening logic
- Changes to alpha-beta pruning or move ordering
- UCI protocol support
