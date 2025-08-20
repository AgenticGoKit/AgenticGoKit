# Migration Guide: From Hardcoded to Configuration-Driven Agents

## Table of Contents

1. [Overview](#overview)
2. [Migration Benefits](#migration-benefits)
3. [Pre-Migration Assessment](#pre-migration-assessment)
4. [Step-by-Step Migration Process](#step-by-step-migration-process)
5. [Code Transformation Examples](#code-transformation-examples)
6. [Testing Your Migration](#testing-your-migration)
7. [Deployment Considerations](#deployment-considerations)
8. [Troubleshooting](#troubleshooting)
9. [Best Practices](#best-practices)
10. [Complete Migration Example](#complete-migration-example)

## Overview

This guide helps you migrate from hardcoded agent configurations to the flexible, configuration-driven approach provided by AgentFlow. The migration process is designed to be gradual and backward-compatible, allowing you to migrate agents one at a time without breaking existing functionality.

### What Changes

**Before (Hardcoded):**
```go
// Hardcoded agent creation
researcher := &ResearchAgent{
    Role:        "researcher",
    Model:       "gpt-4",
    Temperature: 0.3,
    MaxTokens:   1500,
    SystemPrompt: "You are a research specialist...",
}

writer := &WriterAgent{
    Role:        "writer", 
    Model:       "gpt-4",
    Temperature: 0.8,
    MaxTokens:   2000,
    SystemPrompt: "You are a skilled writer...",
}
```

**After (Configuration-Driven):**
```toml
# agentflow.toml
[agents.researcher]
role = "researcher"
description = "Research specialist"
system_prompt = "You are a research specialist..."
capabilities = ["web_search", "document_analysis"]
enabled = true
auto_llm = true

[agents.researcher.llm]
model = "gpt-4"
temperature = 0.3
max_tokens = 1500

[agents.writer]
role = "writer"
description = "Content writer"
system_prompt = "You are a skilled writer..."
capabilities = ["content_creation", "editing"]
enabled = true
auto_llm = true

[agents.writer.llm]
model = "gpt-4"
temperature = 0.8
max_tokens = 2000
```

```go
// Configuration-driven agent creation
factory := core.NewConfigurableAgentFactory(config)
researcher, err := factory.CreateAgent("researcher")
writer, err := factory.CreateAgent("writer")
```

## Migration Benefits

### Immediate Benefits
- **Flexibility**: Change agent behavior without code changes
- **Environment-Specific Configs**: Different settings for dev/staging/prod
- **Hot Reload**: Update configurations without restarting
- **Validation**: Comprehensive validation with helpful error messages
- **Standardization**: Consistent configuration format across all agents

### Long-Term Benefits
- **Maintainability**: Easier to manage complex multi-agent systems
- **Scalability**: Add new agents without code changes
- **Collaboration**: Non-developers can modify agent behavior
- **Experimentation**: A/B test different agent configurations
- **Compliance**: Centralized configuration management for auditing

## Pre-Migration Assessment

### 1. Inventory Your Hardcoded Agents

Create a list of all hardcoded agents in your system:

```bash
# Find hardcoded agent creation patterns
grep -r "Temperature.*=" --include="*.go" .
grep -r "MaxTokens.*=" --include="*.go" .
grep -r "SystemPrompt.*=" --include="*.go" .
grep -r "Model.*=" --include="*.go" .
```

### 2. Document Current Configuration

For each agent, document:
- Role/Purpose
- LLM Settings (model, temperature, max_tokens)
- System Prompt
- Capabilities
- Dependencies
- Environment-specific variations

### 3. Identify Migration Complexity

**Simple Migration** (Low Risk):
- Agents with static configuration
- No complex initialization logic
- Clear separation of concerns

**Complex Migration** (Higher Risk):
- Agents with dynamic configuration
- Complex initialization dependencies
- Shared state between agents
- Custom LLM providers

### 4. Plan Migration Order

Recommended order:
1. **Leaf Agents**: Agents with no dependencies
2. **Independent Agents**: Agents that don't interact with others
3. **Core Agents**: Central agents in your workflow
4. **Complex Agents**: Agents with intricate logic

## Step-by-Step Migration Process

### Step 1: Set Up Configuration Infrastructure

#### 1.1 Create Configuration File

Create `agentflow.toml` in your project root:

```toml
[agent_flow]
name = "my-agent-system"
version = "1.0.0"
description = "Migrating to configuration-driven agents"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 2000
```

#### 1.2 Update Dependencies

Ensure you have the latest AgentFlow version:

```go
// go.mod
require github.com/kunalkushwaha/agenticgokit v1.2.0
```

#### 1.3 Initialize Configuration Loading

```go
// main.go or initialization code
import "github.com/kunalkushwaha/agenticgokit/core"

func main() {
    // Load configuration
    config, err := core.LoadConfig("agentflow.toml")
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }
    
    // Validate configuration
    validator := core.NewDefaultConfigValidator()
    if errors := validator.ValidateConfig(config); len(errors) > 0 {
        for _, err := range errors {
            log.Printf("Config validation: %s", err.Message)
        }
    }
    
    // Create agent factory
    factory := core.NewConfigurableAgentFactory(config)
    
    // Continue with existing code...
}
```

### Step 2: Migrate Your First Agent

#### 2.1 Choose a Simple Agent

Start with a simple, independent agent:

```go
// Before: Hardcoded researcher agent
type ResearchAgent struct {
    Role         string
    Model        string
    Temperature  float32
    MaxTokens    int
    SystemPrompt string
}

func NewResearchAgent() *ResearchAgent {
    return &ResearchAgent{
        Role:         "researcher",
        Model:        "gpt-4",
        Temperature:  0.3,
        MaxTokens:    1500,
        SystemPrompt: "You are a research specialist focused on gathering accurate information.",
    }
}
```

#### 2.2 Add Agent to Configuration

```toml
[agents.researcher]
role = "researcher"
description = "Research and information gathering specialist"
system_prompt = "You are a research specialist focused on gathering accurate information."
capabilities = ["web_search", "document_analysis", "fact_checking"]
enabled = true

[agents.researcher.llm]
model = "gpt-4"
temperature = 0.3
max_tokens = 1500
```

#### 2.3 Update Agent Implementation

```go
// After: Configuration-aware researcher agent
type ResearchAgent struct {
    config *core.ResolvedAgentConfig
    llm    LLMProvider
}

func NewResearchAgentFromConfig(config *core.ResolvedAgentConfig) (*ResearchAgent, error) {
    // Initialize LLM provider with resolved configuration
    llm, err := initializeLLMProvider(config.LLM)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize LLM: %w", err)
    }
    
    return &ResearchAgent{
        config: config,
        llm:    llm,
    }, nil
}

func (r *ResearchAgent) GetRole() string {
    return r.config.Role
}

func (r *ResearchAgent) Run(ctx context.Context, input string) (string, error) {
    // Use configuration-driven system prompt
    prompt := r.config.SystemPrompt + "\n\nUser: " + input
    
    // Use configured LLM settings
    response, err := r.llm.Generate(ctx, prompt)
    if err != nil {
        return "", fmt.Errorf("LLM generation failed: %w", err)
    }
    
    return response, nil
}
```

#### 2.4 Update Agent Creation Code

```go
// Before: Direct instantiation
researcher := NewResearchAgent()

// After: Factory-based creation
researcher, err := factory.CreateAgent("researcher")
if err != nil {
    return fmt.Errorf("failed to create researcher: %w", err)
}
```

### Step 3: Implement Backward Compatibility

During migration, support both hardcoded and configured agents:

```go
type AgentManager struct {
    factory       *core.ConfigurableAgentFactory
    legacyAgents  map[string]Agent
}

func (am *AgentManager) GetAgent(name string) (Agent, error) {
    // Try configuration-driven agent first
    if agent, err := am.factory.CreateAgent(name); err == nil {
        return agent, nil
    }
    
    // Fall back to legacy hardcoded agent
    if agent, exists := am.legacyAgents[name]; exists {
        return agent, nil
    }
    
    return nil, fmt.Errorf("agent %s not found", name)
}
```

### Step 4: Migrate Remaining Agents

Repeat the process for each agent:

1. Add agent configuration to `agentflow.toml`
2. Update agent implementation to use configuration
3. Update agent creation code
4. Test the migrated agent
5. Remove hardcoded fallback when confident

### Step 5: Clean Up Legacy Code

Once all agents are migrated:

1. Remove hardcoded agent creation functions
2. Remove legacy agent fallback logic
3. Update tests to use configuration
4. Update documentation

## Code Transformation Examples

### Example 1: Simple Agent Migration

#### Before (Hardcoded)
```go
type SimpleAgent struct {
    name        string
    model       string
    temperature float32
    prompt      string
}

func NewSimpleAgent() *SimpleAgent {
    return &SimpleAgent{
        name:        "simple",
        model:       "gpt-3.5-turbo",
        temperature: 0.7,
        prompt:      "You are a helpful assistant.",
    }
}

func (s *SimpleAgent) Process(input string) (string, error) {
    // Hardcoded LLM call
    client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
    
    resp, err := client.CreateChatCompletion(
        context.Background(),
        openai.ChatCompletionRequest{
            Model:       s.model,
            Temperature: s.temperature,
            Messages: []openai.ChatCompletionMessage{
                {Role: "system", Content: s.prompt},
                {Role: "user", Content: input},
            },
        },
    )
    
    if err != nil {
        return "", err
    }
    
    return resp.Choices[0].Message.Content, nil
}
```

#### After (Configuration-Driven)
```toml
[agents.simple]
role = "assistant"
description = "Simple helpful assistant"
system_prompt = "You are a helpful assistant."
capabilities = ["general_assistance"]
enabled = true

[agents.simple.llm]
model = "gpt-3.5-turbo"
temperature = 0.7
max_tokens = 1000
```

```go
type SimpleAgent struct {
    config *core.ResolvedAgentConfig
    llm    core.LLMProvider
}

func NewSimpleAgentFromConfig(config *core.ResolvedAgentConfig) (*SimpleAgent, error) {
    llm, err := core.NewLLMProvider(config.LLM)
    if err != nil {
        return nil, err
    }
    
    return &SimpleAgent{
        config: config,
        llm:    llm,
    }, nil
}

func (s *SimpleAgent) GetRole() string {
    return s.config.Role
}

func (s *SimpleAgent) Process(ctx context.Context, input string) (string, error) {
    // Use configuration-driven settings
    return s.llm.Generate(ctx, s.config.SystemPrompt, input)
}
```

### Example 2: Complex Agent with Dependencies

#### Before (Hardcoded)
```go
type ComplexAgent struct {
    name         string
    primaryLLM   LLMProvider
    fallbackLLM  LLMProvider
    memory       MemoryProvider
    tools        []Tool
    retryPolicy  RetryPolicy
}

func NewComplexAgent(memoryProvider MemoryProvider) *ComplexAgent {
    return &ComplexAgent{
        name: "complex",
        primaryLLM: &OpenAIProvider{
            Model:       "gpt-4",
            Temperature: 0.7,
            MaxTokens:   2000,
        },
        fallbackLLM: &OpenAIProvider{
            Model:       "gpt-3.5-turbo",
            Temperature: 0.7,
            MaxTokens:   1000,
        },
        memory: memoryProvider,
        tools: []Tool{
            &WebSearchTool{},
            &CalculatorTool{},
        },
        retryPolicy: RetryPolicy{
            MaxRetries: 3,
            BaseDelay:  time.Second,
        },
    }
}
```

#### After (Configuration-Driven)
```toml
[agents.complex]
role = "complex_processor"
description = "Complex agent with multiple capabilities"
system_prompt = "You are a sophisticated AI agent with access to multiple tools."
capabilities = ["web_search", "calculations", "memory_access"]
enabled = true

[agents.complex.llm]
model = "gpt-4"
temperature = 0.7
max_tokens = 2000

[agents.complex.fallback_llm]
model = "gpt-3.5-turbo"
temperature = 0.7
max_tokens = 1000

[agents.complex.retry_policy]
max_retries = 3
base_delay_ms = 1000
max_delay_ms = 30000
backoff_multiplier = 2.0

[agents.complex.tools]
enabled_tools = ["web_search", "calculator"]

[agents.complex.memory]
enabled = true
max_entries = 1000
```

```go
type ComplexAgent struct {
    config      *core.ResolvedAgentConfig
    primaryLLM  core.LLMProvider
    fallbackLLM core.LLMProvider
    memory      core.MemoryProvider
    tools       map[string]core.Tool
    retryPolicy *core.RetryPolicy
}

func NewComplexAgentFromConfig(
    config *core.ResolvedAgentConfig,
    memoryProvider core.MemoryProvider,
    toolRegistry core.ToolRegistry,
) (*ComplexAgent, error) {
    // Initialize primary LLM
    primaryLLM, err := core.NewLLMProvider(config.LLM)
    if err != nil {
        return nil, fmt.Errorf("failed to create primary LLM: %w", err)
    }
    
    // Initialize fallback LLM if configured
    var fallbackLLM core.LLMProvider
    if config.FallbackLLM != nil {
        fallbackLLM, err = core.NewLLMProvider(*config.FallbackLLM)
        if err != nil {
            return nil, fmt.Errorf("failed to create fallback LLM: %w", err)
        }
    }
    
    // Initialize tools based on configuration
    tools := make(map[string]core.Tool)
    for _, toolName := range config.Tools.EnabledTools {
        tool, err := toolRegistry.GetTool(toolName)
        if err != nil {
            return nil, fmt.Errorf("failed to get tool %s: %w", toolName, err)
        }
        tools[toolName] = tool
    }
    
    return &ComplexAgent{
        config:      config,
        primaryLLM:  primaryLLM,
        fallbackLLM: fallbackLLM,
        memory:      memoryProvider,
        tools:       tools,
        retryPolicy: config.RetryPolicy,
    }, nil
}
```

### Example 3: Agent with Environment-Specific Configuration

#### Configuration for Different Environments

**Development (`agentflow.dev.toml`):**
```toml
[agents.researcher]
role = "researcher"
description = "Development researcher with limited capabilities"
system_prompt = "You are a researcher. Keep responses brief for development."
capabilities = ["basic_search"]
enabled = true

[agents.researcher.llm]
model = "gpt-3.5-turbo"  # Cheaper for development
temperature = 0.5
max_tokens = 500         # Smaller responses
```

**Production (`agentflow.prod.toml`):**
```toml
[agents.researcher]
role = "researcher"
description = "Production researcher with full capabilities"
system_prompt = "You are a comprehensive research specialist with access to multiple data sources."
capabilities = ["web_search", "document_analysis", "fact_checking", "citation_generation"]
enabled = true

[agents.researcher.llm]
model = "gpt-4"          # Better model for production
temperature = 0.3
max_tokens = 3000        # Longer, detailed responses

[agents.researcher.retry_policy]
max_retries = 5
base_delay_ms = 1000
max_delay_ms = 60000
```

**Environment-Aware Loading:**
```go
func loadEnvironmentConfig() (*core.Config, error) {
    env := os.Getenv("ENVIRONMENT")
    if env == "" {
        env = "development"
    }
    
    configFile := fmt.Sprintf("agentflow.%s.toml", env)
    
    // Fall back to default if environment-specific config doesn't exist
    if _, err := os.Stat(configFile); os.IsNotExist(err) {
        configFile = "agentflow.toml"
    }
    
    return core.LoadConfig(configFile)
}
```

## Testing Your Migration

### 1. Unit Tests for Configuration Loading

```go
func TestConfigurationLoading(t *testing.T) {
    // Test configuration loading
    config, err := core.LoadConfig("testdata/agentflow.toml")
    require.NoError(t, err)
    
    // Validate configuration structure
    assert.Equal(t, "test-system", config.AgentFlow.Name)
    assert.Contains(t, config.Agents, "researcher")
    
    // Test agent configuration resolution
    resolver := core.NewConfigResolver(config)
    agentConfig, err := resolver.ResolveAgentConfig("researcher")
    require.NoError(t, err)
    
    assert.Equal(t, "researcher", agentConfig.Role)
    assert.Equal(t, "gpt-4", agentConfig.LLM.Model)
}
```

### 2. Integration Tests for Agent Creation

```go
func TestAgentCreationFromConfig(t *testing.T) {
    config, err := core.LoadConfig("testdata/agentflow.toml")
    require.NoError(t, err)
    
    factory := core.NewConfigurableAgentFactory(config)
    
    // Test agent creation
    agent, err := factory.CreateAgent("researcher")
    require.NoError(t, err)
    assert.NotNil(t, agent)
    
    // Test agent functionality
    result, err := agent.Process(context.Background(), "test input")
    require.NoError(t, err)
    assert.NotEmpty(t, result)
}
```

### 3. Backward Compatibility Tests

```go
func TestBackwardCompatibility(t *testing.T) {
    // Test that legacy agents still work
    legacyAgent := NewLegacyResearchAgent()
    result, err := legacyAgent.Process("test input")
    require.NoError(t, err)
    
    // Test that new configuration-driven agents work
    config, err := core.LoadConfig("testdata/agentflow.toml")
    require.NoError(t, err)
    
    factory := core.NewConfigurableAgentFactory(config)
    configAgent, err := factory.CreateAgent("researcher")
    require.NoError(t, err)
    
    configResult, err := configAgent.Process(context.Background(), "test input")
    require.NoError(t, err)
    
    // Both should produce valid results
    assert.NotEmpty(t, result)
    assert.NotEmpty(t, configResult)
}
```

### 4. Configuration Validation Tests

```go
func TestConfigurationValidation(t *testing.T) {
    tests := []struct {
        name        string
        configFile  string
        expectError bool
        errorCount  int
    }{
        {
            name:        "valid configuration",
            configFile:  "testdata/valid-config.toml",
            expectError: false,
            errorCount:  0,
        },
        {
            name:        "missing required fields",
            configFile:  "testdata/invalid-config.toml",
            expectError: true,
            errorCount:  3,
        },
    }
    
    validator := core.NewDefaultConfigValidator()
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            config, err := core.LoadConfig(tt.configFile)
            require.NoError(t, err)
            
            errors := validator.ValidateConfig(config)
            
            if tt.expectError {
                assert.Len(t, errors, tt.errorCount)
            } else {
                assert.Empty(t, errors)
            }
        })
    }
}
```

## Deployment Considerations

### 1. Configuration Management

**Version Control:**
```bash
# Track configuration files in version control
git add agentflow.toml
git add agentflow.dev.toml
git add agentflow.prod.toml

# Use .gitignore for sensitive configurations
echo "agentflow.local.toml" >> .gitignore
```

**Configuration Deployment:**
```bash
# Deploy configuration with application
docker build -t my-agent-app .
docker run -v /path/to/config:/app/config my-agent-app

# Or use environment-specific configurations
docker run -e ENVIRONMENT=production my-agent-app
```

### 2. Environment Variables

Set up environment variables for sensitive data:

```bash
# Production environment variables
export OPENAI_API_KEY="your-production-key"
export AGENTFLOW_LLM_MODEL="gpt-4"
export AGENTFLOW_MEMORY_CONNECTION_STRING="postgresql://prod-db:5432/agentflow"

# Development environment variables
export OPENAI_API_KEY="your-dev-key"
export AGENTFLOW_LLM_MODEL="gpt-3.5-turbo"
export AGENTFLOW_MEMORY_CONNECTION_STRING="postgresql://dev-db:5432/agentflow"
```

### 3. Configuration Validation in CI/CD

```yaml
# .github/workflows/validate-config.yml
name: Validate Configuration

on:
  pull_request:
    paths:
      - '*.toml'
      - 'config/*.toml'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.21
          
      - name: Install AgentCLI
        run: go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest
        
      - name: Validate configurations
        run: |
          agentcli validate agentflow.toml
          agentcli validate agentflow.dev.toml
          agentcli validate agentflow.prod.toml
```

### 4. Monitoring and Alerting

Monitor configuration changes and agent behavior:

```go
// Add configuration change logging
reloader := core.NewConfigReloader("agentflow.toml")
reloader.OnConfigReload(func(newConfig *core.Config) error {
    log.Printf("Configuration reloaded: %s v%s", 
        newConfig.AgentFlow.Name, 
        newConfig.AgentFlow.Version)
    
    // Send metrics to monitoring system
    metrics.Counter("config_reload_total").Inc()
    
    return nil
})
```

## Troubleshooting

### Common Migration Issues

#### 1. Configuration Not Loading

**Problem:** Configuration file not found or invalid syntax

**Solution:**
```bash
# Check file exists
ls -la agentflow.toml

# Validate TOML syntax
agentcli validate agentflow.toml

# Check file permissions
chmod 644 agentflow.toml
```

#### 2. Agent Creation Failures

**Problem:** Agent creation fails with configuration errors

**Solution:**
```bash
# Validate specific agent configuration
agentcli validate --detailed agentflow.toml

# Check for missing required fields
grep -A 10 "agents.problematic_agent" agentflow.toml

# Test with minimal configuration
```

#### 3. Environment Variable Issues

**Problem:** Environment variables not overriding configuration

**Solution:**
```bash
# Check environment variable format
env | grep AGENTFLOW

# Test with explicit values
AGENTFLOW_LLM_MODEL=gpt-4 ./my-app

# Debug environment variable loading
export AGENTFLOW_DEBUG=true
```

#### 4. Performance Issues After Migration

**Problem:** Slower performance with configuration-driven agents

**Solution:**
```go
// Cache resolved configurations
type CachedAgentFactory struct {
    factory *core.ConfigurableAgentFactory
    cache   map[string]*core.ResolvedAgentConfig
    mutex   sync.RWMutex
}

func (c *CachedAgentFactory) CreateAgent(name string) (core.Agent, error) {
    c.mutex.RLock()
    config, exists := c.cache[name]
    c.mutex.RUnlock()
    
    if !exists {
        c.mutex.Lock()
        config, err := c.factory.ResolveAgentConfig(name)
        if err != nil {
            c.mutex.Unlock()
            return nil, err
        }
        c.cache[name] = config
        c.mutex.Unlock()
    }
    
    return c.factory.CreateAgentFromResolvedConfig(config)
}
```

### Debug Mode

Enable debug logging during migration:

```toml
[debug]
enabled = true
log_level = "debug"
log_config_loading = true
log_agent_creation = true
log_validation = true
```

```bash
# Or via environment variable
export AGENTFLOW_DEBUG=true
export AGENTFLOW_LOG_LEVEL=debug
```

## Best Practices

### 1. Gradual Migration

- **Start Small**: Migrate one agent at a time
- **Test Thoroughly**: Validate each migration step
- **Keep Fallbacks**: Maintain backward compatibility during transition
- **Monitor Performance**: Watch for performance impacts

### 2. Configuration Management

- **Version Control**: Track all configuration changes
- **Environment Separation**: Use different configs for different environments
- **Validation**: Always validate configurations before deployment
- **Documentation**: Document configuration changes and rationale

### 3. Testing Strategy

- **Unit Tests**: Test configuration loading and validation
- **Integration Tests**: Test agent creation and functionality
- **Backward Compatibility**: Ensure legacy code still works
- **Performance Tests**: Verify no performance degradation

### 4. Deployment Strategy

- **Blue-Green Deployment**: Deploy new configuration alongside old
- **Canary Releases**: Gradually roll out configuration changes
- **Rollback Plan**: Have a plan to revert to previous configuration
- **Monitoring**: Monitor system behavior after configuration changes

## Complete Migration Example

Here's a complete example showing the migration of a multi-agent content creation system:

### Before Migration

```go
// content_system.go - Before migration
package main

import (
    "context"
    "fmt"
    "log"
    "os"
)

type ContentSystem struct {
    researcher *ResearchAgent
    writer     *WriterAgent
    reviewer   *ReviewerAgent
}

type ResearchAgent struct {
    model       string
    temperature float32
    maxTokens   int
    prompt      string
}

func NewResearchAgent() *ResearchAgent {
    return &ResearchAgent{
        model:       "gpt-4",
        temperature: 0.3,
        maxTokens:   2000,
        prompt:      "You are a research specialist. Gather accurate information on the given topic.",
    }
}

func (r *ResearchAgent) Research(topic string) (string, error) {
    // Hardcoded OpenAI call
    // ... implementation
    return "research results", nil
}

type WriterAgent struct {
    model       string
    temperature float32
    maxTokens   int
    prompt      string
}

func NewWriterAgent() *WriterAgent {
    return &WriterAgent{
        model:       "gpt-4",
        temperature: 0.8,
        maxTokens:   3000,
        prompt:      "You are a skilled content writer. Create engaging content based on research.",
    }
}

func (w *WriterAgent) Write(research string) (string, error) {
    // Hardcoded OpenAI call
    // ... implementation
    return "written content", nil
}

type ReviewerAgent struct {
    model       string
    temperature float32
    maxTokens   int
    prompt      string
}

func NewReviewerAgent() *ReviewerAgent {
    return &ReviewerAgent{
        model:       "gpt-4",
        temperature: 0.2,
        maxTokens:   1500,
        prompt:      "You are a content reviewer. Review and improve the given content.",
    }
}

func (r *ReviewerAgent) Review(content string) (string, error) {
    // Hardcoded OpenAI call
    // ... implementation
    return "reviewed content", nil
}

func NewContentSystem() *ContentSystem {
    return &ContentSystem{
        researcher: NewResearchAgent(),
        writer:     NewWriterAgent(),
        reviewer:   NewReviewerAgent(),
    }
}

func (cs *ContentSystem) CreateContent(topic string) (string, error) {
    research, err := cs.researcher.Research(topic)
    if err != nil {
        return "", err
    }
    
    content, err := cs.writer.Write(research)
    if err != nil {
        return "", err
    }
    
    finalContent, err := cs.reviewer.Review(content)
    if err != nil {
        return "", err
    }
    
    return finalContent, nil
}

func main() {
    system := NewContentSystem()
    content, err := system.CreateContent("AI in healthcare")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(content)
}
```

### After Migration

#### Configuration File (`agentflow.toml`)

```toml
[agent_flow]
name = "content-creation-system"
version = "2.0.0"
description = "Configuration-driven content creation system"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 2000
timeout_seconds = 30

[agents.researcher]
role = "researcher"
description = "Research and information gathering specialist"
system_prompt = "You are a research specialist. Gather accurate, up-to-date information on the given topic. Focus on credible sources and provide comprehensive coverage."
capabilities = ["web_search", "fact_checking", "source_verification"]
enabled = true

[agents.researcher.llm]
temperature = 0.3
max_tokens = 2000

[agents.researcher.retry_policy]
max_retries = 3
base_delay_ms = 1000
max_delay_ms = 10000

[agents.writer]
role = "writer"
description = "Content creation and writing specialist"
system_prompt = "You are a skilled content writer. Create engaging, well-structured content based on research. Focus on clarity, readability, and audience engagement."
capabilities = ["content_creation", "seo_optimization", "copywriting"]
enabled = true

[agents.writer.llm]
temperature = 0.8
max_tokens = 3000

[agents.reviewer]
role = "reviewer"
description = "Content review and quality assurance specialist"
system_prompt = "You are a content reviewer. Review and improve the given content for accuracy, clarity, grammar, and overall quality. Provide constructive feedback."
capabilities = ["quality_assurance", "proofreading", "fact_checking"]
enabled = true

[agents.reviewer.llm]
temperature = 0.2
max_tokens = 1500

[orchestration]
mode = "sequential"
agents = ["researcher", "writer", "reviewer"]

[orchestration.sequential]
stop_on_error = false
pass_context = true
timeout_per_agent = 300

[memory]
enabled = true
provider = "local"
max_entries = 1000
ttl_seconds = 3600

[retry_policy]
max_retries = 3
base_delay_ms = 1000
max_delay_ms = 30000
backoff_multiplier = 2.0

[monitoring]
enabled = true
log_level = "info"
```

#### Migrated Code (`content_system.go`)

```go
// content_system.go - After migration
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

type ContentSystem struct {
    config   *core.Config
    factory  *core.ConfigurableAgentFactory
    manager  *core.AgentManager
}

func NewContentSystem(configPath string) (*ContentSystem, error) {
    // Load configuration
    config, err := core.LoadConfig(configPath)
    if err != nil {
        return nil, fmt.Errorf("failed to load configuration: %w", err)
    }
    
    // Validate configuration
    validator := core.NewDefaultConfigValidator()
    if errors := validator.ValidateConfig(config); len(errors) > 0 {
        for _, err := range errors {
            log.Printf("Config validation warning: %s", err.Message)
        }
    }
    
    // Create agent factory and manager
    factory := core.NewConfigurableAgentFactory(config)
    manager := core.NewAgentManager(config)
    
    // Initialize agents
    if err := manager.InitializeAgents(); err != nil {
        return nil, fmt.Errorf("failed to initialize agents: %w", err)
    }
    
    return &ContentSystem{
        config:  config,
        factory: factory,
        manager: manager,
    }, nil
}

func (cs *ContentSystem) CreateContent(ctx context.Context, topic string) (string, error) {
    // Get agents from manager
    researcher, err := cs.manager.GetAgent("researcher")
    if err != nil {
        return "", fmt.Errorf("failed to get researcher: %w", err)
    }
    
    writer, err := cs.manager.GetAgent("writer")
    if err != nil {
        return "", fmt.Errorf("failed to get writer: %w", err)
    }
    
    reviewer, err := cs.manager.GetAgent("reviewer")
    if err != nil {
        return "", fmt.Errorf("failed to get reviewer: %w", err)
    }
    
    // Execute content creation pipeline
    research, err := researcher.Run(ctx, fmt.Sprintf("Research topic: %s", topic))
    if err != nil {
        return "", fmt.Errorf("research failed: %w", err)
    }
    
    content, err := writer.Run(ctx, fmt.Sprintf("Write content based on research: %s", research))
    if err != nil {
        return "", fmt.Errorf("writing failed: %w", err)
    }
    
    finalContent, err := reviewer.Run(ctx, fmt.Sprintf("Review and improve content: %s", content))
    if err != nil {
        return "", fmt.Errorf("review failed: %w", err)
    }
    
    return finalContent, nil
}

// Alternative: Use orchestrator for automatic pipeline execution
func (cs *ContentSystem) CreateContentWithOrchestrator(ctx context.Context, topic string) (string, error) {
    orchestrator, err := core.NewOrchestrator(cs.config)
    if err != nil {
        return "", fmt.Errorf("failed to create orchestrator: %w", err)
    }
    
    result, err := orchestrator.Execute(ctx, topic)
    if err != nil {
        return "", fmt.Errorf("orchestration failed: %w", err)
    }
    
    return result.FinalOutput, nil
}

func (cs *ContentSystem) EnableHotReload() error {
    reloader := core.NewConfigReloader("agentflow.toml")
    
    reloader.OnConfigReload(func(newConfig *core.Config) error {
        log.Printf("Configuration reloaded: %s v%s", 
            newConfig.AgentFlow.Name, 
            newConfig.AgentFlow.Version)
        
        // Update agent manager with new configuration
        return cs.manager.UpdateConfiguration(newConfig)
    })
    
    go func() {
        if err := reloader.StartWatching(context.Background()); err != nil {
            log.Printf("Config watcher error: %v", err)
        }
    }()
    
    return nil
}

func main() {
    // Create content system with configuration
    system, err := NewContentSystem("agentflow.toml")
    if err != nil {
        log.Fatalf("Failed to create content system: %v", err)
    }
    
    // Enable hot reload for development
    if err := system.EnableHotReload(); err != nil {
        log.Printf("Failed to enable hot reload: %v", err)
    }
    
    // Create content
    ctx := context.Background()
    content, err := system.CreateContent(ctx, "AI in healthcare")
    if err != nil {
        log.Fatalf("Failed to create content: %v", err)
    }
    
    fmt.Println("Generated Content:")
    fmt.Println(content)
    
    // Alternative: Use orchestrator
    orchestratedContent, err := system.CreateContentWithOrchestrator(ctx, "AI in healthcare")
    if err != nil {
        log.Fatalf("Failed to create orchestrated content: %v", err)
    }
    
    fmt.Println("\nOrchestrated Content:")
    fmt.Println(orchestratedContent)
}
```

#### Test File (`content_system_test.go`)

```go
package main

import (
    "context"
    "os"
    "path/filepath"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestContentSystemMigration(t *testing.T) {
    // Create temporary config file
    tempDir, err := os.MkdirTemp("", "content-system-test-*")
    require.NoError(t, err)
    defer os.RemoveAll(tempDir)
    
    configPath := filepath.Join(tempDir, "agentflow.toml")
    
    // Write test configuration
    configContent := `[agent_flow]
name = "test-content-system"
version = "1.0.0"

[llm]
provider = "openai"
model = "gpt-3.5-turbo"
temperature = 0.7

[agents.researcher]
role = "researcher"
description = "Test researcher"
system_prompt = "You are a test researcher."
capabilities = ["research"]
enabled = true

[agents.writer]
role = "writer"
description = "Test writer"
system_prompt = "You are a test writer."
capabilities = ["writing"]
enabled = true

[agents.reviewer]
role = "reviewer"
description = "Test reviewer"
system_prompt = "You are a test reviewer."
capabilities = ["reviewing"]
enabled = true

[orchestration]
mode = "sequential"
agents = ["researcher", "writer", "reviewer"]`
    
    err = os.WriteFile(configPath, []byte(configContent), 0644)
    require.NoError(t, err)
    
    // Test system creation
    system, err := NewContentSystem(configPath)
    require.NoError(t, err)
    assert.NotNil(t, system)
    
    // Test configuration loading
    assert.Equal(t, "test-content-system", system.config.AgentFlow.Name)
    assert.Len(t, system.config.Agents, 3)
    
    // Test agent creation
    researcher, err := system.factory.CreateAgent("researcher")
    require.NoError(t, err)
    assert.Equal(t, "researcher", researcher.GetRole())
    
    writer, err := system.factory.CreateAgent("writer")
    require.NoError(t, err)
    assert.Equal(t, "writer", writer.GetRole())
    
    reviewer, err := system.factory.CreateAgent("reviewer")
    require.NoError(t, err)
    assert.Equal(t, "reviewer", reviewer.GetRole())
}

func TestBackwardCompatibility(t *testing.T) {
    // Test that we can still create legacy agents if needed
    // This ensures the migration doesn't break existing functionality
    
    // Create a minimal configuration
    tempDir, err := os.MkdirTemp("", "backward-compat-test-*")
    require.NoError(t, err)
    defer os.RemoveAll(tempDir)
    
    configPath := filepath.Join(tempDir, "minimal.toml")
    minimalConfig := `[agent_flow]
name = "minimal-system"
version = "1.0.0"`
    
    err = os.WriteFile(configPath, []byte(minimalConfig), 0644)
    require.NoError(t, err)
    
    // Should be able to create system with minimal config
    system, err := NewContentSystem(configPath)
    require.NoError(t, err)
    assert.NotNil(t, system)
}
```

This complete migration example demonstrates:

1. **Full transformation** from hardcoded to configuration-driven
2. **Backward compatibility** during the transition
3. **Enhanced functionality** with orchestration and hot reload
4. **Comprehensive testing** to ensure migration success
5. **Production-ready** configuration management

The migrated system is more flexible, maintainable, and scalable than the original hardcoded version, while maintaining all existing functionality.