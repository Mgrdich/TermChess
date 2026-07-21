//! Ported from `eval_test.go`: position evaluation and its components.

use engine::{Board, Color, PieceType};

use crate::eval::{
    center_distance, compute_game_phase, count_material, count_non_pawn_material, evaluate,
    evaluate_attackers_in_king_zone, evaluate_king_safety, evaluate_mobility, evaluate_mop_up,
    evaluate_open_files_near_king, evaluate_passed_pawns, evaluate_pawn_shield,
    evaluate_piece_positions, find_king, is_passed_pawn, king_distance, piece_value,
    ENDGAME_THRESHOLD, KING_ENDGAME_TABLE, KING_MIDDLEGAME_TABLE, PASSED_PAWN_BONUS,
    TOTAL_STARTING_MATERIAL,
};
use crate::interfaces::Difficulty;

use super::from_fen;

#[test]
fn checkmate() {
    let board = from_fen("7k/6Q1/5K2/8/8/8/8/8 b - - 0 1");
    assert_eq!(evaluate(&board, Difficulty::Easy), 10000.0);

    let board = from_fen("8/8/8/8/8/5k2/6q1/7K w - - 0 1");
    assert_eq!(evaluate(&board, Difficulty::Easy), -10000.0);
}

#[test]
fn stalemate() {
    let board = from_fen("7k/5Q2/5K2/8/8/8/8/8 b - - 0 1");
    assert_eq!(evaluate(&board, Difficulty::Easy), 0.0);
}

#[test]
fn start_position() {
    let board = Board::new();
    assert!(evaluate(&board, Difficulty::Easy).abs() <= 0.01);
}

#[test]
fn material_advantage() {
    let cases: [(&str, f64, f64); 5] = [
        (
            "rnb1kbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
            8.0,
            10.0,
        ),
        (
            "rnbqkbn1/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
            4.0,
            6.0,
        ),
        (
            "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNB1KBNR w KQkq - 0 1",
            -10.0,
            -8.0,
        ),
        (
            "rnbqkb1r/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
            2.0,
            4.0,
        ),
        (
            "rn1qkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
            2.5,
            4.0,
        ),
    ];
    for (fen, want_min, want_max) in cases {
        let board = from_fen(fen);
        let score = evaluate(&board, Difficulty::Easy);
        assert!(
            score >= want_min && score <= want_max,
            "fen {}: score {}",
            fen,
            score
        );
    }
}

#[test]
fn test_count_material() {
    let board = Board::new();
    assert!(count_material(&board).abs() <= 0.01);

    let board = from_fen("rnb1kbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1");
    assert!((count_material(&board) - 9.0).abs() <= 0.1);

    let board = from_fen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNB1KBNR w KQkq - 0 1");
    assert!((count_material(&board) + 9.0).abs() <= 0.1);

    let board = from_fen("7k/8/8/8/8/8/PPPPPPPP/7K w - - 0 1");
    assert!((count_material(&board) - 8.0).abs() <= 0.1);
}

#[test]
fn symmetry() {
    let board1 = from_fen("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1");
    assert!(evaluate(&board1, Difficulty::Easy).abs() <= 0.01);

    let board2 = from_fen("rnbqkbnr/pppp1ppp/8/4p3/8/8/PPPPPPPP/RNBQKBNR w KQkq e6 0 1");
    assert!(evaluate(&board2, Difficulty::Easy).abs() <= 0.01);
}

#[test]
fn draw_by_insufficient_material() {
    for fen in [
        "8/8/8/8/8/4k3/8/4K3 w - - 0 1",
        "8/8/8/8/8/4k3/8/4KB2 w - - 0 1",
        "8/8/8/8/8/4k3/8/4KN2 w - - 0 1",
    ] {
        let board = from_fen(fen);
        assert_eq!(evaluate(&board, Difficulty::Easy), 0.0, "fen {}", fen);
    }
}

#[test]
fn draw_fifty_move_rule() {
    let board = from_fen("8/8/4k3/8/8/4K3/8/8 w - - 100 1");
    assert_eq!(evaluate(&board, Difficulty::Easy), 0.0);

    let board = from_fen("8/8/4k3/8/8/4K3/8/8 w - - 150 1");
    assert_eq!(evaluate(&board, Difficulty::Easy), 0.0);
}

#[test]
fn piece_values() {
    let expected = [
        (PieceType::Pawn, 1.0),
        (PieceType::Knight, 3.0),
        (PieceType::Bishop, 3.25),
        (PieceType::Rook, 5.0),
        (PieceType::Queen, 9.0),
        (PieceType::King, 0.0),
    ];
    for (pt, want) in expected {
        assert_eq!(piece_value(pt), want);
    }
}

#[test]
fn complex_positions() {
    let cases: [(&str, f64, f64); 6] = [
        ("8/8/8/8/8/8/8/8 w - - 0 1", 0.0, 0.01),
        ("4k3/8/8/8/8/8/8/4K3 w - - 0 1", 0.0, 0.01),
        ("4k3/8/8/8/8/8/8/4K2R w - - 0 1", 5.0, 0.1),
        ("4k3/8/8/8/8/8/q7/4K2R w - - 0 1", -4.0, 0.1),
        (
            "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 1",
            0.0,
            0.5,
        ),
        ("8/8/8/8/8/8/8/8 w - - 0 1", 0.0, 0.01),
    ];
    for (fen, want, tol) in cases {
        let board = from_fen(fen);
        let score = evaluate(&board, Difficulty::Easy);
        assert!(
            (score - want).abs() <= tol,
            "fen {}: score {} want {}±{}",
            fen,
            score,
            want,
            tol
        );
    }
}

#[test]
fn difficulty_parameter() {
    let board = Board::new();
    let easy = evaluate(&board, Difficulty::Easy);
    let medium = evaluate(&board, Difficulty::Medium);
    let hard = evaluate(&board, Difficulty::Hard);

    assert!(easy.abs() <= 0.01);
    assert_eq!(medium, hard);
    assert!(medium.abs() <= 3.0);
}

#[test]
fn test_evaluate_piece_positions() {
    let central = from_fen("8/8/8/8/4N3/8/8/8 w - - 0 1");
    let corner = from_fen("N7/8/8/8/8/8/8/8 w - - 0 1");
    assert!(evaluate_piece_positions(&central, 0.5) > evaluate_piece_positions(&corner, 0.5));

    let advanced = from_fen("8/4P3/8/8/8/8/8/8 w - - 0 1");
    let starting = from_fen("8/8/8/8/8/8/4P3/8 w - - 0 1");
    assert!(evaluate_piece_positions(&advanced, 0.5) > evaluate_piece_positions(&starting, 0.5));

    let white = from_fen("8/8/8/8/4N3/8/8/8 w - - 0 1");
    let black = from_fen("8/8/8/4n3/8/8/8/8 w - - 0 1");
    let sw = evaluate_piece_positions(&white, 0.5);
    let sb = evaluate_piece_positions(&black, 0.5);
    assert!((sw + sb).abs() <= 0.01);
}

#[test]
fn test_evaluate_mobility() {
    let open = from_fen("8/8/8/8/8/8/8/R3K3 w - - 0 1");
    let cramped = from_fen("8/8/8/8/8/8/PPP5/RK6 w - - 0 1");
    let mob_open = evaluate_mobility(&open);
    assert!(mob_open > evaluate_mobility(&cramped));
    assert!(mob_open > 0.0);

    let black = from_fen("8/8/8/8/8/8/8/r3k3 b - - 0 1");
    assert!(evaluate_mobility(&black) < 0.0);

    let board = Board::new();
    assert_eq!(evaluate_mobility(&board), 20.0);
}

#[test]
fn difficulty_levels() {
    let board = from_fen("4k3/8/8/8/3NP3/8/8/4K3 w - - 0 1");
    let easy = evaluate(&board, Difficulty::Easy);
    let medium = evaluate(&board, Difficulty::Medium);
    let hard = evaluate(&board, Difficulty::Hard);

    assert!((easy - 4.0).abs() <= 0.01);
    assert!(medium > easy);
    assert!(hard >= medium);

    let start = Board::new();
    assert!(evaluate(&start, Difficulty::Easy).abs() <= 0.01);
    assert!(evaluate(&start, Difficulty::Medium).abs() <= 3.0);
    assert!(evaluate(&start, Difficulty::Hard).abs() <= 3.0);
}

#[test]
fn evaluate_piece_positions_all_pieces() {
    let cases: [(&str, f64, f64); 4] = [
        ("8/4P3/8/8/8/8/8/8 w - - 0 1", 0.6, 0.8),
        ("8/8/8/8/8/8/4p3/8 w - - 0 1", -0.8, -0.6),
        ("8/4R3/8/8/8/8/8/8 w - - 0 1", 0.2, 0.3),
        ("8/8/8/3B4/8/8/8/8 w - - 0 1", 0.05, 0.15),
    ];
    for (fen, min_s, max_s) in cases {
        let board = from_fen(fen);
        let score = evaluate_piece_positions(&board, 0.5);
        assert!(
            score >= min_s && score <= max_s,
            "fen {}: score {}",
            fen,
            score
        );
    }
}

#[test]
fn test_evaluate_king_safety() {
    let good = from_fen("4k3/ppp5/8/8/8/8/PPP5/4K3 w - - 0 1");
    let bad = from_fen("4k3/8/8/8/8/8/8/4K3 w - - 0 1");
    assert!(evaluate_king_safety(&good) >= evaluate_king_safety(&bad));
}

#[test]
fn test_find_king() {
    let board = Board::new();
    assert_eq!(find_king(&board, Color::White), 4);
    assert_eq!(find_king(&board, Color::Black), 60);

    let board = from_fen("4k3/8/8/8/8/8/8/4K3 w - - 0 1");
    assert_eq!(find_king(&board, Color::White), 4);
    assert_eq!(find_king(&board, Color::Black), 60);
}

#[test]
fn test_evaluate_pawn_shield() {
    let complete = from_fen("8/8/8/8/8/8/PPP5/1K6 w - - 0 1");
    let sq = find_king(&complete, Color::White);
    assert!(evaluate_pawn_shield(&complete, sq, Color::White).abs() <= 0.01);

    let partial = from_fen("8/8/8/8/8/8/P7/1K6 w - - 0 1");
    let sq = find_king(&partial, Color::White);
    assert!((evaluate_pawn_shield(&partial, sq, Color::White) - 2.0 * 0.3).abs() <= 0.01);

    let none = from_fen("8/8/8/8/8/8/8/1K6 w - - 0 1");
    let sq = find_king(&none, Color::White);
    assert!((evaluate_pawn_shield(&none, sq, Color::White) - 3.0 * 0.3).abs() <= 0.01);
}

#[test]
fn test_evaluate_open_files_near_king() {
    let closed = from_fen("8/8/8/8/8/ppp5/PPP5/1K1k4 w - - 0 1");
    let sq = find_king(&closed, Color::White);
    assert!(evaluate_open_files_near_king(&closed, sq, Color::White).abs() <= 0.01);

    let one_open = from_fen("8/8/8/8/8/8/PP6/1K6 w - - 0 1");
    let sq = find_king(&one_open, Color::White);
    assert!((evaluate_open_files_near_king(&one_open, sq, Color::White) - 0.25).abs() <= 0.01);

    let all_open = from_fen("8/8/8/8/8/8/8/1K6 w - - 0 1");
    let sq = find_king(&all_open, Color::White);
    assert!(
        (evaluate_open_files_near_king(&all_open, sq, Color::White) - 3.0 * 0.25).abs() <= 0.01
    );
}

#[test]
fn test_evaluate_attackers_in_king_zone() {
    let safe = from_fen("4k3/8/8/8/8/8/8/4K3 w - - 0 1");
    let sq = find_king(&safe, Color::White);
    assert!(evaluate_attackers_in_king_zone(&safe, sq, Color::White).abs() <= 0.01);

    let danger = from_fen("4k3/8/8/8/8/8/8/q3K3 w - - 0 1");
    let sq = find_king(&danger, Color::White);
    assert!(evaluate_attackers_in_king_zone(&danger, sq, Color::White) > 0.0);

    let rook = from_fen("4k3/8/8/8/8/8/8/R3K3 w - - 0 1");
    let bk = find_king(&rook, Color::Black);
    assert_eq!(
        evaluate_attackers_in_king_zone(&rook, bk, Color::Black),
        0.0
    );
}

#[test]
fn king_safety_only_hard() {
    let board = from_fen("8/4k3/3ppp2/8/8/8/8/4K3 w - - 0 1");
    let ks = evaluate_king_safety(&board);
    assert!(ks.abs() >= 0.01);

    let medium = evaluate(&board, Difficulty::Medium);
    let hard = evaluate(&board, Difficulty::Hard);
    assert!(hard < medium);

    let both_safe = from_fen("4k3/ppp5/8/8/8/8/PPP5/4K3 w - - 0 1");
    let m2 = evaluate(&both_safe, Difficulty::Medium);
    let h2 = evaluate(&both_safe, Difficulty::Hard);

    let diff_asym = (hard - medium).abs();
    let diff_both = (h2 - m2).abs();
    assert!(diff_both < diff_asym);
}

#[test]
fn compute_game_phase_starting() {
    let board = Board::new();
    assert!((compute_game_phase(&board) - 1.0).abs() <= 0.01);
}

#[test]
fn compute_game_phase_bare_kings() {
    let board = from_fen("4k3/8/8/8/8/8/8/4K3 w - - 0 1");
    assert_eq!(compute_game_phase(&board), 0.0);
}

#[test]
fn compute_game_phase_kings_and_pawns() {
    let board = from_fen("4k3/pppppppp/8/8/8/8/PPPPPPPP/4K3 w - - 0 1");
    assert_eq!(compute_game_phase(&board), 0.0);
}

#[test]
fn compute_game_phase_minor_above_threshold() {
    let board = from_fen("1rb1k3/8/8/8/8/8/8/1RB1KN2 w - - 0 1");
    let phase = compute_game_phase(&board);
    let expected = (19.5 - ENDGAME_THRESHOLD) / (TOTAL_STARTING_MATERIAL - ENDGAME_THRESHOLD);
    assert!((phase - expected).abs() <= 0.01);
    assert!(phase > 0.0);
}

#[test]
fn compute_game_phase_half_material() {
    let board = from_fen("1r1qkb2/8/8/8/8/8/8/1RRQKB2 w - - 0 1");
    assert!((compute_game_phase(&board) - 0.5).abs() <= 0.01);
}

#[test]
fn test_count_non_pawn_material() {
    let board = Board::new();
    assert!((count_non_pawn_material(&board) - 63.0).abs() <= 0.01);

    let cases: [(&str, f64); 7] = [
        ("4k3/8/8/8/8/8/8/4K3 w - - 0 1", 0.0),
        ("4k3/pppppppp/8/8/8/8/PPPPPPPP/4K3 w - - 0 1", 0.0),
        ("4k3/8/8/8/8/8/8/3QK3 w - - 0 1", 9.0),
        ("r3k3/8/8/8/8/8/8/4K3 w - - 0 1", 5.0),
        ("rnb1k3/8/8/8/8/8/8/RNB1K3 w - - 0 1", 22.5),
        ("4k3/pppppppp/8/8/8/8/PPPPPPPP/4K3 w - - 0 1", 0.0),
        ("r2qk2r/8/8/8/8/8/8/R2QK2R w - - 0 1", 38.0),
    ];
    for (fen, want) in cases {
        let board = from_fen(fen);
        assert!(
            (count_non_pawn_material(&board) - want).abs() <= 0.01,
            "fen {}",
            fen
        );
    }
}

#[test]
fn king_safety_symmetry() {
    let no_shield = from_fen("4k3/8/8/8/8/8/8/4K3 w - - 0 1");
    assert!(evaluate_king_safety(&no_shield).abs() <= 0.1);

    let good = from_fen("4k3/ppp5/8/8/8/8/PPP5/4K3 w - - 0 1");
    assert!(evaluate_king_safety(&good).abs() <= 0.1);
}

#[test]
fn king_phase_interpolation_endgame_center() {
    let board = from_fen("8/8/8/8/4K3/8/8/8 w - - 0 1");
    let eg = evaluate_piece_positions(&board, 0.0);
    let mg = evaluate_piece_positions(&board, 1.0);
    assert!(eg > mg);
}

#[test]
fn king_phase_interpolation_middlegame_castled() {
    let board = from_fen("8/8/8/8/8/8/8/6K1 w - - 0 1");
    let mg = evaluate_piece_positions(&board, 1.0);
    let eg = evaluate_piece_positions(&board, 0.0);
    assert!(mg > eg);
}

#[test]
fn king_phase_interpolation_half_phase() {
    let board = from_fen("8/8/8/8/8/8/8/6K1 w - - 0 1");
    let half = evaluate_piece_positions(&board, 0.5);
    let expected = 0.5 * KING_MIDDLEGAME_TABLE[6] + 0.5 * KING_ENDGAME_TABLE[6];
    assert!((half - expected).abs() <= 0.001);
}

#[test]
fn is_passed_pawn_isolated() {
    let board = from_fen("7k/8/8/4P3/8/8/8/K7 w - - 0 1");
    assert!(is_passed_pawn(&board, 36, Color::White));
}

#[test]
fn is_passed_pawn_blocked_same_file() {
    let board = from_fen("7k/8/4p3/8/4P3/8/8/K7 w - - 0 1");
    assert!(!is_passed_pawn(&board, 28, Color::White));
}

#[test]
fn is_passed_pawn_blocked_adjacent_file() {
    let board = from_fen("7k/8/8/3p4/4P3/8/8/K7 w - - 0 1");
    assert!(!is_passed_pawn(&board, 28, Color::White));
}

#[test]
fn is_passed_pawn_black() {
    let board = from_fen("7k/8/8/8/3p4/8/8/K7 w - - 0 1");
    assert!(is_passed_pawn(&board, 27, Color::Black));
}

#[test]
fn evaluate_passed_pawns_single_white() {
    let board = from_fen("7k/8/4P3/8/8/8/8/K7 w - - 0 1");
    let score = evaluate_passed_pawns(&board, 1.0);
    assert!(score > 0.0);
    assert!((score - PASSED_PAWN_BONUS[5]).abs() <= 0.01);
}

#[test]
fn evaluate_passed_pawns_single_black() {
    let board = from_fen("7k/8/8/8/8/3p4/8/K7 w - - 0 1");
    let score = evaluate_passed_pawns(&board, 1.0);
    assert!(score < 0.0);
    assert!((score + PASSED_PAWN_BONUS[5]).abs() <= 0.01);
}

#[test]
fn evaluate_passed_pawns_rank_bonus() {
    let advanced = from_fen("7k/4P3/8/8/8/8/8/K7 w - - 0 1");
    let early = from_fen("7k/8/8/8/4P3/8/8/K7 w - - 0 1");
    let sa = evaluate_passed_pawns(&advanced, 1.0);
    let se = evaluate_passed_pawns(&early, 1.0);
    assert!(sa > se);
    assert!((sa - PASSED_PAWN_BONUS[6]).abs() <= 0.01);
    assert!((se - PASSED_PAWN_BONUS[3]).abs() <= 0.01);
}

#[test]
fn evaluate_passed_pawns_endgame_amplification() {
    let board = from_fen("7k/8/4P3/8/8/8/8/K7 w - - 0 1");
    let opening = evaluate_passed_pawns(&board, 1.0);
    let endgame = evaluate_passed_pawns(&board, 0.0);
    assert!(endgame > opening);
    assert!((endgame - 2.0 * opening).abs() <= 0.01);
}

#[test]
fn test_center_distance() {
    let cases = [
        (27, 1.0),
        (28, 1.0),
        (35, 1.0),
        (36, 1.0),
        (0, 7.0),
        (7, 7.0),
        (56, 7.0),
        (63, 7.0),
    ];
    for (sq, want) in cases {
        assert!((center_distance(sq) - want).abs() <= 0.01, "sq {}", sq);
    }
}

#[test]
fn test_king_distance() {
    let cases = [
        (28, 28, 0.0),
        (28, 29, 1.0),
        (28, 36, 1.0),
        (28, 37, 1.0),
        (0, 63, 7.0),
        (0, 7, 7.0),
        (0, 56, 7.0),
        (28, 30, 2.0),
    ];
    for (a, b, want) in cases {
        assert!((king_distance(a, b) - want).abs() <= 0.01, "{} {}", a, b);
    }
}

#[test]
fn mop_up_inactive_middlegame() {
    let board = from_fen("4k3/8/8/8/8/8/8/R3K3 w - - 0 1");
    assert_eq!(evaluate_mop_up(&board, 0.8, 5.0), 0.0);
    assert_eq!(evaluate_mop_up(&board, 0.5, 5.0), 0.0);
}

#[test]
fn mop_up_inactive_even_material() {
    let board = from_fen("4k3/8/8/8/8/8/8/4K3 w - - 0 1");
    assert_eq!(evaluate_mop_up(&board, 0.0, 0.0), 0.0);
    assert_eq!(evaluate_mop_up(&board, 0.0, 2.0), 0.0);
}

#[test]
fn mop_up_active_white_winning() {
    let board = from_fen("4k3/8/8/8/4K3/8/8/8 w - - 0 1");
    assert!(evaluate_mop_up(&board, 0.0, 4.0) > 0.0);
}

#[test]
fn mop_up_enemy_king_corner_bonus() {
    let corner = from_fen("k7/8/8/8/4K3/8/8/8 w - - 0 1");
    let center = from_fen("8/8/8/4k3/4K3/8/8/8 w - - 0 1");
    assert!(evaluate_mop_up(&corner, 0.0, 5.0) > evaluate_mop_up(&center, 0.0, 5.0));
}

#[test]
fn mop_up_king_proximity_bonus() {
    let close = from_fen("8/8/8/4k3/4K3/8/8/8 w - - 0 1");
    let far = from_fen("8/8/8/4k3/8/8/8/K7 w - - 0 1");
    assert!(evaluate_mop_up(&close, 0.0, 5.0) > evaluate_mop_up(&far, 0.0, 5.0));
}

#[test]
fn mop_up_black_winning() {
    let board = from_fen("4K3/8/8/8/4k3/8/8/8 w - - 0 1");
    assert!(evaluate_mop_up(&board, 0.0, -4.0) < 0.0);

    let sym = from_fen("8/8/8/4k3/4K3/8/8/8 w - - 0 1");
    let black_winning = evaluate_mop_up(&sym, 0.0, -5.0);
    let white_winning = evaluate_mop_up(&sym, 0.0, 5.0);
    assert!(black_winning < 0.0);
    assert!(white_winning > 0.0);
    assert!((black_winning.abs() - white_winning.abs()).abs() <= 0.01);
}
