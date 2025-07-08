package core

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Memory is the central memory interface - replaces multiple interfaces
// Breaking change: Single, unified interface for all memory operations
type Memory interface {
	// Store automatically handles embedding generation
	Store(ctx context.Context, content string, tags ...string) error

	// Query performs semantic search with automatic embedding
	Query(ctx context.Context, query string, limit ...int) ([]Result, error)

	// Remember stores key-value data (non-semantic)
	Remember(ctx context.Context, key string, value any) error

	// Recall retrieves key-value data
	Recall(ctx context.Context, key string) (any, error)

	// Chat methods - built-in conversation management
	AddMessage(ctx context.Context, role, content string) error
	GetHistory(ctx context.Context, limit ...int) ([]Message, error)

	// Session management
	NewSession() string
	SetSession(ctx context.Context, sessionID string) context.Context
	ClearSession(ctx context.Context) error

	// Lifecycle
	Close() error
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

// AgentMemoryConfig - simplified configuration for agent memory storage
type AgentMemoryConfig struct {
	Provider   string `toml:"provider"`    // pgvector, weaviate, memory
	Connection string `toml:"connection"`  // postgres://..., http://..., or "memory"
	MaxResults int    `toml:"max_results"` // default: 10
	Dimensions int    `toml:"dimensions"`  // default: 1536
	AutoEmbed  bool   `toml:"auto_embed"`  // default: true
}

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
}

type vectorEntry struct {
	Content   string
	Tags      []string
	CreatedAt time.Time
	// For in-memory, we'll use simple string matching instead of real embeddings
}

func newInMemoryProvider(config AgentMemoryConfig) (Memory, error) {
	return &InMemoryProvider{
		vectors:   make(map[string]vectorEntry),
		keyValues: make(map[string]any),
		messages:  make(map[string][]Message),
		sessionID: "default",
		config:    config,
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
			// Basic substring matching (in production, use proper embeddings)
			if query == "" || contains(entry.Content, query) || containsAnyTag(entry.Tags, query) {
				results = append(results, Result{
					Content:   entry.Content,
					Score:     0.95, // Dummy score for in-memory
					Tags:      entry.Tags,
					CreatedAt: entry.CreatedAt,
				})

				if len(results) >= maxResults {
					break
				}
			}
		}
	}

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

	return nil
}

// PgVector provider - production-ready PostgreSQL with pgvector
func newPgVectorProvider(config AgentMemoryConfig) (Memory, error) {
	// TODO: Implement PgVector provider
	// For now, return an error indicating it needs implementation
	return nil, fmt.Errorf("PgVector provider not yet implemented - use 'memory' provider for now")
}

// Weaviate provider - production-ready vector database
func newWeaviateProvider(config AgentMemoryConfig) (Memory, error) {
	// TODO: Implement Weaviate provider
	// For now, return an error indicating it needs implementation
	return nil, fmt.Errorf("Weaviate provider not yet implemented - use 'memory' provider for now")
}

// Utility functions
func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func contains(text, query string) bool {
	return len(text) > 0 && len(query) > 0 &&
		(text == query ||
			len(text) > len(query) &&
				(text[:len(query)] == query ||
					text[len(text)-len(query):] == query ||
					findSubstring(text, query)))
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
