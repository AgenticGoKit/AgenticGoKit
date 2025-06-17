# MCP Integration Quick Reference

**Quick Start Guide for AgentFlow MCP Integration**  
**Created**: June 16, 2025  

## Overview

This document provides a quick reference for developers working on the MCP integration project. It includes key commands, code snippets, and common patterns.

## Project Structure

```
agentflow/
├── core/
│   ├── mcp.go                 # Public MCP interfaces and types  
│   └── mcp_factory.go         # Public factory functions
├── internal/
│   └── mcp/
│       ├── manager.go         # Connection management implementation
│       ├── agent.go           # MCP-aware agent implementation
│       ├── tool.go            # Tool adapter implementation
│       ├── config.go          # Configuration and validation
│       ├── errors.go          # Error types and handling
│       ├── cache.go           # Caching mechanisms
│       └── discovery.go       # Server discovery logic
├── docs/
│   ├── MCP_Integration_Plan.md       # Full implementation plan
│   ├── MCP_Task_Breakdown.md         # Task management
│   ├── MCP_Technical_Specification.md # Technical specs
│   └── MCP_Quick_Reference.md        # This document
└── examples/
    └── mcp/                   # MCP examples (to be created)
```

## Development Workflow

### Setting Up Development Environment

```bash
# 1. Clone and setup
git checkout -b feature/mcp-integration
cd agentflow

# 2. Add MCP Navigator dependency
go get github.com/kunalkushwaha/mcp-navigator-go@latest

# 3. Verify setup
go mod tidy
go build ./...
go test ./...
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test package
go test ./core -v

# Run benchmarks
go test -bench=. ./core

# Run integration tests (when available)
go test -tags=integration ./integration
```

### Code Generation and Tools

```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run

# Generate mocks (if using gomock)
go generate ./...

# Build CLI
go build -o agentcli ./cmd/agentcli
```

## Key Code Patterns

### Creating an MCP Manager

```go
// Import the public API
import "github.com/kunalkushwaha/agentflow/core"

// Basic setup using public interfaces
config := core.MCPConfig{
    AutoDiscover: true,
    DiscoveryTimeout: 10 * time.Second,
    Servers: []core.MCPServerConfig{
        {
            Name: "web_search",
            Type: "tcp",
            Host: "localhost",
            Port: 8811,
            Enabled: true,
        },
    },
}

registry := tools.NewToolRegistry()
manager, err := core.NewMCPManager(config, registry)
if err != nil {
    log.Fatal(err)
}

// Connect to servers
ctx := context.Background()
if err := manager.DiscoverAndConnect(ctx); err != nil {
    log.Fatal(err)
}
defer manager.DisconnectAll()
```

### Creating an MCP Agent

```go
// Create LLM provider
llmProvider := core.NewOpenAIAdapter("your-api-key")

// Create MCP agent using public factory
agent, err := core.NewMCPAgent("mcp-agent", manager, llmProvider)
if err != nil {
    log.Fatal(err)
}

// Use in workflow
inputState := core.NewState()
inputState.Set("query", "search for Go tutorials")

outputState, err := agent.Run(context.Background(), inputState)
if err != nil {
    log.Fatal(err)
}

result := outputState.Get("result")
fmt.Printf("Result: %v\n", result)
```

### Implementing Custom MCP Tool (Internal)

Note: Custom tools would typically be implemented in `internal/` packages.

```go
// This would be in internal/mcp/custom_tool.go
package mcp

import (
    "context"
    "fmt"
    
    "github.com/kunalkushwaha/agentflow/core"
    mcpclient "github.com/kunalkushwaha/mcp-navigator-go/pkg/client"
)

type CustomMCPTool struct {
    name        string
    client      *mcpclient.Client
    serverName  string
}

func (t *CustomMCPTool) Name() string {
    return t.name
}

func (t *CustomMCPTool) Call(ctx context.Context, args map[string]any) (map[string]any, error) {
    // Convert args for MCP
    mcpArgs := convertToMCPArgs(args)
    
    // Call MCP tool
    result, err := t.client.CallTool(ctx, t.name, mcpArgs)
    if err != nil {
        return nil, fmt.Errorf("MCP tool call failed: %w", err)
    }
    
    // Convert result back
    return convertFromMCPResult(result), nil
}
```

### Configuration Examples

#### Simple TCP Server

```toml
[mcp]
auto_discover = true

[[mcp.servers]]
name = "web_search"
type = "tcp"
host = "localhost"
port = 8811
enabled = true
```

#### Docker Server

```toml
[[mcp.servers]]
name = "file_tools"
type = "docker"
container = "mcp-file-server"
enabled = true
timeout = "60s"
```

#### STDIO Server

```toml
[[mcp.servers]]
name = "node_tools"
type = "stdio"
command = "node"
args = ["server.js"]
enabled = true
```

### Error Handling Patterns

```go
// Wrap MCP errors with context
func (m *MCPManager) callTool(ctx context.Context, toolName string, args map[string]any) (map[string]any, error) {
    client, exists := m.getClient(toolName)
    if !exists {
        return nil, &MCPError{
            Type: ErrorTypeConfiguration,
            Message: fmt.Sprintf("no server found for tool: %s", toolName),
            ToolName: toolName,
        }
    }
    
    result, err := client.CallTool(ctx, toolName, args)
    if err != nil {
        return nil, &MCPError{
            Type: ErrorTypeToolExecution,
            Message: "tool execution failed",
            ToolName: toolName,
            ServerName: client.ServerName(),
            Cause: err,
        }
    }
    
    return result, nil
}

// Handle errors gracefully
func handleMCPError(err error) {
    if mcpErr, ok := err.(*MCPError); ok {
        switch mcpErr.Type {
        case ErrorTypeConnection:
            log.Warn("Connection issue, retrying...", "error", err)
        case ErrorTypeToolExecution:
            log.Error("Tool execution failed", "tool", mcpErr.ToolName, "error", err)
        default:
            log.Error("Unknown MCP error", "error", err)
        }
    } else {
        log.Error("Non-MCP error", "error", err)
    }
}
```

## Testing Patterns

### Unit Test Example

```go
func TestMCPTool_Call(t *testing.T) {
    // Setup mock client
    mockClient := &MockMCPClient{
        responses: map[string]mcp.ToolResult{
            "test_tool": {
                Content: []interface{}{
                    map[string]interface{}{
                        "type": "text",
                        "text": "test result",
                    },
                },
            },
        },
    }
    
    // Create tool
    tool := &MCPTool{
        name:   "test_tool",
        client: mockClient,
    }
    
    // Test call
    result, err := tool.Call(context.Background(), map[string]any{
        "query": "test query",
    })
    
    assert.NoError(t, err)
    assert.Equal(t, "test result", result["text"])
}
```

### Integration Test Example

```go
func TestMCPIntegration(t *testing.T) {
    // Skip if no test server
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Setup test server
    server := startTestMCPServer(t)
    defer server.Stop()
    
    // Create manager
    config := MCPConfig{
        Servers: []MCPServerConfig{
            {
                Name: "test_server",
                Type: "tcp",
                Host: "localhost",
                Port: server.Port(),
                Enabled: true,
            },
        },
    }
    
    manager := NewMCPManager(config, tools.NewToolRegistry())
    
    // Test connection
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    err := manager.DiscoverAndConnect(ctx)
    assert.NoError(t, err)
    defer manager.DisconnectAll()
    
    // Test tool execution
    // ... test implementation
}
```

## Common Commands

### Development Commands

```bash
# Add new MCP server for testing
docker run -d --name test-mcp-server -p 8811:8811 mcp-test-server

# Test MCP connection manually
curl http://localhost:8811/health

# Run specific tests
go test -run TestMCP ./core

# Check test coverage
go test -coverprofile=coverage.out ./core
go tool cover -html=coverage.out
```

### CLI Commands (Future)

```bash
# Discover MCP servers
agentcli mcp discover

# Test connection
agentcli mcp connect --server localhost:8811

# List tools
agentcli mcp tools --server web_search

# Test tool
agentcli mcp test-tool --name search --args '{"query": "golang"}'

# Interactive mode
agentcli mcp interactive
```

## Debugging Tips

### Enable Debug Logging

```go
// In your code
logger := log.New(os.Stdout, "[MCP] ", log.LstdFlags|log.Lshortfile)
manager := NewMCPManagerWithLogger(config, registry, logger)

// Via environment
export AGENTFLOW_LOG_LEVEL=debug
export AGENTFLOW_MCP_DEBUG=true
```

### Common Issues and Solutions

#### Connection Refused
```bash
# Check if MCP server is running
netstat -an | grep 8811
curl http://localhost:8811

# Check firewall
sudo ufw status
```

#### Tool Not Found
```go
// Debug tool registration
tools := manager.ListAvailableTools()
for _, tool := range tools {
    fmt.Printf("Tool: %s from server: %s\n", tool.Name, tool.ServerName)
}
```

#### Performance Issues
```go
// Add metrics
metrics := manager.GetMetrics()
fmt.Printf("Average response time: %v\n", metrics.AverageResponseTime)
fmt.Printf("Total calls: %d\n", metrics.TotalCalls)
fmt.Printf("Failed calls: %d\n", metrics.FailedCalls)
```

## Code Review Checklist

### Before Submitting PR

- [ ] All tests pass (`go test ./...`)
- [ ] Code is formatted (`go fmt ./...`)
- [ ] No linting errors (`golangci-lint run`)
- [ ] Documentation is updated
- [ ] Error handling is comprehensive
- [ ] Logging is appropriate
- [ ] Performance impact is considered
- [ ] Backwards compatibility is maintained

### Code Quality Standards

- [ ] Functions are focused and single-purpose
- [ ] Error messages are helpful and actionable
- [ ] Resource cleanup is handled (defer statements)
- [ ] Context cancellation is respected
- [ ] Thread safety is considered for concurrent access
- [ ] Input validation is thorough
- [ ] Return values are documented

## Useful Links

### Documentation
- [MCP Integration Plan](./MCP_Integration_Plan.md)
- [Technical Specification](./MCP_Technical_Specification.md)
- [Task Breakdown](./MCP_Task_Breakdown.md)

### External Resources
- [MCP Navigator Go Library](https://github.com/kunalkushwaha/mcp-navigator-go)
- [MCP Specification](https://spec.modelcontextprotocol.io/)
- [Model Context Protocol](https://modelcontextprotocol.io/)

### AgentFlow Resources
- [AgentFlow Architecture](./Architecture.md)
- [Developer Guide](./DevGuide.md)
- [Library Usage Guide](./LibraryUsageGuide.md)

## Getting Help

### Internal Resources
- Check existing documentation in `docs/`
- Review examples in `examples/`
- Look at test files for usage patterns

### External Resources
- MCP Navigator GitHub issues
- AgentFlow GitHub discussions
- Model Context Protocol community

---

**Last Updated**: June 16, 2025  
**Maintainer**: Development Team  
**Next Update**: After Phase 1 completion
