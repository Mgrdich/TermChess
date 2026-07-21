//! URL and filename builders, ported from the Go `GetAssetURL`,
//! `GetChecksumsURL`, and `GetBinaryFilename` functions.

use crate::{REPO_NAME, REPO_OWNER};

/// The default host for release download URLs.
pub(crate) const GITHUB_DOWNLOAD_BASE: &str = "https://github.com";

/// Builds an asset download URL against an arbitrary base host.
///
/// Used internally so downloads can be redirected to a test server while the
/// public [`get_asset_url`] keeps the hardcoded `github.com` host.
pub(crate) fn asset_url_with_base(base: &str, version: &str, os: &str, arch: &str) -> String {
    let binary_name = format!("termchess-{version}-{os}-{arch}");
    format!("{base}/{REPO_OWNER}/{REPO_NAME}/releases/download/{version}/{binary_name}")
}

/// Builds a checksums file URL against an arbitrary base host.
pub(crate) fn checksums_url_with_base(base: &str, version: &str) -> String {
    format!("{base}/{REPO_OWNER}/{REPO_NAME}/releases/download/{version}/checksums.txt")
}

/// Constructs the download URL for a specific platform binary.
///
/// The version should include the `v` prefix (e.g. `"v0.1.0"`).
/// `os` values: `"darwin"`, `"linux"`. `arch` values: `"amd64"`, `"arm64"`.
pub fn get_asset_url(version: &str, os: &str, arch: &str) -> String {
    asset_url_with_base(GITHUB_DOWNLOAD_BASE, version, os, arch)
}

/// Constructs the download URL for the checksums file.
pub fn get_checksums_url(version: &str) -> String {
    checksums_url_with_base(GITHUB_DOWNLOAD_BASE, version)
}

/// Returns the binary filename for the given version and platform.
pub fn get_binary_filename(version: &str, goos: &str, goarch: &str) -> String {
    format!("termchess-{version}-{goos}-{goarch}")
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_get_asset_url() {
        let cases: &[(&str, &str, &str, &str)] = &[
            (
                "v0.1.0",
                "darwin",
                "amd64",
                "https://github.com/Mgrdich/TermChess/releases/download/v0.1.0/termchess-v0.1.0-darwin-amd64",
            ),
            (
                "v0.1.0",
                "darwin",
                "arm64",
                "https://github.com/Mgrdich/TermChess/releases/download/v0.1.0/termchess-v0.1.0-darwin-arm64",
            ),
            (
                "v0.1.0",
                "linux",
                "amd64",
                "https://github.com/Mgrdich/TermChess/releases/download/v0.1.0/termchess-v0.1.0-linux-amd64",
            ),
            (
                "v0.1.0",
                "linux",
                "arm64",
                "https://github.com/Mgrdich/TermChess/releases/download/v0.1.0/termchess-v0.1.0-linux-arm64",
            ),
            (
                "v1.2.3",
                "darwin",
                "arm64",
                "https://github.com/Mgrdich/TermChess/releases/download/v1.2.3/termchess-v1.2.3-darwin-arm64",
            ),
            (
                "v0.2.0-beta.1",
                "linux",
                "amd64",
                "https://github.com/Mgrdich/TermChess/releases/download/v0.2.0-beta.1/termchess-v0.2.0-beta.1-linux-amd64",
            ),
        ];
        for (version, os, arch, want) in cases {
            assert_eq!(&get_asset_url(version, os, arch), want);
        }
    }

    #[test]
    fn test_get_checksums_url() {
        assert_eq!(
            get_checksums_url("v0.1.0"),
            "https://github.com/Mgrdich/TermChess/releases/download/v0.1.0/checksums.txt"
        );
        assert_eq!(
            get_checksums_url("v1.2.3"),
            "https://github.com/Mgrdich/TermChess/releases/download/v1.2.3/checksums.txt"
        );
        assert_eq!(
            get_checksums_url("v0.2.0-beta.1"),
            "https://github.com/Mgrdich/TermChess/releases/download/v0.2.0-beta.1/checksums.txt"
        );
    }

    #[test]
    fn test_get_binary_filename() {
        assert_eq!(
            get_binary_filename("v0.1.0", "darwin", "amd64"),
            "termchess-v0.1.0-darwin-amd64"
        );
        assert_eq!(
            get_binary_filename("v1.2.3", "linux", "arm64"),
            "termchess-v1.2.3-linux-arm64"
        );
    }
}
