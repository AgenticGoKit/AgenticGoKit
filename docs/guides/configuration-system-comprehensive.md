# AgentFlow Configuration System - Comprehensive Guide

## Table of Contents

1. [Overview](#overview)
2. [Configuration Structure](#configuration-structure)
3. [Agent Configuration](#agent-configuration)
4. [LLM Configuration](#llm-configuration)
5. [Orchestration Configuration](#orchestration-configuration)
6. [Memory and RAG Configuration](#memory-and-rag-configuration)
7. [MCP Configuration](#mcp-configuration)
8. [Environment Variables](#environment-variables)
9. [Configuration Validation](#configuration-validation)
10. [Hot Reload](#hot-reload)
11. [Best Practices](#best-practices)
12. [Migration Guide](#migration-guide)
13. [Troubleshooting](#troubleshooting)
14. [Examples](#examples)

## Overview

The AgentFlow Configuration System provides a comprehensive, flexible way to configure multi-agent AI systems. It supports:

- **Declarative Configuration**: Define agents, LLM settings, and orchestration in configuration files
- **Multiple Formats**: TOML, YAML, and JSON support
- **Environment Overrides**: Override configuration values with environment variables
- **Hot Reload**: Update configuration without restarting the application
- **Validation**: Comprehensive validation with helpful error messages
- **Backward Compatibility**: Seamless migration from hardcoded configurations

## Configuration Structure

### Basic Structure

```toml
[agent_flow]
name = "my-agent-system"
version = "1.0.0"
description = "A comprehensive multi-agent system"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 2000

[agents.researcher]
role = "researcher"
description = "Research and information gathering agent"
system_prompt = "You are a research specialist..."
capabilities = ["web_search", "document_analysis"]
enabled = true

[agents.writer]
role = "writer"
description = "Content creation agent"
system_prompt = "You are a skilled writer..."
capabilities = ["content_creation", "editing"]
enabled = true

[orchestration]
mode = "sequential"
agents = ["researcher", "writer"]
```

### Supported Formats

#### TOML (Recommended)
```toml
[agent_flow]
name = "example-system"

[llm]
provider = "openai"
model = "gpt-4"
```

#### YAML
```yaml
agent_flow:
  name: example-system

llm:
  provider: openai
  model: gpt-4
```

#### JSON
```json
{
  "agent_flow": {
    "name": "example-system"
  },
  "llm": {
    "provider": "openai",
    "model": "gpt-4"
  }
}
```

## Agent Configuration

### Basic Agent Configuration

```toml
[agents.my_agent]
role = "processor"                    # Required: Agent's role identifier
description = "Data processing agent" # Required: Human-readable description
system_prompt = "You are a data processor specialized in analyzing and transforming information."
capabilities = ["data_processing", "analysis", "transformation"]
enabled = true                        # Optional: Enable/disable agent (default: true)
```

### Agent-Specific LLM Configuration

Agents can override global LLM settings:

```toml
[agents.creative_writer]
role = "writer"
description = "Creative content writer"
system_prompt = "You are a creative writer with expertise in storytelling."
capabilities = ["creative_writing", "storytelling"]
enabled = true

# Agent-specific LLM settings
[agents.creative_writer.llm]
temperature = 0.9        # Higher creativity
max_tokens = 3000        # Longer responses
model = "gpt-4-turbo"    # Different model
```

### Agent Retry and Rate Limiting

```toml
[agents.api_agent]
role = "api_caller"
description = "Agent that makes API calls"
system_prompt = "You make API calls and process responses."
capabilities = ["api_calls", "data_processing"]

# Retry configuration
[agents.api_agent.retry_policy]
max_retries = 5
base_delay_ms = 1000
max_delay_ms = 30000
backoff_multiplier = 2.0

# Rate limiting
[agents.api_agent.rate_limit]
requests_per_second = 5
burst_size = 10
```

### Advanced Agent Configuration

```toml
[agents.advanced_agent]
role = "advanced_processor"
description = "Advanced agent with full configuration"
system_prompt = "You are an advanced AI agent with sophisticated capabilities."
capabilities = ["advanced_processing", "multi_modal", "reasoning"]
enabled = true

# Custom metadata
[agents.advanced_agent.metadata]
version = "2.0"
author = "AgentFlow Team"
tags = ["advanced", "production"]

# Agent-specific LLM configuration
[agents.advanced_agent.llm]
provider = "openai"
model = "gpt-4-turbo"
temperature = 0.7
max_tokens = 4000
top_p = 0.9
frequency_penalty = 0.1
presence_penalty = 0.1

# Custom parameters for specific providers
[agents.advanced_agent.llm.custom_params]
stream = true
logit_bias = { "50256" = -100 }  # Avoid specific tokens
```

## LLM Configuration

### Global LLM Configuration

```toml
[llm]
provider = "openai"              # Required: LLM provider
model = "gpt-4"                  # Required: Model name
temperature = 0.7                # Optional: Creativity (0.0-2.0)
max_tokens = 2000                # Optional: Maximum response length
top_p = 1.0                      # Optional: Nucleus sampling
frequency_penalty = 0.0          # Optional: Frequency penalty
presence_penalty = 0.0           # Optional: Presence penalty
timeout_seconds = 30             # Optional: Request timeout
```

### Provider-Specific Configuration

#### OpenAI Configuration
```toml
[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 2000
api_key_env = "OPENAI_API_KEY"   # Environment variable for API key
organization = "org-123456"       # Optional: Organization ID
base_url = "https://api.openai.com/v1"  # Optional: Custom base URL
```

#### Anthropic Configuration
```toml
[llm]
provider = "anthropic"
model = "claude-3-opus-20240229"
temperature = 0.7
max_tokens = 4000
api_key_env = "ANTHROPIC_API_KEY"
```

#### Ollama Configuration
```toml
[llm]
provider = "ollama"
model = "llama2"
temperature = 0.7
max_tokens = 2000
base_url = "http://localhost:11434"
```

#### Azure OpenAI Configuration
```toml
[llm]
provider = "azure_openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 2000
api_key_env = "AZURE_OPENAI_API_KEY"
endpoint = "https://your-resource.openai.azure.com/"
api_version = "2023-12-01-preview"
deployment_name = "gpt-4-deployment"
```

## Orchestration Configuration

### Sequential Orchestration

```toml
[orchestration]
mode = "sequential"
agents = ["researcher", "analyzer", "writer", "reviewer"]

# Optional: Sequential-specific settings
[orchestration.sequential]
stop_on_error = true
pass_context = true
timeout_per_agent = 300  # seconds
```

### Collaborative Orchestration

```toml
[orchestration]
mode = "collaborative"
agents = ["researcher", "writer", "reviewer"]

# Collaborative-specific settings
[orchestration.collaborative]
max_iterations = 5
consensus_threshold = 0.8
voting_mechanism = "majority"
discussion_rounds = 3
```

### Parallel Orchestration

```toml
[orchestration]
mode = "parallel"
agents = ["processor_1", "processor_2", "processor_3"]

# Parallel-specific settings
[orchestration.parallel]
max_concurrent = 3
timeout_seconds = 600
merge_strategy = "concatenate"  # or "vote", "average"
```

### Custom Orchestration

```toml
[orchestration]
mode = "custom"
orchestrator_class = "MyCustomOrchestrator"
agents = ["agent_1", "agent_2", "agent_3"]

# Custom orchestration parameters
[orchestration.custom]
custom_param_1 = "value1"
custom_param_2 = 42
custom_param_3 = true
```

## Memory and RAG Configuration

### Memory Configuration

```toml
[memory]
enabled = true
provider = "pgvector"            # or "chroma", "pinecone", "weaviate"
connection_string = "postgresql://user:pass@localhost:5432/agentflow"
collection_name = "agent_memory"
max_entries = 10000
ttl_seconds = 86400             # 24 hours

# Memory-specific settings
[memory.settings]
similarity_threshold = 0.8
max_results = 10
include_metadata = true
```

### Embedding Configuration

```toml
[memory.embedding]
provider = "openai"              # or "huggingface", "sentence_transformers"
model = "text-embedding-ada-002"
dimensions = 1536
batch_size = 100
api_key_env = "OPENAI_API_KEY"

# Provider-specific settings
[memory.embedding.openai]
chunk_size = 8000

[memory.embedding.huggingface]
model_name = "sentence-transformers/all-MiniLM-L6-v2"
device = "cpu"                   # or "cuda"
```

### RAG Configuration

```toml
[rag]
enabled = true
chunk_size = 1000
chunk_overlap = 200
similarity_threshold = 0.8
max_chunks = 5
rerank = true

# Document processing
[rag.document_processing]
supported_formats = ["txt", "pdf", "docx", "md"]
max_file_size_mb = 10
extract_metadata = true

# Retrieval settings
[rag.retrieval]
strategy = "similarity"          # or "mmr", "hybrid"
mmr_diversity_bias = 0.3        # for MMR strategy
hybrid_alpha = 0.5              # for hybrid strategy

# Reranking
[rag.reranking]
enabled = true
model = "cross-encoder/ms-marco-MiniLM-L-6-v2"
top_k = 10
```

## MCP Configuration

### Basic MCP Configuration

```toml
[mcp]
enabled = true
timeout_seconds = 30
max_retries = 3

# MCP Server configurations
[[mcp.servers]]
name = "web_search"
type = "stdio"                   # or "sse", "websocket"
command = "uvx"
args = ["mcp-server-web-search"]
enabled = true

[[mcp.servers]]
name = "filesystem"
type = "stdio"
command = "uvx"
args = ["mcp-server-filesystem", "/workspace"]
enabled = true
env = { "FILESYSTEM_ROOT" = "/workspace" }

[[mcp.servers]]
name = "database"
type = "sse"
url = "http://localhost:8080/mcp"
enabled = true
headers = { "Authorization" = "Bearer ${DB_TOKEN}" }
```

### Advanced MCP Configuration

```toml
[mcp]
enabled = true
timeout_seconds = 60
max_retries = 5
retry_delay_ms = 1000

# Global MCP settings
[mcp.settings]
auto_approve_tools = ["web_search", "file_read"]
require_confirmation = ["file_write", "database_modify"]
log_tool_calls = true
cache_results = true
cache_ttl_seconds = 3600

# Tool-specific configurations
[[mcp.tools]]
name = "web_search"
server = "web_search"
enabled = true
rate_limit = 10                  # requests per minute
cache_results = true

[[mcp.tools]]
name = "file_operations"
server = "filesystem"
enabled = true
permissions = ["read", "write"]
allowed_paths = ["/workspace", "/tmp"]
```

## Environment Variables

### Standard Environment Variables

```bash
# LLM Configuration
AGENTFLOW_LLM_PROVIDER=openai
AGENTFLOW_LLM_MODEL=gpt-4
AGENTFLOW_LLM_TEMPERATURE=0.7
AGENTFLOW_LLM_MAX_TOKENS=2000

# API Keys
OPENAI_API_KEY=your-openai-key
ANTHROPIC_API_KEY=your-anthropic-key
AZURE_OPENAI_API_KEY=your-azure-key

# Agent-specific overrides
AGENTFLOW_AGENTS_RESEARCHER_ENABLED=true
AGENTFLOW_AGENTS_RESEARCHER_LLM_TEMPERATURE=0.3
AGENTFLOW_AGENTS_WRITER_LLM_MODEL=gpt-4-turbo

# Memory configuration
AGENTFLOW_MEMORY_ENABLED=true
AGENTFLOW_MEMORY_PROVIDER=pgvector
AGENTFLOW_MEMORY_CONNECTION_STRING=postgresql://localhost:5432/agentflow

# MCP configuration
AGENTFLOW_MCP_ENABLED=true
AGENTFLOW_MCP_TIMEOUT_SECONDS=30
```

### Environment Variable Patterns

- **Global LLM**: `AGENTFLOW_LLM_<SETTING>`
- **Agent-specific**: `AGENTFLOW_AGENTS_<AGENT_NAME>_<SETTING>`
- **Agent LLM**: `AGENTFLOW_AGENTS_<AGENT_NAME>_LLM_<SETTING>`
- **Memory**: `AGENTFLOW_MEMORY_<SETTING>`
- **RAG**: `AGENTFLOW_RAG_<SETTING>`
- **MCP**: `AGENTFLOW_MCP_<SETTING>`

## Configuration Validation

### Validation Levels

1. **Syntax Validation**: TOML/YAML/JSON syntax
2. **Schema Validation**: Required fields and data types
3. **Semantic Validation**: Logical consistency
4. **Cross-Reference Validation**: Agent references in orchestration
5. **Provider Validation**: LLM provider-specific settings

### Validation Examples

```bash
# Basic validation
agentcli validate agentflow.toml

# Detailed validation with suggestions
agentcli validate --detailed agentflow.toml

# JSON output for automation
agentcli validate --format json agentflow.toml

# Quiet mode (errors only)
agentcli validate --quiet agentflow.toml
```

### Common Validation Errors

#### Missing Required Fields
```
❌ Field: agents.researcher.role
   Issue: Required field is missing
   Suggestion: Add role field to define the agent's purpose
   Value: null
```

#### Invalid Temperature Range
```
❌ Field: llm.temperature
   Issue: Temperature must be between 0.0 and 2.0
   Suggestion: Use a value between 0.0 (deterministic) and 2.0 (creative)
   Value: 3.5
```

#### Agent Reference Error
```
❌ Field: orchestration.agents
   Issue: Agent 'nonexistent_agent' referenced but not defined
   Suggestion: Define the agent in the [agents] section or remove from orchestration
   Value: ["researcher", "nonexistent_agent", "writer"]
```

## Hot Reload

### Enabling Hot Reload

```go
// In your application
reloader := core.NewConfigReloader("agentflow.toml")

// Set up reload callback
reloader.OnConfigReload(func(newConfig *core.Config) error {
    // Validate new configuration
    validator := core.NewDefaultConfigValidator()
    if errors := validator.ValidateConfig(newConfig); len(errors) > 0 {
        return fmt.Errorf("invalid configuration: %v", errors)
    }
    
    // Update agents
    return agentManager.UpdateConfiguration(newConfig)
})

// Start watching for changes
ctx := context.Background()
go reloader.StartWatching(ctx)
```

### Hot Reload Behavior

- **File Changes**: Automatically detected using file system watchers
- **Validation**: New configuration is validated before applying
- **Rollback**: Invalid configurations are rejected, keeping current config
- **Agent Updates**: Agents are updated with new settings
- **Graceful Handling**: In-flight requests complete before updates

### Hot Reload Best Practices

1. **Test Changes**: Validate configuration before saving
2. **Gradual Updates**: Make incremental changes
3. **Monitor Logs**: Watch for reload success/failure messages
4. **Backup Configs**: Keep backup of working configurations
5. **Use Validation**: Always validate before deploying changes

## Best Practices

### Configuration Organization

1. **Use Clear Names**: Choose descriptive agent names and roles
2. **Group Related Settings**: Keep related configuration together
3. **Document Purpose**: Add comments explaining complex configurations
4. **Version Control**: Track configuration changes in version control
5. **Environment Separation**: Use different configs for dev/staging/prod

### Security Best Practices

1. **Environment Variables**: Store sensitive data in environment variables
2. **API Key Management**: Use dedicated environment variables for API keys
3. **Access Control**: Limit file permissions on configuration files
4. **Secrets Management**: Use proper secrets management systems
5. **Audit Trail**: Log configuration changes

### Performance Optimization

1. **Agent Efficiency**: Configure appropriate temperature and token limits
2. **Memory Management**: Set reasonable memory limits and TTL
3. **Rate Limiting**: Configure rate limits to avoid API throttling
4. **Caching**: Enable result caching where appropriate
5. **Resource Monitoring**: Monitor resource usage and adjust accordingly

### Development Workflow

1. **Start Simple**: Begin with minimal configuration
2. **Iterate Gradually**: Add complexity incrementally
3. **Test Thoroughly**: Validate each change
4. **Use Templates**: Start with proven templates
5. **Document Changes**: Keep track of configuration evolution

## Migration Guide

### From Hardcoded to Configuration-Driven

#### Step 1: Identify Hardcoded Values
```go
// Before: Hardcoded agent creation
agent := &ResearchAgent{
    Role:        "researcher",
    Model:       "gpt-4",
    Temperature: 0.3,
    MaxTokens:   1500,
}
```

#### Step 2: Create Configuration
```toml
[agents.researcher]
role = "researcher"
description = "Research and information gathering agent"
system_prompt = "You are a research specialist."
capabilities = ["web_search", "document_analysis"]
enabled = true

[agents.researcher.llm]
model = "gpt-4"
temperature = 0.3
max_tokens = 1500
```

#### Step 3: Update Code
```go
// After: Configuration-driven agent creation
factory := core.NewConfigurableAgentFactory(config)
agent, err := factory.CreateAgent("researcher")
if err != nil {
    return fmt.Errorf("failed to create agent: %w", err)
}
```

### Migration Checklist

- [ ] Identify all hardcoded agent configurations
- [ ] Create configuration file with agent definitions
- [ ] Update agent creation code to use factory
- [ ] Add configuration validation
- [ ] Test with new configuration system
- [ ] Update deployment scripts
- [ ] Document new configuration structure

## Troubleshooting

### Common Issues

#### Configuration Not Loading
```bash
# Check file exists and has correct permissions
ls -la agentflow.toml

# Validate syntax
agentcli validate agentflow.toml

# Check for syntax errors
toml-lint agentflow.toml  # if available
```

#### Agent Creation Failures
```bash
# Check agent is defined in configuration
grep -A 10 "agents.my_agent" agentflow.toml

# Validate agent configuration
agentcli validate --detailed agentflow.toml

# Check for missing required fields
```

#### Environment Variable Issues
```bash
# List all AgentFlow environment variables
env | grep AGENTFLOW

# Check specific variable
echo $AGENTFLOW_LLM_MODEL

# Test with explicit values
AGENTFLOW_LLM_MODEL=gpt-4 ./my-agent-app
```

#### Hot Reload Not Working
```bash
# Check file permissions
ls -la agentflow.toml

# Monitor file changes
inotifywait -m agentflow.toml

# Check application logs for reload messages
tail -f application.log | grep reload
```

### Debug Mode

Enable debug logging to troubleshoot configuration issues:

```bash
# Set debug environment variable
export AGENTFLOW_DEBUG=true

# Or in configuration
[debug]
enabled = true
log_level = "debug"
log_config_loading = true
log_agent_creation = true
log_validation = true
```

## Examples

### Complete Production Configuration

```toml
[agent_flow]
name = "production-content-system"
version = "2.1.0"
description = "Production content creation and optimization system"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 2000
timeout_seconds = 60

[agents.researcher]
role = "researcher"
description = "Research and fact-checking specialist"
system_prompt = """You are a meticulous researcher specializing in gathering accurate, 
up-to-date information from reliable sources. Focus on factual accuracy and cite sources."""
capabilities = ["web_search", "fact_checking", "source_verification", "data_analysis"]
enabled = true

[agents.researcher.llm]
temperature = 0.3
max_tokens = 3000

[agents.researcher.retry_policy]
max_retries = 3
base_delay_ms = 1000
max_delay_ms = 10000

[agents.content_writer]
role = "writer"
description = "Content creation and SEO optimization specialist"
system_prompt = """You are a skilled content writer with expertise in SEO optimization. 
Create engaging, well-structured content that ranks well in search engines."""
capabilities = ["content_creation", "seo_optimization", "copywriting", "editing"]
enabled = true

[agents.content_writer.llm]
temperature = 0.8
max_tokens = 4000

[agents.quality_reviewer]
role = "reviewer"
description = "Quality assurance and final review specialist"
system_prompt = """You are a quality assurance specialist ensuring content meets 
high standards for accuracy, readability, and brand consistency."""
capabilities = ["quality_assurance", "proofreading", "brand_compliance", "fact_checking"]
enabled = true

[agents.quality_reviewer.llm]
temperature = 0.2
max_tokens = 2000

[orchestration]
mode = "sequential"
agents = ["researcher", "content_writer", "quality_reviewer"]

[orchestration.sequential]
stop_on_error = false
pass_context = true
timeout_per_agent = 600

[memory]
enabled = true
provider = "pgvector"
connection_string = "${DATABASE_URL}"
collection_name = "content_memory"
max_entries = 50000
ttl_seconds = 604800  # 7 days

[memory.embedding]
provider = "openai"
model = "text-embedding-ada-002"
dimensions = 1536
batch_size = 100

[rag]
enabled = true
chunk_size = 1000
chunk_overlap = 200
similarity_threshold = 0.8
max_chunks = 5

[rag.document_processing]
supported_formats = ["txt", "pdf", "docx", "md", "html"]
max_file_size_mb = 25
extract_metadata = true

[mcp]
enabled = true
timeout_seconds = 30

[[mcp.servers]]
name = "web_search"
type = "stdio"
command = "uvx"
args = ["mcp-server-web-search"]
enabled = true

[[mcp.servers]]
name = "seo_tools"
type = "stdio"
command = "uvx"
args = ["mcp-server-seo-tools"]
enabled = true

[retry_policy]
max_retries = 3
base_delay_ms = 1000
max_delay_ms = 30000
backoff_multiplier = 2.0

[rate_limit]
requests_per_second = 10
burst_size = 20

[monitoring]
enabled = true
metrics_endpoint = "/metrics"
health_endpoint = "/health"
log_level = "info"
```

### Development Configuration

```toml
[agent_flow]
name = "dev-content-system"
version = "2.1.0-dev"
description = "Development environment for content system"

[llm]
provider = "openai"
model = "gpt-3.5-turbo"  # Cheaper for development
temperature = 0.7
max_tokens = 1000        # Smaller for faster responses

[agents.researcher]
role = "researcher"
description = "Development researcher"
system_prompt = "You are a researcher. Keep responses concise for development."
capabilities = ["web_search", "basic_research"]
enabled = true

[agents.researcher.llm]
temperature = 0.5
max_tokens = 800

[agents.writer]
role = "writer"
description = "Development writer"
system_prompt = "You are a writer. Create brief content for testing."
capabilities = ["content_creation", "basic_editing"]
enabled = true

[orchestration]
mode = "sequential"
agents = ["researcher", "writer"]

[memory]
enabled = false  # Disabled for development

[rag]
enabled = false  # Disabled for development

[mcp]
enabled = false  # Disabled for development

[debug]
enabled = true
log_level = "debug"
log_config_loading = true
log_agent_creation = true
```

This comprehensive guide covers all aspects of the AgentFlow Configuration System. For additional help, consult the API documentation or reach out to the development team.