package orchestrator

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core" // Ensure core is imported

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Using SpyCollaborativeHandler from orchestrator_test_helpers.go

func TestCollaborateOrchestrator_FanOut(t *testing.T) {
	// FIX: Use exported constructor
	o := NewCollaborativeOrchestrator()
	numHandlers := 3
	handlers := make([]*SpyCollaborativeHandler, numHandlers)
	for i := 0; i < numHandlers; i++ {
		handlers[i] = &SpyCollaborativeHandler{AgentName: fmt.Sprintf("handler-%d", i)}
		o.RegisterAgent(fmt.Sprintf("handler-%d", i), handlers[i])
	}

	// FIX: Use correct NewEvent signature and EventData type
	event := agentflow.NewEvent("", agentflow.EventData{"type": "fanout"}, nil)
	o.Dispatch(event)

	for i, handler := range handlers {
		count := handler.EventCount()
		if count != 1 {
			t.Errorf("Handler %d: want 1 event, got %d", i, count)
		}
		// FIX: Use GetID() method
		if count > 0 && handler.events[0] != event.GetID() {
			// FIX: Use GetID() method
			t.Errorf("Handler %d: want event %s, got %s", i, event.GetID(), handler.events[0])
		}
	}
}

func TestCollaborateOrchestrator_ErrorAggregation(t *testing.T) {
	// FIX: Use exported constructor
	o := NewCollaborativeOrchestrator()
	handler1 := &SpyCollaborativeHandler{AgentName: "handler-1"}
	handler2 := &SpyCollaborativeHandler{AgentName: "handler-2", failOn: "evt-fail"}
	handler3 := &SpyCollaborativeHandler{AgentName: "handler-3", failOn: "evt-fail"}
	o.RegisterAgent("handler-1", handler1)
	o.RegisterAgent("handler-2", handler2)
	o.RegisterAgent("handler-3", handler3)

	// FIX: Use correct NewEvent signature and EventData type
	event := agentflow.NewEvent("", agentflow.EventData{"type": "fail"}, nil)
	event.SetID("evt-fail")
	errs := o.Dispatch(event)

	if len(errs) != 2 {
		t.Fatalf("Dispatch: want 2 errors, got %d: %v", len(errs), errs)
	}

	// FIX: Use GetID() method
	if count := handler1.EventCount(); count != 1 || handler1.events[0] != event.GetID() {
		t.Errorf("Handler 1 (success): unexpected events, got %v", handler1.events)
	}
	// FIX: Use GetID() method
	if count := handler2.EventCount(); count != 1 || handler2.events[0] != event.GetID() {
		t.Errorf("Handler 2 (fail): unexpected events, got %v", handler2.events)
	}
	// FIX: Use GetID() method
	if count := handler3.EventCount(); count != 1 || handler3.events[0] != event.GetID() {
		t.Errorf("Handler 3 (fail): unexpected events, got %v", handler3.events)
	}
}

func TestCollaborateOrchestrator_PartialFailure(t *testing.T) {
	// FIX: Use exported constructor
	o := NewCollaborativeOrchestrator()
	handlerOK := &SpyCollaborativeHandler{AgentName: "handler-ok"}
	handlerFail := &SpyCollaborativeHandler{AgentName: "handler-fail", failOn: "evt-partial"}
	o.RegisterAgent("handler-ok", handlerOK)
	o.RegisterAgent("handler-fail", handlerFail)

	// FIX: Use correct NewEvent signature and EventData type
	event := agentflow.NewEvent("", agentflow.EventData{"type": "partial"}, nil)
	event.SetID("evt-partial")
	errs := o.Dispatch(event)

	if len(errs) != 1 {
		t.Fatalf("Dispatch: want 1 error, got %d: %v", len(errs), errs)
	}

	// FIX: Use GetID() method
	if count := handlerOK.EventCount(); count != 1 || handlerOK.events[0] != event.GetID() {
		t.Errorf("Handler OK: unexpected events, got %v", handlerOK.events)
	}
	// FIX: Use GetID() method
	if count := handlerFail.EventCount(); count != 1 || handlerFail.events[0] != event.GetID() {
		t.Errorf("Handler Fail: unexpected events, got %v", handlerFail.events)
	}
}

func TestCollaborateOrchestrator_NoHandlers(t *testing.T) {
	// FIX: Use exported constructor
	o := NewCollaborativeOrchestrator()
	// FIX: Use correct NewEvent signature and EventData type
	errs := o.Dispatch(agentflow.NewEvent("", agentflow.EventData{"type": "no-handler"}, nil))
	if len(errs) > 0 {
		t.Errorf("Dispatch with no handlers returned errors: %v", errs)
	}
}

func TestCollaborateOrchestrator_ConcurrentDispatch(t *testing.T) {
	// FIX: Use exported constructor
	o := NewCollaborativeOrchestrator()
	numHandlers := 3
	handlers := make([]*SpyCollaborativeHandler, numHandlers)
	for i := 0; i < numHandlers; i++ {
		handlers[i] = &SpyCollaborativeHandler{AgentName: fmt.Sprintf("handler-%d", i)}
		o.RegisterAgent(fmt.Sprintf("handler-%d", i), handlers[i])
	}

	numEvents := 50
	var wg sync.WaitGroup
	wg.Add(numEvents)
	for i := 0; i < numEvents; i++ {
		go func(i int) {
			defer wg.Done()
			// FIX: Use correct NewEvent signature and EventData type
			o.Dispatch(agentflow.NewEvent("", agentflow.EventData{"event": i}, nil))
		}(i)
	}
	wg.Wait()

	for i, handler := range handlers {
		count := handler.EventCount()
		if count != numEvents {
			t.Errorf("Handler %d: want %d events, got %d", i, numEvents, count)
		}
	}
}

// Optional: Test with timeout (requires modifying Dispatch to accept context)
/*
func TestCollaborateOrchestrator_Timeout(t *testing.T) {
    // ... implementation using SpyCollaborativeHandler and context ...
}
*/

// --- Tests for DispatchAll ---

func TestCollaborativeOrchestrator_DispatchAll(t *testing.T) {
	// FIX: Use exported constructor
	o := NewCollaborativeOrchestrator()
	numHandlers := 3
	handlers := make([]*SpyCollaborativeHandler, numHandlers)
	for i := 0; i < numHandlers; i++ {
		handlers[i] = &SpyCollaborativeHandler{AgentName: fmt.Sprintf("handler-%d", i)}
		o.RegisterAgent(fmt.Sprintf("handler-%d", i), handlers[i])
	}

	// FIX: Use correct NewEvent signature and EventData type
	event := agentflow.NewEvent("target", agentflow.EventData{"data": "value"}, nil)
	errs := o.DispatchAll(event)

	if len(errs) > 0 {
		t.Errorf("DispatchAll returned unexpected errors: %v", errs)
	}

	for i, handler := range handlers {
		if !handler.HandleCalled {
			t.Errorf("Handler %d was not called", i)
		}
		// FIX: Use GetID() method
		if handler.LastEvent == nil || handler.LastEvent.GetID() != event.GetID() {
			t.Errorf("Handler %d did not receive the correct event", i)
		}
	}
}

func TestCollaborativeOrchestrator_DispatchAll_HandlerFailure(t *testing.T) {
	// FIX: Use exported constructor
	o := NewCollaborativeOrchestrator()
	handler1 := &SpyCollaborativeHandler{AgentName: "handler1"}
	simulatedError := errors.New("handler2 failed")
	handler2 := &SpyCollaborativeHandler{AgentName: "handler2", ReturnError: simulatedError}
	handler3 := &SpyCollaborativeHandler{AgentName: "handler3"}

	o.RegisterAgent("handler1", handler1)
	o.RegisterAgent("handler2", handler2)
	o.RegisterAgent("handler3", handler3)

	// FIX: Use correct NewEvent signature and EventData type
	event := agentflow.NewEvent("target", agentflow.EventData{"data": "value"}, nil)
	errs := o.DispatchAll(event)

	if len(errs) != 1 {
		t.Fatalf("DispatchAll should have returned exactly 1 error, got %d: %v", len(errs), errs)
	}
	found := false
	for _, err := range errs {
		if errors.Is(err, simulatedError) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error '%v' not found in returned errors: %v", simulatedError, errs)
	}

	if !handler1.HandleCalled || !handler2.HandleCalled || !handler3.HandleCalled {
		t.Errorf("Not all handlers were called despite one failing")
	}
}

func TestCollaborativeOrchestrator_DispatchAll_NoHandlers(t *testing.T) {
	// FIX: Use exported constructor
	o := NewCollaborativeOrchestrator()
	// FIX: Use correct NewEvent signature and EventData type
	event := agentflow.NewEvent("target", agentflow.EventData{"data": "value"}, nil)
	errs := o.DispatchAll(event)

	if len(errs) > 0 {
		t.Errorf("DispatchAll with no handlers returned errors: %v", errs)
	}
}

func TestCollaborativeOrchestrator_ConcurrentDispatchAll(t *testing.T) {
	// FIX: Use exported constructor
	o := NewCollaborativeOrchestrator()
	numHandlers := 5
	handlers := make([]*SpyCollaborativeHandler, numHandlers)
	for i := 0; i < numHandlers; i++ {
		handlers[i] = &SpyCollaborativeHandler{AgentName: fmt.Sprintf("handler-%d", i)}
		o.RegisterAgent(fmt.Sprintf("handler-%d", i), handlers[i])
	}

	numEvents := 10
	var wg sync.WaitGroup
	wg.Add(numEvents)
	for i := 0; i < numEvents; i++ {
		go func(i int) {
			defer wg.Done()
			// FIX: Use correct NewEvent signature and EventData type
			event := agentflow.NewEvent("target", agentflow.EventData{"event": i}, nil)
			o.DispatchAll(event)
		}(i)
	}
	wg.Wait()

	expectedCallsPerHandler := numEvents
	for i, handler := range handlers {
		count := handler.EventCount()
		if count != expectedCallsPerHandler {
			t.Errorf("Handler %d: want %d calls, got %d", i, expectedCallsPerHandler, count)
		}
	}
}

func TestCollaborativeOrchestrator_DispatchAll_NilEvent(t *testing.T) {
	// FIX: Use exported constructor
	o := NewCollaborativeOrchestrator()
	handler := &SpyCollaborativeHandler{AgentName: "handler1"}
	o.RegisterAgent("handler1", handler)

	errs := o.DispatchAll(nil)

	if len(errs) > 0 {
		t.Logf("DispatchAll(nil) returned errors: %v (needs review if unexpected)", errs)
	} else {
		t.Log("DispatchAll(nil) returned no errors (as expected or needs review)")
	}

	if handler.HandleCalled {
		t.Errorf("Handler was called for nil event")
	} else {
		t.Logf("Handler was not called for nil event (expected)")
	}
}

// TestCollaborativeOrchestrator_Dispatch verifies concurrent dispatch to multiple handlers.
func TestCollaborativeOrchestrator_Dispatch(t *testing.T) {
	handler1 := &SpyCollaborativeHandler{AgentName: "handler1"}
	handler2 := &SpyCollaborativeHandler{AgentName: "handler2"}
	handler3 := &SpyCollaborativeHandler{AgentName: "handler3", ReturnError: errors.New("handler3 error")}

	// Use NewCollaborativeOrchestrator which doesn't require a registry
	orchestrator := NewCollaborativeOrchestrator()

	// Register handlers using the correct interface type (EventHandler)
	orchestrator.RegisterAgent("h1", handler1) // AgentID is less relevant here
	orchestrator.RegisterAgent("h2", handler2)
	orchestrator.RegisterAgent("h3", handler3)

	testEvent := agentflow.NewEvent("source", agentflow.EventData{"key": "value"}, nil)
	// FIX: Use GetID() method
	eventID := testEvent.GetID() // Store ID for checks

	errs := orchestrator.Dispatch(testEvent)

	// FIX: Use GetID() method
	t.Logf("Dispatch finished for event %s with errors: %v", testEvent.GetID(), errs)

	assert.True(t, handler1.HandleCalled, "Handler 1 should have been called")
	assert.True(t, handler2.HandleCalled, "Handler 2 should have been called")
	assert.True(t, handler3.HandleCalled, "Handler 3 should have been called")

	// FIX: Use stored eventID variable
	assert.Equal(t, eventID, handler1.LastEvent.GetID(), "Handler 1 received wrong event ID")
	assert.Equal(t, eventID, handler2.LastEvent.GetID(), "Handler 2 received wrong event ID")
	assert.Equal(t, eventID, handler3.LastEvent.GetID(), "Handler 3 received wrong event ID")

	require.Len(t, errs, 1, "Expected exactly one error")
	assert.EqualError(t, errs[0], "handler3 error", "Expected error from handler3")
}

// TestCollaborativeOrchestrator_Dispatch_NoHandlers tests dispatch with no registered handlers.
func TestCollaborativeOrchestrator_Dispatch_NoHandlers(t *testing.T) {
	orchestrator := NewCollaborativeOrchestrator()
	testEvent := agentflow.NewEvent("source", agentflow.EventData{"key": "value"}, nil)

	errs := orchestrator.Dispatch(testEvent)

	assert.Empty(t, errs, "Expected no errors when no handlers are registered")
}

// TestCollaborativeOrchestrator_Dispatch_MultipleErrors tests dispatch where multiple handlers return errors.
func TestCollaborativeOrchestrator_Dispatch_MultipleErrors(t *testing.T) {
	err1 := errors.New("handler1 specific error")
	err2 := errors.New("handler2 specific error")
	handler1 := &SpyCollaborativeHandler{AgentName: "handler1", ReturnError: err1}
	handler2 := &SpyCollaborativeHandler{AgentName: "handler2", ReturnError: err2}
	handler3 := &SpyCollaborativeHandler{AgentName: "handler3"} // No error

	orchestrator := NewCollaborativeOrchestrator()
	orchestrator.RegisterAgent("h1", handler1)
	orchestrator.RegisterAgent("h2", handler2)
	orchestrator.RegisterAgent("h3", handler3)

	testEvent := agentflow.NewEvent("source", agentflow.EventData{"data": 123}, nil)
	// FIX: Use GetID() method
	eventID := testEvent.GetID() // Store ID

	errs := orchestrator.Dispatch(testEvent)

	// FIX: Use GetID() method
	t.Logf("Dispatch finished for event %s with errors: %v", testEvent.GetID(), errs)

	assert.True(t, handler1.HandleCalled, "Handler 1 should have been called")
	assert.True(t, handler2.HandleCalled, "Handler 2 should have been called")
	assert.True(t, handler3.HandleCalled, "Handler 3 should have been called")

	// FIX: Use stored eventID variable
	assert.Equal(t, eventID, handler1.LastEvent.GetID(), "Handler 1 received wrong event ID")
	assert.Equal(t, eventID, handler2.LastEvent.GetID(), "Handler 2 received wrong event ID")
	assert.Equal(t, eventID, handler3.LastEvent.GetID(), "Handler 3 received wrong event ID")

	require.Len(t, errs, 2, "Expected two errors")
	// Check if both specific errors are present (order might vary)
	assert.Contains(t, errs, err1)
	assert.Contains(t, errs, err2)
}

// TestCollaborativeOrchestrator_Dispatch_Concurrency simulates concurrent dispatches.
func TestCollaborativeOrchestrator_Dispatch_Concurrency(t *testing.T) {
	numHandlers := 5
	numEvents := 10
	delay := 10 * time.Millisecond // Short delay to encourage concurrency issues

	orchestrator := NewCollaborativeOrchestrator()
	handlers := make([]*SlowSpyEventHandler, numHandlers) // Use SlowSpyEventHandler
	for i := 0; i < numHandlers; i++ {
		// Wrap SpyCollaborativeHandler logic in SlowSpyEventHandler if needed,
		// or adjust SlowSpyEventHandler to match EventHandler signature.
		// Assuming SlowSpyEventHandler is compatible or adjusted:
		handlers[i] = &SlowSpyEventHandler{
			SpyEventHandler: SpyEventHandler{AgentName: fmt.Sprintf("handler-%d", i)},
			delay:           delay,
		}
		orchestrator.RegisterAgent(fmt.Sprintf("h%d", i), handlers[i])
	}

	var wg sync.WaitGroup
	wg.Add(numEvents)

	for i := 0; i < numEvents; i++ {
		go func(eventNum int) {
			defer wg.Done()
			eventData := agentflow.EventData{"eventNum": eventNum}
			event := agentflow.NewEvent("concurrentSource", eventData, nil)
			// FIX: Use GetID() method
			t.Logf("Dispatching concurrent event %s (num %d)", event.GetID(), eventNum)
			errs := orchestrator.Dispatch(event)
			// FIX: Use GetID() method
			if len(errs) > 0 {
				t.Errorf("Concurrent dispatch for event %s (num %d) returned errors: %v", event.GetID(), eventNum, errs)
			}
		}(i)
	}

	wg.Wait() // Wait for all dispatches to complete

	// Verify each handler received all events
	for i, handler := range handlers {
		assert.Equal(t, numEvents, handler.EventCount(), "Handler %d did not receive all events", i)
	}
}

// TestCollaborativeOrchestrator_Stop tests the Stop method (basic check).
func TestCollaborativeOrchestrator_Stop(t *testing.T) {
	orchestrator := NewCollaborativeOrchestrator()
	// Add a handler to ensure it doesn't interfere
	handler := &SpyCollaborativeHandler{AgentName: "handler1"}
	orchestrator.RegisterAgent("h1", handler)

	// Call Stop - primarily checking it doesn't panic
	assert.NotPanics(t, func() { orchestrator.Stop() }, "Stop method should not panic")

	// Optional: Verify state after stop if applicable (e.g., cannot dispatch)
	// testEvent := agentflow.NewEvent("source", agentflow.EventData{}, nil)
	// errs := orchestrator.Dispatch(testEvent)
	// assert.NotEmpty(t, errs, "Dispatch should ideally fail or warn after Stop")
	// Or check an internal 'stopped' flag if implemented.
}

// TestCollaborativeOrchestrator_RegisterAgent_NilHandler tests registering a nil handler.
func TestCollaborativeOrchestrator_RegisterAgent_NilHandler(t *testing.T) {
	orchestrator := NewCollaborativeOrchestrator()

	// Use assert.NotPanics or similar if RegisterAgent logs but doesn't error
	assert.NotPanics(t, func() {
		orchestrator.RegisterAgent("nilAgent", nil)
	}, "Registering nil handler should not panic")

	// Verify no handler was actually added (check internal state if possible,
	// or dispatch an event and ensure no handlers are called).
	testEvent := agentflow.NewEvent("source", agentflow.EventData{}, nil)
	errs := orchestrator.Dispatch(testEvent)
	assert.Empty(t, errs, "Dispatch should have no errors as no handlers should be registered")
}

// TestSpyCollaborativeHandler_ErrorHandling tests the mock handler's error return logic.
func TestSpyCollaborativeHandler_ErrorHandling(t *testing.T) {
	specificError := errors.New("specific test error")
	handler := &SpyCollaborativeHandler{
		AgentName:   "errorTester",
		ReturnError: specificError,
	}

	testEvent := agentflow.NewEvent("source", agentflow.EventData{}, nil)
	err := handler.Handle(testEvent)
	assert.Equal(t, specificError, err, "Handler should return the specified error")

	// Test deliberate failure on specific event ID
	failEventID := "fail-on-this-id"
	handlerWithError := &SpyCollaborativeHandler{
		AgentName: "failer",
		failOn:    failEventID, // Set the ID to fail on
	}

	eventToPass := agentflow.NewEvent("source", agentflow.EventData{"pass": true}, nil)
	// FIX: Use NewEvent and SetID
	eventToFail := agentflow.NewEvent("source", agentflow.EventData{"fail": true}, nil)
	eventToFail.SetID(failEventID) // Set the ID after creation

	errPass := handlerWithError.Handle(eventToPass)
	assert.NoError(t, errPass, "Handler should not error on event %s", eventToPass.GetID())

	errFail := handlerWithError.Handle(eventToFail)
	assert.Error(t, errFail, "Handler should error on event %s", eventToFail.GetID())
	expectedErrMsg := fmt.Sprintf("handler 'failer' failed deliberately for event '%s'", eventToFail.GetID())
	assert.EqualError(t, errFail, expectedErrMsg, "Error message mismatch for deliberate failure")
}

// TestSpyCollaborativeHandler_EventTracking tests if the mock handler correctly tracks events.
func TestSpyCollaborativeHandler_EventTracking(t *testing.T) {
	handler := &SpyCollaborativeHandler{AgentName: "tracker"}
	event1 := agentflow.NewEvent("source", agentflow.EventData{"num": 1}, nil)
	event2 := agentflow.NewEvent("source", agentflow.EventData{"num": 2}, nil)

	_ = handler.Handle(event1)
	_ = handler.Handle(event2)

	assert.Equal(t, 2, handler.EventCount(), "Expected handler to have handled 2 events")
	// FIX: Use GetID() method
	assert.Equal(t, event2.GetID(), handler.LastEvent.GetID(), "Last event ID mismatch")

	handledIDs := handler.GetEvents()
	require.Len(t, handledIDs, 2, "Expected 2 event IDs tracked")
	// FIX: Use GetID() method
	assert.Equal(t, event1.GetID(), handledIDs[0], "First tracked event ID mismatch")
	assert.Equal(t, event2.GetID(), handledIDs[1], "Second tracked event ID mismatch")
}
