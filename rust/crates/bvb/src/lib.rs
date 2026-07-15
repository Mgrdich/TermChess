//! Bot-vs-bot session management for spectator mode.
//!
//! Ported from the Go `internal/bvb` package, preserving behavior and public
//! semantics. Depends on the `bot` and `engine` crates.
//!
//! - [`GameSession`] — single-game controller: orchestrates two [`bot::Engine`]
//!   implementations taking turns, records moves, detects game end.
//! - [`SessionManager`] — multi-game queue: runs up to 50 games concurrently,
//!   collects results.
//! - [`compute_stats`] / [`AggregateStats`] — win-rate, average length,
//!   decisive/draw breakdown.
//! - [`SessionExport`] / [`save_session_export`] — JSON export of completed
//!   games.

mod export;
mod manager;
mod session;
mod stats;
mod types;

pub use export::{save_session_export, ExportError, GameExport, SessionExport};
pub use manager::{calculate_default_concurrency, max_concurrent_games, SessionManager};
pub use session::{GameSession, SharedEngine};
pub use stats::{compute_stats, AggregateStats};
pub use types::{GameResult, PlaybackSpeed, SessionState};
