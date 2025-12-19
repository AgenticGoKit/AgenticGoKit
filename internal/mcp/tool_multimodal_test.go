package mcp

import (
	"testing"

	"github.com/kunalkushwaha/mcp-navigator-go/pkg/mcp"
	"github.com/stretchr/testify/assert"
)

func TestIsImageMimeType(t *testing.T) {
	tests := []struct {
		mimeType string
		expected bool
	}{
		{"image/jpeg", true},
		{"image/png", true},
		{"image/gif", true},
		{"image/webp", true},
		{"image/svg+xml", true},
		{"image/custom", true}, // Generic image/* check
		{"audio/mp3", false},
		{"video/mp4", false},
		{"text/plain", false},
		{"application/json", false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			result := isImageMimeType(tt.mimeType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAudioMimeType(t *testing.T) {
	tests := []struct {
		mimeType string
		expected bool
	}{
		{"audio/mpeg", true},
		{"audio/mp3", true},
		{"audio/wav", true},
		{"audio/ogg", true},
		{"audio/flac", true},
		{"audio/custom", true}, // Generic audio/* check
		{"image/jpeg", false},
		{"video/mp4", false},
		{"text/plain", false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			result := isAudioMimeType(tt.mimeType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsVideoMimeType(t *testing.T) {
	tests := []struct {
		mimeType string
		expected bool
	}{
		{"video/mp4", true},
		{"video/webm", true},
		{"video/ogg", true},
		{"video/quicktime", true},
		{"video/custom", true}, // Generic video/* check
		{"image/jpeg", false},
		{"audio/mp3", false},
		{"text/plain", false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			result := isVideoMimeType(tt.mimeType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractFormatFromMimeType(t *testing.T) {
	tests := []struct {
		mimeType string
		expected string
	}{
		{"audio/mp3", "mp3"},
		{"video/mp4", "mp4"},
		{"image/jpeg", "jpeg"},
		{"audio/x-wav", "x-wav"},
		{"plain", "plain"}, // No slash
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			result := extractFormatFromMimeType(tt.mimeType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildImageData(t *testing.T) {
	tool := &MCPTool{name: "test"}

	content := mcp.Content{
		Type:     "image",
		URI:      "https://example.com/image.jpg",
		Data:     "base64imagedata",
		Name:     "test_image.jpg",
		MimeType: "image/jpeg",
	}

	result := tool.buildImageData(content)

	assert.Equal(t, "https://example.com/image.jpg", result["url"])
	assert.Equal(t, "base64imagedata", result["base64"])

	metadata := result["metadata"].(map[string]string)
	assert.Equal(t, "test_image.jpg", metadata["name"])
	assert.Equal(t, "image/jpeg", metadata["mime_type"])
}

func TestBuildAudioData(t *testing.T) {
	tool := &MCPTool{name: "test"}

	content := mcp.Content{
		Type:     "audio",
		URI:      "https://example.com/audio.mp3",
		Data:     "base64audiodata",
		Name:     "test_audio.mp3",
		MimeType: "audio/mp3",
	}

	result := tool.buildAudioData(content)

	assert.Equal(t, "https://example.com/audio.mp3", result["url"])
	assert.Equal(t, "base64audiodata", result["base64"])
	assert.Equal(t, "mp3", result["format"])

	metadata := result["metadata"].(map[string]string)
	assert.Equal(t, "test_audio.mp3", metadata["name"])
}

func TestBuildVideoData(t *testing.T) {
	tool := &MCPTool{name: "test"}

	content := mcp.Content{
		Type:     "video",
		URI:      "https://example.com/video.mp4",
		Data:     "base64videodata",
		Name:     "test_video.mp4",
		MimeType: "video/mp4",
	}

	result := tool.buildVideoData(content)

	assert.Equal(t, "https://example.com/video.mp4", result["url"])
	assert.Equal(t, "base64videodata", result["base64"])
	assert.Equal(t, "mp4", result["format"])

	metadata := result["metadata"].(map[string]string)
	assert.Equal(t, "test_video.mp4", metadata["name"])
}

func TestBuildAttachmentData(t *testing.T) {
	tool := &MCPTool{name: "test"}

	content := mcp.Content{
		Type:     "file",
		URI:      "https://example.com/doc.pdf",
		Data:     "base64pdfdata",
		Name:     "document.pdf",
		MimeType: "application/pdf",
	}

	result := tool.buildAttachmentData(content)

	assert.Equal(t, "document.pdf", result["name"])
	assert.Equal(t, "application/pdf", result["type"])
	assert.Equal(t, "https://example.com/doc.pdf", result["url"])
	assert.Equal(t, "base64pdfdata", result["data"])
}

func TestConvertMCPResponseToAgentFlow_Multimodal(t *testing.T) {
	tool := &MCPTool{
		name:       "test_tool",
		serverName: "test_server",
	}

	response := &mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: "Here is the analysis",
			},
			{
				Type:     "image",
				URI:      "https://example.com/result.jpg",
				MimeType: "image/jpeg",
			},
			{
				Type:     "audio",
				Data:     "base64audiodata",
				MimeType: "audio/mp3",
			},
			{
				Type:     "video",
				URI:      "https://example.com/result.mp4",
				MimeType: "video/mp4",
			},
		},
		IsError: false,
	}

	result, err := tool.convertMCPResponseToAgentFlow(response)

	assert.NoError(t, err)
	assert.Equal(t, "Here is the analysis", result["text"])
	assert.Equal(t, "test_tool", result["tool_name"])
	assert.Equal(t, "test_server", result["server_name"])
	assert.True(t, result["success"].(bool))

	// Check images
	images := result["images"].([]map[string]interface{})
	assert.Len(t, images, 1)
	assert.Equal(t, "https://example.com/result.jpg", images[0]["url"])

	// Check audio
	audioFiles := result["audio"].([]map[string]interface{})
	assert.Len(t, audioFiles, 1)
	assert.Equal(t, "base64audiodata", audioFiles[0]["base64"])

	// Check video
	videoFiles := result["video"].([]map[string]interface{})
	assert.Len(t, videoFiles, 1)
	assert.Equal(t, "https://example.com/result.mp4", videoFiles[0]["url"])
}

func TestConvertMCPResponseToAgentFlow_MimeTypeDetection(t *testing.T) {
	tool := &MCPTool{
		name:       "test_tool",
		serverName: "test_server",
	}

	// Content without explicit type, relying on MIME type detection
	response := &mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type:     "resource",
				Data:     "imagedata",
				MimeType: "image/png",
			},
			{
				Type:     "resource",
				Data:     "audiodata",
				MimeType: "audio/wav",
			},
			{
				Type:     "resource",
				URI:      "https://example.com/video.webm",
				MimeType: "video/webm",
			},
		},
		IsError: false,
	}

	result, err := tool.convertMCPResponseToAgentFlow(response)

	assert.NoError(t, err)

	// Check that MIME type detection worked
	images := result["images"].([]map[string]interface{})
	assert.Len(t, images, 1)

	audioFiles := result["audio"].([]map[string]interface{})
	assert.Len(t, audioFiles, 1)

	videoFiles := result["video"].([]map[string]interface{})
	assert.Len(t, videoFiles, 1)
}

func TestConvertMCPResponseToAgentFlow_TextOnly(t *testing.T) {
	tool := &MCPTool{
		name:       "test_tool",
		serverName: "test_server",
	}

	response := &mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: "Simple text response",
			},
		},
		IsError: false,
	}

	result, err := tool.convertMCPResponseToAgentFlow(response)

	assert.NoError(t, err)
	assert.Equal(t, "Simple text response", result["text"])

	// Should not have multimodal fields
	_, hasImages := result["images"]
	_, hasAudio := result["audio"]
	_, hasVideo := result["video"]

	assert.False(t, hasImages)
	assert.False(t, hasAudio)
	assert.False(t, hasVideo)
}

func TestConvertMCPResponseToAgentFlow_MultipleTextParts(t *testing.T) {
	tool := &MCPTool{
		name:       "test_tool",
		serverName: "test_server",
	}

	response := &mcp.CallToolResponse{
		Content: []mcp.Content{
			{Type: "text", Text: "First part"},
			{Type: "text", Text: "Second part"},
			{Type: "text", Text: "Third part"},
		},
		IsError: false,
	}

	result, err := tool.convertMCPResponseToAgentFlow(response)

	assert.NoError(t, err)
	assert.Equal(t, "First part", result["text"])

	allText := result["all_text"].([]string)
	assert.Len(t, allText, 3)
	assert.Equal(t, []string{"First part", "Second part", "Third part"}, allText)
}
