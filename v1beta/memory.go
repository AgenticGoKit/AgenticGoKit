package v1beta

import (
	"context"
	"fmt"
	"time"
)

// =============================================================================
// UNIFIED MEMORY INTERFACE AND TYPES
// =============================================================================

// NOTE: Core memory types (Memory, MemoryResult, Document, RAGContext,
// StoreOption, QueryOption, ContextOption, StoreConfig, QueryConfig, ContextConfig)
// are defined in agent.go to avoid duplication.
//
// This file provides the implementation and utility functions for memory operations.

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

// =============================================================================
// STORE OPTIONS
// =============================================================================

// WithContentType sets the content type for stored content
func WithContentType(contentType string) StoreOption {
	return func(config *StoreConfig) {
		config.ContentType = contentType
	}
}

// WithSource sets the source for stored content
func WithSource(source string) StoreOption {
	return func(config *StoreConfig) {
		config.Source = source
	}
}

// WithMetadata adds metadata to stored content
func WithMetadata(metadata map[string]interface{}) StoreOption {
	return func(config *StoreConfig) {
		if config.Metadata == nil {
			config.Metadata = make(map[string]interface{})
		}
		for k, v := range metadata {
			config.Metadata[k] = v
		}
	}
}

// =============================================================================
// QUERY OPTIONS
// =============================================================================

// WithLimit sets the maximum number of results to return
func WithLimit(limit int) QueryOption {
	return func(config *QueryConfig) {
		config.Limit = limit
	}
}

// WithScoreThreshold sets the minimum score threshold for results
func WithScoreThreshold(threshold float32) QueryOption {
	return func(config *QueryConfig) {
		config.ScoreThreshold = threshold
	}
}

// WithIncludeMetadata sets whether to include metadata in results
func WithIncludeMetadata(include bool) QueryOption {
	return func(config *QueryConfig) {
		config.IncludeMetadata = include
	}
}

// =============================================================================
// CONTEXT OPTIONS
// =============================================================================

// WithMaxTokens sets the maximum token count for RAG context
func WithMaxTokens(maxTokens int) ContextOption {
	return func(config *ContextConfig) {
		config.MaxTokens = maxTokens
	}
}

// WithPersonalWeight sets the weight for personal memory in RAG context
func WithPersonalWeight(weight float32) ContextOption {
	return func(config *ContextConfig) {
		config.PersonalWeight = weight
	}
}

// WithKnowledgeWeight sets the weight for knowledge base in RAG context
func WithKnowledgeWeight(weight float32) ContextOption {
	return func(config *ContextConfig) {
		config.KnowledgeWeight = weight
	}
}

// Note: ContextConfig in agent.go only has MaxTokens, PersonalWeight, and KnowledgeWeight
// Additional context options can be added as needed

// =============================================================================
// MEMORY FACTORY FUNCTIONS
// =============================================================================

// NewMemory creates a new memory instance based on configuration
func NewMemory(config *MemoryConfig) (Memory, error) {
	// Set defaults
	if config == nil {
		config = &MemoryConfig{
			Provider: "memory",
		}
	}

	// Apply defaults from configuration
	applyMemoryDefaults(config)

	// Try to create memory using registered factory
	if factory := getMemoryFactory(config.Provider); factory != nil {
		return factory(config)
	}

	// Fallback to no-op implementation
	return &noOpMemory{}, nil
}

// QuickMemory creates an in-memory provider for quick testing
func QuickMemory() Memory {
	config := &MemoryConfig{
		Provider:   "memory",
		Connection: "memory",
	}
	memory, err := NewMemory(config)
	if err != nil {
		return &noOpMemory{}
	}
	return memory
}

// =============================================================================
// MEMORY PROVIDER REGISTRY
// =============================================================================

// MemoryFactory creates a Memory implementation based on MemoryConfig
type MemoryFactory func(*MemoryConfig) (Memory, error)

var memoryFactories = make(map[string]MemoryFactory)

// RegisterMemoryProvider registers a memory provider factory
func RegisterMemoryProvider(name string, factory MemoryFactory) {
	memoryFactories[name] = factory
}

// getMemoryFactory retrieves a memory factory by name
func getMemoryFactory(name string) MemoryFactory {
	return memoryFactories[name]
}

// =============================================================================
// CONFIGURATION HELPERS
// =============================================================================

// applyMemoryDefaults applies default values to memory configuration
func applyMemoryDefaults(config *MemoryConfig) {
	if config.Provider == "" {
		config.Provider = "memory"
	}
	if config.Connection == "" && config.Provider == "memory" {
		config.Connection = "memory"
	}

	// Apply RAG defaults if RAG is configured
	if config.RAG != nil {
		if config.RAG.MaxTokens == 0 {
			config.RAG.MaxTokens = 4000
		}
		if config.RAG.PersonalWeight == 0 {
			config.RAG.PersonalWeight = 0.3
		}
		if config.RAG.KnowledgeWeight == 0 {
			config.RAG.KnowledgeWeight = 0.7
		}
		if config.RAG.HistoryLimit == 0 {
			config.RAG.HistoryLimit = 10
		}
	}
}

// =============================================================================
// NO-OP IMPLEMENTATION
// =============================================================================

// noOpMemory provides a no-op implementation of the Memory interface
type noOpMemory struct{}

func (m *noOpMemory) Store(ctx context.Context, content string, opts ...StoreOption) error {
	return nil
}

func (m *noOpMemory) Query(ctx context.Context, query string, opts ...QueryOption) ([]MemoryResult, error) {
	return []MemoryResult{}, nil
}

func (m *noOpMemory) IngestDocument(ctx context.Context, doc Document) error {
	return nil
}

func (m *noOpMemory) BuildContext(ctx context.Context, query string, opts ...ContextOption) (*RAGContext, error) {
	return &RAGContext{
		PersonalMemory:    []MemoryResult{},
		KnowledgeBase:     []MemoryResult{},
		ChatHistory:       []string{},
		TotalTokens:       0,
		SourceAttribution: []string{},
	}, nil
}

func (m *noOpMemory) NewSession() string {
	return "default"
}

func (m *noOpMemory) SetSession(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, "session_id", sessionID)
}

func (m *noOpMemory) Close() error {
	return nil
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// EstimateTokenCount estimates the token count for a given text
func EstimateTokenCount(text string) int {
	// Simple estimation: ~4 characters per token
	return len(text) / 4
}

// FormatRAGContext formats a RAG context using a template or default format
func FormatRAGContext(context *RAGContext, template string) string {
	if template != "" {
		// Custom template formatting could be implemented here
		// For now, use default formatting
	}

	// Default formatting - combine personal memory and knowledge base
	formatted := ""

	if len(context.PersonalMemory) > 0 {
		formatted += "Personal Context:\n"
		for i, result := range context.PersonalMemory {
			formatted += fmt.Sprintf("%d. %s", i+1, result.Content)
			if result.Source != "" {
				formatted += fmt.Sprintf(" (Source: %s)", result.Source)
			}
			formatted += "\n"
		}
	}

	if len(context.KnowledgeBase) > 0 {
		if formatted != "" {
			formatted += "\n"
		}
		formatted += "Knowledge Base:\n"
		for i, result := range context.KnowledgeBase {
			formatted += fmt.Sprintf("%d. %s", i+1, result.Content)
			if result.Source != "" {
				formatted += fmt.Sprintf(" (Source: %s)", result.Source)
			}
			formatted += "\n"
		}
	}

	return formatted
}

// =============================================================================
// CONVENIENCE FUNCTIONS
// =============================================================================

// StoreSimple stores content with metadata
func StoreSimple(ctx context.Context, memory Memory, content string, metadata map[string]interface{}) error {
	return memory.Store(ctx, content, WithMetadata(metadata))
}

// QuerySimple performs a simple query with a limit
func QuerySimple(ctx context.Context, memory Memory, query string, limit int) ([]MemoryResult, error) {
	return memory.Query(ctx, query, WithLimit(limit))
}

// BuildSimpleContext builds a RAG context with default settings
func BuildSimpleContext(ctx context.Context, memory Memory, query string, maxTokens int) (*RAGContext, error) {
	return memory.BuildContext(ctx, query,
		WithMaxTokens(maxTokens),
		WithPersonalWeight(0.3),
		WithKnowledgeWeight(0.7),
	)
}

// IngestTextDocument ingests a simple text document
func IngestTextDocument(ctx context.Context, memory Memory, title, content, source string) error {
	doc := Document{
		ID:      fmt.Sprintf("doc_%d", time.Now().UnixNano()),
		Title:   title,
		Content: content,
		Source:  source,
		Metadata: map[string]interface{}{
			"type":       string(DocumentTypeText),
			"created_at": time.Now(),
		},
	}
	return memory.IngestDocument(ctx, doc)
}

// =============================================================================
// MEMORY BUILDER PATTERN
// =============================================================================

// MemoryBuilder provides a fluent interface for creating memory instances
type MemoryBuilder struct {
	config *MemoryConfig
}

// NewMemoryBuilder creates a new memory builder
func NewMemoryBuilder() *MemoryBuilder {
	return &MemoryBuilder{
		config: &MemoryConfig{},
	}
}

// WithProvider sets the memory provider
func (b *MemoryBuilder) WithProvider(provider string) *MemoryBuilder {
	b.config.Provider = provider
	return b
}

// WithConnection sets the connection string
func (b *MemoryBuilder) WithConnection(connection string) *MemoryBuilder {
	b.config.Connection = connection
	return b
}

// WithRAGConfig sets the RAG configuration
func (b *MemoryBuilder) WithRAGConfig(rag *RAGConfig) *MemoryBuilder {
	b.config.RAG = rag
	return b
}

// WithOptions sets additional options
func (b *MemoryBuilder) WithOptions(options map[string]string) *MemoryBuilder {
	b.config.Options = options
	return b
}

// Build creates the memory instance
func (b *MemoryBuilder) Build() (Memory, error) {
	return NewMemory(b.config)
}

// =============================================================================
// MEMORY SESSION CONTEXT
// =============================================================================

// SessionContext provides session-aware memory operations
type SessionContext struct {
	memory    Memory
	sessionID string
}

// NewSessionContext creates a new session context
func NewSessionContext(memory Memory, sessionID string) *SessionContext {
	return &SessionContext{
		memory:    memory,
		sessionID: sessionID,
	}
}

// Store stores content in the current session
func (sc *SessionContext) Store(ctx context.Context, content string, opts ...StoreOption) error {
	// Set session context and store
	sessionCtx := sc.memory.SetSession(ctx, sc.sessionID)
	return sc.memory.Store(sessionCtx, content, opts...)
}

// Query queries content from the current session
func (sc *SessionContext) Query(ctx context.Context, query string, opts ...QueryOption) ([]MemoryResult, error) {
	// Set session context and query
	sessionCtx := sc.memory.SetSession(ctx, sc.sessionID)
	return sc.memory.Query(sessionCtx, query, opts...)
}

// BuildContext builds RAG context for the current session
func (sc *SessionContext) BuildContext(ctx context.Context, query string, opts ...ContextOption) (*RAGContext, error) {
	sessionCtx := sc.memory.SetSession(ctx, sc.sessionID)
	return sc.memory.BuildContext(sessionCtx, query, opts...)
}

// =============================================================================
// MEMORY STATISTICS AND MONITORING
// =============================================================================

// MemoryStats provides statistics about memory usage
type MemoryStats struct {
	TotalDocuments int           `json:"total_documents"`
	TotalMemories  int           `json:"total_memories"`
	AverageScore   float32       `json:"average_score"`
	LastQueryTime  time.Duration `json:"last_query_time"`
	SessionCount   int           `json:"session_count"`
	StorageSize    int64         `json:"storage_size_bytes"`
}

// StatsProvider interface for memory implementations that provide statistics
type StatsProvider interface {
	GetStats(ctx context.Context) (*MemoryStats, error)
}

// GetMemoryStats retrieves statistics from a memory instance if supported
func GetMemoryStats(ctx context.Context, memory Memory) (*MemoryStats, error) {
	if statsProvider, ok := memory.(StatsProvider); ok {
		return statsProvider.GetStats(ctx)
	}
	return &MemoryStats{}, fmt.Errorf("memory provider does not support statistics")
}

