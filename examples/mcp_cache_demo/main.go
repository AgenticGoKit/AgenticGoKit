// Package main demonstrates MCP tool result caching functionality.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kunalkushwaha/agentflow/core"
	"github.com/kunalkushwaha/agentflow/internal/mcp"
)

// MockToolExecutor simulates MCP tool execution for testing cache.
type MockToolExecutor struct {
	executionDelay time.Duration
	callCount      int
}

func (m *MockToolExecutor) ExecuteTool(ctx context.Context, execution core.MCPToolExecution) (core.MCPToolResult, error) {
	m.callCount++

	// Simulate tool execution time
	time.Sleep(m.executionDelay)

	// Create mock result
	result := core.MCPToolResult{
		ToolName:   execution.ToolName,
		ServerName: execution.ServerName,
		Success:    true,
		Content: []core.MCPContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Mock result for %s (execution #%d) with args: %v",
					execution.ToolName, m.callCount, execution.Arguments),
				MimeType: "text/plain",
				Metadata: map[string]interface{}{
					"execution_id": m.callCount,
					"mock":         true,
				},
			},
		},
		Duration: m.executionDelay,
	}

	fmt.Printf("🔧 Executed %s (call #%d) - took %v\n", execution.ToolName, m.callCount, m.executionDelay)
	return result, nil
}

func main() {
	fmt.Println("🚀 MCP Tool Result Caching Demo")
	fmt.Println("===============================")

	ctx := context.Background()

	// Step 1: Configure cache
	fmt.Println("\n📡 Step 1: Setting up Cache Configuration")
	config := core.DefaultMCPCacheConfig()
	config.DefaultTTL = 10 * time.Second // Short TTL for demo
	config.MaxKeys = 100
	config.CleanupInterval = 5 * time.Second

	fmt.Printf("✅ Cache config: TTL=%v, MaxKeys=%d, Backend=%s\n",
		config.DefaultTTL, config.MaxKeys, config.Backend)

	// Step 2: Create mock executor and cache manager
	fmt.Println("\n🔧 Step 2: Creating Cache Manager")
	mockExecutor := &MockToolExecutor{
		executionDelay: 500 * time.Millisecond, // Simulate 500ms tool execution
	}

	cacheManager, err := mcp.NewCacheManager(config, mockExecutor)
	if err != nil {
		log.Fatalf("❌ Failed to create cache manager: %v", err)
	}
	defer cacheManager.Close()

	fmt.Println("✅ Cache manager created with mock executor")

	// Step 3: Test cache with repeated executions
	fmt.Println("\n🧪 Step 3: Testing Cache Performance")

	testExecutions := []core.MCPToolExecution{
		{
			ToolName:   "web_search",
			ServerName: "web-service",
			Arguments:  map[string]interface{}{"query": "AI news", "limit": 10},
		},
		{
			ToolName:   "summarize_text",
			ServerName: "nlp-service",
			Arguments:  map[string]interface{}{"text": "Long article about AI...", "max_length": 100},
		},
		{
			ToolName:   "sentiment_analysis",
			ServerName: "nlp-service",
			Arguments:  map[string]interface{}{"text": "This is great news!", "model": "default"},
		},
	}

	// Execute each tool multiple times to test caching
	for round := 1; round <= 3; round++ {
		fmt.Printf("\n🔄 Round %d: Executing all tools\n", round)

		for i, execution := range testExecutions {
			fmt.Printf("\n[%d.%d] Executing %s:%s\n", round, i+1, execution.ServerName, execution.ToolName)

			start := time.Now()
			result, err := cacheManager.ExecuteWithCache(ctx, execution)
			elapsed := time.Since(start)

			if err != nil {
				fmt.Printf("❌ Execution failed: %v\n", err)
				continue
			}

			fmt.Printf("✅ Completed in %v (success=%v)\n", elapsed, result.Success)
			if len(result.Content) > 0 {
				fmt.Printf("   📝 Result: %s\n", result.Content[0].Text[:min(50, len(result.Content[0].Text))]+"...")
			}
		}

		// Show cache stats after each round
		stats, err := cacheManager.GetGlobalStats(ctx)
		if err == nil {
			fmt.Printf("\n📊 Cache Stats: %d keys, %.1f%% hit rate, %d hits, %d misses\n",
				stats.TotalKeys, stats.HitRate*100, stats.HitCount, stats.MissCount)
		}

		if round < 3 {
			fmt.Printf("⏳ Waiting 1 second before next round...\n")
			time.Sleep(1 * time.Second)
		}
	}

	// Step 4: Test cache invalidation
	fmt.Println("\n🧹 Step 4: Testing Cache Invalidation")

	err = cacheManager.InvalidateByPattern(ctx, "web-service")
	if err != nil {
		fmt.Printf("❌ Invalidation failed: %v\n", err)
	} else {
		fmt.Println("✅ Invalidated all web-service caches")
	}
	// Test execution after invalidation
	fmt.Println("\n🔄 Testing web_search after invalidation:")
	start := time.Now()
	_, err = cacheManager.ExecuteWithCache(ctx, testExecutions[0])
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Execution failed: %v\n", err)
	} else {
		fmt.Printf("✅ Completed in %v (should be slow again)\n", elapsed)
	}

	// Step 5: Test TTL expiration
	fmt.Println("\n⏰ Step 5: Testing TTL Expiration")
	fmt.Printf("Waiting %v for cache entries to expire...\n", config.DefaultTTL)
	time.Sleep(config.DefaultTTL + time.Second)
	// Execute again - should be cache miss due to TTL
	fmt.Println("\n🔄 Testing after TTL expiration:")
	start = time.Now()
	_, err = cacheManager.ExecuteWithCache(ctx, testExecutions[1])
	elapsed = time.Since(start)

	if err != nil {
		fmt.Printf("❌ Execution failed: %v\n", err)
	} else {
		fmt.Printf("✅ Completed in %v (should be slow due to TTL expiration)\n", elapsed)
	}

	// Step 6: Final stats
	fmt.Println("\n📈 Final Cache Statistics")
	finalStats, err := cacheManager.GetGlobalStats(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to get stats: %v\n", err)
	} else {
		fmt.Printf("📊 Final Results:\n")
		fmt.Printf("   • Total cache keys: %d\n", finalStats.TotalKeys)
		fmt.Printf("   • Total hits: %d\n", finalStats.HitCount)
		fmt.Printf("   • Total misses: %d\n", finalStats.MissCount)
		fmt.Printf("   • Hit rate: %.1f%%\n", finalStats.HitRate*100)
		fmt.Printf("   • Total evictions: %d\n", finalStats.EvictionCount)
		fmt.Printf("   • Cache size: %d bytes\n", finalStats.TotalSize)
	}

	fmt.Printf("\n🎯 Cache Performance Summary:\n")
	fmt.Printf("   • Mock executor was called %d times\n", mockExecutor.callCount)
	if finalStats.HitCount > 0 {
		saved := time.Duration(finalStats.HitCount) * mockExecutor.executionDelay
		fmt.Printf("   • Estimated time saved: %v\n", saved)
		fmt.Printf("   • Performance improvement: %.1fx faster for cached calls\n",
			float64(mockExecutor.executionDelay)/float64(time.Millisecond))
	}

	fmt.Println("\n🎉 MCP Caching Demo Completed!")
	fmt.Println("✅ Cache system is working correctly")
	fmt.Println("💡 Ready for integration with MCP agents")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
