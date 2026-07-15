//! Build-time injected version metadata.
//!
//! No internal dependencies. Port of Go's `internal/version` package.

mod version;

pub use version::{BUILD_DATE, GIT_COMMIT, VERSION};
