//! Core value types for the chess engine: `Color`, `PieceType`, `Piece`, `Square`.

use std::fmt;

/// The color of a chess piece (White or Black).
#[derive(Clone, Copy, PartialEq, Eq, Debug, Hash)]
pub enum Color {
    /// The white player (encoded value 0).
    White,
    /// The black player (encoded value 1).
    Black,
}

impl Color {
    /// Returns the numeric encoding of the color (White = 0, Black = 1).
    #[inline]
    pub fn as_u8(self) -> u8 {
        match self {
            Color::White => 0,
            Color::Black => 1,
        }
    }

    /// Returns the opposing color.
    #[inline]
    pub fn opponent(self) -> Color {
        match self {
            Color::White => Color::Black,
            Color::Black => Color::White,
        }
    }

    /// Constructs a color from its numeric encoding (0 = White, else Black).
    #[inline]
    pub fn from_u8(v: u8) -> Color {
        if v == 0 {
            Color::White
        } else {
            Color::Black
        }
    }
}

/// The type of a chess piece.
#[derive(Clone, Copy, PartialEq, Eq, Debug, Hash)]
pub enum PieceType {
    /// An empty square (no piece).
    Empty = 0,
    Pawn = 1,
    Knight = 2,
    Bishop = 3,
    Rook = 4,
    Queen = 5,
    King = 6,
}

impl PieceType {
    /// Returns the numeric encoding of the piece type (0-6).
    #[inline]
    pub fn as_u8(self) -> u8 {
        self as u8
    }

    /// Constructs a piece type from its numeric encoding (0-6).
    /// Values outside 0-6 map to `Empty`, mirroring the low-3-bit masking in Go.
    #[inline]
    pub fn from_u8(v: u8) -> PieceType {
        match v {
            1 => PieceType::Pawn,
            2 => PieceType::Knight,
            3 => PieceType::Bishop,
            4 => PieceType::Rook,
            5 => PieceType::Queen,
            6 => PieceType::King,
            _ => PieceType::Empty,
        }
    }
}

/// A chess piece encoded as a single byte.
///
/// The high bit stores the color (0 = White, 1 = Black). The low 3 bits store the
/// piece type.
#[derive(Clone, Copy, PartialEq, Eq, Debug, Hash)]
pub struct Piece(pub u8);

impl Piece {
    /// The empty piece (no piece on the square).
    pub const EMPTY: Piece = Piece(0);

    /// Creates a new piece with the given color and type.
    #[inline]
    pub fn new(color: Color, piece_type: PieceType) -> Piece {
        Piece((color.as_u8() << 7) | piece_type.as_u8())
    }

    /// Returns the color of the piece.
    #[inline]
    pub fn color(self) -> Color {
        Color::from_u8(self.0 >> 7)
    }

    /// Returns the type of the piece.
    #[inline]
    pub fn piece_type(self) -> PieceType {
        PieceType::from_u8(self.0 & 0x07)
    }

    /// Returns true if this is an empty square (no piece).
    #[inline]
    pub fn is_empty(self) -> bool {
        self.piece_type() == PieceType::Empty
    }
}

/// A square on the chess board (0-63), or -1 for no square.
///
/// Indexed as `rank * 8 + file`, where a1 = 0 and h8 = 63.
#[derive(Clone, Copy, PartialEq, Eq, Debug, Hash, PartialOrd, Ord)]
pub struct Square(pub i8);

/// Sentinel value representing an invalid or non-existent square.
pub const NO_SQUARE: Square = Square(-1);

impl Square {
    /// Creates a square from file and rank (both 0-7). Returns [`NO_SQUARE`] if out of range.
    #[inline]
    pub fn new(file: i32, rank: i32) -> Square {
        if !(0..=7).contains(&file) || !(0..=7).contains(&rank) {
            return NO_SQUARE;
        }
        Square((rank * 8 + file) as i8)
    }

    /// Returns the file of the square (0 = a, 7 = h).
    #[inline]
    pub fn file(self) -> i32 {
        (self.0 as i32) % 8
    }

    /// Returns the rank of the square (0 = 1, 7 = 8).
    #[inline]
    pub fn rank(self) -> i32 {
        (self.0 as i32) / 8
    }

    /// Returns true if this is a valid board square (0-63).
    #[inline]
    pub fn is_valid(self) -> bool {
        self.0 >= 0 && self.0 <= 63
    }

    /// Returns the square index as a `usize` for array indexing. Only valid when
    /// [`Square::is_valid`] is true.
    #[inline]
    pub fn index(self) -> usize {
        self.0 as usize
    }
}

impl fmt::Display for Square {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        if !self.is_valid() {
            return write!(f, "-");
        }
        let file = (b'a' as i32 + self.file()) as u8 as char;
        let rank = (b'1' as i32 + self.rank()) as u8 as char;
        write!(f, "{}{}", file, rank)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn square_indexing() {
        let cases = [
            ("a1", 0, 0, 0i8),
            ("b1", 1, 0, 1),
            ("h1", 7, 0, 7),
            ("a2", 0, 1, 8),
            ("a8", 0, 7, 56),
            ("h8", 7, 7, 63),
            ("e4", 4, 3, 28),
            ("d5", 3, 4, 35),
        ];
        for (name, file, rank, expected) in cases {
            let sq = Square::new(file, rank);
            assert_eq!(sq, Square(expected));
            assert_eq!(sq.to_string(), name);
            assert_eq!(sq.file(), file);
            assert_eq!(sq.rank(), rank);
        }
    }

    #[test]
    fn square_validity() {
        for i in 0..=63i8 {
            assert!(Square(i).is_valid());
        }
        for i in [-1i8, -10, 64, 100] {
            assert!(!Square(i).is_valid());
        }
    }

    #[test]
    fn new_square_invalid_inputs() {
        let cases = [(-1, 0), (0, -1), (8, 0), (0, 8), (-1, -1), (8, 8)];
        for (file, rank) in cases {
            assert_eq!(Square::new(file, rank), NO_SQUARE);
        }
    }

    #[test]
    fn invalid_square_display() {
        assert_eq!(NO_SQUARE.to_string(), "-");
        assert_eq!(Square(64).to_string(), "-");
    }

    #[test]
    fn piece_creation() {
        let types = [
            PieceType::Pawn,
            PieceType::Knight,
            PieceType::Bishop,
            PieceType::Rook,
            PieceType::Queen,
            PieceType::King,
        ];
        for color in [Color::White, Color::Black] {
            for &pt in &types {
                let piece = Piece::new(color, pt);
                assert_eq!(piece.color(), color);
                assert_eq!(piece.piece_type(), pt);
                assert!(!piece.is_empty());
            }
        }
    }

    #[test]
    fn empty_piece() {
        let piece = Piece(0);
        assert!(piece.is_empty());
        assert_eq!(piece.piece_type(), PieceType::Empty);
    }
}
