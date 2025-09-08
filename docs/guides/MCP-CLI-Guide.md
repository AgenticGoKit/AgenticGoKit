# MCP CLI Integration Guide

**Complete guide to using MCP (Model Context Protocol) with AgentCLI**

This guide covers how to use the AgentCLI to create projects with MCP integration, configure MCP servers, and work with external tools.

## 🚀 Quick Start

### Enable MCP with Minimal Configuration

The simplest way to enable MCP in your project:

```bash
agentcli create my-project --enable-mcp
```

This creates a project with:
- ✅ MCP capability enabled
- ✅ Commented examples of popular MCP servers
- ✅ Basic configuration ready to use

### MCP Integration Levels

AgentCLI provides three levels of MCP integration:

#### `--mcp minimal`
Basic MCP support with minimal overhead:
```bash
agentcli create my-project --mcp minimal
```
- MCP protocol enabled
- No additional features
- Suitable for development and simple use cases

#### `--mcp standard` 
MCP with performance optimizations:
```bash
agentcli create my-project --mcp standard
```
- MCP protocol enabled
- **Tool result caching** for faster repeated operations
- Suitable for production applications

#### `--mcp advanced`
Full-featured MCP with enterprise capabilities:
```bash
agentcli create my-project --mcp advanced
```
- MCP protocol enabled
- **Tool result caching** for performance
- **Metrics and monitoring** for observability
- **Load balancing** for high availability
- Suitable for production systems with high load

## 📋 Flag Comparison

| Flag | MCP Enabled | Caching | Metrics | Load Balancing | Use Case |
|------|-------------|---------|---------|---------------|----------|
| `--enable-mcp` | ✅ | ❌ | ❌ | ❌ | Development, Learning |
| `--mcp minimal` | ✅ | ❌ | ❌ | ❌ | Simple Applications |
| `--mcp standard` | ✅ | ✅ | ❌ | ❌ | Production Apps |
| `--mcp advanced` | ✅ | ✅ | ✅ | ✅ | Enterprise Systems |

## 🔧 Configuration Examples

### Generated agentflow.toml

When you create a project with MCP enabled, AgentCLI generates an `agentflow.toml` file with comprehensive examples:

```toml
[mcp]
enabled = true
transport = "tcp"
enable_discovery = true
connection_timeout = 5000
max_retries = 3
retry_delay = 1000
enable_caching = true
cache_timeout = 300000
max_connections = 10

# MCP Server Examples - Uncomment and configure as needed
#
# For Docker AI Gateway (provides web search, content fetching, etc.)
# Start with: docker run -p 8812:8812 -p 8813:8813 your-docker-image

[[mcp.servers]]
name = "docker-http-sse"
type = "http_sse"
host = "localhost"
port = 8812
enabled = false

[[mcp.servers]]
name = "docker-http-streaming"
type = "http_streaming"
host = "localhost"
port = 8813
enabled = false

# For file system access
# Install with: npm install -g @modelcontextprotocol/server-filesystem
# [[mcp.servers]]
# name = "filesystem"
# type = "stdio"
# command = "npx @modelcontextprotocol/server-filesystem /path/to/allowed/files"
# enabled = true

# For web search capabilities
# Install with: npm install -g @modelcontextprotocol/server-brave-search
# [[mcp.servers]]
# name = "brave-search"
# type = "stdio"
# command = "npx @modelcontextprotocol/server-brave-search"
# enabled = true
```

### Enabling Specific Tools

To enable specific MCP tools, simply uncomment and configure the relevant sections:

#### 1. Enable File System Access
```toml
[[mcp.servers]]
name = "filesystem"
type = "stdio"
command = "npx @modelcontextprotocol/server-filesystem /home/user/documents"
enabled = true
```

#### 2. Enable Web Search
```toml
[[mcp.servers]]
name = "brave-search"
type = "stdio"
command = "npx @modelcontextprotocol/server-brave-search"
enabled = true
```

#### 3. Enable Database Access
```toml
[[mcp.servers]]
name = "sqlite"
type = "stdio"
command = "npx @modelcontextprotocol/server-sqlite /path/to/database.db"
enabled = true
```

## 🛠 Popular MCP Servers

### Pre-built MCP Servers

| Server | Installation | Description | Tools Provided |
|--------|-------------|-------------|----------------|
| **filesystem** | `npm install -g @modelcontextprotocol/server-filesystem` | File operations | read_file, write_file, list_directory |
| **brave-search** | `npm install -g @modelcontextprotocol/server-brave-search` | Web search | web_search, search_images |
| **sqlite** | `npm install -g @modelcontextprotocol/server-sqlite` | Database queries | execute_query, get_schema |
| **github** | `npm install -g @modelcontextprotocol/server-github` | GitHub integration | create_issue, list_repos, get_file |
| **google-maps** | `npm install -g @modelcontextprotocol/server-google-maps` | Location services | geocode, directions, places |

### Docker AI Gateway

For web search, content fetching, and other online tools:

```bash
# Start Docker AI Gateway
docker run -p 8812:8812 -p 8813:8813 your-docker-image

# Enable in agentflow.toml
[[mcp.servers]]
name = "docker-http-sse"
type = "http_sse"
host = "localhost"
port = 8812
enabled = true
```

## 🎯 Complete Examples

### Research Assistant with Web Search

```bash
# Create research assistant with advanced MCP
agentcli create research-bot --template research-assistant --mcp advanced
```

Then enable web search in `agentflow.toml`:
```toml
[[mcp.servers]]
name = "brave-search"
type = "stdio"
command = "npx @modelcontextprotocol/server-brave-search"
enabled = true
```

### Document Processing with File Access

```bash
# Create document processor with standard MCP
agentcli create doc-processor --mcp standard
```

Then enable file system access in `agentflow.toml`:
```toml
[[mcp.servers]]
name = "filesystem"
type = "stdio"
command = "npx @modelcontextprotocol/server-filesystem /home/user/documents"
enabled = true
```

### Multi-Tool Enterprise System

```bash
# Create enterprise system with full MCP features
agentcli create enterprise-ai --mcp advanced --memory pgvector --rag 1000
```

Then enable multiple tools in `agentflow.toml`:
```toml
# Web search
[[mcp.servers]]
name = "brave-search"
type = "stdio"
command = "npx @modelcontextprotocol/server-brave-search"
enabled = true

# Database access
[[mcp.servers]]
name = "sqlite"
type = "stdio"
command = "npx @modelcontextprotocol/server-sqlite /data/app.db"
enabled = true

# File system
[[mcp.servers]]
name = "filesystem"
type = "stdio"
command = "npx @modelcontextprotocol/server-filesystem /data/files"
enabled = true
```

## 🔍 Testing Your MCP Setup

After creating your project and configuring MCP servers:

1. **Install required MCP servers**:
   ```bash
   npm install -g @modelcontextprotocol/server-brave-search
   npm install -g @modelcontextprotocol/server-filesystem
   ```

2. **Run your project**:
   ```bash
   cd your-project
   go mod tidy
   go run . -m "Search for information about artificial intelligence"
   ```

3. **Check MCP tool discovery**:
   Look for debug output like:
   ```
   DBG MCP tools discovered agent=agent1 tool_count=8
   DBG UnifiedAgent: Tool calls detected, executing agent=agent1 tool_calls=1
   ```

## 🚨 Troubleshooting

### Common Issues

#### "No MCP tools available"
- Ensure MCP servers are running and accessible
- Check `enabled = true` in server configuration
- Verify correct host/port for HTTP servers
- Check command paths for STDIO servers

#### "Tool execution failed"
- Install required MCP server packages
- Set necessary environment variables (e.g., API keys)
- Check file permissions for STDIO servers
- Verify network connectivity for HTTP servers

#### "Connection timeout"
- Increase `connection_timeout` in MCP configuration
- Check if MCP server is responsive
- Verify firewall settings for HTTP connections

### Debug Mode

Enable detailed MCP logging by setting log level to debug in `agentflow.toml`:

```toml
[logging]
level = "debug"
format = "console"
```

## 📚 Next Steps

- [MCP Server Development](../tutorials/mcp/tool-development.md) - Create custom MCP tools
- [Advanced MCP Patterns](../tutorials/mcp/advanced-tool-patterns.md) - Complex integrations
- [MCP API Reference](../reference/api/mcp.md) - Programmatic usage
- [Configuration Reference](../reference/api/configuration.md) - Complete config options

## 🔗 Related Links

- [Model Context Protocol Specification](https://spec.modelcontextprotocol.io/)
- [Official MCP Servers](https://github.com/modelcontextprotocol/servers)
- [AgenticGoKit MCP Examples](../tutorials/mcp/README.md)
