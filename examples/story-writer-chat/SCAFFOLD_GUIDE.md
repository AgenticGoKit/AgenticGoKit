# Scaffold Guide: Story Writer Chat

## Overview

This example is designed as a **golden copy** for generating agentic workflows. It demonstrates best practices for organizing code into reusable infrastructure and application-specific logic.

## Code Organization Analysis

### Current File Structure

```
story-writer-chat/
├── main.go                  # ✅ Entry point with configuration (SCAFFOLD TEMPLATE)
├── workflow_interface.go    # ✅ Generic interface (REUSABLE - DO NOT MODIFY)
├── websocket_server.go      # ✅ WebSocket infrastructure (REUSABLE - DO NOT MODIFY)
├── session_manager.go       # ✅ Session management (REUSABLE - DO NOT MODIFY)
├── story_workflow.go        # 📝 Application logic (SCAFFOLD TEMPLATE - CUSTOMIZE)
└── frontend/                # 🎨 React UI (REUSABLE - DO NOT MODIFY)
```

### Files Categorization

#### 1. **Reusable Infrastructure** (Do Not Modify) ⭐

These files are framework code that should work with ANY workflow:

- **`workflow_interface.go`** (~35 lines)
  - Defines `AgentInfo` struct
  - Defines `WorkflowExecutor` interface
  - Pure interface definition
  - **Scaffold Action**: Copy as-is

- **`websocket_server.go`** (~240 lines)
  - WebSocket connection handling
  - Message routing
  - Session integration
  - HTTP endpoints (health, sessions, info page)
  - **Scaffold Action**: Copy as-is

- **`session_manager.go`** (~90 lines)
  - Thread-safe session storage
  - Message history management
  - Session lifecycle
  - **Scaffold Action**: Copy as-is

- **`frontend/`** (React + TypeScript)
  - Dynamic UI that adapts to any workflow
  - Receives configuration from backend
  - **Scaffold Action**: Copy entire folder as-is

#### 2. **Scaffold Templates** (Customize) 📝

These files should be used as templates for new workflows:

- **`main.go`** (~85 lines)
  - Configuration loading
  - API validation
  - Workflow initialization
  - Server startup
  - **Scaffold Action**: Template with placeholders

- **`story_workflow.go`** (~400 lines)
  - Agent creation and configuration
  - Workflow execution logic
  - Business-specific logic
  - **Scaffold Action**: Template with replaceable sections

## Improvements Made for Scaffolding

### 1. **Configuration Centralization** (`main.go`)

**Before:**
```go
func main() {
    apiKey, _ := ValidateAPIKey()
    workflow, _ := NewStoryWriterWorkflow(apiKey)
    port := getPort()
    server := NewWebSocketServer(port, workflow)
}
```

**After:**
```go
type Config struct {
    APIKey   string
    Port     string
    Provider string
    Model    string
}

func LoadConfig() (*Config, error) {
    // Centralized configuration loading
}

func main() {
    config, _ := LoadConfig()
    workflow, _ := NewStoryWriterWorkflow(config)
    server := NewWebSocketServer(config.Port, workflow)
}
```

**Benefits:**
- ✅ Single source of configuration
- ✅ Easy to extend with new parameters
- ✅ Template-friendly structure
- ✅ Environment variable support

### 2. **Agent Creation Helper** (`story_workflow.go`)

**Before:**
```go
writer, _ := vnext.QuickChatAgentWithConfig("Writer", &vnext.Config{
    Name:    "writer",
    SystemPrompt: "...",
    Timeout: 90 * time.Second,
    Streaming: &vnext.StreamingConfig{...},
    LLM: vnext.LLMConfig{
        Provider:    "openrouter",
        Model:       "openai/gpt-4o-mini",
        Temperature: 0.8,
        MaxTokens:   800,
        APIKey:      apiKey,
    },
})
// Repeated for each agent...
```

**After:**
```go
type AgentConfig struct {
    SystemPrompt string
    Temperature  float32
    MaxTokens    int
}

func createAgent(displayName, name string, config *Config, agentConfig *AgentConfig) (vnext.Agent, error) {
    // Centralized agent creation with consistent configuration
}

writer, _ := createAgent("Writer", "writer", config, &AgentConfig{
    SystemPrompt: "...",
    Temperature:  0.8,
    MaxTokens:    800,
})
```

**Benefits:**
- ✅ DRY principle - no repeated code
- ✅ Consistent agent configuration
- ✅ Easy to modify defaults
- ✅ Cleaner template generation

### 3. **Better Error Messages**

**Before:**
```go
if err != nil {
    log.Fatalf("Failed to create workflow: %v", err)
}
```

**After:**
```go
if err != nil {
    log.Fatalf("❌ Failed to create workflow: %v", err)
}
```

**Benefits:**
- ✅ Visual indicators (emojis) for quick scanning
- ✅ Consistent error formatting
- ✅ Better user experience

## Scaffold Template Structure

### Template Variables

For scaffold generation, these are the key variables to replace:

| Variable | Example Value | Location | Description |
|----------|---------------|----------|-------------|
| `{{WORKFLOW_NAME}}` | `Story Writer Chat` | `story_workflow.go` | Workflow display name |
| `{{WORKFLOW_STRUCT}}` | `StoryWriterWorkflow` | `story_workflow.go` | Struct name |
| `{{WELCOME_MESSAGE}}` | `Welcome to...` | `story_workflow.go` | Initial greeting |
| `{{AGENT_COUNT}}` | `3` | `story_workflow.go` | Number of agents |
| `{{AGENT_1_NAME}}` | `writer` | `story_workflow.go` | First agent identifier |
| `{{AGENT_1_DISPLAY}}` | `Writer` | `story_workflow.go` | First agent display name |
| `{{AGENT_1_ICON}}` | `✍️` | `story_workflow.go` | First agent icon |
| `{{AGENT_1_COLOR}}` | `blue` | `story_workflow.go` | First agent color |
| `{{AGENT_1_DESCRIPTION}}` | `Creates...` | `story_workflow.go` | First agent description |
| `{{AGENT_1_PROMPT}}` | `You are...` | `story_workflow.go` | First agent system prompt |
| `{{AGENT_1_TEMP}}` | `0.8` | `story_workflow.go` | First agent temperature |
| `{{AGENT_1_TOKENS}}` | `800` | `story_workflow.go` | First agent max tokens |
| `{{PORT}}` | `8080` | `main.go` | Default server port |
| `{{PROVIDER}}` | `openrouter` | `main.go` | LLM provider |
| `{{MODEL}}` | `openai/gpt-4o-mini` | `main.go` | Default model |

### Workflow Execution Template

The `Execute()` method should follow this pattern:

```go
func (w *{{WORKFLOW_STRUCT}}) Execute(ctx context.Context, userInput string, sendMessage MessageSender) error {
    // 1. Send workflow start
    sendMessage(WSMessage{
        Type: MsgTypeWorkflowStart,
        Content: "Starting workflow...",
        Timestamp: float64(time.Now().Unix()),
    })
    
    // 2. Execute agents in sequence/parallel/iterative
    // For each agent:
    //   a. Send agent_start
    //   b. Run agent with streaming
    //   c. Send agent_complete
    
    // 3. Send workflow_done
    sendMessage(WSMessage{
        Type: MsgTypeWorkflowDone,
        Content: finalResult,
        Timestamp: float64(time.Now().Unix()),
    })
    
    return nil
}
```

## Scaffold Generation Strategy

### Phase 1: Gather Requirements

```yaml
workflow:
  name: "Your Workflow Name"
  description: "Brief description"
  welcome_message: "Welcome message"
  
agents:
  - name: "agent1"
    display_name: "Agent 1"
    icon: "🔍"
    color: "blue"
    description: "Agent description"
    system_prompt: "You are..."
    temperature: 0.7
    max_tokens: 800
    
  - name: "agent2"
    display_name: "Agent 2"
    icon: "✍️"
    color: "green"
    description: "Agent description"
    system_prompt: "You are..."
    temperature: 0.6
    max_tokens: 1000

workflow_type: "sequential" | "iterative" | "parallel" | "custom"

config:
  port: 8080
  provider: "openrouter"
  model: "openai/gpt-4o-mini"
```

### Phase 2: Generate Files

#### 2.1 Copy Reusable Infrastructure

```bash
cp workflow_interface.go new-workflow/
cp websocket_server.go new-workflow/
cp session_manager.go new-workflow/
cp -r frontend new-workflow/
```

#### 2.2 Generate main.go from Template

```go
// main.go - Generated from template
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    "github.com/kunalkushwaha/agenticgokit/core/vnext"
    _ "github.com/kunalkushwaha/agenticgokit/plugins/llm/{{PROVIDER}}"
)

func main() {
    config, err := LoadConfig()
    if err != nil {
        log.Fatalf("❌ Configuration error: %v", err)
    }

    if err := ValidateAPIConnection(config.APIKey); err != nil {
        log.Fatalf("❌ API validation failed: %v", err)
    }
    log.Println("✅ API connection validated")

    workflow, err := New{{WORKFLOW_STRUCT}}(config)
    if err != nil {
        log.Fatalf("❌ Failed to create workflow: %v", err)
    }

    server := NewWebSocketServer(config.Port, workflow)
    if err := server.Start(); err != nil {
        log.Fatalf("❌ Server error: %v", err)
    }
}

// ... Config struct and helper functions (standard template)
```

#### 2.3 Generate workflow.go from Template

```go
// {{WORKFLOW_FILE}}.go - Generated from YAML
package main

import (
    "context"
    "fmt"
    "strings"
    "time"
    "github.com/kunalkushwaha/agenticgokit/core/vnext"
)

type {{WORKFLOW_STRUCT}} struct {
    {{range .Agents}}
    {{.Name}} vnext.Agent
    {{end}}
    config *Config
}

func New{{WORKFLOW_STRUCT}}(config *Config) (*{{WORKFLOW_STRUCT}}, error) {
    {{range .Agents}}
    {{.Name}}, err := createAgent("{{.DisplayName}}", "{{.Name}}", config, &AgentConfig{
        SystemPrompt: `{{.SystemPrompt}}`,
        Temperature:  {{.Temperature}},
        MaxTokens:    {{.MaxTokens}},
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create {{.Name}} agent: %w", err)
    }
    {{end}}

    return &{{WORKFLOW_STRUCT}}{
        {{range .Agents}}
        {{.Name}}: {{.Name}},
        {{end}}
        config: config,
    }, nil
}

func (w *{{WORKFLOW_STRUCT}}) Name() string {
    return "{{WORKFLOW_NAME}}"
}

func (w *{{WORKFLOW_STRUCT}}) WelcomeMessage() string {
    return "{{WELCOME_MESSAGE}}"
}

func (w *{{WORKFLOW_STRUCT}}) GetAgents() []AgentInfo {
    return []AgentInfo{
        {{range .Agents}}
        {
            Name:        "{{.Name}}",
            DisplayName: "{{.DisplayName}}",
            Icon:        "{{.Icon}}",
            Color:       "{{.Color}}",
            Description: "{{.Description}}",
        },
        {{end}}
    }
}

// Execute method - Generated based on workflow_type
func (w *{{WORKFLOW_STRUCT}}) Execute(ctx context.Context, userInput string, sendMessage MessageSender) error {
    {{if eq .WorkflowType "sequential"}}
        // Sequential workflow implementation
    {{else if eq .WorkflowType "iterative"}}
        // Iterative workflow implementation
    {{else if eq .WorkflowType "parallel"}}
        // Parallel workflow implementation
    {{end}}
}

// ... Helper functions (standard template)
```

### Phase 3: Workflow Type Templates

#### Sequential Workflow

```go
func (w *{{WORKFLOW_STRUCT}}) Execute(ctx context.Context, userInput string, sendMessage MessageSender) error {
    sendMessage(WSMessage{
        Type: MsgTypeWorkflowStart,
        Content: "Starting {{WORKFLOW_NAME}}...",
        Timestamp: float64(time.Now().Unix()),
    })

    var result string
    {{range $index, $agent := .Agents}}
    {{if eq $index 0}}
    result, err := w.runAgentWithStreaming(ctx, w.{{$agent.Name}}, "{{$agent.Name}}", userInput, sendMessage)
    {{else}}
    result, err = w.runAgentWithStreaming(ctx, w.{{$agent.Name}}, "{{$agent.Name}}", result, sendMessage)
    {{end}}
    if err != nil {
        return fmt.Errorf("{{$agent.Name}} failed: %w", err)
    }
    {{end}}

    sendMessage(WSMessage{
        Type: MsgTypeWorkflowDone,
        Content: result,
        Timestamp: float64(time.Now().Unix()),
    })

    return nil
}
```

#### Iterative Workflow (with feedback loop)

```go
func (w *{{WORKFLOW_STRUCT}}) Execute(ctx context.Context, userInput string, sendMessage MessageSender) error {
    maxIterations := 3
    iteration := 0
    approved := false
    var result string

    // Initial creation
    result, err := w.runAgentWithStreaming(ctx, w.{{AGENT_1}}, "{{AGENT_1}}", userInput, sendMessage)
    if err != nil {
        return err
    }

    // Iterative review loop
    for !approved && iteration < maxIterations {
        review, err := w.runAgentWithStreaming(ctx, w.{{AGENT_2}}, "{{AGENT_2}}", result, sendMessage)
        if err != nil {
            return err
        }

        if strings.Contains(strings.ToUpper(review), "APPROVED") {
            approved = true
            result = review
        } else {
            iteration++
            // Send back to first agent for revision
            result, err = w.runAgentWithStreaming(ctx, w.{{AGENT_1}}, "{{AGENT_1}}", 
                fmt.Sprintf("Revise based on feedback:\n%s", review), sendMessage)
            if err != nil {
                return err
            }
        }
    }

    return nil
}
```

#### Parallel Workflow

```go
func (w *{{WORKFLOW_STRUCT}}) Execute(ctx context.Context, userInput string, sendMessage MessageSender) error {
    type agentResult struct {
        name   string
        result string
        err    error
    }

    resultChan := make(chan agentResult, {{len .Agents}})

    // Launch agents in parallel
    {{range .Agents}}
    go func() {
        result, err := w.runAgentWithStreaming(ctx, w.{{.Name}}, "{{.Name}}", userInput, sendMessage)
        resultChan <- agentResult{name: "{{.Name}}", result: result, err: err}
    }()
    {{end}}

    // Collect results
    results := make(map[string]string)
    for i := 0; i < {{len .Agents}}; i++ {
        ar := <-resultChan
        if ar.err != nil {
            return fmt.Errorf("%s failed: %w", ar.name, ar.err)
        }
        results[ar.name] = ar.result
    }

    // Combine results
    finalResult := combineResults(results)

    sendMessage(WSMessage{
        Type: MsgTypeWorkflowDone,
        Content: finalResult,
        Timestamp: float64(time.Now().Unix()),
    })

    return nil
}
```

## Scaffold CLI Design

### Command Structure

```bash
agenticgokit scaffold init <workflow-name>
# Creates project directory and prompts for configuration

agenticgokit scaffold generate --config workflow.yaml
# Generates workflow from YAML configuration

agenticgokit scaffold add-agent --name researcher --display "Researcher" --icon "🔍"
# Adds a new agent to existing workflow

agenticgokit scaffold validate
# Validates workflow configuration and generated code
```

### Interactive Prompts

```
🚀 Creating new agentic workflow...

Workflow name: My Data Analyzer
Description: Analyzes data and generates reports
Welcome message: Welcome! Upload your data to analyze.

How many agents? 2

Agent 1:
  Name (identifier): analyzer
  Display name: Data Analyzer
  Icon: 🔍
  Color: blue
  Description: Analyzes input data
  System prompt: You are a data analyst...
  Temperature (0.0-1.0): 0.5
  Max tokens: 1000

Agent 2:
  Name (identifier): reporter
  Display name: Report Generator
  Icon: 📊
  Color: green
  Description: Generates comprehensive reports
  System prompt: You are a report writer...
  Temperature (0.0-1.0): 0.6
  Max tokens: 1500

Workflow type:
  1) Sequential (agent1 → agent2 → agent3)
  2) Iterative (agent1 → agent2 → agent1 → ...)
  3) Parallel (agent1, agent2, agent3 simultaneously)
  4) Custom (write your own logic)
  
Select: 1

✅ Generated files:
  - my-data-analyzer/main.go
  - my-data-analyzer/data_analyzer_workflow.go
  - my-data-analyzer/workflow_interface.go
  - my-data-analyzer/websocket_server.go
  - my-data-analyzer/session_manager.go
  - my-data-analyzer/frontend/ (copied)
  
📝 Next steps:
  1. cd my-data-analyzer
  2. Set API key: $env:OPENROUTER_API_KEY="your-key"
  3. Run backend: go run .
  4. Run frontend: cd frontend && npm install && npm run dev
```

## Best Practices for Scaffold Templates

### 1. **Keep Infrastructure Generic**

- ❌ Don't add business logic to infrastructure files
- ✅ Keep WebSocket server, session manager workflow-agnostic
- ✅ Use interfaces for extensibility

### 2. **Use Configuration Structs**

- ❌ Don't hardcode values
- ✅ Load from environment variables
- ✅ Provide sensible defaults
- ✅ Document all configuration options

### 3. **Consistent Naming Conventions**

- Agent internal names: lowercase, snake_case (`data_analyzer`)
- Agent display names: Title Case (`Data Analyzer`)
- Struct names: PascalCase (`DataAnalyzerWorkflow`)
- Functions: camelCase (`runAgentWithStreaming`)

### 4. **Error Handling**

- Always return descriptive errors
- Use `fmt.Errorf` with `%w` for error wrapping
- Log errors with context
- Use emojis for visual distinction (❌, ✅, ⚠️)

### 5. **Documentation**

- Document all public interfaces
- Include usage examples
- Explain configuration options
- Provide troubleshooting tips

### 6. **Template Markers**

Use clear markers for replacement:

```go
// {{SCAFFOLD:BEGIN:AGENT_CREATION}}
writer, err := createAgent("Writer", "writer", config, &AgentConfig{
    SystemPrompt: "{{AGENT_PROMPT}}",
    Temperature:  {{AGENT_TEMP}},
    MaxTokens:    {{AGENT_TOKENS}},
})
// {{SCAFFOLD:END:AGENT_CREATION}}
```

## Testing Scaffold Generation

### Validation Checklist

- [ ] All required files generated
- [ ] No syntax errors in generated code
- [ ] Configuration properly loaded
- [ ] All agents initialized correctly
- [ ] WebSocket connection works
- [ ] Frontend displays agents correctly
- [ ] Messages flow properly
- [ ] Workflow executes without errors
- [ ] Error handling works
- [ ] Documentation is complete

### Test Cases

1. **Minimal Workflow** (1 agent, sequential)
2. **Standard Workflow** (2-3 agents, sequential)
3. **Complex Workflow** (3+ agents, iterative)
4. **Parallel Workflow** (multiple agents simultaneously)
5. **Custom Workflow** (user-defined logic)

## Future Enhancements

### 1. **Web-Based Scaffold Generator**

- Visual workflow designer
- Drag-and-drop agent creation
- Real-time preview
- Download generated project

### 2. **Workflow Templates**

- Pre-built templates for common use cases:
  - Content creation (writer → editor → publisher)
  - Data analysis (loader → analyzer → visualizer)
  - Code review (linter → security → reviewer)
  - Research (searcher → summarizer → organizer)

### 3. **Plugin System**

- Custom agent types
- Custom workflow patterns
- Additional LLM providers
- Custom UI components

### 4. **Workflow Marketplace**

- Share workflows
- Download community workflows
- Rate and review
- Fork and customize

## Conclusion

This scaffold structure provides:

✅ **Clear separation** between reusable and customizable code  
✅ **Template-friendly** structure for code generation  
✅ **Consistent patterns** for predictable scaffolding  
✅ **Extensible design** for future enhancements  
✅ **Production-ready** code following best practices  

The story-writer-chat example serves as the **golden copy** for all future workflow scaffolding!
