//! The Easy bot: weighted random move selection (ported from `random.go`).

use std::sync::atomic::{AtomicBool, Ordering};
use std::time::{Duration, Instant};

use engine::{Board, Move, PieceType, Square};
use rand::Rng;

use crate::context::Context;
use crate::error::EngineError;
use crate::interfaces::{Difficulty, Engine, EngineType, Info, Inspectable};

/// The Easy bot: picks moves with a 70% tactical bias toward captures and a
/// secondary 50% bias toward checks.
#[derive(Debug)]
pub struct RandomEngine {
    pub(crate) name: String,
    pub(crate) time_limit: Duration,
    closed: AtomicBool,
}

impl RandomEngine {
    /// Creates a new random engine with the given name and time limit.
    pub(crate) fn new(name: impl Into<String>, time_limit: Duration) -> RandomEngine {
        RandomEngine {
            name: name.into(),
            time_limit,
            closed: AtomicBool::new(false),
        }
    }
}

impl Engine for RandomEngine {
    fn select_move(&self, ctx: &Context, board: &Board) -> Result<Move, EngineError> {
        if self.closed.load(Ordering::SeqCst) {
            return Err(EngineError::Closed);
        }

        let moves = board.legal_moves();
        if moves.is_empty() {
            return Err(EngineError::NoLegalMovesAvailable);
        }

        // Forced move: return immediately.
        if moves.len() == 1 {
            return Ok(moves[0]);
        }

        // Derived deadline from the engine's own time limit, combined with the
        // caller's context (mirrors context.WithTimeout).
        let deadline = Instant::now() + self.time_limit;

        let captures = filter_captures(board, &moves);
        let checks = filter_checks(board, &moves);

        // Respect cancellation / timeout before selecting.
        if ctx.is_done() {
            return Err(ctx.err().unwrap_or(EngineError::ContextCanceled));
        }
        if Instant::now() >= deadline {
            return Err(EngineError::ContextDeadlineExceeded);
        }

        let mut rng = rand::thread_rng();

        // 70% chance to pick a capture if available.
        if rng.gen::<f64>() < 0.7 && !captures.is_empty() {
            return Ok(captures[rng.gen_range(0..captures.len())]);
        }

        // 50% chance to pick a check if available.
        if rng.gen::<f64>() < 0.5 && !checks.is_empty() {
            return Ok(checks[rng.gen_range(0..checks.len())]);
        }

        // Fallback: any random legal move.
        Ok(moves[rng.gen_range(0..moves.len())])
    }

    fn name(&self) -> &str {
        &self.name
    }

    fn close(&self) -> Result<(), EngineError> {
        self.closed.store(true, Ordering::SeqCst);
        Ok(())
    }
}

impl Inspectable for RandomEngine {
    fn info(&self) -> Info {
        let mut features = std::collections::HashMap::new();
        features.insert("random_selection".to_string(), true);
        features.insert("tactical_awareness".to_string(), true);
        features.insert("weighted_selection".to_string(), true);
        Info {
            name: self.name.clone(),
            author: "TermChess".to_string(),
            version: "1.0".to_string(),
            engine_type: EngineType::Internal,
            difficulty: Difficulty::Easy,
            features,
        }
    }
}

/// Returns all moves that capture an opponent's piece, including en passant.
pub(crate) fn filter_captures(board: &Board, moves: &[Move]) -> Vec<Move> {
    let mut captures = Vec::new();
    for &m in moves {
        // Normal capture: destination has an opponent piece.
        let target_piece = board.piece_at(m.to);
        if !target_piece.is_empty() {
            captures.push(m);
            continue;
        }

        // En passant capture: pawn moves to the en passant target square.
        let moving_piece = board.piece_at(m.from);
        if moving_piece.piece_type() == PieceType::Pawn
            && board.en_passant_sq >= 0
            && m.to == Square(board.en_passant_sq)
        {
            captures.push(m);
        }
    }
    captures
}

/// Returns all moves that give check to the opponent's king.
pub(crate) fn filter_checks(board: &Board, moves: &[Move]) -> Vec<Move> {
    let mut checks = Vec::new();
    for &m in moves {
        let mut board_copy = board.copy();
        let _ = board_copy.make_move(m);

        // After the move, is the new active color (opponent) in check?
        if board_copy.in_check() {
            checks.push(m);
        }
    }
    checks
}
