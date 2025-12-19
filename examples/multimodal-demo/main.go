package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
)

func main() {
	// Initialize agent configuration
	// You can switch to "ollama" provider and use models like "llava", "llama3.2-vision", or "moondream"
	provider := os.Getenv("LLM_PROVIDER")
	if provider == "" {
		provider = "openai"
	}

	model := os.Getenv("LLM_MODEL")
	if model == "" {
		if provider == "openai" {
			model = "gpt-4o"
		} else if provider == "ollama" {
			model = "qwen3-vl:235b-cloud" // User requested model
		}
	}

	// For Ollama, API key is not required, but we check for OpenAI
	apiKey := os.Getenv("OPENAI_API_KEY")
	if provider == "openai" && apiKey == "" {
		log.Fatal("Please set OPENAI_API_KEY environment variable for OpenAI provider")
	}

	config := &vnext.Config{
		Name: "MultimodalAgent",
		LLM: vnext.LLMConfig{
			Provider:    provider,
			Model:       model,
			APIKey:      apiKey,
			MaxTokens:   500,
			Modalities:  []string{"text", "image", "audio", "video"},
			OutputTypes: []string{"text"},
		},
		SystemPrompt: "You are a helpful assistant that can analyze images, audio, and video content.",
		Timeout:      120 * time.Second,
	}

	// Build the agent
	agent, err := vnext.NewBuilder("MultimodalAgent").
		WithConfig(config).
		Build()
	if err != nil {
		log.Fatalf("Failed to build agent: %v", err)
	}

	ctx := context.Background()

	// Example 1: Analyze an image from URL
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Example 1: Image Analysis from URL")
	fmt.Println(strings.Repeat("=", 60))

	imageURL := "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"
	fmt.Printf("Image: %s\n\n", imageURL)

	result, err := agent.RunWithOptions(ctx, "What is in this image? Describe it briefly.", &vnext.RunOptions{
		Images: []vnext.ImageData{
			{
				URL: imageURL,
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to run agent: %v", err)
	}

	fmt.Printf("Response: %s\n", result.Content)

	// Example 2: Analyze a local image (if provided as arg)
	if len(os.Args) > 1 {
		localPath := os.Args[1]
		ext := strings.ToLower(filepath.Ext(localPath))

		switch ext {
		case ".jpg", ".jpeg", ".png", ".gif", ".webp":
			analyzeLocalImage(ctx, agent, localPath)
		case ".mp3", ".wav", ".ogg", ".flac", ".m4a":
			analyzeLocalAudio(ctx, agent, localPath)
		case ".mp4", ".webm", ".avi", ".mov":
			analyzeLocalVideo(ctx, agent, localPath)
		default:
			fmt.Printf("Unsupported file type: %s\n", ext)
		}
	}

	// Example 3: Multi-image comparison
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Example 3: Multiple Image Comparison")
	fmt.Println(strings.Repeat("=", 60))

	image1 := "https://upload.wikimedia.org/wikipedia/commons/thumb/4/47/PNG_transparency_demonstration_1.png/300px-PNG_transparency_demonstration_1.png"
	image2 := "https://upload.wikimedia.org/wikipedia/commons/thumb/a/a7/Camponotus_flavomarginatus_ant.jpg/320px-Camponotus_flavomarginatus_ant.jpg"

	fmt.Printf("Image 1: %s\n", image1)
	fmt.Printf("Image 2: %s\n\n", image2)

	result, err = agent.RunWithOptions(ctx, "Compare these two images. What are the main differences?", &vnext.RunOptions{
		Images: []vnext.ImageData{
			{URL: image1},
			{URL: image2},
		},
	})
	if err != nil {
		log.Printf("Multi-image comparison failed: %v", err)
	} else {
		fmt.Printf("Response: %s\n", result.Content)
	}

	// Print usage summary
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Usage Summary")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(`
Multimodal Demo - AgenticGoKit

This demo showcases multimodal capabilities:
  - Image analysis from URLs
  - Local image analysis (base64 encoded)
  - Audio transcription/analysis (OpenAI only)
  - Video analysis (OpenAI only)
  - Multi-image comparison

Usage:
  # Basic run (analyzes sample images)
  go run main.go

  # Analyze a local image
  go run main.go path/to/image.jpg

  # Analyze local audio (requires OpenAI provider)
  go run main.go path/to/audio.mp3

  # Analyze local video (requires OpenAI provider)
  go run main.go path/to/video.mp4

Environment Variables:
  LLM_PROVIDER   - openai, ollama, azure, huggingface, openrouter
  LLM_MODEL      - Model name (e.g., gpt-4o, llava)
  OPENAI_API_KEY - Required for OpenAI provider

Note: Audio and video support is currently only available with OpenAI.
      Other providers will display a warning for unsupported modalities.
`)
}

// analyzeLocalImage reads and analyzes a local image file
func analyzeLocalImage(ctx context.Context, agent vnext.Agent, imagePath string) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Example 2: Local Image Analysis")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Image: %s\n\n", imagePath)

	imageData, err := ioutil.ReadFile(imagePath)
	if err != nil {
		log.Printf("Failed to read local image: %v", err)
		return
	}
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	result, err := agent.RunWithOptions(ctx, "What details can you see in this uploaded image?", &vnext.RunOptions{
		Images: []vnext.ImageData{
			{Base64: base64Image},
		},
	})
	if err != nil {
		log.Printf("Failed to analyze local image: %v", err)
		return
	}

	fmt.Printf("Response: %s\n", result.Content)
}

// analyzeLocalAudio reads and analyzes a local audio file
func analyzeLocalAudio(ctx context.Context, agent vnext.Agent, audioPath string) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Audio Analysis")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Audio: %s\n\n", audioPath)

	audioData, err := ioutil.ReadFile(audioPath)
	if err != nil {
		log.Printf("Failed to read audio file: %v", err)
		return
	}
	base64Audio := base64.StdEncoding.EncodeToString(audioData)

	// Determine format from extension
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(audioPath)), ".")

	result, err := agent.RunWithOptions(ctx, "Please transcribe and describe what you hear in this audio.", &vnext.RunOptions{
		Audio: []vnext.AudioData{
			{
				Base64: base64Audio,
				Format: ext,
			},
		},
	})
	if err != nil {
		log.Printf("Failed to analyze audio: %v", err)
		return
	}

	fmt.Printf("Response: %s\n", result.Content)
}

// analyzeLocalVideo reads and analyzes a local video file
func analyzeLocalVideo(ctx context.Context, agent vnext.Agent, videoPath string) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Video Analysis")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Video: %s\n\n", videoPath)

	// Check file size - warn if too large
	fileInfo, err := os.Stat(videoPath)
	if err != nil {
		log.Printf("Failed to stat video file: %v", err)
		return
	}
	if fileInfo.Size() > 20*1024*1024 { // 20MB limit
		log.Printf("Warning: Video file is large (%d MB). This may take a while or fail.", fileInfo.Size()/(1024*1024))
	}

	videoData, err := ioutil.ReadFile(videoPath)
	if err != nil {
		log.Printf("Failed to read video file: %v", err)
		return
	}
	base64Video := base64.StdEncoding.EncodeToString(videoData)

	// Determine format from extension
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(videoPath)), ".")

	result, err := agent.RunWithOptions(ctx, "Please describe what happens in this video. Include key events and details.", &vnext.RunOptions{
		Video: []vnext.VideoData{
			{
				Base64: base64Video,
				Format: ext,
			},
		},
	})
	if err != nil {
		log.Printf("Failed to analyze video: %v", err)
		return
	}

	fmt.Printf("Response: %s\n", result.Content)
}
