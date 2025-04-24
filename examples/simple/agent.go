package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/llm"
)

// ChatAgent uses a ModelProvider to respond to user prompts.
type ChatAgent struct {
	provider llm.ModelProvider
}

// NewChatAgent creates a ChatAgent.
func NewChatAgent(provider llm.ModelProvider) *ChatAgent {
	if provider == nil {
		log.Fatal("ChatAgent requires a non-nil ModelProvider")
	}
	return &ChatAgent{provider: provider}
}

// Run implements the agentflow.Agent interface.
func (a *ChatAgent) Run(ctx context.Context, in agentflow.State) (agentflow.State, error) {
	log.Println("ChatAgent: Running...")

	// 1. Get user prompt from input state
	userInputVal, ok := in.Get("user_prompt")
	if !ok {
		return in, fmt.Errorf("ChatAgent: 'user_prompt' not found in input state")
	}
	userInput, ok := userInputVal.(string)
	if !ok || userInput == "" {
		return in, fmt.Errorf("ChatAgent: 'user_prompt' is not a non-empty string")
	}
	log.Printf("ChatAgent: Received prompt: %q", userInput)

	// 2. Prepare LLM prompt
	prompt := llm.Prompt{
		System: "You are a helpful assistant.",
		User:   userInput,
		// Parameters can be added here if needed
		// Parameters: llm.ModelParameters{ Temperature: to.Ptr(0.7) }
	}

	// 3. Call the LLM
	log.Println("ChatAgent: Calling LLM...")
	resp, err := a.provider.Call(ctx, prompt)
	if err != nil {
		log.Printf("ChatAgent: LLM call failed: %v", err)
		// Return original state and error
		return in, fmt.Errorf("ChatAgent: LLM call failed: %w", err)
	}
	log.Printf("ChatAgent: LLM response received: FinishReason=%s", resp.FinishReason)

	// 4. Create output state and add response
	out := in.Clone()
	out.Set("llm_response", resp.Content)
	log.Printf("ChatAgent: Added response to state.")

	return out, nil
}

func main() {
	log.Println("Starting Chat Agent Example...")

	// --- Configuration (Read from Environment Variables) ---
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
	chatDeployment := os.Getenv("AZURE_OPENAI_CHAT_DEPLOYMENT")
	// Embedding deployment not strictly needed for chat, but constructor requires it
	embeddingDeployment := os.Getenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT")
	if embeddingDeployment == "" {
		embeddingDeployment = "not-used-in-this-example" // Provide a non-empty dummy if not set
		log.Println("Warning: AZURE_OPENAI_EMBEDDING_DEPLOYMENT not set, using dummy value.")
	}

	if endpoint == "" || apiKey == "" || chatDeployment == "" {
		log.Fatal("Error: Set AZURE_OPENAI_ENDPOINT, AZURE_OPENAI_API_KEY, and AZURE_OPENAI_CHAT_DEPLOYMENT environment variables.")
	}

	// --- Setup ---
	// Create LLM Adapter
	adapterOpts := llm.AzureOpenAIAdapterOptions{
		Endpoint:            endpoint,
		APIKey:              apiKey,
		ChatDeployment:      chatDeployment,
		EmbeddingDeployment: embeddingDeployment,                     // Required by constructor
		HTTPClient:          &http.Client{Timeout: 90 * time.Second}, // Use a reasonable timeout
	}
	azureAdapter, err := llm.NewAzureOpenAIAdapter(adapterOpts)
	if err != nil {
		log.Fatalf("Error creating Azure OpenAI Adapter: %v", err)
	}
	log.Println("Azure OpenAI Adapter created.")

	// Create Chat Agent
	chatAgent := NewChatAgent(azureAdapter)
	log.Println("Chat Agent created.")

	// --- Execution ---
	// Prepare initial state
	initialState := agentflow.NewState()
	initialState.Set("user_prompt", "Explain the concept of Go interfaces in simple terms.")
	log.Printf("Initial state prepared.")

	// Run the agent
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second) // Add a timeout for the agent run
	defer cancel()
	finalState, err := chatAgent.Run(ctx, initialState)
	if err != nil {
		log.Fatalf("Agent execution failed: %v", err)
	}

	// --- Output ---
	log.Println("Agent execution successful.")
	llmResponse, ok := finalState.Get("llm_response")
	if !ok {
		log.Println("LLM response not found in final state.")
	} else {
		fmt.Println("\n--- LLM Response ---")
		fmt.Printf("%v\n", llmResponse) // Use %v in case it's not a string
		fmt.Println("--------------------")
	}

	// Optionally print the full final state
	// log.Printf("Final state data: %+v", finalState.GetData())
}
