package main

import (
	"fmt"

	"github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
	// Example of comprehensive configuration validation
	fmt.Println("=== AgenticGoKit Configuration Validation Example ===\n")

	// Create a configuration with various validation issues
	config := &core.Config{}

	// Set agent flow configuration
	config.AgentFlow.Name = "validation-demo"

	// Set global LLM configuration
	config.LLM = core.AgentLLMConfig{
		Provider:    "openai",
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   2000,
	}

	// Set agents configuration
	config.Agents = map[string]core.AgentConfig{
		"researcher": {
			Role:        "research_specialist",
			Description: "Conducts comprehensive research on various topics",
			Capabilities: []string{
				"information_gathering",
				"fact_checking",
				"source_identification",
			},
			SystemPrompt: "You are a research specialist focused on gathering accurate information.",
			LLM: &core.AgentLLMConfig{
				Temperature: 0.3,
				MaxTokens:   1500,
			},
		},
		"writer": {
			Role:        "content_writer",
			Description: "Creates engaging and informative content",
			Capabilities: []string{
				"text_analysis",
				"summarization",
				"documentation",
			},
			SystemPrompt: "You are a skilled content writer who creates clear, engaging content.",
		},
		"invalid_agent": {
			Role:        "Invalid-Role", // Invalid naming convention
			Description: "Hi",           // Too short description
			Capabilities: []string{
				"unknown_capability", // Unknown capability
				"unknown_capability", // Duplicate capability
			},
			SystemPrompt: "Hi", // Too short system prompt
			Timeout:      -10,  // Invalid timeout
			LLM: &core.AgentLLMConfig{
				Temperature: 3.0, // Invalid temperature
				MaxTokens:   0,   // Invalid max tokens
			},
		},
	}

	// Set orchestration configuration
	config.Orchestration = core.OrchestrationConfigToml{
		SequentialAgents: []string{"researcher", "writer"},
		CollaborativeAgents: []string{
			"researcher",
			"writer",
			"nonexistent_agent", // Reference to non-existent agent
		},
	}

	// Create validator
	validator := core.NewDefaultConfigValidator()

	fmt.Println("1. Basic Configuration Validation:")
	fmt.Println("==================================")
	errors := validator.ValidateConfig(config)

	if len(errors) == 0 {
		fmt.Println("No validation errors found.")
	} else {
		fmt.Printf("Found %d validation issues:\n\n", len(errors))
		for i, err := range errors {
			fmt.Printf("%d. Field: %s\n", i+1, err.Field)
			fmt.Printf("   Issue: %s\n", err.Message)
			if err.Suggestion != "" {
				fmt.Printf("   Suggestion: %s\n", err.Suggestion)
			}
			fmt.Printf("   Current Value: %v\n\n", err.Value)
		}
	}

	fmt.Println("2. Agent-Specific Validation:")
	fmt.Println("=============================")
	for agentName, agentConfig := range config.Agents {
		fmt.Printf("Validating agent: %s\n", agentName)
		agentErrors := validator.ValidateAgentConfig(agentName, &agentConfig)
		if len(agentErrors) == 0 {
			fmt.Printf("Agent '%s' is valid\n\n", agentName)
		} else {
			fmt.Printf("Agent '%s' has %d issues:\n", agentName, len(agentErrors))
			for _, err := range agentErrors {
				fmt.Printf("   - %s: %s\n", err.Field, err.Message)
				if err.Suggestion != "" {
					fmt.Printf("     Suggestion: %s\n", err.Suggestion)
				}
			}
			fmt.Println()
		}
	}

	fmt.Println("3. LLM Configuration Validation:")
	fmt.Println("================================")
	// Validate the global LLM configuration
	llmErrors := validator.ValidateLLMConfig(&config.LLM)
	if len(llmErrors) == 0 {
		fmt.Println("LLM configuration is valid")
	} else {
		fmt.Printf("LLM configuration has %d issues:\n", len(llmErrors))
		for _, err := range llmErrors {
			fmt.Printf("   - %s: %s\n", err.Field, err.Message)
			if err.Suggestion != "" {
				fmt.Printf("     Suggestion: %s\n", err.Suggestion)
			}
		}
	}

	fmt.Println("\n4. Orchestration Validation:")
	fmt.Println("============================")
	orchErrors := validator.ValidateOrchestrationAgents(&config.Orchestration, config.Agents)
	if len(orchErrors) == 0 {
		fmt.Println("Orchestration configuration is valid")
	} else {
		fmt.Printf("Orchestration has %d issues:\n", len(orchErrors))
		for _, err := range orchErrors {
			fmt.Printf("   - %s: %s\n", err.Field, err.Message)
			if err.Suggestion != "" {
				fmt.Printf("     Suggestion: %s\n", err.Suggestion)
			}
		}
	}

	fmt.Println("\n5. Capability Validation:")
	fmt.Println("=========================")
	for agentName, agentConfig := range config.Agents {
		fmt.Printf("Validating capabilities for agent: %s\n", agentName)
		capErrors := validator.ValidateCapabilities(agentConfig.Capabilities)
		if len(capErrors) == 0 {
			fmt.Printf("Capabilities for '%s' are valid\n", agentName)
		} else {
			fmt.Printf("Capabilities for '%s' have issues:\n", agentName)
			for _, err := range capErrors {
				fmt.Printf("   - %s\n", err.Message)
				if err.Suggestion != "" {
					fmt.Printf("     Suggestion: %s\n", err.Suggestion)
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("=== Validation Complete ===")
	fmt.Printf("Total validation issues found: %d\n", len(errors))

	if len(errors) > 0 {
		fmt.Println("\nTip: Address the validation issues above to improve your agent configuration.")
		fmt.Println("   Most issues are suggestions to help optimize your agents' performance.")
	} else {
		fmt.Println("\nYour configuration looks great.")
	}
}
