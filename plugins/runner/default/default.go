package defaultrunner

import (
	"context"

	"github.com/agenticgokit/agenticgokit/core"
)

// memoryAwareAgentHandler wraps an AgentHandler to inject memory + session into context.
type memoryAwareAgentHandler struct {
	inner     core.AgentHandler
	memory    core.Memory
	sessionID string
}

func (m *memoryAwareAgentHandler) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
	if m.memory != nil && m.sessionID != "" {
		ctx = core.WithMemory(ctx, m.memory, m.sessionID)
	}
	return m.inner.Run(ctx, event, state)
}

// init registers a default runner factory built purely from core primitives.
// It avoids internal dependencies and is suitable for third-party consumption.
func init() {
	core.RegisterRunnerFactory(func(cfg core.RunnerConfig) core.Runner {
		// Create the underlying runner with queue size
		queueSize := cfg.QueueSize
		if queueSize <= 0 {
			queueSize = 100
		}
		r := core.NewRunner(queueSize)

		// Callbacks and tracing
		registry := cfg.Callbacks
		if registry == nil {
			registry = core.NewCallbackRegistry()
		}
		r.SetCallbackRegistry(registry)
		r.SetTraceLogger(core.NewInMemoryTraceLogger())

		// Orchestrator
		orch := core.NewRouteOrchestrator(registry)
		r.SetOrchestrator(orch)

		// Register agents (wrap with memory/session when provided)
		for name, handler := range cfg.Agents {
			h := handler
			if cfg.Memory != nil && cfg.SessionID != "" {
				h = &memoryAwareAgentHandler{inner: handler, memory: cfg.Memory, sessionID: cfg.SessionID}
			}
			_ = r.RegisterAgent(name, h)
		}

		// Configure error routing with defaults
		r.SetErrorRouterConfig(core.DefaultErrorRouterConfig())
		return r
	})
}

