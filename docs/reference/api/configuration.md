# Configuration API

**System configuration and setup**

This document covers AgenticGoKit's Configuration API, which provides comprehensive configuration management for agents, orchestration, memory systems, and MCP integration. The configuration system uses TOML files and provides validation and defaults.

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
type MCPConfigToml struct {
    Enabled           bool                  `toml:"enabled"`
    EnableDiscovery   bool                  `toml:"enable_discovery"`
    DiscoveryTimeout  int                   `toml:"discovery_timeout_ms"`
    ScanPorts         []int                 `toml:"scan_ports"`
    ConnectionTimeout int                   `toml:"connection_timeout_ms"`
    MaxRetries        int                   `toml:"max_retries"`
    RetryDelay        int                   `toml:"retry_delay_ms"`
    EnableCaching     bool                  `toml:"enable_caching"`
    CacheTimeout      int                   `toml:"cache_timeout_ms"`
    MaxConnections    int                   `toml:"max_connections"`
    Servers           []MCPServerConfigToml `toml:"servers"`
}

type MCPServerConfigToml struct {
    Name    string `toml:"name"`
    Type    string `toml:"type"` // tcp, stdio, docker, websocket
    Host    string `toml:"host,omitempty"`
    Port    int    `toml:"port,omitempty"`
    Command string `toml:"command,omitempty"` // for stdio transport
    Enabled bool   `toml:"enabled"`
}
```

## 🚀 Basic Configuration

### TOML Configuration Files

AgentFlow supports TOML configuration files for easy setup:

```toml
# agentflow.toml - Basic Configuration
[agent_flow]
name = "my-agent-system"
version = "1.0.0"
provider = "azure"

[logging]
level = "info"
format = "json"

[runtime]
max_concurrent_agents = 10
timeout_seconds = 30

[providers.azure]
# API key will be read from AZURE_OPENAI_API_KEY environment variable
# Endpoint will be read from AZURE_OPENAI_ENDPOINT environment variable
# Deployment will be read from AZURE_OPENAI_DEPLOYMENT environment variable

[providers.openai]
# API key will be read from OPENAI_API_KEY environment variable

[providers.ollama]
endpoint = "http://localhost:11434"
model = "gemma3:1b"

# Ensure you import the LLM plugin so the provider registers:
# import (
#   _ "github.com/kunalkushwaha/agenticgokit/plugins/llm/ollama"
#   _ "github.com/kunalkushwaha/agenticgokit/plugins/logging/zerolog"
#   _ "github.com/kunalkushwaha/agenticgokit/plugins/orchestrator/default"
#   _ "github.com/kunalkushwaha/agenticgokit/plugins/runner/default"
# )

[providers.ollama]
base_url = "http://localhost:11434"
model = "gemma3:1b"
```

**With Memory and RAG Configuration:**
```toml
# agentflow.toml - With Memory and RAG
[agent_flow]
name = "my-rag-system"
version = "1.0.0"
provider = "azure"

[logging]
level = "info"
format = "text"

[runtime]
max_concurrent_agents = 5
timeout_seconds = 30

# RAG-Enhanced Memory Configuration
[agent_memory]
# Core memory settings
provider = "memory"        # Options: memory, pgvector, weaviate
connection = "memory"      # Connection string (for database providers)
max_results = 10          # Maximum results per query
dimensions = 1536         # Embedding dimensions
auto_embed = true         # Automatic embedding generation

# RAG-enhanced settings
enable_knowledge_base = true        # Enable knowledge base functionality
knowledge_max_results = 20          # Maximum results from knowledge base
knowledge_score_threshold = 0.7     # Minimum relevance score for results
chunk_size = 1000                  # Document chunk size in characters
chunk_overlap = 200                # Overlap between chunks in characters

# RAG context assembly
enable_rag = true                  # Enable RAG context building
rag_max_context_tokens = 4000      # Maximum tokens in assembled context
rag_personal_weight = 0.3          # Weight for personal memory (0.0-1.0)
rag_knowledge_weight = 0.7         # Weight for knowledge base (0.0-1.0)
rag_include_sources = true         # Include source attribution in context

# Document processing
[agent_memory.documents]
auto_chunk = true                           # Automatically chunk large documents
supported_types = ["pdf", "txt", "md", "web", "code", "json"]  # Supported file types
max_file_size = "10MB"                     # Maximum file size for processing
enable_metadata_extraction = true          # Extract metadata from documents
enable_url_scraping = true                 # Enable web scraping for URLs

# Embedding service
[agent_memory.embedding]
provider = "azure"                    # Options: azure, openai, local
model = "text-embedding-ada-002"     # Embedding model to use
cache_embeddings = true              # Cache embeddings for performance
max_batch_size = 100                 # Maximum batch size for embeddings
timeout_seconds = 30                 # Request timeout in seconds

# Search configuration
[agent_memory.search]
hybrid_search = true           # Enable hybrid search (semantic + keyword)
keyword_weight = 0.3          # Weight for keyword search (BM25)
semantic_weight = 0.7         # Weight for semantic search (vector similarity)
enable_reranking = false      # Enable advanced re-ranking
enable_query_expansion = false # Enable query expansion for better results

[providers.azure]
# API key will be read from AZURE_OPENAI_API_KEY environment variable
# Endpoint will be read from AZURE_OPENAI_ENDPOINT environment variable
# Deployment will be read from AZURE_OPENAI_DEPLOYMENT environment variable
```

**With MCP Integration:**
```toml
# agentflow.toml - With MCP Integration
[agent_flow]
name = "my-mcp-system"
version = "1.0.0"
provider = "azure"

[logging]
level = "info"
format = "json"

[runtime]
max_concurrent_agents = 10
timeout_seconds = 30

[providers.azure]
# API key will be read from AZURE_OPENAI_API_KEY environment variable
# Endpoint will be read from AZURE_OPENAI_ENDPOINT environment variable
# Deployment will be read from AZURE_OPENAI_DEPLOYMENT environment variable

[mcp]
enabled = true
enable_discovery = true
connection_timeout = 5000
max_retries = 3
retry_delay = 1000
enable_caching = true
cache_timeout = 300000
max_connections = 10

# Example MCP servers - configure as needed
[[mcp.servers]]
name = "docker"
type = "tcp"
host = "localhost"
port = 8811
enabled = false

[[mcp.servers]]
name = "filesystem"
type = "stdio"
command = "npx @modelcontextprotocol/server-filesystem /path/to/allowed/files"
enabled = false

[[mcp.servers]]
name = "brave-search"
type = "stdio"
command = "npx @modelcontextprotocol/server-brave-search"
enabled = false
```

### Loading Configuration

```go
package main

import (
    "fmt"
    "github.com/kunalkushwaha/agenticgokit/core"
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
        // MCP initialization would be handled by the framework
        fmt.Println("MCP configuration loaded")
    }
    
    fmt.Println("System initialized with configuration")
}
```

### Environment Variables

AgenticGoKit supports environment variables for sensitive configuration like API keys:

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
provider = "openai"  # Use OpenAI provider

[logging]
level = "debug"
format = "text"

[runtime]
max_concurrent_agents = 2
timeout_seconds = 10

[providers.openai]
# API key will be read from OPENAI_API_KEY environment variable

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

[providers.azure]
# API key will be read from AZURE_OPENAI_API_KEY environment variable
# Endpoint will be read from AZURE_OPENAI_ENDPOINT environment variable
# Deployment will be read from AZURE_OPENAI_DEPLOYMENT environment variable

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

AgenticGoKit automatically validates your configuration when loading. Common validation errors:

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

This covers the essential configuration patterns you'll need for AgenticGoKit systems.