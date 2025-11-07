package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/kunalkushwaha/agenticgokit/core"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/llm/openrouter"
)

func main() {
	// Check for API key in environment
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENROUTER_API_KEY environment variable not set. Please set it with your OpenRouter API key.")
	}

	// Example 1: Basic usage with API key from environment
	fmt.Println("Example 1: Basic OpenRouter Usage")
	fmt.Println("====================================")

	config := core.LLMProviderConfig{
		Type:        "openrouter",
		APIKey:      apiKey, // Read from OPENROUTER_API_KEY environment variable
		Model:       "openai/gpt-3.5-turbo",
		MaxTokens:   500,
		Temperature: 0.7,
	}

	provider, err := core.NewModelProviderFromConfig(config)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	ctx := context.Background()
	response, err := provider.Call(ctx, core.Prompt{
		System: "You are a helpful assistant.",
		User:   "What is OpenRouter?",
	})

	if err != nil {
		log.Fatalf("Call failed: %v", err)
	}

	fmt.Printf("Response: %s\n", response.Content)
	fmt.Printf("Tokens used: %d (prompt: %d, completion: %d)\n\n",
		response.Usage.TotalTokens,
		response.Usage.PromptTokens,
		response.Usage.CompletionTokens)

	// Example 2: Using different models
	fmt.Println("Example 2: Using Different Models")
	fmt.Println("====================================")

	models := []string{
		"openai/gpt-3.5-turbo",
		"anthropic/claude-3-haiku",
		"google/gemini-2.0-flash-exp:free", // Updated to available free model
		"meta-llama/llama-3.1-8b-instruct",
	}

	for _, model := range models {
		modelConfig := core.LLMProviderConfig{
			Type:        "openrouter",
			APIKey:      apiKey,
			Model:       model,
			MaxTokens:   100,
			Temperature: 0.7,
		}

		modelProvider, err := core.NewModelProviderFromConfig(modelConfig)
		if err != nil {
			log.Printf("Failed to create provider for %s: %v", model, err)
			continue
		}

		modelResponse, err := modelProvider.Call(ctx, core.Prompt{
			User: "Say hello in one sentence.",
		})

		if err != nil {
			log.Printf("Call to %s failed: %v", model, err)
			continue
		}

		fmt.Printf("Model: %s\n", model)
		fmt.Printf("Response: %s\n", modelResponse.Content)
		fmt.Printf("Tokens: %d\n\n", modelResponse.Usage.TotalTokens)
	}

	// Example 3: Streaming responses
	fmt.Println("Example 3: Streaming Responses")
	fmt.Println("====================================")

	streamConfig := core.LLMProviderConfig{
		Type:        "openrouter",
		APIKey:      apiKey,
		Model:       "openai/gpt-3.5-turbo",
		MaxTokens:   200,
		Temperature: 0.8,
	}

	streamProvider, err := core.NewModelProviderFromConfig(streamConfig)
	if err != nil {
		log.Fatalf("Failed to create stream provider: %v", err)
	}

	fmt.Print("Streaming response: ")
	tokenChan, err := streamProvider.Stream(ctx, core.Prompt{
		User: "Write a haiku about coding.",
	})

	if err != nil {
		log.Fatalf("Stream failed: %v", err)
	}

	for token := range tokenChan {
		if token.Error != nil {
			log.Fatalf("Stream error: %v", token.Error)
		}
		fmt.Print(token.Content)
	}
	fmt.Println("\n")

	// Example 4: Using AgentLLMConfig (automatically reads OPENROUTER_API_KEY)
	fmt.Println("Example 4: Using AgentLLMConfig")
	fmt.Println("====================================")
	fmt.Println("AgentLLMConfig automatically reads OPENROUTER_API_KEY environment variable")

	envConfig := core.AgentLLMConfig{
		Provider:    "openrouter",
		Model:       "anthropic/claude-3-haiku",
		MaxTokens:   300,
		Temperature: 0.7,
	}

	envProvider, err := core.NewLLMProvider(envConfig)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	envResponse, err := envProvider.Call(ctx, core.Prompt{
		User: "What are the benefits of AI?",
	})

	if err != nil {
		log.Fatalf("Call failed: %v", err)
	}

	fmt.Printf("Response: %s\n", envResponse.Content)
	fmt.Println()

	// Example 5: Using site tracking for OpenRouter rankings
	fmt.Println("Example 5: Site Tracking")
	fmt.Println("====================================")

	trackingConfig := core.LLMProviderConfig{
		Type:        "openrouter",
		APIKey:      apiKey,
		Model:       "openai/gpt-3.5-turbo",
		MaxTokens:   200,
		Temperature: 0.7,
		SiteURL:     "https://myapp.com",
		SiteName:    "My Awesome App",
	}

	trackingProvider, err := core.NewModelProviderFromConfig(trackingConfig)
	if err != nil {
		log.Fatalf("Failed to create provider with tracking: %v", err)
	}

	fmt.Println("Provider created with site tracking enabled")
	fmt.Println("HTTP-Referer: https://myapp.com")
	fmt.Println("X-Title: My Awesome App")
	fmt.Println("These headers help with OpenRouter rankings and analytics\n")

	trackingResponse, err := trackingProvider.Call(ctx, core.Prompt{
		User: "Explain OpenRouter in one sentence.",
	})

	if err != nil {
		log.Printf("Call failed: %v", err)
	} else {
		fmt.Printf("Response: %s\n", trackingResponse.Content)
	}

	// Example 6: Parameter overrides
	fmt.Println("\nExample 6: Parameter Overrides")
	fmt.Println("====================================")

	// Provider defaults: MaxTokens=500, Temperature=0.7
	paramProvider, err := core.NewModelProviderFromConfig(core.LLMProviderConfig{
		Type:        "openrouter",
		APIKey:      apiKey,
		Model:       "openai/gpt-3.5-turbo",
		MaxTokens:   500,
		Temperature: 0.7,
	})

	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// Override for this specific call
	maxTokens := int32(50)
	temperature := float32(0.3)

	paramResponse, err := paramProvider.Call(ctx, core.Prompt{
		User: "Count from 1 to 5.",
		Parameters: core.ModelParameters{
			MaxTokens:   &maxTokens,
			Temperature: &temperature,
		},
	})

	if err != nil {
		log.Printf("Call failed: %v", err)
	} else {
		fmt.Printf("Response with overridden parameters: %s\n", paramResponse.Content)
		fmt.Printf("Used temperature=0.3 and maxTokens=50 instead of defaults\n")
	}
}
