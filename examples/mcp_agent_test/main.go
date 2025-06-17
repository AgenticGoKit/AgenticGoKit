// Simple test of the MCP-aware agent interface and functionality
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kunalkushwaha/agentflow/core"
)

// SimpleMockLLM is a basic mock LLM for testing
type SimpleMockLLM struct{}

func (m *SimpleMockLLM) Call(ctx context.Context, prompt core.Prompt) (core.Response, error) {
	// Return a simple mock response for tool selection
	response := core.Response{
		Content: `["search", "fetch_content"]`,
		Usage: core.UsageStats{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
		FinishReason: "stop",
	}
	return response, nil
}

func (m *SimpleMockLLM) Stream(ctx context.Context, prompt core.Prompt) (<-chan core.Token, error) {
	// Simple stream implementation
	ch := make(chan core.Token, 1)
	go func() {
		defer close(ch)
		ch <- core.Token{Content: `["search", "fetch_content"]`}
	}()
	return ch, nil
}

func (m *SimpleMockLLM) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	// Return mock embeddings
	embeddings := make([][]float64, len(texts))
	for i := range embeddings {
		embeddings[i] = []float64{0.1, 0.2, 0.3}
	}
	return embeddings, nil
}

// SimpleMockMCPManager is a basic mock MCP manager for testing
type SimpleMockMCPManager struct{}

func (m *SimpleMockMCPManager) Connect(ctx context.Context, serverName string) error {
	return nil
}

func (m *SimpleMockMCPManager) Disconnect(serverName string) error {
	return nil
}

func (m *SimpleMockMCPManager) DisconnectAll() error {
	return nil
}

func (m *SimpleMockMCPManager) DiscoverServers(ctx context.Context) ([]core.MCPServerInfo, error) {
	return []core.MCPServerInfo{}, nil
}

func (m *SimpleMockMCPManager) ListConnectedServers() []string {
	return []string{"mock-server"}
}

func (m *SimpleMockMCPManager) GetServerInfo(serverName string) (*core.MCPServerInfo, error) {
	return &core.MCPServerInfo{
		Name:        "mock-server",
		Type:        "tcp",
		Address:     "localhost",
		Port:        8811,
		Status:      "connected",
		Description: "Mock MCP server for testing",
	}, nil
}

func (m *SimpleMockMCPManager) RefreshTools(ctx context.Context) error {
	return nil
}

func (m *SimpleMockMCPManager) GetAvailableTools() []core.MCPToolInfo {
	return []core.MCPToolInfo{
		{
			Name:        "search",
			Description: "Search for information on the web",
			ServerName:  "mock-server",
			Schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
				},
			},
		},
		{
			Name:        "fetch_content",
			Description: "Fetch content from a URL",
			ServerName:  "mock-server",
			Schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{
						"type":        "string",
						"description": "URL to fetch",
					},
				},
			},
		},
	}
}

func (m *SimpleMockMCPManager) GetToolsFromServer(serverName string) []core.MCPToolInfo {
	return m.GetAvailableTools()
}

func (m *SimpleMockMCPManager) HealthCheck(ctx context.Context) map[string]core.MCPHealthStatus {
	return map[string]core.MCPHealthStatus{
		"mock-server": {
			Status:       "healthy",
			LastCheck:    time.Now(),
			ResponseTime: 10 * time.Millisecond,
			ToolCount:    2,
		},
	}
}

func (m *SimpleMockMCPManager) GetMetrics() core.MCPMetrics {
	return core.MCPMetrics{
		ConnectedServers: 1,
		TotalTools:       2,
		ToolExecutions:   0,
		AverageLatency:   10 * time.Millisecond,
		ErrorRate:        0.0,
		ServerMetrics: map[string]core.MCPServerMetrics{
			"mock-server": {
				ToolCount:       2,
				Executions:      0,
				SuccessfulCalls: 0,
				FailedCalls:     0,
				AverageLatency:  10 * time.Millisecond,
				LastActivity:    time.Now(),
			},
		},
	}
}

func main() {
	fmt.Println("MCP-Aware Agent Interface Test")

	// Create mock dependencies
	llmProvider := &SimpleMockLLM{}
	mcpManager := &SimpleMockMCPManager{}

	// Create agent configuration
	config := core.DefaultMCPAgentConfig()

	// Create MCP-aware agent
	agent := core.NewMCPAwareAgent("test-agent", llmProvider, mcpManager, config)

	fmt.Printf("Agent created: %s\n", agent.Name())

	// Test getting available tools
	tools := agent.GetAvailableMCPTools()
	fmt.Printf("Available MCP tools: %d\n", len(tools))
	for _, tool := range tools {
		fmt.Printf("- %s: %s\n", tool.Name, tool.Description)
	}

	// Create test state
	ctx := context.Background()
	inputState := core.NewState()
	inputState.Set("query", "search for AI agents")
	inputState.Set("url", "https://example.com")

	// Test tool selection
	fmt.Println("\nTesting tool selection...")
	selectedTools, err := agent.SelectTools(ctx, "search for AI agents", inputState)
	if err != nil {
		fmt.Printf("Error selecting tools: %v\n", err)
	} else {
		fmt.Printf("Selected tools: %v\n", selectedTools)
	}

	// Test agent execution
	fmt.Println("\nTesting agent execution...")
	outputState, err := agent.Run(ctx, inputState)
	if err != nil {
		fmt.Printf("Error running agent: %v\n", err)
	} else {
		fmt.Printf("Agent execution completed successfully\n")
		fmt.Printf("Output state keys: %v\n", outputState.Keys())
	}

	fmt.Println("\nMCP-Aware Agent test completed!")
}
