# Configuration API

**System configuration and setup**

This document covers AgentFlow's Configuration API, which provides comprehensive configuration management for agents, orchestration, memory systems, and MCP integration. The configuration system uses TOML files and provides validation and defaults.

## 📋 Core Configuration Types

### Main Configuration Structure

The main configuration structure loaded from `agentflow.toml`:

```go
type Config struct {
    AgentFlow struct {
        Name     string `toml:"name"`
        Version  string `toml:"version"`
        Provider string `toml:"provider"`
    } `toml:"agent_flow"`

    Logging struct {
        Level  string `toml:"level"`
        Format string `toml:"format"`
    } `toml:"logging"`

    Runtime struct {
        MaxConcurrentAgents int `toml:"max_concurrent_agents"`
        TimeoutSeconds      int `toml:"timeout_seconds"`
    } `toml:"runtime"`

    // Agent memory configuration
    AgentMemory AgentMemoryConfig `toml:"agent_memory"`

    // Error routing configuration
    ErrorRouting struct {
        Enabled              bool                     `toml:"enabled"`
        MaxRetries           int                      `toml:"max_retries"`
        RetryDelayMs         int                      `toml:"retry_delay_ms"`
        EnableCircuitBreaker bool                     `toml:"enable_circuit_breaker"`
        ErrorHandlerName     string                   `toml:"error_handler_name"`
        CategoryHandlers     map[string]string        `toml:"category_handlers"`
        SeverityHandlers     map[string]string        `toml:"severity_handlers"`
        CircuitBreaker       CircuitBreakerConfigToml `toml:"circuit_breaker"`
        Retry                RetryConfigToml          `toml:"retry"`
    } `toml:"error_routing"`

    // Provider-specific configurations
    Providers map[string]map[string]interface{} `toml:"providers"`

    // MCP configuration
    MCP MCPConfigToml `toml:"mcp"`

    // Orchestration configuration
    Orchestration OrchestrationConfigToml `toml:"orchestration"`
}
```

### Memory Configuration

Configuration for memory systems and RAG capabilities:

```go
type AgentMemoryConfig struct {
    // Core memory settings
    Provider   string `toml:\"provider\"`    // pgvector, weaviate, memory
    Connection string `toml:\"connection\"`  // Connection string
    MaxResults int    `toml:\"max_results\"` // Maximum search results
    Dimensions int    `toml:\"dimensions\"`  // Vector dimensions
    AutoEmbed  bool   `toml:\"auto_embed\"`  // Auto-generate embeddings

    // RAG settings
    EnableKnowledgeBase     bool    `toml:\"enable_knowledge_base\"`
    KnowledgeMaxResults     int     `toml:\"knowledge_max_results\"`
    KnowledgeScoreThreshold float32 `toml:\"knowledge_score_threshold\"`
    ChunkSize               int     `toml:\"chunk_size\"`
    ChunkOverlap            int     `toml:\"chunk_overlap\"`

    // RAG context assembly
    EnableRAG           bool    `toml:\"enable_rag\"`
    RAGMaxContextTokens int     `toml:\"rag_max_context_tokens\"`
    RAGPersonalWeight   float32 `toml:\"rag_personal_weight\"`
    RAGKnowledgeWeight  float32 `toml:\"rag_knowledge_weight\"`
    RAGIncludeSources   bool    `toml:\"rag_include_sources\"`

    // Document processing
    Documents DocumentConfig `toml:\"documents\"`
    
    // Embedding service
    Embedding EmbeddingConfig `toml:\"embedding\"`
    
    // Search configuration
    Search SearchConfigToml `toml:\"search\"`
}
```

### MCP Configuration

Configuration for Model Context Protocol integration:

```go
type MCPConfig struct {
    // Discovery settings
    EnableDiscovery  bool          `toml:\"enable_discovery\"`
    DiscoveryTimeout time.Duration `toml:\"discovery_timeout\"`
    ScanPorts        []int         `toml:\"scan_ports\"`

    // Connection settings
    ConnectionTimeout time.Duration `toml:\"connection_timeout\"`
    MaxRetries        int           `toml:\"max_retries\"`
    RetryDelay        time.Duration `toml:\"retry_delay\"`

    // Server configurations
    Servers []MCPServerConfig `toml:\"servers\"`

    // Performance settings
    EnableCaching  bool          `toml:\"enable_caching\"`
    CacheTimeout   time.Duration `toml:\"cache_timeout\"`
    MaxConnections int           `toml:\"max_connections\"`
}

type MCPServerConfig struct {
    Name    string `toml:\"name\"`
    Type    string `toml:\"type\"` // tcp, stdio, docker, websocket
    Host    string `toml:\"host,omitempty\"`
    Port    int    `toml:\"port,omitempty\"`
    Command string `toml:\"command,omitempty\"`
    Enabled bool   `toml:\"enabled\"`
}
```

## 🚀 Basic Configuration

### TOML Configuration Files

AgentFlow supports TOML configuration files for easy setup:

```toml
# agentflow.toml
[agent_flow]
name = "my-agent-system"
version = "1.0.0"
provider = "openai"

[logging]
level = "info"
format = "json"

[runtime]
max_concurrent_agents = 10
timeout_seconds = 30

[agent_memory]
provider = "pgvector"
connection = "postgres://user:password@localhost:5432/agentflow"
max_results = 10
dimensions = 1536
auto_embed = true
enable_knowledge_base = true
enable_rag = true
chunk_size = 1000
chunk_overlap = 200
rag_max_context_tokens = 4000
rag_personal_weight = 0.3
rag_knowledge_weight = 0.7
rag_include_sources = true

[agent_memory.documents]
supported_types = ["pdf", "txt", "md", "web", "code"]
max_file_size = "10MB"
auto_chunk = true
enable_metadata_extraction = true
enable_url_scraping = true

[agent_memory.embedding]
provider = "azure"
model = "text-embedding-ada-002"
max_batch_size = 100
timeout_seconds = 30
cache_embeddings = true

[agent_memory.search]
hybrid_search = true
keyword_weight = 0.3
semantic_weight = 0.7

[error_routing]
enabled = true
max_retries = 3
retry_delay_ms = 1000
enable_circuit_breaker = true
error_handler_name = "default-error-handler"

[error_routing.circuit_breaker]
failure_threshold = 5
success_threshold = 3
timeout_ms = 30000
reset_timeout_ms = 60000
half_open_max_calls = 3

[error_routing.retry]
max_retries = 3
base_delay_ms = 1000
max_delay_ms = 30000
backoff_factor = 2.0
enable_jitter = true

[providers.openai]
# API key will be read from OPENAI_API_KEY environment variable

[providers.azure]
# API key will be read from AZURE_OPENAI_API_KEY environment variable
# Endpoint will be read from AZURE_OPENAI_ENDPOINT environment variable
# Deployment will be read from AZURE_OPENAI_DEPLOYMENT environment variable

[providers.ollama]
endpoint = "http://localhost:11434"
model = "llama3.2:latest"

[mcp]
enabled = true
enable_discovery = true
discovery_timeout_ms = 10000
scan_ports = [8080, 8081, 8090, 3000]
connection_timeout_ms = 30000
max_retries = 3
retry_delay_ms = 1000
enable_caching = true
cache_timeout_ms = 300000
max_connections = 10

[[mcp.servers]]
name = "web-search"
type = "tcp"
host = "localhost"
port = 8080
enabled = true

[[mcp.servers]]
name = "file-operations"
type = "stdio"
command = "python -m file_server"
enabled = true

[orchestration]
mode = "collaborative"
timeout_seconds = 30
max_iterations = 5
sequential_agents = ["agent1", "agent2", "agent3"]
collaborative_agents = ["researcher", "analyzer"]
loop_agent = "quality-checker"
```

### Loading Configuration

```go
package main

import (
    "fmt"
    "github.com/kunalkushwaha/agentflow/core"
)

func loadConfigurationExample() {
    // Load configuration from TOML file
    config, err := core.LoadConfig("agentflow.toml")
    if err != nil {
        panic(fmt.Sprintf("Failed to load configuration: %v", err))
    }
    
    // Or load from working directory (looks for agentflow.toml)
    config, err = core.LoadConfigFromWorkingDir()
    if err != nil {
        panic(fmt.Sprintf("Failed to load configuration: %v", err))
    }
    
    // Apply logging configuration
    config.ApplyLoggingConfig()
    
    // Initialize LLM provider from configuration
    provider, err := config.InitializeProvider()
    if err != nil {
        panic(fmt.Sprintf("Failed to initialize provider: %v", err))
    }
    
    // Or use the convenience function
    provider, err = core.NewProviderFromWorkingDir()
    if err != nil {
        panic(fmt.Sprintf("Failed to initialize provider: %v", err))
    }
    
    // Initialize memory if configured
    if config.AgentMemory.Provider != "" {
        memory, err := core.NewMemory(config.AgentMemory)
        if err != nil {
            panic(fmt.Sprintf("Failed to initialize memory: %v", err))
        }
        fmt.Printf("Memory initialized with provider: %s\n", config.AgentMemory.Provider)
    }
    
    // Initialize MCP if enabled
    if config.MCP.Enabled {
        mcpConfig := config.GetMCPConfig()
        err = core.InitializeMCP(mcpConfig)
        if err != nil {
            panic(fmt.Sprintf("Failed to initialize MCP: %v", err))
        }
        fmt.Println("MCP initialized")
    }
    
    // Create runner from configuration
    runner, err := core.NewRunnerFromConfig("agentflow.toml")
    if err != nil {
        panic(fmt.Sprintf("Failed to create runner: %v", err))
    }
    
    fmt.Println("System initialized with configuration")
}
```

### Environment Variables

AgentFlow supports environment variables for sensitive configuration like API keys:

```bash
# LLM Provider Configuration
export OPENAI_API_KEY=your-openai-key
export AZURE_OPENAI_API_KEY=your-azure-key
export AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/
export AZURE_OPENAI_DEPLOYMENT=your-deployment-name

# Database Configuration (for memory providers)
export DATABASE_URL=postgres://user:password@localhost:5432/agentflow

# Optional: Override default settings
export AGENTFLOW_LOG_LEVEL=debug
export AGENTFLOW_TIMEOUT=60
```

These environment variables are automatically used when not specified in the TOML file.

## 🔧 Configuration Examples

### Development vs Production

**Development Configuration (`dev.toml`):**
```toml
[agent_flow]
name = "my-dev-agent"
version = "1.0.0"
provider = "openai"

[logging]
level = "debug"
format = "text"

[agent_memory]
provider = "memory"  # In-memory for fast development
connection = "memory"
max_results = 5
dimensions = 384     # Smaller for faster testing

[agent_memory.embedding]
provider = "dummy"   # No API calls in development
```

**Production Configuration (`prod.toml`):**
```toml
[agent_flow]
name = "my-prod-agent"
version = "1.0.0"
provider = "azure"

[logging]
level = "info"
format = "json"

[runtime]
max_concurrent_agents = 20
timeout_seconds = 60

[agent_memory]
provider = "pgvector"
connection = "postgres://user:pass@db:5432/agentflow"
max_results = 20
dimensions = 1536
enable_knowledge_base = true
enable_rag = true

[agent_memory.embedding]
provider = "azure"
model = "text-embedding-ada-002"
cache_embeddings = true

[mcp]
enabled = true
enable_discovery = false  # Use explicit servers in production
max_connections = 20

[[mcp.servers]]
name = "search-service"
type = "tcp"
host = "search-service"
port = 8080
enabled = true
```

### Multiple Configuration Files

You can use different configuration files for different environments:

```bash
# Development
go run . --config dev.toml

# Production  
go run . --config prod.toml

# Or use environment variable
export AGENTFLOW_CONFIG=prod.toml
go run .
```

## 🔧 Configuration Validation

AgentFlow automatically validates your configuration when loading. Common validation errors:

- **Invalid provider**: Must be one of `openai`, `azure`, `ollama`
- **Missing API keys**: Set environment variables for your chosen provider
- **Invalid memory provider**: Must be one of `memory`, `pgvector`, `weaviate`
- **Invalid orchestration mode**: Must be one of `route`, `collaborative`, `sequential`, `loop`, `mixed`

```bash
# Test your configuration
go run . --validate-config

# Or check configuration loading
go run . --dry-run
```

## 📚 Best Practices

### 1. Keep It Simple
- Use `agentflow.toml` for all configuration
- Store sensitive data in environment variables
- Use different config files for different environments

### 2. Security
- Never commit API keys to version control
- Use environment variables for secrets:
  ```bash
  export OPENAI_API_KEY=your-key
  export DATABASE_URL=postgres://...
  ```

### 3. Environment Management
```bash
# Development
cp agentflow.dev.toml agentflow.toml

# Production
cp agentflow.prod.toml agentflow.toml
```

### 4. Common Configurations

**Minimal Configuration:**
```toml
[agent_flow]
name = "my-agent"
provider = "openai"
```

**Full-Featured Configuration:**
```toml
[agent_flow]
name = "my-agent"
provider = "azure"

[agent_memory]
provider = "pgvector"
enable_rag = true

[mcp]
enabled = true

[orchestration]
mode = "collaborative"
```

This covers the essential configuration patterns you'll need for AgentFlow systems.