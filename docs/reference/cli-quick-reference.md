# CLI Quick Reference

> **Navigation:** [Documentation Home](../README.md) → [Reference](README.md) → **CLI Quick Reference**

Quick reference for the most commonly used AgenticGoKit CLI commands.

## Project Creation

### Templates (Recommended)
```bash
# Research assistant with web search
agentcli create research-bot --template research-assistant

# RAG knowledge base
agentcli create knowledge-base --template rag-system

# Data processing pipeline
agentcli create data-flow --template data-pipeline

# Chat system with memory
agentcli create chat-bot --template chat-system

# Basic multi-agent system
agentcli create my-project --template basic
```

### Custom Configuration
```bash
# Custom RAG system
agentcli create my-kb --memory pgvector --embedding openai --rag 1500

# Production MCP setup
agentcli create my-bot --mcp production --visualize

# Interactive setup
agentcli create --interactive
```

### Template Overrides
```bash
# Use template but customize
agentcli create my-research --template research-assistant --agents 5 --mcp production
```

## Template Management

```bash
# List all templates
agentcli template list

# Create custom template
agentcli template create my-template

# Validate template
agentcli template validate my-template.yaml

# Show search paths
agentcli template paths
```

## Consolidated Flags

| Flag | Description | Examples |
|------|-------------|----------|
| `--template, -t` | Project template | `basic`, `research-assistant`, `rag-system` |
| `--memory` | Memory provider | `memory`, `pgvector`, `weaviate` |
| `--embedding` | Embedding provider | `openai`, `ollama:nomic-embed-text` |
| `--mcp` | MCP integration | `basic`, `production`, `full` |
| `--rag` | RAG chunk size | `default`, `1000`, `2000` |
| `--orchestration` | Agent coordination | `sequential`, `collaborative`, `loop` |
| `--agents, -a` | Number of agents | `2`, `3`, `5` |
| `--provider, -p` | LLM provider | `openai`, `azure`, `ollama` |
| `--visualize` | Generate diagrams | (boolean flag) |
| `--interactive, -i` | Interactive mode | (boolean flag) |

## Built-in Templates

| Template | Description | Use Case |
|----------|-------------|----------|
| `basic` | 2 agents, sequential | Learning, simple projects |
| `research-assistant` | 3 agents, web search | Research, analysis |
| `rag-system` | 3 agents, vector DB | Knowledge bases, Q&A |
| `data-pipeline` | 4 agents, sequential | ETL, data processing |
| `chat-system` | 2 agents, memory | Chatbots, conversations |

## Other Commands

```bash
# Version info
agentcli version

# List sessions
agentcli list

# View traces
agentcli trace session-id

# Memory debug
agentcli memory --stats

# MCP management
agentcli mcp servers

# Cache management
agentcli cache stats

# Shell completion
agentcli completion bash > /etc/bash_completion.d/agentcli
agentcli completion zsh > "${fpath[1]}/_agentcli"
agentcli completion powershell > agentcli.ps1
```

## Common Patterns

### Development Setup
```bash
# Quick development project
agentcli create dev-project --template basic --provider mock
```

### Production Setup
```bash
# Production-ready RAG system
agentcli create prod-kb --template rag-system --mcp production --visualize
```

### Research Workflow
```bash
# Research assistant with custom agents
agentcli create research-team --template research-assistant --agents 4
```

### Data Processing
```bash
# ETL pipeline with visualization
agentcli create etl-system --template data-pipeline --visualize
```

## Help Commands

```bash
# General help
agentcli --help

# Command-specific help
agentcli create --help
agentcli template --help

# Template details
agentcli create help-templates
agentcli template list
```

## Shell Completion

Enable intelligent tab completion for faster CLI usage:

### Bash
```bash
# Load for current session
source <(agentcli completion bash)

# Install permanently (Linux)
agentcli completion bash > /etc/bash_completion.d/agentcli

# Install permanently (macOS with Homebrew)
agentcli completion bash > $(brew --prefix)/etc/bash_completion.d/agentcli
```

### Zsh
```bash
# Enable completion support
echo "autoload -U compinit; compinit" >> ~/.zshrc

# Install completion
agentcli completion zsh > "${fpath[1]}/_agentcli"
```

### PowerShell
```powershell
# Load for current session
agentcli completion powershell | Out-String | Invoke-Expression

# Install permanently
agentcli completion powershell > agentcli.ps1
# Add to your PowerShell profile
```

### Fish
```bash
# Load for current session
agentcli completion fish | source

# Install permanently
agentcli completion fish > ~/.config/fish/completions/agentcli.fish
```

For complete documentation, see the [Full CLI Reference](cli.md).