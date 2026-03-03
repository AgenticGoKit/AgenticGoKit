package main

import (
	"context"
	"fmt"
	"log"
	"time"

	_ "github.com/agenticgokit/agenticgokit/plugins/llm/foundrylocal" // register foundrylocal provider
	vnext "github.com/agenticgokit/agenticgokit/v1beta"
)

const (
	// Default Foundry Local endpoint — matches DefaultFoundryLocalBaseURL in the adapter.
	foundryBaseURL = "http://127.0.0.1:5273/v1"

	// Change this to the alias or model name reported by `foundry model list`.
	foundryModel = "qwen2.5-coder-14b-instruct-generic-gpu:4"
)

func main() {
	fmt.Println("============================================")
	fmt.Println("  Azure AI Foundry Local - QuickStart Demo")
	fmt.Println("============================================")
	fmt.Println()
	fmt.Printf("Endpoint : %s\n", foundryBaseURL)
	fmt.Printf("Model    : %s\n\n", foundryModel)

	ctx := context.Background()

	// ------------------------------------------------------------------
	// 1. Basic chat (Run)
	// ------------------------------------------------------------------
	fmt.Println("── Part 1: Basic Chat ─────────────────────")

	chatAgent, err := vnext.NewBuilder("foundry-chat").
		WithConfig(&vnext.Config{
			Name:         "foundry-chat",
			SystemPrompt: "You are a helpful assistant. Keep answers short and clear.",
			Timeout:      60 * time.Second,
			LLM: vnext.LLMConfig{
				Provider:    "foundrylocal",
				Model:       foundryModel,
				BaseURL:     foundryBaseURL,
				Temperature: 0.7,
				MaxTokens:   300,
			},
		}).
		Build()
	if err != nil {
		log.Fatalf("Failed to build chat agent: %v", err)
	}

	if err := chatAgent.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize chat agent: %v", err)
	}
	defer chatAgent.Cleanup(ctx)

	questions := []string{
		"What is Azure AI Foundry Local in one sentence?",
		"Name three popular open-source LLMs that can run locally.",
		"What are the main advantages of running LLMs on-device?",
		"Can you write a terraform script to deploy an Azure AI Foundry Local instance?",
	}

	for i, q := range questions {
		fmt.Printf("\n[Q%d] %s\n", i+1, q)

		qCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		result, err := chatAgent.Run(qCtx, q)
		cancel()

		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			continue
		}

		fmt.Printf("[A%d] %s\n", i+1, result.Content)
		fmt.Printf("     (tokens: %d | latency: %v)\n", result.TokensUsed, result.Duration)
	}

	// ------------------------------------------------------------------
	// 2. Streaming chat (RunStream)
	// ------------------------------------------------------------------
	fmt.Println()
	fmt.Println("── Part 2: Streaming Chat ──────────────────")

	streamAgent, err := vnext.NewBuilder("foundry-stream").
		WithConfig(&vnext.Config{
			Name:         "foundry-stream",
			SystemPrompt: "You are a creative writer. Be vivid but concise.",
			Timeout:      90 * time.Second,
			LLM: vnext.LLMConfig{
				Provider:    "foundrylocal",
				Model:       foundryModel,
				BaseURL:     foundryBaseURL,
				Temperature: 0.9,
				MaxTokens:   400,
			},
		}).
		Build()
	if err != nil {
		log.Fatalf("Failed to build stream agent: %v", err)
	}

	if err := streamAgent.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize stream agent: %v", err)
	}
	defer streamAgent.Cleanup(ctx)

	streamPrompt := "Write a very short poem (4 lines) about running AI models locally."
	fmt.Printf("\nPrompt: %s\n\n", streamPrompt)
	fmt.Println("Streaming response:")
	fmt.Println("------------------")

	streamCtx, streamCancel := context.WithTimeout(ctx, 90*time.Second)
	defer streamCancel()

	stream, err := streamAgent.RunStream(streamCtx, streamPrompt)
	if err != nil {
		log.Fatalf("Failed to start stream: %v", err)
	}

	tokenCount := 0
	start := time.Now()

	for chunk := range stream.Chunks() {
		if chunk.Error != nil {
			fmt.Printf("\nStream error: %v\n", chunk.Error)
			break
		}

		switch chunk.Type {
		case vnext.ChunkTypeDelta:
			fmt.Print(chunk.Delta)
			tokenCount++
		case vnext.ChunkTypeDone:
			fmt.Println()
		}
	}

	elapsed := time.Since(start)
	fmt.Println("------------------")
	fmt.Printf("Received %d delta chunks in %v\n", tokenCount, elapsed)

	// ------------------------------------------------------------------
	// Done
	// ------------------------------------------------------------------
	fmt.Println()
	fmt.Println("============================================")
	fmt.Println("  Demo complete!")
	fmt.Println("============================================")
	fmt.Println()
	fmt.Println("Tip: Change `foundryModel` at the top of main.go")
	fmt.Println("     to the model alias shown by `foundry model list`.")
}
