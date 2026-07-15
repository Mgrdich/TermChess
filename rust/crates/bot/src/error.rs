//! Error type for the bot crate.
//!
//! Go returns `error` values built with `errors.New` / `fmt.Errorf`; these are
//! mapped to a single [`EngineError`] enum. The `Display` strings are kept
//! byte-identical to the Go messages the tests assert on.

use thiserror::Error;

/// Errors returned by bot engines and their factory functions.
#[derive(Debug, Error, Clone, PartialEq, Eq)]
pub enum EngineError {
    /// The engine has been closed and can no longer select moves.
    #[error("engine is closed")]
    Closed,

    /// `Close` was called on an engine that was already closed.
    #[error("engine already closed")]
    AlreadyClosed,

    /// The position has no legal moves (checkmate or stalemate).
    #[error("no legal moves available")]
    NoLegalMovesAvailable,

    /// `select_best_move` was called with an empty legal-move list.
    #[error("no legal moves")]
    NoLegalMoves,

    /// The RL model has not been loaded (ONNX runtime integration pending).
    #[error("RL model not loaded: ONNX runtime integration pending")]
    ModelNotLoaded,

    /// Inference through the RL session failed; wraps the session error text.
    #[error("inference failed: {0}")]
    InferenceFailed(String),

    /// A configuration or factory validation error (free-form message).
    #[error("{0}")]
    Config(String),

    /// The policy logits vector had an unexpected length.
    #[error("invalid policy logits length: got {0}, want {1}")]
    InvalidPolicyLength(usize, usize),

    /// The context was canceled before a move could be selected.
    #[error("context canceled")]
    ContextCanceled,

    /// The context deadline was exceeded before a move could be selected.
    #[error("context deadline exceeded")]
    ContextDeadlineExceeded,
}
