//go:build ignore

// Standalone program: run with `go run agent-stage-profiler.go`.

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

// AgentStageTimer instruments the agent to measure time in each stage
type AgentStageTimer struct {
	totalStart time.Time
	stages     map[string]time.Duration
}

func NewStageTimer() *AgentStageTimer {
	return &AgentStageTimer{
		stages: make(map[string]time.Duration),
	}
}

func (st *AgentStageTimer) StartTotal() {
	st.totalStart = time.Now()
}

func (st *AgentStageTimer) MeasureStage(name string, fn func() error) error {
	start := time.Now()
	err := fn()
	st.stages[name] = st.stages[name] + time.Since(start)
	return err
}

func (st *AgentStageTimer) Print(title string) {
	total := time.Since(st.totalStart)
	fmt.Printf("\n%s\n", title)
	fmt.Printf("%-40s %12s %8s\n", "Stage", "Duration", "% Total")
	fmt.Printf("%s\n", string(make([]byte, 60, 60)))

	for stageName, duration := range st.stages {
		pct := float64(duration) / float64(total) * 100
		fmt.Printf("%-40s %10v  %6.1f%%\n", stageName, duration, pct)
	}
	fmt.Printf("%-40s %10v  %6.1f%%\n", "TOTAL", total, 100.0)
}

func AgentStageProfiler() {
	flag.Parse()

	ctx := context.Background()

	fmt.Print("=== Go Agent - Stage-by-Stage Profiling ===\n\n")

	// Build agent
	timer := NewStageTimer()

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
	fmt.Printf("Agent built in: %v\n\n", buildDuration)

	cities := []string{"sf", "nyc", "tokyo"}
	if *city != "" {
		cities = []string{*city}
	}

	for i, cityName := range cities {
		fmt.Printf("\n=== Call %d: %s ===\n", i+1, cityName)

		timer = NewStageTimer()
		timer.StartTotal()

		query := fmt.Sprintf("will it rain in %s", cityName)

		// Instrument the Run call
		var result *vnext.Result
		var runErr error

		timer.MeasureStage("Query Preparation", func() error {
			// Just query string building (negligible)
			return nil
		})

		// This is where the agent loop happens
		startRun := time.Now()
		result, runErr = agent.Run(ctx, query)
		runDuration := time.Since(startRun)

		timer.stages["Agent.Run() Total"] = runDuration

		if runErr != nil {
			fmt.Printf("Error: %v\n", runErr)
			continue
		}

		fmt.Printf("\nResponse: %v\n", result.Content)
		fmt.Printf("Response Duration: %v\n", result.Duration)
		fmt.Printf("Tool Calls Made: %d\n\n", len(result.ToolCalls))

		// Print timing breakdown
		timer.Print("Stage Breakdown")

		// Summary for Ollama overhead calculation
		fmt.Printf("\n=== Overhead Analysis ===\n")
		fmt.Printf("Wall-clock time: %.3fs\n", runDuration.Seconds())
		fmt.Printf("Expected Ollama latency: ~2.0s\n")
		fmt.Printf("Framework overhead: %.3fs (%.1f%%)\n",
			(runDuration - 2*time.Second).Seconds(),
			((runDuration - 2*time.Second).Seconds() / runDuration.Seconds() * 100))
	}
}

func main() {
	AgentStageProfiler()
}
