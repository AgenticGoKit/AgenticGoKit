package pgvector

import (
	"github.com/agenticgokit/agenticgokit/core"
	providers "github.com/agenticgokit/agenticgokit/internal/memory/providers"
)

func init() {
	core.RegisterMemoryProviderFactory("pgvector", func(cfg core.AgentMemoryConfig) (core.Memory, error) {
		// Create embedding service via core registry-backed helpers; fails
		// loudly on unregistered factories or unknown providers instead of
		// silently degrading to zero-vector embeddings (see issue #137).
		embed, err := core.NewEmbeddingServiceForConfig(cfg)
		if err != nil {
			return nil, err
		}
		return providers.NewPgVectorProvider(cfg, embed)
	})
}
