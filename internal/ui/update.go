package ui

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Mgrdich/TermChess/internal/bot"
	"github.com/Mgrdich/TermChess/internal/bvb"
	"github.com/Mgrdich/TermChess/internal/config"
	"github.com/Mgrdich/TermChess/internal/engine"
	"github.com/Mgrdich/TermChess/internal/util"
	tea "github.com/charmbracelet/bubbletea"
)

// BvBTickMsg triggers a UI re-render for Bot vs Bot gameplay.
type BvBTickMsg struct{}

// BotMoveMsg is sent when the bot has selected a move.
type BotMoveMsg struct {
	move engine.Move
}

// BotMoveErrorMsg is sent when the bot encounters an error during move selection.
type BotMoveErrorMsg struct {
	err error
}

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
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		return m, nil
	case BvBTickMsg:
		return m.handleBvBTick()
	case BotMoveMsg:
		return m.handleBotMove(msg)
	case BotMoveErrorMsg:
		return m.handleBotMoveError(msg)
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
		// Clean up bot engine if it exists
		if m.botEngine != nil {
			_ = m.botEngine.Close()
		}
		// Clean up BvB manager if running
		if m.bvbManager != nil {
			m.bvbManager.Abort()
			m.bvbManager = nil
		}
		return m, tea.Quit
	case "q":
		// Only quit directly if not in GamePlay screen
		if m.screen != ScreenGamePlay {
			// Clean up bot engine if it exists
			if m.botEngine != nil {
				_ = m.botEngine.Close()
			}
			// Clean up BvB manager if running
			if m.bvbManager != nil {
				m.bvbManager.Abort()
				m.bvbManager = nil
			}
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
	case ScreenBotSelect:
		return m.handleBotSelectKeys(msg)
	case ScreenColorSelect:
		return m.handleColorSelectKeys(msg)
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
	case ScreenBvBBotSelect:
		return m.handleBvBBotSelectKeys(msg)
	case ScreenBvBGameMode:
		return m.handleBvBGameModeKeys(msg)
	case ScreenBvBGridConfig:
		return m.handleBvBGridConfigKeys(msg)
	case ScreenBvBGamePlay:
		return m.handleBvBGamePlayKeys(msg)
	case ScreenBvBStats:
		return m.handleBvBStatsKeys(msg)
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
// Handles all main menu options including the dynamic "Resume Game" option.
func (m Model) handleMainMenuSelection() (tea.Model, tea.Cmd) {
	selected := m.menuOptions[m.menuSelection]

	switch selected {
	case "Resume Game":
		// Load the saved game and start gameplay
		board, err := config.LoadGame()
		if err != nil {
			// Failed to load - show error and stay on main menu
			m.errorMsg = fmt.Sprintf("Failed to load saved game: %v", err)
			return m, nil
		}

		// Successfully loaded - start gameplay with loaded board
		m.board = board
		m.moveHistory = []engine.Move{}
		m.clearNavStack() // Clear nav stack when starting game
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

	case "Exit":
		return m, tea.Quit

	case "New Game":
		// Transition to game type selection screen using navigation stack
		m.pushScreen(ScreenGameTypeSelect)
		// Set up menu options for game type selection
		m.menuOptions = []string{"Player vs Player", "Player vs Bot", "Bot vs Bot"}
		m.menuSelection = 0
		// Clear any previous status messages
		m.statusMsg = ""
		m.errorMsg = ""
		// Clear any previous input
		m.input = ""

	case "Load Game":
		// Transition to FEN input screen using navigation stack
		m.pushScreen(ScreenFENInput)
		// Reset and focus the text input
		m.fenInput.SetValue("")
		m.fenInput.Focus()
		// Clear any previous status messages
		m.statusMsg = ""
		m.errorMsg = ""

	case "Settings":
		// Transition to settings screen using navigation stack
		m.pushScreen(ScreenSettings)
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
		// Return to previous screen using navigation stack
		m.popScreen()
		// Rebuild menu options in case we're back at main menu
		if m.screen == ScreenMainMenu {
			m.menuOptions = buildMainMenuOptions()
		}
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
		// Clear nav stack when starting game
		m.clearNavStack()
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
		// Transition to bot difficulty selection screen using navigation stack
		m.pushScreen(ScreenBotSelect)
		m.menuOptions = []string{"Easy", "Medium", "Hard"}
		m.menuSelection = 0
		m.statusMsg = ""
		m.errorMsg = ""

	case "Bot vs Bot":
		// Set game type to BvB
		m.gameType = GameTypeBvB
		// Start with selecting White bot difficulty
		m.bvbSelectingWhite = true
		m.screen = ScreenBvBBotSelect
		m.menuOptions = []string{"Easy", "Medium", "Hard"}
		m.menuSelection = 0
		m.statusMsg = ""
		m.errorMsg = ""
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

	// Check for 'esc' key to show save prompt before returning to menu
	if msg.String() == "esc" {
		// Show save prompt
		m.screen = ScreenSavePrompt
		m.savePromptSelection = 0
		m.savePromptAction = "menu"
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
// Supports 'n' for new game, 'm' for main menu, 'esc' for main menu, and 'q' for quit.
func (m Model) handleGameOverKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "n", "N":
		// Clean up bot engine if it exists
		if m.botEngine != nil {
			_ = m.botEngine.Close()
			m.botEngine = nil
		}
		// Start a new game - go through game type selection
		m.board = nil
		m.moveHistory = []engine.Move{}
		m.screen = ScreenGameTypeSelect
		m.input = ""
		m.errorMsg = ""
		m.statusMsg = ""
		// Set up menu options for game type selection
		m.menuOptions = []string{"Player vs Player", "Player vs Bot", "Bot vs Bot"}
		m.menuSelection = 0
		// Reset draw offer state
		m.drawOfferedBy = -1
		m.drawOfferedByWhite = false
		m.drawOfferedByBlack = false
		m.drawByAgreement = false

	case "m", "M", "esc":
		// Clean up bot engine if it exists
		if m.botEngine != nil {
			_ = m.botEngine.Close()
			m.botEngine = nil
		}
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
		// Clean up bot engine if it exists
		if m.botEngine != nil {
			_ = m.botEngine.Close()
		}
		// Quit the application
		return m, tea.Quit
	}

	return m, nil
}

// handleSettingsKeys handles keyboard input for the Settings screen.
// Supports arrow keys and vi-style navigation (j/k), Space or Enter to toggle/cycle,
// ESC to return to main menu, and wraps around at top and bottom of the settings.
func (m Model) handleSettingsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear any previous error or status messages when user takes action
	m.errorMsg = ""
	m.statusMsg = ""

	// Number of settings options (5 toggles + 1 theme selector)
	numSettings := 6 // UseUnicode, ShowCoords, UseColors, ShowMoveHistory, ShowHelpText, Theme

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
		// Return to previous screen using navigation stack
		m.popScreen()
		// Rebuild menu options if we're back at main menu
		if m.screen == ScreenMainMenu {
			m.menuOptions = buildMainMenuOptions()
		}
		m.menuSelection = 0
		m.errorMsg = ""
		m.statusMsg = ""
	}

	return m, nil
}

// toggleSelectedSetting toggles the currently selected setting and saves the config.
// For boolean settings, it toggles between true/false.
// For the theme setting, it cycles through: Classic -> Modern -> Minimalist -> Classic.
func (m Model) toggleSelectedSetting() (tea.Model, tea.Cmd) {
	// Toggle or cycle the selected setting based on settingsSelection index
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
	case 5: // Theme
		// Cycle through themes: Classic -> Modern -> Minimalist -> Classic
		m.config.Theme = cycleTheme(m.config.Theme)
		// Update the theme in the model immediately for visual feedback
		m.theme = GetTheme(ParseThemeName(m.config.Theme))
	}

	// Save the configuration immediately
	err := config.SaveConfig(m.config)
	if err != nil {
		m.errorMsg = fmt.Sprintf("Failed to save settings: %v", err)
	} else {
		m.statusMsg = "Setting saved successfully"
	}

	return m, nil
}

// cycleTheme cycles through theme names: classic -> modern -> minimalist -> classic.
func cycleTheme(current string) string {
	switch current {
	case ThemeNameClassic:
		return ThemeNameModern
	case ThemeNameModern:
		return ThemeNameMinimalist
	case ThemeNameMinimalist:
		return ThemeNameClassic
	default:
		// Unknown theme, reset to modern (next after classic)
		return ThemeNameModern
	}
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
		err := config.SaveGame(m.board)
		if err != nil {
			m.errorMsg = fmt.Sprintf("Failed to save game: %v", err)
			return m, nil
		}
		// Clean up bot engine if it exists
		if m.botEngine != nil {
			_ = m.botEngine.Close()
			m.botEngine = nil
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
		m.menuOptions = buildMainMenuOptions()
		m.menuSelection = 0

	case "n", "N":
		// Clean up bot engine if it exists
		if m.botEngine != nil {
			_ = m.botEngine.Close()
			m.botEngine = nil
		}
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
		m.menuOptions = buildMainMenuOptions()
		m.menuSelection = 0

	case "enter":
		// Execute the selected action
		if m.savePromptSelection == 0 {
			// User selected "Yes" - save the game
			err := config.SaveGame(m.board)
			if err != nil {
				m.errorMsg = fmt.Sprintf("Failed to save game: %v", err)
				return m, nil
			}
		}
		// Clean up bot engine if it exists
		if m.botEngine != nil {
			_ = m.botEngine.Close()
			m.botEngine = nil
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
		m.menuOptions = buildMainMenuOptions()
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
		board, err := config.LoadGame()
		if err != nil {
			// Failed to load - show error and go to main menu
			m.errorMsg = fmt.Sprintf("Failed to load saved game: %v", err)
			m.screen = ScreenMainMenu
			m.menuOptions = buildMainMenuOptions()
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
		m.menuOptions = buildMainMenuOptions()
		m.menuSelection = 0
		m.errorMsg = ""
		m.statusMsg = ""

	case "enter":
		// Execute the selected action
		if m.resumePromptSelection == 0 {
			// User selected "Yes" - load the saved game
			board, err := config.LoadGame()
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
			m.menuOptions = buildMainMenuOptions()
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
		// Return to previous screen using navigation stack
		m.popScreen()
		// Rebuild menu options if we're back at main menu
		if m.screen == ScreenMainMenu {
			m.menuOptions = buildMainMenuOptions()
		}
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
		// Clear nav stack when starting game
		m.clearNavStack()
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
	_ = config.DeleteSaveGame()

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
		_ = config.DeleteSaveGame()
		// Clean up bot engine if it exists
		if m.botEngine != nil {
			_ = m.botEngine.Close()
			m.botEngine = nil
		}
		return m, nil
	}

	// If this is a bot game and game is not over, trigger bot move
	if m.gameType == GameTypePvBot {
		return m.makeBotMove()
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
			_ = config.DeleteSaveGame()
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

// handleBotSelectKeys handles keyboard input for the bot difficulty selection screen.
// Supports arrow keys and vi-style navigation (j/k), Enter to select,
// ESC to return to game type selection, and wraps around at top and bottom of the menu.
func (m Model) handleBotSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		return m.handleBotDifficultySelection()

	case "esc":
		// Return to previous screen using navigation stack
		m.popScreen()
		// Rebuild menu options for game type selection if we're back there
		if m.screen == ScreenGameTypeSelect {
			m.menuOptions = []string{"Player vs Player", "Player vs Bot", "Bot vs Bot"}
		}
		m.menuSelection = 0
		m.errorMsg = ""
		m.statusMsg = ""
	}

	return m, nil
}

// handleBvBBotSelectKeys handles keyboard input for the BvB bot difficulty selection screen.
// This screen is used twice: first to select White's bot difficulty, then Black's.
// Supports arrow keys and vi-style navigation (j/k), Enter to select,
// ESC to go back, and wraps around at top and bottom of the menu.
func (m Model) handleBvBBotSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear any previous error or status messages when user takes action
	m.errorMsg = ""
	m.statusMsg = ""

	switch msg.String() {
	case "up", "k":
		if m.menuSelection > 0 {
			m.menuSelection--
		} else {
			m.menuSelection = len(m.menuOptions) - 1
		}

	case "down", "j":
		if m.menuSelection < len(m.menuOptions)-1 {
			m.menuSelection++
		} else {
			m.menuSelection = 0
		}

	case "enter":
		return m.handleBvBBotDifficultySelection()

	case "esc":
		if m.bvbSelectingWhite {
			// Go back to game type selection
			m.screen = ScreenGameTypeSelect
			m.menuOptions = []string{"Player vs Player", "Player vs Bot", "Bot vs Bot"}
			m.menuSelection = 0
			m.errorMsg = ""
			m.statusMsg = ""
		} else {
			// Go back to selecting White bot
			m.bvbSelectingWhite = true
			m.menuSelection = 0
			m.errorMsg = ""
			m.statusMsg = ""
		}
	}

	return m, nil
}

// handleBvBBotDifficultySelection executes the action for the currently selected BvB bot difficulty.
// If selecting White, stores the difficulty and moves to Black selection.
// If selecting Black, stores the difficulty and transitions to game mode selection.
func (m Model) handleBvBBotDifficultySelection() (tea.Model, tea.Cmd) {
	selected := m.menuOptions[m.menuSelection]

	var diff BotDifficulty
	switch selected {
	case "Easy":
		diff = BotEasy
	case "Medium":
		diff = BotMedium
	case "Hard":
		diff = BotHard
	}

	if m.bvbSelectingWhite {
		// Store White difficulty and move to Black selection
		m.bvbWhiteDiff = diff
		m.bvbSelectingWhite = false
		m.menuSelection = 0
		m.statusMsg = ""
		m.errorMsg = ""
	} else {
		// Store Black difficulty and transition to game mode selection
		m.bvbBlackDiff = diff
		m.screen = ScreenBvBGameMode
		m.menuOptions = []string{"Single Game", "Multi-Game"}
		m.menuSelection = 0
		m.bvbInputtingCount = false
		m.bvbCountInput = ""
		m.statusMsg = ""
		m.errorMsg = ""
	}

	return m, nil
}

// handleBvBGameModeKeys handles keyboard input for the BvB game mode selection screen.
// Supports menu navigation for Single/Multi-Game selection, and text input for game count.
func (m Model) handleBvBGameModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.errorMsg = ""

	if m.bvbInputtingCount {
		return m.handleBvBCountInput(msg)
	}

	switch msg.String() {
	case "up", "k":
		if m.menuSelection > 0 {
			m.menuSelection--
		} else {
			m.menuSelection = len(m.menuOptions) - 1
		}

	case "down", "j":
		if m.menuSelection < len(m.menuOptions)-1 {
			m.menuSelection++
		} else {
			m.menuSelection = 0
		}

	case "enter":
		return m.handleBvBGameModeSelection()

	case "esc":
		// Go back to BvB bot select (Black selection)
		m.screen = ScreenBvBBotSelect
		m.bvbSelectingWhite = false
		m.menuOptions = []string{"Easy", "Medium", "Hard"}
		m.menuSelection = 0
		m.statusMsg = ""
	}

	return m, nil
}

// handleBvBGameModeSelection executes the action for the selected game mode.
func (m Model) handleBvBGameModeSelection() (tea.Model, tea.Cmd) {
	selected := m.menuOptions[m.menuSelection]

	switch selected {
	case "Single Game":
		m.bvbGameCount = 1
		m.bvbGridRows = 1
		m.bvbGridCols = 1
		m.bvbViewMode = BvBSingleView
		return m.startBvBSession()

	case "Multi-Game":
		// Switch to text input mode for game count
		m.bvbInputtingCount = true
		m.bvbCountInput = ""
		m.statusMsg = ""
		m.errorMsg = ""
	}

	return m, nil
}

// handleBvBCountInput handles text input for the multi-game count.
func (m Model) handleBvBCountInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Cancel count input, go back to menu mode
		m.bvbInputtingCount = false
		m.bvbCountInput = ""
		m.errorMsg = ""

	case tea.KeyBackspace:
		if len(m.bvbCountInput) > 0 {
			m.bvbCountInput = m.bvbCountInput[:len(m.bvbCountInput)-1]
		}

	case tea.KeyEnter:
		// Validate and submit count
		count, err := parsePositiveInt(m.bvbCountInput)
		if err != nil {
			m.errorMsg = "Please enter a positive integer"
			return m, nil
		}
		m.bvbGameCount = count
		m.bvbInputtingCount = false
		m.screen = ScreenBvBGridConfig
		m.menuOptions = []string{"1x1", "2x2", "2x3", "2x4", "Custom"}
		m.menuSelection = 0
		m.bvbInputtingGrid = false
		m.bvbCustomGridInput = ""
		m.statusMsg = ""
		m.errorMsg = ""

	case tea.KeyRunes:
		// Only allow digits
		for _, r := range msg.Runes {
			if r >= '0' && r <= '9' {
				m.bvbCountInput += string(r)
			}
		}
	}

	return m, nil
}

// parsePositiveInt parses a string as a positive integer (>= 1).
func parsePositiveInt(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty input")
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("not a number")
		}
		n = n*10 + int(r-'0')
	}
	if n < 1 {
		return 0, fmt.Errorf("must be at least 1")
	}
	return n, nil
}

// handleBvBGridConfigKeys handles keyboard input for the BvB grid configuration screen.
// Supports preset selection (1x1, 2x2, 2x3, 2x4) or custom grid input.
func (m Model) handleBvBGridConfigKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.errorMsg = ""

	if m.bvbInputtingGrid {
		return m.handleBvBGridInput(msg)
	}

	switch msg.String() {
	case "up", "k":
		if m.menuSelection > 0 {
			m.menuSelection--
		} else {
			m.menuSelection = len(m.menuOptions) - 1
		}

	case "down", "j":
		if m.menuSelection < len(m.menuOptions)-1 {
			m.menuSelection++
		} else {
			m.menuSelection = 0
		}

	case "enter":
		return m.handleBvBGridSelection()

	case "esc":
		// Go back to game mode selection
		m.screen = ScreenBvBGameMode
		m.menuOptions = []string{"Single Game", "Multi-Game"}
		m.menuSelection = 0
		m.bvbInputtingGrid = false
		m.statusMsg = ""
	}

	return m, nil
}

// handleBvBGridSelection handles the selection of a grid preset or custom option.
func (m Model) handleBvBGridSelection() (tea.Model, tea.Cmd) {
	selected := m.menuOptions[m.menuSelection]

	switch selected {
	case "1x1":
		m.bvbGridRows, m.bvbGridCols = 1, 1
	case "2x2":
		m.bvbGridRows, m.bvbGridCols = 2, 2
	case "2x3":
		m.bvbGridRows, m.bvbGridCols = 2, 3
	case "2x4":
		m.bvbGridRows, m.bvbGridCols = 2, 4
	case "Custom":
		m.bvbInputtingGrid = true
		m.bvbCustomGridInput = ""
		return m, nil
	}

	return m.startBvBSession()
}

// handleBvBGridInput handles text input for custom grid dimensions.
// Expected format: "RxC" (e.g., "2x3").
func (m Model) handleBvBGridInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.bvbInputtingGrid = false
		m.bvbCustomGridInput = ""
		m.errorMsg = ""

	case tea.KeyBackspace:
		if len(m.bvbCustomGridInput) > 0 {
			m.bvbCustomGridInput = m.bvbCustomGridInput[:len(m.bvbCustomGridInput)-1]
		}

	case tea.KeyEnter:
		rows, cols, err := parseGridDimensions(m.bvbCustomGridInput)
		if err != nil {
			m.errorMsg = err.Error()
			return m, nil
		}
		m.bvbGridRows = rows
		m.bvbGridCols = cols
		m.bvbInputtingGrid = false
		return m.startBvBSession()

	case tea.KeyRunes:
		for _, r := range msg.Runes {
			if (r >= '0' && r <= '9') || r == 'x' || r == 'X' {
				m.bvbCustomGridInput += string(r)
			}
		}
	}

	return m, nil
}

// parseGridDimensions parses a grid string like "2x3" into rows and cols.
// Validates that total boards (rows*cols) does not exceed 8.
func parseGridDimensions(s string) (int, int, error) {
	// Find the 'x' or 'X' separator
	sep := -1
	for i, r := range s {
		if r == 'x' || r == 'X' {
			sep = i
			break
		}
	}
	if sep < 0 {
		return 0, 0, fmt.Errorf("use format RxC (e.g., 2x3)")
	}

	rows, err := parsePositiveInt(s[:sep])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid rows: %v", err)
	}
	cols, err := parsePositiveInt(s[sep+1:])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid cols: %v", err)
	}

	if rows*cols > 8 {
		return 0, 0, fmt.Errorf("max 8 boards (got %dx%d = %d)", rows, cols, rows*cols)
	}

	return rows, cols, nil
}

// startBvBSession creates a SessionManager with the configured settings and starts it.
func (m Model) startBvBSession() (tea.Model, tea.Cmd) {
	// Map UI bot difficulty to bvb bot difficulty
	whiteDiff := uiBotDiffToBvB(m.bvbWhiteDiff)
	blackDiff := uiBotDiffToBvB(m.bvbBlackDiff)
	whiteName := botDifficultyName(m.bvbWhiteDiff) + " Bot"
	blackName := botDifficultyName(m.bvbBlackDiff) + " Bot"

	manager := bvb.NewSessionManager(whiteDiff, blackDiff, whiteName, blackName, m.bvbGameCount)
	if err := manager.Start(); err != nil {
		// Engine creation failed - stay on game mode screen and show error
		m.errorMsg = "Failed to start bot session: " + err.Error()
		m.screen = ScreenBvBGameMode
		m.bvbInputtingCount = false
		return m, nil
	}

	m.bvbManager = manager
	m.bvbSpeed = bvb.SpeedNormal
	m.bvbSelectedGame = 0
	m.bvbViewMode = BvBSingleView
	m.bvbPaused = false
	m.screen = ScreenBvBGamePlay
	m.statusMsg = ""
	m.errorMsg = ""
	return m, bvbTickCmd(m.bvbSpeed)
}

// uiBotDiffToBvB maps the UI BotDifficulty to the bot package Difficulty.
func uiBotDiffToBvB(d BotDifficulty) bot.Difficulty {
	switch d {
	case BotEasy:
		return bot.Easy
	case BotMedium:
		return bot.Medium
	case BotHard:
		return bot.Hard
	default:
		return bot.Easy
	}
}

// handleBvBGamePlayKeys handles keyboard input during BvB game viewing.
// Supports pause/resume, speed changes, view toggle, game navigation, and abort.
func (m Model) handleBvBGamePlayKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Abort the session and return to menu
		if m.bvbManager != nil {
			m.bvbManager.Abort()
			m.bvbManager = nil
		}
		m.screen = ScreenMainMenu
		m.menuOptions = buildMainMenuOptions()
		m.menuSelection = 0
		m.statusMsg = ""
		m.errorMsg = ""
		return m, nil

	case " ":
		// Toggle pause/resume
		if m.bvbManager != nil {
			if m.bvbPaused {
				m.bvbManager.Resume()
				m.bvbPaused = false
			} else {
				m.bvbManager.Pause()
				m.bvbPaused = true
			}
		}

	case "1":
		m.bvbSpeed = bvb.SpeedInstant
		if m.bvbManager != nil {
			m.bvbManager.SetSpeed(m.bvbSpeed)
		}
	case "2":
		m.bvbSpeed = bvb.SpeedFast
		if m.bvbManager != nil {
			m.bvbManager.SetSpeed(m.bvbSpeed)
		}
	case "3":
		m.bvbSpeed = bvb.SpeedNormal
		if m.bvbManager != nil {
			m.bvbManager.SetSpeed(m.bvbSpeed)
		}
	case "4":
		m.bvbSpeed = bvb.SpeedSlow
		if m.bvbManager != nil {
			m.bvbManager.SetSpeed(m.bvbSpeed)
		}

	case "tab":
		// Toggle view mode
		if m.bvbViewMode == BvBGridView {
			m.bvbViewMode = BvBSingleView
		} else {
			m.bvbViewMode = BvBGridView
		}

	case "left", "h":
		if m.bvbManager != nil {
			if m.bvbViewMode == BvBSingleView {
				// Previous game in single view
				sessions := m.bvbManager.Sessions()
				if m.bvbSelectedGame > 0 {
					m.bvbSelectedGame--
				} else {
					m.bvbSelectedGame = len(sessions) - 1
				}
			} else {
				// Previous page in grid view
				if m.bvbPageIndex > 0 {
					m.bvbPageIndex--
				}
			}
		}

	case "right", "l":
		if m.bvbManager != nil {
			if m.bvbViewMode == BvBSingleView {
				// Next game in single view
				sessions := m.bvbManager.Sessions()
				if m.bvbSelectedGame < len(sessions)-1 {
					m.bvbSelectedGame++
				} else {
					m.bvbSelectedGame = 0
				}
			} else {
				// Next page in grid view
				sessions := m.bvbManager.Sessions()
				boardsPerPage := m.bvbGridRows * m.bvbGridCols
				totalPages := (len(sessions) + boardsPerPage - 1) / boardsPerPage
				if m.bvbPageIndex < totalPages-1 {
					m.bvbPageIndex++
				}
			}
		}

	case "f":
		// Export FEN of the focused game
		if m.bvbManager != nil {
			sessions := m.bvbManager.Sessions()
			var targetSession *bvb.GameSession
			if m.bvbViewMode == BvBSingleView {
				if m.bvbSelectedGame < len(sessions) {
					targetSession = sessions[m.bvbSelectedGame]
				}
			} else {
				// In grid view, use first visible game on current page
				boardsPerPage := m.bvbGridRows * m.bvbGridCols
				startIdx := m.bvbPageIndex * boardsPerPage
				if startIdx < len(sessions) {
					targetSession = sessions[startIdx]
				}
			}
			if targetSession != nil {
				board := targetSession.CurrentBoard()
				if board != nil {
					fen := board.ToFEN()
					err := util.CopyToClipboard(fen)
					if err != nil {
						m.statusMsg = fmt.Sprintf("FEN: %s (Failed to copy: %v)", fen, err)
					} else {
						m.statusMsg = fmt.Sprintf("FEN copied to clipboard")
					}
				}
			}
		}
	}

	return m, nil
}

// bvbTickCmd returns a command that sends a BvBTickMsg after a delay based on speed.
func bvbTickCmd(speed bvb.PlaybackSpeed) tea.Cmd {
	delay := speed.Duration()
	if delay == 0 {
		// For instant speed, use a short tick interval for rendering
		delay = 100 * time.Millisecond
	}
	return tea.Tick(delay, func(time.Time) tea.Msg {
		return BvBTickMsg{}
	})
}

// handleBvBTick handles tick messages for BvB gameplay updates.
func (m Model) handleBvBTick() (tea.Model, tea.Cmd) {
	if m.screen != ScreenBvBGamePlay || m.bvbManager == nil {
		return m, nil
	}

	if m.bvbManager.AllFinished() {
		m.screen = ScreenBvBStats
		m.bvbStatsSelection = 0
		m.bvbStatsResultsPage = 0
		m.menuOptions = []string{"New Session", "Return to Menu"}
		return m, nil
	}

	// Schedule next tick
	return m, bvbTickCmd(m.bvbSpeed)
}

// handleBvBStatsKeys handles keyboard input on the BvB statistics screen.
func (m Model) handleBvBStatsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.bvbStatsSelection > 0 {
			m.bvbStatsSelection--
		}
	case "down", "j":
		if m.bvbStatsSelection < len(m.menuOptions)-1 {
			m.bvbStatsSelection++
		}
	case "left", "h":
		// Previous page of individual results
		if m.bvbStatsResultsPage > 0 {
			m.bvbStatsResultsPage--
		}
	case "right", "l":
		// Next page of individual results
		if m.bvbManager != nil {
			stats := m.bvbManager.Stats()
			if stats != nil && stats.TotalGames > 1 {
				totalPages := (len(stats.IndividualResults) + 14) / 15 // resultsPerPage = 15
				if m.bvbStatsResultsPage < totalPages-1 {
					m.bvbStatsResultsPage++
				}
			}
		}
	case "enter":
		return m.handleBvBStatsSelection()
	case "esc":
		m.screen = ScreenMainMenu
		m.menuOptions = buildMainMenuOptions()
		m.menuSelection = 0
		m.bvbManager = nil
	}
	return m, nil
}

// handleBvBStatsSelection handles the selected action on the stats screen.
func (m Model) handleBvBStatsSelection() (tea.Model, tea.Cmd) {
	switch m.bvbStatsSelection {
	case 0: // New Session
		m.screen = ScreenBvBBotSelect
		m.menuOptions = []string{"Easy", "Medium", "Hard"}
		m.menuSelection = 0
		m.bvbSelectingWhite = true
		m.bvbManager = nil
	case 1: // Return to Menu
		m.screen = ScreenMainMenu
		m.menuOptions = buildMainMenuOptions()
		m.menuSelection = 0
		m.bvbManager = nil
	}
	return m, nil
}

// handleColorSelectKeys handles keyboard input for the color selection screen.
// Supports arrow keys and vi-style navigation (j/k), Enter to select,
// ESC to return to bot difficulty selection, and wraps around at top and bottom of the menu.
func (m Model) handleColorSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		return m.handleColorSelection()

	case "esc":
		// Return to previous screen using navigation stack
		m.popScreen()
		// Rebuild menu options for bot selection if we're back there
		if m.screen == ScreenBotSelect {
			m.menuOptions = []string{"Easy", "Medium", "Hard"}
		}
		m.menuSelection = 0
		m.errorMsg = ""
		m.statusMsg = ""
	}

	return m, nil
}

// handleColorSelection executes the action for the currently selected color.
// Sets the user's color and starts a new game.
// If user plays Black, triggers bot's opening move.
func (m Model) handleColorSelection() (tea.Model, tea.Cmd) {
	selected := m.menuOptions[m.menuSelection]

	switch selected {
	case "Play as White":
		m.userColor = engine.White
	case "Play as Black":
		m.userColor = engine.Black
	}

	// Create a new board with the standard starting position
	m.board = engine.NewBoard()
	// Clear nav stack when starting game
	m.clearNavStack()
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

	// If user plays Black, bot should make the opening move
	if m.userColor == engine.Black {
		return m.makeBotMove()
	}

	return m, nil
}

// handleBotDifficultySelection executes the action for the currently selected bot difficulty.
// Sets the bot difficulty and transitions to color selection.
func (m Model) handleBotDifficultySelection() (tea.Model, tea.Cmd) {
	selected := m.menuOptions[m.menuSelection]

	switch selected {
	case "Easy":
		m.botDifficulty = BotEasy
	case "Medium":
		m.botDifficulty = BotMedium
	case "Hard":
		m.botDifficulty = BotHard
	}

	// Transition to color selection screen using navigation stack
	m.pushScreen(ScreenColorSelect)
	m.menuOptions = []string{"Play as White", "Play as Black"}
	m.menuSelection = 0
	m.statusMsg = ""
	m.errorMsg = ""

	return m, nil
}

// makeBotMove initiates a bot move calculation asynchronously.
// It displays a thinking message, creates the appropriate bot engine based on difficulty,
// and returns a command that will execute the move selection in a goroutine.
func (m Model) makeBotMove() (Model, tea.Cmd) {
	// Display thinking message
	m.statusMsg = getRandomThinkingMessage()

	// Create bot engine based on difficulty
	var botEngine bot.Engine
	var err error
	switch m.botDifficulty {
	case BotEasy:
		botEngine, err = bot.NewRandomEngine()
	case BotMedium:
		botEngine, err = bot.NewMinimaxEngine(bot.Medium)
	case BotHard:
		botEngine, err = bot.NewMinimaxEngine(bot.Hard)
	}

	if err != nil {
		return m, func() tea.Msg {
			return BotMoveErrorMsg{err: err}
		}
	}

	// Store engine for cleanup
	m.botEngine = botEngine

	// Execute bot move asynchronously
	return m, func() tea.Msg {
		// Track start time for minimum delay enforcement
		startTime := time.Now()

		// Determine minimum delay based on difficulty
		minDelay := getMinimumBotDelay(m.botDifficulty)

		ctx := context.Background()
		move, err := botEngine.SelectMove(ctx, m.board)
		if err != nil {
			return BotMoveErrorMsg{err: err}
		}

		// Enforce minimum delay for natural feel
		elapsed := time.Since(startTime)
		if elapsed < minDelay {
			time.Sleep(minDelay - elapsed)
		}

		return BotMoveMsg{move: move}
	}
}

// getMinimumBotDelay returns the minimum delay for bot moves based on difficulty.
// This ensures bot moves feel natural and not instantaneous, especially for Easy difficulty.
// The delay is randomized within a range to add variety and feel more human-like.
func getMinimumBotDelay(difficulty BotDifficulty) time.Duration {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	switch difficulty {
	case BotEasy:
		// Easy: 1-2 seconds (random for variety)
		// Random engine returns instantly, so this is critical
		minSeconds := 1.0 + rng.Float64() // 1.0 to 2.0 seconds
		return time.Duration(minSeconds * float64(time.Second))
	case BotMedium:
		// Medium: 1-2 seconds minimum
		// Minimax usually takes 2-4 seconds naturally, so this is a safety net
		minSeconds := 1.0 + rng.Float64() // 1.0 to 2.0 seconds
		return time.Duration(minSeconds * float64(time.Second))
	case BotHard:
		// Hard: 1 second minimum
		// Minimax usually takes 4-8 seconds naturally, so delay rarely needed
		return 1 * time.Second
	default:
		// Fallback to 1 second
		return 1 * time.Second
	}
}

// handleBotMove processes a successful bot move.
// It applies the move to the board, clears the status message, adds the move to history,
// and checks if the game is over.
func (m Model) handleBotMove(msg BotMoveMsg) (tea.Model, tea.Cmd) {
	// Try to make the move on the board
	err := m.board.MakeMove(msg.move)
	if err != nil {
		// Invalid move from bot - show error
		m.errorMsg = fmt.Sprintf("Bot generated invalid move: %v", err)
		m.statusMsg = ""
		return m, nil
	}

	// Move was successful - clear status message
	m.statusMsg = ""
	m.errorMsg = ""

	// Add move to history
	m.moveHistory = append(m.moveHistory, msg.move)

	// Check if the game is over after this move
	if m.board.IsGameOver() {
		m.screen = ScreenGameOver
		// Delete the save game file since the game is over
		_ = config.DeleteSaveGame()
		// Clean up bot engine
		if m.botEngine != nil {
			_ = m.botEngine.Close()
			m.botEngine = nil
		}
	}

	return m, nil
}

// handleBotMoveError processes a bot move error.
// It displays the error message to the user and clears the thinking status.
func (m Model) handleBotMoveError(msg BotMoveErrorMsg) (tea.Model, tea.Cmd) {
	m.errorMsg = fmt.Sprintf("Bot error: %v", msg.err)
	m.statusMsg = ""
	return m, nil
}
