# testAgent

**An intelligent multi-agent system powered by AgenticGoKit**

testAgent is a sophisticated multi-agent workflow system that leverages multiple AI agents working sequentially to process and respond to user queries.

## Quick Start

### Prerequisites

- Go 1.21 or later
- OpenAI API key

- PostgreSQL with pgvector extension

- Node.js (for MCP tools)

### Installation

1. **Clone and setup**:
   ```bash
   git clone <your-repository>
   cd testAgent
   go mod tidy
   ```

2. **Configure environment**:
   ```bash
   export OPENAI_API_KEY="your-api-key"
   ```

3. **Setup PostgreSQL with pgvector**:
   ```bash
   # Start database
   docker compose up -d
   
   # Run setup script
   ./setup.sh  # or setup.bat on Windows
   ```

4. **Setup MCP tools** (optional):
   ```bash
   # Install MCP servers
   npm install -g @modelcontextprotocol/server-filesystem
   # Add other MCP servers as needed
   ```

### Running the System

```bash
# Validate configuration first
agentcli validate agentflow.toml

# Interactive mode
go run . 

# Command line mode
go run . -m "Your message here"

# With debug logging
LOG_LEVEL=debug go run . -m "Your message"
```

## Configuration-Driven Architecture

This project uses **AgentFlow's configuration-driven architecture**:

- **No hardcoded agents**: All agents defined in `agentflow.toml`
- **Flexible configuration**: Change behavior without code changes  
- **Hot reload support**: Update config without restarting
- **Environment-specific**: Different settings per environment
- **Built-in validation**: Comprehensive validation with helpful errors

### Key Configuration Files

- **`agentflow.toml`**: Main configuration (agents, LLM, orchestration)
- **`agents/`**: Reference implementations (optional)
- **Environment variables**: Sensitive data (API keys)

### Configuration Management

```bash
# Validate configuration
agentcli validate agentflow.toml

# Generate new configuration from template
agentcli config generate research-assistant my-project

# Get detailed validation report
agentcli validate --detailed agentflow.toml

# Export configuration schema
agentcli config schema --generate
```

### Configuration Example

```toml
# Global LLM settings
[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7

# Agent definitions
[agents.researcher]
role = "researcher"
description = "Researches topics and gathers comprehensive information"
system_prompt = "You are Researcher, Researches topics and gathers comprehensive information"
capabilities = ["general_assistance", "processing"]
enabled = true

# Agent-specific LLM settings
[agents.researcher.llm]
temperature = 0.7
max_tokens = 2000

[agents.analyzer]
role = "analyzer"
description = "Analyzes and processes input data to extract insights"
system_prompt = "You are Analyzer, Analyzes and processes input data to extract insights"
capabilities = ["general_assistance", "processing"]
enabled = true

# Agent-specific LLM settings
[agents.analyzer.llm]
temperature = 0.7
max_tokens = 2000

[agents.synthesizer]
role = "synthesizer"
description = "Synthesizes information and creates comprehensive responses"
system_prompt = "You are Synthesizer, Synthesizes information and creates comprehensive responses"
capabilities = ["general_assistance", "processing"]
enabled = true

# Agent-specific LLM settings
[agents.synthesizer.llm]
temperature = 0.7
max_tokens = 2000

# Orchestration
[orchestration]
mode = "sequential"
agents = ["researcher", "analyzer", "synthesizer"]
```

## Architecture

### System Overview

testAgent implements a sequential multi-agent architecture with 3 specialized agents:

```
User Input -> Agent1 -> Agent2 -> ... -> Final Response
```

### Project Structure

```
testAgent/
|-- agents/                 # Agent implementations
|   |-- researcher.go           # Researcher agent
|   |-- analyzer.go           # Analyzer agent
|   |-- synthesizer.go           # Synthesizer agent
|   `-- README.md           # Agent documentation
|-- internal/               # Internal packages
|   |-- config/             # Configuration utilities
|   `-- handlers/           # Shared handler utilities
|-- docs/                   # Documentation
|   `-- CUSTOMIZATION.md    # Customization guide
|-- main.go                 # Application entry point
|-- agentflow.toml          # Main configuration file
|-- docker-compose.yml        # Database services
|-- setup.sh                 # Database setup script
|-- agentflow.toml          # System configuration
|-- go.mod                  # Go module definition
|-- docker-compose.yml        # Database services
|-- setup.sh                # Database setup script
`-- README.md               # This file
```

### Agent Responsibilities


#### Researcher (`agents/researcher.go`)

**Purpose**: Researches topics and gathers comprehensive information

**Role**: Processes initial user input and prepares data for downstream agents

**Key Features**:
- Memory-enabled for context retention
- Tool integration via MCP
- RAG capabilities for knowledge retrieval


#### Analyzer (`agents/analyzer.go`)

**Purpose**: Analyzes and processes input data to extract insights

**Role**: Processes input from previous agents and passes refined data forward

**Key Features**:
- Memory-enabled for context retention
- Tool integration via MCP
- RAG capabilities for knowledge retrieval


#### Synthesizer (`agents/synthesizer.go`)

**Purpose**: Synthesizes information and creates comprehensive responses

**Role**: Finalizes processing and generates the final response

**Key Features**:
- Memory-enabled for context retention
- Tool integration via MCP
- RAG capabilities for knowledge retrieval



## Configuration

### Core Settings (`agentflow.toml`)

```toml
[agent_flow]
provider = "openai"           # LLM provider
responsible_ai = true        # Responsible AI checks
error_handler = true         # Enhanced error handling

[orchestration]
mode = "sequential"                # Execution pattern
agents = ["researcher", "analyzer", "synthesizer"]
timeout = 0
failure_threshold = 0
```

### Memory Configuration

```toml
[agent_memory]
enabled = true
provider = "pgvector"
connection = "postgres://user:password@localhost:15432/agentflow?sslmode=disable"
dimensions = 1536

[agent_memory.embedding]
provider = "openai"
model = "text-embedding-3-small"


[agent_memory.rag]
enabled = true
chunk_size = 1500
chunk_overlap = 150
top_k = 8
score_threshold = 0.75
hybrid_search = true


[agent_memory.session]
enabled = true

```

**Memory Features**:
- **Storage**: PostgreSQL with pgvector
- **Embeddings**: text-embedding-3-small via openai
- **RAG**: Enabled with 8 top results
- **Sessions**: Conversation context tracking
- **Search**: Hybrid semantic + keyword search


### Tool Integration (MCP)

```toml
[mcp]
enabled = true
production = true
enable_caching = true
enable_metrics = true
metrics_port = 0

transport = "tcp"  # tcp | stdio | websocket
connection_pool_size = 0
retry_policy = ""

# Example MCP servers
[[mcp.servers]]
name = "filesystem"
command = "npx"
args = ["@modelcontextprotocol/server-filesystem", "/allowed/path"]
enabled = true

[[mcp.servers]]
name = "database"
command = "npx"
args = ["@modelcontextprotocol/server-postgres", "postgresql://localhost/db"]
enabled = true
```

**Available Tools**:
- web_search
- academic_search
- fact_check
- citation_formatter


#### MCP Cache Configuration

```toml
[mcp.cache]
enabled = true
default_ttl_ms = 900000      # 15 minutes
max_size_mb = 100
max_keys = 10000
eviction_policy = "lru"     # lru | lfu | ttl
cleanup_interval_ms = 300000 # 5 minutes
backend = "memory"          # memory | redis | file

[mcp.cache.backend_config]
redis_addr = "localhost:6379"
redis_password = ""
redis_db = "0"
file_path = "./cache"

[mcp.cache.tool_ttls_ms]
# web_search = 300000
# content_fetch = 1800000
```



## Usage Examples

### Basic Usage

```bash
# Simple query
go run . -m "Analyze the current market trends"

# Complex query
go run . -m "Create a comprehensive report on renewable energy adoption, including statistics, challenges, and future projections"
```

### Advanced Usage

```bash
# With custom configuration
go run . -config custom-config.toml -m "Your message"

# With debug logging
LOG_LEVEL=debug go run . -m "Debug this workflow"

# Batch processing
echo "Query 1\nQuery 2\nQuery 3" | go run . -batch
```

### Programmatic Usage

```go
package main

import (
    "context"
    "fmt"
    "testAgent/agents"
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Initialize LLM provider
    llmProvider, err := core.NewProviderFromWorkingDir()
    if err != nil {
        panic(err)
    }
    
    // Create agent
    agent := agents.NewAgent1(llmProvider)
    
    // Create event
    event := core.NewEvent("agent1", core.EventData{
        "message": "Your query here",
    }, nil)
    
    // Execute
    result, err := agent.Run(context.Background(), event, core.NewState())
    if err != nil {
        panic(err)
    }
    
    // Process result as needed for your application
}
```

## Customization

### Quick Customizations

1. **Modify Agent Behavior**: Edit files in `agents/` directory
2. **Change Orchestration**: Update `agentflow.toml` orchestration settings
3. **Add Dependencies**: Extend agent constructors in `main.go`
4. **Custom Input/Output**: Modify processing logic in `main.go`

### Advanced Customizations

For comprehensive customization guidance, see:
- [`docs/CUSTOMIZATION.md`](docs/CUSTOMIZATION.md) - Detailed customization guide
- [`agents/README.md`](agents/README.md) - Agent-specific documentation

### Common Patterns

#### Adding Database Integration

```go
// In main.go
db, err := sql.Open("postgres", "your-connection-string")
if err != nil {
    return fmt.Errorf("failed to connect to database: %w", err)
}

// Pass to agents
agent1 := agents.NewAgent1(llmProvider, db)
```

#### Adding API Integration

```go
// In agents/agent1.go
type Agent1Handler struct {
    llm       agenticgokit.ModelProvider
    apiClient *http.Client
}

func (a *Agent1Handler) callExternalAPI(ctx context.Context, data interface{}) (interface{}, error) {
    // Your API integration logic
    return nil, nil
}
```

#### Custom Output Formatting

```go
// In main.go
func formatResults(results []AgentOutput) {
    for _, result := range results {
        // Format and process result as needed
        logger.Info().Str("agent", result.AgentName).Str("content", result.Content).Msg("Agent result")
    }
}
```

## � Monitoring and Observability

### Logging

The system uses structured logging with different levels:

```bash
# Set log level
export LOG_LEVEL=debug  # debug, info, warn, error

# View logs
go run . -m "test" 2>&1 | jq '.'  # Pretty print JSON logs
```

### Metrics

Metrics are available at `http://localhost:0/metrics` when enabled.

Key metrics:
- Agent execution duration
- Success/failure rates
- Memory operations
- Tool usage statistics


### Health Checks

```bash
# Check system health
curl http://localhost:8080/health

# Check individual components
curl http://localhost:8080/health/agents
curl http://localhost:8080/health/memory
curl http://localhost:8080/health/mcp
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestAgent1 ./agents
```

### Integration Testing

```bash
# Test full workflow
go test -tags=integration ./...

# Test with real LLM provider
INTEGRATION_TEST=true go test ./...
```

### Load Testing

```bash
# Install hey
go install github.com/rakyll/hey@latest

# Load test the system
hey -n 100 -c 10 -m POST -d '{"message":"test"}' http://localhost:8080/process
```

## Deployment

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download && go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/agentflow.toml .
CMD ["./main"]
```

```bash
# Build and run
docker build -t testAgent .
docker run -e OPENAI_API_KEY=$OPENAI_API_KEY testAgent
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: testAgent
spec:
  replicas: 3
  selector:
    matchLabels:
      app: testAgent
  template:
    metadata:
      labels:
        app: testAgent
    spec:
      containers:
      - name: testAgent
        image: your-registry/testAgent:latest
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: api-keys
              key: openai
        ports:
        - containerPort: 8080
```

### Cloud Deployment

#### AWS Lambda

```go
package main

import (
    "context"
    "github.com/aws/aws-lambda-go/lambda"
    "testAgent/agents"
)

type Request struct {
    Message string `json:"message"`
}

type Response struct {
    Result string `json:"result"`
}

func handler(ctx context.Context, req Request) (Response, error) {
    // Initialize your agents
    // Process the request
    // Return response
    return Response{Result: "processed"}, nil
}

func main() {
    lambda.Start(handler)
}
```

## Troubleshooting

### Common Issues

#### 1. LLM Provider Connection Issues

```bash
# Check API key
echo $OPENAI_API_KEY

# Test connection
curl -H "Authorization: Bearer $OPENAI_API_KEY" https://api.openai.com/v1/models
```

#### 2. Memory System Issues

```bash
# Check database connection
psql -h localhost -p 15432 -U user -d agentflow -c "SELECT 1;"


# Check embedding model

```

#### 3. Agent Execution Issues

```bash
# Enable debug logging
LOG_LEVEL=debug go run . -m "test message"

# Check agent registration
grep -n "RegisterAgent" main.go
```

#### 4. Configuration Issues

```bash
# Validate TOML syntax
go run . -validate-config

# Check configuration loading
LOG_LEVEL=debug go run . -m "test" 2>&1 | grep -i config
```

### Performance Issues

1. **Slow Response Times**
   - Check LLM provider latency
   - Review agent processing logic
   - Consider caching strategies

2. **Memory Usage**
   - Monitor Go memory usage
   - Check for memory leaks in agents
   - Optimize data structures

3. **Concurrent Processing**
   - Adjust `max_concurrency` settings
   - Review goroutine usage
   - Check for race conditions

### Getting Help

1. **Check Logs**: Enable debug logging for detailed information
2. **Review Configuration**: Validate all settings in `agentflow.toml`
3. **Test Components**: Test LLM provider, memory, and tools individually
4. **Community Support**: Create an issue in the [AgenticGoKit repository](https://github.com/kunalkushwaha/agenticgokit)

## Resources

### Documentation

- [AgenticGoKit Documentation](https://github.com/kunalkushwaha/agenticgokit)
- [Multi-Agent Patterns](https://github.com/kunalkushwaha/agenticgokit/docs/patterns)
- [Configuration Reference](https://github.com/kunalkushwaha/agenticgokit/docs/config)
- [API Documentation](https://pkg.go.dev/github.com/kunalkushwaha/agenticgokit)

### Examples

- [Example Projects](https://github.com/kunalkushwaha/agenticgokit/examples)
- [Integration Patterns](https://github.com/kunalkushwaha/agenticgokit/docs/integrations)
- [Best Practices](https://github.com/kunalkushwaha/agenticgokit/docs/best-practices)

### Community

- [GitHub Issues](https://github.com/kunalkushwaha/agenticgokit/issues)
- [Discussions](https://github.com/kunalkushwaha/agenticgokit/discussions)
- [Contributing Guide](https://github.com/kunalkushwaha/agenticgokit/CONTRIBUTING.md)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [AgenticGoKit](https://github.com/kunalkushwaha/agenticgokit)
- Powered by OpenAI

- Memory system using PostgreSQL with pgvector

- Tool integration via [Model Context Protocol](https://modelcontextprotocol.io/)

---

**Happy coding!**

For questions or support, please refer to the [documentation](https://github.com/kunalkushwaha/agenticgokit) or create an issue.
