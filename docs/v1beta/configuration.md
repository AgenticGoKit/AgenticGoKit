# Configuration Guide

Learn how to configure agents in AgenticGoKit v1beta using the builder pattern, configuration structs, and runtime options.

---

## 🎯 Overview

AgenticGoKit v1beta provides three configuration approaches:

1. **Builder Pattern** - Fluent API for programmatic configuration (recommended)
2. **Configuration Struct** - Direct configuration for advanced use cases
3. **TOML Files** - File-based configuration (legacy, optional)

---

## 🏗️ Builder Pattern (Recommended)

The builder pattern provides a clean, type-safe way to configure agents.

### Basic Configuration

```go
package main

import (
    "log"
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    agent, err := v1beta.NewChatAgent("MyAgent",
        v1beta.WithLLM("openai", "gpt-4"),
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

### Complete Configuration

```go
agent, err := v1beta.NewBuilder("AdvancedAgent").
    WithPreset(v1beta.ResearchAgent).
    WithLLM("openai", "gpt-4").
    WithConfig(&v1beta.Config{
        SystemPrompt: "You are a helpful research assistant",
        LLM: v1beta.LLMConfig{
            Provider:    "openai",
            Model:       "gpt-4",
            Temperature: 0.7,
            MaxTokens:   2000,
        },
        Timeout:      60 * time.Second,
    }).
    WithTools(tools).
    WithMemory(
        v1beta.WithMemoryProvider("memory"),
        v1beta.WithRAG(2000, 0.3, 0.7),
    ).
    WithHandler(customHandler).
    WithMiddleware(loggingMiddleware).
    Build()
```

---

## 🎛️ Builder Methods

### Core Methods

#### WithPreset()
Use preset configurations for common use cases:

```go
// Chat Agent - General conversation
agent, _ := v1beta.NewBuilder("ChatBot").
    WithPreset(v1beta.ChatAgent).
    Build()

// Research Agent - Analysis and investigation
agent, _ := v1beta.NewBuilder("Researcher").
    WithPreset(v1beta.ResearchAgent).
    Build()

// Data Agent - Data processing
agent, _ := v1beta.NewBuilder("DataProcessor").
    WithPreset(v1beta.DataAgent).
    Build()

// Workflow Agent - Multi-step orchestration
agent, _ := v1beta.NewBuilder("Orchestrator").
    WithPreset(v1beta.WorkflowAgent).
    Build()
```



#### WithConfig()
Set advanced configuration options:

```go
agent, _ := v1beta.NewBuilder("Agent").
    WithConfig(&v1beta.Config{
        SystemPrompt: "You are a helpful assistant",
        Timeout:      60 * time.Second,
        
        // LLM Settings
        LLM: v1beta.LLMConfig{
            Temperature: 0.7,    // Creativity (0.0 - 2.0)
            MaxTokens:   2000,   // Response length limit
        },
        
        // Streaming Settings
        Streaming: &v1beta.StreamingConfig{
            BufferSize: 100,
        },
    }).
    Build()
```

#### WithTools()
Add tool capabilities:

```go
tools := []v1beta.Tool{
    searchTool,
    calculatorTool,
    weatherTool,
}

agent, _ := v1beta.NewBuilder("ToolAgent").
    WithTools(tools).
    Build()
```

#### WithMemory()
Enable memory and RAG:

```go
agent, _ := v1beta.NewBuilder("MemoryAgent").
    WithMemory(
        v1beta.WithMemoryProvider("memory"),
        v1beta.WithRAG(2000, 0.3, 0.7), // maxTokens, personalWeight, knowledgeWeight
        v1beta.WithSessionScoped(),
        v1beta.WithContextAware(),
    ).
    Build()
```

#### WithHandler()
Set custom execution logic:

```go
agent, _ := v1beta.NewBuilder("CustomAgent").
    WithHandler(myCustomHandler).
    Build()
```

#### WithMiddleware()
Add middleware for cross-cutting concerns:

```go
agent, _ := v1beta.NewBuilder("LoggedAgent").
    WithMiddleware(loggingMiddleware).
    WithMiddleware(metricsMiddleware).
    Build()
```

---

## 📋 Configuration Struct

For advanced scenarios, use the Config struct directly:

### Config Structure

```go
type Config struct {
    // Core Settings
    Name         string
    SystemPrompt string
    Timeout      time.Duration
    DebugMode    bool
    
    // LLM Configuration
    LLM LLMConfig
    
    // Feature Configurations
    Memory    *MemoryConfig
    Tools     *ToolsConfig
    Workflow  *WorkflowConfig
    Tracing   *TracingConfig
    Streaming *StreamingConfig
}

type LLMConfig struct {
    Provider    string
    Model       string
    Temperature float32
    MaxTokens   int
}
```

### Direct Configuration

```go
config := &v1beta.Config{
    SystemPrompt: "You are a helpful assistant",
    Timeout:      60 * time.Second,
    LLM: v1beta.LLMConfig{
        Provider:    "openai",
        Model:       "gpt-4",
        Temperature: 0.7,
        MaxTokens:   2000,
    },
}

agent, err := v1beta.NewBuilder("Agent").
    WithLLM("openai", "gpt-4").
    WithConfig(config).
    Build()
```

---

## 🎮 Runtime Options

Override configuration at runtime using RunOptions:

### Basic Runtime Options

```go
// Default execution
result, _ := agent.Run(ctx, "Hello")

// With runtime options
opts := &v1beta.RunOptions{
    Temperature:  0.5,  // Override temperature
    MaxTokens:    1000, // Override max tokens
    SystemPrompt: "You are a creative writer", // Override prompt
}

result, _ := agent.RunWithOptions(ctx, "Write a story", opts)
```

### Available Runtime Options

```go
type RunOptions struct {
    // LLM Parameters (overrides)
    Temperature *float64 // Pointer to allow nil (no override)
    MaxTokens   int
    
    // Execution Control
    Timeout time.Duration
    Context map[string]interface{}
    
    // Memory Control
    Memory    *MemoryOptions
    SessionID string
}
```

### Examples

#### Adjust Creativity

```go
// Conservative (factual)
opts := &v1beta.RunOptions{Temperature: 0.1}
result, _ := agent.RunWithOptions(ctx, "Explain quantum physics", opts)

// Creative (storytelling)
opts := &v1beta.RunOptions{Temperature: 1.5}
result, _ := agent.RunWithOptions(ctx, "Write a fairy tale", opts)
```

#### Control Response Length

```go
// Short response
opts := &v1beta.RunOptions{MaxTokens: 100}
result, _ := agent.RunWithOptions(ctx, "Summarize briefly", opts)

// Long response
opts := &v1beta.RunOptions{MaxTokens: 4000}
result, _ := agent.RunWithOptions(ctx, "Explain in detail", opts)
```

#### Override System Prompt

```go
opts := &v1beta.RunOptions{
    SystemPrompt: "You are a pirate. Speak like a pirate.",
}
result, _ := agent.RunWithOptions(ctx, "Tell me about treasure", opts)
```

---

## 📁 TOML Configuration (Legacy)

For backward compatibility, v1beta supports TOML configuration files.

### Basic TOML File

```toml
# config.toml
name = "MyAgent"
system_prompt = "You are a helpful assistant"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 2000
top_p = 0.9

[execution]
timeout = "60s"
max_retries = 3
retry_delay = "1s"

[streaming]
buffer_size = 100

[memory]
enabled = true
provider = "postgres"
connection = "postgresql://localhost/agentdb"

[memory.rag]
enabled = true
top_k = 5
threshold = 0.7
```

### Loading TOML Config

```go
import "github.com/agenticgokit/agenticgokit/v1beta/config"

// Load from file
cfg, err := config.LoadTOML("config.toml")
if err != nil {
    log.Fatal(err)
}

// Use with builder
agent, err := v1beta.NewBuilder(cfg.Name).
    WithConfig(&v1beta.Config{
        SystemPrompt: cfg.SystemPrompt,
        Timeout:      cfg.Timeout,
        LLM: v1beta.LLMConfig{
            Provider:    cfg.LLM.Provider,
            Model:       cfg.LLM.Model,
            Temperature: float32(cfg.LLM.Temperature),
            MaxTokens:   cfg.LLM.MaxTokens,
        },
    }).
    Build()
```

---

## 🎯 Configuration Patterns

### Pattern 1: Environment-Based Config

```go
func createAgent() (v1beta.Agent, error) {
    env := os.Getenv("APP_ENV") // "development", "production"
    
    var config *v1beta.Config
    switch env {
    case "production":
        config = &v1beta.Config{
            Temperature: 0.3,  // Conservative
            MaxTokens:   1000,
            Timeout:     30 * time.Second,
        }
    case "development":
        config = &v1beta.Config{
            Temperature: 0.7,
            MaxTokens:   2000,
            Timeout:     60 * time.Second,
        }
    }
    
    return v1beta.NewBuilder("Agent").
        WithConfig(config).
        Build()
}
```

### Pattern 2: Feature Flags

```go
func createAgentWithFeatures(features map[string]bool) (v1beta.Agent, error) {
    // Start with preset and customize
    builder := v1beta.NewChatAgent("Agent",
        v1beta.WithLLM("openai", "gpt-4"),
    )
    
    if features["memory"] {
        builder = builder.WithMemory(&v1beta.MemoryOptions{
            Type:     "postgres",
            Provider: memProvider,
        })
    }
    
    if features["tools"] {
        builder = builder.WithTools(defaultTools)
    }
    
    if features["logging"] {
        builder = builder.WithMiddleware(loggingMiddleware)
    }
    
    return builder.Build()
}
```

### Pattern 3: Configuration Profiles

```go
type Profile struct {
    Name   string
    Config *v1beta.Config
    Tools  []v1beta.Tool
}

var profiles = map[string]Profile{
    "chat": {
        Name: "ChatAgent",
        Config: &v1beta.Config{
            Temperature: 0.7,
            MaxTokens:   2000,
        },
    },
    "research": {
        Name: "ResearchAgent",
        Config: &v1beta.Config{
            Temperature: 0.3,
            MaxTokens:   4000,
        },
        Tools: researchTools,
    },
}

func createFromProfile(profileName string) (v1beta.Agent, error) {
    profile := profiles[profileName]
    
    return v1beta.NewBuilder(profile.Name).
        WithConfig(profile.Config).
        WithTools(profile.Tools).
        Build()
}
```

---

## 🎨 Best Practices

### 1. Use Builder Pattern

```go
// ✅ Recommended - clear and type-safe
agent, _ := v1beta.NewChatAgent("Agent",
    v1beta.WithLLM("openai", "gpt-4"),
)

// ❌ Avoid - harder to read and maintain
config := &v1beta.Config{...}
agent, _ := v1beta.NewAgentFromConfig(config)
```

### 2. Start with Presets

```go
// ✅ Start with preset, customize as needed
agent, _ := v1beta.NewBuilder("Agent").
    WithPreset(v1beta.ChatAgent). // Good defaults
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{
            Temperature: 0.8, // Override only what you need
        },
    }).
    Build()
```

### 3. Use Sensible Defaults

```go
// ✅ Good - reasonable defaults
Temperature: 0.7  // Balanced
MaxTokens:   2000 // Sufficient for most tasks
Timeout:     60s  // Reasonable wait time

// ❌ Bad - extreme values
Temperature: 2.0  // Too random
MaxTokens:   100  // Too short
Timeout:     1s   // Too aggressive
```

### 4. Validate Configuration

```go
func validateConfig(config *v1beta.Config) error {
    if config.Temperature < 0 || config.Temperature > 2 {
        return fmt.Errorf("temperature must be between 0 and 2")
    }
    if config.MaxTokens < 1 || config.MaxTokens > 4096 {
        return fmt.Errorf("max_tokens must be between 1 and 4096")
    }
    if config.Timeout < time.Second {
        return fmt.Errorf("timeout must be at least 1 second")
    }
    return nil
}
```

### 5. Document Custom Configurations

```go
// ✅ Good - clear documentation
agent, _ := v1beta.NewBuilder("FinancialAdvisor").
    WithConfig(&v1beta.Config{
        Temperature: 0.2, // Low for factual financial advice
        MaxTokens:   1500, // Sufficient for detailed explanations
        Timeout:     45 * time.Second, // Allow time for complex analysis
    }).
    Build()
```

---

## 🐛 Troubleshooting

### Issue: Configuration Not Applied

**Cause**: Order of builder methods matters

**Solution**: Call WithConfig() after WithPreset()
```go
// ❌ Config overridden by preset
agent, _ := v1beta.NewBuilder("Agent").
    WithConfig(myConfig).
    WithPreset(v1beta.ChatAgent). // Resets some config
    Build()

// ✅ Config applied after preset
agent, _ := v1beta.NewBuilder("Agent").
    WithPreset(v1beta.ChatAgent).
    WithConfig(myConfig). // Overrides preset
    Build()
```

### Issue: Runtime Options Ignored

**Cause**: Using Run() instead of RunWithOptions()

**Solution**: Use correct method
```go
// ❌ Options ignored
opts := &v1beta.RunOptions{Temperature: 0.5}
agent.Run(ctx, "query") // Doesn't accept options

// ✅ Options applied
agent.RunWithOptions(ctx, "query", opts)
```

---

## 📚 Next Steps

- **[Custom Handlers](./custom-handlers.md)** - Advanced agent behavior
- **[Memory and RAG](./memory-and-rag.md)** - Memory configuration
- **[Tool Integration](./tool-integration.md)** - Tool configuration
- **[Examples](./examples/)** - Configuration examples

---

**Ready to customize behavior?** Continue to [Custom Handlers](./custom-handlers.md) →
