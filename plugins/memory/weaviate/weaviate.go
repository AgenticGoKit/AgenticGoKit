package weaviate

import (
	"github.com/agenticgokit/agenticgokit/core"
	providers "github.com/agenticgokit/agenticgokit/internal/memory/providers"
)

// Register the Weaviate memory provider. Implementation is currently a stub in internal/providers.
func init() {
	core.RegisterMemoryProviderFactory("weaviate", func(cfg core.AgentMemoryConfig) (core.Memory, error) {
		return providers.NewWeaviateProvider(cfg)
	})
}

