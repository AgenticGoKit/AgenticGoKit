// Package agents provides internal agent implementations as fallbacks when full implementations aren't available.
// This package is NOT part of the public API and should not be imported directly by users.
package agents

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// Initialize the basic implementations in the core package during package initialization
func init() {
	// Basic implementations are available in core; no registration required here post-refactor
}

// basicAgentManager provides a minimal implementation when internal packages aren't imported
type basicAgentManager struct {
	config *core.Config
	agents map[string]core.Agent
	mutex  sync.RWMutex
}

// newBasicAgentManager creates a new basic agent manager
func newBasicAgentManager(config *core.Config) core.AgentManager {
	return &basicAgentManager{
		config: config,
		agents: make(map[string]core.Agent),
	}
}

func (am *basicAgentManager) UpdateAgentConfigurations(config *core.Config) error {
	core.Logger().Warn().
		Str("implementation", "basic").
		Msg("Using basic agent manager - import internal/agents for full functionality")

	am.mutex.Lock()
	defer am.mutex.Unlock()

	am.config = config
	// Basic implementation: just store the config, don't create agents yet
	return nil
}

func (am *basicAgentManager) GetCurrentAgents() map[string]core.Agent {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	// Return a copy to prevent concurrent modification
	agents := make(map[string]core.Agent, len(am.agents))
	for name, agent := range am.agents {
		agents[name] = agent
	}
	return agents
}

func (am *basicAgentManager) CreateAgent(name string, config *core.ResolvedAgentConfig) (core.Agent, error) {
	core.Logger().Warn().
		Str("agent", name).
		Str("implementation", "basic").
		Msg("Using basic agent manager - import internal/agents for full functionality")

	am.mutex.Lock()
	defer am.mutex.Unlock()

	agent := &basicAgent{name: name}
	am.agents[name] = agent
	return agent, nil
}

func (am *basicAgentManager) DisableAgent(name string) error {
	core.Logger().Warn().
		Str("agent", name).
		Str("implementation", "basic").
		Msg("Using basic agent manager - import internal/agents for full functionality")

	am.mutex.Lock()
	defer am.mutex.Unlock()

	delete(am.agents, name)
	return nil
}

func (am *basicAgentManager) InitializeAgents() error {
	core.Logger().Warn().
		Str("implementation", "basic").
		Msg("Using basic agent manager - import internal/agents for full functionality")

	am.mutex.Lock()
	defer am.mutex.Unlock()

	// Basic implementation: create default agents if configured
	if am.config != nil && len(am.config.Agents) > 0 {
		for agentName := range am.config.Agents {
			if _, exists := am.agents[agentName]; !exists {
				am.agents[agentName] = &basicAgent{name: agentName}
			}
		}
	}
	return nil
}

func (am *basicAgentManager) GetActiveAgents() []core.Agent {
	core.Logger().Warn().
		Str("implementation", "basic").
		Msg("Using basic agent manager - import internal/agents for full functionality")

	am.mutex.RLock()
	defer am.mutex.RUnlock()

	agents := make([]core.Agent, 0, len(am.agents))
	for name := range am.agents {
		// Create basic configurable agent with configuration if available
		role := "basic_processor"
		if am.config != nil {
			if agentConfig, exists := am.config.Agents[name]; exists {
				if agentConfig.Role != "" {
					role = agentConfig.Role
				}
			}
		}

		agents = append(agents, &basicConfigurableAgent{
			name:   name,
			role:   role,
			config: am.config,
		})
	}
	return agents
}

// basicConfigurableAgentFactory provides a minimal factory implementation
type basicConfigurableAgentFactory struct {
	config *core.Config
}

// newBasicConfigurableAgentFactory creates a new basic configurable agent factory
func newBasicConfigurableAgentFactory(config *core.Config) core.ConfigurableAgentFactory {
	return &basicConfigurableAgentFactory{
		config: config,
	}
}

func (f *basicConfigurableAgentFactory) CreateAgent(name string, resolvedConfig *core.ResolvedAgentConfig, llmProvider core.ModelProvider) (core.Agent, error) {
	core.Logger().Warn().
		Str("agent", name).
		Str("implementation", "basic").
		Msg("Using basic agent factory - import internal/agents for full functionality")

	return &basicAgent{name: name}, nil
}

func (f *basicConfigurableAgentFactory) CreateAgentFromConfig(name string, globalConfig *core.Config) (core.Agent, error) {
	core.Logger().Warn().
		Str("agent", name).
		Str("implementation", "basic").
		Msg("Using basic agent factory - import internal/agents for full functionality")

	return &basicAgent{name: name}, nil
}

func (f *basicConfigurableAgentFactory) GetAgentCapabilities(name string) []string {
	if f.config != nil {
		return f.config.GetAgentCapabilities(name)
	}
	return []string{"basic_processing"}
}

// basicAgent provides a minimal Agent implementation
type basicAgent struct {
	name string
}

func (a *basicAgent) Run(ctx context.Context, inputState core.State) (core.State, error) {
	// Basic implementation - add minimal metadata and return
	result := inputState.Clone()
	result.Set("processed_by", a.name)
	result.Set("agent_type", "basic")
	result.Set("processed_at", time.Now().Unix())
	return result, nil
}

func (a *basicAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
	// Convert to state processing pattern and back to result pattern
	startTime := time.Now()
	outputState, err := a.Run(ctx, state)
	endTime := time.Now()

	result := core.AgentResult{
		OutputState: outputState,
		StartTime:   startTime,
		EndTime:     endTime,
		Duration:    endTime.Sub(startTime),
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result, nil
}

func (a *basicAgent) Name() string {
	return a.name
}

func (a *basicAgent) GetRole() string {
	return "basic_processor"
}

func (a *basicAgent) GetDescription() string {
	return fmt.Sprintf("Basic agent: %s", a.name)
}

func (a *basicAgent) GetCapabilities() []string {
	return []string{"basic_processing"}
}

func (a *basicAgent) GetSystemPrompt() string {
	return "You are a basic processing agent."
}

func (a *basicAgent) GetTimeout() time.Duration {
	return 30 * time.Second
}

func (a *basicAgent) IsEnabled() bool {
	return true
}

func (a *basicAgent) GetLLMConfig() *core.ResolvedLLMConfig {
	return nil
}

func (a *basicAgent) Initialize(ctx context.Context) error {
	return nil // No initialization needed for basic agent
}

func (a *basicAgent) Shutdown(ctx context.Context) error {
	return nil // No cleanup needed for basic agent
}

// basicConfigurableAgent provides a minimal ConfigurableAgentInterface implementation
type basicConfigurableAgent struct {
	name   string
	role   string
	config *core.Config
}

func (a *basicConfigurableAgent) Run(ctx context.Context, inputState core.State) (core.State, error) {
	// Basic implementation - add metadata and return
	result := inputState.Clone()
	result.Set("processed_by", a.name)
	result.Set("agent_role", a.role)
	result.Set("agent_type", "basic_configurable")
	result.Set("processed_at", time.Now().Unix())
	result.Set("response", fmt.Sprintf("Processed by basic agent %s with role %s", a.name, a.role))
	return result, nil
}

func (a *basicConfigurableAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
	// Convert to state processing pattern and back to result pattern
	startTime := time.Now()
	outputState, err := a.Run(ctx, state)
	endTime := time.Now()

	result := core.AgentResult{
		OutputState: outputState,
		StartTime:   startTime,
		EndTime:     endTime,
		Duration:    endTime.Sub(startTime),
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result, nil
}

func (a *basicConfigurableAgent) Name() string {
	return a.name
}

func (a *basicConfigurableAgent) GetRole() string {
	return a.role
}

func (a *basicConfigurableAgent) GetDescription() string {
	return fmt.Sprintf("Basic configurable agent: %s (role: %s)", a.name, a.role)
}

func (a *basicConfigurableAgent) GetSystemPrompt() string {
	// Try to get from configuration first
	if a.config != nil {
		if agentConfig, exists := a.config.Agents[a.name]; exists {
			if agentConfig.SystemPrompt != "" {
				return agentConfig.SystemPrompt
			}
		}
	}
	return "You are a helpful assistant."
}

func (a *basicConfigurableAgent) GetCapabilities() []string {
	// Try to get from configuration first
	if a.config != nil {
		if agentConfig, exists := a.config.Agents[a.name]; exists {
			if len(agentConfig.Capabilities) > 0 {
				return agentConfig.Capabilities
			}
		}
	}
	return []string{"basic_processing"}
}

func (a *basicConfigurableAgent) IsEnabled() bool {
	// Try to get from configuration first
	if a.config != nil {
		if agentConfig, exists := a.config.Agents[a.name]; exists {
			return agentConfig.Enabled
		}
	}
	return true // Default to enabled
}

func (a *basicConfigurableAgent) GetTimeout() time.Duration {
	// Try to get from configuration first
	if a.config != nil {
		if agentConfig, exists := a.config.Agents[a.name]; exists {
			if agentConfig.Timeout > 0 {
				return time.Duration(agentConfig.Timeout) * time.Second
			}
		}
	}
	return 30 * time.Second // Default timeout
}

func (a *basicConfigurableAgent) GetLLMConfig() *core.ResolvedLLMConfig {
	// Basic implementation doesn't process LLM configurations
	// Return nil - full implementation handles LLM integration
	return nil
}

func (a *basicConfigurableAgent) Initialize(ctx context.Context) error {
	return nil // No initialization needed for basic agent
}

func (a *basicConfigurableAgent) Shutdown(ctx context.Context) error {
	return nil // No cleanup needed for basic agent
}

// NewBasicAgentManager creates a new basic agent manager
func NewBasicAgentManager(config *core.Config) core.AgentManager {
	if config == nil {
		config = &core.Config{}
	}
	return &basicAgentManager{
		config: config,
		agents: make(map[string]core.Agent),
		mutex:  sync.RWMutex{},
	}
}

// NewBasicConfigurableAgentFactory creates a new basic configurable agent factory
func NewBasicConfigurableAgentFactory(config *core.Config) core.ConfigurableAgentFactory {
	if config == nil {
		config = &core.Config{}
	}
	return &basicConfigurableAgentFactory{config: config}
}
