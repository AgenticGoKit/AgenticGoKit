#!/bin/bash
# AgenticGoKit CLI Installation Script for Linux/macOS
# Usage: curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.sh | bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Default values
VERSION="latest"
INSTALL_DIR=""
FORCE=false

# Helper functions
log_info() {
    echo -e "${CYAN}$1${NC}"
}

log_success() {
    echo -e "${GREEN}$1${NC}"
}

log_warning() {
    echo -e "${YELLOW}$1${NC}"
}

log_error() {
    echo -e "${RED}$1${NC}"
}

show_help() {
    echo -e "${CYAN}AgenticGoKit CLI Installer for Linux/macOS${NC}"
    echo -e "${CYAN}===========================================${NC}"
    echo ""
    echo -e "${YELLOW}USAGE:${NC}"
    echo -e "  ${WHITE}curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.sh | bash${NC}"
    echo ""
    echo -e "${YELLOW}OPTIONS:${NC}"
    echo -e "  ${WHITE}-v, --version <version>    Install specific version (default: latest)${NC}"
    echo -e "  ${WHITE}-d, --dir <path>          Custom installation directory${NC}"
    echo -e "  ${WHITE}-f, --force               Overwrite existing installation${NC}"
    echo -e "  ${WHITE}-h, --help                Show this help message${NC}"
    echo ""
    echo -e "${YELLOW}EXAMPLES:${NC}"
    echo -e "  ${WHITE}# Install latest version${NC}"
    echo -e "  ${WHITE}curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.sh | bash${NC}"
    echo ""
    echo -e "  ${WHITE}# Install specific version${NC}"
    echo -e "  ${WHITE}curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.sh | bash -s -- --version v0.3.0${NC}"
}

get_os() {
    case "$(uname -s)" in
        Darwin) echo "darwin" ;;
        Linux) echo "linux" ;;
        *) 
            log_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
}

get_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *)
            log_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
}

get_latest_version() {
    log_info "Fetching latest release information..."
    
    local api_url="https://api.github.com/repos/kunalkushwaha/agenticgokit/releases/latest"
    local response=""
    local http_code=""
    local version=""
    local curl_exit_code=0
    local wget_exit_code=0
    
    if command -v curl >/dev/null 2>&1; then
        # Use curl with detailed error handling
        response=$(curl -fsSL --max-time 30 --write-out "HTTPSTATUS:%{http_code}" "$api_url" 2>/dev/null)
        curl_exit_code=$?
        
        # Check if curl command itself failed
        if [ $curl_exit_code -ne 0 ]; then
            handle_api_error "000" "curl" "$curl_exit_code"
            return 1
        fi
        
        # Extract HTTP status code
        http_code=$(echo "$response" | grep -o "HTTPSTATUS:[0-9]*" | cut -d: -f2)
        response=$(echo "$response" | sed -E 's/HTTPSTATUS:[0-9]*$//')
        
        # Validate HTTP status code extraction
        if [ -z "$http_code" ]; then
            log_error "Failed to extract HTTP status code from response"
            handle_api_error "unknown" "curl" "$curl_exit_code"
            return 1
        fi
        
        # Check for successful HTTP response
        if [ "$http_code" = "200" ]; then
            # Validate response is not empty
            if [ -z "$response" ]; then
                log_error "Received empty response from GitHub API"
                handle_api_error "$http_code" "curl" "$curl_exit_code"
                return 1
            fi
            
            # Validate response contains expected JSON structure
            if ! echo "$response" | grep -q '"tag_name"'; then
                log_error "Invalid JSON response - missing tag_name field"
                log_warning "Response preview: $(echo "$response" | head -c 100)..."
                handle_api_error "$http_code" "curl" "$curl_exit_code"
                return 1
            fi
            
            # Extract version with validation
            version=$(echo "$response" | grep '"tag_name":' | sed -E 's/.*"tag_name":\s*"([^"]+)".*/\1/' | head -1)
            
            # Validate extracted version
            if [ -n "$version" ] && [ "$version" != "null" ]; then
                # Basic version format validation (should start with 'v' or be semantic version)
                if echo "$version" | grep -qE '^v?[0-9]+\.[0-9]+\.[0-9]+'; then
                    echo "$version"
                    return 0
                else
                    log_error "Invalid version format extracted: '$version'"
                fi
            else
                log_error "Failed to extract version from JSON response"
            fi
        fi
        
        # Handle non-200 HTTP status codes
        handle_api_error "$http_code" "curl" "$curl_exit_code"
        
    elif command -v wget >/dev/null 2>&1; then
        # Use wget with error handling
        response=$(wget --timeout=30 --tries=1 -qO- "$api_url" 2>/dev/null)
        wget_exit_code=$?
        
        # Check if wget command itself failed
        if [ $wget_exit_code -ne 0 ]; then
            handle_api_error "unknown" "wget" "$wget_exit_code"
            return 1
        fi
        
        # Validate response is not empty
        if [ -z "$response" ]; then
            log_error "Received empty response from GitHub API"
            handle_api_error "unknown" "wget" "$wget_exit_code"
            return 1
        fi
        
        # Validate response contains expected JSON structure
        if ! echo "$response" | grep -q '"tag_name"'; then
            log_error "Invalid JSON response - missing tag_name field"
            log_warning "Response preview: $(echo "$response" | head -c 100)..."
            handle_api_error "unknown" "wget" "$wget_exit_code"
            return 1
        fi
        
        # Extract version with validation
        version=$(echo "$response" | grep '"tag_name":' | sed -E 's/.*"tag_name":\s*"([^"]+)".*/\1/' | head -1)
        
        # Validate extracted version
        if [ -n "$version" ] && [ "$version" != "null" ]; then
            # Basic version format validation
            if echo "$version" | grep -qE '^v?[0-9]+\.[0-9]+\.[0-9]+'; then
                echo "$version"
                return 0
            else
                log_error "Invalid version format extracted: '$version'"
            fi
        else
            log_error "Failed to extract version from JSON response"
        fi
        
        # If we get here, something went wrong with parsing
        handle_api_error "unknown" "wget" "$wget_exit_code"
        
    else
        log_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi
    
    return 1
}

handle_api_error() {
    local http_code="$1"
    local tool="$2"
    local exit_code="${3:-0}"
    
    log_error "Failed to fetch latest version from GitHub API"
    
    case "$http_code" in
        "403")
            log_warning "  → Rate limit exceeded or access forbidden (HTTP 403)"
            log_warning "  → GitHub API has rate limits for unauthenticated requests"
            log_warning "  → Try again in a few minutes or specify a version manually"
            log_info "  → Example: --version v0.3.0"
            ;;
        "404")
            log_warning "  → Repository not found or releases not available (HTTP 404)"
            log_warning "  → The repository may have been moved or made private"
            log_info "  → Check if the repository exists: https://github.com/kunalkushwaha/agenticgokit"
            ;;
        "5"*)
            log_warning "  → GitHub server error (HTTP $http_code)"
            log_warning "  → This is a temporary issue with GitHub's servers"
            log_warning "  → Try again in a few minutes"
            ;;
        "000"|""|"unknown")
            # Handle tool-specific network errors
            if [ "$tool" = "curl" ] && [ "$exit_code" -ne 0 ]; then
                case "$exit_code" in
                    6) log_warning "  → Could not resolve hostname (curl error 6)" ;;
                    7) log_warning "  → Failed to connect to server (curl error 7)" ;;
                    28) log_warning "  → Operation timeout (curl error 28)" ;;
                    35) log_warning "  → SSL handshake error (curl error 35)" ;;
                    60) log_warning "  → SSL certificate verification failed (curl error 60)" ;;
                    *) log_warning "  → Network error (curl exit code: $exit_code)" ;;
                esac
            elif [ "$tool" = "wget" ] && [ "$exit_code" -ne 0 ]; then
                case "$exit_code" in
                    1) log_warning "  → Generic wget error - check network connection" ;;
                    4) log_warning "  → Network failure (wget error 4)" ;;
                    5) log_warning "  → SSL verification failure (wget error 5)" ;;
                    8) log_warning "  → Server issued an error response (wget error 8)" ;;
                    *) log_warning "  → Network error (wget exit code: $exit_code)" ;;
                esac
            else
                log_warning "  → Network connection failed"
            fi
            log_warning "  → Check your internet connection and try again"
            log_info "  → Verify you can access github.com in your browser"
            ;;
        *)
            log_warning "  → HTTP error code: $http_code"
            log_warning "  → This may be a temporary issue with GitHub"
            log_warning "  → Check your internet connection"
            ;;
    esac
    
    echo ""
    log_info "Alternative installation methods:"
    log_info "  1. Specify version manually: --version v0.3.0"
    log_info "  2. Manual download: https://github.com/kunalkushwaha/agenticgokit/releases"
    log_info "  3. Use Go install: go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest"
    log_info "  4. Check GitHub status: https://www.githubstatus.com/"
}

get_install_dir() {
    if [ -n "$INSTALL_DIR" ]; then
        echo "$INSTALL_DIR"
        return
    fi
    
    # Try to find a good installation directory
    if [ -w "/usr/local/bin" ]; then
        echo "/usr/local/bin"
    elif [ -d "$HOME/.local/bin" ]; then
        echo "$HOME/.local/bin"
    elif [ -d "$HOME/bin" ]; then
        echo "$HOME/bin"
    else
        # Create ~/.local/bin if it doesn't exist
        mkdir -p "$HOME/.local/bin"
        echo "$HOME/.local/bin"
    fi
}

check_path() {
    local install_dir="$1"
    if ! echo "$PATH" | grep -q "$install_dir"; then
        log_warning "$install_dir is not in your PATH."
        
        # Suggest adding to shell profile
        local shell_profile=""
        case "$SHELL" in
            */bash) shell_profile="$HOME/.bashrc" ;;
            */zsh) shell_profile="$HOME/.zshrc" ;;
            */fish) shell_profile="$HOME/.config/fish/config.fish" ;;
            *) shell_profile="$HOME/.profile" ;;
        esac
        
        log_info "Add the following line to your $shell_profile:"
        echo -e "${WHITE}export PATH=\"$install_dir:\$PATH\"${NC}"
        return 1
    fi
    return 0
}

test_installation() {
    local binary_path="$1"
    if [ -x "$binary_path" ]; then
        local version_output
        if version_output=$("$binary_path" version --short 2>/dev/null); then
            log_success "✓ Installation verified: $version_output"
            return 0
        fi
    fi
    log_error "✗ Installation verification failed"
    return 1
}

download_file() {
    local url="$1"
    local output="$2"
    local http_code=""
    
    if command -v curl >/dev/null 2>&1; then
        # Use curl with detailed error handling
        http_code=$(curl -fsSL --max-time 300 --write-out "%{http_code}" "$url" -o "$output" 2>/dev/null)
        local curl_exit=$?
        
        if [ $curl_exit -eq 0 ] && [ "$http_code" = "200" ]; then
            return 0
        fi
        
        # Handle curl-specific errors
        handle_download_error "$http_code" "$curl_exit" "curl" "$url"
        return 1
        
    elif command -v wget >/dev/null 2>&1; then
        # Use wget with error handling
        if wget --timeout=300 --tries=1 -q "$url" -O "$output" 2>/dev/null; then
            return 0
        fi
        
        local wget_exit=$?
        handle_download_error "unknown" "$wget_exit" "wget" "$url"
        return 1
        
    else
        log_error "Neither curl nor wget is available. Please install one of them."
        return 1
    fi
}

handle_download_error() {
    local http_code="$1"
    local exit_code="$2"
    local tool="$3"
    local url="$4"
    
    log_error "Failed to download binary"
    
    case "$http_code" in
        "404")
            log_warning "  → Binary not found (HTTP 404)"
            log_warning "  → The specified version may not exist or binaries aren't available"
            log_info "  → Check available versions: https://github.com/kunalkushwaha/agenticgokit/releases"
            log_info "  → Try a different version: --version v0.3.0"
            ;;
        "403")
            log_warning "  → Access forbidden (HTTP 403)"
            log_warning "  → GitHub may be rate limiting downloads"
            log_warning "  → Try again in a few minutes"
            ;;
        "5"*)
            log_warning "  → GitHub server error (HTTP $http_code)"
            log_warning "  → Try again in a few minutes"
            ;;
        *)
            # Handle tool-specific exit codes
            if [ "$tool" = "curl" ]; then
                case "$exit_code" in
                    6) log_warning "  → Could not resolve hostname - check your internet connection" ;;
                    7) log_warning "  → Failed to connect to server - check your internet connection" ;;
                    28) log_warning "  → Download timeout - check your internet connection" ;;
                    35) log_warning "  → SSL/TLS handshake error - update your certificates" ;;
                    60) log_warning "  → SSL certificate verification failed - update your certificates" ;;
                    *) log_warning "  → Network error (curl exit code: $exit_code)" ;;
                esac
            elif [ "$tool" = "wget" ]; then
                case "$exit_code" in
                    1) log_warning "  → Generic wget error - check your internet connection" ;;
                    2) log_warning "  → Parse error - invalid URL or response" ;;
                    3) log_warning "  → File I/O error - check disk space and permissions" ;;
                    4) log_warning "  → Network failure - check your internet connection" ;;
                    5) log_warning "  → SSL verification failure - update your certificates" ;;
                    6) log_warning "  → Username/password authentication failure" ;;
                    7) log_warning "  → Protocol error" ;;
                    8) log_warning "  → Server issued an error response" ;;
                    *) log_warning "  → Network error (wget exit code: $exit_code)" ;;
                esac
            fi
            ;;
    esac
    
    echo ""
    log_info "Alternative solutions:"
    log_info "  1. Manual download: https://github.com/kunalkushwaha/agenticgokit/releases"
    log_info "  2. Try a different version: --version v0.3.0"
    log_info "  3. Use Go install: go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest"
    log_info "  4. Check your internet connection and try again"
}

install_agenticgokit() {
    log_info "AgenticGoKit CLI Installer"
    log_info "=========================="
    echo ""
    
    # Get version to install
    if [ "$VERSION" = "latest" ]; then
        TARGET_VERSION=$(get_latest_version)
        if [ -z "$TARGET_VERSION" ]; then
            log_error "Failed to determine latest version. Please specify a version manually."
            exit 1
        fi
    else
        TARGET_VERSION="$VERSION"
    fi
    
    echo -e "${WHITE}Target version: $TARGET_VERSION${NC}"
    
    # Determine OS and architecture
    OS=$(get_os)
    ARCH=$(get_arch)
    BINARY_NAME="agentcli-$OS-$ARCH"
    
    echo -e "${WHITE}Platform: $OS/$ARCH${NC}"
    
    # Construct download URL
    DOWNLOAD_URL="https://github.com/kunalkushwaha/agenticgokit/releases/download/$TARGET_VERSION/$BINARY_NAME"
    echo -e "${WHITE}Download URL: $DOWNLOAD_URL${NC}"
    
    # Determine installation directory
    INSTALL_DIR=$(get_install_dir)
    BINARY_PATH="$INSTALL_DIR/agentcli"
    
    echo -e "${WHITE}Installation directory: $INSTALL_DIR${NC}"
    
    # Check if already installed
    if [ -f "$BINARY_PATH" ]; then
        if [ "$FORCE" != "true" ]; then
            log_warning "AgenticGoKit is already installed at $BINARY_PATH"
            log_warning "Use --force to overwrite the existing installation"
            
            # Test current installation
            if test_installation "$BINARY_PATH"; then
                log_success "Current installation is working correctly."
                exit 0
            fi
        else
            log_warning "Overwriting existing installation..."
        fi
    fi
    
    # Create installation directory
    if [ ! -d "$INSTALL_DIR" ]; then
        log_info "Creating installation directory: $INSTALL_DIR"
        mkdir -p "$INSTALL_DIR"
    fi
    
    # Download binary
    log_info "Downloading $BINARY_NAME..."
    TEMP_FILE=$(mktemp)
    if ! download_file "$DOWNLOAD_URL" "$TEMP_FILE"; then
        rm -f "$TEMP_FILE"
        exit 1
    fi
    
    # Move to final location and make executable
    mv "$TEMP_FILE" "$BINARY_PATH"
    chmod +x "$BINARY_PATH"
    
    # Verify download
    if [ ! -f "$BINARY_PATH" ]; then
        log_error "Installation failed: Binary not found at $BINARY_PATH"
        exit 1
    fi
    
    FILE_SIZE=$(du -h "$BINARY_PATH" | cut -f1)
    log_success "Downloaded binary ($FILE_SIZE)"
    
    # Test installation
    log_info "Testing installation..."
    if test_installation "$BINARY_PATH"; then
        echo ""
        log_success "🎉 AgenticGoKit CLI installed successfully!"
        echo ""
        log_warning "NEXT STEPS:"
        
        # Check PATH
        if ! check_path "$INSTALL_DIR"; then
            echo ""
        fi
        
        echo -e "${WHITE}1. Run 'agentcli --help' to get started${NC}"
        echo -e "${WHITE}2. Enable shell completion:${NC}"
        case "$SHELL" in
            */bash) echo -e "${WHITE}   source <(agentcli completion bash)${NC}" ;;
            */zsh) echo -e "${WHITE}   agentcli completion zsh > \"\${fpath[1]}/_agentcli\"${NC}" ;;
            */fish) echo -e "${WHITE}   agentcli completion fish | source${NC}" ;;
            *) echo -e "${WHITE}   agentcli completion bash${NC}" ;;
        esac
        echo -e "${WHITE}3. Create your first project: agentcli create my-project --template basic${NC}"
        echo ""
        log_info "Documentation: https://github.com/kunalkushwaha/agenticgokit"
        exit 0
    else
        log_warning "Installation completed but verification failed."
        log_warning "You may need to restart your shell or check your PATH."
        exit 1
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -d|--dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Run the installer
install_agenticgokit