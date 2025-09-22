# Customization Guide

This guide provides comprehensive instructions for customizing your researchAgent1 multi-agent system to meet your specific requirements.

## Quick Start Customization

### 1. Modify Agent Behavior

The fastest way to customize your system is to modify the agent implementations in the `agents/` directory:

```go
// In agents/agent1.go
func (a *Agent1Handler) Run(ctx context.Context, event agenticgokit.Event, state agenticgokit.State) (agenticgokit.AgentResult, error) {
    // TODO: Replace this with your custom logic
    inputToProcess := extractInput(event, state)
    
    // Add your business logic here
    result := processWithCustomLogic(inputToProcess)
    
    // Return formatted result
    return formatResult(result), nil
}
```

### 2. Add Custom Dependencies

Extend agent constructors to accept additional dependencies:

```go
type Agent1Handler struct {
    llm        agenticgokit.ModelProvider
    database   *sql.DB           // Add database connection
    apiClient  *http.Client      // Add API client
    config     *YourConfig       // Add custom configuration
}

func NewAgent1(llmProvider agenticgokit.ModelProvider, db *sql.DB, config *YourConfig) *Agent1Handler {
    return &Agent1Handler{
        llm:       llmProvider,
        database:  db,
        config:    config,
    }
}
```

### 3. Customize Main Application

Modify `main.go` to initialize your custom dependencies:

```go
func main() {
    // ... existing initialization ...
    
    // Add your custom initialization
    db, err := sql.Open("postgres", "your-connection-string")
    if err != nil {
        log.Fatal(err)
    }
    
    config := loadYourConfig()
    
    // Create agents with custom dependencies
    agent1 := agents.NewAgent1(llmProvider, db, config)
    agents["agent1"] = agent1
    
    // ... rest of main function ...
}
```

## 🏗️ Architecture Customization

### Orchestration Modes

Your system supports different orchestration patterns. Modify `agentflow.toml` to change the execution flow:

#### Sequential Processing
```toml
[orchestration]
mode = "sequential"
agents = ["agent1", "agent2", "agent3"]
```

#### Collaborative Processing  
```toml
[orchestration]
mode = "collaborative"
agents = ["agent1", "agent2", "agent3"]
max_concurrency = 3
```

#### Loop Processing
```toml
[orchestration]
mode = "loop"
agent = "agent1"
max_iterations = 5
```

#### Mixed Processing
```toml
[orchestration]
mode = "mixed"
collaborative_agents = ["agent1", "agent2"]
sequential_agents = ["agent3", "agent4"]
```

### Custom Routing Logic

Implement custom routing by modifying the orchestrator configuration:

```go
// In main.go, after creating the runner
runner.SetCustomRouter(func(ctx context.Context, event core.Event, state core.State) string {
    // Implement your custom routing logic
    if strings.Contains(event.GetData()["message"].(string), "urgent") {
        return "priority_agent"
    }
    return "standard_agent"
})
```

## Architecture Customization

### LLM Provider Configuration

#### OpenAI Configuration
```toml
[agent_flow]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 2000

[openai]
api_key_env = "OPENAI_API_KEY"
organization = "your-org-id"
```

#### Azure OpenAI Configuration
```toml
[agent_flow]
provider = "azure"

[azure]
api_key_env = "AZURE_OPENAI_API_KEY"
endpoint = "https://your-resource.openai.azure.com/"
deployment = "your-deployment-name"
api_version = "2023-12-01-preview"
```

#### Ollama Configuration
```toml
[agent_flow]
provider = "ollama"

[ollama]
base_url = "http://localhost:11434"
model = "llama2"
```

### Memory System Customization


Memory system is not currently enabled. To add memory capabilities:

1. **Enable in configuration**:
```toml
[agent_memory]
enabled = true
provider = "memory"  # or "pgvector", "weaviate"
```

2. **Update agent constructors** to accept memory parameter
3. **Modify main.go** to initialize memory system


### Tool Integration (MCP)


#### Adding New Tools

1. **Configure MCP server** in `agentflow.toml`:
```toml
[[mcp.servers]]
name = "filesystem"
command = "npx"
args = ["@modelcontextprotocol/server-filesystem", "/path/to/allowed/files"]
enabled = true
```

2. **Use tools in agents**:
```go
func (a *YourAgentHandler) useCustomTool(ctx context.Context, query string) (string, error) {
    args := map[string]interface{}{
        "path": "/path/to/file",
        "content": query,
    }
    
    result, err := agenticgokit.ExecuteMCPTool(ctx, "write_file", args)
    if err != nil {
        return "", err
    }
    
    return result.Content[0].Text, nil
}
```

#### Custom Tool Validation
```go
func (a *YourAgentHandler) validateToolCall(toolName string, args map[string]interface{}) error {
    // Add your validation logic
    switch toolName {
    case "sensitive_operation":
        if !a.hasPermission(args["operation"]) {
            return fmt.Errorf("insufficient permissions")
        }
    }
    return nil
}
```


## 🎨 User Interface Customization

### Command Line Interface

#### Adding Custom Flags
```go
// In main.go
func main() {
    messageFlag := flag.String("m", "", "Message to process")
    debugFlag := flag.Bool("debug", false, "Enable debug mode")
    outputFlag := flag.String("output", "", "Output file path")
    configFlag := flag.String("config", "agentflow.toml", "Configuration file")
    
    flag.Parse()
    
    if *debugFlag {
        core.SetLogLevel(core.DEBUG)
    }
    
    // Use custom flags in your logic
}
```

#### Custom Input Processing
```go
func processCustomInput() string {
    if *inputFileFlag != "" {
        content, err := os.ReadFile(*inputFileFlag)
        if err != nil {
            log.Fatal(err)
        }
        return string(content)
    }
    
    // Interactive multi-line input
    fmt.Println("Enter your message (Ctrl+D to finish):")
    scanner := bufio.NewScanner(os.Stdin)
    var lines []string
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    
    return strings.Join(lines, "\n")
}
```

### Output Formatting

#### JSON Output
```go
type OutputResult struct {
    AgentName string    `json:"agent_name"`
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
    Duration  string    `json:"duration"`
}

func formatAsJSON(results []AgentOutput) {
    var jsonResults []OutputResult
    for _, result := range results {
        jsonResults = append(jsonResults, OutputResult{
            AgentName: result.AgentName,
            Content:   result.Content,
            Timestamp: result.Timestamp,
            Duration:  "1.2s", // Calculate actual duration
        })
    }
    
    output, _ := json.MarshalIndent(jsonResults, "", "  ")
    fmt.Println(string(output))
}
```

#### File Output
```go
func saveResultsToFile(results []AgentOutput, filename string) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    for _, result := range results {
        fmt.Fprintf(file, "=== %s ===\n", result.AgentName)
        fmt.Fprintf(file, "%s\n\n", result.Content)
    }
    
    return nil
}
```

## Security Customization

### Input Validation
```go
func validateInput(input string) error {
    if len(input) > 10000 {
        return fmt.Errorf("input too long")
    }
    
    // Check for malicious patterns
    maliciousPatterns := []string{"<script>", "javascript:", "data:"}
    for _, pattern := range maliciousPatterns {
        if strings.Contains(strings.ToLower(input), pattern) {
            return fmt.Errorf("potentially malicious input detected")
        }
    }
    
    return nil
}
```

### Rate Limiting
```go
import "golang.org/x/time/rate"

type RateLimitedAgent struct {
    *Agent1Handler
    limiter *rate.Limiter
}

func (a *RateLimitedAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    if !a.limiter.Allow() {
        return core.AgentResult{}, fmt.Errorf("rate limit exceeded")
    }
    
    return a.Agent1Handler.Run(ctx, event, state)
}
```

### Authentication
```go
func authenticateRequest(ctx context.Context, token string) error {
    // Implement your authentication logic
    if token == "" {
        return fmt.Errorf("authentication token required")
    }
    
    // Validate token with your auth service
    valid, err := validateToken(token)
    if err != nil {
        return fmt.Errorf("authentication failed: %w", err)
    }
    
    if !valid {
        return fmt.Errorf("invalid authentication token")
    }
    
    return nil
}
```

## Monitoring and Observability

### Custom Metrics
```go
import "github.com/prometheus/client_golang/prometheus"

var (
    agentExecutionDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "agent_execution_duration_seconds",
            Help: "Duration of agent execution",
        },
        []string{"agent_name"},
    )
)

func (a *YourAgentHandler) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    start := time.Now()
    defer func() {
        agentExecutionDuration.WithLabelValues(a.agentName).Observe(time.Since(start).Seconds())
    }()
    
    // Your agent logic here
    return core.AgentResult{}, nil
}
```

### Structured Logging
```go
import "github.com/rs/zerolog/log"

func (a *YourAgentHandler) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    logger := log.With().
        Str("agent", a.agentName).
        Str("event_id", event.GetID()).
        Logger()
    
    logger.Debug().Msg("Agent execution started")
    
    // Your logic here
    
    logger.Debug().
        Dur("duration", time.Since(start)).
        Msg("Agent execution completed")
    
    return result, nil
}
```

### Health Checks
```go
func (a *YourAgentHandler) HealthCheck(ctx context.Context) error {
    // Check LLM provider connectivity
    if err := a.llm.HealthCheck(ctx); err != nil {
        return fmt.Errorf("LLM provider unhealthy: %w", err)
    }
    
    // Check database connectivity
    if a.database != nil {
        if err := a.database.PingContext(ctx); err != nil {
            return fmt.Errorf("database unhealthy: %w", err)
        }
    }
    
    return nil
}
```

## Testing Customization

### Unit Testing Agents
```go
func TestAgent1Handler_Run(t *testing.T) {
    // Create mock dependencies
    mockLLM := &MockLLMProvider{}
    mockMemory := &MockMemory{}
    
    agent := NewAgent1(mockLLM, mockMemory)
    
    // Create test event
    event := core.NewEvent("test", core.EventData{
        "message": "test input",
    }, nil)
    
    // Execute agent
    result, err := agent.Run(context.Background(), event, core.NewState())
    
    // Assert results
    assert.NoError(t, err)
    assert.NotNil(t, result.OutputState)
}
```

### Integration Testing
```go
func TestWorkflowIntegration(t *testing.T) {
    // Set up test environment
    config := scaffold.ProjectConfig{
        Name:      "test-workflow",
        NumAgents: 2,
        Provider:  "mock",
    }
    
    // Create test workflow
    runner, err := core.NewRunnerFromConfig("test-agentflow.toml")
    require.NoError(t, err)
    
    // Register test agents
    agent1 := &TestAgent1{}
    agent2 := &TestAgent2{}
    
    runner.RegisterAgent("agent1", agent1)
    runner.RegisterAgent("agent2", agent2)
    
    // Execute workflow
    runner.Start(context.Background())
    
    event := core.NewEvent("agent1", core.EventData{
        "message": "test message",
    }, map[string]string{"route": "agent1"})
    
    err = runner.Emit(event)
    require.NoError(t, err)
    
    // Wait for completion and assert results
    runner.Stop()
    
    // Assert workflow completed successfully
}
```

## Deployment Customization

### Docker Configuration
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/agentflow.toml .

CMD ["./main"]
```

### Environment Configuration
```bash
# .env file
OPENAI_API_KEY=your-api-key
DATABASE_URL=postgres://user:pass@localhost/db
LOG_LEVEL=info
ENVIRONMENT=production
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: researchAgent1
spec:
  replicas: 3
  selector:
    matchLabels:
      app: researchAgent1
  template:
    metadata:
      labels:
        app: researchAgent1
    spec:
      containers:
      - name: researchAgent1
        image: your-registry/researchAgent1:latest
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: api-keys
              key: openai-key
        ports:
        - containerPort: 8080
```

## Advanced Patterns

### Plugin Architecture
```go
type Plugin interface {
    Name() string
    Initialize(config map[string]interface{}) error
    Process(ctx context.Context, input interface{}) (interface{}, error)
}

type PluginManager struct {
    plugins map[string]Plugin
}

func (pm *PluginManager) RegisterPlugin(plugin Plugin) {
    pm.plugins[plugin.Name()] = plugin
}

func (pm *PluginManager) ExecutePlugin(name string, ctx context.Context, input interface{}) (interface{}, error) {
    plugin, exists := pm.plugins[name]
    if !exists {
        return nil, fmt.Errorf("plugin %s not found", name)
    }
    
    return plugin.Process(ctx, input)
}
```

### Event Sourcing
```go
type Event struct {
    ID        string
    Type      string
    Data      interface{}
    Timestamp time.Time
    AgentID   string
}

type EventStore interface {
    Store(ctx context.Context, event Event) error
    GetEvents(ctx context.Context, agentID string) ([]Event, error)
}

func (a *YourAgentHandler) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Store input event
    inputEvent := Event{
        ID:        uuid.New().String(),
        Type:      "agent_input",
        Data:      event.GetData(),
        Timestamp: time.Now(),
        AgentID:   a.agentName,
    }
    a.eventStore.Store(ctx, inputEvent)
    
    // Process and store output event
    result, err := a.processInput(ctx, event, state)
    
    outputEvent := Event{
        ID:        uuid.New().String(),
        Type:      "agent_output",
        Data:      result,
        Timestamp: time.Now(),
        AgentID:   a.agentName,
    }
    a.eventStore.Store(ctx, outputEvent)
    
    return result, err
}
```

## Troubleshooting

### Common Issues and Solutions

1. **Agent not receiving expected input**
   - Check orchestration mode and agent order
   - Verify state key names match between agents
   - Review routing configuration

2. **Memory system connection issues**
   - Verify database is running and accessible
   - Check connection string format
   - Validate embedding model configuration

3. **Tool integration failures**
   - Ensure MCP servers are running
   - Check tool permissions and arguments
   - Review MCP server logs

4. **Performance issues**
   - Profile agent execution times
   - Check for blocking operations
   - Consider implementing caching

### Debug Mode

Enable detailed logging to troubleshoot issues:

```go
// In main.go
core.SetLogLevel(core.DEBUG)

// Or via environment variable
os.Setenv("LOG_LEVEL", "debug")
```

---

This guide covers the most common customization scenarios. For more advanced use cases, refer to the [AgenticGoKit documentation](https://github.com/kunalkushwaha/agenticgokit) or create an issue for support.
