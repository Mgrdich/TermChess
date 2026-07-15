//! Mapping from Rust's target constants to Go's `GOOS`/`GOARCH` naming, so the
//! release asset filenames match those produced by the Go build.

/// Returns the current OS in Go's `GOOS` convention (e.g. `"darwin"`,
/// `"linux"`).
pub fn current_goos() -> &'static str {
    match std::env::consts::OS {
        "macos" => "darwin",
        other => other,
    }
}

/// Returns the current architecture in Go's `GOARCH` convention (e.g.
/// `"amd64"`, `"arm64"`).
pub fn current_goarch() -> &'static str {
    match std::env::consts::ARCH {
        "x86_64" => "amd64",
        "aarch64" => "arm64",
        other => other,
    }
}
