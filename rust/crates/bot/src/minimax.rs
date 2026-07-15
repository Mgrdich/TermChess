//! Medium and Hard bots: minimax with alpha-beta pruning (ported from
//! `minimax.go`).

use std::sync::atomic::{AtomicBool, Ordering};
use std::time::{Duration, Instant};

use engine::{Board, Color, Move};
use rand::Rng;

use crate::context::Context;
use crate::error::EngineError;
use crate::eval::evaluate;
use crate::interfaces::{
    Configurable, Difficulty, Engine, EngineType, Info, Inspectable, MinimaxConfig,
};

/// Weights for the evaluation components. Only `material` is currently applied
/// by the evaluation; the others are placeholders kept for parity with Go.
#[derive(Debug, Clone, Copy, PartialEq)]
pub(crate) struct EvalWeights {
    pub(crate) material: f64,
    pub(crate) piece_square: f64,
    pub(crate) mobility: f64,
    pub(crate) king_safety: f64,
}

/// Returns evaluation weights based on difficulty. All difficulties currently
/// use `material = 1.0` with the remaining terms at 0.0.
pub(crate) fn get_default_weights(_difficulty: Difficulty) -> EvalWeights {
    EvalWeights {
        material: 1.0,
        piece_square: 0.0,
        mobility: 0.0,
        king_safety: 0.0,
    }
}

/// Medium and Hard bots using minimax with alpha-beta pruning.
#[derive(Debug)]
pub struct MinimaxEngine {
    pub(crate) name: String,
    pub(crate) difficulty: Difficulty,
    pub(crate) max_depth: i32,
    pub(crate) time_limit: Duration,
    pub(crate) eval_weights: EvalWeights,
    /// If true, disables random tie-breaking among equal-scored moves.
    pub(crate) deterministic: bool,
    closed: AtomicBool,
}

impl MinimaxEngine {
    pub(crate) fn new(
        name: impl Into<String>,
        difficulty: Difficulty,
        max_depth: i32,
        time_limit: Duration,
        eval_weights: EvalWeights,
        deterministic: bool,
    ) -> MinimaxEngine {
        MinimaxEngine {
            name: name.into(),
            difficulty,
            max_depth,
            time_limit,
            eval_weights,
            deterministic,
            closed: AtomicBool::new(false),
        }
    }

    /// Returns true once the context is canceled or the derived deadline passes.
    fn is_done(ctx: &Context, deadline: Instant) -> bool {
        ctx.is_done() || Instant::now() >= deadline
    }

    /// Performs a minimax search at a specific depth. Returns `None` if the
    /// search was interrupted by a timeout (matching the Go error path).
    fn search_depth(
        &self,
        ctx: &Context,
        deadline: Instant,
        board: &Board,
        depth: i32,
    ) -> Option<Move> {
        let moves = board.legal_moves();
        if moves.is_empty() {
            return None;
        }

        let moves = self.order_moves(board, &moves);

        let mut alpha = f64::NEG_INFINITY;
        let beta = f64::INFINITY;

        let mut best_move: Option<Move> = None;
        let mut best_score = f64::NEG_INFINITY;
        let mut best_count = 0; // moves sharing the best score (for tie-breaking)

        for m in moves {
            if Self::is_done(ctx, deadline) {
                return None;
            }

            let mut board_copy = board.copy();
            if board_copy.make_move(m).is_err() {
                continue;
            }

            // Negamax: negate the score since we switched sides. ply = 1.
            let score = -self.alpha_beta(ctx, deadline, &board_copy, depth - 1, -beta, -alpha, 1);

            if score > best_score {
                best_score = score;
                best_move = Some(m);
                best_count = 1;
            } else if score == best_score && !self.deterministic {
                // Random tie-breaking among equal scores.
                best_count += 1;
                if rand::thread_rng().gen_range(0..best_count) == 0 {
                    best_move = Some(m);
                }
            }

            if score > alpha {
                alpha = score;
            }

            if alpha >= beta {
                break;
            }
        }

        best_move
    }

    /// Recursive negamax search with alpha-beta pruning. Returns the score from
    /// the perspective of the side to move.
    #[allow(clippy::too_many_arguments)]
    fn alpha_beta(
        &self,
        ctx: &Context,
        deadline: Instant,
        board: &Board,
        depth: i32,
        mut alpha: f64,
        beta: f64,
        ply: i32,
    ) -> f64 {
        // Timeout: return a neutral score so partial searches don't corrupt
        // the iterative-deepening results.
        if Self::is_done(ctx, deadline) {
            return 0.0;
        }

        if depth == 0 || board.is_game_over() {
            return self.leaf_score(board, ply);
        }

        let moves = board.legal_moves();
        if moves.is_empty() {
            // Checkmate or stalemate; evaluate() already handles this.
            return self.leaf_score(board, ply);
        }

        let moves = self.order_moves(board, &moves);

        let mut max_score = f64::NEG_INFINITY;

        for m in moves {
            let mut board_copy = board.copy();
            if board_copy.make_move(m).is_err() {
                continue;
            }

            let score = -self.alpha_beta(
                ctx,
                deadline,
                &board_copy,
                depth - 1,
                -beta,
                -alpha,
                ply + 1,
            );

            if score > max_score {
                max_score = score;
            }
            if score > alpha {
                alpha = score;
            }
            if alpha >= beta {
                break; // beta cutoff
            }
        }

        max_score
    }

    /// Evaluates a leaf from White's perspective, adjusts mate scores to prefer
    /// faster mates, then flips for negamax if Black is to move.
    fn leaf_score(&self, board: &Board, ply: i32) -> f64 {
        let mut white_score = evaluate(board, self.difficulty);

        if white_score >= 9999.0 {
            white_score -= ply as f64;
        } else if white_score <= -9999.0 {
            white_score += ply as f64;
        }

        if board.active_color == Color::Black {
            -white_score
        } else {
            white_score
        }
    }

    /// Simple move ordering (captures first) to improve alpha-beta pruning.
    pub(crate) fn order_moves(&self, board: &Board, moves: &[Move]) -> Vec<Move> {
        let mut captures = Vec::new();
        let mut non_captures = Vec::new();

        for &m in moves {
            if !board.piece_at(m.to).is_empty() {
                captures.push(m);
            } else {
                non_captures.push(m);
            }
        }

        captures.extend(non_captures);
        captures
    }
}

impl Engine for MinimaxEngine {
    fn select_move(&self, ctx: &Context, board: &Board) -> Result<Move, EngineError> {
        if self.closed.load(Ordering::SeqCst) {
            return Err(EngineError::Closed);
        }

        let deadline = Instant::now() + self.time_limit;

        let moves = board.legal_moves();
        if moves.is_empty() {
            return Err(EngineError::NoLegalMovesAvailable);
        }
        if moves.len() == 1 {
            return Ok(moves[0]);
        }

        // Iterative deepening from depth 1 to max_depth.
        let mut best_move: Option<Move> = None;

        for depth in 1..=self.max_depth {
            if Self::is_done(ctx, deadline) {
                // Timeout: return the best move from the previous iteration.
                return Ok(best_move.unwrap_or(moves[0]));
            }

            match self.search_depth(ctx, deadline, board, depth) {
                Some(m) => best_move = Some(m),
                None => return Ok(best_move.unwrap_or(moves[0])),
            }
        }

        Ok(best_move.unwrap_or(moves[0]))
    }

    fn name(&self) -> &str {
        &self.name
    }

    fn close(&self) -> Result<(), EngineError> {
        self.closed.store(true, Ordering::SeqCst);
        Ok(())
    }
}

impl Configurable for MinimaxEngine {
    fn configure(&mut self, config: MinimaxConfig) -> Result<(), EngineError> {
        if let Some(depth) = config.search_depth {
            if !(1..=20).contains(&depth) {
                return Err(EngineError::Config(format!(
                    "search depth must be 1-20, got {}",
                    depth
                )));
            }
            self.max_depth = depth;
        }

        if let Some(limit) = config.time_limit {
            if limit.is_zero() {
                return Err(EngineError::Config(
                    "time limit must be positive, got 0s".to_string(),
                ));
            }
            self.time_limit = limit;
        }

        if let Some(w) = config.material_weight {
            self.eval_weights.material = w;
        }
        if let Some(w) = config.piece_square_weight {
            self.eval_weights.piece_square = w;
        }
        if let Some(w) = config.mobility_weight {
            self.eval_weights.mobility = w;
        }
        if let Some(w) = config.king_safety_weight {
            self.eval_weights.king_safety = w;
        }

        Ok(())
    }
}

impl Inspectable for MinimaxEngine {
    fn info(&self) -> Info {
        let mut features = std::collections::HashMap::new();
        features.insert("alpha_beta".to_string(), true);
        features.insert("iterative_deepening".to_string(), true);
        features.insert("move_ordering".to_string(), true);
        features.insert("configurable".to_string(), true);
        features.insert(
            "piece_square_tables".to_string(),
            self.difficulty >= Difficulty::Medium,
        );
        features.insert(
            "mobility".to_string(),
            self.difficulty >= Difficulty::Medium,
        );
        features.insert(
            "king_safety".to_string(),
            self.difficulty >= Difficulty::Hard,
        );
        Info {
            name: self.name.clone(),
            author: "TermChess".to_string(),
            version: "1.0".to_string(),
            engine_type: EngineType::Internal,
            difficulty: self.difficulty,
            features,
        }
    }
}
