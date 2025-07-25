# Setup & Configuration Guides

Guides for setting up and configuring AgenticGoKit components.

## Available Guides

### [LLM Providers](llm-providers.md)
Configure different Large Language Model providers including OpenAI, Anthropic, and local models.

**When to use:** Setting up your first agent or switching LLM providers.

### [Vector Databases](vector-databases.md)
Set up vector storage systems for Retrieval-Augmented Generation (RAG) including pgvector and Weaviate.

**When to use:** Building agents that need to search through documents or knowledge bases.

### [MCP Tools](mcp-tools.md)
Integrate Model Context Protocol tools to extend agent capabilities with external services.

**When to use:** Adding specific tools like web search, file operations, or API integrations.

## Common Setup Patterns

Most AgenticGoKit applications follow this setup pattern:

1. **Choose your LLM provider** - Start with [LLM Providers](llm-providers.md)
2. **Configure memory (optional)** - Add [Vector Databases](vector-databases.md) if needed
3. **Add tools (optional)** - Integrate [MCP Tools](mcp-tools.md) for extended capabilities
4. **Test your configuration** - Use [Testing Agents](../development/testing-agents.md)

## Configuration Files

AgenticGoKit uses TOML configuration files (`agentflow.toml`) for most settings. Each guide shows you how to configure the relevant sections.

## Next Steps

After setup, explore:
- [Development Guides](../development/README.md) for building and testing
- [Deployment Guides](../deployment/README.md) for production deployment
- [Getting Started Tutorials](../../tutorials/getting-started/README.md) for hands-on learning