//! Build-time version metadata.
//!
//! Mirrors Go's `internal/version/version.go`, where three package-level
//! string variables (`Version`, `BuildDate`, `GitCommit`) are set via
//! `ldflags` at build time and default to placeholder values for local
//! builds.
//!
//! In Rust the equivalent of Go's `-ldflags -X` injection is reading
//! environment variables at compile time via [`option_env!`]. The build
//! tooling (e.g. the `Makefile`) sets these env vars before invoking
//! `cargo build`; when unset, the same defaults as the Go package apply.

/// The build version.
///
/// Injected from the `TERMCHESS_VERSION` env var at compile time.
/// Defaults to `"dev"` for local builds (matching Go's default).
pub const VERSION: &str = match option_env!("TERMCHESS_VERSION") {
    Some(v) => v,
    None => "dev",
};

/// The build date (ISO 8601 UTC timestamp).
///
/// Injected from the `TERMCHESS_BUILD_DATE` env var at compile time.
/// Defaults to `"unknown"` when unset (matching Go's default).
pub const BUILD_DATE: &str = match option_env!("TERMCHESS_BUILD_DATE") {
    Some(v) => v,
    None => "unknown",
};

/// The short git commit SHA.
///
/// Injected from the `TERMCHESS_GIT_COMMIT` env var at compile time.
/// Defaults to `"unknown"` when unset (matching Go's default).
pub const GIT_COMMIT: &str = match option_env!("TERMCHESS_GIT_COMMIT") {
    Some(v) => v,
    None => "unknown",
};

#[cfg(test)]
mod tests {
    use super::*;

    // These mirror the Go package's documented defaults for local builds
    // (built without ldflags injection). The crate's own test build does
    // not set the injection env vars, so the defaults must hold.

    #[test]
    fn version_defaults_to_dev() {
        assert_eq!(VERSION, "dev");
    }

    #[test]
    fn build_date_defaults_to_unknown() {
        assert_eq!(BUILD_DATE, "unknown");
    }

    #[test]
    fn git_commit_defaults_to_unknown() {
        assert_eq!(GIT_COMMIT, "unknown");
    }

    #[test]
    fn values_are_non_empty() {
        assert!(!VERSION.is_empty());
        assert!(!BUILD_DATE.is_empty());
        assert!(!GIT_COMMIT.is_empty());
    }
}
