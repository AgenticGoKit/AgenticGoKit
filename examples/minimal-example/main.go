package main

import (
	"context"
	"time"

	"github.com/kunalkushwaha/agentflow/core"
	agentflow "github.com/kunalkushwaha/agentflow/core"
)

// MinimalAgent implements agentflow.AgentHandler
type MinimalAgent struct{}

func (a *MinimalAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	agentflow.Logger().Info().Msg("MinimalAgent ran!")
	return agentflow.AgentResult{OutputState: state}, nil
}

func main() {

	agentflow.SetLogLevel(agentflow.INFO)

	agents := map[string]agentflow.AgentHandler{
		"minimal": &MinimalAgent{},
	}

	runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
		Agents: agents,
	})

	runner.Start(context.Background())
	defer runner.Stop()
	// Emit an event routed to the "minimal" agent
	runner.Emit(core.NewEvent(
		"test",
		nil,
		map[string]string{core.RouteMetadataKey: "minimal"},
	))

	// Give the agent time to run
	time.Sleep(500 * time.Millisecond)
}
