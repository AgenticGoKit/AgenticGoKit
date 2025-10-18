package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	vnext "github.com/kunalkushwaha/agenticgokit/core/vnext"

	// Required: MCP manager plugin for connection and transport support
	_ "github.com/kunalkushwaha/agenticgokit/plugins/mcp/unified"
	// Required: MCP registry plugin for tool storage and discovery
	_ "github.com/kunalkushwaha/agenticgokit/plugins/mcp/default"

	// Import LLM provider
	_ "github.com/kunalkushwaha/agenticgokit/plugins/llm/ollama"
)

func main() {
	fmt.Println("=== vnext MCP Integration Example ===")

	// Example 1: Agent with MCP server
	fmt.Println("Example 1: Agent with Explicit MCP Server")
	if err := exampleWithMCPServer(); err != nil {
		log.Printf("Example 1 error: %v\n", err)
	}

	fmt.Println("\n" + strings.Repeat("-", 60) + "\n")

	// Example 2: Agent with MCP discovery
	fmt.Println("Example 2: Agent with MCP Discovery")
	if err := exampleWithMCPDiscovery(); err != nil {
		log.Printf("Example 2 error: %v\n", err)
	}
}

func exampleWithMCPServer() error {
	ctx := context.Background()

	// Define MCP server configuration
	// NOTE: http_sse and http_streaming transports require the unified MCP plugin
	mcpServer := vnext.MCPServer{
		Name:    "docker-http-sse",
		Type:    "http_sse",
		Address: "localhost",
		Port:    8812,
		Enabled: true,
	}

	// Create agent with MCP integration
	agent, err := vnext.NewBuilder("mcp-agent").
		WithConfig(&vnext.Config{
			Name:         "mcp-agent",
			SystemPrompt: "You are a helpful assistant with access to MCP tools. Use them to help the user.",
			Timeout:      60 * time.Second,
			LLM: vnext.LLMConfig{
				Provider:    "ollama",
				Model:       "gemma3:1b",
				Temperature: 0.7,
				MaxTokens:   200,
			},
		}).
		WithTools(
			vnext.WithMCP(mcpServer),
			vnext.WithToolTimeout(30*time.Second),
		).
		Build()

	if err != nil {
		return fmt.Errorf("failed to build agent: %w", err)
	}

	fmt.Println("✓ Agent created with MCP server")
	fmt.Printf("  Server: %s (%s:%d)\n", mcpServer.Name, mcpServer.Address, mcpServer.Port)

	// List discovered tools
	tools, err := vnext.DiscoverTools()
	if err != nil {
		fmt.Printf("  Warning: Failed to discover tools: %v\n", err)
	} else {
		fmt.Printf("  Discovered %d tools:\n", len(tools))
		for _, tool := range tools {
			fmt.Printf("    - %s: %s\n", tool.Name(), tool.Description())
		}
	}

	// Run agent with a query
	result, err := agent.Run(ctx, "What tools do you have available?")
	if err != nil {
		return fmt.Errorf("failed to run agent: %w", err)
	}

	// Display results
	fmt.Printf("\n📊 Result:\n")
	fmt.Printf("  Response: %s\n", result.Content)
	fmt.Printf("  Duration: %v\n", result.Duration)

	if len(result.ToolsCalled) > 0 {
		fmt.Printf("  Tools Used: %v\n", result.ToolsCalled)
	}

	return nil
}

func exampleWithMCPDiscovery() error {
	ctx := context.Background()

	// Create agent with MCP discovery enabled
	agent, err := vnext.NewBuilder("discovery-agent").
		WithConfig(&vnext.Config{
			Name:         "discovery-agent",
			SystemPrompt: "You are a helpful assistant with access to discovered MCP tools.",
			Timeout:      60 * time.Second,
			LLM: vnext.LLMConfig{
				Provider:    "ollama",
				Model:       "gemma3:1b",
				Temperature: 0.7,
				MaxTokens:   200,
			},
		}).
		WithTools(
			// Enable MCP discovery on common ports
			vnext.WithMCPDiscovery(8080, 8081, 8090, 8100, 8811),
			vnext.WithToolTimeout(30*time.Second),
		).
		Build()

	if err != nil {
		return fmt.Errorf("failed to build agent: %w", err)
	}

	fmt.Println("✓ Agent created with MCP discovery")
	fmt.Printf("  Discovery: Scanning ports 8080, 8081, 8090, 8100, 8811\n")

	// Run agent with a query
	result, err := agent.Run(ctx, "List available tools")
	if err != nil {
		return fmt.Errorf("failed to run agent: %w", err)
	}

	// Display results
	fmt.Printf("\n📊 Result:\n")
	fmt.Printf("  Response: %s\n", result.Content)
	fmt.Printf("  Duration: %v\n", result.Duration)

	if len(result.ToolsCalled) > 0 {
		fmt.Printf("  Tools Used: %v\n", result.ToolsCalled)
	}

	return nil
}
