package factory

import (
	"context"
	"log"
	"os"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/llm"
	"kunalkushwaha/agentflow/internal/orchestrator"
	"kunalkushwaha/agentflow/internal/tools"
)

// RunnerConfig allows customization but provides sensible defaults.
type RunnerConfig struct {
	QueueSize    int
	Orchestrator agentflow.Orchestrator
	Agents       map[string]agentflow.AgentHandler
}

// NewRunnerWithConfig wires up everything, registers agents, and returns a ready-to-use runner.
func NewRunnerWithConfig(cfg RunnerConfig) agentflow.Runner {
	queueSize := cfg.QueueSize
	if queueSize <= 0 {
		queueSize = 10
	}
	runner := agentflow.NewRunner(queueSize)

	// Callbacks and tracing
	callbackRegistry := agentflow.NewCallbackRegistry()
	traceLogger := agentflow.NewInMemoryTraceLogger()
	runner.SetCallbackRegistry(callbackRegistry)
	runner.SetTraceLogger(traceLogger)
	agentflow.RegisterTraceHooks(callbackRegistry, traceLogger)

	// Orchestrator
	var orch agentflow.Orchestrator
	if cfg.Orchestrator != nil {
		orch = cfg.Orchestrator
	} else {
		orch = orchestrator.NewRouteOrchestrator(callbackRegistry)
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
		runner.RegisterAgent("error-handler", agentflow.AgentHandlerFunc(
			func(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
				state.SetMeta(agentflow.RouteMetadataKey, "")
				return agentflow.AgentResult{OutputState: state}, nil
			},
		))
	}

	return runner
}

// NewDefaultToolRegistry returns a ToolRegistry with built-in tools registered.
func NewDefaultToolRegistry() *tools.ToolRegistry {
	registry := tools.NewToolRegistry()
	_ = registry.Register(&tools.WebSearchTool{})
	_ = registry.Register(&tools.ComputeMetricTool{})
	return registry
}

// NewDefaultLLMAdapter returns an Azure OpenAI LLM adapter using environment variables.
func NewDefaultLLMAdapter() llm.ModelProvider {
	options := llm.AzureOpenAIAdapterOptions{
		Endpoint:            os.Getenv("AZURE_OPENAI_ENDPOINT"),
		APIKey:              os.Getenv("AZURE_OPENAI_API_KEY"),
		ChatDeployment:      os.Getenv("AZURE_OPENAI_DEPLOYMENT_ID"),
		EmbeddingDeployment: os.Getenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT"),
	}
	azureLLM, err := llm.NewAzureOpenAIAdapter(options)
	if err != nil {
		log.Fatalf("Failed to create Azure OpenAI adapter: %v", err)
	}
	return azureLLM
}
