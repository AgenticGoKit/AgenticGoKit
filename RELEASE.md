# Release Process Guide

This document describes how to create releases for AgenticGoKit CLI.

## 🚀 Automated Release Process (Recommended)

We use **GitHub Actions** for fully automated releases. Here's how it works:

### 1. Create a Release

```bash
# Make sure you're on main branch and up to date
git checkout main
git pull origin main

# Create and push a version tag
git tag v0.4.0
git push origin v0.4.0
```

### 2. Automatic Process

Once you push a tag, GitHub Actions automatically:

1. ✅ **Builds all platforms**: Linux, macOS, Windows (AMD64 + ARM64)
2. ✅ **Injects version info**: Version, commit, branch, build date (RFC3339 format)
3. ✅ **Creates checksums**: SHA256 for all binaries
4. ✅ **Generates release notes**: Professional release notes with installation instructions
5. ✅ **Creates GitHub release**: With all binaries attached
6. ✅ **Updates installation scripts**: Scripts automatically use the new version

### 3. What Gets Built

| Platform | Binary Name | Description |
|----------|-------------|-------------|
| Linux AMD64 | `agentcli-linux-amd64` | Most Linux distributions |
| Linux ARM64 | `agentcli-linux-arm64` | ARM Linux, Raspberry Pi |
| macOS Intel | `agentcli-darwin-amd64` | Intel-based Macs |
| macOS Apple Silicon | `agentcli-darwin-arm64` | M1/M2/M3 Macs |
| Windows AMD64 | `agentcli-windows-amd64.exe` | 64-bit Windows |
| Windows ARM64 | `agentcli-windows-arm64.exe` | ARM Windows devices |

### 4. Installation Scripts Update

The installation scripts automatically detect and use the latest release:
- `install.ps1` - Windows PowerShell installer
- `install.sh` - Linux/macOS Bash installer

Users can install immediately after release:
```bash
# Latest version (automatic)
curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.sh | bash

# Specific version
curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.sh | bash -s -- --version v0.4.0
```

## 📋 Release Checklist

### Pre-Release
- [ ] All tests passing on main branch
- [ ] Version updated in relevant files (if needed)
- [ ] CHANGELOG updated with new features/fixes
- [ ] Documentation updated
- [ ] Installation scripts tested

### Release
- [ ] Create version tag: `git tag v0.4.0`
- [ ] Push tag: `git push origin v0.4.0`
- [ ] Monitor GitHub Actions workflow
- [ ] Verify release created successfully
- [ ] Test installation scripts with new version

### Post-Release
- [ ] Announce release (if major)
- [ ] Update documentation if needed
- [ ] Close related issues/PRs

## 🏷️ Version Tagging

We follow [Semantic Versioning](https://semver.org/):

- **Major** (v1.0.0): Breaking changes
- **Minor** (v0.4.0): New features, backward compatible
- **Patch** (v0.3.1): Bug fixes, backward compatible

### Tag Format
```bash
# Correct format
git tag v0.4.0
git tag v1.0.0
git tag v0.3.1

# Incorrect format (don't use)
git tag 0.4.0      # Missing 'v' prefix
git tag version-0.4.0  # Wrong format
```

## 🔧 Manual Release (Not Recommended)

If you need to create a release manually:

### 1. Build All Platforms
```bash
# Using Makefile (recommended)
make build-all

# Or using build scripts
./build.sh all     # Linux/macOS
.\build.ps1 all    # Windows PowerShell
```

### 2. Create Checksums
```bash
sha256sum agentcli-* > checksums.txt
```

### 3. Create GitHub Release
1. Go to [GitHub Releases](https://github.com/kunalkushwaha/agenticgokit/releases)
2. Click "Create a new release"
3. Choose tag version
4. Upload all binaries and checksums.txt
5. Write release notes

## 🚨 Troubleshooting

### Release Workflow Failed
1. Check the [Actions tab](https://github.com/kunalkushwaha/agenticgokit/actions)
2. Look at the failed step logs
3. Common issues:
   - Build failures: Check Go code compilation
   - Permission issues: Verify GITHUB_TOKEN permissions
   - Tag format: Ensure tag starts with 'v'

### Installation Scripts Not Working
1. Test scripts locally:
   ```bash
   # Test bash script
   bash install.sh --help
   
   # Test PowerShell script
   pwsh install.ps1 -Help
   ```
2. Check GitHub API rate limits
3. Verify release assets are uploaded correctly

### Binary Issues
1. Test binaries locally:
   ```bash
   ./agentcli-linux-amd64 version
   ./agentcli-windows-amd64.exe version
   ```
2. Check cross-compilation settings in Makefile
3. Verify Go version compatibility

## 📊 Release Metrics

After each release, monitor:
- Download counts per platform
- Installation script usage
- User feedback and issues
- Performance metrics

## 🔄 Hotfix Process

For critical bug fixes:

1. Create hotfix branch from main:
   ```bash
   git checkout main
   git checkout -b hotfix/v0.3.1
   ```

2. Make minimal fix and test

3. Merge to main:
   ```bash
   git checkout main
   git merge hotfix/v0.3.1
   ```

4. Create patch release:
   ```bash
   git tag v0.3.1
   git push origin v0.3.1
   ```

## 📚 Related Documentation

- [BUILD.md](BUILD.md) - Build system documentation
- [INSTALL.md](INSTALL.md) - Installation guide
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
- [GitHub Actions Workflows](.github/workflows/) - CI/CD configuration

---

## 🎯 Quick Release Commands

```bash
# Standard release process
git checkout main
git pull origin main
git tag v0.4.0
git push origin v0.4.0

# Monitor release
open https://github.com/kunalkushwaha/agenticgokit/actions

# Test installation after release
curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.sh | bash -s -- --version v0.4.0
```

**That's it!** The automated system handles everything else. 🚀