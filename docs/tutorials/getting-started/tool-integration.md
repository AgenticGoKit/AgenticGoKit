# Tool Integration Tutorial (15 minutes)

## Overview

Learn how to connect your agents to external tools and APIs using MCP (Model Context Protocol). You'll set up built-in tools, create custom tools, and optimize tool performance with caching.

## Prerequisites

- Complete the [Memory and RAG](memory-and-rag.md) tutorial
- Basic understanding of APIs and external services
- Optional: Docker for advanced tool setups

## Learning Objectives

By the end of this tutorial, you'll understand:
- How to enable MCP tool integration
- Using built-in tools (web search, file operations)
- Creating custom tools for your specific needs
- Tool caching and performance optimization

## What You'll Build

A tool-enabled agent system that can:
1. **Search the web** for real-time information
2. **Process files** and documents
3. **Call custom APIs** for specialized tasks
4. **Cache results** for better performance

---

## Part 1: Basic Tool Integration (5 minutes)

Start with built-in tools to understand MCP concepts.

### Create a Tool-Enabled Project

```bash
# Create project with MCP tools
agentcli create tool-agent --mcp-enabled --agents 2 \
  --mcp-tools "web_search,summarize"
cd tool-agent
```

### Understanding MCP Configuration

The generated `agentflow.toml` includes MCP settings:

```toml
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

**Important Note**: The `--mcp-tools` flag specifies which tools you want to use, but tools are discovered at runtime from MCP servers, not configured individually in the TOML file. The servers above are examples and are disabled by default - you'll need to enable and configure the servers that provide the tools you want to use.

### Test Basic Tools

```bash
# Set your API key
export OPENAI_API_KEY=your-api-key-here

# Run the tool-enabled system
go run main.go
```

The agents can now use web search and summarization tools automatically when needed.

### View Tool Usage

```bash
# Check tool calls in traces
agentcli trace --verbose <session-id>

# Check MCP server status (run from project directory)
agentcli mcp servers

# List available tools (run from project directory)
agentcli mcp tools

# Note: MCP commands require an active MCP configuration and may need
# the agent system to be running to show live server connections
```

---

## Part 2: Production Tool Setup (5 minutes)

Set up production-ready tools with caching, metrics, and load balancing.

### Create a Production Tool System

```bash
# Create production MCP setup
agentcli create production-tools --mcp-production --with-cache --with-metrics \
  --agents 3 --mcp-tools "web_search,summarize,translate"
cd production-tools
```

### Understanding Production Configuration

The production MCP setup generates the same basic TOML configuration as the standard setup:

```toml
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

**Production Features**: The `--mcp-production`, `--with-cache`, and `--with-metrics` flags enable additional runtime features and optimizations, but the core TOML configuration remains the same. Production features are handled at the application level rather than through additional TOML sections.

### Start Production Services

```bash
# Start with Docker Compose (if generated)
docker-compose up -d

# Or run directly
export OPENAI_API_KEY=your-api-key-here
go run main.go
```

### Monitor Tool Performance

```bash
# Check cache statistics (run from project directory)
agentcli cache stats

# View metrics (if enabled and running)
curl http://localhost:8080/metrics

# Monitor MCP servers (run from project directory)
agentcli mcp health

# Note: These commands require the agent system to be running
# and may need active MCP connections to show meaningful data
```

---

## Part 3: Custom Tool Development (5 minutes)

Create custom tools for your specific use cases.

### Create a Custom Tool Project

```bash
# Create project with custom tool setup
agentcli create custom-tools --mcp-enabled --agents 2
cd custom-tools
```

### Add a Custom Weather Tool

Create `tools/weather.go`:

```go
package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "time"
)

// WeatherTool provides weather information
type WeatherTool struct {
    apiKey     string
    httpClient *http.Client
}

func NewWeatherTool(apiKey string) *WeatherTool {
    return &WeatherTool{
        apiKey: apiKey,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

func (w *WeatherTool) Name() string {
    return "weather"
}

func (w *WeatherTool) Description() string {
    return "Gets current weather information for a specified location"
}

func (w *WeatherTool) ParameterSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "location": map[string]interface{}{
                "type":        "string",
                "description": "The city, zip code, or coordinates",
            },
        },
        "required": []string{"location"},
    }
}

func (w *WeatherTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    location, ok := params["location"].(string)
    if !ok || location == "" {
        return nil, fmt.Errorf("location parameter is required")
    }
    
    // Build API URL (using a free weather API)
    apiURL := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", 
        w.apiKey, url.QueryEscape(location))
    
    req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    resp, err := w.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("weather API request failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("weather API returned status %d", resp.StatusCode)
    }
    
    var weatherData map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
        return nil, fmt.Errorf("failed to parse weather data: %w", err)
    }
    
    return weatherData, nil
}
```

### Register the Custom Tool

Update `main.go` to include your custom tool:

```go
// Add to your main.go
import "your-project/tools"

func main() {
    // ... existing code ...
    
    // Create tool manager
    toolManager := core.NewToolManager()
    
    // Register custom weather tool
    weatherTool := tools.NewWeatherTool(os.Getenv("WEATHER_API_KEY"))
    toolManager.RegisterTool(weatherTool)
    
    // Register with agents
    for _, agent := range agents {
        agent.SetToolManager(toolManager)
    }
    
    // ... rest of code ...
}
```

### Test Custom Tools

```bash
# Set API keys
export OPENAI_API_KEY=your-openai-key
export WEATHER_API_KEY=your-weather-api-key

# Run with custom tools
go run main.go
```

Your agents can now use the custom weather tool alongside built-in tools.

---

## Tool Categories and Use Cases

### Built-in Tools

| Tool | Description | Use Case |
|------|-------------|----------|
| **web_search** | Search the internet | Real-time information, research |
| **summarize** | Summarize text content | Document processing, content analysis |
| **translate** | Translate between languages | Multi-language support |
| **file_operations** | Read/write files | Document processing, data import |

### Custom Tool Examples

| Tool Type | Example | Implementation |
|-----------|---------|----------------|
| **API Integration** | Weather, Stock prices | HTTP client with JSON parsing |
| **Database Access** | User lookup, Data queries | SQL client with connection pooling |
| **File Processing** | PDF parsing, Image analysis | Specialized libraries |
| **External Services** | Email sending, SMS | Service-specific SDKs |

## Tool Performance Optimization

### Caching Configuration

Caching is configured in the main MCP section:

```toml
[mcp]
enabled = true
enable_discovery = true
connection_timeout = 5000
max_retries = 3
retry_delay = 1000
enable_caching = true
cache_timeout = 300000          # 5 minutes default cache timeout
max_connections = 10
```

**Note**: Individual tool cache settings are handled at runtime through the MCP manager, not through separate TOML sections. The `cache_timeout` setting applies to all tools by default, and specific tool caching behavior is managed programmatically.

### Load Balancing and Circuit Breaker Protection

**Note**: Load balancing and circuit breaker features are handled at the application runtime level when using `--mcp-production` flag. The TOML configuration remains the same:

```toml
[mcp]
enabled = true
enable_discovery = true
connection_timeout = 5000       # Connection timeout in milliseconds
max_retries = 3                 # Maximum retry attempts
retry_delay = 1000              # Delay between retries in milliseconds
enable_caching = true
cache_timeout = 300000
max_connections = 10            # Maximum concurrent connections
```

Production features like load balancing, circuit breakers, and advanced retry policies are configured programmatically when you use the `--mcp-production` flag during project creation.

## Advanced Tool Patterns

### Tool Composition

```go
// Combine multiple tools for complex operations
type ResearchTool struct {
    webSearch   Tool
    summarizer  Tool
    translator  Tool
}

func (r *ResearchTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // 1. Search for information
    searchResults, err := r.webSearch.Execute(ctx, params)
    if err != nil {
        return nil, err
    }
    
    // 2. Summarize results
    summaryParams := map[string]interface{}{
        "text": searchResults,
    }
    summary, err := r.summarizer.Execute(ctx, summaryParams)
    if err != nil {
        return nil, err
    }
    
    // 3. Translate if needed
    if targetLang, ok := params["language"]; ok {
        translateParams := map[string]interface{}{
            "text":        summary,
            "target_lang": targetLang,
        }
        return r.translator.Execute(ctx, translateParams)
    }
    
    return summary, nil
}
```

### Async Tool Execution

```go
// Execute multiple tools concurrently
func (a *Agent) executeToolsConcurrently(ctx context.Context, tools []ToolCall) (map[string]interface{}, error) {
    results := make(map[string]interface{})
    var wg sync.WaitGroup
    var mu sync.Mutex
    
    for _, toolCall := range tools {
        wg.Add(1)
        go func(tc ToolCall) {
            defer wg.Done()
            
            result, err := a.toolManager.ExecuteTool(ctx, tc.Name, tc.Params)
            
            mu.Lock()
            if err != nil {
                results[tc.Name] = fmt.Sprintf("Error: %v", err)
            } else {
                results[tc.Name] = result
            }
            mu.Unlock()
        }(toolCall)
    }
    
    wg.Wait()
    return results, nil
}
```

## Troubleshooting

### Common Issues

**Tool not found:**
```bash
# Check tool registration
agentcli mcp tools

# Verify MCP server status
agentcli mcp servers

# Check configuration
cat agentflow.toml | grep -A 10 "\[mcp\]"
```

**Tool execution timeout:**
```toml
# Increase timeout in configuration
[mcp]
connection_timeout = 60000  # Increase to 60 seconds (in milliseconds)
```

**Cache not working:**
```bash
# Check cache statistics
agentcli cache stats

# Clear cache if needed
agentcli cache clear --all
```

### Performance Issues

**Slow tool execution:**
- Enable caching for frequently used tools
- Use connection pooling for database tools
- Implement circuit breakers for unreliable services

**High resource usage:**
- Limit concurrent tool executions
- Use tool result caching
- Monitor with metrics

## Tool Security Best Practices

### Input Validation

```go
func (t *CustomTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // Always validate inputs
    input, ok := params["input"].(string)
    if !ok {
        return nil, fmt.Errorf("input parameter must be a string")
    }
    
    // Sanitize inputs
    input = sanitizeInput(input)
    
    // Validate against schema
    if err := validateInput(input); err != nil {
        return nil, fmt.Errorf("invalid input: %w", err)
    }
    
    // ... tool logic ...
}
```

### API Key Management

```go
// Use environment variables for sensitive data
func NewAPITool() *APITool {
    return &APITool{
        apiKey: os.Getenv("API_KEY"),  // Never hardcode keys
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}
```

## Next Steps

Now that your agents can use tools:

1. **Go Production**: Learn [Production Deployment](production-deployment.md) for scaling
2. **Advanced Patterns**: Explore [Advanced Patterns](../advanced-patterns/) for complex workflows
3. **Custom MCP Servers**: Build [MCP Tool Development](../tools/creating-custom-tools.md)

## Key Takeaways

- **MCP Integration**: Standardized way to connect agents with external tools
- **Built-in Tools**: Ready-to-use tools for common tasks
- **Custom Tools**: Easy to create for specific needs
- **Performance**: Caching and load balancing for production use
- **Security**: Always validate inputs and manage credentials properly

## Further Reading

- [MCP Fundamentals](../tools/mcp-fundamentals.md) - Deep dive into MCP concepts
- [Creating Custom Tools](../tools/creating-custom-tools.md) - Advanced tool development
- [Tool Integration Patterns](../tools/tool-integration-patterns.md) - Advanced patterns