package ui

import (
	"fmt"
	"strings"

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
		return "Game Type Selection - Coming Soon"
	case ScreenBotSelect:
		return "Bot Selection - Coming Soon"
	case ScreenFENInput:
		return "FEN Input - Coming Soon"
	case ScreenGamePlay:
		return m.renderGamePlay()
	case ScreenGameOver:
		return "Game Over - Coming Soon"
	case ScreenSettings:
		return "Settings - Coming Soon"
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

// renderGamePlay renders the GamePlay screen showing the chess board.
// For Slice 2, this displays just the board. Turn indicators and move history
// will be added in later slices.
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
