# MCP Tools

**Dynamic Tool Discovery and Execution with Model Context Protocol**

AgenticGoKit uses the Model Context Protocol (MCP) to provide agents with dynamic tool discovery and execution capabilities. This guide covers everything from basic tool usage to building custom MCP servers.

## Overview

The MCP integration in AgenticGoKit provides:
- **Dynamic Discovery**: Tools are discovered at runtime, not hard-coded
- **Schema-Based**: Tools provide their own descriptions and parameters
- **LLM-Driven**: The LLM decides which tools to use based on context
- **Extensible**: Add new tools by connecting MCP servers

## Quick Start (5 minutes)

### 1. Basic MCP Configuration

**agentflow.toml:**
```toml
[mcp]
enabled = true

# Web search capabilities
[mcp.servers.search]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-web-search"]
transport = "stdio"

# Docker container management
[mcp.servers.docker]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-docker"]
transport = "stdio"

# File system operations
[mcp.servers.filesystem]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-filesystem"]
transport = "stdio"
```

### 2. Tool-Enabled Agent

```go
package main

import (
    "context"
    "fmt"
    agentflow "github.com/kunalkushwaha/agenticgokit/core"
)

type ToolEnabledAgent struct {
    llm        agentflow.ModelProvider
    mcpManager agentflow.MCPManager
}

func NewToolEnabledAgent(llm agentflow.ModelProvider, mcp agentflow.MCPManager) *ToolEnabledAgent {
    return &ToolEnabledAgent{
        llm:        llm,
        mcpManager: mcp,
    }
}

func (a *ToolEnabledAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    message := event.GetData()["message"]
    
    // Build system prompt with tool awareness
    systemPrompt := `You are a helpful assistant with access to various tools. 
    Use tools when they can provide current, specific, or actionable information.
    
    When using tools, format calls exactly like this:
    <tool_call>
    {"name": "tool_name", "args": {"param": "value"}}
    </tool_call>`
    
    // Get available tools and add to prompt
    toolPrompt := agentflow.FormatToolsForPrompt(ctx, a.mcpManager)
    fullPrompt := fmt.Sprintf("%s\n\n%s\n\nUser: %s", systemPrompt, toolPrompt, message)
    
    // Get initial LLM response
    response, err := a.llm.Generate(ctx, fullPrompt)
    if err != nil {
        return agentflow.AgentResult{}, fmt.Errorf("LLM generation failed: %w", err)
    }
    
    // Execute any tool calls found in response
    toolResults := agentflow.ParseAndExecuteToolCalls(ctx, a.mcpManager, response)
    
    if len(toolResults) > 0 {
        // Synthesize tool results with original response
        synthesisPrompt := fmt.Sprintf(`Original response: %s
        
Tool execution results: %v

Please provide a comprehensive final answer that incorporates the tool results.`, response, toolResults)
        
        finalResponse, err := a.llm.Generate(ctx, synthesisPrompt)
        if err != nil {
            finalResponse = response // Fallback to original response
        }
        
        state.Set("tools_used", true)
        state.Set("tool_results", toolResults)
        return agentflow.AgentResult{Result: finalResponse, State: state}, nil
    }
    
    return agentflow.AgentResult{Result: response, State: state}, nil
}
```

## Available MCP Servers

### Development & System Tools

```toml
# Docker management
[mcp.servers.docker]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-docker"]
transport = "stdio"

# File system operations
[mcp.servers.filesystem]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-filesystem"]
transport = "stdio"

# GitHub integration
[mcp.servers.github]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-github"]
transport = "stdio"
env = { "GITHUB_TOKEN" = "${GITHUB_TOKEN}" }
```

### Web & Search Tools

```toml
# Web search
[mcp.servers.search]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-web-search"]
transport = "stdio"

# URL content fetching
[mcp.servers.fetch]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-fetch"]
transport = "stdio"

# Brave search API
[mcp.servers.brave]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-brave-search"]
transport = "stdio"
env = { "BRAVE_API_KEY" = "${BRAVE_API_KEY}" }
```

### Database Tools

```toml
# PostgreSQL
[mcp.servers.postgres]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-postgres"]
transport = "stdio"
env = { "DATABASE_URL" = "${DATABASE_URL}" }

# SQLite
[mcp.servers.sqlite]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-sqlite"]
transport = "stdio"

# MongoDB
[mcp.servers.mongodb]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-mongodb"]
transport = "stdio"
env = { "MONGODB_URI" = "${MONGODB_URI}" }
```

## Production Configuration

### With Caching

```toml
[mcp]
enabled = true
cache_enabled = true
cache_ttl = "5m"
connection_timeout = "30s"
max_retries = 3

[mcp.cache]
type = "memory"
max_size = 1000

# Production-ready servers
[mcp.servers.search]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-web-search"]
transport = "stdio"
env = { "SEARCH_API_KEY" = "${SEARCH_API_KEY}" }

[mcp.servers.database]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-postgres"]
transport = "stdio"
env = { "DATABASE_URL" = "${DATABASE_URL}" }
```

## Tool Usage Patterns

### Information Gathering Pattern

```go
func (a *ResearchAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    query := event.GetData()["message"]
    
    // System prompt optimized for research
    systemPrompt := `You are a research agent. For any query:

1. First, search for current information using the search tool
2. If specific URLs are mentioned or found, fetch their content
3. Gather multiple perspectives and sources
4. Organize findings in a structured way

Always use tools for factual, current information rather than relying on training data.`
    
    toolPrompt := agentflow.FormatToolsForPrompt(ctx, a.mcpManager)
    prompt := fmt.Sprintf("%s\n%s\nResearch: %s", systemPrompt, toolPrompt, query)
    
    response, err := a.llm.Generate(ctx, prompt)
    if err != nil {
        return agentflow.AgentResult{}, err
    }
    
    // Execute research tools
    toolResults := agentflow.ParseAndExecuteToolCalls(ctx, a.mcpManager, response)
    
    // Compile comprehensive research report
    if len(toolResults) > 0 {
        reportPrompt := fmt.Sprintf(`Based on this research: %v
        
Create a comprehensive research report with:
1. Executive summary
2. Key findings with sources
3. Detailed information
4. Implications and insights`, toolResults)
        
        finalReport, _ := a.llm.Generate(ctx, reportPrompt)
        state.Set("research_report", finalReport)
        return agentflow.AgentResult{Result: finalReport, State: state}, nil
    }
    
    return agentflow.AgentResult{Result: response, State: state}, nil
}
```

## Custom MCP Servers

### Building a Custom Tool

You can create custom MCP servers for domain-specific tools:

```javascript
// custom-mcp-server.js
const { MCPServer } = require('@modelcontextprotocol/server');

const server = new MCPServer({
    name: "custom-tools",
    version: "1.0.0"
});

// Define custom tool
server.registerTool({
    name: "analyze_data",
    description: "Analyze CSV data and return insights",
    parameters: {
        type: "object",
        properties: {
            data: { type: "string", description: "CSV data to analyze" },
            analysis_type: { type: "string", description: "Type of analysis: summary, trends, outliers" }
        },
        required: ["data", "analysis_type"]
    }
}, async (params) => {
    // Custom analysis logic here
    const { data, analysis_type } = params;
    
    switch(analysis_type) {
        case "summary":
            return { result: "Data summary: ..." };
        case "trends":
            return { result: "Trend analysis: ..." };
        default:
            return { error: "Unknown analysis type" };
    }
});

server.start();
```

**Configure in agentflow.toml:**
```toml
[mcp.servers.custom]
command = "node"
args = ["custom-mcp-server.js"]
transport = "stdio"
```

## Testing Tool Integration

### Mock MCP Manager for Testing

```go
type MockMCPManager struct {
    tools       []agentflow.ToolSchema
    toolResults map[string]interface{}
}

func NewMockMCPManager() *MockMCPManager {
    return &MockMCPManager{
        tools: []agentflow.ToolSchema{
            {
                Name: "search",
                Description: "Search for information",
                Parameters: map[string]interface{}{
                    "query": map[string]interface{}{
                        "type": "string",
                        "description": "Search query",
                    },
                },
            },
        },
        toolResults: map[string]interface{}{
            "search": map[string]interface{}{
                "results": []string{"Mock search result 1", "Mock search result 2"},
            },
        },
    }
}

func (m *MockMCPManager) ListTools(ctx context.Context) ([]agentflow.ToolSchema, error) {
    return m.tools, nil
}

func (m *MockMCPManager) CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
    if result, exists := m.toolResults[name]; exists {
        return result, nil
    }
    return nil, fmt.Errorf("tool not found: %s", name)
}
```

## Performance Considerations

### Tool Execution Optimization

1. **Cache Tool Schemas**: Tool discovery is cached automatically
2. **Parallel Execution**: Multiple tool calls execute concurrently  
3. **Timeout Management**: Tools have configurable timeouts
4. **Connection Pooling**: MCP connections are reused

```go
// Production MCP configuration
config := agentflow.MCPConfig{
    CacheEnabled:     true,
    CacheTTL:        5 * time.Minute,
    ConnectionTimeout: 30 * time.Second,
    MaxRetries:      3,
    MaxConcurrentTools: 5,
}

mcpManager, err := agentflow.InitializeProductionMCP(ctx, config)
```

## Next Steps

- **[LLM Providers](llm-providers.md)** - Configure different LLM providers
- **[Vector Databases](vector-databases.md)** - Set up persistent storage
- **[MCP API Reference](../../reference/api/mcp.md)** - Complete MCP documentation