// Package providers contains internal memory provider implementations.
package providers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
	"github.com/philippgille/chromem-go"
)

// ChromemProvider implements core.Memory using chromem-go
type ChromemProvider struct {
	db          *chromem.DB
	collection  *chromem.Collection
	history     map[string][]core.Message
	kv          map[string]map[string]any
	mu          sync.RWMutex
	dimensions  int
	embeddingFn func(ctx context.Context, text string) ([]float32, error)
}

// NewChromemProvider creates a new chromem-go memory provider
func NewChromemProvider(config core.AgentMemoryConfig, embedder core.EmbeddingService) (core.Memory, error) {
	var db *chromem.DB
	if config.Connection != "" && config.Connection != "memory" {
		var err error
		db, err = chromem.NewPersistentDB(config.Connection, false)
		if err != nil {
			return nil, fmt.Errorf("failed to create persistent chromem db: %w", err)
		}
	} else {
		db = chromem.NewDB()
	}

	// Create a default collection
	var chromemEmbedder chromem.EmbeddingFunc
	if embedder != nil {
		chromemEmbedder = func(ctx context.Context, text string) ([]float32, error) {
			return embedder.GenerateEmbedding(ctx, text)
		}
	}

	col, err := db.CreateCollection("memories", nil, chromemEmbedder)
	if err != nil {
		return nil, fmt.Errorf("failed to create chromem collection: %w", err)
	}

	return &ChromemProvider{
		db:         db,
		collection: col,
		history:    make(map[string][]core.Message),
		kv:         make(map[string]map[string]any),
		dimensions: config.Dimensions,
		embeddingFn: func(ctx context.Context, text string) ([]float32, error) {
			if embedder == nil {
				return make([]float32, config.Dimensions), nil
			}
			return embedder.GenerateEmbedding(ctx, text)
		},
	}, nil
}

// Store saves content to personal memory
func (m *ChromemProvider) Store(ctx context.Context, content string, tags ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessionID := core.GetSessionID(ctx)
	id := fmt.Sprintf("mem_%d", time.Now().UnixNano())
	metadata := map[string]string{
		"session_id": sessionID,
		"type":       "personal",
		"created_at": time.Now().Format(time.RFC3339),
	}
	for i, tag := range tags {
		metadata[fmt.Sprintf("tag_%d", i)] = tag
		if len(tag) > 0 {
			metadata["tag_"+tag] = "true"
		}
	}

	return m.collection.Add(ctx, []string{id}, nil, []map[string]string{metadata}, []string{content})
}

// Query searches personal memory
func (m *ChromemProvider) Query(ctx context.Context, query string, limit ...int) ([]core.Result, error) {
	l := 10
	if len(limit) > 0 {
		l = limit[0]
	}

	// Cap limit by number of documents in collection to avoid chromem error
	count := m.collection.Count()
	if l > count {
		l = count
	}

	if l == 0 {
		return []core.Result{}, nil
	}

	sessionID := core.GetSessionID(ctx)
	// We use the metadata filter. If it fails, we'll know.
	results, err := m.collection.Query(ctx, query, l, map[string]string{"session_id": sessionID}, nil)
	if err != nil {
		return nil, err
	}

	coreResults := make([]core.Result, len(results))
	for i, r := range results {
		coreResults[i] = core.Result{
			Content:   r.Content,
			Score:     r.Similarity,
			CreatedAt: time.Now(),
		}
	}

	return coreResults, nil
}

// Remember stores a key-value pair in memory
func (m *ChromemProvider) Remember(ctx context.Context, key string, value any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessionID := core.GetSessionID(ctx)
	if m.kv[sessionID] == nil {
		m.kv[sessionID] = make(map[string]any)
	}
	m.kv[sessionID][key] = value
	return nil
}

// Recall retrieves a key-value pair from memory
func (m *ChromemProvider) Recall(ctx context.Context, key string) (any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessionID := core.GetSessionID(ctx)
	if m.kv[sessionID] == nil {
		return nil, nil
	}
	return m.kv[sessionID][key], nil
}

// AddMessage adds a message to chat history
func (m *ChromemProvider) AddMessage(ctx context.Context, role, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessionID := core.GetSessionID(ctx)
	m.history[sessionID] = append(m.history[sessionID], core.Message{
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	})
	return nil
}

// GetHistory retrieves chat history
func (m *ChromemProvider) GetHistory(ctx context.Context, limit ...int) ([]core.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessionID := core.GetSessionID(ctx)
	h := m.history[sessionID]
	if len(limit) > 0 && limit[0] < len(h) {
		return h[len(h)-limit[0]:], nil
	}
	return h, nil
}

// NewSession creates a new session ID
func (m *ChromemProvider) NewSession() string {
	return core.GenerateSessionID()
}

// SetSession sets the current session ID
func (m *ChromemProvider) SetSession(ctx context.Context, sessionID string) context.Context {
	return core.WithMemory(ctx, m, sessionID)
}

// ClearSession clears memory for the current session
func (m *ChromemProvider) ClearSession(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessionID := core.GetSessionID(ctx)
	delete(m.history, sessionID)
	delete(m.kv, sessionID)
	// Note: chromem-go doesn't easily support deleting by metadata filter yet in a simple way
	return nil
}

// Close closes the memory provider
func (m *ChromemProvider) Close() error {
	return nil
}

// IngestDocument ingests a document into the knowledge base
func (m *ChromemProvider) IngestDocument(ctx context.Context, doc core.Document) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	metadata := map[string]string{
		"type":   "knowledge",
		"source": doc.Source,
		"title":  doc.Title,
	}

	return m.collection.Add(ctx, []string{doc.ID}, nil, []map[string]string{metadata}, []string{doc.Content})
}

// IngestDocuments ingests multiple documents
func (m *ChromemProvider) IngestDocuments(ctx context.Context, docs []core.Document) error {
	for _, doc := range docs {
		if err := m.IngestDocument(ctx, doc); err != nil {
			return err
		}
	}
	return nil
}

// SearchKnowledge searches the knowledge base
func (m *ChromemProvider) SearchKnowledge(ctx context.Context, query string, options ...core.SearchOption) ([]core.KnowledgeResult, error) {
	config := &core.SearchConfig{Limit: 10}
	for _, opt := range options {
		opt(config)
	}

	limit := config.Limit
	count := m.collection.Count()
	if limit > count {
		limit = count
	}

	if limit == 0 {
		return []core.KnowledgeResult{}, nil
	}

	results, err := m.collection.Query(ctx, query, limit, map[string]string{"type": "knowledge"}, nil)
	if err != nil {
		return nil, err
	}

	kResults := make([]core.KnowledgeResult, len(results))
	for i, r := range results {
		kResults[i] = core.KnowledgeResult{
			Content: r.Content,
			Score:   r.Similarity,
			Source:  r.Metadata["source"],
			Title:   r.Metadata["title"],
		}
	}

	return kResults, nil
}

// SearchAll searches both personal memory and knowledge base
func (m *ChromemProvider) SearchAll(ctx context.Context, query string, options ...core.SearchOption) (*core.HybridResult, error) {
	personal, err := m.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	knowledge, err := m.SearchKnowledge(ctx, query, options...)
	if err != nil {
		return nil, err
	}

	return &core.HybridResult{
		PersonalMemory: personal,
		Knowledge:      knowledge,
		Query:          query,
		TotalResults:   len(personal) + len(knowledge),
	}, nil
}

// BuildContext assembles RAG context
func (m *ChromemProvider) BuildContext(ctx context.Context, query string, options ...core.ContextOption) (*core.RAGContext, error) {
	history, _ := m.GetHistory(ctx, 10)
	searchRes, _ := m.SearchAll(ctx, query)

	contextText := ""
	for _, r := range searchRes.PersonalMemory {
		contextText += fmt.Sprintf("Memory: %s\n", r.Content)
	}
	for _, r := range searchRes.Knowledge {
		contextText += fmt.Sprintf("Knowledge: %s (Source: %s)\n", r.Content, r.Source)
	}

	return &core.RAGContext{
		Query:          query,
		PersonalMemory: searchRes.PersonalMemory,
		Knowledge:      searchRes.Knowledge,
		ChatHistory:    history,
		ContextText:    contextText,
		Timestamp:      time.Now(),
	}, nil
}
