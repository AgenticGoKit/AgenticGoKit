# CLI Reference

**Complete reference for the AgenticGoKit command-line interface**

This document provides comprehensive reference for AgenticGoKit's command-line interface (`agentcli`), covering all commands, options, and usage patterns.

## 🏗️ Installation and Setup

### Installation

```bash
# Install from source
go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest

# Or download binary from releases
curl -L https://github.com/kunalkushwaha/agenticgokit/releases/latest/download/agentcli-${OS}-${ARCH}.tar.gz | tar xz
```

## 📋 Command Structure

```
agentcli [global options] command [command options] [arguments...]
```

### Global Options

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--config` | `-c` | Path to configuration file | `agentflow.toml` |
| `--verbose` | `-v` | Enable verbose output | `false` |
| `--quiet` | `-q` | Suppress non-error output | `false` |
| `--help` | `-h` | Show help information | |
| `--version` |  | Show version information | |

## 🚀 Available Commands

Based on the actual codebase, the following commands are available:

### `trace`
View execution traces with detailed agent flows

```bash
# View a basic trace with all details
agentcli trace <session-id>

# View only the agent flow for a session
agentcli trace --flow-only <session-id>

# Filter trace to see only a specific agent's activity
agentcli trace --filter <agent-name> <session-id>
```

### `mcp`
Manage Model Context Protocol servers and tools

```bash
# List connected MCP servers
agentcli mcp servers

# View MCP server status
agentcli mcp status
```

### `cache`
Monitor and optimize MCP tool result caches

```bash
# View cache statistics
agentcli cache stats

# Clear specific caches
agentcli cache clear --server web-service
```

### `create`
Create new AgenticGoKit projects

```bash
# Create a new project
agentcli create my-project
```

### `list`
List various resources

```bash
# List available resources
agentcli list
```

### `memory`
Memory operations and management

```bash
# Memory operations
agentcli memory
```

## 📚 Usage Examples

### Tracing and Debugging

```bash
# View execution traces
agentcli trace session-123

# Filter traces by agent
agentcli trace --filter chat-agent session-123

# View agent flow only
agentcli trace --flow-only session-123
```

### MCP Management

```bash
# List MCP servers
agentcli mcp servers

# Check MCP status
agentcli mcp status
```

### Cache Management

```bash
# View cache statistics
agentcli cache stats

# Clear cache for specific server
agentcli cache clear --server search-service
```

## 🔧 Configuration

The CLI uses the same `agentflow.toml` configuration file as the main framework. See the [Configuration API](api/configuration.md) for complete configuration options.

## 📝 Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | Network/connectivity error |

For complete CLI documentation and all available options, run:

```bash
agentcli --help
agentcli <command> --help
```