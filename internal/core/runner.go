package agentflow

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// EventHandler processes events. (Renamed from Agent)
type EventHandler interface {
	Handle(Event) error
}

// Runner manages the execution flow, routing events to registered agents.
type Runner interface {
	// Emit sends an event into the processing pipeline.
	Emit(event Event) error

	// RegisterAgent associates an agent name with a handler responsible for invoking it.
	RegisterAgent(name string, handler AgentHandler) error

	// RegisterCallback adds a named callback function for a specific hook point.
	RegisterCallback(hook HookPoint, name string, cb CallbackFunc) error

	// UnregisterCallback removes a named callback function from a specific hook point.
	UnregisterCallback(hook HookPoint, name string)

	// Start begins the event processing loop (non-blocking).
	Start(ctx context.Context) error

	// Stop gracefully shuts down the runner, waiting for active processing to complete.
	Stop()

	// GetCallbackRegistry returns the runner's callback registry.
	GetCallbackRegistry() *CallbackRegistry

	// GetTraceLogger returns the runner's trace logger.
	GetTraceLogger() TraceLogger

	// DumpTrace retrieves the trace entries for a specific session ID from the configured TraceLogger.
	DumpTrace(sessionID string) ([]TraceEntry, error)
}

// RunnerImpl implements the Runner interface.
type RunnerImpl struct {
	queue        chan Event
	orchestrator Orchestrator
	registry     *CallbackRegistry
	traceLogger  TraceLogger  // Use interface type
	tracer       trace.Tracer // OpenTelemetry tracer

	stopOnce sync.Once
	stopChan chan struct{}
	wg       sync.WaitGroup
	mu       sync.RWMutex // Protects access to orchestrator, traceLogger, started flag
	started  bool         // Flag to prevent multiple starts
}

// NewRunner creates a new RunnerImpl.
func NewRunner(queueSize int) *RunnerImpl {
	if queueSize <= 0 {
		queueSize = 100 // Default queue size
	}
	// FIX: Initialize registry here to avoid nil checks later
	return &RunnerImpl{
		queue:    make(chan Event, queueSize),
		stopChan: make(chan struct{}),
		registry: NewCallbackRegistry(), // Initialize registry
		// orchestrator is set via SetOrchestrator
		// traceLogger is set via SetTraceLogger
	}
}

// SetCallbackRegistry assigns the callback registry to the runner.
func (r *RunnerImpl) SetCallbackRegistry(registry *CallbackRegistry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.registry = registry
}

// SetOrchestrator assigns the orchestrator to the runner.
func (r *RunnerImpl) SetOrchestrator(o Orchestrator) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// FIX: Use 'started' field
	if r.started {
		Logger().Warn().Msg("Attempted to set orchestrator while runner is running.")
		return
	}
	r.orchestrator = o
}

// SetTraceLogger assigns the trace logger to the runner.
func (r *RunnerImpl) SetTraceLogger(logger TraceLogger) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// FIX: Use 'started' field
	if r.started {
		Logger().Warn().Msg("Attempted to set trace logger while runner is running.")
		return
	}
	r.traceLogger = logger
	// Optionally, re-register trace callbacks if they depend on the logger
	// r.registerTraceCallbacks() // Example
}

// GetTraceLogger returns the runner's trace logger.
func (r *RunnerImpl) GetTraceLogger() TraceLogger {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// FIX: Use the correct field name 'traceLogger'
	return r.traceLogger
}

// GetCallbackRegistry returns the runner's callback registry.
func (r *RunnerImpl) GetCallbackRegistry() *CallbackRegistry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Lazily initialize if nil? Or require SetCallbackRegistry?
	// For now, return potentially nil registry.
	return r.registry
}

// RegisterCallback delegates to the registry.
func (r *RunnerImpl) RegisterCallback(hook HookPoint, name string, cb CallbackFunc) error {
	r.mu.RLock()
	registry := r.registry // Registry is now guaranteed to be non-nil
	r.mu.RUnlock()
	// FIX: Pass HookPoint directly, ensure argument order matches registry.Register signature
	return registry.Register(hook, name, cb)
}

// UnregisterCallback delegates to the registry.
func (r *RunnerImpl) UnregisterCallback(hook HookPoint, name string) {
	r.mu.RLock()
	registry := r.registry // Registry is now guaranteed to be non-nil
	r.mu.RUnlock()
	// FIX: Pass HookPoint directly, ensure argument order matches registry.Unregister signature
	registry.Unregister(hook, name)
}

// RegisterAgent registers an AgentHandler with the underlying Orchestrator.
// FIX: Change handler type from EventHandler to AgentHandler
func (r *RunnerImpl) RegisterAgent(name string, handler AgentHandler) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.orchestrator == nil {
		return errors.New("orchestrator not set in runner")
	}
	// Assuming orchestrator's RegisterAgent also takes AgentHandler
	return r.orchestrator.RegisterAgent(name, handler)
}

// Emit adds an event to the processing queue.
// It blocks if the queue is full, up to a timeout.
func (r *RunnerImpl) Emit(event Event) error {
	r.mu.RLock()
	// FIX: Use 'started' field
	if !r.started {
		r.mu.RUnlock()
		Logger().Debug().Str("event_id", event.GetID()).Msg("Emit failed: runner not running")
		return errors.New("runner is not running")
	}
	r.mu.RUnlock()

	Logger().Debug().Str("event_id", event.GetID()).Msg("Emit attempting to queue event")

	// Timeout for trying to queue the event
	timeout := 1 * time.Second // Configurable?
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case r.queue <- event:
		Logger().Debug().Str("event_id", event.GetID()).Msg("Emit successfully queued event")
		return nil
	case <-ctx.Done():
		Logger().Debug().Str("event_id", event.GetID()).Msg("Emit timed out")
		return fmt.Errorf("failed to emit event: queue full or blocked")
	case <-r.stopChan: // Check if runner stopped while waiting
		Logger().Debug().Str("event_id", event.GetID()).Msg("Emit failed: runner stopped while waiting to queue")
		return errors.New("runner stopped while emitting")
	}
}

// Start begins the runner's event processing loop in a separate goroutine.
// It returns an error immediately if the runner is already started or not properly configured.
func (r *RunnerImpl) Start(ctx context.Context) error {
	r.mu.Lock()
	if r.started {
		r.mu.Unlock()
		return errors.New("runner already started")
	}
	// Ensure essential components are set before starting
	if r.orchestrator == nil {
		r.mu.Unlock()
		return errors.New("orchestrator must be set before starting runner")
	}
	// Add checks for registry, traceLogger etc. if they are mandatory

	r.started = true
	r.stopChan = make(chan struct{}) // Ensure stopChan is fresh for this run
	r.wg.Add(1)                      // Increment waitgroup for the loop goroutine
	r.mu.Unlock()

	Logger().Info().Msg("Runner started.")
	// Launch the main processing loop
	go r.loop(ctx) // Pass the main context

	// Start is non-blocking; it returns after launching the loop.
	return nil
}

// Stop signals the runner to stop processing events and waits for it to finish.
func (r *RunnerImpl) Stop() {
	r.mu.Lock()
	Logger().Debug().Msg("Runner Stop: Acquired lock")
	if !r.started {
		Logger().Info().Msg("Runner Stop: Already stopped, released lock.")
		r.mu.Unlock()
		return
	}

	Logger().Debug().Msg("Runner Stop: Setting started=false and closing stopChan...")
	r.started = false
	// Close stopChan *before* unlocking to ensure loop sees it if it checks started flag
	close(r.stopChan) // Signal the loop to exit
	Logger().Debug().Msg("Runner Stop: stopChan closed.")

	// Unlock *before* waiting to avoid deadlocks if loop needs the lock (e.g., for registry/orchestrator)
	r.mu.Unlock()
	Logger().Debug().Msg("Runner Stop: Released lock, waiting for loop goroutine (wg.Wait)...")

	// Wait for the loop goroutine to finish processing any in-flight event and exit
	r.wg.Wait()
	Logger().Debug().Msg("Runner Stop: Loop goroutine finished (wg.Wait returned).")

	// Stop the orchestrator after the loop has fully finished
	r.mu.RLock()
	orchestrator := r.orchestrator
	r.mu.RUnlock()
	if orchestrator != nil {
		Logger().Debug().Msg("Runner Stop: Stopping orchestrator...")
	}

	Logger().Info().Msg("Runner Stop: Completed.")
}

// loop is the main event processing goroutine.
func (r *RunnerImpl) loop(ctx context.Context) {
	defer r.wg.Done() // Ensure wg is decremented when loop exits
	for {
		select {
		case <-ctx.Done():
			Logger().Debug().Msg("Runner loop: Context cancelled. Exiting.")
			return // Exit loop
		case <-r.stopChan: // Listen for explicit stop signal
			Logger().Debug().Msg("Runner loop: Stop signal received. Exiting.")
			return // Exit loop
		case event := <-r.queue:
			// FIX: Defer eventCancel call
			eventCtx, eventCancel := context.WithCancel(ctx)
			defer eventCancel() // Ensure context is cancelled when processing for this event finishes

			// --- Session Handling ---
			sessionID, _ := event.GetMetadataValue(SessionIDKey)
			if sessionID == "" {
				sessionID = event.GetID()
				Logger().Warn().Str("event_id", event.GetID()).Msg("Runner loop: Warning - event missing session ID, using event ID as fallback.")
				event.SetMetadata(SessionIDKey, sessionID)
			}
			Logger().Debug().Str("event_id", event.GetID()).Str("session_id", sessionID).Msg("Runner loop: Processing event")

			// FIX: Initialize currentState explicitly as State interface type
			var currentState State = NewState() // Use interface type

			// --- Before Event Handling Callback ---
			if r.registry != nil {
				Logger().Debug().Msg("Runner: Invoking BeforeEventHandling callbacks")
				// FIX: Define callbackArgs before use
				callbackArgs := CallbackArgs{
					Hook:    HookBeforeEventHandling,
					Event:   event,
					State:   currentState,
					AgentID: "", // No specific agent yet
				}
				// FIX: Check and handle error from Invoke
				newState, err := r.registry.Invoke(eventCtx, callbackArgs)
				if err != nil {
					Logger().Error().Str("event_id", event.GetID()).Err(err).Msg("Runner loop: Error during BeforeEventHandling callbacks. Skipping event.")
					continue // Skip to next event
				}
				if newState != nil {
					currentState = newState // Assign State interface to State interface variable
				}
				Logger().Debug().Msg("CallbackRegistry.Invoke: Finished invoking callbacks for hook BeforeEventHandling.")
			}

			// --- Orchestration ---
			var agentResult AgentResult
			var agentErr error
			var invokedAgentID string

			// FIX: Get orchestrator safely
			r.mu.RLock()
			orchestrator := r.orchestrator
			r.mu.RUnlock()

			if orchestrator != nil {
				// --- Determine Target Agent ---
				targetAgentID := "unknown" // Default
				if routeKey, ok := event.GetMetadataValue(RouteMetadataKey); ok {
					targetAgentID = routeKey
				} else if event.GetTargetAgentID() != "" {
					targetAgentID = event.GetTargetAgentID()
				}
				invokedAgentID = targetAgentID // Store the determined agent ID

				// --- Before Agent Run Callback ---
				if r.registry != nil {
					Logger().Debug().Str("agent_id", invokedAgentID).Msgf("Runner: Invoking %s callbacks", HookBeforeAgentRun)
					callbackArgs := CallbackArgs{
						Hook:    HookBeforeAgentRun,
						Event:   event,
						State:   currentState, // Pass state *before* dispatch
						AgentID: invokedAgentID,
					}
					newState, err := r.registry.Invoke(eventCtx, callbackArgs)
					if err != nil {
						Logger().Error().Str("event_id", event.GetID()).Str("agent_id", invokedAgentID).Err(err).Msg("Runner loop: Error during BeforeAgentRun callbacks")
						agentErr = fmt.Errorf("BeforeAgentRun callback failed: %w", err) // Set agentErr
					} else {
						if newState != nil {
							currentState = newState // Assign State interface to State interface variable
						}
						Logger().Debug().Msg("CallbackRegistry.Invoke: Finished invoking callbacks for hook BeforeAgentRun.")
					}
				}

				// --- Dispatch (only if BeforeAgentRun didn't error) ---
				if agentErr == nil {
					Logger().Debug().Str("event_id", event.GetID()).Msg("Runner loop: Dispatching event to orchestrator")
					// FIX: Use the correct orchestrator variable
					agentResult, agentErr = orchestrator.Dispatch(eventCtx, event)
				}

				// --- After Agent Run / Agent Error Callbacks ---
				if agentErr != nil {
					// Dispatch failed OR BeforeAgentRun failed
					Logger().Error().Str("event_id", event.GetID()).Err(agentErr).Msg("Runner loop: Error during agent execution/dispatch")
					// --- Agent Error Callback ---
					if r.registry != nil {
						Logger().Debug().Str("agent_id", invokedAgentID).Msgf("Runner: Invoking %s callbacks", HookAgentError)
						callbackArgs := CallbackArgs{
							Hook:    HookAgentError,
							Event:   event,
							AgentID: invokedAgentID,
							Error:   agentErr,
							State:   currentState,
						}
						// FIX: Check and handle error from Invoke
						newState, cbErr := r.registry.Invoke(eventCtx, callbackArgs)
						if cbErr != nil {
							Logger().Error().Str("event_id", event.GetID()).Err(cbErr).Msg("Runner loop: Error during AgentError callback")
						}
						if newState != nil {
							currentState = newState // Assign State interface to State interface variable
						}
						Logger().Debug().Msg("CallbackRegistry.Invoke: Finished invoking callbacks for hook AgentError.")
					}
				}
			} else {
				Logger().Error().Str("event_id", event.GetID()).Msg("Runner loop: Orchestrator is nil, cannot dispatch event")
				agentErr = errors.New("orchestrator not configured")
				invokedAgentID = "orchestrator" // Indicate orchestrator issue
			}

			// --- Process Agent Result ---
			r.processAgentResult(eventCtx, event, agentResult, agentErr, invokedAgentID)

			// --- After Event Handling Callback ---
			if r.registry != nil {
				Logger().Debug().Msg("Runner: Invoking AfterEventHandling callbacks")
				finalStateForEvent := currentState
				if agentErr == nil && agentResult.OutputState != nil {
					finalStateForEvent = agentResult.OutputState
				}
				callbackArgs := CallbackArgs{
					Hook:    HookAfterEventHandling,
					Event:   event,
					State:   finalStateForEvent,
					AgentID: invokedAgentID,
					Error:   agentErr,
				}
				// FIX: Check and handle error from Invoke
				_, cbErr := r.registry.Invoke(eventCtx, callbackArgs)
				if cbErr != nil {
					Logger().Error().Str("event_id", event.GetID()).Err(cbErr).Msg("Runner loop: Error during AfterEventHandling callbacks")
				}
				Logger().Debug().Msg("CallbackRegistry.Invoke: Finished invoking callbacks for hook AfterEventHandling.")
			}

			Logger().Debug().Str("event_id", event.GetID()).Msg("Runner loop finished processing event")
			// eventCancel() is called by defer at the top of the case block
		}
	}
}

// processAgentResult handles the outcome of an agent execution, potentially emitting new events.
func (r *RunnerImpl) processAgentResult(ctx context.Context, originalEvent Event, result AgentResult, agentErr error, agentID string) {
	sessionID, _ := originalEvent.GetMetadataValue(SessionIDKey)

	if agentErr != nil {
		Logger().Error().
			Str("event_id", originalEvent.GetID()).
			Str("session_id", sessionID).
			Str("agent_id", agentID).
			Err(agentErr).
			Msg("Agent execution failed")

		// Optionally emit a failure event
		failurePayload := EventData{
			"original_event_id": originalEvent.GetID(),
			"error":             agentErr.Error(),
		}
		failureMeta := map[string]string{
			SessionIDKey:     sessionID,
			"status":         "failure",
			RouteMetadataKey: "error-handler", // Add this line
		}
		if agentID != "" && agentID != "unknown" {
			failureMeta["failed_agent_id"] = agentID
		}

		failureEvent := NewEvent(failureMeta[RouteMetadataKey], failurePayload, failureMeta)
		failureEvent.SetSourceAgentID(agentID)

		if err := r.Emit(failureEvent); err != nil {
			Logger().Error().
				Str("event_id", originalEvent.GetID()).
				Err(err).
				Msg("Error emitting failure event")
		}
	} else {
		Logger().Debug().
			Str("event_id", originalEvent.GetID()).
			Str("session_id", sessionID).
			Str("agent_id", agentID).
			Msg("Agent execution successful")

		if result.OutputState != nil {
			// Only emit a new event if the route is present and non-empty
			route, hasRoute := result.OutputState.GetMeta(RouteMetadataKey)
			if hasRoute && route != "" {
				successPayload := make(EventData)
				for _, key := range result.OutputState.Keys() {
					if val, ok := result.OutputState.Get(key); ok {
						successPayload[key] = val
					}
				}

				successMeta := make(map[string]string)
				for _, key := range result.OutputState.MetaKeys() {
					if val, ok := result.OutputState.GetMeta(key); ok {
						successMeta[key] = val
					}
				}
				successMeta[SessionIDKey] = sessionID
				successMeta["status"] = "success"

				successEvent := NewEvent(
					route, // Set the target agent if you want, or leave "" for orchestrator to decide
					successPayload,
					successMeta,
				)
				successEvent.SetSourceAgentID(agentID)
				if err := r.Emit(successEvent); err != nil {
					Logger().Error().
						Str("event_id", originalEvent.GetID()).
						Err(err).
						Msg("Error emitting success event")
				}
			} else {
				Logger().Debug().
					Str("event_id", originalEvent.GetID()).
					Msg("No route present in OutputState after agent execution; not emitting further event.")
			}
		} else {
			Logger().Debug().
				Str("event_id", originalEvent.GetID()).
				Str("session_id", sessionID).
				Str("agent_id", agentID).
				Msg("Agent execution successful, but no OutputState provided in AgentResult. No further event emitted.")
		}
	}
}

// DumpTrace retrieves trace entries.
func (r *RunnerImpl) DumpTrace(sessionID string) ([]TraceEntry, error) {
	r.mu.RLock()
	logger := r.traceLogger // Use correct field name
	r.mu.RUnlock()

	if logger == nil {
		return nil, errors.New("trace logger is not set")
	}
	return logger.GetTrace(sessionID)
}

// internal/core/runner.go (or a constants file)
const (
	callbackStateKeyAgentResult = "__agentResult"
	callbackStateKeyAgentError  = "__agentError"
)
