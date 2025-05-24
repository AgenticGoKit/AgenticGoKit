package factory

import (
	"log"
	"os"

	"github.com/kunalkushwaha/agentflow/core"
	"github.com/kunalkushwaha/agentflow/internal/llm"
	"github.com/kunalkushwaha/agentflow/internal/tools"
)

// RunnerConfig allows customization but provides sensible defaults.
type RunnerConfig = core.RunnerConfig

// NewRunnerWithConfig wires up everything, registers agents, and returns a ready-to-use runner.
var NewRunnerWithConfig = core.NewRunnerWithConfig

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
