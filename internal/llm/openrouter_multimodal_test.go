package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Unit Tests for OpenRouter Multimodal Message Building (No API Calls)

func TestBuildOpenRouterMessages_Multimodal_TextOnly(t *testing.T) {
	prompt := Prompt{
		System: "You are helpful",
		User:   "Hello",
	}

	messages := buildOpenRouterMessages(prompt)

	assert.Len(t, messages, 2)
	assert.Equal(t, "system", messages[0]["role"])
	assert.Equal(t, "user", messages[1]["role"])
	assert.Equal(t, "Hello", messages[1]["content"])
}

func TestBuildOpenRouterMessages_Multimodal_WithImageURL(t *testing.T) {
	prompt := Prompt{
		User: "Describe this image",
		Images: []ImageData{
			{URL: "https://example.com/image.jpg"},
		},
	}

	messages := buildOpenRouterMessages(prompt)

	assert.Len(t, messages, 1)
	assert.Equal(t, "user", messages[0]["role"])

	content, ok := messages[0]["content"].([]map[string]interface{})
	assert.True(t, ok, "content should be an array")
	assert.Len(t, content, 2) // text + image

	assert.Equal(t, "text", content[0]["type"])
	assert.Equal(t, "image_url", content[1]["type"])
	
	imageURL := content[1]["image_url"].(map[string]interface{})
	assert.Equal(t, "https://example.com/image.jpg", imageURL["url"])
}

func TestBuildOpenRouterMessages_Multimodal_WithBase64Image(t *testing.T) {
	prompt := Prompt{
		User: "What's this?",
		Images: []ImageData{
			{Base64: "base64data"},
		},
	}

	messages := buildOpenRouterMessages(prompt)

	content, ok := messages[0]["content"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, content, 2)

	assert.Equal(t, "image_url", content[1]["type"])
	imageURL := content[1]["image_url"].(map[string]interface{})
	url := imageURL["url"].(string)
	assert.Contains(t, url, "data:image/jpeg;base64,")
}

func TestBuildOpenRouterMessages_Multimodal_MultipleImages(t *testing.T) {
	prompt := Prompt{
		User: "Compare these",
		Images: []ImageData{
			{URL: "https://example.com/1.jpg"},
			{URL: "https://example.com/2.jpg"},
		},
	}

	messages := buildOpenRouterMessages(prompt)

	content, ok := messages[0]["content"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, content, 3) // 1 text + 2 images
	
	assert.Equal(t, "text", content[0]["type"])
	assert.Equal(t, "image_url", content[1]["type"])
	assert.Equal(t, "image_url", content[2]["type"])
}

func TestBuildOpenRouterMessages_Multimodal_ImagesOnly(t *testing.T) {
	prompt := Prompt{
		Images: []ImageData{
			{URL: "https://example.com/image.jpg"},
		},
	}

	messages := buildOpenRouterMessages(prompt)

	assert.Len(t, messages, 1)
	content, ok := messages[0]["content"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, content, 1) // Only image, no text
}
