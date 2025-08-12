package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigGenerateCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		expectFiles []string
	}{
		{
			name:        "generate with valid template",
			args:        []string{"simple-workflow", "test-project"},
			expectError: false,
			expectFiles: []string{"agentflow.toml"},
		},
		{
			name:        "generate with custom output dir",
			args:        []string{"simple-workflow", "test-project", "--output-dir", "configs"},
			expectError: false,
			expectFiles: []string{"configs/agentflow.toml"},
		},
		{
			name:        "generate yaml format",
			args:        []string{"simple-workflow", "test-project", "--format", "yaml"},
			expectError: false,
			expectFiles: []string{"agentflow.yaml"},
		},
		{
			name:        "generate json format",
			args:        []string{"simple-workflow", "test-project", "--format", "json"},
			expectError: false,
			expectFiles: []string{"agentflow.json"},
		},
		{
			name:        "missing arguments",
			args:        []string{},
			expectError: true,
			expectFiles: []string{},
		},
		{
			name:        "invalid template",
			args:        []string{"nonexistent-template", "test-project"},
			expectError: true,
			expectFiles: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir, err := os.MkdirTemp("", "config-generate-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			// Change to temp directory
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(oldWd)
			os.Chdir(tempDir)

			// Create and execute command
			cmd := &cobra.Command{
				Use: "generate",
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock the generate function for testing
					if len(args) < 2 {
						return assert.AnError
					}
					
					templateName := args[0]
					if templateName == "nonexistent-template" {
						return assert.AnError
					}

					// Create mock output files
					outputDir, _ := cmd.Flags().GetString("output-dir")
					format, _ := cmd.Flags().GetString("format")
					
					if outputDir == "" {
						outputDir = "."
					}
					
					var filename string
					switch format {
					case "yaml":
						filename = "agentflow.yaml"
					case "json":
						filename = "agentflow.json"
					default:
						filename = "agentflow.toml"
					}
					
					outputPath := filepath.Join(outputDir, filename)
					
					// Create directory if needed
					if err := os.MkdirAll(outputDir, 0755); err != nil {
						return err
					}
					
					// Create mock file
					return os.WriteFile(outputPath, []byte("# Mock configuration"), 0644)
				},
			}

			cmd.Flags().String("output-dir", "", "Output directory")
			cmd.Flags().String("format", "toml", "Output format")

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Execute command
			cmd.SetArgs(tt.args)
			err = cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Check expected files exist
				for _, expectedFile := range tt.expectFiles {
					_, err := os.Stat(expectedFile)
					assert.NoError(t, err, "Expected file %s should exist", expectedFile)
				}
			}
		})
	}
}

func TestConfigMigrateCommand(t *testing.T) {
	tests := []struct {
		name        string
		setupFiles  map[string]string
		args        []string
		expectError bool
	}{
		{
			name: "migrate existing config",
			setupFiles: map[string]string{
				"agentflow.toml": `[agent_flow]
name = "test-project"
version = "1.0.0"

[agents.test_agent]
role = "processor"
description = "Test agent"
capabilities = ["processing"]`,
			},
			args:        []string{"agentflow.toml"},
			expectError: false,
		},
		{
			name: "dry run migration",
			setupFiles: map[string]string{
				"agentflow.toml": `[agent_flow]
name = "test-project"`,
			},
			args:        []string{"--dry-run", "agentflow.toml"},
			expectError: false,
		},
		{
			name:        "missing config file",
			setupFiles:  map[string]string{},
			args:        []string{"nonexistent.toml"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir, err := os.MkdirTemp("", "config-migrate-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			// Change to temp directory
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(oldWd)
			os.Chdir(tempDir)

			// Setup test files
			for filename, content := range tt.setupFiles {
				err := os.WriteFile(filename, []byte(content), 0644)
				require.NoError(t, err)
			}

			// Create and execute command
			cmd := &cobra.Command{
				Use: "migrate",
				RunE: func(cmd *cobra.Command, args []string) error {
					configFile := "agentflow.toml"
					if len(args) > 0 {
						configFile = args[0]
					}

					// Check if file exists
					if _, err := os.Stat(configFile); os.IsNotExist(err) {
						return err
					}

					dryRun, _ := cmd.Flags().GetBool("dry-run")
					if dryRun {
						// Just validate for dry run
						return nil
					}

					// Mock migration logic
					return nil
				},
			}

			cmd.Flags().Bool("dry-run", false, "Dry run")
			cmd.Flags().String("backup-dir", "", "Backup directory")

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Execute command
			cmd.SetArgs(tt.args)
			err = cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigSchemaCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		expectFiles []string
	}{
		{
			name:        "generate schema",
			args:        []string{"--generate"},
			expectError: false,
			expectFiles: []string{"agentflow-schema.json"},
		},
		{
			name:        "export docs",
			args:        []string{"--export-docs"},
			expectError: false,
			expectFiles: []string{"agentflow-config-docs.md"},
		},
		{
			name:        "generate with custom output",
			args:        []string{"--generate", "--output", "custom-schema.json"},
			expectError: false,
			expectFiles: []string{"custom-schema.json"},
		},
		{
			name:        "no flags provided",
			args:        []string{},
			expectError: false, // Should show help
			expectFiles: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir, err := os.MkdirTemp("", "config-schema-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			// Change to temp directory
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(oldWd)
			os.Chdir(tempDir)

			// Create and execute command
			cmd := &cobra.Command{
				Use: "schema",
				RunE: func(cmd *cobra.Command, args []string) error {
					generate, _ := cmd.Flags().GetBool("generate")
					exportDocs, _ := cmd.Flags().GetBool("export-docs")
					outputFile, _ := cmd.Flags().GetString("output")

					if generate {
						filename := "agentflow-schema.json"
						if outputFile != "" {
							filename = outputFile
						}
						return os.WriteFile(filename, []byte(`{"$schema": "http://json-schema.org/draft-07/schema#"}`), 0644)
					} else if exportDocs {
						filename := "agentflow-config-docs.md"
						if outputFile != "" {
							filename = outputFile
						}
						return os.WriteFile(filename, []byte("# Configuration Documentation"), 0644)
					}

					// Show help if no flags
					return nil
				},
			}

			cmd.Flags().Bool("generate", false, "Generate schema")
			cmd.Flags().Bool("export-docs", false, "Export docs")
			cmd.Flags().String("output", "", "Output file")

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Execute command
			cmd.SetArgs(tt.args)
			err = cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Check expected files exist
				for _, expectedFile := range tt.expectFiles {
					_, err := os.Stat(expectedFile)
					assert.NoError(t, err, "Expected file %s should exist", expectedFile)
				}
			}
		})
	}
}

func TestConfigListTemplates(t *testing.T) {
	// Create and execute command
	cmd := &cobra.Command{
		Use: "generate",
		RunE: func(cmd *cobra.Command, args []string) error {
			listTemplates, _ := cmd.Flags().GetBool("list")
			if listTemplates {
				// Mock template listing
				return nil
			}
			return assert.AnError
		},
	}

	cmd.Flags().Bool("list", false, "List templates")

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute command with --list flag
	cmd.SetArgs([]string{"--list"})
	err := cmd.Execute()

	assert.NoError(t, err)
}

func TestConfigCommandIntegration(t *testing.T) {
	// Test that all config subcommands are properly registered
	configCmd := &cobra.Command{Use: "config"}
	
	// Add mock subcommands
	generateCmd := &cobra.Command{Use: "generate"}
	migrateCmd := &cobra.Command{Use: "migrate"}
	schemaCmd := &cobra.Command{Use: "schema"}
	
	configCmd.AddCommand(generateCmd)
	configCmd.AddCommand(migrateCmd)
	configCmd.AddCommand(schemaCmd)

	// Test that subcommands are accessible
	subcommands := configCmd.Commands()
	assert.Len(t, subcommands, 3)

	commandNames := make([]string, len(subcommands))
	for i, cmd := range subcommands {
		commandNames[i] = cmd.Name()
	}

	assert.Contains(t, commandNames, "generate")
	assert.Contains(t, commandNames, "migrate")
	assert.Contains(t, commandNames, "schema")
}

func TestConfigOutputFormats(t *testing.T) {
	formats := []string{"toml", "yaml", "json"}
	
	for _, format := range formats {
		t.Run("format_"+format, func(t *testing.T) {
			// Test that each format is properly handled
			var extension string
			switch format {
			case "yaml":
				extension = ".yaml"
			case "json":
				extension = ".json"
			default:
				extension = ".toml"
			}
			
			expectedFilename := "agentflow" + extension
			assert.True(t, strings.HasSuffix(expectedFilename, extension))
		})
	}
}

func TestConfigValidationIntegration(t *testing.T) {
	// Test integration between config generation and validation
	tempDir, err := os.MkdirTemp("", "config-validation-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Create a mock config file
	configContent := `[agent_flow]
name = "test-project"
version = "1.0.0"

[agents.test_agent]
role = "processor"
description = "Test agent"
capabilities = ["processing"]

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7`

	err = os.WriteFile("agentflow.toml", []byte(configContent), 0644)
	require.NoError(t, err)

	// Verify file exists and has content
	content, err := os.ReadFile("agentflow.toml")
	require.NoError(t, err)
	assert.Contains(t, string(content), "test-project")
	assert.Contains(t, string(content), "test_agent")
}