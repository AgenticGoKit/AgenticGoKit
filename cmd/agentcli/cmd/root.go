package cmd

import (
	"fmt"
	"os"

	"github.com/kunalkushwaha/agenticgokit/cmd/agentcli/version"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "agentcli",
	Short: "AgenticGoKit CLI - Build and manage multi-agent AI systems",
	Long: `AgenticGoKit CLI provides comprehensive tools for creating, managing, and debugging 
multi-agent AI systems built with the AgenticGoKit framework.

PROJECT MANAGEMENT
  create      Create new AgenticGoKit projects with scaffolding
  validate    Validate project structure and configuration  
  info        Show current project information and status
  update      Update project templates and dependencies

DEVELOPMENT & DEBUG  
  trace       View execution traces and agent interactions
  logs        View and filter application logs
  memory      Debug memory system and RAG functionality
  config      Inspect and manage configuration
  health      Check system health and connectivity
  status      Show current system status

MCP & TOOLS
  mcp         Manage Model Context Protocol servers and tools
  cache       Monitor and optimize MCP tool result caches

UTILITIES
  list        List available sessions and projects
  completion  Generate shell completion scripts
  version     Show version information

GETTING STARTED:
  # Create your first project
  agentcli create my-project

  # Create with memory and RAG
  agentcli create my-rag-system --memory-enabled --rag-enabled

  # Create with MCP tools
  agentcli create my-tools --mcp-enabled

  # Check project health
  agentcli health

  # View project information  
  agentcli info

For detailed help on any command, use: agentcli <command> --help`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// CommandCategory represents a group of related commands
type CommandCategory struct {
	Name        string
	Description string
	Icon        string
}

// Command categories for better organization
var commandCategories = map[string]CommandCategory{
	"project": {
		Name:        "Project Management",
		Description: "Commands for creating and managing AgenticGoKit projects",
	},
	"debug": {
		Name:        "Development & Debug",
		Description: "Commands for debugging and development",
	},
	"mcp": {
		Name:        "MCP & Tools",
		Description: "Model Context Protocol and tool management",
	},
	"utility": {
		Name:        "Utilities",
		Description: "General utility commands",
	},
}

// getCommandCategory returns the category for a command
func getCommandCategory(cmdName string) string {
	switch cmdName {
	case "create", "validate", "info", "update":
		return "project"
	case "trace", "logs", "memory", "config", "health", "status":
		return "debug"
	case "mcp", "cache":
		return "mcp"
	case "list", "completion", "version":
		return "utility"
	default:
		return "utility"
	}
}

// customHelpTemplate provides a better organized help output
const customHelpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

func init() {
	// Set custom help template
	rootCmd.SetHelpTemplate(customHelpTemplate)
	
	// Add --version flag to root command
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
	
	// Handle --version flag and default behavior
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
			fmt.Println(version.GetVersionString())
			return
		}
		
		// Show enhanced help when no command is provided
		showEnhancedHelp(cmd)
	}
}

// showEnhancedHelp displays categorized help information
func showEnhancedHelp(cmd *cobra.Command) {
	fmt.Print(cmd.Long)
	fmt.Println()
	
	// Group commands by category
	categoryCommands := make(map[string][]*cobra.Command)
	for _, subCmd := range cmd.Commands() {
		if !subCmd.Hidden {
			category := getCommandCategory(subCmd.Name())
			categoryCommands[category] = append(categoryCommands[category], subCmd)
		}
	}
	
	// Display commands by category
	categoryOrder := []string{"project", "debug", "mcp", "utility"}
	for _, categoryKey := range categoryOrder {
		if commands, exists := categoryCommands[categoryKey]; exists && len(commands) > 0 {
			category := commandCategories[categoryKey]
			fmt.Printf("\n%s %s:\n", category.Icon, category.Name)
			
			for _, subCmd := range commands {
				fmt.Printf("  %-12s %s\n", subCmd.Name(), subCmd.Short)
			}
		}
	}
	
	fmt.Printf("\nFlags:\n")
	fmt.Printf("  -h, --help      Show help information\n")
	fmt.Printf("  -v, --version   Show version information\n")
	
	fmt.Printf("\nUse \"agentcli <command> --help\" for detailed information about a command.\n")
}
