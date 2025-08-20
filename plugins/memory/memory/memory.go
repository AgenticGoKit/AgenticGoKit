package memory

import (
	"github.com/kunalkushwaha/agenticgokit/core"
	im "github.com/kunalkushwaha/agenticgokit/internal/memory/providers"
)

func init() {
	core.RegisterMemoryProviderFactory("memory", func(cfg core.AgentMemoryConfig) (core.Memory, error) {
		return im.NewInMemoryProvider(cfg)
	})
}
