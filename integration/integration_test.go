package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	agenticgokit "github.com/kunalkushwaha/AgenticGoKit/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicEndToEndFlow(t *testing.T) {
	runner := agentflow.NewRunner(1)
	callbackRegistry := agentflow.NewCallbackRegistry()
	var afterEventState agentflow.State
	var once sync.Once
	done := make(chan struct{})

	// Ensure runner uses our callback registry
	runner.SetCallbackRegistry(callbackRegistry)
	// Create route orchestrator with the callback registry
	orch := agentflow.NewRouteOrchestrator(callbackRegistry)
	runner.SetOrchestrator(orch)

	// Add a route terminator callback to prevent infinite loops
	callbackRegistry.Register(agentflow.HookAfterAgentRun, "routeTerminator",
		func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
			if state, ok := args.State.Get("processCount"); ok {
				count := state.(int)
				if count >= 3 {
					args.State.SetMeta(agentflow.RouteMetadataKey, "")
					agentflow.Logger().Info().
						Int("count", count).
						Msg("Route terminator: Breaking cycle after max iterations")
				} else {
					args.State.Set("processCount", count+1)
				}
			} else {
				args.State.Set("processCount", 1)
			}
			// Preserve critical test data in state
			if args.Event != nil && args.Event.GetData() != nil {
				if message, ok := args.Event.GetData()["message"]; ok {
					args.State.Set("original_message", message)
					agentflow.Logger().Info().
						Str("message", fmt.Sprintf("%v", message)).
						Msg("Route terminator: Preserved original message")
				}
			}
			return args.State, nil
		})

	// Add debug callback to monitor events and state
	callbackRegistry.Register(agentflow.HookBeforeEventHandling, "debugRouter",
		func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
			if args.Event != nil {
				agentflow.Logger().Debug().
					Str("event_id", args.Event.GetID()).
					Interface("data_keys", args.Event.GetData()).
					Msg("DEBUG ROUTER: Processing event")
			}
			return args.State, nil
		})

	// Capture the final state after all processing
	callbackRegistry.Register(agentflow.HookAfterEventHandling, "captureFinalState",
		func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
			afterEventState = args.State.Clone()
			agentflow.Logger().Debug().Msg("AfterEventHandling callback: captured state and closing done channel")
			once.Do(func() { close(done) })
			return args.State, nil
		})

	// Create a test agent that properly preserves original data
	var processedEvent agentflow.Event

	testAgent := &TestAgent{
		ProcessFn: func(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
			if processedEvent == nil {
				processedEvent = event
				agentflow.Logger().Info().
					Str("event_id", event.GetID()).
					Msg("Test agent: Captured original event")
			}

			outputState := state.Clone()
			if message, ok := event.GetData()["message"]; ok {
				outputState.Set("original_message", message)
				agentflow.Logger().Info().
					Str("message", fmt.Sprintf("%v", message)).
					Msg("Test agent: Setting original message from event")
			} else if origMessage, ok := state.Get("original_message"); ok {
				outputState.Set("original_message", origMessage)
				agentflow.Logger().Info().
					Str("message", fmt.Sprintf("%v", origMessage)).
					Msg("Test agent: Maintaining original message from state")
			}

			outputState.Set("response", "processed")
			count := 1
			if countVal, ok := state.Get("processCount"); ok {
				count = countVal.(int)
			}
			outputState.Set("processCount", count)

			return agentflow.AgentResult{
				OutputState: outputState,
			}, nil
		},
	}

	// Register the agent with the orchestrator
	err := runner.RegisterAgent("test-agent", testAgent)
	require.NoError(t, err)

	// Start the runner
	ctx := context.Background()
	err = runner.Start(ctx)
	require.NoError(t, err)
	defer runner.Stop()

	// Create and emit an event
	eventData := agentflow.EventData{
		"message": "test-event",
	}
	eventMeta := map[string]string{
		agentflow.RouteMetadataKey: "test-agent",
		agentflow.SessionIDKey:     "test-session",
	}
	event := agentflow.NewEvent("test-source", eventData, eventMeta)
	err = runner.Emit(event)
	require.NoError(t, err)

	// Wait for the AfterEventHandling callback to fire
	select {
	case <-done:
		// Callback fired
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for AfterEventHandling callback")
	}

	// Verify the event was processed
	assert.NotNil(t, processedEvent)

	// Check data in the final state which should have preserved messages
	require.NotNil(t, afterEventState, "AfterEventHandling state should not be nil")
	originalMessage, ok := afterEventState.Get("original_message")
	assert.True(t, ok, "Original message should be preserved in state")
	assert.Equal(t, "test-event", originalMessage)
}

func TestMultipleAgentsSequentialRouting(t *testing.T) {
	runner := agentflow.NewRunner(1)
	callbackRegistry := agentflow.NewCallbackRegistry()
	runner.SetCallbackRegistry(callbackRegistry)
	orch := agentflow.NewRouteOrchestrator(callbackRegistry)
	runner.SetOrchestrator(orch)

	var finalState agentflow.State
	var once sync.Once
	done := make(chan struct{})
	callbackRegistry.Register(agentflow.HookAfterEventHandling, "captureFinalState",
		func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
			step, _ := args.State.Get("step")
			if step == "B" {
				finalState = args.State.Clone()
				once.Do(func() { close(done) })
			}
			return args.State, nil
		})

	agentA := &TestAgent{
		ProcessFn: func(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
			out := state.Clone()
			out.Set("step", "A")
			out.SetMeta(agentflow.RouteMetadataKey, "agent-b")
			return agentflow.AgentResult{OutputState: out}, nil
		},
	}
	agentB := &TestAgent{
		ProcessFn: func(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
			out := state.Clone()
			out.Set("step", "B")
			out.SetMeta(agentflow.RouteMetadataKey, "") // End routing
			return agentflow.AgentResult{OutputState: out}, nil
		},
	}
	require.NoError(t, runner.RegisterAgent("agent-a", agentA))
	require.NoError(t, runner.RegisterAgent("agent-b", agentB))

	require.NoError(t, runner.Start(context.Background()))
	defer runner.Stop()

	event := agentflow.NewEvent("test", agentflow.EventData{}, map[string]string{
		agentflow.RouteMetadataKey: "agent-a",
		agentflow.SessionIDKey:     "multi-seq-session",
	})
	require.NoError(t, runner.Emit(event))

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for AfterEventHandling callback")
	}

	require.NotNil(t, finalState)
	step, ok := finalState.Get("step")
	assert.True(t, ok)
	assert.Equal(t, "B", step)
}

func TestAgentErrorTriggersErrorHandler(t *testing.T) {
	runner := agentflow.NewRunner(1)
	callbackRegistry := agentflow.NewCallbackRegistry()
	runner.SetCallbackRegistry(callbackRegistry)
	orch := agentflow.NewRouteOrchestrator(callbackRegistry)
	runner.SetOrchestrator(orch)

	var errorHandlerCalled bool
	done := make(chan struct{})
	errorHandler := &TestAgent{
		ProcessFn: func(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
			errorHandlerCalled = true
			close(done)
			return agentflow.AgentResult{OutputState: state}, nil
		},
	}
	require.NoError(t, runner.RegisterAgent("error-handler", errorHandler))

	badAgent := &TestAgent{
		ProcessFn: func(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
			return agentflow.AgentResult{}, fmt.Errorf("simulated failure")
		},
	}
	require.NoError(t, runner.RegisterAgent("bad-agent", badAgent))

	require.NoError(t, runner.Start(context.Background()))
	defer runner.Stop()

	event := agentflow.NewEvent("test", agentflow.EventData{}, map[string]string{
		agentflow.RouteMetadataKey: "bad-agent",
		agentflow.SessionIDKey:     "error-session",
	})
	require.NoError(t, runner.Emit(event))

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for error-handler")
	}

	assert.True(t, errorHandlerCalled, "Error handler should be called")
}

func TestSessionStatePersistence(t *testing.T) {
	runner := agentflow.NewRunner(1)
	callbackRegistry := agentflow.NewCallbackRegistry()
	runner.SetCallbackRegistry(callbackRegistry)
	orch := agentflow.NewRouteOrchestrator(callbackRegistry)
	runner.SetOrchestrator(orch)

	var lastState agentflow.State
	var once sync.Once
	done := make(chan struct{})
	callbackRegistry.Register(agentflow.HookAfterEventHandling, "captureFinalState",
		func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
			lastState = args.State.Clone()
			once.Do(func() { close(done) })
			return args.State, nil
		})

	counterAgent := &TestAgent{
		ProcessFn: func(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
			out := state.Clone()
			count := 0
			if v, ok := state.Get("count"); ok {
				count = v.(int)
			}
			out.Set("count", count+1)
			out.SetMeta(agentflow.RouteMetadataKey, "") // End routing
			return agentflow.AgentResult{OutputState: out}, nil
		},
	}
	require.NoError(t, runner.RegisterAgent("counter", counterAgent))

	require.NoError(t, runner.Start(context.Background()))
	defer runner.Stop()

	sessionID := "session-persist"
	state := agentflow.NewState()
	for i := 0; i < 3; i++ {
		done = make(chan struct{})
		once = sync.Once{}
		event := agentflow.NewEvent("test", agentflow.EventData{}, map[string]string{
			agentflow.RouteMetadataKey: "counter",
			agentflow.SessionIDKey:     sessionID,
		})
		// Set the current state as event data for the next event
		for _, key := range state.Keys() {
			val, _ := state.Get(key)
			event.SetData(key, val)
		}
		require.NoError(t, runner.Emit(event))
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for AfterEventHandling callback")
		}
		// Update state for next iteration
		state = lastState.Clone().(*agentflow.SimpleState)
	}

	require.NotNil(t, lastState)
	count, ok := lastState.Get("count")
	assert.True(t, ok)
	assert.Equal(t, 3, count)
}

// TestAgent is a test implementation of AgentHandler
type TestAgent struct {
	ProcessFn func(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error)
}

func (a *TestAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	if a.ProcessFn != nil {
		return a.ProcessFn(ctx, event, state)
	}
	return agentflow.AgentResult{
		OutputState: agentflow.NewState(),
	}, nil
}
