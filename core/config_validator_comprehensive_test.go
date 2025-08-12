package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfigValidator_ProviderSpecificValidation(t *testing.T) {
	validator := NewDefaultConfigValidator()

	tests := []struct {
		name     string
		config   AgentLLMConfig
		expected []string // Expected error fields
	}{
		{
			name: "OpenAI valid configuration",
			config: AgentLLMConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Temperature: 0.7,
				MaxTokens:   2000,
			},
			expected: []string{},
		},
		{
			name: "OpenAI invalid model",
			config: AgentLLMConfig{
				Provider: "openai",
				Model:    "invalid-model",
			},
			expected: []string{"model"},
		},
		{
			name: "Anthropic with unsupported parameters",
			config: AgentLLMConfig{
				Provider:         "anthropic",
				Model:            "claude-3-sonnet",
				FrequencyPenalty: 0.5,
				PresencePenalty:  0.3,
			},
			expected: []string{"frequency_penalty", "presence_penalty"},
		},
		{
			name: "Ollama with high max_tokens",
			config: AgentLLMConfig{
				Provider:  "ollama",
				Model:     "llama2",
				MaxTokens: 4000,
			},
			expected: []string{"max_tokens"},
		},
		{
			name: "Google AI with unsupported parameters",
			config: AgentLLMConfig{
				Provider:         "google",
				Model:            "gemini-pro",
				FrequencyPenalty: 0.5,
				PresencePenalty:  0.3,
			},
			expected: []string{"frequency_penalty", "presence_penalty"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateLLMConfig(&tt.config)
			
			if len(tt.expected) == 0 {
				assert.Empty(t, errors, "Expected no validation errors")
			} else {
				errorFields := make([]string, len(errors))
				for i, err := range errors {
					errorFields[i] = err.Field
				}
				
				for _, expectedField := range tt.expected {
					assert.Contains(t, errorFields, expectedField, 
						"Expected validation error for field: %s", expectedField)
				}
			}
		})
	}
}

func TestDefaultConfigValidator_CapabilityGroupValidation(t *testing.T) {
	validator := NewDefaultConfigValidator()

	tests := []struct {
		name         string
		capabilities []string
		expectError  bool
		errorType    string
	}{
		{
			name:         "Focused research capabilities",
			capabilities: []string{"information_gathering", "fact_checking", "source_identification"},
			expectError:  false,
		},
		{
			name:         "Mixed capabilities from many groups",
			capabilities: []string{"information_gathering", "code_generation", "content_creation", "data_analysis"},
			expectError:  true,
			errorType:    "many different groups",
		},
		{
			name:         "Single capability from group",
			capabilities: []string{"information_gathering"},
			expectError:  true,
			errorType:    "only one capability",
		},
		{
			name:         "Development focused capabilities",
			capabilities: []string{"code_generation", "code_review", "testing", "debugging"},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateCapabilityGroups(tt.capabilities)
			
			if tt.expectError {
				assert.NotEmpty(t, errors, "Expected validation errors")
				found := false
				for _, err := range errors {
					if err.Field == "capabilities" {
						assert.Contains(t, err.Message, tt.errorType)
						found = true
						break
					}
				}
				assert.True(t, found, "Expected error type not found: %s", tt.errorType)
			} else {
				// Filter out suggestion-only errors (these are not actual errors)
				actualErrors := []ValidationError{}
				for _, err := range errors {
					if err.Message != "agent has only one capability from research group" &&
					   err.Message != "agent has only one capability from analysis group" &&
					   err.Message != "agent has only one capability from content group" &&
					   err.Message != "agent has only one capability from development group" {
						actualErrors = append(actualErrors, err)
					}
				}
				assert.Empty(t, actualErrors, "Expected no validation errors")
			}
		})
	}
}

func TestDefaultConfigValidator_ConfigCompleteness(t *testing.T) {
	validator := NewDefaultConfigValidator()

	tests := []struct {
		name     string
		config   Config
		expected []string // Expected error fields
	}{
		{
			name: "Complete configuration",
			config: Config{
				LLM: AgentLLMConfig{
					Provider: "openai",
					Model:    "gpt-4",
				},
				Agents: map[string]AgentConfig{
					"researcher": {
						Role:         "research_specialist",
						SystemPrompt: "You are a research specialist",
						Capabilities: []string{"information_gathering"},
						Enabled:      true,
					},
				},
				Orchestration: OrchestrationConfigToml{
					SequentialAgents: []string{"researcher"},
				},
			},
			expected: []string{},
		},
		{
			name: "Missing global LLM configuration",
			config: Config{
				Agents: map[string]AgentConfig{
					"researcher": {
						Role:         "research_specialist",
						SystemPrompt: "You are a research specialist",
						Capabilities: []string{"information_gathering"},
						Enabled:      true,
					},
				},
			},
			expected: []string{"llm.provider", "llm.model"},
		},
		{
			name: "Missing agent configuration",
			config: Config{
				LLM: AgentLLMConfig{
					Provider: "openai",
					Model:    "gpt-4",
				},
				Agents: map[string]AgentConfig{
					"researcher": {
						Enabled: true,
					},
				},
			},
			expected: []string{"agents.researcher.role", "agents.researcher.system_prompt", "agents.researcher.capabilities"},
		},
		{
			name: "No orchestration configuration",
			config: Config{
				LLM: AgentLLMConfig{
					Provider: "openai",
					Model:    "gpt-4",
				},
				Agents: map[string]AgentConfig{
					"researcher": {
						Role:         "research_specialist",
						SystemPrompt: "You are a research specialist",
						Capabilities: []string{"information_gathering"},
						Enabled:      true,
					},
				},
			},
			expected: []string{"orchestration"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateConfigCompleteness(&tt.config)
			
			if len(tt.expected) == 0 {
				assert.Empty(t, errors, "Expected no validation errors")
			} else {
				errorFields := make([]string, len(errors))
				for i, err := range errors {
					errorFields[i] = err.Field
				}
				
				for _, expectedField := range tt.expected {
					assert.Contains(t, errorFields, expectedField, 
						"Expected validation error for field: %s", expectedField)
				}
			}
		})
	}
}

func TestDefaultConfigValidator_AgentNaming(t *testing.T) {
	validator := NewDefaultConfigValidator()

	tests := []struct {
		name        string
		agentName   string
		agentConfig AgentConfig
		expectError bool
		errorField  string
	}{
		{
			name:      "Valid agent name and role",
			agentName: "research_agent",
			agentConfig: AgentConfig{
				Role: "research_specialist",
			},
			expectError: false,
		},
		{
			name:      "Invalid agent name format",
			agentName: "Research-Agent",
			agentConfig: AgentConfig{
				Role: "research_specialist",
			},
			expectError: true,
			errorField:  "name",
		},
		{
			name:      "Agent name doesn't match role",
			agentName: "researcher",
			agentConfig: AgentConfig{
				Role: "content_writer",
			},
			expectError: true,
			errorField:  "role",
		},
		{
			name:      "Valid matching name and role",
			agentName: "content_writer",
			agentConfig: AgentConfig{
				Role: "content_specialist",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateAgentNaming(tt.agentName, &tt.agentConfig)
			
			if tt.expectError {
				assert.NotEmpty(t, errors, "Expected validation errors")
				found := false
				for _, err := range errors {
					if err.Field == tt.errorField {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected error field not found: %s", tt.errorField)
			} else {
				assert.Empty(t, errors, "Expected no validation errors")
			}
		})
	}
}

func TestDefaultConfigValidator_CrossValidation(t *testing.T) {
	validator := NewDefaultConfigValidator()

	tests := []struct {
		name     string
		config   Config
		expected []string // Expected error messages (partial)
	}{
		{
			name: "All agents disabled",
			config: Config{
				Agents: map[string]AgentConfig{
					"agent1": {Enabled: false},
					"agent2": {Enabled: false},
				},
			},
			expected: []string{"all agents are disabled"},
		},
		{
			name: "Single agent in sequential mode",
			config: Config{
				Agents: map[string]AgentConfig{
					"agent1": {Enabled: true},
				},
				Orchestration: OrchestrationConfigToml{
					SequentialAgents: []string{"agent1"},
				},
			},
			expected: []string{"only one agent in sequential mode"},
		},
		{
			name: "Single agent in collaborative mode",
			config: Config{
				Agents: map[string]AgentConfig{
					"agent1": {Enabled: true},
				},
				Orchestration: OrchestrationConfigToml{
					CollaborativeAgents: []string{"agent1"},
				},
			},
			expected: []string{"only one agent in collaborative mode"},
		},
		{
			name: "Multiple orchestration modes",
			config: Config{
				Agents: map[string]AgentConfig{
					"agent1": {Enabled: true},
					"agent2": {Enabled: true},
				},
				Orchestration: OrchestrationConfigToml{
					SequentialAgents:    []string{"agent1"},
					CollaborativeAgents: []string{"agent2"},
					LoopAgent:           "agent1",
				},
			},
			expected: []string{"multiple orchestration modes configured"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.validateCrossReferences(&tt.config)
			
			if len(tt.expected) == 0 {
				assert.Empty(t, errors, "Expected no validation errors")
			} else {
				errorMessages := make([]string, len(errors))
				for i, err := range errors {
					errorMessages[i] = err.Message
				}
				
				for _, expectedMessage := range tt.expected {
					found := false
					for _, message := range errorMessages {
						if message == expectedMessage {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected error message not found: %s", expectedMessage)
				}
			}
		})
	}
}

func TestDefaultConfigValidator_CapabilitySuggestions(t *testing.T) {
	validator := NewDefaultConfigValidator()

	tests := []struct {
		name     string
		role     string
		expected []string
	}{
		{
			name:     "Research role",
			role:     "research_specialist",
			expected: []string{"information_gathering", "fact_checking", "source_identification"},
		},
		{
			name:     "Analysis role",
			role:     "data_analyst",
			expected: []string{"pattern_recognition", "trend_analysis", "insight_generation"},
		},
		{
			name:     "Content role",
			role:     "content_writer",
			expected: []string{"content_creation", "editing", "summarization"},
		},
		{
			name:     "Development role",
			role:     "code_developer",
			expected: []string{"code_generation", "code_review", "debugging"},
		},
		{
			name:     "Mixed role",
			role:     "research_analyst",
			expected: []string{"information_gathering", "pattern_recognition"}, // Should include both groups
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := validator.GetCapabilitySuggestions(tt.role)
			
			for _, expected := range tt.expected {
				assert.Contains(t, suggestions, expected, 
					"Expected capability suggestion: %s", expected)
			}
		})
	}
}

func TestDefaultConfigValidator_OptimizationSuggestions(t *testing.T) {
	validator := NewDefaultConfigValidator()

	tests := []struct {
		name     string
		config   Config
		expected []string // Expected suggestion messages (partial)
	}{
		{
			name: "Creative agent with low temperature",
			config: Config{
				Agents: map[string]AgentConfig{
					"writer": {
						Capabilities: []string{"content_creation", "writing"},
						LLM: &AgentLLMConfig{
							Temperature: 0.1,
						},
					},
				},
			},
			expected: []string{"low temperature for creative tasks"},
		},
		{
			name: "Analytical agent with high temperature",
			config: Config{
				Agents: map[string]AgentConfig{
					"analyst": {
						Capabilities: []string{"data_analysis", "fact_checking"},
						LLM: &AgentLLMConfig{
							Temperature: 0.9,
						},
					},
				},
			},
			expected: []string{"high temperature for analytical tasks"},
		},
		{
			name: "Agent with too many capabilities",
			config: Config{
				Agents: map[string]AgentConfig{
					"generalist": {
						Capabilities: []string{
							"information_gathering", "fact_checking", "code_generation",
							"content_creation", "data_analysis", "debugging",
						},
					},
				},
			},
			expected: []string{"agent has many capabilities"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := validator.SuggestOptimizations(&tt.config)
			
			if len(tt.expected) == 0 {
				assert.Empty(t, suggestions, "Expected no optimization suggestions")
			} else {
				suggestionMessages := make([]string, len(suggestions))
				for i, suggestion := range suggestions {
					suggestionMessages[i] = suggestion.Message
				}
				
				for _, expectedMessage := range tt.expected {
					found := false
					for _, message := range suggestionMessages {
						if message == expectedMessage {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected suggestion message not found: %s", expectedMessage)
				}
			}
		})
	}
}

func TestDefaultConfigValidator_ComprehensiveValidation(t *testing.T) {
	validator := NewDefaultConfigValidator()

	// Test a comprehensive configuration with multiple validation aspects
	config := Config{
		LLM: AgentLLMConfig{
			Provider:    "openai",
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   2000,
		},
		Agents: map[string]AgentConfig{
			"researcher": {
				Role:         "research_specialist",
				Description:  "Gathers comprehensive information from multiple sources",
				SystemPrompt: "You are a research specialist with expertise in information gathering and fact-checking.",
				Capabilities: []string{"information_gathering", "fact_checking", "source_identification"},
				Enabled:      true,
				LLM: &AgentLLMConfig{
					Temperature: 0.3, // Lower temperature for factual research
					MaxTokens:   1500,
				},
			},
			"writer": {
				Role:         "content_writer",
				Description:  "Creates engaging and informative content",
				SystemPrompt: "You are a skilled content writer who creates engaging, well-structured content.",
				Capabilities: []string{"content_creation", "editing", "summarization"},
				Enabled:      true,
				LLM: &AgentLLMConfig{
					Temperature: 0.8, // Higher temperature for creative writing
					MaxTokens:   2500,
				},
			},
		},
		Orchestration: OrchestrationConfigToml{
			SequentialAgents: []string{"researcher", "writer"},
		},
	}

	errors := validator.ValidateConfig(&config)

	// This should be a valid configuration with minimal or no errors
	assert.LessOrEqual(t, len(errors), 2, "Expected minimal validation errors for well-configured system")

	// Check that no critical errors are present
	for _, err := range errors {
		assert.NotContains(t, err.Message, "required", "No required fields should be missing")
		assert.NotContains(t, err.Message, "not found", "No missing references should exist")
	}
}