//! Mouse handling for gameplay (port of Go `mouse.go`).

use crossterm::event::{MouseButton, MouseEvent, MouseEventKind};
use engine::{PieceType, Square};

use crate::app::App;
use crate::state::{GameType, Screen};

/// Terminal row where the board's first rank (rank 8) is rendered.
const BOARD_START_Y: i32 = 4;
/// Column of the first piece when coordinates are shown.
const BOARD_START_X_WITH_COORDS: i32 = 2;
/// Column of the first piece when coordinates are hidden.
const BOARD_START_X_NO_COORDS: i32 = 0;
/// Width of each square in characters.
const SQUARE_WIDTH: i32 = 2;

/// Converts mouse coordinates to a chess square (Go `squareFromMouse`).
pub fn square_from_mouse(x: i32, y: i32, show_coords: bool) -> Option<Square> {
    let board_start_x = if show_coords {
        BOARD_START_X_WITH_COORDS
    } else {
        BOARD_START_X_NO_COORDS
    };

    if x < board_start_x || y < BOARD_START_Y {
        return None;
    }

    let file = (x - board_start_x) / SQUARE_WIDTH;
    let rank = 7 - (y - BOARD_START_Y);

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

        let sq = match square_from_mouse(msg.column as i32, msg.row as i32, self.config.show_coords)
        {
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

    #[test]
    fn square_from_mouse_with_coords() {
        // Board starts at x=2, y=4 with coords. Top-left cell (a8) = (2,4).
        assert_eq!(square_from_mouse(2, 4, true), Some(Square::new(0, 7)));
        // e4: file 4, rank 3 -> x = 2 + 4*2 = 10, y = 4 + (7-3) = 8.
        assert_eq!(square_from_mouse(10, 8, true), Some(Square::new(4, 3)));
    }

    #[test]
    fn square_from_mouse_no_coords() {
        assert_eq!(square_from_mouse(0, 4, false), Some(Square::new(0, 7)));
        assert_eq!(square_from_mouse(14, 11, false), Some(Square::new(7, 0)));
    }

    #[test]
    fn square_from_mouse_out_of_bounds() {
        assert_eq!(square_from_mouse(0, 0, true), None);
        assert_eq!(square_from_mouse(1, 4, true), None);
        assert_eq!(square_from_mouse(100, 100, false), None);
    }
}
