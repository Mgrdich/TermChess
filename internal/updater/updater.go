// Package updater provides functionality to check for updates and self-upgrade TermChess.
package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
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
