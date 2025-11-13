// Package embedding provides internal embedding service factory implementations.
package embedding

import (
	"fmt"
	"os"

	"github.com/agenticgokit/agenticgokit/core"
	"github.com/agenticgokit/agenticgokit/internal/embedding/providers"
)

// EmbeddingFactory creates embedding services based on provider type
type EmbeddingFactory struct{}

// NewEmbeddingService creates a new embedding service based on provider configuration
func NewEmbeddingService(provider, model, apiKey, baseURL string, dimensions int) (core.EmbeddingService, error) {
	switch provider {
	case "openai":
		// If no API key provided, try environment variable
		if apiKey == "" {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
		if apiKey == "" {
			return nil, fmt.Errorf("OpenAI API key is required for embedding service. Set OPENAI_API_KEY environment variable or provide api_key in configuration")
		}
		return providers.NewOpenAIEmbeddingService(apiKey, model), nil
	case "azure":
		// If no API key provided, try environment variable
		if apiKey == "" {
			apiKey = os.Getenv("AZURE_OPENAI_API_KEY")
		}
		if apiKey == "" {
			return nil, fmt.Errorf("Azure OpenAI API key is required for embedding service. Set AZURE_OPENAI_API_KEY environment variable or provide api_key in configuration")
		}
		return providers.NewOpenAIEmbeddingService(apiKey, model), nil
	case "ollama":
		return providers.NewOllamaEmbeddingService(model, baseURL), nil
	case "dummy":
		return providers.NewDummyEmbeddingService(dimensions), nil
	default:
		return nil, fmt.Errorf("unsupported embedding provider: %s", provider)
	}
}

// Embedding service factory functions for registry pattern
var (
	openAIFactory func(string, string) core.EmbeddingService
	ollamaFactory func(string, string) core.EmbeddingService
	dummyFactory  func(int) core.EmbeddingService
)

// RegisterOpenAIFactory registers the OpenAI embedding service factory
func RegisterOpenAIFactory(factory func(string, string) core.EmbeddingService) {
	openAIFactory = factory
}

// RegisterOllamaFactory registers the Ollama embedding service factory
func RegisterOllamaFactory(factory func(string, string) core.EmbeddingService) {
	ollamaFactory = factory
}

// RegisterDummyFactory registers the Dummy embedding service factory
func RegisterDummyFactory(factory func(int) core.EmbeddingService) {
	dummyFactory = factory
}

// GetOpenAIFactory returns the registered OpenAI factory
func GetOpenAIFactory() func(string, string) core.EmbeddingService {
	return openAIFactory
}

// GetOllamaFactory returns the registered Ollama factory
func GetOllamaFactory() func(string, string) core.EmbeddingService {
	return ollamaFactory
}

// GetDummyFactory returns the registered Dummy factory
func GetDummyFactory() func(int) core.EmbeddingService {
	return dummyFactory
}

// Register all embedding service factories with core to avoid circular imports
func init() {
	// Register internal factories with core - wrap with environment variable support
	core.RegisterOpenAIEmbeddingFactory(func(apiKey, model string) core.EmbeddingService {
		// If no API key provided, try environment variable
		if apiKey == "" {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
		return providers.NewOpenAIEmbeddingService(apiKey, model)
	})
	core.RegisterOllamaEmbeddingFactory(providers.NewOllamaEmbeddingService)
	core.RegisterDummyEmbeddingFactory(providers.NewDummyEmbeddingService)

	// Also register locally for internal use
	RegisterOpenAIFactory(providers.NewOpenAIEmbeddingService)
	RegisterOllamaFactory(providers.NewOllamaEmbeddingService)
	RegisterDummyFactory(providers.NewDummyEmbeddingService)
}

