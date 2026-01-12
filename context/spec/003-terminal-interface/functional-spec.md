# Functional Specification: Terminal Interface

- **Roadmap Item:** Terminal Interface
- **Status:** Complete
- **Author:** Claude & User

---

## 1. Overview and Rationale (The "Why")

The Terminal Interface is the primary interaction layer between the user and TermChess. It addresses the need for chess enthusiasts, developers, and CLI power users to play chess without leaving their terminal environment. Current solutions require switching to GUI applications or web browsers, which breaks workflow and adds unnecessary overhead.

This feature provides a clean, distraction-free chess experience that integrates seamlessly into terminal-based workflows. It enables users to see the board state clearly, input moves using standard chess notation, and navigate through game menus—all within their familiar command-line environment.

**Success will be measured by:**
- Users can complete full games without confusion about board state or move input
- Move notation errors are clearly communicated with helpful feedback
- The interface feels responsive and intuitive for both casual and experienced chess players
- Configurable display options accommodate different user preferences and terminal capabilities

---

## 2. Functional Requirements (The "What")

### 2.1. Board Display

**As a** user, **I want to** see the chess board rendered clearly in my terminal, **so that** I can understand the current game position at a glance.

**Acceptance Criteria:**
- [x] The board displays all 64 squares in an 8x8 grid
- [x] Each piece is represented by a distinct symbol (either Unicode chess symbols ♔♕♖♗♘♙ or ASCII letters K, Q, R, B, N, P)
- [x] White pieces and black pieces are visually distinguishable (e.g., uppercase for white, lowercase for black, or color coding)
- [x] The board includes file labels (a-h) and rank labels (1-8) around the edges
- [x] The board displays from White's perspective (rank 1 at bottom, rank 8 at top)
- [x] Empty squares are clearly distinguishable from occupied squares

**Configuration Options:**
- [x] Users can configure whether to use Unicode symbols or ASCII letters for pieces
- [x] Users can configure whether coordinate labels (files/ranks) are shown or hidden
- [x] Users can configure whether colors are used to distinguish pieces
- [x] Users can configure whether help text (navigation keys and commands) is shown or hidden

### 2.2. Move Input System

**As a** user, **I want to** enter moves using standard algebraic notation, **so that** I can play chess using familiar notation without learning a custom input format.

**Acceptance Criteria:**
- [x] The system accepts moves in Standard Algebraic Notation (SAN): `e4`, `Nf3`, `Bxc5`, `O-O`, `O-O-O`, `e8=Q`
- [x] When multiple pieces of the same type can move to the same square, the system accepts disambiguation by:
  1. File first (preferred): `Nfd2` (knight from f-file to d2)
  2. Rank if needed: `N1d2` (knight from rank 1 to d2)
  3. Both if necessary: `Nf1d2`
- [x] The system validates each move before executing it
- [x] If a move is invalid, the system displays a specific error message explaining why (e.g., "No piece at that square", "That piece cannot move there", "Move would leave king in check")
- [x] The system does not allow illegal moves to be executed
- [x] Pawn promotion moves specify the promoted piece (e.g., `e8=Q`, `a1=N`)
- [x] Castling is entered as `O-O` (kingside) or `O-O-O` (queenside)

**Explicitly NOT Included:**
- [x] Undo/takeback functionality (out of scope)

### 2.3. Move History Display

**As a** user, **I want to** optionally see the history of moves played in the current game, **so that** I can review what has happened so far.

**Acceptance Criteria:**
- [x] Move history display is configurable (can be enabled or disabled)
- [x] By default, move history is hidden (set to false)
- [x] When enabled, the system displays a list of moves in SAN format
- [x] Move history shows move numbers and both players' moves (e.g., "1. e4 e5 2. Nf3 Nc6")

### 2.4. Turn Indicator

**As a** user, **I want to** clearly see whose turn it is to move, **so that** I don't get confused during gameplay.

**Acceptance Criteria:**
- [x] The interface displays a clear indicator showing whose turn it is (e.g., "White to move", "Black to move")
- [x] The turn indicator is visible on every board display
- [x] The turn indicator updates immediately after each move

### 2.5. Game Menu & Flow

**As a** user, **I want to** navigate through menus to start games, configure options, and exit the application, **so that** I have full control over my chess experience.

**Main Menu Acceptance Criteria:**
- [x] When the application starts, a main menu is displayed with the following options:
  - "New Game"
  - "Load Game" (resume a saved game)
  - "Settings"
  - "Exit"
- [x] Users can select menu options using keyboard input
- [x] Selecting "Exit" closes the application gracefully

**New Game Flow Acceptance Criteria:**
- [x] After selecting "New Game", the system prompts the user to choose a game type:
  - "Player vs Player" (local two-player mode)
  - "Player vs Bot" (play against AI)
- [x] If "Player vs Bot" is selected, the system prompts the user to choose:
  - Bot difficulty level (Easy, Medium, Hard)
  - Specific engine/bot type (if multiple bot types are available)
- [x] After all selections are made, the game begins with the board displayed and White to move

**Exit During Active Game Acceptance Criteria:**
- [x] If a user attempts to exit or return to menu during an active game, the system prompts: "Save current game before exiting?"
- [x] If the user chooses "Yes", the game is saved (using FEN) and can be resumed later
- [x] If the user chooses "No", the game is abandoned and the user returns to the main menu
- [x] When the user next launches the application, if a saved game exists, the system offers: "Resume last game?"

**Mid-Game Options Acceptance Criteria:**
- [x] During an active game, users can access the following commands/options:
  - "Resign" (forfeit the game)
  - [x] "Offer Draw" (propose a draw to opponent, if applicable)
  - "Show FEN" (display current position as FEN string)
  - "Menu" (return to main menu with save prompt)
- [x] These options are accessible via commands or a pause menu

### 2.6. Screen Rendering

**As a** user, **I want to** see the board update cleanly after each move, **so that** the interface feels responsive and professional.

**Acceptance Criteria:**
- [x] After each move, the screen clears and the board redraws in place
- [x] The board does not append to scrolling terminal history (redraw in place for clean UX)
- [x] The redraw happens quickly enough to feel instantaneous to the user

### 2.7. Game End Screen

**As a** user, **I want to** see clear information when the game ends, **so that** I understand the outcome and can review the game.

**Acceptance Criteria:**
- [x] When the game ends (checkmate, stalemate, draw, resignation), the system displays:
  - The result announcement (e.g., "Checkmate! White wins", "Stalemate - Draw", "Black resigned - White wins")
  - The final board position
  - Game summary including total move count
- [x] After displaying the end screen, the system offers options to:
  - "Start New Game"
  - "Return to Main Menu"

### 2.8. Universal Navigation

**As a** user, **I want to** be able to navigate back to the main menu from any screen, **so that** I never feel trapped in the interface and can easily change my mind or explore different options.

**Acceptance Criteria:**
- [ ] Every screen in the application provides a way to return to the previous screen or main menu
- [ ] The ESC key is consistently used across all screens to navigate back:
  - From Settings screen → Main Menu
  - From Game Type Selection → Main Menu
  - From FEN Input screen → Main Menu
  - From Bot Selection → Game Type Selection
  - From Game Over screen → Main Menu (via menu option)
- [ ] During active gameplay, ESC key triggers a confirmation prompt before returning to menu
- [ ] Alternative navigation keys are provided where appropriate (e.g., 'b' for back, 'q' for quit to menu)
- [ ] Each screen displays help text showing the available navigation keys (e.g., "Press ESC to return to menu")
- [ ] Navigation actions are instantaneous with no delay
- [ ] Screen transitions are smooth and the user never loses context

**Exit During Active Game Navigation:**
- [ ] When user presses ESC during an active game, system prompts: "Return to menu? Current game will be saved. (y/n)"
- [ ] If user confirms (y), game is saved and user returns to Main Menu
- [ ] If user cancels (n), gameplay continues without interruption
- [ ] Ctrl+C always immediately exits the entire application from any screen

**Help Text Standards:**
- [ ] Every screen includes a help line at the bottom showing navigation options (when enabled)
- [ ] Help text follows consistent format: "ESC: back | q: quit | arrows: navigate"
- [ ] Help text is visually distinct (e.g., dimmed color, separated by whitespace)
- [ ] Help text is shown by default (enabled in default configuration)
- [ ] Users can hide help text via Settings screen by toggling "Show Help Text" option
- [ ] When help text is hidden, screens display only the primary content (board, menu, etc.)
- [ ] Help text setting persists across application restarts (saved in config.toml)

---

## 3. Scope and Boundaries

### In-Scope

- ASCII and Unicode board rendering with configurable display options
- Standard Algebraic Notation (SAN) move input with proper disambiguation
- Move validation and error feedback
- Turn indicators and game state display
- Main menu system for game navigation
- New game flow with game type and difficulty selection
- Automatic game saving on exit with resume capability
- Mid-game options (resign, draw offer, show FEN, menu access)
- Clean screen redrawing for better UX
- Game end screen with result, summary, and next action options
- Configurable move history display
- Universal navigation with ESC key support on all screens
- Consistent help text showing navigation options on every screen (configurable)
- Configurable help text visibility (can be hidden via Settings)
- Save prompts when navigating away from active games

### Out-of-Scope

- Move undo/takeback functionality (not supported)
- Chess Engine Foundation (separate spec, already complete)
- FEN Support for save/load (separate spec, already complete)
- Local Player vs Player game logic (separate spec)
- Bot opponent implementation (separate spec - Phase 2)
- Configuration file management and persistence layer (separate spec)
- Time controls or chess clocks
- Move hints or analysis features
- Opening book integration
- PGN import/export
