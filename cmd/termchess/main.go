// Package main is the entry point for the TermChess application.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Mgrdich/TermChess/internal/config"
	"github.com/Mgrdich/TermChess/internal/ui"
	"github.com/Mgrdich/TermChess/internal/updater"
	"github.com/Mgrdich/TermChess/internal/version"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Parse command-line flags first
	showVersion := flag.Bool("version", false, "Show version information")
	doUpgrade := flag.Bool("upgrade", false, "Upgrade to latest version (or specify version as argument)")
	doUninstall := flag.Bool("uninstall", false, "Uninstall TermChess (remove binary and config)")
	flag.Parse()

	// Handle --version flag (exit before TUI)
	if *showVersion {
		printVersion()
		return
	}

	// Handle --upgrade flag
	if *doUpgrade {
		os.Exit(handleUpgrade(flag.Args()))
	}

	// Handle --uninstall flag
	if *doUninstall {
		os.Exit(handleUninstall())
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

// handleUpgrade handles the --upgrade flag.
// It returns the exit code (0 for success, 1 for error).
func handleUpgrade(args []string) int {
	// Check if installed via go install
	if updater.DetectInstallMethod() == updater.InstallMethodGoInstall {
		fmt.Println(updater.GetGoInstallMessage())
		return 0
	}

	// Get target version from args (if provided)
	var targetVersion string
	if len(args) > 0 {
		targetVersion = args[0]
	}

	client := updater.NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	currentVersion := version.Version

	// If no target version specified, check the latest
	if targetVersion == "" {
		fmt.Printf("Current version: %s\n", currentVersion)
		fmt.Print("Checking for updates...")

		latest, err := client.CheckLatestVersion(ctx)
		if err != nil {
			fmt.Printf("\nError: Failed to check for updates: %v\n", err)
			return 1
		}
		targetVersion = latest
		fmt.Printf("\rLatest version:  %s\n\n", targetVersion)
	} else {
		fmt.Printf("Current version: %s\n", currentVersion)
		fmt.Printf("Target version:  %s\n\n", targetVersion)
	}

	// Create confirmation callback for downgrades
	confirmDowngrade := func() bool {
		fmt.Print("\u26a0 " + targetVersion + " is older than your current version. It might be buggier than a summer porch. Continue? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return false
		}
		response = strings.TrimSpace(strings.ToLower(response))
		return response == "y" || response == "yes"
	}

	// Perform the upgrade
	binaryName := updater.GetBinaryFilename(targetVersion, runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Downloading %s...\n", binaryName)

	result, err := client.Upgrade(ctx, currentVersion, targetVersion, confirmDowngrade)
	if err != nil {
		if errors.Is(err, updater.ErrAlreadyUpToDate) {
			fmt.Printf("Already up to date (%s)\n", currentVersion)
			return 0
		}
		if errors.Is(err, updater.ErrPermissionDenied) {
			fmt.Println("Error: Permission denied. Try running with sudo:")
			fmt.Println("  sudo termchess --upgrade")
			return 1
		}
		if errors.Is(err, updater.ErrChecksumMismatch) {
			fmt.Println("Error: Checksum verification failed. The download may be corrupted.")
			return 1
		}
		if strings.Contains(err.Error(), "cancelled by user") {
			fmt.Println("Upgrade cancelled.")
			return 0
		}
		fmt.Printf("Error: %v\n", err)
		return 1
	}

	fmt.Print("Verifying checksum... \u2713\n")
	fmt.Print("Installing... \u2713\n\n")

	if result.IsDowngrade {
		fmt.Printf("\u2713 TermChess switched from %s to %s\n", result.PreviousVersion, result.NewVersion)
	} else {
		fmt.Printf("\u2713 TermChess upgraded from %s to %s\n", result.PreviousVersion, result.NewVersion)
	}

	return 0
}

// handleUninstall handles the --uninstall flag.
// It returns the exit code (0 for success, 1 for error).
func handleUninstall() int {
	// Prompt for confirmation
	fmt.Print("Are you sure you want to uninstall TermChess? [y/N] ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("\nError reading input: %v\n", err)
		return 1
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("\nUninstall cancelled.")
		return 0
	}

	fmt.Println()

	// Perform uninstall
	if err := updater.Uninstall(); err != nil {
		if errors.Is(err, updater.ErrPermissionDenied) {
			fmt.Println("Error: Permission denied removing binary. Try running with sudo:")
			fmt.Println("  sudo termchess --uninstall")
			return 1
		}
		fmt.Printf("Error: %v\n", err)
		return 1
	}

	fmt.Println("\u2713 TermChess has been uninstalled. Goodbye!")
	return 0
}
