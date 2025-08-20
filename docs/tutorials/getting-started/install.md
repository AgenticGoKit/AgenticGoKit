# Install and Verify (CLI-first)

This page sets up your environment and validates commands with no assumptions.

## Prerequisites
- Go 1.21+ on PATH
- AgenticGoKit CLI (agentcli)
- Recommended: Ollama running with model gemma3:1b for local runs

## Install agentcli

PowerShell (pwsh):
```pwsh
# From a folder on your system PATH
go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest
agentcli --help
agentcli version
```

Bash:
```bash
go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest
agentcli --help
agentcli version
```

## Local LLM (recommended)

Ollama:
```bash
# Bash (if on WSL or macOS/Linux)
ollama pull gemma3:1b
```

Windows: install Ollama and use its UI/CLI to pull gemma3:1b.

## Verify Go
```pwsh
go version
```

## Verify CLI commands exist
```pwsh
agentcli --help
agentcli config --help
agentcli mcp --help
```

If a command or flag isn’t available in your version’s help output, don’t use it.

## Next
- Start with quickstart-cli.md
