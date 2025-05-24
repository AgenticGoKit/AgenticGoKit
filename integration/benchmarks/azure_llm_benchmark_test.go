package benchmarks_test // Changed from llm_test

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kunalkushwaha/agentflow/internal/llm" // Keep importing the package under test
)

// Helper to get Azure Adapter for benchmarks, skipping if not configured
func setupAzureBenchmarkAdapter(b *testing.B) llm.ModelProvider {
	b.Helper()
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
	chatDeployment := os.Getenv("AZURE_OPENAI_CHAT_DEPLOYMENT")
	embeddingDeployment := os.Getenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT")

	if endpoint == "" || apiKey == "" || chatDeployment == "" || embeddingDeployment == "" {
		b.Skip("Skipping Azure benchmark: AZURE_OPENAI_* environment variables not fully set")
	}

	// Use a shared client for benchmarks if appropriate, or create new ones
	client := &http.Client{Timeout: 60 * time.Second}

	opts := llm.AzureOpenAIAdapterOptions{
		Endpoint:            endpoint,
		APIKey:              apiKey,
		ChatDeployment:      chatDeployment,
		EmbeddingDeployment: embeddingDeployment,
		HTTPClient:          client,
	}
	adapter, err := llm.NewAzureOpenAIAdapter(opts)
	if err != nil {
		b.Fatalf("Failed to create Azure adapter for benchmark: %v", err)
	}
	return adapter
}

// BenchmarkAzureCall measures the latency of the Call method.
func BenchmarkAzureCall(b *testing.B) {
	adapter := setupAzureBenchmarkAdapter(b)
	ctx := context.Background()
	prompt := llm.Prompt{
		System: "You are a benchmark assistant.",
		User:   "Briefly explain the concept of Go benchmarks.",
		// Add parameters if needed, e.g., max tokens to limit response size
		// Parameters: llm.ModelParameters{ MaxTokens: to.Ptr(50) },
	}

	b.ResetTimer() // Start timing after setup
	for i := 0; i < b.N; i++ {
		_, err := adapter.Call(ctx, prompt)
		if err != nil {
			// Stop timer before fatal error
			b.StopTimer()
			b.Fatalf("adapter.Call failed during benchmark: %v", err)
		}
	}
	b.StopTimer() // Stop timer after the loop
}

// BenchmarkAzureEmbeddings measures the latency of the Embeddings method.
func BenchmarkAzureEmbeddings(b *testing.B) {
	adapter := setupAzureBenchmarkAdapter(b)
	ctx := context.Background()
	texts := []string{"This is a sample text for benchmarking embeddings."}

	b.ResetTimer() // Start timing after setup
	for i := 0; i < b.N; i++ {
		_, err := adapter.Embeddings(ctx, texts)
		if err != nil {
			// Stop timer before fatal error
			b.StopTimer()
			// Check for the specific known issue with gpt-4o for embeddings
			if strings.Contains(err.Error(), "OperationNotSupported") && strings.Contains(err.Error(), "embeddings operation does not work with the specified model") {
				b.Skipf("Skipping Embeddings benchmark: Configured deployment likely uses a chat model incompatible with embeddings. Error: %v", err)
			}
			b.Fatalf("adapter.Embeddings failed during benchmark: %v", err)
		}
	}
	b.StopTimer() // Stop timer after the loop
}

// TODO: Add benchmarks for OpenAIAdapter and OllamaAdapter in their respective test files.
