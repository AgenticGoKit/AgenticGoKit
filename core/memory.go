// Package core provides essential memory interfaces and types for AgentFlow.
package core

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"
	"time"
)

// =============================================================================
// MEMORY INTERFACE
// =============================================================================

// Memory is the central memory interface for all memory operations including RAG
type Memory interface {
	// Personal memory operations
	Store(ctx context.Context, content string, tags ...string) error
	Query(ctx context.Context, query string, limit ...int) ([]Result, error)
	Remember(ctx context.Context, key string, value any) error
	Recall(ctx context.Context, key string) (any, error)

	// Chat history management
	AddMessage(ctx context.Context, role, content string) error
	GetHistory(ctx context.Context, limit ...int) ([]Message, error)

	// Session management
	NewSession() string
	SetSession(ctx context.Context, sessionID string) context.Context
	ClearSession(ctx context.Context) error
	Close() error

	// RAG-Enhanced Knowledge Base Operations
	IngestDocument(ctx context.Context, doc Document) error
	IngestDocuments(ctx context.Context, docs []Document) error
	SearchKnowledge(ctx context.Context, query string, options ...SearchOption) ([]KnowledgeResult, error)

	// Hybrid Search (Personal Memory + Knowledge Base)
	SearchAll(ctx context.Context, query string, options ...SearchOption) (*HybridResult, error)

	// RAG Context Assembly for LLM Prompts
	BuildContext(ctx context.Context, query string, options ...ContextOption) (*RAGContext, error)
}

// EmbeddingService interface for generating embeddings
type EmbeddingService interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
	GetDimensions() int
}

// =============================================================================
// MEMORY TYPES
// =============================================================================

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

// =============================================================================
// CONFIGURATION TYPES
// =============================================================================

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
	Provider        string `toml:"provider"`         // openai, ollama, dummy
	Model           string `toml:"model"`            // text-embedding-ada-002, mxbai-embed-large, etc.
	CacheEmbeddings bool   `toml:"cache_embeddings"` // default: true
	APIKey          string `toml:"api_key"`          // API key for service
	BaseURL         string `toml:"base_url"`         // Base URL for service (e.g., Ollama endpoint)
	Endpoint        string `toml:"endpoint"`         // Custom endpoint (deprecated, use BaseURL)
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

// =============================================================================
// PUBLIC FACTORY FUNCTIONS
// =============================================================================

// NewMemory creates a new memory instance based on configuration
func NewMemory(config AgentMemoryConfig) (Memory, error) {
	// Set defaults
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
		if config.Embedding.Provider == "dummy" {
			config.KnowledgeScoreThreshold = 0.0
		} else {
			config.KnowledgeScoreThreshold = 0.7
		}
	}
	if config.ChunkSize == 0 {
		config.ChunkSize = 1000
	}
	if config.ChunkOverlap == 0 {
		config.ChunkOverlap = 200
	}

	// Try registry-based factory first (plugin providers)
	if config.Provider != "" {
		if factory, ok := getMemoryProviderFactory(config.Provider); ok {
			return factory(config)
		}
	}

	// Fallback to single internal factory to avoid circular imports (legacy path)
	if memoryFactory != nil {
		return memoryFactory(config)
	}

	Logger().Warn().Msg("No memory factory registered - using no-op memory")
	return &noOpMemory{}, nil
}

// QuickMemory creates an in-memory provider for quick testing
func QuickMemory() Memory {
	config := AgentMemoryConfig{
		Provider:   "memory",
		Connection: "memory",
		MaxResults: 10,
		Dimensions: 1536,
	}
	memory, err := NewMemory(config)
	if err != nil {
		return &noOpMemory{}
	}
	return memory
}

// RegisterMemoryFactory allows internal packages to register their factory function
func RegisterMemoryFactory(factory func(AgentMemoryConfig) (Memory, error)) {
	memoryFactory = factory
}

// =============================================================================
// MEMORY PROVIDER REGISTRY (Plugins register here)
// =============================================================================

// MemoryProviderFactory constructs a Memory implementation based on AgentMemoryConfig.
type MemoryProviderFactory func(AgentMemoryConfig) (Memory, error)

var (
	memoryProviderFactoriesMu sync.RWMutex
	memoryProviderFactories   = map[string]MemoryProviderFactory{}
)

// RegisterMemoryProviderFactory registers a named memory provider factory (e.g., "memory", "pgvector").
func RegisterMemoryProviderFactory(name string, factory MemoryProviderFactory) {
	memoryProviderFactoriesMu.Lock()
	defer memoryProviderFactoriesMu.Unlock()
	memoryProviderFactories[strings.ToLower(name)] = factory
}

func getMemoryProviderFactory(name string) (MemoryProviderFactory, bool) {
	memoryProviderFactoriesMu.RLock()
	defer memoryProviderFactoriesMu.RUnlock()
	f, ok := memoryProviderFactories[strings.ToLower(name)]
	return f, ok
}

// Embedding service factory functions
func NewOpenAIEmbeddingService(apiKey, model string) EmbeddingService {
	if openAIEmbeddingFactory != nil {
		return openAIEmbeddingFactory(apiKey, model)
	}
	Logger().Warn().Msg("No OpenAI embedding factory registered - using no-op service")
	return &noOpEmbeddingService{dimensions: 1536}
}

func NewOllamaEmbeddingService(model, baseURL string) EmbeddingService {
	if ollamaEmbeddingFactory != nil {
		return ollamaEmbeddingFactory(model, baseURL)
	}
	Logger().Warn().Msg("No Ollama embedding factory registered - using no-op service")
	return &noOpEmbeddingService{dimensions: 1024}
}

func NewDummyEmbeddingService(dimensions int) EmbeddingService {
	if dummyEmbeddingFactory != nil {
		return dummyEmbeddingFactory(dimensions)
	}
	if dimensions <= 0 {
		dimensions = 1536
	}
	return &noOpEmbeddingService{dimensions: dimensions}
}

// Register embedding service factories
func RegisterOpenAIEmbeddingFactory(factory func(string, string) EmbeddingService) {
	openAIEmbeddingFactory = factory
}

func RegisterOllamaEmbeddingFactory(factory func(string, string) EmbeddingService) {
	ollamaEmbeddingFactory = factory
}

func RegisterDummyEmbeddingFactory(factory func(int) EmbeddingService) {
	dummyEmbeddingFactory = factory
}

// =============================================================================
// OPTION CONSTRUCTORS
// =============================================================================

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

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func contains(text, query string) bool {
	if len(text) == 0 || len(query) == 0 {
		return false
	}
	text = strings.ToLower(text)
	query = strings.ToLower(query)
	return strings.Contains(text, query)
}

func containsAnyTag(tags []string, query string) bool {
	for _, tag := range tags {
		if contains(tag, query) {
			return true
		}
	}
	return false
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
	return len(text) / 4
}

func calculateScore(content, query string) float32 {
	if len(content) == 0 || len(query) == 0 {
		return 0
	}
	content = strings.ToLower(content)
	query = strings.ToLower(query)
	if strings.Contains(content, query) {
		if content == query {
			return 1.0
		}
		return 0.7
	}
	words := strings.Fields(query)
	matchCount := 0
	for _, word := range words {
		if strings.Contains(content, word) {
			matchCount++
		}
	}
	if matchCount > 0 {
		return float32(matchCount) / float32(len(words)) * 0.5
	}
	return 0
}

// =============================================================================
// INTERNAL IMPLEMENTATIONS
// =============================================================================

var (
	memoryFactory          func(AgentMemoryConfig) (Memory, error)
	openAIEmbeddingFactory func(string, string) EmbeddingService
	ollamaEmbeddingFactory func(string, string) EmbeddingService
	dummyEmbeddingFactory  func(int) EmbeddingService
)

// Temporary no-op implementations during refactoring
type noOpMemory struct{}

func (m *noOpMemory) Store(ctx context.Context, content string, tags ...string) error { return nil }
func (m *noOpMemory) Query(ctx context.Context, query string, limit ...int) ([]Result, error) {
	return []Result{}, nil
}
func (m *noOpMemory) Remember(ctx context.Context, key string, value any) error  { return nil }
func (m *noOpMemory) Recall(ctx context.Context, key string) (any, error)        { return nil, nil }
func (m *noOpMemory) AddMessage(ctx context.Context, role, content string) error { return nil }
func (m *noOpMemory) GetHistory(ctx context.Context, limit ...int) ([]Message, error) {
	return []Message{}, nil
}
func (m *noOpMemory) NewSession() string                                               { return "default" }
func (m *noOpMemory) SetSession(ctx context.Context, sessionID string) context.Context { return ctx }
func (m *noOpMemory) ClearSession(ctx context.Context) error                           { return nil }
func (m *noOpMemory) Close() error                                                     { return nil }
func (m *noOpMemory) IngestDocument(ctx context.Context, doc Document) error           { return nil }
func (m *noOpMemory) IngestDocuments(ctx context.Context, docs []Document) error       { return nil }
func (m *noOpMemory) SearchKnowledge(ctx context.Context, query string, options ...SearchOption) ([]KnowledgeResult, error) {
	return []KnowledgeResult{}, nil
}
func (m *noOpMemory) SearchAll(ctx context.Context, query string, options ...SearchOption) (*HybridResult, error) {
	return &HybridResult{}, nil
}
func (m *noOpMemory) BuildContext(ctx context.Context, query string, options ...ContextOption) (*RAGContext, error) {
	return &RAGContext{}, nil
}

type noOpEmbeddingService struct {
	dimensions int
}

func (s *noOpEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
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
