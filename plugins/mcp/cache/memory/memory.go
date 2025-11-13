package memorycache

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
)

type realMCPCache struct {
	data map[string]*core.MCPCachedResult
	mu   sync.RWMutex
}

func newRealMCPCache() *realMCPCache {
	return &realMCPCache{data: make(map[string]*core.MCPCachedResult)}
}

func (c *realMCPCache) Get(ctx context.Context, key core.MCPCacheKey) (*core.MCPCachedResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keyStr := c.keyToString(key)
	result, exists := c.data[keyStr]
	if !exists {
		return nil, fmt.Errorf("cache miss")
	}
	// Check if expired
	if time.Since(result.Timestamp) > result.TTL {
		delete(c.data, keyStr)
		return nil, fmt.Errorf("cache expired")
	}

	return result, nil
}

func (c *realMCPCache) Set(ctx context.Context, key core.MCPCacheKey, result core.MCPToolResult, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	keyStr := c.keyToString(key)
	cachedResult := &core.MCPCachedResult{
		Key:       key,
		Result:    result,
		Timestamp: time.Now(),
		TTL:       ttl,
		Metadata:  make(map[string]interface{}),
	}

	c.data[keyStr] = cachedResult
	return nil
}

func (c *realMCPCache) Delete(ctx context.Context, key core.MCPCacheKey) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	keyStr := c.keyToString(key)
	delete(c.data, keyStr)
	return nil
}

func (c *realMCPCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*core.MCPCachedResult)
	return nil
}

func (c *realMCPCache) Exists(ctx context.Context, key core.MCPCacheKey) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keyStr := c.keyToString(key)
	_, exists := c.data[keyStr]
	return exists, nil
}

func (c *realMCPCache) Stats(ctx context.Context) (core.MCPCacheStats, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return core.MCPCacheStats{
		HitCount:  0,
		MissCount: 0,
		HitRate:   0.0,
		TotalKeys: int(len(c.data)),
		TotalSize: 0,
	}, nil
}

func (c *realMCPCache) Cleanup(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, result := range c.data {
		if now.Sub(result.Timestamp) > result.TTL {
			delete(c.data, key)
		}
	}

	return nil
}

func (c *realMCPCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*core.MCPCachedResult)
	return nil
}

func (c *realMCPCache) keyToString(key core.MCPCacheKey) string {
	return fmt.Sprintf("%s:%s:%s", key.ToolName, key.ServerName, key.Hash)
}

type realMCPCacheManager struct {
	config core.MCPCacheConfig
	caches map[string]core.MCPCache
	mu     sync.RWMutex
}

func createRealMCPCacheManager(config core.MCPCacheConfig) (core.MCPCacheManager, error) {
	return &realMCPCacheManager{
		config: config,
		caches: make(map[string]core.MCPCache),
	}, nil
}

func (cm *realMCPCacheManager) GetCache(toolName, serverName string) core.MCPCache {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	key := toolName + ":" + serverName
	cache, exists := cm.caches[key]
	if !exists {
		cache = newRealMCPCache()
		cm.caches[key] = cache
	}

	return cache
}

func (cm *realMCPCacheManager) ExecuteWithCache(ctx context.Context, execution core.MCPToolExecution) (core.MCPToolResult, error) {
	args := make(map[string]string)
	for k, v := range execution.Arguments {
		args[k] = fmt.Sprintf("%v", v)
	}

	cacheKey := core.GenerateCacheKey(execution.ToolName, execution.ServerName, args)
	cache := cm.GetCache(execution.ToolName, execution.ServerName)
	if cm.config.Enabled {
		if cached, err := cache.Get(ctx, cacheKey); err == nil {
			return cached.Result, nil
		}
	}

	manager := core.GetMCPManager()
	if manager == nil {
		return core.MCPToolResult{}, fmt.Errorf("MCP manager not initialized")
	}
	exec, ok := manager.(core.MCPToolExecutor)
	if !ok {
		return core.MCPToolResult{}, fmt.Errorf("MCP manager does not support direct tool execution. Import a transport plugin that implements MCPToolExecutor")
	}

	result, err := exec.ExecuteTool(ctx, execution.ToolName, execution.Arguments)
	if err != nil {
		return result, err
	}
	if cm.config.Enabled && result.Success {
		ttl := cm.config.DefaultTTL
		_ = cache.Set(ctx, cacheKey, result, ttl)
	}

	return result, nil
}

func (cm *realMCPCacheManager) InvalidateByPattern(ctx context.Context, pattern string) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for key, cache := range cm.caches {
		if strings.Contains(key, pattern) {
			if err := cache.Clear(ctx); err != nil {
				return fmt.Errorf("failed to clear cache for %s: %w", key, err)
			}
		}
	}
	return nil
}

func (cm *realMCPCacheManager) GetGlobalStats(ctx context.Context) (core.MCPCacheStats, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var total core.MCPCacheStats
	for _, cache := range cm.caches {
		stats, err := cache.Stats(ctx)
		if err != nil {
			continue
		}
		total.HitCount += stats.HitCount
		total.MissCount += stats.MissCount
		total.TotalKeys += stats.TotalKeys
		total.TotalSize += stats.TotalSize
	}
	if total.HitCount+total.MissCount > 0 {
		total.HitRate = float64(total.HitCount) / float64(total.HitCount+total.MissCount)
	}
	return total, nil
}

func (cm *realMCPCacheManager) Shutdown() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	for _, cache := range cm.caches {
		_ = cache.Clear(context.Background())
	}
	cm.caches = make(map[string]core.MCPCache)
	return nil
}

func (cm *realMCPCacheManager) Configure(config core.MCPCacheConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config = config
	return nil
}

func init() {
	// Register the memory cache manager via core factory hook when this plugin is imported.
	core.SetMCPCacheManagerFactory(func(cfg core.MCPCacheConfig) (core.MCPCacheManager, error) {
		return createRealMCPCacheManager(cfg)
	})
}

