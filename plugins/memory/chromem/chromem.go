package chromem

import (
	"github.com/agenticgokit/agenticgokit/core"
	providers "github.com/agenticgokit/agenticgokit/internal/memory/providers"
)

func init() {
	core.RegisterMemoryProviderFactory("chromem", func(config core.AgentMemoryConfig) (core.Memory, error) {
		// Initialize embedding service
		var embedder core.EmbeddingService
		switch config.Embedding.Provider {
		case "openai":
			embedder = core.NewOpenAIEmbeddingService(config.Embedding.APIKey, config.Embedding.Model)
		case "ollama":
			embedder = core.NewOllamaEmbeddingService(config.Embedding.Model, config.Embedding.BaseURL)
		case "dummy":
			embedder = core.NewDummyEmbeddingService(config.Dimensions)
		default:
			// Default to dummy if not specified or unrecognized
			embedder = core.NewDummyEmbeddingService(config.Dimensions)
		}

		return providers.NewChromemProvider(config, embedder)
	})
}
