// Package main demonstrates the complete MCP integration in AgentFlow.
// This example shows how to create a production-ready setup with MCP-aware agents.
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kunalkushwaha/agentflow/core"
	"github.com/kunalkushwaha/agentflow/internal/factory"
	"github.com/kunalkushwaha/agentflow/internal/llm"
)

// OllamaLLMProvider wraps the Ollama adapter for production use with tool selection.
type OllamaLLMProvider struct {
	adapter *llm.OllamaAdapter
}

func NewOllamaLLMProvider() (*OllamaLLMProvider, error) {
	// Create Ollama adapter with llama3.2:latest model (no API key needed for local)
	adapter, err := llm.NewOllamaAdapter("", "llama3.2:latest", 512, 0.7)
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama adapter: %w", err)
	}

	return &OllamaLLMProvider{adapter: adapter}, nil
}

func (p *OllamaLLMProvider) Call(ctx context.Context, prompt core.Prompt) (core.Response, error) {
	fmt.Printf("🧠 Ollama analyzing query: '%s'\n", prompt.User)

	// Add timeout to prevent hanging
	timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Create a well-structured prompt for tool selection
	systemPrompt := `You are an intelligent agent that selects appropriate tools based on user queries. 
Available tools:
- web_search: Advanced web search with filtering
- content_fetch: Fetch and parse web content  
- url_validator: Validate and check URL status
- summarize_text: AI-powered text summarization
- sentiment_analysis: Analyze text sentiment
- entity_extraction: Extract named entities
- compute_metric: Compute statistical metrics
- data_transform: Transform data formats

Based on the user query, select the most relevant tools as a JSON array of tool names.
Only respond with the JSON array, nothing else.

Examples:
Query: "search for AI news" -> ["web_search"]
Query: "get content and summarize it" -> ["content_fetch", "summarize_text"]
Query: "analyze data metrics" -> ["compute_metric", "data_transform"]`

	// Convert core types to internal LLM types
	internalPrompt := llm.Prompt{
		System: systemPrompt,
		User:   prompt.User,
		Parameters: llm.ModelParameters{
			MaxTokens:   func() *int32 { v := int32(100); return &v }(),
			Temperature: func() *float32 { v := float32(0.3); return &v }(),
		},
	}

	// Call Ollama
	response, err := p.adapter.Call(timeoutCtx, internalPrompt)
	if err != nil {
		// Fallback to simple tool selection on error
		fmt.Printf("⚠️  Ollama call failed (using fallback): %v\n", err)
		toolSelection := p.fallbackToolSelection(prompt.User)
		return core.Response{
			Content:      toolSelection,
			Usage:        core.UsageStats{PromptTokens: 50, CompletionTokens: 20, TotalTokens: 70},
			FinishReason: "fallback",
		}, nil
	}

	fmt.Printf("✅ Ollama response received: %s\n", response.Content)

	// Clean up the response to extract just the JSON array
	content := strings.TrimSpace(response.Content)
	if !strings.HasPrefix(content, "[") {
		// Try to extract JSON array from the response
		if start := strings.Index(content, "["); start != -1 {
			if end := strings.LastIndex(content, "]"); end != -1 && end > start {
				content = content[start : end+1]
			}
		} else {
			// Fallback if no JSON found
			fmt.Printf("⚠️  Could not parse JSON from Ollama response, using fallback\n")
			content = p.fallbackToolSelection(prompt.User)
		}
	}

	fmt.Printf("🎯 Ollama selected tools: %s\n", content)

	return core.Response{
		Content:      content,
		Usage:        core.UsageStats{PromptTokens: 75, CompletionTokens: 25, TotalTokens: 100},
		FinishReason: "stop",
	}, nil
}

func (p *OllamaLLMProvider) fallbackToolSelection(query string) string {
	// Simple keyword-based fallback
	query = strings.ToLower(query)
	switch {
	case strings.Contains(query, "search") || strings.Contains(query, "find"):
		return `["web_search", "content_fetch"]`
	case strings.Contains(query, "summarize") || strings.Contains(query, "summary"):
		return `["content_fetch", "summarize_text"]`
	case strings.Contains(query, "analyze") || strings.Contains(query, "data"):
		return `["compute_metric", "sentiment_analysis"]`
	case strings.Contains(query, "extract") || strings.Contains(query, "entities"):
		return `["entity_extraction", "content_fetch"]`
	default:
		return `["web_search"]`
	}
}

func (p *OllamaLLMProvider) Stream(ctx context.Context, prompt core.Prompt) (<-chan core.Token, error) {
	return nil, fmt.Errorf("streaming not implemented")
}

func (p *OllamaLLMProvider) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return nil, fmt.Errorf("embeddings not implemented")
}

// ProductionMCPManager simulates a production MCP manager with multiple servers.
type ProductionMCPManager struct {
	servers map[string][]core.MCPToolInfo
}

func NewProductionMCPManager() *ProductionMCPManager {
	return &ProductionMCPManager{
		servers: map[string][]core.MCPToolInfo{
			"web-services": {
				{Name: "web_search", Description: "Advanced web search with filtering", ServerName: "web-services"},
				{Name: "content_fetch", Description: "Fetch and parse web content", ServerName: "web-services"},
				{Name: "url_validator", Description: "Validate and check URL status", ServerName: "web-services"},
			},
			"nlp-services": {
				{Name: "summarize_text", Description: "AI-powered text summarization", ServerName: "nlp-services"},
				{Name: "sentiment_analysis", Description: "Analyze text sentiment", ServerName: "nlp-services"},
				{Name: "entity_extraction", Description: "Extract named entities", ServerName: "nlp-services"},
			},
			"data-services": {
				{Name: "compute_metric", Description: "Compute statistical metrics", ServerName: "data-services"},
				{Name: "data_transform", Description: "Transform data formats", ServerName: "data-services"},
			},
		},
	}
}

func (m *ProductionMCPManager) Connect(ctx context.Context, serverName string) error {
	if _, exists := m.servers[serverName]; exists {
		fmt.Printf("🔗 Connected to production MCP server: %s\n", serverName)
		return nil
	}
	return fmt.Errorf("server not found: %s", serverName)
}

func (m *ProductionMCPManager) Disconnect(serverName string) error {
	fmt.Printf("🔌 Disconnected from server: %s\n", serverName)
	return nil
}

func (m *ProductionMCPManager) DisconnectAll() error {
	fmt.Println("🔌 Disconnected from all servers")
	return nil
}

func (m *ProductionMCPManager) DiscoverServers(ctx context.Context) ([]core.MCPServerInfo, error) {
	var servers []core.MCPServerInfo
	basePort := 8800

	for serverName := range m.servers {
		servers = append(servers, core.MCPServerInfo{
			Name:        serverName,
			Type:        "tcp",
			Address:     "localhost",
			Port:        basePort,
			Status:      "connected",
			Description: fmt.Sprintf("Production %s server", serverName),
		})
		basePort++
	}

	return servers, nil
}

func (m *ProductionMCPManager) ListConnectedServers() []string {
	var names []string
	for serverName := range m.servers {
		names = append(names, serverName)
	}
	return names
}

func (m *ProductionMCPManager) GetServerInfo(serverName string) (*core.MCPServerInfo, error) {
	servers, _ := m.DiscoverServers(context.Background())
	for _, server := range servers {
		if server.Name == serverName {
			return &server, nil
		}
	}
	return nil, fmt.Errorf("server not found: %s", serverName)
}

func (m *ProductionMCPManager) RefreshTools(ctx context.Context) error {
	fmt.Println("🔄 Refreshing tools from production servers...")
	time.Sleep(100 * time.Millisecond) // Simulate network delay
	return nil
}

func (m *ProductionMCPManager) GetAvailableTools() []core.MCPToolInfo {
	var allTools []core.MCPToolInfo
	for _, tools := range m.servers {
		allTools = append(allTools, tools...)
	}
	return allTools
}

func (m *ProductionMCPManager) GetToolsFromServer(serverName string) []core.MCPToolInfo {
	if tools, exists := m.servers[serverName]; exists {
		return tools
	}
	return []core.MCPToolInfo{}
}

func (m *ProductionMCPManager) HealthCheck(ctx context.Context) map[string]core.MCPHealthStatus {
	health := make(map[string]core.MCPHealthStatus)

	for serverName, tools := range m.servers {
		health[serverName] = core.MCPHealthStatus{
			Status:       "healthy",
			LastCheck:    time.Now(),
			ResponseTime: time.Duration(5+len(serverName)) * time.Millisecond,
			ToolCount:    len(tools),
		}
	}

	return health
}

func (m *ProductionMCPManager) GetMetrics() core.MCPMetrics {
	metrics := core.MCPMetrics{
		ConnectedServers: len(m.servers),
		TotalTools:       len(m.GetAvailableTools()),
		ToolExecutions:   42,
		AverageLatency:   8 * time.Millisecond,
		ErrorRate:        0.02,
		ServerMetrics:    make(map[string]core.MCPServerMetrics),
	}

	for serverName, tools := range m.servers {
		metrics.ServerMetrics[serverName] = core.MCPServerMetrics{
			ToolCount:        len(tools),
			Executions:       15,
			SuccessfulCalls:  14,
			FailedCalls:      1,
			AverageLatency:   8 * time.Millisecond,
			LastActivity:     time.Now(),
			ConnectionUptime: 2 * time.Hour,
		}
	}

	return metrics
}

func main() {
	fmt.Println("🚀 AgentFlow MCP Integration - Production Demo with Ollama")
	fmt.Println("=========================================================")

	ctx := context.Background()

	// Step 1: Initialize production MCP infrastructure
	fmt.Println("\n📡 Step 1: Initializing MCP Infrastructure")

	mcpManager := NewProductionMCPManager()

	// Create Ollama LLM provider
	fmt.Println("🦙 Initializing Ollama (llama3.2:latest)...")
	llmProvider, err := NewOllamaLLMProvider()
	if err != nil {
		log.Fatalf("❌ Failed to initialize Ollama: %v", err)
	}
	fmt.Println("✅ Ollama initialized successfully")

	// Create tool registry with MCP integration
	registry := factory.NewDefaultToolRegistry()
	fmt.Printf("📦 Created tool registry with %d built-in tools\n", len(registry.List()))
	// Auto-discover and register MCP tools
	err = factory.AutoDiscoverMCPTools(ctx, registry, mcpManager)
	if err != nil {
		log.Printf("⚠️  Warning: %v", err)
	}

	fmt.Printf("🛠️  Total tools available: %d\n", len(registry.List()))

	// Step 2: Create MCP-aware agent
	fmt.Println("\n🤖 Step 2: Creating MCP-Aware Agent")

	agentConfig := core.DefaultMCPAgentConfig()
	agentConfig.MaxToolsPerExecution = 3
	agentConfig.ParallelExecution = false
	agentConfig.RetryFailedTools = true

	agent := core.NewMCPAwareAgent("production-agent", llmProvider, mcpManager, agentConfig)
	fmt.Printf("✅ Created agent: %s with %d available tools\n", agent.Name(), len(agent.GetAvailableMCPTools()))
	// Step 3: Test different scenarios with Ollama-based tool selection
	testScenarios := []struct {
		name  string
		query string
	}{
		{"Web Research", "search for the latest developments in AI and machine learning technologies"},
		{"Content Processing", "fetch content from a research paper and provide a comprehensive summary"},
		{"Data Analysis", "analyze sentiment in user reviews and compute statistical metrics"},
		{"Entity Extraction", "extract named entities and key information from the provided text"},
		{"Multi-tool Workflow", "search for AI news, fetch the content, and then summarize the findings"},
	}
	for i, scenario := range testScenarios {
		fmt.Printf("\n🧪 Step %d: Testing Scenario - %s\n", i+3, scenario.name)
		fmt.Printf("📝 Query: %s\n", scenario.query)

		// Add delay between scenarios to prevent overwhelming Ollama
		if i > 0 {
			fmt.Printf("⏳ Waiting 2 seconds before next scenario...\n")
			time.Sleep(2 * time.Second)
		}

		// Create state for the scenario
		state := core.NewState()
		state.Set("query", scenario.query)
		state.Set("scenario", scenario.name)
		state.Set("user_id", "production-user")

		// Run the agent with timeout
		scenarioCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		result, err := agent.Run(scenarioCtx, state)
		cancel()

		if err != nil {
			fmt.Printf("❌ Scenario failed: %v\n", err)
			continue
		}

		// Display results
		if results, exists := result.Get("mcp_results"); exists {
			if resultSlice, ok := results.([]map[string]interface{}); ok {
				fmt.Printf("📊 Executed %d tools:\n", len(resultSlice))
				for j, res := range resultSlice {
					fmt.Printf("   %d. %s: %v\n", j+1, res["tool_name"], res["success"])
				}
			}
		}
	}

	// Step 4: Display comprehensive system status
	fmt.Println("\n📈 Step 6: System Status & Metrics")

	// Health check
	fmt.Println("\n🏥 Server Health:")
	health := mcpManager.HealthCheck(ctx)
	for serverName, status := range health {
		fmt.Printf("   • %s: %s (response: %v, tools: %d)\n",
			serverName, status.Status, status.ResponseTime, status.ToolCount)
	}

	// Metrics
	fmt.Println("\n📊 Performance Metrics:")
	metrics := mcpManager.GetMetrics()
	fmt.Printf("   • Connected servers: %d\n", metrics.ConnectedServers)
	fmt.Printf("   • Total tools: %d\n", metrics.TotalTools)
	fmt.Printf("   • Tool executions: %d\n", metrics.ToolExecutions)
	fmt.Printf("   • Average latency: %v\n", metrics.AverageLatency)
	fmt.Printf("   • Error rate: %.2f%%\n", metrics.ErrorRate*100)

	// Tool registry validation
	fmt.Println("\n✅ Final Validation:")
	err = factory.ValidateToolRegistryIntegration(registry, mcpManager)
	if err != nil {
		fmt.Printf("⚠️  Warning: %v\n", err)
	} else {
		fmt.Println("✅ All systems operational")
	}

	// Display summary
	allTools := registry.List()
	mcpTools := factory.GetMCPToolsFromRegistry(registry, mcpManager)
	builtinTools := len(allTools) - len(mcpTools)

	fmt.Println("\n📋 Integration Summary:")
	fmt.Printf("   • Built-in tools: %d\n", builtinTools)
	fmt.Printf("   • MCP tools: %d\n", len(mcpTools))
	fmt.Printf("   • Total unified tools: %d\n", len(allTools))
	fmt.Printf("   • Connected MCP servers: %d\n", len(mcpManager.ListConnectedServers()))
	fmt.Println("\n🎉 Production Demo with Ollama completed successfully!")
	fmt.Println("💡 The MCP integration with Ollama LLM is ready for production use!")
	fmt.Println("🦙 Dynamic tool selection powered by llama3.2:latest model")
}
