//! Chess board rendering (port of Go `board.go`).
//!
//! `render` produces a structured `ratatui::text::Text` with styled spans;
//! `render_plain` produces the same layout as an unstyled string (used for
//! coordinate math and tests).

use config::Config;
use engine::{Board, Color, Piece, PieceType, Square};
use ratatui::style::{Color as TColor, Modifier, Style};
use ratatui::text::{Line, Span, Text};

use crate::theme::Theme;

/// Renders a chess board for display.
pub struct BoardRenderer<'a> {
    config: &'a Config,
    theme: Theme,
}

impl<'a> BoardRenderer<'a> {
    /// Creates a renderer using the theme parsed from the config.
    pub fn new(config: &'a Config) -> Self {
        let theme = crate::theme::get_theme(crate::theme::parse_theme_name(&config.theme));
        BoardRenderer { config, theme }
    }

    /// Creates a renderer with an explicit theme.
    pub fn with_theme(config: &'a Config, theme: Theme) -> Self {
        BoardRenderer { config, theme }
    }

    /// Renders the board as styled text with optional selection highlighting.
    pub fn render(
        &self,
        b: &Board,
        selected: Option<Square>,
        valid_moves: &[Square],
        blink_on: bool,
    ) -> Text<'static> {
        let mut lines: Vec<Line<'static>> = Vec::new();

        for rank in (0..8).rev() {
            let mut spans: Vec<Span<'static>> = Vec::new();
            if self.config.show_coords {
                spans.push(Span::raw(format!("{} ", rank + 1)));
            }

            for file in 0..8 {
                let sq = Square::new(file, rank);
                let piece = b.piece_at(sq);
                let symbol = self.piece_symbol_str(piece);

                let mut style = self.piece_style(piece);

                if blink_on {
                    if selected == Some(sq) {
                        style = style.bg(self.theme.selected_highlight);
                    } else if valid_moves.contains(&sq) {
                        style = style.bg(self.theme.valid_move_highlight);
                    }
                }

                if file > 0 {
                    spans.push(Span::raw(" "));
                }
                spans.push(Span::styled(symbol, style));
            }

            lines.push(Line::from(spans));
        }

        if self.config.show_coords {
            lines.push(Line::from("  a b c d e f g h".to_string()));
        }

        Text::from(lines)
    }

    /// Renders the board as an unstyled string (matches the Go textual layout).
    pub fn render_plain(&self, b: Option<&Board>) -> String {
        let b = match b {
            Some(b) => b,
            None => return "No board available".to_string(),
        };

        let mut result = String::new();
        for rank in (0..8).rev() {
            if self.config.show_coords {
                result.push_str(&format!("{} ", rank + 1));
            }
            for file in 0..8 {
                let sq = Square::new(file, rank);
                let piece = b.piece_at(sq);
                if file > 0 {
                    result.push(' ');
                }
                result.push_str(&self.piece_symbol_str(piece));
            }
            result.push('\n');
        }
        if self.config.show_coords {
            result.push_str("  a b c d e f g h");
        }
        result
    }

    fn piece_style(&self, p: Piece) -> Style {
        if self.config.use_colors && !p.is_empty() {
            if p.color() == Color::White {
                Style::default()
                    .fg(TColor::Indexed(15))
                    .add_modifier(Modifier::BOLD)
            } else {
                Style::default().fg(TColor::Indexed(8))
            }
        } else {
            Style::default()
        }
    }

    fn piece_symbol_str(&self, p: Piece) -> String {
        if self.config.use_unicode {
            unicode_symbol(p).to_string()
        } else {
            ascii_symbol(p).to_string()
        }
    }
}

fn ascii_symbol(p: Piece) -> char {
    let ch = match p.piece_type() {
        PieceType::Pawn => 'P',
        PieceType::Knight => 'N',
        PieceType::Bishop => 'B',
        PieceType::Rook => 'R',
        PieceType::Queen => 'Q',
        PieceType::King => 'K',
        _ => return '.',
    };
    if p.color() == Color::Black {
        ch.to_ascii_lowercase()
    } else {
        ch
    }
}

fn unicode_symbol(p: Piece) -> char {
    let pt = p.piece_type();
    if pt == PieceType::Empty {
        return '·';
    }
    if p.color() == Color::White {
        match pt {
            PieceType::King => '♔',
            PieceType::Queen => '♕',
            PieceType::Rook => '♖',
            PieceType::Bishop => '♗',
            PieceType::Knight => '♘',
            PieceType::Pawn => '♙',
            _ => '?',
        }
    } else {
        match pt {
            PieceType::King => '♚',
            PieceType::Queen => '♛',
            PieceType::Rook => '♜',
            PieceType::Bishop => '♝',
            PieceType::Knight => '♞',
            PieceType::Pawn => '♟',
            _ => '?',
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn cfg(unicode: bool, coords: bool) -> Config {
        Config {
            use_unicode: unicode,
            show_coords: coords,
            use_colors: false,
            show_move_history: false,
            show_help_text: true,
            theme: "classic".to_string(),
        }
    }

    #[test]
    fn ascii_starting_position() {
        let board = Board::new();
        let c = cfg(false, true);
        let r = BoardRenderer::new(&c);
        let out = r.render_plain(Some(&board));
        assert!(out.contains("r n b q k b n r"));
        assert!(out.contains("R N B Q K B N R"));
        assert!(out.contains("a b c d e f g h"));
        for rank in 1..=8 {
            assert!(out.contains(&rank.to_string()));
        }
    }

    #[test]
    fn nil_board() {
        let c = cfg(false, false);
        let r = BoardRenderer::new(&c);
        assert_eq!(r.render_plain(None), "No board available");
    }

    #[test]
    fn ascii_symbols() {
        assert_eq!(ascii_symbol(Piece::new(Color::White, PieceType::King)), 'K');
        assert_eq!(
            ascii_symbol(Piece::new(Color::Black, PieceType::Queen)),
            'q'
        );
        assert_eq!(ascii_symbol(Piece::EMPTY), '.');
    }

    #[test]
    fn unicode_symbols() {
        assert_eq!(
            unicode_symbol(Piece::new(Color::White, PieceType::King)),
            '♔'
        );
        assert_eq!(
            unicode_symbol(Piece::new(Color::Black, PieceType::Pawn)),
            '♟'
        );
        assert_eq!(unicode_symbol(Piece::EMPTY), '·');
    }

    #[test]
    fn no_coords_omits_labels() {
        let board = Board::new();
        let c = cfg(false, false);
        let r = BoardRenderer::new(&c);
        let out = r.render_plain(Some(&board));
        assert!(!out.contains("a b c d e f g h"));
    }
}
