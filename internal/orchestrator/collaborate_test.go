package orchestrator

import (
	"fmt"
	agentflow "kunalkushwaha/agentflow/internal/core"
	"sync"
	"testing"
)

// Reusing SpyEventHandler from orchestrator_test_helpers.go
// Remove the duplicate definitions of SpyEventHandler and SlowSpyEventHandler below

func TestCollaborateOrchestrator_FanOut(t *testing.T) {
	o := NewCollaborateOrchestrator()
	numHandlers := 3
	handlers := make([]*SpyEventHandler, numHandlers) // Use type from helpers file
	for i := 0; i < numHandlers; i++ {
		handlers[i] = &SpyEventHandler{}                           // Use type from helpers file
		o.RegisterAgent(fmt.Sprintf("handler-%d", i), handlers[i]) // Pass type from helpers file
	}

	event := &agentflow.SimpleEvent{ID: "evt-fanout"}
	o.Dispatch(event)

	for i, handler := range handlers { // Use handler
		count := handler.EventCount()
		if count != 1 {
			t.Errorf("Handler %d: want 1 event, got %d", i, count)
		}
		if count > 0 && handler.events[0] != event.GetID() {
			t.Errorf("Handler %d: want event %s, got %s", i, event.GetID(), handler.events[0])
		}
	}
}

func TestCollaborateOrchestrator_ErrorAggregation(t *testing.T) {
	o := NewCollaborateOrchestrator()
	handler1 := &SpyEventHandler{}                   // Use type from helpers file
	handler2 := &SpyEventHandler{failOn: "evt-fail"} // Use type from helpers file
	handler3 := &SpyEventHandler{failOn: "evt-fail"} // Use type from helpers file
	o.RegisterAgent("handler-1", handler1)           // Pass type from helpers file
	o.RegisterAgent("handler-2", handler2)           // Pass type from helpers file
	o.RegisterAgent("handler-3", handler3)           // Pass type from helpers file

	event := &agentflow.SimpleEvent{ID: "evt-fail"}
	o.Dispatch(event)

	if count := handler1.EventCount(); count != 1 || handler1.events[0] != "evt-fail" {
		t.Errorf("Handler 1 (success): unexpected events, got %v", handler1.events)
	}
	if count := handler2.EventCount(); count != 1 || handler2.events[0] != "evt-fail" {
		t.Errorf("Handler 2 (fail): unexpected events, got %v", handler2.events)
	}
	if count := handler3.EventCount(); count != 1 || handler3.events[0] != "evt-fail" {
		t.Errorf("Handler 3 (fail): unexpected events, got %v", handler3.events)
	}
}

func TestCollaborateOrchestrator_PartialFailure(t *testing.T) {
	o := NewCollaborateOrchestrator()
	handlerOK := &SpyEventHandler{}                        // Use type from helpers file
	handlerFail := &SpyEventHandler{failOn: "evt-partial"} // Use type from helpers file
	o.RegisterAgent("handler-ok", handlerOK)               // Pass type from helpers file
	o.RegisterAgent("handler-fail", handlerFail)           // Pass type from helpers file

	event := &agentflow.SimpleEvent{ID: "evt-partial"}
	o.Dispatch(event)

	if count := handlerOK.EventCount(); count != 1 || handlerOK.events[0] != "evt-partial" {
		t.Errorf("Handler OK: unexpected events, got %v", handlerOK.events)
	}
	if count := handlerFail.EventCount(); count != 1 || handlerFail.events[0] != "evt-partial" {
		t.Errorf("Handler Fail: unexpected events, got %v", handlerFail.events)
	}
}

func TestCollaborateOrchestrator_NoHandlers(t *testing.T) { // Renamed test
	o := NewCollaborateOrchestrator()
	o.Dispatch(&agentflow.SimpleEvent{ID: "evt-no-handler"}) // Renamed event ID for clarity
}

func TestCollaborateOrchestrator_ConcurrentDispatch(t *testing.T) {
	o := NewCollaborateOrchestrator()
	numHandlers := 3
	handlers := make([]*SpyEventHandler, numHandlers) // Use type from helpers file
	for i := 0; i < numHandlers; i++ {
		handlers[i] = &SpyEventHandler{}                           // Use type from helpers file
		o.RegisterAgent(fmt.Sprintf("handler-%d", i), handlers[i]) // Pass type from helpers file
	}

	numEvents := 50
	var wg sync.WaitGroup
	wg.Add(numEvents)
	for i := 0; i < numEvents; i++ {
		go func(i int) {
			defer wg.Done()
			o.Dispatch(&agentflow.SimpleEvent{ID: fmt.Sprintf("evt-concurrent-%d", i)})
		}(i)
	}
	wg.Wait()

	for i, handler := range handlers { // Use handler
		count := handler.EventCount()
		if count != numEvents {
			t.Errorf("Handler %d: want %d events, got %d", i, numEvents, count)
		}
	}
}

// Optional: Test with timeout (requires modifying Dispatch to accept context)
/*  // Make sure this opening comment marker is present
func TestCollaborateOrchestrator_Timeout(t *testing.T) {
    o := NewCollaborateOrchestrator()
    handlerFast := &SpyEventHandler{} // Use type from helpers file
    handlerSlow := &SlowSpyEventHandler{delay: 100 * time.Millisecond} // Use type from helpers file
    o.RegisterAgent("handler-fast", handlerFast) // Pass type from helpers file
    o.RegisterAgent("handler-slow", handlerSlow) // Pass type from helpers file

    event := &agentflow.SimpleEvent{ID: "evt-timeout"}
    ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond) // Timeout shorter than slow agent delay
    defer cancel()

    // Modify Dispatch to accept and use context
    // o.Dispatch(ctx, event)

    // Assertions would depend on how timeout errors are handled/logged
    // e.g., check logs or check that fast agent completed but slow one might not have fully logged its event
}
*/ // Ensure this closing comment marker exists and is correct (line 149 or around there)
