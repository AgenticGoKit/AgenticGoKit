package agents

import (
	"context"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
)

// ConfigAwareAgent represents an agent that can be configured from ResolvedAgentConfig
type ConfigAwareAgent interface {
	core.Agent
	// Configuration methods
	GetRole() string
	GetDescription() string
	GetSystemPrompt() string
	GetCapabilities() []string
	IsEnabled() bool
	GetTimeout() time.Duration
	GetLLMConfig() *core.ResolvedLLMConfig

	// Configuration update methods
	UpdateConfiguration(config *core.ResolvedAgentConfig) error
	ApplySystemPrompt(ctx context.Context, state core.State) (core.State, error)
}

