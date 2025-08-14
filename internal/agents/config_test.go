package agents

import "github.com/kunalkushwaha/agenticgokit/core"

import (
	"testing"
	"time"
)



func TestAgentConfigResolution(t *testing.T) {
	// Create a test configuration
	config := &Config{
		LLM: AgentLLMConfig{
			Provider:    "openai",
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   800,
			TimeoutSeconds: 30,
		},
		Agents: map[string]AgentConfig{
			"researcher": {
				Role:         "research_specialist",
				Description:  "Gathers comprehensive information",
				SystemPrompt: "You are a research specialist...",
				Capabilities: []string{"information_gathering", "fact_checking"},
				Enabled:      true,
				LLM: &AgentLLMConfig{
					Temperature: 0.3, // Override global temperature
					MaxTokens:   1200, // Override global max tokens
				},
				Timeout: 45,
			},
			"analyzer": {
				Role:         "analysis_specialist",
				Description:  "Provides deep insights",
				SystemPrompt: "You are an analysis specialist...",
				Capabilities: []string{"pattern_recognition", "trend_analysis"},
				Enabled:      true,
				// No LLM override - should use global settings
				Timeout: 30,
			},
		},
	}

	// Test resolving researcher config
	researcherConfig, err := config.ResolveAgentConfig("researcher")
	if err != nil {
		t.Fatalf("Failed to resolve researcher config: %v", err)
	}

	// Verify researcher config
	if researcherConfig.Name != "researcher" {
		t.Errorf("Expected name 'researcher', got '%s'", researcherConfig.Name)
	}
	if researcherConfig.Role != "research_specialist" {
		t.Errorf("Expected role 'research_specialist', got '%s'", researcherConfig.Role)
	}
	if researcherConfig.LLMConfig.Temperature != 0.3 {
		t.Errorf("Expected temperature 0.3, got %f", researcherConfig.LLMConfig.Temperature)
	}
	if researcherConfig.LLMConfig.MaxTokens != 1200 {
		t.Errorf("Expected max tokens 1200, got %d", researcherConfig.LLMConfig.MaxTokens)
	}
	if researcherConfig.LLMConfig.Provider != "openai" {
		t.Errorf("Expected provider 'openai', got '%s'", researcherConfig.LLMConfig.Provider)
	}
	if researcherConfig.LLMConfig.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", researcherConfig.LLMConfig.Model)
	}
	if researcherConfig.Timeout != 45*time.Second {
		t.Errorf("Expected timeout 45s, got %v", researcherConfig.Timeout)
	}

	// Test resolving analyzer config
	analyzerConfig, err := config.ResolveAgentConfig("analyzer")
	if err != nil {
		t.Fatalf("Failed to resolve analyzer config: %v", err)
	}

	// Verify analyzer config uses global LLM settings
	if analyzerConfig.LLMConfig.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7 (global), got %f", analyzerConfig.LLMConfig.Temperature)
	}
	if analyzerConfig.LLMConfig.MaxTokens != 800 {
		t.Errorf("Expected max tokens 800 (global), got %d", analyzerConfig.LLMConfig.MaxTokens)
	}

	// Test non-existent agent
	_, err = config.ResolveAgentConfig("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent agent")
	}
}

func TestAgentConfigHelperMethods(t *testing.T) {
	config := &Config{
		Agents: map[string]AgentConfig{
			"enabled_agent": {
				Role:         "test_role",
				Capabilities: []string{"cap1", "cap2"},
				Enabled:      true,
			},
			"disabled_agent": {
				Role:    "test_role",
				Enabled: false,
			},
		},
	}

	// Test GetEnabledAgents
	enabled := config.GetEnabledAgents()
	if len(enabled) != 1 {
		t.Errorf("Expected 1 enabled agent, got %d", len(enabled))
	}
	if enabled[0] != "enabled_agent" {
		t.Errorf("Expected 'enabled_agent', got '%s'", enabled[0])
	}

	// Test IsAgentEnabled
	if !config.IsAgentEnabled("enabled_agent") {
		t.Error("Expected enabled_agent to be enabled")
	}
	if config.IsAgentEnabled("disabled_agent") {
		t.Error("Expected disabled_agent to be disabled")
	}
	if config.IsAgentEnabled("nonexistent") {
		t.Error("Expected nonexistent agent to be disabled")
	}

	// Test GetAgentCapabilities
	caps := config.GetAgentCapabilities("enabled_agent")
	if len(caps) != 2 {
		t.Errorf("Expected 2 capabilities, got %d", len(caps))
	}
	if caps[0] != "cap1" || caps[1] != "cap2" {
		t.Errorf("Expected ['cap1', 'cap2'], got %v", caps)
	}

	// Test capabilities for non-existent agent
	caps = config.GetAgentCapabilities("nonexistent")
	if caps != nil {
		t.Errorf("Expected nil capabilities for non-existent agent, got %v", caps)
	}
}