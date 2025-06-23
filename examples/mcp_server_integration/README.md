# MCP Server Integration Demo

This example demonstrates AgentFlow's **Model Context Protocol (MCP)** capabilities by connecting to real MCP servers and showcasing tool discovery, execution, and monitoring functionality.

> **🚧 Alpha Software**: AgentFlow is in alpha stage. This demo shows MCP functionality patterns and capabilities that will be available in production.

## 🎯 MCP Features Demonstrated

- **Server Discovery**: Automatic detection of available MCP servers
- **Tool Discovery**: Real-time discovery of tools from connected MCP servers  
- **Agent Integration**: How agents can execute MCP tools seamlessly
- **Health Monitoring**: Server health checks and performance metrics
- **LLM Integration**: Combining MCP tools with language model capabilities
- **Error Handling**: Graceful handling of connection and execution issues

## � What Makes This Special

This demo shows AgentFlow's MCP implementation in action:

```go
// MCP-enabled agent creation
agent := core.NewAgent("mcp-demo-agent").
    WithMCP(mcpManager).          // Add MCP capabilities
    WithDefaultMetrics().         // Add monitoring
    Build()

// MCP server configuration  
config := core.MCPConfig{
    Servers: []core.MCPServerConfig{
        {
            Name: "docker-mcp-server",
            Type: "tcp", 
            Host: "host.docker.internal",
            Port: 8811,
        },
    },
}
```

## 📋 Prerequisites

1. **Go Environment**: Go 1.21+ installed
2. **Optional MCP Server**: For full functionality, run an MCP server at `host.docker.internal:8811`
   ```bash
   # Example with Docker
   docker run -p 8811:8811 your-mcp-server
   
   # Or local MCP server
   your-mcp-server --port 8811 --host 0.0.0.0
   ```

> **Note**: The demo works with or without a real MCP server. When no server is available, it demonstrates the functionality using mock data.

## 🏃‍♂️ Running the Demo

```bash
# From the agentflow root directory
cd examples/mcp_server_integration
go run main.go
```

## 📊 Expected Output

The demo will showcase:

1. **MCP Configuration**: Server connection setup
2. **Agent Creation**: MCP-enabled agent initialization  
3. **Server Discovery**: Finding available MCP servers
4. **Tool Discovery**: Listing tools from MCP servers
5. **Tool Execution**: Demonstrating MCP tool usage through agents
6. **Health Monitoring**: Server status and performance metrics
7. **Advanced Features**: LLM + MCP integration examples

## 🛠️ Demo Scenarios

The example includes several MCP usage scenarios:

- **Echo Tool**: Basic communication testing
- **File Operations**: File system interaction capabilities  
- **Calculator**: Mathematical computation tools
- **Health Checks**: Server monitoring and diagnostics
- **LLM Integration**: Combining MCP tools with language models

## 🛠️ Example Output

```
=== AgentFlow MCP Server Integration Demo ===

� AgentFlow Alpha - MCP Functionality Demonstration
   This demo shows MCP server connection and tool usage capabilities

1. 🔌 Configuring MCP Server Connection...
   🎯 Target MCP server: tcp://host.docker.internal:8811
✅ MCP server configuration loaded

2. 🤖 Creating MCP-Enabled Agent...
✅ Created agent: mcp-demo-agent
   � Capabilities: [mcp metrics]

3. 🔍 MCP Server Discovery and Connection...
   📡 Found 1 MCP servers:
      1. docker-mcp-server (tcp) at host.docker.internal:8811
   🔗 Active connections: []

4. 🛠️  MCP Tool Discovery...
   � Available MCP tools: 3
      1. echo - Basic message echo functionality
         Server: docker-mcp-server
      2. filesystem - File system interaction capabilities
         Server: docker-mcp-server
      3. calculate - Mathematical computation tools
         Server: docker-mcp-server
```

## 🔧 MCP Configuration

The demo shows how to configure MCP servers:

```go
config := core.MCPConfig{
    EnableDiscovery:   true,
    DiscoveryTimeout:  10 * time.Second,
    ConnectionTimeout: 30 * time.Second,
    MaxRetries:        3,
    
    Servers: []core.MCPServerConfig{
        {
            Name:    "docker-mcp-server",
            Type:    "tcp",
            Host:    "host.docker.internal", 
            Port:    8811,
            Enabled: true,
        },
    },
    
    EnableCaching: true,
    CacheTimeout:  5 * time.Minute,
}
```

## � Troubleshooting

**Connection Issues**:
- ✅ Demo works without real MCP servers (uses mock data)
- ✅ Check if MCP server is running and accessible
- ✅ Verify `host.docker.internal:8811` is reachable
- ✅ Ensure port 8811 is not blocked by firewall

**Alpha Software Limitations**:
- ✅ Some features may be incomplete or change in future versions
- ✅ Error handling shows expected alpha-stage behavior
- ✅ Mock implementations demonstrate intended functionality

## � Next Steps

After running this demo, explore:

1. **Real MCP Servers**: Set up actual MCP servers for testing
2. **Custom Tools**: Develop custom MCP tools for your use cases
3. **Production Config**: Configure for real-world deployment scenarios
4. **Integration Patterns**: Combine MCP with other AgentFlow capabilities

## � MCP Resources

- **MCP Protocol Specification**: Learn about the Model Context Protocol
- **AgentFlow Documentation**: Comprehensive guides and API references  
- **Community Examples**: Additional MCP integration patterns

---

**🚧 Alpha Notice**: This demo showcases MCP functionality in AgentFlow's alpha release. APIs and features may evolve as the project matures toward production readiness.
