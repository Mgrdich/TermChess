//! JSON export of completed Bot vs Bot sessions (ported from `export.go`).

use std::path::PathBuf;

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

use crate::manager::SessionManager;

/// Errors that can occur while saving a session export.
#[derive(Debug, thiserror::Error)]
pub enum ExportError {
    /// The user's home directory could not be resolved.
    #[error("failed to get home directory")]
    HomeDirUnavailable,
    /// An I/O error occurred creating the directory or writing the file.
    #[error("io error: {0}")]
    Io(#[from] std::io::Error),
    /// The export could not be serialized to JSON.
    #[error("failed to marshal export: {0}")]
    Serialize(#[from] serde_json::Error),
}

/// The complete export data for a Bot vs Bot session, serialized to JSON.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SessionExport {
    /// When the export was generated.
    pub timestamp: DateTime<Utc>,
    /// The name of the white bot.
    pub white_bot: String,
    /// The name of the black bot.
    pub black_bot: String,
    /// The number of completed games.
    pub total_games: i32,
    /// Games won by white.
    pub white_wins: i32,
    /// Games won by black.
    pub black_wins: i32,
    /// Drawn games.
    pub draws: i32,
    /// Average number of moves per game.
    pub average_moves: f64,
    /// Per-game export data.
    pub games: Vec<GameExport>,
}

/// The export data for a single game.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GameExport {
    /// The sequence number of this game.
    pub game_number: i32,
    /// "White", "Black", or "Draw".
    pub result: String,
    /// "Checkmate", "Stalemate", etc.
    #[serde(rename = "termination")]
    pub termination_reason: String,
    /// The total number of moves played.
    pub move_count: i32,
    /// Moves in coordinate notation (e.g., "e2e4").
    pub moves: Vec<String>,
    /// Final position in FEN.
    pub final_fen: String,
}

impl SessionManager {
    /// Generates a [`SessionExport`] from the manager's completed games,
    /// collecting all game data and calculating aggregate statistics.
    pub fn export_stats(&self, white_bot: &str, black_bot: &str) -> SessionExport {
        let (sessions, white_name) = self.export_snapshot();

        let mut export = SessionExport {
            timestamp: Utc::now(),
            white_bot: white_bot.to_string(),
            black_bot: black_bot.to_string(),
            total_games: 0,
            white_wins: 0,
            black_wins: 0,
            draws: 0,
            average_moves: 0.0,
            games: Vec::new(),
        };

        let mut total_moves: i64 = 0;
        for s in &sessions {
            if !s.is_finished() {
                continue;
            }
            let result = match s.result() {
                Some(r) => r,
                None => continue,
            };

            export.total_games += 1;

            let result_str = if result.winner == "Draw" {
                export.draws += 1;
                "Draw"
            } else if result.winner == white_name {
                export.white_wins += 1;
                "White"
            } else {
                export.black_wins += 1;
                "Black"
            };

            let moves: Vec<String> = result.move_history.iter().map(|m| m.to_string()).collect();
            total_moves += result.move_count as i64;

            export.games.push(GameExport {
                game_number: result.game_number,
                result: result_str.to_string(),
                termination_reason: result.end_reason.clone(),
                move_count: result.move_count,
                moves,
                final_fen: result.final_fen.clone(),
            });
        }

        if export.total_games > 0 {
            export.average_moves = total_moves as f64 / export.total_games as f64;
        }

        export
    }
}

/// Saves a [`SessionExport`] to a JSON file.
///
/// If `dir` is empty, uses the default directory `~/.termchess/stats/`. Returns
/// the full path to the created file.
///
/// Unlike Go's `SaveSessionExport`, the export cannot be null (the type system
/// enforces it), so there is no nil-check error path.
pub fn save_session_export(export: &SessionExport, dir: &str) -> Result<PathBuf, ExportError> {
    // Use default directory if not specified.
    let dir_path = if dir.is_empty() {
        let home = dirs::home_dir().ok_or(ExportError::HomeDirUnavailable)?;
        home.join(".termchess").join("stats")
    } else {
        PathBuf::from(dir)
    };

    // Create directory if it doesn't exist.
    std::fs::create_dir_all(&dir_path)?;

    // Generate filename with timestamp.
    let filename = format!(
        "bvb_session_{}.json",
        export.timestamp.format("%Y-%m-%d_%H-%M-%S")
    );
    let full_path = dir_path.join(filename);

    // Marshal to JSON with indentation and write.
    let data = serde_json::to_string_pretty(export)?;
    std::fs::write(&full_path, data)?;

    Ok(full_path)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::types::PlaybackSpeed;
    use bot::Difficulty;
    use chrono::TimeZone;
    use std::thread;
    use std::time::{Duration, Instant};

    fn wait_all_finished(m: &SessionManager, timeout: Duration) -> bool {
        let start = Instant::now();
        while !m.all_finished() {
            if start.elapsed() > timeout {
                return false;
            }
            thread::sleep(Duration::from_millis(50));
        }
        true
    }

    #[test]
    fn export_stats_basic() {
        let m = SessionManager::new(
            Difficulty::Easy,
            Difficulty::Easy,
            "Easy White",
            "Easy Black",
            2,
            0,
        );
        m.set_speed_field(PlaybackSpeed::Instant);
        m.start().expect("start failed");

        assert!(
            wait_all_finished(&m, Duration::from_secs(60)),
            "games did not complete within timeout"
        );

        let export = m.export_stats("Easy", "Easy");
        assert_eq!(export.white_bot, "Easy");
        assert_eq!(export.black_bot, "Easy");
        assert_eq!(export.total_games, 2);
        assert_eq!(
            export.white_wins + export.black_wins + export.draws,
            export.total_games
        );
        assert!(export.average_moves > 0.0);
        assert_eq!(export.games.len(), 2);

        for game in &export.games {
            assert!(game.game_number > 0);
            assert!(matches!(game.result.as_str(), "White" | "Black" | "Draw"));
            assert!(!game.termination_reason.is_empty());
            assert!(game.move_count > 0);
            assert_eq!(game.moves.len() as i32, game.move_count);
            assert!(!game.final_fen.is_empty());
        }
    }

    #[test]
    fn export_stats_move_history() {
        let m = SessionManager::new(
            Difficulty::Easy,
            Difficulty::Easy,
            "Easy White",
            "Easy Black",
            1,
            0,
        );
        m.set_speed_field(PlaybackSpeed::Instant);
        m.start().expect("start failed");

        assert!(
            wait_all_finished(&m, Duration::from_secs(60)),
            "game did not complete within timeout"
        );

        let export = m.export_stats("Easy", "Easy");
        assert_eq!(export.games.len(), 1);
        let game = &export.games[0];

        for mv in &game.moves {
            let bytes = mv.as_bytes();
            assert!(
                mv.len() >= 4 && mv.len() <= 5,
                "invalid move length: {}",
                mv
            );
            assert!(
                (b'a'..=b'h').contains(&bytes[0]),
                "invalid from file: {}",
                mv
            );
            assert!(
                (b'1'..=b'8').contains(&bytes[1]),
                "invalid from rank: {}",
                mv
            );
            assert!((b'a'..=b'h').contains(&bytes[2]), "invalid to file: {}", mv);
            assert!((b'1'..=b'8').contains(&bytes[3]), "invalid to rank: {}", mv);
        }
    }

    #[test]
    fn export_stats_empty() {
        let m = SessionManager::new(Difficulty::Easy, Difficulty::Easy, "W", "B", 3, 0);
        let export = m.export_stats("Easy", "Easy");
        assert_eq!(export.total_games, 0);
        assert_eq!(export.white_wins, 0);
        assert_eq!(export.black_wins, 0);
        assert_eq!(export.draws, 0);
        assert_eq!(export.average_moves, 0.0);
        assert_eq!(export.games.len(), 0);
    }

    fn sample_export(ts: DateTime<Utc>) -> SessionExport {
        SessionExport {
            timestamp: ts,
            white_bot: "Easy".to_string(),
            black_bot: "Medium".to_string(),
            total_games: 3,
            white_wins: 1,
            black_wins: 1,
            draws: 1,
            average_moves: 42.5,
            games: vec![GameExport {
                game_number: 1,
                result: "White".to_string(),
                termination_reason: "Checkmate".to_string(),
                move_count: 40,
                moves: vec!["e2e4".into(), "e7e5".into(), "g1f3".into()],
                final_fen: "rnbqkbnr/pppp1ppp/8/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2"
                    .to_string(),
            }],
        }
    }

    #[test]
    fn save_creates_file() {
        let tmp = tempfile::tempdir().unwrap();
        let export = sample_export(Utc.with_ymd_and_hms(2024, 1, 15, 10, 30, 45).unwrap());

        let path = save_session_export(&export, tmp.path().to_str().unwrap()).expect("save failed");
        assert!(path.exists(), "file was not created at {:?}", path);

        let expected = "bvb_session_2024-01-15_10-30-45.json";
        assert!(
            path.to_str().unwrap().ends_with(expected),
            "filename = {:?}, want suffix {}",
            path,
            expected
        );
    }

    #[test]
    fn save_json_format() {
        let tmp = tempfile::tempdir().unwrap();
        let export = SessionExport {
            timestamp: Utc.with_ymd_and_hms(2024, 2, 20, 14, 0, 0).unwrap(),
            white_bot: "Hard".to_string(),
            black_bot: "Hard".to_string(),
            total_games: 2,
            white_wins: 1,
            black_wins: 0,
            draws: 1,
            average_moves: 50.0,
            games: vec![
                GameExport {
                    game_number: 1,
                    result: "White".to_string(),
                    termination_reason: "Checkmate".to_string(),
                    move_count: 45,
                    moves: vec!["d2d4".into(), "d7d5".into()],
                    final_fen: "some-fen-1".to_string(),
                },
                GameExport {
                    game_number: 2,
                    result: "Draw".to_string(),
                    termination_reason: "Stalemate".to_string(),
                    move_count: 55,
                    moves: vec!["e2e4".into()],
                    final_fen: "some-fen-2".to_string(),
                },
            ],
        };

        let path = save_session_export(&export, tmp.path().to_str().unwrap()).expect("save failed");
        let data = std::fs::read_to_string(&path).unwrap();
        let loaded: SessionExport = serde_json::from_str(&data).expect("failed to unmarshal JSON");

        assert_eq!(loaded.white_bot, "Hard");
        assert_eq!(loaded.black_bot, "Hard");
        assert_eq!(loaded.total_games, 2);
        assert_eq!(loaded.white_wins, 1);
        assert_eq!(loaded.draws, 1);
        assert_eq!(loaded.average_moves, 50.0);
        assert_eq!(loaded.games.len(), 2);
        assert_eq!(loaded.games[0].result, "White");
        assert_eq!(loaded.games[1].result, "Draw");
    }

    #[test]
    fn save_creates_directory() {
        let tmp = tempfile::tempdir().unwrap();
        let nested = tmp.path().join("nested").join("deep").join("stats");

        let export = SessionExport {
            timestamp: Utc::now(),
            white_bot: "Easy".to_string(),
            black_bot: "Easy".to_string(),
            total_games: 0,
            white_wins: 0,
            black_wins: 0,
            draws: 0,
            average_moves: 0.0,
            games: vec![],
        };

        let path = save_session_export(&export, nested.to_str().unwrap()).expect("save failed");
        assert!(nested.is_dir(), "directory was not created");
        assert!(path.exists(), "file was not created at {:?}", path);
    }

    #[test]
    fn save_default_directory() {
        let export = SessionExport {
            timestamp: Utc.with_ymd_and_hms(2024, 3, 1, 12, 0, 0).unwrap(),
            white_bot: "Test".to_string(),
            black_bot: "Test".to_string(),
            total_games: 0,
            white_wins: 0,
            black_wins: 0,
            draws: 0,
            average_moves: 0.0,
            games: vec![],
        };

        let path = save_session_export(&export, "").expect("save failed");
        assert!(path.exists(), "file was not created at {:?}", path);
        let path_str = path.to_str().unwrap();
        assert!(
            path_str.contains(".termchess"),
            "path does not contain .termchess"
        );
        assert!(path_str.contains("stats"), "path does not contain stats");

        let _ = std::fs::remove_file(&path);
    }

    #[test]
    fn game_export_termination_reasons() {
        let reasons = [
            "Checkmate",
            "Stalemate",
            "Insufficient Material",
            "move limit exceeded",
        ];
        for reason in reasons {
            let export = SessionExport {
                timestamp: Utc::now(),
                white_bot: "Test".to_string(),
                black_bot: "Test".to_string(),
                total_games: 1,
                white_wins: 0,
                black_wins: 0,
                draws: 0,
                average_moves: 0.0,
                games: vec![GameExport {
                    game_number: 1,
                    result: "White".to_string(),
                    termination_reason: reason.to_string(),
                    move_count: 10,
                    moves: vec!["e2e4".into()],
                    final_fen: "test-fen".to_string(),
                }],
            };

            let tmp = tempfile::tempdir().unwrap();
            let path =
                save_session_export(&export, tmp.path().to_str().unwrap()).expect("save failed");
            let data = std::fs::read_to_string(&path).unwrap();
            let loaded: SessionExport = serde_json::from_str(&data).unwrap();
            assert_eq!(loaded.games[0].termination_reason, reason);
        }
    }

    #[test]
    fn export_stats_timestamp() {
        let m = SessionManager::new(Difficulty::Easy, Difficulty::Easy, "W", "B", 1, 0);
        let before = Utc::now();
        let export = m.export_stats("Easy", "Easy");
        let after = Utc::now();
        assert!(export.timestamp >= before);
        assert!(export.timestamp <= after);
    }

    #[test]
    fn move_history_recorded_correctly() {
        let m = SessionManager::new(
            Difficulty::Easy,
            Difficulty::Easy,
            "Easy White",
            "Easy Black",
            1,
            0,
        );
        m.set_speed_field(PlaybackSpeed::Instant);
        m.start().expect("start failed");

        assert!(
            wait_all_finished(&m, Duration::from_secs(60)),
            "game did not complete within timeout"
        );

        let sessions = m.sessions();
        assert_eq!(sessions.len(), 1);
        let result = sessions[0].result().expect("session result is nil");

        let export = m.export_stats("Easy", "Easy");
        assert_eq!(export.games.len(), 1);
        assert_eq!(export.games[0].move_count, result.move_count);
        assert_eq!(export.games[0].moves.len(), result.move_history.len());
        for (i, mv) in result.move_history.iter().enumerate() {
            assert_eq!(export.games[0].moves[i], mv.to_string());
        }
    }

    #[test]
    fn export_stats_result_mapping() {
        let m = SessionManager::new(
            Difficulty::Easy,
            Difficulty::Easy,
            "Easy White",
            "Easy Black",
            3,
            0,
        );
        m.set_speed_field(PlaybackSpeed::Instant);
        m.start().expect("start failed");

        assert!(
            wait_all_finished(&m, Duration::from_secs(90)),
            "games did not complete within timeout"
        );

        let export = m.export_stats("Easy", "Easy");
        let mut white = 0;
        let mut black = 0;
        let mut draw = 0;
        for game in &export.games {
            match game.result.as_str() {
                "White" => white += 1,
                "Black" => black += 1,
                "Draw" => draw += 1,
                other => panic!("invalid result {} for game {}", other, game.game_number),
            }
        }
        assert_eq!(white, export.white_wins);
        assert_eq!(black, export.black_wins);
        assert_eq!(draw, export.draws);
    }

    #[test]
    fn final_fen_recorded() {
        let m = SessionManager::new(
            Difficulty::Easy,
            Difficulty::Easy,
            "Easy White",
            "Easy Black",
            1,
            0,
        );
        m.set_speed_field(PlaybackSpeed::Instant);
        m.start().expect("start failed");

        assert!(
            wait_all_finished(&m, Duration::from_secs(60)),
            "game did not complete within timeout"
        );

        let export = m.export_stats("Easy", "Easy");
        assert_eq!(export.games.len(), 1);
        let fen = &export.games[0].final_fen;
        assert!(!fen.is_empty());
        assert!(fen.split(' ').count() >= 4, "FEN has too few parts");
        assert!(
            engine::Board::from_fen(fen).is_ok(),
            "final FEN is not valid: {}",
            fen
        );
    }
}
