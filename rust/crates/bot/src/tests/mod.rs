//! Ported tests from the Go `internal/bot` package. Each submodule mirrors a
//! `*_test.go` file. These are in-crate unit tests so they can exercise
//! crate-private helpers and fields, just as the Go tests (same package) did.

mod difficulty_tests;
mod engine_tests;
mod eval_tests;
mod factory_tests;
mod minimax_tests;
mod performance_tests;
mod random_tests;
mod rl_encoder_tests;
mod rl_tests;
mod tactics_tests;

use engine::Board;

/// Loads a board from a FEN string, panicking on error (test helper).
pub(crate) fn from_fen(fen: &str) -> Board {
    Board::from_fen(fen).unwrap_or_else(|e| panic!("failed to load FEN {:?}: {:?}", fen, e))
}
