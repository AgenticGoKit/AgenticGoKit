package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/factory"
	"kunalkushwaha/agentflow/internal/llm"
	"kunalkushwaha/agentflow/examples/factory_example/simple_agent"
)

func main() {
	llmBuilder := factory.NewAzureLLMBuilder()

	llmBuilder.FromEnvironment()

	llmProvider, err := llmBuilder.Build()
	if err != nil {
		log.Fatalf("Failed to create Azure LLM provider: %v", err)
	}

	llmAgentBuilder := factory.NewLLMAgentBuilder(llmProvider)
	llmAgentBuilder2 := factory.NewLLMAgentBuilder(llmProvider)

	llmAgentBuilder.WithSystemPrompt("You are a helpful assistant that responds concisely.")
	llmAgentBuilder2.WithSystemPrompt("You are a summarization agent that summarizes the provided text.")

	llmAgentHandler1 := llmAgentBuilder.Build()
	llmAgentHandler2 := llmAgentBuilder2.Build()

	simpleAgentHandler := simple_agent.NewSimpleAgentBuilder().Build()

	// 8. Create a Runner Builder
	runnerBuilder := factory.NewRunnerBuilder()

	// 9. Enable trace logging
	runnerBuilder.WithTraceLogging()

	// 10. Build the Runner
	runner, err := runnerBuilder.Build()
	if err != nil {
		log.Fatalf("Failed to create runner: %v", err)
	}

	runner.Start()

	log.Println("--- Running Sequential Workflow ---")
	sequentialWorkflow := factory.NewSequentialAgentHandler(
		llmAgentHandler1,
		llmAgentHandler2,
	)
	sequentialRunner := factory.NewRunnerBuilder().
		WithTraceLogging().
		RegisterAgent("sequentialAgent", sequentialWorkflow).
		BuildOrPanic()
	
	sequentialRunner.Start()
	sequentialEvent := agentflow.NewEvent("sequentialEvent").
		SetData("query", "Tell me about the history of the internet.").
		SetMetadataValue(agentflow.SessionIDKey, "sessionSequential")
	if err := sequentialRunner.Emit(sequentialEvent); err != nil {
		log.Printf("Error emitting sequential event: %v", err)
	}
	time.Sleep(5 * time.Second) // Allow time for processing
	sequentialRunner.Stop()
	log.Println("--- Sequential Workflow Finished ---")
	fmt.Println()

	log.Println("--- Running Parallel Workflow ---")
	parallelWorkflow := factory.NewParallelAgentHandler(
		llmAgentHandler1,
		llmAgentHandler2,
	)
	parallelRunner := factory.NewRunnerBuilder().
		WithTraceLogging().
		RegisterAgent("parallelAgent", parallelWorkflow).
		BuildOrPanic()

	parallelRunner.Start()
	parallelEvent := agentflow.NewEvent("parallelEvent").
		SetData("query1", "What are the key benefits of cloud computing?").
		SetData("query2", "Explain the concept of serverless architecture simply.").
		SetMetadataValue(agentflow.SessionIDKey, "sessionParallel")
	if err := parallelRunner.Emit(parallelEvent); err != nil {
		log.Printf("Error emitting parallel event: %v", err)
	}
	time.Sleep(5 * time.Second) // Allow time for processing
	parallelRunner.Stop()
	log.Println("--- Parallel Workflow Finished ---")
	fmt.Println()

	log.Println("--- Running Loop Workflow ---")
	loopConditionHandler := NewLoopConditionAgent() // We'll create this simple agent
	// Need to use llmAgentHandler1 because llmAgentHandler2 is used in the loop handler
	// and the Run method is blocking, so if we use the same agent handler, it will
	// block the runner.
	llmAgentHandler3 := factory.NewLLMAgentBuilder(llmProvider).\
	WithSystemPrompt("You are a helpful assistant that responds concisely.").Build()
	loopHandler := factory.NewSequentialAgentHandler( // Loop handler is a sequence of increment and LLM call
		NewIncrementCounterAgent(), // We'll create this simple agent
		llmAgentHandler3, // Summarize the state after incrementing
	)

	loopWorkflow := factory.NewLoopAgentHandler(loopConditionHandler, loopHandler)

	loopRunner := factory.NewRunnerBuilder().
		WithTraceLogging().\
		RegisterAgent("loopAgent", loopWorkflow).
		BuildOrPanic()
	time.Sleep(5 * time.Second) // Allow time for processing
	log.Println("--- Loop Workflow Finished ---")

	loopRunner.Start()
	loopEvent := agentflow.NewEvent("startLoop").
		SetData("counter", 0).
		SetData("query", "Briefly explain the concept of artificial intelligence.").
		SetMetadataValue(agentflow.SessionIDKey, "sessionLoop")
	if err := loopRunner.Emit(loopEvent); err != nil {
		log.Printf("Error emitting loop event: %v", err)
	}
	time.Sleep(10 * time.Second) // Allow time for processing
	loopRunner.Stop()
	runner.Stop()
}