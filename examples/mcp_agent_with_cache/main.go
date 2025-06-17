package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kunalkushwaha/agentflow/core"
)

func main() {
	fmt.Println("🚀 MCP Agent with Caching Demo")
	fmt.Println("================================")

	ctx := context.Background()

	// Step 1: Configure MCP with cache-enabled agent
	fmt.Println("📡 Step 1: Setting up MCP Agent with Cache")

	// Configure cache settings
	cacheConfig := core.MCPCacheConfig{
		Enabled:         true,
		DefaultTTL:      5 * time.Minute,
		MaxSize:         50, // 50 MB
		MaxKeys:         1000,
		EvictionPolicy:  "lru",
		CleanupInterval: 2 * time.Minute,
		Backend:         "memory",
		ToolTTLs: map[string]time.Duration{
			"web_search": 2 * time.Minute,
			"summarize":  10 * time.Minute,
			"translate":  30 * time.Minute,
		},
	}

	// Configure agent with caching enabled
	agentConfig := core.MCPAgentConfig{
		MaxToolsPerExecution: 3,
		ToolSelectionTimeout: 30 * time.Second,
		ParallelExecution:    false,
		ExecutionTimeout:     2 * time.Minute,
		RetryFailedTools:     true,
		MaxRetries:           2,
		UseToolDescriptions:  true,
		ResultInterpretation: true,
		EnableCaching:        true,
		CacheConfig:          cacheConfig,
	}

	// Step 2: Create LLM provider (mock for demo)
	fmt.Println("🧠 Step 2: Setting up LLM Provider")
	llmProvider := &mockLLMProvider{}

	// Step 3: Create MCP Manager (mock for demo)
	fmt.Println("🔧 Step 3: Setting up MCP Manager")
	mcpManager := &mockMCPManager{}

	// Step 4: Create MCP-aware agent with caching
	fmt.Println("🤖 Step 4: Creating MCP Agent with Cache")
	agent := core.NewMCPAwareAgent("cache-demo-agent", llmProvider, mcpManager, agentConfig)

	fmt.Printf("✅ Agent created with caching: %t\n", agentConfig.EnableCaching)
	fmt.Printf("   🔧 Cache TTL: %v\n", cacheConfig.DefaultTTL)
	fmt.Printf("   📦 Cache Max Keys: %d\n", cacheConfig.MaxKeys)

	// Step 5: Test agent execution with various tools
	fmt.Println("\n🧪 Step 5: Testing Agent with Tool Execution")

	testCases := []struct {
		name        string
		description string
		query       string
	}{
		{
			name:        "web_search",
			description: "Web search with potential caching",
			query:       "search for latest AI developments",
		},
		{
			name:        "content_analysis",
			description: "Content analysis with caching",
			query:       "analyze the sentiment of user feedback",
		},
		{
			name:        "web_search_repeat",
			description: "Repeat web search (should hit cache)",
			query:       "search for latest AI developments",
		},
	}

	for i, testCase := range testCases {
		fmt.Printf("\n[%d] %s: %s\n", i+1, testCase.name, testCase.description)

		// Create input state with the query
		inputState := core.NewState()
		inputState.SetValue("query", testCase.query)
		inputState.SetValue("task", testCase.description)

		// Execute the agent
		start := time.Now()
		outputState, err := agent.Run(ctx, inputState)
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("❌ Execution failed: %v\n", err)
			continue
		}

		fmt.Printf("✅ Completed in %v\n", elapsed)

		// Extract result from output state
		if result, exists := outputState.GetValue("result"); exists {
			fmt.Printf("   📝 Result: %v\n", result)
		}

		if tools, exists := outputState.GetValue("tools_used"); exists {
			fmt.Printf("   🔧 Tools used: %v\n", tools)
		}
	}

	// Step 6: Summary
	fmt.Println("\n📈 Step 6: Cache Performance Summary")
	fmt.Println("🎯 Key Benefits of MCP Agent Caching:")
	fmt.Println("   • Faster response times for repeated queries")
	fmt.Println("   • Reduced load on external MCP servers")
	fmt.Println("   • Configurable TTL per tool type")
	fmt.Println("   • LRU eviction for memory management")

	fmt.Println("\n🎉 MCP Agent with Cache Demo Completed!")
	fmt.Println("✅ Ready for production use with smart caching")
}

// mockLLMProvider provides a simple mock LLM for demo purposes.
type mockLLMProvider struct{}

func (m *mockLLMProvider) Call(ctx context.Context, prompt core.Prompt) (*core.LLMResponse, error) {
	// Simulate LLM tool selection based on the query
	query := prompt.User

	var tools []string
	if contains(query, "search") {
		tools = append(tools, "web_search")
	}
	if contains(query, "analyze") || contains(query, "sentiment") {
		tools = append(tools, "sentiment_analysis")
	}
	if contains(query, "translate") {
		tools = append(tools, "translate_text")
	}

	// Default fallback
	if len(tools) == 0 {
		tools = []string{"web_search"}
	}

	// Return JSON array of tool names
	content := fmt.Sprintf(`["%s"]`, strings.Join(tools, `", "`))

	return &core.LLMResponse{
		Content: content,
		Usage: core.LLMUsage{
			PromptTokens:     len(prompt.User) / 4,
			CompletionTokens: len(content) / 4,
			TotalTokens:      (len(prompt.User) + len(content)) / 4,
		},
	}, nil
}

func (m *mockLLMProvider) CallWithImage(ctx context.Context, prompt core.Prompt, imageData []byte) (*core.LLMResponse, error) {
	return m.Call(ctx, prompt)
}

// mockMCPManager provides a simple mock MCP manager for demo purposes.
type mockMCPManager struct{}

func (m *mockMCPManager) ConnectToServer(config core.MCPServerConfig) error {
	return nil
}

func (m *mockMCPManager) DisconnectFromServer(serverName string) error {
	return nil
}

func (m *mockMCPManager) DisconnectAll() error {
	return nil
}

func (m *mockMCPManager) ExecuteTool(ctx context.Context, serverName, toolName string, args map[string]interface{}) (core.MCPToolResult, error) {
	// Simulate tool execution with delay
	time.Sleep(100 * time.Millisecond)

	result := core.MCPToolResult{
		ToolName:   toolName,
		ServerName: serverName,
		Success:    true,
		Content: []core.MCPContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Mock result from %s: processed %v", toolName, args),
			},
		},
		Duration: 100 * time.Millisecond,
	}

	return result, nil
}

func (m *mockMCPManager) GetAvailableTools() []core.MCPToolInfo {
	return []core.MCPToolInfo{
		{
			Name:        "web_search",
			Description: "Search the web for information",
			ServerName:  "web-service",
			Parameters: map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query",
				},
			},
		},
		{
			Name:        "sentiment_analysis",
			Description: "Analyze sentiment of text",
			ServerName:  "nlp-service",
			Parameters: map[string]interface{}{
				"text": map[string]interface{}{
					"type":        "string",
					"description": "Text to analyze",
				},
			},
		},
		{
			Name:        "translate_text",
			Description: "Translate text between languages",
			ServerName:  "translation-service",
			Parameters: map[string]interface{}{
				"text": map[string]interface{}{
					"type":        "string",
					"description": "Text to translate",
				},
				"target_language": map[string]interface{}{
					"type":        "string",
					"description": "Target language code",
				},
			},
		},
	}
}

func (m *mockMCPManager) RefreshTools(ctx context.Context) error {
	return nil
}

func (m *mockMCPManager) GetConnectedServers() []string {
	return []string{"web-service", "nlp-service", "translation-service"}
}

func (m *mockMCPManager) GetServerStatus(serverName string) (core.MCPServerStatus, error) {
	return core.MCPServerStatus{
		Name:      serverName,
		Connected: true,
		LastPing:  time.Now(),
	}, nil
}

// Helper functions
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

var strings = struct {
	Join     func([]string, string) string
	Contains func(string, string) bool
	ToLower  func(string) string
}{
	Join: func(strs []string, sep string) string {
		if len(strs) == 0 {
			return ""
		}
		result := strs[0]
		for i := 1; i < len(strs); i++ {
			result += sep + strs[i]
		}
		return result
	},
	Contains: func(s, substr string) bool {
		return len(s) >= len(substr) && (substr == "" || s == substr ||
			(len(s) > len(substr) && (s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				func() bool {
					for i := 1; i < len(s)-len(substr)+1; i++ {
						if s[i:i+len(substr)] == substr {
							return true
						}
					}
					return false
				}())))
	},
	ToLower: func(s string) string {
		result := ""
		for _, r := range s {
			if r >= 'A' && r <= 'Z' {
				result += string(r + 32)
			} else {
				result += string(r)
			}
		}
		return result
	},
}
