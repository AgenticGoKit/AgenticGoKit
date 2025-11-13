package workflow

import (
	"context"
	"fmt"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"
	"github.com/agenticgokit/agenticgokit/examples/story-writer-chat-v2/config"
	"github.com/agenticgokit/agenticgokit/examples/story-writer-chat-v2/infrastructure"
)

// StoryWriterWorkflow demonstrates NESTED SubWorkflowAgent composition with conditional loop termination
//
// Architecture:
//
//	Sequential Pipeline (storyPipeline SubWorkflow)
//	└── Step 1: Revision Loop (revisionLoop SubWorkflow)
//	    └── LoopWorkflow (exits on "APPROVED" OR max 3 iterations)
//	        ├── Writer Agent (creates/revises draft)
//	        └── Editor Agent (reviews, provides feedback or approves)
//	└── Step 2: Publisher Agent (formats approved story)
//
// Key Features:
//   - LoopWorkflow with conditional termination (exits when editor says "APPROVED")
//   - Loop workflow wrapped as a SubWorkflowAgent
//   - SubWorkflowAgent used as a step in Sequential workflow
//   - Sequential workflow also wrapped as a SubWorkflowAgent
//   - Result: Complex nested workflows with smart loop exit, composed as simple agents!
type StoryWriterWorkflow struct {
	// Individual agents
	writer    vnext.Agent
	editor    vnext.Agent
	publisher vnext.Agent

	// Nested SubWorkflows
	revisionLoop  vnext.Agent // SubWorkflow: Loop[Writer<->Editor]
	storyPipeline vnext.Agent // SubWorkflow: Sequential[RevisionLoop -> Publisher]

	config        *config.Config
	messageSender infrastructure.MessageSender // Callback to send messages to UI
}

// NewStoryWriterWorkflow creates nested SubWorkflow composition:
// Sequential[ InitialDraft -> Loop[Writer<->Editor until approved] -> Publisher ]
func NewStoryWriterWorkflow(cfg *config.Config) (*StoryWriterWorkflow, error) {
	// Generate a session ID for memory scoping
	sessionID := fmt.Sprintf("story-session-%d", time.Now().Unix())
	fmt.Printf("\n[DEBUG] Created workflow with sessionID: %s\n", sessionID)

	// Create workflow instance
	sw := &StoryWriterWorkflow{
		config: cfg,
	}

	// Create individual agents
	writer, err := CreateWriter(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create writer: %w", err)
	}

	editor, err := CreateEditor(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create editor: %w", err)
	}

	publisher, err := CreatePublisher(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher: %w", err)
	}

	// Create the Writer<->Editor revision loop as a LoopWorkflow
	// Uses conditional termination to exit when editor approves
	loopWorkflow, err := vnext.NewLoopWorkflowWithCondition(&vnext.WorkflowConfig{
		Mode:          vnext.Loop,
		Timeout:       300 * time.Second,
		MaxIterations: 3,
	}, vnext.Conditions.OutputContains("APPROVED"))
	if err != nil {
		return nil, fmt.Errorf("failed to create loop workflow: %w", err)
	}

	// Add Writer and Editor to the loop with Transform functions
	loopWorkflow.AddStep(vnext.WorkflowStep{
		Name:      "write",
		Agent:     writer,
		Transform: WriterTransform,
	})
	loopWorkflow.AddStep(vnext.WorkflowStep{
		Name:      "review",
		Agent:     editor,
		Transform: EditorTransform,
	}) // Step 3: Wrap the loop as a SubWorkflowAgent
	revisionLoop := vnext.NewSubWorkflowAgent(
		"revision_loop",
		loopWorkflow,
		vnext.WithSubWorkflowDescription("Writer<->Editor revision loop until approval"),
		vnext.WithSubWorkflowMaxDepth(5),
	)

	// Step 4: Create the overall Sequential pipeline
	mainWorkflow, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{
		Mode:    vnext.Sequential,
		Timeout: 300 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create main workflow: %w", err)
	}

	// Add: RevisionLoop (SubWorkflow) -> Publisher
	mainWorkflow.AddStep(vnext.WorkflowStep{
		Name:  "revisions",
		Agent: revisionLoop,
	})
	mainWorkflow.AddStep(vnext.WorkflowStep{
		Name:      "publish",
		Agent:     publisher,
		Transform: PublisherTransform,
	})

	// Wrap the entire pipeline as a SubWorkflowAgent
	storyPipeline := vnext.NewSubWorkflowAgent(
		"story_pipeline",
		mainWorkflow,
		vnext.WithSubWorkflowDescription("Complete story pipeline: Revisions(Loop) -> Publisher"),
		vnext.WithSubWorkflowMaxDepth(10),
	)

	// Populate the workflow struct
	sw.writer = writer
	sw.editor = editor
	sw.publisher = publisher
	sw.revisionLoop = revisionLoop
	sw.storyPipeline = storyPipeline

	return sw, nil
}

func (sw *StoryWriterWorkflow) Name() string {
	return "Story Writer (SubWorkflow Demo)"
}

func (sw *StoryWriterWorkflow) WelcomeMessage() string {
	return "Welcome! Nested SubWorkflow with Smart Loop Demo:\n• Loop exits when Editor approves (or max 3 iterations)\n• Sequential pipeline: [Smart Loop] → Publisher\n• Fully automated workflow composition!\n\nTell me what story you'd like!"
}

func (sw *StoryWriterWorkflow) GetAgents() []infrastructure.AgentInfo {
	return []infrastructure.AgentInfo{
		{Name: "writer", DisplayName: "Writer", Icon: "✍️", Color: "blue", Description: "Creates & revises draft"},
		{Name: "editor", DisplayName: "Editor", Icon: "✏️", Color: "green", Description: "Reviews & approves"},
		{Name: "publisher", DisplayName: "Publisher", Icon: "📚", Color: "purple", Description: "Formats final story"},
	}
}

func (sw *StoryWriterWorkflow) Execute(ctx context.Context, userPrompt string, sendMessage infrastructure.MessageSender) error {
	// Store message sender
	sw.messageSender = sendMessage

	sendMessage(infrastructure.WSMessage{
		Type:      infrastructure.MsgTypeWorkflowStart,
		Content:   "Starting story pipeline with smart loop exit...",
		Timestamp: float64(time.Now().Unix()),
	})

	// Execute the story pipeline SubWorkflow with streaming
	stream, err := sw.storyPipeline.RunStream(ctx, userPrompt)
	if err != nil {
		return err
	}

	// Create stream handler to process chunks
	handler := NewStreamHandler(sendMessage)

	// Process all chunks
	for chunk := range stream.Chunks() {
		switch chunk.Type {
		case vnext.ChunkTypeAgentStart:
			handler.HandleAgentStart(chunk)
		case vnext.ChunkTypeText, vnext.ChunkTypeDelta:
			handler.HandleContent(chunk)
		case vnext.ChunkTypeAgentComplete:
			handler.HandleAgentComplete(chunk)
		case vnext.ChunkTypeMetadata:
			handler.HandleMetadata(chunk)
		case vnext.ChunkTypeError:
			handler.HandleError(chunk)
		}
	}

	// Wait for stream to complete
	result, err := stream.Wait()
	if err != nil {
		return err
	}

	// Get final content
	finalContent := handler.GetFinalContent()
	if finalContent == "" {
		finalContent = result.Content
	}

	// Send workflow completion
	sendMessage(infrastructure.WSMessage{
		Type:      infrastructure.MsgTypeWorkflowDone,
		Content:   finalContent,
		Timestamp: float64(time.Now().Unix()),
		Metadata: map[string]interface{}{
			"success":      true,
			"total_tokens": result.TokensUsed,
		},
	})

	return nil
}

func (sw *StoryWriterWorkflow) Cleanup(ctx context.Context) error {
	// Cleanup individual agents
	if sw.writer != nil {
		sw.writer.Cleanup(ctx)
	}
	if sw.editor != nil {
		sw.editor.Cleanup(ctx)
	}
	if sw.publisher != nil {
		sw.publisher.Cleanup(ctx)
	}
	// Cleanup SubWorkflowAgents (they will cleanup wrapped workflows)
	if sw.revisionLoop != nil {
		sw.revisionLoop.Cleanup(ctx)
	}
	if sw.storyPipeline != nil {
		sw.storyPipeline.Cleanup(ctx)
	}
	return nil
}



