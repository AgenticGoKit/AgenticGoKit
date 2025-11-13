package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCommand(t *testing.T) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "agentcli-validate-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	t.Run("ValidateProjectStructure", func(t *testing.T) {
		results := &ValidationResults{
			ConfigPath: "agentflow.toml",
			Level:      "standard",
			Scope:      "structure-only",
		}

		// Test missing files
		validateProjectStructure(results)
		
		assert.Greater(t, len(results.Errors), 0, "Should have errors for missing files")
		
		// Check for specific missing file errors
		hasConfigError := false
		hasAgentsDirError := false
		for _, err := range results.Errors {
			if err.Field == "agentflow.toml" {
				hasConfigError = true
			}
			if err.Field == "agents/" {
				hasAgentsDirError = true
			}
		}
		assert.True(t, hasConfigError, "Should report missing agentflow.toml")
		assert.True(t, hasAgentsDirError, "Should report missing agents/ directory")
	})

	t.Run("ValidateWithValidProject", func(t *testing.T) {
		// Create minimal valid project structure
		err := os.WriteFile("agentflow.toml", []byte(`
[agent_flow]
name = "test-project"
version = "1.0.0"
provider = "openai"

[agents.test_agent]
role = "test_role"
description = "Test agent"
system_prompt = "You are a test agent"
capabilities = ["testing"]
enabled = true
`), 0644)
		require.NoError(t, err)

		err = os.WriteFile("go.mod", []byte(`
module test-project

go 1.21

require github.com/agenticgokit/agenticgokit v0.1.0
`), 0644)
		require.NoError(t, err)

		err = os.WriteFile("main.go", []byte(`package main

func main() {
	// Test main
}
`), 0644)
		require.NoError(t, err)

		err = os.Mkdir("agents", 0755)
		require.NoError(t, err)

		err = os.WriteFile("agents/test_agent.go", []byte(`package agents

// Test agent implementation
`), 0644)
		require.NoError(t, err)

		results := &ValidationResults{
			ConfigPath: "agentflow.toml",
			Level:      "standard",
			Scope:      "all",
		}

		validateComplete(results)
		
		// Should have fewer errors now
		assert.Equal(t, 0, len(results.Errors), "Should have no critical errors with valid project")
	})
}

func TestValidationIssueTypes(t *testing.T) {
	t.Run("ValidationIssueCreation", func(t *testing.T) {
		issue := ValidationIssue{
			Type:       "error",
			Code:       "TEST_ERROR",
			Field:      "test.field",
			Message:    "Test error message",
			Suggestion: "Test suggestion",
			Severity:   "high",
			Fixable:    true,
		}

		assert.Equal(t, "error", issue.Type)
		assert.Equal(t, "TEST_ERROR", issue.Code)
		assert.Equal(t, "test.field", issue.Field)
		assert.True(t, issue.Fixable)
	})

	t.Run("ValidationResultsSummary", func(t *testing.T) {
		results := &ValidationResults{
			Errors: []ValidationIssue{
				{Type: "error", Fixable: true},
				{Type: "error", Fixable: false},
			},
			Warnings: []ValidationIssue{
				{Type: "warning", Fixable: true},
			},
			Suggestions: []ValidationIssue{
				{Type: "suggestion", Fixable: false},
			},
		}

		// Calculate summary
		results.Summary = ValidationSummary{
			ErrorCount:      len(results.Errors),
			WarningCount:    len(results.Warnings),
			SuggestionCount: len(results.Suggestions),
			TotalIssues:     len(results.Errors) + len(results.Warnings) + len(results.Suggestions),
			IsValid:         len(results.Errors) == 0,
		}

		// Count fixable issues
		for _, issue := range results.Errors {
			if issue.Fixable {
				results.Summary.CanAutoFix++
			}
		}
		for _, issue := range results.Warnings {
			if issue.Fixable {
				results.Summary.CanAutoFix++
			}
		}

		assert.Equal(t, 2, results.Summary.ErrorCount)
		assert.Equal(t, 1, results.Summary.WarningCount)
		assert.Equal(t, 1, results.Summary.SuggestionCount)
		assert.Equal(t, 4, results.Summary.TotalIssues)
		assert.False(t, results.Summary.IsValid)
		assert.Equal(t, 2, results.Summary.CanAutoFix)
	})
}

func TestPathExists(t *testing.T) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "path-exists-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test existing file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	assert.True(t, pathExists(testFile), "Should detect existing file")
	assert.False(t, pathExists(filepath.Join(tempDir, "nonexistent.txt")), "Should not detect non-existent file")

	// Test existing directory
	testDir := filepath.Join(tempDir, "testdir")
	err = os.Mkdir(testDir, 0755)
	require.NoError(t, err)

	assert.True(t, pathExists(testDir), "Should detect existing directory")
}

func TestIsAgenticGoKitProject(t *testing.T) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "project-detection-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Initially should not be detected as AgenticGoKit project
	assert.False(t, isAgenticGoKitProject(), "Empty directory should not be detected as AgenticGoKit project")

	// Add agentflow.toml
	err = os.WriteFile("agentflow.toml", []byte("[agent_flow]\nname = \"test\""), 0644)
	require.NoError(t, err)

	assert.True(t, isAgenticGoKitProject(), "Directory with agentflow.toml should be detected")

	// Remove agentflow.toml and add agents directory
	err = os.Remove("agentflow.toml")
	require.NoError(t, err)

	err = os.Mkdir("agents", 0755)
	require.NoError(t, err)

	assert.True(t, isAgenticGoKitProject(), "Directory with agents/ should be detected")

	// Remove agents directory and add go.mod with dependency
	err = os.Remove("agents")
	require.NoError(t, err)

	err = os.WriteFile("go.mod", []byte(`
module test

require github.com/agenticgokit/agenticgokit v0.1.0
`), 0644)
	require.NoError(t, err)

	assert.True(t, isAgenticGoKitProject(), "Directory with AgenticGoKit dependency should be detected")
}

func TestContainsAgenticGoKitDependency(t *testing.T) {
	// Create temporary test file
	tempFile, err := os.CreateTemp("", "go.mod")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Test with AgenticGoKit dependency
	content := `module test-project

go 1.21

require (
	github.com/agenticgokit/agenticgokit v0.1.0
	github.com/spf13/cobra v1.7.0
)
`
	_, err = tempFile.WriteString(content)
	require.NoError(t, err)
	tempFile.Close()

	assert.True(t, containsAgenticGoKitDependency(tempFile.Name()), "Should detect AgenticGoKit dependency")

	// Test without AgenticGoKit dependency
	tempFile2, err := os.CreateTemp("", "go.mod")
	require.NoError(t, err)
	defer os.Remove(tempFile2.Name())

	content2 := `module test-project

go 1.21

require (
	github.com/spf13/cobra v1.7.0
)
`
	_, err = tempFile2.WriteString(content2)
	require.NoError(t, err)
	tempFile2.Close()

	assert.False(t, containsAgenticGoKitDependency(tempFile2.Name()), "Should not detect AgenticGoKit dependency")

	// Test non-existent file
	assert.False(t, containsAgenticGoKitDependency("nonexistent.mod"), "Should handle non-existent file")
}
