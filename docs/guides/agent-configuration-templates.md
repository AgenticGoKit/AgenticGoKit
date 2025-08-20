# Agent Configuration Templates

This guide covers the enhanced template system in AgenticGoKit that generates comprehensive agent configurations, making it easy to create sophisticated multi-agent systems with minimal setup.

## Overview

The template system has been enhanced to support:

- **Agent-specific configurations** with detailed settings
- **Comprehensive TOML generation** with all configuration options
- **Role-based agent design** with specialized capabilities
- **Performance optimization** settings and tuning
- **Integration configurations** for memory, MCP, and orchestration

## Template Structure

Templates are defined in YAML format and include:

```yaml
name: "Template Name"
description: "Template description"
features:
  - "feature1"
  - "feature2"

config:
  # Project-level configuration
  numAgents: 3
  provider: "openai"
  orchestrationMode: "sequential"
  # ... other settings

agents:
  # Agent-specific configurations
  agent-name:
    role: "agent_role"
    description: "Agent description"
    capabilities: ["cap1", "cap2"]
    systemPrompt: |
      Detailed system prompt...
    llm:
      temperature: 0.3
      maxTokens: 2000
    retryPolicy:
      maxRetries: 3
      baseDelayMs: 1000
    # ... other agent settings

mcpServers:
  # MCP server configurations
  - name: "server_name"
    type: "stdio"
    command: "server_command"
    enabled: true
```

## Available Templates

### 1. Simple Workflow (`simple-workflow.yaml`)

A basic three-agent sequential workflow for general-purpose processing.

**Features:**
- Simple setup with minimal configuration
- Sequential processing pipeline
- Basic error handling
- No external dependencies

**Agents:**
- `processor`: Data processing and validation
- `analyzer`: Data analysis and insights
- `formatter`: Output formatting

**Use Cases:**
- Learning and experimentation
- Simple data processing pipelines
- Proof of concept projects

### 2. Research Assistant (`research-assistant.yaml`)

A comprehensive research system with information gathering, fact-checking, and synthesis.

**Features:**
- Advanced memory system with RAG
- Fact-checking and validation
- Multi-source information gathering
- Citation and reference management

**Agents:**
- `information-gatherer`: Comprehensive information research
- `fact-checker`: Information validation and verification
- `analyzer`: Pattern recognition and trend analysis
- `synthesizer`: Research synthesis and reporting

**Use Cases:**
- Academic research projects
- Market research and analysis
- Due diligence and investigation
- Knowledge base creation

### 3. Content Creation (`content-creation.yaml`)

A multi-agent content creation pipeline with research, writing, editing, and optimization.

**Features:**
- SEO optimization and keyword analysis
- Multi-format content support
- Quality assurance and editing
- Brand consistency and style guides

**Agents:**
- `content-researcher`: Topic research and trend analysis
- `content-writer`: Creative content generation
- `content-editor`: Quality improvement and editing
- `seo-optimizer`: Search engine optimization
- `quality-checker`: Final quality assurance

**Use Cases:**
- Blog and article creation
- Marketing content development
- Technical documentation
- Social media content

### 4. Customer Support (`customer-support.yaml`)

A collaborative customer support system with ticket routing and resolution.

**Features:**
- Intelligent ticket classification
- Sentiment analysis and escalation
- Knowledge base integration
- Performance tracking and analytics

**Agents:**
- `ticket-classifier`: Ticket categorization and routing
- `support-agent`: Issue resolution and customer service
- `escalation-manager`: Complex case management
- `satisfaction-tracker`: Quality and satisfaction analysis

**Use Cases:**
- Customer service automation
- Help desk systems
- Technical support
- Service quality improvement

### 5. Custom RAG System (`custom-rag.yaml`)

A specialized RAG (Retrieval-Augmented Generation) system for knowledge-based applications.

**Features:**
- Advanced document processing
- Semantic search and retrieval
- Multi-modal content support
- Context-aware response generation

**Agents:**
- `document-ingester`: Document processing and indexing
- `query-processor`: Query analysis and optimization
- `retrieval-agent`: Information retrieval and ranking
- `response-generator`: Context-aware response synthesis

**Use Cases:**
- Knowledge management systems
- Document Q&A systems
- Research databases
- Enterprise search solutions

### 6. Advanced Configuration Demo (`advanced-config-demo.yaml`)

A comprehensive demonstration of all configuration features and capabilities.

**Features:**
- Complete feature showcase
- Performance optimization examples
- Enterprise-grade configurations
- Advanced integration patterns

**Agents:**
- `input-processor`: Advanced input processing and validation
- `data-validator`: Comprehensive data quality assurance
- `analyzer`: Deep analysis with ML insights
- `synthesizer`: Knowledge synthesis and integration
- `quality-checker`: Multi-dimensional quality assurance
- `output-formatter`: Multi-channel output optimization

**Use Cases:**
- Learning advanced configurations
- Enterprise system templates
- Performance optimization examples
- Feature exploration and testing

## Agent Configuration Features

### Core Agent Settings

```yaml
agents:
  agent-name:
    role: "agent_role"                    # Agent's functional role
    description: "Detailed description"   # Agent purpose and responsibilities
    capabilities: ["cap1", "cap2"]        # Agent capabilities list
    systemPrompt: |                       # Detailed system prompt
      Multi-line system prompt with
      specific instructions and context
    enabled: true                         # Enable/disable agent
    timeout: 45                          # Timeout in seconds
```

### LLM Configuration

```yaml
    llm:
      provider: "openai"                  # Override global provider
      model: "gpt-4"                      # Specific model for agent
      temperature: 0.3                    # Creativity/randomness control
      maxTokens: 2000                     # Maximum response length
      topP: 0.9                          # Nucleus sampling parameter
      frequencyPenalty: 0.1               # Repetition penalty
      presencePenalty: 0.1                # Topic diversity penalty
```

### Retry Policy Configuration

```yaml
    retryPolicy:
      maxRetries: 3                       # Maximum retry attempts
      baseDelayMs: 1000                   # Initial delay between retries
      maxDelayMs: 5000                    # Maximum delay cap
      backoffFactor: 2.0                  # Exponential backoff multiplier
```

### Rate Limiting Configuration

```yaml
    rateLimit:
      requestsPerSecond: 10               # Maximum requests per second
      burstSize: 20                       # Burst capacity for spikes
```

### Metadata and Tagging

```yaml
    metadata:
      specialization: "data_processing"   # Agent specialization
      priority: "high"                    # Processing priority
      performance_tier: "optimized"       # Performance classification
      compliance_level: "strict"          # Compliance requirements
```

## Memory System Configuration

Templates can include comprehensive memory system settings:

```yaml
config:
  memoryEnabled: true
  memoryProvider: "pgvector"              # or "weaviate", "memory"
  embeddingProvider: "openai"             # Embedding service
  embeddingModel: "text-embedding-3-small"
  
  # RAG Configuration
  ragEnabled: true
  ragChunkSize: 1500                      # Document chunk size
  ragOverlap: 150                         # Chunk overlap
  ragTopK: 8                             # Top results to retrieve
  ragScoreThreshold: 0.75                 # Minimum relevance score
  hybridSearch: true                      # Combine semantic + keyword
  sessionMemory: true                     # Enable session context
```

## MCP Server Configuration

Templates can define MCP (Model Context Protocol) servers:

```yaml
mcpServers:
  - name: "web_search"
    type: "stdio"
    command: "npx @modelcontextprotocol/server-brave-search"
    enabled: true
  - name: "database_query"
    type: "tcp"
    host: "localhost"
    port: 8814
    enabled: false
```

## Using Templates

### 1. List Available Templates

```go
templates, err := scaffold.ListAvailableTemplates()
if err != nil {
    log.Fatalf("Failed to list templates: %v", err)
}

for _, template := range templates {
    fmt.Println(template)
}
```

### 2. Get Template Information

```go
info, err := scaffold.GetTemplateInfo("research-assistant")
if err != nil {
    log.Fatalf("Failed to get template info: %v", err)
}

fmt.Printf("Template: %s\n", info.Name)
fmt.Printf("Description: %s\n", info.Description)
fmt.Printf("Agents: %d\n", info.Config.NumAgents)
```

### 3. Create Project from Template

```go
err := scaffold.CreateAgentProjectFromTemplate("research-assistant", "my-research-project")
if err != nil {
    log.Fatalf("Failed to create project: %v", err)
}
```

### 4. Using the CLI (when available)

```bash
# List available templates
agentcli template list

# Show template information
agentcli template info research-assistant

# Create project from template
agentcli create --template research-assistant my-project
```

## Creating Custom Templates

### 1. Template File Structure

Create a new YAML file in `examples/templates/`:

```yaml
name: "My Custom Template"
description: "Custom template for specific use case"
features:
  - "custom-feature1"
  - "custom-feature2"

config:
  numAgents: 2
  provider: "openai"
  orchestrationMode: "collaborative"
  # ... configuration settings

agents:
  agent1:
    role: "custom_role"
    description: "Custom agent description"
    capabilities: ["custom_capability"]
    systemPrompt: |
      Custom system prompt for this agent...
    # ... agent settings

mcpServers: []
```

### 2. Agent Design Best Practices

**Role Definition:**
- Use descriptive, functional role names
- Follow snake_case naming convention
- Align roles with agent capabilities

**System Prompts:**
- Provide detailed, specific instructions
- Include context and expectations
- Specify output format requirements
- Add examples when helpful

**Capability Selection:**
- Choose capabilities that match the agent's role
- Use standard capability names when possible
- Group related capabilities together

**Performance Tuning:**
- Set appropriate timeouts for agent complexity
- Configure retry policies for reliability
- Use rate limiting for resource management
- Optimize LLM parameters for the use case

### 3. Configuration Validation

Templates are automatically validated when loaded:

- **Required fields** are checked
- **Configuration consistency** is verified
- **Agent references** are validated
- **Resource requirements** are assessed

## Generated Project Structure

When creating a project from a template, the following is generated:

```
my-project/
├── agents/                    # Agent implementations
│   ├── agent1.go             # Generated agent code
│   └── agent2.go
├── agentflow.toml            # Enhanced configuration
├── main.go                   # Main application
├── go.mod                    # Go module file
├── README.md                 # Project documentation
└── docker-compose.yml       # Database setup (if needed)
```

### Enhanced agentflow.toml

The generated configuration includes:

```toml
# Global configuration
[agent_flow]
name = "my-project"
provider = "openai"

# Agent-specific configurations
[agents.agent1]
role = "custom_role"
description = "Agent description"
system_prompt = """Detailed system prompt..."""
capabilities = ["capability1", "capability2"]
enabled = true
timeout_seconds = 30

[agents.agent1.llm]
temperature = 0.3
max_tokens = 2000

[agents.agent1.retry_policy]
max_retries = 3
base_delay_ms = 1000
max_delay_ms = 5000
backoff_factor = 2.0

# Memory configuration (if enabled)
[agent_memory]
provider = "pgvector"
connection = "postgres://..."
# ... memory settings

# MCP configuration (if enabled)
[mcp]
enabled = true
# ... MCP settings

# Orchestration configuration
[orchestration]
mode = "collaborative"
# ... orchestration settings
```

## Best Practices

### Template Design

1. **Start Simple**: Begin with basic configurations and add complexity gradually
2. **Use Descriptive Names**: Make agent roles and capabilities self-explanatory
3. **Provide Context**: Include detailed system prompts and descriptions
4. **Optimize Performance**: Set appropriate timeouts and resource limits
5. **Plan for Scale**: Consider resource usage and performance implications

### Agent Configuration

1. **Role Clarity**: Each agent should have a clear, specific role
2. **Capability Alignment**: Capabilities should match the agent's intended function
3. **Prompt Engineering**: Invest time in crafting effective system prompts
4. **Performance Tuning**: Adjust LLM parameters for optimal results
5. **Error Handling**: Configure appropriate retry policies and timeouts

### System Integration

1. **Memory Strategy**: Choose appropriate memory providers for your use case
2. **MCP Planning**: Select MCP tools that enhance agent capabilities
3. **Orchestration Design**: Match orchestration patterns to workflow requirements
4. **Resource Management**: Plan for computational and memory requirements
5. **Monitoring Setup**: Include metrics and logging configurations

## Troubleshooting

### Common Issues

1. **Template Loading Errors**
   - Check YAML syntax and structure
   - Verify all required fields are present
   - Ensure agent names are valid

2. **Configuration Validation Failures**
   - Review agent capability names
   - Check LLM parameter ranges
   - Verify orchestration agent references

3. **Performance Issues**
   - Adjust timeout values for complex agents
   - Optimize LLM parameters for speed vs. quality
   - Configure appropriate rate limits

4. **Integration Problems**
   - Verify memory provider connections
   - Check MCP server availability
   - Validate environment variables

### Getting Help

- Review the validation error messages for specific guidance
- Check the comprehensive validation system output
- Examine working templates for reference patterns
- Use the template usage examples for guidance

## Conclusion

The enhanced template system provides a powerful way to create sophisticated multi-agent systems with comprehensive configurations. By leveraging templates, you can:

- **Accelerate Development**: Start with proven patterns and configurations
- **Ensure Best Practices**: Benefit from optimized settings and structures
- **Reduce Complexity**: Abstract away configuration details
- **Improve Reliability**: Use tested retry policies and error handling
- **Enable Customization**: Easily modify templates for specific needs

The system supports everything from simple workflows to enterprise-grade multi-agent systems, making it easy to build powerful AI applications with AgenticGoKit.