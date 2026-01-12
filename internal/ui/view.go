package ui

import (
	"fmt"
	"strings"

	"github.com/Mgrdich/TermChess/internal/engine"
	"github.com/charmbracelet/lipgloss"
)

// Define lipgloss styles for consistent UI rendering across the application.
// These styles use colors that work well on both light and dark terminals.
var (
	// titleStyle is used for the main application title
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Align(lipgloss.Center).
			Padding(1, 0)

	// menuItemStyle is used for regular (unselected) menu items
	menuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Padding(0, 2)

	// selectedItemStyle is used for the currently selected menu item
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true).
				Padding(0, 2)

	// helpStyle is used for help text and instructions
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Padding(1, 0)

	// errorStyle is used for error messages
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true).
			Padding(1, 0)

	// statusStyle is used for status messages
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Padding(1, 0)

	// cursorStyle is used for the cursor indicator
	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)
)

// renderHelpText conditionally renders help text based on config.
// Returns empty string if help text is disabled.
func renderHelpText(text string, config Config) string {
	if !config.ShowHelpText {
		return ""
	}
	return helpStyle.Render(text)
}

// View renders the UI based on the current model state.
// This function is called by Bubbletea on every update to generate
// the string that will be displayed in the terminal.
func (m Model) View() string {
	switch m.screen {
	case ScreenMainMenu:
		return m.renderMainMenu()
	case ScreenGameTypeSelect:
		return m.renderGameTypeSelect()
	case ScreenBotSelect:
		return "Bot Selection - Coming Soon"
	case ScreenFENInput:
		return m.renderFENInput()
	case ScreenGamePlay:
		return m.renderGamePlay()
	case ScreenGameOver:
		return m.renderGameOver()
	case ScreenSettings:
		return m.renderSettings()
	case ScreenSavePrompt:
		return m.renderSavePrompt()
	case ScreenResumePrompt:
		return m.renderResumePrompt()
	case ScreenDrawPrompt:
		return m.renderDrawPrompt()
	default:
		return "Unknown screen"
	}
}

// renderMainMenu renders the main menu screen with title, menu options,
// cursor indicator, help text, and any error or status messages.
func (m Model) renderMainMenu() string {
	var b strings.Builder

	// Render the application title
	title := titleStyle.Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render menu options with cursor indicator for selected item
	for i, option := range m.menuOptions {
		cursor := "  " // Two spaces for non-selected items
		optionText := option

		if i == m.menuSelection {
			// Highlight the selected item
			cursor = cursorStyle.Render("> ")
			optionText = selectedItemStyle.Render(option)
		} else {
			// Regular menu item styling
			optionText = menuItemStyle.Render(option)
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, optionText))
	}

	// Render help text
	helpText := renderHelpText("Use arrow keys to navigate, Enter to select, q to quit", m.config)
	if helpText != "" {
		b.WriteString("\n")
		b.WriteString(helpText)
	}

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := errorStyle.Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Render status message if present
	if m.statusMsg != "" {
		b.WriteString("\n\n")
		statusText := statusStyle.Render(m.statusMsg)
		b.WriteString(statusText)
	}

	return b.String()
}

// renderGameTypeSelect renders the GameTypeSelect screen with title, game type options,
// cursor indicator, help text, and any error or status messages.
func (m Model) renderGameTypeSelect() string {
	var b strings.Builder

	// Render the application title
	title := titleStyle.Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render screen header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(0, 0, 1, 0)
	header := headerStyle.Render("Select Game Type:")
	b.WriteString(header)
	b.WriteString("\n")

	// Render menu options with cursor indicator for selected item
	for i, option := range m.menuOptions {
		cursor := "  " // Two spaces for non-selected items
		optionText := option

		if i == m.menuSelection {
			// Highlight the selected item
			cursor = cursorStyle.Render("> ")
			optionText = selectedItemStyle.Render(option)
		} else {
			// Regular menu item styling
			optionText = menuItemStyle.Render(option)
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, optionText))
	}

	// Render help text
	helpText := renderHelpText("Use arrow keys to navigate, Enter to select, q to quit", m.config)
	if helpText != "" {
		b.WriteString("\n")
		b.WriteString(helpText)
	}

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := errorStyle.Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Render status message if present
	if m.statusMsg != "" {
		b.WriteString("\n\n")
		statusText := statusStyle.Render(m.statusMsg)
		b.WriteString(statusText)
	}

	return b.String()
}

// renderGameTypeSelect renders the game type selection screen.
// Displays options for Player vs Player and Player vs Bot with cursor indicator.
func (m Model) renderGameTypeSelect() string {
	var b strings.Builder

	// Render the application title
	title := titleStyle.Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render section header
	sectionHeader := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFDF5")).
		Padding(0, 2).
		Render("Select Game Type")
	b.WriteString(sectionHeader)
	b.WriteString("\n\n")

	// Render game type options with cursor indicator for selected item
	for i, option := range m.menuOptions {
		cursor := "  " // Two spaces for non-selected items
		optionText := option

		if i == m.menuSelection {
			// Highlight the selected item
			cursor = cursorStyle.Render("> ")
			optionText = selectedItemStyle.Render(option)
		} else {
			// Regular menu item styling
			optionText = menuItemStyle.Render(option)
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, optionText))
	}

	// Render help text
	b.WriteString("\n")
	helpText := helpStyle.Render("Use arrow keys to select, Enter to confirm, ESC to go back")
	b.WriteString(helpText)

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := errorStyle.Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Render status message if present
	if m.statusMsg != "" {
		b.WriteString("\n\n")
		statusText := statusStyle.Render(m.statusMsg)
		b.WriteString(statusText)
	}

	return b.String()
}

// renderGamePlay renders the GamePlay screen showing the chess board.
// Displays the title, board, turn indicator, input prompt, help text, and messages.
func (m Model) renderGamePlay() string {
	var b strings.Builder

	// Render the application title
	title := titleStyle.Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render the chess board
	renderer := NewBoardRenderer(m.config)
	boardStr := renderer.Render(m.board)
	b.WriteString(boardStr)

	// Render move history if enabled
	if m.config.ShowMoveHistory && len(m.moveHistory) > 0 {
		b.WriteString("\n\n")
		moveHistoryText := m.formatMoveHistory()
		moveHistoryStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Padding(0, 2)
		b.WriteString(moveHistoryStyle.Render(moveHistoryText))
	}

	// Render turn indicator
	b.WriteString("\n\n")
	turnText := "White to move"
	if m.board.ActiveColor == 1 { // Black
		turnText = "Black to move"
	}
	turnIndicator := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Render(turnText)
	b.WriteString(turnIndicator)

	// Render input prompt
	b.WriteString("\n\n")
	inputPrompt := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Render("Enter move: ")
	inputText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Render(m.input)
	b.WriteString(inputPrompt + inputText)

	// Add help text
	helpText := renderHelpText("Type move (e.g. e4, Nf3) | q: quit | Commands: resign, offerdraw, showfen, menu", m.config)
	if helpText != "" {
		b.WriteString("\n\n")
		b.WriteString(helpText)
	}

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := errorStyle.Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Render status message if present
	if m.statusMsg != "" {
		b.WriteString("\n\n")
		statusText := statusStyle.Render(m.statusMsg)
		b.WriteString(statusText)
	}

	// Render move history if enabled
	if m.config.ShowMoveHistory && len(m.moveHistory) > 0 {
		b.WriteString("\n\n")

		// Move history header
		historyHeaderStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA"))
		historyHeader := historyHeaderStyle.Render("Move History:")
		b.WriteString(historyHeader)
		b.WriteString("\n")

		// Format and display move history
		historyText := FormatMoveHistory(m.moveHistory)
		historyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))
		history := historyStyle.Render(historyText)
		b.WriteString(history)
		b.WriteString("\n")
	}

	return b.String()
}

// getGameResultMessage returns a human-readable message describing the game result.
// It analyzes the game status and winner to generate an appropriate message.
// If resignedBy is not -1, it indicates which player resigned.
// If drawByAgreement is true, the game ended by mutual agreement.
func getGameResultMessage(board *engine.Board, resignedBy int8, drawByAgreement bool) string {
	// Check for draw by agreement first
	if drawByAgreement {
		return "Draw by agreement"
	}

	// Check for resignation
	if resignedBy != -1 {
		if resignedBy == int8(engine.White) {
			return "White resigned - Black wins"
		}
		return "Black resigned - White wins"
	}

	// Otherwise, check the board status
	status := board.Status()

	switch status {
	case engine.Checkmate:
		winner, _ := board.Winner()
		if winner == engine.White {
			return "Checkmate! White wins"
		}
		return "Checkmate! Black wins"

	case engine.Stalemate:
		return "Stalemate - Draw"

	case engine.DrawThreefoldRepetition, engine.DrawFivefoldRepetition:
		return "Draw by repetition"

	case engine.DrawFiftyMoveRule:
		return "Draw by fifty-move rule"

	case engine.DrawSeventyFiveMoveRule:
		return "Draw by seventy-five-move rule"

	case engine.DrawInsufficientMaterial:
		return "Draw by insufficient material"

	default:
		return "Game Over"
	}
}

// renderGameOver renders the GameOver screen showing the game result,
// final board position, move count, and options to continue.
func (m Model) renderGameOver() string {
	var b strings.Builder

	// Render the application title
	title := titleStyle.Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render game result message
	resultMsg := getGameResultMessage(m.board, m.resignedBy, m.drawByAgreement)
	resultStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700")).
		Align(lipgloss.Center).
		Padding(1, 0)
	b.WriteString(resultStyle.Render(resultMsg))
	b.WriteString("\n\n")

	// Render the final board position
	renderer := NewBoardRenderer(m.config)
	boardStr := renderer.Render(m.board)
	b.WriteString(boardStr)

	// Render move count
	b.WriteString("\n\n")
	moveCountMsg := fmt.Sprintf("Game ended after %d moves", m.board.FullMoveNum)
	moveCountStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Align(lipgloss.Center)
	b.WriteString(moveCountStyle.Render(moveCountMsg))

	// Render options
	b.WriteString("\n\n")
	optionsText := "Press 'n' for New Game  |  Press 'm' for Main Menu  |  Press 'q' to Quit"
	optionsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Align(lipgloss.Center)
	b.WriteString(optionsStyle.Render(optionsText))

	// Render help text
	helpText := renderHelpText("n: new game | m: main menu | q: quit", m.config)
	if helpText != "" {
		b.WriteString("\n\n")
		b.WriteString(helpText)
	}

	return b.String()
}

// renderSettings renders the Settings screen showing display configuration options.
// Each option displays its current value and can be toggled by the user.
func (m Model) renderSettings() string {
	var b strings.Builder

	// Render the application title
	title := titleStyle.Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render screen header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(0, 0, 1, 0)
	header := headerStyle.Render("Settings")
	b.WriteString(header)
	b.WriteString("\n")

	// Define settings options with their current values
	// The order here determines the settingsSelection index
	settingsOptions := []struct {
		label   string
		enabled bool
	}{
		{"Use Unicode Pieces", m.config.UseUnicode},
		{"Show Coordinates", m.config.ShowCoords},
		{"Use Colors", m.config.UseColors},
		{"Show Move History", m.config.ShowMoveHistory},
		{"Show Help Text", m.config.ShowHelpText},
	}

	// Render each setting option with its current state
	for i, option := range settingsOptions {
		cursor := "  " // Two spaces for non-selected items

		// Determine checkbox state
		checkbox := "[ ]"
		if option.enabled {
			checkbox = "[X]"
		}

		// Build the option text
		optionText := fmt.Sprintf("%s %s", option.label, checkbox)

		if i == m.settingsSelection {
			// Highlight the selected item
			cursor = cursorStyle.Render("> ")
			optionText = selectedItemStyle.Render(optionText)
		} else {
			// Regular menu item styling
			optionText = menuItemStyle.Render(optionText)
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, optionText))
	}

	// Render help text
	helpText := renderHelpText("Use arrow keys to navigate, Enter to toggle, ESC/q to return to menu", m.config)
	if helpText != "" {
		b.WriteString("\n")
		b.WriteString(helpText)
	}

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := errorStyle.Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Render status message if present
	if m.statusMsg != "" {
		b.WriteString("\n\n")
		statusText := statusStyle.Render(m.statusMsg)
		b.WriteString(statusText)
	}

	return b.String()
}

// renderSavePrompt renders the save prompt screen when the user tries to exit during an active game.
// It shows the current board position and asks if they want to save before exiting.
func (m Model) renderSavePrompt() string {
	var b strings.Builder

	// Render the application title
	title := titleStyle.Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render the current board position
	if m.board != nil {
		renderer := NewBoardRenderer(m.config)
		boardStr := renderer.Render(m.board)
		b.WriteString(boardStr)
		b.WriteString("\n\n")
	}

	// Render the save prompt message
	promptStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700")).
		Align(lipgloss.Center).
		Padding(1, 0)
	promptMsg := "Save current game before exiting?"
	b.WriteString(promptStyle.Render(promptMsg))
	b.WriteString("\n\n")

	// Render options
	optionsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Align(lipgloss.Center)
	optionsText := "y: Yes  |  n: No  |  ESC: Cancel"
	b.WriteString(optionsStyle.Render(optionsText))

	// Render help text
	helpText := renderHelpText("y: save and exit | n: exit without saving | ESC: cancel", m.config)
	if helpText != "" {
		b.WriteString("\n\n")
		b.WriteString(helpText)
	}

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := errorStyle.Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	return b.String()
}

// renderResumePrompt renders the resume prompt screen when a saved game exists on startup.
// It asks the user if they want to resume the saved game or go to the main menu.
func (m Model) renderResumePrompt() string {
	var b strings.Builder

	// Render the application title
	title := titleStyle.Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render the resume prompt message
	promptStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700")).
		Align(lipgloss.Center).
		Padding(1, 0)
	promptMsg := "A saved game was found. Resume last game?"
	b.WriteString(promptStyle.Render(promptMsg))
	b.WriteString("\n\n")

	// Render options
	optionsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Align(lipgloss.Center)
	optionsText := "y: Yes  |  n: No"
	b.WriteString(optionsStyle.Render(optionsText))

	// Render help text
	helpText := renderHelpText("y: resume game | n: go to main menu", m.config)
	if helpText != "" {
		b.WriteString("\n\n")
		b.WriteString(helpText)
	}

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := errorStyle.Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	return b.String()
}

// renderFENInput renders the FEN input screen where users can load a chess position from FEN notation.
// Displays input field, instructions, example FEN, help text, and any error messages.
func (m Model) renderFENInput() string {
	var b strings.Builder

	// Render the application title
	title := titleStyle.Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render screen header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(0, 0, 1, 0)
	header := headerStyle.Render("Load Game from FEN")
	b.WriteString(header)
	b.WriteString("\n")

	// Instructions
	instructions := "Enter a FEN string to load a chess position:\n\n"
	b.WriteString(instructions)

	// Input field with cursor
	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true)
	inputLine := inputStyle.Render(m.fenInput + "█")
	b.WriteString(inputLine)
	b.WriteString("\n\n")

	// Example
	exampleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))
	example := exampleStyle.Render("Example: rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	b.WriteString(example)
	b.WriteString("\n\n")

	// Help text
	helpText := renderHelpText("Enter: load position | ESC: back to menu", m.config)
	if helpText != "" {
		b.WriteString(helpText)
	}

	// Error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := errorStyle.Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	return b.String()
}

// renderSettings renders the Settings screen showing all configurable display options.
// Each setting can be toggled on/off using Space or Enter keys.
// The selected setting is highlighted with a cursor indicator.
func (m Model) renderSettings() string {
	var b strings.Builder

	// Render the application title
	title := titleStyle.Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render section header
	sectionHeader := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFDF5")).
		Padding(0, 2).
		Render("Settings")
	b.WriteString(sectionHeader)
	b.WriteString("\n\n")

	// Define the settings with their current values
	settings := []struct {
		name  string
		value bool
	}{
		{"Use Unicode Pieces", m.config.UseUnicode},
		{"Show Coordinates", m.config.ShowCoords},
		{"Use Colors", m.config.UseColors},
		{"Show Move History", m.config.ShowMoveHistory},
	}

	// Render each setting with checkbox and cursor indicator
	for i, setting := range settings {
		cursor := "  " // Two spaces for non-selected items

		// Checkbox indicator
		checkbox := "[ ]"
		if setting.value {
			checkbox = "[X]"
		}

		// Format the option text
		optionText := fmt.Sprintf("%s: %s", setting.name, checkbox)

		if i == m.settingsSelection {
			// Highlight the selected item
			cursor = cursorStyle.Render("> ")
			optionText = selectedItemStyle.Render(optionText)
		} else {
			// Regular menu item styling
			optionText = menuItemStyle.Render(optionText)
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, optionText))
	}

	// Render help text
	b.WriteString("\n")
	helpText := helpStyle.Render("Use arrow keys to navigate, Space/Enter to toggle, ESC to go back")
	b.WriteString(helpText)

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := errorStyle.Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Render status message if present
	if m.statusMsg != "" {
		b.WriteString("\n\n")
		statusText := statusStyle.Render(m.statusMsg)
		b.WriteString(statusText)
	}

	return b.String()
}

// renderSavePrompt renders the Save Prompt screen asking the user if they want to save the game.
// Displays a title, message, two options (Yes/No), and help text.
func (m Model) renderSavePrompt() string {
	var b strings.Builder

	// Render the application title
	title := titleStyle.Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render prompt title
	promptTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700")).
		Align(lipgloss.Center).
		Padding(1, 0).
		Render("Save Game?")
	b.WriteString(promptTitle)
	b.WriteString("\n\n")

	// Render prompt message
	promptMessage := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Padding(0, 2).
		Render("Save current game before exiting?")
	b.WriteString(promptMessage)
	b.WriteString("\n\n")

	// Define the save prompt options
	options := []string{"Yes", "No"}

	// Render each option with cursor indicator
	for i, option := range options {
		cursor := "  " // Two spaces for non-selected items
		optionText := option

		if i == m.savePromptSelection {
			// Highlight the selected item
			cursor = cursorStyle.Render("> ")
			optionText = selectedItemStyle.Render(option)
		} else {
			// Regular menu item styling
			optionText = menuItemStyle.Render(option)
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, optionText))
	}

	// Render help text
	b.WriteString("\n")
	helpText := helpStyle.Render("Use arrow keys to select, Enter to confirm, ESC to cancel")
	b.WriteString(helpText)

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := errorStyle.Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Render status message if present
	if m.statusMsg != "" {
		b.WriteString("\n\n")
		statusText := statusStyle.Render(m.statusMsg)
		b.WriteString(statusText)
	}

	return b.String()
}

// renderResumePrompt renders the Resume Prompt screen asking the user if they want to resume a saved game.
// Displays a title, message, two options (Yes/No), and help text.
func (m Model) renderResumePrompt() string {
	var b strings.Builder

	// Render the application title
	title := titleStyle.Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render prompt title
	promptTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#50FA7B")).
		Align(lipgloss.Center).
		Padding(1, 0).
		Render("Saved Game Found")
	b.WriteString(promptTitle)
	b.WriteString("\n\n")

	// Render prompt message
	promptMessage := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Padding(0, 2).
		Render("Resume last game?")
	b.WriteString(promptMessage)
	b.WriteString("\n\n")

	// Define the resume prompt options
	options := []string{"Yes", "No"}

	// Render each option with cursor indicator
	for i, option := range options {
		cursor := "  " // Two spaces for non-selected items
		optionText := option

		if i == m.resumePromptSelection {
			// Highlight the selected item
			cursor = cursorStyle.Render("> ")
			optionText = selectedItemStyle.Render(option)
		} else {
			// Regular menu item styling
			optionText = menuItemStyle.Render(option)
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, optionText))
	}

	// Render help text
	b.WriteString("\n")
	helpText := helpStyle.Render("Use arrow keys to select, Enter to confirm")
	b.WriteString(helpText)

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := errorStyle.Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Render status message if present
	if m.statusMsg != "" {
		b.WriteString("\n\n")
		statusText := statusStyle.Render(m.statusMsg)
		b.WriteString(statusText)
	}

	return b.String()
}

// renderFENInput renders the FEN Input screen allowing the user to enter a FEN string.
// Displays a title, instructions, text input field, example FEN strings, and help text.
func (m Model) renderFENInput() string {
	var b strings.Builder

	// Render the application title
	title := titleStyle.Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render section header
	sectionHeader := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFDF5")).
		Padding(0, 2).
		Render("Load Game from FEN")
	b.WriteString(sectionHeader)
	b.WriteString("\n\n")

	// Render instructions
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Padding(0, 2).
		Render("Enter a FEN string to load a chess position:")
	b.WriteString(instructions)
	b.WriteString("\n\n")

	// Render the text input field
	b.WriteString("  ")
	b.WriteString(m.fenInput.View())
	b.WriteString("\n\n")

	// Render example FEN strings
	examplesTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Padding(0, 2).
		Render("Example FEN strings:")
	b.WriteString(examplesTitle)
	b.WriteString("\n")

	examples := []string{
		"Starting position:",
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		"",
		"Mid-game position:",
		"r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4",
	}

	exampleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Padding(0, 2)

	for _, example := range examples {
		b.WriteString(exampleStyle.Render(example))
		b.WriteString("\n")
	}

	// Render help text
	b.WriteString("\n")
	helpText := helpStyle.Render("Press Enter to load, ESC to go back")
	b.WriteString(helpText)

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := errorStyle.Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Render status message if present
	if m.statusMsg != "" {
		b.WriteString("\n\n")
		statusText := statusStyle.Render(m.statusMsg)
		b.WriteString(statusText)
	}

	return b.String()
}

// formatMoveHistory formats the move history as numbered move pairs.
// Format: "1. e4 e5 2. Nf3 Nc6 3. Bc4"
// The function reconstructs the board state for each move to format it correctly.
func (m Model) formatMoveHistory() string {
	if len(m.moveHistory) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString("Move History: ")

	// Create a fresh board to replay moves
	board := engine.NewBoard()

	for i, move := range m.moveHistory {
		// Calculate move number (starts at 1)
		moveNum := i/2 + 1
		isWhiteMove := i%2 == 0

		// Add move number before White's move
		if isWhiteMove {
			result.WriteString(fmt.Sprintf("%d. ", moveNum))
		}

		// Format the move in SAN notation
		san := FormatSAN(board, move)
		result.WriteString(san)

		// Add space after each move (except the last one)
		if i < len(m.moveHistory)-1 {
			result.WriteString(" ")
		}

		// Apply the move to the board for the next iteration
		board.MakeMove(move)
	}

	return result.String()
}

// renderDrawPrompt renders the Draw Prompt screen asking the opponent to accept or decline a draw offer.
// Displays a title, message indicating which player offered the draw, two options (Accept/Decline), and help text.
func (m Model) renderDrawPrompt() string {
	var b strings.Builder

	// Render the application title
	title := titleStyle.Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render prompt title
	promptTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700")).
		Align(lipgloss.Center).
		Padding(1, 0).
		Render("Draw Offer")
	b.WriteString(promptTitle)
	b.WriteString("\n\n")

	// Render prompt message based on who offered the draw
	offerMessage := "White offers a draw. Accept?"
	if m.drawOfferedBy == int8(engine.Black) {
		offerMessage = "Black offers a draw. Accept?"
	}
	promptMessage := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Padding(0, 2).
		Render(offerMessage)
	b.WriteString(promptMessage)
	b.WriteString("\n\n")

	// Define the draw prompt options
	options := []string{"Accept", "Decline"}

	// Render each option with cursor indicator
	for i, option := range options {
		cursor := "  " // Two spaces for non-selected items
		optionText := option

		if i == m.drawPromptSelection {
			// Highlight the selected item
			cursor = cursorStyle.Render("> ")
			optionText = selectedItemStyle.Render(option)
		} else {
			// Regular menu item styling
			optionText = menuItemStyle.Render(option)
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, optionText))
	}

	// Render help text
	b.WriteString("\n")
	helpText := helpStyle.Render("Use arrow keys to select, Enter to confirm, ESC to cancel")
	b.WriteString(helpText)

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := errorStyle.Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Render status message if present
	if m.statusMsg != "" {
		b.WriteString("\n\n")
		statusText := statusStyle.Render(m.statusMsg)
		b.WriteString(statusText)
	}

	return b.String()
}
