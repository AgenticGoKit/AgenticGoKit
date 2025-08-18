# LLM Providers

**Multi-Provider LLM Integration in AgenticGoKit**

AgenticGoKit provides a unified interface for working with different LLM providers. This guide covers configuration, usage patterns, and provider-specific features.

## Provider Overview

AgenticGoKit supports multiple LLM providers through a unified `ModelProvider` interface:

- **Azure OpenAI** (Default) - Enterprise-ready with robust scaling
- **OpenAI** - Direct API access to GPT models
- **Ollama** - Local models for privacy and cost control
- **Mock** - Testing and development

## ModelProvider Interface

All providers implement the same interface:

```go
type ModelProvider interface {
    Generate(ctx context.Context, prompt string) (string, error)
    GenerateWithHistory(ctx context.Context, messages []Message) (string, error)
    Name() string
}

type Message struct {
    Role    string // "system", "user", "assistant"
    Content string
}
```

## Configuration

### Azure OpenAI (Default)

**agentflow.toml:**
```toml
[provider]
type = "azure"
api_key = "${AZURE_OPENAI_API_KEY}"
endpoint = "https://your-resource.openai.azure.com"
deployment = "gpt-4"
api_version = "2024-02-15-preview"
model = "gpt-4"
max_tokens = 2000
temperature = 0.7
```

**Environment variables:**
```bash
export AZURE_OPENAI_API_KEY="your-azure-api-key"
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com"
export AZURE_OPENAI_DEPLOYMENT="gpt-4"
```

### OpenAI

**agentflow.toml:**
```toml
[provider]
type = "openai"
api_key = "${OPENAI_API_KEY}"
model = "gpt-4"
max_tokens = 2000
temperature = 0.7
organization = "your-org-id"  # Optional
```

### Ollama (Local Models)

**agentflow.toml:**
```toml
[provider]
type = "ollama"
host = "http://localhost:11434"
model = "llama3.2:3b"
temperature = 0.7
context_window = 4096
```

**Setup Ollama:**
```bash
# Install Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# Pull a model
ollama pull llama3.2:3b

# Start Ollama server (usually automatic)
ollama serve
```

## Usage Patterns

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Load provider from configuration
    cfg, err := core.LoadConfigFromWorkingDir()
    if err != nil {
        panic(err)
    }
    provider, err := cfg.InitializeProvider()
    if err != nil {
        panic(err)
    }
    
    ctx := context.Background()
    
    // Simple generation
    response, err := provider.Generate(ctx, "Explain Go interfaces in simple terms")
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Response:", response)
    fmt.Println("Provider:", provider.Name())
}
```

### Provider Selection Strategies

```go
func createProvider() (core.ModelProvider, error) {
    providerType := os.Getenv("LLM_PROVIDER")
    
    switch providerType {
    case "azure":
    return core.NewAzureOpenAIAdapter(os.Getenv("AZURE_OPENAI_ENDPOINT"), os.Getenv("AZURE_OPENAI_API_KEY"), os.Getenv("AZURE_OPENAI_DEPLOYMENT"))
    case "openai":
    return core.NewOpenAIAdapter(os.Getenv("OPENAI_API_KEY"), "gpt-4o-mini", 8192, 0.7)
    case "ollama":
    return core.NewOllamaAdapter("http://localhost:11434", "gemma3:1b", 8192, 0.7)
    default:
    return nil, fmt.Errorf("unsupported provider: %s", providerType)
    }
}
```

## Performance Considerations

### Provider Performance Characteristics

| Provider | Latency | Throughput | Cost | Privacy |
|----------|---------|------------|------|---------|
| Azure OpenAI | Medium | High | Medium | Enterprise |
| OpenAI | Medium | High | Medium | Cloud |
| Ollama | Low | Medium | Free | Full |
| Mock | Minimal | Very High | Free | Full |

## Production Deployment

### Environment Configuration

```bash
# Production environment variables
export LLM_PROVIDER="azure"
export AZURE_OPENAI_API_KEY="prod-api-key"
export AZURE_OPENAI_ENDPOINT="https://prod-resource.openai.azure.com"
export AZURE_OPENAI_DEPLOYMENT="gpt-4"

# Fallback provider
export FALLBACK_PROVIDER="openai"
export OPENAI_API_KEY="fallback-api-key"

# Local development
export LLM_PROVIDER="ollama"
export OLLAMA_HOST="http://localhost:11434"
export OLLAMA_MODEL="gemma3:1b"
```

## Next Steps

- **[Vector Databases](vector-databases.md)** - Set up persistent storage for RAG
- **[MCP Tools](mcp-tools.md)** - Add external tool capabilities
- **[Configuration](../../reference/api/configuration.md)** - Advanced configuration options