# Core Concepts

Understanding the fundamental concepts of AgenticGoKit will help you build powerful AI agents. This guide covers the architecture and key components.

---

## 🎯 Overview

AgenticGoKit is built around four core concepts:

1. **Agents** - The primary interface for AI interactions
2. **Handlers** - Custom logic and augmentation functions
3. **Tools** - Extensible capabilities for agents
4. **Memory** - Context and knowledge retention

---

## 🤖 Agents

An **Agent** is the primary abstraction in v1beta. It encapsulates:
- LLM provider configuration
- Execution logic (handlers)
- Tool access
- Memory integration
- Middleware chain

### Agent Interface

Every agent implements this interface:

```go
type Agent interface {
    // Core execution methods
    Name() string
    Run(ctx context.Context, input string) (*Result, error)
    RunWithOptions(ctx context.Context, input string, opts *RunOptions) (*Result, error)

    // Streaming execution methods
    RunStream(ctx context.Context, input string, opts ...StreamOption) (Stream, error)
    RunStreamWithOptions(ctx context.Context, input string, runOpts *RunOptions, streamOpts ...StreamOption) (Stream, error)

    // Configuration access
    Config() *Config
    Capabilities() []string
    Memory() Memory

    // Lifecycle methods
    Initialize(ctx context.Context) error
    Cleanup(ctx context.Context) error
}
```

### Agent Types

AgenticGoKit provides several ways to create agents:

#### 1. Preset Builders (Recommended for Common Cases)

```go
// Chat agent - general purpose conversation
chatAgent, err := v1beta.NewBuilder("Assistant").
    WithPreset(v1beta.ChatAgent).
    Build()

// Research agent - optimized for analysis
researchAgent, err := v1beta.NewBuilder("Researcher").
    WithPreset(v1beta.ResearchAgent).
    Build()

// Quick agent - rapid prototyping (single parameter: model)
quickAgent, _ := v1beta.NewChatAgent("QuickAgent", v1beta.WithLLM("openai", "gpt-4"))
```

#### 2. Custom Builder (Full Control)

```go
// Create agent with full customization
agent, err := v1beta.NewBuilder("CustomAgent").
    WithConfig(&v1beta.Config{
        SystemPrompt: "You are a helpful assistant",
        LLM: v1beta.LLMConfig{
            Provider:    "openai",
            Model:       "gpt-4",
            Temperature: 0.7,
            MaxTokens:   2000,
        },
    }).
    WithTools(myTools).
    WithMemory(&v1beta.MemoryOptions{
        Type:     "simple",
        Provider: memProvider,
    }).
    WithHandler(myHandler).
    Build()
```

### Agent Lifecycle

```
Create Agent → Configure → Run/Stream → Process Result
     ↓            ↓           ↓              ↓
  Builder      Options    Middleware      Handler
```

1. **Create**: Use builder to construct agent
2. **Configure**: Set LLM, tools, memory, handlers
3. **Run/Stream**: Execute with context and query
4. **Process**: Handle result or stream chunks

---

## 🎨 Handlers

Handlers define how agents process queries. AgenticGoKit provides two handler types:

### 1. CustomHandlerFunc

Simple handler with LLM fallback capability:

```go
type CustomHandlerFunc func(
    ctx context.Context,
    query string,
    llmCall func(systemPrompt, userPrompt string) (string, error),
) (string, error)
```

**Use when:**
- You need simple custom logic
- You want automatic LLM fallback
- You don't need tool or memory access

**Example:**

```go
handler := func(ctx context.Context, query string, llmCall func(string, string) (string, error)) (string, error) {
    // Custom logic for specific patterns
    if strings.Contains(strings.ToLower(query), "time") {
        return fmt.Sprintf("Current time: %s", time.Now().Format(time.RFC3339)), nil
    }
    
    // Return empty string to fall back to LLM
    return "", nil
}

agent, _ := v1beta.NewBuilder("TimeAgent").
    WithPreset(v1beta.ChatAgent).
    WithHandler(handler).
    Build()
```

### 2. EnhancedHandlerFunc (AgentHandlerFunc)

Advanced handler with full capabilities:

```go
type EnhancedHandlerFunc func(
    ctx context.Context,
    query string,
    capabilities *HandlerCapabilities,
) (string, error)

type HandlerCapabilities struct {
    LLMCall  func(systemPrompt, userPrompt string) (string, error)
    ToolCall func(toolName string, args map[string]interface{}) (interface{}, error)
    Memory   MemoryProvider // Access to memory storage
    Config   *Config        // Agent configuration
}
```

**Use when:**
- You need access to tools
- You need memory integration
- You need full control over agent logic

**Example:**

```go
handler := func(ctx context.Context, query string, cap *v1beta.HandlerCapabilities) (string, error) {
    // Use tools
    weatherData, err := cap.ToolCall("get_weather", map[string]interface{}{
        "location": "New York",
    })
    if err != nil {
        return "", err
    }
    
    // Use LLM with tool data
    response, err := cap.LLMCall(
        "You are a weather assistant",
        fmt.Sprintf("Weather data: %v\nUser query: %s", weatherData, query),
    )
    
    return response, err
}

agent, _ := v1beta.NewBuilder("WeatherAgent").
    WithPreset(v1beta.ChatAgent).
    WithHandler(handler).
    WithTools(weatherTools).
    Build()
```

### Handler Augmentation

Pre-built augmentation functions automatically enhance handlers:

#### CreateToolAugmentedHandler

Automatically includes tool information in prompts:

```go
handler := v1beta.CreateToolAugmentedHandler(
    func(ctx context.Context, query, toolPrompt string, llmCall func(string, string) (string, error)) (string, error) {
        // toolPrompt contains formatted tool descriptions
        return llmCall("You are an assistant with tools", query)
    },
)
```

#### CreateMemoryAugmentedHandler

Automatically includes relevant memory context:

```go
handler := v1beta.CreateMemoryAugmentedHandler(
    func(ctx context.Context, query, memoryContext string, llmCall func(string, string) (string, error)) (string, error) {
        // memoryContext contains relevant past interactions
        return llmCall("You are an assistant with memory", query)
    },
)
```

#### CreateFullAugmentedHandler

Combines tool and memory augmentation:

```go
handler := v1beta.CreateFullAugmentedHandler(
    func(ctx context.Context, query, toolPrompt, memoryContext string, llmCall func(string, string) (string, error)) (string, error) {
        // Both toolPrompt and memoryContext available
        systemPrompt := fmt.Sprintf("Assistant with:\nTools: %s\nContext: %s", toolPrompt, memoryContext)
        return llmCall(systemPrompt, query)
    },
)
```

---

## 🛠️ Tools

Tools extend agent capabilities beyond LLM interactions.

### Tool Structure

```go
type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
}

type ToolInfo struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Parameters  map[string]interface{} `json:"parameters"`
    Category    string                 `json:"category"`
}
```

### Creating Tools

### Creating Tools

To create a tool, define a struct that implements the `Tool` interface:

```go
type SearchTool struct{}

func (t *SearchTool) Name() string {
    return "web_search"
}

func (t *SearchTool) Description() string {
    return "Search the web for information"
}

func (t *SearchTool) Execute(ctx context.Context, args map[string]interface{}) (*v1beta.ToolResult, error) {
    query, _ := args["query"].(string)
    // Implement search logic...
    return &v1beta.ToolResult{
        Success: true,
        Content: "Search results...",
    }, nil
}
```

### Adding Tools to Agents

Tools can be integrated via MCP (Model Context Protocol) or registered internally:

```go
// 1. Using MCP (Recommended)
mcpServer := v1beta.MCPServer{
    Name:    "filesystem",
    Type:    "stdio",
    Command: "mcp-server-filesystem",
    Enabled: true,
}

agent, err := v1beta.NewBuilder("ToolAgent").
    WithPreset(v1beta.ChatAgent).
    WithTools(
        v1beta.WithMCP(mcpServer),
        v1beta.WithToolTimeout(30*time.Second),
    ).
    Build()

// 2. Registering Internal Tools (Global)
// func init() {
//     v1beta.RegisterInternalTool("web_search", func() v1beta.Tool {
//         return &SearchTool{}
//     })
// }
```

### ToolCallHelper

Simplified tool execution in handlers:

```go
handler := func(ctx context.Context, query string, cap *v1beta.HandlerCapabilities) (string, error) {
    helper := v1beta.NewToolCallHelper(cap)
    
    // Call with map
    result, err := helper.Call("web_search", map[string]interface{}{
        "query": "Go programming",
        "max_results": 10,
    })
    
    // Call with struct
    type SearchParams struct {
        Query      string `json:"query"`
        MaxResults int    `json:"max_results"`
    }
    result, err = helper.CallWithStruct("web_search", SearchParams{
        Query:      "Go programming",
        MaxResults: 10,
    })
    
    return fmt.Sprintf("Search results: %v", result), nil
}
```

### MCP (Model Context Protocol) Tools

AgenticGoKit supports MCP for tool discovery:

```go
import "github.com/agenticgokit/agenticgokit/plugins/mcp"

// Discover tools from MCP servers
mcpTools, err := mcp.DiscoverTools(ctx, mcpServers...)

agent, err := v1beta.NewBuilder("MCPAgent").
    WithPreset(v1beta.ChatAgent).
    WithTools(mcpTools).
    Build()
```

---

## 💾 Memory

Memory provides context retention and knowledge storage. Starting from `v1beta`, memory is **enabled by default** using the `chromem` embedded provider.

### Memory Configuration

You can customize memory behavior or swap providers using the `WithMemory` option:

```go
agent, _ := v1beta.NewBuilder("Assistant").
    WithMemory(
        // Use a different provider
        v1beta.WithMemoryProvider("pgvector"),
        // Enable RAG with custom weights
        v1beta.WithRAG(4000, 0.3, 0.7),
        // Enable session-scoped isolation
        v1beta.WithSessionScoped(),
    ).
    Build()
```

If you need to strictly disable memory, use the `Enabled` flag:

```go
agent, _ := v1beta.NewBuilder("EphemeralAgent").
    WithConfig(&v1beta.Config{
        Memory: &v1beta.MemoryConfig{
            Enabled: false,
        },
    }).
    Build()
```

### Memory Interface

```go
type Memory interface {
    // Basic operations
    Store(ctx context.Context, content string, opts ...StoreOption) error
    Query(ctx context.Context, query string, opts ...QueryOption) ([]MemoryResult, error)

    // Session management
    NewSession() string
    SetSession(ctx context.Context, sessionID string) context.Context

    // RAG operations (if RAG is configured)
    IngestDocument(ctx context.Context, doc Document) error
    IngestDocuments(ctx context.Context, docs []Document) error
    SearchKnowledge(ctx context.Context, query string, opts ...QueryOption) ([]MemoryResult, error)
    BuildContext(ctx context.Context, query string, opts ...ContextOption) (*RAGContext, error)
}

type MemoryResult struct {
    Content   string                 `json:"content"`
    Score     float32                `json:"score"`
    Source    string                 `json:"source"`
    Metadata  map[string]interface{} `json:"metadata"`
    Timestamp time.Time              `json:"timestamp"`
}
```

### Memory Backends

#### In-Memory (Development)

```go
import "github.com/agenticgokit/agenticgokit/v1beta/memory"

memProvider := memory.NewInMemory()

agent, err := v1beta.NewBuilder("MemoryAgent").
    WithPreset(v1beta.ChatAgent).
    WithMemory(&v1beta.MemoryOptions{
        Type:     "simple",
        Provider: memProvider,
    }).
    Build()
```

#### PostgreSQL with pgvector (Production)

```go
// Register pgvector provider
import _ "github.com/agenticgokit/agenticgokit/plugins/memory/pgvector"

agent, err := v1beta.NewBuilder("PostgresAgent").
    WithPreset(v1beta.ChatAgent).
    WithMemory(
        v1beta.WithMemoryProvider("pgvector"),
        v1beta.WithConnection("postgresql://user:pass@localhost/db"),
    ).
    Build()
```

#### Other Providers
AgenticGoKit also supports **Weaviate** and a generic **In-Memory** provider for testing.

### RAG (Retrieval-Augmented Generation)

Memory automatically enables RAG context enrichment when configured:

```go
agent, err := v1beta.NewBuilder("RAGAgent").
    WithPreset(v1beta.ChatAgent).
    WithMemory(
        v1beta.WithRAG(2000, 0.5, 0.5), // maxTokens, personalWeight, knowledgeWeight
    ).
    Build()

// Agent automatically retrieves relevant context
result, err := agent.Run(context.Background(), "What did we discuss about Go?")
// Agent searches memory and augments query with relevant context
```

---

## 🔧 Configuration

### Builder Configuration

```go
agent, err := v1beta.NewBuilder("Agent").
    // LLM Configuration
    WithLLM("openai", "gpt-4").
    WithConfig(&v1beta.Config{
        SystemPrompt: "You are helpful",
        Temperature:  0.7,
        MaxTokens:    2000,
        TopP:         0.9,
        Timeout:      30 * time.Second,
        MaxRetries:   3,
        RetryDelay:   time.Second,
    }).
    
    // Components
    WithTools(tools).
    WithMemory(&v1beta.MemoryOptions{
        Type:     "simple",
        Provider: memProvider,
    }).
    WithHandler(handler).
    WithMiddleware(middleware).
    
    Build()
```

### Runtime Options

Override configuration at runtime:

```go
result, err := agent.RunWithOptions(
    ctx,
    "query",
    &v1beta.RunOptions{
        Temperature:  0.5,    // Override temperature
        MaxTokens:    1000,   // Override max tokens
        SystemPrompt: "...",  // Override system prompt
    },
)
```

### Config Struct

Direct configuration access:

```go
cfg := &v1beta.Config{
    SystemPrompt: "You are helpful",
    Timeout:      30 * time.Second,
    LLM: v1beta.LLMConfig{
        Temperature: 0.7,
        MaxTokens:   2000,
    },
}

agent, err := v1beta.NewBuilder("Agent").
    WithConfig(cfg).
    WithTools(tools).
    WithMemory(&v1beta.MemoryOptions{
        Type:     "simple",
        Provider: memProvider,
    }).
    WithHandler(handler).
    Build()
```

---

## 🔄 Execution Flow

### Standard Run Flow

```
User Query
    ↓
Context + Options
    ↓
Middleware (Before)
    ↓
Handler Selection
    ↓
Custom Handler?
    ↓ No
Default LLM Handler
    ↓ Yes
Custom Logic
    ↓
Tool Calls? ←---→ Tool Execution
    ↓
Memory Access? ←---→ Memory Retrieval
    ↓
LLM Call
    ↓
Response Processing
    ↓
Middleware (After)
    ↓
AgentResult
```

### Streaming Flow

```
User Query
    ↓
Context + Options + Channel
    ↓
Middleware (Before)
    ↓
Handler with Streaming
    ↓
LLM Stream
    ↓
Chunk Processing
    ↓
    ├→ ChunkTypeText ----→ Content chunks
    ├→ ChunkTypeDelta ---→ Token chunks
    ├→ ChunkTypeThought -→ Reasoning
    ├→ ChunkTypeToolCall → Tool execution
    ├→ ChunkTypeMetadata → Extra info
    ├→ ChunkTypeError ---→ Error handling
    └→ ChunkTypeDone ----→ Completion
    ↓
Middleware (After)
    ↓
Close Channel
```

---

## 🎭 Middleware

Middleware intercepts agent execution:

```go
type Middleware interface {
    BeforeRun(ctx context.Context, input string) (context.Context, string, error)
    AfterRun(ctx context.Context, input string, result *AgentResult, err error) (*AgentResult, error)
}
```

### Example: Logging Middleware

```go
type LoggingMiddleware struct{}

func (m *LoggingMiddleware) BeforeRun(ctx context.Context, input string) (context.Context, string, error) {
    log.Printf("Agent executing: %s", input)
    ctx = context.WithValue(ctx, "start_time", time.Now())
    return ctx, input, nil
}

func (m *LoggingMiddleware) AfterRun(ctx context.Context, input string, result *v1beta.Result, err error) (*v1beta.Result, error) {
    startTime := ctx.Value("start_time").(time.Time)
    duration := time.Since(startTime)
    log.Printf("Agent completed in %v: success=%t", duration, result.Success)
    return result, err
}

agent, err := v1beta.NewBuilder("LogAgent").
    WithPreset(v1beta.ChatAgent).
    WithMiddleware(&LoggingMiddleware{}).
    Build()
```

---

## 📊 Result Types

### Result

```go
type Result struct {
    FinalOutput  string                   // Response content
    Success      bool                     // Execution success
    Error        error                    // Error if failed
    Metadata     map[string]interface{}   // Additional data
    ToolCalls    []ToolCall               // Tools executed
    TokenUsage   *TokenUsage              // Token statistics
    StepResults  map[string]*StepResult   // For workflows
    IterationInfo *IterationInfo          // For loop workflows
}

type TokenUsage struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
}
```

### StreamChunk

```go
type StreamChunk struct {
    Type     ChunkType              // Chunk type
    Delta    string                 // Incremental content
    Content  string                 // Complete content (for non-delta)
    Metadata map[string]interface{} // Additional data
    Error    error                  // Error if any
    Done     bool                   // Is final chunk
}

// Chunk types
const (
    ChunkTypeDelta      ChunkType = "delta"       // Incremental token
    ChunkTypeContent    ChunkType = "content"     // Complete text
    ChunkTypeThought    ChunkType = "thought"     // Agent reasoning
    ChunkTypeToolCall   ChunkType = "tool_call"   // Tool execution
    ChunkTypeToolResult ChunkType = "tool_result" // Tool result
    ChunkTypeMetadata   ChunkType = "metadata"    // Extra information
    ChunkTypeError      ChunkType = "error"       // Error chunk
    ChunkTypeDone       ChunkType = "done"        // Completion marker
)
```

---

## 🎯 Best Practices

### 1. Choose the Right Handler Type
- Use **CustomHandlerFunc** for simple logic
- Use **EnhancedHandlerFunc** when you need tools/memory

### 2. Use Preset Builders
- Leverage `NewBuilder(name).WithPreset(ChatAgent)` for common cases
- Use `NewChatAgent(name, WithLLM(provider, model))` for rapid prototyping
- Use custom builder with full configuration for complex needs

### 3. Configure Timeouts
- Always set appropriate timeout values
- Use context cancellation for user-initiated stops

### 4. Handle Errors Properly
- Check both `err` and `result.Success`
- Use typed errors for specific handling

### 5. Leverage Middleware
- Add logging for debugging
- Add metrics for monitoring
- Add validation for input checking

### 6. Memory Management
- Use in-memory for development
- Use vector DB for production
- Configure appropriate retrieval limits

---

## 📚 Next Steps

- **[Streaming Guide](./streaming.md)** - Learn real-time streaming
- **[Workflows](./workflows.md)** - Multi-agent orchestration
- **[Tool Integration](./tool-integration.md)** - Extend capabilities
- **[Memory & RAG](./memory-and-rag.md)** - Add knowledge
- **[Examples](./examples/)** - See it in action

---

**Ready to dive deeper?** Continue to [Streaming Guide](./streaming.md) →
