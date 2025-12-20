package vnext

import (
	"context"
	"strings"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
	"github.com/agenticgokit/agenticgokit/internal/llm"
)

// createLLMProvider creates a ModelProvider instance from LLMConfig.
// This function maps vNext configuration to the appropriate LLM provider
// implementation in internal/llm/ using the existing factory.
//
// Supported providers:
//   - ollama: Local Ollama instance
//   - openai: OpenAI API
//   - azure: Azure OpenAI Service
//
// Returns the initialized provider or an error if the provider is unsupported
// or initialization fails.
func createLLMProvider(config LLMConfig) (llm.ModelProvider, error) {
	// Map vNext config to internal/llm ProviderConfig
	providerType := llm.ProviderType(strings.ToLower(strings.TrimSpace(config.Provider)))

	llmConfig := llm.ProviderConfig{
		Type:        providerType,
		APIKey:      config.APIKey,
		Model:       config.Model,
		MaxTokens:   config.MaxTokens,
		Temperature: config.Temperature,
		BaseURL:     config.BaseURL,
		Endpoint:    config.BaseURL, // For Azure
	}

	// Use the internal/llm factory to create the provider
	factory := llm.NewProviderFactory()
	return factory.CreateProvider(llmConfig)
}

// createMemoryProvider creates a Memory instance from MemoryConfig.
// Returns nil if config is nil (memory disabled).
//
// Supported memory providers:
//   - memory: Simple in-memory storage
//   - pgvector: PostgreSQL with pgvector extension
//   - weaviate: Weaviate vector database
//
// Returns the initialized provider, nil (if disabled), or an error.
func createMemoryProvider(config *MemoryConfig) (core.Memory, error) {
	// Memory is optional - return nil if not configured
	if config == nil {
		return nil, nil
	}

	// Map vNext MemoryConfig to core.AgentMemoryConfig
	agentMemoryConfig := core.AgentMemoryConfig{
		Provider:   config.Provider,
		Connection: config.Connection,
		MaxResults: 10,   // Default value (not in vNext MemoryConfig)
		Dimensions: 1536, // Default embedding dimensions
		AutoEmbed:  true,

		// RAG configuration if available
		EnableRAG:               config.RAG != nil,
		RAGMaxContextTokens:     0,
		RAGPersonalWeight:       0.3,
		RAGKnowledgeWeight:      0.7,
		EnableKnowledgeBase:     config.RAG != nil,
		KnowledgeMaxResults:     10,
		KnowledgeScoreThreshold: 0.3,
	}

	// Apply RAG config if present
	if config.RAG != nil {
		agentMemoryConfig.RAGMaxContextTokens = config.RAG.MaxTokens
		agentMemoryConfig.RAGPersonalWeight = config.RAG.PersonalWeight
		agentMemoryConfig.RAGKnowledgeWeight = config.RAG.KnowledgeWeight
		if config.RAG.HistoryLimit > 0 {
			agentMemoryConfig.KnowledgeMaxResults = config.RAG.HistoryLimit
		}
	}

	// Use core.NewMemory which handles factory lookup internally
	// The factory is registered via init() in internal/memory/factory.go
	return core.NewMemory(agentMemoryConfig)
}

// createTools creates a list of Tool instances from ToolsConfig.
// Returns nil if config is nil or tools are disabled.
//
// If MCP (Model Context Protocol) is enabled, this function will:
//   - Initialize MCP with configured servers
//   - Discover available MCP tools
//   - Combine with internal tools
//
// Returns the list of initialized tools, nil (if disabled), or an error.
func createTools(config *ToolsConfig) ([]Tool, error) {
	// Tools are optional - return nil if not configured or disabled
	if config == nil || !config.Enabled {
		return nil, nil
	}

	var allTools []Tool

	// Step 1: Discover internal tools (always available)
	internalTools, err := DiscoverInternalTools()
	if err != nil {
		// Log warning but continue - internal tools are optional
		core.Logger().Warn().Err(err).Msg("Failed to discover internal tools")
	} else if len(internalTools) > 0 {
		allTools = append(allTools, internalTools...)
		core.Logger().Debug().Int("count", len(internalTools)).Msg("Discovered internal tools")
	}

	// Step 2: Initialize and discover MCP tools if enabled
	if config.MCP != nil && config.MCP.Enabled {
		core.Logger().Debug().Msg("MCP is enabled, initializing...")
		if err := initializeMCP(config.MCP); err != nil {
			// MCP initialization failure is not fatal - log and continue with internal tools
			core.Logger().Warn().Err(err).Msg("Failed to initialize MCP, continuing without MCP tools")
		} else {
			core.Logger().Debug().Msg("MCP initialized successfully, discovering tools...")
			// Discover MCP tools
			mcpTools, err := DiscoverMCPTools()
			if err != nil {
				core.Logger().Warn().Err(err).Msg("Failed to discover MCP tools")
			} else if len(mcpTools) > 0 {
				allTools = append(allTools, mcpTools...)
				core.Logger().Debug().Int("count", len(mcpTools)).Msg("Discovered MCP tools")
			} else {
				core.Logger().Warn().Msg("DiscoverMCPTools returned zero tools")
			}
		}
	}

	// Return the combined list of tools
	core.Logger().Info().Int("total_tools", len(allTools)).Msg("Tool initialization completed")
	return allTools, nil
}

// initializeMCP initializes the MCP manager with the provided configuration.
// This is a helper function that maps vNext MCPConfig to core.MCPConfig.
func initializeMCP(config *MCPConfig) error {
	if config == nil {
		return nil
	}

	// Map vNext MCPConfig to core.MCPConfig
	coreMCPConfig := core.MCPConfig{
		// Server configuration (filter enabled servers only)
		Servers: []core.MCPServerConfig{},

		// Discovery settings
		EnableDiscovery:  config.Discovery,
		DiscoveryTimeout: config.DiscoveryTimeout,
		ScanPorts:        config.ScanPorts,

		// Connection settings
		ConnectionTimeout: config.ConnectionTimeout,
		MaxRetries:        config.MaxRetries,
		RetryDelay:        config.RetryDelay,

		// Cache configuration
		EnableCaching: config.Cache != nil && config.Cache.Enabled,
		CacheTimeout:  10 * time.Minute, // Default cache timeout
	}

	// Map server configurations (only enabled servers)
	for _, server := range config.Servers {
		if !server.Enabled {
			continue // Skip disabled servers
		}

		coreMCPConfig.Servers = append(coreMCPConfig.Servers, core.MCPServerConfig{
			Name:    server.Name,
			Type:    server.Type,
			Host:    server.Address, // Map Address to Host
			Port:    server.Port,
			Command: server.Command,
			Enabled: server.Enabled,
		})
	}

	// Initialize MCP with cache if configured
	if config.Cache != nil && config.Cache.Enabled {
		cacheConfig := core.MCPCacheConfig{
			Enabled:         true,
			DefaultTTL:      config.Cache.TTL,
			MaxSize:         config.Cache.MaxSize,
			MaxKeys:         config.Cache.MaxKeys,
			EvictionPolicy:  config.Cache.EvictionPolicy,
			CleanupInterval: config.Cache.CleanupInterval,
			ToolTTLs:        config.Cache.ToolTTLs,
			Backend:         config.Cache.Backend,
			BackendConfig:   config.Cache.BackendConfig,
		}
		return core.InitializeMCPWithCache(coreMCPConfig, cacheConfig)
	}

	// Initialize MCP without cache
	if err := core.InitializeMCP(coreMCPConfig); err != nil {
		return err
	}

	// Connect to all configured servers
	mgr := core.GetMCPManager()
	if mgr != nil {
		ctx := context.Background()
		for _, server := range config.Servers {
			if server.Enabled {
				core.Logger().Debug().Str("server", server.Name).Msg("Connecting to MCP server...")
				if err := mgr.Connect(ctx, server.Name); err != nil {
					core.Logger().Warn().
						Err(err).
						Str("server", server.Name).
						Msg("Failed to connect to MCP server")
				} else {
					core.Logger().Debug().
						Str("server", server.Name).
						Msg("Successfully connected to MCP server")
				}
			}
		}

		// Initialize MCP tool registry (required for MCP tools to be available)
		core.Logger().Debug().Msg("Initializing MCP tool registry...")
		if err := core.InitializeMCPToolRegistry(); err != nil {
			core.Logger().Warn().Err(err).Msg("Failed to initialize MCP tool registry")
		}

		// Register MCP tools with registry (discovers and registers tools from connected servers)
		core.Logger().Debug().Msg("Registering MCP tools from connected servers...")
		if err := core.RegisterMCPToolsWithRegistry(ctx); err != nil {
			core.Logger().Warn().Err(err).Msg("Failed to register MCP tools")
		} else {
			tools := mgr.GetAvailableTools()
			core.Logger().Debug().Int("tool_count", len(tools)).Msg("MCP tools registered successfully")
		}
	}

	return nil
}

