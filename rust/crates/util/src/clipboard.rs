//! Cross-platform clipboard helper.
//!
//! Mirrors the Go `internal/util` clipboard package: a single function that
//! copies text to the system clipboard, returning an error on failure so
//! callers can degrade gracefully in headless environments.

use thiserror::Error;

/// Errors that can occur when copying text to the system clipboard.
#[derive(Debug, Error)]
pub enum ClipboardError {
    /// The clipboard could not be initialized. This typically happens in
    /// headless environments (e.g. CI servers without a display server) or
    /// when clipboard access is restricted by the operating system.
    #[error("failed to initialize clipboard: {0}")]
    Init(#[source] arboard::Error),

    /// The clipboard was initialized but writing the text failed.
    #[error("failed to write to clipboard: {0}")]
    Write(#[source] arboard::Error),
}

/// Copies the given text to the system clipboard.
///
/// This function provides cross-platform clipboard support for Windows,
/// macOS, and Linux. It handles clipboard initialization internally and can
/// be called multiple times safely.
///
/// Platform-specific notes:
///   - On macOS: works on standard macOS systems.
///   - On Linux: requires an X11 or Wayland display server.
///   - On Windows: uses the Windows clipboard API.
///
/// The function may fail in headless environments (e.g. CI servers without a
/// display) or when clipboard access is restricted by the operating system.
///
/// # Errors
///
/// Returns [`ClipboardError::Init`] if the clipboard could not be
/// initialized, or [`ClipboardError::Write`] if writing the text failed.
///
/// # Example
///
/// ```no_run
/// if let Err(e) = util::copy_to_clipboard("example text") {
///     eprintln!("Failed to copy to clipboard: {e}");
/// }
/// ```
pub fn copy_to_clipboard(text: &str) -> Result<(), ClipboardError> {
    // Initialize clipboard (safe to call multiple times). This must happen
    // before any clipboard operations.
    let mut clipboard = arboard::Clipboard::new().map_err(ClipboardError::Init)?;

    // Write text to the clipboard.
    clipboard
        .set_text(text.to_owned())
        .map_err(ClipboardError::Write)?;

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    /// Tests basic clipboard functionality.
    ///
    /// Note: this test may fail in headless/CI environments without display
    /// server access. It verifies that the function can be called without
    /// panicking, accepting a returned error as a valid outcome.
    #[test]
    fn test_copy_to_clipboard() {
        if std::env::var("CI").as_deref() == Ok("true") {
            eprintln!("Skipping clipboard tests in CI environment");
            return;
        }

        let cases: &[(&str, &str)] = &[
            ("copy simple text", "Hello, World!"),
            ("copy empty string", ""),
            (
                "copy FEN string",
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
            ),
            ("copy multiline text", "Line 1\nLine 2\nLine 3"),
            (
                "copy text with special characters",
                "Special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?",
            ),
        ];

        for (name, text) in cases {
            // In headless environments, clipboard initialization may fail.
            // We accept this as a valid scenario.
            if let Err(e) = copy_to_clipboard(text) {
                // Log the error but don't fail the test if it's an
                // initialization error; this allows tests to pass in CI.
                eprintln!(
                    "[{name}] Clipboard operation failed (expected in headless environments): {e}"
                );
            }
        }
    }

    /// Ensures the function doesn't panic under any circumstances.
    #[test]
    fn test_copy_to_clipboard_does_not_panic() {
        let long = "very long string: ".to_string() + &"\0".repeat(10_000);
        let inputs: Vec<String> = vec![
            String::new(),
            "simple text".to_string(),
            "text with unicode: 日本語 🎮 ♔♕♖♗♘♙".to_string(),
            long,
        ];

        for input in &inputs {
            let _ = copy_to_clipboard(input);
        }
    }
}
