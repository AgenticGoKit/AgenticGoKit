//go:build ignore

// Standalone program: run with `go run weather-tool.go`.

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	_ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
	vnext "github.com/agenticgokit/agenticgokit/v1beta"
)

// WeatherTool is a minimal LangChain-style tool.
type WeatherTool struct{}

func (t *WeatherTool) Name() string { return "check_weather" }
func (t *WeatherTool) Description() string {
	return "Return the weather forecast for the specified location"
}

// JSONSchema provides the schema for native tool calling.
func (t *WeatherTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type":        "string",
				"description": "City name or common abbreviation (e.g., 'sf'→'San Francisco', 'nyc'→'New York', 'la'→'Los Angeles')",
				"examples":    []string{"San Francisco", "sf", "New York", "nyc", "Tokyo"},
			},
		},
		"required": []string{"location"},
	}
}

func (t *WeatherTool) Execute(ctx context.Context, args map[string]interface{}) (*vnext.ToolResult, error) {
	location, ok := args["location"].(string)
	if !ok || location == "" {
		return &vnext.ToolResult{Success: false, Error: "location parameter required"}, fmt.Errorf("location required")
	}

	forecast := fmt.Sprintf("It's always sunny in %s ☀️", location)
	return &vnext.ToolResult{
		Success: true,
		Content: map[string]interface{}{
			"location": location,
			"forecast": forecast,
		},
	}, nil
}

// Register the tool (automatic discovery).
func init() {
	vnext.RegisterInternalTool("check_weather", func() vnext.Tool { return &WeatherTool{} })
}

// No manual extraction or direct-call fallback; rely on native tool calling.

func Weathertool() {
	ctx := context.Background()

	agent, err := vnext.NewBuilder("weather-agent").
		WithConfig(&vnext.Config{
			Name:         "weather-agent",
			SystemPrompt: `You are a helpful assistant. Don't ask followup questions.`,
			LLM: vnext.LLMConfig{
				Provider:    "ollama",
				Model:       "qwen2.5-coder:7b",
				Temperature: 0.0,
				MaxTokens:   150,
			},
			Tools: &vnext.ToolsConfig{ // Ensure tools are attached to the agent
				Enabled: true,
			},
			// Disable memory for this minimal demo to avoid cross-query bleed
			Memory:  &vnext.MemoryConfig{Enabled: false},
			Timeout: 30 * time.Second,
		}).
		WithPreset(vnext.ChatAgent). // Tools auto-available
		Build()
	if err != nil {
		log.Fatal(err)
	}

	// Minimal: rely on registration; no discovery printout.
	res, err := agent.Run(ctx, "will it rain in sf")
	if err != nil {
		fmt.Printf("   error: %v\n\n", err)
		panic(err)
	}
	fmt.Printf("   Assistant: %s\n", res.Content)
	//fmt.Printf("   Tools used: %v\n\n", res)
}

func main() {
	Weathertool()
}
