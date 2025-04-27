package agentflow

import (
	"context"
	"errors"
	"sync"
)

// FIX: Define ErrSessionNotFound here
var ErrSessionNotFound = errors.New("session not found")

// Session represents the state associated with a specific interaction flow.
type Session interface {
	GetID() string
	GetState() State
	SetState(State)
}

// SessionStore defines the interface for managing session persistence.
type SessionStore interface {
	GetSession(ctx context.Context, sessionID string) (Session, error)
	SaveSession(ctx context.Context, session Session) error
	DeleteSession(ctx context.Context, sessionID string) error
}

// --- In-Memory Session Implementation ---

// memorySession implements the Session interface.
type memorySession struct {
	id    string
	state State
	mu    sync.RWMutex
}

// NewMemorySession creates a new session instance.
func NewMemorySession(id string, initialState State) Session {
	if initialState == nil {
		initialState = NewState() // Ensure state is never nil
	}
	return &memorySession{
		id:    id,
		state: initialState,
	}
}

func (s *memorySession) GetID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.id
}

func (s *memorySession) GetState() State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Return a clone to prevent external modification?
	// For now, return direct state. Caller should clone if needed.
	return s.state
}

func (s *memorySession) SetState(newState State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if newState == nil {
		s.state = NewState() // Ensure state is never set to nil
	} else {
		s.state = newState
	}
}

// --- In-Memory Session Store Implementation ---

// MemorySessionStore implements SessionStore using an in-memory map.
type MemorySessionStore struct {
	sessions map[string]Session
	mu       sync.RWMutex
}

// NewMemorySessionStore creates a new in-memory session store.
func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions: make(map[string]Session),
	}
}

func (s *MemorySessionStore) GetSession(ctx context.Context, sessionID string) (Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[sessionID]
	if !ok {
		// FIX: Return the specific ErrSessionNotFound
		return nil, ErrSessionNotFound
	}
	return session, nil
}

func (s *MemorySessionStore) SaveSession(ctx context.Context, session Session) error {
	if session == nil {
		return errors.New("cannot save nil session")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.GetID()] = session
	return nil
}

func (s *MemorySessionStore) DeleteSession(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
	return nil
}
