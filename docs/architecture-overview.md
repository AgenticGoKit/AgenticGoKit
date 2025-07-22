# AgenticGoKit Architecture Overview

**A visual guide to understanding how AgenticGoKit works**

AgenticGoKit is designed around a simple but powerful principle: **agents work together through events to solve complex problems**. This overview shows you how the pieces fit together and how data flows through the system.

## 🏗️ High-Level Architecture

```mermaid
flowchart TB
    subgraph "Your Application"
        APP[Your Go Application]
        CONFIG[agentflow.toml]
    end
    
    subgraph "AgenticGoKit Core"
        RUNNER[Runner]
        ORCHESTRATOR[Orchestrator]
        
        subgraph "Agents"
            AGENT1[Research Agent]
            AGENT2[Analysis Agent]
            AGENT3[Validation Agent]
        end
        
        subgraph "Infrastructure"
            STATE[State Manager]
            MEMORY[Memory System]
            TOOLS[MCP Tools]
        end
    end
    
    subgraph "External Services"
        LLM[LLM Providers<br/>OpenAI, Azure, Ollama]
        MCPSERVERS[MCP Servers<br/>Web Search, Files, APIs]
        DATABASES[Vector Databases<br/>PostgreSQL, Weaviate]
    end
    
    APP --> RUNNER
    CONFIG --> RUNNER
    RUNNER --> ORCHESTRATOR
    ORCHESTRATOR --> AGENT1
    ORCHESTRATOR --> AGENT2
    ORCHESTRATOR --> AGENT3
    
    AGENT1 --> STATE
    AGENT2 --> STATE
    AGENT3 --> STATE
    
    AGENT1 --> MEMORY
    AGENT2 --> MEMORY
    AGENT3 --> MEMORY
    
    AGENT1 --> TOOLS
    AGENT2 --> TOOLS
    AGENT3 --> TOOLS
    
    AGENT1 --> LLM
    AGENT2 --> LLM
    AGENT3 --> LLM
    
    TOOLS --> MCPSERVERS
    MEMORY --> DATABASES
    
    classDef app fill:#e3f2fd,stroke:#1976d2,stroke-width:2px
    classDef core fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px
    classDef agents fill:#e8f5e8,stroke:#388e3c,stroke-width:2px
    classDef infra fill:#fff3e0,stroke:#f57c00,stroke-width:2px
    classDef external fill:#fce4ec,stroke:#c2185b,stroke-width:2px
    
    class APP,CONFIG app
    class RUNNER,ORCHESTRATOR core
    class AGENT1,AGENT2,AGENT3 agents
    class STATE,MEMORY,TOOLS infra
    class LLM,MCPSERVERS,DATABASES external
```

## 🔄 How Data Flows Through the System

```mermaid
sequenceDiagram
    participant User
    participant Runner
    participant Orchestrator
    participant Agent
    participant LLM
    participant Tools
    participant Memory
    
    User->>Runner: "Analyze market trends for Tesla"
    Runner->>Memory: Get conversation context
    Memory-->>Runner: Previous context
    
    Runner->>Orchestrator: Route event with context
    Orchestrator->>Agent: Assign to Research Agent
    
    Agent->>Memory: Get relevant knowledge
    Memory-->>Agent: Related documents
    
    Agent->>Tools: Search for current Tesla data
    Tools-->>Agent: Latest market data
    
    Agent->>LLM: Process with context + tools + memory
    LLM-->>Agent: Analysis response
    
    Agent->>Memory: Store new insights
    Agent-->>Orchestrator: Return analysis
    Orchestrator-->>Runner: Complete result
    Runner-->>User: "Here's the Tesla analysis..."
```

## 🧩 Core Components Explained

### 🎯 **Runner** - The Central Coordinator
The Runner is your main entry point. It:
- Receives your requests and creates events
- Manages the overall workflow
- Handles configuration and setup
- Returns final results to your application

```go
// Simple usage
runner := core.CreateCollaborativeRunner(agents, 30*time.Second)
result := runner.Emit(core.NewEvent("user_query", "Analyze market trends"))
```

### 🎭 **Orchestrator** - The Traffic Director
The Orchestrator decides how agents work together:
- **Route Mode**: Sends events to specific agents
- **Collaborative Mode**: All agents work in parallel
- **Sequential Mode**: Agents work in a pipeline
- **Mixed Mode**: Combines parallel and sequential patterns

### 🤖 **Agents** - The Workers
Agents are where the magic happens:
- Each agent has a specific role (research, analysis, validation)
- They can use tools and access memory
- They communicate with LLMs to process information
- They pass results to other agents or back to the user

### 💾 **State Manager** - The Memory Keeper
Manages data flow between agents:
- Stores conversation context
- Passes data between agents
- Maintains session information
- Handles concurrent access safely

### 🧠 **Memory System** - The Knowledge Base
Provides persistent knowledge and context:
- Stores documents and conversations
- Enables RAG (Retrieval-Augmented Generation)
- Supports vector search for relevant information
- Works with PostgreSQL, Weaviate, or in-memory storage

### 🛠️ **MCP Tools** - The Capabilities
Model Context Protocol tools extend what agents can do:
- Web search for real-time information
- File operations for document processing
- Database queries for data access
- Custom tools for specific needs

## 🎨 Orchestration Patterns

### 🎯 Route Pattern - Single Agent Processing
```mermaid
flowchart LR
    Query[User Query] --> Router{Router}
    Router --> Agent1[Research Agent]
    Router --> Agent2[Analysis Agent]
    Router --> Agent3[Validation Agent]
    Agent1 --> Result1[Research Result]
    Agent2 --> Result2[Analysis Result]
    Agent3 --> Result3[Validation Result]
```
**When to use**: Simple queries that need specific expertise

### 🤝 Collaborative Pattern - Parallel Processing
```mermaid
flowchart TD
    Query[User Query] --> Dispatcher[Dispatcher]
    Dispatcher --> Agent1[Research Agent]
    Dispatcher --> Agent2[Analysis Agent]
    Dispatcher --> Agent3[Validation Agent]
    Agent1 --> Aggregator[Result Aggregator]
    Agent2 --> Aggregator
    Agent3 --> Aggregator
    Aggregator --> FinalResult[Combined Result]
```
**When to use**: Complex queries that benefit from multiple perspectives

### 🔄 Sequential Pattern - Pipeline Processing
```mermaid
flowchart LR
    Query[User Query] --> Agent1[Research Agent]
    Agent1 --> Agent2[Analysis Agent]
    Agent2 --> Agent3[Validation Agent]
    Agent3 --> Agent4[Summary Agent]
    Agent4 --> Result[Final Result]
```
**When to use**: Multi-step processes where each step builds on the previous

## 🚀 Getting Started with the Architecture

### 1. **Start Simple** - Single Agent
```go
// Create a simple agent
agent := &MyAgent{}
runner := core.CreateRunner(map[string]core.AgentHandler{
    "my_agent": agent,
})

// Send a query
result := runner.Emit(core.NewEvent("user_query", "Hello world"))
```

### 2. **Add Collaboration** - Multiple Agents
```go
// Create multiple agents
agents := map[string]core.AgentHandler{
    "researcher": &ResearchAgent{},
    "analyzer":   &AnalysisAgent{},
    "validator":  &ValidationAgent{},
}

// Use collaborative orchestration
runner := core.CreateCollaborativeRunner(agents, 30*time.Second)
```

### 3. **Add Memory** - Persistent Context
```go
// Enable memory system
config := core.Config{
    AgentMemory: core.MemoryConfig{
        Provider:   "pgvector",
        Connection: "postgres://localhost/agentflow",
    },
}
runner := core.NewRunnerFromConfig("agentflow.toml")
```

### 4. **Add Tools** - External Capabilities
```go
// Tools are configured in agentflow.toml
[mcp]
enabled = true

[[mcp.servers]]
name = "web-search"
type = "stdio"
command = "npx @modelcontextprotocol/server-web-search"
enabled = true
```

## 🔧 Configuration Architecture

AgenticGoKit uses a layered configuration approach:

```mermaid
flowchart TD
    subgraph "Configuration Sources"
        CLI[Command Line Args]
        ENV[Environment Variables]
        TOML[agentflow.toml]
        DEFAULTS[Built-in Defaults]
    end
    
    subgraph "Configuration Sections"
        AGENT[Agent Configuration]
        ORCHESTRATION[Orchestration Settings]
        MEMORY[Memory Configuration]
        MCP[MCP Tool Settings]
        PROVIDERS[LLM Provider Settings]
    end
    
    CLI --> AGENT
    ENV --> AGENT
    TOML --> AGENT
    DEFAULTS --> AGENT
    
    CLI --> ORCHESTRATION
    ENV --> ORCHESTRATION
    TOML --> ORCHESTRATION
    DEFAULTS --> ORCHESTRATION
    
    CLI --> MEMORY
    ENV --> MEMORY
    TOML --> MEMORY
    DEFAULTS --> MEMORY
    
    CLI --> MCP
    ENV --> MCP
    TOML --> MCP
    DEFAULTS --> MCP
    
    CLI --> PROVIDERS
    ENV --> PROVIDERS
    TOML --> PROVIDERS
    DEFAULTS --> PROVIDERS
```

**Priority Order**: CLI Args > Environment Variables > TOML File > Defaults

## 🎯 Key Design Principles

### 1. **Simple by Default, Powerful When Needed**
- Start with a single agent and basic configuration
- Add complexity (multiple agents, memory, tools) as you need it
- Sensible defaults for everything

### 2. **Event-Driven Architecture**
- Everything communicates through events
- Loose coupling between components
- Easy to test and debug

### 3. **Provider Agnostic**
- Switch between OpenAI, Azure, Ollama without changing code
- Same interface for all LLM providers
- Easy to add new providers

### 4. **Production Ready**
- Built-in error handling and retry logic
- Monitoring and metrics support
- Horizontal scaling capabilities

## 🔍 Debugging and Observability

AgenticGoKit provides multiple ways to understand what's happening:

### Event Tracing
```bash
# View detailed execution flow
agentcli trace --verbose <session-id>
```

### State Inspection
```go
// Access current state in agents
func (a *MyAgent) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    // Inspect current state
    fmt.Printf("Current state: %+v\n", state.Data)
    
    // Your agent logic here
    return result, nil
}
```

### Performance Monitoring
```toml
# Enable metrics in agentflow.toml
[monitoring]
enabled = true
port = 8080
```

## 🚀 Next Steps

Now that you understand the architecture:

1. **Try the [5-Minute Quickstart](quickstart.md)** - Get hands-on experience
2. **Explore [15-Minute Tutorials](tutorials/15-minute-series/)** - Learn specific patterns
3. **Check [Core Concepts](tutorials/core-concepts/)** - Deep dive into components
4. **Build [Production Systems](tutorials/15-minute-series/production-deployment.md)** - Scale your agents

## 🤔 Common Questions

**Q: How do I choose between orchestration patterns?**
A: Start with Route for simple queries, use Collaborative for parallel processing, and Sequential for multi-step workflows.

**Q: When should I add memory?**
A: Add memory when you need conversation context, document search, or knowledge persistence across sessions.

**Q: How do I add custom tools?**
A: Create MCP servers for your tools, or implement custom tool interfaces directly in your agents.

**Q: Can I mix different LLM providers?**
A: Yes! Different agents can use different providers, and you can switch providers without changing agent code.

The architecture is designed to grow with your needs - start simple and add complexity as your use cases evolve.