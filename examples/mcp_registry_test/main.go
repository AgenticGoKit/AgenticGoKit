// Package main demonstrates MCP tool registry integration.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kunalkushwaha/agentflow/core"
	"github.com/kunalkushwaha/agentflow/internal/factory"
)

// MockMCPManager for testing tool registry integration.
type MockMCPManager struct {
	tools []core.MCPToolInfo
}

func NewMockMCPManager() *MockMCPManager {
	return &MockMCPManager{
		tools: []core.MCPToolInfo{
			{
				Name:        "web_search",
				Description: "Search the web for information",
				Schema:      map[string]interface{}{"query": "string"},
				ServerName:  "web-tools-server",
			},
			{
				Name:        "content_fetch",
				Description: "Fetch content from a URL",
				Schema:      map[string]interface{}{"url": "string"},
				ServerName:  "web-tools-server",
			},
			{
				Name:        "summarize_text",
				Description: "Summarize large text content",
				Schema:      map[string]interface{}{"text": "string", "max_words": "number"},
				ServerName:  "nlp-tools-server",
			},
		},
	}
}

func (m *MockMCPManager) Connect(ctx context.Context, serverName string) error {
	fmt.Printf("📡 Connecting to MCP server: %s\n", serverName)
	return nil
}

func (m *MockMCPManager) Disconnect(serverName string) error {
	fmt.Printf("🔌 Disconnecting from MCP server: %s\n", serverName)
	return nil
}

func (m *MockMCPManager) DisconnectAll() error {
	fmt.Println("🔌 Disconnecting from all MCP servers")
	return nil
}

func (m *MockMCPManager) DiscoverServers(ctx context.Context) ([]core.MCPServerInfo, error) {
	return []core.MCPServerInfo{
		{
			Name:        "web-tools-server",
			Type:        "tcp",
			Address:     "localhost",
			Port:        8811,
			Status:      "connected",
			Description: "Web search and content tools",
		},
		{
			Name:        "nlp-tools-server",
			Type:        "tcp",
			Address:     "localhost",
			Port:        8812,
			Status:      "connected",
			Description: "Natural language processing tools",
		},
	}, nil
}

func (m *MockMCPManager) ListConnectedServers() []string {
	return []string{"web-tools-server", "nlp-tools-server"}
}

func (m *MockMCPManager) GetServerInfo(serverName string) (*core.MCPServerInfo, error) {
	servers, _ := m.DiscoverServers(context.Background())
	for _, server := range servers {
		if server.Name == serverName {
			return &server, nil
		}
	}
	return nil, fmt.Errorf("server not found: %s", serverName)
}

func (m *MockMCPManager) RefreshTools(ctx context.Context) error {
	fmt.Println("🔄 Refreshing tools from MCP servers")
	return nil
}

func (m *MockMCPManager) GetAvailableTools() []core.MCPToolInfo {
	return m.tools
}

func (m *MockMCPManager) GetToolsFromServer(serverName string) []core.MCPToolInfo {
	var serverTools []core.MCPToolInfo
	for _, tool := range m.tools {
		if tool.ServerName == serverName {
			serverTools = append(serverTools, tool)
		}
	}
	return serverTools
}

func (m *MockMCPManager) HealthCheck(ctx context.Context) map[string]core.MCPHealthStatus {
	return map[string]core.MCPHealthStatus{
		"web-tools-server": {
			Status:       "healthy",
			LastCheck:    time.Now(),
			ResponseTime: 12 * time.Millisecond,
			ToolCount:    2,
		},
		"nlp-tools-server": {
			Status:       "healthy",
			LastCheck:    time.Now(),
			ResponseTime: 8 * time.Millisecond,
			ToolCount:    1,
		},
	}
}

func (m *MockMCPManager) GetMetrics() core.MCPMetrics {
	return core.MCPMetrics{
		ConnectedServers: 2,
		TotalTools:       3,
		ToolExecutions:   15,
		AverageLatency:   10 * time.Millisecond,
		ErrorRate:        0.05,
		ServerMetrics: map[string]core.MCPServerMetrics{
			"web-tools-server": {
				ToolCount:       2,
				Executions:      10,
				SuccessfulCalls: 9,
				FailedCalls:     1,
				AverageLatency:  12 * time.Millisecond,
				LastActivity:    time.Now(),
			},
			"nlp-tools-server": {
				ToolCount:       1,
				Executions:      5,
				SuccessfulCalls: 5,
				FailedCalls:     0,
				AverageLatency:  8 * time.Millisecond,
				LastActivity:    time.Now(),
			},
		},
	}
}

func main() {
	fmt.Println("🧪 MCP Tool Registry Integration Test")
	fmt.Println("=====================================")

	ctx := context.Background()

	// Create a default tool registry
	registry := factory.NewDefaultToolRegistry()
	fmt.Printf("📦 Created default tool registry with %d built-in tools\n", len(registry.List()))

	// Display built-in tools
	fmt.Println("\n🔧 Built-in tools:")
	for _, toolName := range registry.List() {
		fmt.Printf("   • %s\n", toolName)
	}

	// Create mock MCP manager
	mcpManager := NewMockMCPManager()
	fmt.Printf("\n🤖 Created MCP manager with %d available tools\n", len(mcpManager.GetAvailableTools()))

	// Display available MCP tools
	fmt.Println("\n🛠️  Available MCP tools:")
	for _, tool := range mcpManager.GetAvailableTools() {
		fmt.Printf("   • %s: %s (from %s)\n", tool.Name, tool.Description, tool.ServerName)
	}

	// Test MCP tool discovery and registration
	fmt.Println("\n🔍 Discovering and registering MCP tools...")
	err := factory.DiscoverAndRegisterMCPTools(ctx, registry, mcpManager)
	if err != nil {
		log.Fatalf("❌ Failed to discover and register MCP tools: %v", err)
	}

	// Display all tools after MCP registration
	allTools := registry.List()
	fmt.Printf("\n📋 All registered tools (%d total):\n", len(allTools))
	for i, toolName := range allTools {
		fmt.Printf("   %d. %s\n", i+1, toolName)
	}

	// Test tool registry validation
	fmt.Println("\n✅ Validating tool registry integration...")
	err = factory.ValidateToolRegistryIntegration(registry, mcpManager)
	if err != nil {
		log.Printf("⚠️  Validation warning: %v", err)
	} else {
		fmt.Println("✅ Tool registry integration validation passed")
	}

	// Get only MCP tools from registry
	mcpTools := factory.GetMCPToolsFromRegistry(registry, mcpManager)
	fmt.Printf("\n🎯 MCP tools in registry (%d): %v\n", len(mcpTools), mcpTools)

	// Test tool execution
	fmt.Println("\n🚀 Testing tool execution...")

	// Test a built-in tool
	if len(registry.List()) > 0 {
		builtinTool := registry.List()[0]
		fmt.Printf("\n🔧 Testing built-in tool: %s\n", builtinTool)
		result, err := registry.CallTool(ctx, builtinTool, map[string]any{"test": "value"})
		if err != nil {
			fmt.Printf("❌ Built-in tool execution failed: %v\n", err)
		} else {
			fmt.Printf("✅ Built-in tool result: %v\n", result)
		}
	}

	// Test an MCP tool
	if len(mcpTools) > 0 {
		mcpTool := mcpTools[0]
		fmt.Printf("\n🛠️  Testing MCP tool: %s\n", mcpTool)
		result, err := registry.CallTool(ctx, mcpTool, map[string]any{
			"query": "test search query",
		})
		if err != nil {
			fmt.Printf("❌ MCP tool execution failed: %v\n", err)
		} else {
			fmt.Printf("✅ MCP tool result: %v\n", result)
		}
	}

	// Display health and metrics
	fmt.Println("\n🏥 MCP Server Health:")
	health := mcpManager.HealthCheck(ctx)
	for serverName, status := range health {
		fmt.Printf("   • %s: %s (response: %v, tools: %d)\n",
			serverName, status.Status, status.ResponseTime, status.ToolCount)
	}

	fmt.Println("\n📊 MCP Metrics:")
	metrics := mcpManager.GetMetrics()
	fmt.Printf("   • Connected servers: %d\n", metrics.ConnectedServers)
	fmt.Printf("   • Total tools: %d\n", metrics.TotalTools)
	fmt.Printf("   • Tool executions: %d\n", metrics.ToolExecutions)
	fmt.Printf("   • Average latency: %v\n", metrics.AverageLatency)
	fmt.Printf("   • Error rate: %.2f%%\n", metrics.ErrorRate*100)

	fmt.Println("\n🎉 MCP Tool Registry Integration Test completed successfully!")
}
