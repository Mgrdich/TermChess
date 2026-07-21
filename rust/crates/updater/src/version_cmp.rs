//! Semantic version comparison, ported from the Go `CompareVersions`,
//! `parseVersion`, and `parseUint` helpers.

/// Compares two semantic version strings.
///
/// Returns `-1` if `v1 < v2`, `0` if `v1 == v2`, `1` if `v1 > v2`.
/// Handles proper semver comparison (e.g. `"1.10.0" > "1.2.0"`).
/// Non-semver versions like `"dev"` are treated as less than any semver.
pub fn compare_versions(v1: &str, v2: &str) -> i32 {
    // Normalize versions by removing 'v' prefix if present.
    let v1 = v1.strip_prefix('v').unwrap_or(v1);
    let v2 = v2.strip_prefix('v').unwrap_or(v2);

    if v1 == v2 {
        return 0;
    }

    let (parts1, pre1) = parse_version(v1);
    let (parts2, pre2) = parse_version(v2);

    match (parts1, parts2) {
        // If either failed to parse (non-semver like "dev"), fall back to
        // string comparison.
        (None, None) => {
            if v1 < v2 {
                -1
            } else {
                1
            }
        }
        (None, Some(_)) => -1, // non-semver is less than semver
        (Some(_), None) => 1,  // semver is greater than non-semver
        (Some(p1), Some(p2)) => {
            // Compare major, minor, patch.
            for i in 0..3 {
                if p1[i] < p2[i] {
                    return -1;
                }
                if p1[i] > p2[i] {
                    return 1;
                }
            }

            // Compare prerelease (if present).
            // Version without prerelease > version with prerelease.
            match (pre1.is_empty(), pre2.is_empty()) {
                (false, true) => -1, // has prerelease < no prerelease
                (true, false) => 1,  // no prerelease > has prerelease
                _ => {
                    // Both have prerelease (or both empty) — compare
                    // lexicographically.
                    if pre1 < pre2 {
                        -1
                    } else if pre1 > pre2 {
                        1
                    } else {
                        0
                    }
                }
            }
        }
    }
}

/// Parses a semver string into `[major, minor, patch]` integers and a
/// prerelease string. Returns `None` for the parts if the string is not a valid
/// semver.
fn parse_version(v: &str) -> (Option<[i64; 3]>, String) {
    let mut v = v;
    let mut prerelease = String::new();

    // Handle prerelease suffix (e.g. "1.0.0-beta.1").
    if let Some(idx) = v.find('-') {
        let mut pre = &v[idx + 1..];
        // Strip build metadata from prerelease if present.
        if let Some(build_idx) = pre.find('+') {
            pre = &pre[..build_idx];
        }
        prerelease = pre.to_string();
        v = &v[..idx];
    } else if let Some(idx) = v.find('+') {
        // Build metadata without prerelease.
        v = &v[..idx];
    }

    let parts: Vec<&str> = v.split('.').collect();
    if parts.is_empty() || parts.len() > 3 {
        return (None, String::new());
    }

    let mut result = [0i64; 3];
    for (i, part) in parts.iter().enumerate() {
        match parse_uint(part) {
            Ok(n) => result[i] = n,
            Err(_) => return (None, String::new()),
        }
    }

    (Some(result), prerelease)
}

/// Parses a string as a non-negative integer, mirroring the Go helper.
fn parse_uint(s: &str) -> Result<i64, ()> {
    if s.is_empty() {
        return Err(());
    }
    let mut n: i64 = 0;
    for c in s.chars() {
        if !c.is_ascii_digit() {
            return Err(());
        }
        n = n * 10 + (c as i64 - '0' as i64);
    }
    Ok(n)
}

/// Removes a leading `v` prefix from a version string.
pub(crate) fn normalize_version(v: &str) -> String {
    v.strip_prefix('v').unwrap_or(v).to_string()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_compare_versions() {
        let cases: &[(&str, &str, i32)] = &[
            ("v1.0.0", "v1.0.0", 0),
            ("1.0.0", "1.0.0", 0),
            ("v1.0.0", "1.0.0", 0),
            ("v1.0.0", "v2.0.0", -1),
            ("v1.0.0", "v1.1.0", -1),
            ("v1.0.0", "v1.0.1", -1),
            ("v2.0.0", "v1.0.0", 1),
            ("v1.1.0", "v1.0.0", 1),
            ("v1.0.1", "v1.0.0", 1),
            ("v1.0.0-alpha", "v1.0.0-beta", -1),
        ];
        for (v1, v2, want) in cases {
            let got = compare_versions(v1, v2);
            assert_eq!(got, *want, "compare_versions({v1:?}, {v2:?})");
        }
    }

    #[test]
    fn test_compare_versions_semver_ordering() {
        // Documented behavior: "1.10.0" > "1.2.0".
        assert_eq!(compare_versions("1.10.0", "1.2.0"), 1);
        assert_eq!(compare_versions("1.2.0", "1.10.0"), -1);
    }

    #[test]
    fn test_compare_versions_non_semver() {
        // Non-semver is less than semver.
        assert_eq!(compare_versions("dev", "1.0.0"), -1);
        assert_eq!(compare_versions("1.0.0", "dev"), 1);
        // Both non-semver falls back to string comparison.
        assert_eq!(compare_versions("abc", "abd"), -1);
    }

    #[test]
    fn test_prerelease_vs_release() {
        // Release is greater than its prerelease.
        assert_eq!(compare_versions("1.0.0", "1.0.0-beta"), 1);
        assert_eq!(compare_versions("1.0.0-beta", "1.0.0"), -1);
    }

    #[test]
    fn test_normalize_version() {
        assert_eq!(normalize_version("v1.0.0"), "1.0.0");
        assert_eq!(normalize_version("1.0.0"), "1.0.0");
    }
}
