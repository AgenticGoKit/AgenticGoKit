package v1beta

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestCollectStream_ToolError(t *testing.T) {
	ctx := context.Background()
	stream, writer := NewStream(ctx, &StreamMetadata{})

	go func() {
		defer writer.Close()
		writer.Write(&StreamChunk{Type: ChunkTypeText, Content: "Thinking about it..."})
		writer.Write(&StreamChunk{
			Type:     ChunkTypeToolRes,
			ToolName: "broken_tool",
			Error:    fmt.Errorf("simulated failure"),
		})
		writer.Write(&StreamChunk{Type: ChunkTypeText, Content: " And I'm done."})
	}()

	output, _, err := CollectStream(stream)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSub := "Tool \"broken_tool\" failed with error: simulated failure"
	if !strings.Contains(output, expectedSub) {
		t.Errorf("output missing expected tool error.\nExpected to contain: %q\nGot: %q", expectedSub, output)
	}
}

func TestCollectStream_ToolSuccess(t *testing.T) {
	ctx := context.Background()
	stream, writer := NewStream(ctx, &StreamMetadata{})

	go func() {
		defer writer.Close()
		writer.Write(&StreamChunk{Type: ChunkTypeText, Content: "Thinking about it..."})
		writer.Write(&StreamChunk{
			Type:     ChunkTypeToolRes,
			ToolName: "working_tool",
			Content:  "42",
		})
		writer.Write(&StreamChunk{Type: ChunkTypeText, Content: " And I'm done."})
	}()

	output, _, err := CollectStream(stream)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSub := "Tool \"working_tool\" returned: 42"
	if !strings.Contains(output, expectedSub) {
		t.Errorf("output missing expected tool result.\nExpected to contain: %q\nGot: %q", expectedSub, output)
	}
}

func TestCollectStream_StreamError(t *testing.T) {
	ctx := context.Background()
	stream, writer := NewStream(ctx, &StreamMetadata{})

	go func() {
		defer writer.Close()
		writer.Write(&StreamChunk{
			Type:  ChunkTypeError,
			Error: fmt.Errorf("connection dropped"),
		})
	}()

	output, _, err := CollectStream(stream)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSub := "Stream error: connection dropped"
	if !strings.Contains(output, expectedSub) {
		t.Errorf("output missing expected stream error.\nExpected to contain: %q\nGot: %q", expectedSub, output)
	}
}

func TestStreamToChannel_ToolError(t *testing.T) {
	ctx := context.Background()
	stream, writer := NewStream(ctx, &StreamMetadata{})

	go func() {
		defer writer.Close()
		writer.Write(&StreamChunk{Type: ChunkTypeText, Content: "Start."})
		writer.Write(&StreamChunk{
			Type:     ChunkTypeToolRes,
			ToolName: "failing_tool",
			Error:    fmt.Errorf("kaboom"),
		})
		writer.Write(&StreamChunk{Type: ChunkTypeText, Content: "End."})
	}()

	ch := StreamToChannel(stream)

	var chunks []string
	for c := range ch {
		chunks = append(chunks, c)
	}
	output := strings.Join(chunks, "")

	expectedSub := "Tool \"failing_tool\" failed with error: kaboom"
	if !strings.Contains(output, expectedSub) {
		t.Errorf("output missing expected tool error.\nExpected to contain: %q\nGot: %q", expectedSub, output)
	}
}

func TestStreamToChannel_StreamError(t *testing.T) {
	ctx := context.Background()
	stream, writer := NewStream(ctx, &StreamMetadata{})

	go func() {
		defer writer.Close()
		writer.Write(&StreamChunk{
			Type:  ChunkTypeError,
			Error: fmt.Errorf("fatal fault"),
		})
	}()

	ch := StreamToChannel(stream)

	var chunks []string
	for c := range ch {
		chunks = append(chunks, c)
	}
	output := strings.Join(chunks, "")

	expectedSub := "Stream error: fatal fault"
	if !strings.Contains(output, expectedSub) {
		t.Errorf("output missing expected stream error.\nExpected to contain: %q\nGot: %q", expectedSub, output)
	}
}
