//! Game-status determination: checkmate, stalemate, and the various draw rules.

use std::fmt;

use crate::board::Board;
use crate::types::{Color, PieceType, Square};

/// The current state of a chess game.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum GameStatus {
    /// The game is still in progress.
    Ongoing,
    /// The player to move is checkmated; the opponent wins.
    Checkmate,
    /// The player to move has no legal moves but is not in check (draw).
    Stalemate,
    /// Draw due to insufficient material to checkmate.
    DrawInsufficientMaterial,
    /// Claimable draw under the fifty-move rule.
    DrawFiftyMoveRule,
    /// Automatic draw under the seventy-five-move rule.
    DrawSeventyFiveMoveRule,
    /// Claimable draw due to threefold repetition.
    DrawThreefoldRepetition,
    /// Automatic draw due to fivefold repetition.
    DrawFivefoldRepetition,
}

impl fmt::Display for GameStatus {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        let s = match self {
            GameStatus::Ongoing => "ongoing",
            GameStatus::Checkmate => "checkmate",
            GameStatus::Stalemate => "stalemate",
            GameStatus::DrawInsufficientMaterial => "draw (insufficient material)",
            GameStatus::DrawFiftyMoveRule => "draw (fifty-move rule)",
            GameStatus::DrawSeventyFiveMoveRule => "draw (seventy-five-move rule)",
            GameStatus::DrawThreefoldRepetition => "draw (threefold repetition)",
            GameStatus::DrawFivefoldRepetition => "draw (fivefold repetition)",
        };
        write!(f, "{}", s)
    }
}

/// Per-color material tally used for insufficient-material detection.
#[derive(Default)]
struct MaterialCount {
    white_pawns: i32,
    white_knights: i32,
    white_bishops: i32,
    white_rooks: i32,
    white_queens: i32,
    black_pawns: i32,
    black_knights: i32,
    black_bishops: i32,
    black_rooks: i32,
    black_queens: i32,
    white_bishop_squares: Vec<Square>,
    black_bishop_squares: Vec<Square>,
}

impl Board {
    /// Determines the current game status, in priority order.
    pub fn status(&self) -> GameStatus {
        let legal_moves = self.legal_moves();

        if legal_moves.is_empty() {
            if self.in_check() {
                return GameStatus::Checkmate;
            }
            return GameStatus::Stalemate;
        }

        // Automatic draws first (these end the game).
        let rep_count = self.repetition_count();
        if rep_count >= 5 {
            return GameStatus::DrawFivefoldRepetition;
        }

        if self.half_move_clock >= 150 {
            return GameStatus::DrawSeventyFiveMoveRule;
        }

        if self.has_insufficient_material() {
            return GameStatus::DrawInsufficientMaterial;
        }

        // Claimable draws.
        if rep_count >= 3 {
            return GameStatus::DrawThreefoldRepetition;
        }

        if self.half_move_clock >= 100 {
            return GameStatus::DrawFiftyMoveRule;
        }

        GameStatus::Ongoing
    }

    /// Returns true if the game has ended due to an automatic (non-claimable) condition.
    pub fn is_game_over(&self) -> bool {
        matches!(
            self.status(),
            GameStatus::Checkmate
                | GameStatus::Stalemate
                | GameStatus::DrawFivefoldRepetition
                | GameStatus::DrawSeventyFiveMoveRule
                | GameStatus::DrawInsufficientMaterial
        )
    }

    /// Returns true if a draw can currently be claimed (threefold or fifty-move).
    pub fn can_claim_draw(&self) -> bool {
        matches!(
            self.status(),
            GameStatus::DrawThreefoldRepetition | GameStatus::DrawFiftyMoveRule
        )
    }

    /// Returns the winning color if the position is checkmate, otherwise `None`.
    pub fn winner(&self) -> Option<Color> {
        if self.status() == GameStatus::Checkmate {
            // The player to move is checkmated, so the opponent wins.
            Some(self.active_color.opponent())
        } else {
            None
        }
    }

    /// Counts how many times the current position occurs in the game history.
    fn repetition_count(&self) -> i32 {
        self.history.iter().filter(|&&h| h == self.hash).count() as i32
    }

    /// Counts all pieces on the board (excluding kings).
    fn count_material(&self) -> MaterialCount {
        let mut mc = MaterialCount::default();

        for sq_idx in 0..64i8 {
            let piece = self.squares[sq_idx as usize];
            if piece.is_empty() {
                continue;
            }
            let sq = Square(sq_idx);
            match (piece.color(), piece.piece_type()) {
                (Color::White, PieceType::Pawn) => mc.white_pawns += 1,
                (Color::White, PieceType::Knight) => mc.white_knights += 1,
                (Color::White, PieceType::Bishop) => {
                    mc.white_bishops += 1;
                    mc.white_bishop_squares.push(sq);
                }
                (Color::White, PieceType::Rook) => mc.white_rooks += 1,
                (Color::White, PieceType::Queen) => mc.white_queens += 1,
                (Color::Black, PieceType::Pawn) => mc.black_pawns += 1,
                (Color::Black, PieceType::Knight) => mc.black_knights += 1,
                (Color::Black, PieceType::Bishop) => {
                    mc.black_bishops += 1;
                    mc.black_bishop_squares.push(sq);
                }
                (Color::Black, PieceType::Rook) => mc.black_rooks += 1,
                (Color::Black, PieceType::Queen) => mc.black_queens += 1,
                _ => {}
            }
        }

        mc
    }

    /// Returns true if neither side can force checkmate.
    fn has_insufficient_material(&self) -> bool {
        let mc = self.count_material();

        if mc.white_pawns > 0
            || mc.black_pawns > 0
            || mc.white_rooks > 0
            || mc.black_rooks > 0
            || mc.white_queens > 0
            || mc.black_queens > 0
        {
            return false;
        }

        let white_minor = mc.white_knights + mc.white_bishops;
        let black_minor = mc.black_knights + mc.black_bishops;

        // K vs K.
        if white_minor == 0 && black_minor == 0 {
            return true;
        }

        // K+B vs K.
        if white_minor == 1 && mc.white_bishops == 1 && black_minor == 0 {
            return true;
        }
        if black_minor == 1 && mc.black_bishops == 1 && white_minor == 0 {
            return true;
        }

        // K+N vs K.
        if white_minor == 1 && mc.white_knights == 1 && black_minor == 0 {
            return true;
        }
        if black_minor == 1 && mc.black_knights == 1 && white_minor == 0 {
            return true;
        }

        // K+B vs K+B with same-color bishops.
        if mc.white_bishops == 1 && mc.black_bishops == 1 && white_minor == 1 && black_minor == 1 {
            let wb = mc.white_bishop_squares[0];
            let bb = mc.black_bishop_squares[0];
            let white_color = (wb.rank() + wb.file()) % 2;
            let black_color = (bb.rank() + bb.file()) % 2;
            if white_color == black_color {
                return true;
            }
        }

        false
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::moves::Move;
    use crate::types::Piece;

    fn sq_from(notation: &str) -> Square {
        let bytes = notation.as_bytes();
        Square::new(bytes[0] as i32 - b'a' as i32, bytes[1] as i32 - b'1' as i32)
    }

    fn setup(pieces: &[(&str, Color, PieceType)], active: Color) -> Board {
        let mut b = Board::default();
        for &(notation, color, pt) in pieces {
            b.squares[sq_from(notation).index()] = Piece::new(color, pt);
        }
        b.active_color = active;
        b.castling_rights = 0;
        b.en_passant_sq = -1;
        b.half_move_clock = 0;
        b.full_move_num = 1;
        b
    }

    #[test]
    fn game_status_display() {
        assert_eq!(GameStatus::Ongoing.to_string(), "ongoing");
        assert_eq!(GameStatus::Checkmate.to_string(), "checkmate");
        assert_eq!(GameStatus::Stalemate.to_string(), "stalemate");
        assert_eq!(
            GameStatus::DrawInsufficientMaterial.to_string(),
            "draw (insufficient material)"
        );
        assert_eq!(
            GameStatus::DrawFiftyMoveRule.to_string(),
            "draw (fifty-move rule)"
        );
        assert_eq!(
            GameStatus::DrawSeventyFiveMoveRule.to_string(),
            "draw (seventy-five-move rule)"
        );
        assert_eq!(
            GameStatus::DrawThreefoldRepetition.to_string(),
            "draw (threefold repetition)"
        );
        assert_eq!(
            GameStatus::DrawFivefoldRepetition.to_string(),
            "draw (fivefold repetition)"
        );
    }

    #[test]
    fn status_ongoing() {
        assert_eq!(Board::new().status(), GameStatus::Ongoing);
        let mut board = Board::new();
        board.make_move(Move::parse("e2e4").unwrap()).unwrap();
        assert_eq!(board.status(), GameStatus::Ongoing);
    }

    #[test]
    fn checkmate_positions() {
        // Fool's Mate
        let mut board = Board::new();
        for m in ["f2f3", "e7e5", "g2g4", "d8h4"] {
            board.make_move(Move::parse(m).unwrap()).unwrap();
        }
        assert_eq!(board.status(), GameStatus::Checkmate);
        assert_eq!(board.active_color, Color::White);

        // Scholar's Mate
        let mut board = Board::new();
        for m in ["e2e4", "e7e5", "f1c4", "b8c6", "d1h5", "g8f6", "h5f7"] {
            board.make_move(Move::parse(m).unwrap()).unwrap();
        }
        assert_eq!(board.status(), GameStatus::Checkmate);
        assert_eq!(board.active_color, Color::Black);

        // Back rank mate
        let board = setup(
            &[
                ("e1", Color::White, PieceType::King),
                ("e8", Color::White, PieceType::Rook),
                ("g8", Color::Black, PieceType::King),
                ("f7", Color::Black, PieceType::Pawn),
                ("g7", Color::Black, PieceType::Pawn),
                ("h7", Color::Black, PieceType::Pawn),
            ],
            Color::Black,
        );
        assert_eq!(board.status(), GameStatus::Checkmate);

        // Queen+King mate
        let board = setup(
            &[
                ("f6", Color::White, PieceType::King),
                ("g7", Color::White, PieceType::Queen),
                ("h8", Color::Black, PieceType::King),
            ],
            Color::Black,
        );
        assert_eq!(board.status(), GameStatus::Checkmate);
    }

    #[test]
    fn check_but_not_checkmate() {
        let board = setup(
            &[
                ("e1", Color::White, PieceType::King),
                ("e8", Color::White, PieceType::Rook),
                ("d8", Color::Black, PieceType::King),
            ],
            Color::Black,
        );
        assert!(board.in_check());
        assert_eq!(board.status(), GameStatus::Ongoing);
    }

    #[test]
    fn stalemate_positions() {
        let board = setup(
            &[
                ("f6", Color::White, PieceType::King),
                ("g6", Color::White, PieceType::Queen),
                ("h8", Color::Black, PieceType::King),
            ],
            Color::Black,
        );
        assert!(!board.in_check());
        assert_eq!(board.status(), GameStatus::Stalemate);

        let board = setup(
            &[
                ("c6", Color::White, PieceType::King),
                ("b6", Color::White, PieceType::Queen),
                ("a8", Color::Black, PieceType::King),
            ],
            Color::Black,
        );
        assert_eq!(board.status(), GameStatus::Stalemate);
    }

    #[test]
    fn is_game_over_and_winner() {
        assert!(!Board::new().is_game_over());
        assert!(Board::new().winner().is_none());

        // White wins (Black checkmated)
        let board = setup(
            &[
                ("f6", Color::White, PieceType::King),
                ("g7", Color::White, PieceType::Queen),
                ("h8", Color::Black, PieceType::King),
            ],
            Color::Black,
        );
        assert!(board.is_game_over());
        assert_eq!(board.winner(), Some(Color::White));

        // Black wins (White checkmated) - back rank
        let board = setup(
            &[
                ("g1", Color::White, PieceType::King),
                ("f2", Color::White, PieceType::Pawn),
                ("g2", Color::White, PieceType::Pawn),
                ("h2", Color::White, PieceType::Pawn),
                ("e1", Color::Black, PieceType::Rook),
                ("e8", Color::Black, PieceType::King),
            ],
            Color::White,
        );
        assert_eq!(board.status(), GameStatus::Checkmate);
        assert_eq!(board.winner(), Some(Color::Black));

        // Stalemate: no winner
        let board = setup(
            &[
                ("f6", Color::White, PieceType::King),
                ("g6", Color::White, PieceType::Queen),
                ("h8", Color::Black, PieceType::King),
            ],
            Color::Black,
        );
        assert!(board.winner().is_none());
    }

    #[test]
    fn threefold_and_fivefold_repetition() {
        let moves = ["g1f3", "g8f6", "f3g1", "f6g8"];
        // threefold
        let mut board = Board::new();
        for _ in 0..2 {
            for m in moves {
                board.make_move(Move::parse(m).unwrap()).unwrap();
            }
        }
        assert_eq!(board.status(), GameStatus::DrawThreefoldRepetition);
        assert!(board.can_claim_draw());
        assert!(!board.is_game_over());

        // two repetitions only -> ongoing
        let mut board = Board::new();
        for m in moves {
            board.make_move(Move::parse(m).unwrap()).unwrap();
        }
        assert_eq!(board.status(), GameStatus::Ongoing);

        // fivefold
        let mut board = Board::new();
        for _ in 0..4 {
            for m in moves {
                board.make_move(Move::parse(m).unwrap()).unwrap();
            }
        }
        assert_eq!(board.status(), GameStatus::DrawFivefoldRepetition);
        assert!(board.is_game_over());
    }

    #[test]
    fn repetition_count_via_status() {
        // four repetitions is threefold, not fivefold, still claimable, not over
        let moves = ["g1f3", "g8f6", "f3g1", "f6g8"];
        let mut board = Board::new();
        for _ in 0..3 {
            for m in moves {
                board.make_move(Move::parse(m).unwrap()).unwrap();
            }
        }
        assert_eq!(board.status(), GameStatus::DrawThreefoldRepetition);
        assert!(!board.is_game_over());
        assert!(board.can_claim_draw());
    }

    #[test]
    fn half_move_clock_behaviour() {
        let mut board = Board::new();
        let seq = ["g1f3", "g8f6", "f3g1", "f6g8"];
        for (i, m) in seq.iter().enumerate() {
            board.make_move(Move::parse(m).unwrap()).unwrap();
            assert_eq!(board.half_move_clock, (i + 1) as u8);
        }
        // pawn move resets
        board.make_move(Move::parse("e2e4").unwrap()).unwrap();
        assert_eq!(board.half_move_clock, 0);
    }

    #[test]
    fn fifty_and_seventy_five_move_rules() {
        let base = || {
            setup(
                &[
                    ("e1", Color::White, PieceType::King),
                    ("d1", Color::White, PieceType::Queen),
                    ("e8", Color::Black, PieceType::King),
                    ("d8", Color::Black, PieceType::Queen),
                ],
                Color::White,
            )
        };

        let mut board = base();
        board.half_move_clock = 99;
        board.make_move(Move::parse("d1d2").unwrap()).unwrap();
        assert_eq!(board.half_move_clock, 100);
        assert_eq!(board.status(), GameStatus::DrawFiftyMoveRule);
        assert!(!board.is_game_over());
        assert!(board.can_claim_draw());

        let mut board = base();
        board.half_move_clock = 99;
        assert_eq!(board.status(), GameStatus::Ongoing);

        let mut board = base();
        board.half_move_clock = 149;
        board.make_move(Move::parse("d1d2").unwrap()).unwrap();
        assert_eq!(board.half_move_clock, 150);
        assert_eq!(board.status(), GameStatus::DrawSeventyFiveMoveRule);
        assert!(board.is_game_over());

        // 149 -> fifty move rule claimable
        let mut board = base();
        board.half_move_clock = 149;
        assert_eq!(board.status(), GameStatus::DrawFiftyMoveRule);

        // priority: 150 is seventy-five
        let mut board = base();
        board.half_move_clock = 150;
        assert_eq!(board.status(), GameStatus::DrawSeventyFiveMoveRule);
    }

    #[test]
    fn promotion_resets_half_move_clock() {
        let mut board = setup(
            &[
                ("e1", Color::White, PieceType::King),
                ("e7", Color::White, PieceType::Pawn),
                ("h8", Color::Black, PieceType::King),
            ],
            Color::White,
        );
        board.half_move_clock = 10;
        board.make_move(Move::parse("e7e8q").unwrap()).unwrap();
        assert_eq!(board.half_move_clock, 0);
    }

    #[test]
    fn insufficient_material() {
        // K vs K
        let board = setup(
            &[
                ("e1", Color::White, PieceType::King),
                ("e8", Color::Black, PieceType::King),
            ],
            Color::White,
        );
        assert_eq!(board.status(), GameStatus::DrawInsufficientMaterial);
        assert!(board.is_game_over());
        assert!(!board.can_claim_draw());
        assert!(board.winner().is_none());

        // K+B vs K
        let board = setup(
            &[
                ("e1", Color::White, PieceType::King),
                ("c1", Color::White, PieceType::Bishop),
                ("e8", Color::Black, PieceType::King),
            ],
            Color::White,
        );
        assert_eq!(board.status(), GameStatus::DrawInsufficientMaterial);

        // K+N vs K
        let board = setup(
            &[
                ("e1", Color::White, PieceType::King),
                ("g1", Color::White, PieceType::Knight),
                ("e8", Color::Black, PieceType::King),
            ],
            Color::White,
        );
        assert_eq!(board.status(), GameStatus::DrawInsufficientMaterial);

        // same color bishops (h1 & a8 both light)
        let board = setup(
            &[
                ("e1", Color::White, PieceType::King),
                ("h1", Color::White, PieceType::Bishop),
                ("e8", Color::Black, PieceType::King),
                ("a8", Color::Black, PieceType::Bishop),
            ],
            Color::White,
        );
        assert_eq!(board.status(), GameStatus::DrawInsufficientMaterial);

        // opposite color bishops (a1 dark, a8 light) -> not insufficient
        let board = setup(
            &[
                ("e1", Color::White, PieceType::King),
                ("a1", Color::White, PieceType::Bishop),
                ("e8", Color::Black, PieceType::King),
                ("a8", Color::Black, PieceType::Bishop),
            ],
            Color::White,
        );
        assert_eq!(board.status(), GameStatus::Ongoing);
    }

    #[test]
    fn sufficient_material() {
        let sufficient: &[&[(&str, Color, PieceType)]] = &[
            &[
                ("e1", Color::White, PieceType::King),
                ("d1", Color::White, PieceType::Queen),
                ("e8", Color::Black, PieceType::King),
            ],
            &[
                ("e1", Color::White, PieceType::King),
                ("a1", Color::White, PieceType::Rook),
                ("e8", Color::Black, PieceType::King),
            ],
            &[
                ("e1", Color::White, PieceType::King),
                ("e2", Color::White, PieceType::Pawn),
                ("e8", Color::Black, PieceType::King),
            ],
            &[
                ("e1", Color::White, PieceType::King),
                ("c1", Color::White, PieceType::Bishop),
                ("f1", Color::White, PieceType::Bishop),
                ("e8", Color::Black, PieceType::King),
            ],
            &[
                ("e1", Color::White, PieceType::King),
                ("b1", Color::White, PieceType::Knight),
                ("g1", Color::White, PieceType::Knight),
                ("e8", Color::Black, PieceType::King),
            ],
            &[
                ("e1", Color::White, PieceType::King),
                ("g1", Color::White, PieceType::Knight),
                ("e8", Color::Black, PieceType::King),
                ("g8", Color::Black, PieceType::Knight),
            ],
        ];
        for pieces in sufficient {
            let board = setup(pieces, Color::White);
            assert_ne!(board.status(), GameStatus::DrawInsufficientMaterial);
        }
    }

    #[test]
    fn checkmate_and_stalemate_not_claimable() {
        // Fool's mate
        let mut board = Board::new();
        for m in ["f2f3", "e7e5", "g2g4", "d8h4"] {
            board.make_move(Move::parse(m).unwrap()).unwrap();
        }
        assert!(board.is_game_over());
        assert!(!board.can_claim_draw());
    }
}
