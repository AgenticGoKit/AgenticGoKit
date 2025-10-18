package vnext

import (
	"context"
	"fmt"
	"strings"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// =============================================================================
// MEMORY AND RAG HELPER FUNCTIONS
// =============================================================================
// This file contains helper functions for enriching prompts with memory context,
// building RAG contexts, and formatting prompts for LLM calls.

// EnrichWithMemory enriches user input with relevant memory context.
// It queries the memory provider for relevant memories and formats them
// into a context-enriched prompt based on the memory configuration.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - memoryProvider: The memory provider to query
//   - input: The user's input/query
//   - config: Memory configuration with RAG settings
//
// Returns the enriched input string with memory context prepended, or the
// original input if no relevant memories are found or an error occurs.
func EnrichWithMemory(ctx context.Context, memoryProvider core.Memory, input string, config *MemoryConfig) string {
	if memoryProvider == nil || config == nil {
		return input
	}

	// Determine how many memories to retrieve
	limit := 5 // Default
	if config.RAG != nil && config.RAG.HistoryLimit > 0 {
		limit = config.RAG.HistoryLimit
	}

	// Query memory for relevant context
	memories, err := memoryProvider.Query(ctx, input, limit)
	if err != nil {
		core.Logger().Warn().Err(err).Msg("Failed to query memory for context")
		return input
	}

	// If no memories found, return original input
	if len(memories) == 0 {
		return input
	}

	// Build RAG context if configured
	if config.RAG != nil {
		return BuildRAGContext(memories, config.RAG, input)
	}

	// Fallback: simple context formatting
	return BuildMemorySimpleContext(memories, input)
}

// buildRAGContext builds a RAG-enhanced context from memories using the provided RAG configuration.
// This function formats memories into a structured context suitable for LLM prompts,
// respecting token limits and weighting preferences.
//
// Parameters:
//   - memories: Slice of memory results to include in context
//   - ragConfig: RAG configuration with weights and limits
//   - query: The original user query
//
// Returns a formatted string with RAG context followed by the user query.
func BuildRAGContext(memories []core.Result, ragConfig *RAGConfig, query string) string {
	if len(memories) == 0 {
		return query
	}

	var context strings.Builder

	// Add context header
	context.WriteString("# Relevant Context\n\n")

	// Calculate max tokens for context (if specified)
	maxTokens := ragConfig.MaxTokens
	if maxTokens == 0 {
		maxTokens = 2000 // Default max tokens for context
	}

	currentTokens := 0
	includedMemories := 0

	// Add memories to context (ordered by relevance/score)
	for i, mem := range memories {
		// Estimate tokens (rough approximation: 1 token ≈ 4 characters)
		memTokens := EstimateTokens(mem.Content)

		// Check if adding this memory would exceed token limit
		if currentTokens+memTokens > maxTokens {
			core.Logger().Debug().
				Int("included_memories", includedMemories).
				Int("skipped_memories", len(memories)-includedMemories).
				Msg("Reached token limit for RAG context")
			break
		}

		// Format memory with metadata
		context.WriteString(fmt.Sprintf("## Memory %d (Relevance: %.2f)\n", i+1, mem.Score))
		context.WriteString(mem.Content)
		context.WriteString("\n\n")

		// Add tags if present
		if len(mem.Tags) > 0 {
			context.WriteString(fmt.Sprintf("*Tags: %s*\n\n", strings.Join(mem.Tags, ", ")))
		}

		currentTokens += memTokens
		includedMemories++
	}

	// Add separator and query
	context.WriteString("---\n\n")
	context.WriteString("# User Query\n\n")
	context.WriteString(query)

	core.Logger().Debug().
		Int("memories_included", includedMemories).
		Int("estimated_tokens", currentTokens).
		Msg("Built RAG context")

	return context.String()
}

// BuildMemorySimpleContext builds a simple memory context without RAG configuration.
// This is a fallback method that provides basic context formatting.
// Note: This is different from memory.BuildSimpleContext which has a different signature.
//
// Parameters:
//   - memories: Slice of memory results to include
//   - query: The original user query
//
// Returns a formatted string with memory context followed by the user query.
func BuildMemorySimpleContext(memories []core.Result, query string) string {
	if len(memories) == 0 {
		return query
	}

	var context strings.Builder

	// Add context header
	context.WriteString("Relevant previous information:\n\n")

	// Add memories (limit to first 3 to avoid overwhelming the prompt)
	limit := len(memories)
	if limit > 3 {
		limit = 3
	}

	for i := 0; i < limit; i++ {
		context.WriteString(fmt.Sprintf("- %s\n", memories[i].Content))
	}

	// Add separator and query
	context.WriteString("\nCurrent query: ")
	context.WriteString(query)

	return context.String()
}

// =============================================================================
// CHAT HISTORY HELPER FUNCTIONS
// =============================================================================

// buildChatHistoryContext builds a formatted chat history context from memory.
// This retrieves recent chat messages and formats them for inclusion in prompts.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - memoryProvider: The memory provider to query for chat history
//   - historyLimit: Maximum number of messages to include (0 for default)
//
// Returns a formatted string with chat history, or empty string if no history available.
func BuildChatHistoryContext(ctx context.Context, memoryProvider core.Memory, historyLimit int) string {
	if memoryProvider == nil {
		return ""
	}

	// Use default limit if not specified
	if historyLimit == 0 {
		historyLimit = 10
	}

	// Get chat history
	messages, err := memoryProvider.GetHistory(ctx, historyLimit)
	if err != nil || len(messages) == 0 {
		return ""
	}

	var context strings.Builder

	// Format chat history
	context.WriteString("# Previous Conversation\n\n")

	for _, msg := range messages {
		// Format based on role
		switch strings.ToLower(msg.Role) {
		case "user":
			context.WriteString(fmt.Sprintf("**User**: %s\n\n", msg.Content))
		case "assistant":
			context.WriteString(fmt.Sprintf("**Assistant**: %s\n\n", msg.Content))
		case "system":
			context.WriteString(fmt.Sprintf("*System*: %s\n\n", msg.Content))
		default:
			context.WriteString(fmt.Sprintf("**%s**: %s\n\n", msg.Role, msg.Content))
		}
	}

	context.WriteString("---\n\n")

	return context.String()
}

// =============================================================================
// PROMPT BUILDING HELPER FUNCTIONS
// =============================================================================

// buildEnrichedPrompt builds a complete enriched prompt combining system prompt,
// memory context, chat history, and user input.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - systemPrompt: The agent's system prompt
//   - userInput: The user's input/query
//   - memoryProvider: Optional memory provider for context
//   - config: Optional memory configuration
//
// Returns a core.Prompt ready for LLM execution.
func BuildEnrichedPrompt(ctx context.Context, systemPrompt, userInput string, memoryProvider core.Memory, config *MemoryConfig) core.Prompt {
	prompt := core.Prompt{
		System: systemPrompt,
		User:   userInput,
	}

	// If no memory provider or config, return basic prompt
	if memoryProvider == nil || config == nil {
		return prompt
	}

	// Enrich with memory context
	enrichedInput := EnrichWithMemory(ctx, memoryProvider, userInput, config)

	// Optionally add chat history
	if config.RAG != nil && config.RAG.HistoryLimit > 0 {
		chatHistory := BuildChatHistoryContext(ctx, memoryProvider, config.RAG.HistoryLimit)
		if chatHistory != "" {
			// Prepend chat history to enriched input
			enrichedInput = chatHistory + enrichedInput
		}
	}

	prompt.User = enrichedInput

	return prompt
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// estimateTokens estimates the number of tokens in a text string.
// This uses a simple heuristic: approximately 1 token per 4 characters.
// For more accurate token counting, consider using a proper tokenizer.
//
// Parameters:
//   - text: The text to estimate tokens for
//
// Returns the estimated number of tokens.
func EstimateTokens(text string) int {
	// Simple approximation: 1 token ≈ 4 characters
	// This works reasonably well for English text
	// For more accuracy, use a proper tokenizer library
	return (len(text) + 3) / 4 // Round up
}

// truncateToTokenLimit truncates text to fit within a token limit.
// This is useful for ensuring context doesn't exceed model limits.
//
// Parameters:
//   - text: The text to truncate
//   - maxTokens: Maximum number of tokens allowed
//
// Returns the truncated text.
func TruncateToTokenLimit(text string, maxTokens int) string {
	if maxTokens <= 0 {
		return text
	}

	// Estimate current tokens
	currentTokens := EstimateTokens(text)

	// If within limit, return as-is
	if currentTokens <= maxTokens {
		return text
	}

	// Calculate approximate character limit
	// Use 4 characters per token with 10% safety margin
	charLimit := int(float64(maxTokens) * 4.0 * 0.9)

	if charLimit >= len(text) {
		return text
	}

	// Truncate and add ellipsis
	return text[:charLimit] + "..."
}

// formatMetadataForPrompt formats metadata map into a readable string for prompts.
// This is useful for including additional context like sources, timestamps, etc.
//
// Parameters:
//   - metadata: Map of metadata key-value pairs
//
// Returns a formatted string representation.
func FormatMetadataForPrompt(metadata map[string]interface{}) string {
	if len(metadata) == 0 {
		return ""
	}

	var formatted strings.Builder
	formatted.WriteString("Metadata:\n")

	for key, value := range metadata {
		formatted.WriteString(fmt.Sprintf("- %s: %v\n", key, value))
	}

	return formatted.String()
}

// extractSources extracts source information from memory results.
// This is useful for attribution and transparency in RAG systems.
//
// Parameters:
//   - memories: Slice of memory results
//
// Returns a slice of unique source strings.
func ExtractSources(memories []core.Result) []string {
	sources := make(map[string]bool)
	var result []string

	for _, mem := range memories {
		// Extract sources from tags (common pattern: "source:url")
		for _, tag := range mem.Tags {
			if strings.HasPrefix(tag, "source:") {
				source := strings.TrimPrefix(tag, "source:")
				if !sources[source] {
					sources[source] = true
					result = append(result, source)
				}
			}
		}
	}

	return result
}

// =============================================================================
// VALIDATION FUNCTIONS
// =============================================================================

// validateRAGConfig validates RAG configuration and applies defaults.
// This ensures the configuration is valid and has reasonable defaults.
//
// Parameters:
//   - config: RAG configuration to validate
//
// Returns the validated config with defaults applied, or nil if invalid.
func ValidateRAGConfig(config *RAGConfig) *RAGConfig {
	if config == nil {
		return nil
	}

	// Apply defaults
	validated := *config

	if validated.MaxTokens == 0 {
		validated.MaxTokens = 2000 // Default max tokens for RAG context
	}

	if validated.PersonalWeight == 0 {
		validated.PersonalWeight = 0.3 // Default weight for personal memories
	}

	if validated.KnowledgeWeight == 0 {
		validated.KnowledgeWeight = 0.7 // Default weight for knowledge base
	}

	// Normalize weights if they don't sum to 1.0
	totalWeight := validated.PersonalWeight + validated.KnowledgeWeight
	if totalWeight > 0 && totalWeight != 1.0 {
		validated.PersonalWeight /= totalWeight
		validated.KnowledgeWeight /= totalWeight
	}

	if validated.HistoryLimit == 0 {
		validated.HistoryLimit = 10 // Default history limit
	}

	return &validated
}
