# Product Roadmap: TermChess

_This roadmap outlines our strategic direction based on user needs and project goals. It focuses on the "what" and "why," not the technical "how."_

---

### Phase 1

_The highest priority features that form the core foundation of the product._

- [x] **Chess Engine Foundation**
  - [x] **Board Representation:** Establish the internal chess board state that tracks piece positions and game status.
  - [x] **Move Validation & Rules:** Implement all standard chess rules including legal move detection, castling, en passant, and pawn promotion.
  - [x] **Game State Detection:** Detect check, checkmate, stalemate, and draw conditions to properly end games.

- [x] **FEN Support**
  - [x] **Export Current Position:** Allow users to copy/display the FEN string of the current game state at any time.
  - [x] **Import FEN to Start Game:** Enable users to start a new game from any valid FEN position.

- [x] **Terminal Interface**
  - [x] **ASCII/Unicode Board Display:** Render the chess board clearly in the terminal with piece symbols and coordinates.
  - [x] **Move Input System:** Accept moves in standard algebraic notation (e.g., `e4`, `Nf3`, `O-O`) with validation and error feedback.
  - [x] **Game Menu & Flow:** Provide a main menu to start new games, select game modes, and exit gracefully.

- [x] **Local Player vs Player**
  - [x] **Two-Player Mode:** Enable two humans to play against each other on the same machine, alternating turns.

- [x] **Configuration & Persistence**
  - [x] **User Config Loading:** Read user preferences from config file on startup (board style, default difficulty, etc.)
  - [x] **Game Saves:** Allow users to save and load game positions using FEN strings stored on disk.
  - [x] **Settings Application:** Apply loaded configuration to customize the game experience accordingly.

---

### Phase 2

_Once the foundational features are complete, we will move on to these high-value additions._

- [x] **Bot Opponents**
  - [x] **Easy Bot:** Create a basic AI opponent that makes legal moves with minimal strategy (random or simple heuristics).
  - [x] **Medium Bot:** Develop a moderately challenging AI using minimax or similar algorithms with position evaluation.
  - [x] **Hard Bot:** Build a stronger AI with deeper search and improved evaluation for experienced players.

---

### Phase 3

_Automated gameplay for testing, entertainment, and analysis._

- [ ] **Bot vs Bot Mode**
  - [ ] **Bot Selection:** Allow users to select two bot difficulties to play against each other (e.g., Easy vs Hard, Medium vs Medium).
  - [ ] **Automated Gameplay:** Bots play a complete game autonomously while the user watches.
  - [ ] **Playback Speed Control:** Users can adjust move speed (instant, fast, normal, slow) to control pacing.
  - [ ] **Pause/Resume:** Users can pause the game at any point to examine the position, then resume play.
  - [ ] **Multi-Game Mode:** Users can run multiple games in sequence (e.g., 10 games) between two bots.
  - [ ] **Aggregate Statistics:** After multi-game runs, display results summary (wins, losses, draws, average moves, average duration).

---

### Phase 4

_Enhanced user interaction and visual experience._

- [ ] **Mouse Interaction & UI/UX Enhancements**
  - [ ] **Click-to-Select Pieces:** Allow users to click on a piece to select it for moving.
  - [ ] **Click-to-Move:** Allow users to click on a destination square to complete a move.
  - [ ] **Visual Feedback:** Highlight selected pieces and valid move destinations on hover/click.
  - [ ] **Beautiful Board Themes:** Offer multiple color schemes and board styles (classic, modern, minimalist, etc.).
  - [ ] **Smooth Animations:** Add subtle animations for piece movement and board transitions.
  - [ ] **Enhanced Typography:** Improve text rendering with better fonts and spacing for readability.
  - [ ] **Visual Polish:** Refine borders, shadows, and spacing to create a polished, professional appearance.
  - [ ] **Accessibility Features:** Ensure color contrast ratios meet standards and add screen reader support.

---

### Phase 5

_Make the application accessible to users via simple command-line installation._

- [ ] **CLI Distribution**
  - [ ] **Release Binary Builds:** Compile standalone binaries for macOS and Linux architectures.
  - [ ] **Hosted Download Endpoint:** Host release binaries at a stable URL (e.g., GitHub Releases).
  - [ ] **Curl Install Script:** Provide a one-liner curl/wget command that downloads and installs the binary to a standard location (e.g., `/usr/local/bin`).
  - [ ] **Installation Instructions:** Document the install process in the README with copy-paste commands.

---

### Phase 6

_Features planned for future consideration. Their priority and scope may be refined based on feedback from earlier phases._

- [ ] **Custom RL Agent**
  - [ ] **RL Training Infrastructure:** Set up the framework for training a reinforcement learning chess agent.
  - [ ] **RL Bot Integration:** Integrate the trained RL model as the top-tier "expert" difficulty opponent.
  - [ ] **Iterative Improvement:** Establish a workflow for retraining and improving the RL agent over time.

- [ ] **UCI Engine Integration**
  - [ ] **UCI Protocol Support:** Implement the Universal Chess Interface protocol to communicate with external engines.
  - [ ] **External Engine Mode:** Allow users to play against any UCI-compatible engine (e.g., Stockfish, Komodo).
  - [ ] **Engine Configuration:** Let users specify engine path and basic settings (skill level, think time).
