package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/kunalkushwaha/agenticgokit/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SpyRouteHandler is a mock handler for testing RouteOrchestrator
type SpyRouteHandler struct {
	ID               string
	ReceivedEvents   []core.Event
	ReceivedStates   []core.State
	ReceivedContexts []context.Context
	ReturnError      error
	mu               sync.Mutex
	RunCalled        bool
}

// NewSpyRouteHandler creates a new spy handler.
func NewSpyRouteHandler(id string) *SpyRouteHandler {
	return &SpyRouteHandler{
		ID:             id,
		ReceivedEvents: make([]core.Event, 0),
		ReceivedStates: make([]core.State, 0),
	}
}

// Run records the call and returns a predefined error or nil.
func (h *SpyRouteHandler) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.RunCalled = true
	h.ReceivedEvents = append(h.ReceivedEvents, event)
	h.ReceivedStates = append(h.ReceivedStates, state)
	h.ReceivedContexts = append(h.ReceivedContexts, ctx)
	log.Printf("SpyRouteHandler '%s' received event %s", h.ID, event.GetID())
	return core.AgentResult{OutputState: state}, h.ReturnError
}

// --- Tests using SpyRouteHandler ---

func TestRouteOrchestrator_RegisterAgent(t *testing.T) {
	o := NewRouteOrchestrator(core.NewCallbackRegistry())
	handler := NewSpyRouteHandler("agent1")

	err := o.RegisterAgent("agent1", handler)
	require.NoError(t, err)

	registeredHandler, exists := o.handlers["agent1"]
	require.True(t, exists, "Agent 'agent1' was not registered")
	assert.Equal(t, handler, registeredHandler, "Registered handler mismatch for 'agent1'")

	// Test duplicate registration
	err = o.RegisterAgent("agent1", handler)
	require.NoError(t, err) // Should not error, just overwrite
	assert.Len(t, o.handlers, 1, "Duplicate registration should overwrite")

	// Test nil handler registration
	err = o.RegisterAgent("agent2", nil)
	assert.Error(t, err, "Nil handler should return error")
}

func TestRouteOrchestrator_Dispatch_Routing(t *testing.T) {
	registry := core.NewCallbackRegistry()
	orch := NewRouteOrchestrator(registry)

	handler1 := NewSpyRouteHandler("agent1")
	handler2 := NewSpyRouteHandler("agent2")

	orch.RegisterAgent("agent1", handler1)
	orch.RegisterAgent("agent2", handler2)

	event := core.NewEvent("test-event-type", map[string]interface{}{}, map[string]string{
		core.RouteMetadataKey: "agent2",
	})

	result, err := orch.Dispatch(context.Background(), event)
	assert.NoError(t, err, "Dispatch to agent2 failed")
	assert.NotNil(t, result.OutputState, "Result should have output state")

	assert.False(t, handler1.RunCalled, "Handler1 should not have been called")
	assert.True(t, handler2.RunCalled, "Handler2 should have been called")
	require.Len(t, handler2.ReceivedEvents, 1, "Handler2 should have received 1 event")
	assert.Equal(t, event.GetID(), handler2.ReceivedEvents[0].GetID(), "Handler2 received wrong event")
}

func TestRouteOrchestrator_Dispatch_HandlerFailure(t *testing.T) {
	registry := core.NewCallbackRegistry()
	orch := NewRouteOrchestrator(registry)
	handler1 := NewSpyRouteHandler("agent1")
	handler2 := NewSpyRouteHandler("agent2")
	handler3 := NewSpyRouteHandler("agent3")
	simulatedError := errors.New("handler 2 failed")
	handler2.ReturnError = simulatedError

	orch.RegisterAgent("agent1", handler1)
	orch.RegisterAgent("agent2", handler2)
	orch.RegisterAgent("agent3", handler3)

	evt0 := core.NewEvent("type0", map[string]interface{}{"id": 0}, map[string]string{
		core.RouteMetadataKey: "agent1",
	})
	evt1 := core.NewEvent("type1", map[string]interface{}{"id": 1}, map[string]string{
		core.RouteMetadataKey: "agent2",
	})
	evt2 := core.NewEvent("type2", map[string]interface{}{"id": 2}, map[string]string{
		core.RouteMetadataKey: "agent3",
	})

	// Dispatch event 0 (success)
	_, err0 := orch.Dispatch(context.Background(), evt0)
	assert.NoError(t, err0, "Dispatch for evt-0 should succeed")

	// Dispatch event 1 (failure)
	_, err1 := orch.Dispatch(context.Background(), evt1)
	assert.Error(t, err1, "Dispatch for evt-1 should fail")
	assert.True(t, errors.Is(err1, simulatedError), "Dispatch for evt-1 returned wrong error")

	// Dispatch event 2 (success)
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
	registry := core.NewCallbackRegistry()
	o := NewRouteOrchestrator(registry)
	handler1 := NewSpyRouteHandler("agent1")
	o.RegisterAgent("agent1", handler1)

	// Create event with NO RouteMetadataKey
	event := core.NewEvent("no-target-type", map[string]interface{}{}, nil)

	_, err := o.Dispatch(context.Background(), event)
	assert.Error(t, err, "Dispatch with no target should return an error")
	assert.Contains(t, err.Error(), "missing routing key", "Error message should indicate missing target")
	assert.False(t, handler1.RunCalled, "Handler should not have been called")
}

func TestRouteOrchestrator_Dispatch_TargetNotFound(t *testing.T) {
	registry := core.NewCallbackRegistry()
	o := NewRouteOrchestrator(registry)
	handler1 := NewSpyRouteHandler("agent1")
	o.RegisterAgent("agent1", handler1)

	event := core.NewEvent("wrong-target-type", map[string]interface{}{}, map[string]string{
		core.RouteMetadataKey: "nonexistent-agent",
	})

	_, err := o.Dispatch(context.Background(), event)
	assert.Error(t, err, "Dispatch to non-existent target should return an error")
	assert.Contains(t, err.Error(), "no agent handler registered for target", "Error message should indicate target not found")
	assert.False(t, handler1.RunCalled, "Handler should not have been called")
}

func TestRouteOrchestrator_Dispatch_NilEvent(t *testing.T) {
	registry := core.NewCallbackRegistry()
	o := NewRouteOrchestrator(registry)
	_, err := o.Dispatch(context.Background(), nil)
	assert.Error(t, err, "Dispatching nil event should return an error")
}

func TestRouteOrchestrator_NoHandlers(t *testing.T) {
	registry := core.NewCallbackRegistry()
	o := NewRouteOrchestrator(registry) // No handlers registered
	event := core.NewEvent("test-type", map[string]interface{}{}, map[string]string{
		core.RouteMetadataKey: "some-agent",
	})

	_, err := o.Dispatch(context.Background(), event)
	assert.Error(t, err, "Dispatch with no handlers registered should return an error")
	assert.Contains(t, err.Error(), "no agent handler registered for target", "Error message should indicate target not found")
}

func TestRouteOrchestrator_ConcurrentDispatch(t *testing.T) {
	registry := core.NewCallbackRegistry()
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
			event := core.NewEvent("concurrent-type", map[string]interface{}{"index": eventIndex}, map[string]string{
				core.RouteMetadataKey: targetAgent,
			})
			_, err := orch.Dispatch(context.Background(), event)
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
	}
	assert.Equal(t, numEvents, totalReceived, "Total received events should match total dispatched events")
}