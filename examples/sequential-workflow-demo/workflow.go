package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"
)

// SimpleSequentialWorkflow demonstrates a basic 2-agent sequential workflow
// Agent1 (Researcher) → Agent2 (Writer)
type SimpleSequentialWorkflow struct {
	researcher vnext.Agent
	writer     vnext.Agent
}

// NewSimpleSequentialWorkflow creates a simple 2-agent workflow
func NewSimpleSequentialWorkflow(apiKey string) (*SimpleSequentialWorkflow, error) {
	// Agent 1: Researcher
	researcher, err := vnext.QuickChatAgentWithConfig("Researcher", &vnext.Config{
		Name: "researcher",
		SystemPrompt: `You are a Research Assistant. Your role is to:
- Research the given topic thoroughly
- Find key facts, statistics, and interesting insights
- Organize information clearly
- Provide concise but informative research notes (150-200 words)

Provide research notes for the topic given by the user.`,
		Timeout: 60 * time.Second,
		Streaming: &vnext.StreamingConfig{
			Enabled:       true,
			BufferSize:    50,
			FlushInterval: 50,
		},
		LLM: vnext.LLMConfig{
			Provider:    "openrouter",
			Model:       "openai/gpt-4o-mini",
			Temperature: 0.7,
			MaxTokens:   500,
			APIKey:      apiKey,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create researcher agent: %w", err)
	}

	// Agent 2: Writer
	writer, err := vnext.QuickChatAgentWithConfig("Writer", &vnext.Config{
		Name: "writer",
		SystemPrompt: `You are a Professional Writer. Your role is to:
- Take research notes and turn them into engaging content
- Write in a clear, accessible style
- Create a well-structured article or summary
- Make the content interesting and easy to read (200-300 words)

Write an article based on the research notes provided.`,
		Timeout: 60 * time.Second,
		Streaming: &vnext.StreamingConfig{
			Enabled:       true,
			BufferSize:    50,
			FlushInterval: 50,
		},
		LLM: vnext.LLMConfig{
			Provider:    "openrouter",
			Model:       "openai/gpt-4o-mini",
			Temperature: 0.8,
			MaxTokens:   600,
			APIKey:      apiKey,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create writer agent: %w", err)
	}

	return &SimpleSequentialWorkflow{
		researcher: researcher,
		writer:     writer,
	}, nil
}

// Name implements WorkflowExecutor interface
func (ssw *SimpleSequentialWorkflow) Name() string {
	return "Research & Write"
}

// WelcomeMessage implements WorkflowExecutor interface
func (ssw *SimpleSequentialWorkflow) WelcomeMessage() string {
	return "Welcome! Give me a topic and I'll research it and write an article about it."
}

// GetAgents implements WorkflowExecutor interface
func (ssw *SimpleSequentialWorkflow) GetAgents() []AgentInfo {
	return []AgentInfo{
		{
			Name:        "researcher",
			DisplayName: "Researcher",
			Icon:        "🔍",
			Color:       "blue",
			Description: "Gathers facts and insights",
		},
		{
			Name:        "writer",
			DisplayName: "Writer",
			Icon:        "✍️",
			Color:       "purple",
			Description: "Creates engaging article",
		},
	}
}

// Execute implements WorkflowExecutor interface
// This is a simple sequential workflow: Researcher → Writer
func (ssw *SimpleSequentialWorkflow) Execute(ctx context.Context, userInput string, sendMessage MessageSender) error {
	// Send workflow start
	sendMessage(WSMessage{
		Type:      MsgTypeWorkflowStart,
		Content:   "Starting research and writing workflow...",
		Timestamp: float64(time.Now().Unix()),
	})

	// Phase 1: Research
	sendMessage(WSMessage{
		Type:      MsgTypeAgentStart,
		Content:   "🔍 Researcher is gathering information...",
		Agent:     "researcher",
		Step:      "research",
		Progress:  50,
		Timestamp: float64(time.Now().Unix()),
	})

	researchNotes, err := ssw.runAgentWithStreaming(ctx, ssw.researcher, "researcher",
		fmt.Sprintf("Research this topic and provide key facts and insights: %s", userInput),
		sendMessage)
	if err != nil {
		return fmt.Errorf("researcher failed: %w", err)
	}

	sendMessage(WSMessage{
		Type:      MsgTypeAgentComplete,
		Content:   researchNotes,
		Agent:     "researcher",
		Timestamp: float64(time.Now().Unix()),
	})

	// Phase 2: Write
	sendMessage(WSMessage{
		Type:      MsgTypeAgentStart,
		Content:   "✍️ Writer is creating an article...",
		Agent:     "writer",
		Step:      "write",
		Progress:  100,
		Timestamp: float64(time.Now().Unix()),
	})

	article, err := ssw.runAgentWithStreaming(ctx, ssw.writer, "writer",
		fmt.Sprintf("Write an engaging article based on these research notes:\n\n%s", researchNotes),
		sendMessage)
	if err != nil {
		return fmt.Errorf("writer failed: %w", err)
	}

	sendMessage(WSMessage{
		Type:      MsgTypeAgentComplete,
		Content:   article,
		Agent:     "writer",
		Timestamp: float64(time.Now().Unix()),
	})

	// Workflow complete
	sendMessage(WSMessage{
		Type:      MsgTypeWorkflowDone,
		Content:   article,
		Timestamp: float64(time.Now().Unix()),
		Metadata: map[string]interface{}{
			"success":      true,
			"total_length": len(article),
		},
	})

	return nil
}

// Cleanup implements WorkflowExecutor interface
func (ssw *SimpleSequentialWorkflow) Cleanup(ctx context.Context) error {
	return nil
}

// runAgentWithStreaming runs an agent and streams the output
func (ssw *SimpleSequentialWorkflow) runAgentWithStreaming(ctx context.Context, agent vnext.Agent, agentName string, prompt string, sendMessage MessageSender) (string, error) {
	stream, err := agent.RunStream(ctx, prompt)
	if err != nil {
		return "", err
	}

	var fullOutput strings.Builder

	for chunk := range stream.Chunks() {
		if chunk.Error != nil {
			sendMessage(WSMessage{
				Type:      MsgTypeError,
				Content:   fmt.Sprintf("Error: %v", chunk.Error),
				Agent:     agentName,
				Timestamp: float64(time.Now().Unix()),
			})
			break
		}

		switch chunk.Type {
		case vnext.ChunkTypeText:
			fullOutput.WriteString(chunk.Content)
			sendMessage(WSMessage{
				Type:      MsgTypeAgentProgress,
				Content:   chunk.Content,
				Agent:     agentName,
				Timestamp: float64(time.Now().Unix()),
			})

		case vnext.ChunkTypeDelta:
			fullOutput.WriteString(chunk.Delta)
			sendMessage(WSMessage{
				Type:      MsgTypeAgentProgress,
				Content:   chunk.Delta,
				Agent:     agentName,
				Timestamp: float64(time.Now().Unix()),
			})
		}
	}

	result, err := stream.Wait()
	if err != nil {
		return "", err
	}

	output := fullOutput.String()
	if output == "" {
		output = result.Content
	}

	return output, nil
}

// ValidateAPIKey checks if the OpenRouter API key is set
func ValidateAPIKey() (string, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENROUTER_API_KEY environment variable not set")
	}
	return apiKey, nil
}



