package ui

import (
	"fmt"
	"strings"

	"github.com/Mgrdich/TermChess/internal/engine"
	"github.com/Mgrdich/TermChess/internal/util"
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
	// Handle global quit keys (work from any screen except GamePlay where 'q' shows save prompt)
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		// Only quit directly if not in GamePlay screen
		if m.screen != ScreenGamePlay {
			return m, tea.Quit
		}
		// Otherwise, let the GamePlay handler deal with it
	}

	// Handle screen-specific keys based on current screen
	switch m.screen {
	case ScreenMainMenu:
		return m.handleMainMenuKeys(msg)
	case ScreenGameTypeSelect:
		return m.handleGameTypeSelectKeys(msg)
	case ScreenFENInput:
		return m.handleFENInputKeys(msg)
	case ScreenGamePlay:
		return m.handleGamePlayKeys(msg)
	case ScreenGameOver:
		return m.handleGameOverKeys(msg)
	case ScreenSettings:
		return m.handleSettingsKeys(msg)
	case ScreenSavePrompt:
		return m.handleSavePromptKeys(msg)
	case ScreenResumePrompt:
		return m.handleResumePromptKeys(msg)
	case ScreenDrawPrompt:
		return m.handleDrawPromptKeys(msg)
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
		// Transition to game type selection screen
		m.screen = ScreenGameTypeSelect
		// Set up menu options for game type selection
		m.menuOptions = []string{"Player vs Player", "Player vs Bot"}
		m.menuSelection = 0
		// Clear any previous status messages
		m.statusMsg = ""
		m.errorMsg = ""
		// Clear any previous input
		m.input = ""

	case "Load Game":
		// Transition to FEN input screen
		m.screen = ScreenFENInput
		// Reset and focus the text input
		m.fenInput.SetValue("")
		m.fenInput.Focus()
		// Clear any previous status messages
		m.statusMsg = ""
		m.errorMsg = ""

	case "Settings":
		// Transition to settings screen
		m.screen = ScreenSettings
		m.settingsSelection = 0
		// Clear any previous status messages
		m.statusMsg = ""
		m.errorMsg = ""
	}

	return m, nil
}

// handleGameTypeSelectKeys handles keyboard input for the game type selection screen.
// Supports arrow keys and vi-style navigation (j/k), Enter to select,
// ESC to return to main menu, and wraps around at top and bottom of the menu.
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

	case "esc":
		// Return to main menu
		m.screen = ScreenMainMenu
		m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
		m.menuSelection = 0
		m.errorMsg = ""
		m.statusMsg = ""
	}

	return m, nil
}

// handleGameTypeSelection executes the action for the currently selected game type option.
// "Player vs Player" starts a new PvP game.
// "Player vs Bot" shows a "coming soon" message and returns to main menu.
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
		// Reset resignation tracking
		m.resignedBy = -1
		// Reset draw offer state
		m.drawOfferedBy = -1
		m.drawOfferedByWhite = false
		m.drawOfferedByBlack = false
		m.drawByAgreement = false

	case "Player vs Bot":
		// Set game type to PvBot
		m.gameType = GameTypePvBot
		// Show coming soon message and keep user on game type select screen
		// They can press ESC or navigate to return to main menu
		m.statusMsg = "Coming soon - Bot play not yet implemented. Press ESC to return to menu."
	}

	return m, nil
}

// handleGamePlayKeys handles keyboard input for the GamePlay screen.
// Supports text input for entering chess moves in coordinate notation (e.g., "e2e4").
// Regular characters are appended to input, backspace deletes, and enter submits.
func (m Model) handleGamePlayKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check for 'q' key to show save prompt
	if msg.String() == "q" || msg.String() == "Q" {
		// Show save prompt
		m.screen = ScreenSavePrompt
		m.savePromptSelection = 0
		m.savePromptAction = "exit"
		m.errorMsg = ""
		m.statusMsg = ""
		return m, nil
	}

	switch msg.Type {
	case tea.KeyBackspace:
		// Remove the last character from input
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
		// Clear error messages when user modifies input
		m.errorMsg = ""

	case tea.KeyEnter:
		// Parse and execute the move or command if input is not empty
		if m.input != "" {
			return m.handleGamePlayInput()
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
		// Start a new game - go through game type selection
		m.board = nil
		m.moveHistory = []engine.Move{}
		m.screen = ScreenGameTypeSelect
		m.input = ""
		m.errorMsg = ""
		m.statusMsg = ""
		// Set up menu options for game type selection
		m.menuOptions = []string{"Player vs Player", "Player vs Bot"}
		m.menuSelection = 0
		// Reset draw offer state
		m.drawOfferedBy = -1
		m.drawOfferedByWhite = false
		m.drawOfferedByBlack = false
		m.drawByAgreement = false

	case "m", "M":
		// Return to main menu
		m.screen = ScreenMainMenu
		m.board = nil
		m.moveHistory = []engine.Move{}
		m.input = ""
		m.errorMsg = ""
		m.statusMsg = ""
		// Reset menu options to main menu
		m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
		m.menuSelection = 0

	case "q", "Q":
		// Quit the application
		return m, tea.Quit
	}

	return m, nil
}

// handleSettingsKeys handles keyboard input for the Settings screen.
// Supports arrow keys and vi-style navigation (j/k), Space or Enter to toggle,
// ESC to return to main menu, and wraps around at top and bottom of the settings.
func (m Model) handleSettingsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear any previous error or status messages when user takes action
	m.errorMsg = ""
	m.statusMsg = ""

	// Number of settings options
	numSettings := 5 // UseUnicode, ShowCoords, UseColors, ShowMoveHistory, ShowHelpText

	switch msg.String() {
	case "up", "k":
		// Move selection up
		if m.settingsSelection > 0 {
			m.settingsSelection--
		} else {
			// Wrap to bottom of settings
			m.settingsSelection = numSettings - 1
		}

	case "down", "j":
		// Move selection down
		if m.settingsSelection < numSettings-1 {
			m.settingsSelection++
		} else {
			// Wrap to top of settings
			m.settingsSelection = 0
		}

	case "enter", " ":
		// Toggle the selected setting
		return m.toggleSelectedSetting()

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

// toggleSelectedSetting toggles the currently selected setting and saves the config.
func (m Model) toggleSelectedSetting() (tea.Model, tea.Cmd) {
	// Toggle the selected setting based on settingsSelection index
	switch m.settingsSelection {
	case 0: // Use Unicode Pieces
		m.config.UseUnicode = !m.config.UseUnicode
	case 1: // Show Coordinates
		m.config.ShowCoords = !m.config.ShowCoords
	case 2: // Use Colors
		m.config.UseColors = !m.config.UseColors
	case 3: // Show Move History
		m.config.ShowMoveHistory = !m.config.ShowMoveHistory
	case 4: // Show Help Text
		m.config.ShowHelpText = !m.config.ShowHelpText
	}

	// Save the configuration immediately
	err := SaveConfig(m.config)
	if err != nil {
		m.errorMsg = fmt.Sprintf("Failed to save settings: %v", err)
	} else {
		m.statusMsg = "Setting saved successfully"
	}

	return m, nil
}

// handleSavePromptKeys handles keyboard input for the Save Prompt screen.
// Supports arrow keys to navigate between Yes/No, Enter to confirm, direct 'y'/'n' keys, and ESC to cancel.
func (m Model) handleSavePromptKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear any previous error or status messages when user takes action
	m.errorMsg = ""
	m.statusMsg = ""

	switch msg.String() {
	case "up", "k":
		// Move selection up (toggle between Yes and No)
		if m.savePromptSelection > 0 {
			m.savePromptSelection--
		} else {
			// Wrap to bottom (only 2 options)
			m.savePromptSelection = 1
		}

	case "down", "j":
		// Move selection down (toggle between Yes and No)
		if m.savePromptSelection < 1 {
			m.savePromptSelection++
		} else {
			// Wrap to top
			m.savePromptSelection = 0
		}

	case "y", "Y":
		// Direct "Yes" - save the game
		err := SaveGame(m.board)
		if err != nil {
			m.errorMsg = fmt.Sprintf("Failed to save game: %v", err)
			return m, nil
		}
		// Save completed successfully, execute the action (exit or menu)
		if m.savePromptAction == "exit" {
			return m, tea.Quit
		}
		// Return to main menu
		m.screen = ScreenMainMenu
		m.board = nil
		m.moveHistory = []engine.Move{}
		m.input = ""
		m.errorMsg = ""
		m.statusMsg = ""
		m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
		m.menuSelection = 0

	case "n", "N":
		// Direct "No" - don't save, just execute the action
		if m.savePromptAction == "exit" {
			return m, tea.Quit
		}
		// Return to main menu without saving
		m.screen = ScreenMainMenu
		m.board = nil
		m.moveHistory = []engine.Move{}
		m.input = ""
		m.errorMsg = ""
		m.statusMsg = ""
		m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
		m.menuSelection = 0

	case "enter":
		// Execute the selected action
		if m.savePromptSelection == 0 {
			// User selected "Yes" - save the game
			err := SaveGame(m.board)
			if err != nil {
				m.errorMsg = fmt.Sprintf("Failed to save game: %v", err)
				return m, nil
			}
		}
		// User selected "No" or save completed successfully
		// Execute the action (exit or menu)
		if m.savePromptAction == "exit" {
			return m, tea.Quit
		}
		// Return to main menu
		m.screen = ScreenMainMenu
		m.board = nil
		m.moveHistory = []engine.Move{}
		m.input = ""
		m.errorMsg = ""
		m.statusMsg = ""
		m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
		m.menuSelection = 0

	case "esc":
		// Cancel and return to game
		m.screen = ScreenGamePlay
		m.errorMsg = ""
		m.statusMsg = ""
	}

	return m, nil
}

// handleResumePromptKeys handles keyboard input for the Resume Prompt screen.
// Supports arrow keys to navigate between Yes/No, Enter to confirm.
// User can choose to resume the saved game or start from main menu.
func (m Model) handleResumePromptKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear any previous error or status messages when user takes action
	m.errorMsg = ""
	m.statusMsg = ""

	switch msg.String() {
	case "up", "k":
		// Move selection up (toggle between Yes and No)
		if m.resumePromptSelection > 0 {
			m.resumePromptSelection--
		} else {
			// Wrap to bottom (only 2 options)
			m.resumePromptSelection = 1
		}

	case "down", "j":
		// Move selection down (toggle between Yes and No)
		if m.resumePromptSelection < 1 {
			m.resumePromptSelection++
		} else {
			// Wrap to top
			m.resumePromptSelection = 0
		}

	case "y", "Y":
		// Direct "Yes" - load the saved game
		board, err := LoadGame()
		if err != nil {
			// Failed to load - show error and go to main menu
			m.errorMsg = fmt.Sprintf("Failed to load saved game: %v", err)
			m.screen = ScreenMainMenu
			m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
			m.menuSelection = 0
			return m, nil
		}

		// Successfully loaded - start gameplay with loaded board
		m.board = board
		m.moveHistory = []engine.Move{}
		m.screen = ScreenGamePlay
		m.input = ""
		m.errorMsg = ""
		m.statusMsg = "Game resumed"
		m.resignedBy = -1
		// Reset draw offer state
		m.drawOfferedBy = -1
		m.drawOfferedByWhite = false
		m.drawOfferedByBlack = false
		m.drawByAgreement = false

	case "n", "N":
		// Direct "No" - go to main menu
		m.screen = ScreenMainMenu
		m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
		m.menuSelection = 0
		m.errorMsg = ""
		m.statusMsg = ""

	case "enter":
		// Execute the selected action
		if m.resumePromptSelection == 0 {
			// User selected "Yes" - load the saved game
			board, err := LoadGame()
			if err != nil {
				// Failed to load - show error and go to main menu
				m.errorMsg = fmt.Sprintf("Failed to load saved game: %v", err)
				m.screen = ScreenMainMenu
				m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
				m.menuSelection = 0
				return m, nil
			}

			// Successfully loaded - start gameplay with loaded board
			m.board = board
			m.moveHistory = []engine.Move{}
			m.screen = ScreenGamePlay
			m.input = ""
			m.errorMsg = ""
			m.statusMsg = "Game resumed"
			m.resignedBy = -1
			// Reset draw offer state
			m.drawOfferedBy = -1
			m.drawOfferedByWhite = false
			m.drawOfferedByBlack = false
			m.drawByAgreement = false
		} else {
			// User selected "No" - go to main menu
			m.screen = ScreenMainMenu
			m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
			m.menuSelection = 0
			m.errorMsg = ""
			m.statusMsg = ""
		}
	}

	return m, nil
}

// handleFENInputKeys handles keyboard input for the FEN Input screen.
// Supports text input for entering FEN strings, Enter to parse and load,
// and Esc to return to main menu.
func (m Model) handleFENInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		// Return to main menu
		m.screen = ScreenMainMenu
		m.menuOptions = []string{"New Game", "Load Game", "Settings", "Exit"}
		m.menuSelection = 0
		m.errorMsg = ""
		m.statusMsg = ""
		m.fenInput.SetValue("")
		return m, nil

	case "enter":
		// Try to parse and load the FEN string
		fenString := m.fenInput.Value()
		if fenString == "" {
			m.errorMsg = "Please enter a FEN string"
			return m, nil
		}

		// Parse the FEN string using the engine
		board, err := engine.FromFEN(fenString)
		if err != nil {
			// Show parsing error to user
			m.errorMsg = fmt.Sprintf("Invalid FEN: %v", err)
			return m, nil
		}

		// Successfully loaded - start gameplay with loaded board
		m.board = board
		m.moveHistory = []engine.Move{}
		m.screen = ScreenGamePlay
		m.gameType = GameTypePvP
		m.input = ""
		m.errorMsg = ""
		m.statusMsg = ""
		m.fenInput.SetValue("")
		m.resignedBy = -1
		// Reset draw offer state
		m.drawOfferedBy = -1
		m.drawOfferedByWhite = false
		m.drawOfferedByBlack = false
		m.drawByAgreement = false
		return m, nil

	default:
		// Delegate to the text input component for regular typing
		m.fenInput, cmd = m.fenInput.Update(msg)
		// Clear error message when user starts typing
		if msg.Type == tea.KeyRunes || msg.Type == tea.KeyBackspace {
			m.errorMsg = ""
		}
	}

	return m, cmd
}

// handleGamePlayInput processes user input during gameplay.
// It first checks if the input is a special command (resign, showfen, menu),
// and if not, attempts to parse and execute it as a chess move.
func (m Model) handleGamePlayInput() (tea.Model, tea.Cmd) {
	// Get the trimmed and lowercased input for command matching
	input := strings.TrimSpace(strings.ToLower(m.input))

	// Check for special commands first
	switch input {
	case "resign":
		return m.handleResignCommand()
	case "showfen":
		return m.handleShowFenCommand()
	case "menu":
		return m.handleMenuCommand()
	case "offerdraw":
		return m.handleOfferDrawCommand()
	default:
		// Not a command, try to parse as a move
		return m.handleMoveInput()
	}
}

// handleResignCommand handles the "resign" command.
// The current player resigns, and the game transitions to GameOver screen.
func (m Model) handleResignCommand() (tea.Model, tea.Cmd) {
	// Mark which player resigned
	m.resignedBy = int8(m.board.ActiveColor)

	// Transition to game over screen
	m.screen = ScreenGameOver

	// Clear input
	m.input = ""
	m.errorMsg = ""
	m.statusMsg = ""

	// Delete the save game file since the game is over
	_ = DeleteSaveGame()

	return m, nil
}

// handleShowFenCommand handles the "showfen" command.
// It displays the current FEN string and copies it to clipboard if possible.
func (m Model) handleShowFenCommand() (tea.Model, tea.Cmd) {
	// Get the FEN string from the current board
	fen := m.board.ToFEN()

	// Try to copy to clipboard
	err := util.CopyToClipboard(fen)
	if err != nil {
		// Show FEN with clipboard error message
		m.statusMsg = fmt.Sprintf("FEN: %s (Failed to copy to clipboard: %v)", fen, err)
	} else {
		// Show FEN with success message
		m.statusMsg = fmt.Sprintf("FEN: %s (Copied to clipboard)", fen)
	}

	// Clear input and error messages
	m.input = ""
	m.errorMsg = ""

	return m, nil
}

// handleMenuCommand handles the "menu" command.
// It shows the save prompt before returning to the main menu.
func (m Model) handleMenuCommand() (tea.Model, tea.Cmd) {
	// Transition to save prompt
	m.screen = ScreenSavePrompt
	m.savePromptSelection = 0
	m.savePromptAction = "menu"

	// Clear input and messages
	m.input = ""
	m.errorMsg = ""
	m.statusMsg = ""

	return m, nil
}

// handleMoveInput parses and executes a chess move.
// It tries SAN notation first, then falls back to coordinate notation.
func (m Model) handleMoveInput() (tea.Model, tea.Cmd) {
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
		// Delete the save game file since the game is over
		_ = DeleteSaveGame()
	}

	return m, nil
}

// handleOfferDrawCommand handles the "offerdraw" command.
// A player offers a draw to their opponent, which can be accepted or declined.
func (m Model) handleOfferDrawCommand() (tea.Model, tea.Cmd) {
	// Check if this player already offered a draw
	if (m.board.ActiveColor == engine.White && m.drawOfferedByWhite) ||
		(m.board.ActiveColor == engine.Black && m.drawOfferedByBlack) {
		m.errorMsg = "You have already offered a draw this game"
		m.input = ""
		return m, nil
	}

	// Mark who offered the draw
	m.drawOfferedBy = int8(m.board.ActiveColor)
	if m.board.ActiveColor == engine.White {
		m.drawOfferedByWhite = true
	} else {
		m.drawOfferedByBlack = true
	}

	// Transition to draw prompt
	m.screen = ScreenDrawPrompt
	m.drawPromptSelection = 0
	m.input = ""
	m.errorMsg = ""
	m.statusMsg = ""

	return m, nil
}

// handleDrawPromptKeys handles keyboard input for the Draw Prompt screen.
// Supports arrow keys to navigate between Accept/Decline, Enter to confirm, and ESC to cancel.
func (m Model) handleDrawPromptKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear any previous error or status messages when user takes action
	m.errorMsg = ""
	m.statusMsg = ""

	switch msg.String() {
	case "up", "k":
		// Move selection up (toggle between Accept and Decline)
		if m.drawPromptSelection > 0 {
			m.drawPromptSelection--
		} else {
			// Wrap to bottom (only 2 options)
			m.drawPromptSelection = 1
		}

	case "down", "j":
		// Move selection down (toggle between Accept and Decline)
		if m.drawPromptSelection < 1 {
			m.drawPromptSelection++
		} else {
			// Wrap to top
			m.drawPromptSelection = 0
		}

	case "enter":
		// Execute the selected action
		if m.drawPromptSelection == 0 {
			// User selected "Accept" - end game in draw
			m.drawByAgreement = true
			m.screen = ScreenGameOver
			m.input = ""
			m.errorMsg = ""
			m.statusMsg = ""
			// Delete the save game file since the game is over
			_ = DeleteSaveGame()
		} else {
			// User selected "Decline" - return to game
			m.screen = ScreenGamePlay
			m.statusMsg = "Draw offer declined"
			m.input = ""
			m.errorMsg = ""
			// Reset draw offered by so another offer can be made
			m.drawOfferedBy = -1
		}

	case "esc":
		// Cancel and return to game
		m.screen = ScreenGamePlay
		m.statusMsg = "Draw offer cancelled"
		m.input = ""
		m.errorMsg = ""
		// Reset draw offered by and the flag for the player who offered
		if m.drawOfferedBy == int8(engine.White) {
			m.drawOfferedByWhite = false
		} else if m.drawOfferedBy == int8(engine.Black) {
			m.drawOfferedByBlack = false
		}
		m.drawOfferedBy = -1
	}

	return m, nil
}
