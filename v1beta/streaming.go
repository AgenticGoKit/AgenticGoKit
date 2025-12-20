// package v1beta provides streaming capabilities for agent execution
package v1beta

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"
)

// =============================================================================
// STREAMING.GO - Unified Streaming System
// =============================================================================
//
// This file provides:
// - Simplified streaming interfaces for agent, workflow, and tool execution
// - Stream chunk types and metadata
// - Stream lifecycle management and error handling
// - Channel-based and callback-based streaming patterns
//
// =============================================================================

// =============================================================================
// STREAM CHUNK TYPES
// =============================================================================

// ChunkType identifies the type of streaming chunk
type ChunkType string

const (
	ChunkTypeText          ChunkType = "text"           // Text content chunk
	ChunkTypeDelta         ChunkType = "delta"          // Incremental update
	ChunkTypeThought       ChunkType = "thought"        // Agent reasoning/thinking
	ChunkTypeToolCall      ChunkType = "tool_call"      // Tool invocation
	ChunkTypeToolRes       ChunkType = "tool_result"    // Tool result
	ChunkTypeMetadata      ChunkType = "metadata"       // Metadata update
	ChunkTypeError         ChunkType = "error"          // Error chunk
	ChunkTypeDone          ChunkType = "done"           // Stream completion
	ChunkTypeAgentStart    ChunkType = "agent_start"    // Agent/step begins execution (workflow lifecycle)
	ChunkTypeAgentComplete ChunkType = "agent_complete" // Agent/step completes execution (workflow lifecycle)
	// New chunk types for multimodal content
	ChunkTypeImage ChunkType = "image"
	ChunkTypeAudio ChunkType = "audio"
	ChunkTypeVideo ChunkType = "video"
)

// StreamChunk represents a single chunk in a streaming response
//
// Chunks are emitted as the agent generates output. They can contain
// text content, tool calls, metadata, or signal stream completion.
type StreamChunk struct {
	Type      ChunkType              `json:"type"`                // Type of chunk
	Content   string                 `json:"content,omitempty"`   // Text content
	Delta     string                 `json:"delta,omitempty"`     // Incremental text delta
	ToolName  string                 `json:"tool_name,omitempty"` // Tool being called
	ToolArgs  map[string]interface{} `json:"tool_args,omitempty"` // Tool arguments
	ToolID    string                 `json:"tool_id,omitempty"`   // Tool call identifier
	Metadata  map[string]interface{} `json:"metadata,omitempty"`  // Additional metadata
	Error     error                  `json:"error,omitempty"`     // Error if type is error
	Timestamp time.Time              `json:"timestamp"`           // When chunk was created
	Index     int                    `json:"index"`               // Chunk sequence number
	// New fields for multimodal content
	ImageData *ImageData `json:"image_data,omitempty"` // Image data for image chunks
	AudioData *AudioData `json:"audio_data,omitempty"` // Audio data for audio chunks
	VideoData *VideoData `json:"video_data,omitempty"` // Video data for video chunks
}

// StreamMetadata contains information about the stream
type StreamMetadata struct {
	AgentName string                 `json:"agent_name"`
	SessionID string                 `json:"session_id,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	StartTime time.Time              `json:"start_time"`
	Model     string                 `json:"model,omitempty"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

// ImageData represents image content in the stream
type ImageData struct {
	URL      string            `json:"url,omitempty"`
	Base64   string            `json:"base64,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// AudioData represents audio content in the stream
type AudioData struct {
	URL      string            `json:"url,omitempty"`
	Base64   string            `json:"base64,omitempty"`
	Format   string            `json:"format,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// VideoData represents video content in the stream
type VideoData struct {
	URL      string            `json:"url,omitempty"`
	Base64   string            `json:"base64,omitempty"`
	Format   string            `json:"format,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Attachment represents generic media attachment
type Attachment struct {
	Name     string            `json:"name,omitempty"`
	Type     string            `json:"type,omitempty"`
	Data     []byte            `json:"data,omitempty"`
	URL      string            `json:"url,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// =============================================================================
// STREAMING INTERFACES
// =============================================================================

// StreamHandler is a function that processes streaming chunks
//
// Return false to stop streaming, true to continue.
// Errors are returned via error chunks, not as return values.
//
// Example:
//
//	handler := func(chunk *StreamChunk) bool {
//	    fmt.Print(chunk.Delta)
//	    return true  // continue streaming
//	}
type StreamHandler func(chunk *StreamChunk) bool

// Stream represents an active streaming session
//
// Provides multiple ways to consume streaming output:
// - Channel-based: Read from Chunks() channel
// - Callback-based: Use handler function
// - Reader-based: Use io.Reader interface
type Stream interface {
	// Chunks returns a channel for receiving stream chunks
	Chunks() <-chan *StreamChunk

	// Wait blocks until the stream is complete or cancelled
	// Returns the final result and any error
	Wait() (*Result, error)

	// Cancel stops the stream
	Cancel()

	// Metadata returns stream metadata
	Metadata() *StreamMetadata

	// AsReader returns an io.Reader that provides text content only
	AsReader() io.Reader
}

// StreamWriter is used internally to write chunks to a stream
type StreamWriter interface {
	Write(chunk *StreamChunk) error
	Close() error
	CloseWithError(err error) error
}

// =============================================================================
// STREAM OPTIONS
// =============================================================================

// StreamOption configures streaming behavior
type StreamOption func(*StreamOptions)

// StreamOptions configures how streaming is performed
type StreamOptions struct {
	BufferSize       int                    // Channel buffer size (default: 100)
	Handler          StreamHandler          // Optional callback handler
	IncludeThoughts  bool                   // Include reasoning/thinking chunks
	IncludeToolCalls bool                   // Include tool call chunks
	IncludeMetadata  bool                   // Include metadata chunks
	TextOnly         bool                   // Only emit text chunks
	FlushInterval    time.Duration          // How often to flush buffers
	Timeout          time.Duration          // Stream timeout
	Metadata         map[string]interface{} // Additional metadata
}

// DefaultStreamOptions returns default streaming options
func DefaultStreamOptions() *StreamOptions {
	return &StreamOptions{
		BufferSize:       100,
		IncludeThoughts:  true,
		IncludeToolCalls: true,
		IncludeMetadata:  false,
		TextOnly:         false,
		FlushInterval:    100 * time.Millisecond,
		Timeout:          5 * time.Minute,
		Metadata:         make(map[string]interface{}),
	}
}

// WithBufferSize sets the stream channel buffer size
func WithBufferSize(size int) StreamOption {
	return func(opts *StreamOptions) {
		opts.BufferSize = size
	}
}

// WithStreamHandler sets a callback handler for chunks
func WithStreamHandler(handler StreamHandler) StreamOption {
	return func(opts *StreamOptions) {
		opts.Handler = handler
	}
}

// WithThoughts includes agent reasoning chunks in the stream
func WithThoughts() StreamOption {
	return func(opts *StreamOptions) {
		opts.IncludeThoughts = true
	}
}

// WithToolCalls includes tool invocation chunks in the stream
func WithToolCalls() StreamOption {
	return func(opts *StreamOptions) {
		opts.IncludeToolCalls = true
	}
}

// WithStreamMetadata includes metadata chunks in the stream
func WithStreamMetadata() StreamOption {
	return func(opts *StreamOptions) {
		opts.IncludeMetadata = true
	}
}

// WithTextOnly emits only text content chunks
func WithTextOnly() StreamOption {
	return func(opts *StreamOptions) {
		opts.TextOnly = true
		opts.IncludeThoughts = false
		opts.IncludeToolCalls = false
		opts.IncludeMetadata = false
	}
}

// WithStreamTimeout sets the stream timeout duration
func WithStreamTimeout(timeout time.Duration) StreamOption {
	return func(opts *StreamOptions) {
		opts.Timeout = timeout
	}
}

// WithFlushInterval sets how often to flush buffered chunks
func WithFlushInterval(interval time.Duration) StreamOption {
	return func(opts *StreamOptions) {
		opts.FlushInterval = interval
	}
}

// =============================================================================
// STREAM IMPLEMENTATION
// =============================================================================

// basicStream implements the Stream interface
type basicStream struct {
	chunks   chan *StreamChunk
	metadata *StreamMetadata
	options  *StreamOptions
	cancel   context.CancelFunc
	ctx      context.Context
	result   *Result
	err      error
	mu       sync.RWMutex
	done     chan struct{}
	chunkIdx int
}

// NewStream creates a new streaming session
func NewStream(ctx context.Context, metadata *StreamMetadata, opts ...StreamOption) (Stream, StreamWriter) {
	options := DefaultStreamOptions()
	for _, opt := range opts {
		opt(options)
	}

	ctx, cancel := context.WithCancel(ctx)
	if options.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
	}

	stream := &basicStream{
		chunks:   make(chan *StreamChunk, options.BufferSize),
		metadata: metadata,
		options:  options,
		cancel:   cancel,
		ctx:      ctx,
		done:     make(chan struct{}),
		chunkIdx: 0,
	}

	return stream, stream
}

// Chunks implements Stream.Chunks
func (s *basicStream) Chunks() <-chan *StreamChunk {
	return s.chunks
}

// Wait implements Stream.Wait
func (s *basicStream) Wait() (*Result, error) {
	<-s.done
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.result, s.err
}

// Cancel implements Stream.Cancel
func (s *basicStream) Cancel() {
	s.cancel()
}

// Metadata implements Stream.Metadata
func (s *basicStream) Metadata() *StreamMetadata {
	return s.metadata
}

// AsReader implements Stream.AsReader
func (s *basicStream) AsReader() io.Reader {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		for chunk := range s.chunks {
			// Only write text content to the reader
			if chunk.Type == ChunkTypeText || chunk.Type == ChunkTypeDelta {
				content := chunk.Content
				if content == "" {
					content = chunk.Delta
				}
				if content != "" {
					if _, err := pw.Write([]byte(content)); err != nil {
						pw.CloseWithError(err)
						return
					}
				}
			}
		}
	}()

	return pr
}

// Write implements StreamWriter.Write
func (s *basicStream) Write(chunk *StreamChunk) error {
	// Check if context is cancelled
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
	}

	// Filter chunks based on options
	if s.options.TextOnly {
		if chunk.Type != ChunkTypeText && chunk.Type != ChunkTypeDelta && chunk.Type != ChunkTypeDone {
			return nil // Skip non-text chunks
		}
	}

	if !s.options.IncludeThoughts && chunk.Type == ChunkTypeThought {
		return nil
	}

	if !s.options.IncludeToolCalls && (chunk.Type == ChunkTypeToolCall || chunk.Type == ChunkTypeToolRes) {
		return nil
	}

	if !s.options.IncludeMetadata && chunk.Type == ChunkTypeMetadata {
		return nil
	}

	// Set chunk metadata
	chunk.Timestamp = time.Now()
	chunk.Index = s.chunkIdx
	s.chunkIdx++

	// Call handler if provided
	if s.options.Handler != nil {
		if !s.options.Handler(chunk) {
			s.cancel()
			return fmt.Errorf("stream cancelled by handler")
		}
	}

	// Send chunk to channel
	select {
	case s.chunks <- chunk:
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

// Close implements StreamWriter.Close
func (s *basicStream) Close() error {
	return s.CloseWithError(nil)
}

// CloseWithError implements StreamWriter.CloseWithError
func (s *basicStream) CloseWithError(err error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Only close once
	select {
	case <-s.done:
		return nil
	default:
	}

	s.err = err
	close(s.chunks)
	close(s.done)
	s.cancel()

	return nil
}

// SetResult sets the final result for the stream
func (s *basicStream) SetResult(result *Result) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.result = result
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// CollectStream collects all text chunks from a stream into a single string
//
// This is a convenience function for when you want to buffer the entire
// streaming output rather than processing chunks incrementally.
//
// Example:
//
//	stream, _ := agent.RunStream(ctx, "Hello")
//	output, result, err := CollectStream(stream)
func CollectStream(stream Stream) (string, *Result, error) {
	var buffer []string

	for chunk := range stream.Chunks() {
		if chunk.Type == ChunkTypeText || chunk.Type == ChunkTypeDelta {
			content := chunk.Content
			if content == "" {
				content = chunk.Delta
			}
			buffer = append(buffer, content)
		}
	}

	result, err := stream.Wait()
	if err != nil {
		return "", nil, err
	}

	output := ""
	for _, s := range buffer {
		output += s
	}

	return output, result, nil
}

// StreamToChannel converts a Stream to a simple text channel
//
// This is useful when you only care about text content and want a simpler API.
//
// Example:
//
//	stream, _ := agent.RunStream(ctx, "Hello")
//	textChan := StreamToChannel(stream)
//	for text := range textChan {
//	    fmt.Print(text)
//	}
func StreamToChannel(stream Stream) <-chan string {
	textChan := make(chan string)

	go func() {
		defer close(textChan)

		for chunk := range stream.Chunks() {
			if chunk.Type == ChunkTypeText || chunk.Type == ChunkTypeDelta {
				content := chunk.Content
				if content == "" {
					content = chunk.Delta
				}
				if content != "" {
					textChan <- content
				}
			}
		}
	}()

	return textChan
}

// PrintStream prints a stream to stdout in real-time
//
// This is a convenience function for demos and testing.
//
// Example:
//
//	stream, _ := agent.RunStream(ctx, "Hello")
//	result, err := PrintStream(stream)
func PrintStream(stream Stream) (*Result, error) {
	for chunk := range stream.Chunks() {
		switch chunk.Type {
		case ChunkTypeText, ChunkTypeDelta:
			content := chunk.Content
			if content == "" {
				content = chunk.Delta
			}
			fmt.Print(content)
		case ChunkTypeThought:
			fmt.Printf("\n[Thinking: %s]\n", chunk.Content)
		case ChunkTypeToolCall:
			fmt.Printf("\n[Tool: %s]\n", chunk.ToolName)
		case ChunkTypeError:
			fmt.Printf("\n[Error: %v]\n", chunk.Error)
		}
	}

	return stream.Wait()
}

// =============================================================================
// STREAM BUILDER
// =============================================================================

// StreamBuilder provides a fluent interface for creating streams
type StreamBuilder struct {
	metadata *StreamMetadata
	options  []StreamOption
}

// NewStreamBuilder creates a new stream builder
func NewStreamBuilder() *StreamBuilder {
	return &StreamBuilder{
		metadata: &StreamMetadata{
			StartTime: time.Now(),
			Extra:     make(map[string]interface{}),
		},
		options: []StreamOption{},
	}
}

// WithAgentName sets the agent name in metadata
func (b *StreamBuilder) WithAgentName(name string) *StreamBuilder {
	b.metadata.AgentName = name
	return b
}

// WithSessionID sets the session ID in metadata
func (b *StreamBuilder) WithSessionID(id string) *StreamBuilder {
	b.metadata.SessionID = id
	return b
}

// WithTraceID sets the trace ID in metadata
func (b *StreamBuilder) WithTraceID(id string) *StreamBuilder {
	b.metadata.TraceID = id
	return b
}

// WithModel sets the model name in metadata
func (b *StreamBuilder) WithModel(model string) *StreamBuilder {
	b.metadata.Model = model
	return b
}

// WithOption adds a stream option
func (b *StreamBuilder) WithOption(opt StreamOption) *StreamBuilder {
	b.options = append(b.options, opt)
	return b
}

// Build creates the stream
func (b *StreamBuilder) Build(ctx context.Context) (Stream, StreamWriter) {
	return NewStream(ctx, b.metadata, b.options...)
}

// =============================================================================
// EXAMPLES
// =============================================================================
//
// EXAMPLE 1: Basic Streaming with Channel
//
//	stream, err := agent.RunStream(ctx, "Explain quantum computing")
//	for chunk := range stream.Chunks() {
//	    fmt.Print(chunk.Delta)
//	}
//	result, err := stream.Wait()
//
// EXAMPLE 2: Streaming with Handler Callback
//
//	handler := func(chunk *StreamChunk) bool {
//	    if chunk.Type == ChunkTypeDelta {
//	        fmt.Print(chunk.Delta)
//	    }
//	    return true
//	}
//	stream, err := agent.RunStream(ctx, "Hello", WithStreamHandler(handler))
//	result, err := stream.Wait()
//
// EXAMPLE 3: Text-Only Streaming
//
//	stream, err := agent.RunStream(ctx, "Hello", WithTextOnly())
//	output, result, err := CollectStream(stream)
//	fmt.Println(output)
//
// EXAMPLE 4: Streaming with Thoughts and Tool Calls
//
//	stream, err := agent.RunStream(ctx, "Search for...",
//	    WithThoughts(),
//	    WithToolCalls())
//
//	for chunk := range stream.Chunks() {
//	    switch chunk.Type {
//	    case ChunkTypeDelta:
//	        fmt.Print(chunk.Delta)
//	    case ChunkTypeThought:
//	        log.Printf("Thinking: %s", chunk.Content)
//	    case ChunkTypeToolCall:
//	        log.Printf("Calling tool: %s", chunk.ToolName)
//	    }
//	}
//
// EXAMPLE 5: Stream to io.Reader
//
//	stream, err := agent.RunStream(ctx, "Hello")
//	reader := stream.AsReader()
//	io.Copy(os.Stdout, reader)
//	result, err := stream.Wait()
//
// EXAMPLE 6: Workflow Streaming
//
//	workflow, _ := QuickWorkflow([]Agent{agent1, agent2})
//	stream, err := workflow.ExecuteStream(ctx, "Hello")
//	PrintStream(stream)
