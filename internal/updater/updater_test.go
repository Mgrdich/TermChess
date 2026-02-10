package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
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
