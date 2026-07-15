//! Bot engines for TermChess: random (Easy), minimax with alpha-beta (Medium
//! and Hard), and an RL/ONNX skeleton.
//!
//! Ported from the Go `internal/bot` package, preserving behavior and public
//! semantics. Depends only on the `engine` crate.
//!
//! Go interfaces map to traits ([`Engine`], [`Configurable`], [`Stateful`],
//! [`Inspectable`]); Go's `context.Context` maps to [`Context`]; Go errors map
//! to the [`EngineError`] enum.

#[cfg(test)]
mod tests;

mod context;
mod error;
mod eval;
mod factory;
mod interfaces;
mod minimax;
mod random;
mod rl;
mod rl_encoder;

pub use context::Context;
pub use error::EngineError;
pub use factory::{
    new_minimax_engine, new_random_engine, with_deterministic, with_options, with_search_depth,
    with_time_limit, EngineConfig, EngineOption, OptionValue,
};
pub use interfaces::{
    Configurable, Difficulty, Engine, EngineType, Info, Inspectable, MinimaxConfig, Stateful,
};
pub use minimax::MinimaxEngine;
pub use random::RandomEngine;
pub use rl::{new_rl_engine, InferenceSession, RLDifficulty, RlEngine};
pub use rl_encoder::encode_board;
