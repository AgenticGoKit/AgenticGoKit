# AgenticGoKit v1beta Examples

> ✅ **STABLE BETA**: The v1beta API is stable and fully functional with real LLM integrations. While still in beta, the core APIs are working and ready for testing and feedback. These examples demonstrate real working agents, workflows, and integrations.

This directory contains examples demonstrating the **v1beta APIs** for AgenticGoKit. All examples use real LLM providers and showcase stable beta patterns from the `v1beta` package.

## 📚 Available Examples

### Getting Started

#### 1. [Ollama Quickstart](./ollama-quickstart/)
**Best for: Beginners & Fast Prototyping**

The simplest way to create an agent using the QuickStart API. Perfect starting point for learning v1beta.

- ✅ QuickStart API with `QuickChatAgentWithConfig()`
- ✅ Minimal setup (< 100 lines)
- ✅ Real LLM integration with Ollama
- ✅ Interactive Q&A loop

```bash
cd ollama-quickstart
go run main.go
```

#### 2. [Ollama Short Answer](./ollama-short-answer/)
**Best for: Learning Builder Pattern**

Complete example using the Builder pattern to create a customized agent.

- ✅ Builder Pattern with `NewBuilder()`
- ✅ Custom configuration
- ✅ ChatAgent preset
- ✅ Short, concise answers
- ✅ Multiple query examples

```bash
cd ollama-short-answer
go run main.go
```

#### 3. [Ollama Config-Based](./ollama-config-based/)
**Best for: Production Deployments**

Demonstrates TOML-based configuration for agents, separating code from configuration.

- ✅ TOML configuration loading
- ✅ Environment variable support
- ✅ Multiple environment configs
- ✅ Easy configuration management

```bash
cd ollama-config-based
go run main.go
# Or with custom config:
go run main.go my-config.toml
```

### Streaming & Real-time

#### 4. [Streaming Demo](./streaming-demo/)
**Best for: Understanding Real-time Streaming**

Comprehensive streaming examples across different LLM providers with performance comparisons.

- ✅ Real-time token streaming
- ✅ Multiple demo modes (basic, advanced, multi-provider, interactive)
- ✅ Performance metrics
- ✅ Provider comparison (Ollama, OpenAI, Azure)

```bash
cd streaming-demo
go run main.go
```

#### 5. [Simple Streaming](./simple-streaming/)
**Best for: Quick Streaming Example**

Minimal streaming example showing token-by-token output.

- ✅ Straightforward implementation
- ✅ Clean streaming pattern
- ✅ Easy to understand

```bash
cd simple-streaming
go run main.go
```

### Multi-Agent Workflows

#### 6. [Sequential Workflow Demo](./sequential-workflow-demo/)
**Best for: Learning Sequential Workflows**

Demonstrates multi-agent sequential workflows where agents execute one after another.

- ✅ Sequential workflow orchestration
- ✅ Multi-agent coordination
- ✅ OpenRouter integration
- ✅ Agent output chaining

```bash
cd sequential-workflow-demo
go run main.go
```

#### 7. [Streaming Workflow](./streaming_workflow/)
**Best for: Real-time Workflow Streaming**

Shows how to stream output from multi-agent workflows in real-time.

- ✅ Workflow streaming
- ✅ Real-time agent outputs
- ✅ Multi-step visualization

```bash
cd streaming_workflow
go run main.go
```

#### 8. [Story Writer Chat v2](./story-writer-chat-v2/)
**Best for: Complete Production Application**

Full-featured multi-agent story writing application with WebSocket streaming.

- ✅ Multi-agent workflow
- ✅ WebSocket server
- ✅ Real-time collaboration
- ✅ HuggingFace integration
- ✅ Production-ready architecture

```bash
cd story-writer-chat-v2
go run main.go
# Open browser to http://localhost:8080
```

#### 9. [Researcher Reporter](./researcher-reporter/)
**Best for: Research Workflows**

Multi-agent workflow for research and report generation.

- ✅ Research agent
- ✅ Analysis agent
- ✅ Report generation
- ✅ Workflow composition

```bash
cd researcher-reporter
go run main.go
```

### Memory & Context

#### 10. [Conversation Memory Demo](./conversation-memory-demo/)
**Best for: Understanding Memory Systems**

Demonstrates conversation memory and context management.

- ✅ Conversation history
- ✅ Memory persistence
- ✅ Context retrieval

```bash
cd conversation-memory-demo
go run main.go
```

#### 11. [Conversation Memory Stream Demo](./conversation-memory-stream-demo/)
**Best for: Streaming with Memory**

Combines streaming with conversation memory.

- ✅ Streaming + memory
- ✅ Real-time responses
- ✅ Context-aware conversations

```bash
cd conversation-memory-stream-demo
go run main.go
```

#### 12. [Memory and Tools](./memory-and-tools/)
**Best for: Advanced Memory + Tools**

Shows integration of memory systems with tool usage.

- ✅ Memory-enhanced agents
- ✅ Tool integration
- ✅ RAG patterns

```bash
cd memory-and-tools
go run main.go
```

### Multimodal & Advanced

#### 13. [Multimodal Demo](./multimodal-demo/)
**Best for: Images, Audio, Video Processing**

Demonstrates multimodal capabilities with images, audio, and video inputs.

- ✅ Image analysis
- ✅ Audio processing
- ✅ Video understanding
- ✅ Base64 and URL inputs
- ✅ OpenAI Vision / Ollama multimodal models

```bash
cd multimodal-demo
go run main.go
```

### Tools & Integrations

#### 14. [MCP Integration](./mcp-integration/)
**Best for: Model Context Protocol**

Shows how to integrate MCP (Model Context Protocol) servers and tools.

- ✅ MCP server connection
- ✅ Tool discovery
- ✅ MCP tool execution
- ✅ Plugin architecture

```bash
cd mcp-integration
go run main.go
```

#### 15. [MCP Tools Blog Demo](./mcp-tools-blog-demo/)
**Best for: MCP Blog-specific Tools**

Specific example of MCP tools for blog operations.

- ✅ Blog-specific MCP tools
- ✅ Real-world MCP usage
- ✅ Tool composition

```bash
cd mcp-tools-blog-demo
go run main.go
```

#### 16. [Marketplace Order Agent](./marketplace-order-agent/)
**Best for: E-commerce Integration**

Demonstrates an agent handling marketplace orders.

- ✅ Order processing
- ✅ Business logic integration
- ✅ Real-world use case

```bash
cd marketplace-order-agent
go run main.go
```

### Provider Examples

#### 17. [HuggingFace Quickstart](./huggingface-quickstart/)
**Best for: HuggingFace Models**

Quick start with HuggingFace model integration.

- ✅ HuggingFace API
- ✅ Open-source models
- ✅ Custom model deployment

```bash
cd huggingface-quickstart
go run main.go
```

#### 18. [OpenRouter Quickstart](./openrouter-quickstart/)
**Best for: OpenRouter API**

Getting started with OpenRouter for multi-model access.

- ✅ OpenRouter integration
- ✅ Multiple model access
- ✅ Unified API

```bash
cd openrouter-quickstart
go run main.go
```

#### 19. [Story Writer Chat (v1)](./story-writer-chat/)
**Best for: Legacy Story Writer**

Earlier version of story writer for comparison.

```bash
cd story-writer-chat
go run main.go
```

## 🎯 Choosing the Right Example

| Use Case | Example | Complexity |
|----------|---------|------------|
| **Just starting** | Ollama Quickstart | ⭐ Simple |
| **Learning builders** | Ollama Short Answer | ⭐⭐ Moderate |
| **Real-time streaming** | Streaming Demo | ⭐⭐ Moderate |
| **Multi-agent workflows** | Sequential Workflow | ⭐⭐⭐ Advanced |
| **Production app** | Story Writer Chat v2 | ⭐⭐⭐⭐ Complex |
| **Images/Audio/Video** | Multimodal Demo | ⭐⭐ Moderate |
| **Memory/Context** | Conversation Memory | ⭐⭐ Moderate |
| **MCP Tools** | MCP Integration | ⭐⭐⭐ Advanced |
| **HuggingFace** | HuggingFace Quickstart | ⭐ Simple |
| **OpenRouter** | OpenRouter Quickstart | ⭐ Simple |

## 📖 Key v1beta APIs Demonstrated

### Agent Creation

```go
// Method 1: Builder Pattern (Recommended)
agent, err := v1beta.NewBuilder("my-agent").
    WithConfig(config).
    WithPreset(v1beta.ChatAgent).
    Build()

// Method 2: QuickStart API (Fast prototyping)
agent, err := v1beta.QuickChatAgentWithConfig(model, config)

// Method 3: Config File (Production)
config, err := v1beta.LoadConfigFromTOML("config.toml")
agent, err := v1beta.NewBuilder(config.Name).WithConfig(config).Build()
```

### Agent Execution

```go
// Basic execution
result, err := agent.Run(ctx, "Hello!")

// With options
opts := v1beta.NewRunOptions().SetTimeout(60 * time.Second)
result, err := agent.RunWithOptions(ctx, input, opts)

// Streaming (real-time token delivery)
stream, err := agent.RunStream(ctx, input)
for chunk := range stream.Chunks() {
    if chunk.Type == v1beta.ChunkTypeDelta {
        fmt.Print(chunk.Delta) // Print token as it arrives
    }
}

// Advanced streaming with options
streamOpts := []v1beta.StreamOption{
    v1beta.WithBufferSize(100),
    v1beta.WithThoughts(),
    v1beta.WithToolCalls(),
}
runOpts := &v1beta.RunOptions{Timeout: 30 * time.Second}
stream, err := agent.RunStreamWithOptions(ctx, input, runOpts, streamOpts...)
```

### Multi-Agent Workflows

```go
// Sequential Workflow
workflow, err := v1beta.NewSequentialWorkflow(&v1beta.WorkflowConfig{
    Name: "research-pipeline",
})
workflow.AddStep(v1beta.WorkflowStep{Name: "research", Agent: researchAgent})
workflow.AddStep(v1beta.WorkflowStep{Name: "analyze", Agent: analyzerAgent})

// Parallel Workflow
parallel, err := v1beta.NewParallelWorkflow(&v1beta.WorkflowConfig{
    Name: "multi-analysis",
})

// DAG Workflow with dependencies
dag, err := v1beta.NewDAGWorkflow(&v1beta.WorkflowConfig{Name: "dag"})
dag.AddStep(v1beta.WorkflowStep{
    Name: "analyze",
    Agent: analyzer,
    Dependencies: []string{"fetch"},
})

// Loop Workflow
loop, err := v1beta.NewLoopWorkflow(&v1beta.WorkflowConfig{
    MaxIterations: 5,
})
loop.SetLoopCondition(v1beta.Conditions.OutputContains("DONE"))
```

### Multimodal Input

```go
// Images, Audio, Video
opts := v1beta.NewRunOptions()
opts.Images = []v1beta.ImageData{
    {URL: "https://example.com/image.jpg"},
    {Base64: base64ImageData},
}
opts.Audio = []v1beta.AudioData{
    {URL: "https://example.com/audio.mp3", Format: "mp3"},
}
opts.Video = []v1beta.VideoData{
    {Base64: base64VideoData, Format: "mp4"},
}
result, err := agent.RunWithOptions(ctx, "Describe this content", opts)
```

### Configuration Types

```go
// Agent Config
config := &v1beta.Config{
    Name:         "my-agent",
    SystemPrompt: "You are helpful",
    Timeout:      30 * time.Second,
    LLM:          v1beta.LLMConfig{...},
}

// LLM Config
llmConfig := v1beta.LLMConfig{
    Provider:    "ollama",
    Model:       "llama3.2",
    Temperature: 0.7,
    MaxTokens:   1000,
    BaseURL:     "http://localhost:11434",
}
```

### Presets

```go
// Available presets
v1beta.ChatAgent        // Conversational agent
v1beta.ResearchAgent    // Research with tools and memory
v1beta.DataAgent        // Data processing
v1beta.WorkflowAgent    // Multi-agent orchestration
```

## 🚀 Prerequisites

### 1. For Ollama Examples

Install Ollama for local LLM execution:

```bash
# macOS/Linux
curl -fsSL https://ollama.com/install.sh | sh

# Windows - Download from https://ollama.com/download
```

Pull required models:

```bash
ollama pull llama3.2
ollama pull gemma3:1b
```

Verify Ollama is running:

```bash
curl http://localhost:11434/api/tags
```

### 2. For OpenAI Examples

Set your API key:

```bash
# Linux/macOS
export OPENAI_API_KEY="your-key-here"

# Windows PowerShell
$env:OPENAI_API_KEY="your-key-here"
```

### 3. For HuggingFace Examples

Set your API key:

```bash
# Linux/macOS
export HUGGINGFACE_API_KEY="your-key-here"

# Windows PowerShell
$env:HUGGINGFACE_API_KEY="your-key-here"
```

### 4. For OpenRouter Examples

Set your API key:

```bash
# Linux/macOS
export OPENROUTER_API_KEY="your-key-here"

# Windows PowerShell
$env:OPENROUTER_API_KEY="your-key-here"
```

## 🔧 Running the Examples

### Run Directly

```bash
cd examples/<example-name>
go run main.go
```

### Build and Run

```bash
cd examples/<example-name>
go build -o agent
./agent  # Linux/macOS
.\agent.exe  # Windows
```

### Run with Go Modules

All examples have their own `go.mod` files and can be run independently:

```bash
cd examples/<example-name>
go mod tidy  # Download dependencies
go run main.go
```

## 📝 Example Code Structure

Each example follows this structure:

```
example-name/
├── main.go           # Main application code
├── go.mod            # Go module definition
├── config.toml       # Configuration (if applicable)
├── README.md         # Example-specific documentation
└── assets/           # Media files (for multimodal examples)
```

## 🎓 Learning Path

1. **Start Here**: [Ollama Quickstart](./ollama-quickstart/) - Simplest introduction
2. **Builder Pattern**: [Ollama Short Answer](./ollama-short-answer/) - Learn flexible agent construction
3. **Streaming**: [Streaming Demo](./streaming-demo/) - Real-time responses
4. **Workflows**: [Sequential Workflow](./sequential-workflow-demo/) - Multi-agent systems
5. **Production**: [Story Writer Chat v2](./story-writer-chat-v2/) - Complete application
6. **Advanced**: [Multimodal Demo](./multimodal-demo/) - Images, audio, video

## 🌟 Key Features Demonstrated

- ✅ **Real-time Streaming**: Token-by-token responses
- ✅ **Builder Pattern**: Flexible agent construction  
- ✅ **QuickStart API**: Rapid development
- ✅ **TOML Configuration**: Declarative setup
- ✅ **Multi-Agent Workflows**: Sequential, parallel, DAG, loop orchestration
- ✅ **Multimodal**: Images, audio, video processing
- ✅ **Memory Systems**: Conversation history and RAG
- ✅ **MCP Integration**: Model Context Protocol tools
- ✅ **Multiple Providers**: Ollama, OpenAI, Azure, HuggingFace, OpenRouter
- ✅ **Error Handling**: Robust error management
- ✅ **Context Management**: Timeout and cancellation
- ✅ **Production Ready**: Real-world patterns

## 🔗 Related Documentation

- [v1beta API Documentation](../v1beta/README.md)
- [Main README](../README.md)
- [Migration Guide](../docs/MIGRATION.md)
- [API Versioning](../docs/API_VERSIONING.md)

## 💡 Next Steps

After completing these examples, explore:

- **Advanced Streaming**: WebSocket-based real-time streaming
- **Memory & RAG**: Retrieval-augmented generation patterns
- **Tool Integration**: Custom tools and MCP servers
- **Complex Workflows**: DAG and loop workflows
- **Production Deployment**: Containerization and scaling
- **Custom Providers**: Implement your own LLM provider

## ❓ Troubleshooting

### Ollama Issues

**Ollama Not Running**
```
Error: failed to connect to Ollama
Solution: Start Ollama service or check http://localhost:11434
```

**Model Not Found**
```
Error: model 'llama3.2' not found
Solution: Run 'ollama pull llama3.2'
```

**Port Conflict**
```
Error: address already in use
Solution: Stop other Ollama instances or change port in config
```

### API Key Issues

**Missing API Key**
```
Error: OPENAI_API_KEY not set
Solution: Set environment variable with your API key
```

**Invalid API Key**
```
Error: authentication failed
Solution: Verify your API key is correct and has proper permissions
```

### Build Issues

**Import Errors**
```
Error: cannot find package
Solution: Run 'go mod tidy' in the example directory
```

**Module Issues**
```
Error: module not found
Solution: Ensure you're in the example directory and run 'go mod download'
```

### Runtime Issues

**Context Timeout**
```
Error: context deadline exceeded
Solution: Increase timeout in config or RunOptions
```

**Memory Issues**
```
Error: out of memory
Solution: Reduce MaxTokens or use smaller models
```

## 🤝 Contributing

Want to add more examples? Please:

1. Follow the existing structure
2. Use v1beta public APIs only
3. Include comprehensive README with setup instructions
4. Add proper error handling
5. Test with at least one LLM provider
6. Include go.mod and go.sum files
7. Document prerequisites and environment variables

## 📄 License

These examples are part of AgenticGoKit and follow the Apache 2.0 license.
