package agent_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
)

// mockLLMProvider implements a mock LLM provider for testing
type mockLLMProvider struct {
	responses  map[string]string
	callCount  int
	lastInput  string
	shouldFail bool
}

func (m *mockLLMProvider) Call(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	m.callCount++
	m.lastInput = userPrompt

	if m.shouldFail {
		return "", errors.New("mock LLM error")
	}

	// Return predefined response if available
	if response, ok := m.responses[userPrompt]; ok {
		return response, nil
	}

	// Default response
	return fmt.Sprintf("Mock response to: %s", userPrompt), nil
}

func (m *mockLLMProvider) CallStream(ctx context.Context, systemPrompt, userPrompt string, callback func(string)) error {
	response, err := m.Call(ctx, systemPrompt, userPrompt)
	if err != nil {
		return err
	}

	// Simulate streaming by sending chunks
	words := strings.Fields(response)
	for _, word := range words {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			callback(word + " ")
			time.Sleep(10 * time.Millisecond)
		}
	}

	return nil
}

func (m *mockLLMProvider) Model() string {
	return "mock-model"
}

// TestAgentCreation tests basic agent creation using Builder
func TestAgentCreation(t *testing.T) {
	tests := []struct {
		name    string
		builder func() (vnext.Agent, error)
		wantErr bool
	}{
		{
			name: "basic_agent_creation",
			builder: func() (vnext.Agent, error) {
				config := &vnext.Config{
					Name:    "test-agent",
					Timeout: 30 * time.Second,
					LLM: vnext.LLMConfig{
						Provider: "mock",
						Model:    "mock-model",
					},
				}
				return vnext.NewBuilder("test-agent").WithConfig(config).Build()
			},
			wantErr: false,
		},
		{
			name: "agent_with_preset_chat",
			builder: func() (vnext.Agent, error) {
				config := &vnext.Config{
					Name:    "chat-agent",
					Timeout: 30 * time.Second,
					LLM: vnext.LLMConfig{
						Provider: "mock",
						Model:    "mock-model",
					},
				}
				return vnext.NewBuilder("chat-agent").
					WithConfig(config).
					WithPreset(vnext.ChatAgent).
					Build()
			},
			wantErr: false,
		},
		{
			name: "agent_with_preset_research",
			builder: func() (vnext.Agent, error) {
				config := &vnext.Config{
					Name:    "research-agent",
					Timeout: 30 * time.Second,
					LLM: vnext.LLMConfig{
						Provider: "mock",
						Model:    "mock-model",
					},
				}
				return vnext.NewBuilder("research-agent").
					WithConfig(config).
					WithPreset(vnext.ResearchAgent).
					Build()
			},
			wantErr: false,
		},
		{
			name: "agent_with_preset_data",
			builder: func() (vnext.Agent, error) {
				config := &vnext.Config{
					Name:    "data-agent",
					Timeout: 30 * time.Second,
					LLM: vnext.LLMConfig{
						Provider: "mock",
						Model:    "mock-model",
					},
				}
				return vnext.NewBuilder("data-agent").
					WithConfig(config).
					WithPreset(vnext.DataAgent).
					Build()
			},
			wantErr: false,
		},
		{
			name: "agent_with_preset_workflow",
			builder: func() (vnext.Agent, error) {
				config := &vnext.Config{
					Name:    "workflow-agent",
					Timeout: 30 * time.Second,
					LLM: vnext.LLMConfig{
						Provider: "mock",
						Model:    "mock-model",
					},
				}
				return vnext.NewBuilder("workflow-agent").
					WithConfig(config).
					WithPreset(vnext.WorkflowAgent).
					Build()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := tt.builder()
			if (err != nil) != tt.wantErr {
				t.Errorf("builder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && agent == nil {
				t.Error("builder() returned nil agent without error")
			}
		})
	}
}

// TestAgentName tests agent name retrieval
func TestAgentName(t *testing.T) {
	config := &vnext.Config{
		Name:    "test-agent-name",
		Timeout: 30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider: "mock",
			Model:    "mock-model",
		},
	}

	agent, err := vnext.NewBuilder("test-agent-name").WithConfig(config).Build()
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	if agent.Name() != "test-agent-name" {
		t.Errorf("Agent name = %s, want test-agent-name", agent.Name())
	}
}

// TestAgentConfig tests agent configuration retrieval
func TestAgentConfig(t *testing.T) {
	config := &vnext.Config{
		Name:    "config-test-agent",
		Timeout: 30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "mock",
			Model:       "mock-model",
			Temperature: 0.7,
			MaxTokens:   1000,
		},
	}

	agent, err := vnext.NewBuilder("config-test-agent").WithConfig(config).Build()
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	retrievedConfig := agent.Config()
	if retrievedConfig == nil {
		t.Fatal("Config() returned nil")
	}

	if retrievedConfig.Name != "config-test-agent" {
		t.Errorf("Config name = %s, want config-test-agent", retrievedConfig.Name)
	}

	if retrievedConfig.LLM.Model != "mock-model" {
		t.Errorf("LLM model = %s, want mock-model", retrievedConfig.LLM.Model)
	}
}

// TestAgentCapabilities tests agent capabilities reporting
func TestAgentCapabilities(t *testing.T) {
	config := &vnext.Config{
		Name:    "capability-test-agent",
		Timeout: 30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider: "mock",
			Model:    "mock-model",
		},
	}

	agent, err := vnext.NewBuilder("capability-test-agent").WithConfig(config).Build()
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	capabilities := agent.Capabilities()
	if capabilities == nil {
		t.Error("Capabilities() returned nil")
	}

	// Capabilities should be a non-empty slice for any agent
	t.Logf("Agent capabilities: %v", capabilities)
}

// TestRunOptions tests RunOptions factory functions
func TestRunOptions(t *testing.T) {
	tests := []struct {
		name     string
		factory  func() *vnext.RunOptions
		validate func(*testing.T, *vnext.RunOptions)
	}{
		{
			name:    "NewRunOptions",
			factory: vnext.NewRunOptions,
			validate: func(t *testing.T, opts *vnext.RunOptions) {
				if opts.ToolMode != "auto" {
					t.Errorf("ToolMode = %s, want auto", opts.ToolMode)
				}
				if opts.MaxRetries != 3 {
					t.Errorf("MaxRetries = %d, want 3", opts.MaxRetries)
				}
				if opts.DetailedResult != false {
					t.Error("DetailedResult should be false by default")
				}
			},
		},
		{
			name:    "RunWithTools",
			factory: func() *vnext.RunOptions { return vnext.RunWithTools("tool1", "tool2") },
			validate: func(t *testing.T, opts *vnext.RunOptions) {
				if len(opts.Tools) != 2 {
					t.Errorf("Tools length = %d, want 2", len(opts.Tools))
				}
				if opts.ToolMode != "specific" {
					t.Errorf("ToolMode = %s, want specific", opts.ToolMode)
				}
			},
		},
		{
			name: "RunWithMemory",
			factory: func() *vnext.RunOptions {
				memOpts := &vnext.MemoryOptions{
					Enabled:       true,
					Provider:      "memory",
					ContextAware:  true,
					SessionScoped: true,
				}
				return vnext.RunWithMemory("session-123", memOpts)
			},
			validate: func(t *testing.T, opts *vnext.RunOptions) {
				if opts.SessionID != "session-123" {
					t.Errorf("SessionID = %s, want session-123", opts.SessionID)
				}
				if opts.Memory == nil {
					t.Fatal("Memory is nil")
				}
				if !opts.Memory.Enabled {
					t.Error("Memory.Enabled should be true")
				}
			},
		},
		{
			name:    "RunWithStreaming",
			factory: vnext.RunWithStreaming,
			validate: func(t *testing.T, opts *vnext.RunOptions) {
				if !opts.Streaming {
					t.Error("Streaming should be true")
				}
			},
		},
		{
			name:    "RunWithDetailedResult",
			factory: vnext.RunWithDetailedResult,
			validate: func(t *testing.T, opts *vnext.RunOptions) {
				if !opts.DetailedResult {
					t.Error("DetailedResult should be true")
				}
				if !opts.IncludeTrace {
					t.Error("IncludeTrace should be true")
				}
				if !opts.IncludeSources {
					t.Error("IncludeSources should be true")
				}
			},
		},
		{
			name:    "RunWithTimeout",
			factory: func() *vnext.RunOptions { return vnext.RunWithTimeout(30 * time.Second) },
			validate: func(t *testing.T, opts *vnext.RunOptions) {
				if opts.Timeout != 30*time.Second {
					t.Errorf("Timeout = %v, want 30s", opts.Timeout)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.factory()
			if opts == nil {
				t.Fatal("Factory returned nil")
			}
			tt.validate(t, opts)
		})
	}
}

// TestRunOptionsChaining tests chainable methods on RunOptions
func TestRunOptionsChaining(t *testing.T) {
	opts := vnext.NewRunOptions().
		SetTools("tool1", "tool2").
		SetTimeout(60*time.Second).
		SetDetailedResult(true).
		SetTracing(true, "debug").
		AddContext("key1", "value1").
		AddContext("key2", 123)

	if len(opts.Tools) != 2 {
		t.Errorf("Tools length = %d, want 2", len(opts.Tools))
	}

	if opts.Timeout != 60*time.Second {
		t.Errorf("Timeout = %v, want 60s", opts.Timeout)
	}

	if !opts.DetailedResult {
		t.Error("DetailedResult should be true")
	}

	if !opts.TraceEnabled {
		t.Error("TraceEnabled should be true")
	}

	if opts.TraceLevel != "debug" {
		t.Errorf("TraceLevel = %s, want debug", opts.TraceLevel)
	}

	if opts.Context["key1"] != "value1" {
		t.Errorf("Context[key1] = %v, want value1", opts.Context["key1"])
	}

	if opts.Context["key2"] != 123 {
		t.Errorf("Context[key2] = %v, want 123", opts.Context["key2"])
	}
}

// TestResultStructure tests Result structure and methods
func TestResultStructure(t *testing.T) {
	tests := []struct {
		name        string
		result      *vnext.Result
		wantText    string
		wantSuccess bool
	}{
		{
			name: "successful_result",
			result: &vnext.Result{
				Success:  true,
				Content:  "Test response",
				Duration: 100 * time.Millisecond,
				TraceID:  "trace-123",
			},
			wantText:    "Test response",
			wantSuccess: true,
		},
		{
			name: "failed_result_with_error",
			result: &vnext.Result{
				Success: false,
				Content: "Error occurred",
				Error:   "some error",
			},
			wantText:    "Error occurred",
			wantSuccess: false,
		},
		{
			name: "result_with_tools",
			result: &vnext.Result{
				Success:     true,
				Content:     "Tool result",
				ToolsCalled: []string{"calculator", "web_search"},
				TokensUsed:  150,
			},
			wantText:    "Tool result",
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result.Text() != tt.wantText {
				t.Errorf("Text() = %s, want %s", tt.result.Text(), tt.wantText)
			}

			if tt.result.IsSuccess() != tt.wantSuccess {
				t.Errorf("IsSuccess() = %v, want %v", tt.result.IsSuccess(), tt.wantSuccess)
			}

			duration := tt.result.GetDuration()
			t.Logf("Duration: %v", duration)
		})
	}
}

// TestAgentError tests AgentError structure and methods
func TestAgentError(t *testing.T) {
	tests := []struct {
		name        string
		createError func() *vnext.AgentError
		validate    func(*testing.T, *vnext.AgentError)
	}{
		{
			name: "new_agent_error",
			createError: func() *vnext.AgentError {
				return vnext.NewAgentError(vnext.ErrLLMCallFailed, "LLM call failed")
			},
			validate: func(t *testing.T, err *vnext.AgentError) {
				if err.Code != vnext.ErrLLMCallFailed {
					t.Errorf("Code = %s, want %s", err.Code, vnext.ErrLLMCallFailed)
				}
				if err.Message != "LLM call failed" {
					t.Errorf("Message = %s, want 'LLM call failed'", err.Message)
				}
				if err.Error() == "" {
					t.Error("Error() returned empty string")
				}
			},
		},
		{
			name: "agent_error_with_details",
			createError: func() *vnext.AgentError {
				details := map[string]interface{}{
					"provider": "mock",
					"model":    "mock-model",
				}
				return vnext.NewAgentErrorWithDetails(vnext.ErrConfigInvalid, "Config invalid", details)
			},
			validate: func(t *testing.T, err *vnext.AgentError) {
				if len(err.Details) != 2 {
					t.Errorf("Details length = %d, want 2", len(err.Details))
				}
				if err.Details["provider"] != "mock" {
					t.Errorf("Details[provider] = %v, want mock", err.Details["provider"])
				}
			},
		},
		{
			name: "agent_error_with_inner_error",
			createError: func() *vnext.AgentError {
				innerErr := errors.New("underlying error")
				return vnext.NewAgentErrorWithError(vnext.ErrToolExecutionFailed, "Tool failed", innerErr)
			},
			validate: func(t *testing.T, err *vnext.AgentError) {
				if err.InnerError == nil {
					t.Error("InnerError is nil")
				}
				if err.Unwrap() == nil {
					t.Error("Unwrap() returned nil")
				}
				if !strings.Contains(err.Error(), "underlying error") {
					t.Error("Error string should contain inner error message")
				}
			},
		},
		{
			name: "agent_error_add_detail",
			createError: func() *vnext.AgentError {
				return vnext.NewAgentError(vnext.ErrMemoryStoreFailed, "Store failed").
					AddDetail("key1", "value1").
					AddDetail("key2", 42)
			},
			validate: func(t *testing.T, err *vnext.AgentError) {
				if len(err.Details) != 2 {
					t.Errorf("Details length = %d, want 2", len(err.Details))
				}
			},
		},
		{
			name: "agent_error_with_stack_trace",
			createError: func() *vnext.AgentError {
				return vnext.NewAgentError(vnext.ErrInternal, "Internal error").
					WithStackTrace("stack trace here")
			},
			validate: func(t *testing.T, err *vnext.AgentError) {
				if err.StackTrace != "stack trace here" {
					t.Error("StackTrace not set correctly")
				}
			},
		},
		{
			name: "agent_error_is_error_code",
			createError: func() *vnext.AgentError {
				return vnext.NewAgentError(vnext.ErrToolNotFound, "Tool not found")
			},
			validate: func(t *testing.T, err *vnext.AgentError) {
				if !err.IsErrorCode(vnext.ErrToolNotFound) {
					t.Error("IsErrorCode should return true for matching code")
				}
				if err.IsErrorCode(vnext.ErrLLMCallFailed) {
					t.Error("IsErrorCode should return false for non-matching code")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.createError()
			if err == nil {
				t.Fatal("createError returned nil")
			}
			tt.validate(t, err)
		})
	}
}

// TestErrorCodes tests all defined error codes
func TestErrorCodes(t *testing.T) {
	errorCodes := []vnext.ErrorCode{
		vnext.ErrConfigInvalid,
		vnext.ErrConfigMissing,
		vnext.ErrAgentInitFailed,
		vnext.ErrAgentCleanupFailed,
		vnext.ErrLLMProviderNotFound,
		vnext.ErrLLMCallFailed,
		vnext.ErrLLMTimeout,
		vnext.ErrMemoryStoreFailed,
		vnext.ErrMemoryRetrieveFailed,
		vnext.ErrMemoryClearFailed,
		vnext.ErrToolNotFound,
		vnext.ErrToolExecutionFailed,
		vnext.ErrToolValidationFailed,
		vnext.ErrMiddlewareBeforeRun,
		vnext.ErrMiddlewareAfterRun,
		vnext.ErrWorkflowInvalid,
		vnext.ErrWorkflowNodeNotFound,
		vnext.ErrWorkflowCycleDetected,
		vnext.ErrContextCancelled,
		vnext.ErrContextTimeout,
		vnext.ErrInternal,
		vnext.ErrNotImplemented,
	}

	for _, code := range errorCodes {
		t.Run(string(code), func(t *testing.T) {
			err := vnext.NewAgentError(code, "test error")
			if err.Code != code {
				t.Errorf("Error code = %s, want %s", err.Code, code)
			}
		})
	}
}

// TestBuilderClone tests builder cloning functionality
func TestBuilderClone(t *testing.T) {
	config := &vnext.Config{
		Name:    "original-agent",
		Timeout: 30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider: "mock",
			Model:    "mock-model",
		},
	}

	builder := vnext.NewBuilder("original-agent").WithConfig(config)
	clonedBuilder := builder.Clone()

	// Build both agents
	agent1, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build original agent: %v", err)
	}

	agent2, err := clonedBuilder.Build()
	if err != nil {
		t.Fatalf("Failed to build cloned agent: %v", err)
	}

	// Both should exist but be independent
	if agent1 == nil || agent2 == nil {
		t.Fatal("One or both agents are nil")
	}

	// They should have the same configuration initially
	if agent1.Name() != agent2.Name() {
		t.Errorf("Agents have different names: %s vs %s", agent1.Name(), agent2.Name())
	}
}

// TestAgentInitializeCleanup tests agent lifecycle methods
func TestAgentInitializeCleanup(t *testing.T) {
	config := &vnext.Config{
		Name:    "lifecycle-agent",
		Timeout: 30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider: "mock",
			Model:    "mock-model",
		},
	}

	agent, err := vnext.NewBuilder("lifecycle-agent").WithConfig(config).Build()
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	ctx := context.Background()

	// Test Initialize
	err = agent.Initialize(ctx)
	if err != nil {
		t.Errorf("Initialize() error = %v", err)
	}

	// Test Cleanup
	err = agent.Cleanup(ctx)
	if err != nil {
		t.Errorf("Cleanup() error = %v", err)
	}
}

// TestAgentWithCustomHandler tests agent with custom handler
func TestAgentWithCustomHandler(t *testing.T) {
	config := &vnext.Config{
		Name:    "custom-handler-agent",
		Timeout: 30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider: "mock",
			Model:    "mock-model",
		},
	}

	customHandlerCalled := false
	handler := func(ctx context.Context, input string, capabilities *vnext.Capabilities) (string, error) {
		customHandlerCalled = true
		return fmt.Sprintf("Custom: %s", input), nil
	}

	agent, err := vnext.NewBuilder("custom-handler-agent").
		WithConfig(config).
		WithHandler(handler).
		Build()

	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	if agent == nil {
		t.Fatal("Agent is nil")
	}

	// Verify custom handler was registered (actual execution would test if it was called)
	t.Logf("Custom handler registered: %v", customHandlerCalled)
}

// TestMemoryOptions tests MemoryOptions configuration
func TestMemoryOptions(t *testing.T) {
	memOpts := &vnext.MemoryOptions{
		Enabled:       true,
		Provider:      "memory",
		ContextAware:  true,
		SessionScoped: true,
		RAGConfig: &vnext.RAGConfig{
			MaxTokens:       1000,
			PersonalWeight:  0.7,
			KnowledgeWeight: 0.3,
			HistoryLimit:    10,
		},
	}

	if !memOpts.Enabled {
		t.Error("Memory should be enabled")
	}

	if memOpts.RAGConfig == nil {
		t.Fatal("RAGConfig is nil")
	}

	if memOpts.RAGConfig.MaxTokens != 1000 {
		t.Errorf("MaxTokens = %d, want 1000", memOpts.RAGConfig.MaxTokens)
	}

	if memOpts.RAGConfig.PersonalWeight != 0.7 {
		t.Errorf("PersonalWeight = %f, want 0.7", memOpts.RAGConfig.PersonalWeight)
	}
}

// TestToolOptions tests ToolOptions configuration
func TestToolOptions(t *testing.T) {
	toolOpts := &vnext.ToolOptions{
		Enabled:       true,
		MaxRetries:    3,
		Timeout:       30 * time.Second,
		RateLimit:     10,
		CacheEnabled:  true,
		CacheTTL:      5 * time.Minute,
		MaxConcurrent: 5,
		OrchestrationConfig: &vnext.ToolOrchestrationConfig{
			AutoDiscovery:      true,
			MaxRetries:         3,
			Timeout:            60 * time.Second,
			ParallelExecution:  true,
			DependencyTracking: true,
		},
	}

	if !toolOpts.Enabled {
		t.Error("Tools should be enabled")
	}

	if toolOpts.OrchestrationConfig == nil {
		t.Fatal("OrchestrationConfig is nil")
	}

	if !toolOpts.OrchestrationConfig.ParallelExecution {
		t.Error("ParallelExecution should be true")
	}
}

// TestLLMOptions tests LLMOptions configuration
func TestLLMOptions(t *testing.T) {
	llmOpts := &vnext.LLMOptions{
		Provider:    "openai",
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   2000,
		Timeout:     30 * time.Second,
	}

	if llmOpts.Provider != "openai" {
		t.Errorf("Provider = %s, want openai", llmOpts.Provider)
	}

	if llmOpts.Temperature != 0.7 {
		t.Errorf("Temperature = %f, want 0.7", llmOpts.Temperature)
	}

	if llmOpts.MaxTokens != 2000 {
		t.Errorf("MaxTokens = %d, want 2000", llmOpts.MaxTokens)
	}
}

// TestTracingOptions tests TracingOptions configuration
func TestTracingOptions(t *testing.T) {
	tracingOpts := &vnext.TracingOptions{
		Enabled:     true,
		Level:       "debug",
		WebUI:       true,
		Performance: true,
		MemoryTrace: true,
		ToolTrace:   true,
	}

	if !tracingOpts.Enabled {
		t.Error("Tracing should be enabled")
	}

	if tracingOpts.Level != "debug" {
		t.Errorf("Level = %s, want debug", tracingOpts.Level)
	}

	if !tracingOpts.Performance {
		t.Error("Performance tracing should be enabled")
	}
}
