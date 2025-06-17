package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kunalkushwaha/agentflow/core"
	"github.com/kunalkushwaha/agentflow/internal/mcp"
)

// MinimalMCPDemo demonstrates basic MCP functionality
func main() {
	fmt.Println("🚀 AgentFlow MCP Minimal Demo")
	fmt.Println("=============================")

	// Get logger
	logger := core.Logger()

	// 1. Initialize Cache Manager with a mock executor
	fmt.Println("📦 Initializing MCP Cache Manager...")
	cacheConfig := core.DefaultMCPCacheConfig()

	// Create a simple mock executor
	mockExecutor := &MockMCPToolExecutor{}

	cacheManager, err := mcp.NewCacheManager(cacheConfig, mockExecutor)
	if err != nil {
		log.Fatalf("Failed to create cache manager: %v", err)
	}

	// 2. Initialize Metrics
	fmt.Println("📊 Initializing MCP Metrics...")
	metricsConfig := &mcp.MetricsConfig{
		Enabled: true,
		Port:    8080,
		Path:    "/metrics",
	}
	metrics := mcp.NewMCPMetrics(metricsConfig, logger)

	// 3. Test cache operations
	fmt.Println("🧪 Testing cache operations...")
	ctx := context.Background()

	// Get cache stats
	stats, err := cacheManager.GetGlobalStats(ctx)
	if err != nil {
		log.Printf("Failed to get cache stats: %v", err)
	} else {
		fmt.Printf("   Cache stats: %d keys, %.2f hit rate\n", stats.TotalKeys, stats.HitRate)
	}

	// 4. Initialize Health Checker
	fmt.Println("🏥 Initializing Health Checker...")
	healthChecker := mcp.NewHealthChecker(nil, cacheManager, metrics, logger)

	// 5. Test health check
	healthResults := healthChecker.CheckHealth(ctx)
	for component, status := range healthResults {
		fmt.Printf("   %s: %s\n", component, status.Status)
		if status.Error != "" {
			fmt.Printf("     Error: %s\n", status.Error)
		}
	}

	fmt.Println("✅ MCP Minimal Demo completed successfully!")
}

// MockMCPToolExecutor is a simple mock implementation for demonstration
type MockMCPToolExecutor struct{}

func (m *MockMCPToolExecutor) ExecuteTool(ctx context.Context, execution core.MCPToolExecution) (core.MCPToolResult, error) {
	// Mock implementation - just return a simple result
	return core.MCPToolResult{
		ToolName:   execution.ToolName,
		ServerName: execution.ServerName,
		Success:    true,
		Content: []core.MCPContent{{
			Type: "text",
			Text: "Mock tool execution result",
		}},
	}, nil
}
