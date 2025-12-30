# Functional Specification: Chess Engine Foundation

- **Roadmap Item:** Chess Engine Foundation — Board Representation, Move Validation & Rules, Game State Detection
- **Status:** Draft
- **Author:** Poe

---

## 1. Overview and Rationale (The "Why")

The Chess Engine Foundation is the core logic layer that powers all gameplay in TermChess. It is the first and most critical component — without a reliable engine, no other feature (UI, bots, FEN support) can function.

**Problem it solves:** Users need a chess application that correctly enforces all standard chess rules. Incorrect rule enforcement would make the app untrustworthy and unusable for serious play.

**User value:** For "Dev Dave" and "CLI Chris," this ensures a legitimate, rule-compliant chess experience where every legal move is accepted, every illegal move is rejected, and games end correctly.

**Success criteria:**
- All standard chess rules are correctly enforced
- Games end appropriately (checkmate, stalemate, draws)
- The engine can initialize from any valid board position
- Move validation is fast and accurate

---

## 2. Functional Requirements (The "What")

### 2.1 Board Representation

The engine must maintain an internal representation of the chess board state.

- **Piece positions:** Track all 64 squares and which piece (if any) occupies each.
- **Active color:** Track whose turn it is (white or black).
- **Castling rights:** Track whether each side can still castle kingside and/or queenside.
- **En passant target:** Track the square where en passant capture is possible (if any).
- **Half-move clock:** Track moves since last pawn move or capture (for 50-move rule).
- **Full-move number:** Track the current move number.
- **Move history:** Maintain history of positions for threefold/fivefold repetition detection.

**Acceptance Criteria:**
- [ ] Given a new game, the board initializes to the standard starting position with all metadata correct.
- [ ] Given an arbitrary position (piece array + metadata), the board initializes correctly.
- [ ] The board state is queryable: what piece is on square X, whose turn, castling rights, etc.

---

### 2.2 Move Validation & Rules

The engine must validate moves and enforce all standard chess rules.

**Move Input Format:**
- Engine accepts **coordinate notation** internally: `e2e4`, `g1f3`, `e1g1` (from-square, to-square).
- For pawn promotion, include the piece: `e7e8q` (promote to queen).
- UI layer is responsible for translating algebraic notation to coordinates.

**Legal Move Generation:**
- Generate all legal moves for the current position.
- A move is legal only if it does not leave the player's own king in check.

**Standard Piece Movement:**
- **Pawn:** Forward one square (or two from starting rank), captures diagonally, en passant, promotion.
- **Knight:** L-shape (2+1), can jump over pieces.
- **Bishop:** Diagonal movement, blocked by pieces.
- **Rook:** Horizontal/vertical movement, blocked by pieces.
- **Queen:** Combination of rook and bishop.
- **King:** One square in any direction.

**Special Rules:**

| Rule | Description |
|------|-------------|
| **Castling** | King moves two squares toward rook; rook jumps over king. Requires: neither piece has moved, no pieces between, king not in check, king doesn't pass through or land in check. |
| **En passant** | Pawn captures enemy pawn that just moved two squares, as if it moved one square. Only available immediately after the two-square move. |
| **Pawn promotion** | Pawn reaching the 8th rank must promote to Queen, Rook, Bishop, or Knight (player's choice). Cannot promote to Pawn or King. |
| **Check** | King is under attack. Player must move out of check, block, or capture the attacker. |

**Acceptance Criteria:**
- [ ] Given a position, the engine returns all legal moves.
- [ ] Given a legal move, the engine applies it and updates the board state correctly.
- [ ] Given an illegal move, the engine rejects it and returns an error.
- [ ] Castling is allowed only when all conditions are met; rejected otherwise.
- [ ] En passant is available only immediately after a two-square pawn move.
- [ ] Pawn promotion requires specifying the promotion piece (Q, R, B, N).
- [ ] A player in check can only make moves that escape check.

---

### 2.3 Game State Detection

The engine must detect and declare game-ending conditions.

**Win Conditions:**
- **Checkmate:** The player to move is in check and has no legal moves. Opponent wins.

**Draw Conditions (auto-declared by engine):**

| Condition                 | Description                                                                                                        |
|---------------------------|--------------------------------------------------------------------------------------------------------------------|
| **Stalemate**             | Player to move has no legal moves but is not in check.                                                             |
| **Insufficient material** | Neither side can checkmate: K vs K, K+B vs K, K+N vs K, K+B vs K+B (same color bishops).                           |
| **Threefold repetition**  | Same position occurs three times (same piece placement, same castling rights, same en passant, same side to move). |
| **Fivefold repetition**   | Same position occurs five times — automatic draw, no claim needed.                                                 |
| **50-move rule**          | 50 consecutive moves by both players without a pawn move or capture.                                               |
| **75-move rule**          | 75 moves without pawn move or capture — automatic draw.                                                            |

**Acceptance Criteria:**
- [ ] Given a checkmate position, the engine declares the winner.
- [ ] Given a stalemate position, the engine declares a draw.
- [ ] Given insufficient material, the engine declares a draw.
- [ ] Given threefold repetition, the engine declares a draw.
- [ ] Given fivefold repetition, the engine declares a draw.
- [ ] Given 50 moves without pawn move/capture, the engine declares a draw.
- [ ] Given 75 moves without pawn move/capture, the engine declares a draw.
- [ ] The engine provides a game status: `ongoing`, `checkmate`, `stalemate`, `draw` (with reason).

---

## 3. Scope and Boundaries

### In-Scope

- Board state representation with all required metadata
- Legal move generation for all pieces
- All standard chess rules (castling, en passant, promotion, check)
- Game state detection (checkmate, stalemate, all draw conditions)
- Coordinate notation move input (`e2e4` format)
- Initialization from standard position or arbitrary position

### Out-of-Scope

The following are separate roadmap items and NOT included in this specification:

- **FEN Support** — FEN string parsing/export (separate spec)
- **Terminal Interface** — Board display, move input UI, menus (separate spec)
- **Local Player vs Player** — Game flow, turn management UI (separate spec)
- **Configuration & Persistence** — Config files, game saves (separate spec)
- **Bot Opponents** — AI move selection (Phase 2)
- **Custom RL Agent** — RL training and integration (Phase 3)
- **UCI Engine Integration** — External engine communication (Phase 3)
- **Mouse Interaction** — Click-based input (Phase 4)
- **Resign / Offer Draw** — UI-level actions, not engine logic
- **Algebraic notation parsing** — UI layer responsibility
