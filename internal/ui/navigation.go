package ui

// Navigation stack management methods for the Model.
// These methods provide push/pop navigation and breadcrumb generation.

// screenName returns a human-readable name for a screen.
func screenName(s Screen) string {
	switch s {
	case ScreenMainMenu:
		return "Main Menu"
	case ScreenGameTypeSelect:
		return "New Game"
	case ScreenBotSelect:
		return "Bot Difficulty"
	case ScreenColorSelect:
		return "Choose Color"
	case ScreenFENInput:
		return "Load Game"
	case ScreenGamePlay:
		return "Game"
	case ScreenGameOver:
		return "Game Over"
	case ScreenSettings:
		return "Settings"
	case ScreenSavePrompt:
		return "Save Game"
	case ScreenResumePrompt:
		return "Resume Game"
	case ScreenDrawPrompt:
		return "Draw Offer"
	case ScreenBvBBotSelect:
		return "Bot vs Bot Setup"
	case ScreenBvBGameMode:
		return "Game Mode"
	case ScreenBvBGridConfig:
		return "Grid Layout"
	case ScreenBvBGamePlay:
		return "Bot vs Bot"
	case ScreenBvBStats:
		return "Statistics"
	case ScreenBvBViewModeSelect:
		return "View Mode"
	case ScreenBvBConcurrencySelect:
		return "Concurrency Select"
	default:
		return "Unknown"
	}
}

// pushScreen navigates to a new screen, saving the current screen to the navigation stack.
// Use this for forward navigation where the user should be able to go back.
func (m *Model) pushScreen(newScreen Screen) {
	// Don't push if we're already on the same screen
	if m.screen == newScreen {
		return
	}
	// Add current screen to stack before navigating
	m.navStack = append(m.navStack, m.screen)
	m.screen = newScreen
}

// popScreen returns to the previous screen in the navigation stack.
// If the stack is empty, it returns to the main menu.
// Returns the screen that was navigated to.
func (m *Model) popScreen() Screen {
	if len(m.navStack) == 0 {
		m.screen = ScreenMainMenu
		return ScreenMainMenu
	}

	// Pop the last screen from the stack
	lastIndex := len(m.navStack) - 1
	previousScreen := m.navStack[lastIndex]
	m.navStack = m.navStack[:lastIndex]
	m.screen = previousScreen
	return previousScreen
}

// clearNavStack clears the navigation stack.
// Use this when starting a new game or returning to a clean state.
func (m *Model) clearNavStack() {
	m.navStack = nil
}

// breadcrumb generates a breadcrumb string showing the navigation path.
// Returns an empty string if at the main menu or if the stack is empty.
func (m Model) breadcrumb() string {
	if m.screen == ScreenMainMenu || len(m.navStack) == 0 {
		return ""
	}

	// Build breadcrumb from the last item in the stack (direct parent)
	// We show just the parent > current for simplicity
	if len(m.navStack) > 0 {
		parent := m.navStack[len(m.navStack)-1]
		return screenName(parent) + " > " + screenName(m.screen)
	}

	return screenName(m.screen)
}

// canGoBack returns true if there's a previous screen to navigate back to.
func (m Model) canGoBack() bool {
	return len(m.navStack) > 0
}
