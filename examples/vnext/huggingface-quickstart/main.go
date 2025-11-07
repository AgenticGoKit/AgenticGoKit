package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/llm/huggingface"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("  HuggingFace QuickStart - vNext API")
	fmt.Println("===========================================")
	fmt.Println()

	// Check for API key in environment
	apiKey := os.Getenv("HUGGINGFACE_API_KEY")
	if apiKey == "" {
		log.Fatal("HUGGINGFACE_API_KEY environment variable not set. Please set it with your HuggingFace API key.")
	}

	// Initialize vNext with defaults (optional but recommended)
	if err := vnext.InitializeDefaults(); err != nil {
		log.Fatalf("Failed to initialize vNext: %v", err)
	}

	ctx := context.Background()

	// Example 1: Basic Usage with New Router API
	fmt.Println("Example 1: Basic Agent with New Router API")
	fmt.Println("============================================")
	fmt.Println("Using router.huggingface.co with OpenAI-compatible format")
	fmt.Println()

	config1 := &vnext.Config{
		Name:         "hf-assistant",
		SystemPrompt: "You are a helpful assistant.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "huggingface",
			Model:       "meta-llama/Llama-3.2-1B-Instruct", // New router-compatible model
			APIKey:      apiKey,
			Temperature: 0.7,
			MaxTokens:   500,
		},
	}

	agent1, err := vnext.NewBuilder("hf-assistant").
		WithConfig(config1).
		Build()

	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	if err := agent1.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize agent: %v", err)
	}
	defer agent1.Cleanup(ctx)

	result1, err := agent1.Run(ctx, "What is artificial intelligence?")
	if err != nil {
		log.Fatalf("Run failed: %v", err)
	}

	fmt.Printf("Response: %s\n", result1.Content)
	fmt.Printf("Duration: %v | Tokens: %d\n\n", result1.Duration, result1.TokensUsed)

	// Example 2: Using Larger Llama Model
	fmt.Println("Example 2: Using Larger Llama Model")
	fmt.Println("=====================================")

	config2 := &vnext.Config{
		Name:         "llama-3b-agent",
		SystemPrompt: "You are a knowledgeable AI assistant specialized in technology.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "huggingface",
			Model:       "meta-llama/Llama-3.2-3B-Instruct", // Larger model for better quality
			APIKey:      apiKey,
			Temperature: 0.7,
			MaxTokens:   200,
		},
	}

	agent2, err := vnext.NewBuilder("llama-3b-agent").
		WithConfig(config2).
		Build()

	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	if err := agent2.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize agent: %v", err)
	}
	defer agent2.Cleanup(ctx)

	result2, err := agent2.Run(ctx, "Explain machine learning in simple terms.")
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
		SystemPrompt: "You are a creative assistant.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "huggingface",
			Model:       "meta-llama/Llama-3.2-1B-Instruct",
			APIKey:      apiKey,
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
	stream, err := streamAgent.RunStream(ctx, "Write a short haiku about programming.",
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

	// Example 4: Conversational Agent with Memory
	fmt.Println("Example 4: Conversational Agent")
	fmt.Println("=================================")

	config4 := &vnext.Config{
		Name:         "conversation-agent",
		SystemPrompt: "You are a helpful conversational assistant. Remember context from previous messages.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "huggingface",
			Model:       "meta-llama/Llama-3.2-1B-Instruct",
			APIKey:      apiKey,
			Temperature: 0.7,
			MaxTokens:   150,
		},
	}

	chatAgent, err := vnext.NewBuilder("conversation-agent").
		WithConfig(config4).
		Build()

	if err != nil {
		log.Fatalf("Failed to create chat agent: %v", err)
	}

	if err := chatAgent.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize chat agent: %v", err)
	}
	defer chatAgent.Cleanup(ctx)

	// Simulating a conversation
	conversation := []string{
		"My favorite color is blue.",
		"What's my favorite color?",
	}

	for i, msg := range conversation {
		fmt.Printf("\n[Turn %d] User: %s\n", i+1, msg)

		result, err := chatAgent.Run(ctx, msg)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		fmt.Printf("Agent: %s\n", result.Content)
	}
	fmt.Println()

	// Example 5: Using Different Temperature Settings
	fmt.Println("Example 5: Temperature Comparison")
	fmt.Println("===================================")

	temperatures := []float32{0.3, 0.7, 1.0}
	prompt := "Complete this sentence: The future of AI is"

	for _, temp := range temperatures {
		tempConfig := &vnext.Config{
			Name:         fmt.Sprintf("temp-%.1f", temp),
			SystemPrompt: "You are a creative assistant.",
			Timeout:      30 * time.Second,
			LLM: vnext.LLMConfig{
				Provider:    "huggingface",
				Model:       "meta-llama/Llama-3.2-1B-Instruct",
				APIKey:      apiKey,
				Temperature: temp,
				MaxTokens:   50,
			},
		}

		tempAgent, err := vnext.NewBuilder(fmt.Sprintf("temp-%.1f", temp)).
			WithConfig(tempConfig).
			Build()

		if err != nil {
			log.Printf("Failed to create agent with temp %.1f: %v", temp, err)
			continue
		}

		if err := tempAgent.Initialize(ctx); err != nil {
			log.Printf("Failed to initialize agent with temp %.1f: %v", temp, err)
			continue
		}

		result, err := tempAgent.Run(ctx, prompt)
		if err != nil {
			log.Printf("Call with temp %.1f failed: %v", temp, err)
			tempAgent.Cleanup(ctx)
			continue
		}

		fmt.Printf("\nTemperature: %.1f\n", temp)
		fmt.Printf("Response: %s\n", result.Content)

		tempAgent.Cleanup(ctx)
	}

	// Example 6: Using Inference Endpoints (Custom Deployment)
	fmt.Println("\nExample 6: Using Inference Endpoints (if available)")
	fmt.Println("====================================================")

	endpointURL := os.Getenv("HUGGINGFACE_ENDPOINT_URL")
	if endpointURL != "" {
		endpointConfig := &vnext.Config{
			Name:         "endpoint-agent",
			SystemPrompt: "You are a helpful assistant.",
			Timeout:      30 * time.Second,
			LLM: vnext.LLMConfig{
				Provider:    "huggingface",
				Model:       "custom-model",
				APIKey:      apiKey,
				BaseURL:     endpointURL,
				Temperature: 0.7,
				MaxTokens:   150,
			},
		}

		endpointAgent, err := vnext.NewBuilder("endpoint-agent").
			WithConfig(endpointConfig).
			Build()

		if err != nil {
			log.Printf("Failed to create endpoint agent: %v", err)
		} else if err := endpointAgent.Initialize(ctx); err != nil {
			log.Printf("Failed to initialize endpoint agent: %v", err)
		} else {
			defer endpointAgent.Cleanup(ctx)
			result, err := endpointAgent.Run(ctx, "Hello from custom endpoint!")
			if err != nil {
				log.Printf("Endpoint call failed: %v", err)
			} else {
				fmt.Printf("Response: %s\n", result.Content)
				fmt.Printf("Duration: %v\n", result.Duration)
			}
		}
	} else {
		fmt.Println("HUGGINGFACE_ENDPOINT_URL not set, skipping endpoint example")
		fmt.Println("To use this example, set up a dedicated Inference Endpoint and export its URL")
	}

	// Example 7: Using RunOptions for Detailed Results
	fmt.Println("\nExample 7: Detailed Results with RunOptions")
	fmt.Println("============================================")

	config7 := &vnext.Config{
		Name:         "detail-agent",
		SystemPrompt: "You are a helpful assistant.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "huggingface",
			Model:       "meta-llama/Llama-3.2-1B-Instruct",
			APIKey:      apiKey,
			Temperature: 0.7,
			MaxTokens:   200,
		},
	}

	detailAgent, err := vnext.NewBuilder("detail-agent").
		WithConfig(config7).
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
		AddContext("request_id", "hf-demo-123")

	detailResult, err := detailAgent.RunWithOptions(ctx, "What is deep learning?", opts)
	if err != nil {
		log.Fatalf("Run with options failed: %v", err)
	}

	fmt.Printf("Response: %s\n", detailResult.Content)
	fmt.Printf("Duration: %v\n", detailResult.Duration)
	fmt.Printf("Success: %v\n", detailResult.Success)
	fmt.Printf("Tokens Used: %d\n", detailResult.TokensUsed)
	fmt.Printf("Metadata: %v\n", detailResult.Metadata)

	// Example 8: Custom Handler with HuggingFace
	fmt.Println("\nExample 8: Custom Handler")
	fmt.Println("==========================")

	customHandler := func(ctx context.Context, input string, caps *vnext.Capabilities) (string, error) {
		// Custom logic: add context for technical questions
		if len(input) > 0 && (containsWord(input, "how") || containsWord(input, "what")) {
			// Add additional context for questions
			enhancedPrompt := fmt.Sprintf("Please provide a clear and concise answer to: %s", input)
			return caps.LLM("You are a technical expert who provides clear, accurate answers.", enhancedPrompt)
		}

		// For other inputs, process normally
		return caps.LLM("You are a helpful assistant.", input)
	}

	config8 := &vnext.Config{
		Name:         "custom-agent",
		SystemPrompt: "You are a helpful assistant.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "huggingface",
			Model:       "meta-llama/Llama-3.2-1B-Instruct",
			APIKey:      apiKey,
			Temperature: 0.7,
			MaxTokens:   150,
		},
	}

	customAgent, err := vnext.NewBuilder("custom-agent").
		WithConfig(config8).
		WithHandler(customHandler).
		Build()

	if err != nil {
		log.Fatalf("Failed to create custom agent: %v", err)
	}

	if err := customAgent.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize custom agent: %v", err)
	}
	defer customAgent.Cleanup(ctx)

	customResult, err := customAgent.Run(ctx, "How does neural network training work?")
	if err != nil {
		log.Printf("Custom handler failed: %v", err)
	} else {
		fmt.Printf("Response: %s\n", customResult.Content)
		fmt.Printf("Duration: %v\n", customResult.Duration)
	}

	// Example 9: Multiple Queries with Same Agent
	fmt.Println("\nExample 9: Multiple Queries")
	fmt.Println("============================")

	config9 := &vnext.Config{
		Name:         "multi-query-agent",
		SystemPrompt: "You are a helpful assistant that provides concise answers.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "huggingface",
			Model:       "meta-llama/Llama-3.2-1B-Instruct",
			APIKey:      apiKey,
			Temperature: 0.7,
			MaxTokens:   100,
		},
	}

	multiAgent, err := vnext.NewBuilder("multi-query-agent").
		WithConfig(config9).
		Build()

	if err != nil {
		log.Fatalf("Failed to create multi-query agent: %v", err)
	}

	if err := multiAgent.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize multi-query agent: %v", err)
	}
	defer multiAgent.Cleanup(ctx)

	queries := []string{
		"What is REST API?",
		"What is GraphQL?",
		"Difference between SQL and NoSQL?",
	}

	for i, query := range queries {
		fmt.Printf("\n[Query %d] %s\n", i+1, query)

		result, err := multiAgent.Run(ctx, query)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		fmt.Printf("Answer: %s\n", result.Content)
		fmt.Printf("Duration: %v\n", result.Duration)
	}

	fmt.Println("\n===========================================")
	fmt.Println("  HuggingFace vNext examples completed!")
	fmt.Println("===========================================")
	fmt.Println()
	fmt.Println("💡 Tips:")
	fmt.Println("  • Use 'meta-llama/Llama-3.2-1B-Instruct' for fast responses")
	fmt.Println("  • Use 'meta-llama/Llama-3.2-3B-Instruct' for better quality")
	fmt.Println("  • Lower temperature (0.3) for factual answers")
	fmt.Println("  • Higher temperature (0.8-1.0) for creative responses")
	fmt.Println("  • The new router API is OpenAI-compatible")
}

// Helper function to check if input contains a word
func containsWord(text, word string) bool {
	return len(text) >= len(word) &&
		(text[:len(word)] == word ||
			fmt.Sprintf(" %s", word) != "" &&
				len(text) > len(word))
}
