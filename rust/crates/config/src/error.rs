//! Error types for the config crate.

use thiserror::Error;

use engine::FenError;

/// Errors returned by config and savegame operations.
#[derive(Debug, Error)]
pub enum ConfigError {
    /// The user's home directory could not be determined.
    #[error("failed to get home directory")]
    HomeDirUnavailable,

    /// An I/O error occurred while reading or writing a file.
    #[error("io error: {0}")]
    Io(#[from] std::io::Error),

    /// The config file could not be encoded to TOML.
    #[error("failed to encode config to TOML: {0}")]
    TomlEncode(#[from] toml::ser::Error),

    /// A saved game's FEN string could not be parsed.
    #[error("failed to parse saved game FEN: {0}")]
    Fen(#[from] FenError),
}
