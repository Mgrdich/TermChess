//! RL-based engine skeleton (ported from `rl.go`).
//!
//! When no inference session is loaded, `select_move` returns
//! [`EngineError::ModelNotLoaded`]. The encoder, policy selector, and
//! mock-session plumbing are wired so the runtime path can be tested.

use std::sync::atomic::{AtomicBool, Ordering};
use std::time::Duration;

use engine::{Board, Move, PieceType};

use crate::context::Context;
use crate::error::EngineError;
use crate::factory::EngineConfig;
use crate::interfaces::{Difficulty, Engine, EngineType, Info, Inspectable};
use crate::rl_encoder::{encode_board, POLICY_SIZE};

/// Difficulty levels for RL-based engines, targeting approximate ELO ratings.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum RLDifficulty {
    /// ~1000 ELO.
    Beginner,
    /// ~1200 ELO.
    Intermediate,
    /// ~1500 ELO.
    Club,
    /// ~2000 ELO.
    Advanced,
    /// ~2200 ELO.
    Master,
}

impl std::fmt::Display for RLDifficulty {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let s = match self {
            RLDifficulty::Beginner => "RL Beginner (1000)",
            RLDifficulty::Intermediate => "RL Intermediate (1200)",
            RLDifficulty::Club => "RL Club (1500)",
            RLDifficulty::Advanced => "RL Advanced (2000)",
            RLDifficulty::Master => "RL Master (2200)",
        };
        write!(f, "{}", s)
    }
}

/// Abstracts ONNX inference so it can be mocked in tests and implemented with a
/// real ONNX Runtime. `close` takes `&self` and uses interior mutability so it
/// can be invoked through the engine's shared reference.
pub trait InferenceSession {
    /// Takes the encoded board (flat `[1, 66, 8, 8]` = 4224 floats) and returns
    /// policy logits (`[4096]`) and a value in `[-1, 1]`.
    fn run_inference(&self, input: &[f32]) -> Result<(Vec<f32>, f32), String>;

    /// Releases resources held by the inference session.
    fn close(&self) -> Result<(), String>;
}

/// RL-based chess engine. When `session` is `None`, `select_move` returns
/// [`EngineError::ModelNotLoaded`].
pub struct RlEngine {
    pub(crate) name: String,
    // Retained for the pending ONNX-runtime integration (spec 008 Slice 11);
    // not yet read until the model is consumed at runtime.
    #[allow(dead_code)]
    pub(crate) difficulty: RLDifficulty,
    #[allow(dead_code)]
    pub(crate) time_limit: Duration,
    closed: AtomicBool,
    /// `None` until a model is loaded.
    pub(crate) session: Option<Box<dyn InferenceSession>>,
}

/// Creates an RL-based engine with the given difficulty.
///
/// The engine is a skeleton until ONNX runtime integration is added;
/// `select_move` will return [`EngineError::ModelNotLoaded`].
pub fn new_rl_engine(
    difficulty: RLDifficulty,
    opts: &[crate::factory::EngineOption],
) -> Result<RlEngine, EngineError> {
    let default_time_limit = match difficulty {
        RLDifficulty::Beginner => Duration::from_secs(3),
        RLDifficulty::Intermediate => Duration::from_secs(4),
        RLDifficulty::Club => Duration::from_secs(5),
        RLDifficulty::Advanced => Duration::from_secs(8),
        RLDifficulty::Master => Duration::from_secs(10),
    };

    let mut cfg = EngineConfig {
        time_limit: default_time_limit,
        ..EngineConfig::default()
    };

    for opt in opts {
        opt(&mut cfg)?;
    }

    Ok(RlEngine {
        name: difficulty.to_string(),
        difficulty,
        time_limit: cfg.time_limit,
        closed: AtomicBool::new(false),
        session: None,
    })
}

impl Engine for RlEngine {
    fn select_move(&self, _ctx: &Context, board: &Board) -> Result<Move, EngineError> {
        if self.closed.load(Ordering::SeqCst) {
            return Err(EngineError::Closed);
        }

        let session = match &self.session {
            Some(s) => s,
            None => return Err(EngineError::ModelNotLoaded),
        };

        // 1. Legal moves.
        let legal_moves = board.legal_moves();
        if legal_moves.is_empty() {
            return Err(EngineError::NoLegalMovesAvailable);
        }
        if legal_moves.len() == 1 {
            return Ok(legal_moves[0]);
        }

        // 2. Encode the board (empty history = zero-filled history planes).
        let input = encode_board(board, &[]);

        // 3. Run inference.
        let (policy_logits, _value) = session
            .run_inference(&input)
            .map_err(EngineError::InferenceFailed)?;

        // 4. Select the best legal move via policy with legal-move masking.
        select_best_move(&legal_moves, &policy_logits)
    }

    fn name(&self) -> &str {
        &self.name
    }

    fn close(&self) -> Result<(), EngineError> {
        // CompareAndSwap: error if already closed.
        if self
            .closed
            .compare_exchange(false, true, Ordering::SeqCst, Ordering::SeqCst)
            .is_err()
        {
            return Err(EngineError::AlreadyClosed);
        }

        if let Some(session) = &self.session {
            return session.close().map_err(EngineError::InferenceFailed);
        }

        Ok(())
    }
}

impl Inspectable for RlEngine {
    fn info(&self) -> Info {
        let mut features = std::collections::HashMap::new();
        features.insert("onnx".to_string(), false);
        Info {
            name: self.name.clone(),
            author: "TermChess".to_string(),
            version: "0.1.0".to_string(),
            engine_type: EngineType::Rl,
            difficulty: Difficulty::Hard,
            features,
        }
    }
}

/// Converts a move to its index in the 4096-element policy vector using a flat
/// from-to encoding: `index = from_square * 64 + to_square`.
pub(crate) fn move_to_policy_index(m: Move) -> usize {
    (m.from.0 as usize) * 64 + (m.to.0 as usize)
}

/// Applies legal-move masking to the policy logits and returns the legal move
/// with the highest policy score. When multiple promotion moves share the same
/// policy index and score, queen promotion is preferred as a tie-breaker.
pub(crate) fn select_best_move(
    legal_moves: &[Move],
    policy_logits: &[f32],
) -> Result<Move, EngineError> {
    if legal_moves.is_empty() {
        return Err(EngineError::NoLegalMoves);
    }
    if policy_logits.len() != POLICY_SIZE {
        return Err(EngineError::InvalidPolicyLength(
            policy_logits.len(),
            POLICY_SIZE,
        ));
    }

    let mut best_move = legal_moves[0];
    let mut best_score = -1e9f32; // Start very negative so any real score wins.

    for &m in legal_moves {
        let idx = move_to_policy_index(m);
        if idx >= policy_logits.len() {
            continue;
        }
        let score = policy_logits[idx];
        if score > best_score {
            best_score = score;
            best_move = m;
        } else if score == best_score && m.promotion == PieceType::Queen {
            // Prefer queen promotion when scores are tied.
            best_move = m;
        }
    }

    Ok(best_move)
}
