//! Path helpers for the TermChess configuration directory and files.
//!
//! Configuration lives under `~/.termchess/`:
//!   - `config.toml` — TOML settings
//!   - `savegame.fen` — FEN game save

use std::path::PathBuf;

use crate::error::ConfigError;

/// Returns the path to the TermChess configuration directory (`~/.termchess/`).
///
/// Returns an error if the home directory cannot be determined.
pub fn get_config_dir() -> Result<PathBuf, ConfigError> {
    let home = home_dir()?;
    Ok(home.join(".termchess"))
}

/// Returns the full path to the configuration file (`~/.termchess/config.toml`).
pub(crate) fn get_config_file_path() -> Result<PathBuf, ConfigError> {
    Ok(get_config_dir()?.join("config.toml"))
}

/// Returns the full path to the save game file (`~/.termchess/savegame.fen`).
///
/// Exported for testing purposes, mirroring the Go `SaveGamePath`.
pub fn save_game_path() -> Result<PathBuf, ConfigError> {
    Ok(get_config_dir()?.join("savegame.fen"))
}

/// Returns the absolute path to the configuration file (`~/.termchess/config.toml`).
pub fn get_config_path() -> Result<PathBuf, ConfigError> {
    let home = home_dir()?;
    Ok(home.join(".termchess").join("config.toml"))
}

/// Resolves the user's home directory, mirroring Go's `os.UserHomeDir`.
fn home_dir() -> Result<PathBuf, ConfigError> {
    dirs::home_dir().ok_or(ConfigError::HomeDirUnavailable)
}
