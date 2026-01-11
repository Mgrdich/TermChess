package ui

import (
	"github.com/Mgrdich/TermChess/internal/engine"
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
	// ScreenFENInput allows the user to load a game from FEN notation
	ScreenFENInput
	// ScreenGamePlay is the main game screen where chess is played
	ScreenGamePlay
	// ScreenGameOver is displayed when the game ends
	ScreenGameOver
	// ScreenSettings allows the user to configure display options
	ScreenSettings
)

// GameType represents the type of chess game being played.
type GameType int

const (
	// GameTypePvP is a player vs player game
	GameTypePvP GameType = iota
	// GameTypePvBot is a player vs bot game (for future bot support)
	GameTypePvBot
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
	// config holds display configuration options
	config Config

	// Input state
	// input holds the current user input text
	input string
	// errorMsg holds any error message to display to the user
	errorMsg string
	// statusMsg holds status information to display to the user
	statusMsg string

	// Menu state
	// menuSelection tracks the currently selected menu item index
	menuSelection int
	// menuOptions holds the list of options available in the current menu
	menuOptions []string

	// Settings state
	// settingsSelection tracks the currently selected setting option index
	settingsSelection int

	// Game metadata
	// gameType indicates whether this is PvP or PvBot
	gameType GameType
	// botDifficulty stores the selected bot difficulty (for future use)
	botDifficulty BotDifficulty
}

// NewModel creates and initializes a new Model with default values.
// The model starts at the main menu screen with configuration loaded from file.
// If no config file exists, default values are used.
func NewModel() Model {
	return Model{
		// Initialize with nil board (created when starting a new game)
		board:       nil,
		moveHistory: []engine.Move{},

		// Start at the main menu
		screen: ScreenMainMenu,

		// Load configuration from ~/.termchess/config.toml (or use defaults if not found)
		config: LoadConfig(),

		// Initialize input state
		input:     "",
		errorMsg:  "",
		statusMsg: "",

		// Initialize main menu
		menuSelection: 0,
		menuOptions:   []string{"New Game", "Load Game", "Settings", "Exit"},

		// Initialize settings
		settingsSelection: 0,

		// Default game metadata
		gameType:      GameTypePvP,
		botDifficulty: BotEasy,
	}
}

// View renders the current state of the UI as a string.
// This is called by Bubbletea to display the interface.
// The actual rendering logic is implemented in view.go.
