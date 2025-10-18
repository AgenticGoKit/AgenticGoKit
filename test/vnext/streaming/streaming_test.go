package streaming_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	vnext "github.com/kunalkushwaha/agenticgokit/core/vnext"
)

// TestStreamChunkTypes tests all chunk type constants
func TestStreamChunkTypes(t *testing.T) {
	tests := []struct {
		name      string
		chunkType vnext.ChunkType
		expected  string
	}{
		{"Text chunk", vnext.ChunkTypeText, "text"},
		{"Delta chunk", vnext.ChunkTypeDelta, "delta"},
		{"Thought chunk", vnext.ChunkTypeThought, "thought"},
		{"Tool call chunk", vnext.ChunkTypeToolCall, "tool_call"},
		{"Tool result chunk", vnext.ChunkTypeToolRes, "tool_result"},
		{"Metadata chunk", vnext.ChunkTypeMetadata, "metadata"},
		{"Error chunk", vnext.ChunkTypeError, "error"},
		{"Done chunk", vnext.ChunkTypeDone, "done"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.chunkType) != tt.expected {
				t.Errorf("ChunkType = %v, want %v", tt.chunkType, tt.expected)
			}
		})
	}
}

// TestStreamBuilder tests the StreamBuilder functionality
func TestStreamBuilder(t *testing.T) {
	ctx := context.Background()

	t.Run("basic stream creation", func(t *testing.T) {
		stream, writer := vnext.NewStreamBuilder().Build(ctx)

		if stream == nil {
			t.Fatal("NewStreamBuilder().Build() returned nil stream")
		}
		if writer == nil {
			t.Fatal("NewStreamBuilder().Build() returned nil writer")
		}

		// Stream should be usable
		chunks := stream.Chunks()
		if chunks == nil {
			t.Error("Stream.Chunks() returned nil")
		}

		// Close the writer
		writer.Close()
	})

	t.Run("with metadata", func(t *testing.T) {
		stream, writer := vnext.NewStreamBuilder().
			WithAgentName("TestAgent").
			WithSessionID("session-123").
			WithTraceID("trace-456").
			Build(ctx)

		defer writer.Close()

		if stream.Metadata() == nil {
			t.Error("Stream.Metadata() returned nil")
		}

		if stream.Metadata().AgentName != "TestAgent" {
			t.Errorf("AgentName = %v, want TestAgent", stream.Metadata().AgentName)
		}
		if stream.Metadata().SessionID != "session-123" {
			t.Errorf("SessionID = %v, want session-123", stream.Metadata().SessionID)
		}
	})

	t.Run("custom buffer size", func(t *testing.T) {
		stream, writer := vnext.NewStreamBuilder().
			WithOption(vnext.WithBufferSize(200)).
			Build(ctx)

		if stream == nil {
			t.Fatal("Stream with custom buffer size is nil")
		}

		// Verify we can write and read chunks
		go func() {
			for i := 0; i < 10; i++ {
				writer.Write(&vnext.StreamChunk{
					Type:    vnext.ChunkTypeDelta,
					Delta:   "test",
					Content: "test content",
				})
			}
			writer.Close()
		}()

		count := 0
		for range stream.Chunks() {
			count++
		}

		if count != 10 {
			t.Errorf("Expected 10 chunks, got %d", count)
		}
	})
}

// TestStreamWriteAndRead tests writing and reading chunks
func TestStreamWriteAndRead(t *testing.T) {
	ctx := context.Background()

	t.Run("write and read single chunk", func(t *testing.T) {
		stream, writer := vnext.NewStreamBuilder().Build(ctx)

		expectedChunk := &vnext.StreamChunk{
			Type:    vnext.ChunkTypeText,
			Content: "Hello, World!",
		}

		go func() {
			writer.Write(expectedChunk)
			writer.Close()
		}()

		chunks := stream.Chunks()
		chunk := <-chunks

		if chunk.Type != expectedChunk.Type {
			t.Errorf("Type = %v, want %v", chunk.Type, expectedChunk.Type)
		}
		if chunk.Content != expectedChunk.Content {
			t.Errorf("Content = %v, want %v", chunk.Content, expectedChunk.Content)
		}
	})

	t.Run("write multiple chunks", func(t *testing.T) {
		stream, writer := vnext.NewStreamBuilder().
			WithOption(vnext.WithBufferSize(100)).
			Build(ctx)

		testChunks := []vnext.StreamChunk{
			{Type: vnext.ChunkTypeDelta, Delta: "Hello"},
			{Type: vnext.ChunkTypeDelta, Delta: " "},
			{Type: vnext.ChunkTypeDelta, Delta: "World"},
			{Type: vnext.ChunkTypeDone, Content: "Complete"},
		}

		go func() {
			for _, chunk := range testChunks {
				c := chunk // Create copy
				writer.Write(&c)
			}
			writer.Close()
		}()

		received := []vnext.StreamChunk{}
		for chunk := range stream.Chunks() {
			received = append(received, *chunk)
		}

		if len(received) != len(testChunks) {
			t.Errorf("Expected %d chunks, got %d", len(testChunks), len(received))
		}

		for i, chunk := range received {
			if chunk.Type != testChunks[i].Type {
				t.Errorf("Chunk %d: Type = %v, want %v", i, chunk.Type, testChunks[i].Type)
			}
		}
	})

	t.Run("concurrent writes", func(t *testing.T) {
		stream, writer := vnext.NewStreamBuilder().
			WithOption(vnext.WithBufferSize(200)).
			Build(ctx)

		numGoroutines := 10
		chunksPerGoroutine := 5
		var wg sync.WaitGroup

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < chunksPerGoroutine; j++ {
					writer.Write(&vnext.StreamChunk{
						Type:    vnext.ChunkTypeDelta,
						Delta:   "test",
						Content: "concurrent write",
						Index:   id*chunksPerGoroutine + j,
					})
				}
			}(i)
		}

		go func() {
			wg.Wait()
			writer.Close()
		}()

		count := 0
		for range stream.Chunks() {
			count++
		}

		expected := numGoroutines * chunksPerGoroutine
		if count != expected {
			t.Errorf("Expected %d chunks, got %d", expected, count)
		}
	})
}

// TestStreamCancellation tests stream cancellation
func TestStreamCancellation(t *testing.T) {
	ctx := context.Background()

	t.Run("cancel stream", func(t *testing.T) {
		stream, writer := vnext.NewStreamBuilder().Build(ctx)

		go func() {
			for i := 0; i < 100; i++ {
				writer.Write(&vnext.StreamChunk{
					Type:  vnext.ChunkTypeDelta,
					Delta: "test",
				})
				time.Sleep(10 * time.Millisecond)
			}
			writer.Close()
		}()

		// Read a few chunks then cancel
		count := 0
		for range stream.Chunks() {
			count++
			if count >= 5 {
				stream.Cancel()
				break
			}
		}

		if count < 5 {
			t.Errorf("Expected at least 5 chunks before cancel, got %d", count)
		}
	})

	t.Run("cancel before reading", func(t *testing.T) {
		stream, writer := vnext.NewStreamBuilder().Build(ctx)

		// Cancel immediately
		stream.Cancel()

		// Close writer
		writer.Close()

		// Should be able to safely cancel - may or may not return error depending on timing
		_, _ = stream.Wait()
		// Test passes if no panic occurs
	})
}

// TestStreamWait tests the Wait functionality
func TestStreamWait(t *testing.T) {
	ctx := context.Background()

	t.Run("wait for completion", func(t *testing.T) {
		stream, writer := vnext.NewStreamBuilder().Build(ctx)

		go func() {
			writer.Write(&vnext.StreamChunk{
				Type:    vnext.ChunkTypeDelta,
				Delta:   "test",
				Content: "test content",
			})
			writer.Close()
		}()

		// Consume chunks
		for range stream.Chunks() {
		}

		result, err := stream.Wait()
		if err != nil {
			t.Errorf("Wait() error = %v", err)
		}

		// Result may be nil if no result was set - this is valid behavior
		// The important part is that Wait() completes without error
		_ = result
	})

	t.Run("wait with error", func(t *testing.T) {
		stream, writer := vnext.NewStreamBuilder().Build(ctx)

		expectedError := errors.New("test error")

		go func() {
			writer.CloseWithError(expectedError)
		}()

		// Consume chunks
		for range stream.Chunks() {
		}

		_, err := stream.Wait()
		if err == nil {
			t.Error("Expected error from Wait(), got nil")
		}
	})
}

// TestStreamMetadata tests metadata handling
func TestStreamMetadata(t *testing.T) {
	ctx := context.Background()

	t.Run("basic metadata", func(t *testing.T) {
		stream, writer := vnext.NewStreamBuilder().
			WithAgentName("TestAgent").
			WithSessionID("session-123").
			WithTraceID("trace-456").
			WithModel("gpt-4").
			Build(ctx)

		defer writer.Close()

		retrieved := stream.Metadata()
		if retrieved == nil {
			t.Fatal("Metadata() returned nil")
		}

		if retrieved.AgentName != "TestAgent" {
			t.Errorf("AgentName = %v, want TestAgent", retrieved.AgentName)
		}
		if retrieved.SessionID != "session-123" {
			t.Errorf("SessionID = %v, want session-123", retrieved.SessionID)
		}
		if retrieved.Model != "gpt-4" {
			t.Errorf("Model = %v, want gpt-4", retrieved.Model)
		}
	})
}

// TestStreamAsReader tests io.Reader functionality
func TestStreamAsReader(t *testing.T) {
	ctx := context.Background()

	t.Run("read as io.Reader", func(t *testing.T) {
		stream, writer := vnext.NewStreamBuilder().Build(ctx)

		testText := "Hello, World! This is a test."

		go func() {
			words := strings.Split(testText, " ")
			for _, word := range words {
				writer.Write(&vnext.StreamChunk{
					Type:  vnext.ChunkTypeDelta,
					Delta: word + " ",
				})
			}
			writer.Close()
		}()

		reader := stream.AsReader()
		if reader == nil {
			t.Fatal("AsReader() returned nil")
		}

		data, err := io.ReadAll(reader)
		if err != nil {
			t.Errorf("ReadAll() error = %v", err)
		}

		result := strings.TrimSpace(string(data))
		expected := strings.TrimSpace(testText)
		if result != expected {
			t.Errorf("Read data = %q, want %q", result, expected)
		}
	})
}

// TestCollectStream tests the CollectStream utility
func TestCollectStream(t *testing.T) {
	ctx := context.Background()

	t.Run("collect text chunks", func(t *testing.T) {
		stream, writer := vnext.NewStreamBuilder().Build(ctx)

		expectedText := "Hello World Test"

		go func() {
			writer.Write(&vnext.StreamChunk{Type: vnext.ChunkTypeDelta, Delta: "Hello "})
			writer.Write(&vnext.StreamChunk{Type: vnext.ChunkTypeDelta, Delta: "World "})
			writer.Write(&vnext.StreamChunk{Type: vnext.ChunkTypeDelta, Delta: "Test"})
			writer.Close()
		}()

		text, _, err := vnext.CollectStream(stream)
		if err != nil {
			t.Errorf("CollectStream() error = %v", err)
		}

		if text != expectedText {
			t.Errorf("CollectStream() = %q, want %q", text, expectedText)
		}
	})

	t.Run("collect with error", func(t *testing.T) {
		stream, writer := vnext.NewStreamBuilder().Build(ctx)

		go func() {
			writer.Write(&vnext.StreamChunk{Type: vnext.ChunkTypeDelta, Delta: "Hello"})
			writer.CloseWithError(errors.New("test error"))
		}()

		_, _, err := vnext.CollectStream(stream)
		if err == nil {
			t.Error("Expected error from CollectStream(), got nil")
		}
	})
}

// TestStreamToChannel tests the StreamToChannel utility
func TestStreamToChannel(t *testing.T) {
	ctx := context.Background()

	t.Run("convert to text channel", func(t *testing.T) {
		stream, writer := vnext.NewStreamBuilder().Build(ctx)

		testTexts := []string{"Hello", "World", "Test"}

		go func() {
			for _, text := range testTexts {
				writer.Write(&vnext.StreamChunk{
					Type:  vnext.ChunkTypeDelta,
					Delta: text,
				})
			}
			writer.Close()
		}()

		textChan := vnext.StreamToChannel(stream)
		received := []string{}

		for text := range textChan {
			received = append(received, text)
		}

		if len(received) != len(testTexts) {
			t.Errorf("Expected %d texts, got %d", len(testTexts), len(received))
		}

		for i, text := range received {
			if text != testTexts[i] {
				t.Errorf("Text %d = %q, want %q", i, text, testTexts[i])
			}
		}
	})
}

// TestStreamOptions tests functional options
func TestStreamOptions(t *testing.T) {
	t.Run("WithBufferSize", func(t *testing.T) {
		opts := &vnext.StreamOptions{}
		vnext.WithBufferSize(200)(opts)

		if opts.BufferSize != 200 {
			t.Errorf("BufferSize = %d, want 200", opts.BufferSize)
		}
	})

	t.Run("WithThoughts", func(t *testing.T) {
		opts := &vnext.StreamOptions{}
		vnext.WithThoughts()(opts)

		if !opts.IncludeThoughts {
			t.Error("IncludeThoughts should be true")
		}
	})

	t.Run("WithToolCalls", func(t *testing.T) {
		opts := &vnext.StreamOptions{}
		vnext.WithToolCalls()(opts)

		if !opts.IncludeToolCalls {
			t.Error("IncludeToolCalls should be true")
		}
	})

	t.Run("WithTextOnly", func(t *testing.T) {
		opts := &vnext.StreamOptions{}
		vnext.WithTextOnly()(opts)

		if !opts.TextOnly {
			t.Error("TextOnly should be true")
		}
	})

	t.Run("WithStreamTimeout", func(t *testing.T) {
		opts := &vnext.StreamOptions{}
		timeout := 30 * time.Second
		vnext.WithStreamTimeout(timeout)(opts)

		if opts.Timeout != timeout {
			t.Errorf("Timeout = %v, want %v", opts.Timeout, timeout)
		}
	})

	t.Run("WithFlushInterval", func(t *testing.T) {
		opts := &vnext.StreamOptions{}
		interval := 100 * time.Millisecond
		vnext.WithFlushInterval(interval)(opts)

		if opts.FlushInterval != interval {
			t.Errorf("FlushInterval = %v, want %v", opts.FlushInterval, interval)
		}
	})

	t.Run("multiple options", func(t *testing.T) {
		opts := &vnext.StreamOptions{}
		vnext.WithBufferSize(150)(opts)
		vnext.WithThoughts()(opts)
		vnext.WithStreamTimeout(60 * time.Second)(opts)

		if opts.BufferSize != 150 {
			t.Errorf("BufferSize = %d, want 150", opts.BufferSize)
		}
		if !opts.IncludeThoughts {
			t.Error("IncludeThoughts should be true")
		}
		if opts.Timeout != 60*time.Second {
			t.Error("Timeout mismatch")
		}
	})
}

// TestStreamErrorHandling tests error handling in streams
func TestStreamErrorHandling(t *testing.T) {
	ctx := context.Background()

	t.Run("error chunk propagation", func(t *testing.T) {
		stream, writer := vnext.NewStreamBuilder().Build(ctx)

		testError := errors.New("test error")

		go func() {
			writer.Write(&vnext.StreamChunk{
				Type:  vnext.ChunkTypeDelta,
				Delta: "before error",
			})
			writer.Write(&vnext.StreamChunk{
				Type:  vnext.ChunkTypeError,
				Error: testError,
			})
			writer.Close()
		}()

		var receivedError error
		for chunk := range stream.Chunks() {
			if chunk.Type == vnext.ChunkTypeError {
				receivedError = chunk.Error
			}
		}

		if receivedError == nil {
			t.Error("Expected error chunk, got none")
		}
		if receivedError.Error() != testError.Error() {
			t.Errorf("Error = %v, want %v", receivedError, testError)
		}
	})

	t.Run("CloseWithError", func(t *testing.T) {
		stream, writer := vnext.NewStreamBuilder().Build(ctx)

		testError := errors.New("close error")

		go func() {
			writer.Write(&vnext.StreamChunk{Type: vnext.ChunkTypeDelta, Delta: "test"})
			writer.CloseWithError(testError)
		}()

		for range stream.Chunks() {
		}

		_, err := stream.Wait()
		if err == nil {
			t.Error("Expected error from Wait(), got nil")
		}
	})
}

// TestStreamConcurrency tests concurrent access patterns
func TestStreamConcurrency(t *testing.T) {
	ctx := context.Background()

	t.Run("multiple writers", func(t *testing.T) {
		stream, writer := vnext.NewStreamBuilder().
			WithOption(vnext.WithBufferSize(100)).
			Build(ctx)

		numChunks := 50
		go func() {
			for i := 0; i < numChunks; i++ {
				writer.Write(&vnext.StreamChunk{
					Type:  vnext.ChunkTypeDelta,
					Delta: "test",
					Index: i,
				})
			}
			writer.Close()
		}()

		count := 0
		for range stream.Chunks() {
			count++
		}

		if count != numChunks {
			t.Errorf("Expected %d chunks, got %d", numChunks, count)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		stream, writer := vnext.NewStreamBuilder().Build(ctx)

		go func() {
			for i := 0; i < 100; i++ {
				select {
				case <-ctx.Done():
					writer.Close()
					return
				default:
					writer.Write(&vnext.StreamChunk{
						Type:  vnext.ChunkTypeDelta,
						Delta: "test",
					})
					time.Sleep(5 * time.Millisecond)
				}
			}
			writer.Close()
		}()

		count := 0
		for range stream.Chunks() {
			count++
			if count >= 10 {
				cancel() // Cancel context
				break
			}
		}

		if count < 10 {
			t.Errorf("Expected at least 10 chunks, got %d", count)
		}
	})
}
