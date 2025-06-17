# AgentFlow Create Command: MCP-Enabled Multi-Agent Workflow Generator

**Created**: June 17, 2025  
**Status**: Design Phase  
**Version**: 1.0  

## Overview

This document outlines the design for extending the `agentcli create` command to generate MCP-enabled multi-agent workflows, leveraging the existing AgentFlow framework with advanced Model Context Protocol capabilities.

## Current AgentFlow Create Command Analysis

Based on README and documentation analysis, the existing `agentcli create` command is designed to support:

### Current Features (From Documentation)
```bash
# Basic project generation
agentcli create myproject --agents 2 --provider ollama

# Interactive mode
agentcli create --interactive

# Provider selection
agentcli create myproject --agents 3 --provider azure
```

### Current Generated Structure
- **Sequential workflow**: agent1 → agent2 → responsible_ai → workflow_finalizer
- **Configuration file**: `agentflow.toml` 
- **Error handlers**: validation, timeout, critical error handling
- **LLM integration**: OpenAI, Azure, Ollama, Mock providers
- **Session management**: Automatic tracking and correlation
- **Modern patterns**: Factory functions and AgentFlow v0.1.1 APIs

## MCP-Enhanced Design

### New Command Options

```bash
# Basic MCP-enabled project
agentcli create myproject --with-mcp

# MCP with specific tools
agentcli create myproject --with-mcp --mcp-tools="web_search,file_ops,database"

# Full MCP production setup
agentcli create myproject --with-mcp --mcp-production --agents 3

# Interactive MCP configuration
agentcli create --interactive --with-mcp

# MCP with caching and metrics
agentcli create myproject --with-mcp --with-cache --with-metrics --provider openai

# MCP multi-server setup
agentcli create myproject --with-mcp --mcp-servers="web-service,file-service,db-service"
```

### Extended CLI Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--with-mcp` | bool | false | Enable MCP tool integration |
| `--mcp-tools` | string | "web_search,summarize" | Comma-separated list of MCP tools |
| `--mcp-servers` | string | "demo-server" | Comma-separated list of MCP server names |
| `--with-cache` | bool | false | Enable MCP result caching |
| `--cache-backend` | string | "memory" | Cache backend (memory, redis) |
| `--with-metrics` | bool | false | Enable Prometheus metrics |
| `--metrics-port` | int | 8080 | Metrics server port |
| `--mcp-production` | bool | false | Include production features (pooling, retry, etc.) |
| `--with-load-balancer` | bool | false | Enable MCP load balancing |
| `--connection-pool-size` | int | 5 | MCP connection pool size |
| `--retry-policy` | string | "exponential" | Retry policy (exponential, linear, fixed) |

### Generated Project Structure

```
myproject/
├── main.go                    # Entry point with MCP setup
├── agentflow.toml            # Enhanced config with MCP settings
├── agents/
│   ├── agent1.go             # MCP-aware agent implementations
│   ├── agent2.go
│   └── agent3.go
├── mcp/
│   ├── config.go             # MCP server configurations
│   ├── tools.go              # MCP tool definitions and registry
│   ├── cache.go              # Cache configuration (if --with-cache)
│   └── metrics.go            # Metrics setup (if --with-metrics)
├── handlers/
│   ├── error_handler.go      # Enhanced with MCP error handling
│   ├── timeout_handler.go
│   └── workflow_finalizer.go
├── configs/
│   ├── mcp_servers.toml      # MCP server connection configs
│   └── cache_config.toml     # Cache settings (if enabled)
└── README.md                 # Project documentation with MCP usage
```

## MCP Integration Levels

### Level 1: Basic MCP Integration (`--with-mcp`)
- Single MCP server connection
- 2-3 predefined tools (web_search, summarize, translate)
- Simple agent with MCP tool execution
- Basic error handling for MCP failures

### Level 2: Enhanced MCP (`--with-mcp --with-cache`)
- MCP result caching with configurable TTL
- Cache hit/miss metrics
- Multiple tool types with different cache strategies
- Cache cleanup and management

### Level 3: Production MCP (`--mcp-production`)
- Connection pooling with health checks
- Advanced retry policies with circuit breakers
- Load balancing across multiple MCP servers
- Comprehensive Prometheus metrics
- Production-ready error handling and recovery

### Level 4: Enterprise MCP (`--mcp-production --with-load-balancer`)
- Multi-server MCP architecture
- Failover and high availability
- Advanced load balancing strategies
- Full observability stack
- Configuration hot-reloading

## Generated Code Templates

### Enhanced main.go Template
```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/kunalkushwaha/agentflow/core"
    "github.com/kunalkushwaha/agentflow/internal/mcp"
    // ... other imports
)

func main() {
    log.Println("Starting MCP-Enhanced AgentFlow Workflow")
    
    ctx := context.Background()
    
    // 1. Load configuration
    config, err := loadConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // 2. Initialize MCP components
    mcpManager, err := initializeMCP(config)
    if err != nil {
        log.Fatalf("Failed to initialize MCP: %v", err)
    }
    
    // 3. Create MCP-aware agents
    agents := createMCPAgents(mcpManager, config)
    
    // 4. Setup workflow with MCP error handling
    runner := setupWorkflowRunner(agents, mcpManager)
    
    // 5. Execute workflow
    result, err := runner.Run(ctx, initialEvent)
    if err != nil {
        log.Fatalf("Workflow failed: %v", err)
    }
    
    log.Printf("Workflow completed: %v", result)
}
```

### MCP Configuration Template (agentflow.toml)
```toml
[agentflow]
provider = "{{.Provider}}"
log_level = "info"
session_timeout = "5m"

[mcp]
enabled = {{.MCPEnabled}}
default_timeout = "30s"
max_retries = 3

{{if .WithCache}}
[mcp.cache]
enabled = true
backend = "{{.CacheBackend}}"
default_ttl = "5m"
max_size_mb = 100
max_keys = 1000

[mcp.cache.tool_ttls]
web_search = "2m"
summarize = "10m"
translate = "30m"
{{end}}

{{if .WithMetrics}}
[mcp.metrics]
enabled = true
port = {{.MetricsPort}}
path = "/metrics"
{{end}}

{{range .MCPServers}}
[[mcp.servers]]
name = "{{.Name}}"
transport = "{{.Transport}}"
{{if eq .Transport "stdio"}}
command = "{{.Command}}"
args = {{.Args}}
{{else if eq .Transport "http"}}
url = "{{.URL}}"
{{end}}
{{end}}
```

### MCP-Aware Agent Template
```go
package agents

import (
    "context"
    "fmt"
    
    "github.com/kunalkushwaha/agentflow/core"
    "github.com/kunalkushwaha/agentflow/internal/mcp"
)

type {{.AgentName}}Agent struct {
    name        string
    mcpManager  *mcp.CacheManager
    tools       []string
}

func New{{.AgentName}}Agent(mcpManager *mcp.CacheManager) *{{.AgentName}}Agent {
    return &{{.AgentName}}Agent{
        name:       "{{.AgentName | lower}}",
        mcpManager: mcpManager,
        tools:      []string{{"{{.Tools}}"}},
    }
}

func (a *{{.AgentName}}Agent) Process(ctx context.Context, event core.Event) (core.State, error) {
    input := event.Data()["input"].(string)
    
    // Execute MCP tool with caching
    execution := core.MCPToolExecution{
        ToolName:   "{{.PrimaryTool}}",
        ServerName: "{{.ServerName}}",
        Arguments: map[string]interface{}{
            "query": input,
        },
    }
    
    result, err := a.mcpManager.ExecuteWithCache(ctx, execution)
    if err != nil {
        return core.State{}, fmt.Errorf("MCP tool execution failed: %w", err)
    }
    
    // Process result and create output state
    response := fmt.Sprintf("%s result: %s", a.name, result.Content[0].Text)
    
    return core.State{
        AgentID: a.name,
        Data: map[string]interface{}{
            "response": response,
            "tool_used": execution.ToolName,
            "duration": result.Duration,
        },
    }, nil
}

func (a *{{.AgentName}}Agent) ID() string {
    return a.name
}
```

## Workflow Patterns for MCP Integration

### Pattern 1: Sequential MCP Chain
```
Input → MCP_Agent1(web_search) → MCP_Agent2(summarize) → MCP_Agent3(translate) → Output
```

### Pattern 2: Parallel MCP Processing
```
Input → [MCP_Agent1(research), MCP_Agent2(sentiment)] → Aggregator → Output
```

### Pattern 3: MCP Tool Router
```
Input → Router → {web_search|file_ops|database} → Processor → Output
```

### Pattern 4: MCP Error Recovery
```
Input → MCP_Agent → [Success → Output | Error → Fallback_Agent → Output]
```

## Implementation Plan

### Phase 1: Basic Command Structure (Week 1)
- [ ] Create `cmd/agentcli/cmd/create.go`
- [ ] Implement basic CLI flag parsing
- [ ] Create project template engine
- [ ] Generate basic MCP-enabled project structure

### Phase 2: MCP Integration (Week 2)
- [ ] Implement MCP configuration generation
- [ ] Create MCP-aware agent templates
- [ ] Add MCP tool registry generation
- [ ] Implement basic error handling templates

### Phase 3: Advanced Features (Week 3)
- [ ] Add caching configuration generation
- [ ] Implement metrics setup templates
- [ ] Create production-ready templates
- [ ] Add load balancer configuration

### Phase 4: Interactive Mode (Week 4)
- [ ] Implement interactive CLI prompts
- [ ] Add project validation
- [ ] Create guided MCP server configuration
- [ ] Add template customization options

## Testing Strategy

### Unit Tests
- [ ] CLI flag parsing tests
- [ ] Template generation tests
- [ ] Configuration validation tests

### Integration Tests
- [ ] Generated project compilation tests
- [ ] MCP connectivity tests
- [ ] End-to-end workflow tests

### Examples Tests
- [ ] Generate and run sample projects
- [ ] Validate all MCP integration levels
- [ ] Performance benchmarking

## Success Metrics

1. **Generated Projects Work**: All templates compile and run successfully
2. **MCP Integration**: Generated projects can connect to MCP servers
3. **Performance**: Generated workflows handle 100+ req/sec with caching
4. **Usability**: Developers can create working MCP projects in <5 minutes
5. **Production Ready**: Generated code includes proper error handling, metrics, and logging

## Next Steps

1. Validate this design with existing AgentFlow patterns
2. Create the basic CLI command structure
3. Implement template generation engine
4. Build comprehensive test suite
5. Create documentation and examples

---

This design enables developers to quickly scaffold production-ready multi-agent workflows with advanced MCP capabilities, from simple tool integration to enterprise-grade distributed architectures.
