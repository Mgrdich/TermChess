// Package main is the entry point for the TermChess application.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Mgrdich/TermChess/internal/config"
	"github.com/Mgrdich/TermChess/internal/ui"
	"github.com/Mgrdich/TermChess/internal/version"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Parse command-line flags first
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Handle --version flag (exit before TUI)
	if *showVersion {
		printVersion()
		return
	}

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

// printVersion prints the version information and exits.
func printVersion() {
	fmt.Printf("termchess %s\n", version.Version)
	fmt.Printf("Build date: %s\n", version.BuildDate)
	fmt.Printf("Git commit: %s\n", version.GitCommit)
}
