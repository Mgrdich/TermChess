//! Ported from `difficulty_test.go`: bot-vs-bot game outcomes.
//!
//! These play full games and are slow, so they are `#[ignore]`d by default,
//! mirroring the Go tests' `testing.Short()` skip. Run with
//! `cargo test -p bot -- --ignored` to exercise them.

use std::time::Duration;

use engine::{Board, Color, GameStatus};

use crate::context::Context;
use crate::factory::{new_minimax_engine, new_random_engine, with_search_depth, with_time_limit};
use crate::interfaces::{Difficulty, Engine};

struct GameResult {
    winner: Option<Color>,
    is_draw: bool,
    move_count: i32,
}

fn run_bot_game(white: &dyn Engine, black: &dyn Engine) -> GameResult {
    let mut board = Board::new();
    let mut move_count = 0;
    let max_moves = 200;

    while move_count < max_moves {
        if board.is_game_over() {
            let winner = board.winner();
            return GameResult {
                winner,
                is_draw: winner.is_none(),
                move_count,
            };
        }
        if board.can_claim_draw() {
            return GameResult {
                winner: None,
                is_draw: true,
                move_count,
            };
        }

        let current: &dyn Engine = if board.active_color == Color::White {
            white
        } else {
            black
        };
        let ctx = Context::with_timeout(Duration::from_secs(10));
        let mv = current
            .select_move(&ctx, &board)
            .expect("bot failed to select move");
        board.make_move(mv).expect("bot selected illegal move");
        move_count += 1;
    }

    GameResult {
        winner: None,
        is_draw: true,
        move_count,
    }
}

#[test]
#[ignore = "slow bot-vs-bot game; mirrors Go's testing.Short() skip"]
fn medium_vs_easy() {
    let easy = new_random_engine(&[]).expect("create");
    let medium = new_minimax_engine(Difficulty::Medium, &[]).expect("create");

    let mut medium_wins = 0;
    for i in 0..10 {
        let result = if i % 2 == 0 {
            let r = run_bot_game(&medium, &easy);
            if !r.is_draw && r.winner == Some(Color::White) {
                medium_wins += 1;
            }
            r
        } else {
            let r = run_bot_game(&easy, &medium);
            if !r.is_draw && r.winner == Some(Color::Black) {
                medium_wins += 1;
            }
            r
        };
        assert!(result.move_count > 0);
    }
    // Go treats a low win-count as a warning, not a failure; we only assert the
    // games completed. `medium_wins` is observed for parity with the Go log.
    let _ = medium_wins;
    easy.close().ok();
    medium.close().ok();
}

#[test]
#[ignore = "slow bot-vs-bot game; mirrors Go's testing.Short() skip"]
fn hard_vs_medium() {
    let medium = new_minimax_engine(
        Difficulty::Medium,
        &[
            with_time_limit(Duration::from_millis(500)),
            with_search_depth(2),
        ],
    )
    .expect("create");
    let hard = new_minimax_engine(
        Difficulty::Hard,
        &[
            with_time_limit(Duration::from_secs(1)),
            with_search_depth(4),
        ],
    )
    .expect("create");

    for i in 0..3 {
        let result = if i % 2 == 0 {
            run_bot_game(&hard, &medium)
        } else {
            run_bot_game(&medium, &hard)
        };
        assert!(result.move_count > 0);
    }
    medium.close().ok();
    hard.close().ok();
}

#[test]
#[ignore = "slow bot-vs-bot game; mirrors Go's testing.Short() skip"]
fn easy_vs_easy() {
    let easy1 = new_random_engine(&[]).expect("create");
    let easy2 = new_random_engine(&[]).expect("create");

    for _ in 0..5 {
        let result = run_bot_game(&easy1, &easy2);
        // Sanity: a completed game either ended or hit the move cap.
        assert!(result.move_count >= 0);
        let _ = result.winner;
        assert!(result.move_count <= 200);
        assert!(result.is_draw || result.move_count > 0);
        let _ = GameStatus::Ongoing; // status enum referenced for parity
    }
    easy1.close().ok();
    easy2.close().ok();
}
