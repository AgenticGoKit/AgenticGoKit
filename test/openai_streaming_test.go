package test

import (
	"context"
	"testing"
	"time"

	"github.com/agenticgokit/agenticgokit/internal/llm"
)

// TestOpenAIStreamingImplementation tests that OpenAI streaming is properly implemented
func TestOpenAIStreamingImplementation(t *testing.T) {
	// Create adapter with dummy API key for testing interface compliance
	adapter, err := llm.NewOpenAIAdapter("dummy-key", "gpt-4o-mini", 100, 0.7)
	if err != nil {
		t.Fatalf("Failed to create OpenAI adapter: %v", err)
	}

	// Test that Stream method exists and returns correct types
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	prompt := llm.Prompt{
		User: "Hello, world!",
	}

	// This should not panic and should return a channel
	tokenChan, err := adapter.Stream(ctx, prompt)

	// We expect an error since we're using a dummy API key, but we want to verify
	// that the method signature is correct and it doesn't panic
	if tokenChan == nil && err == nil {
		t.Error("Stream should return either a channel or an error")
	}

	// If we got a channel, it should be the correct type
	if tokenChan != nil {
		// Just verify the channel type by reading from it (with timeout)
		select {
		case token, ok := <-tokenChan:
			if ok {
				// We got a token (or error token), which is expected
				t.Logf("Received token: %+v", token)
			}
		case <-time.After(1 * time.Second):
			// Timeout is fine, we just wanted to verify the channel type
			t.Log("Stream method timeout (expected with dummy credentials)")
		}
	}

	t.Log("OpenAI streaming implementation has correct interface")
}

// TestOpenAIStreamingMethodSignature verifies the Stream method signature
func TestOpenAIStreamingMethodSignature(t *testing.T) {
	adapter, err := llm.NewOpenAIAdapter("dummy-key", "gpt-4o-mini", 100, 0.7)
	if err != nil {
		t.Fatalf("Failed to create OpenAI adapter: %v", err)
	}

	// Verify that the adapter implements the ModelProvider interface
	var _ llm.ModelProvider = adapter

	t.Log("OpenAI adapter correctly implements ModelProvider interface with streaming")
}

