package main

import (
	"log"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/factory"
)

// MinimalAgent is a very simple agent
type MinimalAgent struct{}

func (a *MinimalAgent) Run(ctx agentflow.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	log.Println("Minimal agent ran")
	return agentflow.AgentResult{}, nil
}

func main() {
	agent := &MinimalAgent{}
	runner := factory.NewRunnerBuilder().RegisterAgent("minimal", agent).BuildOrPanic()
	runner.Start()
	runner.Emit(agentflow.NewEvent("test", nil, nil))
	time.Sleep(1 * time.Second)
	runner.Stop()
}