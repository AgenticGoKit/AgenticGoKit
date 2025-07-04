# AgentFlow Examples

This guide provides practical examples of building AI agents and workflows with AgentFlow, from simple single-agent applications to complex multi-agent orchestrations.

## Table of Contents

- [Quick Start Examples](#quick-start-examples)
- [Single Agent Examples](#single-agent-examples)
- [Multi-Agent Workflows](#multi-agent-workflows)
- [Tool Integration Examples](#tool-integration-examples)
- [Production Examples](#production-examples)
- [Custom Provider Examples](#custom-provider-examples)

## Quick Start Examples

### Simple Query Agent (5 minutes)

The fastest way to create an agent that can answer questions:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // Initialize MCP for tool discovery
    core.QuickStartMCP()
    
    // Create LLM provider (using mock for this example)
    llm := &core.MockLLM{}
    
    // Create an agent
    agent, err := core.NewMCPAgent("helper", llm)
    if err != nil {
        log.Fatal(err)
    }
    
    // Create state with query
    state := core.NewState()
    state.Set("query", "What is the capital of France?")
    
    // Run agent
    result, err := agent.Run(context.Background(), state)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Response:", result.GetResult())
}
```

### Multi-Agent Orchestration (Quick Start)

Generate complete multi-agent workflows with the CLI:

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

An agent that searches for information and provides summaries:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kunalkushwaha/agentflow/core"
)

type ResearchAgent struct {
    agent core.Agent
}

func NewResearchAgent(llm core.ModelProvider) (*ResearchAgent, error) {
    agent, err := core.NewMCPAgent("researcher", llm)
    if err != nil {
        return nil, err
    }
    
    return &ResearchAgent{agent: agent}, nil
}

func (r *ResearchAgent) Research(ctx context.Context, topic string) (string, error) {
    state := core.NewState()
    state.Set("query", fmt.Sprintf("Research the topic: %s. Provide a comprehensive summary with key points.", topic))
    
    result, err := r.agent.Run(ctx, state)
    if err != nil {
        return "", err
    }
    
    return result.GetResult(), nil
}

func main() {
    // Initialize with real LLM provider
    config := core.LLMConfig{
        Provider: "azure-openai",
        APIKey:   "your-api-key",
        BaseURL:  "https://your-resource.openai.azure.com",
    }
    
    llm := core.NewAzureOpenAIAdapter(config)
    
    // Create research agent
    researcher, err := NewResearchAgent(llm)
    if err != nil {
        log.Fatal(err)
    }
    
    // Conduct research
    summary, err := researcher.Research(context.Background(), "quantum computing applications")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Research Summary:", summary)
}
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

All agents work in parallel to process the same task:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // Create specialized agents
    agents := map[string]core.AgentHandler{
        "researcher": NewResearchAgent(),
        "analyzer":   NewAnalysisAgent(),
        "validator":  NewValidationAgent(),
    }
    
    // Create collaborative orchestration
    runner := core.NewOrchestrationBuilder(core.OrchestrationCollaborate).
        WithAgents(agents).
        WithTimeout(2 * time.Minute).
        WithFailureThreshold(0.8).
        WithMaxConcurrency(10).
        Build()
    
    // Create event
    event := core.NewEvent("all", map[string]interface{}{
        "task": "research AI trends and provide comprehensive analysis",
    }, nil)
    
    // All agents process in parallel
    result, err := runner.Run(context.Background(), event)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Collaborative Result: %s\n", result.GetResult())
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
    "time"
    
    "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // Create pipeline agents
    agents := map[string]core.AgentHandler{
        "collector":  NewDataCollectorAgent(),
        "processor":  NewDataProcessorAgent(),
        "formatter":  NewDataFormatterAgent(),
    }
    
    // Create sequential orchestration
    runner := core.NewOrchestrationBuilder(core.OrchestrationSequential).
        WithAgents(agents).
        WithTimeout(5 * time.Minute).
        Build()
    
    // Create pipeline event
    event := core.NewEvent("pipeline", map[string]interface{}{
        "data_source": "market_data",
        "format":      "json",
    }, nil)
    
    // Process through pipeline
    result, err := runner.Run(context.Background(), event)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Pipeline Result: %s\n", result.GetResult())
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
    "time"
    
    "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // Create quality checker agent
    agents := map[string]core.AgentHandler{
        "quality-checker": NewQualityCheckerAgent(),
    }
    
    // Create loop orchestration
    runner := core.NewOrchestrationBuilder(core.OrchestrationLoop).
        WithAgents(agents).
        WithMaxIterations(10).
        WithTimeout(10 * time.Minute).
        Build()
    
    // Create quality check event
    event := core.NewEvent("loop", map[string]interface{}{
        "content":          "document to check",
        "quality_threshold": 0.95,
    }, nil)
    
    // Loop until quality is met
    result, err := runner.Run(context.Background(), event)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Quality Check Result: %s\n", result.GetResult())
}
```

### Mixed Orchestration Workflow

Combine collaborative and sequential patterns:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // Collaborative agents (parallel processing)
    collaborativeAgents := map[string]core.AgentHandler{
        "analyzer":  NewAnalyzerAgent(),
        "validator": NewValidatorAgent(),
    }
    
    // Sequential agents (pipeline processing)
    sequentialAgents := map[string]core.AgentHandler{
        "processor": NewProcessorAgent(),
        "reporter":  NewReporterAgent(),
    }
    
    // Create mixed orchestration
    runner := core.NewOrchestrationBuilder(core.OrchestrationMixed).
        WithCollaborativeAgents(collaborativeAgents).
        WithSequentialAgents(sequentialAgents).
        WithTimeout(8 * time.Minute).
        WithFailureThreshold(0.8).
        Build()
    
    // Create mixed workflow event
    event := core.NewEvent("mixed", map[string]interface{}{
        "task": "analyze data, validate results, process, and generate report",
    }, nil)
    
    // Execute mixed workflow
    result, err := runner.Run(context.Background(), event)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Mixed Workflow Result: %s\n", result.GetResult())
}
```

### Workflow Visualization

Generate Mermaid diagrams for any orchestration:

```go
package main

import (
    "fmt"
    "os"
    "path/filepath"
    
    "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // Create agents
    agents := map[string]core.AgentHandler{
        "researcher": NewResearchAgent(),
        "analyzer":   NewAnalysisAgent(),
        "validator":  NewValidationAgent(),
    }
    
    // Create orchestration builder
    builder := core.NewOrchestrationBuilder(core.OrchestrationCollaborate).
        WithAgents(agents).
        WithTimeout(2 * time.Minute).
        WithFailureThreshold(0.8)
    
    // Generate Mermaid diagram
    diagram := builder.GenerateMermaidDiagram()
    
    // Save to file
    outputDir := "docs/diagrams"
    os.MkdirAll(outputDir, 0755)
    
    filename := filepath.Join(outputDir, "workflow.mmd")
    err := os.WriteFile(filename, []byte(diagram), 0644)
    if err != nil {
        fmt.Printf("Error saving diagram: %v\n", err)
        return
    }
    
    fmt.Printf("Workflow diagram saved to: %s\n", filename)
    fmt.Println("Diagram content:")
    fmt.Println(diagram)
}
```

### Research and Analysis Pipeline

A workflow where one agent researches and another analyzes:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // Initialize MCP
    core.QuickStartMCP()
    
    // Create LLM provider
    llm := core.NewAzureOpenAIAdapter(core.LLMConfig{
        Provider: "azure-openai",
        APIKey:   "your-api-key",
        BaseURL:  "https://your-resource.openai.azure.com",
    })
    
    // Create agents
    researcher, _ := core.NewMCPAgent("researcher", llm)
    analyst, _ := core.NewMCPAgent("analyst", llm)
    
    // Create workflow runner
    runner := core.NewSequentialRunner()
    runner.AddAgent(researcher)
    runner.AddAgent(analyst)
    
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

Process multiple tasks concurrently:

```go
func parallelWorkflow() {
    // Create parallel runner
    runner := core.NewParallelRunner()
    
    // Add multiple agents for different tasks
    webAgent, _ := core.NewMCPAgent("web-searcher", llm)
    dbAgent, _ := core.NewMCPAgent("database-query", llm)
    fileAgent, _ := core.NewMCPAgent("file-processor", llm)
    
    runner.AddAgent(webAgent)
    runner.AddAgent(dbAgent)
    runner.AddAgent(fileAgent)
    
    // Create states for each agent
    states := []core.State{
        core.NewStateWithQuery("search web for latest news"),
        core.NewStateWithQuery("query database for user statistics"),
        core.NewStateWithQuery("process uploaded files"),
    }
    
    // Run all agents in parallel
    results, err := runner.RunParallel(context.Background(), states)
    if err != nil {
        log.Fatal(err)
    }
    
    // Process results
    for i, result := range results {
        fmt.Printf("Agent %d result: %s\n", i+1, result.GetResult())
    }
}
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

Using MCP tools for web search:

```go
func webSearchExample() {
    // Initialize MCP with web search tools
    core.QuickStartMCP()
    
    agent, _ := core.NewMCPAgent("web-searcher", llm)
    
    state := core.NewState()
    state.Set("query", "Search for the latest Docker tutorials and summarize the top 3 results")
    
    result, err := agent.Run(context.Background(), state)
    if err != nil {
        log.Fatal(err)
    }
    
    // The agent automatically uses web search tools
    fmt.Println("Search Results:", result.GetResult())
}
```

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

Add comprehensive monitoring:

```go
func monitoredWorkflow() {
    // Create runner with monitoring
    runner := core.NewSequentialRunner()
    
    // Register callbacks for monitoring
    runner.RegisterCallback(core.HookBeforeAgentRun, func(ctx context.Context, state core.State) {
        log.Printf("Starting agent run: %s", state.Get("agent_name"))
    })
    
    runner.RegisterCallback(core.HookAfterAgentRun, func(ctx context.Context, result core.Result) {
        metrics := result.GetMetrics()
        log.Printf("Agent completed: duration=%v, tokens=%d", 
            metrics.Duration, metrics.TokensUsed)
    })
    
    runner.RegisterCallback(core.HookOnError, func(ctx context.Context, err error) {
        log.Printf("Agent error: %v", err)
        // Send to monitoring system
    })
    
    // Run workflow with monitoring
    result, err := runner.Run(context.Background(), state)
    if err != nil {
        log.Fatal(err)
    }
    
    // Get comprehensive traces
    traces, _ := runner.DumpTrace("session-123")
    for _, trace := range traces {
        fmt.Printf("Trace: %+v\n", trace)
    }
}
```

### Caching for Performance

Implement caching for better performance:

```go
func cachedAgent() {
    // Create agent with caching enabled
    config := core.AgentConfig{
        CacheEnabled:    true,
        CacheTTL:        time.Hour,
        CacheProvider:   "redis",
    }
    
    agent := core.NewAgentBuilder("cached-agent").
        WithLLM(llm).
        WithConfig(config).
        WithMCP().
        WithCache().
        Build()
    
    // First call - will hit LLM
    start := time.Now()
    result1, _ := agent.Run(context.Background(), state)
    duration1 := time.Since(start)
    fmt.Printf("First call: %v\n", duration1)
    
    // Second call - will hit cache
    start = time.Now()
    result2, _ := agent.Run(context.Background(), state)
    duration2 := time.Since(start)
    fmt.Printf("Cached call: %v\n", duration2)
    
    // Results should be identical, but second call much faster
    fmt.Printf("Results match: %v\n", result1.GetResult() == result2.GetResult())
}
```

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

- **[Production Guide](Production.md)** - Deploy your agents to production
- **[Error Handling](ErrorHandling.md)** - Advanced error handling strategies
- **[Performance Guide](Performance.md)** - Optimize your agent performance
- **[Custom Tools](CustomTools.md)** - Build your own MCP tools
- **[API Reference](../api/core.md)** - Complete API documentation

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
