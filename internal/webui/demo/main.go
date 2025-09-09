package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kunalkushwaha/agenticgokit/internal/webui"
)

// Demo program to test the WebUI server implementation with WebSocket support
func main() {
	fmt.Println("🚀 Starting AgenticGoKit WebUI Demo with WebSocket Support")
	fmt.Println("========================================================")

	// Create server configuration
	config := webui.ServerConfig{
		Port:      "8080",
		StaticDir: "./internal/webui/static",
		Config:    nil, // We'll run without agentflow config for this demo
	}

	// Create the server
	server := webui.NewServer(config)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n🛑 Received interrupt signal, shutting down...")
		cancel()
	}()

	// Start the server
	fmt.Printf("🌐 Starting server on port %s...\n", "8080")
	if err := server.Start(ctx); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}

	// Print access information
	fmt.Println("✅ Server started successfully!")
	fmt.Println()
	fmt.Println("📱 Access the WebUI at:")
	fmt.Printf("   🏠 Main Interface: %s\n", server.GetURL())
	fmt.Printf("   🔌 WebSocket:      ws://localhost:8080/ws\n")
	fmt.Printf("   ❤️  Health Check:   %s/api/health\n", server.GetURL())
	fmt.Printf("   ⚙️  Configuration:  %s/api/config\n", server.GetURL())
	fmt.Printf("   👥 Sessions:       %s/api/sessions\n", server.GetURL())
	fmt.Println()
	fmt.Println("🔗 WebSocket Protocol Examples:")
	fmt.Println("   Create Session: {\"type\":\"session_create\",\"timestamp\":\"2025-01-09T10:00:00Z\",\"data\":{}}")
	fmt.Println("   Send Message:   {\"type\":\"chat_message\",\"session_id\":\"<id>\",\"timestamp\":\"2025-01-09T10:00:00Z\",\"data\":{\"content\":\"Hello!\"}}")
	fmt.Println("   Ping:           {\"type\":\"ping\",\"timestamp\":\"2025-01-09T10:00:00Z\"}")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop the server...")

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	fmt.Println("🔄 Shutting down server...")
	if err := server.Stop(); err != nil {
		log.Printf("❌ Error during shutdown: %v", err)
	} else {
		fmt.Println("✅ Server stopped gracefully")
	}

	fmt.Println("👋 Demo completed!")
}
