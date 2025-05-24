package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"testing"

	agentflow "github.com/kunalkushwaha/agentflow/internal/core"

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
	// FIX: Use SetMetadataValue and constant
	event.SetMetadata(agentflow.RouteMetadataKey, "agent2") // Use SetMetadataValue

	// FIX: Pass context and handle two return values
	_, err := orch.Dispatch(context.Background(), event) // Pass context, ignore result if only error matters
	assert.NoError(t, err, "Dispatch to agent2 failed")

	assert.False(t, handler1.RunCalled, "Handler1 should not have been called")
	assert.True(t, handler2.RunCalled, "Handler2 should have been called")
	require.Len(t, handler2.ReceivedEvents, 1, "Handler2 should have received 1 event")
	assert.Equal(t, event.GetID(), handler2.ReceivedEvents[0].GetID(), "Handler2 received wrong event")
}

func TestRouteOrchestrator_Dispatch_HandlerFailure(t *testing.T) {
	registry := agentflow.NewCallbackRegistry()
	orch := NewRouteOrchestrator(registry)
	handler1 := NewSpyRouteHandler("agent1")
	handler2 := NewSpyRouteHandler("agent2")
	handler3 := NewSpyRouteHandler("agent3")
	simulatedError := errors.New("handler 2 failed")
	handler2.ReturnError = simulatedError

	orch.RegisterAgent("agent1", handler1)
	orch.RegisterAgent("agent2", handler2)
	orch.RegisterAgent("agent3", handler3)

	evt0 := agentflow.NewEvent("type0", agentflow.EventData{"id": 0}, nil)
	// FIX: Use SetMetadataValue and constant
	evt0.SetMetadata(agentflow.RouteMetadataKey, "agent1")
	evt1 := agentflow.NewEvent("type1", agentflow.EventData{"id": 1}, nil)
	// FIX: Use SetMetadataValue and constant
	evt1.SetMetadata(agentflow.RouteMetadataKey, "agent2")
	evt2 := agentflow.NewEvent("type2", agentflow.EventData{"id": 2}, nil)
	// FIX: Use SetMetadataValue and constant
	evt2.SetMetadata(agentflow.RouteMetadataKey, "agent3")

	// Dispatch event 0 (success)
	// FIX: Pass context and handle two return values
	_, err0 := orch.Dispatch(context.Background(), evt0)
	assert.NoError(t, err0, "Dispatch for evt-0 should succeed")

	// Dispatch event 1 (failure)
	// FIX: Pass context and handle two return values
	_, err1 := orch.Dispatch(context.Background(), evt1)
	assert.Error(t, err1, "Dispatch for evt-1 should fail")
	assert.True(t, errors.Is(err1, simulatedError), "Dispatch for evt-1 returned wrong error")

	// Dispatch event 2 (success)
	// FIX: Pass context and handle two return values
	_, err2 := orch.Dispatch(context.Background(), evt2)
	assert.NoError(t, err2, "Dispatch for evt-2 should succeed")

	// Verify handler1 got evt-0
	assert.True(t, handler1.RunCalled, "Handler1 should have run")
	require.Len(t, handler1.ReceivedEvents, 1)
	assert.Equal(t, evt0.GetID(), handler1.ReceivedEvents[0].GetID())

	// Verify handler2 received evt-1 (even though it failed)
	assert.True(t, handler2.RunCalled, "Handler2 should have run")
	require.Len(t, handler2.ReceivedEvents, 1)
	assert.Equal(t, evt1.GetID(), handler2.ReceivedEvents[0].GetID())

	// Verify handler3 got evt-2
	assert.True(t, handler3.RunCalled, "Handler3 should have run")
	require.Len(t, handler3.ReceivedEvents, 1)
	assert.Equal(t, evt2.GetID(), handler3.ReceivedEvents[0].GetID())
}

func TestRouteOrchestrator_Dispatch_NoTarget(t *testing.T) {
	registry := agentflow.NewCallbackRegistry()
	o := NewRouteOrchestrator(registry)
	handler1 := NewSpyRouteHandler("agent1")
	o.RegisterAgent("agent1", handler1)

	// Create event with NO RouteMetadataKey
	event := agentflow.NewEvent("no-target-type", agentflow.EventData{}, nil)

	// FIX: Pass context and handle two return values
	_, err := o.Dispatch(context.Background(), event)
	assert.Error(t, err, "Dispatch with no target should return an error")
	// FIX: Check for specific error message if RouteOrchestrator returns one
	assert.Contains(t, err.Error(), "missing routing key", "Error message should indicate missing target")
	assert.False(t, handler1.RunCalled, "Handler should not have been called")
}

func TestRouteOrchestrator_Dispatch_TargetNotFound(t *testing.T) {
	registry := agentflow.NewCallbackRegistry()
	o := NewRouteOrchestrator(registry)
	handler1 := NewSpyRouteHandler("agent1")
	o.RegisterAgent("agent1", handler1)

	event := agentflow.NewEvent("wrong-target-type", agentflow.EventData{}, nil)
	// FIX: Use SetMetadataValue and constant
	event.SetMetadata(agentflow.RouteMetadataKey, "nonexistent-agent")

	// FIX: Pass context and handle two return values
	_, err := o.Dispatch(context.Background(), event)
	assert.Error(t, err, "Dispatch to non-existent target should return an error")
	// FIX: Check for specific error message - update expected substring
	assert.Contains(t, err.Error(), "no agent handler registered for target", "Error message should indicate target not found")
	assert.False(t, handler1.RunCalled, "Handler should not have been called")
}

func TestRouteOrchestrator_Dispatch_NilEvent(t *testing.T) {
	registry := agentflow.NewCallbackRegistry()
	o := NewRouteOrchestrator(registry)
	// FIX: Pass context and handle two return values
	_, err := o.Dispatch(context.Background(), nil) // Pass context
	assert.Error(t, err, "Dispatching nil event should return an error")
}

func TestRouteOrchestrator_NoHandlers(t *testing.T) {
	registry := agentflow.NewCallbackRegistry()
	o := NewRouteOrchestrator(registry) // No handlers registered
	event := agentflow.NewEvent("test-type", agentflow.EventData{}, nil)
	// FIX: Use SetMetadataValue and constant
	event.SetMetadata(agentflow.RouteMetadataKey, "some-agent")

	// FIX: Pass context and handle two return values
	_, err := o.Dispatch(context.Background(), event) // Pass context
	assert.Error(t, err, "Dispatch with no handlers registered should return an error")
	// FIX: Update expected error substring to match actual message
	assert.Contains(t, err.Error(), "no agent handler registered for target", "Error message should indicate target not found")
}

func TestRouteOrchestrator_ConcurrentDispatch(t *testing.T) {
	registry := agentflow.NewCallbackRegistry()
	orch := NewRouteOrchestrator(registry)
	numHandlers := 5
	numEvents := 100
	handlers := make([]*SpyRouteHandler, numHandlers)
	for i := 0; i < numHandlers; i++ {
		id := fmt.Sprintf("agent%d", i)
		handlers[i] = NewSpyRouteHandler(id)
		orch.RegisterAgent(id, handlers[i])
	}

	var wg sync.WaitGroup
	wg.Add(numEvents)

	for i := 0; i < numEvents; i++ {
		go func(eventIndex int) {
			defer wg.Done()
			targetAgent := fmt.Sprintf("agent%d", eventIndex%numHandlers) // Distribute events
			event := agentflow.NewEvent("concurrent-type", agentflow.EventData{"index": eventIndex}, nil)
			// FIX: Use SetMetadataValue and constant
			event.SetMetadata(agentflow.RouteMetadataKey, targetAgent)
			// FIX: Pass context and handle two return values
			_, err := orch.Dispatch(context.Background(), event) // Pass context
			// In a real test, might collect errors in a channel
			if err != nil {
				t.Errorf("Concurrent dispatch %d failed: %v", eventIndex, err)
			}
		}(i)
	}

	wg.Wait()

	totalReceived := 0
	for _, h := range handlers {
		h.mu.Lock() // Lock spy handler to read count safely
		count := len(h.ReceivedEvents)
		h.mu.Unlock()
		totalReceived += count
		assert.Greater(t, count, 0, "Handler %s should have received some events", h.ID)
		// Exact distribution isn't guaranteed, but check it's roughly balanced if needed
	}
	assert.Equal(t, numEvents, totalReceived, "Total received events should match total dispatched events")
}
