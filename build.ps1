# PowerShell build script for AgenticGoKit
# Alternative to Makefile for Windows users

param(
    [string]$Target = "build",
    [switch]$Help
)

# Version information
$VERSION = try { git describe --tags --always --dirty 2>$null } catch { "dev" }
$GIT_COMMIT = try { git rev-parse HEAD 2>$null } catch { "unknown" }
$GIT_BRANCH = try { git rev-parse --abbrev-ref HEAD 2>$null } catch { "unknown" }
$BUILD_DATE = (Get-Date).ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ')

# Build flags
$LDFLAGS = @(
    "-X github.com/kunalkushwaha/agenticgokit/cmd/agentcli/version.Version=$VERSION"
    "-X github.com/kunalkushwaha/agenticgokit/cmd/agentcli/version.GitCommit=$GIT_COMMIT"
    "-X github.com/kunalkushwaha/agenticgokit/cmd/agentcli/version.GitBranch=$GIT_BRANCH"
    "-X github.com/kunalkushwaha/agenticgokit/cmd/agentcli/version.BuildDate=$BUILD_DATE"
) -join " "

function Show-Help {
    Write-Host "AgenticGoKit Build Script (PowerShell)" -ForegroundColor Cyan
    Write-Host "=====================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "USAGE:" -ForegroundColor Yellow
    Write-Host "  .\build.ps1 [target]" -ForegroundColor White
    Write-Host ""
    Write-Host "TARGETS:" -ForegroundColor Yellow
    Write-Host "  build         - Build for current platform (default)" -ForegroundColor White
    Write-Host "  dev           - Quick development build" -ForegroundColor White
    Write-Host "  release       - Optimized release build" -ForegroundColor White
    Write-Host "  clean         - Clean build artifacts" -ForegroundColor White
    Write-Host "  test          - Run tests" -ForegroundColor White
    Write-Host "  linux         - Cross-compile for Linux" -ForegroundColor White
    Write-Host "  darwin        - Cross-compile for macOS" -ForegroundColor White
    Write-Host "  windows       - Cross-compile for Windows" -ForegroundColor White
    Write-Host "  all           - Cross-compile for all platforms" -ForegroundColor White
    Write-Host "  version       - Show version information" -ForegroundColor White
    Write-Host ""
    Write-Host "EXAMPLES:" -ForegroundColor Yellow
    Write-Host "  .\build.ps1 dev       # Quick development build" -ForegroundColor Gray
    Write-Host "  .\build.ps1 all       # Cross-compile for all platforms" -ForegroundColor Gray
    Write-Host "  .\build.ps1 release   # Optimized production build" -ForegroundColor Gray
}

function Build-Current {
    Write-Host "Building agentcli for Windows..." -ForegroundColor Green
    go build -ldflags "$LDFLAGS" -o agentcli.exe ./cmd/agentcli
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Build successful: agentcli.exe" -ForegroundColor Green
    }
}

function Build-Dev {
    Write-Host "Building development version..." -ForegroundColor Green
    go build -o agentcli.exe ./cmd/agentcli
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Development build successful: agentcli.exe" -ForegroundColor Green
    }
}

function Build-Release {
    Write-Host "Building optimized release version..." -ForegroundColor Green
    go build -ldflags "$LDFLAGS -s -w" -o agentcli.exe ./cmd/agentcli
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Release build successful: agentcli.exe" -ForegroundColor Green
    }
}

function Build-Linux {
    Write-Host "Cross-compiling for Linux..." -ForegroundColor Green
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    go build -ldflags "$LDFLAGS" -o agentcli-linux-amd64 ./cmd/agentcli
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Linux build successful: agentcli-linux-amd64" -ForegroundColor Green
    }
}

function Build-Darwin {
    Write-Host "Cross-compiling for macOS..." -ForegroundColor Green
    
    # Intel Mac
    $env:GOOS = "darwin"
    $env:GOARCH = "amd64"
    go build -ldflags "$LDFLAGS" -o agentcli-darwin-amd64 ./cmd/agentcli
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ macOS Intel build successful: agentcli-darwin-amd64" -ForegroundColor Green
    }
    
    # Apple Silicon Mac
    $env:GOOS = "darwin"
    $env:GOARCH = "arm64"
    go build -ldflags "$LDFLAGS" -o agentcli-darwin-arm64 ./cmd/agentcli
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ macOS Apple Silicon build successful: agentcli-darwin-arm64" -ForegroundColor Green
    }
}

function Build-Windows {
    Write-Host "Cross-compiling for Windows..." -ForegroundColor Green
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    go build -ldflags "$LDFLAGS" -o agentcli-windows-amd64.exe ./cmd/agentcli
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Windows build successful: agentcli-windows-amd64.exe" -ForegroundColor Green
    }
}

function Build-All {
    Write-Host "Cross-compiling for all platforms..." -ForegroundColor Cyan
    Build-Linux
    Build-Darwin
    Build-Windows
    Write-Host "✓ All builds completed!" -ForegroundColor Cyan
}

function Clean-Artifacts {
    Write-Host "Cleaning build artifacts..." -ForegroundColor Yellow
    Remove-Item -Path "agentcli.exe", "agentcli-*" -ErrorAction SilentlyContinue
    go clean
    Write-Host "✓ Clean completed!" -ForegroundColor Green
}

function Run-Tests {
    Write-Host "Running tests..." -ForegroundColor Green
    go test ./...
}

function Show-Version {
    Write-Host "Build Information:" -ForegroundColor Cyan
    Write-Host "  OS: Windows" -ForegroundColor White
    Write-Host "  Version: $VERSION" -ForegroundColor White
    Write-Host "  Git Commit: $GIT_COMMIT" -ForegroundColor White
    Write-Host "  Git Branch: $GIT_BRANCH" -ForegroundColor White
    Write-Host "  Build Date: $BUILD_DATE" -ForegroundColor White
}

# Main execution
if ($Help) {
    Show-Help
    exit 0
}

switch ($Target.ToLower()) {
    "build" { Build-Current }
    "dev" { Build-Dev }
    "release" { Build-Release }
    "clean" { Clean-Artifacts }
    "test" { Run-Tests }
    "linux" { Build-Linux }
    "darwin" { Build-Darwin }
    "windows" { Build-Windows }
    "all" { Build-All }
    "version" { Show-Version }
    default {
        Write-Host "Unknown target: $Target" -ForegroundColor Red
        Write-Host "Use -Help to see available targets" -ForegroundColor Yellow
        exit 1
    }
}