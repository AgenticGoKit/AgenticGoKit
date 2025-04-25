package orchestrator

import (
	"context"
	"errors"
	"fmt"
	agentflow "kunalkushwaha/agentflow/internal/core"
	"log"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SpyRouteHandler is a mock handler for testing RouteOrchestrator
type SpyRouteHandler struct {
	ID               string
	ReceivedEvents   []agentflow.Event
	ReceivedStates   []agentflow.State
	ReceivedContexts []context.Context
	ReturnError      error
	mu               sync.Mutex
	RunCalled        bool
}

// NewSpyRouteHandler creates a new spy handler.
func NewSpyRouteHandler(id string) *SpyRouteHandler {
	return &SpyRouteHandler{
		ID:             id,
		ReceivedEvents: make([]agentflow.Event, 0),
		ReceivedStates: make([]agentflow.State, 0),
	}
}

// Run records the call and returns a predefined error or nil.
func (h *SpyRouteHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.RunCalled = true
	h.ReceivedEvents = append(h.ReceivedEvents, event)
	h.ReceivedStates = append(h.ReceivedStates, state)
	h.ReceivedContexts = append(h.ReceivedContexts, ctx)
	log.Printf("SpyRouteHandler '%s' received event %s", h.ID, event.GetID())
	return agentflow.AgentResult{}, h.ReturnError
}

// --- Tests using SpyRouteHandler ---

func TestRouteOrchestrator_RegisterAgent(t *testing.T) {
	o := NewRouteOrchestrator(agentflow.NewCallbackRegistry())
	handler := NewSpyRouteHandler("agent1")

	o.RegisterAgent("agent1", handler)

	registeredHandler, exists := o.handlers["agent1"]
	require.True(t, exists, "Agent 'agent1' was not registered")
	assert.Equal(t, handler, registeredHandler, "Registered handler mismatch for 'agent1'")

	// Test duplicate registration
	o.RegisterAgent("agent1", handler)
	assert.Len(t, o.handlers, 1, "Duplicate registration should not add handler again")

	// Test nil handler registration
	o.RegisterAgent("agent2", nil)
	_, exists = o.handlers["agent2"]
	assert.False(t, exists, "Nil handler should not be registered")
}

func TestRouteOrchestrator_Dispatch_Routing(t *testing.T) {
	registry := agentflow.NewCallbackRegistry()
	orch := NewRouteOrchestrator(registry) // pass registry here

	handler1 := NewSpyRouteHandler("agent1")
	handler2 := NewSpyRouteHandler("agent2")

	orch.RegisterAgent("agent1", handler1)
	orch.RegisterAgent("agent2", handler2)

	event := agentflow.NewEvent("test-event-type", agentflow.EventData{}, nil)
	event.SetTargetAgentID("agent2") // explicitly set target agent ID

	err := orch.Dispatch(event)
	assert.NoError(t, err, "Dispatch to agent2 failed")

	assert.Len(t, handler2.ReceivedEvents, 1, "Handler2 should have received 1 event")
	assert.Len(t, handler1.ReceivedEvents, 0, "Handler1 should not have received any events")
}

func TestRouteOrchestrator_Dispatch_HandlerFailure(t *testing.T) {
	registry := agentflow.NewCallbackRegistry()
	o := NewRouteOrchestrator(registry)
	simulatedError := errors.New("handler2 failed deliberately")

	handler1 := NewSpyRouteHandler("agent1")
	handler2 := NewSpyRouteHandler("agent2")
	handler2.ReturnError = simulatedError
	handler3 := NewSpyRouteHandler("agent3")

	o.RegisterAgent("agent1", handler1)
	o.RegisterAgent("agent2", handler2)
	o.RegisterAgent("agent3", handler3)

	event0 := agentflow.NewEvent("test-type", nil, nil)
	event0.SetID("evt-0")
	event0.SetTargetAgentID("agent1")

	event1 := agentflow.NewEvent("test-type", nil, nil)
	event1.SetID("evt-1")
	event1.SetTargetAgentID("agent2")

	event2 := agentflow.NewEvent("test-type", nil, nil)
	event2.SetID("evt-2")
	event2.SetTargetAgentID("agent3")

	// Dispatch event 0 (success)
	err0 := o.Dispatch(event0)
	assert.NoError(t, err0, "Dispatch failed for event 0")

	// Dispatch event 1 (failure)
	err1 := o.Dispatch(event1)
	assert.Error(t, err1, "Dispatch should have failed for event 1")
	assert.ErrorIs(t, err1, simulatedError, "Dispatch error mismatch for event 1")
	assert.Contains(t, err1.Error(), "handler 'agent2' failed", "Error message should indicate which handler failed")

	// Dispatch event 2 (success)
	err2 := o.Dispatch(event2)
	assert.NoError(t, err2, "Dispatch failed for event 2")

	// Verify handler1 got evt-0
	assert.Len(t, handler1.ReceivedEvents, 1, "Handler 1 event count mismatch")
	assert.Equal(t, "evt-0", handler1.ReceivedEvents[0].GetID(), "Handler 1 event ID mismatch")

	// Verify handler2 received evt-1 (even though it failed)
	assert.Len(t, handler2.ReceivedEvents, 1, "Handler 2 event count mismatch")
	assert.Equal(t, "evt-1", handler2.ReceivedEvents[0].GetID(), "Handler 2 event ID mismatch")

	// Verify handler3 got evt-2
	assert.Len(t, handler3.ReceivedEvents, 1, "Handler 3 event count mismatch")
	assert.Equal(t, "evt-2", handler3.ReceivedEvents[0].GetID(), "Handler 3 event ID mismatch")
}

func TestRouteOrchestrator_Dispatch_NoTarget(t *testing.T) {
	registry := agentflow.NewCallbackRegistry()
	orch := NewRouteOrchestrator(registry)
	handler1 := NewSpyRouteHandler("agent1")
	orch.RegisterAgent("agent1", handler1)

	// Create event with NO TargetAgentID (pass "") and nil metadata.
	// If event type is important, consider adding it to metadata or a dedicated field.
	event := agentflow.NewEvent("", agentflow.EventData{}, nil)
	// Optionally, set a type if your Event interface supports it or use metadata:
	// event.SetMetadata("type", "test-type")

	err := orch.Dispatch(event)
	assert.Error(t, err)
	// Now this assertion should pass
	assert.Contains(t, err.Error(), "has no target agent ID or route metadata key", "Error message mismatch for no target")
}

func TestRouteOrchestrator_Dispatch_NilEvent(t *testing.T) {
	registry := agentflow.NewCallbackRegistry()
	o := NewRouteOrchestrator(registry)
	err := o.Dispatch(nil)
	require.Error(t, err, "Dispatch did not return an error for nil event")
}

func TestRouteOrchestrator_NoHandlers(t *testing.T) {
	o := NewRouteOrchestrator(agentflow.NewCallbackRegistry())
	event := agentflow.NewEvent("agent1", nil, nil)
	err := o.Dispatch(event)
	require.Error(t, err, "Dispatch should fail if no handler is registered for the target")
}

func TestRouteOrchestrator_ConcurrentDispatch(t *testing.T) {
	registry := agentflow.NewCallbackRegistry()
	o := NewRouteOrchestrator(registry)
	numHandlers := 5
	handlers := make([]*SpyRouteHandler, numHandlers)
	for i := 0; i < numHandlers; i++ {
		agentID := fmt.Sprintf("handler-%d", i)
		handlers[i] = NewSpyRouteHandler(agentID)
		o.RegisterAgent(agentID, handlers[i])
	}

	numEvents := 100
	var wg sync.WaitGroup
	wg.Add(numEvents)

	// Distribute events somewhat evenly for testing routing
	for i := 0; i < numEvents; i++ {
		go func(i int) {
			defer wg.Done()
			targetAgent := fmt.Sprintf("handler-%d", i%numHandlers)
			event := agentflow.NewEvent("concurrent-test", agentflow.EventData{"eventNum": i}, nil)
			event.SetID(fmt.Sprintf("evt-%d", i))
			event.SetTargetAgentID(targetAgent)

			err := o.Dispatch(event)
			if err != nil {
				t.Logf("Concurrent dispatch for event %s failed: %v", event.GetID(), err)
			}
		}(i)
	}
	wg.Wait()

	totalEventsHandled := 0
	expectedPerHandler := numEvents / numHandlers
	remainder := numEvents % numHandlers

	for i, handler := range handlers {
		count := len(handler.ReceivedEvents)
		expectedCount := expectedPerHandler
		if i < remainder {
			expectedCount++
		}
		assert.Equal(t, expectedCount, count, "Handler %d event count mismatch", i)
		totalEventsHandled += count
	}
	assert.Equal(t, numEvents, totalEventsHandled, "Total handled events mismatch")
}
