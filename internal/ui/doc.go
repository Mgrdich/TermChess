// Package ui provides the terminal user interface for TermChess.
//
// This package implements a Bubbletea-based TUI that allows users to:
//   - Navigate menus and select game options
//   - Play chess games with SAN move input
//   - View the board in ASCII or Unicode format
//   - Save and load games using FEN notation
//   - Configure display settings
//
// The UI layer is separated from the game logic (internal/engine) and
// uses the Bubbletea framework for reactive, event-driven updates.
package ui
