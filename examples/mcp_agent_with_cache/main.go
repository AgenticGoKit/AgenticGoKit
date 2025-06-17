package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kunalkushwaha/agentflow/core"
	"github.com/kunalkushwaha/agentflow/internal/mcp"
)

func main() {
	fmt.Println("🚀 MCP Agent with Caching Demo")
	fmt.Println("================================")

	ctx := context.Background()

	// Step 1: Configure MCP Cache
	fmt.Println("📡 Step 1: Setting up MCP Cache System")

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

	// Step 2: Create cache manager with mock executor
	fmt.Println("🔧 Step 2: Setting up Cache Manager")
	mockExecutor := &MockMCPToolExecutor{}
	cacheManager, err := mcp.NewCacheManager(cacheConfig, mockExecutor)
	if err != nil {
		fmt.Printf("❌ Failed to create cache manager: %v\n", err)
		return
	}
	// Step 3: Initialize metrics for monitoring (metrics server will be available)
	fmt.Println("📊 Step 3: Setting up Metrics")
	metricsConfig := &mcp.MetricsConfig{
		Enabled: true,
		Port:    8082,
		Path:    "/metrics",
	}
	_ = mcp.NewMCPMetrics(metricsConfig, core.Logger()) // Start metrics server

	// Step 4: Test cache operations with various tools
	fmt.Println("\n🧪 Step 4: Testing Cache with Tool Execution")

	testCases := []struct {
		name        string
		description string
		toolName    string
		arguments   map[string]interface{}
	}{
		{
			name:        "web_search_1",
			description: "First web search (cache miss expected)",
			toolName:    "web_search",
			arguments:   map[string]interface{}{"query": "latest AI developments"},
		},
		{
			name:        "content_analysis",
			description: "Content analysis (new cache entry)",
			toolName:    "sentiment_analysis",
			arguments:   map[string]interface{}{"text": "user feedback analysis"},
		},
		{
			name:        "web_search_2",
			description: "Repeat web search (cache hit expected)",
			toolName:    "web_search",
			arguments:   map[string]interface{}{"query": "latest AI developments"},
		},
		{
			name:        "translate_text",
			description: "Translation task (long TTL)",
			toolName:    "translate_text",
			arguments:   map[string]interface{}{"text": "Hello world", "target": "Spanish"},
		},
	}
	for i, testCase := range testCases {
		fmt.Printf("\n[%d] %s: %s\n", i+1, testCase.name, testCase.description)

		// Create tool execution request
		execution := core.MCPToolExecution{
			ToolName:   testCase.toolName,
			ServerName: "demo-server",
			Arguments:  testCase.arguments,
		}

		// Execute the tool through cache manager (handles caching automatically)
		start := time.Now()
		result, err := cacheManager.ExecuteWithCache(ctx, execution)
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("❌ Execution failed: %v\n", err)
			continue
		}

		fmt.Printf("✅ Executed in %v\n", elapsed)
		fmt.Printf("   📝 Result: %s\n", result.Content[0].Text)
		fmt.Printf("   ⚡ Duration: %v\n", result.Duration)
	}

	// Step 5: Show cache statistics
	fmt.Println("\n📈 Step 5: Cache Performance Summary")
	stats, err := cacheManager.GetGlobalStats(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to get cache stats: %v\n", err)
	} else {
		fmt.Printf("📊 Cache Statistics:\n")
		fmt.Printf("   📦 Total Keys: %d\n", stats.TotalKeys)
		fmt.Printf("   🎯 Hit Rate: %.2f%%\n", stats.HitRate*100)
		fmt.Printf("   ✅ Hit Count: %d\n", stats.HitCount)
		fmt.Printf("   ❌ Miss Count: %d\n", stats.MissCount)
		fmt.Printf("   🗑️  Evictions: %d\n", stats.EvictionCount)
		fmt.Printf("   📏 Total Size: %d bytes\n", stats.TotalSize)
		fmt.Printf("   ⚡ Avg Latency: %v\n", stats.AverageLatency)
	}
	// Step 6: Test cache cleanup
	fmt.Println("\n🧹 Step 6: Testing Cache Management")

	// Get cache for a specific tool
	webSearchCache := cacheManager.GetCache("web_search", "demo-server")

	// Create a test cache key to check existence
	testArgs := map[string]string{"query": "latest AI developments"}
	testCacheKey := core.GenerateCacheKey("web_search", "demo-server", testArgs)

	// Check if entries exist
	exists, err := webSearchCache.Exists(ctx, testCacheKey)
	if err != nil {
		fmt.Printf("❌ Failed to check cache existence: %v\n", err)
	} else {
		fmt.Printf("🔍 Cache entry exists: %t\n", exists)
	}

	// Perform cleanup
	err = webSearchCache.Cleanup(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to cleanup cache: %v\n", err)
	} else {
		fmt.Printf("🧹 Cache cleanup completed successfully\n")
	}

	fmt.Println("\n🎯 Key Benefits of MCP Agent Caching:")
	fmt.Println("   • Faster response times for repeated queries")
	fmt.Println("   • Reduced load on external MCP servers")
	fmt.Println("   • Configurable TTL per tool type")
	fmt.Println("   • LRU eviction for memory management")
	fmt.Println("   • Real-time cache statistics and monitoring")

	fmt.Println("\n🎉 MCP Agent with Cache Demo Completed!")
	fmt.Println("✅ Ready for production use with smart caching")
}

// MockMCPToolExecutor provides a mock implementation for demo purposes
type MockMCPToolExecutor struct{}

func (m *MockMCPToolExecutor) ExecuteTool(ctx context.Context, execution core.MCPToolExecution) (core.MCPToolResult, error) {
	// Simulate different processing times for different tools
	var processingTime time.Duration
	var resultText string

	switch execution.ToolName {
	case "web_search":
		processingTime = 500 * time.Millisecond
		query := execution.Arguments["query"]
		resultText = fmt.Sprintf("Web search results for: %v", query)
	case "sentiment_analysis":
		processingTime = 200 * time.Millisecond
		text := execution.Arguments["text"]
		resultText = fmt.Sprintf("Sentiment analysis result for: %v (Positive: 85%%)", text)
	case "translate_text":
		processingTime = 300 * time.Millisecond
		text := execution.Arguments["text"]
		target := execution.Arguments["target"]
		resultText = fmt.Sprintf("Translated '%v' to %v: 'Hola mundo'", text, target)
	default:
		processingTime = 100 * time.Millisecond
		resultText = fmt.Sprintf("Mock result from %s tool", execution.ToolName)
	}

	// Simulate processing time
	time.Sleep(processingTime)

	return core.MCPToolResult{
		ToolName:   execution.ToolName,
		ServerName: execution.ServerName,
		Success:    true,
		Content: []core.MCPContent{{
			Type: "text",
			Text: resultText,
		}},
		Duration: processingTime,
	}, nil
}
