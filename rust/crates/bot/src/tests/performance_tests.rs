//! Ported from `performance_test.go`: wall-clock time-limit tests.
//!
//! These assert real elapsed time and are `#[ignore]`d by default, mirroring
//! the Go tests' `testing.Short()` skip. The Go benchmarks (`Benchmark*`) are
//! not ported as `#[test]`s since they measure throughput, not correctness.

use std::time::{Duration, Instant};

use engine::Move;

use crate::context::Context;
use crate::factory::{new_minimax_engine, new_random_engine};
use crate::interfaces::{Difficulty, Engine};

use super::from_fen;

fn contains_move(moves: &[Move], m: Move) -> bool {
    moves.contains(&m)
}

const POSITIONS: [(&str, &str); 3] = [
    (
        "Complex middlegame",
        "r1bqk2r/pp1n1ppp/2pbpn2/8/2BP4/2N2N2/PPP2PPP/R1BQK2R w KQkq - 0 8",
    ),
    (
        "Tactical position",
        "r2qkb1r/ppp2ppp/2n1bn2/3pp3/4P3/3P1N2/PPP2PPP/RNBQKB1R w KQkq - 0 6",
    ),
    (
        "Open position",
        "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4",
    ),
];

fn run_time_limit(engine_kind: Difficulty, limit: Duration) {
    let bot: Box<dyn Engine> = match engine_kind {
        Difficulty::Easy => Box::new(new_random_engine(&[]).expect("create")),
        d => Box::new(new_minimax_engine(d, &[]).expect("create")),
    };

    for (_name, fen) in POSITIONS {
        let board = from_fen(fen);
        let start = Instant::now();
        let mv = bot
            .select_move(&Context::background(), &board)
            .expect("select");
        let elapsed = start.elapsed();

        assert!(
            contains_move(&board.legal_moves(), mv),
            "bot returned illegal move"
        );
        let grace = Duration::from_millis(100);
        assert!(
            elapsed <= limit + grace,
            "took {:?}, expected < {:?}",
            elapsed,
            limit
        );
    }
    bot.close().ok();
}

#[test]
#[ignore = "wall-clock timing; mirrors Go's testing.Short() skip"]
fn time_limit_easy_bot() {
    run_time_limit(Difficulty::Easy, Duration::from_secs(2));
}

#[test]
#[ignore = "wall-clock timing; mirrors Go's testing.Short() skip"]
fn time_limit_medium_bot() {
    run_time_limit(Difficulty::Medium, Duration::from_secs(4));
}

#[test]
#[ignore = "wall-clock timing; mirrors Go's testing.Short() skip"]
fn time_limit_hard_bot() {
    run_time_limit(Difficulty::Hard, Duration::from_secs(8));
}

#[test]
#[ignore = "wall-clock timing; mirrors Go's testing.Short() skip"]
fn time_limit_medium_bot_with_timeout() {
    let board = from_fen("r1bqk2r/pp1n1ppp/2pbpn2/8/2BP4/2N2N2/PPP2PPP/R1BQK2R w KQkq - 0 8");
    let bot = new_minimax_engine(Difficulty::Medium, &[]).expect("create");
    let ctx = Context::with_timeout(Duration::from_millis(100));
    let mv = bot.select_move(&ctx, &board).expect("select");
    assert!(contains_move(&board.legal_moves(), mv));
    bot.close().ok();
}

#[test]
#[ignore = "wall-clock timing; mirrors Go's testing.Short() skip"]
fn time_limit_hard_bot_with_timeout() {
    let board = from_fen("r1bqk2r/pp1n1ppp/2pbpn2/8/2BP4/2N2N2/PPP2PPP/R1BQK2R w KQkq - 0 8");
    let bot = new_minimax_engine(Difficulty::Hard, &[]).expect("create");
    let ctx = Context::with_timeout(Duration::from_millis(200));
    let mv = bot.select_move(&ctx, &board).expect("select");
    assert!(contains_move(&board.legal_moves(), mv));
    bot.close().ok();
}
