//! Install-method detection, binary replacement, and uninstall logic.
//!
//! Ported from the Go `InstallMethod`, `DetectInstallMethod`, `ReplaceBinary`,
//! `Uninstall`, and `GetGoInstallMessage` functions.

use std::fmt;
use std::fs;
use std::io::ErrorKind;
use std::path::{Path, PathBuf};

use crate::error::UpdaterError;

/// Represents how TermChess was installed.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum InstallMethod {
    /// Installed via `go install`.
    GoInstall,
    /// Installed via the install script.
    InstallScript,
    /// Installation method could not be determined.
    Unknown,
}

impl InstallMethod {
    /// Returns the string representation of the install method, matching the Go
    /// constant values.
    pub fn as_str(&self) -> &'static str {
        match self {
            InstallMethod::GoInstall => "go-install",
            InstallMethod::InstallScript => "install-script",
            InstallMethod::Unknown => "unknown",
        }
    }
}

impl fmt::Display for InstallMethod {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.write_str(self.as_str())
    }
}

/// Classifies an executable path into an [`InstallMethod`] using the same
/// substring checks as the Go implementation.
pub(crate) fn classify_install_path(path: &str) -> InstallMethod {
    if path.contains("/go/bin/") {
        InstallMethod::GoInstall
    } else if path.contains("/.local/bin/") || path.contains("/usr/local/bin/") {
        InstallMethod::InstallScript
    } else {
        InstallMethod::Unknown
    }
}

/// Identifies how TermChess was installed by examining the executable path.
pub fn detect_install_method() -> InstallMethod {
    let exec_path = match std::env::current_exe() {
        Ok(p) => p,
        Err(_) => return InstallMethod::Unknown,
    };

    let real_path = fs::canonicalize(&exec_path).unwrap_or(exec_path);

    classify_install_path(&real_path.to_string_lossy())
}

/// Appends a literal suffix (e.g. `.new`) to a path, mirroring Go's
/// `realPath + ".new"` string concatenation.
fn with_suffix(path: &Path, suffix: &str) -> PathBuf {
    let mut s = path.as_os_str().to_os_string();
    s.push(suffix);
    PathBuf::from(s)
}

/// Maps an I/O error to [`UpdaterError::PermissionDenied`] when it is a
/// permission error, otherwise wraps it with a contextual message.
fn map_io_err(context: &str, err: std::io::Error) -> UpdaterError {
    if err.kind() == ErrorKind::PermissionDenied {
        UpdaterError::PermissionDenied
    } else {
        UpdaterError::Message(format!("{context}: {err}"))
    }
}

/// Atomically replaces the current executable with new binary data, preserving
/// the original file's permissions.
pub fn replace_binary(new_binary_data: &[u8]) -> Result<(), UpdaterError> {
    let exec_path = std::env::current_exe()
        .map_err(|e| UpdaterError::Message(format!("getting executable path: {e}")))?;

    let real_path = fs::canonicalize(&exec_path).unwrap_or(exec_path);

    replace_binary_at(&real_path, new_binary_data)
}

/// The testable core of [`replace_binary`]: performs the atomic rename-swap for
/// an explicit target path.
pub(crate) fn replace_binary_at(
    real_path: &Path,
    new_binary_data: &[u8],
) -> Result<(), UpdaterError> {
    // Get the current binary's permissions to preserve them.
    let file_info = fs::metadata(real_path)
        .map_err(|e| UpdaterError::Message(format!("getting file info: {e}")))?;
    let permissions = file_info.permissions();

    // 1. Write new binary to temp file.
    let tmp_path = with_suffix(real_path, ".new");
    if let Err(e) = fs::write(&tmp_path, new_binary_data) {
        return Err(map_io_err("writing new binary", e));
    }
    // Preserve the original permissions on the new file.
    let _ = fs::set_permissions(&tmp_path, permissions);

    // 2. Rename current to .old.
    let old_path = with_suffix(real_path, ".old");
    if let Err(e) = fs::rename(real_path, &old_path) {
        // Clean up temp file.
        let _ = fs::remove_file(&tmp_path);
        return Err(map_io_err("backing up current binary", e));
    }

    // 3. Rename new to current.
    if let Err(e) = fs::rename(&tmp_path, real_path) {
        // Try to restore old binary and clean up temp file.
        let _ = fs::rename(&old_path, real_path);
        let _ = fs::remove_file(&tmp_path);
        return Err(map_io_err("installing new binary", e));
    }

    // 4. Delete old (best effort, don't fail if this doesn't work).
    let _ = fs::remove_file(&old_path);

    Ok(())
}

/// Returns the message to show users who installed via `go install`.
pub fn get_go_install_message() -> String {
    "You installed TermChess via 'go install'.\n\
\n\
To upgrade, run:\n\
  go install github.com/Mgrdich/TermChess/cmd/termchess@latest\n\
\n\
Or switch to our install script for automatic upgrades:\n\
  curl -fsSL https://raw.githubusercontent.com/Mgrdich/TermChess/main/scripts/install.sh | bash"
        .to_string()
}

/// Removes the TermChess binary and configuration directory.
pub fn uninstall() -> Result<(), UpdaterError> {
    let exec_path = std::env::current_exe()
        .map_err(|e| UpdaterError::Message(format!("getting executable path: {e}")))?;

    let real_path = fs::canonicalize(&exec_path).unwrap_or(exec_path);

    let config_dir = config::get_config_dir()?;

    uninstall_at(&real_path, &config_dir)
}

/// The testable core of [`uninstall`]: removes an explicit binary path and
/// configuration directory.
pub(crate) fn uninstall_at(binary_path: &Path, config_dir: &Path) -> Result<(), UpdaterError> {
    // Remove the binary.
    if let Err(e) = fs::remove_file(binary_path) {
        return Err(map_io_err("removing binary", e));
    }

    // Remove config directory recursively. Go's os.RemoveAll returns nil when
    // the path does not exist, so treat NotFound as success.
    match fs::remove_dir_all(config_dir) {
        Ok(()) => {}
        Err(e) if e.kind() == ErrorKind::NotFound => {}
        Err(e) => return Err(map_io_err("removing config directory", e)),
    }

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_install_method_string() {
        assert_eq!(InstallMethod::GoInstall.as_str(), "go-install");
        assert_eq!(InstallMethod::InstallScript.as_str(), "install-script");
        assert_eq!(InstallMethod::Unknown.as_str(), "unknown");
        assert_eq!(InstallMethod::GoInstall.to_string(), "go-install");
        assert_eq!(InstallMethod::InstallScript.to_string(), "install-script");
        assert_eq!(InstallMethod::Unknown.to_string(), "unknown");
    }

    #[test]
    fn test_detect_install_method_returns_valid() {
        let method = detect_install_method();
        assert!(matches!(
            method,
            InstallMethod::GoInstall | InstallMethod::InstallScript | InstallMethod::Unknown
        ));
    }

    #[test]
    fn test_detect_install_method_path_parsing() {
        let cases: &[(&str, InstallMethod)] = &[
            ("/home/user/go/bin/termchess", InstallMethod::GoInstall),
            (
                "/Users/developer/go/bin/termchess",
                InstallMethod::GoInstall,
            ),
            (
                "/home/user/.local/bin/termchess",
                InstallMethod::InstallScript,
            ),
            ("/usr/local/bin/termchess", InstallMethod::InstallScript),
            ("/opt/termchess/bin/termchess", InstallMethod::Unknown),
            ("/tmp/termchess", InstallMethod::Unknown),
        ];
        for (path, want) in cases {
            assert_eq!(classify_install_path(path), *want, "path {path:?}");
        }
    }

    #[test]
    fn test_get_go_install_message() {
        let msg = get_go_install_message();
        assert!(msg.contains("go install"));
        assert!(msg.contains("github.com/Mgrdich/TermChess"));
        assert!(msg.contains("install.sh"));
    }

    #[test]
    fn test_replace_binary_at() {
        let dir = tempfile::tempdir().unwrap();
        let current = dir.path().join("termchess");
        fs::write(&current, b"old binary").unwrap();

        let new_data = b"new binary";
        replace_binary_at(&current, new_data).unwrap();

        // The new binary is in place.
        assert_eq!(fs::read(&current).unwrap(), new_data);

        // The .old and .new files were cleaned up / moved.
        assert!(!with_suffix(&current, ".old").exists());
        assert!(!with_suffix(&current, ".new").exists());
    }

    #[test]
    fn test_replace_binary_at_preserves_permissions() {
        let dir = tempfile::tempdir().unwrap();
        let current = dir.path().join("termchess");
        fs::write(&current, b"old binary").unwrap();
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            fs::set_permissions(&current, fs::Permissions::from_mode(0o755)).unwrap();
        }

        replace_binary_at(&current, b"new binary").unwrap();

        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            let mode = fs::metadata(&current).unwrap().permissions().mode();
            assert_eq!(mode & 0o777, 0o755);
        }
    }

    #[test]
    fn test_uninstall_at_logic() {
        let dir = tempfile::tempdir().unwrap();
        let binary = dir.path().join("termchess");
        fs::write(&binary, b"fake binary").unwrap();

        let config_dir = dir.path().join(".termchess");
        fs::create_dir_all(&config_dir).unwrap();
        fs::write(config_dir.join("config.toml"), b"# config").unwrap();

        uninstall_at(&binary, &config_dir).unwrap();

        assert!(!binary.exists());
        assert!(!config_dir.exists());
    }

    #[test]
    fn test_uninstall_at_nested_config_dir() {
        let dir = tempfile::tempdir().unwrap();
        let binary = dir.path().join("termchess");
        fs::write(&binary, b"fake binary").unwrap();

        let config_dir = dir.path().join(".termchess");
        let sub_dir = config_dir.join("subdir");
        fs::create_dir_all(&sub_dir).unwrap();
        fs::write(config_dir.join("config.toml"), b"# config").unwrap();
        fs::write(config_dir.join("savegame.fen"), b"fen string").unwrap();
        fs::write(sub_dir.join("nested.txt"), b"nested").unwrap();

        uninstall_at(&binary, &config_dir).unwrap();

        assert!(!config_dir.exists());
    }

    #[test]
    fn test_uninstall_at_empty_config_dir() {
        let dir = tempfile::tempdir().unwrap();
        let binary = dir.path().join("termchess");
        fs::write(&binary, b"fake binary").unwrap();

        let config_dir = dir.path().join(".termchess");
        fs::create_dir_all(&config_dir).unwrap();

        uninstall_at(&binary, &config_dir).unwrap();

        assert!(!config_dir.exists());
    }

    #[test]
    fn test_uninstall_at_non_existent_binary() {
        let dir = tempfile::tempdir().unwrap();
        let binary = dir.path().join("does-not-exist");
        let config_dir = dir.path().join(".termchess");

        let err = uninstall_at(&binary, &config_dir).unwrap_err();
        // Removing a non-existent binary is an error (not a permission error).
        assert!(matches!(err, UpdaterError::Message(_)));
    }
}
