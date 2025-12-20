# Migration Guide: Legacy APIs to v1beta

**Complete guide to migrating from `core/vnext` to the stable v1beta API.**

---

## Overview

v1beta is the modern, stable API for AgenticGoKit. It provides:

- **Simplified Builder API**: Cleaner, more intuitive agent construction
- **Improved Error Handling**: Typed errors with detailed context
- **Better Type Safety**: Compile-time checks for configuration
- **Enhanced Streaming**: More reliable real-time response handling
- **Workflow Improvements**: Easier composition and state management

**Timeline:**
- v1beta is production-ready and recommended for all new projects
- Legacy APIs (`core/vnext`) remain available but are deprecated
- Legacy support ends 6 months after v1.0 release

---

## Quick Migration Checklist

- [ ] Update imports: `core/vnext` → `v1beta`
- [ ] Replace agent construction with builder pattern
- [ ] Update LLM configuration syntax
- [ ] Migrate memory provider setup
- [ ] Update workflow definitions
- [ ] Replace streaming handlers
- [ ] Update error handling
- [ ] Test thoroughly

---

## Import Changes

### Before (Legacy)
```go
import (
    "github.com/agenticgokit/agenticgokit/core/vnext"
)
```

### After (v1beta)
```go
import (
    "github.com/agenticgokit/agenticgokit/v1beta"
)
```

**Find and Replace:**
- `github.com/agenticgokit/agenticgokit/core/vnext` → `github.com/agenticgokit/agenticgokit/v1beta`
- `vnext.` → `v1beta.`

---

## Basic Agent Creation

### Before (Legacy)
```go
agent := &vnext.Agent{
    Name: "MyAgent",
    LLMConfig: vnext.LLMConfig{
        Provider: "openai",
        Model:    "gpt-4",
        APIKey:   os.Getenv("OPENAI_API_KEY"),
    },
    SystemPrompt: "You are a helpful assistant",
}

err := agent.Initialize(ctx)
if err != nil {
    log.Fatal(err)
}

response, err := agent.Run(ctx, "Hello")
```

### After (v1beta)
```go
agent, err := v1beta.NewBuilder("MyAgent").
    WithLLM("openai", "gpt-4").
    WithSystemPrompt("You are a helpful assistant").
    Build()
if err != nil {
    log.Fatal(err)
}

result, err := agent.Run(ctx, "Hello")
if err != nil {
    log.Fatal(err)
}
response := result.Content
```

**Key Changes:**
- Builder pattern replaces struct initialization
- `Build()` validates configuration before creating agent
- No separate `Initialize()` call needed
- Result contains structured output with metadata

---

## LLM Provider Configuration

### Before (Legacy)
```go
// OpenAI
agent := &vnext.Agent{
    LLMConfig: vnext.LLMConfig{
        Provider:    "openai",
        Model:       "gpt-4",
        APIKey:      os.Getenv("OPENAI_API_KEY"),
        Temperature: 0.7,
        MaxTokens:   1000,
    },
}

// Azure OpenAI
agent := &vnext.Agent{
    LLMConfig: vnext.LLMConfig{
        Provider: "azure",
        Model:    "gpt-4",
        AzureConfig: &vnext.AzureConfig{
            Endpoint:    os.Getenv("AZURE_ENDPOINT"),
            APIKey:      os.Getenv("AZURE_API_KEY"),
            Deployment:  "gpt-4-deployment",
            APIVersion:  "2024-02-01",
        },
    },
}

// Ollama
agent := &vnext.Agent{
    LLMConfig: vnext.LLMConfig{
        Provider: "ollama",
        Model:    "llama3",
        OllamaConfig: &vnext.OllamaConfig{
            BaseURL: "http://localhost:11434",
        },
    },
}
```

### After (v1beta)
```go
// OpenAI
agent, _ := v1beta.NewBuilder("Agent").
    WithLLM("openai", "gpt-4").
    WithTemperature(0.7).
    WithMaxTokens(1000).
    Build()

// Azure OpenAI
agent, _ := v1beta.NewBuilder("Agent").
    WithAzureLLM(
        os.Getenv("AZURE_ENDPOINT"),
        "gpt-4-deployment",
        "2024-02-01",
    ).
    Build()

// Ollama
agent, _ := v1beta.NewBuilder("Agent").
    WithOllamaLLM("llama3", "http://localhost:11434").
    Build()
```

**Key Changes:**
- Dedicated methods for each provider
- API keys read from environment automatically
- Cleaner, more explicit configuration
- Compile-time validation of required parameters

---

## Memory & RAG

### Before (Legacy)
```go
agent := &vnext.Agent{
    Name: "Agent",
    MemoryConfig: &vnext.MemoryConfig{
        Provider: "pgvector",
        ConnectionString: "postgresql://user:pass@localhost/db",
        RAGEnabled: true,
        RAGConfig: &vnext.RAGConfig{
            ContextSize:      4000,
            DiversityWeight:  0.3,
            RelevanceWeight:  0.7,
        },
    },
}
```

### After (v1beta)
```go
agent, _ := v1beta.NewBuilder("Agent").
    WithLLM("openai", "gpt-4").
    WithMemory("pgvector", "postgresql://user:pass@localhost/db").
    WithRAG(4000, 0.3, 0.7).
    Build()
```

**Key Changes:**
- `WithMemory()` for provider setup
- `WithRAG()` for retrieval configuration
- Simplified parameter passing
- Memory and RAG are independent options

**Memory Providers:**
- `pgvector`: PostgreSQL with pgvector extension
- `weaviate`: Weaviate vector database
- `inmemory`: In-memory storage (non-persistent)

---

## Streaming

### Before (Legacy)
```go
stream, err := agent.RunStream(ctx, "Hello")
if err != nil {
    log.Fatal(err)
}

for chunk := range stream {
    if chunk.Type == vnext.ChunkTypeText {
        fmt.Print(chunk.Content)
    } else if chunk.Type == vnext.ChunkTypeError {
        log.Printf("Error: %v", chunk.Error)
    }
}
```

### After (v1beta)
```go
stream, err := agent.RunStream(ctx, "Hello")
if err != nil {
    log.Fatal(err)
}

for chunk := range stream {
    switch chunk.Type {
    case v1beta.ChunkText:
        fmt.Print(chunk.Content)
    case v1beta.ChunkError:
        log.Printf("Error: %v", chunk.Error)
    case v1beta.ChunkDone:
        fmt.Println("\nComplete")
    }
}
```

**Key Changes:**
- Enum-style chunk types (compile-time safe)
- More chunk types: `ChunkThought`, `ChunkToolCall`, `ChunkToolResult`, `ChunkMetadata`
- Explicit `ChunkDone` signal
- Better error propagation

---

## Workflows

### Sequential Workflow

#### Before (Legacy)
```go
workflow := &vnext.SequentialWorkflow{
    Steps: []vnext.WorkflowStep{
        {Agent: agent1, Name: "research"},
        {Agent: agent2, Name: "summarize"},
        {Agent: agent3, Name: "write"},
    },
}

result, err := workflow.Execute(ctx, "AI trends")
```

#### After (v1beta)
```go
workflow := v1beta.NewSequentialWorkflow().
    AddStep("research", agent1).
    AddStep("summarize", agent2).
    AddStep("write", agent3)

result, err := workflow.Run(ctx, "AI trends")
```

### Parallel Workflow

#### Before (Legacy)
```go
workflow := &vnext.ParallelWorkflow{
    Agents: []vnext.Agent{agent1, agent2, agent3, agent4},
}

results, err := workflow.Execute(ctx, "Analyze this text")
```

#### After (v1beta)
```go
workflow := v1beta.NewParallelWorkflow().
    AddAgent("sentiment", agent1).
    AddAgent("topics", agent2).
    AddAgent("summary", agent3).
    AddAgent("keywords", agent4)

results, err := workflow.Run(ctx, "Analyze this text")
```

### DAG Workflow

#### Before (Legacy)
```go
workflow := &vnext.DAGWorkflow{
    Nodes: map[string]vnext.DAGNode{
        "fetch": {Agent: agent1, Dependencies: []string{}},
        "process": {Agent: agent2, Dependencies: []string{"fetch"}},
        "analyze": {Agent: agent3, Dependencies: []string{"fetch"}},
        "report": {Agent: agent4, Dependencies: []string{"process", "analyze"}},
    },
}

result, err := workflow.Execute(ctx, input)
```

#### After (v1beta)
```go
workflow := v1beta.NewDAGWorkflow().
    AddNode("fetch", agent1, nil).
    AddNode("process", agent2, []string{"fetch"}).
    AddNode("analyze", agent3, []string{"fetch"}).
    AddNode("report", agent4, []string{"process", "analyze"})

result, err := workflow.Run(ctx, input)
```

### Loop Workflow

#### Before (Legacy)
```go
workflow := &vnext.LoopWorkflow{
    Agent: agent,
    MaxIterations: 5,
    ConvergenceFunc: func(output string) bool {
        return strings.Contains(output, "FINAL")
    },
}

result, err := workflow.Execute(ctx, "Write code and refine")
```

#### After (v1beta)
```go
workflow := v1beta.NewLoopWorkflow(agent).
    WithMaxIterations(5).
    WithConvergenceFunc(func(result v1beta.Result) bool {
        return strings.Contains(result.Content, "FINAL")
    })

result, err := workflow.Run(ctx, "Write code and refine")
```

**Key Changes:**
- Builder pattern for all workflows
- Named agents for better debugging
- `Run()` instead of `Execute()`
- Consistent result types across workflows

---

## Custom Handlers

### Before (Legacy)
```go
type MyHandler struct{}

func (h *MyHandler) Handle(ctx context.Context, input string, llm vnext.LLM) (string, error) {
    // Pre-process
    input = strings.ToLower(input)
    
    // Call LLM
    response, err := llm.Generate(ctx, "System prompt", input)
    if err != nil {
        return "", err
    }
    
    // Post-process
    return strings.ToUpper(response), nil
}

agent := &vnext.Agent{
    Name: "Agent",
    Handler: &MyHandler{},
}
```

### After (v1beta)
```go
func myHandler(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    // Pre-process
    input = strings.ToLower(input)
    
    // Call LLM
    response, err := caps.LLM("System prompt", input)
    if err != nil {
        return "", err
    }
    
    // Post-process
    return strings.ToUpper(response), nil
}

agent, _ := v1beta.NewBuilder("Agent").
    WithLLM("openai", "gpt-4").
    WithHandler(myHandler).
    Build()
```

**Key Changes:**
- Function-based handlers instead of interface
- `Capabilities` object provides access to LLM, Tools, Memory
- Simpler implementation
- Better composability

---

## Error Handling

### Before (Legacy)
```go
result, err := agent.Run(ctx, input)
if err != nil {
    // Generic error handling
    log.Printf("Error: %v", err)
    return
}
```

### After (v1beta)
```go
result, err := agent.Run(ctx, input)
if err != nil {
    // Type-specific error handling
    var agentErr *v1beta.AgentError
    if errors.As(err, &agentErr) {
        switch agentErr.Code {
        case v1beta.ErrCodeLLM:
            log.Printf("LLM error: %v", agentErr)
        case v1beta.ErrCodeMemory:
            log.Printf("Memory error: %v", agentErr)
        case v1beta.ErrCodeTimeout:
            log.Printf("Timeout: %v", agentErr)
        default:
            log.Printf("Unknown error: %v", agentErr)
        }
    }
    return
}
```

**Key Changes:**
- Typed errors with error codes
- Rich error context (component, operation, details)
- Error wrapping for root cause analysis
- Retry hints for transient errors

**Error Codes:**
- `ErrCodeInvalidInput`: Invalid input parameters
- `ErrCodeLLM`: LLM provider errors
- `ErrCodeMemory`: Memory system errors
- `ErrCodeTool`: Tool execution errors
- `ErrCodeTimeout`: Operation timeout
- `ErrCodeWorkflow`: Workflow execution errors

---

## Tool Integration

### Before (Legacy)
```go
agent := &vnext.Agent{
    Name: "Agent",
    Tools: []vnext.Tool{
        {
            Name: "search",
            Func: func(query string) (string, error) {
                return performSearch(query)
            },
        },
    },
}
```

### After (v1beta)
```go
searchTool := func(args map[string]interface{}) (string, error) {
    query := args["query"].(string)
    return performSearch(query)
}

agent, _ := v1beta.NewBuilder("Agent").
    WithLLM("openai", "gpt-4").
    WithTool("search", "Search the web", searchTool).
    Build()
```

**Key Changes:**
- Explicit tool descriptions
- Map-based arguments for flexibility
- Tools passed directly to builder
- Better LLM integration

---

## Configuration Options

### Before (Legacy)
```go
agent := &vnext.Agent{
    Name: "Agent",
    LLMConfig: vnext.LLMConfig{
        Temperature: 0.7,
        MaxTokens:   1000,
        TopP:        0.9,
    },
    SystemPrompt: "You are helpful",
    Timeout:      30 * time.Second,
}
```

### After (v1beta)
```go
agent, _ := v1beta.NewBuilder("Agent").
    WithLLM("openai", "gpt-4").
    WithSystemPrompt("You are helpful").
    WithTemperature(0.7).
    WithMaxTokens(1000).
    WithTopP(0.9).
    WithTimeout(30 * time.Second).
    Build()
```

**Available Options:**
- `WithSystemPrompt(prompt string)`
- `WithTemperature(temp float64)`
- `WithMaxTokens(max int)`
- `WithTopP(p float64)`
- `WithTimeout(duration time.Duration)`
- `WithRetries(max int)`
- `WithBackoff(strategy BackoffStrategy)`

---

## Testing

### Before (Legacy)
```go
func TestAgent(t *testing.T) {
    agent := &vnext.Agent{
        Name: "TestAgent",
        LLMConfig: vnext.LLMConfig{
            Provider: "mock",
            MockResponses: []string{"Hello!"},
        },
    }
    
    result, err := agent.Run(context.Background(), "Hi")
    assert.NoError(t, err)
    assert.Equal(t, "Hello!", result)
}
```

### After (v1beta)
```go
func TestAgent(t *testing.T) {
    mockLLM := v1beta.NewMockLLM([]string{"Hello!"})
    
    agent, _ := v1beta.NewBuilder("TestAgent").
        WithMockLLM(mockLLM).
        Build()
    
    result, err := agent.Run(context.Background(), "Hi")
    assert.NoError(t, err)
    assert.Equal(t, "Hello!", result.Content)
}
```

**Key Changes:**
- Dedicated mock LLM provider
- Cleaner test setup
- Better assertion helpers
- Structured result validation

---

## Common Migration Patterns

### Pattern 1: Simple Chat Agent

**Before:**
```go
agent := &vnext.Agent{
    Name: "ChatBot",
    LLMConfig: vnext.LLMConfig{
        Provider: "openai",
        Model:    "gpt-4",
    },
    SystemPrompt: "You are a helpful chatbot",
}
agent.Initialize(ctx)
response, _ := agent.Run(ctx, "Hello")
```

**After:**
```go
agent, _ := v1beta.NewBuilder("ChatBot").
    WithLLM("openai", "gpt-4").
    WithSystemPrompt("You are a helpful chatbot").
    Build()

result, _ := agent.Run(ctx, "Hello")
response := result.Content
```

### Pattern 2: RAG Agent

**Before:**
```go
agent := &vnext.Agent{
    Name: "RAGAgent",
    LLMConfig: vnext.LLMConfig{Provider: "openai", Model: "gpt-4"},
    MemoryConfig: &vnext.MemoryConfig{
        Provider: "pgvector",
        ConnectionString: connStr,
        RAGEnabled: true,
        RAGConfig: &vnext.RAGConfig{
            ContextSize: 4000,
        },
    },
}
agent.Initialize(ctx)
agent.StoreMemory(ctx, "Document content")
response, _ := agent.Run(ctx, "Query")
```

**After:**
```go
agent, _ := v1beta.NewBuilder("RAGAgent").
    WithLLM("openai", "gpt-4").
    WithMemory("pgvector", connStr).
    WithRAG(4000, 0.3, 0.7).
    Build()

agent.StoreMemory(ctx, "Document content")
result, _ := agent.Run(ctx, "Query")
response := result.Content
```

### Pattern 3: Multi-Agent Pipeline

**Before:**
```go
agent1 := &vnext.Agent{Name: "A1", LLMConfig: config1}
agent2 := &vnext.Agent{Name: "A2", LLMConfig: config2}
agent1.Initialize(ctx)
agent2.Initialize(ctx)

workflow := &vnext.SequentialWorkflow{
    Steps: []vnext.WorkflowStep{
        {Agent: agent1}, {Agent: agent2},
    },
}
result, _ := workflow.Execute(ctx, "Input")
```

**After:**
```go
agent1, _ := v1beta.NewBuilder("A1").WithLLM("openai", "gpt-4").Build()
agent2, _ := v1beta.NewBuilder("A2").WithLLM("openai", "gpt-4").Build()

workflow := v1beta.NewSequentialWorkflow().
    AddStep("step1", agent1).
    AddStep("step2", agent2)

result, _ := workflow.Run(ctx, "Input")
```

---

## Troubleshooting

### Issue: Import Errors

**Problem:**
```
cannot find package "github.com/agenticgokit/agenticgokit/v1beta"
```

**Solution:**
```bash
go get github.com/agenticgokit/agenticgokit/v1beta@latest
go mod tidy
```

### Issue: Build Errors

**Problem:**
```
agent.Build() returns error: "LLM provider not configured"
```

**Solution:**
Ensure you call `WithLLM()` or provider-specific method before `Build()`:
```go
agent, err := v1beta.NewBuilder("Agent").
    WithLLM("openai", "gpt-4").  // Required
    Build()
```

### Issue: API Key Not Found

**Problem:**
```
LLM error: API key not configured
```

**Solution:**
Set environment variable or pass explicitly:
```bash
export OPENAI_API_KEY="sk-..."
```

Or:
```go
agent, _ := v1beta.NewBuilder("Agent").
    WithLLM("openai", "gpt-4").
    WithAPIKey(os.Getenv("OPENAI_API_KEY")).
    Build()
```

### Issue: Memory Provider Connection

**Problem:**
```
Memory error: connection refused
```

**Solution:**
Verify connection string and database is running:
```bash
# For pgvector
docker run -p 5432:5432 -e POSTGRES_PASSWORD=password ankane/pgvector

# Update connection string
agent, _ := v1beta.NewBuilder("Agent").
    WithMemory("pgvector", "postgresql://postgres:password@localhost:5432/postgres").
    Build()
```

### Issue: Streaming Not Working

**Problem:**
```
stream channel closed immediately
```

**Solution:**
Ensure you iterate until `ChunkDone`:
```go
stream, _ := agent.RunStream(ctx, "Hello")
for chunk := range stream {
    if chunk.Type == v1beta.ChunkDone {
        break
    }
    // Process chunk
}
```

---

## Migration Timeline

**Recommended Approach:**

1. **Week 1: Preparation**
   - Review this migration guide
   - Identify all agent creation points
   - Update dependencies: `go get v1beta@latest`

2. **Week 2: Core Migration**
   - Update imports
   - Migrate basic agent creation
   - Update error handling

3. **Week 3: Advanced Features**
   - Migrate workflows
   - Update memory/RAG configuration
   - Migrate custom handlers

4. **Week 4: Testing & Validation**
   - Comprehensive testing
   - Performance validation
   - Production deployment

---

## Getting Help

- **[v1beta Documentation](v1beta/README.md)** - Complete API reference
- **[Examples](v1beta/examples/)** - Working code samples
- **[GitHub Discussions](https://github.com/agenticgokit/agenticgokit/discussions)** - Community support
- **[Issues](https://github.com/agenticgokit/agenticgokit/issues)** - Bug reports

---

## Next Steps

1. **[Getting Started Guide](v1beta/getting-started.md)** - v1beta fundamentals
2. **[Core Concepts](v1beta/core-concepts.md)** - Deep dive into architecture
3. **[Examples](v1beta/examples/)** - Real-world patterns
4. **[API Versioning](API_VERSIONING.md)** - Version strategy and support
