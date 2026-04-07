package v1beta

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
	"github.com/agenticgokit/agenticgokit/internal/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
	Logger().Debug().Str("tool", name).Msg("Registered vnext internal tool")
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
			Logger().Debug().Str("tool", name).Msg("Discovered vnext internal tool")
		}
	}
	return tools, nil
}

// DiscoverMCPTools discovers tools available through the core MCP manager
func DiscoverMCPTools() ([]Tool, error) {
	mgr := GetMCPManager()
	if mgr == nil {
		Logger().Debug().Msg("MCP manager not available")
		return nil, fmt.Errorf("MCP manager not available")
	}

	mcpToolInfos := mgr.GetAvailableTools()
	Logger().Debug().Int("count", len(mcpToolInfos)).Msg("GetAvailableTools returned")

	var tools []Tool
	for _, info := range mcpToolInfos {
		wrapper := &mcpToolWrapper{
			name:        info.Name,
			description: info.Description,
			parameters:  info.Schema,
			manager:     mgr,
		}
		tools = append(tools, wrapper)
		Logger().Debug().Str("tool", info.Name).Str("server", info.ServerName).Msg("Discovered MCP tool")
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
		Logger().Warn().Err(err).Msg("Failed to discover internal tools")
	}

	// Discover MCP tools
	if mcpTools, err := DiscoverMCPTools(); err == nil {
		allTools = append(allTools, mcpTools...)
	} else {
		Logger().Warn().Err(err).Msg("Failed to discover MCP tools")
	}

	Logger().Debug().Int("tool_count", len(allTools)).Msg("Tool discovery completed")
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
			Logger().Debug().Str("tool", toolName).Interface("args", args).Msg("Executing tool")
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
			Logger().Error().Err(err).Str("tool", toolName).Msg("Tool execution failed")
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

// JSONSchema implements ToolWithSchema so that the MCP tool's InputSchema
// is passed to the LLM for proper function-calling parameter generation.
func (m *mcpToolWrapper) JSONSchema() map[string]interface{} {
	if m.parameters != nil {
		return m.parameters
	}
	// Minimal fallback – no schema info returned from MCP server
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (m *mcpToolWrapper) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	// Create observability span for MCP tool execution
	tracer := otel.Tracer("agenticgokit.mcp")
	ctx, span := tracer.Start(ctx, "agk.mcp.tool.call",
		trace.WithAttributes(
			attribute.String(observability.AttrToolName, m.name),
			attribute.String(observability.AttrMCPServer, "unknown"), // ServerName not directly accessible here
		))
	defer span.End()

	startTime := time.Now()

	// Record input size
	inputSize := 0
	if len(args) > 0 {
		if jsonBytes, err := json.Marshal(args); err == nil {
			inputSize = len(jsonBytes)
		}
	}
	span.SetAttributes(attribute.Int("agk.mcp.tool.input_bytes", inputSize))

	// Execute through core MCP manager
	mcpResult, err := ExecuteMCPTool(ctx, m.name, args)
	if err != nil {
		duration := time.Since(startTime)
		span.RecordError(err)
		span.SetStatus(codes.Error, "MCP tool execution failed")
		span.SetAttributes(attribute.Int64(observability.AttrToolLatencyMs, duration.Milliseconds()))
		return &ToolResult{Success: false, Error: err.Error()}, err
	}

	// Convert MCP result to vnext ToolResult
	// vnext.ToolResult.Content is interface{}, so we use a flexible representation
	var contents []map[string]interface{}
	outputSize := 0
	for _, content := range mcpResult.Content {
		contentMap := map[string]interface{}{
			"type": content.Type,
			"text": content.Text,
			"data": content.Data,
		}
		contents = append(contents, contentMap)

		// Count bytes in content
		if jsonBytes, err := json.Marshal(contentMap); err == nil {
			outputSize += len(jsonBytes)
		}
	}

	duration := time.Since(startTime)
	resultStatus := codes.Ok
	if !mcpResult.Success {
		resultStatus = codes.Error
	}

	span.SetAttributes(
		attribute.Int64(observability.AttrToolLatencyMs, duration.Milliseconds()),
		attribute.Int("agk.mcp.tool.output_bytes", outputSize),
		attribute.Bool("agk.mcp.tool.success", mcpResult.Success),
		attribute.Int("agk.mcp.tool.content_count", len(mcpResult.Content)),
	)

	if !mcpResult.Success {
		span.SetStatus(resultStatus, "MCP tool returned success=false")
	} else {
		span.SetStatus(resultStatus, "MCP tool execution successful")
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
// DISABLED: This example tool can interfere with production usage
// Users can re-enable by uncommenting this block
/*
func init() {
	RegisterInternalTool("echo", func() Tool {
		return &echoTool{}
	})
}
*/
