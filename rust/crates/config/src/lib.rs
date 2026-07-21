//! Configuration and game state persistence for TermChess.
//!
//! Configuration files are stored in `~/.termchess/` and use TOML format.
//! Game saves are stored as FEN strings in `~/.termchess/savegame.fen`.
//!
//! The crate provides:
//!   - Config types and default values
//!   - Config file loading and saving
//!   - Game state save/load/delete operations
//!   - Path helpers for the config directory and files
//!
//! Ported from the Go `internal/config` package, preserving behavior and public
//! semantics.

mod config;
mod error;
mod paths;
mod savegame;

pub use config::{
    load_config, load_game_config, save_config, Config, ConfigFile, DisplayConfig, GameConfig,
    DEFAULT_THEME,
};
pub use error::ConfigError;
pub use paths::{get_config_dir, get_config_path, save_game_path};
pub use savegame::{delete_save_game, load_game, save_game, save_game_exists};

/// A process-wide lock serializing tests that touch the shared `~/.termchess/`
/// files, mirroring Go's sequential in-package test execution.
#[cfg(test)]
pub(crate) fn fs_test_lock() -> std::sync::MutexGuard<'static, ()> {
    use std::sync::{Mutex, OnceLock};
    static LOCK: OnceLock<Mutex<()>> = OnceLock::new();
    LOCK.get_or_init(|| Mutex::new(()))
        .lock()
        .unwrap_or_else(|poisoned| poisoned.into_inner())
}
