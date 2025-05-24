// Package core provides public factory functions for creating runners and tool registries in AgentFlow.
package core

import (
	"context"
	"log"
)

// RunnerConfig allows customization but provides sensible defaults.
type RunnerConfig struct {
	QueueSize    int
	Orchestrator Orchestrator
	Agents       map[string]AgentHandler
}

// NewRunnerWithConfig wires up everything, registers agents, and returns a ready-to-use runner.
func NewRunnerWithConfig(cfg RunnerConfig) Runner {
	queueSize := cfg.QueueSize
	if queueSize <= 0 {
		queueSize = 10
	}
	runner := NewRunner(queueSize)

	// Callbacks and tracing
	callbackRegistry := NewCallbackRegistry()
	traceLogger := NewInMemoryTraceLogger()
	runner.SetCallbackRegistry(callbackRegistry)
	runner.SetTraceLogger(traceLogger)
	RegisterTraceHooks(callbackRegistry, traceLogger)

	// Orchestrator
	var orch Orchestrator
	if cfg.Orchestrator != nil {
		orch = cfg.Orchestrator
	} else {
		orch = NewRouteOrchestrator(callbackRegistry)
	}
	runner.SetOrchestrator(orch)

	// Register agents
	for name, agent := range cfg.Agents {
		if err := runner.RegisterAgent(name, agent); err != nil {
			log.Fatalf("Failed to register agent %s: %v", name, err)
		}
	}

	// Register a default no-op error handler if not present
	if _, ok := cfg.Agents["error-handler"]; !ok {
		runner.RegisterAgent("error-handler", AgentHandlerFunc(
			func(ctx context.Context, event Event, state State) (AgentResult, error) {
				state.SetMeta(RouteMetadataKey, "")
				return AgentResult{OutputState: state}, nil
			},
		))
	}

	return runner
}

// ...other factory functions (e.g., for tool registry, LLM adapter) can be added here...
