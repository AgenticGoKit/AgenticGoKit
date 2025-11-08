package llm

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewHuggingFaceAdapter(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      string
		model       string
		baseURL     string
		apiType     HFAPIType
		maxTokens   int
		temperature float32
		options     HFAdapterOptions
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid inference API configuration",
			apiKey:      "test-key",
			model:       "gpt2",
			baseURL:     "",
			apiType:     HFAPITypeInference,
			maxTokens:   100,
			temperature: 0.7,
			options:     HFAdapterOptions{},
			wantErr:     false,
		},
		{
			name:        "missing api key for inference API",
			apiKey:      "",
			model:       "gpt2",
			baseURL:     "",
			apiType:     HFAPITypeInference,
			maxTokens:   100,
			temperature: 0.7,
			options:     HFAdapterOptions{},
			wantErr:     true,
			errContains: "API key is required",
		},
		{
			name:        "valid endpoint configuration",
			apiKey:      "test-key",
			model:       "gpt2",
			baseURL:     "https://my-endpoint.endpoints.huggingface.cloud",
			apiType:     HFAPITypeEndpoint,
			maxTokens:   150,
			temperature: 0.8,
			options:     HFAdapterOptions{WaitForModel: true},
			wantErr:     false,
		},
		{
			name:        "endpoint without base URL",
			apiKey:      "test-key",
			model:       "gpt2",
			baseURL:     "",
			apiType:     HFAPITypeEndpoint,
			maxTokens:   100,
			temperature: 0.7,
			options:     HFAdapterOptions{},
			wantErr:     true,
			errContains: "base URL is required",
		},
		{
			name:        "valid TGI configuration",
			apiKey:      "",
			model:       "gpt2",
			baseURL:     "http://localhost:8080",
			apiType:     HFAPITypeTGI,
			maxTokens:   200,
			temperature: 0.9,
			options:     HFAdapterOptions{TopP: 0.95, TopK: 50},
			wantErr:     false,
		},
		{
			name:        "valid chat API configuration",
			apiKey:      "test-key",
			model:       "meta-llama/Llama-2-7b-chat-hf",
			baseURL:     "https://api-inference.huggingface.co",
			apiType:     HFAPITypeChat,
			maxTokens:   150,
			temperature: 0.7,
			options:     HFAdapterOptions{},
			wantErr:     false,
		},
		{
			name:        "defaults applied",
			apiKey:      "test-key",
			model:       "",
			baseURL:     "",
			apiType:     "",
			maxTokens:   0,
			temperature: 0,
			options:     HFAdapterOptions{},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := NewHuggingFaceAdapter(
				tt.apiKey,
				tt.model,
				tt.baseURL,
				tt.apiType,
				tt.maxTokens,
				tt.temperature,
				tt.options,
			)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewHuggingFaceAdapter() expected error but got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("NewHuggingFaceAdapter() error = %v, should contain %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewHuggingFaceAdapter() unexpected error = %v", err)
				return
			}

			if adapter == nil {
				t.Error("NewHuggingFaceAdapter() returned nil adapter")
				return
			}

			// Verify defaults were applied
			if tt.model == "" && adapter.model != "gpt2" {
				t.Errorf("NewHuggingFaceAdapter() default model = %v, want gpt2", adapter.model)
			}
			if tt.apiType == "" && adapter.apiType != HFAPITypeInference {
				t.Errorf("NewHuggingFaceAdapter() default apiType = %v, want inference", adapter.apiType)
			}
		})
	}
}

func TestHuggingFaceAdapter_Call_Integration(t *testing.T) {
	// Skip if no API key is set
	apiKey := getTestAPIKey(t, "HUGGINGFACE_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: HUGGINGFACE_API_KEY not set")
	}

	adapter, err := NewHuggingFaceAdapter(
		apiKey,
		"meta-llama/Llama-3.2-1B-Instruct", // Using a model available on the new router
		"",
		HFAPITypeInference,
		50,
		0.7,
		HFAdapterOptions{
			WaitForModel: true,
			UseCache:     true,
		},
	)
	if err != nil {
		t.Fatalf("NewHuggingFaceAdapter() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	prompt := Prompt{
		System: "You are a helpful assistant.",
		User:   "Say hello in one word.",
		Parameters: ModelParameters{
			Temperature: floatPtr(0.7),
			MaxTokens:   int32Ptr(50),
		},
	}

	resp, err := adapter.Call(ctx, prompt)
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}

	if resp.Content == "" {
		t.Error("Call() returned empty content")
	}

	t.Logf("Response: %s", resp.Content)
}

func TestHuggingFaceAdapter_Stream_Integration(t *testing.T) {
	// Skip if no API key is set
	apiKey := getTestAPIKey(t, "HUGGINGFACE_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: HUGGINGFACE_API_KEY not set")
	}

	adapter, err := NewHuggingFaceAdapter(
		apiKey,
		"meta-llama/Llama-3.2-1B-Instruct", // Using a model available on the new router
		"",
		HFAPITypeInference,
		50,
		0.7,
		HFAdapterOptions{
			WaitForModel: true,
		},
	)
	if err != nil {
		t.Fatalf("NewHuggingFaceAdapter() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	prompt := Prompt{
		System: "You are a helpful assistant.",
		User:   "Count from 1 to 5.",
		Parameters: ModelParameters{
			Temperature: floatPtr(0.7),
			MaxTokens:   int32Ptr(50),
		},
	}

	tokenChan, err := adapter.Stream(ctx, prompt)
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}

	var tokens []string
	for token := range tokenChan {
		if token.Error != nil {
			t.Fatalf("Stream() token error = %v", token.Error)
		}
		tokens = append(tokens, token.Content)
	}

	if len(tokens) == 0 {
		t.Error("Stream() returned no tokens")
	}

	fullResponse := strings.Join(tokens, "")
	t.Logf("Streamed response: %s", fullResponse)
}

func TestHuggingFaceAdapter_Embeddings_Integration(t *testing.T) {
	// Skip if no API key is set
	apiKey := getTestAPIKey(t, "HUGGINGFACE_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: HUGGINGFACE_API_KEY not set")
	}

	// Note: Embeddings may require HuggingFace Inference Endpoints or serverless inference.
	// The old api-inference.huggingface.co endpoint is deprecated.
	// For now, skip this test as the embeddings API structure is different
	t.Skip("Skipping embeddings test: Embeddings API requires separate endpoint configuration")

	adapter, err := NewHuggingFaceAdapter(
		apiKey,
		"sentence-transformers/all-MiniLM-L6-v2",
		"",
		HFAPITypeInference,
		0,
		0,
		HFAdapterOptions{},
	)
	if err != nil {
		t.Fatalf("NewHuggingFaceAdapter() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	texts := []string{"Hello world", "Goodbye world"}
	embeddings, err := adapter.Embeddings(ctx, texts)
	if err != nil {
		t.Fatalf("Embeddings() error = %v", err)
	}

	if len(embeddings) != len(texts) {
		t.Errorf("Embeddings() returned %d embeddings, want %d", len(embeddings), len(texts))
	}

	for i, emb := range embeddings {
		if len(emb) == 0 {
			t.Errorf("Embeddings() embedding %d is empty", i)
		}
		t.Logf("Embedding %d has %d dimensions", i, len(emb))
	}
}

func TestHuggingFaceAPITypes(t *testing.T) {
	apiTypes := []HFAPIType{
		HFAPITypeInference,
		HFAPITypeEndpoint,
		HFAPITypeTGI,
		HFAPITypeChat,
	}

	for _, apiType := range apiTypes {
		if string(apiType) == "" {
			t.Errorf("API type is empty: %v", apiType)
		}
	}
}

// Helper function to get test API key from environment
func getTestAPIKey(t *testing.T, envVar string) string {
	t.Helper()
	return strings.TrimSpace(os.Getenv(envVar))
}
