# AgenticGoKit WebUI Agent Bridge - Usage Guide

## Overview

The Agent Bridge Interface connects the AgenticGoKit WebUI with the core agent system, enabling real-time communication between web clients and AI agents through HTTP REST APIs and WebSocket connections.

## Quick Start

### 1. Initialize Components

```go
import "github.com/kunalkushwaha/agenticgokit/internal/webui"

// Create session manager
sessionConfig := webui.DefaultSessionConfig()
sessionManager, err := webui.NewEnhancedSessionManager(coreConfig, sessionConfig)

// Create agent bridge
bridgeConfig := webui.DefaultBridgeConfig()
bridge := webui.NewAgentBridge(agentManager, sessionManager, logger, bridgeConfig)

// Start the bridge
ctx := context.Background()
bridge.Start(ctx)

// Create HTTP handlers
handlers := webui.NewAgentHTTPHandlers(bridge, sessionManager, logger)

// Register routes
mux := http.NewServeMux()
handlers.RegisterHandlers(mux)
```

### 2. REST API Endpoints

#### Process Chat Messages
```bash
POST /api/chat
Content-Type: application/json

{
  "session_id": "unique-session-id",
  "message": "Hello, can you help me?",
  "user_id": "user-123",
  "agent_name": "assistant", // optional
  "metadata": {
    "source": "web",
    "priority": "normal"
  },
  "stream": false // true for streaming responses
}
```

#### Get Available Agents
```bash
GET /api/agents

Response:
{
  "agents": [
    {
      "name": "assistant",
      "description": "General purpose AI assistant",
      "capabilities": ["chat", "text", "analysis"],
      "is_enabled": true,
      "role": "assistant"
    }
  ],
  "count": 1,
  "timestamp": 1694123456
}
```

#### Session Management
```bash
# Get session info
GET /api/session/info?session_id=session-123

# Get session messages
GET /api/session/messages?session_id=session-123&limit=10&offset=0

# Delete session
DELETE /api/session/delete?session_id=session-123
```

#### Health Check
```bash
GET /api/health

Response:
{
  "status": "healthy",
  "timestamp": 1694123456,
  "components": {
    "bridge": true,
    "session_manager": true
  }
}
```

### 3. WebSocket Connection

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/ws?session_id=session-123&user_id=user-123');

// Send chat message
ws.send(JSON.stringify({
  type: 'chat_message',
  session_id: 'session-123',
  message: 'Hello!',
  agent_name: 'assistant',
  metadata: { source: 'websocket' },
  timestamp: Date.now()
}));

// Handle responses
ws.onmessage = (event) => {
  const response = JSON.parse(event.data);
  console.log('Agent response:', response);
};
```

### 4. Programmatic Usage

```go
// Process message programmatically
ctx := context.Background()
sessionID := "session-123"
message := "Hello, agent!"
metadata := map[string]interface{}{"source": "api"}

// Send message to agent
err := bridge.ProcessChatMessage(ctx, sessionID, message, metadata)

// Get response stream
responseStream := bridge.GetResponseStream(sessionID)

// Listen for responses
select {
case response := <-responseStream:
    fmt.Printf("Agent response: %s\n", response.Content)
case <-time.After(30 * time.Second):
    fmt.Println("Timeout waiting for response")
}

// Clean up
bridge.CloseResponseStream(sessionID)
```

## Configuration

### Bridge Configuration
```go
bridgeConfig := &webui.BridgeConfig{
    AgentTimeout:       30 * time.Second,  // Max agent processing time
    ResponseBufferSize: 100,               // Response channel buffer size
    MaxConcurrentTasks: 10,                // Max concurrent agent tasks
    RetryAttempts:      3,                 // Retry failed agent calls
    RetryDelay:         1 * time.Second,   // Delay between retries
    StreamingEnabled:   true,              // Enable response streaming
    ChunkSize:          1024,              // Streaming chunk size
}
```

### Session Configuration
```go
sessionConfig := &webui.SessionConfig{
    StorageType:      "memory",            // "memory" or "file"
    StorageDir:       "./sessions",        // File storage directory
    SessionTimeout:   24 * time.Hour,      // Session expiration time
    CleanupInterval:  1 * time.Hour,       // Cleanup frequency
    MaxSessions:      1000,                // Maximum active sessions
    MaxMessages:      1000,                // Messages per session
    MessageRetention: 7 * 24 * time.Hour,  // Message retention period
    AutoSave:         true,                // Auto-save sessions
    SaveInterval:     5 * time.Minute,     // Save frequency
}
```

## Response Formats

### Chat Response
```json
{
  "session_id": "session-123",
  "agent_name": "assistant",
  "content": "Hello! How can I help you today?",
  "status": "complete", // "processing", "partial", "complete", "error"
  "error": "", // Error message if status is "error"
  "is_streaming": false,
  "chunk_index": 0, // For streaming responses
  "total_chunks": 1,
  "metadata": {
    "processing_time_ms": 150,
    "model": "gpt-4",
    "tokens_used": 25
  },
  "timestamp": 1694123456
}
```

### WebSocket Message Types
- `chat_message` - User message to agent
- `agent_response` - Agent response to user
- `agent_progress` - Processing progress updates
- `session_status` - Session state changes
- `error` - Error notifications
- `ping`/`pong` - Connection health

## Error Handling

### HTTP Errors
- `400 Bad Request` - Invalid request format
- `404 Not Found` - Session not found
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Processing failure
- `503 Service Unavailable` - System overloaded

### WebSocket Errors
```json
{
  "type": "error",
  "session_id": "session-123",
  "error": {
    "code": "AGENT_ERROR",
    "message": "Agent processing failed",
    "details": "Connection timeout"
  },
  "timestamp": 1694123456
}
```

## Best Practices

### 1. Session Management
- Always provide a session ID for message continuity
- Use meaningful session IDs that don't expose sensitive data
- Clean up expired sessions regularly
- Implement session persistence for production use

### 2. Error Handling
- Always check response status before processing content
- Implement retry logic for transient failures
- Handle WebSocket disconnections gracefully
- Monitor agent availability and health

### 3. Performance
- Use streaming for long responses
- Implement client-side response caching
- Batch multiple quick messages when possible
- Monitor response times and implement timeouts

### 4. Security
- Validate all input messages
- Implement rate limiting per session/user
- Use secure WebSocket connections (WSS) in production
- Sanitize agent responses before displaying

## Monitoring & Debugging

### Health Checks
```bash
# Check system health
curl http://localhost:8080/api/health

# Monitor agent availability
curl http://localhost:8080/api/agents
```

### Logging
The bridge provides comprehensive logging for:
- Message processing events
- Agent interactions
- Error conditions
- Performance metrics
- Session lifecycle events

### Metrics
Track these key metrics:
- Messages processed per second
- Average response time
- Error rates by type
- Active session count
- Agent utilization

## Integration Examples

### React Frontend
```javascript
import { useState, useEffect } from 'react';

function ChatComponent() {
  const [messages, setMessages] = useState([]);
  const [socket, setSocket] = useState(null);

  useEffect(() => {
    const ws = new WebSocket('ws://localhost:8080/ws?session_id=session-123');
    ws.onmessage = (event) => {
      const response = JSON.parse(event.data);
      if (response.type === 'agent_response') {
        setMessages(prev => [...prev, {
          role: 'assistant',
          content: response.content,
          timestamp: response.timestamp
        }]);
      }
    };
    setSocket(ws);
    return () => ws.close();
  }, []);

  const sendMessage = (message) => {
    socket.send(JSON.stringify({
      type: 'chat_message',
      session_id: 'session-123',
      message: message,
      timestamp: Date.now()
    }));
  };

  return <div>{/* Chat UI */}</div>;
}
```

### Python Client
```python
import websocket
import json
import threading

class AgentClient:
    def __init__(self, url, session_id):
        self.url = url
        self.session_id = session_id
        self.ws = None
        
    def connect(self):
        self.ws = websocket.WebSocketApp(
            f"{self.url}?session_id={self.session_id}",
            on_message=self.on_message,
            on_error=self.on_error
        )
        self.ws.run_forever()
        
    def send_message(self, message):
        self.ws.send(json.dumps({
            'type': 'chat_message',
            'session_id': self.session_id,
            'message': message,
            'timestamp': int(time.time() * 1000)
        }))
        
    def on_message(self, ws, message):
        response = json.loads(message)
        print(f"Agent: {response.get('content', '')}")

client = AgentClient('ws://localhost:8080/ws', 'session-123')
client.connect()
```

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Check if server is running on correct port
   - Verify firewall settings
   - Ensure WebSocket upgrade is supported

2. **Session Not Found**
   - Verify session ID format
   - Check session expiration settings
   - Ensure session was properly created

3. **Agent Not Responding**
   - Check agent availability via `/api/agents`
   - Verify agent is properly registered
   - Check agent health and configuration

4. **High Latency**
   - Monitor agent processing times
   - Check network connectivity
   - Consider increasing timeout values

For more detailed information, refer to the source code documentation and examples in the `examples/` directory.
