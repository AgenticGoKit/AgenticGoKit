package utils_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
	vnext "github.com/agenticgokit/agenticgokit/v1beta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// MOCK MEMORY PROVIDER FOR TESTING
// =============================================================================

type mockMemoryProvider struct {
	memories []core.Result
	messages []core.Message
	queryErr error
}

func (m *mockMemoryProvider) Query(ctx context.Context, query string, limit ...int) ([]core.Result, error) {
	if m.queryErr != nil {
		return nil, m.queryErr
	}

	maxResults := len(m.memories)
	if len(limit) > 0 && limit[0] < maxResults {
		maxResults = limit[0]
	}

	if maxResults > len(m.memories) {
		maxResults = len(m.memories)
	}

	return m.memories[:maxResults], nil
}

func (m *mockMemoryProvider) GetHistory(ctx context.Context, limit ...int) ([]core.Message, error) {
	maxResults := len(m.messages)
	if len(limit) > 0 && limit[0] < maxResults {
		maxResults = limit[0]
	}

	if maxResults > len(m.messages) {
		maxResults = len(m.messages)
	}

	return m.messages[:maxResults], nil
}

// Stub implementations for other Memory interface methods
func (m *mockMemoryProvider) Store(ctx context.Context, content string, tags ...string) error {
	return nil
}
func (m *mockMemoryProvider) Remember(ctx context.Context, key string, value any) error { return nil }
func (m *mockMemoryProvider) Recall(ctx context.Context, key string) (any, error)       { return nil, nil }
func (m *mockMemoryProvider) AddMessage(ctx context.Context, role, content string) error {
	return nil
}
func (m *mockMemoryProvider) NewSession() string { return "test-session" }
func (m *mockMemoryProvider) SetSession(ctx context.Context, sessionID string) context.Context {
	return ctx
}
func (m *mockMemoryProvider) ClearSession(ctx context.Context) error { return nil }
func (m *mockMemoryProvider) Close() error                           { return nil }
func (m *mockMemoryProvider) IngestDocument(ctx context.Context, doc core.Document) error {
	return nil
}
func (m *mockMemoryProvider) IngestDocuments(ctx context.Context, docs []core.Document) error {
	return nil
}
func (m *mockMemoryProvider) SearchKnowledge(ctx context.Context, query string, options ...core.SearchOption) ([]core.KnowledgeResult, error) {
	return nil, nil
}
func (m *mockMemoryProvider) SearchAll(ctx context.Context, query string, options ...core.SearchOption) (*core.HybridResult, error) {
	return nil, nil
}
func (m *mockMemoryProvider) BuildContext(ctx context.Context, query string, options ...core.ContextOption) (*core.RAGContext, error) {
	return nil, nil
}

// =============================================================================
// UTILITY FUNCTION TESTS
// =============================================================================

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "empty string",
			text:     "",
			expected: 0,
		},
		{
			name:     "short text",
			text:     "Hello",
			expected: 2, // (5+3)/4 = 2
		},
		{
			name:     "medium text",
			text:     "The quick brown fox jumps over the lazy dog",
			expected: 11, // (44+3)/4 = 11
		},
		{
			name:     "long text",
			text:     "This is a longer piece of text that should result in more tokens being estimated for the calculation",
			expected: 25, // (100+3)/4 = 25 (actual length is 100 chars)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vnext.EstimateTokens(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateToTokenLimit(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		maxTokens int
		wantLen   bool // Check if truncated
	}{
		{
			name:      "within limit",
			text:      "Short text",
			maxTokens: 100,
			wantLen:   false,
		},
		{
			name:      "exceeds limit",
			text:      "This is a very long piece of text that should definitely exceed the token limit and be truncated",
			maxTokens: 5,
			wantLen:   true,
		},
		{
			name:      "zero limit",
			text:      "Some text",
			maxTokens: 0,
			wantLen:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vnext.TruncateToTokenLimit(tt.text, tt.maxTokens)

			if tt.wantLen {
				assert.Less(t, len(result), len(tt.text), "Text should be truncated")
				assert.Contains(t, result, "...", "Truncated text should contain ellipsis")
			} else {
				assert.Equal(t, tt.text, result, "Text should not be truncated")
			}
		})
	}
}

func TestExtractSources(t *testing.T) {
	tests := []struct {
		name     string
		memories []core.Result
		expected []string
	}{
		{
			name:     "no memories",
			memories: []core.Result{},
			expected: []string{},
		},
		{
			name: "memories with source tags",
			memories: []core.Result{
				{Content: "Test 1", Tags: []string{"source:https://example.com", "topic:test"}},
				{Content: "Test 2", Tags: []string{"source:https://example.org"}},
			},
			expected: []string{"https://example.com", "https://example.org"},
		},
		{
			name: "memories without source tags",
			memories: []core.Result{
				{Content: "Test 1", Tags: []string{"topic:test", "category:example"}},
				{Content: "Test 2", Tags: []string{"author:someone"}},
			},
			expected: []string{},
		},
		{
			name: "duplicate sources",
			memories: []core.Result{
				{Content: "Test 1", Tags: []string{"source:https://example.com"}},
				{Content: "Test 2", Tags: []string{"source:https://example.com"}},
			},
			expected: []string{"https://example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vnext.ExtractSources(tt.memories)

			if len(tt.expected) == 0 {
				assert.Empty(t, result)
			} else {
				assert.ElementsMatch(t, tt.expected, result)
			}
		})
	}
}

func TestFormatMetadataForPrompt(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]interface{}
		wantKey  string
	}{
		{
			name:     "empty metadata",
			metadata: map[string]interface{}{},
			wantKey:  "",
		},
		{
			name: "with metadata",
			metadata: map[string]interface{}{
				"source": "test.txt",
				"author": "Test Author",
			},
			wantKey: "source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vnext.FormatMetadataForPrompt(tt.metadata)

			if tt.wantKey == "" {
				assert.Empty(t, result)
			} else {
				assert.Contains(t, result, "Metadata:")
				assert.Contains(t, result, tt.wantKey)
			}
		})
	}
}

// =============================================================================
// RAG CONFIG VALIDATION TESTS
// =============================================================================

func TestValidateRAGConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   *vnext.RAGConfig
		expected *vnext.RAGConfig
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: nil,
		},
		{
			name:   "apply defaults",
			config: &vnext.RAGConfig{},
			expected: &vnext.RAGConfig{
				MaxTokens:       2000,
				PersonalWeight:  0.3,
				KnowledgeWeight: 0.7,
				HistoryLimit:    10,
			},
		},
		{
			name: "normalize weights",
			config: &vnext.RAGConfig{
				PersonalWeight:  0.4,
				KnowledgeWeight: 0.8,
			},
			expected: &vnext.RAGConfig{
				MaxTokens:       2000,
				PersonalWeight:  0.333, // Normalized (0.4/1.2)
				KnowledgeWeight: 0.666, // Normalized (0.8/1.2)
				HistoryLimit:    10,
			},
		},
		{
			name: "valid config unchanged",
			config: &vnext.RAGConfig{
				MaxTokens:       3000,
				PersonalWeight:  0.5,
				KnowledgeWeight: 0.5,
				HistoryLimit:    20,
			},
			expected: &vnext.RAGConfig{
				MaxTokens:       3000,
				PersonalWeight:  0.5,
				KnowledgeWeight: 0.5,
				HistoryLimit:    20,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vnext.ValidateRAGConfig(tt.config)

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.MaxTokens, result.MaxTokens)
				assert.InDelta(t, tt.expected.PersonalWeight, result.PersonalWeight, 0.01)
				assert.InDelta(t, tt.expected.KnowledgeWeight, result.KnowledgeWeight, 0.01)
				assert.Equal(t, tt.expected.HistoryLimit, result.HistoryLimit)
			}
		})
	}
}

// =============================================================================
// MEMORY ENRICHMENT TESTS
// =============================================================================

func TestEnrichWithMemory(t *testing.T) {
	ctx := context.Background()

	t.Run("nil memory provider", func(t *testing.T) {
		result := vnext.EnrichWithMemory(ctx, nil, "test query", &vnext.MemoryConfig{})
		assert.Equal(t, "test query", result)
	})

	t.Run("nil config", func(t *testing.T) {
		mock := &mockMemoryProvider{}
		result := vnext.EnrichWithMemory(ctx, mock, "test query", nil)
		assert.Equal(t, "test query", result)
	})

	t.Run("no memories found", func(t *testing.T) {
		mock := &mockMemoryProvider{
			memories: []core.Result{},
		}
		config := &vnext.MemoryConfig{}

		result := vnext.EnrichWithMemory(ctx, mock, "test query", config)
		assert.Equal(t, "test query", result)
	})

	t.Run("with memories and RAG config", func(t *testing.T) {
		mock := &mockMemoryProvider{
			memories: []core.Result{
				{Content: "Previous fact 1", Score: 0.9, CreatedAt: time.Now()},
				{Content: "Previous fact 2", Score: 0.7, CreatedAt: time.Now()},
			},
		}
		config := &vnext.MemoryConfig{
			RAG: &vnext.RAGConfig{
				MaxTokens:      1000,
				HistoryLimit:   5,
				PersonalWeight: 0.5,
			},
		}

		result := vnext.EnrichWithMemory(ctx, mock, "test query", config)

		assert.Contains(t, result, "Relevant Context")
		assert.Contains(t, result, "Previous fact 1")
		assert.Contains(t, result, "test query")
	})

	t.Run("without RAG config (simple context)", func(t *testing.T) {
		mock := &mockMemoryProvider{
			memories: []core.Result{
				{Content: "Previous fact 1", Score: 0.9},
			},
		}
		config := &vnext.MemoryConfig{} // No RAG config

		result := vnext.EnrichWithMemory(ctx, mock, "test query", config)

		assert.Contains(t, result, "Relevant previous information")
		assert.Contains(t, result, "Previous fact 1")
		assert.Contains(t, result, "test query")
	})
}

func TestBuildRAGContext(t *testing.T) {
	t.Run("empty memories", func(t *testing.T) {
		result := vnext.BuildRAGContext([]core.Result{}, &vnext.RAGConfig{}, "test query")
		assert.Equal(t, "test query", result)
	})

	t.Run("with memories", func(t *testing.T) {
		memories := []core.Result{
			{Content: "Memory 1", Score: 0.9, Tags: []string{"tag1", "tag2"}},
			{Content: "Memory 2", Score: 0.7},
		}
		config := &vnext.RAGConfig{
			MaxTokens: 1000,
		}

		result := vnext.BuildRAGContext(memories, config, "test query")

		assert.Contains(t, result, "# Relevant Context")
		assert.Contains(t, result, "Memory 1")
		assert.Contains(t, result, "Memory 2")
		assert.Contains(t, result, "Relevance: 0.90")
		assert.Contains(t, result, "Relevance: 0.70")
		assert.Contains(t, result, "Tags: tag1, tag2")
		assert.Contains(t, result, "# User Query")
		assert.Contains(t, result, "test query")
	})

	t.Run("token limit enforcement", func(t *testing.T) {
		// Create a memory with lots of content
		longContent := strings.Repeat("This is a long memory. ", 100)
		memories := []core.Result{
			{Content: longContent, Score: 0.9},
			{Content: "This should be excluded", Score: 0.8},
		}
		config := &vnext.RAGConfig{
			MaxTokens: 50, // Very small limit
		}

		result := vnext.BuildRAGContext(memories, config, "test query")

		// Should include first memory but not second
		assert.Contains(t, result, "# Relevant Context")
		assert.NotContains(t, result, "should be excluded")
	})
}

func TestBuildMemorySimpleContext(t *testing.T) {
	t.Run("empty memories", func(t *testing.T) {
		result := vnext.BuildMemorySimpleContext([]core.Result{}, "test query")
		assert.Equal(t, "test query", result)
	})

	t.Run("with memories", func(t *testing.T) {
		memories := []core.Result{
			{Content: "Memory 1"},
			{Content: "Memory 2"},
		}

		result := vnext.BuildMemorySimpleContext(memories, "test query")

		assert.Contains(t, result, "Relevant previous information")
		assert.Contains(t, result, "Memory 1")
		assert.Contains(t, result, "Memory 2")
		assert.Contains(t, result, "Current query: test query")
	})

	t.Run("limits to 3 memories", func(t *testing.T) {
		memories := []core.Result{
			{Content: "Memory 1"},
			{Content: "Memory 2"},
			{Content: "Memory 3"},
			{Content: "Memory 4"},
			{Content: "Memory 5"},
		}

		result := vnext.BuildMemorySimpleContext(memories, "test query")

		assert.Contains(t, result, "Memory 1")
		assert.Contains(t, result, "Memory 2")
		assert.Contains(t, result, "Memory 3")
		assert.NotContains(t, result, "Memory 4")
		assert.NotContains(t, result, "Memory 5")
	})
}

// =============================================================================
// CHAT HISTORY TESTS
// =============================================================================

func TestBuildChatHistoryContext(t *testing.T) {
	ctx := context.Background()

	t.Run("nil memory provider", func(t *testing.T) {
		result := vnext.BuildChatHistoryContext(ctx, nil, 10)
		assert.Empty(t, result)
	})

	t.Run("no messages", func(t *testing.T) {
		mock := &mockMemoryProvider{
			messages: []core.Message{},
		}

		result := vnext.BuildChatHistoryContext(ctx, mock, 10)
		assert.Empty(t, result)
	})

	t.Run("with messages", func(t *testing.T) {
		mock := &mockMemoryProvider{
			messages: []core.Message{
				{Role: "user", Content: "Hello", CreatedAt: time.Now()},
				{Role: "assistant", Content: "Hi there!", CreatedAt: time.Now()},
				{Role: "user", Content: "How are you?", CreatedAt: time.Now()},
			},
		}

		result := vnext.BuildChatHistoryContext(ctx, mock, 10)

		assert.Contains(t, result, "# Previous Conversation")
		assert.Contains(t, result, "**User**: Hello")
		assert.Contains(t, result, "**Assistant**: Hi there!")
		assert.Contains(t, result, "**User**: How are you?")
	})

	t.Run("respects history limit", func(t *testing.T) {
		mock := &mockMemoryProvider{
			messages: []core.Message{
				{Role: "user", Content: "Message 1"},
				{Role: "assistant", Content: "Message 2"},
				{Role: "user", Content: "Message 3"},
			},
		}

		result := vnext.BuildChatHistoryContext(ctx, mock, 2)

		// Should only include first 2 messages
		assert.Contains(t, result, "Message 1")
		assert.Contains(t, result, "Message 2")
		assert.NotContains(t, result, "Message 3")
	})
}

// =============================================================================
// PROMPT BUILDING TESTS
// =============================================================================

func TestBuildEnrichedPrompt(t *testing.T) {
	ctx := context.Background()

	t.Run("basic prompt without memory", func(t *testing.T) {
		result := vnext.BuildEnrichedPrompt(ctx, "You are a helpful assistant", "Hello", nil, nil)

		assert.Equal(t, "You are a helpful assistant", result.System)
		assert.Equal(t, "Hello", result.User)
	})

	t.Run("enriched prompt with memory", func(t *testing.T) {
		mock := &mockMemoryProvider{
			memories: []core.Result{
				{Content: "Previous context", Score: 0.9},
			},
		}
		config := &vnext.MemoryConfig{
			RAG: &vnext.RAGConfig{
				MaxTokens:    1000,
				HistoryLimit: 5,
			},
		}

		result := vnext.BuildEnrichedPrompt(ctx, "You are a helpful assistant", "Hello", mock, config)

		assert.Equal(t, "You are a helpful assistant", result.System)
		assert.Contains(t, result.User, "Relevant Context")
		assert.Contains(t, result.User, "Previous context")
		assert.Contains(t, result.User, "Hello")
	})

	t.Run("with chat history", func(t *testing.T) {
		mock := &mockMemoryProvider{
			memories: []core.Result{
				{Content: "Previous context", Score: 0.9},
			},
			messages: []core.Message{
				{Role: "user", Content: "Previous message"},
			},
		}
		config := &vnext.MemoryConfig{
			RAG: &vnext.RAGConfig{
				MaxTokens:    1000,
				HistoryLimit: 10, // Enable chat history
			},
		}

		result := vnext.BuildEnrichedPrompt(ctx, "You are a helpful assistant", "Hello", mock, config)

		assert.Contains(t, result.User, "Previous Conversation")
		assert.Contains(t, result.User, "Previous message")
	})
}



