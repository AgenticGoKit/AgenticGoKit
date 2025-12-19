package llm

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
)

// --- Unit Tests ---

func TestNewAzureOpenAIAdapter(t *testing.T) {
	validOpts := AzureOpenAIAdapterOptions{
		Endpoint:            "https://valid.openai.azure.com",
		APIKey:              "valid-key",
		ChatDeployment:      "gpt-deploy",
		EmbeddingDeployment: "embed-deploy",
	}

	// Test valid configuration
	adapter, err := NewAzureOpenAIAdapter(validOpts)
	if err != nil {
		t.Fatalf("NewAzureOpenAIAdapter failed with valid options: %v", err)
	}
	if adapter == nil {
		t.Fatal("NewAzureOpenAIAdapter returned nil adapter with valid options")
	}
	// Corrected: Check httpClient instead of client
	if adapter.httpClient == nil {
		t.Error("Adapter httpClient should not be nil")
	}
	if adapter.chatDeployment != validOpts.ChatDeployment {
		t.Errorf("ChatDeployment mismatch: got %s, want %s", adapter.chatDeployment, validOpts.ChatDeployment)
	}
	if adapter.embeddingDeployment != validOpts.EmbeddingDeployment {
		t.Errorf("EmbeddingDeployment mismatch: got %s, want %s", adapter.embeddingDeployment, validOpts.EmbeddingDeployment)
	}

	// Test missing fields
	testCases := []struct {
		name      string
		opts      AzureOpenAIAdapterOptions
		expectErr string // Expect the generic error message
	}{
		{"MissingEndpoint", AzureOpenAIAdapterOptions{APIKey: "k", ChatDeployment: "c", EmbeddingDeployment: "e"}, "azure adapter requires endpoint, api key, chat deployment, and embedding deployment"},
		{"MissingAPIKey", AzureOpenAIAdapterOptions{Endpoint: "e", ChatDeployment: "c", EmbeddingDeployment: "e"}, "azure adapter requires endpoint, api key, chat deployment, and embedding deployment"},
		{"MissingChatDeployment", AzureOpenAIAdapterOptions{Endpoint: "e", APIKey: "k", EmbeddingDeployment: "e"}, "azure adapter requires endpoint, api key, chat deployment, and embedding deployment"},
		{"MissingEmbeddingDeployment", AzureOpenAIAdapterOptions{Endpoint: "e", APIKey: "k", ChatDeployment: "c"}, "azure adapter requires endpoint, api key, chat deployment, and embedding deployment"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewAzureOpenAIAdapter(tc.opts)
			if err == nil {
				t.Errorf("Expected error containing '%s', but got nil", tc.expectErr)
			} else if !strings.Contains(err.Error(), tc.expectErr) { // Check if the actual error contains the expected generic message
				t.Errorf("Expected error containing '%s', but got: %v", tc.expectErr, err)
			}
		})
	}

	// Test invalid endpoint (client creation error) - This might require a mock HTTP transport
	// or rely on the SDK's internal validation if it performs any upfront checks.
	// For now, we assume the SDK handles deeper validation during actual calls.
	// _, err = NewAzureOpenAIAdapter(AzureOpenAIAdapterOptions{Endpoint:"invalid-endpoint", APIKey:"k", ChatDeployment:"c", EmbeddingDeployment:"e"})
	// if err == nil {
	//     t.Error("Expected error for invalid endpoint format, but got nil")
	// }
}

// --- Integration Tests (Requires Environment Variables) ---

const (
	envAzureEndpoint            = "AZURE_OPENAI_ENDPOINT"
	envAzureAPIKey              = "AZURE_OPENAI_API_KEY"
	envAzureChatDeployment      = "AZURE_OPENAI_CHAT_DEPLOYMENT"
	envAzureEmbeddingDeployment = "AZURE_OPENAI_EMBEDDING_DEPLOYMENT"
)

// helper function to get integration test adapter or skip
func getIntegrationAdapter(t *testing.T) *AzureOpenAIAdapter {
	t.Helper()
	endpoint := os.Getenv(envAzureEndpoint)
	apiKey := os.Getenv(envAzureAPIKey)
	chatDeploy := os.Getenv(envAzureChatDeployment)
	embedDeploy := os.Getenv(envAzureEmbeddingDeployment)

	if endpoint == "" || apiKey == "" || chatDeploy == "" || embedDeploy == "" {
		t.Skipf("Skipping integration test: Set %s, %s, %s, and %s environment variables",
			envAzureEndpoint, envAzureAPIKey, envAzureChatDeployment, envAzureEmbeddingDeployment)
	}

	// Use the new constructor
	opts := AzureOpenAIAdapterOptions{
		Endpoint:            endpoint,
		APIKey:              apiKey,
		ChatDeployment:      chatDeploy,
		EmbeddingDeployment: embedDeploy,
		// Optionally add a custom HTTP client for testing if needed
		// HTTPClient: &http.Client{Timeout: 45 * time.Second},
	}

	adapter, err := NewAzureOpenAIAdapter(opts) // Call the updated constructor
	if err != nil {
		t.Fatalf("Failed to create adapter for integration test: %v", err)
	}
	return adapter
}

func TestAzureOpenAIAdapter_Call_Integration(t *testing.T) {
	adapter := getIntegrationAdapter(t)
	ctx := context.Background()

	prompt := Prompt{
		System: "You are a helpful AI assistant.",
		User:   "What is the capital of France?",
		Parameters: ModelParameters{
			Temperature: to.Ptr[float32](0.5),
			MaxTokens:   to.Ptr[int32](50),
		},
	}

	resp, err := adapter.Call(ctx, prompt)
	if err != nil {
		t.Fatalf("Call() failed: %v", err)
	}

	if resp.Content == "" {
		t.Error("Call() returned empty content")
	}
	if !strings.Contains(strings.ToLower(resp.Content), "paris") {
		t.Errorf("Call() response content unexpected: got '%s'", resp.Content)
	}
	if resp.FinishReason == "" {
		t.Error("Call() returned empty FinishReason")
	}
	// Usage stats might be zero depending on the model/Azure config, so check > 0 is less reliable
	t.Logf("Call() response: Content='%s', FinishReason='%s', Usage=%+v", resp.Content, resp.FinishReason, resp.Usage)
}

func TestAzureOpenAIAdapter_Stream_Integration(t *testing.T) {
	adapter := getIntegrationAdapter(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Add timeout
	defer cancel()

	prompt := Prompt{
		System: "You are a helpful AI assistant.",
		User:   "Write a short sentence about streaming.",
		Parameters: ModelParameters{
			Temperature: to.Ptr[float32](0.7),
			MaxTokens:   to.Ptr[int32](50),
		},
	}

	tokenChan, err := adapter.Stream(ctx, prompt)
	if err != nil {
		t.Fatalf("Stream() failed: %v", err)
	}

	receivedContent := ""
	tokenCount := 0
	var lastToken Token
	for token := range tokenChan {
		if token.Error != nil {
			t.Fatalf("Stream() received error in token: %v", token.Error)
		}
		if token.Content == "" && token.Error == nil {
			// Allow empty tokens, but log maybe? Some models might send them.
			// t.Logf("Received empty token")
		}
		receivedContent += token.Content
		tokenCount++
		lastToken = token
	}

	if tokenCount == 0 {
		t.Error("Stream() received no tokens")
	}
	if receivedContent == "" {
		t.Error("Stream() received tokens but combined content is empty")
	}
	// Check context error after loop in case of timeout
	if ctx.Err() != nil {
		t.Fatalf("Stream() context cancelled or timed out during consumption: %v", ctx.Err())
	}

	t.Logf("Stream() received %d tokens. Content: '%s'", tokenCount, receivedContent)
	// TODO: Check lastToken for finish reason / usage if the adapter is updated to include it.
	_ = lastToken // Use lastToken if needed later
}

func TestAzureOpenAIAdapter_Embeddings_Integration(t *testing.T) {
	adapter := getIntegrationAdapter(t)
	ctx := context.Background()

	texts := []string{"hello world", "agentflow test"}

	embeddings, err := adapter.Embeddings(ctx, texts)
	if err != nil {
		t.Fatalf("Embeddings() failed: %v", err)
	}

	if len(embeddings) != len(texts) {
		t.Fatalf("Embeddings() returned %d embeddings, want %d", len(embeddings), len(texts))
	}

	for i, emb := range embeddings {
		if len(emb) == 0 {
			t.Errorf("Embeddings() returned empty embedding for text index %d", i)
		}
		// Check a few values are not zero, assuming valid embeddings won't be all zeros
		nonZeroFound := false
		for j := 0; j < 10 && j < len(emb); j++ {
			if emb[j] != 0.0 {
				nonZeroFound = true
				break
			}
		}
		if !nonZeroFound && len(emb) > 0 {
			t.Errorf("Embeddings() returned potentially zero embedding for text index %d (first 10 values are 0)", i)
		}
		t.Logf("Embedding %d length: %d, first few values: %v...", i, len(emb), emb[:min(5, len(emb))])
	}
}

// min helper for logging
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Multimodal Message Building Tests

func TestMapInternalPrompt_TextOnly(t *testing.T) {
	prompt := Prompt{
		System: "You are a helpful assistant",
		User:   "Hello",
	}

	messages := mapInternalPrompt(prompt)

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}
	if messages[0].Role != "system" {
		t.Errorf("Expected system role, got %s", messages[0].Role)
	}
	if messages[1].Role != "user" {
		t.Errorf("Expected user role, got %s", messages[1].Role)
	}
}

func TestMapInternalPrompt_WithImageURL(t *testing.T) {
	prompt := Prompt{
		User: "Describe this image",
		Images: []ImageData{
			{URL: "https://example.com/image.jpg"},
		},
	}

	messages := mapInternalPrompt(prompt)

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	content, ok := messages[0].Content.([]map[string]interface{})
	if !ok {
		t.Fatal("Content should be an array for multimodal")
	}
	if len(content) != 2 {
		t.Fatalf("Expected 2 content items (text + image), got %d", len(content))
	}
	if content[0]["type"] != "text" {
		t.Errorf("First content should be text, got %v", content[0]["type"])
	}
	if content[1]["type"] != "image_url" {
		t.Errorf("Second content should be image_url, got %v", content[1]["type"])
	}
}

func TestMapInternalPrompt_WithBase64Image(t *testing.T) {
	prompt := Prompt{
		User: "What's in this image?",
		Images: []ImageData{
			{Base64: "base64encodeddata"},
		},
	}

	messages := mapInternalPrompt(prompt)

	content, ok := messages[0].Content.([]map[string]interface{})
	if !ok {
		t.Fatal("Content should be an array")
	}

	imageContent := content[1]
	if imageContent["type"] != "image_url" {
		t.Error("Expected image_url type")
	}

	imageURL := imageContent["image_url"].(map[string]string)
	url := imageURL["url"]
	if !strings.Contains(url, "data:image/jpeg;base64,") {
		t.Errorf("Expected data URL prefix, got %s", url)
	}
}

func TestMapInternalPrompt_MultipleImages(t *testing.T) {
	prompt := Prompt{
		User: "Compare these",
		Images: []ImageData{
			{URL: "https://example.com/1.jpg"},
			{URL: "https://example.com/2.jpg"},
		},
	}

	messages := mapInternalPrompt(prompt)

	content, ok := messages[0].Content.([]map[string]interface{})
	if !ok {
		t.Fatal("Content should be an array")
	}
	if len(content) != 3 {
		t.Fatalf("Expected 3 content items (1 text + 2 images), got %d", len(content))
	}
}
