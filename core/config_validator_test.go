package core

import (
	"strings"
	"testing"
)

func TestValidateAgentConfig(t *testing.T) {
	validator := NewDefaultConfigValidator()

	tests := []struct {
		name          string
		agentName     string
		config        AgentConfig
		expectedErrors int
		checkError     func([]ValidationError) bool
	}{
		{
			name:      "valid_config",
			agentName: "test_agent",
			config: AgentConfig{
				Role:         "research_agent",
				Description:  "A research agent for testing",
				SystemPrompt: "You are a helpful research assistant",
				Capabilities: []string{"information_gathering", "fact_checking"},
				Enabled:      true,
				Timeout:      30,
			},
			expectedErrors: 0,
		},
		{
			name:      "missing_role",
			agentName: "test_agent",
			config: AgentConfig{
				Description:  "A test agent",
				SystemPrompt: "You are a helpful assistant",
				Capabilities: []string{"information_gathering"},
				Enabled:      true,
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "role"
			},
		},
		{
			name:      "invalid_role_format",
			agentName: "test_agent",
			config: AgentConfig{
				Role:         "Research-Agent",
				Description:  "A test agent",
				SystemPrompt: "You are a helpful assistant",
				Capabilities: []string{"information_gathering"},
				Enabled:      true,
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "role"
			},
		},
		{
			name:      "missing_system_prompt",
			agentName: "test_agent",
			config: AgentConfig{
				Role:         "test_agent",
				Description:  "A test agent",
				Capabilities: []string{"information_gathering"},
				Enabled:      true,
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "system_prompt"
			},
		},
		{
			name:      "short_system_prompt",
			agentName: "test_agent",
			config: AgentConfig{
				Role:         "test_agent",
				Description:  "A test agent",
				SystemPrompt: "Hi",
				Capabilities: []string{"information_gathering"},
				Enabled:      true,
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "system_prompt"
			},
		},
		{
			name:      "no_capabilities",
			agentName: "test_agent",
			config: AgentConfig{
				Role:         "test_agent",
				Description:  "A test agent",
				SystemPrompt: "You are a helpful assistant",
				Capabilities: []string{},
				Enabled:      true,
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "capabilities"
			},
		},
		{
			name:      "unknown_capabilities",
			agentName: "test_agent",
			config: AgentConfig{
				Role:         "test_agent",
				Description:  "A test agent",
				SystemPrompt: "You are a helpful assistant",
				Capabilities: []string{"unknown_capability", "another_unknown"},
				Enabled:      true,
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "capabilities"
			},
		},
		{
			name:      "duplicate_capabilities",
			agentName: "test_agent",
			config: AgentConfig{
				Role:         "test_agent",
				Description:  "A test agent",
				SystemPrompt: "You are a helpful assistant",
				Capabilities: []string{"information_gathering", "information_gathering"},
				Enabled:      true,
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "capabilities"
			},
		},
		{
			name:      "negative_timeout",
			agentName: "test_agent",
			config: AgentConfig{
				Role:         "test_agent",
				Description:  "A test agent",
				SystemPrompt: "You are a helpful assistant",
				Capabilities: []string{"information_gathering"},
				Enabled:      true,
				Timeout:      -10,
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "timeout_seconds"
			},
		},
		{
			name:      "very_high_timeout",
			agentName: "test_agent",
			config: AgentConfig{
				Role:         "test_agent",
				Description:  "A test agent",
				SystemPrompt: "You are a helpful assistant",
				Capabilities: []string{"information_gathering"},
				Enabled:      true,
				Timeout:      400,
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "timeout_seconds"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateAgentConfig(tt.agentName, &tt.config)
			
			if len(errors) != tt.expectedErrors {
				t.Errorf("Expected %d errors, got %d: %v", tt.expectedErrors, len(errors), errors)
			}
			
			if tt.checkError != nil && !tt.checkError(errors) {
				t.Errorf("Error check failed for test %s: %v", tt.name, errors)
			}
		})
	}
}

func TestValidateLLMConfig(t *testing.T) {
	validator := NewDefaultConfigValidator()

	tests := []struct {
		name          string
		config        AgentLLMConfig
		expectedErrors int
		checkError     func([]ValidationError) bool
	}{
		{
			name: "valid_config",
			config: AgentLLMConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Temperature: 0.7,
				MaxTokens:   800,
				TimeoutSeconds: 30,
			},
			expectedErrors: 0,
		},
		{
			name: "invalid_provider",
			config: AgentLLMConfig{
				Provider:    "unknown_provider",
				Temperature: 0.7,
				MaxTokens:   800,
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "provider"
			},
		},
		{
			name: "invalid_temperature_low",
			config: AgentLLMConfig{
				Provider:    "openai",
				Temperature: -0.5,
				MaxTokens:   800,
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "temperature"
			},
		},
		{
			name: "invalid_temperature_high",
			config: AgentLLMConfig{
				Provider:    "openai",
				Temperature: 2.5,
				MaxTokens:   800,
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "temperature"
			},
		},
		{
			name: "invalid_max_tokens_zero",
			config: AgentLLMConfig{
				Provider:    "openai",
				Temperature: 0.7,
				MaxTokens:   0,
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "max_tokens"
			},
		},
		{
			name: "invalid_max_tokens_high",
			config: AgentLLMConfig{
				Provider:    "openai",
				Temperature: 0.7,
				MaxTokens:   50000,
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "max_tokens"
			},
		},
		{
			name: "invalid_top_p",
			config: AgentLLMConfig{
				Provider:    "openai",
				Temperature: 0.7,
				MaxTokens:   800,
				TopP:        1.5,
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "top_p"
			},
		},
		{
			name: "invalid_frequency_penalty",
			config: AgentLLMConfig{
				Provider:         "openai",
				Temperature:     0.7,
				MaxTokens:       800,
				FrequencyPenalty: 3.0,
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "frequency_penalty"
			},
		},
		{
			name: "invalid_presence_penalty",
			config: AgentLLMConfig{
				Provider:        "openai",
				Temperature:    0.7,
				MaxTokens:      800,
				PresencePenalty: -3.0,
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "presence_penalty"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateLLMConfig(&tt.config)
			
			if len(errors) != tt.expectedErrors {
				t.Errorf("Expected %d errors, got %d: %v", tt.expectedErrors, len(errors), errors)
			}
			
			if tt.checkError != nil && !tt.checkError(errors) {
				t.Errorf("Error check failed for test %s: %v", tt.name, errors)
			}
		})
	}
}

func TestValidateOrchestrationAgents(t *testing.T) {
	validator := NewDefaultConfigValidator()

	agents := map[string]AgentConfig{
		"agent1": {Enabled: true},
		"agent2": {Enabled: true},
		"disabled_agent": {Enabled: false},
	}

	tests := []struct {
		name          string
		orchestration OrchestrationConfigToml
		expectedErrors int
		checkError     func([]ValidationError) bool
	}{
		{
			name: "valid_sequential_agents",
			orchestration: OrchestrationConfigToml{
				SequentialAgents: []string{"agent1", "agent2"},
			},
			expectedErrors: 0,
		},
		{
			name: "nonexistent_sequential_agent",
			orchestration: OrchestrationConfigToml{
				SequentialAgents: []string{"agent1", "nonexistent"},
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "sequential_agents"
			},
		},
		{
			name: "disabled_sequential_agent",
			orchestration: OrchestrationConfigToml{
				SequentialAgents: []string{"agent1", "disabled_agent"},
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "sequential_agents"
			},
		},
		{
			name: "valid_collaborative_agents",
			orchestration: OrchestrationConfigToml{
				CollaborativeAgents: []string{"agent1", "agent2"},
			},
			expectedErrors: 0,
		},
		{
			name: "nonexistent_collaborative_agent",
			orchestration: OrchestrationConfigToml{
				CollaborativeAgents: []string{"agent1", "nonexistent"},
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "collaborative_agents"
			},
		},
		{
			name: "valid_loop_agent",
			orchestration: OrchestrationConfigToml{
				LoopAgent: "agent1",
			},
			expectedErrors: 0,
		},
		{
			name: "nonexistent_loop_agent",
			orchestration: OrchestrationConfigToml{
				LoopAgent: "nonexistent",
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "loop_agent"
			},
		},
		{
			name: "disabled_loop_agent",
			orchestration: OrchestrationConfigToml{
				LoopAgent: "disabled_agent",
			},
			expectedErrors: 1,
			checkError: func(errors []ValidationError) bool {
				return len(errors) > 0 && errors[0].Field == "loop_agent"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateOrchestrationAgents(&tt.orchestration, agents)
			
			if len(errors) != tt.expectedErrors {
				t.Errorf("Expected %d errors, got %d: %v", tt.expectedErrors, len(errors), errors)
			}
			
			if tt.checkError != nil && !tt.checkError(errors) {
				t.Errorf("Error check failed for test %s: %v", tt.name, errors)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	validator := NewDefaultConfigValidator()

	config := &Config{
		LLM: AgentLLMConfig{
			Provider:    "openai",
			Temperature: 0.7,
			MaxTokens:   800,
		},
		Agents: map[string]AgentConfig{
			"valid_agent": {
				Role:         "test_agent",
				Description:  "A test agent",
				SystemPrompt: "You are a helpful assistant",
				Capabilities: []string{"information_gathering"},
				Enabled:      true,
			},
			"invalid_agent": {
				Role:         "", // Missing role
				SystemPrompt: "You are a helpful assistant",
				Capabilities: []string{"unknown_capability"}, // Unknown capability
				Enabled:      true,
			},
		},
		Orchestration: OrchestrationConfigToml{
			SequentialAgents: []string{"valid_agent", "nonexistent_agent"},
		},
	}

	errors := validator.ValidateConfig(config)

	// Should have errors for:
	// 1. invalid_agent missing role
	// 2. invalid_agent unknown capability
	// 3. orchestration referencing nonexistent_agent
	if len(errors) < 3 {
		t.Errorf("Expected at least 3 errors, got %d: %v", len(errors), errors)
	}

	// Check that error fields are properly prefixed
	foundAgentError := false
	foundOrchestrationError := false
	for _, err := range errors {
		if strings.HasPrefix(err.Field, "agents.") {
			foundAgentError = true
		}
		if strings.HasPrefix(err.Field, "orchestration.") {
			foundOrchestrationError = true
		}
	}

	if !foundAgentError {
		t.Error("Expected to find agent validation error with proper field prefix")
	}
	if !foundOrchestrationError {
		t.Error("Expected to find orchestration validation error with proper field prefix")
	}
}

func TestValidationErrorString(t *testing.T) {
	// Test error without suggestion
	err1 := ValidationError{
		Field:   "test_field",
		Value:   "test_value",
		Message: "test message",
	}
	expected1 := "test_field: test message"
	if err1.Error() != expected1 {
		t.Errorf("Expected '%s', got '%s'", expected1, err1.Error())
	}

	// Test error with suggestion
	err2 := ValidationError{
		Field:      "test_field",
		Value:      "test_value",
		Message:    "test message",
		Suggestion: "test suggestion",
	}
	expected2 := "test_field: test message. Suggestion: test suggestion"
	if err2.Error() != expected2 {
		t.Errorf("Expected '%s', got '%s'", expected2, err2.Error())
	}
}

func TestAddKnownCapabilityAndProvider(t *testing.T) {
	validator := NewDefaultConfigValidator()

	// Test adding custom capability
	validator.AddKnownCapability("custom_capability")
	
	config := AgentConfig{
		Role:         "test_agent",
		Description:  "A test agent",
		SystemPrompt: "You are a helpful assistant",
		Capabilities: []string{"custom_capability"},
		Enabled:      true,
	}

	errors := validator.ValidateAgentConfig("test", &config)
	// Should not have capability errors now
	for _, err := range errors {
		if err.Field == "capabilities" {
			t.Errorf("Should not have capability error after adding custom capability: %v", err)
		}
	}

	// Test adding custom provider
	validator.AddValidProvider("custom_provider")
	
	llmConfig := AgentLLMConfig{
		Provider:    "custom_provider",
		Temperature: 0.7,
		MaxTokens:   800,
	}

	llmErrors := validator.ValidateLLMConfig(&llmConfig)
	// Should not have provider errors now
	for _, err := range llmErrors {
		if err.Field == "provider" {
			t.Errorf("Should not have provider error after adding custom provider: %v", err)
		}
	}
}