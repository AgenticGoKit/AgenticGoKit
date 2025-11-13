# Simple Sequential Workflow Demo

A minimal 2-agent sequential workflow demonstrating how to reuse AgenticGoKit's WebSocket infrastructure for rapid application development.

## 🎯 Overview

This example shows how easy it is to create a new agent-based application by reusing infrastructure. It features a simple sequential workflow: **Researcher → Writer**.

### Workflow

```
User Input (topic)
    ↓
Researcher (gathers facts & insights)
    ↓
Writer (creates engaging article)
    ↓
Final Output
```

## ✨ Key Learning Points

1. **Reusability**: Only 3 files (~315 lines) copied from `story-writer-chat`
2. **Simplicity**: Just 2 small application-specific files (~310 lines total)
3. **Zero Infrastructure Code**: No WebSocket, session, or UI code needed
4. **Focus on Logic**: Write only your workflow, everything else is handled
5. **Dynamic UI**: Frontend automatically adapts to your agents - no frontend code changes needed!

## 🏗️ What Was Reused

From `../story-writer-chat`, we copied these **generic, reusable** files:

| File | Lines | Purpose |
|------|-------|---------|
| `workflow_interface.go` | ~20 | Interface definition |
| `websocket_server.go` | ~210 | Complete WebSocket server |
| `session_manager.go` | ~85 | Session management |
| **Total** | **~315** | **Complete infrastructure!** |

## 📝 What We Created

Only 2 application-specific files:

| File | Lines | Purpose |
|------|-------|---------|
| `workflow.go` | ~240 | Our 2-agent sequential logic |
| `main.go` | ~70 | Entry point |
| **Total** | **~310** | **Application logic only!** |

### workflow.go Overview

```go
type SimpleSequentialWorkflow struct {
    researcher vnext.Agent
    writer     vnext.Agent
}

// Implements WorkflowExecutor interface
func (ssw *SimpleSequentialWorkflow) GetAgents() []AgentInfo {
    return []AgentInfo{
        {Name: "researcher", DisplayName: "Researcher", Icon: "🔍", Color: "blue", Description: "Gathers facts and insights"},
        {Name: "writer", DisplayName: "Writer", Icon: "✍️", Color: "purple", Description: "Creates engaging article"},
    }
}

func (ssw *SimpleSequentialWorkflow) Execute(ctx context.Context, userInput string, sendMessage MessageSender) error {
    // Phase 1: Research
    researchNotes := ssw.runAgentWithStreaming(ctx, ssw.researcher, ...)
    
    // Phase 2: Write
    article := ssw.runAgentWithStreaming(ctx, ssw.writer, researchNotes, ...)
    
    return nil
}
```

That's it! The WebSocket server sends your agent configuration to the frontend automatically. The UI builds itself based on your `GetAgents()` response.

## 🚀 Quick Start

### Prerequisites

1. **Go 1.21+** and **Node.js 18+**
2. **OpenRouter API Key**:
   ```bash
   # Windows PowerShell
   $env:OPENROUTER_API_KEY="your-api-key-here"
   
   # Linux/Mac
   export OPENROUTER_API_KEY="your-api-key-here"
   ```

### Running

```bash
# 1. Start backend (runs on port 8081)
go run .

# 2. In another terminal, set up frontend
cp -r ../story-writer-chat/frontend .
cd frontend
npm install  # first time only

# The frontend already has the correct WebSocket URL (port 8081)
# and will automatically display your 2 agents (Researcher, Writer)

npm run dev

# 3. Open browser
# http://localhost:5173
```

## 📖 Usage

### Example Topics

Try these prompts:
- "Artificial Intelligence"
- "Climate Change Solutions"
- "The History of Chess"
- "Quantum Computing Basics"
- "The Future of Space Exploration"

### What Happens

```
You: "Artificial Intelligence"

Researcher 🔍: [Researches and provides key facts, statistics, insights...]
Writer ✍️: [Transforms notes into an engaging, well-structured article...]

Final: [Complete article about AI ready to read!]
```

## 🔧 Customization

### Change Agent Behavior

Edit `workflow.go`:

```go
researcher, err := vnext.QuickChatAgentWithConfig("Researcher", &vnext.Config{
    SystemPrompt: "Your custom research instructions...",
    LLM: vnext.LLMConfig{
        Model:       "openai/gpt-4o",  // Use a different model
        Temperature: 0.5,               // Adjust creativity
        MaxTokens:   1000,              // Change length
    },
})
```

### Add More Agents

Extend the workflow:

```go
type EnhancedWorkflow struct {
    researcher vnext.Agent
    analyzer   vnext.Agent  // New agent
    writer     vnext.Agent
}

func (ew *EnhancedWorkflow) Execute(...) error {
    // Research → Analyze → Write
    notes := runAgent(researcher)
    insights := runAgent(analyzer, notes)
    article := runAgent(writer, insights)
    return nil
}
```

## 📊 Comparison with story-writer-chat

| Aspect | story-writer-chat | sequential-workflow-demo |
|--------|-------------------|--------------------------|
| **Agents** | 3 (Writer, Editor, Publisher) | 2 (Researcher, Writer) |
| **Flow** | Iterative with feedback loop | Simple sequential |
| **Logic Lines** | ~420 | ~240 |
| **Complexity** | Conditional branching, revision cycles | Straightforward A→B |
| **Infrastructure** | Same (~315 lines) | Same (~315 lines) |
| **UI** | React frontend | Same React frontend |

**Both share 100% of infrastructure code!**

## 🎓 Learning Path

### For Beginners

Start with this example:
1. Understand the simple sequential flow
2. See how `WorkflowExecutor` interface works
3. Modify agent prompts to see effects
4. Add a third agent to the sequence

### Intermediate

Move to `story-writer-chat`:
1. Learn iterative workflows with feedback loops
2. Understand conditional branching (approved vs needs_revision)
3. See how to handle revision cycles
4. Implement quality gates

### Advanced

Create your own:
1. Copy the 3 reusable files
2. Design your workflow (parallel? conditional? mixed?)
3. Implement `WorkflowExecutor` interface
4. Run and iterate!

## 🔍 How It Works

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    FRONTEND (React + TypeScript)                │
│                                                                  │
│  No Hardcoded Configuration! Everything Dynamic!                │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  App.tsx                                                 │  │
│  │  ├─ Title: {workflowName} ← From Server                 │  │
│  │  ├─ WorkflowVisualizer                                   │  │
│  │  │  └─ agents.map(agent => <AgentCard {...agent} />)    │  │
│  │  └─ ChatInterface                                        │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Store (chat-store.ts)                                   │  │
│  │  ├─ workflowName: "Research & Write" ← Dynamic          │  │
│  │  ├─ agents: [                         ← Dynamic          │  │
│  │  │     {name: "researcher", icon: "🔍", color: "blue"}  │  │
│  │  │     {name: "writer", icon: "✍️", color: "purple"}    │  │
│  │  │   ]                                                   │  │
│  │  └─ messages, workflowState, etc.                       │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              ↕ WebSocket
┌─────────────────────────────────────────────────────────────────┐
│                          BACKEND (Go)                           │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  websocket_server.go (REUSABLE) ⭐                       │  │
│  │                                                           │  │
│  │  On Connection:                                          │  │
│  │  1. Send session_created + welcome message              │  │
│  │  2. Send agent_config:                                   │  │
│  │     {                                                    │  │
│  │       "workflow_name": workflow.Name(),  ← Dynamic      │  │
│  │       "agents": workflow.GetAgents()     ← Dynamic      │  │
│  │     }                                                    │  │
│  │  3. Route user messages to workflow.Execute()           │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  workflow.go (APPLICATION-SPECIFIC)                      │  │
│  │                                                           │  │
│  │  func (w *Workflow) Name() string {                     │  │
│  │    return "Research & Write"    ← Shown in UI           │  │
│  │  }                                                       │  │
│  │                                                           │  │
│  │  func (w *Workflow) GetAgents() []AgentInfo {           │  │
│  │    return []AgentInfo{                                   │  │
│  │      {Name: "researcher", Icon: "🔍", Color: "blue"},   │  │
│  │      {Name: "writer", Icon: "✍️", Color: "purple"},     │  │
│  │    }                              ← Builds UI cards      │  │
│  │  }                                                       │  │
│  │                                                           │  │
│  │  func (w *Workflow) Execute(...) {                      │  │
│  │    // Researcher → Writer (Your logic)                  │  │
│  │  }                                                       │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### The Magic of Reusable Infrastructure

**WebSocket Server** (`websocket_server.go`):
```go
type WebSocketServer struct {
    workflow WorkflowExecutor  // ← ANY workflow!
    // ... other fields
}

func (s *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    // 1. Upgrade connection
    conn, _ := s.upgrader.Upgrade(w, r, nil)
    
    // 2. Create session
    session := s.sessionManager.CreateSession()
    
    // 3. Send configuration (AUTOMATIC!)
    s.sendWSMessage(conn, WSMessage{
        Type: "agent_config",
        Metadata: map[string]interface{}{
            "workflow_name": s.workflow.Name(),      // ← Your workflow name
            "agents":        s.workflow.GetAgents(), // ← Your agent metadata
        },
    })
    
    // 4. Handle messages
    for {
        var msg WSMessage
        conn.ReadJSON(&msg)
        s.workflow.Execute(...)  // ← Calls YOUR workflow
    }
}
```

Your workflow just needs to implement:
```go
type WorkflowExecutor interface {
    Name() string              // "Research & Write"
    WelcomeMessage() string    // Welcome text
    GetAgents() []AgentInfo    // Agent metadata for UI
    Execute(...)               // Your logic here
    Cleanup()                  // Optional cleanup
}
```

### Dynamic Configuration Flow

**1. Backend Sends Configuration:**

```json
// Sent automatically when WebSocket connects
{
  "type": "agent_config",
  "timestamp": 1699876543,
  "metadata": {
    "workflow_name": "Research & Write",
    "agents": [
      {
        "name": "researcher",
        "displayName": "Researcher",
        "icon": "🔍",
        "color": "blue",
        "description": "Gathers facts and insights"
      },
      {
        "name": "writer",
        "displayName": "Writer",
        "icon": "✍️",
        "color": "purple",
        "description": "Creates engaging article"
      }
    ]
  }
}
```

**2. Frontend Receives & Applies:**

```typescript
// In chat-store.ts (automatic handling)
case 'agent_config':
  if (message.metadata?.workflow_name) {
    set({ workflowName: message.metadata.workflow_name })
  }
  if (message.metadata?.agents) {
    set({ agents: message.metadata.agents })
  }
  break
```

**3. UI Builds Itself:**

```tsx
// In App.tsx
<h1>{workflowName}</h1>  {/* Shows "Research & Write" */}

// In WorkflowVisualizer.tsx
{agents.map(agent => (
  <AgentCard 
    key={agent.name}
    icon={agent.icon}        // 🔍 or ✍️
    name={agent.displayName} // "Researcher" or "Writer"
    color={agent.color}      // blue or purple
    description={agent.description}
  />
))}
```

### Message Protocol

Your workflow communicates with the UI via simple messages:

```go
// Starting an agent
sendMessage(WSMessage{
    Type:      "agent_start",
    Agent:     "researcher",
    Content:   "Gathering information...",
    Progress:  0,
    Timestamp: float64(time.Now().Unix()),
})

// Streaming content
sendMessage(WSMessage{
    Type:      "agent_progress",
    Agent:     "researcher",
    Content:   "chunk of text...",
    Timestamp: float64(time.Now().Unix()),
})

// Agent complete
sendMessage(WSMessage{
    Type:      "agent_complete",
    Agent:     "researcher",
    Progress:  100,
    Timestamp: float64(time.Now().Unix()),
})
```

The WebSocket server and React UI handle everything else!

### Complete Message Types

| Type | Direction | Purpose | Data |
|------|-----------|---------|------|
| `session_created` | Backend → Frontend | Connection established | Session ID, welcome message |
| `agent_config` | Backend → Frontend | **Send configuration** | **Workflow name, agents metadata** |
| `user_message` | Frontend → Backend | User input | Message content |
| `workflow_start` | Backend → Frontend | Workflow begins | Status message |
| `agent_start` | Backend → Frontend | Agent starts work | Agent name, progress |
| `agent_progress` | Backend → Frontend | Streaming output | Agent name, content chunk |
| `agent_complete` | Backend → Frontend | Agent finishes | Agent name, final progress |
| `workflow_done` | Backend → Frontend | Workflow complete | Success message |
| `error` | Backend → Frontend | Error occurred | Error details |

### Files and Their Roles

#### Reusable Infrastructure (Copy These) ⭐

1. **`workflow_interface.go`** (~30 lines)
   ```go
   // Defines AgentInfo struct and WorkflowExecutor interface
   // This is the contract your workflow must implement
   type AgentInfo struct {
       Name        string `json:"name"`
       DisplayName string `json:"displayName"`
       Icon        string `json:"icon"`
       Color       string `json:"color"`
       Description string `json:"description"`
   }
   ```

2. **`websocket_server.go`** (~210 lines)
   - Handles WebSocket connections
   - Automatically sends agent_config with workflow.Name() and workflow.GetAgents()
   - Routes messages between frontend and workflow
   - Manages sessions
   - **Zero workflow-specific logic**

3. **`session_manager.go`** (~85 lines)
   - Thread-safe session storage
   - Message history per session
   - Session lifecycle management

#### Application-Specific (You Write This)

4. **`workflow.go`** (~240 lines for this demo)
   ```go
   type SimpleSequentialWorkflow struct {
       researcher vnext.Agent
       writer     vnext.Agent
   }
   
   func (ssw *SimpleSequentialWorkflow) Name() string {
       return "Research & Write"  // ← Shown in header
   }
   
   func (ssw *SimpleSequentialWorkflow) GetAgents() []AgentInfo {
       return []AgentInfo{
           {
               Name:        "researcher",
               DisplayName: "Researcher",
               Icon:        "🔍",
               Color:       "blue",
               Description: "Gathers facts and insights",
           },
           {
               Name:        "writer",
               DisplayName: "Writer",
               Icon:        "✍️",
               Color:       "purple",
               Description: "Creates engaging article",
           },
       }
   }
   
   func (ssw *SimpleSequentialWorkflow) Execute(ctx context.Context, userInput string, sendMessage MessageSender) error {
       // Phase 1: Research
       sendMessage(WSMessage{Type: "agent_start", Agent: "researcher"})
       researchNotes := runResearcher(...)
       sendMessage(WSMessage{Type: "agent_complete", Agent: "researcher"})
       
       // Phase 2: Write
       sendMessage(WSMessage{Type: "agent_start", Agent: "writer"})
       article := runWriter(researchNotes, ...)
       sendMessage(WSMessage{Type: "agent_complete", Agent: "writer"})
       
       return nil
   }
   ```

5. **`main.go`** (~70 lines)
   ```go
   func main() {
       apiKey := os.Getenv("OPENROUTER_API_KEY")
       workflow := NewSimpleSequentialWorkflow(apiKey)
       server := NewWebSocketServer("8081", workflow)
       server.Start()
   }
   ```

#### Frontend (Copy Entire Folder)

6. **`frontend/src/`**
   - **App.tsx**: Main layout, uses `workflowName` from store
   - **components/WorkflowVisualizer.tsx**: Maps over `agents` array from store
   - **store/chat-store.ts**: Handles `agent_config` messages, updates state
   - **types/index.ts**: TypeScript interfaces matching backend message types
   - **lib/websocket-client.ts**: WebSocket connection management

**Key Point**: The entire frontend has ZERO hardcoded workflow or agent information. Everything comes from the `agent_config` message!

## 💡 Creating Your Own Workflow

### Step 1: Copy Infrastructure

```bash
cp workflow_interface.go ../your-new-app/
cp websocket_server.go ../your-new-app/
cp session_manager.go ../your-new-app/
```

### Step 2: Create Your Workflow

```go
// your_workflow.go
type YourWorkflow struct {
    agent1 vnext.Agent
    agent2 vnext.Agent
}

func (yw *YourWorkflow) Name() string {
    return "Your App Name"
}

func (yw *YourWorkflow) Execute(ctx context.Context, input string, sendMessage MessageSender) error {
    // Your logic here
    result1 := runAgent(yw.agent1, input, sendMessage)
    result2 := runAgent(yw.agent2, result1, sendMessage)
    return nil
}
```

### Step 3: Wire It Up

```go
// main.go
func main() {
    workflow := NewYourWorkflow(apiKey)
    server := NewWebSocketServer("8080", workflow)
    server.Start()
}
```

### Step 4: Run!

```bash
go run .
```

The UI works automatically! 🎉

## 🎯 Example Use Cases

Using this pattern, you could quickly build:

- **Code Review Workflow**: Linter → SecurityChecker → Reviewer
- **Data Analysis**: Loader → Cleaner → Analyzer → Visualizer
- **Content Creation**: Outliner → Drafter → Editor → Formatter
- **Research Assistant**: Searcher → Summarizer → Organizer
- **RAG System**: Retriever → Ranker → Generator

## 📚 Next Steps

1. **Experiment**: Modify agent prompts and see results
2. **Extend**: Add a third agent (e.g., Fact-Checker between Researcher and Writer)
3. **Learn**: Study `story-writer-chat` for iterative workflows
4. **Create**: Build your own workflow for a different use case

## 💡 Tips

- Start simple (2-3 agents)
- Test agents individually before combining
- Use clear, specific agent system prompts
- Monitor agent outputs during development
- Keep workflows focused on a specific task

---

**Ready to build your own workflow? Just copy 3 files and start coding! 🚀**
