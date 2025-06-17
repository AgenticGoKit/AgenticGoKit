# MCP Integration Plan for AgentFlow

**Document Version**: 1.0  
**Created**: June 16, 2025  
**Status**: Planning Phase  

## Overview

This document outlines the implementation plan for integrating Model Context Protocol (MCP) tooling support into AgentFlow using the [mcp-navigator-go](https://github.com/kunalkushwaha/mcp-navigator-go) library. The integration will enable AgentFlow agents to discover, connect to, and utilize MCP servers and their tools seamlessly.

## Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   AgentFlow     │    │  MCP Navigator  │    │   MCP Servers   │
│     Agent       │───▶│    Library      │───▶│   (Tools,       │
│                 │    │                 │    │   Resources)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │              ┌─────────────────┐              │
         └─────────────▶│  Tool Registry  │◀─────────────┘
                        │   (Unified)     │
                        └─────────────────┘
```

## Implementation Phases

### Phase 1: Foundation Integration
**Duration**: 2 weeks  
**Priority**: High  

#### Task 1.1: Add MCP Navigator Dependency
- [ ] Add `github.com/kunalkushwaha/mcp-navigator-go` to go.mod
- [ ] Test basic import and compilation
- [ ] Update dependencies documentation

**Files to modify**:
- `go.mod`
- `go.sum` (auto-generated)
- `docs/DevGuide.md`

**Acceptance Criteria**:
- [x] MCP Navigator library successfully imported
- [x] No compilation errors
- [x] All existing tests pass

#### Task 1.2: Create MCP Tool Adapter
- [ ] Create `core/mcp_tool.go` implementing `FunctionTool` interface
- [ ] Handle MCP tool schema conversion
- [ ] Add error handling for MCP tool execution
- [ ] Create unit tests

**Files to create**:
- `core/mcp_tool.go`
- `core/mcp_tool_test.go`

**Implementation Details**:
```go
type MCPTool struct {
    name        string
    description string
    client      *mcpclient.Client
    schema      map[string]interface{}
    serverName  string
}

func (m *MCPTool) Name() string
func (m *MCPTool) Call(ctx context.Context, args map[string]any) (map[string]any, error)
func NewMCPTool(client *mcpclient.Client, toolInfo mcp.Tool, serverName string) *MCPTool
```

**Acceptance Criteria**:
- [x] MCPTool implements FunctionTool interface
- [x] Proper error handling and logging
- [x] Unit tests with >80% coverage
- [x] Integration with existing tool registry

#### Task 1.3: Create MCP Connection Manager
- [ ] Create `core/mcp_manager.go` for managing MCP connections
- [ ] Implement server discovery and connection pooling
- [ ] Add configuration management
- [ ] Create comprehensive tests

**Files to create**:
- `core/mcp_manager.go`
- `core/mcp_manager_test.go`

**Implementation Details**:
```go
type MCPManager struct {
    clients     map[string]*mcpclient.Client
    discovery   *discovery.Discovery
    registry    *tools.ToolRegistry
    config      MCPConfig
    mu          sync.RWMutex
}

func NewMCPManager(config MCPConfig, registry *tools.ToolRegistry) *MCPManager
func (m *MCPManager) DiscoverAndConnect(ctx context.Context) error
func (m *MCPManager) RegisterMCPTools(serverName string) error
func (m *MCPManager) GetMCPClient(serverName string) *mcpclient.Client
func (m *MCPManager) Disconnect() error
```

**Acceptance Criteria**:
- [x] Thread-safe connection management
- [x] Automatic server discovery
- [x] Graceful error handling
- [x] Comprehensive test coverage

#### Task 1.4: Configuration Integration
- [ ] Extend configuration system to support MCP settings
- [ ] Update `agentflow.toml` schema
- [ ] Add validation for MCP configuration
- [ ] Update configuration documentation

**Files to modify**:
- `core/config.go`
- `core/config_test.go`
- `examples/` (configuration examples)

**Configuration Schema**:
```toml
[mcp]
auto_discover = true
discovery_timeout = "10s"
connection_timeout = "30s"
max_connections = 10

[[mcp.servers]]
name = "web_search"
type = "tcp"
host = "localhost"
port = 8811
enabled = true

[[mcp.servers]]
name = "file_operations"
type = "docker"
container = "mcp-file-server"
enabled = true
```

**Acceptance Criteria**:
- [x] Configuration properly parsed and validated
- [x] Default values work correctly
- [x] Error messages are helpful
- [x] Documentation updated

---

### Phase 2: Core Integration
**Duration**: 2 weeks  
**Priority**: High  

#### Task 2.1: Extend AgentFlow Factory
- [ ] Add MCP-enabled factory functions to `core/factory.go`
- [ ] Create builder patterns for MCP agents
- [ ] Update existing factory tests
- [ ] Add integration examples

**Files to modify**:
- `core/factory.go`
- `core/factory_test.go` (if exists)

**Implementation Details**:
```go
type MCPAgentConfig struct {
    MCPServers       []MCPServerConfig
    AutoDiscover     bool
    DiscoveryTimeout time.Duration
    ToolFilters      []string
}

func NewMCPEnabledRunner(config MCPAgentConfig) (Runner, error)
func NewMCPToolAgent(mcpManager *MCPManager, llmProvider ModelProvider) Agent
```

**Acceptance Criteria**:
- [x] Factory functions create working MCP agents
- [x] Configuration validation works
- [x] Integration with existing agent system
- [x] Comprehensive examples

#### Task 2.2: Create MCP-Aware Agent
- [ ] Create `core/mcp_agent.go` implementing `Agent` interface
- [ ] Add LLM integration for tool selection
- [ ] Implement tool execution workflow
- [ ] Add comprehensive error handling

**Files to create**:
- `core/mcp_agent.go`
- `core/mcp_agent_test.go`

**Implementation Details**:
```go
type MCPAgent struct {
    name        string
    mcpManager  *MCPManager
    llmProvider ModelProvider
    registry    *tools.ToolRegistry
    config      MCPAgentConfig
}

func (a *MCPAgent) Name() string
func (a *MCPAgent) Run(ctx context.Context, inputState State) (State, error)
func NewMCPAgent(name string, mcpManager *MCPManager, llmProvider ModelProvider) *MCPAgent
```

**Workflow**:
1. Analyze input state to determine tool requirements
2. Query available MCP tools
3. Use LLM to select appropriate tools
4. Execute tools with proper parameters
5. Aggregate results into output state

**Acceptance Criteria**:
- [x] Agent implements core Agent interface
- [x] Intelligent tool selection via LLM
- [x] Proper state management
- [x] Error recovery mechanisms

#### Task 2.3: Update Tool Registry Integration
- [ ] Modify `internal/factory/agent_factory.go` to include MCP tools
- [ ] Update tool registration workflow
- [ ] Add MCP tool discovery to default registry
- [ ] Create migration guide

**Files to modify**:
- `internal/factory/agent_factory.go`
- `internal/tools/tool.go` (if needed)

**Implementation Details**:
```go
func NewDefaultToolRegistryWithMCP(mcpConfig MCPConfig) (*tools.ToolRegistry, error)
func RegisterMCPServers(registry *tools.ToolRegistry, mcpManager *MCPManager) error
```

**Acceptance Criteria**:
- [x] MCP tools appear in unified registry
- [x] No conflicts with existing tools
- [x] Backwards compatibility maintained
- [x] Clear documentation

---

### Phase 3: Advanced Features
**Duration**: 2 weeks  
**Priority**: Medium  

#### Task 3.1: LLM-MCP Integration
- [ ] Create intelligent MCP tool selector
- [ ] Implement context-aware tool recommendations
- [ ] Add tool result interpretation
- [ ] Create performance optimization

**Files to create**:
- `core/mcp_tool_selector.go`
- `core/mcp_tool_selector_test.go`

**Implementation Details**:
```go
type MCPToolSelector struct {
    llmProvider ModelProvider
    mcpManager  *MCPManager
    cache       *ToolSelectionCache
}

func (s *MCPToolSelector) SelectTools(ctx context.Context, query string, context State) ([]string, error)
func (s *MCPToolSelector) ExecuteToolChain(ctx context.Context, tools []string, params map[string]any) ([]ToolResult, error)
```

**Acceptance Criteria**:
- [x] Intelligent tool selection based on context
- [x] Performance caching mechanisms
- [x] Error handling and fallbacks
- [x] Metrics and observability

#### Task 3.2: MCP Resource Integration
- [ ] Extend to support MCP Resources
- [ ] Create resource loading agents
- [ ] Add resource caching and management
- [ ] Implement resource workflow patterns

**Files to create**:
- `core/mcp_resource_agent.go`
- `core/mcp_resource_agent_test.go`

**Implementation Details**:
```go
type MCPResourceAgent struct {
    name       string
    mcpManager *MCPManager
    cache      *ResourceCache
}

func (a *MCPResourceAgent) LoadResource(ctx context.Context, uri string) (State, error)
func (a *MCPResourceAgent) ListResources(ctx context.Context, serverName string) ([]Resource, error)
```

**Acceptance Criteria**:
- [x] Resource loading and caching
- [x] Integration with agent workflows
- [x] Performance optimization
- [x] Error handling

#### Task 3.3: Streaming and Real-time Updates
- [ ] Add support for MCP streaming
- [ ] Implement real-time notifications
- [ ] Create streaming agent patterns
- [ ] Add WebSocket transport support

**Files to create**:
- `core/mcp_streaming_agent.go`
- `core/mcp_streaming_agent_test.go`

**Implementation Details**:
```go
type MCPStreamingAgent struct {
    name       string
    mcpManager *MCPManager
    eventBus   chan Event
    subscribers map[string]chan ToolUpdate
}

func (a *MCPStreamingAgent) StreamToolOutput(ctx context.Context, toolName string, args map[string]any) (<-chan ToolUpdate, error)
func (a *MCPStreamingAgent) Subscribe(ctx context.Context, eventType string) (<-chan Event, error)
```

**Acceptance Criteria**:
- [x] Real-time streaming support
- [x] Event-driven architecture
- [x] Scalable subscription model
- [x] Resource management

---

### Phase 4: CLI and Developer Experience
**Duration**: 2 weeks  
**Priority**: Medium  

#### Task 4.1: Extend AgentFlow CLI
- [ ] Add MCP commands to `cmd/agentcli/`
- [ ] Create discovery and connection commands
- [ ] Add tool listing and testing commands
- [ ] Implement interactive MCP shell

**Files to create**:
- `cmd/agentcli/cmd/mcp.go`
- `cmd/agentcli/cmd/mcp_discover.go`
- `cmd/agentcli/cmd/mcp_connect.go`
- `cmd/agentcli/cmd/mcp_tools.go`

**Command Structure**:
```bash
agentcli mcp discover                           # Discover MCP servers
agentcli mcp connect --server localhost:8811   # Test MCP connection
agentcli mcp tools --server web_search         # List available tools
agentcli mcp test-tool --name search --args '{...}'  # Test specific tool
agentcli mcp interactive                        # Interactive MCP shell
```

**Acceptance Criteria**:
- [x] All commands work correctly
- [x] Helpful error messages
- [x] Comprehensive help documentation
- [x] Integration with existing CLI

#### Task 4.2: MCP Agent Templates
- [ ] Create project templates in `internal/scaffold/`
- [ ] Add MCP-enabled agent scaffolding
- [ ] Create example configurations
- [ ] Update project generation

**Files to create**:
- `internal/scaffold/mcp_agent_template.go`
- `internal/scaffold/templates/mcp/`

**Template Features**:
- MCP server configuration
- Agent implementation templates
- Tool integration examples
- Testing frameworks

**Acceptance Criteria**:
- [x] Templates generate working projects
- [x] Good documentation and examples
- [x] Integration with `agentcli create`
- [x] Multiple template variants

#### Task 4.3: Documentation and Examples
- [ ] Update documentation in `docs/`
- [ ] Create comprehensive examples in `examples/mcp/`
- [ ] Add integration guides
- [ ] Create troubleshooting documentation

**Files to create/modify**:
- `docs/MCP_Integration_Guide.md`
- `docs/MCP_Configuration.md`
- `docs/MCP_Troubleshooting.md`
- `examples/mcp/basic_mcp_agent/`
- `examples/mcp/multi_server_agent/`
- `examples/mcp/streaming_agent/`

**Documentation Topics**:
- Quick start guide
- Configuration reference
- API documentation
- Best practices
- Troubleshooting guide

**Acceptance Criteria**:
- [x] Comprehensive documentation
- [x] Working examples
- [x] Clear tutorials
- [x] Troubleshooting guides

---

### Phase 5: Testing and Optimization
**Duration**: 2 weeks  
**Priority**: High  

#### Task 5.1: Integration Tests
- [ ] Create comprehensive integration test suite
- [ ] Add MCP server mocking for tests
- [ ] Create performance benchmarks
- [ ] Add CI/CD integration

**Files to create**:
- `integration/mcp_integration_test.go`
- `integration/mcp_mock_server.go`
- `benchmarks/mcp_performance_test.go`

**Test Categories**:
- Connection management
- Tool discovery and execution
- Error handling and recovery
- Performance and scalability
- Configuration validation

**Acceptance Criteria**:
- [x] >90% test coverage for MCP code
- [x] Integration tests pass consistently
- [x] Performance benchmarks established
- [x] CI/CD pipeline updated

#### Task 5.2: Performance Optimization
- [ ] Implement connection pooling for MCP clients
- [ ] Add caching for tool schemas and results
- [ ] Optimize async tool execution
- [ ] Add performance monitoring

**Files to modify**:
- `core/mcp_manager.go`
- `core/mcp_tool.go`
- Add performance monitoring utilities

**Optimization Areas**:
- Connection reuse and pooling
- Schema caching
- Result caching with TTL
- Async execution patterns
- Memory management

**Acceptance Criteria**:
- [x] 50% improvement in connection overhead
- [x] Effective caching reduces redundant calls
- [x] Memory usage within acceptable bounds
- [x] Performance metrics collection

#### Task 5.3: Error Handling and Resilience
- [ ] Implement circuit breaker patterns
- [ ] Add MCP server failover mechanisms
- [ ] Create graceful degradation
- [ ] Add comprehensive logging and monitoring

**Files to create**:
- `core/mcp_circuit_breaker.go`
- `core/mcp_failover.go`

**Resilience Features**:
- Circuit breaker for unreliable servers
- Automatic failover to backup servers
- Graceful degradation when MCP unavailable
- Comprehensive error reporting
- Health check mechanisms

**Acceptance Criteria**:
- [x] System remains stable with unreliable MCP servers
- [x] Automatic recovery mechanisms work
- [x] Comprehensive error reporting
- [x] Monitoring and alerting

---

## Implementation Schedule

| Week | Phase | Tasks | Deliverables |
|------|-------|-------|--------------|
| 1-2  | Phase 1 | Foundation | MCP integration foundation, basic tool support |
| 3-4  | Phase 2 | Core | MCP agents, factory integration, core workflows |
| 5-6  | Phase 3 | Advanced | LLM integration, resources, streaming |
| 7-8  | Phase 4 | CLI/DX | CLI commands, templates, documentation |
| 9-10 | Phase 5 | Testing | Comprehensive testing, optimization, resilience |

## Success Metrics

### Technical Metrics
- [ ] 100% of existing AgentFlow functionality preserved
- [ ] <100ms average latency for MCP tool calls
- [ ] >99.9% uptime with proper error handling
- [ ] >90% test coverage for all MCP code

### Feature Metrics
- [ ] Automatic discovery of MCP servers
- [ ] Seamless integration with existing agents
- [ ] Support for all major MCP transport types
- [ ] Real-time streaming and notifications

### Developer Experience Metrics
- [ ] <5 minutes to set up MCP-enabled agent
- [ ] Comprehensive documentation and examples
- [ ] Clear error messages and troubleshooting
- [ ] CLI tools for testing and debugging

## Risk Mitigation

### Technical Risks
- **MCP library compatibility**: Pin to specific version, thorough testing
- **Performance impact**: Comprehensive benchmarking, optimization
- **Memory leaks**: Proper connection management, monitoring

### Integration Risks
- **Breaking changes**: Comprehensive backward compatibility testing
- **Configuration complexity**: Good defaults, validation, documentation
- **Learning curve**: Excellent documentation, examples, tutorials

## Dependencies

### External Dependencies
- `github.com/kunalkushwaha/mcp-navigator-go` - MCP client library
- MCP servers for testing and development
- Updated Go version if required

### Internal Dependencies
- Existing tool registry system
- Configuration management
- Agent framework
- CLI infrastructure

## Deliverables

### Code Deliverables
- [ ] MCP integration library (`core/mcp_*.go`)
- [ ] Updated tool registry and factory
- [ ] CLI extensions
- [ ] Comprehensive test suite

### Documentation Deliverables
- [ ] Integration guide
- [ ] API documentation
- [ ] Configuration reference
- [ ] Troubleshooting guide
- [ ] Examples and tutorials

### Infrastructure Deliverables
- [ ] CI/CD pipeline updates
- [ ] Performance benchmarks
- [ ] Monitoring and logging
- [ ] Template projects

---

## Getting Started

To begin implementation:

1. **Set up development environment**:
   ```bash
   git checkout -b feature/mcp-integration
   go get github.com/kunalkushwaha/mcp-navigator-go@latest
   ```

2. **Start with Phase 1, Task 1.1**:
   - Add dependency
   - Verify compilation
   - Update documentation

3. **Follow the task checklist** in order, ensuring each task's acceptance criteria are met before proceeding

4. **Regular testing**: Run the full test suite after each major change

5. **Documentation**: Update docs as you go, don't leave it until the end

## Notes

- This plan assumes familiarity with both AgentFlow and MCP protocol
- Tasks can be parallelized where dependencies allow
- Regular code reviews recommended for each phase
- Consider creating feature flags for gradual rollout
- Monitor performance impact throughout implementation

---

**Last Updated**: June 16, 2025  
**Next Review**: TBD based on implementation progress
