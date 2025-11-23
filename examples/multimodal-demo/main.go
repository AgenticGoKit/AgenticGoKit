package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
		Name: "VisionAgent",
		LLM: vnext.LLMConfig{
			Provider:  provider,
			Model:     model,
			APIKey:    apiKey,
			MaxTokens: 300,
		},
		SystemPrompt: "You are a helpful assistant that can analyze images.",
		Timeout:      60 * time.Second,
	}

	// Build the agent
	agent, err := vnext.NewBuilder("VisionAgent").
		WithConfig(config).
		Build()
	if err != nil {
		log.Fatalf("Failed to build agent: %v", err)
	}

	ctx := context.Background()

	// Example 1: Analyze an image from URL
	imageURL := "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"
	fmt.Printf("\n--- Analyzing Image from URL ---\n")
	fmt.Printf("Image: %s\n", imageURL)
	
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
		localImagePath := os.Args[1]
		fmt.Printf("\n--- Analyzing Local Image ---\n")
		fmt.Printf("Image: %s\n", localImagePath)

		// Read and encode image
		imageData, err := ioutil.ReadFile(localImagePath)
		if err != nil {
			log.Fatalf("Failed to read local image: %v", err)
		}
		base64Image := base64.StdEncoding.EncodeToString(imageData)

		// Construct data URL or just pass base64 depending on provider support
		// For OpenAI, we can pass base64 directly in the adapter logic we wrote, 
		// but let's pass it as a data URL to be safe and explicit if the adapter expects it.
		// Our adapter logic handles raw base64 by checking for "data:" prefix.
		
		result, err = agent.RunWithOptions(ctx, "What details can you see in this uploaded image?", &vnext.RunOptions{
			Images: []vnext.ImageData{
				{
					Base64: base64Image, // Adapter handles adding data:image/... prefix if missing
				},
			},
		})
		if err != nil {
			log.Fatalf("Failed to run agent with local image: %v", err)
		}

		fmt.Printf("Response: %s\n", result.Content)
	}
}
