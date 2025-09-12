package webui

import (
	"encoding/json"
	"time"
)

// WebSocket Message Protocol for AgenticGoKit WebUI
// This defines the message format for bidirectional communication between client and server

// MessageType represents the type of WebSocket message
type MessageType string

const (
	// Client to Server message types
	MsgTypeChatMessage   MessageType = "chat_message"
	MsgTypeSessionCreate MessageType = "session_create"
	MsgTypeSessionJoin   MessageType = "session_join"
	MsgTypePing          MessageType = "ping"
	MsgTypeTyping        MessageType = "typing"

	// Server to Client message types
	MsgTypeAgentResponse  MessageType = "agent_response"
	MsgTypeAgentProgress  MessageType = "agent_progress"
	MsgTypeAgentChunk     MessageType = "agent_chunk"
	MsgTypeAgentComplete  MessageType = "agent_complete"
	MsgTypeAgentError     MessageType = "agent_error"
	MsgTypeSessionStatus  MessageType = "session_status"
	MsgTypePong           MessageType = "pong"
	MsgTypeError          MessageType = "error"
	MsgTypeSystemMessage  MessageType = "system_message"
	MsgTypeWorkflowUpdate MessageType = "workflow_update"
)

// WebSocketMessage represents the base structure for all WebSocket messages
type WebSocketMessage struct {
	Type      MessageType            `json:"type"`
	SessionID string                 `json:"session_id,omitempty"`
	MessageID string                 `json:"message_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// Client to Server Messages

// ChatMessageData represents a chat message from the user
type ChatMessageData struct {
	Content     string                 `json:"content"`
	MessageType string                 `json:"message_type,omitempty"` // "text", "file", "image"
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SessionCreateData represents a request to create a new session
type SessionCreateData struct {
	UserAgent string                 `json:"user_agent,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty"`
}

// SessionJoinData represents a request to join an existing session
type SessionJoinData struct {
	SessionID string `json:"session_id"`
}

// TypingData represents typing indicator
type TypingData struct {
	IsTyping bool `json:"is_typing"`
}

// Server to Client Messages

// AgentResponseData represents a response from an agent
type AgentResponseData struct {
	AgentName   string                 `json:"agent_name"`
	Content     string                 `json:"content"`
	Status      string                 `json:"status"` // "processing", "complete", "error"
	MessageType string                 `json:"message_type,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// AgentProgressData represents the progress of agent processing
type AgentProgressData struct {
	Agents          []AgentStatus `json:"agents"`
	OverallProgress float64       `json:"overall_progress"`
	CurrentAgent    string        `json:"current_agent,omitempty"`
	EstimatedTime   int           `json:"estimated_time,omitempty"` // seconds
}

// AgentChunkData represents a partial chunk of agent output
type AgentChunkData struct {
	AgentName  string                 `json:"agent_name"`
	Content    string                 `json:"content"`
	ChunkIndex int                    `json:"chunk_index"`
	TotalHint  int                    `json:"total_hint,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// AgentCompleteData represents the final completion of an agent response
type AgentCompleteData struct {
	AgentName string                 `json:"agent_name"`
	Content   string                 `json:"content"`
	Usage     map[string]interface{} `json:"usage,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AgentErrorData represents an agent error event
type AgentErrorData struct {
	AgentName string                 `json:"agent_name,omitempty"`
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowUpdateData represents orchestration-level updates
type WorkflowUpdateData struct {
	Step     string                 `json:"step"`
	Mode     string                 `json:"mode"`
	Agents   []string               `json:"agents,omitempty"`
	Message  string                 `json:"message,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AgentStatus represents the status of an individual agent
type AgentStatus struct {
	Name     string  `json:"name"`
	Status   string  `json:"status"`   // "waiting", "processing", "complete", "error"
	Progress float64 `json:"progress"` // 0.0 to 1.0
	Message  string  `json:"message,omitempty"`
}

// SessionStatusData represents session information
type SessionStatusData struct {
	SessionID    string    `json:"session_id"`
	Status       string    `json:"status"` // "active", "inactive", "error"
	MessageCount int       `json:"message_count"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}

// ErrorData represents an error message
type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SystemMessageData represents system notifications
type SystemMessageData struct {
	Level   string `json:"level"` // "info", "warning", "error"
	Message string `json:"message"`
	Action  string `json:"action,omitempty"` // Optional action for client
}

// Message Creation Helpers

// NewChatMessage creates a new chat message from client
func NewChatMessage(sessionID, content string) *WebSocketMessage {
	return &WebSocketMessage{
		Type:      MsgTypeChatMessage,
		SessionID: sessionID,
		MessageID: generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"content":      content,
			"message_type": "text",
		},
	}
}

// NewAgentResponse creates a new agent response message
func NewAgentResponse(sessionID, agentName, content, status string) *WebSocketMessage {
	return &WebSocketMessage{
		Type:      MsgTypeAgentResponse,
		SessionID: sessionID,
		MessageID: generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"agent_name": agentName,
			"content":    content,
			"status":     status,
		},
	}
}

// NewAgentProgress creates a new agent progress message
func NewAgentProgress(sessionID string, agents []AgentStatus, overallProgress float64) *WebSocketMessage {
	return &WebSocketMessage{
		Type:      MsgTypeAgentProgress,
		SessionID: sessionID,
		MessageID: generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"agents":           agents,
			"overall_progress": overallProgress,
		},
	}
}

// NewAgentChunk creates a new agent chunk message
func NewAgentChunk(sessionID, agentName, content string, chunkIndex, totalHint int, metadata map[string]interface{}) *WebSocketMessage {
	return &WebSocketMessage{
		Type:      MsgTypeAgentChunk,
		SessionID: sessionID,
		MessageID: generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"agent_name":  agentName,
			"content":     content,
			"chunk_index": chunkIndex,
			"total_hint":  totalHint,
			"metadata":    metadata,
		},
	}
}

// NewAgentComplete creates a new agent complete message
func NewAgentComplete(sessionID, agentName, content string, usage, metadata map[string]interface{}) *WebSocketMessage {
	return &WebSocketMessage{
		Type:      MsgTypeAgentComplete,
		SessionID: sessionID,
		MessageID: generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"agent_name": agentName,
			"content":    content,
			"usage":      usage,
			"metadata":   metadata,
		},
	}
}

// NewAgentError creates a new agent error message
func NewAgentError(sessionID, agentName, code, message string, metadata map[string]interface{}) *WebSocketMessage {
	return &WebSocketMessage{
		Type:      MsgTypeAgentError,
		SessionID: sessionID,
		MessageID: generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"agent_name": agentName,
			"code":       code,
			"message":    message,
			"metadata":   metadata,
		},
	}
}

// NewWorkflowUpdate creates a new workflow update message
func NewWorkflowUpdate(sessionID, step, mode, message string, agents []string, metadata map[string]interface{}) *WebSocketMessage {
	return &WebSocketMessage{
		Type:      MsgTypeWorkflowUpdate,
		SessionID: sessionID,
		MessageID: generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"step":     step,
			"mode":     mode,
			"agents":   agents,
			"message":  message,
			"metadata": metadata,
		},
	}
}

// NewSessionStatus creates a new session status message
func NewSessionStatus(sessionID, status string, messageCount int, createdAt, lastActivity time.Time) *WebSocketMessage {
	return &WebSocketMessage{
		Type:      MsgTypeSessionStatus,
		SessionID: sessionID,
		MessageID: generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"session_id":    sessionID,
			"status":        status,
			"message_count": messageCount,
			"created_at":    createdAt,
			"last_activity": lastActivity,
		},
	}
}

// NewWSErrorMessage creates a new WebSocket error message
func NewWSErrorMessage(sessionID, code, message, details string) *WebSocketMessage {
	return &WebSocketMessage{
		Type:      MsgTypeError,
		SessionID: sessionID,
		MessageID: generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"code":    code,
			"message": message,
			"details": details,
		},
	}
}

// NewSystemMessage creates a new system message
func NewSystemMessage(sessionID, level, message string) *WebSocketMessage {
	return &WebSocketMessage{
		Type:      MsgTypeSystemMessage,
		SessionID: sessionID,
		MessageID: generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"level":   level,
			"message": message,
		},
	}
}

// NewPong creates a pong response
func NewPong(sessionID string) *WebSocketMessage {
	return &WebSocketMessage{
		Type:      MsgTypePong,
		SessionID: sessionID,
		MessageID: generateMessageID(),
		Timestamp: time.Now(),
	}
}

// Message Parsing Helpers

// ParseChatMessage parses a chat message from WebSocket data
func (msg *WebSocketMessage) ParseChatMessage() (*ChatMessageData, error) {
	data := &ChatMessageData{}
	jsonData, err := json.Marshal(msg.Data)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonData, data)
	return data, err
}

// ParseSessionCreate parses a session create message
func (msg *WebSocketMessage) ParseSessionCreate() (*SessionCreateData, error) {
	data := &SessionCreateData{}
	jsonData, err := json.Marshal(msg.Data)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonData, data)
	return data, err
}

// ParseSessionJoin parses a session join message
func (msg *WebSocketMessage) ParseSessionJoin() (*SessionJoinData, error) {
	data := &SessionJoinData{}
	jsonData, err := json.Marshal(msg.Data)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonData, data)
	return data, err
}

// ParseTyping parses a typing indicator message
func (msg *WebSocketMessage) ParseTyping() (*TypingData, error) {
	data := &TypingData{}
	jsonData, err := json.Marshal(msg.Data)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonData, data)
	return data, err
}

// Validation Methods

// Validate validates the message structure
func (msg *WebSocketMessage) Validate() error {
	if msg.Type == "" {
		return NewProtocolError("INVALID_MESSAGE", "message type is required")
	}

	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	if msg.MessageID == "" {
		msg.MessageID = generateMessageID()
	}

	return nil
}

// Protocol Error Handling

// ProtocolError represents a WebSocket protocol error
type ProtocolError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *ProtocolError) Error() string {
	return e.Message
}

// NewProtocolError creates a new protocol error
func NewProtocolError(code, message string) *ProtocolError {
	return &ProtocolError{
		Code:    code,
		Message: message,
	}
}

// Common error codes
const (
	ErrInvalidMessage  = "INVALID_MESSAGE"
	ErrSessionNotFound = "SESSION_NOT_FOUND"
	ErrUnauthorized    = "UNAUTHORIZED"
	ErrConnectionLimit = "CONNECTION_LIMIT"
	ErrInternalError   = "INTERNAL_ERROR"
	ErrAgentError      = "AGENT_ERROR"
	ErrTimeout         = "TIMEOUT"
)
