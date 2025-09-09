# WebUI Chat Demo

This example demonstrates the AgenticGoKit WebUI chat interface with enhanced agents and comprehensive chat functionality.

## Features

- 🤖 **4 Enhanced Agents**: Assistant, Coder, Writer, and Analyst
- 💬 **Full Chat Interface**: Modern web-based chat UI
- 🔄 **Real-time Communication**: WebSocket support for live updates
- 📊 **Session Management**: Enhanced session handling with persistence
- 🎨 **Rich Responses**: Contextual, personality-driven agent responses
- 📱 **Responsive Design**: Works on desktop and mobile devices

## Quick Start

1. Navigate to this directory:
   ```bash
   cd examples/webui_chat_demo
   ```

2. Run the demo:
   ```bash
   go run main.go
   ```

3. Open your browser to: http://localhost:8080

## Available Agents

### 🤖 Assistant
- **Role**: General Assistant
- **Personality**: Friendly and knowledgeable
- **Capabilities**: General assistance, Q&A, information lookup, problem-solving, learning support

### 💻 Coder  
- **Role**: Code Specialist
- **Personality**: Technical and precise
- **Capabilities**: Code review, debugging, programming help, best practices, architecture design, performance optimization

### ✍️ Writer
- **Role**: Content Creator
- **Personality**: Creative and articulate  
- **Capabilities**: Content creation, editing, creative writing, grammar check, storytelling, copywriting

### 📊 Analyst
- **Role**: Data Specialist
- **Personality**: Analytical and detail-oriented
- **Capabilities**: Data analysis, statistical modeling, visualization, research, reporting, insights generation

## API Endpoints

- `GET /` - Main chat interface
- `POST /api/chat` - Send message to agent
- `GET /api/agents` - List available agents
- `GET /api/health` - Health check
- `WS /ws` - WebSocket connection
- `GET /api/sessions` - List chat sessions

## Usage Examples

### Web Interface
1. Open http://localhost:8080
2. Select an agent from the sidebar
3. Start chatting!

### API Testing
```bash
# List all agents
curl http://localhost:8080/api/agents

# Chat with the assistant
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"agent_name": "assistant", "message": "Hello!"}'

# Chat with the coder about Python
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"agent_name": "coder", "message": "Help me optimize this Python function"}'
```

## Features Demonstrated

- ✅ Enhanced agent responses with personality
- ✅ Multi-agent chat system
- ✅ Session persistence and management
- ✅ WebSocket real-time communication
- ✅ RESTful API endpoints
- ✅ Responsive web interface
- ✅ Agent capability showcasing
- ✅ Error handling and graceful shutdown

## Customization

You can customize the agents by modifying the `NewEnhancedAgentManager()` function to:
- Add new agent types
- Modify agent personalities
- Adjust response patterns
- Add new capabilities

## Integration

This demo shows how to integrate AgenticGoKit's WebUI components:
- `webui.Server` for HTTP/WebSocket serving
- `webui.SessionManager` for chat session handling  
- `webui.AgentBridge` for agent communication
- Custom agent implementations with enhanced responses
