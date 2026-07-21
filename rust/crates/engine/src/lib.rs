//! Chess engine domain: rules, board, FEN, moves, attacks, Zobrist hashing.
//!
//! Pure domain crate with no internal dependencies. Ported from the Go
//! `internal/engine` package, preserving behavior and public semantics.

mod attacks;
mod board;
mod fen;
mod game_state;
mod moves;
mod types;
mod zobrist;

pub use board::{
    Board, MoveError, CASTLE_ALL, CASTLE_BLACK_KING, CASTLE_BLACK_QUEEN, CASTLE_WHITE_KING,
    CASTLE_WHITE_QUEEN,
};
pub use fen::FenError;
pub use game_state::GameStatus;
pub use moves::{Move, ParseMoveError};
pub use types::{Color, Piece, PieceType, Square, NO_SQUARE};
