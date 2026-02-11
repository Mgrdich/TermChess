# Functional Specification: CLI Distribution

- **Roadmap Item:** CLI Distribution (Phase 5) - Make the application accessible to users via simple command-line installation
- **Status:** Complete
- **Author:** Claude

---

## 1. Overview and Rationale (The "Why")

### Problem Statement
Currently, users must clone the repository and build TermChess from source using Go tooling. This creates a barrier for the target audience—developers and power users who expect CLI tools to install with a simple one-liner command, similar to tools like Homebrew packages or rustup.

### Desired Outcome
Users can install TermChess on macOS or Linux with a single `curl` command, manage upgrades through the application itself, and uninstall cleanly when needed. Users are also notified when updates are available.

### Success Metrics
- Users can install TermChess in under 30 seconds with a single command
- Upgrade and uninstall work reliably without manual intervention
- Installation succeeds without requiring sudo in the common case
- Users are informed of available updates without disrupting their experience

---

## 2. Functional Requirements (The "What")

### 2.1 Release Binary Builds

The project must produce standalone executable binaries for supported platforms.

**Supported Platforms:**
| OS | Architecture |
|-------|--------------|
| macOS | amd64, arm64 |
| Linux | amd64, arm64 |

**Binary Naming Convention:** `termchess-<version>-<os>-<arch>`
- Example: `termchess-v1.0.0-darwin-amd64`
- Example: `termchess-v1.0.0-darwin-arm64`
- Example: `termchess-v1.0.0-linux-amd64`
- Example: `termchess-v1.0.0-linux-arm64`

**Acceptance Criteria:**
- [ ] Binaries are compiled as standalone executables with no external dependencies
- [ ] Each release includes binaries for all four supported platform/architecture combinations
- [ ] A `checksums.txt` file is generated containing SHA256 hashes for all binaries

---

### 2.2 Hosted Download Endpoint

Release binaries must be hosted at a stable, publicly accessible URL.

**Release Process:**
1. Developer updates `VERSION` file with new version (e.g., `1.1.0`)
2. Creates PR and merges to main branch
3. GitHub Actions automatically creates git tag and GitHub Release
4. No manual tagging required - releases only happen from main

**Acceptance Criteria:**
- [ ] Binaries are published to GitHub Releases
- [ ] Each release is tagged with semantic versioning (e.g., `v1.0.0`)
- [ ] The `checksums.txt` file is included in each release
- [ ] A "latest" release is always identifiable for upgrade scripts
- [ ] Releases are automatically created when VERSION file changes on main

---

### 2.3 Curl Install Script

A shell script hosted at a stable URL that downloads and installs TermChess.

**Installation Flow:**

1. **Detect Platform:** Identify OS and architecture; abort if unsupported
2. **Check Existing Installation:**
   - If TermChess is already installed, display version comparison:
     ```
     TermChess is already installed (v1.1.0)
     Latest version available: v1.2.0
     Would you like to upgrade? [y/N]
     ```
   - If user declines, exit gracefully
3. **Download Binary:** Fetch the appropriate binary for the platform
4. **Verify Checksum:** Download `checksums.txt` and verify the binary's SHA256 hash
   - If verification fails: `Error: Checksum verification failed. Download may be corrupted.` → abort
5. **Install Binary:**
   - **First attempt:** Install to `~/.local/bin/termchess`
   - **Fallback:** If `~/.local/bin` doesn't exist or isn't writable, prompt for sudo and install to `/usr/local/bin/termchess`
6. **Check PATH:** If installed to `~/.local/bin` and it's not in PATH, display:
   ```
   ⚠ ~/.local/bin is not in your PATH. Add it by running:
     echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
     source ~/.zshrc
   ```
   (Detect shell: `.bashrc` for bash, `.zshrc` for zsh)
7. **Success Message:**
   ```
   ✓ TermChess v1.2.0 installed to ~/.local/bin/termchess
   Run 'termchess' to start playing!
   ```

**Version Selection:**
- Default: Install latest version
- Specific version: `curl ... | bash -s -- v1.1.0`

**Acceptance Criteria:**
- [ ] Script detects unsupported platforms and displays clear error message
- [ ] Script detects existing installation and shows version comparison
- [ ] Script prompts before upgrading an existing installation
- [ ] Script verifies checksum before installing; aborts on mismatch
- [ ] Script attempts user-local install (`~/.local/bin`) before requiring sudo
- [ ] Script warns if `~/.local/bin` is not in PATH with shell-specific instructions
- [ ] Script displays success message with installed version and usage hint
- [ ] Script accepts optional version argument for installing specific versions

**Error Messages:**
| Scenario | Message |
|----------|---------|
| Unsupported OS | `Error: Unsupported operating system. TermChess supports macOS and Linux.` |
| Unsupported arch | `Error: Unsupported architecture. TermChess supports amd64 and arm64 only.` |
| No internet | `Error: Failed to download TermChess. Check your internet connection.` |
| Checksum fail | `Error: Checksum verification failed. Download may be corrupted.` |
| Permission denied | `Error: Permission denied. Try running with sudo.` |

---

### 2.4 Automatic Update Check on Launch

When the application starts, it checks for available updates in the background.

**Behavior:**
1. On application launch, perform a non-blocking check against GitHub Releases for the latest version
2. If a newer version is available, display a notification in **orange color**:
   ```
   Update available: v1.3.0 (current: v1.2.0). Run 'termchess --upgrade' to update.
   ```
3. If no internet connection or the check fails, **fail silently** — do not display any error or warning
4. The update check must not delay application startup; the app should remain responsive

**Acceptance Criteria:**
- [ ] Application checks for updates on every launch
- [ ] Update notification is displayed in orange color when a new version exists
- [ ] Notification includes current version, available version, and upgrade command
- [ ] No error is shown if the update check fails (network issues, timeout, etc.)
- [ ] Update check runs asynchronously and does not block application startup

---

### 2.5 Built-in Upgrade Command

The TermChess binary includes a self-upgrade capability.

**Usage:**
- `termchess --upgrade` — Upgrade to the latest version
- `termchess --upgrade v1.3.0` — Upgrade to a specific version

**Behavior:**
1. Check current version against target version
2. If already on target version: `Already up to date (v1.2.0)`
3. Download new binary, verify checksum, replace current binary
4. Display: `✓ TermChess upgraded from v1.1.0 to v1.2.0`

**Acceptance Criteria:**
- [ ] `--upgrade` without argument upgrades to latest version
- [ ] `--upgrade <version>` upgrades to specified version
- [ ] Displays "Already up to date" if current version matches target
- [ ] Verifies checksum before replacing binary
- [ ] Shows clear success message with old and new versions

---

### 2.6 Built-in Uninstall Command

The TermChess binary includes a self-uninstall capability.

**Usage:** `termchess --uninstall`

**Behavior:**
1. Prompt for confirmation: `Are you sure you want to uninstall TermChess? [y/N]`
2. If confirmed:
   - Remove the binary from its installed location
   - Remove configuration directory (`~/.config/termchess` or equivalent)
3. Display: `✓ TermChess has been uninstalled. Goodbye!`

**Acceptance Criteria:**
- [ ] Prompts for confirmation before uninstalling
- [ ] Removes the binary file
- [ ] Removes the configuration directory and all saved data
- [ ] Displays farewell message on successful uninstall
- [ ] Exits gracefully if user declines confirmation

---

### 2.7 Installation Instructions

The README must document the installation process.

**Required Content:**
1. **One-liner install command** (curl pipe to bash)
2. **Specific version install** example
3. **Manual download instructions** for users who prefer not to pipe to bash
4. **Upgrade instructions** using `--upgrade`
5. **Uninstall instructions** using `--uninstall`

**Acceptance Criteria:**
- [ ] README includes copy-paste curl command for installation
- [ ] README shows how to install a specific version
- [ ] README includes manual download steps (download binary, chmod +x, move to PATH)
- [ ] README documents `--upgrade` and `--uninstall` commands

---

## 3. Scope and Boundaries

### In-Scope
- Binary builds for macOS (amd64, arm64) and Linux (amd64, arm64)
- GitHub Releases hosting with checksums
- Curl install script with upgrade detection
- Automatic update check on application launch (silent fail on no network)
- Built-in `--upgrade` command (latest or specific version)
- Built-in `--uninstall` command with config cleanup
- README documentation for all installation methods

### Out-of-Scope
- Windows support
- Homebrew formula or other package manager integration
- **Phase 6: Custom RL Agent** (separate roadmap item)
- **Phase 6: UCI Engine Integration** (separate roadmap item)
