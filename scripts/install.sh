#!/bin/bash
set -euo pipefail

# TermChess Install Script
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/Mgrdich/TermChess/main/scripts/install.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/Mgrdich/TermChess/main/scripts/install.sh | bash -s -- v1.0.0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Repository configuration
REPO_OWNER="Mgrdich"
REPO_NAME="TermChess"
GITHUB_API="https://api.github.com"
GITHUB_RELEASES="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases"

# Global variables
INSTALL_DIR=""
TEMP_DIR=""

# Cleanup function
cleanup() {
    if [ -n "${TEMP_DIR:-}" ] && [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR"
    fi
}
trap cleanup EXIT

# Output functions
error() {
    echo -e "${RED}Error: $1${NC}" >&2
    exit 1
}

warn() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

success() {
    echo -e "${GREEN}✓ $1${NC}"
}

info() {
    echo -e "$1"
}

# Detect operating system
detect_os() {
    case "$(uname -s)" in
        Darwin) echo "darwin" ;;
        Linux) echo "linux" ;;
        *) echo "unsupported" ;;
    esac
}

# Detect CPU architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *) echo "unsupported" ;;
    esac
}

# Detect user's shell configuration file
detect_shell_rc() {
    local shell_name
    shell_name=$(basename "${SHELL:-/bin/bash}")
    
    case "$shell_name" in
        zsh) echo "$HOME/.zshrc" ;;
        bash)
            # On macOS, bash uses .bash_profile for login shells
            if [ "$(detect_os)" = "darwin" ] && [ -f "$HOME/.bash_profile" ]; then
                echo "$HOME/.bash_profile"
            else
                echo "$HOME/.bashrc"
            fi
            ;;
        *) echo "$HOME/.profile" ;;
    esac
}

# Get the latest version from GitHub releases
get_latest_version() {
    local version
    version=$(curl -fsSL "${GITHUB_API}/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest" 2>/dev/null | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$version" ]; then
        error "Failed to fetch latest version from GitHub. Please check your internet connection or specify a version manually."
    fi
    
    echo "$version"
}

# Find existing termchess installation
find_existing_installation() {
    # Check PATH first
    if command -v termchess >/dev/null 2>&1; then
        command -v termchess
        return 0
    fi
    
    # Check common locations
    local locations=(
        "$HOME/.local/bin/termchess"
        "/usr/local/bin/termchess"
        "/usr/bin/termchess"
    )
    
    for loc in "${locations[@]}"; do
        if [ -x "$loc" ]; then
            echo "$loc"
            return 0
        fi
    done
    
    return 1
}

# Get installed version of termchess
get_installed_version() {
    local binary_path
    if binary_path=$(find_existing_installation); then
        # Try to get version from the binary
        local version
        version=$("$binary_path" --version 2>/dev/null | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | head -1) || true
        echo "$version"
    fi
}

# Prompt user for upgrade confirmation
prompt_upgrade() {
    local installed_version="$1"
    local target_version="$2"
    
    echo ""
    info "TermChess is already installed (${installed_version})"
    info "Latest version available: ${target_version}"
    echo ""
    
    # Check if running interactively
    if [ -t 0 ]; then
        read -r -p "Would you like to upgrade? [y/N] " response
        case "$response" in
            [yY]|[yY][eE][sS])
                return 0
                ;;
            *)
                info "Installation cancelled."
                exit 0
                ;;
        esac
    else
        # Non-interactive mode - skip upgrade by default
        info "Running in non-interactive mode. Use 'curl ... | bash -s -- --force' to force upgrade."
        exit 0
    fi
}

# Download binary from GitHub releases
download_binary() {
    local version="$1"
    local os="$2"
    local arch="$3"
    local output_file="$4"
    
    local binary_name="termchess-${version}-${os}-${arch}"
    local download_url="${GITHUB_RELEASES}/download/${version}/${binary_name}"
    
    info "Downloading TermChess ${version} for ${os}/${arch}..."
    
    if ! curl -fsSL "$download_url" -o "$output_file" 2>/dev/null; then
        error "Failed to download binary from ${download_url}. Please check if the version exists."
    fi
}

# Download and verify checksum
verify_checksum() {
    local version="$1"
    local os="$2"
    local arch="$3"
    local binary_file="$4"
    
    local binary_name="termchess-${version}-${os}-${arch}"
    local checksums_url="${GITHUB_RELEASES}/download/${version}/checksums.txt"
    local checksums_file="${TEMP_DIR}/checksums.txt"
    
    info "Verifying checksum..."
    
    # Download checksums file
    if ! curl -fsSL "$checksums_url" -o "$checksums_file" 2>/dev/null; then
        error "Failed to download checksums file. Cannot verify binary integrity."
    fi
    
    # Extract expected checksum for this binary
    local expected_checksum
    expected_checksum=$(grep "${binary_name}$" "$checksums_file" | awk '{print $1}')
    
    if [ -z "$expected_checksum" ]; then
        error "Could not find checksum for ${binary_name} in checksums.txt"
    fi
    
    # Calculate actual checksum
    local actual_checksum
    if command -v sha256sum >/dev/null 2>&1; then
        actual_checksum=$(sha256sum "$binary_file" | awk '{print $1}')
    elif command -v shasum >/dev/null 2>&1; then
        actual_checksum=$(shasum -a 256 "$binary_file" | awk '{print $1}')
    else
        error "No SHA256 utility found. Please install sha256sum or shasum."
    fi
    
    # Compare checksums
    if [ "$expected_checksum" != "$actual_checksum" ]; then
        error "Checksum verification failed. Download may be corrupted."
    fi
    
    success "Checksum verified"
}

# Install binary to destination
install_binary() {
    local binary_file="$1"
    local install_dir="$2"
    
    # Try to create install directory
    if ! mkdir -p "$install_dir" 2>/dev/null; then
        return 1
    fi
    
    # Try to install
    if ! mv "$binary_file" "${install_dir}/termchess" 2>/dev/null; then
        return 1
    fi
    
    if ! chmod +x "${install_dir}/termchess" 2>/dev/null; then
        return 1
    fi
    
    return 0
}

# Install binary with sudo to system location
install_binary_sudo() {
    local binary_file="$1"
    local install_dir="$2"
    
    info "Installing to ${install_dir} requires administrator privileges..."
    
    if ! sudo mkdir -p "$install_dir"; then
        error "Failed to create directory ${install_dir}"
    fi
    
    if ! sudo mv "$binary_file" "${install_dir}/termchess"; then
        error "Failed to install binary to ${install_dir}"
    fi
    
    if ! sudo chmod +x "${install_dir}/termchess"; then
        error "Failed to set executable permissions"
    fi
    
    return 0
}

# Check if install directory is in PATH
check_path() {
    local install_dir="$1"
    
    # Check if directory is in PATH
    if echo "$PATH" | tr ':' '\n' | grep -q "^${install_dir}$"; then
        return 0
    fi
    
    # Directory not in PATH - show instructions
    local shell_rc
    shell_rc=$(detect_shell_rc)
    local shell_rc_name
    shell_rc_name=$(basename "$shell_rc")
    
    warn "${install_dir} is not in your PATH. Add it by running:"
    echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/${shell_rc_name}"
    echo "  source ~/${shell_rc_name}"
    echo ""
}

# Main installation function
main() {
    local version="${1:-}"
    local force=false
    
    # Parse arguments
    while [ $# -gt 0 ]; do
        case "$1" in
            --force|-f)
                force=true
                shift
                ;;
            v*)
                version="$1"
                shift
                ;;
            *)
                shift
                ;;
        esac
    done
    
    echo ""
    info "TermChess Installer"
    info "==================="
    echo ""
    
    # Detect platform
    local os
    local arch
    os=$(detect_os)
    arch=$(detect_arch)
    
    # Validate platform
    if [ "$os" = "unsupported" ]; then
        error "Unsupported operating system. TermChess supports macOS and Linux."
    fi
    
    if [ "$arch" = "unsupported" ]; then
        error "Unsupported architecture. TermChess supports amd64 and arm64 only."
    fi
    
    info "Detected platform: ${os}/${arch}"
    
    # Get version to install
    if [ -z "$version" ]; then
        info "Fetching latest version..."
        version=$(get_latest_version)
    fi
    
    info "Target version: ${version}"
    
    # Check for existing installation
    local installed_version
    installed_version=$(get_installed_version)
    
    if [ -n "$installed_version" ] && [ "$force" = false ]; then
        if [ "$installed_version" = "$version" ]; then
            success "TermChess ${version} is already installed"
            exit 0
        fi
        prompt_upgrade "$installed_version" "$version"
    fi
    
    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    local temp_binary="${TEMP_DIR}/termchess"
    
    # Download binary
    download_binary "$version" "$os" "$arch" "$temp_binary"
    
    # Verify checksum
    verify_checksum "$version" "$os" "$arch" "$temp_binary"
    
    # Determine install location
    local user_install_dir="$HOME/.local/bin"
    local system_install_dir="/usr/local/bin"
    
    # Try user directory first
    if install_binary "$temp_binary" "$user_install_dir"; then
        INSTALL_DIR="$user_install_dir"
    else
        # Fall back to system directory with sudo
        info "Cannot install to ${user_install_dir}, falling back to ${system_install_dir}"
        if install_binary_sudo "$temp_binary" "$system_install_dir"; then
            INSTALL_DIR="$system_install_dir"
        else
            error "Failed to install TermChess. Please check permissions."
        fi
    fi
    
    # Check PATH
    if [ "$INSTALL_DIR" = "$user_install_dir" ]; then
        check_path "$INSTALL_DIR"
    fi
    
    echo ""
    success "TermChess ${version} installed to ${INSTALL_DIR}/termchess"
    echo "Run 'termchess' to start playing!"
    echo ""
}

main "$@"
