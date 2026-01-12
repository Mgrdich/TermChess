# Task List: Terminal Interface

## Slice 1: Basic Bubbletea Foundation with Main Menu

**Goal:** User can launch the app, see a main menu, and exit gracefully.

- [x] Add Bubbletea dependencies (`bubbletea`, `lipgloss`, `bubbles`, `toml`) to `go.mod`
- [x] Create `internal/ui/` package directory structure
- [x] Create `internal/ui/model.go` with basic Model struct (screen state, menu options)
- [x] Create `internal/ui/view.go` with View function that renders main menu
- [x] Create `internal/ui/update.go` with Update function for keyboard navigation
- [x] Update `cmd/termchess/main.go` to initialize Bubbletea program with alternate screen
- [x] Test: Run app, navigate menu with arrow keys, quit with 'q' or Ctrl+C

## Slice 2: Display Static Board in ASCII

**Goal:** User can select "New Game" and see a chess board (no moves yet).

- [x] Create `internal/ui/config.go` with Config struct (UseUnicode, ShowCoords, UseColors, ShowMoveHistory)
- [x] Create `internal/ui/board.go` with BoardRenderer struct
- [x] Implement ASCII board rendering for starting position (with coordinates)
- [x] Add GamePlay screen state to model
- [x] Add screen transition: MainMenu → "New Game" → GamePlay screen
- [x] Update View to render board when in GamePlay screen
- [x] Test: Select "New Game", see ASCII board with pieces in starting position

## Slice 3: Add Turn Indicator and Basic Input Prompt

**Goal:** User sees whose turn it is and a prompt for move input (input not yet functional).

- [x] Add turn indicator text to board display ("White to move" / "Black to move")
- [x] Add input prompt at bottom of screen ("Enter move: ")
- [x] Add input field state to model (text input, no SAN parsing yet)
- [x] Update View to show input prompt
- [x] Test: See turn indicator and input prompt (typing doesn't execute moves yet)

## Slice 4: Parse and Execute Simple Pawn Moves (Coordinate Notation)

**Goal:** User can enter a pawn move in coordinate notation (`e2e4`) and see the board update.

- [x] Add board state (`*engine.Board`) to model, initialize with `engine.NewBoard()`
- [x] In Update function, handle Enter key to parse input as coordinate move
- [x] Use `engine.ParseMove()` to convert string to Move
- [x] Call `board.MakeMove()` to execute move
- [x] Update board display after move
- [x] Add error message display for invalid moves
- [x] Test: Enter `e2e4`, see pawn move on board; try invalid move, see error

## Slice 5: Implement SAN Parser for Basic Pawn Moves

**Goal:** User can enter pawn moves in SAN (`e4`) instead of coordinate notation.

- [x] Create `internal/ui/san.go` with `ParseSAN()` function skeleton
- [x] Implement SAN parsing for simple pawn moves (e.g., `e4`, `d5`)
- [x] Handle pawn captures with file disambiguation (e.g., `exd5`)
- [x] Update input handler to try SAN parsing first, fall back to coordinate notation
- [x] Add unit tests in `internal/ui/san_test.go` for pawn move parsing
- [x] Test: Enter `e4`, see pawn move; enter `e2e4`, also works

## Slice 6: Extend SAN Parser for Piece Moves (No Disambiguation)

**Goal:** User can move knights, bishops, rooks, queens, kings using SAN (e.g., `Nf3`, `Bc4`).

- [x] Extend `ParseSAN()` to handle piece moves (`Nf3`, `Bc4`, `Qh5`, `Kf1`)
- [x] Handle captures (`Bxc5`, `Nxe5`)
- [x] Handle castling (`O-O`, `O-O-O`)
- [x] Handle check/checkmate symbols by stripping them (`+`, `#`)
- [x] Add unit tests for piece moves, captures, castling
- [x] Test: Play `1. e4 e5 2. Nf3 Nc6 3. Bc4` - all moves work

## Slice 7: Handle SAN Disambiguation and Pawn Promotion

**Goal:** User can enter disambiguated moves (`Nbd2`, `N1f3`) and pawn promotions (`e8=Q`).

- [x] Extend `ParseSAN()` to handle file disambiguation (`Nbd2`, `Rfe1`)
- [x] Handle rank disambiguation (`N1d2`, `R1a3`)
- [x] Handle both file+rank disambiguation (`Nb1d2`)
- [x] Handle pawn promotion (`e8=Q`, `a1=N`, `h8=R`)
- [x] Add unit tests for all disambiguation cases and promotions
- [x] Test: Play game with disambiguation; promote a pawn to queen

## Slice 8: Detect and Display Game Over (Checkmate/Stalemate)

**Goal:** When game ends, user sees game over screen with result and move count.

- [x] Add GameOver screen state to model
- [x] After each move, check `board.Status()` for game end conditions
- [x] If game over, transition to GameOver screen
- [x] Display result message ("Checkmate! White wins", "Stalemate - Draw")
- [x] Display final board position
- [x] Display move count
- [x] Add option to return to main menu or start new game
- [x] Test: Play Scholar's Mate, see game over screen

## Slice 9: Add Unicode Board Rendering Option

**Goal:** User can see board with Unicode chess symbols (♔♕♖♗♘♙) instead of ASCII.

- [x] Add Unicode piece symbol mapping in `board.go`
- [x] Update BoardRenderer to use Unicode symbols when `config.UseUnicode = true`
- [x] Toggle config in code to test (settings screen comes later)
- [x] Add unit tests for Unicode rendering
- [x] Test: Switch config flag, see board with Unicode pieces

## Slice 10: Add Color Support for Pieces

**Goal:** White and black pieces are distinguished by color (if terminal supports it).

- [x] Use `lipgloss` to add color styles to piece symbols
- [x] White pieces: one color (e.g., white/bright)
- [x] Black pieces: another color (e.g., gray/dim)
- [x] Respect `config.UseColors` flag
- [x] Test: See colored pieces on board

## Slice 11: Add Game Type Selection Screen (PvP Only)

**Goal:** After "New Game", user chooses game type (PvP or PvBot), then game starts.

- [x] Create GameTypeSelect screen state
- [x] Add transition: MainMenu → "New Game" → GameTypeSelect → GamePlay
- [x] Display options: "Player vs Player", "Player vs Bot"
- [x] Add keyboard navigation for selection
- [x] Store selected game type in model
- [x] If "PvBot" selected, show "Coming soon" message and return to menu
- [x] Test: Select PvP, game starts; select PvBot, see "Coming soon"

## Slice 12: Implement Configuration File Persistence

**Goal:** User preferences (display settings) are saved to `~/.termchess/config.toml` and loaded on startup.

- [x] Create `internal/ui/persistence.go` (or extend `config.go`)
- [x] Implement `LoadConfig()` to read `~/.termchess/config.toml`
- [x] Implement `SaveConfig()` to write config to file
- [x] Create `~/.termchess/` directory if it doesn't exist
- [x] Load config on app startup, use defaults if file missing
- [x] Test: Modify config in code, restart app, see config persisted

## Slice 13: Add Settings Screen to Change Display Options

**Goal:** User can navigate to Settings, change display options, and see them applied immediately.

- [x] Create Settings screen state
- [x] Add transition: MainMenu → "Settings" → Settings screen
- [x] Display toggleable options: Unicode, Coordinates, Colors, Move History
- [x] Allow keyboard navigation to toggle options
- [x] Call `SaveConfig()` when option changes
- [x] Apply config changes immediately to next board render
- [x] Test: Change Unicode to true, see Unicode pieces; toggle coordinates off, labels disappear
- [x] Add "Show Help Text" option to Settings screen (5th toggleable option)
- [x] Update Config struct to include ShowHelpText field (default: true)
- [x] Update view rendering to conditionally show help text based on config
- [x] Test: Toggle help text off, verify navigation hints disappear from all screens

## Slice 14: Implement Save Game on Exit

**Goal:** When user exits during a game, they're prompted to save; game is saved as FEN.

- [x] Detect exit attempt during active game (Ctrl+C or quit command)
- [x] Show prompt: "Save current game before exiting? (y/n)"
- [x] If yes, save current board state as FEN to `~/.termchess/savegame.fen`
- [x] If no, return to main menu or exit
- [x] Test: Start game, make moves, press Ctrl+C, choose save, check file created

## Slice 15: Implement Resume Game on Startup

**Goal:** When user launches app with saved game, they're prompted to resume.

- [x] On app startup, check for `~/.termchess/savegame.fen`
- [x] If exists, show prompt: "Resume last game? (y/n)"
- [x] If yes, load FEN using `engine.FromFEN()`, start GamePlay screen
- [x] If no, show main menu
- [x] When game ends normally, delete `savegame.fen`
- [x] Test: Save game, restart app, choose resume, see board state restored

## Slice 16: Add "Load Game" Menu Option for FEN Input

**Goal:** User can select "Load Game" from main menu and enter a FEN string to load a position.

- [x] Create FENInput screen state
- [x] Add transition: MainMenu → "Load Game" → FENInput screen
- [x] Display text input for FEN string
- [x] Parse FEN using `engine.FromFEN()` on Enter
- [x] If valid, load board and start GamePlay
- [x] If invalid, show error message
- [x] Test: Enter valid FEN, see position loaded; enter invalid FEN, see error

## Slice 17: Add Move History Display (Optional, Configurable)

**Goal:** When enabled in config, user sees move history during game.

- [x] Add `moveHistory []engine.Move` to model
- [x] After each move, append to moveHistory
- [x] Create `FormatSAN()` function to convert Move to SAN for display
- [x] In View, if `config.ShowMoveHistory = true`, render move history
- [x] Format as "1. e4 e5 2. Nf3 Nc6" (numbered pairs)
- [x] Test: Enable show_move_history in config, play moves, see history displayed

## Slice 18: Add Mid-Game Commands (Resign, Show FEN, Menu)

**Goal:** During gameplay, user can type commands to resign, show FEN, or return to menu.

- [x] Recognize special input commands: `resign`, `showfen`, `menu`
- [x] `resign`: Transition to GameOver screen with resignation message
- [x] `showfen`: Display current FEN string (copy to clipboard if possible)
- [x] `menu`: Prompt to save, then return to MainMenu
- [x] Test: Play game, type `resign`, see game over; type `showfen`, see FEN displayed

## Slice 19: Add Coordinate Label Toggle

**Goal:** User can hide file/rank labels via config.

- [x] Update BoardRenderer to check `config.ShowCoords`
- [x] If false, render board without file/rank labels
- [x] Test: Toggle show_coordinates in config, see labels appear/disappear

## Slice 20: Final Polish and Testing

**Goal:** Ensure all features work together, fix bugs, add final touches.

- [x] Run full game test: main menu → new game → play full game → game over → new game
- [x] Test all screen transitions
- [x] Test save/resume flow multiple times
- [x] Test FEN load with various positions
- [x] Ensure error messages are clear and helpful
- [x] Run `golangci-lint` and fix any issues
- [x] Verify test coverage > 70% for UI logic
- [x] Test on macOS, Linux, and Windows (if possible)
- [x] Ensure no terminal scrollback pollution (clean redraws)

---

## Slice 21: Add Draw Offer Command

**Goal:** Allow players to offer a draw during gameplay in Player vs Player mode.

- [x] Add `offerdraw` command recognition during gameplay
- [x] When player types "offerdraw", show prompt to opponent: "Opponent offers a draw. Accept? (y/n)"
- [x] If opponent accepts (y), transition to GameOver screen with draw message
- [x] If opponent declines (n), display "Draw offer declined" and continue game
- [x] Track draw offer state to prevent spamming (limit one offer per player per game, or cooldown)
- [x] Add draw offer to help text display
- [x] Test: Offer draw, accept it, verify game ends in draw
- [x] Test: Offer draw, decline it, verify game continues

---

## Slice 22: Implement Universal Navigation (ESC to Exit)

**Goal:** Every screen provides consistent navigation back to previous screen or main menu.

- [x] Audit all existing screens for ESC key handling
- [x] Implement ESC key navigation for GameTypeSelect screen → Main Menu
- [x] Implement ESC key navigation for FENInput screen → Main Menu
- [x] Implement ESC key navigation for BotSelect screen → GameTypeSelect
- [x] Add save prompt when ESC pressed during active GamePlay (already exists for 'q')
- [x] Ensure Settings screen ESC navigation is working (already implemented in Slice 13)
- [x] Add consistent help text to all screens showing navigation options (respects ShowHelpText config)
- [x] Update view.go to conditionally display help text for each screen based on config.ShowHelpText:
  - MainMenu: "arrows/jk: navigate | enter: select | q: quit"
  - GameTypeSelect: "ESC: back to menu | arrows: navigate | enter: select"
  - FENInput: "ESC: back to menu | enter: load position"
  - BotSelect: "ESC: back | arrows: navigate | enter: select"
  - GamePlay: "ESC: menu (with save) | type move (e.g. e4, Nf3)"
  - GameOver: "ESC: menu | arrows: navigate | enter: select"
  - Settings: "ESC: back | arrows: navigate | enter: toggle"
- [x] Create helper function `renderHelpText(text string, config Config) string` that returns empty string if ShowHelpText is false
- [x] Ensure help text is visually distinct (dimmed color, bottom of screen, separated by whitespace)
- [x] Ensure Ctrl+C always exits application immediately from any screen
- [x] Add unit tests for ESC key handling on each screen
- [x] Add unit tests for help text display/hide based on config
- [x] Test navigation flow: verify user can navigate from any screen back to menu
- [x] Test that ESC during gameplay prompts for save before returning to menu
- [x] Test that toggling ShowHelpText in Settings immediately affects next screen render

---

## Slice 23: Add Resume Game Option to Main Menu

**Goal:** Show "Resume Game" option in main menu when a saved game exists, providing easier access to continue interrupted games.

- [x] Add function to check if saved game file exists (`~/.termchess/savegame.fen`)
- [x] Modify main menu to dynamically include "Resume Game" option when saved game exists
- [x] Update menu option indices to accommodate dynamic menu items
- [x] Implement "Resume Game" selection handler to load saved game and start gameplay
- [x] Ensure "Resume Game" option appears at the top of the menu (after title, before "New Game")
- [x] Update main menu rendering to highlight the "Resume Game" option distinctly (e.g., different color or indicator)
- [x] Remove or deprecate the startup resume prompt (Slice 15) in favor of menu-based approach
- [x] Update menu navigation tests to handle dynamic menu options
- [x] Add unit tests for saved game detection logic
- [x] Add unit tests for menu option ordering with/without saved game
- [x] Add integration test: save game → return to menu → verify "Resume Game" appears
- [x] Add integration test: select "Resume Game" → verify game state restored correctly
- [x] Add integration test: complete resumed game → verify "Resume Game" option disappears from menu
- [x] Update help text to reflect "Resume Game" option availability
- [x] Test full flow: play game → save → quit → restart → see "Resume Game" in menu → select → continue game
