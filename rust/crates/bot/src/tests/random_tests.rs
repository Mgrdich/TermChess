//! Ported from `random_test.go`: the Easy (random) engine.

use std::collections::HashMap;

use engine::{Board, PieceType};

use crate::context::Context;
use crate::factory::{new_random_engine, with_time_limit};
use crate::interfaces::{Difficulty, Engine, EngineType, Inspectable};
use crate::random::{filter_captures, filter_checks};

use super::from_fen;

#[test]
fn select_move_returns_legal_move() {
    let eng = new_random_engine(&[]).expect("create engine");
    let board = Board::new();

    for i in 0..100 {
        let mv = eng
            .select_move(&Context::background(), &board)
            .unwrap_or_else(|e| panic!("SelectMove failed on iteration {}: {:?}", i, e));

        let legal = board.legal_moves();
        assert!(
            legal
                .iter()
                .any(|lm| lm.from == mv.from && lm.to == mv.to && lm.promotion == mv.promotion),
            "Move {} should be in legal moves list",
            mv
        );
    }
    eng.close().expect("close");
}

#[test]
fn select_move_no_legal_moves() {
    let eng = new_random_engine(&[]).expect("create engine");
    // Checkmate: black king h8, white queen f7, white king g6.
    let board = from_fen("7k/5Q2/6K1/8/8/8/8/8 b - - 0 1");

    assert_eq!(
        board.legal_moves().len(),
        0,
        "position should have no legal moves"
    );

    let err = eng.select_move(&Context::background(), &board).unwrap_err();
    assert_eq!(err.to_string(), "no legal moves available");
    eng.close().ok();
}

#[test]
fn select_move_forced_move() {
    let eng = new_random_engine(&[]).expect("create engine");
    let board = from_fen("7k/8/6K1/8/8/8/8/7R b - - 0 1");

    let legal = board.legal_moves();
    assert_eq!(
        legal.len(),
        1,
        "position should have exactly one legal move"
    );

    let expected = legal[0];
    for _ in 0..10 {
        let mv = eng
            .select_move(&Context::background(), &board)
            .expect("select");
        assert_eq!(
            (mv.from, mv.to, mv.promotion),
            (expected.from, expected.to, expected.promotion)
        );
    }
    eng.close().ok();
}

#[test]
fn select_move_timeout() {
    let eng = new_random_engine(&[]).expect("create engine");
    // Already-cancelled context.
    let ctx = Context::background();
    ctx.cancel();
    let board = Board::new();

    assert!(
        eng.select_move(&ctx, &board).is_err(),
        "expected error for cancelled context"
    );
    eng.close().ok();
}

#[test]
fn select_move_when_closed() {
    let eng = new_random_engine(&[]).expect("create engine");
    eng.close().expect("close");

    let board = Board::new();
    let err = eng.select_move(&Context::background(), &board).unwrap_err();
    assert_eq!(err.to_string(), "engine is closed");
}

#[test]
fn name() {
    let eng = new_random_engine(&[]).expect("create engine");
    assert_eq!(eng.name(), "Easy Bot");
    eng.close().ok();
}

#[test]
fn close_is_idempotent() {
    let eng = new_random_engine(&[]).expect("create engine");
    assert!(eng.close().is_ok());
    assert!(eng.close().is_ok());
    assert!(eng.close().is_ok());
}

#[test]
fn info() {
    let eng = new_random_engine(&[]).expect("create engine");
    let info = eng.info();

    assert_eq!(info.name, "Easy Bot");
    assert_eq!(info.author, "TermChess");
    assert_eq!(info.version, "1.0");
    assert_eq!(info.engine_type, EngineType::Internal);
    assert_eq!(info.difficulty, Difficulty::Easy);
    assert!(info.features["random_selection"]);
    assert!(info.features["tactical_awareness"]);
    assert!(info.features["weighted_selection"]);
    eng.close().ok();
}

#[test]
fn new_random_engine_default_config() {
    let eng = new_random_engine(&[]).expect("create engine");
    assert_eq!(eng.time_limit, std::time::Duration::from_secs(2));
    eng.close().ok();
}

#[test]
fn new_random_engine_custom_time_limit() {
    let eng =
        new_random_engine(&[with_time_limit(std::time::Duration::from_secs(5))]).expect("create");
    assert_eq!(eng.time_limit, std::time::Duration::from_secs(5));
    eng.close().ok();
}

#[test]
fn select_move_distribution_across_moves() {
    let eng = new_random_engine(&[]).expect("create engine");
    let board = Board::new();

    let legal = board.legal_moves();
    assert!(legal.len() > 1);

    let mut move_counts: HashMap<String, i32> = HashMap::new();
    for _ in 0..1000 {
        let mv = eng
            .select_move(&Context::background(), &board)
            .expect("select");
        *move_counts.entry(mv.to_string()).or_insert(0) += 1;
    }

    let min_expected = legal.len() / 2;
    assert!(
        move_counts.len() >= min_expected,
        "should select at least {} different moves, got {}",
        min_expected,
        move_counts.len()
    );

    for move_str in move_counts.keys() {
        assert!(
            legal.iter().any(|lm| lm.to_string() == *move_str),
            "Move {} should be legal",
            move_str
        );
    }
    eng.close().ok();
}

#[test]
fn select_move_various_positions() {
    let eng = new_random_engine(&[]).expect("create engine");

    let cases: [(&str, &str, bool); 4] = [
        (
            "Starting position",
            "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
            false,
        ),
        (
            "After e4",
            "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
            false,
        ),
        ("Endgame position", "8/5k2/8/8/8/8/3K4/8 w - - 0 1", false),
        (
            "Checkmate - no legal moves",
            "7k/5Q2/6K1/8/8/8/8/8 b - - 0 1",
            true,
        ),
    ];

    for (name, fen, expect_error) in cases {
        let board = from_fen(fen);
        let res = eng.select_move(&Context::background(), &board);
        if expect_error {
            assert!(res.is_err(), "expected error for {}", name);
        } else {
            let mv = res.unwrap_or_else(|e| panic!("SelectMove failed for {}: {:?}", name, e));
            let legal = board.legal_moves();
            assert!(
                legal
                    .iter()
                    .any(|lm| lm.from == mv.from && lm.to == mv.to && lm.promotion == mv.promotion),
                "Move {} should be legal in position {}",
                mv,
                name
            );
        }
    }
    eng.close().ok();
}

#[test]
fn select_move_captures_bias() {
    // Position after 1.e4 e5 2.Nf3 Nc6 3.Bc4.
    let board = from_fen("r1bqkbnr/pppp1ppp/2n5/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 2 3");
    let eng = new_random_engine(&[]).expect("create engine");

    let trials = 1000;
    let mut capture_count = 0;
    for _ in 0..trials {
        let mv = eng
            .select_move(&Context::background(), &board)
            .expect("select");
        if !board.piece_at(mv.to).is_empty() {
            capture_count += 1;
        }
    }

    // Probabilistic: Go accepts 70-95% but only warns outside. We assert a
    // generous lower bound so the test is not flaky while still checking bias.
    let pct = capture_count as f64 / trials as f64 * 100.0;
    assert!(
        pct >= 55.0,
        "capture bias {:.1}% unexpectedly low (bias should favor captures)",
        pct
    );
    eng.close().ok();
}

#[test]
fn select_move_checks_bias() {
    let board = from_fen("r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/3P1N2/PPP2PPP/RNBQK2R w KQkq - 0 5");
    let eng = new_random_engine(&[]).expect("create engine");

    let trials = 1000;
    let mut check_count = 0;
    for _ in 0..trials {
        let mv = eng
            .select_move(&Context::background(), &board)
            .expect("select");
        let mut copy = board.copy();
        let _ = copy.make_move(mv);
        if copy.in_check() {
            check_count += 1;
        }
    }

    // Bias exists but is probabilistic; assert a loose lower bound only.
    let pct = check_count as f64 / trials as f64 * 100.0;
    assert!(pct >= 20.0, "check bias {:.1}% unexpectedly low", pct);
    eng.close().ok();
}

#[test]
fn select_move_random_fallback() {
    let board = Board::new();
    let eng = new_random_engine(&[]).expect("create engine");

    let mut moves_seen: HashMap<String, i32> = HashMap::new();
    for _ in 0..50 {
        let mv = eng
            .select_move(&Context::background(), &board)
            .expect("select");
        *moves_seen.entry(mv.to_string()).or_insert(0) += 1;
    }
    assert!(
        moves_seen.len() >= 10,
        "only saw {} unique moves, want more variety",
        moves_seen.len()
    );
    eng.close().ok();
}

#[test]
fn test_filter_captures() {
    let board = from_fen("r1bqkbnr/pppp1ppp/2n5/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 2 3");
    let moves = board.legal_moves();
    let captures = filter_captures(&board, &moves);

    assert!(
        !captures.is_empty(),
        "filterCaptures found 0 captures, expected at least 1"
    );
    for m in captures {
        assert!(
            !board.piece_at(m.to).is_empty() || is_en_passant(&board, m),
            "filterCaptures returned non-capture move: {}",
            m
        );
    }
}

fn is_en_passant(board: &Board, m: engine::Move) -> bool {
    board.piece_at(m.from).piece_type() == PieceType::Pawn
        && board.en_passant_sq >= 0
        && m.to == engine::Square(board.en_passant_sq)
}

#[test]
fn test_filter_checks() {
    let board = from_fen("r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/3P1N2/PPP2PPP/RNBQK2R w KQkq - 0 5");
    let moves = board.legal_moves();
    let checks = filter_checks(&board, &moves);

    assert!(
        !checks.is_empty(),
        "filterChecks found 0 checks, expected at least 1"
    );
    for m in checks {
        let mut copy = board.copy();
        let _ = copy.make_move(m);
        assert!(
            copy.in_check(),
            "filterChecks returned non-check move: {}",
            m
        );
    }
}
