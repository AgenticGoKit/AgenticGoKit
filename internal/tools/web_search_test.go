package tools

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// mockHTTPClient is a mock HTTP client for testing.
type mockHTTPClient struct {
	responseBody string
	statusCode   int
	error        error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.error != nil {
		return nil, m.error
	}
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(m.responseBody)),
		Header:     make(http.Header),
	}, nil
}

func TestWebSearchTool(t *testing.T) {
	ctx := context.Background()

	t.Run("Valid query returns results", func(t *testing.T) {
		mockClient := &mockHTTPClient{
			statusCode: http.StatusOK,
			responseBody: `{
				"web": {
					"results": [
						{"title": "The Go Programming Language", "url": "https://golang.org", "description": "Go is an open source programming language..."}
					]
				}
			}`,
		}
		tool := &WebSearchTool{apiKey: "dummy-key", httpClient: mockClient}

		result, err := tool.Call(ctx, map[string]any{"query": "golang"})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		results, ok := result["results"].([]string)
		if !ok {
			t.Fatalf("Expected results to be []string, got %T", result["results"])
		}
		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}
		if !strings.Contains(results[0], "Title: The Go Programming Language") {
			t.Errorf("Result mismatch: got %s", results[0])
		}
	})

	t.Run("No results found", func(t *testing.T) {
		mockClient := &mockHTTPClient{
			statusCode:   http.StatusOK,
			responseBody: `{"web": {"results": []}}`,
		}
		tool := &WebSearchTool{apiKey: "dummy-key", httpClient: mockClient}

		result, err := tool.Call(ctx, map[string]any{"query": "a query with no results"})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		results := result["results"].([]string)
		if len(results) != 1 || !strings.Contains(results[0], "No results found.") {
			t.Errorf("Expected 'No results found.', got: %v", results)
		}
	})

	t.Run("API error", func(t *testing.T) {
		mockClient := &mockHTTPClient{
			statusCode:   http.StatusUnauthorized,
			responseBody: `{"error": "invalid api key"}`,
		}
		tool := &WebSearchTool{apiKey: "dummy-key", httpClient: mockClient}

		_, err := tool.Call(ctx, map[string]any{"query": "test"})
		if err == nil {
			t.Fatal("Expected an error, but got nil")
		}
		if !strings.Contains(err.Error(), "search request failed") {
			t.Errorf("Expected error to contain 'search request failed', got: %v", err)
		}
	})

	t.Run("Missing query", func(t *testing.T) {
		tool := &WebSearchTool{apiKey: "dummy-key", httpClient: &mockHTTPClient{}}
		_, err := tool.Call(ctx, map[string]any{})
		if err == nil || err.Error() != "missing required argument 'query'" {
			t.Errorf("Expected missing argument error, got: %v", err)
		}
	})

	t.Run("Empty query", func(t *testing.T) {
		tool := &WebSearchTool{apiKey: "dummy-key", httpClient: &mockHTTPClient{}}
		_, err := tool.Call(ctx, map[string]any{"query": ""})
		if err == nil || err.Error() != "argument 'query' must be a non-empty string" {
			t.Errorf("Expected empty string error, got: %v", err)
		}
	})
}
