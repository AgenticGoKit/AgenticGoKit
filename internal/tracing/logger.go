// Package tracing provides internal tracing implementations for AgentFlow.
package tracing

import (
	"sync"
	"time"
)

// TraceEntry represents a single logged event during the execution flow.
type TraceEntry struct {
	Timestamp     time.Time `json:"timestamp"`
	Type          string    `json:"type"`
	EventID       string    `json:"event_id"`
	SessionID     string    `json:"session_id"`
	AgentID       string    `json:"agent_id,omitempty"`
	Error         string    `json:"error,omitempty"`
	Hook          string    `json:"hook,omitempty"`
	TargetAgentID string    `json:"target_agent_id,omitempty"`
	SourceAgentID string    `json:"source_agent_id,omitempty"`
}

// TraceLogger defines the interface for storing and retrieving trace entries.
type TraceLogger interface {
	Log(entry TraceEntry) error
	GetTrace(sessionID string) ([]TraceEntry, error)
}

// InMemoryTraceLogger is a simple in-memory implementation of TraceLogger.
type InMemoryTraceLogger struct {
	mu     sync.RWMutex
	traces map[string][]TraceEntry
}

// NewInMemoryTraceLogger creates a new in-memory trace logger.
func NewInMemoryTraceLogger() *InMemoryTraceLogger {
	return &InMemoryTraceLogger{
		traces: make(map[string][]TraceEntry),
	}
}

// Log adds a trace entry to the logger.
func (l *InMemoryTraceLogger) Log(entry TraceEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	sessionID := entry.SessionID
	if sessionID == "" {
		sessionID = "default"
	}

	l.traces[sessionID] = append(l.traces[sessionID], entry)
	return nil
}

// GetTrace retrieves all trace entries for a given session ID.
func (l *InMemoryTraceLogger) GetTrace(sessionID string) ([]TraceEntry, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if sessionID == "" {
		sessionID = "default"
	}

	entries, exists := l.traces[sessionID]
	if !exists {
		return []TraceEntry{}, nil
	}

	// Return a copy to avoid race conditions
	result := make([]TraceEntry, len(entries))
	copy(result, entries)
	return result, nil
}
