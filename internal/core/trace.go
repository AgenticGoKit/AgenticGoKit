package agentflow

import (
	"context"
	"encoding/json"
	"log"
	"sort"
	"sync"
	"time"
)

// TraceEntry represents a single logged event during the execution flow.
// It captures the state and context at a specific point, often triggered by a callback hook.
type TraceEntry struct {
	Timestamp time.Time    `json:"timestamp"`            // Time the entry was logged
	SessionID string       `json:"session_id,omitempty"` // ID of the session this trace belongs to
	EventID   string       `json:"event_id,omitempty"`   // ID of the specific event being processed
	Hook      HookPoint    `json:"hook"`                 // The hook point that triggered this log entry
	AgentName *string      `json:"agent_name,omitempty"` // Name of the agent involved (if applicable)
	Input     *SimpleState `json:"input,omitempty"`      // Input state (e.g., before agent run)
	Output    *SimpleState `json:"output,omitempty"`     // Output state (e.g., after agent run)
	Error     *string      `json:"error,omitempty"`      // Error message, if any occurred at this step
	Details   interface{}  `json:"details,omitempty"`    // Any other hook-specific details (e.g., model name, tool name, parameters)
}

// Helper to convert error to string pointer for TraceEntry
func errorToStringPtr(err error) *string {
	if err == nil {
		return nil
	}
	errStr := err.Error()
	return &errStr // Return pointer to the string
}

// Helper to convert state to SimpleState pointer for TraceEntry
// Assumes state can be represented as SimpleState for logging.
func stateToSimpleStatePtr(s State) *SimpleState {
	if s == nil {
		return nil
	}
	// Attempt type assertion or conversion if State is an interface
	if ss, ok := s.(*SimpleState); ok {
		// Return a clone to avoid race conditions if the original state is modified later
		cloned := ss.Clone().(*SimpleState)
		return cloned
	}
	// Fallback: Try to marshal/unmarshal or create a basic representation
	// For now, return nil if not *SimpleState
	log.Printf("Warning: Trace logger received state of type %T, expected *SimpleState", s)
	return nil
}

// TraceLogger defines the interface for storing and retrieving trace entries.
type TraceLogger interface {
	Log(entry TraceEntry) error
	GetTrace(sessionID string) ([]TraceEntry, error)
}

// --- InMemoryTraceLogger ---

// InMemoryTraceLogger provides a thread-safe, in-memory implementation of TraceLogger.
type InMemoryTraceLogger struct {
	mu      sync.RWMutex
	entries map[string][]TraceEntry // Keyed by SessionID
}

// NewInMemoryTraceLogger creates a new, empty InMemoryTraceLogger.
func NewInMemoryTraceLogger() *InMemoryTraceLogger {
	return &InMemoryTraceLogger{
		entries: make(map[string][]TraceEntry),
	}
}

// Log records a single trace entry in memory. It is thread-safe.
func (l *InMemoryTraceLogger) Log(entry TraceEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Ensure SessionID is present, default if necessary
	sessionID := entry.SessionID
	if sessionID == "" {
		sessionID = "default_session" // Or handle as an error?
		log.Println("Warning: Trace entry logged without SessionID, using 'default_session'")
		entry.SessionID = sessionID
	}

	l.entries[sessionID] = append(l.entries[sessionID], entry)
	return nil
}

// GetTrace retrieves all trace entries for a given session ID, ordered chronologically.
// It returns a copy of the entries to ensure thread safety.
func (l *InMemoryTraceLogger) GetTrace(sessionID string) ([]TraceEntry, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	sessionEntries, ok := l.entries[sessionID]
	if !ok {
		return []TraceEntry{}, nil // Return empty slice if session not found
	}

	// Return a copy of the slice
	entriesCopy := make([]TraceEntry, len(sessionEntries))
	copy(entriesCopy, sessionEntries)

	// Sort the copy by timestamp (optional, but good practice)
	sort.SliceStable(entriesCopy, func(i, j int) bool {
		return entriesCopy[i].Timestamp.Before(entriesCopy[j].Timestamp)
	})

	return entriesCopy, nil
}

// --- End InMemoryTraceLogger ---

// --- NoOp TraceLogger ---

// noOpTraceLogger provides a TraceLogger implementation that does nothing.
type noOpTraceLogger struct{}

// NewNoOpTraceLogger creates a new no-operation trace logger.
// It implements the TraceLogger interface.
func NewNoOpTraceLogger() TraceLogger {
	return &noOpTraceLogger{}
}

// Log does nothing.
func (n *noOpTraceLogger) Log(entry TraceEntry) error { return nil }

// GetTrace always returns an empty slice and no error.
func (n *noOpTraceLogger) GetTrace(sessionID string) ([]TraceEntry, error) {
	return []TraceEntry{}, nil
}

// Close does nothing.
func (n *noOpTraceLogger) Close() error { return nil }

// Compile-time check
var _ TraceLogger = (*noOpTraceLogger)(nil)

// --- End NoOp TraceLogger ---

// --- Trace Callback ---
const TraceCallbackName = "CoreTraceLoggerCallback"

// NewTraceCallbacks creates a set of standard callback functions for tracing.
func NewTraceCallbacks(logger TraceLogger) map[HookPoint]CallbackFunc {
	if logger == nil {
		log.Println("Warning: NewTraceCallbacks received nil logger, tracing disabled.")
		return map[HookPoint]CallbackFunc{}
	}

	// Define functions matching CallbackFunc signature
	beforeEventFunc := func(ctx context.Context, currentState State, event *Event) (State, error) {
		if event == nil { // Add nil check
			log.Println("Trace: BeforeEventHandling - Event is nil")
			return currentState, nil
		}
		log.Printf("Trace: Hook=%s, EventID=%s, State=%v", HookBeforeEventHandling, (*event).GetID(), currentState)
		return currentState, nil
	}

	afterEventFunc := func(ctx context.Context, currentState State, event *Event) (State, error) {
		if event == nil { // Add nil check
			log.Println("Trace: AfterEventHandling - Event is nil")
			return currentState, nil
		}
		log.Printf("Trace: Hook=%s, EventID=%s, State=%v", HookAfterEventHandling, (*event).GetID(), currentState)
		return currentState, nil
	}

	beforeAgentFunc := func(ctx context.Context, currentState State, event *Event) (State, error) {
		if event == nil { // Add nil check
			log.Println("Trace: BeforeAgentRun - Event is nil")
			return currentState, nil
		}
		log.Printf("Trace: Hook=%s, EventID=%s, State=%v", HookBeforeAgentRun, (*event).GetID(), currentState)
		return currentState, nil
	}

	afterAgentFunc := func(ctx context.Context, currentState State, event *Event) (State, error) {
		if event == nil { // Add nil check
			log.Println("Trace: AfterAgentRun - Event is nil")
			return currentState, nil
		}
		log.Printf("Trace: Hook=%s, EventID=%s, State=%v", HookAfterAgentRun, (*event).GetID(), currentState)
		return currentState, nil
	}

	return map[HookPoint]CallbackFunc{
		HookBeforeEventHandling: beforeEventFunc,
		HookAfterEventHandling:  afterEventFunc,
		HookBeforeAgentRun:      beforeAgentFunc,
		HookAfterAgentRun:       afterAgentFunc,
	}
}

// --- End Trace Callback ---

// MarshalJSON ensures TraceEntry marshals correctly. (Keep as is)
func (t *TraceEntry) MarshalJSON() ([]byte, error) {
	type Alias TraceEntry
	return json.Marshal(&struct {
		Timestamp string `json:"timestamp"` // Format timestamp
		*Alias
	}{
		Timestamp: t.Timestamp.Format(time.RFC3339Nano),
		Alias:     (*Alias)(t),
	})
}
