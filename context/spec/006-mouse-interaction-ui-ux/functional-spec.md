# Functional Specification: Mouse Interaction & UI/UX Enhancements

- **Roadmap Item:** Phase 4 - Mouse Interaction & UI/UX Enhancements
- **Status:** Draft
- **Author:** AI Assistant

---

## 1. Overview and Rationale (The "Why")

**Context:** TermChess currently supports keyboard-only input using algebraic notation. While functional, this creates friction for users who prefer point-and-click interaction and doesn't leverage the visual nature of chess. Additionally, the current UI lacks visual polish, has navigation issues, and the Bot vs Bot mode has performance problems when running multiple games.

**Problem Being Solved:**
- Users comfortable with mouse interaction find algebraic notation cumbersome
- The current single-theme appearance lacks visual appeal and doesn't accommodate different preferences
- Menu navigation is cluttered with too many options and poor visual hierarchy
- Bot vs Bot multi-game mode causes CPU lag due to uncontrolled concurrent engine calculations
- Users cannot easily track statistics during Bot vs Bot sessions
- Navigation lacks shortcuts and clear back-navigation paths

**Desired Outcome:** A polished, intuitive chess experience that supports both mouse and keyboard interaction, offers visual customization through themes, provides accessible color contrast, and delivers smooth Bot vs Bot viewing with comprehensive statistics and performance controls.

**Success Metrics:**
- Users can complete games using mouse-only interaction
- Theme switching works correctly and persists across sessions
- Bot vs Bot multi-game runs without CPU lag when concurrency is properly set
- All themes meet WCAG color contrast standards
- Users can navigate efficiently using keyboard shortcuts

---

## 2. Functional Requirements (The "What")

### 2.1 Mouse Interaction

**As a** user, **I want to** select pieces and make moves by clicking with my mouse, **so that** I can play intuitively without memorizing algebraic notation.

**Acceptance Criteria:**
- [ ] When I click on one of my pieces, that piece becomes selected and displays a blinking color effect (medium speed, ~0.5 second cycle)
- [ ] When a piece is selected, all valid destination squares for that piece are highlighted with a blinking color effect
- [ ] When I click a valid destination square with a piece selected, the move is executed
- [ ] When I click an invalid destination square, the piece remains selected (no deselection, no error message)
- [ ] When I click on a different one of my own pieces while a piece is selected, the new piece becomes selected instead
- [ ] When I click on an empty square or opponent's piece that is not a valid move, the selection is preserved
- [ ] Mouse interaction works alongside keyboard input (both methods remain functional)

### 2.2 Board Themes

**As a** user, **I want to** choose from different visual themes, **so that** I can customize the appearance to my preference.

**Acceptance Criteria:**
- [ ] Three themes are available: Classic, Modern, and Minimalist
- [ ] Each theme defines distinct colors for: light squares, dark squares, selected piece highlight, valid move highlight, and board border
- [ ] White pieces use consistent coloring across all squares within a theme
- [ ] Black pieces use consistent coloring across all squares within a theme
- [ ] Themes can be changed from the settings menu (not during active gameplay)
- [ ] Theme selection persists in user configuration and loads on next startup
- [ ] All themes meet WCAG AA color contrast standards (4.5:1 for normal text, 3:1 for large text)

### 2.3 Turn Indicator and Text Styling

**As a** user, **I want** the interface to clearly show whose turn it is through text color, **so that** I always know who should move.

**Acceptance Criteria:**
- [ ] Move input area and turn status text reflect the current player's color
- [ ] When it's White's turn, relevant UI text appears in a light/white color
- [ ] When it's Black's turn, relevant UI text appears in a dark/black color (with sufficient contrast against background)
- [ ] Turn indicator remains clearly visible in all three themes

### 2.4 Menu and Navigation Improvements

**As a** user, **I want** a cleaner, more intuitive menu system, **so that** I can find options quickly without confusion.

**Acceptance Criteria:**
- [ ] Bottom menu is reorganized to reduce visible options (grouping or hiding less-common actions)
- [ ] Visual hierarchy clearly distinguishes primary actions from secondary ones
- [ ] Breadcrumb or location indicator shows current screen/context (e.g., "Main Menu > Bot vs Bot > Game 3")
- [ ] Back navigation is always available and clearly indicated
- [ ] Fixed keyboard shortcuts are implemented for common actions:
  - `n` - New game
  - `q` - Quit / Exit
  - `s` - Settings
  - `Esc` - Back / Cancel
  - `Space` - Pause/Resume (in Bot vs Bot mode)
  - `?` - Show keyboard shortcuts help
- [ ] Keyboard shortcuts are displayed in a help overlay accessible via `?`

### 2.5 Bot vs Bot Pagination

**As a** user watching a multi-game Bot vs Bot session, **I want to** jump directly to any game number, **so that** I can review specific games without clicking through sequentially.

**Acceptance Criteria:**
- [ ] A "Jump to game" option is available during multi-game Bot vs Bot sessions
- [ ] User can enter a game number to navigate directly to that game
- [ ] Invalid game numbers (out of range, non-numeric) display an appropriate error
- [ ] Current game number and total games are always displayed (e.g., "Game 5 of 10")

### 2.6 Bot vs Bot Live Statistics

**As a** user watching Bot vs Bot games, **I want to** see comprehensive statistics while games are still running, **so that** I can track the match progress in real-time.

**Acceptance Criteria:**
- [ ] Statistics panel is visible during Bot vs Bot sessions
- [ ] Basic counts displayed: current score (wins/losses/draws), games completed, games remaining
- [ ] Detailed stats displayed: average move count per game, current game duration, longest game (moves), shortest game (moves)
- [ ] Comprehensive stats displayed: move history summary for current game, captured pieces for current game, position evaluation (if available from engine)
- [ ] Statistics update in real-time as games progress

### 2.7 Bot vs Bot Speed Options

**As a** user watching Bot vs Bot games, **I want** simplified speed controls, **so that** I can choose between watching moves or seeing instant results.

**Acceptance Criteria:**
- [ ] Only two speed options are available: Normal and Instant
- [ ] Normal speed: approximately 1 second between moves
- [ ] Instant speed: moves execute as fast as possible with no artificial delay
- [ ] Speed can be changed during gameplay
- [ ] Previous speed options (fast, slow) are removed

### 2.8 Bot vs Bot Concurrency Control

**As a** user running multi-game Bot vs Bot sessions, **I want to** control how many games run concurrently, **so that** I can balance speed against CPU load.

**Acceptance Criteria:**
- [ ] Concurrency setting is available for multi-game Bot vs Bot mode
- [ ] Default concurrency is auto-detected based on available CPU cores
- [ ] User can manually override the concurrency setting
- [ ] Setting persists in user configuration
- [ ] UI displays current concurrency setting when starting multi-game session
- [ ] Engine calculations should be optimized using Go routines where beneficial

### 2.9 Accessibility

**As a** user who needs accessible design, **I want** proper color contrast and full keyboard navigation, **so that** I can use the application comfortably.

**Acceptance Criteria:**
- [ ] All three themes meet WCAG AA contrast standards
- [ ] Every interactive element can be reached and activated via keyboard
- [ ] Focus indicators are visible when navigating with keyboard
- [ ] No functionality is mouse-only; keyboard alternatives exist for all actions

### 2.10 Bot vs Bot Stats-Only Mode

**As a** user running many Bot vs Bot games, **I want** an option to hide the game boards and only see statistics, **so that** I can run higher concurrency sessions without terminal lag or rendering issues.

**Acceptance Criteria:**
- [ ] After selecting game count in multi-game mode, a "View Mode" selection screen appears
- [ ] View mode options are: "Grid View" (default), "Single Board", and "Stats Only"
- [ ] Each option has a brief description explaining the mode
- [ ] Stats Only description mentions it's recommended for high game counts (50+)
- [ ] Selected view mode is used when the session starts
- [ ] When Stats Only mode is enabled, no game boards are rendered during the session
- [ ] Statistics panel displays all relevant information: current score, games completed/total, average moves per game, current game being played indicators
- [ ] Progress bar or similar visualization shows overall session progress
- [ ] User can toggle between Stats Only and Grid/Single view modes during a running session using `v` key
- [ ] Stats Only mode allows running higher concurrency without terminal performance degradation
- [ ] Session completes successfully and shows final statistics regardless of view mode

### 2.11 Bot vs Bot Grid Layout Stability

**As a** user watching multiple Bot vs Bot games in grid view, **I want** the board positions to remain stable when games end, **so that** the display doesn't jump around and I can easily track all games.

**Acceptance Criteria:**
- [ ] Each board cell in the grid has a fixed height regardless of game state
- [ ] When a game ends and result text is displayed, the board position does not shift
- [ ] Result text (e.g., "White wins", "Draw", "Black wins") fits within the allocated cell space
- [ ] All boards in a row maintain consistent vertical alignment
- [ ] The grid layout remains stable throughout the entire session
- [ ] No visual jumping or flickering when games complete at different times

### 2.12 Bot vs Bot Statistics Export

**As a** user who has completed a Bot vs Bot session, **I want** to save the statistics and game data to a file, **so that** I can review the results later, share them, or analyze the games offline.

**Acceptance Criteria:**
- [ ] After a BvB session completes, user is prompted with option to save statistics
- [ ] Save option is also available from the BvB stats screen via a key (e.g., `s` for save)
- [ ] Statistics file includes session summary: total games, wins/losses/draws, bot difficulties
- [ ] Statistics file includes per-game details: game number, result, move count, termination reason
- [ ] Statistics file includes move history for each game in standard notation (e.g., PGN or simple algebraic)
- [ ] File is saved in a user-accessible location (e.g., `~/.termchess/stats/` or current directory)
- [ ] Filename includes timestamp for uniqueness (e.g., `bvb_session_2024-01-15_14-30-00.json`)
- [ ] User receives confirmation message with file path after successful save
- [ ] File format is human-readable (JSON or plain text with clear formatting)
- [ ] Error handling for disk write failures with appropriate error message

---

## 3. Scope and Boundaries

### In-Scope

- Mouse click interaction for piece selection and movement
- Blinking color effect for selection and valid move highlighting
- Three board themes (Classic, Modern, Minimalist)
- Turn-colored text indicators
- Menu reorganization and visual hierarchy improvements
- Keyboard shortcuts (fixed, non-customizable)
- Breadcrumb navigation and back-navigation improvements
- Bot vs Bot game number jump navigation
- Bot vs Bot comprehensive live statistics panel
- Bot vs Bot speed simplification (Normal/Instant only)
- Bot vs Bot concurrency control with auto-detection
- Bot vs Bot stats-only mode for high-concurrency sessions
- Bot vs Bot grid layout stability (fixed cell heights)
- Bot vs Bot statistics export to file
- WCAG AA color contrast compliance
- Full keyboard navigation as mouse alternative

### Out-of-Scope

- Phase 5: CLI Distribution (release binaries, curl install script)
- Phase 6: Custom RL Agent (training infrastructure, RL bot integration)
- Phase 6: UCI Engine Integration (external engine support)
- Screen reader support (aria labels, announcements) - deferred to future phase
- Customizable keyboard shortcuts
- Movement animations (pieces sliding)
- Drag-and-drop piece movement
- Right-click context menus
- More than three themes
- Theme switching during active gameplay
