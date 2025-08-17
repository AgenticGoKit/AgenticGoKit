package main

// Plugin bundle for AgentCLI. These blank imports register providers via init().
import (
	_ "github.com/kunalkushwaha/agenticgokit/plugins/llm/azureopenai"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/llm/ollama"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/llm/openai"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/logging/zerolog"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/memory/memory"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/memory/pgvector"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/memory/weaviate"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/orchestrator/default"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/runner/default"
)
