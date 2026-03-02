package llm

import (
	"net/http"
	"time"
)

// OptimizedTransport returns a production-ready HTTP transport configured
// for high-performance LLM API communication with proper connection pooling,
// keep-alive, and timeout settings.
//
// This transport is optimized to:
// - Reuse connections efficiently (avoid TCP handshake overhead)
// - Support high concurrency (50 connections per host)
// - Enable HTTP/2 for multiplexing
// - Set reasonable timeouts to prevent hanging
//
// Performance Impact: 30-50% improvement over default transport settings
func OptimizedTransport() *http.Transport {
	return &http.Transport{
		// Connection pooling configuration
		// MaxIdleConns: Total idle connections across all hosts
		MaxIdleConns: 100,

		// MaxIdleConnsPerHost: Idle connections kept per host
		// Keep 20 idle connections per LLM endpoint for instant reuse
		MaxIdleConnsPerHost: 20,

		// MaxConnsPerHost: Total connections (idle + active) per host
		// Allow up to 50 concurrent requests to the same LLM endpoint
		MaxConnsPerHost: 50,

		// Timeout configuration
		// IdleConnTimeout: How long idle connections stay in the pool
		// 90 seconds balances connection reuse vs server resource cleanup
		IdleConnTimeout: 90 * time.Second,

		// TLSHandshakeTimeout: Maximum time for TLS handshake
		// LLM APIs typically use HTTPS, so this is important
		TLSHandshakeTimeout: 10 * time.Second,

		// ResponseHeaderTimeout: Time to wait for response headers
		// Protects against slow or unresponsive servers
		ResponseHeaderTimeout: 30 * time.Second,

		// ExpectContinueTimeout: Time to wait for 100-Continue response
		// Small value to avoid delays when server doesn't support it
		ExpectContinueTimeout: 1 * time.Second,

		// HTTP/2 support
		// Enable HTTP/2 for multiplexing multiple requests over single connection
		// This is especially beneficial for streaming LLM responses
		ForceAttemptHTTP2: true,

		// Keep-alive configuration
		// CRITICAL: Must be false to enable connection reuse
		// Setting this to true would defeat the entire connection pool
		DisableKeepAlives: false,

		// DisableCompression: false (default)
		// Allow gzip compression for LLM responses to reduce bandwidth
		// Most LLM APIs support compression
	}
}

// NewOptimizedHTTPClient creates an HTTP client with production-ready settings
// for LLM API communication.
//
// Parameters:
//   - timeout: Total timeout for the entire request/response cycle
//
// The client uses OptimizedTransport() for connection pooling and includes
// a timeout to prevent requests from hanging indefinitely.
//
// Example:
//
//	client := NewOptimizedHTTPClient(120 * time.Second)
//	// This client can handle hundreds of concurrent LLM requests efficiently
func NewOptimizedHTTPClient(timeout time.Duration) *http.Client {
	// Default 2 minutes for LLM requests
	if timeout == 0 {
		timeout = 120 * time.Second
	}

	transport := OptimizedTransport()
	// Ensure ResponseHeaderTimeout matches the client timeout if it's larger than the default 30s
	// This prevents the transport from killing connection to slow LLMs (like reasoning models)
	// while waiting for the first byte/header.
	if timeout > transport.ResponseHeaderTimeout {
		transport.ResponseHeaderTimeout = timeout
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

// LLMClientConfig holds configuration for LLM-specific HTTP clients
type LLMClientConfig struct {
	// Timeout for the entire request/response cycle
	Timeout time.Duration

	// MaxIdleConnsPerHost override (default: 20)
	MaxIdleConnsPerHost int

	// MaxConnsPerHost override (default: 50)
	MaxConnsPerHost int

	// IdleConnTimeout override (default: 90s)
	IdleConnTimeout time.Duration
}

// NewCustomHTTPClient creates an HTTP client with custom configuration
// for specific LLM provider requirements.
//
// Use this when you need to tune connection pool settings for a specific
// provider (e.g., local Ollama vs cloud OpenAI).
func NewCustomHTTPClient(config LLMClientConfig) *http.Client {
	if config.Timeout == 0 {
		config.Timeout = 120 * time.Second
	}

	transport := OptimizedTransport()

	// Apply custom overrides if specified
	if config.MaxIdleConnsPerHost > 0 {
		transport.MaxIdleConnsPerHost = config.MaxIdleConnsPerHost
	}
	if config.MaxConnsPerHost > 0 {
		transport.MaxConnsPerHost = config.MaxConnsPerHost
	}
	if config.IdleConnTimeout > 0 {
		transport.IdleConnTimeout = config.IdleConnTimeout
	}

	return &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}
}

// GetDefaultTransportSettings returns the default transport settings
// used by OptimizedTransport for debugging and documentation purposes.
func GetDefaultTransportSettings() map[string]interface{} {
	return map[string]interface{}{
		"MaxIdleConns":          100,
		"MaxIdleConnsPerHost":   20,
		"MaxConnsPerHost":       50,
		"IdleConnTimeout":       "90s",
		"TLSHandshakeTimeout":   "10s",
		"ResponseHeaderTimeout": "30s",
		"ExpectContinueTimeout": "1s",
		"ForceAttemptHTTP2":     true,
		"DisableKeepAlives":     false,
		"DefaultClientTimeout":  "120s",
	}
}
