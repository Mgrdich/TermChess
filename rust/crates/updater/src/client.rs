//! HTTP client for checking versions, downloading assets, and performing the
//! upgrade flow. Ported from the Go `Client` type and its methods.

use std::collections::HashMap;
use std::io::Read;
use std::time::Duration;

use serde::Deserialize;

use crate::checksum::{get_expected_checksum, parse_checksums, verify_checksum};
use crate::context::Context;
use crate::error::UpdaterError;
use crate::install::replace_binary;
use crate::platform::{current_goarch, current_goos};
use crate::urls::{asset_url_with_base, checksums_url_with_base, GITHUB_DOWNLOAD_BASE};
use crate::version_cmp::{compare_versions, normalize_version};
use crate::{GITHUB_API, REPO_NAME, REPO_OWNER};

/// Relevant fields from GitHub's release API response.
#[derive(Deserialize)]
struct GithubRelease {
    #[serde(default)]
    tag_name: String,
}

/// Information about a completed upgrade.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct UpgradeResult {
    /// The version that was running before the upgrade.
    pub previous_version: String,
    /// The version that was installed.
    pub new_version: String,
    /// Whether the operation was a downgrade.
    pub is_downgrade: bool,
}

/// Provides methods for checking and downloading updates.
pub struct Client {
    agent: ureq::Agent,
    /// Base URL for the GitHub API (release metadata).
    base_url: String,
    /// Base host for release asset downloads.
    download_base_url: String,
}

impl Client {
    /// Creates a new updater client with default settings (30s request timeout,
    /// the public GitHub API).
    pub fn new() -> Self {
        let agent = ureq::AgentBuilder::new()
            .timeout_read(Duration::from_secs(30))
            .timeout_write(Duration::from_secs(30))
            .build();
        Self {
            agent,
            base_url: GITHUB_API.to_string(),
            download_base_url: GITHUB_DOWNLOAD_BASE.to_string(),
        }
    }

    /// Creates a new updater client with a custom agent and API base URL.
    ///
    /// Useful for testing with mock servers. Mirrors Go's
    /// `NewClientWithHTTPClient`.
    pub fn with_http_client(agent: ureq::Agent, base_url: impl Into<String>) -> Self {
        Self {
            agent,
            base_url: base_url.into(),
            download_base_url: GITHUB_DOWNLOAD_BASE.to_string(),
        }
    }

    /// Overrides the base host used for release asset downloads (builder-style).
    ///
    /// Downloads normally target `github.com`; this lets tests redirect them to
    /// a local server, replacing the transport rewriting used in the Go tests.
    pub fn with_download_base_url(mut self, url: impl Into<String>) -> Self {
        self.download_base_url = url.into();
        self
    }

    /// Returns the configured API base URL.
    pub fn base_url(&self) -> &str {
        &self.base_url
    }

    /// Queries the GitHub API for the latest release version.
    ///
    /// Returns the version tag (e.g. `"v0.1.0"`) or an error.
    pub fn check_latest_version(&self, ctx: &Context) -> Result<String, UpdaterError> {
        let url = format!(
            "{}/repos/{}/{}/releases/latest",
            self.base_url, REPO_OWNER, REPO_NAME
        );

        let body = self.execute_get(ctx, &url, true)?;

        let release: GithubRelease =
            serde_json::from_slice(&body).map_err(|e| UpdaterError::Parse(e.to_string()))?;

        if release.tag_name.is_empty() {
            return Err(UpdaterError::EmptyTagName);
        }

        Ok(release.tag_name)
    }

    /// Downloads the binary for the specified version and current platform.
    pub fn download_binary(&self, ctx: &Context, version: &str) -> Result<Vec<u8>, UpdaterError> {
        let url = asset_url_with_base(
            &self.download_base_url,
            version,
            current_goos(),
            current_goarch(),
        );
        self.download_file(ctx, &url)
    }

    /// Downloads and parses the `checksums.txt` file for a release.
    pub fn download_checksums(
        &self,
        ctx: &Context,
        version: &str,
    ) -> Result<HashMap<String, String>, UpdaterError> {
        let url = checksums_url_with_base(&self.download_base_url, version);
        let data = self.download_file(ctx, &url).map_err(|e| match e {
            // Preserve the "downloading checksums" context from the Go code
            // while keeping sentinel errors intact.
            UpdaterError::Network(msg) => {
                UpdaterError::Message(format!("downloading checksums: {msg}"))
            }
            other => other,
        })?;

        Ok(parse_checksums(&String::from_utf8_lossy(&data)))
    }

    /// Performs the upgrade to the specified version.
    ///
    /// If `target_version` is empty, upgrades to the latest version.
    /// `confirm_downgrade`, when provided, is invoked if the target is older
    /// than the current version; returning `false` aborts the upgrade.
    pub fn upgrade(
        &self,
        ctx: &Context,
        current_version: &str,
        target_version: &str,
        confirm_downgrade: Option<&dyn Fn() -> bool>,
    ) -> Result<UpgradeResult, UpdaterError> {
        // If no target version specified, get the latest.
        let mut target_version = target_version.to_string();
        if target_version.is_empty() {
            target_version = self
                .check_latest_version(ctx)
                .map_err(|e| UpdaterError::Message(format!("checking latest version: {e}")))?;
        }

        // Normalize versions for comparison.
        let normalized_current = normalize_version(current_version);
        let normalized_target = normalize_version(&target_version);

        // Check if already up to date.
        if normalized_current == normalized_target {
            return Err(UpdaterError::AlreadyUpToDate);
        }

        // Ensure target version has 'v' prefix for URLs.
        if !target_version.starts_with('v') {
            target_version = format!("v{target_version}");
        }

        // Check if this is a downgrade.
        let is_downgrade = compare_versions(&normalized_target, &normalized_current) < 0;
        if is_downgrade {
            if let Some(confirm) = confirm_downgrade {
                if !confirm() {
                    return Err(UpdaterError::DowngradeCancelled);
                }
            }
        }

        // Download checksums first.
        let checksums = self.download_checksums(ctx, &target_version)?;

        // Get expected checksum.
        let expected_checksum = get_expected_checksum(
            &checksums,
            &target_version,
            current_goos(),
            current_goarch(),
        )?;

        // Download the binary.
        let binary_data = self
            .download_binary(ctx, &target_version)
            .map_err(|e| UpdaterError::Message(format!("downloading binary: {e}")))?;

        // Verify checksum.
        if !verify_checksum(&binary_data, &expected_checksum) {
            return Err(UpdaterError::ChecksumMismatch);
        }

        // Replace the binary.
        replace_binary(&binary_data)?;

        Ok(UpgradeResult {
            previous_version: current_version.to_string(),
            new_version: target_version,
            is_downgrade,
        })
    }

    /// Performs an HTTP GET and returns the response body.
    fn download_file(&self, ctx: &Context, url: &str) -> Result<Vec<u8>, UpdaterError> {
        self.execute_get(ctx, url, false)
    }

    /// Core request executor shared by all GET operations.
    ///
    /// `github_json` adds the GitHub API `Accept` header used by
    /// `check_latest_version`.
    fn execute_get(
        &self,
        ctx: &Context,
        url: &str,
        github_json: bool,
    ) -> Result<Vec<u8>, UpdaterError> {
        // Honor cancellation before issuing the request.
        if ctx.is_cancelled() {
            return Err(UpdaterError::Cancelled);
        }

        let mut req = self.agent.get(url).set("User-Agent", "TermChess-Updater");
        if github_json {
            req = req.set("Accept", "application/vnd.github.v3+json");
        }

        // Translate a context deadline into a per-request timeout.
        if let Some(remaining) = ctx.remaining() {
            if remaining.is_zero() {
                return Err(UpdaterError::Timeout);
            }
            req = req.timeout(remaining);
        }

        match req.call() {
            Ok(resp) => {
                if resp.status() != 200 {
                    return Err(UpdaterError::UnexpectedStatusCode(resp.status()));
                }
                let mut buf = Vec::new();
                resp.into_reader()
                    .read_to_end(&mut buf)
                    .map_err(|e| UpdaterError::Message(format!("reading response: {e}")))?;
                Ok(buf)
            }
            Err(ureq::Error::Status(code, _)) => Err(UpdaterError::UnexpectedStatusCode(code)),
            Err(ureq::Error::Transport(t)) => {
                if ctx.is_cancelled() {
                    Err(UpdaterError::Cancelled)
                } else {
                    Err(UpdaterError::Network(t.to_string()))
                }
            }
        }
    }
}

impl Default for Client {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_new_client() {
        let client = Client::new();
        assert_eq!(client.base_url(), GITHUB_API);
    }

    #[test]
    fn test_with_http_client() {
        let custom_url = "https://custom.api.example.com";
        let client = Client::with_http_client(ureq::agent(), custom_url);
        assert_eq!(client.base_url(), custom_url);
    }

    #[test]
    fn test_upgrade_already_up_to_date_with_target() {
        let client = Client::new();
        let err = client
            .upgrade(&Context::background(), "v1.0.0", "v1.0.0", None)
            .unwrap_err();
        assert!(matches!(err, UpdaterError::AlreadyUpToDate));
    }

    #[test]
    fn test_upgrade_downgrade_cancelled() {
        let client = Client::new();
        let confirm = || false; // user says no
        let err = client
            .upgrade(&Context::background(), "v2.0.0", "v1.0.0", Some(&confirm))
            .unwrap_err();
        assert!(matches!(err, UpdaterError::DowngradeCancelled));
        assert!(err.to_string().contains("cancelled by user"));
    }

    #[test]
    fn test_upgrade_result_fields() {
        let result = UpgradeResult {
            previous_version: "v1.0.0".to_string(),
            new_version: "v2.0.0".to_string(),
            is_downgrade: false,
        };
        assert_eq!(result.previous_version, "v1.0.0");
        assert_eq!(result.new_version, "v2.0.0");
        assert!(!result.is_downgrade);
    }
}
