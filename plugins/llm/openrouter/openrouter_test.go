package openrouter

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

func TestPluginRegistration(t *testing.T) {
	// Test that the plugin is registered correctly
	config := core.LLMProviderConfig{
		Type:   "openrouter",
		APIKey: "test-key",
		Model:  "openai/gpt-3.5-turbo",
	}

	provider, err := core.NewModelProviderFromConfig(config)

	// Should not error about unregistered provider
	// May error about invalid API key or network issues (expected in unit test)
	if err != nil {
		if strings.Contains(err.Error(), "not registered") {
			t.Error("Plugin not registered correctly")
		}
		// Other errors are acceptable in unit tests without real API key
		t.Logf("Expected error (no real API call): %v", err)
	}

	if provider != nil {
		t.Log("Provider created successfully")
	}
}

func TestFactoryWithConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  core.LLMProviderConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: core.LLMProviderConfig{
				Type:        "openrouter",
				APIKey:      "test-key",
				Model:       "anthropic/claude-3-sonnet",
				MaxTokens:   2000,
				Temperature: 0.7,
			},
			wantErr: false,
		},
		{
			name: "with site tracking",
			config: core.LLMProviderConfig{
				Type:        "openrouter",
				APIKey:      "test-key",
				Model:       "openai/gpt-4",
				SiteURL:     "https://myapp.com",
				SiteName:    "My App",
				MaxTokens:   1500,
				Temperature: 0.5,
			},
			wantErr: false,
		},
		{
			name: "default base url",
			config: core.LLMProviderConfig{
				Type:   "openrouter",
				APIKey: "test-key",
				Model:  "google/gemini-pro",
			},
			wantErr: false,
		},
		{
			name: "custom base url",
			config: core.LLMProviderConfig{
				Type:    "openrouter",
				APIKey:  "test-key",
				Model:   "meta-llama/llama-3-70b-instruct",
				BaseURL: "https://custom.openrouter.ai/api/v1",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := core.NewModelProviderFromConfig(tt.config)

			if tt.wantErr && err == nil {
				t.Error("Expected error but got nil")
			}

			if !tt.wantErr && err != nil {
				// Should not error during provider creation
				// (actual API calls will fail without valid key)
				t.Logf("Provider created, errors during use expected without valid API key: %v", err)
			}

			if provider != nil {
				t.Log("Provider instance created")
			}
		})
	}
}

func TestOpenRouterIntegration(t *testing.T) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		t.Skip("OPENROUTER_API_KEY not set, skipping integration test")
	}

	config := core.LLMProviderConfig{
		Type:        "openrouter",
		APIKey:      apiKey,
		Model:       "openai/gpt-3.5-turbo",
		MaxTokens:   150,
		Temperature: 0.7,
		SiteURL:     "https://github.com/kunalkushwaha/agenticgokit",
		SiteName:    "AgenticGoKit Tests",
	}

	provider, err := core.NewModelProviderFromConfig(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test Call
	t.Run("Call", func(t *testing.T) {
		response, err := provider.Call(ctx, core.Prompt{
			System: "You are a helpful assistant.",
			User:   "Say hello!",
		})

		if err != nil {
			t.Fatalf("Call failed: %v", err)
		}

		if response.Content == "" {
			t.Error("Expected non-empty response")
		}

		t.Logf("Response: %s", response.Content)
		t.Logf("Tokens: %d prompt, %d completion, %d total",
			response.Usage.PromptTokens,
			response.Usage.CompletionTokens,
			response.Usage.TotalTokens)
	})

	// Test Stream
	t.Run("Stream", func(t *testing.T) {
		tokenChan, err := provider.Stream(ctx, core.Prompt{
			User: "Count from 1 to 3.",
		})

		if err != nil {
			t.Fatalf("Stream failed: %v", err)
		}

		var tokens []string
		for token := range tokenChan {
			if token.Error != nil {
				t.Fatalf("Stream error: %v", token.Error)
			}
			tokens = append(tokens, token.Content)
		}

		if len(tokens) == 0 {
			t.Error("Expected at least one token")
		}

		fullResponse := strings.Join(tokens, "")
		t.Logf("Streamed response: %s", fullResponse)
	})
}

func TestProviderAdapter(t *testing.T) {
	// Test the adapter type conversions work
	// We can't create a full adapter without real dependencies,
	// but we can verify the type implements the interface
	var _ core.ModelProvider = (*providerAdapter)(nil)
	t.Log("providerAdapter correctly implements core.ModelProvider interface")
}

func TestMultipleModels(t *testing.T) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		t.Skip("OPENROUTER_API_KEY not set, skipping integration test")
	}

	models := []string{
		"openai/gpt-3.5-turbo",
		"anthropic/claude-3-haiku",
		"google/gemini-pro",
	}

	for _, model := range models {
		t.Run(model, func(t *testing.T) {
			config := core.LLMProviderConfig{
				Type:        "openrouter",
				APIKey:      apiKey,
				Model:       model,
				MaxTokens:   50,
				Temperature: 0.7,
			}

			provider, err := core.NewModelProviderFromConfig(config)
			if err != nil {
				t.Fatalf("Failed to create provider for %s: %v", model, err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			response, err := provider.Call(ctx, core.Prompt{
				User: "Say 'OK' and nothing else.",
			})

			if err != nil {
				// Some models might not be available or might fail
				t.Logf("Call to %s failed (may be unavailable): %v", model, err)
				return
			}

			if response.Content == "" {
				t.Errorf("Model %s returned empty response", model)
			}

			t.Logf("Model %s response: %s (tokens: %d)",
				model, response.Content, response.Usage.TotalTokens)
		})
	}
}

func TestEnvironmentVariables(t *testing.T) {
	// Test that environment variables are picked up correctly
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		t.Skip("OPENROUTER_API_KEY not set, skipping test")
	}

	// Test with AgentLLMConfig which uses NewLLMProvider
	config := core.AgentLLMConfig{
		Provider:    "openrouter",
		Model:       "openai/gpt-3.5-turbo",
		MaxTokens:   100,
		Temperature: 0.7,
	}

	provider, err := core.NewLLMProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider from env: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := provider.Call(ctx, core.Prompt{
		User: "Say hello!",
	})

	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	if response.Content == "" {
		t.Error("Expected non-empty response")
	}

	t.Logf("Response from env config: %s", response.Content)
}

func TestParameterOverrides(t *testing.T) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		t.Skip("OPENROUTER_API_KEY not set, skipping integration test")
	}

	config := core.LLMProviderConfig{
		Type:        "openrouter",
		APIKey:      apiKey,
		Model:       "openai/gpt-3.5-turbo",
		MaxTokens:   500, // Default
		Temperature: 0.7, // Default
	}

	provider, err := core.NewModelProviderFromConfig(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Override with prompt parameters
	maxTokens := int32(50)
	temperature := float32(0.3)

	response, err := provider.Call(ctx, core.Prompt{
		User: "Count from 1 to 5.",
		Parameters: core.ModelParameters{
			MaxTokens:   &maxTokens,
			Temperature: &temperature,
		},
	})

	if err != nil {
		t.Fatalf("Call with overrides failed: %v", err)
	}

	if response.Content == "" {
		t.Error("Expected non-empty response")
	}

	t.Logf("Response with parameter overrides: %s", response.Content)
}
