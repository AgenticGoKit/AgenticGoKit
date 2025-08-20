# Getting Started

Welcome to the CLI-first, config-first onboarding for AgenticGoKit. Follow this flow:

1) [Install prerequisites and CLI](install.md)

2) [Quickstart (CLI)](quickstart-cli.md) — Fastest path with scaffolding

3) Core building blocks
- [Your first scaffold](first-agent.md)
- [Orchestration basics](orchestration-basics.md)
- [Memory basics](memory-basics.md)
- [Tools basics (MCP)](tools-basics.md)
- [Deploy basics](deploy-basics.md)

AgentCLI Command Palette
- `agentcli create my-project`
- `agentcli validate`
- `agentcli config generate`
- `agentcli config template <name> > agentflow.toml`
- `agentcli mcp servers | agentcli mcp tools | agentcli mcp health`
- `agentcli trace --help`

Verification notes
- All commands and flags in these docs were validated against the current CLI. Use `agentcli --help` to confirm on your system.

Prerequisites
- Go 1.21+ installed
- Optional: Docker for deploy basics
- Local LLM recommended: Ollama with model gemma3:1b (or use OpenAI/Azure OpenAI)

Next
- Dive into [install.md](install.md) to set up your environment.
