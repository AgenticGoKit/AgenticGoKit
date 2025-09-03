# Tools Basics (MCP)

Goal: Connect agents to external systems and APIs through MCP (Model Context Protocol) tools.

## Prerequisites
Complete [memory-basics.md](memory-basics.md) to understand agent state management.

## What are MCP Tools?
MCP (Model Context Protocol) allows agents to:
- Execute shell commands
- Access file systems
- Call external APIs  
- Connect to databases
- Interact with web services

MCP provides standardized tool integration across different providers.

## 1) Create a tool-enabled project
```pwsh
agentcli create tool-demo --template research-assistant
Set-Location tool-demo
```

This creates a project with common research tools pre-configured.

## 2) Configure MCP in agentflow.toml
```toml
[mcp]
enabled = true
enable_discovery = true

[[mcp.servers]]
name = "filesystem"
command = "mcp-server-filesystem"
args = ["--root", "./data"]

[[mcp.servers]]
name = "web"
command = "mcp-server-web"
args = ["--allowed-domains", "example.com,docs.company.com"]
```

Validate the configuration:
```pwsh
agentcli validate
```

Expected: Configuration validation passes without errors.

## 3) Inspect available servers and tools
```pwsh
agentcli mcp servers
agentcli mcp tools
```

Expected results:
- If MCP is not enabled: "MCP manager not configured"
- After enabling MCP: Lists show configured servers and available tools

## 4) Test tool functionality
Create test data:
```pwsh
New-Item -ItemType Directory -Path "data" -Force
"Hello from AgenticGoKit MCP!" | Out-File -FilePath "data/sample.txt"
```

Run agent with tool request:
```pwsh
go run . -m "Read the contents of data/sample.txt using available tools"
```

Expected: Agent uses MCP filesystem tool to read and return file contents.

## 5) Monitor tool usage with tracing
```pwsh
agentcli trace start
go run . -m "List files in the data directory"
agentcli trace stop
```

This captures detailed logs of MCP tool invocations and responses.

## 6) Health monitoring and debugging
```pwsh
agentcli mcp health
agentcli mcp info --server filesystem
```

These commands help diagnose MCP server connectivity and capability issues.

## Common MCP Server Types
| Server | Purpose | Command |
|--------|---------|---------|
| `filesystem` | File operations | `mcp-server-filesystem` |
| `web` | Web scraping | `mcp-server-web` |
| `postgres` | Database access | `mcp-server-postgres` |
| `github` | GitHub API | `mcp-server-github` |

## Notes
- MCP servers are external processes that need separate installation
- Configure server permissions carefully for security
- Use `agentcli trace` to observe tool usage during agent runs
- Server startup may take a few seconds before tools become available

## Next Steps
- Learn deployment strategies: [deploy-basics.md](deploy-basics.md)
- Explore advanced MCP configuration in the [MCP Guide](../../guides/CustomTools.md)

## Verification checklist
- [ ] Research-assistant template project created
- [ ] agentcli validate passed with MCP configuration
- [ ] agentcli mcp servers listed configured servers
- [ ] agentcli mcp tools showed available tools
- [ ] Agent successfully used filesystem tool to read file
- [ ] MCP health check completed without errors
