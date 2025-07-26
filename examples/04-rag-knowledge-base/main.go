package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kunalkushwaha/AgenticGoKit/core"
	"github.com/kunalkushwaha/AgenticGoKit/examples/04-rag-knowledge-base/agents"
	"github.com/kunalkushwaha/AgenticGoKit/examples/04-rag-knowledge-base/services"
)

func main() {
	// Command line flags
	var (
		mode     = flag.String("mode", "query", "Mode: ingest, query, batch-ingest, interactive")
		question = flag.String("question", "", "Question to ask (for query mode)")
		path     = flag.String("path", "", "Path to document or directory (for ingest modes)")
		pattern  = flag.String("pattern", "*.txt,*.md", "File patterns for batch ingest")
		filter   = flag.String("filter", "", "Filter for queries (e.g., 'type:api_doc')")
		batchSize = flag.Int("batch-size", 5, "Batch size for processing")
	)
	flag.Parse()

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize services
	embeddingService, err := services.NewEmbeddingService(config.EmbeddingProvider, config.EmbeddingModel)
	if err != nil {
		log.Fatalf("Failed to initialize embedding service: %v", err)
	}

	vectorStore, err := services.NewVectorStore(config.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize vector store: %v", err)
	}
	defer vectorStore.Close()

	llmProvider, err := initializeLLMProvider(config.LLMProvider)
	if err != nil {
		log.Fatalf("Failed to initialize LLM provider: %v", err)
	}

	// Initialize agents
	ingestionAgent := agents.NewIngestionAgent("ingestion", embeddingService, vectorStore)
	retrievalAgent := agents.NewRetrievalAgent("retrieval", embeddingService, vectorStore)
	synthesisAgent := agents.NewSynthesisAgent("synthesis", llmProvider)

	ctx := context.Background()

	switch *mode {
	case "ingest":
		if *path == "" {
			log.Fatal("Path is required for ingest mode")
		}
		err = ingestDocument(ctx, ingestionAgent, *path)
	case "batch-ingest":
		if *path == "" {
			log.Fatal("Path is required for batch-ingest mode")
		}
		err = batchIngest(ctx, ingestionAgent, *path, *pattern, *batchSize)
	case "query":
		if *question == "" {
			log.Fatal("Question is required for query mode")
		}
		err = queryKnowledgeBase(ctx, retrievalAgent, synthesisAgent, *question, *filter)
	case "interactive":
		err = interactiveMode(ctx, retrievalAgent, synthesisAgent)
	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}

	if err != nil {
		log.Fatalf("Operation failed: %v", err)
	}
}

type Config struct {
	EmbeddingProvider string
	EmbeddingModel    string
	LLMProvider       string
	DatabaseURL       string
}

func loadConfig() (*Config, error) {
	config := &Config{
		EmbeddingProvider: getEnvOrDefault("EMBEDDING_PROVIDER", "openai"),
		EmbeddingModel:    getEnvOrDefault("EMBEDDING_MODEL", "text-embedding-ada-002"),
		LLMProvider:       getEnvOrDefault("LLM_PROVIDER", "openai"),
		DatabaseURL:       getEnvOrDefault("DATABASE_URL", "postgres://agentflow:password@localhost:5432/agentflow?sslmode=disable"),
	}

	// Validate required environment variables
	switch config.EmbeddingProvider {
	case "openai":
		if os.Getenv("OPENAI_API_KEY") == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required for OpenAI embeddings")
		}
	case "ollama":
		config.EmbeddingModel = getEnvOrDefault("EMBEDDING_MODEL", "nomic-embed-text:latest")
	}

	return config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func initializeLLMProvider(provider string) (core.ModelProvider, error) {
	switch provider {
	case "openai":
		return core.NewOpenAIProvider()
	case "azure":
		return core.NewAzureOpenAIProvider()
	case "ollama":
		return core.NewOllamaProvider()
	case "mock":
		return core.NewMockProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}
}

func ingestDocument(ctx context.Context, agent *agents.IngestionAgent, filePath string) error {
	fmt.Printf("Ingesting document: %s\n", filePath)

	event := core.NewEvent("ingest", map[string]interface{}{
		"file_path": filePath,
	})

	state := core.NewState()
	result, err := agent.Execute(ctx, event, state)
	if err != nil {
		return fmt.Errorf("ingestion failed: %w", err)
	}

	fmt.Printf("✅ Document ingested successfully!\n")
	fmt.Printf("   File: %s\n", result.Data["file_processed"])
	fmt.Printf("   Chunks stored: %v\n", result.Data["chunks_stored"])

	return nil
}

func batchIngest(ctx context.Context, agent *agents.IngestionAgent, dirPath, pattern string, batchSize int) error {
	fmt.Printf("Batch ingesting documents from: %s\n", dirPath)
	fmt.Printf("Pattern: %s, Batch size: %d\n", pattern, batchSize)

	// Find files matching pattern
	patterns := strings.Split(pattern, ",")
	var files []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		for _, p := range patterns {
			matched, err := filepath.Match(strings.TrimSpace(p), info.Name())
			if err != nil {
				continue
			}
			if matched {
				files = append(files, path)
				break
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(files) == 0 {
		fmt.Printf("No files found matching pattern: %s\n", pattern)
		return nil
	}

	fmt.Printf("Found %d files to process\n", len(files))

	// Process files in batches
	processed := 0
	failed := 0

	for i := 0; i < len(files); i += batchSize {
		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}

		batch := files[i:end]
		fmt.Printf("\nProcessing batch %d/%d (%d files)...\n", 
			(i/batchSize)+1, (len(files)+batchSize-1)/batchSize, len(batch))

		for _, file := range batch {
			fmt.Printf("  Processing: %s... ", filepath.Base(file))
			
			err := ingestDocument(ctx, agent, file)
			if err != nil {
				fmt.Printf("❌ Failed: %v\n", err)
				failed++
			} else {
				fmt.Printf("✅ Success\n")
				processed++
			}
		}
	}

	fmt.Printf("\n📊 Batch ingestion complete!\n")
	fmt.Printf("   Processed: %d files\n", processed)
	fmt.Printf("   Failed: %d files\n", failed)
	fmt.Printf("   Success rate: %.1f%%\n", float64(processed)/float64(len(files))*100)

	return nil
}

func queryKnowledgeBase(ctx context.Context, retrievalAgent *agents.RetrievalAgent, synthesisAgent *agents.SynthesisAgent, question, filter string) error {
	fmt.Printf("🔍 Searching knowledge base...\n")
	fmt.Printf("Question: %s\n", question)

	start := time.Now()

	// Step 1: Retrieve relevant context
	retrievalEvent := core.NewEvent("retrieve", map[string]interface{}{
		"query":  question,
		"filter": filter,
	})

	retrievalState := core.NewState()
	retrievalResult, err := retrievalAgent.Execute(ctx, retrievalEvent, retrievalState)
	if err != nil {
		return fmt.Errorf("retrieval failed: %w", err)
	}

	retrievalTime := time.Since(start)

	// Check if we found relevant context
	contexts := retrievalResult.Data["retrieved_contexts"].([]string)
	if len(contexts) == 0 {
		fmt.Printf("❌ No relevant information found in the knowledge base.\n")
		fmt.Printf("   Try ingesting more documents or rephrasing your question.\n")
		return nil
	}

	fmt.Printf("📚 Found %d relevant context(s) (%.2fs)\n", len(contexts), retrievalTime.Seconds())

	// Step 2: Synthesize answer
	synthesisEvent := core.NewEvent("synthesize", map[string]interface{}{
		"query":              question,
		"retrieved_contexts": contexts,
		"sources":           retrievalResult.Data["sources"],
	})

	synthesisState := core.NewState()
	synthesisResult, err := synthesisAgent.Execute(ctx, synthesisEvent, synthesisState)
	if err != nil {
		return fmt.Errorf("synthesis failed: %w", err)
	}

	totalTime := time.Since(start)

	// Display results
	fmt.Printf("\n💡 Answer:\n")
	fmt.Printf("%s\n", synthesisResult.Data["answer"])

	if sources, ok := synthesisResult.Data["sources"].([]string); ok && len(sources) > 0 {
		fmt.Printf("\n📖 Sources:\n")
		uniqueSources := make(map[string]bool)
		for _, source := range sources {
			if !uniqueSources[source] {
				fmt.Printf("   • %s\n", source)
				uniqueSources[source] = true
			}
		}
	}

	fmt.Printf("\n⏱️  Query completed in %.2fs\n", totalTime.Seconds())
	fmt.Printf("   Retrieval: %.2fs, Synthesis: %.2fs\n", 
		retrievalTime.Seconds(), 
		(totalTime - retrievalTime).Seconds())

	return nil
}

func interactiveMode(ctx context.Context, retrievalAgent *agents.RetrievalAgent, synthesisAgent *agents.SynthesisAgent) error {
	fmt.Printf("🤖 Interactive Knowledge Base Query\n")
	fmt.Printf("Type your questions or 'exit' to quit.\n\n")

	for {
		fmt.Print("❓ Question: ")
		var question string
		fmt.Scanln(&question)

		if strings.ToLower(question) == "exit" {
			fmt.Printf("👋 Goodbye!\n")
			break
		}

		if strings.TrimSpace(question) == "" {
			continue
		}

		fmt.Println()
		err := queryKnowledgeBase(ctx, retrievalAgent, synthesisAgent, question, "")
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		}
		fmt.Println()
	}

	return nil
}