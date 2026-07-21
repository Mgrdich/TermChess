//! Error type for the updater crate.
//!
//! Mirrors the Go package's sentinel errors (`ErrAlreadyUpToDate`,
//! `ErrChecksumMismatch`, `ErrPermissionDenied`) plus the wrapped
//! `fmt.Errorf` messages, collapsed into a single `thiserror` enum.

use thiserror::Error;

/// Errors returned by updater operations.
#[derive(Debug, Error)]
pub enum UpdaterError {
    /// The current version already matches the target version.
    ///
    /// Equivalent to Go's `ErrAlreadyUpToDate`.
    #[error("already up to date")]
    AlreadyUpToDate,

    /// The downloaded binary's checksum did not match the expected value.
    ///
    /// Equivalent to Go's `ErrChecksumMismatch`.
    #[error("checksum mismatch")]
    ChecksumMismatch,

    /// An operation failed because of insufficient filesystem permissions.
    ///
    /// Equivalent to Go's `ErrPermissionDenied`.
    #[error("permission denied")]
    PermissionDenied,

    /// The user declined a downgrade when prompted.
    #[error("downgrade cancelled by user")]
    DowngradeCancelled,

    /// The server returned a non-200 status code.
    #[error("unexpected status code: {0}")]
    UnexpectedStatusCode(u16),

    /// The GitHub release response had an empty `tag_name`.
    #[error("empty tag_name in response")]
    EmptyTagName,

    /// No checksum entry was found for the requested platform binary.
    #[error("checksum not found for {0}")]
    ChecksumNotFound(String),

    /// The request was cancelled before completing.
    #[error("request cancelled")]
    Cancelled,

    /// The request exceeded its deadline.
    #[error("request timed out")]
    Timeout,

    /// The response body could not be parsed.
    #[error("parsing response: {0}")]
    Parse(String),

    /// A transport/network-level error occurred.
    #[error("{0}")]
    Network(String),

    /// A wrapped error carrying a contextual message (mirrors `fmt.Errorf`).
    #[error("{0}")]
    Message(String),

    /// An underlying I/O error.
    #[error("io error: {0}")]
    Io(#[from] std::io::Error),

    /// An error resolving the configuration directory.
    #[error("config error: {0}")]
    Config(#[from] config::ConfigError),
}
