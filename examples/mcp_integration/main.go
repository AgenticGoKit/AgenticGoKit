// Package main demonstrates MCP integration with AgentFlow.
//
// This example shows how to:
// 1. Configure MCP servers
// 2. Create and initialize an MCP manager
// 3. Connect to MCP servers
// 4. Use MCP tools in AgentFlow
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kunalkushwaha/agentflow/core"
	"github.com/kunalkushwaha/agentflow/internal/mcp"
	"github.com/kunalkushwaha/agentflow/internal/tools"
)

func main() {
	// Create a tool registry
	registry := tools.NewToolRegistry()
	// Create MCP configuration
	config := core.DefaultMCPConfig()
	// Add the MCP server port to the discovery scan
	config.ScanPorts = append(config.ScanPorts, 8811)
	// Add the running MCP server configuration
	tcpServer, err := core.NewTCPServerConfig("docker-mcp-server", "host.docker.internal", 8811)
	if err != nil {
		log.Printf("Error creating TCP server config: %v", err)
	} else {
		config.Servers = append(config.Servers, tcpServer)
	}
	// Comment out the STDIO server for now since we have a working TCP server
	// stdioServer, err := core.NewSTDIOServerConfig("local-tools", "node mcp-server.js")
	// if err != nil {
	// 	log.Printf("Error creating STDIO server config: %v", err)
	// } else {
	// 	config.Servers = append(config.Servers, stdioServer)
	// }

	// Create MCP manager
	manager, err := mcp.NewMCPManager(config, registry, nil)
	if err != nil {
		log.Fatalf("Failed to create MCP manager: %v", err)
	}

	fmt.Println("MCP Manager created successfully!")
	fmt.Printf("Configuration: %+v\n", config)

	// Demonstrate basic manager operations
	ctx := context.Background()

	// Get initial state
	fmt.Println("\n--- Initial State ---")
	servers := manager.ListConnectedServers()
	fmt.Printf("Connected servers: %v\n", servers)

	tools := manager.GetAvailableTools()
	fmt.Printf("Available MCP tools: %d\n", len(tools))

	metrics := manager.GetMetrics()
	fmt.Printf("Metrics: Connected=%d, Tools=%d, Executions=%d\n",
		metrics.ConnectedServers, metrics.TotalTools, metrics.ToolExecutions)

	// Try server discovery (if enabled)
	if config.EnableDiscovery {
		fmt.Println("\n--- Server Discovery ---")
		discovered, err := manager.DiscoverServers(ctx)
		if err != nil {
			fmt.Printf("Discovery failed: %v\n", err)
		} else {
			fmt.Printf("Discovered %d servers\n", len(discovered))
			for _, server := range discovered {
				fmt.Printf("  - %s (%s:%d)\n", server.Name, server.Address, server.Port)
			}
		}
	}

	// Try to connect to configured servers
	fmt.Println("\n--- Connecting to Servers ---")
	for _, serverConfig := range config.Servers {
		if !serverConfig.Enabled {
			continue
		}

		fmt.Printf("Attempting to connect to %s...\n", serverConfig.Name)
		err := manager.Connect(ctx, serverConfig.Name)
		if err != nil {
			fmt.Printf("  Failed: %v\n", err)
		} else {
			fmt.Printf("  Connected successfully!\n")

			// Get server info
			info, err := manager.GetServerInfo(serverConfig.Name)
			if err != nil {
				fmt.Printf("  Could not get server info: %v\n", err)
			} else {
				fmt.Printf("  Server: %s v%s\n", info.Name, info.Version)
			}
		}
	}

	// Show final state
	fmt.Println("\n--- Final State ---")
	servers = manager.ListConnectedServers()
	fmt.Printf("Connected servers: %v\n", servers)

	tools = manager.GetAvailableTools()
	fmt.Printf("Available MCP tools: %d\n", len(tools))
	for _, tool := range tools {
		fmt.Printf("  - %s (from %s): %s\n", tool.Name, tool.ServerName, tool.Description)
	}

	// Health check
	health := manager.HealthCheck(ctx)
	fmt.Printf("Health status:\n")
	for serverName, status := range health {
		fmt.Printf("  - %s: %s (tools: %d, latency: %v)\n",
			serverName, status.Status, status.ToolCount, status.ResponseTime)
		if status.Error != "" {
			fmt.Printf("    Error: %s\n", status.Error)
		}
	}

	// Test calling MCP tools
	fmt.Println("\n--- Testing MCP Tool Calls ---")
	demonstrateToolUsage(manager, registry)

	// Clean up
	fmt.Println("\n--- Cleanup ---")
	err = manager.DisconnectAll()
	if err != nil {
		fmt.Printf("Error during cleanup: %v\n", err)
	} else {
		fmt.Println("All connections closed successfully")
	}

	fmt.Println("\nMCP integration demo completed!")
}

// demonstrateToolUsage shows how to use MCP tools once connected
func demonstrateToolUsage(manager core.MCPManager, registry *tools.ToolRegistry) {
	ctx := context.Background()

	// List available tools in the registry
	toolNames := registry.List()
	fmt.Printf("Available tools in registry: %v\n", toolNames)

	// Test the search tool if available
	searchToolName := "mcp_docker-mcp-server_search"
	if tool, exists := registry.Get(searchToolName); exists {
		fmt.Printf("\n🔍 Testing search tool: %s\n", searchToolName)

		// Search for AgentFlow
		args := map[string]any{
			"query":       "AgentFlow AI framework",
			"max_results": 3,
		}

		fmt.Printf("  Searching for: %s\n", args["query"])
		result, err := tool.Call(ctx, args)
		if err != nil {
			fmt.Printf("  ❌ Error: %v\n", err)
		} else {
			fmt.Printf("  ✅ Search completed successfully!\n")
			if text, ok := result["text"].(string); ok && text != "" {
				// Truncate long results for readability
				if len(text) > 300 {
					text = text[:300] + "..."
				}
				fmt.Printf("  📄 Results preview: %s\n", text)
			}
			fmt.Printf("  📊 Full result structure: %+v\n", result)
		}
	} else {
		fmt.Printf("Search tool not found in registry\n")
	}

	// Test the fetch_content tool if available
	fetchToolName := "mcp_docker-mcp-server_fetch_content"
	if tool, exists := registry.Get(fetchToolName); exists {
		fmt.Printf("\n🌐 Testing fetch content tool: %s\n", fetchToolName)

		// Fetch content from a simple webpage
		args := map[string]any{
			"url": "https://httpbin.org/json", // Simple JSON API for testing
		}

		fmt.Printf("  Fetching URL: %s\n", args["url"])
		result, err := tool.Call(ctx, args)
		if err != nil {
			fmt.Printf("  ❌ Error: %v\n", err)
		} else {
			fmt.Printf("  ✅ Content fetched successfully!\n")
			if text, ok := result["text"].(string); ok && text != "" {
				// Truncate long results for readability
				if len(text) > 200 {
					text = text[:200] + "..."
				}
				fmt.Printf("  📄 Content preview: %s\n", text)
			}
			fmt.Printf("  📊 Full result structure: %+v\n", result)
		}
	} else {
		fmt.Printf("Fetch content tool not found in registry\n")
	}

	// Show a summary
	mcpToolCount := 0
	for _, toolName := range toolNames {
		if len(toolName) > 4 && toolName[:4] == "mcp_" {
			mcpToolCount++
		}
	}
	fmt.Printf("\n📈 Summary: Successfully tested MCP tools (%d MCP tools available)\n", mcpToolCount)
}
