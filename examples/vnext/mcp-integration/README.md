# vnext MCP Integration Example

This example demonstrates how to use Model Context Protocol (MCP) with vnext agents.

## Features

- **Explicit MCP Server Connection**: Connect to specific MCP servers
- **Automatic MCP Discovery**: Discover MCP servers on specified ports
- **Tool Integration**: MCP tools are automatically available to the agent
- **Agentic Loop**: Agent can use MCP tools autonomously

## Prerequisites

1. **Ollama** with `gemma3:1b` model:
   ```bash
   ollama pull gemma3:1b
   ```

2. **MCP Server** (optional for Example 1):
   - Docker MCP server running on `localhost:8811`, or
   - Any other MCP server (update configuration accordingly)

## Running the Example

```bash
go run main.go
```

## Configuration

### Example 1: Explicit MCP Server

Connects to a specific MCP server:

```go
mcpServer := vnext.MCPServer{
    Name:    "docker-mcp",
    Type:    "tcp",
    Address: "localhost",
    Port:    8811,
    Enabled: true,
}

agent, _ := vnext.NewBuilder("mcp-agent").
    WithTools(vnext.WithMCP(mcpServer)).
    Build()
```

### Example 2: MCP Discovery

Automatically discovers MCP servers on specified ports:

```go
agent, _ := vnext.NewBuilder("discovery-agent").
    WithTools(
        vnext.WithMCPDiscovery(8080, 8081, 8090, 8100, 8811),
    ).
    Build()
```

## MCP Server Types

vnext supports multiple MCP transport types:

- **TCP**: `Type: "tcp"` - TCP socket connection
- **STDIO**: `Type: "stdio"` - Standard input/output
- **WebSocket**: `Type: "websocket"` - WebSocket connection
- **HTTP SSE**: `Type: "http_sse"` - HTTP Server-Sent Events
- **HTTP Streaming**: `Type: "http_streaming"` - HTTP streaming

## Expected Output

When tools are discovered:

```
=== vnext MCP Integration Example ===

Example 1: Agent with Explicit MCP Server
✓ Agent created with MCP server
  Server: docker-mcp (localhost:8811)

📊 Result:
  Response: I have access to several tools: echo, docker_ps, docker_images...
  Duration: 1.2s
  Tools Used: [list_tools]
```

## Advanced Configuration

### With Caching

```go
agent, _ := vnext.NewBuilder("mcp-agent").
    WithTools(
        vnext.WithMCP(mcpServer),
        vnext.WithToolCaching(5*time.Minute),
    ).
    Build()
```

### With Timeout

```go
agent, _ := vnext.NewBuilder("mcp-agent").
    WithTools(
        vnext.WithMCP(mcpServer),
        vnext.WithToolTimeout(30*time.Second),
    ).
    Build()
```

### Multiple Servers

```go
server1 := vnext.MCPServer{Name: "docker", Type: "tcp", Address: "localhost", Port: 8811, Enabled: true}
server2 := vnext.MCPServer{Name: "web", Type: "tcp", Address: "localhost", Port: 8812, Enabled: true}

agent, _ := vnext.NewBuilder("multi-mcp-agent").
    WithTools(vnext.WithMCP(server1, server2)).
    Build()
```

## Troubleshooting

**MCP server not found:**
- Ensure the MCP server is running
- Check the port and address are correct
- Verify firewall settings

**No tools discovered:**
- Check MCP server logs
- Verify the server is responding to tool discovery requests
- Try enabling debug mode: `DebugMode: true` in Config

**Tools not executing:**
- Check tool arguments match schema
- Verify LLM is generating proper TOOL_CALL syntax
- Review agent logs for errors

## Learn More

- [vnext README](../../../core/vnext/README.md)
- [MCP Documentation](../../../docs/tutorials/mcp/README.md)
- [Tool Discovery](../../../core/vnext/tool_discovery.go)
