# Ollama QuickStart Agent Example

> ⚠️ **IMPORTANT**: This example shows **API design only**. The vNext builder returns mock responses currently. For working LLM agents, use `core.SimpleAgent` (see [examples/01-simple-agent](../../01-simple-agent/)). Details in [IMPLEMENTATION_STATUS.md](../IMPLEMENTATION_STATUS.md).

This example demonstrates the **QuickStart API** from AgenticGoKit vNext for creating agents with minimal code.

## Features

- ✅ Simplest possible agent creation using `QuickChatAgentWithConfig()`
- ✅ Ollama integration for local LLM (when implemented)
- ✅ Clean, minimal code (~50 lines)
- ✅ Perfect for learning the API design

## Quick Start

```bash
# Ensure Ollama is running and llama3.2 model is available
ollama pull llama3.2

# Run the example
cd examples/vnext/ollama-quickstart
go run main.go
```

## Code Highlights

### QuickStart Agent Creation

```go
// Initialize vNext framework
vnext.InitializeDefaults()

// Create config
config := &vnext.Config{
    Name:         "quick-helper",
    SystemPrompt: "You are a helpful assistant...",
    LLM: vnext.LLMConfig{
        Provider: "ollama",
        Model:    "llama3.2",
    },
}

// Create agent with one line
agent, err := vnext.QuickChatAgentWithConfig("llama3.2", config)
```

## When to Use QuickStart API

✅ **Use QuickStart when:**
- Building simple prototypes
- Learning AgenticGoKit
- Need fast agent creation
- Single-purpose agents

❌ **Use Builder Pattern when:**
- Complex agent configurations
- Multi-agent systems
- Need fine-grained control
- Production applications

## Next Steps

- Try the [Builder Pattern Example](../ollama-short-answer/)
- Try the [TOML Config Example](../ollama-config-based/)
- Add streaming with `agent.RunStream()`
