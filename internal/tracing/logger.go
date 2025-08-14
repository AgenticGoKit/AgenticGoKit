// Package tracing provides internal tracing implementations for AgentFlow.
package tracing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// InMemoryTraceLogger is a simple in-memory implementation of TraceLogger.
type InMemoryTraceLogger struct {
	mu     sync.RWMutex
	traces map[string][]core.TraceEntry
}

// NewInMemoryTraceLogger creates a new in-memory trace logger.
func NewInMemoryTraceLogger() *InMemoryTraceLogger {
	return &InMemoryTraceLogger{
		traces: make(map[string][]core.TraceEntry),
	}
}

// Log adds a trace entry to the logger.
func (l *InMemoryTraceLogger) Log(entry core.TraceEntry) error {
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
func (l *InMemoryTraceLogger) GetTrace(sessionID string) ([]core.TraceEntry, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if sessionID == "" {
		sessionID = "default"
	}

	entries, exists := l.traces[sessionID]
	if !exists {
		return []core.TraceEntry{}, nil
	}

	// Return a copy to avoid race conditions
	result := make([]core.TraceEntry, len(entries))
	copy(result, entries)
	return result, nil
}

// RegisterTraceHooks registers tracing callbacks with the callback registry.
func RegisterTraceHooks(registry *core.CallbackRegistry, logger core.TraceLogger) error {
	if registry == nil {
		return fmt.Errorf("callback registry cannot be nil")
	}
	if logger == nil {
		return fmt.Errorf("trace logger cannot be nil")
	}

	// Register trace callbacks for each hook point
	err := registry.Register(core.HookBeforeEventHandling, "trace_before_event", func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
		entry := core.TraceEntry{
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

	err = registry.Register(core.HookAfterEventHandling, "trace_after_event", func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
		entry := core.TraceEntry{
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