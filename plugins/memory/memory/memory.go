package memory

import (
	"github.com/agenticgokit/agenticgokit/core"
	im "github.com/agenticgokit/agenticgokit/internal/memory/providers"
)

func init() {
	core.RegisterMemoryProviderFactory("memory", func(cfg core.AgentMemoryConfig) (core.Memory, error) {
		return im.NewInMemoryProvider(cfg)
	})
}

