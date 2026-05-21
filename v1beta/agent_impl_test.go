package v1beta

import (
	"context"
	"testing"

	"github.com/agenticgokit/agenticgokit/core"
)

type storeRecordingMemory struct {
	stored []string
}

func (m *storeRecordingMemory) Store(ctx context.Context, content string, tags ...string) error {
	m.stored = append(m.stored, content)
	return nil
}

func (m *storeRecordingMemory) Query(ctx context.Context, query string, limit ...int) ([]core.Result, error) {
	return nil, nil
}

func (m *storeRecordingMemory) Remember(ctx context.Context, key string, value any) error { return nil }
func (m *storeRecordingMemory) Recall(ctx context.Context, key string) (any, error)       { return nil, nil }
func (m *storeRecordingMemory) AddMessage(ctx context.Context, role, content string) error {
	m.stored = append(m.stored, role+":"+content)
	return nil
}
func (m *storeRecordingMemory) GetHistory(ctx context.Context, limit ...int) ([]core.Message, error) {
	return nil, nil
}
func (m *storeRecordingMemory) NewSession() string                                        { return "session-1" }
func (m *storeRecordingMemory) SetSession(ctx context.Context, sessionID string) context.Context { return ctx }
func (m *storeRecordingMemory) ClearSession(ctx context.Context) error                     { return nil }
func (m *storeRecordingMemory) Close() error                                               { return nil }
func (m *storeRecordingMemory) IngestDocument(ctx context.Context, doc core.Document) error {
	return nil
}
func (m *storeRecordingMemory) IngestDocuments(ctx context.Context, docs []core.Document) error {
	return nil
}
func (m *storeRecordingMemory) SearchKnowledge(ctx context.Context, query string, options ...core.SearchOption) ([]core.KnowledgeResult, error) {
	return nil, nil
}
func (m *storeRecordingMemory) SearchAll(ctx context.Context, query string, options ...core.SearchOption) (*core.HybridResult, error) {
	return nil, nil
}
func (m *storeRecordingMemory) BuildContext(ctx context.Context, query string, options ...core.ContextOption) (*core.RAGContext, error) {
	return nil, nil
}

func TestStoreInMemoryStoresOnlyTrustedInputInPersonalMemory(t *testing.T) {
	mem := &storeRecordingMemory{}
	agent := &realAgent{memoryProvider: mem}

	if err := agent.storeInMemory(context.Background(), "user prompt", "assistant output"); err != nil {
		t.Fatalf("storeInMemory returned error: %v", err)
	}

	if len(mem.stored) != 3 {
		t.Fatalf("expected 3 memory writes, got %d: %#v", len(mem.stored), mem.stored)
	}

	if mem.stored[0] != "user prompt" {
		t.Fatalf("expected first personal memory write to be user input, got %q", mem.stored[0])
	}

	if mem.stored[1] != "user:user prompt" {
		t.Fatalf("expected first chat history write to be user input, got %q", mem.stored[1])
	}

	if mem.stored[2] != "assistant:assistant output" {
		t.Fatalf("expected assistant output only in chat history, got %q", mem.stored[2])
	}
}

func TestStoreInMemoryReturnsNoErrorWithoutMemoryProvider(t *testing.T) {
	agent := &realAgent{}

	if err := agent.storeInMemory(context.Background(), "user prompt", "assistant output"); err != nil {
		t.Fatalf("expected nil error with no memory provider, got %v", err)
	}
}
