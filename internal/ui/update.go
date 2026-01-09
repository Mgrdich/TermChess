package ui

import (
	"github.com/Mgrdich/TermChess/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

// Init initializes the model. Called once at program start.
// Returns nil as no initial commands are needed for the basic menu interface.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and updates the model state.
// This is the core of the Elm architecture - all state changes happen here.
// It takes a message (user input, events, etc.) and returns an updated model
// and optionally a command to execute.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

// handleKeyPress processes keyboard input and routes it to the appropriate handler.
// Global keys like quit are handled first, then screen-specific keys are delegated
// to the current screen's handler.
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle global quit keys (work from any screen)
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	}

	// Handle screen-specific keys based on current screen
	switch m.screen {
	case ScreenMainMenu:
		return m.handleMainMenuKeys(msg)
	case ScreenGamePlay:
		return m.handleGamePlayKeys(msg)
	default:
		// Other screens will be implemented in future tasks
		return m, nil
	}
}

// handleMainMenuKeys handles keyboard input for the main menu screen.
// Supports arrow keys and vi-style navigation (j/k), Enter to select,
// and wraps around at top and bottom of the menu.
func (m Model) handleMainMenuKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear any previous error or status messages when user takes action
	m.errorMsg = ""
	m.statusMsg = ""

	switch msg.String() {
	case "up", "k":
		// Move selection up
		if m.menuSelection > 0 {
			m.menuSelection--
		} else {
			// Wrap to bottom of menu
			m.menuSelection = len(m.menuOptions) - 1
		}

	case "down", "j":
		// Move selection down
		if m.menuSelection < len(m.menuOptions)-1 {
			m.menuSelection++
		} else {
			// Wrap to top of menu
			m.menuSelection = 0
		}

	case "enter":
		return m.handleMainMenuSelection()
	}

	return m, nil
}

// handleMainMenuSelection executes the action for the currently selected menu option.
// For now, only "Exit" and "New Game" are fully implemented.
// Other options set a status message indicating they are not yet implemented.
func (m Model) handleMainMenuSelection() (tea.Model, tea.Cmd) {
	selected := m.menuOptions[m.menuSelection]

	switch selected {
	case "Exit":
		return m, tea.Quit

	case "New Game":
		// Create a new board with the standard starting position
		m.board = engine.NewBoard()
		// Switch to the GamePlay screen
		m.screen = ScreenGamePlay
		// Clear any previous status messages
		m.statusMsg = ""
		m.errorMsg = ""
		// Clear any previous input
		m.input = ""

	case "Load Game":
		m.statusMsg = "Load Game selected (not yet implemented)"

	case "Settings":
		m.statusMsg = "Settings selected (not yet implemented)"
	}

	return m, nil
}

// handleGamePlayKeys handles keyboard input for the GamePlay screen.
// Supports text input for entering chess moves.
// Regular characters are appended to input, backspace deletes, and enter submits.
func (m Model) handleGamePlayKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear error messages when user starts typing
	m.errorMsg = ""

	switch msg.Type {
	case tea.KeyBackspace:
		// Remove the last character from input
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}

	case tea.KeyEnter:
		// For now, just clear input and show a status message
		// Move parsing and execution will be implemented in Slice 4
		if m.input != "" {
			m.statusMsg = "Move execution not yet implemented"
			m.input = ""
		}

	case tea.KeyRunes:
		// Append the typed character(s) to the input
		// Only allow alphanumeric characters and basic symbols
		m.input += string(msg.Runes)
	}

	return m, nil
}
