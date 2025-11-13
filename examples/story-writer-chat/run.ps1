# Story Writer Chat App - Quick Start Script (PowerShell)

Write-Host "📖 Story Writer Chat App - Quick Start" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

# Check if Ollama is running
Write-Host "🔍 Checking Ollama connection..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:11434/api/tags" -Method GET -TimeoutSec 5 -UseBasicParsing -ErrorAction Stop
    Write-Host "✅ Ollama is running" -ForegroundColor Green
} catch {
    Write-Host "❌ Ollama is not running!" -ForegroundColor Red
    Write-Host "Please start Ollama first:" -ForegroundColor Yellow
    Write-Host "  ollama serve" -ForegroundColor White
    exit 1
}

# Check if gemma2:2b model is available
Write-Host "🔍 Checking for gemma2:2b model..." -ForegroundColor Yellow
$models = Invoke-WebRequest -Uri "http://localhost:11434/api/tags" -Method GET -UseBasicParsing | ConvertFrom-Json

$hasModel = $false
foreach ($model in $models.models) {
    if ($model.name -like "gemma2:2b*") {
        $hasModel = $true
        break
    }
}

if (-not $hasModel) {
    Write-Host "⚠️  gemma2:2b model not found" -ForegroundColor Yellow
    $response = Read-Host "Would you like to pull it now? (y/n)"
    if ($response -eq "y" -or $response -eq "Y") {
        Write-Host "📥 Pulling gemma2:2b model..." -ForegroundColor Yellow
        ollama pull gemma2:2b
    } else {
        Write-Host "Please pull the model manually:" -ForegroundColor Yellow
        Write-Host "  ollama pull gemma2:2b" -ForegroundColor White
        exit 1
    }
}

Write-Host "✅ Model available" -ForegroundColor Green
Write-Host ""

# Install dependencies
Write-Host "📦 Installing dependencies..." -ForegroundColor Yellow
go mod tidy

# Run the application
Write-Host ""
Write-Host "🚀 Starting Story Writer Chat App..." -ForegroundColor Cyan
Write-Host "Open your browser at: http://localhost:8080" -ForegroundColor Green
Write-Host ""
Write-Host "Press Ctrl+C to stop the server" -ForegroundColor Yellow
Write-Host ""

go run main.go
