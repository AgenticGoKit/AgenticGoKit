# AgenticGoKit Examples

This guide provides practical examples of building AI agents and workflows with AgenticGoKit, from simple single-agent applications to complex multi-agent orchestrations.

## Table of Contents

- [Quick Start Examples](#quick-start-examples)
- [Single Agent Examples](#single-agent-examples)
- [Multi-Agent Workflows](#multi-agent-workflows)
- [Tool Integration Examples](#tool-integration-examples)
- [Production Examples](#production-examples)
- [Custom Provider Examples](#custom-provider-examples)

## Quick Start Examples

### Simple Query Agent (5 minutes)

The fastest way to create an agent that can answer questions (config + handler):

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Build runner from config (use [llm] type = "ollama", model = "gemma3:1b")
    runner, err := core.NewRunnerFromConfig("agentflow.toml")
    if err != nil { log.Fatal(err) }

    // Minimal agent handler
    _ = runner.RegisterAgent("helper", core.AgentHandlerFunc(func(ctx context.Context, ev core.Event, st core.State) (core.AgentResult, error) {
        out := st.Clone()
        out.Set("result", "Paris")
        return core.AgentResult{OutputState: out}, nil
    }))

    ctx := context.Background()
    if err := runner.Start(ctx); err != nil { log.Fatal(err) }
    defer runner.Stop()

    st := core.NewState()
    st.Set("query", "What is the capital of France?")
    if err := runner.Emit(core.NewEvent("helper", st.GetAll(), map[string]string{"session_id": "demo-1"})); err != nil { log.Fatal(err) }
    fmt.Println("Emitted; check logs/callbacks for output")
}
```

### Multi-Agent Orchestration (Quick Start)

Generate complete multi-agent workflows with the CLI (config-driven):

```bash
# Collaborative workflow - all agents work in parallel
agentcli create research-system \
    --orchestration-mode collaborative \
    --collaborative-agents "researcher,analyzer,validator" \
    --visualize \
    --mcp-enabled

# Sequential pipeline - agents process one after another
agentcli create data-pipeline \
    --orchestration-mode sequential \
    --sequential-agents "collector,processor,formatter" \
    --visualize

# Loop-based workflow - single agent repeats with conditions
agentcli create quality-loop \
    --orchestration-mode loop \
    --loop-agent "quality-checker" \
    --max-iterations 5 \
    --visualize
```

### Using CLI Scaffolding

Generate a complete project in seconds:

```bash
# Create a new project with mixed orchestration
agentcli create my-ai-app \
    --orchestration-mode mixed \
    --collaborative-agents "analyzer,validator" \
    --sequential-agents "processor,reporter" \
    --visualize-output "docs/diagrams" \
    --mcp-enabled

cd my-ai-app

# Run with any query
go run . -m "analyze market trends and generate comprehensive report"
```

The generated project includes:
- Multi-agent orchestration configuration
- Automatic workflow visualization (Mermaid diagrams)
- MCP tool integration
- Error handling and fault tolerance
- Logging and tracing
- Production-ready structure

## Single Agent Examples

### Research Agent

An MCP-enabled agent that can use tools to gather information:

```go
// Ensure you import the MCP plugin and the LLM provider plugin in your project:
// _ "github.com/kunalkushwaha/agenticgokit/plugins/mcp/default"
// _ "github.com/kunalkushwaha/agenticgokit/plugins/llm/ollama"

type ResearchAgent struct { agent core.Agent }

func NewResearchAgent(name string) (*ResearchAgent, error) {
    // Create a basic MCP-aware agent (manager initialized elsewhere)
    // or wrap a handler if your logic is simple.
    return &ResearchAgent{agent: &myResearchAgent{name: name}}, nil
}

// myResearchAgent is a minimal example using Agent interface
type myResearchAgent struct { name string }
func (a *myResearchAgent) Name() string { return a.name }
// Implement other Agent methods or use a handler in real apps
```

### Data Analysis Agent

An agent specialized for analyzing and processing data:

```go
type DataAnalysisAgent struct {
    agent core.Agent
}

func (d *DataAnalysisAgent) AnalyzeData(ctx context.Context, data map[string]interface{}) (*AnalysisResult, error) {
    state := core.NewState()
    state.Set("task", "data_analysis")
    state.Set("data", data)
    state.Set("query", "Analyze this data and provide insights, trends, and recommendations.")
    
    result, err := d.agent.Run(ctx, state)
    if err != nil {
        return nil, err
    }
    
    return &AnalysisResult{
        Summary:         result.GetResult(),
        Confidence:      result.GetFloat("confidence"),
        Recommendations: result.GetStringSlice("recommendations"),
    }, nil
}

type AnalysisResult struct {
    Summary         string
    Confidence      float64
    Recommendations []string
}
```

## Multi-Agent Workflows

### Collaborative Research System

All agents work in parallel to process the same task (config-driven):

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Build runner from agentflow.toml (mode: collaborative)
    runner, err := core.NewRunnerFromConfig("agentflow.toml")
    if err != nil { log.Fatal(err) }

    // Register agents
    _ = runner.RegisterAgent("researcher", NewResearchAgent())
    _ = runner.RegisterAgent("analyzer", NewAnalysisAgent())
    _ = runner.RegisterAgent("validator", NewValidationAgent())

    // Create event
    event := core.NewEvent("all", map[string]interface{}{
        "task": "research AI trends and provide comprehensive analysis",
    }, nil)

    // Start runner and emit event
    ctx := context.Background()
    if err := runner.Start(ctx); err != nil { log.Fatal(err) }
    defer runner.Stop()

    if err := runner.Emit(event); err != nil { log.Fatal(err) }
    fmt.Println("Collaborative workflow started; check logs for outputs")
}
```

### Sequential Data Pipeline

Agents process data in sequence, each building on the previous result:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Build runner from agentflow.toml (mode: sequential)
    runner, err := core.NewRunnerFromConfig("agentflow.toml")
    if err != nil { log.Fatal(err) }

    _ = runner.RegisterAgent("collector", NewDataCollectorAgent())
    _ = runner.RegisterAgent("processor", NewDataProcessorAgent())
    _ = runner.RegisterAgent("formatter", NewDataFormatterAgent())

    // Create pipeline event
    event := core.NewEvent("pipeline", map[string]interface{}{
        "data_source": "market_data",
        "format":      "json",
    }, nil)

    // Start runner and emit event
    ctx := context.Background()
    if err := runner.Start(ctx); err != nil { log.Fatal(err) }
    defer runner.Stop()

    if err := runner.Emit(event); err != nil { log.Fatal(err) }
    fmt.Println("Sequential pipeline started; check logs for outputs")
}
```

### Loop-Based Quality Checker

Single agent repeats execution until quality standards are met:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Build runner from agentflow.toml (mode: loop)
    runner, err := core.NewRunnerFromConfig("agentflow.toml")
    if err != nil { log.Fatal(err) }

    _ = runner.RegisterAgent("quality-checker", NewQualityCheckerAgent())

    // Create quality check event
    event := core.NewEvent("loop", map[string]interface{}{
        "content":           "document to check",
        "quality_threshold": 0.95,
    }, nil)

    // Start runner and emit event
    ctx := context.Background()
    if err := runner.Start(ctx); err != nil { log.Fatal(err) }
    defer runner.Stop()

    if err := runner.Emit(event); err != nil { log.Fatal(err) }
    fmt.Println("Quality loop started; check logs for outputs")
}
```

### Mixed Orchestration Workflow

Combine collaborative and sequential patterns via configuration:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Build runner from agentflow.toml (mode: mixed)
    runner, err := core.NewRunnerFromConfig("agentflow.toml")
    if err != nil { log.Fatal(err) }

    // Register agents referenced by config
    _ = runner.RegisterAgent("analyzer", NewAnalyzerAgent())
    _ = runner.RegisterAgent("validator", NewValidatorAgent())
    _ = runner.RegisterAgent("processor", NewProcessorAgent())
    _ = runner.RegisterAgent("reporter", NewReporterAgent())

    // Create mixed workflow event
    event := core.NewEvent("mixed", map[string]interface{}{
        "task": "analyze data, validate results, process, and generate report",
    }, nil)

    // Start runner and emit event
    ctx := context.Background()
    if err := runner.Start(ctx); err != nil { log.Fatal(err) }
    defer runner.Stop()

    if err := runner.Emit(event); err != nil { log.Fatal(err) }
    fmt.Println("Mixed workflow started; check logs for outputs")
}
```

### Workflow Visualization

Use the CLI to generate diagrams. Runtime code doesn’t emit diagrams.

Text: Use `agentcli create ... --visualize` or `agentcli visualize` to create Mermaid diagrams.

### Research and Analysis Pipeline

A workflow where one agent researches and another analyzes:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Create LLM provider (Ollama gemma3:1b for examples)
    llm, _ := core.NewLLMProvider(core.AgentLLMConfig{Provider: "ollama", Model: "gemma3:1b"})
    
    // Create agents
    researcher, _ := core.NewMCPAgent("researcher", llm)
    analyst, _ := core.NewMCPAgent("analyst", llm)
    
    // Build a runner from config-driven setup
    runner, _ := core.NewRunnerFromConfig("agentflow.toml")
    _ = runner.RegisterAgent("researcher", researcher)
    _ = runner.RegisterAgent("analyst", analyst)
    
    // Create initial state
    state := core.NewState()
    state.Set("topic", "artificial intelligence trends 2024")
    
    // Run workflow
    ctx := context.Background()
    
    // Step 1: Research
    state.Set("query", "Research the latest trends in artificial intelligence for 2024")
    result1, err := researcher.Run(ctx, state)
    if err != nil {
        log.Fatal(err)
    }
    
    // Step 2: Analysis
    state.Set("research_data", result1.GetResult())
    state.Set("query", "Analyze the research data and provide key insights and predictions")
    result2, err := analyst.Run(ctx, state)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Final Analysis:", result2.GetResult())
}
```

### Parallel Processing Workflow

Use collaborative mode for parallel fan-out via configuration. Example uses Start/Emit/Stop with a config-driven runner:

```go
ctx := context.Background()
runner, _ := core.NewRunnerFromConfig("agentflow.toml")
// Register agents...
_ = runner.Start(ctx)
defer runner.Stop()
_ = runner.Emit(core.NewEvent("all", map[string]interface{}{"task": "process"}, nil))
```

### Conditional Workflow

Route to different agents based on input:

```go
type ConditionalWorkflow struct {
    router     core.Agent
    classifier core.Agent
    processor  core.Agent
    validator  core.Agent
}

func (w *ConditionalWorkflow) Process(ctx context.Context, input string) (string, error) {
    state := core.NewState()
    state.Set("input", input)
    
    // Step 1: Classify input type
    state.Set("query", "Classify this input type: text, code, data, or question")
    classification, err := w.classifier.Run(ctx, state)
    if err != nil {
        return "", err
    }
    
    inputType := classification.GetString("classification")
    
    // Step 2: Route to appropriate processor
    state.Set("input_type", inputType)
    switch inputType {
    case "code":
        state.Set("query", "Analyze and review this code")
        return w.processWithAgent(ctx, w.processor, state)
    case "question":
        state.Set("query", "Answer this question comprehensively")
        return w.processWithAgent(ctx, w.processor, state)
    default:
        state.Set("query", "Process this general input")
        return w.processWithAgent(ctx, w.processor, state)
    }
}

func (w *ConditionalWorkflow) processWithAgent(ctx context.Context, agent core.Agent, state core.State) (string, error) {
    result, err := agent.Run(ctx, state)
    if err != nil {
        return "", err
    }
    
    // Validate result
    state.Set("result_to_validate", result.GetResult())
    state.Set("query", "Validate this result for accuracy and completeness")
    validation, err := w.validator.Run(ctx, state)
    if err != nil {
        return result.GetResult(), nil // Return original if validation fails
    }
    
    if validation.GetBool("is_valid") {
        return result.GetResult(), nil
    }
    
    return "Result validation failed: " + validation.GetString("reason"), nil
}
```

## Tool Integration Examples

### Web Search Integration

Using MCP tools for web search requires registering an MCP transport plugin and initializing the MCP manager in your app startup. Then create MCP-aware agents via the public factory in core or internal plugin APIs.

### Database Integration

Connect to databases through MCP:

```go
func databaseExample() {
    agent, _ := core.NewMCPAgent("db-agent", llm)
    
    state := core.NewState()
    state.Set("query", "Query the users table and find all active users created in the last 30 days")
    
    // Agent will use database MCP tools automatically
    result, err := agent.Run(context.Background(), state)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Database Results:", result.GetResult())
}
```

### File Processing

Process files using MCP tools:

```go
func fileProcessingExample() {
    agent, _ := core.NewMCPAgent("file-processor", llm)
    
    state := core.NewState()
    state.Set("file_path", "/path/to/document.pdf")
    state.Set("query", "Extract key information from this PDF document and summarize it")
    
    result, err := agent.Run(context.Background(), state)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("File Summary:", result.GetResult())
}
```

## Production Examples

### Error Handling and Retry

Robust error handling for production systems:

```go
type ProductionAgent struct {
    agent      core.Agent
    retryCount int
    timeout    time.Duration
}

func (p *ProductionAgent) ProcessWithRetry(ctx context.Context, query string) (string, error) {
    state := core.NewState()
    state.Set("query", query)
    
    for attempt := 0; attempt < p.retryCount; attempt++ {
        // Set timeout context
        ctxWithTimeout, cancel := context.WithTimeout(ctx, p.timeout)
        defer cancel()
        
        result, err := p.agent.Run(ctxWithTimeout, state)
        if err == nil {
            return result.GetResult(), nil
        }
        
        // Log retry attempt
        log.Printf("Attempt %d failed: %v", attempt+1, err)
        
        // Exponential backoff
        time.Sleep(time.Duration(attempt+1) * time.Second)
    }
    
    return "", fmt.Errorf("failed after %d attempts", p.retryCount)
}
```

### Monitoring and Observability

Add comprehensive monitoring with callbacks and trace dumping:

```go
func monitoredWorkflow() {
    runner, _ := core.NewRunnerFromConfig("agentflow.toml")

    // Register callbacks for monitoring
    _ = runner.RegisterCallback(core.HookBeforeAgentRun, "logBefore", func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
        log.Printf("Starting agent run: %s", args.AgentID)
        return args.State, nil
    })
    _ = runner.RegisterCallback(core.HookAfterEventHandling, "logAfter", func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
        log.Printf("Event handled for agent=%s", args.AgentID)
        return args.State, nil
    })
    _ = runner.RegisterCallback(core.HookAgentError, "logErr", func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
        log.Printf("Agent error: %v", args.Error)
        return args.State, nil
    })

    // Start and emit an event
    ctx := context.Background()
    _ = runner.Start(ctx)
    defer runner.Stop()
    _ = runner.Emit(core.NewEvent("agent1", map[string]interface{}{"message": "hello"}, nil))

    // Get comprehensive traces
    traces, _ := runner.DumpTrace("session-123")
    for _, trace := range traces {
        fmt.Printf("Trace: %+v\n", trace)
    }
}
```

### Caching for Performance

Enable MCP caching through config and plugins. Prefer configuring cache in `agentflow.toml` or production config; avoid ad-hoc builders in code.

## Custom Provider Examples

### Mock Provider for Testing

Create mock providers for unit tests:

```go
type TestMockLLM struct {
    responses map[string]string
}

func (t *TestMockLLM) Generate(ctx context.Context, prompt string) (*core.LLMResponse, error) {
    if response, exists := t.responses[prompt]; exists {
        return &core.LLMResponse{
            Content: response,
            TokensUsed: 100,
        }, nil
    }
    
    return &core.LLMResponse{
        Content: "Mock response for: " + prompt,
        TokensUsed: 100,
    }, nil
}

func TestAgentBehavior(t *testing.T) {
    // Create mock with predefined responses
    mock := &TestMockLLM{
        responses: map[string]string{
            "What is 2+2?": "The answer is 4.",
            "Hello": "Hello! How can I help you?",
        },
    }
    
    agent, err := core.NewMCPAgent("test-agent", mock)
    require.NoError(t, err)
    
    state := core.NewState()
    state.Set("query", "What is 2+2?")
    
    result, err := agent.Run(context.Background(), state)
    require.NoError(t, err)
    assert.Equal(t, "The answer is 4.", result.GetResult())
}
```

### Custom LLM Provider

Implement your own LLM provider:

```go
type CustomLLMProvider struct {
    apiKey  string
    baseURL string
    client  *http.Client
}

func (c *CustomLLMProvider) Generate(ctx context.Context, prompt string) (*core.LLMResponse, error) {
    // Implement your custom LLM API call
    request := CustomLLMRequest{
        Prompt:      prompt,
        MaxTokens:   1000,
        Temperature: 0.7,
    }
    
    response, err := c.callAPI(ctx, request)
    if err != nil {
        return nil, err
    }
    
    return &core.LLMResponse{
        Content:    response.Text,
        TokensUsed: response.TokenCount,
        Model:      response.Model,
    }, nil
}

func (c *CustomLLMProvider) callAPI(ctx context.Context, request CustomLLMRequest) (*CustomLLMResponse, error) {
    // Implement HTTP call to your LLM service
    // This is where you'd integrate with your specific LLM API
    return nil, nil
}

// Register and use custom provider
func useCustomProvider() {
    provider := &CustomLLMProvider{
        apiKey:  "your-api-key",
        baseURL: "https://your-llm-service.com",
        client:  &http.Client{Timeout: 30 * time.Second},
    }
    
    agent, err := core.NewMCPAgent("custom-agent", provider)
    if err != nil {
        log.Fatal(err)
    }
    
    // Use agent normally
    state := core.NewState()
    state.Set("query", "Your query here")
    result, err := agent.Run(context.Background(), state)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Result:", result.GetResult())
}
```

## Best Practices

### 1. Agent Specialization

Create specialized agents for specific tasks:

```go
// Good: Specialized agents
researchAgent := createResearchAgent()    // Web search, data gathering
analysisAgent := createAnalysisAgent()    // Data processing, insights
summaryAgent := createSummaryAgent()      // Final presentation

// Avoid: Generic "do everything" agents
```

### 2. State Management

Use state effectively to pass data between agents:

```go
// Good: Clear state management
state := core.NewState()
state.Set("task_type", "research")
state.Set("topic", "AI trends")
state.Set("output_format", "summary")

// Avoid: Putting everything in the query
```

### 3. Error Handling

Always implement proper error handling:

```go
// Good: Comprehensive error handling
result, err := agent.Run(ctx, state)
if err != nil {
    // Log error
    log.Printf("Agent failed: %v", err)
    
    // Implement fallback
    return fallbackResponse, nil
}

// Avoid: Ignoring errors
```

### 4. Testing

Write tests for your agents:

```go
func TestResearchAgent(t *testing.T) {
    mock := &TestMockLLM{
        responses: map[string]string{
            "research prompt": "research results",
        },
    }
    
    agent, err := NewResearchAgent(mock)
    require.NoError(t, err)
    
    result, err := agent.Research(context.Background(), "test topic")
    require.NoError(t, err)
    assert.Contains(t, result, "research results")
}
```

## Next Steps

- Production - Deploy your agents reliably
- Error Handling - Advanced strategies and hooks
- Performance - Optimize throughput and latency
- Custom Tools - Build MCP tools
- API Reference - Runner and Agent

## Example Projects

Check out the `/examples` directory for complete working projects:

- **Simple Agent** - Basic single-agent application
- **Multi-Agent Workflow** - Research and analysis pipeline
- **Production App** - Enterprise-ready agent system
- **Custom Tools** - MCP server implementation
- **Performance Demo** - Optimized high-throughput system

Each example includes:
- Complete source code
- Configuration files
- Documentation
- Test cases
- Deployment instructions
