# WebSocket Implementation for AgenticGoKit WebUI

## Overview

The WebSocket implementation provides real-time bidirectional communication between the AgenticGoKit WebUI client and server. This enables live chat interactions, session management, and real-time agent responses.

## Architecture

### Components

1. **ConnectionManager** (`websocket.go`)
   - Manages WebSocket connections and message routing
   - Handles client registration/unregistration
   - Provides message broadcasting capabilities
   - Implements heartbeat/ping-pong mechanism

2. **Protocol Definitions** (`protocol.go`)
   - Defines message types and data structures
   - Provides helper functions for message creation and parsing
   - Implements message validation

3. **Session Management** (`types.go`)
   - Enhanced SessionManager with WebSocket integration
   - Thread-safe session operations
   - Session lifecycle management

## Message Protocol

### Message Structure

All WebSocket messages follow this JSON structure:

```json
{
  "type": "message_type",
  "session_id": "optional_session_id",
  "message_id": "unique_message_id",
  "timestamp": "2025-09-09T18:00:00Z",
  "data": {
    // Type-specific payload
  }
}
```

### Message Types

#### Client to Server

- **`session_create`**: Request to create a new session
- **`session_join`**: Request to join an existing session
- **`chat_message`**: Send a chat message to an agent
- **`ping`**: Heartbeat ping
- **`typing`**: Typing indicator

#### Server to Client

- **`session_status`**: Session information and status
- **`agent_response`**: Response from an agent
- **`agent_progress`**: Progress updates during agent processing
- **`pong`**: Heartbeat response
- **`error`**: Error notification
- **`system_message`**: System notifications

## API Examples

### Creating a Session

```javascript
const message = {
  type: "session_create",
  timestamp: new Date().toISOString(),
  data: {
    user_agent: "WebUI/1.0"
  }
};
websocket.send(JSON.stringify(message));
```

### Sending a Chat Message

```javascript
const message = {
  type: "chat_message",
  session_id: "your-session-id",
  timestamp: new Date().toISOString(),
  data: {
    content: "Hello, agent!",
    message_type: "text"
  }
};
websocket.send(JSON.stringify(message));
```

### Handling Agent Responses

```javascript
websocket.onmessage = function(event) {
  const message = JSON.parse(event.data);
  
  switch(message.type) {
    case "agent_response":
      console.log("Agent says:", message.data.content);
      break;
    case "session_status":
      console.log("Session:", message.data.session_id);
      break;
    case "error":
      console.error("Error:", message.data.message);
      break;
  }
};
```

## Connection Flow

1. **Connect**: Client establishes WebSocket connection to `/ws`
2. **Welcome**: Server sends system welcome message
3. **Session**: Client creates or joins a session
4. **Chat**: Real-time message exchange
5. **Heartbeat**: Automatic ping/pong for connection health
6. **Cleanup**: Graceful disconnection and resource cleanup

## Error Handling

The implementation includes comprehensive error handling:

- **Invalid JSON**: Malformed message parsing
- **Unknown Message Types**: Unsupported message types
- **Session Not Found**: Invalid session references
- **Connection Timeouts**: Automatic cleanup of stale connections
- **Protocol Errors**: Message validation failures

## Testing

The WebSocket implementation includes comprehensive tests:

- **Connection Tests**: Basic WebSocket connectivity
- **Protocol Tests**: Message format and parsing
- **Flow Tests**: Complete chat workflows
- **Error Tests**: Error handling scenarios
- **Manager Tests**: Connection management functionality

Run tests with:

```bash
cd internal/webui
go test -v
```

## Configuration

WebSocket behavior can be configured through:

- **Read/Write Timeouts**: Connection timeout settings
- **Buffer Sizes**: Message buffer configurations
- **Ping Interval**: Heartbeat frequency
- **Connection Limits**: Maximum concurrent connections

## Integration

The WebSocket implementation integrates with:

- **HTTP Server**: Endpoint at `/ws`
- **Session Manager**: Thread-safe session operations
- **Agent System**: Future integration for agent responses
- **Middleware**: CORS and security headers

## Demo Usage

Start the demo server:

```bash
cd internal/webui/demo
go run main.go
```

Connect to WebSocket:
- **URL**: `ws://localhost:8080/ws`
- **Interface**: `http://localhost:8080`

## Performance Considerations

- **Goroutine Management**: Each connection runs in separate goroutines
- **Memory Efficiency**: Connection cleanup prevents memory leaks
- **Message Buffering**: Efficient message queuing and delivery
- **Concurrent Safety**: Thread-safe operations throughout

## Security Features

- **Origin Validation**: CORS support (configurable)
- **Message Validation**: Input sanitization and validation
- **Rate Limiting**: Built-in connection and message rate limits
- **Secure Headers**: Security middleware integration

## Future Enhancements

Planned improvements for future tasks:

1. **Agent Integration**: Real agent system integration
2. **Authentication**: User authentication and authorization
3. **Persistence**: Message history persistence
4. **Scaling**: Multi-server WebSocket clustering
5. **Monitoring**: Connection and performance metrics

## Dependencies

- **gorilla/websocket**: WebSocket implementation
- **Standard Library**: JSON, HTTP, time, sync packages
- **AgenticGoKit Core**: Integration with core framework

---

**Status**: ✅ Complete - Task 2 (September 9, 2025)  
**Test Coverage**: 15/15 tests passing  
**Demo**: Functional at http://localhost:8080
