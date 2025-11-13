# Story Writer Chat App 📖

An AI-powered collaborative story writing application featuring a 3-agent iterative workflow with real-time streaming and modern React UI.

## 🎯 Overview

Story Writer demonstrates clean separation between workflow logic and UI infrastructure, making it easy to create new agent-based applications. It features three specialized AI agents working together with an iterative feedback loop to create, refine, and polish stories.

### The Writing Team

1. **Writer Agent 🖊️**
   - Creates initial story drafts with creativity and imagination
   - Revises based on editor feedback
   - Higher temperature (0.8) for creative outputs

2. **Editor Agent ✏️**
   - Reviews stories with high editorial standards
   - Always requests revision on first review (enforced)
   - Provides specific, actionable feedback
   - Approves when quality standards are met

3. **Publisher Agent 📚**
   - Formats the final approved story professionally
   - Adds titles and proper structure
   - Prepares publication-ready version

### Iterative Workflow

```
User Prompt
    ↓
Writer (creates draft)
    ↓
Editor (reviews) ──→ NEEDS_REVISION? ──→ Writer (revises)
    ↓                      ↑                     ↓
    ↓                      └─────────────────────┘
    ↓                    (max 2 revision cycles)
APPROVED?
    ↓
Publisher (formats)
    ↓
Final Story
```

## ✨ Features

### User Experience
- 💬 **Real-time Chat Interface** - Smooth, responsive design with instant feedback
- 📊 **Workflow Visualization** - See which agent is currently working
- 🎨 **Progress Tracking** - Visual progress bar showing workflow completion
- 💾 **Session Management** - Chat history persists across the session
- 📱 **Responsive Design** - Works great on desktop, tablet, and mobile
- ⚡ **WebSocket Streaming** - Real-time content streaming from agents
- 🔧 **Dynamic Agent Configuration** - Frontend automatically adapts to any workflow's agents

### Technical Features
- 🔄 **Iterative Workflow** - Writer-Editor feedback loop (max 2 revisions)
- 🎯 **vnext Framework** - Built on AgenticGoKit's vnext framework
- 🔌 **WebSocket Streaming** - Real-time content delivery
- 🏗️ **Modular Architecture** - Clean separation of concerns
- 🔁 **Reusable Infrastructure** - WebSocket server + UI work with any workflow
- 🛡️ **Error Handling** - Graceful error handling and recovery
- 🎨 **Agent Metadata System** - Server sends agent configuration (names, icons, colors, descriptions) to frontend

## 🏗️ Architecture

This example demonstrates **clean separation** between reusable infrastructure and application-specific logic, with a **dynamic configuration system** that allows the frontend to work with any workflow without code changes.

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         FRONTEND (React)                        │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  App.tsx - Main UI Layout (Dynamic Title)               │  │
│  │  ├─ WorkflowVisualizer - Agent Cards (Dynamic)          │  │
│  │  └─ ChatInterface - Message Display & Input             │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Store (Zustand) - State Management                     │  │
│  │  ├─ agents: Agent[] (populated from server)             │  │
│  │  ├─ workflowName: string (from server)                  │  │
│  │  ├─ messages: ChatMessage[]                             │  │
│  │  └─ workflowState: WorkflowState                        │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  WebSocketClient - Connection & Message Handler         │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              ↕ WebSocket
┌─────────────────────────────────────────────────────────────────┐
│                      BACKEND (Go + vnext)                       │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  main.go - Entry Point                                   │  │
│  │  ├─ Initialize workflow                                  │  │
│  │  ├─ Create WebSocket server                             │  │
│  │  └─ Start HTTP server                                   │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  websocket_server.go - Connection Handler (REUSABLE)    │  │
│  │  ├─ HandleWebSocket()                                    │  │
│  │  ├─ Send session_created + welcome message              │  │
│  │  ├─ Send agent_config (workflow name + agents)          │  │
│  │  └─ Route user messages to workflow                     │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  workflow_interface.go - Contract (REUSABLE)            │  │
│  │  ├─ Name() → workflow title                             │  │
│  │  ├─ WelcomeMessage() → greeting                          │  │
│  │  ├─ GetAgents() → []AgentInfo (metadata)                │  │
│  │  ├─ Execute() → run workflow                            │  │
│  │  └─ Cleanup() → cleanup resources                       │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  story_workflow.go - Application Logic (SPECIFIC)       │  │
│  │  ├─ StoryWriterWorkflow implementation                  │  │
│  │  ├─ Writer, Editor, Publisher agents                    │  │
│  │  ├─ Iterative revision loop                             │  │
│  │  └─ Returns agent metadata (icons, colors, etc.)        │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  session_manager.go - Session State (REUSABLE)          │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### Dynamic Configuration System

The key innovation is that **the frontend receives all configuration from the backend**, making it completely workflow-agnostic:

#### 1. Backend Provides Configuration

When a WebSocket connection is established, the backend sends:

```json
// Message 1: session_created
{
  "type": "session_created",
  "session_id": "abc123",
  "content": "Welcome to Story Writer! Tell me what kind of story you'd like me to create.",
  "timestamp": 1699876543
}

// Message 2: agent_config (AUTOMATIC)
{
  "type": "agent_config",
  "timestamp": 1699876543,
  "metadata": {
    "workflow_name": "Story Writer",
    "agents": [
      {
        "name": "writer",
        "displayName": "Writer",
        "icon": "✍️",
        "color": "blue",
        "description": "Creates initial story draft"
      },
      {
        "name": "editor",
        "displayName": "Editor",
        "icon": "✏️",
        "color": "green",
        "description": "Reviews and provides feedback"
      },
      {
        "name": "publisher",
        "displayName": "Publisher",
        "icon": "📚",
        "color": "purple",
        "description": "Formats final version"
      }
    ]
  }
}
```

#### 2. Frontend Receives & Applies Configuration

The frontend's `chat-store.ts` handles this automatically:

```typescript
case 'agent_config':
  // Update workflow name in header
  if (message.metadata?.workflow_name) {
    set({ workflowName: message.metadata.workflow_name })
  }
  // Update agent cards in sidebar
  if (message.metadata?.agents) {
    set({ agents: message.metadata.agents })
  }
  break
```

#### 3. UI Builds Itself Dynamically

- **Header**: Displays `workflowName` from server
- **Agent Cards**: Generated from `agents[]` array
- **Colors & Icons**: Applied based on agent metadata
- **Descriptions**: Shown in tooltips/cards

### File Structure

```
story-writer-chat/
├── Backend (Go)
│   ├── main.go                  # Entry point (~70 lines)
│   ├── workflow_interface.go    # Generic interface (REUSABLE) ⭐
│   ├── websocket_server.go      # WebSocket handler (REUSABLE) ⭐
│   ├── session_manager.go       # Session state (REUSABLE) ⭐
│   └── story_workflow.go        # Story-specific logic (APPLICATION-SPECIFIC)
│
└── Frontend (React + TypeScript)
    ├── src/
    │   ├── App.tsx                    # Main layout with dynamic title
    │   ├── components/
    │   │   ├── WorkflowVisualizer.tsx # Renders agents from store
    │   │   ├── AgentCard.tsx          # Individual agent display
    │   │   ├── ChatInterface.tsx      # Message history
    │   │   └── InputArea.tsx          # User input
    │   ├── store/
    │   │   └── chat-store.ts          # State management (receives config)
    │   ├── types/
    │   │   └── index.ts               # TypeScript interfaces
    │   └── lib/
    │       ├── websocket-client.ts    # WebSocket connection
    │       └── utils.ts               # Helpers
    └── package.json
```

### Reusable Components (Copy to Any Project) ⭐

**1. `workflow_interface.go`** - Defines the contract any workflow must implement:
```go
type AgentInfo struct {
    Name        string `json:"name"`        // Internal identifier
    DisplayName string `json:"displayName"` // UI display name
    Icon        string `json:"icon"`        // Emoji or icon
    Color       string `json:"color"`       // Theme color (blue, green, purple, etc.)
    Description string `json:"description"` // Brief role description
}

type WorkflowExecutor interface {
    Name() string              // Returns workflow name for UI header
    WelcomeMessage() string    // Initial greeting message
    GetAgents() []AgentInfo    // Returns agent metadata for UI
    Execute(ctx context.Context, userInput string, sendMessage MessageSender) error
    Cleanup(ctx context.Context) error
}
```

**2. `websocket_server.go`** - Generic WebSocket server that works with any `WorkflowExecutor`:
- Handles WebSocket connections and upgrades
- Creates sessions for each connection
- **Automatically sends workflow configuration** (name + agents) to frontend
- Routes messages between client and workflow
- Completely workflow-agnostic - no business logic

**Key Methods:**
```go
// Sends agent_config on connection
func (s *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    // ... connection setup ...
    
    // Send agent configuration automatically
    s.sendWSMessage(conn, WSMessage{
        Type: MsgTypeAgentConfig,
        Metadata: map[string]interface{}{
            "workflow_name": s.workflow.Name(),      // ← Dynamic name
            "agents":        s.workflow.GetAgents(), // ← Dynamic agents
        },
    })
}
```

**3. `session_manager.go`** - Thread-safe session and message history management
- Stores chat history per session
- Thread-safe concurrent access
- Message persistence across reconnects

**4. `frontend/`** - React + TypeScript UI that works with any workflow:
- **Zero hardcoded configuration** - everything from server
- Automatically builds UI based on received metadata
- Adapts to any number of agents (1, 2, 3, or more)
- Color themes applied dynamically
- Truly generic - copy & paste to any workflow project

### Application-Specific Logic

**`story_workflow.go`** - The ONLY file containing story-specific logic:

```go
type StoryWriterWorkflow struct {
    writer       vnext.Agent
    editor       vnext.Agent
    publisher    vnext.Agent
    maxRevisions int
}

// Implements WorkflowExecutor interface
func (sw *StoryWriterWorkflow) Name() string {
    return "Story Writer"  // ← Shown in UI header
}

func (sw *StoryWriterWorkflow) GetAgents() []AgentInfo {
    return []AgentInfo{
        {
            Name:        "writer",
            DisplayName: "Writer",
            Icon:        "✍️",
            Color:       "blue",
            Description: "Creates initial story draft",
        },
        {
            Name:        "editor",
            DisplayName: "Editor",
            Icon:        "✏️",
            Color:       "green",
            Description: "Reviews and provides feedback",
        },
        {
            Name:        "publisher",
            DisplayName: "Publisher",
            Icon:        "📚",
            Color:       "purple",
            Description: "Formats final version",
        },
    }
}

func (sw *StoryWriterWorkflow) Execute(ctx context.Context, userInput string, sendMessage MessageSender) error {
    // Story-specific iterative workflow logic here
    // Writer → Editor → (revisions) → Publisher
}
```

### WebSocket Message Protocol

The system uses a well-defined message protocol for communication:

#### Backend → Frontend Messages

| Message Type | Purpose | Contains |
|--------------|---------|----------|
| `session_created` | Connection established | Session ID, welcome message |
| `agent_config` | **Configuration data** | **Workflow name, agent metadata** |
| `workflow_start` | Workflow begins | Optional status message |
| `agent_start` | Agent begins work | Agent name, progress % |
| `agent_progress` | Streaming content | Agent name, partial content |
| `agent_complete` | Agent finishes | Agent name, final status |
| `workflow_done` | Workflow complete | Success message |
| `error` | Error occurred | Error message |

#### Frontend → Backend Messages

| Message Type | Purpose | Contains |
|--------------|---------|----------|
| `user_message` | User input | Message content, timestamp |

#### Example Message Flow

```
1. Connection Established
   → Backend: session_created
   → Backend: agent_config (with workflow name + agents)
   
2. User Sends Message
   ← Frontend: user_message
   
3. Workflow Processes
   → Backend: workflow_start
   → Backend: agent_start (writer)
   → Backend: agent_progress (streaming text...)
   → Backend: agent_progress (more text...)
   → Backend: agent_complete (writer done)
   → Backend: agent_start (editor)
   → Backend: agent_progress (streaming review...)
   → Backend: agent_complete (editor done)
   → Backend: agent_start (publisher)
   → Backend: agent_progress (streaming final version...)
   → Backend: agent_complete (publisher done)
   → Backend: workflow_done
```

### Benefits of This Architecture

✅ **True Reusability**: Copy 3 backend files + 1 frontend folder = works immediately  
✅ **Zero Configuration**: Frontend has no workflow-specific code  
✅ **Type Safety**: Full TypeScript support throughout  
✅ **Scalability**: Add agents without changing infrastructure  
✅ **Maintainability**: Clear boundaries between generic and specific code  
✅ **Flexibility**: Easy to create new workflows  

### Creating a New Workflow

To create a different workflow (e.g., data analyzer, code reviewer):

1. **Copy reusable backend files:**
   ```bash
   cp workflow_interface.go ../my-workflow/
   cp websocket_server.go ../my-workflow/
   cp session_manager.go ../my-workflow/
   ```

2. **Create your workflow implementation:**
   ```go
   type MyWorkflow struct {
       agent1 vnext.Agent
       agent2 vnext.Agent
   }
   
   func (mw *MyWorkflow) Name() string {
       return "My Custom Workflow"  // ← Shows in UI
   }
   
   func (mw *MyWorkflow) GetAgents() []AgentInfo {
       return []AgentInfo{
           {
               Name:        "analyzer",
               DisplayName: "Data Analyzer",
               Icon:        "🔍",
               Color:       "blue",
               Description: "Analyzes input data",
           },
           {
               Name:        "reporter",
               DisplayName: "Report Generator",
               Icon:        "📊",
               Color:       "green",
               Description: "Creates detailed reports",
           },
       }
   }
   
   func (mw *MyWorkflow) Execute(...) error {
       // Your workflow logic here
   }
   ```

3. **Create minimal main.go:**
   ```go
   func main() {
       workflow := NewMyWorkflow()
       server := NewWebSocketServer("8080", workflow)
       server.Start()
   }
   ```

4. **Copy the frontend folder:**
   ```bash
   cp -r frontend ../my-workflow/frontend
   ```

5. **Done!** Run both backend and frontend - the UI automatically shows:
   - Your workflow name in the header
   - Your custom agents with their icons and colors
   - Dynamic progress tracking for your agents

**No frontend code changes needed!** The UI builds itself from the metadata.

See `../sequential-workflow-demo` for a complete working example!

## 🚀 Quick Start

### Prerequisites

1. **Go 1.21 or higher**
   ```bash
   go version
   ```

2. **Node.js 18 or higher** (for the React frontend)
   ```bash
   node --version
   npm --version
   ```

3. **OpenRouter API Key**
   - Sign up at https://openrouter.ai
   - Get your API key from the dashboard
   - Copy `.env.example` to `.env` and add your key:
   ```bash
   # Copy the example file
   cp .env.example .env
   
   # Edit .env and set your key
   # OPENROUTER_API_KEY=your-api-key-here
   ```
   
   Or set it as an environment variable:
   ```bash
   # Windows PowerShell
   $env:OPENROUTER_API_KEY="your-api-key-here"
   
   # Linux/Mac
   export OPENROUTER_API_KEY="your-api-key-here"
   ```

### Running the Application

1. **Navigate to the project directory:**
   ```bash
   cd examples/vnext/story-writer-chat
   ```

2. **Start the backend server:**
   ```bash
   go run .
   ```
   
   The backend will start on `http://localhost:8080` with:
   - WebSocket API at `ws://localhost:8080/ws`
   - Health check at `/api/health`
   - Sessions API at `/api/sessions`

3. **In a new terminal, start the frontend:**
   ```bash
   cd frontend
   npm install  # First time only
   npm run dev
   ```

4. **Open your browser:**
   ```
   http://localhost:5173
   ```

The application will:
- ✅ Connect to the backend WebSocket server
- 🚀 Display the React-based UI
- 📝 Initialize all three agents (Writer, Editor, Publisher)
- 💬 Open the chat interface ready for story prompts

## 📖 Usage Guide

### Creating a Story

1. **Enter your story prompt** - Be creative! For example:
   - "A robot discovers emotions for the first time"
   - "A time traveler accidentally changes history"
   - "A detective cat solves mysteries in Tokyo"

2. **Watch the iterative workflow**:
   - 🖊️ Writer creates initial draft (may have intentional typos)
   - ✏️ Editor reviews and requests revision (always on first review)
   - 🖊️ Writer revises based on feedback
   - ✏️ Editor reviews again (approves if improved, or cycles again)
   - 📚 Publisher formats the final approved version

3. **View real-time progress** - Each agent's work streams to the UI as it happens

### Example Interaction

```
You: Write a story about a lonely lighthouse keeper who befriends a talking seagull

Writer 🖊️: [Creates imaginative draft with intentional typos...]
Editor ✏️: NEEDS_REVISION: Fix spelling errors, strengthen character development...
Writer 🖊️: [Revises with improvements...]
Editor ✏️: APPROVED: Much better! Grammar fixed, characters more vivid...
Publisher 📚: [Formats with title and professional structure...]

Final Story: "The Keeper and the Gull" - [Beautiful, polished story!]
```

## 🏗️ Architecture

### Backend (main.go)

```
Story Writer Workflow
│
├── StoryWriterWorkflow
│   ├── Writer Agent (gemma2:2b, temp: 0.8)
│   ├── Editor Agent (gemma2:2b, temp: 0.6)
│   └── Publisher Agent (gemma2:2b, temp: 0.4)
│
├── WebSocket Server
│   ├── Connection Management
│   ├── Session Management
│   └── Message Routing
│
└── HTTP Endpoints
    ├── /ws (WebSocket)
    ├── /api/health
    └── /api/sessions
```

### Frontend (static/)

```
Static Assets
│
├── index.html
│   ├── Header with status
│   ├── Workflow visualization
│   ├── Chat interface
│   └── Input area
│
├── style.css
│   ├── Modern responsive design
│   ├── Smooth animations
│   └── Agent-specific styling
│
└── app.js
    ├── WebSocket client
    ├── Message handling
    ├── UI updates
    └── Workflow tracking
```

## 🔧 Configuration

### Customizing Agents

Edit `story_workflow.go` to modify agent behavior:

```go
writer, err := vnext.QuickChatAgentWithConfig("Writer", &vnext.Config{
    Name: "writer",
    SystemPrompt: "Your custom prompt here...",
    Timeout: 90 * time.Second,
    LLM: vnext.LLMConfig{
        Provider:    "openrouter",
        Model:       "openai/gpt-4o-mini",  // Change model
        Temperature: 0.8,                    // Adjust creativity
        MaxTokens:   800,                    // Control length
        APIKey:      apiKey,
    },
})
```

### Adjusting Revision Cycles

In `story_workflow.go`:
```go
return &StoryWriterWorkflow{
    writer:       writer,
    editor:       editor,
    publisher:    publisher,
    maxRevisions: 3,  // Change from 2 to allow more iterations
}
```

### Changing the Port

```bash
PORT=3000 go run .
```

Or modify in `main.go`:
```go
func getPort() string {
    port := "3000"  // Change default port
    if envPort := os.Getenv("PORT"); envPort != "" {
        port = envPort
    }
    return port
}
```

## 🔍 Troubleshooting

### OpenRouter Connection Failed
- Verify your API key is set correctly
- Check your internet connection
- Ensure you have credits in your OpenRouter account
- Try the test connection in startup logs

### WebSocket Connection Issues
- Check browser console for errors
- Verify firewall settings
- Ensure port 8080 is available
- Try a different browser

### Agents Not Responding
- Verify OpenRouter API key is valid
- Check server logs for errors
- Ensure you have API credits
- Try with a simpler prompt first

## 📚 Learn More

- [AgenticGoKit vnext Documentation](../../core/vnext/README.md)
- [Sequential Workflow Demo](../sequential-workflow-demo/) - Simpler 2-agent example
- [Workflow Interface](./workflow_interface.go) - See how to implement your own workflow

## 🎯 Related Examples

- **sequential-workflow-demo** - Simple 2-agent workflow (Researcher → Writer)
- Both examples share the same reusable infrastructure!

## 💡 Tips for Best Results

1. **Be Specific** - Detailed prompts yield better stories
2. **Set the Scene** - Provide context (time, place, characters)
3. **Add Constraints** - Specify tone, length, or style ("write a funny story about...")
4. **Watch the Revision Process** - See how the editor improves the initial draft
5. **Try Different Genres** - Sci-fi, fantasy, mystery, romance, horror

## 📝 Example Prompts

### Short Stories
- "A chef who can taste emotions in food"
- "The last bookstore in a digital world"
- "A painter who can step into their artwork"

### Flash Fiction
- "In three sentences: A door that shouldn't be opened"
- "Sudden twist: A normal day goes unexpectedly wrong"

### Specific Genres
- "Sci-fi: First contact with aliens who communicate through music"
- "Mystery: A detective investigates their own future murder"
- "Fantasy: A dragon who's afraid of heights"

---

**Happy Story Writing! 📖✨**
