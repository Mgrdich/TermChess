package config

import (
	"os"
	"testing"
)

// TestLoadConfig_WithMissingFile tests that LoadConfig returns default config when file doesn't exist
// Note: This test temporarily renames the actual config file if it exists
func TestLoadConfig_WithMissingFile(t *testing.T) {
	// Get the actual config file path
	configPath, err := getConfigFilePath()
	if err != nil {
		t.Fatalf("getConfigFilePath failed: %v", err)
	}

	// If config file exists, temporarily rename it
	backupPath := configPath + ".test-backup"
	fileExisted := false
	if _, err := os.Stat(configPath); err == nil {
		fileExisted = true
		if err := os.Rename(configPath, backupPath); err != nil {
			t.Fatalf("Failed to backup config file: %v", err)
		}
		defer func() {
			// Restore the original file
			os.Rename(backupPath, configPath)
		}()
	}

	// LoadConfig should return defaults without error
	config := LoadConfig()

	// Verify default values
	expectedDefaults := DefaultConfig()
	if config.UseUnicode != expectedDefaults.UseUnicode ||
		config.ShowCoords != expectedDefaults.ShowCoords ||
		config.UseColors != expectedDefaults.UseColors ||
		config.ShowMoveHistory != expectedDefaults.ShowMoveHistory {
		t.Error("LoadConfig did not return default config when file is missing")
	}

	// If file existed originally, it will be restored by defer
	_ = fileExisted
}

// TestSaveAndLoadConfig tests the full save and load cycle
func TestSaveAndLoadConfig(t *testing.T) {
	// Create a custom config
	customConfig := Config{
		UseUnicode:      true,
		ShowCoords:      false,
		UseColors:       false,
		ShowMoveHistory: true,
	}

	// Save the config
	if err := SaveConfig(customConfig); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Load the config
	loadedConfig := LoadConfig()

	// Verify the loaded config matches what we saved
	if loadedConfig.UseUnicode != customConfig.UseUnicode {
		t.Errorf("UseUnicode mismatch: got %v, want %v", loadedConfig.UseUnicode, customConfig.UseUnicode)
	}
	if loadedConfig.ShowCoords != customConfig.ShowCoords {
		t.Errorf("ShowCoords mismatch: got %v, want %v", loadedConfig.ShowCoords, customConfig.ShowCoords)
	}
	if loadedConfig.UseColors != customConfig.UseColors {
		t.Errorf("UseColors mismatch: got %v, want %v", loadedConfig.UseColors, customConfig.UseColors)
	}
	if loadedConfig.ShowMoveHistory != customConfig.ShowMoveHistory {
		t.Errorf("ShowMoveHistory mismatch: got %v, want %v", loadedConfig.ShowMoveHistory, customConfig.ShowMoveHistory)
	}
}

// TestSaveConfig_CreatesDirectory tests that SaveConfig creates the config directory if it doesn't exist
func TestSaveConfig_CreatesDirectory(t *testing.T) {
	// Get the config directory path
	configDir, err := getConfigDir()
	if err != nil {
		t.Fatalf("getConfigDir failed: %v", err)
	}

	// The directory should exist after calling SaveConfig (it may already exist)
	defaultConfig := DefaultConfig()
	if err := SaveConfig(defaultConfig); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify the directory exists
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("SaveConfig did not create config directory")
	}
}

// TestConfigFileToConfig tests the conversion from ConfigFile to Config
func TestConfigFileToConfig(t *testing.T) {
	cf := ConfigFile{
		Display: DisplayConfig{
			UseUnicode:      true,
			ShowCoordinates: false,
			UseColors:       false,
			ShowMoveHistory: true,
		},
		Game: GameConfig{
			DefaultGameType:      "pvbot",
			DefaultBotDifficulty: "hard",
		},
	}

	config := configFileToConfig(cf)

	if config.UseUnicode != cf.Display.UseUnicode {
		t.Error("UseUnicode conversion failed")
	}
	if config.ShowCoords != cf.Display.ShowCoordinates {
		t.Error("ShowCoords conversion failed")
	}
	if config.UseColors != cf.Display.UseColors {
		t.Error("UseColors conversion failed")
	}
	if config.ShowMoveHistory != cf.Display.ShowMoveHistory {
		t.Error("ShowMoveHistory conversion failed")
	}
}

// TestConfigToConfigFile tests the conversion from Config to ConfigFile
func TestConfigToConfigFile(t *testing.T) {
	config := Config{
		UseUnicode:      true,
		ShowCoords:      false,
		UseColors:       false,
		ShowMoveHistory: true,
	}

	cf := configToConfigFile(config)

	if cf.Display.UseUnicode != config.UseUnicode {
		t.Error("UseUnicode conversion failed")
	}
	if cf.Display.ShowCoordinates != config.ShowCoords {
		t.Error("ShowCoordinates conversion failed")
	}
	if cf.Display.UseColors != config.UseColors {
		t.Error("UseColors conversion failed")
	}
	if cf.Display.ShowMoveHistory != config.ShowMoveHistory {
		t.Error("ShowMoveHistory conversion failed")
	}

	// Verify game defaults are preserved
	if cf.Game.DefaultGameType != "pvp" {
		t.Error("DefaultGameType should be 'pvp'")
	}
	if cf.Game.DefaultBotDifficulty != "medium" {
		t.Error("DefaultBotDifficulty should be 'medium'")
	}
}

// TestDefaultConfigFile tests that defaultConfigFile returns expected values
func TestDefaultConfigFile(t *testing.T) {
	cf := defaultConfigFile()

	// Verify display defaults
	if cf.Display.UseUnicode != false {
		t.Error("Default UseUnicode should be false")
	}
	if cf.Display.ShowCoordinates != true {
		t.Error("Default ShowCoordinates should be true")
	}
	if cf.Display.UseColors != true {
		t.Error("Default UseColors should be true")
	}
	if cf.Display.ShowMoveHistory != false {
		t.Error("Default ShowMoveHistory should be false")
	}
	if cf.Display.Theme != DefaultTheme {
		t.Errorf("Default Theme should be %q", DefaultTheme)
	}

	// Verify game defaults
	if cf.Game.DefaultGameType != "pvp" {
		t.Error("Default DefaultGameType should be 'pvp'")
	}
	if cf.Game.DefaultBotDifficulty != "medium" {
		t.Error("Default DefaultBotDifficulty should be 'medium'")
	}
}

// TestThemeSaveAndLoad tests that theme setting is saved and loaded correctly
func TestThemeSaveAndLoad(t *testing.T) {
	// Create a config with a specific theme
	customConfig := Config{
		UseUnicode:      false,
		ShowCoords:      true,
		UseColors:       true,
		ShowMoveHistory: false,
		ShowHelpText:    true,
		Theme:           DefaultTheme,
	}

	// Save the config
	if err := SaveConfig(customConfig); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Load the config
	loadedConfig := LoadConfig()

	// Verify the theme was loaded correctly
	if loadedConfig.Theme != customConfig.Theme {
		t.Errorf("Theme mismatch: got %q, want %q", loadedConfig.Theme, customConfig.Theme)
	}
}

// TestThemeDefaultOnEmpty tests that empty theme in config file defaults to DefaultTheme
func TestThemeDefaultOnEmpty(t *testing.T) {
	cf := ConfigFile{
		Display: DisplayConfig{
			UseUnicode:      false,
			ShowCoordinates: true,
			UseColors:       true,
			ShowMoveHistory: false,
			Theme:           "", // Empty theme
		},
		Game: GameConfig{
			DefaultGameType:      "pvp",
			DefaultBotDifficulty: "medium",
		},
	}

	config := configFileToConfig(cf)

	// Empty theme should default to DefaultTheme
	if config.Theme != DefaultTheme {
		t.Errorf("Expected empty theme to default to %q, got %q", DefaultTheme, config.Theme)
	}
}

// TestDefaultConfig_HasTheme tests that DefaultConfig includes theme field
func TestDefaultConfig_HasTheme(t *testing.T) {
	config := DefaultConfig()

	if config.Theme != DefaultTheme {
		t.Errorf("Expected default theme to be %q, got %q", DefaultTheme, config.Theme)
	}
}
