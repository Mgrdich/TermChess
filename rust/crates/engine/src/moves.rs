//! The `Move` type, move parsing/formatting, and legal/pseudo-legal move generation.

use std::fmt;

use thiserror::Error;

use crate::board::{
    Board, CASTLE_BLACK_KING, CASTLE_BLACK_QUEEN, CASTLE_WHITE_KING, CASTLE_WHITE_QUEEN,
};
use crate::types::{Color, PieceType, Square, NO_SQUARE};

/// A chess move from one square to another, with an optional promotion.
#[derive(Clone, Copy, PartialEq, Eq, Debug, Hash)]
pub struct Move {
    /// Source square.
    pub from: Square,
    /// Destination square.
    pub to: Square,
    /// Promotion piece type ([`PieceType::Empty`] if not a promotion).
    pub promotion: PieceType,
}

/// Error returned by [`Move::parse`] for malformed coordinate notation.
#[derive(Debug, Error, PartialEq, Eq)]
pub enum ParseMoveError {
    /// The input was not 4-5 characters long.
    #[error("invalid move format: expected 4-5 characters")]
    Format,
    /// The from square was out of range.
    #[error("invalid from square: {0}")]
    FromSquare(String),
    /// The to square was out of range.
    #[error("invalid to square: {0}")]
    ToSquare(String),
    /// The promotion character was not one of q/r/b/n.
    #[error("invalid promotion character: {0}")]
    Promotion(char),
}

impl Move {
    /// Creates a non-promotion move.
    pub fn new(from: Square, to: Square) -> Move {
        Move {
            from,
            to,
            promotion: PieceType::Empty,
        }
    }

    /// Creates a promotion move.
    pub fn with_promotion(from: Square, to: Square, promotion: PieceType) -> Move {
        Move {
            from,
            to,
            promotion,
        }
    }

    /// Parses a move from coordinate notation (e.g. `"e2e4"`, `"a7a8q"`).
    pub fn parse(s: &str) -> Result<Move, ParseMoveError> {
        let bytes = s.as_bytes();
        if bytes.len() < 4 || bytes.len() > 5 {
            return Err(ParseMoveError::Format);
        }

        let from_file = bytes[0] as i32 - b'a' as i32;
        let from_rank = bytes[1] as i32 - b'1' as i32;
        if !(0..=7).contains(&from_file) || !(0..=7).contains(&from_rank) {
            return Err(ParseMoveError::FromSquare(s[0..2].to_string()));
        }

        let to_file = bytes[2] as i32 - b'a' as i32;
        let to_rank = bytes[3] as i32 - b'1' as i32;
        if !(0..=7).contains(&to_file) || !(0..=7).contains(&to_rank) {
            return Err(ParseMoveError::ToSquare(s[2..4].to_string()));
        }

        let from = Square::new(from_file, from_rank);
        let to = Square::new(to_file, to_rank);

        let mut promotion = PieceType::Empty;
        if bytes.len() == 5 {
            promotion = match bytes[4] {
                b'q' => PieceType::Queen,
                b'r' => PieceType::Rook,
                b'b' => PieceType::Bishop,
                b'n' => PieceType::Knight,
                other => return Err(ParseMoveError::Promotion(other as char)),
            };
        }

        Ok(Move {
            from,
            to,
            promotion,
        })
    }
}

impl fmt::Display for Move {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}{}", self.from, self.to)?;
        match self.promotion {
            PieceType::Queen => write!(f, "q"),
            PieceType::Rook => write!(f, "r"),
            PieceType::Bishop => write!(f, "b"),
            PieceType::Knight => write!(f, "n"),
            _ => Ok(()),
        }
    }
}

impl Board {
    /// Generates all pseudo-legal pawn moves for the active color.
    fn generate_pawn_moves(&self) -> Vec<Move> {
        let mut moves = Vec::new();

        let (direction, start_rank, promotion_rank) = if self.active_color == Color::White {
            (1, 1, 7)
        } else {
            (-1, 6, 0)
        };

        for sq_idx in 0..64i8 {
            let sq = Square(sq_idx);
            let piece = self.squares[sq_idx as usize];
            if piece.is_empty()
                || piece.piece_type() != PieceType::Pawn
                || piece.color() != self.active_color
            {
                continue;
            }

            let file = sq.file();
            let rank = sq.rank();

            // One square forward.
            let forward_rank = rank + direction;
            if (0..=7).contains(&forward_rank) {
                let forward_sq = Square::new(file, forward_rank);
                if self.squares[forward_sq.index()].is_empty() {
                    if forward_rank == promotion_rank {
                        for promo in [
                            PieceType::Queen,
                            PieceType::Rook,
                            PieceType::Bishop,
                            PieceType::Knight,
                        ] {
                            moves.push(Move::with_promotion(sq, forward_sq, promo));
                        }
                    } else {
                        moves.push(Move::new(sq, forward_sq));

                        // Two squares forward from the starting rank.
                        if rank == start_rank {
                            let two_forward_rank = rank + 2 * direction;
                            let two_forward_sq = Square::new(file, two_forward_rank);
                            if self.squares[two_forward_sq.index()].is_empty() {
                                moves.push(Move::new(sq, two_forward_sq));
                            }
                        }
                    }
                }
            }

            // Diagonal captures.
            for file_offset in [-1, 1] {
                let capture_file = file + file_offset;
                let capture_rank = rank + direction;
                if (0..=7).contains(&capture_file) && (0..=7).contains(&capture_rank) {
                    let capture_sq = Square::new(capture_file, capture_rank);
                    let target = self.squares[capture_sq.index()];
                    if !target.is_empty() && target.color() != self.active_color {
                        if capture_rank == promotion_rank {
                            for promo in [
                                PieceType::Queen,
                                PieceType::Rook,
                                PieceType::Bishop,
                                PieceType::Knight,
                            ] {
                                moves.push(Move::with_promotion(sq, capture_sq, promo));
                            }
                        } else {
                            moves.push(Move::new(sq, capture_sq));
                        }
                    }
                }
            }

            // En passant capture.
            if self.en_passant_sq >= 0 {
                let ep_square = Square(self.en_passant_sq);
                let ep_file = ep_square.file();
                let ep_rank = ep_square.rank();

                let file_diff = (file - ep_file).abs();
                if file_diff == 1
                    && ((self.active_color == Color::White && rank == 4 && ep_rank == 5)
                        || (self.active_color == Color::Black && rank == 3 && ep_rank == 2))
                {
                    moves.push(Move::new(sq, ep_square));
                }
            }
        }

        moves
    }

    /// Generates all pseudo-legal knight moves for the active color.
    fn generate_knight_moves(&self) -> Vec<Move> {
        let offsets = [
            (2, 1),
            (2, -1),
            (-2, 1),
            (-2, -1),
            (1, 2),
            (1, -2),
            (-1, 2),
            (-1, -2),
        ];
        let mut moves = Vec::new();

        for sq_idx in 0..64i8 {
            let sq = Square(sq_idx);
            let piece = self.squares[sq_idx as usize];
            if piece.is_empty()
                || piece.piece_type() != PieceType::Knight
                || piece.color() != self.active_color
            {
                continue;
            }

            let file = sq.file();
            let rank = sq.rank();
            for (df, dr) in offsets {
                let nf = file + df;
                let nr = rank + dr;
                if !(0..=7).contains(&nf) || !(0..=7).contains(&nr) {
                    continue;
                }
                let target_sq = Square::new(nf, nr);
                let target = self.squares[target_sq.index()];
                if target.is_empty() || target.color() != self.active_color {
                    moves.push(Move::new(sq, target_sq));
                }
            }
        }

        moves
    }

    /// Generates pseudo-legal sliding moves for a piece type along the given directions.
    fn generate_sliding_moves(
        &self,
        piece_type: PieceType,
        directions: &[(i32, i32)],
    ) -> Vec<Move> {
        let mut moves = Vec::new();

        for sq_idx in 0..64i8 {
            let sq = Square(sq_idx);
            let piece = self.squares[sq_idx as usize];
            if piece.is_empty()
                || piece.piece_type() != piece_type
                || piece.color() != self.active_color
            {
                continue;
            }

            let file = sq.file();
            let rank = sq.rank();
            for &(df, dr) in directions {
                for dist in 1..=7 {
                    let nf = file + df * dist;
                    let nr = rank + dr * dist;
                    if !(0..=7).contains(&nf) || !(0..=7).contains(&nr) {
                        break;
                    }
                    let target_sq = Square::new(nf, nr);
                    let target = self.squares[target_sq.index()];
                    if target.is_empty() {
                        moves.push(Move::new(sq, target_sq));
                    } else if target.color() != self.active_color {
                        moves.push(Move::new(sq, target_sq));
                        break;
                    } else {
                        break;
                    }
                }
            }
        }

        moves
    }

    fn generate_bishop_moves(&self) -> Vec<Move> {
        self.generate_sliding_moves(PieceType::Bishop, &[(1, 1), (1, -1), (-1, 1), (-1, -1)])
    }

    fn generate_rook_moves(&self) -> Vec<Move> {
        self.generate_sliding_moves(PieceType::Rook, &[(1, 0), (-1, 0), (0, 1), (0, -1)])
    }

    fn generate_queen_moves(&self) -> Vec<Move> {
        self.generate_sliding_moves(
            PieceType::Queen,
            &[
                (1, 1),
                (1, -1),
                (-1, 1),
                (-1, -1),
                (1, 0),
                (-1, 0),
                (0, 1),
                (0, -1),
            ],
        )
    }

    /// Generates all pseudo-legal king moves (including castling) for the active color.
    fn generate_king_moves(&self) -> Vec<Move> {
        let offsets = [
            (1, 1),
            (1, -1),
            (-1, 1),
            (-1, -1),
            (1, 0),
            (-1, 0),
            (0, 1),
            (0, -1),
        ];
        let mut moves = Vec::new();

        for sq_idx in 0..64i8 {
            let sq = Square(sq_idx);
            let piece = self.squares[sq_idx as usize];
            if piece.is_empty()
                || piece.piece_type() != PieceType::King
                || piece.color() != self.active_color
            {
                continue;
            }

            let file = sq.file();
            let rank = sq.rank();
            for (df, dr) in offsets {
                let nf = file + df;
                let nr = rank + dr;
                if !(0..=7).contains(&nf) || !(0..=7).contains(&nr) {
                    continue;
                }
                let target_sq = Square::new(nf, nr);
                let target = self.squares[target_sq.index()];
                if target.is_empty() || target.color() != self.active_color {
                    moves.push(Move::new(sq, target_sq));
                }
            }

            moves.extend(self.generate_castling_moves(sq));
        }

        moves
    }

    /// Generates castling moves for the king on `king_sq`, honoring rights, empty
    /// squares, and check on the king's path.
    fn generate_castling_moves(&self, king_sq: Square) -> Vec<Move> {
        let mut moves = Vec::new();
        let opponent = self.active_color.opponent();

        // King must not currently be in check.
        if self.is_square_attacked(king_sq, opponent) {
            return moves;
        }

        if self.active_color == Color::White {
            if king_sq != Square::new(4, 0) {
                return moves;
            }

            if self.castling_rights & CASTLE_WHITE_KING != 0 {
                let f1 = Square::new(5, 0);
                let g1 = Square::new(6, 0);
                if self.squares[f1.index()].is_empty()
                    && self.squares[g1.index()].is_empty()
                    && !self.is_square_attacked(f1, opponent)
                    && !self.is_square_attacked(g1, opponent)
                {
                    moves.push(Move::new(king_sq, g1));
                }
            }

            if self.castling_rights & CASTLE_WHITE_QUEEN != 0 {
                let b1 = Square::new(1, 0);
                let c1 = Square::new(2, 0);
                let d1 = Square::new(3, 0);
                if self.squares[b1.index()].is_empty()
                    && self.squares[c1.index()].is_empty()
                    && self.squares[d1.index()].is_empty()
                    && !self.is_square_attacked(c1, opponent)
                    && !self.is_square_attacked(d1, opponent)
                {
                    moves.push(Move::new(king_sq, c1));
                }
            }
        } else {
            if king_sq != Square::new(4, 7) {
                return moves;
            }

            if self.castling_rights & CASTLE_BLACK_KING != 0 {
                let f8 = Square::new(5, 7);
                let g8 = Square::new(6, 7);
                if self.squares[f8.index()].is_empty()
                    && self.squares[g8.index()].is_empty()
                    && !self.is_square_attacked(f8, opponent)
                    && !self.is_square_attacked(g8, opponent)
                {
                    moves.push(Move::new(king_sq, g8));
                }
            }

            if self.castling_rights & CASTLE_BLACK_QUEEN != 0 {
                let b8 = Square::new(1, 7);
                let c8 = Square::new(2, 7);
                let d8 = Square::new(3, 7);
                if self.squares[b8.index()].is_empty()
                    && self.squares[c8.index()].is_empty()
                    && self.squares[d8.index()].is_empty()
                    && !self.is_square_attacked(c8, opponent)
                    && !self.is_square_attacked(d8, opponent)
                {
                    moves.push(Move::new(king_sq, c8));
                }
            }
        }

        moves
    }

    /// Generates all pseudo-legal moves for the active color (may leave king in check).
    pub fn pseudo_legal_moves(&self) -> Vec<Move> {
        let mut moves = Vec::new();
        moves.extend(self.generate_pawn_moves());
        moves.extend(self.generate_knight_moves());
        moves.extend(self.generate_bishop_moves());
        moves.extend(self.generate_rook_moves());
        moves.extend(self.generate_queen_moves());
        moves.extend(self.generate_king_moves());
        moves
    }

    /// Generates all legal moves for the active color (does not leave king in check).
    pub fn legal_moves(&self) -> Vec<Move> {
        let pseudo = self.pseudo_legal_moves();
        let mut legal = Vec::new();
        let moving_color = self.active_color;

        for m in pseudo {
            let mut board_copy = self.copy();
            board_copy.apply_move(m);

            // Find the king of the color that just moved.
            let mut king_square = NO_SQUARE;
            for sq_idx in 0..64i8 {
                let piece = board_copy.squares[sq_idx as usize];
                if piece.piece_type() == PieceType::King && piece.color() == moving_color {
                    king_square = Square(sq_idx);
                    break;
                }
            }

            if king_square == NO_SQUARE {
                continue;
            }

            if !board_copy.is_square_attacked(king_square, board_copy.active_color) {
                legal.push(m);
            }
        }

        legal
    }

    /// Returns true if `m` is a legal move in the current position.
    pub fn is_legal_move(&self, m: Move) -> bool {
        self.legal_moves()
            .iter()
            .any(|lm| lm.from == m.from && lm.to == m.to && lm.promotion == m.promotion)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::board::{Board, CASTLE_ALL, CASTLE_WHITE_KING, CASTLE_WHITE_QUEEN};
    use crate::types::Piece;

    fn sq(file: i32, rank: i32) -> Square {
        Square::new(file, rank)
    }

    fn place(b: &mut Board, file: i32, rank: i32, color: Color, pt: PieceType) {
        b.squares[sq(file, rank).index()] = Piece::new(color, pt);
    }

    fn contains(moves: &[Move], from: Square, to: Square) -> bool {
        moves.iter().any(|m| m.from == from && m.to == to)
    }

    fn count_from(moves: &[Move], from: Square) -> usize {
        moves.iter().filter(|m| m.from == from).count()
    }

    fn white_board() -> Board {
        Board {
            active_color: Color::White,
            ..Board::default()
        }
    }

    #[test]
    fn parse_move_cases() {
        assert_eq!(Move::parse("e2e4").unwrap(), Move::new(sq(4, 1), sq(4, 3)));
        assert_eq!(Move::parse("a1h8").unwrap(), Move::new(sq(0, 0), sq(7, 7)));
        assert_eq!(
            Move::parse("a7a8q").unwrap(),
            Move::with_promotion(sq(0, 6), sq(0, 7), PieceType::Queen)
        );
        assert_eq!(Move::parse("h7h8r").unwrap().promotion, PieceType::Rook);
        assert_eq!(Move::parse("b7b8b").unwrap().promotion, PieceType::Bishop);
        assert_eq!(Move::parse("c7c8n").unwrap().promotion, PieceType::Knight);
        for bad in ["e2", "e2e9", "xyz", "i2e4", "e7e8x", "e2e4qq"] {
            assert!(Move::parse(bad).is_err(), "{} should error", bad);
        }
    }

    #[test]
    fn move_string() {
        assert_eq!(Move::new(sq(4, 1), sq(4, 3)).to_string(), "e2e4");
        assert_eq!(
            Move::with_promotion(sq(0, 6), sq(0, 7), PieceType::Queen).to_string(),
            "a7a8q"
        );
        assert_eq!(
            Move::with_promotion(sq(7, 6), sq(7, 7), PieceType::Rook).to_string(),
            "h7h8r"
        );
        assert_eq!(
            Move::with_promotion(sq(1, 6), sq(1, 7), PieceType::Bishop).to_string(),
            "b7b8b"
        );
        assert_eq!(
            Move::with_promotion(sq(2, 6), sq(2, 7), PieceType::Knight).to_string(),
            "c7c8n"
        );
    }

    #[test]
    fn move_round_trip() {
        for s in ["e2e4", "a7a8q", "b1c3", "h1h8", "d7d8r", "e7e8n", "f7f8b"] {
            assert_eq!(Move::parse(s).unwrap().to_string(), s);
        }
    }

    #[test]
    fn pawn_move_generation() {
        let board = Board::new();
        let moves = board.generate_pawn_moves();
        assert_eq!(count_from(&moves, sq(4, 1)), 2);
        assert!(contains(&moves, sq(4, 1), sq(4, 2)));
        assert!(contains(&moves, sq(4, 1), sq(4, 3)));

        // blocked pawn
        let mut board = Board::new();
        place(&mut board, 4, 2, Color::Black, PieceType::Knight);
        let moves = board.generate_pawn_moves();
        assert_eq!(count_from(&moves, sq(4, 1)), 0);
    }

    #[test]
    fn en_passant_generation() {
        // white captures on e6 from d5
        let mut board = Board::new();
        place(&mut board, 3, 4, Color::White, PieceType::Pawn);
        place(&mut board, 4, 4, Color::Black, PieceType::Pawn);
        board.en_passant_sq = sq(4, 5).0;
        let moves = board.generate_pawn_moves();
        assert!(contains(&moves, sq(3, 4), sq(4, 5)));

        // wrong rank
        let mut board = Board::new();
        place(&mut board, 3, 3, Color::White, PieceType::Pawn);
        board.en_passant_sq = sq(4, 5).0;
        let moves = board.generate_pawn_moves();
        assert!(!contains(&moves, sq(3, 3), sq(4, 5)));

        // not adjacent
        let mut board = Board::new();
        place(&mut board, 1, 4, Color::White, PieceType::Pawn);
        board.en_passant_sq = sq(4, 5).0;
        let moves = board.generate_pawn_moves();
        assert!(!contains(&moves, sq(1, 4), sq(4, 5)));

        // both sides can capture
        let mut board = Board::new();
        for i in 8..16 {
            board.squares[i] = Piece::EMPTY;
        }
        for i in 48..56 {
            board.squares[i] = Piece::EMPTY;
        }
        place(&mut board, 3, 4, Color::White, PieceType::Pawn);
        place(&mut board, 5, 4, Color::White, PieceType::Pawn);
        place(&mut board, 4, 4, Color::Black, PieceType::Pawn);
        board.en_passant_sq = sq(4, 5).0;
        let moves = board.generate_pawn_moves();
        assert_eq!(moves.iter().filter(|m| m.to == sq(4, 5)).count(), 2);
    }

    #[test]
    fn knight_move_generation() {
        let mut board = white_board();
        place(&mut board, 4, 3, Color::White, PieceType::Knight);
        let moves = board.generate_knight_moves();
        assert_eq!(moves.len(), 8);

        let mut board = white_board();
        place(&mut board, 0, 0, Color::White, PieceType::Knight);
        assert_eq!(board.generate_knight_moves().len(), 2);
    }

    #[test]
    fn sliding_move_generation() {
        let mut board = white_board();
        place(&mut board, 4, 3, Color::White, PieceType::Bishop);
        assert_eq!(board.generate_bishop_moves().len(), 13);

        let mut board = white_board();
        place(&mut board, 4, 3, Color::White, PieceType::Rook);
        assert_eq!(board.generate_rook_moves().len(), 14);

        let mut board = white_board();
        place(&mut board, 4, 3, Color::White, PieceType::Queen);
        assert_eq!(board.generate_queen_moves().len(), 27);
    }

    #[test]
    fn king_move_generation() {
        let mut board = white_board();
        place(&mut board, 4, 3, Color::White, PieceType::King);
        assert_eq!(board.generate_king_moves().len(), 8);

        let mut board = white_board();
        place(&mut board, 0, 0, Color::White, PieceType::King);
        assert_eq!(board.generate_king_moves().len(), 3);
    }

    #[test]
    fn pseudo_legal_and_legal_counts() {
        let board = Board::new();
        assert_eq!(board.pseudo_legal_moves().len(), 20);
        assert_eq!(board.legal_moves().len(), 20);
        let mut board = Board::new();
        board.active_color = Color::Black;
        assert_eq!(board.pseudo_legal_moves().len(), 20);
    }

    #[test]
    fn pinned_piece_filtering() {
        // pinned bishop has 0 legal moves
        let mut board = white_board();
        place(&mut board, 4, 0, Color::White, PieceType::King);
        place(&mut board, 4, 1, Color::White, PieceType::Bishop);
        place(&mut board, 4, 7, Color::Black, PieceType::Rook);
        place(&mut board, 7, 7, Color::Black, PieceType::King);
        let moves = board.legal_moves();
        assert_eq!(count_from(&moves, sq(4, 1)), 0);

        // pinned rook along a-file has 6 moves
        let mut board = white_board();
        place(&mut board, 0, 0, Color::White, PieceType::King);
        place(&mut board, 0, 3, Color::White, PieceType::Rook);
        place(&mut board, 0, 7, Color::Black, PieceType::Queen);
        place(&mut board, 7, 7, Color::Black, PieceType::King);
        let moves = board.legal_moves();
        assert_eq!(count_from(&moves, sq(0, 3)), 6);
        assert!(contains(&moves, sq(0, 3), sq(0, 4)));
        assert!(!contains(&moves, sq(0, 3), sq(1, 3)));
    }

    #[test]
    fn check_response_filtering() {
        // knight must block on e2 or e4
        let mut board = white_board();
        place(&mut board, 4, 0, Color::White, PieceType::King);
        place(&mut board, 2, 2, Color::White, PieceType::Knight);
        place(&mut board, 4, 7, Color::Black, PieceType::Queen);
        place(&mut board, 7, 7, Color::Black, PieceType::King);
        assert!(board.in_check());
        let moves = board.legal_moves();
        assert!(contains(&moves, sq(2, 2), sq(4, 1)));
        assert!(contains(&moves, sq(2, 2), sq(4, 3)));
        assert_eq!(count_from(&moves, sq(2, 2)), 2);

        // double check: only king moves
        let mut board = white_board();
        place(&mut board, 4, 0, Color::White, PieceType::King);
        place(&mut board, 4, 7, Color::Black, PieceType::Queen);
        place(&mut board, 3, 2, Color::Black, PieceType::Knight);
        place(&mut board, 7, 7, Color::Black, PieceType::King);
        place(&mut board, 0, 1, Color::White, PieceType::Rook);
        assert!(board.in_check());
        let moves = board.legal_moves();
        assert_eq!(count_from(&moves, sq(0, 1)), 0);
        assert!(count_from(&moves, sq(4, 0)) > 0);
    }

    #[test]
    fn king_cannot_move_into_check() {
        let mut board = white_board();
        place(&mut board, 4, 3, Color::White, PieceType::King);
        place(&mut board, 0, 4, Color::Black, PieceType::Rook);
        place(&mut board, 7, 7, Color::Black, PieceType::King);
        let moves = board.legal_moves();
        assert!(!contains(&moves, sq(4, 3), sq(3, 4)));
        assert!(!contains(&moves, sq(4, 3), sq(4, 4)));
        assert!(!contains(&moves, sq(4, 3), sq(5, 4)));
        assert!(contains(&moves, sq(4, 3), sq(4, 2)));
    }

    #[test]
    fn is_legal_move_basics() {
        let board = Board::new();
        assert!(board.is_legal_move(Move::parse("e2e4").unwrap()));
        assert!(board.is_legal_move(Move::parse("g1f3").unwrap()));
        assert!(!board.is_legal_move(Move::parse("e4e5").unwrap()));
        assert!(!board.is_legal_move(Move::parse("e7e6").unwrap()));
        assert!(!board.is_legal_move(Move::parse("e2e5").unwrap()));
        assert!(!board.is_legal_move(Move::parse("a1b1").unwrap()));
        // same from/to square
        assert!(!board.is_legal_move(Move::new(sq(4, 1), sq(4, 1))));
    }

    #[test]
    fn make_move_basics() {
        let mut board = Board::new();
        board.make_move(Move::parse("e2e4").unwrap()).unwrap();
        assert_eq!(
            board.squares[sq(4, 3).index()].piece_type(),
            PieceType::Pawn
        );
        assert!(board.squares[sq(4, 1).index()].is_empty());
        assert_eq!(board.active_color, Color::Black);
        assert_eq!(board.full_move_num, 1);
        board.make_move(Move::parse("e7e5").unwrap()).unwrap();
        assert_eq!(board.active_color, Color::White);
        assert_eq!(board.full_move_num, 2);

        // illegal moves
        let mut board = Board::new();
        assert!(board.make_move(Move::parse("e4e5").unwrap()).is_err());
        assert!(board.make_move(Move::parse("e7e6").unwrap()).is_err());

        // board unchanged after illegal
        let mut board = Board::new();
        let before = board.clone();
        assert!(board.make_move(Move::parse("e4e5").unwrap()).is_err());
        assert_eq!(board, before);

        // error message contains move string
        let mut board = Board::new();
        let err = board.make_move(Move::parse("e4e5").unwrap()).unwrap_err();
        assert!(err.to_string().contains("e4e5"));
    }

    #[test]
    fn castling_move_generation() {
        let mut board = Board {
            active_color: Color::White,
            castling_rights: CASTLE_WHITE_KING | CASTLE_WHITE_QUEEN,
            ..Board::default()
        };
        place(&mut board, 4, 0, Color::White, PieceType::King);
        place(&mut board, 0, 0, Color::White, PieceType::Rook);
        place(&mut board, 7, 0, Color::White, PieceType::Rook);
        place(&mut board, 4, 7, Color::Black, PieceType::King);
        let moves = board.generate_king_moves();
        assert!(contains(&moves, sq(4, 0), sq(6, 0)));
        assert!(contains(&moves, sq(4, 0), sq(2, 0)));

        // no rights -> no castling
        let mut board = Board {
            active_color: Color::White,
            castling_rights: 0,
            ..Board::default()
        };
        place(&mut board, 4, 0, Color::White, PieceType::King);
        place(&mut board, 0, 0, Color::White, PieceType::Rook);
        place(&mut board, 7, 0, Color::White, PieceType::Rook);
        place(&mut board, 4, 7, Color::Black, PieceType::King);
        let moves = board.generate_king_moves();
        assert!(!contains(&moves, sq(4, 0), sq(6, 0)));
        assert!(!contains(&moves, sq(4, 0), sq(2, 0)));

        // blocked on f1
        let mut board = Board {
            active_color: Color::White,
            castling_rights: CASTLE_WHITE_KING,
            ..Board::default()
        };
        place(&mut board, 4, 0, Color::White, PieceType::King);
        place(&mut board, 5, 0, Color::White, PieceType::Bishop);
        place(&mut board, 7, 0, Color::White, PieceType::Rook);
        place(&mut board, 4, 7, Color::Black, PieceType::King);
        assert!(!contains(&board.generate_king_moves(), sq(4, 0), sq(6, 0)));
    }

    #[test]
    fn castling_blocked_by_check() {
        // king in check
        let mut board = Board {
            active_color: Color::White,
            castling_rights: CASTLE_WHITE_KING | CASTLE_WHITE_QUEEN,
            ..Board::default()
        };
        place(&mut board, 4, 0, Color::White, PieceType::King);
        place(&mut board, 0, 0, Color::White, PieceType::Rook);
        place(&mut board, 7, 0, Color::White, PieceType::Rook);
        place(&mut board, 4, 7, Color::Black, PieceType::Rook);
        place(&mut board, 7, 7, Color::Black, PieceType::King);
        let moves = board.generate_king_moves();
        assert!(!contains(&moves, sq(4, 0), sq(6, 0)));
        assert!(!contains(&moves, sq(4, 0), sq(2, 0)));

        // f1 attacked blocks kingside
        let mut board = Board {
            active_color: Color::White,
            castling_rights: CASTLE_WHITE_KING,
            ..Board::default()
        };
        place(&mut board, 4, 0, Color::White, PieceType::King);
        place(&mut board, 7, 0, Color::White, PieceType::Rook);
        place(&mut board, 5, 7, Color::Black, PieceType::Rook); // f8 attacks f1
        place(&mut board, 4, 7, Color::Black, PieceType::King);
        assert!(!contains(&board.generate_king_moves(), sq(4, 0), sq(6, 0)));

        // b1 attacked does NOT block queenside
        let mut board = Board {
            active_color: Color::White,
            castling_rights: CASTLE_WHITE_QUEEN,
            ..Board::default()
        };
        place(&mut board, 4, 0, Color::White, PieceType::King);
        place(&mut board, 0, 0, Color::White, PieceType::Rook);
        place(&mut board, 1, 7, Color::Black, PieceType::Rook); // b8 attacks b1
        place(&mut board, 7, 7, Color::Black, PieceType::King);
        assert!(contains(&board.generate_king_moves(), sq(4, 0), sq(2, 0)));
    }

    #[test]
    fn castling_king_not_on_start_square() {
        let mut board = Board {
            active_color: Color::White,
            castling_rights: CASTLE_ALL,
            ..Board::default()
        };
        place(&mut board, 5, 0, Color::White, PieceType::King); // f1
        place(&mut board, 0, 0, Color::White, PieceType::Rook);
        place(&mut board, 7, 0, Color::White, PieceType::Rook);
        place(&mut board, 4, 7, Color::Black, PieceType::King);
        assert_eq!(board.generate_castling_moves(sq(5, 0)).len(), 0);
    }

    #[test]
    fn castling_in_legal_moves() {
        let mut board = Board {
            active_color: Color::White,
            castling_rights: CASTLE_WHITE_KING | CASTLE_WHITE_QUEEN,
            ..Board::default()
        };
        place(&mut board, 4, 0, Color::White, PieceType::King);
        place(&mut board, 0, 0, Color::White, PieceType::Rook);
        place(&mut board, 7, 0, Color::White, PieceType::Rook);
        place(&mut board, 4, 7, Color::Black, PieceType::King);
        let legal = board.legal_moves();
        assert!(contains(&legal, sq(4, 0), sq(6, 0)));
        assert!(contains(&legal, sq(4, 0), sq(2, 0)));

        // starting position has no castling
        let board = Board::new();
        let legal = board.legal_moves();
        assert!(!contains(&legal, sq(4, 0), sq(6, 0)));
        assert!(!contains(&legal, sq(4, 0), sq(2, 0)));
    }

    #[test]
    fn pawn_promotion_generation() {
        let mut board = Board::new();
        for i in 0..64 {
            board.squares[i] = Piece::EMPTY;
        }
        place(&mut board, 4, 6, Color::White, PieceType::Pawn); // e7
        place(&mut board, 4, 0, Color::White, PieceType::King);
        let moves = board.generate_pawn_moves();
        let e7: Vec<_> = moves.iter().filter(|m| m.from == sq(4, 6)).collect();
        assert_eq!(e7.len(), 4);
        let promos: Vec<_> = e7.iter().map(|m| m.promotion).collect();
        for p in [
            PieceType::Queen,
            PieceType::Rook,
            PieceType::Bishop,
            PieceType::Knight,
        ] {
            assert!(promos.contains(&p));
        }

        // promotion capture: 4 forward + 4 + 4 = 12
        let mut board = Board::new();
        for i in 0..64 {
            board.squares[i] = Piece::EMPTY;
        }
        place(&mut board, 4, 6, Color::White, PieceType::Pawn);
        place(&mut board, 3, 7, Color::Black, PieceType::Rook);
        place(&mut board, 5, 7, Color::Black, PieceType::Knight);
        place(&mut board, 4, 0, Color::White, PieceType::King);
        let moves = board.generate_pawn_moves();
        assert_eq!(moves.iter().filter(|m| m.from == sq(4, 6)).count(), 12);
    }

    #[test]
    fn pawn_promotion_make_move() {
        let mut board = Board::new();
        for i in 0..64 {
            board.squares[i] = Piece::EMPTY;
        }
        place(&mut board, 4, 6, Color::White, PieceType::Pawn);
        place(&mut board, 4, 0, Color::White, PieceType::King);
        place(&mut board, 0, 7, Color::Black, PieceType::King);
        board.make_move(Move::parse("e7e8q").unwrap()).unwrap();
        assert_eq!(
            board.squares[sq(4, 7).index()].piece_type(),
            PieceType::Queen
        );
        assert_eq!(board.squares[sq(4, 7).index()].color(), Color::White);

        // promotion without piece fails with promotion message
        let mut board = Board::new();
        for i in 0..64 {
            board.squares[i] = Piece::EMPTY;
        }
        place(&mut board, 4, 6, Color::White, PieceType::Pawn);
        place(&mut board, 4, 0, Color::White, PieceType::King);
        place(&mut board, 0, 7, Color::Black, PieceType::King);
        let err = board.make_move(Move::parse("e7e8").unwrap()).unwrap_err();
        assert!(err.to_string().contains("promotion"));

        // pawn on 6th rank doesn't require promotion
        let mut board = Board::new();
        for i in 0..64 {
            board.squares[i] = Piece::EMPTY;
        }
        place(&mut board, 4, 5, Color::White, PieceType::Pawn);
        place(&mut board, 4, 0, Color::White, PieceType::King);
        place(&mut board, 0, 7, Color::Black, PieceType::King);
        board.make_move(Move::parse("e6e7").unwrap()).unwrap();
        assert_eq!(
            board.squares[sq(4, 6).index()].piece_type(),
            PieceType::Pawn
        );

        // illegal pawn jump to promotion rank: illegal, not promotion error
        let mut board = Board::new();
        for i in 0..64 {
            board.squares[i] = Piece::EMPTY;
        }
        place(&mut board, 4, 3, Color::White, PieceType::Pawn);
        place(&mut board, 4, 0, Color::White, PieceType::King);
        place(&mut board, 0, 7, Color::Black, PieceType::King);
        let err = board.make_move(Move::parse("e4e8").unwrap()).unwrap_err();
        assert!(!err.to_string().contains("promotion"));
        assert!(err.to_string().contains("illegal"));
    }
}
