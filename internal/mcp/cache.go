// Package mcp provides internal MCP cache implementations.
package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
)

// MemoryCache implements MCPCache using in-memory storage with LRU eviction.
type MemoryCache struct {
	mu          sync.RWMutex
	data        map[string]*cacheEntry
	accessOrder []string // LRU tracking
	config      core.MCPCacheConfig
	stats       cacheStats
	stopCleanup chan struct{}
	cleanupDone chan struct{}
}

// cacheEntry represents an internal cache entry with metadata.
type cacheEntry struct {
	result      core.MCPCachedResult
	lastAccess  time.Time
	accessCount int64
}

// cacheStats tracks cache performance metrics.
type cacheStats struct {
	hits        int64
	misses      int64
	evictions   int64
	totalSize   int64
	lastCleanup time.Time
}

// NewMemoryCache creates a new in-memory cache instance.
func NewMemoryCache(config core.MCPCacheConfig) (*MemoryCache, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("cache is disabled")
	}

	cache := &MemoryCache{
		data:        make(map[string]*cacheEntry),
		accessOrder: make([]string, 0),
		config:      config,
		stats: cacheStats{
			lastCleanup: time.Now(),
		},
		stopCleanup: make(chan struct{}),
		cleanupDone: make(chan struct{}),
	}

	// Start background cleanup routine
	go cache.cleanupRoutine()

	return cache, nil
}

// Get retrieves a cached result by key.
func (c *MemoryCache) Get(ctx context.Context, key core.MCPCacheKey) (*core.MCPCachedResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	keyStr := c.keyToString(key)
	entry, exists := c.data[keyStr]

	if !exists {
		c.stats.misses++
		return nil, nil // Cache miss
	}

	// Check TTL expiration
	if time.Since(entry.result.Timestamp) > entry.result.TTL {
		delete(c.data, keyStr)
		c.removeFromAccessOrder(keyStr)
		c.stats.misses++
		return nil, nil // Expired
	}

	// Update access tracking
	entry.lastAccess = time.Now()
	entry.accessCount++
	entry.result.AccessCount = int(entry.accessCount)
	c.updateAccessOrder(keyStr)
	c.stats.hits++

	// Return a copy to prevent external modification
	result := entry.result
	return &result, nil
}

// Set stores a result in the cache.
func (c *MemoryCache) Set(ctx context.Context, key core.MCPCacheKey, result core.MCPToolResult, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	keyStr := c.keyToString(key)

	// Create cached result
	cachedResult := core.MCPCachedResult{
		Key:         key,
		Result:      result,
		Timestamp:   time.Now(),
		TTL:         ttl,
		AccessCount: 1,
		Metadata:    make(map[string]interface{}),
	}

	// Check if we need to evict entries
	if len(c.data) >= c.config.MaxKeys {
		c.evictLRU()
	}

	// Store the entry
	entry := &cacheEntry{
		result:      cachedResult,
		lastAccess:  time.Now(),
		accessCount: 1,
	}

	c.data[keyStr] = entry
	c.addToAccessOrder(keyStr)
	c.updateSize(keyStr, entry)

	return nil
}

// Delete removes a specific key from the cache.
func (c *MemoryCache) Delete(ctx context.Context, key core.MCPCacheKey) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	keyStr := c.keyToString(key)
	if _, exists := c.data[keyStr]; exists {
		delete(c.data, keyStr)
		c.removeFromAccessOrder(keyStr)
	}

	return nil
}

// Clear removes all entries from the cache.
func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*cacheEntry)
	c.accessOrder = make([]string, 0)
	c.stats.totalSize = 0

	return nil
}

// Exists checks if a key exists in the cache.
func (c *MemoryCache) Exists(ctx context.Context, key core.MCPCacheKey) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keyStr := c.keyToString(key)
	entry, exists := c.data[keyStr]

	if !exists {
		return false, nil
	}

	// Check TTL expiration
	if time.Since(entry.result.Timestamp) > entry.result.TTL {
		return false, nil
	}

	return true, nil
}

// Stats returns cache performance statistics.
func (c *MemoryCache) Stats(ctx context.Context) (core.MCPCacheStats, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.stats.hits + c.stats.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.stats.hits) / float64(total)
	}

	return core.MCPCacheStats{
		TotalKeys:      len(c.data),
		HitCount:       c.stats.hits,
		MissCount:      c.stats.misses,
		HitRate:        hitRate,
		EvictionCount:  c.stats.evictions,
		TotalSize:      c.stats.totalSize,
		AverageLatency: time.Millisecond, // Placeholder - would measure actual latency
		LastCleanup:    c.stats.lastCleanup,
	}, nil
}

// Cleanup performs maintenance operations.
func (c *MemoryCache) Cleanup(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cleanupExpired()
	c.stats.lastCleanup = time.Now()
	return nil
}

// Close closes the cache and releases resources.
func (c *MemoryCache) Close() error {
	close(c.stopCleanup)
	<-c.cleanupDone
	return nil
}

// Helper methods

func (c *MemoryCache) keyToString(key core.MCPCacheKey) string {
	return fmt.Sprintf("%s:%s:%s", key.ServerName, key.ToolName, key.Hash)
}

func (c *MemoryCache) updateAccessOrder(keyStr string) {
	// Remove from current position
	c.removeFromAccessOrder(keyStr)
	// Add to end (most recently used)
	c.accessOrder = append(c.accessOrder, keyStr)
}

func (c *MemoryCache) addToAccessOrder(keyStr string) {
	c.accessOrder = append(c.accessOrder, keyStr)
}

func (c *MemoryCache) removeFromAccessOrder(keyStr string) {
	for i, key := range c.accessOrder {
		if key == keyStr {
			c.accessOrder = append(c.accessOrder[:i], c.accessOrder[i+1:]...)
			break
		}
	}
}

func (c *MemoryCache) evictLRU() {
	if len(c.accessOrder) == 0 {
		return
	}

	// Remove least recently used (first in order)
	lruKey := c.accessOrder[0]
	delete(c.data, lruKey)
	c.accessOrder = c.accessOrder[1:]
	c.stats.evictions++
}

func (c *MemoryCache) updateSize(keyStr string, entry *cacheEntry) {
	// Rough size estimation (could be more sophisticated)
	contentSize := 0
	for _, content := range entry.result.Result.Content {
		contentSize += len(content.Text) + len(content.Data)
	}
	size := int64(len(keyStr) + contentSize + 200) // 200 bytes overhead
	c.stats.totalSize += size
}

func (c *MemoryCache) cleanupExpired() {
	now := time.Now()
	expiredKeys := make([]string, 0)

	for keyStr, entry := range c.data {
		if now.Sub(entry.result.Timestamp) > entry.result.TTL {
			expiredKeys = append(expiredKeys, keyStr)
		}
	}

	for _, keyStr := range expiredKeys {
		delete(c.data, keyStr)
		c.removeFromAccessOrder(keyStr)
	}
}

func (c *MemoryCache) cleanupRoutine() {
	defer close(c.cleanupDone)

	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.Cleanup(context.Background())
		case <-c.stopCleanup:
			return
		}
	}
}

