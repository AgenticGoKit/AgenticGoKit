package core

import (
	"fmt"
	"time"
)

// ConfigurableAgentInterface defines the interface for configuration-driven agents
type ConfigurableAgentInterface interface {
	Agent
	GetRole() string
	GetCapabilities() []string
	GetSystemPrompt() string
	GetTimeout() time.Duration
	IsEnabled() bool
	GetLLMConfig() *ResolvedLLMConfig
	GetDescription() string
}

// ConfigurableAgentFactory creates agents based on resolved configuration
type ConfigurableAgentFactory struct {
	// Implementation will be moved to internal package
	// For now, keep minimal structure for compatibility
	config *Config
}

// NewConfigurableAgentFactory creates a new configurable agent factory
func NewConfigurableAgentFactory(config *Config) *ConfigurableAgentFactory {
	return &ConfigurableAgentFactory{
		config: config,
	}
}

// CreateAgent creates an agent from resolved configuration
func (f *ConfigurableAgentFactory) CreateAgent(name string, resolvedConfig *ResolvedAgentConfig, llmProvider ModelProvider) (Agent, error) {
	// TODO: This will be replaced with internal implementation after refactoring is complete
	return nil, fmt.Errorf("configurable agent factory not yet implemented - use internal/agents package")
}

// CreateAgentFromConfig creates an agent directly from the global configuration
func (f *ConfigurableAgentFactory) CreateAgentFromConfig(name string, globalConfig *Config) (Agent, error) {
	// TODO: This will be replaced with internal implementation after refactoring is complete
	return nil, fmt.Errorf("configurable agent factory not yet implemented - use internal/agents package")
}

// GetAgentCapabilities returns the capabilities for a specific agent
func (f *ConfigurableAgentFactory) GetAgentCapabilities(name string) []string {
	return f.config.GetAgentCapabilities(name)
}

// IsAgentEnabled checks if an agent is enabled
func (f *ConfigurableAgentFactory) IsAgentEnabled(name string) bool {
	return f.config.IsAgentEnabled(name)
}

// CreateAllEnabledAgents creates all enabled agents from the configuration
func (f *ConfigurableAgentFactory) CreateAllEnabledAgents() (map[string]Agent, error) {
	// TODO: This will be replaced with internal implementation after refactoring is complete
	return nil, fmt.Errorf("configurable agent factory not yet implemented - use internal/agents package")
}

// ValidateAgentConfiguration validates that an agent can be created from configuration
func (f *ConfigurableAgentFactory) ValidateAgentConfiguration(name string) error {
	// TODO: This will be replaced with internal implementation after refactoring is complete
	return nil
}

