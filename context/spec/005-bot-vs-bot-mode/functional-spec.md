# Functional Specification: Bot vs Bot Mode

- **Roadmap Item:** Bot vs Bot Mode - Automated gameplay for testing, entertainment, and analysis
- **Status:** Draft
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
    - [ ] "Bot vs Bot" appears as a menu option alongside "Player vs Player" and "Player vs Bot"
    - [ ] Selecting "Bot vs Bot" navigates to the bot selection screen

---

### 2.2 Bot Selection

- **As a** user, **I want to** select which bot difficulty plays as White and which plays as Black, **so that** I can configure the matchup I want to watch.
  - **Acceptance Criteria:**
    - [ ] User selects bot difficulty for White (Easy, Medium, Hard)
    - [ ] User selects bot difficulty for Black (Easy, Medium, Hard)
    - [ ] Same difficulty can be selected for both sides (e.g., Medium vs Medium)
    - [ ] Selection screen clearly indicates which selection is for White and which is for Black
    - [ ] User can navigate back (ESC) to game type selection

---

### 2.3 Game Mode Selection

- **As a** user, **I want to** choose between a single game or multiple games, **so that** I can either watch one game or run a batch for statistics.
  - **Acceptance Criteria:**
    - [ ] After bot selection, user chooses "Single Game" or "Multi-Game"
    - [ ] If "Multi-Game" selected, user enters the number of games (free-form input)
    - [ ] Input validation: must be a positive integer
    - [ ] User can navigate back (ESC) to bot selection

---

### 2.4 Playback Speed Control

- **As a** user, **I want to** control the speed at which moves are played, **so that** I can watch at my preferred pace.
  - **Acceptance Criteria:**
    - [ ] Four speed options available: Instant, Fast, Normal, Slow
    - [ ] Default speed is "Normal"
    - [ ] Speed can be changed at any time during gameplay
    - [ ] Speed values:
      - Instant: 0 delay (moves execute immediately)
      - Fast: ~0.5 seconds per move
      - Normal: ~1.5 seconds per move
      - Slow: ~3 seconds per move
    - [ ] Speed change applies to all running games immediately

---

### 2.5 Display Options (Grid and Single View)

- **As a** user, **I want to** view games in a customizable grid or single-board view, **so that** I can watch multiple games or focus on one.
  - **Acceptance Criteria:**
    - [ ] Preset grid options: 1x1, 2x2, 2x3, 2x4
    - [ ] Custom grid option: user can input rows and columns
    - [ ] Maximum grid size: 8 boards (2x4) for UI clarity
    - [ ] Toggle between grid view and single-board view
    - [ ] In single-board view, user can navigate to watch any specific game (e.g., "Watching Game 3 of 10")
    - [ ] When more games than grid slots, auto-cycle through pages
    - [ ] **Grid View Display (per board - minimal info only):**
      - Current board position
      - Game number
      - Move count
      - Game status (ongoing/finished)
    - [ ] **Single-Board View Display (selected game only - full detail):**
      - Current board position
      - Game number
      - Move count
      - Move history (formatted as "1. e4 e5 2. Nf3 Nc6...")
      - Bot difficulties (e.g., "Easy Bot (White) vs Hard Bot (Black)")
      - Current game status
      - All detailed information
    - [ ] Move history and detailed info are ONLY shown for the selected game in single-board view (not visible in grid view to avoid clutter)

---

### 2.6 Multi-Game Parallel Execution

- **As a** user, **I want** all games to run simultaneously in parallel, **so that** the batch completes faster.
  - **Acceptance Criteria:**
    - [ ] All games in multi-game mode start and run at the same time
    - [ ] Grid view shows different games at different stages of completion
    - [ ] Games that finish early display their final position and result
    - [ ] Completed games are visually distinguished from ongoing games

---

### 2.7 User Actions During Gameplay

- **As a** user, **I want to** pause, resume, export FEN, change speed, and abort games, **so that** I have full control over the viewing experience.
  - **Acceptance Criteria:**
    - [ ] **Pause/Resume:** Pauses ALL running games simultaneously; resume continues all
    - [ ] **Export FEN:** Copy FEN of the currently selected/focused game to clipboard
    - [ ] **Change Speed:** Adjust playback speed at any time (applies to all games)
    - [ ] **Abort:** Cancel current session and return to menu
    - [ ] **Navigate:** In single-board view, navigate between games (next/previous or jump to specific game number)
    - [ ] Help text displays available controls (respects ShowHelpText config)

---

### 2.8 Statistics Display

- **As a** user, **I want to** see comprehensive statistics after games complete, **so that** I can analyze bot performance.
  - **Acceptance Criteria:**
    - [ ] Statistics shown for both single games and multi-game sessions
    - [ ] Statistics reference bot difficulty (e.g., "Easy Bot" not just "White")
    - [ ] **Single Game Statistics:**
      - Winner (e.g., "Hard Bot (White) wins by checkmate")
      - Total moves
      - Game duration
      - Final position displayed
    - [ ] **Multi-Game Statistics:**
      - Wins per bot difficulty (e.g., "Easy Bot: 3 wins, Hard Bot: 7 wins")
      - Draws count
      - Win percentage for each bot
      - Average game length (moves)
      - Average game duration (time)
      - Shortest game (moves and which game number)
      - Longest game (moves and which game number)
      - List of individual game results (Game 1: Hard Bot won in 34 moves, Game 2: Draw by stalemate, etc.)
    - [ ] Option to start new session or return to main menu

---

## 3. Scope and Boundaries

### In-Scope

- Bot vs Bot menu entry point and navigation flow
- Bot difficulty selection for White and Black
- Single game and multi-game modes
- Playback speed control (Instant, Fast, Normal, Slow)
- Grid view with presets (1x1, 2x2, 2x3, 2x4) and custom configuration (max 8 boards)
- Single-board view with game navigation and full detail (move history, etc.)
- Auto-cycling pagination for large game batches
- Parallel game execution for multi-game mode
- Pause/Resume all games
- FEN export during gameplay
- Abort/cancel functionality
- Comprehensive statistics display (single and multi-game)
- Statistics referencing bot difficulty names

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
