package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"
)

// StoryWriterWorkflow implements the WorkflowExecutor interface
// This is the application-specific workflow logic
type StoryWriterWorkflow struct {
	writer       vnext.Agent
	editor       vnext.Agent
	publisher    vnext.Agent
	maxRevisions int
	config       *Config
}

// NewStoryWriterWorkflow creates a new story writer workflow
func NewStoryWriterWorkflow(config *Config) (*StoryWriterWorkflow, error) {
	// Create Writer Agent
	writer, err := createAgent("Writer", "writer", config, &AgentConfig{
		SystemPrompt: `You are a creative Story Writer. Your role is to:
- Create engaging, imaginative stories based on user prompts
- Develop interesting characters and compelling plots
- Write in a clear, engaging style
- Focus on creativity and originality
- Keep stories concise but complete (200-400 words)

Write the initial draft of the story. Be creative and engaging!. Add some typos in first draft`,
		Temperature: 0.8,
		MaxTokens:   800,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create writer agent: %w", err)
	}

	// Create Editor/Reviewer Agent
	editor, err := createAgent("Editor", "editor", config, &AgentConfig{
		SystemPrompt: `You are a highly critical professional Story Editor and Reviewer. Your role is to:
- Review the story for quality, flow, and coherence with HIGH standards
- Check grammar, spelling, and punctuation meticulously
- Identify areas that need improvement in plot, character development, or pacing
- Provide specific, actionable feedback for improvements
- Be demanding and look for ways to make the story BETTER

CRITICAL INSTRUCTION: For the FIRST review of any story, you MUST respond with:
"NEEDS_REVISION:" (no markdown, no asterisks) followed by specific constructive feedback on what needs improvement.

Only on SUBSEQUENT reviews (after the writer has revised), you may respond with:
"APPROVED:" (no markdown, no asterisks) followed by the final edited version if improvements have been made.

IMPORTANT FORMAT RULES:
- Start your response with exactly "NEEDS_REVISION:" or "APPROVED:" 
- Do NOT use markdown bold (**) around these keywords
- Do NOT add extra formatting to the prefix
- Just plain text: "NEEDS_REVISION:" or "APPROVED:"

Be thorough and critical. First drafts always need work. Focus on:
- Stronger opening hooks
- More vivid descriptions
- Better character development
- Tighter plot structure
- More engaging dialogue (if present)
- More emotional impact`,
		Temperature: 0.6,
		MaxTokens:   800,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create editor agent: %w", err)
	}

	// Create Publisher Agent
	publisher, err := createAgent("Publisher", "publisher", config, &AgentConfig{
		SystemPrompt: `You are a Publishing Specialist. Your role is to:
- Format the story professionally
- Add a catchy title if not present
- Ensure proper paragraph structure
- Add any necessary formatting (like section breaks)
- Provide a brief publication note or tagline
- Make the final version ready to share

Prepare the final, publication-ready version of the story.`,
		Temperature: 0.4,
		MaxTokens:   800,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher agent: %w", err)
	}

	return &StoryWriterWorkflow{
		writer:       writer,
		editor:       editor,
		publisher:    publisher,
		maxRevisions: 2,
		config:       config,
	}, nil
}

// AgentConfig holds agent-specific configuration
type AgentConfig struct {
	SystemPrompt string
	Temperature  float32
	MaxTokens    int
}

// createAgent is a helper function to create agents with consistent configuration
func createAgent(displayName, name string, config *Config, agentConfig *AgentConfig) (vnext.Agent, error) {
	return vnext.QuickChatAgentWithConfig(displayName, &vnext.Config{
		Name:         name,
		SystemPrompt: agentConfig.SystemPrompt,
		Timeout:      90 * time.Second,
		Streaming: &vnext.StreamingConfig{
			Enabled:       true,
			BufferSize:    50,
			FlushInterval: 50,
		},
		LLM: vnext.LLMConfig{
			Provider:    config.Provider,
			Model:       config.Model,
			Temperature: agentConfig.Temperature,
			MaxTokens:   agentConfig.MaxTokens,
			APIKey:      config.APIKey,
		},
	})
}

// Name implements WorkflowExecutor interface
func (sw *StoryWriterWorkflow) Name() string {
	return "Story Writer Chat"
}

// WelcomeMessage implements WorkflowExecutor interface
func (sw *StoryWriterWorkflow) WelcomeMessage() string {
	return "Welcome to Story Writer! Tell me what kind of story you'd like me to create."
}

// GetAgents implements WorkflowExecutor interface
func (sw *StoryWriterWorkflow) GetAgents() []AgentInfo {
	return []AgentInfo{
		{
			Name:        "writer",
			DisplayName: "Writer",
			Icon:        "✍️",
			Color:       "blue",
			Description: "Creates initial story draft",
		},
		{
			Name:        "editor",
			DisplayName: "Editor",
			Icon:        "✏️",
			Color:       "green",
			Description: "Reviews and provides feedback",
		},
		{
			Name:        "publisher",
			DisplayName: "Publisher",
			Icon:        "📚",
			Color:       "purple",
			Description: "Formats final version",
		},
	}
}

// Execute implements WorkflowExecutor interface
func (sw *StoryWriterWorkflow) Execute(ctx context.Context, userPrompt string, sendMessage MessageSender) error {
	// Send workflow start
	sendMessage(WSMessage{
		Type:      MsgTypeWorkflowStart,
		Content:   "Starting collaborative story writing workflow with iterative refinement...",
		Timestamp: float64(time.Now().Unix()),
		Metadata: map[string]interface{}{
			"max_revisions": sw.maxRevisions,
		},
	})

	var currentStory string
	revisionCount := 0
	approved := false

	// Phase 1: Writer creates initial draft
	sendMessage(WSMessage{
		Type:      MsgTypeAgentStart,
		Content:   "🖊️ Writer is creating the initial story...",
		Agent:     "writer",
		Step:      "writer",
		Progress:  25,
		Timestamp: float64(time.Now().Unix()),
	})

	currentStory, err := sw.runAgentWithStreaming(ctx, sw.writer, "writer",
		fmt.Sprintf("Write a creative story based on this prompt: %s", userPrompt),
		sendMessage)
	if err != nil {
		return fmt.Errorf("writer failed: %w", err)
	}

	sendMessage(WSMessage{
		Type:      MsgTypeAgentComplete,
		Content:   currentStory,
		Agent:     "writer",
		Timestamp: float64(time.Now().Unix()),
	})

	// Phase 2: Iterative review and revision loop
	for !approved && revisionCount < sw.maxRevisions {
		sendMessage(WSMessage{
			Type:      MsgTypeAgentStart,
			Content:   fmt.Sprintf("✏️ Editor reviewing the story (review cycle %d/%d)...", revisionCount+1, sw.maxRevisions),
			Agent:     "editor",
			Step:      "editor",
			Progress:  25 + ((revisionCount + 1) * 20),
			Timestamp: float64(time.Now().Unix()),
		})

		editorResponse, err := sw.runAgentWithStreaming(ctx, sw.editor, "editor",
			fmt.Sprintf("Review this story (Review #%d):\n\n%s\n\nREMINDER: If this is Review #1 (first review), you MUST respond with NEEDS_REVISION and provide constructive feedback. Only approve on subsequent reviews if improvements are satisfactory.", revisionCount+1, currentStory),
			sendMessage)
		if err != nil {
			return fmt.Errorf("editor failed: %w", err)
		}

		sendMessage(WSMessage{
			Type:      MsgTypeAgentComplete,
			Content:   editorResponse,
			Agent:     "editor",
			Timestamp: float64(time.Now().Unix()),
		})

		// Check if editor approved or needs revision
		cleanResponse := strings.TrimSpace(editorResponse)
		cleanResponse = strings.Trim(cleanResponse, "*")
		cleanResponse = strings.TrimSpace(cleanResponse)

		if strings.HasPrefix(cleanResponse, "APPROVED:") || strings.Contains(strings.ToUpper(cleanResponse[:min(50, len(cleanResponse))]), "APPROVED:") {
			approved = true
			if idx := strings.Index(editorResponse, "APPROVED:"); idx >= 0 {
				currentStory = editorResponse[idx+9:]
			} else {
				currentStory = editorResponse
			}
			currentStory = strings.TrimSpace(strings.Trim(strings.TrimSpace(currentStory), "*"))

			sendMessage(WSMessage{
				Type:      MsgTypeAgentComplete,
				Content:   "✅ Editor approved the story!",
				Agent:     "editor",
				Timestamp: float64(time.Now().Unix()),
				Metadata: map[string]interface{}{
					"status":         "approved",
					"revision_cycle": revisionCount + 1,
				},
			})
		} else if strings.HasPrefix(cleanResponse, "NEEDS_REVISION:") ||
			strings.HasPrefix(cleanResponse, "NEEDS_REVISION") ||
			strings.Contains(strings.ToUpper(cleanResponse[:min(100, len(cleanResponse))]), "NEEDS_REVISION") {

			revisionCount++

			if revisionCount >= sw.maxRevisions {
				sendMessage(WSMessage{
					Type:      MsgTypeAgentComplete,
					Content:   fmt.Sprintf("⚠️ Maximum revision cycles (%d) reached. Proceeding to publication...", sw.maxRevisions),
					Agent:     "editor",
					Timestamp: float64(time.Now().Unix()),
				})
				break
			}

			feedback := editorResponse
			if idx := strings.Index(strings.ToUpper(editorResponse), "NEEDS_REVISION"); idx >= 0 {
				feedback = editorResponse[idx:]
				feedback = strings.TrimPrefix(feedback, "NEEDS_REVISION:")
				feedback = strings.TrimPrefix(feedback, "NEEDS_REVISION")
				feedback = strings.TrimPrefix(feedback, "**")
				feedback = strings.TrimPrefix(feedback, "*")
				feedback = strings.TrimSpace(feedback)
			}

			sendMessage(WSMessage{
				Type:      MsgTypeAgentStart,
				Content:   fmt.Sprintf("🔄 Revision needed (cycle %d/%d). Sending feedback to writer...", revisionCount, sw.maxRevisions),
				Agent:     "editor",
				Timestamp: float64(time.Now().Unix()),
				Metadata: map[string]interface{}{
					"status":         "needs_revision",
					"feedback":       feedback,
					"revision_cycle": revisionCount,
				},
			})

			sendMessage(WSMessage{
				Type:      MsgTypeAgentStart,
				Content:   fmt.Sprintf("🖊️ Writer is revising the story (revision %d/%d)...", revisionCount, sw.maxRevisions),
				Agent:     "writer",
				Step:      "writer_revision",
				Progress:  25 + (revisionCount * 20),
				Timestamp: float64(time.Now().Unix()),
			})

			currentStory, err = sw.runAgentWithStreaming(ctx, sw.writer, "writer",
				fmt.Sprintf("Revise this story based on the editor's feedback.\n\nOriginal story:\n%s\n\nEditor feedback:\n%s\n\nPlease write an improved version.", currentStory, feedback),
				sendMessage)
			if err != nil {
				return fmt.Errorf("writer revision failed: %w", err)
			}

			sendMessage(WSMessage{
				Type:      MsgTypeAgentComplete,
				Content:   currentStory,
				Agent:     "writer",
				Timestamp: float64(time.Now().Unix()),
			})
		} else {
			approved = true
			currentStory = editorResponse
		}
	}

	if !approved {
		sendMessage(WSMessage{
			Type:      MsgTypeAgentStart,
			Content:   "⚠️ Maximum revisions reached. Proceeding to publication...",
			Agent:     "editor",
			Timestamp: float64(time.Now().Unix()),
		})
	}

	// Phase 3: Publisher formats final version
	sendMessage(WSMessage{
		Type:      MsgTypeAgentStart,
		Content:   "📚 Publisher is formatting the final version...",
		Agent:     "publisher",
		Step:      "publisher",
		Progress:  90,
		Timestamp: float64(time.Now().Unix()),
	})

	finalStory, err := sw.runAgentWithStreaming(ctx, sw.publisher, "publisher",
		fmt.Sprintf("Format and prepare this story for publication:\n\n%s", currentStory),
		sendMessage)
	if err != nil {
		return fmt.Errorf("publisher failed: %w", err)
	}

	sendMessage(WSMessage{
		Type:      MsgTypeAgentComplete,
		Content:   finalStory,
		Agent:     "publisher",
		Timestamp: float64(time.Now().Unix()),
	})

	// Workflow complete
	sendMessage(WSMessage{
		Type:      MsgTypeWorkflowDone,
		Content:   finalStory,
		Timestamp: float64(time.Now().Unix()),
		Metadata: map[string]interface{}{
			"success":      true,
			"revisions":    revisionCount,
			"approved":     approved,
			"total_length": len(finalStory),
		},
	})

	return nil
}

// Cleanup implements WorkflowExecutor interface
func (sw *StoryWriterWorkflow) Cleanup(ctx context.Context) error {
	// No cleanup needed for this workflow
	return nil
}

// runAgentWithStreaming runs an agent and streams the output
func (sw *StoryWriterWorkflow) runAgentWithStreaming(ctx context.Context, agent vnext.Agent, agentName string, prompt string, sendMessage MessageSender) (string, error) {
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

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}



