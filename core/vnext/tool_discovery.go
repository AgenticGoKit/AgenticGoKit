package vnext

import (
	"context"
	"fmt"

	"github.com/agenticgokit/agenticgokit/core"
)

// =============================================================================
// VNEXT TOOL DISCOVERY AND EXECUTION
// =============================================================================
// This file provides vnext-local tool discovery and execution helpers.
// It wraps core MCP functionality and manages a vnext-specific tool registry.

// Internal tool registry for vnext-registered tools
var vnextToolRegistry = make(map[string]func() Tool)

// RegisterInternalTool registers a tool factory in the vnext registry
func RegisterInternalTool(name string, factory func() Tool) {
	vnextToolRegistry[name] = factory
	core.Logger().Debug().Str("tool", name).Msg("Registered vnext internal tool")
}

func getInternalToolRegistry() map[string]func() Tool {
	return vnextToolRegistry
}

// DiscoverInternalTools returns all tools registered via RegisterInternalTool
func DiscoverInternalTools() ([]Tool, error) {
	registry := getInternalToolRegistry()
	var tools []Tool
	for name, factory := range registry {
		if tool := factory(); tool != nil {
			tools = append(tools, tool)
			core.Logger().Debug().Str("tool", name).Msg("Discovered vnext internal tool")
		}
	}
	return tools, nil
}

// DiscoverMCPTools discovers tools available through the core MCP manager
func DiscoverMCPTools() ([]Tool, error) {
	mgr := core.GetMCPManager()
	if mgr == nil {
		core.Logger().Debug().Msg("MCP manager not available")
		return nil, fmt.Errorf("MCP manager not available")
	}

	mcpToolInfos := mgr.GetAvailableTools()
	core.Logger().Debug().Int("count", len(mcpToolInfos)).Msg("GetAvailableTools returned")

	var tools []Tool
	for _, info := range mcpToolInfos {
		wrapper := &mcpToolWrapper{
			name:        info.Name,
			description: info.Description,
			parameters:  info.Schema,
			manager:     mgr,
		}
		tools = append(tools, wrapper)
		core.Logger().Debug().Str("tool", info.Name).Str("server", info.ServerName).Msg("Discovered MCP tool")
	}
	return tools, nil
}

// DiscoverTools aggregates all available tools (internal + MCP)
func DiscoverTools() ([]Tool, error) {
	var allTools []Tool

	// Discover internal tools
	if internalTools, err := DiscoverInternalTools(); err == nil {
		allTools = append(allTools, internalTools...)
	} else {
		core.Logger().Warn().Err(err).Msg("Failed to discover internal tools")
	}

	// Discover MCP tools
	if mcpTools, err := DiscoverMCPTools(); err == nil {
		allTools = append(allTools, mcpTools...)
	} else {
		core.Logger().Warn().Err(err).Msg("Failed to discover MCP tools")
	}

	core.Logger().Debug().Int("tool_count", len(allTools)).Msg("Tool discovery completed")
	return allTools, nil
}

// ExecuteToolByName finds and executes a tool by name
func ExecuteToolByName(ctx context.Context, toolName string, args map[string]interface{}) (*ToolResult, error) {
	tools, err := DiscoverTools()
	if err != nil {
		return nil, fmt.Errorf("failed to discover tools: %w", err)
	}

	for _, tool := range tools {
		if tool.Name() == toolName {
			core.Logger().Debug().Str("tool", toolName).Interface("args", args).Msg("Executing tool")
			return tool.Execute(ctx, args)
		}
	}

	return nil, fmt.Errorf("tool not found: %s", toolName)
}

// ExecuteToolsFromLLMResponse parses and executes tool calls from LLM responses
func ExecuteToolsFromLLMResponse(ctx context.Context, llmResponse string) ([]ToolResult, error) {
	toolCalls := ParseLLMToolCalls(llmResponse)
	if len(toolCalls) == 0 {
		return nil, nil
	}

	var results []ToolResult
	for _, toolCall := range toolCalls {
		toolName, ok := toolCall["name"].(string)
		if !ok {
			continue
		}

		args, _ := toolCall["args"].(map[string]interface{})
		if args == nil {
			args = make(map[string]interface{})
		}

		result, err := ExecuteToolByName(ctx, toolName, args)
		if err != nil {
			core.Logger().Error().Err(err).Str("tool", toolName).Msg("Tool execution failed")
			results = append(results, ToolResult{
				Success: false,
				Error:   err.Error(),
			})
		} else {
			results = append(results, *result)
		}
	}

	return results, nil
}

// =============================================================================
// MCP TOOL WRAPPER
// =============================================================================

// mcpToolWrapper adapts MCP tools to implement the vnext Tool interface
type mcpToolWrapper struct {
	name        string
	description string
	parameters  map[string]interface{}
	manager     core.MCPManager
}

func (m *mcpToolWrapper) Name() string {
	return m.name
}

func (m *mcpToolWrapper) Description() string {
	return m.description
}

func (m *mcpToolWrapper) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	// Execute through core MCP manager
	mcpResult, err := core.ExecuteMCPTool(ctx, m.name, args)
	if err != nil {
		return &ToolResult{Success: false, Error: err.Error()}, err
	}

	// Convert MCP result to vnext ToolResult
	// vnext.ToolResult.Content is interface{}, so we use a flexible representation
	var contents []map[string]interface{}
	for _, content := range mcpResult.Content {
		contents = append(contents, map[string]interface{}{
			"type": content.Type,
			"text": content.Text,
			"data": content.Data,
		})
	}

	return &ToolResult{
		Success: mcpResult.Success,
		Content: contents,
	}, nil
}

// =============================================================================
// EXAMPLE INTERNAL TOOL
// =============================================================================

// echoTool is a simple example internal tool
type echoTool struct{}

func (e *echoTool) Name() string {
	return "echo"
}

func (e *echoTool) Description() string {
	return "Echoes back the provided message"
}

func (e *echoTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	message, ok := args["message"].(string)
	if !ok || message == "" {
		return &ToolResult{
			Success: false,
			Error:   "message parameter is required and must be a non-empty string",
		}, nil
	}

	return &ToolResult{
		Success: true,
		Content: fmt.Sprintf("Echo: %s", message),
	}, nil
}

// Register the echo tool on package initialization
func init() {
	RegisterInternalTool("echo", func() Tool {
		return &echoTool{}
	})
}

