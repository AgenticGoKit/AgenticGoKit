package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Unit Tests for Azure OpenAI Multimodal Message Building (No API Calls)

func TestMapInternalPrompt_Multimodal_TextOnly(t *testing.T) {
	prompt := Prompt{
		System: "You are helpful",
		User:   "Hello",
	}

	messages := mapInternalPrompt(prompt)

	assert.Len(t, messages, 2)
	assert.Equal(t, "system", messages[0].Role)
	assert.Equal(t, "user", messages[1].Role)
	assert.Equal(t, "Hello", messages[1].Content)
}

func TestMapInternalPrompt_Multimodal_WithImageURL(t *testing.T) {
	prompt := Prompt{
		User: "Describe this image",
		Images: []ImageData{
			{URL: "https://example.com/image.jpg"},
		},
	}

	messages := mapInternalPrompt(prompt)

	assert.Len(t, messages, 1)
	assert.Equal(t, "user", messages[0].Role)

	content, ok := messages[0].Content.([]map[string]interface{})
	assert.True(t, ok, "content should be an array")
	assert.Len(t, content, 2) // text + image

	assert.Equal(t, "text", content[0]["type"])
	assert.Equal(t, "image_url", content[1]["type"])

	imageURL := content[1]["image_url"].(map[string]interface{})
	assert.Equal(t, "https://example.com/image.jpg", imageURL["url"])
}

func TestMapInternalPrompt_Multimodal_WithBase64Image(t *testing.T) {
	prompt := Prompt{
		User: "What's this?",
		Images: []ImageData{
			{Base64: "base64data"},
		},
	}

	messages := mapInternalPrompt(prompt)

	content, ok := messages[0].Content.([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, content, 2)

	assert.Equal(t, "image_url", content[1]["type"])
	imageURL := content[1]["image_url"].(map[string]interface{})
	url := imageURL["url"].(string)
	assert.Contains(t, url, "data:image/jpeg;base64,")
}

func TestMapInternalPrompt_Multimodal_MultipleImages(t *testing.T) {
	prompt := Prompt{
		User: "Compare these",
		Images: []ImageData{
			{URL: "https://example.com/1.jpg"},
			{URL: "https://example.com/2.jpg"},
		},
	}

	messages := mapInternalPrompt(prompt)

	content, ok := messages[0].Content.([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, content, 3) // 1 text + 2 images

	assert.Equal(t, "text", content[0]["type"])
	assert.Equal(t, "image_url", content[1]["type"])
	assert.Equal(t, "image_url", content[2]["type"])
}

func TestMapInternalPrompt_Multimodal_WithAudio(t *testing.T) {
	prompt := Prompt{
		User:   "Transcribe this audio",
		Images: []ImageData{}, // Need at least empty to trigger multimodal path
		Audio: []AudioData{
			{Base64: "audiobase64data", Format: "mp3"},
		},
	}

	// Force multimodal mode by adding an empty image first
	prompt.Images = append(prompt.Images, ImageData{URL: "https://example.com/img.jpg"})

	messages := mapInternalPrompt(prompt)

	content, ok := messages[0].Content.([]map[string]interface{})
	assert.True(t, ok)
	// Should have text + image + audio
	assert.GreaterOrEqual(t, len(content), 3)

	// Find audio content
	var foundAudio bool
	for _, c := range content {
		if c["type"] == "input_audio" {
			foundAudio = true
			audioData := c["input_audio"].(map[string]interface{})
			assert.Equal(t, "audiobase64data", audioData["data"])
			assert.Equal(t, "mp3", audioData["format"])
		}
	}
	assert.True(t, foundAudio, "should have audio content")
}

func TestMapInternalPrompt_Multimodal_WithVideo(t *testing.T) {
	prompt := Prompt{
		User: "Describe this video",
		Images: []ImageData{
			{URL: "https://example.com/img.jpg"},
		},
		Video: []VideoData{
			{URL: "https://example.com/video.mp4"},
		},
	}

	messages := mapInternalPrompt(prompt)

	content, ok := messages[0].Content.([]map[string]interface{})
	assert.True(t, ok)
	// Should have text + image + video
	assert.GreaterOrEqual(t, len(content), 3)

	// Find video content
	var foundVideo bool
	for _, c := range content {
		if c["type"] == "input_video" {
			foundVideo = true
			videoData := c["input_video"].(map[string]interface{})
			assert.Equal(t, "https://example.com/video.mp4", videoData["url"])
		}
	}
	assert.True(t, foundVideo, "should have video content")
}

func TestMapInternalPrompt_Multimodal_WithVideoBase64(t *testing.T) {
	prompt := Prompt{
		User: "Analyze this video",
		Images: []ImageData{
			{URL: "https://example.com/img.jpg"},
		},
		Video: []VideoData{
			{Base64: "videobase64data", Format: "mp4"},
		},
	}

	messages := mapInternalPrompt(prompt)

	content, ok := messages[0].Content.([]map[string]interface{})
	assert.True(t, ok)

	// Find video content
	var foundVideo bool
	for _, c := range content {
		if c["type"] == "input_video" {
			foundVideo = true
			videoData := c["input_video"].(map[string]interface{})
			url := videoData["url"].(string)
			assert.Contains(t, url, "data:video/mp4;base64,")
		}
	}
	assert.True(t, foundVideo, "should have video content")
}

func TestMapInternalPrompt_Multimodal_AllTypes(t *testing.T) {
	prompt := Prompt{
		User: "Process all these inputs",
		Images: []ImageData{
			{URL: "https://example.com/image.jpg"},
		},
		Audio: []AudioData{
			{Base64: "audiodata", Format: "wav"},
		},
		Video: []VideoData{
			{URL: "https://example.com/video.mp4"},
		},
	}

	messages := mapInternalPrompt(prompt)

	content, ok := messages[0].Content.([]map[string]interface{})
	assert.True(t, ok)
	// Should have text + image + audio + video = 4
	assert.Equal(t, 4, len(content))

	// Verify all types are present
	types := make(map[string]bool)
	for _, c := range content {
		types[c["type"].(string)] = true
	}

	assert.True(t, types["text"], "should have text")
	assert.True(t, types["image_url"], "should have image")
	assert.True(t, types["input_audio"], "should have audio")
	assert.True(t, types["input_video"], "should have video")
}
