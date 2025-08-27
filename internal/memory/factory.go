// Package memory provides internal memory factory implementations for AgentFlow.
package memory

import (
	"context"
	"fmt"

	"github.com/kunalkushwaha/agenticgokit/core"
	"github.com/kunalkushwaha/agenticgokit/internal/embedding"
	"github.com/kunalkushwaha/agenticgokit/internal/memory/providers"
)

// NewMemory creates a new memory instance based on configuration
func NewMemory(config core.AgentMemoryConfig) (core.Memory, error) {
	// Set core memory defaults
	if config.MaxResults == 0 {
		config.MaxResults = 10
	}
	if config.Dimensions == 0 {
		config.Dimensions = 1536
	}
	if config.Connection == "" && config.Provider == "memory" {
		config.Connection = "memory"
	}

	// Set RAG defaults
	if config.KnowledgeMaxResults == 0 {
		config.KnowledgeMaxResults = 20
	}
	if config.KnowledgeScoreThreshold == 0 {
		// Check if threshold was explicitly set to 0.0 in TOML
		// If embedding provider is real (not dummy), use default 0.7 threshold
		// unless explicitly configured to use 0.0 (which we'll preserve)
		if config.Embedding.Provider == "dummy" {
			config.KnowledgeScoreThreshold = 0.0 // No filtering for dummy embeddings
		} else {
			// For real embeddings, use 0.7 as default only if not explicitly set
			// We can't distinguish between unset and explicitly set to 0.0 here,
			// so we'll use a more permissive default for RAG systems
			config.KnowledgeScoreThreshold = 0.3 // More permissive default for real embeddings
		}
	}
	if config.ChunkSize == 0 {
		config.ChunkSize = 1000
	}
	if config.ChunkOverlap == 0 {
		config.ChunkOverlap = 200
	}
	if config.RAGMaxContextTokens == 0 {
		config.RAGMaxContextTokens = 4000
	}
	if config.RAGPersonalWeight == 0 {
		config.RAGPersonalWeight = 0.3
	}
	if config.RAGKnowledgeWeight == 0 {
		config.RAGKnowledgeWeight = 0.7
	}

	// Set document processing defaults
	if len(config.Documents.SupportedTypes) == 0 {
		config.Documents.SupportedTypes = []string{"pdf", "txt", "md", "web", "code"}
	}
	if config.Documents.MaxFileSize == "" {
		config.Documents.MaxFileSize = "10MB"
	}

	// Set embedding service defaults
	if config.Embedding.Provider == "" {
		config.Embedding.Provider = "dummy"
	}
	if config.Embedding.Model == "" {
		config.Embedding.Model = "text-embedding-3-small"
	}
	if config.Embedding.MaxBatchSize == 0 {
		config.Embedding.MaxBatchSize = 100
	}
	if config.Embedding.TimeoutSeconds == 0 {
		config.Embedding.TimeoutSeconds = 30
	}

	// Set search defaults
	if config.Search.KeywordWeight == 0 {
		config.Search.KeywordWeight = 0.3
	}
	if config.Search.SemanticWeight == 0 {
		config.Search.SemanticWeight = 0.7
	}

	// Create embedding service using internal factory
	embeddingService, err := embedding.NewEmbeddingService(
		config.Embedding.Provider,
		config.Embedding.Model,
		config.Embedding.APIKey,
		config.Embedding.BaseURL,
		config.Dimensions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding service: %w", err)
	}

	switch config.Provider {
	case "memory":
		return providers.NewInMemoryProvider(config)
	case "pgvector":
		return providers.NewPgVectorProvider(config, embeddingService)
	case "weaviate":
		return providers.NewWeaviateProvider(config)
	default:
		return nil, fmt.Errorf("unsupported memory provider: %s", config.Provider)
	}
}

// QuickMemory creates an in-memory provider for quick testing
func QuickMemory() core.Memory {
	config := core.AgentMemoryConfig{
		Provider:   "memory",
		Connection: "memory",
		MaxResults: 10,
		Dimensions: 1536,
	}

	memory, err := NewMemory(config)
	if err != nil {
		// Return no-op memory instead of panicking
		return &noOpMemory{}
	}

	return memory
}

// Temporary no-op memory implementation during refactoring
type noOpMemory struct{}

func (m *noOpMemory) Store(ctx context.Context, content string, tags ...string) error {
	return nil
}

func (m *noOpMemory) Query(ctx context.Context, query string, limit ...int) ([]core.Result, error) {
	return []core.Result{}, nil
}

func (m *noOpMemory) Remember(ctx context.Context, key string, value any) error {
	return nil
}

func (m *noOpMemory) Recall(ctx context.Context, key string) (any, error) {
	return nil, nil
}

func (m *noOpMemory) AddMessage(ctx context.Context, role, content string) error {
	return nil
}

func (m *noOpMemory) GetHistory(ctx context.Context, limit ...int) ([]core.Message, error) {
	return []core.Message{}, nil
}

func (m *noOpMemory) NewSession() string {
	return "default"
}

func (m *noOpMemory) SetSession(ctx context.Context, sessionID string) context.Context {
	return ctx
}

func (m *noOpMemory) ClearSession(ctx context.Context) error {
	return nil
}

func (m *noOpMemory) Close() error {
	return nil
}

func (m *noOpMemory) IngestDocument(ctx context.Context, doc core.Document) error {
	return nil
}

func (m *noOpMemory) IngestDocuments(ctx context.Context, docs []core.Document) error {
	return nil
}

func (m *noOpMemory) SearchKnowledge(ctx context.Context, query string, options ...core.SearchOption) ([]core.KnowledgeResult, error) {
	return []core.KnowledgeResult{}, nil
}

func (m *noOpMemory) SearchAll(ctx context.Context, query string, options ...core.SearchOption) (*core.HybridResult, error) {
	return &core.HybridResult{}, nil
}

func (m *noOpMemory) BuildContext(ctx context.Context, query string, options ...core.ContextOption) (*core.RAGContext, error) {
	return &core.RAGContext{}, nil
}

// Register the internal memory factory with core to avoid circular imports
func init() {
	core.RegisterMemoryFactory(NewMemory)
}
