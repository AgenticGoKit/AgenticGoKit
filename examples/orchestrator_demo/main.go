package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"

	// Activate plugins via blank imports
	_ "github.com/kunalkushwaha/agenticgokit/plugins/logging/zerolog"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/memory/memory"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/orchestrator/default"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/runner/default"
)

// echoAgent just echoes payload and routes to "done" for success.
func echoAgent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
	// Merge event data into state
	for k, v := range event.GetData() {
		state.Set(k, v)
	}
	state.Set("handled_by", "echo")
	state.Set("timestamp", time.Now().Format(time.RFC3339))
	state.SetMeta(core.RouteMetadataKey, "done")
	return core.AgentResult{OutputState: state}, nil
}

// doneAgent prints the final state
func doneAgent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
	fmt.Println("Final state:")
	for _, k := range state.Keys() {
		if v, ok := state.Get(k); ok {
			fmt.Printf("- %s: %v\n", k, v)
		}
	}
	return core.AgentResult{OutputState: state}, nil
}

// errorHandler consumes error events and clears routing
func errorHandler(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
	state.SetMeta(core.RouteMetadataKey, "")
	return core.AgentResult{OutputState: state}, nil
}

func main() {
	// Minimal runner config with in-memory provider from plugin
	mem := core.QuickMemory()
	sessionID := core.GenerateSessionID()

	r := core.NewRunnerWithConfig(core.RunnerConfig{
		QueueSize: 100,
		Agents: map[string]core.AgentHandler{
			"echo":          core.AgentHandlerFunc(echoAgent),
			"done":          core.AgentHandlerFunc(doneAgent),
			"error-handler": core.AgentHandlerFunc(errorHandler),
		},
		Memory:    mem,
		Callbacks: core.NewCallbackRegistry(),
		SessionID: sessionID,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = r.Start(ctx)
	defer r.Stop()

	// Emit an event to echo
	evt := core.NewEvent("echo", core.EventData{"message": "hello orchestration"}, map[string]string{
		core.SessionIDKey:     sessionID,
		core.RouteMetadataKey: "echo",
	})
	_ = r.Emit(evt)

	// Give it a moment to process
	time.Sleep(500 * time.Millisecond)
}
