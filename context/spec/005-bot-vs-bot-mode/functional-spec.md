# Functional Specification: Bot vs Bot Mode

- **Roadmap Item:** Bot vs Bot Mode - Automated gameplay for testing, entertainment, and analysis
- **Status:** Implemented
- **Author:** AI Assistant

---

## 1. Overview and Rationale (The "Why")

### Purpose
Bot vs Bot Mode allows users to watch automated chess games between two bot opponents. This feature serves multiple purposes: entertainment (watching bots battle), testing (validating bot behavior and strength), and analysis (observing how different difficulty levels perform against each other).

### Problem Being Solved
Currently, users can only play against bots themselves. There is no way to:
- Observe how bots of different strengths compare
- Run automated test games to validate bot quality
- Watch chess games passively for entertainment or learning
- Gather statistics on bot performance across multiple games

### Desired Outcome
Users can configure and launch automated bot vs bot games, watch them play out in real-time with adjustable speed, run multiple games in parallel with customizable grid displays, and view comprehensive statistics about game outcomes.

### Success Metrics
- Users can successfully start and watch bot vs bot games at all difficulty combinations
- Multi-game mode runs reliably without crashes or freezes
- Statistics accurately reflect game outcomes
- Playback speed controls work smoothly
- Grid display renders correctly at all supported configurations

---

## 2. Functional Requirements (The "What")

### 2.1 Menu Entry Point

- **As a** user, **I want to** see "Bot vs Bot" as a separate option in the game type selection menu, **so that** I can easily access this mode.
  - **Acceptance Criteria:**
    - [x] "Bot vs Bot" appears as a menu option alongside "Player vs Player" and "Player vs Bot"
    - [x] Selecting "Bot vs Bot" navigates to the bot selection screen

---

### 2.2 Bot Selection

- **As a** user, **I want to** select which bot difficulty plays as White and which plays as Black, **so that** I can configure the matchup I want to watch.
  - **Acceptance Criteria:**
    - [x] User selects bot difficulty for White (Easy, Medium, Hard)
    - [x] User selects bot difficulty for Black (Easy, Medium, Hard)
    - [x] Same difficulty can be selected for both sides (e.g., Medium vs Medium)
    - [x] Selection screen clearly indicates which selection is for White and which is for Black
    - [x] User can navigate back (ESC) to game type selection
    - [x] Two-step selection: first White, then Black (ESC from Black goes back to White)

---

### 2.3 Game Mode Selection

- **As a** user, **I want to** choose between a single game or multiple games, **so that** I can either watch one game or run a batch for statistics.
  - **Acceptance Criteria:**
    - [x] After bot selection, user chooses "Single Game" or "Multi-Game"
    - [x] "Single Game" starts gameplay directly (skips grid configuration, uses 1x1 grid in single-board view)
    - [x] If "Multi-Game" selected, user enters the number of games (free-form digit input)
    - [x] Input validation: must be a positive integer, rejects letters and zero
    - [x] Backspace removes characters from input
    - [x] User can navigate back (ESC) to bot selection (ESC from input goes back to menu)

---

### 2.4 Playback Speed Control

- **As a** user, **I want to** control the speed at which moves are played, **so that** I can watch at my preferred pace.
  - **Acceptance Criteria:**
    - [x] Four speed options available: Instant, Fast, Normal, Slow
    - [x] Default speed is "Normal"
    - [x] Speed can be changed at any time during gameplay via keys 1-4
    - [x] Speed values:
      - Instant: 0 delay (moves execute immediately, UI polls at 100ms)
      - Fast: 500ms per move
      - Normal: 1500ms per move
      - Slow: 3000ms per move
    - [x] Speed change applies to all running games immediately

---

### 2.5 Display Options (Grid and Single View)

- **As a** user, **I want to** view games in a customizable grid or single-board view, **so that** I can watch multiple games or focus on one.
  - **Acceptance Criteria:**
    - [x] Preset grid options: 1x1, 2x2, 2x3, 2x4
    - [x] Custom grid option: user can input "RxC" format (e.g., "3x2")
    - [x] Maximum grid size: 8 boards total for UI clarity
    - [x] Toggle between grid view and single-board view via Tab
    - [x] In single-board view, user can navigate between games via ←/→ with wrap-around
    - [x] In grid view, manual page navigation via ←/→ (no wrap, clamped to valid range)
    - [x] **Grid View Display (per board - minimal info only):**
      - Current board position (compact, no coordinates, no color styling)
      - Game number header
      - Move count
      - Game status (winner shown if finished)
      - Finished games visually distinguished with dimmed style
    - [x] **Single-Board View Display (selected game only - full detail):**
      - Current board position (full size with coordinates)
      - Game number and total (e.g., "Game 3 of 10")
      - Move count and active color or result
      - Move history (formatted as "1. e2e4 e7e5 2. g1f3...")
      - Bot matchup (e.g., "Easy Bot (White) vs Hard Bot (Black)")
      - Speed indicator
    - [x] Move history and detailed info are ONLY shown in single-board view
    - [x] Terminal size detection: if terminal too small for grid, shows warning with suggestion to switch to single view

---

### 2.6 Multi-Game Parallel Execution

- **As a** user, **I want** all games to run simultaneously in parallel, **so that** the batch completes faster.
  - **Acceptance Criteria:**
    - [x] All games in multi-game mode start and run at the same time via goroutines
    - [x] Grid view shows different games at different stages of completion
    - [x] Games that finish early display their final position and result
    - [x] Completed games are visually distinguished (dimmed foreground color)

---

### 2.7 User Actions During Gameplay

- **As a** user, **I want to** pause, resume, export FEN, change speed, and abort games, **so that** I have full control over the viewing experience.
  - **Acceptance Criteria:**
    - [x] **Pause/Resume (Space):** Pauses ALL running games simultaneously; resume continues all
    - [x] **Export FEN (f):** Copy FEN of the focused game to clipboard (single view: selected game; grid view: first visible game on current page)
    - [x] **Change Speed (1-4):** Adjust playback speed at any time (applies to all games)
    - [x] **Abort (ESC):** Cancel current session and return to menu
    - [x] **Navigate (←/→):** In single-board view, navigate between games (with wrap); in grid view, navigate between pages (no wrap)
    - [x] **View Toggle (Tab):** Switch between grid and single-board view
    - [x] Help text displays available controls (respects ShowHelpText config)
    - [x] Ctrl+C and 'q' properly clean up bvbManager before exiting

---

### 2.8 Statistics Display

- **As a** user, **I want to** see comprehensive statistics after games complete, **so that** I can analyze bot performance.
  - **Acceptance Criteria:**
    - [x] Statistics shown automatically when all games finish (tick-triggered transition)
    - [x] Statistics reference bot difficulty names (e.g., "Easy Bot" not just "White")
    - [x] **Single Game Statistics:**
      - Winner or draw result with end reason (e.g., "checkmate", "stalemate")
      - Total moves
      - Game duration (rounded to milliseconds)
    - [x] **Multi-Game Statistics:**
      - Wins per bot with win percentage
      - Draws count
      - Average move count and average duration
      - Shortest game (game number and moves)
      - Longest game (game number and moves)
      - List of individual game results (winner/draw, end reason, move count)
    - [x] Options: "New Session" (returns to bot select) or "Return to Menu"
    - [x] Up/down navigation between options, Enter to select, ESC to return to menu

---

### 2.9 Bot Evaluation Improvement (Endgame Awareness)

- **As a** bot engine, **I need to** evaluate positions differently based on game phase, **so that** games produce more decisive results instead of excessive draws.
  - **Acceptance Criteria:**
    - [ ] Game phase detected from remaining non-pawn material (0.0 = endgame, 1.0 = opening)
    - [ ] King evaluation interpolates between middlegame (stay safe) and endgame (centralize) tables
    - [ ] Passed pawns detected and rewarded with rank-scaled bonus (higher = closer to promotion)
    - [ ] Passed pawn bonus amplified in endgame phase
    - [ ] Mop-up evaluation active when significantly ahead in material in endgame:
      - [ ] Rewards enemy king being far from center
      - [ ] Rewards own king being close to enemy king
    - [ ] Pawn advancement bonus increases in endgame phase
    - [ ] Passed pawns available for Medium+ difficulty
    - [ ] Mop-up evaluation available for Hard difficulty only
    - [ ] Draw rate between Medium vs Hard significantly reduced
    - [ ] No regression in opening/middlegame play quality

---

## 3. Scope and Boundaries

### In-Scope

- Bot vs Bot menu entry point and navigation flow
- Bot difficulty selection for White and Black (two-step)
- Single game and multi-game modes
- Playback speed control (Instant, Fast, Normal, Slow)
- Grid view with presets (1x1, 2x2, 2x3, 2x4) and custom configuration (max 8 boards)
- Single-board view with game navigation and full detail (move history, etc.)
- Manual page navigation for large game batches
- Parallel game execution for multi-game mode
- Pause/Resume all games
- FEN export during gameplay
- Abort/cancel functionality
- Comprehensive statistics display (single and multi-game)
- Statistics referencing bot difficulty names
- Terminal size awareness with fallback warning
- Per-move timeout (30 seconds) to prevent infinite computation
- Proper cleanup on quit (Ctrl+C, 'q', ESC)
- Bot evaluation improvement: game phase detection, passed pawns, mop-up evaluation
- Random tie-breaking in minimax engine for varied games

### Out-of-Scope

- **Mouse Interaction & UI/UX Enhancements** (Phase 4 roadmap item)
- **CLI Distribution** (Phase 5 roadmap item)
- **Custom RL Agent** (Phase 6 roadmap item)
- **UCI Engine Integration** (Phase 6 roadmap item)
- Saving/loading bot vs bot game sessions
- Exporting statistics to file
- Tournament bracket mode
- Custom time controls per bot
- Commentary or move annotations during playback
- Scrollable individual results list (rendered inline for now)
- Final board rendering on statistics screen
