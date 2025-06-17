// Package mcp provides cache manager implementation for MCP tools.
package mcp

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kunalkushwaha/agentflow/core"
)

// CacheManager implements MCPCacheManager and manages cache instances for MCP tools.
type CacheManager struct {
	caches   map[string]core.MCPCache
	config   core.MCPCacheConfig
	executor MCPToolExecutor // Interface to execute tools without cache
}

// MCPToolExecutor defines the interface for executing MCP tools.
type MCPToolExecutor interface {
	ExecuteTool(ctx context.Context, execution core.MCPToolExecution) (core.MCPToolResult, error)
}

// NewCacheManager creates a new cache manager instance.
func NewCacheManager(config core.MCPCacheConfig, executor MCPToolExecutor) (*CacheManager, error) {
	if !config.Enabled {
		return &CacheManager{
			caches:   make(map[string]core.MCPCache),
			config:   config,
			executor: executor,
		}, nil
	}

	manager := &CacheManager{
		caches:   make(map[string]core.MCPCache),
		config:   config,
		executor: executor,
	}

	return manager, nil
}

// GetCache returns a cache instance for a specific tool or server.
func (cm *CacheManager) GetCache(toolName, serverName string) core.MCPCache {
	if !cm.config.Enabled {
		return &NoOpCache{} // Return no-op cache when disabled
	}

	cacheKey := fmt.Sprintf("%s:%s", serverName, toolName)

	if cache, exists := cm.caches[cacheKey]; exists {
		return cache
	}

	// Create new cache instance for this tool/server combination
	cache, err := NewMemoryCache(cm.config)
	if err != nil {
		log.Printf("Failed to create cache for %s: %v", cacheKey, err)
		return &NoOpCache{}
	}

	cm.caches[cacheKey] = cache
	return cache
}

// ExecuteWithCache executes a tool with caching support.
func (cm *CacheManager) ExecuteWithCache(ctx context.Context, execution core.MCPToolExecution) (core.MCPToolResult, error) {
	if !cm.config.Enabled {
		return cm.executor.ExecuteTool(ctx, execution)
	}

	// Convert arguments to string map for cache key
	args := make(map[string]string)
	for k, v := range execution.Arguments {
		args[k] = fmt.Sprintf("%v", v)
	}

	// Generate cache key
	cacheKey := core.GenerateCacheKey(execution.ToolName, execution.ServerName, args)
	cache := cm.GetCache(execution.ToolName, execution.ServerName)

	// Try to get from cache first
	cached, err := cache.Get(ctx, cacheKey)
	if err != nil {
		log.Printf("Cache get error for %s: %v", execution.ToolName, err)
	} else if cached != nil {
		log.Printf("📦 Cache HIT for %s:%s", execution.ServerName, execution.ToolName)
		return cached.Result, nil
	}

	log.Printf("📦 Cache MISS for %s:%s", execution.ServerName, execution.ToolName)

	// Execute the tool
	start := time.Now()
	result, err := cm.executor.ExecuteTool(ctx, execution)
	executionTime := time.Since(start)

	if err != nil {
		return result, err
	}

	// Cache the result if successful
	if result.Success {
		ttl := cm.getTTLForTool(execution.ToolName)

		// Add execution time to the first content item's metadata
		if len(result.Content) > 0 {
			if result.Content[0].Metadata == nil {
				result.Content[0].Metadata = make(map[string]interface{})
			}
			result.Content[0].Metadata["execution_time"] = executionTime
			result.Content[0].Metadata["cached_at"] = time.Now()
		}

		err = cache.Set(ctx, cacheKey, result, ttl)
		if err != nil {
			log.Printf("Cache set error for %s: %v", execution.ToolName, err)
		} else {
			log.Printf("📦 Cached result for %s:%s (TTL: %v)", execution.ServerName, execution.ToolName, ttl)
		}
	}

	return result, nil
}

// InvalidateByPattern invalidates cache entries matching a pattern.
func (cm *CacheManager) InvalidateByPattern(ctx context.Context, pattern string) error {
	if !cm.config.Enabled {
		return nil
	}

	invalidated := 0
	for cacheKey, cache := range cm.caches {
		if strings.Contains(cacheKey, pattern) {
			err := cache.Clear(ctx)
			if err != nil {
				log.Printf("Failed to clear cache %s: %v", cacheKey, err)
			} else {
				invalidated++
			}
		}
	}

	log.Printf("📦 Invalidated %d cache instances matching pattern: %s", invalidated, pattern)
	return nil
}

// GetGlobalStats returns aggregated cache statistics.
func (cm *CacheManager) GetGlobalStats(ctx context.Context) (core.MCPCacheStats, error) {
	if !cm.config.Enabled {
		return core.MCPCacheStats{}, nil
	}

	totalStats := core.MCPCacheStats{}
	cacheCount := 0

	for _, cache := range cm.caches {
		stats, err := cache.Stats(ctx)
		if err != nil {
			continue
		}

		totalStats.TotalKeys += stats.TotalKeys
		totalStats.HitCount += stats.HitCount
		totalStats.MissCount += stats.MissCount
		totalStats.EvictionCount += stats.EvictionCount
		totalStats.TotalSize += stats.TotalSize
		cacheCount++
	}

	// Calculate aggregated hit rate
	total := totalStats.HitCount + totalStats.MissCount
	if total > 0 {
		totalStats.HitRate = float64(totalStats.HitCount) / float64(total)
	}

	// Average latency (simplified)
	if cacheCount > 0 {
		totalStats.AverageLatency = time.Millisecond
	}

	return totalStats, nil
}

// Configure updates cache configuration.
func (cm *CacheManager) Configure(config core.MCPCacheConfig) error {
	cm.config = config

	// If caching is disabled, clear all caches
	if !config.Enabled {
		for _, cache := range cm.caches {
			cache.Close()
		}
		cm.caches = make(map[string]core.MCPCache)
	}

	return nil
}

// getTTLForTool returns the TTL for a specific tool.
func (cm *CacheManager) getTTLForTool(toolName string) time.Duration {
	if ttl, exists := cm.config.ToolTTLs[toolName]; exists {
		return ttl
	}
	return cm.config.DefaultTTL
}

// Close closes all cache instances and releases resources.
func (cm *CacheManager) Close() error {
	for _, cache := range cm.caches {
		cache.Close()
	}
	cm.caches = make(map[string]core.MCPCache)
	return nil
}

// Shutdown cleanly shuts down the cache manager and all cache instances.
func (cm *CacheManager) Shutdown() error {
	return cm.Close()
}

// NoOpCache implements MCPCache with no-op operations for when caching is disabled.
type NoOpCache struct{}

func (c *NoOpCache) Get(ctx context.Context, key core.MCPCacheKey) (*core.MCPCachedResult, error) {
	return nil, nil // Always cache miss
}

func (c *NoOpCache) Set(ctx context.Context, key core.MCPCacheKey, result core.MCPToolResult, ttl time.Duration) error {
	return nil // No-op
}

func (c *NoOpCache) Delete(ctx context.Context, key core.MCPCacheKey) error {
	return nil // No-op
}

func (c *NoOpCache) Clear(ctx context.Context) error {
	return nil // No-op
}

func (c *NoOpCache) Exists(ctx context.Context, key core.MCPCacheKey) (bool, error) {
	return false, nil // Never exists
}

func (c *NoOpCache) Stats(ctx context.Context) (core.MCPCacheStats, error) {
	return core.MCPCacheStats{}, nil // Empty stats
}

func (c *NoOpCache) Cleanup(ctx context.Context) error {
	return nil // No-op
}

func (c *NoOpCache) Close() error {
	return nil // No-op
}

// init registers the cache manager factory with the core package.
func init() {
	// Register our cache manager factory with the core package
	core.SetCacheManagerFactory(func(config core.MCPCacheConfig, executor core.MCPToolExecutor) (core.MCPCacheManager, error) {
		// Wrap the core.MCPToolExecutor to match our internal interface
		internalExecutor := &executorWrapper{executor: executor}

		// Create the cache manager using our internal implementation
		return NewCacheManager(config, internalExecutor)
	})
}

// executorWrapper wraps core.MCPToolExecutor to match our internal interface.
type executorWrapper struct {
	executor core.MCPToolExecutor
}

// ExecuteTool implements the internal MCPToolExecutor interface.
func (w *executorWrapper) ExecuteTool(ctx context.Context, execution core.MCPToolExecution) (core.MCPToolResult, error) {
	return w.executor.ExecuteTool(ctx, execution)
}
