# Configuration Guide

AgenticGoKit v1beta offers layered configuration options so you can start with a working agent quickly and extend it as your scenario requires. This guide lays out the recommended builder path, alternative patterns, runtime knobs, and how to fall back to the struct-based configuration you may already know from earlier releases.

## 🎯 Overview

- **Recommended path**: use the builder pattern with `NewChatAgent` (or `NewResearchAgent`, `NewDataAgent`, `NewWorkflowAgent`) to declare intent, then add features (tools, memory, handlers) as needed.
- **When you need more control**: the legacy `Config` struct still exists for advanced scenarios or programmatic assembly of configuration fragments.
- **Other angles**: runtime options allow behavior tuning without changing code, and the TOML loader remains for legacy deployments.

Start with the builder pattern and move on to the other sections only if you hit a blocker.

## 🚀 Quick Start: Builder Pattern

The builder helpers are the fastest way to express what your agent should do, and they keep configuration declarative by chaining options on the constructor.

### Your First Agent

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

    _ = agent
}
```

1. Pick `WithLLM`, `WithPreset`, or `WithConfig` to describe the model and prompting strategy.
2. Stretch the agent with `WithTools`, `WithMemory`, or other helpers when the scenario demands it.
3. When you need to tune runtime behavior, consult the runtime options section below.

## 🔧 Configuring Your Agent

### Essential Builder Methods (choose one)

```go
func WithPreset(preset PresetType) Builder { ... }
```

Use `WithPreset` when you want a pre-configured agent type. Available presets:
- `v1beta.ChatAgent` - Conversational agent (temperature: 0.8, context-aware memory)
- `v1beta.ResearchAgent` - Research agent (temperature: 0.3, tools enabled, extended timeout)
- `v1beta.DataAgent` - Data analysis agent (temperature: 0.1, precise responses)
- `v1beta.WorkflowAgent` - Workflow orchestration (temperature: 0.5, workflow config)

```go
func WithConfig(config *Config) Builder { ... }
```

`WithConfig` lets you pass a fully populated `Config` in one shot, which is helpful when you compose configuration from multiple sources (e.g., CLI flags plus persisted defaults, or loading from TOML files).

```go
func WithLLM(provider, model string) Option { ... }
```

`WithLLM` is used with convenience constructors like `NewChatAgent()` to specify the LLM provider and model. This is the most common starting point.

**Example:**
```go
agent, err := v1beta.NewChatAgent("Assistant",
    v1beta.WithLLM("openai", "gpt-4"),
)
```

**Other standalone options:**
- `WithSystemPrompt(prompt string)` - Set custom system prompt
- `WithAgentTimeout(timeout time.Duration)` - Set execution timeout
- `WithDebugMode(enabled bool)` - Enable debug logging

### Feature Helpers (optional)

```go
func WithTools(opts ...ToolOption) Builder { ... }
```

Add tools to your agent using functional options. Available tool options:
- `WithMCP(servers ...MCPServer)` - Connect to MCP servers
- `WithMCPDiscovery(scanPorts ...int)` - Auto-discover MCP servers
- `WithToolTimeout(timeout time.Duration)` - Set tool execution timeout
- `WithMaxConcurrentTools(max int)` - Limit parallel tool executions
- `WithToolCaching(ttl time.Duration)` - Enable result caching

**Example:**
```go
builder := v1beta.NewBuilder("agent").
    WithPreset(v1beta.ResearchAgent).
    WithTools(
        v1beta.WithMCPDiscovery(),
        v1beta.WithToolTimeout(30 * time.Second),
    )
```

```go
func WithMemory(opts ...MemoryOption) Builder { ... }
```

Configure memory using functional options. Available memory options:
- `WithMemoryProvider(provider string)` - Set memory backend ("chromem", "pgvector", "weaviate")
- `WithRAG(maxTokens int, personalWeight, knowledgeWeight float32)` - Enable RAG with weights
- `WithSessionScoped()` - Enable session-scoped memory
- `WithContextAware()` - Enable context-aware memory

**Example:**
```go
builder := v1beta.NewBuilder("agent").
    WithPreset(v1beta.ChatAgent).
    WithMemory(
        v1beta.WithMemoryProvider("chromem"),
        v1beta.WithRAG(4096, 0.7, 0.3),
        v1beta.WithSessionScoped(),
    )
```

```go
func WithHandler(handler HandlerFunc) Builder { ... }
```

Register a custom handler function that implements your agent's core logic. The handler receives capabilities for LLM, tools, and memory.

**Example:**
```go
handler := func(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    // Call LLM
    response, err := caps.LLM("You are helpful", input)
    if err != nil {
        return "", err
    }
    return response, nil
}

agent, _ := v1beta.NewBuilder("custom").
    WithPreset(v1beta.ChatAgent).
    WithHandler(handler).
    Build()
```

### Advanced Builders (optional)

```go
func WithWorkflow(opts ...WorkflowOption) Builder { ... }
```

Configure workflow orchestration for multi-agent scenarios:
- `WithWorkflowMode(mode string)` - Set mode: "sequential", "parallel", "dag", "loop"
- `WithWorkflowAgents(agents ...string)` - Specify agent names in workflow
- `WithMaxIterations(max int)` - Limit workflow iterations

```go
func WithSubWorkflow(opts ...BuilderSubWorkflowOption) Builder { ... }
```

Wrap a workflow as an agent for hierarchical composition:
- `WithWorkflowInstance(workflow Workflow)` - Set the workflow to wrap
- `WithSubWorkflowMaxDepthBuilder(depth int)` - Limit nesting depth
- `WithSubWorkflowDescriptionBuilder(description string)` - Add description

```go
func Clone() Builder { ... }
```

Create a copy of the builder for reuse with different configurations.

## 🎮 Runtime Options

Runtime options allow you to override configuration per execution without rebuilding the agent. Pass `RunOptions` to `agent.RunWithOptions()`.

```go
type RunOptions struct {
    // Tool configuration
    Tools           []string `json:"tools"`           // Specific tools to enable
    ToolMode        string   `json:"tool_mode"`       // "auto", "specific", "none"
    
    // Memory configuration
    Memory          *MemoryOptions `json:"memory"`    // Memory settings for this run
    SessionID       string   `json:"session_id"`       // Session identifier
    
    // Execution configuration
    Timeout         time.Duration `json:"timeout"`    // Execution timeout
    Context         map[string]interface{} `json:"context"` // Additional context
    MaxRetries      int      `json:"max_retries"`      // Maximum retry attempts
    
    // Performance configuration
    MaxTokens       int      `json:"max_tokens"`       // Override max tokens
    Temperature     *float64 `json:"temperature"`      // Override temperature
    
    // Result configuration
    DetailedResult  bool     `json:"detailed_result"`  // Return detailed execution info
    IncludeTrace    bool     `json:"include_trace"`    // Include trace data
    IncludeSources  bool     `json:"include_sources"`  // Include source attributions
    
    // Multimodal input
    Images          []ImageData `json:"images"`       // Images to include
    Audio           []AudioData `json:"audio"`        // Audio to include
    Video           []VideoData `json:"video"`        // Video to include
}
```

Common use cases:

- **Override temperature**: Adjust creativity per request
- **Override timeout**: Give more time for complex tasks
- **Specify tools**: Enable/disable tools for specific requests
- **Detailed results**: Get execution metrics for debugging
- **Session management**: Group related conversations

Runtime options are ideal when behavior varies per request or user preference.

## 🎯 Configuration Patterns

Pick the pattern that matches your team’s workflow. These snippets illustrate how to reuse the builder model across common scenarios.

### Compose by Preset

```go
// Use a preset with additional features
agent, err := v1beta.NewBuilder("ResearcherAgent").
    WithPreset(v1beta.ResearchAgent).
    WithTools(
        v1beta.WithMCPDiscovery(),
    ).
    WithMemory(
        v1beta.WithRAG(4096, 0.7, 0.3),
    ).
    Build()
```

This is the highest-level approach. Presets package prompts, models, tools, and defaults so new agents stay consistent. Then add only the features you need.

### Compose Programmatically

```go
config := &v1beta.Config{
    Name:         "custom",
    SystemPrompt: "You are a specialized assistant",
    Timeout:      60 * time.Second,
    LLM: v1beta.LLMConfig{
        Provider:    "openai",
        Model:       "gpt-4",
        Temperature: 0.7,
        MaxTokens:   2048,
    },
    Tracing: &v1beta.TracingConfig{
        Enabled: true,
        Level:   "debug",
    },
}
agent, err := v1beta.NewBuilder("CustomAgent").
    WithConfig(config).
    Build()
```

Use this when configuration is assembled from multiple inputs or loaded from files.

### Feature Switches via Runtime Options

```go
// Build agent once
agent, err := v1beta.NewChatAgent("FlexibleAgent",
    v1beta.WithLLM("openai", "gpt-4"),
)

// Adjust behavior per request
temperature := 0.9
runOpts := &v1beta.RunOptions{
    Temperature:    &temperature,
    MaxTokens:      1000,
    DetailedResult: true,
    SessionID:      "user-123",
}

result, err := agent.RunWithOptions(ctx, "Tell me a story", runOpts)
```

Runtime options are best when behavior varies per request, user, or session.

## 📋 Configuration Struct (Advanced)

`Config` exposes every configuration field and is still supported, mainly for teams migrating from Go SDK v0.x or when you need to build configuration dynamically.

Use the struct when:

1. You need to merge fragments from files, databases, or REST APIs before handing them to the builder.
2. You are writing tooling that inspects or validates configuration before creating an agent.
3. Builder helpers do not yet cover the knob you must set.

```go
type Config struct {
    // Core settings
    Name         string        `toml:"name"`
    SystemPrompt string        `toml:"system_prompt"`
    Timeout      time.Duration `toml:"timeout"`
    DebugMode    bool          `toml:"debug_mode"`
    
    // LLM configuration
    LLM LLMConfig `toml:"llm"`
    
    // Feature configurations
    Memory    *MemoryConfig    `toml:"memory,omitempty"`
    Tools     *ToolsConfig     `toml:"tools,omitempty"`
    Workflow  *WorkflowConfig  `toml:"workflow,omitempty"`
    Tracing   *TracingConfig   `toml:"tracing,omitempty"`
    Streaming *StreamingConfig `toml:"streaming,omitempty"`
}

type LLMConfig struct {
    Provider    string  `toml:"provider"`    // "openai", "anthropic", "ollama", etc.
    Model       string  `toml:"model"`       // Model name
    Temperature float32 `toml:"temperature"` // 0.0 to 2.0
    MaxTokens   int     `toml:"max_tokens"`  // Maximum tokens
    BaseURL     string  `toml:"base_url,omitempty"`
    APIKey      string  `toml:"api_key,omitempty"`
}
```

Pass the struct via `WithConfig()` on the builder, or use it with convenience constructors. The builder pattern is still recommended; use `Config` directly only when you need programmatic assembly.

## 🎨 Best Practices

- **Start with presets** for consistency. Each preset encapsulates prompts, models, tools, and runtime defaults. Add or override only what needs to change.
- **Layer defaults** by declaring the core behavior (model, preset, runtime strategy) first, then add auxiliary features like memory or handlers.
- **Validate early**: call `agent.ValidateConfig()` after you build the agent so runtime problems surface during initialization.

```
func validateConfig(agent *v1beta.Agent) error {
    if agent == nil {
        return errors.New("agent is nil")
    }
    return agent.ValidateConfig()
}
```

- **Document custom behavior**: if you deviate from the default builder path, add a short section in your README or architecture guide explaining why.

## 🐛 Troubleshooting

- **`agent lifecycle invalid spec`**: Happens when you call both `WithPreset` and manually configure the prompt/model in a way that conflicts with the preset defaults. Stick to one path—either preset or manual configuration, not both.
- **Presets not picking up runtime options**: Confirm you pass `WithRuntimeOptions`. Runtime fields on `Config` are only respected when they go through the runtime-aware helper.

## 📚 Next Steps

- Review [getting started](../../README.md) for project setup.
- Dive into the [runtime guide](../reference/runtime.md) to understand execution strategies.
- Explore [memory and tools](../reference/memory.md) to extend agent capabilities.

Ready to customize agent behavior? Use the builder flow above, then add your handlers, toolchains, or runtime tweaks.

## 📁 Legacy: TOML Configuration

TOML-based configuration remains for users migrating from older releases. Treat it as a serialization of the `Config` struct—you can always translate TOML into the struct and continue using builder helpers.

```toml
id = "legacy-agent"

[model]
provider = "openai"
model = "gpt-4"

tools = []

[runtime_options]
strategy = "sync"
```

Load the TOML and pass it through `WithConfig`:

```
config, err := v1beta.LoadConfigFromFile("agent.toml")
if err != nil {
    return err
}
agent, err := v1beta.NewChatAgent("LegacyAgent",
    v1beta.WithConfig(config),
)
```

Only keep TOML if you must share configuration files with existing workflows. Otherwise, prefer the builder helpers and runtime switches described earlier.
