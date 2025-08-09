# AgenticGoKit CLI Installation Guide

This guide provides multiple ways to install the AgenticGoKit CLI on your system.

## 🚀 Quick Installation

### Windows (PowerShell)
```powershell
iwr -useb https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.ps1 | iex
```

### Linux/macOS (Bash)
```bash
curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.sh | bash
```

## 📋 Installation Methods

### 1. One-Line Installers (Recommended)

#### Windows PowerShell
```powershell
# Install latest version
iwr -useb https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.ps1 | iex

# Install specific version
iwr -useb 'https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.ps1' | iex -Version v0.3.0

# Install to custom directory
iwr -useb 'https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.ps1' | iex -InstallDir 'C:\tools'

# Force overwrite existing installation
iwr -useb 'https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.ps1' | iex -Force
```

#### Linux/macOS Bash
```bash
# Install latest version
curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.sh | bash

# Install specific version
curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.sh | bash -s -- --version v0.3.0

# Install to custom directory
curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.sh | bash -s -- --dir /usr/local/bin

# Force overwrite existing installation
curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.sh | bash -s -- --force
```

### 2. Manual Download

1. Go to the [Releases page](https://github.com/kunalkushwaha/agenticgokit/releases)
2. Download the appropriate binary for your platform:
   - **Windows**: `agentcli-windows-amd64.exe`
   - **Linux**: `agentcli-linux-amd64`
   - **macOS Intel**: `agentcli-darwin-amd64`
   - **macOS Apple Silicon**: `agentcli-darwin-arm64`
3. Rename to `agentcli` (or `agentcli.exe` on Windows)
4. Make executable: `chmod +x agentcli` (Linux/macOS)
5. Move to a directory in your PATH

### 3. Build from Source

```bash
# Clone the repository
git clone https://github.com/kunalkushwaha/agenticgokit.git
cd agenticgokit

# Build using Go
go install ./cmd/agentcli

# Or build using Make/scripts
make build          # Cross-platform Makefile
./build.sh build    # Linux/macOS
.\build.ps1 build   # Windows PowerShell
```

### 4. Go Install (for Go developers)

```bash
go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest
```

## 🔧 Post-Installation Setup

### 1. Verify Installation
```bash
agentcli version
agentcli --help
```

### 2. Enable Shell Completion

#### Bash
```bash
# Load for current session
source <(agentcli completion bash)

# Install permanently
agentcli completion bash > /etc/bash_completion.d/agentcli  # Linux (requires sudo)
agentcli completion bash > $(brew --prefix)/etc/bash_completion.d/agentcli  # macOS
```

#### Zsh
```bash
# Enable completion support
echo "autoload -U compinit; compinit" >> ~/.zshrc

# Install completion
agentcli completion zsh > "${fpath[1]}/_agentcli"
source ~/.zshrc
```

#### PowerShell
```powershell
# Load for current session
agentcli completion powershell | Out-String | Invoke-Expression

# Install permanently
agentcli completion powershell > agentcli.ps1
# Add to your PowerShell profile
```

#### Fish
```bash
# Load for current session
agentcli completion fish | source

# Install permanently
agentcli completion fish > ~/.config/fish/completions/agentcli.fish
```

### 3. Create Your First Project
```bash
# Interactive setup (recommended for beginners)
agentcli create --interactive

# Quick start with template
agentcli create my-project --template basic

# Advanced RAG system
agentcli create knowledge-base --template rag-system --memory pgvector
```

## 📍 Installation Locations

### Windows
- **Default**: `%LOCALAPPDATA%\Programs\AgenticGoKit\agentcli.exe`
- **Alternative**: `%USERPROFILE%\.agenticgokit\bin\agentcli.exe`
- **Custom**: Use `-InstallDir` parameter

### Linux/macOS
- **System-wide**: `/usr/local/bin/agentcli` (requires sudo)
- **User-local**: `~/.local/bin/agentcli`
- **Homebrew**: `$(brew --prefix)/bin/agentcli`
- **Custom**: Use `--dir` parameter

## 🔍 Troubleshooting

The installation scripts provide detailed error messages and specific guidance for different failure scenarios:

- **GitHub API Issues**: Rate limiting, server errors, network problems, invalid JSON responses
- **Download Failures**: Missing versions, network timeouts, SSL/TLS errors, HTTP status codes
- **Permission Issues**: File system access, PATH modification
- **Platform Detection**: Unsupported OS/architecture combinations
- **Data Validation**: Version format validation, response structure validation

### Enhanced Error Handling Features

- **HTTP Status Code Detection**: Identifies specific GitHub API errors (403, 404, 5xx)
- **Network Error Classification**: Distinguishes between DNS, timeout, and SSL issues
- **Response Validation**: Ensures API responses contain valid JSON and expected fields
- **Version Format Validation**: Verifies extracted versions follow semantic versioning
- **Graceful Degradation**: Provides multiple fallback installation methods

### Common Issues

**"agentcli: command not found"**
- The installation directory is not in your PATH
- Restart your shell or add the directory to PATH manually
- On Windows, restart PowerShell after installation

**"Permission denied" on Linux/macOS**
- The binary is not executable: `chmod +x /path/to/agentcli`
- Or install to a user directory instead of system-wide

**"Execution policy" error on Windows**
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

**Download fails with SSL/TLS errors**
- Update your system's certificates
- Use manual download method instead

**"Version not found" error**
- Check the [releases page](https://github.com/kunalkushwaha/agenticgokit/releases) for available versions
- Use `latest` or omit version parameter for the latest release

**GitHub API rate limiting (HTTP 403)**
- Wait a few minutes and try again
- Specify a version manually: `--version v0.3.0`
- Use manual download method

**Network timeout errors**
- Check your internet connection
- Try again (downloads may be large)
- Use manual download method

**DNS resolution errors**
- Check your internet connection
- Verify you can access github.com
- Check your DNS settings

### Getting Help

- **Documentation**: [GitHub Repository](https://github.com/kunalkushwaha/agenticgokit)
- **Issues**: [GitHub Issues](https://github.com/kunalkushwaha/agenticgokit/issues)
- **CLI Help**: `agentcli --help` or `agentcli <command> --help`

## 🔄 Updating

To update to the latest version, simply run the installation command again:

```bash
# Windows
iwr -useb https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.ps1 | iex -Force

# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.sh | bash -s -- --force
```

## 🗑️ Uninstallation

To uninstall AgenticGoKit CLI:

1. Remove the binary:
   - **Windows**: Delete from installation directory (usually `%LOCALAPPDATA%\Programs\AgenticGoKit`)
   - **Linux/macOS**: `rm /usr/local/bin/agentcli` (or wherever installed)

2. Remove from PATH (if manually added)

3. Remove shell completion (optional):
   - **Bash**: Remove from `/etc/bash_completion.d/` or `$(brew --prefix)/etc/bash_completion.d/`
   - **Zsh**: Remove `_agentcli` from fpath directories
   - **Fish**: Remove from `~/.config/fish/completions/`

4. Remove configuration (optional):
   - `~/.agenticgokit/` directory

## 🎯 Next Steps

After installation:

1. **Read the documentation**: [CLI Reference](docs/reference/cli.md)
2. **Try the tutorials**: [Getting Started](docs/tutorials/getting-started/quickstart.md)
3. **Explore templates**: `agentcli template list`
4. **Join the community**: [GitHub Discussions](https://github.com/kunalkushwaha/agenticgokit/discussions)

Happy building with AgenticGoKit! 🚀