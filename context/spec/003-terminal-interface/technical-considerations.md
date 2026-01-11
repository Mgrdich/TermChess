# Technical Specification: Terminal Interface

- **Feature:** Terminal Interface
- **Status:** Draft
- **Author:** Claude Code
- **Date:** 2026-01-08

---

## 1. Executive Summary

This specification defines the technical architecture for the Terminal Interface feature of TermChess. The interface will provide ASCII/Unicode board rendering, Standard Algebraic Notation (SAN) move input, Bubbletea-based menu navigation, and game state management with save/resume functionality.

**Key Design Principles:**
- Thin entry point with minimal CLI logic in `main.go`
- Separation of UI layer (`internal/ui`) from game logic (`internal/engine`)
- Bubbletea model for reactive, event-driven UI updates
- Clean screen redraws with no terminal scrollback pollution
- Extensive testing of game flow and rendering logic

---

## 2. Project Architecture Analysis

### 2.1 Current Project Structure

```
TermChess/
├── cmd/
│   └── termchess/
│       └── main.go              # Entry point (28 lines, minimal)
├── internal/
│   ├── engine/                  # Chess engine (12,562 lines total)
│   │   ├── types.go             # Core types: Color, PieceType, Piece, Square
│   │   ├── board.go             # Board state and operations
│   │   ├── moves.go             # Move generation and validation
│   │   ├── fen.go               # FEN import/export
│   │   ├── game_state.go        # Game status detection
│   │   ├── attacks.go           # Attack/defense calculations
│   │   ├── zobrist.go           # Position hashing for repetition
│   │   └── *_test.go            # Comprehensive test suite
│   └── util/
│       └── clipboard.go         # Cross-platform clipboard support
├── go.mod                       # Dependencies: clipboard only
├── Makefile                     # Build tasks
└── .golangci.yml                # Linter configuration
```

### 2.2 Chess Engine API

The engine provides a clean, well-tested API:

**Core Types** (`internal/engine/types.go`):
```go
type Color uint8          // White = 0, Black = 1
type PieceType uint8      // Empty, Pawn, Knight, Bishop, Rook, Queen, King
type Piece uint8          // Encoded: color (high bit) + type (low 3 bits)
type Square int8          // 0-63 (a1=0, h8=63), NoSquare=-1

// Square methods
func (s Square) File() int            // 0-7 (a-h)
func (s Square) Rank() int            // 0-7 (1-8)
func (s Square) String() string       // "e4", "a1", etc.
```

**Board State** (`internal/engine/board.go`):
```go
type Board struct {
    Squares [64]Piece       // Piece positions
    ActiveColor Color       // Whose turn
    CastlingRights uint8    // KQkq flags
    EnPassantSq int8        // En passant target (-1 if none)
    HalfMoveClock uint8     // For fifty-move rule
    FullMoveNum uint16      // Move counter
    Hash uint64             // Zobrist hash
    History []uint64        // Position history for repetition
}

// Board methods
func NewBoard() *Board                  // Standard starting position
func (b *Board) Copy() *Board          // Deep copy
func (b *Board) PieceAt(sq Square) Piece
func (b *Board) String() string        // ASCII board representation
```

**Move Operations** (`internal/engine/moves.go`):
```go
type Move struct {
    From      Square
    To        Square
    Promotion PieceType    // Empty if not a promotion
}

// Move methods
func ParseMove(s string) (Move, error)  // "e2e4", "a7a8q"
func (m Move) String() string           // Coordinate notation

// Board move operations
func (b *Board) MakeMove(m Move) error  // Apply and validate
func (b *Board) LegalMoves() []Move     // All legal moves
func (b *Board) IsLegalMove(m Move) bool
func (b *Board) InCheck() bool          // King under attack
```

**Game State** (`internal/engine/game_state.go`):
```go
type GameStatus int  // Ongoing, Checkmate, Stalemate, Draw*

func (b *Board) Status() GameStatus
func (b *Board) IsGameOver() bool
func (b *Board) CanClaimDraw() bool
func (b *Board) Winner() (Color, bool)
```

**FEN Support** (`internal/engine/fen.go`):
```go
func FromFEN(fen string) (*Board, error)
func (b *Board) ToFEN() string
```

### 2.3 Existing Patterns

**Testing:**
- Table-driven tests with subtests (`t.Run()`)
- Comprehensive coverage (12,562 lines total, ~60% tests)
- Perft validation for move generation correctness
- Example from `board_test.go`:
```go
t.Run("White back rank pieces", func(t *testing.T) {
    expectedPieces := []struct {
        square    string
        pieceType PieceType
    }{
        {"a1", Rook},
        {"b1", Knight},
        // ...
    }
    for _, expected := range expectedPieces {
        // Test logic
    }
})
```

**Code Quality:**
- golangci-lint with standard linters enabled
- Clear separation of concerns
- Exported types with documentation
- Error handling with descriptive messages

**No External Dependencies:**
- Only dependency: `golang.design/x/clipboard` for clipboard support
- Pure Go stdlib for everything else
- **No Bubbletea yet** - this will be added

---

## 3. Technical Design

### 3.1 Package Structure

```
internal/
├── engine/           # Existing - DO NOT MODIFY
│   └── ...
├── util/             # Existing - extend as needed
│   ├── clipboard.go
│   └── san.go        # NEW: SAN parser (algebraic notation → coordinate)
└── ui/               # NEW: Terminal interface
    ├── model.go      # Bubbletea model and state
    ├── view.go       # Rendering (board, menus, prompts)
    ├── update.go     # Event handling and state transitions
    ├── board.go      # Board rendering logic
    ├── san.go        # SAN move parsing and formatting
    ├── config.go     # Display configuration
    └── *_test.go     # UI logic tests
```

### 3.2 Bubbletea Integration

**Why Bubbletea?**
- Reactive, event-driven architecture (Elm-inspired)
- Clean separation of Model (state), View (render), Update (events)
- Handles terminal resizing, keyboard input, and screen management
- No global state - everything is explicit
- Well-tested library with excellent documentation

**Installation:**
```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest   # For styling
go get github.com/charmbracelet/bubbles@latest    # For components
```

**Bubbletea Model Pattern:**
```go
type Model struct {
    // State
    board *engine.Board
    screen Screen  // MainMenu, GamePlay, GameOver, etc.
    config Config  // Display settings
    
    // UI State
    input string
    error string
    history []string
}

func (m Model) Init() tea.Cmd { /* ... */ }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { /* ... */ }
func (m Model) View() string { /* ... */ }
```

### 3.3 Screen Flow State Machine

```
┌─────────────┐
│  MainMenu   │
└─────┬───────┘
      │
      ├─→ NewGame ───→ ┌─────────────────┐
      │                │ GameTypeSelect  │
      │                └────────┬────────┘
      │                         │
      │                         ├─→ PvP ──────┐
      │                         └─→ PvBot ────┤
      │                                        │
      ├─→ LoadFEN ───────────────────────────►┤
      │                                        │
      │                                   ┌────▼────┐
      │                                   │GamePlay │
      │                                   └────┬────┘
      │                                        │
      │                                        ├─→ Move Input
      │                                        ├─→ Check Status
      │                                        ├─→ Draw Offer
      │                                        └─→ Resign
      │                                             │
      ├─→ Settings ──→ (Config) ──────────────────►│
      │                                             │
      └────────────────────────────────────────────┴─→ ┌──────────┐
                                                        │ GameOver │
                                                        └──────────┘
```

### 3.4 Data Structures

**UI Model** (`internal/ui/model.go`):
```go
// Screen represents the current UI screen
type Screen int

const (
    ScreenMainMenu Screen = iota
    ScreenGameTypeSelect
    ScreenBotSelect
    ScreenFENInput
    ScreenGamePlay
    ScreenGameOver
    ScreenSettings
)

// Config holds display configuration
type Config struct {
    UseUnicode     bool     // Unicode pieces vs ASCII
    ShowCoords     bool     // Show file/rank labels
    UseColors      bool     // Color piece symbols
    ShowMoveHistory bool    // Display move history
    ShowHelpText   bool     // Display navigation help text
}

// Model is the Bubbletea model
type Model struct {
    // Game state
    board        *engine.Board
    moveHistory  []engine.Move
    
    // UI state
    screen       Screen
    config       Config
    
    // Input state
    input        string
    errorMsg     string
    statusMsg    string
    
    // Menu state
    menuSelection int
    menuOptions   []string
    
    // Game metadata
    gameType     GameType
    botDifficulty BotDifficulty  // For future bot support
}

type GameType int
const (
    GameTypePvP GameType = iota
    GameTypePvBot
)

type BotDifficulty int
const (
    BotEasy BotDifficulty = iota
    BotMedium
    BotHard
)
```

**Message Types** (`internal/ui/update.go`):
```go
// Custom messages for state transitions
type moveMsg struct {
    move engine.Move
}

type errorMsg struct {
    err error
}

type gameOverMsg struct {
    status engine.GameStatus
}
```

### 3.5 Board Rendering

**ASCII vs Unicode:**
```go
// ASCII rendering (default)
8 r n b q k b n r
7 p p p p p p p p
6 . . . . . . . .
5 . . . . . . . .
4 . . . . . . . .
3 . . . . . . . .
2 P P P P P P P P
1 R N B Q K B N R
  a b c d e f g h

// Unicode rendering (config option)
8 ♜ ♞ ♝ ♛ ♚ ♝ ♞ ♜
7 ♟ ♟ ♟ ♟ ♟ ♟ ♟ ♟
6 · · · · · · · ·
5 · · · · · · · ·
4 · · · · · · · ·
3 · · · · · · · ·
2 ♙ ♙ ♙ ♙ ♙ ♙ ♙ ♙
1 ♖ ♘ ♗ ♕ ♔ ♗ ♘ ♖
  a b c d e f g h
```

**Rendering Implementation** (`internal/ui/board.go`):
```go
type BoardRenderer struct {
    config Config
}

func (r *BoardRenderer) Render(b *engine.Board) string {
    // Use lipgloss for styling
    // Return multi-line string representation
}

func (r *BoardRenderer) pieceSymbol(p engine.Piece) string {
    if r.config.UseUnicode {
        return unicodeSymbol(p)
    }
    return asciiSymbol(p)
}
```

**Help Text Rendering** (`internal/ui/view.go`):
```go
// renderHelpText conditionally renders help text based on config
func renderHelpText(text string, config Config) string {
    if !config.ShowHelpText {
        return ""
    }

    // Use dimmed/subtle styling for help text
    helpStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#666666")).
        Padding(1, 0, 0, 0)

    return helpStyle.Render(text)
}

// Example usage in view rendering:
func (m Model) renderGamePlay() string {
    var b strings.Builder

    // ... render board and game state ...

    // Conditionally render help text
    helpText := renderHelpText("ESC: menu | Type move (e.g. e4, Nf3)", m.config)
    if helpText != "" {
        b.WriteString("\n")
        b.WriteString(helpText)
    }

    return b.String()
}
```

### 3.6 SAN Move Parsing

**Challenge:** Engine uses coordinate notation (`e2e4`), but users expect SAN (`e4`, `Nf3`).

**Solution:** Implement SAN parser that converts algebraic notation to coordinate moves.

**SAN Parser** (`internal/ui/san.go`):
```go
// ParseSAN converts SAN to a Move
// Examples: "e4" → e2e4, "Nf3" → g1f3, "Bxc5" → f1c5
func ParseSAN(b *engine.Board, san string) (engine.Move, error) {
    // 1. Strip check/checkmate symbols (+, #)
    // 2. Detect castling (O-O, O-O-O)
    // 3. Parse piece type (K, Q, R, B, N, or pawn)
    // 4. Parse disambiguation (file/rank)
    // 5. Parse destination square
    // 6. Parse capture marker (x)
    // 7. Parse promotion (=Q)
    // 8. Find matching legal move
    // 9. Return move or error
}

// FormatSAN converts a Move to SAN
func FormatSAN(b *engine.Board, m engine.Move) string {
    // Used for move history display
}
```

**Disambiguation Logic:**
- If multiple pieces can reach the destination, disambiguate by:
  1. File (preferred): `Nfd2` (knight from f-file)
  2. Rank: `N1d2` (knight from rank 1)
  3. Both: `Nf1d2` (knight from f1)

**Algorithm:**
```go
func ParseSAN(b *engine.Board, san string) (engine.Move, error) {
    // Get all legal moves for validation
    legalMoves := b.LegalMoves()
    
    // Parse the SAN string to extract:
    // - pieceType (default: Pawn)
    // - fromFile, fromRank (for disambiguation, -1 if not specified)
    // - toSquare
    // - isCapture
    // - promotion
    
    // Filter legal moves that match the SAN specification
    var candidates []engine.Move
    for _, move := range legalMoves {
        piece := b.PieceAt(move.From)
        
        // Check piece type
        if piece.Type() != pieceType {
            continue
        }
        
        // Check destination
        if move.To != toSquare {
            continue
        }
        
        // Check disambiguation
        if fromFile >= 0 && move.From.File() != fromFile {
            continue
        }
        if fromRank >= 0 && move.From.Rank() != fromRank {
            continue
        }
        
        // Check promotion
        if promotion != engine.Empty && move.Promotion != promotion {
            continue
        }
        
        candidates = append(candidates, move)
    }
    
    if len(candidates) == 0 {
        return engine.Move{}, fmt.Errorf("no legal move matches '%s'", san)
    }
    
    if len(candidates) > 1 {
        return engine.Move{}, fmt.Errorf("ambiguous move '%s', please specify file or rank", san)
    }
    
    return candidates[0], nil
}
```

### 3.7 Save/Resume Functionality

**Save on Exit:**
```go
// Save current game state to ~/.termchess/savegame.fen
func SaveGame(b *engine.Board, filename string) error {
    fen := b.ToFEN()
    // Write to file in user's home directory
    return os.WriteFile(filename, []byte(fen), 0644)
}

// Load saved game
func LoadGame(filename string) (*engine.Board, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return engine.FromFEN(string(data))
}
```

**File Location:**
- macOS/Linux: `~/.termchess/savegame.fen`
- Windows: `%USERPROFILE%\.termchess\savegame.fen`

**Auto-save behavior:**
- On exit during active game: prompt "Save game?"
- On launch: check for `savegame.fen`, prompt "Resume last game?"
- After game ends: delete `savegame.fen`

### 3.8 Entry Point Design

**Minimal `main.go`** (`cmd/termchess/main.go`):
```go
package main

import (
    "fmt"
    "os"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/Mgrdich/TermChess/internal/ui"
)

func main() {
    // Initialize the Bubbletea model
    model := ui.NewModel()
    
    // Create the Bubbletea program
    p := tea.NewProgram(
        model,
        tea.WithAltScreen(),       // Use alternate screen buffer
        tea.WithMouseCellMotion(), // Future: mouse support
    )
    
    // Run the program
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

---

## 4. Implementation Plan

### Phase 1: Bubbletea Foundation
1. Add Bubbletea dependencies to `go.mod`
2. Create `internal/ui/` package structure
3. Implement basic Model/View/Update skeleton
4. Add main menu screen with keyboard navigation
5. Test: Menu navigation works, can exit cleanly

### Phase 2: Board Rendering
1. Implement `internal/ui/board.go` with ASCII rendering
2. Add Unicode rendering option
3. Add coordinate labels (configurable)
4. Add color support via lipgloss
5. Test: Board renders correctly for all positions

### Phase 3: SAN Move Input
1. Implement `internal/ui/san.go` SAN parser
2. Add move validation and error feedback
3. Implement move input prompt in GamePlay screen
4. Test: All SAN formats parse correctly

### Phase 4: Game Flow
1. Implement game type selection screen
2. Add FEN input screen
3. Implement gameplay loop (move → validate → update → check status)
4. Add game over screen
5. Test: Full game can be played start to finish

### Phase 5: Save/Resume
1. Implement save/load functions
2. Add save prompt on exit
3. Add resume prompt on launch
4. Test: Games save and resume correctly

### Phase 6: Polish
1. Add settings screen for display config
2. Implement move history display
3. Add resign/draw offer commands
4. Test: All features work together

---

## 5. Testing Strategy

### Unit Tests

**Board Rendering** (`internal/ui/board_test.go`):
```go
func TestBoardRender(t *testing.T) {
    tests := []struct {
        name   string
        fen    string
        config Config
        expect string
    }{
        {
            name: "starting position ASCII",
            fen: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
            config: Config{UseUnicode: false, ShowCoords: true},
            expect: "8 r n b q k b n r\n...",
        },
        // More test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            board, _ := engine.FromFEN(tt.fen)
            renderer := BoardRenderer{config: tt.config}
            got := renderer.Render(board)
            if got != tt.expect {
                t.Errorf("expected:\n%s\ngot:\n%s", tt.expect, got)
            }
        })
    }
}
```

**SAN Parsing** (`internal/ui/san_test.go`):
```go
func TestParseSAN(t *testing.T) {
    tests := []struct {
        name    string
        fen     string
        san     string
        wantErr bool
        wantMove string  // Coordinate notation
    }{
        {
            name: "pawn move",
            fen: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
            san: "e4",
            wantErr: false,
            wantMove: "e2e4",
        },
        {
            name: "knight move",
            fen: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
            san: "Nf3",
            wantErr: false,
            wantMove: "g1f3",
        },
        {
            name: "disambiguation by file",
            fen: "r1bqkb1r/pppp1ppp/2n2n2/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4",
            san: "Nbd2",  // Knight from b-file to d2
            wantErr: false,
            wantMove: "b1d2",
        },
        {
            name: "castling kingside",
            fen: "rnbqkb1r/pppp1ppp/5n2/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 4 3",
            san: "O-O",
            wantErr: false,
            wantMove: "e1g1",
        },
        {
            name: "invalid move",
            fen: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
            san: "e5",  // No pawn can move to e5
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            board, _ := engine.FromFEN(tt.fen)
            move, err := ParseSAN(board, tt.san)
            
            if tt.wantErr {
                if err == nil {
                    t.Error("expected error, got nil")
                }
                return
            }
            
            if err != nil {
                t.Errorf("unexpected error: %v", err)
                return
            }
            
            if move.String() != tt.wantMove {
                t.Errorf("expected move %s, got %s", tt.wantMove, move.String())
            }
        })
    }
}
```

### Integration Tests

**Full Game Flow** (`internal/ui/model_test.go`):
```go
func TestGameFlow(t *testing.T) {
    // Test: New game → play moves → checkmate → game over screen
    model := NewModel()
    
    // Start new game
    model, _ = model.Update(selectMenuOption(0)) // New Game
    model, _ = model.Update(selectMenuOption(0)) // PvP
    
    // Verify game started
    if model.screen != ScreenGamePlay {
        t.Error("expected GamePlay screen")
    }
    
    // Play Scholar's Mate
    moves := []string{"e4", "e5", "Bc4", "Nc6", "Qh5", "Nf6", "Qxf7#"}
    for _, san := range moves {
        model, _ = model.Update(inputMove(san))
        // Check no errors
    }
    
    // Verify game over
    if model.screen != ScreenGameOver {
        t.Error("expected GameOver screen after checkmate")
    }
}
```

---

## 6. Configuration and Defaults

**Default Configuration:**
- UseUnicode: false (ASCII for maximum compatibility)
- ShowCoords: true (Show a-h, 1-8 labels)
- UseColors: true (Use colors if terminal supports)
- ShowMoveHistory: false (Hidden by default per spec)
- ShowHelpText: true (Show navigation help by default)

**Configuration File Persistence:**

Configuration is persisted in TOML format at `~/.termchess/config.toml`:

```toml
[display]
use_unicode = false
show_coordinates = true
use_colors = true
show_move_history = false
show_help_text = true

[game]
default_game_type = "pvp"
default_bot_difficulty = "medium"
```

**File Structure:**
```
~/.termchess/
├── config.toml      # User preferences (persisted permanently)
└── savegame.fen     # Current game state (temporary, deleted on game end)
```

**Configuration Management:**
- Load `config.toml` on application startup
- Create with default values if file doesn't exist
- Save configuration when changed in Settings screen
- Configuration persists across sessions

---

## 7. Error Handling

**Move Input Errors:**
- Invalid SAN format: "Invalid move format. Try 'e4', 'Nf3', 'O-O', etc."
- No matching move: "No legal move matches 'Nf5'"
- Ambiguous move: "Ambiguous move 'Nd2'. Please specify: Nbd2 or Nfd2"
- Move leaves king in check: "That move would leave your king in check"

**FEN Input Errors:**
- Invalid FEN: "Invalid FEN string. Please check the format and try again."
- Displays specific error from `engine.FromFEN()`

**File I/O Errors:**
- Save failed: "Failed to save game: [error details]"
- Load failed: "Failed to load game: [error details]"

---

## 8. Dependencies

**New Dependencies:**
```go
require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/lipgloss v0.9.1
    github.com/charmbracelet/bubbles v0.18.0  // For input components
    github.com/BurntSushi/toml v1.3.2         // For config file parsing
)
```

**Existing Dependencies:**
```go
require golang.design/x/clipboard v0.7.1
```

---

## 9. Compatibility and Constraints

**Terminal Requirements:**
- Minimum: 80x24 characters
- ANSI color support (optional, degrades gracefully)
- UTF-8 support (for Unicode pieces, falls back to ASCII)

**Platform Support:**
- macOS: Full support
- Linux: Full support (requires X11/Wayland for clipboard)
- Windows: Full support (Windows 10+ recommended for Unicode)

**Performance:**
- Board rendering: < 1ms
- Move parsing: < 1ms
- Full screen redraw: < 10ms

---

## 10. Future Enhancements (Out of Scope)

The following are explicitly NOT part of this specification:

1. **Bot Opponents** - AI move selection (Phase 2)
2. **Time Controls** - Chess clocks
3. **Move Undo** - Takeback functionality
4. **PGN Support** - Import/export in PGN format
5. **Analysis Mode** - Move hints, evaluation
6. **Mouse Input** - Click-to-move interface
7. **Network Play** - Remote multiplayer
8. **Opening Book** - Opening move database

---

## 11. Success Criteria

The implementation is complete when:

- [ ] User can start a new PvP game from the main menu
- [ ] User can input moves using SAN notation (e4, Nf3, O-O, etc.)
- [ ] Board renders correctly in ASCII and Unicode modes
- [ ] Move validation provides clear error messages
- [ ] Game detects checkmate/stalemate/draws correctly
- [ ] User can save a game on exit and resume it later
- [ ] User can load a position from FEN string
- [ ] User preferences are saved to config.toml and persist across sessions
- [ ] Settings screen allows configuration changes that are immediately persisted
- [ ] All screens (menu, gameplay, game over) work correctly
- [ ] No terminal scrollback pollution (clean redraws)
- [ ] Test coverage for UI logic > 70%
- [ ] golangci-lint passes with no errors

---

## 12. References

### Existing Codebase
- Chess Engine: `/Users/mgo/Documents/TermChess/internal/engine/`
- Entry Point: `/Users/mgo/Documents/TermChess/cmd/termchess/main.go`
- Functional Spec: `/Users/mgo/Documents/TermChess/context/spec/003-terminal-interface/functional-spec.md`

### External Documentation
- Bubbletea: https://github.com/charmbracelet/bubbletea
- Lipgloss: https://github.com/charmbracelet/lipgloss
- SAN Notation: https://en.wikipedia.org/wiki/Algebraic_notation_(chess)
- FEN Format: https://en.wikipedia.org/wiki/Forsyth%E2%80%93Edwards_Notation

---

## 13. Open Questions and Decisions

1. **SAN Parsing Complexity**: Should we handle all edge cases (e.g., `Nbd2`, `N1d2`) in Phase 3, or start with basic moves and iterate?
   - **Decision:** Implement full SAN support in Phase 3. It's a core feature.

2. **Move History Storage**: Should we store moves as `[]engine.Move` or as `[]string` (SAN)?
   - **Decision:** Store as `[]engine.Move` for accuracy, convert to SAN for display.

3. **Configuration Persistence**: Should display settings persist across sessions?
   - **Decision:** Out of scope for this spec. Use in-memory defaults only.

4. **Error Recovery**: If SAN parsing fails, should we fall back to coordinate notation?
   - **Decision:** Yes. Accept both SAN (`e4`) and coordinate (`e2e4`) input.

5. **Bot Difficulty Selection**: Should UI include bot selection even though bot logic is not implemented?
   - **Decision:** Yes. Show menu option but display "Coming soon" message.

---

## Appendix A: SAN Grammar

```
SAN := [Piece][Disambiguation][Capture]Square[Promotion][Check]
     | Castling[Check]

Piece := 'K' | 'Q' | 'R' | 'B' | 'N' | (empty for pawn)
Disambiguation := File | Rank | FileRank
Capture := 'x'
Square := File Rank
Promotion := '=' ('Q' | 'R' | 'B' | 'N')
Check := '+' | '#'
Castling := 'O-O' | 'O-O-O'

File := 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h'
Rank := '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8'
```

**Examples:**
- `e4` - Pawn to e4
- `Nf3` - Knight to f3
- `Bxc5` - Bishop captures on c5
- `Nbd2` - Knight from b-file to d2
- `N1d2` - Knight from rank 1 to d2
- `e8=Q` - Pawn promotes to queen on e8
- `O-O` - Kingside castling
- `Qh5#` - Queen to h5, checkmate

---

## Appendix B: Bubbletea Example

**Minimal Bubbletea Program:**
```go
package main

import (
    "fmt"
    tea "github.com/charmbracelet/bubbletea"
)

type model struct {
    choices  []string
    cursor   int
    selected string
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "up", "k":
            if m.cursor > 0 {
                m.cursor--
            }
        case "down", "j":
            if m.cursor < len(m.choices)-1 {
                m.cursor++
            }
        case "enter":
            m.selected = m.choices[m.cursor]
            return m, tea.Quit
        }
    }
    return m, nil
}

func (m model) View() string {
    s := "Select an option:\n\n"
    for i, choice := range m.choices {
        cursor := " "
        if m.cursor == i {
            cursor = ">"
        }
        s += fmt.Sprintf("%s %s\n", cursor, choice)
    }
    s += "\nPress q to quit.\n"
    return s
}

func main() {
    p := tea.NewProgram(model{
        choices: []string{"New Game", "Load Game", "Settings", "Exit"},
    })
    
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v", err)
    }
}
```

---

**End of Technical Specification**
