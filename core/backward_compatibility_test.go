package core

import (
	"context"
	"testing"
)

// TestBackwardCompatibility verifies that all essential public APIs work correctly
func TestBackwardCompatibility(t *testing.T) {
	t.Run("LoadConfig", func(t *testing.T) {
		// Test loading default config
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

	t.Run("LLMProviders", func(t *testing.T) {
		// Test OpenAI adapter creation (without actual API call)
		_, err := NewOpenAIAdapter("test-key", "gpt-4", 100, 0.7)
		if err != nil {
			t.Errorf("NewOpenAIAdapter failed: %v", err)
		}

		// Test Ollama adapter creation
		_, err = NewOllamaAdapter("http://localhost:11434", "llama2", 100, 0.7)
		if err != nil {
			t.Errorf("NewOllamaAdapter failed: %v", err)
		}

		// Test Azure adapter creation
		_, err = NewAzureOpenAIAdapter(AzureOpenAIAdapterOptions{
			Endpoint:            "https://test.openai.azure.com/",
			APIKey:              "test-key",
			ChatDeployment:      "gpt-4",
			EmbeddingDeployment: "text-embedding-ada-002",
		})
		if err != nil {
			t.Errorf("NewAzureOpenAIAdapter failed: %v", err)
		}
	})

	t.Run("Memory", func(t *testing.T) {
		// Test memory creation
		_, err := NewMemory(AgentMemoryConfig{
			Provider:   "inmemory",
			Connection: "",
		})
		if err != nil {
			t.Errorf("NewMemory failed: %v", err)
		}
	})

	t.Run("ConfigFromConfig", func(t *testing.T) {
		// Test LLM provider from config
		_, err := NewModelProviderFromConfig(LLMProviderConfig{
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

	t.Run("HelperFunctions", func(t *testing.T) {
		// Test helper functions
		temp := FloatPtr(0.7)
		if *temp != 0.7 {
			t.Error("FloatPtr helper function failed")
		}

		tokens := Int32Ptr(100)
		if *tokens != 100 {
			t.Error("Int32Ptr helper function failed")
		}
	})

	t.Run("TypesAndInterfaces", func(t *testing.T) {
		// Test that we can create core types
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

		// Test Response type
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
	})
}

// TestInterfaceCompatibility verifies that interfaces can be implemented
func TestInterfaceCompatibility(t *testing.T) {
	t.Run("AgentInterface", func(t *testing.T) {
		// Test that we can implement the Agent interface
		var _ Agent = &testAgent{}
	})

	t.Run("AgentHandlerInterface", func(t *testing.T) {
		// Test that we can implement the AgentHandler interface
		var _ AgentHandler = &testAgentHandler{}
	})

	t.Run("ModelProviderInterface", func(t *testing.T) {
		// Test that we can implement the ModelProvider interface
		var _ ModelProvider = &testModelProvider{}
	})

	t.Run("LLMAdapterInterface", func(t *testing.T) {
		// Test that we can implement the LLMAdapter interface
		var _ LLMAdapter = &testLLMAdapter{}
	})
}

// Test implementations to verify interfaces
type testAgent struct{}

func (a *testAgent) Run(ctx context.Context, inputState State) (State, error) {
	return inputState, nil
}

func (a *testAgent) Name() string {
	return "test-agent"
}

type testAgentHandler struct{}

func (h *testAgentHandler) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
	return AgentResult{}, nil
}

type testModelProvider struct{}

func (p *testModelProvider) Call(ctx context.Context, prompt Prompt) (Response, error) {
	return Response{Content: "test response"}, nil
}

func (p *testModelProvider) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
	ch := make(chan Token)
	close(ch)
	return ch, nil
}

func (p *testModelProvider) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return [][]float64{{0.1, 0.2, 0.3}}, nil
}

type testLLMAdapter struct{}

func (a *testLLMAdapter) Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	return "test completion", nil
}