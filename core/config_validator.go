package core

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field      string      `json:"field"`
	Value      interface{} `json:"value"`
	Message    string      `json:"message"`
	Suggestion string      `json:"suggestion"`
}

func (e ValidationError) Error() string {
	if e.Suggestion != "" {
		return fmt.Sprintf("%s: %s. Suggestion: %s", e.Field, e.Message, e.Suggestion)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ConfigValidator interface for agent configuration validation
type ConfigValidator interface {
	ValidateAgentConfig(name string, config *AgentConfig) []ValidationError
	ValidateLLMConfig(config *AgentLLMConfig) []ValidationError
	ValidateOrchestrationAgents(orchestration *OrchestrationConfigToml, agents map[string]AgentConfig) []ValidationError
	ValidateCapabilities(capabilities []string) []ValidationError
	ValidateConfig(config *Config) []ValidationError
}

// DefaultConfigValidator implements ConfigValidator
type DefaultConfigValidator struct {
	knownCapabilities map[string]bool
	validProviders    map[string]bool
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
		},
		validProviders: map[string]bool{
			"openai":  true,
			"azure":   true,
			"ollama":  true,
			"anthropic": true,
		},
	}
}

// ValidateConfig validates the entire configuration
func (v *DefaultConfigValidator) ValidateConfig(config *Config) []ValidationError {
	var errors []ValidationError

	// Validate global LLM configuration
	llmErrors := v.ValidateLLMConfig(&config.LLM)
	for _, err := range llmErrors {
		err.Field = "llm." + err.Field
		errors = append(errors, err)
	}

	// Validate each agent configuration
	for name, agent := range config.Agents {
		agentErrors := v.ValidateAgentConfig(name, &agent)
		for _, err := range agentErrors {
			err.Field = fmt.Sprintf("agents.%s.%s", name, err.Field)
			errors = append(errors, err)
		}
	}

	// Validate orchestration configuration against agents
	orchErrors := v.ValidateOrchestrationAgents(&config.Orchestration, config.Agents)
	for _, err := range orchErrors {
		err.Field = "orchestration." + err.Field
		errors = append(errors, err)
	}

	return errors
}

// ValidateAgentConfig validates agent-specific configuration
func (v *DefaultConfigValidator) ValidateAgentConfig(name string, config *AgentConfig) []ValidationError {
	var errors []ValidationError

	// Validate required fields
	if config.Role == "" {
		errors = append(errors, ValidationError{
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
			errors = append(errors, ValidationError{
				Field:      "role",
				Value:      config.Role,
				Message:    "role must be lowercase with underscores only",
				Suggestion: "use format like 'research_agent' or 'data_processor'",
			})
		}
	}

	// Validate description
	if config.Description == "" {
		errors = append(errors, ValidationError{
			Field:      "description",
			Value:      config.Description,
			Message:    "description is recommended for documentation",
			Suggestion: "provide a brief description of the agent's purpose",
		})
	}

	// Validate system prompt
	if config.SystemPrompt == "" {
		errors = append(errors, ValidationError{
			Field:      "system_prompt",
			Value:      config.SystemPrompt,
			Message:    "system_prompt is required for agent behavior",
			Suggestion: "provide a clear system prompt defining the agent's role and behavior",
		})
	} else if len(config.SystemPrompt) < 10 {
		errors = append(errors, ValidationError{
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
		errors = append(errors, ValidationError{
			Field:      "timeout_seconds",
			Value:      config.Timeout,
			Message:    "timeout cannot be negative",
			Suggestion: "set timeout to a positive value (e.g., 30 seconds)",
		})
	} else if config.Timeout > 300 {
		errors = append(errors, ValidationError{
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
func (v *DefaultConfigValidator) ValidateLLMConfig(config *AgentLLMConfig) []ValidationError {
	var errors []ValidationError

	// Validate provider
	if config.Provider != "" && !v.validProviders[config.Provider] {
		validProviders := make([]string, 0, len(v.validProviders))
		for provider := range v.validProviders {
			validProviders = append(validProviders, provider)
		}
		errors = append(errors, ValidationError{
			Field:      "provider",
			Value:      config.Provider,
			Message:    "unsupported LLM provider",
			Suggestion: fmt.Sprintf("use one of: %s", strings.Join(validProviders, ", ")),
		})
	}

	// Validate temperature
	if config.Temperature < 0 || config.Temperature > 2 {
		errors = append(errors, ValidationError{
			Field:      "temperature",
			Value:      config.Temperature,
			Message:    "temperature must be between 0 and 2",
			Suggestion: "use 0.1-0.3 for factual tasks, 0.7-1.0 for creative tasks",
		})
	}

	// Validate max tokens
	if config.MaxTokens < 1 {
		errors = append(errors, ValidationError{
			Field:      "max_tokens",
			Value:      config.MaxTokens,
			Message:    "max_tokens must be positive",
			Suggestion: "set max_tokens to a reasonable value (e.g., 800-2000)",
		})
	} else if config.MaxTokens > 32000 {
		errors = append(errors, ValidationError{
			Field:      "max_tokens",
			Value:      config.MaxTokens,
			Message:    "max_tokens is very high",
			Suggestion: "consider reducing max_tokens to avoid high costs",
		})
	}

	// Validate timeout
	if config.TimeoutSeconds < 0 {
		errors = append(errors, ValidationError{
			Field:      "timeout_seconds",
			Value:      config.TimeoutSeconds,
			Message:    "timeout_seconds cannot be negative",
			Suggestion: "set timeout_seconds to a positive value (e.g., 30)",
		})
	} else if config.TimeoutSeconds > 300 {
		errors = append(errors, ValidationError{
			Field:      "timeout_seconds",
			Value:      config.TimeoutSeconds,
			Message:    "timeout_seconds is very high (>5 minutes)",
			Suggestion: "consider reducing timeout to avoid long waits",
		})
	}

	// Validate TopP
	if config.TopP != 0 && (config.TopP < 0 || config.TopP > 1) {
		errors = append(errors, ValidationError{
			Field:      "top_p",
			Value:      config.TopP,
			Message:    "top_p must be between 0 and 1",
			Suggestion: "use values like 0.9 for diverse output, 0.1 for focused output",
		})
	}

	// Validate frequency penalty
	if config.FrequencyPenalty < -2 || config.FrequencyPenalty > 2 {
		errors = append(errors, ValidationError{
			Field:      "frequency_penalty",
			Value:      config.FrequencyPenalty,
			Message:    "frequency_penalty must be between -2 and 2",
			Suggestion: "use positive values to reduce repetition, negative to encourage it",
		})
	}

	// Validate presence penalty
	if config.PresencePenalty < -2 || config.PresencePenalty > 2 {
		errors = append(errors, ValidationError{
			Field:      "presence_penalty",
			Value:      config.PresencePenalty,
			Message:    "presence_penalty must be between -2 and 2",
			Suggestion: "use positive values to encourage new topics, negative to stay on topic",
		})
	}

	return errors
}

// ValidateCapabilities validates agent capabilities
func (v *DefaultConfigValidator) ValidateCapabilities(capabilities []string) []ValidationError {
	var errors []ValidationError

	if len(capabilities) == 0 {
		errors = append(errors, ValidationError{
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
		errors = append(errors, ValidationError{
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
		errors = append(errors, ValidationError{
			Field:      "capabilities",
			Value:      duplicates,
			Message:    fmt.Sprintf("duplicate capabilities: %s", strings.Join(duplicates, ", ")),
			Suggestion: "remove duplicate capability entries",
		})
	}

	return errors
}

// ValidateOrchestrationAgents validates orchestration configuration against available agents
func (v *DefaultConfigValidator) ValidateOrchestrationAgents(orchestration *OrchestrationConfigToml, agents map[string]AgentConfig) []ValidationError {
	var errors []ValidationError

	// Validate sequential agents exist
	for _, agentName := range orchestration.SequentialAgents {
		if _, exists := agents[agentName]; !exists {
			errors = append(errors, ValidationError{
				Field:      "sequential_agents",
				Value:      agentName,
				Message:    fmt.Sprintf("agent '%s' not found in configuration", agentName),
				Suggestion: "ensure all referenced agents are defined in the agents section",
			})
		} else if !agents[agentName].Enabled {
			errors = append(errors, ValidationError{
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
			errors = append(errors, ValidationError{
				Field:      "collaborative_agents",
				Value:      agentName,
				Message:    fmt.Sprintf("agent '%s' not found in configuration", agentName),
				Suggestion: "ensure all referenced agents are defined in the agents section",
			})
		} else if !agents[agentName].Enabled {
			errors = append(errors, ValidationError{
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
			errors = append(errors, ValidationError{
				Field:      "loop_agent",
				Value:      orchestration.LoopAgent,
				Message:    fmt.Sprintf("agent '%s' not found in configuration", orchestration.LoopAgent),
				Suggestion: "ensure the loop agent is defined in the agents section",
			})
		} else if !agents[orchestration.LoopAgent].Enabled {
			errors = append(errors, ValidationError{
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
func (v *DefaultConfigValidator) validateRetryPolicy(policy *AgentRetryPolicyConfig) []ValidationError {
	var errors []ValidationError

	if policy.MaxRetries < 0 {
		errors = append(errors, ValidationError{
			Field:      "max_retries",
			Value:      policy.MaxRetries,
			Message:    "max_retries cannot be negative",
			Suggestion: "set max_retries to 0 to disable retries or a positive value",
		})
	} else if policy.MaxRetries > 10 {
		errors = append(errors, ValidationError{
			Field:      "max_retries",
			Value:      policy.MaxRetries,
			Message:    "max_retries is very high",
			Suggestion: "consider reducing max_retries to avoid excessive delays",
		})
	}

	if policy.BaseDelayMs < 0 {
		errors = append(errors, ValidationError{
			Field:      "base_delay_ms",
			Value:      policy.BaseDelayMs,
			Message:    "base_delay_ms cannot be negative",
			Suggestion: "set base_delay_ms to a positive value (e.g., 1000)",
		})
	}

	if policy.MaxDelayMs < policy.BaseDelayMs {
		errors = append(errors, ValidationError{
			Field:      "max_delay_ms",
			Value:      policy.MaxDelayMs,
			Message:    "max_delay_ms must be greater than or equal to base_delay_ms",
			Suggestion: "increase max_delay_ms or decrease base_delay_ms",
		})
	}

	if policy.BackoffFactor <= 0 {
		errors = append(errors, ValidationError{
			Field:      "backoff_factor",
			Value:      policy.BackoffFactor,
			Message:    "backoff_factor must be positive",
			Suggestion: "set backoff_factor to a value like 2.0 for exponential backoff",
		})
	}

	return errors
}

// validateRateLimit validates rate limit configuration
func (v *DefaultConfigValidator) validateRateLimit(rateLimit *RateLimitConfig) []ValidationError {
	var errors []ValidationError

	if rateLimit.RequestsPerSecond <= 0 {
		errors = append(errors, ValidationError{
			Field:      "requests_per_second",
			Value:      rateLimit.RequestsPerSecond,
			Message:    "requests_per_second must be positive",
			Suggestion: "set requests_per_second to a reasonable value (e.g., 10)",
		})
	}

	if rateLimit.BurstSize < 0 {
		errors = append(errors, ValidationError{
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