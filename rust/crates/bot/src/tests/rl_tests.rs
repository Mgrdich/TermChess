//! Ported from `rl_test.go`: the RL engine skeleton, policy indexing and
//! legal-move masking.

use std::cell::Cell;
use std::rc::Rc;

use engine::{Board, Move, PieceType, Square};

use crate::context::Context;
use crate::factory::with_time_limit;
use crate::interfaces::{Engine, EngineType, Inspectable};
use crate::rl::{
    move_to_policy_index, new_rl_engine, select_best_move, InferenceSession, RLDifficulty,
};
use crate::rl_encoder::POLICY_SIZE;

/// Test double for [`InferenceSession`]. The `closed` flag is shared via `Rc`
/// so the test can observe it after the session is moved into the engine.
struct MockInferenceSession {
    policy: Vec<f32>,
    value: f32,
    err: Option<String>,
    closed: Rc<Cell<bool>>,
}

impl InferenceSession for MockInferenceSession {
    fn run_inference(&self, _input: &[f32]) -> Result<(Vec<f32>, f32), String> {
        if let Some(e) = &self.err {
            return Err(e.clone());
        }
        Ok((self.policy.clone(), self.value))
    }

    fn close(&self) -> Result<(), String> {
        self.closed.set(true);
        Ok(())
    }
}

#[test]
fn rl_engine_implements_traits() {
    let eng = new_rl_engine(RLDifficulty::Intermediate, &[]).expect("create");
    let _: &dyn Engine = &eng;
    let _: &dyn Inspectable = &eng;
    eng.close().ok();
}

#[test]
fn new_rl_engine_defaults() {
    let cases = [
        (RLDifficulty::Beginner, "RL Beginner (1000)"),
        (RLDifficulty::Intermediate, "RL Intermediate (1200)"),
        (RLDifficulty::Club, "RL Club (1500)"),
        (RLDifficulty::Advanced, "RL Advanced (2000)"),
        (RLDifficulty::Master, "RL Master (2200)"),
    ];
    for (difficulty, want_name) in cases {
        let eng = new_rl_engine(difficulty, &[]).expect("create");
        assert_eq!(eng.name(), want_name);
        eng.close().ok();
    }
}

#[test]
fn new_rl_engine_with_time_limit() {
    let eng = new_rl_engine(
        RLDifficulty::Intermediate,
        &[with_time_limit(std::time::Duration::from_secs(15))],
    )
    .expect("create");
    assert_eq!(eng.time_limit, std::time::Duration::from_secs(15));
    eng.close().ok();
}

#[test]
fn select_move_returns_model_not_loaded() {
    let eng = new_rl_engine(RLDifficulty::Intermediate, &[]).expect("create");
    let board = Board::new();
    let err = eng.select_move(&Context::background(), &board).unwrap_err();
    assert_eq!(err, crate::error::EngineError::ModelNotLoaded);
    eng.close().ok();
}

#[test]
fn close_twice_errors() {
    let eng = new_rl_engine(RLDifficulty::Advanced, &[]).expect("create");
    assert!(eng.close().is_ok());
    assert!(eng.close().is_err(), "second close should error");
}

#[test]
fn info() {
    let eng = new_rl_engine(RLDifficulty::Master, &[]).expect("create");
    let info = eng.info();
    assert_eq!(info.engine_type, EngineType::Rl);
    assert_eq!(info.name, "RL Master (2200)");
    assert!(!info.features["onnx"]);
    eng.close().ok();
}

#[test]
fn select_move_after_close() {
    let eng = new_rl_engine(RLDifficulty::Intermediate, &[]).expect("create");
    eng.close().expect("close");
    let board = Board::new();
    let err = eng.select_move(&Context::background(), &board).unwrap_err();
    assert_eq!(err.to_string(), "engine is closed");
}

#[test]
fn test_move_to_policy_index() {
    let cases = [
        (
            Move::new(Square::new(4, 1), Square::new(4, 3)),
            12 * 64 + 28,
        ), // e2e4
        (Move::new(Square::new(0, 0), Square::new(0, 1)), 8), // a1a2
        (
            Move::new(Square::new(7, 7), Square::new(7, 6)),
            63 * 64 + 55,
        ), // h8h7
    ];
    for (mv, want) in cases {
        assert_eq!(move_to_policy_index(mv), want);
    }
}

#[test]
fn select_best_move_picks_highest_score() {
    let move_a = Move::new(Square::new(4, 1), Square::new(4, 3)); // e2e4
    let move_b = Move::new(Square::new(3, 1), Square::new(3, 3)); // d2d4
    let move_c = Move::new(Square::new(6, 0), Square::new(5, 2)); // g1f3
    let legal = vec![move_a, move_b, move_c];

    let mut policy = vec![0.0f32; POLICY_SIZE];
    policy[move_to_policy_index(move_a)] = 1.0;
    policy[move_to_policy_index(move_b)] = 5.0;
    policy[move_to_policy_index(move_c)] = 2.0;

    let got = select_best_move(&legal, &policy).expect("select");
    assert_eq!((got.from, got.to), (move_b.from, move_b.to));
}

#[test]
fn select_best_move_legal_masking() {
    let legal_a = Move::new(Square::new(4, 1), Square::new(4, 3));
    let legal_b = Move::new(Square::new(3, 1), Square::new(3, 3));
    let legal = vec![legal_a, legal_b];

    let mut policy = vec![0.0f32; POLICY_SIZE];
    policy[0] = 100.0;
    policy[100] = 50.0;
    policy[2000] = 80.0;
    policy[move_to_policy_index(legal_a)] = 3.0;
    policy[move_to_policy_index(legal_b)] = 7.0;

    let got = select_best_move(&legal, &policy).expect("select");
    assert_eq!((got.from, got.to), (legal_b.from, legal_b.to));
}

#[test]
fn select_best_move_queen_promotion_preference() {
    let from = Square::new(0, 6); // a7
    let to = Square::new(0, 7); // a8
    let idx = move_to_policy_index(Move::new(from, to));

    let promo_moves = vec![
        Move::with_promotion(from, to, PieceType::Rook),
        Move::with_promotion(from, to, PieceType::Knight),
        Move::with_promotion(from, to, PieceType::Bishop),
        Move::with_promotion(from, to, PieceType::Queen),
    ];

    let mut policy = vec![0.0f32; POLICY_SIZE];
    policy[idx] = 5.0;

    let got = select_best_move(&promo_moves, &policy).expect("select");
    assert_eq!(got.promotion, PieceType::Queen);
}

#[test]
fn select_move_with_mock_session() {
    let mut eng = new_rl_engine(RLDifficulty::Intermediate, &[]).expect("create");

    let mut policy = vec![0.0f32; POLICY_SIZE];
    policy[796] = 10.0; // e2e4

    eng.session = Some(Box::new(MockInferenceSession {
        policy,
        value: 0.5,
        err: None,
        closed: Rc::new(Cell::new(false)),
    }));

    let board = Board::new();
    let mv = eng
        .select_move(&Context::background(), &board)
        .expect("select");
    assert_eq!((mv.from, mv.to), (Square::new(4, 1), Square::new(4, 3)));
    eng.close().ok();
}

#[test]
fn select_move_single_legal_move_skips_inference() {
    let mut eng = new_rl_engine(RLDifficulty::Intermediate, &[]).expect("create");

    // Position with exactly one legal move for White.
    let board = Board::from_fen("K1k5/8/8/8/8/8/8/1r6 w - - 0 1").expect("fen");
    let legal = board.legal_moves();
    assert_eq!(
        legal.len(),
        1,
        "expected exactly one legal move, got {}",
        legal.len()
    );

    // A mock that errors if called; single-move path must not invoke inference.
    eng.session = Some(Box::new(MockInferenceSession {
        policy: Vec::new(),
        value: 0.0,
        err: Some("should not be called".to_string()),
        closed: Rc::new(Cell::new(false)),
    }));

    let mv = eng
        .select_move(&Context::background(), &board)
        .expect("select");
    assert_eq!((mv.from, mv.to), (legal[0].from, legal[0].to));
    eng.close().ok();
}

#[test]
fn select_move_inference_error() {
    let mut eng = new_rl_engine(RLDifficulty::Intermediate, &[]).expect("create");
    eng.session = Some(Box::new(MockInferenceSession {
        policy: Vec::new(),
        value: 0.0,
        err: Some("onnx runtime error".to_string()),
        closed: Rc::new(Cell::new(false)),
    }));

    let board = Board::new();
    let err = eng.select_move(&Context::background(), &board).unwrap_err();
    assert_eq!(err.to_string(), "inference failed: onnx runtime error");
    eng.close().ok();
}

#[test]
fn select_move_closed_session() {
    let mut eng = new_rl_engine(RLDifficulty::Intermediate, &[]).expect("create");
    let closed = Rc::new(Cell::new(false));
    eng.session = Some(Box::new(MockInferenceSession {
        policy: vec![0.0f32; POLICY_SIZE],
        value: 0.0,
        err: None,
        closed: closed.clone(),
    }));

    eng.close().expect("close");
    assert!(closed.get(), "expected mock session to be closed");

    let board = Board::new();
    let err = eng.select_move(&Context::background(), &board).unwrap_err();
    assert_eq!(err.to_string(), "engine is closed");
}
