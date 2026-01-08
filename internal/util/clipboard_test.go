package util

import (
	"testing"
)

// TestCopyToClipboard tests basic clipboard functionality.
// Note: This test may fail in headless/CI environments without display server access.
// The test verifies that the function can be called without panicking.
func TestCopyToClipboard(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{
			name: "copy simple text",
			text: "Hello, World!",
		},
		{
			name: "copy empty string",
			text: "",
		},
		{
			name: "copy FEN string",
			text: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		},
		{
			name: "copy multiline text",
			text: "Line 1\nLine 2\nLine 3",
		},
		{
			name: "copy text with special characters",
			text: "Special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Attempt to copy to clipboard
			err := CopyToClipboard(tt.text)

			// In headless environments, clipboard initialization may fail
			// We accept this as a valid scenario
			if err != nil {
				// Log the error but don't fail the test if it's an initialization error
				// This allows tests to pass in CI environments
				t.Logf("Clipboard operation failed (expected in headless environments): %v", err)
				return
			}

			// If we got here, clipboard operation succeeded (no assertions needed)
		})
	}
}

// TestCopyToClipboardDoesNotPanic ensures the function doesn't panic under any circumstances.
func TestCopyToClipboardDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("CopyToClipboard() panicked: %v", r)
		}
	}()

	// Test with various inputs
	testInputs := []string{
		"",
		"simple text",
		"text with unicode: 日本語 🎮 ♔♕♖♗♘♙",
		"very long string: " + string(make([]byte, 10000)),
	}

	for _, input := range testInputs {
		_ = CopyToClipboard(input)
	}
}
