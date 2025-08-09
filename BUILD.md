# AgenticGoKit Build System

This document describes the various ways to build AgenticGoKit CLI across different platforms.

## 🚀 Quick Start

### Windows (PowerShell)
```powershell
# Quick development build
.\build.ps1 dev

# Cross-compile for all platforms
.\build.ps1 all

# Show help
.\build.ps1 -Help
```

### Linux/macOS (Bash)
```bash
# Quick development build
./build.sh dev

# Cross-compile for all platforms
./build.sh all

# Show help
./build.sh help
```

### Make (Cross-platform)
```bash
# Quick development build
make dev

# Cross-compile for all platforms
make build-all

# Show help
make help
```

## 📋 Build Options

### Build Targets

| Target | Description | Output |
|--------|-------------|---------|
| `build` | Build for current platform | `agentcli` / `agentcli.exe` |
| `dev` | Quick development build (no version info) | `agentcli` / `agentcli.exe` |
| `release` | Optimized production build | `agentcli` / `agentcli.exe` |
| `clean` | Remove all build artifacts | - |
| `test` | Run test suite | - |
| `version` | Show build information | - |

### Cross-Compilation Targets

| Target | Description | Output |
|--------|-------------|---------|
| `linux` | Linux AMD64 | `agentcli-linux-amd64` |
| `darwin` | macOS (Intel + Apple Silicon) | `agentcli-darwin-amd64`, `agentcli-darwin-arm64` |
| `windows` | Windows AMD64 | `agentcli-windows-amd64.exe` |
| `all` | All main platforms | All above binaries |

### Extended Targets (Makefile only)

| Target | Description | Output |
|--------|-------------|---------|
| `build-linux-arm64` | Linux ARM64 | `agentcli-linux-arm64` |
| `build-windows-arm64` | Windows ARM64 | `agentcli-windows-arm64.exe` |
| `build-all-extended` | All platforms including ARM64 | All binaries |
| `install` | Install to GOPATH/bin | - |
| `quick` | Build and test | - |

## 🛠️ Build Methods

### 1. PowerShell Script (Windows)

**Best for:** Windows users who prefer PowerShell

```powershell
# Basic usage
.\build.ps1 [target]

# Examples
.\build.ps1 build          # Build for Windows
.\build.ps1 dev            # Quick development build
.\build.ps1 all            # Cross-compile for all platforms
.\build.ps1 clean          # Clean build artifacts
.\build.ps1 version        # Show version info
```

**Features:**
- ✅ Colored output
- ✅ Proper error handling
- ✅ Cross-compilation support
- ✅ Version injection
- ✅ Windows-native commands

### 2. Bash Script (Linux/macOS)

**Best for:** Linux/macOS users or Windows with Git Bash

```bash
# Make executable (first time only)
chmod +x build.sh

# Basic usage
./build.sh [target]

# Examples
./build.sh build          # Build for current platform
./build.sh dev             # Quick development build
./build.sh all             # Cross-compile for all platforms
./build.sh clean           # Clean build artifacts
./build.sh version         # Show version info
```

**Features:**
- ✅ Colored output
- ✅ Unix-style commands
- ✅ Cross-compilation support
- ✅ Version injection
- ✅ POSIX compliant

### 3. Makefile (Universal)

**Best for:** Developers familiar with Make, CI/CD systems

```bash
# Basic usage
make [target]

# Examples
make build                 # Build for current platform
make dev                   # Quick development build
make build-all             # Cross-compile for all platforms
make build-all-extended    # Include ARM64 builds
make clean                 # Clean build artifacts
make install               # Install to GOPATH/bin
make help                  # Show detailed help
```

**Features:**
- ✅ Cross-platform compatibility
- ✅ Advanced targets (install, extended builds)
- ✅ Dependency management
- ✅ Parallel builds
- ✅ CI/CD friendly

## 🔧 Version Information

All build methods inject the following version information:

- **Version**: Git tag or "dev" for development builds
- **Git Commit**: Full commit hash
- **Git Branch**: Current branch name
- **Build Date**: UTC timestamp in RFC3339 format (e.g., "2006-01-02T15:04:05Z")
- **Go Version**: Go compiler version used
- **Platform**: Target OS and architecture

### Build Date Format

The build date is consistently formatted in **RFC3339** format across all build systems:
- **Format**: `YYYY-MM-DDTHH:MM:SSZ`
- **Example**: `2024-01-15T10:30:45Z`
- **Default**: `1970-01-01T00:00:00Z` (when build date cannot be determined)

This ensures consistent parsing in the version display logic.

### Dynamic Version Resolution for Scaffolding

The CLI automatically determines the correct AgenticGoKit library version to use in generated projects:

1. **CLI Version** (preferred): Uses the CLI's own version if it's a proper semantic version
2. **GitHub API** (fallback): Fetches the latest release from GitHub API
3. **Stable Fallback** (last resort): Uses a known stable version (v0.3.4)

This ensures generated projects always use compatible library versions and stay up-to-date with releases.

View version information:
```bash
# Using built binary
./agentcli version --output detailed

# Using build tools
make version           # Makefile
./build.ps1 version    # PowerShell
./build.sh version     # Bash
```

## 📦 Output Files

### Single Platform Builds
- **Windows**: `agentcli.exe`
- **Linux/macOS**: `agentcli`

### Cross-Compilation Builds
- **Linux AMD64**: `agentcli-linux-amd64`
- **Linux ARM64**: `agentcli-linux-arm64` (Makefile only)
- **macOS Intel**: `agentcli-darwin-amd64`
- **macOS Apple Silicon**: `agentcli-darwin-arm64`
- **Windows AMD64**: `agentcli-windows-amd64.exe`
- **Windows ARM64**: `agentcli-windows-arm64.exe` (Makefile only)

## 📦 Installation Scripts

For end users who don't want to build from source, we provide one-line installation scripts:

### PowerShell (Windows)
```powershell
iwr -useb https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/master/install.ps1 | iex
```

### Bash (Linux/macOS)
```bash
curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/master/install.sh | bash
```

### Script Features
- ✅ **Zero Dependencies**: No need to clone repo or install build tools
- ✅ **Automatic Detection**: OS, architecture, and latest version
- ✅ **Smart Installation**: Finds best directory, manages PATH
- ✅ **Version Control**: Install latest or specific versions
- ✅ **Verification**: Tests installation after download
- ✅ **Professional UX**: Colored output, progress indicators

### Advanced Usage
```bash
# Install specific version
curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/master/install.sh | bash -s -- --version v0.3.0

# Install to custom directory
curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/master/install.sh | bash -s -- --dir /usr/local/bin

# Force overwrite existing
curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/master/install.sh | bash -s -- --force
```

See [INSTALL.md](INSTALL.md) for complete installation documentation.

## 🚀 CI/CD Integration

We have comprehensive GitHub Actions workflows for automated building and releasing:

### Workflows
- **[CI Workflow](.github/workflows/ci.yml)**: Tests and builds on every PR/push
- **[Release Workflow](.github/workflows/release.yml)**: Automated releases on version tags
- **[Install Scripts Update](.github/workflows/update-install-scripts.yml)**: Maintains installation scripts

### Release Process
```bash
# Create and push a version tag
git tag v0.4.0
git push origin v0.4.0

# GitHub Actions automatically:
# 1. Builds all platforms
# 2. Creates checksums
# 3. Generates release notes
# 4. Creates GitHub release with binaries
```

### Custom GitHub Actions Example
```yaml
- name: Build all platforms
  run: make build-all

- name: Upload artifacts
  uses: actions/upload-artifact@v3
  with:
    name: agentcli-binaries
    path: agentcli-*
```

See [RELEASE.md](RELEASE.md) for complete release documentation.

### Docker Build Example
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN make build-linux

FROM alpine:latest
COPY --from=builder /app/agentcli-linux-amd64 /usr/local/bin/agentcli
```

## 🔍 Troubleshooting

### Common Issues

**Build fails with "command not found"**
- Ensure Go is installed and in PATH
- Check Go version: `go version` (requires Go 1.21+)

**Cross-compilation fails**
- Ensure target platform is supported by Go
- Check GOOS/GOARCH values are valid

**Version shows as "unknown"**
- Ensure you're in a Git repository
- Check Git is installed and accessible

**PowerShell execution policy error**
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Platform-Specific Notes

**Windows:**
- PowerShell script requires PowerShell 5.0+
- Make requires GNU Make (available with Git for Windows)

**macOS:**
- Bash script requires Bash 4.0+
- Make is included with Xcode Command Line Tools

**Linux:**
- All build methods should work out of the box
- Ensure `make` is installed for Makefile usage

## 📚 Related Documentation

- [CLI Reference](docs/reference/cli.md) - Complete CLI documentation
- [Shell Completion](docs/reference/cli-quick-reference.md#shell-completion) - Tab completion setup
- [Development Guide](docs/guides/development/README.md) - Development workflow
- [Contributing](CONTRIBUTING.md) - Contribution guidelines

## 🤝 Contributing

When contributing build system improvements:

1. Test all three build methods (PowerShell, Bash, Make)
2. Ensure cross-platform compatibility
3. Update this documentation
4. Test on Windows, macOS, and Linux if possible

For questions or issues with the build system, please open an issue on GitHub.
