//! Core value types for Bot vs Bot sessions (ported from `types.go`).

use std::time::Duration;

use engine::{Color, Move};

/// Controls the delay between moves in a Bot vs Bot game.
///
/// Go's `PlaybackSpeed` was an `int` with `iota` constants; here it is an
/// idiomatic enum. Any value maps to a delay via [`PlaybackSpeed::duration`].
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum PlaybackSpeed {
    /// Applies no delay between moves.
    Instant,
    /// Applies a 1 second delay between moves.
    Normal,
}

impl PlaybackSpeed {
    /// Returns the time delay associated with this playback speed.
    pub fn duration(self) -> Duration {
        match self {
            PlaybackSpeed::Instant => Duration::ZERO,
            PlaybackSpeed::Normal => Duration::from_secs(1),
        }
    }
}

/// Represents the current state of a game session.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum SessionState {
    /// The game is currently in progress.
    Running,
    /// The game is paused.
    Paused,
    /// The game has completed.
    Finished,
}

/// Holds the outcome of a completed Bot vs Bot game.
#[derive(Debug, Clone)]
pub struct GameResult {
    /// The sequence number of this game in a series.
    pub game_number: i32,
    /// The name of the winning engine, or "Draw" for draws.
    pub winner: String,
    /// The color of the winning engine.
    pub winner_color: Color,
    /// Describes why the game ended (e.g., "checkmate", "stalemate").
    pub end_reason: String,
    /// The total number of moves played.
    pub move_count: i32,
    /// How long the game took to complete.
    pub duration: Duration,
    /// The FEN string of the final board position.
    pub final_fen: String,
    /// All moves played in order.
    pub move_history: Vec<Move>,
}

impl Default for GameResult {
    fn default() -> Self {
        GameResult {
            game_number: 0,
            winner: String::new(),
            // Go's zero value for engine.Color is White (0).
            winner_color: Color::White,
            end_reason: String::new(),
            move_count: 0,
            duration: Duration::ZERO,
            final_fen: String::new(),
            move_history: Vec::new(),
        }
    }
}
