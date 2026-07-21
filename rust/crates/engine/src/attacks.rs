//! Square-attack detection, working backwards from a target square to attackers.

use crate::board::Board;
use crate::types::{Color, PieceType, Square};

/// Knight move offsets: (file delta, rank delta).
const KNIGHT_OFFSETS: [(i32, i32); 8] = [
    (2, 1),
    (2, -1),
    (-2, 1),
    (-2, -1),
    (1, 2),
    (1, -2),
    (-1, 2),
    (-1, -2),
];

/// King move offsets: all 8 adjacent squares.
const KING_OFFSETS: [(i32, i32); 8] = [
    (1, 1),
    (1, -1),
    (-1, 1),
    (-1, -1),
    (1, 0),
    (-1, 0),
    (0, 1),
    (0, -1),
];

/// Diagonal directions (bishop/queen).
const DIAGONAL_DIRS: [(i32, i32); 4] = [(1, 1), (1, -1), (-1, 1), (-1, -1)];

/// Orthogonal directions (rook/queen).
const ORTHOGONAL_DIRS: [(i32, i32); 4] = [(1, 0), (-1, 0), (0, 1), (0, -1)];

impl Board {
    /// Returns true if `sq` is attacked by any piece of `by_color`.
    pub fn is_square_attacked(&self, sq: Square, by_color: Color) -> bool {
        if !sq.is_valid() {
            return false;
        }

        let file = sq.file();
        let rank = sq.rank();

        self.is_square_attacked_by_pawn(file, rank, by_color)
            || self.is_square_attacked_by_knight(file, rank, by_color)
            || self.is_square_attacked_by_king(file, rank, by_color)
            || self.is_square_attacked_diagonally(file, rank, by_color)
            || self.is_square_attacked_orthogonally(file, rank, by_color)
    }

    fn is_square_attacked_by_pawn(&self, file: i32, rank: i32, by_color: Color) -> bool {
        // A pawn attacking this square sits one rank behind (from its perspective).
        let attacker_rank = if by_color == Color::White {
            rank - 1
        } else {
            rank + 1
        };

        if !(0..=7).contains(&attacker_rank) {
            return false;
        }

        for attacker_file in [file - 1, file + 1] {
            if !(0..=7).contains(&attacker_file) {
                continue;
            }
            let attacker_sq = Square::new(attacker_file, attacker_rank);
            let piece = self.squares[attacker_sq.index()];
            if piece.piece_type() == PieceType::Pawn && piece.color() == by_color {
                return true;
            }
        }
        false
    }

    fn is_square_attacked_by_knight(&self, file: i32, rank: i32, by_color: Color) -> bool {
        for (df, dr) in KNIGHT_OFFSETS {
            let af = file + df;
            let ar = rank + dr;
            if !(0..=7).contains(&af) || !(0..=7).contains(&ar) {
                continue;
            }
            let attacker_sq = Square::new(af, ar);
            let piece = self.squares[attacker_sq.index()];
            if piece.piece_type() == PieceType::Knight && piece.color() == by_color {
                return true;
            }
        }
        false
    }

    fn is_square_attacked_by_king(&self, file: i32, rank: i32, by_color: Color) -> bool {
        for (df, dr) in KING_OFFSETS {
            let af = file + df;
            let ar = rank + dr;
            if !(0..=7).contains(&af) || !(0..=7).contains(&ar) {
                continue;
            }
            let attacker_sq = Square::new(af, ar);
            let piece = self.squares[attacker_sq.index()];
            if piece.piece_type() == PieceType::King && piece.color() == by_color {
                return true;
            }
        }
        false
    }

    fn is_square_attacked_diagonally(&self, file: i32, rank: i32, by_color: Color) -> bool {
        for (df, dr) in DIAGONAL_DIRS {
            for dist in 1..=7 {
                let af = file + df * dist;
                let ar = rank + dr * dist;
                if !(0..=7).contains(&af) || !(0..=7).contains(&ar) {
                    break;
                }
                let attacker_sq = Square::new(af, ar);
                let piece = self.squares[attacker_sq.index()];
                if piece.is_empty() {
                    continue;
                }
                if piece.color() == by_color
                    && (piece.piece_type() == PieceType::Bishop
                        || piece.piece_type() == PieceType::Queen)
                {
                    return true;
                }
                break;
            }
        }
        false
    }

    fn is_square_attacked_orthogonally(&self, file: i32, rank: i32, by_color: Color) -> bool {
        for (df, dr) in ORTHOGONAL_DIRS {
            for dist in 1..=7 {
                let af = file + df * dist;
                let ar = rank + dr * dist;
                if !(0..=7).contains(&af) || !(0..=7).contains(&ar) {
                    break;
                }
                let attacker_sq = Square::new(af, ar);
                let piece = self.squares[attacker_sq.index()];
                if piece.is_empty() {
                    continue;
                }
                if piece.color() == by_color
                    && (piece.piece_type() == PieceType::Rook
                        || piece.piece_type() == PieceType::Queen)
                {
                    return true;
                }
                break;
            }
        }
        false
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::types::Piece;

    fn empty() -> Board {
        Board::default()
    }

    fn place(b: &mut Board, file: i32, rank: i32, color: Color, pt: PieceType) {
        b.squares[Square::new(file, rank).index()] = Piece::new(color, pt);
    }

    #[test]
    fn empty_board_no_attacks() {
        let board = empty();
        let e4 = Square::new(4, 3);
        assert!(!board.is_square_attacked(e4, Color::White));
        assert!(!board.is_square_attacked(e4, Color::Black));
    }

    #[test]
    fn invalid_square_returns_false() {
        let board = Board::new();
        assert!(!board.is_square_attacked(crate::types::NO_SQUARE, Color::White));
        assert!(!board.is_square_attacked(Square(-5), Color::Black));
        assert!(!board.is_square_attacked(Square(64), Color::White));
    }

    #[test]
    fn pawn_attacks() {
        let mut board = empty();
        place(&mut board, 4, 3, Color::White, PieceType::Pawn); // e4
        assert!(board.is_square_attacked(Square::new(3, 4), Color::White)); // d5
        assert!(board.is_square_attacked(Square::new(5, 4), Color::White)); // f5
        assert!(!board.is_square_attacked(Square::new(4, 4), Color::White)); // e5 forward
        assert!(!board.is_square_attacked(Square::new(3, 2), Color::White)); // d3
        assert!(!board.is_square_attacked(Square::new(5, 2), Color::White)); // f3

        let mut board = empty();
        place(&mut board, 4, 4, Color::Black, PieceType::Pawn); // e5
        assert!(board.is_square_attacked(Square::new(3, 3), Color::Black)); // d4
        assert!(board.is_square_attacked(Square::new(5, 3), Color::Black)); // f4
        assert!(!board.is_square_attacked(Square::new(4, 3), Color::Black)); // e4 forward

        // a-file pawn only attacks right
        let mut board = empty();
        place(&mut board, 0, 1, Color::White, PieceType::Pawn); // a2
        assert!(board.is_square_attacked(Square::new(1, 2), Color::White)); // b3

        // wrong color
        let mut board = empty();
        place(&mut board, 4, 3, Color::White, PieceType::Pawn);
        assert!(board.is_square_attacked(Square::new(3, 4), Color::White));
        assert!(!board.is_square_attacked(Square::new(3, 4), Color::Black));
    }

    #[test]
    fn knight_attacks() {
        let mut board = empty();
        place(&mut board, 4, 3, Color::White, PieceType::Knight); // e4
        for (f, r) in [
            (5, 5),
            (6, 4),
            (6, 2),
            (5, 1),
            (3, 1),
            (2, 2),
            (2, 4),
            (3, 5),
        ] {
            assert!(board.is_square_attacked(Square::new(f, r), Color::White));
        }
        assert!(!board.is_square_attacked(Square::new(4, 2), Color::White));
        assert!(!board.is_square_attacked(Square::new(4, 4), Color::White));

        let mut board = empty();
        place(&mut board, 0, 0, Color::White, PieceType::Knight); // a1
        assert!(board.is_square_attacked(Square::new(1, 2), Color::White)); // b3
        assert!(board.is_square_attacked(Square::new(2, 1), Color::White)); // c2
        assert!(!board.is_square_attacked(Square::new(0, 2), Color::White)); // a3
    }

    #[test]
    fn bishop_attacks() {
        let mut board = empty();
        place(&mut board, 4, 3, Color::White, PieceType::Bishop); // e4
        for (f, r) in [
            (5, 4),
            (7, 6),
            (5, 2),
            (7, 0),
            (3, 2),
            (1, 0),
            (3, 4),
            (0, 7),
        ] {
            assert!(board.is_square_attacked(Square::new(f, r), Color::White));
        }
        assert!(!board.is_square_attacked(Square::new(4, 4), Color::White));
        assert!(!board.is_square_attacked(Square::new(3, 3), Color::White));

        // blocked by own piece
        let mut board = empty();
        place(&mut board, 4, 3, Color::White, PieceType::Bishop);
        place(&mut board, 5, 4, Color::White, PieceType::Knight);
        assert!(board.is_square_attacked(Square::new(5, 4), Color::White));
        assert!(!board.is_square_attacked(Square::new(6, 5), Color::White));
        assert!(!board.is_square_attacked(Square::new(7, 6), Color::White));
        assert!(board.is_square_attacked(Square::new(3, 4), Color::White));

        // corner bishop
        let mut board = empty();
        place(&mut board, 0, 0, Color::White, PieceType::Bishop);
        assert!(board.is_square_attacked(Square::new(7, 7), Color::White));
        assert!(board.is_square_attacked(Square::new(3, 3), Color::White));
        assert!(!board.is_square_attacked(Square::new(1, 0), Color::White));
    }

    #[test]
    fn rook_attacks() {
        let mut board = empty();
        place(&mut board, 4, 3, Color::White, PieceType::Rook); // e4
        for (f, r) in [
            (4, 4),
            (4, 7),
            (4, 2),
            (4, 0),
            (5, 3),
            (7, 3),
            (3, 3),
            (0, 3),
        ] {
            assert!(board.is_square_attacked(Square::new(f, r), Color::White));
        }
        assert!(!board.is_square_attacked(Square::new(5, 4), Color::White));
        assert!(!board.is_square_attacked(Square::new(3, 2), Color::White));

        // blocked
        let mut board = empty();
        place(&mut board, 4, 3, Color::White, PieceType::Rook);
        place(&mut board, 4, 5, Color::White, PieceType::Pawn); // e6
        assert!(board.is_square_attacked(Square::new(4, 4), Color::White)); // e5
        assert!(board.is_square_attacked(Square::new(4, 5), Color::White)); // e6
        assert!(!board.is_square_attacked(Square::new(4, 6), Color::White)); // e7
        assert!(!board.is_square_attacked(Square::new(4, 7), Color::White)); // e8

        // blocked by enemy, still attacks the enemy square
        let mut board = empty();
        place(&mut board, 0, 0, Color::White, PieceType::Rook); // a1
        place(&mut board, 0, 3, Color::Black, PieceType::Pawn); // a4
        assert!(board.is_square_attacked(Square::new(0, 3), Color::White));
        assert!(!board.is_square_attacked(Square::new(0, 7), Color::White));
    }

    #[test]
    fn queen_attacks() {
        let mut board = empty();
        place(&mut board, 4, 3, Color::White, PieceType::Queen); // e4
        assert!(board.is_square_attacked(Square::new(7, 6), Color::White)); // h7
        assert!(board.is_square_attacked(Square::new(0, 7), Color::White)); // a8
        assert!(board.is_square_attacked(Square::new(4, 7), Color::White)); // e8
        assert!(board.is_square_attacked(Square::new(0, 3), Color::White)); // a4
    }

    #[test]
    fn king_attacks() {
        let mut board = empty();
        place(&mut board, 4, 3, Color::White, PieceType::King); // e4
        for (f, r) in [
            (3, 2),
            (4, 2),
            (5, 2),
            (3, 3),
            (5, 3),
            (3, 4),
            (4, 4),
            (5, 4),
        ] {
            assert!(board.is_square_attacked(Square::new(f, r), Color::White));
        }
        assert!(!board.is_square_attacked(Square::new(4, 5), Color::White));
        assert!(!board.is_square_attacked(Square::new(6, 5), Color::White));

        let mut board = empty();
        place(&mut board, 0, 0, Color::White, PieceType::King);
        assert!(board.is_square_attacked(Square::new(0, 1), Color::White));
        assert!(board.is_square_attacked(Square::new(1, 0), Color::White));
        assert!(board.is_square_attacked(Square::new(1, 1), Color::White));
        assert!(!board.is_square_attacked(Square::new(0, 2), Color::White));
    }

    #[test]
    fn starting_position_control() {
        let board = Board::new();
        assert!(board.is_square_attacked(Square::new(4, 2), Color::White)); // e3
        assert!(board.is_square_attacked(Square::new(4, 5), Color::Black)); // e6
        assert!(board.is_square_attacked(Square::new(0, 2), Color::White)); // a3 knight
        assert!(board.is_square_attacked(Square::new(2, 2), Color::White)); // c3 knight
        assert!(!board.is_square_attacked(Square::new(4, 3), Color::White)); // e4
        assert!(!board.is_square_attacked(Square::new(4, 3), Color::Black));
    }
}
