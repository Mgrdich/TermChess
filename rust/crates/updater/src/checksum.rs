//! Checksum verification and parsing, ported from the Go `VerifyChecksum`,
//! `ParseChecksums`, and `GetExpectedChecksum` functions.

use std::collections::HashMap;

use sha2::{Digest, Sha256};

use crate::error::UpdaterError;

/// Verifies that the SHA-256 hash of `data` matches the expected hex string.
///
/// Returns `true` if the checksum matches, `false` otherwise. Empty `data` or
/// an empty `expected` string always return `false`. The comparison is
/// case-insensitive (matching Go's `strings.EqualFold`).
pub fn verify_checksum(data: &[u8], expected: &str) -> bool {
    if data.is_empty() || expected.is_empty() {
        return false;
    }

    let hash = Sha256::digest(data);
    let actual = hex::encode(hash);

    actual.eq_ignore_ascii_case(expected)
}

/// Parses a `checksums.txt` file content into a map of filename to checksum.
///
/// Expected format: `"checksum  filename"` (any run of whitespace between the
/// checksum and filename). The last whitespace-separated field on each line is
/// treated as the filename.
pub fn parse_checksums(content: &str) -> HashMap<String, String> {
    let mut checksums = HashMap::new();
    for line in content.lines() {
        let line = line.trim();
        if line.is_empty() {
            continue;
        }
        // Format: "checksum  filename" or "checksum filename".
        let parts: Vec<&str> = line.split_whitespace().collect();
        if parts.len() >= 2 {
            let checksum = parts[0].to_string();
            let filename = parts[parts.len() - 1].to_string(); // last part is filename
            checksums.insert(filename, checksum);
        }
    }
    checksums
}

/// Returns the expected checksum for the specified platform binary.
///
/// `goos`/`goarch` should use Go's naming convention (e.g. `"darwin"`,
/// `"amd64"`).
pub fn get_expected_checksum(
    checksums: &HashMap<String, String>,
    version: &str,
    goos: &str,
    goarch: &str,
) -> Result<String, UpdaterError> {
    let filename = format!("termchess-{version}-{goos}-{goarch}");
    checksums
        .get(&filename)
        .cloned()
        .ok_or(UpdaterError::ChecksumNotFound(filename))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn sha256_hex(data: &[u8]) -> String {
        hex::encode(Sha256::digest(data))
    }

    #[test]
    fn test_verify_checksum() {
        let test_data = b"hello world";
        let correct = sha256_hex(test_data);

        struct Case {
            data: Vec<u8>,
            expected: String,
            want: bool,
        }
        let cases = [
            Case {
                data: test_data.to_vec(),
                expected: correct.clone(),
                want: true,
            },
            Case {
                data: test_data.to_vec(),
                expected: "B94D27B9934D3E08A52E52D7DA7DABFAC484EFE37A5380EE9088F7ACE2EFCDE9"
                    .to_string(),
                want: true,
            },
            Case {
                data: test_data.to_vec(),
                expected: "B94d27b9934D3e08A52e52d7Da7dAbfAc484Efe37a5380Ee9088f7Ace2efCde9"
                    .to_string(),
                want: true,
            },
            Case {
                data: test_data.to_vec(),
                expected: "0000000000000000000000000000000000000000000000000000000000000000"
                    .to_string(),
                want: false,
            },
            Case {
                data: b"different data".to_vec(),
                expected: correct.clone(),
                want: false,
            },
            Case {
                data: vec![],
                expected: correct.clone(),
                want: false,
            },
            Case {
                data: test_data.to_vec(),
                expected: String::new(),
                want: false,
            },
            Case {
                data: vec![],
                expected: String::new(),
                want: false,
            },
        ];

        for (i, c) in cases.iter().enumerate() {
            assert_eq!(verify_checksum(&c.data, &c.expected), c.want, "case {i}");
        }
    }

    #[test]
    fn test_verify_checksum_known_values() {
        // Empty data always returns false, even with the correct empty hash.
        assert!(!verify_checksum(
            b"",
            "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
        ));
        assert!(verify_checksum(
            b"hello world",
            "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
        ));
        assert!(verify_checksum(
            b"a",
            "ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb"
        ));
    }

    #[test]
    fn test_parse_checksums() {
        // Standard format with two spaces.
        let content = "abc123def456  termchess-v0.1.0-darwin-amd64\n\
                       def789ghi012  termchess-v0.1.0-darwin-arm64\n\
                       jkl345mno678  termchess-v0.1.0-linux-amd64";
        let got = parse_checksums(content);
        assert_eq!(got.len(), 3);
        assert_eq!(got["termchess-v0.1.0-darwin-amd64"], "abc123def456");
        assert_eq!(got["termchess-v0.1.0-darwin-arm64"], "def789ghi012");
        assert_eq!(got["termchess-v0.1.0-linux-amd64"], "jkl345mno678");

        // Single space.
        let content = "abc123def456 termchess-v0.1.0-darwin-amd64\n\
                       def789ghi012 termchess-v0.1.0-darwin-arm64";
        let got = parse_checksums(content);
        assert_eq!(got.len(), 2);
        assert_eq!(got["termchess-v0.1.0-darwin-amd64"], "abc123def456");

        // Empty content.
        assert_eq!(parse_checksums("").len(), 0);

        // Content with empty lines.
        let content = "abc123def456  termchess-v0.1.0-darwin-amd64\n\n\
                       def789ghi012  termchess-v0.1.0-darwin-arm64\n\n";
        let got = parse_checksums(content);
        assert_eq!(got.len(), 2);

        // Content with whitespace padding.
        let content = "  abc123def456  termchess-v0.1.0-darwin-amd64\n\
                         def789ghi012  termchess-v0.1.0-darwin-arm64  ";
        let got = parse_checksums(content);
        assert_eq!(got.len(), 2);
        assert_eq!(got["termchess-v0.1.0-darwin-amd64"], "abc123def456");
        assert_eq!(got["termchess-v0.1.0-darwin-arm64"], "def789ghi012");
    }

    #[test]
    fn test_get_expected_checksum() {
        let mut checksums = HashMap::new();
        checksums.insert(
            "termchess-v0.1.0-darwin-amd64".to_string(),
            "abc123".to_string(),
        );
        checksums.insert(
            "termchess-v0.1.0-darwin-arm64".to_string(),
            "def456".to_string(),
        );
        checksums.insert(
            "termchess-v0.1.0-linux-amd64".to_string(),
            "ghi789".to_string(),
        );
        checksums.insert(
            "termchess-v0.1.0-linux-arm64".to_string(),
            "jkl012".to_string(),
        );

        assert_eq!(
            get_expected_checksum(&checksums, "v0.1.0", "darwin", "amd64").unwrap(),
            "abc123"
        );
        assert_eq!(
            get_expected_checksum(&checksums, "v0.1.0", "darwin", "arm64").unwrap(),
            "def456"
        );
        assert_eq!(
            get_expected_checksum(&checksums, "v0.1.0", "linux", "amd64").unwrap(),
            "ghi789"
        );
        assert_eq!(
            get_expected_checksum(&checksums, "v0.1.0", "linux", "arm64").unwrap(),
            "jkl012"
        );

        // Missing platform.
        assert!(get_expected_checksum(&checksums, "v0.1.0", "windows", "amd64").is_err());

        // Empty checksums.
        let empty = HashMap::new();
        assert!(get_expected_checksum(&empty, "v0.1.0", "darwin", "amd64").is_err());
    }
}
