package tools_test

import (
	"context"
	"testing"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// MOCK TOOL FOR TESTING
// =============================================================================

type mockCalculatorTool struct{}

func (m *mockCalculatorTool) Name() string {
	return "calculate"
}

func (m *mockCalculatorTool) Description() string {
	return "Performs basic arithmetic calculations. Usage: calculate(expression=\"2+2\")"
}

func (m *mockCalculatorTool) Execute(ctx context.Context, args map[string]interface{}) (*vnext.ToolResult, error) {
	expr, ok := args["expression"].(string)
	if !ok {
		return &vnext.ToolResult{
			Success: false,
			Error:   "missing or invalid 'expression' argument",
		}, nil
	}

	// Simple calculator mock - just return a fixed result for testing
	switch expr {
	case "2+2":
		return &vnext.ToolResult{
			Success: true,
			Content: "4",
		}, nil
	case "10*5":
		return &vnext.ToolResult{
			Success: true,
			Content: "50",
		}, nil
	default:
		return &vnext.ToolResult{
			Success: false,
			Error:   "unsupported expression",
		}, nil
	}
}

type mockSearchTool struct{}

func (m *mockSearchTool) Name() string {
	return "search"
}

func (m *mockSearchTool) Description() string {
	return "Searches for information. Usage: search(query=\"search term\")"
}

func (m *mockSearchTool) Execute(ctx context.Context, args map[string]interface{}) (*vnext.ToolResult, error) {
	query, ok := args["query"].(string)
	if !ok {
		return &vnext.ToolResult{
			Success: false,
			Error:   "missing or invalid 'query' argument",
		}, nil
	}

	return &vnext.ToolResult{
		Success: true,
		Content: "Search results for: " + query,
	}, nil
}

// =============================================================================
// TOOL PARSING TESTS
// =============================================================================

func TestParseToolCalls_FunctionStyle(t *testing.T) {
	t.Run("single function call", func(t *testing.T) {
		content := `Let me calculate that for you.
calculate(expression="2+2")
The result is 4.`

		calls := vnext.ParseToolCalls(content)
		require.Len(t, calls, 1)
		assert.Equal(t, "calculate", calls[0].Name)
		assert.Equal(t, "2+2", calls[0].Arguments["expression"])
	})

	t.Run("multiple function calls", func(t *testing.T) {
		content := `I'll help you with that.
calculate(expression="2+2")
search(query="weather")
Done!`

		calls := vnext.ParseToolCalls(content)
		require.Len(t, calls, 2)

		assert.Equal(t, "calculate", calls[0].Name)
		assert.Equal(t, "2+2", calls[0].Arguments["expression"])

		assert.Equal(t, "search", calls[1].Name)
		assert.Equal(t, "weather", calls[1].Arguments["query"])
	})

	t.Run("function with no arguments", func(t *testing.T) {
		content := `get_time()`

		calls := vnext.ParseToolCalls(content)
		require.Len(t, calls, 1)
		assert.Equal(t, "get_time", calls[0].Name)
		assert.Empty(t, calls[0].Arguments)
	})

	t.Run("function with multiple arguments", func(t *testing.T) {
		content := `create_user(name="John", age="30", role="admin")`

		calls := vnext.ParseToolCalls(content)
		require.Len(t, calls, 1)
		assert.Equal(t, "create_user", calls[0].Name)
		assert.Equal(t, "John", calls[0].Arguments["name"])
		assert.Equal(t, "30", calls[0].Arguments["age"])
		assert.Equal(t, "admin", calls[0].Arguments["role"])
	})
}

func TestParseToolCalls_ActionStyle(t *testing.T) {
	t.Run("ReAct format", func(t *testing.T) {
		content := `Thought: I need to calculate this.
Action: calculate
Action Input: {"expression": "2+2"}
Observation: The result is 4.`

		calls := vnext.ParseToolCalls(content)
		require.Len(t, calls, 1)
		assert.Equal(t, "calculate", calls[0].Name)
		assert.Equal(t, "2+2", calls[0].Arguments["expression"])
	})

	t.Run("multiple actions", func(t *testing.T) {
		content := `Action: search
Action Input: {"query": "weather"}
Action: calculate
Action Input: {"expression": "10*5"}`

		calls := vnext.ParseToolCalls(content)
		require.Len(t, calls, 2)

		assert.Equal(t, "search", calls[0].Name)
		assert.Equal(t, "weather", calls[0].Arguments["query"])

		assert.Equal(t, "calculate", calls[1].Name)
		assert.Equal(t, "10*5", calls[1].Arguments["expression"])
	})
}

func TestParseToolCalls_NoToolCalls(t *testing.T) {
	t.Run("plain text with no tool calls", func(t *testing.T) {
		content := `This is just a regular response with no tool calls.
I'm answering your question directly without using any tools.`

		calls := vnext.ParseToolCalls(content)
		assert.Empty(t, calls)
	})

	t.Run("text with parentheses but not tool calls", func(t *testing.T) {
		content := `The formula (a + b) = c is correct.
You can use it (if needed) for calculations.`

		calls := vnext.ParseToolCalls(content)
		assert.Empty(t, calls)
	})
}

// =============================================================================
// TOOL FORMATTING TESTS
// =============================================================================

func TestFormatToolsForPrompt(t *testing.T) {
	t.Run("formats multiple tools", func(t *testing.T) {
		tools := []vnext.Tool{
			&mockCalculatorTool{},
			&mockSearchTool{},
		}

		formatted := vnext.FormatToolsForPrompt(tools)

		assert.Contains(t, formatted, "calculate")
		assert.Contains(t, formatted, "search")
		assert.Contains(t, formatted, "arithmetic calculations")
		assert.Contains(t, formatted, "Searches for information")
		assert.Contains(t, formatted, "To use a tool")
	})

	t.Run("returns empty for no tools", func(t *testing.T) {
		formatted := vnext.FormatToolsForPrompt([]vnext.Tool{})
		assert.Empty(t, formatted)
	})
}

func TestFormatToolResult(t *testing.T) {
	t.Run("formats successful result", func(t *testing.T) {
		result := &vnext.ToolResult{
			Success: true,
			Content: "42",
		}

		formatted := vnext.FormatToolResult("calculate", result)
		assert.Contains(t, formatted, "calculate")
		assert.Contains(t, formatted, "returned")
		assert.Contains(t, formatted, "42")
	})

	t.Run("formats failed result", func(t *testing.T) {
		result := &vnext.ToolResult{
			Success: false,
			Error:   "invalid expression",
		}

		formatted := vnext.FormatToolResult("calculate", result)
		assert.Contains(t, formatted, "calculate")
		assert.Contains(t, formatted, "failed")
		assert.Contains(t, formatted, "invalid expression")
	})
}



