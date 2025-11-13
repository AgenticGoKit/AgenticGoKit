// Package providers contains internal embedding service implementations.
package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
)

// OllamaEmbeddingService implements EmbeddingService using Ollama API
type OllamaEmbeddingService struct {
	model      string
	baseURL    string
	dimensions int
	client     *http.Client
}

// Ollama API request/response structures
type ollamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ollamaEmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

// NewOllamaEmbeddingService creates a new Ollama embedding service
func NewOllamaEmbeddingService(model, baseURL string) core.EmbeddingService {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	dimensions := 1024 // Default for mxbai-embed-large
	if strings.Contains(model, "mxbai-embed-large") {
		dimensions = 1024
	} else if strings.Contains(model, "nomic-embed") {
		dimensions = 768
	}

	return &OllamaEmbeddingService{
		model:      model,
		baseURL:    baseURL,
		dimensions: dimensions,
		client: &http.Client{
			Timeout: 60 * time.Second, // Ollama can be slower
		},
	}
}

// GenerateEmbedding generates a single embedding
func (s *OllamaEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Prepare request
	request := ollamaEmbeddingRequest{
		Model:  s.model,
		Prompt: text,
	}

	// Marshal request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/api/embeddings", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse response
	var response ollamaEmbeddingResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Embedding) == 0 {
		return nil, fmt.Errorf("no embedding returned from Ollama")
	}

	return response.Embedding, nil
}

// GenerateEmbeddings generates multiple embeddings in batch (sequential for Ollama)
func (s *OllamaEmbeddingService) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		embedding, err := s.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// GetDimensions returns the embedding dimensions
func (s *OllamaEmbeddingService) GetDimensions() int {
	return s.dimensions
}
