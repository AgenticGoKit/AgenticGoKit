// Package providers contains internal embedding service implementations.
package providers

import (
	"context"
	mathrand "math/rand"

	"github.com/agenticgokit/agenticgokit/core"
)

// DummyEmbeddingService for testing and development
type DummyEmbeddingService struct {
	dimensions int
}

// NewDummyEmbeddingService creates a dummy embedding service for testing
func NewDummyEmbeddingService(dimensions int) core.EmbeddingService {
	if dimensions <= 0 {
		dimensions = 1536 // Default
	}
	return &DummyEmbeddingService{dimensions: dimensions}
}

// GenerateEmbedding generates a dummy embedding
func (s *DummyEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	embedding := make([]float32, s.dimensions)
	// Use text hash for consistent dummy embeddings
	textHash := simpleHash(text)
	rng := mathrand.New(mathrand.NewSource(int64(textHash)))
	for i := range embedding {
		embedding[i] = rng.Float32()*2 - 1 // Random value between -1 and 1
	}
	return embedding, nil
}

// GenerateEmbeddings generates multiple dummy embeddings
func (s *DummyEmbeddingService) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for idx, text := range texts {
		embedding, err := s.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[idx] = embedding
	}
	return embeddings, nil
}

// GetDimensions returns the embedding dimensions
func (s *DummyEmbeddingService) GetDimensions() int {
	return s.dimensions
}

// Simple hash function for consistent dummy embeddings
func simpleHash(s string) uint32 {
	var hash uint32 = 5381
	for _, c := range s {
		hash = ((hash << 5) + hash) + uint32(c)
	}
	return hash
}
