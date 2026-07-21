//! Engine traits and shared metadata types (ported from `engine.go`).

use std::collections::HashMap;
use std::time::Duration;

use engine::{Board, Move};

use crate::context::Context;
use crate::error::EngineError;

/// A chess bot that can select moves.
///
/// This is the minimal trait all engines must implement (Go's `Engine`
/// interface). `select_move` takes `&self`; engines use interior mutability
/// (atomics) for their `closed` state so callers can share a bot behind a
/// shared reference, just as Go shares the interface value.
pub trait Engine {
    /// Returns the bot's chosen move for the given position.
    ///
    /// The context allows cancellation if the bot exceeds time limits.
    fn select_move(&self, ctx: &Context, board: &Board) -> Result<Move, EngineError>;

    /// Returns a human-readable name for this engine.
    fn name(&self) -> &str;

    /// Releases any resources held by the engine.
    ///
    /// Implementations should be idempotent where the Go original is; the RL
    /// engine deliberately returns an error on a second close.
    fn close(&self) -> Result<(), EngineError>;
}

/// Configuration options for minimax engines. All fields are optional (`None`
/// means "not set"), matching the pointer fields of Go's `MinimaxConfig`.
#[derive(Debug, Default, Clone, PartialEq)]
pub struct MinimaxConfig {
    /// Search depth (1-20).
    pub search_depth: Option<i32>,
    /// Time limit per move (must be positive).
    pub time_limit: Option<Duration>,
    /// Weight for material evaluation.
    pub material_weight: Option<f64>,
    /// Weight for piece-square table evaluation.
    pub piece_square_weight: Option<f64>,
    /// Weight for mobility evaluation.
    pub mobility_weight: Option<f64>,
    /// Weight for king safety evaluation.
    pub king_safety_weight: Option<f64>,
}

/// Engines that can accept configuration before or during use.
pub trait Configurable: Engine {
    /// Applies the given configuration, validating individual fields.
    fn configure(&mut self, config: MinimaxConfig) -> Result<(), EngineError>;
}

/// Engines that benefit from knowing position history.
pub trait Stateful: Engine {
    /// Provides the engine with the game's position history.
    fn set_position_history(&mut self, history: Vec<Board>) -> Result<(), EngineError>;
}

/// Metadata about an engine.
#[derive(Debug, Clone, PartialEq)]
pub struct Info {
    /// Human-readable name.
    pub name: String,
    /// Engine author.
    pub author: String,
    /// Engine version.
    pub version: String,
    /// Internal, UCI, or RL.
    pub engine_type: EngineType,
    /// Easy, Medium, Hard (for internal bots).
    pub difficulty: Difficulty,
    /// Supported features.
    pub features: HashMap<String, bool>,
}

/// Engines that can report metadata (Go's `Inspectable`).
pub trait Inspectable: Engine {
    /// Returns metadata about the engine.
    fn info(&self) -> Info;
}

/// Categorizes engine implementations.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum EngineType {
    /// Built-in Rust implementations.
    Internal,
    /// External UCI engines (Phase 5).
    Uci,
    /// RL agents with ONNX models (Phase 6).
    Rl,
}

impl std::fmt::Display for EngineType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let s = match self {
            EngineType::Internal => "Internal",
            EngineType::Uci => "UCI",
            EngineType::Rl => "RL",
        };
        write!(f, "{}", s)
    }
}

/// Difficulty levels for internal engines. Discriminants match the Go `iota`
/// values (Easy = 0, Medium = 1, Hard = 2), and the ordering is used by the
/// evaluation via `difficulty >= Medium` style comparisons.
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord)]
pub enum Difficulty {
    /// Fast responses, simpler evaluation.
    Easy = 0,
    /// Balanced play.
    Medium = 1,
    /// Stronger evaluation, deeper search.
    Hard = 2,
}

impl std::fmt::Display for Difficulty {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let s = match self {
            Difficulty::Easy => "Easy",
            Difficulty::Medium => "Medium",
            Difficulty::Hard => "Hard",
        };
        write!(f, "{}", s)
    }
}
