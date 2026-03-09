package llm

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptimizedTransport_DefaultResponseHeaderTimeout(t *testing.T) {
	transport := OptimizedTransport()
	assert.Equal(t, 30*time.Second, transport.ResponseHeaderTimeout,
		"default ResponseHeaderTimeout should be 30s")
}

func TestOptimizedTransport_CustomResponseHeaderTimeout(t *testing.T) {
	timeout := 5 * time.Minute
	transport := OptimizedTransport(timeout)
	assert.Equal(t, timeout, transport.ResponseHeaderTimeout,
		"ResponseHeaderTimeout should match the provided override")
}

func TestOptimizedTransport_ZeroOverrideFallsBackToDefault(t *testing.T) {
	transport := OptimizedTransport(0)
	assert.Equal(t, 30*time.Second, transport.ResponseHeaderTimeout,
		"zero override should fall back to 30s default")
}

func TestNewOptimizedHTTPClient_PropagatesTimeout(t *testing.T) {
	timeout := 10 * time.Minute
	client := NewOptimizedHTTPClient(timeout)
	require.NotNil(t, client)

	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok, "transport should be *http.Transport")

	assert.Equal(t, timeout, client.Timeout,
		"client Timeout should match the provided timeout")
	assert.Equal(t, timeout, transport.ResponseHeaderTimeout,
		"ResponseHeaderTimeout should match the client timeout")
}

func TestNewOptimizedHTTPClient_DefaultTimeout(t *testing.T) {
	client := NewOptimizedHTTPClient(0)
	require.NotNil(t, client)

	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)

	assert.Equal(t, 120*time.Second, client.Timeout,
		"default client Timeout should be 120s")
	assert.Equal(t, 120*time.Second, transport.ResponseHeaderTimeout,
		"ResponseHeaderTimeout should match the default 120s client timeout")
}

func TestNewCustomHTTPClient_PropagatesTimeout(t *testing.T) {
	timeout := 10 * time.Minute
	client := NewCustomHTTPClient(LLMClientConfig{
		Timeout: timeout,
	})
	require.NotNil(t, client)

	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok, "transport should be *http.Transport")

	assert.Equal(t, timeout, client.Timeout,
		"client Timeout should match the provided config timeout")
	assert.Equal(t, timeout, transport.ResponseHeaderTimeout,
		"ResponseHeaderTimeout should match the config timeout — this was the bug")
}

func TestNewCustomHTTPClient_DefaultTimeout(t *testing.T) {
	client := NewCustomHTTPClient(LLMClientConfig{})
	require.NotNil(t, client)

	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)

	assert.Equal(t, 120*time.Second, client.Timeout)
	assert.Equal(t, 120*time.Second, transport.ResponseHeaderTimeout,
		"default ResponseHeaderTimeout should be 120s when no timeout is configured")
}
