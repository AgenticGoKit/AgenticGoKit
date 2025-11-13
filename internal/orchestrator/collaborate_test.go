package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/agenticgokit/agenticgokit/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollaborateOrchestrator_FanOut(t *testing.T) {
	registry := core.NewCallbackRegistry()
	o := NewCollaborativeOrchestrator(registry)
	numHandlers := 3
	handlers := make([]*SpyAgentHandler, numHandlers)
	for i := 0; i < numHandlers; i++ {
		handlers[i] = &SpyAgentHandler{AgentName: fmt.Sprintf("handler-%d", i)}
		err := o.RegisterAgent(fmt.Sprintf("handler-%d", i), handlers[i])
		require.NoError(t, err)
	}

	event := core.NewEvent("fanout", map[string]interface{}{"type": "fanout"}, nil)
	_, err := o.Dispatch(context.Background(), event)
	assert.NoError(t, err, "Dispatch failed unexpectedly")

	for i, handler := range handlers {
		count := handler.EventCount()
		if count != 1 {
			t.Errorf("Handler %d: want 1 event, got %d", i, count)
		}
		handledEvents := handler.GetEvents()
		if count > 0 && handledEvents[0] != event.GetID() {
			t.Errorf("Handler %d: want event %s, got %s", i, event.GetID(), handledEvents[0])
		}
	}
}

func TestCollaborateOrchestrator_ErrorAggregation(t *testing.T) {
	registry := core.NewCallbackRegistry()
	o := NewCollaborativeOrchestrator(registry)
	handler1 := &SpyAgentHandler{AgentName: "handler-1"}
	handler2 := &SpyAgentHandler{AgentName: "handler-2", failOn: "evt-fail"}
	handler3 := &SpyAgentHandler{AgentName: "handler-3", failOn: "evt-fail"}
	
	err := o.RegisterAgent("handler-1", handler1)
	require.NoError(t, err)
	err = o.RegisterAgent("handler-2", handler2)
	require.NoError(t, err)
	err = o.RegisterAgent("handler-3", handler3)
	require.NoError(t, err)

	event := core.NewEvent("fail", map[string]interface{}{"type": "fail"}, nil)
	event.SetID("evt-fail")
	_, aggErr := o.Dispatch(context.Background(), event)

	require.Error(t, aggErr, "Dispatch should have returned an aggregated error")

	// Check if the aggregated error message contains the individual errors
	errMsg := aggErr.Error()
	assert.Contains(t, errMsg, "failed deliberately", "Aggregated error missing failure message")

	// Check handlers were called
	assert.Equal(t, 1, handler1.EventCount(), "Handler 1 (success): unexpected event count")
	assert.Equal(t, 1, handler2.EventCount(), "Handler 2 (fail): unexpected event count")
	assert.Equal(t, 1, handler3.EventCount(), "Handler 3 (fail): unexpected event count")
	assert.True(t, handler1.RunCalled, "Handler 1 should have been called")
	assert.True(t, handler2.RunCalled, "Handler 2 should have been called")
	assert.True(t, handler3.RunCalled, "Handler 3 should have been called")
}

func TestCollaborateOrchestrator_PartialFailure(t *testing.T) {
	registry := core.NewCallbackRegistry()
	o := NewCollaborativeOrchestrator(registry)
	handlerOK := &SpyAgentHandler{AgentName: "handler-ok"}
	handlerFail := &SpyAgentHandler{AgentName: "handler-fail", failOn: "evt-partial"}
	
	err := o.RegisterAgent("handler-ok", handlerOK)
	require.NoError(t, err)
	err = o.RegisterAgent("handler-fail", handlerFail)
	require.NoError(t, err)

	event := core.NewEvent("partial", map[string]interface{}{"type": "partial"}, nil)
	event.SetID("evt-partial")
	_, aggErr := o.Dispatch(context.Background(), event)

	require.Error(t, aggErr, "Dispatch should have returned an error")
	assert.Contains(t, aggErr.Error(), "failed deliberately", "Expected error message mismatch")

	assert.Equal(t, 1, handlerOK.EventCount(), "Handler OK: unexpected event count")
	assert.Equal(t, 1, handlerFail.EventCount(), "Handler Fail: unexpected event count")
}

func TestCollaborateOrchestrator_NoHandlers(t *testing.T) {
	registry := core.NewCallbackRegistry()
	o := NewCollaborativeOrchestrator(registry)
	_, err := o.Dispatch(context.Background(), core.NewEvent("no-handler", map[string]interface{}{"type": "no-handler"}, nil))
	assert.Error(t, err, "Dispatch with no handlers should return an error")
	assert.Contains(t, err.Error(), "no agents registered", "Error should indicate no agents registered")
}

func TestCollaborateOrchestrator_ConcurrentDispatch(t *testing.T) {
	registry := core.NewCallbackRegistry()
	o := NewCollaborativeOrchestrator(registry)
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
	errChan := make(chan error, numEvents)

	for i := 0; i < numEvents; i++ {
		go func(i int) {
			defer wg.Done()
			event := core.NewEvent("concurrent", map[string]interface{}{"event": i}, nil)
			_, err := o.Dispatch(context.Background(), event)
			if err != nil {
				errChan <- fmt.Errorf("concurrent dispatch %d failed: %w", i, err)
			}
		}(i)
	}
	wg.Wait()
	close(errChan)

	// Check for errors collected from goroutines
	for err := range errChan {
		t.Error(err)
	}

	for i, handler := range handlers {
		count := handler.EventCount()
		if count != numEvents {
			t.Errorf("Handler %d: want %d events, got %d", i, numEvents, count)
		}
	}
}

func TestCollaborativeOrchestrator_Stop(t *testing.T) {
	registry := core.NewCallbackRegistry()
	orchestrator := NewCollaborativeOrchestrator(registry)
	handler := &SpyAgentHandler{AgentName: "handler1"}
	err := orchestrator.RegisterAgent("h1", handler)
	require.NoError(t, err)

	// Call Stop - primarily checking it doesn't panic
	assert.NotPanics(t, func() { orchestrator.Stop() }, "Stop method should not panic")
}

func TestCollaborativeOrchestrator_GetCallbackRegistry(t *testing.T) {
	registry := core.NewCallbackRegistry()
	o := NewCollaborativeOrchestrator(registry)
	returnedRegistry := o.GetCallbackRegistry()
	assert.Equal(t, registry, returnedRegistry, "GetCallbackRegistry should return the provided registry")
}
