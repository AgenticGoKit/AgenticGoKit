package workflow

import (
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"
	"github.com/agenticgokit/agenticgokit/examples/story-writer-chat-v2/config"
)

// AgentConfig holds configuration for creating an agent
type AgentConfig struct {
	SystemPrompt string
	Temperature  float32
	MaxTokens    int
}

// CreateWriter creates the Writer agent with memory enabled
func CreateWriter(cfg *config.Config) (vnext.Agent, error) {
	return vnext.QuickChatAgentWithConfig("Writer", &vnext.Config{
		Name:         "writer",
		SystemPrompt: WriterSystemPrompt,
		Timeout:      90 * time.Second,
		Streaming:    &vnext.StreamingConfig{Enabled: true, BufferSize: 50, FlushInterval: 50},
		Memory: &vnext.MemoryConfig{
			Provider: "memory", // In-memory provider
			RAG: &vnext.RAGConfig{
				MaxTokens:       2000, // Include up to 2000 tokens of conversation history
				PersonalWeight:  0.8,  // Prioritize conversation history
				KnowledgeWeight: 0.2,
				HistoryLimit:    5, // Include last 5 messages
			},
		},
		LLM: vnext.LLMConfig{
			Provider:    cfg.Provider,
			Model:       cfg.Model,
			Temperature: 0.3,
			MaxTokens:   500,
			APIKey:      cfg.APIKey,
		},
	})
}

// CreateEditor creates the Editor agent for spell checking
func CreateEditor(cfg *config.Config) (vnext.Agent, error) {
	return vnext.QuickChatAgentWithConfig("Editor", &vnext.Config{
		Name:         "editor",
		SystemPrompt: EditorSystemPrompt,
		Timeout:      90 * time.Second,
		Streaming:    &vnext.StreamingConfig{Enabled: true, BufferSize: 50, FlushInterval: 50},
		LLM: vnext.LLMConfig{
			Provider:    cfg.Provider,
			Model:       cfg.Model,
			Temperature: 0.0,
			MaxTokens:   300,
			APIKey:      cfg.APIKey,
		},
	})
}

// CreatePublisher creates the Publisher agent for formatting
func CreatePublisher(cfg *config.Config) (vnext.Agent, error) {
	return vnext.QuickChatAgentWithConfig("Publisher", &vnext.Config{
		Name:         "publisher",
		SystemPrompt: PublisherSystemPrompt,
		Timeout:      90 * time.Second,
		Streaming:    &vnext.StreamingConfig{Enabled: true, BufferSize: 50, FlushInterval: 50},
		LLM: vnext.LLMConfig{
			Provider:    cfg.Provider,
			Model:       cfg.Model,
			Temperature: 0.1,
			MaxTokens:   600,
			APIKey:      cfg.APIKey,
		},
	})
}



