package providers

import (
	"context"
	"testing"

	"github.com/agenticgokit/agenticgokit/core"
	"github.com/stretchr/testify/assert"
)

type mockEmbedder struct {
	dimensions int
}

func (e *mockEmbedder) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	vec := make([]float32, e.dimensions)
	if len(text) > 0 {
		vec[0] = float32(len(text))
	}
	return vec, nil
}

func (e *mockEmbedder) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	res := make([][]float32, len(texts))
	for i, t := range texts {
		res[i], _ = e.GenerateEmbedding(ctx, t)
	}
	return res, nil
}

func (e *mockEmbedder) GetDimensions() int {
	return e.dimensions
}

func TestChromemProvider(t *testing.T) {
	ctx := context.Background()
	config := core.AgentMemoryConfig{
		Dimensions: 1536,
	}
	// Use mock embedder to avoid all-zero vectors
	embedder := &mockEmbedder{dimensions: 1536}

	m, err := NewChromemProvider(config, embedder)
	assert.NoError(t, err)
	assert.NotNil(t, m)

	// Test Store and Query
	err = m.Store(ctx, "The sky is blue", "sky", "color")
	assert.NoError(t, err)

	err = m.Store(ctx, "Grass is green", "grass", "color")
	assert.NoError(t, err)

	results, err := m.Query(ctx, "What color is the sky?")
	assert.NoError(t, err)

	t.Logf("Results length: %d", len(results))
	for i, r := range results {
		t.Logf("Result[%d]: content=%q, score=%f", i, r.Content, r.Score)
	}

	assert.NotEmpty(t, results)

	found := false
	for _, r := range results {
		if r.Content == "The sky is blue" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should find 'The sky is blue'")

	// Test KV (Remember/Recall)
	m.SetSession(ctx, "test-session")
	err = m.Remember(ctx, "user_name", "Alice")
	assert.NoError(t, err)

	val, err := m.Recall(ctx, "user_name")
	assert.NoError(t, err)
	assert.Equal(t, "Alice", val)

	// Test Chat History
	err = m.AddMessage(ctx, "user", "Hello")
	assert.NoError(t, err)
	err = m.AddMessage(ctx, "assistant", "Hi Alice!")
	assert.NoError(t, err)

	history, err := m.GetHistory(ctx)
	assert.NoError(t, err)
	assert.Len(t, history, 2)
	assert.Equal(t, "user", history[0].Role)
	assert.Equal(t, "assistant", history[1].Role)
}

func TestChromemProvider_RAG(t *testing.T) {
	ctx := context.Background()
	config := core.AgentMemoryConfig{
		Dimensions: 1536,
	}
	embedder := &mockEmbedder{dimensions: 1536}

	m, err := NewChromemProvider(config, embedder)
	assert.NoError(t, err)

	// Test IngestDocument
	doc := core.Document{
		ID:      "doc1",
		Title:   "Go Performance",
		Content: "Go is fast and efficient.",
		Source:  "wiki",
		Type:    "article",
	}
	err = m.IngestDocument(ctx, doc)
	assert.NoError(t, err)

	// Test SearchKnowledge
	results, err := m.SearchKnowledge(ctx, "Go performance")
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
	assert.Equal(t, "Go is fast and efficient.", results[0].Content)
	assert.Equal(t, "wiki", results[0].Source)

	// Test BuildContext
	ragCtx, err := m.BuildContext(ctx, "How is Go performance?")
	assert.NoError(t, err)
	assert.NotNil(t, ragCtx)
	assert.Contains(t, ragCtx.ContextText, "Knowledge: Go is fast and efficient.")
}
