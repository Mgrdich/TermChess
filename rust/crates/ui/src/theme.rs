//! Color themes for the UI (port of Go `theme.go`).
//!
//! Colors are `ratatui::style::Color`. lipgloss terminal codes ("15", "8",
//! "208") map to `Color::Indexed`; lipgloss hex ("#7D56F4") maps to `Color::Rgb`.

use ratatui::style::Color;

/// A named color theme.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum ThemeName {
    Classic,
    Modern,
    Minimalist,
}

/// Theme name string constants (for serialization/comparison).
pub const THEME_NAME_CLASSIC: &str = "classic";
pub const THEME_NAME_MODERN: &str = "modern";
pub const THEME_NAME_MINIMALIST: &str = "minimalist";

impl ThemeName {
    /// String representation, mirroring Go `ThemeName.String()`.
    pub fn as_str(self) -> &'static str {
        match self {
            ThemeName::Modern => THEME_NAME_MODERN,
            ThemeName::Minimalist => THEME_NAME_MINIMALIST,
            ThemeName::Classic => THEME_NAME_CLASSIC,
        }
    }
}

/// Converts a string to a `ThemeName`, defaulting to Classic (Go `ParseThemeName`).
pub fn parse_theme_name(s: &str) -> ThemeName {
    match s {
        THEME_NAME_MODERN => ThemeName::Modern,
        THEME_NAME_MINIMALIST => ThemeName::Minimalist,
        _ => ThemeName::Classic,
    }
}

/// Parses a "#RRGGBB" string into a `Color::Rgb`.
const fn hex(s: &str) -> Color {
    let b = s.as_bytes();
    // Expect '#' + 6 hex digits.
    let r = (hex_digit(b[1]) << 4) | hex_digit(b[2]);
    let g = (hex_digit(b[3]) << 4) | hex_digit(b[4]);
    let bl = (hex_digit(b[5]) << 4) | hex_digit(b[6]);
    Color::Rgb(r, g, bl)
}

const fn hex_digit(c: u8) -> u8 {
    match c {
        b'0'..=b'9' => c - b'0',
        b'a'..=b'f' => c - b'a' + 10,
        b'A'..=b'F' => c - b'A' + 10,
        _ => 0,
    }
}

/// All color values used throughout the UI.
#[derive(Clone, Copy, Debug)]
pub struct Theme {
    pub name: &'static str,
    pub light_square: Color,
    pub dark_square: Color,
    pub white_piece: Color,
    pub black_piece: Color,
    pub selected_highlight: Color,
    pub valid_move_highlight: Color,
    pub board_border: Color,
    pub menu_selected: Color,
    pub menu_normal: Color,
    pub title_text: Color,
    pub help_text: Color,
    pub error_text: Color,
    pub status_text: Color,
    pub menu_primary: Color,
    pub menu_secondary: Color,
    pub menu_separator: Color,
    pub white_turn_text: Color,
    pub black_turn_text: Color,
}

const CLASSIC: Theme = Theme {
    name: THEME_NAME_CLASSIC,
    light_square: Color::Indexed(15),
    dark_square: Color::Indexed(8),
    white_piece: Color::Indexed(15),
    black_piece: Color::Indexed(8),
    selected_highlight: hex("#7D56F4"),
    valid_move_highlight: hex("#50FA7B"),
    board_border: hex("#FAFAFA"),
    menu_selected: hex("#7D56F4"),
    menu_normal: hex("#FFFDF5"),
    title_text: hex("#FAFAFA"),
    help_text: hex("#626262"),
    error_text: hex("#FF5555"),
    status_text: hex("#50FA7B"),
    menu_primary: hex("#FAFAFA"),
    menu_secondary: hex("#A0A0A0"),
    menu_separator: hex("#444444"),
    white_turn_text: hex("#FAFAFA"),
    black_turn_text: hex("#626262"),
};

const MODERN: Theme = Theme {
    name: THEME_NAME_MODERN,
    light_square: hex("#E8EEF2"),
    dark_square: hex("#5D8AA8"),
    white_piece: hex("#FFFFFF"),
    black_piece: hex("#1A1A2E"),
    selected_highlight: hex("#00A0B0"),
    valid_move_highlight: hex("#4ECDC4"),
    board_border: hex("#B8C5D0"),
    menu_selected: hex("#00A0B0"),
    menu_normal: hex("#E0E0E0"),
    title_text: hex("#E0E0E0"),
    help_text: hex("#8899A6"),
    error_text: hex("#E74C3C"),
    status_text: hex("#4ECDC4"),
    menu_primary: hex("#E0E0E0"),
    menu_secondary: hex("#8899A6"),
    menu_separator: hex("#3D4F5F"),
    white_turn_text: hex("#E0E0E0"),
    black_turn_text: hex("#8899A6"),
};

const MINIMALIST: Theme = Theme {
    name: THEME_NAME_MINIMALIST,
    light_square: hex("#D0D0D0"),
    dark_square: hex("#808080"),
    white_piece: hex("#FFFFFF"),
    black_piece: hex("#2D2D2D"),
    selected_highlight: hex("#A0A0A0"),
    valid_move_highlight: hex("#B8B8B8"),
    board_border: hex("#A0A0A0"),
    menu_selected: hex("#A0A0A0"),
    menu_normal: hex("#C0C0C0"),
    title_text: hex("#C0C0C0"),
    help_text: hex("#707070"),
    error_text: hex("#CC6666"),
    status_text: hex("#88AA88"),
    menu_primary: hex("#C0C0C0"),
    menu_secondary: hex("#888888"),
    menu_separator: hex("#505050"),
    white_turn_text: hex("#C0C0C0"),
    black_turn_text: hex("#707070"),
};

/// Returns the theme for the given name (Go `GetTheme`), defaulting to Classic.
pub fn get_theme(name: ThemeName) -> Theme {
    match name {
        ThemeName::Classic => CLASSIC,
        ThemeName::Modern => MODERN,
        ThemeName::Minimalist => MINIMALIST,
    }
}

/// Display-friendly name for a theme string (Go `getThemeDisplayName`).
pub fn theme_display_name(theme_name: &str) -> &'static str {
    match theme_name {
        THEME_NAME_MODERN => "Modern",
        THEME_NAME_MINIMALIST => "Minimalist",
        _ => "Classic",
    }
}

/// Cycles theme names classic -> modern -> minimalist -> classic (Go `cycleTheme`).
pub fn cycle_theme(current: &str) -> String {
    match current {
        THEME_NAME_CLASSIC => THEME_NAME_MODERN,
        THEME_NAME_MODERN => THEME_NAME_MINIMALIST,
        THEME_NAME_MINIMALIST => THEME_NAME_CLASSIC,
        _ => THEME_NAME_MODERN,
    }
    .to_string()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn parse_defaults_to_classic() {
        assert_eq!(parse_theme_name("classic"), ThemeName::Classic);
        assert_eq!(parse_theme_name("modern"), ThemeName::Modern);
        assert_eq!(parse_theme_name("minimalist"), ThemeName::Minimalist);
        assert_eq!(parse_theme_name("bogus"), ThemeName::Classic);
        assert_eq!(parse_theme_name(""), ThemeName::Classic);
    }

    #[test]
    fn theme_name_roundtrip() {
        for n in [ThemeName::Classic, ThemeName::Modern, ThemeName::Minimalist] {
            assert_eq!(parse_theme_name(n.as_str()), n);
        }
    }

    #[test]
    fn get_theme_names_match() {
        assert_eq!(get_theme(ThemeName::Classic).name, "classic");
        assert_eq!(get_theme(ThemeName::Modern).name, "modern");
        assert_eq!(get_theme(ThemeName::Minimalist).name, "minimalist");
    }

    #[test]
    fn hex_parsing() {
        assert_eq!(hex("#7D56F4"), Color::Rgb(0x7D, 0x56, 0xF4));
        assert_eq!(hex("#FFFFFF"), Color::Rgb(255, 255, 255));
        assert_eq!(hex("#000000"), Color::Rgb(0, 0, 0));
    }

    #[test]
    fn cycle_theme_order() {
        assert_eq!(cycle_theme("classic"), "modern");
        assert_eq!(cycle_theme("modern"), "minimalist");
        assert_eq!(cycle_theme("minimalist"), "classic");
        assert_eq!(cycle_theme("weird"), "modern");
    }

    #[test]
    fn display_names() {
        assert_eq!(theme_display_name("modern"), "Modern");
        assert_eq!(theme_display_name("minimalist"), "Minimalist");
        assert_eq!(theme_display_name("classic"), "Classic");
        assert_eq!(theme_display_name("other"), "Classic");
    }
}
