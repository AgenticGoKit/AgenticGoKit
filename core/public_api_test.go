package core

import (
	"context"
	"testing"
	"time"
)

// TestPublicAPIComprehensive tests all essential public APIs in the core package
func TestPublicAPIComprehensive(t *testing.T) {
	t.Run("ConfigurationAPI", func(t *testing.T) {
		// Test LoadConfig function
		config, err := LoadConfig("")
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}
		if config == nil {
			t.Fatal("LoadConfig returned nil config")
		}
		if config.AgentFlow.Name == "" {
			t.Error("AgentFlow.Name should have a default value")
		}
	})

	t.Run("LLMProviderAPI", func(t *testing.T) {
		// Test LLM provider creation functions
		_, err := NewOpenAIAdapter("test-key", "gpt-4", 100, 0.7)
		if err != nil {
			t.Errorf("NewOpenAIAdapter failed: %v", err)
		}

		_, err = NewOllamaAdapter("http://localhost:11434", "llama2", 100, 0.7)
		if err != nil {
			t.Errorf("NewOllamaAdapter failed: %v", err)
		}

		// Test model provider from config
		_, err = NewModelProviderFromConfig(LLMProviderConfig{
			Type:        "openai",
			APIKey:      "test-key",
			Model:       "gpt-4",
			MaxTokens:   100,
			Temperature: 0.7,
		})
		if err != nil {
			t.Errorf("NewModelProviderFromConfig failed: %v", err)
		}
	})

	t.Run("MemoryAPI", func(t *testing.T) {
		// Test memory creation
		memory, err := NewMemory(AgentMemoryConfig{
			Provider:   "inmemory",
			Connection: "",
		})
		if err != nil {
			t.Errorf("NewMemory failed: %v", err)
		}
		if memory == nil {
			t.Error("NewMemory returned nil memory")
		}

		// Test QuickMemory
		quickMem := QuickMemory()
		if quickMem == nil {
			t.Error("QuickMemory returned nil")
		}
	})

	t.Run("EventAPI", func(t *testing.T) {
		// Test event creation
		eventData := EventData{"test": "data"}
		metadata := map[string]string{"type": "test"}
		event := NewEvent("test-agent", eventData, metadata)
		if event == nil {
			t.Error("NewEvent returned nil")
		}
		if event.GetID() == "" {
			t.Error("Event should have an ID")
		}
		if event.GetTimestamp().IsZero() {
			t.Error("Event should have a timestamp")
		}

		// Test event data manipulation
		event.SetData("key", "value")
		data := event.GetData()
		if data["key"] != "value" {
			t.Error("Event data not set correctly")
		}

		// Test metadata
		event.SetMetadata("meta-key", "meta-value")
		if value, ok := event.GetMetadataValue("meta-key"); !ok || value != "meta-value" {
			t.Error("Event metadata not set correctly")
		}
	})

	t.Run("StateAPI", func(t *testing.T) {
		// Test state creation
		state := NewState()
		if state == nil {
			t.Error("NewState returned nil")
		}

		// Test state operations
		state.Set("key", "value")
		if value, ok := state.Get("key"); !ok || value != "value" {
			t.Error("State data not set correctly")
		}

		state.SetMeta("meta-key", "meta-value")
		if value, ok := state.GetMeta("meta-key"); !ok || value != "meta-value" {
			t.Error("State metadata not set correctly")
		}

		// Test state cloning
		cloned := state.Clone()
		if cloned == nil {
			t.Error("State clone returned nil")
		}
		if value, ok := cloned.Get("key"); !ok || value != "value" {
			t.Error("State clone did not preserve data")
		}
	})

	t.Run("OrchestratorAPI", func(t *testing.T) {
		// Test callback registry creation (core functionality)
		registry := NewCallbackRegistry()
		if registry == nil {
			t.Error("NewCallbackRegistry returned nil")
		}

		// Test orchestrator interface can be implemented
		var _ Orchestrator = &mockOrchestrator{}

		// Test orchestration modes are defined
		if OrchestrationRoute == "" {
			t.Error("OrchestrationRoute should be defined")
		}
		if OrchestrationCollaborate == "" {
			t.Error("OrchestrationCollaborate should be defined")
		}
	})

	t.Run("RunnerAPI", func(t *testing.T) {
		// Test runner creation
		runner := NewRunner(100) // queue size
		if runner == nil {
			t.Error("NewRunner returned nil")
		}

		// Test runner interface can be implemented
		var _ Runner = &mockRunner{}

		// Test hook points are defined
		if HookBeforeAgentRun == "" {
			t.Error("HookBeforeAgentRun should be defined")
		}
		if HookAfterAgentRun == "" {
			t.Error("HookAfterAgentRun should be defined")
		}
	})

	t.Run("FactoryAPI", func(t *testing.T) {
		// Test helper functions
		temp := FloatPtr(0.7)
		if *temp != 0.7 {
			t.Error("FloatPtr helper function failed")
		}

		tokens := Int32Ptr(100)
		if *tokens != 100 {
			t.Error("Int32Ptr helper function failed")
		}

		// Test callback registry creation
		registry := NewCallbackRegistry()
		if registry == nil {
			t.Error("NewCallbackRegistry returned nil")
		}
	})

	t.Run("TypesAndStructures", func(t *testing.T) {
		// Test Prompt structure
		prompt := Prompt{
			System: "You are helpful",
			User:   "Hello",
			Parameters: ModelParameters{
				Temperature: FloatPtr(0.7),
				MaxTokens:   Int32Ptr(100),
			},
		}
		if prompt.System != "You are helpful" {
			t.Error("Prompt creation failed")
		}

		// Test Response structure
		response := Response{
			Content: "Hello back!",
			Usage: UsageStats{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
			FinishReason: "stop",
		}
		if response.Content != "Hello back!" {
			t.Error("Response creation failed")
		}

		// Test AgentResult structure
		result := AgentResult{
			OutputState: NewState(),
			StartTime:   time.Now(),
			EndTime:     time.Now(),
			Duration:    time.Millisecond * 100,
		}
		if result.OutputState == nil {
			t.Error("AgentResult creation failed")
		}
	})

	t.Run("ContextAPI", func(t *testing.T) {
		// Test context with memory
		ctx := context.Background()
		memory := QuickMemory()
		sessionID := "test-session"

		ctxWithMemory := WithMemory(ctx, memory, sessionID)
		if ctxWithMemory == nil {
			t.Error("WithMemory returned nil context")
		}

		// Test memory retrieval from context
		retrievedMemory := GetMemory(ctxWithMemory)
		if retrievedMemory == nil {
			t.Error("GetMemory returned nil")
		}

		// Test session ID retrieval
		retrievedSessionID := GetSessionID(ctxWithMemory)
		if retrievedSessionID != sessionID {
			t.Errorf("Expected session ID %s, got %s", sessionID, retrievedSessionID)
		}
	})

	t.Run("ErrorHandlingAPI", func(t *testing.T) {
		// Test error event data creation
		testEvent := NewEvent("test-agent", EventData{"test": "data"}, map[string]string{})
		errorData := ErrorEventData{
			OriginalEvent: testEvent,
			FailedAgent:   "test-agent",
			ErrorMessage:  "test error",
			ErrorCode:     ErrorCodeValidation,
			RetryCount:    1,
			Timestamp:     time.Now(),
			SessionID:     "test-session",
			Severity:      SeverityMedium,
			ErrorCategory: "validation",
		}

		if errorData.ErrorCode != ErrorCodeValidation {
			t.Error("ErrorEventData creation failed")
		}
		if errorData.Severity != SeverityMedium {
			t.Error("Error severity not set correctly")
		}
	})
}

// TestInterfaceImplementation verifies that all public interfaces can be implemented
func TestInterfaceImplementation(t *testing.T) {
	t.Run("AgentInterface", func(t *testing.T) {
		var _ Agent = &mockAgent{}
	})

	t.Run("AgentHandlerInterface", func(t *testing.T) {
		var _ AgentHandler = &mockAgentHandler{}
	})

	t.Run("ModelProviderInterface", func(t *testing.T) {
		var _ ModelProvider = &mockModelProvider{}
	})

	t.Run("MemoryInterface", func(t *testing.T) {
		var _ Memory = &mockMemory{}
	})

	t.Run("OrchestratorInterface", func(t *testing.T) {
		var _ Orchestrator = &mockOrchestrator{}
	})

	t.Run("RunnerInterface", func(t *testing.T) {
		var _ Runner = &mockRunner{}
	})

	t.Run("EventInterface", func(t *testing.T) {
		var _ Event = &mockEvent{}
	})

	t.Run("StateInterface", func(t *testing.T) {
		var _ State = &mockState{}
	})
}

// Mock implementations for interface testing
type mockAgent struct{}

func (a *mockAgent) Run(ctx context.Context, inputState State) (State, error) { return inputState, nil }
func (a *mockAgent) HandleEvent(ctx context.Context, event Event, state State) (AgentResult, error) {
	return AgentResult{OutputState: state}, nil
}
func (a *mockAgent) Name() string                         { return "mock-agent" }
func (a *mockAgent) GetRole() string                      { return "mock-role" }
func (a *mockAgent) GetDescription() string               { return "Mock agent for testing" }
func (a *mockAgent) GetCapabilities() []string            { return []string{"mock"} }
func (a *mockAgent) GetSystemPrompt() string              { return "You are a mock agent." }
func (a *mockAgent) GetTimeout() time.Duration            { return 30 * time.Second }
func (a *mockAgent) IsEnabled() bool                      { return true }
func (a *mockAgent) GetLLMConfig() *ResolvedLLMConfig     { return nil }
func (a *mockAgent) Initialize(ctx context.Context) error { return nil }
func (a *mockAgent) Shutdown(ctx context.Context) error   { return nil }

type mockAgentHandler struct{}

func (h *mockAgentHandler) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
	return AgentResult{OutputState: state}, nil
}

type mockModelProvider struct{}

func (p *mockModelProvider) Call(ctx context.Context, prompt Prompt) (Response, error) {
	return Response{Content: "mock response"}, nil
}
func (p *mockModelProvider) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
	ch := make(chan Token)
	close(ch)
	return ch, nil
}
func (p *mockModelProvider) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return [][]float64{{0.1, 0.2, 0.3}}, nil
}

type mockMemory struct{}

func (m *mockMemory) Store(ctx context.Context, content string, tags ...string) error { return nil }
func (m *mockMemory) Query(ctx context.Context, query string, limit ...int) ([]Result, error) {
	return nil, nil
}
func (m *mockMemory) Remember(ctx context.Context, key string, value any) error  { return nil }
func (m *mockMemory) Recall(ctx context.Context, key string) (any, error)        { return nil, nil }
func (m *mockMemory) AddMessage(ctx context.Context, role, content string) error { return nil }
func (m *mockMemory) GetHistory(ctx context.Context, limit ...int) ([]Message, error) {
	return nil, nil
}
func (m *mockMemory) NewSession() string                                               { return "mock-session" }
func (m *mockMemory) SetSession(ctx context.Context, sessionID string) context.Context { return ctx }
func (m *mockMemory) ClearSession(ctx context.Context) error                           { return nil }
func (m *mockMemory) Close() error                                                     { return nil }
func (m *mockMemory) IngestDocument(ctx context.Context, doc Document) error           { return nil }
func (m *mockMemory) IngestDocuments(ctx context.Context, docs []Document) error       { return nil }
func (m *mockMemory) SearchKnowledge(ctx context.Context, query string, options ...SearchOption) ([]KnowledgeResult, error) {
	return nil, nil
}
func (m *mockMemory) SearchAll(ctx context.Context, query string, options ...SearchOption) (*HybridResult, error) {
	return nil, nil
}
func (m *mockMemory) BuildContext(ctx context.Context, query string, options ...ContextOption) (*RAGContext, error) {
	return nil, nil
}

type mockOrchestrator struct{}

func (o *mockOrchestrator) Dispatch(ctx context.Context, event Event) (AgentResult, error) {
	return AgentResult{}, nil
}
func (o *mockOrchestrator) RegisterAgent(name string, handler AgentHandler) error { return nil }
func (o *mockOrchestrator) GetCallbackRegistry() *CallbackRegistry                { return NewCallbackRegistry() }
func (o *mockOrchestrator) Stop()                                                 {}

type mockRunner struct{}

func (r *mockRunner) Emit(event Event) error                                              { return nil }
func (r *mockRunner) RegisterAgent(name string, handler AgentHandler) error               { return nil }
func (r *mockRunner) RegisterCallback(hook HookPoint, name string, cb CallbackFunc) error { return nil }
func (r *mockRunner) UnregisterCallback(hook HookPoint, name string)                      {}
func (r *mockRunner) Start(ctx context.Context) error                                     { return nil }
func (r *mockRunner) Stop()                                                               {}
func (r *mockRunner) Wait()                                                               {}
func (r *mockRunner) GetTraceLogger() TraceLogger                                         { return nil }
func (r *mockRunner) DumpTrace(sessionID string) ([]TraceEntry, error)                    { return []TraceEntry{}, nil }
func (r *mockRunner) GetCallbackRegistry() *CallbackRegistry                              { return NewCallbackRegistry() }

type mockEvent struct{}

func (e *mockEvent) GetID() string                              { return "mock-id" }
func (e *mockEvent) GetTimestamp() time.Time                    { return time.Now() }
func (e *mockEvent) GetTargetAgentID() string                   { return "mock-target" }
func (e *mockEvent) GetSourceAgentID() string                   { return "mock-source" }
func (e *mockEvent) GetData() EventData                         { return EventData{} }
func (e *mockEvent) GetMetadata() map[string]string             { return map[string]string{} }
func (e *mockEvent) GetSessionID() string                       { return "mock-session" }
func (e *mockEvent) GetMetadataValue(key string) (string, bool) { return "", false }
func (e *mockEvent) SetID(id string)                            {}
func (e *mockEvent) SetTargetAgentID(id string)                 {}
func (e *mockEvent) SetSourceAgentID(id string)                 {}
func (e *mockEvent) SetData(key string, value any)              {}
func (e *mockEvent) SetMetadata(key string, value string)       {}

type mockState struct{}

func (s *mockState) Get(key string) (any, bool)        { return nil, false }
func (s *mockState) Set(key string, value any)         {}
func (s *mockState) GetMeta(key string) (string, bool) { return "", false }
func (s *mockState) SetMeta(key string, value string)  {}
func (s *mockState) Keys() []string                    { return []string{} }
func (s *mockState) MetaKeys() []string                { return []string{} }
func (s *mockState) Clone() State                      { return &mockState{} }
func (s *mockState) Merge(source State)                {}
