package utils

import (
	"strings"
	"testing"
)

func TestCreateSystemPrompt(t *testing.T) {
	agent := AgentInfo{
		Name:        "test_agent",
		FileName:    "test_agent.go",
		DisplayName: "TestAgent",
		Purpose:     "tests prompt generation functionality",
		Role:        "sequential",
	}

	tests := []struct {
		name              string
		agentIndex        int
		totalAgents       int
		orchestrationMode string
		expectContains    []string
	}{
		{
			name:              "sequential first agent",
			agentIndex:        0,
			totalAgents:       3,
			orchestrationMode: "sequential",
			expectContains: []string{
				"You are TestAgent, the first agent in a sequential multi-agent system",
				"Your primary role is to analyze and process the initial user query comprehensively",
				"Tool Usage Strategy",
				"Response Quality Standards",
				"Sequential Mode Guidelines",
			},
		},
		{
			name:              "sequential last agent",
			agentIndex:        2,
			totalAgents:       3,
			orchestrationMode: "sequential",
			expectContains: []string{
				"You are TestAgent, the final agent in a sequential multi-agent system",
				"Your role is to synthesize and present the final response to the user",
				"Tool Usage Strategy",
				"Response Quality Standards",
			},
		},
		{
			name:              "collaborative agent",
			agentIndex:        1,
			totalAgents:       3,
			orchestrationMode: "collaborative",
			expectContains: []string{
				"You are TestAgent, one of 3 agents working collaboratively",
				"Core Responsibilities",
				"Collaborative Guidelines",
				"Tool Usage Strategy",
			},
		},
		{
			name:              "loop agent",
			agentIndex:        0,
			totalAgents:       1,
			orchestrationMode: "loop",
			expectContains: []string{
				"You are TestAgent, operating in an iterative loop mode",
				"Iterative Processing Guidelines",
				"Loop Mode Strategy",
				"Tool Usage Strategy",
			},
		},
		{
			name:              "route agent",
			agentIndex:        0,
			totalAgents:       1,
			orchestrationMode: "route",
			expectContains: []string{
				"You are TestAgent, an intelligent agent in a multi-agent system",
				"Core Processing Guidelines",
				"Routing Strategy",
				"Tool Usage Strategy",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := CreateSystemPrompt(agent, tt.agentIndex, tt.totalAgents, tt.orchestrationMode)

			// Check that prompt is not empty
			if len(prompt) == 0 {
				t.Errorf("CreateSystemPrompt() returned empty prompt")
				return
			}

			// Check that prompt contains expected content
			for _, expected := range tt.expectContains {
				if !strings.Contains(prompt, expected) {
					t.Errorf("CreateSystemPrompt() missing expected content: %q", expected)
				}
			}

			// Check that prompt is comprehensive (longer than basic prompts)
			if len(prompt) < 500 {
				t.Errorf("CreateSystemPrompt() returned prompt that seems too short (%d chars), expected comprehensive prompt", len(prompt))
			}

			// Check that prompt includes purpose if specified
			if agent.Purpose != "" && !strings.Contains(prompt, agent.Purpose) {
				t.Errorf("CreateSystemPrompt() should include agent purpose: %q", agent.Purpose)
			}
		})
	}
}

func TestCreateSystemPromptComprehensiveness(t *testing.T) {
	agent := AgentInfo{
		Name:        "research_agent",
		FileName:    "research_agent.go",
		DisplayName: "ResearchAgent",
		Purpose:     "conducts comprehensive research and analysis",
		Role:        "sequential",
	}

	prompt := CreateSystemPrompt(agent, 0, 2, "sequential")

	// Check for comprehensive content sections
	expectedSections := []string{
		"1. Thoroughly understand",
		"2. Gather relevant",
		"3. Analyze the information",
		"4. Present your findings",
		"5. Include any relevant",
		"6. Flag any uncertainties",
		"7. Provide substantial content",
		"Tool Usage Strategy:",
		"Response Quality Standards:",
		"Sequential Mode Guidelines:",
	}

	for _, section := range expectedSections {
		if !strings.Contains(prompt, section) {
			t.Errorf("Enhanced prompt missing expected section: %q", section)
		}
	}

	// Verify that this is much more comprehensive than a basic prompt
	basicPromptLength := len("You are ResearchAgent, a specialized AI agent whose purpose is to conducts comprehensive research and analysis. You are the first agent in a sequential workflow. Process the user's input and prepare it for the next agent. Always provide clear, helpful, and accurate responses.")

	if len(prompt) <= basicPromptLength*2 {
		t.Errorf("Enhanced prompt should be significantly longer than basic prompt. Got %d chars, expected >%d", len(prompt), basicPromptLength*2)
	}

	t.Logf("Enhanced prompt length: %d characters", len(prompt))
}
