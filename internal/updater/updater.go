// Package updater provides functionality to check for updates and self-upgrade TermChess.
package updater

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Mgrdich/TermChess/internal/config"
)

const (
	repoOwner = "Mgrdich"
	repoName  = "TermChess"
	githubAPI = "https://api.github.com"
)

// githubRelease represents the relevant fields from GitHub's release API response.
type githubRelease struct {
	TagName string `json:"tag_name"`
}

// Client provides methods for checking and downloading updates.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new updater client with default settings.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: githubAPI,
	}
}

// NewClientWithHTTPClient creates a new updater client with a custom HTTP client.
// This is useful for testing with mock servers.
func NewClientWithHTTPClient(httpClient *http.Client, baseURL string) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// CheckLatestVersion queries the GitHub API to get the latest release version.
// It returns the version tag (e.g., "v0.1.0") or an error if the request fails.
func (c *Client) CheckLatestVersion(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", c.baseURL, repoOwner, repoName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "TermChess-Updater")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if release.TagName == "" {
		return "", fmt.Errorf("empty tag_name in response")
	}

	return release.TagName, nil
}

// GetAssetURL constructs the download URL for a specific platform binary.
// The version should include the 'v' prefix (e.g., "v0.1.0").
// OS values: "darwin", "linux"
// Arch values: "amd64", "arm64"
func GetAssetURL(version, os, arch string) string {
	binaryName := fmt.Sprintf("termchess-%s-%s-%s", version, os, arch)
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s",
		repoOwner, repoName, version, binaryName)
}

// VerifyChecksum verifies that the SHA256 hash of data matches the expected hex string.
// Returns true if the checksum matches, false otherwise.
func VerifyChecksum(data []byte, expected string) bool {
	if len(data) == 0 || expected == "" {
		return false
	}

	hash := sha256.Sum256(data)
	actual := hex.EncodeToString(hash[:])

	return strings.EqualFold(actual, expected)
}

// InstallMethod represents how TermChess was installed.
type InstallMethod string

const (
	InstallMethodGoInstall     InstallMethod = "go-install"
	InstallMethodInstallScript InstallMethod = "install-script"
	InstallMethodUnknown       InstallMethod = "unknown"
)

// DetectInstallMethod identifies how TermChess was installed by examining the executable path.
func DetectInstallMethod() InstallMethod {
	execPath, err := os.Executable()
	if err != nil {
		return InstallMethodUnknown
	}

	realPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		realPath = execPath
	}

	// Check for go install paths
	if strings.Contains(realPath, "/go/bin/") {
		return InstallMethodGoInstall
	}

	// Check common install script locations
	if strings.Contains(realPath, "/.local/bin/") ||
		strings.Contains(realPath, "/usr/local/bin/") {
		return InstallMethodInstallScript
	}

	return InstallMethodUnknown
}

// String returns the string representation of the install method.
func (m InstallMethod) String() string {
	return string(m)
}

// ErrAlreadyUpToDate is returned when the current version matches the target version.
var ErrAlreadyUpToDate = errors.New("already up to date")

// ErrChecksumMismatch is returned when the downloaded binary's checksum doesn't match.
var ErrChecksumMismatch = errors.New("checksum mismatch")

// ErrPermissionDenied is returned when the upgrade fails due to permission issues.
var ErrPermissionDenied = errors.New("permission denied")

// UpgradeResult contains information about a completed upgrade.
type UpgradeResult struct {
	PreviousVersion string
	NewVersion      string
	IsDowngrade     bool
}

// GetChecksumsURL constructs the download URL for the checksums file.
func GetChecksumsURL(version string) string {
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/checksums.txt",
		repoOwner, repoName, version)
}

// DownloadBinary downloads the binary for the specified version and current platform.
func (c *Client) DownloadBinary(ctx context.Context, version string) ([]byte, error) {
	url := GetAssetURL(version, runtime.GOOS, runtime.GOARCH)
	return c.downloadFile(ctx, url)
}

// DownloadChecksums downloads and parses the checksums.txt file for a release.
// Returns a map of filename to checksum.
func (c *Client) DownloadChecksums(ctx context.Context, version string) (map[string]string, error) {
	url := GetChecksumsURL(version)
	data, err := c.downloadFile(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("downloading checksums: %w", err)
	}

	return ParseChecksums(string(data)), nil
}

// ParseChecksums parses a checksums.txt file content into a map of filename to checksum.
// Expected format: "checksum  filename" (two spaces between checksum and filename).
func ParseChecksums(content string) map[string]string {
	checksums := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Format: "checksum  filename" or "checksum filename"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			checksum := parts[0]
			filename := parts[len(parts)-1] // Take last part as filename
			checksums[filename] = checksum
		}
	}
	return checksums
}

// GetExpectedChecksum returns the expected checksum for the specified platform binary.
// Use runtime.GOOS and runtime.GOARCH for the current platform.
func GetExpectedChecksum(checksums map[string]string, version, goos, goarch string) (string, error) {
	filename := fmt.Sprintf("termchess-%s-%s-%s", version, goos, goarch)
	checksum, ok := checksums[filename]
	if !ok {
		return "", fmt.Errorf("checksum not found for %s", filename)
	}
	return checksum, nil
}

// downloadFile performs an HTTP GET request and returns the response body.
func (c *Client) downloadFile(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "TermChess-Updater")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return data, nil
}

// CompareVersions compares two semantic version strings.
// Returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2.
// Handles proper semver comparison (e.g., "1.10.0" > "1.2.0").
// Non-semver versions like "dev" are treated as less than any semver.
func CompareVersions(v1, v2 string) int {
	// Normalize versions by removing 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	if v1 == v2 {
		return 0
	}

	// Parse versions into components
	parts1, pre1 := parseVersion(v1)
	parts2, pre2 := parseVersion(v2)

	// If either failed to parse (non-semver like "dev"), fall back to string comparison
	if parts1 == nil && parts2 == nil {
		if v1 < v2 {
			return -1
		}
		return 1
	}
	if parts1 == nil {
		return -1 // non-semver is less than semver
	}
	if parts2 == nil {
		return 1 // semver is greater than non-semver
	}

	// Compare major, minor, patch
	for i := 0; i < 3; i++ {
		if parts1[i] < parts2[i] {
			return -1
		}
		if parts1[i] > parts2[i] {
			return 1
		}
	}

	// Compare prerelease (if present)
	// Version without prerelease > version with prerelease
	if pre1 == "" && pre2 != "" {
		return 1 // no prerelease > has prerelease
	}
	if pre1 != "" && pre2 == "" {
		return -1 // has prerelease < no prerelease
	}
	// Both have prerelease - compare lexicographically
	if pre1 < pre2 {
		return -1
	}
	if pre1 > pre2 {
		return 1
	}

	return 0
}

// parseVersion parses a semver string into [major, minor, patch] integers and prerelease string.
// Returns nil for parts if the string is not a valid semver.
func parseVersion(v string) ([]int, string) {
	// Handle prerelease suffix (e.g., "1.0.0-beta.1")
	prerelease := ""
	if idx := strings.Index(v, "-"); idx != -1 {
		prerelease = v[idx+1:]
		// Strip build metadata from prerelease if present
		if buildIdx := strings.Index(prerelease, "+"); buildIdx != -1 {
			prerelease = prerelease[:buildIdx]
		}
		v = v[:idx]
	} else if idx := strings.Index(v, "+"); idx != -1 {
		// Build metadata without prerelease
		v = v[:idx]
	}

	parts := strings.Split(v, ".")
	if len(parts) < 1 || len(parts) > 3 {
		return nil, ""
	}

	result := make([]int, 3)
	for i, part := range parts {
		n, err := parseUint(part)
		if err != nil {
			return nil, ""
		}
		result[i] = n
	}

	return result, prerelease
}

// parseUint parses a string as a non-negative integer.
func parseUint(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("not a number")
		}
		n = n*10 + int(r-'0')
	}
	return n, nil
}

// Upgrade performs the upgrade to the specified version.
// If targetVersion is empty, it upgrades to the latest version.
func (c *Client) Upgrade(ctx context.Context, currentVersion, targetVersion string, confirmDowngrade func() bool) (*UpgradeResult, error) {
	// If no target version specified, get the latest
	if targetVersion == "" {
		latest, err := c.CheckLatestVersion(ctx)
		if err != nil {
			return nil, fmt.Errorf("checking latest version: %w", err)
		}
		targetVersion = latest
	}

	// Normalize versions for comparison
	normalizedCurrent := normalizeVersion(currentVersion)
	normalizedTarget := normalizeVersion(targetVersion)

	// Check if already up to date
	if normalizedCurrent == normalizedTarget {
		return nil, ErrAlreadyUpToDate
	}

	// Ensure target version has 'v' prefix for URLs
	if !strings.HasPrefix(targetVersion, "v") {
		targetVersion = "v" + targetVersion
	}

	// Check if this is a downgrade
	isDowngrade := CompareVersions(normalizedTarget, normalizedCurrent) < 0
	if isDowngrade && confirmDowngrade != nil && !confirmDowngrade() {
		return nil, fmt.Errorf("downgrade cancelled by user")
	}

	// Download checksums first
	checksums, err := c.DownloadChecksums(ctx, targetVersion)
	if err != nil {
		return nil, err // DownloadChecksums already wraps the error
	}

	// Get expected checksum
	expectedChecksum, err := GetExpectedChecksum(checksums, targetVersion, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return nil, err
	}

	// Download the binary
	binaryData, err := c.DownloadBinary(ctx, targetVersion)
	if err != nil {
		return nil, fmt.Errorf("downloading binary: %w", err)
	}

	// Verify checksum
	if !VerifyChecksum(binaryData, expectedChecksum) {
		return nil, ErrChecksumMismatch
	}

	// Replace the binary
	if err := ReplaceBinary(binaryData); err != nil {
		return nil, err
	}

	return &UpgradeResult{
		PreviousVersion: currentVersion,
		NewVersion:      targetVersion,
		IsDowngrade:     isDowngrade,
	}, nil
}

// normalizeVersion removes 'v' prefix and returns the version string.
func normalizeVersion(v string) string {
	return strings.TrimPrefix(v, "v")
}

// ReplaceBinary atomically replaces the current executable with new binary data.
// It preserves the original file's permissions.
func ReplaceBinary(newBinaryData []byte) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("getting executable path: %w", err)
	}

	realPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		realPath = execPath
	}

	// Get the current binary's permissions to preserve them
	fileInfo, err := os.Stat(realPath)
	if err != nil {
		return fmt.Errorf("getting file info: %w", err)
	}
	fileMode := fileInfo.Mode()

	// 1. Write new binary to temp file with same permissions as original
	tmpPath := realPath + ".new"
	if err := os.WriteFile(tmpPath, newBinaryData, fileMode); err != nil {
		if os.IsPermission(err) {
			return ErrPermissionDenied
		}
		return fmt.Errorf("writing new binary: %w", err)
	}

	// 2. Rename current to .old
	oldPath := realPath + ".old"
	if err := os.Rename(realPath, oldPath); err != nil {
		// Clean up temp file
		os.Remove(tmpPath)
		if os.IsPermission(err) {
			return ErrPermissionDenied
		}
		return fmt.Errorf("backing up current binary: %w", err)
	}

	// 3. Rename new to current
	if err := os.Rename(tmpPath, realPath); err != nil {
		// Try to restore old binary and clean up temp file
		os.Rename(oldPath, realPath)
		os.Remove(tmpPath) // Clean up temp file on rollback
		if os.IsPermission(err) {
			return ErrPermissionDenied
		}
		return fmt.Errorf("installing new binary: %w", err)
	}

	// 4. Delete old (best effort, don't fail if this doesn't work)
	os.Remove(oldPath)

	return nil
}

// GetGoInstallMessage returns the message to show users who installed via go install.
func GetGoInstallMessage() string {
	return `You installed TermChess via 'go install'.

To upgrade, run:
  go install github.com/Mgrdich/TermChess/cmd/termchess@latest

Or switch to our install script for automatic upgrades:
  curl -fsSL https://raw.githubusercontent.com/Mgrdich/TermChess/main/scripts/install.sh | bash`
}

// GetBinaryFilename returns the binary filename for the given version and platform.
func GetBinaryFilename(version, goos, goarch string) string {
	return fmt.Sprintf("termchess-%s-%s-%s", version, goos, goarch)
}

// Uninstall removes the TermChess binary and configuration directory.
// It returns an error if any removal operation fails.
func Uninstall() error {
	// Get executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("getting executable path: %w", err)
	}

	// Resolve symlinks to get the real path
	realPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		realPath = execPath
	}

	// Get config directory
	configDir, err := config.GetConfigDir()
	if err != nil {
		return fmt.Errorf("getting config directory: %w", err)
	}

	// Remove the binary
	if err := os.Remove(realPath); err != nil {
		if os.IsPermission(err) {
			return ErrPermissionDenied
		}
		return fmt.Errorf("removing binary: %w", err)
	}

	// Remove config directory recursively
	if err := os.RemoveAll(configDir); err != nil {
		if os.IsPermission(err) {
			return ErrPermissionDenied
		}
		return fmt.Errorf("removing config directory: %w", err)
	}

	return nil
}
