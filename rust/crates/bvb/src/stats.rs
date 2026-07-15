//! Aggregate statistics for multi-game sessions (ported from `stats.go`).

use std::time::Duration;

use crate::types::GameResult;

/// Holds computed statistics for a multi-game session.
#[derive(Debug, Clone, Default)]
pub struct AggregateStats {
    /// The number of completed games.
    pub total_games: i32,
    /// The name of the white bot.
    pub white_bot_name: String,
    /// The name of the black bot.
    pub black_bot_name: String,
    /// The number of games won by the white bot.
    pub white_wins: i32,
    /// The number of games won by the black bot.
    pub black_wins: i32,
    /// The number of drawn games.
    pub draws: i32,
    /// The white bot's win percentage (0-100).
    pub white_win_pct: f64,
    /// The black bot's win percentage (0-100).
    pub black_win_pct: f64,
    /// The average number of moves per game.
    pub avg_move_count: f64,
    /// The average game duration.
    pub avg_duration: Duration,
    /// The game with the fewest moves.
    pub shortest_game: GameResult,
    /// The game with the most moves.
    pub longest_game: GameResult,
    /// All game results in order.
    pub individual_results: Vec<GameResult>,
}

/// Calculates aggregate statistics from a slice of game results.
pub fn compute_stats(results: &[GameResult], white_name: &str, black_name: &str) -> AggregateStats {
    if results.is_empty() {
        return AggregateStats {
            white_bot_name: white_name.to_string(),
            black_bot_name: black_name.to_string(),
            ..AggregateStats::default()
        };
    }

    let mut stats = AggregateStats {
        total_games: results.len() as i32,
        white_bot_name: white_name.to_string(),
        black_bot_name: black_name.to_string(),
        individual_results: results.to_vec(),
        shortest_game: results[0].clone(),
        longest_game: results[0].clone(),
        ..AggregateStats::default()
    };

    let mut total_moves: i64 = 0;
    let mut total_duration = Duration::ZERO;

    for r in results {
        // Count wins.
        if r.winner == "Draw" {
            stats.draws += 1;
        } else if r.winner == white_name {
            stats.white_wins += 1;
        } else if r.winner == black_name {
            stats.black_wins += 1;
        }

        // Accumulate for averages.
        total_moves += r.move_count as i64;
        total_duration += r.duration;

        // Track shortest/longest by move count.
        if r.move_count < stats.shortest_game.move_count {
            stats.shortest_game = r.clone();
        }
        if r.move_count > stats.longest_game.move_count {
            stats.longest_game = r.clone();
        }
    }

    // Calculate averages.
    stats.avg_move_count = total_moves as f64 / stats.total_games as f64;
    stats.avg_duration = total_duration / stats.total_games as u32;

    // Calculate win percentages.
    stats.white_win_pct = stats.white_wins as f64 / stats.total_games as f64 * 100.0;
    stats.black_win_pct = stats.black_wins as f64 / stats.total_games as f64 * 100.0;

    stats
}

#[cfg(test)]
mod tests {
    use super::*;
    use engine::Color;

    fn result(
        game_number: i32,
        winner: &str,
        move_count: i32,
        secs: u64,
        end_reason: &str,
    ) -> GameResult {
        GameResult {
            game_number,
            winner: winner.to_string(),
            move_count,
            duration: Duration::from_secs(secs),
            end_reason: end_reason.to_string(),
            ..GameResult::default()
        }
    }

    #[test]
    fn compute_stats_win_counts() {
        let results = vec![
            {
                let mut r = result(1, "White Bot", 30, 5, "checkmate");
                r.winner_color = Color::White;
                r
            },
            {
                let mut r = result(2, "White Bot", 25, 4, "checkmate");
                r.winner_color = Color::White;
                r
            },
            {
                let mut r = result(3, "White Bot", 35, 6, "checkmate");
                r.winner_color = Color::White;
                r
            },
            {
                let mut r = result(4, "Black Bot", 40, 7, "checkmate");
                r.winner_color = Color::Black;
                r
            },
            {
                let mut r = result(5, "Black Bot", 50, 9, "checkmate");
                r.winner_color = Color::Black;
                r
            },
            result(6, "Draw", 60, 10, "stalemate"),
        ];

        let stats = compute_stats(&results, "White Bot", "Black Bot");

        assert_eq!(stats.total_games, 6);
        assert_eq!(stats.white_wins, 3);
        assert_eq!(stats.black_wins, 2);
        assert_eq!(stats.draws, 1);

        assert!((stats.white_win_pct - 50.0).abs() <= 0.01);
        let expected_black = 100.0 * 2.0 / 6.0;
        assert!((stats.black_win_pct - expected_black).abs() <= 0.01);

        assert_eq!(stats.white_bot_name, "White Bot");
        assert_eq!(stats.black_bot_name, "Black Bot");
    }

    #[test]
    fn compute_stats_all_draws() {
        let results = vec![
            result(1, "Draw", 50, 8, "stalemate"),
            result(2, "Draw", 55, 9, "stalemate"),
            result(3, "Draw", 60, 10, "move limit exceeded"),
        ];

        let stats = compute_stats(&results, "Alpha", "Beta");

        assert_eq!(stats.total_games, 3);
        assert_eq!(stats.draws, 3);
        assert_eq!(stats.white_wins, 0);
        assert_eq!(stats.black_wins, 0);
        assert_eq!(stats.white_win_pct, 0.0);
        assert_eq!(stats.black_win_pct, 0.0);
    }

    #[test]
    fn compute_stats_averages() {
        let results = vec![
            result(1, "White Bot", 10, 2, "checkmate"),
            result(2, "Black Bot", 20, 4, "checkmate"),
            result(3, "Draw", 30, 6, "stalemate"),
        ];

        let stats = compute_stats(&results, "White Bot", "Black Bot");

        assert!((stats.avg_move_count - 20.0).abs() <= 0.01);
        assert_eq!(stats.avg_duration, Duration::from_secs(4));
    }

    #[test]
    fn compute_stats_shortest_longest() {
        let results = vec![
            result(1, "White Bot", 40, 5, "checkmate"),
            result(2, "Black Bot", 15, 3, "checkmate"),
            result(3, "Draw", 80, 12, "stalemate"),
            result(4, "White Bot", 55, 7, "checkmate"),
        ];

        let stats = compute_stats(&results, "White Bot", "Black Bot");

        assert_eq!(stats.shortest_game.game_number, 2);
        assert_eq!(stats.shortest_game.move_count, 15);
        assert_eq!(stats.longest_game.game_number, 3);
        assert_eq!(stats.longest_game.move_count, 80);
    }

    #[test]
    fn compute_stats_single_game() {
        let mut r = result(1, "White Bot", 42, 7, "checkmate");
        r.winner_color = Color::White;
        let results = vec![r];

        let stats = compute_stats(&results, "White Bot", "Black Bot");

        assert_eq!(stats.total_games, 1);
        assert_eq!(stats.white_wins, 1);
        assert_eq!(stats.black_wins, 0);
        assert_eq!(stats.draws, 0);
        assert_eq!(stats.white_win_pct, 100.0);
        assert_eq!(stats.black_win_pct, 0.0);
        assert_eq!(stats.avg_move_count, 42.0);
        assert_eq!(stats.avg_duration, Duration::from_secs(7));

        assert_eq!(stats.shortest_game.game_number, 1);
        assert_eq!(stats.longest_game.game_number, 1);
        assert_eq!(
            stats.shortest_game.move_count,
            stats.longest_game.move_count
        );
    }

    #[test]
    fn compute_stats_empty() {
        let stats = compute_stats(&[], "White Bot", "Black Bot");

        assert_eq!(stats.total_games, 0);
        assert_eq!(stats.white_wins, 0);
        assert_eq!(stats.black_wins, 0);
        assert_eq!(stats.draws, 0);
        assert_eq!(stats.white_win_pct, 0.0);
        assert_eq!(stats.black_win_pct, 0.0);
        assert_eq!(stats.avg_move_count, 0.0);
        assert_eq!(stats.avg_duration, Duration::ZERO);
        assert_eq!(stats.white_bot_name, "White Bot");
        assert_eq!(stats.black_bot_name, "Black Bot");
        // Go returns a nil slice; here that maps to an empty Vec.
        assert!(stats.individual_results.is_empty());
    }

    #[test]
    fn compute_stats_empty_slice() {
        let stats = compute_stats(&Vec::<GameResult>::new(), "A", "B");

        assert_eq!(stats.total_games, 0);
        assert_eq!(stats.white_bot_name, "A");
        assert_eq!(stats.black_bot_name, "B");
    }

    #[test]
    fn compute_stats_individual_results() {
        let mut results = vec![
            result(1, "White Bot", 30, 5, "checkmate"),
            result(2, "Black Bot", 45, 8, "checkmate"),
        ];

        let stats = compute_stats(&results, "White Bot", "Black Bot");

        assert_eq!(stats.individual_results.len(), 2);
        assert_eq!(stats.individual_results[0].game_number, 1);
        assert_eq!(stats.individual_results[1].game_number, 2);

        // Modify original to ensure it is a copy.
        results[0].game_number = 99;
        assert_ne!(stats.individual_results[0].game_number, 99);
    }
}
