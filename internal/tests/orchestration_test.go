package tests

import (
	"testing"

	"github.com/agenticgokit/agenticgokit/core"
)

func TestValidateOrchestrationConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      core.OrchestrationConfigToml
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid route mode",
			config: core.OrchestrationConfigToml{
				Mode:           "route",
				TimeoutSeconds: 30,
			},
			expectError: false,
		},
		{
			name: "valid collaborative mode",
			config: core.OrchestrationConfigToml{
				Mode:           "collaborative",
				TimeoutSeconds: 30,
			},
			expectError: false,
		},
		{
			name: "valid sequential mode",
			config: core.OrchestrationConfigToml{
				Mode:             "sequential",
				TimeoutSeconds:   30,
				SequentialAgents: []string{"agent1", "agent2"},
			},
			expectError: false,
		},
		{
			name: "valid loop mode",
			config: core.OrchestrationConfigToml{
				Mode:           "loop",
				TimeoutSeconds: 30,
				MaxIterations:  5,
				LoopAgent:      "agent1",
			},
			expectError: false,
		},
		{
			name: "valid mixed mode with both agent types",
			config: core.OrchestrationConfigToml{
				Mode:                "mixed",
				TimeoutSeconds:      30,
				CollaborativeAgents: []string{"agent1"},
				SequentialAgents:    []string{"agent2", "agent3"},
			},
			expectError: false,
		},
		{
			name: "valid mixed mode with only collaborative agents",
			config: core.OrchestrationConfigToml{
				Mode:                "mixed",
				TimeoutSeconds:      30,
				CollaborativeAgents: []string{"agent1", "agent2"},
			},
			expectError: false,
		},
		{
			name: "valid mixed mode with only sequential agents",
			config: core.OrchestrationConfigToml{
				Mode:             "mixed",
				TimeoutSeconds:   30,
				SequentialAgents: []string{"agent1", "agent2"},
			},
			expectError: false,
		},
		{
			name: "missing mode",
			config: core.OrchestrationConfigToml{
				TimeoutSeconds: 30,
			},
			expectError: true,
			errorMsg:    "orchestration mode is required",
		},
		{
			name: "invalid mode",
			config: core.OrchestrationConfigToml{
				Mode:           "invalid",
				TimeoutSeconds: 30,
			},
			expectError: true,
			errorMsg:    "invalid orchestration mode 'invalid'",
		},
		{
			name: "sequential mode missing agents",
			config: core.OrchestrationConfigToml{
				Mode:           "sequential",
				TimeoutSeconds: 30,
			},
			expectError: true,
			errorMsg:    "sequential orchestration requires 'sequential_agents' array",
		},
		{
			name: "loop mode missing agent",
			config: core.OrchestrationConfigToml{
				Mode:           "loop",
				TimeoutSeconds: 30,
				MaxIterations:  5,
			},
			expectError: true,
			errorMsg:    "loop orchestration requires 'loop_agent' string",
		},
		{
			name: "mixed mode missing both agent types",
			config: core.OrchestrationConfigToml{
				Mode:           "mixed",
				TimeoutSeconds: 30,
			},
			expectError: true,
			errorMsg:    "mixed orchestration requires either 'collaborative_agents' or 'sequential_agents'",
		},
		{
			name: "invalid timeout",
			config: core.OrchestrationConfigToml{
				Mode:           "route",
				TimeoutSeconds: 0,
			},
			expectError: true,
			errorMsg:    "orchestration timeout_seconds must be positive",
		},
		{
			name: "invalid max iterations for loop mode",
			config: core.OrchestrationConfigToml{
				Mode:           "loop",
				TimeoutSeconds: 30,
				MaxIterations:  0,
				LoopAgent:      "agent1",
			},
			expectError: true,
			errorMsg:    "orchestration max_iterations must be positive for loop mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &core.Config{
				Orchestration: tt.config,
			}

			err := config.ValidateOrchestrationConfig()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
