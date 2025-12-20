# Basic Agent Example

A simple chat agent that demonstrates the core v1beta builder pattern.

---

## Overview

This example shows how to:
- Create an agent with the v1beta builder
- Configure an LLM provider
- Execute a simple query
- Handle responses

---

## Complete Code

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/agenticgokit/agenticgokit/v1beta"
    _ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
)

func main() {
    config := &v1beta.Config{
        Name:         "ollama-assistant",
        SystemPrompt: "You are a helpful assistant.",
        Timeout:      120 * time.Second,
        LLM: v1beta.LLMConfig{
            Provider:    "ollama",
            Model:       "gemma3:1b",
            Temperature: 0.7,
            MaxTokens:   100,
        },
    }

    agent, err := v1beta.NewBuilder("ollama-assistant").
        WithConfig(config).
        Build()
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    // Run the agent with a simple query
    result, err := agent.Run(context.Background(), "What is AgenticGoKit?")
    if err != nil {
        log.Fatalf("Agent execution failed: %v", err)
    }

    // Print the response
    fmt.Println("Agent Response:")
    fmt.Println(result.Content)
}
```

---

## Step-by-Step Breakdown

### 1. Import Dependencies

```go
import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/agenticgokit/agenticgokit/v1beta"
    _ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
)
```

### 2. Create Agent with Builder

```go
agent, err := v1beta.NewBuilder("ChatAssistant").
    WithLLM("openai", "gpt-4").
    Build()
```

**Key Points:**
- `NewBuilder(name)` - Creates a new agent builder with a name
- `WithLLM(provider, model)` - Configures the LLM provider and model
- `Build()` - Constructs the agent instance

### 3. Execute Query

```go
result, err := agent.Run(context.Background(), "What is AgenticGoKit?")
```

**Parameters:**
- `context.Context` - For cancellation and timeouts
- `query string` - The input text for the agent

### 4. Handle Response

```go
fmt.Println(result.Content)
```

**Result Structure:**
- `Content` - The agent's response text
- `Metadata` - Additional information about the execution

---

## Running the Example

### Prerequisites

```bash
# Install v1beta
go get github.com/agenticgokit/agenticgokit/v1beta

# Set your API key
export OPENAI_API_KEY="sk-..."
```

### Execute

```bash
go run main.go
```

---

## Common Variations

### With Custom Configuration

```go
agent, err := v1beta.NewBuilder("CustomAgent").
    WithLLM("openai", "gpt-4").
    WithConfig(&v1beta.Config{
        SystemPrompt: "You are a helpful assistant specialized in Go programming.",
        Temperature:  0.7,
        MaxTokens:    1000,
    }).
    Build()
```

### With Azure OpenAI

```go
agent, err := v1beta.NewBuilder("AzureAgent").
    WithLLM("azure", "gpt-4").
    Build()
```

**Environment Variables:**
```bash
export AZURE_OPENAI_API_KEY="your-key"
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com/"
export AZURE_OPENAI_DEPLOYMENT="gpt-4"
```

### With Ollama (Local)

```go
agent, err := v1beta.NewBuilder("LocalAgent").
    WithLLM("ollama", "llama2").
    Build()
```

**Environment Variables:**
```bash
export OLLAMA_HOST="http://localhost:11434"
```

---

## Error Handling

### Basic Error Handling

```go
agent, err := v1beta.NewBuilder("Agent").
    WithLLM("openai", "gpt-4").
    Build()
if err != nil {
    log.Fatalf("Build failed: %v", err)
}

result, err := agent.Run(ctx, query)
if err != nil {
    // Check error type
    if v1beta.IsLLMError(err) {
        log.Printf("LLM error: %v", err)
    } else if v1beta.IsRetryable(err) {
        log.Printf("Retryable error: %v", err)
    } else {
        log.Fatalf("Fatal error: %v", err)
    }
}
```

### With Context Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := agent.Run(ctx, "What is Go?")
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        log.Println("Request timed out")
    } else {
        log.Printf("Error: %v", err)
    }
}
```

---

## Next Steps

- **[Streaming Agent](./streaming-agent.md)** - Add real-time streaming responses
- **[Sequential Workflow](./workflow-sequential.md)** - Chain multiple agents
- **[Memory & RAG](./memory-rag.md)** - Add memory and knowledge base
- **[Custom Handlers](./custom-handlers.md)** - Implement custom logic

---

## Related Documentation

- [Getting Started](../getting-started.md) - Complete beginner guide
- [Core Concepts](../core-concepts.md) - Understanding agents and builders
- [Configuration](../configuration.md) - All configuration options
- [Error Handling](../error-handling.md) - Error patterns and recovery
