package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kunalkushwaha/agentflow/core"
	"github.com/kunalkushwaha/agentflow/internal/mcp"
)

// MCPAgentConnectorDemo demonstrates CLI to MCP agent connection logic
func main() {
	fmt.Println("🔗 AgentFlow MCP Agent Connector Demo")
	fmt.Println("=====================================")

	// Initialize MCP system components
	mcpSystem, err := initializeMCPSystem()
	if err != nil {
		log.Fatalf("Failed to initialize MCP system: %v", err)
	}

	// Simulate CLI commands connecting to MCP agents
	demonstrateAgentConnections(mcpSystem)

	fmt.Println("✅ MCP Agent Connector Demo completed!")
}

type MCPSystemDemo struct {
	cacheManager  *mcp.CacheManager
	metrics       *mcp.MCPMetrics
	healthChecker *mcp.HealthChecker
}

func initializeMCPSystem() (*MCPSystemDemo, error) {
	logger := core.Logger()

	// 1. Initialize cache with mock executor
	cacheConfig := core.DefaultMCPCacheConfig()
	mockExecutor := &MockMCPToolExecutor{}
	cacheManager, err := mcp.NewCacheManager(cacheConfig, mockExecutor)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache manager: %w", err)
	}

	// 2. Initialize metrics
	metricsConfig := &mcp.MetricsConfig{
		Enabled: true,
		Port:    8081, // Different port to avoid conflicts
		Path:    "/metrics",
	}
	metrics := mcp.NewMCPMetrics(metricsConfig, logger)

	// 3. Initialize health checker
	healthChecker := mcp.NewHealthChecker(nil, cacheManager, metrics, logger)

	return &MCPSystemDemo{
		cacheManager:  cacheManager,
		metrics:       metrics,
		healthChecker: healthChecker,
	}, nil
}

func demonstrateAgentConnections(system *MCPSystemDemo) {
	fmt.Println("\n🤖 Simulating CLI-to-Agent Connections...")

	// Simulate different CLI commands interacting with MCP agents
	cliCommands := []struct {
		command     string
		description string
		action      func() error
	}{
		{
			command:     "agentcli mcp connect --server demo-server",
			description: "Connect to MCP server",
			action: func() error {
				return simulateServerConnection("demo-server")
			},
		},
		{
			command:     "agentcli mcp list-tools --server demo-server",
			description: "List available tools",
			action: func() error {
				return simulateToolListing("demo-server")
			},
		},
		{
			command:     "agentcli mcp execute --tool summarize --args '{\"text\":\"demo\"}'",
			description: "Execute MCP tool",
			action: func() error {
				return simulateToolExecution(system, "summarize")
			},
		},
		{
			command:     "agentcli mcp health",
			description: "Check system health",
			action: func() error {
				return simulateHealthCheck(system)
			},
		},
		{
			command:     "agentcli mcp metrics",
			description: "View system metrics",
			action: func() error {
				return simulateMetricsView(system)
			},
		},
	}

	// Execute each CLI command simulation
	for i, cmd := range cliCommands {
		fmt.Printf("\n%d. %s\n", i+1, cmd.description)
		fmt.Printf("   Command: %s\n", cmd.command)

		err := cmd.action()
		if err != nil {
			fmt.Printf("   Result: ❌ Error - %v\n", err)
		} else {
			fmt.Printf("   Result: ✅ Success\n")
		}
	}
}

func simulateServerConnection(serverName string) error {
	fmt.Printf("   → Connecting to MCP server '%s'...\n", serverName)

	// Simulate connection logic
	time.Sleep(100 * time.Millisecond) // Simulate connection time

	fmt.Printf("   → Server '%s' connection established\n", serverName)
	return nil
}

func simulateToolListing(serverName string) error {
	fmt.Printf("   → Listing tools from server '%s'...\n", serverName)

	// Simulate tool discovery
	mockTools := []string{"summarize", "translate", "search", "analyze"}

	fmt.Printf("   → Available tools:\n")
	for _, tool := range mockTools {
		fmt.Printf("     - %s\n", tool)
	}

	return nil
}

func simulateToolExecution(system *MCPSystemDemo, toolName string) error {
	fmt.Printf("   → Executing tool '%s'...\n", toolName)

	ctx := context.Background()

	// Create tool execution request
	execution := core.MCPToolExecution{
		ToolName:   toolName,
		ServerName: "demo-server",
		Arguments:  map[string]interface{}{"text": "demo content"},
	}

	// Simulate tool execution by calling the mock executor directly
	mockExecutor := &MockMCPToolExecutor{}
	result, err := mockExecutor.ExecuteTool(ctx, execution)
	if err != nil {
		return fmt.Errorf("tool execution failed: %w", err)
	}

	fmt.Printf("   → Tool executed successfully\n")
	if len(result.Content) > 0 {
		fmt.Printf("   → Result: %s\n", result.Content[0].Text)
	}

	// Demonstrate cache interaction
	cache := system.cacheManager.GetCache(toolName, "demo-server")
	cacheKey := core.GenerateCacheKey(toolName, "demo-server", map[string]string{"text": "demo content"})

	// Store result in cache
	err = cache.Set(ctx, cacheKey, result, 1*time.Hour)
	if err != nil {
		fmt.Printf("   → Warning: Failed to cache result: %v\n", err)
	} else {
		fmt.Printf("   → Result cached successfully\n")
	}

	return nil
}

func simulateHealthCheck(system *MCPSystemDemo) error {
	fmt.Printf("   → Checking system health...\n")

	ctx := context.Background()
	healthResults := system.healthChecker.CheckHealth(ctx)

	fmt.Printf("   → Health Status:\n")
	for component, status := range healthResults {
		statusIcon := "✅"
		if status.Status != "healthy" {
			statusIcon = "❌"
		}
		fmt.Printf("     %s %s: %s\n", statusIcon, component, status.Status)
		if status.Error != "" {
			fmt.Printf("       Error: %s\n", status.Error)
		}
	}

	return nil
}

func simulateMetricsView(system *MCPSystemDemo) error {
	fmt.Printf("   → Gathering system metrics...\n")

	// Get cache stats
	ctx := context.Background()
	stats, err := system.cacheManager.GetGlobalStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	fmt.Printf("   → Cache Metrics:\n")
	fmt.Printf("     - Total Keys: %d\n", stats.TotalKeys)
	fmt.Printf("     - Hit Rate: %.2f%%\n", stats.HitRate*100)
	fmt.Printf("     - Hit Count: %d\n", stats.HitCount)
	fmt.Printf("     - Miss Count: %d\n", stats.MissCount)

	return nil
}

// MockMCPToolExecutor - reuse from minimal example
type MockMCPToolExecutor struct{}

func (m *MockMCPToolExecutor) ExecuteTool(ctx context.Context, execution core.MCPToolExecution) (core.MCPToolResult, error) {
	// Simulate processing time
	time.Sleep(50 * time.Millisecond)

	return core.MCPToolResult{
		ToolName:   execution.ToolName,
		ServerName: execution.ServerName,
		Success:    true,
		Content: []core.MCPContent{{
			Type: "text",
			Text: fmt.Sprintf("Mock result from %s tool with args %v", execution.ToolName, execution.Arguments),
		}},
		Duration: 50 * time.Millisecond,
	}, nil
}
