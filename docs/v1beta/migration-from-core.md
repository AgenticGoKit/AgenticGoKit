# Migration Guide: Core API → v1beta

This guide helps you migrate from the legacy `core` package APIs to the new `v1beta` package. The v1beta package provides a cleaner, more powerful API while maintaining the core functionality you rely on.

---

## 🎯 Migration Overview

The v1beta package brings significant improvements:

- **Simplified interfaces** - Reduced from 30+ methods to 8 core methods
- **Unified options** - One `RunOptions` struct instead of many separate configs
- **Enhanced streaming** - Dedicated `Stream` interface with 8 chunk types
- **Cleaner configuration** - Single config file with nested sections
- **Better error handling** - Structured errors with actionable suggestions
- **Preset builders** - Quick agent creation for common use cases
- **Type-safe workflows** - Explicit workflow types (Sequential, Parallel, DAG, Loop)

**Estimated migration time:** 1-3 hours for typical projects

---

## 📦 Import Changes

### Update Imports

**Before (core):**
```go
import (
    "github.com/agenticgokit/agenticgokit/core"
)
```

**After (v1beta):**
```go
import (
    "github.com/agenticgokit/agenticgokit/v1beta"
)
```

Or with alias:
```go
import (
    v1 "github.com/agenticgokit/agenticgokit/v1beta"
)
```

---

## 🤖 Agent Creation

### Builder Pattern Changes

**Before (core):**
```go
builder := core.NewAgentBuilder().
    SetName("ChatBot").
    SetSystemPrompt("You are a helpful assistant").
    SetLLMProvider("openai").
    SetLLMModel("gpt-4").
    SetLLMTemperature(0.7).
    SetLLMMaxTokens(1000).
    SetLLMTopP(0.9).
    EnableStreaming(true).
    SetStreamingBufferSize(100)
    
agent, err := builder.Build()
```

**After (v1beta):**
```go
agent, err := v1beta.NewBuilder("ChatBot").
    WithPreset(v1beta.ChatAgent).
    WithLLM("openai", "gpt-4").
    WithConfig(&v1beta.Config{
        SystemPrompt: "You are a helpful assistant",
        Temperature:  0.7,
        MaxTokens:    1000,
        TopP:         0.9,
    }).
    Build()
```

**Key changes:**
- Use `NewBuilder(name)` with `WithPreset(ChatAgent)` pattern
- Consolidate LLM settings with `WithLLM(provider, model)`
- Pass configuration options via `WithConfig()` struct
- Streaming is always available via `RunStream()` method

### Preset Builders

v1beta provides preset builders for common scenarios:

```go
// Chat assistant with preset
chatAgent, _ := v1beta.NewBuilder("Assistant").
    WithPreset(v1beta.ChatAgent).
    Build()

// Research agent with preset
researchAgent, _ := v1beta.NewBuilder("Researcher").
    WithPreset(v1beta.ResearchAgent).
    WithMemory(&v1beta.MemoryOptions{
        Type:     "inmemory",
        Provider: memProvider,
    }).
    Build()

// Custom agent from scratch
customAgent, _ := v1beta.NewBuilder("Custom").
    WithLLM("openai", "gpt-4").
    WithHandler(myCustomHandler).
    Build()
```

### Quick Agent Creation

**Before (core):**
```go
agent, err := core.NewAgentBuilder().
    SetLLMProvider("openai").
    SetLLMModel("gpt-4").
    Build()
```

**After (v1beta):**
```go
// Single line for common cases
agent, err := v1beta.QuickChatAgent("gpt-4")
```

---

## 🚀 Agent Execution

### Basic Execution

**Before (core):**
```go
result, err := agent.Run(ctx, input)
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Content)
```

**After (v1beta):**
```go
// Same interface!
result, err := agent.Run(ctx, input)
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Content)
```

### Execution with Options

**Before (core):**
```go
result, err := agent.RunWithOptions(ctx, input, &core.Options{
    MaxTokens:   1000,
    Temperature: 0.7,
    TopP:        0.9,
})
```

**After (v1beta):**
```go
// Use RunOptions struct
opts := &v1beta.RunOptions{
    MaxTokens:   1000,
    Temperature: 0.7,
    TopP:        0.9,
}
result, err := agent.RunWithOptions(ctx, input, opts)
```

**Key changes:**
- Use `RunWithOptions()` with `RunOptions` struct
- No functional options at Run() level (use struct fields)
- Cleaner and more explicit

---

## 📡 Streaming

Streaming has been completely redesigned in v1beta.

### Simple Text Streaming

**Before (core):**
```go
err := agent.RunWithStreaming(ctx, input, func(text string) {
    fmt.Print(text)
})
```

**After (v1beta):**
```go
stream, err := agent.RunStream(ctx, input)
if err != nil {
    log.Fatal(err)
}

for chunk := range stream.Chunks() {
    if chunk.Type == v1beta.ChunkTypeDelta {
        fmt.Print(chunk.Delta)
    }
}

result, err := stream.Wait()
```

### Stream Interface Benefits

**After (v1beta only):**
```go
stream, err := agent.RunStream(ctx, input)

// Process chunks
for chunk := range stream.Chunks() {
    switch chunk.Type {
    case v1beta.ChunkTypeContent:
        fmt.Printf("Text: %s\n", chunk.Content)
        
    case v1beta.ChunkTypeDelta:
        fmt.Print(chunk.Delta) // Incremental text
        
    case v1beta.ChunkTypeThought:
        log.Printf("💭 Thinking: %s\n", chunk.Content)
        
    case v1beta.ChunkTypeToolCall:
        log.Printf("🔧 Calling tool: %s\n", chunk.ToolName)
        
    case v1beta.ChunkTypeToolResult:
        log.Printf("✓ Tool result: %s\n", chunk.Content)
        
    case v1beta.ChunkTypeMetadata:
        log.Printf("Metadata: %v\n", chunk.Metadata)
        
    case v1beta.ChunkTypeError:
        log.Printf("❌ Error: %s\n", chunk.Error)
        
    case v1beta.ChunkTypeDone:
        log.Println("✓ Complete")
    }
}

result, err := stream.Wait()
```

**Key improvements:**
- 8 distinct chunk types for granular control
- Separate thoughts from output
- Track tool calls and results
- Receive metadata and errors in-stream
- Always call `stream.Wait()` to get final result

---

## ⚙️ Configuration

### Configuration Files

**Before (core):**
```toml
# llm.toml
provider = "openai"
model = "gpt-4"

# memory.toml
provider = "memory"
connection = "inmemory"

# tools.toml
enabled = true
max_retries = 3
```

**After (v1beta):**
```toml
# config.toml - Everything in one file
name = "MyAgent"
system_prompt = "You are a helpful assistant"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 2000

[memory]
provider = "memory"
connection = "inmemory"

[tools]
enabled = true
max_retries = 3
timeout = "30s"

[streaming]
buffer_size = 100
include_thoughts = true
text_only = false
```

### Loading Configuration

**Before (core):**
```go
config, err := core.LoadConfig("llm.toml")
memConfig, err := core.LoadMemoryConfig("memory.toml")
toolsConfig, err := core.LoadToolsConfig("tools.toml")
```

**After (v1beta):**
```go
// Single unified config
config, err := v1beta.LoadConfig("config.toml")

// Use with builder
agent, err := v1beta.NewBuilder("MyAgent").
    WithLLM(config.LLM.Provider, config.LLM.Model).
    WithConfig(config).
    Build()
```

**Key changes:**
- Single config file with nested sections
- All configurations in one place
- Easier to manage and version control

---

## 🔄 Workflows

### Sequential Workflow

**Before (core):**
```go
workflow := core.NewWorkflow("Pipeline")
workflow.AddStep("step1", agent1, "Do task 1")
workflow.AddStep("step2", agent2, "Do task 2")
workflow.SetMode(core.Sequential)

result, err := workflow.Execute(ctx, input)
```

**After (v1beta):**
```go
config := &v1beta.WorkflowConfig{
    Mode:    v1beta.Sequential,
    Timeout: 120 * time.Second,
}

workflow, err := v1beta.NewSequentialWorkflow(config)
if err != nil {
    log.Fatal(err)
}

workflow.AddStep(v1beta.WorkflowStep{Name: "step1", Agent: agent1})
workflow.AddStep(v1beta.WorkflowStep{Name: "step2", Agent: agent2})

result, err := workflow.Run(ctx, input)
```

### Parallel Workflow

**Before (core):**
```go
workflow := core.NewWorkflow("MultiTask")
workflow.AddStep("task1", agent1, "Task 1")
workflow.AddStep("task2", agent2, "Task 2")
workflow.SetMode(core.Parallel)

result, err := workflow.Execute(ctx, input)
```

**After (v1beta):**
```go
config := &v1beta.WorkflowConfig{
    Mode:    v1beta.Parallel,
    Timeout: 120 * time.Second,
}

workflow, err := v1beta.NewParallelWorkflow(config)
if err != nil {
    log.Fatal(err)
}

workflow.AddStep(v1beta.WorkflowStep{Name: "task1", Agent: agent1})
workflow.AddStep(v1beta.WorkflowStep{Name: "task2", Agent: agent2})

result, err := workflow.Run(ctx, input)
```

### DAG Workflow

**Before (core):**
```go
workflow := core.NewWorkflow("DAG")
workflow.AddStep("step1", agent1, "Step 1", nil) // No dependencies
workflow.AddStep("step2", agent2, "Step 2", []string{"step1"})
workflow.AddStep("step3", agent3, "Step 3", []string{"step1"})
workflow.SetMode(core.DAG)

result, err := workflow.Execute(ctx, input)
```

**After (v1beta):**
```go
config := &v1beta.WorkflowConfig{
    Mode:    v1beta.DAG,
    Timeout: 120 * time.Second,
}

workflow, err := v1beta.NewDAGWorkflow(config)
if err != nil {
    log.Fatal(err)
}

workflow.AddStep(v1beta.WorkflowStep{
    Name:         "step1",
    Agent:        agent1,
    Dependencies: nil,
})
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "step2",
    Agent:        agent2,
    Dependencies: []string{"step1"},
})
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "step3",
    Agent:        agent3,
    Dependencies: []string{"step1"},
})

result, err := workflow.Run(ctx, input)
```

### Loop Workflow

**Before (core):**
```go
workflow := core.NewWorkflow("Loop")
workflow.AddStep("step1", agent1, "Iterate")
workflow.SetMode(core.Loop)
workflow.SetStopCondition(func(r map[string]*Result) bool {
    return r["step1"].Metadata["done"].(bool)
})

result, err := workflow.Execute(ctx, input)
```

**After (v1beta):**
```go
shouldContinue := func(ctx context.Context, iteration int, lastResult *v1beta.WorkflowResult) (bool, error) {
    if iteration >= 5 {
        return false, nil // Max iterations
    }
    if lastResult != nil {
        if done, ok := lastResult.Metadata["done"].(bool); ok {
            return !done, nil
        }
    }
    return true, nil
}

config := &v1beta.WorkflowConfig{
    Mode:          v1beta.Loop,
    Timeout:       300 * time.Second,
    MaxIterations: 5,
}

workflow, err := v1beta.NewLoopWorkflowWithCondition(config, shouldContinue)
if err != nil {
    log.Fatal(err)
}

workflow.AddStep(v1beta.WorkflowStep{Name: "step1", Agent: agent1})

result, err := workflow.Run(ctx, input)
```

**Key changes:**
- Workflow mode explicit in constructor (type-safe)
- `Execute()` renamed to `Run()` for consistency
- Loop condition function signature changed (includes context, iteration, lastResult)
- Use `WorkflowStep` struct instead of helper functions

### Workflow Streaming

**After (v1beta only):**
```go
config := &v1beta.WorkflowConfig{
    Mode:    v1beta.Sequential,
    Timeout: 120 * time.Second,
}

workflow, _ := v1beta.NewSequentialWorkflow(config)
workflow.AddStep(v1beta.WorkflowStep{Name: "step1", Agent: agent1})
workflow.AddStep(v1beta.WorkflowStep{Name: "step2", Agent: agent2})

stream, err := workflow.RunStream(ctx, input)

for chunk := range stream.Chunks() {
    if step, ok := chunk.Metadata["step_name"].(string); ok {
        fmt.Printf("→ Step: %s\n", step)
    }
    if chunk.Type == v1beta.ChunkTypeDelta {
        fmt.Print(chunk.Delta)
    }
}

result, err := stream.Wait()
```

---

## 🛠️ Tools

### Tool Definition

**Before (core):**
```go
tool := core.Tool{
    Name:        "calculator",
    Description: "Performs calculations",
    Parameters: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "expression": map[string]interface{}{
                "type":        "string",
                "description": "Math expression to evaluate",
            },
        },
        "required": []string{"expression"},
    },
    Execute: func(args map[string]interface{}) (string, error) {
        expr := args["expression"].(string)
        result := evaluate(expr)
        return fmt.Sprintf("%v", result), nil
    },
}
```

**After (v1beta):**
```go
// Cleaner tool handler signature
tool := v1beta.Tool{
    Name:        "calculator",
    Description: "Performs calculations",
    Parameters: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "expression": map[string]interface{}{
                "type":        "string",
                "description": "Math expression to evaluate",
            },
        },
        "required": []string{"expression"},
    },
    Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
        expr := args["expression"].(string)
        result := evaluate(expr)
        return result, nil
    },
}
```

### Registering Tools

**Before (core):**
```go
agent, _ := core.NewAgentBuilder().
    AddTool(tool1).
    AddTool(tool2).
    Build()
```

**After (v1beta):**
```go
// Same interface
agent, _ := v1beta.NewBuilder("ToolAgent").
    WithPreset(v1beta.ChatAgent).
    WithTools([]v1beta.Tool{tool1, tool2}).
    Build()
```

---

## 💾 Memory & RAG

### In-Memory Storage

**Before (core):**
```go
agent, _ := core.NewAgentBuilder().
    SetMemoryProvider("memory").
    SetMemoryConnection("inmemory").
    EnableRAG(true).
    Build()
```

**After (v1beta):**
```go
memProvider := memory.NewInMemory()

agent, _ := v1beta.NewBuilder("ResearchAgent").
    WithPreset(v1beta.ResearchAgent).
    WithMemory(&v1beta.MemoryOptions{
        Type:     "inmemory",
        Provider: memProvider,
        RAG: &v1beta.RAGConfig{
            Enabled:   true,
            TopK:      5,
            Threshold: 0.7,
        },
    }).
    Build()
// RAG automatically enabled with configuration
```

### Session Management

**Before (core):**
```go
result1, _ := agent.Run(ctx, "Hello", &core.Options{
    SessionID: "session-123",
})

result2, _ := agent.Run(ctx, "Continue", &core.Options{
    SessionID: "session-123",
})
```

**After (v1beta):**
```go
// Sessions are managed by memory provider context
ctx1 := context.WithValue(ctx, "session_id", "session-123")
result1, _ := agent.Run(ctx1, "Hello")

ctx2 := context.WithValue(ctx, "session_id", "session-123")
result2, _ := agent.Run(ctx2, "Continue")
// Memory automatically retrieves context from same session
```

---

## ❌ Error Handling

### Basic Error Handling

**Before (core):**
```go
result, err := agent.Run(ctx, input)
if err != nil {
    log.Printf("Error: %v", err)
    return err
}
```

**After (v1beta):**
```go
// Same for simple cases
result, err := agent.Run(ctx, input)
if err != nil {
    log.Printf("Error: %v", err)
    return err
}
```

### Structured Error Handling

**After (v1beta only):**
```go
result, err := agent.Run(ctx, input)
if err != nil {
    // Check for structured error
    if agentErr, ok := err.(*v1beta.AgentError); ok {
        switch agentErr.Code {
        case "llm_error":
            log.Printf("LLM error: %s", agentErr.Message)
            log.Printf("Suggestion: %s", agentErr.Details)
            // Retry with different model
            
        case "timeout":
            log.Printf("Timeout error: %s", agentErr.Message)
            // Increase timeout
            
        case "tool_error":
            log.Printf("Tool failed: %s", agentErr.Message)
            // Disable tool and retry
            
        case "memory_error":
            log.Printf("Memory error: %s", agentErr.Message)
            // Clear memory and retry
            
        default:
            log.Printf("Unknown error: %s", agentErr.Message)
        }
    }
    return err
}
```

**Error structure:**
```go
type AgentError struct {
    Code    string
    Message string
    Details string
    Cause   error
}
```

---

## 🧪 Testing

### Mock Agents

**Before (core):**
```go
mockAgent := &core.MockAgent{
    RunFunc: func(ctx context.Context, input string) (*core.Result, error) {
        return &core.Result{Content: "mock response"}, nil
    },
}
```

**After (v1beta):**
```go
mockAgent := &v1beta.MockAgent{
    RunFunc: func(ctx context.Context, input string) (*v1beta.Result, error) {
        return &v1beta.Result{FinalOutput: "mock response"}, nil
    },
}
```

---

## 📋 Migration Checklist

Use this checklist to track your migration:

### Imports
- [ ] Update imports to `v1beta` package
- [ ] Remove old `core` imports

### Agent Creation
- [ ] Replace `NewAgentBuilder()` with `NewBuilder(name).WithPreset()`
- [ ] Consolidate `SetXXX()` methods into `WithXXX()` methods
- [ ] Use `WithLLM(provider, model)` instead of separate setters
- [ ] Use `WithConfig()` for temperature, tokens, etc.

### Agent Execution
- [ ] Keep basic `Run()` calls (no changes needed)
- [ ] Use `RunWithOptions()` with `RunOptions` struct
- [ ] Replace callback-based streaming with `RunStream()` + Stream interface

### Streaming
- [ ] Replace string callbacks with Stream interface
- [ ] Add `stream.Wait()` calls after consuming chunks
- [ ] Use `chunk.Delta` for incremental text
- [ ] Use `chunk.Type` for filtering/routing

### Configuration
- [ ] Merge multiple config files into single `config.toml`
- [ ] Update config struct references
- [ ] Add nested sections for different components

### Workflows
- [ ] Replace `NewWorkflow()` with type-specific constructors
- [ ] Use `WorkflowConfig` + `AddStep(WorkflowStep{...})`
- [ ] Replace `Execute()` with `Run()`
- [ ] Update loop condition function signatures (ctx, iteration, lastResult)

### Error Handling
- [ ] Add structured error handling where needed
- [ ] Use error codes for specific cases
- [ ] Log error suggestions for debugging

### Testing
- [ ] Update mock implementations
- [ ] Test streaming functionality
- [ ] Verify workflow behavior

---

## 🚀 Quick Migration Examples

### Example 1: Simple Chat Agent

**Before (core):**
```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/agenticgokit/agenticgokit/core"
)

func main() {
    agent, err := core.NewAgentBuilder().
        SetName("Assistant").
        SetSystemPrompt("You are helpful").
        SetLLMProvider("openai").
        SetLLMModel("gpt-4").
        Build()
    if err != nil {
        log.Fatal(err)
    }
    
    result, err := agent.Run(context.Background(), "Hello!")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(result.Content)
}
```

**After (v1beta):**
```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    agent, err := v1beta.NewBuilder("Assistant").
        WithPreset(v1beta.ChatAgent).
        WithLLM("openai", "gpt-4").
        WithConfig(&v1beta.Config{
            SystemPrompt: "You are helpful",
        }).
        Build()
    if err != nil {
        log.Fatal(err)
    }
    
    result, err := agent.Run(context.Background(), "Hello!")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(result.FinalOutput)
}
```

### Example 2: Streaming Agent

**Before (core):**
```go
func streamResponse(agent core.Agent, query string) error {
    return agent.RunWithStreaming(
        context.Background(),
        query,
        func(text string) {
            fmt.Print(text)
        },
    )
}
```

**After (v1beta):**
```go
func streamResponse(agent v1beta.Agent, query string) error {
    stream, err := agent.RunStream(context.Background(), query)
    if err != nil {
        return err
    }
    
    for chunk := range stream.Chunks() {
        if chunk.Type == v1beta.ChunkTypeDelta {
            fmt.Print(chunk.Delta)
        }
    }
    
    _, err = stream.Wait()
    return err
}
```

### Example 3: Multi-Agent Workflow

**Before (core):**
```go
workflow := core.NewWorkflow("Research")
workflow.AddStep("search", searchAgent, "Find info")
workflow.AddStep("analyze", analyzeAgent, "Analyze {{.search}}")
workflow.AddStep("summarize", summaryAgent, "Summarize {{.analyze}}")
workflow.SetMode(core.Sequential)

results, err := workflow.Execute(context.Background(), "AI trends")
```

**After (v1beta):**
```go
config := &v1beta.WorkflowConfig{
    Mode:    v1beta.Sequential,
    Timeout: 180 * time.Second,
}

workflow, err := v1beta.NewSequentialWorkflow(config)
if err != nil {
    log.Fatal(err)
}

workflow.AddStep(v1beta.WorkflowStep{Name: "search", Agent: searchAgent})
workflow.AddStep(v1beta.WorkflowStep{Name: "analyze", Agent: analyzeAgent})
workflow.AddStep(v1beta.WorkflowStep{Name: "summarize", Agent: summaryAgent})

result, err := workflow.Run(context.Background(), "AI trends")
```

---

## 🆘 Common Migration Issues

### Issue 1: Import Errors

**Problem:** `cannot find package "core"`

**Solution:** Update module dependency
```bash
go get github.com/agenticgokit/agenticgokit/v1beta@latest
go mod tidy
```

### Issue 2: Type Mismatches

**Problem:** `cannot use agent (type core.Agent) as type v1beta.Agent`

**Solution:** Rebuild agents with v1beta builders

### Issue 3: Missing Methods

**Problem:** `agent.RunWithStreaming undefined`

**Solution:** Use `RunStream()` method instead

### Issue 4: Streaming Doesn't Wait

**Problem:** Program exits before streaming completes

**Solution:** Always call `stream.Wait()` at the end

---

## 📚 Additional Resources

- **[Getting Started](./getting-started.md)** - Quick start guide for v1beta
- **[Core Concepts](./core-concepts.md)** - Understanding v1beta architecture
- **[Streaming Guide](./streaming.md)** - Complete streaming documentation
- **[Workflows Guide](./workflows.md)** - Multi-agent workflows
- **[API Reference](./api-reference.md)** - Complete API documentation

---

## 💡 Migration Tips

1. **Start with tests** - Migrate test files first to validate behavior
2. **Use presets** - Leverage preset builders for faster migration
3. **Migrate incrementally** - Update one component at a time
4. **Check deprecations** - Look for deprecation warnings in logs
5. **Test streaming** - Streaming behavior has changed significantly
6. **Update configs** - Merge multiple config files early
7. **Review errors** - Take advantage of new structured errors

---

## 🎉 Migration Complete!

After migration, you'll have:

✅ Cleaner, more readable code  
✅ Better type safety  
✅ Enhanced streaming capabilities  
✅ Improved error handling  
✅ Unified configuration  
✅ Faster development with presets  

**Questions?** Check [Troubleshooting](./troubleshooting.md) or open an issue.

---

**Ready to build?** Continue to [Getting Started](./getting-started.md) →
