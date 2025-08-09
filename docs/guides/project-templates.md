# Project Templates Guide

> **Navigation:** [Documentation Home](../README.md) → [Guides](README.md) → **Project Templates**

Learn how to use built-in templates and create custom project templates for AgenticGoKit projects.

## Overview

AgenticGoKit's template system allows you to quickly scaffold projects with pre-configured settings for common use cases. Templates can be built-in (provided with AgenticGoKit) or custom (defined by you or your team).

## Quick Start

### Using Built-in Templates

```bash
# Create a research assistant with web search capabilities
agentcli create research-bot --template research-assistant

# Create a RAG-powered knowledge base
agentcli create knowledge-base --template rag-system

# Create a data processing pipeline
agentcli create data-pipeline --template data-pipeline

# List all available templates
agentcli template list
```

### Creating Custom Templates

```bash
# Create a new template file
agentcli template create my-company-standard

# Edit the generated template file
# .agenticgokit/templates/my-company-standard.yaml

# Use your custom template
agentcli create my-project --template my-company-standard
```

## Built-in Templates

### `basic`
**Simple Multi-Agent System**
- 2 agents with sequential orchestration
- Basic error handling and responsible AI
- Good starting point for learning

```bash
agentcli create my-project --template basic
```

### `research-assistant`
**Multi-Agent Research System**
- 3 collaborative agents (researcher, analyzer, synthesizer)
- Web search and summarization tools via MCP
- Ideal for information gathering and analysis

```bash
agentcli create research-bot --template research-assistant
```

### `rag-system`
**RAG Knowledge Base**
- 3 collaborative agents for document processing
- PostgreSQL with vector search (pgvector)
- OpenAI embeddings and RAG capabilities
- Perfect for Q&A systems and knowledge bases

```bash
agentcli create knowledge-base --template rag-system
```

### `data-pipeline`
**Sequential Data Processing**
- 4 agents in sequential pipeline (ingester, processor, validator, outputter)
- Error handling and workflow visualization
- Great for ETL and data transformation workflows

```bash
agentcli create data-flow --template data-pipeline
```

### `chat-system`
**Conversational System**
- 2 agents with route-based orchestration
- Session-based memory for multi-user scenarios
- In-memory storage for fast responses
- Ideal for chatbots and conversational interfaces

```bash
agentcli create chat-bot --template chat-system
```

## Template Customization

You can override template defaults with additional flags:

```bash
# Use research-assistant template but with 5 agents and production MCP
agentcli create my-research --template research-assistant --agents 5 --mcp production

# Use rag-system template but with Weaviate instead of PostgreSQL
agentcli create my-kb --template rag-system --memory weaviate

# Use basic template but add memory and visualization
agentcli create my-project --template basic --memory pgvector --visualize
```

## Custom Templates

### Template Locations

Custom templates are searched in the following locations (in priority order):

1. **Current directory**: `.agenticgokit/templates/`
2. **User home**: `~/.agenticgokit/templates/`
3. **System-wide**:
   - Unix/Linux/macOS: `/etc/agenticgokit/templates/`
   - Windows: `%PROGRAMDATA%/AgenticGoKit/templates/`

### Template Format

Templates can be defined in JSON or YAML format:

#### YAML Example

```yaml
name: "E-commerce Assistant"
description: "Multi-agent system for e-commerce operations"
features:
  - "product-search"
  - "inventory-management"
  - "customer-support"

config:
  numAgents: 4
  provider: "openai"
  orchestrationMode: "collaborative"
  collaborativeAgents:
    - "product-searcher"
    - "inventory-manager"
    - "customer-support"
    - "order-processor"
  
  # Memory configuration
  memoryEnabled: true
  memoryProvider: "pgvector"
  embeddingProvider: "openai"
  
  # MCP configuration
  mcpEnabled: true
  mcpTools:
    - "product_catalog"
    - "inventory_api"
    - "customer_db"
  
  # Other options
  responsibleAI: true
  errorHandler: true
  visualize: true
```

#### JSON Example

```json
{
  "name": "Trading Bot System",
  "description": "Multi-agent trading system with market analysis",
  "features": [
    "market-analysis",
    "risk-management",
    "automated-trading"
  ],
  "config": {
    "numAgents": 5,
    "provider": "openai",
    "orchestrationMode": "collaborative",
    "collaborativeAgents": [
      "market-analyzer",
      "sentiment-analyzer",
      "risk-manager",
      "strategy-executor",
      "portfolio-manager"
    ],
    "memoryEnabled": true,
    "memoryProvider": "pgvector",
    "ragEnabled": true,
    "mcpEnabled": true,
    "mcpProduction": true,
    "responsibleAI": true,
    "errorHandler": true
  }
}
```

### Configuration Options

#### Basic Configuration
- `numAgents` (int): Number of agents to create
- `provider` (string): LLM provider ("openai", "azure", "ollama", "mock")
- `orchestrationMode` (string): Agent coordination ("sequential", "collaborative", "loop", "route")

#### Agent Configuration
- `collaborativeAgents` ([]string): Agent names for collaborative mode
- `sequentialAgents` ([]string): Agent names for sequential mode
- `loopAgent` (string): Agent name for loop mode

#### Memory & RAG Configuration
- `memoryEnabled` (bool): Enable memory system
- `memoryProvider` (string): Memory provider ("memory", "pgvector", "weaviate")
- `embeddingProvider` (string): Embedding provider ("openai", "ollama", "dummy")
- `embeddingModel` (string): Specific embedding model
- `ragEnabled` (bool): Enable RAG functionality
- `ragChunkSize` (int): Document chunk size in tokens
- `ragOverlap` (int): Overlap between chunks
- `ragTopK` (int): Number of results to retrieve
- `ragScoreThreshold` (float): Minimum similarity score
- `hybridSearch` (bool): Enable hybrid search
- `sessionMemory` (bool): Enable session-based memory isolation

#### MCP Configuration
- `mcpEnabled` (bool): Enable MCP tool integration
- `mcpProduction` (bool): Enable production MCP features
- `mcpTools` ([]string): List of MCP tools to include
- `withCache` (bool): Enable MCP result caching
- `withMetrics` (bool): Enable Prometheus metrics

#### Other Options
- `responsibleAI` (bool): Include responsible AI features
- `errorHandler` (bool): Include error handling
- `visualize` (bool): Generate workflow diagrams

## Template Management

### Creating Templates

```bash
# Create a YAML template (default)
agentcli template create my-template

# Create a JSON template
agentcli template create my-template --format json

# Create template in specific location
agentcli template create my-template --output /path/to/template.yaml
```

### Validating Templates

```bash
# Validate template syntax and configuration
agentcli template validate my-template.yaml

# Example output:
# Template validation successful!
# Name: My Custom Template
# Description: A custom project template
# Features: [custom-feature, example]
# Agents: 3
# Provider: openai
# Memory: pgvector
# RAG: enabled (chunk size: 1000)
```

### Listing Templates

```bash
# List all available templates
agentcli template list

# Show template search paths
agentcli template paths
```

## Common Use Cases

### Company Standard Template

Create a template that matches your organization's standards:

```yaml
name: "Company Standard"
description: "Standard multi-agent setup for our company"
config:
  numAgents: 3
  provider: "azure"  # Company uses Azure
  memoryEnabled: true
  memoryProvider: "pgvector"  # Company standard
  embeddingProvider: "openai"
  mcpEnabled: true
  mcpProduction: true
  responsibleAI: true
  errorHandler: true
```

### Environment-Specific Templates

Different templates for different environments:

```yaml
# development-template.yaml
name: "Development Setup"
description: "Fast setup for development"
config:
  numAgents: 2
  provider: "mock"  # Fast for testing
  memoryProvider: "memory"  # In-memory for speed
  mcpEnabled: false  # Simplified for dev
```

```yaml
# production-template.yaml
name: "Production Setup"
description: "Production-ready configuration"
config:
  numAgents: 5
  provider: "openai"
  memoryEnabled: true
  memoryProvider: "pgvector"
  ragEnabled: true
  mcpEnabled: true
  mcpProduction: true
  withCache: true
  withMetrics: true
```

### Domain-Specific Templates

Templates for specific industries or use cases:

```yaml
name: "Healthcare Assistant"
description: "HIPAA-compliant healthcare agent system"
config:
  numAgents: 3
  provider: "azure"  # For compliance
  orchestrationMode: "sequential"
  memoryEnabled: true
  memoryProvider: "pgvector"
  ragEnabled: true
  mcpEnabled: true
  mcpTools:
    - "medical_database"
    - "patient_records"
  responsibleAI: true  # Critical for healthcare
  errorHandler: true
```

## Best Practices

### Template Design
1. **Use descriptive names**: Choose template names that clearly indicate their purpose
2. **Include comprehensive features list**: Help users understand what the template provides
3. **Set sensible defaults**: Configure options that work well together
4. **Document your templates**: Include clear descriptions and use cases

### Template Organization
1. **Version your templates**: Consider including version information in descriptions
2. **Share templates**: Place commonly used templates in shared locations for team use
3. **Validate templates**: Always test your templates before sharing
4. **Use consistent naming**: Follow naming conventions for your organization

### Template Usage
1. **Start with templates**: Use templates as starting points, then customize as needed
2. **Override when necessary**: Use flags to override template defaults for specific needs
3. **Create project-specific templates**: For recurring project patterns, create custom templates

## Troubleshooting

### Template Not Found

If your template isn't being found:

1. Check template search paths: `agentcli template paths`
2. Ensure the file has the correct extension (`.json`, `.yaml`, or `.yml`)
3. Validate the template syntax: `agentcli template validate your-template.yaml`

### Template Validation Errors

Common validation errors:

- **Invalid JSON/YAML syntax**: Check for missing commas, quotes, or indentation
- **Missing required fields**: Ensure `name`, `description`, and `config` are present
- **Invalid configuration values**: Check that provider names, orchestration modes, etc. are valid

### Template Override Issues

If template overrides aren't working:

- External templates take priority over built-in templates with the same name
- Check that the template is in the correct search path
- Use `agentcli template list` to see which templates are loaded

## Related Guides

- [CLI Reference](../reference/cli.md) - Complete CLI command reference
- [Configuration](Configuration.md) - Project configuration options
- [Getting Started](../tutorials/getting-started/quickstart.md) - Basic project creation
- [Best Practices](development/best-practices.md) - Development best practices