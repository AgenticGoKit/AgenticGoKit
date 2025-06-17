// Package main provides a minimal MCP demo that should not hang.
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kunalkushwaha/agentflow/core"
	"github.com/kunalkushwaha/agentflow/internal/llm"
)

// SimpleLLMProvider with very short prompts to avoid hanging
type SimpleLLMProvider struct {
	adapter *llm.OllamaAdapter
}

func NewSimpleLLMProvider() (*SimpleLLMProvider, error) {
	adapter, err := llm.NewOllamaAdapter("", "llama3.2:latest", 50, 0.1) // Very short responses
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama adapter: %w", err)
	}

	return &SimpleLLMProvider{adapter: adapter}, nil
}

func (p *SimpleLLMProvider) Call(ctx context.Context, prompt core.Prompt) (core.Response, error) {
	fmt.Printf("🧠 Quick Ollama call for: %s\n", strings.TrimSpace(prompt.User)[:min(50, len(prompt.User))])

	// Very short timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	// Very simple prompt
	systemPrompt := `Respond with a JSON array of 1-2 tool names only. Tools: web_search, content_fetch, summarize_text`

	internalPrompt := llm.Prompt{
		System: systemPrompt,
		User:   prompt.User,
		Parameters: llm.ModelParameters{
			MaxTokens:   func() *int32 { v := int32(30); return &v }(),
			Temperature: func() *float32 { v := float32(0.1); return &v }(),
		},
	}

	response, err := p.adapter.Call(timeoutCtx, internalPrompt)
	if err != nil {
		fmt.Printf("⚠️  Ollama timeout, using fallback\n")
		return core.Response{
			Content:      `["web_search"]`,
			FinishReason: "fallback",
		}, nil
	}

	fmt.Printf("✅ Got response: %s\n", response.Content)
	return core.Response{
		Content:      response.Content,
		FinishReason: "stop",
	}, nil
}

func (p *SimpleLLMProvider) Stream(ctx context.Context, prompt core.Prompt) (<-chan core.Token, error) {
	return nil, fmt.Errorf("streaming not implemented")
}

func (p *SimpleLLMProvider) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return nil, fmt.Errorf("embeddings not implemented")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	fmt.Println("🚀 Minimal MCP Demo - No Hanging Test")
	fmt.Println("=====================================")

	ctx := context.Background()

	// Test simple LLM
	fmt.Println("\n📡 Testing Simple LLM Provider")
	llmProvider, err := NewSimpleLLMProvider()
	if err != nil {
		log.Fatalf("❌ Failed to create LLM provider: %v", err)
	}

	// Test 3 simple queries
	queries := []string{
		"search for AI news",
		"fetch content",
		"analyze data",
	}

	for i, query := range queries {
		fmt.Printf("\n🧪 Test %d: %s\n", i+1, query)

		prompt := core.Prompt{User: query}
		response, err := llmProvider.Call(ctx, prompt)

		if err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
		} else {
			fmt.Printf("✅ Success: %s\n", response.Content)
		}

		if i < len(queries)-1 {
			fmt.Printf("⏳ Waiting 1 second...\n")
			time.Sleep(1 * time.Second)
		}
	}

	fmt.Println("\n🎉 Simple demo completed!")
}
