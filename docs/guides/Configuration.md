# Configuration Management

**Managing AgentFlow Configuration with agentflow.toml**

AgentFlow uses TOML configuration files to manage all aspects of your agent system: LLM providers, MCP servers, multi-agent orchestration settings, workflow visualization, and more.

## Basic Configuration Structure

### agentflow.toml Template

```toml
# Project metadata
name = "My Agent System"
version = "1.0.0"
description = "AgentFlow-powered agent workflow"

# Logging configuration
log_level = "info"  # debug, info, warn, error

# Multi-Agent Orchestration Configuration
[orchestration]
mode = "collaborative"        # collaborative, sequential, loop, mixed
timeout = "60s"
failure_threshold = 0.5       # 0.0-1.0
max_concurrency = 5
max_iterations = 10           # for loop mode

# Collaborative agents (parallel processing)
collaborative_agents = ["researcher", "analyzer", "validator"]

# Sequential agents (pipeline processing)
sequential_agents = ["collector", "processor", "formatter"]

# Loop agent (iterative processing)
loop_agent = "quality-checker"

# Workflow Visualization
[visualization]
enabled = true
output_dir = "./docs/diagrams"
diagram_type = "flowchart"    # flowchart, sequence, etc.
direction = "TD"              # TD, LR, BT, RL
show_metadata = true
show_agent_types = true

# LLM Provider configuration
[provider]
type = "azure"                    # azure, openai, ollama, mock
api_key = "${AZURE_OPENAI_API_KEY}"
endpoint = "https://your-resource.openai.azure.com"
deployment = "gpt-4"
api_version = "2024-02-15-preview"
model = "gpt-4"
max_tokens = 2000
temperature = 0.7
timeout = "30s"

# MCP (Model Context Protocol) configuration
[mcp]
enabled = true
cache_enabled = true
cache_ttl = "5m"
connection_timeout = "30s"
max_retries = 3

# MCP server definitions
[mcp.servers.search]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-web-search"]
transport = "stdio"

[mcp.servers.docker]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-docker"]
transport = "stdio"

# Agent orchestration settings
[orchestration]
mode = "sequential"     # sequential, parallel, collaborative
queue_size = 100
worker_count = 4
timeout = "5m"

# Error handling and routing
[error_routing]
validation_errors = "error_handler"
timeout_errors = "timeout_handler"
critical_errors = "critical_handler"
default_error_handler = "error_handler"

# Optional: Metrics and monitoring
[metrics]
enabled = false
prometheus_port = 8090
```

## Provider Configuration

### Azure OpenAI

```toml
[provider]
type = "azure"
api_key = "${AZURE_OPENAI_API_KEY}"
endpoint = "${AZURE_OPENAI_ENDPOINT}"
deployment = "${AZURE_OPENAI_DEPLOYMENT}"
api_version = "2024-02-15-preview"
model = "gpt-4"
max_tokens = 2000
temperature = 0.7
timeout = "30s"
max_retries = 3
```

**Required Environment Variables:**
```bash
export AZURE_OPENAI_API_KEY="your-azure-api-key"
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com"
export AZURE_OPENAI_DEPLOYMENT="gpt-4"
```

### OpenAI

```toml
[provider]
type = "openai"
api_key = "${OPENAI_API_KEY}"
model = "gpt-4"
max_tokens = 2000
temperature = 0.7
organization = "${OPENAI_ORG}"  # Optional
timeout = "30s"
```

**Required Environment Variables:**
```bash
export OPENAI_API_KEY="your-openai-api-key"
export OPENAI_ORG="your-organization-id"  # Optional
```

### Ollama (Local Models)

```toml
[provider]
type = "ollama"
host = "http://localhost:11434"
model = "llama3.2:3b"
temperature = 0.7
context_window = 4096
timeout = "60s"
```

**Setup Ollama:**
```bash
# Install Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# Pull your preferred model
ollama pull llama3.2:3b
ollama pull codellama:7b
ollama pull mistral:7b

# Start server (usually automatic)
ollama serve
```

### Mock Provider (Testing)

```toml
[provider]
type = "mock"
response = "This is a mock response for testing"
delay = "100ms"
error_rate = 0.0  # 0.0 = no errors, 0.1 = 10% error rate
```

## MCP Configuration

### Basic MCP Setup

```toml
[mcp]
enabled = true
cache_enabled = true
cache_ttl = "5m"

# Web search tools
[mcp.servers.search]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-web-search"]
transport = "stdio"

# Docker management
[mcp.servers.docker]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-docker"]
transport = "stdio"
```

### Production MCP Configuration

```toml
[mcp]
enabled = true
cache_enabled = true
cache_ttl = "10m"
connection_timeout = "30s"
max_retries = 3
max_concurrent_connections = 10

[mcp.cache]
type = "memory"          # memory, redis (future)
max_size = 1000
cleanup_interval = "1m"

# Production-ready servers with environment variables
[mcp.servers.search]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-web-search"]
transport = "stdio"
env = { "SEARCH_API_KEY" = "${SEARCH_API_KEY}" }

[mcp.servers.database]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-postgres"]
transport = "stdio"
env = { "DATABASE_URL" = "${DATABASE_URL}" }

[mcp.servers.github]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-github"]
transport = "stdio"
env = { "GITHUB_TOKEN" = "${GITHUB_TOKEN}" }
```

### Available MCP Servers

```toml
# Development Tools
[mcp.servers.filesystem]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-filesystem"]
transport = "stdio"

[mcp.servers.docker]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-docker"]
transport = "stdio"

# Web & Search
[mcp.servers.brave_search]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-brave-search"]
transport = "stdio"
env = { "BRAVE_API_KEY" = "${BRAVE_API_KEY}" }

[mcp.servers.fetch]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-fetch"]
transport = "stdio"

# Databases
[mcp.servers.postgres]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-postgres"]
transport = "stdio"
env = { "DATABASE_URL" = "${DATABASE_URL}" }

[mcp.servers.sqlite]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-sqlite"]
transport = "stdio"

# Cloud Services
[mcp.servers.aws]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-aws"]
transport = "stdio"
env = { 
    "AWS_ACCESS_KEY_ID" = "${AWS_ACCESS_KEY_ID}",
    "AWS_SECRET_ACCESS_KEY" = "${AWS_SECRET_ACCESS_KEY}",
    "AWS_REGION" = "${AWS_REGION}"
}
```

## Environment Variable Management

### .env File Support

Create a `.env` file in your project root:

```bash
# .env
# LLM Provider
AZURE_OPENAI_API_KEY=your-azure-api-key
AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com
AZURE_OPENAI_DEPLOYMENT=gpt-4

# MCP Tools
SEARCH_API_KEY=your-search-api-key
DATABASE_URL=postgresql://user:pass@localhost/db
GITHUB_TOKEN=your-github-token
BRAVE_API_KEY=your-brave-api-key

# AWS (if using AWS MCP server)
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_REGION=us-east-1
```

**Load environment variables:**
```go
import "github.com/joho/godotenv"

func init() {
    // Load .env file if it exists
    _ = godotenv.Load()
}
```

### Environment-Specific Configuration

Create different config files for different environments:

```bash
# Development
agentflow.dev.toml

# Staging  
agentflow.staging.toml

# Production
agentflow.prod.toml
```

**Load specific config:**
```go
config, err := core.LoadConfigFromFile("agentflow.prod.toml")
if err != nil {
    log.Fatal(err)
}
```

## Configuration Loading

### Automatic Loading

AgentFlow automatically looks for configuration in this order:

1. `agentflow.toml` in current directory
2. `agentflow.toml` in parent directories (up to project root)
3. Environment variables
4. Default values

```go
// Automatic loading
provider, err := core.NewProviderFromWorkingDir()
runner, err := core.NewRunnerFromWorkingDir()
```

### Explicit Configuration

```go
// Load from specific file
config, err := core.LoadConfig("path/to/agentflow.toml")
if err != nil {
    log.Fatal(err)
}

// Create provider from config
provider, err := core.NewProviderFromConfig(config)
if err != nil {
    log.Fatal(err)
}
```

### Programmatic Configuration

```go
// Create configuration in code
config := core.Config{
    Provider: core.ProviderConfig{
        Type:        "azure",
        APIKey:      os.Getenv("AZURE_OPENAI_API_KEY"),
        Endpoint:    os.Getenv("AZURE_OPENAI_ENDPOINT"),
        Deployment:  "gpt-4",
        MaxTokens:   2000,
        Temperature: 0.7,
    },
    MCP: core.MCPConfig{
        Enabled:      true,
        CacheEnabled: true,
        CacheTTL:     5 * time.Minute,
        Servers: map[string]core.MCPServerConfig{
            "search": {
                Command:   "npx",
                Args:      []string{"-y", "@modelcontextprotocol/server-web-search"},
                Transport: "stdio",
            },
        },
    },
}

// Use programmatic config
provider, err := core.NewProviderFromConfig(config)
mcpManager, err := core.InitializeMCPFromConfig(ctx, config.MCP)
```

## Configuration Validation

### Built-in Validation

AgentFlow validates configuration automatically:

```go
config, err := core.LoadConfig("agentflow.toml")
if err != nil {
    // Configuration errors are descriptive
    log.Printf("Configuration error: %v", err)
    // Example: "provider.api_key is required when type is 'azure'"
}
```

### Custom Validation

```go
func validateConfig(config core.Config) error {
    if config.Provider.Type == "azure" {
        if config.Provider.APIKey == "" {
            return fmt.Errorf("Azure provider requires API key")
        }
        if config.Provider.Endpoint == "" {
            return fmt.Errorf("Azure provider requires endpoint")
        }
    }
    
    return nil
}
```

## Dynamic Configuration

### Hot Reloading (Future Feature)

```toml
[config]
hot_reload = true
watch_files = ["agentflow.toml", ".env"]
```

### Runtime Configuration Updates

```go
// Update provider configuration at runtime
newConfig := core.ProviderConfig{
    Temperature: 0.5,  // Lower temperature for more deterministic responses
}

err := provider.UpdateConfig(newConfig)
if err != nil {
    log.Printf("Failed to update provider config: %v", err)
}
```

## Best Practices

### 1. Environment Variable Naming

Use consistent prefixes:

```bash
# AgentFlow settings
AGENTFLOW_LOG_LEVEL=debug
AGENTFLOW_QUEUE_SIZE=200

# Provider settings  
AZURE_OPENAI_API_KEY=...
OPENAI_API_KEY=...
OLLAMA_HOST=...

# Tool settings
SEARCH_API_KEY=...
DATABASE_URL=...
```

### 2. Security

**Never commit secrets:**
```bash
# .gitignore
.env
*.key
agentflow.prod.toml  # If it contains secrets
```

**Use environment variables for secrets:**
```toml
[provider]
api_key = "${AZURE_OPENAI_API_KEY}"  # Good
# api_key = "sk-actual-key-here"      # Never do this
```

### 3. Configuration Organization

**Separate concerns:**
```toml
# Base configuration
[provider]
type = "azure"
model = "gpt-4"

# Development overrides in agentflow.dev.toml
[provider]
type = "mock"
response = "Development response"

# Production overrides in agentflow.prod.toml  
[provider]
max_tokens = 4000
timeout = "60s"
```

### 4. Documentation

Document your configuration:

```toml
# agentflow.toml

# Primary LLM provider for all agents
[provider]
type = "azure"                    # Using Azure OpenAI for enterprise compliance
api_key = "${AZURE_OPENAI_API_KEY}"
deployment = "gpt-4"              # GPT-4 deployment for high-quality responses

# Tools available to all agents
[mcp]
enabled = true

# Web search for research agents
[mcp.servers.search]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-web-search"]
transport = "stdio"

# Docker management for DevOps agents
[mcp.servers.docker]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-docker"]
transport = "stdio"
```

## Troubleshooting

### Common Configuration Issues

**1. Environment Variables Not Loading**
```bash
# Check if environment variables are set
echo $AZURE_OPENAI_API_KEY

# Check if .env file is in the right location
ls -la .env
```

**2. MCP Servers Not Starting**
```bash
# Test MCP server manually
npx -y @modelcontextprotocol/server-web-search

# Check Node.js is installed
node --version
npm --version
```

**3. Configuration File Not Found**
```bash
# Check current directory
pwd
ls -la agentflow.toml

# Check if file has correct permissions
chmod 644 agentflow.toml
```

### Debugging Configuration

```go
// Enable debug logging to see configuration loading
config := core.Config{
    LogLevel: "debug",
}

// Print loaded configuration (sanitized)
fmt.Printf("Loaded config: %+v\n", config.Sanitized())
```

### Configuration Templates

Generate configuration templates:

```bash
# Generate basic template
agentcli config init

# Generate with specific provider
agentcli config init --provider azure

# Generate with MCP servers
agentcli config init --with-mcp

# Generate production template
agentcli config init --production
```

## Next Steps

- **[LLM Providers](Providers.md)** - Learn about specific provider configurations
- **[Tool Integration](ToolIntegration.md)** - Configure MCP servers for your agents  
- **[Production Deployment](Production.md)** - Production configuration patterns
