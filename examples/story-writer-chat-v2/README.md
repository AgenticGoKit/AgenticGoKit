# Story Writer Chat App v2 📖

> **Demonstration of AgenticGoKit's SubWorkflow capabilities with nested agent streaming, conditional loops, and real-time UI updates**

An AI-powered collaborative story writing application featuring a sophisticated 3-agent workflow with Writer↔Editor revision loop, real-time streaming, and modern React UI. This example demonstrates advanced patterns for building production-ready agentic applications.

## 📁 Project Structure

```
story-writer-chat-v2/
├── workflow/                    # Modular workflow components
│   ├── story_workflow.go       # Main workflow orchestration
│   ├── agents.go               # Agent creation & configuration
│   ├── prompts.go              # System prompts (centralized)
│   ├── transforms.go           # Input transformation functions
│   └── handlers.go             # Stream chunk processing
├── infrastructure/              # Communication layer
│   ├── websocket.go            # WebSocket server
│   └── messages.go             # Message protocol definitions
├── config/                      # Configuration management
│   └── config.go               # Environment-based config loading
├── frontend/                    # React UI
│   └── src/
│       ├── App.tsx             # Main application
│       ├── components/         # Chat & agent UI components
│       └── hooks/              # WebSocket & state management
├── ARCHITECTURE.md              # Complete architecture guide
└── main.go                     # Application entry point
```

**Key Design Principles**:
- **Modular Workflow**: Separate files for agents, prompts, transforms, handlers (~150 lines each)
- **Infrastructure Layer**: Reusable WebSocket server and message protocol
- **Streaming First**: Real-time token-by-token updates with agent lifecycle events
- **Nested Workflows**: SubWorkflow pattern for complex agent interactions

## 🎯 Overview

This example showcases **AgenticGoKit's most advanced features**:

- **SubWorkflow Composition**: Workflows wrapped as agents for hierarchical execution
- **Conditional Loop Termination**: Dynamic exit conditions (not just max iterations)
- **Nested Agent Streaming**: Proper lifecycle events for agents within SubWorkflows
- **Real-time UI Updates**: WebSocket protocol with agent status tracking
- **Modular Architecture**: Clean separation of concerns for maintainability

### The Writing Team

Three specialized AI agents collaborate in a sophisticated workflow:

1. **✍️ Writer Agent** (with Memory)
   - Creates initial story drafts based on user prompt
   - Revises stories by retrieving editor feedback from memory
   - Uses Transform function to detect revision vs initial draft
   - Low temperature (0.3) for consistency

2. **✏️ Editor Agent** (Loop Controller)
   - Reviews stories for spelling errors and quality
   - Provides specific corrections: `FIX: misspeling→misspelling`
   - Outputs `APPROVED: [story]` when satisfied
   - **Triggers loop exit** via conditional termination

3. **📚 Publisher Agent**
   - Formats approved stories with markdown structure
   - Adds titles and professional presentation
   - Final step in the pipeline

### Workflow Architecture

```
Sequential Pipeline (SubWorkflow)
    │
    ├─► Step 1: Revision Loop (SubWorkflow)
    │       │
    │       └─► Loop (Conditional Exit on "APPROVED")
    │           ├─► Writer Agent (draft/revise)
    │           │       ↓
    │           └─► Editor Agent (review/approve)
    │                   │
    │                   ├─► Contains "APPROVED"? → EXIT LOOP
    │                   └─► Otherwise → ITERATE (max 3)
    │
    └─► Step 2: Publisher Agent (format)
            ↓
        Final Story
```

**Key Technical Features**:
- **SubWorkflow Nesting**: Loop workflow wrapped as agent, Sequential workflow wraps the loop
- **Conditional Exit**: `OutputContains("APPROVED")` dynamically terminates loop
- **Streaming Preservation**: All nested agents emit proper `agent_start`/`agent_complete` events
- **Memory Integration**: Writer retrieves editor feedback from conversation history

## ✨ Features

### AgenticGoKit Framework Features Demonstrated

- **🔀 SubWorkflow Composition**: Workflows as agents for hierarchical execution
- **🔁 Conditional Loop Termination**: Dynamic exit via `OutputContains()` builder
- **📡 Lifecycle Streaming**: `ChunkTypeAgentStart` / `ChunkTypeAgentComplete` events
- **🧠 Agent Memory**: Writer retrieves editor feedback from conversation history
- **🎯 Input Transforms**: Dynamic prompt modification per agent
- **🔧 Modular Workflow Design**: 5 focused files (agents, prompts, transforms, handlers, workflow)

### User Interface

- **💬 Real-time Chat Bubbles** - Separate bubbles for Writer, Editor, Publisher
- **📊 Agent Status Panel** - Visual tracking of which agent is active
- **⚡ Token-by-Token Streaming** - Smooth, responsive text generation
- **🔄 Loop Progress** - "Iteration 2 starting..." messages
- **✅ Approval Notifications** - "Story approved after 2 iteration(s)"
- **🎨 Agent Icons & Colors** - Visual differentiation (✍️ blue, ✏️ green, 📚 purple)

### Technical Architecture

- **🌐 WebSocket Protocol** - 8 message types (welcome, agent_start, agent_progress, etc.)
- **📦 Clean Separation** - Infrastructure vs Workflow vs Configuration layers
- **🚀 Production Patterns** - Error handling, health checks, CORS configuration
- **📖 Comprehensive Documentation** - ARCHITECTURE.md with sequence diagrams
- **🧪 Test Coverage** - 62 passing workflow tests in core framework

## 🚀 Quick Start

### Prerequisites
- **Go 1.23+** - For backend server
- **Node.js 18+** and npm - For React frontend
- **LLM API Key** - HuggingFace, OpenRouter, or OpenAI

### 1. Configure Environment

Create `.env` file in project root:

```bash
# LLM Provider Configuration
LLM_PROVIDER=huggingface
LLM_MODEL=Qwen/Qwen2.5-72B-Instruct
LLM_API_KEY=your-huggingface-api-key

# Or use OpenRouter
# LLM_PROVIDER=openrouter
# LLM_MODEL=openai/gpt-4o-mini
# LLM_API_KEY=your-openrouter-key
```

**PowerShell Alternative**:
```powershell
$env:LLM_PROVIDER="huggingface"
$env:LLM_MODEL="Qwen/Qwen2.5-72B-Instruct"
$env:LLM_API_KEY="your-api-key"
```

### 2. Start Backend Server

```powershell
# From story-writer-chat-v2 directory
go run main.go
```

Expected output:
```
🚀 Starting Story Writer Chat v2 with SubWorkflow Demo...
✅ LLM API connection validated successfully
📡 WebSocket API: ws://localhost:8080/ws
🌐 Health Check: http://localhost:8080/health
💡 Frontend: http://localhost:5173
```

### 3. Start Frontend (Separate Terminal)

```powershell
cd frontend
npm install    # First time only
npm run dev
```

### 4. Open Browser

Navigate to **http://localhost:5173** and try:

```
Write a story in 100 words about a Cow. Make some typos in story
```

**Expected Behavior**:
- ✍️ Writer creates draft with intentional typos
- ✏️ Editor reviews and requests fixes
- 🔄 Loop iteration 2: Writer revises
- ✏️ Editor approves
- 📚 Publisher formats final story
- ✅ "Story approved after 2 iteration(s)"

## 🏗️ Architecture

### Code Organization (Refactored from 600+ to 5 modular files)

#### `workflow/` Package - Modular Workflow Components

**story_workflow.go** (~150 lines) - Main orchestration
```go
type StoryWriterWorkflow struct {
    writer        vnext.Agent  // Individual agents
    editor        vnext.Agent
    publisher     vnext.Agent
    revisionLoop  vnext.Agent  // SubWorkflow: Loop
    storyPipeline vnext.Agent  // SubWorkflow: Sequential
}

func (sw *StoryWriterWorkflow) Execute(ctx, prompt, sendMessage)
    // Uses StreamHandler to process chunks
    // Converts framework events → WebSocket messages
```

**agents.go** - Agent creation functions
```go
func CreateWriter(cfg *config.Config) (vnext.Agent, error)
func CreateEditor(cfg *config.Config) (vnext.Agent, error)
func CreatePublisher(cfg *config.Config) (vnext.Agent, error)
```

**prompts.go** - System prompts as constants
```go
const WriterSystemPrompt = `You are a story writer...`
const EditorSystemPrompt = `You are a spelling checker...`
const PublisherSystemPrompt = `You format text...`
```

**transforms.go** - Input transformation logic
```go
func WriterTransform(input string) string
func EditorTransform(input string) string
func PublisherTransform(input string) string
```

**handlers.go** - Stream chunk processing
```go
type StreamHandler struct { /* ... */ }
func (h *StreamHandler) HandleAgentStart(chunk vnext.StreamChunk)
func (h *StreamHandler) HandleContent(chunk vnext.StreamChunk)
func (h *StreamHandler) HandleAgentComplete(chunk vnext.StreamChunk)
```

#### `infrastructure/` Package - Communication Layer

**websocket.go** - WebSocket server implementation
- Client connection management
- Message broadcasting
- CORS configuration

**messages.go** - Protocol definitions
```go
type MessageType string  // welcome, agent_start, agent_progress, etc.
type WSMessage struct { Type, Content, Agent, Timestamp, Metadata }
type AgentInfo struct { Name, DisplayName, Icon, Color, Description }
```

#### `config/` Package - Configuration Management

**config.go** - Environment-based settings
```go
type Config struct {
    Provider string  // huggingface, openrouter, openai
    Model    string  // Qwen/Qwen2.5-72B-Instruct
    APIKey   string  // API authentication
}
```

### Data Flow

```
User Input (Frontend)
    ↓ WebSocket
WebSocket Server
    ↓ workflow.Execute()
Story Pipeline (SubWorkflow Sequential)
    ↓ Step 1
Revision Loop (SubWorkflow Loop)
    ├─► Writer Agent → LLM → StreamChunks
    └─► Editor Agent → LLM → StreamChunks
    ↓ Step 2
Publisher Agent → LLM → StreamChunks
    ↓ StreamHandler
WebSocket Messages
    ↓ WebSocket
Frontend UI (Agent Bubbles)
```

### Key Technical Patterns

**1. SubWorkflow Composition**
```go
// Wrap Loop as Agent
revisionLoop := vnext.NewSubWorkflowAgent(loopWorkflow, "revisions")

// Wrap Sequential as Agent
storyPipeline := vnext.NewSubWorkflowAgent(sequentialWorkflow, "story_pipeline")
```

**2. Conditional Loop Exit**
```go
loopWorkflow := vnext.NewLoopWorkflowWithCondition(
    "revision_loop",
    []vnext.Agent{writer, editor},
    3, // max iterations
    vnext.OutputContains("APPROVED"), // exit condition
)
```

**3. Streaming Lifecycle**
```go
for chunk := range stream.Chunks() {
    switch chunk.Type {
    case vnext.ChunkTypeAgentStart:
        // Emit agent_start message
    case vnext.ChunkTypeText:
        // Emit agent_progress message
    case vnext.ChunkTypeAgentComplete:
        // Emit agent_complete message
    }
}
```

## 📖 Documentation

### Comprehensive Guides

**[ARCHITECTURE.md](ARCHITECTURE.md)** - Complete architecture documentation
- System overview with diagrams
- WebSocket protocol specification (8 message types)
- Component integration patterns
- Message flow with sequence diagrams
- Implementation guide for PM/Architects/Engineers
- Frontend integration (React/TypeScript)
- Best practices and troubleshooting

**This README** - Quick start and feature overview

### Learning Path

1. **Beginners**: Read "Overview" and "Quick Start" sections above
2. **Developers**: Study "Architecture" section and code comments in `workflow/`
3. **Architects**: Read `ARCHITECTURE.md` for complete system design
4. **Implementers**: Use as scaffold for similar projects

## 🎓 What You'll Learn

This example demonstrates production-ready patterns for:

### AgenticGoKit Framework
- **SubWorkflow Composition**: Building complex hierarchies
- **Conditional Loops**: Dynamic termination logic
- **Streaming Infrastructure**: Lifecycle events and metadata
- **Memory Integration**: Context retrieval across iterations
- **Agent Configuration**: Temperature, tokens, timeouts

### System Architecture
- **WebSocket Protocol Design**: Message types and flow
- **Real-time UI Updates**: Token-by-token streaming
- **Modular Code Organization**: ~150 lines per file
- **Error Handling**: Graceful degradation patterns
- **State Management**: Backend workflow ↔ Frontend UI

### Best Practices
- **Separation of Concerns**: Infrastructure vs workflow vs config
- **Transform Functions**: Runtime prompt modification
- **Stream Handlers**: Centralized chunk processing
- **Agent Attribution**: Proper nested agent tracking
- **Clean APIs**: Simple interfaces, complex implementations

## 🔧 Customization Guide

### Creating Your Own Workflow

1. **Define Agents** (workflow/agents.go)
   ```go
   func CreateMyAgent(cfg *config.Config) (vnext.Agent, error) {
       return vnext.QuickChatAgentWithConfig("MyAgent", &vnext.Config{
           SystemPrompt: MyAgentPrompt,
           Streaming: &vnext.StreamingConfig{Enabled: true},
           LLM: vnext.LLMConfig{
               Provider: cfg.Provider,
               Model: cfg.Model,
               Temperature: 0.5,
           },
       })
   }
   ```

2. **Compose Workflow** (workflow/story_workflow.go)
   ```go
   workflow := vnext.NewSequentialWorkflow("my_pipeline", []vnext.Agent{
       agent1, agent2, agent3,
   })
   ```

3. **Handle Streams** (workflow/handlers.go)
   - Map chunk types to WebSocket messages
   - Track agent state transitions
   - Accumulate content per agent

4. **Update Frontend** (frontend/src/)
   - Add agent info to `GetAgents()`
   - Customize message bubbles if needed
   - Adjust status panel colors/icons

## 🐛 Troubleshooting

### Backend Issues

**Problem**: `LLM API connection failed`
- **Solution**: Check API key in `.env` or environment variables
- **Verify**: Test API key with provider's playground

**Problem**: Agents having conversations instead of executing tasks
- **Solution**: Lower temperature (0.1-0.3), add enforcement suffixes in transforms

**Problem**: Loop never exits
- **Solution**: Verify condition function checks correct output pattern

### Frontend Issues

**Problem**: WebSocket connection refused
- **Solution**: Ensure backend running on port 8080
- **Check**: CORS settings in `websocket.go` match frontend origin

**Problem**: Messages appearing in wrong bubbles
- **Solution**: Verify `step_name` metadata preserved in `workflow.go:753`

**Problem**: Status panel not updating
- **Solution**: Check `HandleAgentStart/Complete` emit proper messages

## 📊 Performance Metrics

Typical execution for "Write story about cow" prompt:

| Metric | Value |
|--------|-------|
| Total Duration | 15-20 seconds |
| Writer (initial) | 5-7 seconds |
| Editor (first review) | 3-4 seconds |
| Writer (revision) | 4-5 seconds |
| Editor (approval) | 2-3 seconds |
| Publisher | 3-4 seconds |
| Total Tokens | 1200-1500 |
| Loop Iterations | 2 (typical) |

## 🚀 Production Deployment

### Backend Considerations
- Use environment variables for secrets
- Enable HTTPS for WebSocket (wss://)
- Implement rate limiting per user
- Add monitoring (Prometheus/Grafana)
- Use connection pooling for LLM APIs
- Implement proper logging (structured logs)

### Frontend Considerations
- Build optimized bundle: `npm run build`
- Serve via CDN (Cloudflare, Vercel)
- Use environment-specific WebSocket URLs
- Add error boundaries
- Implement reconnection logic with exponential backoff

### Infrastructure
- Load balancer with sticky sessions (WebSocket)
- Redis for session management (multi-server)
- Container orchestration (Docker/Kubernetes)
- Auto-scaling based on WebSocket connections

## 📝 License

Part of the AgenticGoKit project - see main repository for license details.

## 🤝 Contributing

Contributions welcome! Areas for improvement:
- Additional workflow patterns (Parallel, DAG)
- More sophisticated UI components
- Additional LLM provider examples
- Performance optimizations
- Testing utilities

---

**Part of [AgenticGoKit](https://github.com/kunalkushwaha/agenticgokit)** - Build sophisticated AI agent workflows in Go

**Advanced Features**: SubWorkflows • Conditional Loops • Nested Streaming • Real-time UI
