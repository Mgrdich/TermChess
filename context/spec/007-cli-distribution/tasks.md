# Tasks: CLI Distribution

---

## Slice 1: Version Display (`--version` flag)
*After this slice: `termchess --version` displays version info*

- [x] **1.1** Create `internal/version/version.go` with `Version`, `BuildDate`, `GitCommit` variables (default to "dev"/"unknown")
- [x] **1.2** Update `Makefile` to inject version via ldflags when building
- [x] **1.3** Add `flag` parsing to `cmd/termchess/main.go` for `--version` flag
- [x] **1.4** Implement `printVersion()` function that displays version, build date, and commit
- [x] **1.5** Verify: `make build && ./bin/termchess --version` shows "dev" version

---

## Slice 2: Build System & Release Automation
*After this slice: Merging to main with VERSION file change automatically creates a GitHub Release*

- [x] **2.1** Add `build-all` target to `Makefile` for cross-compilation (darwin/linux × amd64/arm64)
- [x] **2.2** Add `checksums` target to `Makefile` to generate `checksums.txt` with SHA256 hashes
- [x] **2.3** Create `VERSION` file at repository root
- [x] **2.4** Create `.github/workflows/release.yml` that triggers on VERSION file changes to main
- [x] **2.5** Configure release workflow to auto-create git tag, build all platforms, generate checksums, and create GitHub Release
- [x] **2.6** Test: Merge PR with VERSION change and verify release is created automatically

---

## Slice 3: Update Check Infrastructure
*After this slice: Internal functions ready for checking updates (not user-facing yet)*

- [ ] **3.1** Create `internal/updater/updater.go` with package structure
- [ ] **3.2** Implement `CheckLatestVersion(ctx) (string, error)` using GitHub API
- [ ] **3.3** Implement `GetAssetURL(version, os, arch string) string` to construct download URLs
- [ ] **3.4** Implement `VerifyChecksum(data []byte, expected string) bool` for SHA256 verification
- [ ] **3.5** Implement `detectInstallMethod() string` to identify go-install vs install-script
- [ ] **3.6** Create `internal/updater/updater_test.go` with unit tests for all functions
- [ ] **3.7** Verify: `make test` passes with new updater tests

---

## Slice 4: Async Update Notification in UI
*After this slice: Users see orange "Update available" message on startup if newer version exists*

- [ ] **4.1** Add `updateAvailable string` field to `Model` in `internal/ui/model.go`
- [ ] **4.2** Add `SetUpdateChannel(chan string)` method to `Model`
- [ ] **4.3** Implement async update check goroutine in `main.go` that sends to channel
- [ ] **4.4** Listen for update channel in TUI and set `updateAvailable` field
- [ ] **4.5** Display orange notification in main menu view when `updateAvailable != ""`
- [ ] **4.6** Ensure silent failure: no error shown if network fails or timeout occurs
- [ ] **4.7** Verify: Build old version, create newer release, run old version → orange notification appears

---

## Slice 5: Self-Upgrade Command (`--upgrade`)
*After this slice: `termchess --upgrade` downloads and replaces the binary*

- [ ] **5.1** Add `--upgrade` flag parsing to `main.go`
- [ ] **5.2** Implement `handleUpgrade(args []string)` function in main
- [ ] **5.3** Check for `go install` users → display redirect message and exit
- [ ] **5.4** Implement `Upgrade(targetVersion string) error` in updater package
- [ ] **5.5** Implement download logic: fetch binary from GitHub Releases
- [ ] **5.6** Implement checksum verification before replacement
- [ ] **5.7** Implement atomic binary replacement (rename dance: old → .old, new → current, delete .old)
- [ ] **5.8** Handle "already up to date" case
- [ ] **5.9** Add downgrade warning: "It might be buggier than a summer porch. Continue? [y/N]"
- [ ] **5.10** Display success message with old → new version
- [ ] **5.11** Handle permission errors with helpful message (suggest sudo)
- [ ] **5.12** Verify: Install via curl script, run `--upgrade`, confirm version changes

---

## Slice 6: Self-Uninstall Command (`--uninstall`)
*After this slice: `termchess --uninstall` removes binary and config*

- [ ] **6.1** Export `GetConfigDir()` function in `internal/config/paths.go`
- [ ] **6.2** Add `--uninstall` flag parsing to `main.go`
- [ ] **6.3** Implement `handleUninstall()` function in main
- [ ] **6.4** Prompt for confirmation: "Are you sure you want to uninstall TermChess? [y/N]"
- [ ] **6.5** Implement `Uninstall() error` in updater package
- [ ] **6.6** Remove binary file (via `os.Executable()` path)
- [ ] **6.7** Remove config directory (`~/.termchess/`)
- [ ] **6.8** Display farewell message: "✓ TermChess has been uninstalled. Goodbye!"
- [ ] **6.9** Verify: Install, create config, run `--uninstall`, confirm both binary and config are gone

---

## Slice 7: Install Script
*After this slice: Users can install via `curl ... | bash`*

- [ ] **7.1** Create `scripts/install.sh` with shebang and error handling (`set -euo pipefail`)
- [ ] **7.2** Implement OS detection (`uname -s` → darwin/linux)
- [ ] **7.3** Implement architecture detection (`uname -m` → amd64/arm64)
- [ ] **7.4** Implement unsupported platform error messages
- [ ] **7.5** Implement existing installation detection and version comparison
- [ ] **7.6** Implement upgrade prompt for existing installations
- [ ] **7.7** Implement binary download from GitHub Releases
- [ ] **7.8** Implement checksum verification
- [ ] **7.9** Implement install to `~/.local/bin` (create dir if needed)
- [ ] **7.10** Implement fallback to `/usr/local/bin` with sudo prompt
- [ ] **7.11** Implement PATH check with shell-specific instructions (bash/zsh)
- [ ] **7.12** Implement specific version install: `bash -s -- v1.0.0`
- [ ] **7.13** Display success message with installed version
- [ ] **7.14** Verify: Test fresh install on macOS and Linux (Docker)

---

## Slice 8: Documentation
*After this slice: README has complete installation instructions*

- [ ] **8.1** Add "Installation" section to README with curl one-liner
- [ ] **8.2** Add specific version install example
- [ ] **8.3** Add manual download instructions (for users who don't trust piping to bash)
- [ ] **8.4** Document `--upgrade` command usage
- [ ] **8.5** Document `--uninstall` command usage
- [ ] **8.6** Document `--version` command
- [ ] **8.7** Verify: README is clear and all commands work as documented
