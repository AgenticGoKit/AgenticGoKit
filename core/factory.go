// Package core provides essential factory functions for creating agents and runners in AgentFlow.
package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Essential factory functions for agents and runners
// Implementation details are moved to internal packages

// Essential factory functions are defined in their respective files
// NewAgent is defined in agent_builder.go
// NewRunner is defined in runner.go

// RouteMetadataKey defines the metadata key used for routing events to specific agents.
const RouteMetadataKey = "route"

// HookPoint defines specific points in the execution flow where callbacks can be triggered.
type HookPoint string

const (
	HookBeforeEventHandling HookPoint = "BeforeEventHandling"
	HookAfterEventHandling  HookPoint = "AfterEventHandling"
	HookBeforeAgentRun      HookPoint = "BeforeAgentRun"
	HookAfterAgentRun       HookPoint = "AfterAgentRun"
	HookAgentError          HookPoint = "AgentError"
	HookAll                 HookPoint = "AllHooks"
)

// CallbackArgs encapsulates all arguments passed to a callback function.
type CallbackArgs struct {
	Ctx         context.Context
	Hook        HookPoint
	Event       Event
	State       State
	AgentID     string
	AgentResult AgentResult
	Error       error
}

// CallbackFunc defines the signature for callback functions.
type CallbackFunc func(ctx context.Context, args CallbackArgs) (State, error)

// CallbackRegistration holds details about a registered callback.
type CallbackRegistration struct {
	ID           string
	Hook         HookPoint
	CallbackFunc CallbackFunc
	AgentName    string
}

// CallbackRegistry manages registered callback functions.
type CallbackRegistry struct {
	mu        sync.RWMutex
	callbacks map[HookPoint][]*CallbackRegistration
}

// NewCallbackRegistry creates a new callback registry.
func NewCallbackRegistry() *CallbackRegistry {
	return &CallbackRegistry{
		callbacks: make(map[HookPoint][]*CallbackRegistration),
	}
}

// Register adds a callback function for a specific hook point.
func (r *CallbackRegistry) Register(hook HookPoint, name string, cb CallbackFunc) error {
	if name == "" {
		return fmt.Errorf("callback name cannot be empty")
	}
	if cb == nil {
		return fmt.Errorf("callback function cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	registration := &CallbackRegistration{
		ID:           name,
		Hook:         hook,
		CallbackFunc: cb,
	}

	for _, existing := range r.callbacks[hook] {
		if existing.ID == name {
			return fmt.Errorf("callback '%s' already registered for hook '%s'", name, hook)
		}
	}

	r.callbacks[hook] = append(r.callbacks[hook], registration)
	// Only log callback registration in debug mode
	if GetLogLevel() == DEBUG {
		Logger().Debug().Str("callback", name).Str("hook", string(hook)).Msg("Callback registered")
	}
	return nil
}

// Unregister removes a callback function.
func (r *CallbackRegistry) Unregister(hook HookPoint, name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	hooks := r.callbacks[hook]
	for i, reg := range hooks {
		if reg.ID == name {
			r.callbacks[hook] = append(hooks[:i], hooks[i+1:]...)
			if GetLogLevel() == DEBUG {
				Logger().Debug().Str("callback", name).Str("hook", string(hook)).Msg("Callback unregistered")
			}
			return
		}
	}
	Logger().Warn().
		Str("callback", name).
		Str("hook", string(hook)).
		Msg("Callback not found during unregister")
}

// Invoke calls all registered callbacks for a specific hook and HookAll.
func (r *CallbackRegistry) Invoke(ctx context.Context, args CallbackArgs) (State, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	currentState := args.State
	if currentState == nil {
		currentState = NewState()
		Logger().Warn().
			Str("hook", string(args.Hook)).
			Msg("Initial state was nil, created new State")
	}

	hookRegistrations := r.callbacks[args.Hook]
	allRegistrations := r.callbacks[HookAll]

	callbacksToRun := make([]CallbackFunc, 0, len(hookRegistrations)+len(allRegistrations))
	for _, reg := range hookRegistrations {
		if reg != nil {
			callbacksToRun = append(callbacksToRun, reg.CallbackFunc)
		}
	}
	for _, reg := range allRegistrations {
		if reg != nil {
			callbacksToRun = append(callbacksToRun, reg.CallbackFunc)
		}
	}

	var lastErr error
	for _, callback := range callbacksToRun {
		currentArgs := args
		currentArgs.State = currentState

		// Reduce per-callback execution logging
		if GetLogLevel() == DEBUG {
			Logger().Debug().Str("hook", string(args.Hook)).Msg("Executing callback")
		}

		returnedState, err := callback(ctx, currentArgs)
		if err != nil {
			Logger().Error().
				Str("hook", string(args.Hook)).
				Err(err).
				Msg("Error executing callback")
			lastErr = err
		}

		if returnedState != nil {
			if GetLogLevel() == DEBUG {
				Logger().Debug().Str("hook", string(args.Hook)).Msg("Callback updated state")
			}
			currentState = returnedState
		} else {
			if GetLogLevel() == DEBUG {
				Logger().Debug().Str("hook", string(args.Hook)).Msg("Callback state unchanged")
			}
		}
	}

	// Only log callback completion in debug mode
	if GetLogLevel() == DEBUG {
		Logger().Debug().Str("hook", string(args.Hook)).Msg("Callbacks complete")
	}
	return currentState, lastErr
}

// Essential logging interface
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// track current log level even if no provider is active (for tests and defaults)
var currentLogLevel LogLevel = INFO

// Essential logging functions - implementations moved to internal packages
func SetLogLevel(level LogLevel) {
	// Delegate to active logging provider if available
	p := getActiveLoggingProvider()
	if p.SetLevel != nil {
		p.SetLevel(level)
	}
	// Always record the desired level locally
	currentLogLevel = level
}

func GetLogLevel() LogLevel {
	// Delegate to active logging provider if available
	p := getActiveLoggingProvider()
	if p.GetLevel != nil {
		return p.GetLevel()
	}
	// Fallback to locally tracked level
	return currentLogLevel
}

func Logger() CoreLogger {
	// Return logger from the active provider or a safe no-op
	p := getActiveLoggingProvider()
	if p.New != nil {
		return p.New()
	}
	return &noopCoreLogger{}
}

// CoreLogger interface for essential logging operations
type CoreLogger interface {
	Debug() LogEvent
	Info() LogEvent
	Warn() LogEvent
	Error() LogEvent
	With() LogEvent
}

// LogEvent interface for building log messages
type LogEvent interface {
	Str(key, val string) LogEvent
	Strs(key string, val []string) LogEvent
	Int(key string, val int) LogEvent
	Bool(key string, val bool) LogEvent
	Float64(key string, val float64) LogEvent
	Dur(key string, val time.Duration) LogEvent
	Time(key string, val time.Time) LogEvent
	Interface(key string, val interface{}) LogEvent
	Err(err error) LogEvent
	Msg(msg string)
	Msgf(format string, args ...interface{})

	// For chaining - allow LogEvent to also behave like a logger
	Debug() LogEvent
	Info() LogEvent
	Warn() LogEvent
	Error() LogEvent

	// Logger creation for With() pattern
	Logger() CoreLogger
}

// =============================================================================
// LOGGING PROVIDER REGISTRY (Plugins register here)
// =============================================================================

// LoggingProvider wires a concrete logger into core.
type LoggingProvider struct {
	// New returns a new CoreLogger instance (can share underlying sink).
	New func() CoreLogger
	// SetLevel sets global log level (optional).
	SetLevel func(LogLevel)
	// GetLevel gets global log level (optional).
	GetLevel func() LogLevel
}

var (
	loggingProvidersMu    sync.RWMutex
	loggingProviders      = map[string]LoggingProvider{}
	activeLoggingProvider string
)

// RegisterLoggingProvider registers a logging provider by name. First registered becomes active by default.
func RegisterLoggingProvider(name string, provider LoggingProvider) {
	if name == "" || provider.New == nil {
		return
	}
	loggingProvidersMu.Lock()
	defer loggingProvidersMu.Unlock()
	loggingProviders[name] = provider
	if activeLoggingProvider == "" {
		activeLoggingProvider = name
	}
}

// UseLoggingProvider selects an already-registered provider by name. Returns true if switched.
func UseLoggingProvider(name string) bool {
	loggingProvidersMu.Lock()
	defer loggingProvidersMu.Unlock()
	if _, ok := loggingProviders[name]; ok {
		activeLoggingProvider = name
		return true
	}
	return false
}

func getActiveLoggingProvider() LoggingProvider {
	loggingProvidersMu.RLock()
	name := activeLoggingProvider
	provider, ok := loggingProviders[name]
	loggingProvidersMu.RUnlock()
	if ok {
		return provider
	}
	return LoggingProvider{New: func() CoreLogger { return &noopCoreLogger{} }}
}

// Safe no-op implementations to avoid nil panics before a provider is registered
type noopCoreLogger struct{}

func (l *noopCoreLogger) Debug() LogEvent { return &noopLogEvent{} }
func (l *noopCoreLogger) Info() LogEvent  { return &noopLogEvent{} }
func (l *noopCoreLogger) Warn() LogEvent  { return &noopLogEvent{} }
func (l *noopCoreLogger) Error() LogEvent { return &noopLogEvent{} }
func (l *noopCoreLogger) With() LogEvent  { return &noopLogEvent{} }

type noopLogEvent struct{}

func (e *noopLogEvent) Str(key, val string) LogEvent                   { return e }
func (e *noopLogEvent) Strs(key string, val []string) LogEvent         { return e }
func (e *noopLogEvent) Int(key string, val int) LogEvent               { return e }
func (e *noopLogEvent) Bool(key string, val bool) LogEvent             { return e }
func (e *noopLogEvent) Float64(key string, val float64) LogEvent       { return e }
func (e *noopLogEvent) Dur(key string, val time.Duration) LogEvent     { return e }
func (e *noopLogEvent) Time(key string, val time.Time) LogEvent        { return e }
func (e *noopLogEvent) Interface(key string, val interface{}) LogEvent { return e }
func (e *noopLogEvent) Err(err error) LogEvent                         { return e }
func (e *noopLogEvent) Msg(msg string)                                 {}
func (e *noopLogEvent) Msgf(format string, args ...interface{})        {}
func (e *noopLogEvent) Debug() LogEvent                                { return e }
func (e *noopLogEvent) Info() LogEvent                                 { return e }
func (e *noopLogEvent) Warn() LogEvent                                 { return e }
func (e *noopLogEvent) Error() LogEvent                                { return e }
func (e *noopLogEvent) Logger() CoreLogger                             { return &noopCoreLogger{} }

// Essential tracing types and interfaces
type TraceEntry struct {
	Timestamp     time.Time    `json:"timestamp"`
	Type          string       `json:"type"`
	EventID       string       `json:"event_id"`
	SessionID     string       `json:"session_id"`
	AgentID       string       `json:"agent_id,omitempty"`
	State         State        `json:"state,omitempty"`
	Error         string       `json:"error,omitempty"`
	Hook          HookPoint    `json:"hook,omitempty"`
	TargetAgentID string       `json:"target_agent_id,omitempty"`
	SourceAgentID string       `json:"source_agent_id,omitempty"`
	AgentResult   *AgentResult `json:"agent_result,omitempty"`
}

// TraceLogger defines the interface for storing and retrieving trace entries.
type TraceLogger interface {
	Log(entry TraceEntry) error
	GetTrace(sessionID string) ([]TraceEntry, error)
}

// Essential tracing factory functions - implementations moved to internal packages
func NewInMemoryTraceLogger() TraceLogger {
	// For now, return a simple implementation during refactoring
	return &inMemoryTraceLogger{
		traces: make(map[string][]TraceEntry),
	}
}

func RegisterTraceHooks(registry *CallbackRegistry, logger TraceLogger) error {
	// Implementation moved to bridge pattern to avoid circular dependencies
	if registry == nil {
		return fmt.Errorf("callback registry cannot be nil")
	}
	if logger == nil {
		return fmt.Errorf("trace logger cannot be nil")
	}

	// Register trace callbacks for each hook point
	err := registry.Register(HookBeforeEventHandling, "trace_before_event", func(ctx context.Context, args CallbackArgs) (State, error) {
		entry := TraceEntry{
			Timestamp:     time.Now(),
			Type:          "event_start",
			EventID:       args.Event.GetID(),
			SessionID:     args.Event.GetSessionID(),
			Hook:          args.Hook,
			TargetAgentID: args.Event.GetTargetAgentID(),
			SourceAgentID: args.Event.GetSourceAgentID(),
			State:         args.State,
		}
		logger.Log(entry)
		return args.State, nil
	})
	if err != nil {
		return fmt.Errorf("failed to register before event trace hook: %w", err)
	}

	err = registry.Register(HookAfterEventHandling, "trace_after_event", func(ctx context.Context, args CallbackArgs) (State, error) {
		entry := TraceEntry{
			Timestamp: time.Now(),
			Type:      "event_end",
			EventID:   args.Event.GetID(),
			SessionID: args.Event.GetSessionID(),
			Hook:      args.Hook,
			AgentID:   args.AgentID,
			State:     args.State,
		}
		if args.Error != nil {
			entry.Error = args.Error.Error()
		}
		logger.Log(entry)
		return args.State, nil
	})
	if err != nil {
		return fmt.Errorf("failed to register after event trace hook: %w", err)
	}
	return nil
}

// Simple in-memory implementation for core package during refactoring
type inMemoryTraceLogger struct {
	mu     sync.RWMutex
	traces map[string][]TraceEntry
}

func (l *inMemoryTraceLogger) Log(entry TraceEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	sessionID := entry.SessionID
	if sessionID == "" {
		sessionID = "default"
	}
	l.traces[sessionID] = append(l.traces[sessionID], entry)
	return nil
}

func (l *inMemoryTraceLogger) GetTrace(sessionID string) ([]TraceEntry, error) {
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

// TODO: Add MCP-enabled factory functions here once the infrastructure is implemented
// These will be moved from internal packages to provide a clean public API
// =============================================================================
// VISUALIZATION SUPPORT
// =============================================================================

// MermaidConfig configures diagram generation options
type MermaidConfig struct {
	DiagramType    string
	Title          string
	Direction      string // TB (top-bottom), LR (left-right), etc.
	Theme          string // default, dark, forest, etc.
	ShowMetadata   bool   // Include metadata like timeouts, error strategies
	ShowAgentTypes bool   // Show agent type information
	CompactMode    bool   // Generate more compact diagrams
}

// DefaultMermaidConfig returns sensible defaults for Mermaid diagram generation
func DefaultMermaidConfig() MermaidConfig {
	return MermaidConfig{
		DiagramType:    "flowchart",
		Direction:      "TD", // Top-Down
		Theme:          "default",
		ShowMetadata:   true,
		ShowAgentTypes: true,
		CompactMode:    false,
	}
}

// MermaidGenerator interface for generating Mermaid diagrams
type MermaidGenerator interface {
	GenerateCompositionDiagram(mode, name string, agents []Agent, config MermaidConfig) string
}

// NewMermaidGenerator creates a new Mermaid generator
// Implementation is provided by internal packages
func NewMermaidGenerator() MermaidGenerator {
	if mermaidGeneratorFactory != nil {
		return mermaidGeneratorFactory()
	}
	// Return a basic implementation
	return &basicMermaidGenerator{}
}

// RegisterMermaidGeneratorFactory registers the Mermaid generator factory function
func RegisterMermaidGeneratorFactory(factory func() MermaidGenerator) {
	mermaidGeneratorFactory = factory
}

var mermaidGeneratorFactory func() MermaidGenerator

// basicMermaidGenerator provides a minimal implementation
type basicMermaidGenerator struct{}

func (g *basicMermaidGenerator) GenerateCompositionDiagram(mode, name string, agents []Agent, config MermaidConfig) string {
	// Basic implementation - internal packages can provide more sophisticated implementations
	return fmt.Sprintf("graph %s\n    %s[%s]\n", config.Direction, name, mode)
}
