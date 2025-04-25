package agentflow

import (
	"context"
	"fmt"
	"sync"
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
		// If state is needed in output, create a copy or handle appropriately
		// For simplicity, just passing it along if non-nil.
		outputState = state
	} else {
		// Provide a default empty state if input state was nil
		// FIX: Use lowercase 'data' field name
		outputState = &SimpleState{data: map[string]interface{}{}}
	}
	return AgentResult{OutputState: &outputState}, nil
}

// --- Mock Orchestrator ---
type mockOrchestrator struct {
	dispatchFunc func(Event) error
	registerFunc func(string, AgentHandler) error
	stopFunc     func() // Add stopFunc field
	registry     *CallbackRegistry
}

func (m *mockOrchestrator) Dispatch(e Event) error {
	if m.dispatchFunc != nil {
		return m.dispatchFunc(e)
	}
	return nil
}

func (m *mockOrchestrator) RegisterAgent(name string, handler AgentHandler) error {
	if m.registerFunc != nil {
		return m.registerFunc(name, handler)
	}
	return nil
}

func (m *mockOrchestrator) GetCallbackRegistry() *CallbackRegistry {
	if m.registry != nil {
		return m.registry
	}
	return NewCallbackRegistry()
}

// FIX: Add Stop method
func (m *mockOrchestrator) Stop() {
	if m.stopFunc != nil {
		m.stopFunc()
	}
	// Default mock behavior: do nothing
}

// --- Mock TraceLogger ---
type mockTraceLogger struct {
	logFunc      func(TraceEntry) error
	getTraceFunc func(string) ([]TraceEntry, error)
}

func (l *mockTraceLogger) Log(entry TraceEntry) error {
	if l.logFunc != nil {
		return l.logFunc(entry)
	}
	return nil
}

func (l *mockTraceLogger) GetTrace(sessionID string) ([]TraceEntry, error) {
	if l.getTraceFunc != nil {
		return l.getTraceFunc(sessionID)
	}
	return []TraceEntry{}, nil
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
	var wg sync.WaitGroup
	wg.Add(1)

	mockOrch := &mockOrchestrator{ // Now implements Stop
		dispatchFunc: func(e Event) error {
			defer wg.Done()
			if e.GetTargetAgentID() != "test-agent" {
				t.Errorf("Expected target agent ID 'test-agent', got '%s'", e.GetTargetAgentID())
			}
			payload := e.GetData()
			if _, ok := payload["payload"]; !ok {
				t.Errorf("Expected 'payload' key in event data, got none")
			} else if payload["payload"] != "hello" {
				t.Errorf("Expected data['payload'] == 'hello', got %v", payload["payload"])
			}
			return nil
		},
	}

	runner := NewRunner(10)
	runner.SetOrchestrator(mockOrch)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := runner.Start(ctx)
	if err != nil {
		t.Fatalf("Runner failed to start: %v", err)
	}
	defer runner.Stop()

	testEvent := NewEvent("test-agent", EventData{"payload": "hello"}, map[string]string{"session_id": "s2"})
	err = runner.Emit(testEvent)
	if err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	if waitTimeout(&wg, 1*time.Second) {
		t.Fatal("Timed out waiting for event to be dispatched")
	}
}

// --- Test Runner Emit Queue Full ---
func TestRunner_EmitQueueFull(t *testing.T) {
	queueSize := 1
	runner := NewRunner(queueSize)

	// FIX: Dispatcher only needs to block longer than Emit's timeout (1s)
	// No need for an external channel here.
	mockOrch := &mockOrchestrator{
		dispatchFunc: func(e Event) error {
			// Block slightly longer than the Emit timeout to ensure it triggers
			time.Sleep(1200 * time.Millisecond)
			return nil
		},
	}
	runner.SetOrchestrator(mockOrch)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure context is cancelled eventually

	err := runner.Start(ctx)
	if err != nil {
		t.Fatalf("Runner failed to start: %v", err)
	}
	// FIX: Defer Stop AFTER all Emit calls to ensure it runs last
	// defer runner.Stop() // Moved lower

	// Emit event 1 - Loop will pick this up and block in the dispatcher
	event1 := NewEvent("agent1", EventData{"num": 1}, map[string]string{"session_id": "s3"})
	err = runner.Emit(event1)
	if err != nil {
		t.Fatalf("Emit 1 failed unexpectedly: %v", err)
	}
	// Give loop a tiny moment to pick up event1 and enter dispatch
	time.Sleep(10 * time.Millisecond)

	// Emit event 2 - Should succeed and fill the queue buffer (size 1)
	event2 := NewEvent("agent2", EventData{"num": 2}, map[string]string{"session_id": "s3"})
	err = runner.Emit(event2)
	if err != nil {
		// If this fails, the timing might still be off, or queue size wrong
		t.Fatalf("Emit 2 failed unexpectedly: %v", err)
	}

	// Emit event 3 - Queue is now full, this Emit should block and time out
	event3 := NewEvent("agent3", EventData{"num": 3}, map[string]string{"session_id": "s3"})
	startTime := time.Now()
	err = runner.Emit(event3)
	duration := time.Since(startTime)

	// Check if Emit 3 failed as expected
	if err == nil {
		t.Fatal("Emit 3 succeeded when queue should be full and blocked, expected timeout error")
	} else {
		expectedErr := "failed to emit event: queue full or blocked"
		if err.Error() != expectedErr {
			t.Errorf("Emit 3 failed with unexpected error. Got '%v', want '%s'", err, expectedErr)
		} else {
			t.Logf("Emit 3 failed with expected error: %v", err)
			// Check if it took roughly the timeout duration
			if duration < 900*time.Millisecond || duration > 1100*time.Millisecond {
				t.Errorf("Emit 3 timeout duration was unexpected: %v (expected ~1s)", duration)
			}
		}
	}

	// FIX: Call Stop explicitly after operations are done, before test exits.
	runner.Stop()
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

	// FIX: Ensure runner.RegisterAgent expects AgentHandler. If the error persists,
	// the RunnerImpl.RegisterAgent signature itself might be wrong in runner.go.
	// Assuming RunnerImpl.RegisterAgent takes AgentHandler:
	err := runner.RegisterAgent(agentName, mockHandler)
	if err != nil {
		// If the error "cannot use mockHandler (*mockAgentHandler) as EventHandler" happens here,
		// it means RunnerImpl.RegisterAgent incorrectly expects EventHandler.
		// You would need to fix RunnerImpl.RegisterAgent in runner.go.
		t.Fatalf("RegisterAgent failed: %v", err)
	}
	if !registered {
		t.Error("Orchestrator's RegisterAgent was not called")
	}
	// FIX: Comparison between interface and concrete type is valid.
	if registeredHandler != mockHandler {
		t.Errorf("Orchestrator's RegisterAgent was called with the wrong handler. Expected %T, Got %T", mockHandler, registeredHandler)
	}
}

// --- Test Runner Callbacks ---
func TestRunner_Callbacks(t *testing.T) {
	var beforeCalled, afterCalled bool
	runner := NewRunner(10)
	mockOrch := &mockOrchestrator{ // Now implements Stop
		dispatchFunc: func(e Event) error { return nil },
	}
	runner.SetOrchestrator(mockOrch)

	// FIX: Update anonymous function signature to accept Event
	err := runner.RegisterCallback(HookBeforeEventHandling, "testBefore", func(ctx context.Context, s State, e Event) (State, error) { // Changed *Event to Event
		beforeCalled = true
		// Add nil check if accessing event 'e'
		if e == nil {
			t.Error("Before callback received nil event") // Keep nil check
		}
		return nil, nil
	})
	if err != nil {
		t.Fatalf("RegisterCallback (before) failed: %v", err)
	}
	// FIX: Update anonymous function signature to accept Event
	err = runner.RegisterCallback(HookAfterEventHandling, "testAfter", func(ctx context.Context, s State, e Event) (State, error) { // Changed *Event to Event
		afterCalled = true
		// Add nil check if accessing event 'e'
		if e == nil {
			t.Error("After callback received nil event") // Keep nil check
		}
		return nil, nil
	})
	if err != nil {
		t.Fatalf("RegisterCallback (after) failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startErr := runner.Start(ctx)
	if startErr != nil {
		t.Fatalf("Runner failed to start: %v", startErr)
	}
	defer runner.Stop()

	testEvent := NewEvent("agent", EventData{"data": "trigger"}, map[string]string{"session_id": "s4"})
	emitErr := runner.Emit(testEvent)
	if emitErr != nil {
		t.Fatalf("Emit failed: %v", emitErr)
	}

	// Allow time for the event to be processed and callbacks to fire
	// Use a WaitGroup or channel for more robust synchronization if needed
	time.Sleep(100 * time.Millisecond) // Increased sleep slightly

	if !beforeCalled {
		t.Error("BeforeEventHandling callback was not called")
	}
	if !afterCalled {
		t.Error("AfterEventHandling callback was not called")
	}
}

// --- Test Runner DumpTrace ---
func TestRunner_DumpTrace(t *testing.T) {
	traceEntry := TraceEntry{EventID: "test-event-id", Hook: HookBeforeEventHandling}
	mockLogger := &mockTraceLogger{
		getTraceFunc: func(sessionID string) ([]TraceEntry, error) {
			if sessionID == "s5" {
				return []TraceEntry{traceEntry}, nil
			}
			return nil, fmt.Errorf("session not found")
		},
	}

	runner := NewRunner(10)
	runner.SetTraceLogger(mockLogger)

	entries, err := runner.DumpTrace("s5")
	if err != nil {
		t.Fatalf("DumpTrace failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 trace entry, got %d", len(entries))
	}
	if entries[0].EventID != "test-event-id" {
		t.Errorf("Expected event ID 'test-event-id', got '%s'", entries[0].EventID)
	}

	_, err = runner.DumpTrace("unknown-session")
	if err == nil {
		t.Fatal("DumpTrace succeeded for unknown session, expected error")
	}
}

// --- END OF FILE ---
// Ensure no duplicate definitions exist below this line
