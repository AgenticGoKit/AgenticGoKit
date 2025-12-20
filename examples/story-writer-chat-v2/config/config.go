package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"
)

// Config holds application configuration
type Config struct {
	APIKey   string
	Port     string
	Provider string // e.g., "huggingface"
	Model    string // e.g., "Qwen/Qwen2.5-72B-Instruct"
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	apiKey := os.Getenv("HUGGINGFACE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("HUGGINGFACE_API_KEY environment variable not set\nPlease set it with: $env:HUGGINGFACE_API_KEY=\"your-key\"")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	provider := os.Getenv("LLM_PROVIDER")
	if provider == "" {
		provider = "huggingface"
	}

	model := os.Getenv("LLM_MODEL")
	if model == "" {
		// Qwen 2.5 72B is excellent at following instructions and structured formats
		// Alternative: meta-llama/Llama-3.1-70b-Instruct, mistralai/Mistral-Large-Instruct-2411
		model = "Qwen/Qwen2.5-72B-Instruct"
	}

	return &Config{
		APIKey:   apiKey,
		Port:     port,
		Provider: provider,
		Model:    model,
	}, nil
}

// ValidateAPIConnection verifies the API key works by making a test request
func ValidateAPIConnection(apiKey string) error {
	log.Println("🔍 Validating API connection...")

	testAgent, err := vnext.QuickChatAgentWithConfig("ValidationTest", &vnext.Config{
		Name:    "validation_test",
		Timeout: 15 * time.Second,
		LLM: vnext.LLMConfig{
			Provider: "huggingface",
			Model:    "Qwen/Qwen2.5-72B-Instruct",
			APIKey:   apiKey,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create test agent: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	_, err = testAgent.Run(ctx, "test")
	if err != nil {
		return fmt.Errorf("API connection test failed: %w", err)
	}

	return nil
}



