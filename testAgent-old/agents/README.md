# Agents Directory

This directory contains all the agent implementations for your multi-agent system. Each agent represents a specific processing step or capability in your workflow.

## Directory Structure

```
agents/
|-- agent1.go          # First agent in your workflow
|-- agent2.go          # Second agent (if configured)
|-- ...                # Additional agents as configured
`-- README.md          # This file
```

## Agent Overview


### Agent1 (agent1.go)

**Purpose**: Processes tasks in sequence as part of a processing pipeline

**Role in Workflow**: Sequential processing - processes input from previous agents

**Key Responsibilities**:
- Process input from user messages
- 
- 
- Generate meaningful responses for downstream processing


### Agent2 (agent2.go)

**Purpose**: Processes tasks in sequence as part of a processing pipeline

**Role in Workflow**: Sequential processing - processes input from previous agents

**Key Responsibilities**:
- Process input from previous agents
- 
- 
- Generate meaningful responses for downstream processing



## Customization Guide

### Adding New Agents

1. **Create a new agent file** (e.g., `new_agent.go`):
   ```go
   package agents
   
   import (
       "context"
       agenticgokit "github.com/kunalkushwaha/agenticgokit/core"
   )
   
   type NewAgentHandler struct {
       llm agenticgokit.ModelProvider
       // Add your custom fields here
   }
   
   func NewNewAgent(llmProvider agenticgokit.ModelProvider) *NewAgentHandler {
       return &NewAgentHandler{llm: llmProvider}
   }
   
   func (a *NewAgentHandler) Run(ctx context.Context, event agenticgokit.Event, state agenticgokit.State) (agenticgokit.AgentResult, error) {
       // Implement your agent logic here
       return agenticgokit.AgentResult{}, nil
   }
   ```

2. **Register the agent** in `main.go`:
   ```go
   newAgent := agents.NewNewAgent(llmProvider)
   agents["new_agent"] = newAgent
   ```

3. **Update configuration** in `agentflow.toml` if needed

### Modifying Existing Agents

Each agent file contains comprehensive TODO comments indicating customization points:

- **Input Processing**: Modify how agents handle and validate input
- **Business Logic**: Add your domain-specific processing logic
- **LLM Interaction**: Customize prompts and response handling
- **Output Formatting**: Change how results are structured and returned
- **Error Handling**: Add custom error handling for your use cases

### Common Customization Patterns

#### 1. Adding External Service Integration

```go
type MyAgentHandler struct {
    llm        agenticgokit.ModelProvider
    database   *sql.DB
    apiClient  *http.Client
}

func (a *MyAgentHandler) callExternalAPI(ctx context.Context, data interface{}) (interface{}, error) {
    // Implement API call logic
    return nil, nil
}
```

#### 2. Adding Custom Validation

```go
func (a *MyAgentHandler) validateInput(input interface{}) error {
    // Add your validation logic
    return nil
}
```

#### 3. Adding Custom Output Processing

```go
func (a *MyAgentHandler) formatOutput(output string) (string, error) {
    // Add your formatting logic
    return output, nil
}
```

## Workflow Integration

### Sequential Mode
Agents process input in order: Agent1 -> Agent2 -> Agent3...
- Each agent receives output from the previous agent
- Final agent produces the workflow result

### Collaborative Mode  
Agents process input in parallel and results are combined
- All agents receive the same initial input
- Results are aggregated based on configuration

### Loop Mode
A single agent processes input iteratively
- Agent processes input multiple times
- Each iteration can refine the previous result

### Mixed Mode
Combination of collaborative and sequential processing
- First phase: collaborative processing
- Second phase: sequential processing of aggregated results

## Memory System Integration


Memory system is not enabled for this project. To enable it, update your `agentflow.toml` configuration.


## Tool Integration (MCP)


MCP (Model Context Protocol) is not enabled for this project. To enable tool integration, update your `agentflow.toml` configuration.


## Best Practices

### 1. Error Handling
- Always handle errors gracefully
- Provide meaningful error messages
- Consider fallback strategies for external service failures

### 2. Logging
- Use structured logging with context
- Log important processing steps
- Include agent name and relevant metadata

### 3. Performance
- Avoid blocking operations in agent logic
- Use context for timeout handling
- Consider caching for expensive operations

### 4. Testing
- Write unit tests for your agent logic
- Test error conditions and edge cases
- Use dependency injection for testability

### 5. Documentation
- Keep TODO comments updated as you customize
- Document your business logic and assumptions
- Maintain this README as you add new agents

## Debugging

### Common Issues

1. **Agent not receiving input**
   - Check orchestration configuration in `agentflow.toml`
   - Verify agent registration in `main.go`
   - Check routing logic and agent names

2. **LLM provider errors**
   - Verify environment variables are set
   - Check API key validity and quotas
   - Ensure network connectivity

3. **Memory system issues**
   - Verify database connections (for pgvector/weaviate)
   - Check embedding model configuration
   - Validate memory provider settings

4. **Tool integration problems**
   - Check MCP server configuration
   - Verify tool availability and permissions
   - Review tool execution logs

### Debug Mode

Run with debug logging to see detailed execution flow:
```bash
go run . -m "your message" --debug
```

## Additional Resources

- [AgenticGoKit Documentation](https://github.com/kunalkushwaha/agenticgokit)
- [Multi-Agent Patterns](https://github.com/kunalkushwaha/agenticgokit/docs/patterns)
- [MCP Protocol Specification](https://modelcontextprotocol.io/)
- [Configuration Reference](../agentflow.toml)

---

**Need help?** Check the main project README or create an issue in the AgenticGoKit repository.
