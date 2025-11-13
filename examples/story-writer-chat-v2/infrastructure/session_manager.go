package infrastructure

import (
	"fmt"
	"sync"
	"time"
)

// ChatMessage represents a stored chat message
type ChatMessage struct {
	Role      string    `json:"role"` // user, agent name, system
	Content   string    `json:"content"`
	Agent     string    `json:"agent,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatSession represents a chat session
type ChatSession struct {
	ID        string        `json:"id"`
	Messages  []ChatMessage `json:"messages"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// SessionManager manages chat sessions (reusable across applications)
type SessionManager struct {
	sessions map[string]*ChatSession
	mu       sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*ChatSession),
	}
}

// CreateSession creates a new chat session
func (sm *SessionManager) CreateSession() *ChatSession {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session := &ChatSession{
		ID:        fmt.Sprintf("session_%d", time.Now().UnixNano()),
		Messages:  []ChatMessage{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	sm.sessions[session.ID] = session
	return session
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(id string) (*ChatSession, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	session, exists := sm.sessions[id]
	return session, exists
}

// AddMessage adds a message to a session
func (sm *SessionManager) AddMessage(sessionID string, msg ChatMessage) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[sessionID]; exists {
		session.Messages = append(session.Messages, msg)
		session.UpdatedAt = time.Now()
	}
}

// ListSessions returns all sessions
func (sm *SessionManager) ListSessions() []*ChatSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*ChatSession, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}
