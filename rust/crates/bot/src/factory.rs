//! Engine construction and functional options (ported from `factory.go`).

use std::collections::HashMap;
use std::time::Duration;

use crate::error::EngineError;
use crate::interfaces::Difficulty;
use crate::minimax::{get_default_weights, MinimaxEngine};
use crate::random::RandomEngine;

/// A value stored in the free-form options map (`map[string]any` in Go).
#[derive(Debug, Clone, PartialEq)]
pub enum OptionValue {
    /// An integer option value.
    Int(i64),
    /// A boolean option value.
    Bool(bool),
    /// A string option value.
    Str(String),
}

/// Configuration accumulated by [`EngineOption`]s during engine creation.
///
/// The struct is public so options can be passed to the factory functions, but
/// its fields are crate-private (matching Go's unexported `engineConfig`).
#[derive(Debug, Clone, PartialEq)]
pub struct EngineConfig {
    pub(crate) difficulty: Difficulty,
    pub(crate) time_limit: Duration,
    pub(crate) search_depth: i32,
    pub(crate) deterministic: bool,
    pub(crate) options: Option<HashMap<String, OptionValue>>,
}

impl Default for EngineConfig {
    fn default() -> Self {
        EngineConfig {
            difficulty: Difficulty::Easy,
            time_limit: Duration::ZERO,
            search_depth: 0,
            deterministic: false,
            options: None,
        }
    }
}

/// A functional option applied to an [`EngineConfig`] during engine creation.
pub type EngineOption = Box<dyn Fn(&mut EngineConfig) -> Result<(), EngineError>>;

/// Sets a custom time limit for move selection. Must be positive.
///
/// Note: Rust's `Duration` is unsigned, so only a zero limit is invalid here
/// (Go additionally rejected negative durations).
pub fn with_time_limit(d: Duration) -> EngineOption {
    Box::new(move |c| {
        if d.is_zero() {
            return Err(EngineError::Config(
                "time limit must be positive".to_string(),
            ));
        }
        c.time_limit = d;
        Ok(())
    })
}

/// Sets a custom search depth for minimax engines (1-20).
pub fn with_search_depth(depth: i32) -> EngineOption {
    Box::new(move |c| {
        if !(1..=20).contains(&depth) {
            return Err(EngineError::Config("search depth must be 1-20".to_string()));
        }
        c.search_depth = depth;
        Ok(())
    })
}

/// Sets custom options as a map.
pub fn with_options(opts: HashMap<String, OptionValue>) -> EngineOption {
    Box::new(move |c| {
        c.options = Some(opts.clone());
        Ok(())
    })
}

/// Disables random tie-breaking for reproducible results (used in tests).
pub fn with_deterministic(deterministic: bool) -> EngineOption {
    Box::new(move |c| {
        c.deterministic = deterministic;
        Ok(())
    })
}

/// Creates an Easy bot with weighted random move selection.
pub fn new_random_engine(opts: &[EngineOption]) -> Result<RandomEngine, EngineError> {
    let mut cfg = EngineConfig {
        difficulty: Difficulty::Easy,
        time_limit: Duration::from_secs(2),
        ..EngineConfig::default()
    };

    for opt in opts {
        opt(&mut cfg)?;
    }

    Ok(RandomEngine::new("Easy Bot", cfg.time_limit))
}

/// Creates a Medium or Hard bot using minimax with alpha-beta pruning.
pub fn new_minimax_engine(
    difficulty: Difficulty,
    opts: &[EngineOption],
) -> Result<MinimaxEngine, EngineError> {
    let mut cfg = EngineConfig {
        difficulty,
        ..EngineConfig::default()
    };

    match difficulty {
        Difficulty::Medium => {
            cfg.time_limit = Duration::from_secs(4);
            cfg.search_depth = 4;
        }
        Difficulty::Hard => {
            cfg.time_limit = Duration::from_secs(8);
            cfg.search_depth = 7;
        }
        _ => {
            return Err(EngineError::Config(format!(
                "invalid difficulty for minimax: {} (expected Medium or Hard)",
                difficulty as i32
            )));
        }
    }

    for opt in opts {
        opt(&mut cfg)?;
    }

    let name = format!("{} Bot", difficulty);

    Ok(MinimaxEngine::new(
        name,
        cfg.difficulty,
        cfg.search_depth,
        cfg.time_limit,
        get_default_weights(cfg.difficulty),
        cfg.deterministic,
    ))
}
