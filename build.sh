#!/bin/bash
# Bash build script for AgenticGoKit
# Alternative to Makefile for Linux/macOS users

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Version information
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
# RFC3339 format for consistent parsing in version.go
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS="-X github.com/kunalkushwaha/agenticgokit/cmd/agentcli/version.Version=$VERSION \
         -X github.com/kunalkushwaha/agenticgokit/cmd/agentcli/version.GitCommit=$GIT_COMMIT \
         -X github.com/kunalkushwaha/agenticgokit/cmd/agentcli/version.GitBranch=$GIT_BRANCH \
         -X github.com/kunalkushwaha/agenticgokit/cmd/agentcli/version.BuildDate=$BUILD_DATE"

show_help() {
    echo -e "${CYAN}AgenticGoKit Build Script (Bash)${NC}"
    echo -e "${CYAN}================================${NC}"
    echo ""
    echo -e "${YELLOW}USAGE:${NC}"
    echo -e "  ${WHITE}./build.sh [target]${NC}"
    echo ""
    echo -e "${YELLOW}TARGETS:${NC}"
    echo -e "  ${WHITE}build         - Build for current platform (default)${NC}"
    echo -e "  ${WHITE}dev           - Quick development build${NC}"
    echo -e "  ${WHITE}release       - Optimized release build${NC}"
    echo -e "  ${WHITE}clean         - Clean build artifacts${NC}"
    echo -e "  ${WHITE}test          - Run tests${NC}"
    echo -e "  ${WHITE}linux         - Cross-compile for Linux${NC}"
    echo -e "  ${WHITE}darwin        - Cross-compile for macOS${NC}"
    echo -e "  ${WHITE}windows       - Cross-compile for Windows${NC}"
    echo -e "  ${WHITE}all           - Cross-compile for all platforms${NC}"
    echo -e "  ${WHITE}version       - Show version information${NC}"
    echo ""
    echo -e "${YELLOW}EXAMPLES:${NC}"
    echo -e "  ${WHITE}./build.sh dev       ${NC}# Quick development build"
    echo -e "  ${WHITE}./build.sh all       ${NC}# Cross-compile for all platforms"
    echo -e "  ${WHITE}./build.sh release   ${NC}# Optimized production build"
}

build_current() {
    echo -e "${GREEN}Building agentcli for $(uname -s)...${NC}"
    go build -ldflags "$LDFLAGS" -o agentcli ./cmd/agentcli
    echo -e "${GREEN}✓ Build successful: agentcli${NC}"
}

build_dev() {
    echo -e "${GREEN}Building development version...${NC}"
    go build -o agentcli ./cmd/agentcli
    echo -e "${GREEN}✓ Development build successful: agentcli${NC}"
}

build_release() {
    echo -e "${GREEN}Building optimized release version...${NC}"
    go build -ldflags "$LDFLAGS -s -w" -o agentcli ./cmd/agentcli
    echo -e "${GREEN}✓ Release build successful: agentcli${NC}"
}

build_linux() {
    echo -e "${GREEN}Cross-compiling for Linux...${NC}"
    GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o agentcli-linux-amd64 ./cmd/agentcli
    echo -e "${GREEN}✓ Linux build successful: agentcli-linux-amd64${NC}"
}

build_darwin() {
    echo -e "${GREEN}Cross-compiling for macOS...${NC}"
    
    # Intel Mac
    GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o agentcli-darwin-amd64 ./cmd/agentcli
    echo -e "${GREEN}✓ macOS Intel build successful: agentcli-darwin-amd64${NC}"
    
    # Apple Silicon Mac
    GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o agentcli-darwin-arm64 ./cmd/agentcli
    echo -e "${GREEN}✓ macOS Apple Silicon build successful: agentcli-darwin-arm64${NC}"
}

build_windows() {
    echo -e "${GREEN}Cross-compiling for Windows...${NC}"
    GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o agentcli-windows-amd64.exe ./cmd/agentcli
    echo -e "${GREEN}✓ Windows build successful: agentcli-windows-amd64.exe${NC}"
}

build_all() {
    echo -e "${CYAN}Cross-compiling for all platforms...${NC}"
    build_linux
    build_darwin
    build_windows
    echo -e "${CYAN}✓ All builds completed!${NC}"
}

clean_artifacts() {
    echo -e "${YELLOW}Cleaning build artifacts...${NC}"
    rm -f agentcli agentcli.exe agentcli-*
    go clean
    echo -e "${GREEN}✓ Clean completed!${NC}"
}

run_tests() {
    echo -e "${GREEN}Running tests...${NC}"
    go test ./...
}

show_version() {
    echo -e "${CYAN}Build Information:${NC}"
    echo -e "  ${WHITE}OS: $(uname -s)${NC}"
    echo -e "  ${WHITE}Version: $VERSION${NC}"
    echo -e "  ${WHITE}Git Commit: $GIT_COMMIT${NC}"
    echo -e "  ${WHITE}Git Branch: $GIT_BRANCH${NC}"
    echo -e "  ${WHITE}Build Date: $BUILD_DATE${NC}"
}

# Main execution
TARGET=${1:-build}

case "$TARGET" in
    "build")
        build_current
        ;;
    "dev")
        build_dev
        ;;
    "release")
        build_release
        ;;
    "clean")
        clean_artifacts
        ;;
    "test")
        run_tests
        ;;
    "linux")
        build_linux
        ;;
    "darwin")
        build_darwin
        ;;
    "windows")
        build_windows
        ;;
    "all")
        build_all
        ;;
    "version")
        show_version
        ;;
    "help"|"-h"|"--help")
        show_help
        ;;
    *)
        echo -e "${RED}Unknown target: $TARGET${NC}"
        echo -e "${YELLOW}Use 'help' to see available targets${NC}"
        exit 1
        ;;
esac