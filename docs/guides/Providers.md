# LLM Providers

**Multi-Provider LLM Integration in AgentFlow**

AgentFlow provides a unified interface for working with different LLM providers. This guide covers configuration, usage patterns, and provider-specific features.

## Provider Overview

AgentFlow supports multiple LLM providers through a unified `ModelProvider` interface:

- **Azure OpenAI** (Default) - Enterprise-ready with robust scaling
- **OpenAI** - Direct API access to GPT models
- **Ollama** - Local models for privacy and cost control
- **Mock** - Testing and development

## ModelProvider Interface

All providers implement the same interface:

```go
type ModelProvider interface {
    Generate(ctx context.Context, prompt string) (string, error)
    GenerateWithHistory(ctx context.Context, messages []Message) (string, error)
    Name() string
}

type Message struct {
    Role    string // "system", "user", "assistant"
    Content string
}
```

## Configuration

### Azure OpenAI (Default)

**agentflow.toml:**
```toml
[provider]
type = "azure"
api_key = "${AZURE_OPENAI_API_KEY}"
endpoint = "https://your-resource.openai.azure.com"
deployment = "gpt-4"
api_version = "2024-02-15-preview"
model = "gpt-4"
max_tokens = 2000
temperature = 0.7
```

**Environment variables:**
```bash
export AZURE_OPENAI_API_KEY="your-azure-api-key"
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com"
export AZURE_OPENAI_DEPLOYMENT="gpt-4"
```

**Programmatic configuration:**
```go
import agenticgokit "github.com/kunalkushwaha/AgenticGoKit/core"

config := agenticgokit.AzureConfig{
    APIKey:     "your-api-key",
    Endpoint:   "https://your-resource.openai.azure.com",
    Deployment: "gpt-4",
    APIVersion: "2024-02-15-preview",
    MaxTokens:  2000,
    Temperature: 0.7,
}

provider, err := agentflow.NewAzureProvider(config)
if err != nil {
    log.Fatal(err)
}
```

### OpenAI

**agentflow.toml:**
```toml
[provider]
type = "openai"
api_key = "${OPENAI_API_KEY}"
model = "gpt-4"
max_tokens = 2000
temperature = 0.7
organization = "your-org-id"  # Optional
```

**Environment variables:**
```bash
export OPENAI_API_KEY="your-openai-api-key"
export OPENAI_ORG="your-org-id"  # Optional
```

**Programmatic configuration:**
```go
config := agentflow.OpenAIConfig{
    APIKey:      "your-api-key",
    Model:       "gpt-4",
    MaxTokens:   2000,
    Temperature: 0.7,
    Organization: "your-org-id", // Optional
}

provider, err := agentflow.NewOpenAIProvider(config)
```

### Ollama (Local Models)

**agentflow.toml:**
```toml
[provider]
type = "ollama"
host = "http://localhost:11434"
model = "llama3.2:3b"
temperature = 0.7
context_window = 4096
```

**Setup Ollama:**
```bash
# Install Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# Pull a model
ollama pull llama3.2:3b

# Start Ollama server (usually automatic)
ollama serve
```

**Programmatic configuration:**
```go
config := agentflow.OllamaConfig{
    Host:          "http://localhost:11434",
    Model:         "llama3.2:3b",
    Temperature:   0.7,
    ContextWindow: 4096,
}

provider, err := agentflow.NewOllamaProvider(config)
```

### Mock Provider (Testing)

**agentflow.toml:**
```toml
[provider]
type = "mock"
response = "This is a mock response"
delay = "100ms"  # Simulate network delay
```

**Programmatic configuration:**
```go
config := agentflow.MockConfig{
    Response: "This is a mock response",
    Delay:    100 * time.Millisecond,
}

provider := agentflow.NewMockProvider(config)
```

## Provider Factory

AgentFlow provides factory functions for easy provider creation:

### From Configuration File

```go
// Automatically load from agentflow.toml
provider, err := agentflow.NewProviderFromWorkingDir()
if err != nil {
    log.Fatal(err)
}

// Or from specific directory
provider, err := agentflow.NewProviderFromDir("/path/to/config")
```

### From Environment

```go
// Auto-detect provider from environment variables
provider, err := agentflow.NewProviderFromEnv()
if err != nil {
    log.Fatal(err)
}
```

### Explicit Configuration

```go
// Create specific provider
azureProvider, err := agentflow.NewAzureProvider(azureConfig)
openaiProvider, err := agentflow.NewOpenAIProvider(openaiConfig)
ollamaProvider, err := agentflow.NewOllamaProvider(ollamaConfig)
```

## Usage Patterns

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    agentflow "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // Load provider from configuration
    provider, err := agentflow.NewProviderFromWorkingDir()
    if err != nil {
        panic(err)
    }
    
    ctx := context.Background()
    
    // Simple generation
    response, err := provider.Generate(ctx, "Explain Go interfaces in simple terms")
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Response:", response)
    fmt.Println("Provider:", provider.Name())
}
```

### Conversation History

```go
func handleConversation(provider agentflow.ModelProvider) {
    ctx := context.Background()
    
    // Build conversation history
    messages := []agentflow.Message{
        {Role: "system", Content: "You are a helpful programming assistant."},
        {Role: "user", Content: "How do I create a web server in Go?"},
        {Role: "assistant", Content: "You can create a web server in Go using the net/http package..."},
        {Role: "user", Content: "Can you show me a more complex example?"},
    }
    
    response, err := provider.GenerateWithHistory(ctx, messages)
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }
    
    fmt.Println("Response:", response)
}
```

### Agent with Provider

```go
type MyAgent struct {
    provider agentflow.ModelProvider
    name     string
}

func NewMyAgent(name string, provider agentflow.ModelProvider) *MyAgent {
    return &MyAgent{
        name:     name,
        provider: provider,
    }
}

func (a *MyAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    message := event.GetData()["message"]
    
    systemPrompt := fmt.Sprintf("You are %s, a specialized assistant.", a.name)
    fullPrompt := fmt.Sprintf("%s\n\nUser: %s", systemPrompt, message)
    
    response, err := a.provider.Generate(ctx, fullPrompt)
    if err != nil {
        return agentflow.AgentResult{}, fmt.Errorf("provider error: %w", err)
    }
    
    state.Set("response", response)
    state.Set("provider", a.provider.Name())
    
    return agentflow.AgentResult{
        Result: response,
        State:  state,
    }, nil
}
```

## Provider Selection Strategies

### Environment-Based Selection

```go
func createProvider() (agentflow.ModelProvider, error) {
    providerType := os.Getenv("LLM_PROVIDER")
    
    switch providerType {
    case "azure":
        return agentflow.NewAzureProvider(agentflow.AzureConfig{
            APIKey:     os.Getenv("AZURE_OPENAI_API_KEY"),
            Endpoint:   os.Getenv("AZURE_OPENAI_ENDPOINT"),
            Deployment: os.Getenv("AZURE_OPENAI_DEPLOYMENT"),
        })
    case "openai":
        return agentflow.NewOpenAIProvider(agentflow.OpenAIConfig{
            APIKey: os.Getenv("OPENAI_API_KEY"),
            Model:  "gpt-4",
        })
    case "ollama":
        return agentflow.NewOllamaProvider(agentflow.OllamaConfig{
            Host:  "http://localhost:11434",
            Model: "llama3.2:3b",
        })
    default:
        return agentflow.NewMockProvider(agentflow.MockConfig{
            Response: "Mock response for testing",
        }), nil
    }
}
```

### Feature-Based Selection

```go
type ProviderSelector struct {
    primary   agentflow.ModelProvider
    fallback  agentflow.ModelProvider
    local     agentflow.ModelProvider
}

func NewProviderSelector() *ProviderSelector {
    azure, _ := agentflow.NewAzureProvider(azureConfig)
    openai, _ := agentflow.NewOpenAIProvider(openaiConfig)
    ollama, _ := agentflow.NewOllamaProvider(ollamaConfig)
    
    return &ProviderSelector{
        primary:  azure,   // Enterprise workloads
        fallback: openai,  // Backup when Azure is down
        local:    ollama,  // Privacy-sensitive tasks
    }
}

func (p *ProviderSelector) SelectProvider(task string) agentflow.ModelProvider {
    if strings.Contains(task, "confidential") || strings.Contains(task, "private") {
        return p.local // Use local model for sensitive content
    }
    
    // Try primary first, fallback if needed
    return p.primary
}

func (p *ProviderSelector) GenerateWithFallback(ctx context.Context, prompt string) (string, error) {
    // Try primary provider
    response, err := p.primary.Generate(ctx, prompt)
    if err == nil {
        return response, nil
    }
    
    // Log primary failure and try fallback
    log.Printf("Primary provider failed: %v, trying fallback", err)
    response, err = p.fallback.Generate(ctx, prompt)
    if err == nil {
        return response, nil
    }
    
    // Both failed
    return "", fmt.Errorf("all providers failed: %w", err)
}
```

## Error Handling

### Provider-Specific Error Handling

```go
func handleProviderError(err error, provider agentflow.ModelProvider) string {
    providerName := provider.Name()
    
    switch {
    case strings.Contains(err.Error(), "rate limit"):
        return fmt.Sprintf("Rate limit reached for %s. Please try again later.", providerName)
    case strings.Contains(err.Error(), "authentication"):
        return fmt.Sprintf("Authentication failed for %s. Please check your API key.", providerName)
    case strings.Contains(err.Error(), "timeout"):
        return fmt.Sprintf("Request to %s timed out. Please try again.", providerName)
    case strings.Contains(err.Error(), "model"):
        return fmt.Sprintf("Model not available on %s. Please check your configuration.", providerName)
    default:
        return fmt.Sprintf("Error with %s: %v", providerName, err)
    }
}

func (a *Agent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    message := event.GetData()["message"]
    
    response, err := a.provider.Generate(ctx, prompt)
    if err != nil {
        errorMsg := handleProviderError(err, a.provider)
        
        // Set error in state for debugging
        state.Set("error", errorMsg)
        state.Set("provider_error", true)
        
        // Return user-friendly error
        return agentflow.AgentResult{
            Result: "I'm having trouble processing your request right now. Please try again.",
            State:  state,
        }, nil // Don't propagate error, handle gracefully
    }
    
    return agentflow.AgentResult{Result: response, State: state}, nil
}
```

### Retry Logic

```go
func generateWithRetry(ctx context.Context, provider agentflow.ModelProvider, prompt string, maxRetries int) (string, error) {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        response, err := provider.Generate(ctx, prompt)
        if err == nil {
            return response, nil
        }
        
        lastErr = err
        
        // Don't retry on authentication errors
        if strings.Contains(err.Error(), "authentication") {
            break
        }
        
        // Exponential backoff
        if i < maxRetries-1 {
            backoff := time.Duration(i+1) * time.Second
            log.Printf("Retry %d/%d after %v: %v", i+1, maxRetries, backoff, err)
            time.Sleep(backoff)
        }
    }
    
    return "", fmt.Errorf("all retries failed: %w", lastErr)
}
```

## Testing with Providers

### Mock Provider for Testing

```go
func TestAgentWithMockProvider(t *testing.T) {
    // Create mock provider
    mockProvider := agentflow.NewMockProvider(agentflow.MockConfig{
        Response: "This is a test response",
        Delay:    10 * time.Millisecond,
    })
    
    // Create agent with mock
    agent := NewMyAgent("test-agent", mockProvider)
    
    // Test
    eventData := agentflow.EventData{"message": "test message"}
    event := agentflow.NewEvent("test", eventData, nil)
    state := agentflow.NewState()
    
    result, err := agent.Run(context.Background(), event, state)
    
    assert.NoError(t, err)
    assert.Equal(t, "This is a test response", result.Result)
    assert.Equal(t, "mock", mockProvider.Name())
}
```

### Provider Integration Tests

```go
func TestProviderIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    tests := []struct {
        name     string
        provider agentflow.ModelProvider
        prompt   string
    }{
        {
            name:     "azure",
            provider: createAzureProvider(t),
            prompt:   "Say hello",
        },
        {
            name:     "openai", 
            provider: createOpenAIProvider(t),
            prompt:   "Say hello",
        },
        {
            name:     "ollama",
            provider: createOllamaProvider(t),
            prompt:   "Say hello",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            defer cancel()
            
            response, err := tt.provider.Generate(ctx, tt.prompt)
            assert.NoError(t, err)
            assert.NotEmpty(t, response)
            assert.Contains(t, strings.ToLower(response), "hello")
        })
    }
}
```

## Performance Considerations

### Provider Performance Characteristics

| Provider | Latency | Throughput | Cost | Privacy |
|----------|---------|------------|------|---------|
| Azure OpenAI | Medium | High | Medium | Enterprise |
| OpenAI | Medium | High | Medium | Cloud |
| Ollama | Low | Medium | Free | Full |
| Mock | Minimal | Very High | Free | Full |

### Optimization Strategies

```go
// Connection pooling for HTTP providers
config := agentflow.AzureConfig{
    APIKey:     apiKey,
    Endpoint:   endpoint,
    Deployment: deployment,
    // Performance settings
    MaxRetries:     3,
    RequestTimeout: 30 * time.Second,
    MaxConnections: 100,
}

// Context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := provider.Generate(ctx, prompt)
```

## Production Deployment

### Environment Configuration

```bash
# Production environment variables
export LLM_PROVIDER="azure"
export AZURE_OPENAI_API_KEY="prod-api-key"
export AZURE_OPENAI_ENDPOINT="https://prod-resource.openai.azure.com"
export AZURE_OPENAI_DEPLOYMENT="gpt-4"

# Fallback provider
export FALLBACK_PROVIDER="openai"
export OPENAI_API_KEY="fallback-api-key"

# Local development
export LLM_PROVIDER="ollama"
export OLLAMA_HOST="http://localhost:11434"
export OLLAMA_MODEL="llama3.2:3b"
```

### Monitoring Provider Health

```go
type ProviderHealthChecker struct {
    provider agentflow.ModelProvider
}

func (h *ProviderHealthChecker) HealthCheck(ctx context.Context) error {
    testPrompt := "Health check"
    
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    
    _, err := h.provider.Generate(ctx, testPrompt)
    return err
}

// Use in health endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
    checker := &ProviderHealthChecker{provider: globalProvider}
    
    if err := checker.HealthCheck(r.Context()); err != nil {
        http.Error(w, "Provider unhealthy: "+err.Error(), http.StatusServiceUnavailable)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
```

## Next Steps

- **[Configuration](Configuration.md)** - Advanced configuration management
- **[Tool Integration](ToolIntegration.md)** - Add MCP tools to your agents
- **[Production Deployment](Production.md)** - Deploy at scale
- **[Custom Providers](../contributors/AddingFeatures.md)** - Build custom provider adapters (for contributors)
