# Web Search Integration

**Give your agents the ability to search the web for real-time information**

This guide shows you how to integrate web search capabilities into your AgenticGoKit agents using MCP (Model Context Protocol) tools. You'll learn to set up web search, handle results, and optimize for different use cases.

## Prerequisites

- Basic AgenticGoKit project setup
- Understanding of MCP tool integration
- API key for a search service (optional for basic setup)

## What You'll Build

An agent system that can:
- Search the web for current information
- Process and summarize search results
- Handle search errors gracefully
- Cache results for performance

## Quick Start

### 1. Create a Project with Web Search

```bash
# Create project with MCP web search enabled
agentcli create web-search-agent --mcp-enabled --agents 2 \\
  --mcp-tools \"web_search,summarize\"
cd web-search-agent
```

### 2. Configure Search Provider

The generated `agentflow.toml` includes MCP server configuration:

```toml
[mcp]
enabled = true
enable_discovery = true
connection_timeout = 5000
max_retries = 3
enable_caching = true
cache_timeout = 300000

# Configure search server (example)
[[mcp.servers]]
name = \"brave-search\"
type = \"stdio\"
command = \"npx @modelcontextprotocol/server-brave-search\"
enabled = true  # Enable this server
```

### 3. Set Up Environment

```bash
# If using Brave Search (free tier available)
export BRAVE_API_KEY=your-brave-api-key

# Or use other search providers
export SERP_API_KEY=your-serp-api-key
```

### 4. Test Web Search

```bash
# Set your LLM provider key
export OPENAI_API_KEY=your-openai-key

# Run the agent
go run . -m \"What are the latest developments in AI?\"
```

## Available Search Providers

| Provider | MCP Server | Free Tier | Rate Limits |
|----------|------------|-----------|-------------|
| **Brave Search** | `@modelcontextprotocol/server-brave-search` | 2,000 queries/month | 1 req/sec |
| **SerpAPI** | `@modelcontextprotocol/server-serpapi` | 100 queries/month | Varies |
| **DuckDuckGo** | `@modelcontextprotocol/server-duckduckgo` | Unlimited | Rate limited |

## Custom Search Integration

If you need a specific search provider, you can create a custom tool:

```go
package tools

import (
    \"context\"
    \"encoding/json\"
    \"fmt\"
    \"net/http\"
    \"net/url\"
)

type CustomSearchTool struct {
    apiKey     string
    httpClient *http.Client
}

func NewCustomSearchTool(apiKey string) *CustomSearchTool {
    return &CustomSearchTool{
        apiKey:     apiKey,
        httpClient: &http.Client{Timeout: 10 * time.Second},
    }
}

func (t *CustomSearchTool) Name() string {
    return \"custom_search\"
}

func (t *CustomSearchTool) Description() string {
    return \"Search the web using custom search API\"
}

func (t *CustomSearchTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    query, ok := params[\"query\"].(string)
    if !ok || query == \"\" {
        return nil, fmt.Errorf(\"query parameter is required\")
    }
    
    numResults := 5
    if nr, ok := params[\"num_results\"].(float64); ok {
        numResults = int(nr)
    }
    
    // Build search URL (example using a hypothetical API)
    searchURL := fmt.Sprintf(\"https://api.example-search.com/search?q=%s&count=%d&key=%s\",
        url.QueryEscape(query), numResults, t.apiKey)
    
    req, err := http.NewRequestWithContext(ctx, \"GET\", searchURL, nil)
    if err != nil {
        return nil, fmt.Errorf(\"failed to create request: %w\", err)
    }
    
    resp, err := t.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf(\"search request failed: %w\", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf(\"search API returned status %d\", resp.StatusCode)
    }
    
    var searchResults map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&searchResults); err != nil {
        return nil, fmt.Errorf(\"failed to parse search results: %w\", err)
    }
    
    return searchResults, nil
}
```

## Advanced Patterns

### Search Result Processing

Create an agent specifically for processing search results:

```go
type SearchProcessorAgent struct {
    name        string
    llmProvider core.ModelProvider
}

func (a *SearchProcessorAgent) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    searchResults, ok := event.Data[\"search_results\"]
    if !ok {
        return nil, fmt.Errorf(\"no search results provided\")
    }
    
    prompt := fmt.Sprintf(`
Analyze these search results and provide a comprehensive summary:

%v

Please provide:
1. Key findings and insights
2. Source credibility assessment
3. Conflicting information (if any)
4. Recommended follow-up searches

Format your response as structured JSON.
`, searchResults)
    
    response, err := a.llmProvider.GenerateResponse(ctx, prompt, nil)
    if err != nil {
        return nil, fmt.Errorf(\"failed to process search results: %w\", err)
    }
    
    return &core.AgentResult{
        Data: map[string]interface{}{
            \"processed_results\": response,
            \"source_count\":      extractSourceCount(searchResults),
        },
    }, nil
}
```

### Error Handling

Implement robust error handling for search operations:

```go
func (a *SearchAgent) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    query := event.Data[\"query\"].(string)
    
    // Try primary search method
    results, err := a.performSearch(ctx, query)
    if err != nil {
        // Log the error but don't fail immediately
        log.Printf(\"Primary search failed: %v\", err)
        
        // Try fallback search method
        results, err = a.performFallbackSearch(ctx, query)
        if err != nil {
            // If both fail, return a graceful response
            return &core.AgentResult{
                Data: map[string]interface{}{
                    \"error\":   \"Search temporarily unavailable\",
                    \"message\": \"Unable to perform web search at this time. Please try again later.\",
                },
            }, nil
        }
    }
    
    return &core.AgentResult{
        Data: map[string]interface{}{
            \"search_results\": results,
            \"query\":         query,
            \"timestamp\":     time.Now().Unix(),
        },
    }, nil
}
```

## Performance Optimization

### Rate Limiting

Implement rate limiting to stay within API limits:

```go
import \"golang.org/x/time/rate\"

type RateLimitedSearchTool struct {
    *CustomSearchTool
    limiter *rate.Limiter
}

func NewRateLimitedSearchTool(apiKey string, requestsPerSecond float64) *RateLimitedSearchTool {
    return &RateLimitedSearchTool{
        CustomSearchTool: NewCustomSearchTool(apiKey),
        limiter:         rate.NewLimiter(rate.Limit(requestsPerSecond), 1),
    }
}

func (t *RateLimitedSearchTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // Wait for rate limiter
    if err := t.limiter.Wait(ctx); err != nil {
        return nil, fmt.Errorf(\"rate limit wait failed: %w\", err)
    }
    
    return t.CustomSearchTool.Execute(ctx, params)
}
```

### Caching Strategy

Configure caching for search results to improve performance and reduce API costs:

```toml
[mcp]
enabled = true
enable_caching = true
cache_timeout = 1800000  # 30 minutes for search results

# Longer cache for stable queries
[mcp.cache_rules]
\"news_*\" = 300000      # 5 minutes for news queries
\"weather_*\" = 1800000  # 30 minutes for weather
\"facts_*\" = 86400000   # 24 hours for factual queries
```

## Troubleshooting

### Common Issues

**Search not working:**
```bash
# Check MCP server status
agentcli mcp list --status

# Verify tool availability
agentcli mcp list --tools

# Check configuration
cat agentflow.toml | grep -A 10 \"\\[mcp\\]\"
```

**API rate limits:**
- Implement exponential backoff
- Use multiple API keys with rotation
- Cache results aggressively
- Filter queries to reduce API calls

**Poor search quality:**
- Refine search queries programmatically
- Use multiple search sources
- Implement result ranking
- Add domain-specific filters

## Security Considerations

### API Key Management

```bash
# Use environment variables
export SEARCH_API_KEY=\"your-key-here\"

# Or use a secrets manager
export SEARCH_API_KEY=$(vault kv get -field=key secret/search-api)
```

### Input Validation

```go
func validateSearchQuery(query string) error {
    if len(query) == 0 {
        return fmt.Errorf(\"search query cannot be empty\")
    }
    
    if len(query) > 500 {
        return fmt.Errorf(\"search query too long (max 500 characters)\")
    }
    
    // Check for potentially harmful content
    if containsHarmfulContent(query) {
        return fmt.Errorf(\"search query contains inappropriate content\")
    }
    
    return nil
}
```

## Next Steps

Now that you have web search capabilities:

1. **Combine with RAG**: Use search results to enhance your knowledge base
2. **Add Fact Checking**: Cross-reference search results with trusted sources
3. **Implement News Monitoring**: Set up automated news tracking
4. **Build Research Workflows**: Create multi-step research processes

## Related Guides

- [Research Assistant](research-assistant.md) - Complete research system
- [MCP Tools](../setup/mcp-tools.md) - MCP tool integration
- [Best Practices](best-practices.md) - Development guidelines
- [Debugging](debugging.md) - Troubleshooting tools

This guide covers web search integration in AgenticGoKit, enabling your agents to access real-time information from the web.