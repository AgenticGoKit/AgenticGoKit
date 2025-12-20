package webui

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
)

// ChatSession represents a chat session between user and agents
type ChatSession struct {
	ID        string          `json:"id"`
	Messages  []ChatMessage   `json:"messages"`
	AgentCtx  context.Context `json:"-"`
	State     core.State      `json:"-"`
	Memory    core.Memory     `json:"-"`
	Config    *core.Config    `json:"-"`
	CreatedAt time.Time       `json:"created_at"`
	LastUsed  time.Time       `json:"last_used"`

	// Session metadata
	UserAgent string `json:"user_agent,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`

	// Session state
	Active bool `json:"active"`
}

// ChatMessage represents a single message in the chat
type ChatMessage struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"` // "user" or "agent"
	Content   string    `json:"content"`
	AgentName string    `json:"agent_name,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"` // "processing", "complete", "error"

	// Message metadata
	MessageType string                 `json:"message_type,omitempty"` // "text", "file", "image", etc.
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// SessionManager handles session lifecycle and operations
type SessionManager struct {
	sessions map[string]*ChatSession
	config   *core.Config
	mu       sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager(config *core.Config) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*ChatSession),
		config:   config,
	}
}

// CreateSession creates a new chat session
func (sm *SessionManager) CreateSession() *ChatSession {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session := NewChatSession()
	sm.sessions[session.ID] = session
	return session
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) *ChatSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.sessions[sessionID]
}

// AddMessage adds a message to a session
func (sm *SessionManager) AddMessage(sessionID string, message ChatMessage) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	session.AddMessage(message)
	return nil
}

// ListSessions returns all active sessions
func (sm *SessionManager) ListSessions() []*ChatSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*ChatSession, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		if session.Active {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

// CleanupExpiredSessions removes expired sessions
func (sm *SessionManager) CleanupExpiredSessions() int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	removed := 0
	for id, session := range sm.sessions {
		if session.IsExpired() {
			delete(sm.sessions, id)
			removed++
		}
	}
	return removed
}

// NewChatSession creates a new chat session
func NewChatSession() *ChatSession {
	return &ChatSession{
		ID:        generateSessionID(),
		Messages:  make([]ChatMessage, 0),
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		Active:    true,
	}
}

// AddMessage adds a message to the session
func (s *ChatSession) AddMessage(message ChatMessage) {
	if message.ID == "" {
		message.ID = generateMessageID()
	}
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	s.Messages = append(s.Messages, message)
	s.LastUsed = time.Now()
}

// GetLastMessage returns the last message in the session
func (s *ChatSession) GetLastMessage() *ChatMessage {
	if len(s.Messages) == 0 {
		return nil
	}
	return &s.Messages[len(s.Messages)-1]
}

// GetMessagesByRole returns all messages by a specific role
func (s *ChatSession) GetMessagesByRole(role string) []ChatMessage {
	var messages []ChatMessage
	for _, msg := range s.Messages {
		if msg.Role == role {
			messages = append(messages, msg)
		}
	}
	return messages
}

// GetMessageCount returns the total number of messages
func (s *ChatSession) GetMessageCount() int {
	return len(s.Messages)
}

// IsExpired checks if the session has expired (not used for more than 24 hours)
func (s *ChatSession) IsExpired() bool {
	return time.Since(s.LastUsed) > 24*time.Hour
}

// Touch updates the last used timestamp
func (s *ChatSession) Touch() {
	s.LastUsed = time.Now()
}

// Close marks the session as inactive
func (s *ChatSession) Close() {
	s.Active = false
}

// generateSessionID creates a unique session identifier
func generateSessionID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return "session_" + hex.EncodeToString([]byte(time.Now().String()))[:16]
	}
	return hex.EncodeToString(bytes)
}

// generateMessageID creates a unique message identifier
func generateMessageID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return "msg_" + hex.EncodeToString([]byte(time.Now().String()))[:12]
	}
	return hex.EncodeToString(bytes)
}

// NewUserMessage creates a new user message
func NewUserMessage(content string) ChatMessage {
	return ChatMessage{
		ID:        generateMessageID(),
		Role:      "user",
		Content:   content,
		Timestamp: time.Now(),
		Status:    "complete",
	}
}

// NewAgentMessage creates a new agent message
func NewAgentMessage(agentName, content string) ChatMessage {
	return ChatMessage{
		ID:        generateMessageID(),
		Role:      "agent",
		Content:   content,
		AgentName: agentName,
		Timestamp: time.Now(),
		Status:    "complete",
	}
}

// NewProcessingMessage creates a message indicating agent is processing
func NewProcessingMessage(agentName string) ChatMessage {
	return ChatMessage{
		ID:        generateMessageID(),
		Role:      "agent",
		Content:   "Processing...",
		AgentName: agentName,
		Timestamp: time.Now(),
		Status:    "processing",
	}
}

// NewErrorMessage creates an error message
func NewErrorMessage(agentName, errorText string) ChatMessage {
	return ChatMessage{
		ID:        generateMessageID(),
		Role:      "agent",
		Content:   "Error occurred during processing",
		AgentName: agentName,
		Timestamp: time.Now(),
		Status:    "error",
		Error:     errorText,
	}
}

// SessionStats provides statistics about a session
type SessionStats struct {
	TotalMessages   int           `json:"total_messages"`
	UserMessages    int           `json:"user_messages"`
	AgentMessages   int           `json:"agent_messages"`
	ErrorMessages   int           `json:"error_messages"`
	SessionDuration time.Duration `json:"session_duration"`
	LastActivity    time.Time     `json:"last_activity"`
	MessagesPerHour float64       `json:"messages_per_hour"`
}

// GetStats returns statistics for the session
func (s *ChatSession) GetStats() SessionStats {
	userMsgs := len(s.GetMessagesByRole("user"))
	agentMsgs := len(s.GetMessagesByRole("agent"))

	var errorMsgs int
	for _, msg := range s.Messages {
		if msg.Status == "error" {
			errorMsgs++
		}
	}

	duration := s.LastUsed.Sub(s.CreatedAt)
	var messagesPerHour float64
	if duration.Hours() > 0 {
		messagesPerHour = float64(len(s.Messages)) / duration.Hours()
	}

	return SessionStats{
		TotalMessages:   len(s.Messages),
		UserMessages:    userMsgs,
		AgentMessages:   agentMsgs,
		ErrorMessages:   errorMsgs,
		SessionDuration: duration,
		LastActivity:    s.LastUsed,
		MessagesPerHour: messagesPerHour,
	}
}

