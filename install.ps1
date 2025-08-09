# AgenticGoKit CLI Installation Script for Windows
# Usage: iwr -useb https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.ps1 | iex

param(
    [string]$Version = "latest",
    [string]$InstallDir = "",
    [switch]$Force,
    [switch]$Help
)

# Colors for output
$Red = "Red"
$Green = "Green"
$Yellow = "Yellow"
$Cyan = "Cyan"
$White = "White"

function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

function Show-Help {
    Write-ColorOutput "AgenticGoKit CLI Installer for Windows" $Cyan
    Write-ColorOutput "========================================" $Cyan
    Write-Host ""
    Write-ColorOutput "USAGE:" $Yellow
    Write-ColorOutput "  iwr -useb https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.ps1 | iex" $White
    Write-Host ""
    Write-ColorOutput "OPTIONS:" $Yellow
    Write-ColorOutput "  -Version <version>    Install specific version (default: latest)" $White
    Write-ColorOutput "  -InstallDir <path>    Custom installation directory" $White
    Write-ColorOutput "  -Force                Overwrite existing installation" $White
    Write-ColorOutput "  -Help                 Show this help message" $White
    Write-Host ""
    Write-ColorOutput "EXAMPLES:" $Yellow
    Write-ColorOutput "  # Install latest version" $White
    Write-ColorOutput "  iwr -useb https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.ps1 | iex" $White
    Write-Host ""
    Write-ColorOutput "  # Install specific version" $White
    Write-ColorOutput "  iwr -useb 'https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/main/install.ps1' | iex -Version v0.3.0" $White
}

function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { return "amd64" }
    }
}

function Get-LatestVersion {
    try {
        Write-ColorOutput "Fetching latest release information..." $Cyan
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/kunalkushwaha/agenticgokit/releases/latest" -Headers @{"User-Agent" = "AgenticGoKit-Installer"}
        return $response.tag_name
    }
    catch {
        Write-ColorOutput "Error fetching latest version: $($_.Exception.Message)" $Red
        Write-ColorOutput "Falling back to manual installation..." $Yellow
        return $null
    }
}

function Get-InstallDirectory {
    if ($InstallDir) {
        return $InstallDir
    }
    
    # Try to find a good installation directory
    $candidates = @(
        "$env:LOCALAPPDATA\Programs\AgenticGoKit",
        "$env:USERPROFILE\.agenticgokit\bin",
        "$env:USERPROFILE\bin"
    )
    
    foreach ($candidate in $candidates) {
        if (Test-Path $candidate -PathType Container) {
            return $candidate
        }
    }
    
    # Default to first candidate
    return $candidates[0]
}

function Add-ToPath {
    param([string]$Directory)
    
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($currentPath -notlike "*$Directory*") {
        Write-ColorOutput "Adding $Directory to PATH..." $Cyan
        $newPath = "$currentPath;$Directory"
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        $env:PATH = "$env:PATH;$Directory"
        return $true
    }
    return $false
}

function Test-Installation {
    param([string]$BinaryPath)
    
    try {
        $output = & $BinaryPath version --short 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-ColorOutput "✓ Installation verified: $output" $Green
            return $true
        }
    }
    catch {
        Write-ColorOutput "✗ Installation verification failed" $Red
        return $false
    }
    return $false
}

function Install-AgenticGoKit {
    # Show help if requested
    if ($Help) {
        Show-Help
        return
    }
    
    Write-ColorOutput "AgenticGoKit CLI Installer" $Cyan
    Write-ColorOutput "==========================" $Cyan
    Write-Host ""
    
    # Get version to install
    if ($Version -eq "latest") {
        $targetVersion = Get-LatestVersion
        if (-not $targetVersion) {
            Write-ColorOutput "Failed to determine latest version. Please specify a version manually." $Red
            return 1
        }
    } else {
        $targetVersion = $Version
    }
    
    Write-ColorOutput "Target version: $targetVersion" $White
    
    # Determine architecture and binary name
    $arch = Get-Architecture
    $binaryName = "agentcli-windows-$arch.exe"
    Write-ColorOutput "Architecture: $arch" $White
    
    # Construct download URL
    $downloadUrl = "https://github.com/kunalkushwaha/agenticgokit/releases/download/$targetVersion/$binaryName"
    Write-ColorOutput "Download URL: $downloadUrl" $White
    
    # Determine installation directory
    $installDir = Get-InstallDirectory
    $binaryPath = Join-Path $installDir "agentcli.exe"
    
    Write-ColorOutput "Installation directory: $installDir" $White
    
    # Check if already installed
    if (Test-Path $binaryPath -PathType Leaf) {
        if (-not $Force) {
            Write-ColorOutput "AgenticGoKit is already installed at $binaryPath" $Yellow
            Write-ColorOutput "Use -Force to overwrite the existing installation" $Yellow
            
            # Test current installation
            if (Test-Installation $binaryPath) {
                Write-ColorOutput "Current installation is working correctly." $Green
                return 0
            }
        } else {
            Write-ColorOutput "Overwriting existing installation..." $Yellow
        }
    }
    
    # Create installation directory
    if (-not (Test-Path $installDir -PathType Container)) {
        Write-ColorOutput "Creating installation directory: $installDir" $Cyan
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    }
    
    # Download binary
    Write-ColorOutput "Downloading $binaryName..." $Cyan
    try {
        $progressPreference = 'SilentlyContinue'
        Invoke-WebRequest -Uri $downloadUrl -OutFile $binaryPath -Headers @{"User-Agent" = "AgenticGoKit-Installer"}
        $progressPreference = 'Continue'
    }
    catch {
        Write-ColorOutput "Error downloading binary: $($_.Exception.Message)" $Red
        Write-ColorOutput "Please check the version and try again." $Yellow
        return 1
    }
    
    # Verify download
    if (-not (Test-Path $binaryPath -PathType Leaf)) {
        Write-ColorOutput "Download failed: Binary not found at $binaryPath" $Red
        return 1
    }
    
    $fileSize = (Get-Item $binaryPath).Length
    Write-ColorOutput "Downloaded binary ($([math]::Round($fileSize/1MB, 2)) MB)" $Green
    
    # Add to PATH
    $pathAdded = Add-ToPath $installDir
    
    # Test installation
    Write-ColorOutput "Testing installation..." $Cyan
    if (Test-Installation $binaryPath) {
        Write-Host ""
        Write-ColorOutput "🎉 AgenticGoKit CLI installed successfully!" $Green
        Write-Host ""
        Write-ColorOutput "NEXT STEPS:" $Yellow
        if ($pathAdded) {
            Write-ColorOutput "1. Restart your PowerShell session to use the new PATH" $White
        }
        Write-ColorOutput "2. Run 'agentcli --help' to get started" $White
        Write-ColorOutput "3. Enable shell completion: agentcli completion powershell | Out-String | Invoke-Expression" $White
        Write-ColorOutput "4. Create your first project: agentcli create my-project --template basic" $White
        Write-Host ""
        Write-ColorOutput "Documentation: https://github.com/kunalkushwaha/agenticgokit" $Cyan
        return 0
    } else {
        Write-ColorOutput "Installation completed but verification failed." $Yellow
        Write-ColorOutput "You may need to restart your shell or check your PATH." $Yellow
        return 1
    }
}

# Run the installer
exit (Install-AgenticGoKit)