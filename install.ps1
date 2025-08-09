# AgenticGoKit CLI Installation Script for Windows
# Usage: iwr -useb https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/master/install.ps1 | iex

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
    Write-ColorOutput "  iwr -useb https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/master/install.ps1 | iex" $White
    Write-Host ""
    Write-ColorOutput "OPTIONS:" $Yellow
    Write-ColorOutput "  -Version <version>    Install specific version (default: latest)" $White
    Write-ColorOutput "  -InstallDir <path>    Custom installation directory" $White
    Write-ColorOutput "  -Force                Overwrite existing installation" $White
    Write-ColorOutput "  -Help                 Show this help message" $White
    Write-Host ""
    Write-ColorOutput "EXAMPLES:" $Yellow
    Write-ColorOutput "  # Install latest version" $White
    Write-ColorOutput "  iwr -useb https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/master/install.ps1 | iex" $White
    Write-Host ""
    Write-ColorOutput "  # Install specific version" $White
    Write-ColorOutput "  iwr -useb 'https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/master/install.ps1' | iex -Version v0.3.0" $White
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
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/kunalkushwaha/agenticgokit/releases/latest" -Headers @{"User-Agent" = "AgenticGoKit-Installer"} -TimeoutSec 30
        
        # Validate response structure
        if (-not $response) {
            Write-ColorOutput "Error: Received empty response from GitHub API" $Red
            Show-ApiErrorGuidance "empty_response"
            return $null
        }
        
        if (-not $response.tag_name) {
            Write-ColorOutput "Error: Invalid JSON response - missing tag_name field" $Red
            if ($response.message) {
                Write-ColorOutput "  → GitHub API message: $($response.message)" $Yellow
            }
            Show-ApiErrorGuidance "invalid_json"
            return $null
        }
        
        $version = $response.tag_name
        
        # Validate version format
        if ([string]::IsNullOrWhiteSpace($version) -or $version -eq "null") {
            Write-ColorOutput "Error: Invalid version value: '$version'" $Red
            Show-ApiErrorGuidance "invalid_version"
            return $null
        }
        
        # Basic version format validation (should be semantic version)
        if (-not ($version -match '^v?\d+\.\d+\.\d+')) {
            Write-ColorOutput "Warning: Unusual version format: '$version'" $Yellow
            Write-ColorOutput "  → Proceeding anyway, but this may cause issues" $Yellow
        }
        
        return $version
    }
    catch {
        $errorMessage = $_.Exception.Message
        $statusCode = $null
        
        # Try to extract HTTP status code for better error handling
        if ($_.Exception -is [System.Net.WebException]) {
            $statusCode = [int]$_.Exception.Response.StatusCode
        } elseif ($_.Exception.Response) {
            $statusCode = [int]$_.Exception.Response.StatusCode
        }
        
        Write-ColorOutput "Error fetching latest version from GitHub API:" $Red
        
        # Provide specific guidance based on error type
        switch ($statusCode) {
            403 {
                Write-ColorOutput "  → Rate limit exceeded or access forbidden (HTTP 403)" $Yellow
                Write-ColorOutput "  → GitHub API has rate limits for unauthenticated requests" $Yellow
                Write-ColorOutput "  → Try again in a few minutes or specify a version manually" $Yellow
                Write-ColorOutput "  → Example: -Version v0.3.0" $White
            }
            404 {
                Write-ColorOutput "  → Repository not found or releases not available (HTTP 404)" $Yellow
                Write-ColorOutput "  → The repository may have been moved or made private" $Yellow
                Write-ColorOutput "  → Check if the repository exists: https://github.com/kunalkushwaha/agenticgokit" $White
            }
            { $_ -ge 500 } {
                Write-ColorOutput "  → GitHub server error (HTTP $statusCode)" $Yellow
                Write-ColorOutput "  → This is a temporary issue with GitHub's servers" $Yellow
                Write-ColorOutput "  → Try again in a few minutes" $Yellow
            }
            default {
                if ($errorMessage -match "timeout|timed out") {
                    Write-ColorOutput "  → Network timeout - check your internet connection" $Yellow
                    Write-ColorOutput "  → GitHub API requests may take time, try again" $Yellow
                } elseif ($errorMessage -match "resolve|dns|name") {
                    Write-ColorOutput "  → DNS resolution failed - check your internet connection" $Yellow
                    Write-ColorOutput "  → Verify you can access github.com in your browser" $Yellow
                } elseif ($errorMessage -match "ssl|tls|certificate") {
                    Write-ColorOutput "  → SSL/TLS certificate error" $Yellow
                    Write-ColorOutput "  → Update your system certificates or try manual installation" $Yellow
                } else {
                    Write-ColorOutput "  → $errorMessage" $Yellow
                    Write-ColorOutput "  → This may be a temporary network or server issue" $Yellow
                }
            }
        }
        
        Show-ApiErrorGuidance "network_error"
        return $null
    }
}

function Show-ApiErrorGuidance {
    param([string]$ErrorType)
    
    Write-ColorOutput "" $White
    Write-ColorOutput "Alternative installation methods:" $Cyan
    Write-ColorOutput "  1. Specify version manually: -Version v0.3.0" $White
    Write-ColorOutput "  2. Manual download: https://github.com/kunalkushwaha/agenticgokit/releases" $White
    Write-ColorOutput "  3. Use Go install: go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest" $White
    Write-ColorOutput "  4. Check GitHub status: https://www.githubstatus.com/" $White
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
        Invoke-WebRequest -Uri $downloadUrl -OutFile $binaryPath -Headers @{"User-Agent" = "AgenticGoKit-Installer"} -TimeoutSec 300
        $progressPreference = 'Continue'
    }
    catch {
        $progressPreference = 'Continue'
        $errorMessage = $_.Exception.Message
        $statusCode = $null
        
        # Try to extract HTTP status code
        if ($_.Exception -is [System.Net.WebException]) {
            $statusCode = [int]$_.Exception.Response.StatusCode
        } elseif ($_.Exception.Response) {
            $statusCode = [int]$_.Exception.Response.StatusCode
        }
        
        Write-ColorOutput "Error downloading binary from $downloadUrl" $Red
        
        # Provide specific guidance based on error type
        switch ($statusCode) {
            404 {
                Write-ColorOutput "  → Binary not found (HTTP 404)" $Yellow
                Write-ColorOutput "  → The version '$targetVersion' may not exist or binaries aren't available" $Yellow
                Write-ColorOutput "  → Check available versions: https://github.com/kunalkushwaha/agenticgokit/releases" $White
                Write-ColorOutput "  → Try a different version: -Version v0.3.0" $White
            }
            403 {
                Write-ColorOutput "  → Access forbidden (HTTP 403)" $Yellow
                Write-ColorOutput "  → GitHub may be rate limiting downloads" $Yellow
                Write-ColorOutput "  → Try again in a few minutes" $Yellow
            }
            { $_ -ge 500 } {
                Write-ColorOutput "  → GitHub server error (HTTP $statusCode)" $Yellow
                Write-ColorOutput "  → Try again in a few minutes" $Yellow
            }
            default {
                if ($errorMessage -match "timeout|timed out") {
                    Write-ColorOutput "  → Download timeout - the binary may be large" $Yellow
                    Write-ColorOutput "  → Check your internet connection and try again" $Yellow
                } elseif ($errorMessage -match "resolve|dns|name") {
                    Write-ColorOutput "  → DNS resolution failed" $Yellow
                    Write-ColorOutput "  → Check your internet connection" $Yellow
                } elseif ($errorMessage -match "ssl|tls|certificate") {
                    Write-ColorOutput "  → SSL/TLS certificate error" $Yellow
                    Write-ColorOutput "  → Update your system certificates" $Yellow
                } else {
                    Write-ColorOutput "  → $errorMessage" $Yellow
                }
            }
        }
        
        Write-ColorOutput "" $White
        Write-ColorOutput "Alternative solutions:" $Cyan
        Write-ColorOutput "  1. Manual download: https://github.com/kunalkushwaha/agenticgokit/releases" $White
        Write-ColorOutput "  2. Try a different version: -Version v0.3.0" $White
        Write-ColorOutput "  3. Use Go install: go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest" $White
        
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
