//! Mouse handling for gameplay (port of Go `mouse.go`).

use crossterm::event::{MouseButton, MouseEvent, MouseEventKind};
use engine::{PieceType, Square};

use crate::app::App;
use crate::state::{GameType, Screen};

/// Width of each square in characters (piece glyph + separating space).
const SQUARE_WIDTH: i32 = 2;

/// Converts terminal mouse coordinates to a chess square (Go `squareFromMouse`).
///
/// `origin` is the screen (column, row) of the board's top-left piece cell —
/// file a, rank 8 — captured from the actual render in `App::draw`. Mapping
/// relative to the real render (rather than hardcoded lipgloss offsets) keeps
/// clicks aligned even if the surrounding layout shifts. Rank 8 is at the top
/// (the board is always drawn from White's perspective; no flip), so rows
/// increase downward as ranks decrease.
pub fn square_from_mouse(x: i32, y: i32, origin: (u16, u16)) -> Option<Square> {
    let origin_x = origin.0 as i32;
    let origin_y = origin.1 as i32;

    if x < origin_x || y < origin_y {
        return None;
    }

    let file = (x - origin_x) / SQUARE_WIDTH;
    let rank = 7 - (y - origin_y);

    if !(0..=7).contains(&file) || !(0..=7).contains(&rank) {
        return None;
    }
    Some(Square::new(file, rank))
}

impl App {
    /// Processes a mouse event during gameplay (Go `handleMouseEvent`).
    pub fn handle_mouse_event(&mut self, msg: MouseEvent) {
        if msg.kind != MouseEventKind::Down(MouseButton::Left) {
            return;
        }

        // The board origin is recorded during `draw`; if the gameplay board has
        // not been drawn yet, ignore the click.
        let origin = match self.board_origin.get() {
            Some(o) => o,
            None => return,
        };

        let sq = match square_from_mouse(msg.column as i32, msg.row as i32, origin) {
            Some(s) => s,
            None => return,
        };

        let board = match &self.board {
            Some(b) => b,
            None => return,
        };

        // For PvBot, only allow interaction on the human's turn.
        if self.game_type == GameType::PvBot && board.active_color != self.user_color {
            return;
        }

        let piece = board.piece_at(sq);

        if self.selected_square.is_some() && self.is_valid_move_destination(sq) {
            self.execute_mouse_move(sq);
            return;
        }

        if !piece.is_empty() && piece.color() == board.active_color {
            self.selected_square = Some(sq);
            self.compute_valid_moves();
            self.blink_on = true;
            self.spawn_blink_tick();
        }
    }

    /// Populates `valid_moves` for the selected piece (Go `computeValidMoves`).
    pub fn compute_valid_moves(&mut self) {
        let (sel, board) = match (self.selected_square, &self.board) {
            (Some(s), Some(b)) => (s, b),
            _ => {
                self.valid_moves.clear();
                return;
            }
        };
        let mut moves = Vec::new();
        for m in board.legal_moves() {
            if m.from == sel {
                moves.push(m.to);
            }
        }
        self.valid_moves = moves;
    }

    fn is_valid_move_destination(&self, sq: Square) -> bool {
        self.valid_moves.contains(&sq)
    }

    /// Executes a move from the selection to `destination` (Go `executeMouseMove`).
    fn execute_mouse_move(&mut self, destination: Square) {
        let sel = match self.selected_square {
            Some(s) => s,
            None => return,
        };

        // Find the matching legal move (prefer queen promotion).
        let mut matching = None;
        if let Some(board) = &self.board {
            for m in board.legal_moves() {
                if m.from == sel && m.to == destination {
                    if matching.is_none() {
                        matching = Some(m);
                    }
                    if m.promotion == PieceType::Queen {
                        matching = Some(m);
                        break;
                    }
                }
            }
        }

        let matching = match matching {
            Some(m) => m,
            None => {
                self.error_msg = "Invalid move".to_string();
                return;
            }
        };

        if let Some(board) = self.board.as_mut() {
            if let Err(e) = board.make_move(matching) {
                self.error_msg = e.to_string();
                return;
            }
        }

        self.selected_square = None;
        self.valid_moves.clear();
        self.blink_on = false;
        self.error_msg.clear();
        self.status_msg.clear();
        self.input.clear();
        self.move_history.push(matching);

        if self
            .board
            .as_ref()
            .map(|b| b.is_game_over())
            .unwrap_or(false)
        {
            self.screen = Screen::GameOver;
            let _ = config::delete_save_game();
            return;
        }

        if self.game_type == GameType::PvBot {
            self.make_bot_move();
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::event::AppEvent;
    use config::Config;
    use engine::Board;
    use ratatui::backend::TestBackend;
    use ratatui::Terminal;
    use std::sync::mpsc::channel;

    // With coordinates shown and the board drawn at screen (0,0), the a8 cell lands
    // at (col=2, row=2): 2 header lines + a 2-char rank label. Derived from the real
    // gameplay render (see `origin_matches_rendered_board`).
    const ORIGIN_COORDS: (u16, u16) = (2, 2);
    // Without coordinates the rank label is gone, so a8 lands at (col=0, row=2).
    const ORIGIN_NO_COORDS: (u16, u16) = (0, 2);

    #[test]
    fn square_from_mouse_with_coords() {
        // Top-left cell (a8).
        assert_eq!(
            square_from_mouse(2, 2, ORIGIN_COORDS),
            Some(Square::new(0, 7))
        );
        // Bottom-right cell (h1): file 7 -> x = 2 + 7*2 = 16, rank 0 -> y = 2 + 7 = 9.
        assert_eq!(
            square_from_mouse(16, 9, ORIGIN_COORDS),
            Some(Square::new(7, 0))
        );
        // e4: file 4, rank 3 -> x = 2 + 4*2 = 10, y = 2 + (7-3) = 6.
        assert_eq!(
            square_from_mouse(10, 6, ORIGIN_COORDS),
            Some(Square::new(4, 3))
        );
    }

    #[test]
    fn square_from_mouse_no_coords() {
        // Top-left cell (a8).
        assert_eq!(
            square_from_mouse(0, 2, ORIGIN_NO_COORDS),
            Some(Square::new(0, 7))
        );
        // Bottom-right cell (h1): file 7 -> x = 14, rank 0 -> y = 9.
        assert_eq!(
            square_from_mouse(14, 9, ORIGIN_NO_COORDS),
            Some(Square::new(7, 0))
        );
    }

    #[test]
    fn square_from_mouse_out_of_bounds() {
        // Above the board.
        assert_eq!(square_from_mouse(2, 0, ORIGIN_COORDS), None);
        // Left of the board (in the rank-label column).
        assert_eq!(square_from_mouse(1, 2, ORIGIN_COORDS), None);
        // Far off the board.
        assert_eq!(square_from_mouse(100, 100, ORIGIN_NO_COORDS), None);
    }

    fn gameplay_app(show_coords: bool) -> App {
        let (tx, _rx) = channel::<AppEvent>();
        let cfg = Config {
            use_unicode: false,
            show_coords,
            use_colors: false,
            show_move_history: false,
            show_help_text: true,
            theme: "classic".to_string(),
        };
        let mut app = App::new(cfg, tx);
        app.screen = Screen::GamePlay;
        app.game_type = GameType::PvP;
        app.board = Some(Board::new());
        app
    }

    /// Render the real gameplay screen into a buffer and confirm the recorded
    /// `board_origin` actually points at the rank-8 pieces, and that clicks at the
    /// board corners resolve to a8 / h1. This ties the mapping to the true render,
    /// so any future header/layout change fails here instead of silently breaking
    /// mouse clicks.
    fn assert_origin_matches_render(show_coords: bool) {
        let app = gameplay_app(show_coords);
        let backend = TestBackend::new(40, 24);
        let mut terminal = Terminal::new(backend).unwrap();
        terminal.draw(|f| app.draw(f)).unwrap();

        let origin = app.board_origin.get().expect("board origin recorded");
        let buffer = terminal.backend().buffer();

        // The rank-8 back rank starts with a rook ('r' for black in the plain ASCII
        // render). Read the glyph at the recorded origin and confirm it is that rook.
        let cell = buffer.cell((origin.0, origin.1)).unwrap();
        assert_eq!(cell.symbol(), "r", "origin should sit on the a8 piece");

        // Corner clicks map to the expected squares.
        assert_eq!(
            square_from_mouse(origin.0 as i32, origin.1 as i32, origin),
            Some(Square::new(0, 7)),
            "top-left click -> a8"
        );
        assert_eq!(
            square_from_mouse(
                origin.0 as i32 + 7 * SQUARE_WIDTH,
                origin.1 as i32 + 7,
                origin
            ),
            Some(Square::new(7, 0)),
            "bottom-right click -> h1"
        );
    }

    #[test]
    fn origin_matches_rendered_board_with_coords() {
        assert_origin_matches_render(true);
    }

    #[test]
    fn origin_matches_rendered_board_no_coords() {
        assert_origin_matches_render(false);
    }

    /// A click on a friendly piece selects it, starts the blink, and computes valid
    /// moves (Go `handleMouseEvent` selection path).
    #[test]
    fn click_selects_own_piece() {
        use crossterm::event::{MouseButton, MouseEvent, MouseEventKind};

        let mut app = gameplay_app(true);
        // Record the origin as draw() would.
        app.board_origin.set(Some(ORIGIN_COORDS));

        // Click e2 (white pawn): file 4, rank 1 -> x = 2 + 8 = 10, y = 2 + 6 = 8.
        let ev = MouseEvent {
            kind: MouseEventKind::Down(MouseButton::Left),
            column: 10,
            row: 8,
            modifiers: crossterm::event::KeyModifiers::empty(),
        };
        app.handle_mouse_event(ev);

        assert_eq!(app.selected_square, Some(Square::new(4, 1)));
        assert!(app.blink_on);
        // e2 pawn can advance to e3 and e4.
        assert!(app.valid_moves.contains(&Square::new(4, 2)));
        assert!(app.valid_moves.contains(&Square::new(4, 3)));
    }

    /// With no recorded origin (board not drawn yet), clicks are ignored.
    #[test]
    fn click_ignored_without_origin() {
        use crossterm::event::{MouseButton, MouseEvent, MouseEventKind};

        let mut app = gameplay_app(true);
        app.board_origin.set(None);
        let ev = MouseEvent {
            kind: MouseEventKind::Down(MouseButton::Left),
            column: 10,
            row: 8,
            modifiers: crossterm::event::KeyModifiers::empty(),
        };
        app.handle_mouse_event(ev);
        assert_eq!(app.selected_square, None);
    }
}
