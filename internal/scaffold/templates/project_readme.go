package templates

const ProjectReadmeTemplate = `# {{.Config.Name}}

**An intelligent multi-agent system powered by AgenticGoKit**

{{.Config.Name}} is a sophisticated multi-agent workflow system that leverages multiple AI agents working {{if eq .Config.OrchestrationMode "sequential"}}sequentially{{else if eq .Config.OrchestrationMode "collaborative"}}collaboratively{{else if eq .Config.OrchestrationMode "loop"}}iteratively{{else}}in coordination{{end}} to process and respond to user queries.

## Quick Start

### Prerequisites

- Go 1.21 or later
- {{if eq .Config.Provider "openai"}}OpenAI API key{{else if eq .Config.Provider "azure"}}Azure OpenAI credentials{{else if eq .Config.Provider "ollama"}}Ollama running locally{{else}}LLM provider credentials{{end}}
{{if .Config.MemoryEnabled}}{{if eq .Config.MemoryProvider "pgvector"}}
- PostgreSQL with pgvector extension{{else if eq .Config.MemoryProvider "weaviate"}}
- Weaviate instance{{end}}{{end}}
{{if .Config.MCPEnabled}}
- Node.js (for MCP tools){{end}}

### Installation

1. **Clone and setup**:
   ` + "```bash" + `
   git clone <your-repository>
   cd {{.Config.Name}}
   go mod tidy
   ` + "```" + `

2. **Configure environment**:
   {{if eq .Config.Provider "openai"}}` + "```bash" + `
   export OPENAI_API_KEY="your-api-key"
   ` + "```" + `{{else if eq .Config.Provider "azure"}}` + "```bash" + `
   export AZURE_OPENAI_API_KEY="your-api-key"
   export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com/"
   export AZURE_OPENAI_DEPLOYMENT="your-deployment-name"
   ` + "```" + `{{else if eq .Config.Provider "ollama"}}` + "```bash" + `
   # Start Ollama
   ollama serve
   
   # Pull required model
   ollama pull llama2
   ` + "```" + `{{end}}

{{if .Config.MemoryEnabled}}{{if eq .Config.MemoryProvider "pgvector"}}3. **Setup PostgreSQL with pgvector**:
   ` + "```bash" + `
   # Start database
   docker compose up -d
   
   # Run setup script
   ./setup.sh  # or setup.bat on Windows
   ` + "```" + `{{else if eq .Config.MemoryProvider "weaviate"}}3. **Setup Weaviate**:
   ` + "```bash" + `
   # Start Weaviate
   docker compose up -d
   ` + "```" + `{{end}}{{end}}

{{if .Config.MCPEnabled}}{{if .Config.MemoryEnabled}}4{{else}}3{{end}}. **Setup MCP tools** (optional):
   ` + "```bash" + `
   # Install MCP servers
   npm install -g @modelcontextprotocol/server-filesystem
   # Add other MCP servers as needed
   ` + "```" + `{{end}}

### Running the System

` + "```bash" + `
# Validate configuration first
agentcli validate agentflow.toml

# Interactive mode
go run . 

# Command line mode
go run . -m "Your message here"

# With debug logging
LOG_LEVEL=debug go run . -m "Your message"
` + "```" + `

## Configuration-Driven Architecture

This project uses **AgentFlow's configuration-driven architecture**:

- **No hardcoded agents**: All agents defined in ` + "`agentflow.toml`" + `
- **Flexible configuration**: Change behavior without code changes  
- **Hot reload support**: Update config without restarting
- **Environment-specific**: Different settings per environment
- **Built-in validation**: Comprehensive validation with helpful errors

### Key Configuration Files

- **` + "`agentflow.toml`" + `**: Main configuration (agents, LLM, orchestration)
- **` + "`agents/`" + `**: Reference implementations (optional)
- **Environment variables**: Sensitive data (API keys)

### Configuration Management

` + "```bash" + `
# Validate configuration
agentcli validate agentflow.toml

# Generate new configuration from template
agentcli config generate research-assistant my-project

# Get detailed validation report
agentcli validate --detailed agentflow.toml

# Export configuration schema
agentcli config schema --generate
` + "```" + `

### Configuration Example

` + "```toml" + `
# Global LLM settings
[llm]
provider = "{{.Config.Provider}}"
model = "gpt-4"
temperature = 0.7

# Agent definitions
{{range $i, $agent := .Agents}}[agents.{{$agent.Name}}]
role = "{{$agent.Name}}"
description = "{{$agent.Purpose}}"
system_prompt = "You are {{$agent.DisplayName}}, {{$agent.Purpose}}"
capabilities = ["general_assistance", "processing"]
enabled = true

# Agent-specific LLM settings
[agents.{{$agent.Name}}.llm]
temperature = 0.7
max_tokens = 2000

{{end}}# Orchestration
[orchestration]
mode = "{{.Config.OrchestrationMode}}"
{{if eq .Config.OrchestrationMode "sequential"}}agents = [{{range $i, $agent := .Agents}}{{if $i}}, {{end}}"{{$agent.Name}}"{{end}}]{{end}}
` + "```" + `

## Architecture

### System Overview

{{.Config.Name}} implements a {{.Config.OrchestrationMode}} multi-agent architecture with {{len .Agents}} specialized agents:

` + "```" + `
{{if eq .Config.OrchestrationMode "sequential"}}User Input -> Agent1 -> Agent2 -> ... -> Final Response{{else if eq .Config.OrchestrationMode "collaborative"}}User Input -> [Agent1, Agent2, Agent3] -> Aggregated Response{{else if eq .Config.OrchestrationMode "loop"}}User Input -> Agent -> Agent -> ... ({{.Config.MaxIterations}} iterations) -> Final Response{{else}}User Input -> Router -> Appropriate Agent -> Response{{end}}
` + "```" + `

### Project Structure

` + "```" + `
{{.Config.Name}}/
|-- agents/                 # Agent implementations
{{range .Agents}}|   |-- {{.FileName}}           # {{.DisplayName}} agent
{{end}}|   ` + "`" + `-- README.md           # Agent documentation
|-- internal/               # Internal packages
|   |-- config/             # Configuration utilities
|   ` + "`" + `-- handlers/           # Shared handler utilities
|-- docs/                   # Documentation
|   ` + "`" + `-- CUSTOMIZATION.md    # Customization guide
|-- main.go                 # Application entry point
|-- agentflow.toml          # Main configuration file
{{if .Config.MemoryEnabled}}|-- docker-compose.yml        # Database services
{{if eq .Config.MemoryProvider "pgvector"}}|-- setup.sh                 # Database setup script{{end}}{{end}}
|-- agentflow.toml          # System configuration
|-- go.mod                  # Go module definition
{{if .Config.MemoryEnabled}}{{if or (eq .Config.MemoryProvider "pgvector") (eq .Config.MemoryProvider "weaviate")}}|-- docker-compose.yml        # Database services
|-- setup.sh                # Database setup script
{{end}}{{end}}` + "`" + `-- README.md               # This file
` + "```" + `

### Agent Responsibilities

{{range $i, $agent := .Agents}}
#### {{$agent.DisplayName}} (` + "`agents/{{$agent.FileName}}`" + `)

**Purpose**: {{$agent.Purpose}}

**Role**: {{if eq $.Config.OrchestrationMode "sequential"}}{{if $agent.IsFirstAgent}}Processes initial user input and prepares data for downstream agents{{else if $agent.IsLastAgent}}Finalizes processing and generates the final response{{else}}Processes input from previous agents and passes refined data forward{{end}}{{else if eq $.Config.OrchestrationMode "collaborative"}}Works in parallel with other agents to process user input from different perspectives{{else if eq $.Config.OrchestrationMode "loop"}}Iteratively processes and refines the input over multiple cycles{{else}}Handles specific types of requests based on routing logic{{end}}

**Key Features**:
- {{if $.Config.MemoryEnabled}}Memory-enabled for context retention{{else}}Memory not enabled{{end}}
- {{if $.Config.MCPEnabled}}Tool integration via MCP{{else}}No tool integration{{end}}
- {{if $.Config.RAGEnabled}}RAG capabilities for knowledge retrieval{{else}}No RAG capabilities{{end}}

{{end}}

## Configuration

### Core Settings (` + "`agentflow.toml`" + `)

` + "```toml" + `
[agent_flow]
provider = "{{.Config.Provider}}"           # LLM provider
{{if .Config.ResponsibleAI}}responsible_ai = true        # Responsible AI checks{{end}}
{{if .Config.ErrorHandler}}error_handler = true         # Enhanced error handling{{end}}

[orchestration]
mode = "{{.Config.OrchestrationMode}}"                # Execution pattern
{{if eq .Config.OrchestrationMode "sequential"}}agents = [{{range $i, $agent := .Agents}}{{if gt $i 0}}, {{end}}"{{$agent.Name}}"{{end}}]
{{else if eq .Config.OrchestrationMode "collaborative"}}agents = [{{range $i, $agent := .Agents}}{{if gt $i 0}}, {{end}}"{{$agent.Name}}"{{end}}]
max_concurrency = {{.Config.MaxConcurrency}}
{{else if eq .Config.OrchestrationMode "loop"}}agent = "{{(index .Agents 0).Name}}"
max_iterations = {{.Config.MaxIterations}}
{{end}}timeout = {{.Config.OrchestrationTimeout}}
failure_threshold = {{.Config.FailureThreshold}}
` + "```" + `

{{if .Config.MemoryEnabled}}### Memory Configuration

` + "```toml" + `
[agent_memory]
enabled = true
provider = "{{.Config.MemoryProvider}}"
{{if eq .Config.MemoryProvider "pgvector"}}connection = "postgres://user:password@localhost:15432/agentflow?sslmode=disable"
{{else if eq .Config.MemoryProvider "weaviate"}}connection = "http://localhost:8080"
{{end}}dimensions = {{.Config.EmbeddingDimensions}}

[agent_memory.embedding]
provider = "{{.Config.EmbeddingProvider}}"
model = "{{.Config.EmbeddingModel}}"
{{if eq .Config.EmbeddingProvider "ollama"}}base_url = "http://localhost:11434"{{end}}

{{if .Config.RAGEnabled}}[agent_memory.rag]
enabled = true
chunk_size = {{.Config.RAGChunkSize}}
chunk_overlap = {{.Config.RAGOverlap}}
top_k = {{.Config.RAGTopK}}
score_threshold = {{.Config.RAGScoreThreshold}}
{{if .Config.HybridSearch}}hybrid_search = true{{end}}
{{end}}

{{if .Config.SessionMemory}}[agent_memory.session]
enabled = true
{{end}}
` + "```" + `

**Memory Features**:
- **Storage**: {{if eq .Config.MemoryProvider "memory"}}In-memory (non-persistent){{else if eq .Config.MemoryProvider "pgvector"}}PostgreSQL with pgvector{{else if eq .Config.MemoryProvider "weaviate"}}Weaviate vector database{{end}}
- **Embeddings**: {{.Config.EmbeddingModel}} via {{.Config.EmbeddingProvider}}
{{if .Config.RAGEnabled}}- **RAG**: Enabled with {{.Config.RAGTopK}} top results{{end}}
{{if .Config.SessionMemory}}- **Sessions**: Conversation context tracking{{end}}
{{if .Config.HybridSearch}}- **Search**: Hybrid semantic + keyword search{{end}}
{{end}}

{{if .Config.MCPEnabled}}### Tool Integration (MCP)

` + "```toml" + `
[mcp]
enabled = true
{{if .Config.MCPProduction}}production = true{{end}}
{{if .Config.WithCache}}enable_caching = true{{end}}
{{if .Config.WithMetrics}}enable_metrics = true
metrics_port = {{.Config.MetricsPort}}{{end}}
{{if .Config.WithLoadBalancer}}enable_load_balancer = true{{end}}
transport = "{{if .Config.MCPTransport}}{{.Config.MCPTransport}}{{else}}tcp{{end}}"  # tcp | stdio | websocket
connection_pool_size = {{.Config.ConnectionPoolSize}}
retry_policy = "{{.Config.RetryPolicy}}"

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
` + "```" + `

**Available Tools**:
{{range .Config.MCPTools}}- {{.}}
{{end}}

{{if .Config.WithCache}}#### MCP Cache Configuration

` + "```toml" + `
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
` + "```" + `
{{end}}
{{end}}

## Usage Examples

### Basic Usage

` + "```bash" + `
# Simple query
go run . -m "Analyze the current market trends"

# Complex query
go run . -m "Create a comprehensive report on renewable energy adoption, including statistics, challenges, and future projections"
` + "```" + `

### Advanced Usage

` + "```bash" + `
# With custom configuration
go run . -config custom-config.toml -m "Your message"

# With debug logging
LOG_LEVEL=debug go run . -m "Debug this workflow"

# Batch processing
echo "Query 1\nQuery 2\nQuery 3" | go run . -batch
` + "```" + `

### Programmatic Usage

` + "```go" + `
package main

import (
    "context"
    "fmt"
    "{{.Config.Name}}/agents"
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
` + "```" + `

## Customization

### Quick Customizations

1. **Modify Agent Behavior**: Edit files in ` + "`agents/`" + ` directory
2. **Change Orchestration**: Update ` + "`agentflow.toml`" + ` orchestration settings
3. **Add Dependencies**: Extend agent constructors in ` + "`main.go`" + `
4. **Custom Input/Output**: Modify processing logic in ` + "`main.go`" + `

### Advanced Customizations

For comprehensive customization guidance, see:
- [` + "`docs/CUSTOMIZATION.md`" + `](docs/CUSTOMIZATION.md) - Detailed customization guide
- [` + "`agents/README.md`" + `](agents/README.md) - Agent-specific documentation

### Common Patterns

#### Adding Database Integration

` + "```go" + `
// In main.go
db, err := sql.Open("postgres", "your-connection-string")
if err != nil {
    return fmt.Errorf("failed to connect to database: %w", err)
}

// Pass to agents
agent1 := agents.NewAgent1(llmProvider, db)
` + "```" + `

#### Adding API Integration

` + "```go" + `
// In agents/agent1.go
type Agent1Handler struct {
    llm       agenticgokit.ModelProvider
    apiClient *http.Client
}

func (a *Agent1Handler) callExternalAPI(ctx context.Context, data interface{}) (interface{}, error) {
    // Your API integration logic
    return nil, nil
}
` + "```" + `

#### Custom Output Formatting

` + "```go" + `
// In main.go
func formatResults(results []AgentOutput) {
    for _, result := range results {
        // Format and process result as needed
        logger.Info().Str("agent", result.AgentName).Str("content", result.Content).Msg("Agent result")
    }
}
` + "```" + `

## � Monitoring and Observability

### Logging

The system uses structured logging with different levels:

` + "```bash" + `
# Set log level
export LOG_LEVEL=debug  # debug, info, warn, error

# View logs
go run . -m "test" 2>&1 | jq '.'  # Pretty print JSON logs
` + "```" + `

### Metrics

{{if .Config.WithMetrics}}Metrics are available at ` + "`http://localhost:{{.Config.MetricsPort}}/metrics`" + ` when enabled.

Key metrics:
- Agent execution duration
- Success/failure rates
- Memory operations
- Tool usage statistics
{{else}}To enable metrics, set ` + "`enable_metrics = true`" + ` in your MCP configuration.{{end}}

### Health Checks

` + "```bash" + `
# Check system health
curl http://localhost:8080/health

# Check individual components
curl http://localhost:8080/health/agents
curl http://localhost:8080/health/memory
curl http://localhost:8080/health/mcp
` + "```" + `

## Testing

### Running Tests

` + "```bash" + `
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestAgent1 ./agents
` + "```" + `

### Integration Testing

` + "```bash" + `
# Test full workflow
go test -tags=integration ./...

# Test with real LLM provider
INTEGRATION_TEST=true go test ./...
` + "```" + `

### Load Testing

` + "```bash" + `
# Install hey
go install github.com/rakyll/hey@latest

# Load test the system
hey -n 100 -c 10 -m POST -d '{"message":"test"}' http://localhost:8080/process
` + "```" + `

## Deployment

### Docker

` + "```dockerfile" + `
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
` + "```" + `

` + "```bash" + `
# Build and run
docker build -t {{.Config.Name}} .
docker run -e OPENAI_API_KEY=$OPENAI_API_KEY {{.Config.Name}}
` + "```" + `

### Kubernetes

` + "```yaml" + `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Config.Name}}
spec:
  replicas: 3
  selector:
    matchLabels:
      app: {{.Config.Name}}
  template:
    metadata:
      labels:
        app: {{.Config.Name}}
    spec:
      containers:
      - name: {{.Config.Name}}
        image: your-registry/{{.Config.Name}}:latest
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: api-keys
              key: openai
        ports:
        - containerPort: 8080
` + "```" + `

### Cloud Deployment

#### AWS Lambda

` + "```go" + `
package main

import (
    "context"
    "github.com/aws/aws-lambda-go/lambda"
    "{{.Config.Name}}/agents"
)

type Request struct {
    Message string ` + "`json:\"message\"`" + `
}

type Response struct {
    Result string ` + "`json:\"result\"`" + `
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
` + "```" + `

## Troubleshooting

### Common Issues

#### 1. LLM Provider Connection Issues

` + "```bash" + `
# Check API key
echo $OPENAI_API_KEY

# Test connection
curl -H "Authorization: Bearer $OPENAI_API_KEY" https://api.openai.com/v1/models
` + "```" + `

#### 2. Memory System Issues

{{if .Config.MemoryEnabled}}` + "```bash" + `
# Check database connection
{{if eq .Config.MemoryProvider "pgvector"}}psql -h localhost -p 15432 -U user -d agentflow -c "SELECT 1;"
{{else if eq .Config.MemoryProvider "weaviate"}}curl http://localhost:8080/v1/meta
{{end}}

# Check embedding model
{{if eq .Config.EmbeddingProvider "ollama"}}curl http://localhost:11434/api/tags
{{end}}
` + "```" + `{{end}}

#### 3. Agent Execution Issues

` + "```bash" + `
# Enable debug logging
LOG_LEVEL=debug go run . -m "test message"

# Check agent registration
grep -n "RegisterAgent" main.go
` + "```" + `

#### 4. Configuration Issues

` + "```bash" + `
# Validate TOML syntax
go run . -validate-config

# Check configuration loading
LOG_LEVEL=debug go run . -m "test" 2>&1 | grep -i config
` + "```" + `

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
   - Adjust ` + "`max_concurrency`" + ` settings
   - Review goroutine usage
   - Check for race conditions

### Getting Help

1. **Check Logs**: Enable debug logging for detailed information
2. **Review Configuration**: Validate all settings in ` + "`agentflow.toml`" + `
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
- Powered by {{if eq .Config.Provider "openai"}}OpenAI{{else if eq .Config.Provider "azure"}}Azure OpenAI{{else if eq .Config.Provider "ollama"}}Ollama{{else}}{{.Config.Provider}}{{end}}
{{if .Config.MemoryEnabled}}
- Memory system using {{if eq .Config.MemoryProvider "pgvector"}}PostgreSQL with pgvector{{else if eq .Config.MemoryProvider "weaviate"}}Weaviate{{else}}in-memory storage{{end}}{{end}}
{{if .Config.MCPEnabled}}
- Tool integration via [Model Context Protocol](https://modelcontextprotocol.io/){{end}}

---

**Happy coding!**

For questions or support, please refer to the [documentation](https://github.com/kunalkushwaha/agenticgokit) or create an issue.
`
