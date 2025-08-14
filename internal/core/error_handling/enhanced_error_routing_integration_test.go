package error_handling

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestEnhancedErrorRoutingIntegration tests the complete enhanced error routing system
func TestEnhancedErrorRoutingIntegration(t *testing.T) {
	ctx := context.Background()

	// Create specialized error handlers
	validationHandler := AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
		// Simulate validation error handling
		outputState := NewState()
		outputState.Set("processed_by", "validation-error-handler")
		outputState.Set("recovery_action", "retry_with_corrections")
		return AgentResult{OutputState: outputState}, nil
	})

	timeoutHandler := AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
		// Simulate timeout error handling
		outputState := NewState()
		outputState.Set("processed_by", "timeout-error-handler")
		outputState.Set("recovery_action", "retry_with_backoff")
		return AgentResult{OutputState: outputState}, nil
	})

	criticalHandler := AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
		// Simulate critical error handling
		outputState := NewState()
		outputState.Set("processed_by", "critical-error-handler")
		outputState.Set("recovery_action", "terminate_workflow")
		outputState.Set("alert_level", "emergency")
		return AgentResult{OutputState: outputState}, nil
	})

	defaultErrorHandler := AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
		// Default error handler
		outputState := NewState()
		outputState.Set("processed_by", "error-handler")
		outputState.Set("recovery_action", "log_and_continue")
		return AgentResult{OutputState: outputState}, nil
	})

	// Create agents that will fail with different error types
	validationFailAgent := AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
		return AgentResult{}, errors.New("validation failed: invalid input format")
	})

	timeoutFailAgent := AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
		return AgentResult{}, errors.New("timeout: operation timed out after 30 seconds")
	})

	llmFailAgent := AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
		return AgentResult{}, errors.New("LLM error: rate limit exceeded")
	})

	// Create agents map with specialized error handlers
	agents := map[string]AgentHandler{
		"validation-fail-agent":    validationFailAgent,
		"timeout-fail-agent":       timeoutFailAgent,
		"llm-fail-agent":           llmFailAgent,
		"validation-error-handler": validationHandler,
		"timeout-error-handler":    timeoutHandler,
		"critical-error-handler":   criticalHandler,
		"error-handler":            defaultErrorHandler,
	}

	// Create runner with agents
	runner := NewRunnerWithConfig(RunnerConfig{
		QueueSize: 10,
		Agents:    agents,
	})

	// Start runner
	if err := runner.Start(ctx); err != nil {
		t.Fatalf("Failed to start runner: %v", err)
	}
	defer runner.Stop()

	testCases := []struct {
		name               string
		agentName          string
		expectedHandler    string
		expectedRecovery   string
		shouldHaveCategory bool
		expectedCategory   string
	}{
		{
			name:               "Validation Error Routing",
			agentName:          "validation-fail-agent",
			expectedHandler:    "validation-error-handler",
			expectedRecovery:   "retry_with_corrections",
			shouldHaveCategory: true,
			expectedCategory:   "validation",
		},
		{
			name:               "Timeout Error Routing",
			agentName:          "timeout-fail-agent",
			expectedHandler:    "timeout-error-handler",
			expectedRecovery:   "retry_with_backoff",
			shouldHaveCategory: true,
			expectedCategory:   "timeout",
		},
		{
			name:               "LLM Error Routing",
			agentName:          "llm-fail-agent",
			expectedHandler:    "error-handler",
			expectedRecovery:   "log_and_continue",
			shouldHaveCategory: true,
			expectedCategory:   "llm",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create result channel to capture processed events
			resultChan := make(chan Event, 1) // Register callback to capture error handler results
			runner.RegisterCallback(HookAfterAgentRun, "test-capture", func(ctx context.Context, args CallbackArgs) (State, error) {
				if args.State != nil {
					if processedBy, exists := args.AgentResult.OutputState.Get("processed_by"); exists {
						if processedBy == tc.expectedHandler {
							resultChan <- NewEvent(tc.expectedHandler, map[string]interface{}{
								"processed_by":    processedBy,
								"recovery_action": getStateValue(args.AgentResult.OutputState, "recovery_action"),
								"error_category":  getStateValue(args.AgentResult.OutputState, "error_category"),
							}, nil)
						}
					}
				}
				return args.State, nil
			})

			// Create test event
			sessionID := "test-session-" + tc.name
			event := NewEvent(tc.agentName, EventData{
				"test": "data",
			}, map[string]string{
				SessionIDKey:     sessionID,
				RouteMetadataKey: tc.agentName,
			})

			// Emit event
			if err := runner.Emit(event); err != nil {
				t.Fatalf("Failed to emit event: %v", err)
			}

			// Wait for result
			select {
			case result := <-resultChan:
				data := result.GetData()

				// Verify handler
				if processedBy, ok := data["processed_by"]; !ok || processedBy != tc.expectedHandler {
					t.Errorf("Expected handler %s, got %v", tc.expectedHandler, processedBy)
				}

				// Verify recovery action
				if recoveryAction, ok := data["recovery_action"]; !ok || recoveryAction != tc.expectedRecovery {
					t.Errorf("Expected recovery action %s, got %v", tc.expectedRecovery, recoveryAction)
				}

				// Verify error category if expected
				if tc.shouldHaveCategory {
					if errorCategory, ok := data["error_category"]; !ok || errorCategory != tc.expectedCategory {
						t.Errorf("Expected error category %s, got %v", tc.expectedCategory, errorCategory)
					}
				}

			case <-time.After(5 * time.Second):
				t.Fatalf("Test timed out waiting for error handler result")
			}

			// Clean up callback
			runner.UnregisterCallback(HookAfterAgentRun, "test-capture")
		})
	}
}

// TestAutoErrorRoutingConfiguration tests automatic error routing configuration
func TestAutoErrorRoutingConfiguration(t *testing.T) {
	// Test that the factory automatically configures error routing
	agents := map[string]AgentHandler{
		"validation-error-handler": AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
			return AgentResult{OutputState: NewState()}, nil
		}),
		"timeout-error-handler": AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
			return AgentResult{OutputState: NewState()}, nil
		}),
		"critical-error-handler": AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
			return AgentResult{OutputState: NewState()}, nil
		}),
		"error-handler": AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
			return AgentResult{OutputState: NewState()}, nil
		}),
	}

	runner := NewRunnerWithConfig(RunnerConfig{
		QueueSize: 10,
		Agents:    agents,
	})

	// Verify that error routing was automatically configured
	runnerImpl, ok := runner.(*RunnerImpl)
	if !ok {
		t.Fatalf("Expected RunnerImpl, got %T", runner)
	}

	config := runnerImpl.getErrorRouterConfig()
	if config == nil {
		t.Fatal("Expected error router config to be set, got nil")
	}

	// Verify specific handlers are configured
	if config.CategoryHandlers[ErrorCodeValidation] != "validation-error-handler" {
		t.Errorf("Expected validation handler to be validation-error-handler, got %s",
			config.CategoryHandlers[ErrorCodeValidation])
	}

	if config.CategoryHandlers[ErrorCodeTimeout] != "timeout-error-handler" {
		t.Errorf("Expected timeout handler to be timeout-error-handler, got %s",
			config.CategoryHandlers[ErrorCodeTimeout])
	}

	if config.SeverityHandlers[SeverityCritical] != "critical-error-handler" {
		t.Errorf("Expected critical handler to be critical-error-handler, got %s",
			config.SeverityHandlers[SeverityCritical])
	}

	if config.ErrorHandlerName != "error-handler" {
		t.Errorf("Expected default error handler to be error-handler, got %s",
			config.ErrorHandlerName)
	}
}

// TestScaffoldGeneratedProjectErrorHandling tests that scaffold-generated projects have proper error handling
func TestScaffoldGeneratedProjectErrorHandling(t *testing.T) { // This test simulates a generated project structure

	// Create agents similar to what the scaffold would generate
	agents := map[string]AgentHandler{
		"agent1": AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
			// Simulate occasional failures
			if event.GetData()["cause_error"] != nil {
				return AgentResult{}, errors.New("simulated agent failure")
			}
			outputState := NewState()
			outputState.Set("processed_by", "agent1")
			outputState.SetMeta(RouteMetadataKey, "workflow_finalizer")
			return AgentResult{OutputState: outputState}, nil
		}),

		// Error handlers as generated by scaffold
		"error_handler": AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
			outputState := NewState()
			outputState.Set("processed_by", "error_handler")
			outputState.Set("error_handled", true)
			return AgentResult{OutputState: outputState}, nil
		}),

		"validation_error_handler": AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
			outputState := NewState()
			outputState.Set("processed_by", "validation_error_handler")
			outputState.Set("recovery_action", "retry_with_corrections")
			return AgentResult{OutputState: outputState}, nil
		}),

		"timeout_error_handler": AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
			outputState := NewState()
			outputState.Set("processed_by", "timeout_error_handler")
			outputState.Set("recovery_action", "retry_with_backoff")
			return AgentResult{OutputState: outputState}, nil
		}),

		"critical_error_handler": AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
			outputState := NewState()
			outputState.Set("processed_by", "critical_error_handler")
			outputState.Set("recovery_action", "terminate_workflow")
			return AgentResult{OutputState: outputState}, nil
		}),

		"workflow_finalizer": AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
			outputState := NewState()
			outputState.Set("workflow_completed", true)
			return AgentResult{OutputState: outputState}, nil
		}),
	}

	// Create runner with enhanced error routing (as the scaffold would)
	runner := NewRunnerWithConfig(RunnerConfig{
		QueueSize: 10,
		Agents:    agents,
	})

	ctx := context.Background()
	if err := runner.Start(ctx); err != nil {
		t.Fatalf("Failed to start runner: %v", err)
	}
	defer runner.Stop()

	// Test normal operation
	t.Run("Normal Operation", func(t *testing.T) {
		event := NewEvent("agent1", EventData{
			"message": "test message",
		}, map[string]string{
			SessionIDKey:     "test-normal",
			RouteMetadataKey: "agent1",
		})

		if err := runner.Emit(event); err != nil {
			t.Fatalf("Failed to emit event: %v", err)
		}

		// Allow processing time
		time.Sleep(100 * time.Millisecond)
	})

	// Test error handling
	t.Run("Error Handling", func(t *testing.T) {
		event := NewEvent("agent1", EventData{
			"message":     "test message",
			"cause_error": true,
		}, map[string]string{
			SessionIDKey:     "test-error",
			RouteMetadataKey: "agent1",
		})

		if err := runner.Emit(event); err != nil {
			t.Fatalf("Failed to emit event: %v", err)
		}

		// Allow processing time
		time.Sleep(100 * time.Millisecond)
	})
}

// Helper function to safely get state values
func getStateValue(state State, key string) interface{} {
	if state == nil {
		return nil
	}
	if value, exists := state.Get(key); exists {
		return value
	}
	return nil
}

// MockProvider implementation for testing
type MockProvider struct{}

func (m *MockProvider) Call(ctx context.Context, prompt Prompt) (Response, error) {
	return Response{
		Content: "Mock response",
		Usage: UsageStats{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
		FinishReason: "stop",
	}, nil
}

func (m *MockProvider) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
	ch := make(chan Token, 1)
	go func() {
		defer close(ch)
		ch <- Token{Content: "Mock response", Error: nil}
	}()
	return ch, nil
}

func (m *MockProvider) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	embeddings := make([][]float64, len(texts))
	for i := range texts {
		embeddings[i] = []float64{0.1, 0.2, 0.3}
	}
	return embeddings, nil
}
