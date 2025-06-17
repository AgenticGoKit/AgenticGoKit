# MCP Integration Example

This example demonstrates how to integrate Model Context Protocol (MCP) tooling into AgentFlow using the `mcp-navigator-go` library.

## What This Example Shows

- How to configure MCP servers (TCP, STDIO, WebSocket)
- How to create and initialize an MCP manager
- How to discover and connect to MCP servers
- How to use MCP tools as first-class AgentFlow tools
- How to monitor MCP connections and performance

## Prerequisites

To run this example with actual MCP servers, you need:

1. An MCP server running on TCP (e.g., port 8080)
2. Or a command that starts an MCP server via STDIO (e.g., Node.js script)

## Running the Example

```bash
# Basic example (will show connection attempts but no actual servers)
go run main.go

# Or build and run
go build -o mcp-example main.go
./mcp-example
```

## Configuration

The example uses the default MCP configuration which includes:

- **Server Discovery**: Enabled (scans common ports)
- **Connection Timeout**: 30 seconds
- **Retry Logic**: 3 retries with 1 second delay
- **Caching**: Enabled with 5 minute timeout

You can customize this by modifying the `config` variable in `main.go`.

## Sample MCP Servers

### TCP Server Example

```go
tcpServer, err := core.NewTCPServerConfig("my-server", "localhost", 8080)
if err == nil {
    config.Servers = append(config.Servers, tcpServer)
}
```

### STDIO Server Example

```go
stdioServer, err := core.NewSTDIOServerConfig("node-tools", "node mcp-server.js")
if err == nil {
    config.Servers = append(config.Servers, stdioServer)
}
```

### WebSocket Server Example

```go
wsServer, err := core.NewWebSocketServerConfig("ws-server", "localhost", 8080)
if err == nil {
    config.Servers = append(config.Servers, wsServer)
}
```

## Architecture

The MCP integration follows AgentFlow's architecture principles:

- **Public API**: Clean interfaces in `core/mcp.go`
- **Implementation**: Concrete types in `internal/mcp/`
- **Tool Integration**: MCP tools implement the `FunctionTool` interface
- **Connection Management**: Automatic discovery, connection pooling, health checks

## Key Components

### MCPManager
- Manages connections to multiple MCP servers
- Handles tool discovery and registration
- Provides health monitoring and metrics

### MCPTool
- Adapts MCP tools to AgentFlow's `FunctionTool` interface
- Handles argument validation and response conversion
- Provides proper error handling and logging

### Configuration
- Flexible server configuration (TCP, STDIO, WebSocket)
- Automatic defaults for timeouts and retries
- Discovery settings and connection limits

## Next Steps

1. **Set up an MCP server** (see [MCP documentation](https://spec.modelcontextprotocol.io/))
2. **Configure the server** in your AgentFlow application
3. **Use MCP tools** in your agents like any other AgentFlow tool
4. **Monitor performance** using the built-in metrics and health checks

For more information, see:
- [MCP Integration Plan](../../docs/MCP_Integration_Plan.md)
- [MCP Technical Specification](../../docs/MCP_Technical_Specification.md)
- [MCP Quick Reference](../../docs/MCP_Quick_Reference.md)
