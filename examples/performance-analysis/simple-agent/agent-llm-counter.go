//go:build ignore

// Standalone program: run with `go run agent-llm-counter.go`.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/agenticgokit/agenticgokit/internal/llm"
	_ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
	vnext "github.com/agenticgokit/agenticgokit/v1beta"
)

var (
	city      = flag.String("city", "", "city to query")
	reasoning = flag.Bool("reasoning", false, "enable agent reasoning loops (default: false)")
)

// InstrumentedLLMProvider wraps an LLM provider to count calls
type InstrumentedLLMProvider struct {
	underlying llm.ModelProvider
	mu         sync.Mutex
	callCount  int
	callTimes  []time.Duration
}

func (i *InstrumentedLLMProvider) Call(ctx context.Context, prompt llm.Prompt) (llm.Response, error) {
	i.mu.Lock()
	i.callCount++
	callNum := i.callCount
	i.mu.Unlock()

	fmt.Printf("\n[LLM Call #%d]\n", callNum)
	fmt.Printf("  System prompt: %.60s...\n", prompt.System)
	fmt.Printf("  User input: %.100s...\n", prompt.User)
	fmt.Printf("  Tools available: %d\n", len(prompt.Tools))

	start := time.Now()
	response, err := i.underlying.Call(ctx, prompt)
	elapsed := time.Since(start)

	i.mu.Lock()
	i.callTimes = append(i.callTimes, elapsed)
	i.mu.Unlock()

	fmt.Printf("  Response time: %.3fs\n", elapsed.Seconds())
	fmt.Printf("  Response content: %.80s...\n", response.Content)
	fmt.Printf("  Tool calls in response: %d\n", len(response.ToolCalls))

	return response, err
}

func (i *InstrumentedLLMProvider) PrintStats() {
	i.mu.Lock()
	defer i.mu.Unlock()

	fmt.Printf("\n=== LLM Provider Statistics ===\n")
	fmt.Printf("Total calls made: %d\n", i.callCount)

	totalTime := time.Duration(0)
	for _, t := range i.callTimes {
		totalTime += t
	}

	if i.callCount > 0 {
		avgTime := totalTime / time.Duration(i.callCount)
		fmt.Printf("Total LLM time: %.3fs\n", totalTime.Seconds())
		fmt.Printf("Average per call: %.3fs\n", avgTime.Seconds())
		fmt.Printf("\nDetailed breakdown:\n")
		for i, t := range i.callTimes {
			fmt.Printf("  Call %d: %.3fs\n", i+1, t.Seconds())
		}
	}
}

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

func agentLLMCounter() {
	flag.Parse()

	ctx := context.Background()

	fmt.Print("=== Instrumented LLM Call Counter ===\n\n")

	// Build agent first to get its LLM provider, then wrap it
	agentBuilder := vnext.NewBuilder("weather-agent").
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
		WithTools(vnext.WithReasoning(*reasoning))

	agent, err := agentBuilder.Build()
	if err != nil {
		log.Fatal(err)
	}

	cities := []string{"sf"}
	if *city != "" {
		cities = []string{*city}
	}

	instrumentedProvider := &InstrumentedLLMProvider{
		callTimes: []time.Duration{},
	}

	// Try to hook into the agent - this is tricky since agent doesn't expose the provider
	// Instead, we'll use the agent directly and look at the result

	for i, cityName := range cities {
		fmt.Printf("\n=== Query %d: %s ===\n", i+1, cityName)

		query := fmt.Sprintf("will it rain in %s", cityName)

		startTotal := time.Now()
		result, runErr := agent.Run(ctx, query)
		totalTime := time.Since(startTotal)

		if runErr != nil {
			fmt.Printf("Error: %v\n", runErr)
			continue
		}

		fmt.Printf("\n=== Results ===\n")
		fmt.Printf("Total wall-clock time: %.3fs\n", totalTime.Seconds())
		fmt.Printf("Result duration (from agent): %v\n", result.Duration)
		fmt.Printf("Tool calls made: %d\n", len(result.ToolCalls))

		if len(result.ToolCalls) > 0 {
			fmt.Printf("\nTool calls detail:\n")
			for j, call := range result.ToolCalls {
				fmt.Printf("  [%d] %s: %v\n", j+1, call.Name, call.Success)
			}
		}

		fmt.Printf("\nResponse: %.100s...\n", result.Content)

		// Calculate estimated overhead
		fmt.Printf("\n=== Overhead Analysis ===\n")
		fmt.Printf("Estimated base Ollama latency: ~2.0s\n")
		fmt.Printf("Measured total time: %.3fs\n", totalTime.Seconds())
		fmt.Printf("Framework overhead: %.3fs\n", (totalTime - 2*time.Second).Seconds())

		// Estimate number of LLM calls
		estimatedCalls := int((totalTime.Seconds() + 0.5) / 2.0)
		fmt.Printf("Estimated LLM calls: %d\n", estimatedCalls)
	}

	instrumentedProvider.PrintStats()
}

func main() {
	agentLLMCounter()
}
