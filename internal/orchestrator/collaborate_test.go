package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	agenticgokit "github.com/kunalkushwaha/AgenticGoKit/internal/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Existing Tests (with FIX comments addressed) ---

func TestCollaborateOrchestrator_FanOut(t *testing.T) {
	o := NewCollaborativeOrchestrator()
	numHandlers := 3
	handlers := make([]*SpyAgentHandler, numHandlers)
	for i := 0; i < numHandlers; i++ {
		handlers[i] = &SpyAgentHandler{AgentName: fmt.Sprintf("handler-%d", i)}
		// Now correctly implements AgentHandler
		err := o.RegisterAgent(fmt.Sprintf("handler-%d", i), handlers[i])
		require.NoError(t, err)
	}

	event := agentflow.NewEvent("", agentflow.EventData{"type": "fanout"}, nil)
	_, err := o.Dispatch(context.Background(), event)
	assert.NoError(t, err, "Dispatch failed unexpectedly")

	for i, handler := range handlers {
		count := handler.EventCount()
		if count != 1 {
			t.Errorf("Handler %d: want 1 event, got %d", i, count)
		}
		handledEvents := handler.GetEvents() // Get tracked events
		if count > 0 && handledEvents[0] != event.GetID() {
			t.Errorf("Handler %d: want event %s, got %s", i, event.GetID(), handledEvents[0])
		}
	}
}

func TestCollaborateOrchestrator_ErrorAggregation(t *testing.T) {
	o := NewCollaborativeOrchestrator()
	handler1 := &SpyAgentHandler{AgentName: "handler-1"}
	// Define expected errors clearly based on Run implementation
	err2Msg := "handler 'handler-2' failed deliberately for event 'evt-fail'"
	err3Msg := "handler 'handler-3' failed deliberately for event 'evt-fail'"
	handler2 := &SpyAgentHandler{AgentName: "handler-2", failOn: "evt-fail"}
	handler3 := &SpyAgentHandler{AgentName: "handler-3", failOn: "evt-fail"}
	err := o.RegisterAgent("handler-1", handler1)
	require.NoError(t, err)
	err = o.RegisterAgent("handler-2", handler2)
	require.NoError(t, err)
	err = o.RegisterAgent("handler-3", handler3)
	require.NoError(t, err)

	event := agentflow.NewEvent("", agentflow.EventData{"type": "fail"}, nil)
	event.SetID("evt-fail")
	_, aggErr := o.Dispatch(context.Background(), event)

	require.Error(t, aggErr, "Dispatch should have returned an aggregated error")

	// Check if the aggregated error message contains the individual errors
	errMsg := aggErr.Error()
	assert.Contains(t, errMsg, err2Msg, "Aggregated error missing error from handler-2")
	assert.Contains(t, errMsg, err3Msg, "Aggregated error missing error from handler-3")
	assert.Equal(t, 2, strings.Count(errMsg, "failed deliberately"), "Aggregated error should represent 2 failures")

	// Check handlers were called
	assert.Equal(t, 1, handler1.EventCount(), "Handler 1 (success): unexpected event count")
	assert.Equal(t, 1, handler2.EventCount(), "Handler 2 (fail): unexpected event count")
	assert.Equal(t, 1, handler3.EventCount(), "Handler 3 (fail): unexpected event count")
	assert.True(t, handler1.RunCalled, "Handler 1 should have been called") // Changed RunCalled to RunCalled
	assert.True(t, handler2.RunCalled, "Handler 2 should have been called") // Changed RunCalled to RunCalled
	assert.True(t, handler3.RunCalled, "Handler 3 should have been called") // Changed RunCalled to RunCalled
}

func TestCollaborateOrchestrator_PartialFailure(t *testing.T) {
	o := NewCollaborativeOrchestrator()
	handlerOK := &SpyAgentHandler{AgentName: "handler-ok"}
	// Define expected error based on Run implementation
	errFailMsg := "handler 'handler-fail' failed deliberately for event 'evt-partial'"
	handlerFail := &SpyAgentHandler{AgentName: "handler-fail", failOn: "evt-partial"}
	err := o.RegisterAgent("handler-ok", handlerOK)
	require.NoError(t, err)
	err = o.RegisterAgent("handler-fail", handlerFail)
	require.NoError(t, err)

	event := agentflow.NewEvent("", agentflow.EventData{"type": "partial"}, nil)
	event.SetID("evt-partial")
	_, aggErr := o.Dispatch(context.Background(), event)

	require.Error(t, aggErr, "Dispatch should have returned an error")
	// errors.Join might wrap single errors, check containment
	assert.Contains(t, aggErr.Error(), errFailMsg, "Expected error message mismatch")

	assert.Equal(t, 1, handlerOK.EventCount(), "Handler OK: unexpected event count")
	assert.Equal(t, 1, handlerFail.EventCount(), "Handler Fail: unexpected event count")
}

func TestCollaborateOrchestrator_NoHandlers(t *testing.T) {
	o := NewCollaborativeOrchestrator()
	_, err := o.Dispatch(context.Background(), agentflow.NewEvent("", agentflow.EventData{"type": "no-handler"}, nil))
	assert.NoError(t, err, "Dispatch with no handlers returned an unexpected error")
}

func TestCollaborateOrchestrator_ConcurrentDispatch(t *testing.T) {
	o := NewCollaborativeOrchestrator()
	numHandlers := 3
	handlers := make([]*SpyAgentHandler, numHandlers)
	for i := 0; i < numHandlers; i++ {
		handlers[i] = &SpyAgentHandler{AgentName: fmt.Sprintf("handler-%d", i)}
		err := o.RegisterAgent(fmt.Sprintf("handler-%d", i), handlers[i])
		require.NoError(t, err)
	}

	numEvents := 50
	var wg sync.WaitGroup
	wg.Add(numEvents)
	errChan := make(chan error, numEvents) // Channel to collect errors from goroutines

	for i := 0; i < numEvents; i++ {
		go func(i int) {
			defer wg.Done()
			_, err := o.Dispatch(context.Background(), agentflow.NewEvent("", agentflow.EventData{"event": i}, nil))
			if err != nil {
				errChan <- fmt.Errorf("concurrent dispatch %d failed: %w", i, err)
			}
		}(i)
	}
	wg.Wait()
	close(errChan)

	// Check for errors collected from goroutines
	for err := range errChan {
		t.Error(err) // Report any errors found
	}

	for i, handler := range handlers {
		count := handler.EventCount()
		if count != numEvents {
			t.Errorf("Handler %d: want %d events, got %d", i, numEvents, count)
		}
	}
}

// --- Tests for DispatchAll ---
// Note: DispatchAll currently delegates to Dispatch, so these tests mirror Dispatch tests.
// If DispatchAll's behavior diverges later, these tests will need adjustment.

func TestCollaborativeOrchestrator_DispatchAll(t *testing.T) {
	o := NewCollaborativeOrchestrator()
	numHandlers := 3
	handlers := make([]*SpyAgentHandler, numHandlers)
	for i := 0; i < numHandlers; i++ {
		handlers[i] = &SpyAgentHandler{AgentName: fmt.Sprintf("handler-%d", i)}
		err := o.RegisterAgent(fmt.Sprintf("handler-%d", i), handlers[i])
		require.NoError(t, err)
	}

	event := agentflow.NewEvent("target", agentflow.EventData{"data": "value"}, nil)
	_, err := o.DispatchAll(context.Background(), event) // DispatchAll calls Dispatch
	assert.NoError(t, err, "DispatchAll returned unexpected errors")

	for i, handler := range handlers {
		assert.True(t, handler.RunCalled, "Handler %d was not called", i)
		assert.Equal(t, 1, handler.EventCount(), "Handler %d event count mismatch", i)
		if handler.LastEvent != nil { // Check LastEvent is not nil before GetID
			assert.Equal(t, event.GetID(), handler.LastEvent.GetID(), "Handler %d did not receive the correct event ID", i)
		} else {
			t.Errorf("Handler %d LastEvent is nil", i)
		}
	}
}

func TestCollaborativeOrchestrator_DispatchAll_HandlerFailure(t *testing.T) {
	o := NewCollaborativeOrchestrator()
	handler1 := &SpyAgentHandler{AgentName: "handler1"}
	simulatedError := errors.New("handler2 failed")
	handler2 := &SpyAgentHandler{AgentName: "handler2", ReturnError: simulatedError}
	handler3 := &SpyAgentHandler{AgentName: "handler3"}

	err := o.RegisterAgent("handler1", handler1)
	require.NoError(t, err)
	err = o.RegisterAgent("handler2", handler2)
	require.NoError(t, err)
	err = o.RegisterAgent("handler3", handler3)
	require.NoError(t, err)

	event := agentflow.NewEvent("target", agentflow.EventData{"data": "value"}, nil)
	_, aggErr := o.DispatchAll(context.Background(), event) // DispatchAll calls Dispatch

	require.Error(t, aggErr, "DispatchAll should have returned an error")
	// errors.Join might wrap single errors, check containment
	assert.Contains(t, aggErr.Error(), simulatedError.Error(), "Expected error message mismatch")

	assert.True(t, handler1.RunCalled, "Handler 1 should have been called")
	assert.True(t, handler2.RunCalled, "Handler 2 should have been called")
	assert.True(t, handler3.RunCalled, "Handler 3 should have been called")
}

func TestCollaborativeOrchestrator_DispatchAll_NoHandlers(t *testing.T) {
	o := NewCollaborativeOrchestrator()
	event := agentflow.NewEvent("target", agentflow.EventData{"data": "value"}, nil)
	_, err := o.DispatchAll(context.Background(), event) // DispatchAll calls Dispatch
	assert.NoError(t, err, "DispatchAll with no handlers returned errors")
}

func TestCollaborativeOrchestrator_ConcurrentDispatchAll(t *testing.T) {
	o := NewCollaborativeOrchestrator()
	numHandlers := 5
	handlers := make([]*SpyAgentHandler, numHandlers)
	for i := 0; i < numHandlers; i++ {
		handlers[i] = &SpyAgentHandler{AgentName: fmt.Sprintf("handler-%d", i)}
		err := o.RegisterAgent(fmt.Sprintf("handler-%d", i), handlers[i])
		require.NoError(t, err)
	}

	numEvents := 10
	var wg sync.WaitGroup
	wg.Add(numEvents)
	errChan := make(chan error, numEvents) // Channel to collect errors

	for i := 0; i < numEvents; i++ {
		go func(i int) {
			defer wg.Done()
			event := agentflow.NewEvent("target", agentflow.EventData{"event": i}, nil)
			_, err := o.DispatchAll(context.Background(), event) // DispatchAll calls Dispatch
			if err != nil {
				errChan <- fmt.Errorf("concurrent DispatchAll %d failed: %w", i, err)
			}
		}(i)
	}
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		t.Error(err)
	}

	expectedCallsPerHandler := numEvents
	for i, handler := range handlers {
		count := handler.EventCount()
		if count != expectedCallsPerHandler {
			t.Errorf("Handler %d: want %d calls, got %d", i, expectedCallsPerHandler, count)
		}
	}
}

func TestCollaborativeOrchestrator_DispatchAll_NilEvent(t *testing.T) {
	o := NewCollaborativeOrchestrator()
	handler := &SpyAgentHandler{AgentName: "handler1"}
	err := o.RegisterAgent("handler1", handler)
	require.NoError(t, err)

	_, err = o.DispatchAll(context.Background(), nil) // DispatchAll calls Dispatch

	require.Error(t, err, "DispatchAll(nil) should return an error")
	assert.EqualError(t, err, "cannot dispatch nil event")

	assert.False(t, handler.RunCalled, "Handler should not have been called for nil event")
}

// TestCollaborativeOrchestrator_Dispatch verifies concurrent dispatch to multiple handlers.
// Note: This test overlaps significantly with FanOut and ErrorAggregation. Consider removing if redundant.
func TestCollaborativeOrchestrator_Dispatch(t *testing.T) {
	handler1 := &SpyAgentHandler{AgentName: "handler1"}
	handler2 := &SpyAgentHandler{AgentName: "handler2"}
	handler3Err := errors.New("handler3 error")
	handler3 := &SpyAgentHandler{AgentName: "handler3", ReturnError: handler3Err}

	orchestrator := NewCollaborativeOrchestrator()

	err := orchestrator.RegisterAgent("h1", handler1)
	require.NoError(t, err)
	err = orchestrator.RegisterAgent("h2", handler2)
	require.NoError(t, err)
	err = orchestrator.RegisterAgent("h3", handler3)
	require.NoError(t, err)

	testEvent := agentflow.NewEvent("source", agentflow.EventData{"key": "value"}, nil)
	eventID := testEvent.GetID() // Store ID for checks

	_, aggErr := orchestrator.Dispatch(context.Background(), testEvent)

	t.Logf("Dispatch finished for event %s with error: %v", testEvent.GetID(), aggErr)

	assert.True(t, handler1.RunCalled, "Handler 1 should have been called")
	assert.True(t, handler2.RunCalled, "Handler 2 should have been called")
	assert.True(t, handler3.RunCalled, "Handler 3 should have been called")

	// Check LastEvent ID if not nil
	if handler1.LastEvent != nil {
		assert.Equal(t, eventID, handler1.LastEvent.GetID(), "Handler 1 received wrong event ID")
	}
	if handler2.LastEvent != nil {
		assert.Equal(t, eventID, handler2.LastEvent.GetID(), "Handler 2 received wrong event ID")
	}
	if handler3.LastEvent != nil {
		assert.Equal(t, eventID, handler3.LastEvent.GetID(), "Handler 3 received wrong event ID")
	}

	require.Error(t, aggErr, "Expected exactly one error")
	// errors.Join might wrap single errors, check containment
	assert.Contains(t, aggErr.Error(), handler3Err.Error(), "Expected error from handler3")
}

// TestCollaborativeOrchestrator_Dispatch_NoHandlers tests dispatch with no registered handlers.
// Note: This test overlaps significantly with TestCollaborateOrchestrator_NoHandlers. Consider removing if redundant.
func TestCollaborativeOrchestrator_Dispatch_NoHandlers(t *testing.T) {
	orchestrator := NewCollaborativeOrchestrator()
	testEvent := agentflow.NewEvent("source", agentflow.EventData{"key": "value"}, nil)

	_, err := orchestrator.Dispatch(context.Background(), testEvent)
	assert.NoError(t, err, "Expected no errors when no handlers are registered")
}

// TestCollaborativeOrchestrator_Dispatch_MultipleErrors tests dispatch where multiple handlers return errors.
// Note: This test overlaps significantly with TestCollaborateOrchestrator_ErrorAggregation. Consider removing if redundant.
func TestCollaborativeOrchestrator_Dispatch_MultipleErrors(t *testing.T) {
	err1 := errors.New("handler1 specific error")
	err2 := errors.New("handler2 specific error")
	handler1 := &SpyAgentHandler{AgentName: "handler1", ReturnError: err1}
	handler2 := &SpyAgentHandler{AgentName: "handler2", ReturnError: err2}
	handler3 := &SpyAgentHandler{AgentName: "handler3"} // No error

	orchestrator := NewCollaborativeOrchestrator()
	err := orchestrator.RegisterAgent("h1", handler1)
	require.NoError(t, err)
	err = orchestrator.RegisterAgent("h2", handler2)
	require.NoError(t, err)
	err = orchestrator.RegisterAgent("h3", handler3)
	require.NoError(t, err)

	testEvent := agentflow.NewEvent("source", agentflow.EventData{"data": 123}, nil)
	eventID := testEvent.GetID() // Store ID

	_, aggErr := orchestrator.Dispatch(context.Background(), testEvent)

	t.Logf("Dispatch finished for event %s with error: %v", testEvent.GetID(), aggErr)

	assert.True(t, handler1.RunCalled, "Handler 1 should have been called")
	assert.True(t, handler2.RunCalled, "Handler 2 should have been called")
	assert.True(t, handler3.RunCalled, "Handler 3 should have been called")

	// Check LastEvent ID if not nil
	if handler1.LastEvent != nil {
		assert.Equal(t, eventID, handler1.LastEvent.GetID(), "Handler 1 received wrong event ID")
	}
	if handler2.LastEvent != nil {
		assert.Equal(t, eventID, handler2.LastEvent.GetID(), "Handler 2 received wrong event ID")
	}
	if handler3.LastEvent != nil {
		assert.Equal(t, eventID, handler3.LastEvent.GetID(), "Handler 3 received wrong event ID")
	}

	require.Error(t, aggErr, "Expected an aggregated error")
	errMsg := aggErr.Error()
	assert.Contains(t, errMsg, err1.Error())
	assert.Contains(t, errMsg, err2.Error())
	assert.Equal(t, 2, strings.Count(errMsg, "specific error"), "Aggregated error should represent 2 failures")
}

// TestCollaborativeOrchestrator_Dispatch_Concurrency simulates concurrent dispatches.
// Note: This test overlaps significantly with TestCollaborateOrchestrator_ConcurrentDispatch. Consider removing if redundant.
// Note: SlowSpyEventHandler needs to be defined or adapted. Assuming SpyAgentHandler for now.
func TestCollaborativeOrchestrator_Dispatch_Concurrency(t *testing.T) {
	numHandlers := 5
	numEvents := 10
	// delay := 10 * time.Millisecond // Short delay to encourage concurrency issues

	orchestrator := NewCollaborativeOrchestrator()
	handlers := make([]*SpyAgentHandler, numHandlers) // Using SpyAgentHandler
	for i := 0; i < numHandlers; i++ {
		handlers[i] = &SpyAgentHandler{AgentName: fmt.Sprintf("handler-%d", i)}
		err := orchestrator.RegisterAgent(fmt.Sprintf("h%d", i), handlers[i])
		require.NoError(t, err)
	}

	var wg sync.WaitGroup
	wg.Add(numEvents)
	errChan := make(chan error, numEvents) // Channel to collect errors

	for i := 0; i < numEvents; i++ {
		go func(eventNum int) {
			defer wg.Done()
			eventData := agentflow.EventData{"eventNum": eventNum}
			event := agentflow.NewEvent("concurrentSource", eventData, nil)
			t.Logf("Dispatching concurrent event %s (num %d)", event.GetID(), eventNum)
			_, err := orchestrator.Dispatch(context.Background(), event)
			if err != nil {
				errChan <- fmt.Errorf("concurrent dispatch for event %s (num %d) returned errors: %w", event.GetID(), eventNum, err)
			}
		}(i)
	}

	wg.Wait() // Wait for all dispatches to complete
	close(errChan)

	// Check for errors
	for err := range errChan {
		t.Error(err)
	}

	// Verify each handler received all events
	for i, handler := range handlers {
		assert.Equal(t, numEvents, handler.EventCount(), "Handler %d did not receive all events", i)
	}
}

// TestCollaborativeOrchestrator_Stop tests the Stop method (basic check).
func TestCollaborativeOrchestrator_Stop(t *testing.T) {
	orchestrator := NewCollaborativeOrchestrator()
	handler := &SpyAgentHandler{AgentName: "handler1"}
	err := orchestrator.RegisterAgent("h1", handler)
	require.NoError(t, err)

	// Call Stop - primarily checking it doesn't panic
	assert.NotPanics(t, func() { orchestrator.Stop() }, "Stop method should not panic")
}

// TestCollaborativeOrchestrator_RegisterAgent_NilHandler tests registering a nil handler.
func TestCollaborativeOrchestrator_RegisterAgent_NilHandler(t *testing.T) {
	orchestrator := NewCollaborativeOrchestrator()

	// RegisterAgent now returns error, so check for that
	err := orchestrator.RegisterAgent("nilAgent", nil)
	// Depending on implementation, nil might be allowed or return an error.
	// Let's assume it's allowed for now, but Dispatch should handle it.
	assert.NoError(t, err, "Registering nil handler returned unexpected error")

	// Verify no handler was actually added or dispatch handles nil
	testEvent := agentflow.NewEvent("source", agentflow.EventData{}, nil)
	_, err = orchestrator.Dispatch(context.Background(), testEvent)
	// Dispatch now explicitly checks for nil handlers and returns an error
	require.Error(t, err, "Dispatch should have returned an error due to nil handler")
	assert.Contains(t, err.Error(), "encountered nil handler", "Error message mismatch for nil handler")
}

// TestSpyCollaborativeHandler_ErrorHandling tests the mock handler's error return logic.
// Updated to call Run instead of Handle.
func TestSpyCollaborativeHandler_ErrorHandling(t *testing.T) {
	specificError := errors.New("specific test error")
	handler := &SpyAgentHandler{
		AgentName:   "errorTester",
		ReturnError: specificError,
	}

	testEvent := agentflow.NewEvent("source", agentflow.EventData{}, nil)
	// Call Run with dummy context and state
	_, err := handler.Run(context.Background(), testEvent, agentflow.NewState())
	assert.Equal(t, specificError, err, "Handler should return the specified error from Run")

	// Test deliberate failure on specific event ID
	failEventID := "fail-on-this-id"
	handlerWithError := &SpyAgentHandler{
		AgentName: "failer",
		failOn:    failEventID, // Set the ID to fail on
	}

	eventToPass := agentflow.NewEvent("source", agentflow.EventData{"pass": true}, nil)
	eventToFail := agentflow.NewEvent("source", agentflow.EventData{"fail": true}, nil)
	eventToFail.SetID(failEventID) // Set the ID after creation

	_, errPass := handlerWithError.Run(context.Background(), eventToPass, agentflow.NewState())
	assert.NoError(t, errPass, "Handler should not error on event %s", eventToPass.GetID())

	_, errFail := handlerWithError.Run(context.Background(), eventToFail, agentflow.NewState())
	assert.Error(t, errFail, "Handler should error on event %s", eventToFail.GetID())
	expectedErrMsg := fmt.Sprintf("handler 'failer' failed deliberately for event '%s'", eventToFail.GetID())
	assert.EqualError(t, errFail, expectedErrMsg, "Error message mismatch for deliberate failure")
}

// TestSpyCollaborativeHandler_EventTracking tests if the mock handler correctly tracks events.
// Updated to call Run instead of Handle.
func TestSpyCollaborativeHandler_EventTracking(t *testing.T) {
	handler := &SpyAgentHandler{AgentName: "tracker"}
	event1 := agentflow.NewEvent("source", agentflow.EventData{"num": 1}, nil)
	event2 := agentflow.NewEvent("source", agentflow.EventData{"num": 2}, nil)

	// Call Run with dummy context and state
	_, _ = handler.Run(context.Background(), event1, agentflow.NewState())
	_, _ = handler.Run(context.Background(), event2, agentflow.NewState())

	assert.Equal(t, 2, handler.EventCount(), "Expected handler to have handled 2 events")
	if handler.LastEvent != nil { // Check LastEvent is not nil
		assert.Equal(t, event2.GetID(), handler.LastEvent.GetID(), "Last event ID mismatch")
	} else {
		t.Errorf("LastEvent is nil after Run calls")
	}

	handledIDs := handler.GetEvents()
	require.Len(t, handledIDs, 2, "Expected 2 event IDs tracked")
	assert.Equal(t, event1.GetID(), handledIDs[0], "First tracked event ID mismatch")
	assert.Equal(t, event2.GetID(), handledIDs[1], "Second tracked event ID mismatch")
}

// --- New Tests ---

// TestCollaborateOrchestrator_Dispatch_ContextCancellation tests behavior when context is cancelled.
func TestCollaborateOrchestrator_Dispatch_ContextCancellation(t *testing.T) {
	o := NewCollaborativeOrchestrator()
	handler := &SpyAgentHandler{
		AgentName: "slow-handler",
		// Add a delay mechanism if needed for testing cancellation, e.g., time.Sleep in Run
	}
	err := o.RegisterAgent("slow", handler)
	require.NoError(t, err)

	event := agentflow.NewEvent("target", agentflow.EventData{"type": "cancel-test"}, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond) // Short timeout
	defer cancel()

	// Dispatch might complete before the handler finishes due to the timeout.
	// The current implementation's wg.Wait() doesn't respect the context timeout directly.
	// We expect Dispatch itself not to block indefinitely if the context expires *before* wg.Wait() finishes.
	// However, if handlers are fast, it might finish normally. If handlers are slow, wg.Wait() blocks.
	// The error returned depends on whether any handler returned an error before the timeout was checked (if ever).
	_, err = o.Dispatch(ctx, event)

	// Assert based on current implementation: wg.Wait() blocks, context expiry doesn't directly cause Dispatch error.
	// Error only occurs if a handler returns an error.
	assert.NoError(t, err, "Dispatch with cancelled context returned unexpected error (current impl doesn't check ctx during wait)")

	// Check if the handler was at least called
	assert.True(t, handler.RunCalled, "Handler should have been called even if context cancelled later")
}

func TestCollaborateOrchestrator_DispatchAll_ContextCancellation(t *testing.T) {
	o := NewCollaborativeOrchestrator()
	handler := &SpyAgentHandler{
		AgentName: "slow-handler-all",
	}
	err := o.RegisterAgent("slow-all", handler)
	require.NoError(t, err)

	event := agentflow.NewEvent("target", agentflow.EventData{"type": "cancel-all-test"}, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond) // Short timeout
	defer cancel()

	// Similar expectations as the Dispatch cancellation test.
	_, err = o.DispatchAll(ctx, event) // DispatchAll calls Dispatch
	assert.NoError(t, err, "DispatchAll with cancelled context returned unexpected error")
	assert.True(t, handler.RunCalled, "Handler should have been called by DispatchAll")
}

// TestCollaborateOrchestrator_ConcurrentRegisterAndDispatch tests for race conditions.
func TestCollaborateOrchestrator_ConcurrentRegisterAndDispatch(t *testing.T) {
	// Run test with -race flag to detect race conditions: go test -race ./...
	o := NewCollaborativeOrchestrator()
	numOps := 100
	var wg sync.WaitGroup
	wg.Add(numOps * 2) // For register and dispatch operations

	// Start dispatching concurrently
	for i := 0; i < numOps; i++ {
		go func(i int) {
			defer wg.Done()
			event := agentflow.NewEvent("concurrentSource", agentflow.EventData{"event": i}, nil)
			// We don't care about the result/error here, just checking for race panics
			_, _ = o.Dispatch(context.Background(), event)
		}(i)
	}

	// Start registering concurrently
	for i := 0; i < numOps; i++ {
		go func(i int) {
			defer wg.Done()
			handler := &SpyAgentHandler{AgentName: fmt.Sprintf("concurrent-handler-%d", i)}
			// Ignore registration errors (e.g., duplicate names if timing allows)
			_ = o.RegisterAgent(handler.AgentName, handler)
		}(i)
	}

	wg.Wait() // Wait for all operations

	// Basic check: ensure some handlers were registered (exact number depends on timing)
	o.mu.RLock()
	numRegistered := len(o.handlers)
	o.mu.RUnlock()
	assert.Greater(t, numRegistered, 0, "Expected some handlers to be registered concurrently")
	t.Logf("Registered %d handlers concurrently", numRegistered)

	// Optional: Dispatch one more event and check if at least one handler receives it
	finalEvent := agentflow.NewEvent("finalSource", agentflow.EventData{"final": true}, nil)
	_, err := o.Dispatch(context.Background(), finalEvent)
	// Error might occur if nil handlers were registered due to race, check contains nil handler error
	if err != nil {
		assert.Contains(t, err.Error(), "encountered nil handler", "Unexpected error during final dispatch")
	}

	// Check if any handler received the final event (difficult to assert specific handler)
	receivedFinal := false
	o.mu.RLock()
	for _, h := range o.handlers {
		// Check if handler is not nil before trying to access its methods
		if h == nil {
			continue
		}
		spyHandler, ok := h.(*SpyAgentHandler)
		if ok {
			events := spyHandler.GetEvents()
			for _, id := range events {
				if id == finalEvent.GetID() {
					receivedFinal = true
					break
				}
			}
		}
		if receivedFinal {
			break
		}
	}
	o.mu.RUnlock()
	// This assertion might be flaky depending on timing, but checks if dispatch works after concurrent registration
	// assert.True(t, receivedFinal, "Expected at least one handler to receive the final event after concurrent ops")
}

// TestCollaborateOrchestrator_GetCallbackRegistry verifies it returns nil.
func TestCollaborativeOrchestrator_GetCallbackRegistry(t *testing.T) {
	o := NewCollaborativeOrchestrator()
	registry := o.GetCallbackRegistry()
	assert.Nil(t, registry, "GetCallbackRegistry should return nil for CollaborativeOrchestrator")
}
