// Package core provides configuration and helper functions for MCP integration.
package core

import (
	"fmt"
	"time"
)

// DefaultMCPConfig returns a default MCP configuration suitable for development.
//
// This configuration enables discovery, sets reasonable timeouts, and includes
// common MCP server ports for scanning.
func DefaultMCPConfig() MCPConfig {
	return MCPConfig{
		EnableDiscovery:   true,
		DiscoveryTimeout:  10 * time.Second,
		ScanPorts:         []int{8080, 8081, 8090, 8100, 3000, 3001},
		ConnectionTimeout: 30 * time.Second,
		MaxRetries:        3,
		RetryDelay:        1 * time.Second,
		EnableCaching:     true,
		CacheTimeout:      5 * time.Minute,
		MaxConnections:    10,
		Servers:           []MCPServerConfig{}, // Empty by default
	}
}

// NewMCPServerConfig creates a new server configuration with validation.
func NewMCPServerConfig(name, serverType, host string, port int) (MCPServerConfig, error) {
	config := MCPServerConfig{
		Name:    name,
		Type:    serverType,
		Host:    host,
		Port:    port,
		Enabled: true,
	}

	// Basic validation
	if name == "" {
		return config, fmt.Errorf("server name cannot be empty")
	}

	switch serverType {
	case "tcp", "websocket":
		if host == "" {
			return config, fmt.Errorf("%s server must specify host", serverType)
		}
		if port <= 0 || port > 65535 {
			return config, fmt.Errorf("%s server must specify valid port (1-65535)", serverType)
		}
	case "stdio":
		// For STDIO, we use the host field as the command
		if host == "" {
			return config, fmt.Errorf("stdio server must specify command")
		}
		config.Command = host
		config.Host = ""
		config.Port = 0
	case "docker":
		// Docker configuration validation could be added here
	default:
		return config, fmt.Errorf("unsupported server type: %s", serverType)
	}

	return config, nil
}

// NewTCPServerConfig creates a TCP server configuration.
func NewTCPServerConfig(name, host string, port int) (MCPServerConfig, error) {
	return NewMCPServerConfig(name, "tcp", host, port)
}

// NewSTDIOServerConfig creates a STDIO server configuration.
func NewSTDIOServerConfig(name, command string) (MCPServerConfig, error) {
	config := MCPServerConfig{
		Name:    name,
		Type:    "stdio",
		Command: command,
		Enabled: true,
	}

	if name == "" {
		return config, fmt.Errorf("server name cannot be empty")
	}
	if command == "" {
		return config, fmt.Errorf("command cannot be empty")
	}

	return config, nil
}

// NewWebSocketServerConfig creates a WebSocket server configuration.
func NewWebSocketServerConfig(name, host string, port int) (MCPServerConfig, error) {
	return NewMCPServerConfig(name, "websocket", host, port)
}
