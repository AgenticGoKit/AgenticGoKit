# Ollama Short Answer Agent - vNext API Example

> ⚠️ **IMPORTANT**: The vNext builder currently returns **mock responses** instead of calling actual LLMs. This example demonstrates the API design and usage patterns. For working LLM integration, use `core.SimpleAgent` (see [examples/01-simple-agent](../../01-simple-agent/)). See [IMPLEMENTATION_STATUS.md](../IMPLEMENTATION_STATUS.md) for details.

This example demonstrates how to create a simple, single-agent application using the AgenticGoKit vNext public APIs. The agent uses Ollama as the LLM provider and is configured to provide short, concise answers to user queries.

## Features

- ✅ Uses **vNext public APIs** (Builder pattern)
- ✅ **Ollama integration** for local LLM execution (when implemented)
- ✅ **Short answer optimization** with system prompts and token limits
- ✅ **Clean error handling** and timeout management
- ✅ **Simple, readable code** for learning purposes

## Current Status

📝 **API Design Complete** - The code demonstrates correct API usage  
⏳ **Implementation Pending** - Actual LLM calls are not yet implemented  
✅ **Compiles Successfully** - All code is syntactically correct

## Prerequisites

1. **Go 1.23+** installed
2. **Ollama** installed and running
3. **Llama 3.2 model** pulled in Ollama

### Install Ollama

```bash
# macOS/Linux
curl -fsSL https://ollama.com/install.sh | sh

# Windows
# Download from https://ollama.com/download
```

### Pull the Model

```bash
ollama pull llama3.2
```

Verify Ollama is running:
```bash
curl http://localhost:11434/api/tags
```

## Project Structure

```
ollama-short-answer/
├── main.go           # Main application with agent setup and execution
├── go.mod            # Go module file
└── README.md         # This file
```

## Code Walkthrough

### 1. Agent Configuration

The agent is configured with:
- **System Prompt**: Instructs the LLM to provide short, 2-3 sentence answers
- **Low Temperature** (0.3): More focused and deterministic responses
- **Limited Tokens** (200): Enforces brevity
- **Ollama Provider**: Uses local Llama 3.2 model

```go
config := &vnext.Config{
    Name:         "short-answer-agent",
    SystemPrompt: systemPrompt,
    Timeout:      30 * time.Second,
    LLM: vnext.LLMConfig{
        Provider:    "ollama",
        Model:       "llama3.2",
        Temperature: 0.3,
        MaxTokens:   200,
    },
}
```

### 2. Builder Pattern

The agent is built using the vNext Builder pattern with the `ChatAgent` preset:

```go
agent, err := vnext.NewBuilder(config.Name).
    WithConfig(config).
    WithPreset(vnext.ChatAgent).
    Build()
```

### 3. Agent Execution

The agent is initialized, runs queries, and cleaned up properly:

```go
ctx := context.Background()
agent.Initialize(ctx)
defer agent.Cleanup(ctx)

result, err := agent.Run(ctx, query)
```

## Running the Example

### Option 1: Direct Execution

```bash
cd examples/vnext/ollama-short-answer
go run main.go
```

### Option 2: Build and Run

```bash
cd examples/vnext/ollama-short-answer
go build -o ollama-agent
./ollama-agent
```

## Expected Output

```
===========================================
  Ollama Short Answer Agent - vNext API
===========================================

[Query 1] What is Go programming language?
---
✓ Answer: Go is a statically typed, compiled programming language developed by Google. It's designed for simplicity, efficiency, and easy concurrency with built-in support for goroutines.
   Duration: 1.2s
   Success: true

[Query 2] Explain what Docker is.
---
✓ Answer: Docker is a platform for developing, shipping, and running applications in containers. Containers package software with all dependencies, ensuring consistent execution across environments.
   Duration: 1.1s
   Success: true

...
```

## Key vNext APIs Used

### Agent Interface
- `agent.Run(ctx, input)` - Execute agent with input
- `agent.Initialize(ctx)` - Initialize agent resources
- `agent.Cleanup(ctx)` - Clean up agent resources

### Builder Pattern
- `vnext.NewBuilder(name)` - Create new agent builder
- `WithConfig(config)` - Set complete configuration
- `WithPreset(preset)` - Apply preset configuration
- `Build()` - Build the final agent

### Configuration Types
- `vnext.Config` - Main agent configuration
- `vnext.LLMConfig` - LLM provider settings
- `vnext.ChatAgent` - Chat agent preset

### Result Type
- `result.Content` - Agent response text
- `result.Duration` - Execution duration
- `result.Success` - Success status

## Customization

### Change the Model

```go
config.LLM.Model = "mistral"  // or "codellama", "gemma", etc.
```

### Adjust Response Length

```go
config.LLM.MaxTokens = 500  // Longer responses
config.LLM.Temperature = 0.7  // More creative
```

### Custom System Prompt

```go
systemPrompt := "You are an expert in [TOPIC]. Provide detailed explanations..."
```

### Add Timeout

```go
queryCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
defer cancel()
```

## Troubleshooting

### Ollama Not Running
```
Error: failed to connect to Ollama
Solution: Start Ollama service
```

### Model Not Found
```
Error: model 'llama3.2' not found
Solution: Run 'ollama pull llama3.2'
```

### Timeout Errors
```
Error: context deadline exceeded
Solution: Increase timeout in config or query context
```

## Next Steps

- **Add Streaming**: Use `agent.RunStream()` for real-time responses
- **Add Memory**: Enable conversation history with `WithMemory()`
- **Add Tools**: Integrate external tools with `WithTools()`
- **Configuration File**: Load settings from TOML file

## Related Examples

- `examples/vnext/ollama-streaming/` - Streaming responses
- `examples/vnext/ollama-with-memory/` - Conversation memory
- `examples/vnext/ollama-with-tools/` - Tool integration

## References

- [vNext Documentation](../../../docs/vnext/)
- [Builder Pattern Guide](../../../core/vnext/builder.go)
- [Configuration Guide](../../../core/vnext/config.go)
- [Ollama Documentation](https://github.com/ollama/ollama)
