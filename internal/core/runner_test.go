package agentflow

import (
	"context"
	"errors" // Import errors package
	"fmt"
	"log" // Import rand for concurrency test
	"strings"
	"sync"        // Import sync
	"sync/atomic" // Import atomic
	"testing"
	"time"
)

// --- Mock EventHandler ---
type mockEventHandler struct {
	handleFunc func(Event) error
}

func (m *mockEventHandler) Handle(e Event) error {
	if m.handleFunc != nil {
		return m.handleFunc(e)
	}
	return nil
}

// --- Mock AgentHandler ---
type mockAgentHandler struct {
	runFunc func(context.Context, Event, State) (AgentResult, error)
}

func (m *mockAgentHandler) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
	if m.runFunc != nil {
		return m.runFunc(ctx, event, state)
	}
	// Return a default result for the mock, ensure state pointer is handled
	var outputState State
	if state != nil {
		outputState = state.Clone() // Clone input state if non-nil
	} else {
		outputState = NewState() // Provide a default empty state if input state was nil
	}
	return AgentResult{OutputState: outputState}, nil
}

// --- Mock Orchestrator ---
type mockOrchestrator struct {
	dispatchFunc   func(context.Context, Event) (AgentResult, error)
	registerFunc   func(string, AgentHandler) error
	stopFunc       func()
	registry       *CallbackRegistry // Registry used by this mock (can be nil if not needed by mock)
	mu             sync.Mutex        // Protect access to shared fields
	dispatchCalled bool
	receivedEvent  Event // Store the last received event
	dispatchErr    error // Store error encountered during dispatch check
	returnErr      error // Error to return from Dispatch call
}

// Dispatch simulates orchestrator dispatch.
// REMOVED the simulation of Before/AfterEventHandling hooks, as the runner handles those.
func (m *mockOrchestrator) Dispatch(ctx context.Context, e Event) (AgentResult, error) {
	m.mu.Lock()
	m.dispatchCalled = true
	m.receivedEvent = e
	m.mu.Unlock()

	// Simulate the actual agent execution (using dispatchFunc or default)
	var dispatchResult AgentResult
	var dispatchErr error
	if m.dispatchFunc != nil {
		// If the mock dispatchFunc needs to simulate agent-specific hooks (like Before/AfterAgentRun),
		// it could potentially use m.registry here, but it should NOT invoke
		// Before/AfterEventHandling hooks.
		dispatchResult, dispatchErr = m.dispatchFunc(ctx, e)
	} else {
		dispatchResult = AgentResult{} // Default empty result
		dispatchErr = m.returnErr      // Use pre-configured error if any
	}

	// Store error for CheckDispatchCalled if needed
	if dispatchErr != nil {
		m.mu.Lock()
		m.dispatchErr = dispatchErr
		m.mu.Unlock()
	}

	// NOTE: No invocation of Before/AfterEventHandling hooks here.
	// The runner loop is responsible for invoking those hooks around this Dispatch call.

	return dispatchResult, dispatchErr
}

func (m *mockOrchestrator) RegisterAgent(name string, handler AgentHandler) error {
	if m.registerFunc != nil {
		return m.registerFunc(name, handler)
	}
	return nil
}

// GetCallbackRegistry is kept for potential use within a custom dispatchFunc,
// but the main Dispatch method no longer uses it directly.
func (m *mockOrchestrator) GetCallbackRegistry() *CallbackRegistry {
	if m.registry == nil {
		m.registry = NewCallbackRegistry()
	}
	return m.registry
}

func (m *mockOrchestrator) Stop() {
	if m.stopFunc != nil {
		m.stopFunc()
	}
}

// --- Mock TraceLogger ---
type mockTraceLogger struct {
	logFunc      func(TraceEntry) error
	getTraceFunc func(string) ([]TraceEntry, error)
	mu           sync.RWMutex
	entries      []TraceEntry
}

func (l *mockTraceLogger) Log(entry TraceEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, entry)
	if l.logFunc != nil {
		return l.logFunc(entry)
	}
	return nil
}

func (l *mockTraceLogger) GetTrace(sessionID string) ([]TraceEntry, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.getTraceFunc != nil {
		return l.getTraceFunc(sessionID)
	}
	// Simple mock: return all entries regardless of sessionID
	// Or filter if needed:
	var filtered []TraceEntry
	for _, e := range l.entries {
		if e.SessionID == sessionID {
			filtered = append(filtered, e)
		}
	}
	return filtered, nil

}

// --- Test Runner Start and Stop ---
func TestRunner_StartStop(t *testing.T) {
	runner := NewRunner(10)
	mockOrch := &mockOrchestrator{} // Now implements Stop

	runner.SetOrchestrator(mockOrch)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := runner.Start(ctx)
	if err != nil {
		t.Fatalf("Runner failed to start: %v", err)
	}

	err = runner.Start(ctx)
	if err == nil {
		t.Fatal("Runner started again without error, should have failed")
	}

	runner.Stop()
	runner.Stop() // Idempotent check

	testEvent := NewEvent("test-agent", EventData{"data": "value"}, map[string]string{"session_id": "s1"})
	err = runner.Emit(testEvent)
	if err == nil {
		t.Fatal("Emit succeeded after stop, should have failed")
	}
}

// --- Test Runner Emit and Dispatch ---
func TestRunner_EmitDispatch(t *testing.T) {
	dispatchChan := make(chan Event, 1) // Channel to signal dispatch occurred
	mockOrch := &mockOrchestrator{
		dispatchFunc: func(ctx context.Context, e Event) (AgentResult, error) {
			log.Printf("MockOrchestrator: Dispatch called for event %s", e.GetID())
			dispatchChan <- e // Signal that dispatch happened
			return AgentResult{}, nil
		},
	}
	runner := NewRunner(10)
	runner.SetOrchestrator(mockOrch)
	runner.SetTraceLogger(NewNoOpTraceLogger()) // Use NoOp logger

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var runnerWg sync.WaitGroup
	runnerWg.Add(1)
	go func() {
		defer runnerWg.Done()
		if err := runner.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			// Log error only if it's not context cancellation
			t.Logf("Runner Start returned error: %v", err)
		}
	}()
	time.Sleep(50 * time.Millisecond) // Allow runner goroutine to start

	testEvent := NewEvent("test-event-dispatch", nil, nil)
	err := runner.Emit(testEvent)
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	// Wait for dispatch signal or timeout
	select {
	case receivedEvent := <-dispatchChan:
		log.Printf("TestRunner_EmitDispatch: Received dispatch signal for event %s", receivedEvent.GetID())
		if receivedEvent.GetID() != testEvent.GetID() {
			t.Errorf("Mock orchestrator received wrong event ID: got %s, want %s", receivedEvent.GetID(), testEvent.GetID())
		}
	case <-time.After(2 * time.Second): // Increased timeout
		t.Fatal("Timeout waiting for orchestrator dispatch")
	}

	cancel()        // Stop the runner
	runnerWg.Wait() // Wait for runner goroutine to finish

	// --- Assertions after runner is stopped ---
	// CheckDispatchCalled is less reliable than the channel, but keep for completeness
	called, _, dispatchErr := mockOrch.CheckDispatchCalled()
	if !called {
		t.Errorf("Expected orchestrator.Dispatch to be called (CheckDispatchCalled), but it wasn't")
	}
	if dispatchErr != nil {
		t.Errorf("Mock orchestrator encountered an error during dispatch check: %v", dispatchErr)
	}
}

// --- Test Runner Emit Queue Full ---
func TestRunner_EmitQueueFull(t *testing.T) {
	// FIX: Use queue size 1
	queueSize := 1
	runner := NewRunner(queueSize)
	processed := make(chan struct{})     // Channel to signal processing completion
	blockDispatch := make(chan struct{}) // Channel to control blocking in dispatch
	dispatchStarted := make(chan string) // Channel to signal dispatch has started

	mockOrch := &mockOrchestrator{
		dispatchFunc: func(ctx context.Context, e Event) (AgentResult, error) {
			log.Printf("MockOrchestrator: Dispatch called for event %s, signalling start...", e.GetID())
			// Signal that dispatch has started *before* blocking
			// Use non-blocking send in case test already moved on/timed out
			select {
			case dispatchStarted <- e.GetID():
			default:
				log.Printf("MockOrchestrator: Warning - dispatchStarted channel blocked or closed for event %s", e.GetID())
			}

			log.Printf("MockOrchestrator: Blocking dispatch for event %s...", e.GetID())
			// Block until told to unblock
			<-blockDispatch
			log.Printf("MockOrchestrator: Unblocked for event %s", e.GetID())

			// Signal processing done *after* unblocking
			// Use a non-blocking send in case the test already timed out
			select {
			case processed <- struct{}{}:
			default:
				log.Printf("MockOrchestrator: Warning - processed channel blocked or closed for event %s", e.GetID())
			}
			return AgentResult{}, nil
		},
	}
	runner.SetOrchestrator(mockOrch)
	runner.SetTraceLogger(NewNoOpTraceLogger())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runner.Start(ctx)
	time.Sleep(50 * time.Millisecond) // Allow runner goroutine to start

	// --- Event 1 ---
	// Emit first event - should enter the queue (size 1)
	event1 := NewEvent("event1", nil, nil)
	err1 := runner.Emit(event1)
	if err1 != nil {
		t.Fatalf("Emit(1) failed unexpectedly: %v", err1)
	}
	log.Printf("Test: Emit(1) succeeded.")

	// Wait for Dispatch to start processing event 1
	select {
	case startedID := <-dispatchStarted:
		if startedID != event1.GetID() {
			// Unblock dispatch before failing
			close(blockDispatch)
			t.Fatalf("Dispatch started for unexpected event ID: got %s, want %s", startedID, event1.GetID())
		}
		log.Printf("Test: Confirmed Dispatch started for event 1.")
	case <-time.After(1 * time.Second):
		// Unblock dispatch before failing
		close(blockDispatch)
		t.Fatal("Timeout waiting for Dispatch to start for event 1")
	}
	// At this point, event 1 is *out* of the queue and blocked in Dispatch. Queue is empty.

	// --- Event 2 ---
	// Emit second event - should fill the queue (size 1) again
	event2 := NewEvent("event2", nil, nil)
	err2 := runner.Emit(event2)
	if err2 != nil {
		// Unblock dispatch before failing
		close(blockDispatch)
		t.Fatalf("Emit(2) failed unexpectedly: %v", err2)
	}
	log.Printf("Test: Emit(2) succeeded (queue should now be full).")
	// At this point, event 1 is blocked in Dispatch, event 2 is in the queue.

	// --- Event 3 ---
	// Emit third event - THIS should block and timeout because queue is full
	event3 := NewEvent("event3", nil, nil)
	err3 := runner.Emit(event3)
	if err3 == nil {
		// Unblock dispatch before failing
		close(blockDispatch)
		t.Fatal("Emit(3) succeeded unexpectedly, expected timeout error")
	}

	// Check the error for Emit(3)
	expectedErrSubstr := "queue full or blocked"
	if !strings.Contains(err3.Error(), expectedErrSubstr) {
		// Unblock dispatch before failing
		close(blockDispatch)
		t.Errorf("Emit(3) error mismatch: got '%v', want error containing '%s'", err3, expectedErrSubstr)
	}
	log.Printf("Emit(3) correctly failed with timeout error: %v", err3)

	// --- Cleanup ---
	// Unblock the first event processing
	log.Printf("Test: Unblocking dispatch...")
	// Ensure close is only called once
	select {
	case <-blockDispatch: // Already closed
	default:
		close(blockDispatch)
	}

	// Wait for the first event to be fully processed
	select {
	case <-processed:
		log.Println("Test: First event processed signal received.")
	case <-time.After(1 * time.Second):
		log.Println("Test: Timeout waiting for first event processing signal.")
	}

	// Stop the runner (cancel context already called by defer)
	log.Printf("Test: Stopping runner...")
	runner.Stop() // Ensure Stop is called for cleanup
	log.Printf("Test: Runner stopped.")
}

// --- Helper waitTimeout ---
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

// --- Test Runner RegisterAgent ---
func TestRunner_RegisterAgent(t *testing.T) {
	registered := false
	agentName := "test-agent"
	var registeredHandler AgentHandler // Keep as interface type

	mockOrch := &mockOrchestrator{ // Now implements Stop
		registerFunc: func(name string, handler AgentHandler) error { // Expects AgentHandler
			if name == agentName && handler != nil {
				registered = true
				registeredHandler = handler // Assign passed handler to interface variable
				return nil
			}
			return fmt.Errorf("mock registration failed for %s", name)
		},
	}
	runner := NewRunner(10)
	runner.SetOrchestrator(mockOrch)

	mockHandler := &mockAgentHandler{} // Implements AgentHandler via Run method

	err := runner.RegisterAgent(agentName, mockHandler)
	if err != nil {
		t.Fatalf("RegisterAgent failed: %v", err)
	}
	if !registered {
		t.Error("Orchestrator's RegisterAgent was not called")
	}
	if registeredHandler != mockHandler {
		t.Errorf("Orchestrator's RegisterAgent was called with the wrong handler. Expected %T, Got %T", mockHandler, registeredHandler)
	}
}

// --- Test Runner Callbacks ---
func TestRunner_Callbacks(t *testing.T) {
	registry := NewCallbackRegistry()
	mockOrchestrator := &mockOrchestrator{}
	runner := NewRunner(1)
	runner.SetOrchestrator(mockOrchestrator)
	runner.SetCallbackRegistry(registry)

	var beforeCalled, afterCalled atomic.Bool // Use atomic flags
	var wg sync.WaitGroup
	wg.Add(2) // Expecting one call for BeforeEventHandling and one for AfterEventHandling

	// Callbacks using CallbackArgs signature
	beforeCallback := func(ctx context.Context, args CallbackArgs) (State, error) {
		// Ensure Done is called only once even if callback runs multiple times unexpectedly
		if !beforeCalled.Swap(true) {
			log.Printf("BeforeEventHandling callback executed for event %s", args.Event.GetID())
			wg.Done()
		} else {
			// Log if called again, helps debugging but shouldn't happen in this test
			log.Printf("WARN: BeforeEventHandling callback executed AGAIN for event %s", args.Event.GetID())
		}
		return args.State, nil
	}
	afterCallback := func(ctx context.Context, args CallbackArgs) (State, error) {
		// Ensure Done is called only once
		if !afterCalled.Swap(true) {
			log.Printf("AfterEventHandling callback executed for event %s", args.Event.GetID())
			wg.Done()
		} else {
			// Log if called again
			log.Printf("WARN: AfterEventHandling callback executed AGAIN for event %s", args.Event.GetID())
		}
		return args.State, nil
	}

	// Register callbacks directly on the registry
	registry.Register(HookBeforeEventHandling, "testBefore", beforeCallback)
	registry.Register(HookAfterEventHandling, "testAfter", afterCallback)

	ctx, cancel := context.WithCancel(context.Background())
	// Defer cancel to ensure cleanup even on test failure before Stop
	defer cancel()

	var runnerWg sync.WaitGroup
	runnerWg.Add(1)
	go func() {
		defer runnerWg.Done()
		// Start runner, log errors other than context cancellation
		if err := runner.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			t.Logf("Runner Start returned error: %v", err)
		}
	}()
	// Short sleep to ensure runner goroutine has likely started
	time.Sleep(50 * time.Millisecond)

	testEvent := NewEvent("agent1", EventData{"data": "test"}, nil)
	err := runner.Emit(testEvent)
	if err != nil {
		// Cancel context before failing to help runner shut down
		cancel()
		runnerWg.Wait() // Wait for runner goroutine before failing test
		t.Fatalf("Emit failed: %v", err)
	}

	// Wait for the two expected callbacks to complete using a channel and select
	waitChan := make(chan struct{})
	go func() {
		wg.Wait() // Wait for wg counter to become 0
		close(waitChan)
	}()

	select {
	case <-waitChan:
		// Callbacks completed successfully
		log.Println("WaitGroup wait completed.")
	case <-time.After(3 * time.Second): // Increased timeout slightly just in case
		// Timeout occurred
		// Cancel context before failing
		cancel()
		runnerWg.Wait() // Wait for runner goroutine before failing test
		t.Fatalf("Timeout waiting for callbacks. BeforeCalled: %v, AfterCalled: %v", beforeCalled.Load(), afterCalled.Load())
	}

	// Now stop the runner *after* confirming callbacks ran (or timed out)
	log.Println("Stopping runner...")
	cancel() // Signal runner loop to stop (might be redundant if timeout occurred, but safe)
	log.Println("Waiting for runner loop goroutine...")
	runnerWg.Wait() // Wait for runner loop to fully exit
	log.Println("Runner loop goroutine finished.")

	// Final assertions
	if !beforeCalled.Load() {
		t.Errorf("BeforeEventHandling callback was not called")
	}
	if !afterCalled.Load() {
		t.Errorf("AfterEventHandling callback was not called")
	}
	log.Println("TestRunner_Callbacks finished.") // Add log to see when test function actually ends
}

// --- Test Runner DumpTrace ---
func TestRunner_DumpTrace(t *testing.T) {
	traceEntry := TraceEntry{EventID: "test-event-id", Hook: HookBeforeEventHandling, SessionID: "s5"}
	mockLogger := &mockTraceLogger{}
	mockLogger.Log(traceEntry) // Pre-populate logger

	runner := NewRunner(10)
	runner.SetTraceLogger(mockLogger)

	entries, err := runner.DumpTrace("s5")
	if err != nil {
		t.Fatalf("DumpTrace failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 trace entry for session 's5', got %d", len(entries))
	}
	if entries[0].EventID != "test-event-id" {
		t.Errorf("Expected event ID 'test-event-id', got '%s'", entries[0].EventID)
	}

	// Test non-existent session
	entries, err = runner.DumpTrace("unknown-session")
	if err != nil {
		t.Fatalf("DumpTrace failed for unknown session: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 trace entries for unknown session, got %d", len(entries))
	}

	// Test with nil logger
	runnerNilLogger := NewRunner(10)
	_, err = runnerNilLogger.DumpTrace("s5")
	if err == nil {
		t.Fatal("DumpTrace succeeded with nil logger, expected error")
	}
	// FIX: Update expected error string
	expectedErr := "trace logger is not set"
	if err.Error() != expectedErr {
		t.Errorf("Expected '%s' error, got: %v", expectedErr, err)
	}
}

// Keep the helper method
func (m *mockOrchestrator) CheckDispatchCalled() (bool, Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.dispatchCalled, m.receivedEvent, m.dispatchErr
}
