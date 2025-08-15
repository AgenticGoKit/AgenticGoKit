# AgenticGoKit Core Package

This package provides the **clean public API** for AgenticGoKit, containing only essential interfaces, types, and factory functions. All implementation details have been moved to internal packages to provide a minimal, easy-to-use API surface.

## 🎯 **Public API Overview**

The core package contains **only the essential components** that third-party developers need:

### Essential Interfaces
- **`Agent`** - Core agent interface for custom implementations
- **`AgentHandler`** - Event-driven agent processing interface  
- **`ModelProvider`** - LLM provider interface for different backends
- **`Memory`** - Memory storage interface for different providers
- **`Orchestrator`** - Agent orchestration interface
- **`LLMAdapter`** - Simplified LLM interaction interface

### Essential Types
- **`Config`** - Main configuration structure
- **`Prompt`**, **`Response`**, **`Token`** - LLM interaction types
- **`Event`**, **`State`**, **`Result`** - Core workflow types
- **`MemoryEntry`** - Memory storage types

### Factory Functions
- **`LoadConfig()`** - Load configuration from file
- **`NewOpenAIAdapter()`**, **`NewAzureOpenAIAdapter()`**, **`NewOllamaAdapter()`** - LLM providers
- **`NewMemory()`** - Memory provider creation
- **`NewModelProviderAdapter()`** - LLM adapter creation

## 🚀 **Quick Start Examples**

### Load Configuration
```go
import "github.com/kunalkushwaha/agenticgokit/core"

// Load from file
config, err := core.LoadConfig("agentflow.toml")

// Load defaults
config, err := core.LoadConfig("")
```

### Create LLM Provider
```go
// OpenAI
provider, err := core.NewOpenAIAdapter("api-key", "gpt-4", 2000, 0.7)

// Azure OpenAI
provider, err := core.NewAzureOpenAIAdapter(core.AzureOpenAIAdapterOptions{
    Endpoint: "https://your-resource.openai.azure.com/",
    APIKey: "your-api-key",
    ChatDeployment: "gpt-4",
})

// Ollama
provider, err := core.NewOllamaAdapter("http://localhost:11434", "llama2", 2000, 0.7)
```

### Create Memory Provider
```go
memory, err := core.NewMemory(core.MemoryConfig{
    Provider: "pgvector",
    Connection: "postgres://user:pass@localhost/db",
})
```

### Simple LLM Interaction
```go
// Direct provider usage
response, err := provider.Call(ctx, core.Prompt{
    System: "You are a helpful assistant",
    User: "Hello, world!",
    Parameters: core.ModelParameters{
        Temperature: core.FloatPtr(0.7),
        MaxTokens: core.Int32Ptr(100),
    },
})

// Simplified adapter usage
adapter := core.NewModelProviderAdapter(provider)
result, err := adapter.Complete(ctx, "You are helpful", "Hello!")
```

## 🏗️ **Architecture Principles**

### Clean API Surface
- **Minimal exports**: Only essential types and functions are public
- **Implementation hiding**: All complex logic moved to internal packages
- **Simple factories**: Easy-to-use creation functions
- **Clear interfaces**: Well-defined contracts for extensibility

### Backward Compatibility
- **No breaking changes**: Existing code continues to work
- **Same import paths**: Public APIs remain in the same locations
- **Consistent signatures**: Function signatures unchanged

### Extensibility
- **Interface-based**: Implement interfaces for custom behavior
- **Factory pattern**: Register custom implementations
- **Configuration-driven**: Extend through configuration

## 📦 **Package Structure**

```
core/                          # Public API only (essential files)
├── agent.go                   # Agent interfaces
├── config.go                  # Configuration types and loading
├── event.go                   # Event types
├── factory.go                 # Factory functions
├── llm.go                     # LLM interfaces and factories
├── memory.go                  # Memory interfaces and factories
├── orchestrator.go            # Orchestrator interface
├── result.go                  # Result types
├── runner.go                  # Runner interface
├── state.go                   # State interface
├── context.go                 # Context types
└── mcp.go                     # MCP interface

internal/                      # Implementation details (hidden)
├── config/                    # Configuration implementation
├── agents/                    # Agent management
├── memory/                    # Memory providers
├── llm/                       # LLM adapters
├── orchestrator/              # Orchestration logic
└── ...                        # Other internal packages
```

## 🔧 **Migration Guide**

If you were using internal APIs (not recommended), here's how to migrate:

### Before (Internal APIs - Don't use)
```go
// ❌ Don't import internal packages
import "github.com/kunalkushwaha/agenticgokit/internal/llm"
```

### After (Public APIs - Recommended)
```go
// ✅ Use public core package
import "github.com/kunalkushwaha/agenticgokit/core"

provider, err := core.NewOpenAIAdapter(apiKey, model, maxTokens, temperature)
```

## 📚 **Additional Resources**

- **Examples**: See `examples/` directory for complete usage examples
- **Configuration**: See `docs/guides/configuration-system.md`
- **Extension Guide**: See `docs/guides/extending-agenticgokit.md`
- **API Reference**: See `docs/reference/api.md`