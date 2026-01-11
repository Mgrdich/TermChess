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
		return "FEN Input - Coming Soon"
	case ScreenGamePlay:
		return m.renderGamePlay()
	case ScreenGameOver:
		return m.renderGameOver()
	case ScreenSettings:
		return m.renderSettings()
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
	b.WriteString("\n")
	helpText := helpStyle.Render("Use arrow keys to navigate, Enter to select, q to quit")
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
	b.WriteString("\n")
	helpText := helpStyle.Render("Use arrow keys to navigate, Enter to select, q to quit")
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
	b.WriteString("\n\n")
	helpText := helpStyle.Render("Press q to quit")
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

// getGameResultMessage returns a human-readable message describing the game result.
// It analyzes the game status and winner to generate an appropriate message.
func getGameResultMessage(board *engine.Board) string {
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
	resultMsg := getGameResultMessage(m.board)
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
	b.WriteString("\n")
	helpText := helpStyle.Render("Use arrow keys to navigate, Enter to toggle, ESC/q to return to menu")
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
