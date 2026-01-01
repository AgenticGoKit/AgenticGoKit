package main

// Plugin bundle for AgentCLI. These blank imports register providers via init().
import (
	_ "github.com/agenticgokit/agenticgokit/plugins/llm/azureopenai"
	_ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
	_ "github.com/agenticgokit/agenticgokit/plugins/llm/openai"
	_ "github.com/agenticgokit/agenticgokit/plugins/logging/zerolog"
	_ "github.com/agenticgokit/agenticgokit/plugins/mcp/default"
	_ "github.com/agenticgokit/agenticgokit/plugins/mcp/registry/memory"
	_ "github.com/agenticgokit/agenticgokit/plugins/mcp/tcp"
	_ "github.com/agenticgokit/agenticgokit/plugins/memory/chromem"
	_ "github.com/agenticgokit/agenticgokit/plugins/memory/pgvector"
	_ "github.com/agenticgokit/agenticgokit/plugins/memory/weaviate"
	_ "github.com/agenticgokit/agenticgokit/plugins/orchestrator/default"
	_ "github.com/agenticgokit/agenticgokit/plugins/runner/default"
)

