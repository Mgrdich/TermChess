# Functional Specification: Mouse Interaction & UI/UX Enhancements

- **Roadmap Item:** Phase 4 - Mouse Interaction & UI/UX Enhancements
- **Status:** Complete
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
- [x] When I click on one of my pieces, that piece becomes selected and displays a blinking color effect (medium speed, ~0.5 second cycle)
- [x] When a piece is selected, all valid destination squares for that piece are highlighted with a blinking color effect
- [x] When I click a valid destination square with a piece selected, the move is executed
- [x] When I click an invalid destination square, the piece remains selected (no deselection, no error message)
- [x] When I click on a different one of my own pieces while a piece is selected, the new piece becomes selected instead
- [x] When I click on an empty square or opponent's piece that is not a valid move, the selection is preserved
- [x] Mouse interaction works alongside keyboard input (both methods remain functional)

### 2.2 Board Themes

**As a** user, **I want to** choose from different visual themes, **so that** I can customize the appearance to my preference.

**Acceptance Criteria:**
- [x] Three themes are available: Classic, Modern, and Minimalist
- [x] Each theme defines distinct colors for: light squares, dark squares, selected piece highlight, valid move highlight, and board border
- [x] White pieces use consistent coloring across all squares within a theme
- [x] Black pieces use consistent coloring across all squares within a theme
- [x] Themes can be changed from the settings menu (not during active gameplay)
- [x] Theme selection persists in user configuration and loads on next startup
- [x] All themes meet WCAG AA color contrast standards (4.5:1 for normal text, 3:1 for large text)

### 2.3 Turn Indicator and Text Styling

**As a** user, **I want** the interface to clearly show whose turn it is through text color, **so that** I always know who should move.

**Acceptance Criteria:**
- [x] Move input area and turn status text reflect the current player's color
- [x] When it's White's turn, relevant UI text appears in a light/white color
- [x] When it's Black's turn, relevant UI text appears in a dark/black color (with sufficient contrast against background)
- [x] Turn indicator remains clearly visible in all three themes

### 2.4 Menu and Navigation Improvements

**As a** user, **I want** a cleaner, more intuitive menu system, **so that** I can find options quickly without confusion.

**Acceptance Criteria:**
- [x] Bottom menu is reorganized to reduce visible options (grouping or hiding less-common actions)
- [x] Visual hierarchy clearly distinguishes primary actions from secondary ones
- [x] Breadcrumb or location indicator shows current screen/context (e.g., "Main Menu > Bot vs Bot > Game 3")
- [x] Back navigation is always available and clearly indicated
- [x] Fixed keyboard shortcuts are implemented for common actions:
  - `n` - New game
  - `q` - Quit / Exit
  - `s` - Settings
  - `Esc` - Back / Cancel
  - `Space` - Pause/Resume (in Bot vs Bot mode)
  - `?` - Show keyboard shortcuts help
- [x] Keyboard shortcuts are displayed in a help overlay accessible via `?`

### 2.5 Bot vs Bot Pagination

**As a** user watching a multi-game Bot vs Bot session, **I want to** jump directly to any game number, **so that** I can review specific games without clicking through sequentially.

**Acceptance Criteria:**
- [x] A "Jump to game" option is available during multi-game Bot vs Bot sessions
- [x] User can enter a game number to navigate directly to that game
- [x] Invalid game numbers (out of range, non-numeric) display an appropriate error
- [x] Current game number and total games are always displayed (e.g., "Game 5 of 10")

### 2.6 Bot vs Bot Live Statistics

**As a** user watching Bot vs Bot games, **I want to** see comprehensive statistics while games are still running, **so that** I can track the match progress in real-time.

**Acceptance Criteria:**
- [x] Statistics panel is visible during Bot vs Bot sessions
- [x] Basic counts displayed: current score (wins/losses/draws), games completed, games remaining
- [x] Detailed stats displayed: average move count per game, current game duration, longest game (moves), shortest game (moves)
- [x] Comprehensive stats displayed: move history summary for current game, captured pieces for current game, position evaluation (if available from engine)
- [x] Statistics update in real-time as games progress

### 2.7 Bot vs Bot Speed Options

**As a** user watching Bot vs Bot games, **I want** simplified speed controls, **so that** I can choose between watching moves or seeing instant results.

**Acceptance Criteria:**
- [x] Only two speed options are available: Normal and Instant
- [x] Normal speed: approximately 1 second between moves
- [x] Instant speed: moves execute as fast as possible with no artificial delay
- [x] Speed can be changed during gameplay
- [x] Previous speed options (fast, slow) are removed

### 2.8 Bot vs Bot Concurrency Control

**As a** user running multi-game Bot vs Bot sessions, **I want to** control how many games run concurrently, **so that** I can balance speed against CPU load.

**Acceptance Criteria:**
- [x] Concurrency setting is available for multi-game Bot vs Bot mode
- [x] Default concurrency is auto-detected based on available CPU cores
- [x] User can manually override the concurrency setting
- [x] Setting persists in user configuration
- [x] UI displays current concurrency setting when starting multi-game session
- [x] Engine calculations should be optimized using Go routines where beneficial

### 2.9 Accessibility

**As a** user who needs accessible design, **I want** proper color contrast and full keyboard navigation, **so that** I can use the application comfortably.

**Acceptance Criteria:**
- [x] All three themes meet WCAG AA contrast standards
- [x] Every interactive element can be reached and activated via keyboard
- [x] Focus indicators are visible when navigating with keyboard
- [x] No functionality is mouse-only; keyboard alternatives exist for all actions

### 2.10 Bot vs Bot Stats-Only Mode

**As a** user running many Bot vs Bot games, **I want** an option to hide the game boards and only see statistics, **so that** I can run higher concurrency sessions without terminal lag or rendering issues.

**Acceptance Criteria:**
- [x] After selecting game count in multi-game mode, a "View Mode" selection screen appears
- [x] View mode options are: "Grid View" (default), "Single Board", and "Stats Only"
- [x] Each option has a brief description explaining the mode
- [x] Stats Only description mentions it's recommended for high game counts (50+)
- [x] Selected view mode is used when the session starts
- [x] When Stats Only mode is enabled, no game boards are rendered during the session
- [x] Statistics panel displays all relevant information: current score, games completed/total, average moves per game, current game being played indicators
- [x] Progress bar or similar visualization shows overall session progress
- [x] User can toggle between Stats Only and Grid/Single view modes during a running session using `v` key
- [x] Stats Only mode allows running higher concurrency without terminal performance degradation
- [x] Session completes successfully and shows final statistics regardless of view mode

### 2.11 Bot vs Bot Grid Layout Stability

**As a** user watching multiple Bot vs Bot games in grid view, **I want** the board positions to remain stable when games end, **so that** the display doesn't jump around and I can easily track all games.

**Acceptance Criteria:**
- [x] Each board cell in the grid has a fixed height regardless of game state
- [x] When a game ends and result text is displayed, the board position does not shift
- [x] Result text (e.g., "White wins", "Draw", "Black wins") fits within the allocated cell space
- [x] All boards in a row maintain consistent vertical alignment
- [x] The grid layout remains stable throughout the entire session
- [x] No visual jumping or flickering when games complete at different times

### 2.12 Bot vs Bot Statistics Export

**As a** user who has completed a Bot vs Bot session, **I want** to save the statistics and game data to a file, **so that** I can review the results later, share them, or analyze the games offline.

**Acceptance Criteria:**
- [x] After a BvB session completes, user is prompted with option to save statistics
- [x] Save option is also available from the BvB stats screen via a key (e.g., `s` for save)
- [x] Statistics file includes session summary: total games, wins/losses/draws, bot difficulties
- [x] Statistics file includes per-game details: game number, result, move count, termination reason
- [x] Statistics file includes move history for each game in standard notation (e.g., PGN or simple algebraic)
- [x] File is saved in a user-accessible location (e.g., `~/.termchess/stats/` or current directory)
- [x] Filename includes timestamp for uniqueness (e.g., `bvb_session_2024-01-15_14-30-00.json`)
- [x] User receives confirmation message with file path after successful save
- [x] File format is human-readable (JSON or plain text with clear formatting)
- [x] Error handling for disk write failures with appropriate error message

### 2.13 Terminal Resize and Responsive Layout

**As a** user with varying terminal sizes, **I want** the UI to adapt to my terminal dimensions, **so that** all content remains visible and usable without horizontal scrolling or truncation.

**Acceptance Criteria:**
- [x] Application detects terminal width and height on startup
- [x] Application responds to terminal resize events (`tea.WindowSizeMsg`)
- [x] Chess board always fits within terminal width (minimum ~20 characters for board)
- [x] Bot vs Bot grid view adjusts columns based on available width
- [x] If terminal is too narrow for grid, automatically switch to single board view
- [x] Menu text and options wrap or truncate gracefully on narrow terminals
- [x] Statistics panel adjusts to available width
- [x] Minimum terminal size warning displayed if terminal is too small (e.g., < 40 columns or < 20 rows)
- [x] No horizontal scrolling required at reasonable terminal sizes (80+ columns)
- [x] Content remains readable after resize without requiring restart

### 2.14 Bot vs Bot Concurrency Selection Screen

**As a** user starting a multi-game Bot vs Bot session, **I want** to choose between recommended concurrency or enter my own custom value, **so that** I can run as many concurrent games as I want even if it may cause lag (especially when using stats-only mode).

**Acceptance Criteria:**
- [x] After selecting game count in multi-game mode, a concurrency selection screen appears (before view mode selection)
- [x] Two options are presented: "Recommended (X concurrent)" and "Custom"
- [x] Recommended option shows the auto-calculated value based on CPU cores
- [x] Selecting Custom allows text input for a custom concurrency value
- [x] Custom input validates: must be positive integer, minimum 1
- [x] Custom input has NO upper limit (user accepts responsibility for lag)
- [x] If custom value exceeds 50, show warning: "High concurrency may cause lag. Consider using Stats Only view mode."
- [x] ESC returns to previous screen (game count input)
- [x] Enter confirms selection and proceeds to view mode selection
- [x] Selected concurrency is used when starting the session
- [x] Help text shows navigation options

### 2.15 Navigation Stack and Linear Back-Navigation

**As a** user navigating through the application, **I want** pressing ESC to always return me to the previous screen in the exact order I navigated, **so that** I can predictably backtrack through the app like mobile navigation.

**Acceptance Criteria:**

**Core Stack Behavior:**
- [x] All screen transitions push the current screen onto a navigation stack before navigating forward
- [x] Pressing ESC pops the stack and returns to the previous screen
- [x] Pressing ESC multiple times in succession navigates back through the entire history until reaching Main Menu
- [x] At Main Menu, ESC has no effect (Main Menu is the navigation root)
- [x] The navigation stack is cleared when entering gameplay (gameplay is a terminal destination)
- [x] The breadcrumb display reflects the current navigation stack path

**Bot vs Bot Multi-Step Flow:**
- [x] The complete BvB multi-game setup flow maintains linear back-navigation:
  ```
  Main Menu → Game Type Select → BvB Bot Select (White) → BvB Bot Select (Black)
  → Game Mode → Game Count Input → Concurrency Select → View Mode Select → Gameplay
  ```
- [x] ESC from any BvB setup screen returns to the immediately previous screen in this flow
- [x] ESC from BvB Bot Select (Black) returns to BvB Bot Select (White)
- [x] ESC from Game Count Input returns to Game Mode selection
- [x] ESC from Concurrency Select returns to Game Count Input
- [x] ESC from View Mode Select returns to Concurrency Select

**Bot vs Bot Single Game Flow:**
- [x] Single game BvB skips Game Count, Concurrency, and View Mode screens:
  ```
  Main Menu → Game Type Select → BvB Bot Select (White) → BvB Bot Select (Black)
  → Game Mode → Gameplay
  ```

**Player vs Bot Flow:**
- [x] The PvBot setup flow maintains linear back-navigation:
  ```
  Main Menu → Game Type Select → Bot Difficulty Select → Color Select → Gameplay
  ```
- [x] ESC from Color Select returns to Bot Difficulty Select
- [x] ESC from Bot Difficulty Select returns to Game Type Select

**Player vs Player Flow:**
- [x] PvP has minimal setup:
  ```
  Main Menu → Game Type Select → Gameplay
  ```

**Settings Flow:**
- [x] Settings follows the navigation stack
- [x] ESC from Settings returns to whatever screen the user navigated from
- [x] Settings can be accessed from multiple screens via 's' shortcut; ESC always returns to the originating screen

**During Active Gameplay:**
- [x] ESC during an active game (PvP, PvBot, or single BvB) shows the Save/Quit confirmation dialog
- [x] If user confirms "Yes" on the save prompt, game is saved and user returns to Main Menu
- [x] If user selects "No", user returns to Main Menu without saving
- [x] ESC on the Save/Quit dialog cancels the dialog and returns to the active game
- [x] ESC during BvB multi-game session shows an "Abort session?" confirmation dialog
- [x] If user confirms abort, session is terminated and user returns to Main Menu
- [x] If user cancels abort, user returns to the running session

**Dialog Behavior:**
- [x] Dialogs (Save Prompt, Draw Offer) overlay the current screen and do not push to the navigation stack
- [x] ESC on any dialog dismisses the dialog without taking action
- [x] After dialog dismissal, user remains on the screen they were on before the dialog appeared

**Deprecations:**
- [x] Remove `ScreenResumePrompt` - saved games are handled via "Resume Game" menu option on Main Menu
- [x] All direct screen assignments in ESC handlers are replaced with `popScreen()` calls

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
- Bot vs Bot concurrency selection screen (recommended vs custom with no upper limit)
- Bot vs Bot stats-only mode for high-concurrency sessions
- Bot vs Bot grid layout stability (fixed cell heights)
- Bot vs Bot statistics export to file
- Terminal resize handling and responsive layout
- WCAG AA color contrast compliance
- Full keyboard navigation as mouse alternative
- Navigation stack for consistent linear back-navigation across all screens
- Confirmation dialogs for aborting active games/sessions

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
- ScreenResumePrompt (deprecated - saved games handled via Main Menu "Resume Game" option)
