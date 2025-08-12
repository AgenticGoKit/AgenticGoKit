package core

import (
	"fmt"
	"sync"
)

// DefaultAgentManager implements the AgentManager interface
type DefaultAgentManager struct {
	factory      *ConfigurableAgentFactory
	agents       map[string]Agent
	agentConfigs map[string]*ResolvedAgentConfig
	mutex        sync.RWMutex
}

// NewAgentManager creates a new agent manager
func NewAgentManager(factory *ConfigurableAgentFactory) *DefaultAgentManager {
	return &DefaultAgentManager{
		factory:      factory,
		agents:       make(map[string]Agent),
		agentConfigs: make(map[string]*ResolvedAgentConfig),
	}
}

// UpdateAgentConfigurations updates all agent configurations from the new config
func (am *DefaultAgentManager) UpdateAgentConfigurations(config *Config) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	Logger().Info().
		Int("agent_count", len(config.Agents)).
		Msg("Updating agent configurations")

	// Track which agents are in the new configuration
	newAgentNames := make(map[string]bool)
	
	// Update or create agents from the new configuration
	for agentName := range config.Agents {
		newAgentNames[agentName] = true
		
		Logger().Debug().
			Str("agent", agentName).
			Msg("Processing agent configuration")
		
		// Create resolver and resolve the agent configuration
		resolver := NewConfigResolver(config)
		err := resolver.ApplyEnvironmentOverrides()
		if err != nil {
			Logger().Error().
				Err(err).
				Str("agent", agentName).
				Msg("Failed to apply environment overrides")
			continue
		}
		
		resolvedConfig, err := resolver.ResolveAgentConfigWithEnv(agentName)
		if err != nil {
			Logger().Error().
				Err(err).
				Str("agent", agentName).
				Msg("Failed to resolve agent configuration")
			continue
		}

		// Handle disabled agents
		if !resolvedConfig.Enabled {
			Logger().Debug().
				Str("agent", agentName).
				Msg("Processing disabled agent")
				
			// If agent exists, disable it (using internal method to avoid deadlock)
			if _, exists := am.agents[agentName]; exists {
				err := am.disableAgentInternal(agentName)
				if err != nil {
					Logger().Error().
						Err(err).
						Str("agent", agentName).
						Msg("Failed to disable agent")
				} else {
					Logger().Info().
						Str("agent", agentName).
						Msg("Disabled agent via configuration")
				}
			} else {
				Logger().Debug().
					Str("agent", agentName).
					Msg("Agent not found, storing disabled configuration")
			}
			// Store the disabled configuration
			am.agentConfigs[agentName] = resolvedConfig
			continue
		}

		// Check if agent already exists
		if existingAgent, exists := am.agents[agentName]; exists {
			// Update existing agent configuration if it supports it
			if configurableAgent, ok := existingAgent.(ConfigurableAgent); ok {
				err := configurableAgent.UpdateConfiguration(resolvedConfig)
				if err != nil {
					Logger().Error().
						Err(err).
						Str("agent", agentName).
						Msg("Failed to update existing agent configuration")
					continue
				}
				
				am.agentConfigs[agentName] = resolvedConfig
				
				Logger().Info().
					Str("agent", agentName).
					Str("role", resolvedConfig.Role).
					Bool("enabled", resolvedConfig.Enabled).
					Msg("Updated existing agent configuration")
			} else {
				// Agent doesn't support configuration updates, recreate it
				newAgent, err := am.factory.CreateAgent(agentName, resolvedConfig, nil)
				if err != nil {
					Logger().Error().
						Err(err).
						Str("agent", agentName).
						Msg("Failed to recreate agent with new configuration")
					continue
				}
				
				am.agents[agentName] = newAgent
				am.agentConfigs[agentName] = resolvedConfig
				
				Logger().Info().
					Str("agent", agentName).
					Str("role", resolvedConfig.Role).
					Bool("enabled", resolvedConfig.Enabled).
					Msg("Recreated agent with new configuration")
			}
		} else {
			// Create new agent
			Logger().Debug().
				Str("agent", agentName).
				Str("role", resolvedConfig.Role).
				Bool("enabled", resolvedConfig.Enabled).
				Msg("Creating new agent from configuration")
				
			newAgent, err := am.factory.CreateAgent(agentName, resolvedConfig, nil)
			if err != nil {
				Logger().Error().
					Err(err).
					Str("agent", agentName).
					Msg("Failed to create new agent")
				continue
			}
			
			am.agents[agentName] = newAgent
			am.agentConfigs[agentName] = resolvedConfig
			
			Logger().Info().
				Str("agent", agentName).
				Str("role", resolvedConfig.Role).
				Bool("enabled", resolvedConfig.Enabled).
				Msg("Created new agent from configuration")
		}
	}

	// Handle agents that are no longer in the configuration
	for agentName := range am.agents {
		if !newAgentNames[agentName] {
			// Agent is no longer in configuration, disable it
			err := am.DisableAgent(agentName)
			if err != nil {
				Logger().Error().
					Err(err).
					Str("agent", agentName).
					Msg("Failed to disable removed agent")
			} else {
				Logger().Info().
					Str("agent", agentName).
					Msg("Disabled agent (removed from configuration)")
			}
		}
	}

	Logger().Info().
		Int("total_agents", len(am.agents)).
		Int("config_agents", len(config.Agents)).
		Msg("Agent configuration update completed")

	return nil
}

// GetCurrentAgents returns a copy of the current agents map
func (am *DefaultAgentManager) GetCurrentAgents() map[string]Agent {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	agents := make(map[string]Agent, len(am.agents))
	for name, agent := range am.agents {
		agents[name] = agent
	}
	return agents
}

// CreateAgent creates a new agent with the given configuration
func (am *DefaultAgentManager) CreateAgent(name string, config *ResolvedAgentConfig) (Agent, error) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	// Create the agent using the factory
	agent, err := am.factory.CreateAgent(name, config, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent %s: %w", name, err)
	}

	// Store the agent and its configuration
	am.agents[name] = agent
	am.agentConfigs[name] = config

	Logger().Info().
		Str("agent", name).
		Str("role", config.Role).
		Bool("enabled", config.Enabled).
		Msg("Created agent")

	return agent, nil
}

// DisableAgent disables an agent by name
func (am *DefaultAgentManager) DisableAgent(name string) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	return am.disableAgentInternal(name)
}

// disableAgentInternal disables an agent without acquiring the mutex (for internal use)
func (am *DefaultAgentManager) disableAgentInternal(name string) error {
	agent, exists := am.agents[name]
	if !exists {
		return fmt.Errorf("agent %s not found", name)
	}

	// If the agent supports configuration, disable it
	if configurableAgent, ok := agent.(ConfigurableAgent); ok {
		config := am.agentConfigs[name]
		if config != nil {
			// Create a disabled version of the configuration
			disabledConfig := *config
			disabledConfig.Enabled = false
			
			err := configurableAgent.UpdateConfiguration(&disabledConfig)
			if err != nil {
				Logger().Error().
					Err(err).
					Str("agent", name).
					Msg("Failed to disable agent via configuration")
			} else {
				am.agentConfigs[name] = &disabledConfig
				Logger().Info().
					Str("agent", name).
					Msg("Disabled agent via configuration")
				return nil
			}
		}
	}

	// If we can't disable via configuration, remove the agent
	delete(am.agents, name)
	delete(am.agentConfigs, name)

	Logger().Info().
		Str("agent", name).
		Msg("Removed disabled agent")

	return nil
}

// GetAgent returns an agent by name
func (am *DefaultAgentManager) GetAgent(name string) (Agent, bool) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	agent, exists := am.agents[name]
	return agent, exists
}

// GetAgentConfig returns an agent's configuration by name
func (am *DefaultAgentManager) GetAgentConfig(name string) (*ResolvedAgentConfig, bool) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	config, exists := am.agentConfigs[name]
	return config, exists
}

// ListAgents returns a list of all agent names
func (am *DefaultAgentManager) ListAgents() []string {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	names := make([]string, 0, len(am.agents))
	for name := range am.agents {
		names = append(names, name)
	}
	return names
}

// ListEnabledAgents returns a list of enabled agent names
func (am *DefaultAgentManager) ListEnabledAgents() []string {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	names := make([]string, 0, len(am.agents))
	for name, config := range am.agentConfigs {
		if config != nil && config.Enabled {
			names = append(names, name)
		}
	}
	return names
}

// GetAgentCount returns the total number of agents
func (am *DefaultAgentManager) GetAgentCount() int {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	return len(am.agents)
}

// GetEnabledAgentCount returns the number of enabled agents
func (am *DefaultAgentManager) GetEnabledAgentCount() int {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	count := 0
	for _, config := range am.agentConfigs {
		if config != nil && config.Enabled {
			count++
		}
	}
	return count
}

// ConfigurableAgent interface for agents that support configuration updates
type ConfigurableAgent interface {
	UpdateConfiguration(config *ResolvedAgentConfig) error
}