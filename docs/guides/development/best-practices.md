# Best Practices

**Development best practices for building robust AgenticGoKit applications**

This guide covers essential best practices for developing, deploying, and maintaining AgenticGoKit applications. Follow these guidelines to build reliable, scalable, and maintainable multi-agent systems.

## Agent Design Principles

### Single Responsibility Principle

Each agent should have one clear, well-defined purpose:

```go
// Good: Focused agent with single responsibility
type EmailValidatorAgent struct {
    name string
}

func (a *EmailValidatorAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    email := event.Data["email"].(string)
    
    if !isValidEmail(email) {
        return core.AgentResult{
            Data: map[string]interface{}{
                "valid": false,
                "error": "Invalid email format",
            },
        }, nil
    }
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "valid": true,
            "email": email,
        },
    }, nil
}

// Bad: Agent trying to do too many things
type EmailProcessorAgent struct {
    // Validates, sends, logs, and analyzes emails - too many responsibilities
}
```

### Stateless Design When Possible

Prefer stateless agents for better scalability and testability:

```go
// Good: Stateless agent
type TextAnalyzerAgent struct {
    llmProvider core.ModelProvider
}

func (a *TextAnalyzerAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    text := event.Data["text"].(string)
    
    // All state comes from event and state parameters
    analysis, err := a.analyzeText(ctx, text)
    if err != nil {
        return core.AgentResult{}, err
    }
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "analysis": analysis,
        },
    }, nil
}

// Use stateful agents only when necessary
type ConversationAgent struct {
    llmProvider core.ModelProvider
    memory      core.Memory  // Stateful for conversation history
}
```

### Robust Error Handling

Always handle errors gracefully and provide meaningful feedback:

```go
func (a *MyAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Validate inputs
    input, ok := event.Data["input"].(string)
    if !ok {
        return core.AgentResult{}, fmt.Errorf("invalid input type: expected string, got %T", event.Data["input"])
    }
    
    if len(input) == 0 {
        return core.AgentResult{}, fmt.Errorf("input cannot be empty")
    }
    
    // Process with timeout
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    result, err := a.processInput(ctx, input)
    if err != nil {
        // Wrap errors with context
        return core.AgentResult{}, fmt.Errorf("processing input %q failed: %w", input, err)
    }
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "result": result,
        },
    }, nil
}
```

## Configuration Management

### Environment-Based Configuration

Use environment variables for deployment-specific settings:

```toml
# agentflow.toml
[llm]
provider = "${LLM_PROVIDER:azure}"
api_key = "${LLM_API_KEY}"
endpoint = "${LLM_ENDPOINT}"

[memory]
provider = "${MEMORY_PROVIDER:memory}"
connection = "${MEMORY_CONNECTION:memory}"

[logging]
level = "${LOG_LEVEL:info}"
format = "${LOG_FORMAT:json}"
```

```go
// Load configuration with validation
func loadConfig() (*Config, error) {
    config, err := core.LoadConfigFromWorkingDir()
    if err != nil {
        return nil, fmt.Errorf("failed to load configuration: %w", err)
    }
    
    // Validate required settings
    if config.LLM.APIKey == "" {
        return nil, fmt.Errorf("LLM API key is required")
    }
    
    return config, nil
}
```

### Configuration Profiles

Use different configurations for different environments:

```bash
# Development
export AGENTFLOW_PROFILE=development
export LLM_PROVIDER=ollama
export MEMORY_PROVIDER=memory

# Production
export AGENTFLOW_PROFILE=production
export LLM_PROVIDER=azure
export MEMORY_PROVIDER=pgvector
export MEMORY_CONNECTION="postgres://..."
```

## Performance Optimization

### Connection Pooling

Configure appropriate connection pools for external services:

```toml
[memory]
provider = "pgvector"
connection = "postgres://user:pass@localhost:5432/db"
max_connections = 25
idle_connections = 5
connection_lifetime = "1h"

[llm]
provider = "azure"
max_connections = 10
request_timeout = "30s"
```

### Caching Strategies

Implement caching for expensive operations:

```go
type CachedAgent struct {
    llmProvider core.ModelProvider
    cache       map[string]string
    cacheMutex  sync.RWMutex
    cacheExpiry time.Duration
}

func (a *CachedAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    input := event.Data["input"].(string)
    
    // Check cache first
    a.cacheMutex.RLock()
    if cached, exists := a.cache[input]; exists {
        a.cacheMutex.RUnlock()
        return core.AgentResult{
            Data: map[string]interface{}{
                "result": cached,
                "cached": true,
            },
        }, nil
    }
    a.cacheMutex.RUnlock()
    
    // Process and cache result
    result, err := a.llmProvider.Generate(ctx, input)
    if err != nil {
        return core.AgentResult{}, err
    }
    
    a.cacheMutex.Lock()
    a.cache[input] = result
    a.cacheMutex.Unlock()
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "result": result,
            "cached": false,
        },
    }, nil
}
```

### Resource Management

Always clean up resources properly:

```go
func (a *MyAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Use context with timeout
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // Clean up resources
    resource, err := acquireResource()
    if err != nil {
        return core.AgentResult{}, err
    }
    defer resource.Close()
    
    // Process with resource
    result, err := processWithResource(ctx, resource, event.Data)
    if err != nil {
        return core.AgentResult{}, err
    }
    
    return core.AgentResult{Data: result}, nil
}
```

## Security Best Practices

### API Key Management

Never hardcode API keys or sensitive information:

```go
// Good: Use environment variables
func createLLMProvider() (core.ModelProvider, error) {
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
    }
    
    return core.NewOpenAIProvider(core.OpenAIConfig{
        APIKey: apiKey,
        Model:  "gpt-4",
    })
}

// Bad: Hardcoded secrets
func createLLMProviderBad() (core.ModelProvider, error) {
    return core.NewOpenAIProvider(core.OpenAIConfig{
        APIKey: "sk-hardcoded-key-here", // Never do this!
        Model:  "gpt-4",
    })
}
```

### Input Validation

Always validate and sanitize inputs:

```go
func (a *MyAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Validate input types
    userInput, ok := event.Data["user_input"].(string)
    if !ok {
        return core.AgentResult{}, fmt.Errorf("user_input must be a string")
    }
    
    // Sanitize input
    userInput = strings.TrimSpace(userInput)
    if len(userInput) == 0 {
        return core.AgentResult{}, fmt.Errorf("user_input cannot be empty")
    }
    
    // Limit input size
    if len(userInput) > 10000 {
        return core.AgentResult{}, fmt.Errorf("user_input too long (max 10000 characters)")
    }
    
    // Remove potentially dangerous content
    userInput = sanitizeInput(userInput)
    
    // Process sanitized input
    return a.processInput(ctx, userInput)
}

func sanitizeInput(input string) string {
    // Remove or escape potentially dangerous content
    // This is a simplified example
    input = strings.ReplaceAll(input, "<script>", "")
    input = strings.ReplaceAll(input, "</script>", "")
    return input
}
```

## Testing Best Practices

### Comprehensive Test Coverage

Write tests at multiple levels:

```go
// Unit tests for individual agents
func TestAgentLogic(t *testing.T) {
    agent := NewMyAgent()
    
    tests := []struct {
        name     string
        input    map[string]interface{}
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    map[string]interface{}{"data": "test"},
            expected: "processed: test",
            wantErr:  false,
        },
        {
            name:    "invalid input",
            input:   map[string]interface{}{"data": 123},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            event := core.NewEvent("test", tt.input)
            result, err := agent.Run(context.Background(), event, core.NewState())
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            require.NoError(t, err)
            assert.Equal(t, tt.expected, result.Data["result"])
        })
    }
}

// Integration tests for agent interactions
func TestAgentWorkflow(t *testing.T) {
    agents := map[string]core.AgentHandler{
        "step1": NewStep1Agent(),
        "step2": NewStep2Agent(),
    }
    
    runner := core.CreateSequentialRunner(agents, []string{"step1", "step2"}, 30*time.Second)
    
    event := core.NewEvent("process", map[string]interface{}{
        "input": "test data",
    })
    
    _ = runner.Start(context.Background())
    defer runner.Stop()
    err := runner.Emit(event)
    
    require.NoError(t, err)
    assert.Len(t, results, 2)
    // Verify workflow results...
}
```

### Use Mocks for External Dependencies

Mock external services for reliable testing:

```go
type MockLLMProvider struct {
    responses map[string]string
    errors    map[string]error
}

func (m *MockLLMProvider) Generate(ctx context.Context, prompt string) (string, error) {
    if err, exists := m.errors[prompt]; exists {
        return "", err
    }
    
    if response, exists := m.responses[prompt]; exists {
        return response, nil
    }
    
    return "default mock response", nil
}

func TestAgentWithMockLLM(t *testing.T) {
    mockLLM := &MockLLMProvider{
        responses: map[string]string{
            "test prompt": "expected response",
        },
    }
    
    agent := NewMyAgent(mockLLM)
    
    // Test with predictable mock responses
    event := core.NewEvent("test", map[string]interface{}{
        "prompt": "test prompt",
    })
    
    result, err := agent.Run(context.Background(), event, core.NewState())
    
    require.NoError(t, err)
    assert.Equal(t, "expected response", result.Data["response"])
}
```

## Monitoring and Observability

### Structured Logging

Use structured logging for better observability:

```go
import (
    "github.com/sirupsen/logrus"
)

func (a *MyAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    logger := logrus.WithFields(logrus.Fields{
        "agent":      a.name,
        "event_type": event.Type,
        "trace_id":   getTraceID(ctx),
    })
    
    logger.Info("Agent execution started")
    start := time.Now()
    
    // Invoke agent normally via the runner orchestration
    _ = runner.Start(ctx)
    defer runner.Stop()
    err := runner.Emit(event)
    
    duration := time.Since(start)
    
    if err != nil {
        logger.WithFields(logrus.Fields{
            "error":    err.Error(),
            "duration": duration,
        }).Error("Agent execution failed")
        return core.AgentResult{}, err
    }
    
    logger.WithFields(logrus.Fields{
        "duration": duration,
        "success":  true,
    }).Info("Agent execution completed")
    
    return result, nil
}
```

### Health Checks

Implement health checks for monitoring:

```go
type HealthChecker interface {
    HealthCheck(ctx context.Context) error
}

func (a *MyAgent) HealthCheck(ctx context.Context) error {
    // Check LLM provider connectivity
    if a.llmProvider != nil {
        _, err := a.llmProvider.Generate(ctx, "health check")
        if err != nil {
            return fmt.Errorf("LLM provider health check failed: %w", err)
        }
    }
    
    // Check memory system
    if a.memory != nil {
        err := a.memory.Store(ctx, "health check", "test")
        if err != nil {
            return fmt.Errorf("memory system health check failed: %w", err)
        }
    }
    
    return nil
}

// HTTP health endpoint
func setupHealthEndpoint(agents map[string]core.AgentHandler) {
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()
        
        health := map[string]interface{}{
            "status":    "healthy",
            "timestamp": time.Now(),
            "agents":    make(map[string]string),
        }
        
        for name, agent := range agents {
            if checker, ok := agent.(HealthChecker); ok {
                if err := checker.HealthCheck(ctx); err != nil {
                    health["agents"].(map[string]string)[name] = "unhealthy: " + err.Error()
                    health["status"] = "degraded"
                } else {
                    health["agents"].(map[string]string)[name] = "healthy"
                }
            } else {
                health["agents"].(map[string]string)[name] = "no health check"
            }
        }
        
        w.Header().Set("Content-Type", "application/json")
        if health["status"] != "healthy" {
            w.WriteHeader(http.StatusServiceUnavailable)
        }
        json.NewEncoder(w).Encode(health)
    })
}
```

## Deployment Best Practices

### Containerization

Use multi-stage Docker builds for efficient containers:

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/agentflow.toml .

CMD ["./main"]
```

### Configuration Management

Use configuration management for different environments:

```yaml
# docker-compose.yml
version: '3.8'
services:
  agent:
    build: .
    environment:
      - LLM_PROVIDER=azure
      - LLM_API_KEY=${AZURE_OPENAI_API_KEY}
      - LLM_ENDPOINT=${AZURE_OPENAI_ENDPOINT}
      - MEMORY_PROVIDER=pgvector
      - MEMORY_CONNECTION=postgres://user:pass@postgres:5432/agentflow
      - LOG_LEVEL=info
    depends_on:
      - postgres
    
  postgres:
    image: pgvector/pgvector:pg15
    environment:
      POSTGRES_DB: agentflow
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

### Graceful Shutdown

Implement graceful shutdown for clean resource cleanup:

```go
func main() {
    // Create agents and runner
    agents := createAgents()
    runner := createRunner(agents)
    
    // Set up graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Handle shutdown signals
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        log.Println("Shutdown signal received, cleaning up...")
        
        // Cancel context to stop all operations
        cancel()
        
        // Clean up resources
        if err := runner.Stop(); err != nil {
            log.Printf("Error stopping runner: %v", err)
        }
        
        // Close memory connections
        for _, agent := range agents {
            if closer, ok := agent.(io.Closer); ok {
                closer.Close()
            }
        }
        
        log.Println("Cleanup completed")
        os.Exit(0)
    }()
    
    // Start application
    if err := runner.Start(ctx); err != nil {
        log.Fatalf("Failed to start runner: %v", err)
    }
    
    // Keep running until shutdown
    <-ctx.Done()
}
```

## Code Organization

### Project Structure

Organize your project with clear separation of concerns:

```
my-agent-app/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── agents/                 # Agent implementations
│   │   ├── analyzer.go
│   │   ├── processor.go
│   │   └── responder.go
│   ├── config/                 # Configuration management
│   │   └── config.go
│   ├── handlers/               # HTTP handlers (if applicable)
│   │   └── api.go
│   └── services/               # Business logic services
│       └── business_logic.go
├── pkg/                        # Public packages (if any)
├── configs/
│   ├── agentflow.toml         # Default configuration
│   ├── development.toml       # Development overrides
│   └── production.toml        # Production overrides
├── deployments/
│   ├── docker-compose.yml
│   └── Dockerfile
├── scripts/
│   ├── setup.sh
│   └── test.sh
├── tests/
│   ├── integration/
│   └── fixtures/
├── go.mod
├── go.sum
└── README.md
```

### Package Design

Keep packages focused and minimize dependencies:

```go
// Good: Focused package with clear interface
package analyzer

import (
    "context"
    "github.com/kunalkushwaha/agenticgokit/core"
)

type Analyzer interface {
    Analyze(ctx context.Context, text string) (*Analysis, error)
}

type Analysis struct {
    Sentiment string  `json:"sentiment"`
    Keywords  []string `json:"keywords"`
    Score     float64  `json:"score"`
}

// Bad: Package with too many responsibilities
package everything

// Contains agents, configuration, HTTP handlers, database logic, etc.
```

## Documentation

### Code Documentation

Document your agents and their behavior:

```go
// TextAnalyzerAgent analyzes text content for sentiment, keywords, and other metrics.
// It uses an LLM provider to perform the analysis and returns structured results.
//
// The agent expects events with the following data:
//   - text (string): The text content to analyze
//   - analysis_type (string, optional): Type of analysis ("sentiment", "keywords", "full")
//
// Returns results with:
//   - sentiment (string): Detected sentiment ("positive", "negative", "neutral")
//   - keywords ([]string): Extracted keywords
//   - confidence (float64): Confidence score (0.0-1.0)
type TextAnalyzerAgent struct {
    llmProvider core.ModelProvider
    name        string
}

// Run processes a text analysis event and returns structured analysis results.
// It validates the input text and analysis type, then uses the LLM provider
// to perform the requested analysis.
func (a *TextAnalyzerAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Implementation...
}
```

### README Documentation

Provide clear setup and usage instructions:

```markdown
# My Agent Application

Brief description of what your application does.

## Quick Start

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Set up environment:
   ```bash
   export OPENAI_API_KEY="your-key-here"
   ```

3. Run the application:
   ```bash
   go run cmd/main.go
   ```

## Configuration

The application uses `agentflow.toml` for configuration. Key settings:

- `llm.provider`: LLM provider ("openai", "azure", "ollama")
- `memory.provider`: Memory provider ("memory", "pgvector", "weaviate")

## Agents

- **Analyzer**: Analyzes text for sentiment and keywords
- **Processor**: Processes and transforms data
- **Responder**: Generates responses based on analysis

## Deployment

See `deployments/` directory for Docker and Kubernetes configurations.
```

## Next Steps

- **[Testing Agents](testing-agents.md)** - Comprehensive testing strategies
- **[Debugging](debugging.md)** - Debug agent interactions effectively
- **[Production Deployment](../deployment/README.md)** - Production deployment setup