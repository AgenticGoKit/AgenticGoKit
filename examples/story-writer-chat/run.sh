#!/bin/bash

# Story Writer Chat App - Quick Start Script

echo "📖 Story Writer Chat App - Quick Start"
echo "======================================"
echo ""

# Check if Ollama is running
echo "🔍 Checking Ollama connection..."
if ! curl -s http://localhost:11434/api/tags > /dev/null; then
    echo "❌ Ollama is not running!"
    echo "Please start Ollama first:"
    echo "  ollama serve"
    exit 1
fi

echo "✅ Ollama is running"

# Check if gemma2:2b model is available
echo "🔍 Checking for gemma2:2b model..."
if ! curl -s http://localhost:11434/api/tags | grep -q "gemma2:2b"; then
    echo "⚠️  gemma2:2b model not found"
    echo "Would you like to pull it now? (y/n)"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        echo "📥 Pulling gemma2:2b model..."
        ollama pull gemma2:2b
    else
        echo "Please pull the model manually:"
        echo "  ollama pull gemma2:2b"
        exit 1
    fi
fi

echo "✅ Model available"
echo ""

# Install dependencies
echo "📦 Installing dependencies..."
go mod tidy

# Run the application
echo ""
echo "🚀 Starting Story Writer Chat App..."
echo "Open your browser at: http://localhost:8080"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

go run main.go
