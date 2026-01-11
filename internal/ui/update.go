package ui

import (
	"fmt"

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
	// Exception: 'q' on Settings screen returns to menu instead of quitting
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		// 'q' on Settings screen returns to menu, otherwise quits
		if m.screen != ScreenSettings {
			return m, tea.Quit
		}
	}

	// Handle screen-specific keys based on current screen
	switch m.screen {
	case ScreenMainMenu:
		return m.handleMainMenuKeys(msg)
	case ScreenGameTypeSelect:
		return m.handleGameTypeSelectKeys(msg)
	case ScreenGamePlay:
		return m.handleGamePlayKeys(msg)
	case ScreenGameOver:
		return m.handleGameOverKeys(msg)
	case ScreenSettings:
		return m.handleSettingsKeys(msg)
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
		// Transition to the GameTypeSelect screen
		m.screen = ScreenGameTypeSelect
		// Initialize the game type selection menu
		m.menuOptions = []string{"Player vs Player", "Player vs Bot"}
		m.menuSelection = 0
		// Clear any previous status messages
		m.statusMsg = ""
		m.errorMsg = ""
		// Clear any previous input
		m.input = ""

	case "Load Game":
		m.statusMsg = "Load Game selected (not yet implemented)"

	case "Settings":
		// Transition to the Settings screen
		m.screen = ScreenSettings
		// Reset settings selection to first option
		m.settingsSelection = 0
		// Clear any previous status messages
		m.statusMsg = ""
		m.errorMsg = ""
	}

	return m, nil
}

// handleGameTypeSelectKeys handles keyboard input for the GameTypeSelect screen.
// Supports arrow keys and vi-style navigation (j/k), Enter to select,
// and wraps around at top and bottom of the menu.
func (m Model) handleGameTypeSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		return m.handleGameTypeSelection()
	}

	return m, nil
}

// handleGameTypeSelection executes the action for the currently selected game type.
// If "Player vs Player" is selected, starts a new PvP game.
// If "Player vs Bot" is selected, shows a "Coming soon" message and returns to main menu.
func (m Model) handleGameTypeSelection() (tea.Model, tea.Cmd) {
	selected := m.menuOptions[m.menuSelection]

	switch selected {
	case "Player vs Player":
		// Set game type to PvP
		m.gameType = GameTypePvP
		// Create a new board with the standard starting position
		m.board = engine.NewBoard()
		// Switch to the GamePlay screen
		m.screen = ScreenGamePlay
		// Clear any previous status messages
		m.statusMsg = ""
		m.errorMsg = ""
		// Clear any previous input
		m.input = ""

	case "Player vs Bot":
		// Set game type to PvBot (for future use)
		m.gameType = GameTypePvBot
		// Show "Coming soon" message and return to main menu
		m.statusMsg = "Player vs Bot mode is coming soon!"
		m.screen = ScreenMainMenu
		// Reset menu options to main menu
		m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
		m.menuSelection = 0
	}

	return m, nil
}

// handleGamePlayKeys handles keyboard input for the GamePlay screen.
// Supports text input for entering chess moves in coordinate notation (e.g., "e2e4").
// Regular characters are appended to input, backspace deletes, and enter submits.
func (m Model) handleGamePlayKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyBackspace:
		// Remove the last character from input
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
		// Clear error messages when user modifies input
		m.errorMsg = ""

	case tea.KeyEnter:
		// Parse and execute the move if input is not empty
		if m.input != "" {
			// Try SAN parsing first
			move, err := ParseSAN(m.board, m.input)
			if err != nil {
				// Fall back to coordinate notation
				move, err = engine.ParseMove(m.input)
				if err != nil {
					// Show parsing error to user
					m.errorMsg = fmt.Sprintf("Invalid move: %v", err)
					return m, nil
				}
			}

			// Try to make the move on the board
			err = m.board.MakeMove(move)
			if err != nil {
				// Show move execution error to user
				m.errorMsg = err.Error()
				return m, nil
			}

			// Move was successful - clear input and error messages
			m.input = ""
			m.errorMsg = ""
			m.statusMsg = ""

			// Add move to history
			m.moveHistory = append(m.moveHistory, move)

			// Check if the game is over after this move
			if m.board.IsGameOver() {
				m.screen = ScreenGameOver
			}
		}

	case tea.KeyRunes:
		// Clear error messages when user starts typing a new move
		m.errorMsg = ""
		// Append the typed character(s) to the input
		// Only allow alphanumeric characters and basic symbols
		m.input += string(msg.Runes)
	}

	return m, nil
}

// handleGameOverKeys handles keyboard input for the GameOver screen.
// Supports 'n' for new game, 'm' for main menu, and 'q' for quit.
func (m Model) handleGameOverKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "n", "N":
		// Start a new game
		m.board = engine.NewBoard()
		m.moveHistory = []engine.Move{}
		m.screen = ScreenGamePlay
		m.input = ""
		m.errorMsg = ""
		m.statusMsg = ""

	case "m", "M":
		// Return to main menu
		m.screen = ScreenMainMenu
		m.board = nil
		m.moveHistory = []engine.Move{}
		m.input = ""
		m.errorMsg = ""
		m.statusMsg = ""

	case "q", "Q":
		// Quit the application
		return m, tea.Quit
	}

	return m, nil
}

// handleSettingsKeys handles keyboard input for the Settings screen.
// Supports arrow keys and vi-style navigation (j/k), Enter to toggle options,
// ESC/q/b/backspace to return to main menu, and wraps around at top and bottom.
func (m Model) handleSettingsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear any previous error or status messages when user takes action
	m.errorMsg = ""
	m.statusMsg = ""

	// Number of settings options
	numOptions := 4 // UseUnicode, ShowCoords, UseColors, ShowMoveHistory

	switch msg.String() {
	case "up", "k":
		// Move selection up
		if m.settingsSelection > 0 {
			m.settingsSelection--
		} else {
			// Wrap to bottom of menu
			m.settingsSelection = numOptions - 1
		}

	case "down", "j":
		// Move selection down
		if m.settingsSelection < numOptions-1 {
			m.settingsSelection++
		} else {
			// Wrap to top of menu
			m.settingsSelection = 0
		}

	case "enter":
		// Toggle the selected option
		switch m.settingsSelection {
		case 0: // Use Unicode Pieces
			m.config.UseUnicode = !m.config.UseUnicode
		case 1: // Show Coordinates
			m.config.ShowCoords = !m.config.ShowCoords
		case 2: // Use Colors
			m.config.UseColors = !m.config.UseColors
		case 3: // Show Move History
			m.config.ShowMoveHistory = !m.config.ShowMoveHistory
		}

		// Save the configuration immediately after toggling
		if err := SaveConfig(m.config); err != nil {
			m.errorMsg = fmt.Sprintf("Failed to save config: %v", err)
		} else {
			m.statusMsg = "Setting saved successfully"
		}

	case "esc", "q", "b", "backspace":
		// Return to main menu
		m.screen = ScreenMainMenu
		m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
		m.menuSelection = 0
		m.errorMsg = ""
		m.statusMsg = ""
	}

	return m, nil
}
