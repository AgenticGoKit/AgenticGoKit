package orchestrator

import (
	"fmt"
	agentflow "kunalkushwaha/agentflow/internal/core"
	"sync"
	"testing"
)

func TestRouteOrchestrator_RoundRobin(t *testing.T) {
	o := NewRouteOrchestrator()
	numHandlers := 3
	handlers := make([]*SpyEventHandler, numHandlers) // Use type from helpers file
	for i := 0; i < numHandlers; i++ {
		handlers[i] = &SpyEventHandler{}                           // Use type from helpers file
		o.RegisterAgent(fmt.Sprintf("handler-%d", i), handlers[i]) // Pass type from helpers file
	}

	numEvents := 9
	for i := 0; i < numEvents; i++ {
		o.Dispatch(&agentflow.SimpleEvent{ID: fmt.Sprintf("evt-%d", i)})
	}

	expectedPerHandler := numEvents / numHandlers // Renamed variable
	for i, handler := range handlers {            // Use handlerUse handler
		count := handler.EventCount()
		if count != expectedPerHandler {
			t.Errorf("Handler %d: want %d events, got %d", i, expectedPerHandler, count)
		}
		firstExpected := fmt.Sprintf("evt-%d", i)
		if len(handler.events) > 0 && handler.events[0] != firstExpected {
			t.Errorf("Handler %d: first event mismatch, want %s, got %s", i, firstExpected, handler.events[0])
		}
	}
}

func TestRouteOrchestrator_HandlerFailure(t *testing.T) {
	o := NewRouteOrchestrator()
	handler1 := &SpyEventHandler{}
	handler2 := &SpyEventHandler{failOn: "evt-1"} // Handler 2 fails on the second event
	handler3 := &SpyEventHandler{}
	o.RegisterAgent("handler-1", handler1)
	o.RegisterAgent("handler-2", handler2)
	o.RegisterAgent("handler-3", handler3)

	for i := 0; i < 3; i++ {
		o.Dispatch(&agentflow.SimpleEvent{ID: fmt.Sprintf("evt-%d", i)})
	}

	// Verify handler1 got evt-0
	if count := handler1.EventCount(); count != 1 || handler1.events[0] != "evt-0" {
		t.Errorf("Handler 1: unexpected events, got %v", handler1.events)
	}
	// Verify handler2 received evt-1 (even though it failed)
	if count := handler2.EventCount(); count != 1 || handler2.events[0] != "evt-1" {
		t.Errorf("Handler 2: unexpected events, got %v", handler2.events)
	}
	// Verify handler3 got evt-2
	if count := handler3.EventCount(); count != 1 || handler3.events[0] != "evt-2" {
		t.Errorf("Handler 3: unexpected events, got %v", handler3.events)
	}
}

func TestRouteOrchestrator_NoHandlers(t *testing.T) {
	o := NewRouteOrchestrator()
	o.Dispatch(&agentflow.SimpleEvent{ID: "evt-no-handler"})
	// No assertion needed, just checking for absence of panic/deadlock
}

func TestRouteOrchestrator_ConcurrentDispatch(t *testing.T) {
	o := NewRouteOrchestrator()
	numHandlers := 5
	handlers := make([]*SpyEventHandler, numHandlers)
	for i := 0; i < numHandlers; i++ {
		handlers[i] = &SpyEventHandler{}
		o.RegisterAgent(fmt.Sprintf("handler-%d", i), handlers[i])
	}

	numEvents := 100
	var wg sync.WaitGroup
	wg.Add(numEvents)
	for i := 0; i < numEvents; i++ {
		go func(i int) {
			defer wg.Done()
			o.Dispatch(&agentflow.SimpleEvent{ID: fmt.Sprintf("evt-%d", i)})
		}(i)
	}
	wg.Wait()

	totalEventsHandled := 0
	for _, handler := range handlers {
		totalEventsHandled += handler.EventCount()
	}

	if totalEventsHandled != numEvents {
		t.Errorf("Total events handled mismatch: want %d, got %d", numEvents, totalEventsHandled)
	}
	// Note: Exact distribution isn't guaranteed with concurrent dispatch due to races
	// in the atomic increment and slice access, but total count should be correct.
}
