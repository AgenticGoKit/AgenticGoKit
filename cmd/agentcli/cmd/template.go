package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	templateFormat string
	templateOutput string
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage project templates",
	Long: `Manage AgenticGoKit project templates including listing, creating, and validating custom templates.

Templates can be defined in JSON or YAML format and placed in:
  - Current directory: .agenticgokit/templates/
  - User home: ~/.agenticgokit/templates/
  - System-wide: /etc/agenticgokit/templates/ (Unix) or %PROGRAMDATA%/AgenticGoKit/templates/ (Windows)

External templates can override built-in templates by using the same name.`,
}

// templateListCmd lists all available templates
var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available templates",
	Long:  "List all available project templates including built-in and custom templates.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(GetTemplateHelp())
	},
}

// templateCreateCmd creates a new template file
var templateCreateCmd = &cobra.Command{
	Use:   "create [template-name]",
	Short: "Create a new template file",
	Long: `Create a new template file with example configuration.

This command creates a template file that you can customize for your specific needs.
The template will be created in the current directory's .agenticgokit/templates/ folder.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateName := args[0]
		
		// Determine output path
		outputPath := templateOutput
		if outputPath == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			
			extension := ".yaml"
			if templateFormat == "json" {
				extension = ".json"
			}
			
			outputPath = filepath.Join(cwd, ".agenticgokit", "templates", templateName+extension)
		}
		
		// Create the example template
		if err := templateLoader.CreateTemplateExample(outputPath, templateFormat); err != nil {
			return fmt.Errorf("failed to create template: %w", err)
		}
		
		fmt.Printf("Template created successfully: %s\n", outputPath)
		fmt.Println("\nNext steps:")
		fmt.Println("1. Edit the template file to customize the configuration")
		fmt.Println("2. Test the template with: agentcli create test-project --template " + templateName)
		fmt.Println("3. Share the template by placing it in a shared template directory")
		
		return nil
	},
}

// templateValidateCmd validates a template file
var templateValidateCmd = &cobra.Command{
	Use:   "validate [template-file]",
	Short: "Validate a template file",
	Long:  "Validate that a template file has correct syntax and configuration.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateFile := args[0]
		
		// Try to load the template
		template, err := templateLoader.loadTemplateFile(templateFile)
		if err != nil {
			return fmt.Errorf("template validation failed: %w", err)
		}
		
		fmt.Printf("Template validation successful!\n")
		fmt.Printf("Name: %s\n", template.Name)
		fmt.Printf("Description: %s\n", template.Description)
		fmt.Printf("Features: %s\n", template.Features)
		fmt.Printf("Agents: %d\n", template.Config.NumAgents)
		fmt.Printf("Provider: %s\n", template.Config.Provider)
		fmt.Printf("Orchestration: %s\n", template.Config.OrchestrationMode)
		
		if template.Config.MemoryEnabled {
			fmt.Printf("Memory: %s\n", template.Config.MemoryProvider)
		}
		
		if template.Config.RAGEnabled {
			fmt.Printf("RAG: enabled (chunk size: %d)\n", template.Config.RAGChunkSize)
		}
		
		if template.Config.MCPEnabled {
			fmt.Printf("MCP: enabled (tools: %v)\n", template.Config.MCPTools)
		}
		
		return nil
	},
}

// templatePathsCmd shows template search paths
var templatePathsCmd = &cobra.Command{
	Use:   "paths",
	Short: "Show template search paths",
	Long:  "Show all directories where AgenticGoKit searches for custom templates.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Template search paths (in order of priority):")
		for i, path := range templateLoader.ListTemplatePaths() {
			fmt.Printf("%d. %s\n", i+1, path)
			
			// Check if path exists and show status
			if _, err := os.Stat(path); os.IsNotExist(err) {
				fmt.Printf("   [does not exist]\n")
			} else {
				// Count templates in this directory
				files, err := filepath.Glob(filepath.Join(path, "*.json"))
				if err == nil {
					yamlFiles, _ := filepath.Glob(filepath.Join(path, "*.yaml"))
					ymlFiles, _ := filepath.Glob(filepath.Join(path, "*.yml"))
					totalFiles := len(files) + len(yamlFiles) + len(ymlFiles)
					fmt.Printf("   [%d template(s) found]\n", totalFiles)
				}
			}
		}
		
		fmt.Println("\nTo create a custom template:")
		fmt.Println("  agentcli template create my-custom-template")
		fmt.Println("\nTo use a custom template:")
		fmt.Println("  agentcli create my-project --template my-custom-template")
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
	
	// Add subcommands
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateCreateCmd)
	templateCmd.AddCommand(templateValidateCmd)
	templateCmd.AddCommand(templatePathsCmd)
	
	// Flags for template list command
	templateListCmd.Flags().BoolP("detailed", "d", false, "Show detailed template information")
	templateListCmd.Flags().StringP("format", "f", "text", "Output format: text, json, yaml")
	
	// Flags for template create command
	templateCreateCmd.Flags().StringVarP(&templateFormat, "format", "f", "yaml", 
		"Template format (yaml, json)")
	templateCreateCmd.Flags().StringVarP(&templateOutput, "output", "o", "", 
		"Output file path (default: .agenticgokit/templates/[name].[format])")
	
	// TODO: Add template info command in the future if needed
	
	// Add completion functions
	templateCreateCmd.RegisterFlagCompletionFunc("format", completeTemplateFormats)
	templateValidateCmd.ValidArgsFunction = completeTemplateFiles
}

// completeTemplateFormats provides completion for template formats
func completeTemplateFormats(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	formats := []string{"yaml", "json"}
	return formats, cobra.ShellCompDirectiveNoFileComp
}

// completeTemplateFiles provides completion for template files
func completeTemplateFiles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Look for template files in current directory and template paths
	var files []string
	
	// Add files from current directory
	if matches, err := filepath.Glob("*.yaml"); err == nil {
		files = append(files, matches...)
	}
	if matches, err := filepath.Glob("*.yml"); err == nil {
		files = append(files, matches...)
	}
	if matches, err := filepath.Glob("*.json"); err == nil {
		files = append(files, matches...)
	}
	
	// Add files from template paths
	for _, path := range templateLoader.ListTemplatePaths() {
		if matches, err := filepath.Glob(filepath.Join(path, "*.yaml")); err == nil {
			for _, match := range matches {
				files = append(files, filepath.Base(match))
			}
		}
		if matches, err := filepath.Glob(filepath.Join(path, "*.yml")); err == nil {
			for _, match := range matches {
				files = append(files, filepath.Base(match))
			}
		}
		if matches, err := filepath.Glob(filepath.Join(path, "*.json")); err == nil {
			for _, match := range matches {
				files = append(files, filepath.Base(match))
			}
		}
	}
	
	return files, cobra.ShellCompDirectiveDefault
}