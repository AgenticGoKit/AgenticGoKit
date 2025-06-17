// Package main demonstrates MCP-aware agent functionality in AgentFlow.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kunalkushwaha/agentflow/core"
)

// ExampleLLMProvider is a simple mock LLM provider for demonstration.
type ExampleLLMProvider struct{}

func (p *ExampleLLMProvider) Call(ctx context.Context, prompt core.Prompt) (core.Response, error) {
	// Simulate LLM tool selection based on the prompt
	if prompt.User == "find information about golang" {
		return core.Response{
			Content: `["search", "fetch_content"]`,
			Usage:   core.UsageStats{PromptTokens: 50, CompletionTokens: 10, TotalTokens: 60},
		}, nil
	}

	return core.Response{
		Content: `["search"]`,
		Usage:   core.UsageStats{PromptTokens: 30, CompletionTokens: 5, TotalTokens: 35},
	}, nil
}

func (p *ExampleLLMProvider) Stream(ctx context.Context, prompt core.Prompt) (<-chan core.Token, error) {
	// Not implemented for this example
	return nil, fmt.Errorf("streaming not implemented in example")
}

func (p *ExampleLLMProvider) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	// Not implemented for this example
	return nil, fmt.Errorf("embeddings not implemented in example")
}

// ExampleMCPManager is a mock MCP manager for demonstration.
type ExampleMCPManager struct {
	tools []core.MCPToolInfo
}

func NewExampleMCPManager() *ExampleMCPManager {
	return &ExampleMCPManager{
		tools: []core.MCPToolInfo{
			{
				Name:        "search",
				Description: "Search for information on the web",
				Schema:      map[string]interface{}{"query": "string"},
				ServerName:  "web-tools",
			},
			{
				Name:        "fetch_content",
				Description: "Fetch content from a URL",
				Schema:      map[string]interface{}{"url": "string"},
				ServerName:  "web-tools",
			},
		},
	}
}

func (m *ExampleMCPManager) Connect(ctx context.Context, serverName string) error {
	fmt.Printf("🔗 Connected to MCP server: %s\n", serverName)
	return nil
}

func (m *ExampleMCPManager) Disconnect(serverName string) error {
	fmt.Printf("🔌 Disconnected from MCP server: %s\n", serverName)
	return nil
}

func (m *ExampleMCPManager) DisconnectAll() error {
	fmt.Println("🔌 Disconnected from all MCP servers")
	return nil
}

func (m *ExampleMCPManager) DiscoverServers(ctx context.Context) ([]core.MCPServerInfo, error) {
	return []core.MCPServerInfo{
		{
			Name:        "web-tools",
			Type:        "tcp",
			Address:     "localhost",
			Port:        8811,
			Status:      "connected",
			Description: "Web search and content fetching tools",
		},
	}, nil
}

func (m *ExampleMCPManager) ListConnectedServers() []string {
	return []string{"web-tools"}
}

func (m *ExampleMCPManager) GetServerInfo(serverName string) (*core.MCPServerInfo, error) {
	if serverName == "web-tools" {
		return &core.MCPServerInfo{
			Name:        "web-tools",
			Type:        "tcp",
			Address:     "localhost",
			Port:        8811,
			Status:      "connected",
			Description: "Web search and content fetching tools",
		}, nil
	}
	return nil, fmt.Errorf("server not found: %s", serverName)
}

func (m *ExampleMCPManager) RefreshTools(ctx context.Context) error {
	fmt.Println("🔄 Refreshed tools from MCP servers")
	return nil
}

func (m *ExampleMCPManager) GetAvailableTools() []core.MCPToolInfo {
	return m.tools
}

func (m *ExampleMCPManager) GetToolsFromServer(serverName string) []core.MCPToolInfo {
	if serverName == "web-tools" {
		return m.tools
	}
	return []core.MCPToolInfo{}
}

func (m *ExampleMCPManager) HealthCheck(ctx context.Context) map[string]core.MCPHealthStatus {
	return map[string]core.MCPHealthStatus{
		"web-tools": {
			Status:       "healthy",
			LastCheck:    time.Now(),
			ResponseTime: 15 * time.Millisecond,
			ToolCount:    len(m.tools),
		},
	}
}

func (m *ExampleMCPManager) GetMetrics() core.MCPMetrics {
	return core.MCPMetrics{
		ConnectedServers: 1,
		TotalTools:       len(m.tools),
		ToolExecutions:   5,
		AverageLatency:   15 * time.Millisecond,
		ErrorRate:        0.02,
		ServerMetrics: map[string]core.MCPServerMetrics{
			"web-tools": {
				ToolCount:       len(m.tools),
				Executions:      5,
				SuccessfulCalls: 5,
				FailedCalls:     0,
				AverageLatency:  15 * time.Millisecond,
				LastActivity:    time.Now(),
			},
		},
	}
}

func main() {
	fmt.Println("🚀 AgentFlow MCP-Aware Agent Demo")
	fmt.Println("==================================")

	// Create dependencies
	llmProvider := &ExampleLLMProvider{}
	mcpManager := NewExampleMCPManager()

	// Create MCP agent configuration
	config := core.DefaultMCPAgentConfig()
	config.MaxToolsPerExecution = 3
	config.ToolSelectionTimeout = 15 * time.Second

	// Create MCP-aware agent
	agent := core.NewMCPAwareAgent("demo-agent", llmProvider, mcpManager, config)

	fmt.Printf("✅ Created MCP-aware agent: %s\n", agent.Name())
	fmt.Printf("📊 Available tools: %d\n", len(agent.GetAvailableMCPTools()))

	// Display available tools
	fmt.Println("\n🛠️  Available MCP Tools:")
	for _, tool := range agent.GetAvailableMCPTools() {
		fmt.Printf("   • %s: %s (from %s)\n", tool.Name, tool.Description, tool.ServerName)
	}

	// Create a test state with a query
	state := core.NewState()
	state.Set("query", "find information about golang")
	state.Set("user_id", "demo-user")

	fmt.Println("\n🧠 Running MCP-aware agent...")
	ctx := context.Background()

	// Run the agent
	result, err := agent.Run(ctx, state)
	if err != nil {
		log.Fatalf("❌ Agent execution failed: %v", err)
	}

	fmt.Println("\n📋 Agent Results:")
	fmt.Println("=================")

	// Display results
	if results, exists := result.Get("mcp_results"); exists {
		if resultSlice, ok := results.([]map[string]interface{}); ok {
			for i, res := range resultSlice {
				fmt.Printf("%d. Tool: %s\n", i+1, res["tool_name"])
				fmt.Printf("   Success: %v\n", res["success"])
				if content, exists := res["content"]; exists {
					fmt.Printf("   Content: %s\n", content)
				}
				if err, exists := res["error"]; exists && err != "" {
					fmt.Printf("   Error: %s\n", err)
				}
				fmt.Println()
			}
		}
	}

	// Display health check
	fmt.Println("🏥 MCP Server Health:")
	health := mcpManager.HealthCheck(ctx)
	for serverName, status := range health {
		fmt.Printf("   • %s: %s (response: %v, tools: %d)\n",
			serverName, status.Status, status.ResponseTime, status.ToolCount)
	}

	// Display metrics
	fmt.Println("\n📈 MCP Metrics:")
	metrics := mcpManager.GetMetrics()
	fmt.Printf("   • Connected servers: %d\n", metrics.ConnectedServers)
	fmt.Printf("   • Total tools: %d\n", metrics.TotalTools)
	fmt.Printf("   • Tool executions: %d\n", metrics.ToolExecutions)
	fmt.Printf("   • Average latency: %v\n", metrics.AverageLatency)
	fmt.Printf("   • Error rate: %.2f%%\n", metrics.ErrorRate*100)

	fmt.Println("\n✨ Demo completed successfully!")
}
