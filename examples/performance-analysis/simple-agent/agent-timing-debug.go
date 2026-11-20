//go:build ignore

// Standalone program: run with `go run agent-timing-debug.go`.

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
	city = flag.String("city", "", "city to query")
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
	forecast := fmt.Sprintf("It's always sunny in %s", location)
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

func AgentTimingDebug() {
	flag.Parse()

	ctx := context.Background()

	fmt.Print("=== Go Agent - Detailed Timing ===\n\n")

	buildStart := time.Now()
	agent, err := vnext.NewBuilder("weather-agent").
		WithConfig(&vnext.Config{
			Name:         "weather-agent",
			SystemPrompt: `You are a helpful assistant. Don't ask follow-up questions.`,
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
		WithPreset(vnext.ChatAgent).
		Build()
	if err != nil {
		log.Fatal(err)
	}
	buildDuration := time.Since(buildStart)
	fmt.Printf("Agent build: %v\n\n", buildDuration)

	cities := []string{"sf", "nyc", "tokyo"}
	if *city != "" {
		cities = []string{*city}
	}

	for i, cityName := range cities {
		fmt.Printf("Run %d (%s):\n", i+1, cityName)

		stageStart := time.Now()
		query := fmt.Sprintf("will it rain in %s", cityName)
		stageDuration := time.Since(stageStart)
		fmt.Printf("  Query preparation: %v\n", stageDuration)

		runStart := time.Now()
		res, err := agent.Run(ctx, query)
		totalDuration := time.Since(runStart)

		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}

		fmt.Printf("  Total agent.Run(): %v\n", totalDuration)
		fmt.Printf("  Response: %v\n\n", res)
	}
}

func main() {
	AgentTimingDebug()
}
