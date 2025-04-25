package agentflow

import (
	"context"
	"errors" // Ensure fmt is imported
	"log"
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
	Start()

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

// NewRunner creates a Runner with a specific queue size.
// It initializes with a NoOpTraceLogger by default.
func NewRunner(queueSize int) *RunnerImpl {
	if queueSize <= 0 {
		queueSize = 100 // Default queue size
	}
	// Initialize tracer provider (replace with actual setup if needed)
	// tracerProvider := trace.NewNoopTracerProvider() // Example No-op provider

	return &RunnerImpl{
		queue:    make(chan Event, queueSize),
		registry: NewCallbackRegistry(),
		stopChan: make(chan struct{}),
		// FIX: Initialize traceLogger with the newly defined NoOp logger
		traceLogger: NewNoOpTraceLogger(),
		// tracer:      tracerProvider.Tracer("agentflow/core/runner"), // Initialize tracer
	}
}

// SetOrchestrator assigns the orchestrator to the runner.
func (r *RunnerImpl) SetOrchestrator(o Orchestrator) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// FIX: Use 'started' field
	if r.started {
		log.Println("Warning: Attempted to set orchestrator while runner is running.")
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
		log.Println("Warning: Attempted to set trace logger while runner is running.")
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
	// Registry is typically set at creation and not changed, so RLock might be sufficient
	// If registry itself can be replaced, use full Lock/RUnlock
	return r.registry
}

// RegisterCallback delegates to the registry.
func (r *RunnerImpl) RegisterCallback(hook HookPoint, name string, cb CallbackFunc) error {
	if r.registry == nil {
		return errors.New("callback registry is not initialized")
	}
	return r.registry.Register(hook, name, cb)
}

// UnregisterCallback delegates to the registry.
func (r *RunnerImpl) UnregisterCallback(hook HookPoint, name string) {
	if r.registry != nil {
		r.registry.Unregister(hook, name)
	}
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

// Emit enqueues an Event for delivery. Returns error if runner is stopped.
func (r *RunnerImpl) Emit(event Event) error {
	r.mu.RLock()
	// FIX: Use 'started' field
	if !r.started {
		r.mu.RUnlock()
		// TEMP LOG
		log.Printf("DEBUG: Emit failed for event %s: runner not running", event.GetID())
		return errors.New("runner is not running")
	}
	r.mu.RUnlock()

	// TEMP LOG
	log.Printf("DEBUG: Emit attempting to queue event %s...", event.GetID())

	select {
	case r.queue <- event:
		// TEMP LOG
		log.Printf("DEBUG: Emit successfully queued event %s", event.GetID())
		return nil
	case <-time.After(1 * time.Second): // Timeout is 1 second
		// TEMP LOG
		log.Printf("DEBUG: Emit timed out for event %s", event.GetID())
		return errors.New("failed to emit event: queue full or blocked") // Returns error on timeout
	case <-r.stopChan:
		// TEMP LOG
		log.Printf("DEBUG: Emit failed for event %s: runner stopped during select", event.GetID())
		return errors.New("failed to emit event: runner is stopped")
	}
}

// Start begins the event processing loop.
func (r *RunnerImpl) Start(ctx context.Context) error {
	r.mu.Lock()
	// FIX: Use 'started' field
	if r.started {
		r.mu.Unlock()
		return errors.New("runner is already running")
	}
	if r.orchestrator == nil {
		r.mu.Unlock()
		return errors.New("orchestrator must be set before starting")
	}
	// FIX: Use 'started' field
	r.started = true
	// Recreate stopChan in case Start is called after Stop
	// Ensure previous one is not leaked if Start is called multiple times (though prevented by r.running check)
	r.stopChan = make(chan struct{})
	r.wg.Add(1) // Increment WaitGroup counter BEFORE starting goroutine
	go r.loop(ctx)
	r.mu.Unlock()                  // Unlock after starting goroutine
	log.Println("Runner started.") // Existing log
	return nil
}

// Stop gracefully shuts down the runner.
func (r *RunnerImpl) Stop() {
	r.mu.Lock()
	log.Println("Runner Stop: Acquired lock") // Add log
	// FIX: Use 'started' field
	if !r.started {
		r.mu.Unlock()
		log.Println("Runner Stop: Already stopped, released lock.") // Add log
		return
	}
	log.Println("Runner Stop: Setting started=false and closing stopChan...") // Add log
	// FIX: Use 'started' field
	r.started = false
	// Ensure stopChan is not nil before closing
	if r.stopChan != nil {
		// Check if already closed to prevent panic
		select {
		case <-r.stopChan:
			log.Println("Runner Stop: stopChan was already closed.")
		default:
			close(r.stopChan)
			log.Println("Runner Stop: stopChan closed.") // Add log
		}
	} else {
		log.Println("Runner Stop: stopChan was nil.") // Add log
	}
	r.mu.Unlock()                                                                      // Unlock before waiting
	log.Println("Runner Stop: Released lock, waiting for loop goroutine (wg.Wait)...") // Add log

	// Wait for the loop goroutine to finish
	r.wg.Wait()
	log.Println("Runner Stop: Loop goroutine finished (wg.Wait returned).") // Add log

	// Stop the orchestrator AFTER the runner loop has finished
	if r.orchestrator != nil {
		log.Println("Runner Stop: Stopping orchestrator...") // Add log
		r.orchestrator.Stop()
		log.Println("Runner Stop: Orchestrator stopped.") // Add log
	}

	log.Println("Runner Stop: Completed.") // Changed log message
}

// loop is the main event processing goroutine.
func (r *RunnerImpl) loop(ctx context.Context) {
	defer r.wg.Done()
	log.Println("Runner loop started.")        // Existing log
	defer log.Println("Runner loop finished.") // Add defer log for exit confirmation

	for {
		log.Println("Runner loop: top of select") // Add log
		select {
		case <-ctx.Done(): // Context cancellation
			log.Println("Runner loop exiting due to context cancellation.")
			return
		case <-r.stopChan: // Explicit stop signal
			log.Println("Runner loop exiting due to stop signal.")
			return
		case event, ok := <-r.queue:
			log.Printf("Runner loop: received from queue (ok=%t)", ok) // Add log
			if !ok {
				log.Println("Runner loop exiting because queue was closed.")
				return // Should not happen if Stop is used correctly
			}
			log.Printf("Runner loop processing event %s", event.GetID()) // Existing log

			// --- Simulate processing ---
			// FIX: Call registry.Invoke for before hook
			if r.registry != nil {
				beforeArgs := CallbackArgs{Ctx: ctx, Hook: HookBeforeEventHandling, Event: &event, CurrentState: nil} // Assuming state is not tracked here
				r.registry.Invoke(beforeArgs)
			}

			if r.orchestrator != nil {
				log.Printf("Runner loop: calling orchestrator.Dispatch for event %s", event.GetID()) // Add log
				err := r.orchestrator.Dispatch(event)                                                // This blocks in the test
				if err != nil {
					log.Printf("Error dispatching event %s: %v", event.GetID(), err)
				}
				log.Printf("Runner loop: orchestrator.Dispatch returned for event %s", event.GetID()) // Add log
			} else {
				log.Printf("Runner loop: orchestrator is nil, skipping dispatch for event %s", event.GetID())
			}

			// FIX: Call registry.Invoke for after hook
			if r.registry != nil {
				afterArgs := CallbackArgs{Ctx: ctx, Hook: HookAfterEventHandling, Event: &event, CurrentState: nil} // Assuming state is not tracked here
				r.registry.Invoke(afterArgs)
			}
			// --- End simulate processing ---

			log.Printf("Runner loop finished processing event %s", event.GetID()) // Existing log
		}
	}
}

// DumpTrace retrieves trace entries.
func (r *RunnerImpl) DumpTrace(sessionID string) ([]TraceEntry, error) {
	r.mu.RLock()
	// FIX: Use the correct field name 'traceLogger'
	logger := r.traceLogger
	r.mu.RUnlock()

	if logger == nil {
		return nil, errors.New("trace logger is not set")
	}
	return logger.GetTrace(sessionID)
}

// processAgentResult handles the outcome of an agent execution, potentially emitting new events.
func (r *RunnerImpl) processAgentResult(ctx context.Context, originalEvent Event, result AgentResult) {
	// ... (tracing code remains the same) ...

	// Check for errors in the result
	if result.Error != "" {
		log.Printf("Agent execution failed for event %s: %s", originalEvent.GetID(), result.Error)
		// Optionally emit a failure event
		// FIX: Create an actual Event object
		failurePayload := EventData{
			"original_event_id": originalEvent.GetID(),
			"error":             result.Error,
		}
		failureMeta := map[string]string{
			SessionIDKey: originalEvent.GetMetadata()[SessionIDKey], // Preserve session ID
			"status":     "failure",
		}
		// Assuming NewEvent creates *SimpleEvent which implements Event
		failureEvent := NewEvent("", failurePayload, failureMeta)       // Use constructor
		failureEvent.SetSourceAgentID(originalEvent.GetTargetAgentID()) // Set source as the agent that failed

		// Emit the failure event
		if err := r.Emit(failureEvent); err != nil {
			log.Printf("Error emitting failure event for original event %s: %v", originalEvent.GetID(), err)
		}
		// span.SetStatus(codes.Error, result.Error) // Already done above
	} else {
		log.Printf("Agent execution successful for event %s", originalEvent.GetID())
		// Optionally emit a success event or an event carrying the result state
		// FIX: Create an actual Event object
		successPayload := EventData{
			"original_event_id": originalEvent.GetID(),
			// Include output state data if needed
			// "output_data": (*result.OutputState).GetData(), // Be careful with size/complexity
		}
		successMeta := map[string]string{
			SessionIDKey: originalEvent.GetMetadata()[SessionIDKey], // Preserve session ID
			"status":     "success",
		}
		// Copy metadata from output state if necessary
		if result.OutputState != nil && *result.OutputState != nil {
			if stateWithMeta, ok := (*result.OutputState).(interface{ GetMetadata() map[string]string }); ok {
				// FIX: Assign map to variable first, then loop
				metaMap := stateWithMeta.GetMetadata() // Get the map
				if metaMap != nil {                    // Check if the map is not nil before ranging
					for k, v := range metaMap { // Loop over the variable
						successMeta[k] = v // Merge/overwrite metadata
					}
				}
			} else {
				log.Printf("Warning: Output state type %T does not have GetMetadata method", *result.OutputState)
			}
		}

		successEvent := NewEvent("", successPayload, successMeta)       // Use constructor
		successEvent.SetSourceAgentID(originalEvent.GetTargetAgentID()) // Set source as the agent that succeeded

		// Emit the success event
		if err := r.Emit(successEvent); err != nil {
			log.Printf("Error emitting success event for original event %s: %v", originalEvent.GetID(), err)
		}
		// span.SetStatus(codes.Ok, "Agent executed successfully") // Already done above
	}
}
