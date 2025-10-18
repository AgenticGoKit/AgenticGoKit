// Package vnext provides unified tools and MCP integration for the vNext API.
// This file consolidates MCP-related functionality into a streamlined interface.
//
// NOTE: Core types (ToolManager, ToolInfo, ToolResult, ToolsConfig, MCPConfig, MCPServer)
// are defined in agent.go and config.go. This file extends them with additional
// functionality, factory functions, and utilities for MCP integration.
package vnext

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// =============================================================================
// SECTION 1: EXTENDED TOOL AND MCP TYPES
// =============================================================================

// ToolExecutionRequest represents a tool execution request with additional options
type ToolExecutionRequest struct {
	ToolName   string                 `json:"tool_name"`
	Arguments  map[string]interface{} `json:"arguments"`
	ServerName string                 `json:"server_name,omitempty"`
	Timeout    time.Duration          `json:"timeout,omitempty"`
	RetryCount int                    `json:"retry_count,omitempty"`
	CacheKey   *CacheKey              `json:"cache_key,omitempty"`
}

// ToolMetrics provides metrics about tool operations
type ToolMetrics struct {
	TotalExecutions  int64                          `json:"total_executions"`
	SuccessfulCalls  int64                          `json:"successful_calls"`
	FailedCalls      int64                          `json:"failed_calls"`
	AverageLatency   time.Duration                  `json:"average_latency"`
	CacheHitRate     float64                        `json:"cache_hit_rate"`
	ToolMetrics      map[string]ToolSpecificMetrics `json:"tool_metrics"`
	MCPServerMetrics map[string]MCPServerMetrics    `json:"mcp_server_metrics,omitempty"`
}

// ToolSpecificMetrics provides metrics for individual tools
type ToolSpecificMetrics struct {
	Executions     int64         `json:"executions"`
	SuccessRate    float64       `json:"success_rate"`
	AverageLatency time.Duration `json:"average_latency"`
	LastUsed       time.Time     `json:"last_used"`
}

// MCPServerInfo represents information about an MCP server
type MCPServerInfo struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Address      string                 `json:"address"`
	Port         int                    `json:"port,omitempty"`
	Version      string                 `json:"version,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Capabilities map[string]interface{} `json:"capabilities,omitempty"`
	Status       string                 `json:"status"` // connected, disconnected, error
	ToolCount    int                    `json:"tool_count"`
}

// MCPServerMetrics provides metrics for individual MCP servers
type MCPServerMetrics struct {
	ToolCount        int           `json:"tool_count"`
	Executions       int64         `json:"executions"`
	SuccessfulCalls  int64         `json:"successful_calls"`
	FailedCalls      int64         `json:"failed_calls"`
	AverageLatency   time.Duration `json:"average_latency"`
	LastActivity     time.Time     `json:"last_activity"`
	ConnectionUptime time.Duration `json:"connection_uptime"`
}

// MCPHealthStatus represents the health status of an MCP server connection
type MCPHealthStatus struct {
	Status       string        `json:"status"` // healthy, unhealthy, unknown
	LastCheck    time.Time     `json:"last_check"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
	ToolCount    int           `json:"tool_count"`
}

// MCPContent represents content returned by MCP tools
type MCPContent struct {
	Type     string                 `json:"type"`
	Text     string                 `json:"text,omitempty"`
	Data     string                 `json:"data,omitempty"`
	MimeType string                 `json:"mime_type,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// =============================================================================
// SECTION 2: CACHING SYSTEM TYPES
// =============================================================================

// ToolCache defines the interface for caching tool results
type ToolCache interface {
	// Get retrieves a cached result by key
	Get(ctx context.Context, key CacheKey) (*CachedResult, error)

	// Set stores a result in the cache
	Set(ctx context.Context, key CacheKey, result *ToolResult, ttl time.Duration) error

	// Delete removes a specific key from the cache
	Delete(ctx context.Context, key CacheKey) error

	// Clear removes all entries from the cache
	Clear(ctx context.Context) error

	// Exists checks if a key exists in the cache
	Exists(ctx context.Context, key CacheKey) (bool, error)

	// Stats returns cache performance statistics
	Stats(ctx context.Context) (*CacheStats, error)

	// Cleanup performs maintenance operations (e.g., TTL expiration)
	Cleanup(ctx context.Context) error
}

// CacheKey represents a unique identifier for cached tool results
type CacheKey struct {
	ToolName   string            `json:"tool_name"`
	ServerName string            `json:"server_name,omitempty"`
	Args       map[string]string `json:"args"`
	Hash       string            `json:"hash"` // SHA256 hash of normalized args
}

// CachedResult represents a cached tool execution result
type CachedResult struct {
	Key         CacheKey               `json:"key"`
	Result      *ToolResult            `json:"result"`
	Timestamp   time.Time              `json:"timestamp"`
	TTL         time.Duration          `json:"ttl"`
	AccessCount int                    `json:"access_count"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CacheStats provides statistics about cache performance
type CacheStats struct {
	TotalKeys      int           `json:"total_keys"`
	HitCount       int64         `json:"hit_count"`
	MissCount      int64         `json:"miss_count"`
	HitRate        float64       `json:"hit_rate"`
	EvictionCount  int64         `json:"eviction_count"`
	TotalSize      int64         `json:"total_size_bytes"`
	AverageLatency time.Duration `json:"average_latency"`
	LastCleanup    time.Time     `json:"last_cleanup"`
}

// =============================================================================
// SECTION 3: FACTORY FUNCTIONS
// =============================================================================

// NewToolManager creates a new ToolManager with the given configuration
func NewToolManager(config *ToolsConfig) (ToolManager, error) {
	if config == nil {
		return nil, fmt.Errorf("tools config cannot be nil")
	}

	// Use factory if registered (for plugin-based implementations)
	factoryMutex.RLock()
	factory := toolManagerFactory
	factoryMutex.RUnlock()

	if factory != nil {
		return factory(config)
	}

	// Return basic implementation
	return &basicToolManager{
		config:  config,
		tools:   make(map[string]ToolInfo),
		metrics: &ToolMetrics{ToolMetrics: make(map[string]ToolSpecificMetrics)},
		mu:      &sync.RWMutex{},
	}, nil
}

// NewToolManagerWithMCP creates a ToolManager with MCP integration enabled
func NewToolManagerWithMCP(config *ToolsConfig, mcpServers ...MCPServer) (ToolManager, error) {
	if config == nil {
		config = &ToolsConfig{Enabled: true}
	}

	if config.MCP == nil {
		config.MCP = &MCPConfig{Enabled: true}
	}

	config.MCP.Enabled = true
	config.MCP.Servers = append(config.MCP.Servers, mcpServers...)

	return NewToolManager(config)
}

// NewToolManagerWithCache creates a ToolManager with caching enabled
func NewToolManagerWithCache(config *ToolsConfig, cacheConfig *CacheConfig) (ToolManager, error) {
	if config == nil {
		config = &ToolsConfig{Enabled: true}
	}

	config.Cache = cacheConfig

	return NewToolManager(config)
}

// =============================================================================
// SECTION 4: DEFAULT CONFIGURATIONS
// =============================================================================

// DefaultToolsConfig returns a default tools configuration
func DefaultToolsConfig() *ToolsConfig {
	return &ToolsConfig{
		Enabled:       true,
		MaxRetries:    3,
		Timeout:       30 * time.Second,
		RateLimit:     100,
		MaxConcurrent: 10,
		Cache: &CacheConfig{
			Enabled:         true,
			TTL:             15 * time.Minute,
			MaxSize:         100, // 100 MB
			MaxKeys:         10000,
			EvictionPolicy:  "lru",
			CleanupInterval: 5 * time.Minute,
			Backend:         "memory",
		},
		CircuitBreaker: &CircuitBreakerConfig{
			Enabled:          true,
			FailureThreshold: 5,
			SuccessThreshold: 2,
			Timeout:          60 * time.Second,
			HalfOpenMaxCalls: 3,
		},
	}
}

// DefaultMCPConfig returns a default MCP configuration
func DefaultMCPConfig() *MCPConfig {
	return &MCPConfig{
		Enabled:           true,
		Discovery:         true,
		ConnectionTimeout: 30 * time.Second,
		MaxRetries:        3,
		RetryDelay:        1 * time.Second,
		DiscoveryTimeout:  10 * time.Second,
		ScanPorts:         []int{8080, 8081, 8090, 8100, 3000, 3001},
		Servers:           []MCPServer{},
		Cache: &CacheConfig{
			Enabled:         true,
			TTL:             15 * time.Minute,
			MaxSize:         100,
			MaxKeys:         10000,
			EvictionPolicy:  "lru",
			CleanupInterval: 5 * time.Minute,
			Backend:         "memory",
			ToolTTLs: map[string]time.Duration{
				"web_search":         5 * time.Minute,
				"content_fetch":      30 * time.Minute,
				"summarize_text":     60 * time.Minute,
				"sentiment_analysis": 45 * time.Minute,
			},
		},
	}
}

// DefaultCacheConfig returns a default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		Enabled:         true,
		TTL:             15 * time.Minute,
		MaxSize:         100, // 100 MB
		MaxKeys:         10000,
		EvictionPolicy:  "lru",
		CleanupInterval: 5 * time.Minute,
		Backend:         "memory",
		ToolTTLs:        make(map[string]time.Duration),
		BackendConfig:   make(map[string]string),
	}
}

// =============================================================================
// SECTION 5: CACHE UTILITIES
// =============================================================================

// GenerateCacheKey creates a standardized cache key for tool execution
func GenerateCacheKey(toolName, serverName string, args map[string]interface{}) CacheKey {
	// Convert interface{} args to string args for hashing
	strArgs := make(map[string]string)
	for k, v := range args {
		strArgs[k] = fmt.Sprintf("%v", v)
	}

	return CacheKey{
		ToolName:   toolName,
		ServerName: serverName,
		Args:       normalizeArgs(strArgs),
		Hash:       generateArgHash(strArgs),
	}
}

// normalizeArgs ensures consistent argument formatting for cache keys
func normalizeArgs(args map[string]string) map[string]string {
	normalized := make(map[string]string)
	for k, v := range args {
		// Normalize whitespace and case for cache consistency
		normalized[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
	}
	return normalized
}

// generateArgHash creates a deterministic hash of the arguments
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

// =============================================================================
// SECTION 6: MCP UTILITY FUNCTIONS FOR LLM INTEGRATION
// =============================================================================

// FormatToolsPromptForLLM creates a prompt section describing available tools
func FormatToolsPromptForLLM(tools []ToolInfo) string {
	if len(tools) == 0 {
		return ""
	}

	prompt := "\n\nAvailable tools:\n"
	for _, tool := range tools {
		prompt += fmt.Sprintf("\n**%s**: %s\n", tool.Name, tool.Description)

		// Include parameters information if available
		if len(tool.Parameters) > 0 {
			prompt += "Parameters: " + formatParametersForLLM(tool.Parameters) + "\n"
		}
	}

	prompt += `
To use a tool, respond with a tool call in this exact JSON format:
TOOL_CALL{"name": "tool_name", "args": {arguments}}

Example:
TOOL_CALL{"name": "search", "args": {"query": "search terms", "max_results": 10}}

Use these tools to provide comprehensive and accurate responses.`

	return prompt
}

// FormatSchemaForLLM converts a tool schema to a readable string format
func FormatSchemaForLLM(schema map[string]interface{}) string {
	if schema == nil {
		return "No schema available"
	}

	var result strings.Builder

	// Handle the "type" field
	if schemaType, ok := schema["type"].(string); ok {
		result.WriteString(fmt.Sprintf("Type: %s", schemaType))
	}

	// Handle "properties" field (for object types)
	if properties, ok := schema["properties"].(map[string]interface{}); ok {
		result.WriteString("\nParameters:\n")
		for propName, propDetails := range properties {
			if propMap, ok := propDetails.(map[string]interface{}); ok {
				propType := "unknown"
				if t, exists := propMap["type"]; exists {
					if typeStr, ok := t.(string); ok {
						propType = typeStr
					}
				}

				description := ""
				if desc, exists := propMap["description"]; exists {
					if descStr, ok := desc.(string); ok {
						description = fmt.Sprintf(" - %s", descStr)
					}
				}

				result.WriteString(fmt.Sprintf("  - %s (%s)%s\n", propName, propType, description))
			}
		}
	}

	// Handle "required" field
	if required, ok := schema["required"].([]interface{}); ok {
		if len(required) > 0 {
			result.WriteString("Required parameters: ")
			for i, req := range required {
				if reqStr, ok := req.(string); ok {
					if i > 0 {
						result.WriteString(", ")
					}
					result.WriteString(reqStr)
				}
			}
			result.WriteString("\n")
		}
	}

	return result.String()
}

// formatParametersForLLM converts parameters to a readable format
func formatParametersForLLM(params map[string]interface{}) string {
	if params == nil || len(params) == 0 {
		return "No parameters"
	}

	var result strings.Builder
	result.WriteString("\n")

	for name, details := range params {
		result.WriteString(fmt.Sprintf("  - %s: %v\n", name, details))
	}

	return result.String()
}

// ParseLLMToolCalls extracts tool calls from LLM response content
func ParseLLMToolCalls(content string) []map[string]interface{} {
	var toolCalls []map[string]interface{}

	// Parse TOOL_CALL{...} patterns from LLM response
	parts := strings.Split(content, "TOOL_CALL")
	for i := 1; i < len(parts); i++ {
		part := parts[i]

		if strings.HasPrefix(part, "{") {
			// Find the closing brace
			braceCount := 0
			endIndex := -1
			for j, char := range part {
				if char == '{' {
					braceCount++
				} else if char == '}' {
					braceCount--
					if braceCount == 0 {
						endIndex = j
						break
					}
				}
			}

			if endIndex > 0 {
				jsonStr := part[:endIndex+1]
				toolCall := ParseToolCallJSON(jsonStr)
				if len(toolCall) > 0 {
					toolCalls = append(toolCalls, toolCall)
				}
			}
		}
	}

	return toolCalls
}

// ParseToolCallJSON parses a tool call JSON string
func ParseToolCallJSON(jsonStr string) map[string]interface{} {
	result := make(map[string]interface{})

	// Try to parse as proper JSON first
	if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
		return result
	}

	// Fall back to simple parsing if JSON unmarshal fails
	jsonStr = strings.Trim(jsonStr, "{}")
	parts := strings.Split(jsonStr, ",")

	for _, part := range parts {
		if strings.Contains(part, ":") {
			keyValue := strings.SplitN(part, ":", 2)
			if len(keyValue) == 2 {
				key := strings.Trim(keyValue[0], " \"")
				value := strings.Trim(keyValue[1], " \"")

				// Try to parse nested objects for args
				if key == "args" && strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
					argsMap := ParseToolCallJSON(value)
					result[key] = argsMap
				} else {
					result[key] = value
				}
			}
		}
	}

	return result
}

// =============================================================================
// SECTION 7: PLUGIN INTEGRATION
// =============================================================================

// ToolManagerFactory is a function type for creating ToolManager instances
type ToolManagerFactory func(config *ToolsConfig) (ToolManager, error)

// Global factory variable for plugin-based implementations
var (
	toolManagerFactory ToolManagerFactory
	factoryMutex       sync.RWMutex
)

// SetToolManagerFactory allows plugins to register a custom ToolManager factory
func SetToolManagerFactory(factory ToolManagerFactory) {
	factoryMutex.Lock()
	defer factoryMutex.Unlock()
	toolManagerFactory = factory
}

// GetToolManagerFactory returns the registered ToolManager factory
func GetToolManagerFactory() ToolManagerFactory {
	factoryMutex.RLock()
	defer factoryMutex.RUnlock()
	return toolManagerFactory
}

// =============================================================================
// SECTION 8: BASIC IMPLEMENTATION (FALLBACK)
// =============================================================================

// basicToolManager provides a minimal implementation when no plugin is available
type basicToolManager struct {
	config  *ToolsConfig
	tools   map[string]ToolInfo
	metrics *ToolMetrics
	mu      *sync.RWMutex
}

// Execute implements ToolManager.Execute
func (tm *basicToolManager) Execute(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	return &ToolResult{
		Success: false,
		Error:   "no tool plugin registered - import a tool plugin package",
	}, fmt.Errorf("tool execution not available without plugin")
}

// List implements ToolManager.List
func (tm *basicToolManager) List() []ToolInfo {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tools := make([]ToolInfo, 0, len(tm.tools))
	for _, tool := range tm.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Available implements ToolManager.Available
func (tm *basicToolManager) Available() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	names := make([]string, 0, len(tm.tools))
	for name := range tm.tools {
		names = append(names, name)
	}
	return names
}

// IsAvailable implements ToolManager.IsAvailable
func (tm *basicToolManager) IsAvailable(name string) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	_, exists := tm.tools[name]
	return exists
}

// GetMetrics implements ToolManager.GetMetrics
func (tm *basicToolManager) GetMetrics() ToolMetrics {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tm.metrics == nil {
		return ToolMetrics{ToolMetrics: make(map[string]ToolSpecificMetrics)}
	}
	return *tm.metrics
}

// Initialize implements ToolManager.Initialize
func (tm *basicToolManager) Initialize(ctx context.Context) error {
	return nil
}

// Shutdown implements ToolManager.Shutdown
func (tm *basicToolManager) Shutdown(ctx context.Context) error {
	return nil
}

// ConnectMCP implements ToolManager.ConnectMCP
func (tm *basicToolManager) ConnectMCP(ctx context.Context, servers ...MCPServer) error {
	return fmt.Errorf("MCP not available without plugin")
}

// DisconnectMCP implements ToolManager.DisconnectMCP
func (tm *basicToolManager) DisconnectMCP(serverName string) error {
	return fmt.Errorf("MCP not available without plugin")
}

// DiscoverMCP implements ToolManager.DiscoverMCP
func (tm *basicToolManager) DiscoverMCP(ctx context.Context) ([]MCPServerInfo, error) {
	return nil, fmt.Errorf("MCP discovery not available without plugin")
}

// HealthCheck implements ToolManager.HealthCheck
func (tm *basicToolManager) HealthCheck(ctx context.Context) map[string]MCPHealthStatus {
	return make(map[string]MCPHealthStatus)
}

// =============================================================================
// SECTION 9: CONFIGURATION VALIDATION
// =============================================================================

// ValidateToolsConfig validates the ToolsConfig
func ValidateToolsConfig(c *ToolsConfig) error {
	if c.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}

	if c.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	if c.RateLimit < 0 {
		return fmt.Errorf("rate_limit cannot be negative")
	}

	if c.MaxConcurrent < 0 {
		return fmt.Errorf("max_concurrent cannot be negative")
	}

	// Validate MCP config if present
	if c.MCP != nil {
		if err := ValidateMCPConfig(c.MCP); err != nil {
			return fmt.Errorf("MCP config validation failed: %w", err)
		}
	}

	// Validate cache config if present
	if c.Cache != nil {
		if err := ValidateCacheConfig(c.Cache); err != nil {
			return fmt.Errorf("cache config validation failed: %w", err)
		}
	}

	// Validate circuit breaker config if present
	if c.CircuitBreaker != nil {
		if err := ValidateCircuitBreakerConfig(c.CircuitBreaker); err != nil {
			return fmt.Errorf("circuit breaker config validation failed: %w", err)
		}
	}

	return nil
}

// ValidateMCPConfig validates the MCPConfig
func ValidateMCPConfig(c *MCPConfig) error {
	if c.ConnectionTimeout < 0 {
		return fmt.Errorf("connection_timeout cannot be negative")
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}

	if c.RetryDelay < 0 {
		return fmt.Errorf("retry_delay cannot be negative")
	}

	// Validate servers
	serverNames := make(map[string]bool)
	for i, server := range c.Servers {
		if err := ValidateMCPServer(&server); err != nil {
			return fmt.Errorf("server %d validation failed: %w", i, err)
		}

		// Check for duplicate server names
		if serverNames[server.Name] {
			return fmt.Errorf("duplicate server name: %s", server.Name)
		}
		serverNames[server.Name] = true
	}

	return nil
}

// ValidateMCPServer validates the MCPServer configuration
func ValidateMCPServer(s *MCPServer) error {
	if s.Name == "" {
		return fmt.Errorf("server name cannot be empty")
	}

	switch s.Type {
	case "tcp":
		if s.Address == "" {
			return fmt.Errorf("TCP server requires address")
		}
		if s.Port <= 0 || s.Port > 65535 {
			return fmt.Errorf("TCP server requires valid port (1-65535)")
		}
	case "websocket":
		if s.Address == "" {
			return fmt.Errorf("WebSocket server requires address")
		}
	case "stdio":
		if s.Command == "" {
			return fmt.Errorf("STDIO server requires command")
		}
	case "http_sse", "http_streaming":
		if s.Address == "" {
			return fmt.Errorf("%s server requires address", s.Type)
		}
	default:
		return fmt.Errorf("unsupported server type: %s", s.Type)
	}

	return nil
}

// ValidateCacheConfig validates the CacheConfig
func ValidateCacheConfig(c *CacheConfig) error {
	if c.TTL < 0 {
		return fmt.Errorf("ttl cannot be negative")
	}

	if c.MaxSize < 0 {
		return fmt.Errorf("max_size_mb cannot be negative")
	}

	if c.MaxKeys < 0 {
		return fmt.Errorf("max_keys cannot be negative")
	}

	if c.CleanupInterval < 0 {
		return fmt.Errorf("cleanup_interval cannot be negative")
	}

	validPolicies := map[string]bool{"lru": true, "lfu": true, "ttl": true}
	if c.EvictionPolicy != "" && !validPolicies[c.EvictionPolicy] {
		return fmt.Errorf("invalid eviction_policy: %s (must be lru, lfu, or ttl)", c.EvictionPolicy)
	}

	validBackends := map[string]bool{"memory": true, "redis": true, "file": true}
	if c.Backend != "" && !validBackends[c.Backend] {
		return fmt.Errorf("invalid backend: %s (must be memory, redis, or file)", c.Backend)
	}

	return nil
}

// ValidateCircuitBreakerConfig validates the CircuitBreakerConfig
func ValidateCircuitBreakerConfig(c *CircuitBreakerConfig) error {
	if c.FailureThreshold < 0 {
		return fmt.Errorf("failure_threshold cannot be negative")
	}

	if c.SuccessThreshold < 0 {
		return fmt.Errorf("success_threshold cannot be negative")
	}

	if c.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	if c.HalfOpenMaxCalls < 0 {
		return fmt.Errorf("half_open_max_calls cannot be negative")
	}

	return nil
}
