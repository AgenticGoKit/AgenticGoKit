package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
)

// BenchmarkResult stores performance metrics
type BenchmarkResult struct {
	TestName        string
	Duration        time.Duration
	TokensProcessed int
	TokensPerSecond float64
	ChunksProcessed int
	ChunksPerSecond float64
	TotalBytes      int
	BytesPerSecond  float64
	Success         bool
	ErrorCount      int
}

// StreamingMetrics tracks streaming performance
type StreamingMetrics struct {
	mu             sync.Mutex
	ChunkCount     int
	TokenCount     int
	BytesProcessed int
	FirstTokenTime time.Time
	LastTokenTime  time.Time
	StartTime      time.Time
	Errors         int
}

func (sm *StreamingMetrics) AddChunk(chunk *vnext.StreamChunk) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.ChunkCount++

	if chunk.Error != nil {
		sm.Errors++
		return
	}

	var content string
	if chunk.Type == vnext.ChunkTypeDelta {
		content = chunk.Delta
	} else if chunk.Type == vnext.ChunkTypeText {
		content = chunk.Content
	}

	if content != "" {
		sm.BytesProcessed += len(content)
		sm.TokenCount++

		now := time.Now()
		if sm.FirstTokenTime.IsZero() {
			sm.FirstTokenTime = now
		}
		sm.LastTokenTime = now
	}
}

func (sm *StreamingMetrics) GetMetrics() (int, int, int, time.Duration) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	var duration time.Duration
	if !sm.FirstTokenTime.IsZero() && !sm.LastTokenTime.IsZero() {
		duration = sm.LastTokenTime.Sub(sm.FirstTokenTime)
	}

	return sm.ChunkCount, sm.TokenCount, sm.BytesProcessed, duration
}

// Create a test agent for benchmarking
func createTestAgent(name string) (vnext.Agent, error) {
	return vnext.QuickChatAgentWithConfig(name, &vnext.Config{
		Name:         name,
		SystemPrompt: "You are a test agent. Respond with exactly 50 words about the given topic.",
		Timeout:      800 * time.Second, // Increased to 13.3 minutes for thorough benchmarking
		LLM: vnext.LLMConfig{
			Provider:    "ollama",
			Model:       "gemma3:1b",
			Temperature: 0.5,
			MaxTokens:   100, // Limited for consistent comparison
			BaseURL:     "http://localhost:11434",
		},
	})
}

// Benchmark 1: Direct Agent Streaming
func benchmarkDirectAgentStreaming(prompt string) BenchmarkResult {
	fmt.Println("\n📊 Benchmarking: Direct Agent Streaming")
	fmt.Println("─────────────────────────────────────")

	agent, err := createTestAgent("DirectAgent")
	if err != nil {
		return BenchmarkResult{TestName: "Direct Agent Streaming", Success: false}
	}

	metrics := &StreamingMetrics{StartTime: time.Now()}

	ctx := context.Background()
	stream, err := agent.RunStream(ctx, prompt)
	if err != nil {
		return BenchmarkResult{TestName: "Direct Agent Streaming", Success: false}
	}

	startTime := time.Now()
	var fullResponse string

	for chunk := range stream.Chunks() {
		metrics.AddChunk(chunk)

		if chunk.Type == vnext.ChunkTypeDelta {
			fullResponse += chunk.Delta
		} else if chunk.Type == vnext.ChunkTypeText {
			fullResponse += chunk.Content
		}
	}

	totalDuration := time.Since(startTime)
	result, err := stream.Wait()

	chunks, tokens, bytes, _ := metrics.GetMetrics()

	return BenchmarkResult{
		TestName:        "Direct Agent Streaming",
		Duration:        totalDuration,
		TokensProcessed: tokens,
		TokensPerSecond: float64(tokens) / totalDuration.Seconds(),
		ChunksProcessed: chunks,
		ChunksPerSecond: float64(chunks) / totalDuration.Seconds(),
		TotalBytes:      bytes,
		BytesPerSecond:  float64(bytes) / totalDuration.Seconds(),
		Success:         err == nil && result != nil && result.Success,
		ErrorCount:      metrics.Errors,
	}
}

// Benchmark 2: Direct Agent Non-Streaming
func benchmarkDirectAgentNonStreaming(prompt string) BenchmarkResult {
	fmt.Println("\n📊 Benchmarking: Direct Agent Non-Streaming")
	fmt.Println("──────────────────────────────────────────")

	agent, err := createTestAgent("DirectAgentSync")
	if err != nil {
		return BenchmarkResult{TestName: "Direct Agent Non-Streaming", Success: false}
	}

	ctx := context.Background()
	startTime := time.Now()

	result, err := agent.Run(ctx, prompt)
	totalDuration := time.Since(startTime)

	if err != nil || result == nil {
		return BenchmarkResult{
			TestName: "Direct Agent Non-Streaming",
			Duration: totalDuration,
			Success:  false,
		}
	}

	return BenchmarkResult{
		TestName:        "Direct Agent Non-Streaming",
		Duration:        totalDuration,
		TokensProcessed: len(strings.Fields(result.Content)), // Rough token estimate
		TokensPerSecond: float64(len(strings.Fields(result.Content))) / totalDuration.Seconds(),
		ChunksProcessed: 1, // Single response
		ChunksPerSecond: 1.0 / totalDuration.Seconds(),
		TotalBytes:      len(result.Content),
		BytesPerSecond:  float64(len(result.Content)) / totalDuration.Seconds(),
		Success:         result.Success,
		ErrorCount:      0,
	}
}

// Benchmark 3: Workflow Streaming (Sequential)
func benchmarkWorkflowStreaming(prompt string) BenchmarkResult {
	fmt.Println("\n📊 Benchmarking: Workflow Streaming (Sequential)")
	fmt.Println("───────────────────────────────────────────────")

	agent1, err := createTestAgent("WorkflowAgent1")
	if err != nil {
		return BenchmarkResult{TestName: "Workflow Streaming", Success: false}
	}

	agent2, err := createTestAgent("WorkflowAgent2")
	if err != nil {
		return BenchmarkResult{TestName: "Workflow Streaming", Success: false}
	}

	workflow, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{
		Mode:    vnext.Sequential,
		Timeout: 800 * time.Second, // Increased to 13.3 minutes for thorough benchmarking
	})
	if err != nil {
		return BenchmarkResult{TestName: "Workflow Streaming", Success: false}
	}

	workflow.AddStep(vnext.WorkflowStep{
		Name:  "step1",
		Agent: agent1,
		Transform: func(input string) string {
			return fmt.Sprintf("First task: %s", input)
		},
	})

	workflow.AddStep(vnext.WorkflowStep{
		Name:  "step2",
		Agent: agent2,
		Transform: func(input string) string {
			return fmt.Sprintf("Second task based on: %s", input)
		},
	})

	ctx := context.Background()
	workflow.Initialize(ctx)

	metrics := &StreamingMetrics{StartTime: time.Now()}
	startTime := time.Now()

	stream, err := workflow.RunStream(ctx, prompt)
	if err != nil {
		return BenchmarkResult{TestName: "Workflow Streaming", Success: false}
	}

	var fullResponse string

	for chunk := range stream.Chunks() {
		metrics.AddChunk(chunk)

		if chunk.Type == vnext.ChunkTypeDelta {
			fullResponse += chunk.Delta
		} else if chunk.Type == vnext.ChunkTypeText {
			fullResponse += chunk.Content
		}
	}

	totalDuration := time.Since(startTime)
	result, err := stream.Wait()

	chunks, tokens, bytes, _ := metrics.GetMetrics()

	workflow.Shutdown(ctx)

	return BenchmarkResult{
		TestName:        "Workflow Streaming",
		Duration:        totalDuration,
		TokensProcessed: tokens,
		TokensPerSecond: float64(tokens) / totalDuration.Seconds(),
		ChunksProcessed: chunks,
		ChunksPerSecond: float64(chunks) / totalDuration.Seconds(),
		TotalBytes:      bytes,
		BytesPerSecond:  float64(bytes) / totalDuration.Seconds(),
		Success:         err == nil && result != nil && result.Success,
		ErrorCount:      metrics.Errors,
	}
}

// Benchmark 4: Workflow Non-Streaming
func benchmarkWorkflowNonStreaming(prompt string) BenchmarkResult {
	fmt.Println("\n📊 Benchmarking: Workflow Non-Streaming")
	fmt.Println("──────────────────────────────────────")

	agent1, err := createTestAgent("WorkflowAgentSync1")
	if err != nil {
		return BenchmarkResult{TestName: "Workflow Non-Streaming", Success: false}
	}

	agent2, err := createTestAgent("WorkflowAgentSync2")
	if err != nil {
		return BenchmarkResult{TestName: "Workflow Non-Streaming", Success: false}
	}

	workflow, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{
		Mode:    vnext.Sequential,
		Timeout: 800 * time.Second, // Increased to 13.3 minutes for thorough benchmarking
	})
	if err != nil {
		return BenchmarkResult{TestName: "Workflow Non-Streaming", Success: false}
	}

	workflow.AddStep(vnext.WorkflowStep{
		Name:  "step1",
		Agent: agent1,
		Transform: func(input string) string {
			return fmt.Sprintf("First task: %s", input)
		},
	})

	workflow.AddStep(vnext.WorkflowStep{
		Name:  "step2",
		Agent: agent2,
		Transform: func(input string) string {
			return fmt.Sprintf("Second task based on: %s", input)
		},
	})

	ctx := context.Background()
	workflow.Initialize(ctx)

	startTime := time.Now()
	result, err := workflow.Run(ctx, prompt)
	totalDuration := time.Since(startTime)

	workflow.Shutdown(ctx)

	if err != nil || result == nil {
		return BenchmarkResult{
			TestName: "Workflow Non-Streaming",
			Duration: totalDuration,
			Success:  false,
		}
	}

	return BenchmarkResult{
		TestName:        "Workflow Non-Streaming",
		Duration:        totalDuration,
		TokensProcessed: len(strings.Fields(result.FinalOutput)),
		TokensPerSecond: float64(len(strings.Fields(result.FinalOutput))) / totalDuration.Seconds(),
		ChunksProcessed: len(result.StepResults),
		ChunksPerSecond: float64(len(result.StepResults)) / totalDuration.Seconds(),
		TotalBytes:      len(result.FinalOutput),
		BytesPerSecond:  float64(len(result.FinalOutput)) / totalDuration.Seconds(),
		Success:         result.Success,
		ErrorCount:      0,
	}
}

// Display benchmark results
func displayBenchmarkResults(results []BenchmarkResult) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🏆 STREAMING OVERHEAD BENCHMARK RESULTS")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("%-30s %-10s %-8s %-12s %-12s %-10s %-10s\n",
		"Test Name", "Success", "Duration", "Tokens/sec", "Chunks/sec", "MB/sec", "Errors")
	fmt.Println(strings.Repeat("-", 80))

	for _, result := range results {
		successIcon := "✅"
		if !result.Success {
			successIcon = "❌"
		}

		mbPerSec := float64(result.TotalBytes) / (1024 * 1024) / result.Duration.Seconds()

		fmt.Printf("%-30s %-10s %-8.2fs %-12.1f %-12.1f %-10.3f %-10d\n",
			result.TestName,
			successIcon,
			result.Duration.Seconds(),
			result.TokensPerSecond,
			result.ChunksPerSecond,
			mbPerSec,
			result.ErrorCount)
	}

	fmt.Println(strings.Repeat("-", 80))

	// Calculate overhead percentages
	if len(results) >= 4 {
		directStream := results[0]
		directNonStream := results[1]
		workflowStream := results[2]
		workflowNonStream := results[3]

		streamingOverhead := ((directStream.Duration.Seconds() - directNonStream.Duration.Seconds()) / directNonStream.Duration.Seconds()) * 100
		workflowOverhead := ((workflowStream.Duration.Seconds() - directStream.Duration.Seconds()) / directStream.Duration.Seconds()) * 100
		totalOverhead := ((workflowStream.Duration.Seconds() - directNonStream.Duration.Seconds()) / directNonStream.Duration.Seconds()) * 100

		fmt.Println("\n📊 OVERHEAD ANALYSIS:")
		fmt.Printf("• Direct Streaming Overhead:     %+.1f%% vs Non-Streaming\n", streamingOverhead)
		fmt.Printf("• Workflow Overhead:             %+.1f%% vs Direct Streaming\n", workflowOverhead)
		fmt.Printf("• Total Workflow+Stream Overhead: %+.1f%% vs Direct Non-Streaming\n", totalOverhead)

		// Additional comparative analysis
		nonStreamWorkflowOverhead := ((workflowNonStream.Duration.Seconds() - directNonStream.Duration.Seconds()) / directNonStream.Duration.Seconds()) * 100
		fmt.Printf("• Non-Stream Workflow Overhead:  %+.1f%% vs Direct Non-Streaming\n", nonStreamWorkflowOverhead)

		if workflowOverhead < 10 {
			fmt.Println("✅ Workflow streaming overhead is minimal (<10%)")
		} else if workflowOverhead < 25 {
			fmt.Println("⚠️  Workflow streaming has moderate overhead (10-25%)")
		} else {
			fmt.Println("🚨 Workflow streaming has significant overhead (>25%)")
		}
	}
}

func main() {
	fmt.Println("🚀 Streaming Overhead Benchmark Suite")
	fmt.Println("=====================================")
	fmt.Println("Testing performance characteristics of different execution modes...")

	// Quick connection test
	fmt.Println("\n🔍 Testing Ollama connection...")
	testAgent, err := createTestAgent("TestConnection")
	if err != nil {
		log.Fatalf("Failed to create test agent: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Increased connection test timeout
	defer cancel()

	_, err = testAgent.Run(ctx, "Hello")
	if err != nil {
		log.Fatalf("Connection test failed: %v", err)
	}
	fmt.Println("✅ Connection successful")

	// Test prompt - consistent across all tests
	testPrompt := "Explain the concept of artificial intelligence"

	fmt.Println("\n⏱️  Running benchmarks (this may take a few minutes)...")

	results := make([]BenchmarkResult, 0, 4)

	// Run all benchmarks
	results = append(results, benchmarkDirectAgentStreaming(testPrompt))
	results = append(results, benchmarkDirectAgentNonStreaming(testPrompt))
	results = append(results, benchmarkWorkflowStreaming(testPrompt))
	results = append(results, benchmarkWorkflowNonStreaming(testPrompt))

	// Display results
	displayBenchmarkResults(results)

	fmt.Println("\n🎯 Key Insights:")
	fmt.Println("• Streaming provides real-time feedback but may add latency")
	fmt.Println("• Workflow adds coordination overhead but enables multi-agent patterns")
	fmt.Println("• Combined workflow+streaming should have acceptable overhead for UX benefits")
	fmt.Println("• Results depend on LLM response speed, network latency, and system resources")

	fmt.Println("\n✨ Benchmark complete!")
}
