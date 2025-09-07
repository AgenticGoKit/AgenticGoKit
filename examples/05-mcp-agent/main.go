package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/mcp/unified" // Import unified transport plugin
)

func main() {
	fmt.Println("🚀 Starting MCP Agent Example with v0.0.2...")
	fmt.Println("This example demonstrates AgenticGoKit MCP integration with the updated mcp-navigator-go v0.0.2")

	ctx := context.Background()

	// Create MCP configuration with various server types
	mcpConfig := core.MCPConfig{
		EnableDiscovery:   false, // Use explicit server configuration
		ConnectionTimeout: 30 * time.Second,
		MaxRetries:        3,
		RetryDelay:        1 * time.Second,
		EnableCaching:     true,
		CacheTimeout:      5 * time.Minute,
		MaxConnections:    10,
		Servers: []core.MCPServerConfig{
			{
				Name:    "docker-mcp-tcp",
				Type:    "tcp",
				Host:    "localhost",
				Port:    8811,
				Enabled: true,
			},
			{
				Name:     "docker-mcp-http-sse",
				Type:     "http_sse",
				Endpoint: "http://localhost:8812",
				Enabled:  true,
			},
			{
				Name:     "docker-mcp-http-streaming",
				Type:     "http_streaming",
				Endpoint: "http://localhost:8813",
				Enabled:  true,
			},
		},
	}

	// Initialize MCP
	fmt.Println("📡 Initializing MCP...")
	if err := core.InitializeMCP(mcpConfig); err != nil {
		log.Fatalf("Failed to initialize MCP: %v", err)
	}

	// Get the MCP manager
	manager := core.GetMCPManager()
	if manager == nil {
		log.Fatal("MCP manager not available")
	}

	// Test connections to all configured servers
	fmt.Println("\n🔌 Testing connections to MCP servers...")

	for _, server := range mcpConfig.Servers {
		fmt.Printf("Connecting to %s (%s)...\n", server.Name, server.Type)

		if err := manager.Connect(ctx, server.Name); err != nil {
			fmt.Printf("❌ Failed to connect to %s: %v\n", server.Name, err)
			continue
		}

		fmt.Printf("✅ Connected to %s successfully\n", server.Name)
	}

	// List connected servers
	fmt.Println("\n📋 Connected servers:")
	connectedServers := manager.ListConnectedServers()
	for _, serverName := range connectedServers {
		fmt.Printf("  • %s\n", serverName)

		// Get server info
		if info, err := manager.GetServerInfo(serverName); err == nil {
			fmt.Printf("    Type: %s, Status: %s\n", info.Type, info.Status)
		}
	}

	// Refresh and discover tools
	fmt.Println("\n🔍 Discovering tools from connected servers...")
	if err := manager.RefreshTools(ctx); err != nil {
		fmt.Printf("Warning: Some tools failed to refresh: %v\n", err)
	}

	// Get available tools
	tools := manager.GetAvailableTools()
	fmt.Printf("Found %d total tools:\n", len(tools))

	for _, tool := range tools {
		fmt.Printf("  🛠️  %s (from %s)\n", tool.Name, tool.ServerName)
		fmt.Printf("      Description: %s\n", tool.Description)
	}

	// Test tool execution if tools are available
	if len(tools) > 0 {
		fmt.Println("\n🧪 Testing tool execution...")

		// Try to execute a search tool if available
		for _, tool := range tools {
			if tool.Name == "search" {
				fmt.Printf("Executing search tool...\n")

				args := map[string]interface{}{
					"query":       "AgenticGoKit MCP integration",
					"max_results": 3,
				}

				if result, err := core.ExecuteMCPTool(ctx, tool.Name, args); err != nil {
					fmt.Printf("❌ Tool execution failed: %v\n", err)
				} else {
					fmt.Printf("✅ Tool executed successfully!\n")
					fmt.Printf("   Duration: %v\n", result.Duration)
					fmt.Printf("   Success: %v\n", result.Success)
					if len(result.Content) > 0 {
						fmt.Printf("   Content: %s\n", result.Content[0].Text)
					}
				}
				break
			}
		}
	}

	// Perform health check
	fmt.Println("\n🏥 Performing health check...")
	healthStatus := manager.HealthCheck(ctx)

	for serverName, status := range healthStatus {
		fmt.Printf("  %s: %s", serverName, status.Status)
		if status.Error != "" {
			fmt.Printf(" (Error: %s)", status.Error)
		}
		if status.ResponseTime > 0 {
			fmt.Printf(" (Response time: %v)", status.ResponseTime)
		}
		fmt.Printf(" (Tools: %d)\n", status.ToolCount)
	}

	// Get metrics
	fmt.Println("\n📊 MCP Metrics:")
	metrics := manager.GetMetrics()
	fmt.Printf("  Connected servers: %d\n", metrics.ConnectedServers)
	fmt.Printf("  Total tools: %d\n", metrics.TotalTools)
	fmt.Printf("  Tool executions: %d\n", metrics.ToolExecutions)
	fmt.Printf("  Average latency: %v\n", metrics.AverageLatency)
	fmt.Printf("  Error rate: %.2f%%\n", metrics.ErrorRate*100)

	// Clean shutdown
	fmt.Println("\n🔄 Shutting down...")
	if err := core.ShutdownMCP(); err != nil {
		fmt.Printf("Warning: Shutdown error: %v\n", err)
	}

	fmt.Println("✅ MCP Agent example completed successfully!")
	fmt.Println("\n📖 Key features demonstrated:")
	fmt.Println("  • mcp-navigator-go v0.0.2 ClientBuilder pattern")
	fmt.Println("  • TCP transport connectivity")
	fmt.Println("  • HTTP transport support (with fallback)")
	fmt.Println("  • Tool discovery and execution")
	fmt.Println("  • Health monitoring and metrics")
	fmt.Println("  • Graceful connection management")
}
