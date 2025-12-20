// package v1beta provides the next generation streamlined API for AgenticGoKit
package v1beta

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// =============================================================================
// CORE HANDLER TYPES
// =============================================================================

// NOTE: HandlerFunc and Capabilities are defined in builder.go
// This file provides augmentations and composition utilities for handlers

// HandlerMetadata provides information about a handler for debugging and logging
type HandlerMetadata struct {
	Name        string
	Description string
	Tags        []string
	Author      string
	Version     string
}

// HandlerContext wraps execution context with additional handler-specific metadata
type HandlerContext struct {
	context.Context
	Metadata  *HandlerMetadata
	StartTime time.Time
	Variables map[string]interface{}
	mu        sync.RWMutex
}

// NewHandlerContext creates a new handler context from a standard context
func NewHandlerContext(ctx context.Context) *HandlerContext {
	return &HandlerContext{
		Context:   ctx,
		StartTime: time.Now(),
		Variables: make(map[string]interface{}),
	}
}

// Set stores a variable in the handler context (thread-safe)
func (hc *HandlerContext) Set(key string, value interface{}) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.Variables[key] = value
}

// Get retrieves a variable from the handler context (thread-safe)
func (hc *HandlerContext) Get(key string) (interface{}, bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	val, ok := hc.Variables[key]
	return val, ok
}

// Duration returns how long the handler has been running
func (hc *HandlerContext) Duration() time.Duration {
	return time.Since(hc.StartTime)
}

// =============================================================================
// HANDLER AUGMENTATIONS (Pre-built Compositions)
// =============================================================================

// WithToolAugmentation wraps a handler to automatically augment it with tool-calling capabilities
// The wrapped handler can use tools without implementing the tool-calling logic manually
// Tools are called based on patterns in the input and tool descriptions
func WithToolAugmentation(handler HandlerFunc) HandlerFunc {
	return func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
		// First, check if tools are available
		if capabilities.Tools == nil {
			// No tools available, just run the base handler
			return handler(ctx, input, capabilities)
		}

		// Get available tools
		tools := capabilities.Tools.List()
		if len(tools) == 0 {
			// No tools registered, run base handler
			return handler(ctx, input, capabilities)
		}

		// Create a tool-augmented prompt with available tools
		toolDescriptions := make([]string, len(tools))
		for i, tool := range tools {
			toolDescriptions[i] = fmt.Sprintf("- %s: %s", tool.Name, tool.Description)
		}
		toolsPrompt := "Available tools:\n" + strings.Join(toolDescriptions, "\n")

		// Augment capabilities with tool-aware LLM
		augmentedCapabilities := *capabilities
		originalLLM := capabilities.LLM
		augmentedCapabilities.LLM = func(system, user string) (string, error) {
			// Inject tool information into system prompt
			augmentedSystem := system + "\n\n" + toolsPrompt + "\n\nYou can use these tools by responding with: TOOL: <tool_name> ARGS: <json_args>"
			return originalLLM(augmentedSystem, user)
		}

		// Run the base handler with augmented capabilities
		response, err := handler(ctx, input, &augmentedCapabilities)
		if err != nil {
			return "", err
		}

		// Check if response indicates tool usage
		if strings.HasPrefix(response, "TOOL:") {
			// Parse tool call from response
			parts := strings.SplitN(response, "ARGS:", 2)
			if len(parts) == 2 {
				toolName := strings.TrimSpace(strings.TrimPrefix(parts[0], "TOOL:"))
				argsStr := strings.TrimSpace(parts[1])

				// Execute the tool (simplified - would need proper JSON parsing in production)
				result, err := capabilities.Tools.Execute(ctx, toolName, map[string]interface{}{
					"input": argsStr,
				})
				if err != nil {
					return "", fmt.Errorf("tool execution failed: %w", err)
				}

				if result.Success {
					return fmt.Sprintf("%v", result.Content), nil
				}
				return "", fmt.Errorf("tool execution failed: %s", result.Error)
			}
		}

		return response, nil
	}
}

// WithMemoryAugmentation wraps a handler to automatically augment it with memory capabilities
// The wrapped handler automatically stores inputs/outputs in memory and can recall relevant context
func WithMemoryAugmentation(handler HandlerFunc) HandlerFunc {
	return func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
		// First, check if memory is available
		if capabilities.Memory == nil {
			// No memory available, just run the base handler
			return handler(ctx, input, capabilities)
		}

		// Query memory for relevant context
		relevantMemories, err := capabilities.Memory.Query(ctx, input, func(qc *QueryConfig) {
			qc.Limit = 5
			qc.ScoreThreshold = 0.7
		})
		if err != nil {
			// Memory query failed, log but continue
			fmt.Printf("Warning: Memory query failed: %v\n", err)
		}

		// Augment capabilities with memory-aware LLM
		augmentedCapabilities := *capabilities
		originalLLM := capabilities.LLM

		if len(relevantMemories) > 0 {
			// Build context from relevant memories
			contextLines := make([]string, len(relevantMemories))
			for i, mem := range relevantMemories {
				contextLines[i] = fmt.Sprintf("- %s", mem.Content)
			}
			memoryContext := "Relevant context from memory:\n" + strings.Join(contextLines, "\n")

			augmentedCapabilities.LLM = func(system, user string) (string, error) {
				// Inject memory context into system prompt
				augmentedSystem := system + "\n\n" + memoryContext
				return originalLLM(augmentedSystem, user)
			}
		}

		// Run the base handler with augmented capabilities
		response, err := handler(ctx, input, &augmentedCapabilities)
		if err != nil {
			return "", err
		}

		// Store the interaction in memory
		conversationEntry := fmt.Sprintf("User: %s\nAssistant: %s", input, response)
		storeErr := capabilities.Memory.Store(ctx, conversationEntry, func(sc *StoreConfig) {
			sc.ContentType = "conversation"
			sc.Source = "handler"
			if sc.Metadata == nil {
				sc.Metadata = make(map[string]interface{})
			}
			sc.Metadata["timestamp"] = time.Now().Unix()
		})
		if storeErr != nil {
			// Failed to store, log but don't fail the handler
			fmt.Printf("Warning: Failed to store in memory: %v\n", storeErr)
		}

		return response, nil
	}
}

// WithLLMAugmentation wraps a handler to provide enhanced LLM capabilities with retries and error handling
// The wrapped handler gets a more robust LLM function with automatic retry logic
func WithLLMAugmentation(handler HandlerFunc, maxRetries int) HandlerFunc {
	if maxRetries <= 0 {
		maxRetries = 3
	}

	return func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
		if capabilities.LLM == nil {
			return "", fmt.Errorf("no LLM configured")
		}

		// Augment capabilities with retry-enabled LLM
		augmentedCapabilities := *capabilities
		originalLLM := capabilities.LLM

		augmentedCapabilities.LLM = func(system, user string) (string, error) {
			var lastErr error
			for attempt := 1; attempt <= maxRetries; attempt++ {
				response, err := originalLLM(system, user)
				if err == nil {
					return response, nil
				}

				lastErr = err
				if attempt < maxRetries {
					// Exponential backoff
					backoff := time.Duration(attempt*attempt) * time.Second
					select {
					case <-ctx.Done():
						return "", ctx.Err()
					case <-time.After(backoff):
						// Retry
					}
				}
			}
			return "", fmt.Errorf("LLM call failed after %d attempts: %w", maxRetries, lastErr)
		}

		return handler(ctx, input, &augmentedCapabilities)
	}
}

// WithRAGAugmentation wraps a handler to automatically augment it with RAG capabilities
// The wrapped handler queries the knowledge base and injects relevant documents into the context
func WithRAGAugmentation(handler HandlerFunc, knowledgeBase string, topK int) HandlerFunc {
	if topK <= 0 {
		topK = 3
	}

	return func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
		// Check if memory is available for RAG
		if capabilities.Memory == nil {
			return handler(ctx, input, capabilities)
		}

		// Query the knowledge base
		docs, err := capabilities.Memory.Query(ctx, input, func(qc *QueryConfig) {
			qc.Limit = topK
			// Note: Knowledge base filtering would be handled by the memory implementation
		})
		if err != nil {
			// RAG query failed, continue without RAG
			fmt.Printf("Warning: RAG query failed: %v\n", err)
			return handler(ctx, input, capabilities)
		}

		if len(docs) == 0 {
			// No relevant documents found
			return handler(ctx, input, capabilities)
		}

		// Augment capabilities with RAG-aware LLM
		augmentedCapabilities := *capabilities
		originalLLM := capabilities.LLM

		// Build knowledge context from documents
		docLines := make([]string, len(docs))
		for i, doc := range docs {
			docLines[i] = fmt.Sprintf("Document %d:\n%s", i+1, doc.Content)
		}
		knowledgeContext := "Relevant knowledge:\n" + strings.Join(docLines, "\n\n")

		augmentedCapabilities.LLM = func(system, user string) (string, error) {
			// Inject knowledge context into system prompt
			augmentedSystem := system + "\n\n" + knowledgeContext + "\n\nUse the above knowledge to answer the user's question."
			return originalLLM(augmentedSystem, user)
		}

		return handler(ctx, input, &augmentedCapabilities)
	}
}

// =============================================================================
// HANDLER COMPOSITION AND CHAINING
// =============================================================================

// Chain combines multiple handlers in sequence
// Each handler receives the output of the previous handler as input
// Useful for building pipelines of processing steps
func Chain(handlers ...HandlerFunc) HandlerFunc {
	return func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
		result := input
		var err error

		for i, handler := range handlers {
			result, err = handler(ctx, result, capabilities)
			if err != nil {
				return "", fmt.Errorf("handler %d failed: %w", i, err)
			}
		}

		return result, nil
	}
}

// ParallelHandlers runs multiple handlers in parallel and combines their results
// All handlers receive the same input and run concurrently
// Results are concatenated with a separator
func ParallelHandlers(separator string, handlers ...HandlerFunc) HandlerFunc {
	return func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
		results := make([]string, len(handlers))
		errors := make([]error, len(handlers))
		var wg sync.WaitGroup

		for i, handler := range handlers {
			wg.Add(1)
			go func(idx int, h HandlerFunc) {
				defer wg.Done()
				result, err := h(ctx, input, capabilities)
				results[idx] = result
				errors[idx] = err
			}(i, handler)
		}

		wg.Wait()

		// Check for errors
		for i, err := range errors {
			if err != nil {
				return "", fmt.Errorf("handler %d failed: %w", i, err)
			}
		}

		// Combine results
		return strings.Join(results, separator), nil
	}
}

// Conditional runs a handler only if the condition function returns true
// Otherwise, it passes through the input unchanged
func Conditional(condition func(ctx context.Context, input string) bool, handler HandlerFunc) HandlerFunc {
	return func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
		if condition(ctx, input) {
			return handler(ctx, input, capabilities)
		}
		return input, nil
	}
}

// Fallback tries the primary handler first, and if it fails, tries the fallback handler
// Useful for implementing graceful degradation
func Fallback(primary HandlerFunc, fallback HandlerFunc) HandlerFunc {
	return func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
		result, err := primary(ctx, input, capabilities)
		if err != nil {
			// Primary failed, try fallback
			return fallback(ctx, input, capabilities)
		}
		return result, nil
	}
}

// Retry wraps a handler with retry logic
// If the handler fails, it will be retried up to maxRetries times with exponential backoff
func Retry(handler HandlerFunc, maxRetries int) HandlerFunc {
	if maxRetries <= 0 {
		maxRetries = 3
	}

	return func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
		var lastErr error
		for attempt := 1; attempt <= maxRetries; attempt++ {
			result, err := handler(ctx, input, capabilities)
			if err == nil {
				return result, nil
			}

			lastErr = err
			if attempt < maxRetries {
				// Exponential backoff
				backoff := time.Duration(attempt*attempt) * time.Second
				select {
				case <-ctx.Done():
					return "", ctx.Err()
				case <-time.After(backoff):
					// Retry
				}
			}
		}
		return "", fmt.Errorf("handler failed after %d attempts: %w", maxRetries, lastErr)
	}
}

// WithTimeout wraps a handler with a timeout
// If the handler takes longer than the timeout, it returns an error
func WithTimeout(handler HandlerFunc, timeout time.Duration) HandlerFunc {
	return func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		resultCh := make(chan string, 1)
		errCh := make(chan error, 1)

		go func() {
			result, err := handler(ctx, input, capabilities)
			if err != nil {
				errCh <- err
			} else {
				resultCh <- result
			}
		}()

		select {
		case result := <-resultCh:
			return result, nil
		case err := <-errCh:
			return "", err
		case <-ctx.Done():
			return "", fmt.Errorf("handler timeout after %v", timeout)
		}
	}
}

// WithLogging wraps a handler with logging for debugging and monitoring
func WithLogging(handler HandlerFunc, logger func(format string, args ...interface{})) HandlerFunc {
	return func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
		start := time.Now()
		logger("Handler started with input: %s", input)

		result, err := handler(ctx, input, capabilities)

		duration := time.Since(start)
		if err != nil {
			logger("Handler failed after %v: %v", duration, err)
		} else {
			logger("Handler completed in %v", duration)
		}

		return result, err
	}
}

// =============================================================================
// PRE-BUILT HANDLERS
// =============================================================================

// SimplePassthrough is a basic handler that just returns the input
// Useful for testing or as a placeholder
func SimplePassthrough() HandlerFunc {
	return func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
		return input, nil
	}
}

// LLMOnly is a basic handler that calls the LLM with a simple prompt
// Useful for basic chat functionality
func LLMOnly(systemPrompt string) HandlerFunc {
	return func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
		if capabilities.LLM == nil {
			return "", fmt.Errorf("no LLM configured")
		}
		return capabilities.LLM(systemPrompt, input)
	}
}

// ToolsFirst tries to use tools first, falling back to LLM if no tools match
// Useful for tool-heavy agents
func ToolsFirst(systemPrompt string) HandlerFunc {
	toolHandler := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
		if capabilities.Tools == nil {
			return "", fmt.Errorf("no tools available")
		}

		// Try to match input to a tool (simplified logic)
		tools := capabilities.Tools.List()

		// Simple keyword matching (production would use more sophisticated matching)
		for _, tool := range tools {
			if strings.Contains(strings.ToLower(input), strings.ToLower(tool.Name)) {
				result, err := capabilities.Tools.Execute(ctx, tool.Name, map[string]interface{}{
					"input": input,
				})
				if err != nil {
					continue
				}
				if result.Success {
					return fmt.Sprintf("%v", result.Content), nil
				}
			}
		}

		return "", fmt.Errorf("no matching tool found")
	}

	llmHandler := LLMOnly(systemPrompt)
	return Fallback(toolHandler, llmHandler)
}

// =============================================================================
// EXAMPLE USAGE AND DOCUMENTATION
// =============================================================================

/*
Package handlers provides a streamlined system for building custom agent logic.

# Basic Handler

The simplest handler just processes input and returns output:

	handler := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
		return "You said: " + input, nil
	}

	agent := NewBuilder().
		WithLLM("openai", "gpt-4").
		WithHandler(handler).
		Build()

# Handler with LLM

Use the LLM capability to call the configured language model:

	handler := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
		return capabilities.LLM("You are a helpful assistant.", input)
	}

# Handler with Tools

Use the Tools capability to execute tools:

	handler := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
		result, err := capabilities.Tools.ExecuteTool(ctx, "calculator", map[string]interface{}{
			"operation": "add",
			"a": 5,
			"b": 3,
		})
		if err != nil {
			return "", err
		}
		return result.Output, nil
	}

# Handler with Memory

Use the Memory capability to store and query information:

	handler := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
		// Query memory for context
		memories, err := capabilities.Memory.Query(ctx, input, &QueryConfig{Limit: 3})
		if err != nil {
			return "", err
		}

		// Build context from memories
		context := ""
		for _, mem := range memories {
			context += mem.Content + "\n"
		}

		// Call LLM with context
		response, err := capabilities.LLM("Use this context: "+context, input)
		if err != nil {
			return "", err
		}

		// Store response in memory
		capabilities.Memory.Store(ctx, response, &StoreConfig{ContentType: "response"})

		return response, nil
	}

# Pre-built Augmentations

Use pre-built augmentations to add capabilities to handlers:

	// Automatically add tool-calling capability
	handler := WithToolAugmentation(func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
		return capabilities.LLM("You are a helpful assistant.", input)
	})

	// Automatically add memory capability
	handler := WithMemoryAugmentation(func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
		return capabilities.LLM("You are a helpful assistant.", input)
	})

	// Add RAG capability to query knowledge base
	handler := WithRAGAugmentation(
		func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
			return capabilities.LLM("Answer based on the provided knowledge.", input)
		},
		"my-knowledge-base",
		5, // top 5 documents
	)

# Handler Composition

Chain handlers together for complex workflows:

	// Sequential processing pipeline
	pipeline := Chain(
		preprocessHandler,
		mainHandler,
		postprocessHandler,
	)

	// Run handlers in parallel
	combined := Parallel("\n---\n",
		summaryHandler,
		detailHandler,
	)

	// Conditional execution
	smartHandler := Conditional(
		func(ctx context.Context, input string) bool {
			return len(input) > 100 // only for long inputs
		},
		complexHandler,
	)

	// Fallback for error handling
	robustHandler := Fallback(
		primaryHandler,
		backupHandler,
	)

	// Add retry logic
	reliableHandler := Retry(unstableHandler, 3)

	// Add timeout protection
	timeoutHandler := WithTimeout(slowHandler, 30*time.Second)

# Complex Example: Multi-Capability Handler

Combine multiple augmentations and composition patterns:

	handler := Chain(
		WithMemoryAugmentation(preprocessHandler),
		WithRAGAugmentation(
			WithToolAugmentation(mainHandler),
			"company-docs",
			3,
		),
		WithMemoryAugmentation(postprocessHandler),
	)

	agent := NewBuilder().
		WithLLM("openai", "gpt-4").
		WithMemory(&MemoryConfig{Backend: "local"}).
		WithTools(tool1, tool2, tool3).
		WithHandler(handler).
		Build()

	result, err := agent.Run(ctx, "Complex query requiring multiple capabilities")

# Pre-built Handlers

Use pre-built handlers for common patterns:

	// Simple LLM-only handler
	agent := NewBuilder().
		WithLLM("openai", "gpt-4").
		WithHandler(LLMOnly("You are a helpful assistant.")).
		Build()

	// Tool-first handler with LLM fallback
	agent := NewBuilder().
		WithLLM("openai", "gpt-4").
		WithTools(calculator, webSearch).
		WithHandler(ToolsFirst("You are a helpful assistant with tools.")).
		Build()
*/

