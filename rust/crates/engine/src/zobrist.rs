//! Zobrist hashing for position identity and repetition detection.
//!
//! The Go implementation seeds `math/rand` with a fixed constant so hashes are
//! deterministic across runs. The exact byte values are never asserted by the test
//! suite (only structural properties such as determinism and incremental/computed
//! parity), so here we fill the tables with a deterministic SplitMix64 stream seeded
//! from the same constant. Determinism within a single build is what matters.

use std::sync::OnceLock;

use crate::board::Board;
use crate::types::{Color, Piece, Square};

/// Deterministic Zobrist tables, initialized once on first use.
pub(crate) struct Zobrist {
    /// `pieces[piece_index][square]` — value for each piece type on each square.
    /// `piece_index = color * 6 + (piece_type - 1)`, giving 12 indices x 64 squares.
    pub pieces: [[u64; 64]; 12],
    /// XORed into the hash when it is Black's turn to move.
    pub side_to_move: u64,
    /// Value for each combination of castling rights (0-15).
    pub castling: [u64; 16],
    /// Value for en passant on each file (0-7).
    pub en_passant: [u64; 8],
}

/// SplitMix64 step — a fast, deterministic pseudo-random generator.
#[inline]
fn splitmix64(state: &mut u64) -> u64 {
    *state = state.wrapping_add(0x9E37_79B9_7F4A_7C15);
    let mut z = *state;
    z = (z ^ (z >> 30)).wrapping_mul(0xBF58_476D_1CE4_E5B9);
    z = (z ^ (z >> 27)).wrapping_mul(0x94D0_49BB_1331_11EB);
    z ^ (z >> 31)
}

fn build() -> Zobrist {
    // Fixed seed mirrors the Go implementation's deterministic-seed intent.
    let mut state: u64 = 0x005D_4E3C_2B1A;

    let mut pieces = [[0u64; 64]; 12];
    for piece in pieces.iter_mut() {
        for sq in piece.iter_mut() {
            *sq = splitmix64(&mut state);
        }
    }

    let side_to_move = splitmix64(&mut state);

    let mut castling = [0u64; 16];
    for c in castling.iter_mut() {
        *c = splitmix64(&mut state);
    }

    let mut en_passant = [0u64; 8];
    for e in en_passant.iter_mut() {
        *e = splitmix64(&mut state);
    }

    Zobrist {
        pieces,
        side_to_move,
        castling,
        en_passant,
    }
}

/// Returns the process-wide Zobrist tables, initializing them on first access.
pub(crate) fn tables() -> &'static Zobrist {
    static Z: OnceLock<Zobrist> = OnceLock::new();
    Z.get_or_init(build)
}

/// Returns the Zobrist table index for a piece, or -1 for empty squares.
///
/// `piece_type` is 1-6 (Pawn to King), so subtract 1 to get 0-5. Color is 0 (White)
/// or 1 (Black), multiplied by 6 to offset into the Black half.
pub(crate) fn piece_zobrist_index(p: Piece) -> i32 {
    if p.is_empty() {
        return -1;
    }
    (p.color().as_u8() as i32) * 6 + (p.piece_type().as_u8() as i32) - 1
}

/// Returns the Zobrist hash contribution for a piece on a square (0 for empty).
///
/// Used for incremental hash updates — XOR to add or remove a piece.
pub(crate) fn hash_piece(p: Piece, sq: Square) -> u64 {
    if p.is_empty() {
        return 0;
    }
    let idx = piece_zobrist_index(p) as usize;
    tables().pieces[idx][sq.index()]
}

impl Board {
    /// Computes the full Zobrist hash for the current position from scratch.
    pub fn compute_hash(&self) -> u64 {
        let z = tables();
        let mut hash: u64 = 0;

        for sq in 0..64usize {
            let piece = self.squares[sq];
            if !piece.is_empty() {
                let idx = piece_zobrist_index(piece) as usize;
                hash ^= z.pieces[idx][sq];
            }
        }

        if self.active_color == Color::Black {
            hash ^= z.side_to_move;
        }

        hash ^= z.castling[self.castling_rights as usize];

        if self.en_passant_sq >= 0 {
            let ep_file = Square(self.en_passant_sq).file() as usize;
            hash ^= z.en_passant[ep_file];
        }

        hash
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::moves::Move;
    use crate::types::{Color, Piece, PieceType};

    #[test]
    fn zobrist_initialization() {
        let z = tables();
        let mut non_zero = 0;
        for p in &z.pieces {
            for &v in p {
                if v != 0 {
                    non_zero += 1;
                }
            }
        }
        assert!(non_zero >= 100);
        assert_ne!(z.side_to_move, 0);
        assert!(z.castling.iter().filter(|&&v| v != 0).count() >= 8);
        assert!(z.en_passant.iter().filter(|&&v| v != 0).count() >= 4);
    }

    #[test]
    fn new_board_has_non_zero_hash() {
        assert_ne!(Board::new().hash, 0);
    }

    #[test]
    fn new_board_hash_is_deterministic() {
        assert_eq!(Board::new().hash, Board::new().hash);
    }

    #[test]
    fn hash_changes_after_move() {
        let mut board = Board::new();
        let initial = board.hash;
        board.make_move(Move::parse("e2e4").unwrap()).unwrap();
        assert_ne!(board.hash, initial);
    }

    #[test]
    fn different_moves_produce_different_hashes() {
        let mut b1 = Board::new();
        b1.make_move(Move::parse("e2e4").unwrap()).unwrap();
        let mut b2 = Board::new();
        b2.make_move(Move::parse("d2d4").unwrap()).unwrap();
        assert_ne!(b1.hash, b2.hash);
    }

    #[test]
    fn history_grows() {
        let mut board = Board::new();
        assert_eq!(board.history.len(), 1);
        board.make_move(Move::parse("e2e4").unwrap()).unwrap();
        assert_eq!(board.history.len(), 2);
        board.make_move(Move::parse("e7e5").unwrap()).unwrap();
        assert_eq!(board.history.len(), 3);
    }

    #[test]
    fn initial_position_in_history() {
        let board = Board::new();
        assert_eq!(board.history[0], board.hash);
    }

    #[test]
    fn threefold_repetition_with_starting_position() {
        let mut board = Board::new();
        let starting = board.hash;
        let count = |b: &Board| b.history.iter().filter(|&&h| h == starting).count();
        assert_eq!(count(&board), 1);
        for m in ["g1f3", "g8f6", "f3g1", "f6g8"] {
            board.make_move(Move::parse(m).unwrap()).unwrap();
        }
        assert_eq!(count(&board), 2);
        assert_eq!(board.hash, starting);
        for m in ["g1f3", "g8f6", "f3g1", "f6g8"] {
            board.make_move(Move::parse(m).unwrap()).unwrap();
        }
        assert_eq!(count(&board), 3);
    }

    #[test]
    fn multiple_games_have_independent_history() {
        let mut b1 = Board::new();
        let b2 = Board::new();
        assert_eq!(b1.history.len(), 1);
        assert_eq!(b2.history.len(), 1);
        b1.make_move(Move::parse("e2e4").unwrap()).unwrap();
        assert_eq!(b1.history.len(), 2);
        assert_eq!(b2.history.len(), 1);
    }

    #[test]
    fn hash_matches_computed_hash() {
        let mut board = Board::new();
        for m in ["e2e4", "e7e5", "g1f3", "b8c6", "f1b5"] {
            board.make_move(Move::parse(m).unwrap()).unwrap();
            assert_eq!(board.hash, board.compute_hash());
        }
    }

    #[test]
    fn same_position_same_hash() {
        let mut b1 = Board::new();
        for m in ["g1f3", "b8c6", "b1c3", "g8f6"] {
            b1.make_move(Move::parse(m).unwrap()).unwrap();
        }
        let mut b2 = Board::new();
        for m in ["b1c3", "g8f6", "g1f3", "b8c6"] {
            b2.make_move(Move::parse(m).unwrap()).unwrap();
        }
        assert_eq!(b1.hash, b2.hash);
    }

    #[test]
    fn copy_preserves_hash() {
        let mut board = Board::new();
        board.make_move(Move::parse("e2e4").unwrap()).unwrap();
        let copy = board.copy();
        assert_eq!(copy.hash, board.hash);
    }

    #[test]
    fn capture_changes_hash() {
        let mut board = Board::new();
        for m in ["e2e4", "d7d5"] {
            board.make_move(Move::parse(m).unwrap()).unwrap();
        }
        let before = board.hash;
        board.make_move(Move::parse("e4d5").unwrap()).unwrap();
        assert_ne!(board.hash, before);
        assert_eq!(board.hash, board.compute_hash());
    }

    #[test]
    fn castling_changes_hash() {
        let mut board = Board::new();
        board.squares[Square::new(5, 0).index()] = Piece::EMPTY;
        board.squares[Square::new(6, 0).index()] = Piece::EMPTY;
        board.hash = board.compute_hash();
        let before = board.hash;
        board.make_move(Move::parse("e1g1").unwrap()).unwrap();
        assert_ne!(board.hash, before);
        assert_eq!(board.hash, board.compute_hash());
    }

    #[test]
    fn en_passant_capture_changes_hash() {
        let mut board = Board::new();
        for m in ["e2e4", "a7a6", "e4e5", "d7d5"] {
            board.make_move(Move::parse(m).unwrap()).unwrap();
        }
        let before = board.hash;
        board.make_move(Move::parse("e5d6").unwrap()).unwrap();
        assert_ne!(board.hash, before);
        assert_eq!(board.hash, board.compute_hash());
    }

    #[test]
    fn promotion_changes_hash() {
        let mut board = Board::new();
        for i in 0..64 {
            board.squares[i] = Piece::EMPTY;
        }
        board.squares[Square::new(4, 0).index()] = Piece::new(Color::White, PieceType::King);
        board.squares[Square::new(4, 7).index()] = Piece::new(Color::Black, PieceType::King);
        board.squares[Square::new(0, 6).index()] = Piece::new(Color::White, PieceType::Pawn);
        board.castling_rights = 0;
        board.en_passant_sq = -1;
        board.hash = board.compute_hash();
        let before = board.hash;
        board.make_move(Move::parse("a7a8q").unwrap()).unwrap();
        assert_ne!(board.hash, before);
        assert_eq!(board.hash, board.compute_hash());
    }

    #[test]
    fn en_passant_file_affects_hash() {
        let mut b1 = Board::new();
        b1.make_move(Move::parse("e2e4").unwrap()).unwrap();
        let mut b2 = Board::new();
        b2.make_move(Move::parse("d2d4").unwrap()).unwrap();
        assert_ne!(b1.hash, b2.hash);
    }

    #[test]
    fn castling_rights_affect_hash() {
        let b1 = Board::new();
        let mut b2 = Board::new();
        b2.castling_rights = 0;
        b2.hash = b2.compute_hash();
        assert_ne!(b1.hash, b2.hash);
    }

    #[test]
    fn side_to_move_affects_hash() {
        let b1 = Board::new();
        let mut b2 = Board::new();
        b2.active_color = Color::Black;
        b2.hash = b2.compute_hash();
        assert_ne!(b1.hash, b2.hash);
    }

    #[test]
    fn piece_zobrist_index_values() {
        let cases = [
            (Piece(0), -1),
            (Piece::new(Color::White, PieceType::Pawn), 0),
            (Piece::new(Color::White, PieceType::Knight), 1),
            (Piece::new(Color::White, PieceType::Bishop), 2),
            (Piece::new(Color::White, PieceType::Rook), 3),
            (Piece::new(Color::White, PieceType::Queen), 4),
            (Piece::new(Color::White, PieceType::King), 5),
            (Piece::new(Color::Black, PieceType::Pawn), 6),
            (Piece::new(Color::Black, PieceType::Knight), 7),
            (Piece::new(Color::Black, PieceType::Bishop), 8),
            (Piece::new(Color::Black, PieceType::Rook), 9),
            (Piece::new(Color::Black, PieceType::Queen), 10),
            (Piece::new(Color::Black, PieceType::King), 11),
        ];
        for (piece, expected) in cases {
            assert_eq!(piece_zobrist_index(piece), expected);
        }
    }
}
