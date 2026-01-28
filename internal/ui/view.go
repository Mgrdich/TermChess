package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Mgrdich/TermChess/internal/bvb"
	"github.com/Mgrdich/TermChess/internal/engine"
	"github.com/charmbracelet/lipgloss"
)

// Style helper methods that use the theme colors.
// These methods return lipgloss styles based on the model's current theme.

// titleStyle returns the style for the main application title.
func (m Model) titleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.TitleText).
		Align(lipgloss.Center).
		Padding(1, 0)
}

// menuItemStyle returns the style for regular (unselected) menu items.
func (m Model) menuItemStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.theme.MenuNormal).
		Padding(0, 2)
}

// selectedItemStyle returns the style for the currently selected menu item.
func (m Model) selectedItemStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.theme.MenuSelected).
		Bold(true).
		Padding(0, 2)
}

// helpStyle returns the style for help text and instructions.
func (m Model) helpStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.theme.HelpText).
		Padding(1, 0)
}

// errorStyle returns the style for error messages.
func (m Model) errorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.theme.ErrorText).
		Bold(true).
		Padding(1, 0)
}

// statusStyle returns the style for status messages.
func (m Model) statusStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.theme.StatusText).
		Padding(1, 0)
}

// cursorStyle returns the style for the cursor indicator.
func (m Model) cursorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.theme.MenuSelected).
		Bold(true)
}

// renderHelpText conditionally renders help text based on config.
// Returns empty string if help text is disabled.
func (m Model) renderHelpText(text string) string {
	if !m.config.ShowHelpText {
		return ""
	}
	return m.helpStyle().Render(text)
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
		return m.renderBotSelect()
	case ScreenColorSelect:
		return m.renderColorSelect()
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
	case ScreenBvBBotSelect:
		return m.renderBvBBotSelect()
	case ScreenBvBGameMode:
		return m.renderBvBGameMode()
	case ScreenBvBGridConfig:
		return m.renderBvBGridConfig()
	case ScreenBvBGamePlay:
		return m.renderBvBGamePlay()
	case ScreenBvBStats:
		return m.renderBvBStats()
	default:
		return "Unknown screen"
	}
}

// renderMainMenu renders the main menu screen with title, menu options,
// cursor indicator, help text, and any error or status messages.
// The "Resume Game" option (if present) is visually distinct with a special indicator and color.
func (m Model) renderMainMenu() string {
	var b strings.Builder

	// Render the application title
	title := m.titleStyle().Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render menu options with cursor indicator for selected item
	for i, option := range m.menuOptions {
		cursor := "  " // Two spaces for non-selected items
		optionText := option

		// Check if this is the "Resume Game" option
		isResumeGame := option == "Resume Game"

		if i == m.menuSelection {
			// Highlight the selected item
			if isResumeGame {
				// Special styling for selected Resume Game option
				cursor = m.cursorStyle().Render("▶ ")
				resumeStyle := lipgloss.NewStyle().
					Foreground(m.theme.StatusText).
					Bold(true).
					Padding(0, 2)
				optionText = resumeStyle.Render(option)
			} else {
				cursor = m.cursorStyle().Render("> ")
				optionText = m.selectedItemStyle().Render(option)
			}
		} else {
			// Regular menu item styling
			if isResumeGame {
				// Special styling for unselected Resume Game option
				resumeStyle := lipgloss.NewStyle().
					Foreground(m.theme.StatusText).
					Padding(0, 2)
				optionText = resumeStyle.Render("▶ " + option)
				cursor = "" // No cursor needed, indicator is part of the text
			} else {
				optionText = m.menuItemStyle().Render(option)
			}
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, optionText))
	}

	// Render help text
	helpText := m.renderHelpText("arrows/jk: navigate | enter: select | q: quit")
	if helpText != "" {
		b.WriteString("\n")
		b.WriteString(helpText)
	}

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := m.errorStyle().Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Render status message if present
	if m.statusMsg != "" {
		b.WriteString("\n\n")
		statusText := m.statusStyle().Render(m.statusMsg)
		b.WriteString(statusText)
	}

	return b.String()
}

// renderGameTypeSelect renders the GameTypeSelect screen with title, game type options,
// cursor indicator, help text, and any error or status messages.
func (m Model) renderGameTypeSelect() string {
	var b strings.Builder

	// Render the application title
	title := m.titleStyle().Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render screen header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.TitleText).
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
			cursor = m.cursorStyle().Render("> ")
			optionText = m.selectedItemStyle().Render(option)
		} else {
			// Regular menu item styling
			optionText = m.menuItemStyle().Render(option)
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, optionText))
	}

	// Render help text
	helpText := m.renderHelpText("ESC: back to menu | arrows/jk: navigate | enter: select")
	if helpText != "" {
		b.WriteString("\n")
		b.WriteString(helpText)
	}

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := m.errorStyle().Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Render status message if present
	if m.statusMsg != "" {
		b.WriteString("\n\n")
		statusText := m.statusStyle().Render(m.statusMsg)
		b.WriteString(statusText)
	}

	return b.String()
}

// renderBotSelect renders the BotSelect screen with title, bot difficulty options,
// cursor indicator, help text, and any error or status messages.
func (m Model) renderBotSelect() string {
	var b strings.Builder

	// Render the application title
	title := m.titleStyle().Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render screen header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.TitleText).
		Padding(0, 0, 1, 0)
	header := headerStyle.Render("Select Bot Difficulty:")
	b.WriteString(header)
	b.WriteString("\n")

	// Render menu options with cursor indicator for selected item
	for i, option := range m.menuOptions {
		cursor := "  " // Two spaces for non-selected items
		optionText := option

		if i == m.menuSelection {
			// Highlight the selected item
			cursor = m.cursorStyle().Render("> ")
			optionText = m.selectedItemStyle().Render(option)
		} else {
			// Regular menu item styling
			optionText = m.menuItemStyle().Render(option)
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, optionText))
	}

	// Render help text
	helpText := m.renderHelpText("ESC: back to game type | arrows/jk: navigate | enter: select")
	if helpText != "" {
		b.WriteString("\n")
		b.WriteString(helpText)
	}

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := m.errorStyle().Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Render status message if present
	if m.statusMsg != "" {
		b.WriteString("\n\n")
		statusText := m.statusStyle().Render(m.statusMsg)
		b.WriteString(statusText)
	}

	return b.String()
}

// renderColorSelect renders the ColorSelect screen with title, color options,
// cursor indicator, help text, and any error or status messages.
func (m Model) renderColorSelect() string {
	var b strings.Builder

	// Render the application title
	title := m.titleStyle().Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render screen header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.TitleText).
		Padding(0, 0, 1, 0)
	header := headerStyle.Render("Select Your Color:")
	b.WriteString(header)
	b.WriteString("\n")

	// Render menu options with cursor indicator for selected item
	for i, option := range m.menuOptions {
		cursor := "  " // Two spaces for non-selected items
		optionText := option

		if i == m.menuSelection {
			// Highlight the selected item
			cursor = m.cursorStyle().Render("> ")
			optionText = m.selectedItemStyle().Render(option)
		} else {
			// Regular menu item styling
			optionText = m.menuItemStyle().Render(option)
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, optionText))
	}

	// Render help text
	helpText := m.renderHelpText("ESC: back to difficulty | arrows/jk: navigate | enter: select")
	if helpText != "" {
		b.WriteString("\n")
		b.WriteString(helpText)
	}

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := m.errorStyle().Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Render status message if present
	if m.statusMsg != "" {
		b.WriteString("\n\n")
		statusText := m.statusStyle().Render(m.statusMsg)
		b.WriteString(statusText)
	}

	return b.String()
}

// renderGamePlay renders the GamePlay screen showing the chess board.
// Displays the title, board, turn indicator, input prompt, help text, and messages.
func (m Model) renderGamePlay() string {
	var b strings.Builder

	// Render the application title
	title := m.titleStyle().Render("TermChess")
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
			Foreground(m.theme.MenuNormal).
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
		Foreground(m.theme.MenuSelected).
		Render(turnText)
	b.WriteString(turnIndicator)

	// Render input prompt
	b.WriteString("\n\n")
	inputPrompt := lipgloss.NewStyle().
		Foreground(m.theme.MenuNormal).
		Render("Enter move: ")
	inputText := lipgloss.NewStyle().
		Foreground(m.theme.MenuSelected).
		Render(m.input)
	b.WriteString(inputPrompt + inputText)

	// Add help text
	helpText := m.renderHelpText("ESC: menu (with save) | type move (e.g. e4, Nf3) | Commands: resign, offerdraw, showfen, menu")
	if helpText != "" {
		b.WriteString("\n\n")
		b.WriteString(helpText)
	}

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := m.errorStyle().Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Render status message if present
	if m.statusMsg != "" {
		b.WriteString("\n\n")
		statusText := m.statusStyle().Render(m.statusMsg)
		b.WriteString(statusText)
	}

	// Render move history if enabled
	if m.config.ShowMoveHistory && len(m.moveHistory) > 0 {
		b.WriteString("\n\n")

		// Move history header
		historyHeaderStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(m.theme.TitleText)
		historyHeader := historyHeaderStyle.Render("Move History:")
		b.WriteString(historyHeader)
		b.WriteString("\n")

		// Format and display move history
		historyText := FormatMoveHistory(m.moveHistory)
		historyStyle := lipgloss.NewStyle().
			Foreground(m.theme.MenuSelected)
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
	title := m.titleStyle().Render("TermChess")
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
		Foreground(m.theme.MenuNormal).
		Align(lipgloss.Center)
	b.WriteString(moveCountStyle.Render(moveCountMsg))

	// Render options
	b.WriteString("\n\n")
	optionsText := "Press 'n' for New Game  |  Press 'm' for Main Menu  |  Press 'q' to Quit"
	optionsStyle := lipgloss.NewStyle().
		Foreground(m.theme.MenuSelected).
		Align(lipgloss.Center)
	b.WriteString(optionsStyle.Render(optionsText))

	// Render help text
	helpText := m.renderHelpText("ESC/m: menu | n: new game | q: quit")
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
	title := m.titleStyle().Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render screen header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.TitleText).
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
			cursor = m.cursorStyle().Render("> ")
			optionText = m.selectedItemStyle().Render(optionText)
		} else {
			// Regular menu item styling
			optionText = m.menuItemStyle().Render(optionText)
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, optionText))
	}

	// Render help text
	helpText := m.renderHelpText("ESC: back | arrows/jk: navigate | enter/space: toggle")
	if helpText != "" {
		b.WriteString("\n")
		b.WriteString(helpText)
	}

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := m.errorStyle().Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Render status message if present
	if m.statusMsg != "" {
		b.WriteString("\n\n")
		statusText := m.statusStyle().Render(m.statusMsg)
		b.WriteString(statusText)
	}

	return b.String()
}

// renderSavePrompt renders the save prompt screen when the user tries to exit during an active game.
// It shows the current board position and asks if they want to save before exiting.
func (m Model) renderSavePrompt() string {
	var b strings.Builder

	// Render the application title
	title := m.titleStyle().Render("TermChess")
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
		Foreground(m.theme.MenuSelected).
		Align(lipgloss.Center)
	optionsText := "y: Yes  |  n: No  |  ESC: Cancel"
	b.WriteString(optionsStyle.Render(optionsText))

	// Render help text
	helpText := m.renderHelpText("y: save and exit | n: exit without saving | ESC: cancel")
	if helpText != "" {
		b.WriteString("\n\n")
		b.WriteString(helpText)
	}

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := m.errorStyle().Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	return b.String()
}

// renderResumePrompt renders the resume prompt screen when a saved game exists on startup.
// It asks the user if they want to resume the saved game or go to the main menu.
func (m Model) renderResumePrompt() string {
	var b strings.Builder

	// Render the application title
	title := m.titleStyle().Render("TermChess")
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
		Foreground(m.theme.MenuSelected).
		Align(lipgloss.Center)
	optionsText := "y: Yes  |  n: No"
	b.WriteString(optionsStyle.Render(optionsText))

	// Render help text
	helpText := m.renderHelpText("y: resume game | n: go to main menu")
	if helpText != "" {
		b.WriteString("\n\n")
		b.WriteString(helpText)
	}

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := m.errorStyle().Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	return b.String()
}

// renderFENInput renders the FEN input screen where users can load a chess position from FEN notation.
// Displays input field, instructions, example FEN, help text, and any error messages.
func (m Model) renderFENInput() string {
	var b strings.Builder

	// Render the application title
	title := m.titleStyle().Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render screen header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.TitleText).
		Padding(0, 0, 1, 0)
	header := headerStyle.Render("Load Game from FEN")
	b.WriteString(header)
	b.WriteString("\n")

	// Instructions
	instructions := "Enter a FEN string to load a chess position:\n\n"
	b.WriteString(instructions)

	// Input field with cursor
	// Render the text input component
	b.WriteString(m.fenInput.View())
	b.WriteString("\n\n")

	// Example
	exampleStyle := lipgloss.NewStyle().
		Foreground(m.theme.HelpText)
	example := exampleStyle.Render("Example: rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	b.WriteString(example)
	b.WriteString("\n\n")

	// Help text
	helpText := m.renderHelpText("ESC: back to menu | enter: load position")
	if helpText != "" {
		b.WriteString(helpText)
	}

	// Error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := m.errorStyle().Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	return b.String()
}

// renderDrawPrompt renders the Draw Prompt screen asking the opponent to accept or decline a draw offer.
// Displays a title, message indicating which player offered the draw, two options (Accept/Decline), and help text.
func (m Model) renderDrawPrompt() string {
	var b strings.Builder

	// Render the application title
	title := m.titleStyle().Render("TermChess")
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
		Foreground(m.theme.MenuNormal).
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
			cursor = m.cursorStyle().Render("> ")
			optionText = m.selectedItemStyle().Render(option)
		} else {
			// Regular menu item styling
			optionText = m.menuItemStyle().Render(option)
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, optionText))
	}

	// Render help text
	b.WriteString("\n")
	helpText := m.helpStyle().Render("Use arrow keys to select, Enter to confirm, ESC to cancel")
	b.WriteString(helpText)

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := m.errorStyle().Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Render status message if present
	if m.statusMsg != "" {
		b.WriteString("\n\n")
		statusText := m.statusStyle().Render(m.statusMsg)
		b.WriteString(statusText)
	}

	return b.String()
}

// renderBvBBotSelect renders the Bot vs Bot bot selection screen.
// This screen is shown twice: once for White bot difficulty, once for Black.
func (m Model) renderBvBBotSelect() string {
	var b strings.Builder

	// Render the application title
	title := m.titleStyle().Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render screen header based on which bot we're selecting
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.TitleText).
		Padding(0, 0, 1, 0)

	headerText := "Select White Bot Difficulty:"
	if !m.bvbSelectingWhite {
		headerText = "Select Black Bot Difficulty:"
	}
	header := headerStyle.Render(headerText)
	b.WriteString(header)
	b.WriteString("\n")

	// Render menu options with cursor indicator for selected item
	for i, option := range m.menuOptions {
		cursor := "  "
		optionText := option

		if i == m.menuSelection {
			cursor = m.cursorStyle().Render("> ")
			optionText = m.selectedItemStyle().Render(option)
		} else {
			optionText = m.menuItemStyle().Render(option)
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, optionText))
	}

	// Show the already-selected White difficulty when selecting Black
	if !m.bvbSelectingWhite {
		b.WriteString("\n")
		infoStyle := lipgloss.NewStyle().
			Foreground(m.theme.StatusText).
			Padding(0, 2)
		diffName := botDifficultyName(m.bvbWhiteDiff)
		b.WriteString(infoStyle.Render(fmt.Sprintf("White: %s Bot", diffName)))
		b.WriteString("\n")
	}

	// Render help text
	helpText := m.renderHelpText("ESC: back | arrows/jk: navigate | enter: select")
	if helpText != "" {
		b.WriteString("\n")
		b.WriteString(helpText)
	}

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := m.errorStyle().Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Render status message if present
	if m.statusMsg != "" {
		b.WriteString("\n\n")
		statusText := m.statusStyle().Render(m.statusMsg)
		b.WriteString(statusText)
	}

	return b.String()
}

// renderBvBGameMode renders the Bot vs Bot game mode selection screen.
// Shows Single Game / Multi-Game options, or a text input for game count.
func (m Model) renderBvBGameMode() string {
	var b strings.Builder

	// Render the application title
	title := m.titleStyle().Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Render screen header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.TitleText).
		Padding(0, 0, 1, 0)
	header := headerStyle.Render("Select Game Mode:")
	b.WriteString(header)
	b.WriteString("\n")

	// Show matchup info
	infoStyle := lipgloss.NewStyle().
		Foreground(m.theme.StatusText).
		Padding(0, 2)
	matchup := fmt.Sprintf("%s Bot (White) vs %s Bot (Black)",
		botDifficultyName(m.bvbWhiteDiff), botDifficultyName(m.bvbBlackDiff))
	b.WriteString(infoStyle.Render(matchup))
	b.WriteString("\n\n")

	if m.bvbInputtingCount {
		// Show text input for game count
		promptStyle := lipgloss.NewStyle().
			Foreground(m.theme.MenuNormal).
			Padding(0, 2)
		b.WriteString(promptStyle.Render("Number of games:"))
		b.WriteString("\n\n")

		inputStyle := lipgloss.NewStyle().
			Foreground(m.theme.MenuSelected).
			Padding(0, 2)
		inputDisplay := m.bvbCountInput
		if inputDisplay == "" {
			inputDisplay = "_"
		}
		b.WriteString(inputStyle.Render("> " + inputDisplay))
		b.WriteString("\n")

		helpText := m.renderHelpText("ESC: back | enter: confirm | type number")
		if helpText != "" {
			b.WriteString("\n")
			b.WriteString(helpText)
		}
	} else {
		// Show menu options
		for i, option := range m.menuOptions {
			cursor := "  "
			optionText := option

			if i == m.menuSelection {
				cursor = m.cursorStyle().Render("> ")
				optionText = m.selectedItemStyle().Render(option)
			} else {
				optionText = m.menuItemStyle().Render(option)
			}

			b.WriteString(fmt.Sprintf("%s%s\n", cursor, optionText))
		}

		helpText := m.renderHelpText("ESC: back | arrows/jk: navigate | enter: select")
		if helpText != "" {
			b.WriteString("\n")
			b.WriteString(helpText)
		}
	}

	// Render error message if present
	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := m.errorStyle().Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Render status message if present
	if m.statusMsg != "" {
		b.WriteString("\n\n")
		statusText := m.statusStyle().Render(m.statusMsg)
		b.WriteString(statusText)
	}

	return b.String()
}

// renderBvBGridView renders multiple games in a grid layout.
func (m Model) renderBvBGridView() string {
	var b strings.Builder

	title := m.titleStyle().Render("TermChess - Bot vs Bot")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Check terminal size - each cell needs ~14 width and ~11 height
	minWidth := m.bvbGridCols * 14
	minHeight := m.bvbGridRows*11 + 8 // 8 lines for header/footer
	if m.termWidth > 0 && m.termHeight > 0 && (m.termWidth < minWidth || m.termHeight < minHeight) {
		warnStyle := lipgloss.NewStyle().
			Foreground(m.theme.ErrorText).
			Padding(0, 2)
		b.WriteString(warnStyle.Render(fmt.Sprintf("Terminal too small for %dx%d grid (need %dx%d, have %dx%d)",
			m.bvbGridRows, m.bvbGridCols, minWidth, minHeight, m.termWidth, m.termHeight)))
		b.WriteString("\n")
		b.WriteString(warnStyle.Render("Press Tab to switch to single-board view"))
		b.WriteString("\n")
		return b.String()
	}

	sessions := m.bvbManager.Sessions()
	if len(sessions) == 0 {
		b.WriteString("No games available.\n")
		return b.String()
	}

	// Calculate pagination
	boardsPerPage := m.bvbGridRows * m.bvbGridCols
	totalPages := (len(sessions) + boardsPerPage - 1) / boardsPerPage
	pageIdx := m.bvbPageIndex
	if pageIdx >= totalPages {
		pageIdx = totalPages - 1
	}
	startIdx := pageIdx * boardsPerPage
	endIdx := startIdx + boardsPerPage
	if endIdx > len(sessions) {
		endIdx = len(sessions)
	}

	// Show matchup and progress info
	infoStyle := lipgloss.NewStyle().
		Foreground(m.theme.StatusText).
		Padding(0, 2)

	finished := 0
	for _, s := range sessions {
		if s.IsFinished() {
			finished++
		}
	}
	running := m.bvbManager.RunningCount()
	queued := m.bvbManager.QueuedCount()
	matchup := fmt.Sprintf("%s Bot (White) vs %s Bot (Black) | Completed: %d/%d | Running: %d | Queued: %d",
		botDifficultyName(m.bvbWhiteDiff), botDifficultyName(m.bvbBlackDiff),
		finished, len(sessions), running, queued)
	b.WriteString(infoStyle.Render(matchup))
	b.WriteString("\n\n")

	// Render the grid
	pageSessions := sessions[startIdx:endIdx]
	gridStr := m.renderBoardGrid(pageSessions, m.bvbGridCols)
	b.WriteString(gridStr)
	b.WriteString("\n")

	// Page indicator
	if totalPages > 1 {
		pageInfo := fmt.Sprintf("Page %d/%d", pageIdx+1, totalPages)
		pageStyle := lipgloss.NewStyle().
			Foreground(m.theme.MenuSelected).
			Bold(true).
			Padding(0, 2)
		b.WriteString(pageStyle.Render(pageInfo))
		b.WriteString("\n")
	}

	// Speed/pause status
	speedNames := map[bvb.PlaybackSpeed]string{
		bvb.SpeedInstant: "Instant",
		bvb.SpeedFast:    "Fast",
		bvb.SpeedNormal:  "Normal",
		bvb.SpeedSlow:    "Slow",
	}
	controlStatus := fmt.Sprintf("Speed: %s", speedNames[m.bvbSpeed])
	if m.bvbPaused {
		controlStatus += " | PAUSED"
	}
	controlStyle := lipgloss.NewStyle().
		Foreground(m.theme.MenuNormal).
		Padding(0, 2)
	b.WriteString(controlStyle.Render(controlStatus))
	b.WriteString("\n")

	// Help text
	helpText := m.renderHelpText("Space: pause/resume | 1-4: speed | ←/→: pages | Tab: single view | f: FEN | ESC: abort")
	if helpText != "" {
		b.WriteString("\n")
		b.WriteString(helpText)
	}

	return b.String()
}

// renderBoardGrid renders a slice of sessions as a grid with the given number of columns.
func (m Model) renderBoardGrid(sessions []*bvb.GameSession, cols int) string {
	if len(sessions) == 0 {
		return ""
	}

	// Render each session as a compact board cell
	cells := make([]string, len(sessions))
	for i, session := range sessions {
		cells[i] = m.renderCompactBoardCell(session)
	}

	// Arrange cells into rows
	var rows []string
	for i := 0; i < len(cells); i += cols {
		end := i + cols
		if end > len(cells) {
			end = len(cells)
		}
		rowCells := cells[i:end]
		row := lipgloss.JoinHorizontal(lipgloss.Top, rowCells...)
		rows = append(rows, row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// renderCompactBoardCell renders a single game session as a compact board cell for the grid.
// Shows: game number, compact board, move count, and status.
func (m Model) renderCompactBoardCell(session *bvb.GameSession) string {
	var b strings.Builder

	board := session.CurrentBoard()
	gameNum := session.GameNumber()
	moveCount := len(session.CurrentMoveHistory())
	isFinished := session.IsFinished()

	// Game header
	headerText := fmt.Sprintf("Game %d", gameNum)
	b.WriteString(headerText)
	b.WriteString("\n")

	// Render compact board (no coords, no color)
	compactConfig := Config{
		UseUnicode: m.config.UseUnicode,
		ShowCoords: false,
		UseColors:  false,
	}
	renderer := NewBoardRenderer(compactConfig)
	boardStr := renderer.Render(board)
	b.WriteString(boardStr)
	b.WriteString("\n")

	// Status line
	if isFinished {
		result := session.Result()
		if result != nil {
			statusText := fmt.Sprintf("Moves: %d | %s", moveCount, result.Winner)
			b.WriteString(statusText)
		}
	} else {
		b.WriteString(fmt.Sprintf("Moves: %d", moveCount))
	}
	b.WriteString("\n")

	// Style the cell with border and padding
	cellStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Margin(0, 1)

	if isFinished {
		// Dimmed style for finished games
		cellStyle = cellStyle.Foreground(m.theme.HelpText)
	}

	return cellStyle.Render(b.String())
}

// renderBvBGridConfig renders the Bot vs Bot grid configuration screen.
func (m Model) renderBvBGridConfig() string {
	var b strings.Builder

	title := m.titleStyle().Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.TitleText).
		Padding(0, 0, 1, 0)
	header := headerStyle.Render("Select Grid Layout:")
	b.WriteString(header)
	b.WriteString("\n")

	// Show game count info
	infoStyle := lipgloss.NewStyle().
		Foreground(m.theme.StatusText).
		Padding(0, 2)
	gameInfo := fmt.Sprintf("%d game(s) | %s Bot (White) vs %s Bot (Black)",
		m.bvbGameCount, botDifficultyName(m.bvbWhiteDiff), botDifficultyName(m.bvbBlackDiff))
	b.WriteString(infoStyle.Render(gameInfo))
	b.WriteString("\n\n")

	if m.bvbInputtingGrid {
		promptStyle := lipgloss.NewStyle().
			Foreground(m.theme.MenuNormal).
			Padding(0, 2)
		b.WriteString(promptStyle.Render("Enter grid dimensions (RxC, max 8 total):"))
		b.WriteString("\n\n")

		inputStyle := lipgloss.NewStyle().
			Foreground(m.theme.MenuSelected).
			Padding(0, 2)
		inputDisplay := m.bvbCustomGridInput
		if inputDisplay == "" {
			inputDisplay = "_"
		}
		b.WriteString(inputStyle.Render("> " + inputDisplay))
		b.WriteString("\n")

		helpText := m.renderHelpText("ESC: back | enter: confirm | e.g. 2x3")
		if helpText != "" {
			b.WriteString("\n")
			b.WriteString(helpText)
		}
	} else {
		for i, option := range m.menuOptions {
			cursor := "  "
			optionText := option

			if i == m.menuSelection {
				cursor = m.cursorStyle().Render("> ")
				optionText = m.selectedItemStyle().Render(option)
			} else {
				optionText = m.menuItemStyle().Render(option)
			}

			b.WriteString(fmt.Sprintf("%s%s\n", cursor, optionText))
		}

		helpText := m.renderHelpText("ESC: back | arrows/jk: navigate | enter: select")
		if helpText != "" {
			b.WriteString("\n")
			b.WriteString(helpText)
		}
	}

	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := m.errorStyle().Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	return b.String()
}

// renderBvBGamePlay renders the Bot vs Bot gameplay screen.
// Shows the current state of the running games in single-board or grid view.
func (m Model) renderBvBGamePlay() string {
	if m.bvbManager == nil {
		return "No session running.\n"
	}

	if m.bvbViewMode == BvBSingleView {
		return m.renderBvBSingleView()
	}
	return m.renderBvBGridView()
}

// renderBvBStats renders the Bot vs Bot statistics screen after all games finish.
func (m Model) renderBvBStats() string {
	var b strings.Builder

	title := m.titleStyle().Render("TermChess - Bot vs Bot Results")
	b.WriteString(title)
	b.WriteString("\n\n")

	if m.bvbManager == nil {
		b.WriteString("No session data available.\n")
		return b.String()
	}

	stats := m.bvbManager.Stats()
	if stats == nil || stats.TotalGames == 0 {
		b.WriteString("No games completed.\n")
		return b.String()
	}

	infoStyle := lipgloss.NewStyle().
		Foreground(m.theme.StatusText).
		Padding(0, 2)

	statStyle := lipgloss.NewStyle().
		Padding(0, 2)

	dimStyle := lipgloss.NewStyle().
		Foreground(m.theme.HelpText).
		Padding(0, 2)

	if stats.TotalGames == 1 {
		// Single game stats
		r := stats.IndividualResults[0]
		b.WriteString(infoStyle.Render(fmt.Sprintf("%s (White) vs %s (Black)", stats.WhiteBotName, stats.BlackBotName)))
		b.WriteString("\n\n")

		if r.Winner == "Draw" {
			b.WriteString(statStyle.Render(fmt.Sprintf("Result: Draw (%s)", r.EndReason)))
		} else {
			b.WriteString(statStyle.Render(fmt.Sprintf("Winner: %s (%s)", r.Winner, r.EndReason)))
		}
		b.WriteString("\n")
		b.WriteString(statStyle.Render(fmt.Sprintf("Total moves: %d", r.MoveCount)))
		b.WriteString("\n")
		b.WriteString(statStyle.Render(fmt.Sprintf("Duration: %s", r.Duration.Round(time.Millisecond))))
		b.WriteString("\n")
	} else {
		// Multi-game stats
		b.WriteString(infoStyle.Render(fmt.Sprintf("%s (White) vs %s (Black) — %d games", stats.WhiteBotName, stats.BlackBotName, stats.TotalGames)))
		b.WriteString("\n\n")

		// Win/loss/draw summary
		b.WriteString(statStyle.Render(fmt.Sprintf("%s wins: %d (%.1f%%)", stats.WhiteBotName, stats.WhiteWins, stats.WhiteWinPct)))
		b.WriteString("\n")
		b.WriteString(statStyle.Render(fmt.Sprintf("%s wins: %d (%.1f%%)", stats.BlackBotName, stats.BlackWins, stats.BlackWinPct)))
		b.WriteString("\n")
		b.WriteString(statStyle.Render(fmt.Sprintf("Draws: %d", stats.Draws)))
		b.WriteString("\n\n")

		// Averages
		b.WriteString(statStyle.Render(fmt.Sprintf("Avg moves: %.1f | Avg duration: %s", stats.AvgMoveCount, stats.AvgDuration.Round(time.Millisecond))))
		b.WriteString("\n")

		// Shortest/longest
		b.WriteString(statStyle.Render(fmt.Sprintf("Shortest game: #%d (%d moves) | Longest game: #%d (%d moves)",
			stats.ShortestGame.GameNumber, stats.ShortestGame.MoveCount,
			stats.LongestGame.GameNumber, stats.LongestGame.MoveCount)))
		b.WriteString("\n\n")

		// Individual results (paginated, 15 per page)
		resultsPerPage := 15
		totalResults := len(stats.IndividualResults)
		totalPages := (totalResults + resultsPerPage - 1) / resultsPerPage
		currentPage := m.bvbStatsResultsPage
		if currentPage >= totalPages {
			currentPage = totalPages - 1
		}
		if currentPage < 0 {
			currentPage = 0
		}

		startIdx := currentPage * resultsPerPage
		endIdx := startIdx + resultsPerPage
		if endIdx > totalResults {
			endIdx = totalResults
		}

		b.WriteString(dimStyle.Render(fmt.Sprintf("Individual Results (Page %d/%d):", currentPage+1, totalPages)))
		b.WriteString("\n")
		for _, r := range stats.IndividualResults[startIdx:endIdx] {
			var resultText string
			if r.Winner == "Draw" {
				resultText = fmt.Sprintf("  Game %d: Draw (%s) — %d moves", r.GameNumber, r.EndReason, r.MoveCount)
			} else {
				resultText = fmt.Sprintf("  Game %d: %s wins (%s) — %d moves", r.GameNumber, r.Winner, r.EndReason, r.MoveCount)
			}
			b.WriteString(dimStyle.Render(resultText))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Menu options
	for i, opt := range m.menuOptions {
		cursor := "  "
		optionText := m.menuItemStyle().Render(opt)
		if i == m.bvbStatsSelection {
			cursor = m.cursorStyle().Render("> ")
			optionText = m.selectedItemStyle().Render(opt)
		}
		b.WriteString(cursor + optionText)
		b.WriteString("\n")
	}

	// Build help text, including pagination controls if multiple pages
	helpStr := "up/down: navigate | Enter: select | ESC: menu"
	if stats.TotalGames > 1 {
		totalPages := (len(stats.IndividualResults) + 14) / 15 // resultsPerPage = 15
		if totalPages > 1 {
			helpStr = "up/down: navigate | left/right: page | Enter: select | ESC: menu"
		}
	}
	helpText := m.renderHelpText(helpStr)
	if helpText != "" {
		b.WriteString("\n")
		b.WriteString(helpText)
	}

	return b.String()
}

// renderBvBSingleView renders a single game with full detail.
func (m Model) renderBvBSingleView() string {
	var b strings.Builder

	title := m.titleStyle().Render("TermChess - Bot vs Bot")
	b.WriteString(title)
	b.WriteString("\n\n")

	sessions := m.bvbManager.Sessions()
	if len(sessions) == 0 {
		b.WriteString("No games available.\n")
		return b.String()
	}

	// Clamp selected game index
	selectedIdx := m.bvbSelectedGame
	if selectedIdx >= len(sessions) {
		selectedIdx = len(sessions) - 1
	}
	session := sessions[selectedIdx]

	// Show game info header
	infoStyle := lipgloss.NewStyle().
		Foreground(m.theme.StatusText).
		Padding(0, 2)

	matchup := fmt.Sprintf("%s Bot (White) vs %s Bot (Black)",
		botDifficultyName(m.bvbWhiteDiff), botDifficultyName(m.bvbBlackDiff))
	b.WriteString(infoStyle.Render(matchup))
	b.WriteString("\n")

	// Game number and progress
	finished := 0
	for _, s := range sessions {
		if s.IsFinished() {
			finished++
		}
	}
	running := m.bvbManager.RunningCount()
	queued := m.bvbManager.QueuedCount()
	var gameInfo string
	if len(sessions) > 1 {
		gameInfo = fmt.Sprintf("Game %d of %d | Completed: %d | Running: %d | Queued: %d",
			selectedIdx+1, len(sessions), finished, running, queued)
	} else {
		gameInfo = fmt.Sprintf("Game %d of %d", selectedIdx+1, len(sessions))
	}
	b.WriteString(infoStyle.Render(gameInfo))
	b.WriteString("\n\n")

	// Render the chess board
	board := session.CurrentBoard()
	renderer := NewBoardRenderer(m.config)
	boardStr := renderer.Render(board)
	b.WriteString(boardStr)
	b.WriteString("\n\n")

	// Move count and status
	moves := session.CurrentMoveHistory()
	moveCount := len(moves)

	statusLine := fmt.Sprintf("Moves: %d", moveCount)
	if session.IsFinished() {
		result := session.Result()
		if result != nil {
			statusLine += fmt.Sprintf(" | Result: %s (%s)", result.Winner, result.EndReason)
		}
	} else {
		if board.ActiveColor == 0 {
			statusLine += " | White to move"
		} else {
			statusLine += " | Black to move"
		}
	}

	statusLineStyle := lipgloss.NewStyle().
		Foreground(m.theme.MenuSelected).
		Bold(true)
	b.WriteString(statusLineStyle.Render(statusLine))
	b.WriteString("\n")

	// Show pause/speed status
	speedNames := map[bvb.PlaybackSpeed]string{
		bvb.SpeedInstant: "Instant",
		bvb.SpeedFast:    "Fast",
		bvb.SpeedNormal:  "Normal",
		bvb.SpeedSlow:    "Slow",
	}
	controlStatus := fmt.Sprintf("Speed: %s", speedNames[m.bvbSpeed])
	if m.bvbPaused {
		controlStatus += " | PAUSED"
	}
	controlStyle := lipgloss.NewStyle().
		Foreground(m.theme.MenuNormal).
		Padding(0, 2)
	b.WriteString(controlStyle.Render(controlStatus))
	b.WriteString("\n")

	// Move history (if enabled and there are moves)
	if m.config.ShowMoveHistory && moveCount > 0 {
		b.WriteString("\n")
		historyHeader := lipgloss.NewStyle().
			Bold(true).
			Foreground(m.theme.TitleText).
			Render("Move History:")
		b.WriteString(historyHeader)
		b.WriteString("\n")

		historyText := FormatMoveHistory(moves)
		historyStyle := lipgloss.NewStyle().
			Foreground(m.theme.MenuSelected)
		b.WriteString(historyStyle.Render(historyText))
		b.WriteString("\n")
	}

	// Help text
	helpStr := "Space: pause/resume | 1-4: speed | "
	if m.bvbGameCount > 1 {
		helpStr += "left/right: games | "
	}
	helpStr += "Tab: view | f: FEN | ESC: abort"
	helpText := m.renderHelpText(helpStr)
	if helpText != "" {
		b.WriteString("\n")
		b.WriteString(helpText)
	}

	return b.String()
}

// botDifficultyName returns the display name for a bot difficulty.
func botDifficultyName(d BotDifficulty) string {
	switch d {
	case BotEasy:
		return "Easy"
	case BotMedium:
		return "Medium"
	case BotHard:
		return "Hard"
	default:
		return "Unknown"
	}
}

// formatMoveHistory formats the move history for display with a header.
// Returns an empty string if there are no moves to display.
// Format: "Move History: 1. e4 e5 2. Nf3 Nc6"
func (m Model) formatMoveHistory() string {
	if len(m.moveHistory) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("Move History: ")
	
	// We need to replay moves on a board to format them as SAN
	board := engine.NewBoard()
	
	for i := 0; i < len(m.moveHistory); i += 2 {
		moveNum := (i / 2) + 1

		// Format white's move
		whiteSAN := FormatSAN(board, m.moveHistory[i])
		board.MakeMove(m.moveHistory[i])

		// Format black's move (if exists)
		if i+1 < len(m.moveHistory) {
			blackSAN := FormatSAN(board, m.moveHistory[i+1])
			board.MakeMove(m.moveHistory[i+1])
			b.WriteString(fmt.Sprintf("%d. %s %s", moveNum, whiteSAN, blackSAN))

			// Add space only if there are more moves to come
			if i+2 < len(m.moveHistory) {
				b.WriteString(" ")
			}
		} else {
			// Only white's move (game in progress)
			b.WriteString(fmt.Sprintf("%d. %s", moveNum, whiteSAN))
		}
	}
	
	return b.String()
}
