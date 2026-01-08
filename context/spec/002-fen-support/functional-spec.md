# Functional Specification: FEN Support

- **Roadmap Item:** FEN Support — Export Current Position, Import FEN to Start Game
- **Status:** Completed
- **Author:** Poe

---

## 1. Overview and Rationale (The "Why")

FEN (Forsyth-Edwards Notation) is the standard format for describing chess positions. FEN Support enables users to save their game state as a portable string and resume play later, or start a game from any valid position.

**Problem it solves:** Users may need to interrupt a game and resume later, practice specific positions, or share game states. Without FEN support, there's no way to capture or restore a position.

**User value:**
- **Dev Dave:** Can quickly save a position when interrupted, paste the FEN somewhere, and resume later
- **CLI Chris:** Can set up specific positions to practice tactics or study endgames

**Success criteria:**
- Users can export the current position as a FEN string with a simple command
- Users can start a new game from any valid FEN position
- Invalid FEN input is rejected with clear feedback

---

## 2. Functional Requirements (The "What")

### 2.1 Export Current Position (`/fen` command)

The user can export the current board position as a FEN string at any time during a game.

**User Flow:**
1. During a game, user types `/fen` (or presses a designated key)
2. The FEN string is printed to the screen
3. The FEN string is also copied to the system clipboard

**Acceptance Criteria:**
- [ ] Given a game in progress, when the user enters `/fen`, then the current position's FEN string is printed to the terminal.
- [ ] Given a game in progress, when the user enters `/fen`, then the FEN string is copied to the system clipboard.
- [ ] The FEN string must be valid and accurately represent: piece positions, active color, castling rights, en passant square, halfmove clock, and fullmove number.

---

### 2.2 Import FEN to Start Game (Menu Option)

The user can start a new game from any valid FEN position via a menu option.

**User Flow:**
1. From the main menu, user selects "Start from FEN"
2. User is prompted to enter a FEN string
3. The system validates the FEN string
4. If valid: "Position loaded successfully" is displayed, then the game starts with the imported position
5. If invalid: "Invalid FEN string. Please check the format and try again." is displayed, user can retry or go back

**Acceptance Criteria:**
- [ ] Given the main menu, when the user selects "Start from FEN", then they are prompted to enter a FEN string.
- [ ] Given a valid FEN string is entered, when submitted, then "Position loaded successfully" is displayed and the game begins with that position.
- [ ] Given an invalid FEN string is entered, when submitted, then "Invalid FEN string. Please check the format and try again." is displayed.
- [ ] After an invalid FEN error, the user can re-enter a FEN string or return to the main menu.

---

### 2.3 FEN Validation

The system must validate FEN strings before accepting them.

**Validation checks:**
- Correct number of fields (6 fields separated by spaces)
- Valid piece placement (8 ranks, valid piece characters, correct square counts)
- Valid active color (`w` or `b`)
- Valid castling rights (`KQkq`, `-`, or subset)
- Valid en passant square (algebraic notation or `-`)
- Valid halfmove clock (non-negative integer)
- Valid fullmove number (positive integer)

**Acceptance Criteria:**
- [ ] Given a FEN string with incorrect field count, when validated, then it is rejected.
- [ ] Given a FEN string with invalid piece characters, when validated, then it is rejected.
- [ ] Given a FEN string with invalid rank structure, when validated, then it is rejected.
- [ ] Given a fully valid FEN string, when validated, then it is accepted.

---

## 3. Scope and Boundaries

### In-Scope

- `/fen` command to export current position (print + clipboard)
- "Start from FEN" menu option to import a position
- FEN string validation with generic error message
- Standard FEN format (6 fields)

### Out-of-Scope

The following are separate roadmap items and NOT included in this specification:

- **Chess Engine Foundation** — Board representation, move validation, game state detection (separate spec)
- **Terminal Interface** — Board display, move input system, game menu (separate spec)
- **Local Player vs Player** — Two-player mode (separate spec)
- **Configuration & Persistence** — Config loading, game saves to disk (separate spec)
- **Bot Opponents** — AI opponents (Phase 2)
- **CLI Distribution** — Binary builds, install scripts (Phase 3)
- **Custom RL Agent** — RL training and integration (Phase 4)
- **UCI Engine Integration** — External engine support (Phase 4)
- **Mouse Interaction** — Click-based input (Phase 5)
