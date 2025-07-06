package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

// HTTPClient is an interface for making HTTP requests.
// It's implemented by *http.Client and our mock client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// BraveAPIResponse defines the structure for the Brave Search API response.
type BraveAPIResponse struct {
	Web struct {
		Results []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
		} `json:"results"`
	} `json:"web"`
}

// WebSearchTool uses the Brave Search API to perform web searches.
type WebSearchTool struct {
	apiKey     string
	httpClient HTTPClient // Use the interface here
}

// NewWebSearchTool creates a new instance of the WebSearchTool.
// It reads the BRAVE_API_KEY environment variable.
func NewWebSearchTool() (*WebSearchTool, error) {
	apiKey := os.Getenv("BRAVE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("BRAVE_API_KEY environment variable not set")
	}
	return &WebSearchTool{
		apiKey:     apiKey,
		httpClient: http.DefaultClient, // http.DefaultClient satisfies the interface
	}, nil
}

// Name returns the tool's name.
func (t *WebSearchTool) Name() string {
	return "web_search"
}

// Call performs a web search using the Brave Search API.
// Expects "query" (string) in args.
// Returns "results" ([]string) in the result map.
func (t *WebSearchTool) Call(ctx context.Context, args map[string]any) (map[string]any, error) {
	queryVal, ok := args["query"]
	if !ok {
		return nil, fmt.Errorf("missing required argument 'query'")
	}
	query, ok := queryVal.(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("argument 'query' must be a non-empty string")
	}

	log.Printf("WebSearchTool: Performing Brave API search for query: %q", query)

	apiURL := "https://api.search.brave.com/res/v1/web/search?q=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", t.apiKey)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch search results: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search request failed: %s", resp.Status)
	}

	var apiResponse BraveAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	if len(apiResponse.Web.Results) == 0 {
		return map[string]any{"results": []string{"No results found."}}, nil
	}

	var results []string
	for _, r := range apiResponse.Web.Results {
		results = append(results, fmt.Sprintf("Title: %s\nURL: %s\nSnippet: %s", r.Title, r.URL, r.Description))
	}

	return map[string]any{
		"results": results,
	}, nil
}
