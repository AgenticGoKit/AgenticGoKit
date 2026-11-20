//go:build ignore

// Standalone program: run with `go run agent-reasoning-toggle.go`.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	_ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
	vnext "github.com/agenticgokit/agenticgokit/v1beta"
)

var (
	enableReasoning = flag.Bool("reasoning", false, "Enable agent reasoning/continuation loops")
	city            = flag.String("city", "sf", "City to query")
)

type WeatherTool struct{}

func (t *WeatherTool) Name() string { return "check_weather" }
func (t *WeatherTool) Description() string {
	return "Return the weather forecast for the specified location"
}

func (t *WeatherTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type":        "string",
				"description": "City name",
			},
		},
		"required": []string{"location"},
	}
}

func (t *WeatherTool) Execute(ctx context.Context, args map[string]interface{}) (*vnext.ToolResult, error) {
	location, _ := args["location"].(string)
	forecast := fmt.Sprintf("It's always sunny in %s ☀️", location)
	return &vnext.ToolResult{
		Success: true,
		Content: map[string]interface{}{
			"location": location,
			"forecast": forecast,
		},
	}, nil
}

func init() {
	vnext.RegisterInternalTool("check_weather", func() vnext.Tool { return &WeatherTool{} })
}

func agentReasoning() {
	flag.Parse()

	ctx := context.Background()

	fmt.Print("=== AgenticGOKit Reasoning Toggle Demo ===\n\n")
	fmt.Printf("Reasoning enabled: %v\n", *enableReasoning)
	fmt.Printf("Query: will it rain in %s\n\n", *city)

	// Build agent
	buildStart := time.Now()

	toolOptions := []vnext.ToolOption{}

	// Add reasoning config based on flag
	if *enableReasoning {
		toolOptions = append(toolOptions, vnext.WithReasoningConfig(5, false))
	} else {
		toolOptions = append(toolOptions, vnext.WithReasoning(false))
	}

	builder := vnext.NewBuilder("weather-agent").
		WithConfig(&vnext.Config{
			Name:         "weather-agent",
			SystemPrompt: `You are a helpful weather assistant. Answer questions about weather using the available tools.`,
			LLM: vnext.LLMConfig{
				Provider:    "ollama",
				Model:       "qwen2.5-coder:7b",
				Temperature: 0.0,
				MaxTokens:   150,
			},
			Tools: &vnext.ToolsConfig{
				Enabled: true,
			},
			Memory:  &vnext.MemoryConfig{Enabled: false},
			Timeout: 30 * time.Second,
		}).
		WithPreset(vnext.ChatAgent)

	// Apply tool options (including reasoning config)
	for _, opt := range toolOptions {
		builder = builder.WithTools(opt)
	}

	agent, err := builder.Build()
	if err != nil {
		log.Fatal(err)
	}

	buildDuration := time.Since(buildStart)
	fmt.Printf("Agent built in: %v\n\n", buildDuration)

	// Run query with timing
	fmt.Print("=== Executing Query ===\n\n")

	runStart := time.Now()
	query := fmt.Sprintf("will it rain in %s", *city)
	result, err := agent.Run(ctx, query)
	runDuration := time.Since(runStart)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Total execution time: %.3fs\n", runDuration.Seconds())
	fmt.Printf("Result duration: %v\n", result.Duration)
	fmt.Printf("Tool calls made: %d\n\n", len(result.ToolCalls))

	if len(result.ToolCalls) > 0 {
		fmt.Println("Tool calls:")
		for i, call := range result.ToolCalls {
			fmt.Printf("  [%d] %s: %v\n", i+1, call.Name, call.Success)
		}
		fmt.Println()
	}

	fmt.Printf("Response:\n%s\n\n", result.Content)

	// Analysis
	fmt.Println("=== Performance Analysis ===")
	if *enableReasoning {
		fmt.Println("Reasoning mode: ENABLED (agent makes multiple LLM calls for reasoning)")
		fmt.Println("Expected: Slower (2-4s) but supports complex multi-step reasoning")
	} else {
		fmt.Println("Reasoning mode: DISABLED (agent makes single LLM call, like Python LangChain)")
		fmt.Println("Expected: Faster (~2s) but no multi-step reasoning capability")
	}
	fmt.Printf("Measured: %.3fs\n", runDuration.Seconds())
}

func main() {
	agentReasoning()
}
