// Package core provides essential factory functions for creating agents and runners in AgentFlow.
package core

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	Logger().Debug().
		Str("callback", name).
		Str("hook", string(hook)).
		Msg("Callback registered")
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
			Logger().Info().
				Str("callback", name).
				Str("hook", string(hook)).
				Msg("Callback unregistered")
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

		Logger().Debug().
			Str("hook", string(args.Hook)).
			Msg("Executing callback")

		returnedState, err := callback(ctx, currentArgs)
		if err != nil {
			Logger().Error().
				Str("hook", string(args.Hook)).
				Err(err).
				Msg("Error executing callback")
			lastErr = err
		}

		if returnedState != nil {
			Logger().Debug().
				Str("hook", string(args.Hook)).
				Msg("Callback returned updated state")
			currentState = returnedState
		} else {
			Logger().Debug().
				Str("hook", string(args.Hook)).
				Msg("Callback returned nil state, state remains unchanged")
		}
	}

	Logger().Debug().
		Str("hook", string(args.Hook)).
		Msg("Finished invoking callbacks, returning final state")
	return currentState, lastErr
}

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var (
	logger   zerolog.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	logLevel LogLevel       = INFO
	mu       sync.RWMutex
)

func SetLogLevel(level LogLevel) {
	mu.Lock()
	defer mu.Unlock()
	logLevel = level
	zerolog.SetGlobalLevel(mapLogLevel(level))
}

func GetLogLevel() LogLevel {
	mu.RLock()
	defer mu.RUnlock()
	return logLevel
}

func Logger() *zerolog.Logger {
	return &logger
}

func mapLogLevel(level LogLevel) zerolog.Level {
	switch level {
	case DEBUG:
		return zerolog.DebugLevel
	case INFO:
		return zerolog.InfoLevel
	case WARN:
		return zerolog.WarnLevel
	case ERROR:
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

// TraceEntry represents a single logged event during the execution flow.
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

// NewInMemoryTraceLogger creates a new in-memory trace logger
func NewInMemoryTraceLogger() TraceLogger {
	// TODO: This will be replaced with internal implementation after refactoring is complete
	return &noOpTraceLogger{}
}

// RegisterTraceHooks registers tracing callbacks with the callback registry
func RegisterTraceHooks(registry *CallbackRegistry, logger TraceLogger) error {
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

// Temporary no-op implementation during refactoring
type noOpTraceLogger struct{}

func (l *noOpTraceLogger) Log(entry TraceEntry) error {
	return nil
}

func (l *noOpTraceLogger) GetTrace(sessionID string) ([]TraceEntry, error) {
	return []TraceEntry{}, nil
}

// TODO: Add MCP-enabled factory functions here once the infrastructure is implemented
// These will be moved from internal packages to provide a clean public API
