//! Game state save/load/delete operations, persisted as FEN.

use std::fs;

use engine::Board;

use crate::error::ConfigError;
use crate::paths::{get_config_dir, save_game_path};

/// Saves the current game state to `~/.termchess/savegame.fen`.
///
/// Converts the board to FEN format and writes it to the file. Creates the
/// config directory if it doesn't exist. Returns an error if the file cannot be
/// written.
pub fn save_game(board: &Board) -> Result<(), ConfigError> {
    let save_path = save_game_path()?;
    let config_dir = get_config_dir()?;
    fs::create_dir_all(&config_dir)?;

    let fen = board.to_fen();
    fs::write(&save_path, fen)?;

    Ok(())
}

/// Loads a saved game from `~/.termchess/savegame.fen`.
///
/// Reads the FEN from the file and creates a `Board` from it. Returns an error
/// if the file cannot be read or the FEN is invalid.
pub fn load_game() -> Result<Board, ConfigError> {
    let save_path = save_game_path()?;
    let data = fs::read_to_string(&save_path)?;
    let board = Board::from_fen(&data)?;
    Ok(board)
}

/// Deletes the saved game file at `~/.termchess/savegame.fen`.
///
/// Returns `Ok` if the file doesn't exist (not an error condition). Returns an
/// error only if deletion fails.
pub fn delete_save_game() -> Result<(), ConfigError> {
    let save_path = save_game_path()?;

    if !save_path.exists() {
        // File doesn't exist, nothing to delete.
        return Ok(());
    }

    fs::remove_file(&save_path)?;
    Ok(())
}

/// Checks if a saved game file exists at `~/.termchess/savegame.fen`.
///
/// Returns `true` if the file exists, `false` otherwise.
pub fn save_game_exists() -> bool {
    match save_game_path() {
        Ok(path) => path.exists(),
        Err(_) => false,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::fs_test_lock;
    use engine::{Board, Move};
    use std::fs as stdfs;

    // TestSaveGamePath
    #[test]
    fn save_game_path_valid() {
        let path = save_game_path().expect("SaveGamePath returned error");
        let s = path.to_string_lossy();
        assert!(!s.is_empty(), "SaveGamePath returned empty string");
        assert!(s.contains(".termchess"), "path does not contain .termchess");
        assert!(
            s.ends_with("savegame.fen"),
            "path does not end with savegame.fen"
        );
    }

    // TestSaveGame
    #[test]
    fn save_game_writes_valid_fen() {
        let _guard = fs_test_lock();
        let board = Board::new();
        save_game(&board).expect("SaveGame failed");

        let path = save_game_path().unwrap();
        assert!(path.exists(), "Savegame file was not created");

        let data = stdfs::read_to_string(&path).expect("Failed to read savegame file");
        assert!(!data.is_empty(), "Savegame file is empty");
        Board::from_fen(&data).expect("Savegame contains invalid FEN");

        let _ = stdfs::remove_file(&path);
    }

    // TestSaveGameCreatesDirectory
    #[test]
    fn save_game_creates_directory() {
        let _guard = fs_test_lock();
        let path = save_game_path().unwrap();
        let save_dir = path.parent().unwrap().to_path_buf();

        // Remove the directory if it exists (to test creation).
        let _ = stdfs::remove_dir_all(&save_dir);

        let board = Board::new();
        save_game(&board).expect("SaveGame failed");

        assert!(
            save_dir.exists(),
            "SaveGame did not create .termchess directory"
        );

        let _ = stdfs::remove_file(&path);
    }

    // TestLoadGame
    #[test]
    fn load_game_matches_original() {
        let _guard = fs_test_lock();
        let mut original = Board::new();
        let m = Move::parse("e2e4").unwrap();
        original.make_move(m).unwrap();

        save_game(&original).expect("SaveGame failed");
        let loaded = load_game().expect("LoadGame failed");

        assert_eq!(
            loaded.to_fen(),
            original.to_fen(),
            "Loaded board FEN does not match original"
        );

        let _ = stdfs::remove_file(save_game_path().unwrap());
    }

    // TestLoadGameNonExistent
    #[test]
    fn load_game_non_existent() {
        let _guard = fs_test_lock();
        let _ = stdfs::remove_file(save_game_path().unwrap());
        assert!(
            load_game().is_err(),
            "LoadGame should return error when file doesn't exist"
        );
    }

    // TestLoadGameInvalidFEN
    #[test]
    fn load_game_invalid_fen() {
        let _guard = fs_test_lock();
        let path = save_game_path().unwrap();
        let save_dir = path.parent().unwrap();
        stdfs::create_dir_all(save_dir).unwrap();
        stdfs::write(&path, "invalid fen string").expect("Failed to write test file");

        assert!(
            load_game().is_err(),
            "LoadGame should return error for invalid FEN"
        );

        let _ = stdfs::remove_file(&path);
    }

    // TestSaveLoadRoundTrip
    #[test]
    fn save_load_round_trip() {
        let _guard = fs_test_lock();
        let mut board = Board::new();
        let moves = ["e2e4", "e7e5", "g1f3", "b8c6", "f1c4"];
        for move_str in moves {
            let m = Move::parse(move_str)
                .unwrap_or_else(|_| panic!("Failed to parse move {}", move_str));
            board
                .make_move(m)
                .unwrap_or_else(|_| panic!("Failed to make move {}", move_str));
        }

        let original_fen = board.to_fen();
        save_game(&board).expect("SaveGame failed");
        let loaded = load_game().expect("LoadGame failed");
        let loaded_fen = loaded.to_fen();

        assert_eq!(original_fen, loaded_fen, "Round-trip FEN mismatch");
        assert_eq!(
            board.active_color, loaded.active_color,
            "ActiveColor mismatch"
        );
        assert_eq!(
            board.castling_rights, loaded.castling_rights,
            "CastlingRights mismatch"
        );
        assert_eq!(
            board.en_passant_sq, loaded.en_passant_sq,
            "EnPassantSq mismatch"
        );
        assert_eq!(
            board.half_move_clock, loaded.half_move_clock,
            "HalfMoveClock mismatch"
        );
        assert_eq!(
            board.full_move_num, loaded.full_move_num,
            "FullMoveNum mismatch"
        );

        let _ = stdfs::remove_file(save_game_path().unwrap());
    }

    // TestDeleteSaveGame
    #[test]
    fn delete_save_game_removes_file() {
        let _guard = fs_test_lock();
        let board = Board::new();
        save_game(&board).expect("SaveGame failed");

        let path = save_game_path().unwrap();
        assert!(path.exists(), "Savegame file was not created");

        delete_save_game().expect("DeleteSaveGame failed");
        assert!(!path.exists(), "Savegame file still exists after deletion");
    }

    // TestDeleteSaveGameNonExistent
    #[test]
    fn delete_save_game_non_existent() {
        let _guard = fs_test_lock();
        let _ = stdfs::remove_file(save_game_path().unwrap());
        delete_save_game().expect("DeleteSaveGame should not error when file doesn't exist");
    }

    // TestSaveGameExists
    #[test]
    fn save_game_exists_reflects_state() {
        let _guard = fs_test_lock();
        let path = save_game_path().unwrap();
        let _ = stdfs::remove_file(&path);

        assert!(
            !save_game_exists(),
            "should return false when no save file exists"
        );

        let board = Board::new();
        save_game(&board).expect("SaveGame failed");
        assert!(
            save_game_exists(),
            "should return true when save file exists"
        );

        let _ = stdfs::remove_file(&path);
    }

    // TestSaveGameFilePermissions
    #[cfg(unix)]
    #[test]
    fn save_game_file_permissions() {
        use std::os::unix::fs::PermissionsExt;
        let _guard = fs_test_lock();
        let board = Board::new();
        save_game(&board).expect("SaveGame failed");

        let path = save_game_path().unwrap();
        let info = stdfs::metadata(&path).expect("Failed to stat save file");
        let mode = info.permissions().mode();
        assert!(
            mode & 0o400 != 0,
            "Save file is not readable by owner: {:o}",
            mode
        );

        let _ = stdfs::remove_file(&path);
    }
}
