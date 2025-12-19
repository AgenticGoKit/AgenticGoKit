package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Unit Tests for Hugging Face Multimodal buildChatRequest (No API Calls)

func TestHuggingFace_BuildChatRequest_Multimodal_TextOnly(t *testing.T) {
	adapter := &HuggingFaceAdapter{
		model:       "test-model",
		temperature: 0.7,
		maxTokens:   100,
	}

	prompt := Prompt{
		System: "You are helpful",
		User:   "Hello",
	}

	request := adapter.buildChatRequest(prompt, 100, 0.7, false)

	messages, ok := request["messages"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, messages, 2)

	assert.Equal(t, "system", messages[0]["role"])
	assert.Equal(t, "user", messages[1]["role"])
}

func TestHuggingFace_BuildChatRequest_Multimodal_WithImageURL(t *testing.T) {
	adapter := &HuggingFaceAdapter{
		model:       "test-model",
		temperature: 0.7,
		maxTokens:   100,
	}

	prompt := Prompt{
		User: "Describe this",
		Images: []ImageData{
			{URL: "https://example.com/image.jpg"},
		},
	}

	request := adapter.buildChatRequest(prompt, 100, 0.7, false)

	messages, ok := request["messages"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, messages, 1)

	content, ok := messages[0]["content"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, content, 2) // text + image

	assert.Equal(t, "text", content[0]["type"])
	assert.Equal(t, "image_url", content[1]["type"])
}

func TestHuggingFace_BuildChatRequest_Multimodal_WithBase64Image(t *testing.T) {
	adapter := &HuggingFaceAdapter{
		model:       "test-model",
		temperature: 0.7,
		maxTokens:   100,
	}

	prompt := Prompt{
		User: "What's this?",
		Images: []ImageData{
			{Base64: "base64data"},
		},
	}

	request := adapter.buildChatRequest(prompt, 100, 0.7, false)

	messages, ok := request["messages"].([]map[string]interface{})
	assert.True(t, ok)

	content, ok := messages[0]["content"].([]map[string]interface{})
	assert.True(t, ok)

	imageContent := content[1]
	assert.Equal(t, "image_url", imageContent["type"])

	imageURL := imageContent["image_url"].(map[string]string)
	url := imageURL["url"]
	assert.Contains(t, url, "data:image/jpeg;base64,")
}

func TestHuggingFace_BuildChatRequest_Multimodal_MultipleImages(t *testing.T) {
	adapter := &HuggingFaceAdapter{
		model:       "test-model",
		temperature: 0.7,
		maxTokens:   100,
	}

	prompt := Prompt{
		User: "Compare",
		Images: []ImageData{
			{URL: "https://example.com/1.jpg"},
			{URL: "https://example.com/2.jpg"},
		},
	}

	request := adapter.buildChatRequest(prompt, 100, 0.7, false)

	messages, ok := request["messages"].([]map[string]interface{})
	assert.True(t, ok)

	content, ok := messages[0]["content"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, content, 3) // 1 text + 2 images
}
