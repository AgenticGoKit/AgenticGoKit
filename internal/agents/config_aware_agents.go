package agents

import (
	"context"
	"time"
)

// ConfigAwareAgent represents an agent that can be configured from ResolvedAgentConfig
type ConfigAwareAgent interface {
	Agent
	// Configuration methods
	GetRole() string
	GetDescription() string
	GetSystemPrompt() string
	GetCapabilities() []string
	IsEnabled() bool
	GetTimeout() time.Duration
	GetLLMConfig() *ResolvedLLMConfig
	
	// Configuration update methods
	UpdateConfiguration(config *ResolvedAgentConfig) error
	ApplySystemPrompt(ctx context.Context, state State) (State, error)
}

