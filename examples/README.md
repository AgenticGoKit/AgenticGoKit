# AgenticGoKit vNext Examples

> ⚠️ **IMPORTANT**: The vNext API is **design-complete** but the builder implementation currently returns **mock responses**. These examples demonstrate intended API usage and patterns. For working agents with actual LLM integration, use `core.SimpleAgent` API (see [examples/01-simple-agent](../01-simple-agent/)). Read [IMPLEMENTATION_STATUS.md](./IMPLEMENTATION_STATUS.md) for details.

This directory contains examples demonstrating the **vNext public APIs** for AgenticGoKit. These examples showcase the intended API design and usage patterns from `core/vnext`.

## 📚 Available Examples

### 1. [Streaming Demo](./streaming-demo/)
**Best for: Understanding Real-time Streaming**

A comprehensive example demonstrating real-time streaming capabilities across different LLM providers. Shows tokens arriving as they're generated.

- ✅ Real-time token streaming
- ✅ Multiple demo modes (basic, advanced, multi-provider, interactive)
- ✅ Performance metrics and comparison
- ✅ Streaming with options and configuration
- ✅ Provider comparison (Ollama, OpenAI, Azure)

```bash
cd streaming-demo
go run main.go
# Or for simple streaming:
go run simple_example.go
```

### 2. [Ollama QuickStart Agent](./ollama-quickstart/)
**Best for: Learning the Builder Pattern**

A complete example showing how to create a single agent using the Builder pattern with Ollama. The agent is configured to provide short, concise answers.

- ✅ Builder Pattern with `NewBuilder()`
- ✅ Custom configuration
- ✅ ChatAgent preset
- ✅ Multiple query examples
- ✅ Full error handling

```bash
cd ollama-short-answer
go run main.go
```

### 2. [Ollama QuickStart Agent](./ollama-quickstart/)
**Best for: Rapid Prototyping**

The simplest way to create an agent using the QuickStart API. Minimal code for maximum results.

- ✅ QuickStart API with `QuickChatAgentWithConfig()`
- ✅ Minimal setup (~50 lines)
- ✅ Perfect for beginners
- ✅ Fast prototyping

```bash
cd ollama-quickstart
go run main.go
```

### 3. [Ollama Short Answer Agent](./ollama-short-answer/)
**Best for: Learning the Builder Pattern**

A complete example showing how to create a single agent using the Builder pattern with Ollama. The agent is configured to provide short, concise answers.

- ✅ Builder Pattern with `NewBuilder()`
- ✅ Custom configuration
- ✅ ChatAgent preset
- ✅ Multiple query examples
- ✅ Full error handling

```bash
cd ollama-short-answer
go run main.go
```

### 4. [Ollama Config-Based Agent](./ollama-config-based/)
**Best for: Production Deployments**

Demonstrates TOML-based configuration for agents, separating code from configuration.

- ✅ TOML configuration loading
- ✅ Environment variable support
- ✅ Easy configuration management
- ✅ Multiple environment configs

```bash
cd ollama-config-based
go run main.go
# Or with custom config:
go run main.go my-config.toml
```

## 🎯 Choosing the Right Example

| Use Case | Example | Complexity |
|----------|---------|------------|
| Understanding streaming | Streaming Demo | ⭐⭐ Moderate |
| Learning basics | Ollama QuickStart | ⭐ Simple |
| Production single agent | Ollama Short Answer | ⭐⭐ Moderate |
| Config-driven apps | Ollama Config-Based | ⭐⭐ Moderate |

## 📖 Key vNext APIs Demonstrated

### Agent Creation

```go
// Method 1: Builder Pattern (Recommended)
agent, err := vnext.NewBuilder("my-agent").
    WithConfig(config).
    WithPreset(vnext.ChatAgent).
    Build()

// Method 2: QuickStart API (Fast prototyping)
agent, err := vnext.QuickChatAgentWithConfig(model, config)

// Method 3: Config File (Production)
config, err := vnext.LoadConfigFromTOML("config.toml")
agent, err := vnext.NewBuilder(config.Name).WithConfig(config).Build()
```

### Agent Execution

```go
// Basic execution
result, err := agent.Run(ctx, "Hello!")

// With options
opts := vnext.NewRunOptions().SetTimeout(60 * time.Second)
result, err := agent.RunWithOptions(ctx, input, opts)

// Streaming (real-time token delivery)
stream, err := agent.RunStream(ctx, input)
for chunk := range stream.Chunks() {
    if chunk.Type == vnext.ChunkTypeDelta {
        fmt.Print(chunk.Delta) // Print token as it arrives
    }
}

// Advanced streaming with options
streamOpts := []vnext.StreamOption{
    vnext.WithBufferSize(100),
    vnext.WithThoughts(),
    vnext.WithToolCalls(),
}
runOpts := &vnext.RunOptions{Timeout: 30 * time.Second}
stream, err := agent.RunStreamWithOptions(ctx, input, runOpts, streamOpts...)
```

### Configuration Types

```go
// Agent Config
config := &vnext.Config{
    Name:         "my-agent",
    SystemPrompt: "You are helpful",
    Timeout:      30 * time.Second,
    LLM:          vnext.LLMConfig{...},
}

// LLM Config
llmConfig := vnext.LLMConfig{
    Provider:    "ollama",
    Model:       "llama3.2",
    Temperature: 0.7,
    MaxTokens:   1000,
}
```

### Presets

```go
// Available presets
vnext.ChatAgent        // Conversational agent
vnext.ResearchAgent    // Research with tools and memory
vnext.DataAgent        // Data processing
vnext.WorkflowAgent    // Multi-agent orchestration
```

## 🚀 Prerequisites

### 1. Install Ollama

```bash
# macOS/Linux
curl -fsSL https://ollama.com/install.sh | sh

# Windows - Download from https://ollama.com/download
```

### 2. Pull Required Model

```bash
ollama pull llama3.2
```

### 3. Verify Ollama is Running

```bash
curl http://localhost:11434/api/tags
```

## 🔧 Running the Examples

### Run Directly

```bash
cd examples/vnext/<example-name>
go run main.go
```

### Build and Run

```bash
cd examples/vnext/<example-name>
go build -o agent
./agent
```

### Run All Examples

```bash
# From the examples/vnext directory
for dir in ollama-*/; do
    echo "Running $dir"
    cd "$dir"
    go run main.go
    cd ..
done
```

## 📝 Example Code Structure

Each example follows this structure:

```
example-name/
├── main.go           # Main application code
├── go.mod            # Go module definition
├── config.toml       # Configuration (if applicable)
└── README.md         # Example-specific documentation
```

## 🎓 Learning Path

1. **Start Here**: [Ollama QuickStart](./ollama-quickstart/) - Simplest introduction
2. **Next**: [Ollama Short Answer](./ollama-short-answer/) - Learn Builder pattern
3. **Advanced**: [Ollama Config-Based](./ollama-config-based/) - Production patterns

## 🌟 Key Features Demonstrated

- ✅ **Real-time Streaming**: See responses being generated token by token
- ✅ **Builder Pattern**: Flexible agent construction
- ✅ **QuickStart API**: Rapid development
- ✅ **TOML Configuration**: Declarative setup
- ✅ **Ollama Integration**: Local LLM execution
- ✅ **Error Handling**: Robust error management
- ✅ **Context Management**: Timeout and cancellation
- ✅ **Clean Code**: Production-ready patterns

## 🔗 Related Documentation

- [vNext API Documentation](../../core/vnext/README.md)
- [Builder Pattern Guide](../../core/vnext/builder.go)
- [Configuration Guide](../../core/vnext/config.go)
- [Streaming Guide](../../core/vnext/STREAMING_GUIDE.md)
- [Migration Guide](../../core/vnext/MIGRATION_GUIDE.md)

## 💡 Next Steps

After completing these examples, explore:

- **Streaming**: Add real-time response streaming
- **Memory**: Enable conversation history
- **Tools**: Integrate external tools and MCP servers
- **Workflows**: Create multi-agent systems
- **RAG**: Add retrieval-augmented generation

## ❓ Troubleshooting

### Ollama Not Running
```
Error: failed to connect to Ollama
Solution: Start Ollama service or check http://localhost:11434
```

### Model Not Found
```
Error: model 'llama3.2' not found
Solution: Run 'ollama pull llama3.2'
```

### Import Errors
```
Error: cannot find package
Solution: Run 'go mod tidy' in the example directory
```

## 🤝 Contributing

Want to add more examples? Please:

1. Follow the existing structure
2. Use vNext public APIs only
3. Include comprehensive README
4. Add error handling
5. Test with Ollama

## 📄 License

These examples are part of AgenticGoKit and follow the same license.
