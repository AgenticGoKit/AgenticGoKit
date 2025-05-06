package factory

import (
	"log"
	"os"

	"kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/llm"
	"kunalkushwaha/agentflow/internal/orchestrator"
	"kunalkushwaha/agentflow/internal/tools"
)

func NewDefaultRunner() *core.Runner {
	orch := orchestrator.NewRouteOrchestrator()
	return core.NewRunner(orch, 10)
}

func NewDefaultAgent(name string) core.Agent {
	return &defaultAgent{name: name}
}

type defaultAgent struct {
	name string
}

func (a *defaultAgent) Name() string {
	return a.name
}

func (a *defaultAgent) Run(ctx core.Context, state core.State) (core.State, error) {
	return state, nil
}

func NewDefaultLLMAdapter() llm.ModelProvider {
	options := llm.AzureOpenAIAdapterOptions{
		Endpoint:            "https://your-resource-name.openai.azure.com",
		APIKey:              os.Getenv("AZURE_OPENAI_API_KEY"),
		ChatDeployment:      "gpt-4-turbo",
		EmbeddingDeployment: "text-embedding-3",
	}

	azureLLM, err := llm.NewAzureOpenAIAdapter(options)
	if err != nil {
		log.Fatalf("Failed to create Azure OpenAI adapter: %v", err)
	}
	return azureLLM
}

func NewDefaultToolRegistry() *tools.ToolRegistry {
	registry := tools.NewToolRegistry()
	registry.Register(&tools.WebSearchTool{})
	registry.Register(&tools.ComputeMetricTool{})
	return registry
}