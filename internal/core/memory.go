package agentflow

import "context"

// QueryResult represents a single item retrieved from a vector similarity search.
type QueryResult struct {
	ID       string         `json:"id"`       // Unique identifier of the stored item
	Metadata map[string]any `json:"metadata"` // Associated metadata stored with the vector
	Score    float32        `json:"score"`    // Similarity score (higher is typically better, but depends on metric)
	// Embedding []float32 `json:"embedding,omitempty"` // Optionally include the embedding itself
}

// VectorMemory defines the interface for storing and querying vector embeddings
// along with associated metadata. Implementations are expected to handle
// the specifics of interacting with different vector databases (e.g., Weaviate, PgVector).
type VectorMemory interface {
	// Store saves a vector embedding and its associated metadata with a unique ID.
	// If an item with the same ID already exists, it should typically be overwritten.
	// The context can be used for cancellation or timeouts.
	Store(ctx context.Context, id string, embedding []float32, metadata map[string]any) error

	// Query performs a similarity search using the given query embedding.
	// It returns the top 'k' most similar items based on the underlying distance metric.
	// The context can be used for cancellation or timeouts.
	Query(ctx context.Context, embedding []float32, topK int) ([]QueryResult, error)

	// TODO: Consider adding Delete(ctx context.Context, id string) error method?
	// TODO: Consider adding BatchStore(ctx context.Context, items []...) error for efficiency?
	// TODO: Consider adding Get(ctx context.Context, id string) (QueryResult, error) if direct lookup is needed?
}
