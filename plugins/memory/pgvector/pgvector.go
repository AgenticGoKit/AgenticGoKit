package pgvector

import (
	"github.com/kunalkushwaha/agenticgokit/core"
	providers "github.com/kunalkushwaha/agenticgokit/internal/memory/providers"
)

func init() {
	core.RegisterMemoryProviderFactory("pgvector", func(cfg core.AgentMemoryConfig) (core.Memory, error) {
		// Create embedding service via core registry-backed helpers
		var embed core.EmbeddingService
		switch cfg.Embedding.Provider {
		case "openai":
			embed = core.NewOpenAIEmbeddingService(cfg.Embedding.APIKey, cfg.Embedding.Model)
		case "ollama":
			embed = core.NewOllamaEmbeddingService(cfg.Embedding.Model, cfg.Embedding.BaseURL)
		case "dummy", "":
			embed = core.NewDummyEmbeddingService(cfg.Dimensions)
		default:
			// Fallback to dummy if unknown provider
			embed = core.NewDummyEmbeddingService(cfg.Dimensions)
		}
		return providers.NewPgVectorProvider(cfg, embed)
	})
}
