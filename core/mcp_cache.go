// Package core provides public interfaces for MCP tool result caching.
package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"time"
)

// MCPCacheKey represents a unique identifier for cached tool results.
type MCPCacheKey struct {
	ToolName   string            `json:"tool_name"`
	ServerName string            `json:"server_name"`
	Args       map[string]string `json:"args"`
	Hash       string            `json:"hash"` // SHA256 hash of normalized args
}

// MCPCachedResult represents a cached tool execution result.
type MCPCachedResult struct {
	Key         MCPCacheKey            `json:"key"`
	Result      MCPToolResult          `json:"result"`
	Timestamp   time.Time              `json:"timestamp"`
	TTL         time.Duration          `json:"ttl"`
	AccessCount int                    `json:"access_count"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MCPCacheStats provides statistics about cache performance.
type MCPCacheStats struct {
	TotalKeys      int           `json:"total_keys"`
	HitCount       int64         `json:"hit_count"`
	MissCount      int64         `json:"miss_count"`
	HitRate        float64       `json:"hit_rate"`
	EvictionCount  int64         `json:"eviction_count"`
	TotalSize      int64         `json:"total_size_bytes"`
	AverageLatency time.Duration `json:"average_latency"`
	LastCleanup    time.Time     `json:"last_cleanup"`
}

// MCPCacheConfig holds configuration for the cache system.
type MCPCacheConfig struct {
	// Cache behavior
	Enabled    bool          `toml:"enabled"`
	DefaultTTL time.Duration `toml:"default_ttl"`
	MaxSize    int64         `toml:"max_size_mb"`
	MaxKeys    int           `toml:"max_keys"`

	// Eviction policy
	EvictionPolicy  string        `toml:"eviction_policy"` // "lru", "lfu", "ttl"
	CleanupInterval time.Duration `toml:"cleanup_interval"`

	// Per-tool TTL overrides
	ToolTTLs map[string]time.Duration `toml:"tool_ttls"`

	// Backend configuration
	Backend       string            `toml:"backend"` // "memory", "redis", "file"
	BackendConfig map[string]string `toml:"backend_config"`
}

// MCPCache defines the interface for caching MCP tool results.
type MCPCache interface {
	// Get retrieves a cached result by key
	Get(ctx context.Context, key MCPCacheKey) (*MCPCachedResult, error)

	// Set stores a result in the cache
	Set(ctx context.Context, key MCPCacheKey, result MCPToolResult, ttl time.Duration) error

	// Delete removes a specific key from the cache
	Delete(ctx context.Context, key MCPCacheKey) error

	// Clear removes all entries from the cache
	Clear(ctx context.Context) error

	// Exists checks if a key exists in the cache
	Exists(ctx context.Context, key MCPCacheKey) (bool, error)

	// Stats returns cache performance statistics
	Stats(ctx context.Context) (MCPCacheStats, error)

	// Cleanup performs maintenance operations (e.g., TTL expiration)
	Cleanup(ctx context.Context) error

	// Close closes the cache and releases resources
	Close() error
}

// MCPCacheManager manages multiple cache instances and provides cache-aware tool execution.
type MCPCacheManager interface {
	// GetCache returns a cache instance for a specific tool or server
	GetCache(toolName, serverName string) MCPCache

	// ExecuteWithCache executes a tool with caching support
	ExecuteWithCache(ctx context.Context, execution MCPToolExecution) (MCPToolResult, error)

	// InvalidateByPattern invalidates cache entries matching a pattern
	InvalidateByPattern(ctx context.Context, pattern string) error

	// GetGlobalStats returns aggregated cache statistics
	GetGlobalStats(ctx context.Context) (MCPCacheStats, error)

	// Configure updates cache configuration
	Configure(config MCPCacheConfig) error
}

// DefaultMCPCacheConfig returns a default cache configuration.
func DefaultMCPCacheConfig() MCPCacheConfig {
	return MCPCacheConfig{
		Enabled:         true,
		DefaultTTL:      15 * time.Minute,
		MaxSize:         100, // 100 MB
		MaxKeys:         10000,
		EvictionPolicy:  "lru",
		CleanupInterval: 5 * time.Minute,
		Backend:         "memory",
		ToolTTLs: map[string]time.Duration{
			"web_search":         5 * time.Minute,  // Search results change frequently
			"content_fetch":      30 * time.Minute, // Content is more stable
			"summarize_text":     60 * time.Minute, // Summaries are expensive to compute
			"sentiment_analysis": 45 * time.Minute, // Analysis results are stable
			"compute_metric":     20 * time.Minute, // Metrics may change
			"entity_extraction":  60 * time.Minute, // Entity extraction is expensive
		},
		BackendConfig: map[string]string{
			"redis_addr":     "localhost:6379",
			"redis_password": "",
			"redis_db":       "0",
			"file_path":      "./cache",
		},
	}
}

// GenerateCacheKey creates a standardized cache key for tool execution.
func GenerateCacheKey(toolName, serverName string, args map[string]string) MCPCacheKey {
	return MCPCacheKey{
		ToolName:   toolName,
		ServerName: serverName,
		Args:       normalizeArgs(args),
		Hash:       generateArgHash(args),
	}
}

// normalizeArgs ensures consistent argument formatting for cache keys.
func normalizeArgs(args map[string]string) map[string]string {
	normalized := make(map[string]string)
	for k, v := range args {
		// Normalize whitespace and case for cache consistency
		normalized[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
	}
	return normalized
}

// generateArgHash creates a deterministic hash of the arguments.
func generateArgHash(args map[string]string) string {
	// Sort keys for deterministic hashing
	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		h.Write([]byte(k + "=" + args[k] + "|"))
	}
	return hex.EncodeToString(h.Sum(nil))[:16] // Use first 16 chars for brevity
}
