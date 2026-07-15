//! Cross-platform utilities (clipboard helper).
//!
//! Ported from the Go `internal/util` package. No internal dependencies.

mod clipboard;

pub use clipboard::{copy_to_clipboard, ClipboardError};
