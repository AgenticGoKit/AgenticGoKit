package core

import (
	"context"
)

// EmbeddingService interface for generating embeddings
type EmbeddingService interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
	GetDimensions() int
}

// Embedding service factory functions using registry pattern to avoid circular imports
var (
	openAIEmbeddingFactory func(string, string) EmbeddingService
	ollamaEmbeddingFactory func(string, string) EmbeddingService
	dummyEmbeddingFactory  func(int) EmbeddingService
)

// RegisterOpenAIEmbeddingFactory allows internal packages to register their factory function
func RegisterOpenAIEmbeddingFactory(factory func(string, string) EmbeddingService) {
	openAIEmbeddingFactory = factory
}

// RegisterOllamaEmbeddingFactory allows internal packages to register their factory function
func RegisterOllamaEmbeddingFactory(factory func(string, string) EmbeddingService) {
	ollamaEmbeddingFactory = factory
}

// RegisterDummyEmbeddingFactory allows internal packages to register their factory function
func RegisterDummyEmbeddingFactory(factory func(int) EmbeddingService) {
	dummyEmbeddingFactory = factory
}

// NewOpenAIEmbeddingService creates a new OpenAI embedding service
func NewOpenAIEmbeddingService(apiKey, model string) EmbeddingService {
	if openAIEmbeddingFactory != nil {
		return openAIEmbeddingFactory(apiKey, model)
	}
	
	// Fallback - should not happen if internal package is imported
	Logger().Warn().Msg("No OpenAI embedding factory registered - using no-op service")
	return &noOpEmbeddingService{dimensions: 1536}
}

// NewOllamaEmbeddingService creates a new Ollama embedding service
func NewOllamaEmbeddingService(model, baseURL string) EmbeddingService {
	if ollamaEmbeddingFactory != nil {
		return ollamaEmbeddingFactory(model, baseURL)
	}
	
	// Fallback - should not happen if internal package is imported
	Logger().Warn().Msg("No Ollama embedding factory registered - using no-op service")
	return &noOpEmbeddingService{dimensions: 1024}
}

// NewDummyEmbeddingService creates a dummy embedding service for testing
func NewDummyEmbeddingService(dimensions int) EmbeddingService {
	if dummyEmbeddingFactory != nil {
		return dummyEmbeddingFactory(dimensions)
	}
	
	// Fallback - should not happen if internal package is imported
	Logger().Warn().Msg("No Dummy embedding factory registered - using no-op service")
	if dimensions <= 0 {
		dimensions = 1536
	}
	return &noOpEmbeddingService{dimensions: dimensions}
}

// Temporary no-op embedding service during refactoring
type noOpEmbeddingService struct {
	dimensions int
}

func (s *noOpEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Return zero vector
	return make([]float32, s.dimensions), nil
}

func (s *noOpEmbeddingService) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i := range embeddings {
		embeddings[i] = make([]float32, s.dimensions)
	}
	return embeddings, nil
}

func (s *noOpEmbeddingService) GetDimensions() int {
	return s.dimensions
}
