# Migration Guide: From Hardcoded to Configuration-Driven Agents

This guide helps you migrate existing AgenticGoKit projects from hardcoded agent implementations to the new configuration-driven approach.

## Overview

The configuration-driven system offers significant advantages over hardcoded implementations:

- **Flexibility**: Change agent behavior without code changes
- **Environment Support**: Different configurations for dev/staging/production
- **Hot-Reload**: Update configurations without restarting
- **Validation**: Comprehensive validation with helpful error messages
- **Templates**: Pre-built configurations for common patterns

## Migration Process

### Phase 1: Assessment and Planning

#### 1.1 Assess Current Implementation

First, analyze your existing agent implementations:

```bash
# Generate configuration from existing code
agentcli config generate --output current-config.toml

# Review generated configuration
agentcli validate current-config.toml --level complete
```

#### 1.2 Identify Hardcoded Values

Look for hardcoded values in your agent implementations:

```go
// Common hardcoded patterns to identify:

// System prompts
systemPrompt := "You are a research specialist..."

// LLM parameters
temperature := 0.3
maxTokens := 2000

// Timeouts and retries
timeout := 30 * time.Second
maxRetries := 3

// Capabilities
capabilities := []string{"research", "analysis"}

// Role definitions
role := "researcher"
```

#### 1.3 Plan Migration Strategy

Choose your migration approach:

1. **Gradual Migration**: Migrate agents one by one
2. **Complete Migration**: Migrate entire system at once
3. **Hybrid Approach**: Keep some agents hardcoded, migrate others

### Phase 2: Configuration Generation

#### 2.1 Generate Initial Configuration

```bash
# Generate configuration from existing agents
agentcli config generate

# Review and customize generated configuration
agentcli validate --level strict --verbose
```

#### 2.2 Enhance Generated Configuration

The generated configuration is a starting point. Enhance it with:

```toml
# Add comprehensive agent descriptions
[agents.researcher]
role = "research_specialist"
description = "Conducts comprehensive research with fact-checking and source validation"
system_prompt = """
You are an expert research specialist with the following responsibilities:
- Gather information from authoritative sources
- Verify facts and check source credibility
- Provide comprehensive analysis with citations
- Maintain objectivity and accuracy in all research
"""

# Add performance optimizations
[agents.researcher.llm]
temperature = 0.2  # Lower for factual accuracy
max_tokens = 2500  # Sufficient for detailed research

# Add reliability settings
[agents.researcher.retry_policy]
max_retries = 3
base_delay_ms = 1000
max_delay_ms = 5000
backoff_factor = 2.0

# Add metadata for organization
[agents.researcher.metadata]
specialization = "research"
priority = "high"
cost_tier = "standard"
```

#### 2.3 Validate Enhanced Configuration

```bash
# Comprehensive validation
agentcli validate --level complete

# Check for optimization opportunities
agentcli config optimize --recommend-only
```

### Phase 3: Code Migration

#### 3.1 Update Agent Constructors

**Before (Hardcoded):**
```go
type ResearchAgent struct {
    llm core.ModelProvider
}

func NewResearchAgent(llmProvider core.ModelProvider) *ResearchAgent {
    return &ResearchAgent{
        llm: llmProvider,
    }
}
```

**After (Configuration-Driven):**
```go
type ResearchAgent struct {
    config core.ResolvedAgentConfig
    llm    core.ModelProvider
}

func NewResearchAgent(config core.ResolvedAgentConfig, llmProvider core.ModelProvider) *ResearchAgent {
    return &ResearchAgent{
        config: config,
        llm:    llmProvider,
    }
}
```

#### 3.2 Update Agent Run Methods

**Before (Hardcoded):**
```go
func (a *ResearchAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Hardcoded system prompt
    systemPrompt := "You are a research specialist focused on gathering accurate information."
    
    // Hardcoded LLM parameters
    prompt := core.Prompt{
        System: systemPrompt,
        User:   fmt.Sprintf("Research: %v", event.GetData()["message"]),
    }
    
    // Hardcoded timeout
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    response, err := a.llm.Call(ctx, prompt)
    if err != nil {
        return core.AgentResult{}, err
    }
    
    // Return result
    outputState := core.NewState()
    outputState.Set("research_response", response.Content)
    
    return core.AgentResult{
        OutputState: outputState,
    }, nil
}
```

**After (Configuration-Driven):**
```go
func (a *ResearchAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Use configuration for system prompt
    systemPrompt := a.config.SystemPrompt
    
    // Build prompt with configured parameters
    prompt := core.Prompt{
        System: systemPrompt,
        User:   fmt.Sprintf("Research: %v", event.GetData()["message"]),
    }
    
    // Use configured timeout
    if a.config.Timeout > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, a.config.Timeout)
        defer cancel()
    }
    
    // Call LLM with configured parameters
    response, err := a.llm.Call(ctx, prompt)
    if err != nil {
        // Use configured retry policy if available
        if a.config.RetryPolicy != nil {
            return a.retryWithPolicy(ctx, prompt)
        }
        return core.AgentResult{}, err
    }
    
    // Return result
    outputState := core.NewState()
    outputState.Set("research_response", response.Content)
    
    return core.AgentResult{
        OutputState: outputState,
    }, nil
}

// Helper method for retry logic
func (a *ResearchAgent) retryWithPolicy(ctx context.Context, prompt core.Prompt) (core.AgentResult, error) {
    policy := a.config.RetryPolicy
    
    for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
        if attempt > 0 {
            delay := time.Duration(policy.BaseDelayMs) * time.Millisecond
            for i := 1; i < attempt; i++ {
                delay = time.Duration(float64(delay) * policy.BackoffFactor)
            }
            if delay > time.Duration(policy.MaxDelayMs)*time.Millisecond {
                delay = time.Duration(policy.MaxDelayMs) * time.Millisecond
            }
            time.Sleep(delay)
        }
        
        response, err := a.llm.Call(ctx, prompt)
        if err == nil {
            outputState := core.NewState()
            outputState.Set("research_response", response.Content)
            return core.AgentResult{OutputState: outputState}, nil
        }
        
        if attempt == policy.MaxRetries {
            return core.AgentResult{}, err
        }
    }
    
    return core.AgentResult{}, fmt.Errorf("max retries exceeded")
}
```

#### 3.3 Update Main Application

**Before (Hardcoded):**
```go
func main() {
    // Initialize LLM provider
    llmProvider, err := core.NewProviderFromWorkingDir()
    if err != nil {
        log.Fatal(err)
    }
    
    // Create hardcoded agents
    researcher := agents.NewResearchAgent(llmProvider)
    writer := agents.NewWriterAgent(llmProvider)
    
    // Register agents
    agentHandlers := map[string]core.AgentHandler{
        "researcher": researcher,
        "writer":     writer,
    }
    
    // Create runner
    runner, err := core.NewRunner(core.RunnerConfig{
        MaxConcurrentAgents: 10,
        TimeoutSeconds:      30,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Register agents
    for name, handler := range agentHandlers {
        runner.RegisterAgent(name, handler)
    }
    
    // Start workflow
    runner.Start(context.Background())
    // ... rest of application
}
```

**After (Configuration-Driven):**
```go
func main() {
    // Load configuration
    config, err := core.LoadConfig("agentflow.toml")
    if err != nil {
        log.Fatal(err)
    }
    
    // Initialize LLM provider
    llmProvider, err := core.NewProviderFromWorkingDir()
    if err != nil {
        log.Fatal(err)
    }
    
    // Create configuration resolver
    resolver := core.NewConfigResolver()
    
    // Create configurable agent factory
    factory := core.NewConfigurableAgentFactory(resolver)
    
    // Create agents from configuration
    agentHandlers := make(map[string]core.AgentHandler)
    
    for agentName := range config.Agents {
        // Resolve agent configuration
        resolvedConfig, err := resolver.ResolveAgentConfig(agentName, config)
        if err != nil {
            log.Fatalf("Failed to resolve config for agent %s: %v", agentName, err)
        }
        
        // Skip disabled agents
        if !resolvedConfig.Enabled {
            continue
        }
        
        // Create agent based on role
        var agent core.AgentHandler
        switch resolvedConfig.Role {
        case "research_specialist":
            agent = agents.NewResearchAgent(resolvedConfig, llmProvider)
        case "content_writer":
            agent = agents.NewWriterAgent(resolvedConfig, llmProvider)
        default:
            log.Printf("Unknown agent role: %s", resolvedConfig.Role)
            continue
        }
        
        agentHandlers[agentName] = agent
    }
    
    // Create runner from configuration
    runner, err := core.NewRunnerFromConfig("agentflow.toml")
    if err != nil {
        log.Fatal(err)
    }
    
    // Register agents
    for name, handler := range agentHandlers {
        runner.RegisterAgent(name, handler)
    }
    
    // Start workflow
    runner.Start(context.Background())
    // ... rest of application
}
```

### Phase 4: Testing and Validation

#### 4.1 Unit Testing

Create tests for configuration-driven agents:

```go
func TestResearchAgentWithConfiguration(t *testing.T) {
    // Create test configuration
    config := core.ResolvedAgentConfig{
        Name:         "test_researcher",
        Role:         "research_specialist",
        SystemPrompt: "You are a test research agent",
        Capabilities: []string{"research", "analysis"},
        Enabled:      true,
        Timeout:      30 * time.Second,
        LLMConfig: &core.ResolvedLLMConfig{
            Provider:    "mock",
            Model:       "test-model",
            Temperature: 0.3,
            MaxTokens:   1000,
        },
    }
    
    // Create mock LLM provider
    mockLLM := &MockLLMProvider{
        response: "Test research response",
    }
    
    // Create agent
    agent := agents.NewResearchAgent(config, mockLLM)
    
    // Test agent execution
    event := core.NewEvent("test", core.EventData{"message": "test query"}, nil)
    result, err := agent.Run(context.Background(), event, core.NewState())
    
    assert.NoError(t, err)
    assert.NotNil(t, result.OutputState)
    
    response, exists := result.OutputState.Get("research_response")
    assert.True(t, exists)
    assert.Equal(t, "Test research response", response)
}
```

#### 4.2 Integration Testing

Test the complete configuration-driven workflow:

```go
func TestConfigurationDrivenWorkflow(t *testing.T) {
    // Load test configuration
    config, err := core.LoadConfig("test-config.toml")
    require.NoError(t, err)
    
    // Validate configuration
    validator := core.NewDefaultConfigValidator()
    errors := validator.ValidateConfig(config)
    assert.Empty(t, errors, "Configuration should be valid")
    
    // Test agent creation from configuration
    resolver := core.NewConfigResolver()
    factory := core.NewConfigurableAgentFactory(resolver)
    
    agents, err := factory.CreateAllEnabledAgents(config)
    require.NoError(t, err)
    assert.Greater(t, len(agents), 0, "Should create agents from configuration")
    
    // Test workflow execution
    runner, err := core.NewRunnerFromConfig("test-config.toml")
    require.NoError(t, err)
    
    for name, agent := range agents {
        err := runner.RegisterAgent(name, agent)
        require.NoError(t, err)
    }
    
    // Execute test workflow
    runner.Start(context.Background())
    defer runner.Stop()
    
    // Test event processing
    event := core.NewEvent("test", core.EventData{"message": "test"}, nil)
    err = runner.Emit(event)
    assert.NoError(t, err)
}
```

#### 4.3 Configuration Validation Testing

```bash
# Validate migrated configuration
agentcli validate --level complete --verbose

# Test configuration with different environments
AGENTFLOW_LLM_PROVIDER=azure agentcli validate
AGENTFLOW_AGENT_RESEARCHER_TIMEOUT_SECONDS=60 agentcli validate

# Test hot-reload functionality
agentcli validate --watch
```

### Phase 5: Deployment and Monitoring

#### 5.1 Environment-Specific Configurations

Create configurations for different environments:

**Development (`agentflow.dev.toml`):**
```toml
[agent_flow]
name = "my-system-dev"
provider = "mock"

[logging]
level = "debug"

[agents.researcher]
timeout_seconds = 60  # Longer timeout for debugging
[agents.researcher.llm]
temperature = 0.5     # Higher creativity for testing
```

**Production (`agentflow.prod.toml`):**
```toml
[agent_flow]
name = "my-system-prod"
provider = "openai"

[logging]
level = "info"

[agents.researcher]
timeout_seconds = 30  # Optimized timeout
[agents.researcher.llm]
temperature = 0.2     # Lower for consistency
```

#### 5.2 Environment Variable Setup

```bash
# Production environment variables
export AGENTFLOW_LLM_PROVIDER="openai"
export OPENAI_API_KEY="your-api-key"
export AGENTFLOW_RUNTIME_MAX_CONCURRENT_AGENTS="20"

# Agent-specific overrides
export AGENTFLOW_AGENT_RESEARCHER_LLM_TEMPERATURE="0.1"
export AGENTFLOW_AGENT_WRITER_TIMEOUT_SECONDS="45"
```

#### 5.3 Monitoring and Alerting

Set up monitoring for configuration-driven systems:

```toml
[logging]
level = "info"
format = "json"

[runtime]
enable_metrics = true
metrics_port = 8080

# Configuration change webhooks
[webhooks]
on_config_change = "http://monitoring:8080/config-changed"
on_validation_error = "http://alerts:8080/validation-error"
```

## Common Migration Challenges

### Challenge 1: Complex Hardcoded Logic

**Problem**: Agents with complex hardcoded decision logic

**Solution**: Break down complex logic into configurable components

```go
// Before: Complex hardcoded logic
func (a *Agent) processData(data interface{}) (interface{}, error) {
    if strings.Contains(data.(string), "urgent") {
        return a.processUrgent(data)
    } else if strings.Contains(data.(string), "research") {
        return a.processResearch(data)
    }
    return a.processDefault(data)
}

// After: Configuration-driven routing
func (a *Agent) processData(data interface{}) (interface{}, error) {
    for _, rule := range a.config.ProcessingRules {
        if rule.Condition.Matches(data) {
            return a.processWithRule(data, rule)
        }
    }
    return a.processDefault(data)
}
```

### Challenge 2: Dynamic Agent Creation

**Problem**: Agents created dynamically based on runtime conditions

**Solution**: Use configuration templates and dynamic resolution

```go
// Dynamic agent creation from configuration
func (f *Factory) CreateAgentForTask(taskType string, config *core.Config) (core.AgentHandler, error) {
    // Find appropriate agent configuration
    for agentName, agentConfig := range config.Agents {
        if agentConfig.Metadata["task_type"] == taskType {
            resolvedConfig, err := f.resolver.ResolveAgentConfig(agentName, config)
            if err != nil {
                return nil, err
            }
            return f.createAgentFromConfig(resolvedConfig)
        }
    }
    return nil, fmt.Errorf("no agent found for task type: %s", taskType)
}
```

### Challenge 3: State Management

**Problem**: Agents with complex internal state

**Solution**: Externalize state configuration

```toml
[agents.stateful_agent]
role = "stateful_processor"
description = "Agent with configurable state management"

[agents.stateful_agent.state]
persistence = "memory"  # or "redis", "database"
ttl_seconds = 3600
max_entries = 1000

[agents.stateful_agent.metadata]
state_key_prefix = "agent_state_"
cleanup_interval = "1h"
```

## Best Practices for Migration

### 1. Incremental Migration

- Start with simple agents
- Migrate one agent at a time
- Test thoroughly at each step
- Keep rollback options available

### 2. Configuration Management

- Use version control for configurations
- Implement configuration validation in CI/CD
- Create environment-specific configurations
- Document configuration changes

### 3. Testing Strategy

- Create comprehensive test suites
- Test with different configurations
- Validate environment variable overrides
- Test hot-reload functionality

### 4. Monitoring and Observability

- Monitor configuration changes
- Track agent performance metrics
- Set up alerts for validation failures
- Log configuration resolution details

### 5. Documentation

- Document configuration options
- Create migration runbooks
- Maintain configuration examples
- Update team training materials

## Rollback Strategy

If migration issues occur, have a rollback plan:

### 1. Code Rollback

Keep the original hardcoded implementation available:

```go
// Feature flag for configuration-driven vs hardcoded
if useConfigDriven {
    agent = agents.NewConfigurableResearchAgent(config, llm)
} else {
    agent = agents.NewResearchAgent(llm) // Original implementation
}
```

### 2. Configuration Rollback

```bash
# Restore previous configuration
cp agentflow.toml.backup agentflow.toml

# Validate restored configuration
agentcli validate
```

### 3. Environment Rollback

```bash
# Clear environment overrides
unset AGENTFLOW_LLM_PROVIDER
unset AGENTFLOW_AGENT_RESEARCHER_ROLE

# Restart with original configuration
```

## Post-Migration Optimization

After successful migration:

### 1. Performance Tuning

```bash
# Analyze performance
agentcli config optimize --focus performance

# Apply optimizations
agentcli config optimize --output optimized-config.toml
```

### 2. Cost Optimization

```bash
# Optimize for cost
agentcli config optimize --focus cost

# Review token usage and model selection
agentcli validate --suggestions
```

### 3. Reliability Improvements

```bash
# Optimize for reliability
agentcli config optimize --focus reliability

# Review retry policies and error handling
agentcli validate --level strict
```

## Conclusion

Migrating from hardcoded to configuration-driven agents is a significant improvement that provides:

- **Flexibility**: Easy configuration changes without code deployment
- **Maintainability**: Centralized configuration management
- **Scalability**: Environment-specific optimizations
- **Reliability**: Comprehensive validation and error handling

The migration process requires careful planning and testing, but the benefits make it worthwhile for any serious AgenticGoKit deployment.

For additional support during migration:

1. Use the CLI tools for validation and optimization
2. Leverage templates for common patterns
3. Test thoroughly with different configurations
4. Monitor performance and adjust as needed
5. Keep documentation updated throughout the process