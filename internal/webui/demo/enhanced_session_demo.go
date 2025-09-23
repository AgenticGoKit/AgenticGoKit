package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
	"github.com/kunalkushwaha/agenticgokit/internal/webui"
)

// Demo program to showcase enhanced session management features
func main() {
	fmt.Println("🚀 Enhanced Session Management Demo")
	fmt.Println("==================================")

	// Create configuration
	config := &core.Config{}
	sessionConfig := webui.DefaultSessionConfig()
	sessionConfig.MaxSessions = 5
	sessionConfig.MaxMessages = 10
	sessionConfig.SessionTimeout = 30 * time.Second

	// Create enhanced session manager
	manager, err := webui.NewEnhancedSessionManager(config, sessionConfig)
	if err != nil {
		log.Fatalf("Failed to create session manager: %v", err)
	}
	defer manager.Stop()

	fmt.Println("✅ Enhanced session manager created")

	// Create some test sessions
	fmt.Println("\n📝 Creating test sessions...")
	session1, err := manager.CreateSession("TestBrowser/1.0", "192.168.1.100")
	if err != nil {
		log.Fatalf("Failed to create session 1: %v", err)
	}
	fmt.Printf("   Session 1: %s\n", session1.ID)

	session2, err := manager.CreateSession("Chrome/120.0", "192.168.1.101")
	if err != nil {
		log.Fatalf("Failed to create session 2: %v", err)
	}
	fmt.Printf("   Session 2: %s\n", session2.ID)

	// Add messages to sessions
	fmt.Println("\n💬 Adding messages to sessions...")
	msg1 := webui.ChatMessage{
		ID:      "msg-1",
		Role:    "user",
		Content: "Hello, world!",
	}

	err = manager.AddMessage(session1.ID, msg1)
	if err != nil {
		log.Printf("Failed to add message: %v", err)
	} else {
		fmt.Println("   Message added to session 1")
	}

	msg2 := webui.ChatMessage{
		ID:      "msg-2",
		Role:    "agent",
		Content: "Hello! How can I help you?",
	}

	err = manager.AddMessage(session1.ID, msg2)
	if err != nil {
		log.Printf("Failed to add message: %v", err)
	} else {
		fmt.Println("   Response added to session 1")
	}

	// Get session metrics
	fmt.Println("\n📊 Session metrics:")
	metrics, err := manager.GetMetrics()
	if err != nil {
		log.Printf("Failed to get metrics: %v", err)
	} else {
		fmt.Printf("   Total sessions: %d\n", metrics.TotalSessions)
		fmt.Printf("   Active sessions: %d\n", metrics.ActiveSessions)
		fmt.Printf("   Total messages: %d\n", metrics.TotalMessages)
	}

	// List sessions with pagination
	fmt.Println("\n📋 Listing sessions (paginated):")
	sessions, total, err := manager.ListSessions(0, 10)
	if err != nil {
		log.Printf("Failed to list sessions: %v", err)
	} else {
		fmt.Printf("   Found %d sessions (total: %d)\n", len(sessions), total)
		for i, session := range sessions {
			fmt.Printf("   %d. %s (%s) - %d messages\n",
				i+1, session.ID, session.UserAgent, len(session.Messages))
		}
	}

	// Test session retrieval
	fmt.Println("\n🔍 Testing session retrieval:")
	retrievedSession, err := manager.GetSession(session1.ID)
	if err != nil {
		log.Printf("Failed to retrieve session: %v", err)
	} else {
		fmt.Printf("   Retrieved session: %s\n", retrievedSession.ID)
		fmt.Printf("   Messages: %d\n", len(retrievedSession.Messages))
		fmt.Printf("   User agent: %s\n", retrievedSession.UserAgent)
		fmt.Printf("   IP address: %s\n", retrievedSession.IPAddress)
	}

	fmt.Println("\n✅ Enhanced session management demo completed successfully!")
	fmt.Println("\nKey features demonstrated:")
	fmt.Println("- ✅ Session creation with metadata")
	fmt.Println("- ✅ Message management")
	fmt.Println("- ✅ Session metrics collection")
	fmt.Println("- ✅ Paginated session listing")
	fmt.Println("- ✅ Session retrieval")
	fmt.Println("- ✅ Enhanced session data structure")
}
