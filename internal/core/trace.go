package agentflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// --- Context Key for Tracer ---
type tracerContextKey struct{}

// WithTracer returns a new context with the provided Tracer embedded.
// FIX: Use Tracer interface type
func WithTracer(ctx context.Context, tracer Tracer) context.Context {
	if tracer == nil {
		return ctx // Avoid adding nil tracer
	}
	return context.WithValue(ctx, tracerContextKey{}, tracer)
}

// GetTracer retrieves the Tracer instance from the context, if present.
// Returns nil if no Tracer is found.
// FIX: Use Tracer interface type
func GetTracer(ctx context.Context) Tracer {
	// FIX: Use Tracer interface type in assertion
	tracer, _ := ctx.Value(tracerContextKey{}).(Tracer)
	return tracer // Returns nil if not found or type assertion fails
}

// TraceEntry represents a single logged event during the execution flow.
// It captures the state and context at a specific point, often triggered by a callback hook.
type TraceEntry struct {
	Timestamp     time.Time    `json:"timestamp"`
	Type          string       `json:"type"` // "event_start", "event_end", "agent_start", "agent_end"
	EventID       string       `json:"event_id"`
	SessionID     string       `json:"session_id"`
	AgentID       string       `json:"agent_id,omitempty"`
	State         State        `json:"state,omitempty"` // State *before* agent run or *after* event handling
	Error         string       `json:"error,omitempty"` // Only for agent_end or event_end (if applicable)
	Hook          HookPoint    `json:"hook,omitempty"`  // The hook point that triggered this entry
	TargetAgentID string       `json:"target_agent_id,omitempty"`
	SourceAgentID string       `json:"source_agent_id,omitempty"`
	AgentResult   *AgentResult `json:"agent_result,omitempty"` // Result from HookAfterAgentRun
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
	Logger().Warn().
		Str("state_type", fmt.Sprintf("%T", s)).
		Msg("Trace logger received state of unexpected type, expected *SimpleState")
	return nil
}

// TraceLogger defines the interface for storing and retrieving trace entries.
type TraceLogger interface {
	Log(entry TraceEntry) error
	GetTrace(sessionID string) ([]TraceEntry, error)
}

// --- Tracer Interface ---
// FIX: Define the Tracer interface
type Tracer interface {
	RecordEventStart(event Event)
	RecordEventEnd(event Event, finalState State, err error)
	RecordAgentStart(event Event, state State)
	RecordAgentEnd(event Event, result AgentResult, err error)
	GetTrace(sessionID string) ([]TraceEntry, error)
}

// --- InMemoryTracer ---
// InMemoryTracer stores trace entries in memory, grouped by session ID.
type InMemoryTracer struct {
	mu     sync.RWMutex
	traces map[string][]TraceEntry // sessionID -> []TraceEntry
	logger TraceLogger             // Underlying logger to also write entries
}

// FIX: Define NewInMemoryTracer constructor
// NewInMemoryTracer creates a new in-memory tracer.
func NewInMemoryTracer(logger TraceLogger, sessionID string) *InMemoryTracer {
	// Note: sessionID passed here isn't strictly needed for the map structure,
	// but might be useful if pre-initializing or for other logic.
	// The map keys are derived from events later.
	return &InMemoryTracer{
		traces: make(map[string][]TraceEntry),
		logger: logger, // Can be nil if no underlying logging is desired
	}
}

// RecordEventStart logs the beginning of event processing.
func (t *InMemoryTracer) RecordEventStart(event Event) {
	// FIX: Handle both return values from GetMetadataValue
	sessionID, _ := event.GetMetadataValue(SessionIDKey) // Ignore the 'ok' boolean
	entry := TraceEntry{
		Timestamp: time.Now(),
		Type:      "event_start",
		EventID:   event.GetID(),
		SessionID: sessionID,
	}
	t.addEntry(sessionID, entry)
}

// RecordEventEnd logs the end of event processing.
func (t *InMemoryTracer) RecordEventEnd(event Event, finalState State, err error) {
	// FIX: Handle both return values from GetMetadataValue
	sessionID, _ := event.GetMetadataValue(SessionIDKey) // Ignore the 'ok' boolean
	entry := TraceEntry{
		Timestamp: time.Now(),
		Type:      "event_end",
		EventID:   event.GetID(),
		SessionID: sessionID,
		State:     finalState, // Record final state
	}
	if err != nil {
		entry.Error = err.Error()
	}
	t.addEntry(sessionID, entry)
}

// RecordAgentStart logs the beginning of an agent's execution.
func (t *InMemoryTracer) RecordAgentStart(event Event, state State) {
	// FIX: Handle both return values from GetMetadataValue
	sessionID, _ := event.GetMetadataValue(SessionIDKey) // Ignore the 'ok' boolean
	entry := TraceEntry{
		Timestamp: time.Now(),
		Type:      "agent_start",
		EventID:   event.GetID(),
		SessionID: sessionID,
		AgentID:   event.GetTargetAgentID(),
		State:     state, // Record state *before* agent runs
	}
	t.addEntry(sessionID, entry)
}

// RecordAgentEnd logs the end of an agent's execution.
func (t *InMemoryTracer) RecordAgentEnd(event Event, result AgentResult, err error) {
	// FIX: Handle both return values from GetMetadataValue
	sessionID, _ := event.GetMetadataValue(SessionIDKey) // Ignore the 'ok' boolean
	entry := TraceEntry{
		Timestamp: time.Now(),
		Type:      "agent_end",
		EventID:   event.GetID(),
		SessionID: sessionID,
		AgentID:   event.GetTargetAgentID(),
		// FIX: Use the correct field name 'AgentResult' (and it's a pointer)
		AgentResult: &result, // Assign to AgentResult (pointer)
	}
	if err != nil {
		entry.Error = err.Error()
	} else if result.Error != "" { // Also capture error string within result
		entry.Error = result.Error
	}
	t.addEntry(sessionID, entry)
}

// addEntry safely adds a trace entry to the map and logs it.
func (t *InMemoryTracer) addEntry(sessionID string, entry TraceEntry) {
	t.mu.Lock()
	t.traces[sessionID] = append(t.traces[sessionID], entry)
	t.mu.Unlock()

	if t.logger != nil {
		t.logger.Log(entry) // Log to underlying logger
	}
}

// GetTrace retrieves all trace entries for a given session ID.
func (t *InMemoryTracer) GetTrace(sessionID string) ([]TraceEntry, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	entries, ok := t.traces[sessionID]
	if !ok {
		// Return empty slice instead of error? Or specific error type?
		return []TraceEntry{}, nil // Return empty slice if not found
	}
	// Return a copy to prevent external modification? For now, return direct slice.
	return entries, nil
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
		Logger().Warn().
			Msg("Trace entry logged without SessionID, using 'default_session'")
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
		Logger().Warn().Msg("NewTraceCallbacks received nil logger, tracing disabled.")
		return map[HookPoint]CallbackFunc{}
	}

	// Define functions matching CallbackFunc signature
	beforeEventFunc := func(ctx context.Context, args CallbackArgs) (State, error) {
		tracer := GetTracer(ctx)
		if tracer == nil {
			return args.State, nil // No tracer, do nothing
		}
		tracer.RecordEventStart(args.Event)
		return args.State, nil
	}

	afterEventFunc := func(ctx context.Context, args CallbackArgs) (State, error) {
		tracer := GetTracer(ctx)
		if tracer == nil {
			return args.State, nil // No tracer, do nothing
		}
		// Note: The 'result' and 'err' for the *overall* event handling are not
		// directly available here. This hook fires after the orchestrator.Dispatch
		// returns, but doesn't inherently know the final outcome beyond the state.
		// If a final event result is needed, the Runner loop would need to manage it.
		tracer.RecordEventEnd(args.Event, args.State, nil) // Pass nil for error for now
		return args.State, nil
	}

	beforeAgentFunc := func(ctx context.Context, args CallbackArgs) (State, error) {
		tracer := GetTracer(ctx)
		if tracer == nil {
			return args.State, nil // No tracer, do nothing
		}
		tracer.RecordAgentStart(args.Event, args.State)
		return args.State, nil
	}

	afterAgentFunc := func(ctx context.Context, args CallbackArgs) (State, error) {
		tracer := GetTracer(ctx)
		if tracer == nil {
			// FIX: Return args.State even if no tracer
			return args.State, nil // No tracer, do nothing
		}
		// FIX: Access result and error from args
		tracer.RecordAgentEnd(args.Event, args.Output, args.Error)
		// FIX: Return args.State (potentially modified by the agent result)
		// Although this specific callback doesn't modify state, return the current one.
		// If the agent returned a new state, args.State might not reflect it yet
		// unless the orchestrator updated it before invoking this hook.
		// Let's assume args.Output.OutputState is the definitive state *after* the agent.
		if args.Output.OutputState != nil {
			return args.Output.OutputState, nil
		}
		return args.State, nil // Fallback to incoming state if agent output state is nil
	}

	return map[HookPoint]CallbackFunc{
		HookBeforeEventHandling: beforeEventFunc,
		HookAfterEventHandling:  afterEventFunc,
		HookBeforeAgentRun:      beforeAgentFunc,
		HookAfterAgentRun:       afterAgentFunc,
	}
}

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

// FileTraceLogger logs trace entries to session-specific JSON files.
type FileTraceLogger struct {
	logDir string
	mu     sync.Mutex
	files  map[string]*os.File
}

// NewFileTraceLogger creates a logger that writes to the specified directory.
func NewFileTraceLogger(logDir string) (*FileTraceLogger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory '%s': %w", logDir, err)
	}
	return &FileTraceLogger{
		logDir: logDir,
		files:  make(map[string]*os.File),
	}, nil
}

// Log writes a trace entry to the appropriate session file.
func (l *FileTraceLogger) Log(entry TraceEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if entry.SessionID == "" {
		return fmt.Errorf("cannot log trace entry with empty SessionID")
	}

	file, ok := l.files[entry.SessionID]
	var err error
	if !ok {
		filename := filepath.Join(l.logDir, fmt.Sprintf("%s.trace.json", entry.SessionID))
		// Open in append mode, create if not exists
		file, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open trace file '%s': %w", filename, err)
		}
		l.files[SessionIDKey] = file
		// Write '[' if the file is new/empty to start the JSON array
		if stat, _ := file.Stat(); stat.Size() == 0 {
			if _, err := file.WriteString("[\n"); err != nil {
				return fmt.Errorf("failed to write initial '[' to trace file '%s': %w", filename, err)
			}
		} else {
			// If file not empty, add comma before new entry
			if _, err := file.WriteString(",\n"); err != nil {
				return fmt.Errorf("failed to write separator to trace file '%s': %w", filename, err)
			}
		}
	} else {
		// If file exists, add comma before new entry
		if _, err := file.WriteString(",\n"); err != nil {
			return fmt.Errorf("failed to write separator to trace file '%s': %w", file.Name(), err)
		}
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty print
	if err := encoder.Encode(entry); err != nil {
		return fmt.Errorf("failed to encode trace entry to file '%s': %w", file.Name(), err)
	}

	return nil
}

// GetTrace reads all trace entries for a session (simplified, reads whole file).
func (l *FileTraceLogger) GetTrace(sessionID string) ([]TraceEntry, error) {
	l.mu.Lock() // Lock needed if files map could change, or file could be closed
	defer l.mu.Unlock()

	filename := filepath.Join(l.logDir, fmt.Sprintf("%s.trace.json", sessionID))
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []TraceEntry{}, nil // Return empty slice if file doesn't exist
		}
		return nil, fmt.Errorf("failed to read trace file '%s': %w", filename, err)
	}

	// Need to make the potentially incomplete JSON valid by adding ']'
	if len(data) > 1 && data[len(data)-1] != ']' {
		data = append(data, ']')
	} else if len(data) == 1 && data[0] == '[' { // Handle case where only '[' was written
		data = append(data, ']')
	} else if len(data) == 0 { // Handle empty file
		return []TraceEntry{}, nil
	}

	var entries []TraceEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		// Log the problematic data for debugging
		// log.Printf("Error unmarshalling trace data from %s: %v\nData:\n%s", filename, err, string(data))
		return nil, fmt.Errorf("failed to unmarshal trace data from '%s': %w", filename, err)
	}

	return entries, nil
}

// Close closes all open trace files.
func (l *FileTraceLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	var firstErr error
	for id, file := range l.files {
		// Add closing bracket ']' to make JSON valid
		if stat, _ := file.Stat(); stat.Size() > 1 { // Avoid writing ']' to empty or just '[' files
			if _, err := file.WriteString("\n]"); err != nil && firstErr == nil {
				firstErr = fmt.Errorf("failed to write closing ']' to trace file '%s': %w", file.Name(), err)
			}
		} else if stat.Size() == 1 { // Handle case where only '[' was written
			if _, err := file.WriteString("]"); err != nil && firstErr == nil {
				firstErr = fmt.Errorf("failed to write closing ']' to trace file '%s': %w", file.Name(), err)
			}
		}

		if err := file.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("failed to close trace file '%s': %w", file.Name(), err)
		}
		delete(l.files, id)
	}
	return firstErr
}

// --- Trace Callback Functions ---

func createTraceCallback(logger TraceLogger) CallbackFunc {
	return func(ctx context.Context, args CallbackArgs) (State, error) {
		if logger == nil {
			return args.State, nil // No logger, do nothing
		}

		sessionID, ok := args.Event.GetMetadataValue(SessionIDKey)
		if !ok || sessionID == "" {
			Logger().Warn().
				Str("event_id", args.Event.GetID()).
				Msg("Cannot log trace entry, missing or empty session ID.")
			return args.State, nil
		}

		entry := TraceEntry{
			Timestamp:     time.Now(),
			SessionID:     sessionID,
			EventID:       args.Event.GetID(),
			Hook:          args.Hook,
			TargetAgentID: args.Event.GetTargetAgentID(),
			SourceAgentID: args.Event.GetSourceAgentID(),
			AgentID:       args.AgentID, // Include AgentID if available in args
		}

		// Add specific fields based on hook
		switch args.Hook {
		case HookAfterAgentRun:
			// FIX: Use args.AgentResult
			entry.AgentResult = &args.AgentResult
			if args.Error != nil { // Agent might succeed but still have an error in result? Check logic.
				entry.Error = args.Error.Error()
			}
		case HookAgentError:
			// FIX: Use args.AgentResult (might contain partial info)
			entry.AgentResult = &args.AgentResult
			if args.Error != nil {
				entry.Error = args.Error.Error()
			}
		}

		if err := logger.Log(entry); err != nil {
			Logger().Error().
				Err(err).
				Msg("Error logging trace entry")
		}
		// Return the state that should proceed
		// FIX: Check AgentResult.OutputState for the definitive state after agent run
		if args.Hook == HookAfterAgentRun && args.AgentResult.OutputState != nil {
			return args.AgentResult.OutputState, nil
		}
		return args.State, nil // Fallback to incoming state
	}
}

// RegisterTraceHooks registers the necessary callbacks for tracing.
func RegisterTraceHooks(registry *CallbackRegistry, logger TraceLogger) {
	if registry == nil || logger == nil {
		Logger().Warn().Msg("Cannot register trace hooks, registry or logger is nil.")
		return
	}
	traceCallback := createTraceCallback(logger)
	// FIX: Pass HookPoint constants directly
	registry.Register(HookBeforeEventHandling, "traceBeforeEvent", traceCallback)
	registry.Register(HookAfterEventHandling, "traceAfterEvent", traceCallback)
	registry.Register(HookBeforeAgentRun, "traceBeforeAgent", traceCallback)
	registry.Register(HookAfterAgentRun, "traceAfterAgent", traceCallback)
	registry.Register(HookAgentError, "traceAgentError", traceCallback)
	Logger().Debug().Msg("Registered trace hooks.")
}
