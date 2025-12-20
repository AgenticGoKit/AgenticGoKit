package main

import (
	"log"

	"github.com/agenticgokit/agenticgokit/examples/story-writer-chat-v2/config"
	"github.com/agenticgokit/agenticgokit/examples/story-writer-chat-v2/infrastructure"
	"github.com/agenticgokit/agenticgokit/examples/story-writer-chat-v2/workflow"

	// Import HuggingFace plugin
	_ "github.com/agenticgokit/agenticgokit/plugins/llm/huggingface"
	// Import memory provider plugin for in-memory storage
	_ "github.com/agenticgokit/agenticgokit/plugins/memory/memory"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Configuration error: %v", err)
	}

	// Validate API connection
	if err := config.ValidateAPIConnection(cfg.APIKey); err != nil {
		log.Fatalf("❌ API validation failed: %v\nCheck your API key and network connection", err)
	}
	log.Println("✅ API connection validated")

	// Create workflow (application-specific)
	wf, err := workflow.NewStoryWriterWorkflow(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to create workflow: %v", err)
	}

	// Create and start WebSocket server (reusable infrastructure)
	server := infrastructure.NewWebSocketServer(cfg.Port, wf)
	if err := server.Start(); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}
