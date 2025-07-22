# 🔍 Research Assistant Example

**A production-ready multi-agent research system that automatically gathers, analyzes, and synthesizes information from multiple sources.**

## What This Demonstrates

- **🤝 Collaborative Orchestration**: Multiple agents working together in parallel
- **🌐 Web Search Integration**: Real-time information gathering via MCP tools
- **🧠 Memory & RAG**: Persistent knowledge storage and retrieval
- **📊 Production Monitoring**: Comprehensive logging and metrics
- **🔄 Error Handling**: Graceful failure recovery and retry logic

## Architecture

```mermaid
graph TB
    User[User Query] --> Coordinator[Research Coordinator]
    
    Coordinator --> Searcher[Web Searcher Agent]
    Coordinator --> Scholar[Academic Searcher Agent]  
    Coordinator --> Analyzer[Analysis Agent]
    
    Searcher --> WebAPI[Web Search API]
    Scholar --> ArxivAPI[ArXiv API]
    
    Searcher --> Memory[(Vector Memory)]
    Scholar --> Memory
    Analyzer --> Memory
    
    Memory --> Synthesizer[Synthesis Agent]
    Synthesizer --> Writer[Report Writer Agent]
    
    Writer --> Report[Final Research Report]
```

## Quick Start

### 1. One-Command Setup
```bash
git clone https://github.com/kunalkushwaha/agenticgokit
cd examples/research-assistant
docker-compose up -d  # Starts PostgreSQL with pgvector
```

### 2. Configure Environment
```bash
cp .env.example .env
# Edit .env with your API keys:
# OPENAI_API_KEY=your-openai-key
# SERP_API_KEY=your-serp-api-key (optional, for web search)
```

### 3. Run the Research Assistant
```bash
go run main.go "Latest developments in quantum computing applications"
```

### 4. Watch the Magic Happen
```
🔍 Research Assistant Starting...
📊 Initializing 5 specialized agents...
🌐 Connecting to web search and academic databases...
💾 Setting up vector memory with pgvector...

🤖 [Web Searcher] Finding recent quantum computing news...
🤖 [Academic Searcher] Searching ArXiv for latest papers...
🤖 [Analyzer] Processing gathered information...
🤖 [Synthesizer] Combining insights from multiple sources...
🤖 [Writer] Generating comprehensive research report...

✅ Research Complete! (23.4s)
📄 Generated 2,847-word research report
🔗 Included 15 sources with citations
💾 Stored findings in knowledge base for future queries
```

## Code Walkthrough

### Core Agent Setup

```go
// main.go - The heart of the research system
func main() {
    // 1. Initialize memory system with vector search
    memory := core.NewMemory(core.AgentMemoryConfig{
        Provider:     "pgvector",
        Connection:   "postgres://user:pass@localhost/research",
        RAGEnabled:   true,
        HybridSearch: true,
    })
    
    // 2. Create specialized research agents
    agents := map[string]core.AgentHandler{
        "web-searcher":     research.NewWebSearchAgent(memory),
        "academic-searcher": research.NewAcademicSearchAgent(memory),
        "analyzer":         research.NewAnalysisAgent(memory),
        "synthesizer":      research.NewSynthesisAgent(memory),
        "writer":           research.NewReportWriterAgent(memory),
    }
    
    // 3. Set up collaborative orchestration
    runner := core.CreateCollaborativeRunner(agents, 120*time.Second)
    
    // 4. Process research query
    result, err := runner.ProcessMessage(context.Background(), os.Args[1])
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(result.Content)
}
```

### Specialized Agent Implementation

```go
// agents/web_searcher.go - Web search specialist
type WebSearchAgent struct {
    llm    core.LLMProvider
    memory core.Memory
    tools  core.MCPManager
}

func NewWebSearchAgent(memory core.Memory) *WebSearchAgent {
    // Initialize with web search MCP tools
    core.QuickStartMCP("web_search", "url_scraper")
    
    return &WebSearchAgent{
        llm:    core.OpenAIProvider{Model: "gpt-4"},
        memory: memory,
        tools:  core.GetMCPManager(),
    }
}

func (w *WebSearchAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    query := state.Get("query").(string)
    
    // 1. Search the web for recent information
    searchResults, err := w.tools.ExecuteTool(ctx, "web_search", map[string]interface{}{
        "query": query,
        "num_results": 10,
        "time_range": "month", // Recent results only
    })
    
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("web search failed: %w", err)
    }
    
    // 2. Process and analyze search results
    analysis, err := w.llm.Generate(ctx, core.LLMRequest{
        Messages: []core.Message{
            {Role: "system", Content: webSearchSystemPrompt},
            {Role: "user", Content: fmt.Sprintf("Analyze these search results for: %s\n\nResults: %v", query, searchResults)},
        },
    })
    
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("analysis failed: %w", err)
    }
    
    // 3. Store findings in memory for other agents
    err = w.memory.Store(ctx, analysis.Content, "web-search-findings")
    if err != nil {
        log.Printf("Warning: Failed to store in memory: %v", err)
    }
    
    // 4. Return structured findings
    return core.AgentResult{
        OutputState: state.Set("web_findings", analysis.Content),
        Metadata: map[string]interface{}{
            "sources_found": len(searchResults.Content),
            "search_time":   time.Now(),
        },
    }, nil
}
```

### Memory Integration

```go
// The memory system automatically handles:
// - Document chunking and embedding
// - Vector similarity search  
// - Hybrid keyword + semantic search
// - Source attribution and citations

// Example: Querying memory across agents
func (s *SynthesisAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    query := state.Get("query").(string)
    
    // Search memory for relevant findings from all agents
    findings, err := s.memory.Search(ctx, query, 
        core.WithLimit(20),
        core.WithScoreThreshold(0.7),
        core.WithIncludePersonal(true),
        core.WithIncludeKnowledge(true),
    )
    
    if err != nil {
        return core.AgentResult{}, err
    }
    
    // Synthesize findings from multiple agents
    synthesis := s.synthesizeFindings(findings)
    
    return core.AgentResult{
        OutputState: state.Set("synthesis", synthesis),
    }, nil
}
```

## Configuration

### agentflow.toml
```toml
[project]
name = "research-assistant"
version = "1.0.0"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7

[memory]
provider = "pgvector"
connection = "postgres://user:pass@localhost:5432/research"
enable_rag = true
rag_chunk_size = 1000
rag_overlap = 200
hybrid_search = true

[mcp]
enable_discovery = true
servers = [
    { name = "web-search", type = "stdio", command = "npx @modelcontextprotocol/server-web-search" },
    { name = "arxiv", type = "stdio", command = "python -m arxiv_mcp_server" }
]

[orchestration]
mode = "collaborative"
timeout = "120s"
max_concurrency = 5
failure_threshold = 0.3

[monitoring]
enable_metrics = true
enable_tracing = true
log_level = "info"
```

### Docker Compose Setup
```yaml
# docker-compose.yml - Complete development environment
version: '3.8'
services:
  postgres:
    image: pgvector/pgvector:pg16
    environment:
      POSTGRES_DB: research
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  research-assistant:
    build: .
    depends_on:
      - postgres
      - redis
    environment:
      - DATABASE_URL=postgres://user:password@postgres:5432/research
      - REDIS_URL=redis://redis:6379
    volumes:
      - .:/app
    working_dir: /app

volumes:
  postgres_data:
```

## Advanced Features

### 1. Custom Research Domains
```go
// Add domain-specific agents
agents["patent-searcher"] = research.NewPatentSearchAgent(memory)
agents["news-analyzer"] = research.NewNewsAnalysisAgent(memory)
agents["trend-detector"] = research.NewTrendDetectionAgent(memory)
```

### 2. Real-time Collaboration Monitoring
```go
// Watch agents collaborate in real-time
runner.OnAgentStart(func(agentName string, input core.State) {
    fmt.Printf("🤖 [%s] Starting research task...\n", agentName)
})

runner.OnAgentComplete(func(agentName string, result core.AgentResult) {
    fmt.Printf("✅ [%s] Completed in %v\n", agentName, result.Duration)
})
```

### 3. Incremental Research
```go
// Build on previous research sessions
previousFindings, _ := memory.Search(ctx, query, core.WithTags([]string{"research-session"}))
state = state.Set("previous_research", previousFindings)
```

## Performance Metrics

**Typical Performance** (with GPT-4 and web search):
- **Research Time**: 15-45 seconds for comprehensive reports
- **Sources Processed**: 20-50 sources per query
- **Report Length**: 1,500-3,000 words with citations
- **Memory Usage**: ~50MB for full research session
- **Concurrent Queries**: Supports 10+ simultaneous research tasks

## Customization Examples

### Add New Research Sources
```go
// Add custom MCP server for specialized databases
core.ConnectMCPServer("pubmed", "stdio", "python -m pubmed_mcp_server")

// Create specialized medical research agent
agents["medical-researcher"] = research.NewMedicalResearchAgent(memory)
```

### Modify Research Focus
```go
// Focus on recent developments only
webSearchAgent.WithTimeRange("week")
academicAgent.WithDateFilter("2024-01-01")

// Add fact-checking agent
agents["fact-checker"] = research.NewFactCheckingAgent(memory)
```

### Export Research Data
```go
// Export research to different formats
exporter := research.NewExporter()
exporter.ToMarkdown(result, "research-report.md")
exporter.ToPDF(result, "research-report.pdf")
exporter.ToJSON(result, "research-data.json")
```

## Troubleshooting

### Common Issues

**🔍 "No search results found"**
- Check your SERP_API_KEY is valid
- Verify internet connectivity
- Try broader search terms

**💾 "Memory connection failed"**
- Ensure PostgreSQL is running: `docker-compose ps`
- Check connection string in agentflow.toml
- Verify pgvector extension: `docker-compose exec postgres psql -U user -d research -c "SELECT * FROM pg_extension WHERE extname='vector';"`

**🤖 "Agent timeout"**
- Increase timeout in configuration
- Check LLM provider rate limits
- Monitor agent performance with built-in metrics

### Performance Optimization

```go
// Optimize for speed
config.Orchestration.MaxConcurrency = 10  // More parallel agents
config.Memory.RAGChunkSize = 500          // Smaller chunks for faster search
config.LLM.Model = "gpt-3.5-turbo"       // Faster model

// Optimize for quality  
config.Memory.RAGTopK = 50                // More context for synthesis
config.LLM.Temperature = 0.3              // More focused responses
config.Orchestration.FailureThreshold = 0.1  // Less tolerance for errors
```

## Next Steps

- **[🔧 Build Custom MCP Tools](../custom-tools/)** - Create specialized research tools
- **[📊 Add Advanced Analytics](../analytics/)** - Track research patterns and insights
- **[🌐 Deploy to Production](../../production/deployment.md)** - Scale your research system
- **[🤝 Contribute](../../contributing/)** - Share your research agent improvements

---

**💡 Pro Tip**: This research assistant gets smarter over time as it builds up a knowledge base of previous research. Each query improves the system's ability to provide better, more comprehensive results!