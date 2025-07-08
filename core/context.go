package core

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

// Context keys for memory and session
type memoryContextKey struct{}
type sessionContextKey struct{}

// WithMemory creates a context with memory and session ID
// Breaking change: Memory is always available through context
func WithMemory(ctx context.Context, memory Memory, sessionID string) context.Context {
	ctx = context.WithValue(ctx, memoryContextKey{}, memory)
	ctx = context.WithValue(ctx, sessionContextKey{}, sessionID)
	return ctx
}

// GetMemory retrieves memory from context
// Breaking change: Never returns nil - returns NoOpMemory instead
func GetMemory(ctx context.Context) Memory {
	if memory, ok := ctx.Value(memoryContextKey{}).(Memory); ok {
		return memory
	}
	// Return no-op memory instead of nil - prevents panics
	return &NoOpMemory{}
}

// GetSessionID retrieves session ID from context
func GetSessionID(ctx context.Context) string {
	if sessionID, ok := ctx.Value(sessionContextKey{}).(string); ok {
		return sessionID
	}
	return "default"
}

// GenerateSessionID creates a new unique session ID
func GenerateSessionID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Convenience functions - Breaking change: Much simpler API
// These functions provide direct access to memory operations through context

// StoreMemory stores content with semantic search capability
func StoreMemory(ctx context.Context, content string, tags ...string) error {
	return GetMemory(ctx).Store(ctx, content, tags...)
}

// QueryMemory performs semantic search
func QueryMemory(ctx context.Context, query string, limit ...int) ([]Result, error) {
	return GetMemory(ctx).Query(ctx, query, limit...)
}

// RememberValue stores key-value data (non-semantic)
func RememberValue(ctx context.Context, key string, value any) error {
	return GetMemory(ctx).Remember(ctx, key, value)
}

// RecallValue retrieves key-value data
func RecallValue(ctx context.Context, key string) (any, error) {
	return GetMemory(ctx).Recall(ctx, key)
}

// AddChatMessage adds a message to conversation history
func AddChatMessage(ctx context.Context, role, content string) error {
	return GetMemory(ctx).AddMessage(ctx, role, content)
}

// GetChatHistory retrieves conversation history
func GetChatHistory(ctx context.Context, limit ...int) ([]Message, error) {
	return GetMemory(ctx).GetHistory(ctx, limit...)
}

// NoOpMemory - prevents nil pointer panics when memory is not available
// Breaking change: Always return working memory interface
type NoOpMemory struct{}

func (n *NoOpMemory) Store(ctx context.Context, content string, tags ...string) error {
	return nil // Silent no-op
}

func (n *NoOpMemory) Query(ctx context.Context, query string, limit ...int) ([]Result, error) {
	return []Result{}, nil // Return empty results
}

func (n *NoOpMemory) Remember(ctx context.Context, key string, value any) error {
	return nil
}

func (n *NoOpMemory) Recall(ctx context.Context, key string) (any, error) {
	return nil, nil
}

func (n *NoOpMemory) AddMessage(ctx context.Context, role, content string) error {
	return nil
}

func (n *NoOpMemory) GetHistory(ctx context.Context, limit ...int) ([]Message, error) {
	return []Message{}, nil
}

func (n *NoOpMemory) NewSession() string {
	return GenerateSessionID()
}

func (n *NoOpMemory) SetSession(ctx context.Context, sessionID string) context.Context {
	return ctx
}

func (n *NoOpMemory) ClearSession(ctx context.Context) error {
	return nil
}

func (n *NoOpMemory) Close() error {
	return nil
}
