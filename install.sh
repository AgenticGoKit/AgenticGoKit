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
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "https://api.github.com/repos/kunalkushwaha/agenticgokit/releases/latest" | \
            grep '"tag_name":' | \
            sed -E 's/.*"tag_name":\s*"([^"]+)".*/\1/'
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "https://api.github.com/repos/kunalkushwaha/agenticgokit/releases/latest" | \
            grep '"tag_name":' | \
            sed -E 's/.*"tag_name":\s*"([^"]+)".*/\1/'
    else
        log_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi
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
    
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$url" -o "$output"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$url" -O "$output"
    else
        log_error "Neither curl nor wget is available. Please install one of them."
        return 1
    fi
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
        log_error "Error downloading binary from $DOWNLOAD_URL"
        log_warning "Please check the version and try again."
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