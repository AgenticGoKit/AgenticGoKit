package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
	"github.com/kunalkushwaha/agenticgokit/core/vnext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// MOCK MEMORY PROVIDER FOR TESTING
// =============================================================================

type mockMemoryProvider struct {
	memories     []core.Result
	messages     []core.Message
	queryErr     error
	storeCount   int
	messageCount int
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

func (m *mockMemoryProvider) Store(ctx context.Context, content string, tags ...string) error {
	m.storeCount++
	// Simulate storing in memory
	m.memories = append(m.memories, core.Result{
		Content:   content,
		Score:     0.9,
		Tags:      tags,
		CreatedAt: time.Now(),
	})
	return nil
}

func (m *mockMemoryProvider) AddMessage(ctx context.Context, role, content string) error {
	m.messageCount++
	m.messages = append(m.messages, core.Message{
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	})
	return nil
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
func (m *mockMemoryProvider) Remember(ctx context.Context, key string, value any) error { return nil }
func (m *mockMemoryProvider) Recall(ctx context.Context, key string) (any, error)       { return nil, nil }
func (m *mockMemoryProvider) NewSession() string                                        { return "test-session" }
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
// MEMORY INTEGRATION TESTS
// =============================================================================

func TestMemoryEnrichment(t *testing.T) {
	t.Run("EnrichWithMemory with RAG config", func(t *testing.T) {
		ctx := context.Background()

		// Create mock memory with some previous context
		mock := &mockMemoryProvider{
			memories: []core.Result{
				{Content: "User previously asked about Go programming", Score: 0.9},
				{Content: "User is interested in web development", Score: 0.8},
			},
		}

		config := &vnext.MemoryConfig{
			Provider: "memory",
			RAG: &vnext.RAGConfig{
				MaxTokens:       1000,
				PersonalWeight:  0.3,
				KnowledgeWeight: 0.7,
				HistoryLimit:    5,
			},
		}

		result := vnext.EnrichWithMemory(ctx, mock, "How do I build a web server?", config)

		// Verify enrichment happened
		assert.Contains(t, result, "Relevant Context", "Should include context header")
		assert.Contains(t, result, "Go programming", "Should include relevant memory")
		assert.Contains(t, result, "web development", "Should include relevant memory")
		assert.Contains(t, result, "How do I build a web server?", "Should include original query")
	})

	t.Run("EnrichWithMemory without RAG config", func(t *testing.T) {
		ctx := context.Background()

		mock := &mockMemoryProvider{
			memories: []core.Result{
				{Content: "Previous fact 1", Score: 0.9},
			},
		}

		config := &vnext.MemoryConfig{
			Provider: "memory",
			// No RAG config - should use simple formatting
		}

		result := vnext.EnrichWithMemory(ctx, mock, "test query", config)

		// Should use simple context formatting
		assert.Contains(t, result, "Relevant previous information")
		assert.Contains(t, result, "Previous fact 1")
		assert.Contains(t, result, "test query")
	})

	t.Run("EnrichWithMemory with no memories", func(t *testing.T) {
		ctx := context.Background()

		mock := &mockMemoryProvider{
			memories: []core.Result{}, // Empty
		}

		config := &vnext.MemoryConfig{
			Provider: "memory",
		}

		result := vnext.EnrichWithMemory(ctx, mock, "test query", config)

		// Should return original query unchanged
		assert.Equal(t, "test query", result)
	})
}

func TestBuildEnrichedPrompt(t *testing.T) {
	t.Run("with memory and chat history", func(t *testing.T) {
		ctx := context.Background()

		mock := &mockMemoryProvider{
			memories: []core.Result{
				{Content: "User is learning Go", Score: 0.9},
			},
			messages: []core.Message{
				{Role: "user", Content: "What is Go?"},
				{Role: "assistant", Content: "Go is a programming language."},
			},
		}

		config := &vnext.MemoryConfig{
			Provider: "memory",
			RAG: &vnext.RAGConfig{
				MaxTokens:    2000,
				HistoryLimit: 10, // Enable history
			},
		}

		result := vnext.BuildEnrichedPrompt(
			ctx,
			"You are a helpful assistant",
			"Tell me more about Go",
			mock,
			config,
		)

		// Check system prompt unchanged
		assert.Equal(t, "You are a helpful assistant", result.System)

		// Check user prompt includes both memory and history
		assert.Contains(t, result.User, "Previous Conversation", "Should include chat history")
		assert.Contains(t, result.User, "What is Go?", "Should include past user message")
		assert.Contains(t, result.User, "Relevant Context", "Should include memory context")
		assert.Contains(t, result.User, "learning Go", "Should include relevant memory")
		assert.Contains(t, result.User, "Tell me more about Go", "Should include current query")
	})

	t.Run("without memory provider", func(t *testing.T) {
		ctx := context.Background()

		result := vnext.BuildEnrichedPrompt(
			ctx,
			"You are a helper",
			"Hello",
			nil, // No memory provider
			nil,
		)

		// Should return basic prompt
		assert.Equal(t, "You are a helper", result.System)
		assert.Equal(t, "Hello", result.User)
	})
}

func TestRAGConfigValidation(t *testing.T) {
	t.Run("validates and applies defaults", func(t *testing.T) {
		config := &vnext.RAGConfig{
			// Empty - should get defaults
		}

		validated := vnext.ValidateRAGConfig(config)

		require.NotNil(t, validated)
		assert.Equal(t, 2000, validated.MaxTokens, "Should apply default max tokens")
		assert.InDelta(t, 0.3, validated.PersonalWeight, 0.01, "Should apply default personal weight")
		assert.InDelta(t, 0.7, validated.KnowledgeWeight, 0.01, "Should apply default knowledge weight")
		assert.Equal(t, 10, validated.HistoryLimit, "Should apply default history limit")
	})

	t.Run("normalizes weights", func(t *testing.T) {
		config := &vnext.RAGConfig{
			PersonalWeight:  0.4,
			KnowledgeWeight: 0.8, // Total is 1.2, should be normalized
		}

		validated := vnext.ValidateRAGConfig(config)

		require.NotNil(t, validated)
		// Weights should be normalized to sum to 1.0
		assert.InDelta(t, 0.333, validated.PersonalWeight, 0.01)
		assert.InDelta(t, 0.667, validated.KnowledgeWeight, 0.01)
	})
}

func TestMemoryStorage(t *testing.T) {
	t.Run("stores both memory and chat messages", func(t *testing.T) {
		ctx := context.Background()

		mock := &mockMemoryProvider{}

		// Simulate storing an interaction
		input := "What is the weather?"
		output := "I don't have access to weather data."

		// Store in memory (simulating what agent does)
		err := mock.Store(ctx, input, "user_message", "conversation")
		require.NoError(t, err)

		err = mock.Store(ctx, output, "agent_response", "conversation")
		require.NoError(t, err)

		err = mock.AddMessage(ctx, "user", input)
		require.NoError(t, err)

		err = mock.AddMessage(ctx, "assistant", output)
		require.NoError(t, err)

		// Verify storage counts
		assert.Equal(t, 2, mock.storeCount, "Should have stored 2 memories")
		assert.Equal(t, 2, mock.messageCount, "Should have stored 2 chat messages")

		// Verify we can retrieve chat history
		messages, err := mock.GetHistory(ctx)
		require.NoError(t, err)
		assert.Len(t, messages, 2)
		assert.Equal(t, "user", messages[0].Role)
		assert.Equal(t, input, messages[0].Content)
		assert.Equal(t, "assistant", messages[1].Role)
		assert.Equal(t, output, messages[1].Content)
	})
}
