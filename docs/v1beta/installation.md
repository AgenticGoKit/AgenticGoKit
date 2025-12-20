# Installation

This guide covers installing and setting up AgenticGoKit in your Go project.

---

## 📋 Requirements

### System Requirements
- **Go 1.21 or later** (Go 1.22+ recommended)
- **Operating System**: Linux, macOS, or Windows
- **Memory**: 512MB minimum (more for LLM operations)

### LLM Provider Requirements
You'll need at least one LLM provider:
- **OpenAI**: API key from [platform.openai.com](https://platform.openai.com)
- **Azure AI**: Azure OpenAI Service credentials
- **Ollama**: Local installation from [ollama.com](https://ollama.com)
- **HuggingFace**: API key from [huggingface.co](https://huggingface.co)
- **OpenRouter**: API key from [openrouter.ai](https://openrouter.ai)

---

## 📦 Installing AgenticGoKit

### Method 1: Using go get (Recommended)

```bash
go get github.com/agenticgokit/agenticgokit/v1beta
```

### Method 2: Adding to go.mod

Add to your `go.mod` file:

```go
require (
    github.com/agenticgokit/agenticgokit/v1beta v0.5.0
)
```

Then run:

```bash
go mod tidy
```

### Method 3: Creating a New Project

```bash
# Create project directory
mkdir myagent && cd myagent

# Initialize Go module
go mod init myagent

# Install v1beta
go get github.com/agenticgokit/agenticgokit/v1beta

# Create main.go
cat > main.go << 'EOF'
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    agent, err := v1beta.PresetChatAgentBuilder().
        WithName("HelloAgent").
        Build()
    if err != nil {
        log.Fatal(err)
    }
    
    result, err := agent.Run(context.Background(), "Hello, world!")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(result.Content)
}
EOF

# Run
go run main.go
```

---

## 🔑 Setting Up LLM Providers

### OpenAI

#### Set API Key

```bash
# Linux/macOS
export OPENAI_API_KEY="sk-..."

# Windows PowerShell
$env:OPENAI_API_KEY="sk-..."

# Windows CMD
set OPENAI_API_KEY=sk-...
```

#### Use in Code

```go
import (
    "os"
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    // Option 1: Set in environment
    os.Setenv("OPENAI_API_KEY", "sk-...")
    
    // Option 2: Pass via configuration
    agent, err := v1beta.NewBuilder("Agent").
        WithLLMProvider("openai").
        WithModel("gpt-4").
        WithConfig(v1beta.Config{
            LLMConfig: map[string]interface{}{
                "api_key": "sk-...",
            },
        }).
        Build()
}
```

### Azure AI (Azure OpenAI)

```bash
# Set Azure credentials
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com"
export AZURE_OPENAI_API_KEY="your-azure-api-key"
```

```go
agent, err := v1beta.NewBuilder("Agent").
    WithModel("azure", "gpt-4").
    WithConfig(v1beta.Config{
        LLMConfig: map[string]interface{}{
            "endpoint": "https://your-resource.openai.azure.com",
            "api_key":  "your-azure-api-key",
            "deployment_name": "gpt-4-deployment", // Your Azure deployment name
        },
    }).
    Build()
```

### Ollama (Local)

#### Install Ollama

```bash
# Linux/macOS
curl -fsSL https://ollama.com/install.sh | sh

# Or download from ollama.com
```

#### Pull a Model

```bash
ollama pull llama2
ollama pull mistral
ollama pull gemma2
```

#### Use in Code

```go
agent, err := v1beta.NewBuilder("Agent").
    WithModel("ollama", "llama2").
    WithConfig(v1beta.Config{
        LLMConfig: map[string]interface{}{
            "base_url": "http://localhost:11434", // Default Ollama URL
        },
    }).
    Build()
```

### HuggingFace

```bash
export HUGGINGFACE_API_KEY="hf_..."
```

```go
agent, err := v1beta.NewBuilder("Agent").
    WithModel("huggingface", "mistralai/Mistral-7B-Instruct-v0.2").
    Build()
```

### OpenRouter

```bash
export OPENROUTER_API_KEY="sk-or-..."
```

```go
agent, err := v1beta.NewBuilder("Agent").
    WithModel("openrouter", "anthropic/claude-3-opus").
    Build()
```

---

## 🔧 Configuration Options

### Environment Variables

AgenticGoKit v1beta supports these environment variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `OPENAI_API_KEY` | OpenAI API key | `sk-...` |
| `AZURE_OPENAI_ENDPOINT` | Azure OpenAI endpoint | `https://your-resource.openai.azure.com` |
| `AZURE_OPENAI_API_KEY` | Azure OpenAI API key | `your-azure-key` |
| `HUGGINGFACE_API_KEY` | HuggingFace API key | `hf_...` |
| `OPENROUTER_API_KEY` | OpenRouter API key | `sk-or-...` |
| `OLLAMA_HOST` | Ollama host URL | `http://localhost:11434` |
| `AGENTICGOKIT_LOG_LEVEL` | Logging level | `debug`, `info`, `warn`, `error` |

### Configuration File (Optional)

Create a `config.toml` file:

```toml
[agent]
name = "MyAgent"
provider = "openai"
model = "gpt-4"
system_prompt = "You are a helpful assistant"
temperature = 0.7
max_tokens = 2000
timeout = "30s"

[llm]
api_key = "${OPENAI_API_KEY}"  # Use environment variable
base_url = "https://api.openai.com/v1"

[logging]
level = "info"
format = "json"
```

Load configuration:

```go
import "github.com/agenticgokit/agenticgokit/v1beta/config"

cfg, err := config.LoadFromFile("config.toml")
if err != nil {
    log.Fatal(err)
}

agent, err := v1beta.NewBuilderFromConfig(cfg).Build()
```

---

## ✅ Verify Installation

Create a test file `test.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    // Set your API key
    os.Setenv("OPENAI_API_KEY", "your-key-here")
    
    // Create a simple agent
    agent, err := v1beta.PresetChatAgentBuilder().
        WithName("TestAgent").
        WithModel("openai", "gpt-3.5-turbo").
        Build()
    if err != nil {
        log.Fatal("Build failed:", err)
    }
    
    // Test the agent
    result, err := agent.Run(context.Background(), "Say 'Installation successful!'")
    if err != nil {
        log.Fatal("Run failed:", err)
    }
    
    fmt.Printf("✓ Installation verified!\n")
    fmt.Printf("Response: %s\n", result.Content)
}
```

Run the test:

```bash
go run test.go
```

Expected output:
```
✓ Installation verified!
Response: Installation successful!
```

---

## 📚 Optional Dependencies

### Memory Backends

#### PostgreSQL with pgvector

```bash
go get github.com/agenticgokit/agenticgokit/plugins/memory/postgres
```

```go
import "github.com/agenticgokit/agenticgokit/plugins/memory/postgres"

memProvider, err := postgres.New("postgresql://user:pass@localhost/db")
agent, err := v1beta.PresetChatAgentBuilder().
    WithMemory(memProvider).
    Build()
```

#### Weaviate

```bash
go get github.com/agenticgokit/agenticgokit/plugins/memory/weaviate
```

```go
import "github.com/agenticgokit/agenticgokit/plugins/memory/weaviate"

memProvider, err := weaviate.New("http://localhost:8080")
agent, err := v1beta.PresetChatAgentBuilder().
    WithMemory(memProvider).
    Build()
```

### Embedding Providers

#### OpenAI Embeddings

```bash
go get github.com/agenticgokit/agenticgokit/plugins/embedding/openai
```

#### HuggingFace Embeddings

```bash
go get github.com/agenticgokit/agenticgokit/plugins/embedding/huggingface
```

### MCP (Model Context Protocol) Support

```bash
go get github.com/agenticgokit/agenticgokit/plugins/mcp
```

---

## 🐳 Docker Setup

### Dockerfile

```dockerfile
FROM golang:1.22-alpine

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN go build -o myagent .

# Run
CMD ["./myagent"]
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  agent:
    build: .
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - AGENTICGOKIT_LOG_LEVEL=info
    volumes:
      - ./data:/app/data
```

Run:

```bash
docker-compose up
```

---

## 🔍 Troubleshooting

### Issue: "cannot find package"

**Solution**: Make sure you're using the correct import path:

```go
// ✅ Correct
import "github.com/agenticgokit/agenticgokit/v1beta"

// ❌ Wrong
import "github.com/agenticgokit/agenticgokit/core"
import "github.com/agenticgokit/agenticgokit/v1beta"
```

### Issue: "API key not found"

**Solution**: Verify environment variable is set:

```bash
echo $OPENAI_API_KEY
```

Or set it in code:

```go
os.Setenv("OPENAI_API_KEY", "sk-...")
```

### Issue: "connection refused" with Ollama

**Solution**: Ensure Ollama is running:

```bash
ollama serve
```

Check Ollama status:

```bash
curl http://localhost:11434/api/tags
```

### Issue: Go version too old

**Solution**: Update Go:

```bash
# Check current version
go version

# Update Go (macOS/Linux with brew)
brew upgrade go

# Or download from golang.org
```

### Issue: Module checksum mismatch

**Solution**: Clear module cache and retry:

```bash
go clean -modcache
go mod tidy
go get github.com/agenticgokit/agenticgokit/v1beta
```

---

## 🚀 Next Steps

Now that you've installed v1beta:

1. **[Getting Started Guide](./getting-started.md)** - Build your first agent
2. **[Core Concepts](./core-concepts.md)** - Understand the architecture
3. **[Examples](./examples/)** - Explore code examples

---

## 📖 Additional Resources

- **[v1beta README](https://github.com/agenticgokit/agenticgokit/tree/main/v1beta)** - Package overview
- **[API Reference](./api-reference.md)** - Complete API documentation
- **[Migration Guide](./migration-from-core.md)** - Upgrading from core/vnext

---

**Ready to build?** Continue to [Getting Started](./getting-started.md) →
