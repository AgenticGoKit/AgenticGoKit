# Code Style Guide

This document defines the coding standards and conventions for AgenticGoKit to ensure consistency, readability, and maintainability across the codebase.

## 🎯 Core Principles

1. **Clarity over Cleverness**: Write code that is easy to understand
2. **Consistency**: Follow established patterns throughout the codebase
3. **Simplicity**: Prefer simple solutions over complex ones
4. **Performance**: Be mindful of performance implications
5. **Documentation**: Code should be self-documenting with helpful comments

## 🏗️ Go Language Standards

### Follow Standard Go Conventions

AgenticGoKit adheres to all standard Go conventions:

- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Best Practices](https://golang.org/doc/faq)

### Formatting and Tools

Use the standard Go toolchain for consistent formatting:

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Check for unused code
go mod tidy

# Vet for common mistakes
go vet ./...
```

### Required Tools Configuration

#### `.golangci.yml`
```yaml
run:
  timeout: 5m
  modules-download-mode: readonly

linters-settings:
  gocyclo:
    min-complexity: 15
  
  goconst:
    min-len: 3
    min-occurrences: 3
  
  goimports:
    local-prefixes: github.com/kunalkushwaha/agentflow
  
  govet:
    check-shadowing: true
  
  misspell:
    locale: US

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - gocyclo
    - goconst
    - misspell
    - ineffassign
    - staticcheck
    - unused
    - errcheck
    - gosimple
    - deadcode
    - varcheck
    - typecheck

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gocyclo
        - errcheck
        - dupl
        - gosec
```

## 📁 Package Organization

### Directory Structure

```
agentflow/
├── cmd/                    # Main applications
│   └── agentcli/          # CLI application
├── core/                   # Public API
│   ├── agent.go           # Core interfaces
│   ├── runner.go          # Public runner interface
│   └── *.go               # Other public APIs
├── internal/               # Private implementation
│   ├── agents/            # Agent implementations
│   ├── mcp/               # MCP implementation
│   ├── llm/               # LLM provider implementations
│   └── */                 # Other internal packages
├── pkg/                    # Utility packages (if needed)
├── examples/               # Example applications
├── docs/                   # Documentation
└── scripts/                # Build and development scripts
```

### Package Naming

- Use lowercase, single-word package names
- Avoid underscores or mixed case
- Package names should be descriptive but concise
- Avoid generic names like `util`, `common`, `base`

```go
// Good
package mcp
package agents
package llm

// Bad
package mcpUtils
package agent_handlers
package LLMProviders
```

### Import Organization

Group imports in this order with blank lines between groups:

```go
package core

import (
    // Standard library
    "context"
    "fmt"
    "time"
    
    // Third-party dependencies
    "github.com/spf13/cobra"
    "go.opentelemetry.io/otel/trace"
    
    // Local imports
    "github.com/kunalkushwaha/agentflow/internal/mcp"
    "github.com/kunalkushwaha/agentflow/pkg/utils"
)
```

## 🏷️ Naming Conventions

### Variables and Functions

Use camelCase for variables and functions:

```go
// Good
var agentCount int
var lastExecutionTime time.Time
func executeAgent() error
func getAgentByName(name string) Agent

// Bad
var agent_count int
var LastExecutionTime time.Time
func execute_agent() error
func GetAgentByName(name string) Agent
```

### Constants

Use camelCase for unexported constants, PascalCase for exported:

```go
// Good
const defaultTimeout = 30 * time.Second
const MaxRetryAttempts = 3

// Bad
const DEFAULT_TIMEOUT = 30 * time.Second
const max_retry_attempts = 3
```

### Types and Interfaces

Use PascalCase for exported types, camelCase for unexported:

```go
// Good
type Agent interface {}
type AgentHandler interface {}
type llmProvider struct {}

// Bad
type agent interface {}
type agentHandler interface {}
type LLMProvider struct {}
```

### Interface Naming

- Use noun or adjective forms
- Single-method interfaces often end with "-er"
- Avoid "I" prefix

```go
// Good
type Runner interface {}
type AgentHandler interface {}
type Executor interface {}

// Bad
type IRunner interface {}
type AgentHandlerInterface interface {}
type ExecutorImpl interface {}
```

## 📝 Documentation Standards

### Package Documentation

Every package should have a package comment:

```go
// Package core provides the public API for AgentFlow.
// It defines the primary interfaces and types used to build
// AI agent systems with dynamic tool integration.
//
// The core package follows a clear separation between interfaces
// (defined here) and implementations (in internal packages).
//
// Example usage:
//
//	config := &core.Config{...}
//	runner, err := core.NewRunner(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//	
//	handler := &MyAgent{}
//	runner.RegisterAgent("my-agent", handler)
package core
```

### Function Documentation

Document all exported functions with their purpose, parameters, return values, and any side effects:

```go
// NewRunner creates a new agent runner with the provided configuration.
// It initializes all configured LLM providers and MCP servers.
//
// The runner will not start processing events until Start() is called.
// Configuration errors will be returned immediately, while connection
// errors to external services may be retried automatically.
//
// Parameters:
//   - config: Configuration for the runner and its dependencies
//
// Returns:
//   - *Runner: Configured runner instance
//   - error: Configuration or initialization error
func NewRunner(config *Config) (*Runner, error) {
    // Implementation...
}
```

### Type Documentation

Document types, especially interfaces:

```go
// AgentHandler defines the interface for implementing agent logic.
//
// Implementations should be stateless and thread-safe, as the same
// handler may be called concurrently by multiple goroutines.
//
// The Run method should process the input event and state, perform
// any necessary operations (including calling tools or LLMs), and
// return the result with any state changes.
type AgentHandler interface {
    // Run processes an event and returns the result.
    //
    // The context may include deadlines, cancellation, and tracing
    // information. Implementations should respect context cancellation.
    //
    // The event contains the input data and metadata. The state
    // represents the current session state and may be modified.
    //
    // Returns AgentResult with response data and updated state,
    // or an error if processing fails.
    Run(ctx context.Context, event Event, state State) (AgentResult, error)
}
```

### Comment Guidelines

- Use complete sentences with proper capitalization and punctuation
- Explain "why" not just "what"
- Include examples for complex functionality
- Document any limitations or gotchas

```go
// validateConfig checks the configuration for common errors and
// provides helpful suggestions for fixes.
//
// This validation is performed at startup to catch configuration
// issues early, before attempting to connect to external services.
// Some validations (like network connectivity) are performed lazily.
func validateConfig(config *Config) error {
    // Check required fields first to provide clear error messages
    if config.LLM.Provider == "" {
        return fmt.Errorf("llm.provider is required")
    }
    
    // Validate provider-specific configuration
    switch config.LLM.Provider {
    case "azure":
        return validateAzureConfig(&config.LLM.Azure)
    case "openai":
        return validateOpenAIConfig(&config.LLM.OpenAI)
    default:
        return fmt.Errorf("unsupported llm provider: %s", config.LLM.Provider)
    }
}
```

## 🔧 Error Handling

### Error Types

Define specific error types for different categories:

```go
// ValidationError represents a configuration or input validation error
type ValidationError struct {
    Field   string
    Value   interface{}
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for field %s: %s", e.Field, e.Message)
}

// TimeoutError represents an operation timeout
type TimeoutError struct {
    Operation string
    Duration  time.Duration
}

func (e *TimeoutError) Error() string {
    return fmt.Sprintf("operation %s timed out after %v", e.Operation, e.Duration)
}
```

### Error Wrapping

Use error wrapping to provide context:

```go
func executeAgent(ctx context.Context, agent AgentHandler, event Event, state State) (AgentResult, error) {
    result, err := agent.Run(ctx, event, state)
    if err != nil {
        return AgentResult{}, fmt.Errorf("failed to execute agent %s: %w", agent.Name(), err)
    }
    return result, nil
}
```

### Error Messages

- Start with lowercase letter (Go convention)
- Be specific and actionable
- Include relevant context
- Avoid implementation details in user-facing errors

```go
// Good
return fmt.Errorf("failed to connect to MCP server %s: %w", serverName, err)
return ValidationError{Field: "timeout", Message: "must be positive"}

// Bad
return fmt.Errorf("Connection failed")
return fmt.Errorf("Error in line 42 of mcp.go")
```

## 🧪 Testing Standards

### Test File Organization

- Test files should be in the same package as the code they test
- Use `_test.go` suffix
- Group related tests in the same file

```go
// agent_test.go
package core

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)
```

### Test Function Naming

Use descriptive test names that explain the scenario:

```go
// Good
func TestAgent_Run_WithValidInput_ReturnsSuccess(t *testing.T) {}
func TestRunner_RegisterAgent_WithNilHandler_ReturnsError(t *testing.T) {}
func TestMCPManager_ExecuteTool_WhenServerUnavailable_RetriesAndFails(t *testing.T) {}

// Bad
func TestAgent(t *testing.T) {}
func TestRunner1(t *testing.T) {}
func TestError(t *testing.T) {}
```

### Test Structure

Use the Arrange-Act-Assert pattern:

```go
func TestAgent_Run_WithValidInput_ReturnsSuccess(t *testing.T) {
    // Arrange
    agent := &TestAgent{name: "test-agent"}
    event := NewEvent("test", map[string]interface{}{
        "query": "Hello world",
    })
    state := NewState()
    ctx := context.Background()
    
    // Act
    result, err := agent.Run(ctx, event, state)
    
    // Assert
    require.NoError(t, err)
    assert.True(t, result.Success)
    assert.Equal(t, "Hello world", result.Data["processed_query"])
}
```

### Table-Driven Tests

Use table-driven tests for multiple scenarios:

```go
func TestValidateConfig(t *testing.T) {
    tests := []struct {
        name    string
        config  Config
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid azure config",
            config: Config{
                LLM: LLMConfig{
                    Provider: "azure",
                    Azure: AzureConfig{
                        Endpoint: "https://test.openai.azure.com",
                        APIKey:   "test-key",
                    },
                },
            },
            wantErr: false,
        },
        {
            name: "missing provider",
            config: Config{
                LLM: LLMConfig{},
            },
            wantErr: true,
            errMsg:  "llm.provider is required",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateConfig(&tt.config)
            
            if tt.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

## 🚀 Performance Guidelines

### Memory Allocation

Minimize allocations in hot paths:

```go
// Good - reuse slice capacity
func processEvents(events []Event) []Result {
    results := make([]Result, 0, len(events))
    for _, event := range events {
        result := processEvent(event)
        results = append(results, result)
    }
    return results
}

// Bad - repeated allocations
func processEvents(events []Event) []Result {
    var results []Result
    for _, event := range events {
        result := processEvent(event)
        results = append(results, result)
    }
    return results
}
```

### String Building

Use strings.Builder for efficient string concatenation:

```go
// Good
func buildPrompt(parts []string) string {
    var builder strings.Builder
    builder.Grow(len(parts) * 50) // Pre-allocate if size is known
    
    for i, part := range parts {
        if i > 0 {
            builder.WriteString("\n")
        }
        builder.WriteString(part)
    }
    
    return builder.String()
}

// Bad
func buildPrompt(parts []string) string {
    result := ""
    for i, part := range parts {
        if i > 0 {
            result += "\n"
        }
        result += part
    }
    return result
}
```

### Context Usage

Always pass context and respect cancellation:

```go
func processWithTimeout(ctx context.Context, data []byte) error {
    // Create timeout context
    timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // Check for cancellation in loops
    for i, item := range data {
        select {
        case <-timeoutCtx.Done():
            return timeoutCtx.Err()
        default:
        }
        
        if err := processItem(timeoutCtx, item); err != nil {
            return fmt.Errorf("failed to process item %d: %w", i, err)
        }
    }
    
    return nil
}
```

## 🔒 Security Guidelines

### Input Validation

Validate all external inputs:

```go
func processUserQuery(query string) error {
    // Validate input length
    if len(query) > maxQueryLength {
        return ValidationError{
            Field:   "query",
            Message: fmt.Sprintf("exceeds maximum length of %d characters", maxQueryLength),
        }
    }
    
    // Check for malicious content
    if containsSQLInjection(query) {
        return ValidationError{
            Field:   "query",
            Message: "contains potentially malicious content",
        }
    }
    
    return nil
}
```

### Secrets Handling

Never log or expose secrets:

```go
// Good
func logConfig(config *Config) {
    log.Printf("LLM Provider: %s", config.LLM.Provider)
    log.Printf("Endpoint: %s", config.LLM.Azure.Endpoint)
    // Don't log API key
}

// Bad
func logConfig(config *Config) {
    log.Printf("Config: %+v", config) // This might expose secrets
}
```

### Resource Limits

Implement appropriate limits:

```go
const (
    maxConcurrentRequests = 100
    maxRequestSize       = 10 * 1024 * 1024 // 10MB
    maxExecutionTime     = 5 * time.Minute
)

func processRequest(ctx context.Context, request Request) error {
    // Check size limits
    if len(request.Data) > maxRequestSize {
        return fmt.Errorf("request too large: %d bytes", len(request.Data))
    }
    
    // Set timeout
    timeoutCtx, cancel := context.WithTimeout(ctx, maxExecutionTime)
    defer cancel()
    
    return doProcessRequest(timeoutCtx, request)
}
```

## 📋 Code Review Checklist

### Before Submitting

- [ ] Code follows Go formatting standards (`go fmt`)
- [ ] All linters pass (`golangci-lint run`)
- [ ] Tests are written and passing
- [ ] Documentation is updated
- [ ] Error handling is appropriate
- [ ] Performance implications considered
- [ ] Security implications considered

### During Review

- [ ] Code is readable and well-structured
- [ ] Variable and function names are clear
- [ ] Comments explain complex logic
- [ ] Error messages are helpful
- [ ] Tests cover edge cases
- [ ] No obvious performance issues
- [ ] Follows established patterns

This code style guide ensures consistency and quality across the AgentFlow codebase, making it easier for contributors to understand, maintain, and extend the system.
