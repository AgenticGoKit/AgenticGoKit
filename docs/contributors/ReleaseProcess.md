# Release Process

This document describes the comprehensive release process for AgentFlow, including versioning, testing, documentation, and deployment procedures.

## 🎯 Release Philosophy

AgentFlow follows a disciplined release process to ensure:
- **Stability**: Thorough testing before releases
- **Predictability**: Regular release schedule and clear versioning
- **Transparency**: Open communication about changes and timeline
- **Quality**: Comprehensive validation and rollback capabilities

## 📅 Release Schedule

### Regular Release Cycle

| Release Type | Frequency | Content | Timeline |
|-------------|-----------|---------|----------|
| **Major** | Every 6 months | Breaking changes, major features | January, July |
| **Minor** | Monthly | New features, improvements | First Tuesday of each month |
| **Patch** | As needed | Bug fixes, security updates | Within 1-2 weeks of issue |
| **Pre-release** | Weekly | Development snapshots | Every Friday |

### Special Releases

- **Security Releases**: Immediate for critical security issues
- **Hotfixes**: Same day for critical production bugs
- **LTS Releases**: Long-term support versions (annual)

## 📝 Versioning Strategy

### Semantic Versioning (SemVer)

AgentFlow uses semantic versioning: `MAJOR.MINOR.PATCH`

```
v2.3.1
│ │ │
│ │ └─ Patch: Bug fixes, security updates
│ └─── Minor: New features, improvements (backward compatible)
└───── Major: Breaking changes, architectural updates
```

### Version Examples

| Version | Type | Description |
|---------|------|-------------|
| `v1.0.0` | Major | Initial stable release |
| `v1.1.0` | Minor | Added MCP caching, new CLI commands |
| `v1.1.1` | Patch | Fixed memory leak in tool execution |
| `v2.0.0` | Major | New agent interface, breaking config changes |
| `v2.1.0-rc.1` | Pre-release | Release candidate with new features |
| `v1.5.0-lts` | LTS | Long-term support release |

### Pre-release Identifiers

- `alpha`: Early development, unstable
- `beta`: Feature complete, testing phase
- `rc`: Release candidate, production ready

Examples: `v1.2.0-alpha.1`, `v1.2.0-beta.2`, `v1.2.0-rc.1`

## 🚀 Release Process Workflow

### 1. Planning Phase (2 weeks before release)

#### Release Planning Meeting
- Review completed features and fixes
- Assess breaking changes and migration needs
- Plan documentation updates
- Set release timeline and responsibilities

#### Create Release Branch
```bash
git checkout main
git pull origin main
git checkout -b release/v1.2.0
git push origin release/v1.2.0
```

#### Update Version Information
```bash
# Update version in all relevant files
./scripts/update-version.sh v1.2.0

# Files typically updated:
# - go.mod
# - cmd/agentcli/main.go
# - docs/README.md
# - CHANGELOG.md
```

### 2. Development Freeze (1 week before release)

#### Code Freeze
- No new features added to release branch
- Only critical bug fixes allowed
- All changes require release manager approval

#### Pre-release Testing
```bash
# Create pre-release
git tag v1.2.0-rc.1
git push origin v1.2.0-rc.1

# Trigger CI/CD pipeline
gh workflow run release.yml --ref v1.2.0-rc.1
```

### 3. Release Testing (3-5 days)

#### Automated Testing
```yaml
# .github/workflows/release-testing.yml
name: Release Testing
on:
  push:
    tags:
      - 'v*-rc.*'

jobs:
  comprehensive-testing:
    runs-on: ubuntu-latest
    steps:
    - name: Unit Tests
      run: go test -v ./...
    
    - name: Integration Tests
      run: go test -v -tags=integration ./integration/...
    
    - name: Performance Tests
      run: go test -bench=. ./benchmarks/...
    
    - name: Security Scan
      uses: securecodewarrior/github-action-add-sarif@v1
      with:
        sarif-file: security-scan.sarif
    
    - name: Compatibility Tests
      run: ./scripts/test-compatibility.sh
    
    - name: Load Tests
      run: ./scripts/run-load-tests.sh
```

#### Manual Testing Checklist

##### Core Functionality
- [ ] Agent creation and execution
- [ ] MCP tool discovery and usage
- [ ] Multi-agent orchestration
- [ ] Configuration management
- [ ] Error handling and recovery

##### CLI Testing
- [ ] Project creation (`agentcli create`)
- [ ] MCP server management
- [ ] Cache operations
- [ ] Tracing and debugging
- [ ] Configuration commands

##### Integration Testing
- [ ] Azure OpenAI integration
- [ ] OpenAI integration
- [ ] Ollama integration
- [ ] Docker-based MCP servers
- [ ] File system operations

##### Performance Testing
- [ ] Memory usage under load
- [ ] Concurrent agent execution
- [ ] Tool execution performance
- [ ] Cache effectiveness

### 4. Documentation Update

#### Update Documentation
```bash
# Generate API documentation
go run ./cmd/gendocs

# Update user guides
./scripts/update-docs.sh v1.2.0

# Update examples
./scripts/update-examples.sh v1.2.0
```

#### Documentation Checklist
- [ ] API reference updated
- [ ] User guides reflect new features
- [ ] Migration guide for breaking changes
- [ ] Examples updated and tested
- [ ] CHANGELOG.md completed
- [ ] README.md version badges updated

### 5. Release Preparation

#### Create Release Notes
```markdown
# AgentFlow v1.2.0 Release Notes

## 🚀 New Features
- **MCP Caching**: Intelligent caching of tool results for improved performance
- **Enhanced CLI**: New `agentcli benchmark` command for performance testing
- **Multi-Provider Support**: Added Anthropic Claude integration

## 🔧 Improvements
- Improved error messages with contextual information
- Better handling of MCP server failures
- Enhanced configuration validation

## 🐛 Bug Fixes
- Fixed memory leak in long-running agent sessions
- Resolved race condition in concurrent tool execution
- Fixed configuration parsing edge cases

## 💥 Breaking Changes
- Changed signature of `AgentHandler.Run()` to include context
- Renamed configuration field `mcp.enabled` to `mcp.enable`
- Removed deprecated `LegacyAgent` interface

## 🔄 Migration Guide
See [Migration Guide](docs/migration/v1.1-to-v1.2.md) for detailed upgrade instructions.

## 📦 Downloads
- [Linux x64](releases/agentcli-linux-amd64.tar.gz)
- [macOS x64](releases/agentcli-darwin-amd64.tar.gz)
- [Windows x64](releases/agentcli-windows-amd64.zip)
```

#### Final Validation
```bash
# Run full test suite
make test-all

# Validate documentation
make docs-validate

# Check for security issues
make security-scan

# Verify build works on all platforms
make build-all
```

### 6. Release Execution

#### Create Final Release
```bash
# Merge release branch to main
git checkout main
git merge release/v1.2.0
git tag v1.2.0
git push origin main
git push origin v1.2.0

# Trigger release pipeline
gh workflow run release.yml --ref v1.2.0
```

#### Automated Release Pipeline
```yaml
# .github/workflows/release.yml
name: Release
on:
  push:
    tags:
      - 'v*'
      - '!v*-rc.*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout
      uses: actions/checkout@v4
    
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run tests
      run: make test-all
    
    - name: Build binaries
      run: make build-all
    
    - name: Generate checksums
      run: make checksums
    
    - name: Create GitHub release
      uses: softprops/action-gh-release@v1
      with:
        files: |
          dist/*
          checksums.txt
        body_path: RELEASE_NOTES.md
        draft: false
        prerelease: false
    
    - name: Update Homebrew formula
      run: ./scripts/update-homebrew.sh
    
    - name: Update documentation site
      run: ./scripts/deploy-docs.sh
    
    - name: Notify stakeholders
      run: ./scripts/notify-release.sh
```

### 7. Post-Release Activities

#### Immediate Post-Release (Same Day)
- [ ] Verify release artifacts are available
- [ ] Test installation from official sources
- [ ] Monitor for immediate issues
- [ ] Update project management boards
- [ ] Announce release on communication channels

#### Short-term Follow-up (1-3 days)
- [ ] Monitor error reports and metrics
- [ ] Address any critical issues with hotfixes
- [ ] Update documentation based on user feedback
- [ ] Review release process effectiveness

#### Long-term Follow-up (1-2 weeks)
- [ ] Analyze adoption metrics
- [ ] Collect user feedback
- [ ] Plan improvements for next release
- [ ] Update contributor documentation

## 🔧 Release Tools and Scripts

### Version Management Script
```bash
#!/bin/bash
# scripts/update-version.sh

set -e

NEW_VERSION=$1
if [[ -z "$NEW_VERSION" ]]; then
    echo "Usage: $0 <version>"
    exit 1
fi

echo "Updating version to $NEW_VERSION"

# Update version in Go files
sed -i "s/Version = \".*\"/Version = \"$NEW_VERSION\"/" cmd/agentcli/main.go

# Update version in documentation
sed -i "s/version: .*/version: $NEW_VERSION/" docs/README.md

# Update CHANGELOG
./scripts/update-changelog.sh "$NEW_VERSION"

echo "Version updated to $NEW_VERSION"
```

### Build Script
```bash
#!/bin/bash
# scripts/build-all.sh

set -e

VERSION=$(git describe --tags --always)
LDFLAGS="-X main.Version=$VERSION -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

mkdir -p dist

for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r OS ARCH <<< "$platform"
    output="dist/agentcli-$OS-$ARCH"
    
    if [[ "$OS" == "windows" ]]; then
        output="$output.exe"
    fi
    
    echo "Building for $OS/$ARCH..."
    GOOS=$OS GOARCH=$ARCH go build -ldflags "$LDFLAGS" -o "$output" ./cmd/agentcli
    
    # Create archive
    if [[ "$OS" == "windows" ]]; then
        zip -q "dist/agentcli-$OS-$ARCH.zip" "$output"
        rm "$output"
    else
        tar -czf "dist/agentcli-$OS-$ARCH.tar.gz" -C dist "$(basename "$output")"
        rm "$output"
    fi
done

echo "Build complete. Artifacts in dist/"
```

### Release Validation Script
```bash
#!/bin/bash
# scripts/validate-release.sh

set -e

VERSION=$1
ARTIFACTS_DIR="dist"

echo "Validating release $VERSION"

# Check all required artifacts exist
REQUIRED_ARTIFACTS=(
    "agentcli-linux-amd64.tar.gz"
    "agentcli-linux-arm64.tar.gz"
    "agentcli-darwin-amd64.tar.gz"
    "agentcli-darwin-arm64.tar.gz"
    "agentcli-windows-amd64.zip"
    "checksums.txt"
)

for artifact in "${REQUIRED_ARTIFACTS[@]}"; do
    if [[ ! -f "$ARTIFACTS_DIR/$artifact" ]]; then
        echo "ERROR: Missing artifact: $artifact"
        exit 1
    fi
done

# Validate checksums
cd "$ARTIFACTS_DIR"
sha256sum -c checksums.txt || {
    echo "ERROR: Checksum validation failed"
    exit 1
}

echo "Release validation passed"
```

## 🚨 Hotfix Process

### Emergency Hotfix Workflow

For critical production issues requiring immediate fixes:

```bash
# 1. Create hotfix branch from main
git checkout main
git checkout -b hotfix/v1.2.1

# 2. Apply minimal fix
# Edit necessary files
git add .
git commit -m "fix: critical security vulnerability in MCP authentication"

# 3. Update version
./scripts/update-version.sh v1.2.1

# 4. Run critical tests
make test-security
make test-integration

# 5. Create hotfix release
git tag v1.2.1
git push origin hotfix/v1.2.1
git push origin v1.2.1

# 6. Merge back to main and develop
git checkout main
git merge hotfix/v1.2.1
git push origin main

# 7. Trigger emergency release
gh workflow run hotfix-release.yml --ref v1.2.1
```

### Hotfix Criteria

Hotfixes are reserved for:
- **Security vulnerabilities** (especially those with public disclosure)
- **Data corruption or loss** issues
- **Complete system failures** affecting all users
- **Legal compliance** issues requiring immediate resolution

## 📊 Release Metrics and KPIs

### Success Metrics

| Metric | Target | Measurement |
|--------|---------|-------------|
| **Release Frequency** | Monthly minor releases | Calendar tracking |
| **Time to Release** | < 2 weeks from freeze | Process timing |
| **Release Quality** | < 5 critical bugs post-release | Issue tracking |
| **Adoption Rate** | > 50% upgrade within 30 days | Download metrics |
| **Rollback Rate** | < 5% of releases | Deployment tracking |

### Release Health Dashboard

Monitor key indicators:
- Build success rate
- Test pass rate
- Security scan results
- Performance regression alerts
- User adoption metrics
- Support ticket volume

## 🤝 Stakeholder Communication

### Internal Communication

#### Release Status Updates
Weekly updates during release cycle to:
- Engineering teams
- Product management
- QA/Testing teams
- DevOps/Infrastructure
- Support teams

#### Release Readiness Review
Final go/no-go meeting with:
- Release manager
- Engineering lead
- QA lead
- Product owner
- Support manager

### External Communication

#### Community Updates
- GitHub Discussions announcements
- Blog post for major releases
- Social media updates
- Newsletter inclusions

#### User Notifications
- In-app update notifications
- Documentation site banners
- CLI version check messages
- Community forum posts

This comprehensive release process ensures AgentFlow maintains high quality and reliability while delivering regular value to users and maintaining contributor confidence.
