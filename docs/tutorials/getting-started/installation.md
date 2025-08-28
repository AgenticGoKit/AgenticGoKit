---
title: "Installation"
description: "Set up AgenticGoKit and verify your development environment"
prev:
  text: "Getting Started"
  link: "./README"
next:
  text: "Understanding Agents"
  link: "./understanding-agents"
---

# Installation

Let's get AgenticGoKit installed and running on your system. This guide covers everything you need to start building AI agent systems.

## Learning Objectives

By the end of this section, you'll have:
- AgenticGoKit CLI installed and working
- Go development environment verified
- LLM provider configured (OpenAI, Azure OpenAI, or local Ollama)
- Created and validated your first project

## System Requirements

Before we begin, ensure your system meets these requirements:

- **Operating System**: Windows 10+, macOS 10.15+, or Linux (Ubuntu 18.04+, CentOS 7+)
- **Go**: Version 1.21 or later
- **Memory**: At least 4GB RAM (8GB recommended for local LLM usage)
- **Disk Space**: 2GB free space for tools and models
- **Network**: Internet connection for downloading dependencies and accessing cloud LLM providers

## Step 1: Verify Go Installation

First, let's make sure Go is properly installed:

```bash
go version
```

You should see output like:
```
go version go1.21.0 linux/amd64
```

If Go isn't installed or is an older version:

### Installing Go

**Windows:**
1. Download Go from [golang.org/dl](https://golang.org/dl)
2. Run the installer and follow the prompts
3. Restart your command prompt/PowerShell

**macOS:**
```bash
# Using Homebrew (recommended)
brew install go

# Or download from golang.org/dl
```

**Linux:**
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# CentOS/RHEL/Fedora
sudo dnf install golang
# or: sudo yum install golang
```

## Step 2: Install AgenticGoKit CLI

The AgenticGoKit CLI (`agentcli`) is your main tool for creating and managing agent projects.

### Recommended Installation

```bash
go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest
```

### Verify Installation

```bash
agentcli version
```

You should see version information like:
```
agentcli version v0.3.0
```

### Test CLI Commands

Let's verify the main commands are available:

```bash
# Show help
agentcli --help

# List available templates
agentcli config template --list

# Show MCP commands
agentcli mcp --help
```

::: tip Success Indicator
If all commands show help text without errors, your CLI installation is working correctly!
:::

## Step 3: Choose Your LLM Provider

AgenticGoKit supports multiple LLM providers. Choose the option that works best for you:

### Option A: OpenAI (Easiest to start)

**Pros**: Reliable, high-quality responses, easy setup
**Cons**: Requires API key and costs money per request

1. Get an API key from [OpenAI](https://platform.openai.com/api-keys)
2. Set your environment variable:

**Windows (PowerShell):**
```powershell
$env:OPENAI_API_KEY = "your-api-key-here"
```

**macOS/Linux (Bash):**
```bash
export OPENAI_API_KEY="your-api-key-here"
```

### Option B: Azure OpenAI (Enterprise choice)

**Pros**: Enterprise features, data privacy, reliable
**Cons**: Requires Azure subscription and setup

1. Set up Azure OpenAI service in Azure portal
2. Get your endpoint, API key, and deployment name
3. Set environment variables:

**Windows (PowerShell):**
```powershell
$env:AZURE_OPENAI_API_KEY = "your-api-key"
$env:AZURE_OPENAI_ENDPOINT = "https://your-resource.openai.azure.com/"
$env:AZURE_OPENAI_DEPLOYMENT = "your-deployment-name"
```

**macOS/Linux (Bash):**
```bash
export AZURE_OPENAI_API_KEY="your-api-key"
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com/"
export AZURE_OPENAI_DEPLOYMENT="your-deployment-name"
```

### Option C: Ollama (Local, free)

**Pros**: Free, private, works offline
**Cons**: Requires more setup, uses local resources

1. **Install Ollama:**

**Windows:**
- Download from [ollama.ai](https://ollama.ai)
- Run the installer

**macOS:**
```bash
# Using Homebrew
brew install ollama

# Or download from ollama.ai
```

**Linux:**
```bash
curl -fsSL https://ollama.ai/install.sh | sh
```

2. **Start Ollama service:**

```bash
# Start Ollama (runs in background)
ollama serve
```

3. **Download a model:**

```bash
# Download a lightweight model (recommended for getting started)
ollama pull gemma2:2b

# Or a more capable model (requires more memory)
ollama pull llama3.1:8b
```

4. **Set environment variable:**

```bash
export OLLAMA_HOST="http://localhost:11434"
```

## Step 4: Create Your First Project

Let's verify everything works by creating a test project:

```bash
# Create a basic agent project
agentcli create test-project --template basic

# Navigate to the project
cd test-project
```

### Examine the Project Structure

Look at what was created:

```bash
# List the files
ls -la

# View the main configuration
cat agentflow.toml
```

You should see:
- `main.go` - The main application entry point
- `agentflow.toml` - Configuration file for your agents
- `go.mod` - Go module file
- `agents/` - Directory containing agent implementations

### Validate the Configuration

```bash
agentcli validate
```

You should see:
```
Status: VALID
Configuration is correct and ready to use.
```

## Step 5: Test Your Setup

Let's run your first agent to make sure everything works:

```bash
# Run the agent with a simple message
go run . -m "Hello, can you introduce yourself?"
```

**Expected Output:**
You should see the agent respond with an introduction. The exact response will vary based on your LLM provider and model.

::: tip Success!
If you see a response from your agent, congratulations! Your AgenticGoKit installation is working correctly.
:::

## Troubleshooting Common Issues

### "agentcli: command not found"

**Problem**: The CLI isn't in your system PATH.

**Solution**:
1. Make sure `$GOPATH/bin` is in your PATH
2. Check where Go installs binaries: `go env GOPATH`
3. Add `$GOPATH/bin` to your PATH in your shell profile

### "provider not registered" error

**Problem**: Missing plugin imports in your Go code.

**Solution**: The generated projects include necessary imports, but if you see this error, ensure your `main.go` includes:

```go
import (
    _ "github.com/kunalkushwaha/agenticgokit/plugins/llm/openai"
    _ "github.com/kunalkushwaha/agenticgokit/plugins/llm/ollama"
    // ... other plugins
)
```

### LLM Connection Issues

**OpenAI/Azure OpenAI**:
- Verify your API key is correct
- Check your internet connection
- Ensure you have sufficient credits/quota

**Ollama**:
- Make sure Ollama service is running: `ollama serve`
- Verify the model is downloaded: `ollama list`
- Check the host URL is correct

### Go Module Issues

If you see Go module errors:

```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download
```

## What You've Learned

✅ **Installed AgenticGoKit CLI** and verified it's working  
✅ **Configured an LLM provider** for your agents to use  
✅ **Created your first project** using the CLI  
✅ **Validated your setup** by running a test agent  
✅ **Learned troubleshooting** techniques for common issues  

## Next Steps

Now that AgenticGoKit is installed and working, let's understand what agents are and how they work in the AgenticGoKit framework.

**[→ Continue to Understanding Agents](./understanding-agents.md)**

---

::: details Quick Navigation

**Previous:** [Getting Started](./README.md) - Tutorial overview and learning path  
**Next:** [Understanding Agents](./understanding-agents.md) - Core concepts and mental models  
**Jump to:** [Your First Agent](./first-agent.md) - Skip concepts and start building  

:::

::: details Need More Help?

**Still having issues?** Here are additional resources:

- **[Troubleshooting Guide](./troubleshooting.md)** - Comprehensive problem-solving guide
- **[GitHub Discussions](https://github.com/kunalkushwaha/agenticgokit/discussions)** - Ask the community
- **[GitHub Issues](https://github.com/kunalkushwaha/agenticgokit/issues)** - Report bugs

**Want to contribute?** Check out our [Contributor Guide](../../contributors/ContributorGuide.md)

**Related Documentation:**
- [Core Concepts](../core-concepts/README.md) - Deep dive into AgenticGoKit architecture
- [API Reference](../../reference/README.md) - Complete API documentation
- [CLI Reference](../../reference/cli.md) - Command-line interface guide

:::