// Package main is the entry point for the TermChess application.
package main

import (
	"fmt"
	"os"

	"github.com/Mgrdich/TermChess/internal/config"
	"github.com/Mgrdich/TermChess/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Load configuration from ~/.termchess/config.toml
	// If the file doesn't exist or cannot be parsed, default values are used
	cfg := config.LoadConfig()

	// Initialize the Bubbletea model with the loaded configuration
	model := ui.NewModel(cfg)

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
