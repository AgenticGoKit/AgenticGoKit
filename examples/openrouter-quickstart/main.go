package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"
	_ "github.com/agenticgokit/agenticgokit/plugins/llm/openrouter"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("  OpenRouter QuickStart - vNext API")
	fmt.Println("===========================================")
	fmt.Println()

	// Check for API key in environment
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENROUTER_API_KEY environment variable not set. Please set it with your OpenRouter API key.")
	}

	// Initialize vNext with defaults (optional but recommended)
	if err := vnext.InitializeDefaults(); err != nil {
		log.Fatalf("Failed to initialize vNext: %v", err)
	}

	ctx := context.Background()

	// Example 1: Basic Usage with Config
	fmt.Println("Example 1: Basic Agent with Config")
	fmt.Println("====================================")

	config1 := &vnext.Config{
		Name:         "openrouter-assistant",
		SystemPrompt: "You are a helpful assistant.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "openrouter",
			Model:       "openai/gpt-3.5-turbo",
			APIKey:      apiKey, // Pass the API key from environment
			Temperature: 0.7,
			MaxTokens:   500,
		},
	}

	agent1, err := vnext.NewBuilder("openrouter-assistant").
		WithConfig(config1).
		Build()

	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	if err := agent1.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize agent: %v", err)
	}
	defer agent1.Cleanup(ctx)

	result1, err := agent1.Run(ctx, "What is OpenRouter?")
	if err != nil {
		log.Fatalf("Run failed: %v", err)
	}

	fmt.Printf("Response: %s\n", result1.Content)
	fmt.Printf("Duration: %v | Tokens: %d\n\n", result1.Duration, result1.TokensUsed)

	// Example 2: Using Preset Chat Agent with Options
	fmt.Println("Example 2: Preset Chat Agent")
	fmt.Println("==============================")

	// Create a config first with API key
	chatConfig := &vnext.Config{
		Name:         "chat-bot",
		SystemPrompt: "You are a conversational assistant focused on providing helpful and friendly responses",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "openrouter",
			Model:       "anthropic/claude-3-haiku",
			APIKey:      apiKey, // Pass the API key
			Temperature: 0.7,
			MaxTokens:   100,
		},
	}

	chatAgent, err := vnext.NewBuilder("chat-bot").
		WithConfig(chatConfig).
		Build()

	if err != nil {
		log.Fatalf("Failed to create chat agent: %v", err)
	}

	if err := chatAgent.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize chat agent: %v", err)
	}
	defer chatAgent.Cleanup(ctx)

	result2, err := chatAgent.Run(ctx, "Say hello in one sentence.")
	if err != nil {
		log.Fatalf("Run failed: %v", err)
	}

	fmt.Printf("Response: %s\n", result2.Content)
	fmt.Printf("Duration: %v\n\n", result2.Duration)

	// Example 3: Streaming Responses
	fmt.Println("Example 3: Streaming Responses")
	fmt.Println("================================")

	config3 := &vnext.Config{
		Name:         "streaming-agent",
		SystemPrompt: "You are a creative writing assistant.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "openrouter",
			Model:       "openai/gpt-3.5-turbo",
			APIKey:      apiKey, // Pass the API key
			Temperature: 0.8,
			MaxTokens:   200,
		},
	}

	streamAgent, err := vnext.NewBuilder("streaming-agent").
		WithConfig(config3).
		Build()

	if err != nil {
		log.Fatalf("Failed to create streaming agent: %v", err)
	}

	if err := streamAgent.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize streaming agent: %v", err)
	}
	defer streamAgent.Cleanup(ctx)

	fmt.Print("Streaming response: ")
	stream, err := streamAgent.RunStream(ctx, "Write a haiku about coding.",
		vnext.WithBufferSize(10),
		vnext.WithThoughts(),
	)

	if err != nil {
		log.Fatalf("Stream failed: %v", err)
	}

	for chunk := range stream.Chunks() {
		if chunk.Error != nil {
			log.Fatalf("Stream error: %v", chunk.Error)
		}
		if chunk.Type == vnext.ChunkTypeDelta {
			fmt.Print(chunk.Delta)
		}
	}

	streamResult, err := stream.Wait()
	if err != nil {
		log.Fatalf("Stream wait failed: %v", err)
	}

	fmt.Printf("\n\nDuration: %v | Success: %v\n\n", streamResult.Duration, streamResult.Success)

	// Example 4: Using Different Models
	fmt.Println("Example 4: Testing Multiple Models")
	fmt.Println("====================================")

	models := []struct {
		name  string
		model string
	}{
		{"OpenAI GPT-3.5", "openai/gpt-3.5-turbo"},
		{"Claude 3 Haiku", "anthropic/claude-3-haiku"},
		{"Gemini 2.0 Flash", "google/gemini-2.0-flash-exp:free"},
		{"Llama 3.1", "meta-llama/llama-3.1-8b-instruct"},
	}

	for _, m := range models {
		modelConfig := &vnext.Config{
			Name:         m.name,
			SystemPrompt: "You are a helpful assistant.",
			Timeout:      30 * time.Second,
			LLM: vnext.LLMConfig{
				Provider:    "openrouter",
				Model:       m.model,
				APIKey:      apiKey, // Pass the API key
				Temperature: 0.7,
				MaxTokens:   100,
			},
		}

		modelAgent, err := vnext.NewBuilder(m.name).
			WithConfig(modelConfig).
			Build()

		if err != nil {
			log.Printf("Failed to create agent for %s: %v", m.name, err)
			continue
		}

		if err := modelAgent.Initialize(ctx); err != nil {
			log.Printf("Failed to initialize agent for %s: %v", m.name, err)
			continue
		}

		result, err := modelAgent.Run(ctx, "What is 2+2?")
		if err != nil {
			log.Printf("Call to %s failed: %v", m.name, err)
			modelAgent.Cleanup(ctx)
			continue
		}

		fmt.Printf("\nModel: %s\n", m.name)
		fmt.Printf("Response: %s\n", result.Content)
		fmt.Printf("Duration: %v\n", result.Duration)

		modelAgent.Cleanup(ctx)
	}

	// Example 5: Using RunOptions for Detailed Results
	fmt.Println("\nExample 5: Detailed Results with RunOptions")
	fmt.Println("============================================")

	config5 := &vnext.Config{
		Name:         "detail-agent",
		SystemPrompt: "You are a helpful assistant.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "openrouter",
			Model:       "openai/gpt-3.5-turbo",
			APIKey:      apiKey, // Pass the API key
			Temperature: 0.7,
			MaxTokens:   300,
		},
	}

	detailAgent, err := vnext.NewBuilder("detail-agent").
		WithConfig(config5).
		Build()

	if err != nil {
		log.Fatalf("Failed to create detail agent: %v", err)
	}

	if err := detailAgent.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize detail agent: %v", err)
	}
	defer detailAgent.Cleanup(ctx)

	// Use RunOptions for detailed execution information
	opts := vnext.RunWithDetailedResult().
		SetTimeout(30*time.Second).
		AddContext("request_id", "demo-123")

	detailResult, err := detailAgent.RunWithOptions(ctx, "Explain AI in one sentence.", opts)
	if err != nil {
		log.Fatalf("Run with options failed: %v", err)
	}

	fmt.Printf("Response: %s\n", detailResult.Content)
	fmt.Printf("Duration: %v\n", detailResult.Duration)
	fmt.Printf("Success: %v\n", detailResult.Success)
	fmt.Printf("Tokens Used: %d\n", detailResult.TokensUsed)
	fmt.Printf("Metadata: %v\n", detailResult.Metadata)

	// Example 6: Site Tracking Configuration
	fmt.Println("\nExample 6: Site Tracking for OpenRouter Rankings")
	fmt.Println("=================================================")

	// Create a custom config with site tracking
	trackingConfig := &vnext.Config{
		Name:         "tracked-agent",
		SystemPrompt: "You are a helpful assistant.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "openrouter",
			Model:       "openai/gpt-3.5-turbo",
			APIKey:      apiKey, // Pass the API key
			Temperature: 0.7,
			MaxTokens:   200,
			// These fields enable site tracking for OpenRouter rankings
			SiteURL:  "https://myapp.com",
			SiteName: "My Awesome App",
		},
	}

	trackingAgent, err := vnext.NewBuilder("tracked-agent").
		WithConfig(trackingConfig).
		Build()

	if err != nil {
		log.Fatalf("Failed to create tracking agent: %v", err)
	}

	if err := trackingAgent.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize tracking agent: %v", err)
	}
	defer trackingAgent.Cleanup(ctx)

	fmt.Println("Agent created with site tracking enabled")
	fmt.Println("HTTP-Referer: https://myapp.com")
	fmt.Println("X-Title: My Awesome App")
	fmt.Println("These headers help with OpenRouter rankings and analytics")
	fmt.Println()

	trackingResult, err := trackingAgent.Run(ctx, "What is machine learning?")
	if err != nil {
		log.Printf("Call failed: %v", err)
	} else {
		fmt.Printf("Response: %s\n", trackingResult.Content)
		fmt.Printf("Duration: %v\n", trackingResult.Duration)
	}

	// Example 7: Using Custom Handler with OpenRouter
	fmt.Println("\nExample 7: Custom Handler")
	fmt.Println("==========================")

	customHandler := func(ctx context.Context, input string, caps *vnext.Capabilities) (string, error) {
		// Custom logic: check for specific keywords
		if len(input) < 10 {
			// For short queries, add context via LLM
			return caps.LLM(
				"You are a helpful assistant that expands short queries into detailed questions.",
				fmt.Sprintf("Expand this query: %s", input),
			)
		}

		// For longer queries, process normally
		return caps.LLM("You are a helpful assistant.", input)
	}

	config7 := &vnext.Config{
		Name:         "custom-agent",
		SystemPrompt: "You are a helpful assistant.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "openrouter",
			Model:       "openai/gpt-3.5-turbo",
			APIKey:      apiKey, // Pass the API key
			Temperature: 0.7,
			MaxTokens:   200,
		},
	}

	customAgent, err := vnext.NewBuilder("custom-agent").
		WithConfig(config7).
		WithHandler(customHandler).
		Build()

	if err != nil {
		log.Fatalf("Failed to create custom agent: %v", err)
	}

	if err := customAgent.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize custom agent: %v", err)
	}
	defer customAgent.Cleanup(ctx)

	// Test with short input
	shortResult, err := customAgent.Run(ctx, "AI?")
	if err != nil {
		log.Printf("Short query failed: %v", err)
	} else {
		fmt.Printf("Short Query Response: %s\n", shortResult.Content)
	}

	// Example 8: Using Preset with Custom Prompt
	fmt.Println("\nExample 8: Using Preset with Custom Prompt")
	fmt.Println("===========================================")

	// Create config for a simple agent
	config8 := &vnext.Config{
		Name:         "simple-agent",
		SystemPrompt: "You are a helpful assistant specialized in programming languages.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "openrouter",
			Model:       "openai/gpt-3.5-turbo",
			APIKey:      apiKey, // Pass the API key
			Temperature: 0.7,
			MaxTokens:   200,
		},
	}

	agent8, err := vnext.NewBuilder("simple-agent").
		WithConfig(config8).
		Build()

	if err != nil {
		log.Printf("Failed to create agent8: %v", err)
	} else if err := agent8.Initialize(ctx); err != nil {
		log.Printf("Failed to initialize agent8: %v", err)
	} else {
		defer agent8.Cleanup(ctx)
		result, err := agent8.Run(ctx, "What is Go programming language?")
		if err != nil {
			log.Printf("Run failed: %v", err)
		} else {
			fmt.Printf("Response: %s\n", result.Content)
			fmt.Printf("Duration: %v\n", result.Duration)
		}
	}

	fmt.Println("\n===========================================")
	fmt.Println("  OpenRouter vNext examples completed!")
	fmt.Println("===========================================")
}



