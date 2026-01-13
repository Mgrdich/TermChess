# Functional Specification: Bot Opponents

- **Roadmap Item:** Bot Opponents (Phase 2)
- **Status:** Draft
- **Author:** AWOS System

---

## 1. Overview and Rationale (The "Why")

### Purpose
Many users want to play chess alone for practice, skill improvement, or casual entertainment without needing another person present. Bot opponents provide this capability by offering AI players at varying skill levels, allowing users to enjoy TermChess offline and on-demand.

### Problem Statement
Currently, TermChess only supports Player vs Player mode, requiring two humans to be present at the same machine. This limits accessibility and makes the application less useful for solo practice. Users like "CLI Chris" (our target persona) want to practice against progressively harder bots to improve their chess skills without needing online connectivity or another player.

### Desired Outcome
Users can select from three bot difficulty levels (Easy, Medium, Hard) when starting a game, choose their color, and play a full chess match against an AI opponent that provides appropriate challenge based on the selected difficulty. The bot should respond within reasonable time limits and provide an engaging experience through personality touches like humorous "thinking" messages.

### Success Metrics
- Bot difficulty levels provide meaningful skill progression from beginner to advanced
- Easy bot can be beaten by novice players; Hard bot challenges experienced players
- Response times feel appropriate (not too fast to be jarring, not too slow to be frustrating)
- Users can complete full games against bots without confusion or technical issues

---

## 2. Functional Requirements (The "What")

### 2.1 Bot Selection and Game Setup

**As a** user, **I want to** select a bot opponent and difficulty level from the main menu, **so that** I can start a game against an AI opponent appropriate for my skill level.

**Acceptance Criteria:**
- [ ] When starting a new game, the main menu displays options: "1) Player vs Player, 2) vs Easy Bot, 3) vs Medium Bot, 4) vs Hard Bot"
- [ ] Selecting options 2, 3, or 4 starts a bot game at the chosen difficulty
- [ ] After selecting a bot difficulty, the user is prompted: "Play as: 1) White, 2) Black, 3) Random"
- [ ] Selecting "Random" assigns the user's color randomly (50/50 chance)
- [ ] Selecting White or Black assigns that color to the user and the opposite to the bot
- [ ] The game board then displays with the user and bot assigned to their respective colors
- [ ] If the bot plays White, it makes the first move immediately after setup

### 2.2 Bot Difficulty Levels

**As a** user, **I want** each bot difficulty to provide a distinct skill level, **so that** I can choose appropriate challenge and practice against progressively harder opponents.

**Acceptance Criteria:**

#### Easy Bot
- [ ] Makes legal moves using random selection or very simple heuristics
- [ ] Does not search deeply (1-2 moves ahead maximum)
- [ ] Response time: 1-2 seconds maximum
- [ ] Should be beatable by novice players who understand basic chess rules
- [ ] May miss obvious tactical opportunities (hanging pieces, simple forks)

#### Medium Bot
- [ ] Makes solid, reasonable moves using minimax or similar algorithm
- [ ] Evaluates position quality using basic heuristics (material count, piece positioning)
- [ ] Search depth: 3-4 moves ahead
- [ ] Response time: 3-4 seconds maximum
- [ ] Should challenge intermediate players but be beatable with good tactics
- [ ] Recognizes and executes basic tactical patterns (forks, pins, skewers)

#### Hard Bot
- [ ] Makes strong, strategic moves with deeper search
- [ ] Uses advanced position evaluation (pawn structure, king safety, piece activity)
- [ ] Search depth: 5+ moves ahead
- [ ] Response time: 5-8 seconds maximum
- [ ] Should challenge experienced players and require strong play to beat
- [ ] Finds complex tactical combinations and long-term strategic plans

### 2.3 Bot Thinking Feedback

**As a** user, **I want to** see entertaining feedback while the bot is calculating, **so that** I understand the game hasn't frozen and stay engaged during wait times.

**Acceptance Criteria:**
- [ ] When the bot is calculating its move, a "thinking" message is displayed to the user
- [ ] The system randomly selects from a pool of 10-15 chess-themed humorous messages
- [ ] Example messages include variations like:
  - "Calculating fork trajectories..."
  - "Consulting the chess gods..."
  - "Pondering pawn structures..."
  - "Analyzing knight maneuvers..."
  - "Contemplating castle formations..."
  - "Evaluating bishop diagonals..."
  - "Reviewing rook highways..."
  - "Meditating on the middle game..."
  - "Channeling chess grandmasters..."
  - "Summoning strategic insights..."
  - [NEEDS CLARIFICATION: Complete list of all 10-15 messages to be finalized during implementation]
- [ ] The message displays at the start of bot calculation and clears when the bot makes its move
- [ ] If the bot exceeds its time limit, it makes the best move found so far (no indefinite waiting)

### 2.4 Game End and Post-Game Flow

**As a** user, **I want** clear feedback when a bot game ends and options for what to do next, **so that** I can understand the outcome and quickly start another game if desired.

**Acceptance Criteria:**

#### Game Result Display
- [ ] When the game ends (checkmate, stalemate, resignation, or draw), display the result type clearly:
  - "Checkmate! You won!"
  - "Checkmate! Bot wins."
  - "Stalemate! Game is a draw."
  - "You resigned. Bot wins."
  - "Draw accepted."
- [ ] Display game statistics after the result:
  - Total moves played (e.g., "Game lasted 32 moves")
  - Game duration in minutes and seconds (e.g., "Game time: 12m 34s")
  - Result type confirmation (e.g., "Victory by checkmate!")
  - Final FEN position displayed and optionally saved

#### Post-Game Options
- [ ] After showing statistics, prompt: "Rematch? (y/n)"
- [ ] If user chooses 'y' (yes to rematch):
  - Start a new game with the same bot difficulty
  - Use the same color assignment (if user was White before, user is White again)
  - Clear the board and begin fresh game
- [ ] If user chooses 'n' (no rematch):
  - Return to the main menu to select a different game mode or exit

### 2.5 In-Game Actions

**As a** user, **I want to** resign or offer a draw during a bot game, **so that** I can end hopeless positions or propose draws in balanced games.

**Acceptance Criteria:**

#### Resignation
- [ ] User can type "resign" (case-insensitive) at any move prompt to resign the game
- [ ] System confirms: "Are you sure you want to resign? (y/n)"
- [ ] If confirmed, game ends with "You resigned. Bot wins." and proceeds to game end flow
- [ ] If not confirmed, user returns to move input

#### Draw Offers
- [ ] User can type "draw" (case-insensitive) at any move prompt to offer a draw
- [ ] Bot evaluates the current position using its evaluation function
- [ ] Bot accepts draw if:
  - Position evaluation is between -0.5 and +0.5 (approximately even)
  - OR bot is in a losing position (evaluation worse than -1.5 from bot's perspective)
- [ ] Bot declines draw if position evaluation is clearly in its favor (> +1.0 from bot's perspective)
- [ ] Display bot's response: "Bot accepts the draw." or "Bot declines the draw. Continue playing."
- [ ] If accepted, game ends with "Draw accepted." and proceeds to game end flow
- [ ] If declined, user continues making their move

---

## 3. Scope and Boundaries

### In-Scope
- Three bot difficulty levels: Easy, Medium, Hard
- Bot selection via main menu game mode options
- User color selection (White, Black, or Random) for each bot game
- Bot move calculation with appropriate time limits per difficulty
- Humorous chess-themed "thinking" messages (10-15 variations)
- Game end detection and result display for bot games
- Post-game statistics: move count, duration, result type, final FEN
- Rematch functionality with same settings (difficulty and color)
- User resignation during bot games
- User draw offers with position-based bot acceptance logic
- Integration with existing chess engine, FEN support, and terminal interface

### Out-of-Scope
- Custom RL Agent (separate Phase 5 roadmap item)
- UCI Engine Integration (separate Phase 5 roadmap item)
- Opening book databases or engine configuration
- Difficulty levels beyond Easy, Medium, Hard
- Bot personality customization or playing style selection
- Bot vs Bot mode (two bots playing against each other)
- Adjustable time controls or configurable bot time limits
- Analysis mode or move hints from bots
- Bot strength ratings (ELO) or formal skill measurement
- Saving/loading bot game progress mid-game (beyond FEN export)
- Online or networked bot play
- Multiple bots playing simultaneously
