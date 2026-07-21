//! FEN (Forsyth-Edwards Notation) parsing and serialization.

use std::fmt::Write as _;

use thiserror::Error;

use crate::board::{
    Board, CASTLE_BLACK_KING, CASTLE_BLACK_QUEEN, CASTLE_WHITE_KING, CASTLE_WHITE_QUEEN,
};
use crate::types::{Color, Piece, PieceType, Square};

/// Error returned when a FEN string cannot be parsed.
#[derive(Debug, Error, PartialEq, Eq)]
pub enum FenError {
    /// The FEN did not have exactly 6 space-separated fields.
    #[error("FEN must have 6 parts, got {0}")]
    WrongPartCount(usize),
    /// The piece placement field did not have exactly 8 ranks.
    #[error("FEN piece placement must have 8 ranks, got {0}")]
    WrongRankCount(usize),
    /// An invalid piece character was encountered.
    #[error("invalid piece character: {0}")]
    InvalidPieceChar(char),
    /// A rank described more or fewer than 8 squares.
    #[error("rank {rank} has {count} squares, expected 8")]
    WrongSquareCount { rank: i32, count: i32 },
    /// A rank overflowed past 8 squares while placing a piece.
    #[error("rank {0} has too many squares, expected 8")]
    RankOverflow(i32),
    /// The active-color field was not `w` or `b`.
    #[error("invalid active color: {0} (expected 'w' or 'b')")]
    InvalidActiveColor(String),
    /// An invalid castling character was encountered.
    #[error("invalid castling character: {0}")]
    InvalidCastlingChar(char),
    /// The en passant field was malformed.
    #[error("invalid en passant square: {0}")]
    InvalidEnPassant(String),
    /// The half-move clock was not a valid non-negative integer.
    #[error("invalid half-move clock: {0}")]
    InvalidHalfMove(String),
    /// The half-move clock exceeded the representable range.
    #[error("half-move clock out of range: {0}")]
    HalfMoveOutOfRange(i64),
    /// The full move number was not a valid integer.
    #[error("invalid full move number: {0}")]
    InvalidFullMove(String),
    /// The full move number was outside the valid range.
    #[error("full move number out of range: {0}")]
    FullMoveOutOfRange(i64),
}

/// Converts a FEN character to a piece. Uppercase = White, lowercase = Black.
fn char_to_piece(c: char) -> Result<Piece, FenError> {
    let (color, upper) = if c.is_ascii_uppercase() {
        (Color::White, c)
    } else if c.is_ascii_lowercase() {
        (Color::Black, c.to_ascii_uppercase())
    } else {
        return Err(FenError::InvalidPieceChar(c));
    };

    let piece_type = match upper {
        'P' => PieceType::Pawn,
        'N' => PieceType::Knight,
        'B' => PieceType::Bishop,
        'R' => PieceType::Rook,
        'Q' => PieceType::Queen,
        'K' => PieceType::King,
        _ => return Err(FenError::InvalidPieceChar(c)),
    };

    Ok(Piece::new(color, piece_type))
}

/// Converts a piece to its FEN character (`?` for empty pieces).
fn piece_to_char(p: Piece) -> char {
    let base = match p.piece_type() {
        PieceType::Pawn => 'P',
        PieceType::Knight => 'N',
        PieceType::Bishop => 'B',
        PieceType::Rook => 'R',
        PieceType::Queen => 'Q',
        PieceType::King => 'K',
        PieceType::Empty => return '?',
    };
    if p.color() == Color::Black {
        base.to_ascii_lowercase()
    } else {
        base
    }
}

impl Board {
    /// Parses a board from a FEN string.
    pub fn from_fen(fen: &str) -> Result<Board, FenError> {
        let parts: Vec<&str> = fen.split_whitespace().collect();
        if parts.len() != 6 {
            return Err(FenError::WrongPartCount(parts.len()));
        }

        let mut b = Board {
            squares: [Piece::EMPTY; 64],
            active_color: Color::White,
            castling_rights: 0,
            en_passant_sq: -1,
            half_move_clock: 0,
            full_move_num: 1,
            hash: 0,
            history: Vec::new(),
        };

        // Field 1: piece placement (rank 8 down to rank 1).
        let ranks: Vec<&str> = parts[0].split('/').collect();
        if ranks.len() != 8 {
            return Err(FenError::WrongRankCount(ranks.len()));
        }

        for (rank_idx, rank_str) in ranks.iter().enumerate() {
            let rank = 7 - rank_idx as i32;
            let rank_str = *rank_str;
            let mut file = 0i32;

            for ch in rank_str.chars() {
                if ('1'..='8').contains(&ch) {
                    file += ch as i32 - '0' as i32;
                } else if ch == '0' || ch == '9' {
                    return Err(FenError::InvalidPieceChar(ch));
                } else {
                    let piece = char_to_piece(ch)?;
                    if file > 7 {
                        return Err(FenError::RankOverflow(rank + 1));
                    }
                    let sq = Square::new(file, rank);
                    b.squares[sq.index()] = piece;
                    file += 1;
                }
            }

            if file != 8 {
                return Err(FenError::WrongSquareCount {
                    rank: rank + 1,
                    count: file,
                });
            }
        }

        // Field 2: active color.
        b.active_color = match parts[1] {
            "w" => Color::White,
            "b" => Color::Black,
            other => return Err(FenError::InvalidActiveColor(other.to_string())),
        };

        // Field 3: castling rights.
        if parts[2] != "-" {
            for ch in parts[2].chars() {
                match ch {
                    'K' => b.castling_rights |= CASTLE_WHITE_KING,
                    'Q' => b.castling_rights |= CASTLE_WHITE_QUEEN,
                    'k' => b.castling_rights |= CASTLE_BLACK_KING,
                    'q' => b.castling_rights |= CASTLE_BLACK_QUEEN,
                    other => return Err(FenError::InvalidCastlingChar(other)),
                }
            }
        }

        // Field 4: en passant square.
        if parts[3] != "-" {
            let ep = parts[3];
            let ep_bytes = ep.as_bytes();
            if ep_bytes.len() != 2 {
                return Err(FenError::InvalidEnPassant(ep.to_string()));
            }
            let file = ep_bytes[0] as i32 - b'a' as i32;
            let rank = ep_bytes[1] as i32 - b'1' as i32;
            if !(0..=7).contains(&file) || !(0..=7).contains(&rank) {
                return Err(FenError::InvalidEnPassant(ep.to_string()));
            }
            b.en_passant_sq = Square::new(file, rank).0;
        }

        // Field 5: half-move clock.
        let half_move: i64 = parts[4]
            .parse()
            .map_err(|_| FenError::InvalidHalfMove(parts[4].to_string()))?;
        if half_move < 0 {
            return Err(FenError::InvalidHalfMove(parts[4].to_string()));
        }
        if half_move > 255 {
            return Err(FenError::HalfMoveOutOfRange(half_move));
        }
        b.half_move_clock = half_move as u8;

        // Field 6: full move number.
        let full_move: i64 = parts[5]
            .parse()
            .map_err(|_| FenError::InvalidFullMove(parts[5].to_string()))?;
        if !(1..=65535).contains(&full_move) {
            return Err(FenError::FullMoveOutOfRange(full_move));
        }
        b.full_move_num = full_move as u16;

        b.hash = b.compute_hash();
        b.history.push(b.hash);

        Ok(b)
    }

    /// Alias for [`Board::from_fen`].
    pub fn parse_fen(fen: &str) -> Result<Board, FenError> {
        Board::from_fen(fen)
    }

    /// Serializes the board to a FEN string.
    pub fn to_fen(&self) -> String {
        let mut fen = String::new();

        // Field 1: piece placement.
        for rank in (0..8i32).rev() {
            let mut empty_count = 0;
            for file in 0..8i32 {
                let sq = Square::new(file, rank);
                let piece = self.squares[sq.index()];
                if piece.is_empty() {
                    empty_count += 1;
                } else {
                    if empty_count > 0 {
                        let _ = write!(fen, "{}", empty_count);
                        empty_count = 0;
                    }
                    fen.push(piece_to_char(piece));
                }
            }
            if empty_count > 0 {
                let _ = write!(fen, "{}", empty_count);
            }
            if rank > 0 {
                fen.push('/');
            }
        }

        // Field 2: active color.
        fen.push(' ');
        fen.push(if self.active_color == Color::White {
            'w'
        } else {
            'b'
        });

        // Field 3: castling rights.
        fen.push(' ');
        let mut castling = String::new();
        if self.castling_rights & CASTLE_WHITE_KING != 0 {
            castling.push('K');
        }
        if self.castling_rights & CASTLE_WHITE_QUEEN != 0 {
            castling.push('Q');
        }
        if self.castling_rights & CASTLE_BLACK_KING != 0 {
            castling.push('k');
        }
        if self.castling_rights & CASTLE_BLACK_QUEEN != 0 {
            castling.push('q');
        }
        if castling.is_empty() {
            castling.push('-');
        }
        fen.push_str(&castling);

        // Field 4: en passant.
        fen.push(' ');
        if self.en_passant_sq < 0 {
            fen.push('-');
        } else {
            let _ = write!(fen, "{}", Square(self.en_passant_sq));
        }

        // Field 5: half-move clock.
        let _ = write!(fen, " {}", self.half_move_clock);

        // Field 6: full move number.
        let _ = write!(fen, " {}", self.full_move_num);

        fen
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::board::{CASTLE_ALL, CASTLE_BLACK_QUEEN, CASTLE_WHITE_KING};

    const START: &str = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1";

    #[test]
    fn char_to_piece_valid() {
        let cases = [
            ('P', Color::White, PieceType::Pawn),
            ('N', Color::White, PieceType::Knight),
            ('B', Color::White, PieceType::Bishop),
            ('R', Color::White, PieceType::Rook),
            ('Q', Color::White, PieceType::Queen),
            ('K', Color::White, PieceType::King),
            ('p', Color::Black, PieceType::Pawn),
            ('n', Color::Black, PieceType::Knight),
            ('b', Color::Black, PieceType::Bishop),
            ('r', Color::Black, PieceType::Rook),
            ('q', Color::Black, PieceType::Queen),
            ('k', Color::Black, PieceType::King),
        ];
        for (c, color, pt) in cases {
            let piece = char_to_piece(c).unwrap();
            assert_eq!(piece.color(), color);
            assert_eq!(piece.piece_type(), pt);
        }
        for c in ['1', '/', 'X', 'x'] {
            assert!(char_to_piece(c).is_err());
        }
    }

    #[test]
    fn from_fen_starting_position() {
        let board = Board::from_fen(START).unwrap();
        let expected = Board::new();
        assert_eq!(board.active_color, expected.active_color);
        assert_eq!(board.castling_rights, expected.castling_rights);
        assert_eq!(board.en_passant_sq, expected.en_passant_sq);
        assert_eq!(board.half_move_clock, expected.half_move_clock);
        assert_eq!(board.full_move_num, expected.full_move_num);
        for i in 0..64 {
            assert_eq!(board.squares[i], expected.squares[i]);
        }
    }

    #[test]
    fn from_fen_en_passant() {
        let board =
            Board::from_fen("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1").unwrap();
        assert_eq!(board.en_passant_sq, Square::new(4, 2).0);
    }

    #[test]
    fn from_fen_castling_rights() {
        let cases = [
            (START, CASTLE_ALL),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w K - 0 1",
                CASTLE_WHITE_KING,
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w Kq - 0 1",
                CASTLE_WHITE_KING | CASTLE_BLACK_QUEEN,
            ),
            ("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w - - 0 1", 0),
        ];
        for (fen, expected) in cases {
            assert_eq!(Board::from_fen(fen).unwrap().castling_rights, expected);
        }
    }

    #[test]
    fn from_fen_clock_values() {
        let board =
            Board::from_fen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 48 30").unwrap();
        assert_eq!(board.half_move_clock, 48);
        assert_eq!(board.full_move_num, 30);
    }

    #[test]
    fn to_fen_empty_and_start() {
        let empty = Board {
            active_color: Color::White,
            castling_rights: 0,
            en_passant_sq: -1,
            full_move_num: 1,
            ..Board::default()
        };
        assert_eq!(empty.to_fen(), "8/8/8/8/8/8/8/8 w - - 0 1");
        assert_eq!(Board::new().to_fen(), START);
    }

    #[test]
    fn round_trip() {
        let fens = [
            START,
            "8/8/8/8/8/8/8/8 w - - 0 1",
            "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
            "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
            "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
            "rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 0 2",
            "rnbqkbnr/pp1ppppp/8/2p5/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2",
            "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R b Kq e3 5 10",
            "5k2/8/8/8/8/8/8/4K2R w K - 0 1",
            "4k3/1P6/8/8/8/8/K7/8 w - - 0 1",
            "8/P1k5/K7/8/8/8/8/8 w - - 0 1",
        ];
        for fen in fens {
            let board = Board::from_fen(fen).unwrap();
            assert_eq!(board.to_fen(), fen);
        }
    }

    #[test]
    fn validation_errors() {
        let cases: &[(&str, &str)] = &[
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0",
                "6 parts",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1 extra",
                "6 parts",
            ),
            ("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP w KQkq - 0 1", "8 ranks"),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
                "8 ranks",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPXPPP/RNBQKBNR w KQkq - 0 1",
                "invalid piece character",
            ),
            (
                "rnbqkbnr/pppppppp/9/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
                "invalid piece character",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPPP/RNBQKBNR w KQkq - 0 1",
                "squares",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPP/RNBQKBNR w KQkq - 0 1",
                "squares",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/54/RNBQKBNR w KQkq - 0 1",
                "squares",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR x KQkq - 0 1",
                "invalid active color",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQX - 0 1",
                "invalid castling character",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq z9 0 1",
                "invalid en passant square",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq a0 0 1",
                "invalid en passant square",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq e 0 1",
                "invalid en passant square",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - -1 1",
                "invalid half-move clock",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - abc 1",
                "invalid half-move clock",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 256 1",
                "half-move clock out of range",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 0",
                "full move number out of range",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 -1",
                "full move number out of range",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 abc",
                "invalid full move number",
            ),
            (
                "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 65536",
                "full move number out of range",
            ),
        ];
        for (fen, substr) in cases {
            let err = Board::from_fen(fen).unwrap_err();
            assert!(
                err.to_string().contains(substr),
                "fen {:?} error {:?} should contain {:?}",
                fen,
                err.to_string(),
                substr
            );
        }

        // valid cases
        for fen in [
            START,
            "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w Kq - 0 1",
            "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 50 1",
            "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1000",
        ] {
            assert!(Board::from_fen(fen).is_ok());
        }
    }

    #[test]
    fn parse_fen_alias() {
        assert!(Board::parse_fen(START).is_ok());
    }
}
