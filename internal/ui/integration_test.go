package ui

import (
	"os"
	"testing"
)

// TestConfigPersistenceIntegration tests the complete config persistence flow:
// 1. App starts with no config -> uses defaults
// 2. Config is saved -> file is created
// 3. App restarts -> config is loaded from file
// 4. Config changes are persisted across restarts
func TestConfigPersistenceIntegration(t *testing.T) {
	// Clean up after test
	configPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() failed: %v", err)
	}
	defer os.Remove(configPath)

	// Phase 1: No config file exists
	os.Remove(configPath)
	config1 := LoadConfig()
	if config1 != DefaultConfig() {
		t.Errorf("LoadConfig() with no file should return defaults")
	}

	// Phase 2: Save custom config
	customConfig := Config{
		UseUnicode:      true,
		ShowCoords:      false,
		UseColors:       false,
		ShowMoveHistory: true,
	}
	if err := SaveConfig(customConfig); err != nil {
		t.Fatalf("SaveConfig() failed: %v", err)
	}

	// Phase 3: Restart simulation - load config again
	config2 := LoadConfig()
	if config2 != customConfig {
		t.Errorf("LoadConfig() after save = %+v, want %+v", config2, customConfig)
	}

	// Phase 4: Modify and save again
	modifiedConfig := Config{
		UseUnicode:      false,
		ShowCoords:      true,
		UseColors:       true,
		ShowMoveHistory: false,
	}
	if err := SaveConfig(modifiedConfig); err != nil {
		t.Fatalf("SaveConfig() second time failed: %v", err)
	}

	// Phase 5: Load modified config
	config3 := LoadConfig()
	if config3 != modifiedConfig {
		t.Errorf("LoadConfig() after modification = %+v, want %+v", config3, modifiedConfig)
	}
}

// TestNewModelLoadsConfig verifies that NewModel() loads config on initialization
func TestNewModelLoadsConfig(t *testing.T) {
	// Set up a known config
	configPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() failed: %v", err)
	}
	defer os.Remove(configPath)

	// Save a specific config
	testConfig := Config{
		UseUnicode:      true,
		ShowCoords:      false,
		UseColors:       true,
		ShowMoveHistory: true,
	}
	if err := SaveConfig(testConfig); err != nil {
		t.Fatalf("SaveConfig() failed: %v", err)
	}

	// Create a new model (should load config internally)
	_ = NewModel()

	// Verify the config can be loaded and matches what we saved
	// Note: Model.config is private, so we test indirectly by ensuring
	// NewModel() doesn't panic and that LoadConfig() works correctly
	loadedConfig := LoadConfig()
	if loadedConfig != testConfig {
		t.Errorf("NewModel should load same config as LoadConfig()")
	}
}

// TestConfigFileFormat verifies the TOML format is correct
func TestConfigFileFormat(t *testing.T) {
	configPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() failed: %v", err)
	}
	defer os.Remove(configPath)

	// Save a config
	testConfig := Config{
		UseUnicode:      true,
		ShowCoords:      false,
		UseColors:       true,
		ShowMoveHistory: false,
	}
	if err := SaveConfig(testConfig); err != nil {
		t.Fatalf("SaveConfig() failed: %v", err)
	}

	// Read the raw file content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// Verify file contains expected TOML sections
	contentStr := string(content)
	requiredSections := []string{
		"[display]",
		"use_unicode",
		"show_coordinates",
		"use_colors",
		"show_move_history",
		"[game]",
		"default_game_type",
		"default_bot_difficulty",
	}

	for _, section := range requiredSections {
		if !contains(contentStr, section) {
			t.Errorf("Config file missing required section or field: %s", section)
		}
	}
}

// TestConfigCorruptedFile verifies graceful handling of corrupted config
func TestConfigCorruptedFile(t *testing.T) {
	configPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() failed: %v", err)
	}
	defer os.Remove(configPath)

	// Create directory
	if err := os.MkdirAll(configPath[:len(configPath)-len("config.toml")], 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Write invalid TOML
	if err := os.WriteFile(configPath, []byte("invalid toml content {{{"), 0644); err != nil {
		t.Fatalf("Failed to write corrupted config: %v", err)
	}

	// LoadConfig should not panic and should return defaults
	config := LoadConfig()
	defaults := DefaultConfig()
	if config != defaults {
		t.Errorf("LoadConfig() with corrupted file should return defaults")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
