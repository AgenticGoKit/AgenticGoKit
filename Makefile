# Makefile for AgenticGoKit

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell powershell -NoProfile -Command "[System.DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ssZ')")

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
	@echo "Building agentcli with version $(VERSION)..."
	go build -ldflags "$(LDFLAGS)" -o agentcli.exe ./cmd/agentcli

# Build everything
build: build-cli
	@echo "Build complete!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f agentcli agentcli.exe
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
	@echo "Building development version..."
	go build -o agentcli.exe ./cmd/agentcli

# Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Git Branch: $(GIT_BRANCH)"
	@echo "Build Date: $(BUILD_DATE)"

# Cross-compilation targets
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o agentcli-linux-amd64 ./cmd/agentcli

build-darwin:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o agentcli-darwin-amd64 ./cmd/agentcli
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o agentcli-darwin-arm64 ./cmd/agentcli

build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o agentcli-windows-amd64.exe ./cmd/agentcli

# Build for all platforms
build-all: build-linux build-darwin build-windows

# Help target
help:
	@echo "Available targets:"
	@echo "  build         - Build the CLI (default)"
	@echo "  build-cli     - Build only the CLI"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run linter"
	@echo "  deps          - Install dependencies"
	@echo "  dev           - Quick development build"
	@echo "  version       - Show version information"
	@echo "  build-linux   - Cross-compile for Linux"
	@echo "  build-darwin  - Cross-compile for macOS"
	@echo "  build-windows - Cross-compile for Windows"
	@echo "  build-all     - Cross-compile for all platforms"
	@echo "  help          - Show this help message"