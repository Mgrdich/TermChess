# internal/updater

In-place binary upgrade via GitHub Releases. Powers `termchess --upgrade`.

## Flow

1. Query `https://api.github.com/repos/Mgrdich/TermChess/releases/latest` (or specified tag).
2. Download the platform-matched asset (darwin/linux × amd64/arm64).
3. Verify SHA-256 checksum against `checksums.txt`.
4. Atomically replace the running binary (rename-swap via `os.Rename`).

## Requirements

The upgrade path only works when the binary is under `~/.local/bin`, `/usr/local/bin`, or a similar writable install location. Go-install managed binaries should be upgraded via `go install`.
