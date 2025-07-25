# Research Assistant

**Create a multi-agent research system with web search, analysis, and synthesis capabilities**

This guide shows you how to build a comprehensive research assistant that can search the web, analyze information, and provide well-structured research reports using AgenticGoKit's multi-agent orchestration.

## What You'll Build

A research assistant system that can:
- Search the web for relevant information
- Analyze and validate sources
- Synthesize findings into coherent reports
- Handle complex multi-step research queries
- Maintain research context across interactions

## Prerequisites

- Basic AgenticGoKit project setup
- Understanding of multi-agent orchestration
- MCP tools configured for web search
- LLM provider configured (OpenAI, Azure, etc.)

## Quick Start

### 1. Create Research Assistant Project

```bash
# Create project with collaborative orchestration
agentcli create research-assistant \\
  --orchestration-mode collaborative \\
  --collaborative-agents \"researcher,analyzer,synthesizer\" \\
  --mcp-enabled \\
  --mcp-tools \"web_search,summarize\" \\
  --visualize
cd research-assistant
```

### 2. Configure Environment

```bash
# Set up API keys
export OPENAI_API_KEY=your-openai-key
export BRAVE_API_KEY=your-brave-search-key  # Optional

# Install dependencies
go mod tidy
```

### 3. Test the System

```bash
# Run a research query
go run . -m \"Research the latest developments in quantum computing and their potential impact on cryptography\"
```

## Architecture Overview

The research assistant uses a collaborative multi-agent architecture:

```mermaid
graph TD
    INPUT[\"🎯 Research Query\"]
    RESEARCHER[\"🔍 Researcher Agent\"]
    ANALYZER[\"📊 Analyzer Agent\"]
    SYNTHESIZER[\"📝 Synthesizer Agent\"]
    OUTPUT[\"📋 Research Report\"]
    
    INPUT --> RESEARCHER
    INPUT --> ANALYZER
    INPUT --> SYNTHESIZER
    
    RESEARCHER --> OUTPUT
    ANALYZER --> OUTPUT
    SYNTHESIZER --> OUTPUT
```

## Agent Implementation

### 1. Researcher Agent

The researcher agent handles web search and information gathering:

```go
package agents

import (
    \"context\"
    \"fmt\"
    \"strings\"
    \"time\"
    
    \"github.com/kunalkushwaha/agenticgokit/core\"
)

type ResearcherAgent struct {
    name        string
    llmProvider core.ModelProvider
    mcpManager  core.MCPManager
}

func NewResearcherAgent(name string, llm core.ModelProvider, mcp core.MCPManager) *ResearcherAgent {
    return &ResearcherAgent{
        name:        name,
        llmProvider: llm,
        mcpManager:  mcp,
    }
}

func (a *ResearcherAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    query, ok := event.GetData()[\"query\"].(string)
    if !ok {
        return core.AgentResult{}, fmt.Errorf(\"missing query in event data\")
    }
    
    // Generate search queries
    searchQueries, err := a.generateSearchQueries(ctx, query)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf(\"failed to generate search queries: %w\", err)
    }
    
    // Perform searches
    var allResults []SearchResult
    for _, searchQuery := range searchQueries {
        results, err := a.performSearch(ctx, searchQuery)
        if err != nil {
            // Log error but continue with other searches
            fmt.Printf(\"Search failed for query '%s': %v\\n\", searchQuery, err)
            continue
        }
        allResults = append(allResults, results...)
    }
    
    // Filter and rank results
    filteredResults := a.filterResults(allResults, query)
    
    // Store research findings in state
    state.Set(\"research_findings\", filteredResults)
    state.Set(\"search_queries\", searchQueries)
    state.SetMeta(\"researcher_completed\", \"true\")
    
    return core.AgentResult{
        OutputState: state,
    }, nil
}

func (a *ResearcherAgent) generateSearchQueries(ctx context.Context, query string) ([]string, error) {
    prompt := fmt.Sprintf(`
Given this research query: \"%s\"

Generate 3-5 specific search queries that would help gather comprehensive information.
Each query should focus on a different aspect of the topic.
Return only the search queries, one per line.
`, query)
    
    response, err := a.llmProvider.Generate(ctx, prompt)
    if err != nil {
        return nil, err
    }
    
    queries := strings.Split(strings.TrimSpace(response), \"\\n\")
    var cleanQueries []string
    for _, q := range queries {
        if cleaned := strings.TrimSpace(q); cleaned != \"\" {
            cleanQueries = append(cleanQueries, cleaned)
        }
    }
    
    return cleanQueries, nil
}

func (a *ResearcherAgent) performSearch(ctx context.Context, query string) ([]SearchResult, error) {
    // Use MCP web search tool
    result, err := a.mcpManager.CallTool(ctx, \"web_search\", map[string]interface{}{
        \"query\": query,
        \"num_results\": 10,
    })
    if err != nil {
        return nil, err
    }
    
    // Parse search results (implementation depends on search provider)
    return parseSearchResults(result), nil
}

func (a *ResearcherAgent) filterResults(results []SearchResult, originalQuery string) []SearchResult {
    // Implement relevance filtering and deduplication
    seen := make(map[string]bool)
    var filtered []SearchResult
    
    for _, result := range results {
        if seen[result.URL] {
            continue
        }
        seen[result.URL] = true
        
        // Add relevance scoring logic here
        if a.isRelevant(result, originalQuery) {
            filtered = append(filtered, result)
        }
    }
    
    return filtered
}

type SearchResult struct {
    Title   string `json:\"title\"`
    URL     string `json:\"url\"`
    Snippet string `json:\"snippet\"`
    Source  string `json:\"source\"`
}
```

### 2. Analyzer Agent

The analyzer agent evaluates source credibility and extracts key insights:

```go
type AnalyzerAgent struct {
    name        string
    llmProvider core.ModelProvider
}

func NewAnalyzerAgent(name string, llm core.ModelProvider) *AnalyzerAgent {
    return &AnalyzerAgent{
        name:        name,
        llmProvider: llm,
    }
}

func (a *AnalyzerAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Get research findings from state
    findings, exists := state.Get(\"research_findings\")
    if !exists {
        return core.AgentResult{}, fmt.Errorf(\"no research findings available\")
    }
    
    results := findings.([]SearchResult)
    query := event.Data[\"query\"].(string)
    
    // Analyze each source
    var analyses []SourceAnalysis
    for _, result := range results {
        analysis, err := a.analyzeSource(ctx, result, query)
        if err != nil {
            // Log error but continue with other sources
            fmt.Printf(\"Analysis failed for source %s: %v\\n\", result.URL, err)
            continue
        }
        analyses = append(analyses, analysis)
    }
    
    // Generate overall analysis
    overallAnalysis, err := a.generateOverallAnalysis(ctx, analyses, query)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf(\"failed to generate overall analysis: %w\", err)
    }
    
    // Store analysis in state
    state.Set(\"source_analyses\", analyses)
    state.Set(\"overall_analysis\", overallAnalysis)
    state.SetMeta(\"analyzer_completed\", \"true\")
    
    return core.AgentResult{
        OutputState: state,
    }, nil
}

func (a *AnalyzerAgent) analyzeSource(ctx context.Context, result SearchResult, query string) (SourceAnalysis, error) {
    prompt := fmt.Sprintf(`
Analyze this source for a research query about \"%s\":

Title: %s
URL: %s
Content: %s

Provide analysis in JSON format:
{
    \"credibility_score\": 0-10,
    \"relevance_score\": 0-10,
    \"key_insights\": [\"insight1\", \"insight2\"],
    \"potential_bias\": \"description\",
    \"source_type\": \"academic|news|blog|government|commercial\"
}
`, query, result.Title, result.URL, result.Snippet)
    
    response, err := a.llmProvider.Generate(ctx, prompt)
    if err != nil {
        return SourceAnalysis{}, err
    }
    
    // Parse JSON response
    var analysis SourceAnalysis
    if err := json.Unmarshal([]byte(response), &analysis); err != nil {
        return SourceAnalysis{}, fmt.Errorf(\"failed to parse analysis: %w\", err)
    }
    
    analysis.Source = result
    return analysis, nil
}

type SourceAnalysis struct {
    Source           SearchResult `json:\"source\"`
    CredibilityScore int          `json:\"credibility_score\"`
    RelevanceScore   int          `json:\"relevance_score\"`
    KeyInsights      []string     `json:\"key_insights\"`
    PotentialBias    string       `json:\"potential_bias\"`
    SourceType       string       `json:\"source_type\"`
}
```

### 3. Synthesizer Agent

The synthesizer agent creates the final research report:

```go
type SynthesizerAgent struct {
    name        string
    llmProvider core.ModelProvider
}

func NewSynthesizerAgent(name string, llm core.ModelProvider) *SynthesizerAgent {
    return &SynthesizerAgent{
        name:        name,
        llmProvider: llm,
    }
}

func (a *SynthesizerAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Get analysis from state
    analyses, exists := state.Get(\"source_analyses\")
    if !exists {
        return core.AgentResult{}, fmt.Errorf(\"no source analyses available\")
    }
    
    overallAnalysis, exists := state.Get(\"overall_analysis\")
    if !exists {
        return core.AgentResult{}, fmt.Errorf(\"no overall analysis available\")
    }
    
    query := event.GetData()[\"query\"].(string)
    sourceAnalyses := analyses.([]SourceAnalysis)
    
    // Generate comprehensive report
    report, err := a.generateReport(ctx, query, sourceAnalyses, overallAnalysis.(string))
    if err != nil {
        return core.AgentResult{}, fmt.Errorf(\"failed to generate report: %w\", err)
    }
    
    // Store final report
    state.Set(\"final_report\", report)
    state.SetMeta(\"synthesizer_completed\", \"true\")
    
    return core.AgentResult{
        OutputState: state,
    }, nil
}

func (a *SynthesizerAgent) generateReport(ctx context.Context, query string, analyses []SourceAnalysis, overallAnalysis string) (string, error) {
    // Prepare source summaries
    var sourceSummaries []string
    for _, analysis := range analyses {
        if analysis.CredibilityScore >= 7 && analysis.RelevanceScore >= 7 {
            summary := fmt.Sprintf(\"- %s (%s): %s\", 
                analysis.Source.Title, 
                analysis.Source.URL, 
                strings.Join(analysis.KeyInsights, \"; \"))
            sourceSummaries = append(sourceSummaries, summary)
        }
    }
    
    prompt := fmt.Sprintf(`
Create a comprehensive research report for the query: \"%s\"

Overall Analysis:
%s

Key Sources and Insights:
%s

Structure the report with:
1. Executive Summary
2. Key Findings
3. Detailed Analysis
4. Sources and References
5. Conclusions and Implications

Make it professional, well-structured, and cite sources appropriately.
`, query, overallAnalysis, strings.Join(sourceSummaries, \"\\n\"))
    
    return a.llmProvider.Generate(ctx, prompt)
}
```

## Configuration

### Agent Flow Configuration

The generated `agentflow.toml` includes:

```toml
[runner]
session_id = \"research-session\"
orchestration_mode = \"collaborate\"
timeout = \"120s\"
max_concurrency = 3
failure_threshold = 0.7

# Collaborative agents process in parallel
collaborative_agents = [\"researcher\", \"analyzer\", \"synthesizer\"]

[mcp]
enabled = true
enable_discovery = true
connection_timeout = \"30s\"
max_retries = 3
enable_caching = true
cache_timeout = \"1800s\"  # 30 minutes

[[mcp.servers]]
name = \"brave-search\"
type = \"stdio\"
command = \"npx @modelcontextprotocol/server-brave-search\"
enabled = true
```

### Main Application

```go
package main

import (
    \"context\"
    \"fmt\"
    \"log\"
    \"os\"
    
    \"github.com/kunalkushwaha/agenticgokit/core\"
    \"./agents\"
)

func main() {
    ctx := context.Background()
    
    // Initialize LLM provider
    llmProvider, err := core.NewOpenAIProvider(core.OpenAIConfig{
        APIKey: os.Getenv(\"OPENAI_API_KEY\"),
        Model:  \"gpt-4\",
    })
    if err != nil {
        log.Fatal(\"Failed to create LLM provider:\", err)
    }
    
    // Initialize MCP manager
    mcpManager, err := core.QuickStartMCP()
    if err != nil {
        log.Fatal(\"Failed to initialize MCP:\", err)
    }
    defer mcpManager.Close()
    
    // Create agents
    agentHandlers := map[string]core.AgentHandler{
        \"researcher\":   agents.NewResearcherAgent(\"researcher\", llmProvider, mcpManager),
        \"analyzer\":     agents.NewAnalyzerAgent(\"analyzer\", llmProvider),
        \"synthesizer\":  agents.NewSynthesizerAgent(\"synthesizer\", llmProvider),
    }
    
    // Create collaborative runner
    runner := core.CreateCollaborativeRunner(agentHandlers, 120*time.Second)
    
    // Get query from command line
    if len(os.Args) < 3 || os.Args[1] != \"-m\" {
        log.Fatal(\"Usage: go run . -m \\\"your research query\\\"\")
    }
    query := os.Args[2]
    
    // Create research event
    event := core.NewEvent(\"\", core.EventData{
        \"query\": query,
    }, nil)
    
    // Process research request
    results, err := runner.ProcessEvent(ctx, event)
    if err != nil {
        log.Fatal(\"Research failed:\", err)
    }
    
    // Display results
    fmt.Println(\"\\n=== RESEARCH REPORT ===\")
    if synthResult, ok := results[\"synthesizer\"]; ok && synthResult.Error == \"\" {
        if report, exists := synthResult.OutputState.Get(\"final_report\"); exists {
            fmt.Println(report)
        }
    }
    
    // Display statistics
    fmt.Println(\"\\n=== RESEARCH STATISTICS ===\")
    for agentName, result := range results {
        if result.Error != \"\" {
            fmt.Printf(\"%s: Error - %s\\n\", agentName, result.Error)
        } else {
            fmt.Printf(\"%s: Success\\n\", agentName)
        }
    }
}
```

## Advanced Features

### Research Context Management

Maintain research context across multiple queries:

```go
type ResearchContext struct {
    Topic           string
    PreviousQueries []string
    KnowledgeBase   map[string]interface{}
    Sources         []SearchResult
}

func (a *ResearcherAgent) RunWithContext(ctx context.Context, event core.Event, state core.State, researchContext *ResearchContext) (core.AgentResult, error) {
    query := event.GetData()[\"query\"].(string)
    
    // Add to context
    researchContext.PreviousQueries = append(researchContext.PreviousQueries, query)
    
    // Use context to refine search
    contextualQuery := a.refineQueryWithContext(query, researchContext)
    
    // Continue with normal processing
    return a.performContextualSearch(ctx, contextualQuery, researchContext)
}
```

### Source Validation

Implement advanced source validation:

```go
func (a *AnalyzerAgent) validateSource(ctx context.Context, result SearchResult) (bool, error) {
    // Check domain reputation
    if a.isBlacklistedDomain(result.URL) {
        return false, nil
    }
    
    // Check for academic or government sources
    if a.isAuthoritative(result.URL) {
        return true, nil
    }
    
    // Use LLM to assess credibility
    prompt := fmt.Sprintf(`
Assess the credibility of this source:
URL: %s
Title: %s
Content: %s

Is this a credible source? Respond with YES or NO and brief explanation.
`, result.URL, result.Title, result.Snippet)
    
    response, err := a.llmProvider.Generate(ctx, prompt)
    if err != nil {
        return false, err
    }
    
    return strings.HasPrefix(strings.ToUpper(response), \"YES\"), nil
}
```

### Report Templates

Create different report formats:

```go
type ReportTemplate struct {
    Name     string
    Sections []string
    Format   string
}

var ReportTemplates = map[string]ReportTemplate{
    \"academic\": {
        Name:     \"Academic Research Report\",
        Sections: []string{\"Abstract\", \"Introduction\", \"Literature Review\", \"Analysis\", \"Conclusion\", \"References\"},
        Format:   \"formal\",
    },
    \"business\": {
        Name:     \"Business Intelligence Report\",
        Sections: []string{\"Executive Summary\", \"Key Findings\", \"Market Analysis\", \"Recommendations\", \"Appendix\"},
        Format:   \"executive\",
    },
    \"news\": {
        Name:     \"News Summary\",
        Sections: []string{\"Headline\", \"Summary\", \"Key Points\", \"Sources\"},
        Format:   \"journalistic\",
    },
}
```

## Testing

### Unit Tests

```go
func TestResearcherAgent(t *testing.T) {
    // Create mock LLM and MCP
    mockLLM := &MockLLMProvider{}
    mockMCP := &MockMCPManager{}
    
    agent := NewResearcherAgent(\"test-researcher\", mockLLM, mockMCP)
    
    event := core.NewEvent(\"research\", map[string]interface{}{
        \"query\": \"test query\",
    })
    state := core.NewState()
    
    result, err := agent.Run(context.Background(), event, state)
    
    assert.NoError(t, err)
    assert.True(t, result.Success)
    assert.Contains(t, result.Data, \"findings\")
}
```

### Integration Tests

```bash
# Test the complete research flow
go test -v ./tests/integration/research_test.go
```

## Deployment

### Docker Configuration

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o research-assistant .

FROM alpine:latest
RUN apk --no-cache add ca-certificates nodejs npm
WORKDIR /app

# Install MCP servers
RUN npm install -g @modelcontextprotocol/server-brave-search

COPY --from=builder /app/research-assistant .
COPY agentflow.toml .

CMD [\"./research-assistant\"]
```

## Performance Optimization

### Parallel Processing

The collaborative orchestration automatically processes agents in parallel, but you can optimize further:

```go
// Optimize search parallelization
func (a *ResearcherAgent) performParallelSearches(ctx context.Context, queries []string) ([]SearchResult, error) {
    resultsChan := make(chan []SearchResult, len(queries))
    errorsChan := make(chan error, len(queries))
    
    for _, query := range queries {
        go func(q string) {
            results, err := a.performSearch(ctx, q)
            if err != nil {
                errorsChan <- err
                return
            }
            resultsChan <- results
        }(query)
    }
    
    var allResults []SearchResult
    for i := 0; i < len(queries); i++ {
        select {
        case results := <-resultsChan:
            allResults = append(allResults, results...)
        case err := <-errorsChan:
            log.Printf(\"Search error: %v\", err)
        case <-ctx.Done():
            return nil, ctx.Err()
        }
    }
    
    return allResults, nil
}
```

## Troubleshooting

### Common Issues

**Agents not collaborating properly:**
- Check orchestration mode is set to \"collaborate\"
- Verify all agents are registered correctly
- Check timeout settings

**Search results poor quality:**
- Refine search query generation
- Implement better result filtering
- Use multiple search providers

**Report generation slow:**
- Optimize LLM prompts
- Implement result caching
- Use streaming responses

## Next Steps

Enhance your research assistant with:

1. **Memory Integration**: Store research findings for future queries
2. **Document Analysis**: Add PDF and document processing capabilities
3. **Citation Management**: Implement proper academic citation formats
4. **Real-time Updates**: Monitor topics for new information
5. **Collaborative Features**: Allow multiple users to contribute to research

## Related Guides

- [Web Search Integration](web-search-integration.md) - Detailed search setup
- [Multi-Agent Orchestration](../../tutorials/core-concepts/orchestration-patterns.md) - Orchestration patterns
- [MCP Tools](../setup/mcp-tools.md) - Tool integration
- [Best Practices](best-practices.md) - Development guidelines

This research assistant demonstrates the power of AgenticGoKit's collaborative multi-agent orchestration for complex, multi-step tasks.
