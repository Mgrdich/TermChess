# Technical Specification: CLI Distribution

- **Functional Specification:** `context/spec/007-cli-distribution/functional-spec.md`
- **Status:** Draft
- **Author(s):** Claude

---

## 1. High-Level Technical Approach

This feature enables users to install, upgrade, and uninstall TermChess via simple CLI commands and a curl-based install script. The implementation involves:

1. **New `internal/version` package** — Holds version/build info injected via ldflags at compile time from git tags
2. **New `internal/updater` package** — Contains self-upgrade, self-uninstall, and update-check logic using GitHub API
3. **Modified `cmd/termchess/main.go`** — Add flag parsing for `--upgrade`, `--uninstall`, `--version`; start async update check before TUI
4. **Modified UI layer** — Display orange update notification in main menu when new version available
5. **New `scripts/install.sh`** — Curl install script hosted in repo
6. **New `.github/workflows/release.yml`** — Automated release on git tag push
7. **Updated `Makefile`** — Add cross-compilation targets and ldflags for version injection

**Version Source of Truth:** Git tags. When a developer pushes a tag (e.g., `v1.0.0`), GitHub Actions automatically builds and releases binaries with the version embedded via ldflags.

---

## 2. Proposed Solution & Implementation Plan (The "How")

### 2.1 Version Package

**New File:** `internal/version/version.go`

```go
package version

// Set via ldflags at build time. Defaults to "dev" for local builds.
var (
    Version   = "dev"
    BuildDate = "unknown"
    GitCommit = "unknown"
)
```

**Build with ldflags:**
```bash
go build -ldflags="-X github.com/Mgrdich/TermChess/internal/version.Version=v1.0.0 \
  -X github.com/Mgrdich/TermChess/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -X github.com/Mgrdich/TermChess/internal/version.GitCommit=$(git rev-parse --short HEAD)" \
  -o bin/termchess ./cmd/termchess
```

---

### 2.2 Updater Package

**New File:** `internal/updater/updater.go`

| Function | Purpose |
|----------|---------|
| `CheckLatestVersion(ctx) (string, error)` | Queries GitHub API for latest release tag |
| `Upgrade(targetVersion string) error` | Downloads binary, verifies checksum, replaces self |
| `Uninstall() error` | Removes binary and `~/.termchess/` directory |
| `GetAssetURL(version, os, arch string) string` | Constructs download URL for specific platform |
| `VerifyChecksum(binary []byte, expected string) bool` | SHA256 verification |
| `detectInstallMethod() string` | Detects if installed via `go install` or install script |

**GitHub API Endpoints:**
```
GET https://api.github.com/repos/Mgrdich/TermChess/releases/latest
GET https://api.github.com/repos/Mgrdich/TermChess/releases/tags/v1.0.0
```

**Install Method Detection:**
```go
func detectInstallMethod() string {
    execPath, err := os.Executable()
    if err != nil {
        return "unknown"
    }
    realPath, _ := filepath.EvalSymlinks(execPath)

    // Check for go install paths
    if strings.Contains(realPath, "/go/bin/") {
        return "go-install"
    }

    // Check common install script locations
    if strings.Contains(realPath, "/.local/bin/") ||
       strings.Contains(realPath, "/usr/local/bin/") {
        return "install-script"
    }

    return "unknown"
}
```

**Self-replacement Strategy (Rename Dance):**
1. Download new binary to temp file
2. Verify checksum against `checksums.txt`
3. Rename current binary to `termchess.old`
4. Move new binary to original path
5. Delete `termchess.old`
6. Exit with success message

---

### 2.3 Entry Point Changes

**Modified File:** `cmd/termchess/main.go`

```go
func main() {
    // 1. Parse flags
    showVersion := flag.Bool("version", false, "Show version")
    upgrade := flag.Bool("upgrade", false, "Upgrade to latest (or specify version)")
    uninstall := flag.Bool("uninstall", false, "Uninstall TermChess")
    flag.Parse()

    // 2. Handle non-TUI commands (exit after)
    if *showVersion { printVersion(); return }
    if *upgrade { handleUpgrade(flag.Args()); return }
    if *uninstall { handleUninstall(); return }

    // 3. Async update check
    updateChan := make(chan string, 1)
    go asyncUpdateCheck(updateChan)

    // 4. Start TUI with update notification channel
    cfg := config.LoadConfig()
    model := ui.NewModel(cfg)
    model.SetUpdateChannel(updateChan)
    // ... run TUI
}
```

---

### 2.4 Upgrade Command Behavior

**Usage:**
```bash
termchess --upgrade          # Latest version
termchess --upgrade v1.2.0   # Specific version (upgrade or downgrade)
```

**Behavior by Install Method:**

| Install Method | `--upgrade` Behavior |
|----------------|---------------------|
| `go install` | Print redirect message, exit |
| Install script (`~/.local/bin`, `/usr/local/bin`) | Perform upgrade |
| Unknown | Attempt upgrade, fail gracefully if issues |

**Redirect Message for `go install` Users:**
```
You installed TermChess via 'go install'.

To upgrade, run:
  go install github.com/Mgrdich/TermChess/cmd/termchess@latest

Or switch to our install script for automatic upgrades:
  curl -fsSL https://raw.githubusercontent.com/Mgrdich/TermChess/main/scripts/install.sh | bash
```

**Upgrading Output:**
```
$ termchess --upgrade

Current version: v1.1.0
Latest version:  v1.2.0

Downloading termchess-v1.2.0-darwin-arm64...
Verifying checksum... ✓
Installing... ✓

✓ TermChess upgraded from v1.1.0 to v1.2.0
```

**Downgrading Output (with warning):**
```
$ termchess --upgrade v1.0.0

Current version: v1.2.0
Target version:  v1.0.0

⚠ v1.0.0 is older than your current version. It might be buggier than a summer porch. Continue? [y/N] y

Downloading termchess-v1.0.0-darwin-arm64...
Verifying checksum... ✓
Installing... ✓

✓ TermChess switched from v1.2.0 to v1.0.0
```

**Already on Target:**
```
$ termchess --upgrade

Already up to date (v1.2.0)
```

---

### 2.5 UI Update Notification

**Modified Files:** `internal/ui/model.go`, `internal/ui/view.go`

- Add `updateAvailable string` field to `Model`
- Add `SetUpdateChannel(chan string)` method
- On main menu render, if `updateAvailable != ""`, display in orange:
  ```
  Update available: v1.3.0 (current: v1.2.0). Run 'termchess --upgrade' to update.
  ```
- Use lipgloss orange color: `lipgloss.Color("208")`
- Update check runs async, does not block TUI startup
- If network fails, fails silently (no error shown)

---

### 2.6 Automated Release Workflow

**Version Source of Truth:** `VERSION` file at repository root (e.g., contains `1.0.0`)

**New Files:**
- `VERSION` - Contains the current version number (e.g., `1.0.0`)
- `.github/workflows/release.yml` - Automated release workflow

```yaml
name: Release

on:
  push:
    branches: [main]
    paths:
      - 'VERSION'

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Get version from VERSION file
        id: version
        run: |
          VERSION=$(cat VERSION | tr -d '[:space:]')
          echo "VERSION=v$VERSION" >> $GITHUB_OUTPUT

      - name: Check if tag already exists
        id: check_tag
        run: |
          if git rev-parse "${{ steps.version.outputs.VERSION }}" >/dev/null 2>&1; then
            echo "exists=true" >> $GITHUB_OUTPUT
          else
            echo "exists=false" >> $GITHUB_OUTPUT
          fi

      - name: Create and push tag
        if: steps.check_tag.outputs.exists == 'false'
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git tag -a ${{ steps.version.outputs.VERSION }} -m "Release ${{ steps.version.outputs.VERSION }}"
          git push origin ${{ steps.version.outputs.VERSION }}

      - uses: actions/setup-go@v5
        if: steps.check_tag.outputs.exists == 'false'
        with:
          go-version: '1.24'

      - name: Install system dependencies
        if: steps.check_tag.outputs.exists == 'false'
        run: sudo apt-get update && sudo apt-get install -y libx11-dev xorg-dev

      - name: Build all platforms
        if: steps.check_tag.outputs.exists == 'false'
        run: make build-all VERSION=${{ steps.version.outputs.VERSION }}

      - name: Generate checksums
        if: steps.check_tag.outputs.exists == 'false'
        run: make checksums

      - name: Create GitHub Release
        if: steps.check_tag.outputs.exists == 'false'
        uses: softprops/action-gh-release@v2
        with:
          tag_name: ${{ steps.version.outputs.VERSION }}
          files: |
            dist/termchess-*
            dist/checksums.txt
          generate_release_notes: true
```

**Automated Release Flow:**
```
1. Developer updates VERSION file (e.g., "1.0.0" → "1.1.0")
    ↓
2. Creates PR and merges to main
    ↓
3. GitHub Actions detects VERSION file changed
    ↓
4. Reads version from file → "v1.1.0"
    ↓
5. Checks if tag exists (skip if already released)
    ↓
6. Creates git tag v1.1.0 automatically
    ↓
7. Builds 4 binaries (darwin/linux × amd64/arm64)
    ↓
8. Generates checksums.txt
    ↓
9. Creates GitHub Release with all assets
    ↓
Users can now install/upgrade to v1.1.0
```

**Key Benefits:**
- No manual tagging required
- Releases only happen from main branch
- Version is explicitly controlled in VERSION file
- Duplicate releases are prevented (tag existence check)

---

### 2.7 Makefile Updates

**Modified File:** `Makefile`

```makefile
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
MODULE := github.com/Mgrdich/TermChess
LDFLAGS := -s -w \
    -X $(MODULE)/internal/version.Version=$(VERSION) \
    -X $(MODULE)/internal/version.BuildDate=$(BUILD_DATE) \
    -X $(MODULE)/internal/version.GitCommit=$(GIT_COMMIT)

.PHONY: build build-all checksums clean test run

build:
	go build -ldflags="$(LDFLAGS)" -o bin/termchess ./cmd/termchess

build-all:
	@mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/termchess-$(VERSION)-darwin-amd64 ./cmd/termchess
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/termchess-$(VERSION)-darwin-arm64 ./cmd/termchess
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/termchess-$(VERSION)-linux-amd64 ./cmd/termchess
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/termchess-$(VERSION)-linux-arm64 ./cmd/termchess

checksums:
	cd dist && sha256sum termchess-* > checksums.txt

clean:
	rm -rf bin/ dist/

test:
	go test -v ./...

run:
	go run ./cmd/termchess
```

---

### 2.8 Install Script

**New File:** `scripts/install.sh`

**URL:** `https://raw.githubusercontent.com/Mgrdich/TermChess/main/scripts/install.sh`

**Invocation:**
```bash
curl -fsSL https://raw.githubusercontent.com/Mgrdich/TermChess/main/scripts/install.sh | bash
curl -fsSL https://raw.githubusercontent.com/Mgrdich/TermChess/main/scripts/install.sh | bash -s -- v1.1.0
```

**Script Responsibilities:**
1. Detect OS (`uname -s`) and arch (`uname -m`)
2. Map to Go naming (`darwin`/`linux`, `amd64`/`arm64`)
3. Check for existing installation, prompt for upgrade if found
4. Download binary and checksums from GitHub Releases
5. Verify SHA256 checksum
6. Install to `~/.local/bin` (create if needed) or `/usr/local/bin` with sudo
7. Check PATH and warn with shell-specific instructions if needed
8. Display success message

---

### 2.9 Config Path Export

**Modified File:** `internal/config/paths.go`

Export `GetConfigDir()` function for use by uninstall:

```go
// GetConfigDir returns the path to the TermChess config directory (~/.termchess/)
func GetConfigDir() (string, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", fmt.Errorf("failed to get home directory: %w", err)
    }
    return filepath.Join(homeDir, ".termchess"), nil
}
```

---

## 3. Impact and Risk Analysis

### System Dependencies

| Dependency | Impact |
|------------|--------|
| GitHub API | Update check, upgrade, install script all rely on GitHub API availability |
| GitHub Releases | Binary hosting; if unavailable, installs/upgrades fail |
| Network access | Required for update check (fails silently), upgrade, install |

### Potential Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| GitHub API rate limiting | Medium | Update checks fail | Use unauthenticated requests (60/hr limit sufficient); fail silently |
| Binary replacement fails mid-write | Low | Corrupted installation | Atomic replacement via rename dance: download → rename old → move new → delete old |
| User lacks write permission | Medium | Upgrade fails | Detect and suggest `sudo termchess --upgrade`; clear error message |
| Checksum mismatch | Low | Install/upgrade aborted | Verify before replacing; suggest re-download in error message |
| `~/.local/bin` not in PATH | High | Users can't run `termchess` | Detect and display shell-specific instructions (bash/zsh) |
| Running binary can't delete itself | Low (Unix) | Uninstall fails | On Unix, running binaries can be deleted; no issue |

---

## 4. Testing Strategy

### Unit Tests

| Component | Test Cases |
|-----------|------------|
| `version` package | Verify default values; verify ldflags injection works |
| `updater.CheckLatestVersion` | Mock GitHub API response; test timeout handling; test parse errors |
| `updater.VerifyChecksum` | Valid checksum passes; invalid fails; empty input handling |
| `updater.GetAssetURL` | Correct URL construction for all OS/arch combinations |
| `updater.detectInstallMethod` | Correctly identifies `go install` vs install script paths |

### Integration Tests

| Scenario | Approach |
|----------|----------|
| Install script | Test in Docker containers (ubuntu, alpine); verify binary works after install |
| `--upgrade` | Mock GitHub releases; verify binary replacement works |
| `--uninstall` | Verify binary and `~/.termchess/` are removed |
| Async update check | Verify TUI starts without delay; verify notification appears |

### Manual/E2E Tests

| Test | Steps |
|------|-------|
| Fresh install | Run curl install on clean macOS/Linux; verify `termchess` runs |
| Upgrade flow | Install old version, run `--upgrade`, verify new version |
| Downgrade flow | Install new version, run `--upgrade v1.0.0`, verify warning and switch |
| Version-specific install | `curl ... \| bash -s -- v1.0.0`; verify correct version |
| Uninstall | Run `--uninstall`, confirm prompt, verify cleanup |
| Update notification | Install old version, run `termchess`, verify orange message |
| `go install` detection | Install via `go install`, run `--upgrade`, verify redirect message |

### CI Testing

- Existing `ci.yml` continues to run `make build` and `make test`
- Release workflow tested by pushing test tags to a fork
- Add build verification for all 4 platform/arch combinations

---

## 5. Files Summary

### New Files

| File | Purpose |
|------|---------|
| `internal/version/version.go` | Version constants (ldflags injected) |
| `internal/updater/updater.go` | Upgrade, uninstall, update-check logic |
| `internal/updater/updater_test.go` | Unit tests for updater |
| `scripts/install.sh` | Curl install script |
| `.github/workflows/release.yml` | Automated release on tag push |

### Modified Files

| File | Changes |
|------|---------|
| `cmd/termchess/main.go` | Add flag parsing (`--version`, `--upgrade`, `--uninstall`), async update check |
| `internal/config/paths.go` | Export `GetConfigDir()` |
| `internal/ui/model.go` | Add `updateAvailable` field and `SetUpdateChannel()` |
| `internal/ui/view.go` | Display orange update notification on main menu |
| `Makefile` | Add ldflags, `build-all`, `checksums` targets |
| `README.md` | Installation instructions (curl, manual, upgrade, uninstall) |
