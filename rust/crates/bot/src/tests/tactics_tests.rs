//! Ported from `tactics_test.go`: tactical puzzle solving by the minimax bots.

use std::time::Duration;

use engine::{GameStatus, Move};

use crate::context::Context;
use crate::factory::{new_minimax_engine, with_deterministic, with_search_depth};
use crate::interfaces::{Difficulty, Engine};
use crate::minimax::MinimaxEngine;

use super::from_fen;

fn make_puzzle_bot(difficulty: Difficulty) -> MinimaxEngine {
    match difficulty {
        Difficulty::Medium => new_minimax_engine(
            Difficulty::Medium,
            &[with_search_depth(5), with_deterministic(true)],
        ),
        Difficulty::Hard => new_minimax_engine(
            Difficulty::Hard,
            &[with_search_depth(6), with_deterministic(true)],
        ),
        _ => panic!("invalid difficulty"),
    }
    .expect("create bot")
}

fn make_mate_bot(difficulty: Difficulty) -> MinimaxEngine {
    new_minimax_engine(difficulty, &[with_deterministic(true)]).expect("create bot")
}

fn test_tactical_puzzle(difficulty: Difficulty, fen: &str, expected: &[&str], desc: &str) {
    let board = from_fen(fen);
    let bot = make_puzzle_bot(difficulty);
    let ctx = Context::with_timeout(Duration::from_secs(10));
    let mv = bot.select_move(&ctx, &board).expect("select");
    let move_str = mv.to_string();
    assert!(
        expected.contains(&move_str.as_str()),
        "{} bot should find {:?}: {} but found {}",
        difficulty,
        expected,
        desc,
        move_str
    );
    bot.close().ok();
}

fn test_mate_delivery(difficulty: Difficulty, fen: &str, desc: &str) {
    let board = from_fen(fen);
    let bot = make_mate_bot(difficulty);
    let ctx = Context::with_timeout(Duration::from_secs(10));
    let mv = bot.select_move(&ctx, &board).expect("select");
    let mut copy = board.copy();
    copy.make_move(mv).expect("make move");
    assert_eq!(
        copy.status(),
        GameStatus::Checkmate,
        "{} bot should find mate-in-1: {}, got move {} status {:?}",
        difficulty,
        desc,
        mv,
        copy.status()
    );
    bot.close().ok();
}

#[test]
fn mate_in_one() {
    let cases = [
        ("6k1/5ppp/8/8/8/8/8/R6K w - - 0 1", "back rank mate"),
        ("7k/8/5Q2/8/8/8/1B6/7K w - - 0 1", "queen and bishop mate"),
        ("7k/R7/8/8/8/8/8/1R5K w - - 0 1", "two rooks mate"),
        ("7k/8/5NQ1/8/8/8/8/7K w - - 0 1", "knight and queen mate"),
        ("6rk/5Npp/8/8/8/8/8/7K w - - 0 1", "smothered mate"),
    ];
    for (fen, desc) in cases {
        test_mate_delivery(Difficulty::Medium, fen, desc);
        test_mate_delivery(Difficulty::Hard, fen, desc);
    }
}

#[test]
fn mate_in_two() {
    let cases: [(&str, &[&str], &str); 3] = [
        (
            "6rk/8/5N2/8/3Q4/8/8/7K w - - 0 1",
            &["d4h4", "d4h8", "f6g8"],
            "queen sac",
        ),
        (
            "7k/R4ppp/8/8/8/8/8/7K w - - 0 1",
            &["a7a8"],
            "rook penetration",
        ),
        (
            "7k/6pp/8/8/3Q4/8/1B6/7K w - - 0 1",
            &["d4g7", "d4h8", "d4h4"],
            "queen and bishop",
        ),
    ];
    for (fen, moves, desc) in cases {
        test_tactical_puzzle(Difficulty::Medium, fen, moves, desc);
        test_tactical_puzzle(Difficulty::Hard, fen, moves, desc);
    }
}

#[test]
fn fork() {
    let cases: [(&str, &[&str], &str); 2] = [
        (
            "8/8/8/3k1r2/8/4N3/8/7K w - - 0 1",
            &["e3d5", "e3f5", "e3g4"],
            "knight fork",
        ),
        (
            "6k1/8/8/3r1r2/4P3/8/8/6K1 w - - 0 1",
            &["e4e5", "e4d5", "e4f5"],
            "pawn fork",
        ),
    ];
    for (fen, moves, desc) in cases {
        test_tactical_puzzle(Difficulty::Medium, fen, moves, desc);
        test_tactical_puzzle(Difficulty::Hard, fen, moves, desc);
    }
}

#[test]
fn pin() {
    let cases: [(&str, &[&str], &str); 2] = [
        (
            "3k4/8/8/3n4/8/8/8/3R2K1 w - - 0 1",
            &["d1d5"],
            "capture pinned piece",
        ),
        (
            "7k/8/8/4q3/8/8/1B6/7K w - - 0 1",
            &["b2e5"],
            "exploit diagonal pin",
        ),
    ];
    for (fen, moves, desc) in cases {
        test_tactical_puzzle(Difficulty::Medium, fen, moves, desc);
        test_tactical_puzzle(Difficulty::Hard, fen, moves, desc);
    }
}

#[test]
fn skewer() {
    let cases: [(&str, &[&str], &str); 2] = [
        (
            "k7/q7/8/8/8/8/8/R6K w - - 0 1",
            &["a1a7", "a1a8"],
            "rook skewer",
        ),
        (
            "7k/5q2/8/8/2B5/8/8/7K w - - 0 1",
            &["c4f7"],
            "bishop captures queen",
        ),
    ];
    for (fen, moves, desc) in cases {
        test_tactical_puzzle(Difficulty::Medium, fen, moves, desc);
        test_tactical_puzzle(Difficulty::Hard, fen, moves, desc);
    }
}

#[test]
fn discovered_attack() {
    let cases: [(&str, &[&str], &str); 2] = [
        (
            "4k3/3q4/8/4N3/8/8/8/7K w - - 0 1",
            &["e5d7"],
            "knight wins queen with check",
        ),
        (
            "4q3/8/8/8/4N3/8/8/4R2K w - - 0 1",
            &[
                "e4d6", "e4f6", "e4g5", "e4g3", "e4f2", "e4d2", "e4c3", "e4c5",
            ],
            "discovered attack",
        ),
    ];
    for (fen, moves, desc) in cases {
        test_tactical_puzzle(Difficulty::Medium, fen, moves, desc);
        test_tactical_puzzle(Difficulty::Hard, fen, moves, desc);
    }
}

#[test]
fn dont_hang_queen() {
    let board = from_fen("4k3/3r4/8/8/8/8/8/3Q1K2 w - - 0 1");
    let blunder = Move::parse("d1d8").unwrap();
    for difficulty in [Difficulty::Medium, Difficulty::Hard] {
        let bot = make_mate_bot(difficulty);
        let ctx = Context::with_timeout(Duration::from_secs(10));
        let mv = bot.select_move(&ctx, &board).expect("select");
        assert_ne!(
            mv, blunder,
            "{} bot should not hang queen with Qd8",
            difficulty
        );
        bot.close().ok();
    }
}

#[test]
fn dont_hang_rook() {
    let board = from_fen("7q/8/8/8/8/8/8/R6K w - - 0 1");
    let blunder = Move::parse("a1a8").unwrap();
    for difficulty in [Difficulty::Medium, Difficulty::Hard] {
        let bot = make_mate_bot(difficulty);
        let ctx = Context::with_timeout(Duration::from_secs(10));
        let mv = bot.select_move(&ctx, &board).expect("select");
        assert_ne!(
            mv, blunder,
            "{} bot should not hang rook with Ra8",
            difficulty
        );
        bot.close().ok();
    }
}

#[test]
fn dont_allow_back_rank_mate() {
    let board = from_fen("4r3/8/8/8/8/8/5PPP/R5K1 w - - 0 1");
    for difficulty in [Difficulty::Medium, Difficulty::Hard] {
        let bot = make_mate_bot(difficulty);
        let ctx = Context::with_timeout(Duration::from_secs(10));
        let mv = bot.select_move(&ctx, &board).expect("select");

        let mut copy = board.copy();
        copy.make_move(mv).expect("make move");

        for black_move in copy.legal_moves() {
            let mut test_board = copy.copy();
            if test_board.make_move(black_move).is_err() {
                continue;
            }
            assert_ne!(
                test_board.status(),
                GameStatus::Checkmate,
                "{} bot should prevent back rank mate; after {} Black mates with {}",
                difficulty,
                mv,
                black_move
            );
        }
        bot.close().ok();
    }
}
