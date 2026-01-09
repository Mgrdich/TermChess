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

- [ ] Use `lipgloss` to add color styles to piece symbols
- [ ] White pieces: one color (e.g., white/bright)
- [ ] Black pieces: another color (e.g., gray/dim)
- [ ] Respect `config.UseColors` flag
- [ ] Test: See colored pieces on board

## Slice 11: Add Game Type Selection Screen (PvP Only)

**Goal:** After "New Game", user chooses game type (PvP or PvBot), then game starts.

- [ ] Create GameTypeSelect screen state
- [ ] Add transition: MainMenu → "New Game" → GameTypeSelect → GamePlay
- [ ] Display options: "Player vs Player", "Player vs Bot"
- [ ] Add keyboard navigation for selection
- [ ] Store selected game type in model
- [ ] If "PvBot" selected, show "Coming soon" message and return to menu
- [ ] Test: Select PvP, game starts; select PvBot, see "Coming soon"

## Slice 12: Implement Configuration File Persistence

**Goal:** User preferences (display settings) are saved to `~/.termchess/config.toml` and loaded on startup.

- [ ] Create `internal/ui/persistence.go` (or extend `config.go`)
- [ ] Implement `LoadConfig()` to read `~/.termchess/config.toml`
- [ ] Implement `SaveConfig()` to write config to file
- [ ] Create `~/.termchess/` directory if it doesn't exist
- [ ] Load config on app startup, use defaults if file missing
- [ ] Test: Modify config in code, restart app, see config persisted

## Slice 13: Add Settings Screen to Change Display Options

**Goal:** User can navigate to Settings, change display options, and see them applied immediately.

- [ ] Create Settings screen state
- [ ] Add transition: MainMenu → "Settings" → Settings screen
- [ ] Display toggleable options: Unicode, Coordinates, Colors, Move History
- [ ] Allow keyboard navigation to toggle options
- [ ] Call `SaveConfig()` when option changes
- [ ] Apply config changes immediately to next board render
- [ ] Test: Change Unicode to true, see Unicode pieces; toggle coordinates off, labels disappear

## Slice 14: Implement Save Game on Exit

**Goal:** When user exits during a game, they're prompted to save; game is saved as FEN.

- [ ] Detect exit attempt during active game (Ctrl+C or quit command)
- [ ] Show prompt: "Save current game before exiting? (y/n)"
- [ ] If yes, save current board state as FEN to `~/.termchess/savegame.fen`
- [ ] If no, return to main menu or exit
- [ ] Test: Start game, make moves, press Ctrl+C, choose save, check file created

## Slice 15: Implement Resume Game on Startup

**Goal:** When user launches app with saved game, they're prompted to resume.

- [ ] On app startup, check for `~/.termchess/savegame.fen`
- [ ] If exists, show prompt: "Resume last game? (y/n)"
- [ ] If yes, load FEN using `engine.FromFEN()`, start GamePlay screen
- [ ] If no, show main menu
- [ ] When game ends normally, delete `savegame.fen`
- [ ] Test: Save game, restart app, choose resume, see board state restored

## Slice 16: Add "Load Game" Menu Option for FEN Input

**Goal:** User can select "Load Game" from main menu and enter a FEN string to load a position.

- [ ] Create FENInput screen state
- [ ] Add transition: MainMenu → "Load Game" → FENInput screen
- [ ] Display text input for FEN string
- [ ] Parse FEN using `engine.FromFEN()` on Enter
- [ ] If valid, load board and start GamePlay
- [ ] If invalid, show error message
- [ ] Test: Enter valid FEN, see position loaded; enter invalid FEN, see error

## Slice 17: Add Move History Display (Optional, Configurable)

**Goal:** When enabled in config, user sees move history during game.

- [ ] Add `moveHistory []engine.Move` to model
- [ ] After each move, append to moveHistory
- [ ] Create `FormatSAN()` function to convert Move to SAN for display
- [ ] In View, if `config.ShowMoveHistory = true`, render move history
- [ ] Format as "1. e4 e5 2. Nf3 Nc6" (numbered pairs)
- [ ] Test: Enable show_move_history in config, play moves, see history displayed

## Slice 18: Add Mid-Game Commands (Resign, Show FEN, Menu)

**Goal:** During gameplay, user can type commands to resign, show FEN, or return to menu.

- [ ] Recognize special input commands: `resign`, `showfen`, `menu`
- [ ] `resign`: Transition to GameOver screen with resignation message
- [ ] `showfen`: Display current FEN string (copy to clipboard if possible)
- [ ] `menu`: Prompt to save, then return to MainMenu
- [ ] Test: Play game, type `resign`, see game over; type `showfen`, see FEN displayed

## Slice 19: Add Coordinate Label Toggle

**Goal:** User can hide file/rank labels via config.

- [ ] Update BoardRenderer to check `config.ShowCoords`
- [ ] If false, render board without file/rank labels
- [ ] Test: Toggle show_coordinates in config, see labels appear/disappear

## Slice 20: Final Polish and Testing

**Goal:** Ensure all features work together, fix bugs, add final touches.

- [ ] Run full game test: main menu → new game → play full game → game over → new game
- [ ] Test all screen transitions
- [ ] Test save/resume flow multiple times
- [ ] Test FEN load with various positions
- [ ] Ensure error messages are clear and helpful
- [ ] Run `golangci-lint` and fix any issues
- [ ] Verify test coverage > 70% for UI logic
- [ ] Test on macOS, Linux, and Windows (if possible)
- [ ] Ensure no terminal scrollback pollution (clean redraws)
