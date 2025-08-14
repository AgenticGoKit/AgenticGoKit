// Package embedding provides internal embedding service factory implementations.
package embedding

import (
	"fmt"

	"github.com/kunalkushwaha/agenticgokit/core"
	"github.com/kunalkushwaha/agenticgokit/internal/embedding/providers"
)

// EmbeddingFactory creates embedding services based on provider type
type EmbeddingFactory struct{}

// NewEmbeddingService creates a new embedding service based on provider configuration
func NewEmbeddingService(provider, model, apiKey, baseURL string, dimensions int) (core.EmbeddingService, error) {
	switch provider {
	case "openai":
		if apiKey == "" {
			return nil, fmt.Errorf("OpenAI API key is required for embedding service")
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
	// Register internal factories with core
	core.RegisterOpenAIEmbeddingFactory(providers.NewOpenAIEmbeddingService)
	core.RegisterOllamaEmbeddingFactory(providers.NewOllamaEmbeddingService)
	core.RegisterDummyEmbeddingFactory(providers.NewDummyEmbeddingService)
	
	// Also register locally for internal use
	RegisterOpenAIFactory(providers.NewOpenAIEmbeddingService)
	RegisterOllamaFactory(providers.NewOllamaEmbeddingService)
	RegisterDummyFactory(providers.NewDummyEmbeddingService)
}