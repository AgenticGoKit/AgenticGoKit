package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/kunalkushwaha/agenticgokit/core"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/llm/huggingface"
)

func main() {
	// Check for API key in environment
	apiKey := os.Getenv("HUGGINGFACE_API_KEY")
	if apiKey == "" {
		log.Fatal("HUGGINGFACE_API_KEY environment variable not set. Please set it with your Hugging Face API key.")
	}

	// Example 1: Basic Inference API usage with new router
	fmt.Println("Example 1: Basic Inference API Usage (New Router)")
	fmt.Println("===================================================")
	fmt.Println("Using the new HuggingFace router API (router.huggingface.co)")
	fmt.Println()

	config := core.LLMProviderConfig{
		Type:        "huggingface",
		APIKey:      apiKey,
		Model:       "meta-llama/Llama-3.2-1B-Instruct", // Router-compatible model
		MaxTokens:   100,
		Temperature: 0.7,
		HFAPIType:   "inference", // Use Inference API (now uses OpenAI-compatible format)
	}

	provider, err := core.NewModelProviderFromConfig(config)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	ctx := context.Background()
	response, err := provider.Call(ctx, core.Prompt{
		System: "You are a helpful assistant.",
		User:   "What is machine learning?",
	})

	if err != nil {
		log.Fatalf("Call failed: %v", err)
	}

	fmt.Printf("Response: %s\n", response.Content)
	fmt.Printf("Tokens used: %d (prompt: %d, completion: %d)\n\n",
		response.Usage.TotalTokens,
		response.Usage.PromptTokens,
		response.Usage.CompletionTokens)

	// Example 2: Using Chat API with different models
	fmt.Println("Example 2: Using Different Models on the Router")
	fmt.Println("================================================")
	fmt.Println("The new router supports various models from different providers")
	fmt.Println()

	chatConfig := core.LLMProviderConfig{
		Type:        "huggingface",
		APIKey:      apiKey,
		Model:       "meta-llama/Llama-3.2-3B-Instruct", // Larger Llama model
		MaxTokens:   200,
		Temperature: 0.8,
		HFAPIType:   "inference", // Router uses OpenAI-compatible format
	}

	chatProvider, err := core.NewModelProviderFromConfig(chatConfig)
	if err != nil {
		log.Fatalf("Failed to create chat provider: %v", err)
	}

	chatResponse, err := chatProvider.Call(ctx, core.Prompt{
		System: "You are a helpful assistant.",
		User:   "Explain neural networks in simple terms.",
	})

	if err != nil {
		log.Fatalf("Chat call failed: %v", err)
	}

	fmt.Printf("Response: %s\n", chatResponse.Content)
	fmt.Printf("Tokens used: %d\n\n", chatResponse.Usage.TotalTokens)

	// Example 3: Using Inference Endpoints (dedicated)
	fmt.Println("Example 3: Using Inference Endpoints (requires endpoint URL)")
	fmt.Println("=============================================================")

	endpointURL := os.Getenv("HUGGINGFACE_ENDPOINT_URL")
	if endpointURL != "" {
		endpointConfig := core.LLMProviderConfig{
			Type:           "huggingface",
			APIKey:         apiKey,
			Model:          "custom-model",
			BaseURL:        endpointURL,
			MaxTokens:      150,
			Temperature:    0.7,
			HFAPIType:      "endpoint", // Use Inference Endpoints
			HFWaitForModel: true,
		}

		endpointProvider, err := core.NewModelProviderFromConfig(endpointConfig)
		if err != nil {
			log.Printf("Failed to create endpoint provider: %v", err)
		} else {
			endpointResponse, err := endpointProvider.Call(ctx, core.Prompt{
				User: "Hello from dedicated endpoint!",
			})

			if err != nil {
				log.Printf("Endpoint call failed: %v", err)
			} else {
				fmt.Printf("Response: %s\n\n", endpointResponse.Content)
			}
		}
	} else {
		fmt.Println("HUGGINGFACE_ENDPOINT_URL not set, skipping endpoint example\n")
	}

	// Example 4: Using Text Generation Inference (TGI) - self-hosted
	fmt.Println("Example 4: Using Text Generation Inference (TGI)")
	fmt.Println("=================================================")

	tgiURL := os.Getenv("HUGGINGFACE_TGI_URL")
	if tgiURL != "" {
		tgiConfig := core.LLMProviderConfig{
			Type:        "huggingface",
			Model:       "tgi-model",
			BaseURL:     tgiURL,
			MaxTokens:   200,
			Temperature: 0.9,
			HFAPIType:   "tgi", // Use TGI
			HFTopP:      0.95,
			HFTopK:      50,
			HFDoSample:  true,
		}

		tgiProvider, err := core.NewModelProviderFromConfig(tgiConfig)
		if err != nil {
			log.Printf("Failed to create TGI provider: %v", err)
		} else {
			tgiResponse, err := tgiProvider.Call(ctx, core.Prompt{
				User: "Generate a creative story opening.",
			})

			if err != nil {
				log.Printf("TGI call failed: %v", err)
			} else {
				fmt.Printf("Response: %s\n\n", tgiResponse.Content)
			}
		}
	} else {
		fmt.Println("HUGGINGFACE_TGI_URL not set, skipping TGI example")
		fmt.Println("To use TGI, set up a local TGI server (e.g., http://localhost:8080)\n")
	}

	// Example 5: Streaming responses
	fmt.Println("Example 5: Streaming Responses")
	fmt.Println("================================")

	streamConfig := core.LLMProviderConfig{
		Type:        "huggingface",
		APIKey:      apiKey,
		Model:       "meta-llama/Llama-3.2-1B-Instruct", // Router-compatible model
		MaxTokens:   150,
		Temperature: 0.8,
		HFAPIType:   "inference",
	}

	streamProvider, err := core.NewModelProviderFromConfig(streamConfig)
	if err != nil {
		log.Fatalf("Failed to create stream provider: %v", err)
	}

	fmt.Print("Streaming response: ")
	tokenChan, err := streamProvider.Stream(ctx, core.Prompt{
		User: "Write a short poem about AI.",
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

	// Example 6: Embeddings (Note: Embeddings require separate configuration)
	fmt.Println("Example 6: Embeddings API")
	fmt.Println("=========================")
	fmt.Println("Note: The embeddings API uses a different endpoint structure.")
	fmt.Println("For production use, consider using HuggingFace Inference Endpoints")
	fmt.Println("or dedicated embedding services.")
	fmt.Println()

	// Skipping embeddings example as it requires different endpoint configuration
	fmt.Println("Embeddings example skipped - requires dedicated endpoint configuration\n")

	// Example 7: Using AgentLLMConfig (automatically reads environment variables)
	fmt.Println("Example 7: Using AgentLLMConfig")
	fmt.Println("=================================")
	fmt.Println("AgentLLMConfig automatically reads HUGGINGFACE_API_KEY, HUGGINGFACE_API_TYPE environment variables")

	envConfig := core.AgentLLMConfig{
		Provider:    "huggingface",
		Model:       "meta-llama/Llama-3.2-1B-Instruct", // Router-compatible model
		MaxTokens:   150,
		Temperature: 0.7,
	}

	envProvider, err := core.NewLLMProvider(envConfig)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	envResponse, err := envProvider.Call(ctx, core.Prompt{
		User: "What is the future of AI?",
	})

	if err != nil {
		log.Fatalf("Call failed: %v", err)
	}

	fmt.Printf("Response: %s\n\n", envResponse.Content)

	// Example 8: Advanced parameters
	fmt.Println("Example 8: Advanced HuggingFace Parameters")
	fmt.Println("===========================================")

	advancedConfig := core.LLMProviderConfig{
		Type:                "huggingface",
		APIKey:              apiKey,
		Model:               "meta-llama/Llama-3.2-1B-Instruct", // Router-compatible model
		MaxTokens:           100,
		Temperature:         0.8,
		HFAPIType:           "inference",
		HFWaitForModel:      true,             // Wait for model to load if needed
		HFUseCache:          false,            // Don't use cached responses
		HFDoSample:          true,             // Use sampling
		HFTopP:              0.9,              // Nucleus sampling
		HFTopK:              50,               // Top-k sampling
		HFRepetitionPenalty: 1.2,              // Penalize repetition
		HFStopSequences:     []string{"\n\n"}, // Stop at double newline
	}

	advancedProvider, err := core.NewModelProviderFromConfig(advancedConfig)
	if err != nil {
		log.Fatalf("Failed to create advanced provider: %v", err)
	}

	advancedResponse, err := advancedProvider.Call(ctx, core.Prompt{
		User: "Generate a creative sentence about space exploration.",
	})

	if err != nil {
		log.Fatalf("Advanced call failed: %v", err)
	}

	fmt.Printf("Response with advanced parameters: %s\n", advancedResponse.Content)
	fmt.Println("Parameters used:")
	fmt.Println("  - Wait for model: true")
	fmt.Println("  - Use cache: false")
	fmt.Println("  - Do sample: true")
	fmt.Println("  - Top-p: 0.9")
	fmt.Println("  - Top-k: 50")
	fmt.Println("  - Repetition penalty: 1.2")
	fmt.Println("  - Stop sequences: ['\\n\\n']")
}
