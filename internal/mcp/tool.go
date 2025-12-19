// Package mcp provides internal implementation for Model Context Protocol (MCP) integration.
//
// This package contains the concrete implementations of MCP tools and managers
// that implement the public interfaces defined in core/mcp.go.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/client"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/mcp"
)

// MCPTool is an adapter that wraps an MCP tool to implement the AgentFlow FunctionTool interface.
// It manages the connection to an MCP server and translates calls between AgentFlow and MCP protocols.
type MCPTool struct {
	name        string
	description string
	schema      map[string]interface{}
	serverName  string
	client      *client.Client
	manager     *MCPManagerImpl
	callTimeout time.Duration
}

// NewMCPTool creates a new MCP tool adapter.
func NewMCPTool(toolInfo mcp.Tool, serverName string, mcpClient *client.Client, manager *MCPManagerImpl) *MCPTool {
	return &MCPTool{
		name:        toolInfo.Name,
		description: toolInfo.Description,
		schema:      toolInfo.InputSchema,
		serverName:  serverName,
		client:      mcpClient,
		manager:     manager,
		callTimeout: 30 * time.Second, // Default timeout
	}
}

// Name returns the unique identifier for the tool.
// This implements the FunctionTool interface.
func (t *MCPTool) Name() string {
	return fmt.Sprintf("mcp_%s_%s", t.serverName, t.name)
}

// Call executes the MCP tool with the given arguments.
// This implements the FunctionTool interface.
func (t *MCPTool) Call(ctx context.Context, args map[string]any) (map[string]any, error) {
	// Create a timeout context for the MCP call
	callCtx, cancel := context.WithTimeout(ctx, t.callTimeout)
	defer cancel()

	// Record call start time for metrics
	startTime := time.Now()

	// Update manager metrics
	t.manager.recordToolCall(t.serverName, startTime)

	// Convert arguments to the format expected by MCP
	mcpArgs, err := t.convertArgumentsToMCP(args)
	if err != nil {
		t.manager.recordToolError(t.serverName, startTime, err)
		return nil, fmt.Errorf("failed to convert arguments: %w", err)
	}

	// Validate arguments against schema if available
	if err := t.validateArguments(mcpArgs); err != nil {
		t.manager.recordToolError(t.serverName, startTime, err)
		return nil, fmt.Errorf("argument validation failed: %w", err)
	}

	// Check if client is still connected
	if !t.client.IsConnected() || !t.client.IsInitialized() {
		err := fmt.Errorf("MCP client for server '%s' is not connected or initialized", t.serverName)
		t.manager.recordToolError(t.serverName, startTime, err)
		return nil, err
	}

	// Execute the MCP tool
	response, err := t.client.CallTool(callCtx, t.name, mcpArgs)
	if err != nil {
		t.manager.recordToolError(t.serverName, startTime, err)
		return nil, fmt.Errorf("MCP tool execution failed: %w", err)
	}

	// Check if the tool execution returned an error
	if response.IsError {
		err := fmt.Errorf("MCP tool returned error: %s", t.formatMCPContent(response.Content))
		t.manager.recordToolError(t.serverName, startTime, err)
		return nil, err
	}

	// Convert MCP response to AgentFlow format
	result, err := t.convertMCPResponseToAgentFlow(response)
	if err != nil {
		t.manager.recordToolError(t.serverName, startTime, err)
		return nil, fmt.Errorf("failed to convert MCP response: %w", err)
	}

	// Record successful call
	t.manager.recordToolSuccess(t.serverName, startTime)

	return result, nil
}

// GetSchema returns the tool's input schema for use by LLMs.
func (t *MCPTool) GetSchema() map[string]interface{} {
	return t.schema
}

// GetDescription returns the tool's description.
func (t *MCPTool) GetDescription() string {
	return t.description
}

// GetServerName returns the name of the MCP server this tool belongs to.
func (t *MCPTool) GetServerName() string {
	return t.serverName
}

// SetTimeout sets the timeout for tool calls.
func (t *MCPTool) SetTimeout(timeout time.Duration) {
	t.callTimeout = timeout
}

// convertArgumentsToMCP converts AgentFlow arguments to MCP format.
func (t *MCPTool) convertArgumentsToMCP(args map[string]any) (map[string]interface{}, error) {
	if args == nil {
		return nil, nil
	}

	// For now, we can pass arguments through directly since both use map[string]interface{}
	// In the future, we might want to add type conversion or validation here
	mcpArgs := make(map[string]interface{}, len(args))
	for key, value := range args {
		mcpArgs[key] = value
	}

	return mcpArgs, nil
}

// validateArguments validates the arguments against the tool's schema.
func (t *MCPTool) validateArguments(args map[string]interface{}) error {
	// Basic validation - check required fields if schema defines them
	if t.schema == nil {
		return nil // No schema to validate against
	}

	// Check if schema defines required properties
	if properties, ok := t.schema["properties"].(map[string]interface{}); ok {
		if required, ok := t.schema["required"].([]interface{}); ok {
			for _, reqField := range required {
				if fieldName, ok := reqField.(string); ok {
					if _, exists := args[fieldName]; !exists {
						return fmt.Errorf("required argument '%s' is missing", fieldName)
					}
				}
			}
		}

		// Basic type checking for provided arguments
		for argName, argValue := range args {
			if propSchema, exists := properties[argName]; exists {
				if err := t.validateArgumentType(argName, argValue, propSchema); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// validateArgumentType performs basic type validation for an argument.
func (t *MCPTool) validateArgumentType(name string, value interface{}, schema interface{}) error {
	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		return nil // Can't validate if schema is not a map
	}

	expectedType, ok := schemaMap["type"].(string)
	if !ok {
		return nil // No type specified in schema
	}

	// Basic type checking
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("argument '%s' must be a string, got %T", name, value)
		}
	case "number", "integer":
		switch value.(type) {
		case int, int32, int64, float32, float64:
			// Valid numeric types
		default:
			return fmt.Errorf("argument '%s' must be a number, got %T", name, value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("argument '%s' must be a boolean, got %T", name, value)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("argument '%s' must be an array, got %T", name, value)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("argument '%s' must be an object, got %T", name, value)
		}
	}

	return nil
}

// convertMCPResponseToAgentFlow converts an MCP response to AgentFlow format.
// This method handles multimodal content (images, audio, video) from MCP tool responses.
func (t *MCPTool) convertMCPResponseToAgentFlow(response *mcp.CallToolResponse) (map[string]any, error) {
	result := make(map[string]any)

	// Convert content array to AgentFlow format
	if len(response.Content) > 0 {
		var contents []map[string]interface{}
		var textParts []string
		var images []map[string]interface{}
		var audioFiles []map[string]interface{}
		var videoFiles []map[string]interface{}
		var attachments []map[string]interface{}

		for _, content := range response.Content {
			contentMap := map[string]interface{}{
				"type": content.Type,
			}

			if content.Text != "" {
				contentMap["text"] = content.Text
				textParts = append(textParts, content.Text)
			}
			if content.Data != "" {
				contentMap["data"] = content.Data
			}
			if content.MimeType != "" {
				contentMap["mime_type"] = content.MimeType
			}
			if content.Name != "" {
				contentMap["name"] = content.Name
			}
			if content.URI != "" {
				contentMap["uri"] = content.URI
			}
			if content.Annotations != nil {
				contentMap["annotations"] = content.Annotations
			}

			contents = append(contents, contentMap)

			// Categorize multimodal content based on type or MIME type
			switch content.Type {
			case "image":
				images = append(images, t.buildImageData(content))
			case "audio":
				audioFiles = append(audioFiles, t.buildAudioData(content))
			case "video":
				videoFiles = append(videoFiles, t.buildVideoData(content))
			default:
				// Check MIME type for content categorization
				if content.MimeType != "" {
					if isImageMimeType(content.MimeType) {
						images = append(images, t.buildImageData(content))
					} else if isAudioMimeType(content.MimeType) {
						audioFiles = append(audioFiles, t.buildAudioData(content))
					} else if isVideoMimeType(content.MimeType) {
						videoFiles = append(videoFiles, t.buildVideoData(content))
					} else if content.Data != "" || content.URI != "" {
						// Generic attachment
						attachments = append(attachments, t.buildAttachmentData(content))
					}
				}
			}
		}

		result["content"] = contents

		// Provide a simplified text output for easy consumption
		if len(textParts) > 0 {
			result["text"] = textParts[0] // First text content
			if len(textParts) > 1 {
				result["all_text"] = textParts // All text content
			}
		}

		// Add categorized multimodal content
		if len(images) > 0 {
			result["images"] = images
		}
		if len(audioFiles) > 0 {
			result["audio"] = audioFiles
		}
		if len(videoFiles) > 0 {
			result["video"] = videoFiles
		}
		if len(attachments) > 0 {
			result["attachments"] = attachments
		}
	}

	// Add metadata
	result["tool_name"] = t.name
	result["server_name"] = t.serverName
	result["success"] = !response.IsError

	return result, nil
}

// formatMCPContent formats MCP content for error messages.
func (t *MCPTool) formatMCPContent(contents []mcp.Content) string {
	if len(contents) == 0 {
		return "no content"
	}

	var parts []string
	for _, content := range contents {
		if content.Text != "" {
			parts = append(parts, content.Text)
		}
	}

	if len(parts) == 0 {
		return fmt.Sprintf("content with %d items", len(contents))
	}

	if len(parts) == 1 {
		return parts[0]
	}

	return fmt.Sprintf("%s (and %d more)", parts[0], len(parts)-1)
}

// buildImageData creates an image data map from MCP content.
func (t *MCPTool) buildImageData(content mcp.Content) map[string]interface{} {
	imageData := map[string]interface{}{}
	if content.URI != "" {
		imageData["url"] = content.URI
	}
	if content.Data != "" {
		imageData["base64"] = content.Data
	}
	if content.Name != "" || content.MimeType != "" {
		metadata := map[string]string{}
		if content.Name != "" {
			metadata["name"] = content.Name
		}
		if content.MimeType != "" {
			metadata["mime_type"] = content.MimeType
		}
		imageData["metadata"] = metadata
	}
	return imageData
}

// buildAudioData creates an audio data map from MCP content.
func (t *MCPTool) buildAudioData(content mcp.Content) map[string]interface{} {
	audioData := map[string]interface{}{}
	if content.URI != "" {
		audioData["url"] = content.URI
	}
	if content.Data != "" {
		audioData["base64"] = content.Data
	}
	// Extract format from MIME type (e.g., "audio/mp3" -> "mp3")
	if content.MimeType != "" {
		audioData["format"] = extractFormatFromMimeType(content.MimeType)
	}
	if content.Name != "" || content.MimeType != "" {
		metadata := map[string]string{}
		if content.Name != "" {
			metadata["name"] = content.Name
		}
		if content.MimeType != "" {
			metadata["mime_type"] = content.MimeType
		}
		audioData["metadata"] = metadata
	}
	return audioData
}

// buildVideoData creates a video data map from MCP content.
func (t *MCPTool) buildVideoData(content mcp.Content) map[string]interface{} {
	videoData := map[string]interface{}{}
	if content.URI != "" {
		videoData["url"] = content.URI
	}
	if content.Data != "" {
		videoData["base64"] = content.Data
	}
	// Extract format from MIME type (e.g., "video/mp4" -> "mp4")
	if content.MimeType != "" {
		videoData["format"] = extractFormatFromMimeType(content.MimeType)
	}
	if content.Name != "" || content.MimeType != "" {
		metadata := map[string]string{}
		if content.Name != "" {
			metadata["name"] = content.Name
		}
		if content.MimeType != "" {
			metadata["mime_type"] = content.MimeType
		}
		videoData["metadata"] = metadata
	}
	return videoData
}

// buildAttachmentData creates an attachment data map from MCP content.
func (t *MCPTool) buildAttachmentData(content mcp.Content) map[string]interface{} {
	attachmentData := map[string]interface{}{}
	if content.Name != "" {
		attachmentData["name"] = content.Name
	}
	if content.MimeType != "" {
		attachmentData["type"] = content.MimeType
	}
	if content.URI != "" {
		attachmentData["url"] = content.URI
	}
	if content.Data != "" {
		attachmentData["data"] = content.Data
	}
	if content.Name != "" || content.MimeType != "" {
		metadata := map[string]string{}
		if content.Name != "" {
			metadata["name"] = content.Name
		}
		if content.MimeType != "" {
			metadata["mime_type"] = content.MimeType
		}
		attachmentData["metadata"] = metadata
	}
	return attachmentData
}

// isImageMimeType checks if the MIME type is an image type.
func isImageMimeType(mimeType string) bool {
	imageTypes := []string{
		"image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp",
		"image/bmp", "image/svg+xml", "image/tiff", "image/x-icon",
	}
	for _, t := range imageTypes {
		if mimeType == t {
			return true
		}
	}
	return len(mimeType) > 6 && mimeType[:6] == "image/"
}

// isAudioMimeType checks if the MIME type is an audio type.
func isAudioMimeType(mimeType string) bool {
	audioTypes := []string{
		"audio/mpeg", "audio/mp3", "audio/wav", "audio/ogg", "audio/flac",
		"audio/aac", "audio/webm", "audio/x-wav", "audio/mp4",
	}
	for _, t := range audioTypes {
		if mimeType == t {
			return true
		}
	}
	return len(mimeType) > 6 && mimeType[:6] == "audio/"
}

// isVideoMimeType checks if the MIME type is a video type.
func isVideoMimeType(mimeType string) bool {
	videoTypes := []string{
		"video/mp4", "video/mpeg", "video/webm", "video/ogg", "video/avi",
		"video/quicktime", "video/x-msvideo", "video/x-matroska",
	}
	for _, t := range videoTypes {
		if mimeType == t {
			return true
		}
	}
	return len(mimeType) > 6 && mimeType[:6] == "video/"
}

// extractFormatFromMimeType extracts the format from a MIME type.
// e.g., "audio/mp3" -> "mp3", "video/mp4" -> "mp4"
func extractFormatFromMimeType(mimeType string) string {
	for i := len(mimeType) - 1; i >= 0; i-- {
		if mimeType[i] == '/' {
			return mimeType[i+1:]
		}
	}
	return mimeType
}

// ToMCPToolInfo converts this MCPTool to a core.MCPToolInfo.
func (t *MCPTool) ToMCPToolInfo() core.MCPToolInfo {
	return core.MCPToolInfo{
		Name:        t.name,
		Description: t.description,
		Schema:      t.schema,
		ServerName:  t.serverName,
	}
}

// JSONString returns a JSON representation of the tool for debugging.
func (t *MCPTool) JSONString() string {
	data := map[string]interface{}{
		"name":        t.name,
		"description": t.description,
		"schema":      t.schema,
		"server_name": t.serverName,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("MCPTool{name: %s, server: %s}", t.name, t.serverName)
	}

	return string(jsonData)
}
