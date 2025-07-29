# basic-test

**An intelligent multi-agent system powered by AgenticGoKit**

basic-test is a sophisticated multi-agent workflow system that leverages multiple AI agents working sequentially to process and respond to user queries.

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or later
- OpenAI API key



### Installation

1. **Clone and setup**:
   ```bash
   git clone <your-repository>
   cd basic-test
   go mod tidy
   ```

2. **Configure environment**:
   ```bash
   export OPENAI_API_KEY="your-api-key"
   ```





### Running the System

```bash
# Interactive mode
go run . 

# Command line mode
go run . -m "Your message here"

# With debug logging
LOG_LEVEL=debug go run . -m "Your message"
```

## 🏗️ Architecture

### System Overview

basic-test implements a sequential multi-agent architecture with 2 specialized agents:

```
User Input → Agent1 → Agent2 → ... → Final Response
```

### Project Structure

```
basic-test/
├── 📁 agents/                 # Agent implementations
│   ├── agent1.go           # Agent1 agent
│   ├── agent2.go           # Agent2 agent
│   └── README.md           # Agent documentation
├── 📁 internal/               # Internal packages
│   ├── config/               # Configuration utilities
│   └── handlers/             # Shared handler utilities
├── 📁 docs/                  # Documentation
│   └── CUSTOMIZATION.md      # Customization guide
├── 📄 main.go                # Application entry point
├── 📄 agentflow.toml         # System configuration
├── 📄 go.mod                 # Go module definition
└── 📄 README.md              # This file
```

### Agent Responsibilities


#### Agent1 (`agents/agent1.go`)

**Purpose**: Processes tasks in sequence as part of a processing pipeline

**Role**: Processes initial user input and prepares data for downstream agents

**Key Features**:
- ❌ Memory not enabled
- ❌ No tool integration
- ❌ No RAG capabilities


#### Agent2 (`agents/agent2.go`)

**Purpose**: Processes tasks in sequence as part of a processing pipeline

**Role**: Finalizes processing and generates the final response

**Key Features**:
- ❌ Memory not enabled
- ❌ No tool integration
- ❌ No RAG capabilities



## ⚙️ Configuration

### Core Settings (`agentflow.toml`)

```toml
[agent_flow]
provider = "openai"           # LLM provider
responsible_ai = true        # Responsible AI checks
error_handler = true         # Enhanced error handling

[orchestration]
mode = "sequential"                # Execution pattern
agents = ["agent1", "agent2"]
timeout = 0
failure_threshold = 0
```





## 🎯 Usage Examples

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
    "basic-test/agents"
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
    
    fmt.Println("Result:", result)
}
```

## 🛠️ Customization

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
    log.Fatal(err)
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
        fmt.Printf("🤖 %s: %s\n", result.AgentName, result.Content)
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

To enable metrics, set `enable_metrics = true` in your MCP configuration.

### Health Checks

```bash
# Check system health
curl http://localhost:8080/health

# Check individual components
curl http://localhost:8080/health/agents
curl http://localhost:8080/health/memory
curl http://localhost:8080/health/mcp
```

## 🧪 Testing

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

## 🚀 Deployment

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
docker build -t basic-test .
docker run -e OPENAI_API_KEY=$OPENAI_API_KEY basic-test
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: basic-test
spec:
  replicas: 3
  selector:
    matchLabels:
      app: basic-test
  template:
    metadata:
      labels:
        app: basic-test
    spec:
      containers:
      - name: basic-test
        image: your-registry/basic-test:latest
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
    "basic-test/agents"
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

## 🔧 Troubleshooting

### Common Issues

#### 1. LLM Provider Connection Issues

```bash
# Check API key
echo $OPENAI_API_KEY

# Test connection
curl -H "Authorization: Bearer $OPENAI_API_KEY" https://api.openai.com/v1/models
```

#### 2. Memory System Issues



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

## 📚 Resources

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

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [AgenticGoKit](https://github.com/kunalkushwaha/agenticgokit)
- Powered by OpenAI



---

**Happy coding! 🚀**

For questions or support, please refer to the [documentation](https://github.com/kunalkushwaha/agenticgokit) or create an issue.
