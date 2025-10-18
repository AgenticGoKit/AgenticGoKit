package benchmarks_test

import (
	"context"
	"testing"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
)

// =============================================================================
// BUILDER BENCHMARKS
// =============================================================================

func BenchmarkBuilder_Creation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = vnext.NewBuilder("test-agent")
	}
}

func BenchmarkBuilder_WithPreset(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		builder := vnext.NewBuilder("test-agent")
		_ = builder.WithPreset(vnext.ChatAgent)
	}
}

func BenchmarkBuilder_Build(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		builder := vnext.NewBuilder("test-agent")
		_, _ = builder.Build()
	}
}

func BenchmarkBuilder_BuildWithPreset(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = vnext.NewBuilder("test-agent").
			WithPreset(vnext.ChatAgent).
			Build()
	}
}

func BenchmarkBuilder_Clone(b *testing.B) {
	builder := vnext.NewBuilder("original")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = builder.Clone()
	}
}

func BenchmarkBuilder_ResearchAgent(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		agent, _ := vnext.NewResearchAgent("test")
		_ = agent
	}
}

// =============================================================================
// CONFIG BENCHMARKS
// =============================================================================

func BenchmarkConfig_Creation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = &vnext.Config{
			Name:         "test-agent",
			SystemPrompt: "You are a helpful assistant",
			Timeout:      30 * time.Second,
			LLM: vnext.LLMConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Temperature: 0.7,
				MaxTokens:   2000,
			},
		}
	}
}

func BenchmarkConfig_Validation(b *testing.B) {
	config := &vnext.Config{
		Name:         "test-agent",
		SystemPrompt: "You are a helpful assistant",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   2000,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = vnext.ValidateConfig(config)
	}
}

func BenchmarkConfig_Complex(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = &vnext.Config{
			Name:         "test-agent",
			SystemPrompt: "You are a helpful assistant",
			Timeout:      30 * time.Second,
			DebugMode:    true,
			LLM: vnext.LLMConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Temperature: 0.7,
				MaxTokens:   2000,
			},
			Memory: &vnext.MemoryConfig{
				Provider:   "memory",
				Connection: "local",
				RAG: &vnext.RAGConfig{
					MaxTokens:       1000,
					PersonalWeight:  0.7,
					KnowledgeWeight: 0.3,
				},
			},
			Tools: &vnext.ToolsConfig{
				Enabled: true,
			},
		}
	}
}

// =============================================================================
// STREAMING BENCHMARKS
// =============================================================================

func BenchmarkStream_Creation(b *testing.B) {
	ctx := context.Background()
	metadata := &vnext.StreamMetadata{
		AgentName: "test-agent",
		StartTime: time.Now(),
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = vnext.NewStream(ctx, metadata)
	}
}

func BenchmarkStream_WithOptions(b *testing.B) {
	ctx := context.Background()
	metadata := &vnext.StreamMetadata{
		AgentName: "test-agent",
		StartTime: time.Now(),
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = vnext.NewStream(ctx, metadata,
			vnext.WithBufferSize(1024),
			vnext.WithThoughts(),
		)
	}
}

func BenchmarkStream_Write(b *testing.B) {
	ctx := context.Background()
	metadata := &vnext.StreamMetadata{
		AgentName: "test-agent",
		StartTime: time.Now(),
	}
	_, writer := vnext.NewStream(ctx, metadata)
	chunk := &vnext.StreamChunk{
		Type:      vnext.ChunkTypeText,
		Content:   "test content",
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = writer.Write(chunk)
	}
	writer.Close()
}

func BenchmarkStream_ChunkCreation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = &vnext.StreamChunk{
			Type:      vnext.ChunkTypeText,
			Content:   "test content with some data",
			Timestamp: time.Now(),
		}
	}
}

// =============================================================================
// WORKFLOW BENCHMARKS
// =============================================================================

func BenchmarkWorkflow_Sequential(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = &vnext.WorkflowConfig{
			Mode:          vnext.Sequential,
			Agents:        []string{"agent1", "agent2"},
			Timeout:       30 * time.Second,
			MaxIterations: 10,
		}
	}
}

func BenchmarkWorkflow_Parallel(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = &vnext.WorkflowConfig{
			Mode:          vnext.Parallel,
			Agents:        []string{"agent1", "agent2"},
			Timeout:       30 * time.Second,
			MaxIterations: 10,
		}
	}
}

func BenchmarkWorkflow_DAG(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = &vnext.WorkflowConfig{
			Mode:          vnext.DAG,
			Agents:        []string{"agent1", "agent2", "agent3"},
			Timeout:       30 * time.Second,
			MaxIterations: 10,
		}
	}
}

// =============================================================================
// MEMORY, TOOLS, MCP BENCHMARKS
// =============================================================================

func BenchmarkMemoryConfig_Creation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = &vnext.MemoryConfig{
			Provider:   "memory",
			Connection: "local",
			RAG: &vnext.RAGConfig{
				MaxTokens:       1000,
				PersonalWeight:  0.7,
				KnowledgeWeight: 0.3,
			},
		}
	}
}

func BenchmarkToolsConfig_Creation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = &vnext.ToolsConfig{
			Enabled:       true,
			MaxRetries:    3,
			Timeout:       10 * time.Second,
			RateLimit:     10,
			MaxConcurrent: 5,
		}
	}
}

func BenchmarkMCPConfig_Creation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = &vnext.MCPConfig{
			Enabled:           true,
			Discovery:         true,
			ConnectionTimeout: 30 * time.Second,
			MaxRetries:        3,
			RetryDelay:        1 * time.Second,
		}
	}
}

// =============================================================================
// ERROR HANDLING BENCHMARKS
// =============================================================================

func BenchmarkAgentError_Creation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = &vnext.AgentError{
			Code:      "EXECUTION_ERROR",
			Message:   "Test error message",
			Timestamp: time.Now(),
		}
	}
}

func BenchmarkValidationError_Creation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = &vnext.ValidationError{
			Field:   "name",
			Message: "field is required",
		}
	}
}
