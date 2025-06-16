// Package core provides public MCP interfaces for AgentFlow.
//
// This package exposes the minimal public API for Model Context Protocol (MCP) integration.
// All implementation details are kept in internal/mcp to maintain a clean public API.
package core

import (
	"context"
	"time"
)

// MCPManager provides the main interface for managing MCP connections and tools.
// It handles server discovery, connection management, and tool registration.
type MCPManager interface {
	// Connection Management
	Connect(ctx context.Context, serverName string) error
	Disconnect(serverName string) error
	DisconnectAll() error

	// Server Discovery and Management
	DiscoverServers(ctx context.Context) ([]MCPServerInfo, error)
	ListConnectedServers() []string
	GetServerInfo(serverName string) (*MCPServerInfo, error)

	// Tool Management
	RefreshTools(ctx context.Context) error
	GetAvailableTools() []MCPToolInfo
	GetToolsFromServer(serverName string) []MCPToolInfo

	// Health and Monitoring
	HealthCheck(ctx context.Context) map[string]MCPHealthStatus
	GetMetrics() MCPMetrics
}

// MCPAgent represents an agent that can utilize MCP tools.
// It extends the basic Agent interface with MCP-specific capabilities.
type MCPAgent interface {
	Agent

	// MCP-specific methods
	SelectTools(ctx context.Context, query string, context State) ([]string, error)
	ExecuteTools(ctx context.Context, tools []MCPToolExecution) ([]MCPToolResult, error)
	GetAvailableMCPTools() []MCPToolInfo
}

// MCPConfig holds configuration for MCP integration.
type MCPConfig struct {
	// Discovery settings
	EnableDiscovery  bool          `toml:"enable_discovery"`
	DiscoveryTimeout time.Duration `toml:"discovery_timeout"`
	ScanPorts        []int         `toml:"scan_ports"`

	// Connection settings
	ConnectionTimeout time.Duration `toml:"connection_timeout"`
	MaxRetries        int           `toml:"max_retries"`
	RetryDelay        time.Duration `toml:"retry_delay"`

	// Server configurations
	Servers []MCPServerConfig `toml:"servers"`

	// Performance settings
	EnableCaching  bool          `toml:"enable_caching"`
	CacheTimeout   time.Duration `toml:"cache_timeout"`
	MaxConnections int           `toml:"max_connections"`
}

// MCPServerConfig defines configuration for individual MCP servers.
type MCPServerConfig struct {
	Name    string `toml:"name"`
	Type    string `toml:"type"` // tcp, stdio, docker, websocket
	Host    string `toml:"host,omitempty"`
	Port    int    `toml:"port,omitempty"`
	Command string `toml:"command,omitempty"` // for stdio transport
	Enabled bool   `toml:"enabled"`
}

// MCPServerInfo represents information about an MCP server.
type MCPServerInfo struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Address      string                 `json:"address"`
	Port         int                    `json:"port"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Capabilities map[string]interface{} `json:"capabilities"`
	Status       string                 `json:"status"` // connected, disconnected, error
}

// MCPToolInfo represents metadata about an available MCP tool.
type MCPToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema"`
	ServerName  string                 `json:"server_name"`
}

// MCPToolExecution represents a tool execution request.
type MCPToolExecution struct {
	ToolName   string                 `json:"tool_name"`
	Arguments  map[string]interface{} `json:"arguments"`
	ServerName string                 `json:"server_name,omitempty"`
}

// MCPToolResult represents the result of an MCP tool execution.
type MCPToolResult struct {
	ToolName   string        `json:"tool_name"`
	ServerName string        `json:"server_name"`
	Success    bool          `json:"success"`
	Content    []MCPContent  `json:"content,omitempty"`
	Error      string        `json:"error,omitempty"`
	Duration   time.Duration `json:"duration"`
}

// MCPContent represents content returned by MCP tools.
type MCPContent struct {
	Type     string                 `json:"type"`
	Text     string                 `json:"text,omitempty"`
	Data     string                 `json:"data,omitempty"`
	MimeType string                 `json:"mime_type,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// MCPHealthStatus represents the health status of an MCP server connection.
type MCPHealthStatus struct {
	Status       string        `json:"status"` // healthy, unhealthy, unknown
	LastCheck    time.Time     `json:"last_check"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
	ToolCount    int           `json:"tool_count"`
}

// MCPMetrics provides metrics about MCP operations.
type MCPMetrics struct {
	ConnectedServers int                         `json:"connected_servers"`
	TotalTools       int                         `json:"total_tools"`
	ToolExecutions   int64                       `json:"tool_executions"`
	AverageLatency   time.Duration               `json:"average_latency"`
	ErrorRate        float64                     `json:"error_rate"`
	ServerMetrics    map[string]MCPServerMetrics `json:"server_metrics"`
}

// MCPServerMetrics provides metrics for individual servers.
type MCPServerMetrics struct {
	ToolCount        int           `json:"tool_count"`
	Executions       int64         `json:"executions"`
	SuccessfulCalls  int64         `json:"successful_calls"`
	FailedCalls      int64         `json:"failed_calls"`
	AverageLatency   time.Duration `json:"average_latency"`
	LastActivity     time.Time     `json:"last_activity"`
	ConnectionUptime time.Duration `json:"connection_uptime"`
}
