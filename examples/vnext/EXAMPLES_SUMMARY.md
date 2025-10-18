# vNext Examples Summary

## Overview

I've created three comprehensive examples demonstrating the AgenticGoKit vNext public APIs. All examples use Ollama as the LLM provider and showcase different approaches to creating a single agent that helps with user queries in short answers.

## Created Examples

### 1. **ollama-short-answer/** (Main Example)
**Approach**: Builder Pattern with full configuration

**Key Features**:
- Uses `vnext.NewBuilder()` with the Builder pattern
- Complete `vnext.Config` configuration
- `ChatAgent` preset for conversational setup
- System prompt optimized for short answers (2-3 sentences)
- Low temperature (0.3) for focused responses
- Token limit (200) to enforce brevity
- Multiple example queries with error handling
- Full lifecycle: Initialize → Run → Cleanup

**Code Highlights**:
```go
config := &vnext.Config{
    Name:         "short-answer-agent",
    SystemPrompt: "You are a helpful assistant that provides short, concise answers...",
    Timeout:      30 * time.Second,
    LLM: vnext.LLMConfig{
        Provider:    "ollama",
        Model:       "llama3.2",
        Temperature: 0.3,
        MaxTokens:   200,
    },
}

agent, err := vnext.NewBuilder(config.Name).
    WithConfig(config).
    WithPreset(vnext.ChatAgent).
    Build()
```

### 2. **ollama-quickstart/** (Simplified)
**Approach**: QuickStart API for rapid development

**Key Features**:
- Uses `vnext.QuickChatAgentWithConfig()` 
- Minimal code (~50 lines)
- Same functionality as main example
- Perfect for prototyping
- `vnext.InitializeDefaults()` for framework setup

**Code Highlights**:
```go
vnext.InitializeDefaults()

agent, err := vnext.QuickChatAgentWithConfig("llama3.2", config)
```

### 3. **ollama-config-based/** (Production-Ready)
**Approach**: TOML configuration file

**Key Features**:
- Configuration loaded from `config.toml`
- Separation of code and configuration
- Environment variable support (${VAR})
- Easy to modify without recompiling
- Multiple environment configs (dev/prod)
- Uses `vnext.LoadConfigFromTOML()`

**Code Highlights**:
```go
config, err := vnext.LoadConfigFromTOML("config.toml")

agent, err := vnext.NewBuilder(config.Name).
    WithConfig(config).
    WithPreset(vnext.ChatAgent).
    Build()
```

## File Structure

```
examples/vnext/
├── README.md                          # Overview and learning path
├── ollama-short-answer/               # Main example (Builder Pattern)
│   ├── main.go                        # Complete implementation
│   ├── go.mod                         # Module definition
│   ├── go.sum                         # Dependencies
│   └── README.md                      # Detailed documentation
├── ollama-quickstart/                 # QuickStart API example
│   ├── main.go                        # Minimal implementation
│   ├── go.mod                         # Module definition
│   └── README.md                      # QuickStart guide
└── ollama-config-based/               # TOML config example
    ├── main.go                        # Config-driven implementation
    ├── config.toml                    # Agent configuration
    ├── go.mod                         # Module definition
    └── README.md                      # Configuration guide
```

## vNext Public APIs Used

### Core APIs
1. **Builder Pattern**
   - `vnext.NewBuilder(name)` - Create agent builder
   - `WithConfig(config)` - Set configuration
   - `WithPreset(preset)` - Apply preset (ChatAgent, ResearchAgent, etc.)
   - `Build()` - Build final agent

2. **QuickStart Functions**
   - `vnext.InitializeDefaults()` - Initialize framework
   - `vnext.QuickChatAgentWithConfig(model, config)` - Quick agent creation

3. **Configuration Loading**
   - `vnext.LoadConfigFromTOML(path)` - Load TOML config

4. **Agent Interface**
   - `agent.Initialize(ctx)` - Initialize agent
   - `agent.Run(ctx, input)` - Execute agent
   - `agent.Cleanup(ctx)` - Cleanup resources

### Configuration Types
- `vnext.Config` - Main agent configuration
- `vnext.LLMConfig` - LLM provider settings
- `vnext.ChatAgent` - Chat preset type

### Result Type
- `result.Content` - Response text
- `result.Duration` - Execution time
- `result.Success` - Success status

## Running the Examples

### Prerequisites
```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull model
ollama pull llama3.2

# Verify Ollama is running
curl http://localhost:11434/api/tags
```

### Run Examples
```bash
# Main example
cd examples/vnext/ollama-short-answer
go run main.go

# QuickStart example
cd examples/vnext/ollama-quickstart
go run main.go

# Config-based example
cd examples/vnext/ollama-config-based
go run main.go
```

## Key Features Demonstrated

✅ **Builder Pattern**: Flexible, production-ready agent construction  
✅ **QuickStart API**: Rapid prototyping and development  
✅ **TOML Configuration**: Declarative, environment-friendly setup  
✅ **Ollama Integration**: Local LLM execution  
✅ **Short Answer Optimization**: System prompts + token limits  
✅ **Error Handling**: Comprehensive error management  
✅ **Context Management**: Timeouts and cancellation  
✅ **Lifecycle Management**: Initialize → Run → Cleanup pattern  

## Learning Path

1. **Start**: `ollama-quickstart/` - Understand basic agent creation
2. **Learn**: `ollama-short-answer/` - Master Builder pattern
3. **Deploy**: `ollama-config-based/` - Production configuration

## Example Output

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
```

## Next Steps

After completing these examples, developers can:

1. **Add Streaming**: Use `agent.RunStream()` for real-time responses
2. **Add Memory**: Use `WithMemory()` for conversation history
3. **Add Tools**: Use `WithTools()` for external tool integration
4. **Add Workflows**: Create multi-agent systems
5. **Add RAG**: Implement retrieval-augmented generation

## Documentation

All examples include:
- ✅ Comprehensive README files
- ✅ Inline code comments
- ✅ Error handling patterns
- ✅ Configuration examples
- ✅ Troubleshooting guides

## Status

All examples:
- ✅ Compile successfully
- ✅ Use only vNext public APIs
- ✅ Follow Go best practices
- ✅ Include proper error handling
- ✅ Ready for testing (requires Ollama)
