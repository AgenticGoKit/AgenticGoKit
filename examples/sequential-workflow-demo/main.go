package main

import (
	"context"
	"log"
	"os"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"

	// Import OpenRouter plugin
	_ "github.com/agenticgokit/agenticgokit/plugins/llm/openrouter"
)

func main() {
	// Validate API key
	apiKey, err := ValidateAPIKey()
	if err != nil {
		log.Fatalf("❌ %v\nPlease set it with: $env:OPENROUTER_API_KEY=\"your-key\"", err)
	}

	// Test OpenRouter connection
	log.Println("🔍 Checking OpenRouter connection...")
	if err := testOpenRouterConnection(apiKey); err != nil {
		log.Fatalf("❌ OpenRouter connection failed: %v\nCheck your API key and network connection", err)
	}
	log.Println("✅ OpenRouter connection successful")

	// Create workflow (application-specific: 2-agent sequential)
	workflow, err := NewSimpleSequentialWorkflow(apiKey)
	if err != nil {
		log.Fatalf("Failed to create workflow: %v", err)
	}

	// Create WebSocket server (reusable infrastructure)
	port := getPort()
	server := NewWebSocketServer(port, workflow)

	// Start server
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// testOpenRouterConnection verifies the API key works
func testOpenRouterConnection(apiKey string) error {
	testAgent, err := vnext.QuickChatAgentWithConfig("Test", &vnext.Config{
		Name:    "test",
		Timeout: 10 * time.Second,
		LLM: vnext.LLMConfig{
			Provider: "openrouter",
			Model:    "openai/gpt-4o-mini",
			APIKey:   apiKey,
		},
	})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err = testAgent.Run(ctx, "test")
	return err
}

// getPort returns the port to use (from env or default)
func getPort() string {
	port := "8080" // Different port to avoid conflict with story-writer
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	return port
}



