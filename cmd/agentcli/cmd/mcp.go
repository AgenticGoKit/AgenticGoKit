package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
	"github.com/spf13/cobra"
)

// mcpCmd represents the mcp command
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage and monitor MCP (Model Context Protocol) integration",
	Long: `The mcp command provides comprehensive management and monitoring capabilities 
for Model Context Protocol integration. This includes server management, tool discovery, 
connection monitoring, and performance analysis.

Key capabilities:
  * List and manage MCP server connections
  * Discover available tools from connected servers
  * Monitor server health and connection status
  * Test tool execution and performance
  * Manage server configurations
  * View detailed server and tool information

Examples:
  # List all connected MCP servers
  agentcli mcp servers

  # List all available tools
  agentcli mcp tools

  # List tools from a specific server
  agentcli mcp tools --server web-service

  # Show server health status
  agentcli mcp health

  # Test a specific tool
  agentcli mcp test --server web-service --tool web_search

  # Show detailed server information
  agentcli mcp info --server web-service`,
	Run: func(cmd *cobra.Command, args []string) {
		// Show help if no subcommand is provided
		cmd.Help()
	},
}

// MCP command flags
var (
	mcpServer  string
	mcpTool    string
	mcpFormat  string
	mcpVerbose bool
	mcpTimeout time.Duration
	mcpArgs    []string
)

// mcpServersCmd lists MCP servers
var mcpServersCmd = &cobra.Command{
	Use:   "servers",
	Short: "List MCP server connections",
	Long: `List all MCP server connections with their status, capabilities, and connection details.
Shows which servers are currently connected and available for tool execution.`, RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize MCP manager
		mcpManager, err := initializeMCPManager()
		if err != nil {
			return fmt.Errorf("failed to initialize MCP manager: %w", err)
		}

		// Get connected servers
		servers := mcpManager.ListConnectedServers()

		// Display servers based on format
		switch mcpFormat {
		case "json":
			return displayServersJSON(servers)
		case "table":
			return displayServersTable(mcpManager, servers)
		default:
			return displayServersDefault(mcpManager, servers)
		}
	},
}

// mcpToolsCmd lists available tools
var mcpToolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "List available MCP tools",
	Long: `List all available tools from connected MCP servers. Can be filtered by server
and shows tool descriptions, parameters, and usage information.`, RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize MCP manager
		mcpManager, err := initializeMCPManager()
		if err != nil {
			return fmt.Errorf("failed to initialize MCP manager: %w", err)
		}

		// Get available tools
		var tools []core.MCPToolInfo
		if mcpServer != "" {
			tools = mcpManager.GetToolsFromServer(mcpServer)
		} else {
			tools = mcpManager.GetAvailableTools()
		}

		// Display tools based on format
		switch mcpFormat {
		case "json":
			return displayToolsJSON(tools)
		case "table":
			return displayToolsTable(tools)
		default:
			return displayToolsDefault(tools)
		}
	},
}

// mcpHealthCmd shows server health
var mcpHealthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check MCP server health status",
	Long: `Check the health status of all connected MCP servers. Shows connection status,
response times, and any error conditions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Initialize MCP manager
		mcpManager, err := initializeMCPManager()
		if err != nil {
			return fmt.Errorf("failed to initialize MCP manager: %w", err)
		}

		// Perform health check
		healthStatus := mcpManager.HealthCheck(ctx)

		// Display health status
		return displayHealthStatus(healthStatus)
	},
}

// mcpTestCmd tests tool execution
var mcpTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test MCP tool execution",
	Long: `Test the execution of a specific MCP tool with provided arguments. This helps
verify that tools are working correctly and measures their performance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if mcpServer == "" || mcpTool == "" {
			return fmt.Errorf("both --server and --tool are required")
		}

		ctx, cancel := context.WithTimeout(context.Background(), mcpTimeout)
		defer cancel()

		// Initialize MCP manager
		mcpManager, err := initializeMCPManager()
		if err != nil {
			return fmt.Errorf("failed to initialize MCP manager: %w", err)
		}

		// Parse arguments
		toolArgs := parseToolArgs(mcpArgs)

		// Execute tool test
		return executeToolTest(ctx, mcpManager, mcpServer, mcpTool, toolArgs)
	},
}

// mcpInfoCmd shows detailed server information
var mcpInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show detailed MCP server information",
	Long: `Show detailed information about a specific MCP server including its capabilities,
available tools, configuration, and performance metrics.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if mcpServer == "" {
			return fmt.Errorf("--server is required")
		}

		ctx := context.Background()

		// Initialize MCP manager
		mcpManager, err := initializeMCPManager()
		if err != nil {
			return fmt.Errorf("failed to initialize MCP manager: %w", err)
		}

		// Get server information
		return showServerInfo(ctx, mcpManager, mcpServer)
	},
}

// mcpRefreshCmd refreshes tool discovery
var mcpRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh tool discovery from MCP servers",
	Long: `Refresh tool discovery by querying all connected MCP servers for their latest
tool capabilities. This updates the available tool list and capabilities.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Initialize MCP manager
		mcpManager, err := initializeMCPManager()
		if err != nil {
			return fmt.Errorf("failed to initialize MCP manager: %w", err)
		}

		// Refresh tools
		fmt.Println("Refreshing tools from MCP servers...")
		err = mcpManager.RefreshTools(ctx)
		if err != nil {
			return fmt.Errorf("failed to refresh tools: %w", err)
		}

		// Show updated tool count
		tools := mcpManager.GetAvailableTools()
		fmt.Printf("Successfully refreshed %d tools from connected servers\n", len(tools))

		return nil
	},
}

// Initialize MCP manager (placeholder - would integrate with actual factory)
func initializeMCPManager() (core.MCPManager, error) {
	// This would integrate with the actual MCP manager factory
	// For now, return an error indicating the feature needs configuration
	return nil, fmt.Errorf("MCP manager not configured - please ensure MCP integration is enabled")
}

// Display functions for servers
func displayServersJSON(servers []string) error {
	data, err := json.MarshalIndent(servers, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func displayServersTable(mcpManager core.MCPManager, servers []string) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "SERVER\tSTATUS\tTOOLS\tLAST_PING\tHEALTH")
	fmt.Fprintln(w, "------\t------\t-----\t---------\t------")

	for _, server := range servers {
		// Get server info (this would require extending the interface)
		tools := mcpManager.GetToolsFromServer(server)

		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
			server,
			"Connected", // Would come from actual status
			len(tools),
			"N/A",     // Would come from actual last ping
			"Healthy") // Would come from actual health check
	}

	return nil
}

func displayServersDefault(mcpManager core.MCPManager, servers []string) error {
	if len(servers) == 0 {
		fmt.Println("No MCP servers are currently connected.")
		return nil
	}

	fmt.Println("Connected MCP Servers")
	fmt.Println("========================")

	for i, server := range servers {
		tools := mcpManager.GetToolsFromServer(server)

		fmt.Printf("%d. %s\n", i+1, server)
		fmt.Printf("   • Tools: %d available\n", len(tools))
		fmt.Printf("   • Status: Connected\n")

		if mcpVerbose && len(tools) > 0 {
			fmt.Printf("   • Available tools: ")
			toolNames := make([]string, len(tools))
			for j, tool := range tools {
				toolNames[j] = tool.Name
			}
			fmt.Printf("%s\n", strings.Join(toolNames, ", "))
		}

		fmt.Println()
	}

	return nil
}

// Display functions for tools
func displayToolsJSON(tools []core.MCPToolInfo) error {
	data, err := json.MarshalIndent(tools, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func displayToolsTable(tools []core.MCPToolInfo) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "SERVER\tTOOL\tDESCRIPTION")
	fmt.Fprintln(w, "------\t----\t-----------")

	for _, tool := range tools {
		desc := tool.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}

		fmt.Fprintf(w, "%s\t%s\t%s\n",
			tool.ServerName,
			tool.Name,
			desc)
	}

	return nil
}

func displayToolsDefault(tools []core.MCPToolInfo) error {
	if len(tools) == 0 {
		fmt.Println("No MCP tools are currently available.")
		if mcpServer != "" {
			fmt.Printf("Server '%s' has no available tools.\n", mcpServer)
		}
		return nil
	}

	fmt.Println("Available MCP Tools")
	fmt.Println("======================")

	// Group tools by server
	serverTools := make(map[string][]core.MCPToolInfo)
	for _, tool := range tools {
		serverTools[tool.ServerName] = append(serverTools[tool.ServerName], tool)
	}

	for server, toolList := range serverTools {
		fmt.Printf("\n📡 %s (%d tools)\n", server, len(toolList))
		fmt.Println(strings.Repeat("-", len(server)+15))

		for _, tool := range toolList {
			fmt.Printf("  • %s: %s\n", tool.Name, tool.Description)

			if mcpVerbose {
				// Show tool schema if available (would need to extend the interface)
				fmt.Printf("    Schema: (tool parameter schema not yet implemented)\n")
			}
		}
	}

	return nil
}

// Health status display
func displayHealthStatus(healthStatus map[string]core.MCPHealthStatus) error {
	fmt.Println("🏥 MCP Server Health Status")
	fmt.Println("===========================")

	if len(healthStatus) == 0 {
		fmt.Println("No MCP servers to check.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "SERVER\tSTATUS\tRESPONSE_TIME\tLAST_ERROR")
	fmt.Fprintln(w, "------\t------\t-------------\t----------")
	for server, status := range healthStatus {
		errorMsg := "None"
		if status.Error != "" {
			errorMsg = status.Error
			if len(errorMsg) > 30 {
				errorMsg = errorMsg[:27] + "..."
			}
		}

		statusIcon := "🟢"
		statusText := "Healthy"
		if status.Status != "healthy" {
			statusIcon = "🔴"
			statusText = "Unhealthy"
		}

		fmt.Fprintf(w, "%s\t%s %s\t%v\t%s\n",
			server,
			statusIcon,
			statusText,
			status.ResponseTime,
			errorMsg)
	}

	return nil
}

// Tool testing
func parseToolArgs(args []string) map[string]interface{} {
	result := make(map[string]interface{})

	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}

	return result
}

func executeToolTest(ctx context.Context, mcpManager core.MCPManager, server, tool string, args map[string]interface{}) error {
	fmt.Printf("🧪 Testing tool execution: %s:%s\n", server, tool)
	fmt.Printf("Arguments: %v\n", args)
	fmt.Println("------------------------")

	// Note: This would require extending the MCPManager interface to include ExecuteTool
	// For now, show what the test would do
	fmt.Println("Tool execution test not yet implemented.")
	fmt.Println("Future implementation will:")
	fmt.Printf("   • Execute %s:%s with provided arguments\n", server, tool)
	fmt.Println("   • Measure execution time and performance")
	fmt.Println("   • Validate result format and content")
	fmt.Println("   • Report any errors or warnings")

	return nil
}

// Server information display
func showServerInfo(ctx context.Context, mcpManager core.MCPManager, serverName string) error {
	fmt.Printf("🔍 Server Information: %s\n", serverName)
	fmt.Println("=========================")

	// Check if server is connected
	servers := mcpManager.ListConnectedServers()
	connected := false
	for _, s := range servers {
		if s == serverName {
			connected = true
			break
		}
	}

	if !connected {
		return fmt.Errorf("server '%s' is not connected", serverName)
	}

	// Get server tools
	tools := mcpManager.GetToolsFromServer(serverName)

	fmt.Printf("Basic Information:\n")
	fmt.Printf("   • Status: Connected\n")
	fmt.Printf("   • Available Tools: %d\n", len(tools))

	// Show tools
	if len(tools) > 0 {
		fmt.Printf("\nAvailable Tools:\n")
		for i, tool := range tools {
			fmt.Printf("   %d. %s\n", i+1, tool.Name)
			fmt.Printf("      Description: %s\n", tool.Description)
			if mcpVerbose {
				fmt.Printf("      Schema: (parameter schema not yet implemented)\n")
			}
			fmt.Println()
		}
	}
	// Show health information
	healthStatus := mcpManager.HealthCheck(ctx)
	if health, exists := healthStatus[serverName]; exists {
		fmt.Printf("🏥 Health Status:\n")
		statusText := "Healthy"
		if health.Status != "healthy" {
			statusText = "Unhealthy"
		}
		fmt.Printf("   • Status: %s\n", statusText)
		fmt.Printf("   • Response Time: %v\n", health.ResponseTime)
		if health.Error != "" {
			fmt.Printf("   • Last Error: %s\n", health.Error)
		}
	}

	return nil
}

func init() {
	// Add mcp command to root
	rootCmd.AddCommand(mcpCmd)

	// Add subcommands
	mcpCmd.AddCommand(mcpServersCmd)
	mcpCmd.AddCommand(mcpToolsCmd)
	mcpCmd.AddCommand(mcpHealthCmd)
	mcpCmd.AddCommand(mcpTestCmd)
	mcpCmd.AddCommand(mcpInfoCmd)
	mcpCmd.AddCommand(mcpRefreshCmd)

	// Global MCP flags
	mcpCmd.PersistentFlags().StringVar(&mcpFormat, "format", "default", "Output format (default, table, json)")
	mcpCmd.PersistentFlags().BoolVar(&mcpVerbose, "verbose", false, "Show verbose output")

	// Tools command flags
	mcpToolsCmd.Flags().StringVar(&mcpServer, "server", "", "Filter tools by server name")

	// Test command flags
	mcpTestCmd.Flags().StringVar(&mcpServer, "server", "", "MCP server name (required)")
	mcpTestCmd.Flags().StringVar(&mcpTool, "tool", "", "Tool name to test (required)")
	mcpTestCmd.Flags().DurationVar(&mcpTimeout, "timeout", 30*time.Second, "Test timeout duration")
	mcpTestCmd.Flags().StringSliceVar(&mcpArgs, "arg", []string{}, "Tool arguments in key=value format")

	// Info command flags
	mcpInfoCmd.Flags().StringVar(&mcpServer, "server", "", "Server name to show info for (required)")
}

