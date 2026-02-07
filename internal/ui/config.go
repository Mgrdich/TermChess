package ui

import "github.com/Mgrdich/TermChess/internal/config"

// Type aliases for backward compatibility during refactoring.
// These allow UI code to continue using Config types while
// the actual definitions live in the config package.

// Config holds display configuration options that control how the UI is rendered.
type Config = config.Config

// ConfigFile represents the structure of the TOML configuration file.
type ConfigFile = config.ConfigFile

// DisplayConfig holds display-related configuration options for the TOML file.
type DisplayConfig = config.DisplayConfig

// GameConfig holds game-related configuration options for the TOML file.
type GameConfig = config.GameConfig

// Function aliases for backward compatibility

// DefaultConfig returns a Config with default values for maximum compatibility
// and user-friendliness.
var DefaultConfig = config.DefaultConfig

// GetConfigPath returns the absolute path to the configuration file.
// The config file is stored at ~/.termchess/config.toml
var GetConfigPath = config.GetConfigPath

// LoadConfig reads the configuration file from ~/.termchess/config.toml.
// If the file doesn't exist or cannot be parsed, it returns the default configuration.
var LoadConfig = config.LoadConfig

// LoadGameConfig reads the game configuration from ~/.termchess/config.toml.
// If the file doesn't exist or cannot be parsed, it returns the default game configuration.
var LoadGameConfig = config.LoadGameConfig

// SaveConfig writes the configuration to ~/.termchess/config.toml.
var SaveConfig = config.SaveConfig
