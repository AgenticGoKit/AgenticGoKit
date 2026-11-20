//go:build ignore

// Standalone program: run with `go run weather-profile.go`.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"runtime/trace"
	"time"

	_ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
	vnext "github.com/agenticgokit/agenticgokit/v1beta"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile = flag.String("memprofile", "", "write memory profile to file")
	tracefile  = flag.String("trace", "", "write runtime trace to file")
	verbose    = flag.Bool("v", false, "verbose timing output")
	city       = flag.String("city", "", "city to query (if empty, tests multiple cities)")
	reasoning  = flag.Bool("reasoning", false, "enable agent reasoning loops (default: false for fast path)")
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

func WeatherProfile() {
	flag.Parse()

	// Start CPU profiling if requested
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	// Start runtime trace if requested (helps show two LLM calls in reasoning mode)
	if *tracefile != "" {
		f, err := os.Create(*tracefile)
		if err != nil {
			log.Fatal("could not create trace file: ", err)
		}
		defer f.Close()
		if err := trace.Start(f); err != nil {
			log.Fatal("could not start trace: ", err)
		}
		defer trace.Stop()
	}

	startTime := time.Now()
	if *verbose {
		fmt.Printf("[TIMING] Program start\n")
	}

	ctx := context.Background()

	buildStart := time.Now()
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
		WithPreset(vnext.ChatAgent).                // Tools auto-available
		WithTools(vnext.WithReasoning(*reasoning)). // Apply reasoning toggle
		Build()
	if err != nil {
		log.Fatal(err)
	}
	if *verbose {
		fmt.Printf("[TIMING] Agent build took: %v\n", time.Since(buildStart))
	}

	// Show reasoning mode
	reasoningMode := "disabled (fast path)"
	if *reasoning {
		reasoningMode = "enabled (multi-step reasoning)"
	}
	fmt.Printf("Reasoning mode: %s\n", reasoningMode)

	// Test single city or multiple cities
	var cities []string
	if *city != "" {
		cities = []string{*city}
	} else {
		cities = []string{"sf", "nyc", "tokyo", "london", "paris", "sydney"}
	}
	var totalRunTime time.Duration

	if len(cities) > 1 {
		fmt.Printf("\n=== Testing %d cities ===\n", len(cities))
	}
	for i, cityName := range cities {
		runStart := time.Now()
		query := fmt.Sprintf("will it rain in %s", cityName)
		res, err := agent.Run(ctx, query)
		runDuration := time.Since(runStart)
		totalRunTime += runDuration

		if err != nil {
			fmt.Printf("   [%d] Error for %s: %v\n", i+1, cityName, err)
			continue
		}

		if len(cities) > 1 {
			fmt.Printf("   [%d] %s: %v (took %v)\n", i+1, cityName, res, runDuration)
		} else {
			fmt.Printf("   %s: %v\n", cityName, res)
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("   Total calls: %d\n", len(cities))
	fmt.Printf("   Total time: %v\n", totalRunTime)
	fmt.Printf("   Average per call: %v\n", totalRunTime/time.Duration(len(cities)))

	if *verbose {
		fmt.Printf("[TIMING] Total execution: %v\n", time.Since(startTime))
	}

	// Write memory profile if requested
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

func main() {
	WeatherProfile()
}
