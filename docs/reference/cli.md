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

### Shell Completion

AgenticGoKit CLI supports intelligent tab completion for all major shells. This provides faster command usage and reduces typing errors.

#### Bash

```bash
# Load completion for current session
source <(agentcli completion bash)

# Install permanently on Linux
sudo agentcli completion bash > /etc/bash_completion.d/agentcli

# Install permanently on macOS (with Homebrew)
agentcli completion bash > $(brew --prefix)/etc/bash_completion.d/agentcli
```

#### Zsh

```bash
# Enable completion support (if not already enabled)
echo "autoload -U compinit; compinit" >> ~/.zshrc

# Install completion
agentcli completion zsh > "${fpath[1]}/_agentcli"

# Restart your shell or reload configuration
source ~/.zshrc
```

#### Fish

```bash
# Load completion for current session
agentcli completion fish | source

# Install permanently
agentcli completion fish > ~/.config/fish/completions/agentcli.fish
```

#### PowerShell

```powershell
# Load completion for current session
agentcli completion powershell | Out-String | Invoke-Expression

# Install permanently
agentcli completion powershell > agentcli.ps1
# Add the following line to your PowerShell profile:
# . /path/to/agentcli.ps1
```

**Completion Features:**
- **Command completion**: Tab completion for all available commands
- **Flag completion**: Tab completion for all command flags and options
- **Template completion**: Intelligent completion for `--template` flag with all available templates
- **Provider completion**: Tab completion for `--provider` flag (openai, azure, ollama, mock)
- **Memory provider completion**: Tab completion for `--memory` flag (memory, pgvector, weaviate)
- **File completion**: Smart file completion for template validation and other file operations

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
Create new AgenticGoKit projects with multi-agent workflows

```bash
# Create a basic project
agentcli create my-project

# Create from template
agentcli create my-project --template research-assistant

# Custom configuration with consolidated flags
agentcli create my-project --memory pgvector --embedding openai --rag 1500 --mcp production

# Interactive mode for guided setup
agentcli create --interactive

# Show available templates
agentcli create help-templates
```

**Key Features:**
- **Template System**: Pre-configured project templates for common use cases
- **Consolidated Flags**: Simplified flag structure (12 flags instead of 32)
- **Intelligent Defaults**: Automatic dependency resolution and sensible defaults
- **External Templates**: Support for custom templates via JSON/YAML files
- **Interactive Mode**: Guided project setup

**Available Templates:**
- `basic` - Simple 2-agent sequential system
- `research-assistant` - Multi-agent research with web search and analysis
- `rag-system` - Document Q&A with vector search and RAG
- `data-pipeline` - Sequential data processing workflow
- `chat-system` - Conversational agents with memory

**Consolidated Flags:**
- `--template, -t` - Project template name
- `--agents, -a` - Number of agents to create
- `--provider, -p` - LLM provider (openai, azure, ollama, mock)
- `--memory` - Memory system provider (memory, pgvector, weaviate)
- `--embedding` - Embedding provider and model (openai, ollama:model, dummy)
- `--mcp` - MCP integration level (basic, production, full)
- `--rag` - Enable RAG with optional chunk size
- `--orchestration` - Orchestration mode (sequential, collaborative, loop, route)
- `--visualize` - Generate Mermaid workflow diagrams
- `--interactive, -i` - Interactive mode for guided setup

### `template`
Manage project templates

```bash
# List all available templates (built-in + custom)
agentcli template list

# Create a new custom template
agentcli template create my-template

# Validate a template file
agentcli template validate my-template.yaml

# Show template search paths
agentcli template paths
```

**Template Locations:**
Templates are searched in the following locations (in priority order):
1. Current directory: `.agenticgokit/templates/`
2. User home: `~/.agenticgokit/templates/`
3. System-wide: `/etc/agenticgokit/templates/` (Unix) or `%PROGRAMDATA%/AgenticGoKit/templates/` (Windows)

**Template Formats:**
- JSON format: `my-template.json`
- YAML format: `my-template.yaml` or `my-template.yml`

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

### `completion`
Generate shell completion scripts

```bash
# Generate completion for different shells
agentcli completion bash
agentcli completion zsh
agentcli completion fish
agentcli completion powershell

# Install completion (examples)
agentcli completion bash > /etc/bash_completion.d/agentcli
agentcli completion zsh > "${fpath[1]}/_agentcli"
agentcli completion fish > ~/.config/fish/completions/agentcli.fish
```

**Supported Shells:**
- Bash (Linux, macOS, Windows with Git Bash)
- Zsh (macOS default, Linux)
- Fish (cross-platform)
- PowerShell (Windows, cross-platform PowerShell Core)

### `version`
Show version information

```bash
# Show basic version
agentcli version

# Show detailed version information
agentcli version --output detailed

# Show version in JSON format
agentcli version --output json

# Show short version only
agentcli version --short
```

## 📚 Usage Examples

### Project Creation

```bash
# Quick project creation with templates
agentcli create research-bot --template research-assistant
agentcli create knowledge-base --template rag-system
agentcli create data-flow --template data-pipeline

# Custom configuration with consolidated flags
agentcli create custom-bot --memory pgvector --embedding openai --rag 1500 --mcp production

# Interactive mode for guided setup
agentcli create --interactive

# Override template defaults
agentcli create my-research --template research-assistant --agents 5 --mcp production
```

### Template Management

```bash
# List all available templates
agentcli template list

# Create a custom template
agentcli template create my-company-standard

# Validate template syntax
agentcli template validate my-template.yaml

# Show template search paths
agentcli template paths
```

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