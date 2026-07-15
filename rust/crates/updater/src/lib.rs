//! In-place binary upgrade via GitHub Releases. Powers `termchess --upgrade`.
//!
//! Port of the Go `internal/updater` package, preserving behavior and public
//! semantics.
//!
//! ## Flow
//!
//! 1. Query the GitHub API for the latest (or a specified) release tag.
//! 2. Download the platform-matched asset (darwin/linux × amd64/arm64).
//! 3. Verify the SHA-256 checksum against `checksums.txt`.
//! 4. Atomically replace the running binary (rename-swap via [`replace_binary`]).

mod checksum;
mod client;
mod context;
mod error;
mod install;
mod platform;
mod urls;
mod version_cmp;

/// GitHub repository owner.
pub(crate) const REPO_OWNER: &str = "Mgrdich";
/// GitHub repository name.
pub(crate) const REPO_NAME: &str = "TermChess";
/// Base URL for the GitHub API.
pub(crate) const GITHUB_API: &str = "https://api.github.com";

pub use checksum::{get_expected_checksum, parse_checksums, verify_checksum};
pub use client::{Client, UpgradeResult};
pub use context::{CancelHandle, Context};
pub use error::UpdaterError;
pub use install::{
    detect_install_method, get_go_install_message, replace_binary, uninstall, InstallMethod,
};
pub use platform::{current_goarch, current_goos};
pub use urls::{get_asset_url, get_binary_filename, get_checksums_url};
pub use version_cmp::compare_versions;
