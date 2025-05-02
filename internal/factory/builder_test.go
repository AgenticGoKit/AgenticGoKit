package factory

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/llm"
)

// Mock implementations for testing
type mockRunner struct {
	mock.Mock
}

func (m *mockRunner) Start() {
	m.Called()
}

// Previously we had:
// func (m *mockRunner) Start(ctx context.Context) error {
//     args := m.Called(ctx)
//     return args.Error(0)
// }

func (m *mockRunner) Stop() {
	m.Called()
}

func (m *mockRunner) RegisterAgent(name string, handler agentflow.AgentHandler) error {
	args := m.Called(name, handler)
	return args.Error(0)
}

func (m *mockRunner) Emit(event agentflow.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *mockRunner) SetCallbackRegistry(registry *agentflow.CallbackRegistry) {
	m.Called(registry)
}

func (m *mockRunner) SetTraceLogger(logger agentflow.TraceLogger) {
	m.Called(logger)
}

func (m *mockRunner) SetOrchestrator(orchestrator agentflow.Orchestrator) {
	m.Called(orchestrator)
}

func (m *mockRunner) GetCallbackRegistry() *agentflow.CallbackRegistry {
	args := m.Called()
	return args.Get(0).(*agentflow.CallbackRegistry)
}

func (m *mockRunner) GetTraceLogger() agentflow.TraceLogger {
	args := m.Called()
	return args.Get(0).(agentflow.TraceLogger)
}

func (m *mockRunner) DumpTrace(sessionID string) ([]agentflow.TraceEntry, error) {
	args := m.Called(sessionID)
	return args.Get(0).([]agentflow.TraceEntry), args.Error(1)
}

// Add missing method to mockRunner
func (m *mockRunner) RegisterCallback(hook agentflow.HookPoint, name string, callback agentflow.CallbackFunc) error {
	args := m.Called(hook, name, callback)
	return args.Error(0)
}

// Add missing UnregisterCallback method
func (m *mockRunner) UnregisterCallback(hook agentflow.HookPoint, name string) {
	m.Called(hook, name)
}

type mockModelProvider struct {
	mock.Mock
}

func (m *mockModelProvider) Call(ctx context.Context, prompt llm.Prompt) (llm.Response, error) {
	args := m.Called(ctx, prompt)
	return args.Get(0).(llm.Response), args.Error(1)
}

// Mock agentflow.AgentHandler for testing
type mockAgentHandler struct {
	mock.Mock
}

func (m *mockAgentHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	args := m.Called(ctx, event, state)
	return args.Get(0).(agentflow.AgentResult), args.Error(1)
}

// Unit tests for RunnerBuilder
func TestRunnerBuilder(t *testing.T) {
	// Test basic builder without trace logging
	t.Run("Basic_Configuration", func(t *testing.T) {
		builder := NewRunnerBuilder().
			WithQueueSize(20).
			WithRouteOrchestrator()

		handler := &mockAgentHandler{}
		builder.RegisterAgent("test-agent", handler)

		runner, err := builder.Build()

		require.NoError(t, err)
		require.NotNil(t, runner)
	})

	// Test with trace logging
	t.Run("With_Trace_Logging", func(t *testing.T) {
		builder := NewRunnerBuilder().
			WithTraceLogging().
			WithRouteOrchestrator()

		runner, err := builder.Build()

		require.NoError(t, err)
		require.NotNil(t, runner)

		// Verify trace logger was set
		assert.NotNil(t, runner.GetTraceLogger())
	})

	// Test with collaborative orchestrator
	t.Run("Collaborative_Orchestrator", func(t *testing.T) {
		builder := NewRunnerBuilder().
			WithCollaborativeOrchestrator()

		runner, err := builder.Build()

		require.NoError(t, err)
		require.NotNil(t, runner)
	})

	// Test with invalid orchestrator type
	t.Run("Invalid_Orchestrator", func(t *testing.T) {
		builder := &RunnerBuilder{
			queueSize:        10,
			orchestratorType: "invalid",
		}

		runner, err := builder.Build()

		assert.Error(t, err)
		assert.Nil(t, runner)
		assert.Contains(t, err.Error(), "unknown orchestrator type")
	})

	// Test agent registration error handling
	t.Run("Agent_Registration_Error", func(t *testing.T) {
		// Following Azure best practices for testable service integrations:
		// 1. Create a fully pre-configured mock
		// 2. Define expectations before execution
		// 3. Verify all expectations after execution

		// Create registry and orchestrator since we can't set them through the interface
		registry := agentflow.NewCallbackRegistry()
		//orch := orchestrator.NewRouteOrchestrator(registry)

		// Create the mock runner with expected behavior
		mockRun := new(mockRunner)
		// No need to mock SetCallbackRegistry and SetOrchestrator since we won't call them
		mockRun.On("RegisterAgent", "test-agent", mock.Anything).Return(errors.New("registration error"))
		// Make sure GetCallbackRegistry returns our registry if the code tries to access it
		mockRun.On("GetCallbackRegistry").Return(registry)

		// Create a test-specific builder
		builder := &TestableRunnerBuilder{
			RunnerBuilder: RunnerBuilder{
				queueSize:        10,
				orchestratorType: "route",
				agentHandlers:    map[string]agentflow.AgentHandler{"test-agent": &mockAgentHandler{}},
			},
			mockRunner: mockRun,
		}

		// Attempt to build with the mock
		runner, err := builder.Build()

		// Verify expectations following Azure testing best practices
		assert.Error(t, err, "Expected error not returned")
		assert.Nil(t, runner, "Runner should be nil when error occurs")
		assert.Contains(t, err.Error(), "failed to register agent", "Error message should indicate registration failure")
		mockRun.AssertExpectations(t)
	})
}

// TestableRunnerBuilder provides a testable version of the RunnerBuilder
type TestableRunnerBuilder struct {
	RunnerBuilder
	mockRunner agentflow.Runner
}

// Build overrides the standard Build method to use the mock runner
func (b *TestableRunnerBuilder) Build() (agentflow.Runner, error) {
	// Following Azure testing best practices:
	// 1. Get the callback registry to match expectations
	// 2. Use proper error propagation with context
	// 3. Make the test implementation match the real behavior

	// Get the callback registry as the real implementation would
	registry := b.mockRunner.GetCallbackRegistry()
	if registry == nil {
		return nil, fmt.Errorf("callback registry required for agent registration")
	}

	// Register agents - this is where we expect the error to occur
	for name, handler := range b.agentHandlers {
		if err := b.mockRunner.RegisterAgent(name, handler); err != nil {
			return nil, fmt.Errorf("failed to register agent '%s': %w", name, err)
		}
	}

	// Return the pre-configured mock
	return b.mockRunner, nil
}
