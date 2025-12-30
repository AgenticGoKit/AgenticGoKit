package main

import (
	"context"
	"fmt"
	"time"

	"github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
	// Create an agent with memory enabled
	// We use "memory" provider which is ephemeral (in-memory) for this demo
	agent, err := v1beta.NewBuilder("Simple Agent").
		WithConfig(&v1beta.Config{
			Name:         "memory-agent",
			SystemPrompt: "You are a helpful assistant. Answer concisely.",
			Timeout:      30 * time.Second,
			LLM: v1beta.LLMConfig{
				Provider: "ollama",
				Model:    "gemma3:1b", // Using the model from user request
				BaseURL:  "http://localhost:11434",
			},
			Memory: &v1beta.MemoryConfig{
				Provider: "chromem",
				Options: map[string]string{
					"dimensions": "1536", // Dummy embedder dimensions
				},
				RAG: &v1beta.RAGConfig{
					MaxTokens:    2000,
					HistoryLimit: 5,
				},
			},
		}).Build()

	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// 1. First interaction
	fmt.Println("--- Interaction 1: User asks a question ---")
	query1 := "What is the capital of France?"
	fmt.Printf("User: %s\n", query1)

	response, err := agent.Run(ctx, query1)
	if err != nil {
		fmt.Printf("Error running agent: %v\n", err)
	} else {
		fmt.Printf("Agent: %s\n", response.Content)
	}

	// 2. Programmatic Memory Inspection & Manual Usage
	fmt.Println("\n--- Inspecting Memory Programmatically ---")
	mem := agent.Memory()
	if mem == nil {
		fmt.Println("Error: Agent memory is nil!")
	} else {
		// Test Manual Storage
		fmt.Println("Storing manual entry 'My favorite color is blue'...")
		if err := mem.Store(ctx, "My favorite color is blue"); err != nil {
			fmt.Printf("Manual store failed: %v\n", err)
		}

		// Query memory
		fmt.Println("Querying memory for 'blue'...")
		results, err := mem.Query(ctx, "blue", v1beta.WithLimit(5))
		if err != nil {
			fmt.Printf("Error querying memory: %v\n", err)
		} else {
			fmt.Printf("Found %d memory entries:\n", len(results))
			for i, res := range results {
				fmt.Printf(" [%d] Content: %q (Score: %.2f)\n", i+1, res.Content, res.Score)
			}
		}
	}
	fmt.Println()

	// 3. Second interaction - testing recall
	fmt.Println("--- Interaction 2: Testing Recall ---")
	query2 := "What did I just ask you?"
	fmt.Printf("User: %s\n", query2)

	response, err = agent.Run(ctx, query2)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Agent: %s\n", response.Content)
	fmt.Printf("Memory Queries: %d\n", response.MemoryQueries)

	// Check if we hit the memory context
	if response.MemoryContext != nil {
		fmt.Printf("Memory Context Used: Yes (Tokens: %d)\n", response.MemoryContext.TotalTokens)
		if len(response.MemoryContext.ChatHistory) > 0 {
			fmt.Println("Chat History captured.")
		}
	} else {
		fmt.Println("Memory Context Used: No (or nil)")
	}
	// 4. Testing RAG (Knowledge Base)
	fmt.Println("\n--- Testing RAG (Knowledge Base) ---")
	fmt.Println("Ingesting facts about AgenticGoKit...")
	docs := []v1beta.Document{
		{
			ID:      "agk-1",
			Title:   "What is AgenticGoKit?",
			Content: "AgenticGoKit is a powerful framework for building agentic AI applications in Go.",
			Source:  "internal-docs",
		},
		{
			ID:      "agk-2",
			Title:   "Memory Support",
			Content: "AgenticGoKit supports multiple memory providers including in-memory, PGVector, and Chromem.",
			Source:  "internal-docs",
		},
	}

	if err := mem.IngestDocuments(ctx, docs); err != nil {
		fmt.Printf("Ingest documents failed: %v\n", err)
	}

	fmt.Println("Searching knowledge base for 'PGVector'...")
	kResults, err := mem.SearchKnowledge(ctx, "PGVector", v1beta.WithLimit(2))
	if err != nil {
		fmt.Printf("Search knowledge failed: %v\n", err)
	} else {
		fmt.Printf("Found %d knowledge entries:\n", len(kResults))
		for i, res := range kResults {
			fmt.Printf(" [%d] Content: %q (Source: %s, Score: %.2f)\n", i+1, res.Content, res.Source, res.Score)
		}
	}

	// Ask the agent something it only knows from the ingested documents
	fmt.Println("\n--- Interaction 3: Asking about ingested knowledge ---")
	query3 := "Which memory providers does AgenticGoKit support?"
	fmt.Printf("User: %s\n", query3)
	response, err = agent.Run(ctx, query3)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Agent: %s\n", response.Content)
	}

	fmt.Printf("\nTraceID: %s\n", response.TraceID)
}
