package agentflow

import (
	"context"
	"fmt"
	"sync"
)

// SessionIDKey is the standard metadata key used to store the session identifier.
const SessionIDKey = "session_id"

// SessionStore defines the interface for managing agent session state.
// Implementations must be thread-safe.
type SessionStore interface {
	// GetSession retrieves the state associated with a session ID.
	// It returns the State, a boolean indicating if the session was found, and an error.
	GetSession(ctx context.Context, sessionID string) (state State, found bool, err error)

	// SaveSession stores or updates the state for a given session ID.
	SaveSession(ctx context.Context, sessionID string, state State) error

	// TODO: Consider adding DeleteSession, ListSessions etc. later if needed
	// DeleteSession(ctx context.Context, sessionID string) error
}

// --- In-Memory Session Store Implementation ---

// MemorySessionStore provides an in-memory, concurrency-safe implementation of SessionStore
// storing the agent State interface.
type MemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]State // <<< Store State interface directly
}

// NewMemorySessionStore creates and initializes a new MemorySessionStore.
func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions: make(map[string]State), // <<< Initialize map for State
	}
}

// Compile-time check to ensure MemorySessionStore implements SessionStore
var _ SessionStore = (*MemorySessionStore)(nil)

// GetSession retrieves the state for a session ID from memory.
func (s *MemorySessionStore) GetSession(ctx context.Context, sessionID string) (State, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Retrieve the State interface directly
	state, found := s.sessions[sessionID]
	if !found {
		return nil, false, nil // Not found, return nil state, false, nil error
	}

	// Return a clone to prevent external modification of the stored state
	// state is already a State interface, Clone() returns State
	return state.Clone(), true, nil // <<< Correct: Call Clone() on the State interface
}

// SaveSession stores or updates the state for a session ID in memory.
func (s *MemorySessionStore) SaveSession(ctx context.Context, sessionID string, state State) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}
	if state == nil {
		return fmt.Errorf("cannot save nil state for session %s", sessionID)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Store a clone to prevent the caller from modifying the stored state later
	// state.Clone() returns State, which matches the map type
	s.sessions[sessionID] = state.Clone() // <<< Correct: Assign State interface to map
	return nil
}
