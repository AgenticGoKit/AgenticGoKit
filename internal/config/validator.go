// Package config provides internal configuration validation functionality.
package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/agenticgokit/agenticgokit/core"
)

// DefaultConfigValidator implements core.ConfigValidator
type DefaultConfigValidator struct {
	knownCapabilities map[string]bool
	validProviders    map[string]bool
	providerModels    map[string][]string
	capabilityGroups  map[string][]string
}

// NewDefaultConfigValidator creates a new default configuration validator
func NewDefaultConfigValidator() *DefaultConfigValidator {
	return &DefaultConfigValidator{
		knownCapabilities: map[string]bool{
			"information_gathering":  true,
			"fact_checking":         true,
			"source_identification": true,
			"pattern_recognition":   true,
			"trend_analysis":        true,
			"insight_generation":    true,
			"data_processing":       true,
			"text_analysis":         true,
			"summarization":         true,
			"translation":           true,
			"code_generation":       true,
			"code_review":           true,
			"debugging":             true,
			"testing":               true,
			"documentation":         true,
			"content_creation":      true,
			"editing":               true,
			"data_analysis":         true,
			"research":              true,
			"writing":               true,
			"analysis":              true,
		},
		validProviders: map[string]bool{
			"openai":    true,
			"azure":     true,
			"ollama":    true,
			"anthropic": true,
			"google":    true,
			"cohere":    true,
		},
		providerModels: map[string][]string{
			"openai": {
				"gpt-4", "gpt-4-turbo", "gpt-4o", "gpt-4o-mini",
				"gpt-3.5-turbo", "gpt-3.5-turbo-16k",
			},
			"azure": {
				"gpt-4", "gpt-4-turbo", "gpt-35-turbo", "gpt-35-turbo-16k",
			},
			"anthropic": {
				"claude-3-opus", "claude-3-sonnet", "claude-3-haiku",
				"claude-2.1", "claude-2.0", "claude-instant-1.2",
			},
			"google": {
				"gemini-pro", "gemini-pro-vision", "gemini-1.5-pro",
				"gemini-1.5-flash", "text-bison", "chat-bison",
			},
			"cohere": {
				"command", "command-light", "command-nightly",
			},
			"ollama": {
				"llama2", "llama2:13b", "llama2:70b", "codellama",
				"mistral", "mixtral", "phi", "gemma",
			},
		},
		capabilityGroups: map[string][]string{
			"research": {
				"information_gathering", "fact_checking", "source_identification",
				"data_processing", "text_analysis",
			},
			"analysis": {
				"pattern_recognition", "trend_analysis", "insight_generation",
				"data_analysis", "text_analysis",
			},
			"content": {
				"content_creation", "editing", "summarization", "translation",
				"documentation", "writing",
			},
			"development": {
				"code_generation", "code_review", "debugging", "testing",
				"documentation",
			},
		},
	}
}

// ValidateConfig validates the entire configuration
func (v *DefaultConfigValidator) ValidateConfig(config *core.Config) []core.ValidationError {
	var errors []core.ValidationError

	// 1. Validate configuration completeness
	completenessErrors := v.ValidateConfigCompleteness(config)
	errors = append(errors, completenessErrors...)

	// 2. Validate global LLM configuration
	llmErrors := v.ValidateLLMConfig(&config.LLM)
	for _, err := range llmErrors {
		err.Field = "llm." + err.Field
		errors = append(errors, err)
	}

	// 3. Validate each agent configuration
	for name, agent := range config.Agents {
		// Basic agent validation
		agentErrors := v.ValidateAgentConfig(name, &agent)
		for _, err := range agentErrors {
			err.Field = fmt.Sprintf("agents.%s.%s", name, err.Field)
			errors = append(errors, err)
		}

		// Agent naming validation
		namingErrors := v.ValidateAgentNaming(name, &agent)
		for _, err := range namingErrors {
			err.Field = fmt.Sprintf("agents.%s.%s", name, err.Field)
			errors = append(errors, err)
		}

		// Capability group validation
		capGroupErrors := v.ValidateCapabilityGroups(agent.Capabilities)
		for _, err := range capGroupErrors {
			err.Field = fmt.Sprintf("agents.%s.%s", name, err.Field)
			errors = append(errors, err)
		}
	}

	// 4. Validate orchestration configuration against agents
	orchErrors := v.ValidateOrchestrationAgents(&config.Orchestration, config.Agents)
	for _, err := range orchErrors {
		err.Field = "orchestration." + err.Field
		errors = append(errors, err)
	}

	// 5. Cross-validation checks
	crossValidationErrors := v.validateCrossReferences(config)
	errors = append(errors, crossValidationErrors...)

	return errors
}

// ValidateAgentConfig validates agent-specific configuration
func (v *DefaultConfigValidator) ValidateAgentConfig(name string, config *core.AgentConfig) []core.ValidationError {
	var errors []core.ValidationError

	// Validate required fields
	if config.Role == "" {
		errors = append(errors, core.ValidationError{
			Field:      "role",
			Value:      config.Role,
			Message:    "role is required",
			Suggestion: fmt.Sprintf("set role to '%s_agent' or a descriptive role name", name),
		})
	}

	// Validate role format (should be lowercase with underscores)
	if config.Role != "" {
		rolePattern := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
		if !rolePattern.MatchString(config.Role) {
			errors = append(errors, core.ValidationError{
				Field:      "role",
				Value:      config.Role,
				Message:    "role must be lowercase with underscores only",
				Suggestion: "use format like 'research_agent' or 'data_processor'",
			})
		}
	}

	// Validate description
	if config.Description == "" {
		errors = append(errors, core.ValidationError{
			Field:      "description",
			Value:      config.Description,
			Message:    "description is recommended for documentation",
			Suggestion: "provide a brief description of the agent's purpose",
		})
	}

	// Validate system prompt
	if config.SystemPrompt == "" {
		errors = append(errors, core.ValidationError{
			Field:      "system_prompt",
			Value:      config.SystemPrompt,
			Message:    "system_prompt is required for agent behavior",
			Suggestion: "provide a clear system prompt defining the agent's role and behavior",
		})
	} else if len(config.SystemPrompt) < 10 {
		errors = append(errors, core.ValidationError{
			Field:      "system_prompt",
			Value:      config.SystemPrompt,
			Message:    "system_prompt is too short",
			Suggestion: "provide a more detailed system prompt (at least 10 characters)",
		})
	}

	// Validate capabilities
	capErrors := v.ValidateCapabilities(config.Capabilities)
	errors = append(errors, capErrors...)

	// Validate timeout
	if config.Timeout < 0 {
		errors = append(errors, core.ValidationError{
			Field:      "timeout_seconds",
			Value:      config.Timeout,
			Message:    "timeout cannot be negative",
			Suggestion: "set timeout to a positive value (e.g., 30 seconds)",
		})
	} else if config.Timeout > 300 {
		errors = append(errors, core.ValidationError{
			Field:      "timeout_seconds",
			Value:      config.Timeout,
			Message:    "timeout is very high (>5 minutes)",
			Suggestion: "consider reducing timeout to avoid long waits",
		})
	}

	// Validate LLM configuration if provided
	if config.LLM != nil {
		llmErrors := v.ValidateLLMConfig(config.LLM)
		for _, err := range llmErrors {
			err.Field = "llm." + err.Field
			errors = append(errors, err)
		}
	}

	// Validate retry policy if provided
	if config.RetryPolicy != nil {
		retryErrors := v.validateRetryPolicy(config.RetryPolicy)
		for _, err := range retryErrors {
			err.Field = "retry_policy." + err.Field
			errors = append(errors, err)
		}
	}

	// Validate rate limit if provided
	if config.RateLimit != nil {
		rateLimitErrors := v.validateRateLimit(config.RateLimit)
		for _, err := range rateLimitErrors {
			err.Field = "rate_limit." + err.Field
			errors = append(errors, err)
		}
	}

	return errors
}

// ValidateLLMConfig validates LLM configuration
func (v *DefaultConfigValidator) ValidateLLMConfig(config *core.AgentLLMConfig) []core.ValidationError {
	var errors []core.ValidationError

	// Validate provider
	if config.Provider != "" && !v.validProviders[config.Provider] {
		validProviders := make([]string, 0, len(v.validProviders))
		for provider := range v.validProviders {
			validProviders = append(validProviders, provider)
		}
		errors = append(errors, core.ValidationError{
			Field:      "provider",
			Value:      config.Provider,
			Message:    "unsupported LLM provider",
			Suggestion: fmt.Sprintf("use one of: %s", strings.Join(validProviders, ", ")),
		})
	}

	// Validate model for specific providers
	if config.Provider != "" && config.Model != "" {
		if models, exists := v.providerModels[config.Provider]; exists {
			validModel := false
			for _, model := range models {
				if config.Model == model {
					validModel = true
					break
				}
			}
			if !validModel {
				errors = append(errors, core.ValidationError{
					Field:      "model",
					Value:      config.Model,
					Message:    fmt.Sprintf("unsupported model for provider '%s'", config.Provider),
					Suggestion: fmt.Sprintf("use one of: %s", strings.Join(models, ", ")),
				})
			}
		}
	}

	// Validate temperature
	if config.Temperature < 0 || config.Temperature > 2 {
		errors = append(errors, core.ValidationError{
			Field:      "temperature",
			Value:      config.Temperature,
			Message:    "temperature must be between 0 and 2",
			Suggestion: "use 0.1-0.3 for factual tasks, 0.7-1.0 for creative tasks",
		})
	}

	// Validate max tokens
	if config.MaxTokens < 1 {
		errors = append(errors, core.ValidationError{
			Field:      "max_tokens",
			Value:      config.MaxTokens,
			Message:    "max_tokens must be positive",
			Suggestion: "set max_tokens to a reasonable value (e.g., 800-2000)",
		})
	} else if config.MaxTokens > 32000 {
		errors = append(errors, core.ValidationError{
			Field:      "max_tokens",
			Value:      config.MaxTokens,
			Message:    "max_tokens is very high",
			Suggestion: "consider reducing max_tokens to avoid high costs",
		})
	}

	// Validate timeout
	if config.TimeoutSeconds < 0 {
		errors = append(errors, core.ValidationError{
			Field:      "timeout_seconds",
			Value:      config.TimeoutSeconds,
			Message:    "timeout_seconds cannot be negative",
			Suggestion: "set timeout_seconds to a positive value (e.g., 30)",
		})
	} else if config.TimeoutSeconds > 300 {
		errors = append(errors, core.ValidationError{
			Field:      "timeout_seconds",
			Value:      config.TimeoutSeconds,
			Message:    "timeout_seconds is very high (>5 minutes)",
			Suggestion: "consider reducing timeout to avoid long waits",
		})
	}

	// Validate TopP
	if config.TopP != 0 && (config.TopP < 0 || config.TopP > 1) {
		errors = append(errors, core.ValidationError{
			Field:      "top_p",
			Value:      config.TopP,
			Message:    "top_p must be between 0 and 1",
			Suggestion: "use values like 0.9 for diverse output, 0.1 for focused output",
		})
	}

	// Validate frequency penalty
	if config.FrequencyPenalty < -2 || config.FrequencyPenalty > 2 {
		errors = append(errors, core.ValidationError{
			Field:      "frequency_penalty",
			Value:      config.FrequencyPenalty,
			Message:    "frequency_penalty must be between -2 and 2",
			Suggestion: "use positive values to reduce repetition, negative to encourage it",
		})
	}

	// Validate presence penalty
	if config.PresencePenalty < -2 || config.PresencePenalty > 2 {
		errors = append(errors, core.ValidationError{
			Field:      "presence_penalty",
			Value:      config.PresencePenalty,
			Message:    "presence_penalty must be between -2 and 2",
			Suggestion: "use positive values to encourage new topics, negative to stay on topic",
		})
	}

	// Provider-specific validation
	providerErrors := v.validateProviderSpecificParams(config)
	errors = append(errors, providerErrors...)

	return errors
}

// ValidateCapabilities validates agent capabilities
func (v *DefaultConfigValidator) ValidateCapabilities(capabilities []string) []core.ValidationError {
	var errors []core.ValidationError

	if len(capabilities) == 0 {
		errors = append(errors, core.ValidationError{
			Field:      "capabilities",
			Value:      capabilities,
			Message:    "at least one capability should be specified",
			Suggestion: "add capabilities like 'information_gathering', 'text_analysis', etc.",
		})
	}

	// Check for unknown capabilities
	var unknownCaps []string
	for _, cap := range capabilities {
		if !v.knownCapabilities[cap] {
			unknownCaps = append(unknownCaps, cap)
		}
	}

	if len(unknownCaps) > 0 {
		knownCaps := make([]string, 0, len(v.knownCapabilities))
		for cap := range v.knownCapabilities {
			knownCaps = append(knownCaps, cap)
		}
		errors = append(errors, core.ValidationError{
			Field:      "capabilities",
			Value:      unknownCaps,
			Message:    fmt.Sprintf("unknown capabilities: %s", strings.Join(unknownCaps, ", ")),
			Suggestion: fmt.Sprintf("use known capabilities: %s", strings.Join(knownCaps[:5], ", ")) + "...",
		})
	}

	// Check for duplicates
	seen := make(map[string]bool)
	var duplicates []string
	for _, cap := range capabilities {
		if seen[cap] {
			duplicates = append(duplicates, cap)
		}
		seen[cap] = true
	}

	if len(duplicates) > 0 {
		errors = append(errors, core.ValidationError{
			Field:      "capabilities",
			Value:      duplicates,
			Message:    fmt.Sprintf("duplicate capabilities: %s", strings.Join(duplicates, ", ")),
			Suggestion: "remove duplicate capability entries",
		})
	}

	return errors
}

// ValidateOrchestrationAgents validates orchestration configuration against available agents
func (v *DefaultConfigValidator) ValidateOrchestrationAgents(orchestration *core.OrchestrationConfigToml, agents map[string]core.AgentConfig) []core.ValidationError {
	var errors []core.ValidationError

	// Validate sequential agents exist
	for _, agentName := range orchestration.SequentialAgents {
		if _, exists := agents[agentName]; !exists {
			errors = append(errors, core.ValidationError{
				Field:      "sequential_agents",
				Value:      agentName,
				Message:    fmt.Sprintf("agent '%s' not found in configuration", agentName),
				Suggestion: "ensure all referenced agents are defined in the agents section",
			})
		} else if !agents[agentName].Enabled {
			errors = append(errors, core.ValidationError{
				Field:      "sequential_agents",
				Value:      agentName,
				Message:    fmt.Sprintf("agent '%s' is disabled", agentName),
				Suggestion: "enable the agent or remove it from sequential_agents",
			})
		}
	}

	// Validate collaborative agents exist
	for _, agentName := range orchestration.CollaborativeAgents {
		if _, exists := agents[agentName]; !exists {
			errors = append(errors, core.ValidationError{
				Field:      "collaborative_agents",
				Value:      agentName,
				Message:    fmt.Sprintf("agent '%s' not found in configuration", agentName),
				Suggestion: "ensure all referenced agents are defined in the agents section",
			})
		} else if !agents[agentName].Enabled {
			errors = append(errors, core.ValidationError{
				Field:      "collaborative_agents",
				Value:      agentName,
				Message:    fmt.Sprintf("agent '%s' is disabled", agentName),
				Suggestion: "enable the agent or remove it from collaborative_agents",
			})
		}
	}

	// Validate loop agent exists
	if orchestration.LoopAgent != "" {
		if _, exists := agents[orchestration.LoopAgent]; !exists {
			errors = append(errors, core.ValidationError{
				Field:      "loop_agent",
				Value:      orchestration.LoopAgent,
				Message:    fmt.Sprintf("agent '%s' not found in configuration", orchestration.LoopAgent),
				Suggestion: "ensure the loop agent is defined in the agents section",
			})
		} else if !agents[orchestration.LoopAgent].Enabled {
			errors = append(errors, core.ValidationError{
				Field:      "loop_agent",
				Value:      orchestration.LoopAgent,
				Message:    fmt.Sprintf("agent '%s' is disabled", orchestration.LoopAgent),
				Suggestion: "enable the agent or choose a different loop agent",
			})
		}
	}

	return errors
}

// validateRetryPolicy validates retry policy configuration
func (v *DefaultConfigValidator) validateRetryPolicy(policy *core.AgentRetryPolicyConfig) []core.ValidationError {
	var errors []core.ValidationError

	if policy.MaxRetries < 0 {
		errors = append(errors, core.ValidationError{
			Field:      "max_retries",
			Value:      policy.MaxRetries,
			Message:    "max_retries cannot be negative",
			Suggestion: "set max_retries to 0 to disable retries or a positive value",
		})
	} else if policy.MaxRetries > 10 {
		errors = append(errors, core.ValidationError{
			Field:      "max_retries",
			Value:      policy.MaxRetries,
			Message:    "max_retries is very high",
			Suggestion: "consider reducing max_retries to avoid excessive delays",
		})
	}

	if policy.BaseDelayMs < 0 {
		errors = append(errors, core.ValidationError{
			Field:      "base_delay_ms",
			Value:      policy.BaseDelayMs,
			Message:    "base_delay_ms cannot be negative",
			Suggestion: "set base_delay_ms to a positive value (e.g., 1000)",
		})
	}

	if policy.MaxDelayMs < policy.BaseDelayMs {
		errors = append(errors, core.ValidationError{
			Field:      "max_delay_ms",
			Value:      policy.MaxDelayMs,
			Message:    "max_delay_ms must be greater than or equal to base_delay_ms",
			Suggestion: "increase max_delay_ms or decrease base_delay_ms",
		})
	}

	if policy.BackoffFactor <= 0 {
		errors = append(errors, core.ValidationError{
			Field:      "backoff_factor",
			Value:      policy.BackoffFactor,
			Message:    "backoff_factor must be positive",
			Suggestion: "set backoff_factor to a value like 2.0 for exponential backoff",
		})
	}

	return errors
}

// validateRateLimit validates rate limit configuration
func (v *DefaultConfigValidator) validateRateLimit(rateLimit *core.RateLimitConfig) []core.ValidationError {
	var errors []core.ValidationError

	if rateLimit.RequestsPerSecond <= 0 {
		errors = append(errors, core.ValidationError{
			Field:      "requests_per_second",
			Value:      rateLimit.RequestsPerSecond,
			Message:    "requests_per_second must be positive",
			Suggestion: "set requests_per_second to a reasonable value (e.g., 10)",
		})
	}

	if rateLimit.BurstSize < 0 {
		errors = append(errors, core.ValidationError{
			Field:      "burst_size",
			Value:      rateLimit.BurstSize,
			Message:    "burst_size cannot be negative",
			Suggestion: "set burst_size to 0 for no burst or a positive value",
		})
	}

	return errors
}

// AddKnownCapability adds a capability to the known capabilities list
func (v *DefaultConfigValidator) AddKnownCapability(capability string) {
	v.knownCapabilities[capability] = true
}

// AddValidProvider adds a provider to the valid providers list
func (v *DefaultConfigValidator) AddValidProvider(provider string) {
	v.validProviders[provider] = true
}

// validateProviderSpecificParams validates parameters specific to each provider
func (v *DefaultConfigValidator) validateProviderSpecificParams(config *core.AgentLLMConfig) []core.ValidationError {
	var errors []core.ValidationError

	switch config.Provider {
	case "openai":
		errors = append(errors, v.validateOpenAIParams(config)...)
	case "anthropic":
		errors = append(errors, v.validateAnthropicParams(config)...)
	case "azure":
		errors = append(errors, v.validateAzureParams(config)...)
	case "ollama":
		errors = append(errors, v.validateOllamaParams(config)...)
	case "google":
		errors = append(errors, v.validateGoogleParams(config)...)
	case "cohere":
		errors = append(errors, v.validateCohereParams(config)...)
	}

	return errors
}

// validateOpenAIParams validates OpenAI-specific parameters
func (v *DefaultConfigValidator) validateOpenAIParams(config *core.AgentLLMConfig) []core.ValidationError {
	var errors []core.ValidationError

	// OpenAI supports all standard parameters
	// Validate max tokens based on model
	if strings.HasPrefix(config.Model, "gpt-4") {
		if config.MaxTokens > 8192 && !strings.Contains(config.Model, "turbo") {
			errors = append(errors, core.ValidationError{
				Field:      "max_tokens",
				Value:      config.MaxTokens,
				Message:    "max_tokens exceeds model limit for GPT-4",
				Suggestion: "use max_tokens <= 8192 for GPT-4 or use gpt-4-turbo for higher limits",
			})
		}
	} else if strings.HasPrefix(config.Model, "gpt-3.5") {
		if config.MaxTokens > 4096 && !strings.Contains(config.Model, "16k") {
			errors = append(errors, core.ValidationError{
				Field:      "max_tokens",
				Value:      config.MaxTokens,
				Message:    "max_tokens exceeds model limit for GPT-3.5",
				Suggestion: "use max_tokens <= 4096 for GPT-3.5 or use gpt-3.5-turbo-16k",
			})
		}
	}

	return errors
}

// validateAnthropicParams validates Anthropic-specific parameters
func (v *DefaultConfigValidator) validateAnthropicParams(config *core.AgentLLMConfig) []core.ValidationError {
	var errors []core.ValidationError

	// Anthropic doesn't support frequency_penalty and presence_penalty
	if config.FrequencyPenalty != 0 {
		errors = append(errors, core.ValidationError{
			Field:      "frequency_penalty",
			Value:      config.FrequencyPenalty,
			Message:    "frequency_penalty is not supported by Anthropic",
			Suggestion: "remove frequency_penalty or use a different provider",
		})
	}

	if config.PresencePenalty != 0 {
		errors = append(errors, core.ValidationError{
			Field:      "presence_penalty",
			Value:      config.PresencePenalty,
			Message:    "presence_penalty is not supported by Anthropic",
			Suggestion: "remove presence_penalty or use a different provider",
		})
	}

	// Validate max tokens for Claude models
	if strings.HasPrefix(config.Model, "claude-3") {
		if config.MaxTokens > 4096 {
			errors = append(errors, core.ValidationError{
				Field:      "max_tokens",
				Value:      config.MaxTokens,
				Message:    "max_tokens exceeds recommended limit for Claude-3",
				Suggestion: "use max_tokens <= 4096 for optimal performance",
			})
		}
	}

	return errors
}

// validateAzureParams validates Azure OpenAI-specific parameters
func (v *DefaultConfigValidator) validateAzureParams(config *core.AgentLLMConfig) []core.ValidationError {
	var errors []core.ValidationError

	// Azure OpenAI has similar limits to OpenAI but with different model names
	if strings.HasPrefix(config.Model, "gpt-35") {
		if config.MaxTokens > 4096 && !strings.Contains(config.Model, "16k") {
			errors = append(errors, core.ValidationError{
				Field:      "max_tokens",
				Value:      config.MaxTokens,
				Message:    "max_tokens exceeds model limit for GPT-3.5 on Azure",
				Suggestion: "use max_tokens <= 4096 or use gpt-35-turbo-16k",
			})
		}
	}

	return errors
}

// validateOllamaParams validates Ollama-specific parameters
func (v *DefaultConfigValidator) validateOllamaParams(config *core.AgentLLMConfig) []core.ValidationError {
	var errors []core.ValidationError

	// Ollama doesn't support some OpenAI parameters
	if config.FrequencyPenalty != 0 {
		errors = append(errors, core.ValidationError{
			Field:      "frequency_penalty",
			Value:      config.FrequencyPenalty,
			Message:    "frequency_penalty may not be supported by all Ollama models",
			Suggestion: "test with your specific model or remove this parameter",
		})
	}

	if config.PresencePenalty != 0 {
		errors = append(errors, core.ValidationError{
			Field:      "presence_penalty",
			Value:      config.PresencePenalty,
			Message:    "presence_penalty may not be supported by all Ollama models",
			Suggestion: "test with your specific model or remove this parameter",
		})
	}

	// Ollama models typically have lower context limits
	if config.MaxTokens > 2048 {
		errors = append(errors, core.ValidationError{
			Field:      "max_tokens",
			Value:      config.MaxTokens,
			Message:    "max_tokens is high for Ollama models",
			Suggestion: "consider reducing max_tokens for better performance with local models",
		})
	}

	return errors
}

// validateGoogleParams validates Google AI-specific parameters
func (v *DefaultConfigValidator) validateGoogleParams(config *core.AgentLLMConfig) []core.ValidationError {
	var errors []core.ValidationError

	// Google AI has different parameter names and limits
	if config.FrequencyPenalty != 0 {
		errors = append(errors, core.ValidationError{
			Field:      "frequency_penalty",
			Value:      config.FrequencyPenalty,
			Message:    "frequency_penalty is not directly supported by Google AI",
			Suggestion: "remove frequency_penalty or use equivalent Google AI parameters",
		})
	}

	if config.PresencePenalty != 0 {
		errors = append(errors, core.ValidationError{
			Field:      "presence_penalty",
			Value:      config.PresencePenalty,
			Message:    "presence_penalty is not directly supported by Google AI",
			Suggestion: "remove presence_penalty or use equivalent Google AI parameters",
		})
	}

	return errors
}

// validateCohereParams validates Cohere-specific parameters
func (v *DefaultConfigValidator) validateCohereParams(config *core.AgentLLMConfig) []core.ValidationError {
	var errors []core.ValidationError

	// Cohere has different parameter support
	if config.TopP != 0 && config.TopP > 0.99 {
		errors = append(errors, core.ValidationError{
			Field:      "top_p",
			Value:      config.TopP,
			Message:    "top_p should be less than 0.99 for Cohere",
			Suggestion: "use top_p values between 0.1 and 0.99",
		})
	}

	return errors
}

// ValidateConfigCompleteness validates configuration completeness and applies defaults
func (v *DefaultConfigValidator) ValidateConfigCompleteness(config *core.Config) []core.ValidationError {
	var errors []core.ValidationError

	// Check if global LLM configuration is provided
	if config.LLM.Provider == "" {
		errors = append(errors, core.ValidationError{
			Field:      "llm.provider",
			Value:      config.LLM.Provider,
			Message:    "global LLM provider not specified",
			Suggestion: "set a default LLM provider (e.g., 'openai', 'anthropic')",
		})
	}

	if config.LLM.Model == "" {
		errors = append(errors, core.ValidationError{
			Field:      "llm.model",
			Value:      config.LLM.Model,
			Message:    "global LLM model not specified",
			Suggestion: "set a default LLM model (e.g., 'gpt-4', 'claude-3-sonnet')",
		})
	}

	// Check agent configuration completeness
	for name, agent := range config.Agents {
		if agent.Role == "" {
			errors = append(errors, core.ValidationError{
				Field:      fmt.Sprintf("agents.%s.role", name),
				Value:      agent.Role,
				Message:    "agent role not specified",
				Suggestion: fmt.Sprintf("set role to '%s_agent' or a descriptive role", name),
			})
		}

		if agent.SystemPrompt == "" {
			errors = append(errors, core.ValidationError{
				Field:      fmt.Sprintf("agents.%s.system_prompt", name),
				Value:      agent.SystemPrompt,
				Message:    "agent system prompt not specified",
				Suggestion: "provide a clear system prompt defining the agent's behavior",
			})
		}

		if len(agent.Capabilities) == 0 {
			errors = append(errors, core.ValidationError{
				Field:      fmt.Sprintf("agents.%s.capabilities", name),
				Value:      agent.Capabilities,
				Message:    "agent capabilities not specified",
				Suggestion: "add relevant capabilities based on the agent's role",
			})
		}
	}

	// Check orchestration configuration
	if len(config.Orchestration.SequentialAgents) == 0 && 
	   len(config.Orchestration.CollaborativeAgents) == 0 && 
	   config.Orchestration.LoopAgent == "" {
		errors = append(errors, core.ValidationError{
			Field:      "orchestration",
			Value:      nil,
			Message:    "no orchestration configuration specified",
			Suggestion: "configure sequential_agents, collaborative_agents, or loop_agent",
		})
	}

	return errors
}

// ValidateCapabilityGroups validates that agent capabilities make sense together
func (v *DefaultConfigValidator) ValidateCapabilityGroups(capabilities []string) []core.ValidationError {
	var errors []core.ValidationError

	if len(capabilities) == 0 {
		return errors
	}

	// Check for conflicting capability groups
	groupCounts := make(map[string]int)
	for _, cap := range capabilities {
		for group, groupCaps := range v.capabilityGroups {
			for _, groupCap := range groupCaps {
				if cap == groupCap {
					groupCounts[group]++
					break
				}
			}
		}
	}

	// Suggest capability groups if agent has mixed capabilities
	if len(groupCounts) > 2 {
		var groups []string
		for group := range groupCounts {
			groups = append(groups, group)
		}
		errors = append(errors, core.ValidationError{
			Field:      "capabilities",
			Value:      capabilities,
			Message:    "agent has capabilities from many different groups",
			Suggestion: fmt.Sprintf("consider focusing on specific capability groups: %s", strings.Join(groups, ", ")),
		})
	}

	// Suggest additional capabilities based on existing ones
	for group, count := range groupCounts {
		if count == 1 && len(v.capabilityGroups[group]) > 1 {
			errors = append(errors, core.ValidationError{
				Field:      "capabilities",
				Value:      capabilities,
				Message:    fmt.Sprintf("agent has only one capability from %s group", group),
				Suggestion: fmt.Sprintf("consider adding more %s capabilities: %s", group, strings.Join(v.capabilityGroups[group], ", ")),
			})
		}
	}

	return errors
}

// ValidateAgentNaming validates agent naming conventions
func (v *DefaultConfigValidator) ValidateAgentNaming(name string, config *core.AgentConfig) []core.ValidationError {
	var errors []core.ValidationError

	// Validate agent name format
	namePattern := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	if !namePattern.MatchString(name) {
		errors = append(errors, core.ValidationError{
			Field:      "name",
			Value:      name,
			Message:    "agent name must be lowercase with underscores only",
			Suggestion: "use format like 'research_agent' or 'data_processor'",
		})
	}

	// Check if role matches name convention
	if config.Role != "" && !strings.Contains(config.Role, name) && !strings.Contains(name, strings.Replace(config.Role, "_agent", "", 1)) {
		errors = append(errors, core.ValidationError{
			Field:      "role",
			Value:      config.Role,
			Message:    "role doesn't match agent name convention",
			Suggestion: fmt.Sprintf("consider using role '%s_agent' to match name '%s'", name, name),
		})
	}

	return errors
}

// validateCrossReferences validates cross-references between different parts of the configuration
func (v *DefaultConfigValidator) validateCrossReferences(config *core.Config) []core.ValidationError {
	var errors []core.ValidationError

	// Check for agents with different LLM providers than global
	globalProvider := config.LLM.Provider
	for name, agent := range config.Agents {
		if agent.LLM != nil && agent.LLM.Provider != "" && agent.LLM.Provider != globalProvider {
			errors = append(errors, core.ValidationError{
				Field:      fmt.Sprintf("agents.%s.llm.provider", name),
				Value:      agent.LLM.Provider,
				Message:    "agent uses different LLM provider than global configuration",
				Suggestion: "ensure this is intentional for agent-specific optimization",
			})
		}
	}

	// Validate orchestration makes sense with available agents
	totalAgents := len(config.Agents)
	enabledAgents := 0
	for _, agent := range config.Agents {
		if agent.Enabled {
			enabledAgents++
		}
	}

	if enabledAgents == 0 && totalAgents > 0 {
		errors = append(errors, core.ValidationError{
			Field:      "agents",
			Value:      nil,
			Message:    "all agents are disabled",
			Suggestion: "enable at least one agent for the system to function",
		})
	}

	// Check if orchestration mode matches agent configuration
	sequentialCount := len(config.Orchestration.SequentialAgents)
	collaborativeCount := len(config.Orchestration.CollaborativeAgents)
	hasLoopAgent := config.Orchestration.LoopAgent != ""

	if sequentialCount > 0 && collaborativeCount > 0 && hasLoopAgent {
		errors = append(errors, core.ValidationError{
			Field:      "orchestration",
			Value:      nil,
			Message:    "multiple orchestration modes configured",
			Suggestion: "choose one primary orchestration mode for clarity",
		})
	}

	if sequentialCount == 1 {
		errors = append(errors, core.ValidationError{
			Field:      "orchestration.sequential_agents",
			Value:      config.Orchestration.SequentialAgents,
			Message:    "only one agent in sequential mode",
			Suggestion: "add more agents for sequential processing or use single agent mode",
		})
	}

	if collaborativeCount == 1 {
		errors = append(errors, core.ValidationError{
			Field:      "orchestration.collaborative_agents",
			Value:      config.Orchestration.CollaborativeAgents,
			Message:    "only one agent in collaborative mode",
			Suggestion: "add more agents for collaboration or use single agent mode",
		})
	}

	return errors
}
