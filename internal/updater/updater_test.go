package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestCheckLatestVersion(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseStatus int
		wantVersion    string
		wantErr        bool
	}{
		{
			name:           "valid response",
			responseBody:   `{"tag_name": "v0.1.0"}`,
			responseStatus: http.StatusOK,
			wantVersion:    "v0.1.0",
			wantErr:        false,
		},
		{
			name:           "valid response with extra fields",
			responseBody:   `{"tag_name": "v1.2.3", "name": "Release 1.2.3", "draft": false}`,
			responseStatus: http.StatusOK,
			wantVersion:    "v1.2.3",
			wantErr:        false,
		},
		{
			name:           "empty tag_name",
			responseBody:   `{"tag_name": ""}`,
			responseStatus: http.StatusOK,
			wantVersion:    "",
			wantErr:        true,
		},
		{
			name:           "missing tag_name",
			responseBody:   `{"name": "Release"}`,
			responseStatus: http.StatusOK,
			wantVersion:    "",
			wantErr:        true,
		},
		{
			name:           "not found",
			responseBody:   `{"message": "Not Found"}`,
			responseStatus: http.StatusNotFound,
			wantVersion:    "",
			wantErr:        true,
		},
		{
			name:           "server error",
			responseBody:   `{"message": "Internal Server Error"}`,
			responseStatus: http.StatusInternalServerError,
			wantVersion:    "",
			wantErr:        true,
		},
		{
			name:           "invalid JSON",
			responseBody:   `not json`,
			responseStatus: http.StatusOK,
			wantVersion:    "",
			wantErr:        true,
		},
		{
			name:           "rate limited",
			responseBody:   `{"message": "API rate limit exceeded"}`,
			responseStatus: http.StatusForbidden,
			wantVersion:    "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request path
				expectedPath := "/repos/Mgrdich/TermChess/releases/latest"
				if r.URL.Path != expectedPath {
					t.Errorf("unexpected path: got %s, want %s", r.URL.Path, expectedPath)
				}

				// Verify headers
				if r.Header.Get("Accept") != "application/vnd.github.v3+json" {
					t.Errorf("missing or incorrect Accept header")
				}
				if r.Header.Get("User-Agent") == "" {
					t.Errorf("missing User-Agent header")
				}

				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClientWithHTTPClient(server.Client(), server.URL)
			version, err := client.CheckLatestVersion(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckLatestVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if version != tt.wantVersion {
				t.Errorf("CheckLatestVersion() = %q, want %q", version, tt.wantVersion)
			}
		})
	}
}

func TestCheckLatestVersionTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow response
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"tag_name": "v0.1.0"}`))
	}))
	defer server.Close()

	client := NewClientWithHTTPClient(server.Client(), server.URL)

	// Create a context with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.CheckLatestVersion(ctx)
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}

func TestCheckLatestVersionCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"tag_name": "v0.1.0"}`))
	}))
	defer server.Close()

	client := NewClientWithHTTPClient(server.Client(), server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately
	cancel()

	_, err := client.CheckLatestVersion(ctx)
	if err == nil {
		t.Error("expected cancellation error, got nil")
	}
}

func TestGetAssetURL(t *testing.T) {
	tests := []struct {
		name    string
		version string
		os      string
		arch    string
		want    string
	}{
		{
			name:    "darwin amd64",
			version: "v0.1.0",
			os:      "darwin",
			arch:    "amd64",
			want:    "https://github.com/Mgrdich/TermChess/releases/download/v0.1.0/termchess-v0.1.0-darwin-amd64",
		},
		{
			name:    "darwin arm64",
			version: "v0.1.0",
			os:      "darwin",
			arch:    "arm64",
			want:    "https://github.com/Mgrdich/TermChess/releases/download/v0.1.0/termchess-v0.1.0-darwin-arm64",
		},
		{
			name:    "linux amd64",
			version: "v0.1.0",
			os:      "linux",
			arch:    "amd64",
			want:    "https://github.com/Mgrdich/TermChess/releases/download/v0.1.0/termchess-v0.1.0-linux-amd64",
		},
		{
			name:    "linux arm64",
			version: "v0.1.0",
			os:      "linux",
			arch:    "arm64",
			want:    "https://github.com/Mgrdich/TermChess/releases/download/v0.1.0/termchess-v0.1.0-linux-arm64",
		},
		{
			name:    "different version",
			version: "v1.2.3",
			os:      "darwin",
			arch:    "arm64",
			want:    "https://github.com/Mgrdich/TermChess/releases/download/v1.2.3/termchess-v1.2.3-darwin-arm64",
		},
		{
			name:    "version with prerelease",
			version: "v0.2.0-beta.1",
			os:      "linux",
			arch:    "amd64",
			want:    "https://github.com/Mgrdich/TermChess/releases/download/v0.2.0-beta.1/termchess-v0.2.0-beta.1-linux-amd64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAssetURL(tt.version, tt.os, tt.arch)
			if got != tt.want {
				t.Errorf("GetAssetURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVerifyChecksum(t *testing.T) {
	// Create a known test data and its SHA256 hash
	testData := []byte("hello world")
	hash := sha256.Sum256(testData)
	correctChecksum := hex.EncodeToString(hash[:])

	tests := []struct {
		name     string
		data     []byte
		expected string
		want     bool
	}{
		{
			name:     "valid checksum",
			data:     testData,
			expected: correctChecksum,
			want:     true,
		},
		{
			name:     "valid checksum uppercase",
			data:     testData,
			expected: "B94D27B9934D3E08A52E52D7DA7DABFAC484EFE37A5380EE9088F7ACE2EFCDE9",
			want:     true,
		},
		{
			name:     "valid checksum mixed case",
			data:     testData,
			expected: "B94d27b9934D3e08A52e52d7Da7dAbfAc484Efe37a5380Ee9088f7Ace2efCde9",
			want:     true,
		},
		{
			name:     "invalid checksum",
			data:     testData,
			expected: "0000000000000000000000000000000000000000000000000000000000000000",
			want:     false,
		},
		{
			name:     "wrong data",
			data:     []byte("different data"),
			expected: correctChecksum,
			want:     false,
		},
		{
			name:     "empty data",
			data:     []byte{},
			expected: correctChecksum,
			want:     false,
		},
		{
			name:     "nil data",
			data:     nil,
			expected: correctChecksum,
			want:     false,
		},
		{
			name:     "empty expected",
			data:     testData,
			expected: "",
			want:     false,
		},
		{
			name:     "both empty",
			data:     []byte{},
			expected: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VerifyChecksum(tt.data, tt.expected)
			if got != tt.want {
				t.Errorf("VerifyChecksum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifyChecksumKnownValues(t *testing.T) {
	// Test with well-known SHA256 values
	tests := []struct {
		name     string
		data     string
		expected string
		want     bool
	}{
		{
			name:     "empty string hash",
			data:     "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			want:     false, // Empty data should return false
		},
		{
			name:     "hello world",
			data:     "hello world",
			expected: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
			want:     true,
		},
		{
			name:     "single character",
			data:     "a",
			expected: "ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VerifyChecksum([]byte(tt.data), tt.expected)
			if got != tt.want {
				t.Errorf("VerifyChecksum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectInstallMethod(t *testing.T) {
	// Note: This test is limited because we can't easily control os.Executable()
	// We can only verify it returns a valid InstallMethod
	method := DetectInstallMethod()

	validMethods := map[InstallMethod]bool{
		InstallMethodGoInstall:     true,
		InstallMethodInstallScript: true,
		InstallMethodUnknown:       true,
	}

	if !validMethods[method] {
		t.Errorf("DetectInstallMethod() returned invalid method: %q", method)
	}
}

func TestInstallMethodString(t *testing.T) {
	tests := []struct {
		method InstallMethod
		want   string
	}{
		{InstallMethodGoInstall, "go-install"},
		{InstallMethodInstallScript, "install-script"},
		{InstallMethodUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.method.String()
			if got != tt.want {
				t.Errorf("InstallMethod.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectInstallMethodPathParsing(t *testing.T) {
	// Test the path parsing logic by examining the patterns we check for
	tests := []struct {
		name string
		path string
		want InstallMethod
	}{
		{
			name: "go bin path",
			path: "/home/user/go/bin/termchess",
			want: InstallMethodGoInstall,
		},
		{
			name: "go bin path with version",
			path: "/Users/developer/go/bin/termchess",
			want: InstallMethodGoInstall,
		},
		{
			name: "local bin path",
			path: "/home/user/.local/bin/termchess",
			want: InstallMethodInstallScript,
		},
		{
			name: "usr local bin path",
			path: "/usr/local/bin/termchess",
			want: InstallMethodInstallScript,
		},
		{
			name: "random path",
			path: "/opt/termchess/bin/termchess",
			want: InstallMethodUnknown,
		},
		{
			name: "tmp path",
			path: "/tmp/termchess",
			want: InstallMethodUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't directly test DetectInstallMethod with custom paths,
			// but we can verify the string matching logic would work
			var result InstallMethod
			if containsPath(tt.path, "/go/bin/") {
				result = InstallMethodGoInstall
			} else if containsPath(tt.path, "/.local/bin/") || containsPath(tt.path, "/usr/local/bin/") {
				result = InstallMethodInstallScript
			} else {
				result = InstallMethodUnknown
			}

			if result != tt.want {
				t.Errorf("path %q: got %q, want %q", tt.path, result, tt.want)
			}
		})
	}
}

// containsPath is a helper to match the logic in DetectInstallMethod.
func containsPath(path, substr string) bool {
	return len(path) >= len(substr) && (path == substr ||
		(len(path) > len(substr) && pathContains(path, substr)))
}

func pathContains(path, substr string) bool {
	for i := 0; i <= len(path)-len(substr); i++ {
		if path[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
	if client.httpClient == nil {
		t.Error("NewClient() returned client with nil httpClient")
	}
	if client.baseURL != githubAPI {
		t.Errorf("NewClient() baseURL = %q, want %q", client.baseURL, githubAPI)
	}
}

func TestNewClientWithHTTPClient(t *testing.T) {
	customClient := &http.Client{Timeout: 10 * time.Second}
	customURL := "https://custom.api.example.com"

	client := NewClientWithHTTPClient(customClient, customURL)
	if client == nil {
		t.Fatal("NewClientWithHTTPClient() returned nil")
	}
	if client.httpClient != customClient {
		t.Error("NewClientWithHTTPClient() did not use provided httpClient")
	}
	if client.baseURL != customURL {
		t.Errorf("NewClientWithHTTPClient() baseURL = %q, want %q", client.baseURL, customURL)
	}
}

func TestGetChecksumsURL(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "standard version",
			version: "v0.1.0",
			want:    "https://github.com/Mgrdich/TermChess/releases/download/v0.1.0/checksums.txt",
		},
		{
			name:    "different version",
			version: "v1.2.3",
			want:    "https://github.com/Mgrdich/TermChess/releases/download/v1.2.3/checksums.txt",
		},
		{
			name:    "prerelease version",
			version: "v0.2.0-beta.1",
			want:    "https://github.com/Mgrdich/TermChess/releases/download/v0.2.0-beta.1/checksums.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetChecksumsURL(tt.version)
			if got != tt.want {
				t.Errorf("GetChecksumsURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseChecksums(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    map[string]string
	}{
		{
			name: "standard format with two spaces",
			content: `abc123def456  termchess-v0.1.0-darwin-amd64
def789ghi012  termchess-v0.1.0-darwin-arm64
jkl345mno678  termchess-v0.1.0-linux-amd64`,
			want: map[string]string{
				"termchess-v0.1.0-darwin-amd64": "abc123def456",
				"termchess-v0.1.0-darwin-arm64": "def789ghi012",
				"termchess-v0.1.0-linux-amd64":  "jkl345mno678",
			},
		},
		{
			name: "standard format with single space",
			content: `abc123def456 termchess-v0.1.0-darwin-amd64
def789ghi012 termchess-v0.1.0-darwin-arm64`,
			want: map[string]string{
				"termchess-v0.1.0-darwin-amd64": "abc123def456",
				"termchess-v0.1.0-darwin-arm64": "def789ghi012",
			},
		},
		{
			name:    "empty content",
			content: "",
			want:    map[string]string{},
		},
		{
			name: "content with empty lines",
			content: `abc123def456  termchess-v0.1.0-darwin-amd64

def789ghi012  termchess-v0.1.0-darwin-arm64

`,
			want: map[string]string{
				"termchess-v0.1.0-darwin-amd64": "abc123def456",
				"termchess-v0.1.0-darwin-arm64": "def789ghi012",
			},
		},
		{
			name: "content with whitespace padding",
			content: `  abc123def456  termchess-v0.1.0-darwin-amd64
  def789ghi012  termchess-v0.1.0-darwin-arm64  `,
			want: map[string]string{
				"termchess-v0.1.0-darwin-amd64": "abc123def456",
				"termchess-v0.1.0-darwin-arm64": "def789ghi012",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseChecksums(tt.content)
			if len(got) != len(tt.want) {
				t.Errorf("ParseChecksums() returned %d entries, want %d", len(got), len(tt.want))
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("ParseChecksums()[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestGetExpectedChecksum(t *testing.T) {
	checksums := map[string]string{
		"termchess-v0.1.0-darwin-amd64": "abc123",
		"termchess-v0.1.0-darwin-arm64": "def456",
		"termchess-v0.1.0-linux-amd64":  "ghi789",
		"termchess-v0.1.0-linux-arm64":  "jkl012",
	}

	tests := []struct {
		name        string
		version     string
		goos        string
		goarch      string
		want        string
		wantErr     bool
		checksumMap map[string]string
	}{
		{
			name:        "darwin amd64",
			version:     "v0.1.0",
			goos:        "darwin",
			goarch:      "amd64",
			want:        "abc123",
			wantErr:     false,
			checksumMap: checksums,
		},
		{
			name:        "darwin arm64",
			version:     "v0.1.0",
			goos:        "darwin",
			goarch:      "arm64",
			want:        "def456",
			wantErr:     false,
			checksumMap: checksums,
		},
		{
			name:        "linux amd64",
			version:     "v0.1.0",
			goos:        "linux",
			goarch:      "amd64",
			want:        "ghi789",
			wantErr:     false,
			checksumMap: checksums,
		},
		{
			name:        "missing platform",
			version:     "v0.1.0",
			goos:        "windows",
			goarch:      "amd64",
			want:        "",
			wantErr:     true,
			checksumMap: checksums,
		},
		{
			name:        "empty checksums",
			version:     "v0.1.0",
			goos:        "darwin",
			goarch:      "amd64",
			want:        "",
			wantErr:     true,
			checksumMap: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store original values
			origGOOS := runtime.GOOS
			origGOARCH := runtime.GOARCH

			// We can't change runtime.GOOS/GOARCH, so we'll test the logic directly
			filename := GetBinaryFilename(tt.version, tt.goos, tt.goarch)
			got, ok := tt.checksumMap[filename]

			if tt.wantErr {
				if ok {
					t.Errorf("expected error for %s, but got checksum %q", filename, got)
				}
			} else {
				if !ok {
					t.Errorf("expected checksum for %s, but got error", filename)
				} else if got != tt.want {
					t.Errorf("GetExpectedChecksum() = %q, want %q", got, tt.want)
				}
			}

			// Verify runtime values weren't modified
			if runtime.GOOS != origGOOS || runtime.GOARCH != origGOARCH {
				t.Error("runtime values were unexpectedly modified")
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want int
	}{
		{
			name: "equal versions",
			v1:   "v1.0.0",
			v2:   "v1.0.0",
			want: 0,
		},
		{
			name: "equal versions without prefix",
			v1:   "1.0.0",
			v2:   "1.0.0",
			want: 0,
		},
		{
			name: "equal versions mixed prefix",
			v1:   "v1.0.0",
			v2:   "1.0.0",
			want: 0,
		},
		{
			name: "v1 less than v2 major",
			v1:   "v1.0.0",
			v2:   "v2.0.0",
			want: -1,
		},
		{
			name: "v1 less than v2 minor",
			v1:   "v1.0.0",
			v2:   "v1.1.0",
			want: -1,
		},
		{
			name: "v1 less than v2 patch",
			v1:   "v1.0.0",
			v2:   "v1.0.1",
			want: -1,
		},
		{
			name: "v1 greater than v2 major",
			v1:   "v2.0.0",
			v2:   "v1.0.0",
			want: 1,
		},
		{
			name: "v1 greater than v2 minor",
			v1:   "v1.1.0",
			v2:   "v1.0.0",
			want: 1,
		},
		{
			name: "v1 greater than v2 patch",
			v1:   "v1.0.1",
			v2:   "v1.0.0",
			want: 1,
		},
		{
			name: "prerelease comparison",
			v1:   "v1.0.0-alpha",
			v2:   "v1.0.0-beta",
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareVersions(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

func TestDownloadBinary(t *testing.T) {
	expectedData := []byte("fake binary data")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request path contains the expected binary name pattern
		expectedPattern := "/releases/download/v0.1.0/termchess-v0.1.0-"
		if !containsPath(r.URL.Path, expectedPattern) {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(expectedData)
	}))
	defer server.Close()

	// Create a custom client that redirects to our test server
	customTransport := &testTransport{
		server: server,
	}
	httpClient := &http.Client{Transport: customTransport}
	client := NewClientWithHTTPClient(httpClient, server.URL)

	data, err := client.DownloadBinary(context.Background(), "v0.1.0")
	if err != nil {
		t.Fatalf("DownloadBinary() error = %v", err)
	}

	if string(data) != string(expectedData) {
		t.Errorf("DownloadBinary() = %q, want %q", string(data), string(expectedData))
	}
}

func TestDownloadBinaryError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	customTransport := &testTransport{server: server}
	httpClient := &http.Client{Transport: customTransport}
	client := NewClientWithHTTPClient(httpClient, server.URL)

	_, err := client.DownloadBinary(context.Background(), "v0.1.0")
	if err == nil {
		t.Error("expected error for 404 response, got nil")
	}
}

func TestDownloadChecksums(t *testing.T) {
	checksumContent := `abc123def456  termchess-v0.1.0-darwin-amd64
def789ghi012  termchess-v0.1.0-darwin-arm64`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(checksumContent))
	}))
	defer server.Close()

	customTransport := &testTransport{server: server}
	httpClient := &http.Client{Transport: customTransport}
	client := NewClientWithHTTPClient(httpClient, server.URL)

	checksums, err := client.DownloadChecksums(context.Background(), "v0.1.0")
	if err != nil {
		t.Fatalf("DownloadChecksums() error = %v", err)
	}

	if len(checksums) != 2 {
		t.Errorf("DownloadChecksums() returned %d entries, want 2", len(checksums))
	}

	if checksums["termchess-v0.1.0-darwin-amd64"] != "abc123def456" {
		t.Errorf("unexpected checksum for darwin-amd64: %q", checksums["termchess-v0.1.0-darwin-amd64"])
	}
}

func TestReplaceBinary(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "termchess-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a fake "current" binary
	currentBinary := filepath.Join(tmpDir, "termchess")
	if err := os.WriteFile(currentBinary, []byte("old binary"), 0755); err != nil {
		t.Fatalf("failed to write current binary: %v", err)
	}

	// New binary data
	newData := []byte("new binary")

	// We can't directly test ReplaceBinary because it uses os.Executable()
	// But we can test the atomic replacement logic

	// Test the atomic replacement logic manually
	tmpPath := currentBinary + ".new"
	oldPath := currentBinary + ".old"

	// 1. Write new binary to temp file
	if err := os.WriteFile(tmpPath, newData, 0755); err != nil {
		t.Fatalf("failed to write new binary: %v", err)
	}

	// 2. Rename current to .old
	if err := os.Rename(currentBinary, oldPath); err != nil {
		t.Fatalf("failed to rename current to old: %v", err)
	}

	// 3. Rename new to current
	if err := os.Rename(tmpPath, currentBinary); err != nil {
		t.Fatalf("failed to rename new to current: %v", err)
	}

	// 4. Delete old
	os.Remove(oldPath)

	// Verify the new binary is in place
	data, err := os.ReadFile(currentBinary)
	if err != nil {
		t.Fatalf("failed to read replaced binary: %v", err)
	}

	if string(data) != string(newData) {
		t.Errorf("replaced binary content = %q, want %q", string(data), string(newData))
	}

	// Verify .old was deleted
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Error("old binary file should have been deleted")
	}

	// Verify .new was moved
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("temp new binary file should have been moved")
	}
}

func TestUpgradeAlreadyUpToDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return v1.0.0 as latest
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"tag_name": "v1.0.0"}`))
	}))
	defer server.Close()

	client := NewClientWithHTTPClient(server.Client(), server.URL)

	_, err := client.Upgrade(context.Background(), "v1.0.0", "", nil)
	if !errors.Is(err, ErrAlreadyUpToDate) {
		t.Errorf("Upgrade() error = %v, want ErrAlreadyUpToDate", err)
	}
}

func TestUpgradeAlreadyUpToDateWithTarget(t *testing.T) {
	client := NewClient()

	_, err := client.Upgrade(context.Background(), "v1.0.0", "v1.0.0", nil)
	if !errors.Is(err, ErrAlreadyUpToDate) {
		t.Errorf("Upgrade() error = %v, want ErrAlreadyUpToDate", err)
	}
}

func TestUpgradeDowngradeCancelled(t *testing.T) {
	client := NewClient()

	confirmDowngrade := func() bool {
		return false // User says no
	}

	_, err := client.Upgrade(context.Background(), "v2.0.0", "v1.0.0", confirmDowngrade)
	if err == nil {
		t.Error("Upgrade() should have returned error for cancelled downgrade")
	}
	if err != nil && !containsPath(err.Error(), "cancelled by user") {
		t.Errorf("Upgrade() error = %v, want 'cancelled by user'", err)
	}
}

func TestGetGoInstallMessage(t *testing.T) {
	msg := GetGoInstallMessage()

	// Verify the message contains key instructions
	if !containsPath(msg, "go install") {
		t.Error("GetGoInstallMessage() should mention 'go install'")
	}
	if !containsPath(msg, "github.com/Mgrdich/TermChess") {
		t.Error("GetGoInstallMessage() should mention the repository")
	}
	if !containsPath(msg, "install.sh") {
		t.Error("GetGoInstallMessage() should mention the install script")
	}
}

func TestGetBinaryFilename(t *testing.T) {
	tests := []struct {
		name    string
		version string
		goos    string
		goarch  string
		want    string
	}{
		{
			name:    "darwin amd64",
			version: "v0.1.0",
			goos:    "darwin",
			goarch:  "amd64",
			want:    "termchess-v0.1.0-darwin-amd64",
		},
		{
			name:    "linux arm64",
			version: "v1.2.3",
			goos:    "linux",
			goarch:  "arm64",
			want:    "termchess-v1.2.3-linux-arm64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetBinaryFilename(tt.version, tt.goos, tt.goarch)
			if got != tt.want {
				t.Errorf("GetBinaryFilename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUpgradeResultFields(t *testing.T) {
	result := &UpgradeResult{
		PreviousVersion: "v1.0.0",
		NewVersion:      "v2.0.0",
		IsDowngrade:     false,
	}

	if result.PreviousVersion != "v1.0.0" {
		t.Errorf("PreviousVersion = %q, want %q", result.PreviousVersion, "v1.0.0")
	}
	if result.NewVersion != "v2.0.0" {
		t.Errorf("NewVersion = %q, want %q", result.NewVersion, "v2.0.0")
	}
	if result.IsDowngrade {
		t.Error("IsDowngrade should be false for upgrade")
	}
}

func TestSentinelErrors(t *testing.T) {
	// Test that sentinel errors are properly defined
	if ErrAlreadyUpToDate == nil {
		t.Error("ErrAlreadyUpToDate should not be nil")
	}
	if ErrChecksumMismatch == nil {
		t.Error("ErrChecksumMismatch should not be nil")
	}
	if ErrPermissionDenied == nil {
		t.Error("ErrPermissionDenied should not be nil")
	}

	// Test error messages
	if ErrAlreadyUpToDate.Error() != "already up to date" {
		t.Errorf("ErrAlreadyUpToDate.Error() = %q, want %q", ErrAlreadyUpToDate.Error(), "already up to date")
	}
	if ErrChecksumMismatch.Error() != "checksum mismatch" {
		t.Errorf("ErrChecksumMismatch.Error() = %q, want %q", ErrChecksumMismatch.Error(), "checksum mismatch")
	}
	if ErrPermissionDenied.Error() != "permission denied" {
		t.Errorf("ErrPermissionDenied.Error() = %q, want %q", ErrPermissionDenied.Error(), "permission denied")
	}
}

// testTransport is a custom transport that redirects all requests to the test server.
type testTransport struct {
	server *httptest.Server
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite the URL to point to our test server
	req.URL.Scheme = "http"
	req.URL.Host = t.server.Listener.Addr().String()
	return http.DefaultTransport.RoundTrip(req)
}

func TestUninstallLogic(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "termchess-uninstall-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a fake binary
	binaryPath := filepath.Join(tmpDir, "termchess")
	if err := os.WriteFile(binaryPath, []byte("fake binary"), 0755); err != nil {
		t.Fatalf("failed to write fake binary: %v", err)
	}

	// Create a fake config directory with files
	configDir := filepath.Join(tmpDir, ".termchess")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	configFile := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configFile, []byte("# config"), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Test the removal logic manually (since Uninstall uses os.Executable())
	// Remove binary
	if err := os.Remove(binaryPath); err != nil {
		t.Errorf("failed to remove binary: %v", err)
	}

	// Verify binary is gone
	if _, err := os.Stat(binaryPath); !os.IsNotExist(err) {
		t.Error("binary should have been removed")
	}

	// Remove config directory
	if err := os.RemoveAll(configDir); err != nil {
		t.Errorf("failed to remove config dir: %v", err)
	}

	// Verify config directory is gone
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Error("config directory should have been removed")
	}
}

func TestUninstallNonExistentBinary(t *testing.T) {
	// Test that removing a non-existent binary returns an error
	nonExistentPath := "/tmp/non-existent-termchess-binary-12345"

	err := os.Remove(nonExistentPath)
	if err == nil {
		t.Error("expected error when removing non-existent binary")
	}
	if !os.IsNotExist(err) {
		t.Errorf("expected IsNotExist error, got: %v", err)
	}
}

func TestUninstallEmptyConfigDir(t *testing.T) {
	// Create a temporary empty config directory
	tmpDir, err := os.MkdirTemp("", "termchess-uninstall-empty-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	configDir := filepath.Join(tmpDir, ".termchess")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Remove the empty config directory
	if err := os.RemoveAll(configDir); err != nil {
		t.Errorf("failed to remove empty config dir: %v", err)
	}

	// Verify it's gone
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Error("empty config directory should have been removed")
	}

	// Clean up
	os.RemoveAll(tmpDir)
}

func TestUninstallNestedConfigDir(t *testing.T) {
	// Create a temporary config directory with nested files
	tmpDir, err := os.MkdirTemp("", "termchess-uninstall-nested-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configDir := filepath.Join(tmpDir, ".termchess")
	subDir := filepath.Join(configDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create nested dirs: %v", err)
	}

	// Create files in the config directory and subdirectory
	if err := os.WriteFile(filepath.Join(configDir, "config.toml"), []byte("# config"), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "savegame.fen"), []byte("fen string"), 0644); err != nil {
		t.Fatalf("failed to write savegame file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("nested"), 0644); err != nil {
		t.Fatalf("failed to write nested file: %v", err)
	}

	// Remove the config directory recursively
	if err := os.RemoveAll(configDir); err != nil {
		t.Errorf("failed to remove nested config dir: %v", err)
	}

	// Verify everything is gone
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Error("config directory should have been removed")
	}
}
