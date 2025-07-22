# Configuration API

**System configuration and setup**

This document covers AgenticGoKit's Configuration API, which provides comprehensive configuration management for agents, orchestration, memory systems, and MCP integration. The configuration system supports multiple formats and provides validation and defaults.

## 📋 Core Configuration Types

### Runner Configuration

The main configuration for agent runners and orchestration:

```go
type RunnerConfig struct {
    // Agent configuration
    Agents    map[string]AgentHandler `toml:"-"`
    Memory    Memory                  `toml:"-"`
    SessionID string                  `toml:"session_id"`
    
    // Orchestration settings
    OrchestrationMode   OrchestrationMode   `toml:"orchestration_mode"`
    CollaborativeAgents []string            `toml:"collaborative_agents"`
    SequentialAgents    []string            `toml:"sequential_agents"`
    
    // Performance settings
    Timeout        time.Duration `toml:"timeout"`
    MaxConcurrency int           `toml:"max_concurrency"`
    
    // Error handling
    FailureThreshold float64      `toml:"failure_threshold"`
    RetryPolicy      *RetryPolicy `toml:"retry_policy"`
}

type EnhancedRunnerConfig struct {
    RunnerConfig
    OrchestrationMode   OrchestrationMode   `toml:"orchestration_mode"`
    Config              OrchestrationConfig `toml:"orchestration_config"`
    CollaborativeAgents []string            `toml:"collaborative_agents"`
    SequentialAgents    []string            `toml:"sequential_agents"`
}
```

### Memory Configuration

Configuration for memory systems and RAG capabilities:

```go
type AgentMemoryConfig struct {
    // Core memory settings
    Provider   string `toml:"provider"`    // pgvector, weaviate, memory
    Connection string `toml:"connection"`  // Connection string
    MaxResults int    `toml:"max_results"` // Maximum search results
    Dimensions int    `toml:"dimensions"`  // Vector dimensions
    AutoEmbed  bool   `toml:"auto_embed"`  // Auto-generate embeddings

    // RAG settings
    EnableKnowledgeBase     bool    `toml:"enable_knowledge_base"`
    KnowledgeMaxResults     int     `toml:"knowledge_max_results"`
    KnowledgeScoreThreshold float32 `toml:"knowledge_score_threshold"`
    ChunkSize               int     `toml:"chunk_size"`
    ChunkOverlap            int     `toml:"chunk_overlap"`

    // RAG context assembly
    EnableRAG           bool    `toml:"enable_rag"`
    RAGMaxContextTokens int     `toml:"rag_max_context_tokens"`
    RAGPersonalWeight   float32 `toml:"rag_personal_weight"`
    RAGKnowledgeWeight  float32 `toml:"rag_knowledge_weight"`
    RAGIncludeSources   bool    `toml:"rag_include_sources"`

    // Document processing
    Documents DocumentConfig `toml:"documents"`
    
    // Embedding service
    Embedding EmbeddingConfig `toml:"embedding"`
    
    // Search configuration
    Search SearchConfigToml `toml:"search"`
}
```

### MCP Configuration

Configuration for Model Context Protocol integration:

```go
type MCPConfig struct {
    // Discovery settings
    EnableDiscovery  bool          `toml:"enable_discovery"`
    DiscoveryTimeout time.Duration `toml:"discovery_timeout"`
    ScanPorts        []int         `toml:"scan_ports"`

    // Connection settings
    ConnectionTimeout time.Duration `toml:"connection_timeout"`
    MaxRetries        int           `toml:"max_retries"`
    RetryDelay        time.Duration `toml:"retry_delay"`

    // Server configurations
    Servers []MCPServerConfig `toml:"servers"`

    // Performance settings
    EnableCaching  bool          `toml:"enable_caching"`
    CacheTimeout   time.Duration `toml:"cache_timeout"`
    MaxConnections int           `toml:"max_connections"`
}

type MCPServerConfig struct {
    Name    string `toml:"name"`
    Type    string `toml:"type"` // tcp, stdio, docker, websocket
    Host    string `toml:"host,omitempty"`
    Port    int    `toml:"port,omitempty"`
    Command string `toml:"command,omitempty"`
    Enabled bool   `toml:"enabled"`
}
```

## 🚀 Basic Configuration

### TOML Configuration Files

AgenticGoKit supports TOML configuration files for easy setup:

```toml
# agentflow.toml
[runner]
session_id = "main-session"
orchestration_mode = "collaborate"
timeout = "30s"
max_concurrency = 10
failure_threshold = 0.7

[runner.retry_policy]
max_retries = 3
backoff_factor = 1.5
max_delay = "10s"

[memory]
provider = "pgvector"
connection = "postgres://user:password@localhost:5432/agentflow"
max_results = 10
dimensions = 1536
auto_embed = true
enable_knowledge_base = true
chunk_size = 1000
chunk_overlap = 200

[memory.embedding]
provider = "openai"
model = "text-embedding-3-small"
cache_embeddings = true

[mcp]
enable_discovery = true
discovery_timeout = "10s"
scan_ports = [8080, 8081, 8090, 3000]
connection_timeout = "30s"
max_retries = 3
retry_delay = "1s"
enable_caching = true
cache_timeout = "15m"

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
    config, err := core.LoadConfigFromFile("agentflow.toml")
    if err != nil {
        panic(fmt.Sprintf("Failed to load configuration: %v", err))
    }
    
    // Initialize systems with loaded configuration
    
    // Initialize memory
    memory, err := core.NewMemory(config.Memory)
    if err != nil {
        panic(fmt.Sprintf("Failed to initialize memory: %v", err))
    }
    
    // Initialize MCP
    err = core.InitializeMCP(config.MCP)
    if err != nil {
        panic(fmt.Sprintf("Failed to initialize MCP: %v", err))
    }
    
    // Create runner with configuration
    runnerConfig := core.EnhancedRunnerConfig{
        RunnerConfig: core.RunnerConfig{
            Memory:    memory,
            SessionID: config.Runner.SessionID,
        },
        OrchestrationMode: config.Runner.OrchestrationMode,
        Config: core.OrchestrationConfig{
            Timeout:          config.Runner.Timeout,
            MaxConcurrency:   config.Runner.MaxConcurrency,
            FailureThreshold: config.Runner.FailureThreshold,
            RetryPolicy:      config.Runner.RetryPolicy,
        },
    }
    
    runner := core.NewRunnerWithOrchestration(runnerConfig)
    
    fmt.Println("System initialized with configuration")
}
```

### Programmatic Configuration

```go
func programmaticConfigExample() {
    // Create memory configuration
    memoryConfig := core.AgentMemoryConfig{
        Provider:   "pgvector",
        Connection: "postgres://user:pass@localhost:5432/agents",
        MaxResults: 10,
        Dimensions: 1536,
        AutoEmbed:  true,
        
        // RAG configuration
        EnableKnowledgeBase:     true,
        KnowledgeMaxResults:     20,
        KnowledgeScoreThreshold: 0.7,
        ChunkSize:               1000,
        ChunkOverlap:            200,
        
        // Embedding configuration
        Embedding: core.EmbeddingConfig{
            Provider: "openai",
            Model:    "text-embedding-3-small",
            APIKey:   "your-api-key",
        },
    }
    
    // Create MCP configuration
    mcpConfig := core.MCPConfig{
        EnableDiscovery:   true,
        DiscoveryTimeout:  10 * time.Second,
        ConnectionTimeout: 30 * time.Second,
        MaxRetries:        3,
        RetryDelay:        1 * time.Second,
        EnableCaching:     true,
        CacheTimeout:      15 * time.Minute,
        
        Servers: []core.MCPServerConfig{
            {
                Name:    "search-server",
                Type:    "tcp",
                Host:    "localhost",
                Port:    8080,
                Enabled: true,
            },
        },
    }
    
    // Create orchestration configuration
    orchestrationConfig := core.OrchestrationConfig{
        Timeout:          30 * time.Second,
        MaxConcurrency:   10,
        FailureThreshold: 0.7,
        RetryPolicy: &core.RetryPolicy{
            MaxRetries:    3,
            BackoffFactor: 1.5,
            MaxDelay:      10 * time.Second,
        },
    }
    
    // Initialize systems
    memory, err := core.NewMemory(memoryConfig)
    if err != nil {
        panic(err)
    }
    
    err = core.InitializeMCP(mcpConfig)
    if err != nil {
        panic(err)
    }
    
    // Create agents
    agents := map[string]core.AgentHandler{
        "processor": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            return core.AgentResult{
                Data: map[string]interface{}{
                    "processed": true,
                },
            }, nil
        }),
    }
    
    // Create runner with full configuration
    runnerConfig := core.EnhancedRunnerConfig{
        RunnerConfig: core.RunnerConfig{
            Agents:    agents,
            Memory:    memory,
            SessionID: "configured-session",
        },
        OrchestrationMode: core.OrchestrationCollaborate,
        Config:            orchestrationConfig,
    }
    
    runner := core.NewRunnerWithOrchestration(runnerConfig)
    
    // Use the configured runner
    event := core.NewEvent("test", map[string]interface{}{
        "input": "test data",
    })
    
    results, err := runner.ProcessEvent(context.Background(), event)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Results: %+v\n", results)
}
```

## 🔧 Advanced Configuration

### Environment-Based Configuration

```go
func environmentConfigExample() {
    // Create configuration with environment variable support
    config := core.AgentMemoryConfig{
        Provider:   getEnvOrDefault("AGENTFLOW_MEMORY_PROVIDER", "memory"),
        Connection: getEnvOrDefault("AGENTFLOW_MEMORY_CONNECTION", "memory"),
        MaxResults: getEnvIntOrDefault("AGENTFLOW_MEMORY_MAX_RESULTS", 10),
        Dimensions: getEnvIntOrDefault("AGENTFLOW_MEMORY_DIMENSIONS", 1536),
        AutoEmbed:  getEnvBoolOrDefault("AGENTFLOW_MEMORY_AUTO_EMBED", true),
        
        Embedding: core.EmbeddingConfig{
            Provider: getEnvOrDefault("AGENTFLOW_EMBEDDING_PROVIDER", "openai"),
            Model:    getEnvOrDefault("AGENTFLOW_EMBEDDING_MODEL", "text-embedding-3-small"),
            APIKey:   getEnvOrDefault("OPENAI_API_KEY", ""),
        },
    }
    
    // Validate configuration
    if err := validateMemoryConfig(config); err != nil {
        panic(fmt.Sprintf("Invalid configuration: %v", err))
    }
    
    memory, err := core.NewMemory(config)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Memory initialized with provider: %s\n", config.Provider)
}

// Helper functions for environment variables
func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        if boolValue, err := strconv.ParseBool(value); err == nil {
            return boolValue
        }
    }
    return defaultValue
}

func validateMemoryConfig(config core.AgentMemoryConfig) error {
    if config.Provider == "" {
        return fmt.Errorf("memory provider is required")
    }
    
    if config.Provider == "pgvector" && config.Connection == "" {
        return fmt.Errorf("connection string is required for pgvector provider")
    }
    
    if config.Dimensions <= 0 {
        return fmt.Errorf("dimensions must be positive")
    }
    
    return nil
}
```

### Configuration Profiles

```go
// Configuration profiles for different environments
type ConfigProfile struct {
    Name        string
    Description string
    Memory      core.AgentMemoryConfig
    MCP         core.MCPConfig
    Runner      core.EnhancedRunnerConfig
}

func configurationProfilesExample() {
    profiles := map[string]ConfigProfile{
        "development": {
            Name:        "Development",
            Description: "Local development with in-memory storage",
            Memory: core.AgentMemoryConfig{
                Provider:   "memory",
                Connection: "memory",
                MaxResults: 5,
                Dimensions: 384, // Smaller for faster testing
                AutoEmbed:  true,
                Embedding: core.EmbeddingConfig{
                    Provider: "dummy", // No API calls in development
                },
            },
            MCP: core.MCPConfig{
                EnableDiscovery:   true,
                DiscoveryTimeout:  5 * time.Second,
                ConnectionTimeout: 10 * time.Second,
                MaxRetries:        1,
                EnableCaching:     false, // Disable caching for development
            },
        },
        
        "testing": {
            Name:        "Testing",
            Description: "Testing environment with mock services",
            Memory: core.AgentMemoryConfig{
                Provider:   "memory",
                Connection: "memory",
                MaxResults: 10,
                Dimensions: 384,
                AutoEmbed:  true,
                Embedding: core.EmbeddingConfig{
                    Provider: "dummy",
                },
            },
            MCP: core.MCPConfig{
                EnableDiscovery:   false, // No discovery in tests
                ConnectionTimeout: 5 * time.Second,
                MaxRetries:        1,
                EnableCaching:     true,
                CacheTimeout:      1 * time.Minute, // Short cache for testing
            },
        },
        
        "production": {
            Name:        "Production",
            Description: "Production environment with full features",
            Memory: core.AgentMemoryConfig{
                Provider:   "pgvector",
                Connection: "postgres://user:pass@db:5432/agentflow",
                MaxResults: 20,
                Dimensions: 1536,
                AutoEmbed:  true,
                EnableKnowledgeBase: true,
                ChunkSize:  1000,
                ChunkOverlap: 200,
                Embedding: core.EmbeddingConfig{
                    Provider:        "openai",
                    Model:           "text-embedding-3-small",
                    CacheEmbeddings: true,
                    MaxBatchSize:    100,
                    TimeoutSeconds:  30,
                },
            },
            MCP: core.MCPConfig{
                EnableDiscovery:   false, // Explicit servers in production
                ConnectionTimeout: 30 * time.Second,
                MaxRetries:        3,
                RetryDelay:        1 * time.Second,
                EnableCaching:     true,
                CacheTimeout:      15 * time.Minute,
                MaxConnections:    20,
                Servers: []core.MCPServerConfig{
                    {
                        Name:    "search-service",
                        Type:    "tcp",
                        Host:    "search-service",
                        Port:    8080,
                        Enabled: true,
                    },
                    {
                        Name:    "data-service",
                        Type:    "tcp",
                        Host:    "data-service",
                        Port:    8081,
                        Enabled: true,
                    },
                },
            },
        },
    }
    
    // Select profile based on environment
    profileName := getEnvOrDefault("AGENTFLOW_PROFILE", "development")
    profile, exists := profiles[profileName]
    if !exists {
        panic(fmt.Sprintf("Unknown profile: %s", profileName))
    }
    
    fmt.Printf("Using profile: %s - %s\n", profile.Name, profile.Description)
    
    // Initialize with profile configuration
    memory, err := core.NewMemory(profile.Memory)
    if err != nil {
        panic(err)
    }
    
    err = core.InitializeMCP(profile.MCP)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("System initialized with %s profile\n", profile.Name)
}
```

### Dynamic Configuration Updates

```go
type ConfigManager struct {
    currentConfig *core.AgentMemoryConfig
    watchers      []func(*core.AgentMemoryConfig)
    mutex         sync.RWMutex
}

func NewConfigManager(initialConfig *core.AgentMemoryConfig) *ConfigManager {
    return &ConfigManager{
        currentConfig: initialConfig,
        watchers:      make([]func(*core.AgentMemoryConfig), 0),
    }
}

func (cm *ConfigManager) GetConfig() *core.AgentMemoryConfig {
    cm.mutex.RLock()
    defer cm.mutex.RUnlock()
    
    // Return a copy to prevent external modification
    config := *cm.currentConfig
    return &config
}

func (cm *ConfigManager) UpdateConfig(newConfig *core.AgentMemoryConfig) error {
    cm.mutex.Lock()
    defer cm.mutex.Unlock()
    
    // Validate new configuration
    if err := validateMemoryConfig(*newConfig); err != nil {
        return fmt.Errorf("invalid configuration: %w", err)
    }
    
    // Update configuration
    cm.currentConfig = newConfig
    
    // Notify watchers
    for _, watcher := range cm.watchers {
        go watcher(newConfig) // Run in goroutine to avoid blocking
    }
    
    return nil
}

func (cm *ConfigManager) AddWatcher(watcher func(*core.AgentMemoryConfig)) {
    cm.mutex.Lock()
    defer cm.mutex.Unlock()
    cm.watchers = append(cm.watchers, watcher)
}

func dynamicConfigExample() {
    // Initial configuration
    initialConfig := &core.AgentMemoryConfig{
        Provider:   "memory",
        MaxResults: 10,
        Dimensions: 384,
        AutoEmbed:  true,
    }
    
    configManager := NewConfigManager(initialConfig)
    
    // Add watcher for configuration changes
    configManager.AddWatcher(func(newConfig *core.AgentMemoryConfig) {
        fmt.Printf("Configuration updated: Provider=%s, MaxResults=%d\n", 
            newConfig.Provider, newConfig.MaxResults)
        
        // Reinitialize memory system with new configuration
        memory, err := core.NewMemory(*newConfig)
        if err != nil {
            fmt.Printf("Failed to reinitialize memory: %v\n", err)
            return
        }
        
        // Update global memory instance (implementation specific)
        updateGlobalMemory(memory)
    })
    
    // Simulate configuration updates
    time.Sleep(2 * time.Second)
    
    // Update to production configuration
    productionConfig := &core.AgentMemoryConfig{
        Provider:   "pgvector",
        Connection: "postgres://localhost:5432/agentflow",
        MaxResults: 20,
        Dimensions: 1536,
        AutoEmbed:  true,
    }
    
    err := configManager.UpdateConfig(productionConfig)
    if err != nil {
        fmt.Printf("Failed to update configuration: %v\n", err)
    }
    
    // Configuration change will trigger watcher
    time.Sleep(1 * time.Second)
}

func updateGlobalMemory(memory core.Memory) {
    // Implementation would update the global memory instance
    // This is a placeholder for the actual implementation
    fmt.Println("Global memory instance updated")
}
```

## 🔧 Configuration Validation

### Schema Validation

```go
type ConfigValidator struct {
    rules []ValidationRule
}

type ValidationRule interface {
    Validate(config interface{}) error
    Name() string
}

type MemoryProviderRule struct{}

func (r *MemoryProviderRule) Name() string {
    return "memory_provider"
}

func (r *MemoryProviderRule) Validate(config interface{}) error {
    memConfig, ok := config.(*core.AgentMemoryConfig)
    if !ok {
        return fmt.Errorf("expected AgentMemoryConfig")
    }
    
    validProviders := []string{"memory", "pgvector", "weaviate"}
    for _, provider := range validProviders {
        if memConfig.Provider == provider {
            return nil
        }
    }
    
    return fmt.Errorf("invalid memory provider: %s, must be one of %v", 
        memConfig.Provider, validProviders)
}

type ConnectionStringRule struct{}

func (r *ConnectionStringRule) Name() string {
    return "connection_string"
}

func (r *ConnectionStringRule) Validate(config interface{}) error {
    memConfig, ok := config.(*core.AgentMemoryConfig)
    if !ok {
        return fmt.Errorf("expected AgentMemoryConfig")
    }
    
    if memConfig.Provider == "pgvector" {
        if memConfig.Connection == "" {
            return fmt.Errorf("connection string is required for pgvector provider")
        }
        
        // Validate PostgreSQL connection string format
        if !strings.HasPrefix(memConfig.Connection, "postgres://") {
            return fmt.Errorf("invalid PostgreSQL connection string format")
        }
    }
    
    return nil
}

func configValidationExample() {
    validator := &ConfigValidator{
        rules: []ValidationRule{
            &MemoryProviderRule{},
            &ConnectionStringRule{},
        },
    }
    
    // Test valid configuration
    validConfig := &core.AgentMemoryConfig{
        Provider:   "pgvector",
        Connection: "postgres://user:pass@localhost:5432/db",
        MaxResults: 10,
        Dimensions: 1536,
    }
    
    err := validator.ValidateConfig(validConfig)
    if err != nil {
        fmt.Printf("Valid config failed validation: %v\n", err)
    } else {
        fmt.Println("Valid configuration passed validation")
    }
    
    // Test invalid configuration
    invalidConfig := &core.AgentMemoryConfig{
        Provider:   "invalid_provider",
        Connection: "",
        MaxResults: 10,
        Dimensions: 1536,
    }
    
    err = validator.ValidateConfig(invalidConfig)
    if err != nil {
        fmt.Printf("Invalid config failed validation (expected): %v\n", err)
    }
}

func (cv *ConfigValidator) ValidateConfig(config interface{}) error {
    var errors []string
    
    for _, rule := range cv.rules {
        if err := rule.Validate(config); err != nil {
            errors = append(errors, fmt.Sprintf("%s: %v", rule.Name(), err))
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
    }
    
    return nil
}
```

### Configuration Testing

```go
func TestConfigurationLoading(t *testing.T) {
    // Create temporary config file
    configContent := `
[memory]
provider = "memory"
max_results = 10
dimensions = 384
auto_embed = true

[mcp]
enable_discovery = true
discovery_timeout = "5s"
`
    
    tmpFile, err := os.CreateTemp("", "test-config-*.toml")
    require.NoError(t, err)
    defer os.Remove(tmpFile.Name())
    
    _, err = tmpFile.WriteString(configContent)
    require.NoError(t, err)
    tmpFile.Close()
    
    // Load configuration
    config, err := core.LoadConfigFromFile(tmpFile.Name())
    require.NoError(t, err)
    
    // Verify memory configuration
    assert.Equal(t, "memory", config.Memory.Provider)
    assert.Equal(t, 10, config.Memory.MaxResults)
    assert.Equal(t, 384, config.Memory.Dimensions)
    assert.True(t, config.Memory.AutoEmbed)
    
    // Verify MCP configuration
    assert.True(t, config.MCP.EnableDiscovery)
    assert.Equal(t, 5*time.Second, config.MCP.DiscoveryTimeout)
}

func TestConfigurationValidation(t *testing.T) {
    tests := []struct {
        name        string
        config      core.AgentMemoryConfig
        expectError bool
        errorMsg    string
    }{
        {
            name: "valid memory config",
            config: core.AgentMemoryConfig{
                Provider:   "memory",
                MaxResults: 10,
                Dimensions: 384,
            },
            expectError: false,
        },
        {
            name: "invalid provider",
            config: core.AgentMemoryConfig{
                Provider:   "invalid",
                MaxResults: 10,
                Dimensions: 384,
            },
            expectError: true,
            errorMsg:    "invalid memory provider",
        },
        {
            name: "missing connection for pgvector",
            config: core.AgentMemoryConfig{
                Provider:   "pgvector",
                Connection: "",
                MaxResults: 10,
                Dimensions: 384,
            },
            expectError: true,
            errorMsg:    "connection string is required",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateMemoryConfig(tt.config)
            
            if tt.expectError {
                assert.Error(t, err)
                if tt.errorMsg != "" {
                    assert.Contains(t, err.Error(), tt.errorMsg)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## 📚 Best Practices

### 1. Configuration Organization

```go
// Good: Organized configuration structure
type AppConfig struct {
    Environment string                    `toml:"environment"`
    Memory      core.AgentMemoryConfig   `toml:"memory"`
    MCP         core.MCPConfig           `toml:"mcp"`
    Runner      core.EnhancedRunnerConfig `toml:"runner"`
    Logging     LoggingConfig            `toml:"logging"`
    Metrics     MetricsConfig            `toml:"metrics"`
}

type LoggingConfig struct {
    Level  string `toml:"level"`
    Format string `toml:"format"`
    Output string `toml:"output"`
}

type MetricsConfig struct {
    Enabled bool   `toml:"enabled"`
    Port    int    `toml:"port"`
    Path    string `toml:"path"`
}

// Bad: Flat, unorganized configuration
type BadConfig struct {
    MemoryProvider     string `toml:"memory_provider"`
    MemoryConnection   string `toml:"memory_connection"`
    MCPDiscovery       bool   `toml:"mcp_discovery"`
    MCPTimeout         string `toml:"mcp_timeout"`
    LogLevel           string `toml:"log_level"`
    MetricsEnabled     bool   `toml:"metrics_enabled"`
    // ... many more flat fields
}
```

### 2. Environment-Specific Configuration

```go
// Good: Environment-aware configuration
func loadEnvironmentConfig() (*AppConfig, error) {
    env := getEnvOrDefault("ENVIRONMENT", "development")
    
    configFile := fmt.Sprintf("config/%s.toml", env)
    if _, err := os.Stat(configFile); os.IsNotExist(err) {
        configFile = "config/default.toml"
    }
    
    config, err := loadConfigFromFile(configFile)
    if err != nil {
        return nil, err
    }
    
    // Override with environment variables
    overrideWithEnvVars(config)
    
    // Validate configuration
    if err := validateConfig(config); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }
    
    return config, nil
}
```

### 3. Configuration Security

```go
// Good: Secure configuration handling
func secureConfigExample() {
    config := &core.AgentMemoryConfig{
        Provider:   "pgvector",
        Connection: buildSecureConnectionString(),
        Embedding: core.EmbeddingConfig{
            Provider: "openai",
            APIKey:   getSecretFromVault("OPENAI_API_KEY"),
        },
    }
    
    // Don't log sensitive configuration
    logConfig := *config
    logConfig.Connection = maskConnectionString(config.Connection)
    logConfig.Embedding.APIKey = "***masked***"
    
    fmt.Printf("Loaded configuration: %+v\n", logConfig)
}

func buildSecureConnectionString() string {
    host := getEnvOrDefault("DB_HOST", "localhost")
    port := getEnvOrDefault("DB_PORT", "5432")
    user := getEnvOrDefault("DB_USER", "agentflow")
    password := getSecretFromVault("DB_PASSWORD")
    database := getEnvOrDefault("DB_NAME", "agentflow")
    
    return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require", 
        user, password, host, port, database)
}

func getSecretFromVault(key string) string {
    // Implementation would fetch from secure vault
    // For demo, fall back to environment variable
    return os.Getenv(key)
}

func maskConnectionString(connStr string) string {
    // Mask password in connection string for logging
    re := regexp.MustCompile(`://([^:]+):([^@]+)@`)
    return re.ReplaceAllString(connStr, "://$1:***@")
}
```

### 4. Configuration Documentation

```go
// Good: Well-documented configuration
type DocumentedConfig struct {
    // Memory configuration for agent storage and retrieval
    Memory MemoryConfig `toml:"memory" doc:"Configuration for agent memory systems"`
    
    // MCP configuration for tool integration
    MCP MCPConfig `toml:"mcp" doc:"Model Context Protocol integration settings"`
    
    // Runner configuration for agent orchestration
    Runner RunnerConfig `toml:"runner" doc:"Agent runner and orchestration settings"`
}

type MemoryConfig struct {
    // Provider specifies the memory backend (memory, pgvector, weaviate)
    Provider string `toml:"provider" doc:"Memory provider type" default:"memory"`
    
    // Connection string for database providers
    Connection string `toml:"connection" doc:"Database connection string" example:"postgres://user:pass@host:5432/db"`
    
    // Maximum number of results to return from searches
    MaxResults int `toml:"max_results" doc:"Maximum search results" default:"10" min:"1" max:"100"`
    
    // Vector dimensions for embeddings
    Dimensions int `toml:"dimensions" doc:"Embedding vector dimensions" default:"1536"`
}
```

## 🔗 Related APIs

- **[Agent API](agent.md)** - Building individual agents
- **[Orchestration API](orchestration.md)** - Multi-agent coordination
- **[Memory API](memory.md)** - Persistent storage and RAG
- **[MCP API](mcp.md)** - Tool integration

---

*This documentation covers the current Configuration API in AgenticGoKit. The framework is actively developed, so some interfaces may evolve.*