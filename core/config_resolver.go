package core

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// ConfigResolver handles configuration resolution with environment variable overrides
type ConfigResolver struct {
	config *Config
}

// NewConfigResolver creates a new configuration resolver
func NewConfigResolver(config *Config) *ConfigResolver {
	return &ConfigResolver{
		config: config,
	}
}

// ResolveAgentConfigWithEnv resolves agent configuration with environment variable overrides
func (r *ConfigResolver) ResolveAgentConfigWithEnv(agentName string) (*ResolvedAgentConfig, error) {
	agent, exists := r.config.Agents[agentName]
	if !exists {
		return nil, fmt.Errorf("agent '%s' not found in configuration", agentName)
	}

	// Apply environment variable overrides to agent config
	resolvedAgent := r.applyAgentEnvOverrides(agentName, agent)

	// Resolve LLM configuration with environment overrides
	llmConfig := r.resolveLLMConfigWithEnv(agentName, &resolvedAgent)

	// Create resolved configuration
	resolved := &ResolvedAgentConfig{
		Name:         agentName,
		Role:         resolvedAgent.Role,
		Description:  resolvedAgent.Description,
		SystemPrompt: resolvedAgent.SystemPrompt,
		Capabilities: resolvedAgent.Capabilities,
		Enabled:      resolvedAgent.Enabled,
		LLMConfig:    llmConfig,
		RetryPolicy:  resolvedAgent.RetryPolicy,
		RateLimit:    resolvedAgent.RateLimit,
		Timeout:      time.Duration(resolvedAgent.Timeout) * time.Second,
	}

	return resolved, nil
}

// applyAgentEnvOverrides applies environment variable overrides to agent configuration
func (r *ConfigResolver) applyAgentEnvOverrides(agentName string, agent AgentConfig) AgentConfig {
	// Create a copy to avoid modifying the original
	resolved := agent

	// Environment variable patterns:
	// AGENTFLOW_AGENT_{AGENT_NAME}_{FIELD}
	// AGENTFLOW_AGENTS_{AGENT_NAME}_{FIELD}
	agentNameUpper := strings.ToUpper(agentName)
	
	// Override role
	if envRole := r.getEnvVar(fmt.Sprintf("AGENTFLOW_AGENT_%s_ROLE", agentNameUpper)); envRole != "" {
		resolved.Role = envRole
		Logger().Info().
			Str("agent", agentName).
			Str("field", "role").
			Str("value", envRole).
			Msg("Applied environment override")
	}

	// Override description
	if envDesc := r.getEnvVar(fmt.Sprintf("AGENTFLOW_AGENT_%s_DESCRIPTION", agentNameUpper)); envDesc != "" {
		resolved.Description = envDesc
		Logger().Info().
			Str("agent", agentName).
			Str("field", "description").
			Str("value", envDesc).
			Msg("Applied environment override")
	}

	// Override system prompt
	if envPrompt := r.getEnvVar(fmt.Sprintf("AGENTFLOW_AGENT_%s_SYSTEM_PROMPT", agentNameUpper)); envPrompt != "" {
		resolved.SystemPrompt = envPrompt
		Logger().Info().
			Str("agent", agentName).
			Str("field", "system_prompt").
			Str("value", envPrompt).
			Msg("Applied environment override")
	}

	// Override capabilities (comma-separated)
	if envCaps := r.getEnvVar(fmt.Sprintf("AGENTFLOW_AGENT_%s_CAPABILITIES", agentNameUpper)); envCaps != "" {
		capabilities := strings.Split(envCaps, ",")
		for i, cap := range capabilities {
			capabilities[i] = strings.TrimSpace(cap)
		}
		resolved.Capabilities = capabilities
		Logger().Info().
			Str("agent", agentName).
			Str("field", "capabilities").
			Strs("value", capabilities).
			Msg("Applied environment override")
	}

	// Override enabled status
	if envEnabled := r.getEnvVar(fmt.Sprintf("AGENTFLOW_AGENT_%s_ENABLED", agentNameUpper)); envEnabled != "" {
		if enabled, err := strconv.ParseBool(envEnabled); err == nil {
			resolved.Enabled = enabled
			Logger().Info().
				Str("agent", agentName).
				Str("field", "enabled").
				Bool("value", enabled).
				Msg("Applied environment override")
		} else {
			Logger().Warn().
				Str("agent", agentName).
				Str("field", "enabled").
				Str("value", envEnabled).
				Err(err).
				Msg("Invalid boolean value for environment override")
		}
	}

	// Override timeout
	if envTimeout := r.getEnvVar(fmt.Sprintf("AGENTFLOW_AGENT_%s_TIMEOUT_SECONDS", agentNameUpper)); envTimeout != "" {
		if timeout, err := strconv.Atoi(envTimeout); err == nil {
			resolved.Timeout = timeout
			Logger().Info().
				Str("agent", agentName).
				Str("field", "timeout_seconds").
				Int("value", timeout).
				Msg("Applied environment override")
		} else {
			Logger().Warn().
				Str("agent", agentName).
				Str("field", "timeout_seconds").
				Str("value", envTimeout).
				Err(err).
				Msg("Invalid integer value for environment override")
		}
	}

	return resolved
}

// resolveLLMConfigWithEnv resolves LLM configuration with inheritance and environment overrides
func (r *ConfigResolver) resolveLLMConfigWithEnv(agentName string, agent *AgentConfig) *ResolvedLLMConfig {
	// Start with global LLM config
	resolved := &ResolvedLLMConfig{
		Provider:         r.config.LLM.Provider,
		Model:            r.config.LLM.Model,
		Temperature:      r.config.LLM.Temperature,
		MaxTokens:        r.config.LLM.MaxTokens,
		Timeout:          time.Duration(r.config.LLM.TimeoutSeconds) * time.Second,
		TopP:             r.config.LLM.TopP,
		FrequencyPenalty: r.config.LLM.FrequencyPenalty,
		PresencePenalty:  r.config.LLM.PresencePenalty,
	}

	// Apply global environment overrides first (lower priority)
	agentNameUpper := strings.ToUpper(agentName)

	// Global LLM environment overrides
	if envProvider := r.getEnvVar("AGENTFLOW_LLM_PROVIDER"); envProvider != "" {
		resolved.Provider = envProvider
		Logger().Info().
			Str("field", "llm.provider").
			Str("value", envProvider).
			Msg("Applied global LLM environment override")
	}

	if envModel := r.getEnvVar("AGENTFLOW_LLM_MODEL"); envModel != "" {
		resolved.Model = envModel
		Logger().Info().
			Str("field", "llm.model").
			Str("value", envModel).
			Msg("Applied global LLM environment override")
	}

	if envTemp := r.getEnvVar("AGENTFLOW_LLM_TEMPERATURE"); envTemp != "" {
		if temp, err := strconv.ParseFloat(envTemp, 64); err == nil {
			resolved.Temperature = temp
			Logger().Info().
				Str("field", "llm.temperature").
				Float64("value", temp).
				Msg("Applied global LLM environment override")
		}
	}

	if envTokens := r.getEnvVar("AGENTFLOW_LLM_MAX_TOKENS"); envTokens != "" {
		if tokens, err := strconv.Atoi(envTokens); err == nil {
			resolved.MaxTokens = tokens
			Logger().Info().
				Str("field", "llm.max_tokens").
				Int("value", tokens).
				Msg("Applied global LLM environment override")
		}
	}

	// Override with agent-specific LLM config if provided (higher priority than global env)
	if agent.LLM != nil {
		if agent.LLM.Provider != "" {
			resolved.Provider = agent.LLM.Provider
		}
		if agent.LLM.Model != "" {
			resolved.Model = agent.LLM.Model
		}
		if agent.LLM.Temperature != 0 {
			resolved.Temperature = agent.LLM.Temperature
		}
		if agent.LLM.MaxTokens != 0 {
			resolved.MaxTokens = agent.LLM.MaxTokens
		}
		if agent.LLM.TimeoutSeconds != 0 {
			resolved.Timeout = time.Duration(agent.LLM.TimeoutSeconds) * time.Second
		}
		if agent.LLM.TopP != 0 {
			resolved.TopP = agent.LLM.TopP
		}
		if agent.LLM.FrequencyPenalty != 0 {
			resolved.FrequencyPenalty = agent.LLM.FrequencyPenalty
		}
		if agent.LLM.PresencePenalty != 0 {
			resolved.PresencePenalty = agent.LLM.PresencePenalty
		}
	}

	// Agent-specific LLM environment overrides (highest priority)
	if envProvider := r.getEnvVar(fmt.Sprintf("AGENTFLOW_AGENT_%s_LLM_PROVIDER", agentNameUpper)); envProvider != "" {
		resolved.Provider = envProvider
		Logger().Info().
			Str("agent", agentName).
			Str("field", "llm.provider").
			Str("value", envProvider).
			Msg("Applied agent-specific LLM environment override")
	}

	if envModel := r.getEnvVar(fmt.Sprintf("AGENTFLOW_AGENT_%s_LLM_MODEL", agentNameUpper)); envModel != "" {
		resolved.Model = envModel
		Logger().Info().
			Str("agent", agentName).
			Str("field", "llm.model").
			Str("value", envModel).
			Msg("Applied agent-specific LLM environment override")
	}

	if envTemp := r.getEnvVar(fmt.Sprintf("AGENTFLOW_AGENT_%s_LLM_TEMPERATURE", agentNameUpper)); envTemp != "" {
		if temp, err := strconv.ParseFloat(envTemp, 64); err == nil {
			resolved.Temperature = temp
			Logger().Info().
				Str("agent", agentName).
				Str("field", "llm.temperature").
				Float64("value", temp).
				Msg("Applied agent-specific LLM environment override")
		}
	}

	if envTokens := r.getEnvVar(fmt.Sprintf("AGENTFLOW_AGENT_%s_LLM_MAX_TOKENS", agentNameUpper)); envTokens != "" {
		if tokens, err := strconv.Atoi(envTokens); err == nil {
			resolved.MaxTokens = tokens
			Logger().Info().
				Str("agent", agentName).
				Str("field", "llm.max_tokens").
				Int("value", tokens).
				Msg("Applied agent-specific LLM environment override")
		}
	}

	if envTimeout := r.getEnvVar(fmt.Sprintf("AGENTFLOW_AGENT_%s_LLM_TIMEOUT_SECONDS", agentNameUpper)); envTimeout != "" {
		if timeout, err := strconv.Atoi(envTimeout); err == nil {
			resolved.Timeout = time.Duration(timeout) * time.Second
			Logger().Info().
				Str("agent", agentName).
				Str("field", "llm.timeout_seconds").
				Int("value", timeout).
				Msg("Applied agent-specific LLM environment override")
		}
	}

	return resolved
}

// ApplyEnvironmentOverrides applies environment variable overrides to the entire configuration
func (r *ConfigResolver) ApplyEnvironmentOverrides() error {
	// Apply global configuration overrides
	if err := r.applyGlobalEnvOverrides(); err != nil {
		return fmt.Errorf("failed to apply global environment overrides: %w", err)
	}

	// Apply agent-specific overrides
	for agentName := range r.config.Agents {
		agent := r.config.Agents[agentName]
		resolvedAgent := r.applyAgentEnvOverrides(agentName, agent)
		r.config.Agents[agentName] = resolvedAgent
	}

	return nil
}

// applyGlobalEnvOverrides applies environment overrides to global configuration
func (r *ConfigResolver) applyGlobalEnvOverrides() error {
	// Override global LLM configuration
	if envProvider := r.getEnvVar("AGENTFLOW_LLM_PROVIDER"); envProvider != "" {
		r.config.LLM.Provider = envProvider
		Logger().Info().
			Str("field", "llm.provider").
			Str("value", envProvider).
			Msg("Applied global environment override")
	}

	if envModel := r.getEnvVar("AGENTFLOW_LLM_MODEL"); envModel != "" {
		r.config.LLM.Model = envModel
		Logger().Info().
			Str("field", "llm.model").
			Str("value", envModel).
			Msg("Applied global environment override")
	}

	if envTemp := r.getEnvVar("AGENTFLOW_LLM_TEMPERATURE"); envTemp != "" {
		if temp, err := strconv.ParseFloat(envTemp, 64); err == nil {
			r.config.LLM.Temperature = temp
			Logger().Info().
				Str("field", "llm.temperature").
				Float64("value", temp).
				Msg("Applied global environment override")
		} else {
			return fmt.Errorf("invalid temperature value in AGENTFLOW_LLM_TEMPERATURE: %s", envTemp)
		}
	}

	if envTokens := r.getEnvVar("AGENTFLOW_LLM_MAX_TOKENS"); envTokens != "" {
		if tokens, err := strconv.Atoi(envTokens); err == nil {
			r.config.LLM.MaxTokens = tokens
			Logger().Info().
				Str("field", "llm.max_tokens").
				Int("value", tokens).
				Msg("Applied global environment override")
		} else {
			return fmt.Errorf("invalid max_tokens value in AGENTFLOW_LLM_MAX_TOKENS: %s", envTokens)
		}
	}

	if envTimeout := r.getEnvVar("AGENTFLOW_LLM_TIMEOUT_SECONDS"); envTimeout != "" {
		if timeout, err := strconv.Atoi(envTimeout); err == nil {
			r.config.LLM.TimeoutSeconds = timeout
			Logger().Info().
				Str("field", "llm.timeout_seconds").
				Int("value", timeout).
				Msg("Applied global environment override")
		} else {
			return fmt.Errorf("invalid timeout_seconds value in AGENTFLOW_LLM_TIMEOUT_SECONDS: %s", envTimeout)
		}
	}

	return nil
}

// getEnvVar gets an environment variable with logging
func (r *ConfigResolver) getEnvVar(key string) string {
	value := os.Getenv(key)
	if value != "" {
		Logger().Debug().
			Str("env_var", key).
			Str("value", value).
			Msg("Found environment variable")
	}
	return value
}

// GetResolvedConfig returns the configuration with all overrides applied
func (r *ConfigResolver) GetResolvedConfig() *Config {
	return r.config
}

// ResolveAllAgents resolves all agent configurations with environment overrides
func (r *ConfigResolver) ResolveAllAgents() (map[string]*ResolvedAgentConfig, error) {
	resolved := make(map[string]*ResolvedAgentConfig)
	
	for agentName := range r.config.Agents {
		agentConfig, err := r.ResolveAgentConfigWithEnv(agentName)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve agent '%s': %w", agentName, err)
		}
		resolved[agentName] = agentConfig
	}
	
	return resolved, nil
}

// ValidateResolvedConfig validates the configuration after environment overrides
func (r *ConfigResolver) ValidateResolvedConfig() []ValidationError {
	validator := NewDefaultConfigValidator()
	return validator.ValidateConfig(r.config)
}