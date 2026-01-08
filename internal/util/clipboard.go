package util

import (
	"fmt"
	"golang.design/x/clipboard"
)

// CopyToClipboard copies the given text to the system clipboard.
//
// This function provides cross-platform clipboard support for Windows, macOS, and Linux.
// It handles clipboard initialization internally and can be called multiple times safely.
//
// Platform-specific notes:
//   - On macOS: Requires Cocoa framework (works on standard macOS systems)
//   - On Linux: Requires X11 or Wayland display server
//   - On Windows: Uses the Windows clipboard API
//
// The function may fail in headless environments (e.g., CI servers without display)
// or when clipboard access is restricted by the operating system.
//
// Parameters:
//   - text: The string to copy to the clipboard
//
// Returns:
//   - error: nil on success, or an error if clipboard initialization or write fails
//
// Example:
//
//	err := CopyToClipboard("example text")
//	if err != nil {
//	    log.Printf("Failed to copy to clipboard: %v", err)
//	}
func CopyToClipboard(text string) error {
	// Initialize clipboard (safe to call multiple times)
	// This must happen before any clipboard operations
	err := clipboard.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize clipboard: %w", err)
	}

	// Write text to clipboard
	// Note: clipboard.Write does not return an error, but initialization
	// failure would have been caught above
	clipboard.Write(clipboard.FmtText, []byte(text))

	return nil
}
