// Package agents provides internal agent factory implementations for AgentFlow.
package agents

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// ConfigurableAgentFactory creates agents based on resolved configuration
type ConfigurableAgentFactory struct {
	config *core.Config
}

// NewConfigurableAgentFactory creates a new configurable agent factory
func NewConfigurableAgentFactory(config *core.Config) *ConfigurableAgentFactory {
	return &ConfigurableAgentFactory{
		config: config,
	}
}

// CreateAgent creates an agent from resolved configuration
func (f *ConfigurableAgentFactory) CreateAgent(name string, resolvedConfig *core.ResolvedAgentConfig, llmProvider core.ModelProvider) (core.Agent, error) {
	if resolvedConfig == nil {
		return nil, fmt.Errorf("resolved configuration is required for agent '%s'", name)
	}

	if !resolvedConfig.Enabled {
		return nil, fmt.Errorf("agent '%s' is disabled in configuration", name)
	}

	// Start building the agent using internal builder
	builder := NewAgent(name)

	// Add LLM capability if provider is available
	if llmProvider != nil && resolvedConfig.LLMConfig != nil {
		llmConfig := f.createLLMConfigFromResolved(resolvedConfig.LLMConfig)
		builder = builder.WithLLMAndConfig(llmProvider, llmConfig)
	}

	// Add metrics capability if enabled (default for all agents)
	builder = builder.WithDefaultMetrics()

	// Add retry policy if configured
	if resolvedConfig.RetryPolicy != nil {
		// Note: This would require implementing a retry capability
		// For now, we'll log that retry policy is configured
		core.Logger().Info().
			Str("agent", name).
			Int("max_retries", resolvedConfig.RetryPolicy.MaxRetries).
			Msg("Agent configured with retry policy")
	}

	// Add rate limiting if configured
	if resolvedConfig.RateLimit != nil {
		// Note: This would require implementing a rate limiting capability
		// For now, we'll log that rate limiting is configured
		core.Logger().Info().
			Str("agent", name).
			Int("requests_per_second", resolvedConfig.RateLimit.RequestsPerSecond).
			Msg("Agent configured with rate limiting")
	}

	// Build the agent
	agent, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build agent '%s': %w", name, err)
	}

	// Apply AutoLLM configuration if the agent supports it
	if unifiedAgent, ok := agent.(*core.UnifiedAgent); ok {
		unifiedAgent.SetAutoLLM(resolvedConfig.AutoLLM)
		core.Logger().Info().
			Str("agent", name).
			Bool("auto_llm", resolvedConfig.AutoLLM).
			Msg("Applied AutoLLM configuration")
	}

	// Create a wrapper that includes the configuration metadata
	configuredAgent := &ConfiguredAgent{
		Agent:          agent,
		AgentName:      name,
		Config:         resolvedConfig,
		OriginalConfig: f.config,
	}

	core.Logger().Info().
		Str("agent", name).
		Str("role", resolvedConfig.Role).
		Strs("capabilities", resolvedConfig.Capabilities).
		Bool("enabled", resolvedConfig.Enabled).
		Dur("timeout", resolvedConfig.Timeout).
		Msg("Agent created from configuration")

	return configuredAgent, nil
}

// CreateAgentFromConfig creates an agent directly from the global configuration
func (f *ConfigurableAgentFactory) CreateAgentFromConfig(name string, globalConfig *core.Config) (core.Agent, error) {
	// Resolve the agent configuration
	resolver := core.NewConfigResolver(globalConfig)
	resolvedConfig, err := resolver.ResolveAgentConfigWithEnv(name)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve configuration for agent '%s': %w", name, err)
	}

	// Create LLM provider from resolved configuration
	llmProvider, err := f.createLLMProviderFromConfig(resolvedConfig.LLMConfig, globalConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider for agent '%s': %w", name, err)
	}

	return f.CreateAgent(name, resolvedConfig, llmProvider)
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
func (f *ConfigurableAgentFactory) CreateAllEnabledAgents() (map[string]core.Agent, error) {
	agents := make(map[string]core.Agent)
	enabledAgentNames := f.config.GetEnabledAgents()

	for _, agentName := range enabledAgentNames {
		agent, err := f.CreateAgentFromConfig(agentName, f.config)
		if err != nil {
			return nil, fmt.Errorf("failed to create agent '%s': %w", agentName, err)
		}
		agents[agentName] = agent
	}

	core.Logger().Info().
		Int("agent_count", len(agents)).
		Strs("agent_names", enabledAgentNames).
		Msg("Created all enabled agents from configuration")

	return agents, nil
}

// ValidateAgentConfiguration validates that an agent can be created from configuration
func (f *ConfigurableAgentFactory) ValidateAgentConfiguration(name string) error {
	// Check if agent exists in configuration
	if !f.config.IsAgentEnabled(name) {
		return fmt.Errorf("agent '%s' is not enabled or does not exist in configuration", name)
	}

	// Resolve configuration to check for errors
	resolver := core.NewConfigResolver(f.config)
	resolvedConfig, err := resolver.ResolveAgentConfigWithEnv(name)
	if err != nil {
		return fmt.Errorf("failed to resolve configuration for agent '%s': %w", name, err)
	}

	// Validate resolved configuration
	if resolvedConfig.Role == "" {
		return fmt.Errorf("agent '%s' has no role defined", name)
	}

	if resolvedConfig.SystemPrompt == "" {
		return fmt.Errorf("agent '%s' has no system prompt defined", name)
	}

	if len(resolvedConfig.Capabilities) == 0 {
		return fmt.Errorf("agent '%s' has no capabilities defined", name)
	}

	// Validate LLM configuration if present
	if resolvedConfig.LLMConfig != nil {
		if resolvedConfig.LLMConfig.Provider == "" {
			return fmt.Errorf("agent '%s' has LLM configuration but no provider specified", name)
		}
	}

	return nil
}

// createLLMConfigFromResolved converts ResolvedLLMConfig to LLMConfig for the builder
func (f *ConfigurableAgentFactory) createLLMConfigFromResolved(resolved *core.ResolvedLLMConfig) core.LLMConfig {
	return core.LLMConfig{
		Temperature:    resolved.Temperature,
		MaxTokens:      resolved.MaxTokens,
		TimeoutSeconds: int(resolved.Timeout.Seconds()),
	}
}

// createLLMProviderFromConfig creates an LLM provider from resolved configuration
func (f *ConfigurableAgentFactory) createLLMProviderFromConfig(llmConfig *core.ResolvedLLMConfig, globalConfig *core.Config) (core.ModelProvider, error) {
	if llmConfig == nil {
		return nil, fmt.Errorf("LLM configuration is required")
	}

	// Get provider-specific configuration from global config
	providerConfig, exists := globalConfig.Providers[llmConfig.Provider]
	if !exists {
		return nil, fmt.Errorf("no provider configuration found for '%s'", llmConfig.Provider)
	}

	// Create provider based on type
	switch llmConfig.Provider {
	case "openai":
		return f.createOpenAIProvider(providerConfig, llmConfig)
	case "azure":
		return f.createAzureProvider(providerConfig, llmConfig)
	case "ollama":
		return f.createOllamaProvider(providerConfig, llmConfig)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", llmConfig.Provider)
	}
}

// createOpenAIProvider creates an OpenAI provider with resolved configuration
func (f *ConfigurableAgentFactory) createOpenAIProvider(providerConfig map[string]interface{}, llmConfig *core.ResolvedLLMConfig) (core.ModelProvider, error) {
	apiKey := f.getStringValue(providerConfig, "api_key")
	if apiKey == "" {
		// Fallback to environment variable
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OpenAI API key not found in provider configuration or OPENAI_API_KEY environment variable")
		}
	}

	return core.NewOpenAIAdapter(
		apiKey,
		llmConfig.Model,
		llmConfig.MaxTokens,
		float32(llmConfig.Temperature),
	)
}

// createAzureProvider creates an Azure OpenAI provider with resolved configuration
func (f *ConfigurableAgentFactory) createAzureProvider(providerConfig map[string]interface{}, llmConfig *core.ResolvedLLMConfig) (core.ModelProvider, error) {
	endpoint := f.getStringValue(providerConfig, "endpoint")
	if endpoint == "" {
		// Fallback to environment variable
		endpoint = os.Getenv("AZURE_OPENAI_ENDPOINT")
		if endpoint == "" {
			return nil, fmt.Errorf("Azure OpenAI endpoint not found in provider configuration or AZURE_OPENAI_ENDPOINT environment variable")
		}
	}

	apiKey := f.getStringValue(providerConfig, "api_key")
	if apiKey == "" {
		// Fallback to environment variable
		apiKey = os.Getenv("AZURE_OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("Azure OpenAI API key not found in provider configuration or AZURE_OPENAI_API_KEY environment variable")
		}
	}

	chatDeployment := f.getStringValue(providerConfig, "chat_deployment")
	if chatDeployment == "" {
		// Fallback to environment variable
		chatDeployment = os.Getenv("AZURE_OPENAI_DEPLOYMENT")
		if chatDeployment == "" {
			return nil, fmt.Errorf("Azure OpenAI chat deployment not found in provider configuration or AZURE_OPENAI_DEPLOYMENT environment variable")
		}
	}

	embeddingDeployment := f.getStringValue(providerConfig, "embedding_deployment")
	if embeddingDeployment == "" {
		// Fallback to environment variable, then default
		embeddingDeployment = os.Getenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT")
		if embeddingDeployment == "" {
			embeddingDeployment = "text-embedding-ada-002" // default
		}
	}

	return core.NewAzureOpenAIAdapter(core.AzureOpenAIAdapterOptions{
		Endpoint:            endpoint,
		APIKey:              apiKey,
		ChatDeployment:      chatDeployment,
		EmbeddingDeployment: embeddingDeployment,
	})
}

// createOllamaProvider creates an Ollama provider with resolved configuration
func (f *ConfigurableAgentFactory) createOllamaProvider(providerConfig map[string]interface{}, llmConfig *core.ResolvedLLMConfig) (core.ModelProvider, error) {
	baseURL := f.getStringValue(providerConfig, "base_url")
	if baseURL == "" {
		baseURL = f.getStringValue(providerConfig, "endpoint") // alias support
	}
	if baseURL == "" {
		baseURL = "http://localhost:11434" // default
	}

	return core.NewOllamaAdapter(
		baseURL,
		llmConfig.Model,
		llmConfig.MaxTokens,
		float32(llmConfig.Temperature),
	)
}

// Helper methods to safely extract values from provider configuration
func (f *ConfigurableAgentFactory) getStringValue(config map[string]interface{}, key string) string {
	if val, exists := config[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// ConfiguredAgent wraps an agent with its configuration metadata
type ConfiguredAgent struct {
	Agent          core.Agent
	AgentName      string
	Config         *core.ResolvedAgentConfig
	OriginalConfig *core.Config
}

// Run implements the Agent interface with configuration-aware behavior
func (ca *ConfiguredAgent) Run(ctx context.Context, inputState core.State) (core.State, error) {
	// Apply timeout from configuration
	if ca.Config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, ca.Config.Timeout)
		defer cancel()
	}

	// Add configuration metadata to state
	inputState.Set("agent_name", ca.AgentName)
	inputState.Set("agent_role", ca.Config.Role)
	inputState.Set("agent_capabilities", ca.Config.Capabilities)
	inputState.Set("agent_system_prompt", ca.Config.SystemPrompt)
	// Set system_prompt for UnifiedAgent compatibility
	inputState.Set("system_prompt", ca.Config.SystemPrompt)

	// Log agent execution
	core.Logger().Debug().
		Str("agent", ca.AgentName).
		Str("role", ca.Config.Role).
		Strs("capabilities", ca.Config.Capabilities).
		Msg("Executing configured agent")

	// Execute the underlying agent
	outputState, err := ca.Agent.Run(ctx, inputState)
	if err != nil {
		core.Logger().Error().
			Str("agent", ca.AgentName).
			Err(err).
			Msg("Agent execution failed")
		return outputState, err
	}

	// Add execution metadata to output state
	outputState.Set("executed_by", ca.AgentName)
	outputState.Set("execution_role", ca.Config.Role)

	core.Logger().Debug().
		Str("agent", ca.AgentName).
		Msg("Agent execution completed successfully")

	return outputState, nil
}

// Name implements the Agent interface
func (ca *ConfiguredAgent) Name() string {
	return ca.AgentName
}

// GetRole returns the agent's configured role
func (ca *ConfiguredAgent) GetRole() string {
	return ca.Config.Role
}

// GetCapabilities returns the agent's configured capabilities
func (ca *ConfiguredAgent) GetCapabilities() []string {
	return ca.Config.Capabilities
}

// GetSystemPrompt returns the agent's configured system prompt
func (ca *ConfiguredAgent) GetSystemPrompt() string {
	return ca.Config.SystemPrompt
}

// GetTimeout returns the agent's configured timeout
func (ca *ConfiguredAgent) GetTimeout() time.Duration {
	return ca.Config.Timeout
}

// IsEnabled returns whether the agent is enabled
func (ca *ConfiguredAgent) IsEnabled() bool {
	return ca.Config.Enabled
}

// GetLLMConfig returns the agent's resolved LLM configuration
func (ca *ConfiguredAgent) GetLLMConfig() *core.ResolvedLLMConfig {
	return ca.Config.LLMConfig
}

// GetDescription returns the agent's description
func (ca *ConfiguredAgent) GetDescription() string {
	return ca.Config.Description
}

// HandleEvent implements the event-driven pattern by delegating to Run
func (ca *ConfiguredAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
	start := time.Now()
	out, err := ca.Run(ctx, state)
	end := time.Now()
	res := core.AgentResult{OutputState: out, StartTime: start, EndTime: end, Duration: end.Sub(start)}
	if err != nil {
		res.Error = err.Error()
	}
	return res, err
}

// Initialize delegates to underlying agent if it implements Initialize
func (ca *ConfiguredAgent) Initialize(ctx context.Context) error {
	return ca.Agent.Initialize(ctx)
}

// Shutdown delegates to underlying agent if it implements Shutdown
func (ca *ConfiguredAgent) Shutdown(ctx context.Context) error {
	return ca.Agent.Shutdown(ctx)
}

// Register this factory as the enhanced implementation for core
func init() {
	core.RegisterConfigurableAgentFactory(func(cfg *core.Config) core.ConfigurableAgentFactory {
		return NewConfigurableAgentFactory(cfg)
	})
}
