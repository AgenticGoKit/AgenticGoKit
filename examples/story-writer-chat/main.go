package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"

	// Import OpenRouter plugin
	_ "github.com/agenticgokit/agenticgokit/plugins/llm/openrouter"
)

func main() {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("❌ Configuration error: %v", err)
	}

	// Validate API connection
	if err := ValidateAPIConnection(config.APIKey); err != nil {
		log.Fatalf("❌ API validation failed: %v\nCheck your API key and network connection", err)
	}
	log.Println("✅ API connection validated")

	// Create workflow (application-specific)
	workflow, err := NewStoryWriterWorkflow(config)
	if err != nil {
		log.Fatalf("❌ Failed to create workflow: %v", err)
	}

	// Create and start WebSocket server (reusable infrastructure)
	server := NewWebSocketServer(config.Port, workflow)
	if err := server.Start(); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}

// Config holds application configuration
type Config struct {
	APIKey   string
	Port     string
	Provider string // e.g., "openrouter"
	Model    string // e.g., "openai/gpt-4o-mini"
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY environment variable not set\nPlease set it with: $env:OPENROUTER_API_KEY=\"your-key\"")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	provider := os.Getenv("LLM_PROVIDER")
	if provider == "" {
		provider = "openrouter"
	}

	model := os.Getenv("LLM_MODEL")
	if model == "" {
		model = "openai/gpt-4o-mini"
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
			Provider: "openrouter",
			Model:    "openai/gpt-4o-mini",
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



