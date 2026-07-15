//! The `Board` type: full game state plus move application and Zobrist maintenance.

use std::collections::HashMap;
use std::fmt;

use thiserror::Error;

use crate::moves::Move;
use crate::types::{Color, Piece, PieceType, Square};
use crate::zobrist::{hash_piece, tables};

/// Castling right: White kingside (K).
pub const CASTLE_WHITE_KING: u8 = 1 << 0;
/// Castling right: White queenside (Q).
pub const CASTLE_WHITE_QUEEN: u8 = 1 << 1;
/// Castling right: Black kingside (k).
pub const CASTLE_BLACK_KING: u8 = 1 << 2;
/// Castling right: Black queenside (q).
pub const CASTLE_BLACK_QUEEN: u8 = 1 << 3;
/// All castling rights combined.
pub const CASTLE_ALL: u8 =
    CASTLE_WHITE_KING | CASTLE_WHITE_QUEEN | CASTLE_BLACK_KING | CASTLE_BLACK_QUEEN;

/// Error returned by [`Board::make_move`] when a move cannot be applied.
#[derive(Debug, Error, PartialEq, Eq)]
pub enum MoveError {
    /// A pawn reached its promotion rank but no promotion piece was specified.
    #[error("pawn promotion requires specifying a piece (q, r, b, n)")]
    PromotionRequired,
    /// The move is not legal in the current position.
    #[error("illegal move: {0}")]
    Illegal(String),
}

/// The complete state of a chess game.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Board {
    /// All 64 squares, indexed `rank * 8 + file` (a1 = 0, h8 = 63).
    pub squares: [Piece; 64],
    /// The color of the player to move.
    pub active_color: Color,
    /// Castling rights bitmask (see the `CASTLE_*` constants).
    pub castling_rights: u8,
    /// En passant target square, or -1 if none.
    pub en_passant_sq: i8,
    /// Half-moves since the last pawn move or capture (fifty-move rule).
    pub half_move_clock: u8,
    /// Current full move number, starting at 1.
    pub full_move_num: u16,
    /// Zobrist hash of the current position.
    pub hash: u64,
    /// Zobrist hashes of previous positions (for repetition detection).
    pub history: Vec<u64>,
}

impl Default for Board {
    /// Mirrors Go's zero-value `Board{}`: all squares empty, White to move, no
    /// castling rights, en passant square 0, clocks 0, empty history.
    fn default() -> Self {
        Board {
            squares: [Piece::EMPTY; 64],
            active_color: Color::White,
            castling_rights: 0,
            en_passant_sq: 0,
            half_move_clock: 0,
            full_move_num: 0,
            hash: 0,
            history: Vec::new(),
        }
    }
}

impl Board {
    /// Creates a board in the standard starting position.
    pub fn new() -> Board {
        let mut b = Board {
            squares: [Piece::EMPTY; 64],
            active_color: Color::White,
            castling_rights: CASTLE_ALL,
            en_passant_sq: -1,
            half_move_clock: 0,
            full_move_num: 1,
            hash: 0,
            history: Vec::new(),
        };

        let back_rank = [
            PieceType::Rook,
            PieceType::Knight,
            PieceType::Bishop,
            PieceType::Queen,
            PieceType::King,
            PieceType::Bishop,
            PieceType::Knight,
            PieceType::Rook,
        ];

        for (file, &pt) in back_rank.iter().enumerate() {
            b.squares[file] = Piece::new(Color::White, pt);
            b.squares[8 + file] = Piece::new(Color::White, PieceType::Pawn);
            b.squares[48 + file] = Piece::new(Color::Black, PieceType::Pawn);
            b.squares[56 + file] = Piece::new(Color::Black, pt);
        }

        b.hash = b.compute_hash();
        b.history.push(b.hash);

        b
    }

    /// Returns the piece at the given square, or [`Piece::EMPTY`] for invalid squares.
    pub fn piece_at(&self, sq: Square) -> Piece {
        if !sq.is_valid() {
            return Piece::EMPTY;
        }
        self.squares[sq.index()]
    }

    /// Returns a deep copy of the board.
    pub fn copy(&self) -> Board {
        self.clone()
    }

    /// Applies a move after validating that it is legal.
    ///
    /// Returns [`MoveError::PromotionRequired`] if a pawn reaches its promotion rank
    /// without a promotion piece, or [`MoveError::Illegal`] if the move is not legal.
    pub fn make_move(&mut self, m: Move) -> Result<(), MoveError> {
        let piece = self.squares[m.from.index()];

        // Reject a pawn that reaches the promotion rank without a promotion piece.
        if piece.piece_type() == PieceType::Pawn {
            let from_rank = m.from.rank();
            let to_rank = m.to.rank();
            let is_valid_promotion =
                (piece.color() == Color::White && from_rank == 6 && to_rank == 7)
                    || (piece.color() == Color::Black && from_rank == 1 && to_rank == 0);
            if is_valid_promotion && m.promotion == PieceType::Empty {
                return Err(MoveError::PromotionRequired);
            }
        }

        if !self.is_legal_move(m) {
            return Err(MoveError::Illegal(m.to_string()));
        }

        self.apply_move(m);
        Ok(())
    }

    /// Applies a move without any legality checking. Internal to move generation and
    /// perft; external callers should use [`Board::make_move`].
    pub(crate) fn apply_move(&mut self, m: Move) {
        let z = tables();
        let piece = self.squares[m.from.index()];
        let captured_piece = self.squares[m.to.index()];

        let old_castling_rights = self.castling_rights;
        let old_en_passant_sq = self.en_passant_sq;

        // Fifty-move clock: reset on pawn moves or captures, increment otherwise.
        let is_capture = !captured_piece.is_empty();
        let is_pawn_move = piece.piece_type() == PieceType::Pawn;
        if is_pawn_move || is_capture {
            self.half_move_clock = 0;
        } else {
            self.half_move_clock = self.half_move_clock.wrapping_add(1);
        }

        // Zobrist: XOR out the old en passant file.
        if self.en_passant_sq >= 0 {
            let old_ep_file = Square(self.en_passant_sq).file() as usize;
            self.hash ^= z.en_passant[old_ep_file];
        }

        // Zobrist: XOR out the moving piece from its source square.
        self.hash ^= hash_piece(piece, m.from);

        // Zobrist: XOR out the captured piece (if any).
        if !captured_piece.is_empty() {
            self.hash ^= hash_piece(captured_piece, m.to);
        }

        // Handle en passant capture: remove the captured pawn.
        if piece.piece_type() == PieceType::Pawn
            && old_en_passant_sq >= 0
            && m.to == Square(old_en_passant_sq)
        {
            let mut captured_pawn_rank = m.to.rank();
            if piece.color() == Color::White {
                captured_pawn_rank -= 1;
            } else {
                captured_pawn_rank += 1;
            }
            let ep_captured_sq = Square::new(m.to.file(), captured_pawn_rank);
            let captured_pawn = self.squares[ep_captured_sq.index()];
            self.hash ^= hash_piece(captured_pawn, ep_captured_sq);
            self.squares[ep_captured_sq.index()] = Piece::EMPTY;
        }

        // Move the piece.
        self.squares[m.to.index()] = piece;
        self.squares[m.from.index()] = Piece::EMPTY;

        // Handle pawn promotion.
        let mut final_piece = piece;
        if piece.piece_type() == PieceType::Pawn && m.promotion != PieceType::Empty {
            let promoted = Piece::new(piece.color(), m.promotion);
            self.squares[m.to.index()] = promoted;
            final_piece = promoted;
        }

        // Zobrist: XOR in the final piece at the destination.
        self.hash ^= hash_piece(final_piece, m.to);

        // Handle castling rook movement.
        if piece.piece_type() == PieceType::King {
            let file_diff = m.to.file() - m.from.file();
            if file_diff == 2 {
                // Kingside: rook h-file -> f-file.
                let rank = m.from.rank();
                let rook_from = Square::new(7, rank);
                let rook_to = Square::new(5, rank);
                let rook = self.squares[rook_from.index()];
                self.hash ^= hash_piece(rook, rook_from);
                self.hash ^= hash_piece(rook, rook_to);
                self.squares[rook_to.index()] = rook;
                self.squares[rook_from.index()] = Piece::EMPTY;
            } else if file_diff == -2 {
                // Queenside: rook a-file -> d-file.
                let rank = m.from.rank();
                let rook_from = Square::new(0, rank);
                let rook_to = Square::new(3, rank);
                let rook = self.squares[rook_from.index()];
                self.hash ^= hash_piece(rook, rook_from);
                self.hash ^= hash_piece(rook, rook_to);
                self.squares[rook_to.index()] = rook;
                self.squares[rook_from.index()] = Piece::EMPTY;
            }
        }

        // Update castling rights: king move removes both rights for that color.
        if piece.piece_type() == PieceType::King {
            if piece.color() == Color::White {
                self.castling_rights &= !(CASTLE_WHITE_KING | CASTLE_WHITE_QUEEN);
            } else {
                self.castling_rights &= !(CASTLE_BLACK_KING | CASTLE_BLACK_QUEEN);
            }
        }

        // Rook moving off its original square removes that side's right.
        if piece.piece_type() == PieceType::Rook {
            self.clear_castle_for_square(m.from);
        }

        // Capturing on a rook's original square removes that castling right.
        self.clear_castle_for_square(m.to);

        // Zobrist: update castling rights hash if changed.
        if self.castling_rights != old_castling_rights {
            self.hash ^= z.castling[old_castling_rights as usize];
            self.hash ^= z.castling[self.castling_rights as usize];
        }

        // Set en passant square if a pawn moved two squares.
        if piece.piece_type() == PieceType::Pawn {
            let rank_diff = m.to.rank() - m.from.rank();
            if rank_diff == 2 || rank_diff == -2 {
                let ep_rank = (m.from.rank() + m.to.rank()) / 2;
                self.en_passant_sq = Square::new(m.from.file(), ep_rank).0;
            } else {
                self.en_passant_sq = -1;
            }
        } else {
            self.en_passant_sq = -1;
        }

        // Zobrist: XOR in the new en passant file.
        if self.en_passant_sq >= 0 {
            let new_ep_file = Square(self.en_passant_sq).file() as usize;
            self.hash ^= z.en_passant[new_ep_file];
        }

        // Zobrist: toggle side to move.
        self.hash ^= z.side_to_move;

        // Toggle active color and bump full move number after Black's move.
        if self.active_color == Color::White {
            self.active_color = Color::Black;
        } else {
            self.active_color = Color::White;
            self.full_move_num += 1;
        }

        self.history.push(self.hash);
    }

    /// Clears the castling right associated with a rook's home square, if `sq` is one.
    fn clear_castle_for_square(&mut self, sq: Square) {
        if sq == Square::new(0, 0) {
            self.castling_rights &= !CASTLE_WHITE_QUEEN;
        } else if sq == Square::new(7, 0) {
            self.castling_rights &= !CASTLE_WHITE_KING;
        } else if sq == Square::new(0, 7) {
            self.castling_rights &= !CASTLE_BLACK_QUEEN;
        } else if sq == Square::new(7, 7) {
            self.castling_rights &= !CASTLE_BLACK_KING;
        }
    }

    /// Returns true if the active color's king is under attack.
    pub fn in_check(&self) -> bool {
        let mut king_square = crate::types::NO_SQUARE;
        for sq in 0..64i8 {
            let piece = self.squares[sq as usize];
            if piece.piece_type() == PieceType::King && piece.color() == self.active_color {
                king_square = Square(sq);
                break;
            }
        }

        if king_square == crate::types::NO_SQUARE {
            return false;
        }

        self.is_square_attacked(king_square, self.active_color.opponent())
    }

    /// Counts all leaf nodes at the given depth (perft), the gold standard for
    /// validating move generation.
    pub fn perft(&self, depth: i32) -> u64 {
        if depth == 0 {
            return 1;
        }

        let moves = self.legal_moves();
        if depth == 1 {
            return moves.len() as u64;
        }

        let mut nodes: u64 = 0;
        for m in moves {
            let mut board_copy = self.copy();
            board_copy.apply_move(m);
            nodes += board_copy.perft(depth - 1);
        }
        nodes
    }

    /// Perft with a per-move breakdown: maps each move to its node count at depth-1.
    pub fn divide(&self, depth: i32) -> HashMap<String, u64> {
        let mut result = HashMap::new();
        for m in self.legal_moves() {
            let mut board_copy = self.copy();
            board_copy.apply_move(m);

            let nodes = if depth <= 1 {
                1
            } else {
                board_copy.perft(depth - 1)
            };
            result.insert(m.to_string(), nodes);
        }
        result
    }
}

impl fmt::Display for Board {
    /// Text representation from White's perspective (rank 8 at top). Uppercase for
    /// White pieces, lowercase for Black, `.` for empty squares.
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        let piece_chars: [u8; 7] = [b'.', b'P', b'N', b'B', b'R', b'Q', b'K'];
        let mut result = String::new();

        for rank in (0..8i32).rev() {
            result.push((b'1' + rank as u8) as char);
            result.push(' ');

            for file in 0..8i32 {
                let sq = (rank * 8 + file) as usize;
                let piece = self.squares[sq];

                let mut ch = if piece.is_empty() {
                    b'.'
                } else {
                    piece_chars[piece.piece_type().as_u8() as usize]
                };
                if !piece.is_empty() && piece.color() == Color::Black {
                    ch = ch - b'A' + b'a';
                }
                result.push(ch as char);
                if file < 7 {
                    result.push(' ');
                }
            }
            result.push('\n');
        }

        result.push_str("  a b c d e f g h");
        write!(f, "{}", result)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::moves::Move;

    fn sq(file: i32, rank: i32) -> Square {
        Square::new(file, rank)
    }

    fn place(b: &mut Board, file: i32, rank: i32, color: Color, pt: PieceType) {
        b.squares[sq(file, rank).index()] = Piece::new(color, pt);
    }

    fn white_board() -> Board {
        Board {
            active_color: Color::White,
            ..Board::default()
        }
    }

    #[test]
    fn new_board_metadata() {
        let board = Board::new();
        assert_eq!(board.active_color, Color::White);
        assert_eq!(board.castling_rights, CASTLE_ALL);
        assert_eq!(board.en_passant_sq, -1);
        assert_eq!(board.half_move_clock, 0);
        assert_eq!(board.full_move_num, 1);
        assert_ne!(board.hash, 0);
        assert_eq!(board.history.len(), 1);
        assert_eq!(board.history[0], board.hash);
    }

    #[test]
    fn new_board_starting_position() {
        let board = Board::new();
        let back = [
            PieceType::Rook,
            PieceType::Knight,
            PieceType::Bishop,
            PieceType::Queen,
            PieceType::King,
            PieceType::Bishop,
            PieceType::Knight,
            PieceType::Rook,
        ];
        for (i, &pt) in back.iter().enumerate() {
            let p = board.piece_at(Square(i as i8));
            assert_eq!(p.piece_type(), pt);
            assert_eq!(p.color(), Color::White);
            let wp = board.piece_at(Square((8 + i) as i8));
            assert_eq!(wp.piece_type(), PieceType::Pawn);
            assert_eq!(wp.color(), Color::White);
            let bp = board.piece_at(Square((48 + i) as i8));
            assert_eq!(bp.piece_type(), PieceType::Pawn);
            assert_eq!(bp.color(), Color::Black);
            let b = board.piece_at(Square((56 + i) as i8));
            assert_eq!(b.piece_type(), pt);
            assert_eq!(b.color(), Color::Black);
        }
        for i in 16..48 {
            assert!(board.piece_at(Square(i)).is_empty());
        }
        let count = (0..64)
            .filter(|&i| !board.piece_at(Square(i)).is_empty())
            .count();
        assert_eq!(count, 32);
    }

    #[test]
    fn piece_at_invalid_square() {
        let board = Board::new();
        for s in [crate::types::NO_SQUARE, Square(-5), Square(64), Square(100)] {
            assert!(board.piece_at(s).is_empty());
        }
    }

    #[test]
    fn castling_rights_bits() {
        assert_eq!(CASTLE_WHITE_KING, 1);
        assert_eq!(CASTLE_WHITE_QUEEN, 2);
        assert_eq!(CASTLE_BLACK_KING, 4);
        assert_eq!(CASTLE_BLACK_QUEEN, 8);
        assert_eq!(CASTLE_ALL, 15);
    }

    #[test]
    fn board_string() {
        let board = Board::new();
        let expected = "8 r n b q k b n r\n7 p p p p p p p p\n6 . . . . . . . .\n5 . . . . . . . .\n4 . . . . . . . .\n3 . . . . . . . .\n2 P P P P P P P P\n1 R N B Q K B N R\n  a b c d e f g h";
        assert_eq!(board.to_string(), expected);
    }

    #[test]
    fn in_check_scenarios() {
        let board = Board::new();
        assert!(!board.in_check());
        let mut board = Board::new();
        board.active_color = Color::Black;
        assert!(!board.in_check());

        // rook horizontal check
        let mut b = white_board();
        place(&mut b, 4, 3, Color::White, PieceType::King);
        place(&mut b, 0, 3, Color::Black, PieceType::Rook);
        assert!(b.in_check());

        // rook blocked by own piece
        let mut b = white_board();
        place(&mut b, 4, 0, Color::White, PieceType::King);
        place(&mut b, 4, 7, Color::Black, PieceType::Rook);
        place(&mut b, 4, 1, Color::White, PieceType::Pawn);
        assert!(!b.in_check());

        // bishop check
        let mut b = white_board();
        place(&mut b, 4, 3, Color::White, PieceType::King);
        place(&mut b, 7, 6, Color::Black, PieceType::Bishop);
        assert!(b.in_check());

        // knight check
        let mut b = white_board();
        place(&mut b, 4, 3, Color::White, PieceType::King);
        place(&mut b, 5, 5, Color::Black, PieceType::Knight);
        assert!(b.in_check());

        // pawn check
        let mut b = white_board();
        place(&mut b, 4, 3, Color::White, PieceType::King);
        place(&mut b, 3, 4, Color::Black, PieceType::Pawn);
        assert!(b.in_check());

        // pawn in front does not check
        let mut b = white_board();
        place(&mut b, 4, 3, Color::White, PieceType::King);
        place(&mut b, 4, 4, Color::Black, PieceType::Pawn);
        assert!(!b.in_check());

        // own pieces do not check
        let mut b = white_board();
        place(&mut b, 4, 3, Color::White, PieceType::King);
        place(&mut b, 4, 7, Color::White, PieceType::Queen);
        place(&mut b, 0, 3, Color::White, PieceType::Rook);
        assert!(!b.in_check());

        // double check
        let mut b = white_board();
        place(&mut b, 4, 3, Color::White, PieceType::King);
        place(&mut b, 4, 7, Color::Black, PieceType::Queen);
        place(&mut b, 5, 5, Color::Black, PieceType::Knight);
        assert!(b.in_check());

        // no king returns false
        let b = white_board();
        assert!(!b.in_check());
    }

    #[test]
    fn apply_move_castling() {
        // white kingside
        let mut b = Board {
            active_color: Color::White,
            castling_rights: CASTLE_ALL,
            ..Board::default()
        };
        place(&mut b, 4, 0, Color::White, PieceType::King);
        place(&mut b, 7, 0, Color::White, PieceType::Rook);
        b.apply_move(Move::new(sq(4, 0), sq(6, 0)));
        assert_eq!(b.squares[sq(6, 0).index()].piece_type(), PieceType::King);
        assert_eq!(b.squares[sq(5, 0).index()].piece_type(), PieceType::Rook);
        assert!(b.squares[sq(4, 0).index()].is_empty());
        assert!(b.squares[sq(7, 0).index()].is_empty());

        // white queenside
        let mut b = Board {
            active_color: Color::White,
            castling_rights: CASTLE_ALL,
            ..Board::default()
        };
        place(&mut b, 4, 0, Color::White, PieceType::King);
        place(&mut b, 0, 0, Color::White, PieceType::Rook);
        b.apply_move(Move::new(sq(4, 0), sq(2, 0)));
        assert_eq!(b.squares[sq(2, 0).index()].piece_type(), PieceType::King);
        assert_eq!(b.squares[sq(3, 0).index()].piece_type(), PieceType::Rook);
        assert!(b.squares[sq(0, 0).index()].is_empty());

        // black kingside
        let mut b = Board {
            active_color: Color::Black,
            castling_rights: CASTLE_ALL,
            ..Board::default()
        };
        place(&mut b, 4, 7, Color::Black, PieceType::King);
        place(&mut b, 7, 7, Color::Black, PieceType::Rook);
        b.apply_move(Move::new(sq(4, 7), sq(6, 7)));
        assert_eq!(b.squares[sq(6, 7).index()].piece_type(), PieceType::King);
        assert_eq!(b.squares[sq(5, 7).index()].piece_type(), PieceType::Rook);

        // normal king move doesn't move rook
        let mut b = Board {
            active_color: Color::White,
            castling_rights: CASTLE_ALL,
            ..Board::default()
        };
        place(&mut b, 4, 0, Color::White, PieceType::King);
        place(&mut b, 7, 0, Color::White, PieceType::Rook);
        b.apply_move(Move::new(sq(4, 0), sq(5, 0)));
        assert_eq!(b.squares[sq(5, 0).index()].piece_type(), PieceType::King);
        assert_eq!(b.squares[sq(7, 0).index()].piece_type(), PieceType::Rook);
    }

    #[test]
    fn en_passant_square_set_and_cleared() {
        // white e2-e4 sets e3
        let mut b = Board::new();
        b.apply_move(Move::new(sq(4, 1), sq(4, 3)));
        assert_eq!(b.en_passant_sq, sq(4, 2).0);

        // black d7-d5 sets d6
        let mut b = Board::new();
        b.active_color = Color::Black;
        b.apply_move(Move::new(sq(3, 6), sq(3, 4)));
        assert_eq!(b.en_passant_sq, sq(3, 5).0);

        // single square pawn move clears
        let mut b = Board::new();
        b.en_passant_sq = sq(4, 2).0;
        b.apply_move(Move::new(sq(4, 1), sq(4, 2)));
        assert_eq!(b.en_passant_sq, -1);

        // knight move clears
        let mut b = Board::new();
        b.en_passant_sq = sq(4, 2).0;
        b.apply_move(Move::new(sq(6, 0), sq(5, 2)));
        assert_eq!(b.en_passant_sq, -1);

        // all files white double push
        for file in 0..8 {
            let mut b = Board::new();
            b.apply_move(Move::new(sq(file, 1), sq(file, 3)));
            assert_eq!(b.en_passant_sq, sq(file, 2).0);
        }
    }

    #[test]
    fn castling_rights_update() {
        // white king move removes both white rights
        let mut b = Board {
            active_color: Color::White,
            castling_rights: CASTLE_ALL,
            ..Board::default()
        };
        place(&mut b, 4, 0, Color::White, PieceType::King);
        b.apply_move(Move::new(sq(4, 0), sq(4, 1)));
        assert_eq!(
            b.castling_rights & (CASTLE_WHITE_KING | CASTLE_WHITE_QUEEN),
            0
        );
        assert_ne!(
            b.castling_rights & (CASTLE_BLACK_KING | CASTLE_BLACK_QUEEN),
            0
        );

        // h1 rook move removes white kingside only
        let mut b = Board {
            active_color: Color::White,
            castling_rights: CASTLE_ALL,
            ..Board::default()
        };
        place(&mut b, 7, 0, Color::White, PieceType::Rook);
        b.apply_move(Move::new(sq(7, 0), sq(7, 1)));
        assert_eq!(b.castling_rights & CASTLE_WHITE_KING, 0);
        assert_ne!(b.castling_rights & CASTLE_WHITE_QUEEN, 0);

        // capture on a1 removes white queenside
        let mut b = Board {
            active_color: Color::Black,
            castling_rights: CASTLE_ALL,
            ..Board::default()
        };
        place(&mut b, 0, 0, Color::White, PieceType::Rook);
        place(&mut b, 1, 1, Color::Black, PieceType::Bishop);
        b.apply_move(Move::new(sq(1, 1), sq(0, 0)));
        assert_eq!(b.castling_rights & CASTLE_WHITE_QUEEN, 0);
        assert_ne!(b.castling_rights & CASTLE_WHITE_KING, 0);

        // pawn move doesn't affect
        let mut b = Board {
            active_color: Color::White,
            castling_rights: CASTLE_ALL,
            ..Board::default()
        };
        place(&mut b, 4, 1, Color::White, PieceType::Pawn);
        b.apply_move(Move::new(sq(4, 1), sq(4, 3)));
        assert_eq!(b.castling_rights, CASTLE_ALL);
    }

    #[test]
    fn en_passant_capture_execution() {
        // white captures e5xd6 removing black pawn on d5
        let mut b = white_board();
        place(&mut b, 4, 0, Color::White, PieceType::King);
        place(&mut b, 4, 7, Color::Black, PieceType::King);
        place(&mut b, 4, 4, Color::White, PieceType::Pawn); // e5
        place(&mut b, 3, 4, Color::Black, PieceType::Pawn); // d5
        b.en_passant_sq = sq(3, 5).0; // d6
        let legal = b.legal_moves();
        let ep = legal
            .iter()
            .find(|m| m.from == sq(4, 4) && m.to == sq(3, 5))
            .copied();
        assert!(ep.is_some());
        b.make_move(ep.unwrap()).unwrap();
        assert_eq!(b.squares[sq(3, 5).index()].piece_type(), PieceType::Pawn);
        assert_eq!(b.squares[sq(3, 5).index()].color(), Color::White);
        assert!(b.squares[sq(4, 4).index()].is_empty());
        assert!(b.squares[sq(3, 4).index()].is_empty());
        assert_eq!(b.en_passant_sq, -1);
    }

    #[test]
    fn en_passant_pinned_not_legal() {
        let mut b = white_board();
        place(&mut b, 4, 4, Color::White, PieceType::King); // e5
        place(&mut b, 3, 4, Color::White, PieceType::Pawn); // d5
        place(&mut b, 2, 4, Color::Black, PieceType::Pawn); // c5
        place(&mut b, 0, 4, Color::Black, PieceType::Rook); // a5
        b.en_passant_sq = sq(2, 5).0; // c6
        let legal = b.legal_moves();
        assert!(!legal.iter().any(|m| m.from == sq(3, 4) && m.to == sq(2, 5)));
    }

    #[test]
    fn perft_starting_position() {
        let board = Board::new();
        assert_eq!(board.perft(0), 1);
        assert_eq!(board.perft(1), 20);
        assert_eq!(board.perft(2), 400);
        assert_eq!(board.perft(3), 8902);
    }

    #[test]
    fn perft_positions() {
        let cases: &[(&str, &[(i32, u64)])] = &[
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
                &[(1, 20), (2, 400), (3, 8902), (4, 197281)],
            ),
            (
                "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
                &[(1, 48), (2, 2039), (3, 97862), (4, 4085603)],
            ),
            (
                "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
                &[(1, 14), (2, 191), (3, 2812), (4, 43238)],
            ),
            (
                "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
                &[(1, 6), (2, 264), (3, 9467)],
            ),
            (
                "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
                &[(1, 44), (2, 1486), (3, 62379)],
            ),
            (
                "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
                &[(1, 46), (2, 2079), (3, 89890)],
            ),
        ];
        for (fen, depths) in cases {
            let board = Board::from_fen(fen).unwrap();
            for (depth, expected) in *depths {
                assert_eq!(
                    board.perft(*depth),
                    *expected,
                    "fen={} depth={}",
                    fen,
                    depth
                );
            }
        }
    }

    #[test]
    fn divide_starting_position() {
        let board = Board::new();
        let d1 = board.divide(1);
        assert_eq!(d1.len(), 20);
        assert!(d1.values().all(|&c| c == 1));
        assert_eq!(d1.values().sum::<u64>(), 20);

        let d2 = board.divide(2);
        assert_eq!(d2.len(), 20);
        for m in [
            "a2a3", "b2b3", "e2e4", "d2d4", "b1a3", "b1c3", "g1f3", "g1h3",
        ] {
            assert_eq!(d2[m], 20);
        }
        assert_eq!(d2.values().sum::<u64>(), 400);
    }

    #[test]
    fn board_copy_independence() {
        let mut original = Board::new();
        original.history.push(12345);
        let mut copied = original.copy();
        assert_eq!(copied, original);
        copied.active_color = Color::Black;
        copied.castling_rights = 0;
        copied.squares[0] = Piece::EMPTY;
        copied.history.push(11111);
        assert_eq!(original.active_color, Color::White);
        assert_eq!(original.castling_rights, CASTLE_ALL);
        assert!(!original.squares[0].is_empty());
        assert_eq!(original.history.len(), 2);
    }
}
