package v1beta

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agenticgokit/agenticgokit/core"
)

type storeRecordingMemory struct {
	stored       []string
	storeCalls   []string
	messageCalls []string
}

func (m *storeRecordingMemory) Store(ctx context.Context, content string, tags ...string) error {
	m.stored = append(m.stored, content)
	m.storeCalls = append(m.storeCalls, content)
	return nil
}

func (m *storeRecordingMemory) Query(ctx context.Context, query string, limit ...int) ([]core.Result, error) {
	return nil, nil
}

func (m *storeRecordingMemory) Remember(ctx context.Context, key string, value any) error { return nil }
func (m *storeRecordingMemory) Recall(ctx context.Context, key string) (any, error)       { return nil, nil }
func (m *storeRecordingMemory) AddMessage(ctx context.Context, role, content string) error {
	entry := role + ":" + content
	m.stored = append(m.stored, entry)
	m.messageCalls = append(m.messageCalls, entry)
	return nil
}
func (m *storeRecordingMemory) GetHistory(ctx context.Context, limit ...int) ([]core.Message, error) {
	return nil, nil
}
func (m *storeRecordingMemory) NewSession() string { return "session-1" }
func (m *storeRecordingMemory) SetSession(ctx context.Context, sessionID string) context.Context {
	return ctx
}
func (m *storeRecordingMemory) ClearSession(ctx context.Context) error { return nil }
func (m *storeRecordingMemory) Close() error                           { return nil }
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

func TestStoreInMemoryPromptInjectionPoCBlocksLLMOutputFromPersonalMemory(t *testing.T) {
	preGuardMem := &storeRecordingMemory{}
	mem := &storeRecordingMemory{}
	agent := &realAgent{memoryProvider: mem}
	payload := "POC_MARKER model output that must not reach personal memory"

	if err := preGuardMem.Store(context.Background(), payload, "assistant_output", "conversation"); err != nil {
		t.Fatalf("pre-guard fake memory store returned error: %v", err)
	}
	if len(preGuardMem.storeCalls) != 1 || !strings.Contains(preGuardMem.storeCalls[0], "POC_MARKER") {
		t.Fatalf("PoC setup failed: fake DB recorder did not store model-controlled marker: %#v", preGuardMem.storeCalls)
	}

	if err := agent.storeInMemory(context.Background(), "trusted user prompt", payload); err != nil {
		t.Fatalf("storeInMemory returned error: %v", err)
	}

	for _, stored := range mem.storeCalls {
		if strings.Contains(stored, "POC_MARKER") {
			t.Fatalf("fixed code sent model-controlled payload to personal memory DB sink: %#v", mem.storeCalls)
		}
	}
	if len(mem.storeCalls) != 1 || mem.storeCalls[0] != "trusted user prompt" {
		t.Fatalf("expected only trusted user input in personal memory, got %#v", mem.storeCalls)
	}

	if len(mem.messageCalls) != 2 {
		t.Fatalf("expected chat history messages to be preserved, got %#v", mem.messageCalls)
	}
	if mem.messageCalls[1] != "assistant:"+payload {
		t.Fatalf("expected assistant output to remain only in chat history, got %#v", mem.messageCalls)
	}
}

func TestStoreInMemoryReturnsNoErrorWithoutMemoryProvider(t *testing.T) {
	agent := &realAgent{}

	if err := agent.storeInMemory(context.Background(), "user prompt", "assistant output"); err != nil {
		t.Fatalf("expected nil error with no memory provider, got %v", err)
	}
}

func TestGenerateManifestAggregatesLLMSpanTokens(t *testing.T) {
	runDir := t.TempDir()
	trace := []map[string]interface{}{
		{
			"Name":      "agk.agent.run",
			"StartTime": "2026-01-01T00:00:00Z",
			"EndTime":   "2026-01-01T00:00:01Z",
			"Attributes": []map[string]interface{}{
				{"Key": "agk.llm.tokens.input", "Value": map[string]interface{}{"Type": "INT64", "Value": 99}},
				{"Key": "agk.llm.tokens.output", "Value": map[string]interface{}{"Type": "INT64", "Value": 101}},
			},
		},
		{
			"Name":      "llm.openai.call",
			"StartTime": "2026-01-01T00:00:00Z",
			"EndTime":   "2026-01-01T00:00:01Z",
			"Attributes": []map[string]interface{}{
				{"Key": "agk.llm.tokens.total", "Value": map[string]interface{}{"Type": "INT64", "Value": 7}},
			},
		},
		{
			"Name":      "llm.anthropic.call",
			"StartTime": "2026-01-01T00:00:01Z",
			"EndTime":   "2026-01-01T00:00:02Z",
			"Attributes": []map[string]interface{}{
				{"Key": "agk.llm.tokens.prompt", "Value": map[string]interface{}{"Type": "INT64", "Value": 4}},
				{"Key": "agk.llm.tokens.completion", "Value": map[string]interface{}{"Type": "INT64", "Value": 6}},
			},
		},
	}
	var data []byte
	for _, span := range trace {
		encoded, err := json.Marshal(span)
		if err != nil {
			t.Fatalf("marshal trace fixture: %v", err)
		}
		data = append(data, encoded...)
		data = append(data, '\n')
	}
	if err := os.WriteFile(filepath.Join(runDir, "trace.jsonl"), data, 0644); err != nil {
		t.Fatalf("write trace fixture: %v", err)
	}

	agent := &realAgent{runDir: runDir, runID: "run-1", config: &Config{Name: "test-agent"}}
	if err := agent.generateManifest(); err != nil {
		t.Fatalf("generateManifest returned error: %v", err)
	}

	manifestData, err := os.ReadFile(filepath.Join(runDir, "manifest.json"))
	if err != nil {
		t.Fatalf("read generated manifest: %v", err)
	}
	var manifest struct {
		LLMCalls    int `json:"llm_calls"`
		TotalTokens int `json:"total_tokens"`
	}
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatalf("decode generated manifest: %v", err)
	}
	if manifest.LLMCalls != 2 {
		t.Fatalf("expected 2 LLM calls, got %d", manifest.LLMCalls)
	}
	if manifest.TotalTokens != 17 {
		t.Fatalf("expected 17 total tokens without root-span double counting, got %d", manifest.TotalTokens)
	}
}
