# vNext MCP + Tools Blog Demo

A minimal example that accompanies the blog "MCP + Tool Integration in AgenticGoKit vNext".

## What it shows

- Enabling MCP in vNext Tools
- Explicit server (HTTP SSE) and Discovery modes
- Listing discovered tools (internal + MCP)
- Direct tool execution by name (internal `echo`)
- Driving execution via LLM-style TOOL_CALL output
- Running an agent end-to-end

## Prerequisites

- Go 1.24+
- An LLM provider plugin; this demo uses Ollama and the `gemma3:1b` model
  - `ollama pull gemma3:1b`
- An MCP server (for explicit server mode); example uses HTTP SSE on `localhost:8812`
  - Or use Discovery mode to scan common ports

## Run it (Windows PowerShell)

```powershell
# From repo root
pwsh -NoProfile -Command "cd examples/vnext/mcp-tools-blog-demo; go run ."
```

## Notes

- Plugins are imported blank inside `main.go`:
  - `plugins/mcp/unified` (transport)
  - `plugins/mcp/default` (registry/cache)
  - `plugins/llm/ollama` (LLM provider)
- Swap the LLM provider/model to match your environment if you prefer OpenAI/Azure/etc.
