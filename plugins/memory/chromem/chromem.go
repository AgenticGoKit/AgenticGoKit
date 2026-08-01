package chromem

import (
	"github.com/agenticgokit/agenticgokit/core"
	providers "github.com/agenticgokit/agenticgokit/internal/memory/providers"
)

func init() {
	core.RegisterMemoryProviderFactory("chromem", func(config core.AgentMemoryConfig) (core.Memory, error) {
		// Initialize embedding service; fails loudly on unregistered factories
		// or unknown providers instead of silently degrading to zero-vector
		// embeddings (see issue #137).
		embedder, err := core.NewEmbeddingServiceForConfig(config)
		if err != nil {
			return nil, err
		}

		return providers.NewChromemProvider(config, embedder)
	})
}
