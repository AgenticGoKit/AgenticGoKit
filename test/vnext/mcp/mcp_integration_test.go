package mcp_test

import (
	"context"
	"testing"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"

	// Import MCP plugin
	_ "github.com/agenticgokit/agenticgokit/plugins/mcp/default"

	// Import LLM provider
	_ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
)

// TestMCPConfigurationSetup tests that MCP configuration is properly set up
func TestMCPConfigurationSetup(t *testing.T) {
	t.Run("WithMCP sets up server configuration", func(t *testing.T) {
		server := vnext.MCPServer{
			Name:    "test-server",
			Type:    "tcp",
			Address: "localhost",
			Port:    8811,
			Enabled: true,
		}

		agent, err := vnext.NewBuilder("test-agent").
			WithConfig(&vnext.Config{
				Name:         "test-agent",
				SystemPrompt: "Test agent",
				Timeout:      30 * time.Second,
				LLM: vnext.LLMConfig{
					Provider:    "ollama",
					Model:       "gemma3:1b",
					Temperature: 0.7,
					MaxTokens:   100,
				},
			}).
			WithTools(vnext.WithMCP(server)).
			Build()

		if err != nil {
			t.Fatalf("Failed to build agent: %v", err)
		}

		if agent == nil {
			t.Fatal("Agent should not be nil")
		}

		// Get config to verify MCP is set up
		config := agent.Config()
		if config.Tools == nil {
			t.Fatal("Tools config should not be nil")
		}

		if config.Tools.MCP == nil {
			t.Fatal("MCP config should not be nil")
		}

		if !config.Tools.MCP.Enabled {
			t.Error("MCP should be enabled")
		}

		if len(config.Tools.MCP.Servers) != 1 {
			t.Errorf("Expected 1 MCP server, got %d", len(config.Tools.MCP.Servers))
		}

		if config.Tools.MCP.Servers[0].Name != "test-server" {
			t.Errorf("Expected server name 'test-server', got '%s'", config.Tools.MCP.Servers[0].Name)
		}
	})

	t.Run("WithMCP supports multiple servers", func(t *testing.T) {
		server1 := vnext.MCPServer{
			Name:    "server-1",
			Type:    "tcp",
			Address: "localhost",
			Port:    8811,
			Enabled: true,
		}

		server2 := vnext.MCPServer{
			Name:    "server-2",
			Type:    "tcp",
			Address: "localhost",
			Port:    8812,
			Enabled: true,
		}

		agent, err := vnext.NewBuilder("multi-server-agent").
			WithConfig(&vnext.Config{
				Name:         "multi-server-agent",
				SystemPrompt: "Test agent",
				Timeout:      30 * time.Second,
				LLM: vnext.LLMConfig{
					Provider:    "ollama",
					Model:       "gemma3:1b",
					Temperature: 0.7,
					MaxTokens:   100,
				},
			}).
			WithTools(vnext.WithMCP(server1, server2)).
			Build()

		if err != nil {
			t.Fatalf("Failed to build agent: %v", err)
		}

		config := agent.Config()
		if len(config.Tools.MCP.Servers) != 2 {
			t.Errorf("Expected 2 MCP servers, got %d", len(config.Tools.MCP.Servers))
		}
	})
}

// TestMCPDiscoveryConfiguration tests MCP discovery setup
func TestMCPDiscoveryConfiguration(t *testing.T) {
	t.Run("WithMCPDiscovery enables discovery", func(t *testing.T) {
		agent, err := vnext.NewBuilder("discovery-agent").
			WithConfig(&vnext.Config{
				Name:         "discovery-agent",
				SystemPrompt: "Test agent",
				Timeout:      30 * time.Second,
				LLM: vnext.LLMConfig{
					Provider:    "ollama",
					Model:       "gemma3:1b",
					Temperature: 0.7,
					MaxTokens:   100,
				},
			}).
			WithTools(vnext.WithMCPDiscovery(8080, 8081, 8090)).
			Build()

		if err != nil {
			t.Fatalf("Failed to build agent: %v", err)
		}

		config := agent.Config()
		if config.Tools == nil || config.Tools.MCP == nil {
			t.Fatal("MCP config should not be nil")
		}

		if !config.Tools.MCP.Discovery {
			t.Error("MCP discovery should be enabled")
		}

		if len(config.Tools.MCP.ScanPorts) != 3 {
			t.Errorf("Expected 3 scan ports, got %d", len(config.Tools.MCP.ScanPorts))
		}

		expectedPorts := []int{8080, 8081, 8090}
		for i, port := range expectedPorts {
			if config.Tools.MCP.ScanPorts[i] != port {
				t.Errorf("Expected port %d at index %d, got %d", port, i, config.Tools.MCP.ScanPorts[i])
			}
		}
	})

	t.Run("WithMCPDiscovery uses default ports when none provided", func(t *testing.T) {
		agent, err := vnext.NewBuilder("discovery-default-agent").
			WithConfig(&vnext.Config{
				Name:         "discovery-default-agent",
				SystemPrompt: "Test agent",
				Timeout:      30 * time.Second,
				LLM: vnext.LLMConfig{
					Provider:    "ollama",
					Model:       "gemma3:1b",
					Temperature: 0.7,
					MaxTokens:   100,
				},
			}).
			WithTools(vnext.WithMCPDiscovery()). // No ports specified
			Build()

		if err != nil {
			t.Fatalf("Failed to build agent: %v", err)
		}

		config := agent.Config()
		if !config.Tools.MCP.Discovery {
			t.Error("MCP discovery should be enabled")
		}

		// Should have default ports
		if len(config.Tools.MCP.ScanPorts) == 0 {
			t.Error("Expected default scan ports to be set")
		}

		t.Logf("Default scan ports: %v", config.Tools.MCP.ScanPorts)
	})
}

// TestMCPToolDiscovery tests that tool discovery functions work
func TestMCPToolDiscovery(t *testing.T) {
	t.Run("DiscoverInternalTools works", func(t *testing.T) {
		tools, err := vnext.DiscoverInternalTools()
		if err != nil {
			t.Fatalf("Failed to discover internal tools: %v", err)
		}

		// Should have at least the echo tool
		if len(tools) == 0 {
			t.Log("No internal tools discovered (this is acceptable)")
		} else {
			t.Logf("Discovered %d internal tools", len(tools))
			for _, tool := range tools {
				t.Logf("  - %s: %s", tool.Name(), tool.Description())
			}
		}
	})

	t.Run("DiscoverMCPTools handles no MCP manager gracefully", func(t *testing.T) {
		// This should not panic even if no MCP manager is initialized
		tools, err := vnext.DiscoverMCPTools()

		// We expect an error since no MCP manager is set up
		if err == nil {
			t.Log("No error when MCP manager not available (acceptable)")
		} else {
			t.Logf("Expected error when no MCP manager: %v", err)
		}

		if tools != nil && len(tools) > 0 {
			t.Logf("Discovered %d MCP tools", len(tools))
		}
	})
}

// TestMCPValidation tests MCP configuration validation
func TestMCPValidation(t *testing.T) {
	t.Run("Invalid server type is caught", func(t *testing.T) {
		config := &vnext.Config{
			Name:         "test-agent",
			SystemPrompt: "Test",
			LLM: vnext.LLMConfig{
				Provider: "ollama",
				Model:    "gemma3:1b",
			},
			Tools: &vnext.ToolsConfig{
				Enabled: true,
				MCP: &vnext.MCPConfig{
					Enabled: true,
					Servers: []vnext.MCPServer{
						{
							Name:    "bad-server",
							Type:    "invalid-type", // Invalid!
							Address: "localhost",
							Port:    8080,
							Enabled: true,
						},
					},
				},
			},
		}

		err := vnext.ValidateConfig(config)
		if err == nil {
			t.Error("Expected validation error for invalid server type")
		} else {
			t.Logf("Validation correctly caught error: %v", err)
		}
	})

	t.Run("Missing server name is caught", func(t *testing.T) {
		config := &vnext.Config{
			Name:         "test-agent",
			SystemPrompt: "Test",
			LLM: vnext.LLMConfig{
				Provider: "ollama",
				Model:    "gemma3:1b",
			},
			Tools: &vnext.ToolsConfig{
				Enabled: true,
				MCP: &vnext.MCPConfig{
					Enabled: true,
					Servers: []vnext.MCPServer{
						{
							Name:    "", // Missing!
							Type:    "tcp",
							Address: "localhost",
							Port:    8080,
							Enabled: true,
						},
					},
				},
			},
		}

		err := vnext.ValidateConfig(config)
		if err == nil {
			t.Error("Expected validation error for missing server name")
		} else {
			t.Logf("Validation correctly caught error: %v", err)
		}
	})

	t.Run("TCP server without address/port is caught", func(t *testing.T) {
		config := &vnext.Config{
			Name:         "test-agent",
			SystemPrompt: "Test",
			LLM: vnext.LLMConfig{
				Provider: "ollama",
				Model:    "gemma3:1b",
			},
			Tools: &vnext.ToolsConfig{
				Enabled: true,
				MCP: &vnext.MCPConfig{
					Enabled: true,
					Servers: []vnext.MCPServer{
						{
							Name:    "tcp-server",
							Type:    "tcp",
							Address: "", // Missing!
							Port:    0,  // Missing!
							Enabled: true,
						},
					},
				},
			},
		}

		err := vnext.ValidateConfig(config)
		if err == nil {
			t.Error("Expected validation error for missing address and port")
		} else {
			t.Logf("Validation correctly caught error: %v", err)
		}
	})

	t.Run("STDIO server without command is caught", func(t *testing.T) {
		config := &vnext.Config{
			Name:         "test-agent",
			SystemPrompt: "Test",
			LLM: vnext.LLMConfig{
				Provider: "ollama",
				Model:    "gemma3:1b",
			},
			Tools: &vnext.ToolsConfig{
				Enabled: true,
				MCP: &vnext.MCPConfig{
					Enabled: true,
					Servers: []vnext.MCPServer{
						{
							Name:    "stdio-server",
							Type:    "stdio",
							Command: "", // Missing!
							Enabled: true,
						},
					},
				},
			},
		}

		err := vnext.ValidateConfig(config)
		if err == nil {
			t.Error("Expected validation error for missing STDIO command")
		} else {
			t.Logf("Validation correctly caught error: %v", err)
		}
	})

	t.Run("Valid configurations pass validation", func(t *testing.T) {
		validConfigs := []vnext.MCPServer{
			{
				Name:    "tcp-server",
				Type:    "tcp",
				Address: "localhost",
				Port:    8080,
				Enabled: true,
			},
			{
				Name:    "stdio-server",
				Type:    "stdio",
				Command: "python mcp_server.py",
				Enabled: true,
			},
			{
				Name:    "ws-server",
				Type:    "websocket",
				Address: "localhost",
				Port:    8081,
				Enabled: true,
			},
			{
				Name:    "http-sse-server",
				Type:    "http_sse",
				Address: "http://localhost:8082/mcp",
				Enabled: true,
			},
		}

		for _, server := range validConfigs {
			config := &vnext.Config{
				Name:         "test-agent",
				SystemPrompt: "Test",
				LLM: vnext.LLMConfig{
					Provider: "ollama",
					Model:    "gemma3:1b",
				},
				Tools: &vnext.ToolsConfig{
					Enabled: true,
					MCP: &vnext.MCPConfig{
						Enabled: true,
						Servers: []vnext.MCPServer{server},
					},
				},
			}

			err := vnext.ValidateConfig(config)
			if err != nil {
				t.Errorf("Valid %s server config failed validation: %v", server.Type, err)
			}
		}
	})
}

// TestMCPTimeouts tests timeout configuration
func TestMCPTimeouts(t *testing.T) {
	t.Run("Connection timeout is set", func(t *testing.T) {
		agent, err := vnext.NewBuilder("timeout-agent").
			WithConfig(&vnext.Config{
				Name:         "timeout-agent",
				SystemPrompt: "Test agent",
				Timeout:      30 * time.Second,
				LLM: vnext.LLMConfig{
					Provider:    "ollama",
					Model:       "gemma3:1b",
					Temperature: 0.7,
					MaxTokens:   100,
				},
			}).
			WithTools(
				vnext.WithMCP(vnext.MCPServer{
					Name:    "test-server",
					Type:    "tcp",
					Address: "localhost",
					Port:    8811,
					Enabled: true,
				}),
				vnext.WithToolTimeout(45*time.Second),
			).
			Build()

		if err != nil {
			t.Fatalf("Failed to build agent: %v", err)
		}

		config := agent.Config()
		if config.Tools.Timeout != 45*time.Second {
			t.Errorf("Expected tool timeout 45s, got %v", config.Tools.Timeout)
		}
	})
}

// TestMCPRealIntegration tests with actual MCP (requires MCP server running)
func TestMCPRealIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real MCP integration test in short mode")
	}

	t.Run("Agent can be created with MCP configuration", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server := vnext.MCPServer{
			Name:    "docker-mcp",
			Type:    "tcp",
			Address: "localhost",
			Port:    8812,
			Enabled: true,
		}

		agent, err := vnext.NewBuilder("real-mcp-agent").
			WithConfig(&vnext.Config{
				Name:         "real-mcp-agent",
				SystemPrompt: "You are a helpful assistant with MCP tools.",
				Timeout:      60 * time.Second,
				LLM: vnext.LLMConfig{
					Provider:    "ollama",
					Model:       "gemma3:1b",
					Temperature: 0.7,
					MaxTokens:   150,
				},
			}).
			WithTools(
				vnext.WithMCP(server),
				vnext.WithToolTimeout(30*time.Second),
			).
			Build()

		if err != nil {
			t.Fatalf("Failed to build agent: %v", err)
		}

		// Try to run the agent (will fail if no MCP server, but agent creation should succeed)
		result, err := agent.Run(ctx, "Hello")
		if err != nil {
			t.Logf("Agent run failed (expected if no MCP server): %v", err)
		} else {
			t.Logf("Agent response: %s", result.Content)
			t.Logf("Duration: %v", result.Duration)
		}
	})
}


