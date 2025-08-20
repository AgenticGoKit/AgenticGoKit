// Package core provides the public Runner interface and related types for AgentFlow.
package core

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"
)

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
	queue             chan Event
	orchestrator      Orchestrator
	registry          *CallbackRegistry
	traceLogger       TraceLogger
	tracer            trace.Tracer
	errorRouterConfig *ErrorRouterConfig

	stopOnce sync.Once
	stopChan chan struct{}
	wg       sync.WaitGroup
	mu       sync.RWMutex
	started  bool
}

// NewRunner creates a new RunnerImpl.
func NewRunner(queueSize int) *RunnerImpl {
	if queueSize <= 0 {
		queueSize = 100
	}
	return &RunnerImpl{
		queue:    make(chan Event, queueSize),
		stopChan: make(chan struct{}),
		registry: NewCallbackRegistry(),
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
	if r.started {
		Logger().Warn().Msg("Attempted to set trace logger while runner is running.")
		return
	}
	r.traceLogger = logger
}

// SetErrorRouterConfig assigns the error router configuration to the runner.
func (r *RunnerImpl) SetErrorRouterConfig(config *ErrorRouterConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.started {
		Logger().Warn().Msg("Attempted to set error router config while runner is running.")
		return
	}
	r.errorRouterConfig = config
}

// getErrorRouterConfig returns the error router configuration or default if not set.
func (r *RunnerImpl) getErrorRouterConfig() *ErrorRouterConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.errorRouterConfig != nil {
		return r.errorRouterConfig
	}
	return DefaultErrorRouterConfig()
}

// GetTraceLogger returns the runner's trace logger.
func (r *RunnerImpl) GetTraceLogger() TraceLogger {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.traceLogger
}

// GetCallbackRegistry returns the runner's callback registry.
func (r *RunnerImpl) GetCallbackRegistry() *CallbackRegistry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.registry
}

// RegisterCallback delegates to the registry.
func (r *RunnerImpl) RegisterCallback(hook HookPoint, name string, cb CallbackFunc) error {
	r.mu.RLock()
	registry := r.registry
	r.mu.RUnlock()
	return registry.Register(hook, name, cb)
}

// UnregisterCallback delegates to the registry.
func (r *RunnerImpl) UnregisterCallback(hook HookPoint, name string) {
	r.mu.RLock()
	registry := r.registry
	r.mu.RUnlock()
	registry.Unregister(hook, name)
}

// RegisterAgent registers an AgentHandler with the underlying Orchestrator.
func (r *RunnerImpl) RegisterAgent(name string, handler AgentHandler) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.orchestrator == nil {
		return errors.New("orchestrator not set in runner")
	}
	return r.orchestrator.RegisterAgent(name, handler)
}

// Emit adds an event to the processing queue.
func (r *RunnerImpl) Emit(event Event) error {
	r.mu.RLock()
	if !r.started {
		r.mu.RUnlock()
		Logger().Debug().Str("event_id", event.GetID()).Msg("Emit failed: runner not running")
		return errors.New("runner is not running")
	}
	r.mu.RUnlock()

	Logger().Debug().Str("event_id", event.GetID()).Msg("Emit attempting to queue event")

	timeout := 1 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case r.queue <- event:
		Logger().Debug().Str("event_id", event.GetID()).Msg("Emit successfully queued event")
		return nil
	case <-ctx.Done():
		Logger().Debug().Str("event_id", event.GetID()).Msg("Emit timed out")
		return fmt.Errorf("failed to emit event: queue full or blocked")
	case <-r.stopChan:
		Logger().Debug().Str("event_id", event.GetID()).Msg("Emit failed: runner stopped while waiting to queue")
		return errors.New("runner stopped while emitting")
	}
}

// Start begins the runner's event processing loop in a separate goroutine.
func (r *RunnerImpl) Start(ctx context.Context) error {
	r.mu.Lock()
	if r.started {
		r.mu.Unlock()
		return errors.New("runner already started")
	}
	if r.orchestrator == nil {
		r.mu.Unlock()
		return errors.New("orchestrator must be set before starting runner")
	}
	r.started = true
	r.stopChan = make(chan struct{})
	r.wg.Add(1)
	r.mu.Unlock()

	Logger().Info().Msg("Runner started.")
	go r.loop(ctx)
	return nil
}

// Stop signals the runner to stop processing events and waits for it to finish.
func (r *RunnerImpl) Stop() {
	r.mu.Lock()
	Logger().Debug().Msg("Runner Stop: Acquired lock")
	if !r.started {
		Logger().Debug().Msg("Runner Stop: Already stopped, released lock.")
		r.mu.Unlock()
		return
	}

	Logger().Debug().Msg("Runner Stop: Setting started=false and closing stopChan...")
	r.started = false
	close(r.stopChan)
	Logger().Debug().Msg("Runner Stop: stopChan closed.")

	r.mu.Unlock()
	Logger().Debug().Msg("Runner Stop: Released lock, waiting for loop goroutine (wg.Wait)...")

	r.wg.Wait()
	Logger().Debug().Msg("Runner Stop: Loop goroutine finished (wg.Wait returned).")

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
	defer r.wg.Done()
	for {
		select {
		case <-ctx.Done():
			Logger().Debug().Msg("Runner loop: Context cancelled. Exiting.")
			return
		case <-r.stopChan:
			Logger().Debug().Msg("Runner loop: Stop signal received. Exiting.")
			return
		case event := <-r.queue:
			eventCtx, eventCancel := context.WithCancel(ctx)
			defer eventCancel()

			sessionID, _ := event.GetMetadataValue(SessionIDKey)
			if sessionID == "" {
				sessionID = event.GetID()
				Logger().Warn().Str("event_id", event.GetID()).Msg("Runner loop: Warning - event missing session ID, using event ID as fallback.")
				event.SetMetadata(SessionIDKey, sessionID)
			}
			Logger().Debug().Str("event_id", event.GetID()).Str("session_id", sessionID).Msg("Runner loop: Processing event")

			var currentState State = NewState()

			if r.registry != nil {
				Logger().Debug().Msg("Runner: Invoking BeforeEventHandling callbacks")
				callbackArgs := CallbackArgs{
					Hook:    HookBeforeEventHandling,
					Event:   event,
					State:   currentState,
					AgentID: "",
				}
				newState, err := r.registry.Invoke(eventCtx, callbackArgs)
				if err != nil {
					Logger().Error().Str("event_id", event.GetID()).Err(err).Msg("Runner loop: Error during BeforeEventHandling callbacks. Skipping event.")
					continue
				}
				if newState != nil {
					currentState = newState
				}
				Logger().Debug().Msg("CallbackRegistry.Invoke: Finished invoking callbacks for hook BeforeEventHandling.")
			}

			var agentResult AgentResult
			var agentErr error
			var invokedAgentID string

			r.mu.RLock()
			orchestrator := r.orchestrator
			r.mu.RUnlock()

			if orchestrator != nil {
				targetAgentID := "unknown"
				if routeKey, ok := event.GetMetadataValue(RouteMetadataKey); ok {
					targetAgentID = routeKey
				} else if event.GetTargetAgentID() != "" {
					targetAgentID = event.GetTargetAgentID()
				}
				invokedAgentID = targetAgentID

				if r.registry != nil {
					Logger().Debug().Str("agent_id", invokedAgentID).Msgf("Runner: Invoking %s callbacks", HookBeforeAgentRun)
					callbackArgs := CallbackArgs{
						Hook:    HookBeforeAgentRun,
						Event:   event,
						State:   currentState,
						AgentID: invokedAgentID,
					}
					newState, err := r.registry.Invoke(eventCtx, callbackArgs)
					if err != nil {
						Logger().Error().Str("event_id", event.GetID()).Str("agent_id", invokedAgentID).Err(err).Msg("Runner loop: Error during BeforeAgentRun callbacks")
						agentErr = fmt.Errorf("BeforeAgentRun callback failed: %w", err)
					} else {
						if newState != nil {
							currentState = newState
						}
						Logger().Debug().Msg("CallbackRegistry.Invoke: Finished invoking callbacks for hook BeforeAgentRun.")
					}
				}

				if agentErr == nil {
					Logger().Debug().Str("event_id", event.GetID()).Msg("Runner loop: Dispatching event to orchestrator")
					agentResult, agentErr = orchestrator.Dispatch(eventCtx, event)
				}

				if agentErr != nil {
					Logger().Error().Str("event_id", event.GetID()).Err(agentErr).Msg("Runner loop: Error during agent execution/dispatch")
					if r.registry != nil {
						Logger().Debug().Str("agent_id", invokedAgentID).Msgf("Runner: Invoking %s callbacks", HookAgentError)
						callbackArgs := CallbackArgs{
							Hook:    HookAgentError,
							Event:   event,
							AgentID: invokedAgentID,
							Error:   agentErr,
							State:   currentState,
						}
						newState, cbErr := r.registry.Invoke(eventCtx, callbackArgs)
						if cbErr != nil {
							Logger().Error().Str("event_id", event.GetID()).Err(cbErr).Msg("Runner loop: Error during AgentError callback")
						}
						if newState != nil {
							currentState = newState
						}
						Logger().Debug().Msg("CallbackRegistry.Invoke: Finished invoking callbacks for hook AgentError.")
					}
				}
			} else {
				Logger().Error().Str("event_id", event.GetID()).Msg("Runner loop: Orchestrator is nil, cannot dispatch event")
				agentErr = errors.New("orchestrator not configured")
				invokedAgentID = "orchestrator"
			}

			r.processAgentResult(eventCtx, event, agentResult, agentErr, invokedAgentID)

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
				_, cbErr := r.registry.Invoke(eventCtx, callbackArgs)
				if cbErr != nil {
					Logger().Error().Str("event_id", event.GetID()).Err(cbErr).Msg("Runner loop: Error during AfterEventHandling callbacks")
				}
				Logger().Debug().Msg("CallbackRegistry.Invoke: Finished invoking callbacks for hook AfterEventHandling.")
			}

			Logger().Debug().Str("event_id", event.GetID()).Msg("Runner loop finished processing event")
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

		// Use enhanced error routing system
		errorRouterConfig := r.getErrorRouterConfig()
		failureEvent := CreateEnhancedErrorEvent(originalEvent, agentID, agentErr, errorRouterConfig)

		if err := r.Emit(failureEvent); err != nil {
			Logger().Error().
				Str("event_id", originalEvent.GetID()).
				Err(err).
				Msg("Error emitting enhanced failure event")
		}
	} else {
		Logger().Debug().
			Str("event_id", originalEvent.GetID()).
			Str("session_id", sessionID).
			Str("agent_id", agentID).
			Msg("Agent execution successful")

		if result.OutputState != nil {
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
					route,
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
	logger := r.traceLogger
	r.mu.RUnlock()

	if logger == nil {
		return nil, errors.New("trace logger is not set")
	}
	return logger.GetTrace(sessionID)
}

const (
	callbackStateKeyAgentResult = "__agentResult"
	callbackStateKeyAgentError  = "__agentError"
)

// =============================================================================
// RUNNER CONFIGURATION
// =============================================================================

// RunnerConfig holds configuration for creating runners
type RunnerConfig struct {
	QueueSize int                     `json:"queue_size"`
	Agents    map[string]AgentHandler `json:"agents"`
	Memory    Memory                  `json:"memory"`
	Callbacks *CallbackRegistry       `json:"callbacks"`
	SessionID string                  `json:"session_id"`
}

// DefaultRunnerConfig returns sensible defaults for runner configuration
func DefaultRunnerConfig() RunnerConfig {
	return RunnerConfig{
		QueueSize: 100,
		Agents:    make(map[string]AgentHandler),
		Memory:    nil,
		Callbacks: NewCallbackRegistry(),
	}
}

// NewRunnerWithConfig creates a new runner with the specified configuration
func NewRunnerWithConfig(config RunnerConfig) Runner {
	if runnerFactory != nil {
		return runnerFactory(config)
	}
	// Return a basic implementation
	return &basicRunner{config: config}
}

// RegisterRunnerFactory registers the runner factory function
func RegisterRunnerFactory(factory func(RunnerConfig) Runner) {
	runnerFactory = factory
}

var runnerFactory func(RunnerConfig) Runner

// basicRunner provides a minimal implementation
type basicRunner struct {
	config RunnerConfig
}

func (r *basicRunner) Emit(event Event) error {
	Logger().Warn().Msg("Using basic runner - import internal/factory for full functionality")
	return nil
}

func (r *basicRunner) RegisterAgent(name string, handler AgentHandler) error {
	Logger().Warn().Msg("Using basic runner - import internal/factory for full functionality")
	return nil
}

func (r *basicRunner) RegisterCallback(hook HookPoint, name string, cb CallbackFunc) error {
	Logger().Warn().Msg("Using basic runner - import internal/factory for full functionality")
	return nil
}

func (r *basicRunner) UnregisterCallback(hook HookPoint, name string) {
	Logger().Warn().Msg("Using basic runner - import internal/factory for full functionality")
}

func (r *basicRunner) Start(ctx context.Context) error {
	Logger().Warn().Msg("Using basic runner - import internal/factory for full functionality")
	return nil
}

func (r *basicRunner) Stop() {
	Logger().Warn().Msg("Using basic runner - import internal/factory for full functionality")
}

func (r *basicRunner) GetCallbackRegistry() *CallbackRegistry {
	return r.config.Callbacks
}

func (r *basicRunner) GetTraceLogger() TraceLogger {
	Logger().Warn().Msg("Using basic runner - import internal/factory for full functionality")
	return &basicTraceLogger{}
}

func (r *basicRunner) DumpTrace(sessionID string) ([]TraceEntry, error) {
	Logger().Warn().Msg("Using basic runner - import internal/factory for full functionality")
	return []TraceEntry{}, nil
}

// basicTraceLogger provides a minimal implementation
type basicTraceLogger struct{}

func (t *basicTraceLogger) Log(entry TraceEntry) error {
	// Basic implementation - do nothing
	return nil
}

func (t *basicTraceLogger) GetTrace(sessionID string) ([]TraceEntry, error) {
	return []TraceEntry{}, nil
}

// NewRunnerFromConfig creates a new runner by loading configuration from a file path
func NewRunnerFromConfig(configPath string) (Runner, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %w", configPath, err)
	}

	// Map the loaded config values to runnerConfig
	runnerConfig := RunnerConfig{
		QueueSize: 1000,                          // Default queue size - could be made configurable
		Agents:    make(map[string]AgentHandler), // Will be populated when agents are registered
		Memory:    nil,                           // Can be set later if needed from config.AgentMemory
		Callbacks: NewCallbackRegistry(),
		SessionID: "", // Will be set when starting
	}

	// Apply runtime configuration if available
	if config.Runtime.MaxConcurrentAgents > 0 {
		// Could potentially influence queue size
		runnerConfig.QueueSize = config.Runtime.MaxConcurrentAgents * 100 // Heuristic
	}

	return NewRunnerWithConfig(runnerConfig), nil
}
