package llm

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewOpenRouterAdapter(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      string
		model       string
		baseURL     string
		maxTokens   int
		temperature float32
		siteURL     string
		siteName    string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid configuration",
			apiKey:      "test-key",
			model:       "openai/gpt-3.5-turbo",
			baseURL:     "https://openrouter.ai/api/v1",
			maxTokens:   2000,
			temperature: 0.7,
			siteURL:     "https://myapp.com",
			siteName:    "My App",
			wantErr:     false,
		},
		{
			name:        "missing api key",
			apiKey:      "",
			model:       "openai/gpt-3.5-turbo",
			baseURL:     "https://openrouter.ai/api/v1",
			maxTokens:   2000,
			temperature: 0.7,
			wantErr:     true,
			errContains: "API key cannot be empty",
		},
		{
			name:        "defaults applied",
			apiKey:      "test-key",
			model:       "",
			baseURL:     "",
			maxTokens:   0,
			temperature: 0,
			wantErr:     false,
		},
		{
			name:        "optional site fields empty",
			apiKey:      "test-key",
			model:       "anthropic/claude-3-sonnet",
			baseURL:     "https://openrouter.ai/api/v1",
			maxTokens:   1500,
			temperature: 0.5,
			siteURL:     "",
			siteName:    "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := NewOpenRouterAdapter(
				tt.apiKey,
				tt.model,
				tt.baseURL,
				tt.maxTokens,
				tt.temperature,
				tt.siteURL,
				tt.siteName,
			)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewOpenRouterAdapter() expected error but got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("NewOpenRouterAdapter() error = %v, should contain %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewOpenRouterAdapter() unexpected error = %v", err)
				return
			}

			if adapter == nil {
				t.Error("NewOpenRouterAdapter() returned nil adapter")
				return
			}

			// Verify defaults were applied
			if tt.model == "" && adapter.model != "openai/gpt-3.5-turbo" {
				t.Errorf("Expected default model 'openai/gpt-3.5-turbo', got %v", adapter.model)
			}
			if tt.baseURL == "" && adapter.baseURL != "https://openrouter.ai/api/v1" {
				t.Errorf("Expected default baseURL 'https://openrouter.ai/api/v1', got %v", adapter.baseURL)
			}
			if tt.maxTokens == 0 && adapter.maxTokens != 2000 {
				t.Errorf("Expected default maxTokens 2000, got %v", adapter.maxTokens)
			}
			if tt.temperature == 0 && adapter.temperature != 0.7 {
				t.Errorf("Expected default temperature 0.7, got %v", adapter.temperature)
			}
		})
	}
}

func TestOpenRouterAdapter_Call(t *testing.T) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		t.Skip("OPENROUTER_API_KEY not set, skipping integration test")
	}

	adapter, err := NewOpenRouterAdapter(
		apiKey,
		"openai/gpt-3.5-turbo",
		"https://openrouter.ai/api/v1",
		150,
		0.7,
		"https://github.com/kunalkushwaha/agenticgokit",
		"AgenticGoKit",
	)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := adapter.Call(ctx, Prompt{
		System: "You are a helpful assistant.",
		User:   "Say 'Hello, World!' and nothing else.",
	})

	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	if resp.Content == "" {
		t.Error("Expected non-empty response content")
	}

	if resp.Usage.TotalTokens == 0 {
		t.Error("Expected non-zero token usage")
	}

	t.Logf("Response: %s", resp.Content)
	t.Logf("Tokens: %d prompt, %d completion, %d total",
		resp.Usage.PromptTokens,
		resp.Usage.CompletionTokens,
		resp.Usage.TotalTokens)
}

func TestOpenRouterAdapter_Call_WithParameters(t *testing.T) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		t.Skip("OPENROUTER_API_KEY not set, skipping integration test")
	}

	adapter, err := NewOpenRouterAdapter(
		apiKey,
		"openai/gpt-3.5-turbo",
		"https://openrouter.ai/api/v1",
		150,
		0.7,
		"",
		"",
	)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Override parameters
	maxTokens := int32(50)
	temperature := float32(0.3)

	resp, err := adapter.Call(ctx, Prompt{
		User: "Count from 1 to 3.",
		Parameters: ModelParameters{
			MaxTokens:   &maxTokens,
			Temperature: &temperature,
		},
	})

	if err != nil {
		t.Fatalf("Call with parameters failed: %v", err)
	}

	if resp.Content == "" {
		t.Error("Expected non-empty response content")
	}

	t.Logf("Response with params: %s", resp.Content)
}

func TestOpenRouterAdapter_Stream(t *testing.T) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		t.Skip("OPENROUTER_API_KEY not set, skipping integration test")
	}

	adapter, err := NewOpenRouterAdapter(
		apiKey,
		"openai/gpt-3.5-turbo",
		"https://openrouter.ai/api/v1",
		150,
		0.7,
		"",
		"",
	)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tokenChan, err := adapter.Stream(ctx, Prompt{
		User: "Count from 1 to 5.",
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
	t.Logf("Streamed response (%d tokens): %s", len(tokens), fullResponse)
}

func TestOpenRouterAdapter_Stream_ContextCancellation(t *testing.T) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		t.Skip("OPENROUTER_API_KEY not set, skipping integration test")
	}

	adapter, err := NewOpenRouterAdapter(
		apiKey,
		"openai/gpt-3.5-turbo",
		"https://openrouter.ai/api/v1",
		1000,
		0.7,
		"",
		"",
	)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tokenChan, err := adapter.Stream(ctx, Prompt{
		User: "Write a long story about a robot.",
	})

	if err != nil {
		t.Fatalf("Stream failed: %v", err)
	}

	// Cancel after receiving first few tokens
	tokenCount := 0
	for token := range tokenChan {
		if token.Error != nil {
			// Expected to get context cancellation error
			if ctx.Err() != context.Canceled {
				t.Errorf("Expected context.Canceled error, got: %v", token.Error)
			}
			return
		}
		tokenCount++
		if tokenCount >= 3 {
			cancel()
		}
	}

	if tokenCount < 3 {
		t.Logf("Received %d tokens before cancellation", tokenCount)
	}
}

func TestOpenRouterAdapter_Embeddings(t *testing.T) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		t.Skip("OPENROUTER_API_KEY not set, skipping integration test")
	}

	adapter, err := NewOpenRouterAdapter(
		apiKey,
		"openai/gpt-3.5-turbo",
		"https://openrouter.ai/api/v1",
		150,
		0.7,
		"",
		"",
	)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ctx := context.Background()

	_, err = adapter.Embeddings(ctx, []string{"test"})

	// Embeddings should return an error indicating not supported
	if err == nil {
		t.Error("Expected Embeddings to return error (not supported)")
	}

	if !strings.Contains(err.Error(), "not currently supported") {
		t.Errorf("Expected 'not supported' error, got: %v", err)
	}
}

func TestOpenRouterAdapter_CallWithEmptyPrompt(t *testing.T) {
	adapter, err := NewOpenRouterAdapter(
		"test-key",
		"openai/gpt-3.5-turbo",
		"https://openrouter.ai/api/v1",
		150,
		0.7,
		"",
		"",
	)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ctx := context.Background()

	_, err = adapter.Call(ctx, Prompt{
		System: "You are helpful",
		User:   "", // Empty user prompt
	})

	if err == nil {
		t.Error("Expected error for empty user prompt")
	}

	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("Expected 'cannot be empty' error, got: %v", err)
	}
}

func TestBuildOpenRouterMessages(t *testing.T) {
	tests := []struct {
		name      string
		prompt    Prompt
		wantLen   int
		hasUser   bool
		hasSystem bool
	}{
		{
			name: "both system and user",
			prompt: Prompt{
				System: "You are helpful",
				User:   "Hello",
			},
			wantLen:   2,
			hasUser:   true,
			hasSystem: true,
		},
		{
			name: "user only",
			prompt: Prompt{
				User: "Hello",
			},
			wantLen:   1,
			hasUser:   true,
			hasSystem: false,
		},
		{
			name: "system only",
			prompt: Prompt{
				System: "You are helpful",
			},
			wantLen:   1,
			hasUser:   false,
			hasSystem: true,
		},
		{
			name:      "empty prompt",
			prompt:    Prompt{},
			wantLen:   0,
			hasUser:   false,
			hasSystem: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages := buildOpenRouterMessages(tt.prompt)

			if len(messages) != tt.wantLen {
				t.Errorf("Expected %d messages, got %d", tt.wantLen, len(messages))
			}

			hasSystem := false
			hasUser := false
			for _, msg := range messages {
				if role, ok := msg["role"].(string); ok {
					if role == "system" {
						hasSystem = true
					}
					if role == "user" {
						hasUser = true
					}
				}
			}

			if hasSystem != tt.hasSystem {
				t.Errorf("Expected hasSystem=%v, got %v", tt.hasSystem, hasSystem)
			}
			if hasUser != tt.hasUser {
				t.Errorf("Expected hasUser=%v, got %v", tt.hasUser, hasUser)
			}
		})
	}
}
