package ui

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/Mgrdich/TermChess/internal/bvb"
	"github.com/Mgrdich/TermChess/internal/engine"
	"github.com/Mgrdich/TermChess/internal/updater"
	"github.com/Mgrdich/TermChess/internal/version"
	"github.com/charmbracelet/lipgloss"
)

// Terminal size and Bot vs Bot grid constants.
const (
	// minTerminalWidth is the minimum terminal width for the UI to render properly.
	minTerminalWidth = 40

	// minTerminalHeight is the minimum terminal height for the UI to render properly.
	minTerminalHeight = 20

	// bvbCellHeight is the fixed height for each grid cell in lines.
	// Breakdown: header (1) + board (8) + status (1) + result (1) + spacing (1) = 12 lines
	bvbCellHeight = 12

	// bvbCellWidth is the fixed width for each grid cell in characters.
	// Board width without coords is 15 chars (8 pieces + 7 spaces).
	// Adding padding and margin gives us 22 characters.
	bvbCellWidth = 22
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

// whiteTurnStyle returns the style for white's turn indicator.
func (m Model) whiteTurnStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.theme.WhiteTurnText).
		Bold(true)
}

// blackTurnStyle returns the style for black's turn indicator.
func (m Model) blackTurnStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.theme.BlackTurnText).
		Bold(true)
}

// turnStyle returns the appropriate style for the current turn.
func (m Model) turnStyle() lipgloss.Style {
	if m.board != nil && m.board.ActiveColor == 1 { // Black
		return m.blackTurnStyle()
	}
	return m.whiteTurnStyle()
}

// breadcrumbStyle returns the style for navigation breadcrumbs.
func (m Model) breadcrumbStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.theme.HelpText).
		Italic(true)
}

// menuPrimaryStyle returns the style for primary menu items (New Game, Start, Resume).
func (m Model) menuPrimaryStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.theme.MenuPrimary).
		Bold(true).
		Padding(0, 2)
}

// menuSecondaryStyle returns the style for secondary menu items (Settings, Load Game, Exit).
func (m Model) menuSecondaryStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.theme.MenuSecondary).
		Padding(0, 2)
}

// selectedPrimaryStyle returns the style for selected primary menu items.
func (m Model) selectedPrimaryStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.theme.MenuSelected).
		Bold(true).
		Padding(0, 2)
}

// selectedSecondaryStyle returns the style for selected secondary menu items.
func (m Model) selectedSecondaryStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.theme.MenuSelected).
		Padding(0, 2)
}

// menuSeparatorStyle returns the style for menu separators.
func (m Model) menuSeparatorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.theme.MenuSeparator)
}

// renderMenuSeparator returns a styled horizontal separator line for menus.
func (m Model) renderMenuSeparator() string {
	separator := "  ────────────────"
	return m.menuSeparatorStyle().Render(separator)
}

// isPrimaryAction returns true if the menu option is a primary action.
// Primary actions: New Game, Resume Game, Start, Play Again, New Session
func isPrimaryAction(option string) bool {
	switch option {
	case "New Game", "Resume Game", "Start", "Play Again", "New Session":
		return true
	default:
		return false
	}
}

// renderBreadcrumb renders the navigation breadcrumb if present.
// Returns an empty string if there's no breadcrumb to display.
func (m Model) renderBreadcrumb() string {
	bc := m.breadcrumb()
	if bc == "" {
		return ""
	}
	return m.breadcrumbStyle().Render(bc) + "\n\n"
}

// renderHelpText conditionally renders help text based on config.
// Returns empty string if help text is disabled.
func (m Model) renderHelpText(text string) string {
	if !m.config.ShowHelpText {
		return ""
	}
	return m.helpStyle().Render(text)
}

// renderMinSizeWarning renders a warning when the terminal is too small.
func (m Model) renderMinSizeWarning() string {
	var b strings.Builder

	warnStyle := lipgloss.NewStyle().
		Foreground(m.theme.ErrorText).
		Bold(true)

	b.WriteString(warnStyle.Render("Terminal too small"))
	b.WriteString("\n\n")

	infoStyle := lipgloss.NewStyle().
		Foreground(m.theme.HelpText)

	b.WriteString(infoStyle.Render(fmt.Sprintf("Current: %dx%d", m.termWidth, m.termHeight)))
	b.WriteString("\n")
	b.WriteString(infoStyle.Render(fmt.Sprintf("Minimum: %dx%d", minTerminalWidth, minTerminalHeight)))
	b.WriteString("\n\n")
	b.WriteString(infoStyle.Render("Please resize your terminal."))

	return b.String()
}

// View renders the UI based on the current model state.
// This function is called by Bubbletea on every update to generate
// the string that will be displayed in the terminal.
func (m Model) View() string {
	// Check if terminal is too small to render properly
	if m.termWidth > 0 && m.termHeight > 0 {
		if m.termWidth < minTerminalWidth || m.termHeight < minTerminalHeight {
			return m.renderMinSizeWarning()
		}
	}

	// If the shortcuts overlay is active, render it over the current view
	if m.showShortcutsOverlay {
		return m.renderShortcutsOverlay()
	}

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
	case ScreenBvBViewModeSelect:
		return m.renderBvBViewModeSelect()
	case ScreenBvBConcurrencySelect:
		return m.renderBvBConcurrencySelect()
	default:
		return "Unknown screen"
	}
}

// renderMainMenu renders the main menu screen with title, menu options,
// cursor indicator, help text, and any error or status messages.
// The "Resume Game" option (if present) is visually distinct with a special indicator and color.
// Menu is organized with visual separators between primary actions (game-related) and
// secondary actions (settings/exit).
func (m Model) renderMainMenu() string {
	var b strings.Builder

	// Render the application title
	title := m.titleStyle().Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Track when separator has been inserted
	// Main menu structure: [Resume Game], New Game, Load Game | Settings, Exit
	// Primary: Resume Game, New Game
	// Secondary: Load Game, Settings, Exit
	separatorInserted := false

	// Render menu options with cursor indicator for selected item
	for i, option := range m.menuOptions {
		// Check if we need to insert a separator before this item
		// Insert separator before "Settings" to separate game actions from app actions
		if option == "Settings" && !separatorInserted {
			b.WriteString(m.renderMenuSeparator())
			b.WriteString("\n")
			separatorInserted = true
		}

		cursor := "  " // Two spaces for non-selected items
		optionText := option

		// Check if this is the "Resume Game" option
		isResumeGame := option == "Resume Game"
		isPrimary := isPrimaryAction(option)

		if i == m.menuSelection {
			// Highlight the selected item with focus indicator
			if isResumeGame {
				// Special styling for selected Resume Game option
				cursor = m.cursorStyle().Render(">> ")
				resumeStyle := lipgloss.NewStyle().
					Foreground(m.theme.StatusText).
					Bold(true).
					Padding(0, 2)
				optionText = resumeStyle.Render(option)
			} else if isPrimary {
				cursor = m.cursorStyle().Render(">> ")
				optionText = m.selectedPrimaryStyle().Render(option)
			} else {
				cursor = m.cursorStyle().Render(" > ")
				optionText = m.selectedSecondaryStyle().Render(option)
			}
		} else {
			// Regular menu item styling
			if isResumeGame {
				// Special styling for unselected Resume Game option
				resumeStyle := lipgloss.NewStyle().
					Foreground(m.theme.StatusText).
					Padding(0, 2)
				optionText = resumeStyle.Render(option)
			} else if isPrimary {
				optionText = m.menuPrimaryStyle().Render(option)
			} else {
				optionText = m.menuSecondaryStyle().Render(option)
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

	// Render update notification if available
	if m.updateAvailable != "" {
		b.WriteString("\n\n")
		updateStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")). // Orange color
			Bold(true)

		var updateText string
		installMethod := updater.DetectInstallMethod()
		if installMethod == updater.InstallMethodGoInstall {
			updateText = fmt.Sprintf("Update available: %s (current: %s). Run 'go install github.com/Mgrdich/TermChess/cmd/termchess@latest' to update.",
				m.updateAvailable, version.Version)
		} else {
			updateText = fmt.Sprintf("Update available: %s (current: %s). Run 'termchess --upgrade' to update.",
				m.updateAvailable, version.Version)
		}
		b.WriteString(updateStyle.Render(updateText))
	}

	return b.String()
}

// renderGameTypeSelect renders the GameTypeSelect screen with title, game type options,
// cursor indicator, help text, and any error or status messages.
// Game type options are styled with visual hierarchy - game modes are primary actions.
func (m Model) renderGameTypeSelect() string {
	var b strings.Builder

	// Render the application title
	title := m.titleStyle().Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n")

	// Render breadcrumb navigation
	b.WriteString(m.renderBreadcrumb())

	// Render screen header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.TitleText).
		Padding(0, 0, 1, 0)
	header := headerStyle.Render("Select Game Type:")
	b.WriteString(header)
	b.WriteString("\n")

	// Render menu options with cursor indicator for selected item
	// All game type options are primary actions
	for i, option := range m.menuOptions {
		cursor := "  " // Two spaces for non-selected items
		optionText := option

		if i == m.menuSelection {
			// Highlight the selected item with prominent focus indicator
			cursor = m.cursorStyle().Render(">> ")
			optionText = m.selectedPrimaryStyle().Render(option)
		} else {
			// Primary styling for all game type options
			optionText = m.menuPrimaryStyle().Render(option)
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
	b.WriteString("\n")

	// Render breadcrumb navigation
	b.WriteString(m.renderBreadcrumb())

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
			// Highlight the selected item with focus indicator
			cursor = m.cursorStyle().Render(">> ")
			optionText = m.selectedPrimaryStyle().Render(option)
		} else {
			// Primary styling for difficulty options
			optionText = m.menuPrimaryStyle().Render(option)
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
	b.WriteString("\n")

	// Render breadcrumb navigation
	b.WriteString(m.renderBreadcrumb())

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
			// Highlight the selected item with focus indicator
			cursor = m.cursorStyle().Render(">> ")
			optionText = m.selectedPrimaryStyle().Render(option)
		} else {
			// Primary styling for color options
			optionText = m.menuPrimaryStyle().Render(option)
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

	// Render the chess board with selection highlighting
	renderer := NewBoardRendererWithTheme(m.config, m.theme)
	boardStr := renderer.RenderWithSelection(m.board, m.selectedSquare, m.validMoves, m.blinkOn)
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

	// Render turn indicator with turn-based color
	b.WriteString("\n\n")
	turnText := "White to move"
	turnStyle := m.whiteTurnStyle()
	if m.board.ActiveColor == 1 { // Black
		turnText = "Black to move"
		turnStyle = m.blackTurnStyle()
	}
	b.WriteString(turnStyle.Render(turnText))

	// Render input prompt with turn-based color for the input text
	b.WriteString("\n\n")
	inputPrompt := lipgloss.NewStyle().
		Foreground(m.theme.MenuNormal).
		Render("Enter move: ")
	inputText := turnStyle.Render(m.input)
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
// Settings are grouped with visual separators between display options and appearance options.
func (m Model) renderSettings() string {
	var b strings.Builder

	// Render the application title
	title := m.titleStyle().Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n")

	// Render breadcrumb navigation
	b.WriteString(m.renderBreadcrumb())

	// Render screen header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.TitleText).
		Padding(0, 0, 1, 0)
	header := headerStyle.Render("Settings")
	b.WriteString(header)
	b.WriteString("\n")

	// Define toggle settings options with their current values
	// The order here determines the settingsSelection index (0-4 for toggles)
	// Group 1 (Display): Use Unicode, Show Coordinates, Use Colors
	// Group 2 (Info): Show Move History, Show Help Text
	toggleOptions := []struct {
		label   string
		enabled bool
		group   int // 1 = display, 2 = info
	}{
		{"Use Unicode Pieces", m.config.UseUnicode, 1},
		{"Show Coordinates", m.config.ShowCoords, 1},
		{"Use Colors", m.config.UseColors, 1},
		{"Show Move History", m.config.ShowMoveHistory, 2},
		{"Show Help Text", m.config.ShowHelpText, 2},
	}

	currentGroup := 0

	// Render each toggle option with its current state
	for i, option := range toggleOptions {
		// Insert separator when changing groups
		if option.group != currentGroup && currentGroup != 0 {
			b.WriteString(m.renderMenuSeparator())
			b.WriteString("\n")
		}
		currentGroup = option.group

		cursor := "  " // Two spaces for non-selected items

		// Determine checkbox state
		checkbox := "[ ]"
		if option.enabled {
			checkbox = "[X]"
		}

		// Build the option text
		optionText := fmt.Sprintf("%s %s", option.label, checkbox)

		if i == m.settingsSelection {
			// Highlight the selected item with focus indicator
			cursor = m.cursorStyle().Render(">> ")
			optionText = m.selectedItemStyle().Render(optionText)
		} else {
			// Regular menu item styling
			optionText = m.menuItemStyle().Render(optionText)
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, optionText))
	}

	// Add separator before theme option
	b.WriteString(m.renderMenuSeparator())
	b.WriteString("\n")

	// Render the Theme option (index 5)
	// Get theme display name with proper capitalization
	themeDisplayName := getThemeDisplayName(m.config.Theme)
	themeCursor := "  "
	themeText := fmt.Sprintf("Theme: %s", themeDisplayName)

	if m.settingsSelection == 5 {
		themeCursor = m.cursorStyle().Render(">> ")
		themeText = m.selectedItemStyle().Render(themeText)
	} else {
		themeText = m.menuItemStyle().Render(themeText)
	}
	b.WriteString(fmt.Sprintf("%s%s\n", themeCursor, themeText))

	// Render help text
	helpText := m.renderHelpText("ESC: back | arrows/jk: navigate | enter/space: toggle/cycle")
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
	optionsText := "y: Save & Exit  |  n: Exit without saving  |  ESC: Cancel"
	b.WriteString(optionsStyle.Render(optionsText))

	// Render help text
	helpText := m.renderHelpText("y: save & exit | n: exit without saving | ESC: cancel")
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
	b.WriteString("\n")

	// Render breadcrumb navigation
	b.WriteString(m.renderBreadcrumb())

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
			// Highlight the selected item with focus indicator
			cursor = m.cursorStyle().Render(">> ")
			optionText = m.selectedPrimaryStyle().Render(option)
		} else {
			// Primary styling for options
			optionText = m.menuPrimaryStyle().Render(option)
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
			cursor = m.cursorStyle().Render(">> ")
			optionText = m.selectedPrimaryStyle().Render(option)
		} else {
			optionText = m.menuPrimaryStyle().Render(option)
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
		b.WriteString(inputStyle.Render(">> " + inputDisplay))
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
				cursor = m.cursorStyle().Render(">> ")
				optionText = m.selectedPrimaryStyle().Render(option)
			} else {
				optionText = m.menuPrimaryStyle().Render(option)
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
	concurrency := m.bvbManager.Concurrency()
	matchup := fmt.Sprintf("%s Bot (White) vs %s Bot (Black) | Completed: %d/%d | Running: %d | Queued: %d | Concurrency: %d",
		botDifficultyName(m.bvbWhiteDiff), botDifficultyName(m.bvbBlackDiff),
		finished, len(sessions), running, queued, concurrency)
	b.WriteString(infoStyle.Render(matchup))
	b.WriteString("\n\n")

	// Render live statistics panel
	liveStats := m.renderBvBLiveStats()
	if liveStats != "" {
		b.WriteString(liveStats)
		b.WriteString("\n\n")
	}

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
		bvb.SpeedNormal:  "Normal",
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

	// Jump prompt (if showing)
	if m.bvbShowJumpPrompt {
		b.WriteString("\n")
		jumpPromptStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(m.theme.MenuSelected).
			Padding(0, 2)
		inputDisplay := m.bvbJumpInput
		if inputDisplay == "" {
			inputDisplay = "_"
		}
		jumpPrompt := fmt.Sprintf("Jump to game (1-%d): %s", m.bvbGameCount, inputDisplay)
		b.WriteString(jumpPromptStyle.Render(jumpPrompt))
		b.WriteString("\n")

		jumpHintStyle := lipgloss.NewStyle().
			Foreground(m.theme.HelpText).
			Italic(true).
			Padding(0, 2)
		b.WriteString(jumpHintStyle.Render("Enter: jump | Esc: cancel"))
		b.WriteString("\n")
	}

	// Error message if present
	if m.errorMsg != "" {
		b.WriteString("\n")
		errorText := m.errorStyle().Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Help text
	helpText := m.renderHelpText("Space: pause/resume | t: toggle speed | ←/→: pages | g: jump to game | Tab: single view | f: FEN | ESC: abort")
	if helpText != "" {
		b.WriteString("\n")
		b.WriteString(helpText)
	}

	return b.String()
}

// renderBoardGrid renders a slice of sessions as a grid with the given number of columns.
// Each cell has fixed dimensions to prevent layout shifts when games complete.
func (m Model) renderBoardGrid(sessions []*bvb.GameSession, cols int) string {
	if len(sessions) == 0 {
		return ""
	}

	// Render each session as a fixed-dimension compact board cell
	cells := make([]string, len(sessions))
	for i, session := range sessions {
		cells[i] = m.renderCompactBoardCell(session)
	}

	// Arrange cells into rows with consistent alignment
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

// renderBvBGridCell renders a grid cell for the game at the given index.
// Returns an empty string if the index is out of bounds or manager is nil.
// The cell has fixed dimensions (bvbCellHeight x bvbCellWidth) to prevent layout shifts.
func (m Model) renderBvBGridCell(gameIndex int) string {
	if m.bvbManager == nil {
		return ""
	}

	sessions := m.bvbManager.Sessions()
	if gameIndex < 0 || gameIndex >= len(sessions) {
		return ""
	}

	return m.renderCompactBoardCell(sessions[gameIndex])
}

// renderCompactBoardCell renders a single game session as a compact board cell for the grid.
// Shows: game number, compact board, move count, and status.
// The cell has fixed dimensions (bvbCellHeight x bvbCellWidth) to prevent layout shifts.
func (m Model) renderCompactBoardCell(session *bvb.GameSession) string {
	board := session.CurrentBoard()
	gameNum := session.GameNumber()
	moveCount := len(session.CurrentMoveHistory())
	isFinished := session.IsFinished()

	// Build cell content lines
	var lines []string

	// Line 1: Game header
	lines = append(lines, fmt.Sprintf("Game %d", gameNum))

	// Lines 2-9: Board (8 lines)
	compactConfig := Config{
		UseUnicode: m.config.UseUnicode,
		ShowCoords: false,
		UseColors:  false,
	}
	renderer := NewBoardRenderer(compactConfig)
	boardStr := renderer.Render(board)
	boardLines := strings.Split(strings.TrimSuffix(boardStr, "\n"), "\n")
	lines = append(lines, boardLines...)

	// Line 10: Status line (always shows move count)
	lines = append(lines, fmt.Sprintf("Moves: %d", moveCount))

	// Line 11: Result line (empty for in-progress, result for finished)
	if isFinished {
		result := session.Result()
		if result != nil {
			lines = append(lines, result.Winner)
		} else {
			lines = append(lines, "") // Empty placeholder
		}
	} else {
		lines = append(lines, "") // Empty placeholder for in-progress games
	}

	// Line 12: Spacing (empty line)
	lines = append(lines, "")

	// Pad or truncate to exactly bvbCellHeight lines
	for len(lines) < bvbCellHeight {
		lines = append(lines, "")
	}
	if len(lines) > bvbCellHeight {
		lines = lines[:bvbCellHeight]
	}

	// Normalize each line to bvbCellWidth characters
	for i, line := range lines {
		lineWidth := lipgloss.Width(line)
		if lineWidth < bvbCellWidth {
			// Pad with spaces to reach target width
			lines[i] = line + strings.Repeat(" ", bvbCellWidth-lineWidth)
		} else if lineWidth > bvbCellWidth {
			// Truncate to target width (keeping ANSI codes intact is tricky,
			// but for our simple case we just truncate)
			lines[i] = truncateToWidth(line, bvbCellWidth)
		}
	}

	// Join lines and apply styling
	content := strings.Join(lines, "\n")

	// Style the cell with consistent dimensions
	cellStyle := lipgloss.NewStyle().
		Width(bvbCellWidth).
		Height(bvbCellHeight).
		Margin(0, 1)

	if isFinished {
		// Dimmed style for finished games
		cellStyle = cellStyle.Foreground(m.theme.HelpText)
	}

	return cellStyle.Render(content)
}

// truncateToWidth truncates a string to fit within the specified width.
// This is a simple implementation that handles most common cases.
func truncateToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	return string(runes[:width])
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
		b.WriteString(inputStyle.Render(">> " + inputDisplay))
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
				cursor = m.cursorStyle().Render(">> ")
				optionText = m.selectedPrimaryStyle().Render(option)
			} else {
				optionText = m.menuPrimaryStyle().Render(option)
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
// Shows the current state of the running games in single-board, grid, or stats-only view.
func (m Model) renderBvBGamePlay() string {
	if m.bvbManager == nil {
		return "No session running.\n"
	}

	// Render the abort confirmation dialog as an overlay if showing
	if m.bvbShowAbortConfirm {
		return m.renderBvBAbortConfirm()
	}

	switch m.bvbViewMode {
	case BvBSingleView:
		return m.renderBvBSingleView()
	case BvBStatsOnlyView:
		return m.renderBvBStatsOnly()
	default:
		return m.renderBvBGridView()
	}
}

// renderBvBAbortConfirm renders the abort confirmation dialog for BvB sessions.
func (m Model) renderBvBAbortConfirm() string {
	var b strings.Builder

	// Title
	title := m.titleStyle().Render("TermChess - Bot vs Bot")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Dialog box
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.MenuSelected).
		Padding(1, 2)

	var dialogContent strings.Builder
	dialogTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.TitleText)
	dialogContent.WriteString(dialogTitleStyle.Render("  Abort Session?  "))
	dialogContent.WriteString("\n\n")
	dialogContent.WriteString("  Games in progress will be lost.\n\n")

	// Options
	normalStyle := m.menuItemStyle()
	selectedStyle := m.selectedItemStyle()

	if m.bvbAbortSelection == 0 {
		dialogContent.WriteString(selectedStyle.Render("  > Cancel"))
	} else {
		dialogContent.WriteString(normalStyle.Render("    Cancel"))
	}
	dialogContent.WriteString("\n")
	if m.bvbAbortSelection == 1 {
		dialogContent.WriteString(selectedStyle.Render("  > Abort Session"))
	} else {
		dialogContent.WriteString(normalStyle.Render("    Abort Session"))
	}
	dialogContent.WriteString("\n\n")
	dialogContent.WriteString(m.helpStyle().Render("  esc: cancel | enter: select"))

	b.WriteString(dialogStyle.Render(dialogContent.String()))

	return b.String()
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

	// Menu options with visual hierarchy
	for i, opt := range m.menuOptions {
		cursor := "  "
		var optionText string
		isPrimary := isPrimaryAction(opt)

		if i == m.bvbStatsSelection {
			cursor = m.cursorStyle().Render(">> ")
			if isPrimary {
				optionText = m.selectedPrimaryStyle().Render(opt)
			} else {
				optionText = m.selectedSecondaryStyle().Render(opt)
			}
		} else {
			if isPrimary {
				optionText = m.menuPrimaryStyle().Render(opt)
			} else {
				optionText = m.menuSecondaryStyle().Render(opt)
			}
		}
		b.WriteString(cursor + optionText)
		b.WriteString("\n")
	}

	// Show status or error messages
	if m.statusMsg != "" {
		statusStyle := lipgloss.NewStyle().
			Foreground(m.theme.StatusText).
			Padding(0, 2)
		b.WriteString("\n")
		b.WriteString(statusStyle.Render(m.statusMsg))
	}
	if m.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(m.theme.ErrorText).
			Padding(0, 2)
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(m.errorMsg))
	}

	// Build help text, including pagination controls if multiple pages
	helpStr := "up/down: navigate | s: export | Enter: select | ESC: menu"
	if stats.TotalGames > 1 {
		totalPages := (len(stats.IndividualResults) + 14) / 15 // resultsPerPage = 15
		if totalPages > 1 {
			helpStr = "up/down: navigate | left/right: page | s: export | Enter: select | ESC: menu"
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

	// Prominent "Game X of Y" indicator for multi-game mode
	if len(sessions) > 1 {
		gameIndicatorStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(m.theme.MenuSelected).
			Padding(0, 2)
		gameIndicator := fmt.Sprintf(">>> Game %d of %d <<<", selectedIdx+1, len(sessions))
		b.WriteString(gameIndicatorStyle.Render(gameIndicator))
		b.WriteString("\n")
	}

	// Game progress info
	finished := 0
	for _, s := range sessions {
		if s.IsFinished() {
			finished++
		}
	}
	running := m.bvbManager.RunningCount()
	queued := m.bvbManager.QueuedCount()
	concurrency := m.bvbManager.Concurrency()
	var gameInfo string
	if len(sessions) > 1 {
		gameInfo = fmt.Sprintf("Completed: %d/%d | Running: %d | Queued: %d | Concurrency: %d",
			finished, len(sessions), running, queued, concurrency)
	} else {
		gameInfo = fmt.Sprintf("Game %d of %d | Concurrency: %d", selectedIdx+1, len(sessions), concurrency)
	}
	b.WriteString(infoStyle.Render(gameInfo))
	b.WriteString("\n\n")

	// Render live statistics panel
	liveStats := m.renderBvBLiveStats()
	if liveStats != "" {
		b.WriteString(liveStats)
		b.WriteString("\n\n")
	}

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
		bvb.SpeedNormal:  "Normal",
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

	// Jump prompt (if showing)
	if m.bvbShowJumpPrompt {
		b.WriteString("\n")
		jumpPromptStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(m.theme.MenuSelected).
			Padding(0, 2)
		inputDisplay := m.bvbJumpInput
		if inputDisplay == "" {
			inputDisplay = "_"
		}
		jumpPrompt := fmt.Sprintf("Jump to game (1-%d): %s", m.bvbGameCount, inputDisplay)
		b.WriteString(jumpPromptStyle.Render(jumpPrompt))
		b.WriteString("\n")

		jumpHintStyle := lipgloss.NewStyle().
			Foreground(m.theme.HelpText).
			Italic(true).
			Padding(0, 2)
		b.WriteString(jumpHintStyle.Render("Enter: jump | Esc: cancel"))
		b.WriteString("\n")
	}

	// Error message if present
	if m.errorMsg != "" {
		b.WriteString("\n")
		errorText := m.errorStyle().Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Help text
	helpStr := "Space: pause/resume | t: toggle speed | "
	if m.bvbGameCount > 1 {
		helpStr += "left/right: games | g: jump to game | "
	}
	helpStr += "Tab: view | f: FEN | ESC: abort"
	helpText := m.renderHelpText(helpStr)
	if helpText != "" {
		b.WriteString("\n")
		b.WriteString(helpText)
	}

	return b.String()
}

// renderBvBLiveStats renders a live statistics panel for Bot vs Bot gameplay.
// Shows current score (White Wins / Black Wins / Draws) and progress (Completed / Total).
// Also shows detailed statistics: average moves, longest/shortest games, current game duration,
// last 10 moves, and captured pieces.
// This panel updates on each BvBTickMsg as the manager's Stats() are recalculated.
func (m Model) renderBvBLiveStats() string {
	if m.bvbManager == nil {
		return ""
	}

	stats := m.bvbManager.Stats()
	sessions := m.bvbManager.Sessions()
	totalGames := len(sessions)

	if totalGames == 0 {
		return ""
	}

	var sb strings.Builder

	// Stats panel header with box-drawing characters
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.TitleText)
	sb.WriteString(headerStyle.Render("══════ Statistics ══════"))
	sb.WriteString("\n")

	// Score line: White Wins | Black Wins | Draws
	scoreStyle := lipgloss.NewStyle().
		Foreground(m.theme.MenuSelected)

	whiteWins := 0
	blackWins := 0
	draws := 0
	completed := 0

	if stats != nil {
		whiteWins = stats.WhiteWins
		blackWins = stats.BlackWins
		draws = stats.Draws
		completed = stats.TotalGames
	}

	scoreLine := fmt.Sprintf("Score: White %d | Black %d | Draws %d", whiteWins, blackWins, draws)
	sb.WriteString(scoreStyle.Render(scoreLine))
	sb.WriteString("\n")

	// Progress line: Games Completed / Total
	progressStyle := lipgloss.NewStyle().
		Foreground(m.theme.StatusText)
	progressLine := fmt.Sprintf("Progress: %d / %d games", completed, totalGames)
	sb.WriteString(progressStyle.Render(progressLine))
	sb.WriteString("\n")

	// Detailed stats (only show if we have completed games)
	detailStyle := lipgloss.NewStyle().
		Foreground(m.theme.MenuNormal)

	if stats != nil && completed > 0 {
		// Average moves per game
		avgMovesLine := fmt.Sprintf("Avg Moves: %.1f", stats.AvgMoveCount)
		sb.WriteString(detailStyle.Render(avgMovesLine))
		sb.WriteString("\n")

		// Longest and shortest games
		longestShortestLine := fmt.Sprintf("Longest: %d moves | Shortest: %d moves",
			stats.LongestGame.MoveCount, stats.ShortestGame.MoveCount)
		sb.WriteString(detailStyle.Render(longestShortestLine))
		sb.WriteString("\n")
	}

	// Current game info (selected game in single view or first running game in grid view)
	currentSession := m.getCurrentBvBSession()
	if currentSession != nil {
		sb.WriteString(headerStyle.Render("─── Current Game ───"))
		sb.WriteString("\n")

		// Game duration timer
		duration := currentSession.Duration()
		durationStr := formatBvBDuration(duration)
		durationLine := fmt.Sprintf("Duration: %s", durationStr)
		sb.WriteString(detailStyle.Render(durationLine))
		sb.WriteString("\n")

		// Last 10 moves
		moves := currentSession.CurrentMoveHistory()
		if len(moves) > 0 {
			lastMoves := formatLastMoves(moves, 10)
			movesLine := fmt.Sprintf("Last moves: %s", lastMoves)
			sb.WriteString(detailStyle.Render(movesLine))
			sb.WriteString("\n")
		}

		// Captured pieces
		board := currentSession.CurrentBoard()
		if board != nil {
			capturedWhite, capturedBlack := computeCapturedPieces(board)
			if len(capturedWhite) > 0 || len(capturedBlack) > 0 {
				capturedLine := fmt.Sprintf("Captured: %s | %s", capturedWhite, capturedBlack)
				sb.WriteString(detailStyle.Render(capturedLine))
				sb.WriteString("\n")
			}
		}
	}

	sb.WriteString(headerStyle.Render("═════════════════════════"))

	return sb.String()
}

// getCurrentBvBSession returns the current session to display detailed stats for.
// In single view mode, returns the selected game.
// In grid view mode, returns the first non-finished game, or nil if all finished.
func (m Model) getCurrentBvBSession() *bvb.GameSession {
	if m.bvbManager == nil {
		return nil
	}

	if m.bvbViewMode == BvBSingleView {
		return m.bvbManager.GetSession(m.bvbSelectedGame)
	}

	// In grid view, find first running game
	sessions := m.bvbManager.Sessions()
	for _, s := range sessions {
		if s != nil && !s.IsFinished() {
			return s
		}
	}
	return nil
}

// formatBvBDuration formats a duration as MM:SS for display.
func formatBvBDuration(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// formatLastMoves formats the last N moves from a move history as a comma-separated string.
// Uses coordinate notation (e.g., "e2e4, e7e5, g1f3").
func formatLastMoves(moves []engine.Move, n int) string {
	if len(moves) == 0 {
		return ""
	}

	start := 0
	if len(moves) > n {
		start = len(moves) - n
	}

	lastMoves := moves[start:]
	moveStrs := make([]string, len(lastMoves))
	for i, m := range lastMoves {
		moveStrs[i] = m.String()
	}

	return strings.Join(moveStrs, ", ")
}

// computeCapturedPieces compares the current board state to a starting position
// and returns strings representing captured pieces for each side.
// Returns (whiteCaptured, blackCaptured) where whiteCaptured shows white pieces
// that have been captured (displayed with white symbols) and blackCaptured shows
// black pieces that have been captured (displayed with black symbols).
func computeCapturedPieces(board *engine.Board) (string, string) {
	// Starting piece counts for each side
	// White: 8 pawns, 2 knights, 2 bishops, 2 rooks, 1 queen, 1 king
	// Black: same
	startingCounts := map[engine.PieceType]int{
		engine.Pawn:   8,
		engine.Knight: 2,
		engine.Bishop: 2,
		engine.Rook:   2,
		engine.Queen:  1,
		engine.King:   1,
	}

	// Count current pieces on board
	whiteCounts := make(map[engine.PieceType]int)
	blackCounts := make(map[engine.PieceType]int)

	for sq := 0; sq < 64; sq++ {
		piece := board.Squares[sq]
		if piece.IsEmpty() {
			continue
		}
		if piece.Color() == engine.White {
			whiteCounts[piece.Type()]++
		} else {
			blackCounts[piece.Type()]++
		}
	}

	// Unicode symbols for pieces
	// White pieces (captured by black): ♙♘♗♖♕
	// Black pieces (captured by white): ♟♞♝♜♛
	whitePieceSymbols := map[engine.PieceType]rune{
		engine.Pawn:   '♙',
		engine.Knight: '♘',
		engine.Bishop: '♗',
		engine.Rook:   '♖',
		engine.Queen:  '♕',
	}
	blackPieceSymbols := map[engine.PieceType]rune{
		engine.Pawn:   '♟',
		engine.Knight: '♞',
		engine.Bishop: '♝',
		engine.Rook:   '♜',
		engine.Queen:  '♛',
	}

	// Build captured strings
	// Order: Queen, Rook, Bishop, Knight, Pawn (most valuable first)
	pieceOrder := []engine.PieceType{engine.Queen, engine.Rook, engine.Bishop, engine.Knight, engine.Pawn}

	var whiteCaptured strings.Builder // White pieces captured (by black)
	var blackCaptured strings.Builder // Black pieces captured (by white)

	for _, pt := range pieceOrder {
		// White pieces captured
		whiteRemaining := whiteCounts[pt]
		whiteMissing := startingCounts[pt] - whiteRemaining
		for i := 0; i < whiteMissing; i++ {
			whiteCaptured.WriteRune(whitePieceSymbols[pt])
		}

		// Black pieces captured
		blackRemaining := blackCounts[pt]
		blackMissing := startingCounts[pt] - blackRemaining
		for i := 0; i < blackMissing; i++ {
			blackCaptured.WriteRune(blackPieceSymbols[pt])
		}
	}

	return whiteCaptured.String(), blackCaptured.String()
}

// renderBvBViewModeSelect renders the Bot vs Bot view mode selection screen.
// Shows three options: Grid View, Single Board, Stats Only with descriptions.
func (m Model) renderBvBViewModeSelect() string {
	var b strings.Builder

	title := m.titleStyle().Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.TitleText).
		Padding(0, 0, 1, 0)
	header := headerStyle.Render("Select View Mode:")
	b.WriteString(header)
	b.WriteString("\n")

	// Show session info
	infoStyle := lipgloss.NewStyle().
		Foreground(m.theme.StatusText).
		Padding(0, 2)
	sessionInfo := fmt.Sprintf("%d game(s) | %s Bot (White) vs %s Bot (Black) | Grid: %dx%d",
		m.bvbGameCount, botDifficultyName(m.bvbWhiteDiff), botDifficultyName(m.bvbBlackDiff),
		m.bvbGridRows, m.bvbGridCols)
	b.WriteString(infoStyle.Render(sessionInfo))
	b.WriteString("\n\n")

	// Define view mode options with descriptions
	type viewModeOption struct {
		name        string
		description string
		hint        string
	}
	options := []viewModeOption{
		{"Grid View", "Watch multiple games in a grid layout", ""},
		{"Single Board", "Focus on one game at a time", ""},
		{"Stats Only", "No boards, just statistics", "(Recommended for 50+ games)"},
	}

	descStyle := lipgloss.NewStyle().
		Foreground(m.theme.HelpText).
		Italic(true).
		Padding(0, 4)

	hintStyle := lipgloss.NewStyle().
		Foreground(m.theme.StatusText).
		Bold(true)

	for i, opt := range options {
		cursor := "  "
		var optionText string

		if i == m.bvbViewModeSelection {
			cursor = m.cursorStyle().Render(">> ")
			optionText = m.selectedPrimaryStyle().Render(opt.name)
		} else {
			optionText = m.menuPrimaryStyle().Render(opt.name)
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, optionText))

		// Show description
		descText := opt.description
		if opt.hint != "" {
			descText += " " + hintStyle.Render(opt.hint)
		}
		b.WriteString(descStyle.Render(descText))
		b.WriteString("\n")
	}

	helpText := m.renderHelpText("ESC: back | arrows/jk: navigate | enter: select")
	if helpText != "" {
		b.WriteString("\n")
		b.WriteString(helpText)
	}

	if m.errorMsg != "" {
		b.WriteString("\n\n")
		errorText := m.errorStyle().Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	return b.String()
}

// renderBvBStatsOnly renders the stats-only view for Bot vs Bot gameplay.
// Displays progress bar, score summary, average moves, in-progress count, and recent completions.
func (m Model) renderBvBStatsOnly() string {
	var b strings.Builder

	title := m.titleStyle().Render("TermChess - Bot vs Bot (Stats Only)")
	b.WriteString(title)
	b.WriteString("\n\n")

	if m.bvbManager == nil {
		b.WriteString("No session running.\n")
		return b.String()
	}

	sessions := m.bvbManager.Sessions()
	totalGames := len(sessions)
	if totalGames == 0 {
		b.WriteString("No games available.\n")
		return b.String()
	}

	// Calculate stats
	stats := m.bvbManager.Stats()
	completed := 0
	inProgress := 0
	whiteWins := 0
	blackWins := 0
	draws := 0
	totalMoves := 0

	for _, s := range sessions {
		if s.IsFinished() {
			completed++
			result := s.Result()
			if result != nil {
				totalMoves += result.MoveCount
				switch result.Winner {
				case "White":
					whiteWins++
				case "Black":
					blackWins++
				case "Draw":
					draws++
				}
			}
		} else {
			inProgress++
		}
	}

	// Override with stats from manager if available
	if stats != nil {
		whiteWins = stats.WhiteWins
		blackWins = stats.BlackWins
		draws = stats.Draws
	}

	infoStyle := lipgloss.NewStyle().
		Foreground(m.theme.StatusText).
		Padding(0, 2)

	// Matchup header
	matchup := fmt.Sprintf("%s Bot (White) vs %s Bot (Black)",
		botDifficultyName(m.bvbWhiteDiff), botDifficultyName(m.bvbBlackDiff))
	b.WriteString(infoStyle.Render(matchup))
	b.WriteString("\n\n")

	// Progress bar
	progressBar := renderProgressBar(completed, totalGames, 40)
	progressStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.MenuSelected).
		Padding(0, 2)
	b.WriteString(progressStyle.Render(progressBar))
	b.WriteString("\n\n")

	// Score summary
	statStyle := lipgloss.NewStyle().
		Foreground(m.theme.MenuNormal).
		Padding(0, 2)

	scoreLine := fmt.Sprintf("Score:  White: %d  |  Black: %d  |  Draws: %d", whiteWins, blackWins, draws)
	scoreStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.TitleText).
		Padding(0, 2)
	b.WriteString(scoreStyle.Render(scoreLine))
	b.WriteString("\n\n")

	// Average moves per completed game
	avgMoves := 0.0
	if completed > 0 {
		if stats != nil {
			avgMoves = stats.AvgMoveCount
		} else {
			avgMoves = float64(totalMoves) / float64(completed)
		}
	}
	avgLine := fmt.Sprintf("Average moves per game: %.1f", avgMoves)
	b.WriteString(statStyle.Render(avgLine))
	b.WriteString("\n")

	// In-progress indicator
	inProgressLine := fmt.Sprintf("%d game(s) in progress", inProgress)
	inProgressStyle := lipgloss.NewStyle().
		Foreground(m.theme.StatusText).
		Padding(0, 2)
	b.WriteString(inProgressStyle.Render(inProgressLine))
	b.WriteString("\n\n")

	// Concurrency info
	running := m.bvbManager.RunningCount()
	queued := m.bvbManager.QueuedCount()
	concurrency := m.bvbManager.Concurrency()
	concurrencyLine := fmt.Sprintf("Running: %d | Queued: %d | Concurrency: %d", running, queued, concurrency)
	b.WriteString(statStyle.Render(concurrencyLine))
	b.WriteString("\n\n")

	// Recent completions log (last 5 results)
	recentHeader := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.TitleText).
		Padding(0, 2)
	b.WriteString(recentHeader.Render("Recent Completions:"))
	b.WriteString("\n")

	recentStyle := lipgloss.NewStyle().
		Foreground(m.theme.HelpText).
		Padding(0, 4)

	if len(m.bvbRecentCompletions) == 0 {
		b.WriteString(recentStyle.Render("(none yet)"))
		b.WriteString("\n")
	} else {
		for _, entry := range m.bvbRecentCompletions {
			b.WriteString(recentStyle.Render(entry))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Speed/pause status
	speedNames := map[bvb.PlaybackSpeed]string{
		bvb.SpeedInstant: "Instant",
		bvb.SpeedNormal:  "Normal",
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

	// Error message if present
	if m.errorMsg != "" {
		b.WriteString("\n")
		errorText := m.errorStyle().Render(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString(errorText)
	}

	// Help text
	helpText := m.renderHelpText("[Space] Pause/Resume | [v] Change view | [t] Speed | [q/ESC] Quit")
	if helpText != "" {
		b.WriteString("\n")
		b.WriteString(helpText)
	}

	return b.String()
}

// renderProgressBar creates a text-based progress bar.
// Parameters: completed items, total items, width of the bar in characters.
func renderProgressBar(completed, total, width int) string {
	if total == 0 {
		return "[" + strings.Repeat("░", width) + "] 0% (0/0)"
	}
	percent := float64(completed) / float64(total)
	filled := int(percent * float64(width))
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("[%s] %d%% (%d/%d)", bar, int(percent*100), completed, total)
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

// getThemeDisplayName returns a display-friendly name for a theme.
// Converts the internal theme name string to a capitalized display name.
func getThemeDisplayName(themeName string) string {
	switch themeName {
	case ThemeNameModern:
		return "Modern"
	case ThemeNameMinimalist:
		return "Minimalist"
	case ThemeNameClassic:
		return "Classic"
	default:
		return "Classic"
	}
}

// renderShortcutsOverlay renders a full-screen modal overlay displaying all keyboard shortcuts.
// The overlay is organized by context (Global, Menu, Settings, Gameplay, Bot vs Bot).
func (m Model) renderShortcutsOverlay() string {
	var b strings.Builder

	// Title style for the overlay
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.TitleText).
		Align(lipgloss.Center).
		Padding(1, 0)

	// Section header style
	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.MenuSelected).
		Padding(1, 0, 0, 0)

	// Shortcut key style (left column)
	keyStyle := lipgloss.NewStyle().
		Foreground(m.theme.MenuSelected).
		Bold(true).
		Width(15)

	// Description style (right column)
	descStyle := lipgloss.NewStyle().
		Foreground(m.theme.MenuNormal)

	// Hint style for the footer
	hintStyle := lipgloss.NewStyle().
		Foreground(m.theme.HelpText).
		Italic(true).
		Padding(2, 0, 0, 0)

	// Render title
	b.WriteString(titleStyle.Render("Keyboard Shortcuts"))
	b.WriteString("\n")

	// Helper function to render a shortcut line
	renderShortcut := func(key, description string) {
		b.WriteString(keyStyle.Render(key))
		b.WriteString(descStyle.Render(description))
		b.WriteString("\n")
	}

	// Global shortcuts
	b.WriteString(sectionStyle.Render("Global"))
	b.WriteString("\n")
	renderShortcut("?", "Show this help overlay")
	renderShortcut("n", "Start new game")
	renderShortcut("s", "Open settings")
	renderShortcut("Ctrl+C", "Quit application")
	renderShortcut("q", "Quit (or show save prompt in game)")
	renderShortcut("Esc", "Go back / Cancel")

	// Menu navigation
	b.WriteString(sectionStyle.Render("Menu Navigation"))
	b.WriteString("\n")
	renderShortcut("Up / k", "Move selection up")
	renderShortcut("Down / j", "Move selection down")
	renderShortcut("Enter", "Select / Confirm")

	// Settings
	b.WriteString(sectionStyle.Render("Settings"))
	b.WriteString("\n")
	renderShortcut("Up / k", "Previous setting")
	renderShortcut("Down / j", "Next setting")
	renderShortcut("Enter/Space", "Toggle / Cycle setting")

	// Gameplay
	b.WriteString(sectionStyle.Render("Gameplay"))
	b.WriteString("\n")
	renderShortcut("Type move", "Enter move (e.g., e4, Nf3, O-O)")
	renderShortcut("Enter", "Submit move")
	renderShortcut("resign", "Resign the game")
	renderShortcut("offerdraw", "Offer a draw")
	renderShortcut("showfen", "Show/copy FEN position")
	renderShortcut("menu", "Return to menu (with save)")

	// Bot vs Bot
	b.WriteString(sectionStyle.Render("Bot vs Bot"))
	b.WriteString("\n")
	renderShortcut("Space", "Pause / Resume")
	renderShortcut("Left / h", "Previous game / page")
	renderShortcut("Right / l", "Next game / page")
	renderShortcut("g", "Jump to game (enter game number)")
	renderShortcut("Tab", "Toggle grid / single view")
	renderShortcut("t", "Toggle speed (Normal / Instant)")
	renderShortcut("f", "Copy FEN of current game")

	// Footer hint
	b.WriteString(hintStyle.Render("Press any key to close"))

	return b.String()
}

// renderBvBConcurrencySelect renders the Bot vs Bot concurrency selection screen.
// Shows two options: Recommended (auto-calculated based on CPU) and Custom.
func (m Model) renderBvBConcurrencySelect() string {
	var b strings.Builder

	title := m.titleStyle().Render("TermChess")
	b.WriteString(title)
	b.WriteString("\n\n")

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.TitleText).
		Padding(0, 0, 1, 0)
	header := headerStyle.Render("Select Concurrency:")
	b.WriteString(header)
	b.WriteString("\n")

	// Show session info
	infoStyle := lipgloss.NewStyle().
		Foreground(m.theme.StatusText).
		Padding(0, 2)
	sessionInfo := fmt.Sprintf("%d game(s) | %s Bot (White) vs %s Bot (Black)",
		m.bvbGameCount, botDifficultyName(m.bvbWhiteDiff), botDifficultyName(m.bvbBlackDiff))
	b.WriteString(infoStyle.Render(sessionInfo))
	b.WriteString("\n\n")

	// Get recommended concurrency and CPU count
	recommendedConcurrency := bvb.CalculateDefaultConcurrency()
	numCPU := getCPUCount()

	// Define options
	type concurrencyOption struct {
		name        string
		description string
	}
	options := []concurrencyOption{
		{
			name:        fmt.Sprintf("Recommended (%d concurrent games)", recommendedConcurrency),
			description: fmt.Sprintf("Based on your CPU (%d cores)", numCPU),
		},
		{
			name:        "Custom",
			description: "Enter your own value (may cause lag)",
		},
	}

	descStyle := lipgloss.NewStyle().
		Foreground(m.theme.HelpText).
		Italic(true).
		Padding(0, 4)

	// If inputting custom value, show input field instead of menu
	if m.bvbInputtingConcurrency {
		// Show input prompt
		inputStyle := lipgloss.NewStyle().
			Foreground(m.theme.TitleText).
			Padding(0, 2)
		b.WriteString(inputStyle.Render("Enter concurrency: "))

		// Show the input with cursor
		inputValueStyle := lipgloss.NewStyle().
			Foreground(m.theme.MenuSelected).
			Bold(true)
		inputText := m.bvbCustomConcurrency + "_"
		b.WriteString(inputValueStyle.Render(inputText))
		b.WriteString("\n")

		// Show warning if value exceeds 50
		if val := parseConcurrencyValue(m.bvbCustomConcurrency); val > 50 {
			warnStyle := lipgloss.NewStyle().
				Foreground(m.theme.ErrorText).
				Bold(true).
				Padding(0, 2)
			b.WriteString("\n")
			b.WriteString(warnStyle.Render("Warning: High concurrency may cause lag. Consider using Stats Only view mode."))
			b.WriteString("\n")
		}

		helpText := m.renderHelpText("enter: confirm | esc: cancel")
		if helpText != "" {
			b.WriteString("\n")
			b.WriteString(helpText)
		}
	} else {
		// Show menu options
		for i, opt := range options {
			cursor := "  "
			var optionText string

			if i == m.bvbConcurrencySelection {
				cursor = m.cursorStyle().Render(">> ")
				optionText = m.selectedPrimaryStyle().Render(opt.name)
			} else {
				optionText = m.menuPrimaryStyle().Render(opt.name)
			}

			b.WriteString(fmt.Sprintf("%s%s\n", cursor, optionText))
			b.WriteString(descStyle.Render(opt.description))
			b.WriteString("\n")
		}

		helpText := m.renderHelpText("arrows/jk: navigate | enter: select | esc: back")
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

// parseConcurrencyValue parses a string into an integer, returning 0 if invalid.
func parseConcurrencyValue(s string) int {
	if s == "" {
		return 0
	}
	var val int
	for _, r := range s {
		if r >= '0' && r <= '9' {
			val = val*10 + int(r-'0')
		}
	}
	return val
}

// getCPUCount returns the number of CPUs available.
func getCPUCount() int {
	return runtime.NumCPU()
}
