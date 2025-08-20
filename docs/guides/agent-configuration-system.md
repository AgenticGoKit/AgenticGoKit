# Agent Configuration System

This comprehensive guide covers the complete agent configuration system in AgenticGoKit, including configuration-driven development, validation, hot-reload, and management tools.

## Table of Contents

1. [Overview](#overview)
2. [Configuration Structure](#configuration-structure)
3. [Agent Configuration](#agent-configuration)
4. [Configuration Resolution](#configuration-resolution)
5. [Validation System](#validation-system)
6. [Hot-Reload Support](#hot-reload-support)
7. [CLI Management Tools](#cli-management-tools)
8. [Template System](#template-system)
9. [Best Practices](#best-practices)
10. [Migration Guide](#migration-guide)
11. [Troubleshooting](#troubleshooting)

## Overview

The AgenticGoKit agent configuration system provides a comprehensive, flexible approach to defining and managing multi-agent systems through configuration files rather than hardcoded values.

### Key Features

- **Configuration-Driven Development**: Define agents through TOML configuration
- **Comprehensive Validation**: Extensive validation with helpful error messages
- **Hot-Reload Support**: Dynamic configuration updates without restart
- **Environment Overrides**: Environment variable support for deployment
- **Template System**: Pre-built templates for common use cases
- **CLI Management**: Powerful CLI tools for configuration management
- **Performance Optimization**: Built-in performance recommendations

### Benefits

- **Rapid Development**: Create sophisticated agents without coding
- **Easy Deployment**: Environment-specific configurations
- **Maintainability**: Centralized configuration management
- **Flexibility**: Runtime configuration changes
- **Reliability**: Comprehensive validation and error handling

## Configuration Structure

### Basic Configuration File (`agentflow.toml`)

```toml
# Project Information
[agent_flow]
name = "my-agent-system"
version = "1.0.0"
provider = "openai"

# Global LLM Configuration
[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 2000

# Runtime Settings
[runtime]
max_concurrent_agents = 10
timeout_seconds = 30

# Logging Configuration
[logging]
level = "info"
format = "json"

# Agent Definitions
[agents.researcher]
role = "research_specialist"
description = "Conducts comprehensive research on various topics"
system_prompt = """
You are a research specialist focused on gathering accurate information.
Your role is to find authoritative sources and provide well-researched content.
"""
capabilities = ["information_gathering", "fact_checking", "source_identification"]
enabled = true
auto_llm = true
timeout_seconds = 45

# Agent-specific LLM overrides
[agents.researcher.llm]
temperature = 0.3
max_tokens = 2500

# Retry policy for reliability
[agents.researcher.retry_policy]
max_retries = 3
base_delay_ms = 1000
max_delay_ms = 5000
backoff_factor = 2.0

# Rate limiting
[agents.researcher.rate_limit]
requests_per_second = 10
burst_size = 20

# Agent metadata
[agents.researcher.metadata]
specialization = "research"
priority = "high"
cost_tier = "standard"

# Orchestration Configuration
[orchestration]
mode = "sequential"
timeout_seconds = 300
sequential_agents = ["researcher", "analyzer", "writer"]

# Memory System (optional)
[agent_memory]
provider = "pgvector"
connection = "postgres://user:password@localhost:15432/agentflow"
dimensions = 1536
enable_rag = true
chunk_size = 1500
chunk_overlap = 150

# MCP Configuration (optional)
[mcp]
enabled = true
enable_discovery = true
connection_timeout = 5000

[[mcp.servers]]
name = "web_search"
type = "stdio"
command = "npx @modelcontextprotocol/server-brave-search"
enabled = true
```

## Agent Configuration

### Core Agent Settings

Every agent requires these fundamental settings:

```toml
[agents.agent_name]
role = "agent_role"                    # Functional role identifier
description = "Agent description"       # Human-readable description
system_prompt = """Multi-line prompt""" # Detailed system instructions
capabilities = ["cap1", "cap2"]         # List of agent capabilities
enabled = true                          # Enable/disable agent
timeout_seconds = 30                    # Execution timeout
```

### Advanced Agent Settings

#### LLM Configuration Override

```toml
[agents.agent_name.llm]
provider = "openai"                     # Override global provider
model = "gpt-4"                        # Specific model
temperature = 0.3                       # Creativity control
max_tokens = 2000                      # Response length limit
top_p = 0.9                           # Nucleus sampling
frequency_penalty = 0.1                # Repetition penalty
presence_penalty = 0.1                 # Topic diversity
```

#### Retry Policy Configuration

```toml
[agents.agent_name.retry_policy]
max_retries = 3                        # Maximum retry attempts
base_delay_ms = 1000                   # Initial delay
max_delay_ms = 5000                    # Maximum delay cap
backoff_factor = 2.0                   # Exponential backoff
```

#### Rate Limiting

```toml
[agents.agent_name.rate_limit]
requests_per_second = 10               # Rate limit
burst_size = 20                        # Burst capacity
```

#### Metadata and Tagging

```toml
[agents.agent_name.metadata]
specialization = "data_processing"      # Agent specialization
priority = "high"                      # Processing priority
cost_tier = "premium"                  # Cost classification
team = "research"                      # Team assignment
version = "2.1"                        # Agent version
```

### Capability System

AgenticGoKit includes a comprehensive capability system:

#### Research Capabilities
- `information_gathering` - Web search and data collection
- `fact_checking` - Information verification
- `source_identification` - Source credibility assessment
- `data_processing` - Data cleaning and transformation

#### Analysis Capabilities
- `pattern_recognition` - Pattern detection in data
- `trend_analysis` - Trend identification and analysis
- `insight_generation` - Insight extraction from data
- `data_analysis` - Statistical and analytical processing

#### Content Capabilities
- `text_analysis` - Text processing and analysis
- `summarization` - Content summarization
- `translation` - Language translation
- `content_creation` - Original content generation
- `editing` - Content editing and improvement

#### Development Capabilities
- `code_generation` - Code creation and scaffolding
- `code_review` - Code quality assessment
- `debugging` - Error identification and fixing
- `testing` - Test creation and execution
- `documentation` - Documentation generation

## Configuration Resolution

The configuration system uses a hierarchical resolution approach:

### Resolution Order

1. **Environment Variables** (highest priority)
2. **Agent-Specific Configuration**
3. **Global Configuration**
4. **System Defaults** (lowest priority)

### Environment Variable Overrides

```bash
# Global LLM settings
export AGENTFLOW_LLM_PROVIDER="azure"
export AGENTFLOW_LLM_TEMPERATURE="0.5"
export AGENTFLOW_LLM_MAX_TOKENS="1500"

# Agent-specific settings
export AGENTFLOW_AGENT_RESEARCHER_ROLE="senior_researcher"
export AGENTFLOW_AGENT_RESEARCHER_TIMEOUT_SECONDS="60"

# Agent-specific LLM settings
export AGENTFLOW_AGENT_RESEARCHER_LLM_TEMPERATURE="0.2"
export AGENTFLOW_AGENT_RESEARCHER_LLM_MAX_TOKENS="3000"
```

### Configuration Resolution Example

```go
// Load and resolve configuration
config, err := core.LoadConfig("agentflow.toml")
if err != nil {
    log.Fatal(err)
}

// Create resolver
resolver := core.NewConfigResolver()

// Resolve agent configuration
resolvedConfig, err := resolver.ResolveAgentConfig("researcher", config)
if err != nil {
    log.Fatal(err)
}

// Use resolved configuration
fmt.Printf("Agent: %s\n", resolvedConfig.Name)
fmt.Printf("Role: %s\n", resolvedConfig.Role)
fmt.Printf("Temperature: %.2f\n", resolvedConfig.LLMConfig.Temperature)
```

## Validation System

### Comprehensive Validation

The validation system performs extensive checks:

#### Configuration Structure Validation
- TOML syntax and structure
- Required field presence
- Data type validation
- Value range validation

#### Agent Configuration Validation
- Role and capability validation
- System prompt completeness
- LLM parameter ranges
- Timeout and retry settings

#### Cross-Reference Validation
- Orchestration agent references
- Capability consistency
- Provider compatibility

#### Performance Validation
- Resource usage optimization
- Timeout recommendations
- Cost optimization suggestions

### Using the Validation System

```go
// Create validator
validator := core.NewDefaultConfigValidator()

// Validate complete configuration
errors := validator.ValidateConfig(config)

// Process validation results
for _, err := range errors {
    fmt.Printf("Field: %s\n", err.Field)
    fmt.Printf("Issue: %s\n", err.Message)
    fmt.Printf("Suggestion: %s\n", err.Suggestion)
}
```

### CLI Validation

```bash
# Basic validation
agentcli validate

# Comprehensive validation with suggestions
agentcli validate --level strict --verbose

# Validate specific aspects
agentcli validate --scope agents-only
agentcli validate --scope config-only

# Output in different formats
agentcli validate --output json
agentcli validate --output yaml
```

## Hot-Reload Support

### Configuration Hot-Reload

The system supports dynamic configuration updates:

```go
// Create config reloader
reloader, err := core.NewConfigReloader("agentflow.toml")
if err != nil {
    log.Fatal(err)
}

// Set up change handler
reloader.OnConfigChange(func(newConfig *core.Config) {
    fmt.Println("Configuration updated!")
    // Update agents with new configuration
})

// Start watching for changes
reloader.Start()
defer reloader.Stop()
```

### Safe Reload Process

1. **File Change Detection**: Monitor configuration file changes
2. **Validation**: Validate new configuration before applying
3. **Rollback**: Automatic rollback on validation failure
4. **Agent Update**: Update running agents with new settings
5. **Notification**: Notify system of successful updates

### Hot-Reload Features

- **Validation Before Apply**: Prevents invalid configurations
- **Atomic Updates**: All-or-nothing configuration changes
- **Rollback Support**: Automatic rollback on failures
- **Event Notifications**: Configuration change events
- **Graceful Updates**: Non-disruptive agent updates

## CLI Management Tools

### Configuration Validation

```bash
# Validate current project
agentcli validate

# Validate with comprehensive checks
agentcli validate --level complete --verbose

# Validate specific configuration file
agentcli validate my-config.toml

# Show optimization suggestions
agentcli validate --suggestions
```

### Configuration Generation

```bash
# Generate configuration from existing code
agentcli config generate

# Generate to specific file
agentcli config generate my-config.toml

# Generate in different format
agentcli config generate --format yaml config.yaml
```

### Configuration Migration

```bash
# Migrate to latest format
agentcli config migrate

# Migrate specific file
agentcli config migrate old-config.toml

# Preview migration changes
agentcli config migrate --dry-run
```

### Configuration Optimization

```bash
# Optimize for performance
agentcli config optimize --focus performance

# Show optimization recommendations
agentcli config optimize --recommend-only

# Optimize specific aspects
agentcli config optimize --focus cost
agentcli config optimize --focus reliability
```

### Template-Based Configuration

```bash
# List available templates
agentcli config template --list

# Generate from template
agentcli config template research-assistant > config.toml

# Customize template
agentcli config template rag-system --memory pgvector --embedding openai
```

## Template System

### Available Templates

#### Research Assistant
```bash
agentcli create research-project --template research-assistant
```
- Multi-agent research system
- Web search and fact-checking
- Information synthesis
- Citation management

#### Content Creation Pipeline
```bash
agentcli create content-system --template content-creation
```
- Research and writing workflow
- SEO optimization
- Quality assurance
- Multi-format output

#### Customer Support System
```bash
agentcli create support-system --template customer-support
```
- Ticket classification and routing
- Automated resolution
- Escalation management
- Satisfaction tracking

#### RAG System
```bash
agentcli create knowledge-base --template custom-rag
```
- Document processing and indexing
- Semantic search and retrieval
- Context-aware responses
- Multi-modal support

### Custom Templates

Create custom templates in YAML format:

```yaml
name: "Custom Template"
description: "Template for specific use case"
features:
  - "custom-feature"

config:
  numAgents: 3
  provider: "openai"
  orchestrationMode: "sequential"
  
agents:
  agent1:
    role: "custom_role"
    description: "Custom agent"
    capabilities: ["custom_capability"]
    systemPrompt: |
      Custom system prompt...
```

## Best Practices

### Configuration Organization

1. **Logical Grouping**: Group related agents together
2. **Consistent Naming**: Use descriptive, consistent names
3. **Documentation**: Include clear descriptions and comments
4. **Version Control**: Track configuration changes
5. **Environment Separation**: Separate configs for dev/staging/prod

### Agent Design

1. **Single Responsibility**: Each agent should have a clear, focused role
2. **Capability Alignment**: Capabilities should match agent functions
3. **Prompt Engineering**: Invest time in effective system prompts
4. **Performance Tuning**: Optimize LLM parameters for use case
5. **Error Handling**: Configure appropriate retry policies

### Performance Optimization

1. **Timeout Management**: Set appropriate timeouts for each agent
2. **Resource Limits**: Configure memory and token limits
3. **Rate Limiting**: Prevent API rate limit issues
4. **Caching**: Use caching for repeated operations
5. **Monitoring**: Track performance metrics

### Security Considerations

1. **Credential Management**: Use environment variables for secrets
2. **Input Validation**: Validate all configuration inputs
3. **Access Control**: Limit configuration file access
4. **Audit Logging**: Log configuration changes
5. **Backup Strategy**: Regular configuration backups

## Migration Guide

### From Hardcoded to Configuration-Driven

#### Step 1: Generate Initial Configuration

```bash
# Generate configuration from existing code
agentcli config generate
```

#### Step 2: Review and Customize

```bash
# Validate generated configuration
agentcli validate --level complete

# Optimize configuration
agentcli config optimize
```

#### Step 3: Update Agent Code

```go
// Before: Hardcoded agent
type ResearchAgent struct {
    llm core.ModelProvider
}

func (a *ResearchAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    systemPrompt := "You are a research specialist..." // Hardcoded
    temperature := 0.3 // Hardcoded
    // ...
}

// After: Configuration-driven agent
type ConfigurableResearchAgent struct {
    config core.ResolvedAgentConfig
    llm    core.ModelProvider
}

func NewConfigurableResearchAgent(config core.ResolvedAgentConfig, llm core.ModelProvider) *ConfigurableResearchAgent {
    return &ConfigurableResearchAgent{
        config: config,
        llm:    llm,
    }
}

func (a *ConfigurableResearchAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    systemPrompt := a.config.SystemPrompt // From configuration
    temperature := a.config.LLMConfig.Temperature // From configuration
    // ...
}
```

#### Step 4: Test and Validate

```bash
# Test with new configuration
go run . -m "test message"

# Validate configuration
agentcli validate --level strict
```

### Version Migration

#### Configuration Format Updates

When AgenticGoKit introduces new configuration formats:

```bash
# Check current version compatibility
agentcli config migrate --dry-run

# Migrate to latest format
agentcli config migrate --backup

# Validate migrated configuration
agentcli validate
```

## Troubleshooting

### Common Issues

#### Configuration Loading Errors

**Problem**: Configuration file not found or invalid syntax

**Solution**:
```bash
# Check file exists and has correct syntax
agentcli validate --scope config-only

# Generate new configuration if needed
agentcli config generate
```

#### Agent Configuration Errors

**Problem**: Agent configuration validation failures

**Solution**:
```bash
# Validate agent configurations specifically
agentcli validate --scope agents-only --verbose

# Check capability names and requirements
agentcli validate --level complete
```

#### Environment Override Issues

**Problem**: Environment variables not being applied

**Solution**:
```bash
# Check environment variable names
echo $AGENTFLOW_LLM_PROVIDER
echo $AGENTFLOW_AGENT_RESEARCHER_ROLE

# Validate resolved configuration
agentcli validate --verbose
```

#### Hot-Reload Problems

**Problem**: Configuration changes not being applied

**Solution**:
```go
// Check file watcher is running
reloader.IsWatching() // Should return true

// Check for validation errors in logs
// Ensure file permissions allow reading
```

### Performance Issues

#### High Memory Usage

**Problem**: Agents consuming too much memory

**Solution**:
```bash
# Optimize for memory usage
agentcli config optimize --focus memory

# Check token limits and batch sizes
agentcli validate --level strict
```

#### Slow Response Times

**Problem**: Agents taking too long to respond

**Solution**:
```bash
# Optimize for performance
agentcli config optimize --focus performance

# Check timeout settings and LLM parameters
agentcli validate --suggestions
```

### Debugging Tools

#### Configuration Inspection

```bash
# Show resolved configuration
agentcli config show --resolved

# Compare configurations
agentcli config diff config1.toml config2.toml

# Validate specific agent
agentcli validate --agent researcher
```

#### Logging and Monitoring

```toml
[logging]
level = "debug"
format = "json"

[runtime]
enable_metrics = true
metrics_port = 8080
```

## Advanced Topics

### Custom Validation Rules

```go
// Create custom validation rule
type CustomValidationRule struct{}

func (r *CustomValidationRule) Name() string {
    return "custom_rule"
}

func (r *CustomValidationRule) Description() string {
    return "Custom validation logic"
}

func (r *CustomValidationRule) Validate(config *core.Config) []core.ValidationError {
    var errors []core.ValidationError
    // Custom validation logic
    return errors
}

// Add to validator
validator := core.NewDefaultConfigValidator()
validator.AddValidationRule(&CustomValidationRule{})
```

### Configuration Preprocessing

```go
// Custom configuration preprocessor
type ConfigPreprocessor struct{}

func (p *ConfigPreprocessor) Preprocess(config *core.Config) (*core.Config, error) {
    // Apply custom transformations
    for name, agent := range config.Agents {
        if agent.Role == "" {
            agent.Role = fmt.Sprintf("%s_role", name)
            config.Agents[name] = agent
        }
    }
    return config, nil
}
```

### Integration with External Systems

```toml
# External configuration sources
[external_config]
enabled = true
source = "consul"
endpoint = "http://consul:8500"
prefix = "agentflow/"

# Configuration webhooks
[webhooks]
on_config_change = "http://monitoring:8080/config-changed"
on_validation_error = "http://alerts:8080/validation-error"
```

## Conclusion

The AgenticGoKit agent configuration system provides a powerful, flexible foundation for building and managing sophisticated multi-agent systems. By leveraging configuration-driven development, comprehensive validation, hot-reload capabilities, and powerful management tools, you can create maintainable, scalable, and reliable AI applications.

Key takeaways:

1. **Start with Templates**: Use pre-built templates for common patterns
2. **Validate Early and Often**: Use comprehensive validation to catch issues
3. **Optimize Continuously**: Regularly review and optimize configurations
4. **Monitor Performance**: Track metrics and adjust settings as needed
5. **Plan for Scale**: Design configurations with growth in mind

For additional help and examples, refer to the template system documentation and CLI reference guides.