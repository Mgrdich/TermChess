//! Position evaluation (ported from `eval.go`).
//!
//! `evaluate` returns a score from White's perspective: positive favors White,
//! negative favors Black. Higher difficulties layer in more terms.

use engine::{Board, Color, GameStatus, PieceType, Square};

use crate::interfaces::Difficulty;

/// Sum of non-pawn, non-king piece values at game start:
/// `2*Q(9) + 4*R(5) + 4*B(3.25) + 4*N(3) = 63`.
pub(crate) const TOTAL_STARTING_MATERIAL: f64 = 63.0;

/// Material level below which the position is a pure endgame.
pub(crate) const ENDGAME_THRESHOLD: f64 = 16.0;

/// Material advantage (in pawns) required to activate endgame mop-up.
pub(crate) const MOP_UP_MATERIAL_THRESHOLD: f64 = 3.0;

/// Standard chess piece values in pawns. King is invaluable (0.0, not counted).
pub(crate) fn piece_value(pt: PieceType) -> f64 {
    match pt {
        PieceType::Pawn => 1.0,
        PieceType::Knight => 3.0,
        PieceType::Bishop => 3.25,
        PieceType::Rook => 5.0,
        PieceType::Queen => 9.0,
        PieceType::King | PieceType::Empty => 0.0,
    }
}

/// Encourages pawn advancement and central control.
#[rustfmt::skip]
pub(crate) const PAWN_TABLE: [f64; 64] = [
    0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
    0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
    0.1, 0.1, 0.2, 0.3, 0.3, 0.2, 0.1, 0.1,
    0.15, 0.15, 0.2, 0.35, 0.35, 0.2, 0.15, 0.15,
    0.2, 0.2, 0.3, 0.4, 0.4, 0.3, 0.2, 0.2,
    0.3, 0.3, 0.4, 0.5, 0.5, 0.4, 0.3, 0.3,
    0.5, 0.5, 0.6, 0.7, 0.7, 0.6, 0.5, 0.5,
    0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
];

/// Encourages centralization and development.
#[rustfmt::skip]
pub(crate) const KNIGHT_TABLE: [f64; 64] = [
    -0.5, -0.4, -0.3, -0.3, -0.3, -0.3, -0.4, -0.5,
    -0.4, -0.2, 0.0, 0.0, 0.0, 0.0, -0.2, -0.4,
    -0.3, 0.0, 0.1, 0.15, 0.15, 0.1, 0.0, -0.3,
    -0.3, 0.05, 0.15, 0.2, 0.2, 0.15, 0.05, -0.3,
    -0.3, 0.0, 0.15, 0.2, 0.2, 0.15, 0.0, -0.3,
    -0.3, 0.05, 0.1, 0.15, 0.15, 0.1, 0.05, -0.3,
    -0.4, -0.2, 0.0, 0.05, 0.05, 0.0, -0.2, -0.4,
    -0.5, -0.4, -0.3, -0.3, -0.3, -0.3, -0.4, -0.5,
];

/// Encourages long diagonals and central control.
#[rustfmt::skip]
pub(crate) const BISHOP_TABLE: [f64; 64] = [
    -0.2, -0.1, -0.1, -0.1, -0.1, -0.1, -0.1, -0.2,
    -0.1, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, -0.1,
    -0.1, 0.0, 0.05, 0.1, 0.1, 0.05, 0.0, -0.1,
    -0.1, 0.05, 0.05, 0.1, 0.1, 0.05, 0.05, -0.1,
    -0.1, 0.0, 0.1, 0.1, 0.1, 0.1, 0.0, -0.1,
    -0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, -0.1,
    -0.1, 0.05, 0.0, 0.0, 0.0, 0.0, 0.05, -0.1,
    -0.2, -0.1, -0.1, -0.1, -0.1, -0.1, -0.1, -0.2,
];

/// Encourages 7th rank occupation and central files.
#[rustfmt::skip]
pub(crate) const ROOK_TABLE: [f64; 64] = [
    0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
    0.05, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.05,
    -0.05, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, -0.05,
    -0.05, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, -0.05,
    -0.05, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, -0.05,
    -0.05, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, -0.05,
    0.25, 0.25, 0.25, 0.25, 0.25, 0.25, 0.25, 0.25,
    0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
];

/// Rewards castled king positions and penalizes exposed kings (middlegame).
#[rustfmt::skip]
pub(crate) const KING_MIDDLEGAME_TABLE: [f64; 64] = [
    0.2, 0.3, 0.1, 0.0, 0.0, 0.1, 0.3, 0.2,
    0.2, 0.2, 0.0, 0.0, 0.0, 0.0, 0.2, 0.2,
    -0.1, -0.2, -0.2, -0.3, -0.3, -0.2, -0.2, -0.1,
    -0.2, -0.3, -0.3, -0.4, -0.4, -0.3, -0.3, -0.2,
    -0.3, -0.4, -0.4, -0.5, -0.5, -0.4, -0.4, -0.3,
    -0.3, -0.4, -0.4, -0.5, -0.5, -0.4, -0.4, -0.3,
    -0.3, -0.4, -0.4, -0.5, -0.5, -0.4, -0.4, -0.3,
    -0.3, -0.4, -0.4, -0.5, -0.5, -0.4, -0.4, -0.3,
];

/// Encourages king centralization in the endgame.
#[rustfmt::skip]
pub(crate) const KING_ENDGAME_TABLE: [f64; 64] = [
    -0.3, -0.4, -0.4, -0.5, -0.5, -0.4, -0.4, -0.3,
    -0.3, -0.4, -0.4, -0.5, -0.5, -0.4, -0.4, -0.3,
    -0.3, -0.4, -0.2, 0.0, 0.0, -0.2, -0.4, -0.3,
    -0.3, -0.3, 0.0, 0.2, 0.2, 0.0, -0.3, -0.3,
    -0.3, -0.3, 0.0, 0.2, 0.2, 0.0, -0.3, -0.3,
    -0.3, -0.4, -0.2, 0.0, 0.0, -0.2, -0.4, -0.3,
    -0.3, -0.4, -0.4, -0.5, -0.5, -0.4, -0.4, -0.3,
    -0.3, -0.4, -0.4, -0.5, -0.5, -0.4, -0.4, -0.3,
];

/// Bonuses for passed pawns indexed by rank (0-7); higher rank = bigger bonus.
#[rustfmt::skip]
pub(crate) const PASSED_PAWN_BONUS: [f64; 8] = [
    0.0, 0.0, 0.1, 0.2, 0.35, 0.6, 1.0, 1.5,
];

/// Returns a score for the position from White's perspective.
pub(crate) fn evaluate(board: &Board, difficulty: Difficulty) -> f64 {
    // 1. Terminal states first.
    let status = board.status();

    if status == GameStatus::Checkmate {
        // The player to move is checkmated; the opponent wins.
        return match board.winner() {
            Some(Color::White) => 10000.0,
            _ => -10000.0,
        };
    }

    if matches!(
        status,
        GameStatus::Stalemate
            | GameStatus::DrawThreefoldRepetition
            | GameStatus::DrawFiftyMoveRule
            | GameStatus::DrawInsufficientMaterial
            | GameStatus::DrawFivefoldRepetition
            | GameStatus::DrawSeventyFiveMoveRule
    ) {
        return 0.0;
    }

    // 2. Material count (all difficulties).
    let material = count_material(board);
    let mut score = material;

    // 3. Piece-square tables, passed pawns, and mobility (Medium+).
    let mut phase = 0.0;
    if difficulty >= Difficulty::Medium {
        phase = compute_game_phase(board);
        score += evaluate_piece_positions(board, phase);
        score += evaluate_passed_pawns(board, phase);
        score += evaluate_mobility(board) * 0.1; // 10% weight
    }

    // 4. King safety and mop-up (Hard only).
    if difficulty >= Difficulty::Hard {
        score += evaluate_king_safety(board);
        score += evaluate_mop_up(board, phase, material);
    }

    score
}

/// Calculates the material balance from White's perspective.
pub(crate) fn count_material(board: &Board) -> f64 {
    let mut score = 0.0;
    for sq in 0..64i8 {
        let piece = board.piece_at(Square(sq));
        if piece.is_empty() {
            continue;
        }
        let value = piece_value(piece.piece_type());
        if piece.color() == Color::White {
            score += value;
        } else {
            score -= value;
        }
    }
    score
}

/// Calculates positional bonuses using piece-square tables. `phase`
/// (0.0 = endgame, 1.0 = opening) interpolates the king tables.
pub(crate) fn evaluate_piece_positions(board: &Board, phase: f64) -> f64 {
    let mut score = 0.0;

    for sq in 0..64usize {
        let piece = board.piece_at(Square(sq as i8));
        if piece.is_empty() {
            continue;
        }

        let piece_type = piece.piece_type();
        let color = piece.color();

        // Flip square for Black pieces (Black plays from rank 7).
        let square_index = if color == Color::Black {
            let rank = sq / 8;
            let file = sq % 8;
            (7 - rank) * 8 + file
        } else {
            sq
        };

        let bonus = match piece_type {
            PieceType::Pawn => PAWN_TABLE[square_index],
            PieceType::Knight => KNIGHT_TABLE[square_index],
            PieceType::Bishop => BISHOP_TABLE[square_index],
            PieceType::Rook => ROOK_TABLE[square_index],
            PieceType::King => {
                let mg = KING_MIDDLEGAME_TABLE[square_index];
                let eg = KING_ENDGAME_TABLE[square_index];
                phase * mg + (1.0 - phase) * eg
            }
            // Queens don't have a specific table.
            PieceType::Queen | PieceType::Empty => 0.0,
        };

        if color == Color::White {
            score += bonus;
        } else {
            score -= bonus;
        }
    }

    score
}

/// Mobility score based on legal move count, from White's perspective.
pub(crate) fn evaluate_mobility(board: &Board) -> f64 {
    let mobility_score = board.legal_moves().len() as f64;
    if board.active_color == Color::Black {
        -mobility_score
    } else {
        mobility_score
    }
}

/// King safety scores for both kings, from White's perspective.
pub(crate) fn evaluate_king_safety(board: &Board) -> f64 {
    let mut score = 0.0;

    let white_king_sq = find_king(board, Color::White);
    if white_king_sq != -1 {
        score += evaluate_king_safety_for_color(board, white_king_sq, Color::White);
    }

    let black_king_sq = find_king(board, Color::Black);
    if black_king_sq != -1 {
        score -= evaluate_king_safety_for_color(board, black_king_sq, Color::Black);
    }

    score
}

/// Returns the square index of the king for `color`, or -1 if not found.
pub(crate) fn find_king(board: &Board, color: Color) -> i32 {
    for sq in 0..64i32 {
        let piece = board.piece_at(Square(sq as i8));
        if piece.piece_type() == PieceType::King && piece.color() == color {
            return sq;
        }
    }
    -1
}

/// Manhattan distance from board center (0-7 range).
pub(crate) fn center_distance(sq: i32) -> f64 {
    let file = (sq % 8) as f64;
    let rank = (sq / 8) as f64;
    (file - 3.5).abs() + (rank - 3.5).abs()
}

/// Chebyshev distance (max of file/rank diff) between two squares.
pub(crate) fn king_distance(sq1: i32, sq2: i32) -> f64 {
    let (file1, rank1) = (sq1 % 8, sq1 / 8);
    let (file2, rank2) = (sq2 % 8, sq2 / 8);
    let file_diff = (file1 - file2).abs() as f64;
    let rank_diff = (rank1 - rank2).abs() as f64;
    file_diff.max(rank_diff)
}

/// Bonus for the winning side to push the enemy king to a corner. Only active
/// when `phase < 0.5` and `|material_balance| >= MOP_UP_MATERIAL_THRESHOLD`.
pub(crate) fn evaluate_mop_up(board: &Board, phase: f64, material_balance: f64) -> f64 {
    if phase >= 0.5 || material_balance.abs() < MOP_UP_MATERIAL_THRESHOLD {
        return 0.0;
    }

    let white_king_sq = find_king(board, Color::White);
    let black_king_sq = find_king(board, Color::Black);
    if white_king_sq == -1 || black_king_sq == -1 {
        return 0.0;
    }

    let phase_scale = 1.0 - phase; // 0.5 to 1.0 range when active

    if material_balance > 0.0 {
        let enemy_corner_bonus = center_distance(black_king_sq) * 0.1;
        let king_proximity_bonus = (7.0 - king_distance(white_king_sq, black_king_sq)) * 0.05;
        (enemy_corner_bonus + king_proximity_bonus) * phase_scale
    } else {
        let enemy_corner_bonus = center_distance(white_king_sq) * 0.1;
        let king_proximity_bonus = (7.0 - king_distance(black_king_sq, white_king_sq)) * 0.05;
        -(enemy_corner_bonus + king_proximity_bonus) * phase_scale
    }
}

/// King safety for a specific color; returns a negative penalty.
pub(crate) fn evaluate_king_safety_for_color(board: &Board, king_sq: i32, color: Color) -> f64 {
    let mut penalty = 0.0;
    penalty += evaluate_pawn_shield(board, king_sq, color);
    penalty += evaluate_open_files_near_king(board, king_sq, color);
    penalty += evaluate_attackers_in_king_zone(board, king_sq, color);
    -penalty
}

/// Penalty for missing pawns in the king's shield.
pub(crate) fn evaluate_pawn_shield(board: &Board, king_sq: i32, color: Color) -> f64 {
    let king_file = king_sq % 8;
    let king_rank = king_sq / 8;

    let mut pawn_count = 0;

    for file_offset in -1..=1i32 {
        let file = king_file + file_offset;
        if !(0..8).contains(&file) {
            continue;
        }

        let target_rank = if color == Color::White {
            king_rank + 1
        } else {
            king_rank - 1
        };

        if (0..8).contains(&target_rank) {
            let sq = target_rank * 8 + file;
            let piece = board.piece_at(Square(sq as i8));
            if piece.piece_type() == PieceType::Pawn && piece.color() == color {
                pawn_count += 1;
            }
        }
    }

    let missing_pawns = 3 - pawn_count;
    missing_pawns as f64 * 0.3
}

/// Penalty for open files (no pawns) near the king.
pub(crate) fn evaluate_open_files_near_king(board: &Board, king_sq: i32, _color: Color) -> f64 {
    let king_file = king_sq % 8;
    let mut penalty = 0.0;

    for file_offset in -1..=1i32 {
        let file = king_file + file_offset;
        if !(0..8).contains(&file) {
            continue;
        }

        let mut has_pawn = false;
        for rank in 0..8i32 {
            let sq = rank * 8 + file;
            let piece = board.piece_at(Square(sq as i8));
            if piece.piece_type() == PieceType::Pawn {
                has_pawn = true;
                break;
            }
        }

        if !has_pawn {
            penalty += 0.25;
        }
    }

    penalty
}

/// Penalty based on the number of squares in the 3x3 king zone attacked by the
/// opponent.
pub(crate) fn evaluate_attackers_in_king_zone(board: &Board, king_sq: i32, color: Color) -> f64 {
    let king_file = king_sq % 8;
    let king_rank = king_sq / 8;

    let opponent_color = if color == Color::Black {
        Color::White
    } else {
        Color::Black
    };

    let mut attacker_count = 0;

    for rank_offset in -1..=1i32 {
        for file_offset in -1..=1i32 {
            let target_rank = king_rank + rank_offset;
            let target_file = king_file + file_offset;

            if !(0..8).contains(&target_rank) || !(0..8).contains(&target_file) {
                continue;
            }

            let target_sq = Square((target_rank * 8 + target_file) as i8);
            if board.is_square_attacked(target_sq, opponent_color) {
                attacker_count += 1;
            }
        }
    }

    attacker_count as f64 * 0.1
}

/// Returns a value between 0.0 (endgame) and 1.0 (opening) based on remaining
/// non-pawn material.
pub(crate) fn compute_game_phase(board: &Board) -> f64 {
    let material = count_non_pawn_material(board);
    if material <= ENDGAME_THRESHOLD {
        return 0.0;
    }
    if material >= TOTAL_STARTING_MATERIAL {
        return 1.0;
    }
    (material - ENDGAME_THRESHOLD) / (TOTAL_STARTING_MATERIAL - ENDGAME_THRESHOLD)
}

/// Sums piece values for all non-pawn, non-king pieces (both colors).
pub(crate) fn count_non_pawn_material(board: &Board) -> f64 {
    let mut material = 0.0;
    for sq in 0..64i8 {
        let piece = board.piece_at(Square(sq));
        if piece.is_empty() {
            continue;
        }
        let pt = piece.piece_type();
        if pt == PieceType::Pawn || pt == PieceType::King {
            continue;
        }
        material += piece_value(pt);
    }
    material
}

/// Reports whether the pawn at `sq` (of `color`) is a passed pawn.
pub(crate) fn is_passed_pawn(board: &Board, sq: i32, color: Color) -> bool {
    let file = sq % 8;
    let rank = sq / 8;

    let f_start = 0.max(file - 1);
    let f_end = 7.min(file + 1);

    for f in f_start..=f_end {
        if color == Color::White {
            for r in (rank + 1)..=7 {
                let check_sq = r * 8 + f;
                let piece = board.piece_at(Square(check_sq as i8));
                if !piece.is_empty()
                    && piece.piece_type() == PieceType::Pawn
                    && piece.color() == Color::Black
                {
                    return false;
                }
            }
        } else {
            for r in (0..=(rank - 1)).rev() {
                let check_sq = r * 8 + f;
                let piece = board.piece_at(Square(check_sq as i8));
                if !piece.is_empty()
                    && piece.piece_type() == PieceType::Pawn
                    && piece.color() == Color::White
                {
                    return false;
                }
            }
        }
    }
    true
}

/// Scores passed pawns from White's perspective. Bonus scales by
/// `1.0 + (1.0 - phase)` to double in a pure endgame.
pub(crate) fn evaluate_passed_pawns(board: &Board, phase: f64) -> f64 {
    let mut score = 0.0;
    let phase_multiplier = 1.0 + (1.0 - phase);

    for sq in 0..64i32 {
        let piece = board.piece_at(Square(sq as i8));
        if piece.is_empty() || piece.piece_type() != PieceType::Pawn {
            continue;
        }

        let color = piece.color();
        if !is_passed_pawn(board, sq, color) {
            continue;
        }

        let rank = (sq / 8) as usize;
        if color == Color::White {
            score += PASSED_PAWN_BONUS[rank] * phase_multiplier;
        } else {
            let flipped_rank = 7 - rank;
            score -= PASSED_PAWN_BONUS[flipped_rank] * phase_multiplier;
        }
    }
    score
}
