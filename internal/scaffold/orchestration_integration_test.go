package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldOrchestrationGeneration(t *testing.T) {
	tests := []struct {
		name                    string
		config                  ProjectConfig
		expectedOrchestration   string
		expectedMainGoContent   []string
		unexpectedMainGoContent []string
	}{
		{
			name: "sequential orchestration",
			config: ProjectConfig{
				Name:                 "sequential-test",
				NumAgents:            3,
				Provider:             "mock",
				OrchestrationMode:    "sequential",
				SequentialAgents:     []string{"agent1", "agent2", "agent3"},
				OrchestrationTimeout: 30,
				MemoryEnabled:        true,
				MemoryProvider:       "memory",
				EmbeddingProvider:    "dummy",
				EmbeddingModel:       "dummy",
				EmbeddingDimensions:  768,
			},
			expectedOrchestration: `[orchestration]
mode = "sequential"
timeout_seconds = 30
sequential_agents = ["agent1", "agent2", "agent3"]`,
			expectedMainGoContent: []string{
				"core.NewRunnerFromConfig(\"agentflow.toml\")",
				"runner.RegisterAgent(name, handler)",
			},
			unexpectedMainGoContent: []string{
				"core.NewRunnerWithOrchestration",
				"OrchestrationSequential",
				"CreateCollaborativeRunner",
			},
		},
		{
			name: "loop orchestration",
			config: ProjectConfig{
				Name:                 "loop-test",
				NumAgents:            3,
				Provider:             "mock",
				OrchestrationMode:    "loop",
				LoopAgent:            "agent1",
				MaxIterations:        5,
				OrchestrationTimeout: 30,
				MemoryEnabled:        true,
				MemoryProvider:       "memory",
				EmbeddingProvider:    "dummy",
				EmbeddingModel:       "dummy",
				EmbeddingDimensions:  768,
			},
			expectedOrchestration: `[orchestration]
mode = "loop"
timeout_seconds = 30
loop_agent = "agent1"
max_iterations = 5`,
			expectedMainGoContent: []string{
				"core.NewRunnerFromConfig(\"agentflow.toml\")",
				"runner.RegisterAgent(name, handler)",
			},
			unexpectedMainGoContent: []string{
				"core.NewRunnerWithOrchestration",
				"OrchestrationLoop",
				"CreateCollaborativeRunner",
			},
		},
		{
			name: "mixed orchestration",
			config: ProjectConfig{
				Name:                 "mixed-test",
				NumAgents:            4,
				Provider:             "mock",
				OrchestrationMode:    "mixed",
				CollaborativeAgents:  []string{"agent1"},
				SequentialAgents:     []string{"agent2", "agent3", "agent4"},
				OrchestrationTimeout: 30,
				MemoryEnabled:        true,
				MemoryProvider:       "memory",
				EmbeddingProvider:    "dummy",
				EmbeddingModel:       "dummy",
				EmbeddingDimensions:  768,
			},
			expectedOrchestration: `[orchestration]
mode = "mixed"
timeout_seconds = 30
collaborative_agents = ["agent1"]
sequential_agents = ["agent2", "agent3", "agent4"]`,
			expectedMainGoContent: []string{
				"core.NewRunnerFromConfig(\"agentflow.toml\")",
				"runner.RegisterAgent(name, handler)",
			},
			unexpectedMainGoContent: []string{
				"core.NewRunnerWithOrchestration",
				"OrchestrationMixed",
				"CreateCollaborativeRunner",
			},
		},
		{
			name: "collaborative orchestration",
			config: ProjectConfig{
				Name:                 "collaborative-test",
				NumAgents:            3,
				Provider:             "mock",
				OrchestrationMode:    "collaborative",
				OrchestrationTimeout: 30,
				MemoryEnabled:        true,
				MemoryProvider:       "memory",
				EmbeddingProvider:    "dummy",
				EmbeddingModel:       "dummy",
				EmbeddingDimensions:  768,
			},
			expectedOrchestration: `[orchestration]
mode = "collaborative"
timeout_seconds = 30`,
			expectedMainGoContent: []string{
				"core.NewRunnerFromConfig(\"agentflow.toml\")",
				"runner.RegisterAgent(name, handler)",
			},
			unexpectedMainGoContent: []string{
				"core.NewRunnerWithOrchestration",
				"OrchestrationCollaborate",
				"CreateCollaborativeRunner",
			},
		},
		{
			name: "route orchestration",
			config: ProjectConfig{
				Name:                 "route-test",
				NumAgents:            3,
				Provider:             "mock",
				OrchestrationMode:    "route",
				OrchestrationTimeout: 30,
				MemoryEnabled:        true,
				MemoryProvider:       "memory",
				EmbeddingProvider:    "dummy",
				EmbeddingModel:       "dummy",
				EmbeddingDimensions:  768,
			},
			expectedOrchestration: `[orchestration]
mode = "route"
timeout_seconds = 30`,
			expectedMainGoContent: []string{
				"core.NewRunnerFromConfig(\"agentflow.toml\")",
				"runner.RegisterAgent(name, handler)",
			},
			unexpectedMainGoContent: []string{
				"core.NewRunnerWithOrchestration",
				"OrchestrationRoute",
				"CreateCollaborativeRunner",
			},
		},
		{
			name: "mixed orchestration with default agent split",
			config: ProjectConfig{
				Name:                 "mixed-default-test",
				NumAgents:            4,
				Provider:             "mock",
				OrchestrationMode:    "mixed",
				OrchestrationTimeout: 30,
				MemoryEnabled:        true,
				MemoryProvider:       "memory",
				EmbeddingProvider:    "dummy",
				EmbeddingModel:       "dummy",
				EmbeddingDimensions:  768,
				// No explicit CollaborativeAgents or SequentialAgents - should use defaults
			},
			expectedOrchestration: `[orchestration]
mode = "mixed"
timeout_seconds = 30
collaborative_agents = ["agent1"]
sequential_agents = ["agent2", "agent3", "agent4"]`,
			expectedMainGoContent: []string{
				"core.NewRunnerFromConfig(\"agentflow.toml\")",
				"runner.RegisterAgent(name, handler)",
			},
			unexpectedMainGoContent: []string{
				"core.NewRunnerWithOrchestration",
				"OrchestrationMixed",
				"CreateCollaborativeRunner",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tmpDir := t.TempDir()
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get working directory: %v", err)
			}
			defer os.Chdir(originalDir)

			err = os.Chdir(tmpDir)
			if err != nil {
				t.Fatalf("Failed to change to temp directory: %v", err)
			}

			// Create the project
			err = CreateAgentProject(tt.config)
			if err != nil {
				t.Fatalf("Failed to create project: %v", err)
			}

			// Verify agentflow.toml contains correct orchestration configuration
			configPath := filepath.Join(tt.config.Name, "agentflow.toml")
			configContent, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("Failed to read agentflow.toml: %v", err)
			}

			configStr := string(configContent)
			if !strings.Contains(configStr, tt.expectedOrchestration) {
				t.Errorf("Expected orchestration configuration not found in agentflow.toml.\nExpected:\n%s\n\nActual config:\n%s", 
					tt.expectedOrchestration, configStr)
			}

			// Verify main.go uses NewRunnerFromConfig() instead of hardcoded orchestration
			mainGoPath := filepath.Join(tt.config.Name, "main.go")
			mainGoContent, err := os.ReadFile(mainGoPath)
			if err != nil {
				t.Fatalf("Failed to read main.go: %v", err)
			}

			mainGoStr := string(mainGoContent)

			// Check for expected content
			for _, expected := range tt.expectedMainGoContent {
				if !strings.Contains(mainGoStr, expected) {
					t.Errorf("Expected content not found in main.go: %s", expected)
				}
			}

			// Check for unexpected content (hardcoded orchestration)
			for _, unexpected := range tt.unexpectedMainGoContent {
				if strings.Contains(mainGoStr, unexpected) {
					t.Errorf("Unexpected hardcoded orchestration content found in main.go: %s", unexpected)
				}
			}

			// Verify the project compiles (basic syntax check)
			// This is a simple check - we're not actually running go build
			// but we can check for basic Go syntax issues
			if !strings.Contains(mainGoStr, "package main") {
				t.Errorf("main.go does not contain 'package main'")
			}
			if !strings.Contains(mainGoStr, "func main()") {
				t.Errorf("main.go does not contain 'func main()'")
			}
		})
	}
}

func TestGenerateOrchestrationConfig(t *testing.T) {
	tests := []struct {
		name           string
		config         ProjectConfig
		expectedConfig string
	}{
		{
			name: "sequential mode with explicit agents",
			config: ProjectConfig{
				OrchestrationMode:    "sequential",
				OrchestrationTimeout: 30,
				SequentialAgents:     []string{"agent1", "agent2", "agent3"},
			},
			expectedConfig: `
[orchestration]
mode = "sequential"
timeout_seconds = 30
sequential_agents = ["agent1", "agent2", "agent3"]
`,
		},
		{
			name: "sequential mode with default agents",
			config: ProjectConfig{
				OrchestrationMode:    "sequential",
				OrchestrationTimeout: 30,
				NumAgents:            3,
			},
			expectedConfig: `
[orchestration]
mode = "sequential"
timeout_seconds = 30
sequential_agents = ["agent1", "agent2", "agent3"]
`,
		},
		{
			name: "loop mode",
			config: ProjectConfig{
				OrchestrationMode:    "loop",
				OrchestrationTimeout: 30,
				LoopAgent:            "agent1",
				MaxIterations:        5,
			},
			expectedConfig: `
[orchestration]
mode = "loop"
timeout_seconds = 30
loop_agent = "agent1"
max_iterations = 5
`,
		},
		{
			name: "loop mode with default agent",
			config: ProjectConfig{
				OrchestrationMode:    "loop",
				OrchestrationTimeout: 30,
				MaxIterations:        5,
			},
			expectedConfig: `
[orchestration]
mode = "loop"
timeout_seconds = 30
loop_agent = "agent1"
max_iterations = 5
`,
		},
		{
			name: "mixed mode with explicit agents",
			config: ProjectConfig{
				OrchestrationMode:    "mixed",
				OrchestrationTimeout: 30,
				CollaborativeAgents:  []string{"agent1"},
				SequentialAgents:     []string{"agent2", "agent3"},
			},
			expectedConfig: `
[orchestration]
mode = "mixed"
timeout_seconds = 30
collaborative_agents = ["agent1"]
sequential_agents = ["agent2", "agent3"]
`,
		},
		{
			name: "mixed mode with default agent split",
			config: ProjectConfig{
				OrchestrationMode:    "mixed",
				OrchestrationTimeout: 30,
				NumAgents:            4,
			},
			expectedConfig: `
[orchestration]
mode = "mixed"
timeout_seconds = 30
collaborative_agents = ["agent1"]
sequential_agents = ["agent2", "agent3", "agent4"]
`,
		},
		{
			name: "collaborative mode",
			config: ProjectConfig{
				OrchestrationMode:    "collaborative",
				OrchestrationTimeout: 30,
			},
			expectedConfig: `
[orchestration]
mode = "collaborative"
timeout_seconds = 30
`,
		},
		{
			name: "route mode",
			config: ProjectConfig{
				OrchestrationMode:    "route",
				OrchestrationTimeout: 30,
			},
			expectedConfig: `
[orchestration]
mode = "route"
timeout_seconds = 30
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateOrchestrationConfig(tt.config)
			
			// Normalize whitespace for comparison
			expected := strings.TrimSpace(tt.expectedConfig)
			actual := strings.TrimSpace(result)
			
			if actual != expected {
				t.Errorf("Generated orchestration config does not match expected.\nExpected:\n%s\n\nActual:\n%s", expected, actual)
			}
		})
	}
}