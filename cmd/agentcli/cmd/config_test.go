package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agenticgokit/agenticgokit/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigGenerate(t *testing.T) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "config-generate-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	t.Run("AnalyzeAgentFiles", func(t *testing.T) {
		// Create agents directory with test files
		err := os.Mkdir("agents", 0755)
		require.NoError(t, err)

		// Create test agent files
		agentFiles := map[string]string{
			"research_agent.go": `package agents

type ResearchAgent struct {
	// Agent implementation
}

func (a *ResearchAgent) Run() {
	// Implementation
}
`,
			"writer_agent.go": `package agents

type WriterAgent struct {
	// Agent implementation  
}

func (a *WriterAgent) Run() {
	// Implementation
}
`,
		}

		for filename, content := range agentFiles {
			err := os.WriteFile(filepath.Join("agents", filename), []byte(content), 0644)
			require.NoError(t, err)
		}

		// Analyze agent files
		configs, err := analyzeAgentFiles("agents")
		require.NoError(t, err)

		assert.Len(t, configs, 2, "Should find 2 agent configurations")
		assert.Contains(t, configs, "research_agent", "Should find research_agent")
		assert.Contains(t, configs, "writer_agent", "Should find writer_agent")

		// Check extracted configuration
		researchConfig := configs["research_agent"]
		assert.Equal(t, "research_agent", researchConfig.Name)
		assert.Equal(t, "research_agent_role", researchConfig.Role)
		assert.True(t, researchConfig.Enabled)
		assert.Equal(t, 30, researchConfig.Timeout)
	})

	t.Run("GenerateConfigFromAgents", func(t *testing.T) {
		// Create test agent configurations
		agentConfigs := map[string]AgentConfigExtract{
			"test_agent": {
				Name:         "test_agent",
				Role:         "test_role",
				Description:  "Test agent description",
				Capabilities: []string{"testing", "validation"},
				SystemPrompt: "You are a test agent",
				Enabled:      true,
				Timeout:      45,
			},
		}

		// Generate configuration
		config := generateConfigFromAgents(agentConfigs)

		assert.Equal(t, "generated-project", config.AgentFlow.Name)
		assert.Equal(t, "1.0.0", config.AgentFlow.Version)
		assert.Equal(t, "openai", config.AgentFlow.Provider)

		assert.Len(t, config.Agents, 1)
		assert.Contains(t, config.Agents, "test_agent")

		testAgent := config.Agents["test_agent"]
		assert.Equal(t, "test_role", testAgent.Role)
		assert.Equal(t, "Test agent description", testAgent.Description)
		assert.Equal(t, "You are a test agent", testAgent.SystemPrompt)
		assert.Equal(t, []string{"testing", "validation"}, testAgent.Capabilities)
		assert.True(t, testAgent.Enabled)
		assert.Equal(t, 45, testAgent.Timeout)
	})
}

func TestConfigMigration(t *testing.T) {
	t.Run("AnalyzeMigrationNeeds", func(t *testing.T) {
		// Test configuration that needs migration
		config := &core.Config{}
		config.AgentFlow.Name = "test-project"
		// Missing version field

		needed, changes := analyzeMigrationNeeds(config)

		assert.True(t, needed, "Should detect migration is needed")
		assert.Greater(t, len(changes), 0, "Should have migration changes")
		
		// Check for specific change
		hasVersionChange := false
		for _, change := range changes {
			if change == "Add version field to agent_flow section" {
				hasVersionChange = true
				break
			}
		}
		assert.True(t, hasVersionChange, "Should detect missing version field")
	})

	t.Run("ApplyMigration", func(t *testing.T) {
		// Test configuration before migration
		config := &core.Config{}
		config.AgentFlow.Name = "test-project"
		// Missing version field

		// Apply migration
		migratedConfig := applyMigration(config, "1.0")

		assert.Equal(t, "1.0.0", migratedConfig.AgentFlow.Version, "Should add version field")
		assert.Equal(t, "test-project", migratedConfig.AgentFlow.Name, "Should preserve existing fields")
	})
}

func TestConfigOptimization(t *testing.T) {
	t.Run("AnalyzeOptimizations", func(t *testing.T) {
		// Create configuration with optimization opportunities
		config := &core.Config{}
		config.Agents = map[string]core.AgentConfig{
			"slow_agent": {
				Role:        "slow_role",
				Description: "Slow agent",
				Timeout:     120, // High timeout
				Enabled:     true,
			},
			"fast_agent": {
				Role:        "fast_role", 
				Description: "Fast agent",
				Timeout:     30, // Normal timeout
				Enabled:     true,
			},
		}

		optimizations := analyzeOptimizations(config, "performance")

		assert.Greater(t, len(optimizations), 0, "Should find optimization opportunities")
		
		// Check for timeout optimization
		hasTimeoutOpt := false
		for _, opt := range optimizations {
			if opt.Area == "Performance" && opt.Description == "Agent slow_agent has high timeout (120 seconds)" {
				hasTimeoutOpt = true
				break
			}
		}
		assert.True(t, hasTimeoutOpt, "Should detect high timeout optimization")
	})

	t.Run("ApplyOptimizations", func(t *testing.T) {
		// Create configuration with high timeout
		config := &core.Config{}
		config.Agents = map[string]core.AgentConfig{
			"slow_agent": {
				Role:        "slow_role",
				Description: "Slow agent", 
				Timeout:     120, // High timeout
				Enabled:     true,
			},
		}

		optimizations := []OptimizationOpportunity{
			{
				Area:        "Performance",
				Description: "Agent slow_agent has high timeout (120 seconds)",
				Impact:      "Reduce timeout for better responsiveness",
			},
		}

		// Apply optimizations
		optimizedConfig := applyOptimizations(config, optimizations)

		assert.Equal(t, 30, optimizedConfig.Agents["slow_agent"].Timeout, "Should reduce timeout to 30 seconds")
	})
}

func TestCopyFile(t *testing.T) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "copy-file-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create source file
	srcFile := filepath.Join(tempDir, "source.txt")
	srcContent := "test content for copy"
	err = os.WriteFile(srcFile, []byte(srcContent), 0644)
	require.NoError(t, err)

	// Copy file
	dstFile := filepath.Join(tempDir, "destination.txt")
	err = copyFile(srcFile, dstFile)
	require.NoError(t, err)

	// Verify copy
	dstContent, err := os.ReadFile(dstFile)
	require.NoError(t, err)

	assert.Equal(t, srcContent, string(dstContent), "Copied file should have same content")

	// Test copying non-existent file
	err = copyFile(filepath.Join(tempDir, "nonexistent.txt"), filepath.Join(tempDir, "dest2.txt"))
	assert.Error(t, err, "Should error when copying non-existent file")
}

func TestAgentConfigExtract(t *testing.T) {
	t.Run("AgentConfigExtractCreation", func(t *testing.T) {
		config := AgentConfigExtract{
			Name:         "test_agent",
			Role:         "test_role",
			Description:  "Test agent for validation",
			Capabilities: []string{"testing", "validation", "analysis"},
			SystemPrompt: "You are a comprehensive test agent",
			Enabled:      true,
			Timeout:      60,
		}

		assert.Equal(t, "test_agent", config.Name)
		assert.Equal(t, "test_role", config.Role)
		assert.Equal(t, "Test agent for validation", config.Description)
		assert.Len(t, config.Capabilities, 3)
		assert.Contains(t, config.Capabilities, "testing")
		assert.Contains(t, config.Capabilities, "validation")
		assert.Contains(t, config.Capabilities, "analysis")
		assert.Equal(t, "You are a comprehensive test agent", config.SystemPrompt)
		assert.True(t, config.Enabled)
		assert.Equal(t, 60, config.Timeout)
	})
}

func TestOptimizationOpportunity(t *testing.T) {
	t.Run("OptimizationOpportunityCreation", func(t *testing.T) {
		opt := OptimizationOpportunity{
			Area:        "Performance",
			Description: "High timeout detected in agent configuration",
			Impact:      "Reducing timeout will improve system responsiveness",
		}

		assert.Equal(t, "Performance", opt.Area)
		assert.Equal(t, "High timeout detected in agent configuration", opt.Description)
		assert.Equal(t, "Reducing timeout will improve system responsiveness", opt.Impact)
	})
}

// Integration test for the full config generate workflow
func TestConfigGenerateIntegration(t *testing.T) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "config-generate-integration")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Create agents directory with realistic agent files
	err = os.Mkdir("agents", 0755)
	require.NoError(t, err)

	agentContent := `package agents

import (
	"context"
	"fmt"
	"github.com/agenticgokit/agenticgokit/core"
)

type DataProcessorHandler struct {
	llm core.ModelProvider
}

func NewDataProcessor(llmProvider core.ModelProvider) *DataProcessorHandler {
	return &DataProcessorHandler{
		llm: llmProvider,
	}
}

func (a *DataProcessorHandler) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
	// System prompt would be extracted from here
	systemPrompt := "You are a data processing specialist"
	
	// Implementation details...
	return core.AgentResult{}, nil
}
`

	err = os.WriteFile("agents/data_processor.go", []byte(agentContent), 0644)
	require.NoError(t, err)

	// Test the analysis
	configs, err := analyzeAgentFiles("agents")
	require.NoError(t, err)

	assert.Len(t, configs, 1)
	assert.Contains(t, configs, "data_processor")

	// Test configuration generation
	config := generateConfigFromAgents(configs)
	assert.NotNil(t, config)
	assert.Len(t, config.Agents, 1)
	assert.Contains(t, config.Agents, "data_processor")
}
