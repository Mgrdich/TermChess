// Package main is the entry point for the TermChess application.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Mgrdich/TermChess/internal/ui"
)

func main() {
	// Initialize the Bubbletea model
	model := ui.NewModel()

	// Create the Bubbletea program with options:
	// - WithAltScreen: Use alternate screen buffer for clean TUI experience
	// - WithMouseCellMotion: Enable mouse support for future interactions
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Future: mouse support
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
