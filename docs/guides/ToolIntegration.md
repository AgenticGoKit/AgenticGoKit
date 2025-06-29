# Tool Integration

**Dynamic Tool Discovery and Execution with MCP Protocol**

AgentFlow uses the Model Context Protocol (MCP) to provide agents with dynamic tool discovery and execution capabilities. This guide covers everything from basic tool usage to building custom MCP servers.

## Overview

The MCP integration in AgentFlow provides:
- **Dynamic Discovery**: Tools are discovered at runtime, not hard-coded
- **Schema-Based**: Tools provide their own descriptions and parameters
- **LLM-Driven**: The LLM decides which tools to use based on context
- **Extensible**: Add new tools by connecting MCP servers

## Core MCP Functions

AgentFlow provides a simple interface for MCP tool execution:

### ExecuteMCPTool()

The simplest way to execute a tool with automatic caching:

```go
result, err := agentflow.ExecuteMCPTool(ctx, "search", map[string]interface{}{
    "query": "latest Go tutorials",
})
```

**Example usage:**
```go
func searchExample(ctx context.Context) error {
    result, err := agentflow.ExecuteMCPTool(ctx, "search", map[string]interface{}{
        "query": "AgentFlow documentation",
        "limit": 5,
    })
    if err != nil {
        return fmt.Errorf("tool execution failed: %w", err)
    }
    
    if !result.Success {
        return fmt.Errorf("tool returned error: %s", result.Error)
    }
    
    for _, content := range result.Content {
        fmt.Printf("Result: %s\n", content.Text)
    }
    
    return nil
}
```

### GetAvailableTools()

Discover tools from connected MCP servers:

```go
mcpManager := agentflow.GetMCPManager()
tools := mcpManager.GetAvailableTools()

for _, tool := range tools {
    fmt.Printf("Tool: %s - %s (from %s)\n", 
        tool.Name, tool.Description, tool.ServerName)
}

Let me also fetch the official Go documentation.

<tool_call>
{"name": "fetch_content", "args": {"url": "https://golang.org/doc/"}}
</tool_call>
```

## Complete Agent with Tools

Here's a complete MCP-aware agent that can use tools intelligently:

```go
package main

import (
    "context"
    "fmt"
    agentflow "github.com/kunalkushwaha/agentflow/core"
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
    logger := agentflow.Logger()
    
    // Extract user message
    message := event.GetData()["message"].(string)
    
    // Get available tools
    tools := a.mcpManager.GetAvailableTools()
    logger.Info().Int("available_tools", len(tools)).Msg("Tools discovered")
    
    // Build prompt with tool information
    toolsDesc := a.buildToolsDescription(tools)
    
    systemPrompt := fmt.Sprintf(`You are a helpful assistant with access to tools.

Available tools:
%s

Instructions:
- Analyze the user's request carefully
- Determine if any tools would be helpful
- If yes, respond with: USE_TOOL: tool_name with arguments
- If no tools needed, provide a direct answer

User request: %s`, toolsDesc, message)
    
    // Get LLM response to determine tool usage
    response, err := a.llm.Generate(ctx, systemPrompt)
    if err != nil {
        return agentflow.AgentResult{}, fmt.Errorf("LLM generation failed: %w", err)
    }
    
    // Check if LLM wants to use a tool
    if a.shouldUseTool(response) {
        toolName, args := a.parseToolRequest(response)
        
        // Execute the tool
        result, err := agentflow.ExecuteMCPTool(ctx, toolName, args)
        if err != nil {
            logger.Warn().Err(err).Str("tool", toolName).Msg("Tool execution failed")
            return agentflow.AgentResult{
                Data: map[string]interface{}{
                    "response": "I'm sorry, I couldn't execute the requested tool.",
                    "error":    err.Error(),
                },
            }, nil
        }
        
        // Generate final response incorporating tool results
        finalPrompt := fmt.Sprintf(`Based on the tool execution results below, provide a helpful response to the user.

User request: %s
Tool used: %s
Tool results: %v

Please provide a clear, helpful response incorporating this information:`, 
            message, toolName, result)
        
        finalResponse, err := a.llm.Generate(ctx, finalPrompt)
        if err != nil {
            return agentflow.AgentResult{}, fmt.Errorf("final response generation failed: %w", err)
        }
        
        return agentflow.AgentResult{
            Data: map[string]interface{}{
                "response":    finalResponse,
                "tool_used":   toolName,
                "tool_result": result,
            },
        }, nil
    }
    
    // No tool needed, return direct response
    return agentflow.AgentResult{
        Data: map[string]interface{}{
            "response": response,
        },
    }, nil
}

func (a *ToolEnabledAgent) buildToolsDescription(tools []agentflow.MCPToolInfo) string {
    if len(tools) == 0 {
        return "No tools available."
    }
    
    desc := ""
    for _, tool := range tools {
        desc += fmt.Sprintf("- %s: %s\n", tool.Name, tool.Description)
    }
    return desc
}

func (a *ToolEnabledAgent) shouldUseTool(response string) bool {
    return strings.Contains(response, "USE_TOOL:")
}

func (a *ToolEnabledAgent) parseToolRequest(response string) (string, map[string]interface{}) {
    // Simple parsing for demo - in production, use more robust parsing
    parts := strings.Split(response, "USE_TOOL:")
    if len(parts) < 2 {
        return "", nil
    }
    
    toolPart := strings.TrimSpace(parts[1])
    // For demo, assume format: "tool_name arg1=value1 arg2=value2"
    words := strings.Fields(toolPart)
    if len(words) == 0 {
        return "", nil
    }
    
    toolName := words[0]
    args := make(map[string]interface{})
    
    // Parse simple key=value pairs
    for _, word := range words[1:] {
        if kv := strings.Split(word, "="); len(kv) == 2 {
            args[kv[0]] = kv[1]
        }
    }
    
    return toolName, args
}
            if err != nil {
                // Fallback to original response if synthesis fails
                logger.Warn().Err(err).Msg("Tool result synthesis failed, using original response")
                finalResponse = response
            }
        } else {
            finalResponse = response
        }
    } else {
        finalResponse = response
    }
    
    // Update state with results
    state.Set("response", finalResponse)
    state.Set("tools_used", len(toolResults) > 0)
    if len(toolResults) > 0 {
        state.Set("tool_results", toolResults)
    }
    
    return agentflow.AgentResult{
        Result: finalResponse,
        State:  state,
    }, nil
}
```

## MCP Configuration

### Basic Configuration (agentflow.toml)

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

### Production Configuration with Caching

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

### Technical Operations Pattern

```go
func (a *DevOpsAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    task := event.GetData()["message"]
    
    systemPrompt := `You are a DevOps agent with access to Docker and system tools.

For technical tasks:
1. Analyze what needs to be done
2. Use docker commands for container operations
3. Use filesystem tools for file operations  
4. Provide clear explanations of actions taken
5. Include any warnings or important notes

Be careful with destructive operations and always explain what you're doing.`
    
    toolPrompt := agentflow.FormatToolsForPrompt(ctx, a.mcpManager)
    prompt := fmt.Sprintf("%s\n%s\nTask: %s", systemPrompt, toolPrompt, task)
    
    response, err := a.llm.Generate(ctx, prompt)
    if err != nil {
        return agentflow.AgentResult{}, err
    }
    
    // Execute technical tools with logging
    logger := agentflow.Logger()
    toolResults := agentflow.ParseAndExecuteToolCalls(ctx, a.mcpManager, response)
    
    if len(toolResults) > 0 {
        logger.Info().Interface("results", toolResults).Msg("Technical operations completed")
        
        // Provide detailed summary of actions
        summaryPrompt := fmt.Sprintf(`Technical operations completed: %v
        
Provide a summary that includes:
1. What actions were taken
2. Results of each operation
3. Current state
4. Any follow-up recommendations`, toolResults)
        
        summary, _ := a.llm.Generate(ctx, summaryPrompt)
        state.Set("operations_summary", summary)
        return agentflow.AgentResult{Result: summary, State: state}, nil
    }
    
    return agentflow.AgentResult{Result: response, State: state}, nil
}
```

## Error Handling for Tools

### Graceful Tool Failure Handling

```go
func (a *Agent) handleToolExecution(ctx context.Context, response string) (string, error) {
    logger := agentflow.Logger()
    
    // Attempt tool execution
    toolResults := agentflow.ParseAndExecuteToolCalls(ctx, a.mcpManager, response)
    
    // Check if tools were expected but failed
    if strings.Contains(response, "<tool_call>") && len(toolResults) == 0 {
        logger.Warn().Msg("Tool calls were attempted but none succeeded")
        
        // Generate fallback response
        fallbackPrompt := fmt.Sprintf(`The following response contained tool calls that failed to execute:

%s

Please provide a helpful response based on your knowledge instead, and mention that some real-time information couldn't be retrieved.`, response)
        
        fallbackResponse, err := a.llm.Generate(ctx, fallbackPrompt)
        if err != nil {
            // If even fallback fails, return original response
            return response, nil
        }
        return fallbackResponse, nil
    }
    
    // If tools succeeded, synthesize results
    if len(toolResults) > 0 {
        synthesisPrompt := fmt.Sprintf("Response: %s\nTool results: %v\nFinal answer:", response, toolResults)
        finalResponse, err := a.llm.Generate(ctx, synthesisPrompt)
        if err != nil {
            // Fallback to original response
            return response, nil
        }
        return finalResponse, nil
    }
    
    // No tools were called, return original response
    return response, nil
}
```

## Custom MCP Servers

### Building a Custom Tool

You can create custom MCP servers for domain-specific tools. Here's a conceptual example:

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

// Test with mock
func TestAgentWithTools(t *testing.T) {
    mockLLM := &MockModelProvider{}
    mockMCP := NewMockMCPManager()
    agent := NewToolEnabledAgent(mockLLM, mockMCP)
    
    // Test tool integration
    eventData := agentflow.EventData{"message": "Search for Go tutorials"}
    event := agentflow.NewEvent("test", eventData, nil)
    state := agentflow.NewState()
    
    result, err := agent.Run(context.Background(), event, state)
    assert.NoError(t, err)
    assert.NotEmpty(t, result.Result)
    
    // Verify tools were used
    toolsUsed, _ := result.State.Get("tools_used")
    assert.True(t, toolsUsed.(bool))
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

- **[LLM Providers](Providers.md)** - Configure different LLM providers
- **[Configuration](Configuration.md)** - Advanced configuration options
- **[Custom Tools](CustomTools.md)** - Build your own MCP servers
- **[Production Deployment](Production.md)** - Deploy with tool integration
