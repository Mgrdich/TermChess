package ui

import (
	"github.com/Mgrdich/TermChess/internal/bot"
	"github.com/Mgrdich/TermChess/internal/bvb"
	"github.com/Mgrdich/TermChess/internal/config"
	"github.com/Mgrdich/TermChess/internal/engine"
	"github.com/charmbracelet/bubbles/textinput"
)

// Screen represents the current UI screen state in the application.
// The application navigates between different screens based on user interaction.
type Screen int

const (
	// ScreenMainMenu is the initial screen showing main menu options
	ScreenMainMenu Screen = iota
	// ScreenGameTypeSelect allows the user to choose between PvP or PvBot
	ScreenGameTypeSelect
	// ScreenBotSelect allows the user to choose bot difficulty
	ScreenBotSelect
	// ScreenColorSelect allows the user to choose their color in bot games
	ScreenColorSelect
	// ScreenFENInput allows the user to load a game from FEN notation
	ScreenFENInput
	// ScreenGamePlay is the main game screen where chess is played
	ScreenGamePlay
	// ScreenGameOver is displayed when the game ends
	ScreenGameOver
	// ScreenSettings allows the user to configure display options
	ScreenSettings
	// ScreenSavePrompt is displayed when the user tries to quit during an active game
	ScreenSavePrompt
	// ScreenResumePrompt is displayed on startup when a saved game exists
	ScreenResumePrompt
	// ScreenDrawPrompt is displayed when one player offers a draw
	ScreenDrawPrompt
	// ScreenBvBBotSelect allows the user to choose bot difficulties for Bot vs Bot mode
	ScreenBvBBotSelect
	// ScreenBvBGameMode allows the user to choose single game or multi-game mode
	ScreenBvBGameMode
	// ScreenBvBGridConfig allows the user to select grid layout for viewing games
	ScreenBvBGridConfig
	// ScreenBvBGamePlay is the main screen for watching Bot vs Bot games
	ScreenBvBGamePlay
	// ScreenBvBStats is displayed after all Bot vs Bot games finish
	ScreenBvBStats
	// ScreenBvBViewModeSelect allows the user to select view mode before starting session
	ScreenBvBViewModeSelect
	// ScreenBvBConcurrencySelect allows the user to choose concurrency for multi-game BvB sessions
	ScreenBvBConcurrencySelect
)

// GameType represents the type of chess game being played.
type GameType int

const (
	// GameTypePvP is a player vs player game
	GameTypePvP GameType = iota
	// GameTypePvBot is a player vs bot game (for future bot support)
	GameTypePvBot
	// GameTypeBvB is a bot vs bot game
	GameTypeBvB
)

// BotDifficulty represents the difficulty level of the chess bot.
// This is defined for future bot implementation.
type BotDifficulty int

const (
	// BotEasy is the easiest bot difficulty level
	BotEasy BotDifficulty = iota
	// BotMedium is the medium bot difficulty level
	BotMedium
	// BotHard is the hardest bot difficulty level
	BotHard
)

// Model is the Bubbletea application model that holds all application state.
// It implements the tea.Model interface (Init, Update, View methods).
type Model struct {
	// Game state
	// board holds the current chess board state from the engine
	board *engine.Board
	// moveHistory stores all moves made in the current game
	moveHistory []engine.Move

	// UI state
	// screen tracks which screen is currently being displayed
	screen Screen
	// navStack tracks the navigation history for back navigation
	navStack []Screen
	// config holds display configuration options
	config Config
	// theme holds the current color theme for UI rendering
	theme Theme
	// termWidth holds the current terminal width in characters
	termWidth int
	// termHeight holds the current terminal height in lines
	termHeight int

	// Input state
	// input holds the current user input text
	input string
	// fenInput holds the text input component for FEN string entry
	fenInput textinput.Model
	// errorMsg holds any error message to display to the user
	errorMsg string
	// statusMsg holds status information to display to the user
	statusMsg string

	// Menu state
	// menuSelection tracks the currently selected menu item index
	menuSelection int
	// menuOptions holds the list of options available in the current menu
	menuOptions []string
	// settingsSelection tracks the currently selected setting in the settings screen
	settingsSelection int
	// savePromptSelection tracks the currently selected option in the save prompt (0=Yes, 1=No)
	savePromptSelection int
	// savePromptAction indicates what action to take after save decision ("exit" or "menu")
	savePromptAction string
	// resumePromptSelection tracks the currently selected option in the resume prompt (0=Yes, 1=No)
	resumePromptSelection int
	// drawPromptSelection tracks the currently selected option in the draw prompt (0=Accept, 1=Decline)
	drawPromptSelection int

	// Game metadata
	// gameType indicates whether this is PvP or PvBot
	gameType GameType
	// botDifficulty stores the selected bot difficulty (for future use)
	botDifficulty BotDifficulty
	// botEngine holds the chess bot engine instance for PvBot games
	botEngine bot.Engine
	// userColor stores the color the user is playing (White or Black) in bot games
	userColor engine.Color
	// resignedBy indicates which player resigned (White, Black, or -1 for no resignation)
	resignedBy int8
	// drawOfferedBy indicates which color offered a draw (-1 if none)
	drawOfferedBy int8
	// drawOfferedByWhite tracks if White has already offered a draw this game
	drawOfferedByWhite bool
	// drawOfferedByBlack tracks if Black has already offered a draw this game
	drawOfferedByBlack bool
	// drawByAgreement indicates if the game ended by draw agreement
	drawByAgreement bool

	// Bot vs Bot fields
	// bvbWhiteDiff stores the selected bot difficulty for White in BvB mode
	bvbWhiteDiff BotDifficulty
	// bvbBlackDiff stores the selected bot difficulty for Black in BvB mode
	bvbBlackDiff BotDifficulty
	// bvbSelectingWhite indicates whether we're selecting the White bot (true) or Black bot (false)
	bvbSelectingWhite bool
	// bvbGameCount stores the number of games to play in multi-game mode
	bvbGameCount int
	// bvbCountInput holds the text input for game count entry
	bvbCountInput string
	// bvbInputtingCount indicates whether we're in text input mode for game count
	bvbInputtingCount bool
	// bvbGridRows stores the number of rows in the grid layout
	bvbGridRows int
	// bvbGridCols stores the number of columns in the grid layout
	bvbGridCols int
	// bvbCustomGridInput holds the text input for custom grid dimensions
	bvbCustomGridInput string
	// bvbInputtingGrid indicates whether we're in text input mode for custom grid
	bvbInputtingGrid bool
	// bvbManager holds the session manager for the current BvB session
	bvbManager *bvb.SessionManager
	// bvbSpeed stores the current playback speed
	bvbSpeed bvb.PlaybackSpeed
	// bvbSelectedGame tracks which game is focused in single view (0-indexed)
	bvbSelectedGame int
	// bvbViewMode tracks whether we're in grid or single-board view
	bvbViewMode BvBViewMode
	// bvbPaused tracks whether games are paused
	bvbPaused bool
	// bvbPageIndex tracks the current page in grid view
	bvbPageIndex int
	// bvbStatsSelection tracks the selected option on the stats screen (0=New Session, 1=Return to Menu)
	bvbStatsSelection int
	// bvbStatsResultsPage tracks the current page of individual results on the stats screen
	bvbStatsResultsPage int
	// bvbJumpInput holds the text input for game number entry during jump navigation
	bvbJumpInput string
	// bvbShowJumpPrompt indicates whether the jump prompt is visible
	bvbShowJumpPrompt bool
	// bvbViewModeSelection tracks the currently selected option in view mode selection screen
	bvbViewModeSelection int
	// bvbRecentCompletions stores the last 5 game completion results for stats-only view
	bvbRecentCompletions []string
	// bvbConcurrencySelection tracks the selected option (0 = Recommended, 1 = Custom)
	bvbConcurrencySelection int
	// bvbCustomConcurrency holds the text input for custom concurrency value
	bvbCustomConcurrency string
	// bvbInputtingConcurrency indicates whether we're in text input mode for custom concurrency
	bvbInputtingConcurrency bool
	// bvbConcurrency stores the selected concurrency value for the session
	bvbConcurrency int

	// Overlay state
	// showShortcutsOverlay indicates whether the keyboard shortcuts help overlay is displayed
	showShortcutsOverlay bool

	// Mouse interaction state
	// selectedSquare holds the currently selected piece's square for mouse interaction
	// nil means no piece is currently selected
	selectedSquare *engine.Square
	// validMoves stores the valid destination squares for the currently selected piece
	// This is computed when a piece is selected and used to validate move execution
	validMoves []engine.Square
	// blinkOn controls the blinking highlight state for selected squares
	// Toggles every 500ms when a piece is selected to create a blinking effect
	blinkOn bool
}

// BvBViewMode represents the display mode for BvB gameplay.
type BvBViewMode int

const (
	// BvBGridView shows multiple boards in a grid
	BvBGridView BvBViewMode = iota
	// BvBSingleView shows a single board with full details
	BvBSingleView
	// BvBStatsOnlyView shows only statistics without board rendering
	BvBStatsOnlyView
)

// NewModel creates and initializes a new Model with the provided configuration.
// The model always starts at the main menu screen.
// If a saved game exists, the menu will include a "Resume Game" option.
func NewModel(config Config) Model {
	// Initialize the text input for FEN entry
	ti := textinput.New()
	ti.Placeholder = "Enter FEN string..."
	ti.CharLimit = 100
	ti.Width = 80

	// Build menu options dynamically based on saved game existence
	menuOptions := buildMainMenuOptions()

	// Load theme based on config
	theme := GetTheme(ParseThemeName(config.Theme))

	return Model{
		// Initialize with nil board (created when starting a new game)
		board:       nil,
		moveHistory: []engine.Move{},

		// Always start at main menu
		screen: ScreenMainMenu,

		// Use the provided configuration
		config: config,

		// Use the loaded theme
		theme: theme,

		// Initialize input state
		input:     "",
		fenInput:  ti,
		errorMsg:  "",
		statusMsg: "",

		// Initialize main menu with dynamic options
		menuSelection: 0,
		menuOptions:   menuOptions,

		// Initialize settings
		settingsSelection: 0,

		// Default game metadata
		gameType:      GameTypePvP,
		botDifficulty: BotEasy,
		resignedBy:    -1, // No resignation

		// Initialize draw offer state
		drawOfferedBy:      -1, // No draw offer
		drawOfferedByWhite: false,
		drawOfferedByBlack: false,
		drawByAgreement:    false,
	}
}

// buildMainMenuOptions constructs the main menu options array.
// If a saved game exists, it includes "Resume Game" at the top of the menu.
func buildMainMenuOptions() []string {
	if config.SaveGameExists() {
		return []string{"Resume Game", "New Game", "Load Game", "Settings", "Exit"}
	}
	return []string{"New Game", "Load Game", "Settings", "Exit"}
}

// View renders the current state of the UI as a string.
// This is called by Bubbletea to display the interface.
// The actual rendering logic is implemented in view.go.
