package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "agentcli",
	Short: "Command-line interface for inspecting and managing AgentFlow executions",
	Long: `AgentFlow CLI provides comprehensive tools for inspecting, visualizing, and managing 
agent executions and interactions within the AgentFlow framework.

Key capabilities:
  * Trace visualization - View execution traces with detailed agent flows
  * Sequence diagrams - See agent interactions in chronological order
  * Condensed routes - View simplified linear agent execution paths
  * Per-event analysis - Examine individual event flows and requeue patterns
  * MCP management - Manage Model Context Protocol servers and tools
  * Cache management - Monitor and optimize MCP tool result caches

Examples:
  # View a basic trace with all details
  agentcli trace <session-id>

  # View only the agent flow for a session
  agentcli trace --flow-only <session-id>

  # Filter trace to see only a specific agent's activity
  agentcli trace --filter <agent-name> <session-id>

  # List connected MCP servers
  agentcli mcp servers

  # View cache statistics
  agentcli cache stats

  # Clear specific caches
  agentcli cache clear --server web-service`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.agentcli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
