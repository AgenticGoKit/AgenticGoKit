package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"

	// Plugins (blank imports activate providers)
	_ "github.com/kunalkushwaha/agenticgokit/plugins/logging/zerolog"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/memory/memory"
)

func main() {
	// Configure in-memory provider
	memCfg := core.AgentMemoryConfig{
		Provider:   "memory",
		Connection: "memory",
		MaxResults: 5,
		Dimensions: 1536,
		EnableRAG:  true,
		Documents:  core.DocumentConfig{AutoChunk: true},
		Embedding:  core.EmbeddingConfig{Provider: "dummy", MaxBatchSize: 50},
	}

	mem, err := core.NewMemory(memCfg)
	if err != nil {
		fmt.Println("error creating memory:", err)
		return
	}
	defer mem.Close()

	ctx := context.Background()
	ctx = mem.SetSession(ctx, mem.NewSession())

	// Personal memory
	_ = mem.Store(ctx, "Kunal likes Go and vectors", "profile", "likes")
	_ = mem.Store(ctx, "Kunal works on AgenticGoKit", "work")

	res, _ := mem.Query(ctx, "AgenticGoKit")
	fmt.Println("Personal search results:")
	for _, r := range res {
		fmt.Printf("- %s (%.2f)\n", r.Content, r.Score)
	}

	// Chat history
	_ = mem.AddMessage(ctx, "user", "Hello")
	_ = mem.AddMessage(ctx, "assistant", "Hi! How can I help?")
	hist, _ := mem.GetHistory(ctx, 10)
	fmt.Println("History:")
	for _, m := range hist {
		fmt.Printf("%s: %s\n", m.Role, m.Content)
	}

	// RAG Knowledge demo
	doc := core.Document{Content: "AgenticGoKit adds plugin registries for memory and logging.", Source: "docs", CreatedAt: time.Now()}
	_ = mem.IngestDocument(ctx, doc)

	rag, _ := mem.BuildContext(ctx, "What are plugins?", core.WithIncludeSources(true))
	fmt.Println("\nContext sources:")
	for _, s := range rag.Sources {
		fmt.Println("-", s)
	}
}
