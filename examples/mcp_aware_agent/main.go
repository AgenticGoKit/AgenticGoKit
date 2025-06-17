// Example demonstrating the MCP-aware agent functionality
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kunalkushwaha/agentflow/core"
)

func main() {
	fmt.Println("MCP-Aware Agent Example")

	// Create a context
	ctx := context.Background()
	// Create an LLM provider (mock for demonstration)
	llmProvider := &core.MockModelProvider{}

	// Configure MCP settings
	mcpConfig := core.MCPConfig{
		EnableDiscovery:   true,
		ConnectionTimeout: 30000000000, // 30 seconds in nanoseconds
		MaxRetries:        3,
		Servers: []core.MCPServerConfig{
			{
				Name:    "web-tools",
				Type:    "tcp",
				Host:    "host.docker.internal",
				Port:    8811,
				Enabled: true,
			},
		},
	}

	// Configure agent settings
	agentConfig := core.DefaultMCPAgentConfig()

	// Initialize MCP infrastructure
	fmt.Println("Initializing MCP manager...")
	if err := core.InitializeMCPManager(mcpConfig); err != nil {
		log.Printf("Warning: Failed to initialize MCP manager: %v", err)
		log.Println("Continuing with example without MCP tools...")
		return
	}

	// Initialize MCP tool registry
	fmt.Println("Initializing MCP tool registry...")
	if err := core.InitializeMCPToolRegistry(); err != nil {
		log.Printf("Warning: Failed to initialize MCP tool registry: %v", err)
		return
	}

	// Create MCP-aware agent using factory functions
	fmt.Println("Creating MCP-aware agent...")
	agent, err := core.CreateMCPAgentWithLLMAndTools(ctx, "demo-agent", llmProvider, mcpConfig, agentConfig)
	if err != nil {
		log.Printf("Warning: Failed to create MCP agent: %v", err)
		return
	}

	// Create a simple state with a query
	inputState := core.NewState()
	inputState.Set("query", "search for information about AI agents")
	inputState.Set("url", "https://example.com/ai-agents")

	// Run the agent
	fmt.Println("Running MCP-aware agent...")
	outputState, err := agent.Run(ctx, inputState)
	if err != nil {
		log.Printf("Agent execution error: %v", err)
		return
	}

	// Display results
	fmt.Println("\nAgent execution completed!")
	fmt.Printf("Available MCP tools: %d\n", len(agent.GetAvailableMCPTools()))

	// Print available tools
	tools := agent.GetAvailableMCPTools()
	if len(tools) > 0 {
		fmt.Println("\nAvailable MCP tools:")
		for _, tool := range tools {
			fmt.Printf("- %s: %s (server: %s)\n", tool.Name, tool.Description, tool.ServerName)
		}
	}

	// Check for MCP results in the output state
	if results, exists := outputState.Get("mcp_results"); exists {
		fmt.Printf("\nMCP Results: %v\n", results)
	}

	// Check for specific tool results
	keys := outputState.Keys()
	fmt.Printf("\nOutput state keys: %v\n", keys)

	// Clean up
	fmt.Println("\nShutting down MCP manager...")
	if err := core.ShutdownMCPManager(); err != nil {
		log.Printf("Warning: Error during shutdown: %v", err)
	}

	fmt.Println("Example completed successfully!")
}
