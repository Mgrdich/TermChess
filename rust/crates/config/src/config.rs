//! Config types, defaults, and TOML load/save.

use std::fs;

use serde::{Deserialize, Serialize};

use crate::error::ConfigError;
use crate::paths::{get_config_dir, get_config_file_path};

/// The default theme name.
///
/// Valid theme values are: `"classic"`, `"modern"`, `"minimalist"`.
/// These must match the `ui` theme-name constants. Invalid theme values are
/// normalized to `DEFAULT_THEME` by the UI layer's `ParseThemeName`.
pub const DEFAULT_THEME: &str = "classic";

/// Display configuration options that control how the UI is rendered.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Config {
    /// Whether to use Unicode chess pieces (♔♕) or ASCII (K, Q).
    pub use_unicode: bool,
    /// Whether to show file/rank labels (a-h, 1-8).
    pub show_coords: bool,
    /// Whether to color piece symbols.
    pub use_colors: bool,
    /// Whether to display the move history panel.
    pub show_move_history: bool,
    /// Whether to display navigation help text at the bottom of screens.
    pub show_help_text: bool,
    /// The name of the color theme to use (e.g. `"classic"`).
    pub theme: String,
}

impl Config {
    /// Returns a `Config` with default values for maximum compatibility and
    /// user-friendliness.
    pub fn default_config() -> Config {
        Config {
            use_unicode: false,               // ASCII for maximum compatibility
            show_coords: true,                // Show a-h, 1-8 labels
            use_colors: true,                 // Use colors if terminal supports
            show_move_history: false,         // Hidden by default
            show_help_text: true,             // Show help text by default
            theme: DEFAULT_THEME.to_string(), // Classic theme by default
        }
    }
}

impl Default for Config {
    fn default() -> Self {
        Config::default_config()
    }
}

/// The structure of the TOML configuration file, with separate `display` and
/// `game` sections.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
#[serde(default)]
pub struct ConfigFile {
    pub display: DisplayConfig,
    pub game: GameConfig,
}

/// Display-related configuration options for the TOML file.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
#[serde(default)]
pub struct DisplayConfig {
    #[serde(rename = "use_unicode")]
    pub use_unicode: bool,
    #[serde(rename = "show_coordinates")]
    pub show_coordinates: bool,
    #[serde(rename = "use_colors")]
    pub use_colors: bool,
    #[serde(rename = "show_move_history")]
    pub show_move_history: bool,
    #[serde(rename = "show_help_text")]
    pub show_help_text: bool,
    #[serde(rename = "theme")]
    pub theme: String,
}

/// Game-related configuration options for the TOML file.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
#[serde(default)]
pub struct GameConfig {
    #[serde(rename = "default_game_type")]
    pub default_game_type: String,
    #[serde(rename = "default_bot_difficulty")]
    pub default_bot_difficulty: String,
    /// Controls how many Bot vs Bot games run simultaneously.
    /// 0 = auto-detect based on CPU count, positive values specify exact count.
    #[serde(rename = "bvb_concurrency")]
    pub bvb_concurrency: i64,
    /// The default view mode for Bot vs Bot sessions.
    /// Valid values: `"grid"`, `"single"`, `"stats_only"`.
    #[serde(rename = "bvb_default_view_mode")]
    pub bvb_default_view_mode: String,
}

/// Returns a `ConfigFile` with default values.
pub(crate) fn default_config_file() -> ConfigFile {
    ConfigFile {
        display: DisplayConfig {
            use_unicode: false,               // ASCII for maximum compatibility
            show_coordinates: true,           // Show a-h, 1-8 labels
            use_colors: true,                 // Use colors if terminal supports
            show_move_history: false,         // Hidden by default
            show_help_text: false,            // Zero value (not set in Go default)
            theme: DEFAULT_THEME.to_string(), // Classic theme by default
        },
        game: GameConfig {
            default_game_type: "pvp".to_string(), // Default to player vs player
            default_bot_difficulty: "medium".to_string(), // Default bot difficulty
            bvb_concurrency: 0,
            bvb_default_view_mode: "grid".to_string(), // Default to grid view for BvB
        },
    }
}

/// Converts a `ConfigFile` to a `Config`.
pub(crate) fn config_file_to_config(cf: &ConfigFile) -> Config {
    let theme = if cf.display.theme.is_empty() {
        DEFAULT_THEME.to_string()
    } else {
        cf.display.theme.clone()
    };
    Config {
        use_unicode: cf.display.use_unicode,
        show_coords: cf.display.show_coordinates,
        use_colors: cf.display.use_colors,
        show_move_history: cf.display.show_move_history,
        show_help_text: cf.display.show_help_text,
        theme,
    }
}

/// Converts a `Config` to a `ConfigFile`.
pub(crate) fn config_to_config_file(c: &Config) -> ConfigFile {
    let theme = if c.theme.is_empty() {
        DEFAULT_THEME.to_string()
    } else {
        c.theme.clone()
    };
    ConfigFile {
        display: DisplayConfig {
            use_unicode: c.use_unicode,
            show_coordinates: c.show_coords,
            use_colors: c.use_colors,
            show_move_history: c.show_move_history,
            show_help_text: c.show_help_text,
            theme,
        },
        game: GameConfig {
            default_game_type: "pvp".to_string(),         // Preserve default
            default_bot_difficulty: "medium".to_string(), // Preserve default
            bvb_concurrency: 0,
            bvb_default_view_mode: "grid".to_string(), // Preserve default
        },
    }
}

/// Reads the configuration file from `~/.termchess/config.toml`.
///
/// If the file doesn't exist or cannot be parsed, returns the default
/// configuration. This function never returns an error — it always returns a
/// valid configuration.
pub fn load_config() -> Config {
    let config_path = match get_config_file_path() {
        Ok(p) => p,
        Err(_) => return Config::default_config(),
    };

    let contents = match fs::read_to_string(&config_path) {
        Ok(c) => c,
        Err(_) => return Config::default_config(),
    };

    match toml::from_str::<ConfigFile>(&contents) {
        Ok(cf) => config_file_to_config(&cf),
        Err(_) => Config::default_config(),
    }
}

/// Reads the game configuration from `~/.termchess/config.toml`.
///
/// If the file doesn't exist or cannot be parsed, returns the default game
/// configuration. This function never returns an error.
pub fn load_game_config() -> GameConfig {
    let config_path = match get_config_file_path() {
        Ok(p) => p,
        Err(_) => return default_config_file().game,
    };

    let contents = match fs::read_to_string(&config_path) {
        Ok(c) => c,
        Err(_) => return default_config_file().game,
    };

    match toml::from_str::<ConfigFile>(&contents) {
        Ok(cf) => cf.game,
        Err(_) => default_config_file().game,
    }
}

/// Writes the configuration to `~/.termchess/config.toml`.
///
/// Creates the `~/.termchess/` directory if it doesn't exist. Returns an error
/// if the file cannot be written.
pub fn save_config(config: &Config) -> Result<(), ConfigError> {
    let config_dir = get_config_dir()?;
    fs::create_dir_all(&config_dir)?;

    let config_path = get_config_file_path()?;
    let cf = config_to_config_file(config);
    let encoded = toml::to_string(&cf)?;
    fs::write(&config_path, encoded)?;

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::fs_test_lock;
    use std::fs as stdfs;

    // TestLoadConfig_WithMissingFile
    #[test]
    fn load_config_with_missing_file() {
        let _guard = fs_test_lock();
        let config_path = get_config_file_path().expect("get_config_file_path failed");

        // If config file exists, temporarily rename it.
        let backup_path = {
            let mut p = config_path.clone().into_os_string();
            p.push(".test-backup");
            std::path::PathBuf::from(p)
        };
        let file_existed = config_path.exists();
        if file_existed {
            stdfs::rename(&config_path, &backup_path).expect("failed to backup config file");
        }

        // LoadConfig should return defaults.
        let config = load_config();
        let expected = Config::default_config();
        assert_eq!(config.use_unicode, expected.use_unicode);
        assert_eq!(config.show_coords, expected.show_coords);
        assert_eq!(config.use_colors, expected.use_colors);
        assert_eq!(config.show_move_history, expected.show_move_history);

        // Restore the original file.
        if file_existed {
            let _ = stdfs::rename(&backup_path, &config_path);
        }
    }

    // TestSaveAndLoadConfig
    #[test]
    fn save_and_load_config() {
        let _guard = fs_test_lock();
        let custom = Config {
            use_unicode: true,
            show_coords: false,
            use_colors: false,
            show_move_history: true,
            show_help_text: false,
            theme: String::new(),
        };

        save_config(&custom).expect("SaveConfig failed");
        let loaded = load_config();

        assert_eq!(
            loaded.use_unicode, custom.use_unicode,
            "UseUnicode mismatch"
        );
        assert_eq!(
            loaded.show_coords, custom.show_coords,
            "ShowCoords mismatch"
        );
        assert_eq!(loaded.use_colors, custom.use_colors, "UseColors mismatch");
        assert_eq!(
            loaded.show_move_history, custom.show_move_history,
            "ShowMoveHistory mismatch"
        );
    }

    // TestSaveConfig_CreatesDirectory
    #[test]
    fn save_config_creates_directory() {
        let _guard = fs_test_lock();
        let config_dir = get_config_dir().expect("GetConfigDir failed");

        let default_config = Config::default_config();
        save_config(&default_config).expect("SaveConfig failed");

        assert!(
            config_dir.exists(),
            "SaveConfig did not create config directory"
        );
    }

    // TestConfigFileToConfig
    #[test]
    fn config_file_to_config_conversion() {
        let cf = ConfigFile {
            display: DisplayConfig {
                use_unicode: true,
                show_coordinates: false,
                use_colors: false,
                show_move_history: true,
                show_help_text: false,
                theme: String::new(),
            },
            game: GameConfig {
                default_game_type: "pvbot".to_string(),
                default_bot_difficulty: "hard".to_string(),
                bvb_concurrency: 0,
                bvb_default_view_mode: String::new(),
            },
        };

        let config = config_file_to_config(&cf);
        assert_eq!(config.use_unicode, cf.display.use_unicode);
        assert_eq!(config.show_coords, cf.display.show_coordinates);
        assert_eq!(config.use_colors, cf.display.use_colors);
        assert_eq!(config.show_move_history, cf.display.show_move_history);
    }

    // TestConfigToConfigFile
    #[test]
    fn config_to_config_file_conversion() {
        let config = Config {
            use_unicode: true,
            show_coords: false,
            use_colors: false,
            show_move_history: true,
            show_help_text: false,
            theme: String::new(),
        };

        let cf = config_to_config_file(&config);
        assert_eq!(cf.display.use_unicode, config.use_unicode);
        assert_eq!(cf.display.show_coordinates, config.show_coords);
        assert_eq!(cf.display.use_colors, config.use_colors);
        assert_eq!(cf.display.show_move_history, config.show_move_history);

        // Game defaults are preserved.
        assert_eq!(cf.game.default_game_type, "pvp");
        assert_eq!(cf.game.default_bot_difficulty, "medium");
    }

    // TestDefaultConfigFile
    #[test]
    fn default_config_file_values() {
        let cf = default_config_file();
        assert!(
            !cf.display.use_unicode,
            "Default UseUnicode should be false"
        );
        assert!(
            cf.display.show_coordinates,
            "Default ShowCoordinates should be true"
        );
        assert!(cf.display.use_colors, "Default UseColors should be true");
        assert!(
            !cf.display.show_move_history,
            "Default ShowMoveHistory should be false"
        );
        assert_eq!(cf.display.theme, DEFAULT_THEME);

        assert_eq!(cf.game.default_game_type, "pvp");
        assert_eq!(cf.game.default_bot_difficulty, "medium");
    }

    // TestThemeSaveAndLoad
    #[test]
    fn theme_save_and_load() {
        let _guard = fs_test_lock();
        let custom = Config {
            use_unicode: false,
            show_coords: true,
            use_colors: true,
            show_move_history: false,
            show_help_text: true,
            theme: DEFAULT_THEME.to_string(),
        };

        save_config(&custom).expect("SaveConfig failed");
        let loaded = load_config();
        assert_eq!(loaded.theme, custom.theme, "Theme mismatch");
    }

    // TestThemeDefaultOnEmpty
    #[test]
    fn theme_default_on_empty() {
        let cf = ConfigFile {
            display: DisplayConfig {
                use_unicode: false,
                show_coordinates: true,
                use_colors: true,
                show_move_history: false,
                show_help_text: false,
                theme: String::new(),
            },
            game: GameConfig {
                default_game_type: "pvp".to_string(),
                default_bot_difficulty: "medium".to_string(),
                bvb_concurrency: 0,
                bvb_default_view_mode: String::new(),
            },
        };

        let config = config_file_to_config(&cf);
        assert_eq!(config.theme, DEFAULT_THEME, "empty theme should default");
    }

    // TestDefaultConfig_HasTheme
    #[test]
    fn default_config_has_theme() {
        let config = Config::default_config();
        assert_eq!(config.theme, DEFAULT_THEME);
    }
}
