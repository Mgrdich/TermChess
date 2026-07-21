//! Ported from `minimax_test.go`: the Medium/Hard minimax engine.

use std::time::{Duration, Instant};

use engine::{Board, GameStatus, Move};

use crate::context::Context;
use crate::factory::{new_minimax_engine, with_time_limit};
use crate::interfaces::{Configurable, Difficulty, Engine, EngineType, Inspectable, MinimaxConfig};
use crate::minimax::{get_default_weights, EvalWeights};

use super::from_fen;

fn contains_move(moves: &[Move], m: Move) -> bool {
    moves.contains(&m)
}

#[test]
fn name() {
    for (difficulty, expected) in [
        (Difficulty::Medium, "Medium Bot"),
        (Difficulty::Hard, "Hard Bot"),
    ] {
        let eng = new_minimax_engine(difficulty, &[]).expect("create");
        assert_eq!(eng.name(), expected);
        eng.close().ok();
    }
}

#[test]
fn close() {
    let eng = new_minimax_engine(Difficulty::Medium, &[]).expect("create");
    eng.close().expect("close");
    let board = Board::new();
    let err = eng.select_move(&Context::background(), &board).unwrap_err();
    assert!(err.to_string().contains("closed"));
}

#[test]
fn info() {
    let medium = new_minimax_engine(Difficulty::Medium, &[]).expect("create");
    let info = medium.info();
    assert_eq!(info.name, "Medium Bot");
    assert_eq!(info.author, "TermChess");
    assert_eq!(info.version, "1.0");
    assert_eq!(info.engine_type, EngineType::Internal);
    assert_eq!(info.difficulty, Difficulty::Medium);
    assert!(info.features["alpha_beta"]);
    assert!(info.features["iterative_deepening"]);
    assert!(info.features["configurable"]);
    assert!(info.features["piece_square_tables"]);
    assert!(info.features["mobility"]);
    assert!(!info.features["king_safety"]);
    medium.close().ok();

    let hard = new_minimax_engine(Difficulty::Hard, &[]).expect("create");
    let info = hard.info();
    assert_eq!(info.name, "Hard Bot");
    assert_eq!(info.difficulty, Difficulty::Hard);
    assert!(info.features["king_safety"]);
    hard.close().ok();
}

#[test]
fn forced_move() {
    let board = from_fen("4k3/8/8/8/8/8/4r3/4K2R w - - 0 1");
    let moves = board.legal_moves();
    assert!(!moves.is_empty());

    let eng = new_minimax_engine(Difficulty::Medium, &[]).expect("create");
    if moves.len() == 1 {
        let start = Instant::now();
        let mv = eng
            .select_move(&Context::background(), &board)
            .expect("select");
        assert!(start.elapsed() < Duration::from_millis(100));
        assert_eq!(mv, moves[0]);
    } else {
        let mv = eng
            .select_move(&Context::background(), &board)
            .expect("select");
        assert!(contains_move(&moves, mv));
    }
    eng.close().ok();
}

#[test]
fn finds_mate_in_one() {
    let cases = [
        "6k1/5ppp/8/8/8/8/8/R6K w - - 0 1",
        "k7/8/1K6/8/8/8/8/Q7 w - - 0 1",
        "7k/5Q2/6K1/8/8/8/8/8 w - - 0 1",
    ];
    for fen in cases {
        let board = from_fen(fen);
        let eng = new_minimax_engine(Difficulty::Medium, &[]).expect("create");
        let ctx = Context::with_timeout(Duration::from_secs(5));
        let mv = eng.select_move(&ctx, &board).expect("select");
        let mut copy = board.copy();
        copy.make_move(mv).expect("make move");
        assert_eq!(
            copy.status(),
            GameStatus::Checkmate,
            "fen {} should be mate-in-1",
            fen
        );
        eng.close().ok();
    }
}

#[test]
fn avoid_blunder() {
    let board = from_fen("4k3/3r4/8/8/8/8/8/3Q1K2 w - - 0 1");
    let eng = new_minimax_engine(Difficulty::Medium, &[]).expect("create");
    let ctx = Context::with_timeout(Duration::from_secs(5));
    let mv = eng.select_move(&ctx, &board).expect("select");
    let blunder = Move::parse("d1d8").unwrap();
    assert_ne!(mv, blunder, "engine should not hang the queen with Qd8");
    let mut copy = board.copy();
    copy.make_move(mv).expect("make move");
    eng.close().ok();
}

#[test]
fn capture_priority() {
    let board = from_fen("6k1/8/5q2/4P3/8/8/8/6K1 w - - 0 1");
    let eng = new_minimax_engine(Difficulty::Medium, &[]).expect("create");
    let ctx = Context::with_timeout(Duration::from_secs(5));
    let mv = eng.select_move(&ctx, &board).expect("select");
    assert_eq!(
        mv,
        Move::parse("e5f6").unwrap(),
        "engine should capture the queen"
    );
    eng.close().ok();
}

#[test]
fn timeout() {
    let board = Board::new();
    let eng = new_minimax_engine(
        Difficulty::Medium,
        &[with_time_limit(Duration::from_nanos(1))],
    )
    .expect("create");
    // With a 1ns limit the engine returns the first legal move immediately.
    let mv = eng
        .select_move(&Context::background(), &board)
        .expect("select");
    assert!(contains_move(&board.legal_moves(), mv));
    eng.close().ok();
}

#[test]
fn no_legal_moves() {
    let board = from_fen("rnb1kbnr/pppp1ppp/8/4p3/6Pq/5P2/PPPPP2P/RNBQKBNR w KQkq - 1 3");
    assert_eq!(board.status(), GameStatus::Checkmate);

    let eng = new_minimax_engine(Difficulty::Medium, &[]).expect("create");
    let err = eng.select_move(&Context::background(), &board).unwrap_err();
    assert!(err.to_string().contains("no legal moves"));
    eng.close().ok();
}

#[test]
fn depth2_search_finds_fork() {
    let board = from_fen("6k1/8/3r1r2/8/4P3/8/8/6K1 w - - 0 1");
    let eng = new_minimax_engine(Difficulty::Medium, &[]).expect("create");
    let ctx = Context::with_timeout(Duration::from_secs(5));
    let mv = eng.select_move(&ctx, &board).expect("select");
    assert_eq!(
        mv,
        Move::parse("e4e5").unwrap(),
        "engine should find the pawn fork e4-e5"
    );
    eng.close().ok();
}

#[test]
fn get_default_weights_all_difficulties() {
    let expected = EvalWeights {
        material: 1.0,
        piece_square: 0.0,
        mobility: 0.0,
        king_safety: 0.0,
    };
    for difficulty in [Difficulty::Medium, Difficulty::Hard, Difficulty::Easy] {
        assert_eq!(get_default_weights(difficulty), expected);
    }
}

#[test]
fn move_ordering() {
    let board = from_fen("6k1/8/3p1p2/4P3/8/8/8/6K1 w - - 0 1");
    let eng = new_minimax_engine(Difficulty::Medium, &[]).expect("create");
    let moves = board.legal_moves();
    let ordered = eng.order_moves(&board, &moves);

    let mut capture_phase = true;
    for m in ordered {
        let is_capture = !board.piece_at(m.to).is_empty();
        if !is_capture {
            capture_phase = false;
        }
        if is_capture && !capture_phase {
            panic!("all captures should come before non-captures");
        }
    }
    eng.close().ok();
}

#[test]
fn iterative_deepening_timeout() {
    let eng = new_minimax_engine(
        Difficulty::Medium,
        &[with_time_limit(Duration::from_millis(100))],
    )
    .expect("create");
    let board = from_fen("r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 1");
    let mv = eng
        .select_move(&Context::background(), &board)
        .expect("select");
    assert!(contains_move(&board.legal_moves(), mv));
    eng.close().ok();
}

#[test]
fn iterative_deepening_multiple_depths() {
    let eng = new_minimax_engine(
        Difficulty::Medium,
        &[with_time_limit(Duration::from_secs(5))],
    )
    .expect("create");
    let board = from_fen("8/8/8/4k3/8/8/4K3/4R3 w - - 0 1");
    let mv = eng
        .select_move(&Context::background(), &board)
        .expect("select");
    assert!(contains_move(&board.legal_moves(), mv));
    eng.close().ok();
}

#[test]
fn returns_last_completed_depth() {
    let eng = new_minimax_engine(
        Difficulty::Hard,
        &[with_time_limit(Duration::from_millis(500))],
    )
    .expect("create");
    let board = from_fen("r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 1");
    let mv = eng
        .select_move(&Context::background(), &board)
        .expect("select");
    assert!(contains_move(&board.legal_moves(), mv));
    eng.close().ok();
}

#[test]
fn configure_search_depth() {
    let mut eng = new_minimax_engine(Difficulty::Medium, &[]).expect("create");
    eng.configure(MinimaxConfig {
        search_depth: Some(8),
        ..Default::default()
    })
    .expect("configure");
    assert_eq!(eng.max_depth, 8);

    assert!(eng
        .configure(MinimaxConfig {
            search_depth: Some(0),
            ..Default::default()
        })
        .is_err());
    assert!(eng
        .configure(MinimaxConfig {
            search_depth: Some(21),
            ..Default::default()
        })
        .is_err());
}

#[test]
fn configure_time_limit() {
    let mut eng = new_minimax_engine(Difficulty::Medium, &[]).expect("create");
    eng.configure(MinimaxConfig {
        time_limit: Some(Duration::from_secs(5)),
        ..Default::default()
    })
    .expect("configure");
    assert_eq!(eng.time_limit, Duration::from_secs(5));

    assert!(eng
        .configure(MinimaxConfig {
            time_limit: Some(Duration::ZERO),
            ..Default::default()
        })
        .is_err());
}

#[test]
fn configure_eval_weights() {
    let mut eng = new_minimax_engine(Difficulty::Hard, &[]).expect("create");
    eng.configure(MinimaxConfig {
        material_weight: Some(1.5),
        piece_square_weight: Some(0.2),
        mobility_weight: Some(0.15),
        king_safety_weight: Some(0.3),
        ..Default::default()
    })
    .expect("configure");
    assert_eq!(eng.eval_weights.material, 1.5);
    assert_eq!(eng.eval_weights.piece_square, 0.2);
    assert_eq!(eng.eval_weights.mobility, 0.15);
    assert_eq!(eng.eval_weights.king_safety, 0.3);
}

#[test]
fn configure_empty_config() {
    let mut eng = new_minimax_engine(Difficulty::Medium, &[]).expect("create");
    eng.configure(MinimaxConfig::default()).expect("configure");
    assert_eq!(eng.max_depth, 4);
}

#[test]
#[ignore = "timing-sensitive; mirrors Go's testing.Short() skip"]
fn alpha_beta_pruning_completes_quickly() {
    let board = Board::new();
    let eng = new_minimax_engine(Difficulty::Medium, &[]).expect("create");
    let ctx = Context::with_timeout(Duration::from_secs(5));
    let start = Instant::now();
    let mv = eng.select_move(&ctx, &board).expect("select");
    assert!(start.elapsed() < Duration::from_secs(1));
    assert!(mv.from.is_valid());
    eng.close().ok();
}
