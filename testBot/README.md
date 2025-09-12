# testBot

**An intelligent multi-agent system powered by AgenticGoKit**

testBot is a sophisticated multi-agent workflow system that leverages multiple AI agents working sequentially to process and respond to user queries.

## Quick Start

### Prerequisites

- Go 1.21 or later
- OpenAI API key



### Installation

1. **Clone and setup**:
   ```bash
   git clone <your-repository>
   cd testBot
   go mod tidy
   ```

2. **Configure environment**:
   ```bash
   export OPENAI_API_KEY="your-api-key"
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
[agents.agent1]
role = "agent1"
description = "Processes tasks in sequence as part of a processing pipeline"
system_prompt = "You are Agent1, Processes tasks in sequence as part of a processing pipeline"
capabilities = ["general_assistance", "processing"]
enabled = true

# Agent-specific LLM settings
[agents.agent1.llm]
temperature = 0.7
max_tokens = 2000

[agents.agent2]
role = "agent2"
description = "Processes tasks in sequence as part of a processing pipeline"
system_prompt = "You are Agent2, Processes tasks in sequence as part of a processing pipeline"
capabilities = ["general_assistance", "processing"]
enabled = true

# Agent-specific LLM settings
[agents.agent2.llm]
temperature = 0.7
max_tokens = 2000

# Orchestration
[orchestration]
mode = "sequential"
agents = ["agent1", "agent2"]
```

## Architecture

### System Overview

testBot implements a sequential multi-agent architecture with 2 specialized agents:

```
User Input -> Agent1 -> Agent2 -> ... -> Final Response
```

### Project Structure

```
testBot/
|-- agents/                 # Agent implementations
|   |-- agent1.go           # Agent1 agent
|   |-- agent2.go           # Agent2 agent
|   `-- README.md           # Agent documentation
|-- internal/               # Internal packages
|   |-- config/             # Configuration utilities
|   `-- handlers/           # Shared handler utilities
|-- docs/                   # Documentation
|   `-- CUSTOMIZATION.md    # Customization guide
|-- main.go                 # Application entry point
|-- agentflow.toml          # Main configuration file

|-- agentflow.toml          # System configuration
|-- go.mod                  # Go module definition
`-- README.md               # This file
```

### Agent Responsibilities


#### Agent1 (`agents/agent1.go`)

**Purpose**: Processes tasks in sequence as part of a processing pipeline

**Role**: Processes initial user input and prepares data for downstream agents

**Key Features**:
- Memory not enabled
- No tool integration
- No RAG capabilities


#### Agent2 (`agents/agent2.go`)

**Purpose**: Processes tasks in sequence as part of a processing pipeline

**Role**: Finalizes processing and generates the final response

**Key Features**:
- Memory not enabled
- No tool integration
- No RAG capabilities



## Configuration

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
    "testBot/agents"
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
# testBot

An agentic, configuration-driven multi-agent system with a built-in Web UI, streaming, config editor, and flow visualization.

## Highlights (what’s new)

- Web UI with WebSocket-first streaming (HTTP fallback)
- Config editor endpoints (read/write `agentflow.toml` safely)
- Flow tracing and Mermaid sequence diagrams for debugging
- Clean architecture: network handlers moved to `internal/handlers`, config utilities in `internal/config`, tracing in `internal/tracing`

## Prerequisites

- Go 1.21+
- LLM credentials (e.g., set `OPENAI_API_KEY` for OpenAI)

Optional:
- Set `AGENTFLOW_CONFIG_PATH` to point to a non-default `agentflow.toml`

## Quick start

Windows PowerShell:

```powershell
# From repo root
pwsh -NoProfile -Command "Set-Location 'c:\Users\Kunal\work\ZynkWorks\AgenticGoKit\testBot'; go build .; .\testBot.exe -webui"
```

Then open `http://localhost:8080` in your browser.

CLI (no Web UI):

```powershell
pwsh -NoProfile -Command "Set-Location 'c:\Users\Kunal\work\ZynkWorks\AgenticGoKit\testBot'; go run . -m 'Hello there'"
```

## Web UI overview

- Chat area on the right; debug trace pane can be toggled on the left (grid layout, resizes chat)
- Real-time streaming via WebSocket; falls back to HTTP when needed
- Flow view shows actual agent-to-agent transitions (self-loops included), with label tooltips

## API surface

HTTP (JSON):

- `GET /api/agents` → list available agents
- `POST /api/chat` → body: `{ "message": string, "agent": string, "useOrchestration": bool }`
- `GET /api/config/raw` → returns `{ content, path, size }` for `agentflow.toml`
- `PUT /api/config/raw` → body: `{ "toml": string }` (validated then atomically written)
- `GET /api/config` → server features and orchestration info
- `GET /api/visualization/composition` → composition diagram (Mermaid)
- `GET /api/visualization/trace?session=webui-session&linear=true` → sequence diagram + labels

WebSocket (`/ws`):

- Send: `{ type: "chat", agent: string, message: string, useOrchestration?: boolean }`
- Receive events:
    - `welcome`
    - `agent_progress`
    - `agent_chunk` (chunked content for smooth streaming)
    - `agent_complete`

## Configuration

The app is configuration-driven via `agentflow.toml`.

- Read/write endpoints: `GET/PUT /api/config/raw`
- Validation uses `core.LoadConfig` and `ValidateOrchestrationConfig`
- Atomic writes ensure safe updates
- Override path with `AGENTFLOW_CONFIG_PATH`

Tip: validate configs with the CLI from repo root:

```powershell
pwsh -NoProfile -Command "Set-Location 'c:\Users\Kunal\work\ZynkWorks\AgenticGoKit'; .\agentcli.exe validate agentflow.toml"
```

## Project structure (key parts)

```
testBot/
    agents/                      # Agent registrations (imported in main)
    internal/
        config/                    # Config helpers and raw config handlers
        handlers/                  # All HTTP/WS handlers (Server)
        tracing/                   # In-memory tracing + Mermaid builders
        webui/                     # Static assets (HTML/CSS/JS)
    main.go                      # Wiring, startup, orchestration setup
    agentflow.toml               # Main configuration
    go.mod
    README.md
```

## Notes on tracing and visualization

- We record the initial `User -> FirstAgent` hop at entrypoints only
- Outbound edges come from agent results (`RouteMetadataKey`), preserving real routing
- Mermaid sequence diagrams are labeled (compressed `M1/M2…` with hover tooltips)

## Troubleshooting

- Missing API key: set `OPENAI_API_KEY` (or provider-specific vars for Azure/Ollama)
- No responses captured: check logs and your `agents` config in `agentflow.toml`
- Blank trace: ensure tracing is enabled; the app registers framework trace hooks on start

## License

MIT. See `LICENSE` at repo root.

---

Happy building! If you need help, open an issue in the main repository.
