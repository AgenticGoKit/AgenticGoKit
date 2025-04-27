package agentflow

import (
	"context"
	"errors"
	"fmt"
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

// NewRunner creates a new RunnerImpl.
func NewRunner(queueSize int) *RunnerImpl {
	if queueSize <= 0 {
		queueSize = 100 // Default queue size
	}
	return &RunnerImpl{
		queue:    make(chan Event, queueSize),
		stopChan: make(chan struct{}),
		// registry is initialized lazily or via SetCallbackRegistry
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
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Lazily initialize if nil? Or require SetCallbackRegistry?
	// For now, return potentially nil registry.
	return r.registry
}

// RegisterCallback delegates to the registry.
func (r *RunnerImpl) RegisterCallback(hook HookPoint, name string, cb CallbackFunc) error {
	r.mu.RLock()
	registry := r.registry
	r.mu.RUnlock()
	if registry == nil {
		return errors.New("callback registry is not set on the runner")
	}
	// FIX: Pass HookPoint directly, ensure argument order matches registry.Register signature
	// Assuming signature is Register(hook HookPoint, name string, cb CallbackFunc)
	return registry.Register(hook, name, cb)
}

// UnregisterCallback delegates to the registry.
func (r *RunnerImpl) UnregisterCallback(hook HookPoint, name string) {
	r.mu.RLock()
	registry := r.registry
	r.mu.RUnlock()
	if registry != nil {
		// FIX: Pass HookPoint directly, ensure argument order matches registry.Unregister signature
		// Assuming signature is Unregister(hook HookPoint, name string)
		registry.Unregister(hook, name)
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

// Emit adds an event to the processing queue.
// It blocks if the queue is full, up to a timeout.
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

	// Timeout for trying to queue the event
	timeout := 1 * time.Second // Configurable?
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case r.queue <- event:
		// TEMP LOG
		log.Printf("DEBUG: Emit successfully queued event %s", event.GetID())
		return nil
	case <-ctx.Done():
		// TEMP LOG
		log.Printf("DEBUG: Emit timed out for event %s", event.GetID())
		return fmt.Errorf("failed to emit event: queue full or blocked")
	case <-r.stopChan: // Check if runner stopped while waiting
		// TEMP LOG
		log.Printf("DEBUG: Emit failed for event %s: runner stopped while waiting to queue", event.GetID())
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

	log.Println("Runner started.")
	// Launch the main processing loop
	go r.loop(ctx) // Pass the main context

	// Start is non-blocking; it returns after launching the loop.
	return nil
}

// Stop signals the runner to stop processing events and waits for it to finish.
func (r *RunnerImpl) Stop() {
	r.mu.Lock()
	log.Println("Runner Stop: Acquired lock")
	if !r.started {
		log.Println("Runner Stop: Already stopped, released lock.")
		r.mu.Unlock()
		return
	}

	log.Println("Runner Stop: Setting started=false and closing stopChan...")
	r.started = false
	// Close stopChan *before* unlocking to ensure loop sees it if it checks started flag
	close(r.stopChan) // Signal the loop to exit
	log.Println("Runner Stop: stopChan closed.")

	// Unlock *before* waiting to avoid deadlocks if loop needs the lock (e.g., for registry/orchestrator)
	r.mu.Unlock()
	log.Println("Runner Stop: Released lock, waiting for loop goroutine (wg.Wait)...")

	// Wait for the loop goroutine to finish processing any in-flight event and exit
	r.wg.Wait()
	log.Println("Runner Stop: Loop goroutine finished (wg.Wait returned).")

	// Stop the orchestrator after the loop has fully finished
	r.mu.RLock()
	orchestrator := r.orchestrator
	r.mu.RUnlock()
	if orchestrator != nil {
		log.Println("Runner Stop: Stopping orchestrator...")
		// Assuming orchestrator has a Stop method - add if needed
		// orchestrator.Stop()
		log.Println("Runner Stop: Orchestrator stopped (or stop not implemented).")
	}

	log.Println("Runner Stop: Completed.")
}

// loop is the main event processing goroutine.
func (r *RunnerImpl) loop(ctx context.Context) {
	defer func() {
		log.Println("Runner loop exiting...")
		r.wg.Done()
		log.Println("Runner loop: wg.Done() called.")
	}()
	log.Println("Runner loop started.")

	for {
		select {
		case <-ctx.Done():
			log.Println("Runner loop: Main context cancelled.")
			return
		case <-r.stopChan:
			log.Println("Runner loop: Stop signal received.")
			return
		case event, ok := <-r.queue:
			if !ok {
				log.Println("Runner loop: Queue channel closed.")
				return
			}
			log.Printf("Runner loop: Processing event %s...", event.GetID())

			eventCtx := ctx // Inherit main context for this event

			// --- Start Event Processing ---
			sessionID, _ := event.GetMetadataValue(SessionIDKey) // Correctly ignore bool if not needed here
			log.Printf("Runner loop: Event %s belongs to session %s", event.GetID(), sessionID)

			var currentState State = NewState() // Fresh state per event

			// --- Before Event Handling Callback ---
			if r.registry != nil {
				log.Printf("Runner: Invoking %s callbacks", HookBeforeEventHandling)
				callbackArgs := CallbackArgs{
					Hook:  HookBeforeEventHandling,
					Event: event,
					State: currentState,
				}
				returnedState, err := r.registry.Invoke(eventCtx, callbackArgs)
				if err != nil {
					log.Printf("Runner: Error invoking %s callbacks: %v", HookBeforeEventHandling, err)
				}
				if returnedState != nil {
					currentState = returnedState
				}
			}

			// --- Dispatch to Orchestrator ---
			r.mu.RLock()
			orchestrator := r.orchestrator
			r.mu.RUnlock()

			var agentResult AgentResult // Type definition needs verification
			var agentErr error

			if orchestrator != nil {
				// Corrected Dispatch call: Pass only ctx and event
				agentResult, agentErr = orchestrator.Dispatch(eventCtx, event) // Removed currentState
				if agentErr != nil {
					log.Printf("Runner loop: Error dispatching event %s: %v", event.GetID(), agentErr)
					// --- Agent Error Callback ---
					if r.registry != nil {
						log.Printf("Runner: Invoking %s callbacks", HookAgentError)
						errorArgs := CallbackArgs{
							Hook:  HookAgentError,
							Event: event,
							State: currentState,
							Error: agentErr,
							// AgentID is likely unknown here if Dispatch failed
							// AgentResult might be nil or contain partial info
							AgentResult: agentResult,
						}
						returnedState, err := r.registry.Invoke(eventCtx, errorArgs)
						if err != nil {
							log.Printf("Runner: Error invoking %s callbacks: %v", HookAgentError, err)
						}
						if returnedState != nil {
							currentState = returnedState
						}
					}
				} else {
					// Dispatch successful (agentErr == nil)
					// NOTE: We cannot access agentResult.OutputState or agentResult.AgentID
					// because the type definition seems to lack them.
					// How is the resulting state or the agent ID obtained?
					// This logic might need revision based on the actual AgentResult structure
					// and how the orchestrator communicates results.
					// For now, we assume currentState might be modified *within* the orchestrator/agent
					// or the AgentResult itself *is* the new state (needs type assertion).

					// --- After Agent Run Callback ---
					if r.registry != nil {
						// We don't know the AgentID from agentResult based on previous errors.
						// Where does the AgentID come from after successful dispatch?
						// Placeholder: Maybe the event contains it? Or Dispatch needs to return it?
						invokedAgentID := "unknown_agent" // Placeholder! Needs correct source.
						log.Printf("Runner: Invoking %s callbacks for agent %s", HookAfterAgentRun, invokedAgentID)

						callbackArgs := CallbackArgs{
							Hook:        HookAfterAgentRun,
							Event:       event,
							AgentID:     invokedAgentID, // Placeholder! Needs correct source.
							AgentResult: agentResult,    // Pass the result we got
							State:       currentState,   // Pass state *after* dispatch attempt
						}
						returnedState, err := r.registry.Invoke(eventCtx, callbackArgs)
						if err != nil {
							log.Printf("Runner: Error invoking %s callbacks: %v", HookAfterAgentRun, err)
						}
						if returnedState != nil {
							currentState = returnedState
						}
					}
				}
			} else {
				log.Printf("Runner loop: Orchestrator is nil, cannot dispatch event %s", event.GetID())
				agentErr = errors.New("orchestrator not configured")
			}

			// --- Process Agent Result ---
			// Pass the agentResult and agentErr obtained from dispatch
			r.processAgentResult(eventCtx, event, agentResult, agentErr)

			// --- After Event Handling Callback ---
			if r.registry != nil {
				log.Printf("Runner: Invoking %s callbacks", HookAfterEventHandling)
				callbackArgs := CallbackArgs{
					Hook:        HookAfterEventHandling,
					Event:       event,
					State:       currentState, // Final state for this event
					AgentResult: agentResult,  // Pass result obtained
					Error:       agentErr,     // Pass error obtained
				}
				returnedState, err := r.registry.Invoke(eventCtx, callbackArgs)
				if err != nil {
					log.Printf("Runner: Error invoking %s callbacks: %v", HookAfterEventHandling, err)
				}
				if returnedState != nil {
					currentState = returnedState
				}
			}

			log.Printf("Runner loop finished processing event %s", event.GetID())
		} // end select
	} // end for
} // end loop

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

// processAgentResult handles the outcome of an agent execution, potentially emitting new events.
// FIX: Added agentErr parameter
func (r *RunnerImpl) processAgentResult(ctx context.Context, originalEvent Event, result AgentResult, agentErr error) {
	// TODO: Add tracing spans if needed

	// Use agentErr passed from the loop to determine success/failure
	if agentErr != nil {
		// Correct usage of GetMetadataValue
		sessionID, _ := originalEvent.GetMetadataValue(SessionIDKey) // Assign both return values
		log.Printf("Agent execution failed for event %s (session: %s): %v", originalEvent.GetID(), sessionID, agentErr)

		// Optionally emit a failure event
		failurePayload := EventData{
			"original_event_id": originalEvent.GetID(),
			"error":             agentErr.Error(), // Use agentErr
		}
		// Preserve session ID, potentially add AgentID if available in result even on error
		failureMeta := map[string]string{
			SessionIDKey: sessionID, // Use captured sessionID
			"status":     "failure",
		}
		// if result.AgentID != "" { // Cannot access result.AgentID
		// 	failureMeta["failed_agent_id"] = result.AgentID
		// }

		failureEvent := NewEvent("", failurePayload, failureMeta)
		// Set source if known
		// if result.AgentID != "" { // Cannot access result.AgentID
		// 	failureEvent.SetSourceAgentID(result.AgentID)
		// } else {
		// Fallback if agent ID isn't known (e.g., orchestrator error before agent selection)
		failureEvent.SetSourceAgentID("orchestrator_or_agent") // Generic source
		// }

		// Emit the failure event
		if err := r.Emit(failureEvent); err != nil {
			log.Printf("Error emitting failure event for original event %s: %v", originalEvent.GetID(), err)
		}
	} else {
		// This block executes only if agentErr is nil
		// Correct usage of GetMetadataValue
		sessionID, _ := originalEvent.GetMetadataValue(SessionIDKey)                                                                        // Assign both return values
		agentIDForResult := "unknown_agent"                                                                                                 // Placeholder! Still need source for AgentID
		log.Printf("Agent execution successful for event %s (session: %s) by agent %s", originalEvent.GetID(), sessionID, agentIDForResult) // Using placeholder

		// Optionally emit a success event or an event carrying the result state
		successPayload := EventData{
			"original_event_id": originalEvent.GetID(),
			// Include output state data if needed and safe
			// "output_data": result.OutputState.GetData(), // Cannot access result.OutputState
		}
		successMeta := map[string]string{
			SessionIDKey: sessionID, // Use captured sessionID
			"status":     "success",
			"agent_id":   agentIDForResult, // Using placeholder
		}
		// Copy metadata from output state if necessary
		// if result.OutputState != nil { // Cannot access result.OutputState
		// 	if stateWithMeta, ok := result.OutputState.(interface{ GetMetadata() map[string]string }); ok {
		// 		metaMap := stateWithMeta.GetMetadata()
		// 		if metaMap != nil {
		// 			for k, v := range metaMap {
		// 				if _, exists := successMeta[k]; !exists { // Avoid overwriting core meta like sessionID, status
		// 					successMeta[k] = v
		// 				}
		// 			}
		// 		}
		// 	} else {
		// 		log.Printf("Warning: Output state type %T does not have GetMetadata method", result.OutputState)
		// 	}
		// }

		successEvent := NewEvent("", successPayload, successMeta)
		successEvent.SetSourceAgentID(agentIDForResult) // Using placeholder

		// Emit the success event
		if err := r.Emit(successEvent); err != nil {
			log.Printf("Error emitting success event for original event %s: %v", originalEvent.GetID(), err)
		}
	}
}

// internal/core/runner.go (or a constants file)
const (
	callbackStateKeyAgentResult = "__agentResult"
	callbackStateKeyAgentError  = "__agentError"
)
