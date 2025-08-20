# Troubleshooting Guide

**Quick solutions for common AgenticGoKit issues**

This guide helps you diagnose and resolve common problems when building with AgenticGoKit. Issues are organized by category with step-by-step solutions and prevention tips.

## 🚨 Quick Diagnostics

### Health Check Commands

```bash
# Check AgenticGoKit installation
agentcli version

# Verify project structure
agentcli validate

# Test basic functionality
go run . -m "Hello, world!" --dry-run
```

### Common Error Patterns

| Error Pattern | Likely Cause | Quick Fix |
|---------------|--------------|-----------|
| `connection refused` | Service not running | Start required services |
| `API key not found` | Missing environment variables | Set API keys |
| `agent not found` | Registration issue | Check agent registration |
| `timeout` | Performance or network issue | Increase timeouts |
| `out of memory` | Resource exhaustion | Optimize memory usage |

## 🔧 Installation & Setup Issues

### AgenticGoKit CLI Not Found

**Symptoms:**
```bash
agentcli: command not found
```

**Solutions:**

1. **Install AgenticGoKit CLI:**
```bash
# Using Go
go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest

# Verify installation
agentcli version
```

2. **Check PATH:**
```bash
# Add Go bin to PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Make permanent (add to ~/.bashrc or ~/.zshrc)
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
```

3. **Alternative installation:**
```bash
# Clone and build
git clone https://github.com/kunalkushwaha/agenticgokit.git
cd agenticgokit
go build -o agentcli ./cmd/agentcli
sudo mv agentcli /usr/local/bin/
```

### Go Module Issues

**Symptoms:**
```bash
go: module not found
go: cannot find module providing package
```

**Solutions:**

1. **Initialize Go module:**
```bash
go mod init your-project-name
go mod tidy
```

2. **Update dependencies:**
```bash
go get -u github.com/kunalkushwaha/agenticgokit@latest
go mod tidy
```

3. **Clear module cache:**
```bash
go clean -modcache
go mod download
```

### Project Creation Fails

**Symptoms:**
```bash
Error: failed to create project directory
Permission denied
```

**Solutions:**

1. **Check permissions:**
```bash
# Ensure write permissions in current directory
ls -la
chmod 755 .
```

2. **Use different directory:**
```bash
cd ~/projects
agentcli create my-project
```

3. **Run with elevated permissions (if necessary):**
```bash
sudo agentcli create my-project
sudo chown -R $USER:$USER my-project
```

## 🤖 Agent Execution Issues

### Agent Not Responding

**Symptoms:**
- No output from agents
- Workflow hangs indefinitely
- Silent failures

**Diagnosis:**

1. **Enable debug logging:**
```bash
export AGENTFLOW_LOG_LEVEL=debug
go run . -m "test message"
```

2. **Check agent registration:**
```go
// Verify agents are registered
func main() {
    agents := map[string]core.AgentHandler{
        "agent1": myAgent1,
        "agent2": myAgent2,
    }
    
    // Debug: Print registered agents
    for name := range agents {
        log.Printf("Registered agent: %s", name)
    }
    
    runner := core.CreateSequentialRunner(agents, []string{"agent1", "agent2"}, 30*time.Second)
    // ...
}
```

**Solutions:**

1. **Verify agent implementation:**
```go
func (a *MyAgent) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    log.Printf("Agent %s executing with event type: %s", a.name, event.Type)
    
    // Ensure you return a result
    return &core.AgentResult{
        Data: map[string]interface{}{
            "response": "Agent executed successfully",
        },
    }, nil
}
```

2. **Check event routing:**
```go
// Ensure event types match what agents expect
event := core.NewEvent("process", map[string]interface{}{
    "input": "test data",
})
```

3. **Add timeout handling:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := agent.Execute(ctx, event, state)
```

### Infinite Loops

**Symptoms:**
- High CPU usage
- Agents keep executing without finishing
- No final result

**Diagnosis:**
```go
// Add loop detection
type LoopDetector struct {
    executions map[string]int
    maxExecs   int
}

func (ld *LoopDetector) CheckLoop(agentName string) bool {
    ld.executions[agentName]++
    if ld.executions[agentName] > ld.maxExecs {
        log.Printf("WARNING: Agent %s executed %d times - possible loop", 
            agentName, ld.executions[agentName])
        return true
    }
    return false
}
```

**Solutions:**

1. **Set maximum iterations:**
```go
runner := core.CreateSequentialRunner(agents, agentOrder, 30*time.Second)
runner.SetMaxIterations(10) // Prevent infinite loops
```

2. **Add termination conditions:**
```go
func (a *MyAgent) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    // Check if work is already done
    if state.Data["completed"] == true {
        return &core.AgentResult{
            Data: map[string]interface{}{
                "message": "Work already completed",
            },
        }, nil
    }
    
    // Do work and mark as completed
    result := &core.AgentResult{
        Data: map[string]interface{}{
            "response": "Work done",
            "completed": true,
        },
    }
    
    return result, nil
}
```

3. **Use proper orchestration:**
```go
// For sequential processing, ensure proper agent order
agents := []string{"input", "process", "output"} // Clear sequence

// For collaborative, ensure agents don't interfere
runner, _ := core.NewRunnerFromConfig("agentflow.toml")
```

## 🔌 LLM Provider Issues

### API Key Errors

**Symptoms:**
```bash
Error: API key not found
Error: unauthorized
Error: invalid API key
```

**Solutions:**

1. **Set environment variables:**
```bash
# OpenAI
export OPENAI_API_KEY="your-key-here"

# Azure OpenAI
export AZURE_OPENAI_API_KEY="your-key-here"
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com/"
export AZURE_OPENAI_DEPLOYMENT="your-deployment-name"

# Ollama (if using local)
export OLLAMA_HOST="http://localhost:11434"
```

2. **Verify keys are loaded:**
```go
func main() {
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("OPENAI_API_KEY environment variable not set")
    }
    log.Printf("API key loaded: %s...", apiKey[:8]) // Show first 8 chars
}
```

3. **Use .env files:**
```bash
# Create .env file
cat > .env << EOF
OPENAI_API_KEY=your-key-here
AZURE_OPENAI_API_KEY=your-azure-key
EOF

# Load in your application
go get github.com/joho/godotenv
```

```go
import "github.com/joho/godotenv"

func init() {
    if err := godotenv.Load(); err != nil {
        log.Printf("No .env file found")
    }
}
```

### Rate Limiting

**Symptoms:**
```bash
Error: rate limit exceeded
Error: too many requests
HTTP 429 errors
```

**Solutions:**

1. **Implement exponential backoff:**
```go
import "time"

func retryWithBackoff(fn func() error, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }
        
        if strings.Contains(err.Error(), "rate limit") {
            delay := time.Duration(1<<i) * time.Second // Exponential backoff
            log.Printf("Rate limited, waiting %v before retry %d/%d", delay, i+1, maxRetries)
            time.Sleep(delay)
            continue
        }
        
        return err // Non-rate-limit error
    }
    return fmt.Errorf("max retries exceeded")
}
```

2. **Add request throttling:**
```go
import "golang.org/x/time/rate"

type ThrottledProvider struct {
    provider core.ModelProvider
    limiter  *rate.Limiter
}

func NewThrottledProvider(provider core.ModelProvider, requestsPerSecond float64) *ThrottledProvider {
    return &ThrottledProvider{
        provider: provider,
        limiter:  rate.NewLimiter(rate.Limit(requestsPerSecond), 1),
    }
}

func (tp *ThrottledProvider) GenerateResponse(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
    if err := tp.limiter.Wait(ctx); err != nil {
        return "", err
    }
    return tp.provider.GenerateResponse(ctx, prompt, options)
}
```

3. **Use multiple API keys:**
```go
type RotatingProvider struct {
    providers []core.ModelProvider
    current   int
    mutex     sync.Mutex
}

func (rp *RotatingProvider) GenerateResponse(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
    rp.mutex.Lock()
    provider := rp.providers[rp.current]
    rp.current = (rp.current + 1) % len(rp.providers)
    rp.mutex.Unlock()
    
    return provider.GenerateResponse(ctx, prompt, options)
}
```

### Connection Timeouts

**Symptoms:**
```bash
Error: context deadline exceeded
Error: connection timeout
Error: request timeout
```

**Solutions:**

1. **Increase timeouts:**
```toml
# In agentflow.toml
[runtime]
timeout_seconds = 60  # Increase from default 30

[providers.openai]
timeout_seconds = 45

[providers.azure]
timeout_seconds = 45
```

2. **Implement timeout handling:**
```go
func (a *MyAgent) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    // Create timeout context
    timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // Use timeout context for LLM calls
    response, err := a.llmProvider.GenerateResponse(timeoutCtx, prompt, nil)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return nil, fmt.Errorf("LLM request timed out after 30 seconds")
        }
        return nil, err
    }
    
    return &core.AgentResult{Data: map[string]interface{}{"response": response}}, nil
}
```

3. **Add retry logic:**
```go
func callWithRetry(ctx context.Context, provider core.ModelProvider, prompt string, maxRetries int) (string, error) {
    for i := 0; i < maxRetries; i++ {
        response, err := provider.GenerateResponse(ctx, prompt, nil)
        if err == nil {
            return response, nil
        }
        
        if errors.Is(err, context.DeadlineExceeded) && i < maxRetries-1 {
            log.Printf("Timeout on attempt %d/%d, retrying...", i+1, maxRetries)
            continue
        }
        
        return "", err
    }
    return "", fmt.Errorf("all retry attempts failed")
}
```

## 💾 Memory & Database Issues

### Database Connection Failed

**Symptoms:**
```bash
Error: failed to connect to database
Error: connection refused
Error: database does not exist
```

**Solutions:**

1. **Check database is running:**
```bash
# For PostgreSQL
docker compose ps
docker compose logs postgres

# Test connection manually
psql -h localhost -U agentflow -d agentflow -c "SELECT 1;"
```

2. **Verify connection string:**
```toml
[agent_memory]
provider = "pgvector"
connection = "postgres://agentflow:password@localhost:5432/agentflow?sslmode=disable"
```

3. **Initialize database:**
```bash
# Run setup script
./setup.sh

# Or manually
psql -h localhost -U agentflow -d agentflow -f init-db.sql
```

4. **Check Docker networking:**
```bash
# If using Docker, check network connectivity
docker network ls
docker compose exec postgres pg_isready -U agentflow
```

### Vector Search Not Working

**Symptoms:**
```bash
Error: vector extension not found
Error: operator does not exist for vector
Poor search results
```

**Solutions:**

1. **Verify pgvector extension:**
```sql
-- Connect to database
psql -h localhost -U agentflow -d agentflow

-- Check extension
SELECT * FROM pg_extension WHERE extname = 'vector';

-- Install if missing
CREATE EXTENSION IF NOT EXISTS vector;
```

2. **Check vector dimensions:**
```sql
-- Verify table structure
\d+ embeddings

-- Check if dimensions match your embedding model
SELECT embedding FROM embeddings LIMIT 1;
```

3. **Rebuild indexes:**
```sql
-- Drop and recreate vector indexes
DROP INDEX IF EXISTS embeddings_embedding_idx;
CREATE INDEX embeddings_embedding_idx 
ON embeddings USING ivfflat (embedding vector_cosine_ops) 
WITH (lists = 100);
```

### Memory Usage Too High

**Symptoms:**
- Application crashes with out of memory
- Slow performance
- High RAM usage

**Solutions:**

1. **Monitor memory usage:**
```go
import "runtime"

func logMemoryUsage() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    log.Printf("Memory: Alloc=%d KB, Sys=%d KB, NumGC=%d", 
        m.Alloc/1024, m.Sys/1024, m.NumGC)
}

// Call periodically
go func() {
    for range time.Tick(30 * time.Second) {
        logMemoryUsage()
    }
}()
```

2. **Optimize batch sizes:**
```toml
[agent_memory.embedding]
max_batch_size = 50  # Reduce from default 100

[agent_memory]
max_connections = 10  # Reduce connection pool
```

3. **Implement memory cleanup:**
```go
func (a *MyAgent) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    defer runtime.GC() // Force garbage collection after execution
    
    // Your agent logic...
    
    return result, nil
}
```

## 🔧 MCP Tool Issues

### Tools Not Available

**Symptoms:**
```bash
Error: tool not found
Error: MCP server not configured
No tools discovered
```

**Solutions:**

1. **Check MCP configuration:**
```bash
# Verify MCP is enabled
cat agentflow.toml | grep -A 10 "\[mcp\]"

# Check server status
agentcli mcp servers

# List available tools
agentcli mcp tools
```

2. **Enable MCP servers:**
```toml
[[mcp.servers]]
name = "brave-search"
type = "stdio"
command = "npx @modelcontextprotocol/server-brave-search"
enabled = true  # Make sure this is true
```

3. **Install MCP servers:**
```bash
# Install required MCP servers
npm install -g @modelcontextprotocol/server-brave-search
npm install -g @modelcontextprotocol/server-filesystem

# Test server directly
npx @modelcontextprotocol/server-brave-search
```

### Tool Execution Fails

**Symptoms:**
```bash
Error: tool execution failed
Error: timeout executing tool
Tool returned empty result
```

**Solutions:**

1. **Check tool parameters:**
```go
// Ensure proper parameter format
params := map[string]interface{}{
    "query": "search term",
    "num_results": 5,
}

result, err := toolManager.ExecuteTool(ctx, "web_search", params)
```

2. **Add error handling:**
```go
func (a *MyAgent) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    result, err := a.toolManager.ExecuteTool(ctx, "web_search", params)
    if err != nil {
        log.Printf("Tool execution failed: %v", err)
        
        // Provide fallback response
        return &core.AgentResult{
            Data: map[string]interface{}{
                "error": "Tool temporarily unavailable",
                "fallback": "Unable to search at this time",
            },
        }, nil
    }
    
    return &core.AgentResult{Data: map[string]interface{}{"result": result}}, nil
}
```

3. **Increase timeouts:**
```toml
[mcp]
connection_timeout = 30000  # 30 seconds
max_retries = 5
retry_delay = 2000
```

## 🚀 Performance Issues

### Slow Response Times

**Symptoms:**
- Long delays between agent responses
- High latency
- Timeouts

**Diagnosis:**
```go
// Add timing measurements
func (a *MyAgent) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    start := time.Now()
    defer func() {
        log.Printf("Agent %s took %v", a.name, time.Since(start))
    }()
    
    // Your agent logic...
}
```

**Solutions:**

1. **Optimize LLM calls:**
```go
// Use shorter prompts
prompt := fmt.Sprintf("Summarize: %s", truncateText(input, 1000))

// Cache responses
type CachedProvider struct {
    provider core.ModelProvider
    cache    map[string]string
    mutex    sync.RWMutex
}

func (cp *CachedProvider) GenerateResponse(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
    cp.mutex.RLock()
    if cached, exists := cp.cache[prompt]; exists {
        cp.mutex.RUnlock()
        return cached, nil
    }
    cp.mutex.RUnlock()
    
    response, err := cp.provider.GenerateResponse(ctx, prompt, options)
    if err == nil {
        cp.mutex.Lock()
        cp.cache[prompt] = response
        cp.mutex.Unlock()
    }
    
    return response, err
}
```

2. **Parallel processing:**
```go
// Process multiple agents concurrently
func processAgentsConcurrently(ctx context.Context, agents []core.AgentHandler, event core.Event, state *core.State) ([]*core.AgentResult, error) {
    results := make([]*core.AgentResult, len(agents))
    errors := make([]error, len(agents))
    
    var wg sync.WaitGroup
    for i, agent := range agents {
        wg.Add(1)
        go func(idx int, a core.AgentHandler) {
            defer wg.Done()
            result, err := a.Execute(ctx, event, state)
            results[idx] = result
            errors[idx] = err
        }(i, agent)
    }
    
    wg.Wait()
    
    // Check for errors
    for _, err := range errors {
        if err != nil {
            return nil, err
        }
    }
    
    return results, nil
}
```

3. **Database optimization:**
```sql
-- Add indexes for common queries
CREATE INDEX IF NOT EXISTS idx_agent_memory_session 
ON agent_memory(session_id, created_at);

-- Optimize vector search
CREATE INDEX embeddings_embedding_hnsw_idx 
ON embeddings USING hnsw (embedding vector_cosine_ops);
```

### High Resource Usage

**Solutions:**

1. **Limit concurrent operations:**
```go
// Use semaphore to limit concurrency
semaphore := make(chan struct{}, 5) // Max 5 concurrent operations

func (a *MyAgent) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    semaphore <- struct{}{} // Acquire
    defer func() { <-semaphore }() // Release
    
    // Your agent logic...
}
```

2. **Optimize memory usage:**
```go
// Use object pools for frequently allocated objects
var resultPool = sync.Pool{
    New: func() interface{} {
        return &core.AgentResult{
            Data: make(map[string]interface{}),
        }
    },
}

func (a *MyAgent) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    result := resultPool.Get().(*core.AgentResult)
    defer resultPool.Put(result)
    
    // Clear previous data
    for k := range result.Data {
        delete(result.Data, k)
    }
    
    // Use result...
    result.Data["response"] = "..."
    
    // Return a copy, not the pooled object
    return &core.AgentResult{Data: copyMap(result.Data)}, nil
}
```

## 🛠️ Development Issues

### Build Failures

**Symptoms:**
```bash
go: build failed
undefined: core.AgentHandler
package not found
```

**Solutions:**

1. **Update dependencies:**
```bash
go get -u github.com/kunalkushwaha/agenticgokit@latest
go mod tidy
```

2. **Check imports:**
```go
import (
    "github.com/kunalkushwaha/agenticgokit/core"
    // Not: "github.com/kunalkushwaha/agenticgokit/internal/core"
)
```

3. **Verify Go version:**
```bash
go version  # Should be 1.21 or later
```

### Testing Issues

**Solutions:**

1. **Mock LLM provider for tests:**
```go
type MockProvider struct {
    responses map[string]string
}

func (m *MockProvider) GenerateResponse(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
    if response, exists := m.responses[prompt]; exists {
        return response, nil
    }
    return "mock response", nil
}

// Use in tests
func TestMyAgent(t *testing.T) {
    mockProvider := &MockProvider{
        responses: map[string]string{
            "test prompt": "expected response",
        },
    }
    
    agent := NewMyAgent("test", mockProvider)
    // Test agent...
}
```

2. **Test with timeout:**
```go
func TestAgentTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    result, err := agent.Execute(ctx, event, state)
    if err != nil {
        t.Fatalf("Agent execution failed: %v", err)
    }
    
    // Verify result...
}
```

## 📞 Getting Help

### Before Asking for Help

1. **Check logs:**
```bash
export AGENTFLOW_LOG_LEVEL=debug
go run . -m "test" 2>&1 | tee debug.log
```

2. **Gather system info:**
```bash
go version
agentcli version
docker --version
cat agentflow.toml
```

3. **Create minimal reproduction:**
```go
// Simplest possible example that shows the issue
func main() {
    provider := core.NewMockProvider()
    agent := agents.NewSimpleAgent("test", provider)
    
    event := core.NewEvent("test", map[string]interface{}{"input": "test"})
    state := core.NewState()
    
    result, err := agent.Execute(context.Background(), event, state)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Result: %+v", result)
}
```

### Community Resources

- **[GitHub Issues](https://github.com/kunalkushwaha/AgenticGoKit/issues)** - Bug reports and feature requests
- **[GitHub Discussions](https://github.com/kunalkushwaha/AgenticGoKit/discussions)** - Questions and community help
- **[Discord Community](https://discord.gg/agenticgokit)** - Real-time chat and support

### Issue Template

When reporting issues, include:

```
**AgenticGoKit Version:** (agentcli version output)
**Go Version:** (go version output)
**Operating System:** (OS and version)

**Expected Behavior:**
What you expected to happen

**Actual Behavior:**
What actually happened

**Steps to Reproduce:**
1. Step one
2. Step two
3. Step three

**Code Sample:**
```go
// Minimal code that reproduces the issue
```

**Logs:**
```
// Relevant log output with debug level enabled
```

**Configuration:**
```toml
// Relevant parts of agentflow.toml
```
```

---

*This troubleshooting guide covers common issues with AgenticGoKit. The framework is actively developed, so some issues may be resolved in newer versions.*