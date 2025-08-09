# Makefile for AgenticGoKit
# Cross-platform Makefile that works on Windows (PowerShell/CMD), Linux, and macOS

# Detect operating system
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    EXE_EXT := .exe
    RM_CMD := del /Q
    DATE_CMD := powershell -NoProfile -Command "(Get-Date).ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ')"
    ENV_SET := set
    ENV_SEP := &&
else
    DETECTED_OS := $(shell uname -s)
    EXE_EXT := 
    RM_CMD := rm -f
    DATE_CMD := date -u +"%Y-%m-%dT%H:%M:%SZ"
    ENV_SET := 
    ENV_SEP := 
endif

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell $(DATE_CMD) 2>/dev/null || echo "1970-01-01T00:00:00Z")

# Build flags
LDFLAGS = -X github.com/kunalkushwaha/agenticgokit/cmd/agentcli/version.Version=$(VERSION) \
          -X github.com/kunalkushwaha/agenticgokit/cmd/agentcli/version.GitCommit=$(GIT_COMMIT) \
          -X github.com/kunalkushwaha/agenticgokit/cmd/agentcli/version.GitBranch=$(GIT_BRANCH) \
          -X github.com/kunalkushwaha/agenticgokit/cmd/agentcli/version.BuildDate=$(BUILD_DATE)

# Build targets
.PHONY: build build-cli clean test lint help

# Default target
all: build

# Build the CLI
build-cli:
	@echo "Building agentcli with version $(VERSION) on $(DETECTED_OS)..."
	go build -ldflags "$(LDFLAGS)" -o agentcli$(EXE_EXT) ./cmd/agentcli

# Build everything
build: build-cli
	@echo "Build complete!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts on $(DETECTED_OS)..."
ifeq ($(OS),Windows_NT)
	-$(RM_CMD) agentcli agentcli.exe agentcli-*.exe agentcli-* 2>nul || echo "No files to clean"
else
	-$(RM_CMD) agentcli agentcli.exe agentcli-*
endif
	go clean

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint code
lint:
	@echo "Running linter..."
	golangci-lint run

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Development build (faster, no version injection)
dev:
	@echo "Building development version on $(DETECTED_OS)..."
	go build -o agentcli$(EXE_EXT) ./cmd/agentcli

# Show version information
version:
	@echo "Build Information:"
	@echo "  OS: $(DETECTED_OS)"
	@echo "  Version: $(VERSION)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Git Branch: $(GIT_BRANCH)"
	@echo "  Build Date: $(BUILD_DATE)"

# Cross-compilation targets
build-linux:
	@echo "Building for linux/amd64..."
ifeq ($(OS),Windows_NT)
	$(ENV_SET) GOOS=linux$(ENV_SEP) $(ENV_SET) GOARCH=amd64$(ENV_SEP) go build -ldflags "$(LDFLAGS)" -o agentcli-linux-amd64 ./cmd/agentcli
else
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o agentcli-linux-amd64 ./cmd/agentcli
endif

build-linux-arm64:
	@echo "Building for linux/arm64..."
ifeq ($(OS),Windows_NT)
	$(ENV_SET) GOOS=linux$(ENV_SEP) $(ENV_SET) GOARCH=arm64$(ENV_SEP) go build -ldflags "$(LDFLAGS)" -o agentcli-linux-arm64 ./cmd/agentcli
else
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o agentcli-linux-arm64 ./cmd/agentcli
endif

build-darwin:
	@echo "Building for darwin/amd64..."
ifeq ($(OS),Windows_NT)
	$(ENV_SET) GOOS=darwin$(ENV_SEP) $(ENV_SET) GOARCH=amd64$(ENV_SEP) go build -ldflags "$(LDFLAGS)" -o agentcli-darwin-amd64 ./cmd/agentcli
else
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o agentcli-darwin-amd64 ./cmd/agentcli
endif
	@echo "Building for darwin/arm64..."
ifeq ($(OS),Windows_NT)
	$(ENV_SET) GOOS=darwin$(ENV_SEP) $(ENV_SET) GOARCH=arm64$(ENV_SEP) go build -ldflags "$(LDFLAGS)" -o agentcli-darwin-arm64 ./cmd/agentcli
else
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o agentcli-darwin-arm64 ./cmd/agentcli
endif

build-windows:
	@echo "Building for windows/amd64..."
ifeq ($(OS),Windows_NT)
	$(ENV_SET) GOOS=windows$(ENV_SEP) $(ENV_SET) GOARCH=amd64$(ENV_SEP) go build -ldflags "$(LDFLAGS)" -o agentcli-windows-amd64.exe ./cmd/agentcli
else
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o agentcli-windows-amd64.exe ./cmd/agentcli
endif

build-windows-arm64:
	@echo "Building for windows/arm64..."
ifeq ($(OS),Windows_NT)
	$(ENV_SET) GOOS=windows$(ENV_SEP) $(ENV_SET) GOARCH=arm64$(ENV_SEP) go build -ldflags "$(LDFLAGS)" -o agentcli-windows-arm64.exe ./cmd/agentcli
else
	GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o agentcli-windows-arm64.exe ./cmd/agentcli
endif

# Build for all platforms
build-all: build-linux build-darwin build-windows

# Build for all platforms including ARM64
build-all-extended: build-linux build-linux-arm64 build-darwin build-windows build-windows-arm64

# Install the CLI to GOPATH/bin
install:
	@echo "Installing agentcli to GOPATH/bin..."
	go install -ldflags "$(LDFLAGS)" ./cmd/agentcli

# Quick build and test
quick: dev test

# Release build with optimizations
release:
	@echo "Building optimized release version..."
	go build -ldflags "$(LDFLAGS) -s -w" -o agentcli$(EXE_EXT) ./cmd/agentcli

# Help target
help:
	@echo "AgenticGoKit Build System ($(DETECTED_OS))"
	@echo "==========================================="
	@echo ""
	@echo "BASIC TARGETS:"
	@echo "  build              - Build the CLI (default)"
	@echo "  dev                - Quick development build (no version injection)"
	@echo "  release            - Optimized release build"
	@echo "  clean              - Clean build artifacts"
	@echo "  install            - Install to GOPATH/bin"
	@echo ""
	@echo "TESTING:"
	@echo "  test               - Run tests"
	@echo "  test-coverage      - Run tests with coverage report"
	@echo "  lint               - Run linter"
	@echo "  quick              - Quick build and test"
	@echo ""
	@echo "CROSS-COMPILATION:"
	@echo "  build-linux        - Build for Linux AMD64"
	@echo "  build-linux-arm64  - Build for Linux ARM64"
	@echo "  build-darwin       - Build for macOS (Intel + Apple Silicon)"
	@echo "  build-windows      - Build for Windows AMD64"
	@echo "  build-windows-arm64- Build for Windows ARM64"
	@echo "  build-all          - Build for all main platforms"
	@echo "  build-all-extended - Build for all platforms including ARM64"
	@echo ""
	@echo "UTILITIES:"
	@echo "  deps               - Install dependencies"
	@echo "  version            - Show version information"
	@echo "  help               - Show this help message"
	@echo ""
	@echo "EXAMPLES:"
	@echo "  make dev           # Quick development build"
	@echo "  make build-all     # Cross-compile for all platforms"
	@echo "  make release       # Optimized production build"