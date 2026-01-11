package ui

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGetConfigPath verifies that GetConfigPath returns a valid path
func TestGetConfigPath(t *testing.T) {
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() returned error: %v", err)
	}

	if path == "" {
		t.Fatal("GetConfigPath() returned empty path")
	}

	// Verify path contains .termchess directory
	if !filepath.IsAbs(path) {
		t.Errorf("GetConfigPath() returned non-absolute path: %s", path)
	}

	if filepath.Base(filepath.Dir(path)) != ".termchess" {
		t.Errorf("GetConfigPath() path doesn't contain .termchess directory: %s", path)
	}

	if filepath.Base(path) != "config.toml" {
		t.Errorf("GetConfigPath() doesn't end with config.toml: %s", path)
	}
}

// TestDefaultConfig verifies that DefaultConfig returns expected defaults
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.UseUnicode != false {
		t.Errorf("DefaultConfig().UseUnicode = %v, want false", config.UseUnicode)
	}

	if config.ShowCoords != true {
		t.Errorf("DefaultConfig().ShowCoords = %v, want true", config.ShowCoords)
	}

	if config.UseColors != true {
		t.Errorf("DefaultConfig().UseColors = %v, want true", config.UseColors)
	}

	if config.ShowMoveHistory != false {
		t.Errorf("DefaultConfig().ShowMoveHistory = %v, want false", config.ShowMoveHistory)
	}

	if config.ShowHelpText != true {
		t.Errorf("DefaultConfig().ShowHelpText = %v, want true", config.ShowHelpText)
	}
}

// TestLoadConfigNoFile verifies that LoadConfig returns defaults when no file exists
func TestLoadConfigNoFile(t *testing.T) {
	// Get config path and ensure it doesn't exist
	configPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() returned error: %v", err)
	}

	// Remove config file if it exists
	os.Remove(configPath)

	// Load config should return defaults
	config := LoadConfig()
	defaultConfig := DefaultConfig()

	if config.UseUnicode != defaultConfig.UseUnicode {
		t.Errorf("LoadConfig().UseUnicode = %v, want %v", config.UseUnicode, defaultConfig.UseUnicode)
	}

	if config.ShowCoords != defaultConfig.ShowCoords {
		t.Errorf("LoadConfig().ShowCoords = %v, want %v", config.ShowCoords, defaultConfig.ShowCoords)
	}

	if config.UseColors != defaultConfig.UseColors {
		t.Errorf("LoadConfig().UseColors = %v, want %v", config.UseColors, defaultConfig.UseColors)
	}

	if config.ShowMoveHistory != defaultConfig.ShowMoveHistory {
		t.Errorf("LoadConfig().ShowMoveHistory = %v, want %v", config.ShowMoveHistory, defaultConfig.ShowMoveHistory)
	}
}

// TestSaveAndLoadConfig verifies that SaveConfig writes and LoadConfig reads correctly
func TestSaveAndLoadConfig(t *testing.T) {
	// Create a test config with non-default values
	testConfig := Config{
		UseUnicode:      true,
		ShowCoords:      false,
		UseColors:       false,
		ShowMoveHistory: true,
	}

	// Save the config
	err := SaveConfig(testConfig)
	if err != nil {
		t.Fatalf("SaveConfig() returned error: %v", err)
	}

	// Verify config file was created
	configPath, _ := GetConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created at %s", configPath)
	}

	// Load the config
	loadedConfig := LoadConfig()

	// Verify all fields match
	if loadedConfig.UseUnicode != testConfig.UseUnicode {
		t.Errorf("LoadConfig().UseUnicode = %v, want %v", loadedConfig.UseUnicode, testConfig.UseUnicode)
	}

	if loadedConfig.ShowCoords != testConfig.ShowCoords {
		t.Errorf("LoadConfig().ShowCoords = %v, want %v", loadedConfig.ShowCoords, testConfig.ShowCoords)
	}

	if loadedConfig.UseColors != testConfig.UseColors {
		t.Errorf("LoadConfig().UseColors = %v, want %v", loadedConfig.UseColors, testConfig.UseColors)
	}

	if loadedConfig.ShowMoveHistory != testConfig.ShowMoveHistory {
		t.Errorf("LoadConfig().ShowMoveHistory = %v, want %v", loadedConfig.ShowMoveHistory, testConfig.ShowMoveHistory)
	}

	// Clean up
	os.Remove(configPath)
}

// TestSaveConfigCreatesDirectory verifies that SaveConfig creates the directory if needed
func TestSaveConfigCreatesDirectory(t *testing.T) {
	// Get config path
	configPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() returned error: %v", err)
	}

	// Remove the entire .termchess directory
	configDir := filepath.Dir(configPath)
	os.RemoveAll(configDir)

	// Verify directory doesn't exist
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Fatalf("Config directory still exists at %s", configDir)
	}

	// Save config
	err = SaveConfig(DefaultConfig())
	if err != nil {
		t.Fatalf("SaveConfig() returned error: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Config directory was not created at %s", configDir)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created at %s", configPath)
	}

	// Clean up
	os.RemoveAll(configDir)
}

// TestConfigPersistence verifies complete round-trip persistence
func TestConfigPersistence(t *testing.T) {
	// Clean up first
	configPath, _ := GetConfigPath()
	os.Remove(configPath)

	// Create custom config
	customConfig := Config{
		UseUnicode:      true,
		ShowCoords:      true,
		UseColors:       false,
		ShowMoveHistory: true,
	}

	// Save it
	if err := SaveConfig(customConfig); err != nil {
		t.Fatalf("SaveConfig() failed: %v", err)
	}

	// Load it back
	loadedConfig := LoadConfig()

	// Verify exact match
	if loadedConfig != customConfig {
		t.Errorf("Loaded config doesn't match saved config.\nGot: %+v\nWant: %+v", loadedConfig, customConfig)
	}

	// Clean up
	os.Remove(configPath)
}
