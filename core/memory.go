package core

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Memory is the central memory interface - replaces multiple interfaces
// Breaking change: Single, unified interface for all memory operations INCLUDING RAG
type Memory interface {
	// Personal memory operations (existing)
	Store(ctx context.Context, content string, tags ...string) error
	Query(ctx context.Context, query string, limit ...int) ([]Result, error)
	Remember(ctx context.Context, key string, value any) error
	Recall(ctx context.Context, key string) (any, error)

	// Chat history management (existing)
	AddMessage(ctx context.Context, role, content string) error
	GetHistory(ctx context.Context, limit ...int) ([]Message, error)

	// Session management (existing)
	NewSession() string
	SetSession(ctx context.Context, sessionID string) context.Context
	ClearSession(ctx context.Context) error
	Close() error

	// NEW: RAG-Enhanced Knowledge Base Operations
	// Breaking change: Add RAG capabilities to core memory interface
	IngestDocument(ctx context.Context, doc Document) error
	IngestDocuments(ctx context.Context, docs []Document) error
	SearchKnowledge(ctx context.Context, query string, options ...SearchOption) ([]KnowledgeResult, error)

	// NEW: Hybrid Search (Personal Memory + Knowledge Base)
	SearchAll(ctx context.Context, query string, options ...SearchOption) (*HybridResult, error)

	// NEW: RAG Context Assembly for LLM Prompts
	BuildContext(ctx context.Context, query string, options ...ContextOption) (*RAGContext, error)
}

// Result - simplified result structure
type Result struct {
	Content   string    `json:"content"`
	Score     float32   `json:"score"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Message - conversation message
type Message struct {
	Role      string    `json:"role"` // user, assistant, system
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// NEW: RAG-Enhanced Types for Knowledge Base and Document Management

// Document structure for knowledge ingestion
type Document struct {
	ID         string         `json:"id"`
	Title      string         `json:"title,omitempty"`
	Content    string         `json:"content"`
	Source     string         `json:"source,omitempty"` // URL, file path, etc.
	Type       DocumentType   `json:"type,omitempty"`   // PDF, TXT, WEB, etc.
	Metadata   map[string]any `json:"metadata,omitempty"`
	Tags       []string       `json:"tags,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at,omitempty"`
	ChunkIndex int            `json:"chunk_index,omitempty"` // For chunked documents
	ChunkTotal int            `json:"chunk_total,omitempty"`
}

// DocumentType represents the type of document being ingested
type DocumentType string

const (
	DocumentTypePDF      DocumentType = "pdf"
	DocumentTypeText     DocumentType = "txt"
	DocumentTypeMarkdown DocumentType = "md"
	DocumentTypeWeb      DocumentType = "web"
	DocumentTypeCode     DocumentType = "code"
	DocumentTypeJSON     DocumentType = "json"
)

// KnowledgeResult represents search results from the knowledge base
type KnowledgeResult struct {
	Content    string         `json:"content"`
	Score      float32        `json:"score"`
	Source     string         `json:"source"`
	Title      string         `json:"title,omitempty"`
	DocumentID string         `json:"document_id"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	Tags       []string       `json:"tags,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	ChunkIndex int            `json:"chunk_index,omitempty"`
}

// HybridResult combines personal memory and knowledge base search results
type HybridResult struct {
	PersonalMemory []Result          `json:"personal_memory"`
	Knowledge      []KnowledgeResult `json:"knowledge"`
	Query          string            `json:"query"`
	TotalResults   int               `json:"total_results"`
	SearchTime     time.Duration     `json:"search_time"`
}

// RAGContext provides assembled context for LLM prompts
type RAGContext struct {
	Query          string            `json:"query"`
	PersonalMemory []Result          `json:"personal_memory"`
	Knowledge      []KnowledgeResult `json:"knowledge"`
	ChatHistory    []Message         `json:"chat_history"`
	ContextText    string            `json:"context_text"` // Formatted for LLM
	Sources        []string          `json:"sources"`      // Source attribution
	TokenCount     int               `json:"token_count"`  // Estimated tokens
	Timestamp      time.Time         `json:"timestamp"`
}

// Search and context configuration options
type SearchOption func(*SearchConfig)
type ContextOption func(*ContextConfig)

type SearchConfig struct {
	Limit            int            `json:"limit"`
	ScoreThreshold   float32        `json:"score_threshold"`
	Sources          []string       `json:"sources"`           // Filter by source
	DocumentTypes    []DocumentType `json:"document_types"`    // Filter by type
	Tags             []string       `json:"tags"`              // Filter by tags
	DateRange        *DateRange     `json:"date_range"`        // Filter by date
	HybridWeight     float32        `json:"hybrid_weight"`     // Semantic vs keyword weight
	IncludePersonal  bool           `json:"include_personal"`  // Include personal memory
	IncludeKnowledge bool           `json:"include_knowledge"` // Include knowledge base
}

type ContextConfig struct {
	MaxTokens       int     `json:"max_tokens"`       // Context size limit
	PersonalWeight  float32 `json:"personal_weight"`  // Weight for personal memory
	KnowledgeWeight float32 `json:"knowledge_weight"` // Weight for knowledge base
	HistoryLimit    int     `json:"history_limit"`    // Chat history messages
	IncludeSources  bool    `json:"include_sources"`  // Include source attribution
	FormatTemplate  string  `json:"format_template"`  // Custom context formatting
}

type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// AgentMemoryConfig - enhanced configuration for agent memory storage with RAG support
type AgentMemoryConfig struct {
	// Core memory settings
	Provider   string `toml:"provider"`    // pgvector, weaviate, memory
	Connection string `toml:"connection"`  // postgres://..., http://..., or "memory"
	MaxResults int    `toml:"max_results"` // default: 10
	Dimensions int    `toml:"dimensions"`  // default: 1536
	AutoEmbed  bool   `toml:"auto_embed"`  // default: true

	// RAG-enhanced settings
	EnableKnowledgeBase     bool    `toml:"enable_knowledge_base"`     // default: true
	KnowledgeMaxResults     int     `toml:"knowledge_max_results"`     // default: 20
	KnowledgeScoreThreshold float32 `toml:"knowledge_score_threshold"` // default: 0.7
	ChunkSize               int     `toml:"chunk_size"`                // default: 1000
	ChunkOverlap            int     `toml:"chunk_overlap"`             // default: 200

	// RAG context assembly settings
	EnableRAG           bool    `toml:"enable_rag"`             // default: true
	RAGMaxContextTokens int     `toml:"rag_max_context_tokens"` // default: 4000
	RAGPersonalWeight   float32 `toml:"rag_personal_weight"`    // default: 0.3
	RAGKnowledgeWeight  float32 `toml:"rag_knowledge_weight"`   // default: 0.7
	RAGIncludeSources   bool    `toml:"rag_include_sources"`    // default: true

	// Document processing settings
	Documents DocumentConfig `toml:"documents"`

	// Embedding service settings
	Embedding EmbeddingConfig `toml:"embedding"`

	// Search settings
	Search SearchConfigToml `toml:"search"`
}

// DocumentConfig represents document processing configuration
type DocumentConfig struct {
	AutoChunk                bool     `toml:"auto_chunk"`                 // default: true
	SupportedTypes           []string `toml:"supported_types"`            // default: ["pdf", "txt", "md", "web", "code"]
	MaxFileSize              string   `toml:"max_file_size"`              // default: "10MB"
	EnableMetadataExtraction bool     `toml:"enable_metadata_extraction"` // default: true
	EnableURLScraping        bool     `toml:"enable_url_scraping"`        // default: true
}

// EmbeddingConfig represents embedding service configuration
type EmbeddingConfig struct {
	Provider        string `toml:"provider"`         // azure, openai, local
	Model           string `toml:"model"`            // text-embedding-ada-002, etc.
	CacheEmbeddings bool   `toml:"cache_embeddings"` // default: true
	APIKey          string `toml:"api_key"`          // API key for service
	Endpoint        string `toml:"endpoint"`         // Custom endpoint
	MaxBatchSize    int    `toml:"max_batch_size"`   // default: 100
	TimeoutSeconds  int    `toml:"timeout_seconds"`  // default: 30
}

// SearchConfigToml represents search configuration
type SearchConfigToml struct {
	HybridSearch         bool    `toml:"hybrid_search"`          // default: true
	KeywordWeight        float32 `toml:"keyword_weight"`         // default: 0.3
	SemanticWeight       float32 `toml:"semantic_weight"`        // default: 0.7
	EnableReranking      bool    `toml:"enable_reranking"`       // default: false
	RerankingModel       string  `toml:"reranking_model"`        // Model for reranking
	EnableQueryExpansion bool    `toml:"enable_query_expansion"` // default: false
}

// NewMemory creates a new memory instance based on configuration
func NewMemory(config AgentMemoryConfig) (Memory, error) {
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
		config.KnowledgeScoreThreshold = 0.7
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
		config.Embedding.Provider = "azure"
	}
	if config.Embedding.Model == "" {
		config.Embedding.Model = "text-embedding-ada-002"
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

	// Set boolean defaults (these are false by default in Go)
	if !config.EnableKnowledgeBase {
		config.EnableKnowledgeBase = true
	}
	if !config.EnableRAG {
		config.EnableRAG = true
	}
	if !config.RAGIncludeSources {
		config.RAGIncludeSources = true
	}
	if !config.Documents.AutoChunk {
		config.Documents.AutoChunk = true
	}
	if !config.Documents.EnableMetadataExtraction {
		config.Documents.EnableMetadataExtraction = true
	}
	if !config.Documents.EnableURLScraping {
		config.Documents.EnableURLScraping = true
	}
	if !config.Embedding.CacheEmbeddings {
		config.Embedding.CacheEmbeddings = true
	}
	if !config.Search.HybridSearch {
		config.Search.HybridSearch = true
	}

	switch config.Provider {
	case "memory":
		return newInMemoryProvider(config)
	case "pgvector":
		return newPgVectorProvider(config)
	case "weaviate":
		return newWeaviateProvider(config)
	default:
		return nil, fmt.Errorf("unsupported memory provider: %s", config.Provider)
	}
}

// QuickMemory creates an in-memory provider for quick testing
func QuickMemory() Memory {
	config := AgentMemoryConfig{
		Provider:   "memory",
		Connection: "memory",
		MaxResults: 10,
		AutoEmbed:  true,
	}

	memory, err := NewMemory(config)
	if err != nil {
		// Return no-op memory instead of panicking
		return &NoOpMemory{}
	}

	return memory
}

// Provider implementations

// InMemoryProvider - fast in-memory implementation for development/testing
type InMemoryProvider struct {
	mutex     sync.RWMutex
	vectors   map[string]vectorEntry
	keyValues map[string]any
	messages  map[string][]Message // sessionID -> messages
	sessionID string
	config    AgentMemoryConfig

	// NEW: Knowledge base storage (global, not session-scoped)
	knowledge map[string]knowledgeEntry // documentID -> document content
	documents map[string]Document       // documentID -> document metadata
}

type vectorEntry struct {
	Content   string
	Tags      []string
	CreatedAt time.Time
	// For in-memory, we'll use simple string matching instead of real embeddings
}

// NEW: Knowledge base entry for in-memory storage
type knowledgeEntry struct {
	Content   string
	Document  Document
	CreatedAt time.Time
}

func newInMemoryProvider(config AgentMemoryConfig) (Memory, error) {
	return &InMemoryProvider{
		vectors:   make(map[string]vectorEntry),
		keyValues: make(map[string]any),
		messages:  make(map[string][]Message),
		sessionID: "default",
		config:    config,
		knowledge: make(map[string]knowledgeEntry),
		documents: make(map[string]Document),
	}, nil
}

func (m *InMemoryProvider) Store(ctx context.Context, content string, tags ...string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sessionID := GetSessionID(ctx)
	key := sessionID + ":" + generateID()

	m.vectors[key] = vectorEntry{
		Content:   content,
		Tags:      tags,
		CreatedAt: time.Now(),
	}

	return nil
}

func (m *InMemoryProvider) Query(ctx context.Context, query string, limit ...int) ([]Result, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	maxResults := m.config.MaxResults
	if len(limit) > 0 && limit[0] > 0 {
		maxResults = limit[0]
	}

	sessionID := GetSessionID(ctx)
	var results []Result

	// Simple text matching for in-memory implementation
	sessionPrefix := sessionID + ":"
	for key, entry := range m.vectors {
		if strings.HasPrefix(key, sessionPrefix) {
			score := calculateScore(entry.Content, query)
			tagScore := float32(0)

			// Check tag matching
			for _, tag := range entry.Tags {
				if tagScore < calculateScore(tag, query) {
					tagScore = calculateScore(tag, query)
				}
			}

			// Use the best score
			finalScore := score
			if tagScore > finalScore {
				finalScore = tagScore
			}

			// Only include if there's a reasonable match
			if finalScore > 0.1 {
				results = append(results, Result{
					Content:   entry.Content,
					Score:     finalScore,
					Tags:      entry.Tags,
					CreatedAt: entry.CreatedAt,
				})

				if len(results) >= maxResults {
					break
				}
			}
		}
	}

	// Sort results by score (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

func (m *InMemoryProvider) Remember(ctx context.Context, key string, value any) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sessionID := GetSessionID(ctx)
	fullKey := sessionID + ":" + key
	m.keyValues[fullKey] = value

	return nil
}

func (m *InMemoryProvider) Recall(ctx context.Context, key string) (any, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	sessionID := GetSessionID(ctx)
	fullKey := sessionID + ":" + key

	if value, exists := m.keyValues[fullKey]; exists {
		return value, nil
	}

	return nil, nil
}

func (m *InMemoryProvider) AddMessage(ctx context.Context, role, content string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sessionID := GetSessionID(ctx)

	message := Message{
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	}

	m.messages[sessionID] = append(m.messages[sessionID], message)

	return nil
}

func (m *InMemoryProvider) GetHistory(ctx context.Context, limit ...int) ([]Message, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	sessionID := GetSessionID(ctx)
	messages := m.messages[sessionID]

	if len(limit) > 0 && limit[0] > 0 && limit[0] < len(messages) {
		// Return last N messages
		start := len(messages) - limit[0]
		return messages[start:], nil
	}

	return messages, nil
}

func (m *InMemoryProvider) NewSession() string {
	return generateID()
}

func (m *InMemoryProvider) SetSession(ctx context.Context, sessionID string) context.Context {
	return WithMemory(ctx, m, sessionID)
}

func (m *InMemoryProvider) ClearSession(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sessionID := GetSessionID(ctx)
	sessionPrefix := sessionID + ":"

	// Clear vectors for this session
	for key := range m.vectors {
		if strings.HasPrefix(key, sessionPrefix) {
			delete(m.vectors, key)
		}
	}

	// Clear key-values for this session
	for key := range m.keyValues {
		if strings.HasPrefix(key, sessionPrefix) {
			delete(m.keyValues, key)
		}
	}

	// Clear messages for this session
	delete(m.messages, sessionID)

	return nil
}

func (m *InMemoryProvider) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Clear all data
	m.vectors = make(map[string]vectorEntry)
	m.keyValues = make(map[string]any)
	m.messages = make(map[string][]Message)
	m.knowledge = make(map[string]knowledgeEntry)
	m.documents = make(map[string]Document)

	return nil
}

// NEW: RAG-Enhanced Knowledge Base Operations Implementation

func (m *InMemoryProvider) IngestDocument(ctx context.Context, doc Document) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Generate ID if not provided
	if doc.ID == "" {
		doc.ID = generateID()
	}

	// Set timestamps
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now()
	}
	doc.UpdatedAt = time.Now()

	// Store document metadata
	m.documents[doc.ID] = doc

	// Store document content for search
	m.knowledge[doc.ID] = knowledgeEntry{
		Content:   doc.Content,
		Document:  doc,
		CreatedAt: doc.CreatedAt,
	}

	return nil
}

func (m *InMemoryProvider) IngestDocuments(ctx context.Context, docs []Document) error {
	for _, doc := range docs {
		if err := m.IngestDocument(ctx, doc); err != nil {
			return fmt.Errorf("failed to ingest document %s: %w", doc.ID, err)
		}
	}
	return nil
}

func (m *InMemoryProvider) SearchKnowledge(ctx context.Context, query string, options ...SearchOption) ([]KnowledgeResult, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Apply search options
	config := &SearchConfig{
		Limit:            m.config.MaxResults,
		ScoreThreshold:   0.0,
		IncludeKnowledge: true,
	}
	for _, opt := range options {
		opt(config)
	}

	var results []KnowledgeResult

	// Search through knowledge base
	for docID, entry := range m.knowledge {
		score := calculateScore(entry.Content, query)
		titleScore := calculateScore(entry.Document.Title, query)
		tagScore := float32(0)

		// Check tag matching
		for _, tag := range entry.Document.Tags {
			if tagScore < calculateScore(tag, query) {
				tagScore = calculateScore(tag, query)
			}
		}

		// Use the best score from content, title, or tags
		finalScore := score
		if titleScore > finalScore {
			finalScore = titleScore
		}
		if tagScore > finalScore {
			finalScore = tagScore
		}

		// Only include if score is above threshold
		if finalScore < config.ScoreThreshold {
			continue
		}

		// Apply filters
		if len(config.Sources) > 0 && !contains(entry.Document.Source, config.Sources[0]) {
			continue
		}
		if len(config.DocumentTypes) > 0 && !containsDocumentType(config.DocumentTypes, entry.Document.Type) {
			continue
		}
		if len(config.Tags) > 0 && !hasAnyTag(entry.Document.Tags, config.Tags) {
			continue
		}

		results = append(results, KnowledgeResult{
			Content:    entry.Content,
			Score:      finalScore,
			Source:     entry.Document.Source,
			Title:      entry.Document.Title,
			DocumentID: docID,
			Metadata:   entry.Document.Metadata,
			Tags:       entry.Document.Tags,
			CreatedAt:  entry.CreatedAt,
			ChunkIndex: entry.Document.ChunkIndex,
		})

		if len(results) >= config.Limit {
			break
		}
	}

	// Sort results by score (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

func (m *InMemoryProvider) SearchAll(ctx context.Context, query string, options ...SearchOption) (*HybridResult, error) {
	start := time.Now()

	// Apply search options
	config := &SearchConfig{
		Limit:            m.config.MaxResults,
		ScoreThreshold:   0.0,
		IncludePersonal:  true,
		IncludeKnowledge: true,
	}
	for _, opt := range options {
		opt(config)
	}

	result := &HybridResult{
		Query:          query,
		PersonalMemory: []Result{},
		Knowledge:      []KnowledgeResult{},
	}

	// Search personal memory if enabled
	if config.IncludePersonal {
		personalResults, err := m.Query(ctx, query, config.Limit)
		if err != nil {
			return nil, fmt.Errorf("failed to search personal memory: %w", err)
		}
		result.PersonalMemory = personalResults
	}

	// Search knowledge base if enabled
	if config.IncludeKnowledge {
		knowledgeResults, err := m.SearchKnowledge(ctx, query, options...)
		if err != nil {
			return nil, fmt.Errorf("failed to search knowledge base: %w", err)
		}
		result.Knowledge = knowledgeResults
	}

	result.TotalResults = len(result.PersonalMemory) + len(result.Knowledge)
	result.SearchTime = time.Since(start)

	return result, nil
}

func (m *InMemoryProvider) BuildContext(ctx context.Context, query string, options ...ContextOption) (*RAGContext, error) {
	// Apply context options
	config := &ContextConfig{
		MaxTokens:       4000,
		PersonalWeight:  0.5,
		KnowledgeWeight: 0.5,
		HistoryLimit:    5,
		IncludeSources:  true,
		FormatTemplate:  "", // Use default formatting
	}
	for _, opt := range options {
		opt(config)
	}

	// Get hybrid search results
	searchResults, err := m.SearchAll(ctx, query,
		WithLimit(config.MaxTokens/100), // Rough estimate: 100 tokens per result
		WithIncludePersonal(config.PersonalWeight > 0),
		WithIncludeKnowledge(config.KnowledgeWeight > 0),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search for context: %w", err)
	}

	// Get chat history
	history, err := m.GetHistory(ctx, config.HistoryLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat history: %w", err)
	}

	// Build context text
	contextText := m.formatContextText(query, searchResults, history, config)

	// Collect sources
	sources := []string{}
	for _, result := range searchResults.Knowledge {
		if result.Source != "" {
			sources = append(sources, result.Source)
		}
	}

	// Remove duplicates
	sources = removeDuplicates(sources)

	return &RAGContext{
		Query:          query,
		PersonalMemory: searchResults.PersonalMemory,
		Knowledge:      searchResults.Knowledge,
		ChatHistory:    history,
		ContextText:    contextText,
		Sources:        sources,
		TokenCount:     estimateTokenCount(contextText),
		Timestamp:      time.Now(),
	}, nil
}

// Helper functions for RAG implementation

func (m *InMemoryProvider) formatContextText(query string, results *HybridResult, history []Message, config *ContextConfig) string {
	if config.FormatTemplate != "" {
		// TODO: Implement custom template formatting
		return config.FormatTemplate
	}

	var builder strings.Builder

	// Add query
	builder.WriteString(fmt.Sprintf("Query: %s\n\n", query))

	// Add personal memory context
	if len(results.PersonalMemory) > 0 {
		builder.WriteString("Personal Memory:\n")
		for i, result := range results.PersonalMemory {
			builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, result.Content))
		}
		builder.WriteString("\n")
	}

	// Add knowledge base context
	if len(results.Knowledge) > 0 {
		builder.WriteString("Knowledge Base:\n")
		for i, result := range results.Knowledge {
			source := ""
			if config.IncludeSources && result.Source != "" {
				source = fmt.Sprintf(" (Source: %s)", result.Source)
			}
			builder.WriteString(fmt.Sprintf("%d. %s%s\n", i+1, result.Content, source))
		}
		builder.WriteString("\n")
	}

	// Add recent chat history
	if len(history) > 0 {
		builder.WriteString("Recent Conversation:\n")
		for _, msg := range history {
			builder.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
	}

	return builder.String()
}

func containsDocumentType(types []DocumentType, docType DocumentType) bool {
	for _, t := range types {
		if t == docType {
			return true
		}
	}
	return false
}

func hasAnyTag(documentTags, searchTags []string) bool {
	for _, searchTag := range searchTags {
		for _, docTag := range documentTags {
			if contains(docTag, searchTag) {
				return true
			}
		}
	}
	return false
}

func calculateScore(content, query string) float32 {
	content = strings.ToLower(content)
	query = strings.ToLower(query)

	// Exact match = 1.0
	if content == query {
		return 1.0
	}

	// Full substring match = 0.9
	if strings.Contains(content, query) {
		return 0.9
	}

	// Word-based matching
	contentWords := strings.Fields(content)
	queryWords := strings.Fields(query)

	matches := 0
	for _, qWord := range queryWords {
		for _, cWord := range contentWords {
			if strings.Contains(cWord, qWord) || strings.Contains(qWord, cWord) {
				matches++
				break
			}
		}
	}

	if matches > 0 {
		return 0.5 + (float32(matches)/float32(len(queryWords)))*0.4
	}

	return 0.1
}

func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

func estimateTokenCount(text string) int {
	// Rough estimation: ~4 characters per token
	return len(text) / 4
}

// Search option constructors
func WithLimit(limit int) SearchOption {
	return func(config *SearchConfig) {
		config.Limit = limit
	}
}

func WithScoreThreshold(threshold float32) SearchOption {
	return func(config *SearchConfig) {
		config.ScoreThreshold = threshold
	}
}

func WithSources(sources []string) SearchOption {
	return func(config *SearchConfig) {
		config.Sources = sources
	}
}

func WithDocumentTypes(types []DocumentType) SearchOption {
	return func(config *SearchConfig) {
		config.DocumentTypes = types
	}
}

func WithTags(tags []string) SearchOption {
	return func(config *SearchConfig) {
		config.Tags = tags
	}
}

func WithIncludePersonal(include bool) SearchOption {
	return func(config *SearchConfig) {
		config.IncludePersonal = include
	}
}

func WithIncludeKnowledge(include bool) SearchOption {
	return func(config *SearchConfig) {
		config.IncludeKnowledge = include
	}
}

// Context option constructors
func WithMaxTokens(maxTokens int) ContextOption {
	return func(config *ContextConfig) {
		config.MaxTokens = maxTokens
	}
}

func WithPersonalWeight(weight float32) ContextOption {
	return func(config *ContextConfig) {
		config.PersonalWeight = weight
	}
}

func WithKnowledgeWeight(weight float32) ContextOption {
	return func(config *ContextConfig) {
		config.KnowledgeWeight = weight
	}
}

func WithHistoryLimit(limit int) ContextOption {
	return func(config *ContextConfig) {
		config.HistoryLimit = limit
	}
}

func WithIncludeSources(include bool) ContextOption {
	return func(config *ContextConfig) {
		config.IncludeSources = include
	}
}

func WithFormatTemplate(template string) ContextOption {
	return func(config *ContextConfig) {
		config.FormatTemplate = template
	}
}

// PgVector provider - production-ready PostgreSQL with pgvector
func newPgVectorProvider(config AgentMemoryConfig) (Memory, error) {
	// TODO: Implement PgVector provider
	// For now, return an error indicating it needs implementation
	return nil, fmt.Errorf("PgVector provider not yet implemented - use 'memory' provider for now")
}

// PgVectorProvider - production-ready PostgreSQL with pgvector (stub)
type PgVectorProvider struct {
	config AgentMemoryConfig
}

func (p *PgVectorProvider) Store(ctx context.Context, content string, tags ...string) error {
	return fmt.Errorf("PgVector provider not yet implemented")
}

func (p *PgVectorProvider) Query(ctx context.Context, query string, limit ...int) ([]Result, error) {
	return nil, fmt.Errorf("PgVector provider not yet implemented")
}

func (p *PgVectorProvider) Remember(ctx context.Context, key string, value any) error {
	return fmt.Errorf("PgVector provider not yet implemented")
}

func (p *PgVectorProvider) Recall(ctx context.Context, key string) (any, error) {
	return nil, fmt.Errorf("PgVector provider not yet implemented")
}

func (p *PgVectorProvider) AddMessage(ctx context.Context, role, content string) error {
	return fmt.Errorf("PgVector provider not yet implemented")
}

func (p *PgVectorProvider) GetHistory(ctx context.Context, limit ...int) ([]Message, error) {
	return nil, fmt.Errorf("PgVector provider not yet implemented")
}

func (p *PgVectorProvider) NewSession() string {
	return generateID()
}

func (p *PgVectorProvider) SetSession(ctx context.Context, sessionID string) context.Context {
	return WithMemory(ctx, p, sessionID)
}

func (p *PgVectorProvider) ClearSession(ctx context.Context) error {
	return fmt.Errorf("PgVector provider not yet implemented")
}

func (p *PgVectorProvider) Close() error {
	return fmt.Errorf("PgVector provider not yet implemented")
}

func (p *PgVectorProvider) IngestDocument(ctx context.Context, doc Document) error {
	return fmt.Errorf("PgVector provider not yet implemented")
}

func (p *PgVectorProvider) IngestDocuments(ctx context.Context, docs []Document) error {
	return fmt.Errorf("PgVector provider not yet implemented")
}

func (p *PgVectorProvider) SearchKnowledge(ctx context.Context, query string, options ...SearchOption) ([]KnowledgeResult, error) {
	return nil, fmt.Errorf("PgVector provider not yet implemented")
}

func (p *PgVectorProvider) SearchAll(ctx context.Context, query string, options ...SearchOption) (*HybridResult, error) {
	return nil, fmt.Errorf("PgVector provider not yet implemented")
}

func (p *PgVectorProvider) BuildContext(ctx context.Context, query string, options ...ContextOption) (*RAGContext, error) {
	return nil, fmt.Errorf("PgVector provider not yet implemented")
}

// Weaviate provider - production-ready vector database
func newWeaviateProvider(config AgentMemoryConfig) (Memory, error) {
	// TODO: Implement Weaviate provider
	// For now, return an error indicating it needs implementation
	return nil, fmt.Errorf("Weaviate provider not yet implemented - use 'memory' provider for now")
}

// WeaviateProvider - production-ready vector database (stub)
type WeaviateProvider struct {
	config AgentMemoryConfig
}

func (w *WeaviateProvider) Store(ctx context.Context, content string, tags ...string) error {
	return fmt.Errorf("Weaviate provider not yet implemented")
}

func (w *WeaviateProvider) Query(ctx context.Context, query string, limit ...int) ([]Result, error) {
	return nil, fmt.Errorf("Weaviate provider not yet implemented")
}

func (w *WeaviateProvider) Remember(ctx context.Context, key string, value any) error {
	return fmt.Errorf("Weaviate provider not yet implemented")
}

func (w *WeaviateProvider) Recall(ctx context.Context, key string) (any, error) {
	return nil, fmt.Errorf("Weaviate provider not yet implemented")
}

func (w *WeaviateProvider) AddMessage(ctx context.Context, role, content string) error {
	return fmt.Errorf("Weaviate provider not yet implemented")
}

func (w *WeaviateProvider) GetHistory(ctx context.Context, limit ...int) ([]Message, error) {
	return nil, fmt.Errorf("Weaviate provider not yet implemented")
}

func (w *WeaviateProvider) NewSession() string {
	return generateID()
}

func (w *WeaviateProvider) SetSession(ctx context.Context, sessionID string) context.Context {
	return WithMemory(ctx, w, sessionID)
}

func (w *WeaviateProvider) ClearSession(ctx context.Context) error {
	return fmt.Errorf("Weaviate provider not yet implemented")
}

func (w *WeaviateProvider) Close() error {
	return fmt.Errorf("Weaviate provider not yet implemented")
}

func (w *WeaviateProvider) IngestDocument(ctx context.Context, doc Document) error {
	return fmt.Errorf("Weaviate provider not yet implemented")
}

func (w *WeaviateProvider) IngestDocuments(ctx context.Context, docs []Document) error {
	return fmt.Errorf("Weaviate provider not yet implemented")
}

func (w *WeaviateProvider) SearchKnowledge(ctx context.Context, query string, options ...SearchOption) ([]KnowledgeResult, error) {
	return nil, fmt.Errorf("Weaviate provider not yet implemented")
}

func (w *WeaviateProvider) SearchAll(ctx context.Context, query string, options ...SearchOption) (*HybridResult, error) {
	return nil, fmt.Errorf("Weaviate provider not yet implemented")
}

func (w *WeaviateProvider) BuildContext(ctx context.Context, query string, options ...ContextOption) (*RAGContext, error) {
	return nil, fmt.Errorf("Weaviate provider not yet implemented")
}

// Utility functions
func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func contains(text, query string) bool {
	if len(text) == 0 || len(query) == 0 {
		return false
	}

	// Convert to lowercase for case-insensitive matching
	text = strings.ToLower(text)
	query = strings.ToLower(query)

	// Direct substring match
	return strings.Contains(text, query)
}

func findSubstring(text, query string) bool {
	if len(query) > len(text) {
		return false
	}
	for i := 0; i <= len(text)-len(query); i++ {
		if text[i:i+len(query)] == query {
			return true
		}
	}
	return false
}

func containsAnyTag(tags []string, query string) bool {
	for _, tag := range tags {
		if contains(tag, query) {
			return true
		}
	}
	return false
}
