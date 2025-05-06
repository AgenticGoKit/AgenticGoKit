package factory

import (
	"context"
	"fmt"
	"log"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
)

// Helper function for registering trace callbacks
func registerTraceCallbacks(registry *agentflow.CallbackRegistry, logger agentflow.TraceLogger) {
	// Register callbacks for different hook points

	// Before Event Handling
	registry.Register(agentflow.HookBeforeEventHandling, "trace_before_event",
		func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
			sessionID := getSessionID(args.Event)
			// Log trace with relevant information for this hook
			logTrace(logger, "before_event", sessionID, args.Event, args.AgentID, nil, args.Error)
			return args.State, nil
		})

	// After Event Handling
	registry.Register(agentflow.HookAfterEventHandling, "trace_after_event",
		func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
			sessionID := getSessionID(args.Event)
			// Log trace with relevant information for this hook
			logTrace(logger, "after_event", sessionID, args.Event, args.AgentID, args.Result, args.Error)
			return args.State, nil
		})

	// Before Agent Run
	registry.Register(agentflow.HookBeforeAgentRun, "trace_before_agent_run",
		func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
			sessionID := getSessionID(args.Event)
			// Log trace with relevant information for this hook
			logTrace(logger, "before_agent_run", sessionID, args.Event, args.AgentID, nil, args.Error)
			return args.State, nil
		})

	// After Agent Run
	registry.Register(agentflow.HookAfterAgentRun, "trace_after_agent_run",
		func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
			sessionID := getSessionID(args.Event)
			// Log trace with relevant information for this hook
			logTrace(logger, "after_agent_run", sessionID, args.Event, args.AgentID, args.Result, args.Error)
			return args.State, nil
		})

	// Before LLM Call
	registry.Register(agentflow.HookBeforeLLMCall, "trace_before_llm_call",
		func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
			sessionID := getSessionID(args.Event)
			// Log trace with relevant information for this hook
			// Assuming AgentID is available in CallbackArgs for LLM hooks if applicable
			logTrace(logger, "before_llm_call", sessionID, args.Event, args.AgentID, nil, args.Error)
			return args.State, nil
		})

	// After LLM Call
	registry.Register(agentflow.HookAfterLLMCall, "trace_after_llm_call",
		func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
			sessionID := getSessionID(args.Event)
			// Log trace with relevant information for this hook
			// Assuming AgentID and Result are available for LLM hooks
			logTrace(logger, "after_llm_call", sessionID, args.Event, args.AgentID, args.Result, args.Error)
			return args.State, nil
		})

}

// Safe helper for logging traces, matching the actual TraceLogger interface
func logTrace(logger agentflow.TraceLogger, entryType, sessionID string,
	event agentflow.Event, agentID string, result *agentflow.AgentResult, err error) {

	entry := agentflow.TraceEntry{
		SessionID: sessionID,
		EventID:   event.GetID(),
		AgentID:   agentID,
		Type:      entryType,
		Timestamp: time.Now(),
	}

	if err != nil {
		// Log the full error details
		entry.Error = fmt.Sprintf("%+v", err)
	}

	if result != nil {
		// Include the result details in the trace entry
		// Using %+v to get detailed struct representation
		entry.Result = fmt.Sprintf("%+v", result)
	}

	logger.Log(entry)
}

// Safe helper to get session ID from an event
func getSessionID(event agentflow.Event) string {
	// Attempt to get the session ID from event metadata
	sessionID, ok := event.GetMetadataValue(agentflow.SessionIDKey)
	if !ok || sessionID == "" {
		return event.GetID()
	}
	return sessionID
}